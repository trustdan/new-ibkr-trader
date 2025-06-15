package streaming

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/rs/zerolog/log"
)

// FilterManager manages dynamic filter updates via WebSocket
type FilterManager struct {
	currentConfig filters.FilterConfig
	filterChain   *filters.AdvancedFilterChain
	presets       map[string]filters.FilterConfig
	mu            sync.RWMutex
	
	// Callback for filter updates
	onUpdate      func(*filters.AdvancedFilterChain)
	
	// History of filter changes
	history       []FilterChange
	maxHistory    int
}

// FilterChange records a filter configuration change
type FilterChange struct {
	Timestamp   time.Time              `json:"timestamp"`
	ChangeType  string                 `json:"change_type"` // update, preset, reset
	Previous    filters.FilterConfig   `json:"previous"`
	Current     filters.FilterConfig   `json:"current"`
	ChangedBy   string                 `json:"changed_by"`
	Description string                 `json:"description"`
}

// FilterUpdateMessage represents a filter update via WebSocket
type FilterUpdateMessage struct {
	Type        string                 `json:"type"`        // "update", "preset", "reset"
	Filters     *filters.FilterConfig  `json:"filters,omitempty"`
	PresetName  string                 `json:"preset_name,omitempty"`
	ClientID    string                 `json:"client_id"`
}

// NewFilterManager creates a new filter manager
func NewFilterManager(initialConfig filters.FilterConfig) *FilterManager {
	fm := &FilterManager{
		currentConfig: initialConfig,
		filterChain:   filters.NewAdvancedFilterChain(initialConfig, true, true),
		presets:       loadDefaultPresets(),
		history:       make([]FilterChange, 0),
		maxHistory:    100,
	}
	
	return fm
}

// loadDefaultPresets loads default filter presets
func loadDefaultPresets() map[string]filters.FilterConfig {
	presets := make(map[string]filters.FilterConfig)
	
	// Conservative preset
	presets["conservative"] = filters.FilterConfig{
		Delta:        &filters.DeltaFilter{MinDelta: 0.15, MaxDelta: 0.30},
		DTE:          &filters.DTEFilter{MinDTE: 30, MaxDTE: 60},
		Liquidity:    &filters.LiquidityFilter{MinOpenInterest: 100, MinVolume: 50},
		IVPercentile: &filters.IVPercentileFilter{MinPercentile: 30, MaxPercentile: 70},
		ProbOfProfit: &filters.PoPFilter{MinPoP: 0.70, MaxPoP: 0.90},
		MaxPositions: 5,
		RiskLimit:    5000,
	}
	
	// Moderate preset
	presets["moderate"] = filters.FilterConfig{
		Delta:        &filters.DeltaFilter{MinDelta: 0.20, MaxDelta: 0.40},
		DTE:          &filters.DTEFilter{MinDTE: 20, MaxDTE: 45},
		Liquidity:    &filters.LiquidityFilter{MinOpenInterest: 50, MinVolume: 25},
		IVPercentile: &filters.IVPercentileFilter{MinPercentile: 40, MaxPercentile: 80},
		ProbOfProfit: &filters.PoPFilter{MinPoP: 0.60, MaxPoP: 0.85},
		MaxPositions: 10,
		RiskLimit:    10000,
	}
	
	// Aggressive preset
	presets["aggressive"] = filters.FilterConfig{
		Delta:        &filters.DeltaFilter{MinDelta: 0.25, MaxDelta: 0.50},
		DTE:          &filters.DTEFilter{MinDTE: 7, MaxDTE: 30},
		Liquidity:    &filters.LiquidityFilter{MinOpenInterest: 25, MinVolume: 10},
		IVPercentile: &filters.IVPercentileFilter{MinPercentile: 50, MaxPercentile: 90},
		ProbOfProfit: &filters.PoPFilter{MinPoP: 0.50, MaxPoP: 0.80},
		MaxPositions: 20,
		RiskLimit:    20000,
	}
	
	// High IV preset
	presets["high_iv"] = filters.FilterConfig{
		Delta:     &filters.DeltaFilter{MinDelta: 0.10, MaxDelta: 0.25},
		DTE:       &filters.DTEFilter{MinDTE: 30, MaxDTE: 60},
		Liquidity: &filters.LiquidityFilter{MinOpenInterest: 100, MinVolume: 50},
		IV:        &filters.IVFilter{MinIV: 0.30, MaxIV: 1.0},
		IVPercentile: &filters.IVPercentileFilter{MinPercentile: 70, MaxPercentile: 100},
		Vega:      &filters.VegaFilter{MinVega: 0.05, MaxVega: 0.20},
		MaxPositions: 8,
		RiskLimit: 8000,
	}
	
	// Theta harvesting preset
	presets["theta_harvest"] = filters.FilterConfig{
		Delta:        &filters.DeltaFilter{MinDelta: 0.20, MaxDelta: 0.35},
		DTE:          &filters.DTEFilter{MinDTE: 15, MaxDTE: 45},
		Liquidity:    &filters.LiquidityFilter{MinOpenInterest: 50, MinVolume: 25},
		Theta:        &filters.ThetaFilter{MinTheta: 0.02, MaxTheta: 0.10},
		ProbOfProfit: &filters.PoPFilter{MinPoP: 0.65, MaxPoP: 0.85},
		MaxPositions: 15,
		RiskLimit:    15000,
	}
	
	return presets
}

