package watch

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeReset mirrors the shape of bot.ResetView (watch.reset/1) without importing the bot.
type fakeReset struct {
	state, reason string
	last          map[string]any
	refuse        *ResetRefusal
	sets          [][2]bool
}

func (f *fakeReset) body() map[string]any {
	m := map[string]any{"schema": "watch.reset/1", "state": f.state}
	if f.state != "idle" {
		m["since"], m["reason"] = "2026-11-15T10:00:00Z", f.reason
	}
	if f.last != nil {
		m["last"] = f.last
	}
	return m
}
func (f *fakeReset) View() any { return f.body() }
func (f *fakeReset) Set(hard, force bool) (any, *ResetRefusal) {
	f.sets = append(f.sets, [2]bool{hard, force})
	if f.refuse != nil {
		return nil, f.refuse
	}
	m := f.body()
	if f.state == "resetting" || f.state == "rejoining" {
		m["already"] = true // a double press
		return m, nil
	}
	f.state, f.reason = "resetting", "waiting for a safe moment to leave the game"
	return f.body(), nil
}

func TestResetRouteGates(t *testing.T) {
	api := NewResetAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	post, get := api.PostHandler(o), api.Handler()
	auth := map[string]string{"Authorization": "Bearer tok"}

	if w := invPost(post, ResetRoute, `{}`, auth); w.Code != 503 {
		t.Errorf("POST unbound: %d", w.Code)
	}
	w := httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("GET", ResetRoute, nil))
	if w.Code != 503 {
		t.Errorf("GET unbound: %d", w.Code)
	}
	f := &fakeReset{state: "idle"}
	api.Bind(f)
	if w := invPost(post, ResetRoute, `{"hard":false}`, auth); w.Code != 200 || len(f.sets) != 1 || f.sets[0] != [2]bool{false, false} {
		t.Errorf("soft: %d %s %v", w.Code, w.Body, f.sets)
	}
	// the second press reaches the desk, which says it is already running; the route answers 200 either way
	if w := invPost(post, ResetRoute, `{"hard":true,"force":true}`, auth); w.Code != 200 || !strings.Contains(w.Body.String(), `"already":true`) {
		t.Errorf("double press: %d %s", w.Code, w.Body)
	}
	f.refuse = &ResetRefusal{Code: "measured_run", Error: "a measured run never resets"}
	if w := invPost(post, ResetRoute, `{}`, auth); w.Code != 409 || !strings.Contains(w.Body.String(), `"code":"measured_run"`) {
		t.Errorf("refusal: %d %s", w.Code, w.Body)
	}
	n := len(f.sets)
	for name, c := range map[string]struct {
		body string
		hdr  map[string]string
		want int
	}{
		"no token":      {`{}`, nil, 401},
		"foreign":       {`{}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}, 403},
		"unknown field": {`{"x":1}`, auth, 400},
		"wrong type":    {`{"hard":"yes"}`, auth, 400},
	} {
		if w := invPost(post, ResetRoute, c.body, c.hdr); w.Code != c.want {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	if len(f.sets) != n {
		t.Errorf("a refused request reached the desk: %v", f.sets)
	}
	w = httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("POST", ResetRoute, nil))
	if w.Code != 405 {
		t.Errorf("viewer accepts POST: %d", w.Code)
	}
}

// Goldens: one per state; fixtures are only added, never edited (UPDATE_GOLDEN=1 writes missing ones).
func TestResetRouteGoldens(t *testing.T) {
	done := map[string]any{"result": "done", "at": "2026-11-15T09:00:00Z", "reason": "left the game and rejoined"}
	for _, st := range []struct {
		name, state, reason string
		last                map[string]any
	}{
		{"idle", "idle", "", nil},
		{"idle-after-done", "idle", "", done},
		{"resetting", "resetting", "waiting for a safe moment to leave the game", nil},
		{"rejoining", "rejoining", "leaving the game and rejoining", nil},
		{"failed", "failed", "the game did not come back", map[string]any{"result": "failed", "at": "2026-11-15T10:01:00Z", "reason": "the game did not come back"}},
	} {
		api := NewResetAPI()
		api.Bind(&fakeReset{state: st.state, reason: st.reason, last: st.last})
		w := httptest.NewRecorder()
		api.Handler().ServeHTTP(w, httptest.NewRequest("GET", ResetRoute, nil))
		path := filepath.Join("testdata", "reset", "view-"+st.name+".json")
		if _, err := os.Stat(path); err != nil && os.Getenv("UPDATE_GOLDEN") != "" {
			os.WriteFile(path, w.Body.Bytes(), 0o644)
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(want) != w.Body.String() {
			t.Errorf("%s:\nwant %s\ngot  %s", st.name, want, w.Body)
		}
	}
	// the refusal body
	api := NewResetAPI()
	api.Bind(&fakeReset{state: "idle", refuse: &ResetRefusal{Code: "measured_run", Error: "a measured run never resets"}})
	w := invPost(api.PostHandler(AnswerOptions{Token: "tok"}), ResetRoute, `{}`, map[string]string{"Authorization": "Bearer tok"})
	path := filepath.Join("testdata", "reset", "post-refused-measured_run.json")
	if _, err := os.Stat(path); err != nil && os.Getenv("UPDATE_GOLDEN") != "" {
		os.WriteFile(path, w.Body.Bytes(), 0o644)
	}
	if want, err := os.ReadFile(path); err != nil || string(want) != w.Body.String() {
		t.Errorf("refusal golden: %v\nwant %s\ngot  %s", err, want, w.Body)
	}
}
