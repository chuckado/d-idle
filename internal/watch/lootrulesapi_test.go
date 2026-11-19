package watch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/lootrules"
)

type lrDesk struct{ s *lootrules.Store }

func (d lrDesk) View() any {
	return d.s.View(1, []lootrules.Recent{{Scope: "ground", Item: "El Rune", Today: "pickup", Rules: "pickup", Rule: "g-runes", Why: "Will pick up: rune"}})
}
func (d lrDesk) Put(b []byte) (any, error) { return d.s.Put(b) }

func lrReq(h http.Handler, method, body string, hdr map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/api/lootrules", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestLootRulesAPIGetPutGates(t *testing.T) {
	auth := map[string]string{"Authorization": "Bearer tok"}
	opts := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	api := NewLootRulesAPI()
	if w := lrReq(api.PutHandler(opts), "PUT", `{}`, auth); w.Code != 503 {
		t.Errorf("unbound PUT %d", w.Code)
	}
	if w := lrReq(api.Handler(), "GET", ``, nil); w.Code != 503 {
		t.Errorf("unbound GET %d", w.Code)
	}
	for _, class := range []string{"Amazon", "Sorceress"} {
		st := lootrules.OpenStore(filepath.Join(t.TempDir(), "lr.json"), []string{"Amazon", "Sorceress"})
		api := NewLootRulesAPI()
		api.Bind(lrDesk{st})
		put, get := api.PutHandler(opts), api.Handler()

		w := lrReq(get, "GET", "", nil)
		var v lootrules.ViewDoc
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &v) != nil || v.Schema != lootrules.ViewSchema || v.Seq != 1 || len(v.Recent) != 1 || v.Plain["g-runes"] == "" || len(v.Presets) != 3 {
			t.Fatalf("GET %d %s", w.Code, w.Body)
		}
		if lrReq(get, "PUT", `{}`, auth).Code != http.StatusMethodNotAllowed {
			t.Error("PUT on the viewer listener")
		}
		// gates
		if w := lrReq(put, "PUT", `{}`, nil); w.Code != 401 {
			t.Errorf("no token %d", w.Code)
		}
		if w := lrReq(put, "PUT", `{}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}); w.Code != 403 {
			t.Errorf("bad origin %d", w.Code)
		}
		if w := lrReq(put, "POST", `{}`, auth); w.Code != 405 {
			t.Errorf("POST %d", w.Code)
		}
		if w := lrReq(put, "PUT", `{}`, map[string]string{"Authorization": "Bearer tok", "Content-Type": "text/plain"}); w.Code != 415 {
			t.Errorf("content type %d", w.Code)
		}
		if w := lrReq(put, "PUT", strings.Repeat("x", int(lootRulesMaxBody)+10), auth); w.Code != 413 {
			t.Errorf("big body %d", w.Code)
		}
		// a preset for the class with options
		ok := `{"schema":"lootrules/1","seq":1,"name":"x","limits":{"maxActions":1},"preset":"class:` + class + `","options":{"minRune":7},"rules":[]}`
		w = lrReq(put, "PUT", ok, auth)
		var rep lootRulesReply
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &rep) != nil || !rep.OK || rep.Seq != 2 {
			t.Fatalf("PUT %d %s", w.Code, w.Body)
		}
		// the same base again is a conflict that names the stored seq
		if w = lrReq(put, "PUT", ok, auth); w.Code != 409 || !strings.Contains(w.Body.String(), `"seq":2`) {
			t.Errorf("conflict %d %s", w.Code, w.Body)
		}
		// field and syntax errors are 400 with a path or a line
		w = lrReq(put, "PUT", `{"schema":"lootrules/1","seq":2,"name":"x","limits":{"maxActions":1},"rules":[{"id":"a","action":"explode","reason":"r"}]}`, auth)
		if w.Code != 400 || !strings.Contains(w.Body.String(), `"path":"rules[0].action"`) {
			t.Errorf("field error %d %s", w.Code, w.Body)
		}
		w = lrReq(put, "PUT", "{\n\"schema\": \n}", auth)
		if w.Code != 400 || !strings.Contains(w.Body.String(), `"line":3`) {
			t.Errorf("syntax error %d %s", w.Code, w.Body)
		}
		if st.Active().Seq != 2 {
			t.Error("refused bodies changed the store")
		}
	}
}
