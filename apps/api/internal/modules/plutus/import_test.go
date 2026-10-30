package plutus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

func jrow(amount int64, merchant, date, ref string) string {
	return fmt.Sprintf(`{"amount_minor":%d,"currency":"THB","merchant":%q,"category":"food","occurred_on":%q,"note":"","reference":%q}`,
		amount, merchant, date, ref)
}

func fileOf(rows ...string) string { return backupJSON(strings.Join(rows, ",")) }

func doImport(t *testing.T, m *Module, user, content string) ImportResult {
	t.Helper()
	res, err := m.ImportBackup(context.Background(), user, strings.NewReader(content), 1<<20)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	return res
}

func seed(t *testing.T, st *memStore, user string, n NewTransaction) {
	t.Helper()
	if _, _, err := st.AddTransaction(context.Background(), user, n); err != nil {
		t.Fatal(err)
	}
}

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestPlanImport(t *testing.T) {
	grab := func(merchant, ref string) NewTransaction {
		return NewTransaction{AmountMinor: 12900, Currency: "THB", Merchant: merchant, OccurredOn: day("2026-10-28"), Reference: ref}
	}
	cases := []struct {
		name                  string
		stored                []NewTransaction
		rows                  []NewTransaction
		wantInsert, wantSkips int
	}{
		{"all new", nil, []NewTransaction{grab("Grab", ""), grab("Grab", "R1")}, 2, 0},
		{"stored reference skipped", []NewTransaction{grab("Old", "R1")}, []NewTransaction{grab("Grab", "R1")}, 0, 1},
		{"same reference twice in file", nil, []NewTransaction{grab("Grab", "R1"), grab("Grab", "R1")}, 1, 1},
		{"no-ref row matches stored no-ref triple, case-insensitive", []NewTransaction{grab("GRAB", "")}, []NewTransaction{grab("Grab", "")}, 0, 1},
		{"two identical no-ref rows into empty store are both kept", nil, []NewTransaction{grab("Grab", ""), grab("Grab", "")}, 2, 0},
		{"two identical no-ref rows, one already stored", []NewTransaction{grab("Grab", "")}, []NewTransaction{grab("Grab", ""), grab("Grab", "")}, 1, 1},
		{"two identical no-ref rows, both already stored", []NewTransaction{grab("Grab", ""), grab("Grab", "")}, []NewTransaction{grab("Grab", ""), grab("Grab", "")}, 0, 2},
		{"row with its own reference is not a duplicate of a stored no-ref triple", []NewTransaction{grab("Grab", "")}, []NewTransaction{grab("Grab", "R9")}, 1, 0},
		{"stored referenced row does not hide a no-ref row", []NewTransaction{grab("Grab", "R1")}, []NewTransaction{grab("Grab", "")}, 1, 0},
		{"ref row earlier in file does not hide a no-ref row", nil, []NewTransaction{grab("Grab", "R1"), grab("Grab", "")}, 2, 0},
	}
	for _, tc := range cases {
		idx := NewExistingIndex()
		for _, s := range tc.stored {
			idx.Add(s.AmountMinor, s.Currency, s.OccurredOn, s.Merchant, s.Reference)
		}
		insert, skipped := PlanImport(tc.rows, idx)
		if len(insert) != tc.wantInsert || skipped != tc.wantSkips {
			t.Errorf("%s: insert=%d skipped=%d, want %d and %d", tc.name, len(insert), skipped, tc.wantInsert, tc.wantSkips)
		}
	}
	if insert, skipped := PlanImport([]NewTransaction{grab("Grab", "")}, nil); len(insert) != 1 || skipped != 0 {
		t.Errorf("a nil index must behave like an empty one, got insert=%d skipped=%d", len(insert), skipped)
	}
}

func TestImport_AddsAndCounts(t *testing.T) {
	m, st := newBackupModule()
	res := doImport(t, m, "u1", fileOf(jrow(100, "A", "2026-10-01", ""), jrow(200, "B", "2026-10-02", ""), jrow(300, "C", "2026-10-03", "")))
	if res != (ImportResult{Added: 3}) || st.count("u1") != 3 {
		t.Fatalf("res=%+v stored=%d, want 3 added", res, st.count("u1"))
	}
}

