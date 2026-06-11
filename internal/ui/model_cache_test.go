package ui

import "testing"

func TestResolveCacheDirPrefersOverride(t *testing.T) {
	got := resolveCacheDir("/config/cache", "/runtime/cache")
	if got != "/runtime/cache" {
		t.Fatalf("expected runtime override to win, got %q", got)
	}
}

func TestResolveCacheDirFallsBackToConfigValue(t *testing.T) {
	got := resolveCacheDir("/config/cache", "")
	if got != "/config/cache" {
		t.Fatalf("expected config cache dir, got %q", got)
	}
}
