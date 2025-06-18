package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error     string `json:"error"`
	Status    int    `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Status, e.Error)
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data      interface{} `json:"data"`
	Status    string      `json:"status"`
	Timestamp int64       `json:"timestamp"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Status     string      `json:"status"`
	Timestamp  int64       `json:"timestamp"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
	Version   string  `json:"version"`
	Uptime    float64 `json:"uptime"`
}

// ScanRequest represents a scan request
type ScanRequest struct {
	Symbols []string               `json:"symbols"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// ScanFilters represents basic scan filters
type ScanFilters struct {
	DeltaMin float64 `json:"delta_min,omitempty"`
	DeltaMax float64 `json:"delta_max,omitempty"`
	DTEMin   int     `json:"dte_min,omitempty"`
	DTEMax   int     `json:"dte_max,omitempty"`
}

// ScanResult represents a scan result
type ScanResult struct {
	Symbol             string         `json:"symbol"`
	ScanTime           time.Time      `json:"scan_time"`
	TotalContracts     int            `json:"total_contracts"`
	FilteredContracts  int            `json:"filtered_contracts"`
	Spreads           []SpreadResult `json:"spreads"`
}

// SpreadResult represents a spread opportunity
type SpreadResult struct {
	LongStrike        float64 `json:"long_strike"`
	ShortStrike       float64 `json:"short_strike"`
	Expiration        string  `json:"expiration"`
	NetCredit         float64 `json:"net_credit"`
	MaxProfit         float64 `json:"max_profit"`
	MaxLoss           float64 `json:"max_loss"`
	ProbabilityProfit float64 `json:"probability_profit"`
	Score             float64 `json:"score"`
	DTE               int     `json:"dte"`
	SpreadWidth       float64 `json:"spread_width"`
	RiskReward        float64 `json:"risk_reward"`
}

// FilterConfig represents filter configuration
type FilterConfig struct {
	Delta     *DeltaFilter     `json:"delta,omitempty"`
	DTE       *DTEFilter       `json:"dte,omitempty"`
	Liquidity *LiquidityFilter `json:"liquidity,omitempty"`
	Spread    *SpreadFilter    `json:"spread,omitempty"`
	Advanced  *AdvancedFilter  `json:"advanced,omitempty"`
}

// DeltaFilter represents delta filter criteria
type DeltaFilter struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// DTEFilter represents days to expiration filter
type DTEFilter struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// LiquidityFilter represents liquidity filter criteria
type LiquidityFilter struct {
	MinOpenInterest int `json:"min_open_interest"`
	MinVolume       int `json:"min_volume"`
}

// SpreadFilter represents spread filter criteria
type SpreadFilter struct {
	MinCredit     float64 `json:"min_credit"`
	MaxWidth      float64 `json:"max_width"`
	MinRiskReward float64 `json:"min_risk_reward"`
}

// AdvancedFilter represents advanced filter criteria
type AdvancedFilter struct {
	MinPoP           float64 `json:"min_pop,omitempty"`
	MaxBidAskSpread  float64 `json:"max_bid_ask_spread,omitempty"`
	MinIVPercentile  float64 `json:"min_iv_percentile,omitempty"`
}

// FilterPreset represents a saved filter preset
type FilterPreset struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Filters     FilterConfig `json:"filters"`
	Tags        []string     `json:"tags"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	UsageCount  int          `json:"usage_count"`
}

// PresetRequest represents a preset creation request
type PresetRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Filters     FilterConfig `json:"filters"`
	Tags        []string     `json:"tags,omitempty"`
}

// HistoryParams represents history query parameters
type HistoryParams struct {
	Symbol    string `json:"symbol,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
}

// WSMessage represents an outgoing WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSResponse represents an incoming WebSocket message
type WSResponse struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
}