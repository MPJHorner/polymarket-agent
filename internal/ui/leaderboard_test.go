package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"polytracker/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

func setupTestDB(t *testing.T) *db.DB {
	tmpFile, err := os.CreateTemp("", "leaderboard_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	database, err := db.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
	})

	return database
}

func TestLeaderboardCreation(t *testing.T) {
	styles := DefaultStyles()
	lb := NewLeaderboard(styles)

	if lb == nil {
		t.Fatal("NewLeaderboard returned nil")
	}

	if lb.sortField != db.SortByProfitLoss {
		t.Errorf("Expected default sort field to be profit_loss, got %s", lb.sortField)
	}

	if lb.sortOrder != db.SortDesc {
		t.Errorf("Expected default sort order to be DESC, got %s", lb.sortOrder)
	}

	if lb.currentPage != 0 {
		t.Errorf("Expected current page to be 0, got %d", lb.currentPage)
	}
}

func TestLeaderboardLoadTraders(t *testing.T) {
	database := setupTestDB(t)

	// Add test traders
	traders := []db.Trader{
		{Address: "0x1111111111111111", Username: "trader1", WinRate: 0.75, ProfitLoss: 1500.0, ROI: 0.5, Volume: 10000.0, LastScanned: time.Now()},
		{Address: "0x2222222222222222", Username: "trader2", WinRate: 0.6, ProfitLoss: 2500.0, ROI: 0.4, Volume: 20000.0, LastScanned: time.Now()},
		{Address: "0x3333333333333333", Username: "trader3", WinRate: 0.8, ProfitLoss: 500.0, ROI: 0.3, Volume: 5000.0, LastScanned: time.Now()},
	}

	for _, trader := range traders {
		if err := database.SaveTrader(&trader); err != nil {
			t.Fatalf("Failed to save trader: %v", err)
		}
	}

	styles := DefaultStyles()
	lb := NewLeaderboard(styles)

	cmd := lb.LoadTraders(database)
	if cmd == nil {
		t.Fatal("LoadTraders should return a command")
	}

	msg := cmd()
	loadedMsg, ok := msg.(tradersLoadedMsg)
	if !ok {
		t.Fatal("Expected tradersLoadedMsg")
	}

	if loadedMsg.totalCount != 3 {
		t.Errorf("Expected 3 traders, got %d", loadedMsg.totalCount)
	}

	// Should be sorted by profit_loss DESC, so trader2 should be first
	if len(loadedMsg.traders) < 1 || loadedMsg.traders[0].Address != "0x2222222222222222" {
		t.Error("Expected traders to be sorted by profit_loss DESC")
	}
}

func TestLeaderboardSortToggle(t *testing.T) {
	styles := DefaultStyles()
	lb := NewLeaderboard(styles)

	// Default should be ProfitLoss DESC
	if lb.sortField != db.SortByProfitLoss || lb.sortOrder != db.SortDesc {
		t.Error("Expected default to be ProfitLoss DESC")
	}

	// Press 'w' to switch to WinRate
	lb, _ = lb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	if lb.sortField != db.SortByWinRate {
		t.Errorf("Expected WinRate, got %s", lb.sortField)
	}
	if lb.sortOrder != db.SortDesc {
		t.Error("Expected DESC when first switching to WinRate")
	}

	// Press 'w' again to toggle order
	lb, _ = lb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	if lb.sortOrder != db.SortAsc {
		t.Error("Expected ASC after toggling")
	}

	// Press 'p' to switch to ProfitLoss
	lb, _ = lb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if lb.sortField != db.SortByProfitLoss {
		t.Errorf("Expected ProfitLoss, got %s", lb.sortField)
	}
	if lb.sortOrder != db.SortDesc {
		t.Error("Expected DESC when switching to ProfitLoss")
	}
}

func TestLeaderboardView(t *testing.T) {
	styles := DefaultStyles()
	lb := NewLeaderboard(styles)

	view := lb.View()

	expectedParts := []string{
		"Sorted by:",
		"P&L",
		"Page",
	}

	for _, part := range expectedParts {
		if !strings.Contains(view, part) {
			t.Errorf("View missing expected part: %s", part)
		}
	}
}

func TestLeaderboardHelpText(t *testing.T) {
	styles := DefaultStyles()
	lb := NewLeaderboard(styles)

	help := lb.HelpText()

	if !strings.Contains(help, "enter") {
		t.Error("Help text should mention enter key")
	}
	if !strings.Contains(help, "w:") {
		t.Error("Help text should mention 'w' for win sort")
	}
	if !strings.Contains(help, "p:") {
		t.Error("Help text should mention 'p' for P&L sort")
	}
}

func TestFormatPNL(t *testing.T) {
	cases := []struct {
		input    float64
		expected string
	}{
		{100.0, "+$100.00"},
		{-50.0, "-$50.00"},
		{0.0, "+$0.00"},
	}

	for _, tc := range cases {
		result := formatPNL(tc.input)
		if result != tc.expected {
			t.Errorf("formatPNL(%f) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestFormatVolume(t *testing.T) {
	cases := []struct {
		input    float64
		expected string
	}{
		{500.0, "$500"},
		{1500.0, "$1.5K"},
		{1000000.0, "$1.0M"},
		{2500000.0, "$2.5M"},
	}

	for _, tc := range cases {
		result := formatVolume(tc.input)
		if result != tc.expected {
			t.Errorf("formatVolume(%f) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}
