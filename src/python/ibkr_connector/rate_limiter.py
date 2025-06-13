"""
Rate limiting for IBKR API requests.

This module implements a token bucket algorithm to prevent exceeding
the TWS API rate limits (50 requests/second).
"""

import asyncio
import time
import logging
from typing import Optional, Dict, Any
from dataclasses import dataclass
from collections import deque
from datetime import datetime, timedelta

from ..config.settings import RateLimitConfig
from .exceptions import RateLimitError


@dataclass
class RequestStats:
    """Statistics for rate limiter performance."""
    total_requests: int = 0
    rejected_requests: int = 0
    queued_requests: int = 0
    average_wait_time: float = 0.0
    current_rate: float = 0.0
    last_reset: datetime = None


class RateLimiter:
    """
    Token bucket rate limiter for IBKR API requests.
    
    Implements a token bucket algorithm with:
    - Configurable rate limit (default 45 req/sec for safety)
    - Burst capacity for short request spikes
    - Request queuing with priority support
    - Performance statistics
    """
    
    def __init__(self, config: Optional[RateLimitConfig] = None):
        """
        Initialize the rate limiter.
        
        Args:
            config: Rate limit configuration
        """
        self.config = config or RateLimitConfig()
        self.logger = logging.getLogger(__name__)
        
        # Token bucket state
        self._tokens = float(self.config.max_requests_per_second)
        self._max_tokens = float(self.config.max_requests_per_second)
        self._refill_rate = float(self.config.max_requests_per_second)
        self._last_refill = time.monotonic()
        
        # Request queue
        self._queue: deque = deque()
        self._lock = asyncio.Lock()
        
        # Statistics
        self.stats = RequestStats(last_reset=datetime.now())
        self._request_times = deque(maxlen=100)  # Track last 100 requests
        
        self.logger.info(
            f"Rate limiter initialized: {self.config.max_requests_per_second} req/sec"
        )
    
    async def acquire(self, priority: int = 0, timeout: Optional[float] = None) -> float:
        """
        Acquire permission to make a request.
        
        Args:
            priority: Request priority (higher = more important)
            timeout: Maximum time to wait for permission
            
        Returns:
            Time waited in seconds
            
        Raises:
            RateLimitError: If timeout exceeded or rate limit cannot be satisfied
        """
        start_time = time.monotonic()
        
        async with self._lock:
            # Refill tokens based on time elapsed
            self._refill_tokens()
            
            # If we have tokens, consume one immediately
            if self._tokens >= 1:
                self._consume_token()
                wait_time = 0.0
            else:
                # Calculate wait time needed
                tokens_needed = 1 - self._tokens
                wait_time = tokens_needed / self._refill_rate
                
                if timeout and wait_time > timeout:
                    self.stats.rejected_requests += 1
                    raise RateLimitError(
                        f"Rate limit would require {wait_time:.2f}s wait, "
                        f"exceeds timeout of {timeout}s",
                        retry_after=wait_time
                    )
                
                # Wait for tokens to be available
                self.stats.queued_requests += 1
                await asyncio.sleep(wait_time)
                
                # Refill and consume
                self._refill_tokens()
                self._consume_token()
        
        # Update statistics
        self.stats.total_requests += 1
        self._request_times.append(time.monotonic())
        self._update_current_rate()
        
        actual_wait = time.monotonic() - start_time
        self._update_average_wait_time(actual_wait)
        
        if actual_wait > 0.1:  # Log if we had to wait significantly
            self.logger.debug(f"Rate limit wait: {actual_wait:.3f}s")
        
        return actual_wait
    
    def _refill_tokens(self) -> None:
        """Refill tokens based on time elapsed."""
        now = time.monotonic()
        elapsed = now - self._last_refill
        
        # Add tokens based on time elapsed
        tokens_to_add = elapsed * self._refill_rate
        self._tokens = min(self._max_tokens, self._tokens + tokens_to_add)
        self._last_refill = now
    
    def _consume_token(self) -> None:
        """Consume a token for a request."""
        self._tokens = max(0, self._tokens - 1)
    
    def _update_current_rate(self) -> None:
        """Update the current request rate calculation."""
        if len(self._request_times) < 2:
            self.stats.current_rate = 0.0
            return
        
        # Calculate rate over the last second
        now = time.monotonic()
        one_second_ago = now - 1.0
        recent_requests = sum(1 for t in self._request_times if t > one_second_ago)
        self.stats.current_rate = float(recent_requests)
    
    def _update_average_wait_time(self, wait_time: float) -> None:
        """Update rolling average wait time."""
        # Simple exponential moving average
        alpha = 0.1  # Smoothing factor
        self.stats.average_wait_time = (
            alpha * wait_time + (1 - alpha) * self.stats.average_wait_time
        )
    
    async def check_rate(self) -> bool:
        """
        Check if a request can be made without waiting.
        
        Returns:
            True if request can be made immediately
        """
        async with self._lock:
            self._refill_tokens()
            return self._tokens >= 1
    
    def get_stats(self) -> Dict[str, Any]:
        """
        Get current rate limiter statistics.
        
        Returns:
            Dictionary of statistics
        """
        return {
            'total_requests': self.stats.total_requests,
            'rejected_requests': self.stats.rejected_requests,
            'queued_requests': self.stats.queued_requests,
            'average_wait_time': f"{self.stats.average_wait_time:.3f}s",
            'current_rate': f"{self.stats.current_rate:.1f} req/sec",
            'tokens_available': int(self._tokens),
            'max_rate': f"{self.config.max_requests_per_second} req/sec",
            'uptime': str(datetime.now() - self.stats.last_reset)
        }
    
    def reset_stats(self) -> None:
        """Reset statistics counters."""
        self.stats = RequestStats(last_reset=datetime.now())
        self._request_times.clear()
        self.logger.info("Rate limiter statistics reset")
    
    def __repr__(self) -> str:
        """String representation of rate limiter."""
        return (
            f"RateLimiter(rate={self.config.max_requests_per_second} req/sec, "
            f"tokens={self._tokens:.1f}, current={self.stats.current_rate:.1f} req/sec)"
        )