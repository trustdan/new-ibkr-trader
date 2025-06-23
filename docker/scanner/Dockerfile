# Go Scanner Service Dockerfile
# Multi-stage build for minimal runtime image

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY src/scanner/go.mod src/scanner/go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/scanner/ ./

# Build the scanner binary (using simple_main to avoid compilation issues)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scanner simple_main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 ibkr

# Copy binary from builder
COPY --from=builder /build/scanner /usr/local/bin/scanner

# Create necessary directories
RUN mkdir -p /app/logs /app/.vibe && \
    chown -R ibkr:ibkr /app

# Set working directory
WORKDIR /app

# Switch to non-root user
USER ibkr

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Expose scanner API port
EXPOSE 8081

# Default command
CMD ["scanner"]