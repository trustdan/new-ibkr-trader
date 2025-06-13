FROM python:3.11-slim

# Install system dependencies for ib-insync and async performance
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Install Python dependencies with specific versions for stability
RUN pip install --no-cache-dir \
    uvloop==0.19.0 \
    aiohttp==3.9.1 \
    ib-insync==0.9.86 \
    prometheus-client==0.19.0 \
    aiocache[msgpack]==0.12.2 \
    python-json-logger==2.0.7 \
    aiofiles==23.2.1

# Set up async-friendly environment
ENV PYTHONUNBUFFERED=1
ENV PYTHONASYNCIODEBUG=1
ENV IB_ASYNC_MODE=1

# Default environment variables for TWS connection
ENV TWS_HOST=host.docker.internal
ENV TWS_PORT=7497
ENV TWS_CLIENT_ID=1
ENV WATCHDOG_TIMEOUT=60
ENV MAX_SUBSCRIPTIONS=90
ENV ACCOUNT_TYPE=paper
ENV LOG_LEVEL=INFO

# Copy application code (will be mounted in development)
COPY src/python /app

# Create directories for logs and data
RUN mkdir -p /app/logs /app/data

# Health check that doesn't block event loop
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD python -c "import asyncio, aiohttp; asyncio.run(aiohttp.ClientSession().get('http://localhost:8080/health'))" || exit 1

# Expose ports for API and metrics
EXPOSE 8080 9090

# Run with uvloop for maximum async performance
CMD ["python", "-m", "uvloop", "main.py"]