package nav

import (
	"errors"
	"testing"

	"d2lod-bot/internal/input"
)

// The fork's waypoint verb (docs/process/input-verbs2.md): one message instead of the act tab and row click, only with
// the knob input.planner on AND the backend's input.waypoint capability; a named rule refusal ends the attempt, any
// other failure leaves the clicks as they were.

// wpVerbIn wraps the sim controller and offers the verbs; an accepted TakeWaypoint does what the row click does.
type wpVerbIn struct {
	input.Controller
	w     *wpWorld
	err   error
	asked int
}

func (v *wpVerbIn) WorldVerbs() (input.WorldVerbs, bool) { return v, true }
func (v *wpVerbIn) TakeWaypoint(level int) error {
	v.asked++
	if v.err != nil {
		return v.err
	}
	v.w.s.px, v.w.s.py = 1300.5, 1050.5
	v.w.s.ui[0x14] = 0
	return nil
}
func (v *wpVerbIn) BuyAt(int, int, int) error { return errors.New("unused") }
func (v *wpVerbIn) SellItem(uint32) error     { return errors.New("unused") }
func (v *wpVerbIn) EnterPortal(int) error     { return errors.New("unused") }
func (v *wpVerbIn) PickupItem(uint32) error   { return errors.New("unused") }

func wpVerbNav(t *testing.T, on bool, caps ...string) (*wpWorld, *Navigator, *wpVerbIn) {
	w, n := newWPWorld(t, 1)
	v := &wpVerbIn{Controller: n.in, w: w}
	n.in = v
	n.e.Planner = &input.Planner{Enabled: func() bool { return on }, Caps: &input.BackendCaps{Verbs: caps}}
	return w, n, v
}

func TestWaypointVerbReplacesTheClicks(t *testing.T) {
	w, n, v := wpVerbNav(t, true, input.CapWaypoint)
	if err := n.TakeWaypoint(ctxT(t), 3); err != nil {
		t.Fatal(err)
	}
	if v.asked != 1 || w.rowClicks != 0 {
		t.Fatalf("verb asked %d times, %d row clicks: want the verb and no row click", v.asked, w.rowClicks)
	}
	if got := n.e.Planner.Summary(); got != "1/0/0" {
		t.Fatalf("path counts %s, want 1/0/0", got)
	}
}

func TestWaypointVerbNeedsKnobAndCapability(t *testing.T) {
	for name, mk := range map[string]func() (*wpWorld, *Navigator, *wpVerbIn){
		"knob off": func() (*wpWorld, *Navigator, *wpVerbIn) { return wpVerbNav(t, false, input.CapWaypoint) },
		"no cap":   func() (*wpWorld, *Navigator, *wpVerbIn) { return wpVerbNav(t, true) },
	} {
		w, n, v := mk()
		if err := n.TakeWaypoint(ctxT(t), 3); err != nil {
			t.Fatal(err)
		}
		if v.asked != 0 || w.rowClicks != 1 {
			t.Errorf("%s: verb asked %d, row clicks %d: the click path must run unchanged", name, v.asked, w.rowClicks)
		}
	}
}

func TestWaypointVerbRefusalEndsTheAttempt(t *testing.T) {
	w, n, v := wpVerbNav(t, true, input.CapWaypoint)
	v.err = input.Refusal("inactive")
	if err := n.TakeWaypoint(ctxT(t), 3); !errors.Is(err, ErrNoWaypoint) {
		t.Fatalf("err = %v, want ErrNoWaypoint naming the refusal", err)
	}
	if w.rowClicks != 0 {
		t.Fatalf("%d row clicks after a named refusal: a click could not fix an inactive waypoint", w.rowClicks)
	}
}

func TestWaypointVerbTransportErrorFallsBackToClicks(t *testing.T) {
	for _, err := range []error{errors.New("unknown verb waypoint"), input.Refusal("closed")} {
		w, n, v := wpVerbNav(t, true, input.CapWaypoint)
		v.err = err
		if got := n.TakeWaypoint(ctxT(t), 3); got != nil {
			t.Fatalf("%v: %v", err, got)
		}
		if w.rowClicks != 1 {
			t.Fatalf("%v: %d row clicks, want the fallback", err, w.rowClicks)
		}
	}
}
