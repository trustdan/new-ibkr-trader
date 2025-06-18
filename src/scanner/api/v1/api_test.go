package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	v1 "github.com/ibkr-trader/scanner/api/v1"
	"github.com/ibkr-trader/scanner/internal/analytics"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/history"
	"github.com/ibkr-trader/scanner/internal/metrics"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/ibkr-trader/scanner/internal/service"
	"github.com/ibkr-trader/scanner/internal/streaming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuite holds test dependencies
type TestSuite struct {
	api      *v1.API
	router   *mux.Router
	scanner  *service.Scanner
	streamer *streaming.Manager
}

// setupTestSuite creates a test suite with mocked dependencies
func setupTestSuite(t *testing.T) *TestSuite {
	// Create mocked dependencies
	scanner := service.NewScanner(nil) // Mock scanner
	streamer := streaming.NewManager()
	analytics := analytics.NewEngine()
	history := history.NewStore()
	metrics := metrics.NewCollector()
	
	// Create API
	api := v1.NewAPI(scanner, streamer, analytics, history, metrics)
	
	// Create router
	router := mux.NewRouter()
	api.RegisterRoutes(router)
	
	return &TestSuite{
		api:      api,
		router:   router,
		scanner:  scanner,
		streamer: streamer,
	}
}

// TestHealthCheck tests the health endpoint
func TestHealthCheck(t *testing.T) {
	suite := setupTestSuite(t)
	
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "version")
	assert.Contains(t, response, "uptime")
}

// TestScanSingleSymbol tests scanning a single symbol
func TestScanSingleSymbol(t *testing.T) {
	suite := setupTestSuite(t)
	
	tests := []struct {
		name       string
		symbol     string
		query      string
		wantStatus int
	}{
		{
			name:       "valid symbol",
			symbol:     "AAPL",
			query:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "with filters",
			symbol:     "AAPL",
			query:      "?delta_min=0.20&delta_max=0.35&dte_min=30&dte_max=60",
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty symbol",
			symbol:     "",
			query:      "",
			wantStatus: http.StatusNotFound,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/v1/scan/%s%s", tt.symbol, tt.query)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
			
			if tt.wantStatus == http.StatusOK {
				var response v1.SuccessResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "success", response.status)
			}
		})
	}
}

