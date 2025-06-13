# ADR-001: Async Event-Driven Architecture

## Status
Accepted

## Context
The IBKR TWS API requires careful handling of connections, rate limits, and market data subscriptions. The ib-insync library provides a Python wrapper around the native API, using asyncio for non-blocking operations. We need to decide on the fundamental architecture pattern for our Python service.

## Decision
We will build the Python IBKR service using a **fully async, event-driven architecture** that embraces ib-insync's patterns rather than fighting them.

## Consequences

### Positive
1. **Natural fit with ib-insync**: The library is built on asyncio and events
2. **Non-blocking operations**: GUI remains responsive during API calls  
3. **Built-in rate limiting**: ib-insync handles throttling automatically
4. **Efficient resource usage**: Single thread handles many operations
5. **Real-time updates**: Events provide immediate notification of changes
6. **Simplified error handling**: Exceptions don't block other operations

### Negative
1. **Learning curve**: Developers must understand async/await patterns
2. **Debugging complexity**: Async stack traces can be harder to follow
3. **Testing challenges**: Async tests require special handling
4. **Library constraints**: Must use async-compatible libraries

### Neutral
1. **Code style change**: All I/O operations must use await
2. **Event handler discipline**: Handlers must be lightweight
3. **Different mental model**: Think in terms of events, not requests

## Implementation Guidelines

### The One Rule
**User code may not block for too long** - This is ib-insync's fundamental constraint.

### Do's
```python
# Correct async patterns
await ib.connectAsync(host, port, clientId)
await ib.sleep(1)  # Not time.sleep()!
trade = await ib.placeOrderAsync(contract, order)

# Event-driven updates
ib.pendingTickersEvent += on_ticker_update
ib.orderStatusEvent += on_order_status
```

### Don'ts
```python
# These will break the event loop
time.sleep(1)  # NEVER do this
requests.get(url)  # Use aiohttp instead
heavy_computation()  # Farm out to process pool
```

### Event Handler Pattern
```python
def on_order_status(trade: Trade):
    # Quick processing only
    logger.info(f"Order {trade.order.orderId}: {trade.orderStatus.status}")
    
    # DON'T place new orders here (recursion risk)
    # DON'T do heavy computation
    # DON'T make blocking calls
```

## Alternatives Considered

### 1. Traditional Threading Model
- **Rejected**: Would fight against ib-insync's design
- Multiple threads would complicate state management
- No real benefit over async approach

### 2. Polling-Based Architecture  
- **Rejected**: Inefficient and increases latency
- Would require custom rate limiting
- Misses real-time updates

### 3. Hybrid Sync/Async
- **Rejected**: Complexity without benefit
- Difficult to maintain consistency
- Prone to blocking accidents

## References
- [ib-insync documentation](https://ib-insync.readthedocs.io/)
- [Python asyncio documentation](https://docs.python.org/3/library/asyncio.html)
- TWS API documentation (rate limits, threading requirements)

## Review Date
To be reviewed after Phase 1 implementation to validate decision.