package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func multiProfileModel(t *testing.T) Model {
	t.Helper()
	m := New("test")
	cfg := cfgpkg.Default()
	cfg.Profiles = map[string]cfgpkg.Profile{
		"prod": {
			Server: cfgpkg.ServerConfig{URL: "https://prod.example.com", Token: "a", InsecureSkipTLSVerify: true},
			SSH:    cfgpkg.SSHConfig{Username: "user@example.com", Host: "prod.example.com", Port: 2222, Binary: "ssh", ExtraArgs: []string{"-v"}},
		},
		"staging": {
			Server: cfgpkg.ServerConfig{URL: "https://staging.example.com", Token: "b"},
			SSH:    cfgpkg.SSHConfig{Username: "user@example.com", Host: "staging.example.com", Port: 2222, Binary: "ssh"},
		},
	}
	cfg.DefaultProfile = "prod"
	if err := cfg.Use("prod"); err != nil {
		t.Fatalf("Use: %v", err)
	}
	m.applyConfig(cfg)
	return m
}

func TestCacheDirIsScopedToProfile(t *testing.T) {
	m := multiProfileModel(t)
	m.cacheDirOverride = "/tmp/wgt"
	if got := m.cacheDir(); got != "/tmp/wgt/prod" {
		t.Fatalf("expected profile cache dir, got %q", got)
	}
}

func TestNewProfileFormAddsProfileWithoutInheritingActiveSettings(t *testing.T) {
	m := multiProfileModel(t)
	m.startNewProfileForm()
	requireField(t, m.fields, "profile.name").input.SetValue("dev")
	requireField(t, m.fields, "server.url").input.SetValue("https://dev.example.com")
	requireField(t, m.fields, "server.token").input.SetValue("c")
	requireField(t, m.fields, "ssh.username").input.SetValue("user@example.com")

	cfg, err := m.formConfig()
	if err != nil {
		t.Fatalf("formConfig: %v", err)
	}
	if cfg.Profile != "dev" || cfg.DefaultProfile != "prod" || len(cfg.Profiles) != 3 {
		t.Fatalf("expected dev to be added alongside existing profiles, got active=%q default=%q %#v", cfg.Profile, cfg.DefaultProfile, cfg.Profiles)
	}
	dev := cfg.Profiles["dev"]
	if dev.Server.InsecureSkipTLSVerify || len(dev.SSH.ExtraArgs) != 0 {
		t.Fatalf("expected new profile not to inherit active profile settings, got %#v", dev)
	}
	if dev.SSH.Host != "dev.example.com" {
		t.Fatalf("expected ssh host derived from server url, got %q", dev.SSH.Host)
	}
}

func TestNewProfileFormRejectsExistingName(t *testing.T) {
	m := multiProfileModel(t)
	m.startNewProfileForm()
	requireField(t, m.fields, "profile.name").input.SetValue("staging")
	requireField(t, m.fields, "server.url").input.SetValue("https://dev.example.com")
	requireField(t, m.fields, "server.token").input.SetValue("c")
	requireField(t, m.fields, "ssh.username").input.SetValue("user@example.com")

	if _, err := m.formConfig(); err == nil {
		t.Fatal("expected duplicate profile name to fail")
	}
}

func TestEditFormRenamesActiveProfile(t *testing.T) {
	m := multiProfileModel(t)
	m.startForm(true)
	requireField(t, m.fields, "profile.name").input.SetValue("production")

	cfg, err := m.formConfig()
	if err != nil {
		t.Fatalf("formConfig: %v", err)
	}
	if _, ok := cfg.Profiles["prod"]; ok {
		t.Fatal("expected old profile name to be removed")
	}
	if cfg.Profile != "production" || cfg.DefaultProfile != "production" {
		t.Fatalf("expected rename to carry active and default, got active=%q default=%q", cfg.Profile, cfg.DefaultProfile)
	}
}

func TestRefreshFromPreviousProfileIsIgnored(t *testing.T) {
	m := multiProfileModel(t)
	m.setTargets([]api.Target{{Name: "prod-db"}})

	updatedAny, _ := m.handleRefresh(refreshMsg{profile: "staging", snapshot: cache.Snapshot{Targets: []api.Target{{Name: "staging-db"}}}})
	updated := updatedAny.(Model)

	if len(updated.targets) != 1 || updated.targets[0].Name != "prod-db" {
		t.Fatalf("expected stale refresh to be ignored, got %#v", updated.targets)
	}
}

func TestProfilePickerNOpensNewProfileForm(t *testing.T) {
	m := multiProfileModel(t)
	m.openProfilePicker()

	updatedAny, _ := m.updateProfilePicker(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	updated := updatedAny.(Model)

	if updated.mode != modeConfig || updated.formProfile != "" {
		t.Fatalf("expected new profile form, got mode=%v formProfile=%q", updated.mode, updated.formProfile)
	}
	if got := requireField(t, updated.fields, "profile.name").input.Value(); got != "" {
		t.Fatalf("expected blank profile name, got %q", got)
	}
}

func TestSwitchProfileRemembersLastUsed(t *testing.T) {
	m := multiProfileModel(t)
	m.cacheDirOverride = t.TempDir()

	updatedAny, _ := m.switchProfile("staging")
	updated := updatedAny.(Model)

	if got := cache.LoadLastProfile(updated.cacheDirOverride); got != "staging" {
		t.Fatalf("expected last profile to be remembered, got %q", got)
	}
}
