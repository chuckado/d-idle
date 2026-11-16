package main

import (
	"net/http"
	"net/url"

	"d2lod-bot/internal/lootrules"
)

// lootRulesFixture serves /api/lootrules for the fit pages (internal/watch/web/lootrulesui.js): GET answers the real
// Starter view (the real plain-language lines, presets and protected list) with a few recent decisions; the page's query
// lootrules=custom opens it as a Custom set, lootrules=empty has no recent decisions. PUT never saves: it answers a
// refusal with a line and a field, so the error text can be pictured.
func lootRulesFixture(w http.ResponseWriter, r *http.Request) {
	variant := ""
	if ref, err := url.Parse(r.Referer()); err == nil {
		variant = ref.Query().Get("lootrules")
	}
	switch r.URL.Path {
	case "/api/lootrules/preview":
		js(w, lootrules.PreviewDoc{Schema: lootrules.PreviewSchema, Name: "Starter", AsOfItems: 31, Changed: 4, Totals: map[string]int{"sold": 9, "stashed": 6, "dropped": 0, "kept": 16},
			Note: "The dry run does not know your keep and sell marks; an item you marked keeps its mark.",
			Items: []lootrules.PreviewRow{
				{ID: 1, Name: "Rare Ring of a very long name that has to wrap on a phone", Place: "pack", Changed: true, WasTidy: "stash", WasSell: "sell",
					Tidy: lootrules.PreviewAct{Action: "stash", Rule: "t-rare-jewelry", Why: "Will stash: rare jewelry or circlet"}, Sell: lootrules.PreviewAct{Action: "keep", Rule: "s-jewelry-stash", Why: "Will stash: you own a Tome of Identify"}},
				{ID: 2, Name: "Chipped Skull", Place: "pack", Changed: true, WasTidy: "stash", WasSell: "sell",
					Tidy: lootrules.PreviewAct{Action: "keep", Rule: "default", Why: "Left alone: no rule"}, Sell: lootrules.PreviewAct{Action: "keep", Rule: "default", Why: "Left alone: no rule"}},
				{ID: 3, Name: "El Rune", Place: "stash", Changed: true, WasTidy: "keep",
					Tidy: lootrules.PreviewAct{Action: "sell", Rule: "t-sell", Why: "Will sell: vendor junk, value 1"}},
				{ID: 4, Name: "Short Sword", Place: "pack", Changed: true, WasTidy: "sell", WasSell: "sell",
					Tidy: lootrules.PreviewAct{Action: "keep", Rule: "default", Why: "Left alone"}, Sell: lootrules.PreviewAct{Action: "keep", Rule: "default", Why: "Left alone"}},
				{ID: 5, Name: "Cap", Place: "pack", WasTidy: "keep", WasSell: "keep"},
			}})
		return
	case "/api/lootrules/undo", "/api/lootrules/import":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"ok":false,"code":"no-undo","error":"there is no earlier version to go back to"}`))
		return
	case "/api/lootrules/export":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="lootrules.json"`)
		w.Write([]byte("{}\n"))
		return
	}
	if r.Method == http.MethodPut {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false,"code":"invalid","error":"the rule set has errors","errors":[{"path":"options.minRune","msg":"1 to 33"},{"path":"rules[3].when.all[1].v","msg":"unknown rune group, a very long field message that has to wrap on a phone width without a sideways scroll"}]}`))
		return
	}
	st := lootrules.OpenStore("/nonexistent/lootrules.json", []string{"Amazon", "Sorceress", "Paladin"})
	recent := []lootrules.Recent{
		{Scope: "ground", Item: "Rune: El", Today: "pickup", Rules: "pickup", Rule: "g-runes", Why: "Will pick up: rune"},
		{Scope: "ground", Item: "Chipped Skull", Today: "ignore", Rules: "pickup", Rule: "g-low-gems", Why: "Will pick up: surplus low gem, sells"},
		{Scope: "tidy", Item: "Grand Charm of a very long name that has to wrap on a phone", Today: "keep", Rules: "stash", Rule: "t-charms", Why: "Will stash: charm"},
		{Scope: "sell", Item: "Short Sword", Today: "sell", Rules: "sell", Rule: "t-sell", Why: "Will sell: vendor junk"},
		{Scope: "ground", Item: "Plain Cap", Today: "ignore", Rules: "ignore", Rule: "g-junk", Why: "Left alone: plain or inferior item"},
		{Scope: "tidy", Item: "Ring", Today: "stash", Rules: "stash", Rule: "t-rare-jewelry", Why: "Will stash: rare jewelry"},
		{Scope: "sell", Item: "Buckler", Today: "keep", Rules: "sell", Rule: "t-sell", Why: "Will sell: vendor junk"},
		{Scope: "ground", Item: "Gold: 54", Today: "pickup", Rules: "pickup", Rule: "g-gold", Why: "Will pick up: gold"},
		{Scope: "ground", Item: "Jewel", Today: "pickup", Rules: "pickup", Rule: "g-jewels", Why: "Will pick up: jewel"},
		{Scope: "tidy", Item: "Scroll of Town Portal", Today: "keep", Rules: "keep", Rule: "t-supplies", Why: "Kept: tome, scroll or key"},
	}
	if variant == "empty" {
		recent = nil
	}
	v := st.View(3, recent)
	v.CanUndo = variant != "empty"
	if variant == "custom" {
		cp := *v.Active
		cp.Preset = lootrules.PresetCustom
		cp.Rules = append([]lootrules.Rule(nil), cp.Rules...)
		cp.Rules[2].Off = true
		v.Active = &cp
	}
	js(w, v)
}
