# Scanner Performance Guide

## Overview

This guide provides comprehensive strategies for optimizing the IBKR Scanner's performance, including benchmarking methodologies, optimization techniques, and memory management best practices. Our target is sub-100ms scanning of 10,000+ contracts while maintaining minimal memory footprint.

## Performance Targets

### Primary Metrics
| Metric | Target | Critical Threshold |
|--------|--------|-------------------|
| Full Scan Latency | < 100ms | < 500ms |
| Filter Throughput | > 100K contracts/sec | > 50K contracts/sec |
| Memory Usage | < 500MB | < 1GB |
| CPU Usage | < 50% (4 cores) | < 80% |
| Concurrent Scans | 10+ | 5+ |
| WebSocket Latency | < 50ms | < 200ms |

### Business Metrics
| Metric | Target | Minimum |
|--------|--------|---------|
| Results Accuracy | 100% | 99.9% |
| Cache Hit Rate | > 80% | > 60% |
| Error Rate | < 0.01% | < 0.1% |
| Recovery Time | < 5 seconds | < 30 seconds |

## Benchmarking Methodology

### 1. Micro-benchmarks

```go
package benchmark

import (
    "testing"
    "github.com/ibkr-scanner/filters"
)

// Benchmark individual filter performance
func BenchmarkDeltaFilter(b *testing.B) {
    contracts := generateContracts(10000)
    filter := &filters.DeltaFilter{
        MinDelta: 0.25,
        MaxDelta: 0.35,
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _ = filter.Apply(contracts)
    }
}

// Benchmark filter chain
func BenchmarkFilterChain(b *testing.B) {
    contracts := generateContracts(10000)
    chain := createStandardFilterChain()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _ = chain.Apply(contracts)
    }
}

// Benchmark with different data sizes
func BenchmarkScalability(b *testing.B) {
    sizes := []int{100, 1000, 10000, 50000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            contracts := generateContracts(size)
            scanner := NewScanner()
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = scanner.Scan(contracts)
            }
        })
    }
}
```

### 2. Load Testing

```go
// Load test with realistic scenarios
func TestScannerUnderLoad(t *testing.T) {
    scanner := NewScanner()
    contracts := loadRealMarketData()
    
    // Simulate concurrent users
    users := 50
    requestsPerUser := 100
    
    var wg sync.WaitGroup
    errors := make(chan error, users*requestsPerUser)
    latencies := make(chan time.Duration, users*requestsPerUser)
    
    start := time.Now()
    
    for u := 0; u < users; u++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            
            for r := 0; r < requestsPerUser; r++ {
                reqStart := time.Now()
                
                _, err := scanner.Scan(contracts)
                if err != nil {
                    errors <- err
                    continue
                }
                
                latencies <- time.Since(reqStart)
            }
        }(u)
    }
    
    wg.Wait()
    close(errors)
    close(latencies)
    
    // Analyze results
    totalTime := time.Since(start)
    errorCount := len(errors)
    
    var totalLatency time.Duration
    var maxLatency time.Duration
    count := 0
    
    for lat := range latencies {
        totalLatency += lat
        if lat > maxLatency {
            maxLatency = lat
        }
        count++
    }
    
    avgLatency := totalLatency / time.Duration(count)
    
    t.Logf("Load Test Results:")
    t.Logf("  Total Requests: %d", users*requestsPerUser)
    t.Logf("  Total Time: %v", totalTime)
    t.Logf("  Requests/sec: %.2f", float64(count)/totalTime.Seconds())
    t.Logf("  Average Latency: %v", avgLatency)
    t.Logf("  Max Latency: %v", maxLatency)
    t.Logf("  Error Rate: %.2f%%", float64(errorCount)/float64(users*requestsPerUser)*100)
    
    // Assert performance targets
    assert.Less(t, avgLatency, 100*time.Millisecond)
    assert.Less(t, float64(errorCount)/float64(users*requestsPerUser), 0.01)
}
```

### 3. Profiling Tools

