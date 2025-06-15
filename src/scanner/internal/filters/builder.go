package filters

import (
	"encoding/json"
	"fmt"
)

// FilterBuilder provides a fluent interface for building filter chains
type FilterBuilder struct {
	config      FilterConfig
	advanced    *AdvancedFilterChain
	errors      []error
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		config: FilterConfig{},
		errors: make([]error, 0),
	}
}

// WithDeltaFilter adds a delta filter
func (fb *FilterBuilder) WithDeltaFilter(minDelta, maxDelta float64) *FilterBuilder {
	fb.config.Delta = &DeltaFilter{
		MinDelta: minDelta,
		MaxDelta: maxDelta,
	}
	return fb
}

// WithDTEFilter adds a DTE filter
func (fb *FilterBuilder) WithDTEFilter(minDTE, maxDTE int) *FilterBuilder {
	fb.config.DTE = &DTEFilter{
		MinDTE: minDTE,
		MaxDTE: maxDTE,
	}
	return fb
}

// WithLiquidityFilter adds a liquidity filter
func (fb *FilterBuilder) WithLiquidityFilter(minOI, minVolume int) *FilterBuilder {
	fb.config.Liquidity = &LiquidityFilter{
		MinOpenInterest: int64(minOI),
		MinVolume:       int64(minVolume),
	}
	return fb
}

// WithThetaFilter adds a theta filter
func (fb *FilterBuilder) WithThetaFilter(minTheta, maxTheta float64) *FilterBuilder {
	fb.config.Theta = &ThetaFilter{
		MinTheta: minTheta,
		MaxTheta: maxTheta,
	}
	return fb
}

// WithVegaFilter adds a vega filter
func (fb *FilterBuilder) WithVegaFilter(minVega, maxVega float64) *FilterBuilder {
	fb.config.Vega = &VegaFilter{
		MinVega: minVega,
		MaxVega: maxVega,
	}
	return fb
}

// WithIVFilter adds an IV filter
func (fb *FilterBuilder) WithIVFilter(minIV, maxIV float64) *FilterBuilder {
	fb.config.IV = &IVFilter{
		MinIV: minIV,
		MaxIV: maxIV,
	}
	return fb
}

// WithIVPercentileFilter adds an IV percentile filter
func (fb *FilterBuilder) WithIVPercentileFilter(minPercentile, maxPercentile float64) *FilterBuilder {
	fb.config.IVPercentile = &IVPercentileFilter{
		MinPercentile: minPercentile,
		MaxPercentile: maxPercentile,
	}
	return fb
}

// WithSpreadWidthFilter adds a spread width filter
func (fb *FilterBuilder) WithSpreadWidthFilter(minWidth, maxWidth float64) *FilterBuilder {
	fb.config.SpreadWidth = &SpreadWidthFilter{
		MinWidth: minWidth,
		MaxWidth: maxWidth,
	}
	return fb
}

// WithPoPFilter adds a probability of profit filter
func (fb *FilterBuilder) WithPoPFilter(minPoP, maxPoP float64) *FilterBuilder {
	fb.config.ProbOfProfit = &PoPFilter{
		MinPoP: minPoP,
		MaxPoP: maxPoP,
	}
	return fb
}

// WithMaxPositions sets maximum positions limit
func (fb *FilterBuilder) WithMaxPositions(max int) *FilterBuilder {
	fb.config.MaxPositions = max
	return fb
}

// WithRiskLimit sets risk limit
func (fb *FilterBuilder) WithRiskLimit(limit float64) *FilterBuilder {
	fb.config.RiskLimit = limit
	return fb
}

// Build creates the filter chain
func (fb *FilterBuilder) Build() (*AdvancedFilterChain, error) {
	if len(fb.errors) > 0 {
		return nil, fmt.Errorf("filter builder has %d errors: %v", len(fb.errors), fb.errors[0])
	}
	
	// Create advanced filter chain
	chain := NewAdvancedFilterChain(fb.config, false, false)
	
	// Validate all filters
	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("filter validation failed: %w", err)
	}
	
	return chain, nil
}

