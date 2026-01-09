package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"polytracker/internal/claude"
	"polytracker/internal/db"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type analysisState int

const (
	analysisStateFetching analysisState = iota
	analysisStateAnalyzing
	analysisStateComplete
	analysisStateError
)

type AnalysisKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Back  key.Binding
	Save  key.Binding
	Copy  key.Binding
	Retry key.Binding
}

var analysisKeys = AnalysisKeyMap{
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
	Save: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "save"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	Retry: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	),
}

type Analysis struct {
	trader       *db.Trader
	trades       []db.Trade
	markets      map[string]*db.Market
	thesis       string
	state        analysisState
	styles       Styles
	width        int
	height       int
	scrollOffset int
	spinner      spinner.Model
	err          error
	savedPath    string
	copied       bool
	claudeClient *claude.Client
}

// Messages for analysis flow
type AnalysisDataFetchedMsg struct {
	Trades  []db.Trade
	Markets map[string]*db.Market
}

type AnalysisCompleteMsg struct {
	Thesis string
}

type AnalysisErrorMsg struct {
	Err error
}

type AnalysisSavedMsg struct {
	Path string
}

type AnalysisCopiedMsg struct{}

type AnalysisRetryMsg struct{}

func NewAnalysis(trader *db.Trader, styles Styles, claudeClient *claude.Client) *Analysis {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Header.GetBackground())

	return &Analysis{
		trader:       trader,
		trades:       nil,
		markets:      make(map[string]*db.Market),
		thesis:       "",
		state:        analysisStateFetching,
		styles:       styles,
		scrollOffset: 0,
		spinner:      s,
		claudeClient: claudeClient,
	}
}

func (a *Analysis) SetSize(width, height int) {
	a.width = width
	a.height = height
}

// FetchData fetches trader data from the database
func (a *Analysis) FetchData(database *db.DB) tea.Cmd {
	return func() tea.Msg {
		if a.trader == nil {
			return AnalysisErrorMsg{Err: fmt.Errorf("no trader selected")}
		}

		trades, err := database.GetTradesByTrader(a.trader.Address)
		if err != nil {
			return AnalysisErrorMsg{Err: fmt.Errorf("failed to fetch trades: %w", err)}
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

		return AnalysisDataFetchedMsg{
			Trades:  trades,
			Markets: markets,
		}
	}
}

// RunAnalysis sends the data to Claude for analysis
func (a *Analysis) RunAnalysis() tea.Cmd {
	return func() tea.Msg {
		if a.claudeClient == nil {
			return AnalysisErrorMsg{Err: claude.ErrNoAPIKey}
		}

		data := claude.TraderData{
			Trader:  a.trader,
			Trades:  a.trades,
			Markets: a.markets,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		result, err := a.claudeClient.AnalyzeTrader(ctx, data)
		if err != nil {
			return AnalysisErrorMsg{Err: err}
		}

		return AnalysisCompleteMsg{Thesis: result.Thesis}
	}
}

// SaveThesis saves the thesis to a markdown file
func (a *Analysis) SaveThesis() tea.Cmd {
	return func() tea.Msg {
		if a.thesis == "" {
			return AnalysisErrorMsg{Err: fmt.Errorf("no thesis to save")}
		}

		// Create exports directory if it doesn't exist
		exportDir := "exports"
		if err := os.MkdirAll(exportDir, 0755); err != nil {
			return AnalysisErrorMsg{Err: fmt.Errorf("failed to create exports directory: %w", err)}
		}

		// Generate filename with trader address and timestamp
		shortAddr := a.trader.Address
		if len(shortAddr) > 10 {
			shortAddr = shortAddr[:10]
		}
		filename := fmt.Sprintf("thesis_%s_%s.md", shortAddr, time.Now().Format("20060102_150405"))
		path := filepath.Join(exportDir, filename)

		// Build markdown content
		var content strings.Builder
		content.WriteString(fmt.Sprintf("# Trading Thesis: %s\n\n", a.trader.Address))
		content.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		content.WriteString("---\n\n")
		content.WriteString(a.thesis)

		if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
			return AnalysisErrorMsg{Err: fmt.Errorf("failed to save thesis: %w", err)}
		}

		return AnalysisSavedMsg{Path: path}
	}
}

func (a *Analysis) Init() tea.Cmd {
	return a.spinner.Tick
}

func (a *Analysis) Update(msg tea.Msg) (*Analysis, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if a.state == analysisStateFetching || a.state == analysisStateAnalyzing {
			a.spinner, cmd = a.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case AnalysisDataFetchedMsg:
		a.trades = msg.Trades
		a.markets = msg.Markets
		a.state = analysisStateAnalyzing
		cmds = append(cmds, a.RunAnalysis())
		cmds = append(cmds, a.spinner.Tick)

	case AnalysisCompleteMsg:
		a.thesis = msg.Thesis
		a.state = analysisStateComplete
		a.scrollOffset = 0

	case AnalysisErrorMsg:
		a.err = msg.Err
		a.state = analysisStateError

	case AnalysisSavedMsg:
		a.savedPath = msg.Path

	case AnalysisCopiedMsg:
		a.copied = true

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, analysisKeys.Back):
			return a, func() tea.Msg { return GoBackMsg{} }

		case key.Matches(msg, analysisKeys.Up):
			if a.scrollOffset > 0 {
				a.scrollOffset--
			}

		case key.Matches(msg, analysisKeys.Down):
			a.scrollOffset++

		case key.Matches(msg, analysisKeys.Save):
			if a.state == analysisStateComplete && a.thesis != "" {
				return a, a.SaveThesis()
			}

		case key.Matches(msg, analysisKeys.Copy):
			if a.state == analysisStateComplete && a.thesis != "" {
				a.copied = true
				// Note: Full clipboard support would require github.com/atotto/clipboard
				// For now we just set a flag to show the user the action was attempted
			}

		case key.Matches(msg, analysisKeys.Retry):
			if a.state == analysisStateError || a.state == analysisStateComplete {
				a.state = analysisStateFetching
				a.thesis = ""
				a.err = nil
				a.savedPath = ""
				a.copied = false
				return a, func() tea.Msg { return AnalysisRetryMsg{} }
			}
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *Analysis) View() string {
	if a.trader == nil {
		return "No trader selected"
	}

	var sections []string

	// Header
	sections = append(sections, a.renderHeader())

	switch a.state {
	case analysisStateFetching:
		sections = append(sections, a.renderProgress("Fetching trader data..."))
	case analysisStateAnalyzing:
		sections = append(sections, a.renderProgress("Analyzing with Claude AI..."))
	case analysisStateError:
		sections = append(sections, a.renderError())
	case analysisStateComplete:
		sections = append(sections, a.renderThesis())
	}

	// Status messages
	if a.savedPath != "" {
		sections = append(sections, "")
		sections = append(sections, a.styles.Highlight.Render(fmt.Sprintf("Saved to: %s", a.savedPath)))
	}
	if a.copied {
		sections = append(sections, "")
		sections = append(sections, a.styles.Highlight.Render("Thesis content ready for copy"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply scrolling
	lines := strings.Split(content, "\n")
	if a.scrollOffset >= len(lines) {
		a.scrollOffset = len(lines) - 1
	}
	if a.scrollOffset < 0 {
		a.scrollOffset = 0
	}

	visibleHeight := a.height - 4
	if visibleHeight < 1 {
		visibleHeight = 20
	}

	endIdx := a.scrollOffset + visibleHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	visibleLines := lines[a.scrollOffset:endIdx]
	return strings.Join(visibleLines, "\n")
}

func (a *Analysis) renderHeader() string {
	shortAddr := a.trader.Address
	if len(shortAddr) > 20 {
		shortAddr = shortAddr[:10] + "..." + shortAddr[len(shortAddr)-8:]
	}

	return a.styles.Header.Render(fmt.Sprintf(" ANALYSIS: %s ", shortAddr))
}

func (a *Analysis) renderProgress(message string) string {
	progressBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.styles.Header.GetBackground()).
		Padding(2, 4).
		Width(a.width - 10).
		Align(lipgloss.Center)

	spinnerView := a.spinner.View()
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		spinnerView+" "+message,
		"",
		a.styles.Subtle.Render("This may take a moment..."),
		"",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		"",
		progressBox.Render(content),
	)
}

func (a *Analysis) renderError() string {
	errorBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff5555")).
		Padding(2, 4).
		Width(a.width - 10)

	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))

	errMsg := "Unknown error"
	if a.err != nil {
		errMsg = a.err.Error()
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		errorStyle.Bold(true).Render("Error"),
		"",
		errorStyle.Render(errMsg),
		"",
		a.styles.Subtle.Render("Press 'r' to retry or 'esc' to go back"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		"",
		errorBox.Render(content),
	)
}

func (a *Analysis) renderThesis() string {
	thesisBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.styles.Header.GetBackground()).
		Padding(1, 2).
		Width(a.width - 6)

	// Render markdown-like content with basic styling
	renderedContent := a.renderMarkdown(a.thesis)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		thesisBox.Render(renderedContent),
	)
}

