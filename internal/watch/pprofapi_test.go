package watch

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProfilerRoutes(t *testing.T) {
	p := &Profiler{Dir: t.TempDir()}
	h := p.Handler(AnswerOptions{Token: "tok", Origins: []string{"http://ok"}})
	do := func(method, path, tok, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		if tok != "" {
			r.Header.Set("Authorization", "Bearer "+tok)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := do("POST", "/api/debug/cpuprofile?secs=2", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", w.Code)
	}
	if w := do("POST", "/api/debug/cpuprofile?secs=2", "tok", "http://evil"); w.Code != http.StatusForbidden {
		t.Fatalf("bad origin: %d", w.Code)
	}
	if w := do("POST", "/api/debug/cpuprofile?secs=x", "tok", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("bad secs: %d", w.Code)
	}
	if w := do("GET", "/api/debug/cpuprofile/file", "tok", ""); w.Code != http.StatusNotFound {
		t.Fatalf("file before any profile: %d", w.Code)
	}
	if w := do("POST", "/api/debug/cpuprofile?secs=999", "tok", ""); w.Code != http.StatusAccepted {
		t.Fatalf("start: %d %s", w.Code, w.Body)
	}
	if w := do("POST", "/api/debug/cpuprofile", "tok", ""); w.Code != http.StatusConflict {
		t.Fatalf("second start: %d", w.Code)
	}
	if w := do("GET", "/api/debug/cpuprofile/file", "tok", ""); w.Code != http.StatusConflict {
		t.Fatalf("file while running: %d", w.Code)
	}
	if left := time.Until(p.until); left > MaxProfileSecs*time.Second+time.Second {
		t.Fatalf("cap not applied: %v", left)
	}
	end := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(end) { // give the profiler something to sample
	}
	if w := do("POST", "/api/debug/cpuprofile/stop", "tok", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"running":false`) {
		t.Fatalf("stop: %d %s", w.Code, w.Body)
	}
	w := do("GET", "/api/debug/cpuprofile/file", "tok", "")
	if w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Fatalf("file: %d len %d", w.Code, w.Body.Len())
	}
	if fi, err := os.Stat(p.path); err != nil || fi.Size() == 0 {
		t.Fatalf("stored file: %v", err)
	}
}
