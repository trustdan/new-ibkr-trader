package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/streaming"
)

func main() {
	// Configure client
	config := streaming.ClientConfig{
		URL:            "ws://localhost:8080/scan/stream",
		ReconnectDelay: 1 * time.Second,
		OnConnect: func() {
			log.Println("Connected to scanner stream")
		},
		OnDisconnect: func(err error) {
			log.Printf("Disconnected: %v", err)
		},
		OnMessage: func(msg streaming.Message) {
			log.Printf("Received: %s", msg.Type)
		},
		OnError: func(err error) {
			log.Printf("Error: %v", err)
		},
	}
	
	// Create client
	client := streaming.NewWebSocketClient(config)
	
	// Create subscription manager
	manager := streaming.NewSubscriptionManager(client)
	
	// Connect
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	
	// Add subscription for specific symbols
	err := manager.AddSubscription("main", map[string]interface{}{
		"symbols": []string{"SPY", "QQQ", "IWM"},
		"filters": map[string]interface{}{
			"min_delta": 0.20,
			"max_delta": 0.40,
			"min_dte":   30,
			"max_dte":   60,
		},
	}, func(update streaming.ScanUpdate) {
		// Handle scan updates
		fmt.Printf("\n=== Scan Update ===\n")
		fmt.Printf("Symbol: %s\n", update.Symbol)
		fmt.Printf("Type: %s\n", update.UpdateType)
		fmt.Printf("Spreads: %d\n", len(update.Spreads))
		
		// Display top spreads
		for i, spread := range update.Spreads {
			if i >= 3 {
				break // Show only top 3
			}
			
			fmt.Printf("\nSpread %d:\n", i+1)
			fmt.Printf("  Short: %.0f @ %.2f\n", spread.ShortLeg.Strike, spread.ShortLeg.Bid)
			fmt.Printf("  Long:  %.0f @ %.2f\n", spread.LongLeg.Strike, spread.LongLeg.Ask)
			fmt.Printf("  Credit: %.2f\n", spread.Credit)
			fmt.Printf("  PoP: %.1f%%\n", spread.ProbOfProfit*100)
			fmt.Printf("  Score: %.2f\n", spread.Score)
		}
		
		if meta, ok := update.Metadata["total_found"]; ok {
			fmt.Printf("\nTotal found: %v\n", meta)
		}
		if meta, ok := update.Metadata["filtered"]; ok {
			fmt.Printf("After filters: %v\n", meta)
		}
	})
	
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	
	log.Println("Streaming client started. Press Ctrl+C to exit.")
	
	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	
	<-sigChan
	
	log.Println("Shutting down...")
	client.Disconnect()
}