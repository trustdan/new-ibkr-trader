package v1

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// getPatterns returns detected trading patterns
func (api *API) getPatterns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")
	patternType := r.URL.Query().Get("type")
	
	// Get patterns from analytics engine
	patterns := api.analytics.GetPatterns(symbol, startTime, endTime, patternType)
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      patterns,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// getStatistics returns statistical analysis
func (api *API) getStatistics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	metric := r.URL.Query().Get("metric")
	period := r.URL.Query().Get("period")
	
	if period == "" {
		period = "24h"
	}
	
	// Get statistics
	stats := api.analytics.GetStatistics(symbol, metric, period)
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      stats,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// getPerformance returns performance metrics
func (api *API) getPerformance(w http.ResponseWriter, r *http.Request) {
	// Get performance metrics
	perf := api.analytics.GetPerformanceMetrics()
	
	// Add system metrics
	perf["api_metrics"] = map[string]interface{}{
		"requests_per_second": api.metrics.GetRequestRate(),
		"average_latency_ms":  api.metrics.GetAverageLatency(),
		"error_rate":          api.metrics.GetErrorRate(),
		"active_connections":  api.streamer.GetActiveConnections(),
	}
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      perf,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// exportAnalytics exports analytics data in various formats
func (api *API) exportAnalytics(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Format    string   `json:"format"`     // json, csv, excel
		StartDate string   `json:"start_date"`
		EndDate   string   `json:"end_date"`
		Symbols   []string `json:"symbols,omitempty"`
		Metrics   []string `json:"metrics,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate format
	if request.Format == "" {
		request.Format = "json"
	}
	
	if request.Format != "json" && request.Format != "csv" {
		api.sendError(w, http.StatusBadRequest, "Unsupported format. Use 'json' or 'csv'")
		return
	}
	
	// Get export data
	data := api.analytics.Export(request.StartDate, request.EndDate, request.Symbols, request.Metrics)
	
	switch request.Format {
	case "csv":
		api.sendCSV(w, data)
	default:
		api.sendJSON(w, http.StatusOK, SuccessResponse{
			Data:      data,
			Status:    "success",
			Timestamp: time.Now().Unix(),
		})
	}
}

// getHistory returns historical scan data
func (api *API) getHistory(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	// Parse filters
	symbol := r.URL.Query().Get("symbol")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	
	// Get history
	results, total := api.history.Query(symbol, startDate, endDate, page, pageSize)
	
	// Calculate pagination
	totalPages := (total + pageSize - 1) / pageSize
	
	api.sendJSON(w, http.StatusOK, PaginatedResponse{
		Data: results,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// getSymbolHistory returns history for a specific symbol
func (api *API) getSymbolHistory(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		api.sendError(w, http.StatusBadRequest, "Symbol is required")
		return
	}
	
	// Parse time range
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 7
	}
	
	// Get symbol history
	history := api.history.GetSymbolHistory(symbol, days)
	
	// Calculate statistics
	stats := api.analytics.CalculateSymbolStats(symbol, history)
	
	response := map[string]interface{}{
		"symbol":   symbol,
		"history":  history,
		"stats":    stats,
		"period":   fmt.Sprintf("%d days", days),
	}
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      response,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// clearHistory clears historical data
func (api *API) clearHistory(w http.ResponseWriter, r *http.Request) {
	// Parse parameters
	symbol := r.URL.Query().Get("symbol")
	before := r.URL.Query().Get("before")
	
	// Clear history
	count := api.history.Clear(symbol, before)
	
	api.sendJSON(w, http.StatusOK, map[string]interface{}{
		"cleared":   count,
		"status":    "success",
		"timestamp": time.Now().Unix(),
	})
}

// getMetrics returns current metrics
func (api *API) getMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := api.metrics.GetAll()
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      metrics,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// prometheusMetrics returns metrics in Prometheus format
func (api *API) prometheusMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	api.metrics.WritePrometheus(w)
}

// sendCSV sends data as CSV
func (api *API) sendCSV(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export_%d.csv", time.Now().Unix()))
	
	writer := csv.NewWriter(w)
	defer writer.Flush()
	
	// Convert data to CSV format
	// This is a simplified version - in production, handle different data types
	if records, ok := data.([][]string); ok {
		for _, record := range records {
			writer.Write(record)
		}
	}
}