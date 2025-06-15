package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/ibkr-automation/scanner/internal/scanner"
	"github.com/ibkr-automation/scanner/pkg/models"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
	symbol     = flag.String("symbol", "SPY", "symbol to scan")
	iterations = flag.Int("iterations", 100, "number of scan iterations")
	concurrent = flag.Int("concurrent", 1, "number of concurrent scans")
)

func main() {
	flag.Parse()

	// Setup CPU profiling if requested
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	// Create scanner
	c := cache.New(5*time.Minute, 10*time.Minute)
	s := scanner.New(c, sugar)

	// Create test request
	req := &models.ScanRequest{
		Symbol: *symbol,
		Filters: []models.FilterConfig{
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
					"min_volume":        100,
					"min_open_interest": 500,
					"max_bid_ask_spread": 0.10,
				},
			},
		},
		Limit: 50,
	}

	// Run benchmark
	fmt.Printf("Running %d iterations with %d concurrent workers...\n", *iterations, *concurrent)
	
	start := time.Now()
	results := runBenchmark(s, req, *iterations, *concurrent)
	duration := time.Since(start)

	// Print results
	printResults(results, duration)

	// Memory profiling if requested
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}

type benchmarkResult struct {
	Duration time.Duration
	Spreads  int
	Error    error
}

func runBenchmark(s *scanner.Scanner, req *models.ScanRequest, iterations, workers int) []benchmarkResult {
	results := make([]benchmarkResult, iterations)
	
	// Create work channel
	work := make(chan int, iterations)
	for i := 0; i < iterations; i++ {
		work <- i
	}
	close(work)

	// Create result channel
	resultCh := make(chan struct {
		idx    int
		result benchmarkResult
	}, iterations)

	// Start workers
	for w := 0; w < workers; w++ {
		go func() {
			for idx := range work {
				start := time.Now()
				resp, err := s.ScanOptions(context.Background(), req)
				
				var spreads int
				if resp != nil {
					spreads = len(resp.Options)
				}
				
				resultCh <- struct {
					idx    int
					result benchmarkResult
				}{
					idx: idx,
					result: benchmarkResult{
						Duration: time.Since(start),
						Spreads:  spreads,
						Error:    err,
					},
				}
			}
		}()
	}

	// Collect results
	for i := 0; i < iterations; i++ {
		r := <-resultCh
		results[r.idx] = r.result
	}

	return results
}

func printResults(results []benchmarkResult, totalDuration time.Duration) {
	var (
		totalSpreads   int
		totalErrors    int
		totalScanTime  time.Duration
		minDuration    = time.Hour
		maxDuration    time.Duration
	)

	for _, r := range results {
		if r.Error != nil {
			totalErrors++
			continue
		}
		
		totalSpreads += r.Spreads
		totalScanTime += r.Duration
		
		if r.Duration < minDuration {
			minDuration = r.Duration
		}
		if r.Duration > maxDuration {
			maxDuration = r.Duration
		}
	}

	successfulScans := len(results) - totalErrors
	avgDuration := totalScanTime / time.Duration(successfulScans)
	scansPerSecond := float64(successfulScans) / totalDuration.Seconds()

	fmt.Println("\n=== Benchmark Results ===")
	fmt.Printf("Total iterations: %d\n", len(results))
	fmt.Printf("Successful scans: %d\n", successfulScans)
	fmt.Printf("Failed scans: %d\n", totalErrors)
	fmt.Printf("Total duration: %s\n", totalDuration)
	fmt.Printf("Scans per second: %.2f\n", scansPerSecond)
	fmt.Printf("\nPer-scan statistics:\n")
	fmt.Printf("  Average duration: %s\n", avgDuration)
	fmt.Printf("  Min duration: %s\n", minDuration)
	fmt.Printf("  Max duration: %s\n", maxDuration)
	fmt.Printf("  Average spreads found: %.2f\n", float64(totalSpreads)/float64(successfulScans))
	
	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\nMemory statistics:\n")
	fmt.Printf("  Alloc: %d MB\n", m.Alloc/1024/1024)
	fmt.Printf("  TotalAlloc: %d MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("  Sys: %d MB\n", m.Sys/1024/1024)
	fmt.Printf("  NumGC: %d\n", m.NumGC)
}