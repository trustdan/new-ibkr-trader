FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY src/go/go.mod src/go/go.sum ./

# Download dependencies (will fail gracefully if go.sum is empty)
RUN go mod download || true

# Copy source code
COPY src/go/ .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scanner cmd/scanner/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/scanner .

# Create directories for logs
RUN mkdir -p /app/logs

# Environment variables
ENV SCANNER_PORT=8081
ENV PYTHON_SERVICE_URL=http://python-ibkr:8080
ENV LOG_LEVEL=INFO
ENV CACHE_TTL=300

# Expose scanner port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

CMD ["./scanner"]