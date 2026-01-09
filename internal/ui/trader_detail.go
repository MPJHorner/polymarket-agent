package ui

import (
	"fmt"
	"strings"

	"polytracker/internal/db"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	recentTradesLimit = 10
)

type TraderDetailKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Back    key.Binding
	Analyze key.Binding
	Watch   key.Binding
	Trades  key.Binding
}

var traderDetailKeys = TraderDetailKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "scroll down"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Analyze: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "analyze"),
	),
	Watch: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "watch"),
	),
	Trades: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "all trades"),
	),
}

type TraderDetail struct {
	trader       *db.Trader
	trades       []db.Trade
	markets      map[string]*db.Market
	styles       Styles
	width        int
	height       int
	scrollOffset int
	showAllTrades bool
	isOnWatchlist bool
}

type tradesLoadedMsg struct {
	trades  []db.Trade
	markets map[string]*db.Market
}

type watchlistStatusMsg struct {
	isOnWatchlist bool
}

type GoBackMsg struct{}

type AnalyzeTraderMsg struct {
	Trader *db.Trader
}

type ToggleWatchlistMsg struct {
	Trader *db.Trader
	Add    bool
}

func NewTraderDetail(trader *db.Trader, styles Styles) *TraderDetail {
	return &TraderDetail{
		trader:        trader,
		trades:        nil,
		markets:       make(map[string]*db.Market),
		styles:        styles,
		scrollOffset:  0,
		showAllTrades: false,
		isOnWatchlist: false,
	}
}

func (td *TraderDetail) SetSize(width, height int) {
	td.width = width
	td.height = height
}

func (td *TraderDetail) LoadTrades(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		if td.trader == nil {
			return nil
		}

		trades, err := database.GetTradesByTrader(td.trader.Address)
		if err != nil {
			return nil
		}

		markets := make(map[string]*db.Market)
		for _, t := range trades {
			if _, exists := markets[t.MarketID]; !exists {
				market, err := database.GetMarket(t.MarketID)
				if err == nil && market != nil {
					markets[t.MarketID] = market
				}
			}
		}

		return tradesLoadedMsg{
			trades:  trades,
			markets: markets,
		}
	}
}

func (td *TraderDetail) CheckWatchlistStatus(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		if td.trader == nil {
			return nil
		}

		item, err := database.GetWatchlistItem(td.trader.Address)
		if err != nil {
			return watchlistStatusMsg{isOnWatchlist: false}
		}

		return watchlistStatusMsg{isOnWatchlist: item != nil}
	}
}

func (td *TraderDetail) Update(msg tea.Msg) (*TraderDetail, tea.Cmd) {
	switch msg := msg.(type) {
	case tradesLoadedMsg:
		td.trades = msg.trades
		td.markets = msg.markets
		return td, nil

	case watchlistStatusMsg:
		td.isOnWatchlist = msg.isOnWatchlist
		return td, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, traderDetailKeys.Back):
			return td, func() tea.Msg { return GoBackMsg{} }

		case key.Matches(msg, traderDetailKeys.Up):
			if td.scrollOffset > 0 {
				td.scrollOffset--
			}
			return td, nil

		case key.Matches(msg, traderDetailKeys.Down):
			td.scrollOffset++
			return td, nil

		case key.Matches(msg, traderDetailKeys.Analyze):
			if td.trader != nil {
				return td, func() tea.Msg { return AnalyzeTraderMsg{Trader: td.trader} }
			}
			return td, nil

		case key.Matches(msg, traderDetailKeys.Watch):
			if td.trader != nil {
				return td, func() tea.Msg {
					return ToggleWatchlistMsg{Trader: td.trader, Add: !td.isOnWatchlist}
				}
			}
			return td, nil

		case key.Matches(msg, traderDetailKeys.Trades):
			td.showAllTrades = !td.showAllTrades
			td.scrollOffset = 0
			return td, nil
		}
	}

	return td, nil
}

func (td *TraderDetail) View() string {
	if td.trader == nil {
		return "No trader selected"
	}

	var sections []string

	// Profile section
	sections = append(sections, td.renderProfile())

	// Stats section
	sections = append(sections, td.renderStats())

	// Recent trades section
	sections = append(sections, td.renderTrades())

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply scrolling if content is taller than viewport
	lines := strings.Split(content, "\n")
	if td.scrollOffset >= len(lines) {
		td.scrollOffset = len(lines) - 1
	}
	if td.scrollOffset < 0 {
		td.scrollOffset = 0
	}

	visibleHeight := td.height - 4 // Account for padding
	if visibleHeight < 1 {
		visibleHeight = 20
	}

	endIdx := td.scrollOffset + visibleHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	visibleLines := lines[td.scrollOffset:endIdx]
	return strings.Join(visibleLines, "\n")
}

