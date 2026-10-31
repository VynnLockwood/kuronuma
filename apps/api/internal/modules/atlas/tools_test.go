package atlas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/hermes"
)

// ---- in-memory store, per user, mirroring the Postgres store's behaviour ----

type memStore struct {
	mu    sync.Mutex
	items map[string][]*Memory
	names map[string][]string
	seq   int
}

func (s *memStore) setNames(user string, names ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.names == nil {
		s.names = map[string][]string{}
	}
	s.names[user] = append([]string(nil), names...)
}

func (s *memStore) CategoryNames(_ context.Context, user string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.names[user]...), nil
}

func (s *memStore) AddMemory(_ context.Context, user string, n NewMemory) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items == nil {
		s.items = map[string][]*Memory{}
	}
	s.seq++
	m := &Memory{
		ID: fmt.Sprintf("00000000-0000-0000-0000-%012d", s.seq), Kind: n.Kind, Title: n.Title, Body: n.Body,
		Status: StatusOpen, Priority: n.Priority, Category: n.Category, Tags: append([]string{}, n.Tags...),
		CreatedAt: time.Unix(int64(s.seq), 0), UpdatedAt: time.Unix(int64(s.seq), 0),
	}
	if n.DueOn != nil {
		d := *n.DueOn
		m.DueOn = &d
	}
	s.items[user] = append(s.items[user], m)
	return *m, nil
}

