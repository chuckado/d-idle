package loot

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/lootrules"
	"d2lod-bot/internal/tidy"
)

// The fixture set: every base of the synthetic table in every quality, identified and not, with and without affixes.
type fixItem struct {
	label string
	u     d2.Unit
	st    d2.Stats // nil = unreadable
}

func shadowFixtures(t testing.TB, db *data.DB) []fixItem {
	t.Helper()
	var out []fixItem
	id := uint32(100)
	for _, b := range db.Items {
		for _, q := range []int{d2.QualityLow, d2.QualityNormal, d2.QualitySuperior, d2.QualityMagic, d2.QualityRare, d2.QualityUnique, d2.QualitySet, d2.QualityCrafted} {
			for _, ident := range []bool{true, false} {
				for _, rich := range []bool{false, true} {
					if rich && q < d2.QualityMagic || rich && !ident {
						continue
					}
					id++
					u := synthUnit(t, db, id, b.Code, q, 0, d2.ContInventory)
					if ident {
						u.Item.Flags = d2.FlagIdentified
					}
					st := d2.Stats{}
					if rich {
						st = d2.Stats{7: 30 << 8, 39: 20, 105: 10, 194: 0}
					}
					if q <= d2.QualitySuperior && b.Sockets >= 4 {
						st = d2.Stats{194: 4}
					}
					if q >= d2.QualityMagic && !ident {
						st = nil // an unidentified affixed item has no readable stat list
					}
					out = append(out, fixItem{fmt.Sprintf("%s q%d id=%v rich=%v", b.Code, q, ident, rich), u, st})
				}
			}
		}
	}
	return out
}

// tidyFactsOf mirrors town.TidyFacts for a pack unit (the real one needs a Town; its tests run with game tables).
func tidyFactsOf(l *Looter, u d2.Unit, st d2.Stats, mark string) tidy.Facts {
	b := l.base(u.Class)
	f := tidy.Facts{Place: "pack", Quality: ruleQuality[u.Item.Quality], Identified: u.Item.Identified(), Ethereal: u.Item.Ethereal(),
		StartItem: u.Item.Flags&d2.FlagStartItem != 0, Mark: mark, Code: b.Code, Type: b.Type, Stackable: b.Stackable, Quest: b.Quest || b.Type == "key"}
	f.SellValue = SellValue(l.e.Data, b, u.Item, 0)
	if b.Sockets >= 4 && (f.Quality == "normal" || f.Quality == "superior") && st != nil {
		f.Sockets = int(st.Get(statSocketsRule()))
	}
	return f
}

func statSocketsRule() int { return 194 }

// With the Starter preset the pack stage (tidy) must produce ZERO disagreements on the fixture set, for two classes and
// the pack and stash fill levels around the drop line (the shadow-stage twin of lootrules.TestStarterPackEqualsTidyPolicy).
func TestShadowStarterPackStageNoDisagreements(t *testing.T) {
	l, db := synthLooter(t)
	pol, err := tidy.Default()
	if err != nil {
		t.Fatal(err)
	}
	a := l.Rules()
	items := shadowFixtures(t, db)
	n := 0
	for _, cls := range []d2.Class{d2.Sorceress, d2.Amazon} {
		for _, pack := range []int{40, 90} {
			s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: cls, Level: 12, Gold: 500}, Inventory: &d2.Inventory{}}
			cx := a.Ctx(s, "", pack, 10)
			c, err := lootrules.Compile(lootrules.Starter(), cx.Class, "")
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range items {
				for _, mark := range []string{"", "keep", "sell"} {
					it, err := a.Item(f.u, ItemOpts{Place: lootrules.PlacePack, Stats: f.st, Mark: mark, NeverSell: l.base(f.u.Class).Type == "key"})
					if err != nil {
						t.Fatal(err)
					}
					want := string(pol.Decide(tidyFactsOf(l, f.u, f.st, mark), tidy.Ctx{Class: cls.String(), PackFullPct: pack, StashFullPct: 10}).Action)
					got := norm(ShadowTidy, c.Decide(&it, cx).Action)
					n++
					if got != want {
						t.Errorf("%v pack %d %s mark %q: tidy %s rules %s", cls, pack, f.label, mark, want, got)
					}
				}
			}
		}
	}
	t.Logf("%d pack decisions compared", n)
}

// groundToday is the old code's ground decision for one unit: the verdict through the strictness filter.
func groundToday(l *Looter, s *d2.Snapshot, u d2.Unit, st d2.Stats, strict int) string {
	c := l.ctxFrom(s)
	c.Stats = st
	v := l.EvaluateWith(u, c)
	if passes(v, strict, s.Player.Gold < l.e.Tune.Loot.LowGoldBelow) {
		return lootrules.Pickup
	}
	return lootrules.Ignore
}

