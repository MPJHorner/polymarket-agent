package ui

import (
	"fmt"

	"polytracker/internal/db"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	colRank     = "rank"
	colAddress  = "address"
	colUsername = "username"
	colWinRate  = "win_rate"
	colPNL      = "pnl"
	colROI      = "roi"
	colVolume   = "volume"

	pageSize = 20
)

type LeaderboardKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Enter    key.Binding
	SortWin  key.Binding
	SortPNL  key.Binding
}

var leaderboardKeys = LeaderboardKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", "page down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view details"),
	),
	SortWin: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "sort by win%"),
	),
	SortPNL: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "sort by P&L"),
	),
}

type Leaderboard struct {
	table       table.Model
	traders     []db.Trader
	sortField   db.SortField
	sortOrder   db.SortOrder
	currentPage int
	totalPages  int
	totalCount  int
	width       int
	height      int
	styles      Styles
	selected    *db.Trader
}

type tradersLoadedMsg struct {
	traders    []db.Trader
	totalCount int
}

type TraderSelectedMsg struct {
	Trader *db.Trader
}

func NewLeaderboard(styles Styles) *Leaderboard {
	l := &Leaderboard{
		sortField:   db.SortByProfitLoss,
		sortOrder:   db.SortDesc,
		currentPage: 0,
		styles:      styles,
	}
	l.table = l.createTable()
	return l
}

func (l *Leaderboard) createTable() table.Model {
	columns := []table.Column{
		table.NewColumn(colRank, "#", 4).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewColumn(colAddress, "Address", 14),
		table.NewColumn(colUsername, "Username", 16),
		table.NewColumn(colWinRate, "Win %", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewColumn(colPNL, "P&L", 12).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewColumn(colROI, "ROI %", 10).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewColumn(colVolume, "Volume", 14).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
	}

	t := table.New(columns).
		WithRows([]table.Row{}).
		Focused(true).
		WithPageSize(pageSize).
		HeaderStyle(lipgloss.NewStyle().Bold(true).Foreground(l.styles.Highlight.GetForeground())).
		HighlightStyle(lipgloss.NewStyle().
			Bold(true).
			Background(l.styles.ActiveTab.GetBorderBottomForeground()).
			Foreground(lipgloss.Color("#FFF")))

	return t
}

func (l *Leaderboard) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.table = l.table.WithTargetWidth(width - 4)
}

func (l *Leaderboard) LoadTraders(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		count, err := database.CountTraders()
		if err != nil {
			return nil
		}

		traders, err := database.ListTradersWithOptions(db.ListTradersOptions{
			SortBy: l.sortField,
			Order:  l.sortOrder,
			Limit:  pageSize,
			Offset: l.currentPage * pageSize,
		})
		if err != nil {
			return nil
		}

		return tradersLoadedMsg{
			traders:    traders,
			totalCount: count,
		}
	}
}

func (l *Leaderboard) Update(msg tea.Msg) (*Leaderboard, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tradersLoadedMsg:
		l.traders = msg.traders
		l.totalCount = msg.totalCount
		l.totalPages = (msg.totalCount + pageSize - 1) / pageSize
		if l.totalPages == 0 {
			l.totalPages = 1
		}
		l.table = l.table.WithRows(l.buildRows())
		return l, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, leaderboardKeys.Enter):
			if row := l.table.HighlightedRow(); row.Data != nil {
				if idx, ok := row.Data[colRank].(int); ok && idx > 0 && idx <= len(l.traders) {
					trader := l.traders[idx-1-(l.currentPage*pageSize)]
					l.selected = &trader
					return l, func() tea.Msg {
						return TraderSelectedMsg{Trader: &trader}
					}
				}
			}
		case key.Matches(msg, leaderboardKeys.SortWin):
			if l.sortField == db.SortByWinRate {
				l.toggleSortOrder()
			} else {
				l.sortField = db.SortByWinRate
				l.sortOrder = db.SortDesc
			}
			l.currentPage = 0
			return l, nil
		case key.Matches(msg, leaderboardKeys.SortPNL):
			if l.sortField == db.SortByProfitLoss {
				l.toggleSortOrder()
			} else {
				l.sortField = db.SortByProfitLoss
				l.sortOrder = db.SortDesc
			}
			l.currentPage = 0
			return l, nil
		}
	}

	l.table, cmd = l.table.Update(msg)
	return l, cmd
}

func (l *Leaderboard) toggleSortOrder() {
	if l.sortOrder == db.SortDesc {
		l.sortOrder = db.SortAsc
	} else {
		l.sortOrder = db.SortDesc
	}
}

func (l *Leaderboard) buildRows() []table.Row {
	rows := make([]table.Row, len(l.traders))
	for i, t := range l.traders {
		rank := l.currentPage*pageSize + i + 1
		address := t.Address
		if len(address) > 12 {
			address = address[:6] + "..." + address[len(address)-4:]
		}

		username := t.Username
		if username == "" {
			username = "-"
		}

		rows[i] = table.NewRow(table.RowData{
			colRank:     rank,
			colAddress:  address,
			colUsername: username,
			colWinRate:  fmt.Sprintf("%.1f%%", t.WinRate*100),
			colPNL:      formatPNL(t.ProfitLoss),
			colROI:      fmt.Sprintf("%.1f%%", t.ROI*100),
			colVolume:   formatVolume(t.Volume),
		})
	}
	return rows
}

func formatPNL(pnl float64) string {
	if pnl >= 0 {
		return fmt.Sprintf("+$%.2f", pnl)
	}
	return fmt.Sprintf("-$%.2f", -pnl)
}

func formatVolume(vol float64) string {
	if vol >= 1000000 {
		return fmt.Sprintf("$%.1fM", vol/1000000)
	}
	if vol >= 1000 {
		return fmt.Sprintf("$%.1fK", vol/1000)
	}
	return fmt.Sprintf("$%.0f", vol)
}

func (l *Leaderboard) View() string {
	var sortIndicator string
	switch l.sortField {
	case db.SortByWinRate:
		sortIndicator = "Win %"
	case db.SortByProfitLoss:
		sortIndicator = "P&L"
	case db.SortByROI:
		sortIndicator = "ROI"
	case db.SortByVolume:
		sortIndicator = "Volume"
	}
	if l.sortOrder == db.SortDesc {
		sortIndicator += " ↓"
	} else {
		sortIndicator += " ↑"
	}

	header := l.styles.Subtle.Render(fmt.Sprintf(
		"Sorted by: %s | Page %d/%d | Total: %d traders",
		sortIndicator,
		l.currentPage+1,
		l.totalPages,
		l.totalCount,
	))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		l.table.View(),
	)
}

func (l *Leaderboard) GetSortField() db.SortField {
	return l.sortField
}

func (l *Leaderboard) GetSortOrder() db.SortOrder {
	return l.sortOrder
}

func (l *Leaderboard) GetCurrentPage() int {
	return l.currentPage
}

func (l *Leaderboard) NextPage() {
	if l.currentPage < l.totalPages-1 {
		l.currentPage++
	}
}

func (l *Leaderboard) PrevPage() {
	if l.currentPage > 0 {
		l.currentPage--
	}
}

func (l *Leaderboard) HelpText() string {
	return "↑/↓: navigate • enter: view details • w: sort by win% • p: sort by P&L"
}
