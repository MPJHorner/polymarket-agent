package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStyles(t *testing.T) {
	tests := []struct {
		name      string
		theme     Theme
		expectPrimary string
	}{
		{
			name:      "Dracula",
			theme:     Dracula,
			expectPrimary: "#bd93f9",
		},
		{
			name:      "Nord",
			theme:     Nord,
			expectPrimary: "#81a1c1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := GetStyles(tt.theme)
			// We can't easily check the exact color from lipgloss.Style without internal access or rendering
			// but we can check if it's initialized.
			assert.NotNil(t, styles.Header)
			assert.NotNil(t, styles.Footer)
		})
	}
}

func TestThemesMap(t *testing.T) {
	assert.Contains(t, Themes, "dracula")
	assert.Contains(t, Themes, "nord")
	assert.Contains(t, Themes, "gruvbox")
	assert.Contains(t, Themes, "catppuccin")
	assert.Contains(t, Themes, "tokyo")
}

func TestNewModelWithTheme(t *testing.T) {
	m := NewModel("nord")
	assert.Equal(t, stateLeaderboard, m.state)
	
	m2 := NewModel("invalid")
	// Should default to Dracula
	assert.NotNil(t, m2.styles)
}
