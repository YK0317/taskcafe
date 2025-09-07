# Simple Taskcafe Build Automation
.PHONY: help build test clean dev deps

# Default target
help: ## Show this help message
	@echo "Simple Taskcafe Build Automation"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing frontend dependencies..."
	cd frontend && npm ci

build: deps ## Build entire application
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Building backend..."
	go build -o build/taskcafe ./cmd/taskcafe

test: ## Run all tests
	@echo "Running Go tests..."
	go test ./...
	@echo "Running frontend tests..."
	cd frontend && npm test -- --watchAll=false

test-backend: ## Run Go tests only
	@echo "Running Go tests..."
	go test -v ./...

test-frontend: ## Run React tests only
	@echo "Running frontend tests..."
	cd frontend && npm test -- --coverage --watchAll=false

clean: ## Clean build artifacts
	rm -rf build/
	rm -rf frontend/build/

dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose up --build

docker: ## Build Docker image
	docker build -t taskcafe:latest .

# CI targets
ci-test: deps test ## CI test pipeline
	@echo "CI tests completed successfully"

ci-build: deps test build ## CI build pipeline
	@echo "CI build completed successfully"
