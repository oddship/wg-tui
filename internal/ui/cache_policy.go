package ui

import (
	"github.com/oddship/wg-tui/internal/cache"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

// Empty state means no cache was loaded and the API must supply the targets.
func cacheRejectionReason(cfg cfgpkg.Config, snapshot cache.Snapshot) string {
	if snapshot.FetchedAt.IsZero() && len(snapshot.Targets) == 0 {
		return ""
	}
	if !cache.IsUsable(snapshot, cfg.Cache.MaxAge) {
		return "cached targets exceeded max_age or have no valid timestamp"
	}
	if cache.IsStale(snapshot, cfg.Cache.TTL) && !cfg.Cache.UseStaleOnError {
		return "stale cached targets disabled by use_stale_on_error"
	}
	return ""
}
