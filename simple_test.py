"""
Simple synchronous TWS connection test.
"""
from ib_insync import IB
import time

def test_tws_connection():
    """Test TWS connection synchronously."""
    print("🔌 Testing TWS connection...")
    
    # Connection parameters
    host = "127.0.0.1"
    port = 7497  # Paper trading port
    client_id = 1
    
    ib = IB()
    
    try:
        print(f"📊 Connecting to {host}:{port} (client ID: {client_id})")
        
        # Connect (synchronous)
        ib.connect(host, port, clientId=client_id)
        
        if ib.isConnected():
            print("✅ CONNECTION SUCCESSFUL!")
            
            print("\n📋 Connection Details:")
            print(f"  Connected: {ib.isConnected()}")
            print(f"  Client ID: {ib.client.clientId}")
            print(f"  Server Version: {ib.client.serverVersion()}")
            
            # Test server time
            try:
                server_time = ib.reqCurrentTime()
                print(f"  Server Time: {server_time}")
            except Exception as e:
                print(f"  ⚠️ Server time failed: {e}")
            
            # Test account summary
            try:
                print("  Requesting account summary...")
                account_summary = ib.accountSummary()
                print(f"  Account Items: {len(account_summary)}")
                
                if account_summary:
                    for item in account_summary[:3]:  # Show first 3 items
                        print(f"    {item.tag}: {item.value}")
                        
            except Exception as e:
                print(f"  ⚠️ Account summary failed: {e}")
            
            # Clean disconnect
            ib.disconnect()
            print("👋 Disconnected successfully")
            return True
            
        else:
            print("❌ Connection failed")
            return False
            
    except Exception as e:
        print(f"💥 Connection error: {e}")
        print(f"Error type: {type(e).__name__}")
        
        if ib.isConnected():
            ib.disconnect()
        return False

if __name__ == "__main__":
    print("=" * 50)
    print("🚀 Phase 1A: TWS Connection Validation")
    print("=" * 50)
    
    success = test_tws_connection()
    
    print("\n" + "=" * 50)
    if success:
        print("🎉 CONNECTION TEST PASSED!")
        print("✅ Phase 1A validation complete - ready for integration tests")
    else:
        print("⚠️ CONNECTION TEST FAILED!")
        print("Check that TWS is running with API enabled on port 7497")
    print("=" * 50) 