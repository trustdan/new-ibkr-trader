# Monitoring and Metrics Guide

## Overview

Comprehensive monitoring is critical for a trading system. This guide covers the metrics, alerts, and dashboards needed to maintain system health and trading performance.

## Metrics Architecture

```
┌─────────────────┐     Metrics      ┌──────────────────┐
│  Python Service │ ───────────────► │                  │
│  (TWS Client)   │                  │   Prometheus     │
└─────────────────┘                  │                  │
                                     └────────┬─────────┘
┌─────────────────┐                          │
│   Go Scanner    │ ───────────────►         │
│    Service      │     Metrics              │
└─────────────────┘                          ▼
                                     ┌──────────────────┐
┌─────────────────┐                  │                  │
│  Node Exporter  │ ───────────────► │     Grafana     │
│  (System Stats) │                  │   (Dashboards)  │
└─────────────────┘                  └──────────────────┘
```

## Core Metrics Categories

### 1. System Health Metrics

```python
from prometheus_client import Counter, Gauge, Histogram, Summary
import psutil
import asyncio

class SystemMetrics:
    """System-level health metrics"""
    
    def __init__(self):
        # CPU metrics
        self.cpu_usage = Gauge('system_cpu_usage_percent', 
                              'CPU usage percentage')
        self.cpu_cores = Gauge('system_cpu_cores_available', 
                              'Number of CPU cores')
        
        # Memory metrics
        self.memory_used = Gauge('system_memory_used_bytes', 
                                'Memory used in bytes')
        self.memory_available = Gauge('system_memory_available_bytes', 
                                     'Memory available in bytes')
        self.memory_percent = Gauge('system_memory_usage_percent', 
                                   'Memory usage percentage')
        
        # Disk metrics
        self.disk_usage = Gauge('system_disk_usage_percent', 
                               'Disk usage percentage', ['mount'])
        self.disk_io_read = Counter('system_disk_io_read_bytes', 
                                   'Disk read bytes')
        self.disk_io_write = Counter('system_disk_io_write_bytes', 
                                    'Disk write bytes')
        
        # Network metrics
        self.network_sent = Counter('system_network_sent_bytes', 
                                   'Network bytes sent')
        self.network_recv = Counter('system_network_recv_bytes', 
                                   'Network bytes received')
        
    async def update_metrics(self):
        """Update all system metrics"""
        while True:
            # CPU
            self.cpu_usage.set(psutil.cpu_percent(interval=1))
            self.cpu_cores.set(psutil.cpu_count())
            
            # Memory
            mem = psutil.virtual_memory()
            self.memory_used.set(mem.used)
            self.memory_available.set(mem.available)
            self.memory_percent.set(mem.percent)
            
            # Disk
            for partition in psutil.disk_partitions():
                usage = psutil.disk_usage(partition.mountpoint)
                self.disk_usage.labels(mount=partition.mountpoint).set(usage.percent)
            
            # Network
            net = psutil.net_io_counters()
            self.network_sent.inc(net.bytes_sent)
            self.network_recv.inc(net.bytes_recv)
            
            await asyncio.sleep(10)
```

### 2. TWS Connection Metrics

```python
class TWSConnectionMetrics:
    """TWS connection and API metrics"""
    
    def __init__(self):
        # Connection status
        self.connection_status = Gauge('tws_connection_status', 
                                      'TWS connection status (1=connected, 0=disconnected)')
        self.connection_uptime = Gauge('tws_connection_uptime_seconds', 
                                      'Time since last connection')
        self.reconnection_count = Counter('tws_reconnection_total', 
                                         'Number of reconnection attempts')
        
        # API metrics
        self.api_requests = Counter('tws_api_requests_total', 
                                   'Total API requests', ['method'])
        self.api_errors = Counter('tws_api_errors_total', 
                                 'API errors', ['error_code'])
        self.api_latency = Histogram('tws_api_latency_seconds', 
                                    'API request latency', ['method'])
        
        # Rate limiting
        self.rate_limit_hits = Counter('tws_rate_limit_hits_total', 
                                      'Rate limit violations')
        self.request_rate = Gauge('tws_request_rate_per_second', 
                                 'Current request rate')
        
        # Socket metrics
        self.socket_bytes_sent = Counter('tws_socket_bytes_sent_total', 
                                        'Bytes sent over socket')
        self.socket_bytes_recv = Counter('tws_socket_bytes_received_total', 
                                        'Bytes received over socket')
        self.socket_errors = Counter('tws_socket_errors_total', 
                                    'Socket errors', ['error_type'])
```

### 3. Trading Metrics

