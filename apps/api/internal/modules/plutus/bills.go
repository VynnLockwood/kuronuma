package plutus

import (
	"fmt"
	"sort"
	"time"
)

// Bill is a fixed monthly bill. It counts in every month from StartMonth up
// to, but not including, EndedFrom (when set).
type Bill struct {
	ID          string
	Name        string
	AmountMinor int64
	Currency    string
	Category    string
	StartMonth  time.Time
	EndedFrom   *time.Time
}

// NewBill is the validated input for a new bill.
type NewBill struct {
	Name        string
	AmountMinor int64
	Currency    string
	Category    string
	StartMonth  time.Time
}

// BillPatch holds optional changes to a bill. ClearEnded removes EndedFrom.
type BillPatch struct {
	Name, Category        *string
	AmountMinor           *int64
	StartMonth, EndedFrom *time.Time
	ClearEnded            bool
}

// BillTotal is one currency's planned and paid figures for a month.
type BillTotal struct {
	Currency     string
	PlannedMinor int64
	PaidMinor    int64
	Paid, Unpaid int
}

// MonthStart returns the first day of t's month at midnight UTC.
func MonthStart(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// ParseMonth parses a YYYY-MM string into the first day of that month, UTC.
func ParseMonth(s string) (time.Time, error) {
	if len(s) != 7 {
		return time.Time{}, fmt.Errorf("month %q must be YYYY-MM", s)
	}
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("month %q must be YYYY-MM", s)
	}
	return MonthStart(t), nil
}

// Counts reports whether b counts in month: StartMonth <= month and
// (EndedFrom is nil or month < EndedFrom).
func Counts(b Bill, month time.Time) bool {
	month = MonthStart(month)
	if month.Before(MonthStart(b.StartMonth)) {
		return false
	}
	if b.EndedFrom != nil && !month.Before(MonthStart(*b.EndedFrom)) {
		return false
	}
	return true
}

// BillReference is the payment reference for a bill in a month:
// bill:<bill_id>:<YYYY-MM>.
func BillReference(billID string, month time.Time) string {
	return fmt.Sprintf("bill:%s:%s", billID, MonthStart(month).Format("2006-01"))
}

// Totals sums the bills that count in month per currency, sorted by currency.
// payments is keyed by payment reference; a bill is paid when its reference is
// present, and the paid amount is that transaction's amount.
func Totals(bills []Bill, month time.Time, payments map[string]Transaction) []BillTotal {
	byCurrency := map[string]*BillTotal{}
	for _, b := range bills {
		if !Counts(b, month) {
			continue
		}
		t := byCurrency[b.Currency]
		if t == nil {
			t = &BillTotal{Currency: b.Currency}
			byCurrency[b.Currency] = t
		}
		t.PlannedMinor += b.AmountMinor
		if p, ok := payments[BillReference(b.ID, month)]; ok {
			t.PaidMinor += p.AmountMinor
			t.Paid++
		} else {
			t.Unpaid++
		}
	}
	out := make([]BillTotal, 0, len(byCurrency))
	for _, t := range byCurrency {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Currency < out[j].Currency })
	return out
}
