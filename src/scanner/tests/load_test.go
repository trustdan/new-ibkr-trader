// +build load

package tests

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ibkr-trader/scanner/client"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadTestConfig holds load test configuration
type LoadTestConfig struct {
	NumClients       int
	NumSymbols       int
	TestDuration     time.Duration
	RequestsPerSec   int
	ContractsPerScan int
}

// LoadTestResults holds test results
type LoadTestResults struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalLatency       int64
	MaxLatency         int64
	MinLatency         int64
	ContractsProcessed int64
	StartTime          time.Time
	EndTime            time.Time
}

// TestLoad10kContracts tests scanning with 10k+ contracts
func TestLoad10kContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	// Configure load test
	config := LoadTestConfig{
		NumClients:       50,
		NumSymbols:       100,
		TestDuration:     5 * time.Minute,
		RequestsPerSec:   100,
		ContractsPerScan: 150, // Average contracts per symbol
	}
	
	// Generate test symbols
	symbols := generateTestSymbols(config.NumSymbols)
	
	// Create clients
	clients := make([]*client.Client, config.NumClients)
	for i := 0; i < config.NumClients; i++ {
		cfg := client.Config{
			BaseURL: baseURL,
			Timeout: 30 * time.Second,
		}
		clients[i] = client.NewClient(cfg)
	}
	
	// Run load test
	results := runLoadTest(t, clients, symbols, config)
	
	// Analyze results
	analyzeResults(t, results, config)
}

// TestLoadConcurrentBatchScans tests concurrent batch scanning
func TestLoadConcurrentBatchScans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	cfg := client.Config{
		BaseURL: baseURL,
		Timeout: 60 * time.Second,
	}
	c := client.NewClient(cfg)
	
	// Test parameters
	numBatches := 20
	symbolsPerBatch := 50
	
	// Generate symbol batches
	batches := make([][]string, numBatches)
	for i := 0; i < numBatches; i++ {
		batches[i] = generateTestSymbols(symbolsPerBatch)
	}
	
	// Track results
	var wg sync.WaitGroup
	var successCount, failCount int64
	var totalContracts int64
	
	// Execute concurrent batch scans
	startTime := time.Now()
	
	for i, batch := range batches {
		wg.Add(1)
		go func(batchNum int, symbols []string) {
			defer wg.Done()
			
			results, err := c.ScanMultiple(ctx, symbols, nil)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				t.Logf("Batch %d failed: %v", batchNum, err)
				return
			}
			
			atomic.AddInt64(&successCount, 1)
			
			// Count total contracts
			for _, result := range results {
				atomic.AddInt64(&totalContracts, int64(result.TotalContracts))
			}
		}(i, batch)
		
		// Control request rate
		time.Sleep(100 * time.Millisecond)
	}
	
	wg.Wait()
	elapsed := time.Since(startTime)
	
	// Verify results
	t.Logf("Batch scan results:")
	t.Logf("  Total batches: %d", numBatches)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", failCount)
	t.Logf("  Total contracts processed: %d", totalContracts)
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Contracts/second: %.2f", float64(totalContracts)/elapsed.Seconds())
	
	// Assertions
	assert.Greater(t, successCount, int64(numBatches*80/100)) // At least 80% success
	assert.Greater(t, totalContracts, int64(10000)) // At least 10k contracts
}

// TestLoadWebSocketConnections tests WebSocket connection limits
func TestLoadWebSocketConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	ctx := context.Background()
	server, baseURL := startTestServer(t)
	defer server.Stop(ctx)
	
	// Test parameters
	targetConnections := 100
	symbols := generateTestSymbols(50)
	
	// Create connections
	connections := make([]*client.StreamingClient, 0, targetConnections)
	var connMutex sync.Mutex
	var successCount, failCount int64
	
	var wg sync.WaitGroup
	for i := 0; i < targetConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			cfg := client.Config{
				BaseURL: baseURL,
				Timeout: 30 * time.Second,
			}
			c := client.NewClient(cfg)
			
			stream, err := c.Connect(ctx)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}
			
			connMutex.Lock()
			connections = append(connections, stream)
			connMutex.Unlock()
			
			// Subscribe to random symbols
			subSymbols := make([]string, 5)
			for j := 0; j < 5; j++ {
				subSymbols[j] = symbols[rand.Intn(len(symbols))]
			}
			
			if err := stream.Subscribe(subSymbols, nil); err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}
			
			atomic.AddInt64(&successCount, 1)
		}(i)
		
		// Stagger connections
		time.Sleep(10 * time.Millisecond)
	}
	
	wg.Wait()
	
	// Keep connections alive for a bit
	time.Sleep(10 * time.Second)
	
	// Clean up
	connMutex.Lock()
	for _, conn := range connections {
		conn.Close()
	}
	connMutex.Unlock()
	
	// Results
	t.Logf("WebSocket connection test:")
	t.Logf("  Target connections: %d", targetConnections)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", failCount)
	
	assert.Greater(t, successCount, int64(targetConnections*80/100)) // At least 80% success
}

