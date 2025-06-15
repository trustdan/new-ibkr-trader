package analytics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/history"
	"github.com/ibkr-trader/scanner/internal/metrics"
)

// Aggregator aggregates metrics from multiple sources
type Aggregator struct {
	metricsCollector *metrics.MetricsCollector
	historyStore     *history.HistoryStore
	analyzer         *Analyzer
	
	// Aggregation intervals
	intervals        []AggregationInterval
}

// AggregationInterval defines time-based aggregation
type AggregationInterval struct {
	Name     string
	Duration time.Duration
	Points   int
}

// AggregatedMetrics contains aggregated metrics data
type AggregatedMetrics struct {
	Period          string                    `json:"period"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         time.Time                 `json:"end_time"`
	
	// Scanner metrics
	TotalScans      int64                     `json:"total_scans"`
	AvgScanRate     float64                   `json:"avg_scan_rate"`
	PeakScanRate    float64                   `json:"peak_scan_rate"`
	
	// Result metrics
	TotalResults    int64                     `json:"total_results"`
	UniqueSymbols   int                       `json:"unique_symbols"`
	AvgSpreadsFound float64                   `json:"avg_spreads_found"`
	
	// Performance metrics
	AvgScanDuration float64                   `json:"avg_scan_duration_ms"`
	CacheHitRate    float64                   `json:"cache_hit_rate"`
	ErrorRate       float64                   `json:"error_rate"`
	
	// Filter metrics
	FilterStats     map[string]FilterStats    `json:"filter_stats"`
	
	// WebSocket metrics
	AvgConnections  float64                   `json:"avg_connections"`
	TotalMessages   int64                     `json:"total_messages"`
	
	// Top performers
	TopSymbols      []SymbolPerformance       `json:"top_symbols"`
	TopSpreads      []SpreadSummary           `json:"top_spreads"`
	
	// Patterns and trends
	DetectedPatterns []string                 `json:"detected_patterns"`
	MarketRegimes   map[string]float64        `json:"market_regimes"`
}

// FilterStats contains aggregated filter statistics
type FilterStats struct {
	Executions      int64   `json:"executions"`
	AvgDuration     float64 `json:"avg_duration_ms"`
	AvgReduction    float64 `json:"avg_reduction"`
	TotalFiltered   int64   `json:"total_filtered"`
}

// SymbolPerformance represents symbol-level performance
type SymbolPerformance struct {
	Symbol          string  `json:"symbol"`
	ScanCount       int     `json:"scan_count"`
	AvgSpreads      float64 `json:"avg_spreads"`
	BestScore       float64 `json:"best_score"`
	TotalVolume     int64   `json:"total_volume"`
}

// SpreadSummary summarizes a high-performing spread
type SpreadSummary struct {
	Symbol          string    `json:"symbol"`
	Strike1         float64   `json:"strike1"`
	Strike2         float64   `json:"strike2"`
	Credit          float64   `json:"credit"`
	Score           float64   `json:"score"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	Occurrences     int       `json:"occurrences"`
}

// NewAggregator creates a new metrics aggregator
func NewAggregator(collector *metrics.MetricsCollector, store *history.HistoryStore, analyzer *Analyzer) *Aggregator {
	return &Aggregator{
		metricsCollector: collector,
		historyStore:     store,
		analyzer:         analyzer,
		intervals: []AggregationInterval{
			{Name: "1hour", Duration: 1 * time.Hour, Points: 60},
			{Name: "1day", Duration: 24 * time.Hour, Points: 24},
			{Name: "1week", Duration: 7 * 24 * time.Hour, Points: 7},
			{Name: "1month", Duration: 30 * 24 * time.Hour, Points: 30},
		},
	}
}

