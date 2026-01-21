.PHONY: build test test-e2e clean install help lint run-dev static-lint lint-fix build-e2e clean-e2e test-all

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=gomarklint
E2E_BINARY=gomarklint-e2e-test
BUILD_DIR=.

# Go parameters
GOCMD=go
GOLINT=golangci-lint
LINT_CONFIG=--config=./.golangci.yml
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .

test: ## Run unit tests (excluding E2E)
	@echo "Running tests..."
	$(GOTEST) ./... -v -skip '^TestE2E'

# E2E test binary
test-e2e: build-e2e ## Run end-to-end tests
	@echo "Running E2E tests..."
	$(GOTEST) ./e2e/... -v
	@$(MAKE) clean-e2e

build-e2e: ## Build binary for E2E tests
	@echo "Building E2E test binary..."
	$(GOBUILD) -o e2e/$(E2E_BINARY) .

clean-e2e: ## Clean E2E test binary
	@echo "Cleaning E2E test binary..."
	rm -f e2e/$(E2E_BINARY)

test-all: test test-e2e ## Run all tests (unit + E2E)
	@echo "All tests completed!"

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

lint-fix: ## Run golangci-lint and fix issues automatically
	@echo "Running golangci-lint fix..."
	$(GOLINT) run $(LINT_CONFIG) --fix

lint-self: ## Run gomarklint on the repo's README
	@echo "Running gomarklint on README.md"
	$(GORUN) . README.md --min-heading=1

static-lint: ## Run golangci-lint for static analysis
	@echo "Running golangci-lint..."
	$(GOLINT) run $(LINT_CONFIG)

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
