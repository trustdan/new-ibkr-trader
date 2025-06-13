# Event-Driven Architecture for IBKR Trading System

## Overview

Our system uses an event-driven architecture that mirrors TWS's callback-based API design. This approach provides natural handling of asynchronous market events, order updates, and system state changes.

```
┌─────────────────┐     Events      ┌──────────────────┐
│                 │ ◄────────────── │                  │
│   TWS Client    │                 │   Go Scanner     │
│  (Python/async) │ ───Requests───► │   (REST API)     │
└────────┬────────┘                 └──────────────────┘
         │                                    │
         │ Events                             │ Results
         ▼                                    ▼
┌─────────────────┐                 ┌──────────────────┐
│  Event Router   │                 │  Results Cache   │
│  & Dispatcher   │                 │   (In-Memory)    │
└────────┬────────┘                 └──────────────────┘
         │
         ├─────────────┬──────────────┬───────────────┐
         ▼             ▼              ▼               ▼
   Order Handler  Market Data    Scanner Handler  Position Handler
```

## Core Components

### 1. TWS Event Stream

The TWS API generates events for all market activity:

```python
class TWS EventTypes(Enum):
    # Connection Events
    CONNECTED = "connected"
    DISCONNECTED = "disconnected"
    ERROR = "error"
    
    # Market Data Events
    TICK_PRICE = "tickPrice"
    TICK_SIZE = "tickSize"
    TICK_OPTION_COMPUTATION = "tickOptionComputation"
    TICK_GENERIC = "tickGeneric"
    
    # Order Events
    ORDER_STATUS = "orderStatus"
    OPEN_ORDER = "openOrder"
    EXEC_DETAILS = "execDetails"
    COMMISSION_REPORT = "commissionReport"
    
    # Account Events
    UPDATE_ACCOUNT_VALUE = "updateAccountValue"
    UPDATE_PORTFOLIO = "updatePortfolio"
    ACCOUNT_DOWNLOAD_END = "accountDownloadEnd"
    
    # Scanner Events
    SCANNER_DATA = "scannerData"
    SCANNER_DATA_END = "scannerDataEnd"
```

### 2. Event Router Implementation

```python
import asyncio
from typing import Dict, List, Callable, Any
from collections import defaultdict

class EventRouter:
    """Central event routing system for TWS events"""
    
    def __init__(self):
        self._handlers: Dict[str, List[Callable]] = defaultdict(list)
        self._async_handlers: Dict[str, List[Callable]] = defaultdict(list)
        self._middleware: List[Callable] = []
        self._event_queue = asyncio.Queue(maxsize=10000)
        self._running = False
        
    def on(self, event_type: str, handler: Callable):
        """Register synchronous event handler"""
        self._handlers[event_type].append(handler)
        
    def on_async(self, event_type: str, handler: Callable):
        """Register asynchronous event handler"""
        self._async_handlers[event_type].append(handler)
        
    def use_middleware(self, middleware: Callable):
        """Add middleware for all events"""
        self._middleware.append(middleware)
        
    async def emit(self, event_type: str, data: Any):
        """Emit event to all registered handlers"""
        event = Event(event_type, data)
        
        # Apply middleware
        for mw in self._middleware:
            event = await mw(event)
            if event is None:
                return  # Middleware filtered event
        
        # Queue for async processing
        await self._event_queue.put(event)
        
    async def _process_events(self):
        """Main event processing loop"""
        while self._running:
            try:
                event = await asyncio.wait_for(
                    self._event_queue.get(), 
                    timeout=1.0
                )
                
                # Process sync handlers
                for handler in self._handlers[event.type]:
                    try:
                        handler(event)
                    except Exception as e:
                        logger.error(f"Sync handler error: {e}")
                
                # Process async handlers
                tasks = [
                    handler(event) 
                    for handler in self._async_handlers[event.type]
                ]
                if tasks:
                    await asyncio.gather(*tasks, return_exceptions=True)
                    
            except asyncio.TimeoutError:
                continue
            except Exception as e:
                logger.error(f"Event processing error: {e}")
```

### 3. Domain Event Handlers

