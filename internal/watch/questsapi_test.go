package watch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/questboard"
)

type questRig struct {
	t    *testing.T
	f    *questboard.Fake
	x    *questboard.Exec
	desk *questboard.Desk
	api  *QuestsAPI
	now  time.Time
	post http.Handler
	view http.Handler
}

var questAuth = map[string]string{"Authorization": "Bearer tok"}

// ledgerFake is the account wallet in memory: a key pays once.
type ledgerFake map[string]int

func (l ledgerFake) k(q, c, d string) string {
	if c == "" {
		c = "-"
	}
	return "quest/" + q + "/" + c + "/" + d
}
func (l ledgerFake) Earn(q, c, d string, n int) (questboard.LedgerResult, error) {
	k := l.k(q, c, d)
	if _, dup := l[k]; dup {
		return questboard.LedgerResult{Key: k, Already: true}, nil
	}
	l[k] = n
	return questboard.LedgerResult{Key: k, Credited: n}, nil
}
func (l ledgerFake) Paid(q, c, d string) (int, bool) { n, ok := l[l.k(q, c, d)]; return n, ok }

func newQuestRig(t *testing.T, class string, clvl int) *questRig {
	return newQuestRigWith(t, class, clvl, nil)
}

func newQuestRigWith(t *testing.T, class string, clvl int, led questboard.Ledger) *questRig {
	t.Helper()
	b, err := questboard.Load()
	if err != nil {
		t.Fatal(err)
	}
	st, err := questboard.Open(b, t.TempDir(), "Testy")
	if err != nil {
		t.Fatal(err)
	}
	f := questboard.NewFake(b, class, clvl)
	x := questboard.NewExec(b, st, f, f, questboard.Options{Enabled: true, Ledger: func() questboard.Ledger { return led }})
	r := &questRig{t: t, f: f, x: x, api: NewQuestsAPI(), now: time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)}
	r.desk = questboard.NewDesk(b, questboard.DeskConfig{
		Exec:  func() (*questboard.Exec, *questboard.Store) { return x, st },
		Class: func() string { return class },
		Now:   func() time.Time { return r.now },
	})
	r.api.Bind(r.desk)
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	r.post, r.view = r.api.PostHandler(o), r.api.Handler()
	return r
}

func (r *questRig) get() string {
	r.now = r.now.Add(2 * time.Second) // past the desk's one-second cache
	w := httptest.NewRecorder()
	r.view.ServeHTTP(w, httptest.NewRequest("GET", QuestsRoute, nil))
	if w.Code != 200 {
		r.t.Fatalf("GET: %d %s", w.Code, w.Body)
	}
	return w.Body.String()
}

func (r *questRig) do(path, body string) *httptest.ResponseRecorder {
	return invPost(r.post, path, body, questAuth)
}

func (r *questRig) tick(n int) {
	for i := 0; i < n; i++ {
		if _, err := r.x.Tick(context.Background(), r.f.Step); err != nil {
			r.t.Fatal(err)
		}
	}
}

func TestQuestsRouteGoldens(t *testing.T) {
	for _, c := range []struct{ class, name string }{{"Paladin", "paladin"}, {"Sorceress", "sorceress"}} {
		r := newQuestRig(t, c.class, 6)
		golden(t, "quests/offers-"+c.name, r.get())
		w := r.do("/api/quests/den_of_evil/start", `{}`)
		golden(t, "quests/reply-start-"+c.name, w.Body.String())
		if w.Code != 200 {
			t.Fatalf("start: %d", w.Code)
		}
		w = r.do("/api/quests/den_of_evil/start", `{}`)
		golden(t, "quests/reply-start-again-"+c.name, w.Body.String())
		w = r.do("/api/quests/search_for_cain/start", `{}`)
		golden(t, "quests/reply-busy-"+c.name, w.Body.String())
		if w.Code != 409 {
			t.Fatalf("busy: %d", w.Code)
		}
		r.tick(1)
		golden(t, "quests/running-"+c.name, r.get())
		r.tick(3)
		golden(t, "quests/awaiting-"+c.name, r.get())
	}
}

