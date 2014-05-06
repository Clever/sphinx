SHELL := /bin/bash
PKG = github.com/Clever/sphinx
SUBPKGS = $(addprefix $(PKG)/,common handlers limitkeys matchers main)
PKGS = $(PKG) $(SUBPKGS)

.PHONY: test $(PKGS)

test: $(PKGS)

golint:
	go get github.com/golang/lint/golint

$(PKGS): PATH := $(PATH):$(GOPATH)/bin
$(PKGS): golint
	@echo ""
	@echo "FORMATTING $@..."
	go get -d -t $@
	gofmt -w=true $(GOPATH)/src/$@*/**.go
	@echo ""
ifneq ($(NOLINT),1)
	@echo "LINTING $@..."
	golint $(GOPATH)/src/$@*/**.go
	@echo ""
endif
ifeq ($(COVERAGE),1)
	go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING $@..."
	go test -v -bench=. $@
endif
