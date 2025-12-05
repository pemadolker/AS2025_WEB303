.PHONY: help test test-unit test-integration test-e2e test-coverage install-deps proto-generate docker-build docker-up docker-down clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-deps: ## Install testing dependencies
	@echo "Installing test dependencies..."
	@cd user-service && go get github.com/stretchr/testify/assert github.com/stretchr/testify/require github.com/stretchr/testify/mock gorm.io/driver/sqlite
	@cd menu-service && go get github.com/stretchr/testify/assert github.com/stretchr/testify/require gorm.io/driver/sqlite
	@cd order-service && go get github.com/stretchr/testify/assert github.com/stretchr/testify/require github.com/stretchr/testify/mock gorm.io/driver/sqlite
	@cd tests/integration && go mod init integration-tests 2>/dev/null || true
	@cd tests/integration && go get github.com/stretchr/testify/assert github.com/stretchr/testify/require github.com/stretchr/testify/mock google.golang.org/grpc google.golang.org/grpc/test/bufconn gorm.io/driver/sqlite gorm.io/gorm github.com/douglasswm/student-cafe-protos
	@cd tests/e2e && go mod init e2e-tests 2>/dev/null || true
	@cd tests/e2e && go get github.com/stretchr/testify/assert github.com/stretchr/testify/require
	@echo "Dependencies installed!"

proto-generate: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@cd student-cafe-protos && make generate
	@echo "Protobuf code generated!"

test-unit: ## Run unit tests for all services
	@echo "Running unit tests..."
	@echo "\n=== User Service Unit Tests ==="
	@cd user-service && go test -v ./grpc/...
	@echo "\n=== Menu Service Unit Tests ==="
	@cd menu-service && go test -v ./grpc/...
	@echo "\n=== Order Service Unit Tests ==="
	@cd order-service && go test -v ./grpc/...
	@echo "\nAll unit tests completed!"

test-unit-user: ## Run user service unit tests only
	@echo "Running user service unit tests..."
	@cd user-service && go test -v ./grpc/...

test-unit-menu: ## Run menu service unit tests only
	@echo "Running menu service unit tests..."
	@cd menu-service && go test -v ./grpc/...

test-unit-order: ## Run order service unit tests only
	@echo "Running order service unit tests..."
	@cd order-service && go test -v ./grpc/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@cd tests/integration && go test -v -timeout 30s
	@echo "Integration tests completed!"

test-e2e: ## Run end-to-end tests (requires services to be running)
	@echo "Running end-to-end tests..."
	@echo "Note: Ensure services are running with 'make docker-up' first"
	@cd tests/e2e && go test -v -timeout 60s
	@echo "End-to-end tests completed!"

test-e2e-docker: docker-up ## Run end-to-end tests after starting services
	@echo "Waiting for services to be ready..."
	@sleep 10
	@$(MAKE) test-e2e
	@$(MAKE) docker-down

test: test-unit test-integration ## Run unit and integration tests

test-all: test-unit test-integration test-e2e-docker ## Run all tests including e2e

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@echo "\n=== User Service Coverage ==="
	@cd user-service && go test -v -coverprofile=coverage.out ./grpc/... && go tool cover -html=coverage.out -o coverage.html
	@echo "\n=== Menu Service Coverage ==="
	@cd menu-service && go test -v -coverprofile=coverage.out ./grpc/... && go tool cover -html=coverage.out -o coverage.html
	@echo "\n=== Order Service Coverage ==="
	@cd order-service && go test -v -coverprofile=coverage.out ./grpc/... && go tool cover -html=coverage.out -o coverage.html
	@echo "\nCoverage reports generated!"
	@echo "User Service: user-service/coverage.html"
	@echo "Menu Service: menu-service/coverage.html"
	@echo "Order Service: order-service/coverage.html"

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker compose build
	@echo "Docker images built!"

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	@docker compose up -d
	@echo "Services started!"
	@echo "API Gateway: http://localhost:8080"
	@docker compose ps

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker compose down
	@echo "Services stopped!"

docker-logs: ## Show logs from all services
	@docker compose logs -f

docker-logs-user: ## Show logs from user service
	@docker compose logs -f user-service

docker-logs-menu: ## Show logs from menu service
	@docker compose logs -f menu-service

docker-logs-order: ## Show logs from order service
	@docker compose logs -f order-service

docker-logs-gateway: ## Show logs from API gateway
	@docker compose logs -f api-gateway

clean: ## Clean up test artifacts and Docker volumes
	@echo "Cleaning up..."
	@find . -name "coverage.out" -delete
	@find . -name "coverage.html" -delete
	@docker compose down -v
	@echo "Cleanup completed!"

dev-setup: install-deps proto-generate docker-build ## Complete development setup
	@echo "Development setup completed!"
	@echo "Run 'make docker-up' to start services"
	@echo "Run 'make test' to run tests"

# Continuous Integration targets
ci-test: proto-generate test-unit test-integration ## Run tests suitable for CI
	@echo "CI tests completed successfully!"

ci-full: proto-generate test-unit test-integration docker-build test-e2e-docker ## Full CI pipeline
	@echo "Full CI pipeline completed successfully!"
