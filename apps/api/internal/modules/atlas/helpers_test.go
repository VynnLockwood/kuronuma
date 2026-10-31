package atlas

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNormalizeTags(t *testing.T) {
	ok := map[string][]string{
		"case and spacing":                {"Work", "work", " WORK "},
		"empty and invisible are dropped": {"", "​", "work"},
	}
	for name, in := range ok {
		got, err := NormalizeTags(in)
		if err != nil || !reflect.DeepEqual(got, []string{"work"}) {
			t.Errorf("%s: got %v err=%v, want [work]", name, got, err)
		}
	}
	if got, err := NormalizeTags([]string{strings.Repeat("a", 40)}); err != nil || len([]rune(got[0])) != 30 {
		t.Errorf("a 40-character tag must be cut to 30: %v %v", got, err)
	}
	if got, err := NormalizeTags([]string{"งาน", "ครอบครัว"}); err != nil || !reflect.DeepEqual(got, []string{"งาน", "ครอบครัว"}) {
		t.Errorf("Thai tags must be kept: %v %v", got, err)
	}
	if got, err := NormalizeTags(nil); err != nil || got == nil || len(got) != 0 {
		t.Errorf("nil gives an empty, non-nil list: %v %v", got, err)
	}
	var eleven, tenWithDups []string
	for i := 0; i < 11; i++ {
		eleven = append(eleven, string(rune('a'+i)))
	}
	tenWithDups = append(append(tenWithDups, eleven[:10]...), "A", "b", "C")
	if _, err := NormalizeTags(eleven); err == nil {
		t.Error("11 distinct tags must be refused")
	}
	if got, err := NormalizeTags(tenWithDups); err != nil || len(got) != 10 {
		t.Errorf("10 distinct tags plus duplicates are fine: %v %v", got, err)
	}
}

func TestParseTerms(t *testing.T) {
	if got, err := ParseTerms("  buy   Milk "); err != nil || !reflect.DeepEqual(got, []string{"buy", "Milk"}) {
		t.Errorf("got %v err=%v, want [buy Milk]", got, err)
	}
	if got, err := ParseTerms("bu​y"); err != nil || !reflect.DeepEqual(got, []string{"buy"}) {
		t.Errorf("invisible characters must be removed: %v %v", got, err)
	}
	if got, err := ParseTerms("   "); err != nil || len(got) != 0 {
		t.Errorf("an empty query gives no terms: %v %v", got, err)
	}
	if _, err := ParseTerms("a b c d e f"); err == nil {
		t.Error("6 terms must be refused")
	}
	if got, err := ParseTerms("a b c d e"); err != nil || len(got) != 5 {
		t.Errorf("5 terms are fine: %v %v", got, err)
	}
	if got, _ := ParseTerms(strings.Repeat("x", 80)); len([]rune(got[0])) != 50 {
		t.Errorf("an 80-character term must be cut to 50, got %d", len([]rune(got[0])))
	}
	if got, err := ParseTerms("ร้านอาหาร"); err != nil || !reflect.DeepEqual(got, []string{"ร้านอาหาร"}) {
		t.Errorf("Thai terms must be kept: %v %v", got, err)
	}
}

func TestLikePattern(t *testing.T) {
	if got, want := LikePattern("50%_off\\x"), "%50\\%\\_off\\\\x%"; got != want {
		t.Errorf("LikePattern = %q, want %q", got, want)
	}
	if got := LikePattern("milk"); got != "%milk%" {
		t.Errorf("plain text must only be wrapped: %q", got)
	}
}

func TestLess(t *testing.T) {
	d := func(s string) *time.Time {
		v, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return &v
	}
	at := func(n int) time.Time { return time.Unix(int64(n), 0) }
	open := Memory{Status: StatusOpen, Priority: PriorityNormal, CreatedAt: at(1)}
	cases := []struct {
		name string
		a, b Memory
	}{
		{"open before done", open, Memory{Status: StatusDone, DueOn: d("2026-01-01"), Priority: PriorityHigh, CreatedAt: at(9)}},
		{"open before archived", open, Memory{Status: StatusArchived, Priority: PriorityHigh, CreatedAt: at(9)}},
		{"earlier due first", Memory{Status: StatusOpen, DueOn: d("2026-10-01"), Priority: PriorityLow}, Memory{Status: StatusOpen, DueOn: d("2026-10-05"), Priority: PriorityHigh}},
		{"dated before undated", Memory{Status: StatusOpen, DueOn: d("2030-01-01"), Priority: PriorityLow}, Memory{Status: StatusOpen, Priority: PriorityHigh}},
		{"high before normal", Memory{Status: StatusOpen, Priority: PriorityHigh}, Memory{Status: StatusOpen, Priority: PriorityNormal}},
		{"normal before low", Memory{Status: StatusOpen, Priority: PriorityNormal}, Memory{Status: StatusOpen, Priority: PriorityLow}},
		{"newest first as the last tie-break", Memory{Status: StatusOpen, Priority: PriorityNormal, CreatedAt: at(5)}, Memory{Status: StatusOpen, Priority: PriorityNormal, CreatedAt: at(2)}},
	}
	for _, tc := range cases {
		if !Less(tc.a, tc.b) || Less(tc.b, tc.a) {
			t.Errorf("%s: Less(a,b)=%v Less(b,a)=%v, want true and false", tc.name, Less(tc.a, tc.b), Less(tc.b, tc.a))
		}
	}
	if Less(open, open) {
		t.Error("Less must be false for equal items")
	}
}
