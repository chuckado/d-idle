package nav

import (
	"context"
	"errors"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// wpWorld: Rogue Encampment (1) with a waypoint object and Cold Plains (3). The panel opens after `needClicks`
// interactions (a lit waypoint: 1, a new one: 2, D2D-V) and a row click only travels once `ready` says so.
type wpWorld struct {
	s          *sim
	interacts  int
	needClicks int
	rowClicks  int
	arrived    time.Time // virtual time the character entered the level
	cooldown   time.Duration
}

func newWPWorld(t *testing.T, needClicks int) (*wpWorld, *Navigator) {
	g := openField(1000, 1000, 400, 100)
	l1 := &simLevel{no: 1, rect: d2map.Rect{X: 1000, Y: 1000, W: 200, H: 100}}
	l3 := &simLevel{no: 3, rect: d2map.Rect{X: 1200, Y: 1000, W: 200, H: 100}}
	s := newSim(g, input.Point{X: 1020, Y: 1050}, l1, l3)
	s.objects = []d2.Unit{d2Unit(11, 119, 1060, 1050)}
	w := &wpWorld{s: s, needClicks: needClicks, cooldown: 0, arrived: s.now}
	s.onInteract = func(s *sim, _ input.Target) error {
		w.interacts++
		if w.interacts >= w.needClicks {
			s.ui[0x14] = 1
		}
		return nil
	}
	s.onClick = func(s *sim, x, y int) {
		if y != wpPanel.RowY[1] {
			return
		}
		w.rowClicks++
		if s.now.Sub(w.arrived) >= w.cooldown { // the game eats a click inside the cooldown
			s.px, s.py = 1300.5, 1050.5
			s.ui[0x14] = 0
		}
	}
	return w, s.navigator()
}

// Default (0): no wait; the game here has no cooldown.
func TestWaypointNoCooldownWait(t *testing.T) {
	w, n := newWPWorld(t, 1)
	start := w.s.now
	if err := n.TakeWaypoint(ctxT(t), 3); err != nil {
		t.Fatal(err)
	}
	if got := w.s.now.Sub(start); got >= 9*time.Second {
		t.Fatalf("took %v on the fake clock, no cooldown wait expected", got)
	}
	if w.rowClicks != 1 {
		t.Fatalf("row clicks = %d, want 1", w.rowClicks)
	}
}

// The knob still works for an old binary: a positive value waits it out.
func TestWaypointKnobWaitsOutCooldown(t *testing.T) {
	w, n := newWPWorld(t, 1)
	n.e.Tune.Nav.WpCooldownSec = 10
	start := w.s.now
	if err := n.TakeWaypoint(ctxT(t), 3); err != nil {
		t.Fatal(err)
	}
	if got := w.s.now.Sub(start); got < 10*time.Second {
		t.Fatalf("clicked after %v, the knob says 10 s", got)
	}
	if w.rowClicks != 1 {
		t.Fatalf("row clicks = %d, want 1", w.rowClicks)
	}
}

func TestWaypointSecondClickActivates(t *testing.T) {
	w, n := newWPWorld(t, 2)
	if err := n.TakeWaypoint(ctxT(t), 3); err != nil {
		t.Fatal(err)
	}
	if w.interacts != 2 || n.wpClicks < 3 {
		t.Fatalf("interacts %d, clicks %d: want 2 interactions then the row click", w.interacts, n.wpClicks)
	}
}

// A row click the game ate (our clock started late) is retried while the arrival is recent.
func TestWaypointRetriesEatenClick(t *testing.T) {
	w, n := newWPWorld(t, 1)
	w.cooldown = 9 * time.Second // a refusal for another reason: the retry round clicks again, bounded
	if err := n.TakeWaypoint(ctxT(t), 3); err != nil {
		t.Fatal(err)
	}
	if w.rowClicks < 2 {
		t.Fatalf("row clicks = %d, want a retry round", w.rowClicks)
	}
}

func TestWaypointNoRetryForOldArrival(t *testing.T) {
	w, n := newWPWorld(t, 1)
	w.cooldown = time.Hour // an inactive waypoint: never travels
	n.wpLevel, n.wpSince = 1, w.s.now.Add(-time.Hour)
	w.s.now = w.s.now.Add(time.Minute)
	if err := n.TakeWaypoint(ctxT(t), 3); !errors.Is(err, ErrNoWaypoint) {
		t.Fatalf("err = %v", err)
	}
	if w.rowClicks > 2 {
		t.Fatalf("row clicks = %d: a long-standing character has no cooldown excuse, one round only", w.rowClicks)
	}
}

func TestWaypointCursorItem(t *testing.T) {
	w, n := newWPWorld(t, 1)
	w.s.cursor = &d2.Unit{ID: 5}
	if err := n.TakeWaypoint(ctxT(t), 3); !errors.Is(err, ErrCursorItem) {
		t.Fatalf("no hook: err = %v, want ErrCursorItem", err)
	}
	if w.interacts != 0 {
		t.Fatal("the waypoint was clicked with an item on the cursor")
	}
	cleared := 0
	n.ClearCursor = func(context.Context) error { cleared++; w.s.cursor = nil; return nil }
	if err := n.TakeWaypoint(ctxT(t), 3); err != nil || cleared != 1 {
		t.Fatalf("err %v cleared %d", err, cleared)
	}
}

// W2: the nine row centres sit inside game.exe's hit boxes (tops 60 96 132 168 205 241 277 313 349, +60 on screen,
// 30 px tall, exclusive) and every row of an act maps to its own box; the target act's tab is clicked.
func TestWaypointRowCentresHitBoxes(t *testing.T) {
	tops := [9]int{60, 96, 132, 168, 205, 241, 277, 313, 349}
	for r, top := range tops {
		y := wpPanel.RowY[r]
		if y <= top+60 || y >= top+90 {
			t.Errorf("row %d: y=%d outside (%d,%d)", r, y, top+60, top+90)
		}
		if r > 0 && y-wpPanel.RowY[r-1] > 37 {
			t.Errorf("row %d pitch %d", r, y-wpPanel.RowY[r-1])
		}
	}
	if wpPanel.RowX <= 97 || wpPanel.RowX >= 377 {
		t.Errorf("RowX %d outside 97..377", wpPanel.RowX)
	}
	for a, x := range wpPanel.TabX {
		if lo := 80 + 64*a; x <= lo || x >= lo+64 || wpPanel.TabY > 90 {
			t.Errorf("tab %d at %d,%d outside its box", a, x, wpPanel.TabY)
		}
	}
}

// Every waypoint level is clicked at its own act tab and row centre (rows 7 and 8 included).
func TestWaypointClicksEveryLevel(t *testing.T) {
	var all []int
	for _, act := range wpLevels {
		all = append(all, act...)
	}
	for _, lv := range all {
		act, row, ok := waypointSlot(lv)
		if !ok {
			t.Fatalf("level %d has no slot", lv)
		}
		cur := 1
		if act > 0 {
			cur = []int{1, 40, 75, 103, 109}[act]
		}
		g := openField(1000, 1000, 400, 100)
		l0 := &simLevel{no: cur, rect: d2map.Rect{X: 1000, Y: 1000, W: 200, H: 100}}
		l1 := &simLevel{no: lv, rect: d2map.Rect{X: 1200, Y: 1000, W: 200, H: 100}}
		if cur == lv {
			continue
		}
		s := newSim(g, input.Point{X: 1020, Y: 1050}, l0, l1)
		s.objects, s.units = nil, nil
		s.objects = append(s.objects, d2Unit(11, 119, 1060, 1050))
		s.onInteract = func(s *sim, _ input.Target) error { s.ui[0x14] = 1; return nil }
		var xs, ys []int
		s.onClick = func(s *sim, x, y int) {
			xs, ys = append(xs, x), append(ys, y)
			if y == wpPanel.RowY[row] {
				s.px, s.py = 1300.5, 1050.5
				s.ui[0x14] = 0
			}
		}
		if err := s.navigator().TakeWaypoint(ctxT(t), lv); err != nil {
			t.Fatalf("level %d: %v", lv, err)
		}
		if xs[0] != wpPanel.TabX[act] || ys[0] != wpPanel.TabY || xs[len(xs)-1] != wpPanel.RowX || ys[len(ys)-1] != wpPanel.RowY[row] {
			t.Errorf("level %d act %d row %d: clicks %v / %v", lv, act, row, xs, ys)
		}
	}
}

// The log signature: a waypoint with class -1 (what d2dsrc filed before W1) is never found; a waypoint class
// is, and the nearest wins.
func TestNearestWaypointUnit(t *testing.T) {
	o := &obs{pos: input.Point{X: 10, Y: 10}, snap: &d2.Snapshot{Objects: []d2.Unit{
		d2Unit(1, -1, 11, 10), d2Unit(2, 156, 50, 50), d2Unit(3, 119, 20, 10)}}}
	if u := (&Navigator{}).nearestWaypointUnit(o); u == nil || u.ID != 3 {
		t.Fatalf("nearest = %+v, want unit 3", u)
	}
	o.snap.Objects = o.snap.Objects[:1]
	if u := (&Navigator{}).nearestWaypointUnit(o); u != nil {
		t.Fatalf("class -1 found: %+v", u)
	}
}
