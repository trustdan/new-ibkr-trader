package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/models"
)

// Scanner is the main scanning service
type Scanner struct {
	mu              sync.RWMutex
	filterChain     *filters.FilterChain
	dataProvider    DataProvider
	cache           *ContractCache
	concurrentScans int
}

// DataProvider interface for getting market data
type DataProvider interface {
	GetOptionChain(ctx context.Context, symbol string) ([]models.OptionContract, error)
	GetQuote(ctx context.Context, symbol string) (float64, error)
}

// ContractCache for caching option data
type ContractCache struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	contracts []models.OptionContract
	timestamp time.Time
}

// NewScanner creates a new scanner instance
func NewScanner(provider DataProvider, config filters.FilterConfig) *Scanner {
	return &Scanner{
		filterChain:     filters.NewFilterChain(config),
		dataProvider:    provider,
		cache:           NewContractCache(5 * time.Minute),
		concurrentScans: 5, // Default concurrency
	}
}

// NewContractCache creates a new cache instance
func NewContractCache(ttl time.Duration) *ContractCache {
	return &ContractCache{
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

// ScanSymbol scans a single symbol for vertical spread opportunities
func (s *Scanner) ScanSymbol(ctx context.Context, symbol string) (*models.ScanResult, error) {
	start := time.Now()
	
	// Get option chain (with caching)
	contracts, err := s.getContractsWithCache(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get option chain: %w", err)
	}
	
	// Apply contract filters
	filtered := s.filterChain.ApplyToContracts(contracts)
	
	// Generate vertical spreads
	spreads := s.generateVerticalSpreads(filtered)
	
	// Apply spread filters
	validSpreads := make([]models.VerticalSpread, 0)
	for _, spread := range spreads {
		if s.filterChain.ApplyToSpread(spread) {
			validSpreads = append(validSpreads, spread)
		}
	}
	
	// Sort by score
	sort.Slice(validSpreads, func(i, j int) bool {
		return validSpreads[i].Score > validSpreads[j].Score
	})
	
	return &models.ScanResult{
		ScanID:     generateScanID(),
		Timestamp:  time.Now(),
		Symbol:     symbol,
		Spreads:    validSpreads,
		TotalFound: len(contracts),
		Filtered:   len(filtered),
		Duration:   time.Since(start),
	}, nil
}

// ScanMultiple scans multiple symbols concurrently
func (s *Scanner) ScanMultiple(ctx context.Context, symbols []string) ([]*models.ScanResult, error) {
	results := make([]*models.ScanResult, 0, len(symbols))
	resultCh := make(chan *models.ScanResult, len(symbols))
	errorCh := make(chan error, len(symbols))
	
	// Create worker pool
	sem := make(chan struct{}, s.concurrentScans)
	var wg sync.WaitGroup
	
	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			result, err := s.ScanSymbol(ctx, sym)
			if err != nil {
				errorCh <- fmt.Errorf("scan failed for %s: %w", sym, err)
				return
			}
			resultCh <- result
		}(symbol)
	}
	
	// Wait for all scans to complete
	go func() {
		wg.Wait()
		close(resultCh)
		close(errorCh)
	}()
	
	// Collect results
	for result := range resultCh {
		results = append(results, result)
	}
	
	// Check for errors
	var errs []error
	for err := range errorCh {
		errs = append(errs, err)
	}
	
	if len(errs) > 0 {
		return results, fmt.Errorf("scan completed with %d errors", len(errs))
	}
	
	return results, nil
}

// getContractsWithCache gets contracts with caching
func (s *Scanner) getContractsWithCache(ctx context.Context, symbol string) ([]models.OptionContract, error) {
	// Check cache first
	if cached, ok := s.cache.Get(symbol); ok {
		return cached, nil
	}
	
	// Fetch from provider
	contracts, err := s.dataProvider.GetOptionChain(ctx, symbol)
	if err != nil {
		return nil, err
	}
	
	// Cache the results
	s.cache.Set(symbol, contracts)
	
	return contracts, nil
}

