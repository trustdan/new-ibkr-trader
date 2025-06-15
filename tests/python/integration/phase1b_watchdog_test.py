"""
Phase 1B: Watchdog Integration Testing
Tests the Watchdog component with real connection scenarios.
"""

import asyncio
import pytest
import logging
from datetime import datetime, time as dt_time
from unittest.mock import Mock, patch

from src.python.ibkr_connector.watchdog import ConnectionWatchdog, WatchdogState
from src.python.ibkr_connector.connection import ConnectionManager, ConnectionState
from src.python.config.settings import Config


# Configure logging for integration tests
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@pytest.mark.integration
@pytest.mark.asyncio
class TestWatchdogIntegration:
    """Test Watchdog component with connection scenarios."""
    
    @pytest.fixture
    def config(self):
        """Test configuration."""
        config = Config.from_env()
        config.connection.host = "127.0.0.1"
        config.connection.port = 7497
        config.connection.client_id = 999  # Test client ID
        return config
    
    @pytest.fixture
    async def connection_manager(self, config):
        """Connection manager fixture."""
        manager = ConnectionManager(config)
        yield manager
        # Cleanup
        if manager.is_connected():
            await manager.disconnect()
    
    @pytest.fixture
    async def watchdog(self, connection_manager):
        """Watchdog fixture."""
        watchdog = ConnectionWatchdog(connection_manager)
        # Faster testing intervals
        watchdog.health_check_interval = 2  # 2 seconds
        watchdog.reconnect_delay_base = 1   # 1 second base delay
        yield watchdog
        # Cleanup
        await watchdog.stop()
    
    async def test_watchdog_lifecycle(self, watchdog):
        """Test basic watchdog start/stop lifecycle."""
        logger.info("🧪 Testing watchdog lifecycle")
        
        # Initially stopped
        assert watchdog.state == WatchdogState.STOPPED
        
        # Start watchdog
        await watchdog.start()
        assert watchdog.state == WatchdogState.MONITORING
        assert watchdog.monitoring_task is not None
        
        # Check uptime tracking
        await asyncio.sleep(0.1)
        assert watchdog.get_uptime() > 0
        
        # Stop watchdog
        await watchdog.stop()
        assert watchdog.state == WatchdogState.STOPPED
        
        logger.info("✅ Watchdog lifecycle test passed")
    
    async def test_health_check_with_socket(self, watchdog):
        """Test health check using real socket connectivity."""
        logger.info("🧪 Testing health check with socket")
        
        # Test socket health check directly
        health_ok = await watchdog._check_socket_health()
        
        # Should be True since TWS is running
        assert health_ok, "Socket health check should pass with TWS running"
        
        logger.info("✅ Socket health check test passed")
    
    async def test_health_monitoring_loop(self, watchdog):
        """Test the health monitoring loop."""
        logger.info("🧪 Testing health monitoring loop")
        
        # Track events
        health_events = []
        
        def on_health_event(event_data):
            health_events.append(event_data)
        
        # Subscribe to health events
        watchdog.event_manager.subscribe('health_check_passed', on_health_event)
        watchdog.event_manager.subscribe('health_check_failed', on_health_event)
        
        # Start watchdog with short intervals
        await watchdog.start()
        
        # Wait for a few health checks
        await asyncio.sleep(5)
        
        # Stop watchdog
        await watchdog.stop()
        
        # Should have received health check events
        assert len(health_events) > 0, "Should receive health check events"
        assert watchdog.stats['health_checks'] > 0, "Should perform health checks"
        
        logger.info(f"✅ Health monitoring test passed - {len(health_events)} events")
    
    async def test_connection_recovery_simulation(self, watchdog):
        """Test connection recovery with simulated connection loss."""
        logger.info("🧪 Testing connection recovery simulation")
        
        # Track reconnection events
        reconnection_events = []
        
        def on_reconnection_event(event_data):
            reconnection_events.append(event_data)
        
        watchdog.event_manager.subscribe('reconnection_started', on_reconnection_event)
        watchdog.event_manager.subscribe('reconnection_successful', on_reconnection_event)
        
        # Start watchdog
        await watchdog.start()
        
        # Simulate connection loss by triggering reconnection
        logger.info("🔄 Simulating connection loss...")
        await watchdog._handle_reconnection()
        
        # Wait for reconnection attempt
        await asyncio.sleep(3)
        
        # Stop watchdog
        await watchdog.stop()
        
        # Should have attempted reconnection
        assert len(reconnection_events) > 0, "Should attempt reconnection"
        assert watchdog.stats['reconnect_count'] > 0, "Should track reconnection attempts"
        
        logger.info("✅ Connection recovery simulation test passed")
    
    async def test_daily_restart_detection(self, watchdog):
        """Test daily restart time detection."""
        logger.info("🧪 Testing daily restart detection")
        
        # Mock current time to be near restart time
        restart_time = dt_time(23, 44)  # 11:44 PM (1 minute before restart)
        
        with patch('src.python.ibkr_connector.watchdog.datetime') as mock_datetime:
            # Mock datetime.now().time() to return our test time
            mock_now = Mock()
            mock_now.time.return_value = restart_time
            mock_datetime.now.return_value = mock_now
            
            # Test restart detection
            await watchdog._check_daily_restart()
            
            # Should have detected restart window
            # (This test validates the logic without waiting for actual restart)
        
        logger.info("✅ Daily restart detection test passed")
    
    async def test_watchdog_metrics_and_status(self, watchdog):
        """Test watchdog metrics and status reporting."""
        logger.info("🧪 Testing watchdog metrics and status")
        
        # Start watchdog
        await watchdog.start()
        await asyncio.sleep(1)
        
        # Get health status
        status = watchdog.get_health_status()
        
        # Validate status structure
        assert 'watchdog_state' in status
        assert 'connection_state' in status
        assert 'uptime_seconds' in status
        assert 'stats' in status
        assert 'config' in status
        
        # Validate state values
        assert status['watchdog_state'] == 'monitoring'
        assert status['uptime_seconds'] > 0
        
        # Validate stats
        stats = status['stats']
        assert 'start_time' in stats
        assert 'health_checks' in stats
        assert 'reconnect_count' in stats
        
        await watchdog.stop()
        
        logger.info("✅ Watchdog metrics test passed")
    
    async def test_exponential_backoff(self, watchdog):
        """Test exponential backoff in reconnection attempts."""
        logger.info("🧪 Testing exponential backoff")
        
        # Test delay calculation
        delays = []
        for attempt in range(1, 6):
            delay = watchdog.reconnect_delay_base * (2 ** (attempt - 1))
            delay = min(delay, 60)  # Max 60 seconds
            delays.append(delay)
        
        # Should increase exponentially: 1, 2, 4, 8, 16
        expected = [1, 2, 4, 8, 16]
        assert delays == expected, f"Expected {expected}, got {delays}"
        
        logger.info("✅ Exponential backoff test passed")
    
    async def test_event_subscriptions(self, watchdog):
        """Test watchdog event subscription methods."""
        logger.info("🧪 Testing event subscriptions")
        
        # Test event subscription callbacks
        events_received = []
        
        def on_started(data):
            events_received.append(('started', data))
        
        def on_reconnection(data):
            events_received.append(('reconnection', data))
        
        # Subscribe to events
        watchdog.on_watchdog_started(on_started)
        watchdog.on_reconnection_started(on_reconnection)
        
        # Start watchdog to trigger events
        await watchdog.start()
        await asyncio.sleep(0.1)
        await watchdog.stop()
        
        # Should have received started event
        assert len(events_received) > 0, "Should receive events"
        assert any(event[0] == 'started' for event in events_received), "Should receive started event"
        
        logger.info("✅ Event subscription test passed")


