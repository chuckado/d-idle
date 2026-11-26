package watch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExplainRoute(t *testing.T) {
	h := &Hub{}
	get := func() string {
		rr := httptest.NewRecorder()
		h.explainHandler()(rr, httptest.NewRequest(http.MethodGet, ExplainRoute, nil))
		return rr.Body.String()
	}
	if got := get(); !strings.Contains(got, `"enabled":false`) {
		t.Fatalf("no source: %s", got)
	}
	h.SetExplainSource(func() any { return map[string]string{"why": "x"} })
	if got := get(); !strings.Contains(got, `"enabled":true`) || !strings.Contains(got, `"why":"x"`) {
		t.Fatalf("source: %s", got)
	}
	h.SetExplainSource(nil)
	if got := get(); !strings.Contains(got, `"enabled":false`) {
		t.Fatalf("cleared: %s", got)
	}
	rr := httptest.NewRecorder()
	h.explainHandler()(rr, httptest.NewRequest(http.MethodPost, ExplainRoute, nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST: %d", rr.Code)
	}
}
