# TWS API Precautions and Rate Limits Guide
## Navigate the Minefield of TWS Restrictions

> ⚠️ **Critical**: TWS has undocumented limits that will bite you. This guide covers both official and discovered limits.

### API Precautions Settings

Navigate to: `Configure > Settings > API > Precautions`

#### Essential Precautions to DISABLE for Automation

```
[✓] Bypass Order Precautions for API Orders     <-- MUST ENABLE
[ ] Bypass Bond warning for bonds                <-- Keep disabled
[✓] Bypass negative yield confirmation           <-- Enable for options
[ ] Bypass Called Bond warning                   <-- Keep disabled  
[✓] Bypass "same action pair trade" warning     <-- Enable for spreads
[✓] Allow order to be transmitted without warning <-- MUST ENABLE
```

> **WARNING**: Disabling precautions means YOU are responsible for validation!

### Rate Limits - The Real Numbers

#### Official Limits
```python
OFFICIAL_LIMITS = {
    'messages_per_second': 50,          # Hard limit
    'identical_ticks_per_second': 1,    # Per contract
    'market_depth_per_second': 10,      # Level 2 data
    'scanner_results': 50,              # Per scan
    'historical_data_concurrent': 6,    # Parallel requests
    'positions_per_request': 50,        # Account positions
}
```

#### Discovered Limits (Undocumented)
```python
REAL_WORLD_LIMITS = {
    'safe_messages_per_second': 40,     # 80% of official
    'burst_messages_allowed': 100,      # Over 2 seconds
    'scanner_concurrent': 3,            # More = errors
    'market_data_lines': {
        'basic': 100,
        'pro': 1000,
        'vip': 10000
    },
    'order_modifications_per_minute': 60,
    'contract_details_batch': 20,       # Per request
    'news_requests_per_minute': 10,
}
```

### Request Pacing Implementation

```python
import asyncio
from collections import deque
from datetime import datetime, timedelta

class TWS RateLimiter:
    """Intelligent rate limiter for TWS API"""
    
    def __init__(self):
        self.request_times = deque(maxlen=1000)
        self.burst_bucket = 100
        self.refill_rate = 40  # per second
        self.last_refill = datetime.now()
        
    async def acquire(self, priority: int = 5):
        """Acquire permission to send request"""
        while True:
            now = datetime.now()
            self._refill_bucket(now)
            
            # Check rate limit
            self._clean_old_requests(now)
            
            if len(self.request_times) < 40:  # Under rate limit
                if self.burst_bucket > 0:
                    self.burst_bucket -= 1
                    self.request_times.append(now)
                    return
                    
            # High priority requests get preference
            wait_time = 0.025 * (10 - priority)  # 25-250ms
            await asyncio.sleep(wait_time)
    
    def _refill_bucket(self, now: datetime):
        """Refill burst bucket"""
        elapsed = (now - self.last_refill).total_seconds()
        tokens_to_add = int(elapsed * self.refill_rate)
        
        if tokens_to_add > 0:
            self.burst_bucket = min(100, self.burst_bucket + tokens_to_add)
            self.last_refill = now
    
    def _clean_old_requests(self, now: datetime):
        """Remove requests older than 1 second"""
        cutoff = now - timedelta(seconds=1)
        while self.request_times and self.request_times[0] < cutoff:
            self.request_times.popleft()
```

### Message Type Priority System

```python
class MessagePriority:
    """Priority levels for different message types"""
    
    # Highest priority - trading operations
    ORDER_NEW = 10
    ORDER_CANCEL = 10
    ORDER_MODIFY = 9
    
    # High priority - critical data
    POSITION_UPDATE = 8
    ACCOUNT_UPDATE = 8
    ORDER_STATUS = 8
    
    # Medium priority - market data
    MARKET_DATA_SUBSCRIBE = 6
    MARKET_DATA_UNSUBSCRIBE = 6
    SCANNER_SUBSCRIBE = 5
    
    # Low priority - historical/info
    HISTORICAL_DATA = 3
    CONTRACT_DETAILS = 3
    NEWS_REQUEST = 2
    
    # Lowest priority
    SERVER_TIME = 1
    
async def send_with_priority(client, request_func, priority: int, *args):
    """Send request with rate limiting and priority"""
    await rate_limiter.acquire(priority)
    return await request_func(*args)
```

