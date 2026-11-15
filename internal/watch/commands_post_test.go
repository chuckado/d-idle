package watch

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// POST /api/commands is a dispatcher over the old routes: each test posts through it and checks the state the old route
// gives, over the real Fun wiring of the rigs.

func cmdPost(r *funRig, token, route, body string) (int, map[string]any) {
	resp, b := r.do("POST", r.answer, route, token, body)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return resp.StatusCode, m
}

func openOrderTexts(r *funRig) []string {
	var out []string
	for _, o := range r.hub.OrdersView().Orders {
		if o.State == OrderQueued || o.State == OrderActive {
			out = append(out, o.Text)
		}
	}
	return out
}

func TestCommandsPostGatesAndBadInput(t *testing.T) {
	r := newFunRig(t, "Sorceress", 5)
	good := `{"intent":"town"}`
	for name, c := range map[string]struct {
		token, body string
		code        int
	}{
		"no token":      {"", good, 401},
		"unknown field": {funToken, `{"intent":"town","x":1}`, 400},
		"unknown":       {funToken, `{"intent":"dance"}`, 400},
		"bad mode":      {funToken, `{"intent":"go","mode":"later"}`, 400},
		"town queue":    {funToken, `{"intent":"town","mode":"queue"}`, 400},
		"not json":      {funToken, `nope`, 400},
	} {
		if code, m := cmdPost(r, c.token, CommandsRoute, c.body); code != c.code {
			t.Errorf("%s: %d %v, want %d", name, code, m, c.code)
		}
	}
	if resp, _ := r.do("GET", r.answer, CommandsRoute, funToken, ""); resp.StatusCode != 405 {
		t.Errorf("GET on the answer listener: %d", resp.StatusCode)
	}
	if code, m := cmdPost(r, funToken, CommandsCancelRoute, `{"id":"order:99"}`); code != 404 || m["code"] != "not-found" {
		t.Errorf("cancel unknown: %d %v", code, m)
	}
}

func TestCommandsPostZoneIntentsMatchTheOldRoute(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			a, b := newFunRig(t, class, 5), newFunRig(t, class, 5)
			area := "Blood Moor"
			cases := []struct{ cmd, old string }{
				{`{"intent":"go","args":{"area":"` + area + `","until":{"minutes":20}}}`, `{"kind":"go","area":"` + area + `","minutes":20,"replace":true}`},
				{`{"intent":"go","args":{"area":"` + area + `","until":"forever"}}`, `{"kind":"go","area":"` + area + `","until":"forever","replace":true}`},
				{`{"intent":"fight_here","args":{"minutes":15}}`, `{"kind":"fight","minutes":15,"replace":true}`},
				{`{"intent":"town"}`, `{"kind":"home","replace":true}`},
			}
			for _, c := range cases {
				code, m := cmdPost(a, funToken, CommandsRoute, c.cmd)
				ocode, _ := cmdPost(b, funToken, "/api/orders", c.old)
				if code != ocode || code != 200 {
					t.Fatalf("%s: %d, old route %d, %v", c.cmd, code, ocode, m)
				}
				if m["intent"] == nil || m["order"] == nil {
					t.Errorf("answer lacks intent or order: %v", m)
				}
				if x, y := strings.Join(openOrderTexts(a), "|"), strings.Join(openOrderTexts(b), "|"); x != y {
					t.Errorf("state differs: %q vs old %q", x, y)
				}
			}
		})
	}
}

func TestCommandsUndoCancelsTheNewOrder(t *testing.T) {
	r := newFunRig(t, "Paladin", 5)
	code, m := cmdPost(r, funToken, CommandsRoute, `{"intent":"go","args":{"area":"Blood Moor","until":{"minutes":10}}}`)
	if code != 200 {
		t.Fatalf("%d %v", code, m)
	}
	u, _ := m["undo"].(map[string]any)
	body, _ := json.Marshal(u["body"])
	if u["path"] != CommandsCancelRoute || len(openOrderTexts(r)) != 1 {
		t.Fatalf("undo %v open %v", u, openOrderTexts(r))
	}
	if code, m := cmdPost(r, funToken, CommandsCancelRoute, string(body)); code != 200 {
		t.Fatalf("undo: %d %v", code, m)
	}
	if got := openOrderTexts(r); len(got) != 0 {
		t.Errorf("after undo: %v", got)
	}
}

