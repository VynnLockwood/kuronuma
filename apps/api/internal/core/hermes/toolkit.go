package hermes

import (
	"bytes"
	"encoding/json"
)

// Schema builds the JSON Schema for a tool's arguments: an object with the given properties, no extra
// properties allowed, and the named ones required.
func Schema(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

// Prop describes one argument of a tool.
func Prop(typ, desc string) map[string]any { return map[string]any{"type": typ, "description": desc} }

// DecodeArgs strictly decodes tool arguments into dst: unknown fields and wrong types are errors,
// returned as a ToolError the assistant can read.
func DecodeArgs(raw json.RawMessage, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return Errorf("Invalid arguments: %s", Untrusted(err.Error(), 200))
	}
	return nil
}

// Respond marshals a tool result and adds the standing note that text fields are user data.
func Respond(fields map[string]any) (string, error) {
	fields["_note"] = UntrustedNote
	b, err := json.Marshal(fields)
	return string(b), err
}
