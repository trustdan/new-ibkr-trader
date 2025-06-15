package analytics

import (
	"math"
	"sort"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/models"
)

// Analyzer provides advanced analytics for scan results
type Analyzer struct {
	// Moving averages
	scoreMA      *MovingAverage
	spreadCountMA *MovingAverage
	creditMA     *MovingAverage
	
	// Volatility tracking
	volTracker   *VolatilityTracker
	
	// Pattern detection
	patterns     []PatternDetector
	
	// Market regime
	regimeDetector *MarketRegimeDetector
}

// AnalysisResult contains comprehensive analysis
type AnalysisResult struct {
	Timestamp       time.Time              `json:"timestamp"`
	Symbol          string                 `json:"symbol"`
	
	// Basic metrics
	SpreadCount     int                    `json:"spread_count"`
	TopScore        float64                `json:"top_score"`
	AvgScore        float64                `json:"avg_score"`
	TopCredit       float64                `json:"top_credit"`
	AvgCredit       float64                `json:"avg_credit"`
	
	// Advanced metrics
	ScoreDistribution  Distribution         `json:"score_distribution"`
	CreditDistribution Distribution         `json:"credit_distribution"`
	
	// Moving averages
	ScoreMA         float64                `json:"score_ma"`
	SpreadCountMA   float64                `json:"spread_count_ma"`
	CreditMA        float64                `json:"credit_ma"`
	
	// Volatility metrics
	ScoreVolatility float64                `json:"score_volatility"`
	SpreadVolatility float64               `json:"spread_volatility"`
	
	// Market analysis
	MarketRegime    string                 `json:"market_regime"`
	RegimeConfidence float64               `json:"regime_confidence"`
	
	// Opportunity ranking
	OpportunityScore float64               `json:"opportunity_score"`
	Rank            int                    `json:"rank"`
	
	// Patterns detected
	Patterns        []DetectedPattern      `json:"patterns"`
	
	// Recommendations
	Recommendations []string               `json:"recommendations"`
}

// Distribution represents statistical distribution
type Distribution struct {
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	Mean       float64   `json:"mean"`
	Median     float64   `json:"median"`
	StdDev     float64   `json:"std_dev"`
	Percentiles map[int]float64 `json:"percentiles"`
}

