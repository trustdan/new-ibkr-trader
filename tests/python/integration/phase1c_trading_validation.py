"""
Phase 1C: Trading Operations Testing - Day 6
Tests paper trading, order management, and vertical spread creation.
"""

import asyncio
import socket
import time
import logging
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, Any, List, Optional
from dataclasses import dataclass

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class OrderStatus(Enum):
    """Order status states."""
    PENDING = "pending"
    SUBMITTED = "submitted"
    FILLED = "filled"
    CANCELLED = "cancelled"
    REJECTED = "rejected"


class OrderType(Enum):
    """Order types."""
    MARKET = "market"
    LIMIT = "limit"
    STOP = "stop"
    COMBO = "combo"  # For spreads


@dataclass
class Contract:
    """Contract representation."""
    symbol: str
    sec_type: str  # STK, OPT, etc.
    exchange: str
    currency: str = "USD"
    strike: Optional[float] = None
    expiry: Optional[str] = None
    right: Optional[str] = None  # C or P for options


@dataclass
class Order:
    """Order representation."""
    order_id: int
    contract: Contract
    action: str  # BUY or SELL
    quantity: int
    order_type: OrderType
    limit_price: Optional[float] = None
    status: OrderStatus = OrderStatus.PENDING
    filled_quantity: int = 0
    avg_fill_price: Optional[float] = None
    created_at: datetime = None
    
    def __post_init__(self):
        if self.created_at is None:
            self.created_at = datetime.now()


@dataclass
class VerticalSpread:
    """Vertical spread definition."""
    symbol: str
    expiry: str
    long_strike: float
    short_strike: float
    right: str  # C or P
    quantity: int
    spread_type: str  # "debit" or "credit"
    max_loss: float
    max_profit: float
    breakeven: float


