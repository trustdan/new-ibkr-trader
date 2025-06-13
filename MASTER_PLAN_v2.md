# 🚀 IBKR Spread Automation - Master Plan v2.0
## Updated with TWS API Requirements & ib-insync Best Practices

## 📋 Executive Summary

This updated master plan incorporates critical adjustments based on TWS API documentation and ib-insync library patterns. The key change is embracing an **event-driven, async architecture** throughout the Python service, with proper handling of market data limits, rate limiting, and connection management.

### Updated Architecture
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

### Critical Changes from v1
1. **Async Everything**: Python service built on asyncio/ib-insync patterns
2. **Event-Driven**: Replace polling with event subscriptions
3. **Built-in Throttling**: Use ib-insync's rate limiting, not custom
4. **Subscription Manager**: Track market data line usage
5. **Request Coordinator**: Manage flow between Go scanner and Python

---

## 🌊 Phase 0: Foundation & Vibe Setup (5 Days)

### Day 1: Project Genesis & Vibe Initialization
**Morning (4 hours)**
```gherkin
Feature: Project Foundation Creation
  As a vibe-conscious developer
  I want to establish a project structure that promotes flow state
  So that development feels natural and productive

  Scenario: Initialize vibe-friendly repository
    Given I have an empty project directory
    When I create the foundational structure
    Then I have these directories:
      | Directory           | Purpose                           |
      | src/               | Clean production code             |
      | experiments/       | Safe playground for ideas         |
      | docs/              | Living documentation              |
      | .vibe/             | Flow logs and inspiration         |
      | flow_journal/      | Daily development insights        |
      | docker/            | Container configurations          |
      | tests/             | Test suites                       |
      | monitoring/        | Dashboards and metrics            |
    And I have essential files:
      | File               | Content                           |
      | .gitignore         | Proper exclusions                 |
      | LICENSE            | MIT License                       |
      | CHANGELOG.md       | Version history template          |
      | IDEAS.md           | Future feature brainstorming      |
      | ADR/               | Architecture decisions            |
      | MONITORING.md      | System metrics tracking           |
```

**Afternoon (4 hours)**
- Create `.vibe/templates/` with async code snippets
- Set up `flow_journal/template.md` for daily entries
- Initialize git repository with meaningful first commit
- Create `experiments/async-patterns/` for testing ib-insync
- Document project philosophy in `.vibe/manifesto.md`
- Add `monitoring/README.md` for metrics strategy

### Day 2: Docker Environment Architecture (Async-First)
**Morning (4 hours)**
```gherkin
Feature: Async-First Container Environment
  As a developer
  I want containers optimized for event-driven architecture
  So that the system works harmoniously with TWS API

  Scenario: Create async Python IBKR container
    Given I need event-driven TWS connection
    When I create docker/python-ibkr/Dockerfile
    Then it includes:
      | Component           | Version/Config                    |
      | Base Image         | python:3.11-slim                  |
      | ib-insync          | 0.9.86                           |
      | asyncio            | Core event loop                  |
      | uvloop             | High-performance event loop      |
      | aiohttp            | Async HTTP server                |
      | prometheus-client  | Metrics export                   |
    And environment variables:
      | Variable           | Default                           |
      | TWS_HOST          | host.docker.internal              |
      | TWS_PORT          | 7497                              |
      | CLIENT_ID         | 1                                 |
      | WATCHDOG_TIMEOUT  | 60                                |
      | MAX_SUBSCRIPTIONS | 90                                |
      | ACCOUNT_TYPE      | paper                             |
```

**Afternoon (4 hours)**
```dockerfile
# docker/python-ibkr/Dockerfile
FROM python:3.11-slim

# Install uvloop for performance
RUN pip install uvloop aiohttp ib-insync==0.9.86 \
    prometheus-client aiocache msgpack

# Set up async-friendly environment
ENV PYTHONUNBUFFERED=1
ENV IB_ASYNC_MODE=1

COPY src/python /app
WORKDIR /app

# Health check that doesn't block event loop
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD python -c "import aiohttp; aiohttp.ClientSession().get('http://localhost:8080/health')"

CMD ["python", "-m", "uvloop", "main.py"]
```

