package senses

import (
	"os"
	"testing"

	"d2lod-bot/internal/d2map"
)

// BenchmarkObserveWalk is the Observe cost of walking one subtile at a time across the recorded Rogue Encampment, as
// the OnObserve hook does after every nav poll (docs/probes/bot-cpu-2026-11-16.md).
func BenchmarkObserveWalk(b *testing.B) {
	fh, err := os.Open("../d2map/testdata/rogue_encampment.d2mi")
	if err != nil {
		b.Fatal(err)
	}
	defer fh.Close()
	img, err := d2map.LoadImage(fh)
	if err != nil {
		b.Fatal(err)
	}
	tr := d2map.NewTracker(img)
	st, err := tr.Update()
	if err != nil {
		b.Fatal(err)
	}
	lm := tr.Current()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := &Publisher{}
		for k := 0; k < 100; k++ { // 100 subtiles of walking: about 10 s of a town walk
			if err := p.Observe(lm, "Rogue Encampment", d2map.Point{X: st.Pos.X + k, Y: st.Pos.Y + k/2}); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// The dense fog lookup answers exactly like the map it replaced, inside and outside the grid bounds.
func TestFogLookupMatchesMap(t *testing.T) {
	sv := map[block]struct{}{}
	for _, c := range []d2map.Point{{X: 100, Y: 100}, {X: 130, Y: 90}, {X: -20, Y: 7}, {X: 400, Y: 400}} {
		reveal(sv, c)
	}
	for _, b := range []d2map.Rect{{X: 90, Y: 80, W: 61, H: 50}, {X: -30, Y: -5, W: 20, H: 30}, {}} {
		f := fogLookup(sv, b)
		for y := -60; y < 460; y += 3 {
			for x := -60; x < 460; x += 3 {
				_, want := sv[block{int32(floorDiv(x)), int32(floorDiv(y))}]
				if f(x, y) != want {
					t.Fatalf("rect %v (%d,%d): got %v want %v", b, x, y, !want, want)
				}
			}
		}
	}
}
