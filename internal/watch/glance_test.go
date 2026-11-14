package watch

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/telemetry"
)

// glanceClass is the hero of a golden: nothing here knows a class, the two builds only differ in this data.
var glanceClass = map[string]HeroInfo{
	"sorceress": {Name: "Simmy", Class: "Sorceress", Level: 7, Difficulty: "normal", Area: "Cold Plains", AreaID: 3, Activity: "Fighting in Cold Plains",
		Life: 86, MaxLife: 120, Mana: 105, MaxMana: 140, XP: 15000, XPNext: 18000},
	"paladin": {Name: "Pally", Class: "Paladin", Level: 12, Difficulty: "normal", Area: "Black Marsh", AreaID: 6, Activity: "Fighting in Black Marsh",
		Life: 200, MaxLife: 260, Mana: 41, MaxMana: 60, XP: 90000, XPNext: 120000},
}

// halfway puts a hero half way through his level, from the game's experience table.
func halfway(h HeroInfo) HeroInfo {
	lo, _ := telemetry.DefaultExpForLevel(h.Level)
	hi, _ := telemetry.DefaultExpForLevel(h.Level + 1)
	h.XP, h.XPNext = int64(lo+(hi-lo)/2), 0
	return h
}

func glanceRig(t *testing.T, build, scenario string) *rig {
	r := newRig(t, build)
	h := r.sim.Hub
	hero := halfway(glanceClass[build])
	h.SetGoldSource(func() (int, int, bool) { return 1240, 5300, true })
	h.SetHeroStatsSource(func() (HeroStats, bool) {
		return HeroStats{Resists: HeroResists{Fire: 12}, Readiness: &HeroReadiness{Light: "amber", Area: hero.Area, Reason: "Lightning resistance is under the line."}}, true
	})
	h.Whyline().Observe(goals.Why{Code: "fight", Params: map[string]string{"area": hero.Area, "life": "72"}}, hero.Activity, "g1", "")
	var src GlanceSources
	switch scenario {
	case "daily":
		h.Activity(KindProgress, "fighting")
		early := hero
		early.XP -= 600 // the baseline of xp/h, three minutes ago
		h.SetHero(early)
		r.sim.Clock.Advance(3 * time.Minute)
		if _, err := h.GiveOrder(OrderGo, hero.Area, 30); err != nil {
			t.Fatal(err)
		}
	case "needs":
		h.Activity(KindProgress, "fighting")
		src.Needs = func() []GlanceNeed {
			return []GlanceNeed{
				{ID: "gear", Kind: "gear", Sev: "danger", Text: "Weapon broken", Sub: "repair is about 180 gold",
					Actions: []GlanceVerb{{Label: "Repair now", Post: "/api/repair", Body: json.RawMessage(`{"now":true}`)}}},
				{ID: "proposal", Kind: "proposal", Sev: "warn", Text: "Go for Blood Raven at level 9?",
					Actions: []GlanceVerb{{Label: "Go now", Post: "/api/orders"}, {Label: "Not yet", Post: "/api/orders"}, {Label: "x", Post: "https://elsewhere/"}}},
			}
		}
	case "blocked":
		h.Activity(KindStall, "town: the quest path is blocked")
		src.Needs = func() []GlanceNeed {
			return []GlanceNeed{{ID: "journey", Kind: "journey_blocked", Sev: "warn", Text: "The quest path is blocked", Sub: "no mercenary yet"}}
		}
		src.Journey = func() (GlanceJourney, bool) {
			return GlanceJourney{N: 3, Of: 9, Chapters: []GlanceChapter{{ID: "a", State: "done"}, {ID: "b", State: "done"}, {ID: "c", State: "open"}},
				Next: &GlanceNextStep{ID: "c", Title: "Blood Raven", Band: "too_early", Text: "floor 9"}}, true
		}
	case "paused":
		h.Activity(KindIdle, "paused")
		src.Paused = func() (bool, string) { return true, "walking home first, then stopping" }
	case "dead":
		hero.Life = 0
		h.Activity(KindDeath, "died in "+hero.Area)
	case "unknown":
		hero = HeroInfo{}
		h.SetHeroStatsSource(nil)
	}
	h.SetHero(hero)
	h.SetGlanceSources(src)
	return r
}

// The handler answers for the six situations and both builds; the golden files are the contract the panels build on.
func TestGlanceGoldens(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} {
		for _, sc := range []string{"daily", "needs", "blocked", "paused", "dead", "unknown"} {
			r := glanceRig(t, build, sc)
			w := r.get(GlanceRoute, nil)
			if w.Code != 200 {
				t.Fatalf("%s %s: %d", build, sc, w.Code)
			}
			var buf strings.Builder
			var v any
			json.Unmarshal(w.Body.Bytes(), &v)
			b, _ := json.MarshalIndent(v, "", "  ")
			buf.Write(b)
			buf.WriteByte('\n')
			golden(t, "glance/view-"+sc+"-"+build, buf.String())
		}
	}
}

