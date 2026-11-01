package resource

import (
	"fmt"
	"testing"

	"d2lod-bot/internal/d2"
)

// PALPROBE part B (e) and (f): the Paladin belt plan and the melee model of the resource simulator against a
// Sorceress staff rung. Model inputs are U unless a comment says otherwise: the Blood Moor foe, monster damage and
// the Paladin's to-hit factor are the nearest defaults of BloodMoorSorc, not measurements.

func TestPalProbeBeltPlan(t *testing.T) {
	for _, lv := range []int{1, 8, 12, 13, 30} {
		p := PlanFor(d2.Paladin, lv)
		hp, mp := 0, 0
		for _, c := range p.Cols {
			switch c {
			case PotionLife:
				hp++
			case PotionMana:
				mp++
			}
		}
		// Paladin: ManaNeed 1 (occasional): three healing columns and one mana column; the Sorceress has two mana columns
		if hp != 3 || mp != 1 {
			t.Errorf("clvl %d: %d life and %d mana columns, want 3 and 1 (%v)", lv, hp, mp, p.Cols)
		}
		if s := PlanFor(d2.Sorceress, lv); s.Cols == p.Cols {
			t.Errorf("clvl %d: Paladin and Sorceress share a belt plan", lv)
		}
	}
	// Normal Act 1 caps at tier 2 up to clvl 12 and 4 up to 30, like every class
	if PlanFor(d2.Paladin, 8).MaxTier != 2 || PlanFor(d2.Paladin, 20).MaxTier != 4 {
		t.Error("tier caps")
	}
}

// The endgame Hammerdin table is not on the mana-heavy list although Blessed Hammer drains one mana column in Hell:
// recorded as a finding, not fixed here.
func TestPalProbeHammerdinManaHeavy(t *testing.T) {
	if !ManaHeavyBuild("paladin_hammerdin_endgame90") {
		t.Skip("PALPROBE finding F5: paladin_hammerdin_endgame90 is not in manaHeavyBuilds (resource/manaheavy.go); Hell and Nightmare get the plain Paladin belt (hp, hp, hp, mp) although Blessed Hammer at 5 to 10 mana a cast empties one column; U until a Hell soak")
	}
}

// palStats is a Paladin of the v2 plan at clvl L with a short sword (2-7) and a buckler: Str +2 and Vit +3 a level.
func palScenario(l int) Scenario {
	life := 55 + (l-1)*2 + (l-1)*3*3 // 2 a level, 3 Vit a level at 3 life each (charstats; matches the observed 88 at clvl 4)
	mana := 15 + (l-1)*3/2
	str := 25 + 2*(l-1)
	return Scenario{
		Name: fmt.Sprintf("Blood Moor, clvl %d Paladin, melee only", l), Seconds: 300, Class: d2.Paladin, Level: l,
		MaxLife: life, MaxMana: mana, StartMana: mana, Gold: 120, HP: 3, MP: 0, TravelSec: 9, PackMin: 1, PackMax: 4,
		MonsterHP: []int{12, 24, 24, 45}, MonsterDps: 1.6, Exposure: 1, // in contact for the whole fight (no kiting)
		GoldPerKill: 6, SpellCost: 999, SpellDmg: 0, CastSec: 0.45, // no ranged spell: the cast slot is never affordable
		MeleeSec: 0.6, TripSec: 75, Seed: 7, HasMelee: true, Belt: 1,
		// To-hit: AR (dex-7) x 5 + factor, factor 20 is U (charstats ToHitFactor not read here); damage: Str through
		// the sword's StrBonus 100 (U: short sword in 1.14d), no skill ED yet.
		Blow: &Blow{AR: AttackRating(0, 20, 20), Level: l, Weapon: true, Min: 2, Max: 7, Str: str, DmgPct: StatDmgPct(0, str, 20, 100, 0, 0)},
		Foe:  Foe{Defense: 15, Level: 3},
	}
}

func sorcStaffOnly() Scenario {
	sc := BloodMoorSorc()
	sc.Name = "Blood Moor, clvl 8 Sorceress, staff only"
	sc.SpellCost = 999
	return sc
}