```go
// CPU Profiling
func profileCPU() {
    f, _ := os.Create("cpu.prof")
    defer f.Close()
    
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Run scanner workload
    runScannerWorkload()
}

// Memory Profiling
func profileMemory() {
    f, _ := os.Create("mem.prof")
    defer f.Close()
    
    // Run scanner workload
    runScannerWorkload()
    
    runtime.GC()
    pprof.WriteHeapProfile(f)
}

// Trace Analysis
func profileTrace() {
    f, _ := os.Create("trace.out")
    defer f.Close()
    
    trace.Start(f)
    defer trace.Stop()
    
    // Run scanner workload
    runScannerWorkload()
}
```

## Optimization Techniques

### 1. Algorithmic Optimizations

#### Early Exit Strategy
```go
func (fc *FilterChain) ApplyWithEarlyExit(contracts []Contract) []Contract {
    result := contracts
    
    for _, filter := range fc.filters {
        if len(result) == 0 {
            // Early exit - no point continuing
            return result
        }
        
        // Check if filter is worth applying
        if fc.shouldSkipFilter(filter, len(result)) {
            continue
        }
        
        result = filter.Apply(result)
        
        // Update filter statistics
        fc.updateStats(filter, len(contracts), len(result))
    }
    
    return result
}

func (fc *FilterChain) shouldSkipFilter(filter Filter, remainingCount int) bool {
    // Skip expensive filters on small datasets
    if filter.Cost() > High && remainingCount < 100 {
        return true
    }
    
    // Skip filters with low selectivity late in chain
    stats := fc.getStats(filter)
    if stats.Position > 5 && stats.AvgSelectivity > 0.9 {
        return true
    }
    
    return false
}
```

#### Batch Processing
```go
func (s *Scanner) ScanBatch(contractBatches [][]Contract) [][]Contract {
    results := make([][]Contract, len(contractBatches))
    
    // Process batches in parallel
    var wg sync.WaitGroup
    for i, batch := range contractBatches {
        wg.Add(1)
        go func(idx int, contracts []Contract) {
            defer wg.Done()
            results[idx] = s.Scan(contracts)
        }(i, batch)
    }
    
    wg.Wait()
    return results
}
```

### 2. Concurrency Optimizations

#### Worker Pool Pattern
```go
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    results   chan Result
}

func NewWorkerPool(workers int) *WorkerPool {
    wp := &WorkerPool{
        workers:   workers,
        taskQueue: make(chan Task, workers*2),
        results:   make(chan Result, workers*2),
    }
    
    // Start workers
    for i := 0; i < workers; i++ {
        go wp.worker()
    }
    
    return wp
}

func (wp *WorkerPool) worker() {
    for task := range wp.taskQueue {
        result := task.Execute()
        wp.results <- result
    }
}

func (wp *WorkerPool) Submit(task Task) {
    wp.taskQueue <- task
}
```

#### Lock-Free Data Structures
```go
// Use atomic operations for counters
type Metrics struct {
    scansTotal    atomic.Uint64
    errorsTotal   atomic.Uint64
    contractsTotal atomic.Uint64
}

func (m *Metrics) IncrementScans() {
    m.scansTotal.Add(1)
}

// Use sync.Map for concurrent access
type FilterCache struct {
    cache sync.Map // map[string]*CacheEntry
}

func (fc *FilterCache) Get(key string) (*CacheEntry, bool) {
    value, ok := fc.cache.Load(key)
    if !ok {
        return nil, false
    }
    return value.(*CacheEntry), true
}
```

### 3. Memory Management

#### Object Pooling
```go
var contractSlicePool = sync.Pool{
    New: func() interface{} {
        // Pre-allocate with reasonable capacity
        return make([]Contract, 0, 1000)
    },
}

func getContractSlice() []Contract {
    return contractSlicePool.Get().([]Contract)[:0]
}

func putContractSlice(s []Contract) {
    // Clear the slice to help GC
    for i := range s {
        s[i] = Contract{}
    }
    contractSlicePool.Put(s)
}

// Usage in filter
func (f *DeltaFilter) Apply(contracts []Contract) []Contract {
    filtered := getContractSlice()
    defer func() {
        if cap(filtered) > 10000 {
            // Don't pool very large slices
            return
        }
        putContractSlice(filtered)
    }()
    
    for _, contract := range contracts {
        if f.matches(contract) {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}
```

