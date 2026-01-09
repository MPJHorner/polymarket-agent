package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polytracker/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantErr   error
		wantNil   bool
	}{
		{
			name: "valid config",
			config: Config{
				APIKey: "test-api-key",
			},
			wantErr: nil,
			wantNil: false,
		},
		{
			name: "missing API key",
			config: Config{
				APIKey: "",
			},
			wantErr: ErrNoAPIKey,
			wantNil: true,
		},
		{
			name: "with custom endpoint",
			config: Config{
				APIKey:   "test-api-key",
				Endpoint: "https://custom.endpoint.com",
			},
			wantErr: nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}

func TestGenerateThesisPrompt(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef",
		Username:    "TestTrader",
		WinRate:     0.75,
		ProfitLoss:  1500.50,
		ROI:         0.25,
		Volume:      10000.00,
		LastScanned: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	trades := []db.Trade{
		{
			ID:        "trade1",
			TraderID:  "0x1234567890abcdef",
			MarketID:  "market1",
			Type:      "BUY",
			Side:      "YES",
			Price:     0.65,
			Size:      100,
			Timestamp: time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:        "trade2",
			TraderID:  "0x1234567890abcdef",
			MarketID:  "market2",
			Type:      "SELL",
			Side:      "NO",
			Price:     0.35,
			Size:      50,
			Timestamp: time.Date(2024, 1, 13, 8, 0, 0, 0, time.UTC),
		},
	}

	markets := map[string]*db.Market{
		"market1": {
			ID:       "market1",
			Question: "Will Bitcoin reach $100k by end of 2024?",
			Status:   "active",
		},
		"market2": {
			ID:       "market2",
			Question: "Will the Fed cut rates in March 2024?",
			Status:   "active",
		},
	}

	data := TraderData{
		Trader:  trader,
		Trades:  trades,
		Markets: markets,
	}

	prompt := GenerateThesisPrompt(data)

	// Verify prompt contains key elements
	assert.Contains(t, prompt, "Trader Profile")
	assert.Contains(t, prompt, "0x1234567890abcdef")
	assert.Contains(t, prompt, "TestTrader")
	assert.Contains(t, prompt, "75.00%") // Win rate
	assert.Contains(t, prompt, "$1500.50") // P&L
	assert.Contains(t, prompt, "25.00%") // ROI
	assert.Contains(t, prompt, "$10000.00") // Volume

	// Verify trades are included
	assert.Contains(t, prompt, "Trade 1")
	assert.Contains(t, prompt, "Trade 2")
	assert.Contains(t, prompt, "Will Bitcoin reach $100k by end of 2024?")
	assert.Contains(t, prompt, "Will the Fed cut rates in March 2024?")
	assert.Contains(t, prompt, "BUY")
	assert.Contains(t, prompt, "YES")

	// Verify analysis sections
	assert.Contains(t, prompt, "Trading Strategy Summary")
	assert.Contains(t, prompt, "Market Focus")
	assert.Contains(t, prompt, "Risk Profile")
	assert.Contains(t, prompt, "Overall Thesis")
}

func TestGenerateThesisPrompt_NoTrades(t *testing.T) {
	trader := &db.Trader{
		Address:     "0xemptytrader",
		WinRate:     0,
		ProfitLoss:  0,
		ROI:         0,
		Volume:      0,
		LastScanned: time.Now(),
	}

	data := TraderData{
		Trader:  trader,
		Trades:  []db.Trade{},
		Markets: make(map[string]*db.Market),
	}

	prompt := GenerateThesisPrompt(data)

	assert.Contains(t, prompt, "No trades available for analysis")
	assert.Contains(t, prompt, "0xemptytrader")
}

func TestGenerateThesisPrompt_ManyTrades(t *testing.T) {
	trader := &db.Trader{
		Address:     "0xheavytrader",
		WinRate:     0.60,
		ProfitLoss:  5000,
		ROI:         0.15,
		Volume:      50000,
		LastScanned: time.Now(),
	}

	// Generate 100 trades
	trades := make([]db.Trade, 100)
	for i := 0; i < 100; i++ {
		trades[i] = db.Trade{
			ID:        string(rune('A' + i)),
			TraderID:  "0xheavytrader",
			MarketID:  "market1",
			Type:      "BUY",
			Side:      "YES",
			Price:     0.5,
			Size:      10,
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	markets := map[string]*db.Market{
		"market1": {
			ID:       "market1",
			Question: "Test Market",
		},
	}

	data := TraderData{
		Trader:  trader,
		Trades:  trades,
		Markets: markets,
	}

	prompt := GenerateThesisPrompt(data)

	// Should show truncation notice
	assert.Contains(t, prompt, "Showing 50 of 100 total trades")

	// Should have exactly 50 trade sections
	count := strings.Count(prompt, "### Trade")
	assert.Equal(t, 50, count)
}

func TestGenerateThesisPrompt_UnknownMarket(t *testing.T) {
	trader := &db.Trader{
		Address:     "0xtest",
		LastScanned: time.Now(),
	}

	trades := []db.Trade{
		{
			ID:        "trade1",
			TraderID:  "0xtest",
			MarketID:  "unknown_market",
			Type:      "BUY",
			Side:      "YES",
			Price:     0.5,
			Size:      10,
			Timestamp: time.Now(),
		},
	}

	data := TraderData{
		Trader:  trader,
		Trades:  trades,
		Markets: make(map[string]*db.Market), // Empty markets map
	}

	prompt := GenerateThesisPrompt(data)

	assert.Contains(t, prompt, "Unknown Market")
}

func TestAnalyzeTrader_InvalidTrader(t *testing.T) {
	client, err := NewClient(Config{APIKey: "test-key"})
	require.NoError(t, err)

	_, err = client.AnalyzeTrader(context.Background(), TraderData{
		Trader: nil,
	})

	assert.ErrorIs(t, err, ErrInvalidTrader)
}

func TestAnalyzeTrader_MockAPI(t *testing.T) {
	// Create a mock server
	mockResponse := map[string]interface{}{
		"id":   "msg_123",
		"type": "message",
		"role": "assistant",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "## Trading Strategy Summary\n\nThis trader shows a pattern of buying YES positions on political markets...",
			},
		},
		"model":         "claude-3-5-sonnet-20241022",
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  500,
			"output_tokens": 200,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("x-api-key"), "test-api-key")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIKey:   "test-api-key",
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	trader := &db.Trader{
		Address:     "0x123",
		Username:    "TestUser",
		WinRate:     0.65,
		ProfitLoss:  1000,
		ROI:         0.20,
		Volume:      5000,
		LastScanned: time.Now(),
	}

	data := TraderData{
		Trader:  trader,
		Trades:  []db.Trade{},
		Markets: make(map[string]*db.Market),
	}

	result, err := client.AnalyzeTrader(context.Background(), data)
	require.NoError(t, err)

	assert.Contains(t, result.Thesis, "Trading Strategy Summary")
	assert.Equal(t, "claude-3-5-sonnet-20241022", result.Model)
	assert.Equal(t, int64(500), result.InputTokens)
	assert.Equal(t, int64(200), result.OutputTokens)
	assert.Equal(t, "end_turn", result.StopReason)
}

func TestAnalyzeTrader_TokenLimit(t *testing.T) {
	// Create a mock server that returns max_tokens stop reason
	mockResponse := map[string]interface{}{
		"id":   "msg_123",
		"type": "message",
		"role": "assistant",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Partial response that was truncated...",
			},
		},
		"model":         "claude-3-5-sonnet-20241022",
		"stop_reason":   "max_tokens",
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  500,
			"output_tokens": 4096,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIKey:   "test-api-key",
		Endpoint: server.URL,
	})
	require.NoError(t, err)

	trader := &db.Trader{
		Address:     "0x123",
		LastScanned: time.Now(),
	}

	data := TraderData{
		Trader:  trader,
		Trades:  []db.Trade{},
		Markets: make(map[string]*db.Market),
	}

	result, err := client.AnalyzeTrader(context.Background(), data)

	// Should return both result and error
	assert.ErrorIs(t, err, ErrTokenLimit)
	assert.NotNil(t, result)
	assert.Contains(t, result.Thesis, "Partial response")
}
