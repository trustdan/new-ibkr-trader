package scanner

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ibkr-automation/scanner/pkg/models"
)

// generateVerticalSpreads creates all possible vertical spreads from filtered options
func (s *Scanner) generateVerticalSpreads(options []models.Option, req *models.ScanRequest) []*models.VerticalSpread {
	s.logger.Debug("Generating vertical spreads from options")
	
	// Group options by expiry
	expiryGroups := make(map[string][]models.Option)
	for _, opt := range options {
		expiryGroups[opt.Expiry] = append(expiryGroups[opt.Expiry], opt)
	}
	
	// Generate spreads for each expiry
	var allSpreads []*models.VerticalSpread
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	for expiry, opts := range expiryGroups {
		wg.Add(1)
		go func(exp string, options []models.Option) {
			defer wg.Done()
			
			spreads := s.generateSpreadsForExpiry(options, req)
			
			mu.Lock()
			allSpreads = append(allSpreads, spreads...)
			mu.Unlock()
		}(expiry, opts)
	}
	
	wg.Wait()
	
	// Score and analyze all spreads
	s.scoreAndAnalyzeSpreads(allSpreads)
	
	// Sort by score (descending)
	sort.Slice(allSpreads, func(i, j int) bool {
		return allSpreads[i].Score > allSpreads[j].Score
	})
	
	// Apply limit if specified
	if req.Limit > 0 && len(allSpreads) > req.Limit {
		allSpreads = allSpreads[:req.Limit]
	}
	
	s.logger.Infof("Generated %d vertical spreads", len(allSpreads))
	return allSpreads
}

// generateSpreadsForExpiry creates spreads for a single expiry
func (s *Scanner) generateSpreadsForExpiry(options []models.Option, req *models.ScanRequest) []*models.VerticalSpread {
	// Separate calls and puts
	var calls, puts []models.Option
	for _, opt := range options {
		if opt.Right == "C" {
			calls = append(calls, opt)
		} else {
			puts = append(puts, opt)
		}
	}
	
	// Sort by strike
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Strike < calls[j].Strike
	})
	sort.Slice(puts, func(i, j int) bool {
		return puts[i].Strike < puts[j].Strike
	})
	
	var spreads []*models.VerticalSpread
	
	// Generate call spreads (bull call spreads)
	spreads = append(spreads, s.generateCallSpreads(calls, req)...)
	
	// Generate put spreads (bull put spreads)
	spreads = append(spreads, s.generatePutSpreads(puts, req)...)
	
	return spreads
}

// generateCallSpreads creates bull call spreads
func (s *Scanner) generateCallSpreads(calls []models.Option, req *models.ScanRequest) []*models.VerticalSpread {
	var spreads []*models.VerticalSpread
	
	// For each potential long leg
	for i := 0; i < len(calls)-1; i++ {
		longLeg := calls[i]
		
		// Find suitable short legs (higher strikes)
		for j := i + 1; j < len(calls); j++ {
			shortLeg := calls[j]
			
			// Check spread width constraints
			width := shortLeg.Strike - longLeg.Strike
			if width < 1.0 || width > 10.0 {
				continue
			}
			
			// Create spread
			spread := s.createVerticalSpread(&longLeg, &shortLeg, "CALL")
			if spread != nil {
				spreads = append(spreads, spread)
			}
		}
	}
	
	return spreads
}

// generatePutSpreads creates bull put spreads
func (s *Scanner) generatePutSpreads(puts []models.Option, req *models.ScanRequest) []*models.VerticalSpread {
	var spreads []*models.VerticalSpread
	
	// For each potential short leg
	for i := 0; i < len(puts)-1; i++ {
		shortLeg := puts[i]
		
		// Find suitable long legs (higher strikes for puts)
		for j := i + 1; j < len(puts); j++ {
			longLeg := puts[j]
			
			// Check spread width constraints
			width := longLeg.Strike - shortLeg.Strike
			if width < 1.0 || width > 10.0 {
				continue
			}
			
			// Create spread
			spread := s.createVerticalSpread(&longLeg, &shortLeg, "PUT")
			if spread != nil {
				spreads = append(spreads, spread)
			}
		}
	}
	
	return spreads
}

// createVerticalSpread creates a vertical spread from two options
func (s *Scanner) createVerticalSpread(longLeg, shortLeg *models.Option, spreadType string) *models.VerticalSpread {
	// Calculate spread metrics
	netDebit := longLeg.Ask - shortLeg.Bid
	if netDebit <= 0 {
		return nil // Invalid spread
	}
	
	// Calculate max profit/loss
	strikeWidth := math.Abs(shortLeg.Strike - longLeg.Strike)
	maxProfit := strikeWidth - netDebit
	maxLoss := netDebit
	
	// Calculate breakeven
	var breakeven float64
	if spreadType == "CALL" {
		breakeven = longLeg.Strike + netDebit
	} else {
		breakeven = shortLeg.Strike - netDebit
	}
	
	// Estimate probability of profit (simplified)
	// In practice, this would use proper options pricing models
	probabilityProfit := s.estimateProbabilityOfProfit(longLeg, shortLeg, breakeven)
	
	spread := &models.VerticalSpread{
		ID:                generateSpreadID(longLeg, shortLeg),
		Symbol:           longLeg.Symbol,
		Type:             spreadType,
		LongLeg:          *longLeg,
		ShortLeg:         *shortLeg,
		NetDebit:         netDebit,
		MaxProfit:        maxProfit,
		MaxLoss:          maxLoss,
		Breakeven:        breakeven,
		ProbabilityProfit: probabilityProfit,
		CreatedAt:        time.Now(),
	}
	
	return spread
}

// scoreAndAnalyzeSpreads applies scoring and Greeks analysis to all spreads
func (s *Scanner) scoreAndAnalyzeSpreads(spreads []*models.VerticalSpread) {
	var wg sync.WaitGroup
	
	// Process in batches for efficiency
	batchSize := 100
	for i := 0; i < len(spreads); i += batchSize {
		end := i + batchSize
		if end > len(spreads) {
			end = len(spreads)
		}
		
		wg.Add(1)
		go func(batch []*models.VerticalSpread) {
			defer wg.Done()
			
			for _, spread := range batch {
				// Calculate score
				spread.Score = s.scorer.ScoreSpread(spread)
				
				// Analyze Greeks
				greeksReport := s.greeksAnalyzer.AnalyzeSpread(spread)
				spread.GreeksAnalysis = greeksReport
			}
		}(spreads[i:end])
	}
	
	wg.Wait()
}

// estimateProbabilityOfProfit estimates PoP based on delta
func (s *Scanner) estimateProbabilityOfProfit(longLeg, shortLeg *models.Option, breakeven float64) float64 {
	// Simplified estimation using delta as proxy for probability
	// For call spreads: PoP ≈ 1 - long delta
	// For put spreads: PoP ≈ short delta
	
	if longLeg.Right == "C" {
		// Bull call spread
		return 1.0 - math.Abs(longLeg.Delta)
	} else {
		// Bull put spread  
		return math.Abs(shortLeg.Delta)
	}
}

// generateSpreadID creates a unique ID for a spread
func generateSpreadID(longLeg, shortLeg *models.Option) string {
	return fmt.Sprintf("%s_%s_%s_%.2f_%.2f",
		longLeg.Symbol,
		longLeg.Expiry,
		longLeg.Right,
		longLeg.Strike,
		shortLeg.Strike,
	)
}