// generateVerticalSpreads creates all possible vertical spreads
func (s *Scanner) generateVerticalSpreads(contracts []models.OptionContract) []models.VerticalSpread {
	spreads := make([]models.VerticalSpread, 0)
	
	// Group by expiry and type
	grouped := make(map[string][]models.OptionContract)
	for _, contract := range contracts {
		key := fmt.Sprintf("%s_%s", contract.Expiry.Format("2006-01-02"), contract.OptionType)
		grouped[key] = append(grouped[key], contract)
	}
	
	// Generate spreads within each group
	for _, group := range grouped {
		// Sort by strike
		sort.Slice(group, func(i, j int) bool {
			return group[i].Strike < group[j].Strike
		})
		
		// Create spreads (adjacent strikes for now)
		for i := 0; i < len(group)-1; i++ {
			for j := i + 1; j < len(group) && j <= i+3; j++ { // Max 3 strikes wide
				spread := s.createSpread(&group[i], &group[j])
				if spread != nil {
					spreads = append(spreads, *spread)
				}
			}
		}
	}
	
	return spreads
}

// createSpread creates a vertical spread from two contracts
func (s *Scanner) createSpread(lower, higher *models.OptionContract) *models.VerticalSpread {
	if lower.OptionType != higher.OptionType || lower.Expiry != higher.Expiry {
		return nil
	}
	
	spread := &models.VerticalSpread{}
	
	// Determine spread type based on option type
	if lower.OptionType == "CALL" {
		// Bull call spread (debit)
		spread.LongLeg = lower
		spread.ShortLeg = higher
		spread.SpreadType = "DEBIT"
		spread.NetDebit = lower.Ask - higher.Bid
		spread.MaxProfit = (higher.Strike - lower.Strike) - spread.NetDebit
		spread.MaxLoss = spread.NetDebit
		spread.Breakeven = lower.Strike + spread.NetDebit
	} else {
		// Bear put spread (debit)
		spread.LongLeg = higher
		spread.ShortLeg = lower
		spread.SpreadType = "DEBIT"
		spread.NetDebit = higher.Ask - lower.Bid
		spread.MaxProfit = (higher.Strike - lower.Strike) - spread.NetDebit
		spread.MaxLoss = spread.NetDebit
		spread.Breakeven = higher.Strike - spread.NetDebit
	}
	
	// Calculate combined Greeks
	spread.NetDelta = spread.LongLeg.Delta - spread.ShortLeg.Delta
	spread.NetTheta = spread.LongLeg.Theta - spread.ShortLeg.Theta
	spread.NetVega = spread.LongLeg.Vega - spread.ShortLeg.Vega
	
	// Simple scoring algorithm (can be enhanced)
	spread.Score = s.calculateSpreadScore(spread)
	
	return spread
}

// calculateSpreadScore calculates a score for ranking spreads
func (s *Scanner) calculateSpreadScore(spread *models.VerticalSpread) float64 {
	score := 0.0
	
	// Reward good risk/reward ratio
	if spread.MaxLoss > 0 {
		score += (spread.MaxProfit / spread.MaxLoss) * 10
	}
	
	// Reward positive theta (time decay in our favor)
	score += spread.NetTheta * 5
	
	// Penalize wide bid-ask spreads
	avgSpread := (spread.LongLeg.BidAskSpread + spread.ShortLeg.BidAskSpread) / 2
	score -= avgSpread * 2
	
	// Reward probability of profit (simplified calculation)
	if spread.ProbOfProfit > 0 {
		score += spread.ProbOfProfit * 20
	}
	
	return score
}

// Cache methods
func (c *ContractCache) Get(symbol string) ([]models.OptionContract, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, ok := c.cache[symbol]
	if !ok {
		return nil, false
	}
	
	// Check if cache is still valid
	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}
	
	return entry.contracts, true
}

func (c *ContractCache) Set(symbol string, contracts []models.OptionContract) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.cache[symbol] = cacheEntry{
		contracts: contracts,
		timestamp: time.Now(),
	}
}

// generateScanID creates a unique scan ID
func generateScanID() string {
	return fmt.Sprintf("scan_%d", time.Now().UnixNano())
}