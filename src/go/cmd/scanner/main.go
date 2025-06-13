package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/ibkr-automation/scanner/internal/api"
	"github.com/ibkr-automation/scanner/internal/cache"
	"github.com/ibkr-automation/scanner/internal/scanner"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize logger
	logger, _ := zap.NewProduction()
	if os.Getenv("ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Info("🚀 Starting IBKR Go Scanner Service")

	// Initialize components
	cacheService := cache.New(5*time.Minute, 10*time.Minute)
	scannerService := scanner.New(cacheService, sugar)
	
	// Setup Gin router
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(api.LoggerMiddleware(sugar))
	router.Use(api.MetricsMiddleware())

	// Setup routes
	api.SetupRoutes(router, scannerService, sugar)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start server
	port := os.Getenv("SCANNER_PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Failed to start server: %v", err)
		}
	}()

	sugar.Infof("✅ Scanner service running on port %s", port)
	sugar.Info("🌊 Ready to scan the options waves!")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("🛑 Shutting down scanner service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorf("Server forced to shutdown: %v", err)
	}

	sugar.Info("👋 Scanner service shutdown complete")
}