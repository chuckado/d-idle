package combat

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
)

// verbGame is the popup game plus a select verb. effect false is the devp13 behaviour of R08: the verb is accepted
// and leaves the hand unchanged.
type verbGame struct {
	*popupGame
	effect bool
	verbs  int
	clicks int
	left   int
}

func (g *verbGame) SelectSkillID(id int, left bool) error {
	g.verbs++
	if g.effect {
		if left {
			g.left = id
		} else {
			g.right = id
		}
	}
	return nil
}

func (g *verbGame) Click(b input.Button, x, y int, m ...input.Mod) error {
	g.clicks++
	return g.popupGame.Click(b, x, y, m...)
}

func plannerHarness(t *testing.T, on bool, caps ...string) (*harness, *verbGame) {
	h, pg := newPopupHarness(t)
	g := &verbGame{popupGame: pg}
	h.f.e.In = g
	h.f.D2D = true
	h.f.e.Planner = &input.Planner{Enabled: func() bool { return on }, Caps: &input.BackendCaps{Verbs: caps}}
	return h, g
}

func TestPlannerVerbSelectsWithoutClicks(t *testing.T) {
	for _, id := range []int{38, 2} { // two skills of two different rows: the rule is data, not one build
		h, g := plannerHarness(t, true, input.CapSelect)
		g.effect = true
		if err := h.f.SetRight(context.Background(), id); err != nil {
			t.Fatal(err)
		}
		v, k, c, _ := h.f.e.Planner.Counts()
		if g.right != id || g.clicks != 0 || g.opened != 0 || v != 1 || k != 0 || c != 0 {
			t.Fatalf("skill %d: right %d clicks %d opened %d counts %d/%d/%d", id, g.right, g.clicks, g.opened, v, k, c)
		}
	}
}

func TestPlannerKnobOffKeepsLegacyPopup(t *testing.T) {
	h, g := plannerHarness(t, false, input.CapSelect)
	g.effect = true
	if err := h.f.SetRight(context.Background(), 38); err != nil {
		t.Fatal(err)
	}
	if g.verbs != 0 || g.opened == 0 || g.right != 38 {
		t.Fatalf("knob off must click the popup: verbs %d opened %d right %d", g.verbs, g.opened, g.right)
	}
	if v, k, c, _ := h.f.e.Planner.Counts(); v+k+c != 0 {
		t.Fatalf("legacy path must not count: %d/%d/%d", v, k, c)
	}
}

func TestPlannerNoCapFallsBackToVerifiedClick(t *testing.T) {
	h, g := plannerHarness(t, true) // devp13: select not advertised
	if err := h.f.SetRight(context.Background(), 38); err != nil {
		t.Fatal(err)
	}
	if g.verbs != 0 || g.right != 38 || g.opened == 0 {
		t.Fatalf("verbs %d right %d opened %d", g.verbs, g.right, g.opened)
	}
	if v, k, c, _ := h.f.e.Planner.Counts(); v != 0 || k != 0 || c != 1 {
		t.Fatalf("counts %d/%d/%d", v, k, c)
	}
}

func TestPlannerBrokenVerbFallsBackAndIsNotCounted(t *testing.T) {
	h, g := plannerHarness(t, true, input.CapSelect) // advertised but without effect (R08)
	if err := h.f.SetRight(context.Background(), 38); err != nil {
		t.Fatal(err)
	}
	v, k, c, _ := h.f.e.Planner.Counts()
	if g.verbs != 1 || g.right != 38 || v != 0 || c != 1 || k != 0 {
		t.Fatalf("verbs %d right %d counts %d/%d/%d", g.verbs, g.right, v, k, c)
	}
}

func TestPlannerLeftVerbAndNoPathError(t *testing.T) {
	h, g := plannerHarness(t, true, input.CapSelect)
	g.effect = true
	h.f.snapFn = func() (*d2.Snapshot, error) { s, _ := g.snap(); s.Player.LeftSkill = g.left; return s, nil }
	if err := h.f.SetLeft(context.Background(), 36); err != nil {
		t.Fatal(err)
	}
	if g.left != 36 || g.clicks != 0 {
		t.Fatalf("left %d clicks %d", g.left, g.clicks)
	}
	g.effect = false
	h.f.SetListRows(nil)
	if err := h.f.SetLeft(context.Background(), 38); !errors.Is(err, ErrBindFailed) {
		t.Fatalf("no verb effect and no popup rows must be ErrBindFailed: %v", err)
	}
}