func TestCommandsCharterIsPaidInline(t *testing.T) {
	r := newFunRig(t, "Sorceress", 5)
	r.feed(funStream()...) // earns the tally the charter is paid from
	code, m := cmdPost(r, funToken, CommandsRoute, `{"intent":"go","args":{"area":"Cold Plains","until":{"minutes":30}}}`)
	ch, _ := m["charter"].(map[string]any)
	if code != 402 || m["code"] != "charter-needed" || ch == nil || ch["price"].(float64) != 20 {
		t.Fatalf("unapproved: %d %v", code, m)
	}
	if code, m := cmdPost(r, funToken, CommandsRoute, `{"intent":"go","args":{"area":"Cold Plains","until":{"minutes":30},"approve":true}}`); code != 200 {
		t.Fatalf("approved: %d %v", code, m)
	}
}

func TestCommandsQueueAndLoopShareTheOldGates(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			g := newLoopRig(t, class)
			q := `{"intent":"go","mode":"queue","args":{"area":"Blood Moor","until":{"minutes":10}}}`
			l := `{"intent":"loop","args":{"area":"Blood Moor","minutes":20,"repair":"damaged","rounds":2}}`
			if code, m := cmdPost(g.funRig, funToken, CommandsRoute, q); code != 403 || m["code"] != "locked" {
				t.Errorf("queue locked: %d %v", code, m)
			}
			if code, m := cmdPost(g.funRig, funToken, CommandsRoute, l); code != 409 || m["code"] != "locked" {
				t.Errorf("loop locked: %d %v", code, m)
			}
			g.openLoop()
			code, m := cmdPost(g.funRig, funToken, CommandsRoute, l)
			if code != 200 || g.lview().State != LoopActive {
				t.Fatalf("loop start: %d %v %+v", code, m, g.lview())
			}
			u, _ := m["undo"].(map[string]any)
			if u == nil || u["path"] != CommandsCancelRoute {
				t.Fatalf("loop undo: %v", m)
			}
			body, _ := json.Marshal(u["body"])
			if code, m := cmdPost(g.funRig, funToken, CommandsCancelRoute, string(body)); code != 200 {
				t.Fatalf("loop undo: %d %v", code, m)
			}
			if s := g.lview().State; s == LoopActive {
				t.Errorf("loop still active after undo")
			}
			g.denCleared()
			if code, m := cmdPost(g.funRig, funToken, CommandsRoute, q); code != 200 || m["entry"] == nil {
				t.Errorf("queue add: %d %v", code, m)
			}
		})
	}
}

func TestCommandsUnboundDeskIsOff(t *testing.T) {
	r := newFunRig(t, "Sorceress", 5)
	for _, c := range []string{`{"intent":"pause"}`, `{"intent":"farm_gold"}`, `{"intent":"repair"}`, `{"intent":"revive_merc"}`} {
		code, m := cmdPost(r, funToken, CommandsRoute, c)
		if code != http.StatusServiceUnavailable && code != http.StatusConflict {
			t.Errorf("%s: %d %v, want 503 off", c, code, m)
		}
	}
}

func TestCommandsTranslateTable(t *testing.T) {
	for _, c := range []struct {
		in   commandBody
		path string
	}{
		{commandBody{Intent: "go", Args: map[string]any{"area": "X"}}, "/api/orders"},
		{commandBody{Intent: "go", Mode: "queue", Args: map[string]any{"area": "X"}}, "/api/orderqueue"},
		{commandBody{Intent: "loop"}, "/api/orderloop"},
		{commandBody{Intent: "repair"}, "/api/repair/request"},
		{commandBody{Intent: "repair", Args: map[string]any{"cancel": true}}, "/api/repair/cancel"},
		{commandBody{Intent: "farm_gold"}, GoldFarmRoute},
		{commandBody{Intent: "pause"}, "/api/pause"},
		{commandBody{Intent: "resume"}, "/api/pause"},
		{commandBody{Intent: "sell_now"}, RouteSellNow},
		{commandBody{Intent: "revive_merc"}, "/api/merc/revive"},
		{commandBody{Intent: "identify"}, "/api/inventory/identify"},
		{commandBody{Intent: "journey"}, CampaignIntentRoute},
	} {
		if got, why := c.in.translate(); got.path != c.path || why != "" {
			t.Errorf("%s/%s: %q (%s), want %q", c.in.Intent, c.in.Mode, got.path, why, c.path)
		}
	}
}
