package plutus

import (
	"strings"
	"testing"
	"time"
)

func TestBuildTransaction(t *testing.T) {
	m := New(&memStore{}, func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) })
	build := func(minor int64, currency, merchant, date, category string) (NewTransaction, error) {
		return m.buildTransaction(minor, currency, merchant, date, category, "", "")
	}

	rejected := []struct {
		name string
		call func() (NewTransaction, error)
	}{
		{"zero amount", func() (NewTransaction, error) { return build(0, "THB", "Grab", "2026-10-28", "") }},
		{"negative amount", func() (NewTransaction, error) { return build(-5, "THB", "Grab", "2026-10-28", "") }},
		{"amount above max", func() (NewTransaction, error) { return build(maxMinor+1, "THB", "Grab", "2026-10-28", "") }},
		{"two-letter currency", func() (NewTransaction, error) { return build(100, "TH", "Grab", "2026-10-28", "") }},
		{"four-letter currency", func() (NewTransaction, error) { return build(100, "THBB", "Grab", "2026-10-28", "") }},
		{"blank merchant", func() (NewTransaction, error) { return build(100, "THB", "  ", "2026-10-28", "") }},
		{"impossible date", func() (NewTransaction, error) { return build(100, "THB", "Grab", "2026-13-01", "") }},
		{"date three days ahead", func() (NewTransaction, error) { return build(100, "THB", "Grab", "2026-11-02", "") }},
		{"date before 1990", func() (NewTransaction, error) { return build(100, "THB", "Grab", "1989-12-31", "") }},
	}
	for _, tc := range rejected {
		if _, err := tc.call(); err == nil {
			t.Errorf("%s: expected an error", tc.name)
		}
	}

	got, err := build(maxMinor, "thb", "  Grab  ", "2026-10-31", "Food")
	if err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	if got.AmountMinor != maxMinor || got.Currency != "THB" || got.Merchant != "Grab" || got.Category != "food" {
		t.Errorf("unexpected normalisation: %+v", got)
	}
	if !got.OccurredOn.Equal(time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("tomorrow must be accepted, got date %v", got.OccurredOn)
	}

	if got, _ := build(100, "THB", "Grab", "2026-10-28", ""); got.Category != "uncategorized" {
		t.Errorf("empty category = %q, want uncategorized", got.Category)
	}
	if got, _ := build(100, "THB", strings.Repeat("ก", 150), "2026-10-28", ""); len([]rune(got.Merchant)) != 100 {
		t.Errorf("merchant has %d runes, want 100", len([]rune(got.Merchant)))
	}
}
