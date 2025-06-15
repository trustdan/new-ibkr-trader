package streaming

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ibkr-trader/scanner/internal/filters"
	"github.com/ibkr-trader/scanner/internal/models"
)

// MockScanner is a mock scanner for testing
type MockScanner struct {
	mock.Mock
}

func (m *MockScanner) GetOptionChain(ctx context.Context, symbol string) ([]models.OptionContract, error) {
	args := m.Called(ctx, symbol)
	if contracts, ok := args.Get(0).([]models.OptionContract); ok {
		return contracts, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScanner) FindVerticalSpreads(contracts []models.OptionContract) []models.VerticalSpread {
	args := m.Called(contracts)
	if spreads, ok := args.Get(0).([]models.VerticalSpread); ok {
		return spreads
	}
	return nil
}

func (m *MockScanner) ScanSymbol(ctx context.Context, symbol string, config interface{}) (models.ScanResult, error) {
	args := m.Called(ctx, symbol, config)
	return args.Get(0).(models.ScanResult), args.Error(1)
}

// Test WebSocket Server
func TestWebSocketServer(t *testing.T) {
	t.Run("Basic Connection", func(t *testing.T) {
		// Create server
		wsServer := NewWebSocketServer()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		wsServer.Start(ctx)
		
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(wsServer.HandleWebSocket))
		defer server.Close()
		
		// Connect client
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer conn.Close()
		
		// Wait for welcome message
		var msg Message
		err = conn.ReadJSON(&msg)
		assert.NoError(t, err)
		assert.Equal(t, MessageTypeStatus, msg.Type)
	})
	
	t.Run("Broadcast Messages", func(t *testing.T) {
		wsServer := NewWebSocketServer()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		wsServer.Start(ctx)
		
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(wsServer.HandleWebSocket))
		defer server.Close()
		
		// Connect multiple clients
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		
		client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer client1.Close()
		
		client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer client2.Close()
		
		// Wait for connections to establish
		time.Sleep(100 * time.Millisecond)
		
		// Broadcast a message
		update := ScanUpdate{
			ScanID:     "test-scan",
			Symbol:     "TEST",
			UpdateType: "new",
		}
		wsServer.BroadcastScanUpdate(update)
		
		// Both clients should receive the message
		var msg1, msg2 Message
		
		// Skip welcome messages
		client1.ReadJSON(&msg1)
		client2.ReadJSON(&msg2)
		
		// Read broadcast message
		err = client1.ReadJSON(&msg1)
		assert.NoError(t, err)
		assert.Equal(t, MessageTypeResult, msg1.Type)
		
		err = client2.ReadJSON(&msg2)
		assert.NoError(t, err)
		assert.Equal(t, MessageTypeResult, msg2.Type)
	})
}

// Test Streaming Scanner
func TestStreamingScanner(t *testing.T) {
	t.Run("Continuous Scanning", func(t *testing.T) {
		// Create mocks
		mockScanner := new(MockScanner)
		wsServer := NewWebSocketServer()
		filterChain := filters.NewAdvancedFilterChain(filters.FilterConfig{}, false, false)
		
		// Create streaming scanner
		scanner := NewStreamingScanner(mockScanner, wsServer, filterChain)
		scanner.scanInterval = 100 * time.Millisecond // Fast for testing
		
		// Setup mock expectations
		contracts := []models.OptionContract{
			{Symbol: "TEST", Strike: 100, Delta: 0.30},
			{Symbol: "TEST", Strike: 105, Delta: 0.25},
		}
		
		spreads := []models.VerticalSpread{
			{
				Symbol:    "TEST",
				ShortLeg:  contracts[0],
				LongLeg:   contracts[1],
				Credit:    1.50,
			},
		}
		
		mockScanner.On("GetOptionChain", mock.Anything, "TEST").Return(contracts, nil)
		mockScanner.On("FindVerticalSpreads", mock.Anything).Return(spreads)
		
		// Start scanning
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		wsServer.Start(ctx)
		err := scanner.Start(ctx, []string{"TEST"})
		assert.NoError(t, err)
		
		// Wait for scans
		time.Sleep(300 * time.Millisecond)
		
		// Verify scans occurred
		stats := scanner.GetStats()
		assert.True(t, stats["scan_count"].(int64) >= 2)
		
		// Stop scanning
		scanner.Stop()
	})
	
	t.Run("Result Deduplication", func(t *testing.T) {
		cache := NewResultCache(1*time.Minute, 100)
		
		result1 := models.ScanResult{
			Symbol: "TEST",
			Spreads: []models.VerticalSpread{
				{
					Symbol:   "TEST",
					ShortLeg: models.OptionContract{Strike: 100},
					LongLeg:  models.OptionContract{Strike: 105},
					Credit:   1.50,
				},
			},
		}
		
		// First check - not duplicate
		assert.False(t, cache.IsDuplicate(result1))
		cache.Store(result1)
		
		// Second check - is duplicate
		assert.True(t, cache.IsDuplicate(result1))
		
		// Different result - not duplicate
		result2 := result1
		result2.Spreads[0].Credit = 1.60
		assert.False(t, cache.IsDuplicate(result2))
	})
}

