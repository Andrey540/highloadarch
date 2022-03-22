# This file defines GNU Make targets

GOPATH=$(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

all: build test check

.PHONY: modules
modules:
	go mod tidy

# Builds Go binaries
.PHONY: build
build: modules
	bin/run-go-build $(foreach name,$(APP_CMD_NAMES), "$(name)")

# Removes built Go binaries
.PHONY: clean
clean:
	rm -f $(foreach name,$(APP_CMD_NAMES), "bin/$(name)")

.PHONY: test
test:
	go test ./...

.PHONY: check
check:
	golangci-lint run
