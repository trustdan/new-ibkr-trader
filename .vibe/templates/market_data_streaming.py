# Market Data Streaming Patterns
from ib_insync import IB, Stock, Option, Contract, util
import asyncio
from collections import OrderedDict
import logging

class MarketDataStreamer:
    """Template for efficient market data streaming with subscription management"""
    
    def __init__(self, ib: IB, max_subscriptions=90):
        self.ib = ib
        self.max_subscriptions = max_subscriptions
        self.active_subscriptions = OrderedDict()  # LRU tracking
        self.tickers = {}  # Contract -> Ticker mapping
        
    async def subscribe_option_chain(self, symbol: str, exchange='SMART'):
        """Subscribe to option chain with smart subscription management"""
        
        # Get the underlying stock first
        stock = Stock(symbol, exchange, 'USD')
        await self.ib.qualifyContractsAsync(stock)
        
        # Get option chain
        chains = await self.ib.reqSecDefOptParamsAsync(
            stock.symbol, 
            stock.exchange, 
            stock.secType, 
            stock.conId
        )
        
        if not chains:
            logging.warning(f"No option chains found for {symbol}")
            return []
        
        chain = chains[0]
        
        # Create option contracts for desired strikes/expirations
        options = []
        for strike in chain.strikes:
            for expiry in chain.expirations[:3]:  # Limit to next 3 expirations
                for right in ['C', 'P']:
                    opt = Option(
                        symbol, 
                        expiry, 
                        strike, 
                        right, 
                        exchange
                    )
                    options.append(opt)
        
        # Subscribe to market data efficiently
        subscribed = []
        for opt in options:
            if await self._subscribe_with_limit(opt):
                subscribed.append(opt)
                
        return subscribed
        
    async def _subscribe_with_limit(self, contract: Contract):
        """Subscribe with automatic eviction if at limit"""
        
        key = self._get_contract_key(contract)
        
        # Already subscribed? Just update LRU
        if key in self.active_subscriptions:
            self.active_subscriptions.move_to_end(key)
            return True
            
        # At limit? Evict oldest
        if len(self.active_subscriptions) >= self.max_subscriptions:
            oldest_key, oldest_contract = self.active_subscriptions.popitem(False)
            self.ib.cancelMktData(oldest_contract)
            del self.tickers[oldest_key]
            logging.info(f"Evicted subscription: {oldest_key}")
            
        # Subscribe to market data
        ticker = self.ib.reqMktData(
            contract,
            genericTickList='',
            snapshot=False,
            regulatorySnapshot=False
        )
        
        # Store references
        self.active_subscriptions[key] = contract
        self.tickers[key] = ticker
        
        return True
        
    def _get_contract_key(self, contract: Contract):
        """Generate unique key for contract"""
        if contract.secType == 'OPT':
            return f"{contract.symbol}_{contract.lastTradeDateOrContractMonth}_{contract.strike}_{contract.right}"
        else:
            return f"{contract.symbol}_{contract.secType}_{contract.exchange}"
            
    async def stream_updates(self, callback):
        """Stream market data updates to callback function"""
        
        while True:
            # Process pending tickers
            await self.ib.pendingTickersEvent.wait()
            
            for ticker in self.ib.pendingTickers():
                if ticker.last is not None:  # Has valid price
                    await callback(ticker)
                    
            # Small delay to prevent tight loop
            await asyncio.sleep(0.1)
            
    def get_subscription_stats(self):
        """Get current subscription statistics"""
        return {
            'active': len(self.active_subscriptions),
            'max': self.max_subscriptions,
            'usage_pct': (len(self.active_subscriptions) / self.max_subscriptions) * 100,
            'contracts': list(self.active_subscriptions.keys())
        }

# Example: Greeks calculation pattern
class GreeksCalculator:
    """Template for calculating option Greeks"""
    
    def __init__(self, ib: IB):
        self.ib = ib
        self.greek_cache = {}  # Cache computed Greeks
        
    async def get_option_greeks(self, contract: Option):
        """Get Greeks for an option contract"""
        
        # Check cache first
        cache_key = f"{contract.symbol}_{contract.strike}_{contract.right}"
        if cache_key in self.greek_cache:
            cached = self.greek_cache[cache_key]
            if cached['timestamp'] > asyncio.get_event_loop().time() - 60:
                return cached['greeks']
                
        # Request market data for Greeks calculation
        ticker = self.ib.reqMktData(contract, '106', False, False)
        
        # Wait for Greeks to populate
        max_wait = 5
        start = asyncio.get_event_loop().time()
        
        while asyncio.get_event_loop().time() - start < max_wait:
            if all([
                ticker.modelGreeks.delta is not None,
                ticker.modelGreeks.gamma is not None,
                ticker.modelGreeks.theta is not None,
                ticker.modelGreeks.vega is not None
            ]):
                break
            await asyncio.sleep(0.1)
            
        # Extract Greeks
        greeks = {
            'delta': ticker.modelGreeks.delta,
            'gamma': ticker.modelGreeks.gamma,
            'theta': ticker.modelGreeks.theta,
            'vega': ticker.modelGreeks.vega,
            'iv': ticker.modelGreeks.impliedVol
        }
        
        # Cache the results
        self.greek_cache[cache_key] = {
            'greeks': greeks,
            'timestamp': asyncio.get_event_loop().time()
        }
        
        # Cancel market data subscription
        self.ib.cancelMktData(contract)
        
        return greeks

# Example usage
async def stream_option_data():
    """Example of streaming option data with Greeks"""
    
    ib = IB()
    await ib.connectAsync('localhost', 7497, clientId=1)
    
    streamer = MarketDataStreamer(ib)
    calculator = GreeksCalculator(ib)
    
    # Subscribe to SPY options
    options = await streamer.subscribe_option_chain('SPY')
    logging.info(f"Subscribed to {len(options)} option contracts")
    
    # Define update handler
    async def handle_update(ticker):
        if ticker.contract.secType == 'OPT' and ticker.last:
            # Get Greeks for significant price moves
            greeks = await calculator.get_option_greeks(ticker.contract)
            
            logging.info(
                f"{ticker.contract.symbol} "
                f"{ticker.contract.strike}{ticker.contract.right}: "
                f"${ticker.last:.2f} "
                f"Delta={greeks['delta']:.3f} "
                f"IV={greeks['iv']:.1%}"
            )
    
    # Stream updates
    try:
        await streamer.stream_updates(handle_update)
    except KeyboardInterrupt:
        logging.info("Stopping stream...")
    finally:
        ib.disconnect()

if __name__ == '__main__':
    util.run(stream_option_data())