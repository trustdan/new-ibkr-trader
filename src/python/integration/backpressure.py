"""
Backpressure handling for scanner requests

This module implements various backpressure strategies to prevent overwhelming
the scanner service and maintain system stability.
"""

import asyncio
from typing import Dict, Optional, Callable, Any
from datetime import datetime, timedelta
from collections import deque
import logging
from enum import Enum
from dataclasses import dataclass

logger = logging.getLogger(__name__)


class BackpressureStrategy(str, Enum):
    """Available backpressure strategies"""
    FIXED_WINDOW = "fixed_window"
    SLIDING_WINDOW = "sliding_window"
    TOKEN_BUCKET = "token_bucket"
    ADAPTIVE = "adaptive"


@dataclass
class RequestMetrics:
    """Metrics for a single request"""
    timestamp: datetime
    duration: float
    success: bool
    queue_time: float = 0.0


class BackpressureHandler:
    """
    Handles backpressure for scanner requests
    
    Features:
    - Multiple rate limiting strategies
    - Queue management with priorities
    - Adaptive throttling based on response times
    - Circuit breaker pattern
    """
    
    def __init__(
        self,
        strategy: BackpressureStrategy = BackpressureStrategy.TOKEN_BUCKET,
        requests_per_second: float = 10.0,
        burst_size: int = 20,
        queue_size: int = 100,
        circuit_breaker_threshold: int = 5,
        circuit_breaker_timeout: int = 60
    ):
        self.strategy = strategy
        self.requests_per_second = requests_per_second
        self.burst_size = burst_size
        self.queue_size = queue_size
        
        # Circuit breaker
        self.circuit_breaker_threshold = circuit_breaker_threshold
        self.circuit_breaker_timeout = circuit_breaker_timeout
        self.circuit_breaker_failures = 0
        self.circuit_breaker_opened_at: Optional[datetime] = None
        
        # Request queue
        self.request_queue: asyncio.Queue = asyncio.Queue(maxsize=queue_size)
        self.priority_queue: asyncio.PriorityQueue = asyncio.PriorityQueue()
        
        # Rate limiting
        self.tokens = float(burst_size)
        self.last_refill = datetime.now()
        
        # Metrics
        self.request_history: deque = deque(maxlen=1000)
        self.current_qps = 0.0
        self.average_response_time = 0.0
        
        # Adaptive throttling
        self.adaptive_rate = requests_per_second
        self.target_response_time = 1.0  # Target 1 second response
        
    async def acquire(self, priority: int = 5) -> bool:
        """
        Acquire permission to make a request
        
        Args:
            priority: Request priority (1=highest, 10=lowest)
            
        Returns:
            True if request can proceed, False if rejected
        """
        # Check circuit breaker
        if self._is_circuit_open():
            logger.warning("Circuit breaker is open, rejecting request")
            return False
            
        # Apply rate limiting based on strategy
        if self.strategy == BackpressureStrategy.TOKEN_BUCKET:
            return await self._token_bucket_acquire()
        elif self.strategy == BackpressureStrategy.SLIDING_WINDOW:
            return await self._sliding_window_acquire()
        elif self.strategy == BackpressureStrategy.ADAPTIVE:
            return await self._adaptive_acquire()
        else:
            return await self._fixed_window_acquire()
            
    async def _token_bucket_acquire(self) -> bool:
        """Token bucket rate limiting"""
        now = datetime.now()
        
        # Refill tokens
        time_passed = (now - self.last_refill).total_seconds()
        tokens_to_add = time_passed * self.requests_per_second
        self.tokens = min(self.burst_size, self.tokens + tokens_to_add)
        self.last_refill = now
        
        # Try to acquire token
        if self.tokens >= 1:
            self.tokens -= 1
            return True
            
        # Calculate wait time
        tokens_needed = 1 - self.tokens
        wait_time = tokens_needed / self.requests_per_second
        
        if wait_time < 0.1:  # Wait if it's short
            await asyncio.sleep(wait_time)
            self.tokens = 0
            return True
            
        return False
        
    async def _sliding_window_acquire(self) -> bool:
        """Sliding window rate limiting"""
        now = datetime.now()
        window_start = now - timedelta(seconds=1)
        
        # Count recent requests
        recent_requests = [
            r for r in self.request_history
            if r.timestamp > window_start
        ]
        
        if len(recent_requests) < self.requests_per_second:
            return True
            
        # Calculate when next slot opens
        oldest_in_window = min(recent_requests, key=lambda r: r.timestamp)
        wait_time = (oldest_in_window.timestamp + timedelta(seconds=1) - now).total_seconds()
        
        if wait_time > 0 and wait_time < 0.1:
            await asyncio.sleep(wait_time)
            return True
            
        return False
        
    async def _adaptive_acquire(self) -> bool:
        """Adaptive rate limiting based on response times"""
        # Adjust rate based on response times
        if self.average_response_time > self.target_response_time * 1.5:
            # Slow down if responses are too slow
            self.adaptive_rate = max(1, self.adaptive_rate * 0.9)
            logger.info(f"Reducing rate to {self.adaptive_rate:.1f} req/s")
        elif self.average_response_time < self.target_response_time * 0.5:
            # Speed up if responses are fast
            self.adaptive_rate = min(
                self.requests_per_second * 2,
                self.adaptive_rate * 1.1
            )
            logger.info(f"Increasing rate to {self.adaptive_rate:.1f} req/s")
            
        # Use token bucket with adaptive rate
        original_rate = self.requests_per_second
        self.requests_per_second = self.adaptive_rate
        result = await self._token_bucket_acquire()
        self.requests_per_second = original_rate
        
        return result
        
    async def _fixed_window_acquire(self) -> bool:
        """Fixed window rate limiting"""
        # Simple implementation - just use token bucket
        return await self._token_bucket_acquire()
        
    def record_request(
        self,
        duration: float,
        success: bool,
        queue_time: float = 0.0
    ):
        """
        Record metrics for a completed request
        
        Args:
            duration: Request duration in seconds
            success: Whether request succeeded
            queue_time: Time spent in queue
        """
        metric = RequestMetrics(
            timestamp=datetime.now(),
            duration=duration,
            success=success,
            queue_time=queue_time
        )
        
        self.request_history.append(metric)
        
        # Update circuit breaker
        if success:
            self.circuit_breaker_failures = 0
        else:
            self.circuit_breaker_failures += 1
            if self.circuit_breaker_failures >= self.circuit_breaker_threshold:
                self._open_circuit_breaker()
                
        # Update metrics
        self._update_metrics()
        
    def _open_circuit_breaker(self):
        """Open the circuit breaker"""
        self.circuit_breaker_opened_at = datetime.now()
        logger.warning(
            f"Circuit breaker opened after {self.circuit_breaker_failures} failures"
        )
        
    def _is_circuit_open(self) -> bool:
        """Check if circuit breaker is open"""
        if not self.circuit_breaker_opened_at:
            return False
            
        # Check if timeout has passed
        time_open = (datetime.now() - self.circuit_breaker_opened_at).total_seconds()
        if time_open > self.circuit_breaker_timeout:
            # Reset circuit breaker
            self.circuit_breaker_opened_at = None
            self.circuit_breaker_failures = 0
            logger.info("Circuit breaker reset")
            return False
            
        return True
        
    def _update_metrics(self):
        """Update performance metrics"""
        now = datetime.now()
        recent_window = now - timedelta(seconds=10)
        
        recent_requests = [
            r for r in self.request_history
            if r.timestamp > recent_window
        ]
        
        if recent_requests:
            # Calculate QPS
            time_span = (now - min(r.timestamp for r in recent_requests)).total_seconds()
            self.current_qps = len(recent_requests) / max(1, time_span)
            
            # Calculate average response time
            successful_requests = [r for r in recent_requests if r.success]
            if successful_requests:
                self.average_response_time = sum(
                    r.duration for r in successful_requests
                ) / len(successful_requests)
                
    def get_metrics(self) -> Dict[str, Any]:
        """Get current backpressure metrics"""
        return {
            "strategy": self.strategy.value,
            "current_qps": round(self.current_qps, 2),
            "average_response_time": round(self.average_response_time, 3),
            "tokens_available": round(self.tokens, 2),
            "circuit_breaker_open": self._is_circuit_open(),
            "circuit_breaker_failures": self.circuit_breaker_failures,
            "queue_size": self.request_queue.qsize() if hasattr(self.request_queue, 'qsize') else 0,
            "adaptive_rate": round(self.adaptive_rate, 2) if self.strategy == BackpressureStrategy.ADAPTIVE else None
        }
        
    async def wait_if_needed(self):
        """Wait if necessary to respect rate limits"""
        retries = 0
        while retries < 3:
            if await self.acquire():
                return
            retries += 1
            await asyncio.sleep(0.1 * retries)
            
        raise Exception("Failed to acquire rate limit permit")


