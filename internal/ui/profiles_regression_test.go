package ui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestSwitchProfileBlocksQueuedActionsUntilCacheLoads(t *testing.T) {
	m := multiProfileModel(t)
	m.cacheDirOverride = t.TempDir()
	m.setTargets([]api.Target{{Name: "prod-only"}})
	m.recentTargets = []string{"prod-only"}
	staging := filepath.Join(m.cacheDirOverride, "staging")
	if err := cache.Save(staging, cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "stage-only"}}}); err != nil {
		t.Fatal(err)
	}
	if err := cache.SaveState(staging, cache.State{RecentTargets: []string{"stage-only"}}); err != nil {
		t.Fatal(err)
	}
	updated, load := m.switchProfile("staging")
	m = updated.(Model)
	if !m.profileLoading || len(m.filtered) != 0 || len(m.recentTargets) != 0 {
		t.Fatal("old targets or state available during switch")
	}
	for _, key := range []tea.KeyMsg{{Type: tea.KeyEnter}, {Type: tea.KeyRunes, Runes: []rune("t")}, {Type: tea.KeyRunes, Runes: []rune("s")}, {Type: tea.KeyRunes, Runes: []rune("e")}} {
		updated, cmd := m.Update(key)
		m = updated.(Model)
		if cmd != nil || m.mode != modeBrowse {
			t.Fatal("queued action ran while switching")
		}
	}
	// A refresh already in flight must not repopulate targets while loading.
	updated, _ = m.handleRefresh(refreshMsg{client: m.client, profile: "staging", snapshot: cache.Snapshot{Targets: []api.Target{{Name: "outdated"}}}})
	m = updated.(Model)
	if len(m.filtered) != 0 {
		t.Fatal("refresh bypassed switch lock")
	}
	updated, _ = m.Update(load())
	m = updated.(Model)
	if m.profileLoading || len(m.filtered) != 1 || m.filtered[0].Name != "stage-only" || m.recentTargets[0] != "stage-only" {
		t.Fatalf("destination did not load: %#v", m.filtered)
	}
}

func writeLegacyConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.huml")
	const data = "%HUML v0.2.0\nserver::\n  url: \"https://prod.example.com\"\n  token: \"token\"\nssh::\n  username: \"user\"\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLegacyStartupMigratesCacheOnceAndDisplaysRecoveryNotice(t *testing.T) {
	root, path := t.TempDir(), writeLegacyConfig(t)
	if err := cache.Save(root, cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "legacy-db"}}}); err != nil {
		t.Fatal(err)
	}
	if err := cache.SaveState(root, cache.State{RecentTargets: []string{"legacy-db"}, Transfer: cache.TransferState{RemotePath: "/saved"}}); err != nil {
		t.Fatal(err)
	}
	msg := loadCmd(path, root, "")().(loadedMsg)
	if msg.err != nil || len(msg.snapshot.Targets) != 1 || len(msg.state.RecentTargets) != 1 || msg.state.Transfer.RemotePath != "/saved" {
		t.Fatalf("legacy data not loaded: %#v", msg)
	}
	if !strings.Contains(msg.notice, "RESTORE.txt") {
		t.Fatalf("missing recovery notice: %s", msg.notice)
	}
	backups, err := filepath.Glob(filepath.Join(root, ".backups", "*"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected one backup: %v %v", backups, err)
	}
	if _, err := os.Stat(filepath.Join(backups[0], "config.huml")); err != nil {
		t.Fatal(err)
	}
	// Current profile data wins on subsequent starts, even while config remains legacy.
	if err := cache.SaveState(filepath.Join(root, "default"), cache.State{RecentTargets: []string{"newer"}}); err != nil {
		t.Fatal(err)
	}
	again := loadCmd(path, root, "")().(loadedMsg)
	if again.state.RecentTargets[0] != "newer" {
		t.Fatal("migration overwrote current profile data")
	}
	backups, _ = filepath.Glob(filepath.Join(root, ".backups", "*"))
	if len(backups) != 1 {
		t.Fatal("migration repeated")
	}
	m := New(path)
	updated, _ := m.handleLoaded(msg)
	m = updated.(Model)
	updated, _ = m.handleRefresh(refreshMsg{client: m.client, profile: "default", err: os.ErrPermission})
	m = updated.(Model)
	if !strings.Contains(m.View(), "RESTORE.txt") {
		t.Fatal("refresh hid recovery instructions")
	}
}

