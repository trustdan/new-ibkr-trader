"""
Unit tests for connection manager.

These tests use mocks to test connection logic without requiring
an actual TWS connection. Integration tests on Windows will verify
actual TWS connectivity.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from datetime import datetime

from src.python.ibkr_connector.connection import ConnectionManager, ConnectionState
from src.python.ibkr_connector.exceptions import ConnectionError, ConfigurationError
from src.python.config.settings import Config, ConnectionConfig


class TestConnectionManager:
    """Test cases for ConnectionManager."""
    
    @pytest.fixture
    def config(self):
        """Create test configuration."""
        config = Config()
        config.connection.host = "localhost"
        config.connection.port = 7497
        config.connection.client_id = 999
        config.connection.timeout = 5.0
        return config
    
    @pytest.fixture
    def manager(self, config):
        """Create connection manager with test config."""
        return ConnectionManager(config)
    
    def test_initialization(self, manager, config):
        """Test connection manager initialization."""
        assert manager.config == config
        assert manager.state == ConnectionState.DISCONNECTED
        assert not manager.is_connected()
        assert manager._connected_time is None
        assert manager._reconnect_count == 0
    
    def test_invalid_config(self):
        """Test initialization with invalid config."""
        config = Config()
        config.connection.port = 9999  # Invalid port
        
        with pytest.raises(ValueError):
            ConnectionManager(config)
    
    @pytest.mark.asyncio
    async def test_connect_success(self, manager):
        """Test successful connection."""
        # Mock IB.connectAsync
        with patch.object(manager.ib, 'connectAsync', new_callable=AsyncMock) as mock_connect:
            mock_connect.return_value = None
            
            # Mock isConnected to return True after connection
            with patch.object(manager.ib, 'isConnected', return_value=True):
                # Track emitted events
                events = []
                async def capture_event(event_data):
                    events.append(event_data)
                
                manager.event_manager.on('connection_established', capture_event)
                
                # Connect
                await manager.connect()
                
                # Verify connection was attempted
                mock_connect.assert_called_once_with(
                    host="localhost",
                    port=7497,
                    clientId=999,
                    timeout=5.0
                )
                
                # Verify state changes
                assert manager.state == ConnectionState.CONNECTED
                assert manager.is_connected()
                assert manager._connected_time is not None
                assert manager._reconnect_count == 0
                
                # Verify event was emitted
                assert len(events) == 1
                assert events[0]['host'] == "localhost"
                assert events[0]['port'] == 7497
                assert events[0]['client_id'] == 999
    
    @pytest.mark.asyncio
    async def test_connect_already_connected(self, manager):
        """Test connecting when already connected."""
        manager.state = ConnectionState.CONNECTED
        
        with patch.object(manager.ib, 'connectAsync', new_callable=AsyncMock) as mock_connect:
            await manager.connect()
            
            # Should not attempt to connect again
            mock_connect.assert_not_called()
    
    @pytest.mark.asyncio
    async def test_connect_timeout(self, manager):
        """Test connection timeout."""
        with patch.object(manager.ib, 'connectAsync', new_callable=AsyncMock) as mock_connect:
            mock_connect.side_effect = asyncio.TimeoutError()
            
            with pytest.raises(ConnectionError) as exc_info:
                await manager.connect()
            
            assert "Connection timeout" in str(exc_info.value)
            assert manager.state == ConnectionState.ERROR
    
    @pytest.mark.asyncio
    async def test_connect_failure(self, manager):
        """Test connection failure."""
        with patch.object(manager.ib, 'connectAsync', new_callable=AsyncMock) as mock_connect:
            mock_connect.side_effect = Exception("Network error")
            
            with pytest.raises(ConnectionError) as exc_info:
                await manager.connect()
            
            assert "Failed to connect" in str(exc_info.value)
            assert manager.state == ConnectionState.ERROR
    
    @pytest.mark.asyncio
    async def test_disconnect(self, manager):
        """Test disconnection."""
        # Set up as connected
        manager.state = ConnectionState.CONNECTED
        manager._connected_time = datetime.now()
        
        # Track events
        events = []
        async def capture_event(event_data):
            events.append(event_data)
        
        manager.event_manager.on('connection_closed', capture_event)
        
        with patch.object(manager.ib, 'disconnect') as mock_disconnect:
            await manager.disconnect()
            
            mock_disconnect.assert_called_once()
            assert manager.state == ConnectionState.DISCONNECTED
            assert manager._connected_time is None
            
            # Verify event
            assert len(events) == 1
            assert 'timestamp' in events[0]
    
    @pytest.mark.asyncio
    async def test_ensure_connected_when_disconnected(self, manager):
        """Test ensure_connected reconnects when disconnected."""
        manager.state = ConnectionState.DISCONNECTED
        
        with patch.object(manager, 'connect', new_callable=AsyncMock) as mock_connect:
            await manager.ensure_connected()
            mock_connect.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_ensure_connected_when_connected(self, manager):
        """Test ensure_connected does nothing when already connected."""
        manager.state = ConnectionState.CONNECTED
        
        with patch.object(manager, 'connect', new_callable=AsyncMock) as mock_connect:
            await manager.ensure_connected()
            mock_connect.assert_not_called()
    
    def test_connection_info(self, manager):
        """Test connection info property."""
        info = manager.connection_info
        
        assert info['state'] == 'disconnected'
        assert info['connected'] is False
        assert info['connected_time'] is None
        assert info['reconnect_count'] == 0
        assert info['client_id'] is None
        assert info['server_version'] is None
    
    @pytest.mark.asyncio
    async def test_error_event_handler(self, manager):
        """Test error event handling."""
        events = []
        async def capture_event(event_data):
            events.append(event_data)
        
        manager.event_manager.on('error_occurred', capture_event)
        
        # Simulate error event
        manager._on_error(reqId=123, errorCode=100, errorString="Pacing violation")
        
        # Wait for async event processing
        await asyncio.sleep(0.1)
        
        assert len(events) == 1
        assert events[0]['req_id'] == 123
        assert events[0]['error_code'] == 100
        assert events[0]['error_string'] == "Pacing violation"
    
    def test_disconnected_event_handler(self, manager):
        """Test disconnected event handling."""
        manager.state = ConnectionState.CONNECTED
        manager._connected_time = datetime.now()
        
        manager._on_disconnected()
        
        assert manager.state == ConnectionState.DISCONNECTED


# TODO: Windows Integration Tests
# The following tests require actual TWS connection and should be run on Windows:
# - test_real_connection_to_tws
# - test_real_disconnection_handling  
# - test_real_error_handling
# - test_real_reconnection_after_failure
# - test_daily_restart_handling
# - test_rate_limit_with_real_requests