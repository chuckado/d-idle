package main

import (
	"os"
	"strings"
	"testing"
)

// segment returns src from start up to end, as the portal_copy.go generator took it.
func segment(t *testing.T, file, start, end string) string {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	i := strings.Index(s, start)
	if i < 0 {
		t.Fatalf("%s: marker %q not found", file, start)
	}
	j := strings.Index(s[i:], end)
	if j < 0 {
		t.Fatalf("%s: marker %q not found", file, end)
	}
	return s[i : i+j]
}

// TestPortalCopyInStep fails when cmd/d2bot's portal verb, d2dTown.ToTown or townAdapter change: portal_copy.go holds
// verbatim copies (package main cannot be imported), so the drills would otherwise test code the bot no longer runs.
// Refresh: copy the three segments below from the originals into portal_copy.go.
func TestPortalCopyInStep(t *testing.T) {
	cp, err := os.ReadFile("portal_copy.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []struct{ file, start, end string }{
		{"../d2bot/d2dwire.go", "const (\n\ttownPortalObject", "// lateEmit is"},
		{"../d2bot/d2dwire.go", "type d2dTown struct {", "// d2dHealth maps"},
		{"../d2bot/adapters.go", "type townAdapter struct {", "func (a townAdapter) Chores("},
	} {
		want := strings.TrimRight(segment(t, s.file, s.start, s.end), "\n")
		if !strings.Contains(string(cp), want) {
			t.Errorf("%s from %q is no longer in portal_copy.go: refresh the copy", s.file, s.start)
		}
	}
}
