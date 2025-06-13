# Scanner Configuration & API Documentation

## Overview

This document provides comprehensive documentation for configuring and integrating with the IBKR Scanner service, including REST API endpoints, WebSocket protocol, preset management, and configuration best practices.

## REST API Endpoints

### Base URL
```
http://localhost:8081/api/v1
```

### Authentication
```http
Authorization: Bearer <jwt-token>
X-API-Key: <api-key>
```

### Endpoints Reference

#### Health & Status

##### GET /health
Check scanner service health

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600,
  "scanner": {
    "active": true,
    "scansRunning": 2,
    "lastScanTime": "2025-01-13T10:30:00Z"
  },
  "dependencies": {
    "pythonService": "connected",
    "cache": "healthy",
    "database": "connected"
  }
}
```

##### GET /metrics
Prometheus-compatible metrics endpoint

**Response:**
```
# HELP scanner_scans_total Total number of scans performed
# TYPE scanner_scans_total counter
scanner_scans_total 12543

# HELP scanner_scan_duration_seconds Scan duration in seconds
# TYPE scanner_scan_duration_seconds histogram
scanner_scan_duration_seconds_bucket{le="0.05"} 9423
scanner_scan_duration_seconds_bucket{le="0.1"} 11234
scanner_scan_duration_seconds_bucket{le="0.5"} 12500
```

#### Scanner Control

##### POST /scan/start
Start scanning with specified configuration

**Request Body:**
```json
{
  "symbols": ["SPY", "QQQ", "IWM"],
  "filters": {
    "dte": {
      "min": 30,
      "max": 60
    },
    "delta": {
      "min": 0.25,
      "max": 0.35
    },
    "liquidity": {
      "minVolume": 50,
      "minOpenInterest": 100,
      "maxBidAskSpread": 0.10
    }
  },
  "scanInterval": 60,
  "maxResults": 100,
  "sortBy": "pop",
  "sortOrder": "desc"
}
```

**Response:**
```json
{
  "scanId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "started",
  "message": "Scanner started successfully",
  "websocketUrl": "/ws/scan/550e8400-e29b-41d4-a716-446655440000"
}
```

##### DELETE /scan/stop/{scanId}
Stop an active scan

**Response:**
```json
{
  "scanId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "stopped",
  "message": "Scanner stopped successfully",
  "statistics": {
    "totalScans": 45,
    "totalResults": 234,
    "runtime": 2700
  }
}
```

##### GET /scan/status/{scanId}
Get current scan status

**Response:**
```json
{
  "scanId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "startTime": "2025-01-13T10:00:00Z",
  "lastScanTime": "2025-01-13T10:29:45Z",
  "nextScanTime": "2025-01-13T10:30:45Z",
  "statistics": {
    "scansCompleted": 30,
    "averageScanTime": 87,
    "totalResultsFound": 156,
    "errorsEncountered": 0
  },
  "currentFilters": {
    "dte": {"min": 30, "max": 60},
    "delta": {"min": 0.25, "max": 0.35}
  }
}
```

#### Filter Management

##### POST /scan/filters/{scanId}
Update filters for running scan

**Request Body:**
```json
{
  "filters": {
    "delta": {
      "min": 0.30,
      "max": 0.40
    },
    "iv_percentile": {
      "min": 70,
      "max": 100,
      "lookbackDays": 252
    }
  }
}
```

**Response:**
```json
{
  "scanId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "filters_updated",
  "message": "Filters updated successfully",
  "appliedFilters": {
    "delta": {"min": 0.30, "max": 0.40},
    "iv_percentile": {"min": 70, "max": 100, "lookbackDays": 252}
  }
}
```

##### GET /scan/filters/{scanId}
Get current active filters

**Response:**
```json
{
  "scanId": "550e8400-e29b-41d4-a716-446655440000",
  "activeFilters": {
    "dte": {"min": 30, "max": 60, "enabled": true},
    "delta": {"min": 0.30, "max": 0.40, "enabled": true},
    "liquidity": {"minVolume": 50, "enabled": true},
    "greeks": {"maxGamma": 0.05, "enabled": false}
  },
  "filterStats": {
    "dte": {"avgSelectivity": 0.3, "avgExecutionTime": 0.5},
    "delta": {"avgSelectivity": 0.5, "avgExecutionTime": 0.8},
    "liquidity": {"avgSelectivity": 0.7, "avgExecutionTime": 1.2}
  }
}
```

#### Results Retrieval

##### GET /scan/results/{scanId}
Get latest scan results (polling endpoint)

**Query Parameters:**
- `limit` (int): Maximum results to return (default: 100)
- `offset` (int): Pagination offset (default: 0)
- `since` (timestamp): Only results after this time

**Response:**
```json
{
  "scanId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-13T10:30:00Z",
  "results": [
    {
      "id": "result-001",
      "symbol": "SPY",
      "spreadType": "call_debit",
      "expiry": "2025-02-21",
      "longStrike": 450,
      "shortStrike": 455,
      "metrics": {
        "maxProfit": 325,
        "maxLoss": 175,
        "pop": 0.68,
        "debitPaid": 1.75,
        "delta": 0.15,
        "theta": -0.08,
        "iv": 0.18,
        "ivPercentile": 82
      },
      "liquidity": {
        "longVolume": 1250,
        "longOI": 5400,
        "shortVolume": 890,
        "shortOI": 3200,
        "bidAskSpread": 0.05
      },
      "score": 8.5
    }
  ],
  "pagination": {
    "total": 156,
    "limit": 100,
    "offset": 0,
    "hasMore": true
  }
}
```

#### Preset Management

##### GET /filters/presets
Get all saved filter presets

**Response:**
```json
{
  "presets": [
    {
      "id": "conservative-credit",
      "name": "Conservative Credit Spreads",
      "description": "High probability credit spreads in high IV",
      "filters": {
        "dte": {"min": 30, "max": 60},
        "delta": {"min": 0.20, "max": 0.30},
        "iv_percentile": {"min": 70, "max": 100}
      },
      "tags": ["credit", "conservative", "high-iv"],
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-10T15:30:00Z"
    }
  ]
}
```

##### POST /filters/presets
Save a new filter preset

**Request Body:**
```json
{
  "name": "Weekly Income",
  "description": "Weekly credit spreads for income",
  "filters": {
    "dte": {"min": 7, "max": 14},
    "delta": {"min": 0.15, "max": 0.25},
    "liquidity": {"minVolume": 100}
  },
  "tags": ["weekly", "income", "credit"]
}
```

**Response:**
```json
{
  "id": "weekly-income-001",
  "message": "Preset saved successfully",
  "preset": {
    "id": "weekly-income-001",
    "name": "Weekly Income",
    "filters": {...}
  }
}
```

##### PUT /filters/presets/{presetId}
Update existing preset

##### DELETE /filters/presets/{presetId}
Delete a preset

## WebSocket Protocol

### Connection

```javascript
const ws = new WebSocket('ws://localhost:8081/ws/scan/{scanId}');

