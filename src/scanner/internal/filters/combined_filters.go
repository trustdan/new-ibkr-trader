package filters

import (
	"github.com/ibkr-trader/scanner/internal/models"
	"sort"
)

// CorrelationFilter filters based on correlation between contracts and spreads
type CorrelationFilter struct {
	MaxCorrelation float64 `json:"max_correlation"`
	SymbolGroups   map[string][]string `json:"symbol_groups"` // Groups of correlated symbols
}

func (f *CorrelationFilter) Name() string { return "CorrelationFilter" }

func (f *CorrelationFilter) ApplyToCombined(contracts []models.OptionContract, spreads []models.VerticalSpread) ([]models.OptionContract, []models.VerticalSpread) {
	if f == nil || f.SymbolGroups == nil {
		return contracts, spreads
	}
	
	// Track symbols already in spreads by group
	groupSymbols := make(map[string]map[string]bool)
	
	// Initialize groups
	for group := range f.SymbolGroups {
		groupSymbols[group] = make(map[string]bool)
	}
	
	// Mark symbols used in spreads
	for _, spread := range spreads {
		symbol := spread.Symbol
		for group, symbols := range f.SymbolGroups {
			for _, s := range symbols {
				if s == symbol {
					groupSymbols[group][symbol] = true
					break
				}
			}
		}
	}
	
	// Filter out contracts from same correlation group
	filteredContracts := make([]models.OptionContract, 0)
	for _, contract := range contracts {
		canAdd := true
		
		// Check if contract's symbol is in a group with existing positions
		for group, symbols := range f.SymbolGroups {
			for _, s := range symbols {
				if s == contract.Symbol && len(groupSymbols[group]) > 0 {
					// Check if we'd exceed correlation limit
					if float64(len(groupSymbols[group])+1) / float64(len(symbols)) > f.MaxCorrelation {
						canAdd = false
						break
					}
				}
			}
			if !canAdd {
				break
			}
		}
		
		if canAdd {
			filteredContracts = append(filteredContracts, contract)
		}
	}
	
	// Similarly filter spreads
	filteredSpreads := make([]models.VerticalSpread, 0)
	groupCounts := make(map[string]int)
	
	for _, spread := range spreads {
		canAdd := true
		spreadGroup := ""
		
		// Find which group this spread belongs to
		for group, symbols := range f.SymbolGroups {
			for _, s := range symbols {
				if s == spread.Symbol {
					spreadGroup = group
					break
				}
			}
			if spreadGroup != "" {
				break
			}
		}
		
		if spreadGroup != "" {
			// Check correlation limit
			if float64(groupCounts[spreadGroup]+1) / float64(len(f.SymbolGroups[spreadGroup])) > f.MaxCorrelation {
				canAdd = false
			}
		}
		
		if canAdd {
			filteredSpreads = append(filteredSpreads, spread)
			if spreadGroup != "" {
				groupCounts[spreadGroup]++
			}
		}
	}
	
	return filteredContracts, filteredSpreads
}

func (f *CorrelationFilter) Validate() error { return nil }

// PortfolioBalanceFilter ensures portfolio balance across strategies
type PortfolioBalanceFilter struct {
	MaxAllocation      float64            `json:"max_allocation"`       // Max % allocation per symbol
	StrategyLimits     map[string]int     `json:"strategy_limits"`      // Max positions per strategy
	SectorLimits       map[string]float64 `json:"sector_limits"`        // Max % per sector
	SymbolToSector     map[string]string  `json:"symbol_to_sector"`     // Symbol to sector mapping
}

func (f *PortfolioBalanceFilter) Name() string { return "PortfolioBalanceFilter" }

func (f *PortfolioBalanceFilter) ApplyToCombined(contracts []models.OptionContract, spreads []models.VerticalSpread) ([]models.OptionContract, []models.VerticalSpread) {
	if f == nil {
		return contracts, spreads
	}
	
	// Track current allocations
	symbolAllocations := make(map[string]float64)
	sectorAllocations := make(map[string]float64)
	strategyCount := make(map[string]int)
	
	// Count existing spread positions
	totalValue := 0.0
	for _, spread := range spreads {
		value := spread.Credit * 100 // Assuming 1 contract = 100 shares
		symbolAllocations[spread.Symbol] += value
		totalValue += value
		
		// Track sector allocation
		if sector, exists := f.SymbolToSector[spread.Symbol]; exists {
			sectorAllocations[sector] += value
		}
		
		// Track strategy count (simplified - all spreads are "vertical")
		strategyCount["vertical"]++
	}
	
	// Filter contracts based on allocation limits
	filteredContracts := make([]models.OptionContract, 0)
	for _, contract := range contracts {
		// Check symbol allocation
		currentAlloc := symbolAllocations[contract.Symbol] / totalValue
		if currentAlloc >= f.MaxAllocation {
			continue
		}
		
		// Check sector allocation
		if sector, exists := f.SymbolToSector[contract.Symbol]; exists {
			sectorAlloc := sectorAllocations[sector] / totalValue
			if limit, hasLimit := f.SectorLimits[sector]; hasLimit && sectorAlloc >= limit {
				continue
			}
		}
		
		filteredContracts = append(filteredContracts, contract)
	}
	
	// Filter spreads based on limits
	filteredSpreads := make([]models.VerticalSpread, 0)
	for _, spread := range spreads {
		// Check strategy limits
		if limit, exists := f.StrategyLimits["vertical"]; exists && strategyCount["vertical"] >= limit {
			continue
		}
		
		filteredSpreads = append(filteredSpreads, spread)
	}
	
	return filteredContracts, filteredSpreads
}

