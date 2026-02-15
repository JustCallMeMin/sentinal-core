# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install build essentials
RUN apk add --no-cache git make

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application using Makefile
RUN make build

# Stage 2: Run the application
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user for security
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/sentinal-api .
COPY --from=builder /app/.env.example .env

# Use the non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./sentinal-api"]
