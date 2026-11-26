package main

import (
	"strings"
	"testing"

	"d2lod-bot/internal/scenario"
)

func res(st, why string) scenario.Result { return scenario.Result{Status: st, Reason: why} }

// Three runs, two green needed: a flaky row (GREEN, RED, GREEN) is GREEN, (GREEN, RED, RED) is RED with the tally and the
// first failure. Must fail while one run decides.
func TestRepeatTwoOfThree(t *testing.T) {
	if repeatNeed(3) != 2 || repeatNeed(1) != 1 || repeatNeed(5) != 4 {
		t.Fatalf("need %d %d %d", repeatNeed(3), repeatNeed(1), repeatNeed(5))
	}
	g := aggregateRepeats([]scenario.Result{res("GREEN", ""), res("RED", "died"), res("GREEN", "")}, 2)
	if g.Status != "GREEN" || !strings.Contains(g.Reason, "2/3") {
		t.Fatalf("%+v", g)
	}
	r := aggregateRepeats([]scenario.Result{res("RED", "character died"), res("GREEN", ""), res("RED", "x")}, 2)
	if r.Status != "RED" || !strings.Contains(r.Reason, "1/3") || !strings.Contains(r.Reason, "character died") {
		t.Fatalf("%+v", r)
	}
	if b := aggregateRepeats([]scenario.Result{res("BLOCKED", "slot"), res("BLOCKED", "slot"), res("GREEN", "")}, 2); b.Status != "BLOCKED" {
		t.Fatalf("%+v", b)
	}
}
