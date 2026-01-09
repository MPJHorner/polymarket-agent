package ui

import (
	"fmt"
	"strings"

	"polytracker/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateScan sessionState = iota
	stateLeaderboard
	stateWatchlist
	stateSettings
	stateTraderDetail
)

type Model struct {
	state          sessionState
	previousState  sessionState
	styles         Styles
	width          int
	height         int
	err            error
	leaderboard    *Leaderboard
	traderDetail   *TraderDetail
	db             *db.DB
	selectedTrader *db.Trader
}

func NewModel(themeName string) Model {
	t, ok := Themes[strings.ToLower(themeName)]
	if !ok {
		t = Dracula
	}
	styles := GetStyles(t)
	return Model{
		state:       stateLeaderboard,
		styles:      styles,
		leaderboard: NewLeaderboard(styles),
	}
}

func NewModelWithDB(themeName string, database *db.DB) Model {
	m := NewModel(themeName)
	m.db = database
	return m
}

func (m Model) Init() tea.Cmd {
	if m.db != nil && m.leaderboard != nil {
		return m.leaderboard.LoadTraders(m.db)
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys first
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.state = stateScan
			return m, nil
		case "2":
			m.state = stateLeaderboard
			if m.db != nil && m.leaderboard != nil {
				cmds = append(cmds, m.leaderboard.LoadTraders(m.db))
			}
			return m, tea.Batch(cmds...)
		case "3":
			m.state = stateWatchlist
			return m, nil
		case "4":
			m.state = stateSettings
			return m, nil
		case "?":
			// Toggle help? For now just stay.
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := m.height - 6 // Account for header, tabs, footer
		if m.leaderboard != nil {
			m.leaderboard.SetSize(m.width, contentHeight)
		}
		if m.traderDetail != nil {
			m.traderDetail.SetSize(m.width, contentHeight)
		}

	case TraderSelectedMsg:
		if msg.Trader != nil {
			m.selectedTrader = msg.Trader
			m.previousState = m.state
			m.state = stateTraderDetail
			m.traderDetail = NewTraderDetail(msg.Trader, m.styles)
			m.traderDetail.SetSize(m.width, m.height-6)
			var cmds []tea.Cmd
			if m.db != nil {
				cmds = append(cmds, m.traderDetail.LoadTrades(m.db))
				cmds = append(cmds, m.traderDetail.CheckWatchlistStatus(m.db))
			}
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case GoBackMsg:
		m.state = stateLeaderboard
		m.selectedTrader = nil
		m.traderDetail = nil
		return m, nil

	case AnalyzeTraderMsg:
		// Will be implemented in ai-001/ai-002
		return m, nil

	case ToggleWatchlistMsg:
		if m.db != nil && msg.Trader != nil {
			if msg.Add {
				_ = m.db.AddToWatchlist(msg.Trader.Address, "")
			} else {
				_ = m.db.RemoveFromWatchlist(msg.Trader.Address)
			}
			if m.traderDetail != nil {
				return m, m.traderDetail.CheckWatchlistStatus(m.db)
			}
		}
		return m, nil

	case tradesLoadedMsg:
		if m.traderDetail != nil {
			m.traderDetail, cmd = m.traderDetail.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case watchlistStatusMsg:
		if m.traderDetail != nil {
			m.traderDetail, cmd = m.traderDetail.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tradersLoadedMsg:
		if m.leaderboard != nil {
			m.leaderboard, cmd = m.leaderboard.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Pass messages to leaderboard when in leaderboard state
	if m.state == stateLeaderboard && m.leaderboard != nil {
		oldSort := m.leaderboard.GetSortField()
		oldOrder := m.leaderboard.GetSortOrder()
		oldPage := m.leaderboard.GetCurrentPage()

		m.leaderboard, cmd = m.leaderboard.Update(msg)
		cmds = append(cmds, cmd)

		// Reload data if sort or page changed
		newSort := m.leaderboard.GetSortField()
		newOrder := m.leaderboard.GetSortOrder()
		newPage := m.leaderboard.GetCurrentPage()

		if m.db != nil && (oldSort != newSort || oldOrder != newOrder || oldPage != newPage) {
			cmds = append(cmds, m.leaderboard.LoadTraders(m.db))
		}
	}

	// Pass messages to trader detail when in trader detail state
	if m.state == stateTraderDetail && m.traderDetail != nil {
		m.traderDetail, cmd = m.traderDetail.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	// Header
	header := m.styles.Header.Render(" POLYTRACKER ")
	s.WriteString(header + "\n")

	// Tabs
	tabs := m.renderTabs()
	s.WriteString(tabs + "\n\n")

	// Content
	content := m.renderContent()
	s.WriteString(content)

	// Footer
	footer := m.renderFooter()
	// Push footer to bottom
	contentHeight := lipgloss.Height(s.String())
	if m.height > contentHeight+1 {
		s.WriteString(strings.Repeat("\n", m.height-contentHeight-1))
	}
	s.WriteString(footer)

	return m.styles.Doc.Width(m.width).Height(m.height).Render(s.String())
}

func (m Model) renderTabs() string {
	var tabs []string
	labels := []string{"1. Scan", "2. Leaderboard", "3. Watchlist", "4. Settings"}
	states := []sessionState{stateScan, stateLeaderboard, stateWatchlist, stateSettings}

	for i, label := range labels {
		if m.state == states[i] {
			tabs = append(tabs, m.styles.ActiveTab.Render(label))
		} else {
			tabs = append(tabs, m.styles.Tab.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderContent() string {
	var content string
	switch m.state {
	case stateScan:
		content = "Scan View (Work in Progress)"
	case stateLeaderboard:
		if m.leaderboard != nil {
			return m.styles.Content.Render(m.leaderboard.View())
		}
		content = "Leaderboard View (Loading...)"
	case stateWatchlist:
		content = "Watchlist View (Work in Progress)"
	case stateSettings:
		content = "Settings View (Work in Progress)"
	case stateTraderDetail:
		content = m.renderTraderDetail()
	}

	return m.styles.Content.Render(content)
}

func (m Model) renderTraderDetail() string {
	if m.traderDetail != nil {
		return m.traderDetail.View()
	}
	if m.selectedTrader == nil {
		return "No trader selected"
	}
	return "Loading trader details..."
}

func formatStat(label, value string) string {
	return "  " + label + ": " + value
}

func formatPercent(val float64) string {
	return fmt.Sprintf("%.1f%%", val*100)
}

func (m Model) renderFooter() string {
	var help string
	switch m.state {
	case stateLeaderboard:
		if m.leaderboard != nil && m.db != nil {
			help = m.leaderboard.HelpText() + " | q: quit | 1-4: tabs"
		} else {
			help = "q: quit | 1-4: change tab | ?: help"
		}
	case stateTraderDetail:
		if m.traderDetail != nil {
			help = m.traderDetail.HelpText() + " | q: quit"
		} else {
			help = "esc: back | q: quit"
		}
	default:
		help = "q: quit | 1-4: change tab | ?: help"
	}
	return m.styles.Footer.Width(m.width).Render(help)
}

func Start(themeName string) error {
	p := tea.NewProgram(NewModel(themeName), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func StartWithDB(themeName string, database *db.DB) error {
	p := tea.NewProgram(NewModelWithDB(themeName, database), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
