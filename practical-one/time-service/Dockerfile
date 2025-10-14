# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum from root
COPY go.mod go.sum ./
RUN go mod download

# Copy everything (root context)
COPY . .

# Build the time-service binary
WORKDIR /app/time-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50052
CMD ["/app/server"]
