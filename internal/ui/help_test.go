package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestBrowseHelpShowsOmittedActionsAndFitsLayout(t *testing.T) {
	for _, position := range []string{"bottom", "right"} {
		m := newPagingModel(t, 100)
		m.cfg.UI.DetailsPane = position
		m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 24})
		before := m.View()
		m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		view := m.View()
		for _, label := range []string{"Keyboard shortcuts", "page up", "page down", "copy ssh", "help"} {
			if !strings.Contains(view, label) {
				t.Fatalf("missing %q in help", label)
			}
		}
		if lipgloss.Height(view) > 24 || lipgloss.Width(view) > 80 {
			t.Fatalf("expanded help overflow: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
		}
		m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		if before != m.View() {
			t.Fatal("help did not close")
		}
	}
}

func TestTunnelHelpShowsFullActionsAndFitsLayout(t *testing.T) {
	m := newPagingModel(t, 1)
	m.mode = modeTunnel
	m.tunnel.target = "host"
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 24})
	before := m.View()
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !strings.Contains(m.View(), "Keyboard shortcuts") || !strings.Contains(m.View(), "help") {
		t.Fatal("tunnel full help missing")
	}
	if lipgloss.Height(m.View()) > 24 || lipgloss.Width(m.View()) > 80 {
		t.Fatal("tunnel help exceeds terminal")
	}
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m.View() != before {
		t.Fatal("tunnel help did not close")
	}
}
