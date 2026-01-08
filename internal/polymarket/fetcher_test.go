package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"polytracker/internal/db"
	"testing"
	"time"
)

func TestFetcher_FetchTraderHistory(t *testing.T) {
	// Setup DB
	dbPath := "test_fetcher.db"
	defer os.Remove(dbPath)
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer database.Close()

	// Mock Gamma API
	mockMarket := Market{
		ID:       "m1",
		Question: "Will it rain?",
		Tokens: []Token{
			{TokenID: "t1", Outcome: "Yes", Price: 0.6},
			{TokenID: "t2", Outcome: "No", Price: 0.4},
		},
		Closed: false,
	}

	gammaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockMarket)
	}))
	defer gammaServer.Close()

	// Mock CLOB API
	mockTrades := []Trade{
		{
			ID:        "trade-1",
			MarketID:  "m1",
			Price:     0.55,
			Size:      100,
			Side:      "BUY",
			Timestamp: time.Now().Unix(),
		},
	}

	clobServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockTrades)
	}))
	defer clobServer.Close()

	// Setup Client
	client := NewClient(Config{
		GammaBaseURL: gammaServer.URL,
		CLOBBaseURL:  clobServer.URL,
	})

	fetcher := NewFetcher(client, database)

	// Test FetchTraderHistory
	address := "0xabc"
	err = fetcher.FetchTraderHistory(context.Background(), address)
	if err != nil {
		t.Fatalf("FetchTraderHistory failed: %v", err)
	}

	// Verify Data
	trades, err := database.GetTradesByTrader(address)
	if err != nil {
		t.Fatalf("Failed to get trades from DB: %v", err)
	}
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	} else {
		if trades[0].MarketID != "m1" {
			t.Errorf("Expected market ID m1, got %s", trades[0].MarketID)
		}
		if trades[0].Price != 0.55 {
			t.Errorf("Expected price 0.55, got %f", trades[0].Price)
		}
	}

	market, err := database.GetMarket("m1")
	if err != nil {
		t.Fatalf("Failed to get market from DB: %v", err)
	}
	if market == nil {
		t.Fatal("Market not found in DB")
	}
	if market.Question != "Will it rain?" {
		t.Errorf("Expected question 'Will it rain?', got '%s'", market.Question)
	}

	snapshot, err := database.GetLatestMarketSnapshot("m1")
	if err != nil {
		t.Fatalf("Failed to get snapshot from DB: %v", err)
	}
	if snapshot == nil {
		t.Fatal("Snapshot not found in DB")
	}
	if snapshot.YesPrice != 0.6 {
		t.Errorf("Expected yes price 0.6, got %f", snapshot.YesPrice)
	}
}
