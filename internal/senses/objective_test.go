package senses

import (
	"testing"

	"d2lod-bot/internal/d2map"
)

// The objective marker is not fogged, and only the current area's is published.
func TestObjectiveShownBeyondTheFogForTheCurrentAreaOnly(t *testing.T) {
	for name, lm := range fogLevels() { // an open field and a cave: the build differs, the rule does not
		t.Run(name, func(t *testing.T) {
			var p Publisher
			if _, ok := p.ObjectiveOf(); ok {
				t.Fatal("an objective before the first observation")
			}
			if err := p.Observe(lm, name, d2map.Point{X: 8, Y: 30}); err != nil {
				t.Fatal(err)
			}
			far := Objective{Area: lm.No, X: 230, Y: 30, Label: "Countess"}
			p.SetObjective(far, true)
			if p.revealedLocked(lm.No, far.X, far.Y) {
				t.Fatal("test geometry: the goal must lie in the fog")
			}
			if o, ok := p.ObjectiveOf(); !ok || o != far {
				t.Fatalf("objective %+v %v: the marker must show although its block is unexplored", o, ok)
			}
			p.SetObjective(Objective{Area: lm.No + 1, X: 1, Y: 1}, true)
			if _, ok := p.ObjectiveOf(); ok {
				t.Fatal("the objective of another area was published")
			}
			p.SetObjective(far, false)
			if _, ok := p.ObjectiveOf(); ok {
				t.Fatal("a cleared objective was published")
			}
		})
	}
}

// A new game lays the levels out again: the blocks the hero walked in the last game must not reveal the new one.
func TestNewGameForgetsTheWalkedGround(t *testing.T) {
	lm := fogLevels()["field"]
	var p Publisher
	for x := 8; x < 200; x += 20 {
		if err := p.Observe(lm, "field", d2map.Point{X: x, Y: 30}); err != nil {
			t.Fatal(err)
		}
	}
	if !p.revealedLocked(lm.No, 150, 30) {
		t.Fatal("test geometry: the walk must have revealed x=150")
	}
	p.SetObjective(Objective{Area: lm.No, X: 5, Y: 5}, true)
	p.NewGame()
	if p.revealedLocked(lm.No, 150, 30) {
		t.Fatal("the walked ground of the last game still reveals the new one")
	}
	if _, ok := p.ObjectiveOf(); ok {
		t.Fatal("the objective of the last game survived")
	}
	if err := p.Observe(lm, "field", d2map.Point{X: 8, Y: 30}); err != nil {
		t.Fatal(err)
	}
	if p.revealedLocked(lm.No, 150, 30) || !p.revealedLocked(lm.No, 8, 30) {
		t.Fatal("after the new game only the ground around the hero is revealed")
	}
}