### Market Data Line Management

```python
class MarketDataManager:
    """Manage market data lines efficiently"""
    
    def __init__(self, max_lines: int = 100):
        self.max_lines = max_lines
        self.active_subscriptions = {}
        self.subscription_times = {}
        self.lock = asyncio.Lock()
        
    async def subscribe(self, contract, tick_types: str = ""):
        """Subscribe with automatic cleanup"""
        async with self.lock:
            # Check if at limit
            if len(self.active_subscriptions) >= self.max_lines:
                # Remove oldest subscription
                await self._remove_oldest()
            
            req_id = next_req_id()
            self.active_subscriptions[req_id] = contract
            self.subscription_times[req_id] = datetime.now()
            
            await client.reqMktData(req_id, contract, tick_types, False)
            return req_id
    
    async def _remove_oldest(self):
        """Remove oldest market data subscription"""
        if not self.subscription_times:
            return
            
        oldest_id = min(self.subscription_times, 
                       key=self.subscription_times.get)
        
        await client.cancelMktData(oldest_id)
        del self.active_subscriptions[oldest_id]
        del self.subscription_times[oldest_id]
```

### Error Code Reference

```python
ERROR_CODES = {
    # Rate limit errors
    100: "Max rate of messages exceeded",        # Slow down!
    101: "Max number of tickers reached",        # Too many subscriptions
    102: "Duplicate ticker ID",                  # ID already in use
    103: "Duplicate order ID",                   # Order ID collision
    
    # Pacing violations
    354: "Requested market data is not subscribed",
    420: "Historical data request pacing violation",
    162: "Historical data service error message",
    
    # Connection limits
    503: "Max number of API clients reached",    # Too many connections
    504: "Not connected",                        # Connection lost
    
    # Data limits
    10090: "Part of requested market data is not subscribed",
    10167: "Requested market data requires additional subscription",
}

class ErrorHandler:
    def __init__(self):
        self.error_counts = defaultdict(int)
        self.last_errors = deque(maxlen=100)
        
    async def handle_error(self, req_id: int, error_code: int, error_msg: str):
        """Handle TWS errors intelligently"""
        self.error_counts[error_code] += 1
        self.last_errors.append((datetime.now(), error_code, error_msg))
        
        if error_code == 100:  # Rate limit
            logger.warning("Rate limit hit - increasing delay")
            await rate_limiter.emergency_slowdown()
            
        elif error_code == 420:  # Historical pacing
            logger.warning("Historical data pacing - waiting 10s")
            await asyncio.sleep(10)
            
        elif error_code in [503, 504]:  # Connection issues
            logger.error("Connection issue - triggering reconnect")
            await client.reconnect()
```

### Scanner Subscription Limits

```python
class ScannerManager:
    """Manage scanner subscriptions within limits"""
    
    MAX_CONCURRENT_SCANNERS = 3
    MAX_RESULTS_PER_SCAN = 50
    
    def __init__(self):
        self.active_scanners = {}
        self.scanner_queue = asyncio.Queue()
        self.scanner_semaphore = asyncio.Semaphore(self.MAX_CONCURRENT_SCANNERS)
        
    async def request_scanner(self, scan_params: ScannerSubscription):
        """Request scanner with queuing if at limit"""
        async with self.scanner_semaphore:
            req_id = next_req_id()
            
            # Enforce result limit
            scan_params.numberOfRows = min(
                scan_params.numberOfRows, 
                self.MAX_RESULTS_PER_SCAN
            )
            
            self.active_scanners[req_id] = scan_params
            results = await client.reqScannerSubscription(req_id, scan_params)
            
            # Auto-cleanup after receiving results
            await client.cancelScannerSubscription(req_id)
            del self.active_scanners[req_id]
            
            return results
```

### Order Management Limits

