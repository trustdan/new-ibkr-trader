package filters

import (
	"math"
	"github.com/ibkr-automation/scanner/pkg/models"
)

// DeltaFilter filters options by delta range
type DeltaFilter struct {
	minDelta float64
	maxDelta float64
}

// NewDeltaFilter creates a new delta filter
func NewDeltaFilter(params map[string]interface{}) *DeltaFilter {
	filter := &DeltaFilter{
		minDelta: -1.0,
		maxDelta: 1.0,
	}
	
	if min, ok := params["min"].(float64); ok {
		filter.minDelta = min
	}
	if max, ok := params["max"].(float64); ok {
		filter.maxDelta = max
	}
	
	return filter
}

// Apply checks if option passes delta filter
func (f *DeltaFilter) Apply(option *models.Option) bool {
	absDelta := math.Abs(option.Delta)
	return absDelta >= math.Abs(f.minDelta) && absDelta <= math.Abs(f.maxDelta)
}

// Name returns the filter name
func (f *DeltaFilter) Name() string {
	return "delta"
}