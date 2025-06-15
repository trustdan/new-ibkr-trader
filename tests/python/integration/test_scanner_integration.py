"""
Integration tests for Python-Go scanner communication

Tests the complete integration between the Python IBKR service and Go scanner.
"""

import asyncio
import pytest
import httpx
from datetime import datetime
from unittest.mock import Mock, AsyncMock, patch

from src.python.scanner_client import (
    ScannerClient, ScanRequest, ScanFilter, FilterType,
    OptionContract, VerticalSpread
)
from src.python.integration.scanner_coordinator import (
    ScannerCoordinator, ScanJob
)
from src.python.integration.backpressure import (
    BackpressureHandler, BackpressureStrategy
)


class TestScannerClient:
    """Test the scanner HTTP client"""
    
    @pytest.mark.asyncio
    async def test_health_check(self):
        """Test health check endpoint"""
        async with ScannerClient(base_url="http://localhost:8080") as client:
            # Mock the HTTP response
            with patch.object(client._client, 'get') as mock_get:
                mock_response = Mock()
                mock_response.json.return_value = {"status": "healthy", "version": "1.0"}
                mock_response.raise_for_status = Mock()
                mock_get.return_value = mock_response
                
                health = await client.health_check()
                
                assert health["status"] == "healthy"
                mock_get.assert_called_once_with("/health")
                
    @pytest.mark.asyncio
    async def test_scan_request(self):
        """Test basic scan request"""
        async with ScannerClient() as client:
            # Create scan request
            request = ScanRequest(
                symbol="AAPL",
                filters=[
                    ScanFilter(
                        type=FilterType.DELTA,
                        params={"min": 0.25, "max": 0.35}
                    )
                ],
                limit=10
            )
            
            # Mock response data
            mock_spread_data = {
                "long_leg": {
                    "symbol": "AAPL",
                    "expiry": "2025-02-21",
                    "strike": 150.0,
                    "right": "C",
                    "delta": 0.30,
                    "theta": -0.05,
                    "vega": 0.15,
                    "iv": 0.25,
                    "volume": 1000,
                    "open_interest": 5000,
                    "bid": 5.50,
                    "ask": 5.60,
                    "last": 5.55
                },
                "short_leg": {
                    "symbol": "AAPL",
                    "expiry": "2025-02-21",
                    "strike": 155.0,
                    "right": "C",
                    "delta": 0.20,
                    "theta": -0.04,
                    "vega": 0.12,
                    "iv": 0.24,
                    "volume": 800,
                    "open_interest": 4000,
                    "bid": 3.20,
                    "ask": 3.30,
                    "last": 3.25
                },
                "net_debit": 2.30,
                "max_profit": 2.70,
                "max_loss": 2.30,
                "breakeven": 152.30,
                "probability_profit": 0.65,
                "score": 85.5
            }
            
            # Mock the HTTP response
            with patch.object(client._client, 'post') as mock_post:
                mock_response = Mock()
                mock_response.json.return_value = {"spreads": [mock_spread_data]}
                mock_response.raise_for_status = Mock()
                mock_post.return_value = mock_response
                
                spreads = await client.scan(request)
                
                assert len(spreads) == 1
                spread = spreads[0]
                assert spread.score == 85.5
                assert spread.probability_profit == 0.65
                assert spread.long_leg.strike == 150.0
                assert spread.short_leg.strike == 155.0
                
    @pytest.mark.asyncio
    async def test_scan_retry_on_rate_limit(self):
        """Test retry logic on rate limit"""
        async with ScannerClient(max_retries=2) as client:
            request = ScanRequest(symbol="SPY", filters=[], limit=10)
            
            # Mock rate limit response then success
            call_count = 0
            
            async def mock_post(*args, **kwargs):
                nonlocal call_count
                call_count += 1
                
                if call_count == 1:
                    # First call: rate limited
                    response = Mock()
                    response.status_code = 429
                    raise httpx.HTTPStatusError(
                        "Rate limited",
                        request=Mock(),
                        response=response
                    )
                else:
                    # Second call: success
                    response = Mock()
                    response.json.return_value = {"spreads": []}
                    response.raise_for_status = Mock()
                    return response
                    
            with patch.object(client._client, 'post', mock_post):
                spreads = await client.scan(request)
                
                assert call_count == 2
                assert spreads == []


