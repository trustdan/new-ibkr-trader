# Python-Go Integration Architecture

## Overview

The IBKR Spread Automation system uses a microservices architecture with Python handling IBKR API communication and Go providing high-performance options scanning. This document describes how these services integrate.

## Architecture Diagram

```
┌─────────────────────┐         ┌─────────────────────┐
│   Python Service    │         │    Go Scanner       │
│                     │         │                     │
│ ┌─────────────────┐ │ HTTP    │ ┌─────────────────┐│
│ │ IBKR Connection │ │ ──────> │ │   Scanner API   ││
│ └─────────────────┘ │         │ └─────────────────┘│
│                     │         │                     │
│ ┌─────────────────┐ │         │ ┌─────────────────┐│
│ │Scanner Client   │ │ <────── │ │  Filter Engine  ││
│ └─────────────────┘ │ WebSocket│ └─────────────────┘│
│                     │         │                     │
│ ┌─────────────────┐ │         │ ┌─────────────────┐│
│ │  Coordinator    │ │         │ │  Cache Layer    ││
│ └─────────────────┘ │         │ └─────────────────┘│
└─────────────────────┘         └─────────────────────┘
```

## Components

### Python Service Components

#### 1. Scanner Client (`scanner_client.py`)
- HTTP client for Go scanner API
- Handles request/response serialization
- Implements retry logic with exponential backoff
- Manages connection pooling

**Key Features:**
- Async/await support
- Type-safe request/response models
- Automatic retry on rate limiting
- Health check endpoint

#### 2. Scanner Coordinator (`scanner_coordinator.py`)
- Orchestrates data flow between IBKR and scanner
- Manages scan job lifecycle
- Implements caching layer (5-minute TTL)
- Handles concurrent scan limiting

**Key Features:**
- Job queue management
- Cache-first strategy
- Metrics collection
- Event emission

#### 3. Backpressure Handler (`backpressure.py`)
- Multiple rate limiting strategies
- Circuit breaker pattern
- Adaptive throttling
- Request metrics tracking

**Strategies:**
- Token Bucket (default)
- Sliding Window
- Fixed Window
- Adaptive (adjusts based on response times)

### Go Scanner Components

#### 1. Scanner API
- RESTful HTTP endpoints
- WebSocket support for real-time updates
- Request validation
- Response formatting

#### 2. Filter Engine
- Modular filter system
- Concurrent processing
- Score calculation
- Spread generation

#### 3. Cache Layer
- In-memory caching
- TTL-based expiration
- Concurrent-safe access

## Integration Flow

### 1. Scan Request Flow

```python
# Python initiates scan
request = ScanRequest(
    symbol="AAPL",
    filters=[...],
    limit=100
)

# Coordinator checks cache
if cached_result:
    return cached_result

# Submit to scanner via HTTP
response = await scanner_client.scan(request)

# Cache and return results
cache[key] = response
return response
```

### 2. Backpressure Flow

```python
# Acquire rate limit permit
await backpressure.wait_if_needed()

# Execute request
result = await scanner.scan(request)

# Record metrics
backpressure.record_request(
    duration=elapsed_time,
    success=True,
    queue_time=wait_time
)
```

### 3. Error Handling

```python
try:
    spreads = await scanner.scan(request)
except HTTPStatusError as e:
    if e.response.status_code == 429:
        # Rate limited - retry with backoff
        await asyncio.sleep(backoff_time)
        return await retry_scan(request)
    elif e.response.status_code == 503:
        # Service unavailable - circuit breaker
        circuit_breaker.record_failure()
        raise ServiceUnavailableError()
```

## Configuration

### Python Service

```python
# Scanner client configuration
SCANNER_BASE_URL = "http://localhost:8080"
SCANNER_TIMEOUT = 30.0
SCANNER_MAX_RETRIES = 3

# Coordinator configuration
MAX_CONCURRENT_SCANS = 5
SCAN_CACHE_TTL = 300  # 5 minutes

# Backpressure configuration
REQUESTS_PER_SECOND = 10.0
BURST_SIZE = 20
CIRCUIT_BREAKER_THRESHOLD = 5
CIRCUIT_BREAKER_TIMEOUT = 60
```

### Go Scanner

```go
// API configuration
API_PORT = ":8080"
MAX_REQUEST_SIZE = 1MB
REQUEST_TIMEOUT = 30s

// Scanner configuration
MAX_CONCURRENT_SCANS = 10
CACHE_TTL = 5 * time.Minute
```

## Monitoring & Metrics

### Python Metrics

```python
{
    "total_scans": 1523,
    "successful_scans": 1498,
    "failed_scans": 25,
    "cache_hits": 342,
    "average_scan_time": 0.45,
    "active_jobs": 3,
    "queued_jobs": 7,
    "backpressure": {
        "current_qps": 8.5,
        "average_response_time": 0.42,
        "circuit_breaker_open": false,
        "adaptive_rate": 9.2
    }
}
```

### Go Scanner Metrics

```json
{
    "total_requests": 1523,
    "cache_hits": 456,
    "average_filter_time": 0.023,
    "active_scans": 3,
    "error_rate": 0.016
}
```

## Testing

### Unit Tests
- Scanner client HTTP mocking
- Coordinator job lifecycle
- Backpressure strategies
- Circuit breaker behavior

### Integration Tests
- End-to-end scan flow
- Concurrent request handling
- Cache validation
- Error recovery

### Load Tests
```bash
# Test high concurrency
python scripts/test_scanner_integration.py

# Monitor metrics during load
curl http://localhost:8080/metrics
```

## Deployment Considerations

### 1. Service Discovery
- Use environment variables for service URLs
- Consider service mesh for production
- Implement health checks

### 2. Scaling
- Python service scales with IBKR connections
- Go scanner scales horizontally
- Use load balancer for scanner instances

### 3. Resilience
- Circuit breakers prevent cascade failures
- Caching reduces scanner load
- Graceful degradation on scanner unavailability

## Security

### 1. Authentication
- API key authentication (future)
- Service-to-service mTLS (production)

### 2. Rate Limiting
- Per-client rate limits
- Global rate limits
- DDoS protection

### 3. Input Validation
- Request size limits
- Parameter validation
- SQL injection prevention

## Future Enhancements

1. **gRPC Support**
   - Binary protocol for better performance
   - Streaming support for real-time updates
   - Strong typing with protobuf

2. **Service Mesh Integration**
   - Istio/Linkerd support
   - Automatic retries and circuit breaking
   - Distributed tracing

3. **Event Streaming**
   - Kafka/NATS for event bus
   - Async processing pipeline
   - Event sourcing

4. **Advanced Caching**
   - Redis for distributed cache
   - Cache warming strategies
   - Predictive cache invalidation