### Day 3: Documentation Framework & TWS Setup Guide
**Morning (4 hours)**
```gherkin
Feature: TWS-Specific Documentation
  As a developer
  I want clear TWS configuration docs
  So that setup is foolproof

  Scenario: Create comprehensive TWS guide
    Given TWS has strict requirements
    When I document setup procedures
    Then docs/setup/TWS_CONFIGURATION.md includes:
      | Section                    | Critical Details                |
      | Socket Configuration      | Enable ActiveX and Socket       |
      | Read-Only Mode           | MUST be disabled                |
      | Memory Settings          | 4GB minimum                     |
      | Daily Restart Time       | Configure for 11:45 PM EST      |
      | Port Setup               | 7497 (paper), 7496 (live)      |
      | Trusted IPs              | Add Docker network range        |
      | Market Data Lines        | Note subscription level         |
      | API Precautions          | Bypass all warnings            |
      | Master Client ID         | Reserve 0 for manual orders    |
```

**Afternoon (4 hours)**
- Create `docs/architecture/event-driven-design.md`
- Write `docs/api/subscription-limits.md`
- Document order ID management in `docs/trading/order-management.md`
- Create `ADR/001-async-architecture.md`
- Add `docs/monitoring/metrics.md`

### Day 4: Development Tools & Async Scripts
**Morning (4 hours)**
```python
# scripts/test_connection.py
import asyncio
from ib_insync import IB, util

async def test_connection():
    """Test TWS connection with proper async handling"""
    ib = IB()
    try:
        await ib.connectAsync('localhost', 7497, clientId=999)
        print(f"Connected: {ib.isConnected()}")
        
        # Test event system
        ib.errorEvent += lambda reqId, err, errStr, contract: 
            print(f"Error {err}: {errStr}")
            
        await ib.sleep(2)  # Proper async sleep
        
    finally:
        ib.disconnect()

if __name__ == '__main__':
    util.run(test_connection())
```

**Afternoon (4 hours)**
```makefile
# Makefile with async-aware commands
.PHONY: dev test monitor clean

dev:
	@echo "Starting async development environment..."
	docker-compose up -d
	@echo "Waiting for services..."
	@sleep 5
	@make health-check

test:
	@echo "Running async tests..."
	docker-compose run --rm python-ibkr pytest -v --asyncio-mode=auto

monitor:
	@echo "Opening monitoring dashboards..."
	open http://localhost:9090  # Prometheus
	open http://localhost:3000  # Grafana

health-check:
	@docker-compose ps
	@curl -s http://localhost:8080/health | jq
```

### Day 5: Environment Validation & Flow Check
**Morning (4 hours)**
```gherkin
Feature: Async Environment Validation
  As a developer
  I want to verify async patterns work correctly
  So that I avoid blocking the event loop

  Scenario: Validate event-driven setup
    Given I have async containers running
    When I run validation tests
    Then:
      | Check                    | Expected Result              |
      | Event loop running      | No blocking operations       |
      | IB connection async     | Connects without blocking    |
      | Watchdog active         | Auto-reconnects on failure   |
      | Events firing           | Callbacks execute properly   |
      | Metrics exported        | Prometheus scraping works    |
```

**Afternoon (4 hours)**
- Create first flow journal entry
- Test ib-insync examples in `experiments/`
- Document async gotchas discovered
- Set up monitoring dashboard templates
- Review ib-insync event patterns

---

## 🔌 Phase 1: IBKR Connection Layer - Event-Driven (12 Days)

