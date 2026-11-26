package arbiter

import "testing"

const questyOutcomes = "../../tools/provingground/arbiter-outcomes/questy-2026-11-25.json"

func hardOnly(n string) bool { return n == "potion-not-drunk" }

func TestQuestyOutcomes(t *testing.T) {
	f, err := LoadOutcomes(questyOutcomes)
	if err != nil {
		t.Fatal(err)
	}
	var scores []CaseScore
	for _, c := range f.Cases {
		scores = append(scores, f.Score(c, DefaultConfig()))
	}
	got := map[string]Verdict{}
	for _, v := range Judge(scores, hardOnly) {
		got[v.Family] = v
	}
	if v := got["survive"]; !v.Pass || v.ChainHard != 1 || v.ArbHard != 0 {
		t.Fatalf("survive verdict %+v", v)
	}
	if v := got["chores"]; !v.Pass || v.ChainSoft != 3 || v.ArbSoft != 0 {
		t.Fatalf("chores verdict %+v", v)
	}
}

// Two builds: a different drink line (a Paladin drinks later or earlier than a Sorceress) is data in the facts, so the
// verdict must not depend on one class's number. MUST-FAIL half: when the arbiter has no survive source reading the line
// (survive_below 0 = no flee line) it picks the same family as the chain and the case scores no better.
func TestSurviveOutcomeTwoBuildsAndMustFail(t *testing.T) {
	f, err := LoadOutcomes(questyOutcomes)
	if err != nil {
		t.Fatal(err)
	}
	base := f.Cases[0]
	for _, line := range []int{50, 70} {
		c := base
		c.Facts = map[string]int{}
		for k, v := range base.Facts {
			c.Facts[k] = v
		}
		c.Facts[FactSurviveBelow] = line
		if s := f.Score(c, DefaultConfig()); s.Arbiter != "survive" || len(s.ArbiterHits) != 0 || len(s.ChainHits) != 1 {
			t.Fatalf("line %d: %+v", line, s)
		}
	}
	c := base
	c.Facts = map[string]int{"in_town": 0, "life_pct": 30, "hostile": 12, "potions": 3, "trip": 1} // no drink line known
	s := f.Score(c, DefaultConfig())
	if s.Arbiter == "survive" {
		t.Fatalf("without a drink line the arbiter must not claim a reflex: %+v", s)
	}
	for _, v := range Judge([]CaseScore{s}, hardOnly) {
		if v.Pass {
			t.Fatalf("a case with no improvement must not pass: %+v", v)
		}
	}
}

func TestJudgeRejectsNewKind(t *testing.T) {
	s := CaseScore{Family: "x", ChainHits: []string{"xp-frozen"}, ArbiterHits: []string{"corpse-unresolved"}, NewKinds: []string{"corpse-unresolved"}}
	if v := Judge([]CaseScore{s}, func(string) bool { return true })[0]; v.Pass {
		t.Fatalf("a new kind of hit must fail: %+v", v)
	}
}

func TestLoadOutcomesRejectsBadCondition(t *testing.T) {
	if _, _, _, err := parseCond("hostile"); err == nil {
		t.Fatal("no operator must be an error")
	}
}

// Entry 1 (trip, chores, claim): the stranded-walk and the pending-claim cases pass for the arbiter. Two builds: the belt
// refill line is data (2 or 6 potions), the verdict does not depend on one number. MUST-FAIL: with no trip and no claim in
// the facts the arbiter has nothing to add, so the case scores no better and the verdict fails.
func TestChoresClaimOutcomes(t *testing.T) {
	f, err := LoadOutcomes(questyOutcomes)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"arb.questy-stranded-walk-home", "arb.questy-claim-pending-quiet-field"} {
		var c OutcomeCase
		for _, x := range f.Cases {
			if x.ID == id {
				c = x
			}
		}
		for _, min := range []int{2, 6} {
			cc := c
			cc.Facts = map[string]int{}
			for k, v := range c.Facts {
				cc.Facts[k] = v
			}
			cc.Facts[FactPotionsMin] = min
			s := f.Score(cc, DefaultConfig())
			for _, v := range Judge([]CaseScore{s}, hardOnly) {
				if !v.Pass || len(s.ChainHits) == 0 || len(s.ArbiterHits) != 0 {
					t.Fatalf("%s min %d: %+v %+v", id, min, s, v)
				}
			}
		}
		bare := c
		bare.Facts = map[string]int{"in_town": 0, "life_pct": 100, "potions": 4, "potions_min": 4, "survive_below": 50}
		s := f.Score(bare, DefaultConfig())
		for _, v := range Judge([]CaseScore{s}, hardOnly) {
			if v.Pass {
				t.Fatalf("%s: with nothing wanted the arbiter must not look better: %+v", id, s)
			}
		}
	}
}

