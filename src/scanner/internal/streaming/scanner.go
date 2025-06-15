package streaming

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/ibkr-trader/scanner/internal/service"
	"github.com/rs/zerolog/log"
)

// StreamingScanner manages continuous scanning with real-time updates
type StreamingScanner struct {
	scanner         service.Scanner
	wsServer        *WebSocketServer
	filterChain     *filters.AdvancedFilterChain
	
	// Scan configuration
	scanInterval    time.Duration
	symbols         []string
	
	// Deduplication
	resultCache     *ResultCache
	
	// State management
	isRunning       bool
	mu              sync.RWMutex
	
	// Channels
	scanRequests    chan ScanRequest
	stopChan        chan struct{}
	
	// Pacing control
	lastScanTime    map[string]time.Time
	minScanInterval time.Duration
	
	// Metrics
	scanCount       int64
	resultCount     int64
	errorCount      int64
}

// ScanRequest represents a request to scan specific symbols
type ScanRequest struct {
	Symbols       []string
	FilterConfig  filters.FilterConfig
	ResponseChan  chan<- ScanResponse
}

// ScanResponse contains the response to a scan request
type ScanResponse struct {
	Results []models.ScanResult
	Error   error
}

// ResultCache manages deduplication of results
type ResultCache struct {
	cache      map[string]*CachedResult
	mu         sync.RWMutex
	ttl        time.Duration
	maxSize    int
}

// CachedResult stores a cached scan result
type CachedResult struct {
	Result     models.ScanResult
	Hash       string
	Timestamp  time.Time
}

// NewStreamingScanner creates a new streaming scanner
func NewStreamingScanner(scanner service.Scanner, wsServer *WebSocketServer, filterChain *filters.AdvancedFilterChain) *StreamingScanner {
	return &StreamingScanner{
		scanner:         scanner,
		wsServer:        wsServer,
		filterChain:     filterChain,
		scanInterval:    5 * time.Second,
		resultCache:     NewResultCache(5*time.Minute, 10000),
		scanRequests:    make(chan ScanRequest, 100),
		stopChan:        make(chan struct{}),
		lastScanTime:    make(map[string]time.Time),
		minScanInterval: 1 * time.Second,
	}
}

// Start begins continuous scanning
func (ss *StreamingScanner) Start(ctx context.Context, symbols []string) error {
	ss.mu.Lock()
	if ss.isRunning {
		ss.mu.Unlock()
		return fmt.Errorf("streaming scanner already running")
	}
	
	ss.symbols = symbols
	ss.isRunning = true
	ss.mu.Unlock()
	
	// Start workers
	go ss.continuousScan(ctx)
	go ss.handleRequests(ctx)
	
	// Broadcast status
	ss.wsServer.BroadcastStatus(StatusUpdate{
		Status:  "scanning_started",
		Message: fmt.Sprintf("Started scanning %d symbols", len(symbols)),
	})
	
	log.Info().Int("symbols", len(symbols)).Msg("Streaming scanner started")
	return nil
}

// Stop halts continuous scanning
func (ss *StreamingScanner) Stop() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	if !ss.isRunning {
		return
	}
	
	ss.isRunning = false
	close(ss.stopChan)
	
	// Broadcast status
	ss.wsServer.BroadcastStatus(StatusUpdate{
		Status:  "scanning_stopped",
		Message: "Scanning halted",
	})
	
	log.Info().Msg("Streaming scanner stopped")
}

// continuousScan performs continuous scanning
func (ss *StreamingScanner) continuousScan(ctx context.Context) {
	ticker := time.NewTicker(ss.scanInterval)
	defer ticker.Stop()
	
	// Initial scan
	ss.performScan(ctx, ss.symbols)
	
	for {
		select {
		case <-ctx.Done():
			return
			
		case <-ss.stopChan:
			return
			
		case <-ticker.C:
			ss.performScan(ctx, ss.symbols)
		}
	}
}

