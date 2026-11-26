package bot

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
)

// The decision ring gets typed refusals: a step refusal and a hold show their need code and wake fact in the explain view
// (and so in `ops status --why`), through the one outcome handler. Two classes. Must fail while the ring takes refusals
// only from "action failed" and the legacy gate line.
func TestExplainShowsTypedRefusalsWithNeedAndWake(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		g := newRig(t)
		g.w.set(func(w *world) { w.state = d2.StateInGame; w.area = 2; w.class = class })
		ref := refusalOf("not ready: early road (belt 0)")
		g.r.do(context.Background(), g.r.control(), StateTraveling, "x", 0, "walk", func(context.Context) error { return ref })
		g.r.holdRefused(holdNeed("needs_gold_farm", "kit trip held back: gold below the price"))
		v := g.r.Explain()
		got := map[string]ExplainRefusal{}
		for _, e := range v.Refusals {
			got[e.What] = e
		}
		s, h := got["step:route_closed"], got["hold:needs_gold_farm"]
		if s.Code != "route_closed" || s.Wake == "" || !strings.Contains(s.Need, "early road") {
			t.Fatalf("%v: step entry %+v in %+v", class, s, v.Refusals)
		}
		if h.Code != "needs_gold_farm" || h.Wake == "" {
			t.Fatalf("%v: hold entry %+v", class, h)
		}
		b, _ := json.Marshal(v)
		if !strings.Contains(string(b), `"wake":`) || !strings.Contains(string(b), `"code":"route_closed"`) {
			t.Fatalf("json: %s", b)
		}
		// the same refusal again only counts
		g.r.do(context.Background(), g.r.control(), StateTraveling, "x", 0, "walk", func(context.Context) error { return ref })
		for _, e := range g.r.Explain().Refusals {
			if e.What == "step:route_closed" && e.Count < 2 {
				t.Fatalf("count %d", e.Count)
			}
		}
	}
}
