#!/bin/bash

# Practical 5B Deployment Script
# Pure gRPC Backend with HTTP→gRPC Gateway

set -e  # Exit on error

echo "========================================"
echo "Student Cafe - Practical 5B Deployment"
echo "Pure gRPC Backend with HTTP Gateway"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
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
echo "Waiting 15 seconds for databases and services to initialize..."
sleep 15
echo -e "${GREEN}✓ Services should be ready${NC}"
echo ""

# Step 6: Check service health
echo -e "${BLUE}Step 6: Checking service health${NC}"
docker-compose ps
echo ""

# Step 7: Verify gRPC architecture
echo -e "${BLUE}Step 7: Verifying gRPC-only architecture${NC}"
echo "Checking gateway gRPC client initialization..."
if docker-compose logs api-gateway | grep -q "gRPC clients initialized"; then
    echo -e "${GREEN}✓ Gateway gRPC clients initialized${NC}"
else
    echo -e "${RED}⚠ Gateway may not have initialized gRPC clients${NC}"
fi

echo "Checking backend services are gRPC-only..."
if docker-compose logs user-service | grep -q "gRPC only"; then
    echo -e "${GREEN}✓ User service running as gRPC-only${NC}"
else
    echo -e "${YELLOW}⚠ User service logs don't show 'gRPC only'${NC}"
fi

if docker-compose logs menu-service | grep -q "gRPC only"; then
    echo -e "${GREEN}✓ Menu service running as gRPC-only${NC}"
else
    echo -e "${YELLOW}⚠ Menu service logs don't show 'gRPC only'${NC}"
fi

if docker-compose logs order-service | grep -q "gRPC only"; then
    echo -e "${GREEN}✓ Order service running as gRPC-only${NC}"
else
    echo -e "${YELLOW}⚠ Order service logs don't show 'gRPC only'${NC}"
fi
echo ""

# Step 8: Display access information
echo -e "${GREEN}========================================"
echo "Deployment Complete!"
echo "========================================${NC}"
echo ""
echo -e "${YELLOW}Architecture:${NC}"
echo "  External: HTTP/REST"
echo "  Internal: Pure gRPC"
echo ""
echo -e "${YELLOW}Service Endpoints:${NC}"
echo ""
echo "External Access (HTTP):"
echo "  - API Gateway:    http://localhost:8080"
echo ""
echo "Internal gRPC Ports (not directly accessible via HTTP):"
echo "  - User Service:   localhost:9091"
echo "  - Menu Service:   localhost:9092"
echo "  - Order Service:  localhost:9093"
echo ""
echo "Database Ports:"
echo "  - User DB:        localhost:5434"
echo "  - Menu DB:        localhost:5433"
echo "  - Order DB:       localhost:5435"
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
echo "# Create an order (demonstrates full gRPC flow)"
echo "curl -X POST http://localhost:8080/api/orders \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"user_id\": 1, \"items\": [{\"menu_item_id\": 1, \"quantity\": 2}]}'"
echo ""
echo -e "${YELLOW}Verify gRPC Architecture:${NC}"
echo ""
echo "# These should FAIL (gRPC ports, not HTTP):"
echo "curl http://localhost:9091"
echo "curl http://localhost:9092"
echo "curl http://localhost:9093"
echo ""
echo "# This should WORK (HTTP gateway):"
echo "curl http://localhost:8080/api/menu"
echo ""
echo -e "${YELLOW}View Logs:${NC}"
echo "  docker-compose logs -f [service-name]"
echo ""
echo "  Useful log checks:"
echo "  docker-compose logs api-gateway | grep 'gRPC clients'"
echo "  docker-compose logs user-service | grep 'gRPC only'"
echo ""
echo -e "${YELLOW}Stop Services:${NC}"
echo "  docker-compose down"
echo ""
echo "========================================"
echo ""
echo -e "${GREEN}Key Changes from Practical 5A:${NC}"
echo "  ✓ Gateway now uses gRPC clients (not HTTP proxy)"
echo "  ✓ Backend services are gRPC-only (HTTP removed)"
echo "  ✓ All internal communication is gRPC"
echo "  ✓ External API still HTTP/REST (backwards compatible)"
echo ""
echo -e "${BLUE}Architecture Achievement: Pure gRPC Backend!${NC}"
echo ""
