package watch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlanAPIRoute(t *testing.T) {
	a := NewPlanAPI()
	h := a.Handler()
	get := func(method, path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, path, nil))
		return w
	}
	if w := get("GET", PlanRoute); w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), `"off"`) {
		t.Fatalf("unbound: %d %s", w.Code, w.Body)
	}
	a.Bind(func() any { return map[string]any{"enabled": true, "line": "your build: Zeal 4"} })
	w := get("GET", PlanRoute)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "your build: Zeal 4") {
		t.Fatalf("bound: %d %s", w.Code, w.Body)
	}
	if w := get("POST", PlanRoute); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("the plan route is read-only: POST answered %d", w.Code)
	}
	if w := get("GET", "/api/hero/plan/x"); w.Code != http.StatusNotFound {
		t.Errorf("other paths: %d", w.Code)
	}
}
