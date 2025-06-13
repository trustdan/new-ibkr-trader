"""
Event management system for IBKR connector.

This module provides a centralized event bus for handling and distributing
events throughout the IBKR connector system.
"""

import asyncio
import logging
from typing import Dict, List, Callable, Any, Optional
from collections import defaultdict
from datetime import datetime


class EventManager:
    """
    Centralized event manager for the IBKR connector.
    
    Provides pub-sub functionality for system-wide event handling,
    supporting both sync and async event handlers.
    """
    
    def __init__(self):
        """Initialize the event manager."""
        self._handlers: Dict[str, List[Callable]] = defaultdict(list)
        self._async_handlers: Dict[str, List[Callable]] = defaultdict(list)
        self.logger = logging.getLogger(__name__)
        self._event_history: List[Dict[str, Any]] = []
        self._history_limit = 1000
    
    def on(self, event_name: str, handler: Callable) -> None:
        """
        Register a synchronous event handler.
        
        Args:
            event_name: Name of the event to listen for
            handler: Callable to invoke when event occurs
        """
        if asyncio.iscoroutinefunction(handler):
            self._async_handlers[event_name].append(handler)
            self.logger.debug(f"Registered async handler for '{event_name}'")
        else:
            self._handlers[event_name].append(handler)
            self.logger.debug(f"Registered sync handler for '{event_name}'")
    
    def off(self, event_name: str, handler: Callable) -> None:
        """
        Unregister an event handler.
        
        Args:
            event_name: Name of the event
            handler: Handler to remove
        """
        if handler in self._handlers[event_name]:
            self._handlers[event_name].remove(handler)
            self.logger.debug(f"Removed sync handler for '{event_name}'")
        elif handler in self._async_handlers[event_name]:
            self._async_handlers[event_name].remove(handler)
            self.logger.debug(f"Removed async handler for '{event_name}'")
    
    async def emit(self, event_name: str, data: Optional[Dict[str, Any]] = None) -> None:
        """
        Emit an event to all registered handlers.
        
        Args:
            event_name: Name of the event to emit
            data: Optional event data
        """
        data = data or {}
        data['_event_name'] = event_name
        data['_timestamp'] = datetime.now()
        
        # Store in history
        self._add_to_history(event_name, data)
        
        self.logger.debug(f"Emitting event '{event_name}' with data: {data}")
        
        # Call sync handlers
        for handler in self._handlers[event_name]:
            try:
                handler(data)
            except Exception as e:
                self.logger.error(
                    f"Error in sync handler for '{event_name}': {e}",
                    exc_info=True
                )
        
        # Call async handlers
        tasks = []
        for handler in self._async_handlers[event_name]:
            tasks.append(self._call_async_handler(handler, event_name, data))
        
        if tasks:
            await asyncio.gather(*tasks, return_exceptions=True)
    
    async def _call_async_handler(
        self, 
        handler: Callable, 
        event_name: str, 
        data: Dict[str, Any]
    ) -> None:
        """Call an async handler with error handling."""
        try:
            await handler(data)
        except Exception as e:
            self.logger.error(
                f"Error in async handler for '{event_name}': {e}",
                exc_info=True
            )
    
    def _add_to_history(self, event_name: str, data: Dict[str, Any]) -> None:
        """Add event to history, maintaining size limit."""
        self._event_history.append({
            'event': event_name,
            'data': data.copy(),
            'timestamp': datetime.now()
        })
        
        # Trim history if needed
        if len(self._event_history) > self._history_limit:
            self._event_history = self._event_history[-self._history_limit:]
    
    def get_history(
        self, 
        event_name: Optional[str] = None,
        limit: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        """
        Get event history.
        
        Args:
            event_name: Filter by specific event name
            limit: Maximum number of events to return
            
        Returns:
            List of historical events
        """
        history = self._event_history
        
        if event_name:
            history = [e for e in history if e['event'] == event_name]
        
        if limit:
            history = history[-limit:]
        
        return history
    
    def clear_history(self) -> None:
        """Clear event history."""
        self._event_history.clear()
    
    def handler_count(self, event_name: Optional[str] = None) -> Dict[str, int]:
        """
        Get count of registered handlers.
        
        Args:
            event_name: Specific event to check, or None for all
            
        Returns:
            Dictionary of event names to handler counts
        """
        if event_name:
            sync_count = len(self._handlers.get(event_name, []))
            async_count = len(self._async_handlers.get(event_name, []))
            return {event_name: sync_count + async_count}
        
        counts = {}
        for name in set(self._handlers.keys()) | set(self._async_handlers.keys()):
            sync_count = len(self._handlers.get(name, []))
            async_count = len(self._async_handlers.get(name, []))
            counts[name] = sync_count + async_count
        
        return counts
    
    def __repr__(self) -> str:
        """String representation of event manager."""
        total_handlers = sum(self.handler_count().values())
        return f"EventManager(handlers={total_handlers}, events={len(self._event_history)})"


# Common event names as constants
class Events:
    """Standard event names used throughout the system."""
    
    # Connection events
    CONNECTION_ESTABLISHED = "connection_established"
    CONNECTION_LOST = "connection_lost"
    CONNECTION_RECOVERED = "connection_recovered"
    CONNECTION_ERROR = "connection_error"
    
    # Order events
    ORDER_PLACED = "order_placed"
    ORDER_FILLED = "order_filled"
    ORDER_CANCELLED = "order_cancelled"
    ORDER_REJECTED = "order_rejected"
    ORDER_STATUS_CHANGED = "order_status_changed"
    
    # Market data events
    TICKER_UPDATE = "ticker_update"
    OPTION_CHAIN_UPDATE = "option_chain_update"
    MARKET_DATA_ERROR = "market_data_error"
    
    # System events
    RATE_LIMIT_WARNING = "rate_limit_warning"
    RATE_LIMIT_EXCEEDED = "rate_limit_exceeded"
    ERROR_OCCURRED = "error_occurred"
    SYSTEM_READY = "system_ready"
    SYSTEM_SHUTDOWN = "system_shutdown"
    
    # Watchdog events
    WATCHDOG_RECONNECTING = "watchdog_reconnecting"
    WATCHDOG_RECONNECTED = "watchdog_reconnected"
    DAILY_RESTART_PENDING = "daily_restart_pending"