include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: build run vendor
SHELL := /bin/bash
PKG := github.com/Clever/cron-admin
SUBPKGS := $(addprefix $(PKG)/, db server)
PKGS := $(PKG) $(SUBPKGS)
EXECUTABLE := cron-admin
GOLINT := $(GOPATH)/bin/golint
GODEP := $(GOPATH)/bin/godep


$(eval $(call golang-version-check,1.5))

export MONGO_TEST_DB ?= 127.0.0.1:27017

all: build test

clean:
	rm -rf $(GOPATH)/src/$(PKG)/build

build: clean
	go build -o build/$(EXECUTABLE) $(PKG)

test: $(PKGS)

$(PKGS): golang-test-all-deps
		$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
		$(call golang-godep-vendor,$(PKGS))

run: build
	./build/$(EXECUTABLE)
