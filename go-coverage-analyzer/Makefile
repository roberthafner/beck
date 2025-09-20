.PHONY: build test clean install deps lint fmt vet run-example help

# Variables
BINARY_NAME=gcov
BUILD_DIR=bin
MAIN_PATH=./cmd/gcov
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Build targets
build: deps
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-release: deps
	@echo "🚀 Building release version..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Release binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy

# Testing
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📋 Coverage report generated: coverage.html"

test-integration:
	@echo "🔧 Running integration tests..."
	@go test -v -tags=integration ./...

# Code quality
lint:
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run

fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@goimports -w . 2>/dev/null || echo "goimports not available, run: go install golang.org/x/tools/cmd/goimports@latest"

vet:
	@echo "🔎 Vetting code..."
	@go vet ./...

# Installation
install: build
	@echo "📥 Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ $(BINARY_NAME) installed to /usr/local/bin/"

uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) uninstalled"

# Examples and testing
run-example: build
	@echo "🚀 Running example analysis..."
	@./$(BUILD_DIR)/$(BINARY_NAME) analyze ./examples/simple-project --verbose

run-generate-example: build
	@echo "🛠️  Running example test generation..."
	@./$(BUILD_DIR)/$(BINARY_NAME) generate ./examples/simple-project --dry-run --verbose

demo: build
	@echo "🎬 Running full demo..."
	@./$(BUILD_DIR)/$(BINARY_NAME) analyze ./examples/simple-project --verbose --output console
	@echo ""
	@./$(BUILD_DIR)/$(BINARY_NAME) generate ./examples/simple-project --dry-run --verbose

# Development workflow
dev: fmt vet test

ci: deps fmt vet lint test

pre-commit: fmt vet lint test

# Documentation
docs:
	@echo "📚 Generating documentation..."
	@go doc -all ./... > docs/API.md
	@echo "✅ Documentation generated in docs/API.md"

# Benchmarking
bench:
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Profiling
profile-cpu:
	@echo "🔥 Running CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./...
	@go tool pprof cpu.prof

profile-mem:
	@echo "💾 Running memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./...
	@go tool pprof mem.prof

# Clean up
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f cpu.prof mem.prof
	@rm -f *.test
	@echo "✅ Cleanup complete"

clean-examples:
	@echo "🧹 Cleaning example projects..."
	@find examples -name "*.test" -delete
	@find examples -name "coverage.out" -delete
	@find examples -name "*_test.go" -path "*/generated/*" -delete

# Docker support
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: docker-build
	@echo "🐳 Running in Docker..."
	@docker run --rm -v $(PWD):/workspace $(BINARY_NAME):$(VERSION)

# Release management
tag:
	@echo "🏷️  Creating git tag..."
	@read -p "Enter tag version (e.g., v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version

release: clean build-release test
	@echo "🎉 Release ready: $(BUILD_DIR)/$(BINARY_NAME)"

# Utility targets
version:
	@echo "📌 Version: $(VERSION)"

info:
	@echo "📋 Project Information:"
	@echo "   Name: $(BINARY_NAME)"
	@echo "   Version: $(VERSION)"
	@echo "   Go Version: $(shell go version)"
	@echo "   Build Dir: $(BUILD_DIR)"
	@echo "   Main Path: $(MAIN_PATH)"

check-tools:
	@echo "🔧 Checking required tools..."
	@which go >/dev/null || (echo "❌ Go not found" && exit 1)
	@which git >/dev/null || echo "⚠️  Git not found (optional)"
	@which golangci-lint >/dev/null || echo "⚠️  golangci-lint not found (run: make lint to install)"
	@which goimports >/dev/null || echo "⚠️  goimports not found (run: go install golang.org/x/tools/cmd/goimports@latest)"
	@echo "✅ Tool check complete"

# Help
help:
	@echo "🚀 Go Coverage Analyzer & Test Generator"
	@echo "========================================"
	@echo ""
	@echo "Build Commands:"
	@echo "  build           Build the binary"
	@echo "  build-release   Build optimized release binary"
	@echo "  install         Install binary to /usr/local/bin"
	@echo "  uninstall       Remove binary from /usr/local/bin"
	@echo ""
	@echo "Development Commands:"
	@echo "  deps            Install dependencies"
	@echo "  test            Run all tests"
	@echo "  test-coverage   Run tests with coverage report"
	@echo "  lint            Run linter"
	@echo "  fmt             Format code"
	@echo "  vet             Vet code"
	@echo "  dev             Run fmt, vet, and test"
	@echo "  ci              Full CI workflow"
	@echo ""
	@echo "Example Commands:"
	@echo "  run-example     Analyze example project"
	@echo "  run-generate-example  Generate tests for example (dry-run)"
	@echo "  demo            Run full demo"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean           Clean build artifacts"
	@echo "  version         Show version"
	@echo "  info            Show project information"
	@echo "  check-tools     Check required tools"
	@echo "  help            Show this help"