package combat

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/input"
)

// gateAt mirrors nav.MouthClickOK for one unwanted exit: a click target within 15 of it breaks the invariant.
type gateAt struct {
	Mover
	ex    input.Point
	asked int
}

func (m *gateAt) MouthClickOK(from, to input.Point) bool {
	m.asked++
	dx, dy := to.X-m.ex.X, to.Y-m.ex.Y
	return dx*dx+dy*dy >= 225
}

// FIXMOUTH4, V (DRILLS18): basic-attack swings at a Fallen by the mouth walked the character from 13 to 3.2 subtiles from
// the exit. With the gate refusing, the swing is held (no left click) and the fight still spends time.
func TestSwingHeldTowardMouth(t *testing.T) {
	h := newHarness(t)
	h.s.mana, h.s.manaRegen, h.s.meleeDmg = 0, 0, 20
	h.s.addMon(10, clsFallen, 104, 100, 30, 1, 0) // 4 away: a swing walks
	mv := &gateAt{Mover: h.s, ex: input.Point{X: 108, Y: 100}}
	h.f.mv = mv
	t0 := h.s.Now()
	_, _ = h.f.Clear(context.Background(), ClearOpts{Timeout: 20 * time.Second})
	if h.s.melee != 0 {
		t.Fatalf("%d swings toward the mouth", h.s.melee)
	}
	if mv.asked == 0 {
		t.Fatal("the gate was never asked")
	}
	if h.s.Now().Sub(t0) <= 0 {
		t.Fatal("no time passed")
	}
}

// A target already in melee range (2) is swung at: the character does not walk.
func TestSwingAllowedInMeleeRange(t *testing.T) {
	h := newHarness(t)
	h.s.mana, h.s.manaRegen, h.s.meleeDmg = 0, 0, 20
	h.s.addMon(10, clsFallen, 102, 100, 30, 1, 0)
	h.f.mv = &gateAt{Mover: h.s, ex: input.Point{X: 108, Y: 100}}
	res, err := h.f.Clear(context.Background(), ClearOpts{Timeout: time.Minute})
	if err != nil || res.Kills != 1 || h.s.melee == 0 {
		t.Fatalf("adjacent target must be swung at: %+v %v melee %d", res, err, h.s.melee)
	}
}

// Far from any exit the gate allows the swing: the fight is unchanged.
func TestSwingAllowedAwayFromMouth(t *testing.T) {
	h := newHarness(t)
	h.s.mana, h.s.manaRegen, h.s.meleeDmg = 0, 0, 20
	h.s.addMon(10, clsFallen, 104, 100, 30, 1, 0)
	h.f.mv = &gateAt{Mover: h.s, ex: input.Point{X: 160, Y: 100}}
	res, err := h.f.Clear(context.Background(), ClearOpts{Timeout: time.Minute})
	if err != nil || res.Kills != 1 {
		t.Fatalf("%+v %v", res, err)
	}
}