// With a wallet the offers carry `pays`, the hand-in reports `awarded` once and `already` on a repeat; without one
// neither field appears (the goldens above).
func TestQuestsRouteEtherGoldens(t *testing.T) {
	for _, c := range []struct{ class, name string }{{"Paladin", "paladin"}, {"Sorceress", "sorceress"}} {
		r := newQuestRigWith(t, c.class, 6, ledgerFake{})
		golden(t, "quests/offers-ether-"+c.name, r.get())
		r.do("/api/quests/den_of_evil/start", `{}`)
		r.tick(4)
		w := r.do("/api/quests/den_of_evil/handin", `{}`)
		if w.Code != 202 {
			t.Fatalf("handin: %d %s", w.Code, w.Body)
		}
		var done string
		for i := 0; i < 200 && done == ""; i++ {
			time.Sleep(10 * time.Millisecond)
			if g := r.get(); strings.Contains(g, `"state":"done"`) && strings.Contains(g, `"awarded"`) {
				done = g
			}
		}
		if done == "" {
			t.Fatal("hand-in never paid")
		}
		for _, want := range []string{`"awarded":[{"kind":"ether","amount":2,"key":"quest/den_of_evil/-/normal"}]`, `"paid":[{"kind":"ether","amount":2`} {
			if !strings.Contains(done, want) {
				t.Fatalf("hand-in view lacks %s", want)
			}
		}
		w = r.do("/api/quests/den_of_evil/handin", `{}`)
		golden(t, "quests/reply-handin-again-"+c.name, w.Body.String())
		if !strings.Contains(w.Body.String(), `"already":true`) || strings.Contains(w.Body.String(), "awarded") {
			t.Fatalf("repeat: %s", w.Body)
		}
	}
}