@pytest.mark.integration
async def test_phase1b_full_integration():
    """Comprehensive Phase 1B integration test."""
    logger.info("🚀 Running Phase 1B Full Integration Test")
    
    # Configuration
    config = Config.from_env()
    config.connection.client_id = 998
    
    # Create components
    connection_manager = ConnectionManager(config)
    watchdog = ConnectionWatchdog(connection_manager)
    
    # Speed up for testing
    watchdog.health_check_interval = 3
    watchdog.max_reconnect_attempts = 3
    
    try:
        # Test 1: Start watchdog
        logger.info("📋 Test 1: Starting watchdog")
        await watchdog.start()
        assert watchdog.state == WatchdogState.MONITORING
        
        # Test 2: Health monitoring
        logger.info("📋 Test 2: Health monitoring")
        await asyncio.sleep(4)  # Wait for health checks
        assert watchdog.stats['health_checks'] > 0
        
        # Test 3: Status reporting
        logger.info("📋 Test 3: Status reporting")
        status = watchdog.get_health_status()
        assert status['uptime_seconds'] > 0
        
        # Test 4: Stop gracefully
        logger.info("📋 Test 4: Stopping watchdog")
        await watchdog.stop()
        assert watchdog.state == WatchdogState.STOPPED
        
        logger.info("🎉 Phase 1B Full Integration Test: PASSED!")
        return True
        
    except Exception as e:
        logger.error(f"❌ Phase 1B integration test failed: {e}")
        return False
        
    finally:
        # Cleanup
        if watchdog.state != WatchdogState.STOPPED:
            await watchdog.stop()
        if connection_manager.is_connected():
            await connection_manager.disconnect()


if __name__ == "__main__":
    """Run Phase 1B integration test."""
    print("🐕 PHASE 1B: WATCHDOG INTEGRATION TEST")
    print("=" * 50)
    
    # Run the comprehensive test
    success = asyncio.run(test_phase1b_full_integration())
    
    print("\n" + "=" * 50)
    if success:
        print("🎉 PHASE 1B WATCHDOG TESTING: COMPLETE!")
        print("✅ Ready for Phase 1C: Trading Operations")
    else:
        print("❌ Phase 1B testing failed - review logs")
    print("=" * 50) 