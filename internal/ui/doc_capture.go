package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

const (
	docCaptureWidth  = 120
	docCaptureHeight = 34
)

type DocCaptureFixtures struct {
	Config  cfgpkg.Config
	Targets []api.Target
	State   cache.State
}

type DocCaptureScenario struct {
	Name   string
	View   string
	Frames []string
}

type docCaptureVHSGroup string

const (
	docCaptureVHSOverview docCaptureVHSGroup = "overview"
	docCaptureVHSForms    docCaptureVHSGroup = "forms"
)

type docCaptureSceneSpec struct {
	Name        string
	WaitPattern string
	Screenshot  string
	Group       docCaptureVHSGroup
	ShowSleep   string
	HoldSleep   string
	TourOrder   int
	BuildModel  func(DocCaptureFixtures) (Model, error)
}

type docCaptureVHSSettings struct {
	FontFamily string
	FontSize   int
	LineHeight float64
	Width      int
	Height     int
}

var docCaptureOverviewSettings = docCaptureVHSSettings{
	FontFamily: "JetBrains Mono",
	FontSize:   19,
	LineHeight: 1.05,
	Width:      1600,
	Height:     860,
}

var docCaptureFormSettings = docCaptureVHSSettings{
	FontFamily: "JetBrains Mono",
	FontSize:   17,
	LineHeight: 1.05,
	Width:      1600,
	Height:     1100,
}

var docCaptureSceneRegistry = []docCaptureSceneSpec{
	{
		Name:        "browse-overview",
		WaitPattern: "Browse targets",
		Screenshot:  "wgt-browse-overview.png",
		Group:       docCaptureVHSOverview,
		ShowSleep:   "450ms",
		HoldSleep:   "900ms",
		TourOrder:   1,
		BuildModel:  docCaptureBrowseOverviewModel,
	},
	{
		Name:        "search-prod",
		WaitPattern: "search> prod",
		Screenshot:  "wgt-search-prod.png",
		Group:       docCaptureVHSOverview,
		ShowSleep:   "450ms",
		HoldSleep:   "900ms",
		TourOrder:   2,
		BuildModel:  docCaptureSearchProdModel,
	},
	{
		Name:        "onboarding",
		WaitPattern: "Base URL",
		Screenshot:  "wgt-onboarding.png",
		Group:       docCaptureVHSForms,
		ShowSleep:   "450ms",
		HoldSleep:   "600ms",
		BuildModel:  docScenarioOnboardingModel,
	},
	{
		Name:        "config-editor",
		WaitPattern: "edit config and press enter to save",
		Screenshot:  "wgt-config-editor.png",
		Group:       docCaptureVHSForms,
		ShowSleep:   "450ms",
		HoldSleep:   "600ms",
		BuildModel:  docScenarioConfigEditorModel,
	},
	{
		Name:        "tunnel-form",
		WaitPattern: "Open Tunnel",
		Screenshot:  "wgt-tunnel-form.png",
		Group:       docCaptureVHSOverview,
		ShowSleep:   "450ms",
		HoldSleep:   "950ms",
		TourOrder:   3,
		BuildModel:  docScenarioTunnelFormModel,
	},
	{
		Name:        "transfer-form",
		WaitPattern: "Transfer - prod-reporting-db",
		Screenshot:  "wgt-transfer-form.png",
		Group:       docCaptureVHSOverview,
		ShowSleep:   "450ms",
		HoldSleep:   "950ms",
		TourOrder:   4,
		BuildModel:  docScenarioTransferFormModel,
	},
	{
		Name:        "tunnel-status",
		WaitPattern: "tunnel open:",
		Screenshot:  "wgt-tunnel-status.png",
		Group:       docCaptureVHSOverview,
		ShowSleep:   "450ms",
		HoldSleep:   "950ms",
		TourOrder:   5,
		BuildModel:  docScenarioTunnelStatusModel,
	},
}

func DocCaptureScenarioNames() []string {
	names := make([]string, 0, len(docCaptureSceneRegistry)+1)
	for _, spec := range docCaptureSceneRegistry {
		names = append(names, spec.Name)
	}
	names = append(names, "tour-sequence")
	return names
}

