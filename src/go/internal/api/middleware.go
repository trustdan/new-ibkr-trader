package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	// Prometheus metrics
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scanner_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "scanner_http_request_duration_seconds",
		Help: "HTTP request duration in seconds",
	}, []string{"method", "path"})

	scanOperationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "scanner_scan_operations_total",
		Help: "Total number of scan operations",
	})
)

// LoggerMiddleware provides request logging
func LoggerMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		
		// Process request
		c.Next()
		
		// Skip logging for metrics endpoint
		if path == "/metrics" {
			return
		}
		
		// Log request details
		latency := time.Since(start)
		status := c.Writer.Status()
		
		logger.Infow("Request processed",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency,
			"client_ip", c.ClientIP(),
		)
	}
}

// MetricsMiddleware collects Prometheus metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		
		// Skip metrics for the metrics endpoint itself
		if path == "/metrics" {
			c.Next()
			return
		}
		
		// Process request
		c.Next()
		
		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			string(rune(status)),
		).Inc()
		
		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)
		
		// Track scan operations
		if path == "/api/v1/scan" && status == 200 {
			scanOperationsTotal.Inc()
		}
	}
}