package scoring

import (
	"math"
	"github.com/new-ibkr-trader/src/go/pkg/models"
)

// ScoringConfig defines weights for different scoring factors
type ScoringConfig struct {
	// Base weights
	ProbabilityWeight   float64 `json:"probability_weight"`
	RiskRewardWeight    float64 `json:"risk_reward_weight"`
	LiquidityWeight     float64 `json:"liquidity_weight"`
	
	// Greeks weights
	DeltaWeight         float64 `json:"delta_weight"`
	ThetaWeight         float64 `json:"theta_weight"`
	VegaWeight          float64 `json:"vega_weight"`
	
	// Advanced weights
	VolatilityWeight    float64 `json:"volatility_weight"`
	SpreadWidthWeight   float64 `json:"spread_width_weight"`
	TimeDecayWeight     float64 `json:"time_decay_weight"`
}

// DefaultScoringConfig returns a balanced scoring configuration
func DefaultScoringConfig() *ScoringConfig {
	return &ScoringConfig{
		ProbabilityWeight:   0.25,
		RiskRewardWeight:    0.20,
		LiquidityWeight:     0.15,
		DeltaWeight:         0.10,
		ThetaWeight:         0.10,
		VegaWeight:          0.05,
		VolatilityWeight:    0.05,
		SpreadWidthWeight:   0.05,
		TimeDecayWeight:     0.05,
	}
}

// ConservativeScoringConfig returns a risk-averse scoring configuration
func ConservativeScoringConfig() *ScoringConfig {
	return &ScoringConfig{
		ProbabilityWeight:   0.35,  // Higher weight on probability
		RiskRewardWeight:    0.15,
		LiquidityWeight:     0.20,  // Higher weight on liquidity
		DeltaWeight:         0.10,
		ThetaWeight:         0.05,
		VegaWeight:          0.05,
		VolatilityWeight:    0.03,
		SpreadWidthWeight:   0.05,
		TimeDecayWeight:     0.02,
	}
}

// AggressiveScoringConfig returns a profit-focused scoring configuration
func AggressiveScoringConfig() *ScoringConfig {
	return &ScoringConfig{
		ProbabilityWeight:   0.15,
		RiskRewardWeight:    0.30,  // Higher weight on risk/reward
		LiquidityWeight:     0.10,
		DeltaWeight:         0.15,  // Higher weight on delta
		ThetaWeight:         0.15,  // Higher weight on theta
		VegaWeight:          0.05,
		VolatilityWeight:    0.05,
		SpreadWidthWeight:   0.03,
		TimeDecayWeight:     0.02,
	}
}

// Scorer calculates spread scores using configurable weights
type Scorer struct {
	config *ScoringConfig
}

// NewScorer creates a new scorer with the given configuration
func NewScorer(config *ScoringConfig) *Scorer {
	if config == nil {
		config = DefaultScoringConfig()
	}
	return &Scorer{config: config}
}

// ScoreSpread calculates a comprehensive score for a vertical spread
func (s *Scorer) ScoreSpread(spread *models.VerticalSpread) float64 {
	// Calculate individual component scores (0-100 scale)
	probabilityScore := s.scoreProbability(spread.ProbabilityProfit)
	riskRewardScore := s.scoreRiskReward(spread)
	liquidityScore := s.scoreLiquidity(spread)
	deltaScore := s.scoreDelta(spread)
	thetaScore := s.scoreTheta(spread)
	vegaScore := s.scoreVega(spread)
	volatilityScore := s.scoreVolatility(spread)
	spreadWidthScore := s.scoreSpreadWidth(spread)
	timeDecayScore := s.scoreTimeDecay(spread)
	
	// Apply weights and calculate final score
	weightedScore := 
		probabilityScore * s.config.ProbabilityWeight +
		riskRewardScore * s.config.RiskRewardWeight +
		liquidityScore * s.config.LiquidityWeight +
		deltaScore * s.config.DeltaWeight +
		thetaScore * s.config.ThetaWeight +
		vegaScore * s.config.VegaWeight +
		volatilityScore * s.config.VolatilityWeight +
		spreadWidthScore * s.config.SpreadWidthWeight +
		timeDecayScore * s.config.TimeDecayWeight
	
	return math.Round(weightedScore*100) / 100 // Round to 2 decimal places
}

// scoreProbability scores based on probability of profit
func (s *Scorer) scoreProbability(probability float64) float64 {
	// Linear scaling: 50% PoP = 0 score, 90% PoP = 100 score
	if probability < 0.5 {
		return 0
	}
	return math.Min(100, (probability-0.5)*250)
}

// scoreRiskReward scores based on risk/reward ratio
func (s *Scorer) scoreRiskReward(spread *models.VerticalSpread) float64 {
	if spread.MaxLoss == 0 {
		return 0
	}
	
	ratio := spread.MaxProfit / spread.MaxLoss
	// Ratio of 1:1 = 50 score, 2:1 = 100 score
	score := ratio * 50
	return math.Min(100, score)
}

// scoreLiquidity scores based on volume and open interest
func (s *Scorer) scoreLiquidity(spread *models.VerticalSpread) float64 {
	// Average liquidity metrics of both legs
	avgVolume := float64(spread.LongLeg.Volume + spread.ShortLeg.Volume) / 2
	avgOI := float64(spread.LongLeg.OpenInterest + spread.ShortLeg.OpenInterest) / 2
	
	// Volume score (0-50 points)
	volumeScore := math.Min(50, avgVolume/100)
	
	// Open Interest score (0-50 points)
	oiScore := math.Min(50, avgOI/1000)
	
	return volumeScore + oiScore
}

