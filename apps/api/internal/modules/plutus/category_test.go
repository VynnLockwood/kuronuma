package plutus

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/themis"
)

func (m *memStore) setNames(user string, names ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.names == nil {
		m.names = map[string][]string{}
	}
	m.names[user] = append([]string(nil), names...)
}

func (m *memStore) CategoryNames(_ context.Context, userID string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.names[userID]...), nil
}

func withRef(i int, extra map[string]any) map[string]any {
	a := base(extra)
	a["reference"] = fmt.Sprintf("R%d", i)
	return a
}

func (m *memStore) CreateCategory(_ context.Context, userID, name string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.names == nil {
		m.names = map[string][]string{}
	}
	for _, n := range m.names[userID] {
		if n == name {
			return false, nil
		}
	}
	m.names[userID] = append(m.names[userID], name)
	return true, nil
}

func (m *memStore) listed(user, name string) bool {
	names, _ := m.CategoryNames(context.Background(), user)
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func TestUnknownCategoryIsCreated(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("alice", "food")

	// A name already in the list is used as is, whatever its case or invisible characters; nothing is created.
	for i, c := range []string{"food", "Food", " fo\u200bod "} {
		data, text, isErr := e.call("alice", "add_transaction", withRef(i, map[string]any{"category": c}))
		if isErr {
			t.Fatalf("category %q refused: %s", c, text)
		}
		if got := data["transaction"].(map[string]any)["category"]; got != "food" {
			t.Errorf("category %q saved as %v, want food", c, got)
		}
		if _, has := data["category_created"]; has {
			t.Errorf("category %q already existed, so category_created must be absent: %v", c, data)
		}
	}

	// A new name is created in the user's list and used.
	data, text, isErr := e.call("alice", "add_transaction", withRef(10, map[string]any{"category": "Travel"}))
	if isErr {
		t.Fatalf("a new category must be created, not refused: %s", text)
	}
	if got := data["transaction"].(map[string]any)["category"]; got != "travel" {
		t.Errorf("saved as %v, want travel", got)
	}
	if data["category_created"] != true {
		t.Errorf("the result must say the category was created: %v", data)
	}
	if !e.store.listed("alice", "travel") {
		t.Error("travel must now be in alice's list")
	}
	if e.store.listed("bob", "travel") {
		t.Error("another user's list must not change")
	}

	// The second save finds it: no second creation notice.
	data, _, _ = e.call("alice", "add_transaction", withRef(11, map[string]any{"category": "travel"}))
	if _, has := data["category_created"]; has {
		t.Errorf("travel already exists now: %v", data)
	}

	// Leaving the category out or using uncategorized never creates anything.
	before, _ := e.store.CategoryNames(context.Background(), "alice")
	for i, args := range []map[string]any{{}, {"category": "uncategorized"}, {"category": "Uncategorized"}} {
		data, text, isErr := e.call("alice", "add_transaction", withRef(20+i, args))
		if isErr {
			t.Fatalf("args %v refused: %s", args, text)
		}
		if got := data["transaction"].(map[string]any)["category"]; got != "uncategorized" {
			t.Errorf("args %v saved as %v, want uncategorized", args, got)
		}
	}
	if after, _ := e.store.CategoryNames(context.Background(), "alice"); len(after) != len(before) {
		t.Errorf("no category should have been created: %v -> %v", before, after)
	}

	// A user with no categories at all can start one.
	e.store.setNames("bob")
	data, text, isErr = e.call("bob", "add_transaction", withRef(30, map[string]any{"category": "x"}))
	if isErr || data["category_created"] != true {
		t.Errorf("first category ever: isErr=%v text=%q", isErr, text)
	}
}

func TestThaiCategory(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("alice", "อาหาร")
	data, text, isErr := e.call("alice", "add_transaction", withRef(1, map[string]any{"category": "อาหาร"}))
	if isErr {
		t.Fatalf("Thai category refused: %s", text)
	}
	if got := data["transaction"].(map[string]any)["category"]; got != "อาหาร" {
		t.Errorf("saved as %v", got)
	}
	data, text, isErr = e.call("alice", "add_transaction", withRef(2, map[string]any{"category": "เดินทาง"}))
	if isErr || data["category_created"] != true || data["transaction"].(map[string]any)["category"] != "เดินทาง" {
		t.Errorf("a new Thai category must be created intact: isErr=%v %s", isErr, text)
	}
	if !e.store.listed("alice", "เดินทาง") {
		t.Error("the Thai name must be in the list")
	}
}

func TestCategoryCap(t *testing.T) {
	e := newEnv(t)
	var names []string
	for i := 1; i <= themis.MaxCategories; i++ {
		names = append(names, fmt.Sprintf("cat%03d", i))
	}
	e.store.setNames("alice", names...)

	_, text, isErr := e.call("alice", "add_transaction", withRef(1, map[string]any{"category": "zzz"}))
	if !isErr {
		t.Fatal("a new category beyond the cap must be refused")
	}
	for _, want := range []string{`"zzz"`, "merge_categories", fmt.Sprint(themis.MaxCategories)} {
		if !strings.Contains(text, want) {
			t.Errorf("refusal lacks %q: %s", want, text)
		}
	}
	if e.store.count("alice") != 0 || e.store.listed("alice", "zzz") {
		t.Fatal("a refused save must store nothing and create nothing")
	}
	if _, text, isErr := e.call("alice", "add_transaction", withRef(2, map[string]any{"category": "cat007"})); isErr {
		t.Fatalf("a listed category still works at the cap: %s", text)
	}
}

// set_payee_rule keeps the old rule: its category must already be in the list, and the refusal names the
// first 50 of the user's categories.
func TestPayeeRuleRefusalListIsCapped(t *testing.T) {
	e := newEnv(t)
	var names []string
	for i := 1; i <= 120; i++ {
		names = append(names, fmt.Sprintf("cat%03d", i))
	}
	e.store.setNames("alice", names...)
	_, text, isErr := e.call("alice", "set_payee_rule", map[string]any{"payee": "Grab", "category": "zzz"})
	if !isErr {
		t.Fatal("want a refusal")
	}
	if !strings.Contains(text, "cat050") || strings.Contains(text, "cat051") || !strings.Contains(text, "and 70 more") {
		t.Errorf("want the first 50 names and 'and 70 more', got %s", text)
	}
}

func TestCategoryListIsPerUser(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("bob", "work")
	e.store.setNames("alice", "food")
	data, text, isErr := e.call("alice", "add_transaction", withRef(1, map[string]any{"category": "work"}))
	if isErr || data["category_created"] != true {
		t.Fatalf("alice has no work category yet, so it must be created for her, even though bob has one: %v %s", isErr, text)
	}
	if names, _ := e.store.CategoryNames(context.Background(), "bob"); len(names) != 1 {
		t.Errorf("bob's list must not change: %v", names)
	}
	data, text, isErr = e.call("bob", "add_transaction", withRef(2, map[string]any{"category": "work"}))
	if isErr {
		t.Fatalf("bob's own category refused: %s", text)
	}
	if _, has := data["category_created"]; has {
		t.Errorf("bob already had work: %v", data)
	}
}
