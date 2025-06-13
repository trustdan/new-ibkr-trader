# 🚀 IBKR Spread Automation - Unified Master Plan
## Reconciling v1 & v2 into the Ultimate Vibe-Driven Development Journey

## 📋 Executive Summary

This unified master plan combines the comprehensive structure of v1 with the critical async-first architecture and monitoring insights from v2. We're building an automated vertical spread options trading system that honors both TWS API requirements and vibe coding principles, creating a flow-state-preserving development experience while delivering production-ready trading automation.

### Core Architecture Evolution
```
┌─────────────────────────────────────────────────────────────────┐
│                     Windows GUI Application                      │
│                   (Go Backend + Svelte Frontend)                 │
│                    Real-time System Monitoring                   │
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
│  - Built-in rate limiting       │   - Adaptive request flow     │
└─────────────────────────────────┴───────────────────────────────┘
                      │                   │
             TCP Socket (Async)      Metrics Export
                      │                   │
              ┌───────┴──────┐   ┌───────┴──────┐
              │   TWS/IB     │   │ Prometheus/  │
              │   Gateway    │   │   Grafana    │
              └──────────────┘   └──────────────┘
```

### Key Architectural Decisions (from v2 learnings)
1. **Async Everything**: Python service built entirely on asyncio/ib-insync patterns
2. **Event-Driven Core**: Replace all polling with event subscriptions
3. **Built-in Throttling**: Leverage ib-insync's rate limiting instead of custom implementation
4. **Subscription Manager**: Track and manage market data line usage with LRU eviction
5. **Request Coordinator**: Intelligent flow control between Go scanner and Python service
6. **Monitoring First**: Prometheus/Grafana integration from day one
7. **Living Documentation**: Flow journals and real-time metric dashboards

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
      | ADR/               | Architecture decisions            |
    And I have essential files:
      | File               | Content                           |
      | .gitignore         | Proper exclusions                 |
      | LICENSE            | MIT License                       |
      | CHANGELOG.md       | Version history template          |
      | IDEAS.md           | Future feature brainstorming      |
      | RULES.md           | Quick reference for vibe coding   |
      | MONITORING.md      | System metrics tracking           |
```

**Afternoon (4 hours)**
- Create `.vibe/templates/` with async code snippets
- Set up `flow_journal/template.md` for daily entries
- Initialize git repository with meaningful first commit: "🌟 Birth of IBKR Spread Automation - The journey begins"
- Create `experiments/async-patterns/` for testing ib-insync patterns
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
      | aiocache           | Async caching                    |
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
  I want crystal clear TWS configuration docs
  So that setup is foolproof and repeatable

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
- Write `docs/api/subscription-limits.md` with TWS tier details
- Document order ID management in `docs/trading/order-management.md`
- Create `ADR/001-async-architecture.md` - why we chose event-driven
- Add `docs/monitoring/metrics.md` with key performance indicators
- Set up quick-start guide in `docs/QUICKSTART.md`

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
.PHONY: dev test monitor clean vibe

dev:
	@echo "🚀 Starting async development environment..."
	docker-compose up -d
	@echo "⏳ Waiting for services..."
	@sleep 5
	@make health-check

test:
	@echo "🧪 Running async tests..."
	docker-compose run --rm python-ibkr pytest -v --asyncio-mode=auto

monitor:
	@echo "📊 Opening monitoring dashboards..."
	open http://localhost:9090  # Prometheus
	open http://localhost:3000  # Grafana

vibe:
	@echo "🌊 Checking the vibe..."
	@cat .vibe/manifesto.md
	@echo "\n📝 Latest flow journal entry:"
	@ls -t flow_journal/*.md | head -1 | xargs tail -20

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
- Write first flow journal entry documenting setup experience
- Test ib-insync examples in `experiments/connection-test/`
- Document async gotchas discovered
- Set up monitoring dashboard templates
- Update IDEAS.md with initial feature thoughts
- Prepare Phase 1 work breakdown

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
- Implement event handlers for all connection states
- Add comprehensive error handling (Error 502, 507, 1100)
- Create connection state persistence
- Build async health check endpoint

**Day 7 Morning (4 hours)**
```gherkin
Feature: Robust Connection Management
  As the IBKR service
  I want automatic connection recovery
  So that daily TWS restarts don't break the system

  Scenario: Handle daily TWS restart gracefully
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
- Add connection metrics collection
- Create restart schedule awareness
- Test watchdog recovery scenarios

### Day 8-9: Market Data Subscription Management
**Day 8 Morning (4 hours)**
```python
# src/python/core/subscription_manager.py
from collections import OrderedDict
from typing import Dict, Set
import asyncio
import logging

class SubscriptionManager:
    """Manages market data subscriptions within TWS limits"""
    
    def __init__(self, ib, max_lines=90):  # Leave headroom
        self.ib = ib
        self.max_lines = max_lines
        self.active_subscriptions: OrderedDict[str, Contract] = OrderedDict()
        self.subscription_counts: Dict[str, int] = {}
        self.eviction_count = 0
        
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
            self.eviction_count += 1
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
            'usage_pct': len(self.active_subscriptions) / self.max_lines * 100,
            'evictions': self.eviction_count
        }
```

**Day 8 Afternoon (4 hours)**
- Implement subscription pooling for similar contracts
- Add subscription request queuing
- Create usage monitoring endpoint
- Test with various account subscription levels

**Day 9 Morning (4 hours)**
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
    And eviction is logged for monitoring
```

**Day 9 Afternoon (4 hours)**
- Optimize subscription patterns
- Document best practices
- Create subscription dashboard in Grafana
- Flow journal on data management insights

### Day 10-11: Request Coordination & Rate Limiting
**Day 10 Morning (4 hours)**
```python
# src/python/core/request_coordinator.py
import asyncio
from asyncio import Queue, Semaphore
from dataclasses import dataclass
from typing import Any, Callable
import time

@dataclass
class Request:
    """Async request wrapper"""
    func: Callable
    args: tuple
    kwargs: dict
    future: asyncio.Future
    priority: int = 0
    submitted_at: float = 0

