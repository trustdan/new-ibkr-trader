package benchmark

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/ibkr-automation/scanner/internal/filters"
	"github.com/ibkr-automation/scanner/internal/scoring"
	"github.com/ibkr-automation/scanner/internal/greeks"
	"github.com/ibkr-automation/scanner/pkg/models"
)

// BenchmarkFilterChain tests the performance of filter chains
func BenchmarkFilterChain(b *testing.B) {
	// Create test options
	options := generateTestOptions(1000)
	
	// Create filter configs
	filterConfigs := []models.FilterConfig{
		{
			Type: "delta",
			Params: map[string]interface{}{
				"min": 0.25,
				"max": 0.35,
			},
		},
		{
			Type: "dte",
			Params: map[string]interface{}{
				"min": 30,
				"max": 60,
			},
		},
		{
			Type: "liquidity",
			Params: map[string]interface{}{
				"min_volume": 100,
				"min_open_interest": 500,
			},
		},
	}
	
	// Build filter chain
	chain := filters.BuildFilterChain(filterConfigs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filtered := 0
		for _, opt := range options {
			if chain.Apply(&opt) {
				filtered++
			}
		}
	}
}

// BenchmarkScoring tests the performance of the scoring algorithm
func BenchmarkScoring(b *testing.B) {
	scorer := scoring.NewScorer(scoring.DefaultScoringConfig())
	spread := generateTestSpread()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.ScoreSpread(spread)
	}
}

// BenchmarkGreeksAnalysis tests the performance of Greeks analysis
func BenchmarkGreeksAnalysis(b *testing.B) {
	analyzer := greeks.NewGreeksAnalyzer(greeks.DefaultGreeksConfig())
	spread := generateTestSpread()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.AnalyzeSpread(spread)
	}
}

// BenchmarkConcurrentScanning tests concurrent scanning performance
func BenchmarkConcurrentScanning(b *testing.B) {
	options := generateTestOptions(10000)
	
	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			processOptionsSequential(options)
		}
	})
	
	b.Run("Concurrent-4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			processOptionsConcurrent(options, 4)
		}
	})
	
	b.Run("Concurrent-8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			processOptionsConcurrent(options, 8)
		}
	})
	
	b.Run("Concurrent-16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			processOptionsConcurrent(options, 16)
		}
	})
}

// BenchmarkCachePerformance tests cache hit/miss performance
func BenchmarkCachePerformance(b *testing.B) {
	c := cache.New(5*time.Minute, 10*time.Minute)
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		c.Set(key, generateTestOptions(100), cache.DefaultExpiration)
	}
	
	b.Run("CacheHit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", rand.Intn(1000))
			_, _ = c.Get(key)
		}
	})
	
	b.Run("CacheMiss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("miss_key_%d", i)
			_, _ = c.Get(key)
		}
	})
}

// BenchmarkSpreadGeneration tests spread generation performance
func BenchmarkSpreadGeneration(b *testing.B) {
	options := generateTestOptions(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSpreadsFromOptions(options)
	}
}

// Helper functions

func generateTestOptions(count int) []models.Option {
	options := make([]models.Option, count)
	
	for i := 0; i < count; i++ {
		strike := 100.0 + float64(i)
		options[i] = models.Option{
			Symbol:       "TEST",
			Strike:       strike,
			Expiration:   time.Now().AddDate(0, 0, 45),
			OptionType:   "call",
			DTE:          45,
			Bid:          rand.Float64() * 10,
			Ask:          rand.Float64() * 10 + 0.1,
			Volume:       rand.Int63n(10000),
			OpenInterest: rand.Int63n(50000),
			Delta:        rand.Float64() * 0.5,
			Gamma:        rand.Float64() * 0.1,
			Theta:        -rand.Float64() * 0.1,
			Vega:         rand.Float64() * 0.2,
			IV:           0.2 + rand.Float64() * 0.3,
		}
	}
	
	return options
}

func generateTestSpread() *models.VerticalSpread {
	longLeg := models.OptionContract{
		Symbol: "TEST",
		Strike: 100,
		Delta:  0.35,
		Theta:  -0.08,
		Vega:   0.15,
		IV:     0.25,
		Volume: 1000,
		OpenInterest: 5000,
		Bid:    4.50,
		Ask:    4.60,
	}
	
	shortLeg := models.OptionContract{
		Symbol: "TEST",
		Strike: 105,
		Delta:  0.25,
		Theta:  -0.06,
		Vega:   0.12,
		IV:     0.23,
		Volume: 800,
		OpenInterest: 4000,
		Bid:    2.20,
		Ask:    2.30,
	}
	
	return &models.VerticalSpread{
		LongLeg:           longLeg,
		ShortLeg:          shortLeg,
		NetDebit:          2.40,
		MaxProfit:         2.60,
		MaxLoss:           2.40,
		Breakeven:         102.40,
		ProbabilityProfit: 0.65,
	}
}

func processOptionsSequential(options []models.Option) int {
	count := 0
	for _, opt := range options {
		if opt.Delta > 0.25 && opt.Delta < 0.35 {
			count++
		}
	}
	return count
}

func processOptionsConcurrent(options []models.Option, workers int) int {
	ch := make(chan int, workers)
	chunkSize := len(options) / workers
	
	for i := 0; i < workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == workers-1 {
			end = len(options)
		}
		
		go func(opts []models.Option) {
			count := 0
			for _, opt := range opts {
				if opt.Delta > 0.25 && opt.Delta < 0.35 {
					count++
				}
			}
			ch <- count
		}(options[start:end])
	}
	
	total := 0
	for i := 0; i < workers; i++ {
		total += <-ch
	}
	
	return total
}

func generateSpreadsFromOptions(options []models.Option) []models.VerticalSpread {
	var spreads []models.VerticalSpread
	
	for i := 0; i < len(options)-1; i++ {
		for j := i + 1; j < len(options) && j < i+10; j++ {
			spread := models.VerticalSpread{
				LongLeg: models.OptionContract{
					Strike: options[i].Strike,
					Delta:  options[i].Delta,
				},
				ShortLeg: models.OptionContract{
					Strike: options[j].Strike,
					Delta:  options[j].Delta,
				},
			}
			spreads = append(spreads, spread)
		}
	}
	
	return spreads
}