package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/oddship/wg-tui/internal/api"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestTargetTableColumnOrderAndAlignment(t *testing.T) {
	ui := cfgpkg.Default().UI
	ui.TargetFields = "kind, name, group, description"
	targets := []api.Target{makeRenderTarget("host-a", "", "first\nnode"), makeRenderTarget("東京-server", "prod", "second")}
	rows := renderTargetTable(targets, 90, 1, func(s string) bool { return s == "host-a" }, ui)
	for i := range rows {
		rows[i] = ansiPattern.ReplaceAllString(rows[i], "")
	}
	headings := strings.Fields(strings.ReplaceAll(rows[0], "│", ""))
	if strings.Join(headings, ",") != "Kind,Name,Group,Description" {
		t.Fatalf("wrong columns: %q", rows[0])
	}
	if !strings.Contains(rows[1], "host-a ↺") || !strings.Contains(rows[1], "first node") || !strings.HasPrefix(rows[2], "• ") {
		t.Fatalf("wrong cells: %#v", rows)
	}
	positions := func(s string) []int {
		out := []int{}
		for _, part := range strings.Split(s, "│") {
			out = append(out, lipgloss.Width(part))
		}
		return out
	}
	for _, row := range rows[1:] {
		got, want := positions(row), positions(rows[0])
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("unaligned cells: %#v", rows)
			}
		}
	}
}

func TestTargetTableRetainsNameOnNarrowPanes(t *testing.T) {
	ui := cfgpkg.Default().UI
	ui.TargetFields = "description, kind, group, name"
	for width := 0; width <= 100; width++ {
		rows := renderTargetTable([]api.Target{makeRenderTarget("東京-host", "prod", "long description")}, width, 0, func(string) bool { return false }, ui)
		for _, row := range rows {
			if lipgloss.Width(row) > width {
				t.Fatalf("overflow at %d: %q", width, row)
			}
		}
		if width == 20 && !strings.Contains(rows[0], "Name") {
			t.Fatal("Name must survive hidden columns")
		}
	}
}

func TestBrowsePaneLayoutAndPagingFitTerminal(t *testing.T) {
	for _, position := range []string{"bottom", "right"} {
		for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 120, Height: 40}} {
			m := newPagingModel(t, 100)
			m.cfg.UI.DetailsPane = position
			m.setWindowSize(size)
			m.moveSelection(55)
			view := ansiPattern.ReplaceAllString(m.View(), "")
			if lipgloss.Width(view) > size.Width || lipgloss.Height(view) > size.Height {
				t.Fatalf("%s overflow %dx%d: got %dx%d", position, size.Width, size.Height, lipgloss.Width(view), lipgloss.Height(view))
			}
			lines := strings.Split(view, "\n")
			targets, details := -1, -1
			for i, line := range lines {
				if strings.Contains(line, "Targets") {
					targets = i
				}
				if strings.Contains(line, "Details") {
					details = i
				}
			}
			if position == "bottom" && details <= targets {
				t.Fatalf("expected Details below Targets")
			}
			if position == "right" && details != targets {
				t.Fatalf("expected side-by-side panes")
			}
			if position == "bottom" && !strings.Contains(view, "SSH:") {
				t.Fatal("bottom details must keep command visible")
			}
			if !strings.Contains(view, "host-55") {
				t.Fatal("selected row is hidden")
			}
		}
	}
}

func TestBottomDetailsCanScrollThroughFullMetadata(t *testing.T) {
	m := newPagingModel(t, 1)
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 24})
	target := makeRenderTarget("host", strings.Repeat("group-", 20), strings.Repeat("description words ", 40)+"FINAL-METADATA")
	target.ExternalHost = strings.Repeat("external-", 20)
	m.setTargets([]api.Target{target})
	for i := 0; i < 200; i++ {
		view := m.View()
		if lipgloss.Height(view) > 24 || lipgloss.Width(view) > 80 {
			t.Fatal("scrolling details overflow terminal")
		}
		if !strings.Contains(view, "SSH:") {
			t.Fatal("SSH command should stay pinned")
		}
		if strings.Contains(view, "FINAL-METADATA") {
			break
		}
		m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyDown, Alt: true})
	}
	if !strings.Contains(m.View(), "FINAL-METADATA") {
		t.Fatal("description tail is unreachable")
	}
	if m.selected != 0 {
		t.Fatal("detail scrolling changed selected target")
	}
	offset := m.detailScroll
	m = pressKey(t, m, tea.KeyMsg{Type: tea.KeyUp, Alt: true})
	if m.detailScroll != offset-1 {
		t.Fatal("cannot scroll back up")
	}
	m.moveSelection(0)
	if m.detailScroll != 0 {
		t.Fatal("new selection should reset detail scrolling")
	}
}
