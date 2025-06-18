# IBKR Scanner Go Client SDK

Official Go client SDK for the IBKR Scanner API.

## Installation

```bash
go get github.com/ibkr-trader/scanner/client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ibkr-trader/scanner/client"
)

func main() {
    // Create client
    cfg := client.DefaultConfig()
    c := client.NewClient(cfg)
    
    // Scan a symbol
    result, err := c.ScanSymbol(context.Background(), "AAPL", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d spreads for %s\n", len(result.Spreads), result.Symbol)
}
```

## Configuration

```go
cfg := client.Config{
    BaseURL: "https://api.ibkr-trader.com/v1",
    APIKey:  "your-api-key",
    Timeout: 30 * time.Second,
}

c := client.NewClient(cfg)
```

## Features

### Health Check

```go
health, err := c.Health(ctx)
fmt.Printf("Service status: %s\n", health.Status)
```

### Single Symbol Scan

```go
filters := &client.ScanFilters{
    DeltaMin: 0.20,
    DeltaMax: 0.35,
    DTEMin:   30,
    DTEMax:   60,
}

result, err := c.ScanSymbol(ctx, "AAPL", filters)
```

### Multiple Symbol Scan

```go
symbols := []string{"AAPL", "MSFT", "GOOGL"}
filters := map[string]interface{}{
    "delta": map[string]float64{"min": 0.20, "max": 0.35},
    "dte":   map[string]int{"min": 30, "max": 60},
}

results, err := c.ScanMultiple(ctx, symbols, filters)
```

### Filter Management

```go
// Get current filters
config, err := c.GetFilters(ctx)

// Update filters
newConfig := &client.FilterConfig{
    Delta: &client.DeltaFilter{Min: 0.25, Max: 0.35},
    DTE:   &client.DTEFilter{Min: 30, Max: 60},
}
err = c.UpdateFilters(ctx, newConfig)

// Get presets
presets, err := c.GetPresets(ctx)

// Create preset
preset := &client.PresetRequest{
    Name:        "My Strategy",
    Description: "Custom trading strategy",
    Filters:     *newConfig,
    Tags:        []string{"custom", "moderate"},
}
id, err := c.CreatePreset(ctx, preset)
```

### Real-time Streaming

```go
// Connect to WebSocket
stream, err := c.Connect(ctx)
defer stream.Close()

// Register handlers
stream.OnMessage("scan_result", func(msgType string, payload json.RawMessage) {
    var result client.ScanResult
    json.Unmarshal(payload, &result)
    fmt.Printf("New scan result for %s\n", result.Symbol)
})

// Subscribe to symbols
err = stream.Subscribe([]string{"AAPL", "TSLA"}, nil)
```

### Analytics & History

```go
// Get statistics
stats, err := c.GetStatistics(ctx, "AAPL", "spread_width", "24h")

// Get historical data
params := &client.HistoryParams{
    Symbol:   "AAPL",
    Page:     1,
    PageSize: 20,
}
history, err := c.GetHistory(ctx, params)
```

## Error Handling

The client returns typed errors for API responses:

```go
result, err := c.ScanSymbol(ctx, "INVALID", nil)
if err != nil {
    if apiErr, ok := err.(*client.ErrorResponse); ok {
        fmt.Printf("API Error %d: %s\n", apiErr.Status, apiErr.Error)
    } else {
        fmt.Printf("Network Error: %v\n", err)
    }
}
```

## Examples

See the [example directory](example/) for complete usage examples.

## Filter Reference

### Delta Filter
- `min`: Minimum delta value (0-1)
- `max`: Maximum delta value (0-1)

### DTE Filter
- `min`: Minimum days to expiration
- `max`: Maximum days to expiration

### Liquidity Filter
- `min_open_interest`: Minimum open interest
- `min_volume`: Minimum daily volume

### Spread Filter
- `min_credit`: Minimum net credit
- `max_width`: Maximum spread width
- `min_risk_reward`: Minimum risk/reward ratio

### Advanced Filter
- `min_pop`: Minimum probability of profit
- `max_bid_ask_spread`: Maximum bid-ask spread
- `min_iv_percentile`: Minimum IV percentile

## WebSocket Message Types

### Outgoing
- `subscribe`: Subscribe to symbols
- `unsubscribe`: Unsubscribe from symbols
- `update_filters`: Update real-time filters
- `scan`: Request immediate scan
- `ping`: Keep-alive ping

### Incoming
- `welcome`: Connection established
- `scan_result`: New scan results
- `subscribed`: Subscription confirmed
- `unsubscribed`: Unsubscription confirmed
- `filters_updated`: Filter update confirmed
- `error`: Error message
- `pong`: Keep-alive response

## Rate Limits

- HTTP API: 100 requests per minute
- WebSocket: 1000 concurrent connections
- Batch scans: 100 symbols per request

## Support

For issues and feature requests, please visit:
https://github.com/ibkr-trader/scanner/issues