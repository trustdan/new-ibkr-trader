package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"github.com/ibkr-trader/scanner/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

// WebSocketServer handles real-time scanner updates via WebSocket
type WebSocketServer struct {
	upgrader    websocket.Upgrader
	subscribers map[string]*Subscriber
	mu          sync.RWMutex
	
	// Channels
	broadcast    chan Message
	register     chan *Subscriber
	unregister   chan *Subscriber
	
	// Metrics
	metrics      *StreamingMetrics
	
	// Configuration
	pingInterval time.Duration
	pongTimeout  time.Duration
	writeTimeout time.Duration
	maxMessageSize int64
}

// Subscriber represents a WebSocket client
type Subscriber struct {
	ID           string
	conn         *websocket.Conn
	send         chan Message
	filters      map[string]interface{} // Client-specific filters
	lastActivity time.Time
	mu           sync.RWMutex
}

// Message types for WebSocket communication
type Message struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	ID        string      `json:"id,omitempty"`
}

// MessageType constants
const (
	MessageTypeResult       = "result"
	MessageTypeStatus       = "status"
	MessageTypeError        = "error"
	MessageTypeSubscribe    = "subscribe"
	MessageTypeUnsubscribe  = "unsubscribe"
	MessageTypePing         = "ping"
	MessageTypePong         = "pong"
	MessageTypeAck          = "ack"
)

// ScanUpdate represents a real-time scan update
type ScanUpdate struct {
	ScanID    string                   `json:"scan_id"`
	Symbol    string                   `json:"symbol"`
	Spreads   []models.VerticalSpread  `json:"spreads,omitempty"`
	Contracts []models.OptionContract  `json:"contracts,omitempty"`
	UpdateType string                  `json:"update_type"` // "new", "update", "remove"
	Metadata  map[string]interface{}   `json:"metadata,omitempty"`
}

// StatusUpdate represents a status change
type StatusUpdate struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Progress  int    `json:"progress,omitempty"`
	Total     int    `json:"total,omitempty"`
}

// StreamingMetrics tracks WebSocket metrics
type StreamingMetrics struct {
	activeConnections   prometheus.Gauge
	messagesSent        *prometheus.CounterVec
	messagesReceived    *prometheus.CounterVec
	connectionDuration  *prometheus.HistogramVec
	messageLatency      *prometheus.HistogramVec
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer() *WebSocketServer {
	ws := &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking for production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		subscribers:    make(map[string]*Subscriber),
		broadcast:      make(chan Message, 100),
		register:       make(chan *Subscriber),
		unregister:     make(chan *Subscriber),
		pingInterval:   30 * time.Second,
		pongTimeout:    60 * time.Second,
		writeTimeout:   10 * time.Second,
		maxMessageSize: 512 * 1024, // 512KB
	}
	
	ws.initMetrics()
	return ws
}

// initMetrics initializes Prometheus metrics
func (ws *WebSocketServer) initMetrics() {
	ws.metrics = &StreamingMetrics{
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "websocket_active_connections",
				Help: "Number of active WebSocket connections",
			},
		),
		messagesSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_sent_total",
				Help: "Total number of messages sent",
			},
			[]string{"type"},
		),
		messagesReceived: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_received_total",
				Help: "Total number of messages received",
			},
			[]string{"type"},
		),
		connectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "websocket_connection_duration_seconds",
				Help: "WebSocket connection duration",
			},
			[]string{"status"},
		),
		messageLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "websocket_message_latency_seconds",
				Help: "Message delivery latency",
			},
			[]string{"type"},
		),
	}
}

// Start begins the WebSocket server
func (ws *WebSocketServer) Start(ctx context.Context) {
	go ws.run(ctx)
}

// run is the main event loop
func (ws *WebSocketServer) run(ctx context.Context) {
	ticker := time.NewTicker(ws.pingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			ws.shutdown()
			return
			
		case subscriber := <-ws.register:
			ws.addSubscriber(subscriber)
			
		case subscriber := <-ws.unregister:
			ws.removeSubscriber(subscriber)
			
		case message := <-ws.broadcast:
			ws.broadcastMessage(message)
			
		case <-ticker.C:
			ws.pingClients()
		}
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (ws *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}
	
	subscriber := &Subscriber{
		ID:           uuid.New().String(),
		conn:         conn,
		send:         make(chan Message, 256),
		filters:      make(map[string]interface{}),
		lastActivity: time.Now(),
	}
	
	// Configure connection
	conn.SetReadLimit(ws.maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(ws.pongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(ws.pongTimeout))
		subscriber.updateActivity()
		return nil
	})
	
	// Register subscriber
	ws.register <- subscriber
	
	// Start goroutines
	go subscriber.writePump(ws)
	go subscriber.readPump(ws)
	
	// Send welcome message
	welcome := Message{
		Type:      MessageTypeStatus,
		Timestamp: time.Now(),
		Data: StatusUpdate{
			Status:  "connected",
			Message: fmt.Sprintf("Connected with ID: %s", subscriber.ID),
		},
	}
	subscriber.send <- welcome
}

// addSubscriber adds a new subscriber
func (ws *WebSocketServer) addSubscriber(sub *Subscriber) {
	ws.mu.Lock()
	ws.subscribers[sub.ID] = sub
	ws.mu.Unlock()
	
	ws.metrics.activeConnections.Inc()
	log.Info().Str("subscriber_id", sub.ID).Msg("New subscriber connected")
}

