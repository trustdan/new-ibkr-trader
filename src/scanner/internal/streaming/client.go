package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// WebSocketClient provides a robust WebSocket client with automatic reconnection
type WebSocketClient struct {
	url             string
	conn            *websocket.Conn
	mu              sync.RWMutex
	
	// Channels
	send            chan Message
	receive         chan Message
	
	// Reconnection
	reconnectDelay  time.Duration
	maxReconnectDelay time.Duration
	reconnectAttempts int
	
	// State
	isConnected     bool
	shouldReconnect bool
	
	// Callbacks
	onConnect       func()
	onDisconnect    func(error)
	onMessage       func(Message)
	onError         func(error)
	
	// Context for shutdown
	ctx             context.Context
	cancel          context.CancelFunc
}

// ClientConfig contains client configuration
type ClientConfig struct {
	URL               string
	ReconnectDelay    time.Duration
	MaxReconnectDelay time.Duration
	OnConnect         func()
	OnDisconnect      func(error)
	OnMessage         func(Message)
	OnError           func(error)
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(config ClientConfig) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &WebSocketClient{
		url:               config.URL,
		send:              make(chan Message, 100),
		receive:           make(chan Message, 100),
		reconnectDelay:    config.ReconnectDelay,
		maxReconnectDelay: config.MaxReconnectDelay,
		shouldReconnect:   true,
		ctx:               ctx,
		cancel:           cancel,
		onConnect:         config.OnConnect,
		onDisconnect:      config.OnDisconnect,
		onMessage:         config.OnMessage,
		onError:           config.OnError,
	}
	
	// Set defaults
	if client.reconnectDelay == 0 {
		client.reconnectDelay = 1 * time.Second
	}
	if client.maxReconnectDelay == 0 {
		client.maxReconnectDelay = 1 * time.Minute
	}
	
	return client
}

// Connect establishes the WebSocket connection
func (c *WebSocketClient) Connect() error {
	return c.connect()
}

// connect performs the actual connection
func (c *WebSocketClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.isConnected {
		return nil
	}
	
	// Parse URL
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	
	// Connect
	log.Info().Str("url", c.url).Msg("Connecting to WebSocket")
	
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	
	c.conn = conn
	c.isConnected = true
	c.reconnectAttempts = 0
	
	// Start goroutines
	go c.readPump()
	go c.writePump()
	go c.processMessages()
	
	// Call connect callback
	if c.onConnect != nil {
		go c.onConnect()
	}
	
	log.Info().Msg("WebSocket connected")
	return nil
}

// Disconnect closes the WebSocket connection
func (c *WebSocketClient) Disconnect() {
	c.mu.Lock()
	c.shouldReconnect = false
	c.mu.Unlock()
	
	c.cancel()
	
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.isConnected = false
	c.mu.Unlock()
}

// Send sends a message to the server
func (c *WebSocketClient) Send(msg Message) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}
	
	select {
	case c.send <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

// Subscribe sends a subscription request
func (c *WebSocketClient) Subscribe(filters map[string]interface{}) error {
	msg := Message{
		Type:      MessageTypeSubscribe,
		Timestamp: time.Now(),
		Data:      filters,
	}
	return c.Send(msg)
}

// Unsubscribe sends an unsubscription request
func (c *WebSocketClient) Unsubscribe() error {
	msg := Message{
		Type:      MessageTypeUnsubscribe,
		Timestamp: time.Now(),
	}
	return c.Send(msg)
}

// IsConnected returns the connection status
func (c *WebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// readPump reads messages from the WebSocket
func (c *WebSocketClient) readPump() {
	defer func() {
		c.handleDisconnect(nil)
	}()
	
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			c.handleDisconnect(err)
			return
		}
		
		// Reset read deadline
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		// Send to receive channel
		select {
		case c.receive <- msg:
		case <-c.ctx.Done():
			return
		}
	}
}

// writePump writes messages to the WebSocket
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
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
			
			if err := c.conn.WriteJSON(message); err != nil {
				log.Error().Err(err).Msg("Write error")
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error().Err(err).Msg("Ping error")
				return
			}
			
		case <-c.ctx.Done():
			return
		}
	}
}

