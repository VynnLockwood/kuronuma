package plutus

import (
	"context"
	"fmt"
	"strings"
	"testing"
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

func TestCategoryMustBeInTheList(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("alice", "food")

	for i, c := range []string{"food", "Food", " fo​od "} {
		data, text, isErr := e.call("alice", "add_transaction", withRef(i, map[string]any{"category": c}))
		if isErr {
			t.Fatalf("category %q refused: %s", c, text)
		}
		if got := data["transaction"].(map[string]any)["category"]; got != "food" {
			t.Errorf("category %q saved as %v, want food", c, got)
		}
	}
	before := e.store.count("alice")

	_, text, isErr := e.call("alice", "add_transaction", withRef(10, map[string]any{"category": "travel"}))
	if !isErr {
		t.Fatal("an unknown category must be refused")
	}
	for _, want := range []string{`"travel"`, "food", "create_category", "omit", "uncategorized"} {
		if !strings.Contains(text, want) {
			t.Errorf("refusal message lacks %q: %s", want, text)
		}
	}
	if e.store.count("alice") != before {
		t.Fatal("a refused transaction must not be stored")
	}

	for i, args := range []map[string]any{{}, {"category": "uncategorized"}, {"category": "Uncategorized"}} {
		data, text, isErr := e.call("alice", "add_transaction", withRef(20+i, args))
		if isErr {
			t.Fatalf("args %v refused: %s", args, text)
		}
		if got := data["transaction"].(map[string]any)["category"]; got != "uncategorized" {
			t.Errorf("args %v saved as %v, want uncategorized", args, got)
		}
	}

	e.store.setNames("bob")
	_, text, isErr = e.call("bob", "add_transaction", withRef(30, map[string]any{"category": "x"}))
	if !isErr || !strings.Contains(text, "no categories yet") {
		t.Errorf("a user with no categories: isErr=%v text=%q", isErr, text)
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
	_, text, isErr = e.call("alice", "add_transaction", withRef(2, map[string]any{"category": "เดินทาง"}))
	if !isErr || !strings.Contains(text, "อาหาร") || !strings.Contains(text, "เดินทาง") {
		t.Errorf("refusal must show the Thai names intact: isErr=%v %s", isErr, text)
	}
}

func TestRefusalListIsCapped(t *testing.T) {
	e := newEnv(t)
	var names []string
	for i := 1; i <= 120; i++ {
		names = append(names, fmt.Sprintf("cat%03d", i))
	}
	e.store.setNames("alice", names...)
	_, text, isErr := e.call("alice", "add_transaction", withRef(1, map[string]any{"category": "zzz"}))
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
	if _, text, isErr := e.call("alice", "add_transaction", withRef(1, map[string]any{"category": "work"})); !isErr {
		t.Fatalf("alice must not use bob's category: %s", text)
	}
	if _, text, isErr := e.call("bob", "add_transaction", withRef(2, map[string]any{"category": "work"})); isErr {
		t.Fatalf("bob's own category refused: %s", text)
	}
}
