package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
	"github.com/oddship/wg-tui/internal/search"
	sshpkg "github.com/oddship/wg-tui/internal/ssh"
)

type mode int

const (
	modeBrowse mode = iota
	modeOnboarding
	modeConfig
	modeTunnelForm
	modeRsyncForm
	modeTunnel
)

type Model struct {
	cfgPath          string
	cacheDirOverride string
	cfg              cfgpkg.Config
	keys             cfgpkg.KeyMap
	client           *api.Client
	tunnel           tunnelSession

	searchBox     textinput.Model
	searchFocused bool
	help          help.Model
	mode          mode
	status        string
	showHelp      bool

	targets  []api.Target
	filtered []api.Target
	selected int
	index    search.Index
	snapshot cache.Snapshot

	formTitle            string
	formIndex            int
	fields               []formField
	formStatus           string
	pendingTunnelTarget  string
	pendingRsyncTarget   string
	rsyncLastTool        string
	rsyncLastDirection   string
	rsyncLastRsyncFlags  string
	rsyncLastScpFlags    string
	rsyncLastLocalPath   string
	rsyncLastRemotePath  string
	recentTargets        []string
	tunnelLastRemotePort int
	tunnelLastLocalPort  int
	tunnelLocalTouched   bool
	tunnelAutoLocalValue string
	width                int
	height               int
	tunnelBack           key.Binding
	tunnelReconnect      key.Binding
	tunnelClose          key.Binding
}

type formField struct {
	key         string
	label       string
	input       textinput.Model
	hint        string
	options     []string
	optionIndex int
}

type loadedMsg struct {
	cfg      cfgpkg.Config
	snapshot cache.Snapshot
	state    cache.State
	status   string
	err      error
}

type refreshMsg struct {
	snapshot cache.Snapshot
	status   string
	err      error
}

type configSavedMsg struct {
	cfg      cfgpkg.Config
	snapshot cache.Snapshot
	status   string
	err      error
}

const (
	searchPlaceholderIdle    = "Type / to search targets"
	searchPlaceholderFocused = "Type to search - esc to exit"
	recentTargetLimit        = 10
	recentTargetMarker       = "↺"
)

type Option func(*Model)

func WithCacheDirOverride(dir string) Option {
	return func(m *Model) {
		m.cacheDirOverride = strings.TrimSpace(dir)
	}
}

func New(cfgPath string, opts ...Option) Model {
	ti := textinput.New()
	ti.Placeholder = searchPlaceholderIdle
	ti.Prompt = "search> "
	ti.Blur()

	m := Model{
		cfgPath:   cfgPath,
		searchBox: ti,
		help:      help.New(),
		status:    "loading...",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&m)
		}
	}
	m.help.ShowAll = false
	m.initTunnelBindings()
	return m
}

func (m Model) Init() tea.Cmd { return loadCmd(m.cfgPath, m.cacheDirOverride) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setWindowSize(msg)
		return m, nil
	case loadedMsg:
		return m.handleLoaded(msg)
	case refreshMsg:
		return m.handleRefresh(msg)
	case configSavedMsg:
		return m.handleConfigSaved(msg)
	case sshpkg.ExecFinishedMsg:
		return m.handleSSHFinished(msg)
	case tunnelStartedMsg:
		return m.handleTunnelStarted(msg)
	case rsyncFinishedMsg:
		return m.handleRsyncFinished(msg)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.stopTunnel()
			return m, tea.Quit
		}
		switch m.mode {
		case modeBrowse:
			return m.updateBrowse(msg)
		case modeOnboarding, modeConfig:
			return m.updateConfigForm(msg)
		case modeTunnelForm:
			return m.updateTunnelForm(msg)
		case modeRsyncForm:
			return m.updateRsyncForm(msg)
		case modeTunnel:
			return m.updateTunnel(msg)
		default:
			return m, nil
		}
	}

	if m.mode == modeBrowse && m.searchFocused {
		return m.updateSearchInput(msg)
	}

	return m, nil
}

