package drill

import (
	"context"
	"strings"
	"testing"
)

// simWP: a character that travels to any activated waypoint and arrives next to it, or does nothing.
type simWP struct {
	w      *fakeWorld
	area   int
	active bool
	broken bool // the verb does not travel
	tick   int
}

func (s *simWP) WView(context.Context) (WPView, error) {
	return WPView{Area: s.area, X: 100, Y: 100, Tick: s.tick, Waypoints: [][2]int{{103, 103}}}, nil
}
func (s *simWP) Tick(_ context.Context, n int) error { s.tick += n; return nil }
func (s *simWP) Goto(_ context.Context, r *Run, level int) error {
	_, err := r.Cheat(context.Background(), "debug goto", "start")
	s.area = level
	return err
}
func (s *simWP) ActivateAll(_ context.Context, r *Run) error {
	_, err := r.Cheat(context.Background(), "debug wp all", "activate")
	s.active = true
	return err
}
func (s *simWP) UnlockAct(_ context.Context, r *Run, act int) error {
	_, err := r.Cheat(context.Background(), "debug act", "unlock")
	return err
}
func (s *simWP) Tabs() string { return "[2 rects]" }
func (s *simWP) TakeWaypoint(_ context.Context, level int) WPCall {
	if !s.active || s.broken {
		return WPCall{Err: "waypoint unavailable: unit not found", Ticks: 10}
	}
	s.area = level
	return WPCall{Ticks: 300, Clicks: 3}
}

func TestDW(t *testing.T) {
	run := func(s *simWP) map[string]Verdict {
		s.w = &fakeWorld{}
		out := map[string]Verdict{}
		for _, d := range DW(s) {
			if d.Do == nil {
				out[d.ID] = Blocked
				continue
			}
			r := &Run{W: s.w, Drill: d.ID}
			out[d.ID] = d.Do(context.Background(), r).Verdict
		}
		return out
	}
	got := run(&simWP{area: 1})
	for id, v := range got {
		if v != Pass {
			t.Errorf("%s = %s", id, v)
		}
	}
	// the log signature: the verb finds no unit, nothing travels
	bad := run(&simWP{area: 1, broken: true})
	if bad["DW.a1"] != Fail || bad["DW.c"] != Fail {
		t.Errorf("broken verb: %v", bad)
	}
	if bad["DW.b"] != Pass {
		t.Errorf("inactive drill with a failing verb should pass: %v", bad)
	}
	if !strings.Contains(DW(&simWP{})[1].Title, "Cold Plains") {
		t.Error("drill order")
	}
}
