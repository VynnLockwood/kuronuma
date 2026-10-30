package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/modules/plutus"
)

// TestPostgresImport needs a real database, so it only runs when TEST_DATABASE_URL is set
// (CI has no database service). It creates and removes its own user.
func TestPostgresImport(t *testing.T) {
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

	sub := fmt.Sprintf("import-test-%d", time.Now().UnixNano())
	var uid string
	if err := db.pool.QueryRow(ctx,
		`INSERT INTO users (google_sub, email) VALUES ($1, $2) RETURNING id::text`, sub, sub+"@example.test").Scan(&uid); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, uid) })

	count := func() int {
		var n int
		if err := db.pool.QueryRow(ctx, `SELECT count(*) FROM transactions WHERE user_id = $1::uuid`, uid).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	date := func(s string) time.Time {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	bus := plutus.NewTransaction{AmountMinor: 2000, Currency: "THB", Merchant: "Bus", Category: "transport", OccurredOn: date("2026-10-28")}
	grab := plutus.NewTransaction{AmountMinor: 12900, Currency: "THB", Merchant: "Grab", Category: "transport", OccurredOn: date("2026-10-28"), Reference: "R1"}
	file := []plutus.NewTransaction{bus, bus, grab}

	res, err := db.ImportTransactions(ctx, uid, file)
	if err != nil || res != (plutus.ImportResult{Added: 3}) || count() != 3 {
		t.Fatalf("first import: res=%+v err=%v rows=%d, want 3 added", res, err, count())
	}
	res, err = db.ImportTransactions(ctx, uid, file)
	if err != nil || res != (plutus.ImportResult{Skipped: 3}) || count() != 3 {
		t.Fatalf("re-import: res=%+v err=%v rows=%d, want 3 skipped and nothing added", res, err, count())
	}

	// A later row the database rejects (lowercase currency violates the CHECK) must undo the earlier insert.
	bad := []plutus.NewTransaction{
		{AmountMinor: 1, Currency: "THB", Merchant: "Fresh", Category: "x", OccurredOn: date("2026-10-01")},
		{AmountMinor: 1, Currency: "thb", Merchant: "Broken", Category: "x", OccurredOn: date("2026-10-01")},
	}
	if _, err := db.ImportTransactions(ctx, uid, bad); err == nil {
		t.Fatal("expected the database to reject the second row")
	}
	if n := count(); n != 3 {
		t.Fatalf("failed import left %d rows, want the original 3 (rollback)", n)
	}

	// Two imports of the same new rows at once must not both insert them.
	conc := []plutus.NewTransaction{
		{AmountMinor: 5, Currency: "THB", Merchant: "Conc1", Category: "x", OccurredOn: date("2026-10-02")},
		{AmountMinor: 6, Currency: "THB", Merchant: "Conc2", Category: "x", OccurredOn: date("2026-10-02")},
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := db.ImportTransactions(ctx, uid, conc)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent import: %v", err)
		}
	}
	if n := count(); n != 5 {
		t.Fatalf("after two concurrent imports there are %d rows, want 5 (3 + 2, not 3 + 4)", n)
	}
}