func (m *Model) setWindowSize(msg tea.WindowSizeMsg) {
	m.width, m.height = msg.Width, msg.Height
	m.searchBox.Width = max(20, msg.Width-14)
	m.help.Width = msg.Width
}

func (m Model) handleLoaded(msg loadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.cfg = cfgpkg.Default()
		m.keys = cfgpkg.NewKeyMap(m.cfg.Keys)
		m.mode = modeOnboarding
		m.startForm(false)
		m.formStatus = firstNonEmpty(msg.status, msg.err.Error())
		m.status = m.formStatus
		return m, nil
	}

	m.applyConfig(msg.cfg)
	m.applyState(msg.state)
	m.snapshot = msg.snapshot
	m.setTargets(msg.snapshot.Targets)
	m.status = msg.status

	if msg.snapshot.FetchedAt.IsZero() || cache.IsStale(msg.snapshot, m.cfg.Cache.TTL) {
		m.status = joinStatus(m.status, "refreshing in background...")
		return m, refreshCmd(m.client)
	}

	return m, nil
}

func (m Model) handleRefresh(msg refreshMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = msg.err.Error()
		return m, nil
	}

	m.applySnapshot(msg.snapshot)
	m.status = msg.status
	return m, nil
}

func (m Model) handleConfigSaved(msg configSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.formStatus = msg.err.Error()
		m.status = msg.err.Error()
		return m, nil
	}

	m.applyConfig(msg.cfg)
	m.applySnapshot(msg.snapshot)
	m.status = msg.status
	m.mode = modeBrowse
	m.formStatus = ""
	return m, nil
}

func (m Model) handleSSHFinished(msg sshpkg.ExecFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.ConfigUpdated {
		m.applyConfig(msg.Config)
	}
	if msg.Err != nil {
		m.status = msg.Err.Error()
	} else {
		m.status = firstNonEmpty(msg.Status, "ssh session ended")
	}
	return m, nil
}

func (m Model) handleRsyncFinished(msg rsyncFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = fmt.Sprintf("%s %s failed: %v", msg.tool, msg.direction, msg.err)
		return m, nil
	}
	m.status = fmt.Sprintf("%s %s complete", msg.tool, msg.direction)
	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchFocused {
		return m.updateBrowseWhileSearching(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.toggleHelp()
	case key.Matches(msg, m.keys.Search):
		m.focusSearch("search focused")
	case key.Matches(msg, m.keys.Clear):
		m.clearSearch()
	case key.Matches(msg, m.keys.Up):
		m.moveSelection(-1)
	case key.Matches(msg, m.keys.Down):
		m.moveSelection(1)
	case key.Matches(msg, m.keys.Refresh):
		m.status = "refreshing..."
		return m, refreshCmd(m.client)
	case key.Matches(msg, m.keys.EditConfig):
		m.openConfigEditor()
	case key.Matches(msg, m.keys.Connect):
		return m.connectSelected()
	case key.Matches(msg, m.keys.Tunnel):
		m.startTunnelFormForSelection()
	case key.Matches(msg, m.keys.Rsync):
		return m.startRsyncFormForSelection()
	case key.Matches(msg, m.keys.Copy):
		m.copySelectedCommand()
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.focusSearch("")
			return m.updateSearchInput(msg)
		}
	}

	return m, nil
}

func (m Model) updateBrowseWhileSearching(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyUp:
		m.moveSelection(-1)
		return m, nil
	case msg.Type == tea.KeyDown:
		m.moveSelection(1)
		return m, nil
	case key.Matches(msg, m.keys.Search):
		return m, nil
	case key.Matches(msg, m.keys.Clear):
		m.clearOrBlurSearch()
		return m, nil
	case key.Matches(msg, m.keys.Connect):
		return m.connectSelected()
	default:
		return m.updateSearchInput(msg)
	}
}

func (m Model) updateSearchInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.searchBox, cmd = m.searchBox.Update(msg)
	m.applyFilter()
	return m, cmd
}

func (m *Model) applyConfig(cfg cfgpkg.Config) {
	m.cfg = cfg
	m.keys = cfgpkg.NewKeyMap(cfg.Keys)
	m.client = api.New(cfg)
}

