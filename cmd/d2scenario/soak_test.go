package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/scenario"
)

func intp(n int) *int { return &n }

// A fake clock that moves a minute per run and a fake runner: the soak runs until the duration is used, bumps the seed after a pass,
// splits the hits by invariants.json and counts survival.
func TestSoakLoop(t *testing.T) {
	now := time.Unix(0, 0)
	clock := func() time.Time { return now }
	var seeds []int64
	run := func(r scenario.Run) scenario.Result {
		now = now.Add(time.Minute)
		seeds = append(seeds, r.Seed)
		res := scenario.Result{Status: scenario.Green, State: &scenario.State{LifePct: intp(80), Invariants: map[string]int{"town-thrash": 2}}}
		res.Faults = &scenario.FaultReport{Set: "x", Counts: map[string]int{"cancel": 3}}
		switch len(seeds) {
		case 2:
			res.Status = scenario.Red
			res.State.Invariants = map[string]int{"potion-not-drunk": 1}
		case 3:
			res.State.LifePct = intp(0)
		case 4:
			res = scenario.Result{Status: scenario.Blocked}
		}
		return res
	}
	rows := []scenario.Run{{ID: "a", Class: "paladin", Seed: 10}, {ID: "b", Class: "sorceress", Seed: 20}}
	rep := soakLoop(context.Background(), rows, 5*time.Minute, clock, run, loadInvariantPolicy())
	if len(rep.Runs) != 5 {
		t.Fatalf("%d runs", len(rep.Runs))
	}
	if seeds[2] != 11 || seeds[3] != 21 {
		t.Fatalf("seeds %v: the second pass must bump", seeds)
	}
	if rep.Hard["potion-not-drunk"] != 1 || rep.Soft["town-thrash"] != 6 || rep.Injected["cancel"] != 15 {
		t.Fatalf("%+v", rep)
	}
	if rep.Survived != 2 || rep.Died != 1 || rep.Blocked != 1 {
		t.Fatalf("survived %d died %d blocked %d", rep.Survived, rep.Died, rep.Blocked)
	}
	var b bytes.Buffer
	rep.print(&b)
	if !strings.Contains(b.String(), "hard invariant hits: potion-not-drunk=1") {
		t.Fatal(b.String())
	}
}

func TestSoakLoopAlwaysOneRunAndStopsOnCancel(t *testing.T) {
	now := time.Unix(0, 0)
	n := 0
	run := func(scenario.Run) scenario.Result {
		n++
		now = now.Add(time.Hour)
		return scenario.Result{Status: scenario.Green}
	}
	rep := soakLoop(context.Background(), []scenario.Run{{ID: "a"}}, time.Minute, func() time.Time { return now }, run, loadInvariantPolicy())
	if len(rep.Runs) != 1 || n != 1 {
		t.Fatalf("%d runs", len(rep.Runs))
	}
	ctx, stop := context.WithCancel(context.Background())
	stop()
	n = 0
	rep = soakLoop(ctx, []scenario.Run{{ID: "a"}}, time.Hour, func() time.Time { return now }, run, loadInvariantPolicy())
	if len(rep.Runs) != 1 {
		t.Fatalf("a cancelled soak still ran %d", len(rep.Runs))
	}
}

func TestSurvivedNeedsLifeAndNoLifeZero(t *testing.T) {
	ok := scenario.Result{Status: scenario.Red, State: &scenario.State{LifePct: intp(10)}}
	if !survived(ok) {
		t.Fatal("a RED run with life left survived")
	}
	ok.State.Invariants = map[string]int{"life-zero-alive": 1}
	if survived(ok) {
		t.Fatal("life-zero-alive is not survival")
	}
	if survived(scenario.Result{Status: scenario.Green}) {
		t.Fatal("no state, no proof")
	}
}

func TestFaultSetsFileIsValidAndNamed(t *testing.T) {
	sets, err := faultSets()
	if err != nil || len(sets) < 4 {
		t.Fatalf("%v %d", err, len(sets))
	}
	if _, err := faultSet("nope"); err == nil || !strings.Contains(err.Error(), "socket-timeouts") {
		t.Fatalf("%v", err)
	}
}

func TestFaultsEnvBeatsScenario(t *testing.T) {
	s := scenario.Scenario{Faults: "mixed"}
	t.Setenv(faultsEnv, "")
	if faultsFor(s) != "mixed" {
		t.Fatal("scenario set")
	}
	t.Setenv(faultsEnv, "cancel-storm")
	if faultsFor(s) != "cancel-storm" {
		t.Fatal("env override")
	}
}

func TestSoakDryRun(t *testing.T) {
	var b bytes.Buffer
	err := cmdSoak([]string{"-dir", "../../tools/provingground/scenarios", "-id", "smoke.", "-faults", "mixed", "-dry-run", "-duration", "1h"}, &b)
	if err != nil && !strings.Contains(err.Error(), "no row to soak") {
		t.Fatal(err)
	}
	if err == nil && !strings.Contains(b.String(), "DRY RUN") {
		t.Fatal(b.String())
	}
	if err := cmdSoak([]string{"-faults", "nope", "-dry-run"}, &b); err == nil {
		t.Fatal("unknown set accepted")
	}
}
