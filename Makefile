.PHONY: build run vendor
SHELL := /bin/bash
PKG := github.com/Clever/cron-admin
SUBKGS := $(addprefix $(PKG)/, db server)
EXECUTABLE := cron-admin
GOLINT := $(GOPATH)/bin/golint
GODEP := $(GOPATH)/bin/godep

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
	$(error must be running Go 1.5)
endif
export GO15VENDOREXPERIMENT=1

all: build clean

$(GOLINT):
	go get github.com/golang/lint/golint

clean:
	rm -rf $(GOPATH)/src/$(PKG)/build

build: clean
	go build -o build/$(EXECUTABLE) $(PKG)

run: build
	./build/$(EXECUTABLE)

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories

