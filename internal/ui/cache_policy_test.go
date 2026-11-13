package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

func TestStartupHonorsCachePolicy(t *testing.T) {
	for _, tc := range []struct {
		name              string
		age               time.Duration
		fallback, visible bool
	}{
		{"fresh_without_fallback", time.Minute, false, true},
		{"fresh_with_fallback", time.Minute, true, true},
		{"stale_without_fallback", 30 * time.Minute, false, false},
		{"stale_with_fallback", 30 * time.Minute, true, true},
		{"expired_without_fallback", 30 * 24 * time.Hour, false, false},
		{"expired_with_fallback", 30 * 24 * time.Hour, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := cfgpkg.Default()
			cfg.Cache.Dir = t.TempDir()
			cfg.Cache.UseStaleOnError = tc.fallback
			path := filepath.Join(t.TempDir(), "config.huml")
			if err := cfgpkg.Save(path, cfg); err != nil {
				t.Fatal(err)
			}
			snapshot := cache.Snapshot{FetchedAt: time.Now().Add(-tc.age), Targets: []api.Target{{Name: "cached-host"}}}
			if err := cache.Save(cfg.Cache.Dir, snapshot); err != nil {
				t.Fatal(err)
			}
			loaded := loadCmd(path, "")().(loadedMsg)
			next, cmd := New(path).handleLoaded(loaded)
			m := next.(Model)
			if (len(m.filtered) > 0) != tc.visible {
				t.Fatalf("wrong cache visibility: %v", m.filtered)
			}
			wantRefresh := tc.age > cfg.Cache.TTL
			if (cmd != nil) != wantRefresh {
				t.Fatalf("wrong refresh command for %s", tc.name)
			}
		})
	}
}

func TestRefreshFailureHonorsCachePolicy(t *testing.T) {
	for _, tc := range []struct {
		name           string
		age            time.Duration
		fallback, keep bool
	}{
		{"fresh", time.Minute, false, true},
		{"stale_allowed", 30 * time.Minute, true, true},
		{"stale_disabled", 30 * time.Minute, false, false},
		{"expired_even_with_fallback", 30 * 24 * time.Hour, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newPagingModel(t, 1)
			m.cfg.Cache.UseStaleOnError = tc.fallback
			m.snapshot = cache.Snapshot{FetchedAt: time.Now().Add(-tc.age), Targets: append([]api.Target(nil), m.targets...)}
			next, _ := m.handleRefresh(refreshMsg{err: fmt.Errorf("offline")})
			result := next.(Model)
			if (len(result.filtered) > 0) != tc.keep {
				t.Fatal("wrong fallback behavior")
			}
			if !strings.Contains(result.status, "offline") {
				t.Fatal("refresh failure not reported")
			}
			if !tc.keep {
				_, cmd := result.connectSelected()
				if cmd != nil {
					t.Fatal("discarded cache can still connect")
				}
				if !result.snapshot.FetchedAt.IsZero() {
					t.Fatal("discarded snapshot remains in memory")
				}
			}
		})
	}
}

func TestFreshRefreshReplacesRejectedCache(t *testing.T) {
	m := newPagingModel(t, 1)
	m.cfg.Cache.Dir = t.TempDir()
	m.cfg.Cache.UseStaleOnError = false
	m.snapshot = cache.Snapshot{FetchedAt: time.Now().Add(-time.Hour), Targets: m.targets}
	next, _ := m.handleRefresh(refreshMsg{err: fmt.Errorf("offline")})
	m = next.(Model)
	next, _ = m.handleRefresh(refreshMsg{snapshot: cache.Snapshot{FetchedAt: time.Now(), Targets: []api.Target{{Name: "live-host"}}}})
	m = next.(Model)
	if len(m.filtered) != 1 || m.filtered[0].Name != "live-host" {
		t.Fatal("fresh results did not recover")
	}
}

func TestRejectedCacheRefreshesWhenMaxAgeIsShorterThanTTL(t *testing.T) {
	cfg := cfgpkg.Default()
	cfg.Cache.MaxAge = time.Hour
	cfg.Cache.TTL = 3 * time.Hour
	snapshot := cache.Snapshot{FetchedAt: time.Now().Add(-2 * time.Hour), Targets: []api.Target{{Name: "expired-host"}}}
	next, cmd := New("test").handleLoaded(loadedMsg{cfg: cfg, snapshot: snapshot})
	m := next.(Model)
	if len(m.filtered) != 0 {
		t.Fatal("expired targets should be discarded")
	}
	if cmd == nil {
		t.Fatal("rejected cache must trigger a refresh even within TTL")
	}
}