func (m Model) cacheDir() string {
	return resolveCacheDir(m.cfg.Cache.Dir, m.cacheDirOverride)
}

func resolveCacheDir(configDir, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return strings.TrimSpace(configDir)
}

func (m *Model) applyState(state cache.State) {
	m.rsyncLastTool = state.Transfer.Tool
	m.rsyncLastDirection = state.Transfer.Direction
	m.rsyncLastRsyncFlags = state.Transfer.RsyncFlags
	m.rsyncLastScpFlags = state.Transfer.ScpFlags
	m.rsyncLastLocalPath = state.Transfer.LocalPath
	m.rsyncLastRemotePath = state.Transfer.RemotePath
	m.recentTargets = append([]string(nil), state.RecentTargets...)
}

func (m Model) stateSnapshot() cache.State {
	return cache.State{
		Transfer: cache.TransferState{
			Tool:       m.rsyncLastTool,
			Direction:  m.rsyncLastDirection,
			RsyncFlags: m.rsyncLastRsyncFlags,
			ScpFlags:   m.rsyncLastScpFlags,
			LocalPath:  m.rsyncLastLocalPath,
			RemotePath: m.rsyncLastRemotePath,
		},
		RecentTargets: append([]string(nil), m.recentTargets...),
	}
}

func (m *Model) recordTargetUse(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	m.recentTargets = updatedRecentTargets(m.recentTargets, name, recentTargetLimit)
	m.rebuildTargetIndex()
	m.selectTargetByName(name)
	m.persistState()
}

func updatedRecentTargets(current []string, name string, limit int) []string {
	updated := []string{name}
	for _, existing := range current {
		existing = strings.TrimSpace(existing)
		if existing == "" || existing == name {
			continue
		}
		updated = append(updated, existing)
		if len(updated) == limit {
			break
		}
	}
	return updated
}

func (m *Model) rebuildTargetIndex() {
	m.index = search.New(m.targets, m.recentTargets...)
	m.applyFilter()
}

func (m *Model) selectTargetByName(name string) {
	for i, target := range m.filtered {
		if target.Name == name {
			m.selected = i
			return
		}
	}
}

func (m *Model) persistState() {
	if strings.TrimSpace(m.cacheDir()) == "" {
		return
	}
	_ = cache.SaveState(m.cacheDir(), m.stateSnapshot())
}

func (m *Model) applySnapshot(snap cache.Snapshot) {
	m.snapshot = snap
	if cacheDir := m.cacheDir(); cacheDir != "" {
		_ = cache.Save(cacheDir, snap)
	}
	if snap.Info.Ports.SSH != 0 && (m.cfg.SSH.Port == 0 || m.cfg.SSH.Port == 2222) {
		m.cfg.SSH.Port = snap.Info.Ports.SSH
	}
	m.setTargets(snap.Targets)
}

func (m *Model) setTargets(targets []api.Target) {
	m.targets = targets
	m.rebuildTargetIndex()
}

func (m *Model) applyFilter() {
	m.filtered = m.index.Filter(m.searchBox.Value())
	if m.selected >= len(m.filtered) {
		m.selected = max(0, len(m.filtered)-1)
	}
}

func (m *Model) moveSelection(delta int) {
	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}
	m.selected = clamp(m.selected+delta, 0, len(m.filtered)-1)
}

func (m *Model) toggleHelp() {
	m.showHelp = !m.showHelp
	m.help.ShowAll = m.showHelp
}

func (m *Model) focusSearch(status string) {
	m.searchFocused = true
	m.searchBox.Placeholder = searchPlaceholderFocused
	m.searchBox.Focus()
	if status != "" {
		m.status = status
	}
}

func (m *Model) blurSearch() {
	m.searchFocused = false
	m.searchBox.Placeholder = searchPlaceholderIdle
	m.searchBox.Blur()
}

func (m *Model) clearSearch() {
	if m.searchBox.Value() == "" {
		return
	}
	m.searchBox.SetValue("")
	m.applyFilter()
	m.status = fmt.Sprintf("showing %d targets", len(m.filtered))
}

