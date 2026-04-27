.PHONY: build test test-e2e test-coverage check-coverage clean install help lint run-dev static-lint lint-fix build-e2e clean-e2e test-all bench bench-compare lint-self install-hooks

# Default target
.DEFAULT_GOAL := help

# Coverage threshold (percentage, integer)
COVERAGE_THRESHOLD ?= 100

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
	$(GOTEST) $(shell go list ./... | grep -v '/e2e') -v

# E2E test binary
test-e2e: build-e2e ## Run end-to-end tests
	@echo "Running E2E tests..."
	@$(GOTEST) ./e2e/... -v; \
	EXIT_CODE=$$?; \
	$(MAKE) clean-e2e; \
	exit $$EXIT_CODE

build-e2e: ## Build binary for E2E tests
	@echo "Building E2E test binary..."
	$(GOBUILD) -o e2e/$(E2E_BINARY) .

clean-e2e: ## Clean E2E test binary
	@echo "Cleaning E2E test binary..."
	rm -f e2e/$(E2E_BINARY)

test-all: test test-e2e ## Run all tests (unit + E2E)
	@echo "All tests completed!"

test-coverage: ## Run tests with coverage (report only)
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	$(GOTEST) $(shell go list ./... | grep -v '/e2e') -coverprofile=coverage/coverage.out
	$(GOCMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

check-coverage: ## Run tests with coverage and enforce minimum threshold
	@echo "Running tests with coverage (threshold: $(COVERAGE_THRESHOLD)%)..."
	@mkdir -p coverage
	$(GOTEST) $(shell go list ./... | grep -v '/e2e') -coverprofile=coverage/coverage.out
	@total=$$($(GOCMD) tool cover -func=coverage/coverage.out | grep '^total' | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $${total}%"; \
	if [ "$$(echo "$${total} < $(COVERAGE_THRESHOLD)" | bc)" = "1" ]; then \
		echo "FAIL: coverage $${total}% is below threshold $(COVERAGE_THRESHOLD)%"; exit 1; \
	fi
	@echo "Coverage OK."

bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem $(shell go list ./... | grep -v '/e2e') -run=^$$

bench-compare: ## Compare benchmarks against origin/main; blocks on ⚠️ +10%+ regression
	@bash scripts/bench-compare.sh

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf coverage

install: ## Install the binary locally
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) .

lint: ## Run gomarklint on testdata
	@echo "Running gomarklint on testdata..."
	$(GORUN) . testdata

lint-fix: ## Run golangci-lint and fix issues automatically
	@echo "Running golangci-lint fix..."
	$(GOLINT) run $(LINT_CONFIG) --fix

lint-self: ## Run gomarklint on repo markdown files
	@echo "Running gomarklint on repo markdown..."
	$(GORUN) . README.md README.ja.md docs/content --config .gomarklint.ci.json

static-lint: ## Run golangci-lint for static analysis
	@echo "Running golangci-lint..."
	$(GOLINT) run $(LINT_CONFIG)

run-dev: ## Run gomarklint with arguments (usage: make run-dev ARGS="path/to/file.md")
	$(GORUN) . $(ARGS)

init: ## Generate default .gomarklint.json config
	@echo "Generating default .gomarklint.json..."
	$(GORUN) . init

install-hooks: ## Install git hooks (pre-push)
	@echo "Installing git hooks..."
	@HOOKS_DIR=$$(git rev-parse --git-path hooks); \
	mkdir -p "$$HOOKS_DIR"; \
	cp scripts/pre-push "$$HOOKS_DIR/pre-push"; \
	chmod +x "$$HOOKS_DIR/pre-push"; \
	echo "pre-push hook installed to $$HOOKS_DIR/pre-push."

mod-tidy: ## Tidy go.mod and go.sum
	@echo "Tidying go.mod..."
	$(GOMOD) tidy

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