// BuildWithCache creates a filter chain with caching enabled
func (fb *FilterBuilder) BuildWithCache() (*AdvancedFilterChain, error) {
	if len(fb.errors) > 0 {
		return nil, fmt.Errorf("filter builder has %d errors: %v", len(fb.errors), fb.errors[0])
	}
	
	// Create advanced filter chain with cache
	chain := NewAdvancedFilterChain(fb.config, true, false)
	
	// Validate all filters
	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("filter validation failed: %w", err)
	}
	
	return chain, nil
}

// BuildParallel creates a filter chain with parallel execution
func (fb *FilterBuilder) BuildParallel() (*AdvancedFilterChain, error) {
	if len(fb.errors) > 0 {
		return nil, fmt.Errorf("filter builder has %d errors: %v", len(fb.errors), fb.errors[0])
	}
	
	// Create advanced filter chain with parallel execution
	chain := NewAdvancedFilterChain(fb.config, true, true)
	
	// Validate all filters
	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("filter validation failed: %w", err)
	}
	
	return chain, nil
}

// FromJSON loads configuration from JSON
func (fb *FilterBuilder) FromJSON(jsonData []byte) *FilterBuilder {
	if err := json.Unmarshal(jsonData, &fb.config); err != nil {
		fb.errors = append(fb.errors, fmt.Errorf("failed to parse JSON: %w", err))
	}
	return fb
}

// ToJSON exports configuration to JSON
func (fb *FilterBuilder) ToJSON() ([]byte, error) {
	return json.MarshalIndent(fb.config, "", "  ")
}

// Reset clears the builder
func (fb *FilterBuilder) Reset() *FilterBuilder {
	fb.config = FilterConfig{}
	fb.errors = make([]error, 0)
	return fb
}

// FilterPresets provides common filter configurations
type FilterPresets struct{}

// NewFilterPresets creates filter presets
func NewFilterPresets() *FilterPresets {
	return &FilterPresets{}
}

// Conservative returns a conservative filter configuration
func (fp *FilterPresets) Conservative() *FilterBuilder {
	return NewFilterBuilder().
		WithDeltaFilter(0.15, 0.30).
		WithDTEFilter(30, 60).
		WithLiquidityFilter(100, 50).
		WithIVPercentileFilter(30, 70).
		WithPoPFilter(0.70, 0.90).
		WithMaxPositions(5).
		WithRiskLimit(5000)
}

// Moderate returns a moderate filter configuration
func (fp *FilterPresets) Moderate() *FilterBuilder {
	return NewFilterBuilder().
		WithDeltaFilter(0.20, 0.40).
		WithDTEFilter(20, 45).
		WithLiquidityFilter(50, 25).
		WithIVPercentileFilter(40, 80).
		WithPoPFilter(0.60, 0.85).
		WithMaxPositions(10).
		WithRiskLimit(10000)
}

// Aggressive returns an aggressive filter configuration
func (fp *FilterPresets) Aggressive() *FilterBuilder {
	return NewFilterBuilder().
		WithDeltaFilter(0.25, 0.50).
		WithDTEFilter(7, 30).
		WithLiquidityFilter(25, 10).
		WithIVPercentileFilter(50, 90).
		WithPoPFilter(0.50, 0.80).
		WithMaxPositions(20).
		WithRiskLimit(20000)
}

// HighIV returns a configuration for high IV environments
func (fp *FilterPresets) HighIV() *FilterBuilder {
	return NewFilterBuilder().
		WithDeltaFilter(0.10, 0.25).
		WithDTEFilter(30, 60).
		WithLiquidityFilter(100, 50).
		WithIVFilter(0.30, 1.0).
		WithIVPercentileFilter(70, 100).
		WithVegaFilter(0.05, 0.20).
		WithMaxPositions(8).
		WithRiskLimit(8000)
}

// ThetaHarvesting returns a configuration for theta harvesting
func (fp *FilterPresets) ThetaHarvesting() *FilterBuilder {
	return NewFilterBuilder().
		WithDeltaFilter(0.20, 0.35).
		WithDTEFilter(15, 45).
		WithLiquidityFilter(50, 25).
		WithThetaFilter(0.02, 0.10).
		WithPoPFilter(0.65, 0.85).
		WithMaxPositions(15).
		WithRiskLimit(15000)
}