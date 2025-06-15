package filters

import (
	"context"
	"runtime"
	"sync"
	"time"
	"github.com/ibkr-trader/scanner/internal/models"
)

// BatchProcessor handles efficient filtering of large datasets
type BatchProcessor struct {
	batchSize     int
	workers       int
	filterChain   *AdvancedFilterChain
	progressChan  chan BatchProgress
	resultBuffer  int
}

// BatchProgress reports processing progress
type BatchProgress struct {
	ProcessedItems int
	TotalItems     int
	FilteredItems  int
	BatchNumber    int
	TotalBatches   int
}

// BatchResult contains filtered results from a batch
type BatchResult struct {
	Contracts []models.OptionContract
	Spreads   []models.VerticalSpread
	BatchID   int
}

// NewBatchProcessor creates an optimized batch processor
func NewBatchProcessor(filterChain *AdvancedFilterChain, options ...BatchOption) *BatchProcessor {
	bp := &BatchProcessor{
		batchSize:    1000,
		workers:      runtime.NumCPU(),
		filterChain:  filterChain,
		resultBuffer: 10,
	}
	
	// Apply options
	for _, opt := range options {
		opt(bp)
	}
	
	return bp
}

// BatchOption configures the batch processor
type BatchOption func(*BatchProcessor)

// WithBatchSize sets the batch size
func WithBatchSize(size int) BatchOption {
	return func(bp *BatchProcessor) {
		bp.batchSize = size
	}
}

// WithWorkers sets the number of workers
func WithWorkers(workers int) BatchOption {
	return func(bp *BatchProcessor) {
		bp.workers = workers
	}
}

// WithProgressReporting enables progress reporting
func WithProgressReporting(progressChan chan BatchProgress) BatchOption {
	return func(bp *BatchProcessor) {
		bp.progressChan = progressChan
	}
}

// ProcessContracts processes contracts in batches
func (bp *BatchProcessor) ProcessContracts(ctx context.Context, contracts []models.OptionContract) []models.OptionContract {
	if len(contracts) <= bp.batchSize {
		// Small dataset, process directly
		return bp.filterChain.ApplyToContracts(contracts)
	}
	
	// Calculate batches
	totalBatches := (len(contracts) + bp.batchSize - 1) / bp.batchSize
	
	// Create channels
	jobChan := make(chan contractBatch, totalBatches)
	resultChan := make(chan []models.OptionContract, bp.resultBuffer)
	
	// Start workers
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	for i := 0; i < bp.workers; i++ {
		wg.Add(1)
		go bp.contractWorker(ctx, &wg, jobChan, resultChan)
	}
	
	// Send batches
	go func() {
		for i := 0; i < len(contracts); i += bp.batchSize {
			end := i + bp.batchSize
			if end > len(contracts) {
				end = len(contracts)
			}
			
			batch := contractBatch{
				contracts:   contracts[i:end],
				batchNumber: i / bp.batchSize,
				totalBatches: totalBatches,
				startIndex:  i,
			}
			
			select {
			case jobChan <- batch:
			case <-ctx.Done():
				return
			}
		}
		close(jobChan)
	}()
	
	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Aggregate results
	results := make([]models.OptionContract, 0, len(contracts)/2) // Estimate 50% pass rate
	processedCount := 0
	
	for filtered := range resultChan {
		results = append(results, filtered...)
		processedCount += bp.batchSize
		
		// Report progress if enabled
		if bp.progressChan != nil {
			progress := BatchProgress{
				ProcessedItems: processedCount,
				TotalItems:     len(contracts),
				FilteredItems:  len(results),
			}
			
			select {
			case bp.progressChan <- progress:
			default:
				// Don't block on progress reporting
			}
		}
	}
	
	return results
}

// contractWorker processes contract batches
func (bp *BatchProcessor) contractWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan contractBatch, results chan<- []models.OptionContract) {
	defer wg.Done()
	
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			
			// Process batch
			filtered := bp.filterChain.ApplyToContracts(job.contracts)
			
			// Send results
			select {
			case results <- filtered:
			case <-ctx.Done():
				return
			}
			
			// Report progress for this batch
			if bp.progressChan != nil {
				progress := BatchProgress{
					BatchNumber:  job.batchNumber,
					TotalBatches: job.totalBatches,
				}
				
				select {
				case bp.progressChan <- progress:
				default:
				}
			}
			
		case <-ctx.Done():
			return
		}
	}
}

// ProcessSpreads processes spreads in batches
func (bp *BatchProcessor) ProcessSpreads(ctx context.Context, spreads []models.VerticalSpread) []models.VerticalSpread {
	if len(spreads) <= bp.batchSize {
		return bp.filterChain.ApplyToSpreads(spreads)
	}
	
	// Similar batch processing for spreads
	totalBatches := (len(spreads) + bp.batchSize - 1) / bp.batchSize
	
	jobChan := make(chan spreadBatch, totalBatches)
	resultChan := make(chan []models.VerticalSpread, bp.resultBuffer)
	
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Start workers
	for i := 0; i < bp.workers; i++ {
		wg.Add(1)
		go bp.spreadWorker(ctx, &wg, jobChan, resultChan)
	}
	
	// Send batches
	go func() {
		for i := 0; i < len(spreads); i += bp.batchSize {
			end := i + bp.batchSize
			if end > len(spreads) {
				end = len(spreads)
			}
			
			batch := spreadBatch{
				spreads:      spreads[i:end],
				batchNumber:  i / bp.batchSize,
				totalBatches: totalBatches,
			}
			
			select {
			case jobChan <- batch:
			case <-ctx.Done():
				return
			}
		}
		close(jobChan)
	}()
	
	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Aggregate results
	results := make([]models.VerticalSpread, 0, len(spreads)/2)
	for filtered := range resultChan {
		results = append(results, filtered...)
	}
	
	return results
}

