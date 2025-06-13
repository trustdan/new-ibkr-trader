# Scanner Core Architecture

## Overview

The IBKR Spread Scanner is a high-performance Go service designed to efficiently scan thousands of option contracts in real-time while respecting TWS API limitations. Built with concurrent processing at its core, the scanner coordinates with the Python IBKR service to maximize throughput without overwhelming the TWS connection.

## Architecture Principles

### 1. Concurrent Processing
- **Goroutine-based parallelism** for filter execution
- **Channel-based communication** for thread-safe data flow
- **Worker pool pattern** for controlled concurrency

### 2. Intelligent Coordination
- **Adaptive rate control** based on Python service health
- **Backpressure mechanisms** to prevent queue overflow
- **Circuit breaker pattern** for fault tolerance

### 3. Performance Optimization
- **In-memory caching** for frequently accessed data
- **Filter chain optimization** based on selectivity
- **Lazy evaluation** for expensive operations

## Core Components

### Scanner Engine

```go
type Scanner struct {
    coordinator  *RequestCoordinator
    filterChain  *FilterChain
    cache        *ResultCache
    metrics      *Metrics
    broadcaster  *Broadcaster
}
```

The scanner engine orchestrates the entire scanning process:

1. **Request Coordination**: Manages communication with Python service
2. **Filter Chain Execution**: Applies filters in optimal order
3. **Result Caching**: Stores recent results for performance
4. **Metrics Collection**: Tracks performance and health
5. **Result Broadcasting**: Streams results to connected clients

### Request Coordinator

```go
type RequestCoordinator struct {
    pythonClient  *PythonAPIClient
    maxConcurrent int
    semaphore     chan struct{}
    adaptiveDelay time.Duration
}
```

The coordinator ensures smooth interaction with the Python service:

- **Semaphore-based concurrency control**
- **Health-aware request pacing**
- **Adaptive backpressure based on queue depth**
- **Intelligent request batching**

### Filter Chain Architecture

```go
type FilterChain struct {
    filters       []Filter
    cache         *FilterCache
    optimizer     *FilterOptimizer
}

type Filter interface {
    Apply(contracts []Contract) []Contract
    Name() string
    Priority() int
    Selectivity() float64
}
```

The filter chain implements:

1. **Dynamic Reordering**: Filters arranged by selectivity for optimal performance
2. **Parallel Execution**: Independent filters run concurrently
3. **Result Caching**: Cache filter results when appropriate
4. **Short-circuit Evaluation**: Stop early when no contracts remain

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         Scanner Service                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  API Request → Filter Config → Scanner Engine                    │
│                                  │                               │
│                                  ├─→ Request Coordinator         │
│                                  │     │                         │
│                                  │     ├─→ Python Service        │
│                                  │     │   (Market Data)         │
│                                  │     │                         │
│                                  │     ←── Contract Data         │
│                                  │                               │
│                                  ├─→ Filter Chain               │
│                                  │     │                         │
│                                  │     ├─→ Liquidity Filter     │
│                                  │     ├─→ DTE Filter           │
│                                  │     ├─→ Delta Filter         │
│                                  │     ├─→ Greeks Filter        │
│                                  │     └─→ Custom Filters       │
│                                  │                               │
│                                  ├─→ Result Aggregation         │
│                                  │                               │
│                                  └─→ WebSocket Broadcast        │
│                                      Real-time Updates           │
└─────────────────────────────────────────────────────────────────┘
```

## Coordination Protocol

### Health-Based Rate Control

The scanner continuously monitors Python service health and adjusts its behavior:

```go
func (rc *RequestCoordinator) calculateBackpressure(health HealthStatus) time.Duration {
    switch {
    case health.QueueSize > 100:
        return 500 * time.Millisecond  // Heavy backpressure
    case health.QueueSize > 75:
        return 100 * time.Millisecond  // Moderate backpressure
    case health.QueueSize > 50:
        return 50 * time.Millisecond   // Light backpressure
    case health.QueueSize > 25:
        return 25 * time.Millisecond   // Minimal backpressure
    default:
        return 10 * time.Millisecond   // Normal operation
    }
}
```

### Request Batching

Intelligent batching reduces API calls and improves efficiency:

```go
func (rc *RequestCoordinator) createOptimalBatches(contracts []Contract) [][]Contract {
    // Group by underlying symbol for efficient API usage
    // Respect Python service batch size limits
    // Consider current queue depth
    return optimizedBatches
}
```

## Performance Patterns

### 1. Concurrent Filter Execution

```go
func (fc *FilterChain) ApplyConcurrent(contracts []Contract) []Contract {
    // Split independent filters into groups
    independentGroups := fc.groupIndependentFilters()
    
    // Execute each group concurrently
    results := make(chan []Contract, len(independentGroups))
    
    for _, group := range independentGroups {
        go func(filters []Filter) {
            filtered := contracts
            for _, filter := range filters {
                filtered = filter.Apply(filtered)
            }
            results <- filtered
        }(group)
    }
    
    // Merge results
    return fc.mergeResults(results)
}
```

### 2. Smart Caching Strategy

```go
type FilterCache struct {
    cache     map[string]CacheEntry
    ttl       time.Duration
    maxSize   int
    evictLRU  bool
}

