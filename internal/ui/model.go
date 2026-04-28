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
)

type Model struct {
	cfgPath string
	cfg     cfgpkg.Config
	keys    cfgpkg.KeyMap
	client  *api.Client

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

	formTitle  string
	formIndex  int
	fields     []formField
	formStatus string
	width      int
	height     int
}

type formField struct {
	key   string
	label string
	input textinput.Model
	hint  string
}

type loadedMsg struct {
	cfg      cfgpkg.Config
	snapshot cache.Snapshot
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

func New(cfgPath string) Model {
	ti := textinput.New()
	ti.Placeholder = "Type / to search targets"
	ti.Prompt = "search> "
	ti.Blur()

	m := Model{
		cfgPath:   cfgPath,
		searchBox: ti,
		help:      help.New(),
		status:    "loading...",
	}
	m.help.ShowAll = false
	return m
}

func (m Model) Init() tea.Cmd { return loadCmd(m.cfgPath) }

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
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.mode != modeBrowse {
			return m.updateForm(msg)
		}
		return m.updateBrowse(msg)
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
	m.snapshot = msg.snapshot
	m.setTargets(msg.snapshot.Targets)
	m.status = msg.status

	if msg.snapshot.FetchedAt.IsZero() || cache.IsStale(msg.snapshot, m.cfg.Cache.TTL) {
		m.status = joinStatus(msg.status, "refreshing in background...")
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
	m.mode = modeBrowse
	m.formStatus = ""
	m.status = msg.status
	return m, nil
}

func (m Model) handleSSHFinished(msg sshpkg.ExecFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.status = msg.Err.Error()
	} else {
		m.status = "ssh session ended"
	}
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

func (m *Model) applySnapshot(snap cache.Snapshot) {
	m.snapshot = snap
	_ = cache.Save(m.cfg.Cache.Dir, snap)
	if snap.Info.Ports.SSH != 0 && (m.cfg.SSH.Port == 0 || m.cfg.SSH.Port == 2222) {
		m.cfg.SSH.Port = snap.Info.Ports.SSH
	}
	m.setTargets(snap.Targets)
}

func (m *Model) setTargets(targets []api.Target) {
	m.targets = targets
	m.index = search.New(targets)
	m.applyFilter()
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
	m.searchBox.Focus()
	if status != "" {
		m.status = status
	}
}

func (m *Model) blurSearch() {
	m.searchFocused = false
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
	if m.searchBox.Value() != "" {
		m.clearSearch()
		return
	}
	m.blurSearch()
	m.status = "search blurred"
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
	return m, sshpkg.ExecCmd(m.cfg, m.filtered[m.selected].Name)
}

func (m *Model) copySelectedCommand() {
	if len(m.filtered) == 0 {
		return
	}
	bin, args := sshpkg.Command(m.cfg, m.filtered[m.selected].Name)
	if err := clipboard.WriteAll(bin + " " + strings.Join(args, " ")); err != nil {
		m.status = err.Error()
		return
	}
	m.status = "ssh command copied"
}

func loadCmd(cfgPath string) tea.Cmd {
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

		snap, err := cache.Load(cfg.Cache.Dir)
		status := "loaded config"
		switch {
		case err == nil:
			status = fmt.Sprintf("loaded %d cached targets", len(snap.Targets))
		case err != nil && !cache.Missing(err):
			status = "cache unavailable"
		}

		return loadedMsg{cfg: cfg, snapshot: snap, status: status}
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
	if m.mode != modeBrowse {
		return m.viewForm()
	}
	return m.viewBrowse()
}

var (
	titleStyle       = lipgloss.NewStyle().Bold(true)
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	panelStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	headerStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statusBarStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("238")).Padding(0, 1)
	searchFocusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

func (m Model) viewBrowse() string {
	if m.width == 0 {
		m.width = 100
	}

	leftWidth := max(36, m.width/2-2)
	rightWidth := max(28, m.width-leftWidth-4)
	list := panelStyle.Width(leftWidth).Render(m.listView(leftWidth - 4))
	detail := panelStyle.Width(rightWidth).Render(m.detailView(rightWidth - 4))
	header := headerStyle.Render("wgt") + "  " + mutedStyle.Render(m.summaryText())
	searchLabel := mutedStyle.Render("Search")
	if m.searchFocused {
		searchLabel = searchFocusStyle.Render("Search")
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, list, detail)

	return strings.Join([]string{
		header,
		searchLabel,
		m.searchBox.View(),
		body,
		statusBarStyle.Width(max(0, m.width-2)).Render(m.status),
		m.help.View(m.keys),
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
		t := m.filtered[i]
		line := truncate(fmt.Sprintf("%s [%s] %s", t.Name, t.Kind, t.Group.Name), width)
		if i == m.selected {
			rows = append(rows, selectedStyle.Render("• "+line))
		} else {
			rows = append(rows, "  "+line)
		}
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
	if !m.snapshot.FetchedAt.IsZero() {
		freshness := "fresh"
		if cache.IsStale(m.snapshot, m.cfg.Cache.TTL) {
			freshness = "stale"
		}
		rows = append(rows, "", mutedStyle.Render(fmt.Sprintf("cache: %s (%s)", freshness, m.snapshot.FetchedAt.Format(time.RFC822))))
	}
	return strings.Join(rows, "\n")
}

func (m Model) summaryText() string {
	parts := []string{fmt.Sprintf("%d targets", len(m.filtered))}
	if q := strings.TrimSpace(m.searchBox.Value()); q != "" {
		parts = append(parts, fmt.Sprintf("query=%q", q))
	}
	if len(m.filtered) > 0 {
		parts = append(parts, fmt.Sprintf("selected=%d", m.selected+1))
	}
	if m.searchFocused {
		parts = append(parts, "typing", "scroll with arrows")
	}
	return strings.Join(parts, " - ")
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

func truncate(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width-1]) + "…"
	}
	return s
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