func TestImport_Idempotent(t *testing.T) {
	m, st := newBackupModule()
	content := fileOf(jrow(100, "A", "2026-10-01", ""), jrow(200, "B", "2026-10-02", "R2"), jrow(300, "C", "2026-10-03", ""))
	doImport(t, m, "u1", content)
	res := doImport(t, m, "u1", content)
	if res != (ImportResult{Skipped: 3}) || st.count("u1") != 3 {
		t.Fatalf("second import: res=%+v stored=%d, want 0 added 3 skipped", res, st.count("u1"))
	}
}

func TestImport_SkipsExistingReference(t *testing.T) {
	m, st := newBackupModule()
	seed(t, st, "u1", NewTransaction{AmountMinor: 1, Currency: "THB", Merchant: "Old", Category: "x", OccurredOn: day("2026-09-01"), Reference: "R1"})
	res := doImport(t, m, "u1", fileOf(jrow(100, "A", "2026-10-01", "R1"), jrow(200, "B", "2026-10-02", "")))
	if res != (ImportResult{Added: 1, Skipped: 1}) {
		t.Fatalf("res=%+v", res)
	}
}

func TestImport_SkipsSameAmountDateMerchantIgnoringCase(t *testing.T) {
	m, st := newBackupModule()
	seed(t, st, "u1", NewTransaction{AmountMinor: 12900, Currency: "THB", Merchant: "GRAB", Category: "x", OccurredOn: day("2026-10-28")})
	res := doImport(t, m, "u1", fileOf(jrow(12900, "Grab", "2026-10-28", "")))
	if res != (ImportResult{Skipped: 1}) || st.count("u1") != 1 {
		t.Fatalf("res=%+v stored=%d", res, st.count("u1"))
	}
}

func TestImport_IdenticalRowsInFileAreKept(t *testing.T) {
	m, st := newBackupModule()
	content := fileOf(jrow(2000, "Bus", "2026-10-28", ""), jrow(2000, "Bus", "2026-10-28", ""))
	if res := doImport(t, m, "u1", content); res != (ImportResult{Added: 2}) {
		t.Fatalf("restore into empty store: %+v, want both kept", res)
	}
	if res := doImport(t, m, "u1", content); res != (ImportResult{Skipped: 2}) || st.count("u1") != 2 {
		t.Fatalf("re-import: %+v stored=%d, want nothing added", res, st.count("u1"))
	}
}

func TestImport_DoubleSubmitAddsNothingTheSecondTime(t *testing.T) {
	m, st := newBackupModule()
	content := fileOf(jrow(100, "A", "2026-10-01", ""), jrow(100, "A", "2026-10-01", ""), jrow(5, "B", "2026-10-02", "R5"))
	doImport(t, m, "u1", content)
	before := st.count("u1")
	if res := doImport(t, m, "u1", content); res.Added != 0 || st.count("u1") != before {
		t.Fatalf("second submit added rows: %+v", res)
	}
}

func TestImport_BadRowWritesNothing(t *testing.T) {
	m, st := newBackupModule()
	_, err := m.ImportBackup(context.Background(), "u1",
		strings.NewReader(fileOf(jrow(100, "A", "2026-10-01", ""), jrow(0, "bad", "2026-10-02", ""))), 1<<20)
	var re *RowsError
	if !errors.As(err, &re) || st.count("u1") != 0 {
		t.Fatalf("err=%v stored=%d, want *RowsError and nothing stored", err, st.count("u1"))
	}
}

func TestImport_BadFileWritesNothing(t *testing.T) {
	m, st := newBackupModule()
	_, err := m.ImportBackup(context.Background(), "u1", strings.NewReader("date,amount\n2026-10-28,450\n"), 1<<20)
	if !errors.Is(err, ErrBackup) || st.count("u1") != 0 {
		t.Fatalf("err=%v stored=%d", err, st.count("u1"))
	}
}

