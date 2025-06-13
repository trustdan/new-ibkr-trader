"""
Async IBKR Connection Manager
Handles all TWS API communication with event-driven patterns
"""
import asyncio
import os
from typing import Optional, Dict, Any
from datetime import datetime

from ib_insync import IB, Contract, Trade, util, MarketOrder, LimitOrder
from ib_insync.ibcontroller import Watchdog, IBC
import logging

from ..monitoring.metrics import connection_status, api_requests, api_errors

logger = logging.getLogger(__name__)


class AsyncIBKRService:
    """Event-driven IBKR service using ib-insync patterns"""
    
    def __init__(self):
        self.ib = IB()
        self.watchdog: Optional[Watchdog] = None
        self.connected = False
        self.subscriptions: Dict[int, Contract] = {}  # reqId -> Contract
        self.max_subscriptions = int(os.getenv("MAX_SUBSCRIPTIONS", "90"))
        
        # Connection parameters
        self.host = os.getenv("TWS_HOST", "host.docker.internal")
        self.port = int(os.getenv("TWS_PORT", "7497"))
        self.client_id = int(os.getenv("TWS_CLIENT_ID", "1"))
        self.account_type = os.getenv("ACCOUNT_TYPE", "paper")
        
        # Setup event handlers
        self._setup_event_handlers()
        
    def _setup_event_handlers(self):
        """Configure all event handlers for async operation"""
        # Connection events
        self.ib.connectedEvent += self._on_connected
        self.ib.disconnectedEvent += self._on_disconnected
        
        # Order events  
        self.ib.orderStatusEvent += self._on_order_status
        self.ib.execDetailsEvent += self._on_exec_details
        
        # Market data events
        self.ib.pendingTickersEvent += self._on_pending_tickers
        
        # Error handling
        self.ib.errorEvent += self._on_error
        
    async def connect(self, retry_count: int = 3) -> bool:
        """Establish connection to TWS with retry logic"""
        for attempt in range(retry_count):
            try:
                logger.info(f"🔌 Connecting to TWS at {self.host}:{self.port} (attempt {attempt + 1}/{retry_count})")
                
                await self.ib.connectAsync(
                    host=self.host,
                    port=self.port,
                    clientId=self.client_id,
                    readonly=False
                )
                
                # Setup watchdog for auto-reconnection
                if os.getenv("WATCHDOG_ENABLED", "true").lower() == "true":
                    self._setup_watchdog()
                
                return True
                
            except Exception as e:
                logger.error(f"Connection attempt {attempt + 1} failed: {e}")
                if attempt < retry_count - 1:
                    await asyncio.sleep(2 ** attempt)  # Exponential backoff
                else:
                    raise
                    
        return False
        
    def _setup_watchdog(self):
        """Configure Watchdog for automatic TWS restart and reconnection"""
        watchdog_config = {
            'host': self.host,
            'port': self.port,
            'clientId': self.client_id,
            'connectTimeout': int(os.getenv("WATCHDOG_TIMEOUT", "60")),
        }
        
        # Only setup IBC if TWS path is provided
        tws_path = os.getenv("TWS_PATH")
        if tws_path:
            ibc = IBC(
                twsPath=tws_path,
                gateway=False,
                tradingMode=self.account_type
            )
            self.watchdog = Watchdog(ibc, self.ib, **watchdog_config)
        else:
            # Watchdog without IBC for reconnection only
            self.watchdog = Watchdog(None, self.ib, **watchdog_config)
            
        self.watchdog.start()
        logger.info("🐕 Watchdog started for connection monitoring")
        
    async def disconnect(self):
        """Gracefully disconnect from TWS"""
        if self.watchdog:
            self.watchdog.stop()
            
        if self.ib.isConnected():
            self.ib.disconnect()
            
        logger.info("👋 Disconnected from TWS")
        
    # Event Handlers
    def _on_connected(self):
        """Handle successful connection"""
        self.connected = True
        connection_status.labels(status="connected").set(1)
        connection_status.labels(status="disconnected").set(0)
        logger.info("✅ Connected to TWS successfully")
        
    def _on_disconnected(self):
        """Handle disconnection"""
        self.connected = False
        connection_status.labels(status="connected").set(0)
        connection_status.labels(status="disconnected").set(1)
        logger.warning("❌ Disconnected from TWS")
        
    def _on_order_status(self, trade: Trade):
        """Handle order status updates"""
        logger.info(f"📊 Order status: {trade.order.orderId} - {trade.orderStatus.status}")
        
    def _on_exec_details(self, trade: Trade, fill):
        """Handle execution details"""
        logger.info(f"✅ Execution: {fill.contract.symbol} - {fill.execution.shares} @ {fill.execution.price}")
        
    def _on_pending_tickers(self, tickers):
        """Handle market data updates"""
        # Process tickers without blocking
        for ticker in tickers:
            if ticker.last is not None:
                logger.debug(f"📈 {ticker.contract.symbol}: {ticker.last}")
                
    def _on_error(self, reqId: int, errorCode: int, errorString: str, contract: Contract):
        """Handle API errors with proper categorization"""
        api_errors.labels(error_code=str(errorCode)).inc()
        
        if errorCode == 1100:
            logger.warning("🔌 Connectivity lost - Watchdog should reconnect")
        elif errorCode == 100:
            logger.error("⚡ Pacing violation - slow down requests!")
        elif errorCode == 502:
            logger.error("❌ TWS not connected")
        elif errorCode == 507:
            logger.error("🔌 Socket error - bad message")
        else:
            logger.error(f"⚠️ Error {errorCode}: {errorString}")
            
    # Public API Methods
    async def get_account_summary(self) -> Dict[str, Any]:
        """Get account summary information"""
        if not self.connected:
            raise ConnectionError("Not connected to TWS")
            
        api_requests.labels(endpoint="account_summary").inc()
        
        summary = self.ib.accountSummary()
        return {
            item.tag: item.value 
            for item in summary
        }
        
    async def test_connection(self) -> Dict[str, Any]:
        """Test connection and return status"""
        return {
            "connected": self.connected,
            "host": self.host,
            "port": self.port,
            "client_id": self.client_id,
            "account_type": self.account_type,
            "server_time": self.ib.reqCurrentTime() if self.connected else None
        }