// Aggregate performs metrics aggregation for a time period
func (a *Aggregator) Aggregate(startTime, endTime time.Time) (*AggregatedMetrics, error) {
	agg := &AggregatedMetrics{
		Period:       fmt.Sprintf("%s to %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02")),
		StartTime:    startTime,
		EndTime:      endTime,
		FilterStats:  make(map[string]FilterStats),
		TopSymbols:   make([]SymbolPerformance, 0),
		TopSpreads:   make([]SpreadSummary, 0),
		MarketRegimes: make(map[string]float64),
	}
	
	// Aggregate scanner metrics
	a.aggregateScannerMetrics(agg, startTime, endTime)
	
	// Aggregate filter metrics
	a.aggregateFilterMetrics(agg)
	
	// Aggregate historical data
	a.aggregateHistoricalData(agg, startTime, endTime)
	
	// Analyze patterns and trends
	a.analyzePatterns(agg, startTime, endTime)
	
	return agg, nil
}

// aggregateScannerMetrics aggregates scanner-level metrics
func (a *Aggregator) aggregateScannerMetrics(agg *AggregatedMetrics, start, end time.Time) {
	// This is simplified - in production you'd query Prometheus
	// For now, we'll use mock data
	
	agg.TotalScans = 1000
	agg.AvgScanRate = 2.5
	agg.PeakScanRate = 5.0
	agg.TotalResults = 5000
	agg.AvgSpreadsFound = 15.5
	agg.AvgScanDuration = 150.0
	agg.CacheHitRate = 0.85
	agg.ErrorRate = 0.02
	agg.AvgConnections = 10.5
	agg.TotalMessages = 50000
}

// aggregateFilterMetrics aggregates filter performance
func (a *Aggregator) aggregateFilterMetrics(agg *AggregatedMetrics) {
	// Mock filter statistics
	filters := []string{"DeltaFilter", "DTEFilter", "LiquidityFilter", "IVFilter"}
	
	for _, filterName := range filters {
		agg.FilterStats[filterName] = FilterStats{
			Executions:    1000,
			AvgDuration:   5.0,
			AvgReduction:  0.75,
			TotalFiltered: 750,
		}
	}
}

// aggregateHistoricalData aggregates historical scan data
func (a *Aggregator) aggregateHistoricalData(agg *AggregatedMetrics, start, end time.Time) {
	// Get historical results
	results := a.historyStore.GetByDateRange(start, end, nil)
	
	// Symbol performance tracking
	symbolStats := make(map[string]*SymbolPerformance)
	spreadMap := make(map[string]*SpreadSummary)
	regimeCount := make(map[string]int)
	
	for _, hist := range results {
		// Update symbol stats
		sym, exists := symbolStats[hist.Symbol]
		if !exists {
			sym = &SymbolPerformance{Symbol: hist.Symbol}
			symbolStats[hist.Symbol] = sym
		}
		
		sym.ScanCount++
		sym.AvgSpreads = (sym.AvgSpreads*float64(sym.ScanCount-1) + float64(hist.Metrics.SpreadCount)) / float64(sym.ScanCount)
		if hist.Metrics.TopSpreadScore > sym.BestScore {
			sym.BestScore = hist.Metrics.TopSpreadScore
		}
		
		// Track spreads
		for _, spread := range hist.ScanResult.Spreads {
			key := fmt.Sprintf("%s-%.0f-%.0f", spread.Symbol, spread.ShortLeg.Strike, spread.LongLeg.Strike)
			
			summary, exists := spreadMap[key]
			if !exists {
				summary = &SpreadSummary{
					Symbol:    spread.Symbol,
					Strike1:   spread.ShortLeg.Strike,
					Strike2:   spread.LongLeg.Strike,
					FirstSeen: hist.Timestamp,
				}
				spreadMap[key] = summary
			}
			
			summary.LastSeen = hist.Timestamp
			summary.Occurrences++
			if spread.Credit > summary.Credit {
				summary.Credit = spread.Credit
			}
			if spread.Score > summary.Score {
				summary.Score = spread.Score
			}
		}
		
		// Count market regimes
		if hist.MarketState.MarketTrend != "" {
			regimeCount[hist.MarketState.MarketTrend]++
		}
	}
	
	// Convert to sorted lists
	agg.UniqueSymbols = len(symbolStats)
	
	// Top symbols by score
	for _, sym := range symbolStats {
		agg.TopSymbols = append(agg.TopSymbols, *sym)
	}
	// Sort by best score (simplified - would use sort.Slice)
	
	// Top spreads by score
	for _, spread := range spreadMap {
		if spread.Score > 0.8 { // High score threshold
			agg.TopSpreads = append(agg.TopSpreads, *spread)
		}
	}
	
	// Market regime percentages
	total := len(results)
	for regime, count := range regimeCount {
		agg.MarketRegimes[regime] = float64(count) / float64(total)
	}
}

