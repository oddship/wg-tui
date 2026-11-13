package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oddship/wg-tui/internal/api"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func newPagingModel(t *testing.T, count int) Model {
	t.Helper()
	m := New("test")
	m.mode = modeBrowse
	m.applyConfig(cfgpkg.Default())
	targets := make([]api.Target, 0, count)
	for i := range count {
		targets = append(targets, makeRenderTarget(fmt.Sprintf("host-%02d", i), "alpha", ""))
	}
	m.setTargets(targets)
	m.applyFilter()
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 20})
	return m
}

func pressKey(t *testing.T, m Model, msg tea.KeyMsg) Model {
	t.Helper()
	next, _ := m.Update(msg)
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("expected ui.Model from Update, got %T", next)
	}
	return updated
}

func TestPageDownMovesSelectionByVisibleRows(t *testing.T) {
	m := newPagingModel(t, 100)
	page := m.visibleRows()

	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyPgDown})
	if m.selected != page {
		t.Fatalf("expected selection at %d after page down, got %d", page, m.selected)
	}

	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyPgUp})
	if m.selected != 0 {
		t.Fatalf("expected selection back at 0 after page up, got %d", m.selected)
	}
}

func TestPageKeysClampAtListBounds(t *testing.T) {
	m := newPagingModel(t, 4)
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 40})

	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyPgDown})
	if m.selected != 3 {
		t.Fatalf("expected selection clamped to last row, got %d", m.selected)
	}

	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyPgUp})
	if m.selected != 0 {
		t.Fatalf("expected selection clamped to first row, got %d", m.selected)
	}
}

func TestPageDownAliasWorksInBrowseMode(t *testing.T) {
	m := newPagingModel(t, 100)
	page := m.visibleRows()

	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyCtrlF})
	if m.selected != page {
		t.Fatalf("expected ctrl+f to page down to %d, got %d", page, m.selected)
	}
}

// With search focused, ctrl+f belongs to the text cursor, not the list.
func TestPageAliasIsLeftToSearchInputWhenSearching(t *testing.T) {
	m := newPagingModel(t, 100)
	m.focusSearch("")

	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyCtrlF})
	if m.selected != 0 {
		t.Fatalf("expected ctrl+f not to page the list while searching, got selection %d", m.selected)
	}

	page := m.visibleRows()
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyPgDown})
	if m.selected != page {
		t.Fatalf("expected pgdown to page down to %d while searching, got %d", page, m.selected)
	}
}
