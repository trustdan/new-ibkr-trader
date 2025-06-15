package filters

import (
	"fmt"
	"sync"
	"time"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

// AdvancedFilterChain provides sophisticated filter management with performance tracking
type AdvancedFilterChain struct {
	contractFilters []Filter
	spreadFilters   []SpreadFilter
	combinedFilters []CombinedFilter
	
	// Performance tracking
	metrics         *FilterMetrics
	executionStats  map[string]*FilterStats
	mu              sync.RWMutex
	
	// Configuration
	parallelExecution bool
	cacheEnabled      bool
	cache             *FilterCache
}

// CombinedFilter can filter both contracts and spreads with complex logic
type CombinedFilter interface {
	Name() string
	ApplyToCombined(contracts []models.OptionContract, spreads []models.VerticalSpread) ([]models.OptionContract, []models.VerticalSpread)
	Validate() error
}

// FilterStats tracks execution statistics
type FilterStats struct {
	ExecutionCount   int64
	TotalDuration    time.Duration
	AverageDuration  time.Duration
	ItemsProcessed   int64
	ItemsFiltered    int64
	LastExecution    time.Time
}

// FilterMetrics for Prometheus monitoring
type FilterMetrics struct {
	filterExecutionTime   *prometheus.HistogramVec
	filterItemsProcessed  *prometheus.CounterVec
	filterItemsRemoved    *prometheus.CounterVec
	cacheHitRate         prometheus.Gauge
}

// NewAdvancedFilterChain creates an advanced filter chain
func NewAdvancedFilterChain(config FilterConfig, enableCache bool, parallel bool) *AdvancedFilterChain {
	chain := &AdvancedFilterChain{
		contractFilters:   make([]Filter, 0),
		spreadFilters:     make([]SpreadFilter, 0),
		combinedFilters:   make([]CombinedFilter, 0),
		executionStats:    make(map[string]*FilterStats),
		parallelExecution: parallel,
		cacheEnabled:      enableCache,
	}
	
	if enableCache {
		chain.cache = NewFilterCache(5 * time.Minute)
	}
	
	// Initialize metrics
	chain.initMetrics()
	
	// Build filter chain from config
	chain.buildFromConfig(config)
	
	return chain
}

// initMetrics initializes Prometheus metrics
func (fc *AdvancedFilterChain) initMetrics() {
	// Create metrics without registering them to avoid conflicts in tests
	fc.metrics = &FilterMetrics{
		filterExecutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "filter_execution_duration_seconds",
				Help: "Filter execution duration in seconds",
			},
			[]string{"filter_name"},
		),
		filterItemsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "filter_items_processed_total",
				Help: "Total number of items processed by filter",
			},
			[]string{"filter_name"},
		),
		filterItemsRemoved: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "filter_items_removed_total",
				Help: "Total number of items removed by filter",
			},
			[]string{"filter_name"},
		),
		cacheHitRate: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "filter_cache_hit_rate",
				Help: "Filter cache hit rate",
			},
		),
	}
}

// buildFromConfig builds filter chain from configuration
func (fc *AdvancedFilterChain) buildFromConfig(config FilterConfig) {
	// Add contract filters
	if config.Delta != nil {
		fc.AddContractFilter(config.Delta)
	}
	if config.DTE != nil {
		fc.AddContractFilter(config.DTE)
	}
	if config.Liquidity != nil {
		fc.AddContractFilter(config.Liquidity)
	}
	if config.Theta != nil {
		fc.AddContractFilter(config.Theta)
	}
	if config.Vega != nil {
		fc.AddContractFilter(config.Vega)
	}
	if config.IV != nil {
		fc.AddContractFilter(config.IV)
	}
	if config.IVPercentile != nil {
		fc.AddContractFilter(config.IVPercentile)
	}
	
	// Add spread filters
	if config.SpreadWidth != nil {
		fc.AddSpreadFilter(config.SpreadWidth)
	}
	if config.ProbOfProfit != nil {
		fc.AddSpreadFilter(config.ProbOfProfit)
	}
}

// AddContractFilter adds a contract filter to the chain
func (fc *AdvancedFilterChain) AddContractFilter(filter Filter) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	fc.contractFilters = append(fc.contractFilters, filter)
	fc.executionStats[filter.Name()] = &FilterStats{}
}

// AddSpreadFilter adds a spread filter to the chain
func (fc *AdvancedFilterChain) AddSpreadFilter(filter SpreadFilter) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	fc.spreadFilters = append(fc.spreadFilters, filter)
	fc.executionStats[filter.Name()] = &FilterStats{}
}

// AddCombinedFilter adds a combined filter to the chain
func (fc *AdvancedFilterChain) AddCombinedFilter(filter CombinedFilter) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	fc.combinedFilters = append(fc.combinedFilters, filter)
	fc.executionStats[filter.Name()] = &FilterStats{}
}

// ApplyToContracts applies all contract filters with performance tracking
func (fc *AdvancedFilterChain) ApplyToContracts(contracts []models.OptionContract) []models.OptionContract {
	// Check cache if enabled
	if fc.cacheEnabled {
		if cached, found := fc.cache.GetContracts(contracts); found {
			fc.metrics.cacheHitRate.Add(1)
			return cached
		}
	}
	
	result := contracts
	
	if fc.parallelExecution && len(fc.contractFilters) > 1 {
		result = fc.applyContractsParallel(contracts)
	} else {
		result = fc.applyContractsSequential(contracts)
	}
	
	// Cache result if enabled
	if fc.cacheEnabled {
		fc.cache.SetContracts(contracts, result)
	}
	
	return result
}

