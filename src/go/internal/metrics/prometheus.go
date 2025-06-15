package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

// Metrics holds all Prometheus metrics for the scanner
type Metrics struct {
	// Request metrics
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestsActive  prometheus.Gauge
	
	// Scan metrics
	ScansTotal      *prometheus.CounterVec
	ScanDuration    *prometheus.HistogramVec
	SpreadsFound    *prometheus.HistogramVec
	
	// Filter metrics
	FiltersApplied  *prometheus.CounterVec
	FilterDuration  *prometheus.HistogramVec
	FilterPassRate  *prometheus.GaugeVec
	
	// Scoring metrics
	ScoreCalculated *prometheus.CounterVec
	ScoreDuration   *prometheus.HistogramVec
	ScoreDistribution *prometheus.HistogramVec
	
	// Greeks metrics
	GreeksAnalyzed  *prometheus.CounterVec
	GreeksDuration  *prometheus.HistogramVec
	RiskScores      *prometheus.HistogramVec
	
	// Cache metrics
	CacheHits       prometheus.Counter
	CacheMisses     prometheus.Counter
	CacheSize       prometheus.Gauge
	CacheEvictions  prometheus.Counter
	
	// System metrics
	GoRoutines      prometheus.Gauge
	MemoryUsage     prometheus.Gauge
	CPUUsage        prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// Request metrics
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scanner_requests_total",
				Help: "Total number of scan requests",
			},
			[]string{"method", "status"},
		),
		
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		
		RequestsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scanner_requests_active",
				Help: "Number of active requests",
			},
		),
		
		// Scan metrics
		ScansTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scanner_scans_total",
				Help: "Total number of scans performed",
			},
			[]string{"symbol", "status"},
		),
		
		ScanDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_scan_duration_seconds",
				Help:    "Scan duration in seconds",
				Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"symbol"},
		),
		
		SpreadsFound: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_spreads_found",
				Help:    "Number of spreads found per scan",
				Buckets: []float64{0, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"symbol"},
		),
		
		// Filter metrics
		FiltersApplied: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scanner_filters_applied_total",
				Help: "Total number of filters applied",
			},
			[]string{"filter_type"},
		),
		
		FilterDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_filter_duration_seconds",
				Help:    "Filter execution duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"filter_type"},
		),
		
		FilterPassRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "scanner_filter_pass_rate",
				Help: "Percentage of options passing each filter",
			},
			[]string{"filter_type"},
		),
		
		// Scoring metrics
		ScoreCalculated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scanner_scores_calculated_total",
				Help: "Total number of scores calculated",
			},
			[]string{"scoring_mode"},
		),
		
		ScoreDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_score_calculation_duration_seconds",
				Help:    "Score calculation duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"scoring_mode"},
		),
		
		ScoreDistribution: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_score_distribution",
				Help:    "Distribution of calculated scores",
				Buckets: []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			},
			[]string{"symbol"},
		),
		
		// Greeks metrics
		GreeksAnalyzed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "scanner_greeks_analyzed_total",
				Help: "Total number of Greeks analyses performed",
			},
			[]string{"symbol"},
		),
		
		GreeksDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_greeks_analysis_duration_seconds",
				Help:    "Greeks analysis duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"symbol"},
		),
		
		RiskScores: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "scanner_risk_score_distribution",
				Help:    "Distribution of risk scores",
				Buckets: []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			},
			[]string{"symbol"},
		),
		
		// Cache metrics
		CacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "scanner_cache_hits_total",
				Help: "Total number of cache hits",
			},
		),
		
		CacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "scanner_cache_misses_total",
				Help: "Total number of cache misses",
			},
		),
		
		CacheSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scanner_cache_size",
				Help: "Current cache size",
			},
		),
		
		CacheEvictions: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "scanner_cache_evictions_total",
				Help: "Total number of cache evictions",
			},
		),
		
		// System metrics
		GoRoutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scanner_goroutines",
				Help: "Number of goroutines",
			},
		),
		
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scanner_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
		),
		
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "scanner_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),
	}
	
	return m
}

// RecordRequest records metrics for an HTTP request
func (m *Metrics) RecordRequest(method, status string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, status).Inc()
	m.RequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordScan records metrics for a scan operation
func (m *Metrics) RecordScan(symbol, status string, duration time.Duration, spreadsFound int) {
	m.ScansTotal.WithLabelValues(symbol, status).Inc()
	m.ScanDuration.WithLabelValues(symbol).Observe(duration.Seconds())
	if status == "success" {
		m.SpreadsFound.WithLabelValues(symbol).Observe(float64(spreadsFound))
	}
}

// RecordFilter records metrics for filter application
func (m *Metrics) RecordFilter(filterType string, duration time.Duration, passRate float64) {
	m.FiltersApplied.WithLabelValues(filterType).Inc()
	m.FilterDuration.WithLabelValues(filterType).Observe(duration.Seconds())
	m.FilterPassRate.WithLabelValues(filterType).Set(passRate)
}

// RecordScore records metrics for score calculation
func (m *Metrics) RecordScore(mode string, duration time.Duration, score float64, symbol string) {
	m.ScoreCalculated.WithLabelValues(mode).Inc()
	m.ScoreDuration.WithLabelValues(mode).Observe(duration.Seconds())
	m.ScoreDistribution.WithLabelValues(symbol).Observe(score)
}

// RecordGreeks records metrics for Greeks analysis
func (m *Metrics) RecordGreeks(symbol string, duration time.Duration, riskScore float64) {
	m.GreeksAnalyzed.WithLabelValues(symbol).Inc()
	m.GreeksDuration.WithLabelValues(symbol).Observe(duration.Seconds())
	m.RiskScores.WithLabelValues(symbol).Observe(riskScore)
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.CacheHits.Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.CacheMisses.Inc()
}

// UpdateCacheSize updates the current cache size
func (m *Metrics) UpdateCacheSize(size int) {
	m.CacheSize.Set(float64(size))
}

// RecordCacheEviction records a cache eviction
func (m *Metrics) RecordCacheEviction() {
	m.CacheEvictions.Inc()
}

// UpdateSystemMetrics updates system resource metrics
func (m *Metrics) UpdateSystemMetrics(goroutines int, memoryBytes uint64, cpuPercent float64) {
	m.GoRoutines.Set(float64(goroutines))
	m.MemoryUsage.Set(float64(memoryBytes))
	m.CPUUsage.Set(cpuPercent)
}

// Handler returns the Prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// Middleware provides HTTP middleware for recording request metrics
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Track active requests
		m.RequestsActive.Inc()
		defer m.RequestsActive.Dec()
		
		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Record metrics
		duration := time.Since(start)
		status := statusClass(wrapped.statusCode)
		m.RecordRequest(r.Method, status, duration)
	})
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

// statusClass converts HTTP status code to class (2xx, 4xx, 5xx)
func statusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}