// analyzePatterns analyzes patterns in the data
func (a *Aggregator) analyzePatterns(agg *AggregatedMetrics, start, end time.Time) {
	// Simplified pattern detection
	patterns := []string{}
	
	// High activity pattern
	if agg.AvgSpreadsFound > 20 {
		patterns = append(patterns, "high_opportunity_period")
	}
	
	// Efficiency pattern
	if agg.CacheHitRate > 0.9 {
		patterns = append(patterns, "high_cache_efficiency")
	}
	
	// Error pattern
	if agg.ErrorRate > 0.05 {
		patterns = append(patterns, "elevated_error_rate")
	}
	
	agg.DetectedPatterns = patterns
}

// Export methods

// ExportJSON exports aggregated metrics as JSON
func (a *Aggregator) ExportJSON(agg *AggregatedMetrics, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(agg)
}

// ExportCSV exports aggregated metrics as CSV
func (a *Aggregator) ExportCSV(agg *AggregatedMetrics, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()
	
	// Write headers
	headers := []string{
		"Period", "Total Scans", "Avg Scan Rate", "Total Results",
		"Unique Symbols", "Avg Spreads Found", "Cache Hit Rate",
		"Error Rate", "Avg Connections",
	}
	if err := writer.Write(headers); err != nil {
		return err
	}
	
	// Write data
	record := []string{
		agg.Period,
		fmt.Sprintf("%d", agg.TotalScans),
		fmt.Sprintf("%.2f", agg.AvgScanRate),
		fmt.Sprintf("%d", agg.TotalResults),
		fmt.Sprintf("%d", agg.UniqueSymbols),
		fmt.Sprintf("%.2f", agg.AvgSpreadsFound),
		fmt.Sprintf("%.2f%%", agg.CacheHitRate*100),
		fmt.Sprintf("%.2f%%", agg.ErrorRate*100),
		fmt.Sprintf("%.1f", agg.AvgConnections),
	}
	if err := writer.Write(record); err != nil {
		return err
	}
	
	// Write filter stats
	writer.Write([]string{}) // Empty line
	writer.Write([]string{"Filter Statistics"})
	writer.Write([]string{"Filter", "Executions", "Avg Duration (ms)", "Avg Reduction", "Total Filtered"})
	
	for name, stats := range agg.FilterStats {
		record := []string{
			name,
			fmt.Sprintf("%d", stats.Executions),
			fmt.Sprintf("%.2f", stats.AvgDuration),
			fmt.Sprintf("%.2f%%", stats.AvgReduction*100),
			fmt.Sprintf("%d", stats.TotalFiltered),
		}
		writer.Write(record)
	}
	
	// Write top symbols
	writer.Write([]string{}) // Empty line
	writer.Write([]string{"Top Symbols"})
	writer.Write([]string{"Symbol", "Scan Count", "Avg Spreads", "Best Score"})
	
	for _, sym := range agg.TopSymbols {
		record := []string{
			sym.Symbol,
			fmt.Sprintf("%d", sym.ScanCount),
			fmt.Sprintf("%.1f", sym.AvgSpreads),
			fmt.Sprintf("%.3f", sym.BestScore),
		}
		writer.Write(record)
	}
	
	return nil
}

