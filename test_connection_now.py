"""
Quick connection test to validate TWS integration.
"""
import asyncio
import sys
import os

# Add the project root to Python path
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

# Direct imports using the ib-insync library for now
from ib_insync import IB, util


async def test_connection():
    """Test basic TWS connection using ib-insync directly."""
    print("🔌 Testing connection to TWS...")
    
    # Connection parameters (matching our config defaults)
    host = "127.0.0.1"
    port = 7497  # Paper trading port
    client_id = 1
    
    ib = IB()
    
    try:
        print(f"📊 Attempting connection to {host}:{port} with client ID {client_id}")
        
        # Test connection
        await ib.connectAsync(host, port, clientId=client_id)
        
        if ib.isConnected():
            print("✅ CONNECTION SUCCESSFUL!")
            
            # Get connection info
            print("\n📋 Connection Details:")
            print(f"  Connected: {ib.isConnected()}")
            print(f"  Client ID: {ib.client.clientId}")
            print(f"  Server Version: {ib.client.serverVersion()}")
            
            # Test a simple API call
            try:
                server_time = ib.reqCurrentTime()
                print(f"  Server Time: {server_time}")
                
                # Test account summary request
                account_summary = ib.accountSummary()
                if account_summary:
                    print(f"  Account Data: {len(account_summary)} items received")
                    print(f"  Account Number: {account_summary[0].account if account_summary else 'N/A'}")
                
            except Exception as e:
                print(f"⚠️ API request failed: {e}")
            
            # Clean disconnect
            ib.disconnect()
            print("👋 Disconnected cleanly")
            
        else:
            print("❌ Connection failed - not connected")
            return False
            
    except Exception as e:
        print(f"💥 Connection test failed: {e}")
        print(f"Error type: {type(e).__name__}")
        if ib.isConnected():
            ib.disconnect()
        return False
    
    return True


if __name__ == "__main__":
    result = asyncio.run(test_connection())
    if result:
        print("\n🎉 Phase 1A Connection Validation: PASSED!")
    else:
        print("\n⚠️ Phase 1A Connection Validation: FAILED!")
        print("Check TWS is running on port 7497 with API enabled") 