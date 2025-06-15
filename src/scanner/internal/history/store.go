package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/rs/zerolog/log"
)

// HistoryStore manages historical scan data storage
type HistoryStore struct {
	dataDir       string
	retentionDays int
	mu            sync.RWMutex
	
	// In-memory cache for recent data
	recentResults map[string]*RecentData
	
	// Indexes for fast lookups
	symbolIndex   map[string][]string // symbol -> []resultIDs
	dateIndex     map[string][]string // date -> []resultIDs
	
	// Background workers
	flushTicker   *time.Ticker
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// RecentData holds recent scan results in memory
type RecentData struct {
	Results      []HistoricalResult
	LastUpdate   time.Time
	mu           sync.RWMutex
}

// HistoricalResult represents a scan result with additional metadata
type HistoricalResult struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Symbol       string                 `json:"symbol"`
	ScanResult   models.ScanResult      `json:"scan_result"`
	MarketState  MarketConditions       `json:"market_state"`
	FilterConfig map[string]interface{} `json:"filter_config"`
	Metrics      ResultMetrics          `json:"metrics"`
}

// MarketConditions captures market state at scan time
type MarketConditions struct {
	VIX          float64 `json:"vix"`
	SPYPrice     float64 `json:"spy_price"`
	MarketTrend  string  `json:"market_trend"` // up, down, sideways
	Volume       int64   `json:"volume"`
	Volatility   float64 `json:"volatility"`
}

// ResultMetrics contains calculated metrics
type ResultMetrics struct {
	TopSpreadScore    float64 `json:"top_spread_score"`
	AvgSpreadScore    float64 `json:"avg_spread_score"`
	BestCredit        float64 `json:"best_credit"`
	AvgCredit         float64 `json:"avg_credit"`
	SpreadCount       int     `json:"spread_count"`
	FilterReduction   float64 `json:"filter_reduction"`
}

// NewHistoryStore creates a new history store
func NewHistoryStore(dataDir string, retentionDays int) (*HistoryStore, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	hs := &HistoryStore{
		dataDir:       dataDir,
		retentionDays: retentionDays,
		recentResults: make(map[string]*RecentData),
		symbolIndex:   make(map[string][]string),
		dateIndex:     make(map[string][]string),
		stopChan:      make(chan struct{}),
	}
	
	// Load existing indexes
	if err := hs.loadIndexes(); err != nil {
		log.Warn().Err(err).Msg("Failed to load indexes, starting fresh")
	}
	
	// Start background workers
	hs.startWorkers()
	
	return hs, nil
}

// Store saves a scan result to history
func (hs *HistoryStore) Store(result models.ScanResult, marketState MarketConditions, filterConfig map[string]interface{}) error {
	// Create historical result
	hist := HistoricalResult{
		ID:           fmt.Sprintf("%s-%d", result.Symbol, time.Now().UnixNano()),
		Timestamp:    result.Timestamp,
		Symbol:       result.Symbol,
		ScanResult:   result,
		MarketState:  marketState,
		FilterConfig: filterConfig,
		Metrics:      hs.calculateMetrics(result),
	}
	
	// Add to memory cache
	hs.addToCache(hist)
	
	// Update indexes
	hs.updateIndexes(hist)
	
	// Write to disk asynchronously
	go hs.writeToDisk(hist)
	
	return nil
}

// addToCache adds result to memory cache
func (hs *HistoryStore) addToCache(hist HistoricalResult) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	// Get or create recent data for symbol
	recent, exists := hs.recentResults[hist.Symbol]
	if !exists {
		recent = &RecentData{
			Results:    make([]HistoricalResult, 0),
			LastUpdate: time.Now(),
		}
		hs.recentResults[hist.Symbol] = recent
	}
	
	recent.mu.Lock()
	recent.Results = append(recent.Results, hist)
	recent.LastUpdate = time.Now()
	
	// Keep only last 100 results per symbol
	if len(recent.Results) > 100 {
		recent.Results = recent.Results[len(recent.Results)-100:]
	}
	recent.mu.Unlock()
}

// updateIndexes updates lookup indexes
func (hs *HistoryStore) updateIndexes(hist HistoricalResult) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	// Update symbol index
	hs.symbolIndex[hist.Symbol] = append(hs.symbolIndex[hist.Symbol], hist.ID)
	
	// Update date index
	dateKey := hist.Timestamp.Format("2006-01-02")
	hs.dateIndex[dateKey] = append(hs.dateIndex[dateKey], hist.ID)
}

// writeToDisk writes result to disk storage
func (hs *HistoryStore) writeToDisk(hist HistoricalResult) {
	// Create directory structure: dataDir/YYYY/MM/DD/symbol/
	dateDir := filepath.Join(hs.dataDir,
		fmt.Sprintf("%d", hist.Timestamp.Year()),
		fmt.Sprintf("%02d", hist.Timestamp.Month()),
		fmt.Sprintf("%02d", hist.Timestamp.Day()),
		hist.Symbol,
	)
	
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create date directory")
		return
	}
	
	// Write file
	filename := filepath.Join(dateDir, fmt.Sprintf("%s.json", hist.ID))
	data, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal historical result")
		return
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to write historical result")
		return
	}
}

