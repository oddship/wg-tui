package config

import (
	"errors"
	"os"
	"testing"
)

func TestConfigPathUsesUserConfigDir(t *testing.T) {
	reset := stubPathResolvers(t)
	defer reset()
	userConfigDirFn = func() (string, error) { return "/tmp/config-home", nil }

	if got := ConfigPath(); got != "/tmp/config-home/wgt/config.huml" {
		t.Fatalf("expected config path %q, got %q", "/tmp/config-home/wgt/config.huml", got)
	}
}

func TestConfigPathFallsBackToHome(t *testing.T) {
	reset := stubPathResolvers(t)
	defer reset()
	userConfigDirFn = func() (string, error) { return "", errors.New("boom") }
	userHomeDirFn = func() (string, error) { return "/home/tester", nil }

	if got := ConfigPath(); got != "/home/tester/.config/wgt/config.huml" {
		t.Fatalf("expected fallback config path %q, got %q", "/home/tester/.config/wgt/config.huml", got)
	}
}

func TestDefaultCacheDirUsesUserCacheDir(t *testing.T) {
	reset := stubPathResolvers(t)
	defer reset()
	userCacheDirFn = func() (string, error) { return "/tmp/cache-home", nil }

	if got := DefaultCacheDir(); got != "/tmp/cache-home/wgt" {
		t.Fatalf("expected cache dir %q, got %q", "/tmp/cache-home/wgt", got)
	}
}

func TestDefaultCacheDirFallsBackToHome(t *testing.T) {
	reset := stubPathResolvers(t)
	defer reset()
	userCacheDirFn = func() (string, error) { return "", errors.New("boom") }
	userHomeDirFn = func() (string, error) { return "/home/tester", nil }

	if got := DefaultCacheDir(); got != "/home/tester/.cache/wgt" {
		t.Fatalf("expected fallback cache dir %q, got %q", "/home/tester/.cache/wgt", got)
	}
}

func TestDefaultUsesDefaultCacheDir(t *testing.T) {
	reset := stubPathResolvers(t)
	defer reset()
	userCacheDirFn = func() (string, error) { return "/tmp/cache-home", nil }

	if got := Default().Cache.Dir; got != "/tmp/cache-home/wgt" {
		t.Fatalf("expected default cache dir %q, got %q", "/tmp/cache-home/wgt", got)
	}
}

func stubPathResolvers(t *testing.T) func() {
	t.Helper()
	prevConfigDir := userConfigDirFn
	prevCacheDir := userCacheDirFn
	prevHomeDir := userHomeDirFn
	return func() {
		userConfigDirFn = prevConfigDir
		userCacheDirFn = prevCacheDir
		userHomeDirFn = prevHomeDir
	}
}

func TestTargetDisplayRoundTrip(t *testing.T) {
	cfg := Default()
	cfg.UI.TargetFields = "kind, description"
	cfg.UI.TargetContrast = "subtle"
	path := t.TempDir() + "/config.huml"
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.UI.TargetFields != cfg.UI.TargetFields || loaded.UI.TargetContrast != cfg.UI.TargetContrast {
		t.Fatalf("display settings lost: %+v", loaded.UI)
	}
}

func TestTargetDisplayValidation(t *testing.T) {
	for _, fields := range []string{"description,group", "none", "", "kind"} {
		if err := ValidateTargetDisplay(UIConfig{TargetFields: fields, TargetContrast: "normal"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, ui := range []UIConfig{{TargetFields: "unknown"}, {TargetFields: "group,group"}, {TargetFields: "none,kind"}, {TargetContrast: "dim"}} {
		if err := ValidateTargetDisplay(ui); err == nil {
			t.Fatalf("accepted invalid settings: %+v", ui)
		}
	}
}

func TestLegacyConfigUsesTargetDisplayDefaults(t *testing.T) {
	path := t.TempDir() + "/config.huml"
	if err := os.WriteFile(path, []byte("%HUML v0.2.0\nui::\n  details_pane: \"right\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UI.TargetFields != Default().UI.TargetFields || cfg.UI.TargetContrast != "normal" {
		t.Fatalf("missing defaults: %+v", cfg.UI)
	}
}
