package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#F456D0")
	ColorAccent    = lipgloss.Color("#56F4A0")
	ColorBg        = lipgloss.Color("#1A1B26")
	ColorFg        = lipgloss.Color("#A9B1D6")
	ColorGray      = lipgloss.Color("#565F89")

	// Styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(ColorPrimary).
			Padding(0, 1).
			Bold(true)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Padding(0, 1).
			Italic(true)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 2)

	ActiveTabStyle = TabStyle.Copy().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorSecondary).
			Bold(true).
			Foreground(ColorSecondary)

	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	DocStyle = lipgloss.NewStyle().
			Margin(0, 0)
)
