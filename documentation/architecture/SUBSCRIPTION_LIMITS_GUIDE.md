# TWS Subscription Limits by Tier

## Overview

Interactive Brokers enforces strict limits on market data subscriptions based on your account tier. Exceeding these limits results in errors, rejected requests, and potential additional charges.

## Subscription Tiers

### 1. IBKR Lite (Free Tier)
```
Market Data Lines: 100
Concurrent Scanners: 1
Historical Data Requests: Limited
Real-time Bars: Not available
Market Depth: Not available
```

### 2. IBKR Pro (Standard)
```
Market Data Lines: 100 (base)
Concurrent Scanners: 3
Historical Data Requests: 6 concurrent
Real-time Bars: 5 concurrent
Market Depth: 10 concurrent
```

### 3. Market Data Subscriptions (Add-ons)

#### US Securities Snapshot and Futures Value Bundle ($10/month)
```
Market Data Lines: +1000
Enhanced quote snapshot data
Futures market data included
```

#### US Equity and Options Add-On Streaming Bundle ($4.50/month)
```
Market Data Lines: +10000
Real-time streaming quotes
Options chains included
```

#### Market Depth Subscription ($30/month per exchange)
```
Level 2 data access
Full order book visibility
Price by exchange
```

## Subscription Management Strategy

### 1. Tiered Subscription Manager

```python
from enum import Enum
from dataclasses import dataclass
from typing import Dict, Set, Optional
import asyncio

class SubscriptionTier(Enum):
    LITE = "lite"
    PRO = "pro"
    PRO_PLUS = "pro_plus"
    PREMIUM = "premium"

@dataclass
class TierLimits:
    market_data_lines: int
    concurrent_scanners: int
    historical_concurrent: int
    realtime_bars: int
    market_depth: int

TIER_LIMITS = {
    SubscriptionTier.LITE: TierLimits(100, 1, 1, 0, 0),
    SubscriptionTier.PRO: TierLimits(100, 3, 6, 5, 10),
    SubscriptionTier.PRO_PLUS: TierLimits(1100, 3, 6, 5, 10),
    SubscriptionTier.PREMIUM: TierLimits(10100, 3, 6, 5, 10),
}

class SubscriptionManager:
    """Manage subscriptions within tier limits"""
    
    def __init__(self, tier: SubscriptionTier):
        self.tier = tier
        self.limits = TIER_LIMITS[tier]
        
        # Track active subscriptions
        self.market_data: Dict[int, Contract] = {}
        self.scanners: Set[int] = set()
        self.historical: Set[int] = set()
        self.realtime_bars: Set[int] = set()
        self.market_depth: Set[int] = set()
        
        # LRU cache for market data
        self.access_times: Dict[int, datetime] = {}
        
    async def can_subscribe_market_data(self) -> bool:
        """Check if new market data subscription allowed"""
        return len(self.market_data) < self.limits.market_data_lines
    
    async def subscribe_market_data(self, req_id: int, contract: Contract):
        """Subscribe to market data with limit checking"""
        if not await self.can_subscribe_market_data():
            # Remove least recently used
            await self._evict_lru_market_data()
        
        self.market_data[req_id] = contract
        self.access_times[req_id] = datetime.now()
        
    async def _evict_lru_market_data(self):
        """Remove least recently used market data subscription"""
        if not self.access_times:
            return
            
        # Find LRU subscription
        lru_id = min(self.access_times.items(), key=lambda x: x[1])[0]
        
        # Cancel subscription
        await self.client.cancelMktData(lru_id)
        
        # Remove from tracking
        del self.market_data[lru_id]
        del self.access_times[lru_id]
        
        logger.info(f"Evicted market data subscription {lru_id}")
```

### 2. Priority-Based Subscription System