```python
class TradingMetrics:
    """Trading activity metrics"""
    
    def __init__(self):
        # Order metrics
        self.orders_placed = Counter('trading_orders_placed_total', 
                                    'Orders placed', ['order_type', 'action'])
        self.orders_filled = Counter('trading_orders_filled_total', 
                                    'Orders filled', ['order_type', 'action'])
        self.orders_cancelled = Counter('trading_orders_cancelled_total', 
                                       'Orders cancelled', ['reason'])
        self.orders_rejected = Counter('trading_orders_rejected_total', 
                                      'Orders rejected', ['reason'])
        
        # Fill metrics
        self.fill_price = Summary('trading_fill_price', 
                                 'Fill prices', ['symbol', 'action'])
        self.fill_latency = Histogram('trading_fill_latency_seconds', 
                                     'Time from order to fill')
        self.slippage = Summary('trading_slippage_dollars', 
                               'Slippage in dollars', ['symbol'])
        
        # Position metrics
        self.positions_open = Gauge('trading_positions_open', 
                                   'Open positions', ['symbol'])
        self.position_pnl = Gauge('trading_position_pnl_dollars', 
                                 'Position P&L', ['symbol'])
        self.total_pnl = Gauge('trading_total_pnl_dollars', 
                              'Total P&L')
        
        # Risk metrics
        self.margin_used = Gauge('trading_margin_used_dollars', 
                                'Margin used')
        self.buying_power = Gauge('trading_buying_power_dollars', 
                                 'Available buying power')
        self.risk_exposure = Gauge('trading_risk_exposure_dollars', 
                                  'Total risk exposure')
```

### 4. Scanner Metrics

```python
class ScannerMetrics:
    """Options scanner metrics"""
    
    def __init__(self):
        # Scan performance
        self.scans_completed = Counter('scanner_scans_completed_total', 
                                      'Completed scans')
        self.scan_duration = Histogram('scanner_scan_duration_seconds', 
                                      'Scan duration', 
                                      buckets=[0.1, 0.5, 1, 2, 5, 10])
        self.results_found = Histogram('scanner_results_found', 
                                      'Results per scan', 
                                      buckets=[0, 10, 50, 100, 500, 1000])
        
        # Filter metrics
        self.filters_active = Gauge('scanner_filters_active', 
                                   'Active filter count')
        self.filter_efficiency = Gauge('scanner_filter_efficiency_percent', 
                                      'Percentage of results passing filters')
        
        # Queue metrics
        self.scan_queue_depth = Gauge('scanner_queue_depth', 
                                     'Pending scans in queue')
        self.scan_queue_wait = Histogram('scanner_queue_wait_seconds', 
                                        'Time spent in queue')
        
        # Cache metrics
        self.cache_hits = Counter('scanner_cache_hits_total', 
                                 'Cache hit count')
        self.cache_misses = Counter('scanner_cache_misses_total', 
                                   'Cache miss count')
        self.cache_size = Gauge('scanner_cache_size_bytes', 
                               'Cache memory usage')
```

### 5. Market Data Metrics

```python
class MarketDataMetrics:
    """Market data flow metrics"""
    
    def __init__(self):
        # Data flow
        self.ticks_received = Counter('market_data_ticks_received_total', 
                                     'Ticks received', ['tick_type'])
        self.ticks_processed = Counter('market_data_ticks_processed_total', 
                                      'Ticks processed', ['tick_type'])
        self.tick_latency = Histogram('market_data_tick_latency_seconds', 
                                     'Tick processing latency')
        
        # Subscription metrics
        self.subscriptions_active = Gauge('market_data_subscriptions_active', 
                                         'Active subscriptions')
        self.subscription_limit = Gauge('market_data_subscription_limit', 
                                       'Subscription limit')
        self.subscription_errors = Counter('market_data_subscription_errors_total', 
                                          'Subscription errors', ['error_type'])
        
        # Data quality
        self.data_gaps = Counter('market_data_gaps_total', 
                                'Detected data gaps')
        self.stale_data = Gauge('market_data_stale_seconds', 
                               'Seconds since last update', ['symbol'])
        self.invalid_ticks = Counter('market_data_invalid_ticks_total', 
                                    'Invalid ticks received')
```

## Alert Rules

### Critical Alerts

```yaml
# prometheus/alerts/critical.yml
groups:
  - name: critical_alerts
    interval: 30s
    rules:
      # TWS Connection Lost
      - alert: TWSConnectionLost
        expr: tws_connection_status == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "TWS connection lost"
          description: "TWS has been disconnected for {{ $value }} minutes"
      
      # High Error Rate
      - alert: HighAPIErrorRate
        expr: rate(tws_api_errors_total[5m]) > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High API error rate"
          description: "API error rate is {{ $value }} errors/sec"
      
      # Order Rejection Spike
      - alert: OrderRejectionSpike
        expr: rate(trading_orders_rejected_total[5m]) > 5
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "High order rejection rate"
          description: "{{ $value }} orders/sec being rejected"
```