#### Order Event Handler
```python
class OrderEventHandler:
    """Handles all order-related events"""
    
    def __init__(self, event_router: EventRouter, order_manager: OrderManager):
        self.router = event_router
        self.order_manager = order_manager
        
        # Register for order events
        self.router.on_async("orderStatus", self.handle_order_status)
        self.router.on_async("openOrder", self.handle_open_order)
        self.router.on_async("execDetails", self.handle_execution)
        
    async def handle_order_status(self, event: Event):
        """Process order status updates"""
        data = event.data
        order_id = data['orderId']
        status = data['status']
        
        # Update internal state
        await self.order_manager.update_order_status(order_id, status)
        
        # Handle specific statuses
        if status == "Filled":
            await self.handle_order_filled(order_id, data)
        elif status == "Cancelled":
            await self.handle_order_cancelled(order_id, data)
        elif status in ["Inactive", "ApiCancelled"]:
            await self.handle_order_rejected(order_id, data)
            
    async def handle_order_filled(self, order_id: int, data: dict):
        """Handle filled orders"""
        logger.info(f"Order {order_id} filled")
        
        # Emit domain event
        await self.router.emit("order.filled", {
            'orderId': order_id,
            'avgFillPrice': data['avgFillPrice'],
            'filled': data['filled'],
            'timestamp': datetime.now()
        })
```

#### Market Data Event Handler
```python
class MarketDataHandler:
    """Handles market data events with aggregation"""
    
    def __init__(self, event_router: EventRouter):
        self.router = event_router
        self.price_aggregator = PriceAggregator()
        self.option_calculator = OptionGreeksCalculator()
        
        # Register handlers
        self.router.on("tickPrice", self.handle_tick_price)
        self.router.on("tickSize", self.handle_tick_size)
        self.router.on("tickOptionComputation", self.handle_option_tick)
        
    def handle_tick_price(self, event: Event):
        """Process price ticks"""
        data = event.data
        req_id = data['reqId']
        tick_type = data['tickType']
        price = data['price']
        
        # Aggregate prices
        self.price_aggregator.update(req_id, tick_type, price)
        
        # Emit aggregated data every N ticks or T seconds
        if self.price_aggregator.should_emit(req_id):
            aggregated = self.price_aggregator.get_aggregated(req_id)
            asyncio.create_task(
                self.router.emit("marketData.aggregated", aggregated)
            )
```

### 4. Event Aggregation Patterns

#### Time-Window Aggregation
```python
class TimeWindowAggregator:
    """Aggregate events over time windows"""
    
    def __init__(self, window_seconds: int = 1):
        self.window_seconds = window_seconds
        self.windows: Dict[str, List[Event]] = defaultdict(list)
        self.window_starts: Dict[str, datetime] = {}
        
    def add_event(self, key: str, event: Event):
        """Add event to aggregation window"""
        now = datetime.now()
        
        # Initialize window if needed
        if key not in self.window_starts:
            self.window_starts[key] = now
            
        # Check if window expired
        if (now - self.window_starts[key]).total_seconds() > self.window_seconds:
            # Process window
            aggregated = self._aggregate_window(key)
            
            # Reset window
            self.windows[key] = [event]
            self.window_starts[key] = now
            
            return aggregated
        else:
            self.windows[key].append(event)
            return None
    
    def _aggregate_window(self, key: str) -> dict:
        """Aggregate events in window"""
        events = self.windows[key]
        if not events:
            return None
            
        return {
            'key': key,
            'count': len(events),
            'window_start': self.window_starts[key],
            'window_end': datetime.now(),
            'events': events
        }
```

#### Batch Processing Pattern
```python
class BatchProcessor:
    """Process events in batches for efficiency"""
    
    def __init__(self, batch_size: int = 100, timeout: float = 1.0):
        self.batch_size = batch_size
        self.timeout = timeout
        self.batch: List[Event] = []
        self.last_process = time.time()
        
    async def add_event(self, event: Event):
        """Add event to batch"""
        self.batch.append(event)
        
        # Process if batch full or timeout reached
        should_process = (
            len(self.batch) >= self.batch_size or
            time.time() - self.last_process > self.timeout
        )
        
        if should_process:
            await self.process_batch()
    
    async def process_batch(self):
        """Process current batch"""
        if not self.batch:
            return
            
        # Process batch
        batch_to_process = self.batch
        self.batch = []
        self.last_process = time.time()
        
        # Batch operations (e.g., database inserts)
        await self._execute_batch(batch_to_process)
```

### 5. Event Store Pattern

