package combat

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
)

// hitWorld turns the player into hit recovery during the first n casts, as a monster hit that cuts the cast.
type hitWorld struct {
	*nworld
	cut int
}

func (h *hitWorld) CastAt(key string, at input.Point) error {
	err := h.nworld.CastAt(key, at)
	if h.cut > 0 {
		h.cut--
		h.snap.Player.Mode = modeGetHit
	}
	return err
}

func armorCutSetup(t *testing.T, mana int) (*harness, *hitWorld) {
	t.Helper()
	h, w := newNecro(t)
	cat := necroCat()
	cat.Skills[40] = frozenArmorRow(true)
	h.f.SetCatalog(cat)
	p := w.snap.Player
	p.Class, p.RightSkill, p.Mana, p.MaxMana = d2.Sorceress, 40, mana, 400
	p.Skills = []d2.Skill{{ID: 40, Level: 1}}
	h.f.SetHotkeys(map[int]string{40: "f2"})
	hw := &hitWorld{nworld: w}
	h.f.e.In = hw
	return h, hw
}

func buffNow(t *testing.T, h *harness) {
	t.Helper()
	snap, _ := h.f.snapFn()
	if _, err := h.f.buffPass(context.Background(), snap); err != nil {
		t.Fatal(err)
	}
}

// The refresh cast at the limit is cut by a hit: the old cast time stays, so the next pass retries at once.
func TestBuffCutByHitRetriesAtOnce(t *testing.T) {
	h, w := armorCutSetup(t, 400)
	buffNow(t, h) // opening cast
	w.snap.Player.States[d2.StateFrozenArmor/32] |= 1 << (d2.StateFrozenArmor % 32)
	h.s.Sleep(115 * time.Second) // past the 114 s limit, armor still up
	w.cut = 1
	buffNow(t, h)
	if len(w.casts) != 2 {
		t.Fatalf("casts %d, want 2", len(w.casts))
	}
	w.snap.Player.Mode = 1 // hit recovery over
	buffNow(t, h)
	if len(w.casts) != 3 {
		t.Fatalf("casts %d, want a retry after the cut cast", len(w.casts))
	}
	buffNow(t, h)
	if len(w.casts) != 3 {
		t.Fatalf("casts %d: a completed cast must not repeat", len(w.casts))
	}
}

// Nothing is cast during hit recovery or while casting; retries are bounded and need the mana.
func TestBuffNoCastMidHitAndBoundedRetries(t *testing.T) {
	h, w := armorCutSetup(t, 400)
	w.snap.Player.Mode = modeGetHit
	buffNow(t, h)
	w.snap.Player.Mode = modeCast
	buffNow(t, h)
	if len(w.casts) != 0 {
		t.Fatalf("cast during hit recovery or a cast: %d", len(w.casts))
	}
	w.snap.Player.Mode = 1
	w.cut = 100 // every cast is cut
	for i := 0; i < 10; i++ {
		w.snap.Player.Mode = 1
		buffNow(t, h)
	}
	if len(w.casts) != buffTriesMax+1 {
		t.Fatalf("casts %d, want %d (bounded retries)", len(w.casts), buffTriesMax+1)
	}
}

func TestBuffCutRetryNeedsMana(t *testing.T) {
	h, w := armorCutSetup(t, 400)
	w.cost = 7
	w.cut = 1
	buffNow(t, h)
	w.snap.Player.Mode = 1
	w.snap.Player.Mana = 0
	buffNow(t, h)
	if len(w.casts) != 1 {
		t.Fatalf("casts %d: no retry without mana", len(w.casts))
	}
}
