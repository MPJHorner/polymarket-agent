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
