package ui

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
	sshpkg "github.com/oddship/wg-tui/internal/ssh"
)

func (m *Model) startForm(edit bool) {
	cfg := m.cfg
	if !edit {
		cfg = cfgpkg.Default()
	}

	m.formTitle = formTitle(edit)
	m.fields = []formField{
		newField("server.url", "Server URL", cfg.Server.URL, "Base URL, for example https://warpgate.example.com"),
		newPasswordField("server.token", "API Token", cfg.Server.Token, apiTokenHint(cfg.Server.URL)),
		newField("ssh.username", "SSH Username", cfg.SSH.Username, "Warpgate username or email used for SSH"),
		newField("ssh.host", "SSH Host", cfg.SSH.Host, "Optional override. Leave blank to derive from server URL"),
		newField("ssh.port", "SSH Port", portString(cfg.SSH.Port), "Optional override. Usually auto-detected from /info"),
	}
	m.formIndex = 0
	m.formStatus = ""
	m.focusField(0)
}

type rsyncFinishedMsg struct {
	tool       string
	direction  string
	localPath  string
	remotePath string
	err        error
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

func (m *Model) startRsyncForm(target string) {
	m.pendingRsyncTarget = target
	m.mode = modeRsyncForm
	m.formTitle = fmt.Sprintf("Transfer - %s", target)

	tool := m.rsyncLastTool
	if strings.TrimSpace(tool) == "" {
		tool = "rsync"
	}
	direction := m.rsyncLastDirection
	if strings.TrimSpace(direction) == "" {
		direction = "upload"
	}

	m.fields = []formField{
		newSelectField("rsync.tool", "Tool", []string{"rsync", "scp"}, tool, "Use left/right to choose rsync or scp"),
		newSelectField("rsync.direction", "Direction", []string{"upload", "download"}, direction, "Use left/right to choose transfer direction"),
		newField("rsync.flags", "Flags", m.transferFlagsForTool(tool), "Editable transfer flags for the selected tool"),
		newField("rsync.local_path", "Local Path", m.rsyncLastLocalPath, ""),
		newField("rsync.remote_path", "Remote Path", m.rsyncLastRemotePath, ""),
	}
	m.syncRsyncHints()
	m.formIndex = 0
	m.formStatus = "choose tool and direction, set paths, then press enter to run transfer"
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
	return formField{key: key, label: label, input: newTextInput(value), hint: hint}
}

func newSelectField(key, label string, options []string, value, hint string) formField {
	index := 0
	for i, option := range options {
		if strings.EqualFold(strings.TrimSpace(option), strings.TrimSpace(value)) {
			index = i
			break
		}
	}
	ti := newTextInput(options[index])
	ti.Blur()
	return formField{key: key, label: label, input: ti, hint: hint, options: options, optionIndex: index}
}

func newPasswordField(key, label, value, hint string) formField {
	ti := newTextInput(value)
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	return formField{key: key, label: label, input: ti, hint: hint}
}

func newTextInput(value string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
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
	m.syncConfigHints()
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

func (m Model) updateRsyncForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	current := m.currentField()
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.formStatus = ""
		m.status = "transfer cancelled"
		return m, nil
	case "tab", "down":
		m.focusField((m.formIndex + 1) % len(m.fields))
		return m, nil
	case "shift+tab", "up":
		m.focusField((m.formIndex - 1 + len(m.fields)) % len(m.fields))
		return m, nil
	case "left", "h":
		if current.isSelect() {
			m.shiftFieldOption(m.formIndex, -1)
			return m, nil
		}
	case "right", "l":
		if current.isSelect() {
			m.shiftFieldOption(m.formIndex, 1)
			return m, nil
		}
	case "enter":
		if current.isSelect() {
			m.focusField((m.formIndex + 1) % len(m.fields))
			return m, nil
		}
		tool, direction, flags, localPath, remotePath, err := m.rsyncFormValues()
		if err != nil {
			m.formStatus = err.Error()
			return m, nil
		}
		m.rsyncLastTool = tool
		m.rsyncLastDirection = direction
		m.rememberTransferFlags(tool, strings.Join(flags, " "))
		m.rsyncLastLocalPath = localPath
		m.rsyncLastRemotePath = remotePath
		m.persistState()
		m.mode = modeBrowse
		m.formStatus = ""
		m.status = fmt.Sprintf("running %s %s for %s", tool, direction, m.pendingRsyncTarget)
		return m, rsyncCmd(m.cfg, m.pendingRsyncTarget, tool, direction, flags, localPath, remotePath)
	}

	var cmd tea.Cmd
	m.fields[m.formIndex].input, cmd = m.fields[m.formIndex].input.Update(msg)
	m.rememberTransferDraft(m.formIndex)
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
		if i == idx && !m.fields[i].isSelect() {
			m.fields[i].input.Focus()
		} else {
			m.fields[i].input.Blur()
		}
	}
	m.formIndex = idx
}

