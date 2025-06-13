# System Monitoring Strategy

## Overview

Real-time monitoring is critical for IBKR trading systems. This directory contains configurations and dashboards for comprehensive system observability.

## Key Metrics to Track

### 1. Connection Health
- **TWS Connection Status**: Binary up/down
- **Connection Duration**: Time since last disconnect
- **Reconnection Count**: Daily reconnections
- **Watchdog Status**: Active/inactive

### 2. API Usage
- **Request Rate**: Requests per second (target: <45)
- **Throttle Events**: Count and duration
- **Queue Depth**: Pending requests
- **Error Rate**: By error code

### 3. Market Data
- **Active Subscriptions**: Current vs max
- **Subscription Churn**: Evictions per minute
- **Data Latency**: Time from request to receipt
- **Cache Hit Rate**: Percentage of cached responses

### 4. Order Execution
- **Order Fill Rate**: Successful fills percentage
- **Execution Time**: Order submission to fill
- **Rejection Rate**: By reason
- **Active Orders**: Current count

### 5. System Performance
- **CPU Usage**: By service
- **Memory Usage**: Especially Python service
- **Event Loop Lag**: Python async health
- **Network Latency**: Between services

## Dashboard Layout

### Main Dashboard
```
┌─────────────────────────────────────────────────────┐
│                  System Health                       │
├─────────────────┬─────────────────┬─────────────────┤
│ TWS Connection  │ API Usage       │ Active Orders   │
│ ✅ Connected    │ 32/45 req/s     │ 3 pending       │
├─────────────────┴─────────────────┴─────────────────┤
│                Market Data Usage                     │
│ ████████████████████░░░░░░  72/90 subscriptions    │
├─────────────────────────────────────────────────────┤
│                  Recent Alerts                       │
│ ⚠️ 14:32 - Throttling started (2s)                 │
│ ℹ️ 14:28 - Subscription evicted: SPY_20240315_C400 │
└─────────────────────────────────────────────────────┘
```

### Detailed Metrics
- Request latency histogram
- Error rate by type
- Subscription lifetime distribution
- Order execution timeline

## Alert Rules

### Critical Alerts
1. **TWS Disconnected** > 30 seconds
2. **Error Rate** > 5% over 1 minute
3. **Queue Depth** > 200 requests
4. **No Market Data** > 10 seconds

### Warning Alerts
1. **Subscription Usage** > 80%
2. **Throttle Events** > 5 per minute
3. **Memory Usage** > 80%
4. **Order Rejection Rate** > 10%

### Info Alerts
1. Daily TWS restart approaching
2. New error code detected
3. Performance degradation
4. Unusual trading patterns

## Implementation

### Prometheus Metrics
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'python-ibkr'
    static_configs:
      - targets: ['python-ibkr:8080']
  
  - job_name: 'go-scanner'
    static_configs:
      - targets: ['go-scanner:8081']
```

### Grafana Dashboards
1. `system-overview.json` - Main health dashboard
2. `api-usage.json` - Detailed API metrics
3. `market-data.json` - Subscription management
4. `order-execution.json` - Trading performance
5. `alerts-config.json` - Alert configuration

## Best Practices

1. **Monitor First, Alert Second**: Observe patterns before setting thresholds
2. **Context in Alerts**: Include relevant data in alert messages
3. **Dashboard Hierarchy**: Overview → Component → Detail
4. **Regular Reviews**: Weekly threshold adjustments based on patterns
5. **Correlation**: Link metrics to identify root causes

## Troubleshooting Guide

### High Throttle Rate
1. Check scanner request patterns
2. Verify batching is working
3. Review subscription churn
4. Consider increasing delays

### Connection Issues
1. Verify TWS is running
2. Check network connectivity
3. Review Watchdog logs
4. Confirm port accessibility

### Memory Growth
1. Check subscription cleanup
2. Review event handler efficiency
3. Verify cache eviction
4. Look for memory leaks

---

Remember: Good monitoring prevents bad surprises. When in doubt, add a metric!