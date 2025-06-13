# IBKR Spread Automation 🚀

An elegant, event-driven system for automated vertical spread options trading through Interactive Brokers.

## Vision

Transform complex options trading into a flowing, intuitive experience where technology amplifies human decision-making rather than replacing it.

## Features

- **Automated Vertical Spread Trading** - Focus on high-probability debit spreads with intelligent credit spread switching
- **Real-time Options Scanning** - High-performance Go scanner with customizable filters
- **Event-Driven Architecture** - Built on ib-insync's async patterns for maximum efficiency
- **Comprehensive Monitoring** - Prometheus/Grafana dashboards for complete system observability
- **Smart Rate Limiting** - Automatic request throttling with queue management
- **Market Data Management** - LRU subscription cache within TWS limits
- **3-Click Interface** - Intuitive GUI where any action is 3 clicks away

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Windows GUI Application                      │
│                   (Go Backend + Svelte Frontend)                 │
└─────────────────────┬───────────────────┬───────────────────────┘
                      │                   │
                 WebSocket            REST API
                      │                   │
┌─────────────────────┴───────────────────┴───────────────────────┐
│                      Docker Container Network                     │
├─────────────────────────────────┬───────────────────────────────┤
│  Python IBKR Interface (Async)  │      Go Scanner Engine         │
│  - Event-driven architecture    │   - Request coordination      │
│  - ib-insync with Watchdog     │   - Backpressure handling     │
│  - Subscription management      │   - High-performance filter   │
└─────────────────────────────────┴───────────────────────────────┘
                      │
             TCP Socket (Async)
                      │
              ┌───────┴──────┐
              │   TWS/IB     │
              │   Gateway    │
              └──────────────┘
```

## Quick Start

### Prerequisites

1. **Interactive Brokers Account** with options trading permissions
2. **TWS or IB Gateway** installed and configured
3. **Docker** and **Docker Compose**
4. **Market Data Subscription** (for real-time data)

### TWS Configuration

1. Enable API connections in TWS:
   - File → Global Configuration → API → Settings
   - ✅ Enable ActiveX and Socket Clients
   - ❌ Read-Only API (must be unchecked)
   - ✅ Download open orders on connection
   - ✅ Include market data in snapshot

2. Configure ports:
   - Socket port: 7497 (paper) or 7496 (live)
   - Add `127.0.0.1` to Trusted IPs

3. Set memory allocation to 4GB minimum

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/ibkr-spread-automation.git
cd ibkr-spread-automation

# Copy environment template
cp .env.example .env

# Start services
docker-compose up -d

# Check health
docker-compose ps
curl http://localhost:8080/health
```

### First Run

1. Test connection:
   ```bash
   docker-compose run --rm python-ibkr python experiments/sandbox/test_connection.py
   ```

2. Open GUI:
   ```
   http://localhost:3000
   ```

3. View monitoring:
   ```
   http://localhost:3001  # Grafana (admin/admin)
   http://localhost:9090  # Prometheus
   ```

## Development

### Project Structure

```
.
├── src/
│   ├── python/         # IBKR interface service
│   ├── go/            # High-performance scanner
│   └── gui/           # Svelte frontend + Go backend
├── docker/            # Container configurations
├── experiments/       # Safe testing playground
├── monitoring/        # Dashboards and alerts
├── flow_journal/      # Development diary
├── .vibe/            # Templates and inspiration
└── docs/             # Living documentation
```

### Core Principles

1. **The One Rule**: Never block the event loop
2. **Events Over Polling**: React to changes, don't ask for them
3. **Monitor Everything**: If it matters, measure it
4. **Flow State First**: Match tasks to energy levels

### Running Tests

```bash
# Python tests
docker-compose run --rm python-ibkr pytest

# Go tests  
docker-compose run --rm go-scanner go test ./...

# Integration tests
make test-integration
```

## Configuration

### Scanner Filters

All standard options parameters are supported:
- Greeks (delta, gamma, theta, vega)
- Days to Expiration (DTE)
- Implied Volatility (level and percentile)
- Liquidity metrics (volume, open interest)
- Bid-ask spread limits
- Probability metrics (ITM, PoP)
- And many more...

### Environment Variables

See `.env.example` for all configuration options.

Key settings:
- `TWS_HOST`: TWS hostname (use `host.docker.internal` for Docker)
- `TWS_PORT`: 7497 (paper) or 7496 (live)
- `MAX_SUBSCRIPTIONS`: Market data line limit (default: 90)
- `CLIENT_ID`: Unique ID per connection (max 32)

## Monitoring

The system includes comprehensive monitoring:

- **Connection Health**: TWS connection status and uptime
- **API Usage**: Request rates and throttling events  
- **Market Data**: Subscription usage and evictions
- **Order Execution**: Fill rates and timing
- **System Performance**: CPU, memory, event loop health

Access dashboards at `http://localhost:3001`

## Troubleshooting

### Connection Issues
1. Verify TWS is running and logged in
2. Check API settings in TWS
3. Confirm ports are accessible
4. Review logs: `docker-compose logs python-ibkr`

### Rate Limiting
- System automatically handles throttling
- Monitor dashboard shows current usage
- Adjust `MAX_CONCURRENT_REQUESTS` if needed

### Market Data Limits
- Check subscription level in TWS
- Monitor active subscriptions in dashboard
- System uses LRU eviction when at capacity

## Contributing

We follow vibe coding principles:

1. Read `RULES.md` before starting
2. Check `flow_journal/` for context
3. Use `experiments/` for testing ideas
4. Update docs as you code
5. Maintain flow state!

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [ib-insync](https://github.com/erdewit/ib_insync) - the excellent async IB API wrapper
- Inspired by flow state development practices
- Designed for traders who value both automation and control

---

*"The best trading system feels like an extension of the trader's mind, not a replacement for it."*

Ready to trade with flow? Let's go! 🚀