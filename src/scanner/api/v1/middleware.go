package v1

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ibkr-trader/scanner/internal/metrics"
	"github.com/rs/zerolog/log"
)

// RequestIDKey is the context key for request ID
type RequestIDKey struct{}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Generate request ID
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
		r = r.WithContext(ctx)
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Log request
		log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("Request started")
		
		// Call the next handler
		next.ServeHTTP(wrapped, r)
		
		// Log response
		log.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", time.Since(start)).
			Msg("Request completed")
	})
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// metricsMiddleware records request metrics
func metricsMiddleware(collector *metrics.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Call the next handler
			next.ServeHTTP(wrapped, r)
			
			// Record metrics
			duration := time.Since(start).Seconds()
			path := cleanPath(r.URL.Path)
			
			collector.RecordHTTPRequest(r.Method, path, wrapped.statusCode, duration)
		})
	}
}

// authMiddleware validates API authentication (optional)
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		
		// For now, accept any non-empty API key
		// In production, validate against stored keys
		if apiKey == "" {
			// Allow unauthenticated requests for now
			// In production, return 401
		}
		
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements rate limiting
func rateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter
	// In production, use Redis or similar
	clients := make(map[string][]time.Time)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()
			
			// Clean old requests
			if requests, exists := clients[clientIP]; exists {
				var validRequests []time.Time
				for _, t := range requests {
					if now.Sub(t) < time.Minute {
						validRequests = append(validRequests, t)
					}
				}
				clients[clientIP] = validRequests
			}
			
			// Check rate limit
			if len(clients[clientIP]) >= requestsPerMinute {
				w.Header().Set("X-RateLimit-Limit", string(requestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", string(now.Add(time.Minute).Unix()))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			// Record request
			clients[clientIP] = append(clients[clientIP], now)
			
			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", string(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", string(requestsPerMinute-len(clients[clientIP])))
			
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// cleanPath removes variable parts from URL path for metrics
func cleanPath(path string) string {
	// Replace UUIDs and common patterns
	path = strings.ReplaceAll(path, "/api/v1", "")
	
	// Replace common dynamic segments
	parts := strings.Split(path, "/")
	for i, part := range parts {
		// Check if part looks like UUID
		if len(part) == 36 && strings.Count(part, "-") == 4 {
			parts[i] = ":id"
		}
		// Check if part is a symbol (uppercase letters)
		if len(part) > 0 && part == strings.ToUpper(part) && !strings.Contains(part, "/") {
			parts[i] = ":symbol"
		}
	}
	
	return strings.Join(parts, "/")
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}