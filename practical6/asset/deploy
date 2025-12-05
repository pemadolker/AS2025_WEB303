#!/bin/bash

# Practical 5A Deployment Script
# This script builds and deploys the Student Cafe microservices with gRPC support

set -e  # Exit on error

echo "===================================="
echo "Student Cafe - Practical 5A Deployment"
echo "gRPC Microservices with Centralized Proto Repository"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Generate Proto Code
echo -e "${BLUE}Step 1: Generating Proto Code${NC}"
cd student-cafe-protos
echo "Cleaning previous generated code..."
rm -rf gen/
echo "Generating Go code from proto files..."
export PATH=$PATH:$(go env GOPATH)/bin
make generate
cd ..
echo -e "${GREEN}✓ Proto code generated successfully${NC}"
echo ""

# Step 2: Stop and remove existing containers
echo -e "${BLUE}Step 2: Cleaning up existing containers${NC}"
docker-compose down -v 2>/dev/null || true
echo -e "${GREEN}✓ Cleanup complete${NC}"
echo ""

# Step 3: Build Docker images
echo -e "${BLUE}Step 3: Building Docker images${NC}"
echo "This may take a few minutes..."
docker-compose build --no-cache
echo -e "${GREEN}✓ Docker images built successfully${NC}"
echo ""

# Step 4: Start services
echo -e "${BLUE}Step 4: Starting all services${NC}"
docker-compose up -d
echo -e "${GREEN}✓ All services started${NC}"
echo ""

# Step 5: Wait for services to be ready
echo -e "${BLUE}Step 5: Waiting for services to be ready${NC}"
echo "Waiting 10 seconds for databases and services to initialize..."
sleep 10
echo -e "${GREEN}✓ Services should be ready${NC}"
echo ""

# Step 6: Check service health
echo -e "${BLUE}Step 6: Checking service health${NC}"
docker-compose ps
echo ""

# Step 7: Display access information
echo -e "${GREEN}===================================="
echo "Deployment Complete!"
echo "====================================${NC}"
echo ""
echo -e "${YELLOW}Service Endpoints:${NC}"
echo ""
echo "HTTP Endpoints (REST):"
echo "  - API Gateway:    http://localhost:8080"
echo "  - User Service:   http://localhost:8081"
echo "  - Menu Service:   http://localhost:8082"
echo "  - Order Service:  http://localhost:8083"
echo ""
echo "gRPC Endpoints (Internal):"
echo "  - User Service:   localhost:9091"
echo "  - Menu Service:   localhost:9092"
echo "  - Order Service:  localhost:9093"
echo ""
echo -e "${YELLOW}Test Commands:${NC}"
echo ""
echo "# Create a menu item"
echo "curl -X POST http://localhost:8080/api/menu \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"name\": \"Coffee\", \"description\": \"Hot coffee\", \"price\": 2.50}'"
echo ""
echo "# Create a user"
echo "curl -X POST http://localhost:8080/api/users \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"name\": \"John Doe\", \"email\": \"john@example.com\", \"is_cafe_owner\": false}'"
echo ""
echo "# Create an order (demonstrates gRPC inter-service communication)"
echo "curl -X POST http://localhost:8080/api/orders \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"user_id\": 1, \"items\": [{\"menu_item_id\": 1, \"quantity\": 2}]}'"
echo ""
echo -e "${YELLOW}View Logs:${NC}"
echo "  docker-compose logs -f [service-name]"
echo ""
echo -e "${YELLOW}Stop Services:${NC}"
echo "  docker-compose down"
echo ""
echo "===================================="
