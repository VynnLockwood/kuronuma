package plutus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

const goodRow = `{"amount_minor":12900,"currency":"THB","merchant":"Grab","category":"transport","occurred_on":"2026-10-28","note":"","reference":""}`

func backupJSON(rows string) string {
	return `{"format":"nemosyne-backup","version":1,"transactions":[` + rows + `]}`
}

func newBackupModule() (*Module, *memStore) {
	st := &memStore{}
	return New(st, func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) }), st
}

func TestWriteBackup_Header(t *testing.T) {
	m, st := newBackupModule()
	ctx := context.Background()
	for _, n := range []NewTransaction{
		{AmountMinor: 12900, Currency: "THB", Merchant: "Grab", Category: "transport", OccurredOn: time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC), Reference: "R1"},
		{AmountMinor: 4550, Currency: "THB", Merchant: "7-Eleven", Category: "food", OccurredOn: time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC), Note: "snacks"},
	} {
		if _, _, err := st.AddTransaction(ctx, "u1", n); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := st.AddTransaction(ctx, "u2", NewTransaction{AmountMinor: 1, Currency: "USD", Merchant: "Other", Category: "x", OccurredOn: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := m.WriteBackup(ctx, &buf, "u1"); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if got["format"] != "nemosyne-backup" || got["version"] != float64(1) {
		t.Errorf("bad header: format=%v version=%v", got["format"], got["version"])
	}
	if _, err := time.Parse(time.RFC3339, got["exported_at"].(string)); err != nil {
		t.Errorf("exported_at is not RFC 3339: %v", err)
	}
	rows := got["transactions"].([]any)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want only u1's 2", len(rows))
	}
	for _, r := range rows {
		keys := r.(map[string]any)
		for _, forbidden := range []string{"id", "user_id", "user"} {
			if _, ok := keys[forbidden]; ok {
				t.Errorf("row exposes %q; the file must be portable", forbidden)
			}
		}
		if len(keys) != 7 {
			t.Errorf("row has %d keys, want the 7 documented fields: %v", len(keys), keys)
		}
	}
}

func TestWriteBackup_EmptyIsAnEmptyList(t *testing.T) {
	m, _ := newBackupModule()
	var buf bytes.Buffer
	if err := m.WriteBackup(context.Background(), &buf, "nobody"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"transactions": []`) {
		t.Errorf("want an empty JSON list, got %s", buf.String())
	}
}

func TestParseBackup_Valid(t *testing.T) {
	m, _ := newBackupModule()
	rows, err := m.parseBackup(strings.NewReader(backupJSON(goodRow+","+goodRow)), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Merchant != "Grab" || rows[0].AmountMinor != 12900 || rows[0].Category != "transport" {
		t.Errorf("unexpected rows: %+v", rows)
	}
}

func TestParseBackup_EmptyList(t *testing.T) {
	m, _ := newBackupModule()
	rows, err := m.parseBackup(strings.NewReader(backupJSON("")), 1<<20)
	if err != nil || len(rows) != 0 {
		t.Fatalf("rows=%v err=%v, want zero rows and no error", rows, err)
	}
}

func TestParseBackup_EmptyObjectWithHeader(t *testing.T) {
	m, _ := newBackupModule()
	rows, err := m.parseBackup(strings.NewReader(`{"format":"nemosyne-backup","version":1}`), 1<<20)
	if err != nil || len(rows) != 0 {
		t.Fatalf("rows=%v err=%v, want zero rows and no error", rows, err)
	}
}

func TestParseBackup_Rejects(t *testing.T) {
	m, _ := newBackupModule()
	withID := strings.Replace(goodRow, "}", `,"id":"x"}`, 1)
	withUser := strings.Replace(goodRow, "}", `,"user_id":"someone-else"}`, 1)
	cases := map[string]string{
		"empty input":       "",
		"not json":          "hello",
		"truncated":         backupJSON(goodRow)[:40],
		"csv":               "date,amount\n2026-10-28,450\n",
		"png":               "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR",
		"bom prefix":        "\xef\xbb\xbf" + backupJSON(goodRow),
		"bare object":       `{}`,
		"wrong format":      `{"format":"other","version":1,"transactions":[]}`,
		"newer version":     `{"format":"nemosyne-backup","version":2,"transactions":[]}`,
		"version zero":      `{"format":"nemosyne-backup","version":0,"transactions":[]}`,
		"unknown top field": `{"format":"nemosyne-backup","version":1,"transactions":[],"extra":1}`,
		"unknown row field": backupJSON(withID),
		"user id in row":    backupJSON(withUser),
		"trailing object":   backupJSON(goodRow) + "{}",
		"trailing garbage":  backupJSON(goodRow) + " x",
	}
	for name, in := range cases {
		rows, err := m.parseBackup(strings.NewReader(in), 1<<20)
		if !errors.Is(err, ErrBackup) {
			t.Errorf("%s: err=%v, want ErrBackup", name, err)
		}
		if rows != nil {
			t.Errorf("%s: must return no rows", name)
		}
	}
}

func TestParseBackup_TooLarge(t *testing.T) {
	m, _ := newBackupModule()
	in := backupJSON(goodRow)
	if _, err := m.parseBackup(strings.NewReader(in), int64(len(in))); err != nil {
		t.Fatalf("file exactly at the limit must pass: %v", err)
	}
	if _, err := m.parseBackup(strings.NewReader(in), int64(len(in))-1); !errors.Is(err, ErrBackup) {
		t.Fatalf("file one byte over the limit: err=%v, want ErrBackup", err)
	}
}

func TestParseBackup_BadRowsReportedTogether(t *testing.T) {
	m, _ := newBackupModule()
	badAmount := strings.Replace(goodRow, `"amount_minor":12900`, `"amount_minor":0`, 1)
	badCurrency := strings.Replace(strings.Replace(goodRow, `"THB"`, `"XX"`, 1), `"Grab"`, `"SecretMerchantName"`, 1)
	in := backupJSON(strings.Join([]string{goodRow, badAmount, goodRow, goodRow, badCurrency}, ","))

	rows, err := m.parseBackup(strings.NewReader(in), 1<<20)
	var re *RowsError
	if !errors.As(err, &re) {
		t.Fatalf("err=%v, want *RowsError", err)
	}
	if rows != nil {
		t.Error("must return no rows when any row is invalid")
	}
	if len(re.Rows) != 2 || re.Rows[0].Row != 2 || re.Rows[1].Row != 5 {
		t.Fatalf("rows reported: %+v, want rows 2 and 5 (1-based)", re.Rows)
	}
	for _, r := range re.Rows {
		if r.Reason == "" || strings.Contains(r.Reason, "SecretMerchantName") || strings.Contains(r.Reason, "XX") {
			t.Errorf("reason must be static text, never file content: %q", r.Reason)
		}
	}
}
