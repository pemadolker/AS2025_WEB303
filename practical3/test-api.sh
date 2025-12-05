#!/bin/bash

# Test script for the microservices API
BASE_URL="http://localhost:8080"

echo "🚀 Testing Microservices API..."
echo "================================"

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 20

# Test 1: Create a user
echo "📝 Creating a user..."
USER_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
     -d '{"name": "Jane Doe", "email": "jane.doe@example.com"}' \
     ${BASE_URL}/api/users)
echo "User created: $USER_RESPONSE"
echo ""

# Test 2: Get the user
echo "👤 Getting user with ID 1..."
curl -s ${BASE_URL}/api/users/1 | jq '.'
echo ""

# Test 3: Create a product
echo "🛍️ Creating a product..."
PRODUCT_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
     -d '{"name": "Laptop", "price": 1200.50}' \
     ${BASE_URL}/api/products)
echo "Product created: $PRODUCT_RESPONSE"
echo ""

# Test 4: Get the product
echo "💻 Getting product with ID 1..."
curl -s ${BASE_URL}/api/products/1 | jq '.'
echo ""

# Test 5: Get combined purchase data
echo "🛒 Getting combined purchase data..."
curl -s ${BASE_URL}/api/purchases/user/1/product/1 | jq '.'
echo ""

echo "✅ API testing complete!"
