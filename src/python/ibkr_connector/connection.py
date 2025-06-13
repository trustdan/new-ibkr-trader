"""
Core connection manager for IBKR TWS API.

This module provides the main connection interface to Interactive Brokers,
handling connection lifecycle, state management, and event coordination.
"""

import asyncio
import logging
from typing import Optional, Dict, Any, Callable
from datetime import datetime
from enum import Enum

from ib_insync import IB, util
from ib_insync.contract import Contract
from ib_insync.order import Order

from ..config.settings import Config, ConnectionConfig
from .exceptions import ConnectionError, AuthenticationError, ConfigurationError
from .events import EventManager


class ConnectionState(Enum):
    """Connection state enumeration."""
    DISCONNECTED = "disconnected"
    CONNECTING = "connecting"
    CONNECTED = "connected"
    RECONNECTING = "reconnecting"
    ERROR = "error"


class ConnectionManager:
    """
    Manages the connection to Interactive Brokers TWS/Gateway.
    
    This class provides a high-level interface for establishing and maintaining
    a connection to the IBKR API, handling events, and managing connection state.
    """
    
    def __init__(self, config: Optional[Config] = None):
        """
        Initialize the connection manager.
        
        Args:
            config: Configuration object. If None, uses default config.
        """
        self.config = config or Config.from_env()
        self.config.validate()
        
        self.ib = IB()
        self.event_manager = EventManager()
        self.state = ConnectionState.DISCONNECTED
        self.logger = logging.getLogger(__name__)
        
        self._connected_time: Optional[datetime] = None
        self._reconnect_count = 0
        self._setup_event_handlers()
    
    def _setup_event_handlers(self) -> None:
        """Set up IB event handlers."""
        # Connection events
        self.ib.connectedEvent += self._on_connected
        self.ib.disconnectedEvent += self._on_disconnected
        self.ib.errorEvent += self._on_error
        
        # TODO: Add more event handlers as needed
        # These will be tested on Windows with real TWS connection
    
    async def connect(
        self, 
        host: Optional[str] = None,
        port: Optional[int] = None,
        client_id: Optional[int] = None
    ) -> None:
        """
        Establish connection to TWS/Gateway.
        
        Args:
            host: Override config host
            port: Override config port
            client_id: Override config client_id
            
        Raises:
            ConnectionError: If connection fails
            ConfigurationError: If configuration is invalid
        """
        if self.state == ConnectionState.CONNECTED:
            self.logger.warning("Already connected")
            return
        
        # Use provided values or fall back to config
        conn_config = self.config.connection
        host = host or conn_config.host
        port = port or conn_config.port
        client_id = client_id or conn_config.client_id
        
        self.state = ConnectionState.CONNECTING
        self.logger.info(f"Connecting to {host}:{port} with client_id={client_id}")
        
        try:
            # TODO: Windows testing - verify connection parameters
            await self.ib.connectAsync(
                host=host,
                port=port,
                clientId=client_id,
                timeout=conn_config.timeout
            )
            
            # Connection successful
            self.state = ConnectionState.CONNECTED
            self._connected_time = datetime.now()
            self._reconnect_count = 0
            
            await self.event_manager.emit('connection_established', {
                'host': host,
                'port': port,
                'client_id': client_id,
                'timestamp': self._connected_time
            })
            
            self.logger.info("Successfully connected to IBKR")
            
        except asyncio.TimeoutError:
            self.state = ConnectionState.ERROR
            raise ConnectionError(f"Connection timeout after {conn_config.timeout}s")
        except Exception as e:
            self.state = ConnectionState.ERROR
            self.logger.error(f"Connection failed: {e}")
            raise ConnectionError(f"Failed to connect: {str(e)}")
    
    async def disconnect(self) -> None:
        """Disconnect from TWS/Gateway."""
        if self.state == ConnectionState.DISCONNECTED:
            return
        
        self.logger.info("Disconnecting from IBKR")
        self.ib.disconnect()
        self.state = ConnectionState.DISCONNECTED
        self._connected_time = None
        
        await self.event_manager.emit('connection_closed', {
            'timestamp': datetime.now()
        })
    
    async def ensure_connected(self) -> None:
        """Ensure connection is active, reconnect if necessary."""
        if self.state != ConnectionState.CONNECTED:
            await self.connect()
    
    def is_connected(self) -> bool:
        """Check if currently connected."""
        return self.state == ConnectionState.CONNECTED and self.ib.isConnected()
    
    @property
    def connection_info(self) -> Dict[str, Any]:
        """Get current connection information."""
        return {
            'state': self.state.value,
            'connected': self.is_connected(),
            'connected_time': self._connected_time,
            'reconnect_count': self._reconnect_count,
            'client_id': self.ib.client.clientId if self.ib.client else None,
            'server_version': self.ib.client.serverVersion() if self.ib.client else None
        }
    
    # Event handlers
    def _on_connected(self) -> None:
        """Handle connected event."""
        self.logger.debug("Connected event received")
    
    def _on_disconnected(self) -> None:
        """Handle disconnected event."""
        self.logger.warning("Disconnected event received")
        self.state = ConnectionState.DISCONNECTED
        asyncio.create_task(self.event_manager.emit('connection_lost', {
            'timestamp': datetime.now(),
            'was_connected_for': (
                datetime.now() - self._connected_time 
                if self._connected_time else None
            )
        }))
    
    def _on_error(self, reqId: int, errorCode: int, errorString: str, 
                  contract: Optional[Contract] = None) -> None:
        """
        Handle error events from TWS.
        
        Args:
            reqId: Request ID that caused the error
            errorCode: TWS error code
            errorString: Error description
            contract: Contract related to error (if any)
        """
        self.logger.error(f"TWS Error {errorCode}: {errorString} (reqId={reqId})")
        
        # TODO: Windows testing - verify error handling with real TWS
        asyncio.create_task(self.event_manager.emit('error_occurred', {
            'req_id': reqId,
            'error_code': errorCode,
            'error_string': errorString,
            'contract': contract,
            'timestamp': datetime.now()
        }))
    
    def __repr__(self) -> str:
        """String representation of connection manager."""
        return (
            f"ConnectionManager(state={self.state.value}, "
            f"connected={self.is_connected()})"
        )