func (fc *FilterCache) GetOrCompute(key string, compute func() []Contract) []Contract {
    if entry, ok := fc.cache[key]; ok && !entry.IsExpired() {
        fc.updateAccessTime(key)
        return entry.Contracts
    }
    
    result := compute()
    fc.Set(key, result)
    return result
}
```

### 3. Memory Pool for Allocations

```go
var contractPool = sync.Pool{
    New: func() interface{} {
        return make([]Contract, 0, 1000)
    },
}

func getContractSlice() []Contract {
    return contractPool.Get().([]Contract)[:0]
}

func putContractSlice(s []Contract) {
    contractPool.Put(s)
}
```

## Error Handling

### Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    failures     int
    lastFailTime time.Time
    state        State
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.state = HalfOpen
        } else {
            return ErrCircuitOpen
        }
    }
    
    err := fn()
    if err != nil {
        cb.recordFailure()
    } else {
        cb.reset()
    }
    
    return err
}
```

## Metrics and Monitoring

The scanner exposes comprehensive metrics for monitoring:

### Performance Metrics
- `scanner_scan_duration_seconds` - Time to complete full scan
- `scanner_filter_duration_seconds` - Time per filter execution
- `scanner_cache_hit_rate` - Cache effectiveness
- `scanner_contracts_processed_total` - Total contracts scanned

### Health Metrics
- `scanner_active_scans` - Currently running scans
- `scanner_queue_depth` - Pending scan requests
- `scanner_coordinator_backpressure` - Current backpressure delay
- `scanner_circuit_breaker_state` - Circuit breaker status

### Business Metrics
- `scanner_results_found_total` - Opportunities identified
- `scanner_filter_selectivity` - Filter effectiveness
- `scanner_websocket_clients` - Connected real-time clients

## Best Practices

### 1. Filter Ordering
Always order filters by selectivity (most restrictive first):
```
Liquidity → DTE → Delta → Greeks → Custom
```

### 2. Batch Size Tuning
Optimal batch sizes depend on:
- Python service capacity
- Network latency
- Contract complexity
- Current system load

### 3. Cache Management
- Cache filter results with appropriate TTL
- Use LRU eviction for memory management
- Clear cache on configuration changes
- Monitor cache hit rates

### 4. Error Recovery
- Implement exponential backoff for retries
- Use circuit breakers for failing dependencies
- Log all errors with context
- Gracefully degrade functionality

## Integration Points

### Python Service API
```
POST /api/market-data/batch
GET  /api/health
GET  /api/metrics
```

### Scanner REST API
```
POST   /scan/start
DELETE /scan/stop
GET    /scan/status
POST   /scan/filters
GET    /scan/results
```

### WebSocket Protocol
```
Connected → Subscribe → Receive Updates → Handle Reconnect
```

## Performance Benchmarks

Target performance metrics:
- **Scan Latency**: < 100ms for 10,000 contracts
- **Filter Throughput**: > 100,000 contracts/second
- **Memory Usage**: < 500MB under normal load
- **CPU Usage**: < 50% on 4-core system
- **WebSocket Latency**: < 50ms for updates