func TestLegacyMigrationFailureUsesOriginalDataAndWarns(t *testing.T) {
	root, path := t.TempDir(), writeLegacyConfig(t)
	if err := cache.Save(root, cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "legacy-db"}}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".backups"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	msg := loadCmd(path, root, "")().(loadedMsg)
	if msg.err != nil || !msg.migrationFailed || len(msg.snapshot.Targets) != 1 || !strings.Contains(msg.notice, "migration stopped") {
		t.Fatalf("expected fallback and warning: %#v", msg)
	}
	m := New(path)
	updated, cmd := m.handleLoaded(msg)
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("refresh scheduled after failed migration")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("target action permitted after failed migration")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("cannot exit after failed migration")
	}
}

func TestRenamePersistsStateWithBackupAndRecoveryNotice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/info") {
			_, _ = w.Write([]byte(`{}`))
		} else {
			_, _ = w.Write([]byte(`[{"name":"prod-db"}]`))
		}
	}))
	defer server.Close()
	m := multiProfileModel(t)
	m.cacheDirOverride = t.TempDir()
	m.cfgPath = filepath.Join(t.TempDir(), "config.huml")
	m.cfg.Server.URL = server.URL
	m.cfg.CommitActive()
	if err := cfgpkg.Save(m.cfgPath, m.cfg); err != nil {
		t.Fatal(err)
	}
	m.recentTargets = []string{"prod-db"}
	m.rsyncLastRemotePath = "/saved"
	m.persistState()
	m.startForm(true)
	requireField(t, m.fields, "profile.name").input.SetValue("production")
	cfg, err := m.formConfig()
	if err != nil {
		t.Fatal(err)
	}
	msg := validateAndSaveCmd(m.cfgPath, cfg, m.cacheDirOverride, m.formProfile, m.stateSnapshot())().(configSavedMsg)
	if msg.err != nil {
		t.Fatal(msg.err)
	}
	updated, _ := m.handleConfigSaved(msg)
	m = updated.(Model)
	reloaded := loadCmd(m.cfgPath, m.cacheDirOverride, "")().(loadedMsg)
	if reloaded.err != nil || reloaded.cfg.Profile != "production" || len(reloaded.state.RecentTargets) != 1 || reloaded.state.Transfer.RemotePath != "/saved" {
		t.Fatalf("rename lost state on restart: %#v", reloaded)
	}
	if _, err := cache.LoadState(filepath.Join(m.cacheDirOverride, "prod")); err != nil {
		t.Fatalf("original data removed: %v", err)
	}
	if !strings.Contains(m.recoveryNotice, "RESTORE.txt") {
		t.Fatal("missing restore instructions")
	}
	backups, _ := filepath.Glob(filepath.Join(m.cacheDirOverride, ".backups", "*"))
	if len(backups) != 1 {
		t.Fatal("expected rename backup")
	}
	old, err := cfgpkg.Load(filepath.Join(backups[0], "config.huml"))
	if err != nil || old.Profile != "prod" {
		t.Fatalf("old configuration not backed up: %v", err)
	}
}

func TestRenameBackupFailureDoesNotSaveConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/info") {
			_, _ = w.Write([]byte(`{}`))
		} else {
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	defer server.Close()
	m := multiProfileModel(t)
	root := t.TempDir()
	path := filepath.Join(t.TempDir(), "config.huml")
	m.cfg.Server.URL = server.URL
	m.cfg.CommitActive()
	if err := cfgpkg.Save(path, m.cfg); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".backups"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := m.cfg
	cfg.RenameActive("production")
	cfg.CommitActive()
	msg := validateAndSaveCmd(path, cfg, root, "prod", cache.State{})().(configSavedMsg)
	if msg.err == nil || !strings.Contains(msg.err.Error(), "rename stopped") {
		t.Fatalf("expected backup failure: %v", msg.err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(original) {
		t.Fatalf("config changed after backup failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "production")); !os.IsNotExist(err) {
		t.Fatalf("destination created on failure: %v", err)
	}
}
