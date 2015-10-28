SHELL := /bin/bash
PKG := github.com/Clever/leakybucket
SUBPKGSREL := memory redis
SUBPKGS := $(addprefix $(PKG)/,$(SUBPKGSREL))
PKGS := $(PKG) $(SUBPKGS)
GOLINT := $(GOPATH)/bin/golint
.PHONY: test $(PKGS) $(SUBPKGSREL)
GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif

export REDIS_URL ?= localhost:6379

export GO15VENDOREXPERIMENT = 1

test: $(PKGS)

$(GOLINT):
	go get github.com/golang/lint/golint

$(PKGS): $(GOLINT)
	go get -d -t $@
	$(GOLINT) $(GOPATH)/src/$@/*.go
	gofmt -w=true $(GOPATH)/src/$@/*.go
	go test $@ -test.v

$(SUBPKGSREL): %: $(addprefix $(PKG)/, %)
