package ui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oddship/wg-tui/internal/api"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestOnboardingHeaderShowsCurrentPageOnNarrowScreens(t *testing.T) {
	m := New("test")
	m.mode = modeOnboarding
	m.startForm(false)
	m.setWindowSize(tea.WindowSizeMsg{Width: 60, Height: 20})

	line := visibleLinePrefix(m.View(), 60)
	if !strings.Contains(line, "onboarding") {
		t.Fatalf("expected onboarding page label in visible header, got %q", line)
	}
}

func TestConfigHeaderShowsCurrentPageOnNarrowScreens(t *testing.T) {
	m := New("test")
	m.mode = modeConfig
	m.startForm(true)
	m.setWindowSize(tea.WindowSizeMsg{Width: 60, Height: 20})

	line := visibleLinePrefix(m.View(), 60)
	if !strings.Contains(line, "config") {
		t.Fatalf("expected config page label in visible header, got %q", line)
	}
}

func TestSearchHeaderShowsCurrentPageOnVeryNarrowScreens(t *testing.T) {
	m := New("test")
	m.mode = modeBrowse
	m.searchFocused = true
	m.searchBox.Focus()
	m.setWindowSize(tea.WindowSizeMsg{Width: 24, Height: 20})

	line := visibleLinePrefix(m.View(), 24)
	if !strings.Contains(line, "search") {
		t.Fatalf("expected search page label in visible header, got %q", line)
	}
}

func TestFocusedSearchDoesNotRenderDuplicateSearchHeading(t *testing.T) {
	m := New("test")
	m.mode = modeBrowse
	m.searchFocused = true
	m.searchBox.Focus()
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 20})

	view := ansiPattern.ReplaceAllString(m.View(), "")
	if countExactLines(view, "Search") != 1 {
		t.Fatalf("expected one Search heading while typing, got view %q", view)
	}
}

func TestSearchPlaceholderSwitchesToEscapeHintWhenFocused(t *testing.T) {
	m := New("test")
	if m.searchBox.Placeholder != searchPlaceholderIdle {
		t.Fatalf("expected idle placeholder %q, got %q", searchPlaceholderIdle, m.searchBox.Placeholder)
	}

	m.focusSearch("")
	if m.searchBox.Placeholder != searchPlaceholderFocused {
		t.Fatalf("expected focused placeholder %q, got %q", searchPlaceholderFocused, m.searchBox.Placeholder)
	}

	m.blurSearch()
	if m.searchBox.Placeholder != searchPlaceholderIdle {
		t.Fatalf("expected placeholder to reset to %q, got %q", searchPlaceholderIdle, m.searchBox.Placeholder)
	}
}

func TestBrowseViewUsesPageTitleInsteadOfSelectionSummary(t *testing.T) {
	m := New("test")
	m.mode = modeBrowse
	m.setTargets([]api.Target{{Name: "one", Kind: "ssh"}})
	m.selected = 0
	m.status = "ready"
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 20})

	view := ansiPattern.ReplaceAllString(m.View(), "")
	if !strings.Contains(view, "Browse targets") {
		t.Fatalf("expected browse page title, got %q", view)
	}
	if strings.Contains(view, "selected=") {
		t.Fatalf("expected no selection summary in browse header, got %q", view)
	}
}

func TestBrowseHeaderDoesNotPretendAllPagesAreReachable(t *testing.T) {
	m := New("test")
	m.mode = modeBrowse
	m.setWindowSize(tea.WindowSizeMsg{Width: 100, Height: 20})

	line := visibleLinePrefix(m.View(), 100)
	if strings.Contains(line, "onboarding") || strings.Contains(line, "tunnel form") || strings.Contains(line, "rsync form") {
		t.Fatalf("expected browse header to show current page only, got %q", line)
	}
	if !strings.Contains(line, "browse") {
		t.Fatalf("expected browse page label in header, got %q", line)
	}
}

func TestBrowseViewShowsConfigShortcutEvenWhenShortcutsWrap(t *testing.T) {
	m := New("test")
	m.applyConfig(cfgpkg.Default())
	m.mode = modeBrowse
	m.setWindowSize(tea.WindowSizeMsg{Width: 60, Height: 20})

	view := ansiPattern.ReplaceAllString(m.View(), "")
	if !strings.Contains(view, "e config") {
		t.Fatalf("expected browse shortcuts to advertise config action, got %q", view)
	}
}

func TestRenderTargetListLineShowsDescriptionAsSecondaryText(t *testing.T) {
	target := makeRenderTarget("host-01", "alpha", "zone-a1")

	line := ansiPattern.ReplaceAllString(renderTargetListLine(target, 80, false), "")
	if !strings.Contains(line, "host-01 [Ssh] alpha zone-a1") {
		t.Fatalf("expected list row to include description, got %q", line)
	}
}

func TestRenderTargetListLineTrimsDescriptionToAvailableWidth(t *testing.T) {
	target := makeRenderTarget("host-01", "alpha", "zone-a1-extra-long")

	width := 28
	line := ansiPattern.ReplaceAllString(renderTargetListLine(target, width, false), "")
	if len([]rune(line)) > width {
		t.Fatalf("expected rendered list row to fit width %d, got %q", width, line)
	}
}

func TestRenderTargetListLineShowsRecentIndicator(t *testing.T) {
	target := makeRenderTarget("host-01", "alpha", "")

	line := ansiPattern.ReplaceAllString(renderTargetListLineWithRecent(target, 80, false, true), "")
	if !strings.Contains(line, recentTargetMarker) {
		t.Fatalf("expected recent indicator in list row, got %q", line)
	}
}

func TestBrowseListShowsRecentIndicatorForRecentTarget(t *testing.T) {
	m := New("test")
	m.mode = modeBrowse
	m.recentTargets = []string{"prod-api"}
	m.setTargets([]api.Target{
		makeRenderTarget("prod-api", "alpha", ""),
		makeRenderTarget("prod-db", "alpha", ""),
	})
	m.setWindowSize(tea.WindowSizeMsg{Width: 80, Height: 20})

	view := ansiPattern.ReplaceAllString(m.View(), "")
	if !strings.Contains(view, "prod-api [Ssh] alpha "+recentTargetMarker) {
		t.Fatalf("expected recent indicator in browse list, got %q", view)
	}
}

func makeRenderTarget(name, group, description string) api.Target {
	var target api.Target
	target.Name = name
	target.Kind = "Ssh"
	target.Group.Name = group
	target.Description = description
	return target
}

func visibleLinePrefix(view string, width int) string {
	for _, raw := range strings.Split(view, "\n") {
		line := ansiPattern.ReplaceAllString(raw, "")
		if strings.TrimSpace(line) == "" {
			continue
		}
		runes := []rune(line)
		if len(runes) > width {
			runes = runes[:width]
		}
		return string(runes)
	}
	return ""
}

func countExactLines(view, want string) int {
	count := 0
	for _, raw := range strings.Split(view, "\n") {
		line := strings.TrimSpace(ansiPattern.ReplaceAllString(raw, ""))
		if line == want {
			count++
		}
	}
	return count
}