func (m *Model) clearOrBlurSearch() {
	if m.searchFocused {
		m.blurSearch()
		m.status = fmt.Sprintf("search blurred - %d targets filtered", len(m.filtered))
		return
	}
	if m.searchBox.Value() != "" {
		m.clearSearch()
		return
	}
	m.status = "search cleared"
}

func (m *Model) openConfigEditor() {
	m.mode = modeConfig
	m.startForm(true)
	m.formStatus = "edit config and press enter to save"
}

func (m Model) connectSelected() (tea.Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		return m, nil
	}
	target := m.filtered[m.selected].Name
	m.recordTargetUse(target)
	return m, sshpkg.ExecCmd(m.cfgPath, m.cfg, target)
}

func (m *Model) copySelectedCommand() {
	if len(m.filtered) == 0 {
		return
	}
	bin, args := sshpkg.ShellCommand(m.cfg, m.filtered[m.selected].Name)
	if err := clipboard.WriteAll(commandText(bin, args)); err != nil {
		m.status = err.Error()
		return
	}
	m.status = "ssh command copied"
}

func loadCmd(cfgPath, cacheDirOverride string) tea.Cmd {
	return func() tea.Msg {
		if _, err := os.Stat(cfgPath); err != nil {
			if os.IsNotExist(err) {
				return loadedMsg{err: err, status: "no config found - complete onboarding"}
			}
			return loadedMsg{err: err, status: err.Error()}
		}

		cfg, err := cfgpkg.Load(cfgPath)
		if err != nil {
			return loadedMsg{err: err, status: "failed to load config - fix values below"}
		}

		cacheDir := resolveCacheDir(cfg.Cache.Dir, cacheDirOverride)
		snap, err := cache.Load(cacheDir)
		state, stateErr := cache.LoadState(cacheDir)
		status := "loaded config"
		switch {
		case err == nil:
			status = fmt.Sprintf("loaded %d cached targets", len(snap.Targets))
		case err != nil && !cache.Missing(err):
			status = "cache unavailable"
		}

		if stateErr != nil && !cache.Missing(stateErr) {
			status = joinStatus(status, "state unavailable")
		}

		return loadedMsg{cfg: cfg, snapshot: snap, state: state, status: status}
	}
}

func refreshCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return refreshMsg{err: fmt.Errorf("client not configured")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		info, err := client.GetInfo(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}
		targets, err := client.GetTargets(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		snap := cache.Snapshot{FetchedAt: time.Now(), Info: info, Targets: targets}
		return refreshMsg{snapshot: snap, status: fmt.Sprintf("refreshed %d targets", len(targets))}
	}
}

func validateAndSaveCmd(cfgPath string, cfg cfgpkg.Config) tea.Cmd {
	return func() tea.Msg {
		client := api.New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		info, err := client.GetInfo(ctx)
		if err != nil {
			return configSavedMsg{err: fmt.Errorf("validation failed: %w", err)}
		}
		targets, err := client.GetTargets(ctx)
		if err != nil {
			return configSavedMsg{err: fmt.Errorf("validation failed: %w", err)}
		}

		snap := cache.Snapshot{FetchedAt: time.Now(), Info: info, Targets: targets}
		if err := cfgpkg.Save(cfgPath, cfg); err != nil {
			return configSavedMsg{err: err}
		}

		return configSavedMsg{
			cfg:      cfg,
			snapshot: snap,
			status:   fmt.Sprintf("config saved - loaded %d targets", len(targets)),
		}
	}
}

func (m Model) View() string {
	switch m.mode {
	case modeBrowse:
		return m.viewBrowse()
	case modeTunnel:
		return m.viewTunnel()
	default:
		return m.viewForm()
	}
}

var (
	titleStyle        = lipgloss.NewStyle().Bold(true)
	mutedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	panelStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	headerStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statusBarStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("238")).Padding(0, 1)
	searchFocusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	appBadgeStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("12")).Padding(0, 1)
	modeActiveStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("205")).Padding(0, 1)
	modeInactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("239")).Padding(0, 1)
	actionsBarStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("236")).Padding(0, 1)
	actionKeyStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	actionDescStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	actionSepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)

