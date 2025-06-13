# Async IB Connection Template
import asyncio
from ib_insync import IB, util
import logging

async def connect_to_tws():
    """Basic async connection pattern with error handling"""
    ib = IB()
    
    try:
        # Connect asynchronously
        await ib.connectAsync(
            host='localhost',  # or 'host.docker.internal' from Docker
            port=7497,        # 7497 for paper, 7496 for live
            clientId=1,       # Unique ID for this connection
            timeout=10        # Connection timeout
        )
        
        logging.info(f"Connected to TWS: {ib.isConnected()}")
        
        # Your async operations here
        await do_something_async(ib)
        
    except Exception as e:
        logging.error(f"Connection failed: {e}")
        raise
    finally:
        # Always disconnect cleanly
        ib.disconnect()

async def do_something_async(ib):
    """Example async operation"""
    # Let the event loop breathe
    await ib.sleep(1)
    
    # Get account info asynchronously
    account = await ib.accountSummaryAsync()
    print(f"Account: {account}")

if __name__ == '__main__':
    # Run with ib_insync's event loop
    util.run(connect_to_tws())