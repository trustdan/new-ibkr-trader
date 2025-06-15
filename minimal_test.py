"""
Minimal TWS test - just connection, no fancy features.
"""
import os
os.environ['PYTHONWARNINGS'] = 'ignore'  # Suppress warnings

print("Starting minimal test...")

try:
    from ib_insync import IB
    print("✅ ib_insync imported")
    
    ib = IB()
    print("✅ IB object created")
    
    print("🔄 Attempting connection to 127.0.0.1:7497...")
    ib.connect('127.0.0.1', 7497, clientId=1)
    
    print(f"Connected: {ib.isConnected()}")
    
    if ib.isConnected():
        print("🎉 SUCCESS!")
        ib.disconnect()
        print("Disconnected")
    else:
        print("❌ Failed to connect")

except Exception as e:
    print(f"Error: {e}")
    print(f"Type: {type(e)}")

print("Test complete.") 