### Day 6-7: Async TWS Connection Foundation
**Day 6 Morning (4 hours)**
```python
# src/python/core/connection.py
import asyncio
from typing import Optional
from ib_insync import IB, Contract, Trade, util
from ib_insync.ibcontroller import Watchdog
import logging

class AsyncIBKRService:
    """Event-driven IBKR service using ib-insync patterns"""
    
    def __init__(self):
        self.ib = IB()
        self.watchdog: Optional[Watchdog] = None
        self.next_order_id: int = 0
        self.subscriptions = {}  # Track market data subscriptions
        self._setup_event_handlers()
        
    def _setup_event_handlers(self):
        """Configure all event handlers"""
        # Connection events
        self.ib.connectedEvent += self._on_connected
        self.ib.disconnectedEvent += self._on_disconnected
        
        # Order events  
        self.ib.orderStatusEvent += self._on_order_status
        self.ib.execDetailsEvent += self._on_exec_details
        
        # Market data events
        self.ib.pendingTickersEvent += self._on_pending_tickers
        
        # Error handling
        self.ib.errorEvent += self._on_error
        
        # Throttling events
        self.ib.client.throttleStart += self._on_throttle_start
        self.ib.client.throttleEnd += self._on_throttle_end
        
    async def start(self, host='localhost', port=7497, client_id=1):
        """Start service with watchdog"""
        # Configure watchdog for auto-reconnection
        self.watchdog = IB.Watchdog(
            controller=self.ib,
            host=host,
            port=port,
            clientId=client_id,
            connectTimeout=10,
            appStartupTime=15,
            readonly=False
        )
        
        # Start watchdog (handles connection)
        self.watchdog.start()
        
        # Wait for initial connection
        await self._wait_for_connection()
```

**Day 6 Afternoon (4 hours)**
```python
    # Event handlers
    def _on_connected(self):
        """Handle connection established"""
        logging.info("Connected to TWS")
        self.next_order_id = self.ib.client.getReqId()
        
    def _on_disconnected(self):
        """Handle disconnection"""
        logging.warning("Disconnected from TWS")
        # Watchdog will handle reconnection
        
    def _on_error(self, reqId, errorCode, errorString, contract):
        """Handle API errors with specific actions"""
        if errorCode == 1100:
            logging.error("Connectivity lost - Watchdog will reconnect")
        elif errorCode == 100:
            logging.warning("Pacing violation - throttling active")
        elif errorCode == 2110:
            logging.warning("Connectivity restored")
        else:
            logging.error(f"Error {errorCode}: {errorString}")
            
    async def _wait_for_connection(self, timeout=30):
        """Wait for connection to be established"""
        start = asyncio.get_event_loop().time()
        while not self.ib.isConnected():
            if asyncio.get_event_loop().time() - start > timeout:
                raise TimeoutError("Connection timeout")
            await self.ib.sleep(0.1)
```

**Day 7 Morning (4 hours)**
```gherkin
Feature: Robust Connection Management
  As the IBKR service
  I want automatic connection recovery
  So that daily TWS restarts don't break the system

  Scenario: Handle daily TWS restart
    Given TWS is configured to restart at 11:45 PM EST
    When the restart time approaches
    Then the watchdog:
      | Time      | Action                           |
      | 11:40 PM  | Log upcoming restart            |
      | 11:45 PM  | Detect disconnection            |
      | 11:45 PM  | Begin reconnection attempts     |
      | 11:50 PM  | Restore connection              |
      | 11:51 PM  | Resubscribe market data        |
      | 11:52 PM  | Resume normal operation        |
```

**Day 7 Afternoon (4 hours)**
- Implement subscription restoration after reconnect
- Add connection state persistence
- Create health check endpoint
- Test watchdog recovery scenarios

### Day 8-9: Market Data Subscription Management
**Day 8 Morning (4 hours)**
```python
# src/python/core/subscription_manager.py
from collections import OrderedDict
from typing import Dict, Set
import asyncio

class SubscriptionManager:
    """Manages market data subscriptions within TWS limits"""
    
    def __init__(self, ib, max_lines=90):  # Leave headroom
        self.ib = ib
        self.max_lines = max_lines
        self.active_subscriptions: OrderedDict[str, Contract] = OrderedDict()
        self.subscription_counts: Dict[str, int] = {}
        
    async def subscribe(self, contract: Contract) -> bool:
        """Subscribe with automatic eviction if needed"""
        key = self._get_key(contract)
        
        if key in self.active_subscriptions:
            # Move to end (LRU)
            self.active_subscriptions.move_to_end(key)
            return True
            
        # Check if we need to evict
        if len(self.active_subscriptions) >= self.max_lines:
            # Evict least recently used
            oldest_key, oldest_contract = self.active_subscriptions.popitem(False)
            self.ib.cancelMktData(oldest_contract)
            logging.info(f"Evicted subscription: {oldest_key}")
            
        # Subscribe
        self.ib.reqMktData(contract, '', False, False)
        self.active_subscriptions[key] = contract
        return True
        
    def get_usage(self) -> dict:
        """Get subscription metrics"""
        return {
            'active': len(self.active_subscriptions),
            'max': self.max_lines,
            'usage_pct': len(self.active_subscriptions) / self.max_lines * 100
        }
```

