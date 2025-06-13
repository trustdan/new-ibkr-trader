# Order ID Management System

## Critical Importance

Order ID management is one of the most critical aspects of TWS integration. Improper handling leads to:
- Rejected orders
- Duplicate executions
- System crashes
- Compliance violations

> ⚠️ **NEVER reuse an Order ID. NEVER skip Order IDs. ALWAYS use sequential IDs.**

## Order ID Architecture

```
┌─────────────────┐     nextValidId      ┌──────────────────┐
│                 │ ◄─────────────────── │                  │
│   Order ID      │                      │   TWS Server     │
│   Manager       │ ──placeOrder(id)──► │                  │
└────────┬────────┘                      └──────────────────┘
         │
         ├──────────────┬──────────────┬────────────────┐
         ▼              ▼              ▼                ▼
    Main Orders    Scanner Orders  Bracket Orders  Algo Orders
    (0-999)        (1000-1999)     (2000-2999)    (3000-3999)
```

## Core Implementation

### 1. Thread-Safe Order ID Manager

```python
import threading
import asyncio
from typing import Optional, Dict, Set
from dataclasses import dataclass
from datetime import datetime
import pickle

@dataclass
class OrderIdState:
    """Persistent order ID state"""
    next_valid_id: int
    reserved_ranges: Dict[str, tuple]
    used_ids: Set[int]
    last_updated: datetime

class OrderIdManager:
    """Thread-safe order ID management with persistence"""
    
    def __init__(self, state_file: str = "order_id_state.pkl"):
        self.state_file = state_file
        self._lock = threading.RLock()
        self._async_lock = asyncio.Lock()
        
        # Initialize state
        self.state = self._load_state()
        
        # Reserve ID ranges for different purposes
        self.ranges = {
            'manual': (0, 999),
            'scanner': (1000, 1999),
            'bracket': (2000, 2999),
            'algo': (3000, 3999),
            'test': (9000, 9999)
        }
        
        # Track active orders
        self.active_orders: Dict[int, dict] = {}
        
    def _load_state(self) -> OrderIdState:
        """Load persisted state or create new"""
        try:
            with open(self.state_file, 'rb') as f:
                state = pickle.load(f)
                logger.info(f"Loaded order ID state: next_id={state.next_valid_id}")
                return state
        except FileNotFoundError:
            logger.info("Creating new order ID state")
            return OrderIdState(
                next_valid_id=1,
                reserved_ranges={},
                used_ids=set(),
                last_updated=datetime.now()
            )
    
    def _save_state(self):
        """Persist current state"""
        self.state.last_updated = datetime.now()
        with open(self.state_file, 'wb') as f:
            pickle.dump(self.state, f)
    
    def set_next_valid_id(self, next_id: int):
        """Update next valid ID from TWS"""
        with self._lock:
            if next_id > self.state.next_valid_id:
                logger.info(f"Updating next valid ID: {self.state.next_valid_id} -> {next_id}")
                self.state.next_valid_id = next_id
                self._save_state()
            else:
                logger.warning(f"Received lower next_id {next_id}, keeping {self.state.next_valid_id}")
    
    def get_next_id(self, purpose: str = 'manual') -> int:
        """Get next available order ID"""
        with self._lock:
            # Check if we need ID from specific range
            if purpose in self.ranges:
                range_start, range_end = self.ranges[purpose]
                
                # Find next available in range
                for id_ in range(max(range_start, self.state.next_valid_id), range_end + 1):
                    if id_ not in self.state.used_ids:
                        self.state.used_ids.add(id_)
                        self._save_state()
                        return id_
                
                raise Exception(f"No available IDs in {purpose} range")
            
            # Use next sequential ID
            order_id = self.state.next_valid_id
            self.state.next_valid_id += 1
            self.state.used_ids.add(order_id)
            self._save_state()
            
            return order_id
    
    async def get_next_id_async(self, purpose: str = 'manual') -> int:
        """Async version for asyncio code"""
        async with self._async_lock:
            return self.get_next_id(purpose)
    
    def reserve_id_range(self, start: int, count: int, purpose: str) -> List[int]:
        """Reserve a range of IDs for batch operations"""
        with self._lock:
            reserved = []
            
            # Ensure we start from valid ID
            start = max(start, self.state.next_valid_id)
            
            for i in range(count):
                id_ = start + i
                if id_ in self.state.used_ids:
                    raise Exception(f"ID {id_} already used!")
                
                self.state.used_ids.add(id_)
                reserved.append(id_)
            
            # Update next valid ID if needed
            if start + count > self.state.next_valid_id:
                self.state.next_valid_id = start + count
            
            self._save_state()
            return reserved
    
    def mark_order_active(self, order_id: int, order_info: dict):
        """Track active order"""
        with self._lock:
            self.active_orders[order_id] = {
                **order_info,
                'created_at': datetime.now()
            }
    
    def mark_order_completed(self, order_id: int):
        """Mark order as completed"""
        with self._lock:
            if order_id in self.active_orders:
                del self.active_orders[order_id]
```

