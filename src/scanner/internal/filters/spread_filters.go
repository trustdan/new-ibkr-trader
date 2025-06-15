package filters

import (
	"math"
	"github.com/ibkr-trader/scanner/internal/models"
)

// RiskRewardFilter filters spreads by risk/reward ratio
type RiskRewardFilter struct {
	MinRatio float64 `json:"min_ratio"`
	MaxRatio float64 `json:"max_ratio"`
}

func (f *RiskRewardFilter) Name() string { return "RiskRewardFilter" }

func (f *RiskRewardFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Calculate risk/reward ratio
	maxProfit := spread.Credit
	maxLoss := (spread.ShortLeg.Strike - spread.LongLeg.Strike) - spread.Credit
	
	if maxLoss <= 0 {
		return false // Invalid spread
	}
	
	ratio := maxProfit / maxLoss
	return ratio >= f.MinRatio && ratio <= f.MaxRatio
}

func (f *RiskRewardFilter) Validate() error { return nil }

// BreakEvenFilter filters by breakeven distance
type BreakEvenFilter struct {
	MinDistance float64 `json:"min_distance"` // Minimum distance to breakeven in %
	MaxDistance float64 `json:"max_distance"` // Maximum distance to breakeven in %
}

func (f *BreakEvenFilter) Name() string { return "BreakEvenFilter" }

func (f *BreakEvenFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Calculate breakeven point
	breakeven := spread.ShortLeg.Strike - spread.Credit
	currentPrice := spread.UnderlyingPrice
	
	// Calculate distance as percentage
	distance := math.Abs((breakeven - currentPrice) / currentPrice * 100)
	
	return distance >= f.MinDistance && distance <= f.MaxDistance
}

func (f *BreakEvenFilter) Validate() error { return nil }

// ExpectedValueFilter filters by expected value calculation
type ExpectedValueFilter struct {
	MinEV float64 `json:"min_ev"` // Minimum expected value
}

func (f *ExpectedValueFilter) Name() string { return "ExpectedValueFilter" }

func (f *ExpectedValueFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Calculate expected value
	maxProfit := spread.Credit
	maxLoss := (spread.ShortLeg.Strike - spread.LongLeg.Strike) - spread.Credit
	
	// Simple EV calculation: (Profit * PoP) - (Loss * (1 - PoP))
	ev := (maxProfit * spread.ProbOfProfit) - (maxLoss * (1 - spread.ProbOfProfit))
	
	return ev >= f.MinEV
}

func (f *ExpectedValueFilter) Validate() error { return nil }

// DeltaNeutralFilter filters for delta-neutral spreads
type DeltaNeutralFilter struct {
	MaxNetDelta float64 `json:"max_net_delta"` // Maximum absolute net delta
}

func (f *DeltaNeutralFilter) Name() string { return "DeltaNeutralFilter" }

func (f *DeltaNeutralFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Calculate net delta of the spread
	netDelta := math.Abs(spread.ShortLeg.Delta + spread.LongLeg.Delta)
	
	return netDelta <= f.MaxNetDelta
}

func (f *DeltaNeutralFilter) Validate() error { return nil }

// MarginEfficiencyFilter filters by margin efficiency
type MarginEfficiencyFilter struct {
	MinEfficiency float64 `json:"min_efficiency"` // Minimum return on margin
}

func (f *MarginEfficiencyFilter) Name() string { return "MarginEfficiencyFilter" }

func (f *MarginEfficiencyFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Calculate margin requirement (simplified)
	marginReq := (spread.ShortLeg.Strike - spread.LongLeg.Strike) * 100 // Per contract
	
	// Calculate return on margin
	returnOnMargin := (spread.Credit * 100) / marginReq
	
	return returnOnMargin >= f.MinEfficiency
}

func (f *MarginEfficiencyFilter) Validate() error { return nil }

// VolatilityEdgeFilter filters based on IV edge
type VolatilityEdgeFilter struct {
	MinIVDiff float64 `json:"min_iv_diff"` // Minimum IV difference between legs
}

func (f *VolatilityEdgeFilter) Name() string { return "VolatilityEdgeFilter" }

func (f *VolatilityEdgeFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Check for volatility edge (selling higher IV, buying lower IV)
	ivDiff := spread.ShortLeg.IV - spread.LongLeg.IV
	
	return ivDiff >= f.MinIVDiff
}

func (f *VolatilityEdgeFilter) Validate() error { return nil }

// CombinedGreeksFilter filters based on combined Greeks metrics
type CombinedGreeksFilter struct {
	MaxGammaRisk  float64 `json:"max_gamma_risk"`
	MaxVegaRisk   float64 `json:"max_vega_risk"`
	MinThetaDecay float64 `json:"min_theta_decay"`
}

func (f *CombinedGreeksFilter) Name() string { return "CombinedGreeksFilter" }

func (f *CombinedGreeksFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Calculate net Greeks
	netGamma := math.Abs(spread.ShortLeg.Gamma + spread.LongLeg.Gamma)
	netVega := math.Abs(spread.ShortLeg.Vega + spread.LongLeg.Vega)
	netTheta := spread.ShortLeg.Theta + spread.LongLeg.Theta // Want positive theta
	
	return netGamma <= f.MaxGammaRisk && 
	       netVega <= f.MaxVegaRisk && 
	       netTheta >= f.MinThetaDecay
}

func (f *CombinedGreeksFilter) Validate() error { return nil }

// LiquiditySpreadFilter ensures both legs have sufficient liquidity
type LiquiditySpreadFilter struct {
	MinBidAskRatio float64 `json:"min_bid_ask_ratio"` // Min bid/ask ratio for each leg
	MaxSpreadWidth float64 `json:"max_spread_width"`  // Max bid-ask spread width
}

func (f *LiquiditySpreadFilter) Name() string { return "LiquiditySpreadFilter" }

func (f *LiquiditySpreadFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	
	// Check bid-ask ratios for both legs
	shortBidAskRatio := spread.ShortLeg.Bid / spread.ShortLeg.Ask
	longBidAskRatio := spread.LongLeg.Bid / spread.LongLeg.Ask
	
	if shortBidAskRatio < f.MinBidAskRatio || longBidAskRatio < f.MinBidAskRatio {
		return false
	}
	
	// Check spread width
	shortSpreadWidth := (spread.ShortLeg.Ask - spread.ShortLeg.Bid) / spread.ShortLeg.Ask
	longSpreadWidth := (spread.LongLeg.Ask - spread.LongLeg.Bid) / spread.LongLeg.Ask
	
	return shortSpreadWidth <= f.MaxSpreadWidth && longSpreadWidth <= f.MaxSpreadWidth
}

func (f *LiquiditySpreadFilter) Validate() error { return nil }