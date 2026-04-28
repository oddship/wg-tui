package ui

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func (m *Model) startForm(edit bool) {
	cfg := m.cfg
	if !edit {
		cfg = cfgpkg.Default()
	}

	m.formTitle = formTitle(edit)
	m.fields = []formField{
		newField("server.url", "Server URL", cfg.Server.URL, "Base URL, for example https://warpgate.example.com"),
		newPasswordField("server.token", "API Token", cfg.Server.Token, "Generate this once from the Warpgate web UI"),
		newField("ssh.username", "SSH Username", cfg.SSH.Username, "Warpgate username or email used for SSH"),
		newField("ssh.host", "SSH Host", cfg.SSH.Host, "Optional override. Leave blank to derive from server URL"),
		newField("ssh.port", "SSH Port", portString(cfg.SSH.Port), "Optional override. Usually auto-detected from /info"),
	}
	m.formIndex = 0
	m.formStatus = ""
	m.focusField(0)
}

func formTitle(edit bool) string {
	if edit {
		return "Edit Config"
	}
	return "Onboarding"
}

func portString(port int) string {
	if port <= 0 {
		return strconv.Itoa(cfgpkg.Default().SSH.Port)
	}
	return strconv.Itoa(port)
}

func newField(key, label, value, hint string) formField {
	return formField{key: key, label: label, input: newTextInput(label, value), hint: hint}
}

func newPasswordField(key, label, value, hint string) formField {
	ti := newTextInput(label, value)
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	return formField{key: key, label: label, input: ti, hint: hint}
}

func newTextInput(label, value string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = label + ": "
	ti.SetValue(value)
	ti.Width = 64
	return ti
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.mode == modeConfig {
			m.mode = modeBrowse
			m.formStatus = ""
			m.status = "config edit cancelled"
			return m, nil
		}
	case "tab", "down":
		m.focusField((m.formIndex + 1) % len(m.fields))
		return m, nil
	case "shift+tab", "up":
		m.focusField((m.formIndex - 1 + len(m.fields)) % len(m.fields))
		return m, nil
	case "enter":
		cfg, err := m.formConfig()
		if err != nil {
			m.formStatus = err.Error()
			return m, nil
		}
		m.formStatus = "validating config and loading targets..."
		return m, validateAndSaveCmd(m.cfgPath, cfg)
	}

	var cmd tea.Cmd
	m.fields[m.formIndex].input, cmd = m.fields[m.formIndex].input.Update(msg)
	return m, cmd
}

func (m *Model) focusField(idx int) {
	for i := range m.fields {
		if i == idx {
			m.fields[i].input.Focus()
		} else {
			m.fields[i].input.Blur()
		}
	}
	m.formIndex = idx
}

func (m Model) formConfig() (cfgpkg.Config, error) {
	cfg := m.cfg
	if cfg.Server.URL == "" && m.mode == modeOnboarding {
		cfg = cfgpkg.Default()
	}

	for _, f := range m.fields {
		value := strings.TrimSpace(f.input.Value())
		switch f.key {
		case "server.url":
			cfg.Server.URL = value
		case "server.token":
			cfg.Server.Token = value
		case "ssh.username":
			cfg.SSH.Username = value
		case "ssh.host":
			cfg.SSH.Host = value
		case "ssh.port":
			port, err := parsePort(value)
			if err != nil {
				return cfg, err
			}
			cfg.SSH.Port = port
		}
	}

	if err := validateRequired(cfg); err != nil {
		return cfg, err
	}
	if err := deriveSSHHost(&cfg); err != nil {
		return cfg, err
	}
	if cfg.SSH.Binary == "" {
		cfg.SSH.Binary = cfgpkg.Default().SSH.Binary
	}
	return cfg, nil
}

func parsePort(v string) (int, error) {
	if v == "" {
		return cfgpkg.Default().SSH.Port, nil
	}
	p, err := strconv.Atoi(v)
	if err != nil || p <= 0 {
		return 0, fmt.Errorf("invalid ssh port")
	}
	return p, nil
}

func validateRequired(cfg cfgpkg.Config) error {
	if cfg.Server.URL == "" || cfg.Server.Token == "" || cfg.SSH.Username == "" {
		return fmt.Errorf("server url, token, and username are required")
	}
	return nil
}

func deriveSSHHost(cfg *cfgpkg.Config) error {
	u, err := url.Parse(cfg.Server.URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid server url")
	}
	if cfg.SSH.Host == "" {
		cfg.SSH.Host = u.Hostname()
	}
	return nil
}

func (m Model) viewForm() string {
	lines := []string{
		titleStyle.Render(m.formTitle),
		mutedStyle.Render("tab/shift+tab to move - enter to validate and save - esc to cancel config edit"),
		"",
	}
	for i, f := range m.fields {
		label := f.label
		if i == m.formIndex {
			label = selectedStyle.Render(label)
		}
		lines = append(lines, label, f.input.View(), mutedStyle.Render(f.hint), "")
	}
	if m.formStatus != "" {
		lines = append(lines, statusBarStyle.Render(m.formStatus))
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(lines, "\n"))
}
