package ui

import (
	"testing"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/oddship/wg-tui/internal/search"
)

func TestRecordTargetUseMovesExistingTargetToFront(t *testing.T) {
	m := New("test")
	m.recentTargets = []string{"prod-api", "prod-db", "prod-cache"}

	m.recordTargetUse("prod-db")

	want := []string{"prod-db", "prod-api", "prod-cache"}
	assertRecentTargets(t, m.recentTargets, want)
}

func TestRecordTargetUseCapsRecentTargets(t *testing.T) {
	m := New("test")
	for i := 0; i < recentTargetLimit+3; i++ {
		m.recordTargetUse(targetName(i))
	}

	if len(m.recentTargets) != recentTargetLimit {
		t.Fatalf("expected %d recent targets, got %d", recentTargetLimit, len(m.recentTargets))
	}
	if m.recentTargets[0] != targetName(recentTargetLimit+2) {
		t.Fatalf("expected newest target first, got %q", m.recentTargets[0])
	}
	if m.recentTargets[len(m.recentTargets)-1] != targetName(3) {
		t.Fatalf("expected oldest retained target to be %q, got %q", targetName(3), m.recentTargets[len(m.recentTargets)-1])
	}
}

func TestRecordTargetUseIgnoresBlankNames(t *testing.T) {
	m := New("test")
	m.recentTargets = []string{"prod-api"}

	m.recordTargetUse("   ")

	want := []string{"prod-api"}
	assertRecentTargets(t, m.recentTargets, want)
}

func TestConnectSelectedRecordsTargetUse(t *testing.T) {
	m := New("test")
	m.filtered = []api.Target{{Name: "prod-api"}}

	updatedAny, _ := m.connectSelected()
	updated := updatedAny.(Model)

	want := []string{"prod-api"}
	assertRecentTargets(t, updated.recentTargets, want)
}

func TestRecordTargetUseKeepsUsedTargetSelectedAfterReorder(t *testing.T) {
	m := New("test")
	m.targets = []api.Target{{Name: "one"}, {Name: "two"}, {Name: "three"}}
	m.filtered = append([]api.Target(nil), m.targets...)
	m.index = search.New(m.targets)
	m.selected = 2

	m.recordTargetUse("three")

	if m.selected != 0 {
		t.Fatalf("expected used target to stay selected after moving to the top, got index %d", m.selected)
	}
	if len(m.filtered) == 0 || m.filtered[m.selected].Name != "three" {
		t.Fatalf("expected selected target to remain three, got %#v", m.filtered)
	}
}

func assertRecentTargets(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d recent targets, got %d (%#v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected recent target order: want %#v got %#v", want, got)
		}
	}
}

func targetName(i int) string {
	return "target-" + string(rune('a'+i))
}
