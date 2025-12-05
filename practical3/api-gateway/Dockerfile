FROM golang:1.24.2-alpine AS builder
WORKDIR /app
COPY api-gateway/go.mod ./
COPY api-gateway/go.sum ./
COPY proto/ ./proto/
RUN go mod download
COPY api-gateway/ ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["/app/server"]
