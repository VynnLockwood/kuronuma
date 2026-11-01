package plutus

import (
	"fmt"
	"strings"
	"testing"
)

var gentleOpeners = []string{"Sorry,", "Just so you know,"}

// Every fixed sentence Plutus writes is gentle, short, plain, free of user data, and still says what happened and
// what to do next.
func TestFixedMessagesAreGentleAndExact(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		facts []string
	}{
		{"reference exists", msgRefExists, []string{"already exists", "nothing new was saved"}},
		{"possible duplicate", msgPossibleDuplicate, []string{"already exists", "nothing was saved", "Ask the user", "allow_duplicate=true"}},
		{"rule applies later", msgRuleAppliesLater, []string{"new transactions only", "saved records are unchanged"}},
		{"transaction limit", fmt.Sprintf(msgTransactionLimit, 5), []string{"5 transactions", "Ask the user", "delete_transaction"}},
		{"payee rule limit", fmt.Sprintf(msgRuleLimit, 5), []string{"5 payee rules", "delete_payee_rule"}},
	}
	for _, c := range cases {
		ok := false
		for _, o := range gentleOpeners {
			ok = ok || strings.HasPrefix(c.text, o)
		}
		if !ok {
			t.Errorf("%s: %q does not start with one of %v", c.name, c.text, gentleOpeners)
		}
		for _, f := range c.facts {
			if !strings.Contains(c.text, f) {
				t.Errorf("%s: lost the fact %q in %q", c.name, f, c.text)
			}
		}
		if len(c.text) > 240 || !strings.HasSuffix(c.text, ".") {
			t.Errorf("%s: want at most 240 characters ending with a full stop, got %d: %q", c.name, len(c.text), c.text)
		}
		for i := 0; i < len(c.text); i++ {
			if c.text[i] > 0x7e {
				t.Errorf("%s: non-plain character at %d", c.name, i)
			}
		}
	}
}
