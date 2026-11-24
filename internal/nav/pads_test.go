package nav

import (
	"math"
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// The Arcane Sanctuary is platforms over a void joined by teleport pads (V lab, pads.go). Three platforms in a row: pad P1
// on A leads to B, B has the pad it landed on (P1b, back to A) and P2 (to C); the goal is on C. Must fail without hopPad:
// GoTo ends in ErrNoPath on A. Two builds walk it (the hop is class-free).
func TestGoToHopsTeleportPadsToAnIslandGoal(t *testing.T) {
	for _, cls := range []string{"Paladin", "Sorceress"} {
		t.Run(cls, func(t *testing.T) {
			g := openField(1000, 1000, 600, 60)
			blockRect(g, d2map.Rect{X: 1100, Y: 1000, W: 100, H: 60}) // the void between A and B
			blockRect(g, d2map.Rect{X: 1300, Y: 1000, W: 100, H: 60}) // and between B and C
			l1 := &simLevel{no: 74, rect: d2map.Rect{X: 1000, Y: 1000, W: 600, H: 60}}
			s := newSim(g, input.Point{X: 1010, Y: 1030}, l1)
			s.reveal = 400
			p1, p1b, p2 := input.Point{X: 1060, Y: 1052}, input.Point{X: 1220, Y: 1052}, input.Point{X: 1280, Y: 1052}
			for i, p := range []input.Point{p1, p1b, p2} {
				s.objects = append(s.objects, d2Unit(uint32(40+i), teleportPadClass, p.X, p.Y))
			}
			dest := map[input.Point]input.Point{p1: {X: 1220, Y: 1049}, p1b: {X: 1060, Y: 1049}, p2: {X: 1450, Y: 1030}}
			armed := map[input.Point]bool{p1: true, p1b: true, p2: true}
			s.onStep = func(s *sim) {
				for p, to := range dest {
					d := math.Hypot(s.px-float64(p.X), s.py-float64(p.Y))
					if d > 6 {
						armed[p] = true
					} else if d < 2 && armed[p] {
						armed[p] = false
						armed[to] = false
						s.px, s.py = float64(to.X)+0.5, float64(to.Y)+0.5
						for q := range dest { // landing beside the pair stands on it, which fires nothing
							if math.Hypot(float64(q.X-to.X), float64(q.Y-to.Y)) < 6 {
								armed[q] = false
							}
						}
						return
					}
				}
			}
			n := s.navigator()
			goal := input.Point{X: 1500, Y: 1030}
			if err := n.GoTo(ctxT(t), goal); err != nil {
				t.Fatalf("goto: %v (at %v)", err, s.pos())
			}
			if dist(s.pos(), goal) > 6 {
				t.Fatalf("ended at %v", s.pos())
			}
		})
	}
}

// Without a pad the same void stays a no-path (the old behaviour).
func TestGoToWithoutPadsStillFails(t *testing.T) {
	g := openField(1000, 1000, 600, 60)
	blockRect(g, d2map.Rect{X: 1100, Y: 1000, W: 100, H: 60})
	l1 := &simLevel{no: 74, rect: d2map.Rect{X: 1000, Y: 1000, W: 600, H: 60}}
	s := newSim(g, input.Point{X: 1010, Y: 1030}, l1)
	s.reveal = 400
	if err := s.navigator().GoTo(ctxT(t), input.Point{X: 1500, Y: 1030}); err == nil {
		t.Fatal("reached an island without a pad")
	}
}
