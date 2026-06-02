.PHONY: build build-static install clean test

BINARY := browser-cli
BUILD_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Support sandboxed Go environment
GOPATH ?= /tmp/gopath
GOMODCACHE ?= $(GOPATH)/pkg/mod
GO_ENV := GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE)

build:
	@mkdir -p $(BUILD_DIR)
	$(GO_ENV) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/browser-cli

# Static build without CGO (portable, no system library dependencies)
build-static:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO_ENV) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/browser-cli

install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GO_ENV) go test -v ./...

# Install Playwright browsers
setup-browsers:
	$(GO_ENV) go run ./cmd/browser-cli setup

# Development
dev: build
	./$(BUILD_DIR)/$(BINARY) --help

# Download dependencies
deps:
	$(GO_ENV) go mod download
	$(GO_ENV) go mod tidy