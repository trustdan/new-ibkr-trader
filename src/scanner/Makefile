# IBKR Scanner Makefile

.PHONY: all build test clean run docker-build docker-run lint docs

# Variables
BINARY_NAME=scanner
DOCKER_IMAGE=ibkr-scanner
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: test build

# Build the scanner
build:
	@echo "Building scanner..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/scanner/main.go

# Run tests
test:
	@echo "Running unit tests..."
	go test -v -cover ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/...

# Run load tests
test-load:
	@echo "Running load tests..."
	go test -v -tags=load -timeout=30m ./tests/...

# Run all tests
test-all: test test-integration test-load

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run the scanner
run: build
	@echo "Starting scanner..."
	./bin/$(BINARY_NAME)

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

# Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):latest

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Generate mocks
mocks:
	@echo "Generating mocks..."
	go generate ./...

# Update dependencies
deps:
	@echo "Updating dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	go clean -cache

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@echo "Starting temporary server for OpenAPI spec..."
	@./bin/$(BINARY_NAME) &
	@sleep 2
	@curl -s http://localhost:8080/api/v1/openapi.json > docs/openapi.json
	@pkill -f $(BINARY_NAME)
	@echo "OpenAPI spec saved to docs/openapi.json"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang/mock/mockgen@latest

# Run API server with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	air

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Performance profiling
profile:
	@echo "Running performance profiling..."
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "CPU profile: cpu.prof"
	@echo "Memory profile: mem.prof"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 cmd/scanner/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 cmd/scanner/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/scanner/main.go

# Create release
release: test-all build-all
	@echo "Creating release $(VERSION)..."
	@mkdir -p releases/$(VERSION)
	@cp dist/* releases/$(VERSION)/
	@cp README.md releases/$(VERSION)/
	@tar -czf releases/scanner-$(VERSION).tar.gz -C releases/$(VERSION) .
	@echo "Release created: releases/scanner-$(VERSION).tar.gz"

# API testing with curl examples
test-api:
	@echo "Testing API endpoints..."
	@echo "Health check:"
	curl -s http://localhost:8080/api/v1/health | jq .
	@echo "\nScanning AAPL:"
	curl -s http://localhost:8080/api/v1/scan/AAPL | jq .
	@echo "\nGetting filters:"
	curl -s http://localhost:8080/api/v1/filters | jq .

# Help
help:
	@echo "Available targets:"
	@echo "  make build           - Build the scanner"
	@echo "  make test            - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-load       - Run load tests"
	@echo "  make test-all        - Run all tests"
	@echo "  make bench           - Run benchmarks"
	@echo "  make run             - Build and run the scanner"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Run Docker container"
	@echo "  make lint            - Run linter"
	@echo "  make fmt             - Format code"
	@echo "  make deps            - Update dependencies"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make docs            - Generate API documentation"
	@echo "  make coverage        - Generate coverage report"
	@echo "  make profile         - Run performance profiling"
	@echo "  make release         - Create release package"