class RequestCoordinator:
    """Coordinates requests between services respecting ib-insync throttling"""
    
    def __init__(self, ib):
        self.ib = ib
        self.request_queue: Queue[Request] = Queue()
        self.processing = False
        self.throttled = False
        self.metrics = {
            'total_requests': 0,
            'throttle_events': 0,
            'current_queue_size': 0,
            'avg_processing_time': 0
        }
        
        # Monitor throttling
        ib.client.throttleStart += self._on_throttle_start
        ib.client.throttleEnd += self._on_throttle_end
        
    async def submit_request(self, func, *args, priority=0, **kwargs):
        """Submit request for coordinated execution"""
        future = asyncio.Future()
        request = Request(func, args, kwargs, future, priority, time.time())
        
        await self.request_queue.put(request)
        self.metrics['current_queue_size'] = self.request_queue.qsize()
        
        if not self.processing:
            asyncio.create_task(self._process_requests())
            
        return await future
```

**Day 10 Afternoon (4 hours)**
- Implement request batching for similar operations
- Add priority queue for time-sensitive requests
- Create backpressure mechanism for Go scanner
- Test under heavy load with realistic patterns

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
      | Tracks metrics       | For dashboard display    |
```

**Day 11 Afternoon (4 hours)**
- Add request deduplication
- Implement smart batching algorithms
- Create performance metrics dashboard
- Update flow journal with coordination insights

### Day 12-13: Order Execution Engine (Async)
**Day 12 Morning (4 hours)**
```python
# src/python/trading/async_order_engine.py
from ib_insync import Contract, Order, Trade, ComboLeg
import asyncio
from typing import List, Optional
from datetime import datetime

class AsyncOrderEngine:
    """Event-driven order execution engine"""
    
    def __init__(self, ib, coordinator):
        self.ib = ib
        self.coordinator = coordinator
        self.active_trades: Dict[int, Trade] = {}
        self.execution_metrics = {
            'total_orders': 0,
            'successful_fills': 0,
            'rejected_orders': 0,
            'avg_fill_time': 0
        }
        
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
        quantity: int = 1,
        order_type: str = 'LMT'
    ) -> Trade:
        """Execute vertical spread with full event handling"""
        
        start_time = datetime.now()
        
        # Create combo contract
        combo = await self._create_spread_combo(
            symbol, expiry, long_strike, short_strike, right
        )
        
        # Create order
        order = Order(
            action='BUY' if long_strike < short_strike else 'SELL',
            totalQuantity=quantity,
            orderType=order_type,
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
            
        # Set limit price based on preview and market
        order.lmtPrice = await self._calculate_optimal_limit_price(combo, preview)
        order.transmit = True
        
        # Place order
        trade = await self.coordinator.submit_request(
            self.ib.placeOrderAsync, combo, order
        )
        
        # Track active trade
        self.active_trades[trade.order.orderId] = trade
        self.execution_metrics['total_orders'] += 1
        
        # Wait for fill or timeout
        filled = await self._wait_for_fill(trade, timeout=60)
        
        if filled:
            fill_time = (datetime.now() - start_time).total_seconds()
            self._update_fill_metrics(fill_time)
            
        return trade
```

**Day 12 Afternoon (4 hours)**
- Implement combo contract creation helpers
- Add OCA group management for risk
- Create order validation logic
- Build fill monitoring system

**Day 13 Morning (4 hours)**
```gherkin
Feature: Reliable Spread Execution
  As a trader
  I want confident spread execution
  So that orders fill at optimal prices

  Scenario: Execute debit spread with smart pricing
    Given I want to buy a call spread
    When I submit the order
    Then the system:
      | Step          | Action                      | Validation           |
      | Build combo   | Create spread contract     | Both legs qualified  |
      | Preview       | whatIfOrder check          | Margin acceptable    |
      | Smart price   | Calculate optimal limit    | Within bid-ask       |
      | Submit        | Place combo order          | Order accepted       |
      | Monitor       | Track via events           | Real-time updates    |
      | Fill          | Both legs execute          | Prices reasonable    |
      | Report        | Log execution metrics      | Update dashboard     |
```

**Day 13 Afternoon (4 hours)**
- Test various order types and scenarios
- Implement execution reporting
- Create trade history tracking
- Document order execution patterns

### Day 14-15: Integration Testing & Monitoring Setup
**Day 14 Morning (4 hours)**
```python
# src/python/monitoring/metrics_server.py
from aiohttp import web
from prometheus_client import Counter, Gauge, Histogram, generate_latest

# Define comprehensive metrics
connection_status = Gauge('ibkr_connection_status', 'TWS connection status')
active_subscriptions = Gauge('ibkr_active_subscriptions', 'Active market data subscriptions')
subscription_usage_pct = Gauge('ibkr_subscription_usage_pct', 'Subscription usage percentage')
eviction_count = Counter('ibkr_subscription_evictions_total', 'Total subscription evictions')
request_queue_size = Gauge('ibkr_request_queue_size', 'Pending requests in queue')
order_execution_time = Histogram('ibkr_order_execution_seconds', 'Order execution time')
throttle_events = Counter('ibkr_throttle_events_total', 'Total throttle events')
api_errors = Counter('ibkr_api_errors_total', 'API errors by code', ['error_code'])

async def metrics_handler(request):
    """Prometheus metrics endpoint"""
    metrics = generate_latest()
    return web.Response(text=metrics.decode('utf-8'), content_type='text/plain')

async def health_handler(request):
    """Health check endpoint with detailed status"""
    app = request.app
    ib_service = app['ib_service']
    
    health = {
        'status': 'healthy' if ib_service.ib.isConnected() else 'unhealthy',
        'connected': ib_service.ib.isConnected(),
        'subscriptions': ib_service.subscription_manager.get_usage(),
        'queue_size': ib_service.coordinator.metrics['current_queue_size'],
        'throttled': ib_service.coordinator.throttled,
        'uptime': ib_service.get_uptime(),
        'next_order_id': ib_service.next_order_id
    }
    
    status_code = 200 if health['status'] == 'healthy' else 503
    return web.json_response(health, status=status_code)
```

**Day 14 Afternoon (4 hours)**
- Create comprehensive Grafana dashboards
- Set up alerting rules for critical metrics
- Test monitoring under various load scenarios
- Document all metrics and their meanings

**Day 15 Morning (4 hours)**
- Integration test suite for connection layer
- Test daily restart handling
- Verify subscription management
- Test order execution flows

