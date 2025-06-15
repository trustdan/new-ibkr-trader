package filters

import (
	"github.com/ibkr-trader/scanner/internal/models"
)

// ThetaFilter filters by theta (time decay)
type ThetaFilter struct {
	MinTheta float64 `json:"min_theta"`
	MaxTheta float64 `json:"max_theta"`
}

func (f *ThetaFilter) Name() string { return "ThetaFilter" }
func (f *ThetaFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	filtered := make([]models.OptionContract, 0)
	for _, c := range contracts {
		if c.Theta >= f.MinTheta && c.Theta <= f.MaxTheta {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
func (f *ThetaFilter) Validate() error { return nil }

// VegaFilter filters by vega (volatility sensitivity)
type VegaFilter struct {
	MinVega float64 `json:"min_vega"`
	MaxVega float64 `json:"max_vega"`
}

func (f *VegaFilter) Name() string { return "VegaFilter" }
func (f *VegaFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	filtered := make([]models.OptionContract, 0)
	for _, c := range contracts {
		if c.Vega >= f.MinVega && c.Vega <= f.MaxVega {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
func (f *VegaFilter) Validate() error { return nil }

// IVFilter filters by implied volatility
type IVFilter struct {
	MinIV float64 `json:"min_iv"`
	MaxIV float64 `json:"max_iv"`
}

func (f *IVFilter) Name() string { return "IVFilter" }
func (f *IVFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	filtered := make([]models.OptionContract, 0)
	for _, c := range contracts {
		if c.IV >= f.MinIV && c.IV <= f.MaxIV {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
func (f *IVFilter) Validate() error { return nil }

// IVPercentileFilter filters by IV percentile
type IVPercentileFilter struct {
	MinPercentile float64 `json:"min_percentile"`
	MaxPercentile float64 `json:"max_percentile"`
}

func (f *IVPercentileFilter) Name() string { return "IVPercentileFilter" }
func (f *IVPercentileFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	filtered := make([]models.OptionContract, 0)
	for _, c := range contracts {
		if c.IVPercentile >= f.MinPercentile && c.IVPercentile <= f.MaxPercentile {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
func (f *IVPercentileFilter) Validate() error { return nil }

// SpreadWidthFilter filters spreads by width
type SpreadWidthFilter struct {
	MinWidth float64 `json:"min_width"`
	MaxWidth float64 `json:"max_width"`
}

func (f *SpreadWidthFilter) Name() string { return "SpreadWidthFilter" }
func (f *SpreadWidthFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	width := spread.ShortLeg.Strike - spread.LongLeg.Strike
	return width >= f.MinWidth && width <= f.MaxWidth
}
func (f *SpreadWidthFilter) Validate() error { return nil }

// PoPFilter filters by probability of profit
type PoPFilter struct {
	MinPoP float64 `json:"min_pop"`
	MaxPoP float64 `json:"max_pop"`
}

func (f *PoPFilter) Name() string { return "PoPFilter" }
func (f *PoPFilter) ApplyToSpread(spread models.VerticalSpread) bool {
	if f == nil {
		return true
	}
	return spread.ProbOfProfit >= f.MinPoP && spread.ProbOfProfit <= f.MaxPoP
}
func (f *PoPFilter) Validate() error { return nil }