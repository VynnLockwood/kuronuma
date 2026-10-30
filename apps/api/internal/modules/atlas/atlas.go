// Package atlas is the memory module: tasks, to-dos and notes the user's assistant keeps for them, found
// again by keyword search plus filters. Categories come from the shared list kept by themis.
package atlas

import (
	"context"
	"time"
)

// OAuth scopes used by the tools.
const (
	ScopeRead  = "atlas:read"
	ScopeWrite = "atlas:write"
)

// Kinds of memory: a task is a work item, a todo a personal checklist item, a note anything to remember.
const (
	KindTask = "task"
	KindTodo = "todo"
	KindNote = "note"
)

// Statuses. completed_at is set when an item becomes done and cleared when it leaves done.
const (
	StatusOpen     = "open"
	StatusDone     = "done"
	StatusArchived = "archived"
)

// Priorities.
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
)

// Limits.
const (
	MaxTitleLen  = 200
	MaxBodyLen   = 4000
	MaxTags      = 10
	MaxTagLen    = 30
	MaxTerms     = 5
	MaxTermLen   = 50
	MaxLimit     = 200
	DefaultLimit = 50
)

// Memory is one stored item.
type Memory struct {
	ID, Kind, Title, Body, Status, Priority, Category string
	Tags                                              []string
	DueOn, CompletedAt                                *time.Time
	CreatedAt, UpdatedAt                              time.Time
}

// NewMemory is the validated input for a new item.
type NewMemory struct {
	Kind, Title, Body, Priority, Category string
	Tags                                  []string
	DueOn                                 *time.Time
}

// Update changes the non-nil fields of an item. SetDue true with a nil DueOn clears the due date; SetDue false
// leaves it alone.
type Update struct {
	Title, Body, Status, Priority, Category *string
	Tags                                    *[]string
	SetDue                                  bool
	DueOn                                   *time.Time
}

// Filter selects items. Every non-empty field must match. Terms must each appear (case-insensitively) as a
// substring of the title, body or a tag. DueBefore and DueAfter are inclusive; items without a due date never
// match them.
type Filter struct {
	Terms                       []string
	Kind, Status, Category, Tag string
	DueBefore, DueAfter         *time.Time
	Limit                       int
}

// Store persists memories. Every method is scoped to userID.
type Store interface {
	AddMemory(ctx context.Context, userID string, n NewMemory) (Memory, error)
	ListMemories(ctx context.Context, userID string, f Filter) ([]Memory, error)
	// UpdateMemory applies u and returns the item. found is false when the user has no item with that id.
	// completed_at follows the status: set on becoming done, cleared on leaving done, kept otherwise.
	UpdateMemory(ctx context.Context, userID, id string, u Update) (m Memory, found bool, err error)
	DeleteMemory(ctx context.Context, userID, id string) (bool, error)
	// CategoryNames returns the names in the user's category list, sorted (themis.Uncategorized is built in).
	CategoryNames(ctx context.Context, userID string) ([]string, error)
}
