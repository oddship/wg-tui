package config

import (
	"errors"
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