// scoreDelta scores based on delta positioning
func (s *Scorer) scoreDelta(spread *models.VerticalSpread) float64 {
	// Ideal delta for long leg: 0.25-0.40
	longDelta := math.Abs(spread.LongLeg.Delta)
	
	if longDelta < 0.20 || longDelta > 0.50 {
		return 0
	}
	
	// Peak score at 0.30-0.35 delta
	if longDelta >= 0.30 && longDelta <= 0.35 {
		return 100
	}
	
	// Linear falloff outside ideal range
	if longDelta < 0.30 {
		return (longDelta - 0.20) * 1000 // 0.20->0.30 maps to 0->100
	}
	
	return (0.50 - longDelta) * 667 // 0.35->0.50 maps to 100->0
}

// scoreTheta scores based on theta (time decay)
func (s *Scorer) scoreTheta(spread *models.VerticalSpread) float64 {
	// Net theta should be positive (we collect time decay)
	netTheta := spread.ShortLeg.Theta - spread.LongLeg.Theta
	
	if netTheta <= 0 {
		return 0
	}
	
	// Score based on daily theta as percentage of net debit
	if spread.NetDebit > 0 {
		thetaPercent := (netTheta / spread.NetDebit) * 100
		// 1% daily theta = 100 score
		return math.Min(100, thetaPercent * 100)
	}
	
	return 50 // Default score if we can't calculate percentage
}

// scoreVega scores based on vega exposure
func (s *Scorer) scoreVega(spread *models.VerticalSpread) float64 {
	// Net vega should be low (neutral to volatility)
	netVega := math.Abs(spread.LongLeg.Vega - spread.ShortLeg.Vega)
	
	// Lower vega = higher score
	if netVega < 0.05 {
		return 100
	}
	
	if netVega > 0.20 {
		return 0
	}
	
	// Linear scale between 0.05 and 0.20
	return (0.20 - netVega) * 667
}

// scoreVolatility scores based on IV levels
func (s *Scorer) scoreVolatility(spread *models.VerticalSpread) float64 {
	// Average IV of both legs
	avgIV := (spread.LongLeg.IV + spread.ShortLeg.IV) / 2
	
	// Ideal IV range: 0.20-0.40 (20-40%)
	if avgIV < 0.15 || avgIV > 0.60 {
		return 0
	}
	
	if avgIV >= 0.20 && avgIV <= 0.40 {
		return 100
	}
	
	if avgIV < 0.20 {
		return (avgIV - 0.15) * 2000
	}
	
	return (0.60 - avgIV) * 250
}

// scoreSpreadWidth scores based on strike width
func (s *Scorer) scoreSpreadWidth(spread *models.VerticalSpread) float64 {
	width := spread.ShortLeg.Strike - spread.LongLeg.Strike
	
	// Ideal width depends on underlying price
	// For now, use fixed ranges
	if width < 1 {
		return 0 // Too narrow
	}
	
	if width > 10 {
		return 0 // Too wide
	}
	
	// Peak score at 2.5-5 width
	if width >= 2.5 && width <= 5 {
		return 100
	}
	
	if width < 2.5 {
		return (width - 1) * 67
	}
	
	return (10 - width) * 20
}

// scoreTimeDecay scores based on DTE and decay curve
func (s *Scorer) scoreTimeDecay(spread *models.VerticalSpread) float64 {
	// This would use actual DTE from expiry date
	// For now, return a default score
	return 75
}

// ScoreReport provides detailed scoring breakdown
type ScoreReport struct {
	TotalScore       float64            `json:"total_score"`
	ComponentScores  map[string]float64 `json:"component_scores"`
	Recommendations  []string           `json:"recommendations"`
}

// GenerateReport creates a detailed scoring report
func (s *Scorer) GenerateReport(spread *models.VerticalSpread) *ScoreReport {
	report := &ScoreReport{
		ComponentScores:  make(map[string]float64),
		Recommendations:  []string{},
	}
	
	// Calculate all component scores
	report.ComponentScores["probability"] = s.scoreProbability(spread.ProbabilityProfit)
	report.ComponentScores["risk_reward"] = s.scoreRiskReward(spread)
	report.ComponentScores["liquidity"] = s.scoreLiquidity(spread)
	report.ComponentScores["delta"] = s.scoreDelta(spread)
	report.ComponentScores["theta"] = s.scoreTheta(spread)
	report.ComponentScores["vega"] = s.scoreVega(spread)
	report.ComponentScores["volatility"] = s.scoreVolatility(spread)
	report.ComponentScores["spread_width"] = s.scoreSpreadWidth(spread)
	report.ComponentScores["time_decay"] = s.scoreTimeDecay(spread)
	
	// Calculate total score
	report.TotalScore = s.ScoreSpread(spread)
	
	// Generate recommendations
	if report.ComponentScores["probability"] < 50 {
		report.Recommendations = append(report.Recommendations, 
			"Low probability of profit - consider more conservative strikes")
	}
	
	if report.ComponentScores["liquidity"] < 30 {
		report.Recommendations = append(report.Recommendations,
			"Low liquidity - wider bid-ask spreads expected")
	}
	
	if report.ComponentScores["theta"] < 20 {
		report.Recommendations = append(report.Recommendations,
			"Low theta collection - consider different strike selection")
	}
	
	return report
}