// HandleFilterUpdate handles a filter update message
func (fm *FilterManager) HandleFilterUpdate(msg FilterUpdateMessage) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	// Record previous config
	previous := fm.currentConfig
	
	switch msg.Type {
	case "update":
		if msg.Filters == nil {
			return fmt.Errorf("filters required for update")
		}
		
		// Apply filter updates
		if err := fm.applyFilterUpdate(*msg.Filters); err != nil {
			return err
		}
		
		// Record change
		fm.recordChange(FilterChange{
			Timestamp:   time.Now(),
			ChangeType:  "update",
			Previous:    previous,
			Current:     fm.currentConfig,
			ChangedBy:   msg.ClientID,
			Description: "Manual filter update",
		})
		
	case "preset":
		preset, exists := fm.presets[msg.PresetName]
		if !exists {
			return fmt.Errorf("preset not found: %s", msg.PresetName)
		}
		
		// Apply preset
		fm.currentConfig = preset
		fm.filterChain = filters.NewAdvancedFilterChain(preset, true, true)
		
		// Record change
		fm.recordChange(FilterChange{
			Timestamp:   time.Now(),
			ChangeType:  "preset",
			Previous:    previous,
			Current:     fm.currentConfig,
			ChangedBy:   msg.ClientID,
			Description: fmt.Sprintf("Applied preset: %s", msg.PresetName),
		})
		
	case "reset":
		// Reset to initial config
		fm.currentConfig = filters.FilterConfig{}
		fm.filterChain = filters.NewAdvancedFilterChain(fm.currentConfig, true, true)
		
		// Record change
		fm.recordChange(FilterChange{
			Timestamp:   time.Now(),
			ChangeType:  "reset",
			Previous:    previous,
			Current:     fm.currentConfig,
			ChangedBy:   msg.ClientID,
			Description: "Reset to default filters",
		})
		
	default:
		return fmt.Errorf("unknown update type: %s", msg.Type)
	}
	
	// Notify callback
	if fm.onUpdate != nil {
		fm.onUpdate(fm.filterChain)
	}
	
	log.Info().
		Str("type", msg.Type).
		Str("client", msg.ClientID).
		Msg("Filter configuration updated")
	
	return nil
}

// applyFilterUpdate applies incremental filter updates
func (fm *FilterManager) applyFilterUpdate(update filters.FilterConfig) error {
	// Update individual filters if provided
	if update.Delta != nil {
		fm.currentConfig.Delta = update.Delta
	}
	if update.DTE != nil {
		fm.currentConfig.DTE = update.DTE
	}
	if update.Liquidity != nil {
		fm.currentConfig.Liquidity = update.Liquidity
	}
	if update.Theta != nil {
		fm.currentConfig.Theta = update.Theta
	}
	if update.Vega != nil {
		fm.currentConfig.Vega = update.Vega
	}
	if update.IV != nil {
		fm.currentConfig.IV = update.IV
	}
	if update.IVPercentile != nil {
		fm.currentConfig.IVPercentile = update.IVPercentile
	}
	if update.SpreadWidth != nil {
		fm.currentConfig.SpreadWidth = update.SpreadWidth
	}
	if update.ProbOfProfit != nil {
		fm.currentConfig.ProbOfProfit = update.ProbOfProfit
	}
	if update.MaxPositions > 0 {
		fm.currentConfig.MaxPositions = update.MaxPositions
	}
	if update.RiskLimit > 0 {
		fm.currentConfig.RiskLimit = update.RiskLimit
	}
	
	// Rebuild filter chain
	fm.filterChain = filters.NewAdvancedFilterChain(fm.currentConfig, true, true)
	
	return nil
}

// recordChange records a filter change in history
func (fm *FilterManager) recordChange(change FilterChange) {
	fm.history = append(fm.history, change)
	
	// Trim history if needed
	if len(fm.history) > fm.maxHistory {
		fm.history = fm.history[len(fm.history)-fm.maxHistory:]
	}
}

// GetCurrentConfig returns the current filter configuration
func (fm *FilterManager) GetCurrentConfig() filters.FilterConfig {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	return fm.currentConfig
}