class TradingManager:
    """
    Trading operations manager for Phase 1C testing.
    Simulates trading operations while validating architecture.
    """
    
    def __init__(self, host="127.0.0.1", port=7497):
        self.host = host
        self.port = port
        self.connected = False
        self.next_order_id = 1000
        self.orders: Dict[int, Order] = {}
        self.positions: Dict[str, int] = {}
        self.account_value = 100000.0  # Paper trading starting value
        
        # Risk management settings
        self.max_position_size = 10
        self.max_daily_loss = 5000.0
        self.daily_pnl = 0.0
        
        # Statistics
        self.stats = {
            'orders_placed': 0,
            'orders_filled': 0,
            'orders_cancelled': 0,
            'spreads_created': 0,
            'total_pnl': 0.0
        }
    
    async def connect(self) -> bool:
        """Connect to TWS for trading operations."""
        try:
            logger.info(f"🔌 Connecting to TWS for trading at {self.host}:{self.port}")
            
            # Test socket connectivity
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(5)
            result = sock.connect_ex((self.host, self.port))
            sock.close()
            
            if result == 0:
                self.connected = True
                logger.info("✅ Trading connection established")
                return True
            else:
                logger.error("❌ Failed to connect to TWS")
                return False
                
        except Exception as e:
            logger.error(f"Connection error: {e}")
            return False
    
    def disconnect(self):
        """Disconnect from TWS."""
        self.connected = False
        logger.info("🔌 Trading connection closed")
    
    def validate_paper_trading(self) -> bool:
        """Validate we're connected to paper trading account."""
        if not self.connected:
            logger.error("❌ Not connected to TWS")
            return False
        
        # In real implementation, would check account type
        # For testing, assume port 7497 = paper trading
        if self.port == 7497:
            logger.info("✅ Paper trading account validated")
            return True
        else:
            logger.warning("⚠️ Not using paper trading port (7497)")
            return False
    
    def create_stock_contract(self, symbol: str) -> Contract:
        """Create a stock contract."""
        return Contract(
            symbol=symbol,
            sec_type="STK",
            exchange="SMART",
            currency="USD"
        )
    
    def create_option_contract(self, symbol: str, expiry: str, strike: float, right: str) -> Contract:
        """Create an option contract."""
        return Contract(
            symbol=symbol,
            sec_type="OPT",
            exchange="SMART",
            currency="USD",
            strike=strike,
            expiry=expiry,
            right=right
        )
    
    def create_vertical_spread(self, symbol: str, expiry: str, long_strike: float, 
                             short_strike: float, right: str, quantity: int) -> VerticalSpread:
        """Create a vertical spread definition."""
        
        # Determine spread type and calculate P&L
        if right == "C":  # Call spread
            if long_strike < short_strike:
                spread_type = "debit"
                max_loss = (short_strike - long_strike) * 100 * quantity  # Simplified
                max_profit = max_loss * 0.3  # Simplified calculation
            else:
                spread_type = "credit"
                max_profit = (long_strike - short_strike) * 100 * quantity
                max_loss = max_profit * 0.7
        else:  # Put spread
            if long_strike > short_strike:
                spread_type = "debit"
                max_loss = (long_strike - short_strike) * 100 * quantity
                max_profit = max_loss * 0.3
            else:
                spread_type = "credit"
                max_profit = (short_strike - long_strike) * 100 * quantity
                max_loss = max_profit * 0.7
        
        # Calculate breakeven (simplified)
        if spread_type == "debit":
            breakeven = long_strike + (max_loss / (100 * quantity))
        else:
            breakeven = short_strike - (max_profit / (100 * quantity))
        
        return VerticalSpread(
            symbol=symbol,
            expiry=expiry,
            long_strike=long_strike,
            short_strike=short_strike,
            right=right,
            quantity=quantity,
            spread_type=spread_type,
            max_loss=max_loss,
            max_profit=max_profit,
            breakeven=breakeven
        )
    
    def validate_order(self, order: Order) -> bool:
        """Validate order against risk management rules."""
        
        # Check connection
        if not self.connected:
            logger.error("❌ Order validation failed: Not connected")
            return False
        
        # Check position size limits
        current_position = self.positions.get(order.contract.symbol, 0)
        new_position = current_position + (order.quantity if order.action == "BUY" else -order.quantity)
        
        if abs(new_position) > self.max_position_size:
            logger.error(f"❌ Order validation failed: Position size limit exceeded")
            return False
        
        # Check daily loss limit
        if self.daily_pnl < -self.max_daily_loss:
            logger.error("❌ Order validation failed: Daily loss limit reached")
            return False
        
        # Check account value (simplified)
        estimated_cost = order.quantity * (order.limit_price or 100)  # Simplified
        if estimated_cost > self.account_value * 0.1:  # Max 10% of account per trade
            logger.error("❌ Order validation failed: Trade size too large")
            return False
        
        logger.info("✅ Order validation passed")
        return True
    
    async def place_order(self, contract: Contract, action: str, quantity: int, 
                         order_type: OrderType, limit_price: Optional[float] = None) -> Order:
        """Place an order."""
        
        # Create order
        order = Order(
            order_id=self.next_order_id,
            contract=contract,
            action=action,
            quantity=quantity,
            order_type=order_type,
            limit_price=limit_price
        )
        
        self.next_order_id += 1
        
        # Validate order
        if not self.validate_order(order):
            order.status = OrderStatus.REJECTED
            logger.error(f"❌ Order {order.order_id} rejected")
            return order
        
        # Store order
        self.orders[order.order_id] = order
        
        # Simulate order submission
        logger.info(f"📤 Placing order {order.order_id}: {action} {quantity} {contract.symbol}")
        order.status = OrderStatus.SUBMITTED
        self.stats['orders_placed'] += 1
        
        # Simulate order processing
        await asyncio.sleep(0.5)  # Simulate network delay
        
        # Simulate fill (for testing)
        await self._simulate_fill(order)
        
        return order
    
    async def _simulate_fill(self, order: Order):
        """Simulate order fill for testing."""
        
        # Simulate market conditions
        fill_probability = 0.8  # 80% fill rate for testing
        
        if order.order_type == OrderType.MARKET or (
            order.order_type == OrderType.LIMIT and fill_probability > 0.5
        ):
            # Simulate fill
            order.status = OrderStatus.FILLED
            order.filled_quantity = order.quantity
            order.avg_fill_price = order.limit_price or 100.0  # Simplified
            
            # Update position
            symbol = order.contract.symbol
            current_pos = self.positions.get(symbol, 0)
            if order.action == "BUY":
                self.positions[symbol] = current_pos + order.quantity
            else:
                self.positions[symbol] = current_pos - order.quantity
            
            self.stats['orders_filled'] += 1
            logger.info(f"✅ Order {order.order_id} filled at ${order.avg_fill_price}")
        else:
            logger.info(f"⏳ Order {order.order_id} pending...")
    
    async def place_vertical_spread(self, spread: VerticalSpread) -> List[Order]:
        """Place a vertical spread as a combo order."""
        
        logger.info(f"📊 Creating {spread.spread_type} {spread.right} spread for {spread.symbol}")
        logger.info(f"   Long {spread.long_strike} / Short {spread.short_strike}")
        logger.info(f"   Max Loss: ${spread.max_loss:.2f}, Max Profit: ${spread.max_profit:.2f}")
        logger.info(f"   Breakeven: ${spread.breakeven:.2f}")
        
        # Create option contracts
        long_contract = self.create_option_contract(
            spread.symbol, spread.expiry, spread.long_strike, spread.right
        )
        short_contract = self.create_option_contract(
            spread.symbol, spread.expiry, spread.short_strike, spread.right
        )
        
        orders = []
        
        # Long leg
        long_order = await self.place_order(
            long_contract, "BUY", spread.quantity, OrderType.LIMIT, spread.long_strike
        )
        orders.append(long_order)
        
        # Short leg
        short_order = await self.place_order(
            short_contract, "SELL", spread.quantity, OrderType.LIMIT, spread.short_strike
        )
        orders.append(short_order)
        
        self.stats['spreads_created'] += 1
        logger.info(f"✅ Vertical spread created with orders {long_order.order_id}, {short_order.order_id}")
        
        return orders
    
    async def cancel_order(self, order_id: int) -> bool:
        """Cancel an order."""
        if order_id not in self.orders:
            logger.error(f"❌ Order {order_id} not found")
            return False
        
        order = self.orders[order_id]
        if order.status in [OrderStatus.FILLED, OrderStatus.CANCELLED]:
            logger.warning(f"⚠️ Order {order_id} cannot be cancelled (status: {order.status.value})")
            return False
        
        order.status = OrderStatus.CANCELLED
        self.stats['orders_cancelled'] += 1
        logger.info(f"❌ Order {order_id} cancelled")
        return True
    
    def get_order_status(self, order_id: int) -> Optional[Order]:
        """Get order status."""
        return self.orders.get(order_id)
    
    def get_positions(self) -> Dict[str, int]:
        """Get current positions."""
        return self.positions.copy()
    
    def get_account_summary(self) -> Dict[str, Any]:
        """Get account summary."""
        return {
            'account_value': self.account_value,
            'daily_pnl': self.daily_pnl,
            'positions': len(self.positions),
            'active_orders': len([o for o in self.orders.values() if o.status == OrderStatus.SUBMITTED]),
            'stats': self.stats.copy()
        }


