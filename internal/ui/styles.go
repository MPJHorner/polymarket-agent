package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Bg        lipgloss.Color
	Fg        lipgloss.Color
	Gray      lipgloss.Color
}

var (
	Dracula = Theme{
		Name:      "Dracula",
		Primary:   lipgloss.Color("#bd93f9"),
		Secondary: lipgloss.Color("#ff79c6"),
		Accent:    lipgloss.Color("#50fa7b"),
		Bg:        lipgloss.Color("#282a36"),
		Fg:        lipgloss.Color("#f8f8f2"),
		Gray:      lipgloss.Color("#6272a4"),
	}

	Nord = Theme{
		Name:      "Nord",
		Primary:   lipgloss.Color("#81a1c1"),
		Secondary: lipgloss.Color("#88c0d0"),
		Accent:    lipgloss.Color("#a3be8c"),
		Bg:        lipgloss.Color("#2e3440"),
		Fg:        lipgloss.Color("#eceff4"),
		Gray:      lipgloss.Color("#4c566a"),
	}

	Gruvbox = Theme{
		Name:      "Gruvbox",
		Primary:   lipgloss.Color("#458588"),
		Secondary: lipgloss.Color("#b16286"),
		Accent:    lipgloss.Color("#689d6a"),
		Bg:        lipgloss.Color("#282828"),
		Fg:        lipgloss.Color("#ebdbb2"),
		Gray:      lipgloss.Color("#928374"),
	}

	Catppuccin = Theme{
		Name:      "Catppuccin",
		Primary:   lipgloss.Color("#cba6f7"),
		Secondary: lipgloss.Color("#f5c2e7"),
		Accent:    lipgloss.Color("#a6e3a1"),
		Bg:        lipgloss.Color("#1e1e2e"),
		Fg:        lipgloss.Color("#cdd6f4"),
		Gray:      lipgloss.Color("#585b70"),
	}

	TokyoNight = Theme{
		Name:      "Tokyo Night",
		Primary:   lipgloss.Color("#7aa2f7"),
		Secondary: lipgloss.Color("#bb9af7"),
		Accent:    lipgloss.Color("#9ece6a"),
		Bg:        lipgloss.Color("#1a1b26"),
		Fg:        lipgloss.Color("#c0caf5"),
		Gray:      lipgloss.Color("#565f89"),
	}
)

var Themes = map[string]Theme{
	"dracula":    Dracula,
	"nord":       Nord,
	"gruvbox":    Gruvbox,
	"catppuccin": Catppuccin,
	"tokyo":      TokyoNight,
}

type Styles struct {
	Header     lipgloss.Style
	Footer     lipgloss.Style
	Tab        lipgloss.Style
	ActiveTab  lipgloss.Style
	Content    lipgloss.Style
	Doc        lipgloss.Style
	Highlight  lipgloss.Style
	Subtle     lipgloss.Style
}

func GetStyles(theme Theme) Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(theme.Primary).
			Padding(0, 1).
			Bold(true),

		Footer: lipgloss.NewStyle().
			Foreground(theme.Gray).
			Padding(0, 1).
			Italic(true),

		Tab: lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(theme.Fg),

		ActiveTab: lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(theme.Secondary).
			Bold(true).
			Foreground(theme.Secondary),

		Content: lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(theme.Fg),

		Doc: lipgloss.NewStyle().
			Margin(0, 0).
			Background(theme.Bg).
			Foreground(theme.Fg),

		Highlight: lipgloss.NewStyle().
			Foreground(theme.Accent),

		Subtle: lipgloss.NewStyle().
			Foreground(theme.Gray),
	}
}

// DefaultStyles returns styles for the default theme (Dracula)
func DefaultStyles() Styles {
	return GetStyles(Dracula)
}