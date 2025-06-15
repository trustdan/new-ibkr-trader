package metrics

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DashboardHandler provides metrics dashboard functionality
type DashboardHandler struct {
	collector *MetricsCollector
	history   *MetricsHistory
}

// MetricsHistory tracks historical metrics data
type MetricsHistory struct {
	ScanRate      []TimeSeriesPoint `json:"scan_rate"`
	ResultRate    []TimeSeriesPoint `json:"result_rate"`
	WSConnections []TimeSeriesPoint `json:"ws_connections"`
	MemoryUsage   []TimeSeriesPoint `json:"memory_usage"`
	maxPoints     int
}

// TimeSeriesPoint represents a data point in time
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// DashboardMetrics contains current metrics for the dashboard
type DashboardMetrics struct {
	// Overview
	TotalScans       int64   `json:"total_scans"`
	ScanRate         float64 `json:"scan_rate"`
	ActiveScans      int     `json:"active_scans"`
	ErrorRate        float64 `json:"error_rate"`
	
	// Performance
	AvgScanDuration  float64 `json:"avg_scan_duration_ms"`
	ResultThroughput float64 `json:"result_throughput"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	
	// WebSocket
	WSConnections    int     `json:"ws_connections"`
	WSMessageRate    float64 `json:"ws_message_rate"`
	
	// System
	MemoryUsageMB    float64 `json:"memory_usage_mb"`
	GoroutineCount   int     `json:"goroutine_count"`
	CPUUsage         float64 `json:"cpu_usage"`
	
	// Filters
	TopFilters       []FilterMetric `json:"top_filters"`
	FilterEfficiency float64        `json:"filter_efficiency"`
	
	// Alerts
	AlertsTriggered  int64   `json:"alerts_triggered"`
	AlertQueueSize   int     `json:"alert_queue_size"`
	
	// History
	History          *MetricsHistory `json:"history"`
}

// FilterMetric represents metrics for a single filter
type FilterMetric struct {
	Name           string  `json:"name"`
	Executions     int64   `json:"executions"`
	AvgDuration    float64 `json:"avg_duration_ms"`
	AvgReduction   float64 `json:"avg_reduction"`
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(collector *MetricsCollector) *DashboardHandler {
	return &DashboardHandler{
		collector: collector,
		history: &MetricsHistory{
			ScanRate:      make([]TimeSeriesPoint, 0),
			ResultRate:    make([]TimeSeriesPoint, 0),
			WSConnections: make([]TimeSeriesPoint, 0),
			MemoryUsage:   make([]TimeSeriesPoint, 0),
			maxPoints:     100,
		},
	}
}

// RegisterRoutes registers dashboard routes
func (h *DashboardHandler) RegisterRoutes(router *gin.Engine) {
	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// Dashboard endpoints
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("/", h.handleDashboard)
		dashboard.GET("/metrics", h.handleMetricsJSON)
		dashboard.GET("/history", h.handleHistoryJSON)
	}
}

// handleDashboard serves the metrics dashboard HTML
func (h *DashboardHandler) handleDashboard(c *gin.Context) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Scanner Metrics Dashboard</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }
        .dashboard {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .metric-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric-value {
            font-size: 2em;
            font-weight: bold;
            color: #333;
            margin: 10px 0;
        }
        .metric-label {
            color: #666;
            font-size: 0.9em;
        }
        .chart-container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
            height: 300px;
        }
        .status-good { color: #4CAF50; }
        .status-warning { color: #FF9800; }
        .status-error { color: #F44336; }
        .refresh-info {
            text-align: right;
            color: #999;
            font-size: 0.8em;
        }
        canvas { max-width: 100%; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="dashboard">
        <div class="header">
            <h1>Scanner Metrics Dashboard</h1>
            <div class="refresh-info">Auto-refreshes every 5 seconds</div>
        </div>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-label">Scan Rate</div>
                <div class="metric-value" id="scanRate">-</div>
                <div class="metric-label">scans/sec</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-label">Active Scans</div>
                <div class="metric-value" id="activeScans">-</div>
                <div class="metric-label">concurrent</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-label">WebSocket Connections</div>
                <div class="metric-value" id="wsConnections">-</div>
                <div class="metric-label">active</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-label">Cache Hit Rate</div>
                <div class="metric-value" id="cacheHitRate">-</div>
                <div class="metric-label">percent</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-label">Memory Usage</div>
                <div class="metric-value" id="memoryUsage">-</div>
                <div class="metric-label">MB</div>
            </div>
            
            <div class="metric-card">
                <div class="metric-label">Error Rate</div>
                <div class="metric-value" id="errorRate">-</div>
                <div class="metric-label">percent</div>
            </div>
        </div>
        
        <div class="chart-container">
            <h3>Scan Rate History</h3>
            <canvas id="scanRateChart"></canvas>
        </div>
        
        <div class="chart-container">
            <h3>Memory Usage History</h3>
            <canvas id="memoryChart"></canvas>
        </div>
        
        <div class="chart-container">
            <h3>Top Filters by Execution Time</h3>
            <canvas id="filterChart"></canvas>
        </div>
    </div>
    
    <script>
        // Initialize charts
        const scanRateChart = new Chart(document.getElementById('scanRateChart'), {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Scans/sec',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
        
        const memoryChart = new Chart(document.getElementById('memoryChart'), {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Memory (MB)',
                    data: [],
                    borderColor: 'rgb(255, 99, 132)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
        
        const filterChart = new Chart(document.getElementById('filterChart'), {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Avg Duration (ms)',
                    data: [],
                    backgroundColor: 'rgb(54, 162, 235)'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
        
        // Update function
        async function updateDashboard() {
            try {
                const response = await fetch('/dashboard/metrics');
                const data = await response.json();
                
                // Update metric cards
                document.getElementById('scanRate').textContent = data.scan_rate.toFixed(2);
                document.getElementById('activeScans').textContent = data.active_scans;
                document.getElementById('wsConnections').textContent = data.ws_connections;
                document.getElementById('cacheHitRate').textContent = (data.cache_hit_rate * 100).toFixed(1);
                document.getElementById('memoryUsage').textContent = data.memory_usage_mb.toFixed(1);
                document.getElementById('errorRate').textContent = (data.error_rate * 100).toFixed(1);
                
                // Update charts with history
                if (data.history) {
                    // Scan rate chart
                    const scanLabels = data.history.scan_rate.map(p => 
                        new Date(p.timestamp).toLocaleTimeString()
                    );
                    scanRateChart.data.labels = scanLabels;
                    scanRateChart.data.datasets[0].data = data.history.scan_rate.map(p => p.value);
                    scanRateChart.update('none');
                    
                    // Memory chart
                    const memLabels = data.history.memory_usage.map(p => 
                        new Date(p.timestamp).toLocaleTimeString()
                    );
                    memoryChart.data.labels = memLabels;
                    memoryChart.data.datasets[0].data = data.history.memory_usage.map(p => p.value);
                    memoryChart.update('none');
                }
                
                // Update filter chart
                if (data.top_filters && data.top_filters.length > 0) {
                    filterChart.data.labels = data.top_filters.map(f => f.name);
                    filterChart.data.datasets[0].data = data.top_filters.map(f => f.avg_duration_ms);
                    filterChart.update('none');
                }
                
                // Color code metrics
                if (data.error_rate > 0.05) {
                    document.getElementById('errorRate').className = 'metric-value status-error';
                } else if (data.error_rate > 0.01) {
                    document.getElementById('errorRate').className = 'metric-value status-warning';
                } else {
                    document.getElementById('errorRate').className = 'metric-value status-good';
                }
                
            } catch (error) {
                console.error('Failed to update dashboard:', error);
            }
        }
        
        // Initial update and refresh interval
        updateDashboard();
        setInterval(updateDashboard, 5000);
    </script>
</body>
</html>
`
	
	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Type", "text/html")
	t.Execute(c.Writer, nil)
}

