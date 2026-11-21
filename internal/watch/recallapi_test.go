package watch

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

type fakeRecall struct {
	calls []string
	err   string
}

func (f *fakeRecall) body(state string) any {
	return map[string]any{"schema": "watch.recall/1", "enabled": f.err == "", "state": state, "in_town": false, "error": f.err}
}
func (f *fakeRecall) View() any   { return f.body("idle") }
func (f *fakeRecall) Recall() any { f.calls = append(f.calls, "recall"); return f.body("waiting") }
func (f *fakeRecall) Cancel() any { f.calls = append(f.calls, "cancel"); return f.body("idle") }

func TestRecallRoute(t *testing.T) {
	api := NewRecallAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	post, get := api.PostHandler(o), api.Handler()
	auth := map[string]string{"Authorization": "Bearer tok"}
	if w := invPost(post, RecallRoute, `{}`, auth); w.Code != 503 {
		t.Errorf("POST unbound: %d", w.Code)
	}
	w := httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("GET", RecallRoute, nil))
	if w.Code != 503 {
		t.Errorf("GET unbound: %d", w.Code)
	}
	f := &fakeRecall{}
	api.Bind(f)
	w = invPost(post, RecallRoute, `{}`, auth)
	var m map[string]any
	if json.Unmarshal(w.Body.Bytes(), &m) != nil || w.Code != 200 || m["ok"] != true || m["state"] != "waiting" {
		t.Fatalf("recall: %d %s", w.Code, w.Body)
	}
	if w = invPost(post, RecallRoute, `{"cancel":true}`, auth); w.Code != 200 || len(f.calls) != 2 || f.calls[1] != "cancel" {
		t.Errorf("cancel: %d %v", w.Code, f.calls)
	}
	if w = invPost(post, RecallRoute, `{}`, nil); w.Code == 200 {
		t.Errorf("POST without the token answered 200")
	}
	f.err = "immediate management is off (knob ui.managementFast)"
	w = invPost(post, RecallRoute, `{}`, auth)
	m = nil
	if json.Unmarshal(w.Body.Bytes(), &m) != nil || m["ok"] != false || m["code"] != "recall_refused" || m["error"] == "" {
		t.Errorf("refusal: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("GET", RecallRoute, nil))
	if w.Code != 200 {
		t.Errorf("GET: %d", w.Code)
	}
}
