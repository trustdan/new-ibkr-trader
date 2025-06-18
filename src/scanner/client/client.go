// Package client provides a Go SDK for the IBKR Scanner API
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// Client is the IBKR Scanner API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	wsConn     *websocket.Conn
	wsURL      string
}

// Config holds client configuration
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// DefaultConfig returns default client configuration
func DefaultConfig() Config {
	return Config{
		BaseURL: "http://localhost:8080/api/v1",
		Timeout: 30 * time.Second,
	}
}

// NewClient creates a new API client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	// Parse WebSocket URL from base URL
	u, _ := url.Parse(config.BaseURL)
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}
	wsURL := fmt.Sprintf("%s://%s/api/v1/ws", wsScheme, u.Host)
	
	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		wsURL: wsURL,
	}
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	return resp, nil
}

// decodeResponse decodes the response body
func decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return &errResp
	}
	
	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

// Health checks the service health
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.doRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	
	var health HealthResponse
	if err := decodeResponse(resp, &health); err != nil {
		return nil, err
	}
	
	return &health, nil
}

// ScanSymbol scans a single symbol
func (c *Client) ScanSymbol(ctx context.Context, symbol string, filters *ScanFilters) (*ScanResult, error) {
	path := fmt.Sprintf("/scan/%s", symbol)
	
	// Add query parameters for filters
	if filters != nil {
		params := url.Values{}
		if filters.DeltaMin > 0 {
			params.Add("delta_min", fmt.Sprintf("%f", filters.DeltaMin))
		}
		if filters.DeltaMax > 0 {
			params.Add("delta_max", fmt.Sprintf("%f", filters.DeltaMax))
		}
		if filters.DTEMin > 0 {
			params.Add("dte_min", strconv.Itoa(filters.DTEMin))
		}
		if filters.DTEMax > 0 {
			params.Add("dte_max", strconv.Itoa(filters.DTEMax))
		}
		
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	var result SuccessResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	
	// Convert data to ScanResult
	scanResult := &ScanResult{}
	if data, err := json.Marshal(result.Data); err == nil {
		json.Unmarshal(data, scanResult)
	}
	
	return scanResult, nil
}

// ScanMultiple scans multiple symbols
func (c *Client) ScanMultiple(ctx context.Context, symbols []string, filters map[string]interface{}) ([]*ScanResult, error) {
	req := ScanRequest{
		Symbols: symbols,
		Filters: filters,
	}
	
	resp, err := c.doRequest(ctx, "POST", "/scan", req)
	if err != nil {
		return nil, err
	}
	
	var result SuccessResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	
	// Convert data to []*ScanResult
	var scanResults []*ScanResult
	if data, err := json.Marshal(result.Data); err == nil {
		json.Unmarshal(data, &scanResults)
	}
	
	return scanResults, nil
}

// GetFilters returns current filter configuration
func (c *Client) GetFilters(ctx context.Context) (*FilterConfig, error) {
	resp, err := c.doRequest(ctx, "GET", "/filters", nil)
	if err != nil {
		return nil, err
	}
	
	var result SuccessResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	
	// Convert data to FilterConfig
	config := &FilterConfig{}
	if data, err := json.Marshal(result.Data); err == nil {
		json.Unmarshal(data, config)
	}
	
	return config, nil
}

// UpdateFilters updates filter configuration
func (c *Client) UpdateFilters(ctx context.Context, filters *FilterConfig) error {
	resp, err := c.doRequest(ctx, "PUT", "/filters", filters)
	if err != nil {
		return err
	}
	
	return decodeResponse(resp, nil)
}

// GetPresets returns all filter presets
func (c *Client) GetPresets(ctx context.Context) (map[string]*FilterPreset, error) {
	resp, err := c.doRequest(ctx, "GET", "/filters/presets", nil)
	if err != nil {
		return nil, err
	}
	
	var result SuccessResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	
	// Convert data to map[string]*FilterPreset
	presets := make(map[string]*FilterPreset)
	if data, err := json.Marshal(result.Data); err == nil {
		json.Unmarshal(data, &presets)
	}
	
	return presets, nil
}

// CreatePreset creates a new filter preset
func (c *Client) CreatePreset(ctx context.Context, preset *PresetRequest) (string, error) {
	resp, err := c.doRequest(ctx, "POST", "/filters/presets", preset)
	if err != nil {
		return "", err
	}
	
	var result map[string]interface{}
	if err := decodeResponse(resp, &result); err != nil {
		return "", err
	}
	
	if id, ok := result["id"].(string); ok {
		return id, nil
	}
	
	return "", fmt.Errorf("no ID returned")
}

// GetStatistics returns analytics statistics
func (c *Client) GetStatistics(ctx context.Context, symbol, metric, period string) (map[string]interface{}, error) {
	path := "/analytics/statistics"
	params := url.Values{}
	
	if symbol != "" {
		params.Add("symbol", symbol)
	}
	if metric != "" {
		params.Add("metric", metric)
	}
	if period != "" {
		params.Add("period", period)
	}
	
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	var result SuccessResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	
	if stats, ok := result.Data.(map[string]interface{}); ok {
		return stats, nil
	}
	
	return nil, fmt.Errorf("unexpected statistics format")
}

// GetHistory returns historical scan data
func (c *Client) GetHistory(ctx context.Context, params *HistoryParams) (*PaginatedResponse, error) {
	path := "/history"
	queryParams := url.Values{}
	
	if params != nil {
		if params.Symbol != "" {
			queryParams.Add("symbol", params.Symbol)
		}
		if params.StartDate != "" {
			queryParams.Add("start_date", params.StartDate)
		}
		if params.EndDate != "" {
			queryParams.Add("end_date", params.EndDate)
		}
		if params.Page > 0 {
			queryParams.Add("page", strconv.Itoa(params.Page))
		}
		if params.PageSize > 0 {
			queryParams.Add("page_size", strconv.Itoa(params.PageSize))
		}
	}
	
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}
	
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	var result PaginatedResponse
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// StreamingClient handles WebSocket streaming
type StreamingClient struct {
	client   *Client
	conn     *websocket.Conn
	handlers map[string]MessageHandler
	done     chan struct{}
}

// MessageHandler handles incoming WebSocket messages
type MessageHandler func(msgType string, payload json.RawMessage)

// Connect establishes WebSocket connection
func (c *Client) Connect(ctx context.Context) (*StreamingClient, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	
	header := http.Header{}
	if c.apiKey != "" {
		header.Set("X-API-Key", c.apiKey)
	}
	
	conn, _, err := dialer.DialContext(ctx, c.wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	
	sc := &StreamingClient{
		client:   c,
		conn:     conn,
		handlers: make(map[string]MessageHandler),
		done:     make(chan struct{}),
	}
	
	// Start reading messages
	go sc.readLoop()
	
	return sc, nil
}

// Subscribe subscribes to symbols
func (sc *StreamingClient) Subscribe(symbols []string, filters map[string]interface{}) error {
	msg := WSMessage{
		Type: "subscribe",
		Payload: map[string]interface{}{
			"symbols": symbols,
			"filters": filters,
		},
	}
	
	return sc.conn.WriteJSON(msg)
}

// Unsubscribe unsubscribes from symbols
func (sc *StreamingClient) Unsubscribe(symbols []string) error {
	msg := WSMessage{
		Type: "unsubscribe",
		Payload: map[string]interface{}{
			"symbols": symbols,
		},
	}
	
	return sc.conn.WriteJSON(msg)
}

// OnMessage registers a message handler
func (sc *StreamingClient) OnMessage(msgType string, handler MessageHandler) {
	sc.handlers[msgType] = handler
}

// Close closes the WebSocket connection
func (sc *StreamingClient) Close() error {
	close(sc.done)
	return sc.conn.Close()
}

// readLoop reads messages from WebSocket
func (sc *StreamingClient) readLoop() {
	defer sc.conn.Close()
	
	for {
		select {
		case <-sc.done:
			return
		default:
			var msg WSResponse
			if err := sc.conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					// Log error
				}
				return
			}
			
			// Call handler if registered
			if handler, ok := sc.handlers[msg.Type]; ok {
				handler(msg.Type, msg.Payload)
			}
		}
	}
}