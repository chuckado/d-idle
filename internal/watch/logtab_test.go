package watch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// logRig serves the four source routes of /api/log from fixtures whose times are relative to the hub's own clock, so the
// handler's "seconds ago" is exact whatever the build.
func logRig(t *testing.T, build string) (*Hub, http.Handler) {
	t.Helper()
	s, err := NewSim(build, 1)
	if err != nil {
		t.Fatal(err)
	}
	h := s.Hub
	now := h.rel(h.cfg.Clock.Now())
	m := http.NewServeMux()
	reply := func(w http.ResponseWriter, v any) { writeJSON(w, 200, v) }
	m.HandleFunc("/api/thoughts", func(w http.ResponseWriter, r *http.Request) {
		reply(w, ThoughtsView{Schema: thoughtsSchema, Now: now, Thoughts: []ThoughtItem{{ID: 1, Kind: "question", Question: "Go for Blood Raven?"}},
			Feed: []FeedItem{{ID: "f1", T: now - 30, Kind: "story", Text: "A quiet moor."}, {ID: "f2", T: now - 5, Kind: "order", Text: "Order changed."}}})
	})
	m.HandleFunc("/api/journal", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "80" {
			t.Errorf("journal limit: %q", r.URL.RawQuery)
		}
		reply(w, JournalView{Schema: journalSchema, Now: now, Lines: []JournalLine{{ID: "j1", T: now - 90, Kind: "answer", Text: "You said yes."}, {ID: "j2", T: now - 20, Kind: "learned", Text: "It will remember."}}})
	})
	m.HandleFunc("/api/chronicle", func(w http.ResponseWriter, r *http.Request) {
		reply(w, ChronicleView{Schema: "watch.chronicle/1", Open: false, Title: "Story", Teaser: "The story opens at level 5."})
	})
	// /api/alerts is not mounted: the tab says so
	return h, h.LogHandler(m)
}

func getLog(t *testing.T, h http.Handler, q string) (int, LogView) {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", LogRoute+q, nil))
	var v LogView
	if w.Code == 200 {
		if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
			t.Fatalf("%s: %v\n%s", q, err, w.Body.String())
		}
	}
	return w.Code, v
}

func TestLogTabs(t *testing.T) {
	for _, build := range []string{"sorceress", "paladin"} {
		_, h := logRig(t, build)
		code, v := getLog(t, h, "?tab=now")
		if code != 200 || v.Schema != logSchema || v.Tab != "now" || len(v.Rows) != 2 {
			t.Fatalf("%s now: %d %+v", build, code, v)
		}
		if v.Rows[0].ID != "f2" || v.Rows[1].ID != "f1" || v.Rows[0].Ago < 4.9 || v.Rows[0].Ago > 5.1 {
			t.Errorf("%s now: newest first with seconds ago: %+v", build, v.Rows)
		}
		for _, r := range v.Rows {
			if strings.Contains(r.Text, "Blood Raven") {
				t.Errorf("an open question is a Needs you row, not a log line: %+v", r)
			}
		}
		if code, def := getLog(t, h, ""); code != 200 || def.Tab != "now" {
			t.Errorf("no tab means now: %d %q", code, def.Tab)
		}
		if _, j := getLog(t, h, "?tab=journal"); len(j.Rows) != 2 || j.Rows[0].ID != "j2" {
			t.Errorf("%s journal: %+v", build, j.Rows)
		}
		if _, s := getLog(t, h, "?tab=story"); len(s.Rows) != 0 || s.Note != "The story opens at level 5." {
			t.Errorf("%s story (locked) says why: %+v", build, s)
		}
		if _, a := getLog(t, h, "?tab=alerts"); len(a.Rows) != 0 || a.Note == "" || a.Rows == nil {
			t.Errorf("%s alerts (route absent): empty rows, a note: %+v", build, a)
		}
		if code, _ := getLog(t, h, "?tab=nope"); code != 400 {
			t.Errorf("unknown tab: %d", code)
		}
	}
}

// The log drawer's panel in a stub DOM: a golden per tab and the tab handling (testdata/logpanel/log_dom.js).
func TestLogPanelDOM(t *testing.T) {
	runNode(t, "testdata/logpanel/log_dom.js", "web/panels")
}
