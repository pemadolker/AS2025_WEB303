FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /order-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /order-service /order-service
EXPOSE 8083
CMD ["/order-service"]
