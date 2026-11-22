package d2dworld

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2dctl"
)

func (s *stub) WarpTriggers(context.Context) (*d2dctl.WarpTriggers, error) {
	s.n("warptriggers")
	if s.wt == nil {
		return nil, errors.New("err unknown debug verb")
	}
	return s.wt, nil
}

// One warp (to level 2, as the stub's "debug warps" lists it) with a two-row click area and a reach cell.
func stubTriggers() *d2dctl.WarpTriggers {
	return &d2dctl.WarpTriggers{Level: 1, Unit: "subtile", Scan: 40, ReachRule: 5, Warps: []d2dctl.WarpTrigger{{
		Index: 0, To: 2, Unit: [2]int{14, 3},
		Click: []d2dctl.Run{{2, 10, 12}, {3, 10, 12}}, Reach: []d2dctl.Run{{3, 8, 9}}}}}
}

func TestWarpReadGivesExitsTheirTriggerAreas(t *testing.T) {
	s := newStub()
	s.wt = stubTriggers()
	w := New(s, Options{WarpRead: func() bool { return true }})
	if _, err := w.Update(); err != nil {
		t.Fatal(err)
	}
	ex := w.Current().Exits
	if len(ex) != 1 || ex[0].Trigger == nil {
		t.Fatalf("exits %+v (%s)", ex, w.LastTriggerNote)
	}
	if !ex[0].Trigger.ClickNear(11, 3, 0) || ex[0].Trigger.ClickNear(13, 3, 0) || len(ex[0].Trigger.Reach) != 1 {
		t.Fatalf("trigger %+v", ex[0].Trigger)
	}
}

func TestWarpReadOffAbsentOrMismatchedChangesNothing(t *testing.T) {
	bad := stubTriggers()
	bad.Warps[0].To = 9 // another level than the exit list says
	cases := map[string]func() *World{
		"knob off": func() *World {
			s := newStub()
			s.wt = stubTriggers()
			return New(s, Options{WarpRead: func() bool { return false }})
		},
		"knob unset": func() *World { s := newStub(); s.wt = stubTriggers(); return New(s, Options{}) },
		"verb absent": func() *World {
			return New(struct{ Verbs }{newStub()}, Options{WarpRead: func() bool { return true }})
		},
		"verb errors": func() *World { return New(newStub(), Options{WarpRead: func() bool { return true }}) },
		"mismatch": func() *World {
			s := newStub()
			s.wt = bad
			return New(s, Options{WarpRead: func() bool { return true }})
		},
	}
	for name, mk := range cases {
		w := mk()
		if _, err := w.Update(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if ex := w.Current().Exits; len(ex) != 1 || ex[0].Trigger != nil {
			t.Errorf("%s: exits %+v", name, ex)
		}
	}
}
