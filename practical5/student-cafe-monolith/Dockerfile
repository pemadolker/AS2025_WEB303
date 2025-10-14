FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /monolith .

FROM alpine:latest
WORKDIR /
COPY --from=builder /monolith /monolith
EXPOSE 8080
CMD ["/monolith"]