func TestQuestsRouteGate(t *testing.T) {
	r := newQuestRig(t, "Paladin", 6)
	path := "/api/quests/den_of_evil/start"
	for name, c := range map[string]struct {
		path, body string
		hdr        map[string]string
		code       int
	}{
		"no token":       {path, `{}`, nil, 401},
		"foreign origin": {path, `{}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}, 403},
		"unknown field":  {path, `{"x":1}`, questAuth, 400},
		"unknown quest":  {"/api/quests/nope/start", `{}`, questAuth, 404},
		"unknown action": {"/api/quests/den_of_evil/explode", `{}`, questAuth, 404},
		"bad mode":       {"/api/quests/mode", `{"mode":"sprint"}`, questAuth, 400},
		"no mode":        {"/api/quests/mode", `{}`, questAuth, 400},
	} {
		w := invPost(r.post, c.path, c.body, c.hdr)
		if w.Code != c.code {
			t.Errorf("%s: %d, want %d: %s", name, w.Code, c.code, w.Body)
		}
		if c.code >= 400 && c.code != 401 && c.code != 403 && !strings.Contains(w.Body.String(), `"ok":false`) && !strings.Contains(w.Body.String(), `"error"`) {
			t.Errorf("%s: no reason in %s", name, w.Body)
		}
	}
	// refusals always say ok:false with code and reason
	w := r.do("/api/quests/nope/start", `{}`)
	if !strings.Contains(w.Body.String(), `"ok":false`) || !strings.Contains(w.Body.String(), `"code":"unknown"`) || !strings.Contains(w.Body.String(), `"reason"`) {
		t.Errorf("refusal body: %s", w.Body)
	}
	if w := httptest.NewRecorder(); true {
		r.view.ServeHTTP(w, httptest.NewRequest("POST", QuestsRoute, nil))
		if w.Code != 405 {
			t.Errorf("viewer accepts POST: %d", w.Code)
		}
	}
	// not bound: 503
	empty := NewQuestsAPI()
	if w := invPost(empty.PostHandler(AnswerOptions{Token: "tok"}), path, `{}`, questAuth); w.Code != 503 {
		t.Errorf("unbound: %d", w.Code)
	}
	if w := httptest.NewRecorder(); true {
		empty.Handler().ServeHTTP(w, httptest.NewRequest("GET", QuestsRoute, nil))
		if w.Code != 503 {
			t.Errorf("unbound GET: %d", w.Code)
		}
	}
}

func TestQuestsHandInIsAsyncAndIdempotent(t *testing.T) {
	r := newQuestRig(t, "Sorceress", 6)
	r.do("/api/quests/den_of_evil/start", `{}`)
	if w := r.do("/api/quests/den_of_evil/handin", `{}`); w.Code != 409 || !strings.Contains(w.Body.String(), `"code":"not_awaiting"`) {
		t.Fatalf("early hand-in: %d %s", w.Code, w.Body)
	}
	r.tick(4)
	w := r.do("/api/quests/den_of_evil/handin", `{}`)
	if w.Code != 202 || !strings.Contains(w.Body.String(), `"pending":true`) || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatalf("hand-in: %d %s", w.Code, w.Body)
	}
	done := false
	for i := 0; i < 300 && !done; i++ {
		done = strings.Contains(r.get(), `"handIn":{"quest":"den_of_evil","chunk":"clear_den","state":"done"`)
		time.Sleep(5 * time.Millisecond)
	}
	if !done {
		t.Fatalf("hand-in never reported done: %s", r.get())
	}
	w = r.do("/api/quests/den_of_evil/handin", `{}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"already":true`) {
		t.Fatalf("repeat hand-in: %d %s", w.Code, w.Body)
	}
	if r.f.Count("handin") != 1 {
		t.Fatalf("dialogue ran %d times", r.f.Count("handin"))
	}
}

func TestGlanceCarriesTheQuestStrip(t *testing.T) {
	r := newQuestRig(t, "Paladin", 6)
	r.do("/api/quests/den_of_evil/start", `{}`)
	h := glanceRig(t, "paladin", "daily").sim.Hub
	h.SetGlanceSources(GlanceSources{Quest: func() (GlanceQuest, bool) {
		s := r.desk.View().Strip
		if s == nil {
			return GlanceQuest{}, false
		}
		return GlanceQuest{Quest: s.Quest, Title: s.QuestTitle, Chunk: s.Chunk, Part: s.Title, State: string(s.State), Step: s.Step, NPC: s.NPC}, true
	}})
	v := h.Glance()
	if v.Quest == nil || v.Quest.Quest != "den_of_evil" || v.Quest.State == "" {
		t.Fatalf("glance quest: %+v", v.Quest)
	}
}

// Slice S8: the opt-in route and the away card dismissal, for two builds.
func TestQuestsAutoRoute(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		r := newQuestRig(t, class, 6)
		w := r.do("/api/quests/den_of_evil/auto", `{"on":true}`)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) || strings.Contains(w.Body.String(), `"already"`) {
			t.Fatalf("%s on: %d %s", class, w.Code, w.Body)
		}
		if !strings.Contains(r.get(), `"id":"den_of_evil","title":"Den of Evil","act":1,"state":"ready","auto":true`) {
			t.Fatalf("%s: the row does not show the flag: %s", class, r.get())
		}
		if w = r.do("/api/quests/den_of_evil/auto", `{"on":true}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"already":true`) {
			t.Fatalf("%s again: %d %s", class, w.Code, w.Body)
		}
		if w = r.do("/api/quests/den_of_evil/auto", `{}`); w.Code != 400 {
			t.Fatalf("%s no body: %d %s", class, w.Code, w.Body)
		}
		w = r.do("/api/quests/search_for_cain/auto", `{"on":true}`)
		if w.Code != 409 || !strings.Contains(w.Body.String(), `"code":"auto_never"`) || !strings.Contains(w.Body.String(), "always needs you") {
			t.Fatalf("%s never: %d %s", class, w.Code, w.Body)
		}
		if w = r.do("/api/quests/den_of_evil/auto", `{"on":false}`); w.Code != 200 || strings.Contains(r.get(), `"auto":true`) {
			t.Fatalf("%s off: %d %s", class, w.Code, w.Body)
		}
		if w = r.do("/api/quests/away/dismiss", `{}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) {
			t.Fatalf("%s dismiss: %d %s", class, w.Code, w.Body)
		}
		if w = r.do("/api/quests/den_of_evil/bogus", `{}`); w.Code != 404 {
			t.Fatalf("%s bogus: %d", class, w.Code)
		}
	}
}