**Day 15 Afternoon (4 hours)**
- Performance optimization based on metrics
- Update documentation with learnings
- Flow journal reflection on Phase 1
- Prepare for scanner integration

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
    adaptiveDelay time.Duration
    mu            sync.RWMutex
}

func NewRequestCoordinator(client *PythonAPIClient, maxConcurrent int) *RequestCoordinator {
    return &RequestCoordinator{
        pythonClient:  client,
        maxConcurrent: maxConcurrent,
        semaphore:     make(chan struct{}, maxConcurrent),
        metrics:       NewMetrics(),
        adaptiveDelay: 10 * time.Millisecond,
    }
}

func (rc *RequestCoordinator) RequestMarketData(ctx context.Context, contracts []Contract) error {
    // Implement intelligent backpressure
    select {
    case rc.semaphore <- struct{}{}:
        defer func() { <-rc.semaphore }()
        
        // Check Python service health first
        health, err := rc.pythonClient.GetHealth(ctx)
        if err != nil {
            return err
        }
        
        // Apply adaptive backpressure
        if health.QueueSize > 50 {
            delay := rc.calculateBackpressure(health)
            time.Sleep(delay)
        }
        
        // Batch request intelligently
        batches := rc.createOptimalBatches(contracts)
        for _, batch := range batches {
            if err := rc.pythonClient.RequestMarketDataBatch(ctx, batch); err != nil {
                return err
            }
        }
        
        return nil
        
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

**Day 16 Afternoon (4 hours)**
```gherkin
Feature: Coordinated Scanner Operations
  As the Go scanner
  I want to coordinate intelligently with Python service
  So that we maximize throughput without overwhelming TWS

  Scenario: Adaptive rate control
    Given Python service reports queue depth
    When scanner adjusts sending rate
    Then:
      | Queue Depth | Action                    | Delay        |
      | 0-25        | Send normally            | 10ms         |
      | 26-50       | Slight backpressure      | 25ms         |
      | 51-75       | Moderate backpressure    | 50ms         |
      | 76-100      | Heavy backpressure       | 100ms        |
      | >100        | Pause sending            | Until <50    |
```

**Day 17 Morning (4 hours)**
- Implement adaptive rate control algorithm
- Add circuit breaker for service protection
- Create intelligent request batching
- Test coordination under stress

**Day 17 Afternoon (4 hours)**
- Build comprehensive metrics collection
- Add performance monitoring
- Document coordination protocol
- Create scanner health endpoints

### Day 18-19: Enhanced Filter Implementation
**Day 18 Morning (4 hours)**
```go
// src/go/scanner/filters/filter_chain.go
package filters

import (
    "sync"
    "time"
)

type FilterChain struct {
    filters       []Filter
    cache         *FilterCache
    metrics       *FilterMetrics
}

type Filter interface {
    Apply(contracts []Contract) []Contract
    Name() string
    Priority() int  // For optimal ordering
}

// Comprehensive filter suite
type DeltaFilter struct {
    MinDelta float64
    MaxDelta float64
}

type DTEFilter struct {
    MinDays int
    MaxDays int
}

type LiquidityFilter struct {
    MinVolume      int
    MinOpenInterest int
}

type SpreadFilter struct {
    MaxBidAskSpread float64
    MinSpreadWidth  float64
}

type GreeksFilter struct {
    MaxGamma float64
    MinTheta float64
    MaxVega  float64
}

type IVFilter struct {
    MinIVPercentile int
    MaxIVPercentile int
    LookbackDays    int
}

type ProbabilityFilter struct {
    MinPoP float64  // Probability of Profit
    MaxITM float64  // In The Money probability
}
```

**Day 18 Afternoon (4 hours)**
- Implement all filter types with caching
- Add filter chain optimization (order by selectivity)
- Create filter performance benchmarks
- Document filter best practices

**Day 19 Morning (4 hours)**
```gherkin
Feature: High-Performance Filter Chain
  As a trader
  I want lightning-fast filtering
  So that opportunities are found quickly

  Scenario: Apply complex filter chain
    Given 10,000 option contracts
    When applying these filters:
      | Filter            | Configuration        | Expected Selectivity |
      | Liquidity        | Vol>100, OI>50       | Filters 80%         |
      | DTE              | 30-60 days           | Filters 70%         |
      | Delta            | 0.25-0.35            | Filters 60%         |
      | Spread Width     | <$0.10               | Filters 50%         |
      | IV Percentile    | >50                  | Filters 40%         |
    Then results return in <100ms
    And filters apply in optimal order
    And cache hit rate >80%
```

**Day 19 Afternoon (4 hours)**
- Complete advanced filter implementations
- Test filter combinations
- Optimize for large datasets
- Create filter chain visualizer

### Day 20-21: Real-time Streaming Integration
**Day 20 Morning (4 hours)**
```go
// src/go/scanner/streaming/websocket_server.go
package streaming

import (
    "github.com/gorilla/websocket"
    "encoding/json"
    "sync"
)

type StreamingScanner struct {
    scanner      *Scanner
    coordinator  *RequestCoordinator
    broadcaster  *Broadcaster
    subscribers  map[string]*Subscriber
    mu           sync.RWMutex
}

type ScanUpdate struct {
    Type      string      `json:"type"`  // "result", "status", "error"
    Timestamp int64       `json:"timestamp"`
    Data      interface{} `json:"data"`
}

func (s *StreamingScanner) HandleWebSocket(ws *websocket.Conn) {
    subscriber := NewSubscriber(ws)
    s.addSubscriber(subscriber)
    defer s.removeSubscriber(subscriber)
    
    // Send initial status
    s.sendStatus(subscriber)
    
    // Handle incoming messages
    for {
        var msg ScanRequest
        if err := ws.ReadJSON(&msg); err != nil {
            break
        }
        
        // Process scan request
        go s.processScanRequest(subscriber, msg)
    }
}

func (s *StreamingScanner) continuousScan(criteria ScanCriteria) {
    resultsChan := make(chan []ScanResult, 10)
    
    // Scan with intelligent pacing
    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            if s.shouldScan() {
                results := s.scanner.ScanWithCoordination(criteria)
                resultsChan <- results
            }
        }
    }()
    
    // Broadcast results
    for results := range resultsChan {
        update := ScanUpdate{
            Type:      "results",
            Timestamp: time.Now().Unix(),
            Data:      results,
        }
        s.broadcast(update)
    }
}
```

**Day 20 Afternoon (4 hours)**
- Implement WebSocket protocol
- Add result deduplication
- Create subscription management
- Test real-time streaming

**Day 21 Morning (4 hours)**
```gherkin
Feature: Real-time Scanner Streaming
  As a GUI client
  I want live scanner updates
  So that I see opportunities immediately

  Scenario: Stream scanner results
    Given scanner is running continuously
    When results are found
    Then:
      | Event              | Client Receives        | Latency    |
      | New result         | Full result data      | <50ms      |
      | Result update      | Delta only            | <50ms      |
      | Status change      | Status message        | <10ms      |
      | Error             | Error details         | Immediate  |
      | Reconnection      | Restore subscriptions | Automatic  |
```

**Day 21 Afternoon (4 hours)**
- Complete streaming implementation
- Add reconnection handling
- Performance optimization
- Update API documentation

### Day 22-23: API Finalization & Testing
**Day 22 Morning (4 hours)**
- Finalize REST API endpoints
- Create OpenAPI/Swagger documentation
- Build Go client SDK
- Add API versioning

**Day 22 Afternoon (4 hours)**
```go
// src/go/scanner/api/routes.go
package api

// REST API endpoints
// GET    /health                 - Scanner health status
// GET    /metrics                - Prometheus metrics
// POST   /scan/start             - Start scanning
// DELETE /scan/stop              - Stop scanning
// GET    /scan/status            - Current scan status
// POST   /scan/filters           - Update filters
// GET    /scan/filters           - Get current filters
// GET    /scan/results           - Get latest results (polling)
// WS     /scan/stream            - WebSocket streaming
// GET    /filters/presets        - Get saved presets
// POST   /filters/presets        - Save new preset
```

**Day 23 Morning (4 hours)**
- Comprehensive integration testing
- Load testing with 10k+ contracts
- Stress testing coordination
- Document performance characteristics

**Day 23 Afternoon (4 hours)**
- Code review and refactoring
- Update flow journal
- Polish documentation
- Prepare for GUI integration

### Day 24-25: Scanner Polish & Optimization
**Day 24 (8 hours)**
- Profile and optimize hot paths
- Implement advanced caching strategies
- Add comprehensive logging
- Create deployment configurations

**Day 25 (8 hours)**
- Final integration tests with Python service
- Performance benchmarking suite
- Documentation review
- Create scanner operation guide

---

## 🖥️ Phase 3: GUI Development with Real-time Monitoring (15 Days)

### Day 26-28: GUI Foundation with System Monitoring
**Day 26 Morning (4 hours)**
```gherkin
Feature: Desktop Application with Live Monitoring
  As a trader
  I want a responsive GUI with system status
  So that I always know the system state

  Scenario: Application architecture
    Given I need a desktop GUI with monitoring
    When I design the architecture
    Then I have:
      | Component         | Technology           | Purpose              |
      | Backend          | Go + Gin            | API & coordination   |
      | Frontend         | Svelte + Vite       | Reactive UI          |
      | Desktop Wrapper  | Wails               | Native experience    |
      | State Management | Svelte stores       | Real-time updates    |
      | UI Framework     | Tailwind + DaisyUI  | Beautiful components |
      | Charts           | Chart.js            | Data visualization   |
```

**Day 26 Afternoon (4 hours)**
```javascript
// src/gui/frontend/src/lib/stores/system.js
import { writable, derived } from 'svelte/store';
import { createWebSocketStore } from './websocket';

// System health store
export const systemHealth = writable({
    tws: { connected: false, uptime: 0 },
    subscriptions: { active: 0, max: 100, usage_pct: 0 },
    queue: { size: 0, processing: false },
    throttling: false,
    errors: []
});

// Scanner state store
export const scannerState = writable({
    running: false,
    filters: {},
    results: [],
    lastUpdate: null
});

// WebSocket connection for real-time updates
export const ws = createWebSocketStore('/api/ws', {
    onMessage: (data) => {
        if (data.type === 'health') {
            systemHealth.set(data.payload);
        } else if (data.type === 'scanner') {
            scannerState.update(s => ({ ...s, ...data.payload }));
        }
    }
});

// Derived stores for UI
export const isHealthy = derived(
    systemHealth,
    $health => $health.tws.connected && !$health.throttling
);

export const subscriptionWarning = derived(
    systemHealth,
    $health => $health.subscriptions.usage_pct > 80
);
```

**Day 27 Morning (4 hours)**
- Set up Svelte project with Vite
- Configure Wails for desktop packaging
- Implement WebSocket client
- Create reactive store architecture

**Day 27 Afternoon (4 hours)**
```svelte
<!-- src/gui/frontend/src/components/SystemStatusBar.svelte -->
<script>
    import { systemHealth, isHealthy, subscriptionWarning } from '$lib/stores/system';
    import StatusIndicator from './StatusIndicator.svelte';
    import ProgressBar from './ProgressBar.svelte';
</script>

<div class="status-bar">
    <StatusIndicator 
        label="TWS" 
        status={$systemHealth.tws.connected ? 'connected' : 'disconnected'}
        pulse={!$isHealthy}
    />
    
    <div class="subscription-meter">
        <span class="label">Market Data</span>
        <ProgressBar 
            value={$systemHealth.subscriptions.usage_pct} 
            max={100}
            color={$subscriptionWarning ? 'warning' : 'primary'}
        />
        <span class="value">
            {$systemHealth.subscriptions.active}/{$systemHealth.subscriptions.max}
        </span>
    </div>
    
    <div class="queue-status">
        <span class="label">Queue</span>
        <span class="value" class:processing={$systemHealth.queue.processing}>
            {$systemHealth.queue.size}
        </span>
    </div>
    
    {#if $systemHealth.throttling}
        <div class="alert alert-warning">
            ⚠️ Rate Limited
        </div>
    {/if}
</div>

<style>
    .status-bar {
        @apply flex items-center gap-4 p-2 bg-base-200 border-b;
    }
    
    .subscription-meter {
        @apply flex items-center gap-2;
    }
    
    .queue-status .processing {
        @apply text-warning animate-pulse;
    }
</style>
```

**Day 28 Morning (4 hours)**
- Create main application layout
- Build navigation system
- Implement theme system (light/dark)
- Add responsive design

**Day 28 Afternoon (4 hours)**
- Set up build pipeline
- Configure hot reloading
- Create component library
- Document UI patterns

### Day 29-31: Parameter Control Interface
**Day 29 Morning (4 hours)**
```svelte
<!-- src/gui/frontend/src/components/filters/FilterPanel.svelte -->
<script>
    import { scannerConfig } from '$lib/stores/trading';
    import DeltaFilter from './DeltaFilter.svelte';
    import DTEFilter from './DTEFilter.svelte';
    import LiquidityFilter from './LiquidityFilter.svelte';
    import SpreadFilter from './SpreadFilter.svelte';
    import GreeksFilter from './GreeksFilter.svelte';
    import IVFilter from './IVFilter.svelte';
    import PresetManager from './PresetManager.svelte';
    
    let activeFilters = {
        delta: true,
        dte: true,
        liquidity: true,
        spread: false,
        greeks: false,
        iv: false
    };
</script>

<div class="filter-panel">
    <div class="panel-header">
        <h2>Scanner Filters</h2>
        <PresetManager bind:config={$scannerConfig} />
    </div>
    
    <div class="filter-grid">
        {#if activeFilters.delta}
            <DeltaFilter bind:config={$scannerConfig.filters.delta} />
        {/if}
        
        {#if activeFilters.dte}
            <DTEFilter bind:config={$scannerConfig.filters.dte} />
        {/if}
        
        {#if activeFilters.liquidity}
            <LiquidityFilter bind:config={$scannerConfig.filters.liquidity} />
        {/if}
        
        <!-- Additional filters with toggle switches -->
    </div>
    
    <div class="filter-actions">
        <button class="btn btn-primary" on:click={applyFilters}>
            Apply Filters
        </button>
        <button class="btn btn-ghost" on:click={resetFilters}>
            Reset
        </button>
    </div>
</div>

<style>
    .filter-panel {
        @apply p-4 bg-base-100 rounded-lg shadow-lg;
    }
    
    .filter-grid {
        @apply grid grid-cols-1 md:grid-cols-2 gap-4 my-4;
    }
</style>
```

**Day 29 Afternoon (4 hours)**
- Implement all filter components
- Create dual-handle range sliders
- Add input validation
- Build preset save/load system

**Day 30 Morning (4 hours)**
```gherkin
Feature: Intuitive Filter Controls
  As a trader
  I want easy-to-use filter controls
  So that I can quickly adjust my strategy

  Scenario: Configure complex filters
    Given I need to set multiple parameters
    When I use the filter panel
    Then I can:
      | Action              | Result                    |
      | Toggle filters      | Enable/disable instantly  |
      | Adjust ranges       | Smooth slider interaction |
      | Enter exact values  | Keyboard input works      |
      | Save configuration  | Named preset saved        |
      | Load preset         | All values restored       |
      | See preview count   | Estimated results shown   |
```

**Day 30 Afternoon (4 hours)**
- Create filter preview system
- Add tooltips and help text
- Implement keyboard shortcuts
- Build quick-access toolbar

**Day 31 Morning (4 hours)**
- Polish filter interactions
- Add filter validation
- Create filter templates
- Test usability

**Day 31 Afternoon (4 hours)**
- Implement filter import/export
- Add filter history
- Create filter suggestions
- Document filter patterns

### Day 32-34: Real-time Data Visualization
**Day 32 Morning (4 hours)**
```svelte
<!-- src/gui/frontend/src/components/ScannerResults.svelte -->
<script>
    import { scanResults, selectedSpread } from '$lib/stores/trading';
    import VirtualList from '@tanstack/svelte-virtual';
    import ResultRow from './ResultRow.svelte';
    import SpreadChart from './SpreadChart.svelte';
    
    let sortColumn = 'score';
    let sortDirection = 'desc';
    let selectedRows = new Set();
    
    $: sortedResults = sortResults($scanResults, sortColumn, sortDirection);
    
    function handleRowClick(result) {
        selectedSpread.set(result);
    }
</script>

<div class="results-container">
    <div class="results-header">
        <h3>Scanner Results ({sortedResults.length})</h3>
        <div class="result-actions">
            <button class="btn btn-sm" on:click={exportResults}>
                Export
            </button>
        </div>
    </div>
    
    <div class="results-grid">
        <VirtualList 
            height={600}
            itemCount={sortedResults.length}
            itemSize={50}
            overscan={5}
        >
            <div slot="item" let:index let:style {style}>
                <ResultRow 
                    result={sortedResults[index]}
                    selected={selectedRows.has(index)}
                    on:click={() => handleRowClick(sortedResults[index])}
                />
            </div>
        </VirtualList>
    </div>
    
    {#if $selectedSpread}
        <div class="spread-preview">
            <SpreadChart spread={$selectedSpread} />
        </div>
    {/if}
</div>

<style>
    .results-container {
        @apply flex flex-col h-full;
    }
    
    .results-grid {
        @apply flex-1 overflow-hidden border rounded-lg;
    }
    
    .spread-preview {
        @apply mt-4 p-4 bg-base-200 rounded-lg;
    }
</style>
```

**Day 32 Afternoon (4 hours)**
- Implement virtual scrolling for performance
- Add real-time result updates
- Create sorting system
- Build column customization

**Day 33 Morning (4 hours)**
```gherkin
Feature: Live Scanner Visualization
  As a trader
  I want to see results as they arrive
  So that I can act on opportunities quickly

  Scenario: Real-time result display
    Given scanner is streaming results
    When new results arrive
    Then:
      | Feature              | Behavior                   |
      | Update animation     | Smooth fade-in            |
      | Sort maintenance     | Position updates          |
      | Selection persist    | Selected items tracked    |
      | Performance          | 60fps with 1000+ rows    |
      | Greeks display       | Live updates              |
      | Profit calculation   | Real-time P&L             |
```

**Day 33 Afternoon (4 hours)**
- Add result highlighting and badges
- Implement Greeks visualization
- Create mini profit/loss charts
- Build detailed result panels

**Day 34 Morning (4 hours)**
- Add position tracking view
- Create P&L dashboard
- Implement trade history
- Build performance metrics

**Day 34 Afternoon (4 hours)**
- Test data visualization performance
- Optimize rendering
- Add export capabilities
- Document visualization patterns

### Day 35-37: Trade Execution Interface
**Day 35 Morning (4 hours)**
```svelte
<!-- src/gui/frontend/src/components/TradeExecution.svelte -->
<script>
    import { selectedSpread, accountInfo } from '$lib/stores/trading';
    import SpreadVisualizer from './SpreadVisualizer.svelte';
    import OrderPreview from './OrderPreview.svelte';
    import RiskCalculator from './RiskCalculator.svelte';
    
    let orderParams = {
        quantity: 1,
        orderType: 'LMT',
        limitPrice: 0,
        timeInForce: 'DAY'
    };
    
    let preview = null;
    let executing = false;
    
    async function previewOrder() {
        const response = await fetch('/api/orders/preview', {
            method: 'POST',
            body: JSON.stringify({
                spread: $selectedSpread,
                params: orderParams
            })
        });
        preview = await response.json();
    }
    
    async function executeOrder() {
        executing = true;
        try {
            const response = await fetch('/api/orders/execute', {
                method: 'POST',
                body: JSON.stringify({
                    spread: $selectedSpread,
                    params: orderParams
                })
            });
            const result = await response.json();
            // Handle execution result
        } finally {
            executing = false;
        }
    }
</script>

<div class="execution-panel">
    <SpreadVisualizer spread={$selectedSpread} />
    
    <div class="order-params">
        <h3>Order Parameters</h3>
        <div class="param-grid">
            <label>
                Quantity
                <input type="number" bind:value={orderParams.quantity} min="1" />
            </label>
            
            <label>
                Order Type
                <select bind:value={orderParams.orderType}>
                    <option value="LMT">Limit</option>
                    <option value="MKT">Market</option>
                </select>
            </label>
            
            {#if orderParams.orderType === 'LMT'}
                <label>
                    Limit Price
                    <input type="number" bind:value={orderParams.limitPrice} step="0.01" />
                </label>
            {/if}
        </div>
    </div>
    
    <RiskCalculator spread={$selectedSpread} quantity={orderParams.quantity} />
    
    {#if preview}
        <OrderPreview {preview} />
    {/if}
    
    <div class="execution-actions">
        <button 
            class="btn btn-secondary" 
            on:click={previewOrder}
            disabled={executing}
        >
            Preview Order
        </button>
        
        <button 
            class="btn btn-primary" 
            on:click={executeOrder}
            disabled={!preview || executing}
            class:loading={executing}
        >
            Execute Spread
        </button>
    </div>
</div>

<style>
    .execution-panel {
        @apply p-6 bg-base-100 rounded-lg shadow-xl;
    }
    
    .param-grid {
        @apply grid grid-cols-2 gap-4;
    }
    
    .execution-actions {
        @apply flex gap-4 mt-6;
    }
</style>
```

**Day 35 Afternoon (4 hours)**
- Create spread visualizer with payoff diagram
- Build comprehensive order preview
- Add risk calculations and warnings
- Implement confirmation flow

**Day 36 Morning (4 hours)**
```gherkin
Feature: Safe Trade Execution
  As a trader
  I want clear execution confirmation
  So that I never place unintended trades

  Scenario: Execute vertical spread safely
    Given I select a spread opportunity
    When I configure the order
    Then I see:
      | Information         | Display                   |
      | Spread visualization| Interactive payoff chart  |
      | Max risk           | Dollar amount + %         |
      | Max profit         | Dollar amount + %         |
      | Breakeven          | Price level + chart      |
      | Margin impact      | Current → After          |
      | Commission         | Estimated fees           |
      | Account impact     | Buying power change      |
    And I must confirm before submission
    And execution shows real-time status
```

**Day 36 Afternoon (4 hours)**
- Implement whatIfOrder integration
- Add margin impact calculator
- Create position sizing tool
- Build order modification interface

**Day 37 Morning (4 hours)**
- Add order status tracking
- Implement fill notifications
- Create execution log
- Build comprehensive error handling

**Day 37 Afternoon (4 hours)**
- Test execution flows end-to-end
- Add keyboard shortcuts
- Polish confirmation dialogs
- Document trade execution flow

### Day 38-40: Integration & Polish
**Day 38 Morning (4 hours)**
- Connect all UI components
- Implement state persistence
- Add comprehensive keyboard navigation
- Create context-sensitive help

**Day 38 Afternoon (4 hours)**
```svelte
<!-- src/gui/frontend/src/App.svelte -->
<script>
    import { onMount } from 'svelte';
    import SystemStatusBar from './components/SystemStatusBar.svelte';
    import FilterPanel from './components/filters/FilterPanel.svelte';
    import ScannerResults from './components/ScannerResults.svelte';
    import TradeExecution from './components/TradeExecution.svelte';
    import NotificationToast from './components/NotificationToast.svelte';
    import { initializeStores } from './lib/stores';
    
    let activeTab = 'scanner';
    
    onMount(() => {
        initializeStores();
        
        // Set up global keyboard shortcuts
        window.addEventListener('keydown', handleGlobalShortcuts);
        
        return () => {
            window.removeEventListener('keydown', handleGlobalShortcuts);
        };
    });
    
    function handleGlobalShortcuts(e) {
        if (e.ctrlKey || e.metaKey) {
            switch(e.key) {
                case 's': // Start/stop scanner
                    e.preventDefault();
                    toggleScanner();
                    break;
                case 'f': // Focus filter panel
                    e.preventDefault();
                    focusFilters();
                    break;
                case 'e': // Execute selected spread
                    e.preventDefault();
                    if (canExecute()) executeSelected();
                    break;
            }
        }
    }
</script>

<div class="app">
    <SystemStatusBar />
    
    <div class="app-content">
        <aside class="sidebar">
            <FilterPanel />
        </aside>
        
        <main class="main-content">
            <div class="tabs">
                <button 
                    class="tab" 
                    class:active={activeTab === 'scanner'}
                    on:click={() => activeTab = 'scanner'}
                >
                    Scanner
                </button>
                <button 
                    class="tab" 
                    class:active={activeTab === 'execution'}
                    on:click={() => activeTab = 'execution'}
                >
                    Execution
                </button>
            </div>
            
            <div class="tab-content">
                {#if activeTab === 'scanner'}
                    <ScannerResults />
                {:else if activeTab === 'execution'}
                    <TradeExecution />
                {/if}
            </div>
        </main>
    </div>
    
    <NotificationToast />
</div>

<style>
    .app {
        @apply h-screen flex flex-col bg-base-300;
    }
    
    .app-content {
        @apply flex-1 flex overflow-hidden;
    }
    
    .sidebar {
        @apply w-96 p-4 overflow-y-auto;
    }
    
    .main-content {
        @apply flex-1 flex flex-col p-4;
    }
</style>
```

**Day 39 Morning (4 hours)**
- Performance optimization
- Accessibility improvements (ARIA labels, keyboard nav)
- Theme customization system
- Settings management

**Day 39 Afternoon (4 hours)**
- Create onboarding flow
- Add tooltips and guided tours
- Implement user preferences
- Build help documentation

**Day 40 Morning (4 hours)**
- User acceptance testing
- Bug fixes and polish
- Create user guide
- Prepare deployment package

**Day 40 Afternoon (4 hours)**
- Final UI/UX review
- Performance benchmarking
- Create video tutorials
- Package application for distribution

---

## 🔧 Phase 4: System Integration & Production (10 Days)

### Day 41-42: Service Orchestration with Full Monitoring
**Day 41 Morning (4 hours)**
```yaml
# docker-compose.yml
version: '3.8'

services:
  python-ibkr:
    build: ./docker/python-ibkr
    container_name: ibkr-python
    environment:
      - TWS_HOST=host.docker.internal
      - TWS_PORT=7497
      - CLIENT_ID=1
      - WATCHDOG_TIMEOUT=60
      - MAX_SUBSCRIPTIONS=90
      - LOG_LEVEL=INFO
    ports:
      - "8080:8080"  # API and metrics
    volumes:
      - ./src/python:/app
      - ./logs/python:/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped
      
  go-scanner:
    build: ./docker/go-scanner
    container_name: ibkr-scanner
    depends_on:
      python-ibkr:
        condition: service_healthy
    environment:
      - IBKR_API_URL=http://python-ibkr:8080
      - MAX_CONCURRENT_REQUESTS=10
      - CACHE_TTL=60
      - LOG_LEVEL=INFO
    ports:
      - "8081:8081"  # Scanner API
    volumes:
      - ./src/go:/app
      - ./logs/scanner:/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
    restart: unless-stopped
      
  gui-backend:
    build: ./docker/gui
    container_name: ibkr-gui
    depends_on:
      - go-scanner
    ports:
      - "3000:3000"  # GUI backend
    volumes:
      - ./src/gui:/app
      - ./logs/gui:/logs
    environment:
      - SCANNER_URL=http://go-scanner:8081
      - PYTHON_URL=http://python-ibkr:8080
    restart: unless-stopped
      
  prometheus:
    image: prom/prometheus:latest
    container_name: ibkr-prometheus
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
    restart: unless-stopped
      
  grafana:
    image: grafana/grafana:latest
    container_name: ibkr-grafana
    volumes:
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards
      - grafana_data:/var/lib/grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=vibetrading
      - GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource
    depends_on:
      - prometheus
    restart: unless-stopped
      
  # Optional: Log aggregation
  loki:
    image: grafana/loki:latest
    container_name: ibkr-loki
    ports:
      - "3100:3100"
    volumes:
      - ./monitoring/loki:/etc/loki
      - loki_data:/loki
    restart: unless-stopped
      
volumes:
  prometheus_data:
  grafana_data:
  loki_data:
```

**Day 41 Afternoon (4 hours)**
- Configure Prometheus scraping
- Set up Grafana dashboards
- Create alert rules
- Test monitoring stack

**Day 42 Morning (4 hours)**
```gherkin
Feature: Production-Ready Orchestration
  As a system operator
  I want reliable service management
  So that the system runs smoothly

  Scenario: Service health monitoring
    Given all services are running
    When I check the monitoring dashboard
    Then I see:
      | Service       | Metrics Available              |
      | Python IBKR   | Connection, subscriptions, RPS |
      | Go Scanner    | Scan rate, filter performance  |
      | GUI Backend   | Response times, active users   |
      | Overall       | System health score            |
```

**Day 42 Afternoon (4 hours)**
- Implement service discovery
- Create backup/restore procedures
- Document deployment process
- Set up log rotation

### Day 43-44: End-to-End Workflows
**Day 43 Morning (4 hours)**
```gherkin
Feature: Complete Trading Workflow
  As a trader
  I want seamless end-to-end functionality
  So that trading is efficient and reliable

  Scenario: Full trade execution flow
    Given I start the application
    When I execute a complete workflow
    Then:
      | Step                 | System Action              | Validation          |
      | Launch GUI          | All services healthy       | Status green        |
      | Configure filters   | Filters sent to scanner    | Confirmation shown  |
      | Start scanning      | Results stream in          | Real-time updates   |
      | Select spread       | Details displayed          | Greeks calculated   |
      | Preview order       | whatIfOrder executed       | Margin checked      |
      | Execute trade       | Order placed via TWS       | Fill confirmed      |
      | Monitor position    | P&L updates live           | Accurate tracking   |
```

**Day 43 Afternoon (4 hours)**
- Test complete workflows
- Verify data consistency
- Check error propagation
- Validate state management

**Day 44 Morning (4 hours)**
- Integration test suite
- Performance testing (throughput, latency)
- Load testing (concurrent users)
- Stress testing (edge cases)

**Day 44 Afternoon (4 hours)**
- Document test results
- Create runbooks
- Update flow journal
- Plan optimization work

### Day 45-46: Production Hardening
**Day 45 Morning (4 hours)**
```bash
# scripts/deploy.sh
#!/bin/bash
set -e

echo "🚀 Deploying IBKR Spread Automation..."

# Pre-flight checks
./scripts/pre-deploy-checks.sh

# Build production images
docker-compose -f docker-compose.prod.yml build

# Run database migrations (if any)
# docker-compose run --rm python-ibkr python manage.py migrate

# Deploy with zero downtime
docker-compose -f docker-compose.prod.yml up -d --scale python-ibkr=2

# Wait for health checks
./scripts/wait-for-healthy.sh

# Switch traffic to new instances
docker-compose -f docker-compose.prod.yml up -d --remove-orphans

echo "✅ Deployment complete!"
```

**Day 45 Afternoon (4 hours)**
- Create deployment automation
- Set up CI/CD pipeline
- Configure monitoring alerts
- Build backup systems

**Day 46 Morning (4 hours)**
```gherkin
Feature: Production Security
  As a system administrator
  I want robust security measures
  So that trading is safe and compliant

  Scenario: Security hardening
    Given production deployment
    When security audit runs
    Then:
      | Check               | Status    | Action              |
      | API authentication  | Required  | JWT tokens          |
      | Data encryption     | Enabled   | TLS everywhere      |
      | Secrets management  | Vault     | No hardcoded values |
      | Audit logging       | Active    | All trades logged   |
      | Access control      | RBAC      | Role-based perms    |
```

**Day 46 Afternoon (4 hours)**
- Security audit
- Performance optimization
- Documentation review
- Deployment dry run

### Day 47-48: Testing & Documentation
**Day 47 Morning (4 hours)**
- End-to-end system testing
- User acceptance testing
- Performance benchmarking
- Security penetration testing

**Day 47 Afternoon (4 hours)**
- Fix identified issues
- Optimize bottlenecks
- Update documentation
- Create troubleshooting guide

**Day 48 Morning (4 hours)**
- Complete system documentation
- Create operation runbooks
- Write troubleshooting guides
- Document known issues

**Day 48 Afternoon (4 hours)**
- Create video tutorials
- Write quick-start guide
- Prepare training materials
- Document FAQ

### Day 49-50: Launch Preparation
**Day 49 Morning (4 hours)**
- Final system review
- Performance validation
- Security sign-off
- Documentation approval

**Day 49 Afternoon (4 hours)**
- Create launch checklist
- Prepare rollback plan
- Set up monitoring alerts
- Brief support team

**Day 50 Morning (4 hours)**
- Production deployment
- Monitor system health
- Verify all integrations
- Document any issues

**Day 50 Afternoon (4 hours)**
- Post-launch review
- Update flow journal
- Plan future enhancements
- Celebrate achievement! 🎉

---

## 📊 Success Metrics & Monitoring

### System Health Dashboard
```gherkin
Feature: Comprehensive System Monitoring
  As a system operator
  I want real-time health metrics
  So that I can ensure reliable operation

  Scenario: Monitor all critical metrics
    Given the system is running
    When I check the Grafana dashboard
    Then I see:
      | Metric                  | Target              | Alert Threshold |
      | TWS Connection         | Always Connected    | Any disconnect  |
      | Subscription Usage     | <85%                | >90%           |
      | Request Queue Depth    | <50                 | >100           |
      | Scanner Performance    | <100ms              | >500ms         |
      | Order Fill Rate        | >98%                | <95%           |
      | API Error Rate         | <0.1%               | >1%            |
      | System Uptime          | >99.9%              | Any downtime   |
      | Response Time (p99)    | <200ms              | >1000ms        |
```

### Key Performance Indicators
1. **Trading Efficiency**
   - Average time from signal to execution: <2 seconds
   - Order fill rate: >98%
   - Slippage: <0.5% of spread width

2. **System Performance**
   - Scanner throughput: >1000 contracts/second
   - GUI responsiveness: <50ms for all actions
   - API latency (p99): <200ms

3. **Reliability**
   - Uptime: >99.9%
   - Successful TWS reconnections: 100%
   - Zero data loss during restarts

---

## 📝 Daily Flow Journal Template

```markdown
# Flow Journal - Day [X] - [Date]

## 🌅 Morning Intention
- Energy level: [1-10]
- Focus area: [What excites me today]
- Vibe: [Current mood/energy]

## 🚀 Session Highlights

### Breakthroughs
- [Key insights or aha moments]
- [Code patterns that clicked]
- [Problems solved elegantly]

### Challenges
- [What blocked flow]
- [How I worked around it]
- [Lessons learned]

### Code Snippets
```[language]
// Interesting patterns discovered
// Beautiful solutions found
```

## 📚 API Learnings
- [TWS API specific discoveries]
- [ib-insync patterns that worked well]
- [Event handling insights]
- [Performance optimizations]

## 🎯 Progress Check
- [ ] Maintained flow state
- [ ] Updated documentation
- [ ] Committed with story
- [ ] No pacing violations
- [ ] Tests passing

## 🌊 Tomorrow's Flow
- [What I'm excited to tackle next]
- [Energy-matched tasks]
- [Ideas to explore]

## 🎨 Vibe Check
- Flow state achieved: [Yes/No/Partial]
- Best working music: [What helped focus]
- Environment notes: [What worked/didn't]
- Overall satisfaction: [1-10]
```

---

## 🚀 Beyond MVP: Future Enhancements

### Phase 5: Advanced Strategies (15 Days)
- Iron Condors and Butterflies
- Calendar spreads
- Diagonal spreads
- Advanced rolling logic
- Multi-leg optimization

### Phase 6: Machine Learning (20 Days)
- Pattern recognition for opportunities
- Optimal parameter learning
- Entry/exit timing models
- Risk prediction
- Automated strategy tuning

### Phase 7: Scale & Enterprise (20 Days)
- Multi-account support
- Team collaboration
- Advanced audit trails
- Compliance reporting
- White-label capabilities

---

## 🎯 Final Success Checklist

### Technical Excellence
- [x] Sub-second scanner performance
- [x] 99.9% order execution reliability
- [x] Zero pacing violations design
- [x] Smooth TWS restart handling
- [x] Comprehensive error recovery
- [x] Real-time monitoring

### User Experience
- [x] 3-click navigation to any feature
- [x] Intuitive parameter controls
- [x] Real-time visual feedback
- [x] Clear execution confirmations
- [x] Helpful error messages
- [x] System status visibility

### Development Experience
- [x] Maintained flow state principles
- [x] Living documentation
- [x] Clean async architecture
- [x] Comprehensive testing
- [x] Smooth deployment
- [x] Monitoring from day one

---

This unified master plan combines the comprehensive structure of v1 with the critical async architecture and monitoring insights from v2, creating a complete roadmap for building a production-ready automated trading system while maintaining vibe coding principles throughout the journey.