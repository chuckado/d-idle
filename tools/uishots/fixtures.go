package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Fixture panels for the fit check and the cockpit pictures: the real cockpit panels (orders, zone-map, area-map) are used as they
// are; needs-you and the strip are other slices, so the fixture viewer adds stand-ins that fill the slots the way those will: fixed boxes,
// one-line rows cut with an ellipsis, a long list that scrolls inside (data-scroll). They exist only here: panels.json
// and /watch/layouts/cockpit-demo.json, -overflow.json are extended by this viewer and nothing else serves them.

type fxPanel struct{ id, title, js string }

var fxPanels = []fxPanel{
	{"fx-needs", "Needs you", `const n = Math.max(0, Math.min(9, +(new URLSearchParams(location.search).get('needs') ?? 2)));
  ctx.hide(n === 0);
  const rows = ['Weapon broken, repair now?', 'Go for Blood Raven? Level 9 is the guide', 'Belt is empty, buy potions', 'Mercenary is dead', 'Stash is full'];
  for (let i = 0; i < Math.min(n, 3); i++) { const r = ctx.el('div', 'row line'); r.append(ctx.el('span', 'txt', rows[i]), ctx.ui.btn('Do it', { size: 'sm' })); ctx.body.append(r); }
  if (n > 3) ctx.body.append(ctx.el('div', 'row', '+' + (n - 3) + ' more'));`},
	{"fx-orders", "Orders", `ctx.body.append(ctx.el('div', 'row', 'Orders  1 open'));
  const verbs = ctx.el('div', 'row verbs');
  for (const v of ['Go to', 'Town', 'Repair', 'Pause', 'More']) verbs.append(ctx.ui.btn(v, { size: 'sm' }));
  const list = ctx.el('div', 'list'); list.dataset.scroll = '1';
  const n = Math.max(1, Math.min(30, +(new URLSearchParams(location.search).get('orders') ?? 4)));
  for (let i = 0; i < n; i++) list.append(ctx.el('div', 'row', i === 0 ? 'Farm Cold Plains 30 min  ACTIVE' : 'then: level 10, order number ' + i + ' of the plan that runs long'));
  ctx.body.append(verbs, list);`},
	{"fx-progress", "Progress", `ctx.body.append(ctx.el('div', 'row', 'Progress  level 7'), ctx.el('div', 'row', 'to 10 in 25 min'), ctx.el('div', 'row', 'xp/h 12.4k  gold/h 2.1k'));`},
	{"fx-quest", "Quest", `ctx.body.append(ctx.el('div', 'row', 'Quest  Sisters' + "'" + ' Burial Grounds'), ctx.el('div', 'row', 'Blood Raven: not ready for this hero yet, a long line that is cut'));`},
	{"fx-health", "Health", `ctx.body.append(ctx.el('div', 'row', 'Health'), ctx.el('div', 'row', 'Life ok  Gear ok  Ready ok'));`},
	// the crafted pages the check must fail (docs/ui/redesign-2026-11-14/reviews/s1b-review.md): content cut one level below
	// the panel root, a data-scroll that does not scroll, a slot whose only panel is hidden, a frame taller than the window
	{"fx-nested", "Nested", `const box = ctx.el('div', 'nest'); for (let i = 0; i < 60; i++) box.append(ctx.el('div', 'row', 'row ' + i + ' in a box that cuts')); ctx.body.append(box);`},
	{"fx-fakescroll", "Fake scroll", `const box = ctx.el('div', 'fake'); box.dataset.scroll = '1'; for (let i = 0; i < 60; i++) box.append(ctx.el('div', 'row', 'row ' + i + ' under a data-scroll that does not scroll')); ctx.body.append(box);`},
	{"fx-blank", "Blank", `ctx.hide(true);`},
	{"fx-tall", "Tall", `ctx.body.append(ctx.el('div', 'row', 'a frame taller than the window'));`},
	// not a stand-in: a panel that overflows its box on purpose, for the fixture the fit check must fail
	{"fx-overflow", "Overflow", `for (let i = 0; i < 60; i++) ctx.body.append(ctx.el('div', 'row', 'row ' + i + ' of a panel with no height of its own'));`},
}

const fxCSS = `/* unlayered on purpose: the tall-frame fixture must beat the shell's layer */
.d-main[data-template="cockpit"]:has([data-panel="fx-tall"]) { block-size: 130dvh !important; }
@layer panels {
  [data-panel^="fx-"] .row { min-block-size: 28px; line-height: 28px; font-size: var(--d-fs-1); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  [data-panel^="fx-"] .row.line, [data-panel^="fx-"] .row.verbs { display: flex; align-items: center; gap: var(--d-sp-2); line-height: normal; }
  [data-panel^="fx-"] .row .txt { flex: 1 1 auto; min-inline-size: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  [data-cockpit-slot="strip"] [data-panel^="fx-"] .row { min-block-size: 24px; line-height: 24px; }
  [data-panel="fx-nested"] .nest { block-size: 100%; overflow: hidden; }
  [data-panel="fx-nested"] .d-slotbody { overflow: hidden; }
  [data-panel="fx-orders"] .d-slotbody { display: flex; flex-direction: column; }
  [data-panel="fx-orders"] .row { flex: none; }
  [data-panel="fx-orders"] .list { overflow: auto; min-block-size: 0; }
}
`

// orderedTabs marshals as a JSON object in the order given (a Go map would sort the tab names).
type orderedTabs []struct {
	Name string
	IDs  []string
}

