"""
Integration tests for ConnectionManager against real TWS.

These tests validate our connection components work with actual TWS instances.
Run only on Windows with TWS running.

Usage:
    pytest tests/python/integration/test_connection_integration.py -m integration
"""

import asyncio
import pytest
import logging
from datetime import datetime, timedelta

from src.python.ibkr_connector.connection import ConnectionManager, ConnectionState
from src.python.config.settings import Config


# Configure logging for integration tests
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@pytest.mark.integration
@pytest.mark.asyncio
class TestConnectionIntegration:
    """Test real TWS connection scenarios."""
    
    @pytest.fixture
    def config(self):
        """Test configuration for paper trading."""
        config = Config.from_env()
        # Override for paper trading
        config.connection.host = "127.0.0.1"
        config.connection.port = 7497
        config.connection.client_id = 998  # Test client ID
        config.connection.timeout = 10.0
        return config
    
    @pytest.fixture
    async def connection_manager(self, config):
        """Provide a connection manager instance."""
        manager = ConnectionManager(config)
        yield manager
        # Cleanup
        if manager.is_connected():
            await manager.disconnect()
    
    async def test_basic_connection_establishment(self, connection_manager):
        """Test basic connection to TWS."""
        logger.info("🔌 Testing basic TWS connection...")
        
        # Initially disconnected
        assert connection_manager.state == ConnectionState.DISCONNECTED
        assert not connection_manager.is_connected()
        
        # Attempt connection
        await connection_manager.connect()
        
        # Verify connection
        assert connection_manager.state == ConnectionState.CONNECTED
        assert connection_manager.is_connected()
        
        # Check connection info
        info = connection_manager.connection_info
        assert info['connected'] is True
        assert info['connected_time'] is not None
        assert info['client_id'] == 998
        assert info['server_version'] is not None
        
        logger.info("✅ Basic connection test passed!")
    
    async def test_connection_with_invalid_port(self, config):
        """Test connection failure with invalid port."""
        logger.info("🚫 Testing connection with invalid port...")
        
        # Use invalid port
        config.connection.port = 9999
        config.connection.timeout = 2.0  # Quick timeout
        
        manager = ConnectionManager(config)
        
        with pytest.raises(Exception):  # Should raise ConnectionError
            await manager.connect()
        
        assert manager.state == ConnectionState.ERROR
        logger.info("✅ Invalid port test passed!")
    
    async def test_connection_info_details(self, connection_manager):
        """Test detailed connection information."""
        logger.info("📊 Testing connection info details...")
        
        start_time = datetime.now()
        await connection_manager.connect()
        
        info = connection_manager.connection_info
        
        # Validate connection details
        assert info['state'] == 'connected'
        assert info['connected'] is True
        assert info['reconnect_count'] == 0
        assert info['client_id'] == 998
        assert isinstance(info['server_version'], str)
        
        # Check timing
        connected_time = info['connected_time']
        assert isinstance(connected_time, datetime)
        assert connected_time >= start_time
        assert connected_time <= datetime.now()
        
        logger.info("✅ Connection info test passed!")
    
    async def test_disconnect_functionality(self, connection_manager):
        """Test proper disconnection."""
        logger.info("👋 Testing disconnect functionality...")
        
        # Connect first
        await connection_manager.connect()
        assert connection_manager.is_connected()
        
        # Disconnect
        await connection_manager.disconnect()
        
        # Verify disconnection
        assert connection_manager.state == ConnectionState.DISCONNECTED
        assert not connection_manager.is_connected()
        
        info = connection_manager.connection_info
        assert info['connected'] is False
        assert info['connected_time'] is None
        
        logger.info("✅ Disconnect test passed!")
    
    async def test_ensure_connected_functionality(self, connection_manager):
        """Test ensure_connected method."""
        logger.info("🔄 Testing ensure_connected functionality...")
        
        # Should connect when disconnected
        assert not connection_manager.is_connected()
        await connection_manager.ensure_connected()
        assert connection_manager.is_connected()
        
        # Should be no-op when already connected
        await connection_manager.ensure_connected()
        assert connection_manager.is_connected()
        
        logger.info("✅ Ensure connected test passed!")
    
    async def test_multiple_connection_attempts(self, connection_manager):
        """Test multiple connection attempts don't cause issues."""
        logger.info("🔁 Testing multiple connection attempts...")
        
        # Connect
        await connection_manager.connect()
        assert connection_manager.is_connected()
        
        # Try connecting again (should be no-op with warning)
        await connection_manager.connect()
        assert connection_manager.is_connected()
        
        logger.info("✅ Multiple connection test passed!")


@pytest.mark.integration
@pytest.mark.asyncio
class TestTWSRequirements:
    """Test TWS-specific requirements and behaviors."""
    
    async def test_tws_server_time(self):
        """Test TWS server time request."""
        logger.info("⏰ Testing TWS server time...")
        
        config = Config.from_env()
        config.connection.client_id = 997
        
        manager = ConnectionManager(config)
        
        try:
            await manager.connect()
            
            # Test direct IB access for server time
            server_time = manager.ib.reqCurrentTime()
            assert isinstance(server_time, datetime)
            
            # Should be reasonably current
            now = datetime.now()
            time_diff = abs((server_time - now).total_seconds())
            assert time_diff < 300  # Within 5 minutes
            
            logger.info(f"✅ Server time test passed! Server time: {server_time}")
            
        finally:
            if manager.is_connected():
                await manager.disconnect()
    
    async def test_account_information(self):
        """Test account information retrieval."""
        logger.info("📊 Testing account information...")
        
        config = Config.from_env()
        config.connection.client_id = 996
        
        manager = ConnectionManager(config)
        
        try:
            await manager.connect()
            
            # Test account info
            accounts = manager.ib.managedAccounts()
            assert isinstance(accounts, list)
            assert len(accounts) > 0
            
            logger.info(f"✅ Account test passed! Accounts: {accounts}")
            
        finally:
            if manager.is_connected():
                await manager.disconnect()


# Utility function for manual testing
async def manual_connection_test():
    """Manual connection test - useful for debugging."""
    print("🧪 Manual Connection Test")
    print("=" * 30)
    
    config = Config.from_env()
    config.connection.client_id = 999
    
    manager = ConnectionManager(config)
    
    try:
        print("Connecting...")
        await manager.connect()
        
        print("Connection successful!")
        print(f"Info: {manager.connection_info}")
        
        # Keep alive for a few seconds
        await asyncio.sleep(2)
        
    except Exception as e:
        print(f"Connection failed: {e}")
    
    finally:
        if manager.is_connected():
            await manager.disconnect()
            print("Disconnected.")


if __name__ == "__main__":
    # Run manual test
    asyncio.run(manual_connection_test()) 