ws.onopen = () => {
  console.log('Connected to scanner stream');
  
  // Subscribe to specific events
  ws.send(JSON.stringify({
    type: 'subscribe',
    events: ['results', 'status', 'errors']
  }));
};
```

### Message Types

#### Client → Server

##### Subscribe
```json
{
  "type": "subscribe",
  "events": ["results", "status", "errors", "metrics"]
}
```

##### Unsubscribe
```json
{
  "type": "unsubscribe",
  "events": ["metrics"]
}
```

##### Ping
```json
{
  "type": "ping",
  "timestamp": 1234567890
}
```

#### Server → Client

##### Result Update
```json
{
  "type": "result",
  "timestamp": "2025-01-13T10:30:00Z",
  "data": {
    "action": "new",
    "result": {
      "id": "result-002",
      "symbol": "QQQ",
      "spreadType": "put_credit",
      "expiry": "2025-02-21",
      "longStrike": 380,
      "shortStrike": 375,
      "metrics": {...}
    }
  }
}
```

##### Status Update
```json
{
  "type": "status",
  "timestamp": "2025-01-13T10:30:00Z",
  "data": {
    "scanId": "550e8400-e29b-41d4-a716-446655440000",
    "status": "running",
    "progress": {
      "symbolsScanned": 2,
      "totalSymbols": 3,
      "contractsProcessed": 5420,
      "resultsFound": 45
    }
  }
}
```

##### Error Notification
```json
{
  "type": "error",
  "timestamp": "2025-01-13T10:30:00Z",
  "data": {
    "code": "RATE_LIMIT",
    "message": "Rate limit exceeded, backing off",
    "severity": "warning",
    "details": {
      "currentRate": 48,
      "maxRate": 45,
      "backoffMs": 5000
    }
  }
}
```

##### Metrics Update
```json
{
  "type": "metrics",
  "timestamp": "2025-01-13T10:30:00Z",
  "data": {
    "scanDuration": 87,
    "filtersApplied": 5,
    "cacheHitRate": 0.82,
    "memoryUsage": 245,
    "cpuUsage": 0.35
  }
}
```

### WebSocket Best Practices

```javascript
class ScannerWebSocket {
  constructor(scanId) {
    this.scanId = scanId;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
  }
  
