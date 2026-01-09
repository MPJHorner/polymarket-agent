package ui

import (
	"testing"
	"time"

	"polytracker/internal/claude"
	"polytracker/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewAnalysis(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)

	assert.NotNil(t, analysis)
	assert.Equal(t, trader, analysis.trader)
	assert.Equal(t, analysisStateFetching, analysis.state)
	assert.Empty(t, analysis.thesis)
}

func TestAnalysisView_Fetching(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)

	view := analysis.View()
	assert.Contains(t, view, "ANALYSIS")
	assert.Contains(t, view, "Fetching trader data")
}

func TestAnalysisView_Analyzing(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)
	analysis.state = analysisStateAnalyzing

	view := analysis.View()
	assert.Contains(t, view, "ANALYSIS")
	assert.Contains(t, view, "Analyzing with Claude")
}

func TestAnalysisUpdate_DataFetched(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)

	trades := []db.Trade{
		{
			ID:        "trade1",
			TraderID:  trader.Address,
			MarketID:  "market1",
			Type:      "buy",
			Side:      "yes",
			Price:     0.65,
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

	msg := AnalysisDataFetchedMsg{
		Trades:  trades,
		Markets: markets,
	}

	analysis, _ = analysis.Update(msg)

	assert.Equal(t, trades, analysis.trades)
	assert.Equal(t, markets, analysis.markets)
	assert.Equal(t, analysisStateAnalyzing, analysis.state)
}

func TestAnalysisUpdate_Complete(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)
	analysis.state = analysisStateAnalyzing

	thesis := "## Trading Strategy\n\nThis trader shows consistent patterns..."

	msg := AnalysisCompleteMsg{Thesis: thesis}
	analysis, _ = analysis.Update(msg)

	assert.Equal(t, thesis, analysis.thesis)
	assert.Equal(t, analysisStateComplete, analysis.state)
}

func TestAnalysisUpdate_Error(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)

	testErr := claude.ErrNoAPIKey
	msg := AnalysisErrorMsg{Err: testErr}
	analysis, _ = analysis.Update(msg)

	assert.Equal(t, testErr, analysis.err)
	assert.Equal(t, analysisStateError, analysis.state)
}

func TestAnalysisUpdate_KeyBack(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	analysis, cmd := analysis.Update(msg)

	// Command should return GoBackMsg
	assert.NotNil(t, cmd)
	result := cmd()
	_, ok := result.(GoBackMsg)
	assert.True(t, ok)
}

func TestAnalysisUpdate_KeyRetry(t *testing.T) {
	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "testuser",
		WinRate:     0.65,
		ProfitLoss:  1500.50,
		ROI:         0.35,
		Volume:      10000.00,
		LastScanned: time.Now(),
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)
	analysis.state = analysisStateError
	analysis.err = claude.ErrNoAPIKey

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	analysis, cmd := analysis.Update(msg)

	assert.Equal(t, analysisStateFetching, analysis.state)
	assert.Nil(t, analysis.err)
	assert.NotNil(t, cmd)

	// Command should return AnalysisRetryMsg
	result := cmd()
	_, ok := result.(AnalysisRetryMsg)
	assert.True(t, ok)
}

func TestAnalysisHelpText(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234567890abcdef1234567890abcdef12345678",
		Username: "testuser",
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)

	// Fetching state
	analysis.state = analysisStateFetching
	assert.Contains(t, analysis.HelpText(), "cancel")

	// Error state
	analysis.state = analysisStateError
	assert.Contains(t, analysis.HelpText(), "retry")
	assert.Contains(t, analysis.HelpText(), "back")

	// Complete state
	analysis.state = analysisStateComplete
	assert.Contains(t, analysis.HelpText(), "save")
	assert.Contains(t, analysis.HelpText(), "retry")
	assert.Contains(t, analysis.HelpText(), "scroll")
}

func TestAnalysisRenderMarkdown(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234567890abcdef1234567890abcdef12345678",
		Username: "testuser",
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.SetSize(80, 24)

	content := `# Main Title
## Section Header
### Subsection
- Bullet point
1. Numbered item
**Bold text** normal text`

	rendered := analysis.renderMarkdown(content)

	// Rendered content should contain the text (minus markdown syntax)
	assert.Contains(t, rendered, "Main Title")
	assert.Contains(t, rendered, "Section Header")
	assert.Contains(t, rendered, "Subsection")
	assert.Contains(t, rendered, "Bullet point")
	assert.Contains(t, rendered, "Numbered item")
	assert.Contains(t, rendered, "normal text")
}

func TestAnalysisGetTrader(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234567890abcdef1234567890abcdef12345678",
		Username: "testuser",
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)

	assert.Equal(t, trader, analysis.GetTrader())
}

func TestAnalysisGetThesis(t *testing.T) {
	trader := &db.Trader{
		Address:  "0x1234567890abcdef1234567890abcdef12345678",
		Username: "testuser",
	}

	styles := DefaultStyles()
	analysis := NewAnalysis(trader, styles, nil)
	analysis.thesis = "Test thesis content"

	assert.Equal(t, "Test thesis content", analysis.GetThesis())
}
