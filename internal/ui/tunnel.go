package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	sshpkg "github.com/oddship/wg-tui/internal/ssh"
)

type tunnelState string

const (
	tunnelClosed  tunnelState = "closed"
	tunnelOpening tunnelState = "opening"
	tunnelOpen    tunnelState = "open"
	tunnelFailed  tunnelState = "failed"
)

type tunnelSession struct {
	target           string
	remotePort       int
	localPort        int
	state            tunnelState
	lastError        string
	process          *sshpkg.TunnelProcess
	activeGeneration int
	nextGeneration   int
}

type tunnelStartedMsg struct {
	generation int
	err        error
	stderr     string
}

type helpBindings struct {
	short []key.Binding
	full  [][]key.Binding
}

func (h helpBindings) ShortHelp() []key.Binding  { return h.short }
func (h helpBindings) FullHelp() [][]key.Binding { return h.full }

func (m *Model) initTunnelBindings() {
	m.tunnelBack = key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "back"))
	m.tunnelReconnect = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reconnect"))
	m.tunnelClose = key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "close tunnel"))
}

func (m *Model) startTunnelFormForSelection() {
	if len(m.filtered) == 0 {
		m.status = "no target selected"
		return
	}
	m.startTunnelForm(m.filtered[m.selected].Name)
}

func (m Model) startRsyncFormForSelection() (tea.Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		m.status = "no target selected"
		return m, nil
	}
	m.blurSearch()
	m.startRsyncForm(m.filtered[m.selected].Name)
	return m, nil
}

func (m Model) updateTunnel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		_ = m.stopTunnel()
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.toggleHelp()
	case key.Matches(msg, m.keys.Copy):
		m.copyTunnelCommand()
	case key.Matches(msg, m.tunnelClose):
		if err := m.stopTunnel(); err != nil {
			m.tunnel.state = tunnelFailed
			m.tunnel.lastError = err.Error()
			m.status = err.Error()
			return m, nil
		}
		m.tunnel.state = tunnelClosed
		m.tunnel.lastError = ""
		m.status = "tunnel closed"
	case key.Matches(msg, m.tunnelReconnect):
		return m.reconnectTunnel()
	case key.Matches(msg, m.tunnelBack):
		if err := m.stopTunnel(); err != nil {
			m.tunnel.lastError = err.Error()
			m.status = err.Error()
			m.mode = modeBrowse
			return m, nil
		}
		m.mode = modeBrowse
		m.status = "returned to target list"
	}
	return m, nil
}

func (m Model) reconnectTunnel() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.tunnel.target) == "" {
		m.status = "no tunnel target selected"
		return m, nil
	}
	if err := m.stopTunnel(); err != nil {
		m.tunnel.lastError = err.Error()
		m.status = err.Error()
	}
	return m.startTunnel(m.tunnel.target, m.tunnel.remotePort, m.tunnel.localPort)
}

func (m Model) startTunnel(target string, remotePort, localPort int) (tea.Model, tea.Cmd) {
	if err := m.stopTunnel(); err != nil {
		m.tunnel.lastError = err.Error()
	}
	m.blurSearch()
	m.mode = modeTunnel
	m.tunnel.target = target
	m.tunnel.remotePort = remotePort
	m.tunnel.localPort = localPort
	m.tunnel.state = tunnelOpening
	m.tunnel.lastError = ""
	m.status = fmt.Sprintf("complete SSH approval to open tunnel for %s", target)

	proc, err := sshpkg.PrepareTunnel(m.cfg, target, remotePort, localPort)
	if err != nil {
		m.tunnel.state = tunnelFailed
		m.tunnel.lastError = err.Error()
		m.status = fmt.Sprintf("tunnel failed: %s", err)
		return m, nil
	}

	m.tunnel.process = proc
	m.tunnel.nextGeneration++
	m.tunnel.activeGeneration = m.tunnel.nextGeneration
	generation := m.tunnel.activeGeneration
	return m, proc.StartCmd(func(err error, stderr string) tea.Msg {
		return tunnelStartedMsg{generation: generation, err: err, stderr: stderr}
	})
}

func (m *Model) stopTunnel() error {
	if m.tunnel.process == nil {
		return nil
	}
	proc := m.tunnel.process
	m.tunnel.process = nil
	m.tunnel.activeGeneration = 0
	if m.tunnel.state != tunnelFailed {
		m.tunnel.state = tunnelClosed
	}
	return proc.Close()
}

