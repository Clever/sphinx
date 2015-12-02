SHELL := /bin/bash
PKG := github.com/Clever/sphinx
PKGS := $(shell go list ./... | grep -v /vendor)
READMES := $(addsuffix README.md, $(PKGS))
VERSION := $(shell cat deb/sphinx/DEBIAN/control | grep Version | cut -d " " -f 2)
RELEASE_NAME := $(shell cat CHANGES.md | head -n 1 | tail -c+3)
RELEASE_DOCS := $(shell cat CHANGES.md | tail -n+2 | sed -n '/\#/q;p')
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
TESTS := $(shell find . -name "*_test.go" | sed s/\.go// | grep -v "./vendor")
BENCHES := $(addsuffix "_bench", $(TESTS))
.PHONY: test $(PKGS) run clean build-release vendor

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif
export GO15VENDOREXPERIMENT = 1

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get github.com/golang/lint/golint

GODEP := $(GOPATH)/bin/godep
$(GODEP):
	go get -u github.com/tools/godep

test: $(TESTS)
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
$(TESTS): $(GOLINT)
	@echo ""
	@echo "FORMATTING $@..."
	gofmt -w=true $(GOPATH)/src/$(THE_PKG)*.go
	@echo ""
	@echo "LINTING $@..."
	$(GOLINT) $(GOPATH)/src/$(THE_PKG)*.go
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

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories
