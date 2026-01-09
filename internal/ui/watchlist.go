package ui

import (
	"fmt"

	"polytracker/internal/db"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	colWatchIndex    = "index"
	colWatchTrader   = "trader"
	colWatchUsername = "username"
	colWatchNotes    = "notes"
	colWatchAdded    = "added"
	colWatchWinRate  = "win_rate"
	colWatchPNL      = "pnl"
)

type WatchlistKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Remove   key.Binding
	EditNote key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
}

var watchlistKeys = WatchlistKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view details"),
	),
	Remove: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "remove"),
	),
	EditNote: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "edit note"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

type Watchlist struct {
	table           table.Model
	items           []db.WatchlistItem
	traders         map[string]*db.Trader
	width           int
	height          int
	styles          Styles
	editingNote     bool
	noteInput       textinput.Model
	selectedTraderID string
}

type watchlistLoadedMsg struct {
	items   []db.WatchlistItem
	traders map[string]*db.Trader
}

type WatchlistTraderSelectedMsg struct {
	Trader *db.Trader
}

type WatchlistRemoveMsg struct {
	TraderID string
}

type WatchlistNoteUpdatedMsg struct {
	TraderID string
	Note     string
}

func NewWatchlist(styles Styles) *Watchlist {
	ti := textinput.New()
	ti.Placeholder = "Enter note..."
	ti.CharLimit = 200
	ti.Width = 50

	w := &Watchlist{
		styles:    styles,
		traders:   make(map[string]*db.Trader),
		noteInput: ti,
	}
	w.table = w.createTable()
	return w
}

func (w *Watchlist) createTable() table.Model {
	columns := []table.Column{
		table.NewColumn(colWatchTrader, "Address", 14),
		table.NewColumn(colWatchUsername, "Username", 16),
		table.NewColumn(colWatchWinRate, "Win %", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewColumn(colWatchPNL, "P&L", 12).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewColumn(colWatchNotes, "Notes", 30),
		table.NewColumn(colWatchAdded, "Added", 12),
	}

	t := table.New(columns).
		WithRows([]table.Row{}).
		Focused(true).
		WithPageSize(15).
		HeaderStyle(lipgloss.NewStyle().Bold(true).Foreground(w.styles.Highlight.GetForeground())).
		HighlightStyle(lipgloss.NewStyle().
			Bold(true).
			Background(w.styles.ActiveTab.GetBorderBottomForeground()).
			Foreground(lipgloss.Color("#FFF")))

	return t
}

func (w *Watchlist) SetSize(width, height int) {
	w.width = width
	w.height = height
	w.table = w.table.WithTargetWidth(width - 4)
	w.noteInput.Width = width - 20
}

func (w *Watchlist) LoadWatchlist(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		items, err := database.ListWatchlist()
		if err != nil {
			return watchlistLoadedMsg{items: nil, traders: nil}
		}

		traders := make(map[string]*db.Trader)
		for _, item := range items {
			trader, err := database.GetTrader(item.TraderID)
			if err == nil && trader != nil {
				traders[item.TraderID] = trader
			}
		}

		return watchlistLoadedMsg{
			items:   items,
			traders: traders,
		}
	}
}

func (w *Watchlist) Update(msg tea.Msg) (*Watchlist, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case watchlistLoadedMsg:
		w.items = msg.items
		w.traders = msg.traders
		w.table = w.table.WithRows(w.buildRows())
		return w, nil

	case tea.KeyMsg:
		// Handle note editing mode
		if w.editingNote {
			switch {
			case key.Matches(msg, watchlistKeys.Cancel):
				w.editingNote = false
				w.noteInput.Blur()
				return w, nil
			case key.Matches(msg, watchlistKeys.Confirm):
				w.editingNote = false
				w.noteInput.Blur()
				note := w.noteInput.Value()
				traderID := w.selectedTraderID
				return w, func() tea.Msg {
					return WatchlistNoteUpdatedMsg{TraderID: traderID, Note: note}
				}
			default:
				w.noteInput, cmd = w.noteInput.Update(msg)
				return w, cmd
			}
		}

		// Normal mode
		switch {
		case key.Matches(msg, watchlistKeys.Enter):
			if row := w.table.HighlightedRow(); row.Data != nil {
				if idx, ok := row.Data[colWatchIndex].(int); ok && idx >= 0 && idx < len(w.items) {
					traderID := w.items[idx].TraderID
					if trader, exists := w.traders[traderID]; exists {
						return w, func() tea.Msg {
							return WatchlistTraderSelectedMsg{Trader: trader}
						}
					}
				}
			}
		case key.Matches(msg, watchlistKeys.Remove):
			if row := w.table.HighlightedRow(); row.Data != nil {
				if idx, ok := row.Data[colWatchIndex].(int); ok && idx >= 0 && idx < len(w.items) {
					traderID := w.items[idx].TraderID
					return w, func() tea.Msg {
						return WatchlistRemoveMsg{TraderID: traderID}
					}
				}
			}
		case key.Matches(msg, watchlistKeys.EditNote):
			if row := w.table.HighlightedRow(); row.Data != nil {
				if idx, ok := row.Data[colWatchIndex].(int); ok && idx >= 0 && idx < len(w.items) {
					traderID := w.items[idx].TraderID
					w.editingNote = true
					w.selectedTraderID = traderID
					w.noteInput.SetValue(w.items[idx].Notes)
					w.noteInput.Focus()
					return w, textinput.Blink
				}
			}
		}
	}

	w.table, cmd = w.table.Update(msg)
	return w, cmd
}

