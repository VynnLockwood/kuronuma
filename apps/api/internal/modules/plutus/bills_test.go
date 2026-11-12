package plutus

import (
	"testing"
	"time"
)

func m(y int, mo time.Month) time.Time {
	return time.Date(y, mo, 1, 0, 0, 0, 0, time.UTC)
}

func ptr[T any](v T) *T { return &v }

func TestCountsMonthEdges(t *testing.T) {
	// Starts in Nov 2026, ends from Jan 2027: counts Nov and Dec, not Jan.
	b := Bill{StartMonth: m(2026, 11), EndedFrom: ptr(m(2027, 1))}
	cases := []struct {
		month time.Time
		want  bool
	}{
		{m(2026, 10), false},
		{m(2026, 11), true},
		{m(2026, 12), true},
		{m(2027, 1), false},
	}
	for _, c := range cases {
		if got := Counts(b, c.month); got != c.want {
			t.Errorf("Counts(%s) = %v, want %v", c.month.Format("2006-01"), got, c.want)
		}
	}

	// Leap year: a bill starting 2028-02 counts in 2028-02 and not 2028-01.
	leap := Bill{StartMonth: m(2028, 2)}
	if !Counts(leap, m(2028, 2)) {
		t.Error("bill starting 2028-02 should count in 2028-02")
	}
	if Counts(leap, m(2028, 1)) {
		t.Error("bill starting 2028-02 should not count in 2028-01")
	}

	// Year roll: ended from 2027-01 still counts in 2026-12.
	rolled := Bill{StartMonth: m(2026, 1), EndedFrom: ptr(m(2027, 1))}
	if !Counts(rolled, m(2026, 12)) {
		t.Error("bill ended from 2027-01 should count in 2026-12")
	}
	if Counts(rolled, m(2027, 1)) {
		t.Error("bill ended from 2027-01 should not count in 2027-01")
	}
}

func TestParseMonthAcceptsYearMonth(t *testing.T) {
	got, err := ParseMonth("2026-11")
	if err != nil {
		t.Fatalf("ParseMonth: %v", err)
	}
	if want := m(2026, 11); !got.Equal(want) {
		t.Errorf("ParseMonth = %v, want %v", got, want)
	}
}

func TestParseMonthRejectsOtherForms(t *testing.T) {
	for _, s := range []string{"2026-11-05", "11/2026", "2026-13", ""} {
		if _, err := ParseMonth(s); err == nil {
			t.Errorf("ParseMonth(%q) succeeded, want an error", s)
		}
	}
}

func TestBillReference(t *testing.T) {
	got := BillReference("3f2a", m(2026, 11))
	if want := "bill:3f2a:2026-11"; got != want {
		t.Errorf("BillReference = %q, want %q", got, want)
	}
}

func TestTotalsPerCurrencyAndPaid(t *testing.T) {
	nov := m(2026, 11)
	bills := []Bill{
		{ID: "usd1", AmountMinor: 999, Currency: "USD", StartMonth: m(2026, 1)},
		{ID: "rent", AmountMinor: 1200000, Currency: "THB", StartMonth: m(2026, 1)},
		{ID: "mobile", AmountMinor: 39900, Currency: "THB", StartMonth: m(2026, 1)},
		{ID: "old", AmountMinor: 50000, Currency: "THB", StartMonth: m(2026, 1), EndedFrom: ptr(m(2026, 10))},
	}
	payments := map[string]Transaction{
		BillReference("mobile", nov): {AmountMinor: 35000, Currency: "THB"},
	}

	got := Totals(bills, nov, payments)
	want := []BillTotal{
		{Currency: "THB", PlannedMinor: 1239900, PaidMinor: 35000, Paid: 1, Unpaid: 1},
		{Currency: "USD", PlannedMinor: 999, PaidMinor: 0, Paid: 0, Unpaid: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("Totals = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Totals[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestTotalsEmpty(t *testing.T) {
	got := Totals(nil, m(2026, 11), nil)
	if got == nil || len(got) != 0 {
		t.Errorf("Totals(nil) = %#v, want empty non-nil slice", got)
	}
}