// GenerateReport generates a comprehensive report
func (a *Aggregator) GenerateReport(period string) (*Report, error) {
	var start, end time.Time
	
	// Parse period
	switch period {
	case "daily":
		end = time.Now()
		start = end.AddDate(0, 0, -1)
	case "weekly":
		end = time.Now()
		start = end.AddDate(0, 0, -7)
	case "monthly":
		end = time.Now()
		start = end.AddDate(0, -1, 0)
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}
	
	// Aggregate metrics
	agg, err := a.Aggregate(start, end)
	if err != nil {
		return nil, err
	}
	
	// Generate report
	report := &Report{
		Title:       fmt.Sprintf("Scanner Performance Report - %s", period),
		Period:      period,
		Generated:   time.Now(),
		Metrics:     agg,
		Summary:     a.generateSummary(agg),
		Insights:    a.generateInsights(agg),
		Recommendations: a.generateRecommendations(agg),
	}
	
	return report, nil
}

// Report represents a comprehensive performance report
type Report struct {
	Title           string              `json:"title"`
	Period          string              `json:"period"`
	Generated       time.Time           `json:"generated"`
	Metrics         *AggregatedMetrics  `json:"metrics"`
	Summary         string              `json:"summary"`
	Insights        []string            `json:"insights"`
	Recommendations []string            `json:"recommendations"`
}

// generateSummary generates executive summary
func (a *Aggregator) generateSummary(agg *AggregatedMetrics) string {
	return fmt.Sprintf(
		"During the period %s, the scanner performed %d scans across %d unique symbols "+
		"with an average scan rate of %.1f scans/second. The system found an average of %.1f spreads per scan "+
		"with a cache hit rate of %.1f%% and error rate of %.1f%%.",
		agg.Period, agg.TotalScans, agg.UniqueSymbols, agg.AvgScanRate,
		agg.AvgSpreadsFound, agg.CacheHitRate*100, agg.ErrorRate*100,
	)
}

// generateInsights generates data insights
func (a *Aggregator) generateInsights(agg *AggregatedMetrics) []string {
	insights := []string{}
	
	// Performance insights
	if agg.CacheHitRate > 0.8 {
		insights = append(insights, fmt.Sprintf("Excellent cache performance with %.1f%% hit rate", agg.CacheHitRate*100))
	}
	
	if agg.ErrorRate < 0.01 {
		insights = append(insights, "Very low error rate indicates stable operation")
	}
	
	// Market insights
	for regime, pct := range agg.MarketRegimes {
		if pct > 0.5 {
			insights = append(insights, fmt.Sprintf("Market was predominantly in %s regime (%.1f%% of time)", regime, pct*100))
		}
	}
	
	// Top performer insights
	if len(agg.TopSymbols) > 0 && agg.TopSymbols[0].BestScore > 0.9 {
		insights = append(insights, fmt.Sprintf("%s showed exceptional opportunities with score %.3f",
			agg.TopSymbols[0].Symbol, agg.TopSymbols[0].BestScore))
	}
	
	return insights
}

// generateRecommendations generates actionable recommendations
func (a *Aggregator) generateRecommendations(agg *AggregatedMetrics) []string {
	recommendations := []string{}
	
	// Performance recommendations
	if agg.CacheHitRate < 0.7 {
		recommendations = append(recommendations, "Consider increasing cache size to improve hit rate")
	}
	
	if agg.ErrorRate > 0.05 {
		recommendations = append(recommendations, "Investigate and address sources of errors")
	}
	
	if agg.AvgScanDuration > 200 {
		recommendations = append(recommendations, "Optimize scan performance to reduce latency")
	}
	
	// Filter recommendations
	for name, stats := range agg.FilterStats {
		if stats.AvgDuration > 10 {
			recommendations = append(recommendations, fmt.Sprintf("Optimize %s filter - currently taking %.1fms", name, stats.AvgDuration))
		}
	}
	
	// Trading recommendations
	if len(agg.TopSpreads) > 10 {
		recommendations = append(recommendations, "Many recurring opportunities detected - consider automation")
	}
	
	return recommendations
}