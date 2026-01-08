package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelUpdate(t *testing.T) {
	m := NewModel()

	// Test tab switching
	t.Run("tab switching", func(t *testing.T) {
		msgs := []struct {
			msg      tea.Msg
			expected sessionState
		}{
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")}, stateScan},
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}, stateLeaderboard},
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")}, stateWatchlist},
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")}, stateSettings},
		}

		for _, tc := range msgs {
			newModel, _ := m.Update(tc.msg)
			m = newModel.(Model)
			if m.state != tc.expected {
				t.Errorf("Expected state %v, got %v", tc.expected, m.state)
			}
		}
	})

	// Test window resizing
	t.Run("window resize", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)
		if m.width != 100 || m.height != 50 {
			t.Errorf("Expected size 100x50, got %dx%d", m.width, m.height)
		}
	})
}

func TestModelView(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24

	view := m.View()
	
	if view == "" {
		t.Fatal("View returned empty string")
	}

	expectedParts := []string{
		"POLYTRACKER",
		"Scan",
		"Leaderboard",
		"Watchlist",
		"Settings",
		"q: quit",
	}

	for _, part := range expectedParts {
		if !strings.Contains(view, part) {
			t.Errorf("View missing expected part: %s", part)
		}
	}
}
