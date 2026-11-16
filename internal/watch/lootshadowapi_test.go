package watch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLootShadowAPI(t *testing.T) {
	a := NewLootShadowAPI()
	h := a.Handler()
	get := func(method string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, "/api/lootrules/shadow", nil))
		return w
	}
	if w := get("GET"); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("unbound: %d", w.Code)
	}
	a.Bind(func() any { return map[string]any{"schema": "watch.lootshadow/1", "decisions": 3} })
	w := get("GET")
	var got map[string]any
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil || got["schema"] != "watch.lootshadow/1" {
		t.Fatalf("bound: %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Error("cache header")
	}
	if w := get("POST"); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST must be refused: %d", w.Code)
	}
}
