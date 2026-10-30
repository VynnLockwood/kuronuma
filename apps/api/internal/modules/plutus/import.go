package plutus

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// ImportResult counts what a merge import did.
type ImportResult struct{ Added, Skipped int }

// DuplicateKey identifies "the same payment" when merging: amount, currency, date and merchant
// (case-insensitive). The merchant comes last so the key is unambiguous.
func DuplicateKey(amountMinor int64, currency string, occurredOn time.Time, merchant string) string {
	return fmt.Sprintf("%d|%s|%s|%s", amountMinor, currency, occurredOn.Format(dateLayout), strings.ToLower(merchant))
}

// ExistingIndex summarises what a user already has stored, for PlanImport. Feed it every stored
// record with Add.
type ExistingIndex struct {
	counts map[string]int
	refs   map[string]bool
}

// NewExistingIndex returns an empty index.
func NewExistingIndex() *ExistingIndex {
	return &ExistingIndex{counts: map[string]int{}, refs: map[string]bool{}}
}

// Add records one stored record. A record with a reference is known only by that reference; it never
// counts toward the amount/date/merchant match used for rows without one. Both stores build their
// index through this method, so they cannot disagree about what "already stored" means.
func (x *ExistingIndex) Add(amountMinor int64, currency string, occurredOn time.Time, merchant, reference string) {
	if reference != "" {
		x.refs[reference] = true
		return
	}
	x.counts[DuplicateKey(amountMinor, currency, occurredOn, merchant)]++
}

// PlanImport decides which rows to insert when merging a backup into existing records.
//
// Records with a reference and records without one are two separate worlds, inside the file and
// against what is stored, so the result never depends on which of them arrived first:
//
//   - A row with a reference is skipped only when that reference is already stored or appeared
//     earlier in the same file. It is never a duplicate of another record just because amount, date
//     and merchant match: a different slip can look the same.
//   - A row without a reference is matched on DuplicateKey against stored records that also have no
//     reference; a file holding N identical rows adds max(0, N-stored). So two real identical
//     payments survive a restore, and importing the same file again adds nothing.
func PlanImport(rows []NewTransaction, existing *ExistingIndex) (insert []NewTransaction, skipped int) {
	if existing == nil {
		existing = NewExistingIndex()
	}
	seenKeys := map[string]int{}
	seenRefs := map[string]bool{}
	for _, r := range rows {
		if r.Reference != "" {
			if existing.refs[r.Reference] || seenRefs[r.Reference] {
				skipped++
				continue
			}
			seenRefs[r.Reference] = true
			insert = append(insert, r)
			continue
		}
		k := DuplicateKey(r.AmountMinor, r.Currency, r.OccurredOn, r.Merchant)
		seenKeys[k]++
		if existing.counts[k] >= seenKeys[k] {
			skipped++
			continue
		}
		insert = append(insert, r)
	}
	return insert, skipped
}

// ImportBackup merges a backup file into the user's records. A file with any problem is rejected
// as a whole and nothing is written. Nothing is ever deleted.
func (m *Module) ImportBackup(ctx context.Context, userID string, r io.Reader, maxBytes int64) (ImportResult, error) {
	rows, err := m.parseBackup(r, maxBytes)
	if err != nil {
		return ImportResult{}, err
	}
	if len(rows) == 0 {
		return ImportResult{}, nil
	}
	m.importMu.Lock()
	defer m.importMu.Unlock()
	return m.store.ImportTransactions(ctx, userID, rows)
}
