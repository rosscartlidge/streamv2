# StreamV2 Makefile

.PHONY: test build clean examples benchmark fmt lint vet

# Build configuration
BINARY_NAME=streamv2
BUILD_DIR=build
PKG=./pkg/stream

# Go commands
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOFMT=gofmt
GOLINT=golint
GOVET=$(GO) vet

# Test and build targets
all: test build

test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -html=coverage.out -o coverage.html

test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short $(PKG)

build:
	@echo "Building..."
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./examples/basic_example.go

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

examples:
	@echo "Running examples..."
	@echo "=== Basic Example ==="
	$(GO) run ./examples/basic_example.go
	@echo "=== Network Example ==="
	$(GO) run ./examples/network_example.go

benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem $(PKG)

fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

lint:
	@echo "Running linter..."
	$(GOLINT) $(PKG)

vet:
	@echo "Running go vet..."
	$(GOVET) $(PKG)

check: fmt vet lint test

# Installation targets
install:
	@echo "Installing StreamV2..."
	$(GO) install ./...

# Development targets
dev-setup:
	@echo "Setting up development environment..."
	$(GOGET) -u golang.org/x/lint/golint
	$(GOGET) -u golang.org/x/tools/cmd/goimports

# Documentation
docs:
	@echo "Generating documentation..."
	$(GO) doc -all $(PKG) > docs/API.md

# Docker targets (if needed in future)
docker-build:
	@echo "Building Docker image..."
	docker build -t streamv2:latest .

# Release targets
version:
	@git describe --tags --always --dirty

release-test:
	@echo "Testing release build..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./examples/basic_example.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./examples/basic_example.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./examples/basic_example.go

# Help target
help:
	@echo "Available targets:"
	@echo "  test         - Run all tests with coverage"
	@echo "  test-short   - Run short tests"
	@echo "  build        - Build the examples"
	@echo "  clean        - Clean build artifacts"
	@echo "  examples     - Run all examples"
	@echo "  benchmark    - Run benchmarks"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  vet          - Run go vet"
	@echo "  check        - Run fmt, vet, lint, and test"
	@echo "  install      - Install the package"
	@echo "  dev-setup    - Set up development tools"
	@echo "  docs         - Generate documentation"
	@echo "  help         - Show this help message"