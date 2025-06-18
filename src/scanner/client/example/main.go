package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ibkr-trader/scanner/client"
)

func main() {
	// Create client with default configuration
	cfg := client.DefaultConfig()
	// Optionally set API key
	// cfg.APIKey = "your-api-key"
	
	c := client.NewClient(cfg)
	ctx := context.Background()
	
	// Example 1: Check service health
	fmt.Println("=== Health Check ===")
	health, err := c.Health(ctx)
	if err != nil {
		log.Fatal("Health check failed:", err)
	}
	fmt.Printf("Service Status: %s\n", health.Status)
	fmt.Printf("Version: %s\n", health.Version)
	fmt.Printf("Uptime: %.2f seconds\n\n", health.Uptime)
	
	// Example 2: Scan a single symbol
	fmt.Println("=== Single Symbol Scan ===")
	filters := &client.ScanFilters{
		DeltaMin: 0.20,
		DeltaMax: 0.35,
		DTEMin:   30,
		DTEMax:   60,
	}
	
	result, err := c.ScanSymbol(ctx, "AAPL", filters)
	if err != nil {
		log.Fatal("Scan failed:", err)
	}
	
	fmt.Printf("Symbol: %s\n", result.Symbol)
	fmt.Printf("Total Contracts: %d\n", result.TotalContracts)
	fmt.Printf("Filtered Contracts: %d\n", result.FilteredContracts)
	fmt.Printf("Found %d spread opportunities\n\n", len(result.Spreads))
	
	// Display top 3 spreads
	for i := 0; i < 3 && i < len(result.Spreads); i++ {
		spread := result.Spreads[i]
		fmt.Printf("Spread %d:\n", i+1)
		fmt.Printf("  Strikes: %.2f/%.2f\n", spread.LongStrike, spread.ShortStrike)
		fmt.Printf("  Credit: $%.2f\n", spread.NetCredit)
		fmt.Printf("  PoP: %.2f%%\n", spread.ProbabilityProfit*100)
		fmt.Printf("  Score: %.2f\n\n", spread.Score)
	}
	
	// Example 3: Scan multiple symbols
	fmt.Println("=== Multiple Symbol Scan ===")
	symbols := []string{"MSFT", "GOOGL", "AMZN"}
	filterMap := map[string]interface{}{
		"delta": map[string]float64{
			"min": 0.25,
			"max": 0.35,
		},
		"dte": map[string]int{
			"min": 21,
			"max": 45,
		},
		"liquidity": map[string]int{
			"min_open_interest": 100,
			"min_volume":        50,
		},
	}
	
	results, err := c.ScanMultiple(ctx, symbols, filterMap)
	if err != nil {
		log.Fatal("Multiple scan failed:", err)
	}
	
	for _, result := range results {
		fmt.Printf("%s: %d spreads found\n", result.Symbol, len(result.Spreads))
	}
	fmt.Println()
	
	// Example 4: Get and use filter presets
	fmt.Println("=== Filter Presets ===")
	presets, err := c.GetPresets(ctx)
	if err != nil {
		log.Fatal("Failed to get presets:", err)
	}
	
	for _, preset := range presets {
		fmt.Printf("- %s: %s\n", preset.Name, preset.Description)
	}
	fmt.Println()
	
	// Example 5: WebSocket streaming
	fmt.Println("=== WebSocket Streaming ===")
	stream, err := c.Connect(ctx)
	if err != nil {
		log.Fatal("Failed to connect to WebSocket:", err)
	}
	defer stream.Close()
	
	// Register handlers
	stream.OnMessage("scan_result", func(msgType string, payload json.RawMessage) {
		var result client.ScanResult
		if err := json.Unmarshal(payload, &result); err != nil {
			log.Printf("Failed to parse scan result: %v", err)
			return
		}
		fmt.Printf("[STREAM] %s: %d spreads found\n", result.Symbol, len(result.Spreads))
	})
	
	stream.OnMessage("error", func(msgType string, payload json.RawMessage) {
		var errMsg map[string]string
		if err := json.Unmarshal(payload, &errMsg); err != nil {
			log.Printf("Failed to parse error: %v", err)
			return
		}
		fmt.Printf("[ERROR] %s\n", errMsg["message"])
	})
	
	// Subscribe to symbols
	err = stream.Subscribe([]string{"AAPL", "TSLA"}, nil)
	if err != nil {
		log.Fatal("Failed to subscribe:", err)
	}
	
	fmt.Println("Subscribed to AAPL and TSLA. Listening for updates...")
	
	// Listen for 30 seconds
	time.Sleep(30 * time.Second)
	
	// Example 6: Get statistics
	fmt.Println("\n=== Analytics Statistics ===")
	stats, err := c.GetStatistics(ctx, "AAPL", "", "24h")
	if err != nil {
		log.Fatal("Failed to get statistics:", err)
	}
	
	fmt.Printf("Statistics for AAPL (24h):\n")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	// Example 7: Get historical data
	fmt.Println("\n=== Historical Data ===")
	historyParams := &client.HistoryParams{
		Symbol:   "AAPL",
		Page:     1,
		PageSize: 10,
	}
	
	history, err := c.GetHistory(ctx, historyParams)
	if err != nil {
		log.Fatal("Failed to get history:", err)
	}
	
	fmt.Printf("Found %d total historical records\n", history.Pagination.Total)
	fmt.Printf("Showing page %d of %d\n", history.Pagination.Page, history.Pagination.TotalPages)
}

// Helper function to pretty print JSON
func prettyPrint(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}