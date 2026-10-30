// Package themis keeps the user's category list. Spending (and later memory items) refer to a category by
// name; this package owns what names are valid and the tools to list, create, rename and merge them.
package themis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/hermes"
)

// OAuth scopes used by the tools below.
const (
	ScopeRead  = "categories:read"
	ScopeWrite = "categories:write"
)

// Uncategorized is always a valid category. It is built in: it has no row, and cannot be created,
// renamed or merged away.
const Uncategorized = "uncategorized"

const maxMergeSources = 10

// maxCategoriesInMessage caps how many of the user's categories a refusal lists.
const maxCategoriesInMessage = 50

// ErrExists and ErrNotFound are returned by a Store; the module turns them into messages for the assistant.
var (
	ErrExists   = errors.New("themis: category already exists")
	ErrNotFound = errors.New("themis: category not found")
)

// Category is one name with how much of the user's data uses it. Uses is keyed by module label ("spending",
// "payee_rules", "memory"); the Postgres store includes every label, zero when unused. Listed is false for a
// name that is in use but missing from the list.
type Category struct {
	Name   string
	Listed bool
	Uses   map[string]int
}

// Store persists the list. Every method is scoped to userID.
type Store interface {
	// ListCategories returns every listed name plus every in-use name missing from the list
	// (Listed false). Uncategorized need not appear.
	ListCategories(ctx context.Context, userID string) ([]Category, error)
	// CreateCategory adds name; created is false when it was already listed.
	CreateCategory(ctx context.Context, userID, name string) (created bool, err error)
	// RenameCategory renames a listed category and every record using it, atomically.
	// ErrNotFound if from is not listed; ErrExists if to is listed. updated counts records moved per module.
	RenameCategory(ctx context.Context, userID, from, to string) (updated map[string]int, err error)
	// MergeCategories moves every record from the sources to target and removes the sources from the
	// list, atomically. ErrNotFound if target is neither listed nor Uncategorized, or a source is
	// neither listed nor in use. moved counts records moved per module.
	MergeCategories(ctx context.Context, userID string, sources []string, target string) (moved map[string]int, err error)
}

// NormalizeName is the one rule for category names: lower case, whitespace collapsed, invisible and
// control characters removed, at most 40 characters.
func NormalizeName(s string) string { return strings.ToLower(hermes.Untrusted(s, 40)) }

// Check refuses a category that is not in the user's list (names). Uncategorized is always valid. The
// error is written for the assistant: it names the category, lists the user's categories (capped), and says
// how to recover: ask the user, tell them they can reconnect (an assistant without categories:write cannot
// create one), or omit the category. Every module that stores a category calls this.
func Check(category string, names []string) error {
	if category == Uncategorized {
		return nil
	}
	for _, n := range names {
		if n == category {
			return nil
		}
	}
	list := "They have no categories yet."
	if len(names) > 0 {
		shown := names
		if len(shown) > maxCategoriesInMessage {
			shown = shown[:maxCategoriesInMessage]
		}
		clean := make([]string, len(shown))
		for i, n := range shown {
			clean[i] = hermes.Untrusted(n, 40)
		}
		list = "Their categories: " + strings.Join(clean, ", ")
		if more := len(names) - len(shown); more > 0 {
			list += fmt.Sprintf(" and %d more", more)
		}
		list += "."
	}
	return hermes.Errorf("Category %q is not in the user's list. %s Ask the user whether to create it (create_category) or to choose one of these. "+
		"If create_category is not available to you, tell the user they can reconnect Nemosyne to manage categories. "+
		"If you cannot ask, omit the category to file it under uncategorized.", category, list)
}

// Module wires the category tools into the MCP server.
type Module struct{ store Store }

// New creates the module.
func New(store Store) *Module { return &Module{store: store} }

// Register adds the category tools to s.
func (m *Module) Register(s *hermes.Server) {
	s.AddTool(hermes.Tool{
		Name:  "list_categories",
		Title: "List categories",
		Description: "List the user's categories and how many records use each. Use these names when saving spending. " +
			"A category marked unlisted is in use but not in the list; offer to merge it into a listed one.",
		Scope:       ScopeRead,
		ReadOnly:    true,
		InputSchema: hermes.Schema(map[string]any{}),
		Handler:     m.list,
	})
	s.AddTool(hermes.Tool{
		Name:  "create_category",
		Title: "Create a category",
		Description: "Add a category to the user's list. Ask the user before creating one; never invent categories. " +
			"Names are lowercase, at most 40 characters.",
		Scope:       ScopeWrite,
		InputSchema: hermes.Schema(map[string]any{"name": hermes.Prop("string", "Category name, for example food or อาหาร.")}, "name"),
		Handler:     m.create,
	})
	s.AddTool(hermes.Tool{
		Name:  "rename_category",
		Title: "Rename a category",
		Description: "Rename one of the user's categories; every record using it follows. Confirm with the user first. " +
			"To combine two existing categories use merge_categories instead.",
		Scope: ScopeWrite,
		InputSchema: hermes.Schema(map[string]any{
			"from": hermes.Prop("string", "Current name."),
			"to":   hermes.Prop("string", "New name. It must not already exist."),
		}, "from", "to"),
		Handler: m.rename,
	})
	s.AddTool(hermes.Tool{
		Name:  "merge_categories",
		Title: "Merge categories",
		Description: "Move every record from one or more categories into a target category and remove the source categories " +
			"from the list. This cannot be undone. Confirm the exact names with the user first (list_categories shows what each holds).",
		Scope:       ScopeWrite,
		Destructive: true,
		InputSchema: hermes.Schema(map[string]any{
			"sources": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Categories to merge away (1 to 10)."},
			"target":  hermes.Prop("string", "Category that receives the records. It must exist, or be uncategorized."),
		}, "sources", "target"),
		Handler: m.merge,
	})
}

