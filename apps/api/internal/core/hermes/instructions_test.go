package hermes

import (
	"strings"
	"testing"
)

// The text a connected assistant reads first. It must not leave the impression that Nemosyne is only for money,
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
	if strings.Contains(DefaultInstructions, "ask the user before creating") {
		t.Error("categories are created without asking now")
	}
}

func TestDefaultInstructionsStayShort(t *testing.T) {
	if n := len([]rune(DefaultInstructions)); n > 1200 {
		t.Errorf("instructions are %d characters; keep them under 1200 so assistants read them", n)
	}
}