// calculateMetrics calculates result metrics
func (hs *HistoryStore) calculateMetrics(result models.ScanResult) ResultMetrics {
	metrics := ResultMetrics{
		SpreadCount: len(result.Spreads),
	}
	
	if result.TotalFound > 0 {
		metrics.FilterReduction = float64(result.TotalFound-result.Filtered) / float64(result.TotalFound)
	}
	
	if len(result.Spreads) > 0 {
		var totalScore, totalCredit float64
		bestCredit := 0.0
		bestScore := 0.0
		
		for _, spread := range result.Spreads {
			totalScore += spread.Score
			totalCredit += spread.Credit
			
			if spread.Credit > bestCredit {
				bestCredit = spread.Credit
			}
			if spread.Score > bestScore {
				bestScore = spread.Score
			}
		}
		
		metrics.TopSpreadScore = bestScore
		metrics.AvgSpreadScore = totalScore / float64(len(result.Spreads))
		metrics.BestCredit = bestCredit
		metrics.AvgCredit = totalCredit / float64(len(result.Spreads))
	}
	
	return metrics
}

// Query methods

// GetRecentBySymbol returns recent results for a symbol
func (hs *HistoryStore) GetRecentBySymbol(symbol string, limit int) []HistoricalResult {
	hs.mu.RLock()
	recent, exists := hs.recentResults[symbol]
	hs.mu.RUnlock()
	
	if !exists {
		return []HistoricalResult{}
	}
	
	recent.mu.RLock()
	defer recent.mu.RUnlock()
	
	// Return most recent results
	start := len(recent.Results) - limit
	if start < 0 {
		start = 0
	}
	
	return recent.Results[start:]
}

// GetByDateRange returns results within a date range
func (hs *HistoryStore) GetByDateRange(startDate, endDate time.Time, symbols []string) []HistoricalResult {
	results := make([]HistoricalResult, 0)
	
	// Iterate through dates
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		
		hs.mu.RLock()
		resultIDs, exists := hs.dateIndex[dateKey]
		hs.mu.RUnlock()
		
		if !exists {
			continue
		}
		
		// Load results from disk
		for _, id := range resultIDs {
			hist, err := hs.loadFromDisk(id, d)
			if err != nil {
				continue
			}
			
			// Filter by symbols if specified
			if len(symbols) > 0 {
				found := false
				for _, sym := range symbols {
					if hist.Symbol == sym {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			
			results = append(results, hist)
		}
	}
	
	return results
}

// loadFromDisk loads a historical result from disk
func (hs *HistoryStore) loadFromDisk(id string, date time.Time) (HistoricalResult, error) {
	// This is simplified - in production you'd parse the ID to get the symbol
	// For now, we'll check each symbol directory
	
	var result HistoricalResult
	dateDir := filepath.Join(hs.dataDir,
		fmt.Sprintf("%d", date.Year()),
		fmt.Sprintf("%02d", date.Month()),
		fmt.Sprintf("%02d", date.Day()),
	)
	
	// Walk symbol directories
	err := filepath.Walk(dateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		if filepath.Base(path) == fmt.Sprintf("%s.json", id) {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			
			return json.Unmarshal(data, &result)
		}
		
		return nil
	})
	
	return result, err
}

// Analysis methods

// GetStatsBySymbol returns statistics for a symbol
func (hs *HistoryStore) GetStatsBySymbol(symbol string, days int) SymbolStats {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	
	results := hs.GetByDateRange(startDate, endDate, []string{symbol})
	
	return hs.calculateStats(results)
}

// SymbolStats contains statistical analysis
type SymbolStats struct {
	Symbol          string         `json:"symbol"`
	TotalScans      int            `json:"total_scans"`
	AvgSpreadsFound float64        `json:"avg_spreads_found"`
	BestScore       float64        `json:"best_score"`
	AvgScore        float64        `json:"avg_score"`
	BestCredit      float64        `json:"best_credit"`
	AvgCredit       float64        `json:"avg_credit"`
	TrendData       []TrendPoint   `json:"trend_data"`
	Patterns        []Pattern      `json:"patterns"`
}

// TrendPoint represents a point in trend analysis
type TrendPoint struct {
	Date       string  `json:"date"`
	AvgScore   float64 `json:"avg_score"`
	SpreadCount int    `json:"spread_count"`
}

// Pattern represents a detected pattern
type Pattern struct {
	Type        string    `json:"type"` // daily, weekly, etc
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	LastSeen    time.Time `json:"last_seen"`
}

// calculateStats calculates statistics from results
func (hs *HistoryStore) calculateStats(results []HistoricalResult) SymbolStats {
	if len(results) == 0 {
		return SymbolStats{}
	}
	
	stats := SymbolStats{
		Symbol:     results[0].Symbol,
		TotalScans: len(results),
	}
	
	var totalSpreads int
	var totalScore, totalCredit float64
	
	// Daily aggregation for trends
	dailyData := make(map[string]*TrendPoint)
	
	for _, hist := range results {
		metrics := hist.Metrics
		
		// Update totals
		totalSpreads += metrics.SpreadCount
		if metrics.SpreadCount > 0 {
			totalScore += metrics.AvgSpreadScore
			totalCredit += metrics.AvgCredit
			
			if metrics.TopSpreadScore > stats.BestScore {
				stats.BestScore = metrics.TopSpreadScore
			}
			if metrics.BestCredit > stats.BestCredit {
				stats.BestCredit = metrics.BestCredit
			}
		}
		
		// Update daily data
		dateKey := hist.Timestamp.Format("2006-01-02")
		if daily, exists := dailyData[dateKey]; exists {
			daily.AvgScore = (daily.AvgScore + metrics.AvgSpreadScore) / 2
			daily.SpreadCount += metrics.SpreadCount
		} else {
			dailyData[dateKey] = &TrendPoint{
				Date:        dateKey,
				AvgScore:    metrics.AvgSpreadScore,
				SpreadCount: metrics.SpreadCount,
			}
		}
	}
	
	// Calculate averages
	if stats.TotalScans > 0 {
		stats.AvgSpreadsFound = float64(totalSpreads) / float64(stats.TotalScans)
		stats.AvgScore = totalScore / float64(stats.TotalScans)
		stats.AvgCredit = totalCredit / float64(stats.TotalScans)
	}
	
	// Convert daily data to trend
	for _, point := range dailyData {
		stats.TrendData = append(stats.TrendData, *point)
	}
	
	// Detect patterns (simplified)
	stats.Patterns = hs.detectPatterns(results)
	
	return stats
}

// detectPatterns detects trading patterns
func (hs *HistoryStore) detectPatterns(results []HistoricalResult) []Pattern {
	patterns := make([]Pattern, 0)
	
	// Example: High opportunity days
	highOpDays := 0
	for _, hist := range results {
		if hist.Metrics.SpreadCount > 10 {
			highOpDays++
		}
	}
	
	if highOpDays > len(results)/3 {
		patterns = append(patterns, Pattern{
			Type:        "high_opportunity",
			Description: "Frequently shows high number of spreads",
			Confidence:  float64(highOpDays) / float64(len(results)),
			LastSeen:    time.Now(),
		})
	}
	
	return patterns
}

// Background workers

// startWorkers starts background workers
func (hs *HistoryStore) startWorkers() {
	// Flush cache periodically
	hs.flushTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-hs.flushTicker.C:
				hs.flushCache()
			case <-hs.stopChan:
				return
			}
		}
	}()
	
	// Cleanup old data
	hs.cleanupTicker = time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-hs.cleanupTicker.C:
				hs.cleanupOldData()
			case <-hs.stopChan:
				return
			}
		}
	}()
}

