package combat

import (
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// frozenArmorRow is the Skills.txt row of Frozen Armor (Param3 3000, Param4 300; auralencalc ln34 + synergy).
func frozenArmorRow(withDuration bool) *data.Skill {
	s := &data.Skill{ID: 40, Key: "Frozen Armor", Name: "Frozen Armor", Mana: 7, ManaShift: 8, MinMana: 1}
	if withDuration {
		s.AuraLenCalc = "ln34+(skill('Shiver Armor'.blvl)+skill('Chilling Armor'.blvl))*par7"
		s.Param[2], s.Param[3], s.Param[6] = 3000, 300, 250
	}
	return s
}

// armorRun plays a 10 minute Frozen Armor life at 1 s steps: the armor is up for dur after each cast, the bot asks
// buffDue every step. state=false hides the state array (as before the state path). Returns the casts and the seconds
// the armor was down.
func armorRun(t *testing.T, withDuration bool, level int, dur time.Duration, state bool) (casts, down int) {
	t.Helper()
	h := newHarness(t)
	cat := testCatalog()
	cat.Skills[40] = frozenArmorRow(withDuration)
	h.f.SetCatalog(cat)
	p := &d2.Player{ID: 1, Class: d2.Sorceress, Level: 10, Skills: []d2.Skill{{ID: 40, Level: level}}}
	sp := SkillSpec{ID: 40, Name: "Frozen Armor", Level: level}
	var castAt time.Time
	up := func() bool { return !castAt.IsZero() && h.f.now().Sub(castAt) < dur }
	for i := 0; i < 600; i++ {
		p.States = d2.StateSet{}
		if state && up() {
			p.States[d2.StateFrozenArmor/32] |= 1 << (d2.StateFrozenArmor % 32)
		}
		if h.f.buffDue(p, sp) {
			castAt = h.f.now()
			h.f.buffAt[sp.ID] = castAt
			casts++
		} else if !up() {
			down++
		}
		h.s.Sleep(time.Second)
	}
	return casts, down
}

// Before the fix: a 150 s timer on a 120 s armor leaves it down 30 s of every cycle (level 1), 18 s at level 2.
func TestFrozenArmorTimerLeavesGap(t *testing.T) {
	_, down := armorRun(t, false, 1, 120*time.Second, false)
	if down < 100 {
		t.Fatalf("old behaviour should leave a gap, down %d s", down)
	}
}

func TestFrozenArmorRecastBeforeExpiryByClock(t *testing.T) {
	for _, c := range []struct {
		level int
		dur   time.Duration
	}{{1, 120 * time.Second}, {2, 132 * time.Second}} {
		casts, down := armorRun(t, true, c.level, c.dur, false)
		if down != 0 {
			t.Errorf("level %d: armor down %d s (casts %d)", c.level, down, casts)
		}
		// one cast per (duration - 6 s margin)
		if want := 600 / int((c.dur - 6*time.Second).Seconds()); casts < want || casts > want+2 {
			t.Errorf("level %d: %d casts, want about %d", c.level, casts, want)
		}
	}
}

func TestFrozenArmorRecastBeforeExpiryByState(t *testing.T) {
	_, down := armorRun(t, true, 1, 120*time.Second, true)
	if down != 0 {
		t.Fatalf("armor down %d s with the state readable", down)
	}
}

// Learned state: the armor killed early (state gone) is recast after the 4 s grace, not at the clock limit.
func TestFrozenArmorStateGoneRecastsEarly(t *testing.T) {
	h := newHarness(t)
	cat := testCatalog()
	cat.Skills[40] = frozenArmorRow(true)
	h.f.SetCatalog(cat)
	p := &d2.Player{Skills: []d2.Skill{{ID: 40, Level: 1}}}
	sp := SkillSpec{ID: 40, Name: "Frozen Armor", Level: 1}
	h.f.buffAt[40] = h.f.now()
	p.States[0] |= 1 << d2.StateFrozenArmor
	if h.f.buffDue(p, sp) || !h.f.necro.learned[d2.StateFrozenArmor] {
		t.Fatal("up and fresh: not due, learned")
	}
	p.States = d2.StateSet{}
	h.s.Sleep(5 * time.Second)
	if !h.f.buffDue(p, sp) {
		t.Fatal("state gone: recast")
	}
}

func TestBuffLimit(t *testing.T) {
	h := newHarness(t)
	cat := testCatalog()
	cat.Skills[40] = frozenArmorRow(true)
	h.f.SetCatalog(cat)
	for _, c := range []struct {
		lvl  int
		want time.Duration
	}{{1, 114 * time.Second}, {2, 126 * time.Second}, {5, 150 * time.Second}} { // L5: 4200 frames = 168 s, capped by the knob
		if got, ok := h.f.buffLimit(SkillSpec{ID: 40, Level: c.lvl}); !ok || got != c.want {
			t.Errorf("level %d: %v %v, want %v", c.lvl, got, ok, c.want)
		}
	}
	h.f.e.Tune.Combat.BuffMarginSec = 0
	if got, _ := h.f.buffLimit(SkillSpec{ID: 40, Level: 1}); got != 120*time.Second {
		t.Errorf("margin 0: %v", got)
	}
	h.f.e.Tune.Combat.BuffRecastSec = 100
	if got, _ := h.f.buffLimit(SkillSpec{ID: 40, Level: 1}); got != 100*time.Second {
		t.Errorf("knob stays the upper bound: %v", got)
	}
}

func TestBuffLimitFallbackNoDuration(t *testing.T) {
	h := newHarness(t)
	if got, ok := h.f.buffLimit(SkillSpec{ID: 40, Level: 1}); ok || got != 150*time.Second {
		t.Fatalf("no auralencalc: %v %v", got, ok)
	}
}
