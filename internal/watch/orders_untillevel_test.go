package watch

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// Until-level for the zone orders (docs/ui/redesign-2026-11-14, B4): POST /api/orders {until: "level", level: N}.
func TestOrderUntilLevelValidation(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} { // sim heroes at level 6 and 12
		r := newRig(t, build)
		lvl := r.sim.Hub.heroNow().Level
		for _, c := range []struct {
			name string
			body map[string]any
			ok   bool
		}{
			{"go above the hero", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level", "level": lvl + 2}, true},
			{"fight above the hero", map[string]any{"kind": "fight", "until": "level", "level": lvl + 1}, true},
			{"at the hero's level", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level", "level": lvl}, false},
			{"below the hero", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level", "level": 2}, false},
			{"level 1", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level", "level": 1}, false},
			{"above the table", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level", "level": 100}, false},
			{"no level", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level"}, false},
			{"with minutes", map[string]any{"kind": "go", "area": "Blood Moor", "until": "level", "level": lvl + 3, "minutes": 5}, false},
			{"home", map[string]any{"kind": "home", "until": "level", "level": lvl + 3}, false},
			{"level without until", map[string]any{"kind": "go", "area": "Blood Moor", "minutes": 5, "level": lvl + 3}, false},
			{"unknown until", map[string]any{"kind": "go", "area": "Blood Moor", "until": "sunrise"}, false},
		} {
			var ov OrdersView
			r.get("/api/orders", &ov)
			for _, o := range ov.Orders {
				if o.State == OrderQueued || o.State == OrderActive {
					r.post("/api/orders/cancel", map[string]any{"id": o.ID})
				}
			}
			code, rp := r.post("/api/orders", c.body)
			switch {
			case c.ok && (code != 200 || !rp.OK || rp.Order == nil || rp.Order.UntilLevel == 0):
				t.Errorf("%s/%s: %d %+v", build, c.name, code, rp)
			case !c.ok && (code != 400 || rp.Code != "bad-order"):
				t.Errorf("%s/%s: %d %+v", build, c.name, code, rp)
			}
		}
	}
}

// An until-level order reads "until level N", reaches the bot as an endless order, ends done when the hero gets there,
// and the reply of POST /api/orders keeps every old field.
func TestOrderUntilLevelEndsAtTheLevel(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} {
		r := newRig(t, build)
		h := r.sim.Hub
		start := h.heroNow().Level
		code, rp := r.post("/api/orders", map[string]any{"kind": "go", "area": "Dark Wood", "until": "level", "level": start + 2})
		o := rp.Order
		if code != 200 || o == nil || o.Text != "Farm Dark Wood until level "+strconv.Itoa(start+2) || !o.Forever || o.Minutes != 0 || o.UntilLevel != start+2 {
			t.Fatalf("%s: %d %+v", build, code, rp)
		}
		b, _ := json.Marshal(o)
		for _, k := range []string{`"id"`, `"kind"`, `"area"`, `"forever"`, `"state"`, `"text"`, `"givenAt"`, `"untilLevel"`} {
			if !strings.Contains(string(b), k) {
				t.Errorf("%s: order JSON lacks %s: %s", build, k, b)
			}
		}
		h.TickOrders()
		if len(h.openIDs()) != 1 {
			t.Fatalf("%s: ended before the level", build)
		}
		hero := h.heroNow()
		hero.Level = start + 1
		h.SetHero(hero)
		h.TickOrders()
		if len(h.openIDs()) != 1 {
			t.Fatalf("%s: ended one level early", build)
		}
		hero.Level = start + 2
		h.SetHero(hero)
		h.TickOrders()
		var ov OrdersView
		r.get("/api/orders", &ov)
		last := ov.Orders[len(ov.Orders)-1]
		if len(h.openIDs()) != 0 || last.State != OrderDone || !strings.Contains(last.Note, "reached level") {
			t.Fatalf("%s: %+v", build, last)
		}
		h.TickOrders() // idempotent
	}
}

// A plain timed order and the old forever order are unchanged.
func TestOrderUntilLevelLeavesOldOrdersAlone(t *testing.T) {
	r := newRig(t, "sorceress")
	_, rp := r.post("/api/orders", map[string]any{"kind": "go", "area": "Dark Wood", "minutes": 15})
	if rp.Order == nil || rp.Order.UntilLevel != 0 || rp.Order.Text != "Farm Dark Wood for 15 min" {
		t.Fatalf("%+v", rp.Order)
	}
	b, _ := json.Marshal(rp.Order)
	if strings.Contains(string(b), "untilLevel") {
		t.Errorf("old order carries untilLevel: %s", b)
	}
	_, rp = r.post("/api/orders", map[string]any{"kind": "go", "area": "Dark Wood", "until": "forever", "replace": true})
	if rp.Order == nil || rp.Order.Text != "Farm Dark Wood until you cancel" {
		t.Fatalf("%+v", rp.Order)
	}
	hero := r.sim.Hub.heroNow()
	hero.Level = 90
	r.sim.Hub.SetHero(hero)
	r.sim.Hub.TickOrders()
	if len(r.sim.Hub.openIDs()) != 1 {
		t.Error("a forever order ended on a level")
	}
}

// Fun.Settle is what ends the order in the real process.
func TestOrderUntilLevelEndsThroughSettle(t *testing.T) {
	for _, class := range []string{"Sorceress", "Paladin"} {
		r := newFunRig(t, class, 5)
		o, err := r.hub.Give(OrderRequest{Kind: OrderGo, Area: "Blood Moor", UntilLevel: 7})
		if err != nil {
			t.Fatal(err)
		}
		r.fun.Settle()
		if len(r.hub.openIDs()) != 1 {
			t.Fatalf("%s: ended early", class)
		}
		r.hub.SetHero(HeroInfo{Class: class, Level: 7, Area: "Blood Moor"})
		r.fun.Settle()
		if st, _ := r.hub.orderStateOf(o.ID); st != OrderDone {
			t.Fatalf("%s: state %q", class, st)
		}
	}
}
