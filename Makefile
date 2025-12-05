.PHONY: help build test test-v test-cover clean fmt vet lint install run all

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^##' Makefile | sed 's/^## /  /'
	@echo ""

## build: Build the CLI binary
build:
	@echo "Building diagtool..."
	@go build -o bin/diagtool ./cmd/diagtool
	@echo "✓ Built bin/diagtool"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing diagtool..."
	@go install ./cmd/diagtool
	@echo "✓ Installed diagtool"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test ./...

## test-v: Run tests with verbose output
test-v:
	@echo "Running tests (verbose)..."
	@go test -v ./...

## test-cover: Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@echo ""
	@echo "For detailed coverage:"
	@echo "  go test -coverprofile=coverage.out ./..."
	@echo "  go tool cover -html=coverage.out"

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ No issues found"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out
	@echo "✓ Cleaned"

## run: Build and run the CLI
run: build
	@echo "Running diagtool..."
	@./bin/diagtool

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	@go mod tidy
	@echo "✓ Modules tidied"

## verify: Run all verification checks (fmt, vet, test)
verify: fmt vet test
	@echo "✓ All verification checks passed"

## all: Build and run all checks
all: clean verify build
	@echo "✓ Build complete and all checks passed"
