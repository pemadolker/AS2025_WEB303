FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service .

FROM alpine:latest
WORKDIR /
COPY --from=builder /user-service /user-service
EXPOSE 8081
CMD ["/user-service"]
