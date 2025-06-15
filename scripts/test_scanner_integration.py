#!/usr/bin/env python3
"""
Quick test script for Python-Go scanner integration

This script validates that the Python service can communicate with the Go scanner.
"""

import asyncio
import sys
import logging
from datetime import datetime

# Add src to path
sys.path.insert(0, '/home/kali/new-ibkr-trader')

from src.python.scanner_client import (
    ScannerClient, ScanRequest, ScanFilter, FilterType
)
from src.python.integration.scanner_coordinator import ScannerCoordinator
from src.python.integration.backpressure import BackpressureStrategy

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class MockIBKR:
    """Mock IBKR connection for testing"""
    async def emit_event(self, event):
        logger.info(f"Event: {event.type} - {event.data}")


async def test_scanner_client():
    """Test direct scanner client communication"""
    print("\n=== Testing Scanner Client ===")
    
    async with ScannerClient() as client:
        # Test health check
        print("\n1. Testing health check...")
        try:
            health = await client.health_check()
            print(f"✓ Scanner health: {health}")
        except Exception as e:
            print(f"✗ Health check failed: {e}")
            return False
            
        # Test scan
        print("\n2. Testing scan request...")
        request = ScanRequest(
            symbol="SPY",
            filters=[
                ScanFilter(
                    type=FilterType.DELTA,
                    params={"min": 0.25, "max": 0.35}
                ),
                ScanFilter(
                    type=FilterType.DTE,
                    params={"min": 30, "max": 60}
                )
            ],
            limit=5
        )
        
        try:
            spreads = await client.scan(request)
            print(f"✓ Found {len(spreads)} spreads")
            
            if spreads:
                spread = spreads[0]
                print(f"\nTop spread:")
                print(f"  Score: {spread.score:.2f}")
                print(f"  Long: {spread.long_leg.strike} @ {spread.long_leg.delta:.2f}Δ")
                print(f"  Short: {spread.short_leg.strike} @ {spread.short_leg.delta:.2f}Δ")
                print(f"  Net Debit: ${spread.net_debit:.2f}")
                print(f"  Max Profit: ${spread.max_profit:.2f}")
                print(f"  PoP: {spread.probability_profit:.1%}")
                
        except Exception as e:
            print(f"✗ Scan failed: {e}")
            return False
            
    return True


async def test_coordinator():
    """Test scanner coordinator with backpressure"""
    print("\n=== Testing Scanner Coordinator ===")
    
    mock_ibkr = MockIBKR()
    
    async with ScannerClient() as scanner_client:
        coordinator = ScannerCoordinator(
            ibkr_connection=mock_ibkr,
            scanner_client=scanner_client,
            backpressure_strategy=BackpressureStrategy.ADAPTIVE
        )
        
        await coordinator.start()
        
        try:
            # Test single scan
            print("\n1. Testing single scan with caching...")
            filters = [
                ScanFilter(
                    type=FilterType.DELTA,
                    params={"min": 0.3, "max": 0.4}
                ),
                ScanFilter(
                    type=FilterType.LIQUIDITY,
                    params={
                        "min_volume": 100,
                        "min_open_interest": 1000,
                        "max_bid_ask_spread": 0.10
                    }
                )
            ]
            
            start = datetime.now()
            spreads1 = await coordinator.scan_symbol("AAPL", filters)
            time1 = (datetime.now() - start).total_seconds()
            print(f"✓ First scan: {len(spreads1)} spreads in {time1:.2f}s")
            
            # Test cache
            start = datetime.now()
            spreads2 = await coordinator.scan_symbol("AAPL", filters)
            time2 = (datetime.now() - start).total_seconds()
            print(f"✓ Cached scan: {len(spreads2)} spreads in {time2:.2f}s")
            
            # Test concurrent scans
            print("\n2. Testing concurrent scans...")
            symbols = ["MSFT", "GOOGL", "AMZN", "TSLA", "META"]
            
            tasks = [
                coordinator.scan_symbol(symbol, filters)
                for symbol in symbols
            ]
            
            start = datetime.now()
            results = await asyncio.gather(*tasks, return_exceptions=True)
            total_time = (datetime.now() - start).total_seconds()
            
            successful = sum(1 for r in results if not isinstance(r, Exception))
            print(f"✓ Concurrent scans: {successful}/{len(symbols)} successful in {total_time:.2f}s")
            
            # Show metrics
            print("\n3. Coordinator Metrics:")
            metrics = coordinator.get_metrics()
            print(f"  Total scans: {metrics['total_scans']}")
            print(f"  Successful: {metrics['successful_scans']}")
            print(f"  Failed: {metrics['failed_scans']}")
            print(f"  Cache hits: {metrics['cache_hits']}")
            print(f"  Avg scan time: {metrics['average_scan_time']:.2f}s")
            
            print("\n4. Backpressure Metrics:")
            bp_metrics = metrics['backpressure']
            print(f"  Strategy: {bp_metrics['strategy']}")
            print(f"  Current QPS: {bp_metrics['current_qps']}")
            print(f"  Avg response time: {bp_metrics['average_response_time']}s")
            print(f"  Circuit breaker: {'OPEN' if bp_metrics['circuit_breaker_open'] else 'CLOSED'}")
            
        finally:
            await coordinator.stop()
            
    return True