class PriorityBackpressureHandler(BackpressureHandler):
    """Extended backpressure handler with priority queue support"""
    
    async def submit_request(
        self,
        request_fn: Callable,
        priority: int = 5,
        timeout: float = 30.0
    ) -> Any:
        """
        Submit a request with priority handling
        
        Args:
            request_fn: Async function to execute
            priority: Request priority (1=highest)
            timeout: Request timeout
            
        Returns:
            Result of request_fn
        """
        # Add to priority queue
        queue_entry = (priority, datetime.now(), request_fn)
        await self.priority_queue.put(queue_entry)
        
        # Wait for our turn
        start_time = datetime.now()
        
        while True:
            # Get next item
            _, queued_time, fn = await self.priority_queue.get()
            
            if fn == request_fn:
                # Our turn
                queue_time = (datetime.now() - queued_time).total_seconds()
                
                # Acquire rate limit
                await self.wait_if_needed()
                
                # Execute request
                try:
                    result = await asyncio.wait_for(fn(), timeout=timeout)
                    duration = (datetime.now() - start_time).total_seconds()
                    self.record_request(duration, True, queue_time)
                    return result
                except Exception as e:
                    duration = (datetime.now() - start_time).total_seconds()
                    self.record_request(duration, False, queue_time)
                    raise
            else:
                # Not our turn, put it back
                await self.priority_queue.put((priority, queued_time, fn))
                await asyncio.sleep(0.01)


# Example usage
async def example_backpressure():
    """Example of using backpressure handler"""
    
    handler = BackpressureHandler(
        strategy=BackpressureStrategy.ADAPTIVE,
        requests_per_second=5.0
    )
    
    async def make_request(i: int):
        """Simulate a request"""
        if await handler.acquire():
            start = datetime.now()
            await asyncio.sleep(0.1)  # Simulate work
            duration = (datetime.now() - start).total_seconds()
            handler.record_request(duration, True)
            return f"Request {i} completed"
        else:
            return f"Request {i} rejected"
            
    # Make concurrent requests
    tasks = [make_request(i) for i in range(20)]
    results = await asyncio.gather(*tasks)
    
    for result in results:
        print(result)
        
    # Show metrics
    metrics = handler.get_metrics()
    print(f"\nMetrics: {metrics}")


if __name__ == "__main__":
    asyncio.run(example_backpressure())