// Entry 2 (goldfarm, recovery): the poverty spiral and the quest walk while broke score better for the arbiter; the
// player's farm order stays above recovery (arbiter and chain agree, no hit either way). Two builds: the refill line is
// data. MUST-FAIL: without the recovery fact the arbiter proposes the refill again and the spiral case does not pass.
func TestGoldFarmRecoveryOutcomes(t *testing.T) {
	f, err := LoadOutcomes(questyOutcomes)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]OutcomeCase{}
	for _, c := range f.Cases {
		byID[c.ID] = c
	}
	for _, id := range []string{"arb.questy-poverty-spiral", "arb.questy-poverty-quest-walk"} {
		for _, min := range []int{2, 6} {
			c := byID[id]
			c.Facts = map[string]int{}
			for k, v := range byID[id].Facts {
				c.Facts[k] = v
			}
			c.Facts[FactPotionsMin] = min
			s := f.Score(c, DefaultConfig())
			if s.Arbiter != "recovery" || len(s.ChainHits) == 0 || len(s.ArbiterHits) != 0 {
				t.Fatalf("%s min %d: %+v", id, min, s)
			}
		}
		c := byID[id]
		c.Facts = map[string]int{}
		for k, v := range byID[id].Facts {
			if k != FactRecovery {
				c.Facts[k] = v
			}
		}
		if s := f.Score(c, DefaultConfig()); s.Arbiter == "recovery" {
			t.Fatalf("%s: without recovery the arbiter must not farm: %+v", id, s)
		}
	}
	if s := f.Score(byID["arb.goldfarm-order-broke-trip"], DefaultConfig()); s.Arbiter != "goldfarm" || len(s.ArbiterHits) != 0 || len(s.ChainHits) == 0 {
		t.Fatalf("the order must beat the unpayable trip: %+v", s)
	}
	if s := f.Score(byID["arb.goldfarm-order-over-recovery"], DefaultConfig()); s.Arbiter != "goldfarm" || len(s.NewKinds) != 0 {
		t.Fatalf("the order must stay above recovery: %+v", s)
	}
}

// Entry 3 (quest): a standing typed refusal silences the quest source (the arbiter takes the filler instead of asking
// again), and an affordable refill comes before the quest walk. Two builds (refill line 2 or 6). MUST-FAIL: without the
// refusal fact the quest source asks again and the case scores no better.
func TestQuestOutcomes(t *testing.T) {
	f, err := LoadOutcomes(questyOutcomes)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]OutcomeCase{}
	for _, c := range f.Cases {
		byID[c.ID] = c
	}
	for _, min := range []int{2, 6} {
		c := byID["arb.questy-quest-refusal-storm"]
		c.Facts = map[string]int{}
		for k, v := range byID["arb.questy-quest-refusal-storm"].Facts {
			c.Facts[k] = v
		}
		c.Facts[FactPotionsMin] = min
		c.Facts[FactPotions] = min
		s := f.Score(c, DefaultConfig())
		if s.Arbiter != "level" || len(s.ChainHits) != 1 || len(s.ArbiterHits) != 0 {
			t.Fatalf("min %d: %+v", min, s)
		}
		c.Facts[FactQuestRefused] = 0
		s = f.Score(c, DefaultConfig())
		if s.Arbiter != "quest" || len(s.ChainHits) != 0 {
			t.Fatalf("without the refusal the quest source asks again: %+v", s)
		}
	}
	if s := f.Score(byID["arb.questy-quest-walk-affordable-refill"], DefaultConfig()); s.Arbiter != "chores" || len(s.ChainHits) != 1 || len(s.ArbiterHits) != 0 {
		t.Fatalf("%+v", s)
	}
}