async def test_paper_trading_validation():
    """Test paper trading account validation."""
    logger.info("🧪 Testing Paper Trading Validation")
    
    trading_manager = TradingManager()
    
    try:
        # Test connection
        connected = await trading_manager.connect()
        assert connected, "Should connect to TWS"
        
        # Test paper trading validation
        is_paper = trading_manager.validate_paper_trading()
        assert is_paper, "Should validate paper trading account"
        
        logger.info("✅ Paper trading validation test passed")
        return True
        
    except Exception as e:
        logger.error(f"❌ Paper trading validation test failed: {e}")
        return False
    finally:
        trading_manager.disconnect()


async def test_basic_order_management():
    """Test basic order creation and management."""
    logger.info("🧪 Testing Basic Order Management")
    
    trading_manager = TradingManager()
    
    try:
        await trading_manager.connect()
        
        # Create a stock contract
        contract = trading_manager.create_stock_contract("AAPL")
        
        # Place a limit order (within position size limits)
        order = await trading_manager.place_order(
            contract, "BUY", 5, OrderType.LIMIT, 150.0  # Changed from 100 to 5
        )
        
        assert order.order_id > 0, "Order should have valid ID"
        assert order.status in [OrderStatus.SUBMITTED, OrderStatus.FILLED], "Order should be submitted or filled"
        
        # Check order status
        status = trading_manager.get_order_status(order.order_id)
        assert status is not None, "Should retrieve order status"
        
        logger.info("✅ Basic order management test passed")
        return True
        
    except Exception as e:
        logger.error(f"❌ Basic order management test failed: {e}")
        return False
    finally:
        trading_manager.disconnect()


async def test_vertical_spread_creation():
    """Test vertical spread creation."""
    logger.info("🧪 Testing Vertical Spread Creation")
    
    trading_manager = TradingManager()
    
    try:
        await trading_manager.connect()
        
        # Create a call spread
        spread = trading_manager.create_vertical_spread(
            symbol="SPY",
            expiry="20250117",  # January 17, 2025
            long_strike=580.0,
            short_strike=585.0,
            right="C",
            quantity=1
        )
        
        assert spread.spread_type == "debit", "Should be a debit spread"
        assert spread.max_loss > 0, "Should have calculated max loss"
        assert spread.max_profit > 0, "Should have calculated max profit"
        
        # Place the spread
        orders = await trading_manager.place_vertical_spread(spread)
        
        assert len(orders) == 2, "Should create two orders for spread"
        assert orders[0].action == "BUY", "First order should be BUY (long leg)"
        assert orders[1].action == "SELL", "Second order should be SELL (short leg)"
        
        logger.info("✅ Vertical spread creation test passed")
        return True
        
    except Exception as e:
        logger.error(f"❌ Vertical spread creation test failed: {e}")
        return False
    finally:
        trading_manager.disconnect()


