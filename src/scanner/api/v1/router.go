package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ibkr-trader/scanner/internal/analytics"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/history"
	"github.com/ibkr-trader/scanner/internal/metrics"
	"github.com/ibkr-trader/scanner/internal/service"
	"github.com/ibkr-trader/scanner/internal/streaming"
)

// API represents the v1 API
type API struct {
	scanner      *service.Scanner
	streamer     *streaming.Manager
	analytics    *analytics.Engine
	history      *history.Store
	metrics      *metrics.Collector
	filterCache  *filters.PresetCache
}

// NewAPI creates a new v1 API instance
func NewAPI(
	scanner *service.Scanner,
	streamer *streaming.Manager,
	analytics *analytics.Engine,
	history *history.Store,
	metrics *metrics.Collector,
) *API {
	return &API{
		scanner:      scanner,
		streamer:     streamer,
		analytics:    analytics,
		history:      history,
		metrics:      metrics,
		filterCache:  filters.NewPresetCache(),
	}
}

// RegisterRoutes registers all v1 API routes
func (api *API) RegisterRoutes(r *mux.Router) {
	// Apply API versioning prefix
	v1 := r.PathPrefix("/api/v1").Subrouter()
	
	// Middleware
	v1.Use(loggingMiddleware)
	v1.Use(corsMiddleware)
	v1.Use(metricsMiddleware(api.metrics))
	
	// Health and Info endpoints
	v1.HandleFunc("/health", api.healthCheck).Methods("GET", "OPTIONS")
	v1.HandleFunc("/info", api.serviceInfo).Methods("GET", "OPTIONS")
	
	// Scanner endpoints
	v1.HandleFunc("/scan", api.scanSymbols).Methods("POST", "OPTIONS")
	v1.HandleFunc("/scan/{symbol}", api.scanSymbol).Methods("GET", "OPTIONS")
	v1.HandleFunc("/scan/batch", api.batchScan).Methods("POST", "OPTIONS")
	
	// Filter management
	v1.HandleFunc("/filters", api.getFilters).Methods("GET", "OPTIONS")
	v1.HandleFunc("/filters", api.updateFilters).Methods("PUT", "OPTIONS")
	v1.HandleFunc("/filters/validate", api.validateFilters).Methods("POST", "OPTIONS")
	v1.HandleFunc("/filters/presets", api.getPresets).Methods("GET", "OPTIONS")
	v1.HandleFunc("/filters/presets", api.createPreset).Methods("POST", "OPTIONS")
	v1.HandleFunc("/filters/presets/{id}", api.getPreset).Methods("GET", "OPTIONS")
	v1.HandleFunc("/filters/presets/{id}", api.updatePreset).Methods("PUT", "OPTIONS")
	v1.HandleFunc("/filters/presets/{id}", api.deletePreset).Methods("DELETE", "OPTIONS")
	
	// Analytics endpoints
	v1.HandleFunc("/analytics/patterns", api.getPatterns).Methods("GET", "OPTIONS")
	v1.HandleFunc("/analytics/statistics", api.getStatistics).Methods("GET", "OPTIONS")
	v1.HandleFunc("/analytics/performance", api.getPerformance).Methods("GET", "OPTIONS")
	v1.HandleFunc("/analytics/export", api.exportAnalytics).Methods("POST", "OPTIONS")
	
	// History endpoints
	v1.HandleFunc("/history", api.getHistory).Methods("GET", "OPTIONS")
	v1.HandleFunc("/history/{symbol}", api.getSymbolHistory).Methods("GET", "OPTIONS")
	v1.HandleFunc("/history/clear", api.clearHistory).Methods("DELETE", "OPTIONS")
	
	// Metrics endpoints
	v1.HandleFunc("/metrics", api.getMetrics).Methods("GET", "OPTIONS")
	v1.HandleFunc("/metrics/prometheus", api.prometheusMetrics).Methods("GET", "OPTIONS")
	
	// WebSocket endpoint for streaming
	v1.HandleFunc("/ws", api.handleWebSocket)
	
	// OpenAPI documentation
	v1.HandleFunc("/openapi.json", api.openAPISpec).Methods("GET", "OPTIONS")
}

// Response helpers
func (api *API) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (api *API) sendError(w http.ResponseWriter, status int, message string) {
	api.sendJSON(w, status, map[string]interface{}{
		"error": message,
		"status": status,
		"timestamp": time.Now().Unix(),
	})
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Status    int    `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data      interface{} `json:"data"`
	Status    string      `json:"status"`
	Timestamp int64       `json:"timestamp"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Status     string      `json:"status"`
	Timestamp  int64       `json:"timestamp"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}