func (m Model) viewBrowse() string {
	if m.width == 0 {
		m.width = 100
	}

	leftWidth := max(36, m.width/2-2)
	rightWidth := max(28, m.width-leftWidth-4)
	list := panelStyle.Width(leftWidth).Render(m.listView(leftWidth - 4))
	detail := panelStyle.Width(rightWidth).Render(m.detailView(rightWidth - 4))
	searchLabel := mutedStyle.Render("Search")
	if m.searchFocused {
		searchLabel = searchFocusStyle.Render("Search")
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, list, detail)

	return strings.Join([]string{
		m.viewTopChrome("wgt", m.browseTitleText(), m.browseHelpBindings().short),
		searchLabel,
		m.searchBox.View(),
		body,
		statusBarStyle.Width(max(0, m.width-2)).Render(m.status),
	}, "\n")
}

func (m Model) listView(width int) string {
	rows := []string{headerStyle.Render("Targets")}
	if len(m.filtered) == 0 {
		rows = append(rows, mutedStyle.Render("No targets match the current query."))
		return strings.Join(rows, "\n")
	}

	available := max(5, m.height-10)
	start := max(0, m.selected-available+1)
	if m.selected < available {
		start = 0
	}
	end := min(len(m.filtered), start+available)

	for i := start; i < end; i++ {
		rows = append(rows, renderTargetListLineWithRecent(m.filtered[i], width, i == m.selected, m.isRecentTarget(m.filtered[i].Name)))
	}

	if start > 0 || end < len(m.filtered) {
		rows = append(rows, "", mutedStyle.Render(fmt.Sprintf("showing %d-%d of %d", start+1, end, len(m.filtered))))
	}
	return strings.Join(rows, "\n")
}

func (m Model) detailView(width int) string {
	rows := []string{headerStyle.Render("Details")}
	if len(m.filtered) == 0 {
		rows = append(rows, mutedStyle.Render("No target selected."))
		return strings.Join(rows, "\n")
	}

	t := m.filtered[m.selected]
	rows = append(rows,
		titleStyle.Render(t.Name),
		fmt.Sprintf("kind: %s", fallback(t.Kind, "-")),
		fmt.Sprintf("group: %s", fallback(t.Group.Name, "-")),
	)
	if t.ExternalHost != "" {
		rows = append(rows, fmt.Sprintf("external host: %s", t.ExternalHost))
	}
	if t.DefaultDatabaseName != "" {
		rows = append(rows, fmt.Sprintf("default db: %s", t.DefaultDatabaseName))
	}
	if t.Description != "" {
		rows = append(rows, "", lipgloss.NewStyle().Width(width).Render(t.Description))
	}
	bin, args := sshpkg.ShellCommand(m.cfg, t.Name)
	rows = append(rows, commandPreviewRows("SSH command", commandText(bin, args), width)...)
	if !m.snapshot.FetchedAt.IsZero() {
		freshness := "fresh"
		if cache.IsStale(m.snapshot, m.cfg.Cache.TTL) {
			freshness = "stale"
		}
		rows = append(rows, "", mutedStyle.Render(fmt.Sprintf("cache: %s (%s)", freshness, m.snapshot.FetchedAt.Format(time.RFC822))))
	}
	return strings.Join(rows, "\n")
}

func (m Model) browseTitleText() string {
	if m.searchFocused {
		return ""
	}
	if strings.TrimSpace(m.searchBox.Value()) != "" {
		return "Search results"
	}
	return "Browse targets"
}

func (m Model) viewTopChrome(title, summary string, bindings []key.Binding) string {
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		appBadgeStyle.Render(title),
		" ",
		modeActiveStyle.Render(m.currentPageLabel()),
	)
	rows := []string{header}
	if strings.TrimSpace(summary) != "" {
		rows = append(rows, mutedStyle.Render(summary))
	}
	if shortcuts := renderShortcutBar(bindings, m.width); shortcuts != "" {
		rows = append(rows, shortcuts)
	}
	return strings.Join(rows, "\n")
}

