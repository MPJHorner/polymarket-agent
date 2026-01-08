package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateScan sessionState = iota
	stateLeaderboard
	stateWatchlist
	stateSettings
)

type Model struct {
	state  sessionState
	styles Styles
	width  int
	height int
	err    error
}

func NewModel(themeName string) Model {
	t, ok := Themes[strings.ToLower(themeName)]
	if !ok {
		t = Dracula
	}
	return Model{
		state:  stateLeaderboard,
		styles: GetStyles(t),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.state = stateScan
		case "2":
			m.state = stateLeaderboard
		case "3":
			m.state = stateWatchlist
		case "4":
			m.state = stateSettings
		case "?":
			// Toggle help? For now just stay.
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
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
		content = "Leaderboard View (Work in Progress)"
	case stateWatchlist:
		content = "Watchlist View (Work in Progress)"
	case stateSettings:
		content = "Settings View (Work in Progress)"
	}

	return m.styles.Content.Render(content)
}

func (m Model) renderFooter() string {
	help := "q: quit • 1-4: change tab • ?: help"
	return m.styles.Footer.Width(m.width).Render(help)
}

func Start(themeName string) error {
	p := tea.NewProgram(NewModel(themeName), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