func (w *Watchlist) buildRows() []table.Row {
	rows := make([]table.Row, len(w.items))
	for i, item := range w.items {
		address := item.TraderID
		shortAddress := address
		if len(address) > 12 {
			shortAddress = address[:6] + "..." + address[len(address)-4:]
		}

		username := "-"
		winRate := "-"
		pnl := "-"

		if trader, exists := w.traders[item.TraderID]; exists && trader != nil {
			if trader.Username != "" {
				username = trader.Username
			}
			winRate = fmt.Sprintf("%.1f%%", trader.WinRate*100)
			pnl = formatPNL(trader.ProfitLoss)
		}

		notes := item.Notes
		if len(notes) > 28 {
			notes = notes[:28] + ".."
		}
		if notes == "" {
			notes = "-"
		}

		rows[i] = table.NewRow(table.RowData{
			colWatchIndex:    i,
			colWatchTrader:   shortAddress,
			colWatchUsername: username,
			colWatchWinRate:  winRate,
			colWatchPNL:      pnl,
			colWatchNotes:    notes,
			colWatchAdded:    item.CreatedAt.Format("01/02/06"),
		}).WithStyle(lipgloss.NewStyle())
	}
	return rows
}

func (w *Watchlist) View() string {
	var sections []string

	header := w.styles.Subtle.Render(fmt.Sprintf(
		"Watching %d traders",
		len(w.items),
	))
	sections = append(sections, header)
	sections = append(sections, "")

	if len(w.items) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(w.styles.Subtle.GetForeground()).
			Padding(2, 4).
			Render("No traders in watchlist.\n\nVisit the Leaderboard (2) and press 'w' on a trader to add them.")
		sections = append(sections, emptyMsg)
	} else {
		sections = append(sections, w.table.View())
	}

	// Show note editing input if active
	if w.editingNote {
		sections = append(sections, "")
		editBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(w.styles.Highlight.GetForeground()).
			Padding(1, 2).
			Width(w.width - 10)

		editContent := lipgloss.JoinVertical(
			lipgloss.Left,
			w.styles.Header.Render(" Edit Note "),
			"",
			w.noteInput.View(),
			"",
			w.styles.Subtle.Render("Enter: save | Esc: cancel"),
		)
		sections = append(sections, editBox.Render(editContent))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (w *Watchlist) HelpText() string {
	if w.editingNote {
		return "enter: save note | esc: cancel"
	}
	return "↑/↓: navigate | enter: view details | d: remove | n: edit note"
}

func (w *Watchlist) GetItems() []db.WatchlistItem {
	return w.items
}

func (w *Watchlist) IsEditingNote() bool {
	return w.editingNote
}

// UpdateNote updates the note for a specific trader in the local state
func (w *Watchlist) UpdateNote(traderID, note string) {
	for i := range w.items {
		if w.items[i].TraderID == traderID {
			w.items[i].Notes = note
			break
		}
	}
	w.table = w.table.WithRows(w.buildRows())
}

// RemoveItem removes a trader from the local watchlist state
func (w *Watchlist) RemoveItem(traderID string) {
	for i := range w.items {
		if w.items[i].TraderID == traderID {
			w.items = append(w.items[:i], w.items[i+1:]...)
			break
		}
	}
	delete(w.traders, traderID)
	w.table = w.table.WithRows(w.buildRows())
}
