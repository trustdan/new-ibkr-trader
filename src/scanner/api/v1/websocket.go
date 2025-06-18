package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// handleWebSocket handles WebSocket connections for real-time streaming
func (api *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract client info
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = generateClientID()
	}
	
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	
	// Create client
	client := &wsClient{
		id:        clientID,
		conn:      conn,
		send:      make(chan []byte, 256),
		api:       api,
		filters:   make(map[string]interface{}),
		isClosing: false,
	}
	
	// Register client with streamer
	api.streamer.Register(client)
	
	// Start client goroutines
	go client.writePump()
	go client.readPump()
	
	// Send welcome message
	welcome := map[string]interface{}{
		"type":      "welcome",
		"client_id": clientID,
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"features": []string{
			"real_time_scans",
			"filter_updates",
			"batch_results",
			"performance_metrics",
		},
	}
	
	if data, err := json.Marshal(welcome); err == nil {
		client.send <- data
	}
}

// wsClient represents a WebSocket client
type wsClient struct {
	id        string
	conn      *websocket.Conn
	send      chan []byte
	api       *API
	filters   map[string]interface{}
	isClosing bool
}

// readPump handles incoming messages from the client
func (c *wsClient) readPump() {
	defer func() {
		c.api.streamer.Unregister(c)
		c.conn.Close()
	}()
	
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("client_id", c.id).Msg("WebSocket read error")
			}
			break
		}
		
		// Process message
		c.handleMessage(message)
	}
}

// writePump handles outgoing messages to the client
func (c *wsClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages to current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *wsClient) handleMessage(message []byte) {
	var msg wsMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError("Invalid message format")
		return
	}
	
	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg.Payload)
		
	case "unsubscribe":
		c.handleUnsubscribe(msg.Payload)
		
	case "update_filters":
		c.handleUpdateFilters(msg.Payload)
		
	case "scan":
		c.handleScan(msg.Payload)
		
	case "ping":
		c.sendPong()
		
	default:
		c.sendError("Unknown message type: " + msg.Type)
	}
}

// handleSubscribe handles subscription requests
func (c *wsClient) handleSubscribe(payload json.RawMessage) {
	var sub struct {
		Symbols []string               `json:"symbols"`
		Filters map[string]interface{} `json:"filters,omitempty"`
	}
	
	if err := json.Unmarshal(payload, &sub); err != nil {
		c.sendError("Invalid subscription data")
		return
	}
	
	// Update client filters
	if sub.Filters != nil {
		c.filters = sub.Filters
	}
	
	// Subscribe to symbols
	for _, symbol := range sub.Symbols {
		c.api.streamer.Subscribe(c.id, symbol)
	}
	
	// Send confirmation
	c.sendMessage("subscribed", map[string]interface{}{
		"symbols": sub.Symbols,
		"filters": c.filters,
	})
}

// handleUnsubscribe handles unsubscription requests
func (c *wsClient) handleUnsubscribe(payload json.RawMessage) {
	var unsub struct {
		Symbols []string `json:"symbols"`
	}
	
	if err := json.Unmarshal(payload, &unsub); err != nil {
		c.sendError("Invalid unsubscription data")
		return
	}
	
	// Unsubscribe from symbols
	for _, symbol := range unsub.Symbols {
		c.api.streamer.Unsubscribe(c.id, symbol)
	}
	
	// Send confirmation
	c.sendMessage("unsubscribed", map[string]interface{}{
		"symbols": unsub.Symbols,
	})
}

// handleUpdateFilters handles filter update requests
func (c *wsClient) handleUpdateFilters(payload json.RawMessage) {
	var filters map[string]interface{}
	
	if err := json.Unmarshal(payload, &filters); err != nil {
		c.sendError("Invalid filter data")
		return
	}
	
	// Update client filters
	c.filters = filters
	
	// Send confirmation
	c.sendMessage("filters_updated", c.filters)
}

// handleScan handles real-time scan requests
func (c *wsClient) handleScan(payload json.RawMessage) {
	var scanReq struct {
		Symbol  string                 `json:"symbol"`
		Filters map[string]interface{} `json:"filters,omitempty"`
	}
	
	if err := json.Unmarshal(payload, &scanReq); err != nil {
		c.sendError("Invalid scan request")
		return
	}
	
	// Use client filters if none provided
	if scanReq.Filters == nil {
		scanReq.Filters = c.filters
	}
	
	// Perform scan
	go func() {
		result, err := c.api.scanner.ScanSymbol(nil, scanReq.Symbol, scanReq.Filters)
		if err != nil {
			c.sendError(fmt.Sprintf("Scan failed: %v", err))
			return
		}
		
		// Send result
		c.sendMessage("scan_result", result)
	}()
}

// Helper methods
func (c *wsClient) sendMessage(msgType string, payload interface{}) {
	msg := map[string]interface{}{
		"type":      msgType,
		"payload":   payload,
		"timestamp": time.Now().Unix(),
	}
	
	if data, err := json.Marshal(msg); err == nil {
		select {
		case c.send <- data:
		default:
			// Client's send channel is full, close connection
			c.conn.Close()
		}
	}
}

func (c *wsClient) sendError(error string) {
	c.sendMessage("error", map[string]string{"message": error})
}

func (c *wsClient) sendPong() {
	c.sendMessage("pong", map[string]int64{"timestamp": time.Now().Unix()})
}

// Send implements the streaming.Client interface
func (c *wsClient) Send(data []byte) error {
	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("client send buffer full")
	}
}

// GetID implements the streaming.Client interface
func (c *wsClient) GetID() string {
	return c.id
}

// Close implements the streaming.Client interface
func (c *wsClient) Close() error {
	if !c.isClosing {
		c.isClosing = true
		close(c.send)
	}
	return nil
}

// wsMessage represents a WebSocket message
type wsMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("ws-%d-%s", time.Now().Unix(), generateRandomString(8))
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}