package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/ibkr-trader/scanner/api"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/ibkr-trader/scanner/internal/service"
)

// Config holds application configuration
type Config struct {
	Port            string
	PythonAPIURL    string
	ConcurrentScans int
	CacheTTL        time.Duration
}

func main() {
	// Parse command line flags
	config := parseFlags()
	
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Starting IBKR Scanner Service on port %s", config.Port)
	
	// Create data provider (connects to Python container)
	dataProvider := NewPythonDataProvider(config.PythonAPIURL)
	
	// Create default filter configuration
	filterConfig := filters.FilterConfig{
		Delta: &filters.DeltaFilter{
			MinDelta: 0.25,
			MaxDelta: 0.35,
			Absolute: true,
		},
		DTE: &filters.DTEFilter{
			MinDTE: 30,
			MaxDTE: 60,
		},
		Liquidity: &filters.LiquidityFilter{
			MinVolume:       100,
			MinOpenInterest: 500,
			MaxBidAskSpread: 0.10,
		},
	}
	
	// Create scanner service
	scanner := service.NewScanner(dataProvider, filterConfig)
	
	// Create API server
	server := api.NewServer(scanner)
	
	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      server,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		log.Printf("Scanner API listening on port %s", config.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server stopped")
}

// parseFlags parses command line arguments
func parseFlags() Config {
	var config Config
	
	flag.StringVar(&config.Port, "port", "8081", "Port to listen on")
	flag.StringVar(&config.PythonAPIURL, "python-api", "http://python-service:8080", "Python API URL")
	flag.IntVar(&config.ConcurrentScans, "concurrent", 5, "Number of concurrent scans")
	flag.DurationVar(&config.CacheTTL, "cache-ttl", 5*time.Minute, "Cache TTL duration")
	
	flag.Parse()
	return config
}

// PythonDataProvider implements DataProvider by calling Python API
type PythonDataProvider struct {
	apiURL string
	client *http.Client
}

// NewPythonDataProvider creates a new Python API data provider
func NewPythonDataProvider(apiURL string) *PythonDataProvider {
	return &PythonDataProvider{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetOptionChain fetches option chain from Python API
func (p *PythonDataProvider) GetOptionChain(ctx context.Context, symbol string) ([]models.OptionContract, error) {
	// TODO: Implement actual API call to Python service
	// For now, return mock data
	log.Printf("Fetching option chain for %s", symbol)
	return []models.OptionContract{}, nil
}

// GetQuote fetches current quote from Python API
func (p *PythonDataProvider) GetQuote(ctx context.Context, symbol string) (float64, error) {
	// TODO: Implement actual API call to Python service
	log.Printf("Fetching quote for %s", symbol)
	return 100.0, nil
}