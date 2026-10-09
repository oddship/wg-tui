package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

type profileLoadedMsg struct {
	profile  string
	snapshot cache.Snapshot
	state    cache.State
	status   string
}

func (m *Model) openProfilePicker() {
	m.blurSearch()
	m.mode = modeProfiles
	m.profileIndex = max(0, slices.Index(m.cfg.ProfileNames(), m.cfg.Profile))
}

func (m Model) updateProfilePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := m.cfg.ProfileNames()
	switch {
	case msg.String() == "esc":
		m.mode = modeBrowse
		m.status = "profile switch cancelled"
	case key.Matches(msg, m.keys.Up):
		m.profileIndex = clamp(m.profileIndex-1, 0, len(options)-1)
	case key.Matches(msg, m.keys.Down):
		m.profileIndex = clamp(m.profileIndex+1, 0, len(options)-1)
	case msg.String() == "n":
		m.startNewProfileForm()
	case msg.String() == "enter" && len(options) > 0:
		return m.switchProfile(options[m.profileIndex])
	}
	return m, nil
}

func (m Model) switchProfile(name string) (tea.Model, tea.Cmd) {
	m.mode = modeBrowse
	if name == m.cfg.Profile {
		m.status = fmt.Sprintf("already on profile %s", name)
		return m, nil
	}
	cfg := m.cfg
	if err := cfg.Use(name); err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.applyConfig(cfg)
	m.persistLastProfile()
	m.selected = 0
	m.status = fmt.Sprintf("switching to profile %s...", name)
	cacheDir := m.cacheDir()
	return m, func() tea.Msg {
		snap, state, status := loadCache(cacheDir)
		return profileLoadedMsg{profile: name, snapshot: snap, state: state, status: status}
	}
}

func (m Model) handleProfileLoaded(msg profileLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.profile != m.cfg.Profile {
		return m, nil
	}
	return m.showCached(msg.snapshot, msg.state, joinStatus("profile "+msg.profile+":", msg.status))
}

func (m *Model) startNewProfileForm() {
	cfg := m.cfg
	cfg.Profile = ""
	cfg.Server = cfgpkg.ServerConfig{}
	cfg.SSH = cfgpkg.Default().SSH
	m.mode = modeConfig
	m.startFormFrom(cfg, "New Profile", "")
	m.formStatus = "fill in the new profile and press enter to save"
}

func (m Model) viewProfiles() string {
	lines := []string{
		m.viewTopChrome("wgt", "Switch profile", nil),
		mutedStyle.Render("up/down to move - enter to select - n for new profile - esc to cancel"),
		"",
	}
	for i, name := range m.cfg.ProfileNames() {
		label := name
		if name == m.cfg.Profile {
			label += " (active)"
		}
		row := "  " + label
		if i == m.profileIndex {
			row = selectedStyle.Render("• " + label)
		}
		if p, ok := m.cfg.Profiles[name]; ok && strings.TrimSpace(p.Server.URL) != "" {
			row += " " + mutedStyle.Render(p.Server.URL)
		}
		lines = append(lines, row)
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(lines, "\n"))
}
