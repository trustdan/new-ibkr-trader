package filters

import (
	"github.com/ibkr-automation/scanner/pkg/models"
)

// VolumeFilter filters by minimum volume
type VolumeFilter struct {
	minVolume int64
}

func NewVolumeFilter(params map[string]interface{}) *VolumeFilter {
	filter := &VolumeFilter{minVolume: 0}
	if min, ok := params["min"].(float64); ok {
		filter.minVolume = int64(min)
	}
	return filter
}

func (f *VolumeFilter) Apply(option *models.Option) bool {
	return option.Volume >= f.minVolume
}

func (f *VolumeFilter) Name() string {
	return "volume"
}

// OpenInterestFilter filters by minimum open interest
type OpenInterestFilter struct {
	minOI int64
}

func NewOpenInterestFilter(params map[string]interface{}) *OpenInterestFilter {
	filter := &OpenInterestFilter{minOI: 0}
	if min, ok := params["min"].(float64); ok {
		filter.minOI = int64(min)
	}
	return filter
}

func (f *OpenInterestFilter) Apply(option *models.Option) bool {
	return option.OpenInterest >= f.minOI
}

func (f *OpenInterestFilter) Name() string {
	return "open_interest"
}

// IVPercentileFilter filters by IV percentile
type IVPercentileFilter struct {
	minIVP float64
	maxIVP float64
}

func NewIVPercentileFilter(params map[string]interface{}) *IVPercentileFilter {
	filter := &IVPercentileFilter{minIVP: 0, maxIVP: 100}
	if min, ok := params["min"].(float64); ok {
		filter.minIVP = min
	}
	if max, ok := params["max"].(float64); ok {
		filter.maxIVP = max
	}
	return filter
}

func (f *IVPercentileFilter) Apply(option *models.Option) bool {
	return option.IVPercentile >= f.minIVP && option.IVPercentile <= f.maxIVP
}

func (f *IVPercentileFilter) Name() string {
	return "iv_percentile"
}

// BidAskSpreadFilter filters by maximum bid-ask spread percentage
type BidAskSpreadFilter struct {
	maxSpreadPct float64
}

func NewBidAskSpreadFilter(params map[string]interface{}) *BidAskSpreadFilter {
	filter := &BidAskSpreadFilter{maxSpreadPct: 100}
	if max, ok := params["max"].(float64); ok {
		filter.maxSpreadPct = max
	}
	return filter
}

func (f *BidAskSpreadFilter) Apply(option *models.Option) bool {
	return option.BidAskSpreadPct <= f.maxSpreadPct
}

func (f *BidAskSpreadFilter) Name() string {
	return "bid_ask_spread"
}