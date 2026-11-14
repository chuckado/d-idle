package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The fit check (docs/ui/redesign-2026-11-14/briefs/S1b-cockpit-template.md, "The fit rule"): a cockpit page must fill
// the viewport exactly. The serve command offers /__fit?u=<page>&w=<px>&h=<px> (fitHarness), which measures the page
// (fit.js) and prints the numbers into <pre id="fit-result">; tools/uishots.sh dumps that DOM in headless Chromium at each
// size and `uishots fit DIR` judges the dumps. The page is judged as it is: the harness only reads.

//go:embed fit.js
var fitJS []byte

// fitHarness is the page /__fit serves: the page under test in an iframe of exactly w by h (so innerWidth, innerHeight, dvh and
// the container queries are those of a browser window of that size whatever the real window is: a headless Chromium
// window carries browser chrome and has a least width), and the numbers of fit.js written into <pre id="fit-result">
// once the panels have mounted and polled. The pages forbid framing (CSP frame-ancestors); framedPage relaxes that for
// the copy this viewer serves with fitframe=1 and nothing else.
func fitHarness(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	u, ww, hh := q.Get("u"), atoi(q.Get("w")), atoi(q.Get("h"))
	if !strings.HasPrefix(u, "/watch") || strings.HasPrefix(u, "//") || ww < 200 || hh < 200 || ww > 4000 || hh > 4000 {
		http.Error(w, "u must be a /watch path on this server; w and h 200 to 4000", http.StatusBadRequest)
		return
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><title>fit</title></head><body style="margin:0;background:#888">
<iframe id="f" src="%s" width="%d" height="%d" style="border:0;display:block"></iframe><pre id="fit-result"></pre>
<script type="module">
import { measure } from '/__fit/fit.js';
const f = document.getElementById('f'), out = document.getElementById('fit-result');
f.addEventListener('load', () => setTimeout(() => {
  try { out.textContent = JSON.stringify(Object.assign({ page: %q }, measure(f.contentWindow))); }
  catch (e) { out.textContent = JSON.stringify({ error: String(e) }); }
}, 5000));
</script></body></html>`, html.EscapeString(u+sep+"fitframe=1"), ww, hh, u)
}

// framedPage serves the page with the framing ban relaxed to this origin.
func framedPage(h http.Handler, w http.ResponseWriter, r *http.Request) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.Header().Del("X-Frame-Options")
	if csp := w.Header().Get("Content-Security-Policy"); csp != "" {
		w.Header().Set("Content-Security-Policy", strings.Replace(csp, "frame-ancestors 'none'", "frame-ancestors 'self'", 1))
	}
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
}

func serveFitJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Write(fitJS)
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

// fitResult is what fit.js returns.
type fitResult struct {
	Page     string `json:"page"`
	Error    string `json:"error"`
	VW       int    `json:"vw"`
	VH       int    `json:"vh"`
	ScrollH  int    `json:"scrollH"`
	ScrollW  int    `json:"scrollW"`
	Overflow []struct {
		El string `json:"el"`
		DH int    `json:"dh"`
		DW int    `json:"dw"`
	} `json:"overflow"`
	Slots []struct {
		Slot string `json:"slot"`
	} `json:"slots"`
	Covered int    `json:"covered"`
	Hole    int    `json:"hole"`
	HoleAt  [2]int `json:"holeAt"`
}

// fitGap is the widest empty strip between slots that is not dead space: the cockpit gap (8 px) and two pixels.
const fitGap = 10

// judge returns what is wrong with a measured page; an empty list is a fit.
func (r fitResult) judge() []string {
	var bad []string
	switch {
	case r.Error != "":
		return []string{"the measurement failed: " + r.Error}
	case r.VW == 0:
		return []string{"no measurement"}
	}
	if r.ScrollH > r.VH {
		bad = append(bad, fmt.Sprintf("the page scrolls: scrollHeight %d > innerHeight %d", r.ScrollH, r.VH))
	}
	if r.ScrollW > r.VW {
		bad = append(bad, fmt.Sprintf("the page scrolls sideways: scrollWidth %d > innerWidth %d", r.ScrollW, r.VW))
	}
	for _, o := range r.Overflow {
		bad = append(bad, fmt.Sprintf("%s overflows its box by %d px high, %d px wide", o.El, o.DH, o.DW))
	}
	if len(r.Slots) == 0 {
		bad = append(bad, "no [data-slot] on the page")
	}
	if r.Hole > fitGap {
		bad = append(bad, fmt.Sprintf("dead space: a %d px square is covered by no slot at %d,%d", r.Hole, r.HoleAt[0], r.HoleAt[1]))
	}
	return bad
}

var reSize = regexp.MustCompile(`-(\d+)x(\d+)$`)

var reFitResult = regexp.MustCompile(`(?s)<pre id="fit-result">(.*?)</pre>`)

// fitDir judges every fit-*.html dump of a run. A dump named fit-overflow-* is the deliberately overflowing fixture: it
// must FAIL the check, and a pass there is itself a failure (the check would not catch a real overflow). Returns the
// lines to print and whether everything is as expected.
func fitDir(dir string) ([]string, bool, error) {
	names, err := filepath.Glob(filepath.Join(dir, "fit-*.html"))
	if err != nil || len(names) == 0 {
		return nil, false, fmt.Errorf("no fit-*.html dumps in %q", dir)
	}
	sort.Strings(names)
	var lines []string
	ok := true
	for _, n := range names {
		b, err := os.ReadFile(n)
		if err != nil {
			return lines, false, err
		}
		name := strings.TrimSuffix(filepath.Base(n), ".html")
		m := reFitResult.FindSubmatch(b)
		if m == nil {
			lines = append(lines, "FAIL "+name+": no fit result in the dump")
			ok = false
			continue
		}
		if strings.TrimSpace(html.UnescapeString(string(m[1]))) == "" {
			lines = append(lines, "FAIL "+name+": the page gave no measurement in the time budget")
			ok = false
			continue
		}
		var r fitResult
		if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &r); err != nil {
			lines = append(lines, "FAIL "+name+": "+err.Error())
			ok = false
			continue
		}
		bad := r.judge()
		// the browser window must really be the size the name says, or the numbers are for another page
		if wh := reSize.FindStringSubmatch(name); wh != nil && (strconv.Itoa(r.VW) != wh[1] || strconv.Itoa(r.VH) != wh[2]) {
			bad = append(bad, fmt.Sprintf("the viewport is %dx%d, not %sx%s", r.VW, r.VH, wh[1], wh[2]))
		}
		want := strings.HasPrefix(name, "fit-overflow-")
		switch {
		case want && len(bad) == 0:
			lines = append(lines, "FAIL "+name+": the overflowing fixture passed the fit check")
			ok = false
		case want:
			lines = append(lines, fmt.Sprintf("ok   %s: failed as it must (%s)", name, bad[0]))
		case len(bad) > 0:
			lines = append(lines, "FAIL "+name+": "+strings.Join(bad, "; "))
			ok = false
		default:
			lines = append(lines, fmt.Sprintf("ok   %s: %dx%d, no scroll, %d slots, largest gap %d px", name, r.VW, r.VH, len(r.Slots), r.Hole))
		}
	}
	return lines, ok, nil
}