// spreadWorker processes spread batches
func (bp *BatchProcessor) spreadWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan spreadBatch, results chan<- []models.VerticalSpread) {
	defer wg.Done()
	
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			
			filtered := bp.filterChain.ApplyToSpreads(job.spreads)
			
			select {
			case results <- filtered:
			case <-ctx.Done():
				return
			}
			
		case <-ctx.Done():
			return
		}
	}
}

// StreamingProcessor processes data in a streaming fashion
type StreamingProcessor struct {
	filterChain  *AdvancedFilterChain
	bufferSize   int
	flushTimeout time.Duration
}

// NewStreamingProcessor creates a streaming processor
func NewStreamingProcessor(filterChain *AdvancedFilterChain, bufferSize int) *StreamingProcessor {
	return &StreamingProcessor{
		filterChain:  filterChain,
		bufferSize:   bufferSize,
		flushTimeout: 100 * time.Millisecond,
	}
}

// ProcessContractStream processes a stream of contracts
func (sp *StreamingProcessor) ProcessContractStream(ctx context.Context, input <-chan models.OptionContract) <-chan models.OptionContract {
	output := make(chan models.OptionContract, sp.bufferSize)
	
	go func() {
		defer close(output)
		
		buffer := make([]models.OptionContract, 0, sp.bufferSize)
		ticker := time.NewTicker(sp.flushTimeout)
		defer ticker.Stop()
		
		for {
			select {
			case contract, ok := <-input:
				if !ok {
					// Input closed, process remaining buffer
					if len(buffer) > 0 {
						filtered := sp.filterChain.ApplyToContracts(buffer)
						for _, c := range filtered {
							select {
							case output <- c:
							case <-ctx.Done():
								return
							}
						}
					}
					return
				}
				
				buffer = append(buffer, contract)
				
				// Process when buffer is full
				if len(buffer) >= sp.bufferSize {
					filtered := sp.filterChain.ApplyToContracts(buffer)
					buffer = buffer[:0] // Reset buffer
					
					for _, c := range filtered {
						select {
						case output <- c:
						case <-ctx.Done():
							return
						}
					}
				}
				
			case <-ticker.C:
				// Flush buffer periodically
				if len(buffer) > 0 {
					filtered := sp.filterChain.ApplyToContracts(buffer)
					buffer = buffer[:0]
					
					for _, c := range filtered {
						select {
						case output <- c:
						case <-ctx.Done():
							return
						}
					}
				}
				
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return output
}

// Internal batch types
type contractBatch struct {
	contracts    []models.OptionContract
	batchNumber  int
	totalBatches int
	startIndex   int
}

type spreadBatch struct {
	spreads      []models.VerticalSpread
	batchNumber  int
	totalBatches int
}

// OptimizedFilterChain provides memory-efficient filtering
type OptimizedFilterChain struct {
	*AdvancedFilterChain
	indexPool    sync.Pool
	contractPool sync.Pool
}

// NewOptimizedFilterChain creates an optimized filter chain
func NewOptimizedFilterChain(config FilterConfig) *OptimizedFilterChain {
	return &OptimizedFilterChain{
		AdvancedFilterChain: NewAdvancedFilterChain(config, true, true),
		indexPool: sync.Pool{
			New: func() interface{} {
				return make([]int, 0, 1000)
			},
		},
		contractPool: sync.Pool{
			New: func() interface{} {
				return make([]models.OptionContract, 0, 1000)
			},
		},
	}
}

// ApplyToContractsOptimized applies filters with memory optimization
func (ofc *OptimizedFilterChain) ApplyToContractsOptimized(contracts []models.OptionContract) []models.OptionContract {
	// Get pooled slice
	indices := ofc.indexPool.Get().([]int)
	defer func() {
		indices = indices[:0]
		ofc.indexPool.Put(indices)
	}()
	
	// Mark passing indices instead of copying
	for i, contract := range contracts {
		passed := true
		
		for _, filter := range ofc.contractFilters {
			// Apply filter to single contract
			temp := []models.OptionContract{contract}
			if len(filter.Apply(temp)) == 0 {
				passed = false
				break
			}
		}
		
		if passed {
			indices = append(indices, i)
		}
	}
	
	// Get result slice from pool
	result := ofc.contractPool.Get().([]models.OptionContract)
	defer func() {
		if cap(result) > 10000 { // Don't pool very large slices
			return
		}
		result = result[:0]
		ofc.contractPool.Put(result)
	}()
	
	// Copy passing contracts
	result = result[:0]
	for _, idx := range indices {
		result = append(result, contracts[idx])
	}
	
	// Return a copy to avoid pool issues
	output := make([]models.OptionContract, len(result))
	copy(output, result)
	
	return output
}