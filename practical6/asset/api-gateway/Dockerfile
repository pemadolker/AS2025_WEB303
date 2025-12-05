FROM golang:1.24-alpine AS builder
WORKDIR /build

# Copy proto module first (required for imports)
COPY student-cafe-protos student-cafe-protos

# Copy gateway files
WORKDIR /build/app
COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download
COPY api-gateway/ .

# Build the gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway .

FROM alpine:latest
WORKDIR /
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8080
CMD ["/api-gateway"]
