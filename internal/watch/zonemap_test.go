package watch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The zone map: the handler goldens of GET /api/orders/map with the S5 fields (here, target, two pins, a failed mark, the
// Loop and queue ladder) for both classes, and the DOM goldens of the zone-map panel over them (testdata/zonemap/zonemap_dom.js).
// UPDATE=1 rewrites the files.

type zmClass struct {
	level                        int
	area                         string
	areaID                       int
	target, pin1, pin2, failedAt string
}

var zmClasses = map[string]zmClass{
	"sorceress": {7, "Blood Moor", 2, "Cold Plains", "Stony Field", "Dark Wood", "Burial Grounds"},
	"paladin":   {12, "Stony Field", 4, "Dark Wood", "Black Marsh", "Cold Plains", "Crypt"},
}

func zoneMapRig(t *testing.T, class string) (*funRig, MapView) {
	c := zmClasses[class]
	r := newFunRig(t, class, c.level)
	r.hub.SetHero(HeroInfo{Class: class, Level: c.level, Area: c.area, AreaID: c.areaID})
	first, err := r.hub.Give(OrderRequest{Kind: OrderGo, Area: c.target, Minutes: 30})
	if err != nil {
		t.Fatal(err)
	}
	r.hub.OrderState(first.ID, OrderActive, "")
	for _, a := range []string{c.pin1, c.pin2} {
		if _, err := r.hub.Give(OrderRequest{Kind: OrderGo, Area: a, Minutes: 20}); err != nil {
			t.Fatal(err)
		}
	}
	bad, err := r.hub.Give(OrderRequest{Kind: OrderGo, Area: c.failedAt, Minutes: 20})
	if err != nil {
		t.Fatal(err)
	}
	r.hub.OrderState(bad.ID, OrderFailed, "the hero's guard turned it down")
	r.hub.SetMapLadder(func() *MapLadder {
		return &MapLadder{LoopOpen: true, QueueOpen: true, QueueSlots: 2, QueueWaiting: 1, QueueTeaser: "Clear the Den of Evil for a second slot."}
	})
	return r, r.hub.OrdersMap(1)
}

func TestZoneMapGoldens(t *testing.T) {
	if err := os.MkdirAll("testdata/zonemap", 0o755); err != nil {
		t.Fatal(err)
	}
	for class := range zmClasses {
		c := zmClasses[class]
		r, v := zoneMapRig(t, class)
		byID := map[string]MapNodeView{}
		for _, n := range v.Nodes {
			byID[n.Name] = n
		}
		if n := byID[c.area]; !n.Status.Here {
			t.Errorf("%s: %s is not marked here", class, c.area)
		}
		if n := byID[c.target]; !n.Status.Order {
			t.Errorf("%s: %s is not the target", class, c.target)
		}
		if byID[c.pin1].Status.Queued != 1 || byID[c.pin2].Status.Queued != 2 {
			t.Errorf("%s: pins %d %d", class, byID[c.pin1].Status.Queued, byID[c.pin2].Status.Queued)
		}
		if n := byID[c.failedAt]; !n.Status.Failed || n.Status.FailedWhy == "" {
			t.Errorf("%s: no failed mark on %s", class, c.failedAt)
		}
		if v.Ladder == nil || !v.Ladder.LoopOpen {
			t.Errorf("%s: no ladder", class)
		}
		if a2 := r.hub.OrdersMap(2); a2.Populated || len(a2.Areas) == 0 {
			t.Errorf("act 2 lists no orderable areas: %+v", a2.Areas)
		}
		b, _ := json.MarshalIndent(v, "", " ")
		write := func(name string, b []byte) {
			p := filepath.Join("testdata/zonemap", name)
			if os.Getenv("UPDATE") != "" {
				if err := os.WriteFile(p, append(b, '\n'), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(p)
			if err != nil || string(want) != string(append(b, '\n')) {
				t.Errorf("%s differs (UPDATE=1 rewrites)", p)
			}
		}
		write("map-fixture-"+class+".json", b)
		a2, _ := json.MarshalIndent(r.hub.OrdersMap(2), "", " ")
		write("map-act2-"+class+".json", a2)
	}
}

func TestZoneMapPanelDOM(t *testing.T) {
	runNode(t, "testdata/zonemap/zonemap_dom.js", "web/panels")
}