### Warning Alerts

```yaml
  - name: warning_alerts
    rules:
      # High Memory Usage
      - alert: HighMemoryUsage
        expr: system_memory_usage_percent > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage at {{ $value }}%"
      
      # Approaching Rate Limit
      - alert: ApproachingRateLimit
        expr: tws_request_rate_per_second > 40
        for: 30s
        labels:
          severity: warning
        annotations:
          summary: "Approaching TWS rate limit"
          description: "Request rate at {{ $value }} req/s (limit: 50)"
      
      # Low Cache Hit Rate
      - alert: LowCacheHitRate
        expr: |
          rate(scanner_cache_hits_total[5m]) / 
          (rate(scanner_cache_hits_total[5m]) + rate(scanner_cache_misses_total[5m])) 
          < 0.7
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low scanner cache hit rate"
          description: "Cache hit rate at {{ $value }}%"
```

## Grafana Dashboards

### 1. System Overview Dashboard

```json
{
  "dashboard": {
    "title": "IBKR Trading System Overview",
    "panels": [
      {
        "title": "System Health",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [
          {
            "expr": "system_cpu_usage_percent",
            "legendFormat": "CPU Usage"
          },
          {
            "expr": "system_memory_usage_percent",
            "legendFormat": "Memory Usage"
          }
        ]
      },
      {
        "title": "TWS Connection Status",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [
          {
            "expr": "tws_connection_status",
            "legendFormat": "Connection"
          }
        ]
      },
      {
        "title": "Order Flow",
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 8},
        "targets": [
          {
            "expr": "rate(trading_orders_placed_total[5m])",
            "legendFormat": "Orders Placed"
          },
          {
            "expr": "rate(trading_orders_filled_total[5m])",
            "legendFormat": "Orders Filled"
          }
        ]
      }
    ]
  }
}
```

### 2. Trading Performance Dashboard

```json
{
  "dashboard": {
    "title": "Trading Performance",
    "panels": [
      {
        "title": "P&L Tracking",
        "gridPos": {"h": 10, "w": 24, "x": 0, "y": 0},
        "targets": [
          {
            "expr": "trading_total_pnl_dollars",
            "legendFormat": "Total P&L"
          }
        ]
      },
      {
        "title": "Fill Quality",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 10},
        "targets": [
          {
            "expr": "histogram_quantile(0.5, trading_slippage_dollars)",
            "legendFormat": "Median Slippage"
          }
        ]
      },
      {
        "title": "Risk Metrics",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 10},
        "targets": [
          {
            "expr": "trading_margin_used_dollars / trading_buying_power_dollars",
            "legendFormat": "Margin Utilization"
          }
        ]
      }
    ]
  }
}
```

## Logging Strategy

### Structured Logging

```python
import structlog
from datetime import datetime

# Configure structured logging
structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
        structlog.processors.JSONRenderer()
    ],
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    cache_logger_on_first_use=True,
)

class TradingLogger:
    """Structured logging for trading events"""
    
    def __init__(self, component: str):
        self.logger = structlog.get_logger(component=component)
        
    def log_order(self, order_id: int, action: str, **kwargs):
        """Log order-related events"""
        self.logger.info(
            "order_event",
            order_id=order_id,
            action=action,
            timestamp=datetime.now().isoformat(),
            **kwargs
        )
    
    def log_fill(self, order_id: int, fill_price: float, **kwargs):
        """Log fill events"""
        self.logger.info(
            "fill_event",
            order_id=order_id,
            fill_price=fill_price,
            timestamp=datetime.now().isoformat(),
            **kwargs
        )
    
    def log_error(self, error_type: str, error_msg: str, **kwargs):
        """Log errors with context"""
        self.logger.error(
            "error_event",
            error_type=error_type,
            error_msg=error_msg,
            timestamp=datetime.now().isoformat(),
            **kwargs
        )
```

## Best Practices

1. **Metric Naming**
   - Use consistent prefixes (tws_, trading_, scanner_)
   - Include units in metric names (_seconds, _bytes, _percent)
   - Use labels for dimensions, not metric proliferation

2. **Dashboard Design**
   - Overview dashboard for quick health check
   - Detailed dashboards for specific components
   - Use consistent color schemes
   - Include relevant time ranges

3. **Alert Tuning**
   - Start with conservative thresholds
   - Reduce noise by using "for" durations
   - Include runbooks in alert descriptions
   - Test alerts in staging environment

4. **Performance Considerations**
   - Limit cardinality of labels
   - Use recording rules for complex queries
   - Set appropriate retention policies
   - Monitor Prometheus resource usage