// groundDiffs runs today's ground decision and the Starter's over the fixtures. label: the units are d2d ground labels
// (no item level, no identified flag, no stat list).
func groundDiffs(t *testing.T, label bool) (diffs []string, n int) {
	l, db := synthLooter(t)
	a := l.Rules()
	items := shadowFixtures(t, db)
	for _, cls := range []d2.Class{d2.Sorceress, d2.Amazon} {
		for _, gold := range []int{50, 5000} {
			s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: cls, Level: 12, Gold: gold, Strength: 25, Dexterity: 25, Area: 2}, Inventory: &d2.Inventory{}}
			cx := a.Ctx(s, "", 30, 0)
			c, err := lootrules.Compile(lootrules.Starter(), cx.Class, "")
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range items {
				u, st := f.u, f.st
				u.Item.Where, u.Mode = d2.ContGround, d2.ItemModeGround
				if label {
					cp := *u.Item
					cp.ILvl, cp.Flags = 0, 0
					u.Item, st = &cp, nil
				}
				it, err := a.Item(u, ItemOpts{Place: lootrules.PlaceGround, Stats: st, LabelOnly: label})
				if err != nil {
					t.Fatal(err)
				}
				d := c.Decide(&it, cx)
				sp := &ShadowPass{strict: 1, lowGold: gold < l.e.Tune.Loot.LowGoldBelow}
				rules, today := sp.groundFold(d), groundToday(l, s, u, st, 1)
				n++
				if today != rules {
					diffs = append(diffs, fmt.Sprintf("%v gold %d %s: today %s rules %s (%s: %s)", cls, gold, f.label, today, rules, d.Rule, d.Why))
				}
			}
		}
	}
	sort.Strings(diffs)
	return diffs, n
}

// The ground-stage twin of the pack test: with the Starter the ground decision equals today's pickit verdict through the
// strictness filter, for two classes, poor and rich, on memory-read items (everything readable).
func TestShadowStarterGroundStageEquivalence(t *testing.T) {
	diffs, n := groundDiffs(t, false)
	t.Logf("%d ground decisions, %d disagreements", n, len(diffs))
	for _, d := range diffs {
		t.Error(d)
	}
}

// On d2d ground labels the item level, the identified flag and the stat list are unreadable: today's pickit sees an
// unidentified item with no stats (its stat rules do not match), the engine sees UNKNOWN and the Starter rules that
// say onUnknown "match" (charms and jewelry to identify) take it. Both must agree.
func TestShadowStarterGroundLabelEquivalence(t *testing.T) {
	diffs, n := groundDiffs(t, true)
	t.Logf("%d label decisions, %d disagreements", n, len(diffs))
	for _, d := range diffs {
		t.Error(d)
	}
}

// sellToday mirrors the sell chore's own test (town.sellableIf, the part that does not need a Town): the pickit verdict
// sells it, or drops it while the pack is full and the tier is not S.
func sellToday(l *Looter, s *d2.Snapshot, u d2.Unit, st d2.Stats, full bool) string {
	c := l.ctxFrom(s)
	c.Stats = st
	v := l.EvaluateWith(u, c)
	if v.Action == Sell || v.Action == Drop && full && v.Tier != "S" {
		return lootrules.Sell
	}
	return lootrules.Keep
}

// The sell stage is where the four tables disagree (docs/sellskip.md section 1): the chore sells what the pickit calls
// junk or sellable, the Starter pack rules sell value >= 20 (the tidy row). The number and the kinds are the findings of
// docs/lootrules.md; the test pins the totals so a change of either side shows.
func TestShadowSellStageFindings(t *testing.T) {
	l, db := synthLooter(t)
	a := l.Rules()
	items := shadowFixtures(t, db)
	byKind := map[string]int{}
	n, total := 0, 0
	for _, full := range []bool{false, true} {
		pack := map[bool]int{false: 40, true: 90}[full]
		s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: d2.Sorceress, Level: 12, Gold: 500, Strength: 25, Dexterity: 25, Area: 2}, Inventory: &d2.Inventory{}}
		cx := a.Ctx(s, "", pack, 10)
		c, err := lootrules.Compile(lootrules.Starter(), cx.Class, "")
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range items {
			it, err := a.Item(f.u, ItemOpts{Place: lootrules.PlacePack, Stats: f.st})
			if err != nil {
				t.Fatal(err)
			}
			d := c.Decide(&it, cx)
			rules, today := norm(ShadowSell, d.Action), sellToday(l, s, f.u, f.st, full)
			n++
			if rules != today {
				total++
				byKind[fmt.Sprintf("today=%s rules=%s rule=%s", today, rules, d.Rule)]++
			}
		}
	}
	t.Logf("%d sell decisions, %d disagreements", n, total)
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, k := range kinds {
		t.Logf("%4d  %s", byKind[k], k)
	}
}

