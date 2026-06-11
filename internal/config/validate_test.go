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
