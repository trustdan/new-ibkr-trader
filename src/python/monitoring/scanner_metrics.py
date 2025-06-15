"""
Prometheus metrics for Python scanner integration

This module provides comprehensive metrics for monitoring the scanner integration
performance and health.
"""

from prometheus_client import Counter, Histogram, Gauge, Info
import time
from functools import wraps
from typing import Callable, Any

# Scanner request metrics
scanner_requests_total = Counter(
    'scanner_requests_total',
    'Total number of scanner requests',
    ['method', 'status']
)

scanner_request_duration = Histogram(
    'scanner_request_duration_seconds',
    'Scanner request duration in seconds',
    ['method'],
    buckets=(0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0)
)

scanner_active_requests = Gauge(
    'scanner_active_requests',
    'Number of active scanner requests'
)

# Integration metrics
integration_scans_total = Counter(
    'integration_scans_total',
    'Total number of scans through integration layer',
    ['symbol', 'status']
)

integration_cache_hits = Counter(
    'integration_cache_hits_total',
    'Total number of cache hits in integration layer'
)

integration_cache_size = Gauge(
    'integration_cache_size',
    'Current size of integration cache'
)

# Batch processing metrics
batch_requests_total = Counter(
    'batch_requests_total',
    'Total number of batch requests',
    ['status']
)

batch_size_histogram = Histogram(
    'batch_size',
    'Size of processed batches',
    buckets=(1, 5, 10, 25, 50, 100)
)

batch_processing_duration = Histogram(
    'batch_processing_duration_seconds',
    'Batch processing duration in seconds',
    buckets=(0.5, 1.0, 2.5, 5.0, 10.0, 30.0)
)

# Backpressure metrics
backpressure_rejections = Counter(
    'backpressure_rejections_total',
    'Total number of requests rejected by backpressure'
)

backpressure_wait_time = Histogram(
    'backpressure_wait_time_seconds',
    'Time spent waiting for backpressure permit',
    buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1.0)
)

circuit_breaker_state = Gauge(
    'circuit_breaker_state',
    'Circuit breaker state (0=closed, 1=open)'
)

# Performance metrics
scanner_qps = Gauge(
    'scanner_qps',
    'Current queries per second to scanner'
)

scanner_latency_p50 = Gauge(
    'scanner_latency_p50_seconds',
    '50th percentile latency'
)

scanner_latency_p95 = Gauge(
    'scanner_latency_p95_seconds',
    '95th percentile latency'
)

scanner_latency_p99 = Gauge(
    'scanner_latency_p99_seconds',
    '99th percentile latency'
)

# Error metrics
scanner_errors_total = Counter(
    'scanner_errors_total',
    'Total number of scanner errors',
    ['error_type']
)

scanner_timeouts_total = Counter(
    'scanner_timeouts_total',
    'Total number of scanner timeouts'
)

# System info
scanner_info = Info(
    'scanner_info',
    'Scanner service information'
)

# Update system info
scanner_info.info({
    'version': '2.0',
    'integration': 'python-go',
    'features': 'advanced-scoring,greeks-analysis,batching'
})


# Decorators for metric collection

def track_request_metrics(method: str):
    """Decorator to track request metrics"""
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs) -> Any:
            scanner_active_requests.inc()
            start_time = time.time()
            status = 'success'
            
            try:
                result = await func(*args, **kwargs)
                return result
            except Exception as e:
                status = 'error'
                error_type = type(e).__name__
                scanner_errors_total.labels(error_type=error_type).inc()
                raise
            finally:
                duration = time.time() - start_time
                scanner_requests_total.labels(method=method, status=status).inc()
                scanner_request_duration.labels(method=method).observe(duration)
                scanner_active_requests.dec()
                
        return wrapper
    return decorator


def track_scan_metrics(func: Callable) -> Callable:
    """Decorator to track scan operation metrics"""
    @wraps(func)
    async def wrapper(self, symbol: str, *args, **kwargs) -> Any:
        start_time = time.time()
        status = 'success'
        
        try:
            result = await func(self, symbol, *args, **kwargs)
            integration_scans_total.labels(symbol=symbol, status=status).inc()
            return result
        except Exception as e:
            status = 'error'
            integration_scans_total.labels(symbol=symbol, status=status).inc()
            raise
            
    return wrapper


def track_backpressure_metrics(func: Callable) -> Callable:
    """Decorator to track backpressure metrics"""
    @wraps(func)
    async def wrapper(self, *args, **kwargs) -> Any:
        start_time = time.time()
        
        try:
            result = await func(self, *args, **kwargs)
            wait_time = time.time() - start_time
            
            if wait_time > 0.001:  # Only record if there was actual waiting
                backpressure_wait_time.observe(wait_time)
                
            return result
        except Exception as e:
            if 'rate limit' in str(e).lower():
                backpressure_rejections.inc()
            raise
            
    return wrapper


class MetricsCollector:
    """Collects and updates performance metrics"""
    
    def __init__(self):
        self.latencies = []
        self.last_update = time.time()
        
    def record_latency(self, latency: float):
        """Record a latency measurement"""
        self.latencies.append(latency)
        
        # Update metrics every 10 seconds
        if time.time() - self.last_update > 10:
            self._update_percentiles()
            
    def _update_percentiles(self):
        """Update percentile metrics"""
        if not self.latencies:
            return
            
        sorted_latencies = sorted(self.latencies)
        
        # Calculate percentiles
        p50_idx = int(len(sorted_latencies) * 0.50)
        p95_idx = int(len(sorted_latencies) * 0.95)
        p99_idx = int(len(sorted_latencies) * 0.99)
        
        scanner_latency_p50.set(sorted_latencies[p50_idx])
        scanner_latency_p95.set(sorted_latencies[p95_idx])
        scanner_latency_p99.set(sorted_latencies[p99_idx])
        
        # Calculate QPS
        time_window = time.time() - self.last_update
        qps = len(self.latencies) / time_window
        scanner_qps.set(qps)
        
        # Reset for next window
        self.latencies = []
        self.last_update = time.time()
        
    def update_cache_metrics(self, cache_size: int, hit_rate: float):
        """Update cache-related metrics"""
        integration_cache_size.set(cache_size)
        
    def update_circuit_breaker(self, is_open: bool):
        """Update circuit breaker state"""
        circuit_breaker_state.set(1 if is_open else 0)


# Global metrics collector instance
metrics_collector = MetricsCollector()


# Example usage in scanner client
class MetricsEnabledScannerClient:
    """Scanner client with metrics collection"""
    
    def __init__(self, base_client):
        self.client = base_client
        
    @track_request_metrics('scan')
    async def scan(self, request):
        """Scan with metrics tracking"""
        start = time.time()
        
        try:
            result = await self.client.scan(request)
            latency = time.time() - start
            metrics_collector.record_latency(latency)
            return result
        except TimeoutError:
            scanner_timeouts_total.inc()
            raise
            
    @track_request_metrics('health_check')
    async def health_check(self):
        """Health check with metrics tracking"""
        return await self.client.health_check()