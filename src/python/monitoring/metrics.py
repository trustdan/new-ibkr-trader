"""
Prometheus metrics for monitoring IBKR service health
Tracks the vibe while maintaining observability
"""
from prometheus_client import Counter, Histogram, Gauge, Info
import time


# Connection metrics
connection_status = Gauge(
    'ibkr_connection_status',
    'TWS connection status',
    ['status']
)

connection_attempts = Counter(
    'ibkr_connection_attempts_total',
    'Total number of connection attempts'
)

# API metrics
api_requests = Counter(
    'ibkr_api_requests_total',
    'Total API requests',
    ['endpoint']
)

api_errors = Counter(
    'ibkr_api_errors_total',
    'Total API errors',
    ['error_code']
)

api_request_duration = Histogram(
    'ibkr_api_request_duration_seconds',
    'API request duration',
    ['endpoint']
)

# Trading metrics
orders_placed = Counter(
    'ibkr_orders_placed_total',
    'Total orders placed',
    ['order_type', 'symbol']
)

orders_filled = Counter(
    'ibkr_orders_filled_total',
    'Total orders filled',
    ['symbol']
)

# Market data metrics
market_data_subscriptions = Gauge(
    'ibkr_market_data_subscriptions',
    'Current market data subscriptions'
)

ticks_received = Counter(
    'ibkr_ticks_received_total',
    'Total market data ticks received',
    ['symbol']
)

# System metrics
vibe_level = Gauge(
    'ibkr_vibe_level',
    'Current system vibe level (0-10)'
)

# Service info
service_info = Info(
    'ibkr_service',
    'Service information'
)


def setup_metrics():
    """Initialize metrics with default values"""
    # Set initial connection status
    connection_status.labels(status="connected").set(0)
    connection_status.labels(status="disconnected").set(1)
    
    # Set service info
    service_info.info({
        'version': '0.1.0',
        'environment': 'development',
        'vibe': 'async-first'
    })
    
    # Set initial vibe level (starting strong!)
    vibe_level.set(9.0)


class MetricsTimer:
    """Context manager for timing operations"""
    
    def __init__(self, histogram: Histogram, labels: dict = None):
        self.histogram = histogram
        self.labels = labels or {}
        self.start_time = None
        
    def __enter__(self):
        self.start_time = time.time()
        return self
        
    def __exit__(self, exc_type, exc_val, exc_tb):
        duration = time.time() - self.start_time
        if self.labels:
            self.histogram.labels(**self.labels).observe(duration)
        else:
            self.histogram.observe(duration)