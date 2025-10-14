FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /menu-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /menu-service /menu-service
EXPOSE 8082
CMD ["/menu-service"]
