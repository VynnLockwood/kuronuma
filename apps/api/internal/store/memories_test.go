package store

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VynnLockwood/kuronuma/apps/api/internal/core/themis"
	"github.com/VynnLockwood/kuronuma/apps/api/internal/modules/atlas"
	"github.com/VynnLockwood/kuronuma/apps/api/internal/modules/plutus"
)

// TestPostgresMemories needs a real database, so it only runs when TEST_DATABASE_URL is set
// (CI has no database service). It creates and removes its own users.
func TestPostgresMemories(t *testing.T) {
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
		sub := fmt.Sprintf("memories-test-%d", time.Now().UnixNano())
		var id string
		if err := db.pool.QueryRow(ctx,
			`INSERT INTO users (google_sub, email) VALUES ($1, $2) RETURNING id::text`, sub, sub+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, id) })
		return id
	}
	day := func(s string) *time.Time {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return &d
	}
	add := func(user string, n atlas.NewMemory) atlas.Memory {
		m, err := db.AddMemory(ctx, user, n)
		if err != nil {
			t.Fatalf("add %q: %v", n.Title, err)
		}
		return m
	}
	list := func(user string, f atlas.Filter) []atlas.Memory {
		ms, err := db.ListMemories(ctx, user, f)
		if err != nil {
			t.Fatal(err)
		}
		return ms
	}
	titles := func(ms []atlas.Memory) string {
		var out []string
		for _, m := range ms {
			out = append(out, m.Title)
		}
		return strings.Join(out, " | ")
	}

	alice, bob := newUser(), newUser()
	a := add(alice, atlas.NewMemory{Kind: "task", Title: "Write report", Body: "quarterly numbers", Priority: "high", Category: "work", Tags: []string{"q4", "report"}, DueOn: day("2026-11-05")})
	b := add(alice, atlas.NewMemory{Kind: "todo", Title: "Buy milk", Priority: "normal", Category: "uncategorized"})
	c := add(alice, atlas.NewMemory{Kind: "note", Title: "100% sure", Body: `a_b and c\d`, Priority: "low", Category: "uncategorized", Tags: []string{"misc"}, DueOn: day("2026-10-20")})
	d := add(alice, atlas.NewMemory{Kind: "task", Title: "ร้านอาหารครัวไทย", Priority: "normal", Category: "work"})
	add(alice, atlas.NewMemory{Kind: "note", Title: "aXb", Priority: "normal", Category: "uncategorized"})

	// Round trip and defaults.
	if a.ID == "" || a.Status != "open" || a.CompletedAt != nil || a.DueOn == nil || a.DueOn.Format("2006-01-02") != "2026-11-05" || len(a.Tags) != 2 {
		t.Fatalf("round trip: %+v", a)
	}
	if b.DueOn != nil || b.Tags == nil || len(b.Tags) != 0 {
		t.Fatalf("an item without a due date or tags: %+v", b)
	}

	// The list order: open first, due date ascending with undated last, then priority, then newest.
	if got := titles(list(alice, atlas.Filter{})); !strings.HasPrefix(got, "100% sure | Write report | ") {
		t.Fatalf("order: %s", got)
	}

	// Filters.
	cases := []struct {
		name string
		f    atlas.Filter
		want string
	}{
		{"kind", atlas.Filter{Kind: "note", Terms: []string{"sure"}}, "100% sure"},
		{"category", atlas.Filter{Category: "work"}, "Write report | ร้านอาหารครัวไทย"},
		{"tag", atlas.Filter{Tag: "q4"}, "Write report"},
		{"due before is inclusive and skips undated", atlas.Filter{DueBefore: day("2026-10-20")}, "100% sure"},
		{"due after is inclusive and skips undated", atlas.Filter{DueAfter: day("2026-11-05")}, "Write report"},
		{"term in the body", atlas.Filter{Terms: []string{"quarterly"}}, "Write report"},
		{"terms are ANDed", atlas.Filter{Terms: []string{"quarterly", "report"}}, "Write report"},
		{"ANDed terms that never meet", atlas.Filter{Terms: []string{"quarterly", "milk"}}, ""},
		{"a tag matches a term, any case", atlas.Filter{Terms: []string{"Q4"}}, "Write report"},
		{"Thai substring", atlas.Filter{Terms: []string{"อาหาร"}}, "ร้านอาหารครัวไทย"},
		{"percent is literal", atlas.Filter{Terms: []string{"100%"}}, "100% sure"},
		{"underscore is literal", atlas.Filter{Terms: []string{"a_b"}}, "100% sure"},
		{"backslash is literal", atlas.Filter{Terms: []string{`c\d`}}, "100% sure"},
	}
	for _, tc := range cases {
		if got := titles(list(alice, tc.f)); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
	if n := len(list(alice, atlas.Filter{Limit: 2})); n != 2 {
		t.Errorf("limit 2 returned %d", n)
	}

	// Update: completed_at follows the status.
	done, open, newTitle := "done", "open", "Buy oat milk"
	m, found, err := db.UpdateMemory(ctx, alice, b.ID, atlas.Update{Status: &done})
	if err != nil || !found || m.Status != "done" || m.CompletedAt == nil {
		t.Fatalf("complete: %+v found=%v err=%v", m, found, err)
	}
	completed := *m.CompletedAt
	m, _, err = db.UpdateMemory(ctx, alice, b.ID, atlas.Update{Title: &newTitle})
	if err != nil || m.CompletedAt == nil || !m.CompletedAt.Equal(completed) || m.Title != newTitle {
		t.Fatalf("editing a done item must keep completed_at: %+v err=%v", m, err)
	}
	m, _, err = db.UpdateMemory(ctx, alice, b.ID, atlas.Update{Status: &done})
	if err != nil || m.CompletedAt == nil || !m.CompletedAt.Equal(completed) {
		t.Fatalf("done to done must keep completed_at: %+v err=%v", m, err)
	}
	m, _, err = db.UpdateMemory(ctx, alice, b.ID, atlas.Update{Status: &open})
	if err != nil || m.CompletedAt != nil || m.Status != "open" {
		t.Fatalf("reopening must clear completed_at: %+v err=%v", m, err)
	}
	// Due date: clear, then set; tags replaced.
	m, _, err = db.UpdateMemory(ctx, alice, a.ID, atlas.Update{SetDue: true})
	if err != nil || m.DueOn != nil {
		t.Fatalf("clearing the due date: %+v err=%v", m, err)
	}
	m, _, err = db.UpdateMemory(ctx, alice, a.ID, atlas.Update{SetDue: true, DueOn: day("2026-12-01"), Tags: &[]string{"x"}})
	if err != nil || m.DueOn == nil || m.DueOn.Format("2006-01-02") != "2026-12-01" || len(m.Tags) != 1 || m.Tags[0] != "x" {
		t.Fatalf("setting the due date and tags: %+v err=%v", m, err)
	}

	// Another user can neither change nor delete nor see these items.
	if _, found, err := db.UpdateMemory(ctx, bob, a.ID, atlas.Update{Title: &newTitle}); err != nil || found {
		t.Fatalf("bob updated alice's item: found=%v err=%v", found, err)
	}
	if ok, err := db.DeleteMemory(ctx, bob, a.ID); err != nil || ok {
		t.Fatalf("bob deleted alice's item: ok=%v err=%v", ok, err)
	}
	if n := len(list(bob, atlas.Filter{})); n != 0 {
		t.Fatalf("bob sees %d of alice's items", n)
	}

	// Delete.
	if ok, err := db.DeleteMemory(ctx, alice, d.ID); err != nil || !ok {
		t.Fatalf("delete: ok=%v err=%v", ok, err)
	}
	if ok, err := db.DeleteMemory(ctx, alice, d.ID); err != nil || ok {
		t.Fatalf("second delete: ok=%v err=%v, want false", ok, err)
	}

	// A category rename and merge move memories, and use of a name counts them.
	if _, err := db.CreateCategory(ctx, alice, "work"); err != nil {
		t.Fatal(err)
	}
	cats, err := db.ListCategories(ctx, alice)
	if err != nil {
		t.Fatal(err)
	}
	for _, cat := range cats {
		for _, label := range []string{"spending", "payee_rules", "memory"} {
			if _, ok := cat.Uses[label]; !ok {
				t.Fatalf("category %q lacks the %q count: %+v", cat.Name, label, cat.Uses)
			}
		}
		if cat.Name == "work" && cat.Uses["memory"] != 1 {
			t.Fatalf("work should count one memory: %+v", cat)
		}
	}
	if updated, err := db.RenameCategory(ctx, alice, "work", "job"); err != nil || updated["memory"] != 1 {
		t.Fatalf("rename: updated=%v err=%v", updated, err)
	}
	if got := list(alice, atlas.Filter{Category: "job"}); len(got) != 1 || got[0].ID != a.ID {
		t.Fatalf("the item did not follow the rename: %v", titles(got))
	}
	if moved, err := db.MergeCategories(ctx, alice, []string{"job"}, themis.Uncategorized); err != nil || moved["memory"] != 1 {
		t.Fatalf("merge: moved=%v err=%v", moved, err)
	}
	if got := list(alice, atlas.Filter{Category: themis.Uncategorized, Terms: []string{"report"}}); len(got) != 1 {
		t.Fatalf("the item did not follow the merge: %v", titles(got))
	}

	// Two adds at once both succeed.
	before := len(list(alice, atlas.Filter{Limit: 200}))
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := db.AddMemory(ctx, alice, atlas.NewMemory{Kind: "todo", Title: "same item", Priority: "normal", Category: "uncategorized"}); err != nil {
				t.Errorf("concurrent add: %v", err)
			}
		}()
	}
	wg.Wait()
	if after := len(list(alice, atlas.Filter{Limit: 200})); after != before+2 {
		t.Fatalf("after two concurrent adds there are %d items, want %d", after, before+2)
	}
	_ = c
}

