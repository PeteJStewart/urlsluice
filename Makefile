.PHONY: all build test coverage lint clean docs help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=urlsluice
COVERAGE_FILE=coverage.out

# Build parameters
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

all: test build ## Run tests and build

build: ## Build the binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/urlsluice

test: ## Run tests
	$(GOTEST) -v -race ./...

coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE)

lint: ## Run linters
	golangci-lint run

clean: ## Clean build directory
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

docs: ## Generate documentation
	godoc -http=:6060

install: ## Install project dependencies
	$(GOGET) -u golang.org/x/lint/golint
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

.DEFAULT_GOAL := help 