  connect() {
    this.ws = new WebSocket(`ws://localhost:8081/ws/scan/${this.scanId}`);
    
    this.ws.onopen = () => {
      console.log('Connected');
      this.reconnectAttempts = 0;
      this.subscribe(['results', 'status', 'errors']);
    };
    
    this.ws.onclose = () => {
      console.log('Disconnected');
      this.handleReconnect();
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    this.ws.onmessage = (event) => {
      this.handleMessage(JSON.parse(event.data));
    };
  }
  
  handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`Reconnecting in ${delay}ms...`);
      setTimeout(() => this.connect(), delay);
    }
  }
  
  subscribe(events) {
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'subscribe',
        events: events
      }));
    }
  }
  
  handleMessage(message) {
    switch (message.type) {
      case 'result':
        this.onResult(message.data);
        break;
      case 'status':
        this.onStatus(message.data);
        break;
      case 'error':
        this.onError(message.data);
        break;
      case 'metrics':
        this.onMetrics(message.data);
        break;
    }
  }
}
```

## Configuration Files

### Scanner Configuration (config.yaml)

```yaml
scanner:
  # Core settings
  workers: 8
  maxConcurrentScans: 10
  defaultScanInterval: 60
  
  # Performance
  cache:
    enabled: true
    size: 10000
    ttl: 300
    evictionPolicy: lru
  
  # Request coordination
  coordinator:
    maxConcurrentRequests: 10
    adaptiveBackpressure: true
    healthCheckInterval: 5
    queueThresholds:
      low: 25
      medium: 50
      high: 75
      critical: 100
  
  # Filter defaults
  filters:
    liquidity:
      minVolume: 10
      minOpenInterest: 50
    dte:
      min: 7
      max: 180
    
  # Result management
  results:
    maxPerScan: 500
    sortBy: score
    sortOrder: desc
    
  # API settings
  api:
    port: 8081
    timeout: 30
    maxRequestSize: 1048576
    cors:
      enabled: true
      origins: ["http://localhost:3000"]
    
  # WebSocket
  websocket:
    pingInterval: 30
    maxConnections: 100
    messageQueueSize: 1000
    
  # Monitoring
  metrics:
    enabled: true
    port: 9091
    path: /metrics
```

### Environment Variables

```bash
# Scanner Service
SCANNER_PORT=8081
SCANNER_WORKERS=8
SCANNER_CACHE_SIZE=10000
SCANNER_LOG_LEVEL=info

# Python Service Connection
PYTHON_SERVICE_URL=http://python-ibkr:8080
PYTHON_SERVICE_TIMEOUT=30
PYTHON_HEALTH_CHECK_INTERVAL=5

# Performance Tuning
SCANNER_MAX_CONCURRENT=10
SCANNER_QUEUE_SIZE=1000
SCANNER_BATCH_SIZE=100

# Security
JWT_SECRET=your-secret-key
API_KEY_HEADER=X-API-Key
CORS_ENABLED=true
CORS_ORIGINS=http://localhost:3000

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9091
TRACE_ENABLED=false
```

### Docker Configuration

```dockerfile
# docker/go-scanner/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o scanner cmd/scanner/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/scanner .
COPY --from=builder /app/config ./config

# Configuration volume
VOLUME ["/config"]

# Expose ports
EXPOSE 8081 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:8081/health || exit 1

