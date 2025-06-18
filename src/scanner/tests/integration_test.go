// +build integration

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ibkr-trader/scanner/api"
	"github.com/ibkr-trader/scanner/client"
	"github.com/ibkr-trader/scanner/internal/analytics"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/history"
	"github.com/ibkr-trader/scanner/internal/metrics"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/ibkr-trader/scanner/internal/service"
	"github.com/ibkr-trader/scanner/internal/streaming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationFullFlow tests the complete scanner flow
func TestIntegrationFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	
	// Start test server
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	// Create client
	cfg := client.Config{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	}
	c := client.NewClient(cfg)
	
	// Step 1: Health check
	t.Run("HealthCheck", func(t *testing.T) {
		health, err := c.Health(ctx)
		require.NoError(t, err)
		assert.Equal(t, "healthy", health.Status)
	})
	
	// Step 2: Get and update filters
	t.Run("FilterManagement", func(t *testing.T) {
		// Get current filters
		currentFilters, err := c.GetFilters(ctx)
		require.NoError(t, err)
		assert.NotNil(t, currentFilters)
		
		// Update filters
		newFilters := &client.FilterConfig{
			Delta: &client.DeltaFilter{
				Min: 0.20,
				Max: 0.35,
			},
			DTE: &client.DTEFilter{
				Min: 30,
				Max: 60,
			},
			Liquidity: &client.LiquidityFilter{
				MinOpenInterest: 100,
				MinVolume:       50,
			},
		}
		
		err = c.UpdateFilters(ctx, newFilters)
		require.NoError(t, err)
		
		// Verify update
		updatedFilters, err := c.GetFilters(ctx)
		require.NoError(t, err)
		assert.Equal(t, newFilters.Delta.Min, updatedFilters.Delta.Min)
	})
	
	// Step 3: Scan single symbol
	t.Run("SingleSymbolScan", func(t *testing.T) {
		result, err := c.ScanSymbol(ctx, "AAPL", nil)
		require.NoError(t, err)
		assert.Equal(t, "AAPL", result.Symbol)
		assert.Greater(t, result.TotalContracts, 0)
		
		// Check spreads
		if len(result.Spreads) > 0 {
			spread := result.Spreads[0]
			assert.Greater(t, spread.ShortStrike, spread.LongStrike)
			assert.Greater(t, spread.NetCredit, 0.0)
			assert.Greater(t, spread.Score, 0.0)
		}
	})
	
	// Step 4: Batch scan
	t.Run("BatchScan", func(t *testing.T) {
		symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}
		results, err := c.ScanMultiple(ctx, symbols, nil)
		require.NoError(t, err)
		assert.Len(t, results, len(symbols))
		
		for _, result := range results {
			assert.Contains(t, symbols, result.Symbol)
			assert.NotZero(t, result.ScanTime)
		}
	})
	
	// Step 5: Preset management
	t.Run("PresetManagement", func(t *testing.T) {
		// Get existing presets
		presets, err := c.GetPresets(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, presets)
		
		// Create custom preset
		customPreset := &client.PresetRequest{
			Name:        "Integration Test Preset",
			Description: "Test preset for integration testing",
			Filters: client.FilterConfig{
				Delta: &client.DeltaFilter{Min: 0.15, Max: 0.25},
				DTE:   &client.DTEFilter{Min: 21, Max: 45},
			},
			Tags: []string{"test", "integration"},
		}
		
		presetID, err := c.CreatePreset(ctx, customPreset)
		require.NoError(t, err)
		assert.NotEmpty(t, presetID)
	})
	
	// Step 6: WebSocket streaming
	t.Run("WebSocketStreaming", func(t *testing.T) {
		stream, err := c.Connect(ctx)
		require.NoError(t, err)
		defer stream.Close()
		
		// Track received messages
		var mu sync.Mutex
		received := make(map[string]int)
		
		// Register handlers
		stream.OnMessage("scan_result", func(msgType string, payload json.RawMessage) {
			mu.Lock()
			received["scan_result"]++
			mu.Unlock()
		})
		
		stream.OnMessage("subscribed", func(msgType string, payload json.RawMessage) {
			mu.Lock()
			received["subscribed"]++
			mu.Unlock()
		})
		
		// Subscribe to symbols
		err = stream.Subscribe([]string{"AAPL", "TSLA"}, nil)
		require.NoError(t, err)
		
		// Wait for messages
		time.Sleep(2 * time.Second)
		
		mu.Lock()
		assert.Greater(t, received["subscribed"], 0)
		mu.Unlock()
	})
	
	// Step 7: Analytics
	t.Run("Analytics", func(t *testing.T) {
		// Get statistics
		stats, err := c.GetStatistics(ctx, "AAPL", "", "24h")
		require.NoError(t, err)
		assert.NotEmpty(t, stats)
		
		// Get history
		historyParams := &client.HistoryParams{
			Symbol:   "AAPL",
			Page:     1,
			PageSize: 10,
		}
		
		history, err := c.GetHistory(ctx, historyParams)
		require.NoError(t, err)
		assert.NotNil(t, history.Pagination)
	})
}

