# Order Execution Patterns for Vertical Spreads
from ib_insync import IB, Contract, Order, ComboLeg, TagValue, util
import asyncio
import logging
from datetime import datetime

class VerticalSpreadTrader:
    """Template for executing vertical spread combo orders"""
    
    def __init__(self, ib: IB):
        self.ib = ib
        self.active_orders = {}
        self._setup_order_events()
        
    def _setup_order_events(self):
        """Set up order event handlers"""
        self.ib.orderStatusEvent += self._on_order_status
        self.ib.execDetailsEvent += self._on_execution
        self.ib.commissionReportEvent += self._on_commission
        
    async def create_vertical_spread(
        self,
        symbol: str,
        expiry: str,  # YYYYMMDD format
        long_strike: float,
        short_strike: float,
        right: str = 'C',  # 'C' for call, 'P' for put
        exchange: str = 'SMART'
    ):
        """Create a vertical spread combo contract"""
        
        # Create the leg contracts
        long_leg = Option(symbol, expiry, long_strike, right, exchange)
        short_leg = Option(symbol, expiry, short_strike, right, exchange)
        
        # Qualify the contracts to get conIds
        qualified = await self.ib.qualifyContractsAsync(long_leg, short_leg)
        
        if len(qualified) != 2:
            raise ValueError("Failed to qualify option contracts")
            
        long_leg, short_leg = qualified
        
        # Create combo contract
        combo = Contract()
        combo.symbol = symbol
        combo.secType = 'BAG'
        combo.currency = 'USD'
        combo.exchange = exchange
        
        # Define combo legs
        leg1 = ComboLeg()
        leg1.conId = long_leg.conId
        leg1.ratio = 1
        leg1.action = 'BUY'
        leg1.exchange = exchange
        
        leg2 = ComboLeg()
        leg2.conId = short_leg.conId
        leg2.ratio = 1
        leg2.action = 'SELL'
        leg2.exchange = exchange
        
        combo.comboLegs = [leg1, leg2]
        
        return combo, long_leg, short_leg
        
    async def execute_vertical_spread(
        self,
        symbol: str,
        expiry: str,
        long_strike: float,
        short_strike: float,
        right: str = 'C',
        quantity: int = 1,
        limit_price: float = None,
        preview_only: bool = False
    ):
        """Execute a vertical spread order"""
        
        # Create the combo contract
        combo, long_leg, short_leg = await self.create_vertical_spread(
            symbol, expiry, long_strike, short_strike, right
        )
        
        # Determine order action (debit or credit spread)
        is_debit = long_strike < short_strike if right == 'C' else long_strike > short_strike
        
        # Create order
        order = Order()
        order.action = 'BUY' if is_debit else 'SELL'
        order.totalQuantity = quantity
        order.orderType = 'LMT' if limit_price else 'MKT'
        
        if limit_price:
            order.lmtPrice = limit_price
            
        # Add smart routing
        order.smartComboRoutingParams = [
            TagValue('NonGuaranteed', '1')
        ]
        
        # Preview order with whatIf
        if preview_only:
            order.whatIf = True
            preview = await self.ib.whatIfOrderAsync(combo, order)
            return self._format_preview(preview, combo, order)
            
        # Place the order
        trade = await self.ib.placeOrderAsync(combo, order)
        self.active_orders[trade.order.orderId] = trade
        
        logging.info(f"Placed {order.action} {quantity} "
                    f"{symbol} {long_strike}/{short_strike} "
                    f"{right} spread @ {limit_price or 'MKT'}")
        
        return trade
        
    def _format_preview(self, preview, combo, order):
        """Format order preview results"""
        return {
            'action': order.action,
            'quantity': order.totalQuantity,
            'commission': float(preview.commission),
            'margin': {
                'initial': float(preview.initMarginChange),
                'maintenance': float(preview.maintMarginChange)
            },
            'equity': {
                'before': float(preview.equityWithLoanBefore),
                'after': float(preview.equityWithLoanAfter)
            },
            'warning': preview.warningText
        }
        
    async def modify_order(self, order_id: int, new_limit_price: float):
        """Modify an existing order's limit price"""
        
        if order_id not in self.active_orders:
            raise ValueError(f"Order {order_id} not found")
            
        trade = self.active_orders[order_id]
        
        # Only modify if order is still active
        if trade.orderStatus.status in ['PreSubmitted', 'Submitted']:
            trade.order.lmtPrice = new_limit_price
            await self.ib.placeOrderAsync(trade.contract, trade.order)
            logging.info(f"Modified order {order_id} limit price to {new_limit_price}")
        else:
            logging.warning(f"Cannot modify order {order_id} - status: {trade.orderStatus.status}")
            
    async def cancel_order(self, order_id: int):
        """Cancel an active order"""
        
        if order_id not in self.active_orders:
            raise ValueError(f"Order {order_id} not found")
            
        trade = self.active_orders[order_id]
        self.ib.cancelOrder(trade.order)
        logging.info(f"Cancelled order {order_id}")
        
    def _on_order_status(self, trade):
        """Handle order status updates"""
        status = trade.orderStatus.status
        order_id = trade.order.orderId
        
        logging.info(f"Order {order_id} status: {status}")
        
        if status == 'Filled':
            logging.info(f"✅ Order {order_id} filled at "
                        f"{trade.orderStatus.avgFillPrice}")
            # Remove from active orders
            self.active_orders.pop(order_id, None)
            
        elif status in ['Cancelled', 'ApiCancelled']:
            logging.info(f"❌ Order {order_id} cancelled")
            self.active_orders.pop(order_id, None)
            
    def _on_execution(self, trade, fill):
        """Handle execution details"""
        logging.info(f"Execution: {fill.execution.side} "
                    f"{fill.execution.shares} "
                    f"{fill.contract.localSymbol} "
                    f"@ {fill.execution.price}")
        
    def _on_commission(self, trade, fill, report):
        """Handle commission reports"""
        logging.info(f"Commission: ${report.commission:.2f}")

