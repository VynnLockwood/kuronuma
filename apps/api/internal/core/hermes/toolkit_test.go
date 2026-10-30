package hermes

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestSchema(t *testing.T) {
	s := Schema(map[string]any{"a": Prop("string", "the a")})
	if s["type"] != "object" || s["additionalProperties"] != false {
		t.Fatalf("unexpected schema: %v", s)
	}
	if _, ok := s["required"]; ok {
		t.Error("required must be absent when none is given")
	}
	props := s["properties"].(map[string]any)["a"].(map[string]any)
	if props["type"] != "string" || props["description"] != "the a" {
		t.Errorf("Prop built %v", props)
	}
	req, ok := Schema(map[string]any{}, "a", "b")["required"].([]string)
	if !ok || len(req) != 2 || req[0] != "a" || req[1] != "b" {
		t.Errorf("required = %v, want [a b]", req)
	}
}

func TestDecodeArgsStrict(t *testing.T) {
	var dst struct {
		Name string `json:"name"`
	}
	if err := DecodeArgs(json.RawMessage(`{"name":"x"}`), &dst); err != nil || dst.Name != "x" {
		t.Fatalf("valid input: name=%q err=%v", dst.Name, err)
	}
	for _, in := range []string{`{"name":"x","extra":1}`, `{"name":5}`, `[1]`} {
		var te ToolError
		err := DecodeArgs(json.RawMessage(in), &dst)
		if !errors.As(err, &te) || !strings.HasPrefix(te.Msg, "Invalid arguments") {
			t.Errorf("%s: err=%v, want a ToolError starting 'Invalid arguments'", in, err)
		}
	}
}

func TestRespondAddsNote(t *testing.T) {
	out, err := Respond(map[string]any{"status": "ok"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got["status"] != "ok" || got["_note"] != UntrustedNote {
		t.Errorf("got %v", got)
	}
}
