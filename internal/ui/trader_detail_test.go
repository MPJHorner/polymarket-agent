package ui

import (
	"testing"
	"time"

	"polytracker/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewTraderDetail(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.25,
		Volume:      50000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)

	assert.NotNil(t, td)
	assert.Equal(t, trader, td.trader)
	assert.Empty(t, td.trades)
	assert.False(t, td.showAllTrades)
	assert.False(t, td.isOnWatchlist)
}

func TestTraderDetailUpdate(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.25,
		Volume:      50000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)
	td.SetSize(100, 50)

	// Test loading trades
	trades := []db.Trade{
		{
			ID:        "trade1",
			TraderID:  trader.Address,
			MarketID:  "market1",
			Type:      "buy",
			Side:      "yes",
			Price:     0.55,
			Size:      100,
			Timestamp: time.Now(),
		},
	}
	markets := map[string]*db.Market{
		"market1": {
			ID:       "market1",
			Question: "Will it rain tomorrow?",
		},
	}

	td, _ = td.Update(tradesLoadedMsg{trades: trades, markets: markets})
	assert.Equal(t, 1, len(td.trades))

	// Test watchlist status update
	td, _ = td.Update(watchlistStatusMsg{isOnWatchlist: true})
	assert.True(t, td.isOnWatchlist)

	// Test toggle all trades
	td, _ = td.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	assert.True(t, td.showAllTrades)

	// Test scroll down
	td, _ = td.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assert.Equal(t, 1, td.scrollOffset)

	// Test scroll up
	td, _ = td.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	assert.Equal(t, 0, td.scrollOffset)
}

func TestTraderDetailBackNavigation(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234",
		Username: "test",
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)
	td.SetSize(100, 50)

	// Test back navigation with ESC
	td, cmd := td.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg := cmd()
	_, isGoBack := msg.(GoBackMsg)
	assert.True(t, isGoBack)
}

func TestTraderDetailAnalyze(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234",
		Username: "test",
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)
	td.SetSize(100, 50)

	// Test analyze action
	td, cmd := td.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	msg := cmd()
	analyzeMsg, isAnalyze := msg.(AnalyzeTraderMsg)
	assert.True(t, isAnalyze)
	assert.Equal(t, trader.Address, analyzeMsg.Trader.Address)
}

func TestTraderDetailWatchlist(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234",
		Username: "test",
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)
	td.SetSize(100, 50)

	// Test watch action (add)
	td.isOnWatchlist = false
	td, cmd := td.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	msg := cmd()
	watchMsg, isWatch := msg.(ToggleWatchlistMsg)
	assert.True(t, isWatch)
	assert.True(t, watchMsg.Add)

	// Test watch action (remove)
	td.isOnWatchlist = true
	td, cmd = td.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	msg = cmd()
	watchMsg, isWatch = msg.(ToggleWatchlistMsg)
	assert.True(t, isWatch)
	assert.False(t, watchMsg.Add)
}

func TestTraderDetailView(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.25,
		Volume:      50000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)
	td.SetSize(100, 50)

	view := td.View()

	// Check that key elements are present
	assert.Contains(t, view, "TRADER PROFILE")
	assert.Contains(t, view, "STATISTICS")
	assert.Contains(t, view, "RECENT TRADES")
	assert.Contains(t, view, "testuser")
}

func TestTraderDetailHelpText(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234",
		Username: "test",
	}

	styles := DefaultStyles()
	td := NewTraderDetail(trader, styles)

	// Not on watchlist
	td.isOnWatchlist = false
	help := td.HelpText()
	assert.Contains(t, help, "add to watchlist")

	// On watchlist
	td.isOnWatchlist = true
	help = td.HelpText()
	assert.Contains(t, help, "remove from watchlist")
}

func TestTraderDetailNilTrader(t *testing.T) {
	styles := DefaultStyles()
	td := NewTraderDetail(nil, styles)
	td.SetSize(100, 50)

	view := td.View()
	assert.Equal(t, "No trader selected", view)
}
