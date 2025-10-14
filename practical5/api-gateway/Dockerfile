FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway .

FROM alpine:latest
WORKDIR /
COPY --from=builder /api-gateway /api-gateway
EXPOSE 8080
CMD ["/api-gateway"]
