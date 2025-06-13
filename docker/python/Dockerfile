# Python IBKR Service Dockerfile
# Multi-stage build for optimization

# Build stage
FROM python:3.11-slim as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --user --no-cache-dir -r requirements.txt

# Runtime stage
FROM python:3.11-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    netcat-traditional \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash ibkr

# Copy Python dependencies from builder
COPY --from=builder /root/.local /home/ibkr/.local

# Set working directory
WORKDIR /app

# Copy application code
COPY src/python/ ./src/python/
COPY scripts/ ./scripts/

# Set Python path
ENV PYTHONPATH=/app:$PYTHONPATH
ENV PATH=/home/ibkr/.local/bin:$PATH

# Create necessary directories
RUN mkdir -p logs .vibe && \
    chown -R ibkr:ibkr /app

# Switch to non-root user
USER ibkr

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import sys; sys.exit(0)"

# Default command (will be overridden by docker-compose)
CMD ["python", "-m", "src.python.ibkr_connector"]