// Whatever the bot says, the band has a reason: every code of whychange.json gives a non-empty sentence, and so does a
// hero that has nothing to say yet.
func TestGlanceReasonForEveryWhyCode(t *testing.T) {
	r := newRig(t, "sorceress")
	if len(whyChange()) == 0 {
		t.Fatal("whychange.json is empty")
	}
	tab := (&Whyline{}).tab
	if err := json.Unmarshal(whylineJSON, &tab); err != nil {
		t.Fatal(err)
	}
	for code := range whyChange() {
		rc := newRig(t, "sorceress")
		params := map[string]string{}
		for _, p := range tab.Codes[code].Params {
			params[p] = "x"
		}
		rc.sim.Hub.SetHero(glanceClass["sorceress"])
		rc.sim.Hub.Whyline().Observe(goals.Why{Code: code, Params: params}, "", "g1", "")
		v := rc.sim.Hub.Glance()
		if strings.TrimSpace(v.Sentence.Reason) == "" {
			t.Errorf("code %s: empty reason", code)
		}
		if v.Sentence.Why.Code != code || v.Sentence.Why.Now == "" {
			t.Errorf("code %s: why %+v", code, v.Sentence.Why)
		}
	}
	if v := r.sim.Hub.Glance(); v.Sentence.Reason == "" || v.Sentence.Word == "" {
		t.Errorf("nothing known: %+v", v.Sentence)
	}
}

// Text from the bot is plain and bounded; an action may only post to an API route; the margin shows as words.
func TestGlanceTextAndVerbs(t *testing.T) {
	r := glanceRig(t, "sorceress", "needs")
	long := "<b>" + strings.Repeat("x", 400)
	r.sim.Hub.SetGlanceSources(GlanceSources{
		Needs: func() []GlanceNeed {
			n := make([]GlanceNeed, 30)
			for i := range n {
				n[i] = GlanceNeed{ID: "n", Sev: "weird", Text: long, Actions: []GlanceVerb{{Label: long, Post: "/api/x"}, {Label: "b", Post: "javascript:1"}}}
			}
			return n
		},
		Voice: func() string { return "hi\n" + long },
	})
	v := r.sim.Hub.Glance()
	if len(v.Needs) != glanceMaxNeed {
		t.Errorf("needs %d", len(v.Needs))
	}
	n := v.Needs[0]
	if len(n.Text) > glanceTextMax || n.Sev != "info" || len(n.Actions) != 1 || len(n.Actions[0].Label) > 24 {
		t.Errorf("need not cleaned: %+v", n)
	}
	if len(v.Sentence.Voice) > glanceTextMax || strings.Contains(v.Sentence.Voice, "\n") {
		t.Errorf("voice %q", v.Sentence.Voice)
	}
	if v.Sentence.Margin != "risky" {
		t.Errorf("margin %q", v.Sentence.Margin)
	}
	if b, _ := json.Marshal(v); strings.Contains(string(b), `"margin":1`) {
		t.Error("a number for the margin")
	}
}

// The xp/h and gold/h series fill one point a minute while somebody polls, and a reset empties them.
func TestGlanceSeries(t *testing.T) {
	r := glanceRig(t, "paladin", "daily")
	h := r.sim.Hub
	hero := glanceClass["paladin"]
	gold := 1000
	h.SetGoldSource(func() (int, int, bool) { return gold, 0, true })
	for i := 0; i < 4; i++ {
		if i > 0 {
			r.sim.Clock.Advance(time.Minute)
		}
		hero.XP += 600
		gold += 60
		h.SetHero(hero)
		h.Glance()
	}
	v := h.Glance()
	if n := len(v.Progress.XPSeries); n != 3 {
		t.Fatalf("series %v", v.Progress.XPSeries)
	}
	for _, x := range v.Progress.XPSeries {
		if x != 36000 {
			t.Errorf("rate %v", v.Progress.XPSeries)
		}
	}
	if v.Progress.GoldPerHour == nil || *v.Progress.GoldPerHour != 3600 {
		t.Errorf("gold/h %v", v.Progress.GoldPerHour)
	}
	h.Glance() // same minute: no new point
	if n := len(h.Glance().Progress.XPSeries); n != 3 {
		t.Errorf("a second sample in one minute: %d", n)
	}
	hero.XP = 10 // another character
	h.SetHero(hero)
	if n := len(h.Glance().Progress.XPSeries); n != 0 {
		t.Errorf("series after a reset: %d", n)
	}
	for i := 0; i < 40; i++ {
		r.sim.Clock.Advance(time.Minute)
		hero.XP += 100
		h.SetHero(hero)
		h.Glance()
	}
	if n := len(h.Glance().Progress.XPSeries); n != glancePoints-1 {
		t.Errorf("ring not bounded: %d", n)
	}
}

// DOM goldens of the band panels in node: statusline (ok, warn, danger, info, loading, error, a long reason), needs-count
// (0, 1, 3, error) and the slim form of hero-meters.
func TestGlancePanelDOMGoldens(t *testing.T) {
	runNode(t, "testdata/glance_dom.js", "web/panels")
}

// The band panels keep the page rules: no timers, requests or queries of their own, text through textContent, no style
// writes, and the free-grid layouts never mount them.
func TestGlancePanelsFollowTheContract(t *testing.T) {
	bad := regexp.MustCompile(`setInterval|setTimeout|fetch\(|EventSource|document\.|innerHTML|\.style\b|window\.`)
	for _, id := range []string{"statusline", "needs-count"} {
		b, err := webFS.ReadFile("web/panels/" + id + "/panel.js")
		if err != nil {
			t.Fatal(err)
		}
		if m := bad.Find(b); m != nil {
			t.Errorf("%s: %q breaks the panel contract", id, m)
		}
		mb, _ := webFS.ReadFile("web/panels/" + id + "/panel.json")
		var m PanelManifest
		if err := json.Unmarshal(mb, &m); err != nil {
			t.Fatal(err)
		}
		if !m.OptIn || !m.Slot || !m.CockpitOnly || m.Endpoint != GlanceRoute {
			t.Errorf("%s manifest: %+v", id, m)
		}
	}
}
