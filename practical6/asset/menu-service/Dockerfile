FROM golang:1.24-alpine AS builder
WORKDIR /build

# Copy proto module first (needed for go mod download)
COPY student-cafe-protos student-cafe-protos

# Copy service files
WORKDIR /build/app
COPY menu-service/go.mod menu-service/go.sum ./
RUN go mod download
COPY menu-service/ .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /menu-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /menu-service /menu-service

# Expose only gRPC port
EXPOSE 9092

CMD ["/menu-service"]
