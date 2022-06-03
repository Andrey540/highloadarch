# This file defines GNU Make targets

GOPATH=$(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

all: build test check

# Generates Go code before 'build' target
.PHONY: generate
generate:
	bin/grpc-generate $(foreach path,$(APP_PROTO_FILES), "$(path)")

.PHONY: modules
modules:
	go mod tidy

# Builds Go binaries
.PHONY: build
build: generate modules
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
