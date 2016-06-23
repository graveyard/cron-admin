include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: all build run vendor $(PKGS)
SHELL := /bin/bash
PKG := github.com/Clever/cron-admin
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.6))

export MONGO_TEST_DB ?= 127.0.0.1:27017

all: build test

clean:
	-rm bin/*

build: clean
	go build -o bin/$(EXECUTABLE) $(PKG)

test: $(PKGS)
$(PKGS): golang-test-all-strict-deps
	$(call golang-test-all-strict,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

run: build
	bin/$(EXECUTABLE)