// runLoadTest executes the load test
func runLoadTest(t *testing.T, clients []*client.Client, symbols []string, config LoadTestConfig) *LoadTestResults {
	results := &LoadTestResults{
		StartTime:  time.Now(),
		MinLatency: int64(^uint64(0) >> 1), // Max int64
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	defer cancel()
	
	// Rate limiter
	ticker := time.NewTicker(time.Second / time.Duration(config.RequestsPerSec))
	defer ticker.Stop()
	
	// Worker pool
	var wg sync.WaitGroup
	requestCh := make(chan int, config.RequestsPerSec)
	
	// Start workers
	for i := 0; i < config.NumClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			client := clients[clientID]
			
			for {
				select {
				case <-ctx.Done():
					return
				case <-requestCh:
					// Random symbol
					symbol := symbols[rand.Intn(len(symbols))]
					
					// Time the request
					start := time.Now()
					result, err := client.ScanSymbol(ctx, symbol, nil)
					latency := time.Since(start).Milliseconds()
					
					// Update results
					atomic.AddInt64(&results.TotalRequests, 1)
					atomic.AddInt64(&results.TotalLatency, latency)
					
					if err != nil {
						atomic.AddInt64(&results.FailedRequests, 1)
					} else {
						atomic.AddInt64(&results.SuccessfulRequests, 1)
						atomic.AddInt64(&results.ContractsProcessed, int64(result.TotalContracts))
						
						// Update min/max latency
						for {
							oldMax := atomic.LoadInt64(&results.MaxLatency)
							if latency <= oldMax || atomic.CompareAndSwapInt64(&results.MaxLatency, oldMax, latency) {
								break
							}
						}
						
						for {
							oldMin := atomic.LoadInt64(&results.MinLatency)
							if latency >= oldMin || atomic.CompareAndSwapInt64(&results.MinLatency, oldMin, latency) {
								break
							}
						}
					}
				}
			}
		}(i)
	}
	
	// Generate load
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(requestCh)
				return
			case <-ticker.C:
				select {
				case requestCh <- 1:
				default:
					// Channel full, skip this tick
				}
			}
		}
	}()
	
	// Progress reporting
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				total := atomic.LoadInt64(&results.TotalRequests)
				success := atomic.LoadInt64(&results.SuccessfulRequests)
				contracts := atomic.LoadInt64(&results.ContractsProcessed)
				
				t.Logf("Progress: %d requests, %d successful, %d contracts processed",
					total, success, contracts)
			}
		}
	}()
	
	// Wait for completion
	wg.Wait()
	results.EndTime = time.Now()
	
	return results
}

// analyzeResults analyzes and reports load test results
func analyzeResults(t *testing.T, results *LoadTestResults, config LoadTestConfig) {
	duration := results.EndTime.Sub(results.StartTime)
	avgLatency := float64(results.TotalLatency) / float64(results.SuccessfulRequests)
	successRate := float64(results.SuccessfulRequests) / float64(results.TotalRequests) * 100
	requestsPerSec := float64(results.TotalRequests) / duration.Seconds()
	contractsPerSec := float64(results.ContractsProcessed) / duration.Seconds()
	
	t.Logf("\n=== Load Test Results ===")
	t.Logf("Duration: %v", duration)
	t.Logf("Total Requests: %d", results.TotalRequests)
	t.Logf("Successful: %d (%.2f%%)", results.SuccessfulRequests, successRate)
	t.Logf("Failed: %d", results.FailedRequests)
	t.Logf("Requests/sec: %.2f", requestsPerSec)
	t.Logf("Contracts Processed: %d", results.ContractsProcessed)
	t.Logf("Contracts/sec: %.2f", contractsPerSec)
	t.Logf("Latency - Avg: %.2fms, Min: %dms, Max: %dms",
		avgLatency, results.MinLatency, results.MaxLatency)
	
	// Performance assertions
	assert.Greater(t, successRate, 95.0, "Success rate should be > 95%")
	assert.Greater(t, results.ContractsProcessed, int64(10000), "Should process > 10k contracts")
	assert.Less(t, avgLatency, 1000.0, "Average latency should be < 1 second")
	
	// Calculate percentiles
	t.Logf("\n=== Performance Metrics ===")
	t.Logf("Theoretical max contracts: %d", config.NumSymbols*config.ContractsPerScan)
	t.Logf("Efficiency: %.2f%%", float64(results.ContractsProcessed)/float64(config.NumSymbols*config.ContractsPerScan)*100)
}

// generateTestSymbols generates test stock symbols
func generateTestSymbols(count int) []string {
	// Common prefixes and suffixes for realistic symbols
	prefixes := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	
	symbols := make([]string, count)
	used := make(map[string]bool)
	
	for i := 0; i < count; i++ {
		var symbol string
		for {
			// Generate 3-4 character symbol
			length := 3 + rand.Intn(2)
			symbol = ""
			for j := 0; j < length; j++ {
				symbol += prefixes[rand.Intn(len(prefixes))]
			}
			
			if !used[symbol] {
				used[symbol] = true
				break
			}
		}
		symbols[i] = symbol
	}
	
	// Add some real symbols for realism
	realSymbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "JPM", "BAC", "WMT"}
	for i := 0; i < len(realSymbols) && i < count; i++ {
		symbols[i] = realSymbols[i]
	}
	
	return symbols
}

// BenchmarkLoadTest runs a benchmark version of the load test
func BenchmarkLoadTest(b *testing.B) {
	ctx := context.Background()
	server, baseURL := startTestServer(nil)
	defer server.Stop(ctx)
	
	// Create client pool
	numClients := 10
	clients := make([]*client.Client, numClients)
	for i := 0; i < numClients; i++ {
		cfg := client.Config{
			BaseURL: baseURL,
			Timeout: 30 * time.Second,
		}
		clients[i] = client.NewClient(cfg)
	}
	
	symbols := generateTestSymbols(50)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		clientIdx := 0
		for pb.Next() {
			client := clients[clientIdx%numClients]
			symbol := symbols[rand.Intn(len(symbols))]
			
			client.ScanSymbol(ctx, symbol, nil)
			clientIdx++
		}
	})
}