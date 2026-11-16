package watch

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"d2lod-bot/internal/decide"
)

func TestDropsFeedAndItemsTab(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} { // the feed is build-independent
		now := func() float64 { return 100 }
		l := NewAwayLog(now)
		l.Add(AwayEvent{T: 90, Kind: AwayDrop, Item: "Short Staff", Quality: "normal", Area: "Blood Moor"})
		l.Add(AwayEvent{T: 95, Kind: AwayKill, Monster: "Fallen"})
		l.Add(AwayEvent{T: 98, Kind: AwayDrop, Item: "Stormshield", Quality: "unique"})
		w := httptest.NewRecorder()
		DropsHandler(l, now).ServeHTTP(w, httptest.NewRequest("GET", DropsRoute, nil))
		var v DropsView
		if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || len(v.Rows) != 2 || v.Rows[0].Item != "Stormshield" || v.Rows[0].Ago != 2 {
			t.Fatalf("%s: %v %s", build, err, w.Body.String())
		}
		var lv LogView
		logFill(&lv, "items", w.Body.Bytes(), 100)
		if len(lv.Rows) != 2 || lv.Rows[0].Q != "unique" || !strings.HasPrefix(lv.Rows[0].Text, "picked up Stormshield") {
			t.Errorf("%s: %+v", build, lv.Rows)
		}
	}
}

func TestFirstCardIsInformation(t *testing.T) {
	r := newRig(t, "sorceress")
	info := decide.ThoughtInfo{Point: decide.FirstsPoint, Key: "t0.first|act=1,first=corpsefire", Sit: decide.Situation{
		"question": "Corpsefire is a unique. Kite him?", "yes": "kite him", "no": "stand and fight"}}
	c := r.sim.Hub.BuildCard(info, "yes")
	if !c.Specific || !strings.Contains(c.Headline, "Corpsefire") || strings.Contains(c.Headline, "in the situation") {
		t.Errorf("headline: %+v", c)
	}
	if !strings.Contains(c.IfNothing, "Nothing is required") || !strings.Contains(c.IfNothing, "kite him") {
		t.Errorf("if nothing: %q", c.IfNothing)
	}
	if fl := firstLabels(info); fl["no"] != "stand and fight" {
		t.Errorf("labels: %v", fl)
	}
}
