"""
Unit tests for rate limiting functionality.
"""

import pytest
import asyncio
import time
from unittest.mock import patch

from src.python.ibkr_connector.rate_limiter import RateLimiter, RequestStats
from src.python.ibkr_connector.exceptions import RateLimitError
from src.python.config.settings import RateLimitConfig


class TestRateLimiter:
    """Test cases for RateLimiter."""
    
    @pytest.fixture
    def config(self):
        """Create test rate limit config."""
        return RateLimitConfig(
            max_requests_per_second=10.0,  # Lower rate for testing
            burst_size=5,
            throttle_wait=0.01
        )
    
    @pytest.fixture
    def rate_limiter(self, config):
        """Create rate limiter with test config."""
        return RateLimiter(config)
    
    def test_initialization(self, rate_limiter, config):
        """Test rate limiter initialization."""
        assert rate_limiter.config == config
        assert rate_limiter._tokens == 10.0
        assert rate_limiter._max_tokens == 10.0
        assert rate_limiter._refill_rate == 10.0
        assert isinstance(rate_limiter.stats, RequestStats)
    
    @pytest.mark.asyncio
    async def test_acquire_with_available_tokens(self, rate_limiter):
        """Test acquiring when tokens are available."""
        wait_time = await rate_limiter.acquire()
        
        assert wait_time == 0.0
        assert rate_limiter._tokens < 10.0  # Token was consumed
        assert rate_limiter.stats.total_requests == 1
        assert rate_limiter.stats.queued_requests == 0
    
    @pytest.mark.asyncio
    async def test_acquire_multiple_fast(self, rate_limiter):
        """Test acquiring multiple tokens quickly."""
        # Acquire 5 tokens rapidly
        wait_times = []
        for _ in range(5):
            wait_time = await rate_limiter.acquire()
            wait_times.append(wait_time)
        
        # First few should be immediate (burst capacity)
        assert wait_times[0] == 0.0
        assert rate_limiter.stats.total_requests == 5
    
    @pytest.mark.asyncio
    async def test_acquire_with_wait(self, rate_limiter):
        """Test acquiring when tokens need to be refilled."""
        # Consume all tokens
        rate_limiter._tokens = 0.5  # Less than 1 token
        
        start = time.monotonic()
        wait_time = await rate_limiter.acquire()
        elapsed = time.monotonic() - start
        
        # Should have waited for tokens
        assert wait_time > 0
        assert elapsed >= wait_time
        assert rate_limiter.stats.queued_requests == 1
    
    @pytest.mark.asyncio
    async def test_acquire_with_timeout(self, rate_limiter):
        """Test timeout when waiting for tokens."""
        # Consume all tokens
        rate_limiter._tokens = 0
        
        with pytest.raises(RateLimitError) as exc_info:
            await rate_limiter.acquire(timeout=0.01)  # Very short timeout
        
        assert "exceeds timeout" in str(exc_info.value)
        assert exc_info.value.retry_after > 0
        assert rate_limiter.stats.rejected_requests == 1
    
    @pytest.mark.asyncio
    async def test_token_refill(self, rate_limiter):
        """Test token refill mechanism."""
        # Set tokens to half
        rate_limiter._tokens = 5.0
        
        # Wait for refill
        await asyncio.sleep(0.1)  # 100ms should add ~1 token at 10/sec
        
        # Force refill calculation
        async with rate_limiter._lock:
            rate_limiter._refill_tokens()
        
        # Should have more tokens now
        assert rate_limiter._tokens > 5.0
        assert rate_limiter._tokens <= 10.0  # Capped at max
    
    @pytest.mark.asyncio
    async def test_check_rate(self, rate_limiter):
        """Test checking if request can be made."""
        # With tokens available
        can_proceed = await rate_limiter.check_rate()
        assert can_proceed is True
        
        # Without tokens
        rate_limiter._tokens = 0.5
        can_proceed = await rate_limiter.check_rate()
        assert can_proceed is False
    
    @pytest.mark.asyncio
    async def test_concurrent_requests(self, rate_limiter):
        """Test handling concurrent requests."""
        # Create multiple concurrent requests
        async def make_request(index):
            wait_time = await rate_limiter.acquire()
            return index, wait_time
        
        # Launch 20 concurrent requests (more than burst capacity)
        tasks = [make_request(i) for i in range(20)]
        results = await asyncio.gather(*tasks)
        
        # Verify all completed
        assert len(results) == 20
        assert rate_limiter.stats.total_requests == 20
        
        # Some should have waited
        wait_times = [r[1] for r in results]
        assert any(w > 0 for w in wait_times)
    
    def test_get_stats(self, rate_limiter):
        """Test statistics retrieval."""
        stats = rate_limiter.get_stats()
        
        assert stats['total_requests'] == 0
        assert stats['rejected_requests'] == 0
        assert stats['queued_requests'] == 0
        assert 'current_rate' in stats
        assert 'tokens_available' in stats
        assert stats['max_rate'] == '10.0 req/sec'
    
    @pytest.mark.asyncio
    async def test_reset_stats(self, rate_limiter):
        """Test resetting statistics."""
        # Generate some stats
        await rate_limiter.acquire()
        await rate_limiter.acquire()
        
        assert rate_limiter.stats.total_requests == 2
        
        # Reset
        rate_limiter.reset_stats()
        
        assert rate_limiter.stats.total_requests == 0
        assert rate_limiter.stats.rejected_requests == 0
        assert len(rate_limiter._request_times) == 0
    
    @pytest.mark.asyncio
    async def test_current_rate_calculation(self, rate_limiter):
        """Test current request rate calculation."""
        # Make several requests
        for _ in range(5):
            await rate_limiter.acquire()
            await asyncio.sleep(0.1)  # Space them out
        
        stats = rate_limiter.get_stats()
        # Current rate should reflect recent requests
        assert rate_limiter.stats.current_rate >= 0
    
    def test_rate_limiter_repr(self, rate_limiter):
        """Test string representation."""
        repr_str = repr(rate_limiter)
        assert "RateLimiter" in repr_str
        assert "10.0 req/sec" in repr_str
        assert "tokens=" in repr_str