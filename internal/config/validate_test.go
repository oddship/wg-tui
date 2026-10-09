package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/huml"
)

func TestValidateAllowsDefaultKeymap(t *testing.T) {
	if err := validate(Default()); err != nil {
		t.Fatalf("validate(default): %v", err)
	}
}

func TestValidateRejectsConflictingBindings(t *testing.T) {
	cfg := Default()
	cfg.Keys.EditConfig = []string{"e"}
	cfg.Keys.Search = []string{"e"}

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected conflicting keybindings to fail validation")
	}
}

func TestValidateRejectsEmptyBinding(t *testing.T) {
	cfg := Default()
	cfg.Keys.Copy = []string{""}

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected empty keybinding to fail validation")
	}
}

func TestValidateRejectsInvalidApprovalOpenMode(t *testing.T) {
	cfg := Default()
	cfg.UI.ApprovalOpenMode = "sometimes"

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected invalid approval open mode to fail validation")
	}
}

func TestValidateRejectsApprovalOpenCommandWithoutPlaceholder(t *testing.T) {
	cfg := Default()
	cfg.UI.ApprovalOpenCommand = "xdg-open"

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected approval open command without placeholder to fail validation")
	}
}

func TestValidateAllowsEmptyApprovalOpenCommandWhenModeIsNever(t *testing.T) {
	cfg := Default()
	cfg.UI.ApprovalOpenMode = ApprovalOpenModeNever
	cfg.UI.ApprovalOpenCommand = ""

	if err := validate(cfg); err != nil {
		t.Fatalf("expected never mode to allow empty approval command: %v", err)
	}
}

func TestValidateRejectsTunnelBindingConflict(t *testing.T) {
	cfg := Default()
	cfg.Keys.Tunnel = []string{"enter"}

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected tunnel keybinding conflict to fail validation")
	}
}

func TestValidateRejectsRsyncBindingConflict(t *testing.T) {
	cfg := Default()
	cfg.Keys.Rsync = []string{"enter"}

	err := validate(cfg)
	if err == nil {
		t.Fatal("expected rsync keybinding conflict to fail validation")
	}
}

func TestNewKeyMapLeavesLegacyTunnelUnsetWhenTIsAlreadyUsed(t *testing.T) {
	cfg := Default()
	cfg.Keys.Tunnel = nil
	cfg.Keys.Search = []string{"t"}

	km := NewKeyMap(cfg.Keys)
	if km.Tunnel.Enabled() {
		t.Fatal("expected tunnel binding to stay disabled when legacy config already uses t")
	}
}

func TestNewKeyMapLeavesLegacyRsyncUnsetWhenSIsAlreadyUsed(t *testing.T) {
	cfg := Default()
	cfg.Keys.Rsync = nil
	cfg.Keys.Search = []string{"s"}

	km := NewKeyMap(cfg.Keys)
	if km.Rsync.Enabled() {
		t.Fatal("expected rsync binding to stay disabled when legacy config already uses s")
	}
}

func TestLoadLegacyConfigWithoutTunnelKeyStillLoads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.huml")
	content, err := huml.Parser().Marshal(map[string]any{
		"server": map[string]any{
			"url":   "https://warpgate.example.com",
			"token": "token",
		},
		"ssh": map[string]any{
			"username": "user@example.com",
		},
		"keys": map[string]any{
			"search": []string{"t"},
		},
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load legacy config: %v", err)
	}
	if len(cfg.Keys.Tunnel) != 0 {
		t.Fatalf("expected legacy config tunnel keys to remain unset, got %#v", cfg.Keys.Tunnel)
	}
	if km := NewKeyMap(cfg.Keys); km.Tunnel.Enabled() {
		t.Fatal("expected runtime tunnel key to stay disabled for conflicting legacy config")
	}
}

