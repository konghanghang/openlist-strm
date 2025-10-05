.PHONY: build clean run test help

# Variables
BINARY_NAME=openlist-strm
BUILD_DIR=bin
BACKEND_DIR=backend
CMD_DIR=$(BACKEND_DIR)/cmd/server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@cd $(BACKEND_DIR) && go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run the application (compiled)
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME) --config config.yaml

## run-dev: Run from source without building
run-dev:
	@echo "Running from source..."
	@cd $(BACKEND_DIR) && go run cmd/server/main.go -config ../configs/config.example.yaml

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf data/
	@rm -rf logs/
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	@cd $(BACKEND_DIR) && go test -v ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@cd $(BACKEND_DIR) && go mod download
	@cd $(BACKEND_DIR) && go mod tidy

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@cd $(BACKEND_DIR) && go fmt ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@cd $(BACKEND_DIR) && golangci-lint run

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' | sed -e 's/^/ /'
