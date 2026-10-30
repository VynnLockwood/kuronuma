package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/themis"
	"github.com/VynnLockwood/nemosyne/apps/api/internal/modules/plutus"
)

// TestPostgresPayeeRules needs a real database, so it only runs when TEST_DATABASE_URL is set
// (CI has no database service). It creates and removes its own users.
func TestPostgresPayeeRules(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	newUser := func() string {
		sub := fmt.Sprintf("payee-test-%d", time.Now().UnixNano())
		var id string
		if err := db.pool.QueryRow(ctx,
			`INSERT INTO users (google_sub, email) VALUES ($1, $2) RETURNING id::text`, sub, sub+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, id) })
		return id
	}
	alice, bob := newUser(), newUser()

	first, created, err := db.SetPayeeRule(ctx, alice, plutus.PayeeRule{Payee: "7-Eleven", PayeeNorm: "7-eleven", Category: "food", Description: "convenience store"})
	if err != nil || !created || first.ID == "" {
		t.Fatalf("first set: rule=%+v created=%v err=%v", first, created, err)
	}
	time.Sleep(10 * time.Millisecond)
	second, created, err := db.SetPayeeRule(ctx, alice, plutus.PayeeRule{Payee: "7-ELEVEN", PayeeNorm: "7-eleven", Category: "bills", Description: "changed"})
	if err != nil || created {
		t.Fatalf("second set: created=%v err=%v, want an update", created, err)
	}
	if second.ID != first.ID || second.Payee != "7-ELEVEN" || second.Category != "bills" || second.Description != "changed" || !second.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("update did not refresh the rule: before=%+v after=%+v", first, second)
	}

	rules, err := db.PayeeRules(ctx, alice)
	if err != nil || len(rules) != 1 || rules[0].Category != "bills" {
		t.Fatalf("PayeeRules: %+v err=%v", rules, err)
	}
	if got, _ := db.PayeeRules(ctx, bob); len(got) != 0 {
		t.Fatalf("bob sees alice's rules: %+v", got)
	}
	if ok, err := db.DeletePayeeRule(ctx, bob, "7-eleven"); err != nil || ok {
		t.Fatalf("bob deleted alice's rule: ok=%v err=%v", ok, err)
	}

	// A rule's category follows a rename and a merge, and counts as use of the name.
	if _, err := db.CreateCategory(ctx, alice, "bills"); err != nil {
		t.Fatal(err)
	}
	cats, err := db.ListCategories(ctx, alice)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cats {
		if c.Name == "bills" && c.Uses["payee_rules"] != 1 {
			t.Fatalf("bills should count one rule: %+v", c)
		}
	}
	if updated, err := db.RenameCategory(ctx, alice, "bills", "utilities"); err != nil || updated["payee_rules"] != 1 {
		t.Fatalf("rename: updated=%v err=%v", updated, err)
	}
	if rules, _ := db.PayeeRules(ctx, alice); len(rules) != 1 || rules[0].Category != "utilities" {
		t.Fatalf("the rule did not follow the rename: %+v", rules)
	}
	if moved, err := db.MergeCategories(ctx, alice, []string{"utilities"}, themis.Uncategorized); err != nil || moved["payee_rules"] != 1 {
		t.Fatalf("merge: moved=%v err=%v", moved, err)
	}
	if rules, _ := db.PayeeRules(ctx, alice); len(rules) != 1 || rules[0].Category != themis.Uncategorized {
		t.Fatalf("the rule did not follow the merge: %+v", rules)
	}

	if ok, err := db.DeletePayeeRule(ctx, alice, "7-eleven"); err != nil || !ok {
		t.Fatalf("delete: ok=%v err=%v", ok, err)
	}
	if ok, err := db.DeletePayeeRule(ctx, alice, "7-eleven"); err != nil || ok {
		t.Fatalf("second delete: ok=%v err=%v, want false", ok, err)
	}
}