CMD ["./scanner", "-config", "/config/config.yaml"]
```

## Configuration Best Practices

### 1. Performance Tuning

```yaml
# For high-throughput environments
scanner:
  workers: 16  # 2x CPU cores
  cache:
    size: 50000  # Larger cache
    ttl: 600     # Longer TTL
  coordinator:
    maxConcurrentRequests: 20
  results:
    maxPerScan: 1000
```

### 2. Resource Constraints

```yaml
# For limited resources
scanner:
  workers: 4
  cache:
    size: 5000
    ttl: 120
  coordinator:
    maxConcurrentRequests: 5
  results:
    maxPerScan: 100
```

### 3. Development Configuration

```yaml
# For development/testing
scanner:
  workers: 2
  cache:
    enabled: false  # Disable caching
  api:
    cors:
      enabled: true
      origins: ["*"]  # Allow all origins
  metrics:
    enabled: true
  trace:
    enabled: true   # Enable tracing
```

### 4. Production Configuration

```yaml
# For production
scanner:
  workers: ${SCANNER_WORKERS:-8}
  cache:
    enabled: true
    size: ${SCANNER_CACHE_SIZE:-10000}
  api:
    timeout: 30
    cors:
      enabled: true
      origins: ${CORS_ORIGINS}
  metrics:
    enabled: true
  logging:
    level: info
    format: json
```

## Integration Examples

### Python Client

```python
import aiohttp
import asyncio
import json

class ScannerClient:
    def __init__(self, base_url="http://localhost:8081"):
        self.base_url = base_url
        self.session = None
        
    async def __aenter__(self):
        self.session = aiohttp.ClientSession()
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.session.close()
        
    async def start_scan(self, config):
        """Start a new scan with configuration"""
        async with self.session.post(
            f"{self.base_url}/api/v1/scan/start",
            json=config
        ) as resp:
            return await resp.json()
            
    async def get_results(self, scan_id, limit=100):
        """Get latest scan results"""
        async with self.session.get(
            f"{self.base_url}/api/v1/scan/results/{scan_id}",
            params={"limit": limit}
        ) as resp:
            return await resp.json()
            
    async def stream_results(self, scan_id):
        """Stream results via WebSocket"""
        ws_url = f"ws://localhost:8081/ws/scan/{scan_id}"
        
        async with self.session.ws_connect(ws_url) as ws:
            # Subscribe to results
            await ws.send_json({
                "type": "subscribe",
                "events": ["results"]
            })
            
            async for msg in ws:
                if msg.type == aiohttp.WSMsgType.TEXT:
                    data = json.loads(msg.data)
                    if data["type"] == "result":
                        yield data["data"]
                elif msg.type == aiohttp.WSMsgType.ERROR:
                    break
```

### JavaScript/TypeScript Client

```typescript
interface ScanConfig {
  symbols: string[];
  filters: Record<string, any>;
  scanInterval?: number;
  maxResults?: number;
}

class ScannerAPI {
  private baseUrl: string;
  private apiKey: string;
  
  constructor(baseUrl = 'http://localhost:8081', apiKey: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }
  
  async startScan(config: ScanConfig): Promise<ScanResponse> {
    const response = await fetch(`${this.baseUrl}/api/v1/scan/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(config)
    });
    
    if (!response.ok) {
      throw new Error(`Scanner error: ${response.statusText}`);
    }
    
    return response.json();
  }
  
  streamResults(scanId: string, onResult: (result: any) => void): WebSocket {
    const ws = new WebSocket(`ws://localhost:8081/ws/scan/${scanId}`);
    
    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: 'subscribe',
        events: ['results', 'status', 'errors']
      }));
    };
    
    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'result') {
        onResult(message.data);
      }
    };
    
    return ws;
  }
}
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check scanner service is running
   - Verify port configuration
   - Check firewall rules

2. **WebSocket Disconnects**
   - Implement reconnection logic
   - Check for rate limiting
   - Monitor network stability

3. **No Results**
   - Verify filter configuration
   - Check Python service connection
   - Review scanner logs

4. **High Latency**
   - Monitor cache hit rate
   - Check concurrent scan count
   - Review filter complexity

### Debug Endpoints

```http
# Get detailed scanner state
GET /debug/state

# Force cache clear
POST /debug/cache/clear

# Get filter statistics
GET /debug/filters/stats

# Trigger manual GC
POST /debug/gc
```