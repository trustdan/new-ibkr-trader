package filters

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
)

// FilterVisualizer creates visual representations of filter chains
type FilterVisualizer struct {
	chain *AdvancedFilterChain
}

// NewFilterVisualizer creates a new visualizer
func NewFilterVisualizer(chain *AdvancedFilterChain) *FilterVisualizer {
	return &FilterVisualizer{chain: chain}
}

// VisualizationData contains data for visualization
type VisualizationData struct {
	Filters       []FilterInfo       `json:"filters"`
	Stats         map[string]*FilterStats `json:"stats"`
	FlowDiagram   string            `json:"flow_diagram"`
	PerformanceData []PerformancePoint `json:"performance_data"`
}

// FilterInfo describes a filter for visualization
type FilterInfo struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Position    int                    `json:"position"`
}

// PerformancePoint represents a performance data point
type PerformancePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	FilterName   string    `json:"filter_name"`
	Duration     float64   `json:"duration_ms"`
	ItemsIn      int       `json:"items_in"`
	ItemsOut     int       `json:"items_out"`
	ReductionPct float64   `json:"reduction_pct"`
}

// GenerateVisualization creates visualization data
func (fv *FilterVisualizer) GenerateVisualization() VisualizationData {
	data := VisualizationData{
		Filters:         fv.extractFilterInfo(),
		Stats:          fv.chain.GetStats(),
		FlowDiagram:    fv.generateFlowDiagram(),
		PerformanceData: fv.extractPerformanceData(),
	}
	
	return data
}

// extractFilterInfo extracts filter information
func (fv *FilterVisualizer) extractFilterInfo() []FilterInfo {
	infos := make([]FilterInfo, 0)
	position := 0
	
	// Contract filters
	for _, filter := range fv.chain.contractFilters {
		info := FilterInfo{
			Name:       filter.Name(),
			Type:       "contract",
			Position:   position,
			Parameters: fv.extractParameters(filter),
		}
		infos = append(infos, info)
		position++
	}
	
	// Spread filters
	for _, filter := range fv.chain.spreadFilters {
		info := FilterInfo{
			Name:       filter.Name(),
			Type:       "spread",
			Position:   position,
			Parameters: fv.extractParameters(filter),
		}
		infos = append(infos, info)
		position++
	}
	
	// Combined filters
	for _, filter := range fv.chain.combinedFilters {
		info := FilterInfo{
			Name:       filter.Name(),
			Type:       "combined",
			Position:   position,
			Parameters: fv.extractParameters(filter),
		}
		infos = append(infos, info)
		position++
	}
	
	return infos
}

// extractParameters extracts filter parameters via reflection
func (fv *FilterVisualizer) extractParameters(filter interface{}) map[string]interface{} {
	// Simplified - in real implementation would use reflection
	params := make(map[string]interface{})
	
	// Marshal to JSON and back to get parameters
	data, _ := json.Marshal(filter)
	json.Unmarshal(data, &params)
	
	return params
}