func (m *Model) currentField() formField {
	if m.formIndex < 0 || m.formIndex >= len(m.fields) {
		return formField{}
	}
	return m.fields[m.formIndex]
}

func (m *Model) shiftFieldOption(index, delta int) {
	if index < 0 || index >= len(m.fields) {
		return
	}
	field := &m.fields[index]
	if !field.isSelect() || len(field.options) == 0 {
		return
	}
	field.optionIndex = (field.optionIndex + delta + len(field.options)) % len(field.options)
	field.input.SetValue(field.options[field.optionIndex])
	if field.key == "rsync.direction" {
		m.rsyncLastDirection = strings.TrimSpace(field.input.Value())
		m.syncRsyncHints()
	}
	if field.key == "rsync.tool" {
		m.rsyncLastTool = strings.TrimSpace(field.input.Value())
		m.syncTransferFlags()
	}
	m.persistState()
}

func (m *Model) syncTransferFlags() {
	m.setFieldValue("rsync.flags", m.transferFlagsForTool(m.fieldValue("rsync.tool")))
}

func (m *Model) rememberTransferDraft(index int) {
	if index < 0 || index >= len(m.fields) {
		return
	}
	field := m.fields[index]
	value := strings.TrimSpace(field.input.Value())
	switch field.key {
	case "rsync.local_path":
		m.rsyncLastLocalPath = value
	case "rsync.remote_path":
		m.rsyncLastRemotePath = value
	case "rsync.flags":
		m.rememberTransferFlags(m.fieldValue("rsync.tool"), value)
	default:
		return
	}
	m.persistState()
}

func (m *Model) transferFlagsForTool(tool string) string {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "scp":
		if strings.TrimSpace(m.rsyncLastScpFlags) != "" {
			return m.rsyncLastScpFlags
		}
		return defaultTransferFlags("scp")
	case "rsync":
		fallthrough
	default:
		if strings.TrimSpace(m.rsyncLastRsyncFlags) != "" {
			return m.rsyncLastRsyncFlags
		}
		return defaultTransferFlags("rsync")
	}
}

func (m *Model) rememberTransferFlags(tool, flags string) {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "scp":
		m.rsyncLastScpFlags = flags
	default:
		m.rsyncLastRsyncFlags = flags
	}
}

func (m *Model) syncConfigHints() {
	m.setFieldHint("server.token", apiTokenHint(m.rawFieldValue("server.url")))
}

func apiTokenHint(serverURL string) string {
	const fallback = "Generate this once from the Warpgate web UI"
	serverURL = strings.TrimSpace(serverURL)
	if serverURL == "" {
		return fallback
	}
	base, err := url.Parse(serverURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return fallback
	}
	link := base.ResolveReference(&url.URL{Path: "/@warpgate/", Fragment: "/profile/api-tokens"})
	return link.String()
}

func (m *Model) syncRsyncHints() {
	direction := m.fieldValue("rsync.direction")
	localHint := "Source path on this machine"
	remoteHint := "Destination path on the target"
	if direction == "download" {
		localHint = "Destination path on this machine"
		remoteHint = "Source path on the target"
	}
	m.setFieldHint("rsync.local_path", localHint)
	m.setFieldHint("rsync.remote_path", remoteHint)
}

func (m *Model) fieldValue(key string) string {
	return strings.ToLower(m.rawFieldValue(key))
}