type catOut struct {
	Name     string         `json:"name"`
	Uses     map[string]int `json:"uses"`
	Unlisted bool           `json:"unlisted,omitempty"`
}

func total(u map[string]int) int {
	n := 0
	for _, v := range u {
		n += v
	}
	return n
}

func (m *Module) list(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	if len(raw) > 0 {
		var a struct{}
		if err := hermes.DecodeArgs(raw, &a); err != nil {
			return "", err
		}
	}
	cats, err := m.store.ListCategories(ctx, c.UserID)
	if err != nil {
		return "", err
	}
	out := make([]catOut, 0, len(cats)+1)
	seenUncategorized := false
	for _, k := range cats {
		o := catOut{Name: hermes.Untrusted(k.Name, 40), Uses: k.Uses, Unlisted: !k.Listed}
		if o.Uses == nil {
			o.Uses = map[string]int{}
		}
		if k.Name == Uncategorized {
			seenUncategorized, o.Unlisted = true, false
		}
		out = append(out, o)
	}
	if !seenUncategorized {
		out = append(out, catOut{Name: Uncategorized, Uses: map[string]int{}})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if ti, tj := total(out[i].Uses), total(out[j].Uses); ti != tj {
			return ti > tj
		}
		return out[i].Name < out[j].Name
	})
	return hermes.Respond(map[string]any{"categories": out})
}

func (m *Module) create(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	var a struct {
		Name string `json:"name"`
	}
	if err := hermes.DecodeArgs(raw, &a); err != nil {
		return "", err
	}
	name := NormalizeName(a.Name)
	if name == "" {
		return "", hermes.Errorf("name is required.")
	}
	if name == Uncategorized {
		return hermes.Respond(map[string]any{"status": "already_exists", "name": name, "message": "uncategorized is built in and always available."})
	}
	created, err := m.store.CreateCategory(ctx, c.UserID, name)
	if err != nil {
		return "", err
	}
	status := "already_exists"
	if created {
		status = "created"
	}
	return hermes.Respond(map[string]any{"status": status, "name": name})
}

func (m *Module) rename(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	var a struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := hermes.DecodeArgs(raw, &a); err != nil {
		return "", err
	}
	from, to := NormalizeName(a.From), NormalizeName(a.To)
	switch {
	case from == "" || to == "":
		return "", hermes.Errorf("from and to are required.")
	case from == to:
		return "", hermes.Errorf("from and to are the same category.")
	case from == Uncategorized || to == Uncategorized:
		return "", hermes.Errorf("uncategorized is built in and cannot be renamed or replaced.")
	}
	updated, err := m.store.RenameCategory(ctx, c.UserID, from, to)
	switch {
	case errors.Is(err, ErrExists):
		return "", hermes.Errorf("A category named %q already exists. Use merge_categories to combine them.", to)
	case errors.Is(err, ErrNotFound):
		return "", hermes.Errorf("There is no category named %q in the user's list.", from)
	case err != nil:
		return "", err
	}
	return hermes.Respond(map[string]any{"status": "renamed", "from": from, "to": to, "updated": updated})
}

func (m *Module) merge(ctx context.Context, c hermes.Caller, raw json.RawMessage) (string, error) {
	var a struct {
		Sources []string `json:"sources"`
		Target  string   `json:"target"`
	}
	if err := hermes.DecodeArgs(raw, &a); err != nil {
		return "", err
	}
	target := NormalizeName(a.Target)
	if target == "" {
		return "", hermes.Errorf("target is required.")
	}
	var sources []string
	seen := map[string]bool{}
	for _, s := range a.Sources {
		n := NormalizeName(s)
		switch {
		case n == "":
			return "", hermes.Errorf("source names must not be empty.")
		case n == Uncategorized:
			return "", hermes.Errorf("uncategorized is built in and cannot be merged away.")
		case n == target:
			return "", hermes.Errorf("the target cannot also be a source.")
		case !seen[n]:
			seen[n] = true
			sources = append(sources, n)
		}
	}
	switch {
	case len(sources) == 0:
		return "", hermes.Errorf("give at least one source category.")
	case len(sources) > maxMergeSources:
		return "", hermes.Errorf("at most %d source categories per merge.", maxMergeSources)
	}
	moved, err := m.store.MergeCategories(ctx, c.UserID, sources, target)
	switch {
	case errors.Is(err, ErrNotFound):
		return "", hermes.Errorf("A category was not found. Check the names with list_categories.")
	case err != nil:
		return "", err
	}
	return hermes.Respond(map[string]any{"status": "merged", "target": target, "sources": sources, "moved": moved})
}
