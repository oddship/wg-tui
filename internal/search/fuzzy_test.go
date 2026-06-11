package search

import (
	"testing"

	"github.com/oddship/wg-tui/internal/api"
)

func makeTarget(name, group, description string) api.Target {
	var t api.Target
	t.Name = name
	t.Kind = "Ssh"
	t.Group.Name = group
	t.Description = description
	return t
}

func TestNormalizeTreatsSeparatorsAsEquivalent(t *testing.T) {
	if got := normalize("zone-a1"); got != "zone a1" {
		t.Fatalf("expected normalized value %q, got %q", "zone a1", got)
	}
	if got := normalize("zone a1"); got != "zone a1" {
		t.Fatalf("expected normalized value %q, got %q", "zone a1", got)
	}
}

func TestFilterRanksDescriptionExactMatchAbovePrimaryFuzzy(t *testing.T) {
	idx := New([]api.Target{
		makeTarget("service-zone-j-1", "beta", "secondary node"),
		makeTarget("host-01", "alpha", "zone-a1"),
		makeTarget("service-zone-h-1", "beta", "another node"),
	})

	got := idx.Filter("zone a1")
	if len(got) == 0 {
		t.Fatal("expected at least one match")
	}
	if got[0].Name != "host-01" {
		t.Fatalf("expected description exact match to rank first, got %q", got[0].Name)
	}
}

func TestFilterTreatsSpaceAndHyphenQueriesTheSame(t *testing.T) {
	idx := New([]api.Target{
		makeTarget("host-01", "alpha", "zone-a1"),
		makeTarget("service-zone-j-1", "beta", "secondary node"),
	})

	for _, query := range []string{"zone a1", "zone-a1"} {
		got := idx.Filter(query)
		if len(got) == 0 {
			t.Fatalf("expected matches for query %q", query)
		}
		if got[0].Name != "host-01" {
			t.Fatalf("expected top result for %q to be %q, got %q", query, "host-01", got[0].Name)
		}
	}
}

func TestFilterEmptyQueryReturnsRecentTargetsFirst(t *testing.T) {
	idx := New([]api.Target{
		makeTarget("one", "alpha", ""),
		makeTarget("two", "alpha", ""),
		makeTarget("three", "alpha", ""),
	}, "three", "one")

	got := idx.Filter("")
	if len(got) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(got))
	}
	if got[0].Name != "three" || got[1].Name != "one" || got[2].Name != "two" {
		t.Fatalf("expected recents first ordering, got %#v", []string{got[0].Name, got[1].Name, got[2].Name})
	}
}

func TestFilterPrefersRecentTargetAmongSimilarMatches(t *testing.T) {
	idx := New([]api.Target{
		makeTarget("prod-api-b", "alpha", "secondary"),
		makeTarget("prod-api-a", "alpha", "primary"),
	}, "prod-api-a")

	got := idx.Filter("prod api")
	if len(got) == 0 {
		t.Fatal("expected at least one match")
	}
	if got[0].Name != "prod-api-a" {
		t.Fatalf("expected recent target to rank first, got %q", got[0].Name)
	}
}

func TestFilterExactMatchBeatsWeakerRecentMatch(t *testing.T) {
	idx := New([]api.Target{
		makeTarget("prod-api-helper", "alpha", "helper"),
		makeTarget("prod-api", "alpha", "exact"),
	}, "prod-api-helper")

	got := idx.Filter("prod api")
	if len(got) == 0 {
		t.Fatal("expected at least one match")
	}
	if got[0].Name != "prod-api" {
		t.Fatalf("expected exact match to outrank weaker recent match, got %q", got[0].Name)
	}
}
