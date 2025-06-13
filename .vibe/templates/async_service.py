"""
Async Service Template - The foundation for all TWS-connected services

Remember: The One Rule - Never block the event loop!
"""
import asyncio
import logging
from typing import Optional
from ib_insync import IB, util
import signal

logger = logging.getLogger(__name__)


class AsyncService:
    """Base template for async IBKR services"""
    
    def __init__(self, name: str):
        self.name = name
        self.ib = IB()
        self.running = False
        self._setup_event_handlers()
        self._setup_shutdown_handlers()
        
    def _setup_event_handlers(self):
        """Wire up event handlers - the heart of our system"""
        # Connection events
        self.ib.connectedEvent += self._on_connected
        self.ib.disconnectedEvent += self._on_disconnected
        
        # Error handling
        self.ib.errorEvent += self._on_error
        
        # Add service-specific handlers in subclasses
        
    def _setup_shutdown_handlers(self):
        """Graceful shutdown on SIGINT/SIGTERM"""
        for sig in (signal.SIGINT, signal.SIGTERM):
            signal.signal(sig, lambda s, f: asyncio.create_task(self.shutdown()))
            
    async def start(self, host='localhost', port=7497, client_id=1):
        """Start the service with connection"""
        logger.info(f"Starting {self.name}...")
        self.running = True
        
        try:
            await self.ib.connectAsync(host, port, client_id)
            logger.info(f"{self.name} connected successfully")
            
            # Main service loop
            while self.running:
                await self.ib.sleep(1)  # The One Rule!
                await self._heartbeat()
                
        except Exception as e:
            logger.error(f"{self.name} error: {e}")
        finally:
            await self.shutdown()
            
    async def _heartbeat(self):
        """Override in subclasses for periodic tasks"""
        pass
        
    async def shutdown(self):
        """Graceful shutdown"""
        logger.info(f"Shutting down {self.name}...")
        self.running = False
        
        if self.ib.isConnected():
            # Clean up any subscriptions
            await self._cleanup()
            self.ib.disconnect()
            
        logger.info(f"{self.name} shutdown complete")
        
    async def _cleanup(self):
        """Override in subclasses for cleanup tasks"""
        pass
        
    # Event handlers
    def _on_connected(self):
        logger.info(f"{self.name}: Connected to TWS")
        
    def _on_disconnected(self):
        logger.warning(f"{self.name}: Disconnected from TWS")
        
    def _on_error(self, reqId, errorCode, errorString, contract):
        if errorCode == 2104:
            # Market data farm connected - info only
            logger.info(errorString)
        elif errorCode == 2106:
            # HMDS data farm connected - info only  
            logger.info(errorString)
        else:
            logger.error(f"Error {errorCode}: {errorString}")


# Example usage pattern
if __name__ == '__main__':
    # This pattern ensures proper event loop handling
    service = AsyncService("ExampleService")
    util.run(service.start())  # util.run handles the event loop properly