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

func (m *Model) startTunnelForm(target string) {
	m.pendingTunnelTarget = target
	m.mode = modeTunnelForm
	m.formTitle = fmt.Sprintf("Open Tunnel - %s", target)

	remote := ""
	local := ""
	if m.tunnelLastRemotePort > 0 {
		remote = strconv.Itoa(m.tunnelLastRemotePort)
	}
	if m.tunnelLastLocalPort > 0 {
		local = strconv.Itoa(m.tunnelLastLocalPort)
	}
	if remote != "" && local == "" {
		local = remote
	}

	m.fields = []formField{
		newField("tunnel.remote_port", "Remote Port", remote, "Service port on the selected target"),
		newField("tunnel.local_port", "Local Port", local, "Local port to bind on 127.0.0.1"),
	}
	m.formIndex = 0
	m.formStatus = "enter ports and press enter to start tunnel"
	m.tunnelLocalTouched = strings.TrimSpace(local) != "" && local != remote
	if !m.tunnelLocalTouched {
		m.tunnelAutoLocalValue = local
	} else {
		m.tunnelAutoLocalValue = ""
	}
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

func (m Model) updateConfigForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m Model) updateTunnelForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.formStatus = ""
		m.status = "tunnel cancelled"
		return m, nil
	case "tab", "down":
		m.focusField((m.formIndex + 1) % len(m.fields))
		return m, nil
	case "shift+tab", "up":
		m.focusField((m.formIndex - 1 + len(m.fields)) % len(m.fields))
		return m, nil
	case "enter":
		remotePort, localPort, err := m.tunnelFormPorts()
		if err != nil {
			m.formStatus = err.Error()
			return m, nil
		}
		m.tunnelLastRemotePort = remotePort
		m.tunnelLastLocalPort = localPort
		m.formStatus = ""
		return m.startTunnel(m.pendingTunnelTarget, remotePort, localPort)
	}

	if len(m.fields) == 0 {
		return m, nil
	}

	before := m.fields[m.formIndex].input.Value()
	var cmd tea.Cmd
	m.fields[m.formIndex].input, cmd = m.fields[m.formIndex].input.Update(msg)
	after := m.fields[m.formIndex].input.Value()
	m.syncTunnelFormFields(m.formIndex, before, after)
	return m, cmd
}

func (m *Model) syncTunnelFormFields(index int, before, after string) {
	if len(m.fields) < 2 {
		return
	}

	switch m.fields[index].key {
	case "tunnel.remote_port":
		localValue := m.fields[1].input.Value()
		if !m.tunnelLocalTouched || strings.TrimSpace(localValue) == "" || localValue == m.tunnelAutoLocalValue {
			m.fields[1].input.SetValue(after)
			m.tunnelAutoLocalValue = after
			m.tunnelLocalTouched = false
		}
	case "tunnel.local_port":
		if after != m.tunnelAutoLocalValue {
			m.tunnelLocalTouched = true
		}
		if strings.TrimSpace(after) == "" && strings.TrimSpace(before) != "" {
			m.tunnelLocalTouched = false
		}
	}
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

func (m Model) tunnelFormPorts() (int, int, error) {
	if strings.TrimSpace(m.pendingTunnelTarget) == "" {
		return 0, 0, fmt.Errorf("no target selected")
	}

	values := make(map[string]string, len(m.fields))
	for _, f := range m.fields {
		values[f.key] = strings.TrimSpace(f.input.Value())
	}

	remotePort, err := parseTunnelPort(values["tunnel.remote_port"], "remote")
	if err != nil {
		return 0, 0, err
	}
	localPort, err := parseTunnelPort(values["tunnel.local_port"], "local")
	if err != nil {
		return 0, 0, err
	}
	return remotePort, localPort, nil
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

func parseTunnelPort(v, label string) (int, error) {
	p, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || p <= 0 || p > 65535 {
		return 0, fmt.Errorf("invalid %s port", label)
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
	helpText := "tab/shift+tab to move - enter to submit"
	if m.mode == modeConfig {
		helpText = "tab/shift+tab to move - enter to validate and save - esc to cancel config edit"
	}
	if m.mode == modeOnboarding {
		helpText = "tab/shift+tab to move - enter to validate and save"
	}
	if m.mode == modeTunnelForm {
		helpText = "tab/shift+tab to move - enter to start tunnel - esc to cancel"
	}

	lines := []string{
		m.viewTopChrome("wgt", m.formTitle, nil),
		mutedStyle.Render(helpText),
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
