# IBKR Spread Automation - Quick Start Guide

Welcome! This guide will get you up and running with the IBKR Spread Automation system in under 30 minutes.

## Prerequisites

- Windows 10/11 (64-bit)
- Docker Desktop installed and running
- Interactive Brokers TWS installed
- IBKR account with API access enabled
- 8GB RAM minimum (16GB recommended)
- Git installed

## 1. Clone and Setup (5 minutes)

```bash
# Clone the repository
git clone https://github.com/your-org/ibkr-spread-automation.git
cd ibkr-spread-automation

# Create environment file
cp .env.example .env
```

Edit `.env` with your settings:
```env
# TWS Connection
TWS_HOST=host.docker.internal
TWS_PORT=7497  # 7497 for paper, 7496 for live
TWS_CLIENT_ID=1

# Trading Settings
TRADING_MODE=paper  # paper or live
MAX_POSITION_SIZE=10000
MAX_DAILY_TRADES=50

# Scanner Settings
SCANNER_INTERVAL_SECONDS=60
MIN_OPEN_INTEREST=100
MIN_VOLUME=1000
```

## 2. Configure TWS (5 minutes)

1. **Start TWS** and log in to your account

2. **Enable API Access**:
   - File → Global Configuration → API → Settings
   - ✅ Enable ActiveX and Socket Clients
   - ✅ Socket port: 7497 (paper) or 7496 (live)
   - ✅ Allow connections from localhost only

3. **Set Memory** (critical for stability):
   - Configure → Settings → Memory
   - Initial heap: 2048 MB
   - Maximum heap: 4096 MB

4. **Disable Precautions** (for automation):
   - Configure → Settings → API → Precautions
   - ✅ Bypass Order Precautions for API Orders

## 3. Start the System (5 minutes)

```bash
# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps

# Expected output:
# ibkr-python     running   0.0.0.0:8001->8001/tcp
# ibkr-scanner    running   0.0.0.0:8002->8002/tcp
# prometheus      running   0.0.0.0:9090->9090/tcp
# grafana         running   0.0.0.0:3000->3000/tcp
```

## 4. Verify Connection (2 minutes)

```bash
# Check TWS connection
curl http://localhost:8001/health

# Expected response:
{
  "status": "healthy",
  "tws_connected": true,
  "account": "DU1234567",
  "next_order_id": 1
}

# Test scanner
curl http://localhost:8002/scan/test

# Expected response:
{
  "status": "success",
  "results": [...]
}
```

## 5. Access Dashboards (2 minutes)

1. **Grafana Monitoring**: http://localhost:3000
   - Username: admin
   - Password: admin (change on first login)
   - Navigate to: Dashboards → IBKR System Overview

2. **Prometheus Metrics**: http://localhost:9090
   - Check targets: Status → Targets
   - All should show as "UP"

## 6. Run Your First Scan (5 minutes)

Create a file `first_scan.json`:
```json
{
  "underlying": "SPY",
  "expiry_days": [30, 60],
  "filters": {
    "min_delta": 0.2,
    "max_delta": 0.4,
    "min_open_interest": 1000,
    "min_volume": 100,
    "max_spread_width": 0.10
  },
  "spread_type": "put_credit"
}
```

Run the scan:
```bash
curl -X POST http://localhost:8002/scan \
  -H "Content-Type: application/json" \
  -d @first_scan.json
```

## 7. Common Operations

### View Logs
```bash
# Python service logs
docker-compose logs -f ibkr-python

# Scanner logs
docker-compose logs -f ibkr-scanner

# All services
docker-compose logs -f
```

### Stop Services
```bash
# Stop all services
docker-compose down

# Stop and remove volumes (full reset)
docker-compose down -v
```

### Update Configuration
```bash
# Edit configuration
nano .env

# Restart services to apply
docker-compose restart
```

## Troubleshooting

### TWS Connection Issues

1. **"No connection to TWS"**
   - Ensure TWS is running and logged in
   - Check firewall settings
   - Verify port numbers match

2. **"Socket connection broken"**
   - Restart TWS
   - Check memory settings
   - Review logs for details

3. **"Rate limit exceeded"**
   - Reduce scanner frequency
   - Check subscription limits
   - Implement request throttling

### Docker Issues

1. **"Cannot connect to Docker daemon"**
   - Ensure Docker Desktop is running
   - Run as administrator (Windows)

2. **"Port already in use"**
   - Check for conflicting services
   - Change ports in docker-compose.yml

## Next Steps

1. **Read the Documentation**:
   - [Architecture Overview](documentation/architecture/EVENT_DRIVEN_ARCHITECTURE.md)
   - [TWS Configuration Guide](documentation/tws-setup/TWS_CONFIGURATION_COMPLETE.md)
   - [Scanner Strategies](documentation/scanners/SCANNER_STRATEGIES.md)

2. **Configure Scanners**:
   - Create custom scan configurations
   - Set up automated scanning schedules
   - Define trading strategies

3. **Set Up Alerts**:
   - Configure Grafana alerts
   - Set up email/SMS notifications
   - Define risk thresholds

4. **Paper Trade First**:
   - Test all strategies in paper account
   - Verify order execution logic
   - Monitor system performance

## Quick Commands Reference

```bash
# Service Management
docker-compose up -d          # Start all services
docker-compose down           # Stop all services
docker-compose restart        # Restart all services
docker-compose logs -f        # View all logs

# Health Checks
curl http://localhost:8001/health    # Python service
curl http://localhost:8002/health    # Scanner service
curl http://localhost:9090/-/healthy # Prometheus
curl http://localhost:3000/api/health # Grafana

# Scanner Operations
curl http://localhost:8002/scan/list    # List active scans
curl http://localhost:8002/scan/results  # Get latest results
curl http://localhost:8002/scan/stop     # Stop all scans

# System Info
docker-compose ps             # Service status
docker stats                  # Resource usage
docker-compose top            # Running processes
```

## Support

- **Documentation**: See `/documentation` folder
- **Issues**: GitHub Issues page
- **Logs**: Check service logs first
- **Community**: Discord/Slack channel

---

🚀 **You're ready to start automated options trading!** Remember to always test thoroughly in paper trading before going live.