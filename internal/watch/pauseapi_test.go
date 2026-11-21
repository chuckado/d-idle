package watch

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// fakePause mirrors the shape of bot.PlayerPauseView (watch.pause/1) without importing the bot.
type fakePause struct {
	state, reason string
	sets          []bool
}

func (f *fakePause) body() any {
	m := map[string]any{"schema": "watch.pause/1", "state": f.state}
	if f.reason != "" {
		m["since"] = "2026-11-14T10:00:00Z"
		m["reason"] = f.reason
	}
	return m
}
func (f *fakePause) View() any { return f.body() }
func (f *fakePause) Set(on bool) any {
	f.sets = append(f.sets, on)
	if on {
		f.state, f.reason = "going_home", "waiting for a safe moment to go home"
	} else {
		f.state, f.reason = "running", ""
	}
	return f.body()
}

func TestPauseRouteGates(t *testing.T) {
	api := NewPauseAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	post, get := api.PostHandler(o), api.Handler()
	auth := map[string]string{"Authorization": "Bearer tok"}

	if w := invPost(post, PauseRoute, `{"on":true}`, auth); w.Code != 503 {
		t.Errorf("POST unbound: %d", w.Code)
	}
	w := httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("GET", PauseRoute, nil))
	if w.Code != 503 {
		t.Errorf("GET unbound: %d", w.Code)
	}
	f := &fakePause{state: "running"}
	api.Bind(f)
	if w := invPost(post, PauseRoute, `{"on":true}`, auth); w.Code != 200 || len(f.sets) != 1 || !f.sets[0] {
		t.Errorf("pause: %d %s", w.Code, w.Body)
	}
	if w := invPost(post, PauseRoute, `{"on":false}`, auth); w.Code != 200 || len(f.sets) != 2 || f.sets[1] {
		t.Errorf("resume: %d %s", w.Code, w.Body)
	}
	for name, c := range map[string]struct {
		body string
		hdr  map[string]string
		want int
	}{
		"no token":      {`{"on":true}`, nil, 401},
		"foreign":       {`{"on":true}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}, 403},
		"unknown field": {`{"on":true,"x":1}`, auth, 400},
		"missing on":    {`{}`, auth, 400},
	} {
		if w := invPost(post, PauseRoute, c.body, c.hdr); w.Code != c.want {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	if len(f.sets) != 2 {
		t.Errorf("a refused request reached the desk: %v", f.sets)
	}
	w = httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("POST", PauseRoute, nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("viewer accepts POST: %d", w.Code)
	}
}

// Goldens: one per state; fixtures are only added, never edited (UPDATE_GOLDEN=1 writes missing ones).
func TestPauseRouteGoldens(t *testing.T) {
	for _, st := range []struct{ state, reason string }{
		{"running", ""}, {"going_home", "going home"}, {"paused", "paused in town"},
	} {
		api := NewPauseAPI()
		api.Bind(&fakePause{state: st.state, reason: st.reason})
		w := httptest.NewRecorder()
		api.Handler().ServeHTTP(w, httptest.NewRequest("GET", PauseRoute, nil))
		path := filepath.Join("testdata", "pause", "view-"+st.state+".json")
		if _, err := os.Stat(path); err != nil && os.Getenv("UPDATE_GOLDEN") != "" {
			os.WriteFile(path, w.Body.Bytes(), 0o644)
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(want) != w.Body.String() {
			t.Errorf("%s:\nwant %s\ngot  %s", st.state, want, w.Body)
		}
	}
}

// modePause is a desk that also takes the mode.
type modePause struct {
	fakePause
	modes []string
}

func (m *modePause) SetMode(on bool, freeze *bool) any {
	m.modes = append(m.modes, map[bool]string{true: "on", false: "off"}[on]+":"+map[bool]string{true: "freeze", false: "walk"}[*freeze])
	return m.fakePause.Set(on)
}

// {"on":true,"freeze":true|false} reaches a desk that takes a mode; a desk without it (or no freeze field) gets Set.
func TestPauseRouteFreezeField(t *testing.T) {
	api := NewPauseAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	post := api.PostHandler(o)
	auth := map[string]string{"Authorization": "Bearer tok"}
	m := &modePause{fakePause: fakePause{state: "running"}}
	api.Bind(m)
	for _, body := range []string{`{"on":true,"freeze":true}`, `{"on":true,"freeze":false}`, `{"on":false}`} {
		if w := invPost(post, PauseRoute, body, auth); w.Code != 200 {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body)
		}
	}
	if len(m.modes) != 2 || m.modes[0] != "on:freeze" || m.modes[1] != "on:walk" || len(m.sets) != 3 {
		t.Errorf("modes %v sets %v", m.modes, m.sets)
	}
	if w := invPost(post, PauseRoute, `{"on":true,"freeze":"yes"}`, auth); w.Code != 400 {
		t.Errorf("a non-boolean freeze: %d", w.Code)
	}
	plain := &fakePause{state: "running"}
	api.Bind(plain)
	if w := invPost(post, PauseRoute, `{"on":true,"freeze":true}`, auth); w.Code != 200 || len(plain.sets) != 1 {
		t.Errorf("plain desk: %d %v", w.Code, plain.sets)
	}
}

// mapPause serves a fixed view (the frozen pause's fields).
type mapPause struct{ v map[string]any }

func (m mapPause) View() any    { return m.v }
func (m mapPause) Set(bool) any { return m.v }

// Goldens of the frozen pause: the fit fixtures of tools/uishots serve these two files, so a change of shape shows here.
func TestPauseRouteFrozenGoldens(t *testing.T) {
	for name, v := range map[string]map[string]any{
		"canfreeze": {"schema": "watch.pause/1", "state": "running", "can_freeze": true},
		"frozen": {"schema": "watch.pause/1", "state": "paused", "since": "2026-11-14T10:00:00Z", "reason": "the game is frozen", "frozen": true,
			"can_freeze": true, "cap_min": 60, "ends_at": "2026-11-14T11:00:00Z", "monsters_near": true},
	} {
		api := NewPauseAPI()
		api.Bind(mapPause{v})
		w := httptest.NewRecorder()
		api.Handler().ServeHTTP(w, httptest.NewRequest("GET", PauseRoute, nil))
		path := filepath.Join("testdata", "pause", "view-"+name+".json")
		if _, err := os.Stat(path); err != nil && os.Getenv("UPDATE_GOLDEN") != "" {
			os.WriteFile(path, w.Body.Bytes(), 0o644)
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(want) != w.Body.String() {
			t.Errorf("%s:\nwant %s\ngot  %s", name, want, w.Body)
		}
	}
}