func (td *TraderDetail) renderProfile() string {
	t := td.trader

	address := t.Address
	shortAddress := address
	if len(address) > 20 {
		shortAddress = address[:10] + "..." + address[len(address)-8:]
	}

	username := t.Username
	if username == "" {
		username = "(no username)"
	}

	watchStatus := ""
	if td.isOnWatchlist {
		watchStatus = " [WATCHING]"
	}

	header := td.styles.Header.Render(" TRADER PROFILE " + watchStatus)

	profileBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(td.styles.Header.GetBackground()).
		Padding(1, 2).
		Width(td.width - 6)

	profileContent := lipgloss.JoinVertical(
		lipgloss.Left,
		td.styles.Highlight.Render("Address:  ")+shortAddress,
		td.styles.Highlight.Render("Full:     ")+td.styles.Subtle.Render(address),
		td.styles.Highlight.Render("Username: ")+username,
		td.styles.Highlight.Render("Scanned:  ")+t.LastScanned.Format("2006-01-02 15:04:05"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		profileBox.Render(profileContent),
	)
}

func (td *TraderDetail) renderStats() string {
	t := td.trader

	header := td.styles.Header.Render(" STATISTICS ")

	statsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(td.styles.Header.GetBackground()).
		Padding(1, 2).
		Width(td.width - 6)

	// Format stats with colors
	winRateStyle := td.styles.Highlight
	if t.WinRate >= 0.6 {
		winRateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")) // Green
	} else if t.WinRate < 0.4 {
		winRateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")) // Red
	}

	pnlStyle := td.styles.Highlight
	if t.ProfitLoss >= 0 {
		pnlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")) // Green
	} else {
		pnlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")) // Red
	}

	roiStyle := td.styles.Highlight
	if t.ROI >= 0 {
		roiStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")) // Green
	} else {
		roiStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")) // Red
	}

	statsContent := lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("%-12s %s", "Win Rate:", winRateStyle.Render(fmt.Sprintf("%.1f%%", t.WinRate*100))),
		fmt.Sprintf("%-12s %s", "P&L:", pnlStyle.Render(formatPNL(t.ProfitLoss))),
		fmt.Sprintf("%-12s %s", "ROI:", roiStyle.Render(fmt.Sprintf("%.1f%%", t.ROI*100))),
		fmt.Sprintf("%-12s %s", "Volume:", td.styles.Highlight.Render(formatVolume(t.Volume))),
		fmt.Sprintf("%-12s %s", "Trades:", td.styles.Highlight.Render(fmt.Sprintf("%d", len(td.trades)))),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		header,
		"",
		statsBox.Render(statsContent),
	)
}

func (td *TraderDetail) renderTrades() string {
	title := " RECENT TRADES "
	if td.showAllTrades {
		title = " ALL TRADES "
	}
	header := td.styles.Header.Render(title)

	if len(td.trades) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			header,
			"",
			td.styles.Subtle.Render("  No trades found"),
		)
	}

	tradesBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(td.styles.Header.GetBackground()).
		Padding(1, 2).
		Width(td.width - 6)

	// Build trades list
	var tradeLines []string

	// Header row
	headerLine := fmt.Sprintf("%-10s %-6s %-6s %-10s %-10s %-30s",
		"Date", "Type", "Side", "Price", "Size", "Market")
	tradeLines = append(tradeLines, td.styles.Subtle.Render(headerLine))
	tradeLines = append(tradeLines, td.styles.Subtle.Render(strings.Repeat("-", 75)))

	limit := recentTradesLimit
	if td.showAllTrades {
		limit = len(td.trades)
	}
	if limit > len(td.trades) {
		limit = len(td.trades)
	}

	for i := 0; i < limit; i++ {
		trade := td.trades[i]

		typeStyle := td.styles.Highlight
		if trade.Type == "buy" {
			typeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b"))
		} else {
			typeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
		}

		marketQuestion := trade.MarketID[:8] + "..."
		if market, exists := td.markets[trade.MarketID]; exists && market != nil {
			marketQuestion = market.Question
			if len(marketQuestion) > 28 {
				marketQuestion = marketQuestion[:28] + ".."
			}
		}

		tradeLine := fmt.Sprintf("%-10s %s %-6s %-10s %-10s %-30s",
			trade.Timestamp.Format("01/02 15:04"),
			typeStyle.Render(fmt.Sprintf("%-6s", trade.Type)),
			trade.Side,
			fmt.Sprintf("$%.3f", trade.Price),
			fmt.Sprintf("%.2f", trade.Size),
			marketQuestion,
		)
		tradeLines = append(tradeLines, tradeLine)
	}

	if !td.showAllTrades && len(td.trades) > recentTradesLimit {
		tradeLines = append(tradeLines, "")
		tradeLines = append(tradeLines, td.styles.Subtle.Render(
			fmt.Sprintf("  ... and %d more trades (press 't' to show all)", len(td.trades)-recentTradesLimit)))
	}

	tradesContent := strings.Join(tradeLines, "\n")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		header,
		"",
		tradesBox.Render(tradesContent),
	)
}

func (td *TraderDetail) HelpText() string {
	watchAction := "w: add to watchlist"
	if td.isOnWatchlist {
		watchAction = "w: remove from watchlist"
	}
	return fmt.Sprintf("esc: back | a: analyze | %s | t: toggle all trades | j/k: scroll", watchAction)
}

func (td *TraderDetail) GetTrader() *db.Trader {
	return td.trader
}