// DetectedPattern represents a detected trading pattern
type DetectedPattern struct {
	Type        string    `json:"type"`
	Confidence  float64   `json:"confidence"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		scoreMA:       NewMovingAverage(20),
		spreadCountMA: NewMovingAverage(20),
		creditMA:      NewMovingAverage(20),
		volTracker:    NewVolatilityTracker(20),
		patterns:      initPatternDetectors(),
		regimeDetector: NewMarketRegimeDetector(),
	}
}

// Analyze performs comprehensive analysis on scan results
func (a *Analyzer) Analyze(result models.ScanResult, historicalData []models.ScanResult) AnalysisResult {
	analysis := AnalysisResult{
		Timestamp:    result.Timestamp,
		Symbol:       result.Symbol,
		SpreadCount:  len(result.Spreads),
	}
	
	// Basic metrics
	if len(result.Spreads) > 0 {
		analysis.TopScore, analysis.AvgScore = a.calculateScoreMetrics(result.Spreads)
		analysis.TopCredit, analysis.AvgCredit = a.calculateCreditMetrics(result.Spreads)
		
		// Distributions
		analysis.ScoreDistribution = a.calculateDistribution(a.extractScores(result.Spreads))
		analysis.CreditDistribution = a.calculateDistribution(a.extractCredits(result.Spreads))
	}
	
	// Update moving averages
	a.scoreMA.Add(analysis.AvgScore)
	a.spreadCountMA.Add(float64(analysis.SpreadCount))
	a.creditMA.Add(analysis.AvgCredit)
	
	analysis.ScoreMA = a.scoreMA.Value()
	analysis.SpreadCountMA = a.spreadCountMA.Value()
	analysis.CreditMA = a.creditMA.Value()
	
	// Volatility analysis
	a.volTracker.Update(analysis.AvgScore)
	analysis.ScoreVolatility = a.volTracker.GetVolatility()
	analysis.SpreadVolatility = a.calculateSpreadVolatility(historicalData)
	
	// Market regime detection
	regime, confidence := a.regimeDetector.DetectRegime(result, historicalData)
	analysis.MarketRegime = regime
	analysis.RegimeConfidence = confidence
	
	// Pattern detection
	analysis.Patterns = a.detectPatterns(result, historicalData)
	
	// Calculate opportunity score
	analysis.OpportunityScore = a.calculateOpportunityScore(analysis)
	
	// Generate recommendations
	analysis.Recommendations = a.generateRecommendations(analysis)
	
	return analysis
}

// calculateScoreMetrics calculates score metrics
func (a *Analyzer) calculateScoreMetrics(spreads []models.VerticalSpread) (top, avg float64) {
	if len(spreads) == 0 {
		return 0, 0
	}
	
	sum := 0.0
	top = spreads[0].Score
	
	for _, spread := range spreads {
		sum += spread.Score
		if spread.Score > top {
			top = spread.Score
		}
	}
	
	avg = sum / float64(len(spreads))
	return
}

// calculateCreditMetrics calculates credit metrics
func (a *Analyzer) calculateCreditMetrics(spreads []models.VerticalSpread) (top, avg float64) {
	if len(spreads) == 0 {
		return 0, 0
	}
	
	sum := 0.0
	top = spreads[0].Credit
	
	for _, spread := range spreads {
		sum += spread.Credit
		if spread.Credit > top {
			top = spread.Credit
		}
	}
	
	avg = sum / float64(len(spreads))
	return
}

// extractScores extracts scores from spreads
func (a *Analyzer) extractScores(spreads []models.VerticalSpread) []float64 {
	scores := make([]float64, len(spreads))
	for i, spread := range spreads {
		scores[i] = spread.Score
	}
	return scores
}

// extractCredits extracts credits from spreads
func (a *Analyzer) extractCredits(spreads []models.VerticalSpread) []float64 {
	credits := make([]float64, len(spreads))
	for i, spread := range spreads {
		credits[i] = spread.Credit
	}
	return credits
}

// calculateDistribution calculates statistical distribution
func (a *Analyzer) calculateDistribution(values []float64) Distribution {
	if len(values) == 0 {
		return Distribution{}
	}
	
	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	dist := Distribution{
		Min: sorted[0],
		Max: sorted[len(sorted)-1],
		Mean: a.mean(values),
		Median: a.median(sorted),
		StdDev: a.stdDev(values),
		Percentiles: make(map[int]float64),
	}
	
	// Calculate percentiles
	percentiles := []int{10, 25, 50, 75, 90}
	for _, p := range percentiles {
		dist.Percentiles[p] = a.percentile(sorted, p)
	}
	
	return dist
}

// Statistical helper functions
func (a *Analyzer) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (a *Analyzer) median(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func (a *Analyzer) stdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	
	mean := a.mean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

func (a *Analyzer) percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	
	index := float64(p) / 100 * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sorted[lower]
	}
	
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// calculateSpreadVolatility calculates volatility of spread counts
func (a *Analyzer) calculateSpreadVolatility(historical []models.ScanResult) float64 {
	if len(historical) < 2 {
		return 0
	}
	
	counts := make([]float64, len(historical))
	for i, result := range historical {
		counts[i] = float64(len(result.Spreads))
	}
	
	return a.stdDev(counts)
}

// detectPatterns detects trading patterns
func (a *Analyzer) detectPatterns(current models.ScanResult, historical []models.ScanResult) []DetectedPattern {
	patterns := make([]DetectedPattern, 0)
	
	for _, detector := range a.patterns {
		if pattern := detector.Detect(current, historical); pattern != nil {
			patterns = append(patterns, *pattern)
		}
	}
	
	return patterns
}

// calculateOpportunityScore calculates overall opportunity score
func (a *Analyzer) calculateOpportunityScore(analysis AnalysisResult) float64 {
	score := 0.0
	
	// Factor in spread count (normalized)
	spreadScore := math.Min(float64(analysis.SpreadCount)/20.0, 1.0) * 0.2
	score += spreadScore
	
	// Factor in average score
	scoreScore := analysis.AvgScore * 0.3
	score += scoreScore
	
	// Factor in top credit (normalized)
	creditScore := math.Min(analysis.TopCredit/5.0, 1.0) * 0.2
	score += creditScore
	
	// Factor in consistency (inverse of volatility)
	if analysis.ScoreVolatility > 0 {
		consistencyScore := (1.0 - math.Min(analysis.ScoreVolatility, 1.0)) * 0.2
		score += consistencyScore
	}
	
	// Factor in market regime
	regimeBonus := 0.0
	switch analysis.MarketRegime {
	case "high_volatility":
		regimeBonus = 0.1
	case "trending":
		regimeBonus = 0.05
	}
	score += regimeBonus
	
	return math.Min(score, 1.0)
}

// generateRecommendations generates trading recommendations
func (a *Analyzer) generateRecommendations(analysis AnalysisResult) []string {
	recommendations := make([]string, 0)
	
	// High opportunity score
	if analysis.OpportunityScore > 0.8 {
		recommendations = append(recommendations, "High opportunity - consider increasing position size")
	}
	
	// High volatility
	if analysis.ScoreVolatility > 0.3 {
		recommendations = append(recommendations, "High volatility - use tighter risk management")
	}
	
	// Many spreads available
	if analysis.SpreadCount > 20 {
		recommendations = append(recommendations, "Many opportunities - be selective with best scores")
	}
	
	// Pattern-based recommendations
	for _, pattern := range analysis.Patterns {
		if pattern.Confidence > 0.7 && pattern.Action != "" {
			recommendations = append(recommendations, pattern.Action)
		}
	}
	
	// Market regime recommendations
	switch analysis.MarketRegime {
	case "high_volatility":
		recommendations = append(recommendations, "High volatility regime - focus on premium collection")
	case "low_volatility":
		recommendations = append(recommendations, "Low volatility regime - consider longer DTE")
	case "trending":
		recommendations = append(recommendations, "Trending market - watch for directional risk")
	}
	
	return recommendations
}

// MovingAverage calculates simple moving average
type MovingAverage struct {
	window int
	values []float64
}

func NewMovingAverage(window int) *MovingAverage {
	return &MovingAverage{
		window: window,
		values: make([]float64, 0, window),
	}
}

func (ma *MovingAverage) Add(value float64) {
	ma.values = append(ma.values, value)
	if len(ma.values) > ma.window {
		ma.values = ma.values[len(ma.values)-ma.window:]
	}
}

func (ma *MovingAverage) Value() float64 {
	if len(ma.values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range ma.values {
		sum += v
	}
	return sum / float64(len(ma.values))
}

// VolatilityTracker tracks volatility
type VolatilityTracker struct {
	ma *MovingAverage
	values []float64
	window int
}

func NewVolatilityTracker(window int) *VolatilityTracker {
	return &VolatilityTracker{
		ma:     NewMovingAverage(window),
		values: make([]float64, 0, window),
		window: window,
	}
}

func (vt *VolatilityTracker) Update(value float64) {
	vt.values = append(vt.values, value)
	if len(vt.values) > vt.window {
		vt.values = vt.values[len(vt.values)-vt.window:]
	}
	vt.ma.Add(value)
}

func (vt *VolatilityTracker) GetVolatility() float64 {
	if len(vt.values) < 2 {
		return 0
	}
	
	mean := vt.ma.Value()
	sumSquares := 0.0
	for _, v := range vt.values {
		diff := v - mean
		sumSquares += diff * diff
	}
	
	return math.Sqrt(sumSquares / float64(len(vt.values)-1))
}

// Pattern detectors
type PatternDetector interface {
	Detect(current models.ScanResult, historical []models.ScanResult) *DetectedPattern
}

func initPatternDetectors() []PatternDetector {
	return []PatternDetector{
		&TrendPatternDetector{},
		&VolumePatternDetector{},
		&ConsistencyPatternDetector{},
	}
}

// TrendPatternDetector detects trending patterns
type TrendPatternDetector struct{}

func (d *TrendPatternDetector) Detect(current models.ScanResult, historical []models.ScanResult) *DetectedPattern {
	if len(historical) < 5 {
		return nil
	}
	
	// Check for increasing spread counts
	increasing := 0
	for i := len(historical) - 5; i < len(historical)-1; i++ {
		if len(historical[i+1].Spreads) > len(historical[i].Spreads) {
			increasing++
		}
	}
	
	if increasing >= 3 {
		return &DetectedPattern{
			Type:        "increasing_opportunities",
			Confidence:  float64(increasing) / 4.0,
			Description: "Opportunities are increasing",
			Action:      "Monitor closely for entry points",
		}
	}
	
	return nil
}

// VolumePatternDetector detects volume patterns
type VolumePatternDetector struct{}

func (d *VolumePatternDetector) Detect(current models.ScanResult, historical []models.ScanResult) *DetectedPattern {
	// Simplified - would check actual volume data
	if len(current.Spreads) > 30 {
		return &DetectedPattern{
			Type:        "high_activity",
			Confidence:  0.8,
			Description: "Unusually high option activity",
			Action:      "Check for news or events",
		}
	}
	return nil
}

// ConsistencyPatternDetector detects consistency patterns
type ConsistencyPatternDetector struct{}

func (d *ConsistencyPatternDetector) Detect(current models.ScanResult, historical []models.ScanResult) *DetectedPattern {
	if len(historical) < 10 {
		return nil
	}
	
	// Check spread count consistency
	counts := make([]float64, len(historical))
	for i, result := range historical {
		counts[i] = float64(len(result.Spreads))
	}
	
	mean := 0.0
	for _, c := range counts {
		mean += c
	}
	mean /= float64(len(counts))
	
	// Calculate standard deviation
	var sumSquares float64
	for _, c := range counts {
		diff := c - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(counts)-1))
	
	// Low volatility indicates consistency
	if stdDev/mean < 0.2 {
		return &DetectedPattern{
			Type:        "consistent_opportunities",
			Confidence:  0.9,
			Description: "Stable opportunity flow",
			Action:      "Good for systematic trading",
		}
	}
	
	return nil
}

// MarketRegimeDetector detects market regimes
type MarketRegimeDetector struct {
	volThresholds map[string]float64
}

func NewMarketRegimeDetector() *MarketRegimeDetector {
	return &MarketRegimeDetector{
		volThresholds: map[string]float64{
			"low":  0.15,
			"high": 0.30,
		},
	}
}

func (d *MarketRegimeDetector) DetectRegime(current models.ScanResult, historical []models.ScanResult) (regime string, confidence float64) {
	// Simplified regime detection based on spread metrics
	if len(current.Spreads) == 0 {
		return "unknown", 0.0
	}
	
	// Calculate average IV from spreads
	avgIV := 0.0
	for _, spread := range current.Spreads {
		avgIV += (spread.ShortLeg.IV + spread.LongLeg.IV) / 2
	}
	avgIV /= float64(len(current.Spreads))
	
	// Determine regime based on IV levels
	if avgIV < d.volThresholds["low"] {
		return "low_volatility", 0.8
	} else if avgIV > d.volThresholds["high"] {
		return "high_volatility", 0.8
	}
	
	// Check for trending behavior
	if len(historical) >= 5 {
		// Simplified trend detection
		recentCounts := make([]int, 5)
		for i := 0; i < 5; i++ {
			idx := len(historical) - 5 + i
			if idx >= 0 {
				recentCounts[i] = len(historical[idx].Spreads)
			}
		}
		
		// Check if counts are increasing or decreasing
		increasing := 0
		decreasing := 0
		for i := 1; i < len(recentCounts); i++ {
			if recentCounts[i] > recentCounts[i-1] {
				increasing++
			} else if recentCounts[i] < recentCounts[i-1] {
				decreasing++
			}
		}
		
		if increasing >= 3 || decreasing >= 3 {
			return "trending", 0.7
		}
	}
	
	return "sideways", 0.6
}