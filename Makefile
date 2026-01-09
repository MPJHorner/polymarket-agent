# Polytracker Makefile

BINARY_NAME=polytracker
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
BUILD_DIR=build

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build test run clean lint tidy help build-all build-linux build-darwin build-windows

all: build test

build: tidy
	@echo "Building Polytracker..."
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

test:
	@echo "Running tests..."
	go test ./...

run: build
	./$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	go clean

lint:
	@echo "Linting..."
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Cross-compilation targets
build-all: build-linux build-darwin build-windows
	@echo "All cross-compilation builds complete!"

build-linux: tidy
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go

build-darwin: tidy
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go

build-windows: tidy
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go

help:
	@echo "Available targets:"
	@echo "  build         - Build the binary for current platform"
	@echo "  test          - Run all tests"
	@echo "  run           - Build and run the binary"
	@echo "  clean         - Remove the binary and clean go cache"
	@echo "  lint          - Run go vet and golangci-lint (if available)"
	@echo "  tidy          - Run go mod tidy"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  build-linux   - Build for Linux (amd64, arm64)"
	@echo "  build-darwin  - Build for macOS (amd64, arm64)"
	@echo "  build-windows - Build for Windows (amd64)"
	@echo "  help          - Show this help message"

