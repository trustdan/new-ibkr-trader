"""
Simple test to verify TWS connectivity without full service startup
"""
import sys
import os
import time

# Add project root to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

print("🚀 Starting simple TWS connectivity test...")

try:
    print("📦 Testing ib_insync import...")
    from ib_insync import IB
    print("✅ ib_insync imported successfully")
    
    print("🔌 Attempting TWS connection...")
    ib = IB()
    
    # Try to connect to TWS
    try:
        ib.connect('127.0.0.1', 7497, clientId=999)
        print(f"✅ Connected to TWS: {ib.isConnected()}")
        
        # Get account info
        accounts = ib.managedAccounts()
        print(f"📊 Available accounts: {accounts}")
        
        # Disconnect
        ib.disconnect()
        print("👋 Disconnected successfully")
        
    except Exception as e:
        print(f"❌ Connection failed: {e}")
        print("💡 Make sure TWS is running with API enabled on port 7497")
        
except Exception as e:
    print(f"❌ Import failed: {e}")
    print("💡 This might be the numpy compatibility issue")

print("🏁 Test complete") 