package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ibkr-automation/scanner/internal/scanner"
	"github.com/ibkr-automation/scanner/pkg/models"
)

var startTime = time.Now()

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, scanner *scanner.Scanner, logger *zap.SugaredLogger) {
	// Health check
	router.GET("/health", healthHandler)
	
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Scanner endpoints
		v1.POST("/scan", scanHandler(scanner, logger))
		v1.GET("/stats", statsHandler(scanner))
		
		// Test endpoints
		v1.GET("/ping", pingHandler)
	}
}

// healthHandler returns service health status
func healthHandler(c *gin.Context) {
	health := models.HealthCheck{
		Status:    "healthy",
		Service:   "ibkr-go-scanner",
		Uptime:    time.Since(startTime).String(),
		Version:   "0.1.0",
		Timestamp: time.Now(),
	}
	
	c.JSON(http.StatusOK, health)
}

// scanHandler handles scan requests
func scanHandler(scanner *scanner.Scanner, logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ScanRequest
		
		// Bind and validate request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format",
				"details": err.Error(),
			})
			return
		}
		
		// Set default max results if not specified
		if req.MaxResults == 0 {
			req.MaxResults = 100
		}
		
		// Perform scan
		logger.Infof("Received scan request for %s", req.Symbol)
		
		response, err := scanner.ScanOptions(c.Request.Context(), &req)
		if err != nil {
			logger.Errorf("Scan failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Scan failed",
				"details": err.Error(),
			})
			return
		}
		
		// Limit results if needed
		if len(response.Options) > req.MaxResults {
			response.Options = response.Options[:req.MaxResults]
			response.ResultCount = req.MaxResults
		}
		
		c.JSON(http.StatusOK, response)
	}
}

// statsHandler returns scanner statistics
func statsHandler(scanner *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := scanner.GetStats()
		stats["uptime"] = time.Since(startTime).String()
		
		c.JSON(http.StatusOK, stats)
	}
}

// pingHandler is a simple test endpoint
func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time": time.Now(),
	})
}