func (f *PortfolioBalanceFilter) Validate() error { return nil }

// RankingFilter ranks and limits results based on scoring
type RankingFilter struct {
	MaxContracts   int     `json:"max_contracts"`
	MaxSpreads     int     `json:"max_spreads"`
	ScoreThreshold float64 `json:"score_threshold"`
}

func (f *RankingFilter) Name() string { return "RankingFilter" }

func (f *RankingFilter) ApplyToCombined(contracts []models.OptionContract, spreads []models.VerticalSpread) ([]models.OptionContract, []models.VerticalSpread) {
	if f == nil {
		return contracts, spreads
	}
	
	// Sort contracts by score (descending)
	sort.Slice(contracts, func(i, j int) bool {
		return contracts[i].Score > contracts[j].Score
	})
	
	// Filter contracts by score threshold and limit
	filteredContracts := make([]models.OptionContract, 0)
	for i, contract := range contracts {
		if contract.Score < f.ScoreThreshold {
			break
		}
		if i >= f.MaxContracts {
			break
		}
		filteredContracts = append(filteredContracts, contract)
	}
	
	// Sort spreads by score (descending)
	sort.Slice(spreads, func(i, j int) bool {
		return spreads[i].Score > spreads[j].Score
	})
	
	// Filter spreads by score threshold and limit
	filteredSpreads := make([]models.VerticalSpread, 0)
	for i, spread := range spreads {
		if spread.Score < f.ScoreThreshold {
			break
		}
		if i >= f.MaxSpreads {
			break
		}
		filteredSpreads = append(filteredSpreads, spread)
	}
	
	return filteredContracts, filteredSpreads
}

func (f *RankingFilter) Validate() error { return nil }

// TimeDecayOptimizer optimizes for theta collection
type TimeDecayOptimizer struct {
	MinDailyTheta   float64 `json:"min_daily_theta"`
	MaxThetaRisk    float64 `json:"max_theta_risk"`
	PreferredDTE    int     `json:"preferred_dte"`
	DTEWeight       float64 `json:"dte_weight"`
}

func (f *TimeDecayOptimizer) Name() string { return "TimeDecayOptimizer" }

func (f *TimeDecayOptimizer) ApplyToCombined(contracts []models.OptionContract, spreads []models.VerticalSpread) ([]models.OptionContract, []models.VerticalSpread) {
	if f == nil {
		return contracts, spreads
	}
	
	// Filter contracts optimized for theta
	filteredContracts := make([]models.OptionContract, 0)
	for _, contract := range contracts {
		// For selling, we want positive theta
		if contract.Theta >= f.MinDailyTheta {
			// Adjust score based on DTE preference
			dteDistance := float64(abs(contract.DTE - f.PreferredDTE))
			adjustedScore := contract.Score * (1 - f.DTEWeight*dteDistance/100)
			contract.Score = adjustedScore
			
			filteredContracts = append(filteredContracts, contract)
		}
	}
	
	// Filter spreads optimized for theta collection
	filteredSpreads := make([]models.VerticalSpread, 0)
	for _, spread := range spreads {
		netTheta := spread.ShortLeg.Theta + spread.LongLeg.Theta
		
		if netTheta >= f.MinDailyTheta && netTheta <= f.MaxThetaRisk {
			// Adjust score based on DTE preference
			avgDTE := (spread.ShortLeg.DTE + spread.LongLeg.DTE) / 2
			dteDistance := float64(abs(avgDTE - f.PreferredDTE))
			adjustedScore := spread.Score * (1 - f.DTEWeight*dteDistance/100)
			spread.Score = adjustedScore
			
			filteredSpreads = append(filteredSpreads, spread)
		}
	}
	
	return filteredContracts, filteredSpreads
}

func (f *TimeDecayOptimizer) Validate() error { return nil }

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}