package greeks

import (
	"fmt"
	"math"
	"github.com/new-ibkr-trader/src/go/pkg/models"
)

// GreeksConfig defines thresholds and targets for Greeks analysis
type GreeksConfig struct {
	// Delta configuration
	DeltaMin       float64 `json:"delta_min"`
	DeltaMax       float64 `json:"delta_max"`
	DeltaNeutral   bool    `json:"delta_neutral"`
	
	// Gamma configuration
	GammaMaxRisk   float64 `json:"gamma_max_risk"`
	GammaWarning   float64 `json:"gamma_warning"`
	
	// Theta configuration
	ThetaMinDaily  float64 `json:"theta_min_daily"`
	ThetaAsPercent bool    `json:"theta_as_percent"`
	
	// Vega configuration
	VegaMaxExposure float64 `json:"vega_max_exposure"`
	VegaNeutral     bool    `json:"vega_neutral"`
	
	// Rho configuration (interest rate sensitivity)
	RhoMaxExposure  float64 `json:"rho_max_exposure"`
}

// DefaultGreeksConfig returns a balanced Greeks configuration
func DefaultGreeksConfig() *GreeksConfig {
	return &GreeksConfig{
		DeltaMin:        0.20,
		DeltaMax:        0.40,
		DeltaNeutral:    false,
		GammaMaxRisk:    0.10,
		GammaWarning:    0.05,
		ThetaMinDaily:   0.01, // 1% of premium per day
		ThetaAsPercent:  true,
		VegaMaxExposure: 0.20,
		VegaNeutral:     false,
		RhoMaxExposure:  0.05,
	}
}

// GreeksReport contains the analysis results
type GreeksReport struct {
	// Net Greeks
	NetDelta   float64 `json:"net_delta"`
	NetGamma   float64 `json:"net_gamma"`
	NetTheta   float64 `json:"net_theta"`
	NetVega    float64 `json:"net_vega"`
	NetRho     float64 `json:"net_rho"`
	
	// Risk metrics
	DeltaRisk      string  `json:"delta_risk"`      // LOW, MEDIUM, HIGH
	GammaRisk      string  `json:"gamma_risk"`
	ThetaCapture   float64 `json:"theta_capture"`   // Daily theta as % of investment
	VegaExposure   string  `json:"vega_exposure"`
	
	// Analysis
	IsBalanced     bool     `json:"is_balanced"`
	RiskScore      float64  `json:"risk_score"`      // 0-100, lower is better
	Warnings       []string `json:"warnings"`
	Recommendations []string `json:"recommendations"`
}

// GreeksAnalyzer performs comprehensive Greeks analysis
type GreeksAnalyzer struct {
	config *GreeksConfig
}

// NewGreeksAnalyzer creates a new analyzer with the given configuration
func NewGreeksAnalyzer(config *GreeksConfig) *GreeksAnalyzer {
	if config == nil {
		config = DefaultGreeksConfig()
	}
	return &GreeksAnalyzer{config: config}
}

// AnalyzeSpread performs comprehensive Greeks analysis on a vertical spread
func (a *GreeksAnalyzer) AnalyzeSpread(spread *models.VerticalSpread) *GreeksReport {
	report := &GreeksReport{
		Warnings:        []string{},
		Recommendations: []string{},
	}
	
	// Calculate net Greeks
	report.NetDelta = spread.ShortLeg.Delta - spread.LongLeg.Delta
	report.NetGamma = spread.ShortLeg.Gamma - spread.LongLeg.Gamma
	report.NetTheta = spread.ShortLeg.Theta - spread.LongLeg.Theta
	report.NetVega = spread.ShortLeg.Vega - spread.LongLeg.Vega
	report.NetRho = a.calculateNetRho(spread)
	
	// Analyze each Greek
	a.analyzeDelta(spread, report)
	a.analyzeGamma(spread, report)
	a.analyzeTheta(spread, report)
	a.analyzeVega(spread, report)
	a.analyzeRho(spread, report)
	
	// Calculate overall risk score
	report.RiskScore = a.calculateRiskScore(report)
	
	// Determine if spread is well-balanced
	report.IsBalanced = a.isBalanced(report)
	
	// Generate recommendations
	a.generateRecommendations(spread, report)
	
	return report
}

