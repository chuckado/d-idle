package bot

import (
	"testing"
)

type drainMover struct {
	*fakeMover
	pending []int
}

func (m *drainMover) DrainActivated() []int { a := m.pending; m.pending = nil; return a }

// FIXNAV: a waypoint the navigator touched on the town's waypoint-to-town trip counts as active for the picker.
func TestNavigatorTouchedWaypointIsActive(t *testing.T) {
	q := &fakeQuests{}
	q.v.WPs = map[int]bool{}
	g := campaignRig(t, q, 12)
	g.r.d.Mover = &drainMover{fakeMover: g.mover, pending: []int{6}}
	g.w.set(func(w *world) { w.area = 4 })
	g.r.poll()
	active, _ := g.r.activeWaypoints(g.r.latest())
	if !active[6] {
		t.Fatalf("active %v: want area 6", active)
	}
}
