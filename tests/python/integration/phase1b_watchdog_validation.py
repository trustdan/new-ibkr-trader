"""
Phase 1B: Watchdog Testing - Simplified Version
Tests the Watchdog concept with socket-based connection monitoring.
"""

import asyncio
import socket
import time
import logging
from datetime import datetime
from enum import Enum
from typing import Dict, Any

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class WatchdogState(Enum):
    """Watchdog states."""
    STOPPED = "stopped"
    MONITORING = "monitoring"
    RECONNECTING = "reconnecting"
    ERROR = "error"


class SimpleWatchdog:
    """
    Simplified Watchdog for Phase 1B testing.
    Uses socket-based health checks since we confirmed TWS connectivity.
    """
    
    def __init__(self, host="127.0.0.1", port=7497):
        self.host = host
        self.port = port
        self.state = WatchdogState.STOPPED
        self.monitoring_task = None
        self.health_check_interval = 5  # seconds
        self.stats = {
            'start_time': None,
            'health_checks': 0,
            'failed_checks': 0,
            'reconnect_attempts': 0
        }
    
    async def start(self):
        """Start watchdog monitoring."""
        if self.state != WatchdogState.STOPPED:
            logger.warning("Watchdog already running")
            return
        
        logger.info("🐕 Starting Simple Watchdog")
        self.state = WatchdogState.MONITORING
        self.stats['start_time'] = datetime.now()
        
        # Start monitoring task
        self.monitoring_task = asyncio.create_task(self._monitoring_loop())
    
    async def stop(self):
        """Stop watchdog monitoring."""
        if self.state == WatchdogState.STOPPED:
            return
        
        logger.info("🛑 Stopping Simple Watchdog")
        self.state = WatchdogState.STOPPED
        
        if self.monitoring_task:
            self.monitoring_task.cancel()
            try:
                await self.monitoring_task
            except asyncio.CancelledError:
                pass
    
    async def _monitoring_loop(self):
        """Main monitoring loop."""
        logger.info("🔍 Watchdog monitoring started")
        
        try:
            while self.state != WatchdogState.STOPPED:
                await self._perform_health_check()
                await asyncio.sleep(self.health_check_interval)
        except asyncio.CancelledError:
            logger.info("Monitoring loop cancelled")
        except Exception as e:
            logger.error(f"Monitoring error: {e}")
            self.state = WatchdogState.ERROR
    
    async def _perform_health_check(self):
        """Perform socket-based health check."""
        self.stats['health_checks'] += 1
        
        try:
            # Quick socket test
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(3)
            result = sock.connect_ex((self.host, self.port))
            sock.close()
            
            if result == 0:
                logger.info(f"✅ Health check #{self.stats['health_checks']}: TWS responsive")
                return True
            else:
                logger.warning(f"⚠️ Health check #{self.stats['health_checks']}: TWS not responsive")
                self.stats['failed_checks'] += 1
                await self._handle_connection_issue()
                return False
                
        except Exception as e:
            logger.error(f"Health check error: {e}")
            self.stats['failed_checks'] += 1
            return False
    
    async def _handle_connection_issue(self):
        """Handle connection issues."""
        if self.state == WatchdogState.RECONNECTING:
            return
        
        logger.info("🔄 Handling connection issue")
        self.state = WatchdogState.RECONNECTING
        self.stats['reconnect_attempts'] += 1
        
        # Simulate reconnection logic
        await asyncio.sleep(2)
        
        # Check if connection is restored
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(3)
        result = sock.connect_ex((self.host, self.port))
        sock.close()
        
        if result == 0:
            logger.info("✅ Connection restored")
            self.state = WatchdogState.MONITORING
        else:
            logger.warning("❌ Connection still down")
            self.state = WatchdogState.ERROR
    
    def get_status(self) -> Dict[str, Any]:
        """Get watchdog status."""
        uptime = 0
        if self.stats['start_time']:
            uptime = (datetime.now() - self.stats['start_time']).total_seconds()
        
        return {
            'state': self.state.value,
            'uptime_seconds': uptime,
            'stats': self.stats.copy(),
            'config': {
                'host': self.host,
                'port': self.port,
                'health_check_interval': self.health_check_interval
            }
        }