// TestIntegrationConcurrentScans tests concurrent scanning
func TestIntegrationConcurrentScans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	// Create multiple clients
	numClients := 10
	clients := make([]*client.Client, numClients)
	for i := 0; i < numClients; i++ {
		cfg := client.Config{
			BaseURL: baseURL,
			Timeout: 30 * time.Second,
		}
		clients[i] = client.NewClient(cfg)
	}
	
	// Symbols to scan
	symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "JPM", "BAC", "WMT"}
	
	// Concurrent scans
	var wg sync.WaitGroup
	results := make(chan *client.ScanResult, len(symbols)*numClients)
	errors := make(chan error, len(symbols)*numClients)
	
	for i, c := range clients {
		for j, symbol := range symbols {
			wg.Add(1)
			go func(client *client.Client, sym string, idx int) {
				defer wg.Done()
				
				// Add some jitter
				time.Sleep(time.Duration(idx*10) * time.Millisecond)
				
				result, err := client.ScanSymbol(ctx, sym, nil)
				if err != nil {
					errors <- err
					return
				}
				results <- result
			}(c, symbol, i*len(symbols)+j)
		}
	}
	
	// Wait for completion
	wg.Wait()
	close(results)
	close(errors)
	
	// Check results
	var errorCount int
	for err := range errors {
		t.Logf("Scan error: %v", err)
		errorCount++
	}
	
	var resultCount int
	for result := range results {
		assert.NotEmpty(t, result.Symbol)
		resultCount++
	}
	
	// Allow some errors but most should succeed
	assert.Less(t, errorCount, numClients*len(symbols)/10) // Less than 10% errors
	assert.Greater(t, resultCount, numClients*len(symbols)/2) // More than 50% success
}

// TestIntegrationWebSocketReconnect tests WebSocket reconnection
func TestIntegrationWebSocketReconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	cfg := client.Config{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	}
	c := client.NewClient(cfg)
	
	// Connect
	stream1, err := c.Connect(ctx)
	require.NoError(t, err)
	
	// Subscribe
	err = stream1.Subscribe([]string{"AAPL"}, nil)
	require.NoError(t, err)
	
	// Close connection
	stream1.Close()
	
	// Reconnect
	stream2, err := c.Connect(ctx)
	require.NoError(t, err)
	defer stream2.Close()
	
	// Subscribe again
	err = stream2.Subscribe([]string{"AAPL"}, nil)
	require.NoError(t, err)
}

// TestIntegrationFilterValidation tests filter validation
func TestIntegrationFilterValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	cfg := client.Config{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	}
	c := client.NewClient(cfg)
	
	// Test invalid filters
	invalidFilters := &client.FilterConfig{
		Delta: &client.DeltaFilter{
			Min: 0.5,  // Invalid: min > max
			Max: 0.3,
		},
	}
	
	err := c.UpdateFilters(ctx, invalidFilters)
	assert.Error(t, err)
	
	// Test valid edge cases
	edgeCaseFilters := &client.FilterConfig{
		Delta: &client.DeltaFilter{
			Min: 0.0,
			Max: 1.0,
		},
		DTE: &client.DTEFilter{
			Min: 0,
			Max: 365,
		},
	}
	
	err = c.UpdateFilters(ctx, edgeCaseFilters)
	assert.NoError(t, err)
}

// TestIntegrationHistoricalData tests historical data accumulation
func TestIntegrationHistoricalData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	cfg := client.Config{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	}
	c := client.NewClient(cfg)
	
	// Perform multiple scans
	symbol := "AAPL"
	numScans := 5
	
	for i := 0; i < numScans; i++ {
		_, err := c.ScanSymbol(ctx, symbol, nil)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}
	
	// Check history
	historyParams := &client.HistoryParams{
		Symbol:   symbol,
		Page:     1,
		PageSize: 10,
	}
	
	history, err := c.GetHistory(ctx, historyParams)
	require.NoError(t, err)
	
	// Should have accumulated history
	assert.Greater(t, history.Pagination.Total, 0)
}

// Helper function to start test server
func startTestServer(t *testing.T) (*api.Server, string) {
	// Create dependencies
	scanner := service.NewScanner(nil)
	streamer := streaming.NewManager()
	analytics := analytics.NewEngine()
	history := history.NewStore()
	metrics := metrics.NewCollector()
	
	// Create server with random port
	config := api.DefaultConfig()
	config.Port = 0 // Random available port
	
	server := api.NewServer(config, scanner, streamer, analytics, history, metrics)
	
	// Start server
	err := server.Start()
	require.NoError(t, err)
	
	// Get actual port
	baseURL := fmt.Sprintf("http://localhost:%d/api/v1", config.Port)
	
	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	
	return server, baseURL
}

// BenchmarkIntegrationScan benchmarks end-to-end scanning
func BenchmarkIntegrationScan(b *testing.B) {
	ctx := context.Background()
	server, baseURL := startTestServer(nil)
	defer server.Stop(ctx)
	
	cfg := client.Config{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	}
	c := client.NewClient(cfg)
	
	// Warm up
	c.ScanSymbol(ctx, "AAPL", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ScanSymbol(ctx, "AAPL", nil)
	}
}

// BenchmarkIntegrationConcurrentScans benchmarks concurrent scanning
func BenchmarkIntegrationConcurrentScans(b *testing.B) {
	ctx := context.Background()
	server, baseURL := startTestServer(nil)
	defer server.Stop(ctx)
	
	numClients := 10
	clients := make([]*client.Client, numClients)
	for i := 0; i < numClients; i++ {
		cfg := client.Config{
			BaseURL: baseURL,
			Timeout: 30 * time.Second,
		}
		clients[i] = client.NewClient(cfg)
	}
	
	symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		clientIdx := 0
		symbolIdx := 0
		for pb.Next() {
			client := clients[clientIdx%numClients]
			symbol := symbols[symbolIdx%len(symbols)]
			
			_, _ = client.ScanSymbol(ctx, symbol, nil)
			
			clientIdx++
			symbolIdx++
		}
	})
}