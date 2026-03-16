# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine AS builder

# Install build tools and security updates
RUN apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

# Create non-root user for building
RUN adduser -D -g '' appuser

WORKDIR /src

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/server ./main.go

# Runtime stage - use Alpine minimal (more reliable than distroless registries)
FROM alpine:latest

# Install only CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# Create non-root user
RUN adduser -D -g '' -s /bin/sh appuser

# Copy the binary
COPY --from=builder /app/server /app/server
COPY --from=builder /src/keys /app/keys

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/health || exit 1

# Run as non-root user
USER appuser

# Start the application
ENTRYPOINT ["/app/server"]