// handleRequests processes ad-hoc scan requests
func (ss *StreamingScanner) handleRequests(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
			
		case <-ss.stopChan:
			return
			
		case req := <-ss.scanRequests:
			results := make([]models.ScanResult, 0)
			
			// Apply custom filter config if provided
			var chain *filters.AdvancedFilterChain
			if req.FilterConfig.Delta != nil || req.FilterConfig.DTE != nil {
				chain = filters.NewAdvancedFilterChain(req.FilterConfig, true, true)
			} else {
				chain = ss.filterChain
			}
			
			// Scan requested symbols
			for _, symbol := range req.Symbols {
				if ss.shouldScan(symbol) {
					result, err := ss.scanSymbol(ctx, symbol, chain)
					if err != nil {
						log.Error().Err(err).Str("symbol", symbol).Msg("Scan failed")
						continue
					}
					results = append(results, result)
				}
			}
			
			// Send response
			if req.ResponseChan != nil {
				req.ResponseChan <- ScanResponse{
					Results: results,
					Error:   nil,
				}
			}
		}
	}
}

// performScan scans all configured symbols
func (ss *StreamingScanner) performScan(ctx context.Context, symbols []string) {
	startTime := time.Now()
	successCount := 0
	
	// Broadcast scan start
	ss.wsServer.BroadcastStatus(StatusUpdate{
		Status:   "scanning",
		Message:  "Scan in progress",
		Progress: 0,
		Total:    len(symbols),
	})
	
	// Scan each symbol
	for i, symbol := range symbols {
		if !ss.shouldScan(symbol) {
			continue
		}
		
		result, err := ss.scanSymbol(ctx, symbol, ss.filterChain)
		if err != nil {
			ss.errorCount++
			log.Error().Err(err).Str("symbol", symbol).Msg("Scan failed")
			ss.wsServer.BroadcastError(err)
			continue
		}
		
		ss.scanCount++
		successCount++
		
		// Check for changes and broadcast
		if ss.hasChanges(result) {
			ss.broadcastResult(result)
			ss.resultCount++
		}
		
		// Update progress
		if i%10 == 0 {
			ss.wsServer.BroadcastStatus(StatusUpdate{
				Status:   "scanning",
				Message:  fmt.Sprintf("Scanning %s", symbol),
				Progress: i + 1,
				Total:    len(symbols),
			})
		}
	}
	
	// Broadcast completion
	duration := time.Since(startTime)
	ss.wsServer.BroadcastStatus(StatusUpdate{
		Status:  "scan_complete",
		Message: fmt.Sprintf("Scanned %d symbols in %v", successCount, duration),
		Total:   len(symbols),
	})
	
	log.Info().
		Int("symbols", successCount).
		Dur("duration", duration).
		Int64("total_scans", ss.scanCount).
		Msg("Scan cycle complete")
}

// scanSymbol scans a single symbol
func (ss *StreamingScanner) scanSymbol(ctx context.Context, symbol string, chain *filters.AdvancedFilterChain) (models.ScanResult, error) {
	// Record scan time
	ss.mu.Lock()
	ss.lastScanTime[symbol] = time.Now()
	ss.mu.Unlock()
	
	// Get option chain
	contracts, err := ss.scanner.GetOptionChain(ctx, symbol)
	if err != nil {
		return models.ScanResult{}, fmt.Errorf("failed to get option chain: %w", err)
	}
	
	// Apply filters
	filtered := chain.ApplyToContracts(contracts)
	
	// Find spreads
	spreads := ss.scanner.FindVerticalSpreads(filtered)
	
	// Apply spread filters
	filteredSpreads := chain.ApplyToSpreads(spreads)
	
	// Create result
	result := models.ScanResult{
		ScanID:     fmt.Sprintf("%s-%d", symbol, time.Now().Unix()),
		Timestamp:  time.Now(),
		Symbol:     symbol,
		Spreads:    filteredSpreads,
		TotalFound: len(spreads),
		Filtered:   len(filteredSpreads),
	}
	
	return result, nil
}

// shouldScan checks if a symbol should be scanned
func (ss *StreamingScanner) shouldScan(symbol string) bool {
	ss.mu.RLock()
	lastScan, exists := ss.lastScanTime[symbol]
	ss.mu.RUnlock()
	
	if !exists {
		return true
	}
	
	return time.Since(lastScan) >= ss.minScanInterval
}

// hasChanges checks if results have changed
func (ss *StreamingScanner) hasChanges(result models.ScanResult) bool {
	return !ss.resultCache.IsDuplicate(result)
}

