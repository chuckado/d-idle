package watch

import (
	"strings"
	"testing"
)

// The shell must start its poll loop before it imports the panel modules one by one: the imports queue behind the
// picture and event streams (the browser's per-host connection cap), and a stalled one held every panel in its first
// state ("waiting for the bot") because the loop only started after the last import.
func TestShellStartsPollLoopBeforePanelImports(t *testing.T) {
	b, err := webFS.ReadFile("web/shell.js")
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	start := strings.Index(src, "sched.start();")
	imp := strings.Index(src, "import('/watch/panels/' + m.id + '/panel.js')")
	if start < 0 || imp < 0 {
		t.Fatalf("anchors missing: start=%d import=%d", start, imp)
	}
	if start > imp {
		t.Error("sched.start() runs after the panel imports: one stalled import freezes every panel")
	}
	if strings.Count(src, "sched.start();") != 1 {
		t.Error("the poll loop must start exactly once")
	}
	if !strings.Contains(src[imp-60:imp+200], "Promise.race") {
		t.Error("a panel import needs a timeout, or one stalled module blocks the later panels for good")
	}
}
