# Makefile for Huifu Configuration System

# Variables
BINARY_NAME=huifu-server
GO=go
GOFLAGS=-v
PORT=8080

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) main.go
	@echo "Build complete!"

# Run the application
run:
	@echo "Starting server on port $(PORT)..."
	$(GO) run main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@if [ -f $(BINARY_NAME) ]; then rm $(BINARY_NAME); fi
	@if [ -f temp_config.json ]; then rm temp_config.json; fi
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies installed!"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "Tests complete!"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted!"

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Build for different platforms
build-all: build-linux build-windows build-darwin

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-amd64 main.go

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-windows-amd64.exe main.go

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-arm64 main.go

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t huifu-config-system:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) huifu-config-system:latest

# Development mode with hot reload (requires air)
dev:
	@if command -v air &> /dev/null; then \
		air; \
	else \
		echo "Air not installed. Run: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		$(MAKE) run; \
	fi

# Help command
help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run        - Run the application"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make deps       - Install dependencies"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Lint code"
	@echo "  make build-all  - Build for all platforms"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run - Run Docker container"
	@echo "  make dev        - Run in development mode with hot reload"
	@echo "  make help       - Show this help message"

.PHONY: build run clean deps test fmt lint build-all build-linux build-windows build-darwin docker-build docker-run dev help

# Default target
.DEFAULT_GOAL := help