// broadcastResult broadcasts a scan result
func (ss *StreamingScanner) broadcastResult(result models.ScanResult) {
	// Determine update type
	updateType := "new"
	if ss.resultCache.Exists(result.Symbol) {
		updateType = "update"
	}
	
	// Cache result
	ss.resultCache.Store(result)
	
	// Create update
	update := ScanUpdate{
		ScanID:     result.ScanID,
		Symbol:     result.Symbol,
		Spreads:    result.Spreads,
		UpdateType: updateType,
		Metadata: map[string]interface{}{
			"total_found": result.TotalFound,
			"filtered":    result.Filtered,
			"timestamp":   result.Timestamp,
		},
	}
	
	// Broadcast to subscribers
	ss.wsServer.BroadcastScanUpdate(update)
}

// SubmitScanRequest submits an ad-hoc scan request
func (ss *StreamingScanner) SubmitScanRequest(symbols []string, config filters.FilterConfig) <-chan ScanResponse {
	responseChan := make(chan ScanResponse, 1)
	
	request := ScanRequest{
		Symbols:      symbols,
		FilterConfig: config,
		ResponseChan: responseChan,
	}
	
	select {
	case ss.scanRequests <- request:
	default:
		// Queue full, return error
		responseChan <- ScanResponse{
			Error: fmt.Errorf("scan request queue full"),
		}
	}
	
	return responseChan
}

// GetStats returns scanner statistics
func (ss *StreamingScanner) GetStats() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	
	return map[string]interface{}{
		"is_running":    ss.isRunning,
		"scan_count":    ss.scanCount,
		"result_count":  ss.resultCount,
		"error_count":   ss.errorCount,
		"symbol_count":  len(ss.symbols),
		"cache_size":    ss.resultCache.Size(),
	}
}

// ResultCache implementation

// NewResultCache creates a new result cache
func NewResultCache(ttl time.Duration, maxSize int) *ResultCache {
	cache := &ResultCache{
		cache:   make(map[string]*CachedResult),
		ttl:     ttl,
		maxSize: maxSize,
	}
	
	// Start cleanup routine
	go cache.cleanup()
	
	return cache
}

// IsDuplicate checks if a result is a duplicate
func (rc *ResultCache) IsDuplicate(result models.ScanResult) bool {
	hash := rc.computeHash(result)
	
	rc.mu.RLock()
	cached, exists := rc.cache[result.Symbol]
	rc.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// Check if expired
	if time.Since(cached.Timestamp) > rc.ttl {
		return false
	}
	
	return cached.Hash == hash
}

// Store stores a result in the cache
func (rc *ResultCache) Store(result models.ScanResult) {
	hash := rc.computeHash(result)
	
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	// Check size limit
	if len(rc.cache) >= rc.maxSize {
		// Remove oldest entry
		var oldestKey string
		var oldestTime time.Time
		
		for key, cached := range rc.cache {
			if oldestKey == "" || cached.Timestamp.Before(oldestTime) {
				oldestKey = key
				oldestTime = cached.Timestamp
			}
		}
		
		if oldestKey != "" {
			delete(rc.cache, oldestKey)
		}
	}
	
	rc.cache[result.Symbol] = &CachedResult{
		Result:    result,
		Hash:      hash,
		Timestamp: time.Now(),
	}
}

// Exists checks if a symbol exists in cache
func (rc *ResultCache) Exists(symbol string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	_, exists := rc.cache[symbol]
	return exists
}

// Size returns the cache size
func (rc *ResultCache) Size() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return len(rc.cache)
}

// computeHash computes a hash of the result for deduplication
func (rc *ResultCache) computeHash(result models.ScanResult) string {
	// Simple hash based on spread count and top spread details
	if len(result.Spreads) == 0 {
		return fmt.Sprintf("%s-0", result.Symbol)
	}
	
	topSpread := result.Spreads[0]
	return fmt.Sprintf("%s-%d-%.2f-%.2f-%.2f",
		result.Symbol,
		len(result.Spreads),
		topSpread.ShortLeg.Strike,
		topSpread.LongLeg.Strike,
		topSpread.Credit,
	)
}

// cleanup removes expired entries
func (rc *ResultCache) cleanup() {
	ticker := time.NewTicker(rc.ttl / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		rc.mu.Lock()
		
		now := time.Now()
		for symbol, cached := range rc.cache {
			if now.Sub(cached.Timestamp) > rc.ttl {
				delete(rc.cache, symbol)
			}
		}
		
		rc.mu.Unlock()
	}
}