#### Memory-Efficient Data Structures
```go
// Use value types where possible
type CompactContract struct {
    Symbol     [8]byte    // Fixed size instead of string
    Expiry     uint32     // Unix timestamp instead of time.Time
    Strike     float32    // float32 sufficient for prices
    Right      byte       // 'C' or 'P' instead of string
    Greeks     CompactGreeks
}

type CompactGreeks struct {
    Delta float32
    Gamma float32
    Theta float32
    Vega  float32
}

// Bit-packed flags
type ContractFlags uint32

const (
    FlagCall ContractFlags = 1 << iota
    FlagPut
    FlagWeekly
    FlagQuarterly
    FlagLiquid
    FlagATM
)
```

#### Zero-Allocation Techniques
```go
// Reuse buffers
type Scanner struct {
    buffer []Contract
}

func (s *Scanner) Scan(contracts []Contract) []Contract {
    // Reuse internal buffer
    s.buffer = s.buffer[:0]
    
    // Process without additional allocations
    for i := range contracts {
        if s.matches(&contracts[i]) {
            s.buffer = append(s.buffer, contracts[i])
        }
    }
    
    // Return a copy to avoid data races
    result := make([]Contract, len(s.buffer))
    copy(result, s.buffer)
    
    return result
}
```

### 4. Cache Optimization

#### Multi-Level Cache
```go
type MultiLevelCache struct {
    l1 *LRUCache    // Hot data, small, fast
    l2 *LRUCache    // Warm data, larger
    l3 *DiskCache   // Cold data, persistent
}

func (mlc *MultiLevelCache) Get(key string) (interface{}, bool) {
    // Check L1
    if val, ok := mlc.l1.Get(key); ok {
        return val, true
    }
    
    // Check L2
    if val, ok := mlc.l2.Get(key); ok {
        // Promote to L1
        mlc.l1.Set(key, val)
        return val, true
    }
    
    // Check L3
    if val, ok := mlc.l3.Get(key); ok {
        // Promote to L2
        mlc.l2.Set(key, val)
        return val, true
    }
    
    return nil, false
}
```

#### Cache Warming
```go
func (s *Scanner) WarmCache(ctx context.Context) error {
    // Pre-populate cache with common queries
    commonFilters := []FilterConfig{
        {DTE: Range{30, 60}, Delta: Range{0.25, 0.35}},
        {DTE: Range{7, 30}, Delta: Range{0.30, 0.40}},
        // ... more common configurations
    }
    
    for _, config := range commonFilters {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            contracts := s.fetchContracts(config)
            s.cache.Set(config.Key(), contracts)
        }
    }
    
    return nil
}
```

## Performance Monitoring

### Real-time Metrics

```go
type PerformanceMonitor struct {
    scanDuration    prometheus.Histogram
    filterDuration  *prometheus.HistogramVec
    cacheHitRate    prometheus.Gauge
    memoryUsage     prometheus.Gauge
    goroutineCount  prometheus.Gauge
}

func (pm *PerformanceMonitor) RecordScan(duration time.Duration) {
    pm.scanDuration.Observe(duration.Seconds())
}

func (pm *PerformanceMonitor) StartCollection() {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            pm.memoryUsage.Set(float64(m.Alloc) / 1024 / 1024) // MB
            pm.goroutineCount.Set(float64(runtime.NumGoroutine()))
        }
    }()
}
```

### Performance Dashboard

