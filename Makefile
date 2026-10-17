# Go parameters
BINARY_NAME=akrobat-msg
MAIN_PATH=./cmd/server
BUILD_DIR=./bin
GO=go

.PHONY: all build run test clean fmt vet tidy help

help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

build:
	@echo "Building $(BINARY_NAME).."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@echo "Running $(BINARY_NAME).."
	@$(BUILD_DIR)/$(BINARY_NAME)

test:
	$(GO) test -v -race -cover ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

clean:
	@echo "Cleaning.."
	@rm -rf $(BUILD_DIR)

all: fmt vet test build