### 2. Order ID Recovery System

```python
class OrderIdRecovery:
    """Recover from order ID issues"""
    
    def __init__(self, order_manager: OrderIdManager, tws_client):
        self.order_manager = order_manager
        self.client = tws_client
        self.recovery_in_progress = False
        
    async def recover_order_state(self):
        """Recover order state after disconnect"""
        if self.recovery_in_progress:
            logger.warning("Recovery already in progress")
            return
            
        self.recovery_in_progress = True
        
        try:
            logger.info("Starting order ID recovery")
            
            # 1. Get next valid ID from TWS
            next_id = await self.client.reqNextValidId()
            self.order_manager.set_next_valid_id(next_id)
            
            # 2. Get all open orders
            open_orders = await self.client.reqOpenOrders()
            
            # 3. Reconcile with our state
            tws_order_ids = {order.orderId for order in open_orders}
            our_order_ids = set(self.order_manager.active_orders.keys())
            
            # Orders in TWS but not in our state
            unknown_orders = tws_order_ids - our_order_ids
            if unknown_orders:
                logger.warning(f"Found {len(unknown_orders)} unknown orders in TWS: {unknown_orders}")
                
            # Orders in our state but not in TWS
            missing_orders = our_order_ids - tws_order_ids
            if missing_orders:
                logger.warning(f"Found {len(missing_orders)} orders missing from TWS: {missing_orders}")
                # Mark as completed
                for order_id in missing_orders:
                    self.order_manager.mark_order_completed(order_id)
            
            logger.info("Order ID recovery completed")
            
        finally:
            self.recovery_in_progress = False
```

### 3. Order ID Validation

```python
class OrderIdValidator:
    """Validate order IDs before submission"""
    
    def __init__(self, order_manager: OrderIdManager):
        self.order_manager = order_manager
        self.validation_rules = [
            self._check_not_reused,
            self._check_sequential,
            self._check_not_reserved,
            self._check_within_range
        ]
        
    def validate(self, order_id: int, purpose: str = 'manual') -> bool:
        """Validate order ID"""
        for rule in self.validation_rules:
            if not rule(order_id, purpose):
                return False
        return True
    
    def _check_not_reused(self, order_id: int, purpose: str) -> bool:
        """Ensure ID hasn't been used"""
        if order_id in self.order_manager.state.used_ids:
            logger.error(f"Order ID {order_id} has already been used!")
            return False
        return True
    
    def _check_sequential(self, order_id: int, purpose: str) -> bool:
        """Ensure ID is sequential (with some tolerance)"""
        if purpose == 'manual':
            expected_range = range(
                self.order_manager.state.next_valid_id - 100,
                self.order_manager.state.next_valid_id + 100
            )
            if order_id not in expected_range:
                logger.warning(f"Order ID {order_id} not in expected range")
                return False
        return True
    
    def _check_not_reserved(self, order_id: int, purpose: str) -> bool:
        """Ensure ID isn't in reserved range"""
        for range_purpose, (start, end) in self.order_manager.ranges.items():
            if range_purpose != purpose and start <= order_id <= end:
                logger.error(f"Order ID {order_id} is in reserved range for {range_purpose}")
                return False
        return True
    
    def _check_within_range(self, order_id: int, purpose: str) -> bool:
        """Ensure ID is within valid range"""
        if order_id < 0 or order_id > 2147483647:  # Max 32-bit int
            logger.error(f"Order ID {order_id} outside valid range")
            return False
        return True
```

