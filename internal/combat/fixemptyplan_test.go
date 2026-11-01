package combat

import (
	"context"
	"fmt"
	"testing"

	"d2lod-bot/internal/input"
)

// snapMover models a navigator whose goal cells in `blocked` snap onto the player's own cell: PathTo and GoTo
// return the typed already-there result for them (nav.ErrAlreadyThere), and walk normally elsewhere.
type snapMover struct {
	Mover
	blocked func(input.Point) bool
	gotos   []input.Point
	pathTos int
}

var errThere = fmt.Errorf("nav: no path: %w: empty path", input.ErrAlreadyThere)

func (m *snapMover) PathTo(p input.Point) ([]input.Point, error) {
	m.pathTos++
	if m.blocked(p) {
		return nil, errThere
	}
	return m.Mover.PathTo(p)
}

func (m *snapMover) GoTo(ctx context.Context, p input.Point) error {
	m.gotos = append(m.gotos, p)
	if m.blocked(p) {
		return errThere
	}
	return m.Mover.GoTo(ctx, p)
}

// Approach vs back-off for a goal that snaps onto the start cell (docs/fixemptyplan.md).
func TestValidApproachVsBackoff(t *testing.T) {
	cases := []struct {
		name         string
		blocked      bool
		backoff, app bool
	}{
		{"goal snaps onto the player", true, false, true},
		{"goal reachable", false, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			h.f.mv = &snapMover{Mover: h.s, blocked: func(input.Point) bool { return c.blocked }}
			p := input.Point{X: 90, Y: 90}
			if got := h.f.valid(p); got != c.backoff {
				t.Fatalf("valid (back-off) = %v, want %v", got, c.backoff)
			}
			if got := h.f.validApproach(p); got != c.app {
				t.Fatalf("validApproach = %v, want %v", got, c.app)
			}
		})
	}
}

// The back-off picks another retreat cell when the best one snaps onto the player, with a bounded number of probes.
func TestAwayPointSkipsSnappedCandidates(t *testing.T) {
	h := newHarness(t)
	me := input.Point{X: 100, Y: 100}
	threat := []input.Point{{X: 95, Y: 100}}
	first, ok := awayPoint(me, threat, nil, 10, func(input.Point) bool { return true })
	if !ok {
		t.Fatal("no retreat in the open")
	}
	m := &snapMover{Mover: h.s, blocked: func(p input.Point) bool { return p == first }}
	h.f.mv = m
	got, ok := awayPoint(me, threat, nil, 10, h.f.valid)
	if !ok || got == first || got == me {
		t.Fatalf("got %v ok=%v; want a different cell than %v", got, ok, first)
	}
	if m.pathTos > 3*len(retreatAngles) {
		t.Fatalf("%d probes: not bounded", m.pathTos)
	}
	// every cell snaps onto the player: no retreat at all, so the caller falls through to its next option
	m2 := &snapMover{Mover: h.s, blocked: func(input.Point) bool { return true }}
	h.f.mv = m2
	if _, ok := awayPoint(me, threat, nil, 10, h.f.valid); ok {
		t.Fatal("a retreat was offered although no cell can be reached")
	}
	if m2.pathTos > 3*len(retreatAngles) {
		t.Fatalf("%d probes: not bounded", m2.pathTos)
	}
}

// move: an arrival is not a failed walk (no error, a tick passes); a real failure still logs and passes a tick.
func TestMoveAlreadyThereIsQuietTick(t *testing.T) {
	h := newHarness(t)
	h.f.mv = &snapMover{Mover: h.s, blocked: func(input.Point) bool { return true }}
	t0 := h.s.Now()
	if err := h.f.move(context.Background(), input.Point{X: 90, Y: 90}); err != nil {
		t.Fatal(err)
	}
	if !h.s.Now().After(t0) {
		t.Fatal("no tick passed")
	}
}
