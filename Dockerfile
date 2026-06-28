# Dockerfile
# Builds a production-style container image for the Go URL shortener API.
# Connects to: compose.yaml, go.mod, src/main.go
# Created: 2026-06-28

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY src ./src
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/url-shortener-api ./src

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /bin/url-shortener-api /usr/local/bin/url-shortener-api

EXPOSE 8080

CMD ["url-shortener-api"]
