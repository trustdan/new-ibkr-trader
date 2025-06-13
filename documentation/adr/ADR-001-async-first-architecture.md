# ADR-001: Async-First Architecture

## Status
Accepted

## Context

The IBKR TWS API is fundamentally event-driven, sending callbacks for market data, order updates, and system events. We need to choose an architecture that:

1. Handles thousands of events per second efficiently
2. Maintains responsiveness during high market volatility
3. Scales across multiple CPU cores
4. Integrates naturally with TWS's callback model
5. Supports both Python (for TWS integration) and Go (for performance-critical components)

## Decision

We will adopt an **async-first architecture** using:
- Python's `asyncio` with `uvloop` for the TWS integration layer
- Go's goroutines for the high-performance scanner service
- Event-driven communication between components
- Non-blocking I/O throughout the system

## Rationale

### Why Async Over Sync

1. **Natural fit with TWS callbacks**
   - TWS sends events asynchronously
   - Blocking on one event delays processing others
   - Async allows concurrent event handling

2. **Performance benefits**
   ```python
   # Sync approach - sequential processing
   def handle_tick(tick):
       process_tick(tick)        # 10ms
       update_database(tick)     # 50ms
       notify_subscribers(tick)  # 20ms
       # Total: 80ms per tick
   
   # Async approach - concurrent processing
   async def handle_tick(tick):
       await asyncio.gather(
           process_tick(tick),      # 10ms
           update_database(tick),   # 50ms
           notify_subscribers(tick) # 20ms
       )
       # Total: 50ms per tick (limited by slowest operation)
   ```

3. **Resource efficiency**
   - Single thread can handle thousands of concurrent operations
   - Lower memory footprint than thread-per-connection
   - Better CPU cache utilization

### Why Async Over Threading

1. **GIL limitations in Python**
   - Python's Global Interpreter Lock prevents true parallelism
   - Threads add complexity without performance benefit
   - Async provides concurrency without GIL contention

2. **Predictable behavior**
   - No race conditions within single event loop
   - Easier to reason about shared state
   - Deterministic execution order

3. **Integration advantages**
   - Modern libraries support async natively
   - Better compatibility with web frameworks
   - Natural fit for WebSocket/streaming APIs

### Implementation Details

1. **Python Service (TWS Integration)**
   ```python
   # Using uvloop for 2-4x performance over standard asyncio
   import uvloop
   asyncio.set_event_loop_policy(uvloop.EventLoopPolicy())
   ```

2. **Go Service (Scanner)**
   ```go
   // Goroutines for concurrent scanning
   for _, filter := range filters {
       go func(f Filter) {
           results := scanner.Scan(f)
           resultsChan <- results
       }(filter)
   }
   ```

3. **Inter-Service Communication**
   - REST APIs for request/response
   - WebSockets for streaming data
   - Message queues for event distribution

## Consequences

### Positive
- 10-100x better performance than synchronous approach
- Natural handling of TWS event stream
- Scales horizontally and vertically
- Modern, maintainable codebase
- Better resource utilization

### Negative
- Steeper learning curve for developers new to async
- Debugging can be more complex
- Need careful design to avoid callback hell
- Some libraries may not support async

### Mitigation Strategies

1. **Learning curve**
   - Comprehensive documentation
   - Code examples and patterns
   - Gradual onboarding

2. **Debugging complexity**
   - Structured logging with correlation IDs
   - Distributed tracing
   - Async-aware debugging tools

3. **Callback hell**
   - Use async/await syntax
   - Break complex flows into small functions
   - Leverage async context managers

## References
- [Python asyncio documentation](https://docs.python.org/3/library/asyncio.html)
- [uvloop performance benchmarks](https://github.com/MagicStack/uvloop)
- [Go concurrency patterns](https://go.dev/blog/pipelines)
- [TWS API asynchronous considerations](https://interactivebrokers.github.io/tws-api/)

## Decision Date
2025-01-13

## Participants
- Architecture Team
- Development Team
- Operations Team