FROM golang:1.24-alpine AS builder
WORKDIR /build

# Copy proto module first (needed for go mod download)
COPY student-cafe-protos student-cafe-protos

# Copy service files
WORKDIR /build/app
COPY user-service/go.mod user-service/go.sum ./
RUN go mod download
COPY user-service/ .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /user-service /user-service

# Expose only gRPC port
EXPOSE 9091

CMD ["/user-service"]