// generateFlowDiagram creates an ASCII flow diagram
func (fv *FilterVisualizer) generateFlowDiagram() string {
	var builder strings.Builder
	
	builder.WriteString("Filter Chain Flow:\n")
	builder.WriteString("=================\n\n")
	
	// Input
	builder.WriteString("    [Input Contracts/Spreads]\n")
	builder.WriteString("              |\n")
	builder.WriteString("              v\n")
	
	// Contract filters
	if len(fv.chain.contractFilters) > 0 {
		builder.WriteString("    +-------------------+\n")
		builder.WriteString("    | Contract Filters  |\n")
		builder.WriteString("    +-------------------+\n")
		
		for i, filter := range fv.chain.contractFilters {
			stats := fv.chain.executionStats[filter.Name()]
			reduction := float64(0)
			if stats != nil && stats.ItemsProcessed > 0 {
				reduction = float64(stats.ItemsFiltered) / float64(stats.ItemsProcessed) * 100
			}
			
			builder.WriteString(fmt.Sprintf("    | %-17s |\n", filter.Name()))
			builder.WriteString(fmt.Sprintf("    | Reduction: %5.1f%% |\n", reduction))
			
			if i < len(fv.chain.contractFilters)-1 {
				builder.WriteString("    |         |         |\n")
				builder.WriteString("    |         v         |\n")
			}
		}
		
		builder.WriteString("    +-------------------+\n")
		builder.WriteString("              |\n")
		builder.WriteString("              v\n")
	}
	
	// Spread filters
	if len(fv.chain.spreadFilters) > 0 {
		builder.WriteString("    +-------------------+\n")
		builder.WriteString("    |  Spread Filters   |\n")
		builder.WriteString("    +-------------------+\n")
		
		for _, filter := range fv.chain.spreadFilters {
			builder.WriteString(fmt.Sprintf("    | %-17s |\n", filter.Name()))
		}
		
		builder.WriteString("    +-------------------+\n")
		builder.WriteString("              |\n")
		builder.WriteString("              v\n")
	}
	
	// Combined filters
	if len(fv.chain.combinedFilters) > 0 {
		builder.WriteString("    +-------------------+\n")
		builder.WriteString("    | Combined Filters  |\n")
		builder.WriteString("    +-------------------+\n")
		
		for _, filter := range fv.chain.combinedFilters {
			builder.WriteString(fmt.Sprintf("    | %-17s |\n", filter.Name()))
		}
		
		builder.WriteString("    +-------------------+\n")
		builder.WriteString("              |\n")
		builder.WriteString("              v\n")
	}
	
	// Output
	builder.WriteString("    [Filtered Results]\n")
	
	// Add performance summary
	builder.WriteString("\n\nPerformance Summary:\n")
	builder.WriteString("-------------------\n")
	
	if fv.chain.parallelExecution {
		builder.WriteString("Execution Mode: Parallel\n")
	} else {
		builder.WriteString("Execution Mode: Sequential\n")
	}
	
	if fv.chain.cacheEnabled {
		if fv.chain.cache != nil {
			hits, misses, evictions, hitRate := fv.chain.cache.GetStats()
			builder.WriteString(fmt.Sprintf("Cache: Enabled (Hit Rate: %.1f%%, Hits: %d, Misses: %d, Evictions: %d)\n",
				hitRate*100, hits, misses, evictions))
		}
	} else {
		builder.WriteString("Cache: Disabled\n")
	}
	
	return builder.String()
}

// extractPerformanceData extracts performance metrics
func (fv *FilterVisualizer) extractPerformanceData() []PerformancePoint {
	points := make([]PerformancePoint, 0)
	
	for filterName, stats := range fv.chain.executionStats {
		if stats.ExecutionCount > 0 {
			point := PerformancePoint{
				Timestamp:    stats.LastExecution,
				FilterName:   filterName,
				Duration:     float64(stats.AverageDuration.Microseconds()) / 1000.0,
				ItemsIn:      int(stats.ItemsProcessed / stats.ExecutionCount),
				ItemsOut:     int((stats.ItemsProcessed - stats.ItemsFiltered) / stats.ExecutionCount),
			}
			
			if point.ItemsIn > 0 {
				point.ReductionPct = float64(point.ItemsIn-point.ItemsOut) / float64(point.ItemsIn) * 100
			}
			
			points = append(points, point)
		}
	}
	
	return points
}

