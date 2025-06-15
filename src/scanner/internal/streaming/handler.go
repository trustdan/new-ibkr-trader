package streaming

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/rs/zerolog/log"
)

// StreamingHandler handles HTTP endpoints for streaming functionality
type StreamingHandler struct {
	wsServer         *WebSocketServer
	streamingScanner *StreamingScanner
}

// NewStreamingHandler creates a new streaming handler
func NewStreamingHandler(wsServer *WebSocketServer, scanner *StreamingScanner) *StreamingHandler {
	return &StreamingHandler{
		wsServer:         wsServer,
		streamingScanner: scanner,
	}
}

// RegisterRoutes registers streaming routes
func (h *StreamingHandler) RegisterRoutes(router *gin.Engine) {
	stream := router.Group("/scan")
	{
		// WebSocket endpoint
		stream.GET("/stream", h.handleWebSocket)
		
		// Control endpoints
		stream.POST("/stream/start", h.handleStartStreaming)
		stream.POST("/stream/stop", h.handleStopStreaming)
		stream.GET("/stream/status", h.handleStreamStatus)
		
		// Ad-hoc scan request
		stream.POST("/stream/request", h.handleScanRequest)
	}
}

// handleWebSocket upgrades the connection to WebSocket
func (h *StreamingHandler) handleWebSocket(c *gin.Context) {
	h.wsServer.HandleWebSocket(c.Writer, c.Request)
}

// handleStartStreaming starts the streaming scanner
func (h *StreamingHandler) handleStartStreaming(c *gin.Context) {
	var req struct {
		Symbols []string `json:"symbols" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// Start streaming scanner
	ctx := context.Background()
	if err := h.streamingScanner.Start(ctx, req.Symbols); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "started",
		"symbols": req.Symbols,
	})
}

// handleStopStreaming stops the streaming scanner
func (h *StreamingHandler) handleStopStreaming(c *gin.Context) {
	h.streamingScanner.Stop()
	
	c.JSON(http.StatusOK, gin.H{
		"status": "stopped",
	})
}

// handleStreamStatus returns streaming status
func (h *StreamingHandler) handleStreamStatus(c *gin.Context) {
	stats := h.streamingScanner.GetStats()
	
	c.JSON(http.StatusOK, stats)
}

// handleScanRequest handles ad-hoc scan requests
func (h *StreamingHandler) handleScanRequest(c *gin.Context) {
	var req struct {
		Symbols      []string               `json:"symbols" binding:"required"`
		FilterConfig filters.FilterConfig   `json:"filter_config,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// Submit scan request
	responseChan := h.streamingScanner.SubmitScanRequest(req.Symbols, req.FilterConfig)
	
	// Wait for response with timeout
	select {
	case response := <-responseChan:
		if response.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": response.Error.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"results": response.Results,
		})
		
	case <-time.After(30 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{
			"error": "scan timeout",
		})
	}
}

// StreamingMiddleware provides middleware for streaming endpoints
type StreamingMiddleware struct {
	wsServer *WebSocketServer
}

// NewStreamingMiddleware creates streaming middleware
func NewStreamingMiddleware(wsServer *WebSocketServer) *StreamingMiddleware {
	return &StreamingMiddleware{
		wsServer: wsServer,
	}
}

// NotifyClients is middleware that notifies WebSocket clients of API changes
func (m *StreamingMiddleware) NotifyClients() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:          []byte{},
		}
		c.Writer = writer
		
		// Process request
		c.Next()
		
		// If successful modification, notify clients
		if c.Request.Method != "GET" && writer.statusCode >= 200 && writer.statusCode < 300 {
			// Parse response to determine what changed
			var response map[string]interface{}
			if err := json.Unmarshal(writer.body, &response); err == nil {
				// Broadcast status update
				m.wsServer.BroadcastStatus(StatusUpdate{
					Status:  "api_update",
					Message: c.Request.URL.Path,
				})
			}
		}
	}
}

// responseWriter captures response for middleware
type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// StreamingService provides high-level streaming functionality
type StreamingService struct {
	wsServer         *WebSocketServer
	streamingScanner *StreamingScanner
	handler          *StreamingHandler
}

// NewStreamingService creates a complete streaming service
func NewStreamingService(scanner *StreamingScanner, filterChain *filters.AdvancedFilterChain) *StreamingService {
	wsServer := NewWebSocketServer()
	streamingScanner := NewStreamingScanner(scanner, wsServer, filterChain)
	handler := NewStreamingHandler(wsServer, streamingScanner)
	
	return &StreamingService{
		wsServer:         wsServer,
		streamingScanner: streamingScanner,
		handler:          handler,
	}
}

// Start starts all streaming components
func (s *StreamingService) Start(ctx context.Context) error {
	// Start WebSocket server
	s.wsServer.Start(ctx)
	
	log.Info().Msg("Streaming service started")
	return nil
}

// Stop stops all streaming components
func (s *StreamingService) Stop() {
	s.streamingScanner.Stop()
	log.Info().Msg("Streaming service stopped")
}

// RegisterRoutes registers all streaming routes
func (s *StreamingService) RegisterRoutes(router *gin.Engine) {
	s.handler.RegisterRoutes(router)
	
	// Add middleware
	middleware := NewStreamingMiddleware(s.wsServer)
	router.Use(middleware.NotifyClients())
}