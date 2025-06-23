"""
Minimal IBKR service for GUI integration testing
Provides basic API endpoints without complex ib-insync dependencies
"""
import asyncio
import json
from aiohttp import web, log
import logging

# Simple logging setup
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("minimal-ibkr-service")

class MinimalIBKRService:
    """Minimal service to test GUI integration"""
    
    def __init__(self):
        self.connected = False
        self.accounts = ["DU123456"]  # Mock paper account
        
    async def get_health(self):
        """Health check endpoint"""
        return {
            "status": "healthy",
            "tws_connected": self.connected,
            "service": "ibkr-python-minimal",
            "accounts": self.accounts
        }
        
    async def connect_tws(self):
        """Mock TWS connection"""
        # Simulate connection
        await asyncio.sleep(0.1)
        self.connected = True
        logger.info("Mock TWS connection established")
        return True
        
    async def get_option_chain(self, symbol):
        """Mock option chain data"""
        return {
            "symbol": symbol,
            "chains": [
                {
                    "strike": 580,
                    "expiry": "2025-02-21",
                    "call_bid": 2.50,
                    "call_ask": 2.60,
                    "put_bid": 1.20,
                    "put_ask": 1.30,
                    "delta": 0.30,
                    "volume": 150
                },
                {
                    "strike": 585,
                    "expiry": "2025-02-21", 
                    "call_bid": 2.10,
                    "call_ask": 2.20,
                    "put_bid": 1.50,
                    "put_ask": 1.60,
                    "delta": 0.25,
                    "volume": 200
                }
            ]
        }

# Create global service instance
ibkr_service = MinimalIBKRService()

async def health_handler(request):
    """Health check endpoint"""
    health_data = await ibkr_service.get_health()
    return web.json_response(health_data)

async def connect_handler(request):
    """Connect to TWS endpoint"""
    success = await ibkr_service.connect_tws()
    return web.json_response({"connected": success})

async def option_chain_handler(request):
    """Get option chain endpoint"""
    symbol = request.match_info.get('symbol', 'SPY')
    chain_data = await ibkr_service.get_option_chain(symbol)
    return web.json_response(chain_data)

async def create_app():
    """Create the web application"""
    app = web.Application()
    
    # Add CORS headers for GUI
    @web.middleware
    async def cors_handler(request, handler):
        response = await handler(request)
        response.headers['Access-Control-Allow-Origin'] = '*'
        response.headers['Access-Control-Allow-Methods'] = 'GET, POST, OPTIONS'
        response.headers['Access-Control-Allow-Headers'] = 'Content-Type'
        return response
    
    app.middlewares.append(cors_handler)
    
    # Routes
    app.router.add_get('/health', health_handler)
    app.router.add_post('/connect', connect_handler)
    app.router.add_get('/option_chain/{symbol}', option_chain_handler)
    
    return app

async def main():
    """Main service entry point"""
    logger.info("🚀 Starting Minimal IBKR Service...")
    
    # Create and start app
    app = await create_app()
    runner = web.AppRunner(app)
    await runner.setup()
    
    site = web.TCPSite(runner, '0.0.0.0', 8080)
    await site.start()
    
    logger.info("✅ Minimal IBKR Service running on port 8080")
    logger.info("🔗 Health check: http://localhost:8080/health")
    
    # Auto-connect on startup
    await ibkr_service.connect_tws()
    
    # Keep running
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        logger.info("👋 Service shutting down...")
        await runner.cleanup()

if __name__ == "__main__":
    asyncio.run(main()) 