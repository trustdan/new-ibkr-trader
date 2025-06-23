package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Simple health response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Service   string `json:"service"`
}

// Simple scan result
type ScanResult struct {
	Symbol    string  `json:"symbol"`
	Timestamp int64   `json:"timestamp"`
	Results   []Trade `json:"results"`
}

type Trade struct {
	Strike  float64 `json:"strike"`
	Expiry  string  `json:"expiry"`
	Delta   float64 `json:"delta"`
	Premium float64 `json:"premium"`
	Volume  int     `json:"volume"`
}

func main() {
	port := flag.String("port", "8081", "Port to listen on")
	pythonAPI := flag.String("python-api", "http://localhost:8080", "Python API URL")
	flag.Parse()

	log.Printf("Starting Simple Scanner on port %s", *port)
	log.Printf("Python API: %s", *pythonAPI)

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().Unix(),
			Service:   "scanner-simple",
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(response)
	})

	// Simple scan endpoint
	mux.HandleFunc("/api/v1/scan/", func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Path[len("/api/v1/scan/"):]
		if symbol == "" {
			symbol = "SPY"
		}

		// Mock scan result
		result := ScanResult{
			Symbol:    symbol,
			Timestamp: time.Now().Unix(),
			Results: []Trade{
				{
					Strike:  580.0,
					Expiry:  "2025-02-21",
					Delta:   0.30,
					Premium: 2.50,
					Volume:  150,
				},
				{
					Strike:  585.0,
					Expiry:  "2025-02-21",
					Delta:   0.25,
					Premium: 2.10,
					Volume:  200,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(result)
	})

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	// Start server
	go func() {
		log.Printf("Scanner listening on port %s", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}