func (m *Model) rawFieldValue(key string) string {
	for _, field := range m.fields {
		if field.key == key {
			return strings.TrimSpace(field.input.Value())
		}
	}
	return ""
}

func (m *Model) setFieldHint(key, hint string) {
	for i := range m.fields {
		if m.fields[i].key == key {
			m.fields[i].hint = hint
			return
		}
	}
}

func (m *Model) setFieldValue(key, value string) {
	for i := range m.fields {
		if m.fields[i].key == key {
			m.fields[i].input.SetValue(value)
			return
		}
	}
}

func (f formField) isSelect() bool {
	return len(f.options) > 0
}

func (f formField) selectView(active bool) string {
	value := f.input.Value()
	if strings.TrimSpace(value) == "" && len(f.options) > 0 {
		value = f.options[f.optionIndex]
	}
	text := fmt.Sprintf("< %s >", value)
	if active {
		return selectedStyle.Render(text)
	}
	return text
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

func (m Model) rsyncFormValues() (string, string, []string, string, string, error) {
	if strings.TrimSpace(m.pendingRsyncTarget) == "" {
		return "", "", nil, "", "", fmt.Errorf("no target selected")
	}

	values := make(map[string]string, len(m.fields))
	for _, f := range m.fields {
		values[f.key] = strings.TrimSpace(f.input.Value())
	}

	tool, err := parseTransferTool(values["rsync.tool"])
	if err != nil {
		return "", "", nil, "", "", err
	}
	direction, err := parseRsyncDirection(values["rsync.direction"])
	if err != nil {
		return "", "", nil, "", "", err
	}
	flags := parseTransferFlags(values["rsync.flags"])
	localPath := values["rsync.local_path"]
	if localPath == "" {
		return "", "", nil, "", "", fmt.Errorf("local path is required")
	}
	remotePath := values["rsync.remote_path"]
	if remotePath == "" {
		return "", "", nil, "", "", fmt.Errorf("remote path is required")
	}
	return tool, direction, flags, localPath, remotePath, nil
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

func parseTransferTool(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "rsync", "sync":
		return "rsync", nil
	case "scp", "copy":
		return "scp", nil
	default:
		return "", fmt.Errorf("invalid transfer tool - use rsync or scp")
	}
}

func parseRsyncDirection(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "upload", "up", "push", "host-to-remote", "local-to-remote":
		return "upload", nil
	case "download", "down", "pull", "remote-to-host", "remote-to-local":
		return "download", nil
	default:
		return "", fmt.Errorf("invalid rsync direction - use upload or download")
	}
}

func parseTransferFlags(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.Fields(v)
}

func defaultTransferFlags(tool string) string {
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "scp":
		return "-C -p"
	default:
		return "-avz"
	}
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

func rsyncCmd(cfg cfgpkg.Config, target, tool, direction string, flags []string, localPath, remotePath string) tea.Cmd {
	return func() tea.Msg {
		var (
			bin  string
			args []string
			err  error
		)
		switch tool {
		case "rsync":
			bin, args, err = sshpkg.RsyncCommand(cfg, target, direction, flags, localPath, remotePath)
		case "scp":
			bin, args, err = sshpkg.ScpCommand(cfg, target, direction, flags, localPath, remotePath)
		default:
			err = fmt.Errorf("invalid transfer tool: %s", tool)
		}
		if err != nil {
			return rsyncFinishedMsg{tool: tool, direction: direction, localPath: localPath, remotePath: remotePath, err: err}
		}
		cmd := exec.Command(bin, args...)
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return rsyncFinishedMsg{tool: tool, direction: direction, localPath: localPath, remotePath: remotePath, err: err}
		})()
	}
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
	if m.mode == modeRsyncForm {
		helpText = "tab/shift+tab to move - left/right to choose options - enter to continue or run transfer - esc to cancel"
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
		fieldView := f.input.View()
		if f.isSelect() {
			fieldView = f.selectView(i == m.formIndex)
		}
		lines = append(lines, label, fieldView, mutedStyle.Render(f.hint), "")
	}
	if m.formStatus != "" {
		lines = append(lines, statusBarStyle.Render(m.formStatus))
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(lines, "\n"))
}
