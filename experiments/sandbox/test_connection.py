#!/usr/bin/env python3
"""
Simple TWS Connection Test

This is our "Hello World" for the IBKR API.
Run this to verify your TWS is set up correctly.

Prerequisites:
1. TWS running and logged in
2. API connections enabled in TWS settings
3. ib-insync installed: pip install ib-insync
"""
import asyncio
from datetime import datetime
from ib_insync import IB, util
import logging

# Set up logging to see what's happening
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)


async def test_basic_connection():
    """Test basic connection to TWS"""
    print("\n=== Basic Connection Test ===")
    
    ib = IB()
    try:
        # Connect to TWS
        await ib.connectAsync('localhost', 7497, clientId=999)
        print(f"✅ Connected: {ib.isConnected()}")
        
        # Get account info
        account = ib.client.account
        print(f"📊 Account: {account}")
        
        # Check server time
        server_time = ib.reqCurrentTime()
        print(f"🕐 Server time: {datetime.fromtimestamp(server_time)}")
        
        # Wait a bit to see if we get any events
        print("\n⏳ Listening for events for 5 seconds...")
        await ib.sleep(5)
        
    except Exception as e:
        print(f"❌ Connection failed: {e}")
    finally:
        if ib.isConnected():
            ib.disconnect()
            print("🔌 Disconnected")


async def test_event_system():
    """Test the event system"""
    print("\n=== Event System Test ===")
    
    ib = IB()
    events_received = []
    
    # Set up event handlers
    def on_connected():
        events_received.append('connected')
        print("📡 Event: Connected!")
        
    def on_error(reqId, errorCode, errorString, contract):
        events_received.append(f'error_{errorCode}')
        print(f"⚠️  Event: Error {errorCode} - {errorString}")
        
    # Wire up events
    ib.connectedEvent += on_connected
    ib.errorEvent += on_error
    
    try:
        await ib.connectAsync('localhost', 7497, clientId=998)
        await ib.sleep(2)
        
        print(f"\n📊 Events received: {events_received}")
        
    finally:
        if ib.isConnected():
            ib.disconnect()


async def test_market_data():
    """Test market data subscription"""
    print("\n=== Market Data Test ===")
    
    ib = IB()
    try:
        await ib.connectAsync('localhost', 7497, clientId=997)
        
        # Create a simple stock contract
        from ib_insync import Stock
        contract = Stock('AAPL', 'SMART', 'USD')
        
        # Qualify the contract (get full details)
        await ib.qualifyContractsAsync(contract)
        print(f"📈 Contract qualified: {contract}")
        
        # Request market data
        ticker = ib.reqMktData(contract, '', False, False)
        
        # Wait for some ticks
        print("⏳ Waiting for market data...")
        for i in range(5):
            await ib.sleep(1)
            print(f"   Bid: {ticker.bid}, Ask: {ticker.ask}, Last: {ticker.last}")
            
        # Cancel market data
        ib.cancelMktData(contract)
        
    except Exception as e:
        print(f"❌ Market data test failed: {e}")
    finally:
        if ib.isConnected():
            ib.disconnect()


async def run_all_tests():
    """Run all connection tests"""
    print("🚀 Starting IBKR Connection Tests")
    print("=" * 50)
    
    await test_basic_connection()
    await test_event_system()
    
    # Only test market data if basic tests pass
    print("\n❓ Run market data test? (requires market data subscription)")
    print("   Press Enter to skip, or type 'yes' to run: ", end='')
    
    # Note: In a real async app, we'd handle input differently
    # For now, we'll skip the market data test in automation
    # response = input().strip().lower()
    # if response == 'yes':
    #     await test_market_data()
    
    print("\n✅ All tests completed!")
    print("\n💡 Next steps:")
    print("   1. Check TWS settings if connection failed")
    print("   2. Try experiments/sandbox/test_watchdog.py")
    print("   3. Explore event patterns in .vibe/templates/")


if __name__ == '__main__':
    # This is the proper way to run async code with ib-insync
    util.run(run_all_tests())