// renderMarkdown applies basic markdown styling to the thesis content
func (a *Analysis) renderMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(a.styles.Header.GetBackground())
	boldStyle := lipgloss.NewStyle().Bold(true)
	listStyle := a.styles.Highlight

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Headers
		if strings.HasPrefix(trimmed, "## ") {
			result = append(result, "")
			result = append(result, headerStyle.Render(strings.TrimPrefix(trimmed, "## ")))
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			result = append(result, "")
			result = append(result, headerStyle.Render(strings.TrimPrefix(trimmed, "# ")))
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			result = append(result, "")
			result = append(result, boldStyle.Render(strings.TrimPrefix(trimmed, "### ")))
			continue
		}

		// Numbered list items
		if len(trimmed) > 2 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
			result = append(result, listStyle.Render(trimmed[:2])+trimmed[2:])
			continue
		}

		// Bullet points
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			result = append(result, listStyle.Render(trimmed[:2])+trimmed[2:])
			continue
		}

		// Bold text (simple pattern: **text**)
		processed := a.processBoldText(line)

		result = append(result, processed)
	}

	return strings.Join(result, "\n")
}

// processBoldText handles **bold** text patterns
func (a *Analysis) processBoldText(line string) string {
	boldStyle := lipgloss.NewStyle().Bold(true)
	result := line

	for {
		start := strings.Index(result, "**")
		if start == -1 {
			break
		}
		end := strings.Index(result[start+2:], "**")
		if end == -1 {
			break
		}
		end = start + 2 + end

		boldText := result[start+2 : end]
		result = result[:start] + boldStyle.Render(boldText) + result[end+2:]
	}

	return result
}

func (a *Analysis) HelpText() string {
	switch a.state {
	case analysisStateFetching, analysisStateAnalyzing:
		return "esc: cancel"
	case analysisStateError:
		return "r: retry | esc: back"
	case analysisStateComplete:
		return "s: save | r: retry | j/k: scroll | esc: back"
	default:
		return "esc: back"
	}
}

func (a *Analysis) GetTrader() *db.Trader {
	return a.trader
}

func (a *Analysis) GetThesis() string {
	return a.thesis
}
