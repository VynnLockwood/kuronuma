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

// PlanImport decides which rows to insert when merging a backup into existing records.
//
//   - A row with a reference is skipped only when that reference already exists (existingRefs)
//     or appeared earlier in the same file. It is never treated as a duplicate of another record
//     just because amount, date and merchant match: a different slip can look the same.
//   - A row without a reference is matched on DuplicateKey. existingCounts says how many records
//     with that key are already stored; a file holding N identical rows adds max(0, N-existing).
//     So two real identical payments survive a restore, and importing the same file again adds nothing.
//
// existingRefs only needs the references that occur in rows.
func PlanImport(rows []NewTransaction, existingCounts map[string]int, existingRefs map[string]bool) (insert []NewTransaction, skipped int) {
	seenKeys := map[string]int{}
	seenRefs := map[string]bool{}
	for _, r := range rows {
		if r.Reference != "" {
			if existingRefs[r.Reference] || seenRefs[r.Reference] {
				skipped++
				continue
			}
			seenRefs[r.Reference] = true
			insert = append(insert, r)
			continue
		}
		k := DuplicateKey(r.AmountMinor, r.Currency, r.OccurredOn, r.Merchant)
		seenKeys[k]++
		if existingCounts[k] >= seenKeys[k] {
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
	return m.store.ImportTransactions(ctx, userID, rows)
}
