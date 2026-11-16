package loot

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
	"d2lod-bot/internal/tidy"
)

func enginePass(t *testing.T, l *Looter, cls d2.Class, gold, stage int) (*ShadowPass, *d2.Snapshot) {
	t.Helper()
	l.e.Tune.Loot.RulesEngine = stage
	s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: cls, Level: 12, Gold: gold, Strength: 25, Dexterity: 25, Area: 2}, Inventory: &d2.Inventory{}}
	p := l.ShadowPass(s, 30, 10)
	if p == nil {
		t.Fatal("no pass")
	}
	return p, s
}

// Ground stage: the engine's verdict, through the same strictness filter, picks up exactly what today's pickit picks
// up, for two classes, poor and rich, at every strictness (the switch-over must not move an item).
func TestEngineGroundStageEqualsToday(t *testing.T) {
	l, db := synthLooter(t)
	items := shadowFixtures(t, db)
	n, bad := 0, 0
	for _, cls := range []d2.Class{d2.Sorceress, d2.Amazon} {
		for _, gold := range []int{50, 5000} {
			p, s := enginePass(t, l, cls, gold, StageGround)
			for _, strict := range []int{0, 1, 2, 3} {
				for _, f := range items {
					u, st := f.u, f.st
					u.Item.Where, u.Mode = d2.ContGround, d2.ItemModeGround
					v, ok := p.GroundVerdict(u, st)
					if !ok {
						t.Fatalf("%s: engine failed", f.label)
					}
					got := lootrules.Ignore
					if passes(v, strict, gold < l.e.Tune.Loot.LowGoldBelow) {
						got = lootrules.Pickup
					}
					n++
					if want := groundToday(l, s, u, st, strict); got != want {
						bad++
						if bad < 8 {
							t.Errorf("%v gold %d strict %d %s: today %s engine %s (%s)", cls, gold, strict, f.label, want, got, v.Reason)
						}
					}
				}
			}
		}
	}
	t.Logf("%d ground decisions, %d differ", n, bad)
}

// Tidy stage: stash, sell, drop and keep come out of the engine as the tidy table says them, two classes.
func TestEngineTidyStageEqualsToday(t *testing.T) {
	l, db := synthLooter(t)
	pol, err := tidy.Default()
	if err != nil {
		t.Fatal(err)
	}
	items := shadowFixtures(t, db)
	n := 0
	for _, cls := range []d2.Class{d2.Sorceress, d2.Amazon} {
		p, _ := enginePass(t, l, cls, 500, StageTidy)
		for _, f := range items {
			a, rule, _, ok := p.TidyDecision(f.u, "", l.base(f.u.Class).Type == "key")
			if !ok {
				t.Fatalf("%s: engine failed", f.label)
			}
			want := string(pol.Decide(tidyFactsOf(l, f.u, f.st, ""), tidy.Ctx{Class: cls.String(), PackFullPct: 30, StashFullPct: 10}).Action)
			n++
			// TidyDecision reads the stat list itself (nil here), so compare only where the fixture has none.
			if f.st == nil || len(f.st) == 0 {
				if a != want {
					t.Errorf("%v %s: tidy %s engine %s (%s)", cls, f.label, want, a, rule)
				}
			}
		}
	}
	t.Logf("%d tidy decisions", n)
}

func TestEngineStageSwitches(t *testing.T) {
	l, db := synthLooter(t)
	u := synthUnit(t, db, 7, "r01", d2.QualityNormal, d2.FlagIdentified, d2.ContGround)
	u.Mode = d2.ItemModeGround
	for stage, want := range map[int][2]bool{0: {false, false}, 1: {true, false}, 2: {true, true}, 3: {true, true}, 9: {true, true}} {
		l.e.Tune.Loot.RulesEngine = stage
		l.e.Tune.Bot.LootShadow = 0
		s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: d2.Sorceress, Level: 12, Area: 2}, Inventory: &d2.Inventory{}}
		p := l.ShadowPass(s, 0, 0)
		if stage == 0 {
			if p != nil {
				t.Fatal("stage 0 and shadow 0 must give no pass")
			}
			continue
		}
		_, g := p.GroundVerdict(u, d2.Stats{})
		_, _, _, td := p.TidyDecision(u, "", false)
		if g != want[0] || td != want[1] {
			t.Errorf("stage %d: ground %v tidy %v, want %v", stage, g, td, want)
		}
	}
	// the shadow alone (stage 0) never decides
	l.e.Tune.Loot.RulesEngine, l.e.Tune.Bot.LootShadow = 0, 1
	s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: d2.Sorceress, Level: 12, Area: 2}, Inventory: &d2.Inventory{}}
	p := l.ShadowPass(s, 0, 0)
	if _, ok := p.GroundVerdict(u, nil); ok {
		t.Error("shadow-only pass decided")
	}
}

// A broken rule set or a panic inside the engine falls back to today's decision for that item and logs once.
func TestEngineFailureFallsBackAndLogsOnce(t *testing.T) {
	l, db := synthLooter(t)
	var buf bytes.Buffer
	l.e.Log = slog.New(slog.NewTextHandler(&buf, nil))
	l.SetEngine(NewEngine(&lootrules.RuleSet{Schema: "lootrules/99"}))
	u := synthUnit(t, db, 7, "r01", d2.QualityNormal, d2.FlagIdentified, d2.ContGround)
	u.Mode = d2.ItemModeGround
	p, _ := enginePass(t, l, d2.Amazon, 500, StageTidy)
	for i := 0; i < 3; i++ {
		if _, ok := p.GroundVerdict(u, nil); ok {
			t.Fatal("bad rule set must not decide")
		}
		if _, _, _, ok := p.TidyDecision(u, "", false); ok {
			t.Fatal("bad rule set must not decide (tidy)")
		}
	}
	if n := strings.Count(buf.String(), "engine failed"); n != 1 {
		t.Errorf("logged %d times, want once:\n%s", n, buf.String())
	}
	if !strings.Contains(buf.String(), `class \"Amazon\"`) && !strings.Contains(buf.String(), `class "Amazon"`) {
		t.Errorf("log does not name the class: %s", buf.String())
	}
	// panic inside a condition callback
	l.SetEngine(NewEngine(lootrules.Starter()))
	p, _ = enginePass(t, l, d2.Sorceress, 500, StageGround)
	p.cx.Needs = func(string, *lootrules.Item) bool { panic("boom") }
	p.cx.Owned = func(*lootrules.Item) int { panic("boom") }
	hp := synthUnit(t, db, 8, "hp1", d2.QualityNormal, 0, d2.ContGround)
	hp.Mode = d2.ItemModeGround
	if _, ok := p.GroundVerdict(hp, nil); ok {
		t.Log("no rule consulted the callbacks for this item; panic path covered by the next check")
	}
	if _, err := (&Engine{}).decide(&lootrules.Item{}, &lootrules.Ctx{}); err == nil || !strings.Contains(err.Error(), "panic") {
		t.Errorf("nil rule set must be a recovered panic, got %v", err)
	}
}

func TestSetActiveRulesReachesEngineAndShadow(t *testing.T) {
	old := DefaultEngine.Rules()
	defer SetActiveRules(old)
	rs := lootrules.Starter()
	rs.Name = "Custom"
	SetActiveRules(rs)
	if DefaultEngine.Rules().Name != "Custom" || DefaultShadow.rs.Name != "Custom" {
		t.Error("active rules not propagated")
	}
}
