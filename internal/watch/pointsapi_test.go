package watch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/points"
)

// pointsRig is the real points desk over a tiny two-class table and a game that takes every verb.
type pointsRig struct {
	t        *testing.T
	p        d2.Player
	desk     *points.Desk
	post     http.Handler
	view     http.Handler
	statSent []int
}

func newPointsRig(t *testing.T, class d2.Class) *pointsRig {
	r := &pointsRig{t: t, p: d2.Player{Class: class, Level: 10, Strength: 20, Vitality: 20, StatPoints: 3, SkillPoints: 1}}
	tb := points.Tables{
		Classes: []*data.CharClass{{Name: "Sorceress", Code: "sor", Str: 10, LifePerVit: 8, ManaPerEnergy: 8}, {Name: "Paladin", Code: "pal", Str: 25, LifePerVit: 12, ManaPerEnergy: 6}},
		Skills: []*data.Skill{
			{ID: 36, Key: "Fire Bolt", Name: "Fire Bolt", Class: "sor", Page: 1, ReqLevel: 1, MaxLevel: 20},
			{ID: 96, Key: "Might", Name: "Might", Class: "pal", Page: 1, ReqLevel: 1, MaxLevel: 20, Aura: true},
		},
	}
	r.desk = points.NewDesk(points.Config{
		Hero: "Testy", T: tb, Dir: t.TempDir(),
		Builds: func() []points.BuildInfo {
			return []points.BuildInfo{{Name: "sor_a", Class: "sorceress", Tier: points.BuildSupported}, {Name: "pal_a", Class: "paladin", Tier: points.BuildSupported}}
		},
		Player: func() *d2.Player { q := r.p; return &q },
		Fresh:  func(context.Context) (*d2.Player, error) { q := r.p; return &q, nil },
		Verbs:  func() points.Verbs { return points.Verbs{Stat: true, Skill: true, Via: "d2d verbs"} },
		SpendStat: func(_ context.Context, id int) error {
			r.statSent = append(r.statSent, id)
			r.p.StatPoints--
			r.p.Vitality++
			return nil
		},
		SpendSkill: func(_ context.Context, idx int) error {
			r.p.SkillPoints--
			id := map[d2.Class]int{d2.Sorceress: 36, d2.Paladin: 96}[r.p.Class]
			r.p.Skills = append(r.p.Skills, d2.Skill{ID: id, Level: 1})
			return nil
		},
	})
	api := NewPointsAPI()
	api.Bind(r.desk)
	r.post, r.view = api.PostHandler(AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}), api.Handler()
	return r
}

func (r *pointsRig) do(path, body string) *httptest.ResponseRecorder {
	return invPost(r.post, path, body, questAuth)
}

func (r *pointsRig) get() (int, string) {
	w := httptest.NewRecorder()
	r.view.ServeHTTP(w, httptest.NewRequest("GET", PointsRoute, nil))
	return w.Code, w.Body.String()
}