func (m Model) currentPageLabel() string {
	if m.mode == modeBrowse && (m.searchFocused || strings.TrimSpace(m.searchBox.Value()) != "") {
		return "search"
	}
	switch m.mode {
	case modeBrowse:
		return "browse"
	case modeTunnelForm, modeTunnel:
		return "tunnel"
	case modeRsyncForm:
		return "transfer"
	case modeConfig:
		return "config"
	case modeOnboarding:
		return "onboarding"
	default:
		return "browse"
	}
}

func commandText(bin string, args []string) string {
	parts := append([]string{bin}, args...)
	return strings.TrimSpace(strings.Join(parts, " "))
}

func commandPreviewRows(label, command string, width int) []string {
	return commandPreviewRowsWithHint(label, command, "press c to copy this command", width)
}

func commandPreviewRowsWithHint(label, command, hint string, width int) []string {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}
	return []string{
		"",
		headerStyle.Render(label),
		lipgloss.NewStyle().Width(width).Render(command),
		mutedStyle.Render(hint),
	}
}

func renderShortcutBar(bindings []key.Binding, width int) string {
	parts := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if !binding.Enabled() {
			continue
		}
		help := binding.Help()
		if strings.TrimSpace(help.Key) == "" || strings.TrimSpace(help.Desc) == "" {
			continue
		}
		parts = append(parts, actionKeyStyle.Render(help.Key)+" "+actionDescStyle.Render(help.Desc))
	}
	if len(parts) == 0 {
		return ""
	}

	availableWidth := max(1, width-2)
	sep := actionSepStyle.Render("  •  ")
	lines := make([]string, 0, 2)
	current := ""
	for _, part := range parts {
		candidate := part
		if current != "" {
			candidate = current + sep + part
		}
		if current != "" && lipgloss.Width(candidate) > availableWidth {
			lines = append(lines, current)
			current = part
			continue
		}
		current = candidate
	}
	if current != "" {
		lines = append(lines, current)
	}

	return actionsBarStyle.Width(availableWidth).Render(strings.Join(lines, "\n"))
}

func joinStatus(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func fallback(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func (m Model) isRecentTarget(name string) bool {
	name = strings.TrimSpace(name)
	for _, recent := range m.recentTargets {
		if recent == name {
			return true
		}
	}
	return false
}

func renderTargetListLine(t api.Target, width int, selected bool) string {
	return renderTargetListLineWithRecent(t, width, selected, false)
}

func renderTargetListLineWithRecent(t api.Target, width int, selected, recent bool) string {
	prefix := "  "
	prefixRendered := prefix
	main := fmt.Sprintf("%s [%s] %s", t.Name, t.Kind, t.Group.Name)
	available := max(0, width-lipgloss.Width(prefix))
	secondaryParts := make([]string, 0, 2)
	if recent {
		secondaryParts = append(secondaryParts, recentTargetMarker)
	}
	if description := strings.TrimSpace(t.Description); description != "" {
		secondaryParts = append(secondaryParts, description)
	}
	secondary := strings.Join(secondaryParts, " ")

	if secondary != "" {
		mainWidth := lipgloss.Width(main)
		switch {
		case mainWidth >= available:
			main = truncate(main, available)
			secondary = ""
		case mainWidth+1+lipgloss.Width(secondary) > available:
			secondary = truncate(secondary, max(0, available-mainWidth-1))
		}
	} else {
		main = truncate(main, available)
	}

	if selected {
		prefixRendered = selectedStyle.Render("• ")
		main = selectedStyle.Render(main)
	}
	if secondary != "" {
		return prefixRendered + main + " " + mutedStyle.Render(secondary)
	}
	return prefixRendered + main
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}

	limit := width - 1
	runes := []rune(s)
	out := make([]rune, 0, len(runes))
	used := 0
	for _, r := range runes {
		rw := lipgloss.Width(string(r))
		if used+rw > limit {
			break
		}
		out = append(out, r)
		used += rw
	}
	return string(out) + "…"
}

func clamp(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
