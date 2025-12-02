# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies for PostgreSQL
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o auth-svc ./cmd/auth-svc

# Final stage
FROM alpine:latest

# Install runtime dependencies for PostgreSQL
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root

COPY --from=builder /app/auth-svc .

EXPOSE 8080

CMD ["./auth-svc"]