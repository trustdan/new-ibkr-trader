package models

import (
	"time"
)

// OptionContract represents a single option contract with all relevant data
type OptionContract struct {
	// Basic identification
	Symbol       string    `json:"symbol"`
	ContractID   string    `json:"contract_id"`
	Strike       float64   `json:"strike"`
	Expiry       time.Time `json:"expiry"`
	OptionType   string    `json:"option_type"` // "CALL" or "PUT"
	Underlying   string    `json:"underlying"`
	
	// Market data
	Bid          float64   `json:"bid"`
	Ask          float64   `json:"ask"`
	Last         float64   `json:"last"`
	Volume       int64     `json:"volume"`
	OpenInterest int64     `json:"open_interest"`
	
	// Greeks
	Delta        float64   `json:"delta"`
	Gamma        float64   `json:"gamma"`
	Theta        float64   `json:"theta"`
	Vega         float64   `json:"vega"`
	Rho          float64   `json:"rho"`
	
	// Implied Volatility
	IV           float64   `json:"iv"`
	IVRank       float64   `json:"iv_rank"`
	IVPercentile float64   `json:"iv_percentile"`
	
	// Calculated fields
	DTE          int       `json:"dte"` // Days to expiration
	BidAskSpread float64   `json:"bid_ask_spread"`
	Moneyness    float64   `json:"moneyness"` // Distance from current price
	
	// Score for ranking
	Score        float64   `json:"score"`
	
	// Metadata
	LastUpdate   time.Time `json:"last_update"`
}

// VerticalSpread represents a vertical spread strategy
type VerticalSpread struct {
	Symbol       string          `json:"symbol"`
	LongLeg      OptionContract  `json:"long_leg"`
	ShortLeg     OptionContract  `json:"short_leg"`
	SpreadType   string          `json:"spread_type"` // "DEBIT" or "CREDIT"
	
	// Spread metrics
	Credit       float64         `json:"credit"`      // For credit spreads
	NetDebit     float64         `json:"net_debit"`
	MaxProfit    float64         `json:"max_profit"`
	MaxLoss      float64         `json:"max_loss"`
	Breakeven    float64         `json:"breakeven"`
	ProbOfProfit float64         `json:"prob_of_profit"`
	
	// Combined Greeks
	NetDelta     float64         `json:"net_delta"`
	NetTheta     float64         `json:"net_theta"`
	NetVega      float64         `json:"net_vega"`
	
	// Underlying info
	UnderlyingPrice float64      `json:"underlying_price"`
	
	// Score for ranking
	Score        float64         `json:"score"`
}

// ScanResult contains the results of a scan operation
type ScanResult struct {
	ScanID       string           `json:"scan_id"`
	Timestamp    time.Time        `json:"timestamp"`
	Symbol       string           `json:"symbol"`
	Spreads      []VerticalSpread `json:"spreads"`
	TotalFound   int              `json:"total_found"`
	Filtered     int              `json:"filtered"`
	Duration     time.Duration    `json:"duration"`
}