class TestBackpressureHandler:
    """Test backpressure handling"""
    
    @pytest.mark.asyncio
    async def test_token_bucket_rate_limiting(self):
        """Test token bucket rate limiting"""
        handler = BackpressureHandler(
            strategy=BackpressureStrategy.TOKEN_BUCKET,
            requests_per_second=5.0,
            burst_size=10
        )
        
        # Should allow burst
        results = []
        for i in range(10):
            result = await handler.acquire()
            results.append(result)
            
        assert all(results)  # All should succeed
        
        # Next should fail (tokens exhausted)
        assert not await handler.acquire()
        
    @pytest.mark.asyncio
    async def test_circuit_breaker(self):
        """Test circuit breaker functionality"""
        handler = BackpressureHandler(
            circuit_breaker_threshold=3,
            circuit_breaker_timeout=1
        )
        
        # Record failures
        for i in range(3):
            handler.record_request(0.1, success=False)
            
        # Circuit should be open
        assert not await handler.acquire()
        
        # Wait for timeout
        await asyncio.sleep(1.1)
        
        # Circuit should be closed
        assert await handler.acquire()
        
    @pytest.mark.asyncio
    async def test_adaptive_rate_limiting(self):
        """Test adaptive rate limiting"""
        handler = BackpressureHandler(
            strategy=BackpressureStrategy.ADAPTIVE,
            requests_per_second=10.0
        )
        
        # Record slow responses
        for i in range(5):
            handler.record_request(2.0, success=True)  # 2 second responses
            
        # Rate should decrease
        assert handler.adaptive_rate < 10.0
        
        # Record fast responses
        for i in range(10):
            handler.record_request(0.1, success=True)  # 0.1 second responses
            
        # Rate should increase
        assert handler.adaptive_rate > 5.0


class TestScannerCoordinator:
    """Test the scanner coordinator"""
    
    @pytest.mark.asyncio
    async def test_scan_with_cache(self):
        """Test scanning with cache"""
        # Mock dependencies
        mock_ibkr = Mock()
        mock_ibkr.emit_event = AsyncMock()
        
        mock_scanner = Mock()
        mock_scanner.scan = AsyncMock(return_value=[])
        
        coordinator = ScannerCoordinator(
            ibkr_connection=mock_ibkr,
            scanner_client=mock_scanner,
            scan_cache_ttl=60
        )
        
        await coordinator.start()
        
        try:
            # First scan
            filters = [
                ScanFilter(FilterType.DELTA, {"min": 0.3, "max": 0.4})
            ]
            
            result1 = await coordinator.scan_symbol("TSLA", filters)
            assert mock_scanner.scan.call_count == 1
            
            # Second scan (should hit cache)
            result2 = await coordinator.scan_symbol("TSLA", filters)
            assert mock_scanner.scan.call_count == 1  # No new call
            assert coordinator.metrics["cache_hits"] == 1
            
        finally:
            await coordinator.stop()
            
    @pytest.mark.asyncio
    async def test_concurrent_scan_limiting(self):
        """Test concurrent scan limiting"""
        mock_ibkr = Mock()
        mock_ibkr.emit_event = AsyncMock()
        
        # Mock scanner with delay
        async def slow_scan(*args, **kwargs):
            await asyncio.sleep(0.1)
            return []
            
        mock_scanner = Mock()
        mock_scanner.scan = slow_scan
        
        coordinator = ScannerCoordinator(
            ibkr_connection=mock_ibkr,
            scanner_client=mock_scanner,
            max_concurrent_scans=2
        )
        
        await coordinator.start()
        
        try:
            # Submit multiple scans
            tasks = []
            for i in range(5):
                task = asyncio.create_task(
                    coordinator.scan_symbol(f"STOCK{i}", [])
                )
                tasks.append(task)
                
            # Should process with max 2 concurrent
            await asyncio.gather(*tasks)
            
            # Check metrics
            metrics = coordinator.get_metrics()
            assert metrics["successful_scans"] == 5
            
        finally:
            await coordinator.stop()
            
    @pytest.mark.asyncio
    async def test_scan_job_lifecycle(self):
        """Test scan job lifecycle tracking"""
        mock_ibkr = Mock()
        mock_ibkr.emit_event = AsyncMock()
        
        mock_scanner = Mock()
        mock_scanner.scan = AsyncMock(return_value=[])
        
        coordinator = ScannerCoordinator(
            ibkr_connection=mock_ibkr,
            scanner_client=mock_scanner
        )
        
        await coordinator.start()
        
        try:
            # Track job states
            job_states = []
            
            # Patch to track job state changes
            original_process = coordinator._process_scan_job
            
            async def track_process(job):
                job_states.append(job.status)
                await original_process(job)
                job_states.append(job.status)
                
            coordinator._process_scan_job = track_process
            
            # Run scan
            await coordinator.scan_symbol("AMD", [])
            
            # Should have gone through states
            assert "pending" in job_states
            assert "processing" in job_states
            assert "completed" in job_states
            
        finally:
            await coordinator.stop()


