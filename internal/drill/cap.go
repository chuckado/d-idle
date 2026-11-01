package drill

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// DefaultCap is the size a run log kept in the repo stays under (the repo
// hygiene gate fails at 1 MB; tools/ci/check.sh).
const DefaultCap = 400 * 1024

var keepRe = regexp.MustCompile(`FAIL|BLOCKED|PASS|ERR|stuck|timeout|^DRILL |^H\.|^cheat |^group `)

// CapLog shrinks a run log to at most about max bytes: the head and the tail
// (a quarter of max each), plus every middle line that carries a verdict or
// error marker, while those fit; one marker line says what was dropped. A log
// already under max is returned unchanged.
func CapLog(text string, max int) string {
	if len(text) <= max {
		return text
	}
	lines := strings.SplitAfter(text, "\n")
	head, tail := 0, len(lines)
	for n := 0; head < len(lines) && n+len(lines[head]) <= max/4; head++ {
		n += len(lines[head])
	}
	for n := 0; tail > head && n+len(lines[tail-1]) <= max/4; tail-- {
		n += len(lines[tail-1])
	}
	var keep []string
	budget := max / 2
	dropL, dropB := 0, 0
	for _, l := range lines[head:tail] {
		if keepRe.MatchString(l) && len(l) <= budget {
			if len(l) > 600 {
				l = l[:600] + "...\n"
			}
			keep = append(keep, l)
			budget -= len(l)
			continue
		}
		dropL++
		dropB += len(l)
	}
	var b strings.Builder
	b.WriteString(strings.Join(lines[:head], ""))
	fmt.Fprintf(&b, "[capped: %d lines (%d bytes) of the middle dropped, %d marker lines kept below; the full log is not kept in the repo (it stayed on the worker host)]\n", dropL, dropB, len(keep))
	b.WriteString(strings.Join(keep, ""))
	b.WriteString("[end of kept marker lines; tail of the log follows]\n")
	b.WriteString(strings.Join(lines[tail:], ""))
	return b.String()
}

// CapFile caps path in place (see CapLog); it reports the sizes.
func CapFile(path string, max int) (before, after int, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	out := CapLog(string(raw), max)
	if len(out) == len(raw) {
		return len(raw), len(raw), nil
	}
	return len(raw), len(out), os.WriteFile(path, []byte(out), 0o644)
}
