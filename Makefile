SHELL := /bin/bash
PKG := github.com/Clever/sphinx
VERSION := 0.1
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
TESTS := $(shell find . -name "*_test.go")
BENCHES := $(addsuffix "_bench", $(TESTS))
.PHONY: test $(PKGS) run clean

test: $(TESTS)
bench: $(BENCHES)
build: bin/sphinxd

bin/sphinxd: *.go **/*.go
	go build -o bin/sphinxd -ldflags "-X main.version $(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)/main

golint:
	go get github.com/golang/lint/golint

$(TESTS): PATH := $(PATH):$(GOPATH)/bin
$(TESTS): THE_PKG = $(addprefix $(PKG)/, $(dir $@))
$(TESTS): golint
	@echo ""
	@echo "FORMATTING $@..."
	go get -d -t $(THE_PKG)
	gofmt -w=true $(GOPATH)/src/$(THE_PKG)*.go
	@echo ""
ifneq ($(NOLINT),1)
	@echo "LINTING $@..."
	golint $(GOPATH)/src/$(THE_PKG)*.go
	@echo ""
endif
ifeq ($(COVERAGE),1)
	@echo "TESTING COVERAGE $@..."
	go test -cover -coverprofile=$(GOPATH)/src/$(THE_PKG)/c.out $(THE_PKG) -test.v
	go tool cover -html=$(GOPATH)/src/$(THE_PKG)/c.out
else
	@echo "TESTING $@..."
	go test -v $(THE_PKG)
endif

$(BENCHES): THE_PKG = $(addprefix $(PKG)/, $(dir $@))
$(BENCHES): READABLE_NAME = $(shell echo $@ | sed s/_bench//)
$(BENCHES):
	@echo ""
	@echo "BENCHMARKING $(READABLE_NAME)..."
	go test -bench=. $(THE_PKG)

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
