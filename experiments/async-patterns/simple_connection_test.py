"""
What I'm testing: Basic async connection to TWS
Expected outcome: Connect, get account info, disconnect cleanly
Actual result: TBD
Learnings: TBD
"""

import asyncio
from ib_insync import IB, util
import logging

logging.basicConfig(level=logging.INFO)

async def test_basic_connection():
    """First test - can we connect at all?"""
    
    ib = IB()
    
    try:
        # Try to connect to paper trading
        print("🔌 Attempting connection to TWS...")
        await ib.connectAsync('localhost', 7497, clientId=999)
        
        print(f"✅ Connected: {ib.isConnected()}")
        print(f"📊 Client ID: {ib.client.clientId}")
        print(f"🔧 Server Version: {ib.serverVersion()}")
        
        # Get account info
        account_values = await ib.accountValuesAsync()
        print(f"\n💰 Account values count: {len(account_values)}")
        
        # Get positions
        positions = await ib.positionsAsync()
        print(f"📈 Open positions: {len(positions)}")
        
        # Test the event loop is responsive
        print("\n⏱️ Testing async responsiveness...")
        start = asyncio.get_event_loop().time()
        await ib.sleep(1)
        elapsed = asyncio.get_event_loop().time() - start
        print(f"✓ Sleep test: {elapsed:.3f} seconds")
        
    except Exception as e:
        print(f"❌ Connection failed: {type(e).__name__}: {e}")
        print("\nTroubleshooting:")
        print("1. Is TWS running?")
        print("2. Is API enabled in TWS settings?")
        print("3. Is port 7497 correct for paper trading?")
        print("4. Check TWS Global Configuration > API > Settings")
        
    finally:
        if ib.isConnected():
            print("\n👋 Disconnecting...")
            ib.disconnect()
            print("✅ Disconnected cleanly")

if __name__ == '__main__':
    print("🧪 IB-Insync Basic Connection Test\n")
    util.run(test_basic_connection())
    print("\n✨ Test complete!")