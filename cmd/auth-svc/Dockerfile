# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-svc ./cmd/auth-svc

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /app/auth-svc .

EXPOSE 8080

CMD ["./auth-svc"]