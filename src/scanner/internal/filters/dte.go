package filters

import (
	"fmt"
	
	"github.com/ibkr-trader/scanner/internal/models"
)

// DTEFilter filters options by days to expiration
type DTEFilter struct {
	MinDTE int `json:"min_dte"`
	MaxDTE int `json:"max_dte"`
}

func (f *DTEFilter) Name() string {
	return "DTEFilter"
}

func (f *DTEFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	
	filtered := make([]models.OptionContract, 0, len(contracts)/2)
	for _, contract := range contracts {
		if contract.DTE >= f.MinDTE && contract.DTE <= f.MaxDTE {
			filtered = append(filtered, contract)
		}
	}
	
	return filtered
}

func (f *DTEFilter) Validate() error {
	if f.MinDTE < 0 {
		return fmt.Errorf("min_dte cannot be negative")
	}
	if f.MaxDTE < 0 {
		return fmt.Errorf("max_dte cannot be negative")
	}
	if f.MinDTE > f.MaxDTE {
		return fmt.Errorf("min_dte cannot be greater than max_dte")
	}
	return nil
}