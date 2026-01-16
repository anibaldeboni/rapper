# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Build parameters
BINARY_NAME=rapper
BUILD_DIR=./build
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Build flags
LDFLAGS=-s -w
BUILD_FLAGS=-buildvcs=true -trimpath
RACE_FLAGS=-race

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: all build build-all clean test test-coverage lint install_deps mocks run help dev release

# Default target
default: help

# Build and test everything
all: lint test build

## help: Display this help message
help:
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@echo ""
	@echo "  $(COLOR_GREEN)build$(COLOR_RESET)          - Build the application binary"
	@echo "  $(COLOR_GREEN)build-all$(COLOR_RESET)      - Build binaries for all platforms"
	@echo "  $(COLOR_GREEN)clean$(COLOR_RESET)          - Remove build artifacts and cache"
	@echo "  $(COLOR_GREEN)test$(COLOR_RESET)           - Run tests"
	@echo "  $(COLOR_GREEN)test-coverage$(COLOR_RESET)  - Run tests with coverage report"
	@echo "  $(COLOR_GREEN)lint$(COLOR_RESET)           - Run linter"
	@echo "  $(COLOR_GREEN)install_deps$(COLOR_RESET)   - Download and install dependencies"
	@echo "  $(COLOR_GREEN)mocks$(COLOR_RESET)          - Generate mocks"
	@echo "  $(COLOR_GREEN)run$(COLOR_RESET)            - Build and run the application"
	@echo "  $(COLOR_GREEN)dev$(COLOR_RESET)            - Build with race detector (for development)"
	@echo "  $(COLOR_GREEN)release$(COLOR_RESET)        - Build optimized binary for release"
	@echo "  $(COLOR_GREEN)all$(COLOR_RESET)            - Run lint, test, and build"
	@echo ""

## build: Build the application binary with VCS info
build: install_deps
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH) -ldflags "$(LDFLAGS)" .
	@echo "$(COLOR_GREEN)✓ Build output: $(COLOR_BOLD)$(BINARY_PATH)$(COLOR_RESET)\n"

## build-all: Build binaries for all platforms (linux, darwin, windows)
build-all: install_deps
	@echo "$(COLOR_BLUE)Building for all platforms...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@echo "  Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags "$(LDFLAGS)" .
	@echo "  Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -ldflags "$(LDFLAGS)" .
	@echo "  Building for Darwin/macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags "$(LDFLAGS)" .
	@echo "  Building for Darwin/macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -ldflags "$(LDFLAGS)" .
	@echo "  Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -ldflags "$(LDFLAGS)" .
	@echo "$(COLOR_GREEN)✓ All binaries built in $(BUILD_DIR)/$(COLOR_RESET)\n"

## dev: Build with race detector for development
dev: install_deps
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME) with race detector...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) $(RACE_FLAGS) -o $(BINARY_PATH) .
	@echo "$(COLOR_GREEN)✓ Development build: $(COLOR_BOLD)$(BINARY_PATH)$(COLOR_RESET)\n"

## release: Build optimized binary for release
release: clean install_deps
	@echo "$(COLOR_BLUE)Building optimized release binary...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH) -ldflags "$(LDFLAGS)" .
	@echo "$(COLOR_GREEN)✓ Release binary: $(COLOR_BOLD)$(BINARY_PATH)$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Binary size: $(COLOR_RESET)$$(du -h $(BINARY_PATH) | cut -f1)\n"

## clean: Remove build artifacts and cache
clean:
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)\n"

## test: Run tests
test: install_deps
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GOTEST) -v -cover -race -timeout 30s ./...
	@echo ""

## test-coverage: Run tests with coverage report
test-coverage: install_deps
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GOTEST) -v -race -coverprofile=$(BUILD_DIR)/coverage.out -covermode=atomic -timeout 30s ./...
	@$(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: $(COLOR_BOLD)$(BUILD_DIR)/coverage.html$(COLOR_RESET)\n"

## lint: Run linter
lint:
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@golangci-lint run -v
	@echo ""

## install_deps: Download and install dependencies
install_deps:
	@echo "$(COLOR_BLUE)Installing dependencies...$(COLOR_RESET)"
	@$(GOGET) -v ./...
	@$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✓ Dependencies installed$(COLOR_RESET)\n"

## mocks: Generate mocks
mocks:
	@echo "$(COLOR_BLUE)Generating mocks...$(COLOR_RESET)"
	@$(GOCMD) generate ./...
	@echo "$(COLOR_GREEN)✓ Mocks generated$(COLOR_RESET)\n"

## run: Build and run the application
run: build
	@echo "$(COLOR_BLUE)Running $(BINARY_NAME)...$(COLOR_RESET)"
	@$(BINARY_PATH)
