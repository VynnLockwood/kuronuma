package plutus

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/hermes"
)

// Limits for payee rules.
const (
	maxPayeeLen         = 100
	minPayeeLen         = 2
	maxPayeeDescription = 200
	maxPayeeRules       = 500
)

// PayeeRule teaches Nemo what a payee is: new spending whose merchant contains PayeeNorm is filed under
// Category when the assistant gives none. Payee is what the user said; PayeeNorm is what is matched on.
type PayeeRule struct {
	ID          string
	Payee       string
	PayeeNorm   string
	Category    string
	Description string
	UpdatedAt   time.Time
}

// NormalizePayee is the one rule for payee text, for rules and for merchants alike: lower case, whitespace
// collapsed, invisible and control characters removed, at most 100 characters.
func NormalizePayee(s string) string { return strings.ToLower(hermes.Untrusted(s, maxPayeeLen)) }

// matchRule returns the rule that applies to merchant: among the rules whose PayeeNorm the normalised merchant
// contains, the longest (counted in characters, not bytes), then the most recently updated, then the
// smallest PayeeNorm so the result never depends on input order. A rule with an empty PayeeNorm never matches.
func matchRule(rules []PayeeRule, merchant string) (PayeeRule, bool) {
	mn := NormalizePayee(merchant)
	var best PayeeRule
	found := false
	for _, r := range rules {
		if r.PayeeNorm == "" || !strings.Contains(mn, r.PayeeNorm) {
			continue
		}
		if !found || better(r, best) {
			best, found = r, true
		}
	}
	return best, found
}

func better(a, b PayeeRule) bool {
	if la, lb := utf8.RuneCountInString(a.PayeeNorm), utf8.RuneCountInString(b.PayeeNorm); la != lb {
		return la > lb
	}
	if !a.UpdatedAt.Equal(b.UpdatedAt) {
		return a.UpdatedAt.After(b.UpdatedAt)
	}
	return a.PayeeNorm < b.PayeeNorm
}
