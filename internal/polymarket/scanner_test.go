package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"polytracker/internal/db"
	"testing"
)

func TestScanner_ScanRecentActivity(t *testing.T) {
	// Setup mock DB
	dbPath := "test_scanner.db"
	defer os.Remove(dbPath)
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer database.Close()

	// Mock Data
	mockMarkets := []Market{
		{ID: "m1", Question: "Market 1"},
	}
	mockTrades := []Trade{
		{ID: "t1", MarketID: "m1", Price: 0.5, Size: 100, Maker: "addr1", Taker: "addr2"},
	}

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/markets" {
			json.NewEncoder(w).Encode(mockMarkets)
		} else if r.URL.Path == "/trades" {
			json.NewEncoder(w).Encode(mockTrades)
		}
	}))
	defer server.Close()

	client := NewClient(Config{
		GammaBaseURL: server.URL,
		CLOBBaseURL:  server.URL,
	})

	scanner := NewScanner(client, database)
	err = scanner.ScanRecentActivity(context.Background(), 1)
	if err != nil {
		t.Fatalf("ScanRecentActivity failed: %v", err)
	}

	// Verify DB state
	traders, err := database.ListTraders()
	if err != nil {
		t.Fatalf("Failed to list traders: %v", err)
	}

	if len(traders) != 2 {
		t.Errorf("Expected 2 traders (maker and taker), got %d", len(traders))
	}

	foundAddr1 := false
	for _, tr := range traders {
		if tr.Address == "addr1" {
			foundAddr1 = true
			if tr.Volume != 50.0 { // 0.5 * 100
				t.Errorf("Expected volume 50.0 for addr1, got %f", tr.Volume)
			}
		}
	}
	if !foundAddr1 {
		t.Errorf("addr1 not found in DB")
	}
}