func BuildDocCaptureScenarios() ([]DocCaptureScenario, error) {
	prepareDocCaptureRendering()
	fixtures, err := LoadDocCaptureFixtures()
	if err != nil {
		return nil, err
	}

	scenarios := make([]DocCaptureScenario, 0, len(docCaptureSceneRegistry)+1)
	for _, spec := range docCaptureSceneRegistry {
		model, err := spec.BuildModel(fixtures)
		if err != nil {
			return nil, fmt.Errorf("build %s: %w", spec.Name, err)
		}
		scenarios = append(scenarios, DocCaptureScenario{Name: spec.Name, View: model.View()})
	}

	tourFrames := make([]string, 0, len(docCaptureSceneRegistry))
	for _, spec := range docCaptureTourSpecs() {
		model, err := spec.BuildModel(fixtures)
		if err != nil {
			return nil, fmt.Errorf("build %s: %w", spec.Name, err)
		}
		tourFrames = append(tourFrames, model.View())
	}
	scenarios = append(scenarios, DocCaptureScenario{Name: "tour-sequence", Frames: tourFrames})

	return scenarios, nil
}

func NewDocCaptureScenarioModel(name string) (Model, error) {
	prepareDocCaptureRendering()
	fixtures, err := LoadDocCaptureFixtures()
	if err != nil {
		return Model{}, err
	}

	spec, err := docCaptureSceneSpecByName(strings.TrimSpace(name))
	if err != nil {
		return Model{}, err
	}

	model, err := spec.BuildModel(fixtures)
	if err != nil {
		return Model{}, fmt.Errorf("build %s: %w", spec.Name, err)
	}
	model.skipInit = true
	return model, nil
}

func WriteDocCaptureVHSTapes(dir, binaryPath, assetDir, scratchDir string) error {
	themeJSON, err := loadDocCaptureThemeJSON()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	if err := writeDocCaptureOverviewTape(filepath.Join(dir, "overview.tape"), binaryPath, filepath.ToSlash(assetDir), themeJSON); err != nil {
		return err
	}
	if err := writeDocCaptureExtrasTape(filepath.Join(dir, "extras.tape"), binaryPath, filepath.ToSlash(assetDir), filepath.ToSlash(scratchDir), themeJSON); err != nil {
		return err
	}
	return nil
}

func LoadDocCaptureFixtures() (DocCaptureFixtures, error) {
	repoRoot, err := docCaptureRepoRoot()
	if err != nil {
		return DocCaptureFixtures{}, err
	}
	fixtureDir := filepath.Join(repoRoot, "docs", "demo", "fixtures")

	cfg, err := cfgpkg.Load(filepath.Join(fixtureDir, "config.huml"))
	if err != nil {
		return DocCaptureFixtures{}, fmt.Errorf("load config fixture: %w", err)
	}

	var targets []api.Target
	if err := readJSONFixture(filepath.Join(fixtureDir, "targets.json"), &targets); err != nil {
		return DocCaptureFixtures{}, err
	}

	var state cache.State
	if err := readJSONFixture(filepath.Join(fixtureDir, "state.json"), &state); err != nil {
		return DocCaptureFixtures{}, err
	}

	return DocCaptureFixtures{Config: cfg, Targets: targets, State: state}, nil
}

func prepareDocCaptureRendering() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
}

func readJSONFixture(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read fixture %s: %w", path, err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decode fixture %s: %w", path, err)
	}
	return nil
}

func docCaptureRepoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..")), nil
}

func frameFileName(index int) string {
	return fmt.Sprintf("frame-%04d.ansi", index)
}

func docCaptureSceneSpecByName(name string) (docCaptureSceneSpec, error) {
	if name == "" {
		name = "browse-overview"
	}
	for _, spec := range docCaptureSceneRegistry {
		if spec.Name == name {
			return spec, nil
		}
	}
	return docCaptureSceneSpec{}, fmt.Errorf("unknown docs scenario %q", name)
}