// RenderHTML renders visualization as HTML
func (fv *FilterVisualizer) RenderHTML(w io.Writer) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Filter Chain Visualization</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .filter-box { 
            border: 2px solid #333; 
            border-radius: 5px; 
            padding: 10px; 
            margin: 10px;
            background-color: #f0f0f0;
        }
        .contract-filter { border-color: #0066cc; }
        .spread-filter { border-color: #00cc66; }
        .combined-filter { border-color: #cc6600; }
        .stats-table { 
            border-collapse: collapse; 
            width: 100%; 
            margin-top: 20px;
        }
        .stats-table th, .stats-table td { 
            border: 1px solid #ddd; 
            padding: 8px; 
            text-align: left;
        }
        .stats-table th { background-color: #f2f2f2; }
        .flow-diagram { 
            font-family: monospace; 
            white-space: pre; 
            background-color: #f8f8f8;
            padding: 15px;
            border: 1px solid #ddd;
            border-radius: 5px;
        }
        .performance-chart {
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <h1>Filter Chain Visualization</h1>
    
    <h2>Active Filters</h2>
    {{range .Filters}}
    <div class="filter-box {{.Type}}-filter">
        <h3>{{.Name}}</h3>
        <p>Type: {{.Type}}</p>
        <p>Position: {{.Position}}</p>
        {{if .Parameters}}
        <h4>Parameters:</h4>
        <ul>
        {{range $key, $value := .Parameters}}
            <li>{{$key}}: {{$value}}</li>
        {{end}}
        </ul>
        {{end}}
    </div>
    {{end}}
    
    <h2>Flow Diagram</h2>
    <div class="flow-diagram">{{.FlowDiagram}}</div>
    
    <h2>Performance Statistics</h2>
    <table class="stats-table">
        <thead>
            <tr>
                <th>Filter Name</th>
                <th>Executions</th>
                <th>Avg Duration</th>
                <th>Items Processed</th>
                <th>Items Filtered</th>
                <th>Filter Rate</th>
            </tr>
        </thead>
        <tbody>
        {{range $name, $stats := .Stats}}
            <tr>
                <td>{{$name}}</td>
                <td>{{$stats.ExecutionCount}}</td>
                <td>{{$stats.AverageDuration}}</td>
                <td>{{$stats.ItemsProcessed}}</td>
                <td>{{$stats.ItemsFiltered}}</td>
                <td>{{if gt $stats.ItemsProcessed 0}}{{printf "%.1f%%" (mulf (divf $stats.ItemsFiltered $stats.ItemsProcessed) 100)}}{{else}}0%{{end}}</td>
            </tr>
        {{end}}
        </tbody>
    </table>
    
    <h2>Performance Timeline</h2>
    <div class="performance-chart">
        <canvas id="perfChart" width="800" height="400"></canvas>
    </div>
    
    <script>
        // Performance data for charting
        const perfData = {{.PerformanceData}};
        
        // Simple performance visualization
        if (perfData && perfData.length > 0) {
            const canvas = document.getElementById('perfChart');
            const ctx = canvas.getContext('2d');
            
            // Draw axes
            ctx.beginPath();
            ctx.moveTo(50, 350);
            ctx.lineTo(750, 350);
            ctx.moveTo(50, 50);
            ctx.lineTo(50, 350);
            ctx.stroke();
            
            // Plot performance points
            const maxDuration = Math.max(...perfData.map(p => p.duration_ms));
            const barWidth = 700 / perfData.length;
            
            perfData.forEach((point, index) => {
                const x = 50 + index * barWidth + barWidth/4;
                const height = (point.duration_ms / maxDuration) * 250;
                const y = 350 - height;
                
                // Draw bar
                ctx.fillStyle = '#0066cc';
                ctx.fillRect(x, y, barWidth/2, height);
                
                // Draw label
                ctx.save();
                ctx.translate(x + barWidth/4, 360);
                ctx.rotate(-Math.PI/4);
                ctx.fillStyle = '#000';
                ctx.font = '10px Arial';
                ctx.fillText(point.filter_name, 0, 0);
                ctx.restore();
            });
        }
    </script>
</body>
</html>
`
	
	// Create template with custom functions
	funcMap := template.FuncMap{
		"mulf": func(a, b float64) float64 { return a * b },
		"divf": func(a, b float64) float64 { return a / b },
	}
	
	t, err := template.New("visualization").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	data := fv.GenerateVisualization()
	return t.Execute(w, data)
}

// RenderJSON renders visualization as JSON
func (fv *FilterVisualizer) RenderJSON(w io.Writer) error {
	data := fv.GenerateVisualization()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// RenderMarkdown renders visualization as Markdown
func (fv *FilterVisualizer) RenderMarkdown(w io.Writer) error {
	data := fv.GenerateVisualization()
	
	fmt.Fprintf(w, "# Filter Chain Visualization\n\n")
	
	// Filters section
	fmt.Fprintf(w, "## Active Filters\n\n")
	for _, filter := range data.Filters {
		fmt.Fprintf(w, "### %s\n", filter.Name)
		fmt.Fprintf(w, "- **Type**: %s\n", filter.Type)
		fmt.Fprintf(w, "- **Position**: %d\n", filter.Position)
		
		if len(filter.Parameters) > 0 {
			fmt.Fprintf(w, "- **Parameters**:\n")
			for key, value := range filter.Parameters {
				fmt.Fprintf(w, "  - %s: %v\n", key, value)
			}
		}
		fmt.Fprintf(w, "\n")
	}
	
	// Flow diagram
	fmt.Fprintf(w, "## Flow Diagram\n\n```\n%s\n```\n\n", data.FlowDiagram)
	
	// Statistics table
	fmt.Fprintf(w, "## Performance Statistics\n\n")
	fmt.Fprintf(w, "| Filter Name | Executions | Avg Duration | Items Processed | Items Filtered | Filter Rate |\n")
	fmt.Fprintf(w, "|-------------|------------|--------------|-----------------|----------------|-------------|\n")
	
	for name, stats := range data.Stats {
		filterRate := float64(0)
		if stats.ItemsProcessed > 0 {
			filterRate = float64(stats.ItemsFiltered) / float64(stats.ItemsProcessed) * 100
		}
		
		fmt.Fprintf(w, "| %s | %d | %v | %d | %d | %.1f%% |\n",
			name,
			stats.ExecutionCount,
			stats.AverageDuration,
			stats.ItemsProcessed,
			stats.ItemsFiltered,
			filterRate,
		)
	}
	
	return nil
}