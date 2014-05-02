SHELL := /bin/bash
PKG = github.com/Clever/sphinx
VERSION := 0.1
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
SUBPKGS = $(addprefix $(PKG)/,common handlers limitkeys matchers main)
PKGS = $(PKG) $(SUBPKGS)
.PHONY: test $(PKGS)

test: $(PKGS)
all: build test

build:
	go build -o bin/sphinxd -ldflags "-X main.version $(VERSION)dev-$(SHA)" $(PKG)/main

$(PKGS):
ifeq ($(LINT),1)
	golint $(GOPATH)/src/$@*/**.go
endif
	go get -d -t $@
ifeq ($(COVERAGE),1)
	go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	go tool cover -html=$(GOPATH)/src/$@/c.out
else
	go test $@ -test.v
endif

# creates a debian package for sphinx
# to install `sudo dpkg -i sphinx.deb`
dpkg:
	mkdir -p deb/sphinx/usr/local/bin
	mkdir -p deb/sphinx/var/lib/sphinx
	mkdir -p deb/sphinx/var/cache/sphinx
	cp bin/sphinxd deb/sphinx/usr/local/bin
	-dpkg-deb --build deb/sphinx

run:
	bin/sphinxd --config="./exmaple.yaml"

clean:
	rm -rf bin/sphinxd
	rm -rf main/main
	rm -rf deb/sphind.deb
	rm -rf deb/sphinx/usr
	rm -rf deb/var
