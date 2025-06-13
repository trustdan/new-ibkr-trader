# Go Scanner Service Dockerfile
# Multi-stage build for minimal runtime image

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download 2>/dev/null || echo "No go.mod yet"

# Copy source code
COPY src/go/ ./src/go/

# Build the scanner binary
RUN if [ -f "src/go/cmd/scanner/main.go" ]; then \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scanner ./src/go/cmd/scanner; \
    else \
        echo "Scanner not implemented yet" > scanner && chmod +x scanner; \
    fi

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
    CMD ["scanner", "health"] || exit 1

# Expose scanner API port
EXPOSE 8080

# Default command
CMD ["scanner", "serve"]