func TestImport_EmptyFileIsFine(t *testing.T) {
	m, _ := newBackupModule()
	if res := doImport(t, m, "u1", fileOf()); res != (ImportResult{}) {
		t.Fatalf("res=%+v", res)
	}
}

func TestImport_ScopedToUser(t *testing.T) {
	m, st := newBackupModule()
	content := fileOf(jrow(100, "A", "2026-10-01", "R1"), jrow(200, "B", "2026-10-02", ""))
	if res := doImport(t, m, "u1", content); res.Added != 2 {
		t.Fatalf("u1: %+v", res)
	}
	if st.count("u2") != 0 {
		t.Fatal("u1's import must not touch u2")
	}
	if res := doImport(t, m, "u2", content); res.Added != 2 || st.count("u2") != 2 {
		t.Fatalf("u2 must get every row regardless of u1's data: %+v", res)
	}
}

func TestImport_RoundTrip(t *testing.T) {
	src, srcStore := newBackupModule()
	for _, n := range []NewTransaction{
		{AmountMinor: 12900, Currency: "THB", Merchant: "Grab", Category: "transport", OccurredOn: day("2026-10-28"), Reference: "R1", Note: "airport"},
		{AmountMinor: 2000, Currency: "THB", Merchant: "Bus", Category: "transport", OccurredOn: day("2026-10-28")},
		{AmountMinor: 2000, Currency: "THB", Merchant: "Bus", Category: "transport", OccurredOn: day("2026-10-28")},
		{AmountMinor: 99, Currency: "USD", Merchant: "Cloud", Category: "bills", OccurredOn: day("2026-10-01"), Reference: "R2"},
	} {
		seed(t, srcStore, "u1", n)
	}
	var buf bytes.Buffer
	if err := src.WriteBackup(context.Background(), &buf, "u1"); err != nil {
		t.Fatal(err)
	}

	dst, dstStore := newBackupModule()
	if res := doImport(t, dst, "owner", buf.String()); res != (ImportResult{Added: 4}) {
		t.Fatalf("res=%+v, want all 4 restored", res)
	}
	fingerprint := func(st *memStore, user string) []string {
		txs, _ := st.ExportTransactions(context.Background(), user)
		var out []string
		for _, x := range txs {
			out = append(out, fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s", x.AmountMinor, x.Currency, x.Merchant, x.Category,
				x.OccurredOn.Format("2006-01-02"), x.Note, x.Reference))
		}
		sort.Strings(out)
		return out
	}
	if a, b := fingerprint(srcStore, "u1"), fingerprint(dstStore, "owner"); strings.Join(a, "\n") != strings.Join(b, "\n") {
		t.Fatalf("round trip changed data:\n%v\n%v", a, b)
	}
}

func TestImport_StoredReferencedRowDoesNotHideARowWithoutReference(t *testing.T) {
	// The merge must not depend on which record arrived first: a stored slip with a reference and a
	// file row without one are two different records, exactly as they are inside a single file.
	m, st := newBackupModule()
	seed(t, st, "u1", NewTransaction{AmountMinor: 12900, Currency: "THB", Merchant: "Grab", Category: "x", OccurredOn: day("2026-10-28"), Reference: "R1"})
	content := fileOf(jrow(12900, "Grab", "2026-10-28", "R1"), jrow(12900, "Grab", "2026-10-28", ""))
	if res := doImport(t, m, "u1", content); res != (ImportResult{Added: 1, Skipped: 1}) || st.count("u1") != 2 {
		t.Fatalf("res=%+v stored=%d, want the no-reference row added and the R1 row skipped", res, st.count("u1"))
	}
	if res := doImport(t, m, "u1", content); res != (ImportResult{Skipped: 2}) || st.count("u1") != 2 {
		t.Fatalf("re-import: res=%+v stored=%d, want nothing added", res, st.count("u1"))
	}
}
