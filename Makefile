# Makefile for Landscaping SaaS Application

.PHONY: help build test clean dev migrate docker-up docker-down docker-build lint format deps security

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Start development environment
	docker-compose -f docker/docker-compose.yml up -d postgres redis mailpit
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Starting API server..."
	ENV=development go run backend/cmd/api/main.go

dev-worker: ## Start development worker
	ENV=development go run backend/cmd/worker/main.go

# Build targets
build: build-api build-worker ## Build all binaries

build-api: ## Build API server binary
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api backend/cmd/api/main.go

build-worker: ## Build worker binary
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/worker backend/cmd/worker/main.go

build-migrate: ## Build migration tool
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/migrate backend/cmd/migrate/main.go

# Testing
test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./backend/...

test-integration: ## Run integration tests
	go test -v -tags=integration ./backend/tests/...

coverage: test ## Show test coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Database
migrate-up: ## Run database migrations up
	go run backend/cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	go run backend/cmd/migrate/main.go down

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	go run backend/cmd/migrate/main.go create $(NAME)

# Docker
docker-up: ## Start all services with docker-compose
	docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop all services
	docker-compose -f docker/docker-compose.yml down

docker-build: ## Build Docker images
	docker build -f docker/Dockerfile.api -t landscaping-app/api .
	docker build -f docker/Dockerfile.worker -t landscaping-app/worker .

docker-logs: ## Show docker logs
	docker-compose -f docker/docker-compose.yml logs -f

# Code quality
lint: ## Run linter
	golangci-lint run backend/...

format: ## Format code
	go fmt ./backend/...
	goimports -w backend/

# Dependencies
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	go get -u ./backend/...
	go mod tidy

# Security
security: ## Run security scan
	gosec ./backend/...

# Clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f

# Install tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Environment setup
setup: install-tools deps ## Setup development environment
	@echo "Development environment setup complete!"
	@echo "Run 'make dev' to start the development server"

# Production build
build-prod: ## Build production binaries with optimizations
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -a -installsuffix cgo -o bin/api backend/cmd/api/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -a -installsuffix cgo -o bin/worker backend/cmd/worker/main.go