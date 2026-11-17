package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/watch"
)

// The quest board fixtures (slice S5, docs/questboard.md "Watch UI"). The real Desk over the scripted Fake world gives GET
// /api/quests the shape the bot gives; the page's own query picks the variant (the poll carries it as the Referer):
//
//	quests=offer      the board on, Den of Evil and Blood Raven offered (a level 6 Paladin)
//	quests=running    Den of Evil accepted and on its first step
//	quests=awaiting   the hero waits at Akara: Hand in
//	quests=handing    awaiting, and the hand-in dialogue runs ("Handing in...")
//	quests=confirming the hand-in is read back
//	quests=failed     the hand-in failed with a reason
//	quests=blocked    the chunk is parked: Try again, Stop
//	quests=training   nothing is startable: training toward the first offer
//	quests=done       Den of Evil handed in; the next offer shows
//	quests=auto       Den of Evil is opted in to auto-start (the switch reads on)
//	quests=countdown  the opted-in Den of Evil counts down to its start by itself: Cancel
//	quests=away       what the board did while the player was away: the card and its Needs-you row
//	quests=off        the knob quests.board is 0
//
// Without the query the route does not exist (a 404), as on a bot that never bound the board. Add class=sorceress for the other build.
var (
	questMu    sync.Mutex
	questCache = map[string][]byte{}
)

func questsFixture(w http.ResponseWriter, r *http.Request) {
	q := url.Values{}
	if ref, err := url.Parse(r.Referer()); err == nil {
		q = ref.Query()
	}
	variant := q.Get("quests")
	if variant == "" {
		http.NotFound(w, r)
		return
	}
	class := "Paladin"
	if q.Get("class") == "sorceress" {
		class = "Sorceress"
	}
	k := variant + "/" + class
	questMu.Lock()
	b, ok := questCache[k]
	if !ok {
		v := questView(variant, class)
		b, _ = json.Marshal(v)
		questCache[k] = b
	}
	questMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(b)
}

func questView(variant, class string) questboard.View {
	if variant == "off" {
		return questboard.View{Schema: questboard.SchemaQuests, Enabled: false, Reason: "the quest board is off (knob quests.board is 0)", Mode: "train", Offers: []questboard.OfferView{}, Quests: []questboard.QuestRow{}, History: []questboard.HistoryRow{}, Needs: []questboard.NeedRow{}}
	}
	b, err := questboard.Load()
	if err != nil {
		panic(err)
	}
	dir, _ := os.MkdirTemp("", "quests-fixture")
	defer os.RemoveAll(dir)
	st, err := questboard.Open(b, dir, "Testy")
	if err != nil {
		panic(err)
	}
	f := questboard.NewFake(b, class, 6)
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	x := questboard.NewExec(b, st, f, f, questboard.Options{Enabled: true, Now: func() time.Time { return now }, Auto: questboard.DefaultAutoOptions()})
	d := questboard.NewDesk(b, questboard.DeskConfig{
		Exec:  func() (*questboard.Exec, *questboard.Store) { return x, st },
		Class: func() string { return class },
		Level: func() int { return f.Level },
		Now:   func() time.Time { return now },
	})
	tick := func(n int) {
		for i := 0; i < n; i++ {
			x.Tick(context.Background(), f.Step)
		}
	}
	switch variant {
	case "running":
		d.Start("den_of_evil", "")
		tick(1)
	case "awaiting", "handing", "confirming", "failed", "done":
		d.Start("den_of_evil", "")
		tick(4)
	case "blocked":
		d.Start("den_of_evil", "")
		tick(1)
	}
	switch variant {
	case "auto":
		st.SetAutoStart("den_of_evil", 0, true)
	case "countdown":
		st.SetAutoStart("den_of_evil", 0, true)
		x.NoteHold() // a player who just looked: the hold is irrelevant to the countdown picture, so lift it by waiting it out
		now = now.Add(16 * time.Minute)
		x.AutoTick(questboard.AutoInput{Calm: true})
		now = now.Add(8 * time.Second)
	case "away":
		at := now.Add(-3 * time.Hour)
		st.AddAutoEvent(questboard.AutoEvent{At: at, Quest: "den_of_evil", Chunk: "clear_den", Kind: questboard.EvAutoStart, Away: true})
		st.AddAutoEvent(questboard.AutoEvent{At: at.Add(18 * time.Minute), Quest: "den_of_evil", Chunk: "clear_den", Kind: questboard.EvAutoHandIn, Away: true})
	}
	if variant == "done" {
		f.DoneQuest["den_of_evil"] = questboard.TierMemory
	}
	now = now.Add(2 * time.Second)
	v := d.View()
	hand := func(state, reason string) {
		if a := v.Awaiting; a != nil {
			v.HandIn = &questboard.HandInView{Quest: a.Quest, Chunk: a.Chunk, State: state, Reason: reason}
		}
	}
	switch variant {
	case "handing":
		hand("running", "")
	case "confirming":
		hand("confirming", "")
	case "failed":
		hand("failed", "Akara did not answer: the dialogue never opened")
	case "blocked":
		if v.Running != nil {
			bl := *v.Running
			bl.State = questboard.Blocked
			bl.Why = "could not find the Cairn Stone"
			v.Blocked, v.Running = &bl, nil
		}
	case "training":
		for i := range v.Offers {
			v.Offers[i].Startable, v.Offers[i].Readiness = false, questboard.ReadyTooHard
			v.Offers[i].Reasons = []questboard.Reason{{Code: "boss", Text: "She would likely kill you. Train first"}}
			v.Offers[i].LevelsToGo = 2
		}
		if len(v.Offers) > 0 {
			v.Meter = &questboard.Meter{Mode: "train", Target: v.Offers[0].Quest, TargetName: v.Offers[0].QuestTitle, Word: questboard.ReadyTooHard, LevelsToGo: 2,
				Line: "Training toward " + v.Offers[0].QuestTitle + ": 2 levels to go."}
		}
	case "mid":
		v.History = []questboard.HistoryRow{{Quest: "den_of_evil", Chunk: "clear_den", Title: "Clear the Den and report to Akara", Confirmed: true}}
	}
	return v
}