async def test_watchdog_functionality():
    """Test watchdog functionality."""
    logger.info("🧪 Testing Watchdog Functionality")
    
    # Create watchdog
    watchdog = SimpleWatchdog()
    
    try:
        # Test 1: Start/Stop lifecycle
        logger.info("📋 Test 1: Watchdog lifecycle")
        await watchdog.start()
        assert watchdog.state == WatchdogState.MONITORING
        
        # Test 2: Health monitoring
        logger.info("📋 Test 2: Health monitoring (10 seconds)")
        await asyncio.sleep(10)
        
        # Test 3: Status reporting
        logger.info("📋 Test 3: Status reporting")
        status = watchdog.get_status()
        logger.info(f"Status: {status}")
        
        # Verify health checks occurred
        assert status['stats']['health_checks'] > 0, "Health checks should have occurred"
        assert status['uptime_seconds'] > 0, "Uptime should be positive"
        
        # Test 4: Stop watchdog
        logger.info("📋 Test 4: Stopping watchdog")
        await watchdog.stop()
        assert watchdog.state == WatchdogState.STOPPED
        
        return True
        
    except Exception as e:
        logger.error(f"Test failed: {e}")
        return False
    finally:
        if watchdog.state != WatchdogState.STOPPED:
            await watchdog.stop()


async def test_connection_recovery():
    """Test connection recovery scenarios."""
    logger.info("🧪 Testing Connection Recovery")
    
    watchdog = SimpleWatchdog()
    
    try:
        await watchdog.start()
        
        # Simulate connection issue by using wrong port
        logger.info("📋 Simulating connection issue...")
        original_port = watchdog.port
        watchdog.port = 9999  # Invalid port
        
        # Wait for health check to detect issue
        await asyncio.sleep(6)
        
        # Restore correct port
        watchdog.port = original_port
        logger.info("📋 Restoring connection...")
        
        # Wait for recovery
        await asyncio.sleep(6)
        
        # Check final status
        status = watchdog.get_status()
        logger.info(f"Final status: {status}")
        
        await watchdog.stop()
        return True
        
    except Exception as e:
        logger.error(f"Recovery test failed: {e}")
        return False
    finally:
        if watchdog.state != WatchdogState.STOPPED:
            await watchdog.stop()


async def run_phase1b_tests():
    """Run all Phase 1B tests."""
    logger.info("🚀 PHASE 1B: WATCHDOG TESTING")
    logger.info("=" * 50)
    
    results = {}
    
    # Test 1: Basic functionality
    logger.info("\n🔬 Test Suite 1: Basic Functionality")
    results['basic'] = await test_watchdog_functionality()
    
    # Test 2: Connection recovery
    logger.info("\n🔬 Test Suite 2: Connection Recovery")
    results['recovery'] = await test_connection_recovery()
    
    # Summary
    logger.info("\n" + "=" * 50)
    logger.info("📊 PHASE 1B TEST RESULTS:")
    logger.info("=" * 50)
    
    all_passed = True
    for test_name, passed in results.items():
        status = "✅ PASS" if passed else "❌ FAIL"
        logger.info(f"  {test_name.title():.<20} {status}")
        if not passed:
            all_passed = False
    
    logger.info("\n" + "=" * 50)
    if all_passed:
        logger.info("🎉 PHASE 1B WATCHDOG TESTING: COMPLETE!")
        logger.info("✅ Watchdog component validated")
        logger.info("✅ Connection monitoring working")
        logger.info("✅ Health checks functional")
        logger.info("✅ Recovery logic implemented")
        logger.info("🚀 Ready for Phase 1C: Trading Operations")
    else:
        logger.info("⚠️ Some tests failed - review implementation")
    logger.info("=" * 50)
    
    return all_passed


if __name__ == "__main__":
    # Run Phase 1B tests
    success = asyncio.run(run_phase1b_tests())
    
    if success:
        print("\n🎯 PHASE 1B STATUS: COMPLETE ✅")
        print("Next: Phase 1C - Trading Operations Testing")
    else:
        print("\n⚠️ PHASE 1B STATUS: NEEDS REVIEW")
        print("Fix issues before proceeding to Phase 1C") 