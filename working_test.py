"""
Working TWS test with numpy warning suppression and proper setup.
"""
import warnings
import sys

# Suppress all numpy/pandas warnings
warnings.filterwarnings('ignore')
import numpy as np
warnings.filterwarnings('ignore', category=RuntimeWarning)

# Set up proper event loop for Windows
import asyncio
if sys.platform == 'win32':
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())

print("🔌 Starting TWS Connection Test")
print("=" * 40)

try:
    print("📦 Importing ib_insync...")
    from ib_insync import IB, util
    print("✅ ib_insync imported successfully")
    
    # Create IB instance
    ib = IB()
    print("✅ IB instance created")
    
    # Connection parameters
    host = "127.0.0.1"
    port = 7497
    client_id = 2  # Different client ID to avoid conflicts
    
    print(f"🔄 Connecting to {host}:{port} (client {client_id})...")
    
    # Use a simple synchronous connection
    ib.connect(host, port, clientId=client_id)
    
    print(f"📊 Connection status: {ib.isConnected()}")
    
    if ib.isConnected():
        print("🎉 CONNECTION SUCCESSFUL!")
        
        # Test basic API functionality
        try:
            print("🔍 Testing server version...")
            version = ib.client.serverVersion()
            print(f"  Server Version: {version}")
            
            print("🔍 Testing current time...")
            current_time = ib.reqCurrentTime()
            print(f"  Server Time: {current_time}")
            
            print("🔍 Testing account data...")
            account_summary = ib.accountSummary()
            print(f"  Account Summary: {len(account_summary)} items")
            
            if account_summary:
                for i, item in enumerate(account_summary[:3]):
                    print(f"    {item.tag}: {item.value}")
                    
        except Exception as e:
            print(f"⚠️ API test failed: {e}")
        
        print("👋 Disconnecting...")
        ib.disconnect()
        print("✅ Disconnected successfully")
        
        print("\n🎉 PHASE 1A CONNECTION VALIDATION: PASSED!")
        print("✅ Ready to proceed with integration tests")
        
    else:
        print("❌ Connection failed")
        
except Exception as e:
    print(f"💥 Error: {e}")
    print(f"Error type: {type(e).__name__}")
    import traceback
    traceback.print_exc()

print("\n" + "=" * 40)
print("🏁 Test completed")
print("=" * 40) 