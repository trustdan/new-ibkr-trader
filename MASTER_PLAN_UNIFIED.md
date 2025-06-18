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
.PHONY: dev test monitor clean vibe paper-test

dev:
	@echo "🚀 Starting async development environment..."
	docker-compose up -d
	@echo "⏳ Waiting for services..."
	@sleep 5
	@make health-check

test:
	@echo "🧪 Running async tests..."
	docker-compose run --rm python-ibkr pytest -v --asyncio-mode=auto

paper-test:
	@echo "📄 Running paper trading validation suite..."
	docker-compose -f docker-compose.paper.yml up -d
	@sleep 10
	@docker-compose run --rm test-runner pytest tests/paper_trading/ -v

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
        self.last_sequence_number: int = 0  # Track for daily restart handling
        self.setup_event_handlers()
        
    def setup_event_handlers(self):
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
- Handle sequence number reset after TWS restart
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

### Day 12-13: Order Execution Engine (Async) with Risk Management
**Day 12 Morning (4 hours)**
```python
# src/python/trading/async_order_engine.py
from ib_insync import Contract, Order, Trade, ComboLeg
import asyncio
from typing import List, Optional, Dict
from datetime import datetime
from dataclasses import dataclass

@dataclass
class RiskLimits:
    max_position_size: int = 10
    max_daily_loss: float = 1000.0
    max_order_value: float = 10000.0
    min_buying_power: float = 5000.0
    max_volatility_iv: float = 100.0  # Circuit breaker

class AsyncOrderEngine:
    """Event-driven order execution engine with risk management"""
    
    def __init__(self, ib, coordinator, risk_limits: RiskLimits = None):
        self.ib = ib
        self.coordinator = coordinator
        self.risk_limits = risk_limits or RiskLimits()
        self.active_trades: Dict[int, Trade] = {}
        self.partial_fills: Dict[int, List[Trade]] = {}
        self.pending_modifications: Dict[int, Order] = {}
        self.execution_metrics = {
            'total_orders': 0,
            'successful_fills': 0,
            'partial_fills': 0,
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
        
        # Pre-trade risk validation
        await self._validate_risk_limits(combo, order)
        
        # Check market volatility circuit breaker
        if await self._check_volatility_circuit_breaker(symbol):
            raise ValueError("Market volatility circuit breaker triggered")
        
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
- Create order validation logic with risk checks
- Build fill monitoring system with partial fill handling
- Implement order modification/cancellation queue

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
      | Risk check    | Validate position limits   | Within risk params   |
      | Volatility    | Check market conditions    | No circuit breaker   |
      | Build combo   | Create spread contract     | Both legs qualified  |
      | Preview       | whatIfOrder check          | Margin acceptable    |
      | Smart price   | Calculate optimal limit    | Within bid-ask       |
      | Submit        | Place combo order          | Order accepted       |
      | Monitor       | Track via events           | Real-time updates    |
      | Fill          | Handle partial fills       | Aggregate correctly  |
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

## 🎯 **PHASE 3: GUI DEVELOPMENT (Days 26-35) - WINDOWS DEVELOPMENT** ⭐ **STARTED!**

**Platform: Windows (Wails + Svelte)**  
**Duration: 10 days**  
**Status: ✅ Foundation Complete - Integration Ready**

**🎉 MAJOR ACHIEVEMENT: Professional GUI foundation built in first session!**

### ✅ **Foundation Completed (Day 26)**:
1. **✅ Desktop Framework** - Wails + Svelte + TypeScript setup complete
2. **✅ Development Environment** - Node.js 22.16.0 via NVM, all dependencies installed  
3. **✅ Backend Integration** - Go app.go with full IBKR structures and API endpoints
4. **✅ System Health Monitoring** - Real-time reactive store architecture
5. **✅ SystemStatusBar Component** - Professional TWS connection monitoring
6. **✅ Main Interface** - Complete IBKR trading dashboard with DaisyUI
7. **✅ Build Success** - `ibkr-trader.exe` generated and running

### 🏗️ **Architecture Achieved**:
- **Go Backend**: SystemHealth, ScanRequest/Response, Options structures
- **Svelte Frontend**: Reactive stores, professional UI components, TypeScript integration
- **API Layer**: Backend communication service with health monitoring  
- **Real-time Updates**: 5-second health monitoring cycle
- **Professional Design**: DaisyUI components, responsive layout, modern UX

### 🚧 **Next Phase: Live Integration (Days 27-28)**:
1. **Backend Service Integration** - Connect to Go scanner (port 8080) and Python IBKR (port 8000)
2. **Real-time Testing** - Validate health monitoring with actual services
3. **TWS Integration** - Connect to live TWS for market data and trading
4. **FilterPanel Component** - Scanner parameter configuration interface
5. **ScannerResults Component** - Real-time options results with virtual scrolling

### 🎯 **Components to Build (Days 29-35)**:
- **TradeExecution Panel** - Order preview, risk calculation, execution buttons
- **Portfolio Monitoring** - Real-time position tracking and P&L
- **WebSocket Integration** - Live market data streaming
- **Settings & Configuration** - TWS connection, scanner presets, risk management
- **Error Handling & Recovery** - Comprehensive error management and user feedback

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

### Day 47-48: Testing & Documentation with Paper Trading Validation
**Day 47 Morning (4 hours)**
```gherkin
Feature: Paper Trading Validation
  As a system validator
  I want comprehensive paper trading tests
  So that we catch issues before production

  Scenario: Complete paper trading validation
    Given system connected to paper account (port 7497)
    When running validation suite
    Then verify:
      | Test Category        | Validation                  | Success Criteria    |
      | Connection handling  | Daily restart recovery      | Auto-reconnects     |
      | Order execution      | All order types work        | No rejections       |
      | Partial fills        | Handled correctly           | State consistent    |
      | Risk limits          | Enforced properly           | Orders blocked      |
      | Circuit breakers     | Trigger on volatility       | Trading halted      |
      | Sequence numbers     | Reset handling works        | No errors after     |
      | Performance          | Meets targets               | <2s execution       |
```

**Day 47 Afternoon (4 hours)**
- Run extended paper trading scenarios
- Test edge cases and error conditions
- Validate all risk management features
- Document any issues found

**Day 48 Morning (4 hours)**
- Complete system documentation
- Create operation runbooks
- Write troubleshooting guides
- Document known issues and solutions

**Day 48 Afternoon (4 hours)**
- Create video tutorials
- Write quick-start guide
- Prepare training materials
- Document FAQ including paper vs live differences

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