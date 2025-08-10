# 18xx CLI Makefile

.PHONY: build test clean run install lint coverage help

# Default target
all: build

# Build the application
build:
	@echo "Building 18xx CLI..."
	go build -o bin/18xx cmd/main.go

# Run tests
test:
	@echo "Running tests..."
	go test ./internal/... -v

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test ./internal/... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the application
run: build
	./bin/18xx

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f bin/18xx
	rm -f coverage.out
	rm -f coverage.html

# Install the application globally
install: build
	@echo "Installing 18xx CLI..."
	cp bin/18xx /usr/local/bin/

# Create release build
release:
	@echo "Building release..."
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/18xx-linux-amd64 cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/18xx-darwin-amd64 cmd/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/18xx-darwin-arm64 cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/18xx-windows-amd64.exe cmd/main.go

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	mkdir -p bin

# Help target
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  run       - Build and run the application"
	@echo "  deps      - Install/update dependencies"
	@echo "  lint      - Run code linter"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install globally"
	@echo "  release   - Build for multiple platforms"
	@echo "  dev-setup - Setup development environment"
	@echo "  help      - Show this help"