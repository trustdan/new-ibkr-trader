package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/models"
)

// healthCheck returns service health status
func (api *API) healthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).Seconds(),
	}
	
	api.sendJSON(w, http.StatusOK, health)
}

// serviceInfo returns detailed service information
func (api *API) serviceInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service":     "IBKR Scanner API",
		"version":     "1.0.0",
		"api_version": "v1",
		"build_date":  buildDate,
		"commit":      gitCommit,
		"features": map[string]bool{
			"streaming":   true,
			"analytics":   true,
			"history":     true,
			"metrics":     true,
			"batch_scan":  true,
		},
		"limits": map[string]int{
			"max_batch_size":        100,
			"max_history_days":      30,
			"websocket_connections": 1000,
		},
	}
	
	api.sendJSON(w, http.StatusOK, info)
}

// scanSymbol scans a single symbol
func (api *API) scanSymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	if symbol == "" {
		api.sendError(w, http.StatusBadRequest, "Symbol is required")
		return
	}
	
	// Get filter overrides from query params
	filterOverrides := parseFilterQuery(r)
	
	// Perform scan
	result, err := api.scanner.ScanSymbol(r.Context(), symbol, filterOverrides)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Scan failed: %v", err))
		return
	}
	
	// Record in history
	api.history.Record(result)
	
	// Send to analytics
	api.analytics.ProcessScanResult(result)
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      result,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// scanSymbols scans multiple symbols
func (api *API) scanSymbols(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Symbols []string               `json:"symbols"`
		Filters map[string]interface{} `json:"filters,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if len(request.Symbols) == 0 {
		api.sendError(w, http.StatusBadRequest, "At least one symbol is required")
		return
	}
	
	if len(request.Symbols) > 100 {
		api.sendError(w, http.StatusBadRequest, "Maximum 100 symbols allowed per request")
		return
	}
	
	// Scan all symbols
	results := make([]*models.ScanResult, 0, len(request.Symbols))
	for _, symbol := range request.Symbols {
		result, err := api.scanner.ScanSymbol(r.Context(), symbol, request.Filters)
		if err != nil {
			// Log error but continue with other symbols
			continue
		}
		
		results = append(results, result)
		
		// Record and analyze
		api.history.Record(result)
		api.analytics.ProcessScanResult(result)
	}
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      results,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// batchScan performs batch scanning with progress updates
func (api *API) batchScan(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Symbols  []string               `json:"symbols"`
		Filters  map[string]interface{} `json:"filters,omitempty"`
		Parallel int                    `json:"parallel,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if len(request.Symbols) == 0 {
		api.sendError(w, http.StatusBadRequest, "At least one symbol is required")
		return
	}
	
	// Set default parallelism
	if request.Parallel == 0 {
		request.Parallel = 10
	}
	
	// Create batch job
	jobID := api.scanner.CreateBatchJob(request.Symbols, request.Filters, request.Parallel)
	
	api.sendJSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":   jobID,
		"status":   "processing",
		"symbols":  len(request.Symbols),
		"progress": fmt.Sprintf("/api/v1/jobs/%s", jobID),
	})
}

// getFilters returns current filter configuration
func (api *API) getFilters(w http.ResponseWriter, r *http.Request) {
	config := api.scanner.GetFilterConfig()
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      config,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// updateFilters updates filter configuration
func (api *API) updateFilters(w http.ResponseWriter, r *http.Request) {
	var config filters.FilterConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid filter configuration")
		return
	}
	
	// Validate filters
	if err := config.Validate(); err != nil {
		api.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid filters: %v", err))
		return
	}
	
	// Update scanner filters
	if err := api.scanner.UpdateFilters(config); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update filters: %v", err))
		return
	}
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      config,
		Status:    "updated",
		Timestamp: time.Now().Unix(),
	})
}

