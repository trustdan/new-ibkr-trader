package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	v1 "github.com/ibkr-trader/scanner/api/v1"
	"github.com/ibkr-trader/scanner/internal/analytics"
	"github.com/ibkr-trader/scanner/internal/history"
	"github.com/ibkr-trader/scanner/internal/metrics"
	"github.com/ibkr-trader/scanner/internal/service"
	"github.com/ibkr-trader/scanner/internal/streaming"
	"github.com/rs/zerolog/log"
)

// Server represents the API server
type Server struct {
	router       *mux.Router
	httpServer   *http.Server
	scanner      *service.Scanner
	streamer     *streaming.Manager
	analytics    *analytics.Engine
	history      *history.Store
	metrics      *metrics.Collector
	port         int
}

// Config holds server configuration
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHeaderBytes  int
}

// DefaultConfig returns default server configuration
func DefaultConfig() Config {
	return Config{
		Port:           8080,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

// NewServer creates a new API server
func NewServer(
	config Config,
	scanner *service.Scanner,
	streamer *streaming.Manager,
	analytics *analytics.Engine,
	history *history.Store,
	metrics *metrics.Collector,
) *Server {
	router := mux.NewRouter()
	
	s := &Server{
		router:    router,
		scanner:   scanner,
		streamer:  streamer,
		analytics: analytics,
		history:   history,
		metrics:   metrics,
		port:      config.Port,
	}
	
	// Setup routes
	s.setupRoutes()
	
	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		Handler:        router,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}
	
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Root health check
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
	
	// API v1
	apiV1 := v1.NewAPI(s.scanner, s.streamer, s.analytics, s.history, s.metrics)
	apiV1.RegisterRoutes(s.router)
	
	// Static files for API documentation (optional)
	s.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./static/docs/"))))
	
	// Catch-all 404 handler
	s.router.NotFoundHandler = http.HandlerFunc(s.handleNotFound)
}

// Start starts the API server
func (s *Server) Start() error {
	log.Info().
		Int("port", s.port).
		Msg("Starting API server")
	
	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
	
	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	
	log.Info().
		Str("address", fmt.Sprintf("http://localhost:%d", s.port)).
		Str("api_docs", fmt.Sprintf("http://localhost:%d/api/v1/openapi.json", s.port)).
		Msg("API server started successfully")
	
	return nil
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping API server")
	
	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	
	// Close streaming connections
	s.streamer.CloseAll()
	
	log.Info().Msg("API server stopped")
	return nil
}

// handleRoot handles root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "IBKR Scanner API",
		"status":  "online",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"health":  "/api/v1/health",
			"api_v1":  "/api/v1",
			"docs":    "/api/v1/openapi.json",
			"metrics": "/api/v1/metrics",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNotFound handles 404 errors
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	
	response := map[string]interface{}{
		"error":     "Not Found",
		"status":    404,
		"path":      r.URL.Path,
		"timestamp": time.Now().Unix(),
	}
	
	json.NewEncoder(w).Encode(response)
}