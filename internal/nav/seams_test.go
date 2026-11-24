package nav

import (
	"testing"

	"d2lod-bot/internal/d2map"
	"d2lod-bot/internal/input"
)

// Two floors that meet at one corner (the Arcane Sanctuary stand-grid, V lab note 49): the plan crosses the seam when no
// other way exists, with the whole grid known. Must fail without the seam retry: the plan is partial.
func TestPlanCrossesADiagonalSeamOnAWholeGrid(t *testing.T) {
	for _, cls := range []string{"Paladin", "Sorceress"} {
		t.Run(cls, func(t *testing.T) {
			g := openField(20, 20, 200, 120)
			blockRect(g, d2map.Rect{X: 100, Y: 20, W: 120, H: 50}) // north-east solid
			blockRect(g, d2map.Rect{X: 20, Y: 70, W: 80, H: 70})   // south-west solid: the floors meet at (99,69) and (100,70)
			_, n := newKnown(t, g, input.Point{X: 40, Y: 40}, true)
			tuningOf(n).Nav.FullGrid = true
			o, err := n.observe()
			if err != nil {
				t.Fatal(err)
			}
			goal := input.Point{X: 150, Y: 100}
			res, err := n.plan(o.lm, input.Point{X: 40, Y: 40}, goal, 2, nil, 0)
			if err != nil || res.partial {
				t.Fatalf("plan: partial %v, err %v", res.partial, err)
			}
		})
	}
}