// validateFilters validates filter configuration without applying
func (api *API) validateFilters(w http.ResponseWriter, r *http.Request) {
	var config filters.FilterConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid filter configuration")
		return
	}
	
	// Validate filters
	validationResult := config.ValidateDetailed()
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      validationResult,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// getPresets returns saved filter presets
func (api *API) getPresets(w http.ResponseWriter, r *http.Request) {
	presets := api.filterCache.GetAll()
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      presets,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// createPreset creates a new filter preset
func (api *API) createPreset(w http.ResponseWriter, r *http.Request) {
	var preset struct {
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Filters     filters.FilterConfig `json:"filters"`
		Tags        []string             `json:"tags,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&preset); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid preset data")
		return
	}
	
	if preset.Name == "" {
		api.sendError(w, http.StatusBadRequest, "Preset name is required")
		return
	}
	
	// Validate filters
	if err := preset.Filters.Validate(); err != nil {
		api.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid filters: %v", err))
		return
	}
	
	// Save preset
	id := api.filterCache.Save(preset.Name, preset.Description, preset.Filters, preset.Tags)
	
	api.sendJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        id,
		"name":      preset.Name,
		"status":    "created",
		"timestamp": time.Now().Unix(),
	})
}

// getPreset returns a specific preset
func (api *API) getPreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	preset, exists := api.filterCache.Get(id)
	if !exists {
		api.sendError(w, http.StatusNotFound, "Preset not found")
		return
	}
	
	api.sendJSON(w, http.StatusOK, SuccessResponse{
		Data:      preset,
		Status:    "success",
		Timestamp: time.Now().Unix(),
	})
}

// updatePreset updates an existing preset
func (api *API) updatePreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	var update struct {
		Name        string               `json:"name,omitempty"`
		Description string               `json:"description,omitempty"`
		Filters     filters.FilterConfig `json:"filters,omitempty"`
		Tags        []string             `json:"tags,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid update data")
		return
	}
	
	// Update preset
	if err := api.filterCache.Update(id, update); err != nil {
		api.sendError(w, http.StatusNotFound, "Preset not found")
		return
	}
	
	api.sendJSON(w, http.StatusOK, map[string]interface{}{
		"id":        id,
		"status":    "updated",
		"timestamp": time.Now().Unix(),
	})
}

// deletePreset deletes a preset
func (api *API) deletePreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if err := api.filterCache.Delete(id); err != nil {
		api.sendError(w, http.StatusNotFound, "Preset not found")
		return
	}
	
	api.sendJSON(w, http.StatusOK, map[string]interface{}{
		"id":        id,
		"status":    "deleted",
		"timestamp": time.Now().Unix(),
	})
}

// Helper functions
func parseFilterQuery(r *http.Request) map[string]interface{} {
	filters := make(map[string]interface{})
	
	// Parse common filter parameters from query
	if deltaMin := r.URL.Query().Get("delta_min"); deltaMin != "" {
		if val, err := strconv.ParseFloat(deltaMin, 64); err == nil {
			if filters["delta"] == nil {
				filters["delta"] = make(map[string]float64)
			}
			filters["delta"].(map[string]float64)["min"] = val
		}
	}
	
	if deltaMax := r.URL.Query().Get("delta_max"); deltaMax != "" {
		if val, err := strconv.ParseFloat(deltaMax, 64); err == nil {
			if filters["delta"] == nil {
				filters["delta"] = make(map[string]float64)
			}
			filters["delta"].(map[string]float64)["max"] = val
		}
	}
	
	if dteMin := r.URL.Query().Get("dte_min"); dteMin != "" {
		if val, err := strconv.Atoi(dteMin); err == nil {
			if filters["dte"] == nil {
				filters["dte"] = make(map[string]int)
			}
			filters["dte"].(map[string]int)["min"] = val
		}
	}
	
	if dteMax := r.URL.Query().Get("dte_max"); dteMax != "" {
		if val, err := strconv.Atoi(dteMax); err == nil {
			if filters["dte"] == nil {
				filters["dte"] = make(map[string]int)
			}
			filters["dte"].(map[string]int)["max"] = val
		}
	}
	
	return filters
}

// Package variables
var (
	startTime = time.Now()
	buildDate = "unknown"
	gitCommit = "unknown"
)