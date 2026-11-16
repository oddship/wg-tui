package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestProfilePickerActionsOverrideConflictingNavigation(t *testing.T) {
	for _, navigation := range []string{"up", "down"} {
		for _, action := range []string{"n", "enter", "esc"} {
			t.Run(navigation+"="+action, func(t *testing.T) {
				m := multiProfileModel(t)
				m.cacheDirOverride = t.TempDir()
				// Release any previous binding for this key so the global config stays valid.
				if action == "enter" {
					m.cfg.Keys.Connect = []string{"ctrl+o"}
				}
				if action == "esc" {
					m.cfg.Keys.Clear = []string{"ctrl+e"}
				}
				if navigation == "up" {
					m.cfg.Keys.Up = []string{action}
				} else {
					m.cfg.Keys.Down = []string{action}
				}
				m.keys = cfgpkg.NewKeyMap(m.cfg.Keys)
				m.openProfilePicker()
				var key tea.KeyMsg
				switch action {
				case "n":
					key = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
				case "enter":
					key = tea.KeyMsg{Type: tea.KeyEnter}
				case "esc":
					key = tea.KeyMsg{Type: tea.KeyEsc}
				}
				updated, _ := m.Update(key)
				m = updated.(Model)
				if action == "n" && (m.mode != modeConfig || m.formProfile != "") {
					t.Fatal("new profile action was hidden by navigation")
				}
				if action != "n" && m.mode != modeBrowse {
					t.Fatal("select or cancel action was hidden by navigation")
				}
			})
		}
	}
}

func TestProfilePickerNativeArrowsAlwaysNavigate(t *testing.T) {
	m := multiProfileModel(t)
	m.cfg.Keys.Up = []string{"n"}
	m.cfg.Keys.Down = []string{"enter"}
	m.keys = cfgpkg.NewKeyMap(m.cfg.Keys)
	m.openProfilePicker()
	if !strings.Contains(m.viewProfiles(), "up/down to move") {
		t.Fatal("missing advertised arrow navigation")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.profileIndex != 1 {
		t.Fatal("native down arrow was disabled")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.profileIndex != 0 {
		t.Fatal("native up arrow was disabled")
	}
}

func TestProfilePickerKeepsNonConflictingNavigationAliases(t *testing.T) {
	m := multiProfileModel(t)
	m.openProfilePicker()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(Model)
	if m.profileIndex != 1 {
		t.Fatal("configured down alias stopped working")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(Model)
	if m.profileIndex != 0 {
		t.Fatal("configured up alias stopped working")
	}
}