// processMessages processes received messages
func (c *WebSocketClient) processMessages() {
	for {
		select {
		case msg := <-c.receive:
			// Handle system messages
			switch msg.Type {
			case MessageTypePing:
				// Respond with pong
				pong := Message{
					Type:      MessageTypePong,
					Timestamp: time.Now(),
				}
				c.send <- pong
				
			case MessageTypeError:
				if c.onError != nil {
					if errData, ok := msg.Data.(map[string]interface{}); ok {
						if errMsg, ok := errData["error"].(string); ok {
							c.onError(fmt.Errorf(errMsg))
						}
					}
				}
				
			default:
				// Pass to user callback
				if c.onMessage != nil {
					c.onMessage(msg)
				}
			}
			
		case <-c.ctx.Done():
			return
		}
	}
}

// handleDisconnect handles disconnection and reconnection
func (c *WebSocketClient) handleDisconnect(err error) {
	c.mu.Lock()
	wasConnected := c.isConnected
	c.isConnected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	shouldReconnect := c.shouldReconnect
	c.mu.Unlock()
	
	if wasConnected {
		log.Warn().Err(err).Msg("WebSocket disconnected")
		
		if c.onDisconnect != nil {
			go c.onDisconnect(err)
		}
	}
	
	// Attempt reconnection
	if shouldReconnect {
		go c.reconnectLoop()
	}
}

// reconnectLoop attempts to reconnect with exponential backoff
func (c *WebSocketClient) reconnectLoop() {
	delay := c.reconnectDelay
	
	for {
		// Check if we should still reconnect
		c.mu.RLock()
		shouldReconnect := c.shouldReconnect
		c.mu.RUnlock()
		
		if !shouldReconnect {
			return
		}
		
		// Wait before reconnecting
		select {
		case <-time.After(delay):
		case <-c.ctx.Done():
			return
		}
		
		// Attempt to reconnect
		c.reconnectAttempts++
		log.Info().
			Int("attempt", c.reconnectAttempts).
			Dur("delay", delay).
			Msg("Attempting to reconnect")
		
		err := c.connect()
		if err == nil {
			// Success
			return
		}
		
		log.Error().Err(err).Msg("Reconnection failed")
		
		// Exponential backoff
		delay = delay * 2
		if delay > c.maxReconnectDelay {
			delay = c.maxReconnectDelay
		}
	}
}

// SubscriptionManager manages multiple subscriptions
type SubscriptionManager struct {
	client        *WebSocketClient
	subscriptions map[string]Subscription
	mu            sync.RWMutex
}

// Subscription represents a subscription configuration
type Subscription struct {
	ID       string
	Filters  map[string]interface{}
	Handler  func(ScanUpdate)
	Created  time.Time
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(client *WebSocketClient) *SubscriptionManager {
	sm := &SubscriptionManager{
		client:        client,
		subscriptions: make(map[string]Subscription),
	}
	
	// Set message handler
	client.onMessage = sm.handleMessage
	
	return sm
}

// AddSubscription adds a new subscription
func (sm *SubscriptionManager) AddSubscription(id string, filters map[string]interface{}, handler func(ScanUpdate)) error {
	sm.mu.Lock()
	sm.subscriptions[id] = Subscription{
		ID:      id,
		Filters: filters,
		Handler: handler,
		Created: time.Now(),
	}
	sm.mu.Unlock()
	
	// Send subscription to server
	return sm.client.Subscribe(filters)
}

// RemoveSubscription removes a subscription
func (sm *SubscriptionManager) RemoveSubscription(id string) {
	sm.mu.Lock()
	delete(sm.subscriptions, id)
	sm.mu.Unlock()
	
	// If no more subscriptions, unsubscribe from server
	if len(sm.subscriptions) == 0 {
		sm.client.Unsubscribe()
	}
}

// handleMessage processes incoming messages
func (sm *SubscriptionManager) handleMessage(msg Message) {
	if msg.Type != MessageTypeResult {
		return
	}
	
	// Parse scan update
	data, err := json.Marshal(msg.Data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message data")
		return
	}
	
	var update ScanUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal scan update")
		return
	}
	
	// Notify all subscriptions
	sm.mu.RLock()
	subscriptions := make([]Subscription, 0, len(sm.subscriptions))
	for _, sub := range sm.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	sm.mu.RUnlock()
	
	for _, sub := range subscriptions {
		if sub.Handler != nil {
			go sub.Handler(update)
		}
	}
}