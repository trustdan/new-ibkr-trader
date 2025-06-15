package filters

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/ibkr-trader/scanner/internal/models"
)

// Test data generators
func createTestContracts(count int) []models.OptionContract {
	contracts := make([]models.OptionContract, count)
	for i := 0; i < count; i++ {
		contracts[i] = models.OptionContract{
			Symbol:       "TEST",
			Strike:       100 + float64(i),
			DTE:         30 + i,
			Delta:       0.20 + float64(i)*0.01,
			Theta:       -0.05 - float64(i)*0.001,
			Vega:        0.10 + float64(i)*0.005,
			Gamma:       0.02 + float64(i)*0.001,
			IV:          0.25 + float64(i)*0.01,
			IVPercentile: 50 + float64(i),
			Bid:         2.00 + float64(i)*0.10,
			Ask:         2.10 + float64(i)*0.10,
			Volume:      int64(1000 - i*10),
			OpenInterest: int64(5000 - i*50),
			Score:       0.75 + float64(i)*0.01,
		}
	}
	return contracts
}

func createTestSpreads(count int) []models.VerticalSpread {
	spreads := make([]models.VerticalSpread, count)
	contracts := createTestContracts(count * 2)
	
	for i := 0; i < count; i++ {
		spreads[i] = models.VerticalSpread{
			Symbol:          "TEST",
			ShortLeg:        contracts[i*2],
			LongLeg:         contracts[i*2+1],
			Credit:          1.00 + float64(i)*0.10,
			ProbOfProfit:    0.70 + float64(i)*0.01,
			UnderlyingPrice: 100.0,
			Score:           0.80 + float64(i)*0.01,
		}
	}
	return spreads
}

// Test AdvancedFilterChain
func TestAdvancedFilterChain(t *testing.T) {
	t.Run("Basic Sequential Filtering", func(t *testing.T) {
		config := FilterConfig{
			Delta: &DeltaFilter{MinDelta: 0.25, MaxDelta: 0.35},
			DTE:   &DTEFilter{MinDTE: 35, MaxDTE: 45},
		}
		
		chain := NewAdvancedFilterChain(config, false, false)
		contracts := createTestContracts(20)
		
		filtered := chain.ApplyToContracts(contracts)
		
		// Verify filters were applied
		assert.Less(t, len(filtered), len(contracts))
		
		// Check all results meet criteria
		for _, c := range filtered {
			assert.GreaterOrEqual(t, c.Delta, 0.25)
			assert.LessOrEqual(t, c.Delta, 0.35)
			assert.GreaterOrEqual(t, c.DTE, 35)
			assert.LessOrEqual(t, c.DTE, 45)
		}
		
		// Check stats were recorded
		stats := chain.GetStats()
		assert.Equal(t, 2, len(stats))
		assert.Greater(t, stats["DeltaFilter"].ExecutionCount, int64(0))
		assert.Greater(t, stats["DTEFilter"].ExecutionCount, int64(0))
	})
	
	t.Run("Parallel Filtering", func(t *testing.T) {
		config := FilterConfig{
			Delta:     &DeltaFilter{MinDelta: 0.20, MaxDelta: 0.40},
			Liquidity: &LiquidityFilter{MinVolume: 800, MinOpenInterest: 4000},
		}
		
		chain := NewAdvancedFilterChain(config, false, true)
		contracts := createTestContracts(50)
		
		filtered := chain.ApplyToContracts(contracts)
		
		// Verify both filters were applied
		for _, c := range filtered {
			assert.GreaterOrEqual(t, c.Delta, 0.20)
			assert.LessOrEqual(t, c.Delta, 0.40)
			assert.GreaterOrEqual(t, c.Volume, int64(800))
			assert.GreaterOrEqual(t, c.OpenInterest, int64(4000))
		}
	})
	
	t.Run("Cached Filtering", func(t *testing.T) {
		config := FilterConfig{
			Delta: &DeltaFilter{MinDelta: 0.25, MaxDelta: 0.35},
		}
		
		chain := NewAdvancedFilterChain(config, true, false)
		contracts := createTestContracts(20)
		
		// First call - cache miss
		filtered1 := chain.ApplyToContracts(contracts)
		
		// Second call - cache hit
		filtered2 := chain.ApplyToContracts(contracts)
		
		// Results should be identical
		assert.Equal(t, len(filtered1), len(filtered2))
		
		// Check cache stats
		hits, misses, _, hitRate := chain.cache.GetStats()
		assert.Equal(t, int64(1), hits)
		assert.Equal(t, int64(1), misses)
		assert.Equal(t, 0.5, hitRate)
	})
}