```python
from queue import PriorityQueue

class PrioritySubscription:
    """Priority-based subscription management"""
    
    def __init__(self, max_subscriptions: int):
        self.max_subscriptions = max_subscriptions
        self.active = {}  # req_id -> (priority, contract)
        self.priority_queue = PriorityQueue()
        
    async def subscribe(self, contract: Contract, priority: int = 5):
        """Subscribe with priority (1=highest, 10=lowest)"""
        req_id = self._get_next_req_id()
        
        # Check if at limit
        if len(self.active) >= self.max_subscriptions:
            # Find lowest priority subscription
            lowest_priority_id = self._find_lowest_priority()
            
            if self.active[lowest_priority_id][0] >= priority:
                # New subscription has higher priority
                await self._cancel_subscription(lowest_priority_id)
            else:
                # Cannot subscribe - priority too low
                raise Exception("Subscription limit reached, priority too low")
        
        # Add subscription
        self.active[req_id] = (priority, contract)
        await self.client.reqMktData(req_id, contract, "", False)
        
        return req_id
    
    def _find_lowest_priority(self) -> int:
        """Find subscription with lowest priority"""
        return min(self.active.items(), key=lambda x: x[1][0])[0]
```

### 3. Dynamic Subscription Optimization

```python
class DynamicSubscriptionOptimizer:
    """Optimize subscriptions based on usage patterns"""
    
    def __init__(self, subscription_manager: SubscriptionManager):
        self.sub_manager = subscription_manager
        self.usage_stats = defaultdict(lambda: {
            'access_count': 0,
            'last_access': None,
            'data_received': 0
        })
        
    async def optimize_subscriptions(self):
        """Periodically optimize subscription allocation"""
        while True:
            await asyncio.sleep(60)  # Run every minute
            
            # Analyze usage patterns
            usage_scores = self._calculate_usage_scores()
            
            # Identify underutilized subscriptions
            underutilized = [
                req_id for req_id, score in usage_scores.items()
                if score < 0.1  # Threshold
            ]
            
            # Cancel underutilized subscriptions
            for req_id in underutilized:
                await self.sub_manager.cancel_subscription(req_id)
                logger.info(f"Cancelled underutilized subscription {req_id}")
    
    def _calculate_usage_scores(self) -> Dict[int, float]:
        """Calculate usage score for each subscription"""
        scores = {}
        now = datetime.now()
        
        for req_id, stats in self.usage_stats.items():
            # Factors: access frequency, recency, data volume
            if stats['last_access']:
                recency = (now - stats['last_access']).total_seconds()
                score = (
                    stats['access_count'] / max(recency, 1) * 
                    min(stats['data_received'] / 1000, 1)
                )
            else:
                score = 0
                
            scores[req_id] = score
            
        return scores
```

### 4. Scanner Quota Management

```python
class ScannerQuotaManager:
    """Manage limited scanner slots efficiently"""
    
    def __init__(self, max_concurrent: int = 3):
        self.max_concurrent = max_concurrent
        self.active_scanners = {}
        self.scanner_queue = asyncio.Queue()
        self.scanner_semaphore = asyncio.Semaphore(max_concurrent)
        
    async def request_scanner(self, params: ScannerSubscription) -> List[Contract]:
        """Request scanner with queuing"""
        # Add to queue
        future = asyncio.Future()
        await self.scanner_queue.put((params, future))
        
        # Process queue
        asyncio.create_task(self._process_scanner_queue())
        
        # Wait for result
        return await future
    
    async def _process_scanner_queue(self):
        """Process queued scanner requests"""
        while not self.scanner_queue.empty():
            params, future = await self.scanner_queue.get()
            
            # Wait for available slot
            async with self.scanner_semaphore:
                try:
                    results = await self._execute_scan(params)
                    future.set_result(results)
                except Exception as e:
                    future.set_exception(e)
    
    async def _execute_scan(self, params: ScannerSubscription) -> List[Contract]:
        """Execute scanner request"""
        req_id = self._get_next_req_id()
        self.active_scanners[req_id] = params
        
        try:
            # Perform scan
            results = await self.client.reqScannerSubscription(req_id, params)
            return results
        finally:
            # Always cleanup
            await self.client.cancelScannerSubscription(req_id)
            del self.active_scanners[req_id]
```

