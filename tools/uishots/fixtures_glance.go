package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"d2lod-bot/internal/watch"
)

// glanceFixture serves /api/glance for the strip panels and Needs you: the hub's own reply (the sentence, the readiness)
// with the blocks the frozen fixture hub cannot build (rates need minutes of play, the journey, gear, mercenary and supplies
// have no source here). The page's own query picks the variant (the poll carries it as the Referer):
//
//	needs=N     how many rows Needs you shows (0 to 9, default 2)
//	glance=     full (default), measuring (rates still building), unknown (mercenary and supplies not reported),
//	            gearbroken, ch1, ch9, blocked (the journey is held), nojourney; further words after commas add a state:
//	            paused (the player pause on the band), mercdead (the mercenary is dead), supplieslow (belt low),
//	            nopotions (no mercenary and no potions), e.g. glance=full,paused,mercdead
func glanceFixture(h *watch.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		if ref, err := url.Parse(r.Referer()); err == nil {
			q = ref.Query()
		}
		v := h.Glance()
		n := 2
		if s := q.Get("needs"); s != "" {
			n, _ = strconv.Atoi(s)
		}
		v.Needs = fixtureNeeds(n)
		words := strings.Split(q.Get("glance"), ",")
		variant := words[0]
		has := func(w string) bool {
			for _, x := range words {
				if x == w {
					return true
				}
			}
			return false
		}
		pctOf := func(f float64) *float64 { return &f }
		xp, gold := 12400.0, 2100.0
		eta := 1500
		v.Progress = watch.GlanceProgress{
			Level:       watch.GlanceLevel{From: 7, To: 8, Pct: pctOf(62), ETASeconds: &eta},
			XPPerHour:   &xp,
			XPSeries:    []float64{8000, 9000, 9500, 10000, 11000, 10000, 12000, 11500, 12400, 12000, 13100, 12200, 11800, 12900, 13300, 12400, 12100, 12600, 13000, 12700, 12300, 12800, 13500, 12900, 12400, 12600, 12200, 12700, 12500, 12400},
			GoldPerHour: &gold,
			GoldSeries:  []float64{1900, 2000, 2100},
		}
		if variant == "measuring" {
			v.Progress.Level.ETASeconds = nil
			v.Progress.Level.Pct = pctOf(4)
			v.Progress.XPPerHour, v.Progress.GoldPerHour = nil, nil
			v.Progress.XPSeries, v.Progress.GoldSeries = []float64{}, []float64{}
		}
		chapters := func(cur int, st string) []watch.GlanceChapter {
			out := make([]watch.GlanceChapter, 9)
			for i := range out {
				s := "todo"
				if i+1 < cur {
					s = "done"
				} else if i+1 == cur {
					s = st
				}
				out[i] = watch.GlanceChapter{ID: fmt.Sprintf("ch%d", i+1), State: s}
			}
			return out
		}
		verbs := []watch.GlanceVerb{{Label: "Go now", Post: "/api/journey/go"}, {Label: "Hold", Post: "/api/journey/hold"}}
		cur, st := 3, "open"
		next := &watch.GlanceNextStep{ID: "ch3", Title: "Blood Raven", Band: "too early (floor 9)", Text: "Level 7 is under the floor of 9."}
		switch variant {
		case "ch1":
			cur, next = 1, &watch.GlanceNextStep{ID: "ch1", Title: "Den of Evil", Band: "ready", Text: "The hero is in range."}
		case "ch9":
			cur, next = 9, &watch.GlanceNextStep{ID: "ch9", Title: "Baal", Band: "far too early (floor 30)", Text: "A long way off."}
		case "blocked":
			st, next = "held", &watch.GlanceNextStep{ID: "ch3", Title: "Blood Raven", Band: "blocked", Text: "The path to the Burial Grounds is blocked."}
		}
		if variant != "nojourney" {
			v.Journey = &watch.GlanceJourney{N: cur, Of: 9, Chapters: chapters(cur, st), Next: next, Actions: verbs}
		} else {
			v.Journey = nil
		}
		light := func(l, w, d string) watch.GlanceLight { return watch.GlanceLight{Light: l, Word: w, Detail: d} }
		v.Health.Gear = light("ok", "sound", "")
		v.Health.Merc = light("ok", "Iron Wolf", "")
		v.Health.Supplies = light("ok", "stocked", "")
		switch variant {
		case "gearbroken":
			v.Health.Gear = light("bad", "broken", "The weapon is broken; a repair is about 180 gold.")
		case "unknown":
			v.Health.Merc, v.Health.Supplies = light("unknown", "unknown", ""), light("unknown", "unknown", "")
		}
		if has("mercdead") {
			v.Health.Merc = light("bad", "Dead", "The mercenary is dead and the hero fights without him")
		}
		if has("supplieslow") {
			v.Health.Supplies = light("warn", "Low", "healing 2 of 4, mana 1 of 4, portal yes")
		}
		if has("nopotions") {
			v.Health.Merc = light("warn", "None", "No mercenary is hired")
			v.Health.Supplies = light("bad", "No potions", "healing 0 of 4, mana 0 of 4, portal yes")
		}
		if has("paused") { // what glancewire_pause.go gives for the going-home state
			v.Sentence.Tone, v.Sentence.Word, v.Sentence.Reason = "info", "Paused", "paused by you, going home first"
		}
		js(w, v)
	}
}

// fixtureNeeds are the first n of the rows the accepted mockup shows and a few more, worst first.
func fixtureNeeds(n int) []watch.GlanceNeed {
	all := []watch.GlanceNeed{
		{ID: "gear", Kind: "gear", Sev: "danger", Text: "Weapon broken", Sub: "repair is about 180 gold", Actions: []watch.GlanceVerb{{Label: "Repair now", Post: "/api/repair", Body: json.RawMessage(`{"why":"gear"}`)}}},
		{ID: "raven", Kind: "proposal", Sev: "warn", Text: "Go for Blood Raven at level 9? The level guide says so", Actions: []watch.GlanceVerb{{Label: "Go now", Post: "/api/orders"}, {Label: "Not yet", Post: "/api/journey/hold"}}},
		{ID: "belt", Kind: "belt", Sev: "info", Text: "Belt is empty", Sub: "buy potions in town", Actions: []watch.GlanceVerb{{Label: "Buy", Post: "/api/buy"}}},
		{ID: "merc", Kind: "merc", Sev: "warn", Text: "Mercenary is dead", Actions: []watch.GlanceVerb{{Label: "Revive", Post: "/api/merc/revive"}}},
		{ID: "stash", Kind: "stash", Sev: "info", Text: "Stash is full"},
		{ID: "key", Kind: "k", Sev: "info", Text: "No keys left"},
		{ID: "rare", Kind: "k", Sev: "info", Text: "A rare ring is waiting in the stash"},
		{ID: "town", Kind: "k", Sev: "info", Text: "Town portal scrolls are low"},
		{ID: "gold", Kind: "k", Sev: "info", Text: "Gold is above the limit"},
	}
	if n < 0 {
		n = 0
	}
	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}
