"""
Quick test script to verify TWS connection
Run this to ensure your setup is working
"""
import asyncio
import sys
from ib_insync import IB, util


async def test_connection():
    """Test basic TWS connection"""
    ib = IB()
    
    try:
        # Connect to TWS
        print("🔌 Attempting to connect to TWS...")
        await ib.connectAsync('host.docker.internal', 7497, clientId=999)
        
        print("✅ Connected successfully!")
        
        # Request current time
        server_time = ib.reqCurrentTime()
        print(f"📅 Server time: {server_time}")
        
        # Get account info
        accounts = ib.managedAccounts()
        print(f"📊 Managed accounts: {accounts}")
        
        # Check connection details
        print(f"🔍 Connection info:")
        print(f"   - Host: {ib.client.host}")
        print(f"   - Port: {ib.client.port}")
        print(f"   - Client ID: {ib.client.clientId}")
        
        return True
        
    except Exception as e:
        print(f"❌ Connection failed: {e}")
        print("\n🔧 Troubleshooting tips:")
        print("1. Ensure TWS is running")
        print("2. Check 'Enable ActiveX and Socket Clients' in TWS")
        print("3. Verify 'Read-Only API' is disabled")
        print("4. Confirm TWS is on port 7497 (paper) or 7496 (live)")
        return False
        
    finally:
        if ib.isConnected():
            ib.disconnect()
            print("👋 Disconnected from TWS")


if __name__ == "__main__":
    # Setup ib-insync for async
    util.startLoop()
    
    # Run test
    success = asyncio.run(test_connection())
    sys.exit(0 if success else 1)