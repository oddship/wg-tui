package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/oddship/wg-tui/internal/api"
)

type Snapshot struct {
	FetchedAt time.Time    `json:"fetched_at"`
	Info      api.Info     `json:"info"`
	Targets   []api.Target `json:"targets"`
}

func Load(dir string) (Snapshot, error) {
	b, err := os.ReadFile(filepath.Join(dir, "snapshot.json"))
	if err != nil {
		return Snapshot{}, err
	}
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}

func Save(dir string, s Snapshot) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "snapshot.json"), b, 0o600)
}

func IsUsable(s Snapshot, maxAge time.Duration) bool {
	if s.FetchedAt.IsZero() {
		return false
	}
	return time.Since(s.FetchedAt) <= maxAge
}

func IsStale(s Snapshot, ttl time.Duration) bool {
	if s.FetchedAt.IsZero() {
		return true
	}
	return time.Since(s.FetchedAt) > ttl
}

func Missing(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
