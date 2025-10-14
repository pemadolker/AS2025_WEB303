# Practical 5: Refactoring Monolith to Microservices - Reference Implementation

This directory contains the complete reference implementation for Practical 5, demonstrating the journey from a monolithic application to microservices architecture.

## Architecture Overview

### Monolithic Application
- **Location:** `student-cafe-monolith/`
- **Port:** 8090
- **Database:** PostgreSQL (port 5432)
- **Features:** All-in-one application with users, menu, and orders

### Microservices Architecture

**Services:**
1. **user-service** (port 8081)
   - Database: user_db (port 5434)
   - Manages user profiles and authentication

2. **menu-service** (port 8082)
   - Database: menu_db (port 5433)
   - Manages food catalog and pricing

3. **order-service** (port 8083)
   - Database: order_db (port 5435)
   - Manages orders and inter-service communication
   - Calls user-service and menu-service to validate orders

4. **api-gateway** (port 8080)
   - Routes requests to appropriate microservices
   - Single entry point for all client requests

## Quick Start

### Run Everything

```bash
docker-compose up --build
```

This will start:
- 4 PostgreSQL databases
- 1 Monolith application
- 3 Microservices
- 1 API Gateway

### Test the Monolith

```bash
# Create menu item
curl -X POST http://localhost:8090/api/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Coffee", "description": "Hot coffee", "price": 2.50}'

# Create user
curl -X POST http://localhost:8090/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Create order
curl -X POST http://localhost:8090/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 2}]}'

# Get orders
curl http://localhost:8090/api/orders
```

### Test the Microservices (via API Gateway)

```bash
# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'

# Create menu item
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{"name": "Tea", "description": "Hot tea", "price": 1.50}'

# Get menu
curl http://localhost:8080/api/menu

# Create order (demonstrates inter-service communication)
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "items": [{"menu_item_id": 1, "quantity": 1}]}'

# Get orders
curl http://localhost:8080/api/orders
```

### Test Individual Microservices

```bash
# User service directly
curl http://localhost:8081/users/1

# Menu service directly
curl http://localhost:8082/menu

# Order service directly
curl http://localhost:8083/orders
```

## Directory Structure

```
practical5/
├── student-cafe-monolith/      # Phase 1: Monolithic application
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── main.go
│   ├── Dockerfile
│   └── docker-compose.yml      # Standalone monolith compose
│
├── menu-service/               # Phase 2: Extracted menu service
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── main.go
│   └── Dockerfile
│
├── user-service/               # Phase 3: Extracted user service
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── main.go
│   └── Dockerfile
│
├── order-service/              # Phase 4: Extracted order service (with inter-service calls)
│   ├── models/
│   ├── handlers/
│   ├── database/
│   ├── main.go
│   └── Dockerfile
│
├── api-gateway/                # Phase 5: API Gateway
│   ├── main.go
│   └── Dockerfile
│
├── docker-compose.yml          # Complete system orchestration
└── README.md                   # This file
```

## Key Learning Points

### Monolith Characteristics
- Single codebase, single deployment
- Shared database with all tables
- Tight coupling between features
- Simple to run but hard to scale independently

### Microservices Characteristics
- Independent codebases and deployments
- Database-per-service pattern
- Loose coupling via HTTP APIs
- Can scale services independently
- Inter-service communication required

### Inter-Service Communication

The order-service demonstrates HTTP-based inter-service communication:

```go
// Validate user exists by calling user-service
userResp, err := http.Get(fmt.Sprintf("%s/users/%d", userServiceURL, req.UserID))

// Validate menu items by calling menu-service
menuResp, err := http.Get(fmt.Sprintf("%s/menu/%d", menuServiceURL, item.MenuItemID))
```

### API Gateway Pattern

The gateway provides:
- Single entry point (port 8080)
- Request routing to appropriate services
- Path transformation (/api/users -> /users)
- Future: Authentication, rate limiting, logging

## Troubleshooting

### Ports Already in Use
```bash
# Stop all containers
docker-compose down

# Remove orphan containers
docker-compose down --remove-orphans

# Check what's using a port
lsof -i :8080
```

### Database Connection Issues
```bash
# Check if databases are running
docker-compose ps

# View database logs
docker-compose logs menu-db
docker-compose logs user-db
docker-compose logs order-db
```

### Service Can't Reach Another Service
```bash
# Check service logs
docker-compose logs order-service
docker-compose logs user-service
docker-compose logs menu-service

# Verify services are on same network
docker network inspect practical5_default
```

### Rebuild Specific Service
```bash
# Rebuild and restart just one service
docker-compose up --build user-service

# Rebuild without cache
docker-compose build --no-cache user-service
```

## Clean Up

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (deletes all data)
docker-compose down -v

# Remove all images
docker-compose down --rmi all
```

## Next Steps

This reference implementation covers Phases 1-5 of the practical:
- ✅ Phase 1: Monolithic application
- ✅ Phase 2: Extract menu-service
- ✅ Phase 3: Extract user-service
- ✅ Phase 4: Extract order-service (with inter-service communication)
- ✅ Phase 5: Add API Gateway

**Future enhancements (Phase 6):**
- Add Consul for dynamic service discovery
- Replace hardcoded URLs with Consul lookups
- Add health checks to services
- Implement service resilience patterns

## Architecture Comparison

| Aspect | Monolith | Microservices |
|--------|----------|---------------|
| Deployment | Single unit | Independent services |
| Database | Shared (1 DB) | Per-service (4 DBs) |
| Scaling | All-or-nothing | Service-specific |
| Development | Simple setup | Requires orchestration |
| Communication | In-process | HTTP/REST |
| Failure Impact | Entire app down | Isolated to service |
| Team Structure | Single team | Team per service |

## Resources

- Student practical guide: `practical5.md`
- Implementation plan: `thoughts/shared/plans/2025-10-08-practical5-monolith-to-microservices.md`
- Related practicals:
  - Practical 2: Consul + API Gateway basics
  - Practical 4: Kubernetes deployment
