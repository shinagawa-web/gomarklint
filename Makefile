.PHONY: build test clean install help lint run-dev

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=gomarklint
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) ./... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) ./... -coverprofile=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

install: ## Install the binary locally
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) .

lint: ## Run gomarklint on testdata
	@echo "Running gomarklint on testdata..."
	$(GORUN) . testdata

run-dev: ## Run gomarklint with arguments (usage: make run-dev ARGS="path/to/file.md")
	$(GORUN) . $(ARGS)

init: ## Generate default .gomarklint.json config
	@echo "Generating default .gomarklint.json..."
	$(GORUN) . init

mod-tidy: ## Tidy go.mod and go.sum
	@echo "Tidying go.mod..."
	$(GOMOD) tidy

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