// questGlance adds what the glance reply carries for the board: the flag that retires the Journey tile, the video card's strip and the
// Needs-you rows of the variant (the same shapes cmd/d2bot/questwire.go builds).
func questGlance(v *watch.GlanceView, variant string, class string) {
	if variant == "" || variant == "off" {
		return
	}
	qv := questView(variant, class)
	v.QuestBoard = true
	if s := qv.Strip; s != nil {
		v.Quest = &watch.GlanceQuest{Quest: s.Quest, Title: s.QuestTitle, Chunk: s.Chunk, Part: s.Title, State: string(s.State), Step: s.Step, NPC: s.NPC}
	}
	for _, n := range qv.Needs {
		g := watch.GlanceNeed{ID: n.ID, Kind: n.Kind, Sev: n.Sev, Text: n.Text, Sub: n.Sub, Actions: []watch.GlanceVerb{}}
		for _, a := range n.Actions {
			body, _ := json.Marshal(a.Body)
			g.Actions = append(g.Actions, watch.GlanceVerb{Label: a.Label, Post: a.Post, Body: body})
		}
		v.Needs = append(v.Needs, g)
	}
}

// Defects of the quest board (fit pages questdefect-NAME must FAIL with the verdict mustFail names): the real page with one defect put
// back. q-clip cuts the reason lines of the sheet; q-small makes the buttons 20 px (a phone's thumb); q-two shows every drawer at once;
// q-words writes a chapter counter and a margin number into the strip tile.
var questJS = map[string]string{}

func init() {
	defectCSS["q-clip"] = `[data-panel="quest-board"] .qb-r { white-space: nowrap !important; overflow: hidden !important; inline-size: 60px !important; }`
	defectCSS["q-small"] = `[data-panel^="quest-"] :is(.qb-btn, .qb-seg, .qb-sw, .qs-btn, .qs-seg) { min-block-size: 0 !important; block-size: 20px !important; }`
	defectCSS["q-two"] = `.d-drawer[hidden] { display: flex !important; }`
	questJS["q-words"] = `setInterval(function () {
  var n = document.querySelector('[data-panel="quest-strip"] .qs-sub');
  if (n && !/chapter/.test(n.textContent)) n.textContent += '  chapter 3 of 9, margin 2.4';
}, 300);`
}