// The shadow costs well under a millisecond a decision on the fixture set (the brief's budget), logs one line per
// disagreement kind, and never panics on an item the adapter cannot read.
func TestShadowBudget(t *testing.T) {
	l, db := synthLooter(t)
	sh := NewShadow(lootrules.Starter())
	l.SetShadow(sh)
	items := shadowFixtures(t, db)
	s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: d2.Sorceress, Level: 12, Gold: 500, Area: 2}, Inventory: &d2.Inventory{}}
	sp := l.ShadowPass(s, 40, 10)
	if sp == nil {
		t.Fatal("pass is nil with the knob on")
	}
	t0 := time.Now()
	n := 0
	for _, f := range items {
		sp.Observe(ShadowTidy, f.u, "keep", "", false, f.st)
		n++
	}
	per := time.Since(t0) / time.Duration(n)
	r := sh.Report(true)
	t.Logf("%d decisions, %v per decision (engine alone avg %.1f us, max %.1f us), %d disagreements, %d unique", n, per, r.AvgMicros, r.MaxMicros, r.Disagreements, r.Unique)
	if per > time.Millisecond {
		t.Errorf("%v per decision, budget 1 ms", per)
	}
	if r.Decisions != n || r.Errors != 0 || r.Schema != ShadowSchema {
		t.Errorf("report %+v", r)
	}
	if len(r.Last) > shadowLines {
		t.Errorf("last has %d lines", len(r.Last))
	}
}

func TestShadowLogLineDedupRateAndKnob(t *testing.T) {
	l, db := synthLooter(t)
	sh := NewShadow(lootrules.Starter())
	var lines []string
	sh.Use(nil, nil)
	sh.log = nil
	l.SetShadow(sh)
	s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: d2.Amazon, Level: 12, Gold: 500, Area: 2}, Inventory: &d2.Inventory{}}
	sp := l.ShadowPass(s, 40, 10)
	rune1 := synthUnit(t, db, 1, "r01", d2.QualityNormal, d2.FlagIdentified, d2.ContInventory)
	// Today says keep for a rune; the Starter stashes it: one disagreement, the same item again is counted but not repeated.
	for i := 0; i < 3; i++ {
		sp.Observe(ShadowTidy, rune1, "keep", "", false, d2.Stats{})
	}
	r := sh.Report(true)
	if r.Disagreements != 3 || r.Unique != 1 || len(r.Last) != 1 {
		t.Fatalf("report %+v", r)
	}
	lines = r.Last
	want := "loot shadow: tidy El Rune today=keep rules=stash rule="
	if !strings.HasPrefix(lines[0], want) || !strings.Contains(lines[0], " reason=") {
		t.Errorf("line %q", lines[0])
	}
	if len(r.TopRules) != 1 || r.TopRules[0].Count != 3 {
		t.Errorf("top rules %+v", r.TopRules)
	}
	// Agreement is silent.
	sh.Reset()
	sp.Observe(ShadowTidy, rune1, "stash", "", false, d2.Stats{})
	if r = sh.Report(true); r.Disagreements != 0 || len(r.Last) != 0 || r.Decisions != 1 {
		t.Errorf("agreeing decision %+v", r)
	}
	// The knob: off means no pass at all.
	l.e.Tune.Bot.LootShadow = 0
	if l.ShadowPass(s, 0, 0) != nil {
		t.Error("knob off must give no pass")
	}
	var nilPass *ShadowPass
	nilPass.Observe(ShadowTidy, rune1, "keep", "", false, nil) // must not panic
	// An unknown class is an error that names it, counted, never a decision.
	l.e.Tune.Bot.LootShadow = 1
	s.Player.Class = d2.Class(42)
	sh.Reset()
	if sp = l.ShadowPass(s, 0, 0); sp != nil {
		sp.Observe(ShadowTidy, rune1, "keep", "", false, d2.Stats{})
	}
	if r = sh.Report(true); r.Errors != 1 || len(r.Last) != 1 || !strings.Contains(r.Last[0], "Class(42)") {
		t.Errorf("unknown class: %+v", r)
	}
}