// Test WebSocket Client
func TestWebSocketClient(t *testing.T) {
	t.Run("Auto Reconnection", func(t *testing.T) {
		// Create a test server that we can control
		var serverConn *websocket.Conn
		serverConnected := make(chan bool, 10)
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			
			serverConn = conn
			serverConnected <- true
			
			// Keep connection open
			for {
				var msg Message
				if err := conn.ReadJSON(&msg); err != nil {
					break
				}
			}
		}))
		defer server.Close()
		
		// Create client
		connectCount := 0
		disconnectCount := 0
		
		config := ClientConfig{
			URL:            "ws" + strings.TrimPrefix(server.URL, "http"),
			ReconnectDelay: 50 * time.Millisecond,
			OnConnect: func() {
				connectCount++
			},
			OnDisconnect: func(err error) {
				disconnectCount++
			},
		}
		
		client := NewWebSocketClient(config)
		
		// Connect
		err := client.Connect()
		assert.NoError(t, err)
		
		// Wait for connection
		<-serverConnected
		time.Sleep(100 * time.Millisecond)
		
		assert.Equal(t, 1, connectCount)
		assert.True(t, client.IsConnected())
		
		// Force disconnect from server side
		serverConn.Close()
		
		// Wait for reconnection
		<-serverConnected
		time.Sleep(100 * time.Millisecond)
		
		assert.Equal(t, 2, connectCount)
		assert.Equal(t, 1, disconnectCount)
		assert.True(t, client.IsConnected())
		
		// Clean disconnect
		client.Disconnect()
	})
	
	t.Run("Message Handling", func(t *testing.T) {
		wsServer := NewWebSocketServer()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		wsServer.Start(ctx)
		
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(wsServer.HandleWebSocket))
		defer server.Close()
		
		// Create client
		receivedMessages := make([]Message, 0)
		
		config := ClientConfig{
			URL: "ws" + strings.TrimPrefix(server.URL, "http"),
			OnMessage: func(msg Message) {
				receivedMessages = append(receivedMessages, msg)
			},
		}
		
		client := NewWebSocketClient(config)
		err := client.Connect()
		assert.NoError(t, err)
		
		// Wait for connection
		time.Sleep(100 * time.Millisecond)
		
		// Send subscription
		err = client.Subscribe(map[string]interface{}{
			"symbols": []string{"TEST"},
		})
		assert.NoError(t, err)
		
		// Broadcast from server
		wsServer.BroadcastScanUpdate(ScanUpdate{
			ScanID:     "test-123",
			Symbol:     "TEST",
			UpdateType: "new",
		})
		
		// Wait for message
		time.Sleep(100 * time.Millisecond)
		
		// Check received messages
		found := false
		for _, msg := range receivedMessages {
			if msg.Type == MessageTypeResult {
				found = true
				break
			}
		}
		assert.True(t, found)
		
		client.Disconnect()
	})
}

// Test Subscription Manager
func TestSubscriptionManager(t *testing.T) {
	t.Run("Multiple Subscriptions", func(t *testing.T) {
		// Create mock client
		client := &WebSocketClient{
			send: make(chan Message, 10),
		}
		
		manager := NewSubscriptionManager(client)
		
		// Add subscriptions
		received1 := make([]ScanUpdate, 0)
		received2 := make([]ScanUpdate, 0)
		
		err := manager.AddSubscription("sub1", map[string]interface{}{
			"symbol": "TEST1",
		}, func(update ScanUpdate) {
			received1 = append(received1, update)
		})
		assert.NoError(t, err)
		
		err = manager.AddSubscription("sub2", map[string]interface{}{
			"symbol": "TEST2",
		}, func(update ScanUpdate) {
			received2 = append(received2, update)
		})
		assert.NoError(t, err)
		
		// Simulate incoming message
		update := ScanUpdate{
			ScanID:     "scan-123",
			Symbol:     "TEST1",
			UpdateType: "new",
		}
		
		msg := Message{
			Type:      MessageTypeResult,
			Timestamp: time.Now(),
			Data:      update,
		}
		
		manager.handleMessage(msg)
		
		// Wait for handlers
		time.Sleep(50 * time.Millisecond)
		
		// Both subscriptions should receive the update
		assert.Len(t, received1, 1)
		assert.Len(t, received2, 1)
		
		// Remove subscription
		manager.RemoveSubscription("sub1")
		assert.Len(t, manager.subscriptions, 1)
	})
}

// Benchmark tests
func BenchmarkWebSocketBroadcast(b *testing.B) {
	wsServer := NewWebSocketServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	wsServer.Start(ctx)
	
	// Add mock subscribers
	for i := 0; i < 100; i++ {
		sub := &Subscriber{
			ID:   fmt.Sprintf("sub-%d", i),
			send: make(chan Message, 256),
		}
		wsServer.subscribers[sub.ID] = sub
	}
	
	update := ScanUpdate{
		ScanID:     "bench-scan",
		Symbol:     "BENCH",
		UpdateType: "update",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		wsServer.BroadcastScanUpdate(update)
	}
}

func BenchmarkResultDeduplication(b *testing.B) {
	cache := NewResultCache(5*time.Minute, 10000)
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		result := models.ScanResult{
			Symbol: fmt.Sprintf("TEST%d", i),
			Spreads: []models.VerticalSpread{
				{
					Symbol:   fmt.Sprintf("TEST%d", i),
					ShortLeg: models.OptionContract{Strike: float64(100 + i)},
					LongLeg:  models.OptionContract{Strike: float64(105 + i)},
					Credit:   1.50,
				},
			},
		}
		cache.Store(result)
	}
	
	testResult := models.ScanResult{
		Symbol: "TEST500",
		Spreads: []models.VerticalSpread{
			{
				Symbol:   "TEST500",
				ShortLeg: models.OptionContract{Strike: 600},
				LongLeg:  models.OptionContract{Strike: 605},
				Credit:   1.50,
			},
		},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = cache.IsDuplicate(testResult)
	}
}