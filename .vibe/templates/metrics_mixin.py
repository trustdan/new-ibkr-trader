"""
Metrics Mixin - Add Prometheus metrics to any service

Usage: Inherit from MetricsMixin and call setup_metrics()
"""
from prometheus_client import Counter, Gauge, Histogram, Summary
import time
from functools import wraps
import asyncio


class MetricsMixin:
    """Add standardized metrics to any service"""
    
    def setup_metrics(self, service_name: str):
        """Initialize standard metrics for a service"""
        prefix = f"ibkr_{service_name}"
        
        # Connection metrics
        self.metric_connected = Gauge(
            f'{prefix}_connected',
            'Is service connected to TWS (1=yes, 0=no)'
        )
        
        # Request metrics
        self.metric_requests_total = Counter(
            f'{prefix}_requests_total',
            'Total API requests',
            ['method', 'status']
        )
        
        self.metric_request_duration = Histogram(
            f'{prefix}_request_duration_seconds',
            'API request duration',
            ['method']
        )
        
        # Error metrics
        self.metric_errors_total = Counter(
            f'{prefix}_errors_total',
            'Total errors',
            ['error_code', 'error_type']
        )
        
        # Service-specific gauges
        self.metric_queue_size = Gauge(
            f'{prefix}_queue_size',
            'Current queue size'
        )
        
        self.metric_active_subscriptions = Gauge(
            f'{prefix}_active_subscriptions',
            'Active market data subscriptions'
        )
        
        # Performance metrics
        self.metric_event_loop_lag = Summary(
            f'{prefix}_event_loop_lag_seconds',
            'Event loop lag'
        )
        
        # Business metrics
        self.metric_orders_placed = Counter(
            f'{prefix}_orders_placed_total',
            'Total orders placed',
            ['order_type', 'symbol']
        )
        
        self.metric_orders_filled = Counter(
            f'{prefix}_orders_filled_total',
            'Total orders filled',
            ['order_type', 'symbol']
        )
        
    def track_connection(self, connected: bool):
        """Update connection status"""
        self.metric_connected.set(1 if connected else 0)
        
    def track_request(self, method: str):
        """Decorator to track API requests"""
        def decorator(func):
            @wraps(func)
            async def async_wrapper(*args, **kwargs):
                start = time.time()
                status = 'success'
                try:
                    result = await func(*args, **kwargs)
                    return result
                except Exception as e:
                    status = 'error'
                    raise
                finally:
                    duration = time.time() - start
                    self.metric_requests_total.labels(method=method, status=status).inc()
                    self.metric_request_duration.labels(method=method).observe(duration)
                    
            @wraps(func)
            def sync_wrapper(*args, **kwargs):
                start = time.time()
                status = 'success'
                try:
                    result = func(*args, **kwargs)
                    return result
                except Exception as e:
                    status = 'error'
                    raise
                finally:
                    duration = time.time() - start
                    self.metric_requests_total.labels(method=method, status=status).inc()
                    self.metric_request_duration.labels(method=method).observe(duration)
                    
            return async_wrapper if asyncio.iscoroutinefunction(func) else sync_wrapper
        return decorator
        
    def track_error(self, error_code: int, error_type: str = 'api'):
        """Track error occurrences"""
        self.metric_errors_total.labels(
            error_code=str(error_code),
            error_type=error_type
        ).inc()
        
    async def track_event_loop_health(self):
        """Monitor event loop responsiveness"""
        while True:
            start = asyncio.get_event_loop().time()
            await asyncio.sleep(0)  # Yield to event loop
            lag = asyncio.get_event_loop().time() - start
            self.metric_event_loop_lag.observe(lag)
            await asyncio.sleep(10)  # Check every 10 seconds
            
    # Convenience methods for common patterns
    def inc_queue_size(self, delta: int = 1):
        """Increment queue size"""
        self.metric_queue_size.inc(delta)
        
    def dec_queue_size(self, delta: int = 1):
        """Decrement queue size"""
        self.metric_queue_size.dec(delta)
        
    def set_subscriptions(self, count: int):
        """Set active subscription count"""
        self.metric_active_subscriptions.set(count)
        
    def track_order_placed(self, order_type: str, symbol: str):
        """Track order placement"""
        self.metric_orders_placed.labels(
            order_type=order_type,
            symbol=symbol
        ).inc()
        
    def track_order_filled(self, order_type: str, symbol: str):
        """Track order fill"""
        self.metric_orders_filled.labels(
            order_type=order_type,
            symbol=symbol
        ).inc()


# Example usage
class MyService(MetricsMixin):
    def __init__(self):
        self.setup_metrics('my_service')
        
    @track_request('get_market_data')
    async def get_market_data(self, symbol: str):
        # Method automatically tracked
        pass
        
    def on_error(self, error_code: int):
        # Track error
        self.track_error(error_code)
        
        # Update connection if needed
        if error_code == 1100:
            self.track_connection(False)