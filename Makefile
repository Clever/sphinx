SHELL := /bin/bash
PKG := github.com/Clever/sphinx
VERSION := $(shell cat deb/sphinx/DEBIAN/control | grep Version | cut -d " " -f 2)
RELEASE_NAME := $(shell cat CHANGES.md | head -n 1 | tail -c+3)
RELEASE_DOCS := $(shell cat CHANGES.md | tail -n+2 | sed -n '/\#/q;p')
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
TESTS := $(shell find . -name "*_test.go" | sed s/\.go//)
BENCHES := $(addsuffix "_bench", $(TESTS))
.PHONY: test $(PKGS) run clean

test: $(TESTS)
bench: $(BENCHES)
build: bin/sphinxd

release: github-release deb

	@if [ "$$GIT_DIRTY" == "+CHANGES" ]; then \
		echo "Uncommited changes. Exciting." ; exit 1 ; \
	fi

	@while [ -z "$$CONTINUE" ]; do \
		read -r -p "Tagging and Releasing v$(VERSION). Anything but Y or y to exit. [y/N] " CONTINUE; \
	done ; \
	if [ ! $$CONTINUE == "y" ]; then \
	if [ ! $$CONTINUE == "Y" ]; then \
		echo "Exiting." ; exit 1 ; \
	fi \
	fi

	github-release release \
		--user Clever \
		--repo sphinx \
		--tag v$(VERSION) \
		--name "$(RELEASE_NAME)" \
		--description "$(RELEASE_DOCS)" \
		--pre-release

	GITHUB_API=https://$(GITHUB_TOKEN):@api.github.com github-release upload \
		--user Clever \
		--repo sphinx \
		--tag v$(VERSION) \
		--name "sphinx-amd64.deb" \
		--file deb/sphinx.deb

bin/sphinxd: *.go **/*.go
	go build -o bin/sphinxd -ldflags "-X main.version v$(VERSION)-$(BRANCH)-$(SHA)$(GIT_DIRTY)" $(PKG)

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
deb: build test bench
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

github-release:
	go get github.com/aktau/github-release
