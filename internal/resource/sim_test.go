package resource

import (
	"fmt"
	"testing"
)

func TestReplayBloodMoor(t *testing.T) {
	sc, tu := BloodMoorSorc(), DefaultTuning()
	res := Replay(sc, ResourcePolicy(), tu)
	leg := Replay(sc, LegacyPolicy(), tu)
	t.Logf("\n%+v\n%+v", leg, res)
	if res2 := Replay(sc, ResourcePolicy(), tu); res2 != res {
		t.Fatal("replay is not deterministic")
	}
	if leg.StarvedSec < 30 {
		t.Fatalf("scenario should starve the legacy policy, starved %.0fs", leg.StarvedSec)
	}
	if res.StarvedSec > leg.StarvedSec/4 {
		t.Fatalf("policy starved %.0fs vs legacy %.0fs", res.StarvedSec, leg.StarvedSec)
	}
	if res.Deaths > leg.Deaths || res.Deaths > 0 {
		t.Fatalf("deaths %d (legacy %d)", res.Deaths, leg.Deaths)
	}
	if res.KillsPerMin < leg.KillsPerMin*0.9 {
		t.Fatalf("policy %.2f kills/min fell below 90%% of legacy %.2f", res.KillsPerMin, leg.KillsPerMin)
	}
}

// The conclusion must not hinge on the U regen figure: halve and double it.
func TestReplayRegenSensitivity(t *testing.T) {
	for _, secs := range []float64{50, 100, 200} {
		sc := BloodMoorSorc()
		sc.ManaFullSecs = secs
		res := Replay(sc, ResourcePolicy(), DefaultTuning())
		leg := Replay(sc, LegacyPolicy(), DefaultTuning())
		t.Run(fmt.Sprint(secs), func(t *testing.T) {
			t.Logf("legacy kills=%d starved=%.0f deaths=%d | policy kills=%d starved=%.0f deaths=%d melee=%d rest=%.0f",
				leg.Kills, leg.StarvedSec, leg.Deaths, res.Kills, res.StarvedSec, res.Deaths, res.MeleeHits, res.RestSec)
			if res.StarvedSec > leg.StarvedSec+1 || res.Deaths > leg.Deaths {
				t.Fatal("policy worse than legacy")
			}
		})
	}
}

// SIMCLASS: the simulator names the class's own attack and weapon skill; the Sorceress defaults stay.
func TestScenarioSkillNames(t *testing.T) {
	sc := BloodMoorSorc()
	if sc.mainAttack() != "Fire Bolt" || sc.weaponSkill() != "Staff" {
		t.Errorf("Sorceress defaults changed: %s / %s", sc.mainAttack(), sc.weaponSkill())
	}
	w := &world{sc: BloodMoorPaladin()}
	in := w.input(1, 0)
	if len(in.Skills) != 2 || in.Skills[0].Name != "Zeal" || in.Skills[0].Cost != 2 || in.Skills[1].Name != "Attack" || !in.Skills[1].Melee {
		t.Errorf("Paladin skills %+v", in.Skills)
	}
	if got := BloodMoorSorc().WithLadder([]string{"Zeal", "Smite", "Attack"}); got.MainAttack != "Zeal" || got.WeaponSkill != "Attack" {
		t.Errorf("WithLadder: %+v", got)
	}
}

// The Paladin stock replay: the Zeal rung is cast while mana lasts, then the plain swing takes over; deterministic.
func TestBloodMoorPaladinReplay(t *testing.T) {
	r := Replay(BloodMoorPaladin(), ResourcePolicy(), DefaultTuning())
	t.Logf("Paladin stock: kills %d casts %d melee %d deaths %d trips %d lifePots %d waits %d", r.Kills, r.Casts, r.MeleeHits, r.Deaths, r.Trips, r.LifePots, r.Waits)
	if r.Kills == 0 || r.Casts == 0 || r.MeleeHits == 0 {
		t.Errorf("ladder not walked: %+v", r)
	}
	if again := Replay(BloodMoorPaladin(), ResourcePolicy(), DefaultTuning()); again != r {
		t.Error("not deterministic")
	}
}

// The Sorceress stock replay is unchanged by SIMCLASS (SKMELEEBOT/WIRE: 22 kills, 28 swings, 66 casts).
func TestSorceressStockReplayPinned(t *testing.T) {
	r := Replay(BloodMoorSorc(), ResourcePolicy(), DefaultTuning())
	if r.Kills != 22 || r.MeleeHits != 28 || r.Casts != 66 {
		t.Errorf("kills %d melee %d casts %d, want 22/28/66", r.Kills, r.MeleeHits, r.Casts)
	}
}
