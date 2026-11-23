package scenario

import (
	"slices"
	"strings"
	"testing"
)

// pgfix4: the ring of loot.rare-ring-classified was dropped in the town because "debug give" ran before "debug goto":
// a ground item is placed beside the hero, so it must come after the position verb. Both classes get the same order.
func TestGroundItemAfterPosition(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		s := Scenario{ID: "x.ring", Kind: KindBehaviour, Pre: Pre{Level: 10, Start: StartField, Area: 2,
			Items: []Item{{Code: "rin", Where: "ground", Quality: 6, Unid: true}}}}
		su, err := BuildSetup(s, class, 3)
		if err != nil {
			t.Fatal(err)
		}
		goto_, give := -1, -1
		for i, v := range su.Verbs {
			switch {
			case strings.HasPrefix(v.Line, "debug goto"):
				goto_ = i
			case strings.HasPrefix(v.Line, "debug give rin"):
				give = i
			}
		}
		if goto_ < 0 || give < 0 || give < goto_ {
			t.Errorf("%s: goto at %d, give at %d: the ring must be dropped after the hero is in the field", class, goto_, give)
		}
	}
}

// pre.stats writes the raw stat after the level verbs (a level write would otherwise reset it) and marks the run CHEATED.
func TestRawStatsFollowTheLevel(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		s := Scenario{ID: "x.stats", Kind: KindBehaviour, Pre: Pre{Level: 20, Stats: []StatSet{{Stat: 7, Value: 307200}}}}
		su, err := BuildSetup(s, class, 3)
		if err != nil {
			t.Fatal(err)
		}
		lvl, raw := -1, -1
		for i, v := range su.Verbs {
			switch v.Line {
			case "debug stat 12 20":
				lvl = i
			case "debug stat 7 307200":
				raw = i
			}
		}
		if lvl < 0 || raw < lvl {
			t.Errorf("%s: level verb at %d, raw stat at %d", class, lvl, raw)
		}
		if !slices.Contains(su.Tags, "CHEATED") {
			t.Errorf("%s: a raw stat write must mark the run CHEATED", class)
		}
		if _, err := BuildSetup(Scenario{ID: "x.bad", Kind: KindBehaviour, Pre: Pre{Stats: []StatSet{{Stat: 9999}}}}, class, 3); err == nil {
			t.Errorf("%s: stat id 9999 accepted", class)
		}
	}
}

// pre.waypoints -1 means every waypoint (`debug wp all`); a bit stays `debug wp <bit>`.
func TestWaypointAllAndBit(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		s := Scenario{ID: "x.wp", Kind: KindBehaviour, Pre: Pre{Waypoints: []int{-1, 3}}}
		su, err := BuildSetup(s, class, 3)
		if err != nil {
			t.Fatal(err)
		}
		var lines []string
		for _, v := range su.Verbs {
			lines = append(lines, v.Line)
		}
		if !slices.Contains(lines, "debug wp all") || !slices.Contains(lines, "debug wp 3") {
			t.Errorf("%s: verbs %v", class, lines)
		}
	}
}
