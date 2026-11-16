package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"d2lod-bot/internal/watch"
)

// commandsFixture serves /api/commands for the Orders panel from the handler goldens of internal/watch/testdata/commands.
// Which one comes from the page's own query (the poll carries it as the Referer): cmd=empty|three-one-failed|blocked|paused|
// safety|cancel-all (default three-one-failed) or cmd=eight, the three-one-failed list with five more waiting orders, for
// the fit check of a list longer than its box; cmd=loop is the same list with a Loop row on top (round 2 of 5).
func commandsFixture(class string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := "three-one-failed"
		if ref, err := url.Parse(r.Referer()); err == nil {
			if c := ref.Query().Get("cmd"); c != "" {
				name = c
			}
		}
		file := name
		if name == "eight" || name == "loop" {
			file = "three-one-failed"
		}
		b, err := os.ReadFile(filepath.Join("internal/watch/testdata/commands", "view-"+file+"-"+class+".json"))
		if err != nil {
			http.Error(w, "no such fixture: "+name, http.StatusNotFound)
			return
		}
		var v watch.CommandsView
		if err := json.Unmarshal(b, &v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if name == "eight" {
			for i := 0; i < 5; i++ {
				v.Rows = append(v.Rows[:1], append([]watch.CommandRow{{ID: fmt.Sprintf("order:%d", 20+i), Intent: "go", Text: fmt.Sprintf("Farm Blood Moor for %d min", 10+i),
					Source: "map", State: "queued", Why: &watch.CommandWhy{Code: "order_queued", Text: "after the running order"},
					Can: watch.CommandCan{Cancel: true}, Cancel: &watch.CommandCancel{Path: "/api/orders/cancel", Body: map[string]any{"id": 20 + i}}}}, v.Rows[1:]...)...)
			}
		}
		if name == "loop" {
			lp := watch.CommandRow{ID: "loop:1", Intent: "loop", Text: "Loop: Cold Plains, round 2 of 5, 20 min each", Source: "loop", State: "active", Phase: "farming",
				Why: &watch.CommandWhy{Code: "loop_wait", Text: "repair after each round"}, Can: watch.CommandCan{Cancel: true},
				Cancel: &watch.CommandCancel{Path: "/api/orderloop", Body: map[string]any{"action": "stop"}}}
			v.Rows = append([]watch.CommandRow{lp}, v.Rows...)
		}
		js(w, v)
	}
}

// resetFixture serves /api/reset for the Reset strip (internal/watch/web/resetui.js): the page's query reset=resetting|
// rejoining|failed picks the state (the poll carries it as the Referer), else the hero is idle. The shapes are the
// goldens of internal/watch/testdata/reset.
func resetFixture(w http.ResponseWriter, r *http.Request) {
	name := "idle"
	if ref, err := url.Parse(r.Referer()); err == nil {
		if c := ref.Query().Get("reset"); c == "resetting" || c == "rejoining" || c == "failed" {
			name = c
		}
	}
	b, err := os.ReadFile(filepath.Join("internal/watch/testdata/reset", "view-"+name+".json"))
	if err != nil {
		http.Error(w, "no such fixture: "+name, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
