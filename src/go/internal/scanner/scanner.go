package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/ibkr-automation/scanner/internal/filters"
	"github.com/ibkr-automation/scanner/pkg/models"
)

// Scanner handles options scanning operations
type Scanner struct {
	cache         *cache.Cache
	logger        *zap.SugaredLogger
	pythonService string
	httpClient    *http.Client
	
	// Performance metrics
	scanCount     uint64
	mu            sync.RWMutex
}

// New creates a new Scanner instance
func New(c *cache.Cache, logger *zap.SugaredLogger) *Scanner {
	pythonService := "http://python-ibkr:8080"
	if url := os.Getenv("PYTHON_SERVICE_URL"); url != "" {
		pythonService = url
	}

	return &Scanner{
		cache:         c,
		logger:        logger,
		pythonService: pythonService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScanOptions performs a scan with the given parameters
func (s *Scanner) ScanOptions(ctx context.Context, req *models.ScanRequest) (*models.ScanResponse, error) {
	s.logger.Infof("📊 Starting scan for %s with %d filters", req.Symbol, len(req.Filters))
	
	// Check cache first
	cacheKey := s.getCacheKey(req)
	if cached, found := s.cache.Get(cacheKey); found {
		s.logger.Debug("Cache hit for scan request")
		return cached.(*models.ScanResponse), nil
	}

	// Fetch options data from Python service
	optionsData, err := s.fetchOptionsData(ctx, req.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch options data: %w", err)
	}

	// Apply filters concurrently
	filtered := s.applyFilters(optionsData, req.Filters)

	// Build response
	response := &models.ScanResponse{
		Symbol:      req.Symbol,
		ScanTime:    time.Now(),
		ResultCount: len(filtered),
		Options:     filtered,
	}

	// Cache the results
	s.cache.Set(cacheKey, response, cache.DefaultExpiration)

	// Update metrics
	s.mu.Lock()
	s.scanCount++
	s.mu.Unlock()

	s.logger.Infof("✅ Scan complete: %d options found", len(filtered))
	return response, nil
}

// fetchOptionsData retrieves options chain from Python service
func (s *Scanner) fetchOptionsData(ctx context.Context, symbol string) ([]models.Option, error) {
	url := fmt.Sprintf("%s/api/v1/options/%s", s.pythonService, symbol)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python service returned status %d", resp.StatusCode)
	}

	var options []models.Option
	if err := json.NewDecoder(resp.Body).Decode(&options); err != nil {
		return nil, err
	}

	return options, nil
}

// applyFilters applies all filters to the options data
func (s *Scanner) applyFilters(options []models.Option, filterConfigs []models.FilterConfig) []models.Option {
	// Use goroutines for parallel filtering
	type result struct {
		idx    int
		passed bool
	}

	results := make(chan result, len(options))
	var wg sync.WaitGroup

	// Create filter chain
	filterChain := filters.BuildFilterChain(filterConfigs)

	// Process each option in parallel
	for i, option := range options {
		wg.Add(1)
		go func(idx int, opt models.Option) {
			defer wg.Done()
			passed := filterChain.Apply(&opt)
			results <- result{idx: idx, passed: passed}
		}(i, option)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	passedMap := make(map[int]bool)
	for r := range results {
		passedMap[r.idx] = r.passed
	}

	// Build filtered slice
	filtered := make([]models.Option, 0)
	for i, option := range options {
		if passedMap[i] {
			filtered = append(filtered, option)
		}
	}

	return filtered
}

// getCacheKey generates a cache key for the scan request
func (s *Scanner) getCacheKey(req *models.ScanRequest) string {
	// Simple key generation - could be improved with hash
	return fmt.Sprintf("%s:%v", req.Symbol, req.Filters)
}

// GetStats returns scanner statistics
func (s *Scanner) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_scans":  s.scanCount,
		"cache_stats":  s.cache.ItemCount(),
		"service":      "go-scanner",
	}
}