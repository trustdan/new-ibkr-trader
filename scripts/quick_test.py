#!/usr/bin/env python
"""
Quick async test runner for development - runs common test scenarios
"""
import asyncio
import sys
from pathlib import Path
from datetime import datetime
import json

# Add src to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent / 'src' / 'python'))

async def test_basic_connection():
    """Test basic TWS connection"""
    print("🔌 Testing basic connection...")
    try:
        from ib_insync import IB
        ib = IB()
        await ib.connectAsync('localhost', 7497, clientId=998)
        print("✅ Connected successfully")
        ib.disconnect()
        return True
    except Exception as e:
        print(f"❌ Connection failed: {e}")
        return False

async def test_event_system():
    """Test event callbacks"""
    print("\n📡 Testing event system...")
    try:
        from ib_insync import IB
        ib = IB()
        
        events_received = []
        
        def on_error(reqId, errorCode, errorString, contract):
            events_received.append(f"Error: {errorCode} - {errorString}")
        
        ib.errorEvent += on_error
        
        await ib.connectAsync('localhost', 7497, clientId=997)
        await asyncio.sleep(1)  # Wait for any initial events
        
        print(f"✅ Event system working - {len(events_received)} events received")
        ib.disconnect()
        return True
    except Exception as e:
        print(f"❌ Event system test failed: {e}")
        return False

async def test_market_data():
    """Test market data subscription"""
    print("\n📊 Testing market data...")
    try:
        from ib_insync import IB, Stock
        ib = IB()
        await ib.connectAsync('localhost', 7497, clientId=996)
        
        # Request market data for SPY
        contract = Stock('SPY', 'SMART', 'USD')
        ib.qualifyContracts(contract)
        
        ticker = ib.reqMktData(contract)
        await asyncio.sleep(2)  # Wait for data
        
        if ticker.bid and ticker.ask:
            print(f"✅ Market data working - SPY: Bid={ticker.bid}, Ask={ticker.ask}")
            result = True
        else:
            print("❌ No market data received")
            result = False
            
        ib.cancelMktData(contract)
        ib.disconnect()
        return result
    except Exception as e:
        print(f"❌ Market data test failed: {e}")
        return False

async def test_async_patterns():
    """Test various async patterns"""
    print("\n⚡ Testing async patterns...")
    try:
        # Test concurrent operations
        tasks = []
        for i in range(3):
            tasks.append(asyncio.create_task(asyncio.sleep(0.1)))
        
        await asyncio.gather(*tasks)
        print("✅ Concurrent operations working")
        
        # Test event loop
        loop = asyncio.get_running_loop()
        print(f"✅ Event loop running: {type(loop).__name__}")
        
        return True
    except Exception as e:
        print(f"❌ Async pattern test failed: {e}")
        return False

async def save_test_results(results):
    """Save test results to vibe folder"""
    report = {
        'timestamp': datetime.now().isoformat(),
        'tests': results,
        'summary': {
            'total': len(results),
            'passed': sum(1 for r in results.values() if r),
            'failed': sum(1 for r in results.values() if not r)
        }
    }
    
    report_path = Path('.vibe/test_results.json')
    report_path.parent.mkdir(exist_ok=True)
    
    with open(report_path, 'w') as f:
        json.dump(report, f, indent=2)
    
    print(f"\n📄 Test results saved to: {report_path}")

async def main():
    """Run all quick tests"""
    print("🚀 IBKR Quick Test Suite")
    print("=" * 40)
    
    # Set up Windows event loop if needed
    if sys.platform == 'win32':
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    
    results = {}
    
    # Run tests
    results['connection'] = await test_basic_connection()
    results['events'] = await test_event_system()
    results['market_data'] = await test_market_data()
    results['async_patterns'] = await test_async_patterns()
    
    # Summary
    print("\n📊 Test Summary:")
    print("-" * 40)
    for test, passed in results.items():
        status = "✅ PASS" if passed else "❌ FAIL"
        print(f"{test:.<30} {status}")
    
    # Save results
    await save_test_results(results)
    
    # Exit code
    all_passed = all(results.values())
    if all_passed:
        print("\n✨ All tests passed!")
        return 0
    else:
        print("\n⚠️  Some tests failed!")
        return 1

if __name__ == '__main__':
    exit_code = asyncio.run(main())
    sys.exit(exit_code)