**Day 8 Afternoon (4 hours)**
```gherkin
Feature: Smart Subscription Management
  As the market data manager
  I want to maximize data availability
  So that scanners get fresh data within limits

  Scenario: LRU eviction when at capacity
    Given I have 90 active subscriptions (limit)
    When a new subscription is requested
    Then the least recently used is cancelled
    And the new subscription is added
    And total stays within limits
    
  Scenario: Subscription metrics tracking
    Given I need to monitor usage
    When I query subscription status
    Then I see:
      | Metric           | Value                |
      | Active Count     | Current subscriptions |
      | Max Allowed      | Account limit        |
      | Usage Percentage | Visual indicator     |
      | Eviction Count   | Total evictions      |
```

**Day 9 Morning (4 hours)**
- Implement subscription pooling for similar contracts
- Add subscription request queuing
- Create usage monitoring endpoint
- Test with various account types

**Day 9 Afternoon (4 hours)**
- Optimize subscription patterns
- Document best practices
- Create subscription dashboard
- Flow journal on data management

### Day 10-11: Request Coordination & Rate Limiting
**Day 10 Morning (4 hours)**
```python
# src/python/core/request_coordinator.py
import asyncio
from asyncio import Queue, Semaphore
from dataclasses import dataclass
from typing import Any, Callable

@dataclass
class Request:
    """Async request wrapper"""
    func: Callable
    args: tuple
    kwargs: dict
    future: asyncio.Future
    priority: int = 0

class RequestCoordinator:
    """Coordinates requests between services respecting ib-insync throttling"""
    
    def __init__(self, ib):
        self.ib = ib
        self.request_queue: Queue[Request] = Queue()
        self.processing = False
        self.metrics = {
            'total_requests': 0,
            'throttle_events': 0,
            'current_queue_size': 0
        }
        
        # Monitor throttling
        ib.client.throttleStart += self._on_throttle_start
        ib.client.throttleEnd += self._on_throttle_end
        
    async def submit_request(self, func, *args, priority=0, **kwargs):
        """Submit request for coordinated execution"""
        future = asyncio.Future()
        request = Request(func, args, kwargs, future, priority)
        
        await self.request_queue.put(request)
        self.metrics['current_queue_size'] = self.request_queue.qsize()
        
        if not self.processing:
            asyncio.create_task(self._process_requests())
            
        return await future
        
    async def _process_requests(self):
        """Process queued requests with respect to throttling"""
        self.processing = True
        
        while not self.request_queue.empty():
            request = await self.request_queue.get()
            
            try:
                # Let ib-insync handle rate limiting
                result = await request.func(*request.args, **request.kwargs)
                request.future.set_result(result)
                
            except Exception as e:
                request.future.set_exception(e)
                
            self.metrics['total_requests'] += 1
            self.metrics['current_queue_size'] = self.request_queue.qsize()
            
            # Small delay to be nice to the API
            await self.ib.sleep(0.01)
            
        self.processing = False
```

**Day 10 Afternoon (4 hours)**
- Implement request batching for similar operations
- Add priority queue for time-sensitive requests
- Create backpressure mechanism for Go scanner
- Test under heavy load

**Day 11 Morning (4 hours)**
```gherkin
Feature: Intelligent Request Coordination
  As the system coordinator
  I want smooth request flow
  So that we maximize throughput without violations

  Scenario: Handle burst from scanner
    Given the Go scanner sends 100 requests
    When they arrive simultaneously
    Then the coordinator:
      | Action               | Result                    |
      | Queues requests      | All accepted             |
      | Processes serially   | Respects rate limits     |
      | Monitors throttling  | Adjusts if needed        |
      | Provides backpressure| Scanner slows if needed  |
```

**Day 11 Afternoon (4 hours)**
- Add request deduplication
- Implement smart batching
- Create performance metrics
- Update flow journal

