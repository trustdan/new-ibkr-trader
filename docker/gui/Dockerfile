# GUI Service Dockerfile
# Multi-stage build for Svelte + Go backend

# Build stage for Svelte frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /build

# Create placeholder structure
RUN mkdir -p dist && echo "<h1>GUI Not Implemented Yet</h1>" > dist/index.html

# Copy and build frontend if it exists
# Note: Since src/gui is empty, we'll skip the copy and use the placeholder
# When GUI is implemented, uncomment these lines:
# COPY src/gui/package*.json ./
# RUN npm ci --only=production
# COPY src/gui/ ./
# RUN npm run build

# Build stage for Go backend
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files
COPY src/go/go.mod src/go/go.sum* ./src/go/

# Download dependencies
RUN cd src/go && go mod download 2>/dev/null || echo "No go.mod yet"

# Copy backend source
COPY src/go/ ./src/go/

# Build GUI backend binary
RUN if [ -f "src/go/cmd/gui/main.go" ]; then \
        cd src/go && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/gui-server ./cmd/gui; \
    else \
        echo "#!/bin/sh\necho 'GUI server not implemented yet'" > gui-server && chmod +x gui-server; \
    fi

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 ibkr

# Copy artifacts from builders
COPY --from=backend-builder /build/gui-server /usr/local/bin/gui-server
COPY --from=frontend-builder /build/dist /app/static

# Create necessary directories
RUN mkdir -p /app/logs /app/.vibe && \
    chown -R ibkr:ibkr /app

# Set working directory
WORKDIR /app

# Switch to non-root user
USER ibkr

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Expose GUI port
EXPOSE 3000

# Default command
CMD ["gui-server", "--static-dir=/app/static", "--port=3000"]