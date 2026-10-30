package plutus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/hermes"
)

// Backup file identity. The file carries no user id and no database ids, so it moves between installs.
const (
	BackupFormat  = "nemosyne-backup"
	BackupVersion = 1
)

// ErrBackup means the input is not a Nemosyne backup this version can read. The wrapped text is a
// fixed description; it never contains file content.
var ErrBackup = errors.New("not a valid Nemosyne backup")

// BackupRow is one transaction in a backup file.
type BackupRow struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
	Merchant    string `json:"merchant"`
	Category    string `json:"category"`
	OccurredOn  string `json:"occurred_on"`
	Note        string `json:"note"`
	Reference   string `json:"reference"`
}

type backupFile struct {
	Format       string      `json:"format"`
	Version      int         `json:"version"`
	ExportedAt   time.Time   `json:"exported_at"`
	Transactions []BackupRow `json:"transactions"`
}

// RowError names an invalid row (1-based) and why. Reason is fixed text, never file content.
type RowError struct {
	Row    int
	Reason string
}

// RowsError reports every invalid row of a backup at once.
type RowsError struct{ Rows []RowError }

func (e *RowsError) Error() string { return fmt.Sprintf("%d invalid row(s) in backup", len(e.Rows)) }

func backupErr(reason string) error { return fmt.Errorf("%w: %s", ErrBackup, reason) }

// WriteBackup writes all of the user's transactions as a backup file.
func (m *Module) WriteBackup(ctx context.Context, w io.Writer, userID string) error {
	txs, err := m.store.ExportTransactions(ctx, userID)
	if err != nil {
		return err
	}
	out := backupFile{
		Format:       BackupFormat,
		Version:      BackupVersion,
		ExportedAt:   m.now().UTC(),
		Transactions: make([]BackupRow, 0, len(txs)),
	}
	for _, t := range txs {
		out.Transactions = append(out.Transactions, BackupRow{
			AmountMinor: t.AmountMinor,
			Currency:    t.Currency,
			Merchant:    t.Merchant,
			Category:    t.Category,
			OccurredOn:  t.OccurredOn.Format(dateLayout),
			Note:        t.Note,
			Reference:   t.Reference,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// parseBackup reads and fully validates a backup. It returns either every row ready to store, an
// ErrBackup, or a *RowsError; a file with any problem yields no rows. A read error from r is returned as is.
func (m *Module) parseBackup(r io.Reader, maxBytes int64) ([]NewTransaction, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, backupErr("file is too large")
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f backupFile
	if err := dec.Decode(&f); err != nil {
		return nil, backupErr("file is not a backup document")
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, backupErr("unexpected data after the backup document")
	}
	if f.Format != BackupFormat {
		return nil, backupErr("unknown file format")
	}
	if f.Version < 1 || f.Version > BackupVersion {
		return nil, backupErr("unsupported backup version")
	}

	var rows []NewTransaction
	var bad []RowError
	for i, row := range f.Transactions {
		n, err := m.buildTransaction(row.AmountMinor, row.Currency, row.Merchant, row.OccurredOn, row.Category, row.Note, row.Reference)
		if err != nil {
			reason := "invalid row"
			var te hermes.ToolError
			if errors.As(err, &te) {
				reason = te.Msg
			}
			bad = append(bad, RowError{Row: i + 1, Reason: reason})
			continue
		}
		rows = append(rows, n)
	}
	if len(bad) > 0 {
		return nil, &RowsError{Rows: bad}
	}
	return rows, nil
}
