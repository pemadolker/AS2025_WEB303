# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy root go.mod & go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy everything (root context)
COPY . .

# Build the greeter-service binary
WORKDIR /app/greeter-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50051
CMD ["/app/server"]