async def test_backpressure():
    """Test backpressure handling under load"""
    print("\n=== Testing Backpressure Under Load ===")
    
    mock_ibkr = MockIBKR()
    
    async with ScannerClient() as scanner_client:
        coordinator = ScannerCoordinator(
            ibkr_connection=mock_ibkr,
            scanner_client=scanner_client,
            max_concurrent_scans=3,
            backpressure_strategy=BackpressureStrategy.TOKEN_BUCKET
        )
        
        await coordinator.start()
        
        try:
            # Generate high load
            print("\n1. Generating high load (20 requests)...")
            filters = [
                ScanFilter(FilterType.DELTA, {"min": 0.25, "max": 0.35})
            ]
            
            tasks = []
            for i in range(20):
                symbol = f"TEST{i}"
                task = coordinator.scan_symbol(symbol, filters, use_cache=False)
                tasks.append(task)
                
            start = datetime.now()
            results = await asyncio.gather(*tasks, return_exceptions=True)
            total_time = (datetime.now() - start).total_seconds()
            
            successful = sum(1 for r in results if not isinstance(r, Exception))
            failed = len(results) - successful
            
            print(f"✓ Load test complete in {total_time:.2f}s")
            print(f"  Successful: {successful}")
            print(f"  Failed: {failed}")
            print(f"  Requests/sec: {len(results) / total_time:.2f}")
            
            # Check backpressure metrics
            metrics = coordinator.get_metrics()
            bp = metrics['backpressure']
            print(f"\n2. Backpressure handled:")
            print(f"  Tokens available: {bp['tokens_available']}")
            print(f"  Queue size: {bp['queue_size']}")
            
        finally:
            await coordinator.stop()
            
    return True


async def main():
    """Run all integration tests"""
    print("Python-Go Scanner Integration Test")
    print("==================================")
    
    # Check if scanner is running
    print("\nChecking scanner availability...")
    try:
        async with ScannerClient() as client:
            health = await client.health_check()
            if health.get("status") != "healthy":
                print("✗ Scanner service is not healthy")
                print("  Please start the Go scanner with: cd src/go && go run cmd/scanner/main.go")
                return 1
    except Exception as e:
        print(f"✗ Scanner service is not available: {e}")
        print("  Please start the Go scanner with: cd src/go && go run cmd/scanner/main.go")
        return 1
        
    print("✓ Scanner service is available")
    
    # Run tests
    tests = [
        ("Scanner Client", test_scanner_client),
        ("Scanner Coordinator", test_coordinator),
        ("Backpressure Handling", test_backpressure)
    ]
    
    passed = 0
    for name, test_func in tests:
        try:
            if await test_func():
                passed += 1
                print(f"\n✓ {name} test passed")
            else:
                print(f"\n✗ {name} test failed")
        except Exception as e:
            print(f"\n✗ {name} test failed with exception: {e}")
            logger.exception(f"Test {name} failed")
            
    print(f"\n{'='*50}")
    print(f"Test Summary: {passed}/{len(tests)} passed")
    
    return 0 if passed == len(tests) else 1


if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)