### Day 12-13: Order Execution Engine (Async)
**Day 12 Morning (4 hours)**
```python
# src/python/trading/async_order_engine.py
from ib_insync import Contract, Order, Trade, ComboLeg
import asyncio
from typing import List, Optional

class AsyncOrderEngine:
    """Event-driven order execution engine"""
    
    def __init__(self, ib, coordinator):
        self.ib = ib
        self.coordinator = coordinator
        self.active_trades: Dict[int, Trade] = {}
        
        # Subscribe to order events
        ib.orderStatusEvent += self._on_order_status
        ib.execDetailsEvent += self._on_execution
        ib.commissionReportEvent += self._on_commission
        
    async def execute_vertical_spread(
        self, 
        symbol: str,
        expiry: str,
        long_strike: float,
        short_strike: float,
        right: str = 'C',
        quantity: int = 1
    ) -> Trade:
        """Execute vertical spread with full event handling"""
        
        # Create combo contract
        combo = await self._create_spread_combo(
            symbol, expiry, long_strike, short_strike, right
        )
        
        # Create order
        order = Order(
            action='BUY' if long_strike < short_strike else 'SELL',
            totalQuantity=quantity,
            orderType='LMT',
            lmtPrice=0.0,  # Will be set after preview
            transmit=False  # Don't transmit yet
        )
        
        # Preview order
        preview = await self.coordinator.submit_request(
            self.ib.whatIfOrderAsync, combo, order
        )
        
        # Validate preview
        if not self._validate_preview(preview):
            raise ValueError(f"Order validation failed: {preview}")
            
        # Set limit price based on preview
        order.lmtPrice = self._calculate_limit_price(preview)
        order.transmit = True
        
        # Place order
        trade = await self.coordinator.submit_request(
            self.ib.placeOrderAsync, combo, order
        )
        
        # Track active trade
        self.active_trades[trade.order.orderId] = trade
        
        # Wait for fill or timeout
        await self._wait_for_fill(trade, timeout=60)
        
        return trade
```

**Day 12 Afternoon (4 hours)**
```python
    async def _create_spread_combo(self, symbol, expiry, long_strike, short_strike, right):
        """Create combo contract for vertical spread"""
        # Base option specs
        base_contract = Contract(
            symbol=symbol,
            secType='OPT',
            exchange='SMART',
            currency='USD',
            lastTradeDateOrContractMonth=expiry,
            right=right
        )
        
        # Get full contract details
        long_contract = Contract(**base_contract.dict(), strike=long_strike)
        short_contract = Contract(**base_contract.dict(), strike=short_strike)
        
        # Qualify contracts
        long_details = await self.ib.qualifyContractsAsync(long_contract)
        short_details = await self.ib.qualifyContractsAsync(short_contract)
        
        # Create combo
        combo = Contract(
            symbol=symbol,
            secType='BAG',
            exchange='SMART',
            currency='USD',
            comboLegs=[
                ComboLeg(conId=long_details[0].conId, ratio=1, action='BUY'),
                ComboLeg(conId=short_details[0].conId, ratio=1, action='SELL')
            ]
        )
        
        return combo
        
    def _on_order_status(self, trade: Trade):
        """Handle order status updates"""
        logging.info(f"Order {trade.order.orderId}: {trade.orderStatus.status}")
        
        if trade.orderStatus.status in ['Filled', 'Cancelled', 'ApiCancelled']:
            # Cleanup
            self.active_trades.pop(trade.order.orderId, None)
```

**Day 13 Morning (4 hours)**
```gherkin
Feature: Reliable Spread Execution
  As a trader
  I want confident spread execution
  So that orders fill at good prices

  Scenario: Execute debit spread with preview
    Given I want to buy a call spread
    When I submit the order
    Then the system:
      | Step          | Action                      | Validation           |
      | Build combo   | Create spread contract     | Both legs qualified  |
      | Preview       | whatIfOrder check          | Margin acceptable    |
      | Price calc    | Determine limit price      | Within bid-ask       |
      | Submit        | Place combo order          | Order accepted       |
      | Monitor       | Track via events           | Status updates flow  |
      | Fill          | Both legs execute          | Prices reasonable    |
```

**Day 13 Afternoon (4 hours)**
- Implement OCA groups for risk management
- Add position tracking
- Create execution reports
- Test various spread types

