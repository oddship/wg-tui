package search

import (
	"sort"
	"strings"
	"unicode"

	"github.com/oddship/wg-tui/internal/api"
	"github.com/sahilm/fuzzy"
)

const (
	scoreNoMatch             = -1
	scoreExactName           = 12000
	scorePrefixName          = 10000
	scoreContainsName        = 9000
	scoreExactDescription    = 8500
	scorePrefixDescription   = 8000
	scoreContainsDescription = 7500
	scorePrimaryContains     = 7000
	scoreSecondaryContains   = 6500
	scorePrimaryFuzzy        = 5000
	scoreSecondaryFuzzy      = 3000
)

type IndexedTarget struct {
	Target      api.Target
	Name        string
	Description string
	Primary     string
	Secondary   string
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
			Target:      t,
			Name:        normalize(t.Name),
			Description: normalize(t.Description),
			Primary:     normalize(strings.Join([]string{t.Name, t.Group.Name, t.Kind}, " ")),
			Secondary:   normalize(strings.Join([]string{t.Description, t.ExternalHost, t.DefaultDatabaseName}, " ")),
		})
	}
	return Index{Items: items}
}

func (idx Index) Filter(q string) []api.Target {
	q = normalize(q)
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
	if score := phraseScore(q, item.Name, scoreExactName, scorePrefixName, scoreContainsName); score != scoreNoMatch {
		return score
	}
	if score := phraseScore(q, item.Description, scoreExactDescription, scorePrefixDescription, scoreContainsDescription); score != scoreNoMatch {
		return score
	}
	if score := containsScore(q, item.Primary, scorePrimaryContains); score != scoreNoMatch {
		return score
	}
	if score := containsScore(q, item.Secondary, scoreSecondaryContains); score != scoreNoMatch {
		return score
	}
	if score, ok := fuzzyScore(q, item.Primary); ok {
		return scorePrimaryFuzzy + score
	}
	if score, ok := fuzzyScore(q, item.Secondary); ok {
		return scoreSecondaryFuzzy + score
	}
	return scoreNoMatch
}

func phraseScore(query, value string, exact, prefix, contains int) int {
	if value == "" {
		return scoreNoMatch
	}
	switch {
	case value == query:
		return exact - len(value)
	case strings.HasPrefix(value, query):
		return prefix - len(value)
	case strings.Contains(value, query):
		return contains - len(value)
	default:
		return scoreNoMatch
	}
}

func containsScore(query, value string, base int) int {
	if value == "" || !strings.Contains(value, query) {
		return scoreNoMatch
	}
	return base - len(value)
}

func fuzzyScore(query, value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	matches := fuzzy.Find(query, []string{value})
	if len(matches) == 0 {
		return 0, false
	}
	return matches[0].Score, true
}

func normalize(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	lastWasSpace := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			b.WriteByte(' ')
			lastWasSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func unwrapTargets(matches []scoredTarget) []api.Target {
	out := make([]api.Target, 0, len(matches))
	for _, match := range matches {
		out = append(out, match.target)
	}
	return out
}
