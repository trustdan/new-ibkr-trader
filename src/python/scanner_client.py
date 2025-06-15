"""
Scanner Client for Python-Go Integration

This module provides an async HTTP client for communicating with the Go scanner service.
It handles request/response serialization, retries, and error handling.
"""

import asyncio
import json
from typing import Dict, List, Optional, Any
from datetime import datetime
import httpx
from dataclasses import dataclass, asdict
import logging
from enum import Enum

logger = logging.getLogger(__name__)


class FilterType(str, Enum):
    """Filter types supported by the scanner"""
    DELTA = "delta"
    DTE = "dte"
    LIQUIDITY = "liquidity"
    THETA = "theta"
    VEGA = "vega"
    IV = "iv"
    IV_PERCENTILE = "iv_percentile"
    SPREAD_WIDTH = "spread_width"
    PROBABILITY_PROFIT = "probability_profit"


@dataclass
class ScanFilter:
    """Individual scan filter configuration"""
    type: FilterType
    params: Dict[str, Any]


@dataclass
class ScanRequest:
    """Request to scan for option spreads"""
    symbol: str
    filters: List[ScanFilter]
    limit: int = 100
    sort_by: str = "score"
    
    def to_dict(self) -> Dict:
        """Convert to dictionary for JSON serialization"""
        return {
            "symbol": self.symbol,
            "filters": [{"type": f.type.value, "params": f.params} for f in self.filters],
            "limit": self.limit,
            "sort_by": self.sort_by
        }


@dataclass
class OptionContract:
    """Option contract details from scanner"""
    symbol: str
    expiry: str
    strike: float
    right: str  # "C" or "P"
    delta: float
    theta: float
    vega: float
    iv: float
    volume: int
    open_interest: int
    bid: float
    ask: float
    last: float


@dataclass
class VerticalSpread:
    """Vertical spread from scanner results"""
    long_leg: OptionContract
    short_leg: OptionContract
    net_debit: float
    max_profit: float
    max_loss: float
    breakeven: float
    probability_profit: float
    score: float
    
    @classmethod
    def from_dict(cls, data: Dict) -> 'VerticalSpread':
        """Create from dictionary response"""
        return cls(
            long_leg=OptionContract(**data['long_leg']),
            short_leg=OptionContract(**data['short_leg']),
            net_debit=data['net_debit'],
            max_profit=data['max_profit'],
            max_loss=data['max_loss'],
            breakeven=data['breakeven'],
            probability_profit=data['probability_profit'],
            score=data['score']
        )


class ScannerClient:
    """HTTP client for Go scanner service"""
    
    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        timeout: float = 30.0,
        max_retries: int = 3
    ):
        self.base_url = base_url
        self.timeout = timeout
        self.max_retries = max_retries
        self._client: Optional[httpx.AsyncClient] = None
        
    async def __aenter__(self):
        """Async context manager entry"""
        self._client = httpx.AsyncClient(
            base_url=self.base_url,
            timeout=self.timeout
        )
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit"""
        if self._client:
            await self._client.aclose()
            
    async def _ensure_client(self):
        """Ensure HTTP client is initialized"""
        if not self._client:
            self._client = httpx.AsyncClient(
                base_url=self.base_url,
                timeout=self.timeout
            )
            
    async def health_check(self) -> Dict[str, Any]:
        """Check scanner service health"""
        await self._ensure_client()
        
        try:
            response = await self._client.get("/health")
            response.raise_for_status()
            return response.json()
        except httpx.HTTPError as e:
            logger.error(f"Health check failed: {e}")
            return {"status": "unhealthy", "error": str(e)}
            
    async def scan(
        self,
        request: ScanRequest,
        retry_count: int = 0
    ) -> List[VerticalSpread]:
        """
        Scan for option spreads matching filters
        
        Args:
            request: Scan configuration
            retry_count: Current retry attempt
            
        Returns:
            List of matching vertical spreads
        """
        await self._ensure_client()
        
        try:
            response = await self._client.post(
                "/api/v1/scan",
                json=request.to_dict()
            )
            response.raise_for_status()
            
            data = response.json()
            spreads = [VerticalSpread.from_dict(s) for s in data.get('spreads', [])]
            
            logger.info(
                f"Scanner returned {len(spreads)} spreads for {request.symbol}"
            )
            return spreads
            
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 429:  # Rate limited
                if retry_count < self.max_retries:
                    wait_time = 2 ** retry_count
                    logger.warning(
                        f"Rate limited, retrying in {wait_time}s..."
                    )
                    await asyncio.sleep(wait_time)
                    return await self.scan(request, retry_count + 1)
                    
            logger.error(f"Scan request failed: {e}")
            raise
            
        except httpx.HTTPError as e:
            if retry_count < self.max_retries:
                wait_time = 2 ** retry_count
                logger.warning(
                    f"Request failed, retrying in {wait_time}s: {e}"
                )
                await asyncio.sleep(wait_time)
                return await self.scan(request, retry_count + 1)
                
            logger.error(f"Scan request failed after retries: {e}")
            raise
            
    async def get_scan_status(self, scan_id: str) -> Dict[str, Any]:
        """Get status of an async scan operation"""
        await self._ensure_client()
        
        response = await self._client.get(f"/api/v1/scan/{scan_id}")
        response.raise_for_status()
        return response.json()
        
    async def cancel_scan(self, scan_id: str) -> bool:
        """Cancel an in-progress scan"""
        await self._ensure_client()
        
        try:
            response = await self._client.delete(f"/api/v1/scan/{scan_id}")
            response.raise_for_status()
            return True
        except httpx.HTTPError:
            return False
            
    async def get_metrics(self) -> Dict[str, Any]:
        """Get scanner performance metrics"""
        await self._ensure_client()
        
        response = await self._client.get("/metrics")
        response.raise_for_status()
        return response.json()


# Example usage
async def example_scan():
    """Example of using the scanner client"""
    
    # Create scan request
    request = ScanRequest(
        symbol="SPY",
        filters=[
            ScanFilter(
                type=FilterType.DELTA,
                params={"min": 0.2, "max": 0.4}
            ),
            ScanFilter(
                type=FilterType.DTE,
                params={"min": 30, "max": 60}
            ),
            ScanFilter(
                type=FilterType.LIQUIDITY,
                params={
                    "min_volume": 100,
                    "min_open_interest": 500,
                    "max_bid_ask_spread": 0.10
                }
            )
        ],
        limit=50
    )
    
    # Execute scan
    async with ScannerClient() as client:
        # Check health first
        health = await client.health_check()
        print(f"Scanner health: {health}")
        
        # Run scan
        spreads = await client.scan(request)
        
        # Display results
        for spread in spreads[:5]:  # Show top 5
            print(f"\nSpread Score: {spread.score:.2f}")
            print(f"Long: {spread.long_leg.strike} {spread.long_leg.expiry}")
            print(f"Short: {spread.short_leg.strike} {spread.short_leg.expiry}")
            print(f"Net Debit: ${spread.net_debit:.2f}")
            print(f"Max Profit: ${spread.max_profit:.2f}")
            print(f"Probability of Profit: {spread.probability_profit:.1%}")


if __name__ == "__main__":
    asyncio.run(example_scan())