class TestIntegration:
    """End-to-end integration tests"""
    
    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_full_scan_flow(self):
        """Test complete scan flow from request to response"""
        # This test requires the Go scanner to be running
        # Skip if not in integration test mode
        
        try:
            # Check if scanner is available
            async with httpx.AsyncClient() as client:
                response = await client.get("http://localhost:8080/health")
                if response.status_code != 200:
                    pytest.skip("Scanner service not available")
        except:
            pytest.skip("Scanner service not available")
            
        # Run actual integration test
        mock_ibkr = Mock()
        mock_ibkr.emit_event = AsyncMock()
        
        async with ScannerClient() as scanner_client:
            coordinator = ScannerCoordinator(
                ibkr_connection=mock_ibkr,
                scanner_client=scanner_client
            )
            
            await coordinator.start()
            
            try:
                # Create realistic filters
                filters = [
                    ScanFilter(
                        type=FilterType.DELTA,
                        params={"min": 0.2, "max": 0.4}
                    ),
                    ScanFilter(
                        type=FilterType.DTE,
                        params={"min": 30, "max": 90}
                    ),
                    ScanFilter(
                        type=FilterType.LIQUIDITY,
                        params={
                            "min_volume": 100,
                            "min_open_interest": 500,
                            "max_bid_ask_spread": 0.15
                        }
                    )
                ]
                
                # Run scan
                spreads = await coordinator.scan_symbol("SPY", filters)
                
                # Verify results
                assert isinstance(spreads, list)
                
                if spreads:
                    spread = spreads[0]
                    assert hasattr(spread, 'score')
                    assert hasattr(spread, 'probability_profit')
                    assert spread.long_leg.delta >= 0.2
                    assert spread.long_leg.delta <= 0.4
                    
                # Check metrics
                metrics = coordinator.get_metrics()
                assert metrics["total_scans"] > 0
                assert "backpressure" in metrics
                
            finally:
                await coordinator.stop()


# Fixtures
@pytest.fixture
def mock_scanner_response():
    """Fixture for mock scanner response"""
    return {
        "spreads": [
            {
                "long_leg": {
                    "symbol": "TEST",
                    "expiry": "2025-03-21",
                    "strike": 100.0,
                    "right": "C",
                    "delta": 0.35,
                    "theta": -0.08,
                    "vega": 0.20,
                    "iv": 0.30,
                    "volume": 2000,
                    "open_interest": 10000,
                    "bid": 8.50,
                    "ask": 8.60,
                    "last": 8.55
                },
                "short_leg": {
                    "symbol": "TEST",
                    "expiry": "2025-03-21",
                    "strike": 105.0,
                    "right": "C",
                    "delta": 0.25,
                    "theta": -0.06,
                    "vega": 0.18,
                    "iv": 0.28,
                    "volume": 1500,
                    "open_interest": 8000,
                    "bid": 5.20,
                    "ask": 5.30,
                    "last": 5.25
                },
                "net_debit": 3.30,
                "max_profit": 1.70,
                "max_loss": 3.30,
                "breakeven": 103.30,
                "probability_profit": 0.58,
                "score": 75.0
            }
        ]
    }


if __name__ == "__main__":
    # Run tests
    pytest.main([__file__, "-v"])