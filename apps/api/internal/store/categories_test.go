package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/themis"
)

// TestPostgresCategories needs a real database, so it only runs when TEST_DATABASE_URL is set
// (CI has no database service). It creates and removes its own users.
func TestPostgresCategories(t *testing.T) {
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
		sub := fmt.Sprintf("categories-test-%d", time.Now().UnixNano())
		var id string
		if err := db.pool.QueryRow(ctx,
			`INSERT INTO users (google_sub, email) VALUES ($1, $2) RETURNING id::text`, sub, sub+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, id) })
		return id
	}
	spend := func(user, category string, n int) {
		for i := 0; i < n; i++ {
			if _, err := db.pool.Exec(ctx, `
				INSERT INTO transactions (user_id, amount_minor, currency, merchant, category, occurred_on)
				VALUES ($1::uuid, 100, 'THB', 'shop', $2, '2026-10-01')`, user, category); err != nil {
				t.Fatal(err)
			}
		}
	}
	inCategory := func(user, category string) int {
		var n int
		if err := db.pool.QueryRow(ctx,
			`SELECT count(*) FROM transactions WHERE user_id = $1::uuid AND category = $2`, user, category).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	list := func(user string) map[string]themis.Category {
		cats, err := db.ListCategories(ctx, user)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]themis.Category{}
		for _, c := range cats {
			m[c.Name] = c
		}
		return m
	}

	alice, bob := newUser(), newUser()
	spend(alice, "food", 2)
	spend(alice, "travel", 1)
	spend(alice, "old", 1)
	spend(alice, "uncategorized", 1)
	if _, err := db.CreateCategory(ctx, bob, "food"); err != nil {
		t.Fatal(err)
	}
	spend(bob, "food", 1)

	// Nothing is listed yet: every in-use name shows as unlisted, with its count.
	got := list(alice)
	if c := got["food"]; c.Listed || c.Uses["spending"] != 2 {
		t.Fatalf("food before listing: %+v", c)
	}

	if created, err := db.CreateCategory(ctx, alice, "food"); err != nil || !created {
		t.Fatalf("create food: created=%v err=%v", created, err)
	}
	if created, err := db.CreateCategory(ctx, alice, "food"); err != nil || created {
		t.Fatalf("second create: created=%v err=%v, want created=false", created, err)
	}
	if _, err := db.CreateCategory(ctx, alice, "travel"); err != nil {
		t.Fatal(err)
	}
	if got := list(alice); !got["food"].Listed || !got["travel"].Listed || got["old"].Listed {
		t.Fatalf("after create: %+v", got)
	}

	// Rename moves the list entry and the records together.
	updated, err := db.RenameCategory(ctx, alice, "food", "meals")
	if err != nil || updated["spending"] != 2 {
		t.Fatalf("rename: updated=%v err=%v", updated, err)
	}
	if inCategory(alice, "food") != 0 || inCategory(alice, "meals") != 2 {
		t.Fatal("rename did not move the transactions")
	}
	if _, err := db.RenameCategory(ctx, alice, "meals", "travel"); !errors.Is(err, themis.ErrExists) {
		t.Fatalf("rename onto an existing name: err=%v, want ErrExists", err)
	}
	if _, err := db.RenameCategory(ctx, alice, "nope", "x"); !errors.Is(err, themis.ErrNotFound) {
		t.Fatalf("rename of an unknown name: err=%v, want ErrNotFound", err)
	}
	if inCategory(alice, "meals") != 2 {
		t.Fatal("a failed rename changed data")
	}

	// Merge folds a listed and an unlisted source into the target.
	moved, err := db.MergeCategories(ctx, alice, []string{"travel", "old"}, "meals")
	if err != nil || moved["spending"] != 2 {
		t.Fatalf("merge: moved=%v err=%v", moved, err)
	}
	if got := list(alice); got["travel"].Listed || got["old"].Name != "" || got["meals"].Uses["spending"] != 4 {
		t.Fatalf("after merge: %+v", got)
	}
	if _, err := db.MergeCategories(ctx, alice, []string{"meals"}, "nope"); !errors.Is(err, themis.ErrNotFound) {
		t.Fatalf("merge into an unknown target: err=%v, want ErrNotFound", err)
	}
	if inCategory(alice, "meals") != 4 {
		t.Fatal("a refused merge changed data")
	}

	// A merge with one unknown source is all or nothing.
	if _, err := db.CreateCategory(ctx, alice, "p"); err != nil {
		t.Fatal(err)
	}
	spend(alice, "p", 1)
	if _, err := db.MergeCategories(ctx, alice, []string{"p", "ghost"}, themis.Uncategorized); !errors.Is(err, themis.ErrNotFound) {
		t.Fatalf("merge with an unknown source: err=%v, want ErrNotFound", err)
	}
	if !list(alice)["p"].Listed || inCategory(alice, "p") != 1 {
		t.Fatal("the failed merge was not rolled back")
	}

	// Merging into uncategorized needs no row for the target.
	if moved, err := db.MergeCategories(ctx, alice, []string{"meals", "p"}, themis.Uncategorized); err != nil || moved["spending"] != 5 {
		t.Fatalf("merge into uncategorized: moved=%v err=%v", moved, err)
	}
	if inCategory(alice, "uncategorized") != 6 {
		t.Fatalf("uncategorized holds %d, want 6", inCategory(alice, "uncategorized"))
	}

	// Bob was never touched.
	if c := list(bob)["food"]; !c.Listed || c.Uses["spending"] != 1 {
		t.Fatalf("bob's category changed: %+v", c)
	}
}
