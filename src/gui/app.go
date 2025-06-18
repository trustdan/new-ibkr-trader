package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// App struct
type App struct {
	ctx          context.Context
	scannerURL   string
	pythonURL    string
	systemHealth *SystemHealth
}

// SystemHealth represents the health status of all services
type SystemHealth struct {
	TWS struct {
		Connected bool  `json:"connected"`
		Uptime    int64 `json:"uptime"`
	} `json:"tws"`
	Subscriptions struct {
		Active   int `json:"active"`
		Max      int `json:"max"`
		UsagePct int `json:"usage_pct"`
	} `json:"subscriptions"`
	Queue struct {
		Size       int  `json:"size"`
		Processing bool `json:"processing"`
	} `json:"queue"`
	Throttling bool     `json:"throttling"`
	Errors     []string `json:"errors"`
}

// ScanRequest represents a scanning request
type ScanRequest struct {
	Symbol     string         `json:"symbol"`
	Filters    []FilterConfig `json:"filters"`
	MaxResults int            `json:"max_results"`
}

// FilterConfig represents filter configuration
type FilterConfig struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// ScanResponse represents scanner results
type ScanResponse struct {
	Symbol      string    `json:"symbol"`
	ScanTime    time.Time `json:"scan_time"`
	ResultCount int       `json:"result_count"`
	Options     []Option  `json:"options"`
}

// Option represents an option contract
type Option struct {
	Symbol    string  `json:"symbol"`
	Strike    float64 `json:"strike"`
	Expiry    string  `json:"expiry"`
	Right     string  `json:"right"`
	Delta     float64 `json:"delta"`
	Gamma     float64 `json:"gamma"`
	Theta     float64 `json:"theta"`
	Vega      float64 `json:"vega"`
	IV        float64 `json:"iv"`
	Volume    int     `json:"volume"`
	OpenInt   int     `json:"open_interest"`
	BidPrice  float64 `json:"bid_price"`
	AskPrice  float64 `json:"ask_price"`
	LastPrice float64 `json:"last_price"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		scannerURL:   "http://localhost:8080", // Go scanner service
		pythonURL:    "http://localhost:8000", // Python IBKR service
		systemHealth: &SystemHealth{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize system health monitoring
	go a.monitorSystemHealth()
}

// GetSystemHealth returns current system health status
func (a *App) GetSystemHealth() (*SystemHealth, error) {
	// Fetch health from Python service
	resp, err := http.Get(a.pythonURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("failed to get Python service health: %w", err)
	}
	defer resp.Body.Close()

	var health SystemHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	a.systemHealth = &health
	return &health, nil
}

// ScanOptions performs an options scan
func (a *App) ScanOptions(request ScanRequest) (*ScanResponse, error) {
	// Send request to Go scanner service
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(a.scannerURL+"/api/v1/scan", "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send scan request: %w", err)
	}
	defer resp.Body.Close()

	var scanResponse ScanResponse
	if err := json.NewDecoder(resp.Body).Decode(&scanResponse); err != nil {
		return nil, fmt.Errorf("failed to decode scan response: %w", err)
	}

	return &scanResponse, nil
}

// monitorSystemHealth continuously monitors system health
func (a *App) monitorSystemHealth() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := a.GetSystemHealth()
			if err != nil {
				// Log error but continue monitoring
				fmt.Printf("Health check failed: %v\n", err)
			}
		case <-a.ctx.Done():
			return
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s! Welcome to IBKR Spread Automation!", name)
}