// handleMetricsJSON returns current metrics as JSON
func (h *DashboardHandler) handleMetricsJSON(c *gin.Context) {
	metrics := h.collectCurrentMetrics()
	c.JSON(http.StatusOK, metrics)
}

// handleHistoryJSON returns historical metrics
func (h *DashboardHandler) handleHistoryJSON(c *gin.Context) {
	c.JSON(http.StatusOK, h.history)
}

// collectCurrentMetrics collects current metrics from Prometheus
func (h *DashboardHandler) collectCurrentMetrics() DashboardMetrics {
	// This is a simplified implementation
	// In production, you would query the Prometheus API
	
	metrics := DashboardMetrics{
		// Mock data - replace with actual Prometheus queries
		TotalScans:       1000,
		ScanRate:         2.5,
		ActiveScans:      3,
		ErrorRate:        0.02,
		AvgScanDuration:  150.0,
		ResultThroughput: 10.5,
		CacheHitRate:     0.85,
		WSConnections:    5,
		WSMessageRate:    25.0,
		MemoryUsageMB:    128.5,
		GoroutineCount:   42,
		CPUUsage:         15.0,
		FilterEfficiency: 0.75,
		AlertsTriggered:  25,
		AlertQueueSize:   2,
		History:          h.history,
	}
	
	// Update history
	h.updateHistory(metrics)
	
	return metrics
}

// updateHistory updates the metrics history
func (h *DashboardHandler) updateHistory(metrics DashboardMetrics) {
	now := time.Now()
	
	// Add new points
	h.history.ScanRate = h.addPoint(h.history.ScanRate, TimeSeriesPoint{
		Timestamp: now,
		Value:     metrics.ScanRate,
	})
	
	h.history.ResultRate = h.addPoint(h.history.ResultRate, TimeSeriesPoint{
		Timestamp: now,
		Value:     metrics.ResultThroughput,
	})
	
	h.history.WSConnections = h.addPoint(h.history.WSConnections, TimeSeriesPoint{
		Timestamp: now,
		Value:     float64(metrics.WSConnections),
	})
	
	h.history.MemoryUsage = h.addPoint(h.history.MemoryUsage, TimeSeriesPoint{
		Timestamp: now,
		Value:     metrics.MemoryUsageMB,
	})
}

// addPoint adds a point to the time series, maintaining max size
func (h *DashboardHandler) addPoint(series []TimeSeriesPoint, point TimeSeriesPoint) []TimeSeriesPoint {
	series = append(series, point)
	
	if len(series) > h.history.maxPoints {
		series = series[len(series)-h.history.maxPoints:]
	}
	
	return series
}

// ExportMetrics exports metrics data in various formats
func (h *DashboardHandler) ExportMetrics(format string) ([]byte, error) {
	metrics := h.collectCurrentMetrics()
	
	switch format {
	case "json":
		return json.MarshalIndent(metrics, "", "  ")
	case "csv":
		// Simplified CSV export
		csv := "timestamp,scan_rate,result_rate,ws_connections,memory_mb\n"
		for i := range h.history.ScanRate {
			csv += fmt.Sprintf("%s,%.2f,%.2f,%.0f,%.2f\n",
				h.history.ScanRate[i].Timestamp.Format(time.RFC3339),
				h.history.ScanRate[i].Value,
				h.history.ResultRate[i].Value,
				h.history.WSConnections[i].Value,
				h.history.MemoryUsage[i].Value,
			)
		}
		return []byte(csv), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}