// removeSubscriber removes a subscriber
func (ws *WebSocketServer) removeSubscriber(sub *Subscriber) {
	ws.mu.Lock()
	if _, exists := ws.subscribers[sub.ID]; exists {
		delete(ws.subscribers, sub.ID)
		close(sub.send)
	}
	ws.mu.Unlock()
	
	ws.metrics.activeConnections.Dec()
	log.Info().Str("subscriber_id", sub.ID).Msg("Subscriber disconnected")
}

// broadcastMessage sends a message to all subscribers
func (ws *WebSocketServer) broadcastMessage(msg Message) {
	ws.mu.RLock()
	subscribers := make([]*Subscriber, 0, len(ws.subscribers))
	for _, sub := range ws.subscribers {
		subscribers = append(subscribers, sub)
	}
	ws.mu.RUnlock()
	
	for _, sub := range subscribers {
		select {
		case sub.send <- msg:
			ws.metrics.messagesSent.WithLabelValues(msg.Type).Inc()
		default:
			// Client's send channel is full, close it
			log.Warn().Str("subscriber_id", sub.ID).Msg("Closing slow subscriber")
			ws.unregister <- sub
		}
	}
}

// BroadcastScanUpdate broadcasts a scan update to all subscribers
func (ws *WebSocketServer) BroadcastScanUpdate(update ScanUpdate) {
	msg := Message{
		Type:      MessageTypeResult,
		Timestamp: time.Now(),
		Data:      update,
	}
	ws.broadcast <- msg
}

// BroadcastStatus broadcasts a status update
func (ws *WebSocketServer) BroadcastStatus(status StatusUpdate) {
	msg := Message{
		Type:      MessageTypeStatus,
		Timestamp: time.Now(),
		Data:      status,
	}
	ws.broadcast <- msg
}

// BroadcastError broadcasts an error
func (ws *WebSocketServer) BroadcastError(err error) {
	msg := Message{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Data: map[string]string{
			"error": err.Error(),
		},
	}
	ws.broadcast <- msg
}

// pingClients sends ping messages to all clients
func (ws *WebSocketServer) pingClients() {
	ping := Message{
		Type:      MessageTypePing,
		Timestamp: time.Now(),
	}
	
	ws.mu.RLock()
	for _, sub := range ws.subscribers {
		select {
		case sub.send <- ping:
		default:
			// Skip if channel is full
		}
	}
	ws.mu.RUnlock()
}

// shutdown closes all connections
func (ws *WebSocketServer) shutdown() {
	ws.mu.Lock()
	for _, sub := range ws.subscribers {
		sub.conn.Close()
		close(sub.send)
	}
	ws.subscribers = make(map[string]*Subscriber)
	ws.mu.Unlock()
}

// Subscriber methods

// readPump reads messages from the WebSocket connection
func (s *Subscriber) readPump(ws *WebSocketServer) {
	defer func() {
		ws.unregister <- s
		s.conn.Close()
	}()
	
	for {
		var msg Message
		err := s.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("subscriber_id", s.ID).Msg("WebSocket read error")
			}
			break
		}
		
		s.updateActivity()
		ws.metrics.messagesReceived.WithLabelValues(msg.Type).Inc()
		
		// Handle different message types
		switch msg.Type {
		case MessageTypeSubscribe:
			s.handleSubscribe(msg, ws)
		case MessageTypeUnsubscribe:
			s.handleUnsubscribe(msg, ws)
		case MessageTypePong:
			// Already handled by SetPongHandler
		default:
			log.Warn().Str("type", msg.Type).Msg("Unknown message type")
		}
	}
}

// writePump writes messages to the WebSocket connection
func (s *Subscriber) writePump(ws *WebSocketServer) {
	ticker := time.NewTicker(ws.pingInterval)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
			if !ok {
				// Channel closed
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := s.conn.WriteJSON(message); err != nil {
				return
			}
			
			// Write queued messages
			n := len(s.send)
			for i := 0; i < n; i++ {
				if err := s.conn.WriteJSON(<-s.send); err != nil {
					return
				}
			}
			
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe handles subscription requests
func (s *Subscriber) handleSubscribe(msg Message, ws *WebSocketServer) {
	// Parse subscription filters
	if data, ok := msg.Data.(map[string]interface{}); ok {
		s.mu.Lock()
		for k, v := range data {
			s.filters[k] = v
		}
		s.mu.Unlock()
		
		// Send acknowledgment
		ack := Message{
			Type:      MessageTypeAck,
			Timestamp: time.Now(),
			ID:        msg.ID,
			Data: map[string]string{
				"status": "subscribed",
			},
		}
		s.send <- ack
	}
}

// handleUnsubscribe handles unsubscription requests
func (s *Subscriber) handleUnsubscribe(msg Message, ws *WebSocketServer) {
	s.mu.Lock()
	s.filters = make(map[string]interface{})
	s.mu.Unlock()
	
	// Send acknowledgment
	ack := Message{
		Type:      MessageTypeAck,
		Timestamp: time.Now(),
		ID:        msg.ID,
		Data: map[string]string{
			"status": "unsubscribed",
		},
	}
	s.send <- ack
}

// updateActivity updates the last activity timestamp
func (s *Subscriber) updateActivity() {
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()
}

// GetFilters returns the subscriber's filters
func (s *Subscriber) GetFilters() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	filters := make(map[string]interface{})
	for k, v := range s.filters {
		filters[k] = v
	}
	return filters
}