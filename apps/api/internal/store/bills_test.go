package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/VynnLockwood/kuronuma/apps/api/internal/modules/plutus"
)

// billsTestDB opens the test database, migrates it, and returns a helper that creates users that are
// removed when the test ends. It skips when TEST_DATABASE_URL is not set.
func billsTestDB(t *testing.T) (*Postgres, func() string) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	newUser := func() string {
		sub := fmt.Sprintf("bills-test-%d", time.Now().UnixNano())
		var id string
		if err := db.pool.QueryRow(ctx,
			`INSERT INTO users (google_sub, email) VALUES ($1, $2) RETURNING id::text`, sub, sub+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, id) })
		return id
	}
	return db, newUser
}

func billMonth(y int, m time.Month) time.Time { return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC) }

func TestPostgresBills(t *testing.T) {
	db, newUser := billsTestDB(t)
	ctx := context.Background()
	alice, bob := newUser(), newUser()
	jan, feb, mar := billMonth(2026, time.January), billMonth(2026, time.February), billMonth(2026, time.March)

	rent, err := db.AddBill(ctx, alice, plutus.NewBill{Name: "Rent", AmountMinor: 1200000, Currency: "THB", Category: "housing", StartMonth: jan})
	if err != nil || rent.ID == "" || rent.EndedFrom != nil || !rent.StartMonth.Equal(jan) {
		t.Fatalf("add: %+v err=%v", rent, err)
	}
	if _, err := db.AddBill(ctx, alice, plutus.NewBill{Name: "Internet", AmountMinor: 59900, Currency: "THB", Category: "bills", StartMonth: jan}); err != nil {
		t.Fatal(err)
	}

	got, ok, err := db.GetBill(ctx, alice, rent.ID)
	if err != nil || !ok || got.Name != "Rent" || got.AmountMinor != 1200000 || got.Currency != "THB" || got.Category != "housing" || !got.StartMonth.Equal(jan) {
		t.Fatalf("get: %+v ok=%v err=%v", got, ok, err)
	}
	list, err := db.ListBills(ctx, alice)
	if err != nil || len(list) != 2 || list[0].Name != "Internet" || list[1].Name != "Rent" {
		t.Fatalf("list should be ordered by name: %+v err=%v", list, err)
	}
	if n, err := db.CountBills(ctx, alice); err != nil || n != 2 {
		t.Fatalf("count: %d err=%v", n, err)
	}

	// A payment for January, then a price change: the payment is untouched.
	pay, created, err := db.AddTransaction(ctx, alice, plutus.NewTransaction{
		AmountMinor: 1200000, Currency: "THB", Merchant: "Rent", Category: "housing",
		OccurredOn: jan, Reference: plutus.BillReference(rent.ID, jan)})
	if err != nil || !created {
		t.Fatalf("payment: %+v created=%v err=%v", pay, created, err)
	}
	newAmount := int64(1300000)
	upd, ok, err := db.UpdateBill(ctx, alice, rent.ID, plutus.BillPatch{AmountMinor: &newAmount})
	if err != nil || !ok || upd.AmountMinor != 1300000 || upd.Name != "Rent" || upd.EndedFrom != nil {
		t.Fatalf("update amount: %+v ok=%v err=%v", upd, ok, err)
	}
	if tx, found, _ := db.TransactionByReference(ctx, alice, plutus.BillReference(rent.ID, jan)); !found || tx.AmountMinor != 1200000 {
		t.Fatalf("past payment changed: %+v found=%v", tx, found)
	}

	// End the bill, then clear the end.
	ended, ok, err := db.UpdateBill(ctx, alice, rent.ID, plutus.BillPatch{EndedFrom: &mar})
	if err != nil || !ok || ended.EndedFrom == nil || !ended.EndedFrom.Equal(mar) {
		t.Fatalf("end: %+v ok=%v err=%v", ended, ok, err)
	}
	cleared, ok, err := db.UpdateBill(ctx, alice, rent.ID, plutus.BillPatch{ClearEnded: true})
	if err != nil || !ok || cleared.EndedFrom != nil {
		t.Fatalf("clear ended: %+v ok=%v err=%v", cleared, ok, err)
	}
	// ended_from before start_month violates the table check, so it is an error, not a change.
	if _, _, err := db.UpdateBill(ctx, alice, rent.ID, plutus.BillPatch{EndedFrom: ptrTime(billMonth(2025, time.December))}); err == nil {
		t.Fatal("ended_from before start_month should be rejected")
	}

	// Payments: only the requested references come back; HasPayments follows the reference prefix.
	if _, _, err := db.AddTransaction(ctx, alice, plutus.NewTransaction{
		AmountMinor: 1300000, Currency: "THB", Merchant: "Rent", Category: "housing",
		OccurredOn: feb, Reference: plutus.BillReference(rent.ID, feb)}); err != nil {
		t.Fatal(err)
	}
	if has, err := db.HasPayments(ctx, alice, rent.ID); err != nil || !has {
		t.Fatalf("HasPayments(rent) = %v err=%v, want true", has, err)
	}
	var internetID string
	for _, b := range list {
		if b.Name == "Internet" {
			internetID = b.ID
		}
	}
	if has, err := db.HasPayments(ctx, alice, internetID); err != nil || has {
		t.Fatalf("HasPayments(internet) = %v err=%v, want false", has, err)
	}
	if has, _ := db.HasPayments(ctx, bob, rent.ID); has {
		t.Fatal("bob sees alice's payments")
	}
	if m, err := db.PaymentsByReference(ctx, alice, nil); err != nil || len(m) != 0 {
		t.Fatalf("empty refs: %v err=%v", m, err)
	}
	m, err := db.PaymentsByReference(ctx, alice, []string{plutus.BillReference(rent.ID, jan), plutus.BillReference(rent.ID, mar)})
	if err != nil || len(m) != 1 || m[plutus.BillReference(rent.ID, jan)].AmountMinor != 1200000 {
		t.Fatalf("PaymentsByReference should return only the January payment: %v err=%v", m, err)
	}
	if m, _ := db.PaymentsByReference(ctx, bob, []string{plutus.BillReference(rent.ID, jan)}); len(m) != 0 {
		t.Fatalf("bob sees alice's payment: %v", m)
	}

	// Isolation: bob cannot read, change or delete alice's bill.
	if _, ok, err := db.GetBill(ctx, bob, rent.ID); err != nil || ok {
		t.Fatalf("bob got alice's bill: ok=%v err=%v", ok, err)
	}
	if _, ok, err := db.UpdateBill(ctx, bob, rent.ID, plutus.BillPatch{AmountMinor: &newAmount}); err != nil || ok {
		t.Fatalf("bob updated alice's bill: ok=%v err=%v", ok, err)
	}
	if ok, err := db.DeleteBill(ctx, bob, rent.ID); err != nil || ok {
		t.Fatalf("bob deleted alice's bill: ok=%v err=%v", ok, err)
	}
	if l, _ := db.ListBills(ctx, bob); len(l) != 0 {
		t.Fatalf("bob lists alice's bills: %+v", l)
	}
	if n, _ := db.CountBills(ctx, bob); n != 0 {
		t.Fatalf("bob counts alice's bills: %d", n)
	}

	// Delete the bill that has no payments (the refusal for paid bills lives in the tool layer).
	if ok, err := db.DeleteBill(ctx, alice, internetID); err != nil || !ok {
		t.Fatalf("delete: ok=%v err=%v", ok, err)
	}
	if ok, err := db.DeleteBill(ctx, alice, internetID); err != nil || ok {
		t.Fatalf("second delete: ok=%v err=%v, want false", ok, err)
	}
	if _, ok, err := db.GetBill(ctx, alice, internetID); err != nil || ok {
		t.Fatalf("deleted bill still found: ok=%v err=%v", ok, err)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestPostgresBillsFollowCategoryRename(t *testing.T) {
	db, newUser := billsTestDB(t)
	ctx := context.Background()
	alice := newUser()
	jan := billMonth(2026, time.January)

	if _, err := db.CreateCategory(ctx, alice, "food"); err != nil {
		t.Fatal(err)
	}
	b, err := db.AddBill(ctx, alice, plutus.NewBill{Name: "Meal plan", AmountMinor: 100000, Currency: "THB", Category: "food", StartMonth: jan})
	if err != nil {
		t.Fatal(err)
	}
	ref := plutus.BillReference(b.ID, jan)
	if _, created, err := db.AddTransaction(ctx, alice, plutus.NewTransaction{
		AmountMinor: 100000, Currency: "THB", Merchant: "Meal plan", Category: "food", OccurredOn: jan, Reference: ref}); err != nil || !created {
		t.Fatalf("payment: created=%v err=%v", created, err)
	}

	cats, err := db.ListCategories(ctx, alice)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cats {
		if c.Name == "food" && (c.Uses["bills"] != 1 || c.Uses["spending"] != 1) {
			t.Fatalf("food should count one bill and one payment: %+v", c)
		}
	}
	updated, err := db.RenameCategory(ctx, alice, "food", "groceries")
	if err != nil || updated["bills"] != 1 || updated["spending"] != 1 {
		t.Fatalf("rename: updated=%v err=%v", updated, err)
	}
	got, ok, err := db.GetBill(ctx, alice, b.ID)
	if err != nil || !ok || got.Category != "groceries" {
		t.Fatalf("the bill did not follow the rename: %+v ok=%v err=%v", got, ok, err)
	}
	tx, found, err := db.TransactionByReference(ctx, alice, ref)
	if err != nil || !found || tx.Category != got.Category {
		t.Fatalf("the payment category should match the bill: tx=%+v bill=%+v err=%v", tx, got, err)
	}
}
