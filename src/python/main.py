"""
IBKR Spread Automation - Python Service Entry Point
Following vibe-driven async patterns for smooth flow
"""
import asyncio
import os
import signal
import sys
from typing import Optional

import uvloop
from aiohttp import web
from ib_insync import IB, util
from ib_insync.ibcontroller import Watchdog

from core.logging import setup_logging
from core.connection import AsyncIBKRService
from api.server import create_app
from monitoring.metrics import setup_metrics

# Configure async logger
logger = setup_logging("ibkr-python-service")


class ServiceManager:
    """Manages the lifecycle of our async IBKR service"""
    
    def __init__(self):
        self.ibkr_service: Optional[AsyncIBKRService] = None
        self.web_app: Optional[web.Application] = None
        self.runner: Optional[web.AppRunner] = None
        self.shutdown_event = asyncio.Event()
        
    async def startup(self):
        """Initialize all services with proper async flow"""
        logger.info("🚀 Starting IBKR Python Service...")
        
        # Setup metrics
        setup_metrics()
        
        # Initialize IBKR service
        self.ibkr_service = AsyncIBKRService()
        await self.ibkr_service.connect()
        
        # Create web application
        self.web_app = await create_app(self.ibkr_service)
        
        # Start web server
        self.runner = web.AppRunner(self.web_app)
        await self.runner.setup()
        
        port = int(os.getenv("API_PORT", "8080"))
        site = web.TCPSite(self.runner, "0.0.0.0", port)
        await site.start()
        
        logger.info(f"✅ Service running on port {port}")
        logger.info("🌊 Ready to ride the trading waves!")
        
    async def shutdown(self):
        """Gracefully shutdown all services"""
        logger.info("🛑 Shutting down services...")
        
        if self.runner:
            await self.runner.cleanup()
            
        if self.ibkr_service:
            await self.ibkr_service.disconnect()
            
        logger.info("👋 Service shutdown complete")
        
    async def run(self):
        """Main service loop"""
        # Setup signal handlers
        for sig in (signal.SIGTERM, signal.SIGINT):
            asyncio.get_running_loop().add_signal_handler(
                sig, lambda: self.shutdown_event.set()
            )
        
        try:
            await self.startup()
            await self.shutdown_event.wait()
        finally:
            await self.shutdown()


async def main():
    """Entry point with uvloop for maximum performance"""
    # Configure ib-insync for async operation
    util.startLoop()
    
    # Run service
    service = ServiceManager()
    await service.run()


if __name__ == "__main__":
    # Use uvloop for better async performance
    uvloop.install()
    
    # Run the service
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt")
        sys.exit(0)