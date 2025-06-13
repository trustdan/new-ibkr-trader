# 📊 Monitoring Strategy

Real-time visibility into our trading system's health and performance.

## Philosophy

"You can't optimize what you don't measure." We implement comprehensive monitoring from day one, not as an afterthought.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Grafana Dashboards                        │
│  - System Health  - Trading Metrics  - Performance Graphs   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    ┌──────┴──────┐
                    │ Prometheus  │
                    │  Metrics DB │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────┴────┐      ┌─────┴─────┐     ┌─────┴─────┐
   │ Python  │      │ Go Scanner│     │    GUI    │
   │ Service │      │  Service  │     │  Backend  │
   └─────────┘      └───────────┘     └───────────┘
```

## Key Metrics

### 🔌 Connection Health
- `ibkr_connection_status` - TWS connection state (0/1)
- `ibkr_connection_uptime` - Seconds since last connection
- `ibkr_reconnection_count` - Total reconnection attempts
- `ibkr_last_error_code` - Most recent TWS error code

### 📈 Market Data
- `ibkr_active_subscriptions` - Current market data subscriptions
- `ibkr_subscription_usage_pct` - Percentage of limit used
- `ibkr_subscription_evictions_total` - LRU eviction count
- `ibkr_market_data_updates_per_second` - Update rate

### 🎯 Scanner Performance
- `scanner_execution_time_ms` - Time to complete scan
- `scanner_contracts_processed` - Contracts evaluated
- `scanner_results_found` - Matching opportunities
- `scanner_filter_performance` - Time per filter type

### 💼 Trading Metrics
- `orders_placed_total` - Total orders by type
- `orders_filled_total` - Successfully filled orders
- `order_execution_time_ms` - Time from submit to fill
- `order_rejection_rate` - Percentage rejected

### ⚡ System Performance
- `request_queue_depth` - Pending API requests
- `api_request_rate` - Requests per second
- `throttle_events_total` - Rate limit violations
- `memory_usage_bytes` - Service memory consumption

## Dashboards

### 1. System Health Overview
Real-time status of all components:
- Connection status indicators
- Service health checks
- Error rate graphs
- Queue depth visualization

### 2. Trading Performance
Track trading effectiveness:
- Order fill rates
- Execution times
- P&L tracking
- Position monitoring

### 3. API Usage
Stay within limits:
- Request rate vs limits
- Subscription usage gauge
- Throttling events
- Historical patterns

### 4. Scanner Analytics
Optimize scanning:
- Scan performance trends
- Filter effectiveness
- Result quality metrics
- Resource utilization

## Alert Rules

### Critical Alerts 🚨
- TWS disconnection > 30 seconds
- API error rate > 5%
- Order rejection rate > 10%
- Service down

### Warning Alerts ⚠️
- Subscription usage > 80%
- Queue depth > 100
- Memory usage > 80%
- Throttle events detected

### Info Alerts ℹ️
- Daily TWS restart approaching
- Scan performance degradation
- Unusual trading patterns

## Implementation Plan

### Phase 1: Basic Metrics
- [ ] Prometheus client in each service
- [ ] Basic health endpoints
- [ ] Connection status tracking
- [ ] Simple Grafana dashboard

### Phase 2: Trading Metrics
- [ ] Order tracking
- [ ] Performance measurements
- [ ] Custom business metrics
- [ ] Advanced dashboards

### Phase 3: Optimization
- [ ] Performance profiling
- [ ] Resource optimization
- [ ] Predictive alerts
- [ ] Capacity planning

## Quick Start

1. **Start monitoring stack**:
   ```bash
   docker-compose up prometheus grafana
   ```

2. **Access dashboards**:
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

3. **Import dashboards**:
   - Located in `monitoring/grafana/dashboards/`

4. **Configure alerts**:
   - Edit `monitoring/prometheus/alerts.yml`

## Best Practices

1. **Metric Naming**: Follow Prometheus conventions
   - `service_subsystem_metric_unit`
   - Use labels for dimensions

2. **Dashboard Design**: Keep it simple
   - Overview → Detail flow
   - Use consistent color schemes
   - Include helpful annotations

3. **Alert Fatigue**: Avoid over-alerting
   - Alert on symptoms, not causes
   - Include remediation steps
   - Use appropriate severity levels

4. **Performance**: Minimize overhead
   - Use histograms for latencies
   - Batch metric updates
   - Reasonable retention periods

## Debugging Guide

When something goes wrong:

1. **Check System Health Dashboard**
   - Are all services green?
   - Any recent error spikes?

2. **Review Logs**
   - Container logs: `docker logs [container]`
   - Aggregated in Loki (if enabled)

3. **Analyze Metrics**
   - Look for anomalies before the issue
   - Compare with historical data

4. **Test Metrics Endpoint**
   ```bash
   curl http://localhost:8080/metrics
   ```

Remember: Good monitoring is like a good co-pilot - it tells you what you need to know when you need to know it. 📊✨