func (t orderedTabs) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, x := range t {
		if i > 0 {
			b.WriteByte(',')
		}
		k, _ := json.Marshal(x.Name)
		v, _ := json.Marshal(x.IDs)
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// fxLayout is a cockpit layout over the fixture panels. Names: "" the demo, "overflow" the panel with no height of its own,
// and the crafted pages of the review, each with its one panel in every arrangement so every width has the fault.
func fxLayout(name string) map[string]any {
	// the area map is the real panel (S6); the rest stand in for the slices that are not built
	rail := []string{"fx-needs", "fx-orders", "zone-map", "area-map"}
	wide := map[string]any{"top": []string{"fx-needs", "fx-orders"}, "left": []string{"zone-map"}, "right": []string{"area-map"}}
	tabs := orderedTabs{{"Orders", []string{"fx-needs", "fx-orders"}}, {"Map", []string{"zone-map", "area-map"}}, {"Progress", []string{"fx-progress", "fx-quest", "fx-health"}}}
	if name == "orders" { // the demo with the real Orders panel (fed by commandsFixture) in place of its stand-in
		rail = []string{"fx-needs", "orders", "zone-map", "area-map"}
		wide["top"] = []string{"fx-needs", "orders"}
		tabs[0].IDs = []string{"fx-needs", "orders"}
	} else if name == "zoneclear" { // the real Zone clear panel (fed by zoneClearFixture) with the needs panel above it, the slot it will be dropped into
		rail = []string{"fx-needs", "zoneclear", "zone-map", "area-map"}
		wide["top"] = []string{"fx-needs"} // the top row sizes itself: a 300 px panel belongs in a column
		wide["left"] = []string{"zoneclear", "zone-map"}
		tabs[0].IDs = []string{"fx-needs", "zoneclear"}
	} else if name != "" {
		one := "fx-" + name
		rail = []string{one}
		wide = map[string]any{"top": []string{one}, "left": []string{}, "right": []string{}}
		tabs = orderedTabs{{"Fault", []string{one}}}
	}
	return map[string]any{
		"schema": "watch.layout/1", "cols": 12, "rowHeight": 48, "density": "dense", "theme": "dark", "template": "cockpit/1",
		"slots": map[string]any{
			"band": []string{"statusline", "hero-meters", "needs-count"}, "stage": "video",
			"strip": []string{"fx-progress", "fx-quest", "fx-health"}, "rail": rail,
			"railWide": wide, "phoneTabs": tabs,
		},
		"sizes":       map[string]any{"strip": 112, "railMin": 400},
		"breakpoints": map[string]any{},
	}
}

// mountFixturePanels wraps h: it adds the fixture manifests to /watch/panels.json and serves the fixture panels and
// layouts; everything else is h's.
func mountFixturePanels(h http.Handler) http.Handler {
	byID := map[string]fxPanel{}
	for _, p := range fxPanels {
		byID[p.id] = p
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case p == "/__fit":
			fitHarness(w, r)
		case (p == "/watch" || p == "/watch/inventory" || p == "/watch/characters" || p == "/watch/vault") && r.URL.Query().Get("fitframe") == "1":
			framedPage(h, w, r)
		case strings.HasPrefix(p, "/__fit/defect-"):
			if !serveDefectCSS(w, r) && !serveDefectJS(w, r) {
				http.NotFound(w, r)
			}
		case p == "/__fit/fit.js":
			serveFitJS(w, r)
		case strings.HasPrefix(p, "/watch/layouts/cockpit-") && strings.HasSuffix(p, ".json"):
			name := strings.TrimSuffix(strings.TrimPrefix(p, "/watch/layouts/cockpit-"), ".json")
			if name == "demo" {
				name = ""
			}
			if name != "" && name != "orders" && name != "zoneclear" && byID["fx-"+name].id == "" {
				http.NotFound(w, r)
				return
			}
			js(w, fxLayout(name))
		case p == "/watch/panels.json":
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)
			var v map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
				w.WriteHeader(rec.Code)
				w.Write(rec.Body.Bytes())
				return
			}
			list, _ := v["panels"].([]any)
			for _, f := range fxPanels {
				man := map[string]any{"schema": "watch.panel/1", "version": 1, "id": f.id, "title": f.title, "band": "side", "size": "s",
					"priority": 50, "tab": "hero", "defaultOpen": true, "endpoint": "", "poll": 0, "unlock": "", "skeleton": 0, "slot": true, "scroll": f.id == "fx-orders" || f.id == "fx-fakescroll",
					"owner": "uishots", "doc": "docs/ui-regression-plan.md"}
				if f.id == "zone-map" {
					man["switchWith"], man["switchLabel"], man["minBlock"] = "area-map", "Act map", 250 // the stand-in for zone-map shares its column with the area map
				}
				list = append(list, man)
			}
			v["panels"] = list
			js(w, v)
		case strings.HasPrefix(p, "/watch/panels/fx-"):
			parts := strings.Split(strings.TrimPrefix(p, "/watch/panels/"), "/")
			f, ok := byID[parts[0]]
			if !ok || len(parts) != 2 {
				http.NotFound(w, r)
				return
			}
			if parts[1] == "panel.css" {
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
				w.Write([]byte(fxCSS))
				return
			}
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			w.Write([]byte("export function mount(root, ctx) {\n  " + f.js + "\n}\n"))
		default:
			h.ServeHTTP(w, r)
		}
	})
}