// analyzeDelta analyzes delta exposure and risk
func (a *GreeksAnalyzer) analyzeDelta(spread *models.VerticalSpread, report *GreeksReport) {
	longDelta := math.Abs(spread.LongLeg.Delta)
	
	// Check if long leg delta is within preferred range
	if longDelta < a.config.DeltaMin {
		report.DeltaRisk = "LOW"
		report.Warnings = append(report.Warnings, 
			fmt.Sprintf("Long delta %.2f below minimum %.2f - low probability of profit", 
				longDelta, a.config.DeltaMin))
	} else if longDelta > a.config.DeltaMax {
		report.DeltaRisk = "HIGH"
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("Long delta %.2f above maximum %.2f - high directional risk",
				longDelta, a.config.DeltaMax))
	} else {
		report.DeltaRisk = "MEDIUM"
	}
	
	// Check delta neutrality if configured
	if a.config.DeltaNeutral && math.Abs(report.NetDelta) > 0.10 {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("Net delta %.2f not neutral - directional bias present", report.NetDelta))
	}
}

// analyzeGamma analyzes gamma risk
func (a *GreeksAnalyzer) analyzeGamma(spread *models.VerticalSpread, report *GreeksReport) {
	// Gamma risk is highest near the money
	maxGamma := math.Max(math.Abs(spread.LongLeg.Gamma), math.Abs(spread.ShortLeg.Gamma))
	
	if maxGamma > a.config.GammaMaxRisk {
		report.GammaRisk = "HIGH"
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("High gamma risk %.3f - large delta changes possible", maxGamma))
	} else if maxGamma > a.config.GammaWarning {
		report.GammaRisk = "MEDIUM"
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("Moderate gamma exposure %.3f", maxGamma))
	} else {
		report.GammaRisk = "LOW"
	}
	
	// Net negative gamma is preferred for spreads
	if report.NetGamma > 0 {
		report.Warnings = append(report.Warnings,
			"Positive net gamma - unusual for credit spreads")
	}
}

// analyzeTheta analyzes time decay capture
func (a *GreeksAnalyzer) analyzeTheta(spread *models.VerticalSpread, report *GreeksReport) {
	// Calculate daily theta capture
	if a.config.ThetaAsPercent && spread.NetDebit > 0 {
		report.ThetaCapture = (report.NetTheta / spread.NetDebit) * 100
		
		if report.ThetaCapture < a.config.ThetaMinDaily {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Low theta capture %.2f%% per day", report.ThetaCapture))
		}
	} else {
		report.ThetaCapture = report.NetTheta
	}
	
	// Theta should be positive (we collect time decay)
	if report.NetTheta <= 0 {
		report.Warnings = append(report.Warnings,
			"Negative net theta - paying time decay instead of collecting")
	}
}

// analyzeVega analyzes volatility exposure
func (a *GreeksAnalyzer) analyzeVega(spread *models.VerticalSpread, report *GreeksReport) {
	absNetVega := math.Abs(report.NetVega)
	
	if absNetVega > a.config.VegaMaxExposure {
		report.VegaExposure = "HIGH"
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("High vega exposure %.3f - sensitive to IV changes", absNetVega))
	} else if absNetVega > a.config.VegaMaxExposure * 0.5 {
		report.VegaExposure = "MEDIUM"
	} else {
		report.VegaExposure = "LOW"
	}
	
	// Check vega neutrality if configured
	if a.config.VegaNeutral && absNetVega > 0.05 {
		report.Warnings = append(report.Warnings,
			"Spread not vega neutral - volatility risk present")
	}
}

// analyzeRho analyzes interest rate sensitivity
func (a *GreeksAnalyzer) analyzeRho(spread *models.VerticalSpread, report *GreeksReport) {
	if math.Abs(report.NetRho) > a.config.RhoMaxExposure {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("High rho exposure %.3f - sensitive to interest rate changes", report.NetRho))
	}
}

// calculateNetRho estimates net rho (not always provided by brokers)
func (a *GreeksAnalyzer) calculateNetRho(spread *models.VerticalSpread) float64 {
	// Simplified rho calculation
	// In practice, this would use the actual rho values if available
	return 0.01 // Placeholder
}

