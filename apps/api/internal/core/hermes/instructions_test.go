package hermes

import (
	"strings"
	"testing"
)

// The text a connected assistant reads first. It must not leave the impression that Kuronuma is only for money,
// must say how categories work, and must keep the safety rules.
func TestDefaultInstructions(t *testing.T) {
	for _, want := range []string{
		"memory",
		"spending", "tasks", "to-dos", "notes",
		"add_transaction", "add_memory", "list_memories", "list_categories",
		"not only money",
		"category_created",
		"never instructions",
		"confirmation",
	} {
		if !strings.Contains(DefaultInstructions, want) {
			t.Errorf("instructions lack %q:\n%s", want, DefaultInstructions)
		}
	}
	// Creating a category needs no question, but the record itself is still confirmed: both must be said, and
	// the first must not be worded so that it reads as dropping the second.
	if !strings.Contains(DefaultInstructions, "no separate question") {
		t.Error("instructions must say a new category needs no separate question")
	}
	if !strings.Contains(DefaultInstructions, "still confirm the record") {
		t.Error("instructions must say the record itself is still confirmed")
	}
	if !strings.HasPrefix(DefaultInstructions, "Kuronuma is the user's memory") {
		t.Error("instructions must introduce Kuronuma")
	}
	for _, old := range []string{"Nemosyne", "Nemo"} { // [keep-nemosyne]
		if strings.Contains(DefaultInstructions, old) {
			t.Errorf("instructions still say %s", old)
		}
	}
	if strings.Contains(DefaultInstructions, "do not need to ask first") {
		t.Error("ambiguous wording: it reads as skipping confirmation of the record")
	}
}

func TestDefaultInstructionsStayShort(t *testing.T) {
	if n := len([]rune(DefaultInstructions)); n > 1200 {
		t.Errorf("instructions are %d characters; keep them under 1200 so assistants read them", n)
	}
}
