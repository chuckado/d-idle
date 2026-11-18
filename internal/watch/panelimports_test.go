package watch

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"
)

var reImport = regexp.MustCompile(`(?m)^\s*(?:import|export)\b[^'"\n]*?from\s+'(\.{1,2}/[^']+)'|^\s*import\s+'(\.{1,2}/[^']+)'`)

// A panel module is loaded by the browser with import('/watch/panels/<id>/panel.js'); every module it imports must be served by
// the same handler, or the shell shows "This panel failed to load." (video/brightness.js was embedded but answered 404).
func TestEveryPanelImportIsServed(t *testing.T) {
	r := newRig(t, "sorceress")
	h := r.sim.Hub.ViewHandler()
	checked := 0
	err := fs.WalkDir(webFS, "web/panels", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Base(p) != "panel.js" {
			return err
		}
		queue := []string{p}
		seen := map[string]bool{p: true}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			b, err := fs.ReadFile(webFS, cur)
			if err != nil {
				t.Errorf("%s: %v", cur, err)
				continue
			}
			for _, m := range reImport.FindAllStringSubmatch(string(b), -1) {
				spec := m[1] + m[2]
				dep := path.Join(path.Dir(cur), spec)
				if seen[dep] {
					continue
				}
				seen[dep] = true
				url := "/" + strings.Replace(dep, "web/", "watch/", 1)
				w := httptest.NewRecorder()
				h.ServeHTTP(w, httptest.NewRequest("GET", url, nil))
				checked++
				if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/javascript") {
					t.Errorf("%s imports %s: GET %s gave %d %q", cur, spec, url, w.Code, w.Header().Get("Content-Type"))
				}
				queue = append(queue, dep)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 4 {
		t.Fatalf("only %d imports checked: the scan lost the video panel's imports", checked)
	}
}

// The must-fail side: a module that is not there, or a name that is not a plain module name, is still a 404.
func TestPanelModuleRouteRefusesOthers(t *testing.T) {
	r := newRig(t, "sorceress")
	h := r.sim.Hub.ViewHandler()
	for _, u := range []string{"/watch/panels/video/nope.js", "/watch/panels/video/.js", "/watch/panels/video/a.b.js", "/watch/panels/video/panel.json", "/watch/panels/video/..%2fpanel.js", "/watch/panels/video/x/brightness.js"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", u, nil))
		if w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently {
			t.Errorf("GET %s gave %d, want 404", u, w.Code)
		}
	}
}
