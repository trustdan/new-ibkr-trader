"""
Event Handler Patterns - Common patterns for TWS event handling

Key principle: Event handlers must be FAST and NON-BLOCKING
"""
from typing import List, Dict, Any
import asyncio
import logging
from collections import defaultdict
from ib_insync import Trade, Ticker, Contract

logger = logging.getLogger(__name__)


class EventPatterns:
    """Collection of common event handling patterns"""
    
    def __init__(self):
        # Event aggregation
        self.event_buffer = defaultdict(list)
        self.buffer_task = None
        
    # Pattern 1: Lightweight handler with async task spawning
    def on_order_status_lightweight(self, trade: Trade):
        """Handle order status without blocking"""
        # Quick synchronous work only
        logger.info(f"Order {trade.order.orderId}: {trade.orderStatus.status}")
        
        # Spawn async task for heavy work
        if trade.isDone():
            asyncio.create_task(self._process_completed_order(trade))
            
    async def _process_completed_order(self, trade: Trade):
        """Heavy processing in separate task"""
        # This won't block the event handler
        await asyncio.sleep(0)  # Yield to event loop
        # Do complex processing here
        
    # Pattern 2: Event buffering for high-frequency events
    def on_ticker_update_buffered(self, ticker: Ticker):
        """Buffer high-frequency events"""
        self.event_buffer['tickers'].append(ticker)
        
        # Start buffer processor if not running
        if not self.buffer_task or self.buffer_task.done():
            self.buffer_task = asyncio.create_task(self._process_buffer())
            
    async def _process_buffer(self):
        """Process buffered events in batch"""
        await asyncio.sleep(0.1)  # Small delay for batching
        
        # Process all buffered events
        tickers = self.event_buffer['tickers']
        self.event_buffer['tickers'] = []
        
        if tickers:
            # Batch processing is more efficient
            await self._update_ticker_display(tickers)
            
    # Pattern 3: Event filtering to reduce noise
    def on_error_filtered(self, reqId, errorCode, errorString, contract):
        """Filter out non-critical errors"""
        # Info codes to ignore
        info_codes = {2104, 2106, 2107, 2108, 2119, 2158}
        
        if errorCode in info_codes:
            logger.debug(f"Info {errorCode}: {errorString}")
        elif errorCode == 1100:
            # Connectivity lost - critical but expected
            logger.warning("Connectivity lost - Watchdog should handle")
        else:
            logger.error(f"Error {errorCode}: {errorString}")
            
    # Pattern 4: State tracking without blocking
    class StateTracker:
        """Track state changes efficiently"""
        def __init__(self):
            self.previous_states = {}
            
        def on_position_update(self, position):
            """Track position changes"""
            key = (position.account, position.contract.symbol)
            prev = self.previous_states.get(key)
            
            # Quick comparison
            if prev != position.position:
                # State changed - handle it
                logger.info(f"Position changed: {key} {prev} -> {position.position}")
                self.previous_states[key] = position.position
                
                # Don't do heavy work here!
                # Instead, emit a custom event or set a flag
                
    # Pattern 5: Event chaining with conditions
    def setup_event_chains(self, ib):
        """Chain events with conditions"""
        # Only process fills for our orders
        def on_exec_details(trade, fill):
            if trade.order.orderId in self.my_order_ids:
                logger.info(f"Our fill: {fill.execution.execId}")
                # Process our fills only
                
        ib.execDetailsEvent += on_exec_details
        
    # Pattern 6: Async event emitter pattern
    class AsyncEventEmitter:
        """Emit events that can be awaited"""
        def __init__(self):
            self.listeners = defaultdict(list)
            
        def on(self, event: str, callback):
            """Register async callback"""
            self.listeners[event].append(callback)
            
        async def emit(self, event: str, *args, **kwargs):
            """Emit event to all listeners"""
            tasks = []
            for callback in self.listeners[event]:
                if asyncio.iscoroutinefunction(callback):
                    tasks.append(asyncio.create_task(callback(*args, **kwargs)))
                else:
                    # Sync callbacks run immediately
                    callback(*args, **kwargs)
                    
            # Wait for all async callbacks
            if tasks:
                await asyncio.gather(*tasks, return_exceptions=True)
                
    # Pattern 7: Debouncing for rapid events
    class Debouncer:
        """Debounce rapid events"""
        def __init__(self, delay: float = 0.5):
            self.delay = delay
            self.pending = {}
            
        def debounce(self, key: str, coro):
            """Debounce a coroutine"""
            # Cancel pending
            if key in self.pending:
                self.pending[key].cancel()
                
            # Schedule new
            self.pending[key] = asyncio.create_task(self._run_after_delay(key, coro))
            
        async def _run_after_delay(self, key: str, coro):
            """Run after delay"""
            await asyncio.sleep(self.delay)
            await coro
            self.pending.pop(key, None)


# Best practices reminder
"""
DO:
✓ Keep handlers fast and lightweight
✓ Use asyncio.create_task() for heavy work
✓ Buffer high-frequency events
✓ Filter noise early
✓ Track state efficiently

DON'T:
✗ Block with time.sleep() or requests
✗ Do heavy computation in handlers
✗ Make API calls from handlers
✗ Use recursive event triggering
✗ Ignore error codes
"""