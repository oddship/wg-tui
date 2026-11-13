package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
)

func TestRefreshKeepsSelectedTargetByName(t *testing.T) {
	m := newPagingModel(t, 3)
	m.cfg.Cache.Dir = t.TempDir()
	m.selected = 1
	next, _ := m.handleRefresh(refreshMsg{snapshot: cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "inserted"}, {Name: "host-00"}, {Name: "host-01"}, {Name: "host-02"}}}})
	updated := next.(Model)
	selected, ok := updated.selectedTarget()
	if !ok || selected.Name != "host-01" || updated.selected != 2 {
		t.Fatalf("refresh changed selection: %+v", selected)
	}
}

func TestRemovedTargetRequiresExplicitReselection(t *testing.T) {
	m := newPagingModel(t, 3)
	m.cfg.Cache.Dir = t.TempDir()
	m.selected = 1
	for i := 0; i < 2; i++ {
		next, _ := m.handleRefresh(refreshMsg{snapshot: cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "host-00"}, {Name: "host-02"}}}})
		m = next.(Model)
		if _, ok := m.selectedTarget(); ok {
			t.Fatal("refresh selected a replacement without input")
		}
	}
	if !strings.Contains(m.View(), "No target selected") {
		t.Fatal("missing selection is not explained")
	}
	next, cmd := m.connectSelected()
	if cmd != nil {
		t.Fatal("connect must not run for removed target")
	}
	m = next.(Model)
	m.startTunnelFormForSelection()
	if m.mode != modeBrowse {
		t.Fatal("tunnel form opened for removed target")
	}
	next, cmd = m.startRsyncFormForSelection()
	if cmd != nil || next.(Model).mode != modeBrowse {
		t.Fatal("transfer opened for removed target")
	}
	m.copySelectedCommand() // Must return before accessing the clipboard.
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	selected, ok := m.selectedTarget()
	if !ok || selected.Name != "host-00" {
		t.Fatal("arrow key did not restore selection")
	}
}
