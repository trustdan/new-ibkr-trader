package filters

import (
	"github.com/ibkr-automation/scanner/pkg/models"
)

// Filter interface for all option filters
type Filter interface {
	Apply(option *models.Option) bool
	Name() string
}

// FilterChain represents a chain of filters
type FilterChain struct {
	filters []Filter
}

// BuildFilterChain creates a filter chain from configuration
func BuildFilterChain(configs []models.FilterConfig) *FilterChain {
	chain := &FilterChain{
		filters: make([]Filter, 0, len(configs)),
	}
	
	for _, config := range configs {
		filter := createFilter(config)
		if filter != nil {
			chain.filters = append(chain.filters, filter)
		}
	}
	
	return chain
}

// Apply runs all filters on an option
func (fc *FilterChain) Apply(option *models.Option) bool {
	// All filters must pass
	for _, filter := range fc.filters {
		if !filter.Apply(option) {
			return false
		}
	}
	return true
}

// createFilter creates a filter based on configuration
func createFilter(config models.FilterConfig) Filter {
	switch config.Type {
	case "delta":
		return NewDeltaFilter(config.Params)
	case "dte":
		return NewDTEFilter(config.Params)
	case "volume":
		return NewVolumeFilter(config.Params)
	case "open_interest":
		return NewOpenInterestFilter(config.Params)
	case "iv_percentile":
		return NewIVPercentileFilter(config.Params)
	case "bid_ask_spread":
		return NewBidAskSpreadFilter(config.Params)
	default:
		return nil
	}
}