```yaml
# Grafana Dashboard Queries
- name: Scanner Performance
  panels:
    - title: Scan Latency (p50, p95, p99)
      query: |
        histogram_quantile(0.5, scanner_scan_duration_seconds)
        histogram_quantile(0.95, scanner_scan_duration_seconds)
        histogram_quantile(0.99, scanner_scan_duration_seconds)
    
    - title: Filter Performance
      query: |
        rate(scanner_filter_duration_seconds_sum[5m]) / 
        rate(scanner_filter_duration_seconds_count[5m])
    
    - title: Cache Effectiveness
      query: |
        scanner_cache_hits_total / 
        (scanner_cache_hits_total + scanner_cache_misses_total)
    
    - title: Memory Usage
      query: scanner_memory_usage_mb
    
    - title: Concurrent Scans
      query: scanner_active_scans
```

## Production Tuning

### 1. Resource Allocation

```go
func configureScannerForProduction() *Scanner {
    // CPU allocation
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Memory limits
    debug.SetMemoryLimit(1 << 30) // 1GB
    
    // GC tuning
    debug.SetGCPercent(100) // Default is 100
    
    return &Scanner{
        workers:      runtime.NumCPU() * 2,
        cacheSize:    10000,
        batchSize:    100,
        maxQueueSize: 1000,
    }
}
```

### 2. Adaptive Performance

```go
type AdaptiveScanner struct {
    scanner         *Scanner
    loadMonitor     *LoadMonitor
    performanceMode PerformanceMode
}

type PerformanceMode int

const (
    ModeNormal PerformanceMode = iota
    ModeHighPerformance
    ModePowerSaving
)

func (as *AdaptiveScanner) Scan(contracts []Contract) []Contract {
    load := as.loadMonitor.CurrentLoad()
    
    // Adjust performance mode based on load
    switch {
    case load > 0.8:
        as.performanceMode = ModePowerSaving
        as.scanner.workers = 2
        as.scanner.cacheSize = 1000
        
    case load < 0.3:
        as.performanceMode = ModeHighPerformance
        as.scanner.workers = runtime.NumCPU() * 2
        as.scanner.cacheSize = 20000
        
    default:
        as.performanceMode = ModeNormal
        as.scanner.workers = runtime.NumCPU()
        as.scanner.cacheSize = 10000
    }
    
    return as.scanner.Scan(contracts)
}
```

### 3. Graceful Degradation

```go
func (s *Scanner) ScanWithDegradation(contracts []Contract) []Contract {
    timeout := 100 * time.Millisecond
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    resultCh := make(chan []Contract, 1)
    
    go func() {
        resultCh <- s.fullScan(contracts)
    }()
    
    select {
    case result := <-resultCh:
        return result
        
    case <-ctx.Done():
        // Timeout - return partial results
        return s.quickScan(contracts)
    }
}

func (s *Scanner) quickScan(contracts []Contract) []Contract {
    // Apply only essential filters
    essentialFilters := []Filter{
        s.liquidityFilter,
        s.dteFilter,
    }
    
    result := contracts
    for _, filter := range essentialFilters {
        result = filter.Apply(result)
    }
    
    return result
}
```

## Troubleshooting Performance Issues

### Common Bottlenecks

1. **Filter Chain Order**
   - Solution: Reorder filters by selectivity
   - Tool: Filter statistics monitoring

2. **Cache Misses**
   - Solution: Increase cache size or adjust TTL
   - Tool: Cache hit rate metrics

3. **Memory Pressure**
   - Solution: Use object pooling and compact types
   - Tool: Memory profiler and GC stats

4. **Lock Contention**
   - Solution: Use lock-free structures or sharding
   - Tool: Mutex profiler and trace analysis

5. **Network Latency**
   - Solution: Batch requests and implement caching
   - Tool: Distributed tracing

### Performance Checklist

- [ ] Benchmark all filters individually
- [ ] Profile CPU and memory usage
- [ ] Monitor production metrics
- [ ] Set up alerts for performance degradation
- [ ] Document performance characteristics
- [ ] Plan capacity for growth
- [ ] Test under realistic load
- [ ] Implement graceful degradation
- [ ] Configure adaptive performance
- [ ] Regular performance reviews