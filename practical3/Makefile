# Makefile for Practical 3 - Microservices Project

.PHONY: help build start stop test clean proto logs consul

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build and start all services
	docker-compose up --build -d

start: ## Start all services (assumes images are built)
	docker-compose up -d

stop: ## Stop all services
	docker-compose down

test: ## Run API tests
	./test-api.sh

clean: ## Stop services and remove volumes
	docker-compose down -v

proto: ## Generate Protocol Buffer code
	buf generate

logs: ## Show logs for all services
	docker-compose logs -f

consul: ## Open Consul UI
	open http://localhost:8500

users-logs: ## Show users-service logs
	docker-compose logs -f users-service

products-logs: ## Show products-service logs
	docker-compose logs -f products-service

gateway-logs: ## Show api-gateway logs
	docker-compose logs -f api-gateway

status: ## Show status of all services
	docker-compose ps

restart: ## Restart all services
	docker-compose restart

dev-db: ## Start only databases and consul for local development
	docker-compose up consul users-db products-db -d
