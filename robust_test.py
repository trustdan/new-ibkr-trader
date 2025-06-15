"""
Robust TWS connection test with timeouts and detailed error reporting.
"""
import sys
import signal
import time
from ib_insync import IB

class TimeoutError(Exception):
    pass

def timeout_handler(signum, frame):
    raise TimeoutError("Connection attempt timed out")

def test_tws_robust():
    """Test TWS connection with robust error handling."""
    print("🔌 Robust TWS Connection Test")
    print("=" * 40)
    
    # Connection parameters
    host = "127.0.0.1"
    port = 7497
    client_id = 1
    
    print(f"Target: {host}:{port} (Client ID: {client_id})")
    
    ib = IB()
    
    try:
        # Set a 10-second timeout for connection
        if sys.platform != 'win32':
            signal.signal(signal.SIGALRM, timeout_handler)
            signal.alarm(10)
        
        print("🔄 Attempting connection...")
        start_time = time.time()
        
        # Try to connect
        ib.connect(host, port, clientId=client_id, timeout=10)
        
        connection_time = time.time() - start_time
        print(f"⏱️  Connection took {connection_time:.2f} seconds")
        
        if ib.isConnected():
            print("✅ CONNECTION SUCCESSFUL!")
            
            # Basic info
            print(f"📊 Server Version: {ib.client.serverVersion()}")
            print(f"📊 Connection Time: {ib.client.connectionTime()}")
            
            # Quick API test
            try:
                print("🔍 Testing server time request...")
                server_time = ib.reqCurrentTime()
                print(f"⏰ Server Time: {server_time}")
            except Exception as e:
                print(f"⚠️ Server time failed: {e}")
            
            # Test with a very simple contract request
            try:
                print("🔍 Testing contract details...")
                from ib_insync import Stock
                contract = Stock('AAPL', 'SMART', 'USD')
                details = ib.reqContractDetails(contract)
                if details:
                    print(f"📈 Contract test passed: Found {len(details)} AAPL contracts")
                else:
                    print("⚠️ No contract details returned")
            except Exception as e:
                print(f"⚠️ Contract test failed: {e}")
            
            print("👋 Disconnecting...")
            ib.disconnect()
            print("✅ Test completed successfully!")
            return True
        else:
            print("❌ Connection failed - not connected")
            return False
            
    except TimeoutError:
        print("⏰ Connection timed out after 10 seconds")
        return False
    except ConnectionRefusedError:
        print("🚫 Connection refused - TWS not accepting connections")
        print("💡 Check TWS API settings: File → Global Configuration → API → Settings")
        return False
    except Exception as e:
        print(f"💥 Unexpected error: {e}")
        print(f"Error type: {type(e).__name__}")
        return False
    finally:
        if sys.platform != 'win32':
            signal.alarm(0)  # Cancel alarm
        if ib.isConnected():
            try:
                ib.disconnect()
            except:
                pass

if __name__ == "__main__":
    print("Ignoring numpy warnings... (these are harmless)")
    
    success = test_tws_robust()
    
    print("\n" + "=" * 40)
    if success:
        print("🎉 TWS CONNECTION VALIDATED!")
        print("Ready for Phase 1 integration tests")
    else:
        print("❌ CONNECTION FAILED")
        print("\nTroubleshooting checklist:")
        print("1. TWS is running and logged in")
        print("2. API settings enabled (File → Global Config → API)")
        print("3. Socket port set to 7497")
        print("4. 'Enable ActiveX and Socket Clients' checked")
        print("5. 'Read-Only API' unchecked")
    print("=" * 40) 