// applyContractsSequential applies filters sequentially
func (fc *AdvancedFilterChain) applyContractsSequential(contracts []models.OptionContract) []models.OptionContract {
	result := contracts
	
	for _, filter := range fc.contractFilters {
		start := time.Now()
		initialCount := len(result)
		
		result = filter.Apply(result)
		
		duration := time.Since(start)
		fc.updateStats(filter.Name(), initialCount, len(result), duration)
	}
	
	return result
}

// applyContractsParallel applies independent filters in parallel
func (fc *AdvancedFilterChain) applyContractsParallel(contracts []models.OptionContract) []models.OptionContract {
	// For parallel execution, we need to identify independent filters
	// This is a simplified version - real implementation would analyze dependencies
	
	result := contracts
	var wg sync.WaitGroup
	resultChan := make(chan []models.OptionContract, len(fc.contractFilters))
	
	// Apply each filter independently and intersect results
	for _, filter := range fc.contractFilters {
		wg.Add(1)
		go func(f Filter) {
			defer wg.Done()
			start := time.Now()
			filtered := f.Apply(contracts)
			duration := time.Since(start)
			
			fc.updateStats(f.Name(), len(contracts), len(filtered), duration)
			resultChan <- filtered
		}(filter)
	}
	
	wg.Wait()
	close(resultChan)
	
	// Intersect all results
	first := true
	for filtered := range resultChan {
		if first {
			result = filtered
			first = false
		} else {
			result = intersectContracts(result, filtered)
		}
	}
	
	return result
}

// ApplyToSpreads applies all spread filters
func (fc *AdvancedFilterChain) ApplyToSpreads(spreads []models.VerticalSpread) []models.VerticalSpread {
	// Check cache if enabled
	if fc.cacheEnabled {
		if cached, found := fc.cache.GetSpreads(spreads); found {
			fc.metrics.cacheHitRate.Add(1)
			return cached
		}
	}
	
	result := make([]models.VerticalSpread, 0)
	
	for _, spread := range spreads {
		passed := true
		for _, filter := range fc.spreadFilters {
			start := time.Now()
			
			if !filter.ApplyToSpread(spread) {
				passed = false
				duration := time.Since(start)
				fc.updateStats(filter.Name(), 1, 0, duration)
				break
			}
			
			duration := time.Since(start)
			fc.updateStats(filter.Name(), 1, 1, duration)
		}
		
		if passed {
			result = append(result, spread)
		}
	}
	
	// Cache result if enabled
	if fc.cacheEnabled {
		fc.cache.SetSpreads(spreads, result)
	}
	
	return result
}

// ApplyCombined applies combined filters to both contracts and spreads
func (fc *AdvancedFilterChain) ApplyCombined(contracts []models.OptionContract, spreads []models.VerticalSpread) ([]models.OptionContract, []models.VerticalSpread) {
	resultContracts := contracts
	resultSpreads := spreads
	
	for _, filter := range fc.combinedFilters {
		start := time.Now()
		
		resultContracts, resultSpreads = filter.ApplyToCombined(resultContracts, resultSpreads)
		
		duration := time.Since(start)
		totalItems := len(contracts) + len(spreads)
		totalResult := len(resultContracts) + len(resultSpreads)
		fc.updateStats(filter.Name(), totalItems, totalResult, duration)
	}
	
	return resultContracts, resultSpreads
}

// updateStats updates execution statistics
func (fc *AdvancedFilterChain) updateStats(filterName string, itemsIn, itemsOut int, duration time.Duration) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	stats, exists := fc.executionStats[filterName]
	if !exists {
		stats = &FilterStats{}
		fc.executionStats[filterName] = stats
	}
	
	stats.ExecutionCount++
	stats.TotalDuration += duration
	stats.AverageDuration = stats.TotalDuration / time.Duration(stats.ExecutionCount)
	stats.ItemsProcessed += int64(itemsIn)
	stats.ItemsFiltered += int64(itemsIn - itemsOut)
	stats.LastExecution = time.Now()
	
	// Update Prometheus metrics
	fc.metrics.filterExecutionTime.WithLabelValues(filterName).Observe(duration.Seconds())
	fc.metrics.filterItemsProcessed.WithLabelValues(filterName).Add(float64(itemsIn))
	fc.metrics.filterItemsRemoved.WithLabelValues(filterName).Add(float64(itemsIn - itemsOut))
}

// GetStats returns execution statistics for all filters
func (fc *AdvancedFilterChain) GetStats() map[string]*FilterStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	statsCopy := make(map[string]*FilterStats)
	for name, stats := range fc.executionStats {
		statsCopy[name] = &FilterStats{
			ExecutionCount:   stats.ExecutionCount,
			TotalDuration:    stats.TotalDuration,
			AverageDuration:  stats.AverageDuration,
			ItemsProcessed:   stats.ItemsProcessed,
			ItemsFiltered:    stats.ItemsFiltered,
			LastExecution:    stats.LastExecution,
		}
	}
	
	return statsCopy
}

// intersectContracts returns contracts that appear in both slices
func intersectContracts(a, b []models.OptionContract) []models.OptionContract {
	contractMap := make(map[string]models.OptionContract)
	for _, contract := range a {
		contractMap[contract.Symbol] = contract
	}
	
	result := make([]models.OptionContract, 0)
	for _, contract := range b {
		if _, exists := contractMap[contract.Symbol]; exists {
			result = append(result, contract)
		}
	}
	
	return result
}

// Validate validates all filters in the chain
func (fc *AdvancedFilterChain) Validate() error {
	for _, filter := range fc.contractFilters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("contract filter %s validation failed: %w", filter.Name(), err)
		}
	}
	
	for _, filter := range fc.spreadFilters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("spread filter %s validation failed: %w", filter.Name(), err)
		}
	}
	
	for _, filter := range fc.combinedFilters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("combined filter %s validation failed: %w", filter.Name(), err)
		}
	}
	
	return nil
}