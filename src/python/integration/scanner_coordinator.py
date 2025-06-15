"""
Scanner Coordinator - Orchestrates data flow between IBKR and Go Scanner

This module coordinates the flow of market data from IBKR through the Go scanner
and back to trading execution.
"""

import asyncio
from typing import Dict, List, Optional, Set
from datetime import datetime, timedelta
import logging
from dataclasses import dataclass
from collections import defaultdict

from ..ibkr_connector.connection import IBKRConnection
from ..scanner_client import ScannerClient, ScanRequest, ScanFilter, FilterType, VerticalSpread
from ..core.events import EventType, Event
from .backpressure import BackpressureHandler, BackpressureStrategy

logger = logging.getLogger(__name__)


@dataclass
class ScanJob:
    """Represents a scanning job"""
    id: str
    symbol: str
    request: ScanRequest
    created_at: datetime
    status: str = "pending"
    result: Optional[List[VerticalSpread]] = None
    error: Optional[str] = None


class ScannerCoordinator:
    """
    Coordinates between IBKR market data and Go scanner service
    
    Responsibilities:
    - Fetch option chains from IBKR
    - Send to Go scanner for analysis
    - Handle backpressure and rate limiting
    - Manage scan job lifecycle
    """
    
    def __init__(
        self,
        ibkr_connection: IBKRConnection,
        scanner_client: ScannerClient,
        max_concurrent_scans: int = 5,
        scan_cache_ttl: int = 300,  # 5 minutes
        backpressure_strategy: BackpressureStrategy = BackpressureStrategy.ADAPTIVE
    ):
        self.ibkr = ibkr_connection
        self.scanner = scanner_client
        self.max_concurrent_scans = max_concurrent_scans
        self.scan_cache_ttl = scan_cache_ttl
        
        # Job management
        self.active_jobs: Dict[str, ScanJob] = {}
        self.job_queue: asyncio.Queue = asyncio.Queue()
        self.scan_semaphore = asyncio.Semaphore(max_concurrent_scans)
        
        # Backpressure handling
        self.backpressure = BackpressureHandler(
            strategy=backpressure_strategy,
            requests_per_second=10.0,  # Start conservative
            burst_size=20,
            queue_size=100
        )
        
        # Cache for recent scans
        self.scan_cache: Dict[str, List[VerticalSpread]] = {}
        self.cache_timestamps: Dict[str, datetime] = {}
        
        # Metrics
        self.metrics = {
            "total_scans": 0,
            "successful_scans": 0,
            "failed_scans": 0,
            "cache_hits": 0,
            "average_scan_time": 0.0
        }
        
        # Background tasks
        self._tasks: Set[asyncio.Task] = set()
        self._running = False
        
    async def start(self):
        """Start the coordinator"""
        self._running = True
        
        # Start worker tasks
        for i in range(self.max_concurrent_scans):
            task = asyncio.create_task(self._scan_worker(i))
            self._tasks.add(task)
            
        # Start cache cleanup task
        cleanup_task = asyncio.create_task(self._cache_cleanup_worker())
        self._tasks.add(cleanup_task)
        
        logger.info(
            f"Scanner coordinator started with {self.max_concurrent_scans} workers"
        )
        
    async def stop(self):
        """Stop the coordinator"""
        self._running = False
        
        # Cancel all tasks
        for task in self._tasks:
            task.cancel()
            
        # Wait for tasks to complete
        await asyncio.gather(*self._tasks, return_exceptions=True)
        self._tasks.clear()
        
        logger.info("Scanner coordinator stopped")
        
    async def scan_symbol(
        self,
        symbol: str,
        filters: List[ScanFilter],
        use_cache: bool = True
    ) -> List[VerticalSpread]:
        """
        Scan a symbol for vertical spreads
        
        Args:
            symbol: Stock symbol to scan
            filters: List of filters to apply
            use_cache: Whether to use cached results
            
        Returns:
            List of vertical spreads matching filters
        """
        # Check cache first
        if use_cache:
            cache_key = self._get_cache_key(symbol, filters)
            if cache_key in self.scan_cache:
                timestamp = self.cache_timestamps.get(cache_key)
                if timestamp and (datetime.now() - timestamp).seconds < self.scan_cache_ttl:
                    self.metrics["cache_hits"] += 1
                    logger.info(f"Cache hit for {symbol}")
                    return self.scan_cache[cache_key]
                    
        # Create scan request
        request = ScanRequest(
            symbol=symbol,
            filters=filters,
            limit=100
        )
        
        # Submit job
        job = ScanJob(
            id=f"{symbol}_{datetime.now().timestamp()}",
            symbol=symbol,
            request=request,
            created_at=datetime.now()
        )
        
        await self.job_queue.put(job)
        self.active_jobs[job.id] = job
        
        # Wait for completion
        while job.status == "pending":
            await asyncio.sleep(0.1)
            
        if job.status == "completed" and job.result:
            # Cache results
            cache_key = self._get_cache_key(symbol, filters)
            self.scan_cache[cache_key] = job.result
            self.cache_timestamps[cache_key] = datetime.now()
            
            return job.result
        elif job.error:
            raise Exception(f"Scan failed: {job.error}")
        else:
            return []
            
    async def _scan_worker(self, worker_id: int):
        """Background worker for processing scan jobs"""
        logger.info(f"Scan worker {worker_id} started")
        
        while self._running:
            try:
                # Get job from queue
                job = await asyncio.wait_for(
                    self.job_queue.get(),
                    timeout=1.0
                )
                
                # Acquire semaphore and backpressure permit
                async with self.scan_semaphore:
                    # Wait for backpressure to allow request
                    await self.backpressure.wait_if_needed()
                    await self._process_scan_job(job)
                    
            except asyncio.TimeoutError:
                continue
            except Exception as e:
                logger.error(f"Worker {worker_id} error: {e}")
                
        logger.info(f"Scan worker {worker_id} stopped")
        
    async def _process_scan_job(self, job: ScanJob):
        """Process a single scan job"""
        start_time = datetime.now()
        
        try:
            logger.info(f"Processing scan job {job.id} for {job.symbol}")
            job.status = "processing"
            
            # Step 1: Fetch option chain data from IBKR
            option_data = await self._fetch_option_chain(job.symbol)
            
            if not option_data:
                raise Exception("No option data available")
                
            # Step 2: Send to scanner (scanner will handle the filtering)
            scan_start = datetime.now()
            spreads = await self.scanner.scan(job.request)
            scan_duration = (datetime.now() - scan_start).total_seconds()
            
            # Record backpressure metrics
            self.backpressure.record_request(
                duration=scan_duration,
                success=True,
                queue_time=(scan_start - job.created_at).total_seconds()
            )
            
            # Update job
            job.status = "completed"
            job.result = spreads
            
            # Update metrics
            self.metrics["total_scans"] += 1
            self.metrics["successful_scans"] += 1
            
            scan_time = (datetime.now() - start_time).total_seconds()
            self._update_average_scan_time(scan_time)
            
            logger.info(
                f"Scan completed for {job.symbol}: "
                f"{len(spreads)} spreads found in {scan_time:.2f}s"
            )
            
            # Emit event
            await self.ibkr.emit_event(Event(
                type=EventType.SCAN_COMPLETED,
                data={
                    "symbol": job.symbol,
                    "spreads_found": len(spreads),
                    "scan_time": scan_time
                }
            ))
            
        except Exception as e:
            logger.error(f"Scan job {job.id} failed: {e}")
            job.status = "failed"
            job.error = str(e)
            self.metrics["failed_scans"] += 1
            
            # Record backpressure failure
            scan_duration = (datetime.now() - start_time).total_seconds()
            self.backpressure.record_request(
                duration=scan_duration,
                success=False,
                queue_time=(start_time - job.created_at).total_seconds()
            )
            
            # Emit error event
            await self.ibkr.emit_event(Event(
                type=EventType.ERROR,
                data={
                    "error_type": "scan_failed",
                    "symbol": job.symbol,
                    "error": str(e)
                }
            ))
            
    async def _fetch_option_chain(self, symbol: str) -> Dict:
        """
        Fetch option chain data from IBKR
        
        Note: This is a placeholder - actual implementation would use
        the IBKR connection to fetch real option data
        """
        # TODO: Implement actual IBKR option chain fetching
        logger.info(f"Fetching option chain for {symbol}")
        
        # For now, return empty dict - scanner has mock data
        return {}
        
    async def _cache_cleanup_worker(self):
        """Periodically clean up expired cache entries"""
        while self._running:
            try:
                await asyncio.sleep(60)  # Run every minute
                
                now = datetime.now()
                expired_keys = []
                
                for key, timestamp in self.cache_timestamps.items():
                    if (now - timestamp).seconds > self.scan_cache_ttl:
                        expired_keys.append(key)
                        
                for key in expired_keys:
                    del self.scan_cache[key]
                    del self.cache_timestamps[key]
                    
                if expired_keys:
                    logger.info(f"Cleaned up {len(expired_keys)} expired cache entries")
                    
            except Exception as e:
                logger.error(f"Cache cleanup error: {e}")
                
    def _get_cache_key(self, symbol: str, filters: List[ScanFilter]) -> str:
        """Generate cache key from symbol and filters"""
        filter_str = "_".join(
            f"{f.type.value}_{hash(str(f.params))}" 
            for f in sorted(filters, key=lambda x: x.type.value)
        )
        return f"{symbol}_{filter_str}"
        
    def _update_average_scan_time(self, scan_time: float):
        """Update rolling average scan time"""
        total_scans = self.metrics["successful_scans"]
        current_avg = self.metrics["average_scan_time"]
        
        # Calculate new average
        new_avg = ((current_avg * (total_scans - 1)) + scan_time) / total_scans
        self.metrics["average_scan_time"] = new_avg
        
    def get_metrics(self) -> Dict:
        """Get coordinator metrics"""
        return {
            **self.metrics,
            "active_jobs": len(self.active_jobs),
            "queued_jobs": self.job_queue.qsize(),
            "cache_size": len(self.scan_cache),
            "backpressure": self.backpressure.get_metrics()
        }


# Example usage
async def example_coordinator():
    """Example of using the scanner coordinator"""
    
    # Mock IBKR connection
    class MockIBKR:
        async def emit_event(self, event):
            print(f"Event: {event.type} - {event.data}")
            
    ibkr = MockIBKR()
    
    # Create coordinator
    async with ScannerClient() as scanner_client:
        coordinator = ScannerCoordinator(
            ibkr_connection=ibkr,
            scanner_client=scanner_client
        )
        
        await coordinator.start()
        
        try:
            # Run a scan
            filters = [
                ScanFilter(
                    type=FilterType.DELTA,
                    params={"min": 0.25, "max": 0.35}
                ),
                ScanFilter(
                    type=FilterType.DTE,
                    params={"min": 45, "max": 60}
                )
            ]
            
            spreads = await coordinator.scan_symbol("AAPL", filters)
            
            print(f"Found {len(spreads)} spreads")
            for spread in spreads[:3]:
                print(f"Score: {spread.score:.2f}, PoP: {spread.probability_profit:.1%}")
                
            # Get metrics
            metrics = coordinator.get_metrics()
            print(f"\nMetrics: {metrics}")
            
        finally:
            await coordinator.stop()


if __name__ == "__main__":
    asyncio.run(example_coordinator())