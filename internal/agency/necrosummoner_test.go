package agency

import (
	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"strings"
	"testing"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/power"
)

// necro1: the shadow planner gives a sane plan for the Summoner table at every level from 1 to 30, with the new
// Necromancer rows, and the same invariants hold for a second class (the Sorceress fire table), so the check is not
// Necromancer-only. Every unsupported skill stays out of every slot (TestUnsupportedNeverChosenAndAlwaysFlagged runs
// the unsupported Necromancer rows too).
func TestNecromancerSummonerShadowPlan(t *testing.T) {
	env := testEnv(t)
	pl := &Planner{Env: env}
	for _, c := range []struct{ table, class, arch string }{
		{"necromancer_summoner_guide", "necromancer", "summoner"},
		{"sorceress_fire_normal30", "sorceress", ""},
	} {
		for clvl := 1; clvl <= 30; clvl++ {
			a, comb := allocFromTable(t, c.table, clvl)
			v, p, err := pl.Evaluate(Input{Hero: "Testy", Alloc: a, Combat: comb, Weapon: Weapon{Class: "1HS", Avg: 20}})
			if err != nil {
				t.Fatalf("%s clvl %d: %v", c.table, clvl, err)
			}
			line := ShadowLine(v, p)
			if !strings.Contains(line, "class="+c.class) || !strings.Contains(v.Line, "your build:") || !strings.Contains(v.Line, "viability: not judged") {
				t.Errorf("%s clvl %d: line %q / view %q", c.table, clvl, line, v.Line)
			}
			for _, ch := range append(append([]*Choice{p.Main, p.Attack, p.Secondary}, ptrs(p.Summons)...), ptrs(p.Buffs)...) {
				if ch != nil && env.Caps.Skills[ch.Skill].Tier == Unsupported {
					t.Errorf("%s clvl %d: planner chose the unsupported %s", c.table, clvl, ch.Skill)
				}
			}
			if c.arch != "" && clvl >= 4 { // a summon owned, so the archetype resolves
				if p.Archetype != c.arch || p.Main == nil || p.Main.Role != RoleSummon {
					t.Errorf("%s clvl %d: archetype %q main %+v", c.table, clvl, p.Archetype, p.Main)
				}
				if len(p.Summons) == 0 {
					t.Errorf("%s clvl %d: no summon in the plan", c.table, clvl)
				}
			}
			if c.arch != "" && clvl == 30 {
				t.Logf("%s clvl 30: %s | warnings %s", c.table, line, warnCodes(p))
				// Hell-only points (Fire Golem chain, Revive) are not there yet; the prerequisite curses are
				// not taken before clvl 50, so nothing at 30 is an unsupported spend.
				if p.Has(WUnsupportedPts) {
					t.Errorf("%s clvl 30: unsupported points in a Normal table (%s)", c.table, warnCodes(p))
				}
			}
		}
	}
	// An unknown class is an error that names it, never a Sorceress plan.
	if _, err := Derive(alloc("witch", 10, map[string]int{"Raise Skeleton": 5}), env); err == nil || !strings.Contains(err.Error(), "witch") {
		t.Errorf("unknown class: %v", err)
	}
}

// The late game of the same table: the points the engine cannot cast are flagged, never played.
func TestNecromancerSummonerLateGameFlagsUnsupported(t *testing.T) {
	env := testEnv(t)
	a, _ := allocFromTable(t, "necromancer_summoner_guide", 90)
	p := mustDerive(t, a, env)
	if !p.Has(WUnsupportedPts) {
		t.Fatalf("clvl 90 holds Decrepify, Terror and golem prerequisite points: expected %s, got %s", WUnsupportedPts, warnCodes(p))
	}
	for _, n := range []string{"Decrepify", "Terror", "BloodGolem", "IronGolem"} {
		if p.Unplayed.Skills[n] != 1 {
			t.Errorf("%s: unplayed %v", n, p.Unplayed.Skills)
		}
	}
}

// Viability from the survival model for a clvl 20 Summoner: a hero with little life finds Act 1 harder than one with a lot.
func TestNecromancerSummonerViability(t *testing.T) {
	env := testEnv(t)
	m, err := power.NewModel(danger.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	pl := &Planner{Env: env, Model: m, Tab: madeUpTable()}
	a, comb := allocFromTable(t, "necromancer_summoner_guide", 20)
	hero := func(life int) power.Hero {
		h := power.Hero{Class: "Necromancer", Level: 20, MaxLife: life, MaxMana: 120, Defence: 60, AR: 100, DmgMin: 5, DmgMax: 9, Potions: 6, Resist: [4]int{0, 0, 0, 0}}
		for _, term := range []string{power.TermLife, power.TermDefence, power.TermBlock, power.TermResist, power.TermOffence, power.TermPotions} {
			h.SetQuality(term, power.Measured)
		}
		return h
	}
	in := func(h power.Hero) Input {
		return Input{Hero: "Testy", Alloc: a, Combat: comb, Weapon: Weapon{Class: "1HS", Avg: 8}, Power: h, Diff: "normal", Act: 1}
	}
	weak, _, err := pl.Evaluate(in(hero(40)))
	if err != nil {
		t.Fatal(err)
	}
	strong, _, err := pl.Evaluate(in(hero(900)))
	if err != nil {
		t.Fatal(err)
	}
	if weak.Viability == nil || strong.Viability == nil {
		t.Fatalf("no viability: %q / %q", weak.Note, strong.Note)
	}
	if wordRank[strong.Viability.Word] > wordRank[weak.Viability.Word] {
		t.Errorf("more life made the act worse: weak %s strong %s", weak.Viability.Word, strong.Viability.Word)
	}
	if !strings.Contains(strong.Line, "your build:") || !strings.Contains(strong.Line, "viability:") {
		t.Errorf("line %q", strong.Line)
	}
	t.Logf("weak: %s\nstrong: %s", weak.Line, strong.Line)
}

// Agreement on the summon slot, where a summoner's plan lives: the summons the planner names are the ones the summoner
// profile has a key for (so KeepSummons can cast them), and every owned summon the profile binds is in the plan.
// Fire Golem is the one exception by design: it has no F-key in the layout (eight keys), so Clay Golem keeps the slot.
func TestNecromancerSummonerPlanAgreesWithProfileOnSummons(t *testing.T) {
	env := testEnv(t)
	prof, err := combat.LookupProfile(d2.Necromancer, "summoner")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, clvl := range []int{6, 12, 18, 24, 30, 60, 90} {
		a, comb := allocFromTable(t, "necromancer_summoner_guide", clvl)
		env.Combat = combat.Build(comb)
		p := mustDerive(t, a, env)
		planned := map[string]bool{}
		for _, s := range p.Summons {
			planned[s.Skill] = true
			if _, ok := prof.Layout[s.Skill]; !ok && s.Skill != "Skeleton Mastery" && s.Skill != "Golem Mastery" && s.Skill != "FireGolem" && s.Skill != "Summon Resist" {
				t.Errorf("clvl %d: the plan summons %s, the profile has no key for it", clvl, s.Skill)
			}
		}
		for _, n := range []string{"Raise Skeleton", "Raise Skeletal Mage", "Clay Golem", "Revive"} {
			if a.Skills[n] > 0 && !planned[n] {
				t.Errorf("clvl %d: %s has %d points and a key but is not in the plan (summons %v)", clvl, n, a.Skills[n], planned)
			}
		}
		checked++
	}
	if checked != 7 {
		t.Fatal(checked)
	}
}