func TestPointsAPITwoClasses(t *testing.T) {
	for _, c := range []struct {
		class   d2.Class
		skill   int
		preview string
	}{{d2.Sorceress, 36, "+1 Vitality: +2 life"}, {d2.Paladin, 96, "+1 Vitality: +3 life"}} {
		r := newPointsRig(t, c.class)
		code, body := r.get()
		if code != 200 || !strings.Contains(body, c.preview) || !strings.Contains(body, `"waiting":"3 stat points and 1 skill point waiting for you"`) || !strings.Contains(body, `"autopilot":false`) {
			t.Fatalf("%s GET %d %s", c.class, code, body)
		}
		if w := r.do("/api/hero/points/stat", `{"stat":"vit"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) || r.statSent[0] != 3 {
			t.Fatalf("stat: %d %s sent %v", w.Code, w.Body, r.statSent)
		}
		if w := r.do("/api/hero/points/skill", `{"skill":9999}`); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "unknown_skill") {
			t.Errorf("illegal skill: %d %s", w.Code, w.Body)
		}
		if w := r.do("/api/hero/points/skill", `{"skill":`+map[int]string{36: "36", 96: "96"}[c.skill]+`}`); w.Code != 200 {
			t.Errorf("legal skill: %d %s", w.Code, w.Body)
		}
		if w := r.do("/api/hero/points/autopilot", `{"on":true}`); w.Code != 200 || !r.desk.Autopilot() {
			t.Errorf("autopilot: %d %s", w.Code, w.Body)
		}
		// the audit trail names the client class and the action, never an address
		ev := r.desk.Recent(10)
		if len(ev) != 4 || ev[3].Action != "stat" || ev[3].Target != "vit" || ev[3].Remote == "" || ev[0].Action != "autopilot" {
			t.Errorf("audit %+v", ev)
		}
	}
}

func TestPointsAPIGate(t *testing.T) {
	r := newPointsRig(t, d2.Paladin)
	req := httptest.NewRequest("POST", "/api/hero/points/stat", strings.NewReader(`{"stat":"str"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.post.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized || len(r.statSent) != 0 {
		t.Errorf("no token: %d sent %v", w.Code, r.statSent)
	}
	if w := r.do("/api/hero/points/nothing", `{}`); w.Code != http.StatusNotFound {
		t.Errorf("route: %d", w.Code)
	}
	if w := r.do("/api/hero/points/stat", `{}`); w.Code != http.StatusBadRequest {
		t.Errorf("empty body: %d", w.Code)
	}
	if w := r.do("/api/hero/points/stat", `{"stat":"vit","extra":1}`); w.Code != http.StatusBadRequest {
		t.Errorf("unknown field: %d", w.Code)
	}
	gw := httptest.NewRecorder()
	r.view.ServeHTTP(gw, httptest.NewRequest("POST", PointsRoute, strings.NewReader("{}")))
	if gw.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST on the viewer route: %d", gw.Code)
	}
	if w := NewPointsAPI(); w.desk() != nil {
		t.Error("unbound")
	}
	un := httptest.NewRecorder()
	NewPointsAPI().Handler().ServeHTTP(un, httptest.NewRequest("GET", PointsRoute, nil))
	if un.Code != http.StatusServiceUnavailable {
		t.Errorf("unbound GET: %d", un.Code)
	}
}

func TestPointsPageAndAssets(t *testing.T) {
	r := newPointsRig(t, d2.Sorceress)
	for path, want := range map[string]string{"/watch/points": "<title>d.idle - points</title>", "/watch/points.js": "/api/hero/points/stat", "/watch/points.css": ".pt-stat"} {
		w := httptest.NewRecorder()
		r.view.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), want) {
			t.Errorf("%s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	r.view.ServeHTTP(w, httptest.NewRequest("GET", "/watch/points.txt", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("unknown asset: %d", w.Code)
	}
}

func TestSkillsAllocateRoutesToThePointsDesk(t *testing.T) {
	r := newPointsRig(t, d2.Paladin)
	sk := NewSkillsAPI()
	sk.EnableAllocate(r.desk)
	w := httptest.NewRecorder()
	sk.PostHandler(AnswerOptions{Token: "tok"}).ServeHTTP(w, newJSONReq("/api/skills/allocate", `{"skill":96,"points":1}`))
	if w.Code != 200 {
		t.Fatalf("allocate: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	sk.PostHandler(AnswerOptions{Token: "tok"}).ServeHTTP(w, newJSONReq("/api/skills/allocate", `{"skill":36,"points":1}`))
	if w.Code != http.StatusConflict {
		t.Errorf("other class skill: %d %s", w.Code, w.Body)
	}
}

func newJSONReq(path, body string) *http.Request {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer tok")
	return req
}

func TestPointsReturnSection(t *testing.T) {
	r := newPointsRig(t, d2.Paladin)
	api := NewPointsAPI()
	sec := PointsSection(func() PointsDesk { return api.desk() })
	if sec.Build(&ReturnData{}) != nil {
		t.Error("unbound desk: no card")
	}
	api.Bind(r.desk)
	sv := sec.Build(&ReturnData{})
	if sv == nil || sv.ID != "points" || !strings.Contains(sv.Lines[0], "3 stat points and 1 skill point waiting for you") {
		t.Fatalf("%+v", sv)
	}
	r.p.StatPoints, r.p.SkillPoints = 0, 0
	if sec.Build(&ReturnData{}) != nil {
		t.Error("nothing banked: no card")
	}
}

func TestBuildAPITwoClasses(t *testing.T) {
	for _, c := range []struct {
		class       d2.Class
		mine, other string
	}{{d2.Sorceress, "sor_a", "pal_a"}, {d2.Paladin, "pal_a", "sor_a"}} {
		r := newPointsRig(t, c.class)
		api := NewPointsAPI()
		api.Bind(r.desk)
		post, view := api.BuildPostHandler(AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}), api.Handler()
		do := func(body string) *httptest.ResponseRecorder { return invPost(post, BuildRoute, body, questAuth) }
		w := httptest.NewRecorder()
		view.ServeHTTP(w, httptest.NewRequest("GET", BuildRoute, nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"locked":false`) {
			t.Fatalf("%s GET %d %s", c.class, w.Code, w.Body)
		}
		if w := do(`{"build":"nope"}`); w.Code != 400 || !strings.Contains(w.Body.String(), "nope") {
			t.Errorf("unknown: %d %s", w.Code, w.Body)
		}
		if w := do(`{}`); w.Code != 400 {
			t.Errorf("empty: %d", w.Code)
		}
		if w := do(`{"build":"` + c.mine + `"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) {
			t.Fatalf("lock: %d %s", w.Code, w.Body)
		}
		if w := do(`{"build":"` + c.mine + `"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"already":true`) {
			t.Errorf("idempotent: %d %s", w.Code, w.Body)
		}
		if w := do(`{"build":"` + c.other + `"}`); w.Code != 409 {
			t.Errorf("switch: %d %s", w.Code, w.Body)
		}
		if w := invPost(post, BuildRoute, `{"build":"`+c.mine+`"}`, nil); w.Code != http.StatusUnauthorized {
			t.Errorf("no token: %d", w.Code)
		}
	}
}
