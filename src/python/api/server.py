"""
Async HTTP API server for IBKR service
Provides REST endpoints for trading operations
"""
from aiohttp import web
from prometheus_client import generate_latest
import json

from ..core.logging import get_logger
from ..monitoring.metrics import api_requests, MetricsTimer, api_request_duration

logger = get_logger(__name__)


async def create_app(ibkr_service) -> web.Application:
    """Create and configure the aiohttp application"""
    app = web.Application()
    
    # Store service reference
    app['ibkr_service'] = ibkr_service
    
    # Setup routes
    app.router.add_get('/health', health_check)
    app.router.add_get('/metrics', metrics_handler)
    app.router.add_get('/api/v1/connection/test', test_connection)
    app.router.add_get('/api/v1/account/summary', get_account_summary)
    
    # Setup middleware
    app.middlewares.append(error_middleware)
    app.middlewares.append(metrics_middleware)
    
    logger.info("API server configured")
    return app


# Health check endpoint
async def health_check(request: web.Request) -> web.Response:
    """Simple health check endpoint"""
    ibkr_service = request.app['ibkr_service']
    
    health = {
        "status": "healthy" if ibkr_service.connected else "unhealthy",
        "connected": ibkr_service.connected,
        "service": "ibkr-python-service"
    }
    
    status_code = 200 if ibkr_service.connected else 503
    return web.json_response(health, status=status_code)


# Metrics endpoint
async def metrics_handler(request: web.Request) -> web.Response:
    """Prometheus metrics endpoint"""
    metrics = generate_latest()
    return web.Response(
        body=metrics,
        content_type="text/plain; version=0.0.4"
    )


# Connection test endpoint
async def test_connection(request: web.Request) -> web.Response:
    """Test IBKR connection status"""
    ibkr_service = request.app['ibkr_service']
    
    try:
        result = await ibkr_service.test_connection()
        return web.json_response(result)
    except Exception as e:
        logger.error(f"Connection test failed: {e}")
        return web.json_response(
            {"error": str(e)}, 
            status=500
        )


# Account summary endpoint
async def get_account_summary(request: web.Request) -> web.Response:
    """Get account summary information"""
    ibkr_service = request.app['ibkr_service']
    
    try:
        summary = await ibkr_service.get_account_summary()
        return web.json_response(summary)
    except ConnectionError:
        return web.json_response(
            {"error": "Not connected to TWS"}, 
            status=503
        )
    except Exception as e:
        logger.error(f"Failed to get account summary: {e}")
        return web.json_response(
            {"error": str(e)}, 
            status=500
        )


# Middleware
@web.middleware
async def error_middleware(request, handler):
    """Global error handling middleware"""
    try:
        return await handler(request)
    except web.HTTPException:
        raise
    except Exception as e:
        logger.error(f"Unhandled error: {e}", exc_info=True)
        return web.json_response(
            {"error": "Internal server error"},
            status=500
        )


@web.middleware
async def metrics_middleware(request, handler):
    """Track request metrics"""
    endpoint = request.path
    
    # Skip metrics endpoint to avoid recursion
    if endpoint == '/metrics':
        return await handler(request)
    
    # Track request
    api_requests.labels(endpoint=endpoint).inc()
    
    # Time the request
    with MetricsTimer(api_request_duration, {'endpoint': endpoint}):
        response = await handler(request)
    
    return response