// Test Spread Filters
func TestSpreadFilters(t *testing.T) {
	t.Run("RiskRewardFilter", func(t *testing.T) {
		filter := &RiskRewardFilter{MinRatio: 0.3, MaxRatio: 0.5}
		spread := models.VerticalSpread{
			ShortLeg: models.OptionContract{Strike: 105},
			LongLeg:  models.OptionContract{Strike: 100},
			Credit:   2.0,
		}
		
		// Risk/Reward = 2.0 / (5 - 2.0) = 2.0 / 3.0 = 0.67
		assert.False(t, filter.ApplyToSpread(spread))
		
		// Adjust for passing case
		spread.Credit = 1.5 // Risk/Reward = 1.5 / 3.5 = 0.43
		assert.True(t, filter.ApplyToSpread(spread))
	})
	
	t.Run("BreakEvenFilter", func(t *testing.T) {
		filter := &BreakEvenFilter{MinDistance: 2.0, MaxDistance: 5.0}
		spread := models.VerticalSpread{
			ShortLeg:        models.OptionContract{Strike: 105},
			Credit:          2.0,
			UnderlyingPrice: 100.0,
		}
		
		// Breakeven = 105 - 2 = 103
		// Distance = |103 - 100| / 100 * 100 = 3%
		assert.True(t, filter.ApplyToSpread(spread))
	})
	
	t.Run("ExpectedValueFilter", func(t *testing.T) {
		filter := &ExpectedValueFilter{MinEV: 0.5}
		spread := models.VerticalSpread{
			ShortLeg:     models.OptionContract{Strike: 105},
			LongLeg:      models.OptionContract{Strike: 100},
			Credit:       2.0,
			ProbOfProfit: 0.7,
		}
		
		// EV = (2.0 * 0.7) - (3.0 * 0.3) = 1.4 - 0.9 = 0.5
		assert.True(t, filter.ApplyToSpread(spread))
	})
	
	t.Run("CombinedGreeksFilter", func(t *testing.T) {
		filter := &CombinedGreeksFilter{
			MaxGammaRisk:  0.05,
			MaxVegaRisk:   0.30,
			MinThetaDecay: 0.02,
		}
		
		spread := models.VerticalSpread{
			ShortLeg: models.OptionContract{
				Gamma: -0.03,
				Vega:  -0.20,
				Theta: 0.05,
			},
			LongLeg: models.OptionContract{
				Gamma: 0.02,
				Vega:  0.15,
				Theta: -0.02,
			},
		}
		
		// Net Gamma = |-0.03 + 0.02| = 0.01 ✓
		// Net Vega = |-0.20 + 0.15| = 0.05 ✓
		// Net Theta = 0.05 - 0.02 = 0.03 ✓
		assert.True(t, filter.ApplyToSpread(spread))
	})
}

// Test Combined Filters
func TestCombinedFilters(t *testing.T) {
	t.Run("CorrelationFilter", func(t *testing.T) {
		filter := &CorrelationFilter{
			MaxCorrelation: 0.5,
			SymbolGroups: map[string][]string{
				"tech": {"AAPL", "MSFT", "GOOGL"},
				"finance": {"JPM", "BAC", "GS"},
			},
		}
		
		contracts := []models.OptionContract{
			{Symbol: "AAPL"},
			{Symbol: "MSFT"},
			{Symbol: "GOOGL"},
			{Symbol: "JPM"},
		}
		
		spreads := []models.VerticalSpread{
			{Symbol: "AAPL"},
		}
		
		filteredContracts, filteredSpreads := filter.ApplyToCombined(contracts, spreads)
		
		// Should filter out MSFT and GOOGL due to correlation limit
		assert.Equal(t, 2, len(filteredContracts)) // Only JPM and one tech stock
		assert.Equal(t, 1, len(filteredSpreads))   // AAPL spread remains
	})
	
	t.Run("RankingFilter", func(t *testing.T) {
		filter := &RankingFilter{
			MaxContracts:   3,
			MaxSpreads:     2,
			ScoreThreshold: 0.75,
		}
		
		contracts := createTestContracts(10)
		spreads := createTestSpreads(5)
		
		filteredContracts, filteredSpreads := filter.ApplyToCombined(contracts, spreads)
		
		// Should limit to top 3 contracts and 2 spreads
		assert.LessOrEqual(t, len(filteredContracts), 3)
		assert.LessOrEqual(t, len(filteredSpreads), 2)
		
		// All should meet score threshold
		for _, c := range filteredContracts {
			assert.GreaterOrEqual(t, c.Score, 0.75)
		}
		for _, s := range filteredSpreads {
			assert.GreaterOrEqual(t, s.Score, 0.75)
		}
	})
}

