package arbiter

import (
	"os"
	"strings"
	"testing"
)

func replayFile(t *testing.T, name string, follow bool) ReplayResult {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	res, err := Replay(f, DefaultConfig(), ShadowConfig{LogGapSec: 60}, follow)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// The Dark Wood ping-pong (docs/probes/treefix3-2026-11-19.md): the chain changed area every 7 s with no progress.
// The fixture is built from the real area-change gaps of internal/bot/flail_test.go flailFixture (V: copied from the sandbox log 2026-11-18); the facts (potions, gold, objective) are assumed (U), not logged.
func TestReplayPingPong(t *testing.T) {
	res := replayFile(t, "darkwood_pingpong.jsonl", true)
	if res.AreaChanges < 300 {
		t.Fatalf("fixture should show the chain flailing, got %d area changes", res.AreaChanges)
	}
	if res.FollowMoves > 4 {
		t.Fatalf("arbiter would have changed area %d times, want at most 4", res.FollowMoves)
	}
	if res.Refused == 0 {
		t.Fatal("expected travel budget refusals in the trace")
	}
	t.Logf("chain %d area changes, arbiter %d, refusals %d", res.AreaChanges, res.FollowMoves, res.Refused)
	goldenCheck(t, "testdata/pingpong_replay.golden", strings.Join(res.Trace[:40], "\n")+"\n")
}

func TestReplayLoggedHeroDivergences(t *testing.T) {
	res := replayFile(t, "darkwood_pingpong.jsonl", false)
	if len(res.Divergences) == 0 || !strings.Contains(res.Divergences[0], "chain chose") {
		t.Fatalf("no divergence line: %v", res.Divergences)
	}
	if !strings.Contains(res.FamilyCounts, "quest=") {
		t.Fatalf("counters: %q", res.FamilyCounts)
	}
}

func TestShadowSurvivesPanicAndKeepsCounting(t *testing.T) {
	bad := FuncSource{N: "bad", F: func(View) *Intent { panic("boom") }}
	s := NewShadow(DefaultConfig(), ShadowConfig{LogGapSec: 1}, bad)
	if line := s.Tick(View{Now: 1}, "quest"); line != "" {
		t.Fatalf("panic must give no line, got %q", line)
	}
	if !strings.Contains(s.Drain(), "panics=1") {
		t.Fatal("panic not counted")
	}
}

func TestShadowRateLimitAndCounters(t *testing.T) {
	s := NewShadow(DefaultConfig(), ShadowConfig{LogGapSec: 60}, StdSources()...)
	v := View{Facts: map[string]int{FactObjective: 1}, Area: "town"}
	n := 0
	for i := 0; i < 100; i++ {
		v.Now = float64(i)
		if s.Tick(v, "chores") != "" {
			n++
		}
	}
	if n != 2 { // t=0 and t=60
		t.Fatalf("rate limit: %d lines, want 2", n)
	}
	if d := s.Drain(); !strings.Contains(d, "chores=100/100") {
		t.Fatalf("counters: %q", d)
	}
	if d := s.Drain(); strings.Contains(d, "chores") {
		t.Fatalf("Drain must reset: %q", d)
	}
}

func BenchmarkShadowTick(b *testing.B) {
	s := NewShadow(DefaultConfig(), ShadowConfig{LogGapSec: 60}, StdSources()...)
	facts := map[string]int{FactObjective: 1, FactPotions: 1, FactPotionsMin: 4, FactPackGold: 500, FactNeedGold: 300}
	for i := 0; i < b.N; i++ {
		s.Tick(View{Now: float64(i), Area: "dark_wood", Facts: facts, ObjectiveArea: "dark_wood"}, "quest")
	}
}

func goldenCheck(t *testing.T, path, got string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Errorf("differs from %s (run with -update after reviewing):\n%s", path, got)
	}
}
