package alerts

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/ibkr-trader/scanner/internal/streaming"
	"github.com/rs/zerolog/log"
)

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeNewOpportunity    AlertType = "new_opportunity"
	AlertTypePriceChange       AlertType = "price_change"
	AlertTypeVolumeSpike       AlertType = "volume_spike"
	AlertTypeVolatilityChange  AlertType = "volatility_change"
	AlertTypeThresholdCrossed  AlertType = "threshold_crossed"
)

// Alert represents a trading alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Symbol      string                 `json:"symbol"`
	Message     string                 `json:"message"`
	Severity    string                 `json:"severity"` // info, warning, critical
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	Acknowledged bool                  `json:"acknowledged"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        AlertType              `json:"type"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     []AlertAction          `json:"actions"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AlertAction defines what to do when an alert is triggered
type AlertAction struct {
	Type   string                 `json:"type"` // webhook, email, sms, websocket
	Config map[string]interface{} `json:"config"`
}

// AlertManager manages alert rules and notifications
type AlertManager struct {
	rules          map[string]*AlertRule
	alerts         map[string]*Alert
	alertHistory   []Alert
	mu             sync.RWMutex
	
	// Channels
	alertChan      chan Alert
	
	// Handlers
	handlers       map[string]AlertHandler
	
	// WebSocket integration
	wsServer       *streaming.WebSocketServer
	
	// Configuration
	maxHistory     int
	alertThrottle  time.Duration
	lastAlertTime  map[string]time.Time
}

// AlertHandler processes alerts
type AlertHandler interface {
	HandleAlert(alert Alert) error
	Type() string
}

// NewAlertManager creates a new alert manager
func NewAlertManager(wsServer *streaming.WebSocketServer) *AlertManager {
	am := &AlertManager{
		rules:         make(map[string]*AlertRule),
		alerts:        make(map[string]*Alert),
		alertHistory:  make([]Alert, 0),
		alertChan:     make(chan Alert, 100),
		handlers:      make(map[string]AlertHandler),
		wsServer:      wsServer,
		maxHistory:    1000,
		alertThrottle: 30 * time.Second,
		lastAlertTime: make(map[string]time.Time),
	}
	
	// Register default handlers
	am.RegisterHandler(NewWebSocketHandler(wsServer))
	am.RegisterHandler(NewLogHandler())
	
	return am
}

// Start begins alert processing
func (am *AlertManager) Start(ctx context.Context) {
	go am.processAlerts(ctx)
}

// RegisterHandler registers an alert handler
func (am *AlertManager) RegisterHandler(handler AlertHandler) {
	am.mu.Lock()
	am.handlers[handler.Type()] = handler
	am.mu.Unlock()
}

// AddRule adds an alert rule
func (am *AlertManager) AddRule(rule AlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID required")
	}
	
	am.mu.Lock()
	am.rules[rule.ID] = &rule
	am.mu.Unlock()
	
	log.Info().Str("rule_id", rule.ID).Str("name", rule.Name).Msg("Alert rule added")
	return nil
}

// RemoveRule removes an alert rule
func (am *AlertManager) RemoveRule(ruleID string) {
	am.mu.Lock()
	delete(am.rules, ruleID)
	am.mu.Unlock()
	
	log.Info().Str("rule_id", ruleID).Msg("Alert rule removed")
}

// CheckScanResult checks a scan result against alert rules
func (am *AlertManager) CheckScanResult(result models.ScanResult) {
	am.mu.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mu.RUnlock()
	
	// Check each rule
	for _, rule := range rules {
		alerts := am.evaluateRule(rule, result)
		for _, alert := range alerts {
			am.TriggerAlert(alert)
		}
	}
}

// evaluateRule evaluates a rule against scan results
func (am *AlertManager) evaluateRule(rule *AlertRule, result models.ScanResult) []Alert {
	alerts := make([]Alert, 0)
	
	switch rule.Type {
	case AlertTypeNewOpportunity:
		if len(result.Spreads) > 0 {
			// Check if spreads meet criteria
			minScore, _ := rule.Conditions["min_score"].(float64)
			minPoP, _ := rule.Conditions["min_pop"].(float64)
			
			for _, spread := range result.Spreads {
				if spread.Score >= minScore && spread.ProbOfProfit >= minPoP {
					alert := Alert{
						ID:       fmt.Sprintf("%s-%s-%d", rule.ID, result.Symbol, time.Now().Unix()),
						Type:     rule.Type,
						Symbol:   result.Symbol,
						Message:  fmt.Sprintf("New opportunity: %s spread with %.2f credit, %.1f%% PoP",
							result.Symbol, spread.Credit, spread.ProbOfProfit*100),
						Severity: "info",
						Data: map[string]interface{}{
							"spread": spread,
							"score":  spread.Score,
						},
						CreatedAt: time.Now(),
					}
					alerts = append(alerts, alert)
					break // One alert per symbol
				}
			}
		}
		
	case AlertTypeThresholdCrossed:
		// Check various thresholds
		if maxSpreads, ok := rule.Conditions["max_spreads"].(float64); ok {
			if float64(len(result.Spreads)) > maxSpreads {
				alert := Alert{
					ID:       fmt.Sprintf("%s-%s-%d", rule.ID, result.Symbol, time.Now().Unix()),
					Type:     rule.Type,
					Symbol:   result.Symbol,
					Message:  fmt.Sprintf("High opportunity count: %d spreads found for %s",
						len(result.Spreads), result.Symbol),
					Severity: "warning",
					Data: map[string]interface{}{
						"spread_count": len(result.Spreads),
						"threshold":    maxSpreads,
					},
					CreatedAt: time.Now(),
				}
				alerts = append(alerts, alert)
			}
		}
	}
	
	return alerts
}

// TriggerAlert triggers an alert
func (am *AlertManager) TriggerAlert(alert Alert) {
	// Check throttling
	throttleKey := fmt.Sprintf("%s-%s", alert.Type, alert.Symbol)
	am.mu.RLock()
	lastTime, exists := am.lastAlertTime[throttleKey]
	am.mu.RUnlock()
	
	if exists && time.Since(lastTime) < am.alertThrottle {
		return // Skip due to throttling
	}
	
	// Update last alert time
	am.mu.Lock()
	am.lastAlertTime[throttleKey] = time.Now()
	am.mu.Unlock()
	
	// Send to processing channel
	select {
	case am.alertChan <- alert:
	default:
		log.Warn().Str("alert_id", alert.ID).Msg("Alert channel full, dropping alert")
	}
}

// processAlerts processes alerts
func (am *AlertManager) processAlerts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
			
		case alert := <-am.alertChan:
			am.handleAlert(alert)
		}
	}
}

// handleAlert processes a single alert
func (am *AlertManager) handleAlert(alert Alert) {
	// Store alert
	am.mu.Lock()
	am.alerts[alert.ID] = &alert
	am.alertHistory = append(am.alertHistory, alert)
	
	// Trim history
	if len(am.alertHistory) > am.maxHistory {
		am.alertHistory = am.alertHistory[len(am.alertHistory)-am.maxHistory:]
	}
	am.mu.Unlock()
	
	// Get rule to determine actions
	am.mu.RLock()
	var rule *AlertRule
	for _, r := range am.rules {
		if r.Type == alert.Type {
			rule = r
			break
		}
	}
	am.mu.RUnlock()
	
	if rule == nil {
		// Use default actions
		rule = &AlertRule{
			Actions: []AlertAction{
				{Type: "websocket"},
				{Type: "log"},
			},
		}
	}
	
	// Execute actions
	for _, action := range rule.Actions {
		if handler, exists := am.handlers[action.Type]; exists {
			if err := handler.HandleAlert(alert); err != nil {
				log.Error().Err(err).
					Str("alert_id", alert.ID).
					Str("handler", action.Type).
					Msg("Failed to handle alert")
			}
		}
	}
}

// GetAlerts returns current alerts
func (am *AlertManager) GetAlerts(acknowledged bool) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	alerts := make([]Alert, 0)
	for _, alert := range am.alerts {
		if alert.Acknowledged == acknowledged {
			alerts = append(alerts, *alert)
		}
	}
	
	return alerts
}

// AcknowledgeAlert acknowledges an alert
func (am *AlertManager) AcknowledgeAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found")
	}
	
	alert.Acknowledged = true
	return nil
}

// GetAlertHistory returns alert history
func (am *AlertManager) GetAlertHistory(limit int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	if limit <= 0 || limit > len(am.alertHistory) {
		limit = len(am.alertHistory)
	}
	
	// Return most recent alerts
	start := len(am.alertHistory) - limit
	return am.alertHistory[start:]
}

// Alert Handlers

// WebSocketHandler sends alerts via WebSocket
type WebSocketHandler struct {
	wsServer *streaming.WebSocketServer
}

func NewWebSocketHandler(wsServer *streaming.WebSocketServer) *WebSocketHandler {
	return &WebSocketHandler{wsServer: wsServer}
}

func (h *WebSocketHandler) Type() string {
	return "websocket"
}

func (h *WebSocketHandler) HandleAlert(alert Alert) error {
	// Create alert message
	msg := streaming.Message{
		Type:      "alert",
		Timestamp: time.Now(),
		Data:      alert,
	}
	
	// Broadcast to all connected clients
	h.wsServer.Broadcast <- msg
	
	return nil
}

// LogHandler logs alerts
type LogHandler struct{}

func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

func (h *LogHandler) Type() string {
	return "log"
}

func (h *LogHandler) HandleAlert(alert Alert) error {
	log.Info().
		Str("alert_id", alert.ID).
		Str("type", string(alert.Type)).
		Str("symbol", alert.Symbol).
		Str("severity", alert.Severity).
		Str("message", alert.Message).
		Msg("Alert triggered")
	
	return nil
}

// PresetAlertRules provides common alert rule configurations
func PresetAlertRules() []AlertRule {
	return []AlertRule{
		{
			ID:   "high-score-opportunities",
			Name: "High Score Opportunities",
			Type: AlertTypeNewOpportunity,
			Conditions: map[string]interface{}{
				"min_score": 0.80,
				"min_pop":   0.70,
			},
			Actions: []AlertAction{
				{Type: "websocket"},
				{Type: "log"},
			},
			Enabled: true,
		},
		{
			ID:   "many-spreads-found",
			Name: "Many Spreads Found",
			Type: AlertTypeThresholdCrossed,
			Conditions: map[string]interface{}{
				"max_spreads": 20,
			},
			Actions: []AlertAction{
				{Type: "websocket"},
				{Type: "log"},
			},
			Enabled: true,
		},
	}
}