package watch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// POST /api/commands and POST /api/commands/cancel (docs/ui/redesign-2026-11-14/C-orders.md section 6): one door for the
// Orders panel and the zone map over the doors that already exist. The dispatcher turns {intent, args, mode} into the
// body of the existing route and serves that route in-process on the same mux, so the gates (origin list, token, JSON,
// small body, known fields), the answers and the error codes are the old route's own, and the Loop order and the zone
// menu share this one code. A route whose desk is not bound answers 503 off.

const (
	// CommandsCancelRoute is POST (answer listener).
	CommandsCancelRoute = CommandsRoute + "/cancel"
	cmdModeNow          = "now"
	cmdModeQueue        = "queue"
)

type commandBody struct {
	Intent string         `json:"intent"`
	Args   map[string]any `json:"args"`
	Mode   string         `json:"mode"` // "" or now (the order runs next and replaces the open ones) or queue (a place in the order queue)
}

type commandCancelBody struct {
	ID string `json:"id"`
}

// itemRoutes are the item intents that pass their args to the inventory route of the same name (the desk checks them).
var itemRoutes = map[string]string{
	"identify": "/api/inventory/identify", "buyid": "/api/inventory/buyid", "cain": "/api/inventory/cain",
	"stash": "/api/inventory/stash", "take": "/api/inventory/take", "gold": "/api/inventory/gold",
	"equip": "/api/inventory/equip", "unequip": "/api/inventory/unequip", "trash": "/api/inventory/trash", "mark": "/api/inventory/mark",
}

type cmdReq struct {
	path string
	body map[string]any
}

func argStr(a map[string]any, k string) string { s, _ := a[k].(string); return s }
func argInt(a map[string]any, k string) int {
	f, _ := a[k].(float64)
	return int(f)
}
func argBool(a map[string]any, k string) bool { b, _ := a[k].(bool); return b }

// orderBody is the body of an order (the orders route and the queue route read the same fields).
func orderBody(b commandBody, kind string) map[string]any {
	a := b.Args
	o := map[string]any{"kind": kind}
	if kind == OrderGo {
		o["area"] = argStr(a, "area")
	}
	if kind != OrderHome {
		switch u := a["until"].(type) {
		case string:
			o["until"] = u
		case map[string]any: // {"minutes": N} | {"level": N} | {"forever": true}, the shape C-orders gives
			switch {
			case u["level"] != nil:
				o["until"], o["level"] = "level", argInt(u, "level")
			case argBool(u, "forever"):
				o["until"] = "forever"
			default:
				o["minutes"] = argInt(u, "minutes")
			}
		}
		if m := argInt(a, "minutes"); m != 0 {
			o["minutes"] = m
		}
		if l := argInt(a, "level"); l != 0 {
			o["level"] = l
		}
	}
	if argBool(a, "approve") {
		o["approve"] = true
	}
	return o
}

// translate maps a command to the existing route; the second result is the error text of a command it cannot map.
func (b commandBody) translate() (cmdReq, string) {
	a := b.Args
	if a == nil {
		a = map[string]any{}
		b.Args = a
	}
	if b.Mode != "" && b.Mode != cmdModeNow && b.Mode != cmdModeQueue {
		return cmdReq{}, "mode must be now or queue"
	}
	zone := func(kind string) (cmdReq, string) {
		o := orderBody(b, kind)
		if b.Mode == cmdModeQueue {
			if kind == OrderHome {
				return cmdReq{}, "town cannot wait in the queue"
			}
			o["action"] = "add"
			return cmdReq{"/api/orderqueue", o}, ""
		}
		o["replace"] = true // Go now: the new order replaces the open ones (what the order sheet has always sent)
		return cmdReq{"/api/orders", o}, ""
	}
	switch b.Intent {
	case "go":
		return zone(OrderGo)
	case "fight_here":
		return zone(OrderHere)
	case "town":
		return zone(OrderHome)
	case "loop":
		if argStr(a, "action") == "stop" {
			return cmdReq{"/api/orderloop", map[string]any{"action": "stop"}}, ""
		}
		o := map[string]any{"action": "start", "area": argStr(a, "area"), "minutes": argInt(a, "minutes"), "repair": argStr(a, "repair")}
		if argBool(a, "forever") {
			o["forever"] = true
		} else if r := argInt(a, "rounds"); r != 0 {
			o["rounds"] = r
		}
		if argBool(a, "approve") {
			o["approve"] = true
		}
		return cmdReq{"/api/orderloop", o}, ""
	case "repair":
		if argBool(a, "cancel") {
			return cmdReq{"/api/repair/cancel", map[string]any{}}, ""
		}
		return cmdReq{"/api/repair/request", map[string]any{}}, ""
	case "farm_gold":
		on, ok := a["on"].(bool)
		if !ok {
			on = true
		}
		return cmdReq{GoldFarmRoute, map[string]any{"on": on}}, ""
	case "pause":
		return cmdReq{PauseRoute, map[string]any{"on": true}}, ""
	case "resume":
		return cmdReq{PauseRoute, map[string]any{"on": false}}, ""
	case "sell_now":
		return cmdReq{RouteSellNow, a}, ""
	case "revive_merc":
		return cmdReq{"/api/merc/revive", a}, ""
	}
	if p, ok := itemRoutes[b.Intent]; ok {
		return cmdReq{p, a}, ""
	}
	return cmdReq{}, "unknown intent " + safeText(b.Intent, 30)
}

