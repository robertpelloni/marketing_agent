# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/marketing_agent ./cmd/marketing_agent

# Final stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/marketing_agent .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./marketing_agent"]