### Day 14-15: Integration Testing & Monitoring
**Day 14 Morning (4 hours)**
```python
# src/python/monitoring/metrics_server.py
from aiohttp import web
from prometheus_client import Counter, Gauge, Histogram, generate_latest

# Define metrics
connection_status = Gauge('ibkr_connection_status', 'TWS connection status')
active_subscriptions = Gauge('ibkr_active_subscriptions', 'Active market data subscriptions')
request_queue_size = Gauge('ibkr_request_queue_size', 'Pending requests in queue')
order_execution_time = Histogram('ibkr_order_execution_seconds', 'Order execution time')
throttle_events = Counter('ibkr_throttle_events_total', 'Total throttle events')

async def metrics_handler(request):
    """Prometheus metrics endpoint"""
    metrics = generate_latest()
    return web.Response(text=metrics.decode('utf-8'), content_type='text/plain')

async def health_handler(request):
    """Health check endpoint"""
    app = request.app
    ib = app['ib_service'].ib
    
    health = {
        'status': 'healthy' if ib.isConnected() else 'unhealthy',
        'connected': ib.isConnected(),
        'subscriptions': app['ib_service'].subscription_manager.get_usage(),
        'queue_size': app['ib_service'].coordinator.metrics['current_queue_size']
    }
    
    return web.json_response(health)
```

**Day 14 Afternoon (4 hours)**
- Create Grafana dashboards
- Set up alerting rules
- Test monitoring under load
- Document metrics

**Day 15 (8 hours)**
- Comprehensive integration tests
- Load testing with realistic scenarios
- Document all event patterns
- Update flow journal with learnings

---

## ⚡ Phase 2: Go Scanner Engine with Coordination (10 Days)

### Day 16-17: Scanner Architecture with Backpressure
**Day 16 Morning (4 hours)**
```go
// src/go/scanner/core/coordinator.go
package core

import (
    "context"
    "sync"
    "time"
)

type RequestCoordinator struct {
    pythonClient  *PythonAPIClient
    maxConcurrent int
    semaphore     chan struct{}
    metrics       *Metrics
    mu            sync.RWMutex
}

func NewRequestCoordinator(client *PythonAPIClient, maxConcurrent int) *RequestCoordinator {
    return &RequestCoordinator{
        pythonClient:  client,
        maxConcurrent: maxConcurrent,
        semaphore:     make(chan struct{}, maxConcurrent),
        metrics:       NewMetrics(),
    }
}

func (rc *RequestCoordinator) RequestMarketData(ctx context.Context, contracts []Contract) error {
    // Implement backpressure
    select {
    case rc.semaphore <- struct{}{}:
        defer func() { <-rc.semaphore }()
        
        // Check Python service queue depth
        queueDepth, err := rc.pythonClient.GetQueueDepth(ctx)
        if err != nil {
            return err
        }
        
        // Apply backpressure if needed
        if queueDepth > 50 {
            delay := time.Duration(queueDepth) * time.Millisecond
            time.Sleep(delay)
        }
        
        // Batch request
        return rc.pythonClient.RequestMarketDataBatch(ctx, contracts)
        
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

**Day 16 Afternoon (4 hours)**
```gherkin
Feature: Coordinated Scanner Operations
  As the Go scanner
  I want to coordinate with Python service
  So that we don't overwhelm the TWS API

  Scenario: Respect Python service capacity
    Given Python service has queue depth of 75
    When scanner wants to send requests
    Then it applies backpressure
    And delays proportionally
    And monitors queue depth
    And adjusts sending rate
```

**Day 17 Morning (4 hours)**
- Implement adaptive rate control
- Add circuit breaker pattern
- Create request batching logic
- Test coordination

**Day 17 Afternoon (4 hours)**
- Build metrics collection
- Add performance monitoring
- Document coordination protocol
- Update scanner API

### Day 18-19: Enhanced Filter Implementation
**Day 18 Morning (4 hours)**
```go
// src/go/scanner/filters/market_aware.go
package filters

import (
    "sync"
    "time"
)

type MarketAwareFilter struct {
    dataCache     *DataCache
    pythonClient  *PythonAPIClient
    cacheTTL      time.Duration
    mu            sync.RWMutex
}