// undoFor is the request that takes an accepted command back (the five-second Undo of the panel); nil where none exists.
func undoFor(b commandBody, inner map[string]any) map[string]any {
	cancel := func(id string) map[string]any {
		return map[string]any{"path": CommandsCancelRoute, "body": map[string]any{"id": id}}
	}
	switch b.Intent {
	case "go", "fight_here", "town":
		if b.Mode == cmdModeQueue {
			if e, ok := inner["entry"].(map[string]any); ok {
				return cancel(fmt.Sprintf("queue:%d", int(argInt(e, "id"))))
			}
			return nil
		}
		if o, ok := inner["order"].(map[string]any); ok {
			return cancel(fmt.Sprintf("order:%d", int(argInt(o, "id"))))
		}
	case "loop":
		if argStr(b.Args, "action") != "stop" {
			return cancel("loop:1")
		}
	case "repair":
		if !argBool(b.Args, "cancel") {
			return cancel("repair:1")
		}
	case "farm_gold":
		on, ok := b.Args["on"].(bool)
		if !ok || on {
			return map[string]any{"path": CommandsRoute, "body": map[string]any{"intent": "farm_gold", "args": map[string]any{"on": false}}}
		}
	case "pause":
		return map[string]any{"path": CommandsRoute, "body": map[string]any{"intent": "resume"}}
	case "resume":
		return map[string]any{"path": CommandsRoute, "body": map[string]any{"intent": "pause"}}
	}
	return nil
}

// serveInner runs the existing route in-process with the caller's headers (origin, token, content type), so its own gates judge.
func serveInner(mux http.Handler, r *http.Request, req cmdReq) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(req.body)
	ir := httptest.NewRequest(http.MethodPost, req.path, bytes.NewReader(raw)).WithContext(r.Context())
	for _, h := range []string{"Origin", "Authorization", "Cookie", "Content-Type"} {
		if v := r.Header.Get(h); v != "" {
			ir.Header.Set(h, v)
		}
	}
	ir.RemoteAddr = r.RemoteAddr
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, ir)
	return rec
}

// relay copies the inner answer; a 404 of the mux itself (the desk is not bound) is 503 off like the desks' own answer.
func relay(w http.ResponseWriter, rec *httptest.ResponseRecorder, extra func(map[string]any)) {
	for k, v := range rec.Header() {
		if strings.HasPrefix(k, "Access-Control") || k == "Vary" {
			w.Header()[k] = v
		}
	}
	var out map[string]any
	if json.Unmarshal(rec.Body.Bytes(), &out) != nil || out == nil {
		out = map[string]any{}
	}
	code := rec.Code
	if code == http.StatusNotFound && out["code"] == "route" {
		code, out = http.StatusServiceUnavailable, map[string]any{"ok": false, "code": "off", "error": "that desk is not available"}
	}
	if extra != nil && code < 300 {
		extra(out)
	}
	writeJSON(w, code, out)
}

// CommandsPostHandler serves POST /api/commands and /api/commands/cancel over mux (the answer listener's own mux, so the
// old routes are the ones served). o is the answer options of the listener; src builds the rows the cancel looks up.
func CommandsPostHandler(o AnswerOptions, mux http.Handler, src func() CommandSources) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !unlockPostGate(o, w, r) {
			return
		}
		fail := func(code int, c, msg string) { writeJSON(w, code, reply{Code: c, Error: msg}) }
		if r.URL.Path == CommandsCancelRoute {
			var b commandCancelBody
			if decodeSmall(r, &b) != nil || b.ID == "" {
				fail(400, "body", "send {\"id\": \"order:12\"}")
				return
			}
			for _, row := range BuildCommands(src()).Rows {
				if row.ID != b.ID {
					continue
				}
				if !row.Can.Cancel || row.Cancel == nil {
					fail(409, "order-state", "that command cannot be cancelled now")
					return
				}
				relay(w, serveInner(mux, r, cmdReq{row.Cancel.Path, row.Cancel.Body}), nil)
				return
			}
			fail(404, "not-found", "no such command")
			return
		}
		var b commandBody
		if decodeSmall(r, &b) != nil {
			fail(400, "body", "body must be one small JSON object with known fields")
			return
		}
		req, why := b.translate()
		if why != "" {
			fail(400, "bad-command", why)
			return
		}
		relay(w, serveInner(mux, r, req), func(out map[string]any) {
			out["intent"] = b.Intent
			if u := undoFor(b, out); u != nil {
				out["undo"] = u
			}
		})
	})
}