func TestPalProbeMeleeModelVersusStaffRung(t *testing.T) {
	tu := DefaultTuning()
	sorc := Replay(sorcStaffOnly(), ResourcePolicy(), tu)
	stock := Replay(BloodMoorSorc(), ResourcePolicy(), tu)
	t.Logf("Sorceress clvl 8 stock (Fire Bolt + staff): kills %d (%.1f/min) deaths %d trips %d lifePots %d", stock.Kills, stock.KillsPerMin, stock.Deaths, stock.Trips, stock.LifePots)
	t.Logf("Sorceress clvl 8 staff rung only: kills %d (%.1f/min) swings %d deaths %d trips %d lifePots %d retreats %d", sorc.Kills, sorc.KillsPerMin, sorc.MeleeHits, sorc.Deaths, sorc.Trips, sorc.LifePots, sorc.Retreats)
	for _, l := range []int{1, 4, 8} {
		sc := palScenario(l)
		r := Replay(sc, ResourcePolicy(), tu)
		exp := sc.Blow.Expected(sc.Foe)
		t.Logf("Paladin clvl %d: life %d, %.2f dmg/swing (hit %d%%): kills %d (%.1f/min) swings %d deaths %d trips %d chickens %d lifePots %d retreats %d",
			l, sc.MaxLife, exp, sc.Blow.HitPct(sc.Foe), r.Kills, r.KillsPerMin, r.MeleeHits, r.Deaths, r.Trips, r.Chickens, r.LifePots, r.Retreats)
		if r.Kills == 0 {
			t.Errorf("clvl %d Paladin killed nothing in the model", l)
		}
	}
	// at the same clvl the Paladin out-damages the staff per swing (sword 2-7 and Str 39 against staff 1-5 and Str 10);
	// at clvl 1 the model's to-hit (foe level 3, U) is only 42 percent, so it is below the clvl 8 staff
	p8 := palScenario(8)
	s8 := BloodMoorSorc()
	if p8.Blow.Expected(p8.Foe) <= s8.Blow.Expected(s8.Foe) {
		t.Errorf("clvl 8 sword %.2f per swing is not above the staff %.2f", p8.Blow.Expected(p8.Foe), s8.Blow.Expected(s8.Foe))
	}
}

// Early-game risk: the expected life lost killing one pack in contact, by clvl, against the line the melee profile
// retreats at (retreatLifePct 40, classdata profiles.json). Closed form of the model: each monster in contact costs
// MonsterDps per second until it dies, the pack dies front first.
func TestPalProbeEarlyLifeCurve(t *testing.T) {
	const retreatPct = 40 // profiles.json paladin melee.retreatLifePct (V file read)
	for _, l := range []int{1, 2, 3, 4, 6, 8} {
		sc := palScenario(l)
		perSwing := sc.Blow.Expected(sc.Foe)
		for _, size := range []int{1, 2, 4} {
			lost, alive := 0.0, size
			for k := 0; k < size; k++ {
				hp := float64(sc.MonsterHP[k%len(sc.MonsterHP)])
				secs := hp / perSwing * sc.MeleeSec
				lost += float64(alive) * sc.MonsterDps * secs
				alive--
			}
			pct := lost / float64(sc.MaxLife) * 100
			flag := ""
			if pct > 100-retreatPct {
				flag = "  crosses the retreat line"
			}
			t.Logf("clvl %d life %d, pack of %d: %.0f life lost (%.0f%%)%s", l, sc.MaxLife, size, lost, pct, flag)
		}
	}
	// a single Blood Moor monster must never take the clvl 1 Paladin through the retreat line in the model
	sc := palScenario(1)
	hp := float64(sc.MonsterHP[3])
	if lost := hp / sc.Blow.Expected(sc.Foe) * sc.MeleeSec * sc.MonsterDps; lost > float64(sc.MaxLife)*0.6 {
		t.Errorf("one 45 hp monster costs %.0f of %d life at clvl 1", lost, sc.MaxLife)
	}
}
