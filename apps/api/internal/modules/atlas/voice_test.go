package atlas

import (
	"fmt"
	"strings"
	"testing"
)

func TestMemoryLimitMessageIsGentleAndExact(t *testing.T) {
	text := fmt.Sprintf(msgMemoryLimit, 5)
	if !strings.HasPrefix(text, "Sorry,") {
		t.Errorf("want a gentle opener, got %q", text)
	}
	for _, f := range []string{"5 tasks, to-dos and notes", "Ask the user", "delete_memory"} {
		if !strings.Contains(text, f) {
			t.Errorf("lost the fact %q in %q", f, text)
		}
	}
	if len(text) > 240 || !strings.HasSuffix(text, ".") {
		t.Errorf("want at most 240 characters ending with a full stop, got %d", len(text))
	}
}
