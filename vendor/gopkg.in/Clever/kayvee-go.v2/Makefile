SHELL := /bin/bash
PKG := gopkg.in/clever/kayvee-go.v0
SUBPKG_NAMES :=
SUBPKGS = $(addprefix $(PKG)/, $(SUBPKG_NAMES))
PKGS = $(PKG) $(SUBPKGS)

.PHONY: test golint README

test: docs tests.json $(PKGS)

golint:
	@go get github.com/golang/lint/golint

README.md: *.go
	@go get github.com/robertkrimen/godocdown/godocdown
	@godocdown $(PKG) > README.md

$(PKGS): golint docs
	@go get -d -t $@
	@gofmt -w=true $(GOPATH)/src/$@*/**.go
ifneq ($(NOLINT),1)
	@echo "LINTING..."
	@PATH=$(PATH):$(GOPATH)/bin golint $(GOPATH)/src/$@*/**.go
	@echo ""
endif
ifeq ($(COVERAGE),1)
	@go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	@go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING..."
	@go test $@ -test.v
	@echo ""
endif

docs: $(addsuffix /README.md, $(SUBPKG_NAMES)) README.md
%/README.md: PATH := $(PATH):$(GOPATH)/bin
%/README.md: %/*.go
	@go get github.com/robertkrimen/godocdown/godocdown
	@godocdown $(PKG)/$(shell dirname $@) > $@

tests.json:
	wget https://raw.githubusercontent.com/Clever/kayvee/master/tests.json -O test/tests.json