### 5. Subscription Cost Optimization

```python
class SubscriptionCostOptimizer:
    """Optimize subscription costs based on usage"""
    
    def __init__(self):
        self.subscription_costs = {
            'base': 0,
            'snapshot_bundle': 10,
            'streaming_bundle': 4.50,
            'market_depth': 30  # per exchange
        }
        self.usage_history = defaultdict(list)
        
    def analyze_subscription_needs(self, 
                                 daily_market_data_requests: int,
                                 daily_scanner_requests: int,
                                 requires_realtime: bool) -> Dict[str, bool]:
        """Recommend optimal subscription mix"""
        recommendations = {
            'snapshot_bundle': False,
            'streaming_bundle': False,
            'market_depth': []
        }
        
        # If exceeding base limits regularly
        if daily_market_data_requests > 100:
            if daily_market_data_requests > 1000:
                # Need streaming bundle
                recommendations['streaming_bundle'] = True
            else:
                # Snapshot bundle sufficient
                recommendations['snapshot_bundle'] = True
        
        # Calculate potential savings
        current_overage_fees = self._calculate_overage_fees(
            daily_market_data_requests
        )
        
        recommended_cost = sum(
            self.subscription_costs[sub] 
            for sub, needed in recommendations.items() 
            if needed and sub != 'market_depth'
        )
        
        savings = current_overage_fees - recommended_cost
        
        return {
            'recommendations': recommendations,
            'monthly_savings': savings,
            'break_even_requests': self._calculate_break_even()
        }
```

### 6. Real-time Monitoring Dashboard

```python
class SubscriptionMonitor:
    """Real-time subscription monitoring"""
    
    def __init__(self):
        self.metrics = {
            'market_data_used': Gauge('tws_market_data_lines_used'),
            'market_data_limit': Gauge('tws_market_data_lines_limit'),
            'scanner_queue_depth': Gauge('tws_scanner_queue_depth'),
            'subscription_evictions': Counter('tws_subscription_evictions'),
            'subscription_denials': Counter('tws_subscription_denials')
        }
        
    async def update_metrics(self, sub_manager: SubscriptionManager):
        """Update Prometheus metrics"""
        self.metrics['market_data_used'].set(len(sub_manager.market_data))
        self.metrics['market_data_limit'].set(sub_manager.limits.market_data_lines)
        
        # Alert if approaching limits
        usage_percent = len(sub_manager.market_data) / sub_manager.limits.market_data_lines
        
        if usage_percent > 0.9:
            logger.warning(f"Market data usage at {usage_percent*100:.1f}% of limit")
        elif usage_percent > 0.95:
            logger.error("CRITICAL: Market data limit nearly reached!")
```

## Best Practices

### 1. Subscription Lifecycle
- Always unsubscribe when done
- Implement automatic cleanup for stale subscriptions
- Use connection pools for different data types

### 2. Cost Management
- Monitor daily usage patterns
- Set alerts for unusual spikes
- Review subscription needs monthly

### 3. Performance Optimization
- Batch similar requests
- Cache frequently accessed data
- Use snapshot data when real-time not required

### 4. Error Handling
```python
SUBSCRIPTION_ERROR_CODES = {
    10090: "Part of requested market data is not subscribed",
    10091: "Max number of market data subscriptions exceeded",
    10092: "Subscription request for invalid security",
    10093: "Market data farm connection is broken",
    10094: "Market data farm is not available"
}
```

## Subscription Limit Quick Reference

| Feature | Lite | Pro | Pro+Snapshot | Pro+Streaming |
|---------|------|-----|--------------|---------------|
| Market Data Lines | 100 | 100 | 1,100 | 10,100 |
| Concurrent Scanners | 1 | 3 | 3 | 3 |
| Historical Data | Limited | 6 | 6 | 6 |
| Real-time Bars | 0 | 5 | 5 | 5 |
| Market Depth | 0 | 10 | 10 | 10 |
| Monthly Cost | $0 | $0 | $10 | $14.50 |