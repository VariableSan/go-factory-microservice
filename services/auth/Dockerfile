# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.work files and all modules
COPY go.work go.work.sum ./
COPY pkg/ pkg/
COPY services/ services/

# Build the auth service
WORKDIR /app/services/auth
RUN go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary
COPY --from=builder /app/services/auth/main .

# Expose ports
EXPOSE 8081 9090

# Run the service
CMD ["./main"]
