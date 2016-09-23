include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

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
$(eval $(call golang-version-check,1.7))

test: $(PKGS)
$(PKGS): golang-test-all-strict-deps
	$(call golang-test-all-strict,$@)

bench: $(BENCHES)
build: bin/sphinxd

bin/sphinxd: *.go **/*.go
	go build -o bin/sphinxd -gcflags "-N -l" -ldflags "-X main.version v$(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)

build-release:
	go build -o bin/sphinxd -ldflags "-X main.version v$(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)

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

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))
