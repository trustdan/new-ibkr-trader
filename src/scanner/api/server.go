package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/service"
)

// Server handles HTTP API requests for the scanner
type Server struct {
	scanner  *service.Scanner
	router   *mux.Router
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

// NewServer creates a new API server
func NewServer(scanner *service.Scanner) *Server {
	s := &Server{
		scanner: scanner,
		router:  mux.NewRouter(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
	
	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// Scanning endpoints
	s.router.HandleFunc("/scan/{symbol}", s.handleScanSymbol).Methods("POST")
	s.router.HandleFunc("/scan/multiple", s.handleScanMultiple).Methods("POST")
	
	// Filter management
	s.router.HandleFunc("/filters", s.handleGetFilters).Methods("GET")
	s.router.HandleFunc("/filters", s.handleUpdateFilters).Methods("PUT")
	s.router.HandleFunc("/filters/presets", s.handleGetPresets).Methods("GET")
	s.router.HandleFunc("/filters/presets", s.handleSavePreset).Methods("POST")
	
	// WebSocket for real-time updates
	s.router.HandleFunc("/ws", s.handleWebSocket)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// handleHealth returns service health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "scanner",
	}
	json.NewEncoder(w).Encode(response)
}

// handleScanSymbol scans a single symbol
func (s *Server) handleScanSymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	// Parse request body for filter overrides
	var filterConfig filters.FilterConfig
	if err := json.NewDecoder(r.Body).Decode(&filterConfig); err != nil {
		// Use default filters if none provided
	}
	
	// Perform scan
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	result, err := s.scanner.ScanSymbol(ctx, symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Broadcast to WebSocket clients
	s.broadcastResult(result)
	
	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleScanMultiple scans multiple symbols
func (s *Server) handleScanMultiple(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Symbols []string             `json:"symbols"`
		Filters filters.FilterConfig `json:"filters"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Perform scans
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	
	results, err := s.scanner.ScanMultiple(ctx, request.Symbols)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Broadcast each result
	for _, result := range results {
		s.broadcastResult(result)
	}
	
	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleGetFilters returns current filter configuration
func (s *Server) handleGetFilters(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement filter retrieval
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "not_implemented"})
}

// handleUpdateFilters updates filter configuration
func (s *Server) handleUpdateFilters(w http.ResponseWriter, r *http.Request) {
	var filterConfig filters.FilterConfig
	if err := json.NewDecoder(r.Body).Decode(&filterConfig); err != nil {
		http.Error(w, "Invalid filter configuration", http.StatusBadRequest)
		return
	}
	
	// TODO: Update scanner filters
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// handleGetPresets returns saved filter presets
func (s *Server) handleGetPresets(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement preset management
	presets := []map[string]interface{}{
		{
			"name": "Conservative",
			"filters": map[string]interface{}{
				"delta": map[string]float64{"min": 0.25, "max": 0.35},
				"dte":   map[string]int{"min": 30, "max": 60},
			},
		},
		{
			"name": "Aggressive",
			"filters": map[string]interface{}{
				"delta": map[string]float64{"min": 0.15, "max": 0.25},
				"dte":   map[string]int{"min": 15, "max": 30},
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presets)
}

// handleSavePreset saves a new filter preset
func (s *Server) handleSavePreset(w http.ResponseWriter, r *http.Request) {
	var preset struct {
		Name    string               `json:"name"`
		Filters filters.FilterConfig `json:"filters"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&preset); err != nil {
		http.Error(w, "Invalid preset data", http.StatusBadRequest)
		return
	}
	
	// TODO: Save preset
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved", "name": preset.Name})
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	
	// Register client
	s.clients[conn] = true
	defer delete(s.clients, conn)
	
	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// broadcastResult sends scan results to all WebSocket clients
func (s *Server) broadcastResult(result interface{}) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "scan_result",
		"payload": result,
	})
	if err != nil {
		return
	}
	
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
}