# Example: Order management with stop loss
class SpreadOrderManager:
    """Advanced order management for spreads"""
    
    def __init__(self, ib: IB):
        self.ib = ib
        self.positions = {}
        
    async def enter_spread_with_stop(
        self,
        spread_params: dict,
        stop_loss_pct: float = 0.5  # 50% of max profit
    ):
        """Enter spread with automatic stop loss"""
        
        trader = VerticalSpreadTrader(self.ib)
        
        # Place the spread order
        trade = await trader.execute_vertical_spread(**spread_params)
        
        # Wait for fill
        while trade.orderStatus.status not in ['Filled', 'Cancelled', 'ApiCancelled']:
            await asyncio.sleep(0.1)
            
        if trade.orderStatus.status == 'Filled':
            # Calculate stop loss level
            fill_price = trade.orderStatus.avgFillPrice
            
            # Create stop order (opposite of entry)
            stop_order = Order()
            stop_order.action = 'SELL' if trade.order.action == 'BUY' else 'BUY'
            stop_order.totalQuantity = trade.order.totalQuantity
            stop_order.orderType = 'STP'
            stop_order.auxPrice = fill_price * (1 - stop_loss_pct)
            
            # Place stop order
            stop_trade = await self.ib.placeOrderAsync(trade.contract, stop_order)
            
            logging.info(f"Placed stop loss at {stop_order.auxPrice}")
            
            return trade, stop_trade
            
        return trade, None

# Example usage
async def example_spread_execution():
    """Example of executing a vertical spread"""
    
    ib = IB()
    await ib.connectAsync('localhost', 7497, clientId=1)
    
    trader = VerticalSpreadTrader(ib)
    
    try:
        # Preview a call spread
        preview = await trader.execute_vertical_spread(
            symbol='SPY',
            expiry='20240220',  # Adjust to valid expiry
            long_strike=450,
            short_strike=455,
            right='C',
            quantity=1,
            limit_price=2.50,
            preview_only=True
        )
        
        print(f"Order Preview: {preview}")
        
        # Execute if preview looks good
        if preview['margin']['initial'] < 5000:  # Check margin requirement
            trade = await trader.execute_vertical_spread(
                symbol='SPY',
                expiry='20240220',
                long_strike=450,
                short_strike=455,
                right='C',
                quantity=1,
                limit_price=2.50
            )
            
            # Monitor until filled or cancelled
            while trade.orderStatus.status in ['PreSubmitted', 'Submitted']:
                await asyncio.sleep(1)
                print(f"Order status: {trade.orderStatus.status}")
                
    except Exception as e:
        logging.error(f"Error: {e}")
    finally:
        ib.disconnect()

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    util.run(example_spread_execution())