func TestLoadLegacyConfigWithoutRsyncKeyStillLoads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.huml")
	content, err := huml.Parser().Marshal(map[string]any{
		"server": map[string]any{
			"url":   "https://warpgate.example.com",
			"token": "token",
		},
		"ssh": map[string]any{
			"username": "user@example.com",
		},
		"keys": map[string]any{
			"search": []string{"s"},
		},
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load legacy config: %v", err)
	}
	if len(cfg.Keys.Rsync) != 0 {
		t.Fatalf("expected legacy config rsync keys to remain unset, got %#v", cfg.Keys.Rsync)
	}
	if km := NewKeyMap(cfg.Keys); km.Rsync.Enabled() {
		t.Fatal("expected runtime rsync key to stay disabled for conflicting legacy config")
	}
}

func TestLoadMigratesLegacyConfigToDefaultProfile(t *testing.T) {
	path := writeHUML(t, map[string]any{
		"server": map[string]any{"url": "https://warpgate.example.com", "token": "token"},
		"ssh":    map[string]any{"username": "user@example.com"},
	})

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load legacy config: %v", err)
	}
	if cfg.Profile != DefaultProfileName || cfg.DefaultProfile != DefaultProfileName {
		t.Fatalf("expected legacy config to load as %q, got active=%q default=%q", DefaultProfileName, cfg.Profile, cfg.DefaultProfile)
	}
	if cfg.Server.URL != "https://warpgate.example.com" || cfg.SSH.Port != Default().SSH.Port {
		t.Fatalf("expected legacy server/ssh to be active, got %#v %#v", cfg.Server, cfg.SSH)
	}
}

func TestSaveAndLoadRoundTripsProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.huml")
	cfg := Default()
	cfg.Profiles = map[string]Profile{
		"prod":    {Server: ServerConfig{URL: "https://prod.example.com", Token: "a"}, SSH: SSHConfig{Username: "user@example.com"}},
		"staging": {Server: ServerConfig{URL: "https://staging.example.com", Token: "b"}, SSH: SSHConfig{Username: "user@example.com", Port: 2022}},
	}
	cfg.DefaultProfile = "staging"
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Profile != "staging" || loaded.Server.URL != "https://staging.example.com" || loaded.SSH.Port != 2022 {
		t.Fatalf("expected staging to be active, got %q %#v %#v", loaded.Profile, loaded.Server, loaded.SSH)
	}
	if err := loaded.Use("prod"); err != nil {
		t.Fatalf("Use(prod): %v", err)
	}
	if loaded.Server.URL != "https://prod.example.com" || loaded.SSH.Port != Default().SSH.Port {
		t.Fatalf("expected prod values with default port, got %#v %#v", loaded.Server, loaded.SSH)
	}
}

func TestValidateProfilesRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", "..", "a/b", "a.b", "a b"} {
		cfg := Default()
		cfg.Profiles = map[string]Profile{name: {}}
		cfg.DefaultProfile = name
		if err := validateProfiles(cfg); err == nil {
			t.Fatalf("expected profile name %q to fail validation", name)
		}
	}
}

func TestValidateProfilesRejectsUnknownDefault(t *testing.T) {
	cfg := Default()
	cfg.Profiles = map[string]Profile{"prod": {}}
	cfg.DefaultProfile = "staging"
	if err := validateProfiles(cfg); err == nil {
		t.Fatal("expected unknown default_profile to fail validation")
	}
}

func TestRenameActiveMovesDefaultProfile(t *testing.T) {
	cfg := Default()
	cfg.Profiles = map[string]Profile{"prod": {}, "staging": {}}
	cfg.DefaultProfile = "prod"
	if err := cfg.Use("prod"); err != nil {
		t.Fatalf("Use: %v", err)
	}

	cfg.RenameActive("production")
	cfg.CommitActive()

	if _, ok := cfg.Profiles["prod"]; ok {
		t.Fatal("expected old profile name to be removed")
	}
	if _, ok := cfg.Profiles["production"]; !ok || cfg.DefaultProfile != "production" {
		t.Fatalf("expected renamed profile to be stored and default, got %#v default=%q", cfg.Profiles, cfg.DefaultProfile)
	}
}

func writeHUML(t *testing.T, data map[string]any) string {
	t.Helper()
	content, err := huml.Parser().Marshal(data)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	path := filepath.Join(t.TempDir(), "config.huml")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
