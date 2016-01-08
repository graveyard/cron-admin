.PHONY: build run vendor
SHELL := /bin/bash
PKG := github.com/Clever/cron-admin
SUBPKGS := $(addprefix $(PKG)/, db server)
PKGS := $(PKG) $(SUBPKGS)
EXECUTABLE := cron-admin
GOLINT := $(GOPATH)/bin/golint
GODEP := $(GOPATH)/bin/godep

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
	$(error must be running Go 1.5)
endif
export GO15VENDOREXPERIMENT=1
export MONGO_TEST_DB ?= 127.0.0.1:27017

all: build test

$(GOLINT):
	go get github.com/golang/lint/golint

clean:
	rm -rf $(GOPATH)/src/$(PKG)/build

build: clean
	go build -o build/$(EXECUTABLE) $(PKG)

test: $(PKGS)

$(PKGS): $(GOLINT)
	@echo ""
	@echo "FORMATTING $@..."
	gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "LINTING $@..."
	$(GOLINT) $(GOPATH)/src/$@/*.go
	@echo "TESTING $@..."
	go test -v $@

run: build
	./build/$(EXECUTABLE)

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories

