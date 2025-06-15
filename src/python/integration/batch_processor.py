"""
Batch Processing for Scanner Requests

This module implements request batching to improve efficiency when scanning
multiple symbols or running multiple scans concurrently.
"""

import asyncio
from typing import List, Dict, Any, Optional, Callable
from datetime import datetime, timedelta
from dataclasses import dataclass, field
from collections import defaultdict
import logging

from ..scanner_client import ScannerClient, ScanRequest, VerticalSpread
from .backpressure import BackpressureHandler

logger = logging.getLogger(__name__)


@dataclass
class BatchRequest:
    """Container for a batch of scan requests"""
    id: str
    requests: List[ScanRequest]
    created_at: datetime = field(default_factory=datetime.now)
    completed_at: Optional[datetime] = None
    results: Dict[str, List[VerticalSpread]] = field(default_factory=dict)
    errors: Dict[str, str] = field(default_factory=dict)


class BatchProcessor:
    """
    Processes scan requests in batches for improved efficiency
    
    Features:
    - Automatic batching of requests
    - Parallel processing within batches
    - Result aggregation
    - Error isolation
    - Progress tracking
    """
    
    def __init__(
        self,
        scanner_client: ScannerClient,
        backpressure_handler: BackpressureHandler,
        batch_size: int = 10,
        batch_timeout: float = 0.5,  # seconds
        max_concurrent_batches: int = 3
    ):
        self.scanner = scanner_client
        self.backpressure = backpressure_handler
        self.batch_size = batch_size
        self.batch_timeout = batch_timeout
        self.max_concurrent_batches = max_concurrent_batches
        
        # Request queue and batching
        self.request_queue: asyncio.Queue = asyncio.Queue()
        self.pending_batch: List[tuple] = []  # (request, future)
        self.batch_lock = asyncio.Lock()
        
        # Batch tracking
        self.active_batches: Dict[str, BatchRequest] = {}
        self.completed_batches: Dict[str, BatchRequest] = {}
        
        # Processing control
        self._running = False
        self._tasks: List[asyncio.Task] = []
        self.batch_semaphore = asyncio.Semaphore(max_concurrent_batches)
        
        # Metrics
        self.metrics = {
            "total_requests": 0,
            "total_batches": 0,
            "successful_requests": 0,
            "failed_requests": 0,
            "average_batch_time": 0.0,
            "current_queue_size": 0
        }
        
    async def start(self):
        """Start the batch processor"""
        self._running = True
        
        # Start batch collector
        collector_task = asyncio.create_task(self._batch_collector())
        self._tasks.append(collector_task)
        
        # Start batch processors
        for i in range(self.max_concurrent_batches):
            processor_task = asyncio.create_task(self._batch_processor(i))
            self._tasks.append(processor_task)
            
        logger.info(f"Batch processor started with size={self.batch_size}")
        
    async def stop(self):
        """Stop the batch processor"""
        self._running = False
        
        # Process any remaining requests
        await self._flush_pending_batch()
        
        # Cancel tasks
        for task in self._tasks:
            task.cancel()
            
        await asyncio.gather(*self._tasks, return_exceptions=True)
        self._tasks.clear()
        
        logger.info("Batch processor stopped")
        
    async def submit_request(self, request: ScanRequest) -> List[VerticalSpread]:
        """
        Submit a scan request for batch processing
        
        Args:
            request: Scan request to process
            
        Returns:
            List of vertical spreads
        """
        future = asyncio.Future()
        
        # Add to queue
        await self.request_queue.put((request, future))
        self.metrics["total_requests"] += 1
        self.metrics["current_queue_size"] = self.request_queue.qsize()
        
        # Wait for result
        return await future
        
    async def submit_batch(self, requests: List[ScanRequest]) -> Dict[str, List[VerticalSpread]]:
        """
        Submit multiple requests as a batch
        
        Args:
            requests: List of scan requests
            
        Returns:
            Dict mapping symbol to spreads
        """
        # Submit all requests
        futures = []
        for req in requests:
            future = asyncio.Future()
            await self.request_queue.put((req, future))
            futures.append((req.symbol, future))
            
        self.metrics["total_requests"] += len(requests)
        
        # Wait for all results
        results = {}
        for symbol, future in futures:
            try:
                spreads = await future
                results[symbol] = spreads
            except Exception as e:
                logger.error(f"Batch request failed for {symbol}: {e}")
                results[symbol] = []
                
        return results
        
    async def _batch_collector(self):
        """Collects requests into batches"""
        while self._running:
            try:
                # Wait for first request or timeout
                try:
                    request, future = await asyncio.wait_for(
                        self.request_queue.get(),
                        timeout=1.0
                    )
                    
                    async with self.batch_lock:
                        self.pending_batch.append((request, future))
                        
                    # Collect more requests up to batch size
                    deadline = datetime.now() + timedelta(seconds=self.batch_timeout)
                    
                    while len(self.pending_batch) < self.batch_size:
                        remaining = (deadline - datetime.now()).total_seconds()
                        if remaining <= 0:
                            break
                            
                        try:
                            req, fut = await asyncio.wait_for(
                                self.request_queue.get(),
                                timeout=remaining
                            )
                            async with self.batch_lock:
                                self.pending_batch.append((req, fut))
                        except asyncio.TimeoutError:
                            break
                            
                    # Create and submit batch
                    await self._flush_pending_batch()
                    
                except asyncio.TimeoutError:
                    # No requests, check if we have pending batch
                    if self.pending_batch:
                        await self._flush_pending_batch()
                        
            except Exception as e:
                logger.error(f"Batch collector error: {e}")
                
    async def _flush_pending_batch(self):
        """Flush pending requests as a batch"""
        async with self.batch_lock:
            if not self.pending_batch:
                return
                
            # Create batch
            batch_id = f"batch_{datetime.now().timestamp()}"
            requests = [req for req, _ in self.pending_batch]
            futures = {req.symbol: fut for req, fut in self.pending_batch}
            
            batch = BatchRequest(
                id=batch_id,
                requests=requests
            )
            
            self.active_batches[batch_id] = batch
            self.pending_batch.clear()
            
            # Submit for processing
            asyncio.create_task(self._process_batch(batch, futures))
            
            self.metrics["total_batches"] += 1
            logger.info(f"Created batch {batch_id} with {len(requests)} requests")
            
    async def _batch_processor(self, worker_id: int):
        """Worker that processes batches"""
        logger.info(f"Batch processor {worker_id} started")
        
        while self._running:
            await asyncio.sleep(0.1)  # Small delay to prevent busy loop
            
        logger.info(f"Batch processor {worker_id} stopped")
        
    async def _process_batch(self, batch: BatchRequest, futures: Dict[str, asyncio.Future]):
        """Process a single batch of requests"""
        start_time = datetime.now()
        
        async with self.batch_semaphore:
            try:
                # Process requests in parallel
                tasks = []
                for request in batch.requests:
                    task = asyncio.create_task(
                        self._process_single_request(request, batch)
                    )
                    tasks.append((request.symbol, task))
                    
                # Wait for all tasks
                for symbol, task in tasks:
                    try:
                        await task
                        self.metrics["successful_requests"] += 1
                    except Exception as e:
                        batch.errors[symbol] = str(e)
                        self.metrics["failed_requests"] += 1
                        
                # Mark batch complete
                batch.completed_at = datetime.now()
                batch_duration = (batch.completed_at - start_time).total_seconds()
                self._update_average_batch_time(batch_duration)
                
                # Resolve futures
                for symbol, future in futures.items():
                    if symbol in batch.errors:
                        future.set_exception(Exception(batch.errors[symbol]))
                    else:
                        future.set_result(batch.results.get(symbol, []))
                        
                # Move to completed
                self.completed_batches[batch.id] = batch
                del self.active_batches[batch.id]
                
                logger.info(
                    f"Batch {batch.id} completed in {batch_duration:.2f}s: "
                    f"{len(batch.results)} successful, {len(batch.errors)} failed"
                )
                
            except Exception as e:
                logger.error(f"Batch processing failed: {e}")
                # Set all futures to error
                for future in futures.values():
                    if not future.done():
                        future.set_exception(e)
                        
    async def _process_single_request(self, request: ScanRequest, batch: BatchRequest):
        """Process a single request within a batch"""
        try:
            # Apply backpressure
            await self.backpressure.wait_if_needed()
            
            # Execute scan
            spreads = await self.scanner.scan(request)
            
            # Store results
            batch.results[request.symbol] = spreads
            
        except Exception as e:
            logger.error(f"Request failed for {request.symbol}: {e}")
            batch.errors[request.symbol] = str(e)
            raise
            
    def _update_average_batch_time(self, batch_time: float):
        """Update rolling average batch processing time"""
        total_batches = self.metrics["total_batches"]
        current_avg = self.metrics["average_batch_time"]
        
        # Calculate new average
        new_avg = ((current_avg * (total_batches - 1)) + batch_time) / total_batches
        self.metrics["average_batch_time"] = new_avg
        
    def get_metrics(self) -> Dict[str, Any]:
        """Get batch processor metrics"""
        return {
            **self.metrics,
            "active_batches": len(self.active_batches),
            "completed_batches": len(self.completed_batches),
            "pending_requests": len(self.pending_batch)
        }
        
    async def get_batch_status(self, batch_id: str) -> Optional[Dict[str, Any]]:
        """Get status of a specific batch"""
        if batch_id in self.active_batches:
            batch = self.active_batches[batch_id]
            return {
                "status": "processing",
                "requests": len(batch.requests),
                "completed": len(batch.results),
                "errors": len(batch.errors),
                "created_at": batch.created_at.isoformat()
            }
        elif batch_id in self.completed_batches:
            batch = self.completed_batches[batch_id]
            return {
                "status": "completed",
                "requests": len(batch.requests),
                "successful": len(batch.results),
                "errors": len(batch.errors),
                "created_at": batch.created_at.isoformat(),
                "completed_at": batch.completed_at.isoformat() if batch.completed_at else None,
                "duration": (batch.completed_at - batch.created_at).total_seconds() if batch.completed_at else None
            }
        else:
            return None


# Example usage
async def example_batch_processing():
    """Example of using the batch processor"""
    
    # Create components
    scanner_client = ScannerClient()
    backpressure = BackpressureHandler()
    
    batch_processor = BatchProcessor(
        scanner_client=scanner_client,
        backpressure_handler=backpressure,
        batch_size=5,
        batch_timeout=0.3
    )
    
    await batch_processor.start()
    
    try:
        # Submit individual requests (will be batched automatically)
        symbols = ["AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", 
                  "META", "NVDA", "AMD", "INTC", "NFLX"]
        
        tasks = []
        for symbol in symbols:
            request = ScanRequest(
                symbol=symbol,
                filters=[],  # Add filters as needed
                limit=10
            )
            task = batch_processor.submit_request(request)
            tasks.append(task)
            
        # Wait for all results
        results = await asyncio.gather(*tasks)
        
        print(f"Processed {len(symbols)} symbols")
        for symbol, spreads in zip(symbols, results):
            print(f"{symbol}: {len(spreads)} spreads found")
            
        # Check metrics
        metrics = batch_processor.get_metrics()
        print(f"\nBatch processor metrics:")
        print(f"  Total batches: {metrics['total_batches']}")
        print(f"  Average batch time: {metrics['average_batch_time']:.2f}s")
        
    finally:
        await batch_processor.stop()


if __name__ == "__main__":
    asyncio.run(example_batch_processing())