// GetFilterChain returns the current filter chain
func (fm *FilterManager) GetFilterChain() *filters.AdvancedFilterChain {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	return fm.filterChain
}

// GetPresets returns available presets
func (fm *FilterManager) GetPresets() map[string]filters.FilterConfig {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	// Return copy
	presets := make(map[string]filters.FilterConfig)
	for k, v := range fm.presets {
		presets[k] = v
	}
	
	return presets
}

// AddPreset adds a custom preset
func (fm *FilterManager) AddPreset(name string, config filters.FilterConfig) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	if _, exists := fm.presets[name]; exists {
		return fmt.Errorf("preset already exists: %s", name)
	}
	
	fm.presets[name] = config
	return nil
}

// GetHistory returns filter change history
func (fm *FilterManager) GetHistory(limit int) []FilterChange {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	if limit <= 0 || limit > len(fm.history) {
		limit = len(fm.history)
	}
	
	// Return most recent changes
	start := len(fm.history) - limit
	return fm.history[start:]
}

// SetUpdateCallback sets the callback for filter updates
func (fm *FilterManager) SetUpdateCallback(callback func(*filters.AdvancedFilterChain)) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	fm.onUpdate = callback
}

// ExportConfig exports the current configuration as JSON
func (fm *FilterManager) ExportConfig() ([]byte, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	return json.MarshalIndent(fm.currentConfig, "", "  ")
}

// ImportConfig imports a configuration from JSON
func (fm *FilterManager) ImportConfig(data []byte, clientID string) error {
	var config filters.FilterConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	previous := fm.currentConfig
	fm.currentConfig = config
	fm.filterChain = filters.NewAdvancedFilterChain(config, true, true)
	
	// Record change
	fm.recordChange(FilterChange{
		Timestamp:   time.Now(),
		ChangeType:  "import",
		Previous:    previous,
		Current:     fm.currentConfig,
		ChangedBy:   clientID,
		Description: "Imported configuration",
	})
	
	// Notify callback
	if fm.onUpdate != nil {
		fm.onUpdate(fm.filterChain)
	}
	
	return nil
}

// WebSocket integration

// ExtendWebSocketServer adds filter update handling to WebSocket server
func ExtendWebSocketServer(ws *WebSocketServer, fm *FilterManager) {
	// Add message handler for filter updates
	originalHandler := ws.subscribers
	
	// Create extended handler
	extendedHandler := func(sub *Subscriber, msg Message) {
		switch msg.Type {
		case "filter_update":
			// Parse filter update message
			data, err := json.Marshal(msg.Data)
			if err != nil {
				sub.send <- Message{
					Type:      MessageTypeError,
					Timestamp: time.Now(),
					Data:      map[string]string{"error": "Invalid filter update"},
				}
				return
			}
			
			var updateMsg FilterUpdateMessage
			if err := json.Unmarshal(data, &updateMsg); err != nil {
				sub.send <- Message{
					Type:      MessageTypeError,
					Timestamp: time.Now(),
					Data:      map[string]string{"error": "Failed to parse filter update"},
				}
				return
			}
			
			updateMsg.ClientID = sub.ID
			
			// Apply filter update
			if err := fm.HandleFilterUpdate(updateMsg); err != nil {
				sub.send <- Message{
					Type:      MessageTypeError,
					Timestamp: time.Now(),
					Data:      map[string]string{"error": err.Error()},
				}
				return
			}
			
			// Send acknowledgment
			sub.send <- Message{
				Type:      MessageTypeAck,
				Timestamp: time.Now(),
				ID:        msg.ID,
				Data: map[string]interface{}{
					"status":  "filter_updated",
					"config":  fm.GetCurrentConfig(),
				},
			}
			
			// Broadcast filter change to all subscribers
			ws.Broadcast <- Message{
				Type:      "filter_change",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"config":     fm.GetCurrentConfig(),
					"changed_by": sub.ID,
				},
			}
			
		case "get_filter_config":
			// Return current filter configuration
			sub.send <- Message{
				Type:      "filter_config",
				Timestamp: time.Now(),
				ID:        msg.ID,
				Data: map[string]interface{}{
					"config":  fm.GetCurrentConfig(),
					"presets": fm.GetPresets(),
				},
			}
			
		case "get_filter_history":
			// Return filter change history
			limit := 10
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if l, ok := data["limit"].(float64); ok {
					limit = int(l)
				}
			}
			
			sub.send <- Message{
				Type:      "filter_history",
				Timestamp: time.Now(),
				ID:        msg.ID,
				Data:      fm.GetHistory(limit),
			}
		}
	}
	
	// Replace handler (simplified - in production you'd properly extend)
	_ = originalHandler
	_ = extendedHandler
}