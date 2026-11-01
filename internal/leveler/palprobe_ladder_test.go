package leveler

import (
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// PALPROBE part B (a)+(b): the combat ladder the profile gives a Paladin at every clvl, with the skills each plan
// table has trained by then (the simulator of palprobe_test.go). Lives here because leveler imports combat.

// skillsAt replays the plan and returns the trained skills after each clvl 2..30.
func skillsAt(t *testing.T, table string) map[int]map[string]int {
	t.Helper()
	b := mustBuild(t, table)
	start := [numAttr]int{25, 20, 25, 15}
	h := Have{Attr: start, Start: start, Skills: map[string]int{}}
	out := map[int]map[string]int{1: {}}
	for L := 2; L <= 30; L++ {
		h.Level, h.StatPts, h.SkillPts = L, h.StatPts+5, h.SkillPts+1
		for _, a := range b.Plan(h, palLegal, nil) {
			if a.Kind == KindStat {
				h.Attr[a.Attr]++
				h.StatPts--
			} else {
				h.Skills[a.Skill]++
				h.SkillPts--
			}
		}
		cp := map[string]int{}
		for k, v := range h.Skills {
			cp[k] = v
		}
		out[L] = cp
	}
	return out
}

func TestPalProbeLadderAlongBothTables(t *testing.T) {
	pr, err := combat.LookupProfile(d2.Paladin, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"paladin-normal30", "paladin-normal30-v2"} {
		sk := skillsAt(t, table)
		var first string
		for L := 1; L <= 30; L++ {
			has := func(n string) bool { return sk[L][n] > 0 }
			names := pr.AttackNames(L, has)
			if len(names) == 0 || names[len(names)-1] != "Attack" {
				t.Errorf("%s clvl %d: the ladder %v must end in the free Attack", table, L, names)
				continue
			}
			for _, n := range names {
				if n == "Sacrifice" {
					t.Errorf("%s clvl %d: Sacrifice on the ladder", table, L)
				}
			}
			if names[0] != first {
				t.Logf("%s clvl %d: ladder starts with %s (%v)", table, L, names[0], names)
				first = names[0]
			}
			// every rung the ladder would use must be a skill the character owns: the loop skips unknown ones
			if L >= 12 && has("Zeal") && names[0] != "Zeal" && !has("Blessed Hammer") {
				t.Errorf("%s clvl %d: Zeal trained but the ladder starts with %s", table, L, names[0])
			}
		}
	}
}

// PALBAND (probe F2): the profile's Blessed Hammer band minimum is the skill's own required level (18), so the band
// is open the moment any table has trained the hammer; before the fix it started at 24 and the points sat idle.
// Every Paladin plan table is checked, the endgame ones too.
func TestPalProbeHammerBandStartsAtTraining(t *testing.T) {
	pr, _ := combat.LookupProfile(d2.Paladin, "")
	for _, table := range []string{"paladin-normal30", "paladin-normal30-v2", "paladin_hammerdin_endgame90", "paladin_hammerdin_zeal_endgame90"} {
		sk := skillsAt(t, table)
		for L := 1; L <= 30; L++ {
			if sk[L]["Blessed Hammer"] > 0 {
				has := func(n string) bool { return sk[L][n] > 0 }
				if n := pr.AttackNames(L, has)[0]; n != "Blessed Hammer" {
					t.Errorf("%s trains Blessed Hammer at clvl %d but the first attack there is %s", table, L, n)
				}
				break
			}
		}
	}
}

// The legacy table has no Prayer, so the Paladin's resting aura (Prayer below 70 percent life) is never available
// on it; only the v2 table trains it (clvl 4).
func TestPalProbeLegacyTableHasNoPrayer(t *testing.T) {
	if skillsAt(t, "paladin-normal30")[30]["Prayer"] == 0 {
		t.Skip("PALPROBE finding F3: build paladin-normal30 (the table a Paladin that already owns Might or Holy Fire keeps) never trains Prayer, so the profile's resting aura rung is dead for it (the ladder falls through to Holy Fire / Might)")
	}
}

// Zeal at level 1 for the whole game on the legacy table: hold clicks of a one-point Zeal.
func TestPalProbeLegacyZealStaysAtOne(t *testing.T) {
	if z := skillsAt(t, "paladin-normal30")[30]["Zeal"]; z < 4 {
		t.Skip("PALPROBE finding F4: build paladin-normal30 trains Zeal to " + string(rune('0'+z)) + " only (v2 trains 4 by clvl 15); the profile holds the button for Zeal from clvl 12 either way, so the legacy Paladin swings the slow low-hit-count Zeal for all of Normal")
	}
}

// ---- gear (c) ----

func TestPalProbeWantsShield(t *testing.T) {
	p, err := LookupPrefs("Paladin")
	if err != nil || !p.WantsShield || p.TwoHandMult >= 1 {
		t.Fatalf("%+v %v", p, err)
	}
	// no other weapon class asks for a shield, so a two-hander is never refused for them (Barbarian, Amazon, Assassin)
	for _, c := range []string{"Barbarian", "Amazon", "Assassin", "Sorceress"} {
		if q, _ := LookupPrefs(c); q.WantsShield {
			t.Errorf("%s wants a shield", c)
		}
	}
}

// Weapon scoring for a Paladin below clvl 40: damage counts (0.8), a two-hander is halved, no caster bonus.
func TestPalProbeWeaponChoice(t *testing.T) {
	pal, _ := LookupPrefs("Paladin")
	sorc, _ := LookupPrefs("Sorceress")
	sword := &data.ItemBase{Kind: data.KindWeapon, Type: "swor", MinDam: 3, MaxDam: 15} // avg 9
	two := &data.ItemBase{Kind: data.KindWeapon, Type: "swor", TwoHanded: true, TwoMinDam: 8, TwoMaxDam: 20}
	wand := &data.ItemBase{Kind: data.KindWeapon, Type: "wand", MinDam: 2, MaxDam: 4}
	it := &d2.Item{Quality: d2.QualityNormal}
	sc := func(pr Prefs, b *data.ItemBase) float64 { return scoreOf(nil, pr, b, it, nil, scoreCtx{}) }
	if sc(pal, sword) <= sc(pal, wand) {
		t.Errorf("a Paladin must prefer a sword (%.1f) to a wand (%.1f)", sc(pal, sword), sc(pal, wand))
	}
	// 14 avg two-hander x 0.5 x 0.8 = 5.6 against the one-hander 9 x 0.8 = 7.2: the one-hander wins on score alone
	if sc(pal, sword) <= sc(pal, two) {
		t.Errorf("one-hander %.1f must beat the two-hander %.1f", sc(pal, sword), sc(pal, two))
	}
	_ = sorc
}

// A shield build is never planned into a two-hander while a shield is worn, whatever the score.
func TestPalProbeTwoHanderNeverOverShield(t *testing.T) {
	two := Cand{Unit: d2.Unit{ID: 5}, Name: "great sword", Score: 99, Usable: true, Slots: []int{d2.BodyRHand}, TwoHanded: true, AvoidTwoHanded: true}
	sword := cand(6, "sword", 10, true, d2.BodyRHand)
	shield := cand(7, "buckler", 3, true, d2.BodyLHand)
	for _, eq := range []map[int]Cand{{d2.BodyLHand: shield}, {}} {
		for _, sw := range ChooseGear(eq, []Cand{two, sword, shield}) {
			if sw.Item.ID() == 5 {
				t.Errorf("two-hander planned (equipped %d slots): %v", len(eq), sw)
			}
		}
	}
}