// TestPostgresCounts checks the per-user counts behind the caps. Like the tests above it needs TEST_DATABASE_URL.
func TestPostgresCounts(t *testing.T) {
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
		sub := fmt.Sprintf("counts-test-%d", time.Now().UnixNano())
		var id string
		if err := db.pool.QueryRow(ctx,
			`INSERT INTO users (google_sub, email) VALUES ($1, $2) RETURNING id::text`, sub, sub+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, id) })
		return id
	}
	a, b := newUser(), newUser()
	for i := 0; i < 2; i++ {
		if _, err := db.AddMemory(ctx, a, atlas.NewMemory{Kind: "note", Title: fmt.Sprintf("n%d", i), Priority: "normal", Category: themis.Uncategorized}); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := db.AddTransaction(ctx, a, plutus.NewTransaction{AmountMinor: 100, Currency: "THB", Merchant: "m", Category: themis.Uncategorized, OccurredOn: time.Now()}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		got  func() (int, error)
		want int
	}{
		{"memories a", func() (int, error) { return db.CountMemories(ctx, a) }, 2},
		{"memories b", func() (int, error) { return db.CountMemories(ctx, b) }, 0},
		{"transactions a", func() (int, error) { return db.CountTransactions(ctx, a) }, 1},
		{"transactions b", func() (int, error) { return db.CountTransactions(ctx, b) }, 0},
	} {
		if n, err := c.got(); err != nil || n != c.want {
			t.Errorf("%s: count = %d, %v; want %d", c.name, n, err, c.want)
		}
	}
}
