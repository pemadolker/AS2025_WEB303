.PHONY: generate clean install-tools lint

# Generate Go code from proto files
generate:
	@echo "Generating Go code from proto files..."
	@mkdir -p gen/go
	@protoc \
		--go_out=gen/go \
		--go_opt=paths=source_relative \
		--go-grpc_out=gen/go \
		--go-grpc_opt=paths=source_relative \
		--proto_path=proto \
		proto/user/v1/user.proto \
		proto/menu/v1/menu.proto \
		proto/order/v1/order.proto
	@echo "✓ Code generation complete!"

# Alternative: use buf for generation (if buf is installed)
generate-buf:
	@echo "Generating Go code using buf..."
	@buf generate
	@echo "✓ Buf generation complete!"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -rf gen/
	@echo "✓ Clean complete!"

# Install required tools
install-tools:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✓ Tools installed!"
	@echo ""
	@echo "Note: Make sure protoc is installed on your system:"
	@echo "  macOS: brew install protobuf"
	@echo "  Linux: apt-get install protobuf-compiler"
	@echo "  Windows: Download from https://github.com/protocolbuffers/protobuf/releases"

# Lint proto files (requires buf)
lint:
	@buf lint

# Format proto files (requires buf)
format:
	@buf format -w

# Run all checks
check: lint
	@go mod tidy
	@go mod verify
	@echo "✓ All checks passed!"
