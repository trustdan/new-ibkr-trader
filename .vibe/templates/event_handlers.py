# Event Handler Patterns for ib-insync
from ib_insync import IB, Contract, Order, Trade
import logging

class EventDrivenTrader:
    """Template for event-driven trading with ib-insync"""
    
    def __init__(self):
        self.ib = IB()
        self._setup_event_handlers()
    
    def _setup_event_handlers(self):
        """Wire up all event handlers - the heart of async trading"""
        
        # Connection events
        self.ib.connectedEvent += self._on_connected
        self.ib.disconnectedEvent += self._on_disconnected
        
        # Order events
        self.ib.orderStatusEvent += self._on_order_status
        self.ib.execDetailsEvent += self._on_exec_details
        self.ib.commissionReportEvent += self._on_commission
        
        # Market data events
        self.ib.pendingTickersEvent += self._on_pending_tickers
        self.ib.barUpdateEvent += self._on_bar_update
        
        # Error events
        self.ib.errorEvent += self._on_error
        
        # Position events
        self.ib.positionEvent += self._on_position
        self.ib.pnlEvent += self._on_pnl
        
        # News events
        self.ib.newsBulletinEvent += self._on_news
        
    # Connection handlers
    def _on_connected(self):
        """Called when connection is established"""
        logging.info("🟢 Connected to TWS")
        # Initialize any startup tasks here
        
    def _on_disconnected(self):
        """Called when connection is lost"""
        logging.warning("🔴 Disconnected from TWS")
        # Handle cleanup or reconnection logic
        
    # Order handlers
    def _on_order_status(self, trade: Trade):
        """Real-time order status updates"""
        logging.info(f"Order {trade.order.orderId}: {trade.orderStatus.status}")
        
        if trade.orderStatus.status == 'Filled':
            self._handle_fill(trade)
        elif trade.orderStatus.status in ['Cancelled', 'ApiCancelled']:
            self._handle_cancellation(trade)
            
    def _on_exec_details(self, trade: Trade, fill):
        """Execution details for filled orders"""
        logging.info(f"Execution: {fill.contract.symbol} "
                    f"{fill.execution.shares}@{fill.execution.price}")
        
    def _on_commission(self, trade: Trade, fill, report):
        """Commission reports"""
        logging.info(f"Commission: ${report.commission} "
                    f"({report.currency})")
        
    # Market data handlers
    def _on_pending_tickers(self, tickers):
        """Handle streaming market data"""
        for ticker in tickers:
            if ticker.last is not None:
                # Process real-time price updates
                self._process_price_update(ticker)
                
    def _on_bar_update(self, bars, has_new_bar):
        """Handle real-time bar updates"""
        if has_new_bar:
            latest_bar = bars[-1]
            logging.debug(f"New bar: {latest_bar}")
            
    # Error handler
    def _on_error(self, reqId, errorCode, errorString, contract):
        """Central error handling"""
        
        # Common error codes to handle
        if errorCode == 1100:
            logging.error("Connectivity lost - will auto-reconnect")
        elif errorCode == 100:
            logging.warning("Pacing violation - slow down requests!")
        elif errorCode == 2110:
            logging.info("Connectivity restored")
        elif errorCode == 502:
            logging.error("TWS not started")
        else:
            logging.error(f"Error {errorCode}: {errorString}")
            
    # Helper methods
    def _handle_fill(self, trade: Trade):
        """Process filled orders"""
        logging.info(f"✅ Order filled: {trade}")
        # Add post-fill logic here
        
    def _handle_cancellation(self, trade: Trade):
        """Process cancelled orders"""
        logging.info(f"❌ Order cancelled: {trade}")
        # Add cancellation logic here
        
    def _process_price_update(self, ticker):
        """Process real-time price updates"""
        # Add your price processing logic here
        pass
        
    def _on_position(self, position):
        """Handle position updates"""
        logging.info(f"Position update: {position}")
        
    def _on_pnl(self, pnl):
        """Handle P&L updates"""
        logging.info(f"P&L update: ${pnl.realizedPnL:.2f} realized, "
                    f"${pnl.unrealizedPnL:.2f} unrealized")
        
    def _on_news(self, msgId, msgType, newsMessage, originExch):
        """Handle news bulletins"""
        logging.info(f"News: {newsMessage}")

# Example usage
async def main():
    trader = EventDrivenTrader()
    await trader.ib.connectAsync('localhost', 7497, clientId=1)
    
    # The event handlers will now process all updates
    # Keep the connection alive
    await trader.ib.sleep(3600)  # Run for 1 hour
    
if __name__ == '__main__':
    util.run(main())