### 4. Bracket Order ID Management

```python
class BracketOrderManager:
    """Special handling for bracket orders"""
    
    def __init__(self, order_manager: OrderIdManager):
        self.order_manager = order_manager
        
    def get_bracket_ids(self) -> Tuple[int, int, int]:
        """Get IDs for bracket order (parent, profit, stop)"""
        with self.order_manager._lock:
            # Reserve 3 consecutive IDs
            parent_id = self.order_manager.get_next_id('bracket')
            profit_id = self.order_manager.get_next_id('bracket')
            stop_id = self.order_manager.get_next_id('bracket')
            
            # Ensure they're consecutive (TWS preference)
            if profit_id != parent_id + 1 or stop_id != parent_id + 2:
                logger.warning("Bracket IDs not consecutive, may cause issues")
            
            return parent_id, profit_id, stop_id
    
    def create_bracket_order(self, 
                           contract: Contract,
                           quantity: int,
                           limit_price: float,
                           take_profit: float,
                           stop_loss: float) -> Tuple[Order, Order, Order]:
        """Create complete bracket order"""
        parent_id, profit_id, stop_id = self.get_bracket_ids()
        
        # Parent order
        parent = Order()
        parent.orderId = parent_id
        parent.action = "BUY"
        parent.orderType = "LMT"
        parent.totalQuantity = quantity
        parent.lmtPrice = limit_price
        parent.transmit = False  # Don't transmit until children attached
        
        # Take profit order
        take_profit_order = Order()
        take_profit_order.orderId = profit_id
        take_profit_order.action = "SELL"
        take_profit_order.orderType = "LMT"
        take_profit_order.totalQuantity = quantity
        take_profit_order.lmtPrice = take_profit
        take_profit_order.parentId = parent_id
        take_profit_order.transmit = False
        
        # Stop loss order
        stop_loss_order = Order()
        stop_loss_order.orderId = stop_id
        stop_loss_order.action = "SELL"
        stop_loss_order.orderType = "STP"
        stop_loss_order.totalQuantity = quantity
        stop_loss_order.auxPrice = stop_loss
        stop_loss_order.parentId = parent_id
        stop_loss_order.transmit = True  # Transmit all orders
        
        # OCA group (one-cancels-all)
        oca_group = f"OCA_{parent_id}"
        take_profit_order.ocaGroup = oca_group
        stop_loss_order.ocaGroup = oca_group
        
        return parent, take_profit_order, stop_loss_order
```

### 5. Order ID Monitoring

```python
class OrderIdMonitor:
    """Monitor order ID health"""
    
    def __init__(self, order_manager: OrderIdManager):
        self.order_manager = order_manager
        self.metrics = {
            'order_id_current': Gauge('tws_order_id_current'),
            'order_id_gaps': Counter('tws_order_id_gaps'),
            'order_id_reuse_attempts': Counter('tws_order_id_reuse_attempts'),
            'active_orders': Gauge('tws_active_orders_count')
        }
        
    async def monitor_loop(self):
        """Continuous monitoring"""
        while True:
            try:
                # Update metrics
                self.metrics['order_id_current'].set(
                    self.order_manager.state.next_valid_id
                )
                self.metrics['active_orders'].set(
                    len(self.order_manager.active_orders)
                )
                
                # Check for gaps
                self._check_for_gaps()
                
                # Check for stale orders
                self._check_stale_orders()
                
                await asyncio.sleep(60)  # Check every minute
                
            except Exception as e:
                logger.error(f"Order ID monitor error: {e}")
    
    def _check_for_gaps(self):
        """Detect gaps in order ID sequence"""
        used_ids = sorted(self.order_manager.state.used_ids)
        
        if len(used_ids) < 2:
            return
            
        for i in range(1, len(used_ids)):
            gap = used_ids[i] - used_ids[i-1]
            if gap > 1:
                logger.warning(f"Gap detected in order IDs: {used_ids[i-1]} -> {used_ids[i]}")
                self.metrics['order_id_gaps'].inc()
    
    def _check_stale_orders(self):
        """Check for orders that have been active too long"""
        now = datetime.now()
        stale_threshold = timedelta(hours=24)
        
        for order_id, order_info in self.order_manager.active_orders.items():
            age = now - order_info['created_at']
            if age > stale_threshold:
                logger.warning(f"Stale order detected: {order_id} (age: {age})")
```

