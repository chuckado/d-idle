package drill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Table is the short text summary: one row per result and the cheat verbs used.
func (g Group) Table() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-9s %-5s %-8s %-4s %s\n", "drill", "prim", "verdict", "det", "reason")
	cheats := map[string]bool{}
	for _, r := range g.Results {
		det := "-"
		if len(r.Runs) > 1 {
			det = map[bool]string{true: "same", false: "DIFF"}[r.Deterministic]
		}
		fmt.Fprintf(&b, "%-9s %-5s %-8s %-4s %s\n", r.ID, r.Primitive, r.Verdict, det, oneLine(r.Reason))
		for _, c := range r.Cheats {
			cheats[c] = true
		}
	}
	if len(cheats) > 0 {
		var l []string
		for c := range cheats {
			l = append(l, c)
		}
		sort.Strings(l)
		fmt.Fprintf(&b, "cheat verbs used (each ledgered with a reason): %s\n", strings.Join(l, ", "))
	}
	if g.Stopped != "" {
		fmt.Fprintf(&b, "group stopped: %s\n", g.Stopped)
	}
	return b.String()
}

// Text is the result lines followed by the table.
func (g Group) Text() string {
	var b strings.Builder
	for _, r := range g.Results {
		b.WriteString(r.Line() + "\n")
	}
	b.WriteString("\n" + g.Table())
	return b.String()
}

// Write stores <dir>/<name>.txt and <dir>/<name>.json.
func (g Group) Write(dir, name string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, name+".txt"), []byte(g.Text()), 0o644); err != nil {
		return err
	}
	j, err := json.MarshalIndent(g, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".json"), append(j, '\n'), 0o644)
}
