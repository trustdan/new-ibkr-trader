# Watchdog Pattern for Automatic Reconnection
from ib_insync import IB, util
from ib_insync.ibcontroller import Watchdog
import logging
import asyncio

class ResilientIBConnection:
    """Template for bulletproof IB connection with auto-recovery"""
    
    def __init__(self):
        self.ib = IB()
        self.watchdog = None
        
    def start_with_watchdog(self, host='localhost', port=7497, clientId=1):
        """Start IB connection with watchdog for auto-reconnection"""
        
        # Configure watchdog
        self.watchdog = Watchdog(
            controller=self.ib,
            host=host,
            port=port,
            clientId=clientId,
            connectTimeout=10,
            appStartupTime=15,
            appTimeout=20,
            retryDelay=2,
            readonly=False,  # Set True for read-only connection
            account='',      # Leave empty for default account
            password='',     # TWS password if needed
            userid=''        # TWS username if needed
        )
        
        # Set up watchdog event handlers
        self.watchdog.startingEvent += self._on_watchdog_starting
        self.watchdog.startedEvent += self._on_watchdog_started
        self.watchdog.stoppingEvent += self._on_watchdog_stopping
        self.watchdog.stoppedEvent += self._on_watchdog_stopped
        self.watchdog.softTimeoutEvent += self._on_soft_timeout
        self.watchdog.hardTimeoutEvent += self._on_hard_timeout
        
        # Start the watchdog (it will handle connection)
        self.watchdog.start()
        logging.info("🐕 Watchdog started - connection resilience active")
        
    def _on_watchdog_starting(self):
        """Called when watchdog is starting up"""
        logging.info("Watchdog starting...")
        
    def _on_watchdog_started(self):
        """Called when watchdog has successfully started"""
        logging.info("✅ Watchdog started - connection protected")
        
    def _on_watchdog_stopping(self):
        """Called when watchdog is shutting down"""
        logging.info("Watchdog stopping...")
        
    def _on_watchdog_stopped(self):
        """Called when watchdog has stopped"""
        logging.info("Watchdog stopped")
        
    def _on_soft_timeout(self):
        """Called on soft timeout - watchdog will try to reconnect"""
        logging.warning("⚠️ Soft timeout - attempting reconnection")
        
    def _on_hard_timeout(self):
        """Called on hard timeout - watchdog will restart TWS/Gateway"""
        logging.error("❌ Hard timeout - restarting TWS/Gateway")
        
    def stop(self):
        """Gracefully stop the watchdog and connection"""
        if self.watchdog:
            self.watchdog.stop()
            
# Example: Running with watchdog protection
async def run_with_watchdog():
    """Example of running with automatic reconnection"""
    
    connection = ResilientIBConnection()
    
    # Start with watchdog protection
    connection.start_with_watchdog()
    
    # Wait for connection
    while not connection.ib.isConnected():
        await asyncio.sleep(0.1)
        
    logging.info(f"Connected: {connection.ib.isConnected()}")
    
    # Your trading logic here
    # The watchdog will automatically reconnect if connection drops
    
    try:
        # Keep running
        await asyncio.sleep(86400)  # Run for 24 hours
    except KeyboardInterrupt:
        logging.info("Shutting down...")
    finally:
        connection.stop()

# Alternative: Simple reconnection without full watchdog
async def simple_reconnect_pattern():
    """Simpler pattern for handling disconnections"""
    
    ib = IB()
    
    async def connect_with_retry(max_retries=5):
        for attempt in range(max_retries):
            try:
                await ib.connectAsync('localhost', 7497, clientId=1)
                logging.info("Connected successfully")
                return True
            except Exception as e:
                logging.error(f"Connection attempt {attempt + 1} failed: {e}")
                if attempt < max_retries - 1:
                    await asyncio.sleep(2 ** attempt)  # Exponential backoff
        return False
    
    # Set up disconnection handler
    def on_disconnected():
        logging.warning("Disconnected! Attempting to reconnect...")
        asyncio.create_task(connect_with_retry())
    
    ib.disconnectedEvent += on_disconnected
    
    # Initial connection
    if await connect_with_retry():
        # Run your trading logic
        await ib.sleep(3600)
    else:
        logging.error("Failed to establish connection")
        
if __name__ == '__main__':
    # Choose your pattern:
    # util.run(run_with_watchdog())  # Full watchdog protection
    util.run(simple_reconnect_pattern())  # Simple reconnection