```python
class OrderLimitManager:
    """Manage order-related limits"""
    
    def __init__(self):
        self.order_modifications = deque(maxlen=60)
        self.pending_orders = {}
        self.order_lock = asyncio.Lock()
        
    async def can_modify_order(self, order_id: int) -> bool:
        """Check if order modification is allowed"""
        now = datetime.now()
        
        # Clean old modifications
        cutoff = now - timedelta(minutes=1)
        while self.order_modifications and self.order_modifications[0] < cutoff:
            self.order_modifications.popleft()
        
        # Check limit
        if len(self.order_modifications) >= 60:
            wait_time = (self.order_modifications[0] + timedelta(minutes=1) - now).total_seconds()
            logger.warning(f"Order modification limit reached, wait {wait_time:.1f}s")
            return False
            
        return True
    
    async def modify_order(self, order_id: int, order: Order):
        """Modify order with limit checking"""
        async with self.order_lock:
            if not await self.can_modify_order(order_id):
                raise Exception("Order modification rate limit exceeded")
                
            self.order_modifications.append(datetime.now())
            await client.placeOrder(order_id, contract, order)
```

### Production Best Practices

#### 1. Request Batching
```python
async def get_contract_details_batch(contracts: List[Contract]):
    """Batch contract detail requests efficiently"""
    results = []
    
    # Process in chunks of 20
    for i in range(0, len(contracts), 20):
        chunk = contracts[i:i+20]
        chunk_results = await asyncio.gather(*[
            client.reqContractDetails(c) for c in chunk
        ])
        results.extend(chunk_results)
        
        # Pause between chunks
        if i + 20 < len(contracts):
            await asyncio.sleep(0.5)
    
    return results
```

#### 2. Subscription Recycling
```python
class SubscriptionRecycler:
    """Reuse subscription IDs efficiently"""
    
    def __init__(self, min_id=1000, max_id=9999):
        self.available_ids = set(range(min_id, max_id))
        self.in_use = set()
        
    def get_id(self) -> int:
        if not self.available_ids:
            raise Exception("No available subscription IDs")
        
        req_id = self.available_ids.pop()
        self.in_use.add(req_id)
        return req_id
    
    def release_id(self, req_id: int):
        self.in_use.discard(req_id)
        self.available_ids.add(req_id)
```

### Monitoring and Alerts

```python
# Prometheus metrics for limit monitoring
limit_metrics = {
    'api_rate_limit_hits': Counter('tws_api_rate_limit_hits', 'Rate limit violations'),
    'api_request_rate': Histogram('tws_api_request_rate', 'Requests per second'),
    'market_data_lines_used': Gauge('tws_market_data_lines_used', 'Active subscriptions'),
    'scanner_queue_depth': Gauge('tws_scanner_queue_depth', 'Pending scanner requests'),
    'order_modification_rate': Counter('tws_order_modification_rate', 'Order mods per minute'),
}

# Alert when approaching limits
if len(active_subscriptions) > max_lines * 0.9:
    alert("Approaching market data line limit", severity="warning")
```

### Emergency Procedures

```python
async def emergency_throttle():
    """Emergency throttling when limits are hit"""
    logger.critical("Entering emergency throttle mode")
    
    # 1. Cancel all non-essential subscriptions
    await cancel_all_scanners()
    await cancel_historical_requests()
    
    # 2. Increase rate limiter delays
    rate_limiter.emergency_mode = True
    
    # 3. Pause for recovery
    await asyncio.sleep(30)
    
    # 4. Gradually resume
    rate_limiter.emergency_mode = False
    logger.info("Exiting emergency throttle mode")
```

### Testing Rate Limits

```python
async def test_rate_limits():
    """Test your rate limiting implementation"""
    
    # Test burst capacity
    print("Testing burst capacity...")
    start = time.time()
    for i in range(100):
        await rate_limiter.acquire(priority=5)
    burst_time = time.time() - start
    print(f"100 requests in {burst_time:.2f}s")
    
    # Test sustained rate
    print("Testing sustained rate...")
    request_times = []
    for i in range(200):
        await rate_limiter.acquire(priority=5)
        request_times.append(time.time())
    
    # Calculate actual rate
    duration = request_times[-1] - request_times[0]
    actual_rate = len(request_times) / duration
    print(f"Sustained rate: {actual_rate:.1f} req/s")
```

### Limit Checklist

- [ ] Rate limiter implemented with burst handling
- [ ] Priority system for critical operations
- [ ] Market data line tracking and cleanup
- [ ] Scanner concurrency management
- [ ] Order modification tracking
- [ ] Error handling for all limit errors
- [ ] Monitoring metrics configured
- [ ] Emergency throttling procedures
- [ ] Regular limit testing in place