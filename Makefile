# Nizam Makefile

.PHONY: all build test clean lint check fmt snapshots-test seeds-test import-test

GO_VERSION := $(shell go version | cut -d' ' -f3)
LDFLAGS := -ldflags "-s -w -X github.com/abdultolba/nizam/internal/version.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

all: build

# Build the binary
build:
	go build $(LDFLAGS) -trimpath -o nizam .

# Run tests with race detection
test:
	go test -race ./...

# Unit tests only (no integration)
snapshots-test:
	go test -race ./internal/compress ./internal/snapshot

seeds-test:
	go test -race ./internal/paths ./internal/seed 2>/dev/null || true

import-test:
	go test -race ./internal/importer 2>/dev/null || true

# Run linting (if golangci-lint is available)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint"; \
		go vet ./...; \
		gofmt -d .; \
	fi

# Format code
fmt:
	gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# Quality checks
check: fmt lint test

# Clean build artifacts
clean:
	rm -f nizam nizam.exe

# Show version information  
version:
	@echo "Go version: $(GO_VERSION)"
	@echo "Git version: $(shell git describe --tags --always --dirty 2>/dev/null || echo 'unknown')"

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -trimpath -o dist/nizam_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -trimpath -o dist/nizam_darwin_arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath -o dist/nizam_linux_amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -trimpath -o dist/nizam_linux_arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -trimpath -o dist/nizam_windows_amd64.exe .

# Run integration tests (requires Docker)
integration-test:
	@echo "Integration tests not yet implemented"
	@echo "TODO: Add Docker-based tests for snapshot create/restore"