func (f *MarketAwareFilter) FilterByGreeks(contracts []Contract, criteria GreeksCriteria) []Contract {
    // Get Greeks data with caching
    greeksData := f.getGreeksData(contracts)
    
    filtered := make([]Contract, 0)
    for i, contract := range contracts {
        greeks := greeksData[i]
        if f.matchesGreeksCriteria(greeks, criteria) {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}

func (f *MarketAwareFilter) getGreeksData(contracts []Contract) []Greeks {
    // Check cache first
    cached := f.checkCache(contracts)
    missing := f.findMissing(contracts, cached)
    
    if len(missing) > 0 {
        // Request only missing data
        fresh := f.requestGreeks(missing)
        f.updateCache(fresh)
        
        // Merge cached and fresh
        return f.mergeGreeksData(cached, fresh)
    }
    
    return cached
}
```

**Day 18 Afternoon (4 hours)**
- Implement all filter types with caching
- Add filter chain optimization
- Create filter performance tests
- Document filter patterns

**Day 19 (8 hours)**
- Complete advanced filters
- Test filter combinations
- Optimize for large datasets
- Create filter benchmarks

### Day 20-21: Real-time Streaming Integration
**Day 20 Morning (4 hours)**
```go
// src/go/scanner/streaming/websocket.go
package streaming

import (
    "github.com/gorilla/websocket"
    "encoding/json"
)

type StreamingScanner struct {
    scanner      *Scanner
    coordinator  *RequestCoordinator
    broadcaster  *Broadcaster
}

func (s *StreamingScanner) StreamResults(criteria ScanCriteria) {
    resultsChan := make(chan ScanResult, 100)
    
    // Start scanning in background
    go s.continuousScan(criteria, resultsChan)
    
    // Broadcast results to all connected clients
    for result := range resultsChan {
        s.broadcaster.Send(result)
    }
}

func (s *StreamingScanner) continuousScan(criteria ScanCriteria, results chan<- ScanResult) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Scan with coordination
            scanResults := s.scanner.ScanWithBackpressure(criteria)
            
            // Send results
            for _, result := range scanResults {
                select {
                case results <- result:
                default:
                    // Drop if channel full
                }
            }
        }
    }
}
```

**Day 20 Afternoon (4 hours)**
- Implement WebSocket server
- Add result deduplication
- Create streaming protocol
- Test real-time updates

**Day 21 (8 hours)**
- Complete streaming integration
- Add reconnection handling
- Performance optimization
- Update documentation

### Day 22-23: API Finalization
**Day 22 (8 hours)**
- Finalize REST API
- Add OpenAPI documentation
- Create client SDK
- Test all endpoints

**Day 23 (8 hours)**
- Integration testing
- Load testing
- Performance tuning
- Flow journal update

### Day 24-25: Scanner Polish
**Day 24 (8 hours)**
- Code review and refactoring
- Add comprehensive logging
- Create deployment configs
- Update tests

**Day 25 (8 hours)**
- Final integration tests
- Documentation review
- Performance benchmarks
- Prepare for GUI integration

---

## 🖥️ Phase 3: GUI Development with Real-time Updates (15 Days)

### Day 26-28: GUI Foundation with Monitoring
**Day 26 Morning (4 hours)**
```svelte
<!-- src/gui/frontend/src/components/SystemStatus.svelte -->
<script>
    import { onMount } from 'svelte';
    import { systemStatus } from '$lib/stores/monitoring';
    
    onMount(() => {
        // Poll health endpoint
        const interval = setInterval(async () => {
            const response = await fetch('/api/health');
            const health = await response.json();
            systemStatus.set(health);
        }, 1000);
        
        return () => clearInterval(interval);
    });
</script>

