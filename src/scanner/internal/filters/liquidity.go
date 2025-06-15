package filters

import (
	"fmt"
	
	"github.com/ibkr-trader/scanner/internal/models"
)

// LiquidityFilter filters options by liquidity metrics
type LiquidityFilter struct {
	MinVolume       int64   `json:"min_volume"`
	MinOpenInterest int64   `json:"min_open_interest"`
	MaxBidAskSpread float64 `json:"max_bid_ask_spread"`
}

func (f *LiquidityFilter) Name() string {
	return "LiquidityFilter"
}

func (f *LiquidityFilter) Apply(contracts []models.OptionContract) []models.OptionContract {
	if f == nil {
		return contracts
	}
	
	filtered := make([]models.OptionContract, 0, len(contracts)/2)
	for _, contract := range contracts {
		// Check volume
		if f.MinVolume > 0 && contract.Volume < f.MinVolume {
			continue
		}
		
		// Check open interest
		if f.MinOpenInterest > 0 && contract.OpenInterest < f.MinOpenInterest {
			continue
		}
		
		// Check bid-ask spread
		if f.MaxBidAskSpread > 0 && contract.BidAskSpread > f.MaxBidAskSpread {
			continue
		}
		
		filtered = append(filtered, contract)
	}
	
	return filtered
}

func (f *LiquidityFilter) Validate() error {
	if f.MinVolume < 0 {
		return fmt.Errorf("min_volume cannot be negative")
	}
	if f.MinOpenInterest < 0 {
		return fmt.Errorf("min_open_interest cannot be negative")
	}
	if f.MaxBidAskSpread < 0 {
		return fmt.Errorf("max_bid_ask_spread cannot be negative")
	}
	return nil
}