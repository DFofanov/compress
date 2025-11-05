# PDF Compressor Makefile
# Build system and project management

# Variables
BINARY_NAME=compressor
MAIN_PATH=./cmd
BUILD_DIR=bin
COVERAGE_DIR=coverage

# Go parameters
GO_VERSION=1.23
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Determine binary extension for the executable
ifeq ($(GOOS),windows)
	BINARY_EXT=.exe
else
	BINARY_EXT=
endif
BINARY=$(BINARY_NAME)$(BINARY_EXT)

# Colors for output
RED=\033[31m
GREEN=\033[32m
YELLOW=\033[33m
BLUE=\033[34m
RESET=\033[0m

# Main commands
.PHONY: help install-deps build run test test-unit test-comprehensive test-all clean coverage lint format check-deps dev docker quickstart release

## install-deps: Install all dependencies
install-deps:
	@echo "$(YELLOW)📦 Installing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(RESET)"

## build: Build the application
build: check-deps
	@echo "$(YELLOW)🔨 Building application...$(RESET)"
	@if not exist "$(BUILD_DIR)" mkdir "$(BUILD_DIR)"
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) $(MAIN_PATH)
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(BINARY)$(RESET)"

## run: Run the application
run:
	@echo "$(BLUE)🚀 Starting PDF Compressor...$(RESET)"
	@go run $(MAIN_PATH)

## test-unit: Run unit tests
test-unit:
	@echo "$(YELLOW)🧪 Running unit tests...$(RESET)"
	@go test -v ./internal/...
	@echo "$(GREEN)✅ Unit tests passed$(RESET)"

## test-comprehensive: Run comprehensive tests
test-comprehensive:
	@echo "$(YELLOW)🔬 Running comprehensive tests...$(RESET)"
	@cd tests && go run comprehensive.go
	@echo "$(GREEN)✅ Comprehensive tests completed$(RESET)"

## test-all: Run all tests (unit + comprehensive)
test-all: test-unit test-comprehensive
	@echo "$(GREEN)✅ All tests passed$(RESET)"

## test: Alias for test-all
test: test-all

## coverage: Code coverage analysis
coverage:
	@echo "$(YELLOW)📊 Analyzing code coverage...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./internal/...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "$(GREEN)✅ Coverage report: $(COVERAGE_DIR)/coverage.html$(RESET)"

## lint: Code quality check
lint:
	@echo "$(YELLOW)🔍 Checking code quality...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)❌ golangci-lint not installed$(RESET)"; \
		echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✅ Linting completed$(RESET)"

## format: Code formatting
format:
	@echo "$(YELLOW)✨ Formatting code...$(RESET)"
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)⚠️  goimports not installed, using go fmt$(RESET)"; \
	fi
	@echo "$(GREEN)✅ Code formatted$(RESET)"

## check-deps: Check dependencies
check-deps:
	@echo "$(YELLOW)🔍 Checking dependencies...$(RESET)"
	@go version
	@go mod verify
	@echo "$(GREEN)✅ Dependencies verified$(RESET)"

## clean: Clean temporary files
clean:
	@echo "$(YELLOW)🧹 Cleaning temporary files...$(RESET)"
	@if exist "$(BUILD_DIR)" rmdir /s /q "$(BUILD_DIR)"
	@if exist "$(COVERAGE_DIR)" rmdir /s /q "$(COVERAGE_DIR)"
	@if exist tests\compressed rmdir /s /q tests\compressed
	@if exist tests\comprehensive_test.exe del tests\comprehensive_test.exe
	@if exist $(BINARY_NAME).exe del $(BINARY_NAME).exe
	@if exist $(BINARY_NAME) del $(BINARY_NAME)
	@if exist $(BINARY_NAME).log del $(BINARY_NAME).log
	@go clean
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

## dev: Development mode with auto-reload
dev:
	@echo "$(BLUE)🔥 Development mode (Ctrl+C to exit)$(RESET)"
	@where air >nul 2>&1 && ( \
		air \
	) || ( \
		echo "$(YELLOW)⚠️  air not installed, using regular run$(RESET)" && \
		echo "Install air: go install github.com/cosmtrek/air@latest" && \
		$(MAKE) run \
	)

## docker: Docker build
docker:
	@echo "$(YELLOW)🐳 Building Docker image...$(RESET)"
	@docker build -t pdf-compressor:latest .
	@echo "$(GREEN)✅ Docker image built$(RESET)"

## release: Создать релиз (Windows PowerShell)
release:
	@echo "$(YELLOW)Creating release...$(RESET)"
	@if [ -f "scripts/release-gitea.ps1" ]; then \
		pwsh -File scripts/release-gitea.ps1; \
	else \
		echo "$(RED)Release script not found$(RESET)"; \
		exit 1; \
	fi

## quickstart: Quick start for new users
quickstart: install-deps build
	@echo "$(GREEN)✅ Quickstart completed!$(RESET)"
	@echo ""
	@echo "$(BLUE)Next steps:$(RESET)"
	@echo "1. Create folders: mkdir test_pdfs output"
	@echo "2. Put PDF files in test_pdfs/"
	@echo "3. Run: make run"
	@echo ""

## help: Show help for commands
help:
	@echo "$(BLUE)PDF Compressor - Available commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Build and Run:$(RESET)"
	@echo "  make install-deps     - Install dependencies"
	@echo "  make build           - Build application"
	@echo "  make run             - Run application"
	@echo "  make dev             - Development mode (auto-reload)"
	@echo ""
	@echo "$(GREEN)Testing and Quality:$(RESET)"
	@echo "  make test            - Run all tests (unit + comprehensive)"
	@echo "  make test-unit       - Run unit tests only"
	@echo "  make test-comprehensive - Run comprehensive tests only"
	@echo "  make coverage        - Code coverage analysis"
	@echo "  make lint            - Code linting"
	@echo "  make format          - Code formatting"
	@echo ""
	@echo "$(GREEN)Utilities:$(RESET)"
	@echo "  make clean           - Clean temporary files"
	@echo "  make check-deps      - Check dependencies"
	@echo "  make docker          - Docker build"
	@echo "  make release         - Create release (PowerShell)"
	@echo "  make quickstart      - Quick start for new users"
	@echo ""

# Show help by default
.DEFAULT_GOAL := help
