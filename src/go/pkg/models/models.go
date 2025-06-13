package models

import (
	"time"
)

// Option represents an options contract
type Option struct {
	Symbol           string    `json:"symbol"`
	Underlying       string    `json:"underlying"`
	Strike           float64   `json:"strike"`
	Expiration       time.Time `json:"expiration"`
	OptionType       string    `json:"option_type"` // "call" or "put"
	DTE              int       `json:"dte"`         // Days to expiration
	
	// Market data
	Bid              float64   `json:"bid"`
	Ask              float64   `json:"ask"`
	Last             float64   `json:"last"`
	Volume           int64     `json:"volume"`
	OpenInterest     int64     `json:"open_interest"`
	
	// Greeks
	Delta            float64   `json:"delta"`
	Gamma            float64   `json:"gamma"`
	Theta            float64   `json:"theta"`
	Vega             float64   `json:"vega"`
	IV               float64   `json:"iv"`              // Implied volatility
	IVPercentile     float64   `json:"iv_percentile"`
	
	// Calculated fields
	BidAskSpread     float64   `json:"bid_ask_spread"`
	BidAskSpreadPct  float64   `json:"bid_ask_spread_pct"`
	Liquidity        float64   `json:"liquidity"`        // Combined volume/OI metric
	ProbabilityITM   float64   `json:"probability_itm"`
	ExpectedMove     float64   `json:"expected_move"`
}

// ScanRequest represents a scan request
type ScanRequest struct {
	Symbol   string         `json:"symbol" binding:"required"`
	Filters  []FilterConfig `json:"filters" binding:"required"`
	MaxResults int          `json:"max_results,omitempty"`
}

// FilterConfig represents a filter configuration
type FilterConfig struct {
	Type     string                 `json:"type" binding:"required"`
	Params   map[string]interface{} `json:"params" binding:"required"`
	Priority int                    `json:"priority,omitempty"`
}

// ScanResponse represents scan results
type ScanResponse struct {
	Symbol      string    `json:"symbol"`
	ScanTime    time.Time `json:"scan_time"`
	ResultCount int       `json:"result_count"`
	Options     []Option  `json:"options"`
}

// SpreadStrategy represents a vertical spread
type SpreadStrategy struct {
	Type        string    `json:"type"` // "debit" or "credit"
	LongLeg     Option    `json:"long_leg"`
	ShortLeg    Option    `json:"short_leg"`
	
	// Spread metrics
	NetDebit    float64   `json:"net_debit"`
	NetCredit   float64   `json:"net_credit"`
	MaxProfit   float64   `json:"max_profit"`
	MaxLoss     float64   `json:"max_loss"`
	Breakeven   float64   `json:"breakeven"`
	RiskReward  float64   `json:"risk_reward"`
	PoP         float64   `json:"pop"` // Probability of profit
}

// HealthCheck represents service health status
type HealthCheck struct {
	Status      string    `json:"status"`
	Service     string    `json:"service"`
	Uptime      string    `json:"uptime"`
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
}