SHELL := /bin/bash
PKG = github.com/Clever/sphinx
SUBPKGS = $(addprefix $(PKG)/,common handlers limitkeys matchers main)
PKGS = $(PKG) $(SUBPKGS)

.PHONY: test $(PKGS)

test: $(PKGS)

golint:
	@go get github.com/golang/lint/golint

$(PKGS): golint
	@go get -d -t $@
	@gofmt -w=true $(GOPATH)/src/$@*/**.go
ifneq ($(NOLINT),1)
	@echo ""
	@echo "LINTING $@..."
	@PATH=$(PATH):$(GOPATH)/bin golint $(GOPATH)/src/$@*/**.go
	@echo ""
endif
ifeq ($(COVERAGE),1)
	@go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	@go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING $@..."
	@go test $@
endif
