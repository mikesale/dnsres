.PHONY: all build test clean lint fmt vet coverage docker-build docker-run help release

# Variables
BINARY_NAME=dnsres
BINARY_TUI_NAME=dnsres-tui
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"
BUILD_DIR=build
RELEASE_DIR=release

# Default target
all: clean build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/dnsres

# Build the TUI application
build-tui:
	@echo "Building $(BINARY_TUI_NAME)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_TUI_NAME) ./cmd/dnsres-tui

# Cross-compilation targets
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe

# Create release packages
release: build-all
	@echo "Creating release packages..."
	@mkdir -p $(RELEASE_DIR)
	@cd $(BUILD_DIR) && \
		tar czf ../$(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar czf ../$(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-arm64 && \
		tar czf ../$(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar czf ../$(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz $(BINARY_NAME)-linux-arm64 && \
		zip ../$(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64-$(VERSION).zip $(BINARY_NAME)-windows-amd64.exe

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf $(BUILD_DIR)
	rm -rf $(RELEASE_DIR)

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 $(BINARY_NAME):$(VERSION)

# Generate mocks
mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/circuitbreaker/circuitbreaker.go -destination=internal/circuitbreaker/mocks/circuitbreaker_mock.go
	mockgen -source=internal/cache/cache.go -destination=internal/cache/mocks/cache_mock.go
	mockgen -source=internal/health/health.go -destination=internal/health/mocks/health_mock.go

# Install development dependencies
deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build the application"
	@echo "  build        - Build the application"
	@echo "  build-tui    - Build the TUI application"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  release      - Create release packages for all platforms"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  mocks        - Generate mocks"
	@echo "  deps         - Install development dependencies"
	@echo "  help         - Show this help message" 
