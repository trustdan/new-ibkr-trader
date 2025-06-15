"""
Direct socket test to TWS - no ib_insync dependencies.
"""
import socket
import time

def test_socket_connection():
    """Test raw socket connection to TWS."""
    host = "127.0.0.1"
    port = 7497
    
    print(f"🔌 Testing socket connection to {host}:{port}")
    
    try:
        # Create socket
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(5)  # 5 second timeout
        
        print("🔄 Connecting...")
        result = s.connect_ex((host, port))
        
        if result == 0:
            print("✅ Socket connection successful!")
            print("📡 TWS is accepting connections")
            s.close()
            return True
        else:
            print(f"❌ Socket connection failed: {result}")
            return False
            
    except socket.error as e:
        print(f"💥 Socket error: {e}")
        return False
    except Exception as e:
        print(f"💥 Unexpected error: {e}")
        return False

def test_telnet_style():
    """Try a telnet-style test."""
    host = "127.0.0.1"
    port = 7497
    
    print(f"\n🔍 Telnet-style test to {host}:{port}")
    
    try:
        import telnetlib
        tn = telnetlib.Telnet(host, port, timeout=3)
        print("✅ Telnet connection established")
        tn.close()
        return True
    except Exception as e:
        print(f"❌ Telnet failed: {e}")
        return False

if __name__ == "__main__":
    print("=" * 50)
    print("🚀 TWS Socket Connectivity Test")
    print("=" * 50)
    
    socket_ok = test_socket_connection()
    telnet_ok = test_telnet_style()
    
    print("\n" + "=" * 50)
    if socket_ok:
        print("🎉 SOCKET TEST PASSED!")
        print("TWS is listening and accepting connections")
        print("The issue is likely with ib_insync library")
    else:
        print("❌ SOCKET TEST FAILED!")
        print("TWS may not be configured correctly")
    print("=" * 50) 