### 6. Emergency Procedures

```python
class OrderIdEmergencyHandler:
    """Handle order ID emergencies"""
    
    def __init__(self, order_manager: OrderIdManager, tws_client):
        self.order_manager = order_manager
        self.client = tws_client
        
    async def handle_order_id_collision(self, order_id: int):
        """Handle case where TWS rejects due to ID collision"""
        logger.error(f"Order ID collision detected for {order_id}")
        
        # 1. Request fresh next valid ID
        next_id = await self.client.reqNextValidId()
        self.order_manager.set_next_valid_id(next_id + 100)  # Skip ahead
        
        # 2. Alert operators
        await self.send_alert(
            "Order ID collision detected",
            f"Skipping to ID {next_id + 100}"
        )
        
    async def reset_order_id_state(self):
        """Complete reset of order ID state"""
        logger.warning("Resetting order ID state")
        
        # 1. Cancel all active orders
        for order_id in list(self.order_manager.active_orders.keys()):
            try:
                await self.client.cancelOrder(order_id)
            except Exception as e:
                logger.error(f"Failed to cancel order {order_id}: {e}")
        
        # 2. Get fresh next valid ID
        next_id = await self.client.reqNextValidId()
        
        # 3. Reset state
        self.order_manager.state = OrderIdState(
            next_valid_id=next_id,
            reserved_ranges={},
            used_ids=set(),
            last_updated=datetime.now()
        )
        self.order_manager._save_state()
        
        logger.info(f"Order ID state reset complete. Next ID: {next_id}")
```

## Best Practices

1. **Always use the OrderIdManager** - Never generate IDs manually
2. **Monitor for gaps** - Gaps indicate potential issues
3. **Persist state** - Always save order ID state to disk
4. **Handle reconnections** - Always sync with TWS after reconnect
5. **Use ID ranges** - Reserve ranges for different order types
6. **Implement validation** - Validate every ID before use
7. **Track active orders** - Know which IDs are in use
8. **Plan for emergencies** - Have procedures for ID issues

## Common Pitfalls

1. **Reusing IDs** - TWS will reject or cause undefined behavior
2. **Skipping IDs** - Can cause synchronization issues
3. **Not persisting state** - Lose track after restart
4. **Race conditions** - Multiple threads getting IDs
5. **Not handling disconnects** - ID state gets out of sync

## Testing Order ID Management

```python
async def test_order_id_system():
    """Comprehensive order ID system test"""
    
    # Test 1: Sequential ID generation
    ids = [order_manager.get_next_id() for _ in range(10)]
    assert ids == list(range(ids[0], ids[0] + 10))
    
    # Test 2: Thread safety
    import concurrent.futures
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        future_ids = [executor.submit(order_manager.get_next_id) for _ in range(100)]
        all_ids = {f.result() for f in future_ids}
        assert len(all_ids) == 100  # All unique
    
    # Test 3: Persistence
    current_id = order_manager.state.next_valid_id
    new_manager = OrderIdManager()  # Reload from disk
    assert new_manager.state.next_valid_id == current_id
    
    # Test 4: Recovery after crash
    await test_crash_recovery()
    
    print("All order ID tests passed!")
```