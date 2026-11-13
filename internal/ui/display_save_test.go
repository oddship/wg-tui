package ui

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestDisplaySettingsSaveOfflineAndPreserveTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("display save contacted API") }))
	url := server.URL
	server.Close()
	cfg := cfgpkg.Default()
	cfg.Server.URL = url
	cfg.Server.Token = "test"
	cfg.SSH.Username = "test"
	cfg.Cache.Dir = t.TempDir()
	path := filepath.Join(t.TempDir(), "config.huml")
	if err := cfgpkg.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	m := New(path)
	m.mode = modeConfig
	m.applyConfig(cfg)
	m.snapshot = cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "selected"}}}
	m.setTargets(m.snapshot.Targets)
	m.startForm(true)
	requireField(t, m.fields, "ui.target_contrast").input.SetValue("subtle")
	updated, err := m.formConfig()
	if err != nil {
		t.Fatal(err)
	}
	msg := m.saveConfigCmd(updated)().(configSavedMsg)
	if msg.err != nil || !msg.localOnly {
		t.Fatalf("offline save failed: %+v", msg)
	}
	saved, err := cfgpkg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.UI.TargetContrast != "subtle" {
		t.Fatal("display change not persisted")
	}
	model, _ := m.handleConfigSaved(msg)
	result := model.(Model)
	if result.snapshot.FetchedAt != m.snapshot.FetchedAt || len(result.filtered) != 1 || result.filtered[0].Name != "selected" {
		t.Fatal("local save replaced cached targets")
	}
	if result.mode != modeBrowse {
		t.Fatal("local save should return to browse")
	}
}

func TestConnectionChangesAndOnboardingStillValidateAPI(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	cfg := cfgpkg.Default()
	cfg.Server.URL = server.URL
	cfg.Server.Token = "old"
	cfg.SSH.Username = "test"
	cfg.SSH.Host = "test-host"
	path := filepath.Join(t.TempDir(), "config.huml")
	if err := cfgpkg.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []mode{modeConfig, modeOnboarding} {
		m := New(path)
		m.mode = mode
		m.applyConfig(cfg)
		next := cfg
		if mode == modeConfig {
			next.Server.Token = "changed"
		}
		msg := m.saveConfigCmd(next)().(configSavedMsg)
		if msg.err == nil || !strings.Contains(msg.err.Error(), "validation failed") {
			t.Fatalf("connection changes need API validation: %+v", msg)
		}
	}
	if requests != 2 {
		t.Fatalf("expected two validations, got %d", requests)
	}
	saved, err := cfgpkg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Server.Token != "old" {
		t.Fatal("failed validation replaced existing config")
	}
}
