package atlas

import (
	"strings"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/hermes"
)

// NormalizeTags cleans tags: lower case, whitespace collapsed, invisible characters removed, each at most 30
// characters, empty ones dropped, duplicates removed (first one kept). More than 10 left is an error.
func NormalizeTags(in []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, t := range in {
		n := strings.ToLower(hermes.Untrusted(t, MaxTagLen))
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	if len(out) > MaxTags {
		return nil, hermes.Errorf("at most %d tags per item.", MaxTags)
	}
	return out, nil
}

// ParseTerms splits a search query into words, each cleaned and cut to 50 characters. More than 5 is an error.
func ParseTerms(q string) ([]string, error) {
	var terms []string
	for _, f := range strings.Fields(hermes.Untrusted(q, 2000)) {
		if t := hermes.Untrusted(f, MaxTermLen); t != "" {
			terms = append(terms, t)
		}
	}
	if len(terms) > MaxTerms {
		return nil, hermes.Errorf("use at most %d search words.", MaxTerms)
	}
	return terms, nil
}

// LikePattern turns a search term into a LIKE pattern that matches it as a substring, with \, % and _ in the
// term matching themselves. Use it with ESCAPE '\'.
func LikePattern(term string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(term) + "%"
}

func priorityRank(p string) int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityNormal:
		return 1
	}
	return 2
}

// Less is the list order: open items first, then due date ascending (undated last), then priority (high first),
// then newest. The SQL ORDER BY in the Postgres store mirrors it.
func Less(a, b Memory) bool {
	if ao, bo := a.Status == StatusOpen, b.Status == StatusOpen; ao != bo {
		return ao
	}
	switch {
	case a.DueOn != nil && b.DueOn != nil:
		if !a.DueOn.Equal(*b.DueOn) {
			return a.DueOn.Before(*b.DueOn)
		}
	case a.DueOn != nil:
		return true
	case b.DueOn != nil:
		return false
	}
	if pa, pb := priorityRank(a.Priority), priorityRank(b.Priority); pa != pb {
		return pa < pb
	}
	return a.CreatedAt.After(b.CreatedAt)
}
