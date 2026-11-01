package drill

import (
	"strings"
	"testing"
)

func TestCapLog(t *testing.T) {
	small := "DRILL a PASS\n"
	if CapLog(small, 1000) != small {
		t.Fatal("small log changed")
	}
	var b strings.Builder
	b.WriteString("first line\n")
	for i := 0; i < 5000; i++ {
		b.WriteString("level=DEBUG msg=\"cast gap\" filler filler filler\n")
		if i == 2500 {
			b.WriteString("DRILL D1.X FAIL middle verdict\n")
		}
	}
	b.WriteString("last line\n")
	out := CapLog(b.String(), 20000)
	if len(out) > 20000 {
		t.Fatalf("size %d over cap", len(out))
	}
	for _, want := range []string{"first line", "last line", "D1.X FAIL middle verdict", "[capped:", "not kept in the repo"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}
