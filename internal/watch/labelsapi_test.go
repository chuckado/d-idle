package watch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeLabels struct {
	on   bool
	sets []bool
	err  string
}

func (f *fakeLabels) Get() LabelsState {
	return LabelsState{OK: f.err == "", On: f.on, Available: f.err != "unsupported", Code: f.err}
}
func (f *fakeLabels) Set(on bool) LabelsState {
	f.sets = append(f.sets, on)
	if f.err == "" {
		f.on = on
	}
	return f.Get()
}

func TestLabelsRouteGates(t *testing.T) {
	api := NewLabelsAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	post, get := api.PostHandler(o), api.Handler()
	auth := map[string]string{"Authorization": "Bearer tok"}

	if w := invPost(post, LabelsRoute, `{"on":true}`, auth); w.Code != 503 {
		t.Errorf("POST unbound: %d", w.Code)
	}
	w := httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("GET", LabelsRoute, nil))
	if w.Code != 503 {
		t.Errorf("GET unbound: %d", w.Code)
	}
	f := &fakeLabels{}
	api.Bind(f)
	w = invPost(post, LabelsRoute, `{"on":true}`, auth)
	var got LabelsState
	if err := json.Unmarshal(w.Body.Bytes(), &got); w.Code != 200 || err != nil || !got.OK || !got.On || !got.Available {
		t.Errorf("set on: %d %s", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("GET", LabelsRoute, nil))
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || !got.On {
		t.Errorf("GET reads the state back: %s", w.Body)
	}
	if w := invPost(post, LabelsRoute, `{"on":false}`, auth); w.Code != 200 || len(f.sets) != 2 || f.on {
		t.Errorf("set off: %d %s", w.Code, w.Body)
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
		if w := invPost(post, LabelsRoute, c.body, c.hdr); w.Code != c.want {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	if len(f.sets) != 2 {
		t.Errorf("a refused request reached the desk: %v", f.sets)
	}
	w = httptest.NewRecorder()
	get.ServeHTTP(w, httptest.NewRequest("POST", LabelsRoute, nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("viewer accepts POST: %d", w.Code)
	}
}
