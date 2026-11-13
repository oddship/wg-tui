package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestColumnEditorLoadsExistingSettings(t *testing.T) {
	for _, tc := range []struct {
		value, first string
		shown        int
	}{
		{"", "name", 4}, {"none", "name", 1}, {"kind, name", "kind", 2}, {"description, group", "name", 3},
	} {
		f := newTargetColumnsField(tc.value)
		if len(f.columns) != 4 || f.columns[0].field != tc.first {
			t.Fatalf("unexpected columns for %q: %+v", tc.value, f.columns)
		}
		shown := 0
		for _, c := range f.columns {
			if c.shown {
				shown++
			}
		}
		if shown != tc.shown {
			t.Fatalf("wrong shown count for %q: %d", tc.value, shown)
		}
	}
}

func TestColumnEditorToggleReorderAndKeepName(t *testing.T) {
	f := newTargetColumnsField("description, group, kind")
	original := f
	f.updateTargetColumns(" ")
	if !f.columns[0].shown {
		t.Fatal("Name must stay enabled")
	}
	f.updateTargetColumns("down")
	f.updateTargetColumns(" ")
	if f.input.Value() != "name, group, kind" {
		t.Fatalf("hide did not update setting: %s", f.input.Value())
	}
	if !original.columns[1].shown {
		t.Fatal("toggle mutated previous model")
	}
	f.updateTargetColumns(" ")
	f.updateTargetColumns("-")
	if f.columnIndex != 0 || f.input.Value() != "description, name, group, kind" {
		t.Fatalf("reorder failed: %s", f.input.Value())
	}
	f.updateTargetColumns("-")
	if f.columnIndex != 0 {
		t.Fatal("moved past first column")
	}
	f.updateTargetColumns("+")
	if f.input.Value() != "name, description, group, kind" {
		t.Fatal("move down did not restore order")
	}
}

func TestColumnEditorNavigationPreviewAndSave(t *testing.T) {
	m := New("test")
	m.cfg = cfgpkg.Default()
	m.cfg.Server.URL = "https://warpgate.example.com"
	m.cfg.Server.Token = "token"
	m.cfg.SSH.Username = "user@example.com"
	m.mode = modeConfig
	m.startForm(true)
	index := requireFieldIndex(t, m.fields, "ui.target_fields")
	m.focusField(index)
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.formIndex != index || m.fields[index].columnIndex != 1 {
		t.Fatal("down should select a column")
	}
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeySpace})
	cfg, err := m.formConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UI.TargetFields != "name, group, kind" {
		t.Fatalf("toggle not saved: %s", cfg.UI.TargetFields)
	}
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 20}, {Width: 100, Height: 24}} {
		m.setWindowSize(size)
		view := m.View()
		if lipgloss.Width(view) > size.Width || lipgloss.Height(view) > size.Height {
			t.Fatalf("editor exceeds %dx%d: %dx%d", size.Width, size.Height, lipgloss.Width(view), lipgloss.Height(view))
		}
		if strings.Contains(view, "Primary API") || !strings.Contains(view, "[ ] Description") {
			t.Fatal("preview or checkbox did not update")
		}
	}
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.formIndex != index+1 {
		t.Fatal("tab should leave column editor")
	}
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyShiftTab})
	_, cmd := m.updateConfigForm(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("Ctrl+S should save from column editor")
	}
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeBrowse || m.cfg.UI.TargetFields != cfgpkg.Default().UI.TargetFields {
		t.Fatal("cancel should discard edits")
	}
}
