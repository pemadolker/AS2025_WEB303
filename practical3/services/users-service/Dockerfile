FROM golang:1.24.2-alpine AS builder
WORKDIR /app
COPY services/users-service/go.mod ./
COPY services/users-service/go.sum ./
COPY proto/ ./proto/
RUN go mod download
COPY services/users-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 50051
CMD ["/app/server"]
