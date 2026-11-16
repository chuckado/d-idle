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

type lrToolsDesk struct {
	lrDesk
	previewBody []byte
}

func (d *lrToolsDesk) Preview(b []byte) (any, error) {
	d.previewBody = b
	return lootrules.PreviewDoc{Schema: lootrules.PreviewSchema, Name: "x", Items: []lootrules.PreviewRow{}, Totals: map[string]int{"sold": 1}}, nil
}
func (d *lrToolsDesk) Undo(seq int) (any, error)             { return d.s.Undo(seq) }
func (d *lrToolsDesk) Import(b []byte, seq int) (any, error) { return d.s.Import(b, seq) }
func (d *lrToolsDesk) Export() ([]byte, error)               { return d.s.Export() }

func lrPath(h http.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestLootRulesToolsRoutes(t *testing.T) {
	auth := map[string]string{"Authorization": "Bearer tok"}
	opts := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	st := lootrules.OpenStore(filepath.Join(t.TempDir(), "lr.json"), []string{"Amazon", "Sorceress"})
	d := &lrToolsDesk{lrDesk: lrDesk{st}}
	api := NewLootRulesAPI()
	if w := lrPath(api.PostHandler(opts), "POST", "/api/lootrules/undo", `{"seq":1}`, auth); w.Code != 503 {
		t.Errorf("unbound %d", w.Code)
	}
	api.Bind(d)
	post, get, put := api.PostHandler(opts), api.Handler(), api.PutHandler(opts)

	// the gates are the PUT gates
	if w := lrPath(post, "POST", "/api/lootrules/preview", `{}`, nil); w.Code != 401 {
		t.Errorf("no token %d", w.Code)
	}
	if w := lrPath(post, "POST", "/api/lootrules/preview", `{}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}); w.Code != 403 {
		t.Errorf("origin %d", w.Code)
	}
	if w := lrPath(post, "PUT", "/api/lootrules/preview", `{}`, auth); w.Code != 405 {
		t.Errorf("method %d", w.Code)
	}
	if w := lrPath(post, "POST", "/api/lootrules/import?seq=1", strings.Repeat("x", int(lootRulesMaxBody)+5), auth); w.Code != 413 {
		t.Errorf("size cap %d", w.Code)
	}

	// preview: the body reaches the desk, nothing is saved
	w := lrPath(post, "POST", "/api/lootrules/preview", `{"preset":"starter"}`, auth)
	var pd lootrules.PreviewDoc
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &pd) != nil || pd.Schema != lootrules.PreviewSchema || string(d.previewBody) != `{"preset":"starter"}` || st.Active().Seq != 1 {
		t.Fatalf("preview %d %s", w.Code, w.Body)
	}

	// export is a download of the active set; it round-trips through import
	w = lrPath(get, "GET", "/api/lootrules/export", "", nil)
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") || !strings.Contains(w.Body.String(), `"schema": "lootrules/1"`) {
		t.Fatalf("export %d %v", w.Code, w.Header())
	}
	file := w.Body.String()
	for _, p := range []string{"token", "password", "secret", "/home/"} {
		if strings.Contains(strings.ToLower(file), p) {
			t.Errorf("export mentions %q", p)
		}
	}
	if w := lrPath(post, "POST", "/api/lootrules/import", file, auth); w.Code != 400 {
		t.Errorf("import without a version %d", w.Code)
	}
	w = lrPath(post, "POST", "/api/lootrules/import?seq=1", file, auth)
	var rep lootRulesReply
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &rep) != nil || !rep.OK || rep.Seq != 2 {
		t.Fatalf("import %d %s", w.Code, w.Body)
	}
	if w := lrPath(post, "POST", "/api/lootrules/import?seq=1", file, auth); w.Code != 409 {
		t.Errorf("stale import %d", w.Code)
	}
	if w := lrPath(post, "POST", "/api/lootrules/import?seq=2", `{"schema":"lootrules/1","junk":1}`, auth); w.Code != 400 || !strings.Contains(w.Body.String(), `"errors"`) {
		t.Errorf("bad file %d %s", w.Code, w.Body)
	}

	// undo: needs a previous version, then goes back
	if w := lrPath(post, "POST", "/api/lootrules/undo", `{"seq":2}`, auth); w.Code != 200 {
		t.Errorf("undo %d %s", w.Code, w.Body)
	}
	if w := lrPath(post, "POST", "/api/lootrules/undo", `{"seq":2}`, auth); w.Code != 409 {
		t.Errorf("stale undo %d", w.Code)
	}
	if w := lrPath(post, "POST", "/api/lootrules/undo", `nope`, auth); w.Code != 400 {
		t.Errorf("bad undo body %d", w.Code)
	}
	_ = put
	// a desk without the tools answers 503 on writes and 404 on export
	api2 := NewLootRulesAPI()
	api2.Bind(lrDesk{st})
	if w := lrPath(api2.PostHandler(opts), "POST", "/api/lootrules/undo", `{"seq":1}`, auth); w.Code != 503 {
		t.Errorf("no tools write %d", w.Code)
	}
	if w := lrPath(api2.Handler(), "GET", "/api/lootrules/export", "", nil); w.Code != 404 {
		t.Errorf("no tools export %d", w.Code)
	}
}
