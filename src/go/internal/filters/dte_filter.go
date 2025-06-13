package filters

import (
	"github.com/ibkr-automation/scanner/pkg/models"
)

// DTEFilter filters options by days to expiration
type DTEFilter struct {
	minDTE int
	maxDTE int
}

// NewDTEFilter creates a new DTE filter
func NewDTEFilter(params map[string]interface{}) *DTEFilter {
	filter := &DTEFilter{
		minDTE: 0,
		maxDTE: 365,
	}
	
	if min, ok := params["min"].(float64); ok {
		filter.minDTE = int(min)
	}
	if max, ok := params["max"].(float64); ok {
		filter.maxDTE = int(max)
	}
	
	return filter
}

// Apply checks if option passes DTE filter
func (f *DTEFilter) Apply(option *models.Option) bool {
	return option.DTE >= f.minDTE && option.DTE <= f.maxDTE
}

// Name returns the filter name
func (f *DTEFilter) Name() string {
	return "dte"
}