func docCaptureTourSpecs() []docCaptureSceneSpec {
	specs := make([]docCaptureSceneSpec, 0, len(docCaptureSceneRegistry))
	for _, spec := range docCaptureSceneRegistry {
		if spec.TourOrder > 0 {
			specs = append(specs, spec)
		}
	}
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].TourOrder < specs[j].TourOrder
	})
	return specs
}

func docCaptureGroupSpecs(group docCaptureVHSGroup) []docCaptureSceneSpec {
	specs := make([]docCaptureSceneSpec, 0, len(docCaptureSceneRegistry))
	for _, spec := range docCaptureSceneRegistry {
		if spec.Group == group {
			specs = append(specs, spec)
		}
	}
	return specs
}

func loadDocCaptureThemeJSON() (string, error) {
	repoRoot, err := docCaptureRepoRoot()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(filepath.Join(repoRoot, "docs", "demo", "vhs-theme.json"))
	if err != nil {
		return "", fmt.Errorf("read docs theme: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

func writeDocCaptureOverviewTape(path, binaryPath, assetDir, themeJSON string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Output %s\n", filepath.ToSlash(filepath.Join(assetDir, "wgt-overview.gif")))
	fmt.Fprintf(&b, "Output %s\n", filepath.ToSlash(filepath.Join(assetDir, "wgt-overview.mp4")))
	fmt.Fprintf(&b, "Output %s\n", filepath.ToSlash(filepath.Join(assetDir, "wgt-overview.webm")))
	writeDocCaptureVHSSettings(&b, docCaptureOverviewSettings, themeJSON)
	for _, spec := range docCaptureGroupSpecs(docCaptureVHSOverview) {
		writeDocCaptureVHSScene(&b, spec, binaryPath, assetDir)
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func writeDocCaptureExtrasTape(path, binaryPath, assetDir, scratchDir, themeJSON string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Output %s\n", filepath.ToSlash(filepath.Join(scratchDir, "extras.gif")))
	writeDocCaptureVHSSettings(&b, docCaptureFormSettings, themeJSON)
	for _, spec := range docCaptureGroupSpecs(docCaptureVHSForms) {
		writeDocCaptureVHSScene(&b, spec, binaryPath, assetDir)
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func writeDocCaptureVHSSettings(b *strings.Builder, settings docCaptureVHSSettings, themeJSON string) {
	fmt.Fprintf(b, "Set FontFamily \"%s\"\n", settings.FontFamily)
	fmt.Fprintf(b, "Set FontSize %d\n", settings.FontSize)
	fmt.Fprintf(b, "Set LineHeight %.2f\n", settings.LineHeight)
	fmt.Fprintf(b, "Set Width %d\n", settings.Width)
	fmt.Fprintf(b, "Set Height %d\n", settings.Height)
	fmt.Fprintf(b, "Set Theme %s\n", themeJSON)
	b.WriteString("Set TypingSpeed 0ms\n")
	b.WriteString("Set CursorBlink false\n")
}

func writeDocCaptureVHSScene(b *strings.Builder, spec docCaptureSceneSpec, binaryPath, assetDir string) {
	fmt.Fprintf(b, "Hide\nType \"%s --scenario %s\"\nEnter\nWait+Screen /%s/\nShow\nSleep %s\nScreenshot %s\nSleep %s\nHide\nCtrl+C\nSleep 150ms\n",
		binaryPath,
		spec.Name,
		spec.WaitPattern,
		spec.ShowSleep,
		filepath.ToSlash(filepath.Join(assetDir, spec.Screenshot)),
		spec.HoldSleep,
	)
}

func newDocCaptureModel(fixtures DocCaptureFixtures) Model {
	m := New("docs/demo/fixtures/config.huml")
	m.skipInit = true
	m.applyConfig(fixtures.Config)
	m.applyState(fixtures.State)
	m.setTargets(fixtures.Targets)
	m.setWindowSize(tea.WindowSizeMsg{Width: docCaptureWidth, Height: docCaptureHeight})
	m.mode = modeBrowse
	m.status = "loaded 8 cached targets"
	return m
}

func docCaptureBrowseOverviewModel(fixtures DocCaptureFixtures) (Model, error) {
	m := newDocCaptureModel(fixtures)
	m.selected = 0
	return m, nil
}

func docCaptureSearchProdModel(fixtures DocCaptureFixtures) (Model, error) {
	m := newDocCaptureModel(fixtures)
	return docReplaySearchQuery(m, "prod")
}

func docScenarioOnboardingModel(fixtures DocCaptureFixtures) (Model, error) {
	m := New("docs/demo/fixtures/config.huml")
	m.skipInit = true
	m.applyConfig(fixtures.Config)
	m.mode = modeOnboarding
	m.startForm(false)
	m.setWindowSize(tea.WindowSizeMsg{Width: docCaptureWidth, Height: docCaptureHeight})
	return m, nil
}

func docScenarioConfigEditorModel(fixtures DocCaptureFixtures) (Model, error) {
	m := New("docs/demo/fixtures/config.huml")
	m.skipInit = true
	m.applyConfig(fixtures.Config)
	m.mode = modeConfig
	m.startForm(true)
	if err := docFocusFieldByKey(&m, "ui.approval_open_mode"); err != nil {
		return Model{}, err
	}
	m.formStatus = "edit config and press enter to save"
	m.setWindowSize(tea.WindowSizeMsg{Width: docCaptureWidth, Height: docCaptureHeight})
	return m, nil
}

func docScenarioTunnelFormModel(fixtures DocCaptureFixtures) (Model, error) {
	m := newDocCaptureModel(fixtures)
	m.startTunnelForm("prod-reporting-db")
	if err := docSetFieldValueByKey(&m, "tunnel.remote_port", "5432"); err != nil {
		return Model{}, err
	}
	if err := docSetFieldValueByKey(&m, "tunnel.local_port", "15432"); err != nil {
		return Model{}, err
	}
	return m, nil
}

func docScenarioTransferFormModel(fixtures DocCaptureFixtures) (Model, error) {
	m := newDocCaptureModel(fixtures)
	m.startRsyncForm("prod-reporting-db")
	if err := docFocusFieldByKey(&m, "rsync.direction"); err != nil {
		return Model{}, err
	}
	return m, nil
}

func docScenarioTunnelStatusModel(fixtures DocCaptureFixtures) (Model, error) {
	m := newDocCaptureModel(fixtures)
	m.mode = modeTunnel
	m.tunnel.target = "prod-reporting-db"
	m.tunnel.remotePort = 5432
	m.tunnel.localPort = 15432
	m.tunnel.state = tunnelOpen
	m.status = "tunnel open: 127.0.0.1:15432 -> prod-reporting-db:127.0.0.1:5432"
	return m, nil
}

func docReplaySearchQuery(m Model, query string) (Model, error) {
	var err error
	m, err = applyDocKey(m, keyRunes("/"))
	if err != nil {
		return Model{}, err
	}
	for _, r := range query {
		m, err = applyDocKey(m, keyRune(r))
		if err != nil {
			return Model{}, err
		}
	}
	return m, nil
}

func docFieldIndexByKey(m *Model, key string) (int, error) {
	for i, field := range m.fields {
		if field.key == key {
			return i, nil
		}
	}
	return -1, fmt.Errorf("field %q not found", key)
}

func docFocusFieldByKey(m *Model, key string) error {
	index, err := docFieldIndexByKey(m, key)
	if err != nil {
		return err
	}
	m.focusField(index)
	return nil
}

func docSetFieldValueByKey(m *Model, key, value string) error {
	index, err := docFieldIndexByKey(m, key)
	if err != nil {
		return err
	}
	before := m.fields[index].input.Value()
	m.fields[index].input.SetValue(value)
	switch m.fields[index].key {
	case "tunnel.remote_port", "tunnel.local_port":
		m.syncTunnelFormFields(index, before, value)
	default:
		m.rememberTransferDraft(index)
	}
	return nil
}

func applyDocKey(m Model, msg tea.KeyMsg) (Model, error) {
	next, cmd := m.Update(msg)
	if cmd != nil {
		_ = cmd()
	}
	updated, ok := next.(Model)
	if !ok {
		return Model{}, fmt.Errorf("unexpected model type %T", next)
	}
	return updated, nil
}

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}
