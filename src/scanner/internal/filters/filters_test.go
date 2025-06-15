package filters

import (
	"testing"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/models"
)

// Helper function to create test contracts
func createTestContracts() []models.OptionContract {
	return []models.OptionContract{
		{
			Symbol:       "SPY240315C00500",
			Strike:       500,
			Expiry:       time.Now().AddDate(0, 0, 30),
			OptionType:   "CALL",
			Delta:        0.30,
			Theta:        -0.05,
			Vega:         0.15,
			IV:           0.25,
			IVPercentile: 75,
			DTE:          30,
			Volume:       1000,
			OpenInterest: 5000,
			BidAskSpread: 0.05,
		},
		{
			Symbol:       "SPY240315C00510",
			Strike:       510,
			Expiry:       time.Now().AddDate(0, 0, 30),
			OptionType:   "CALL",
			Delta:        0.20,
			Theta:        -0.03,
			Vega:         0.10,
			IV:           0.30,
			IVPercentile: 85,
			DTE:          30,
			Volume:       500,
			OpenInterest: 2000,
			BidAskSpread: 0.15,
		},
		{
			Symbol:       "SPY240415C00500",
			Strike:       500,
			Expiry:       time.Now().AddDate(0, 0, 60),
			OptionType:   "CALL",
			Delta:        0.40,
			Theta:        -0.02,
			Vega:         0.20,
			IV:           0.20,
			IVPercentile: 50,
			DTE:          60,
			Volume:       50,
			OpenInterest: 100,
			BidAskSpread: 0.25,
		},
	}
}

func TestDeltaFilter(t *testing.T) {
	contracts := createTestContracts()
	
	tests := []struct {
		name     string
		filter   *DeltaFilter
		expected int
	}{
		{
			name: "Delta range 0.25-0.35",
			filter: &DeltaFilter{
				MinDelta: 0.25,
				MaxDelta: 0.35,
			},
			expected: 1, // Only first contract
		},
		{
			name: "Delta range 0.15-0.45",
			filter: &DeltaFilter{
				MinDelta: 0.15,
				MaxDelta: 0.45,
			},
			expected: 3, // All contracts
		},
		{
			name:     "Nil filter",
			filter:   nil,
			expected: 3, // All contracts
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Apply(contracts)
			if len(result) != tt.expected {
				t.Errorf("expected %d contracts, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestDTEFilter(t *testing.T) {
	contracts := createTestContracts()
	
	tests := []struct {
		name     string
		filter   *DTEFilter
		expected int
	}{
		{
			name: "DTE 25-35",
			filter: &DTEFilter{
				MinDTE: 25,
				MaxDTE: 35,
			},
			expected: 2, // First two contracts
		},
		{
			name: "DTE 50-70",
			filter: &DTEFilter{
				MinDTE: 50,
				MaxDTE: 70,
			},
			expected: 1, // Last contract
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Apply(contracts)
			if len(result) != tt.expected {
				t.Errorf("expected %d contracts, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestLiquidityFilter(t *testing.T) {
	contracts := createTestContracts()
	
	tests := []struct {
		name     string
		filter   *LiquidityFilter
		expected int
	}{
		{
			name: "High liquidity",
			filter: &LiquidityFilter{
				MinVolume:       500,
				MinOpenInterest: 2000,
				MaxBidAskSpread: 0.10,
			},
			expected: 1, // Only first contract
		},
		{
			name: "Low liquidity threshold",
			filter: &LiquidityFilter{
				MinVolume:       50,
				MinOpenInterest: 100,
				MaxBidAskSpread: 0.30,
			},
			expected: 3, // All contracts
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Apply(contracts)
			if len(result) != tt.expected {
				t.Errorf("expected %d contracts, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestFilterChain(t *testing.T) {
	contracts := createTestContracts()
	
	config := FilterConfig{
		Delta: &DeltaFilter{
			MinDelta: 0.25,
			MaxDelta: 0.35,
		},
		DTE: &DTEFilter{
			MinDTE: 25,
			MaxDTE: 35,
		},
		Liquidity: &LiquidityFilter{
			MinVolume:       100,
			MinOpenInterest: 1000,
			MaxBidAskSpread: 0.10,
		},
	}
	
	chain := NewFilterChain(config)
	result := chain.ApplyToContracts(contracts)
	
	// Only the first contract should pass all filters
	if len(result) != 1 {
		t.Errorf("expected 1 contract after filter chain, got %d", len(result))
	}
	
	if result[0].Symbol != "SPY240315C00500" {
		t.Errorf("expected SPY240315C00500, got %s", result[0].Symbol)
	}
}