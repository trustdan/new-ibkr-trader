"""
Unit tests for event management system.
"""

import pytest
import asyncio
from datetime import datetime

from src.python.ibkr_connector.events import EventManager, Events


class TestEventManager:
    """Test cases for EventManager."""
    
    @pytest.fixture
    def event_manager(self):
        """Create event manager instance."""
        return EventManager()
    
    def test_initialization(self, event_manager):
        """Test event manager initialization."""
        assert event_manager._handlers == {}
        assert event_manager._async_handlers == {}
        assert event_manager._event_history == []
        assert event_manager._history_limit == 1000
    
    def test_register_sync_handler(self, event_manager):
        """Test registering synchronous event handler."""
        def handler(data):
            pass
        
        event_manager.on('test_event', handler)
        assert handler in event_manager._handlers['test_event']
        assert event_manager.handler_count('test_event') == {'test_event': 1}
    
    def test_register_async_handler(self, event_manager):
        """Test registering asynchronous event handler."""
        async def handler(data):
            pass
        
        event_manager.on('test_event', handler)
        assert handler in event_manager._async_handlers['test_event']
        assert event_manager.handler_count('test_event') == {'test_event': 1}
    
    def test_unregister_handler(self, event_manager):
        """Test unregistering event handler."""
        def handler(data):
            pass
        
        event_manager.on('test_event', handler)
        event_manager.off('test_event', handler)
        
        assert handler not in event_manager._handlers['test_event']
        assert event_manager.handler_count('test_event') == {'test_event': 0}
    
    @pytest.mark.asyncio
    async def test_emit_to_sync_handler(self, event_manager):
        """Test emitting event to synchronous handler."""
        received_data = []
        
        def handler(data):
            received_data.append(data)
        
        event_manager.on('test_event', handler)
        await event_manager.emit('test_event', {'value': 42})
        
        assert len(received_data) == 1
        assert received_data[0]['value'] == 42
        assert received_data[0]['_event_name'] == 'test_event'
        assert '_timestamp' in received_data[0]
    
    @pytest.mark.asyncio
    async def test_emit_to_async_handler(self, event_manager):
        """Test emitting event to asynchronous handler."""
        received_data = []
        
        async def handler(data):
            received_data.append(data)
        
        event_manager.on('test_event', handler)
        await event_manager.emit('test_event', {'value': 42})
        
        assert len(received_data) == 1
        assert received_data[0]['value'] == 42
    
    @pytest.mark.asyncio
    async def test_emit_to_multiple_handlers(self, event_manager):
        """Test emitting event to multiple handlers."""
        sync_data = []
        async_data = []
        
        def sync_handler(data):
            sync_data.append(data)
        
        async def async_handler(data):
            async_data.append(data)
        
        event_manager.on('test_event', sync_handler)
        event_manager.on('test_event', async_handler)
        
        await event_manager.emit('test_event', {'value': 'test'})
        
        assert len(sync_data) == 1
        assert len(async_data) == 1
        assert sync_data[0]['value'] == 'test'
        assert async_data[0]['value'] == 'test'
    
    @pytest.mark.asyncio
    async def test_handler_exception_handling(self, event_manager):
        """Test that exceptions in handlers don't break event emission."""
        results = []
        
        def bad_handler(data):
            raise ValueError("Handler error")
        
        def good_handler(data):
            results.append("good")
        
        event_manager.on('test_event', bad_handler)
        event_manager.on('test_event', good_handler)
        
        # Should not raise exception
        await event_manager.emit('test_event', {})
        
        assert results == ["good"]
    
    @pytest.mark.asyncio
    async def test_event_history(self, event_manager):
        """Test event history tracking."""
        await event_manager.emit('event1', {'data': 1})
        await event_manager.emit('event2', {'data': 2})
        await event_manager.emit('event1', {'data': 3})
        
        # Get all history
        history = event_manager.get_history()
        assert len(history) == 3
        
        # Filter by event name
        event1_history = event_manager.get_history('event1')
        assert len(event1_history) == 2
        assert all(e['event'] == 'event1' for e in event1_history)
        
        # Limit results
        limited = event_manager.get_history(limit=2)
        assert len(limited) == 2
    
    @pytest.mark.asyncio
    async def test_history_limit(self, event_manager):
        """Test that history respects size limit."""
        event_manager._history_limit = 5
        
        for i in range(10):
            await event_manager.emit('test', {'index': i})
        
        history = event_manager.get_history()
        assert len(history) == 5
        assert history[0]['data']['index'] == 5  # Oldest should be index 5
        assert history[-1]['data']['index'] == 9  # Newest should be index 9
    
    def test_clear_history(self, event_manager):
        """Test clearing event history."""
        event_manager._event_history = [{'test': 'data'}]
        event_manager.clear_history()
        assert event_manager._event_history == []
    
    def test_handler_count(self, event_manager):
        """Test handler counting."""
        def handler1(data): pass
        def handler2(data): pass
        async def handler3(data): pass
        
        event_manager.on('event1', handler1)
        event_manager.on('event1', handler2)
        event_manager.on('event2', handler3)
        
        # Count for specific event
        assert event_manager.handler_count('event1') == {'event1': 2}
        assert event_manager.handler_count('event2') == {'event2': 1}
        
        # Count all
        all_counts = event_manager.handler_count()
        assert all_counts == {'event1': 2, 'event2': 1}
    
    def test_event_constants(self):
        """Test that event constants are defined."""
        assert Events.CONNECTION_ESTABLISHED == "connection_established"
        assert Events.ORDER_FILLED == "order_filled"
        assert Events.RATE_LIMIT_EXCEEDED == "rate_limit_exceeded"
        assert Events.DAILY_RESTART_PENDING == "daily_restart_pending"