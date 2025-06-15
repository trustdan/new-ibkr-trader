package metrics

import (
	"runtime"
	"time"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector collects and exposes scanner metrics
type MetricsCollector struct {
	// Scanner metrics
	ScansTotal        prometheus.Counter
	ScanDuration      prometheus.Histogram
	ScanErrors        prometheus.Counter
	ActiveScans       prometheus.Gauge
	
	// Filter metrics
	FilterExecutions  *prometheus.CounterVec
	FilterDuration    *prometheus.HistogramVec
	FilterReductions  *prometheus.HistogramVec
	
	// Result metrics
	ResultsTotal      *prometheus.CounterVec
	SpreadsFound      prometheus.Histogram
	ResultCacheHits   prometheus.Counter
	ResultCacheMisses prometheus.Counter
	
	// WebSocket metrics
	WSConnections     prometheus.Gauge
	WSMessagesIn      *prometheus.CounterVec
	WSMessagesOut     *prometheus.CounterVec
	WSBroadcastTime   prometheus.Histogram
	
	// Alert metrics
	AlertsTriggered   *prometheus.CounterVec
	AlertsAcked       prometheus.Counter
	AlertQueueSize    prometheus.Gauge
	
	// System metrics
	MemoryUsage       prometheus.Gauge
	GoroutineCount    prometheus.Gauge
	CPUUsage          prometheus.Gauge
	
	// Performance metrics
	ScanThroughput    prometheus.Gauge // scans per second
	ResultThroughput  prometheus.Gauge // results per second
	FilterEfficiency  prometheus.Gauge // average reduction percentage
	
	// Custom metrics tracking
	lastScanTime      time.Time
	scanCount         int64
	resultCount       int64
	ticker            *time.Ticker
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		// Scanner metrics
		ScansTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scanner_scans_total",
			Help: "Total number of scans performed",
		}),
		ScanDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "scanner_scan_duration_seconds",
			Help:    "Duration of scan operations",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to 10s
		}),
		ScanErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scanner_scan_errors_total",
			Help: "Total number of scan errors",
		}),
		ActiveScans: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_active_scans",
			Help: "Number of currently active scans",
		}),
		
		// Filter metrics
		FilterExecutions: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scanner_filter_executions_total",
			Help: "Total number of filter executions",
		}, []string{"filter_name"}),
		FilterDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "scanner_filter_duration_seconds",
			Help:    "Duration of filter executions",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10), // 0.1ms to 100ms
		}, []string{"filter_name"}),
		FilterReductions: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "scanner_filter_reduction_ratio",
			Help:    "Reduction ratio of filters (0-1)",
			Buckets: prometheus.LinearBuckets(0, 0.1, 11), // 0% to 100% in 10% steps
		}, []string{"filter_name"}),
		
		// Result metrics
		ResultsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scanner_results_total",
			Help: "Total number of scan results",
		}, []string{"symbol", "type"}),
		SpreadsFound: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "scanner_spreads_found",
			Help:    "Number of spreads found per scan",
			Buckets: prometheus.LinearBuckets(0, 5, 20), // 0 to 100 in steps of 5
		}),
		ResultCacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scanner_result_cache_hits_total",
			Help: "Total number of result cache hits",
		}),
		ResultCacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scanner_result_cache_misses_total",
			Help: "Total number of result cache misses",
		}),
		
		// WebSocket metrics
		WSConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_websocket_connections",
			Help: "Number of active WebSocket connections",
		}),
		WSMessagesIn: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scanner_websocket_messages_in_total",
			Help: "Total number of WebSocket messages received",
		}, []string{"type"}),
		WSMessagesOut: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scanner_websocket_messages_out_total",
			Help: "Total number of WebSocket messages sent",
		}, []string{"type"}),
		WSBroadcastTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "scanner_websocket_broadcast_duration_seconds",
			Help:    "Duration of WebSocket broadcast operations",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10), // 0.1ms to 100ms
		}),
		
		// Alert metrics
		AlertsTriggered: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scanner_alerts_triggered_total",
			Help: "Total number of alerts triggered",
		}, []string{"type", "severity"}),
		AlertsAcked: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scanner_alerts_acknowledged_total",
			Help: "Total number of alerts acknowledged",
		}),
		AlertQueueSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_alert_queue_size",
			Help: "Current size of alert queue",
		}),
		
		// System metrics
		MemoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		}),
		GoroutineCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_goroutines",
			Help: "Number of active goroutines",
		}),
		CPUUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		}),
		
		// Performance metrics
		ScanThroughput: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_scan_throughput",
			Help: "Scans per second",
		}),
		ResultThroughput: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_result_throughput",
			Help: "Results per second",
		}),
		FilterEfficiency: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "scanner_filter_efficiency_percent",
			Help: "Average filter reduction percentage",
		}),
	}
	
	// Start system metrics collection
	mc.startSystemMetrics()
	
	return mc
}

