package nav

import (
	"testing"

	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
)

// The route graph takes Levels.txt Vis links only where the engine has an exit (docs/probes/exitsread-2026-11-22.md):
// the Burial Grounds list the Pandemonium Run 1 (133) in Vis, but no exit leads there.
func TestVisLinkToPortalOnlyLevelIsNotAnEdge(t *testing.T) {
	for _, tc := range []struct {
		name string
		vis  [8]int
	}{{"Burial Grounds first", [8]int{133, 18, 19}}, {"Burial Grounds last", [8]int{18, 19, 133}}} {
		n := &Navigator{e: &env.Env{Data: &data.DB{Levels: []*data.Level{{ID: 17, Vis: tc.vis}, {ID: 18}, {ID: 19}, {ID: 133}}}}}
		n.defaults()
		adj := n.edges()
		if _, ok := adj[17][133]; ok {
			t.Errorf("%s: 17 -> 133 is an edge", tc.name)
		}
		if _, ok := adj[133][17]; ok {
			t.Errorf("%s: 133 -> 17 is an edge", tc.name)
		}
		for _, to := range []int{18, 19} {
			if _, ok := adj[17][to]; !ok {
				t.Errorf("%s: the real Vis link 17 -> %d was dropped", tc.name, to)
			}
		}
		if r, err := n.Route(17, 133); err == nil {
			t.Errorf("%s: Route(17,133) = %v, want no route", tc.name, r)
		}
	}
}
