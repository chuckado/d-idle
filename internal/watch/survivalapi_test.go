package watch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSurvivalRoute(t *testing.T) {
	h := &Hub{}
	get := func() string {
		rr := httptest.NewRecorder()
		h.survivalHandler()(rr, httptest.NewRequest(http.MethodGet, SurvivalRoute, nil))
		return rr.Body.String()
	}
	if got := get(); !strings.Contains(got, `"enabled":false`) {
		t.Fatalf("no source: %s", got)
	}
	h.SetSurvivalSource(func() any { return map[string]string{"margin": "safe"} })
	if got := get(); !strings.Contains(got, `"enabled":true`) || !strings.Contains(got, `"margin":"safe"`) {
		t.Fatalf("source: %s", got)
	}
	h.SetSurvivalSource(nil)
	if got := get(); !strings.Contains(got, `"enabled":false`) {
		t.Fatalf("cleared: %s", got)
	}
	rr := httptest.NewRecorder()
	h.survivalHandler()(rr, httptest.NewRequest(http.MethodPost, SurvivalRoute, nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST: %d", rr.Code)
	}
}
