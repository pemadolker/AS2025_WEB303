# Student Cafe Proto Repository

This is a centralized, versioned Protocol Buffer repository for the Student Cafe microservices application. It contains all service definitions and generates Go code that is imported by individual services.

## Directory Structure

```
student-cafe-protos/
├── proto/                    # Proto definition files
│   ├── user/v1/
│   │   └── user.proto       # User service definitions
│   ├── menu/v1/
│   │   └── menu.proto       # Menu service definitions
│   └── order/v1/
│       └── order.proto      # Order service definitions
├── gen/go/                  # Generated Go code
│   ├── user/v1/
│   ├── menu/v1/
│   └── order/v1/
├── buf.yaml                 # Buf configuration
├── buf.gen.yaml            # Buf generation configuration
├── Makefile                # Build automation
└── go.mod                  # Go module definition
```

## Prerequisites

1. **Install Protocol Buffer Compiler (protoc)**:
   - macOS: `brew install protobuf`
   - Linux: `sudo apt-get install protobuf-compiler`
   - Windows: Download from [GitHub releases](https://github.com/protocolbuffers/protobuf/releases)

2. **Install Go plugins**:
   ```bash
   make install-tools
   ```

   This installs:
   - `protoc-gen-go` (for generating Go structs)
   - `protoc-gen-go-grpc` (for generating gRPC service code)

## Usage

### Generating Code

After modifying any `.proto` files:

```bash
# Generate Go code
make generate

# Or clean and regenerate
make clean && make generate
```

The generated code will be placed in `gen/go/` directory.

### Importing in Services

In your service's `go.mod`, add the proto module as a dependency. There are two ways to do this:

#### Option 1: Local Development (Recommended for this practical)

Use Go's `replace` directive to reference the local module:

```go
// In your service's go.mod
module user-service

go 1.23

require (
    github.com/douglasswm/student-cafe-protos v0.0.0
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)

// Replace with local path
replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
```

Then in your code:

```go
import (
    userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
    menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
)
```

#### Option 2: Published Module (Production)

If you push the proto module to GitHub and tag it:

```bash
cd student-cafe-protos
git add .
git commit -m "Release v1.0.0"
git tag v1.0.0
git push origin v1.0.0
```

Then services can import directly:

```bash
go get github.com/douglasswm/student-cafe-protos@v1.0.0
```

### Versioning Strategy

We use semantic versioning for the proto repository:

- **v1.0.0**: Initial release
- **v1.0.x**: Backward-compatible bug fixes
- **v1.x.0**: Backward-compatible new features
- **v2.0.0**: Breaking changes

When making changes:

1. Modify proto files
2. Run `make generate`
3. Test with all services
4. Commit and tag with new version
5. Update services to use new version

## Service Definitions

### User Service (`user/v1/user.proto`)

Handles user management operations:
- `CreateUser`: Register a new user
- `GetUser`: Retrieve user by ID
- `GetUsers`: List all users

### Menu Service (`menu/v1/menu.proto`)

Manages menu items:
- `GetMenuItem`: Get a specific menu item
- `GetMenu`: List all menu items
- `CreateMenuItem`: Add new menu item

### Order Service (`order/v1/order.proto`)

Handles order operations:
- `CreateOrder`: Create a new order
- `GetOrders`: List all orders
- `GetOrder`: Get order by ID

## Common Tasks

### Adding a New RPC Method

1. Edit the relevant `.proto` file
2. Add the new method to the service definition
3. Define request/response messages
4. Run `make generate`
5. Update affected services
6. Bump version (e.g., v1.1.0)

### Adding a New Service

1. Create new directory: `proto/newservice/v1/`
2. Create proto file: `proto/newservice/v1/newservice.proto`
3. Update Makefile to include new proto file
4. Run `make generate`
5. Tag as new version

## Troubleshooting

### Import Path Issues

Make sure your `go.mod` has the correct `replace` directive:
```go
replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
```

The relative path should point from the service directory to the proto directory.

### Generation Errors

If proto generation fails:
```bash
# Reinstall tools
make install-tools

# Verify protoc is installed
protoc --version

# Check PATH includes Go bin directory
echo $PATH | grep $(go env GOPATH)/bin
```

### Build Errors in Docker

When building Docker images, ensure the proto module is accessible:

```dockerfile
# Copy proto module first
COPY ../student-cafe-protos /student-cafe-protos
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
```

Or use multi-stage builds with the proto module copied in the build stage.

## Benefits of This Approach

1. **Single Source of Truth**: All service contracts defined in one place
2. **Versioning**: Changes are tracked and services can pin to specific versions
3. **No Code Duplication**: Generated code is reused across services
4. **Type Safety**: Protocol buffers provide strong typing
5. **Breaking Change Detection**: Buf can detect backward-incompatible changes
6. **Easy Updates**: Update proto, regenerate, and all services use the new version

## Development Workflow

1. Make changes to proto files
2. Run `make generate`
3. Test locally with `replace` directive
4. Once stable, tag version and push
5. Update services to use new tagged version
6. Deploy services independently

This approach solves the common proto synchronization and build issues found in previous practicals by centralizing definitions and using proper Go module versioning.
