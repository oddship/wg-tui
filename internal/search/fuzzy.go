package search

import (
	"sort"
	"strings"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/sahilm/fuzzy"
)

const (
	scoreNoMatch        = -1
	scoreExactName      = 10000
	scorePrefixName     = 8000
	scorePrimaryFuzzy   = 5000
	scoreSecondaryFuzzy = 1000
)

type IndexedTarget struct {
	Target    api.Target
	Primary   string
	Secondary string
}

type Index struct {
	Items []IndexedTarget
}

type scoredTarget struct {
	target api.Target
	score  int
}

func New(targets []api.Target) Index {
	items := make([]IndexedTarget, 0, len(targets))
	for _, t := range targets {
		items = append(items, IndexedTarget{
			Target:    t,
			Primary:   strings.TrimSpace(strings.Join([]string{t.Name, t.Group.Name, t.Kind}, " ")),
			Secondary: strings.TrimSpace(strings.Join([]string{t.Description, t.ExternalHost, t.DefaultDatabaseName}, " ")),
		})
	}
	return Index{Items: items}
}

func (idx Index) Filter(q string) []api.Target {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return idx.allTargets()
	}

	matches := make([]scoredTarget, 0, len(idx.Items))
	for _, item := range idx.Items {
		score := scoreTarget(item, q)
		if score == scoreNoMatch {
			continue
		}
		matches = append(matches, scoredTarget{target: item.Target, score: score})
	}

	sort.SliceStable(matches, func(i, j int) bool { return matches[i].score > matches[j].score })
	return unwrapTargets(matches)
}

func (idx Index) allTargets() []api.Target {
	out := make([]api.Target, 0, len(idx.Items))
	for _, item := range idx.Items {
		out = append(out, item.Target)
	}
	return out
}

func scoreTarget(item IndexedTarget, q string) int {
	name := strings.ToLower(item.Target.Name)
	switch {
	case name == q:
		return scoreExactName
	case strings.HasPrefix(name, q):
		return scorePrefixName - len(name)
	}
	if score, ok := fuzzyScore(q, item.Primary); ok {
		return scorePrimaryFuzzy + score
	}
	if score, ok := fuzzyScore(q, item.Secondary); ok {
		return scoreSecondaryFuzzy + score
	}
	return scoreNoMatch
}

func fuzzyScore(query, value string) (int, bool) {
	matches := fuzzy.Find(query, []string{value})
	if len(matches) == 0 {
		return 0, false
	}
	return matches[0].Score, true
}

func unwrapTargets(matches []scoredTarget) []api.Target {
	out := make([]api.Target, 0, len(matches))
	for _, match := range matches {
		out = append(out, match.target)
	}
	return out
}
