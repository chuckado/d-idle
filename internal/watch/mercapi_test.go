package watch

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/merc"
)

// mercRig drives the real request object through the routes: the adapter is the cmd/d2bot one in miniature.
type mercRig struct {
	p  *merc.PlayerRevive
	in merc.Inputs
}

func (m *mercRig) View() any             { return m.p.View(m.in) }
func (m *mercRig) Request() (any, error) { return m.p.Request(m.in) }
func (m *mercRig) Cancel() error         { return m.p.Cancel() }

func newMercRig(class string, level, gold int, st merc.MercState) *mercRig {
	t := time.Unix(1_800_000_000, 0)
	return &mercRig{p: merc.NewPlayerRevive(func() time.Time { return t }), in: merc.Inputs{State: st, Ctx: merc.Ctx{
		Gold: gold, Kit: 300, Level: level, Act: 2, PlayerClass: class, HaveTP: true, HavePotions: true, Knobs: merc.DefaultKnobs()}}}
}

func mercGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if os.Getenv("UPDATE") != "" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != got {
		t.Errorf("%s differs\n got: %s\nwant: %s", path, got, want)
	}
}

func TestMercRoutesGatesAndGoldens(t *testing.T) {
	auth := map[string]string{"Authorization": "Bearer tok"}
	opts := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}

	// 503 off until bound, on both listeners
	api := NewMercAPI()
	if w := invPost(api.PostHandler(opts), "/api/merc/revive", `{}`, auth); w.Code != 503 || !strings.Contains(w.Body.String(), `"off"`) {
		t.Errorf("unbound POST: %d %s", w.Code, w.Body)
	}
	w := httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/merc", nil))
	if w.Code != 503 {
		t.Errorf("unbound GET: %d", w.Code)
	}
	if w := invPost(api.Handler(), "/api/merc", `{}`, auth); w.Code == 200 {
		t.Errorf("POST on the viewer listener answered %d", w.Code)
	}

	dead := merc.MercState{Known: true, Owned: true, Level: 10} // price 750
	for _, class := range []string{"Paladin", "Sorceress"} {
		rig := newMercRig(class, 10, 750, dead)
		api := NewMercAPI()
		api.Bind(rig)
		post := api.PostHandler(opts)
		get := func() string {
			w := httptest.NewRecorder()
			api.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/merc", nil))
			if w.Code != 200 {
				t.Fatalf("GET %d %s", w.Code, w.Body)
			}
			return w.Body.String()
		}
		mercGolden(t, "merc_view_idle_"+class+".json", get())
		if w := invPost(post, "/api/merc/revive", `{}`, nil); w.Code != 401 {
			t.Errorf("no token: %d", w.Code)
		}
		if w := invPost(post, "/api/merc/revive", `{"x":1}`, auth); w.Code != 400 {
			t.Errorf("unknown field: %d", w.Code)
		}
		if w := invPost(post, "/api/merc/revive", `{}`, auth); w.Code != 200 {
			t.Fatalf("revive: %d %s", w.Code, w.Body)
		}
		mercGolden(t, "merc_view_queued_"+class+".json", get())
		if w := invPost(post, "/api/merc/revive", `{}`, auth); w.Code != 409 || !strings.Contains(w.Body.String(), `"queued"`) {
			t.Errorf("second revive: %d %s", w.Code, w.Body)
		}
		if w := invPost(post, "/api/merc/cancel", `{}`, auth); w.Code != 200 {
			t.Errorf("cancel: %d %s", w.Code, w.Body)
		}
		if w := invPost(post, "/api/merc/cancel", `{}`, auth); w.Code != 409 {
			t.Errorf("cancel twice: %d", w.Code)
		}

		short := newMercRig(class, 10, 700, dead)
		api2 := NewMercAPI()
		api2.Bind(short)
		w := invPost(api2.PostHandler(opts), "/api/merc/revive", `{}`, auth)
		if w.Code != 409 {
			t.Fatalf("short of gold: %d %s", w.Code, w.Body)
		}
		mercGolden(t, "merc_revive_short_"+class+".json", w.Body.String())
		w = httptest.NewRecorder()
		api2.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/merc", nil))
		mercGolden(t, "merc_view_short_"+class+".json", w.Body.String())

		alive := newMercRig(class, 10, 5000, merc.MercState{Known: true, Owned: true, Alive: true, Level: 10})
		api3 := NewMercAPI()
		api3.Bind(alive)
		w = httptest.NewRecorder()
		api3.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/merc", nil))
		mercGolden(t, "merc_view_alive_"+class+".json", w.Body.String())
	}
	for _, p := range append(append([]string{}, MercRoutes...), MercPostRoutes...) {
		if !strings.HasPrefix(p, "/api/merc") {
			t.Errorf("odd route %s", p)
		}
	}
}