func (m Model) handleTunnelStarted(msg tunnelStartedMsg) (tea.Model, tea.Cmd) {
	if msg.generation == 0 || msg.generation != m.tunnel.activeGeneration {
		return m, nil
	}

	if msg.err != nil {
		if m.tunnel.process != nil {
			m.tunnel.process.Cleanup()
		}
		m.tunnel.process = nil
		m.tunnel.activeGeneration = 0
		m.tunnel.state = tunnelFailed
		m.tunnel.lastError = firstNonEmpty(msg.stderr, msg.err.Error())
		m.status = fmt.Sprintf("tunnel failed: %s", m.tunnel.lastError)
		return m, nil
	}

	if m.tunnel.process != nil {
		if err := m.tunnel.process.Check(); err != nil {
			m.tunnel.process.Cleanup()
			m.tunnel.process = nil
			m.tunnel.activeGeneration = 0
			m.tunnel.state = tunnelFailed
			m.tunnel.lastError = err.Error()
			m.status = fmt.Sprintf("tunnel failed: %s", err)
			return m, nil
		}
	}

	m.tunnel.state = tunnelOpen
	m.tunnel.lastError = ""
	m.status = fmt.Sprintf("tunnel open: 127.0.0.1:%d -> %s:127.0.0.1:%d", m.tunnel.localPort, m.tunnel.target, m.tunnel.remotePort)
	return m, nil
}

func (m *Model) copyTunnelCommand() {
	if strings.TrimSpace(m.tunnel.target) == "" {
		m.status = "no tunnel target selected"
		return
	}
	bin, args := sshpkg.TunnelCommand(m.cfg, m.tunnel.target, m.tunnel.remotePort, m.tunnel.localPort)
	if err := clipboard.WriteAll(bin + " " + strings.Join(args, " ")); err != nil {
		m.status = err.Error()
		return
	}
	m.status = "tunnel command copied"
}

func (m Model) viewTunnel() string {
	if m.width == 0 {
		m.width = 100
	}

	body := panelStyle.Width(max(44, m.width-2)).Render(m.tunnelDetailView(max(40, m.width-6)))

	return strings.Join([]string{
		m.viewTopChrome("wgt tunnel", fallback(m.tunnel.target, "No target selected"), m.tunnelHelpBindings().short),
		body,
		statusBarStyle.Width(max(0, m.width-2)).Render(m.status),
	}, "\n")
}

func (m Model) tunnelDetailView(width int) string {
	rows := []string{headerStyle.Render("Tunnel")}
	rows = append(rows,
		titleStyle.Render(fallback(m.tunnel.target, "No target selected")),
		fmt.Sprintf("state: %s", m.tunnel.state),
		fmt.Sprintf("local: 127.0.0.1:%d", m.tunnel.localPort),
		fmt.Sprintf("remote: 127.0.0.1:%d", m.tunnel.remotePort),
		fmt.Sprintf("warpgate ssh: %s@%s:%d", fallback(m.cfg.SSH.Username, "-"), fallback(m.cfg.SSH.Host, "-"), m.cfg.SSH.Port),
	)
	if m.tunnel.lastError != "" {
		rows = append(rows, "", lipgloss.NewStyle().Width(width).Render("last error: "+m.tunnel.lastError))
	}
	rows = append(rows, "", mutedStyle.Render("Tunnel startup may briefly hand control to ssh for approval, then return here once the tunnel is backgrounded."))
	return strings.Join(rows, "\n")
}

func (m Model) tunnelHelpBindings() helpBindings {
	copyBinding := cloneBinding(m.keys.Copy, "copy tunnel")
	helpBinding := cloneBinding(m.keys.Help, "help")
	quitBinding := cloneBinding(m.keys.Quit, "quit")
	return helpBindings{
		short: []key.Binding{m.tunnelClose, m.tunnelReconnect, copyBinding, m.tunnelBack, quitBinding},
		full:  [][]key.Binding{{m.tunnelClose, m.tunnelReconnect, copyBinding}, {m.tunnelBack, helpBinding, quitBinding}},
	}
}

func (m Model) browseHelpBindings() helpBindings {
	connectBinding := cloneBinding(m.keys.Connect, "connect")
	tunnelBinding := cloneBinding(m.keys.Tunnel, "tunnel")
	rsyncBinding := cloneBinding(m.keys.Rsync, "rsync")
	copyBinding := cloneBinding(m.keys.Copy, "copy ssh")
	short := []key.Binding{m.keys.Search, connectBinding, tunnelBinding, rsyncBinding, m.keys.EditConfig, m.keys.Refresh, m.keys.Quit}
	full := [][]key.Binding{{m.keys.Up, m.keys.Down, m.keys.Search, m.keys.Clear, connectBinding, tunnelBinding, rsyncBinding}, {m.keys.Refresh, m.keys.EditConfig, copyBinding, m.keys.Help, m.keys.Quit}}
	return helpBindings{short: short, full: full}
}

func cloneBinding(binding key.Binding, desc string) key.Binding {
	help := binding.Help()
	return key.NewBinding(key.WithKeys(binding.Keys()...), key.WithHelp(help.Key, desc))
}