// flushCache saves cached data to disk
func (hs *HistoryStore) flushCache() {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	for _, recent := range hs.recentResults {
		recent.mu.RLock()
		for _, hist := range recent.Results {
			go hs.writeToDisk(hist)
		}
		recent.mu.RUnlock()
	}
	
	// Save indexes
	hs.saveIndexes()
}

// cleanupOldData removes data older than retention period
func (hs *HistoryStore) cleanupOldData() {
	cutoffDate := time.Now().AddDate(0, 0, -hs.retentionDays)
	
	// Walk through old directories and remove them
	yearDir := filepath.Join(hs.dataDir, fmt.Sprintf("%d", cutoffDate.Year()))
	
	filepath.Walk(yearDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Check file date
		if info.ModTime().Before(cutoffDate) {
			os.Remove(path)
		}
		
		return nil
	})
}

// Index persistence

// saveIndexes saves indexes to disk
func (hs *HistoryStore) saveIndexes() {
	indexFile := filepath.Join(hs.dataDir, "indexes.json")
	
	indexes := map[string]interface{}{
		"symbol_index": hs.symbolIndex,
		"date_index":   hs.dateIndex,
	}
	
	data, err := json.MarshalIndent(indexes, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal indexes")
		return
	}
	
	if err := os.WriteFile(indexFile, data, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to save indexes")
	}
}

// loadIndexes loads indexes from disk
func (hs *HistoryStore) loadIndexes() error {
	indexFile := filepath.Join(hs.dataDir, "indexes.json")
	
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return err
	}
	
	var indexes map[string]interface{}
	if err := json.Unmarshal(data, &indexes); err != nil {
		return err
	}
	
	// Type assertions
	if symbolIndex, ok := indexes["symbol_index"].(map[string]interface{}); ok {
		for k, v := range symbolIndex {
			if ids, ok := v.([]interface{}); ok {
				stringIDs := make([]string, len(ids))
				for i, id := range ids {
					stringIDs[i] = id.(string)
				}
				hs.symbolIndex[k] = stringIDs
			}
		}
	}
	
	return nil
}

// Stop stops the history store
func (hs *HistoryStore) Stop() {
	close(hs.stopChan)
	
	if hs.flushTicker != nil {
		hs.flushTicker.Stop()
	}
	if hs.cleanupTicker != nil {
		hs.cleanupTicker.Stop()
	}
	
	// Final flush
	hs.flushCache()
}