SHELL := /bin/bash
PKG := github.com/Clever/sphinx
VERSION := 0.1
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_DIRTY := $(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
SUBPKGS := $(addprefix $(PKG)/,common handlers limitkeys matchers main)
PKGS := $(PKG) $(SUBPKGS)
GO_SRC_FILES := $(shell find . -name "*.go" | grep -v "_test.go")
TESTS := $(shell find . -name "*_test.go")

.PHONY: test $(PKGS) run clean deps

$(TESTS): PATH := $(PATH):$(GOPATH)/bin
$(TESTS): THE_PKG = $(addprefix $(PKG)/, $(dir $@))
$(TESTS): $(GO_SRC_FILES)
	@echo ""
	@echo "FORMATTING $@..."
	go get -d -t $(THE_PKG)
	gofmt -w=true $(GOPATH)/src/$(THE_PKG)/*.go
	@echo ""
	@echo "LINTING... $@"
	golint $(GOPATH)/src/$(THE_PKG)/*.go
ifeq ($(COVERAGE),1)
	@echo "TESTING COVERAGE $@..."
	go test -bench=. -cover -coverprofile=$(GOPATH)/src/$(THE_PKG)/c.out $(THE_PKG) -test.v
	go tool cover -html=$(GOPATH)/src/$(THE_PKG)/c.out
else
	@echo "TESTING $@..."
	go test -v -bench=. $(THE_PKG)
endif
	touch $@

bin/sphinxd: $(GO_SRC_FILES)
	go build -o bin/sphinxd -ldflags "-X main.version $(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)/main

test: $(TESTS)

build: bin/sphinxd

deps: golint cover

cover:
	go get code.google.com/p/go.tools/cmd/cover

golint:
	go get github.com/golang/lint/golint

# creates a debian package for sphinx
# to install `sudo dpkg -i sphinx.deb`
deb: build test
	mkdir -p deb/sphinx/usr/local/bin
	mkdir -p deb/sphinx/var/lib/sphinx
	mkdir -p deb/sphinx/var/cache/sphinx
	cp bin/sphinxd deb/sphinx/usr/local/bin
	-dpkg-deb --build deb/sphinx

run: build
	bin/sphinxd --config="./example.yaml"

clean:
	rm -rf deb/sphinx/usr
	rm -rf deb/var
	rm -f bin/sphinxd
	rm -f main/main
	rm -f deb/sphind.deb
