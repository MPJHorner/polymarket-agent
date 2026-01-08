package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMarket(t *testing.T) {
	mockMarket := Market{
		ID:       "123",
		Question: "Will it rain?",
		Slug:     "will-it-rain",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/123" {
			t.Errorf("Expected path /markets/123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockMarket)
	}))
	defer server.Close()

	client := NewClient(Config{
		GammaBaseURL: server.URL,
	})

	market, err := client.GetMarket(context.Background(), "123")
	if err != nil {
		t.Fatalf("Failed to get market: %v", err)
	}

	if market.ID != mockMarket.ID {
		t.Errorf("Expected market ID %s, got %s", mockMarket.ID, market.ID)
	}
	if market.Question != mockMarket.Question {
		t.Errorf("Expected market question %s, got %s", mockMarket.Question, market.Question)
	}
}

func TestGetTrades(t *testing.T) {
	mockTrades := []Trade{
		{ID: "t1", MarketID: "m1", Price: 0.5},
		{ID: "t2", MarketID: "m1", Price: 0.6},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Errorf("Expected path /trades, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("market_id") != "m1" {
			t.Errorf("Expected market_id m1, got %s", r.URL.Query().Get("market_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockTrades)
	}))
	defer server.Close()

	client := NewClient(Config{
		CLOBBaseURL: server.URL,
	})

	trades, err := client.GetTrades(context.Background(), "m1")
	if err != nil {
		t.Fatalf("Failed to get trades: %v", err)
	}

	if len(trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(trades))
	}
	if trades[0].ID != "t1" {
		t.Errorf("Expected trade ID t1, got %s", trades[0].ID)
	}
}

func TestRateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 100 requests per second, so it shouldn't be too slow but enough to test
	client := NewClient(Config{
		GammaBaseURL: server.URL,
		RateLimit:    100, 
		Burst:        1,
	})

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := client.gammaResty.R().SetContext(ctx).Get("/")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}
}
