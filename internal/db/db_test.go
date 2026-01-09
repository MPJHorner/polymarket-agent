package db

import (
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestDB(t *testing.T) {
	dbPath := "test_polytracker.db"
	defer os.Remove(dbPath)

	database, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// Test Trader CRUD
	trader := &Trader{
		Address:     "0x123",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1200.50,
		ROI:         0.15,
		Volume:      50000,
		LastScanned: time.Now().Truncate(time.Second),
	}

	if err := database.SaveTrader(trader); err != nil {
		t.Fatalf("failed to save trader: %v", err)
	}

	gotTrader, err := database.GetTrader("0x123")
	if err != nil {
		t.Fatalf("failed to get trader: %v", err)
	}
	if gotTrader == nil {
		t.Fatal("trader not found")
	}
	if gotTrader.Username != trader.Username {
		t.Errorf("expected username %s, got %s", trader.Username, gotTrader.Username)
	}

	// Test Trade CRUD
	trade := &Trade{
		ID:        "trade-1",
		TraderID:  "0x123",
		MarketID:  "market-1",
		Type:      "buy",
		Side:      "yes",
		Price:     0.5,
		Size:      100,
		Timestamp: time.Now().Truncate(time.Second),
	}

	if err := database.SaveTrade(trade); err != nil {
		t.Fatalf("failed to save trade: %v", err)
	}

	trades, err := database.GetTradesByTrader("0x123")
	if err != nil {
		t.Fatalf("failed to get trades: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].ID != trade.ID {
		t.Errorf("expected trade ID %s, got %s", trade.ID, trades[0].ID)
	}
}

func TestWatchlistCRUD(t *testing.T) {
	dbPath := "test_watchlist.db"
	defer os.Remove(dbPath)

	database, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// First create a trader to watchlist
	trader := &Trader{
		Address:     "0xWatchTest123",
		Username:    "watchme",
		WinRate:     0.70,
		ProfitLoss:  2500.00,
		ROI:         0.25,
		Volume:      100000,
		LastScanned: time.Now().Truncate(time.Second),
	}
	if err := database.SaveTrader(trader); err != nil {
		t.Fatalf("failed to save trader: %v", err)
	}

	// Test Add to Watchlist
	if err := database.AddToWatchlist("0xWatchTest123", "Great trader to follow"); err != nil {
		t.Fatalf("failed to add to watchlist: %v", err)
	}

	// Test Get Watchlist Item
	item, err := database.GetWatchlistItem("0xWatchTest123")
	if err != nil {
		t.Fatalf("failed to get watchlist item: %v", err)
	}
	if item == nil {
		t.Fatal("watchlist item not found")
	}
	if item.Notes != "Great trader to follow" {
		t.Errorf("expected notes 'Great trader to follow', got '%s'", item.Notes)
	}

	// Test List Watchlist
	items, err := database.ListWatchlist()
	if err != nil {
		t.Fatalf("failed to list watchlist: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 watchlist item, got %d", len(items))
	}

	// Test Update Notes (upsert)
	if err := database.AddToWatchlist("0xWatchTest123", "Updated notes"); err != nil {
		t.Fatalf("failed to update watchlist: %v", err)
	}
	item, err = database.GetWatchlistItem("0xWatchTest123")
	if err != nil {
		t.Fatalf("failed to get updated watchlist item: %v", err)
	}
	if item.Notes != "Updated notes" {
		t.Errorf("expected notes 'Updated notes', got '%s'", item.Notes)
	}

	// Test Remove from Watchlist
	if err := database.RemoveFromWatchlist("0xWatchTest123"); err != nil {
		t.Fatalf("failed to remove from watchlist: %v", err)
	}

	// Verify removal
	item, err = database.GetWatchlistItem("0xWatchTest123")
	if err != nil {
		t.Fatalf("failed to check removal: %v", err)
	}
	if item != nil {
		t.Error("watchlist item should have been removed")
	}

	// Verify list is empty
	items, err = database.ListWatchlist()
	if err != nil {
		t.Fatalf("failed to list watchlist after removal: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty watchlist, got %d items", len(items))
	}
}
