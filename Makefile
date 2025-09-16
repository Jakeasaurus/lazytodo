# lazytodo Makefile
# Electric productivity management

# Variables
BINARY_NAME=lazytodo
VERSION=0.2.0
BUILD_DIR=build
INSTALL_DIR?=/usr/local/bin

# Colors for make output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := build

.PHONY: help build clean install uninstall test fmt vet lint deps cross-compile release

## Show help
help:
	@echo "$(CYAN)⚡ lazytodo Makefile - Electric Build System ⚡$(RESET)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(RESET)"
	@echo "  $(GREEN)build$(RESET)         Build the binary"
	@echo "  $(GREEN)install$(RESET)       Install to system (uses install.sh)"
	@echo "  $(GREEN)uninstall$(RESET)     Remove from system"
	@echo "  $(GREEN)clean$(RESET)         Clean build artifacts"
	@echo "  $(GREEN)test$(RESET)          Run tests"
	@echo "  $(GREEN)fmt$(RESET)           Format code"
	@echo "  $(GREEN)vet$(RESET)           Run go vet"
	@echo "  $(GREEN)lint$(RESET)          Run golangci-lint (if available)"
	@echo "  $(GREEN)deps$(RESET)          Download dependencies"
	@echo "  $(GREEN)cross-compile$(RESET) Build for multiple platforms"
	@echo "  $(GREEN)release$(RESET)       Build release binaries"
	@echo ""

## Build the binary
build: deps
	@echo "$(CYAN)🔨 Building $(BINARY_NAME)...$(RESET)"
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) .
	@echo "$(GREEN)✅ Build complete: $(BINARY_NAME)$(RESET)"

## Build with debug info
build-debug: deps
	@echo "$(CYAN)🔨 Building $(BINARY_NAME) (debug)...$(RESET)"
	go build -o $(BINARY_NAME) .
	@echo "$(GREEN)✅ Debug build complete: $(BINARY_NAME)$(RESET)"

## Download dependencies
deps:
	@echo "$(CYAN)📦 Downloading dependencies...$(RESET)"
	go mod download
	go mod verify
	@echo "$(GREEN)✅ Dependencies ready$(RESET)"

## Run tests
test:
	@echo "$(CYAN)🧪 Running tests...$(RESET)"
	go test -v -race ./...
	@echo "$(GREEN)✅ Tests complete$(RESET)"

## Format code
fmt:
	@echo "$(CYAN)✨ Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(RESET)"

## Run go vet
vet:
	@echo "$(CYAN)🔍 Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✅ Vet complete$(RESET)"

## Run linter (if available)
lint:
	@echo "$(CYAN)🔍 Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✅ Lint complete$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found, skipping$(RESET)"; \
	fi

## Install using install script
install: build
	@echo "$(CYAN)📦 Installing $(BINARY_NAME)...$(RESET)"
	./install.sh
	@echo "$(GREEN)⚡ Installation complete!$(RESET)"

## Uninstall
uninstall:
	@echo "$(CYAN)🗑️  Uninstalling $(BINARY_NAME)...$(RESET)"
	@if [ -f "$(INSTALL_DIR)/$(BINARY_NAME)" ]; then \
		if [ -w "$(INSTALL_DIR)" ]; then \
			rm -f "$(INSTALL_DIR)/$(BINARY_NAME)"; \
		else \
			sudo rm -f "$(INSTALL_DIR)/$(BINARY_NAME)"; \
		fi; \
		echo "$(GREEN)✅ Uninstalled from $(INSTALL_DIR)$(RESET)"; \
	else \
		echo "$(YELLOW)ℹ️  $(BINARY_NAME) not found in $(INSTALL_DIR)$(RESET)"; \
	fi

## Clean build artifacts
clean:
	@echo "$(CYAN)🧹 Cleaning...$(RESET)"
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	go clean
	@echo "$(GREEN)✅ Clean complete$(RESET)"

## Cross-compile for multiple platforms
cross-compile: deps
	@echo "$(CYAN)🌐 Cross-compiling for multiple platforms...$(RESET)"
	mkdir -p $(BUILD_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe .
	
	@echo "$(GREEN)✅ Cross-compile complete in $(BUILD_DIR)/$(RESET)"

## Build release with checksums
release: cross-compile
	@echo "$(CYAN)📦 Creating release package...$(RESET)"
	cd $(BUILD_DIR) && sha256sum * > checksums.txt
	@echo "$(GREEN)⚡ Release ready in $(BUILD_DIR)/$(RESET)"
	@echo "$(CYAN)Files:$(RESET)"
	@ls -la $(BUILD_DIR)/

## Run all quality checks
check: fmt vet lint test
	@echo "$(GREEN)✅ All checks passed!$(RESET)"

## Quick development cycle
dev: clean fmt vet build test
	@echo "$(GREEN)⚡ Development cycle complete!$(RESET)"