// calculateRiskScore computes overall risk score (0-100)
func (a *GreeksAnalyzer) calculateRiskScore(report *GreeksReport) float64 {
	score := 0.0
	
	// Delta risk contribution (0-25)
	switch report.DeltaRisk {
	case "LOW":
		score += 5
	case "MEDIUM":
		score += 15
	case "HIGH":
		score += 25
	}
	
	// Gamma risk contribution (0-25)
	switch report.GammaRisk {
	case "LOW":
		score += 5
	case "MEDIUM":
		score += 15
	case "HIGH":
		score += 25
	}
	
	// Theta contribution (0-25)
	if report.NetTheta <= 0 {
		score += 25
	} else if report.ThetaCapture < 0.5 {
		score += 15
	} else if report.ThetaCapture < 1.0 {
		score += 10
	} else {
		score += 5
	}
	
	// Vega contribution (0-25)
	switch report.VegaExposure {
	case "LOW":
		score += 5
	case "MEDIUM":
		score += 15
	case "HIGH":
		score += 25
	}
	
	return score
}

// isBalanced determines if the spread has balanced Greeks
func (a *GreeksAnalyzer) isBalanced(report *GreeksReport) bool {
	// A balanced spread has:
	// - Moderate delta risk
	// - Low to medium gamma risk
	// - Positive theta
	// - Low vega exposure
	
	return report.DeltaRisk != "HIGH" &&
		report.GammaRisk != "HIGH" &&
		report.NetTheta > 0 &&
		report.VegaExposure != "HIGH" &&
		report.RiskScore < 50
}

// generateRecommendations creates actionable recommendations
func (a *GreeksAnalyzer) generateRecommendations(spread *models.VerticalSpread, report *GreeksReport) {
	// Delta recommendations
	if report.DeltaRisk == "LOW" {
		report.Recommendations = append(report.Recommendations,
			"Consider strikes closer to the money for higher probability")
	} else if report.DeltaRisk == "HIGH" {
		report.Recommendations = append(report.Recommendations,
			"Consider strikes further out of the money to reduce directional risk")
	}
	
	// Gamma recommendations
	if report.GammaRisk == "HIGH" {
		report.Recommendations = append(report.Recommendations,
			"High gamma risk - consider wider strikes or different expiration")
	}
	
	// Theta recommendations
	if report.NetTheta <= 0 {
		report.Recommendations = append(report.Recommendations,
			"Negative theta - restructure spread to collect time decay")
	} else if report.ThetaCapture < 0.5 {
		report.Recommendations = append(report.Recommendations,
			"Low theta capture - consider strikes with higher time decay")
	}
	
	// Vega recommendations
	if report.VegaExposure == "HIGH" {
		report.Recommendations = append(report.Recommendations,
			"High vega exposure - consider different expirations to reduce IV sensitivity")
	}
	
	// Overall balance
	if !report.IsBalanced {
		report.Recommendations = append(report.Recommendations,
			"Spread is not well-balanced - review strike selection and expiration")
	}
}

// CompareGreeks compares Greeks between multiple spreads
func (a *GreeksAnalyzer) CompareGreeks(spreads []*models.VerticalSpread) []*GreeksComparison {
	comparisons := make([]*GreeksComparison, len(spreads))
	
	for i, spread := range spreads {
		report := a.AnalyzeSpread(spread)
		comparisons[i] = &GreeksComparison{
			Spread:     spread,
			Report:     report,
			Ranking:    0, // Will be set after all are analyzed
		}
	}
	
	// Rank by risk score (lower is better)
	// In practice, this would be more sophisticated
	for i := range comparisons {
		rank := 1
		for j := range comparisons {
			if i != j && comparisons[j].Report.RiskScore < comparisons[i].Report.RiskScore {
				rank++
			}
		}
		comparisons[i].Ranking = rank
	}
	
	return comparisons
}

// GreeksComparison holds spread comparison data
type GreeksComparison struct {
	Spread  *models.VerticalSpread
	Report  *GreeksReport
	Ranking int
}