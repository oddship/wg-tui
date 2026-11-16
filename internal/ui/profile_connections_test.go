package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
	sshpkg "github.com/oddship/wg-tui/internal/ssh"
)

func TestReservedProfileNameIsRejected(t *testing.T) {
	for _, name := range []string{"last_profile", "LAST_PROFILE", "Last_Profile"} {
		if err := cfgpkg.ValidateProfileName(name); err == nil {
			t.Fatalf("reserved name %q accepted", name)
		}
		m := multiProfileModel(t)
		m.startNewProfileForm()
		requireField(t, m.fields, "profile.name").input.SetValue(name)
		if _, err := m.formConfig(); err == nil {
			t.Fatalf("form accepted reserved name %q", name)
		}
	}
}

func TestCachedSSHPortSurvivesProfileSwitch(t *testing.T) {
	m := multiProfileModel(t)
	m.cacheDirOverride = t.TempDir()
	var snap cache.Snapshot
	snap.FetchedAt = time.Now()
	snap.Info.Ports.SSH = 2022
	snap.Targets = []api.Target{{Name: "db"}}
	// The API has discovered the upstream's actual port during the current run.
	m.applySnapshot(snap)
	if m.cfg.SSH.Port != 2022 {
		t.Fatal("initial port discovery failed")
	}
	anyModel, cmd := m.switchProfile("staging")
	m = anyModel.(Model)
	anyModel, _ = m.Update(cmd())
	m = anyModel.(Model)
	anyModel, cmd = m.switchProfile("prod")
	m = anyModel.(Model)
	anyModel, _ = m.Update(cmd())
	m = anyModel.(Model)
	if m.cfg.SSH.Port != 2022 {
		t.Fatalf("fresh cache has port 2022 but switching back uses %d", m.cfg.SSH.Port)
	}
}

func TestRefreshBeforeEditingSameProfileIsRejected(t *testing.T) {
	m := multiProfileModel(t)
	m.cacheDirOverride = t.TempDir()
	m.startForm(true)
	// Model an old API request completing after the user's replacement config
	// was validated and saved. Both requests carry the same profile name.
	old := refreshMsg{client: m.client, profile: "prod", snapshot: cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "old-upstream-db"}}}}
	cfg := m.cfg
	cfg.Server.URL = "https://replacement.example.com"
	cfg.SSH.Host = "replacement.example.com"
	cfg.CommitActive()
	saving, _ := m.beginConfigSave(cfg)
	m = saving.(Model)
	anyModel, _ := m.handleConfigSaved(configSavedMsg{generation: m.configSaveGeneration, cfg: cfg, snapshot: cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "replacement-db"}}}})
	m = anyModel.(Model)
	anyModel, _ = m.handleRefresh(old)
	m = anyModel.(Model)
	saved, err := cache.Load(m.cacheDir())
	if err != nil || len(saved.Targets) != 1 || saved.Targets[0].Name != "replacement-db" {
		t.Fatalf("old request changed disk cache: %#v %v", saved, err)
	}
	if len(m.filtered) != 1 || m.filtered[0].Name != "replacement-db" {
		t.Fatalf("old request restored targets for %s: %#v", m.cfg.SSH.Host, m.filtered)
	}
}

func TestLoadedProfileDerivesHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.huml")
	data := []byte("%HUML v0.2.0\ndefault_profile: \"staging\"\nprofiles::\n  staging::\n    server::\n      url: \"https://warpgate-staging.example.com\"\n      token: \"token\"\n    ssh::\n      username: \"user@example.com\"\n")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := cfgpkg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	_, args := sshpkg.ShellCommand(cfg, "db")
	if args[len(args)-1] != "warpgate-staging.example.com" {
		t.Fatalf("documented profile produces SSH host %q", args[len(args)-1])
	}
}

func TestCachedSSHPortPreservesExplicitOverride(t *testing.T) {
	m := multiProfileModel(t)
	m.cfg.SSH.Port = 2200
	var snap cache.Snapshot
	snap.FetchedAt = time.Now()
	snap.Info.Ports.SSH = 2022
	updated, _ := m.showCached(snap, cache.State{}, "")
	m = updated.(Model)
	if m.cfg.SSH.Port != 2200 {
		t.Fatalf("cached info overwrote explicit port: %d", m.cfg.SSH.Port)
	}
}

func TestRefreshCommandCarriesConnectionIdentity(t *testing.T) {
	for _, failPath := range []string{"", "/@warpgate/api/info", "/@warpgate/api/targets"} {
		t.Run(fmt.Sprintf("fail=%s", failPath), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == failPath {
					http.Error(w, "unavailable", 503)
					return
				}
				if r.URL.Path == "/@warpgate/api/info" {
					_, _ = w.Write([]byte(`{}`))
				} else {
					_, _ = w.Write([]byte(`[{"name":"db"}]`))
				}
			}))
			defer server.Close()
			cfg := cfgpkg.Default()
			cfg.Server.URL = server.URL
			client := api.New(cfg)
			msg := refreshCmd("prod", client)().(refreshMsg)
			if msg.client != client || msg.profile != "prod" {
				t.Fatal("refresh lost connection identity")
			}
			if (msg.err != nil) != (failPath != "") {
				t.Fatalf("unexpected result: %v", msg.err)
			}
		})
	}
}

func TestLocalPreferencesKeepValidRefresh(t *testing.T) {
	m := multiProfileModel(t)
	m.cacheDirOverride = t.TempDir()
	client := m.client
	cfg := m.cfg
	cfg.UI.TargetContrast = "subtle"
	m.applyConfig(cfg)
	if m.client != client {
		t.Fatal("local preferences invalidated a valid connection")
	}
	updated, _ := m.handleRefresh(refreshMsg{client: client, profile: m.cfg.Profile, snapshot: cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "db"}}}})
	m = updated.(Model)
	if len(m.filtered) != 1 || m.filtered[0].Name != "db" {
		t.Fatal("valid refresh was discarded")
	}
}

func TestCachedSSHPortIgnoredWhenCachePolicyRejectsSnapshot(t *testing.T) {
	m := multiProfileModel(t)
	var snap cache.Snapshot
	snap.FetchedAt = time.Now().Add(-30 * 24 * time.Hour)
	snap.Info.Ports.SSH = 2022
	updated, _ := m.showCached(snap, cache.State{}, "")
	m = updated.(Model)
	if m.cfg.SSH.Port != 2222 {
		t.Fatalf("rejected cache changed the SSH port: %d", m.cfg.SSH.Port)
	}
}
