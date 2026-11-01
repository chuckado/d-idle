package combat

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
)

// longFight plays a 600 s fight at 1 s steps with no need-buff edge: only fightBuff (the fight loop's check) may
// recast. The armor is up for dur after each cast; states=true shows it in Player.States as game.exe and d2d (with
// Options.BuffStates) do. adjacent monsters stand next to the player the whole time.
func longFight(t *testing.T, level int, dur time.Duration, states bool, adjacent int) (casts, down int) {
	t.Helper()
	h, w := newNecro(t)
	cat := necroCat()
	cat.Skills[40] = frozenArmorRow(true)
	h.f.SetCatalog(cat)
	p := w.snap.Player
	p.Class, p.RightSkill, p.Mana, p.MaxMana = d2.Sorceress, 40, 400, 400
	p.Skills = []d2.Skill{{ID: 40, Level: level}}
	h.f.SetHotkeys(map[int]string{40: "f2"})
	for i := 0; i < adjacent; i++ {
		w.snap.Monsters = append(w.snap.Monsters, mon(uint32(900+i), 1, p.X+1, p.Y))
	}
	var castAt time.Time
	up := func() bool { return !castAt.IsZero() && h.f.now().Sub(castAt) < dur }
	for i := 0; i < 600; i++ {
		p.States = d2.StateSet{}
		if states && up() {
			p.States[d2.StateFrozenArmor/32] |= 1 << (d2.StateFrozenArmor % 32)
		}
		snap, _ := h.f.snapFn()
		n := len(w.casts)
		if _, err := h.f.fightBuff(context.Background(), snap, Hostiles(snap, h.f.hostileCtx(snap.Player))); err != nil {
			t.Fatal(err)
		}
		if len(w.casts) > n {
			castAt = h.f.now()
			casts++
		} else if !up() {
			down++
		}
		h.s.Sleep(time.Second)
	}
	return casts, down
}

func TestFightBuffRecastsInLongFight(t *testing.T) {
	for _, c := range []struct {
		lvl    int
		dur    time.Duration
		states bool
		casts  int
	}{
		{1, 120 * time.Second, false, 6}, {2, 132 * time.Second, false, 5},
		{1, 120 * time.Second, true, 6}, {2, 132 * time.Second, true, 5},
	} {
		casts, down := longFight(t, c.lvl, c.dur, c.states, 0)
		if down != 0 {
			t.Errorf("L%d states=%v: armor down %d s", c.lvl, c.states, down)
		}
		if casts < c.casts-1 || casts > c.casts+1 {
			t.Errorf("L%d states=%v: %d casts, want about %d", c.lvl, c.states, casts, c.casts)
		}
	}
}

func TestFightBuffStateLostRecastsAtOnce(t *testing.T) {
	// Energy Shield style: the state vanishes long before the clock; learned state + 4 s grace recasts.
	h, w := newNecro(t)
	cat := necroCat()
	cat.Skills[40] = frozenArmorRow(true)
	h.f.SetCatalog(cat)
	p := w.snap.Player
	p.Class, p.RightSkill, p.Mana, p.MaxMana = d2.Sorceress, 40, 400, 400
	p.Skills = []d2.Skill{{ID: 40, Level: 1}}
	h.f.SetHotkeys(map[int]string{40: "f2"})
	run := func() int {
		snap, _ := h.f.snapFn()
		n := len(w.casts)
		h.f.fightBuff(context.Background(), snap, nil)
		return len(w.casts) - n
	}
	if run() != 1 {
		t.Fatal("first cast")
	}
	p.States[d2.StateFrozenArmor/32] |= 1 << (d2.StateFrozenArmor % 32)
	h.s.Sleep(10 * time.Second)
	if run() != 0 {
		t.Fatal("up: no cast")
	}
	p.States = d2.StateSet{}
	if run() != 1 {
		t.Fatal("state lost: recast")
	}
}

func TestFightBuffNotWhenSurrounded(t *testing.T) {
	casts, _ := longFight(t, 1, 120*time.Second, false, crowdAdjacentMax)
	if casts != 0 {
		t.Fatalf("cast %d times with %d attackers in reach", casts, crowdAdjacentMax)
	}
}