func (s *memStore) ListMemories(_ context.Context, user string, f Filter) ([]Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Memory
	for _, m := range s.items[user] {
		if f.Kind != "" && m.Kind != f.Kind || f.Status != "" && m.Status != f.Status || f.Category != "" && m.Category != f.Category {
			continue
		}
		if f.Tag != "" {
			found := false
			for _, t := range m.Tags {
				found = found || t == f.Tag
			}
			if !found {
				continue
			}
		}
		if f.DueBefore != nil && (m.DueOn == nil || m.DueOn.After(*f.DueBefore)) || f.DueAfter != nil && (m.DueOn == nil || m.DueOn.Before(*f.DueAfter)) {
			continue
		}
		hay := strings.ToLower(m.Title + "\n" + m.Body + "\n" + strings.Join(m.Tags, "\n"))
		ok := true
		for _, term := range f.Terms {
			ok = ok && strings.Contains(hay, strings.ToLower(term))
		}
		if ok {
			out = append(out, *m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return Less(out[i], out[j]) })
	limit := f.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *memStore) UpdateMemory(_ context.Context, user, id string, u Update) (Memory, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.items[user] {
		if m.ID != id {
			continue
		}
		s.seq++
		now := time.Unix(int64(s.seq), 0)
		if u.Title != nil {
			m.Title = *u.Title
		}
		if u.Body != nil {
			m.Body = *u.Body
		}
		if u.Priority != nil {
			m.Priority = *u.Priority
		}
		if u.Category != nil {
			m.Category = *u.Category
		}
		if u.Tags != nil {
			m.Tags = append([]string{}, *u.Tags...)
		}
		if u.SetDue {
			m.DueOn = u.DueOn
		}
		if u.Status != nil {
			m.Status = *u.Status
			switch {
			case m.Status != StatusDone:
				m.CompletedAt = nil
			case m.CompletedAt == nil:
				m.CompletedAt = &now
			}
		}
		m.UpdatedAt = now
		return *m, true, nil
	}
	return Memory{}, false, nil
}

func (s *memStore) DeleteMemory(_ context.Context, user, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, m := range s.items[user] {
		if m.ID == id {
			s.items[user] = append(s.items[user][:i], s.items[user][i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// ---- test environment: the real MCP server with the module over the fake ----

type auth struct{}

func (auth) Authenticate(_ context.Context, token string) (hermes.Caller, error) {
	all := []string{ScopeRead, ScopeWrite}
	switch token {
	case "alice":
		return hermes.Caller{UserID: "alice", ClientID: "c", Scopes: all}, nil
	case "bob":
		return hermes.Caller{UserID: "bob", ClientID: "c", Scopes: all}, nil
	case "reader":
		return hermes.Caller{UserID: "alice", ClientID: "c", Scopes: []string{ScopeRead}}, nil
	}
	return hermes.Caller{}, hermes.ErrUnauthorized
}

type env struct {
	t     *testing.T
	srv   *hermes.Server
	store *memStore
}

func newEnv(t *testing.T) *env {
	t.Helper()
	st := &memStore{}
	srv := hermes.New(hermes.Options{Name: "t", ResourceMetadataURL: "https://x/meta"}, auth{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	New(st).Register(srv)
	return &env{t: t, srv: srv, store: st}
}

// try calls a tool and never fails the test itself, so it is safe from goroutines.
func (e *env) try(token, tool string, args any) (data map[string]any, text string, isErr bool, ok bool) {
	raw, _ := json.Marshal(args)
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"` + tool + `","arguments":` + string(raw) + `}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.srv.ServeHTTP(rec, req)
	var out map[string]any
	if json.Unmarshal(rec.Body.Bytes(), &out) != nil {
		return nil, rec.Body.String(), false, false
	}
	res, found := out["result"].(map[string]any)
	if !found {
		return nil, rec.Body.String(), false, false
	}
	text = res["content"].([]any)[0].(map[string]any)["text"].(string)
	isErr = res["isError"].(bool)
	if !isErr {
		_ = json.Unmarshal([]byte(text), &data)
	}
	return data, text, isErr, true
}

func (e *env) call(token, tool string, args any) (map[string]any, string, bool) {
	e.t.Helper()
	data, text, isErr, ok := e.try(token, tool, args)
	if !ok {
		e.t.Fatalf("no result for %s: %s", tool, text)
	}
	return data, text, isErr
}

func (e *env) must(token, tool string, args any) map[string]any {
	e.t.Helper()
	data, text, isErr := e.call(token, tool, args)
	if isErr {
		e.t.Fatalf("%s %v failed: %s", tool, args, text)
	}
	return data
}

func (e *env) add(extra map[string]any) map[string]any {
	e.t.Helper()
	a := map[string]any{"kind": "task", "title": "Write report"}
	for k, v := range extra {
		a[k] = v
	}
	return e.must("alice", "add_memory", a)["memory"].(map[string]any)
}

func (e *env) titles(token string, args map[string]any) []string {
	e.t.Helper()
	data := e.must(token, "list_memories", args)
	var out []string
	for _, x := range data["memories"].([]any) {
		out = append(out, x.(map[string]any)["title"].(string))
	}
	return out
}

func join(s []string) string { return strings.Join(s, " | ") }

func runes(v any) int { return len([]rune(v.(string))) }

// ---- tests ----

func TestAddMemory(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("alice", "work")

	m := e.add(nil)
	if m["status"] != "open" || m["priority"] != "normal" || m["category"] != "uncategorized" || m["id"] == "" || len(m["tags"].([]any)) != 0 {
		t.Fatalf("defaults: %v", m)
	}
	if _, has := m["due_on"]; has {
		t.Error("no due date given, so due_on must be absent")
	}

	base := func(extra map[string]any) map[string]any {
		a := map[string]any{"kind": "task", "title": "x"}
		for k, v := range extra {
			a[k] = v
		}
		return a
	}
	eleven := make([]string, 11)
	for i := range eleven {
		eleven[i] = fmt.Sprintf("tag%d", i)
	}
	bad := map[string]map[string]any{
		"invalid kind":     base(map[string]any{"kind": "idea"}),
		"missing kind":     {"title": "x"},
		"missing title":    {"kind": "task"},
		"invisible title":  base(map[string]any{"title": "​"}),
		"bad priority":     base(map[string]any{"priority": "urgent"}),
		"bad due format":   base(map[string]any{"due_on": "28/10/2026"}),
		"impossible date":  base(map[string]any{"due_on": "2026-02-31"}),
		"date before 1990": base(map[string]any{"due_on": "1980-01-01"}),
		"date after 2100":  base(map[string]any{"due_on": "2101-01-01"}),
		"eleven tags":      base(map[string]any{"tags": eleven}),
		"unknown category": base(map[string]any{"category": "nope"}),
		"unknown argument": base(map[string]any{"user_id": "bob"}),
	}
	for name, args := range bad {
		if _, text, isErr := e.call("alice", "add_memory", args); !isErr {
			t.Errorf("%s: want an error, got %s", name, text)
		}
	}
	if _, text, _ := e.call("alice", "add_memory", base(map[string]any{"category": "nope"})); !strings.Contains(text, "create_category") || !strings.Contains(text, "work") {
		t.Errorf("an unknown category gets the shared refusal with the user's list: %s", text)
	}

	long := e.add(map[string]any{"title": strings.Repeat("t", 300), "body": strings.Repeat("b", 5000)})
	if runes(long["title"]) != 200 || runes(long["body"]) != 4000 {
		t.Errorf("title %d and body %d characters, want 200 and 4000", runes(long["title"]), runes(long["body"]))
	}
	tagged := e.add(map[string]any{"tags": []string{"Work", " work", "Q4", ""}, "category": "Work", "due_on": "2026-11-05", "priority": "high", "kind": "todo"})
	if got := fmt.Sprint(tagged["tags"]); got != "[work q4]" || tagged["category"] != "work" || tagged["due_on"] != "2026-11-05" || tagged["priority"] != "high" || tagged["kind"] != "todo" {
		t.Errorf("normalised item: %v", tagged)
	}
	if th := e.add(map[string]any{"title": "ร้านอาหารครัวไทย", "tags": []string{"งาน"}}); th["title"] != "ร้านอาหารครัวไทย" || fmt.Sprint(th["tags"]) != "[งาน]" {
		t.Errorf("Thai round trip: %v", th)
	}
}

func seed(e *env) {
	e.add(map[string]any{"title": "Write report", "body": "quarterly numbers", "priority": "high", "due_on": "2026-11-05", "tags": []string{"q4", "report"}})
	e.add(map[string]any{"kind": "todo", "title": "Buy milk"})
	e.add(map[string]any{"kind": "note", "title": "100% sure", "body": `a_b and c\d`, "priority": "low", "due_on": "2026-10-20", "tags": []string{"misc"}})
	e.add(map[string]any{"title": "ร้านอาหารครัวไทย"})
	e.add(map[string]any{"kind": "note", "title": "aXb"})
}

func TestListFiltersAndOrder(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("alice", "work")
	seed(e)
	all := join(e.titles("alice", map[string]any{}))
	if !strings.HasPrefix(all, "100% sure | Write report | ") {
		t.Fatalf("order: due date first (earlier first), then undated: %s", all)
	}
	cases := map[string]struct {
		args map[string]any
		want string
	}{
		"kind":                    {map[string]any{"kind": "note", "query": "sure"}, "100% sure"},
		"tag":                     {map[string]any{"tag": "Q4"}, "Write report"},
		"due_before is inclusive": {map[string]any{"due_before": "2026-10-20"}, "100% sure"},
		"due_after is inclusive":  {map[string]any{"due_after": "2026-11-05"}, "Write report"},
		"both due filters":        {map[string]any{"due_after": "2026-10-01", "due_before": "2026-12-01"}, "100% sure | Write report"},
		"status open":             {map[string]any{"status": "open", "kind": "todo"}, "Buy milk"},
		"no match":                {map[string]any{"kind": "todo", "tag": "q4"}, ""},
	}
	for name, tc := range cases {
		if got := join(e.titles("alice", tc.args)); got != tc.want {
			t.Errorf("%s: got %q, want %q", name, got, tc.want)
		}
	}
	for name, args := range map[string]map[string]any{
		"invalid kind":     {"kind": "idea"},
		"invalid status":   {"status": "pending"},
		"bad due_before":   {"due_before": "tomorrow"},
		"bad due_after":    {"due_after": "2026-02-31"},
		"unknown argument": {"user_id": "bob"},
	} {
		if _, _, isErr := e.call("alice", "list_memories", args); !isErr {
			t.Errorf("%s: want an error", name)
		}
	}

	done := e.titles("alice", map[string]any{"kind": "todo"})
	id := e.must("alice", "list_memories", map[string]any{"kind": "todo"})["memories"].([]any)[0].(map[string]any)["id"]
	e.must("alice", "update_memory", map[string]any{"id": id, "status": "done"})
	if got := join(e.titles("alice", map[string]any{"status": "done"})); got != done[0] {
		t.Errorf("status done: %q", got)
	}
	if got := e.titles("alice", map[string]any{}); got[len(got)-1] != done[0] {
		t.Errorf("a done item sorts after every open one: %v", got)
	}

	for i := 0; i < 260; i++ {
		_, _ = e.store.AddMemory(context.Background(), "alice", NewMemory{Kind: "note", Title: fmt.Sprintf("bulk %d", i), Priority: "normal", Category: "uncategorized"})
	}
	if n := len(e.titles("alice", map[string]any{})); n != 50 {
		t.Errorf("default limit returned %d, want 50", n)
	}
	if n := len(e.titles("alice", map[string]any{"limit": 500})); n != 200 {
		t.Errorf("limit 500 returned %d, want 200 (the maximum)", n)
	}
	if n := len(e.titles("alice", map[string]any{"limit": 3})); n != 3 {
		t.Errorf("limit 3 returned %d", n)
	}
}

func TestListQuery(t *testing.T) {
	e := newEnv(t)
	seed(e)
	cases := map[string]string{
		"quarterly":        "Write report",
		"quarterly report": "Write report",
		"quarterly milk":   "",
		"QUARTERLY":        "Write report",
		"Q4":               "Write report",
		"อาหาร":            "ร้านอาหารครัวไทย",
		"100%":             "100% sure",
		"a_b":              "100% sure",
		`c\d`:              "100% sure",
		"   ":              join(e.titles("alice", map[string]any{})),
	}
	for q, want := range cases {
		if got := join(e.titles("alice", map[string]any{"query": q})); got != want {
			t.Errorf("query %q: got %q, want %q", q, got, want)
		}
	}
	if _, text, isErr := e.call("alice", "list_memories", map[string]any{"query": "a b c d e f"}); !isErr || !strings.Contains(text, "at most 5") {
		t.Errorf("six search words must be refused: %v %s", isErr, text)
	}
}

func TestUpdateMemory(t *testing.T) {
	e := newEnv(t)
	e.store.setNames("alice", "work")
	e.store.setNames("bob", "work")
	m := e.add(map[string]any{"title": "Draft", "due_on": "2026-11-05", "tags": []string{"a"}})
	id := m["id"]
	upd := func(extra map[string]any) map[string]any {
		a := map[string]any{"id": id}
		for k, v := range extra {
			a[k] = v
		}
		data := e.must("alice", "update_memory", a)
		if data["status"] != "updated" {
			t.Fatalf("update %v: %v", extra, data)
		}
		return data["memory"].(map[string]any)
	}

	if got := upd(map[string]any{"title": "Final", "priority": "high", "category": "Work"}); got["title"] != "Final" || got["priority"] != "high" || got["category"] != "work" {
		t.Errorf("field updates: %v", got)
	}
	done := upd(map[string]any{"status": "done"})
	completed, has := done["completed_at"]
	if !has {
		t.Fatal("completing an item must set completed_at")
	}
	if kept := upd(map[string]any{"body": "edited"}); kept["completed_at"] != completed {
		t.Errorf("editing a done item must keep completed_at: %v vs %v", kept["completed_at"], completed)
	}
	if again := upd(map[string]any{"status": "done"}); again["completed_at"] != completed {
		t.Error("done to done must keep completed_at")
	}
	if reopened := upd(map[string]any{"status": "open"}); reopened["completed_at"] != nil || reopened["status"] != "open" {
		t.Errorf("reopening must clear completed_at: %v", reopened)
	}

	if got := upd(map[string]any{"body": "no date change"}); got["due_on"] != "2026-11-05" {
		t.Errorf("an absent due_on leaves the date alone: %v", got["due_on"])
	}
	if got := upd(map[string]any{"due_on": "2026-12-01"}); got["due_on"] != "2026-12-01" {
		t.Errorf("due_on changed: %v", got["due_on"])
	}
	if got := upd(map[string]any{"due_on": ""}); got["due_on"] != nil {
		t.Errorf(`due_on "" clears the date: %v`, got["due_on"])
	}
	if got := upd(map[string]any{"tags": []string{"X", "y"}}); fmt.Sprint(got["tags"]) != "[x y]" {
		t.Errorf("tags replaced: %v", got["tags"])
	}
	if got := upd(map[string]any{"tags": []string{}}); len(got["tags"].([]any)) != 0 {
		t.Errorf("an empty list clears the tags: %v", got["tags"])
	}
	if got := upd(map[string]any{"body": ""}); got["body"] != "" {
		t.Errorf("an empty body clears it: %v", got["body"])
	}

	bad := map[string]map[string]any{
		"no fields":        {"id": id},
		"bad due":          {"id": id, "due_on": "2026-02-31"},
		"bad status":       {"id": id, "status": "pending"},
		"bad priority":     {"id": id, "priority": "urgent"},
		"empty title":      {"id": id, "title": "​"},
		"unknown category": {"id": id, "category": "nope"},
		"malformed id":     {"id": "not-an-id", "title": "x"},
		"too many tags":    {"id": id, "tags": []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}},
		"no id":            {"title": "x"},
	}
	for name, args := range bad {
		if _, text, isErr := e.call("alice", "update_memory", args); !isErr {
			t.Errorf("%s: want an error, got %s", name, text)
		}
	}
	if data := e.must("alice", "update_memory", map[string]any{"id": "00000000-0000-0000-0000-999999999999", "title": "x"}); data["status"] != "not_found" {
		t.Errorf("a well-formed unknown id: %v", data)
	}
	if data, _, isErr := e.call("bob", "update_memory", map[string]any{"id": id, "title": "stolen"}); isErr || data["status"] != "not_found" {
		t.Errorf("bob must not reach alice's item: %v", data)
	}
	if got := e.titles("alice", map[string]any{}); got[0] != "Final" {
		t.Errorf("alice's item changed by bob: %v", got)
	}
}

func TestDeleteMemory(t *testing.T) {
	e := newEnv(t)
	id := e.add(nil)["id"]
	if data := e.must("bob", "delete_memory", map[string]any{"id": id}); data["status"] != "not_found" {
		t.Errorf("bob must not delete alice's item: %v", data)
	}
	if data := e.must("alice", "delete_memory", map[string]any{"id": id}); data["status"] != "deleted" {
		t.Errorf("delete: %v", data)
	}
	if data := e.must("alice", "delete_memory", map[string]any{"id": id}); data["status"] != "not_found" {
		t.Errorf("second delete: %v", data)
	}
	if _, _, isErr := e.call("alice", "delete_memory", map[string]any{"id": "nope"}); !isErr {
		t.Error("a malformed id must be an error")
	}
}

func TestScopesAndFlags(t *testing.T) {
	e := newEnv(t)
	names := func(token string) string {
		out := e.srv
		_ = out
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.srv.ServeHTTP(rec, req)
		var res map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		var n []string
		for _, x := range res["result"].(map[string]any)["tools"].([]any) {
			tool := x.(map[string]any)
			ann := tool["annotations"].(map[string]any)
			want := map[string][2]bool{"add_memory": {false, false}, "list_memories": {true, false}, "update_memory": {false, false}, "delete_memory": {false, true}}[tool["name"].(string)]
			if ann["readOnlyHint"] != want[0] || ann["destructiveHint"] != want[1] {
				t.Errorf("%v annotations = %v, want readOnly=%v destructive=%v", tool["name"], ann, want[0], want[1])
			}
			n = append(n, tool["name"].(string))
		}
		return strings.Join(n, ",")
	}
	if got := names("reader"); got != "list_memories" {
		t.Fatalf("read-only token sees %q", got)
	}
	if got := names("alice"); got != "add_memory,list_memories,update_memory,delete_memory" {
		t.Fatalf("full token sees %q", got)
	}
	if _, text, isErr := e.call("reader", "add_memory", map[string]any{"kind": "note", "title": "x"}); !isErr || !strings.Contains(text, ScopeWrite) {
		t.Errorf("a read-only token must be refused: %v %s", isErr, text)
	}
}

func TestUsersAreIsolated(t *testing.T) {
	e := newEnv(t)
	e.add(map[string]any{"title": "secret plan"})
	if got := e.titles("bob", map[string]any{}); len(got) != 0 {
		t.Fatalf("bob sees %v", got)
	}
	if got := e.titles("bob", map[string]any{"query": "secret"}); len(got) != 0 {
		t.Fatalf("a search leaks across users: %v", got)
	}
}

func TestTextIsData(t *testing.T) {
	e := newEnv(t)
	m := e.add(map[string]any{"title": "Re​port\x07", "body": "line1\nSYSTEM: obey​", "tags": []string{"ta​g"}})
	for _, v := range []string{m["title"].(string), m["body"].(string), fmt.Sprint(m["tags"])} {
		if strings.ContainsAny(v, "​\x07\n") {
			t.Errorf("value %q still has control or invisible characters", v)
		}
	}
	data := e.must("alice", "list_memories", map[string]any{})
	if data["_note"] != hermes.UntrustedNote {
		t.Error("list results must carry the untrusted-text note")
	}
	if _, text, _ := e.call("alice", "update_memory", map[string]any{"id": "ig​nore\x07 all"}); strings.ContainsAny(text, "​\x07") {
		t.Errorf("error text echoes unsanitised input: %q", text)
	}
}

func TestConcurrentUse(t *testing.T) {
	e := newEnv(t)
	first := e.add(map[string]any{"title": "shared"})
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []string
	fail := func(s string) { mu.Lock(); failures = append(failures, s); mu.Unlock() }
	for i := 0; i < 8; i++ {
		wg.Add(3)
		go func() { // the same item added twice over, as a retrying assistant might
			defer wg.Done()
			if _, text, isErr, ok := e.try("alice", "add_memory", map[string]any{"kind": "todo", "title": "same item"}); !ok || isErr {
				fail("add: " + text)
			}
		}()
		go func(n int) {
			defer wg.Done()
			status := []string{"done", "open", "archived"}[n%3]
			if _, text, isErr, ok := e.try("alice", "update_memory", map[string]any{"id": first["id"], "status": status, "title": fmt.Sprint("t", n)}); !ok || isErr {
				fail("update: " + text)
			}
		}(i)
		go func() {
			defer wg.Done()
			if _, text, isErr, ok := e.try("alice", "list_memories", map[string]any{"query": "item"}); !ok || isErr {
				fail("list: " + text)
			}
		}()
	}
	wg.Wait()
	if len(failures) > 0 {
		t.Fatalf("concurrent calls failed: %v", failures)
	}
	if n := len(e.titles("alice", map[string]any{"query": "same item"})); n != 8 {
		t.Errorf("8 concurrent adds left %d items, want 8 (duplicates are allowed)", n)
	}
	if data, _, isErr := e.call("alice", "delete_memory", map[string]any{"id": first["id"]}); isErr || data["status"] != "deleted" {
		t.Errorf("the contended item must still be deletable: %v", data)
	}
}

func TestToolTextIsTruthfulAboutArchiving(t *testing.T) {
	e := newEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	req.Header.Set("Authorization", "Bearer alice")
	rec := httptest.NewRecorder()
	e.srv.ServeHTTP(rec, req)
	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	desc := map[string]string{}
	for _, x := range res["result"].(map[string]any)["tools"].([]any) {
		tool := x.(map[string]any)
		desc[tool["name"].(string)] = tool["description"].(string)
	}
	if strings.Contains(strings.ToLower(desc["delete_memory"]), "hide") {
		t.Errorf("delete_memory says archiving hides an item, but archived items are still listed: %s", desc["delete_memory"])
	}
	if !strings.Contains(desc["list_memories"], "archived") {
		t.Errorf("list_memories must say that done and archived items are included: %s", desc["list_memories"])
	}
	for _, tool := range []string{"update_memory", "delete_memory"} {
		if !strings.Contains(desc[tool], "still appear in list_memories") {
			t.Errorf("%s must say that archived items still appear in list_memories: %s", tool, desc[tool])
		}
	}
	if !strings.Contains(desc["list_memories"], "status") || !strings.Contains(desc["list_memories"], "open") {
		t.Errorf("list_memories must say how to get only what is still to do: %s", desc["list_memories"])
	}

	// The behaviour the text must describe: archived items are listed (after open ones) unless filtered out.
	e.add(map[string]any{"title": "Open one"})
	id := e.add(map[string]any{"title": "Set aside"})["id"]
	e.must("alice", "update_memory", map[string]any{"id": id, "status": "archived"})
	if got := join(e.titles("alice", map[string]any{})); got != "Open one | Set aside" {
		t.Errorf("an unfiltered list shows archived items after the open ones: %q", got)
	}
	if got := join(e.titles("alice", map[string]any{"status": "open"})); got != "Open one" {
		t.Errorf("status open shows only what is still to do: %q", got)
	}
}
