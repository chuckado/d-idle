package d2dworld

import (
	"context"
	"testing"
)

// openAll answers every probe past the edge as open, so only the geometry decides.
type openAll struct{ *borderStub }

func (openAll) Blocked(context.Context, float64, float64) (bool, error) { return false, nil }

// Card mapverify: "debug layout" never lists the cloister chain (V live, probe strip.py), but "debug nearby" does
// link it, with its origin offset. Without a placement `abut` finds nothing and nav says "no usable gate".
func TestPlaceLinkedGivesAGateToALevelTheLayoutLacks(t *testing.T) {
	own := Placement{Level: 26, X: 3000, Y: 1000, W: 64, H: 18}
	cases := []struct {
		name string
		near Near
		want int // 1 placed, 0 not
		side int
	}{
		{"above (Outer Cloister from the Monastery Gate)", Near{27, 0, -40}, 1, 2},
		{"below", Near{9, 0, 18}, 1, 0},
		{"left", Near{6, -80, 2}, 1, 1},
		{"right", Near{8, 64, -4}, 1, 3},
		{"far away (town listed from Cold Plains)", Near{1, -2232, 32}, 0, 0},
		{"below with a gap", Near{9, 0, 20}, 0, 0},
		{"corner only", Near{9, 64, 18}, 0, 0},
	}
	for _, c := range cases {
		got := placeLinked(own, nil, []Near{c.near})
		if len(got) != c.want {
			t.Fatalf("%s: placed %+v", c.name, got)
		}
		if c.want == 0 {
			continue
		}
		ab := abut(own, got)
		if len(ab) != 1 || ab[0].side != c.side || ab[0].p.Level != c.near.Level {
			t.Fatalf("%s: abut %+v want side %d", c.name, ab, c.side)
		}
	}
	// a level the layout does list is never replaced or doubled
	listed := []Placement{{27, 3000, 960, 56, 40}}
	if got := placeLinked(own, listed, []Near{{27, 0, -40}}); len(got) != 1 || got[0].W != 56 {
		t.Fatalf("listed level changed: %+v", got)
	}
	if got := placeLinked(Placement{Level: 1}, nil, []Near{{27, 0, -40}}); len(got) != 0 {
		t.Fatalf("own size unknown must place nothing: %+v", got)
	}
}

// End to end through the world: level 5 is linked above level 1 but absent from the layout.
func TestLinkedLevelMissingFromLayoutGetsBorderCells(t *testing.T) {
	b := newBorderStub()
	b.near = append(b.near, Near{5, 0, -3})
	w := New(openAll{b}, Options{})
	w.Update()
	lm := w.Current()
	if cells := lm.BorderCells(5); len(cells) == 0 {
		t.Fatalf("no border cells to level 5: adjacent %+v unproven %+v note %q", lm.Adjacent, lm.Unproven, w.LastBorderNote)
	}
	// the same world without the nearby entry has none (the pre-fix behaviour)
	b2 := newBorderStub()
	w2 := New(openAll{b2}, Options{})
	w2.Update()
	if w2.Current().BorderCells(5) != nil {
		t.Fatal("level 5 is not linked")
	}
}
