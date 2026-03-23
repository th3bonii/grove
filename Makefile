.PHONY: build test clean install help

BINARY_NAME=grove
BUILD_DIR=./dist
CMD_DIR=./cmd

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/...

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

## install: Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

## help: Show this help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+.*?:.*?## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
