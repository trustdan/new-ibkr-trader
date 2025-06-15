package filters

import (
	"fmt"
	"math"
	
	"github.com/ibkr-trader/scanner/internal/models"
)

// DeltaFilter filters options by delta range
type DeltaFilter struct {
	MinDelta float64 `json:"min_delta"`
	MaxDelta float64 `json:"max_delta"`
	Absolute bool    `json:"absolute"` // Use absolute value for puts
}

func (f *DeltaFilter) Name() string {
	return "DeltaFilter"
}

func (f *DeltaFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	
	filtered := make([]models.OptionContract, 0, len(contracts)/2)
	for _, contract := range contracts {
		delta := contract.Delta
		if f.Absolute {
			delta = math.Abs(delta)
		}
		
		if delta >= f.MinDelta && delta <= f.MaxDelta {
			filtered = append(filtered, contract)
		}
	}
	
	return filtered
}

func (f *DeltaFilter) Validate() error {
	if f.MinDelta < -1 || f.MinDelta > 1 {
		return fmt.Errorf("min_delta must be between -1 and 1")
	}
	if f.MaxDelta < -1 || f.MaxDelta > 1 {
		return fmt.Errorf("max_delta must be between -1 and 1")
	}
	if f.MinDelta > f.MaxDelta {
		return fmt.Errorf("min_delta cannot be greater than max_delta")
	}
	return nil
}