package main

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// The character screen's fit fixtures (card uic4, review c2c3 R14): the same fixture desk with one thing changed, chosen by the
// page's own address (/watch/inventory?fx=nomerc). The page asks /api/inventory without that query, so the answer follows the
// Referer of the request (the viewer's pages are served with Referrer-Policy: same-origin, which keeps the full address).
// Each variant is the case a defect of the first review came from: a hero with no mercenary, a dead one, names that do not
// fit a slot, requests waiting in the queue.
var fxVariants = map[string]func(map[string]any){
	// a fresh hero: the view carries a merc block that says nobody is hired
	"nomerc": func(v map[string]any) { v["merc"] = map[string]any{"state": "none", "owned": false} },
	// an older backend: no merc block at all
	"absentmerc": func(v map[string]any) { delete(v, "merc") },
	// dead, with the revive verb of the mercenary desk bound (card B3 fills m.revive) and a price
	"deadmerc": func(v map[string]any) {
		m := v["merc"].(map[string]any)
		m["state"], m["alive"], m["lifePct"] = "dead", false, 0
		m["revive"] = map[string]any{"path": "/api/merc/revive", "cost": 12300}
		m["note"] = "Mercenary is dead."
	},
	// names of one-row slots and one-cell items that cannot fit as they are
	"longnames": func(v map[string]any) {
		eq := v["equipment"].(map[string]any)
		eq["belt"] = item(5, "lbl", "Mithril Coil of the Whale's Embrace", "magic", 0, 0, 2, 1, map[string]any{"dur": 12, "maxDur": 12})
		eq["gloves"] = item(4, "lgl", "Gauntlets of the Deep Ember Weaver", "rare", 0, 0, 2, 2, map[string]any{"dur": 11, "maxDur": 12})
		eq["lring"] = item(7, "rin", "Thunderstorm Wedding Ring of the Colossus", "rare", 0, 0, 1, 1, nil)
		eq["neck"] = item(8, "amu", "Amulet of the Everlasting Dawnbringer", "magic", 0, 0, 1, 1, nil)
		pack := v["backpack"].([]any)
		pack[1].(map[string]any)["name"] = "Unbreakable-Thunderstorm-Wedding-Ring"
		pack[3].(map[string]any)["name"] = "Scroll of Town Portal and Other Long Words"
	},
	// requests that wait in the queue, one of them for a trip to town: the cells must carry the dashed outline and the TOWN mark
	"pending": func(v map[string]any) {
		v["inTown"] = false
		v["queue"] = []any{
			map[string]any{"id": 11, "kind": "stash", "slot": "", "name": "Ring", "state": "queued", "at": 0, "wait": "town", "reason": "waiting for a trip home"},
			map[string]any{"id": 10, "slot": "rhand", "name": "Staff of Charged Bolt", "state": "queued", "at": 0},
			map[string]any{"id": 16, "kind": "identify", "name": "Light Plate", "state": "queued", "at": 0},
		}
	},
}

// fxExpect is what a variant's page must show once it has loaded: selectors that must exist and have a size. It is the machine
// form of "the defect does not return"; uishots.sh passes the same ones to the harness (expect=).
var fxExpect = map[string][]string{
	"pending": {"[data-key=item-11][data-pending=stash][data-wait=town]", "[data-key=item-10][data-pending=equip]"},
}

// variantView is the fixture desk's view with the variant applied; "" or an unknown name is the plain view.
func variantView(name string) any {
	v := fixtureDesk{}.View().(map[string]any)
	if f := fxVariants[name]; f != nil {
		f(v)
	}
	return v
}

// withVariants answers /api/inventory from the variant the page's address names, and passes everything else on.
func withVariants(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if artFor(w, r) {
			return
		}
		if r.URL.Path == "/api/inventory" && r.Method == http.MethodGet {
			if ref, err := url.Parse(r.Referer()); err == nil {
				if name := ref.Query().Get("fx"); fxVariants[name] != nil {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Header().Set("Cache-Control", "no-store")
					json.NewEncoder(w).Encode(variantView(name))
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