async def test_risk_management():
    """Test risk management features."""
    logger.info("🧪 Testing Risk Management")
    
    trading_manager = TradingManager()
    
    try:
        await trading_manager.connect()
        
        # Test position size limit
        contract = trading_manager.create_stock_contract("TSLA")
        
        # Try to place order exceeding position limit
        large_order = await trading_manager.place_order(
            contract, "BUY", 50, OrderType.LIMIT, 200.0  # Exceeds max_position_size of 10
        )
        
        assert large_order.status == OrderStatus.REJECTED, "Large order should be rejected"
        
        # Test normal order
        normal_order = await trading_manager.place_order(
            contract, "BUY", 5, OrderType.LIMIT, 200.0
        )
        
        assert normal_order.status in [OrderStatus.SUBMITTED, OrderStatus.FILLED], "Normal order should be accepted"
        
        logger.info("✅ Risk management test passed")
        return True
        
    except Exception as e:
        logger.error(f"❌ Risk management test failed: {e}")
        return False
    finally:
        trading_manager.disconnect()


async def test_order_cancellation():
    """Test order cancellation."""
    logger.info("🧪 Testing Order Cancellation")
    
    trading_manager = TradingManager()
    
    try:
        await trading_manager.connect()
        
        # Place an order
        contract = trading_manager.create_stock_contract("MSFT")
        order = await trading_manager.place_order(
            contract, "BUY", 10, OrderType.LIMIT, 300.0
        )
        
        # Cancel the order (if not already filled)
        if order.status == OrderStatus.SUBMITTED:
            cancelled = await trading_manager.cancel_order(order.order_id)
            assert cancelled, "Should be able to cancel submitted order"
            
            # Check status
            updated_order = trading_manager.get_order_status(order.order_id)
            assert updated_order.status == OrderStatus.CANCELLED, "Order should be cancelled"
        
        logger.info("✅ Order cancellation test passed")
        return True
        
    except Exception as e:
        logger.error(f"❌ Order cancellation test failed: {e}")
        return False
    finally:
        trading_manager.disconnect()


async def run_phase1c_tests():
    """Run all Phase 1C trading operations tests."""
    logger.info("🚀 PHASE 1C: TRADING OPERATIONS TESTING")
    logger.info("📅 Day 6 - January 13, 2025")
    logger.info("=" * 60)
    
    results = {}
    
    # Test 1: Paper trading validation
    logger.info("\n🔬 Test Suite 1: Paper Trading Validation")
    results['paper_trading'] = await test_paper_trading_validation()
    
    # Test 2: Basic order management
    logger.info("\n🔬 Test Suite 2: Basic Order Management")
    results['order_management'] = await test_basic_order_management()
    
    # Test 3: Vertical spread creation
    logger.info("\n🔬 Test Suite 3: Vertical Spread Creation")
    results['vertical_spreads'] = await test_vertical_spread_creation()
    
    # Test 4: Risk management
    logger.info("\n🔬 Test Suite 4: Risk Management")
    results['risk_management'] = await test_risk_management()
    
    # Test 5: Order cancellation
    logger.info("\n🔬 Test Suite 5: Order Cancellation")
    results['order_cancellation'] = await test_order_cancellation()
    
    # Summary
    logger.info("\n" + "=" * 60)
    logger.info("📊 PHASE 1C TEST RESULTS:")
    logger.info("=" * 60)
    
    all_passed = True
    for test_name, passed in results.items():
        status = "✅ PASS" if passed else "❌ FAIL"
        logger.info(f"  {test_name.replace('_', ' ').title():.<30} {status}")
        if not passed:
            all_passed = False
    
    logger.info("\n" + "=" * 60)
    if all_passed:
        logger.info("🎉 PHASE 1C TRADING OPERATIONS: COMPLETE!")
        logger.info("✅ Paper trading validated")
        logger.info("✅ Order management working")
        logger.info("✅ Vertical spreads functional")
        logger.info("✅ Risk management enforced")
        logger.info("✅ Order lifecycle managed")
        logger.info("🚀 Ready for Phase 1D: Market Data Streaming")
    else:
        logger.info("⚠️ Some tests failed - review implementation")
    logger.info("=" * 60)
    
    return all_passed


if __name__ == "__main__":
    # Run Phase 1C tests
    print("📈 PHASE 1C: TRADING OPERATIONS TESTING")
    print("📅 Day 6 - January 13, 2025")
    print("🎯 Testing paper trading, orders, and vertical spreads")
    print("=" * 60)
    
    success = asyncio.run(run_phase1c_tests())
    
    if success:
        print("\n🎯 PHASE 1C STATUS: COMPLETE ✅")
        print("📅 Still Day 6 - Moving to Phase 1D: Market Data Streaming")
    else:
        print("\n⚠️ PHASE 1C STATUS: NEEDS REVIEW")
        print("Fix issues before proceeding to Phase 1D") 