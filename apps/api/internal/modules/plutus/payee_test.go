package plutus

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizePayee(t *testing.T) {
	cases := map[string]string{
		"  7-ELEVEN  Silom ": "7-eleven silom",
		"7​-eleven":          "7-eleven",
		"ร้านอาหาร":          "ร้านอาหาร",
		"​ ":                 "",
	}
	for in, want := range cases {
		if got := NormalizePayee(in); got != want {
			t.Errorf("NormalizePayee(%q) = %q, want %q", in, got, want)
		}
	}
	if got := NormalizePayee(strings.Repeat("ก", 150)); len([]rune(got)) != 100 {
		t.Errorf("a 150-character payee has %d characters after normalising, want 100", len([]rune(got)))
	}
}

func rule(norm string, updated int) PayeeRule {
	return PayeeRule{Payee: norm, PayeeNorm: norm, Category: "c", UpdatedAt: time.Unix(int64(updated), 0)}
}

func TestMatchRule(t *testing.T) {
	cases := []struct {
		name     string
		rules    []PayeeRule
		merchant string
		want     string // PayeeNorm of the winning rule, "" for no match
	}{
		{"case and spacing are ignored", []PayeeRule{rule("7-eleven", 1)}, "7-ELEVEN  Silom", "7-eleven"},
		{"no rule matches", []PayeeRule{rule("7-eleven", 1)}, "Grab", ""},
		{"no rules at all", nil, "Grab", ""},
		{"the longest match wins", []PayeeRule{rule("7-eleven", 1), rule("7-eleven silom", 1)}, "7-Eleven Silom", "7-eleven silom"},
		{"equal length: later update wins", []PayeeRule{rule("ab", 1), rule("bc", 2)}, "abc", "bc"},
		{"equal length: later update wins (reversed)", []PayeeRule{rule("ab", 2), rule("bc", 1)}, "abc", "ab"},
		{"Thai rule matches a Thai merchant", []PayeeRule{rule("ร้านอาหาร", 1)}, "ร้านอาหารครัวไทย", "ร้านอาหาร"},
		{"length counts characters, not bytes", []PayeeRule{rule("กขคงจ", 1), rule("abcdefgh", 1)}, "กขคงจ abcdefgh", "abcdefgh"},
		{"a short rule matches an unrelated merchant containing it", []PayeeRule{rule("go", 1)}, "Google", "go"},
		{"an empty rule never matches", []PayeeRule{rule("", 1)}, "Grab", ""},
	}
	for _, tc := range cases {
		got, ok := matchRule(tc.rules, tc.merchant)
		if tc.want == "" {
			if ok {
				t.Errorf("%s: matched %q, want no match", tc.name, got.PayeeNorm)
			}
			continue
		}
		if !ok || got.PayeeNorm != tc.want {
			t.Errorf("%s: got %q ok=%v, want %q", tc.name, got.PayeeNorm, ok, tc.want)
		}
	}
}
