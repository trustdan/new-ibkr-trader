#!/usr/bin/env python3
"""
Test TWS connection with proper async handling.
Tests the basic connectivity to Interactive Brokers TWS.
"""

import asyncio
import logging
from datetime import datetime
from ib_insync import IB, util

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


async def test_connection():
    """Test basic TWS connection and retrieve account information."""
    ib = IB()
    
    try:
        # Connect to TWS (paper trading by default on port 7497)
        logger.info("Connecting to TWS...")
        await ib.connectAsync('127.0.0.1', 7497, clientId=1)
        
        if ib.isConnected():
            logger.info("✓ Successfully connected to TWS!")
            
            # Get account information
            logger.info("\nRetrieving account information...")
            account_values = await ib.accountValuesAsync()
            
            # Display key account metrics
            logger.info("\nAccount Summary:")
            logger.info("-" * 40)
            
            key_metrics = {
                'NetLiquidation': 'Net Liquidation Value',
                'TotalCashValue': 'Total Cash Value',
                'BuyingPower': 'Buying Power',
                'AvailableFunds': 'Available Funds',
                'MaintMarginReq': 'Maintenance Margin',
                'UnrealizedPnL': 'Unrealized P&L',
                'RealizedPnL': 'Realized P&L'
            }
            
            for account_value in account_values:
                if account_value.tag in key_metrics:
                    logger.info(f"{key_metrics[account_value.tag]}: "
                              f"${float(account_value.value):,.2f}")
            
            # Get positions if any
            positions = await ib.positionsAsync()
            logger.info(f"\nActive Positions: {len(positions)}")
            
            if positions:
                logger.info("\nPosition Details:")
                logger.info("-" * 40)
                for pos in positions[:5]:  # Show first 5 positions
                    logger.info(f"{pos.contract.symbol}: "
                              f"{pos.position} @ ${pos.avgCost:.2f}")
            
            # Test market data subscription (SPY as example)
            logger.info("\nTesting market data subscription...")
            from ib_insync import Stock
            
            spy = Stock('SPY', 'SMART', 'USD')
            await ib.qualifyContractsAsync(spy)
            
            ticker = ib.reqMktData(spy)
            await asyncio.sleep(2)  # Wait for data
            
            if ticker.last:
                logger.info(f"✓ Market data working - SPY last price: ${ticker.last}")
            else:
                logger.warning("⚠ Market data not available (markets may be closed)")
            
            # Check TWS time vs local time
            logger.info("\nTime Synchronization Check:")
            tws_time = await ib.reqCurrentTimeAsync()
            local_time = datetime.now()
            time_diff = abs((tws_time - local_time).total_seconds())
            
            logger.info(f"TWS Time: {tws_time}")
            logger.info(f"Local Time: {local_time}")
            logger.info(f"Time Difference: {time_diff:.1f} seconds")
            
            if time_diff > 5:
                logger.warning("⚠ Time difference > 5 seconds - consider syncing")
            else:
                logger.info("✓ Time synchronization OK")
            
            logger.info("\n✅ All connection tests passed!")
            
        else:
            logger.error("Failed to connect to TWS")
            
    except Exception as e:
        logger.error(f"Connection test failed: {e}")
        logger.error("Make sure TWS is running and API connections are enabled")
        logger.error("Check: File -> Global Configuration -> API -> Settings")
        raise
        
    finally:
        if ib.isConnected():
            logger.info("\nDisconnecting from TWS...")
            ib.disconnect()
            logger.info("Disconnected successfully")


def main():
    """Run the connection test."""
    logger.info("Starting TWS Connection Test")
    logger.info("=" * 50)
    
    # Enable asyncio debugging for better error messages
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    
    try:
        # Run the async test
        asyncio.run(test_connection())
    except KeyboardInterrupt:
        logger.info("\nTest interrupted by user")
    except Exception as e:
        logger.error(f"\nTest failed with error: {e}")
        return 1
    
    return 0


if __name__ == "__main__":
    exit(main())