// Test Filter Builder
func TestFilterBuilder(t *testing.T) {
	t.Run("Basic Builder", func(t *testing.T) {
		chain, err := NewFilterBuilder().
			WithDeltaFilter(0.20, 0.40).
			WithDTEFilter(30, 60).
			WithLiquidityFilter(100, 50).
			Build()
		
		assert.NoError(t, err)
		assert.NotNil(t, chain)
		
		// Test the built chain
		contracts := createTestContracts(20)
		filtered := chain.ApplyToContracts(contracts)
		assert.Less(t, len(filtered), len(contracts))
	})
	
	t.Run("Preset Configurations", func(t *testing.T) {
		presets := NewFilterPresets()
		
		// Test conservative preset
		conservative, err := presets.Conservative().Build()
		assert.NoError(t, err)
		assert.NotNil(t, conservative)
		
		// Test aggressive preset
		aggressive, err := presets.Aggressive().Build()
		assert.NoError(t, err)
		assert.NotNil(t, aggressive)
		
		// Test theta harvesting preset
		theta, err := presets.ThetaHarvesting().Build()
		assert.NoError(t, err)
		assert.NotNil(t, theta)
	})
	
	t.Run("JSON Configuration", func(t *testing.T) {
		jsonConfig := `{
			"delta": {"min_delta": 0.25, "max_delta": 0.35},
			"dte": {"min_dte": 30, "max_dte": 45},
			"max_positions": 10,
			"risk_limit": 5000
		}`
		
		chain, err := NewFilterBuilder().
			FromJSON([]byte(jsonConfig)).
			Build()
		
		assert.NoError(t, err)
		assert.NotNil(t, chain)
	})
}

// Benchmark tests
func BenchmarkSequentialFiltering(b *testing.B) {
	config := FilterConfig{
		Delta:        &DeltaFilter{MinDelta: 0.20, MaxDelta: 0.40},
		DTE:          &DTEFilter{MinDTE: 30, MaxDTE: 60},
		Liquidity:    &LiquidityFilter{MinVolume: 100, MinOpenInterest: 500},
		IVPercentile: &IVPercentileFilter{MinPercentile: 30, MaxPercentile: 70},
	}
	
	chain := NewAdvancedFilterChain(config, false, false)
	contracts := createTestContracts(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chain.ApplyToContracts(contracts)
	}
}

func BenchmarkParallelFiltering(b *testing.B) {
	config := FilterConfig{
		Delta:        &DeltaFilter{MinDelta: 0.20, MaxDelta: 0.40},
		DTE:          &DTEFilter{MinDTE: 30, MaxDTE: 60},
		Liquidity:    &LiquidityFilter{MinVolume: 100, MinOpenInterest: 500},
		IVPercentile: &IVPercentileFilter{MinPercentile: 30, MaxPercentile: 70},
	}
	
	chain := NewAdvancedFilterChain(config, false, true)
	contracts := createTestContracts(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chain.ApplyToContracts(contracts)
	}
}

func BenchmarkCachedFiltering(b *testing.B) {
	config := FilterConfig{
		Delta: &DeltaFilter{MinDelta: 0.20, MaxDelta: 0.40},
		DTE:   &DTEFilter{MinDTE: 30, MaxDTE: 60},
	}
	
	chain := NewAdvancedFilterChain(config, true, false)
	contracts := createTestContracts(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chain.ApplyToContracts(contracts)
	}
}