```python
class EventStore:
    """Store and replay events for recovery and debugging"""
    
    def __init__(self, storage_path: str):
        self.storage_path = storage_path
        self.current_file = None
        self.event_index = {}
        
    async def store_event(self, event: Event):
        """Persist event to storage"""
        # Add metadata
        stored_event = {
            'id': str(uuid.uuid4()),
            'type': event.type,
            'data': event.data,
            'timestamp': event.timestamp.isoformat(),
            'sequence': self._get_next_sequence()
        }
        
        # Write to file
        async with aiofiles.open(self._get_current_file(), 'a') as f:
            await f.write(json.dumps(stored_event) + '\n')
            
        # Update index
        self.event_index[stored_event['id']] = {
            'file': self._get_current_file(),
            'offset': f.tell()
        }
    
    async def replay_events(self, start_time: datetime, 
                          end_time: datetime,
                          event_types: List[str] = None):
        """Replay events from storage"""
        events = []
        
        # Find relevant files
        files = self._get_files_in_range(start_time, end_time)
        
        for file_path in files:
            async with aiofiles.open(file_path, 'r') as f:
                async for line in f:
                    event = json.loads(line)
                    event_time = datetime.fromisoformat(event['timestamp'])
                    
                    # Filter by time and type
                    if start_time <= event_time <= end_time:
                        if event_types is None or event['type'] in event_types:
                            events.append(event)
        
        return events
```

### 6. Error Handling and Recovery

```python
class EventErrorHandler:
    """Centralized error handling for events"""
    
    def __init__(self, event_router: EventRouter):
        self.router = event_router
        self.error_counts = defaultdict(int)
        self.circuit_breakers = {}
        
    async def handle_event_error(self, event: Event, error: Exception):
        """Handle errors in event processing"""
        error_key = f"{event.type}:{type(error).__name__}"
        self.error_counts[error_key] += 1
        
        # Check circuit breaker
        if self._should_circuit_break(error_key):
            logger.error(f"Circuit breaker activated for {error_key}")
            self.circuit_breakers[error_key] = datetime.now()
            return
        
        # Retry logic
        if self._should_retry(event, error):
            await self._retry_event(event)
        else:
            # Dead letter queue
            await self._send_to_dlq(event, error)
    
    def _should_circuit_break(self, error_key: str) -> bool:
        """Check if circuit should break"""
        # Break if too many errors in short time
        return self.error_counts[error_key] > 10
        
    async def _retry_event(self, event: Event):
        """Retry event processing"""
        retry_count = getattr(event, 'retry_count', 0)
        
        if retry_count < 3:
            event.retry_count = retry_count + 1
            
            # Exponential backoff
            delay = 2 ** retry_count
            await asyncio.sleep(delay)
            
            # Re-emit event
            await self.router.emit(event.type, event.data)
```

### 7. Monitoring and Metrics

```python
class EventMetrics:
    """Track event system metrics"""
    
    def __init__(self):
        self.event_counters = defaultdict(int)
        self.event_latencies = defaultdict(list)
        self.queue_depths = {}
        
    def record_event(self, event_type: str, latency: float):
        """Record event metrics"""
        self.event_counters[event_type] += 1
        self.event_latencies[event_type].append(latency)
        
        # Trim old latencies
        if len(self.event_latencies[event_type]) > 1000:
            self.event_latencies[event_type] = \
                self.event_latencies[event_type][-1000:]
    
    def get_metrics(self) -> dict:
        """Get current metrics"""
        metrics = {}
        
        for event_type in self.event_counters:
            latencies = self.event_latencies[event_type]
            metrics[event_type] = {
                'count': self.event_counters[event_type],
                'avg_latency': sum(latencies) / len(latencies) if latencies else 0,
                'p95_latency': self._percentile(latencies, 0.95),
                'p99_latency': self._percentile(latencies, 0.99)
            }
            
        return metrics
```

## Best Practices

1. **Event Naming Convention**
   - Use dot notation: `domain.entity.action`
   - Examples: `order.spread.filled`, `scanner.results.received`

2. **Event Payload Design**
   - Include all necessary data in event
   - Avoid external lookups in handlers
   - Version events for compatibility

3. **Handler Organization**
   - One handler class per domain
   - Keep handlers focused and testable
   - Use dependency injection

4. **Performance Optimization**
   - Batch similar events
   - Use async handlers for I/O operations
   - Monitor queue depths

5. **Testing Strategy**
   - Unit test individual handlers
   - Integration test event flows
   - Load test with production-like volumes