<div class="status-bar">
    <div class="status-item" class:healthy={$systemStatus.connected}>
        <span class="icon">🔌</span>
        TWS: {$systemStatus.connected ? 'Connected' : 'Disconnected'}
    </div>
    
    <div class="status-item">
        <span class="icon">📊</span>
        Subscriptions: {$systemStatus.subscriptions.active}/{$systemStatus.subscriptions.max}
    </div>
    
    <div class="status-item">
        <span class="icon">📬</span>
        Queue: {$systemStatus.queue_size}
    </div>
    
    {#if $systemStatus.throttling}
        <div class="status-item warning">
            <span class="icon">⚠️</span>
            Rate Limited
        </div>
    {/if}
</div>

<style>
    .status-bar {
        display: flex;
        gap: 1rem;
        padding: 0.5rem;
        background: var(--surface);
        border-bottom: 1px solid var(--border);
    }
    
    .status-item {
        display: flex;
        align-items: center;
        gap: 0.25rem;
    }
    
    .status-item.healthy {
        color: var(--success);
    }
    
    .status-item.warning {
        color: var(--warning);
        animation: pulse 1s infinite;
    }
</style>
```

**Day 26-28**: Continue with GUI implementation as in original plan, but with added:
- System status monitoring
- Subscription usage gauge
- Queue depth indicator
- Throttle warnings
- Connection health display

### Day 29-40: Complete GUI as per original plan with monitoring additions

---

## 🔧 Phase 4: System Integration with Monitoring (10 Days)

### Day 41-42: Service Orchestration with Metrics
**Day 41 Morning (4 hours)**
```yaml
# docker-compose.yml
version: '3.8'

services:
  python-ibkr:
    build: ./docker/python-ibkr
    environment:
      - TWS_HOST=host.docker.internal
      - TWS_PORT=7497
      - CLIENT_ID=1
      - WATCHDOG_TIMEOUT=60
      - MAX_SUBSCRIPTIONS=90
    ports:
      - "8080:8080"  # Metrics and health
    volumes:
      - ./src/python:/app
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      
  go-scanner:
    build: ./docker/go-scanner
    depends_on:
      python-ibkr:
        condition: service_healthy
    environment:
      - IBKR_API_URL=http://python-ibkr:8080
      - MAX_CONCURRENT_REQUESTS=10
    volumes:
      - ./src/go:/app
      
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
      
  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./monitoring/dashboards:/var/lib/grafana/dashboards
      - grafana_data:/var/lib/grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      
volumes:
  prometheus_data:
  grafana_data:
```

### Day 43-50: Complete integration as per original plan with monitoring

---

## 📊 New Success Metrics

### System Health Indicators
```gherkin
Feature: Comprehensive System Monitoring
  As a system operator
  I want real-time health metrics
  So that I can ensure reliable operation

  Scenario: Monitor all critical metrics
    Given the system is running
    When I check the dashboard
    Then I see:
      | Metric                  | Threshold           | Alert Level |
      | TWS Connection         | Connected           | Critical    |
      | Subscription Usage     | <90%                | Warning     |
      | Request Queue Depth    | <100                | Warning     |
      | Throttle Events/min    | <5                  | Info        |
      | Order Fill Rate        | >95%                | Warning     |
      | API Error Rate         | <1%                 | Warning     |
      | Scanner Performance    | <100ms              | Info        |
```

---

## 📝 Flow Journal Structure

### Daily Entry Template
```markdown
# Flow Journal - Day X - [Date]

## Morning Intention
- Energy level: [1-10]
- Focus area: [What excites me today]
- Vibe: [Current mood/feeling]

## Session Highlights
### Breakthroughs
- [Key insights or aha moments]

### Challenges
- [What blocked flow]
- [How I worked around it]

### Code Snippets
```[language]
// Interesting patterns discovered
```

## API Learnings
- [TWS API specific discoveries]
- [ib-insync patterns that worked well]
- [Event handling insights]

## Tomorrow's Flow
- [What I'm excited to tackle next]
- [Energy-matched tasks]

## Vibe Check
- Flow state achieved: [Yes/No/Partial]
- Best working music/environment: [Notes]
```

---

## 🚀 Key Improvements Summary

1. **Async-First Architecture**: Complete redesign around ib-insync's event model
2. **Built-in Monitoring**: Prometheus/Grafana from day one
3. **Subscription Management**: LRU cache with automatic eviction
4. **Request Coordination**: Backpressure between services
5. **Event-Driven Design**: Embrace callbacks over polling
6. **Health Visibility**: Real-time metrics in GUI
7. **Flow Journal Integration**: Daily progress tracking

This updated plan aligns perfectly with TWS API requirements while maintaining vibe coding principles.