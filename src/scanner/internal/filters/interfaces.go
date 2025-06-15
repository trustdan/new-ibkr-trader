package filters

import (
	"github.com/ibkr-trader/scanner/internal/models"
)

// Filter is the base interface for all option filters
type Filter interface {
	Name() string
	Apply(contracts []models.OptionContract) []models.OptionContract
	Validate() error
}

// SpreadFilter is specialized for filtering vertical spreads
type SpreadFilter interface {
	Name() string
	ApplyToSpread(spread models.VerticalSpread) bool
	Validate() error
}

// FilterConfig holds all filter configurations
type FilterConfig struct {
	// Basic filters
	Delta        *DeltaFilter        `json:"delta,omitempty"`
	DTE          *DTEFilter          `json:"dte,omitempty"`
	Liquidity    *LiquidityFilter    `json:"liquidity,omitempty"`
	
	// Greeks filters
	Theta        *ThetaFilter        `json:"theta,omitempty"`
	Vega         *VegaFilter         `json:"vega,omitempty"`
	
	// IV filters
	IV           *IVFilter           `json:"iv,omitempty"`
	IVPercentile *IVPercentileFilter `json:"iv_percentile,omitempty"`
	
	// Spread filters
	SpreadWidth  *SpreadWidthFilter  `json:"spread_width,omitempty"`
	ProbOfProfit *PoPFilter          `json:"prob_of_profit,omitempty"`
	
	// Position filters
	MaxPositions int                 `json:"max_positions,omitempty"`
	RiskLimit    float64             `json:"risk_limit,omitempty"`
}

// FilterChain manages a sequence of filters
type FilterChain struct {
	contractFilters []Filter
	spreadFilters   []SpreadFilter
}

// NewFilterChain creates a new filter chain from config
func NewFilterChain(config FilterConfig) *FilterChain {
	chain := &FilterChain{
		contractFilters: make([]Filter, 0),
		spreadFilters:   make([]SpreadFilter, 0),
	}
	
	// Add active filters to chain
	if config.Delta != nil {
		chain.contractFilters = append(chain.contractFilters, config.Delta)
	}
	if config.DTE != nil {
		chain.contractFilters = append(chain.contractFilters, config.DTE)
	}
	if config.Liquidity != nil {
		chain.contractFilters = append(chain.contractFilters, config.Liquidity)
	}
	
	// Add spread filters
	if config.SpreadWidth != nil {
		chain.spreadFilters = append(chain.spreadFilters, config.SpreadWidth)
	}
	if config.ProbOfProfit != nil {
		chain.spreadFilters = append(chain.spreadFilters, config.ProbOfProfit)
	}
	
	return chain
}

// ApplyToContracts applies all contract filters
func (fc *FilterChain) ApplyToContracts(contracts []models.OptionContract) []models.OptionContract {
	result := contracts
	for _, filter := range fc.contractFilters {
		result = filter.Apply(result)
	}
	return result
}

// ApplyToSpread checks if a spread passes all spread filters
func (fc *FilterChain) ApplyToSpread(spread models.VerticalSpread) bool {
	for _, filter := range fc.spreadFilters {
		if !filter.ApplyToSpread(spread) {
			return false
		}
	}
	return true
}