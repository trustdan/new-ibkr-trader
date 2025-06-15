"""
TWS Watchdog Component - Auto-reconnection and Health Monitoring
Implements robust connection recovery for TWS daily restarts and network issues.
"""

import asyncio
import logging
import time
import socket
from datetime import datetime, time as dt_time
from typing import Optional, Dict, Any, Callable
from enum import Enum

from .connection import ConnectionManager, ConnectionState
from .events import EventManager
from .exceptions import ConnectionError, WatchdogError


logger = logging.getLogger(__name__)


class WatchdogState(Enum):
    """Watchdog operational states."""
    STOPPED = "stopped"
    MONITORING = "monitoring"
    RECONNECTING = "reconnecting"
    WAITING_FOR_RESTART = "waiting_for_restart"
    ERROR = "error"


class ConnectionWatchdog:
    """
    Monitors and maintains TWS connection health.
    
    Features:
    - Automatic reconnection on connection loss
    - Daily restart handling (11:45 PM EST)
    - Health monitoring with metrics
    - Configurable retry strategies
    - Event-driven notifications
    """
    
    def __init__(self, connection_manager: ConnectionManager):
        """
        Initialize watchdog.
        
        Args:
            connection_manager: The connection manager to monitor
        """
        self.connection_manager = connection_manager
        self.event_manager = EventManager()
        
        # Watchdog state
        self.state = WatchdogState.STOPPED
        self.monitoring_task: Optional[asyncio.Task] = None
        self.last_health_check = None
        
        # Configuration
        self.health_check_interval = 30  # seconds
        self.reconnect_delay_base = 5    # seconds
        self.max_reconnect_attempts = 10
        self.daily_restart_time = dt_time(23, 45)  # 11:45 PM
        
        # Statistics
        self.stats = {
            'start_time': None,
            'uptime_seconds': 0,
            'reconnect_count': 0,
            'last_reconnect': None,
            'health_checks': 0,
            'failed_health_checks': 0,
            'daily_restarts_handled': 0
        }
        
        # Setup event handlers
        self._setup_event_handlers()
    
    def _setup_event_handlers(self) -> None:
        """Set up event handlers for connection events."""
        # Listen for connection events
        self.connection_manager.event_manager.subscribe(
            'connection_lost', self._on_connection_lost
        )
        self.connection_manager.event_manager.subscribe(
            'connection_established', self._on_connection_established
        )
    
    async def start(self) -> None:
        """Start watchdog monitoring."""
        if self.state != WatchdogState.STOPPED:
            logger.warning("Watchdog already running")
            return
        
        logger.info("🐕 Starting Connection Watchdog")
        self.state = WatchdogState.MONITORING
        self.stats['start_time'] = datetime.now()
        
        # Start monitoring task
        self.monitoring_task = asyncio.create_task(self._monitoring_loop())
        
        await self.event_manager.emit('watchdog_started', {
            'timestamp': datetime.now(),
            'config': {
                'health_check_interval': self.health_check_interval,
                'max_reconnect_attempts': self.max_reconnect_attempts,
                'daily_restart_time': str(self.daily_restart_time)
            }
        })
    
    async def stop(self) -> None:
        """Stop watchdog monitoring."""
        if self.state == WatchdogState.STOPPED:
            return
        
        logger.info("🛑 Stopping Connection Watchdog")
        self.state = WatchdogState.STOPPED
        
        # Cancel monitoring task
        if self.monitoring_task:
            self.monitoring_task.cancel()
            try:
                await self.monitoring_task
            except asyncio.CancelledError:
                pass
        
        await self.event_manager.emit('watchdog_stopped', {
            'timestamp': datetime.now(),
            'uptime_seconds': self.get_uptime(),
            'stats': self.stats.copy()
        })
    
    async def _monitoring_loop(self) -> None:
        """Main monitoring loop."""
        logger.info("🔍 Watchdog monitoring loop started")
        
        try:
            while self.state != WatchdogState.STOPPED:
                await self._perform_health_check()
                await self._check_daily_restart()
                await asyncio.sleep(self.health_check_interval)
                
        except asyncio.CancelledError:
            logger.info("Watchdog monitoring loop cancelled")
        except Exception as e:
            logger.error(f"Watchdog monitoring loop error: {e}")
            self.state = WatchdogState.ERROR
            await self.event_manager.emit('watchdog_error', {
                'error': str(e),
                'timestamp': datetime.now()
            })
    
    async def _perform_health_check(self) -> None:
        """Perform connection health check."""
        self.stats['health_checks'] += 1
        self.last_health_check = datetime.now()
        
        try:
            # Test socket connectivity first (fast check)
            health_ok = await self._check_socket_health()
            
            if health_ok and self.connection_manager.is_connected():
                # Connection is healthy
                await self.event_manager.emit('health_check_passed', {
                    'timestamp': self.last_health_check,
                    'connection_state': self.connection_manager.state.value
                })
                return
            
            # Connection appears unhealthy
            logger.warning("Health check failed - connection unhealthy")
            self.stats['failed_health_checks'] += 1
            
            await self.event_manager.emit('health_check_failed', {
                'timestamp': self.last_health_check,
                'socket_ok': health_ok,
                'connection_state': self.connection_manager.state.value
            })
            
            # Trigger reconnection
            await self._handle_reconnection()
            
        except Exception as e:
            logger.error(f"Health check error: {e}")
            self.stats['failed_health_checks'] += 1
    
    async def _check_socket_health(self) -> bool:
        """Check if TWS socket is responsive."""
        try:
            # Quick socket connectivity test
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(3)
            
            config = self.connection_manager.config.connection
            result = sock.connect_ex((config.host, config.port))
            sock.close()
            
            return result == 0
            
        except Exception as e:
            logger.debug(f"Socket health check failed: {e}")
            return False
    
    async def _check_daily_restart(self) -> None:
        """Check if we're approaching daily TWS restart time."""
        now = datetime.now().time()
        
        # Check if we're within 5 minutes of restart time
        restart_window_start = dt_time(
            self.daily_restart_time.hour,
            max(0, self.daily_restart_time.minute - 5)
        )
        restart_window_end = dt_time(
            self.daily_restart_time.hour,
            min(59, self.daily_restart_time.minute + 5)
        )
        
        if restart_window_start <= now <= restart_window_end:
            logger.info("📅 Approaching TWS daily restart window")
            await self._handle_daily_restart()
    
    async def _handle_daily_restart(self) -> None:
        """Handle TWS daily restart gracefully."""
        if self.state == WatchdogState.WAITING_FOR_RESTART:
            return  # Already handling restart
        
        logger.info("🔄 Preparing for TWS daily restart")
        self.state = WatchdogState.WAITING_FOR_RESTART
        self.stats['daily_restarts_handled'] += 1
        
        await self.event_manager.emit('daily_restart_detected', {
            'timestamp': datetime.now(),
            'restart_time': str(self.daily_restart_time)
        })
        
        # Gracefully disconnect
        if self.connection_manager.is_connected():
            await self.connection_manager.disconnect()
        
        # Wait for restart window to pass
        await asyncio.sleep(300)  # Wait 5 minutes
        
        # Attempt reconnection
        await self._handle_reconnection()
        self.state = WatchdogState.MONITORING
    
    async def _handle_reconnection(self) -> None:
        """Handle connection recovery."""
        if self.state == WatchdogState.RECONNECTING:
            return  # Already reconnecting
        
        logger.info("🔄 Starting connection recovery")
        self.state = WatchdogState.RECONNECTING
        self.stats['reconnect_count'] += 1
        self.stats['last_reconnect'] = datetime.now()
        
        await self.event_manager.emit('reconnection_started', {
            'timestamp': self.stats['last_reconnect'],
            'attempt_number': self.stats['reconnect_count']
        })
        
        for attempt in range(1, self.max_reconnect_attempts + 1):
            try:
                logger.info(f"🔄 Reconnection attempt {attempt}/{self.max_reconnect_attempts}")
                
                # Ensure clean disconnect first
                if self.connection_manager.is_connected():
                    await self.connection_manager.disconnect()
                
                # Wait with exponential backoff
                delay = self.reconnect_delay_base * (2 ** (attempt - 1))
                await asyncio.sleep(min(delay, 60))  # Max 60 seconds
                
                # Attempt reconnection
                await self.connection_manager.connect()
                
                if self.connection_manager.is_connected():
                    logger.info("✅ Reconnection successful!")
                    self.state = WatchdogState.MONITORING
                    
                    await self.event_manager.emit('reconnection_successful', {
                        'timestamp': datetime.now(),
                        'attempt_number': attempt,
                        'total_attempts': attempt
                    })
                    return
                
            except Exception as e:
                logger.warning(f"Reconnection attempt {attempt} failed: {e}")
                
                await self.event_manager.emit('reconnection_attempt_failed', {
                    'timestamp': datetime.now(),
                    'attempt_number': attempt,
                    'error': str(e)
                })
        
        # All reconnection attempts failed
        logger.error("❌ All reconnection attempts failed")
        self.state = WatchdogState.ERROR
        
        await self.event_manager.emit('reconnection_failed', {
            'timestamp': datetime.now(),
            'total_attempts': self.max_reconnect_attempts
        })
        
        raise WatchdogError(
            f"Failed to reconnect after {self.max_reconnect_attempts} attempts"
        )
    
    async def _on_connection_lost(self, event_data: Dict[str, Any]) -> None:
        """Handle connection lost event."""
        logger.warning("🔌 Connection lost event received")
        
        if self.state == WatchdogState.MONITORING:
            await self._handle_reconnection()
    
    async def _on_connection_established(self, event_data: Dict[str, Any]) -> None:
        """Handle connection established event."""
        logger.info("✅ Connection established event received")
        
        if self.state == WatchdogState.RECONNECTING:
            self.state = WatchdogState.MONITORING
    
    def get_uptime(self) -> float:
        """Get watchdog uptime in seconds."""
        if self.stats['start_time']:
            return (datetime.now() - self.stats['start_time']).total_seconds()
        return 0
    
    def get_health_status(self) -> Dict[str, Any]:
        """Get comprehensive health status."""
        return {
            'watchdog_state': self.state.value,
            'connection_state': self.connection_manager.state.value,
            'uptime_seconds': self.get_uptime(),
            'last_health_check': self.last_health_check,
            'stats': self.stats.copy(),
            'config': {
                'health_check_interval': self.health_check_interval,
                'max_reconnect_attempts': self.max_reconnect_attempts,
                'daily_restart_time': str(self.daily_restart_time)
            }
        }
    
    # Event subscription methods for external monitoring
    def on_watchdog_started(self, callback: Callable) -> None:
        """Subscribe to watchdog started events."""
        self.event_manager.subscribe('watchdog_started', callback)
    
    def on_reconnection_started(self, callback: Callable) -> None:
        """Subscribe to reconnection started events."""
        self.event_manager.subscribe('reconnection_started', callback)
    
    def on_reconnection_successful(self, callback: Callable) -> None:
        """Subscribe to successful reconnection events."""
        self.event_manager.subscribe('reconnection_successful', callback)
    
    def on_daily_restart_detected(self, callback: Callable) -> None:
        """Subscribe to daily restart events."""
        self.event_manager.subscribe('daily_restart_detected', callback) 