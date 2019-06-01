include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: all build run vendor $(PKGS)
SHELL := /bin/bash
PKG := github.com/Clever/cron-admin
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.12))

export MONGO_TEST_DB ?= http://127.0.0.1:27017

all: build test

clean:
	-rm bin/*

build: clean
	go build -o bin/$(EXECUTABLE) $(PKG)



start-test-db:
	docker stop cron-admin-mongo; docker rm cron-admin-mongo; docker run --name cron-admin-mongo -p 27017:27017 -d mongo

test: $(PKGS)
$(PKGS): golang-test-all-strict-deps
	$(call golang-test-all-strict,$@)


run: build
	PORT=8080 bin/$(EXECUTABLE)


install_deps: golang-dep-vendor-deps
	$(call golang-dep-vendor)