// startSystemMetrics starts collecting system metrics
func (mc *MetricsCollector) startSystemMetrics() {
	mc.ticker = time.NewTicker(10 * time.Second)
	
	go func() {
		for range mc.ticker.C {
			mc.collectSystemMetrics()
			mc.calculateThroughput()
		}
	}()
}

// collectSystemMetrics collects system-level metrics
func (mc *MetricsCollector) collectSystemMetrics() {
	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mc.MemoryUsage.Set(float64(m.Alloc))
	
	// Goroutine count
	mc.GoroutineCount.Set(float64(runtime.NumGoroutine()))
	
	// CPU usage would require more complex implementation
	// For now, we'll use a placeholder
	mc.CPUUsage.Set(0)
}

// calculateThroughput calculates throughput metrics
func (mc *MetricsCollector) calculateThroughput() {
	now := time.Now()
	if !mc.lastScanTime.IsZero() {
		duration := now.Sub(mc.lastScanTime).Seconds()
		
		// Calculate scans per second
		currentScanCount := mc.getCounterValue(mc.ScansTotal)
		scanRate := float64(currentScanCount-mc.scanCount) / duration
		mc.ScanThroughput.Set(scanRate)
		mc.scanCount = currentScanCount
		
		// Calculate results per second
		currentResultCount := mc.getCounterVecValue(mc.ResultsTotal)
		resultRate := float64(currentResultCount-mc.resultCount) / duration
		mc.ResultThroughput.Set(resultRate)
		mc.resultCount = currentResultCount
	}
	
	mc.lastScanTime = now
}

// Helper methods for recording metrics

// RecordScan records scan metrics
func (mc *MetricsCollector) RecordScan(duration time.Duration, spreadsFound int, err error) {
	mc.ScansTotal.Inc()
	mc.ScanDuration.Observe(duration.Seconds())
	mc.SpreadsFound.Observe(float64(spreadsFound))
	
	if err != nil {
		mc.ScanErrors.Inc()
	}
}

// RecordFilter records filter execution metrics
func (mc *MetricsCollector) RecordFilter(filterName string, duration time.Duration, itemsIn, itemsOut int) {
	mc.FilterExecutions.WithLabelValues(filterName).Inc()
	mc.FilterDuration.WithLabelValues(filterName).Observe(duration.Seconds())
	
	if itemsIn > 0 {
		reduction := float64(itemsIn-itemsOut) / float64(itemsIn)
		mc.FilterReductions.WithLabelValues(filterName).Observe(reduction)
	}
}

// RecordResult records result metrics
func (mc *MetricsCollector) RecordResult(symbol, resultType string) {
	mc.ResultsTotal.WithLabelValues(symbol, resultType).Inc()
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.ResultCacheHits.Inc()
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.ResultCacheMisses.Inc()
}

// RecordWSMessage records WebSocket message metrics
func (mc *MetricsCollector) RecordWSMessage(direction, messageType string) {
	if direction == "in" {
		mc.WSMessagesIn.WithLabelValues(messageType).Inc()
	} else {
		mc.WSMessagesOut.WithLabelValues(messageType).Inc()
	}
}

// RecordBroadcast records broadcast duration
func (mc *MetricsCollector) RecordBroadcast(duration time.Duration) {
	mc.WSBroadcastTime.Observe(duration.Seconds())
}

// RecordAlert records alert metrics
func (mc *MetricsCollector) RecordAlert(alertType, severity string) {
	mc.AlertsTriggered.WithLabelValues(alertType, severity).Inc()
}

// UpdateActiveScans updates the active scan gauge
func (mc *MetricsCollector) UpdateActiveScans(delta int) {
	if delta > 0 {
		mc.ActiveScans.Inc()
	} else {
		mc.ActiveScans.Dec()
	}
}

// SetWSConnections sets the WebSocket connection count
func (mc *MetricsCollector) SetWSConnections(count int) {
	mc.WSConnections.Set(float64(count))
}

// SetAlertQueueSize sets the alert queue size
func (mc *MetricsCollector) SetAlertQueueSize(size int) {
	mc.AlertQueueSize.Set(float64(size))
}

// getCounterValue gets the current value of a counter (for internal use)
func (mc *MetricsCollector) getCounterValue(counter prometheus.Counter) int64 {
	// This is a simplified approach - in production you'd use the Prometheus API
	return 0
}

// getCounterVecValue gets the sum of all counter vec values
func (mc *MetricsCollector) getCounterVecValue(vec *prometheus.CounterVec) int64 {
	// This is a simplified approach - in production you'd use the Prometheus API
	return 0
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop() {
	if mc.ticker != nil {
		mc.ticker.Stop()
	}
}