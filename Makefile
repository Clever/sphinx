SHELL := /bin/bash
PKG := github.com/Clever/sphinx
SUBPKGS := $(shell ls -d */ | grep -v bin | grep -v deb | grep -v vendor | grep -v Godeps)
READMES := $(addsuffix README.md, $(SUBPKGS))
VERSION := $(shell cat deb/sphinx/DEBIAN/control | grep Version | cut -d " " -f 2)
RELEASE_NAME := $(shell cat CHANGES.md | head -n 1 | tail -c+3)
RELEASE_DOCS := $(shell cat CHANGES.md | tail -n+2 | sed -n '/\#/q;p')
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
TESTS := $(shell find . -name "*_test.go" | sed s/\.go// | grep -v "./vendor")
BENCHES := $(addsuffix "_bench", $(TESTS))
.PHONY: test $(PKGS) run clean build-release

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif

export GO15VENDOREXPERIMENT = 1

test: $(TESTS) docs
bench: $(BENCHES)
build: bin/sphinxd

bin/sphinxd: *.go **/*.go
	go build -o bin/sphinxd -gcflags "-N -l" -ldflags "-X main.version v$(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)

build-release:
	go build -o bin/sphinxd -ldflags "-X main.version v$(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)

$(GOPATH)/bin/golint:
	go get github.com/golang/lint/golint

$(TESTS): PATH := $(PATH):$(GOPATH)/bin
$(TESTS): THE_PKG = $(addprefix $(PKG)/, $(dir $@))
$(TESTS): $(GOPATH)/bin/golint
	@echo ""
	@echo "FORMATTING $@..."
	gofmt -w=true $(GOPATH)/src/$(THE_PKG)*.go
	@echo ""
	@echo "LINTING $@..."
	$(GOPATH)/bin/golint $(GOPATH)/src/$(THE_PKG)*.go
	@echo ""
ifeq ($(COVERAGE),1)
	@echo "TESTING COVERAGE $@..."
	go test -cover -coverprofile=$(GOPATH)/src/$(THE_PKG)/c.out $(THE_PKG) -test.v
	go tool cover -html=$(GOPATH)/src/$(THE_PKG)/c.out
	go tool cover -func=$(GOPATH)/src/$(THE_PKG)/c.out | tail -n1 | sed 's/[^0-9]*//' > $(GOPATH)/src/$(THE_PKG)/c.out.percent
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
deb: build-release
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
	rm -f deb/sphinx.deb

docs: $(READMES)
%/README.md: PATH := $(PATH):$(GOPATH)/bin
%/README.md: %/*.go
	@go get github.com/robertkrimen/godocdown/godocdown
	godocdown $(PKG)/$(shell dirname $@) > $@