// TestScanMultipleSymbols tests scanning multiple symbols
func TestScanMultipleSymbols(t *testing.T) {
	suite := setupTestSuite(t)
	
	tests := []struct {
		name       string
		request    interface{}
		wantStatus int
	}{
		{
			name: "valid request",
			request: map[string]interface{}{
				"symbols": []string{"AAPL", "MSFT", "GOOGL"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "with filters",
			request: map[string]interface{}{
				"symbols": []string{"AAPL", "MSFT"},
				"filters": map[string]interface{}{
					"delta": map[string]float64{"min": 0.20, "max": 0.35},
					"dte":   map[string]int{"min": 30, "max": 60},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "empty symbols",
			request: map[string]interface{}{
				"symbols": []string{},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "too many symbols",
			request: map[string]interface{}{
				"symbols": make([]string, 101), // 101 symbols
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid request",
			request:    "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/scan", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestFilterManagement tests filter endpoints
func TestFilterManagement(t *testing.T) {
	suite := setupTestSuite(t)
	
	t.Run("get filters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/filters", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response v1.SuccessResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
	})
	
	t.Run("update filters", func(t *testing.T) {
		filterConfig := filters.FilterConfig{
			Delta: &filters.DeltaFilter{Min: 0.25, Max: 0.35},
			DTE:   &filters.DTEFilter{Min: 30, Max: 60},
		}
		
		body, _ := json.Marshal(filterConfig)
		req := httptest.NewRequest("PUT", "/api/v1/filters", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("validate filters", func(t *testing.T) {
		filterConfig := filters.FilterConfig{
			Delta: &filters.DeltaFilter{Min: 0.25, Max: 0.35},
		}
		
		body, _ := json.Marshal(filterConfig)
		req := httptest.NewRequest("POST", "/api/v1/filters/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestPresetManagement tests preset endpoints
func TestPresetManagement(t *testing.T) {
	suite := setupTestSuite(t)
	
	var presetID string
	
	t.Run("get presets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/filters/presets", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response v1.SuccessResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
	})
	
	t.Run("create preset", func(t *testing.T) {
		preset := map[string]interface{}{
			"name":        "Test Preset",
			"description": "Test description",
			"filters": filters.FilterConfig{
				Delta: &filters.DeltaFilter{Min: 0.20, Max: 0.30},
			},
			"tags": []string{"test", "integration"},
		}
		
		body, _ := json.Marshal(preset)
		req := httptest.NewRequest("POST", "/api/v1/filters/presets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		
		presetID = response["id"].(string)
		assert.NotEmpty(t, presetID)
	})
	
	t.Run("get specific preset", func(t *testing.T) {
		if presetID == "" {
			t.Skip("No preset ID available")
		}
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/filters/presets/%s", presetID), nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("update preset", func(t *testing.T) {
		if presetID == "" {
			t.Skip("No preset ID available")
		}
		
		update := map[string]interface{}{
			"description": "Updated description",
		}
		
		body, _ := json.Marshal(update)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/filters/presets/%s", presetID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("delete preset", func(t *testing.T) {
		if presetID == "" {
			t.Skip("No preset ID available")
		}
		
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/filters/presets/%s", presetID), nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestWebSocket tests WebSocket functionality
func TestWebSocket(t *testing.T) {
	suite := setupTestSuite(t)
	
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert http:// to ws://
	wsURL := "ws" + server.URL[4:] + "/api/v1/ws"
	
	// Connect to WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()
	
	// Test welcome message
	var welcome map[string]interface{}
	err = ws.ReadJSON(&welcome)
	require.NoError(t, err)
	assert.Equal(t, "welcome", welcome["type"])
	
	// Test subscribe
	subscribe := map[string]interface{}{
		"type": "subscribe",
		"payload": map[string]interface{}{
			"symbols": []string{"AAPL", "MSFT"},
		},
	}
	err = ws.WriteJSON(subscribe)
	require.NoError(t, err)
	
	// Test ping/pong
	ping := map[string]interface{}{
		"type": "ping",
	}
	err = ws.WriteJSON(ping)
	require.NoError(t, err)
	
	// Read pong response
	var pong map[string]interface{}
	err = ws.ReadJSON(&pong)
	require.NoError(t, err)
	assert.Equal(t, "pong", pong["type"])
}

// TestAnalyticsEndpoints tests analytics endpoints
func TestAnalyticsEndpoints(t *testing.T) {
	suite := setupTestSuite(t)
	
	t.Run("get patterns", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/analytics/patterns?symbol=AAPL", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("get statistics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/analytics/statistics?period=24h", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("get performance", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/analytics/performance", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("export analytics", func(t *testing.T) {
		exportReq := map[string]interface{}{
			"format":     "json",
			"start_date": "2023-01-01",
			"end_date":   "2023-12-31",
			"symbols":    []string{"AAPL"},
		}
		
		body, _ := json.Marshal(exportReq)
		req := httptest.NewRequest("POST", "/api/v1/analytics/export", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestHistoryEndpoints tests history endpoints
func TestHistoryEndpoints(t *testing.T) {
	suite := setupTestSuite(t)
	
	t.Run("get history", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/history?page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response v1.PaginatedResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 10, response.Pagination.PageSize)
	})
	
	t.Run("get symbol history", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/history/AAPL?days=7", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("clear history", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/history/clear?symbol=AAPL", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestMetricsEndpoints tests metrics endpoints
func TestMetricsEndpoints(t *testing.T) {
	suite := setupTestSuite(t)
	
	t.Run("get metrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/metrics", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("get prometheus metrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/metrics/prometheus", nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain; version=0.0.4", w.Header().Get("Content-Type"))
	})
}

// TestCORSHeaders tests CORS functionality
func TestCORSHeaders(t *testing.T) {
	suite := setupTestSuite(t)
	
	req := httptest.NewRequest("OPTIONS", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	// This would test rate limiting if implemented
	t.Skip("Rate limiting test - implement when rate limiting is added")
}

// TestConcurrentRequests tests handling of concurrent requests
func TestConcurrentRequests(t *testing.T) {
	suite := setupTestSuite(t)
	
	// Number of concurrent requests
	numRequests := 50
	done := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			symbol := fmt.Sprintf("TEST%d", id%10)
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/scan/%s", symbol), nil)
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
		}(i)
	}
	
	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

// BenchmarkScanEndpoint benchmarks the scan endpoint
func BenchmarkScanEndpoint(b *testing.B) {
	suite := setupTestSuite(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/scan/AAPL", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
	}
}