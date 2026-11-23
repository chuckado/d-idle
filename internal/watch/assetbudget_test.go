package watch

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The viewer's asset budget (card uibudget): the JS and CSS a page loads, counted as gzip bytes, with a ceiling a little above today's
// number, so growth shows up in review instead of in a slow phone. The set is what the page's HTML links, the scripts those import, and for
// the cockpit the panels its layout and shell.json name. Numbers are in docs/ui-design-system.md "Asset budget"; raise a ceiling in the
// same change that adds the weight and say why.
func assetSet(t *testing.T, page string, extra []string) (map[string]int, int) {
	src, err := webFS.ReadFile("web/" + page)
	if err != nil {
		t.Fatal(err)
	}
	link := regexp.MustCompile(`(?:src|href)="/watch/([A-Za-z0-9._-]+\.(?:js|css))"`)
	imp := regexp.MustCompile(`(?:from|import)\s*\(?\s*['"](?:/watch/|\./)([A-Za-z0-9._/-]+\.js)['"]`)
	seen := map[string]int{}
	var add func(name string)
	add = func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		b, err := webFS.ReadFile("web/" + name)
		if err != nil {
			return // a route served by Go code (not a file): not counted
		}
		var z bytes.Buffer
		w, _ := gzip.NewWriterLevel(&z, 6)
		w.Write(b)
		w.Close()
		seen[name] = z.Len()
		if strings.HasSuffix(name, ".js") {
			for _, m := range imp.FindAllStringSubmatch(string(b), -1) {
				add(m[1])
			}
		}
	}
	for _, m := range link.FindAllStringSubmatch(string(src), -1) {
		add(m[1])
	}
	for _, e := range extra {
		add(e)
	}
	total := 0
	for _, n := range seen {
		total += n
	}
	return seen, total
}

func cockpitPanelFiles(t *testing.T) []string {
	ids := map[string]bool{}
	var walk func(v any)
	walk = func(v any) {
		switch x := v.(type) {
		case string:
			ids[x] = true
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	for _, f := range []string{"web/layouts/cockpit.json", "web/shell.json"} {
		b, err := webFS.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var v any
		if err := json.Unmarshal(b, &v); err != nil {
			t.Fatal(err)
		}
		walk(v)
	}
	var out []string
	for id := range ids {
		for _, ext := range []string{"panel.js", "panel.css"} {
			if _, err := webFS.ReadFile("web/panels/" + id + "/" + ext); err == nil {
				out = append(out, "panels/"+id+"/"+ext)
			}
		}
	}
	sort.Strings(out)
	return out
}

func TestAssetBudget(t *testing.T) {
	cases := []struct {
		name, page string
		extra      func() []string
		ceiling    int // gzip bytes
	}{
		{"cockpit (/watch)", "watch.html", func() []string { return cockpitPanelFiles(t) }, 0},
		{"vault (/watch/vault)", "vault.html", func() []string { return nil }, 0},
	}
	for i := range cases {
		c := &cases[i]
		files, total := assetSet(t, c.page, c.extra())
		t.Logf("%s: %d files, %d gzip bytes", c.name, len(files), total)
		c.ceiling = budgetCeiling[c.page]
		if total > c.ceiling {
			t.Errorf("%s loads %d gzip bytes of JS and CSS, over the budget of %d: raise the budget in assetbudget_test.go and docs/ui-design-system.md with the reason, or cut weight", c.name, total, c.ceiling)
		}
		if total < c.ceiling*7/10 {
			t.Errorf("%s loads %d gzip bytes, far under the budget %d: lower the budget so growth stays visible", c.name, total, c.ceiling)
		}
	}
}

// budgetCeiling is today's measured gzip total plus about 8 percent (docs/ui-design-system.md "Asset budget").
var budgetCeiling = map[string]int{"watch.html": 193000, "vault.html": 69000}
