package main

import (
	"bytes"
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

//go:embed mapphone.js
var mapPhoneJS []byte

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
	// click: selectors to press, in order, before the second measurement (the "..." menu, then an entry of it; separated by
	// "|"); hit: the selectors that must be hittable afterwards (default: the menu and the Edit layout button)
	hit := q.Get("hit")
	if hit == "" {
		hit = "#menu-more,#btn-layout"
	}
	// hover and ctx: a selector to hover (pointerover, as a mouse does) or to right-click before the measurement; the tooltip or the
	// menu it opens must lie inside the window and, beside a cell, not cover it
	opts, _ := json.Marshal(map[string]any{"page": u, "click": strings.Split(q.Get("click"), "|"), "hit": strings.Split(hit, ","), "hover": q.Get("hover"), "scroll": q.Get("scroll"), "ctx": q.Get("ctx"), "expect": strings.Split(q.Get("expect"), "|"), "flow": q.Get("flow") == "1"})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><title>fit</title></head><body style="margin:0;background:#888">
<iframe id="f" src="%s" width="%d" height="%d" style="border:0;display:block"></iframe><pre id="fit-result"></pre>
<script type="module">
import { measure, hits, expects, overlayState, charsNow, vaultNow, navNow } from '/__fit/fit.js';
const pop = (d, sel, c) => {
  const n = d.querySelector(sel), r = n && !n.hidden ? n.getBoundingClientRect() : null;
  if (!r || !r.width) return { open: false, sel };
  const vw = d.defaultView.innerWidth, vh = d.defaultView.innerHeight, ov = !(r.right <= c.left || r.left >= c.right || r.bottom <= c.top || r.top >= c.bottom);
  return { open: true, sel, left: Math.round(r.left), top: Math.round(r.top), right: Math.round(r.right), bottom: Math.round(r.bottom), vw, vh, inside: r.left >= 0 && r.top >= 0 && r.right <= vw && r.bottom <= vh, covers: ov, sheet: n.dataset.sheet || n.dataset.mode === 'sheet' || '' };
};
const o = %s, f = document.getElementById('f'), out = document.getElementById('fit-result');
const wait = (ms) => new Promise((r) => setTimeout(r, ms));
// settle: a press is judged after the page has answered it, not after a guessed time. Waits until no fetch of the page is in flight and the
// DOM has been quiet for a moment (card fitfix3: the reveal, select and sheet pages flipped between runs when a re-render landed after the judge).
const settle = async (w, quiet = 300, max = 1500) => {
  if (!w.__fitBusy) { w.__fitBusy = { n: 0 }; const of = w.fetch.bind(w); w.fetch = (...a) => { w.__fitBusy.n++; return of(...a).finally(() => { w.__fitBusy.n--; }); }; }
  let since = 0;
  const mo = new w.MutationObserver(() => { since = 0; });
  mo.observe(w.document.documentElement, { subtree: true, childList: true, attributes: true, characterData: true });
  for (let t = 0; t < max; t += 50) { await wait(50); since += 50; if (w.__fitBusy.n === 0 && since >= quiet) break; }
  mo.disconnect();
};
let stage = 'before load';
// a measurement that never finishes says where it stopped (a thrown error in the page or the harness is a verdict, not a silence)
setTimeout(() => { if (!out.textContent) out.textContent = JSON.stringify({ error: 'stalled after: ' + stage }); }, 21000);
window.addEventListener('error', (e) => { if (!out.textContent) out.textContent = JSON.stringify({ error: 'script error: ' + e.message }); });
f.addEventListener('load', async () => {
  stage = 'loaded'; await wait(5000); stage = 'waited';
  // the character screen draws from its first view: measure only once its grids hold cells (a slow first answer is a late page, not a fit result)
  for (let i = 0; i < 30 && f.contentDocument.body.hasAttribute('data-canvas') && !(f.contentDocument.querySelector('#equip .d-slot2') && f.contentDocument.querySelector('#pack [data-key]') && f.contentDocument.getElementById('merc-sub').textContent && f.contentDocument.getElementById('cv-sel').firstChild && f.contentDocument.getElementById('hsheet').children.length > 4); i++) await wait(500);
  if (f.contentDocument.body.hasAttribute('data-canvas')) await wait(1500);
  // scroll: how far down (0 to 1) the Vault list is scrolled before the measurement: the rows far from the top must fit like the first ones
  if (o.scroll) { const sc = f.contentDocument.querySelector('[data-vscroll]'); if (sc) { sc.scrollTop = Number(o.scroll) * (sc.scrollHeight - sc.clientHeight); await wait(1500); } }
  try {
    stage = 'panels.json';
    const pv = await (await fetch('/watch/panels.json')).json();
    stage = 'measure';
    const m = measure(f.contentWindow, { scroll: (pv.panels || []).filter((p) => p.scroll).map((p) => p.id) });
    m.page = o.page;
    if (o.hover || o.ctx) {
      const t = f.contentDocument.querySelector(o.hover || o.ctx), r = t.getBoundingClientRect(), w = f.contentWindow;
      const ev = (type, init) => new w.MouseEvent(type, Object.assign({ bubbles: true, cancelable: true, clientX: r.left + r.width / 2, clientY: r.top + r.height / 2 }, init));
      if (o.hover) t.dispatchEvent(new w.PointerEvent('pointerover', { bubbles: true, pointerType: 'mouse', clientX: r.left + 4, clientY: r.top + 4 }));
      else t.dispatchEvent(ev('contextmenu', {}));
      await wait(600);
      m.pop = pop(f.contentDocument, o.hover ? '#inv-tip' : '#inv-act', r);
    }
    const steps = (o.click || []).filter(Boolean);
    if (steps.length) {
      m.clicked = steps.join(' then ');
      for (const raw of steps) {
        const opt = raw[0] === '?', sel = opt ? raw.slice(1) : raw, b = f.contentDocument.querySelector(sel);
        if (b) { b.click(); await wait(500); await settle(f.contentWindow); (m.trace = m.trace || []).push(sel + ' -> selecting ' + !!f.contentDocument.querySelector('[data-key=select-done]:not([hidden])') + ' sheet ' + !!f.contentDocument.querySelector('.vt-sheet:not([hidden])') + ' cells ' + f.contentDocument.querySelectorAll('.vt-cell').length + ' dp ' + !!f.contentDocument.querySelector('.vt-dcard') + ' sel ' + ((f.contentDocument.querySelector('.vt-ab-sel') || {}).textContent || '').slice(0, 30) + ' disc ' + (() => { const b = f.contentDocument.querySelector('[data-key=discard]'); return b ? (b.disabled ? 'disabled' : 'on') + (b.closest('[hidden]') ? '-hidden' : '') : 'none'; })()); } else if (!opt) { m.missing = sel; break; }
      }
    }
    m.ov = overlayState(f.contentWindow);
    if (steps.length) m.nav = navNow(f.contentWindow); // the page after its pressing: the More sheet is open now
    m.flow = !!o.flow; // a page that is a document (Skills, Points, Return ...): it scrolls on purpose
    if (steps.length) { m.hits = hits(f.contentWindow, o.hit); for (let k = 0; k < 4 && Object.values(m.hits).includes(false); k++) { await settle(f.contentWindow, 400, 800); m.hits = hits(f.contentWindow, o.hit); } } // a target that is not there yet is waited for; one that never comes is the verdict
    if (steps.length && f.contentDocument.body.hasAttribute('data-chars')) m.chars = charsNow(f.contentWindow); // the page with its sheet open
    if (steps.length && f.contentDocument.body.hasAttribute('data-vault')) m.vault = vaultNow(f.contentWindow);
    const ex = (o.expect || []).filter(Boolean);
    if (ex.length) m.expect = expects(f.contentWindow, ex);
    out.textContent = JSON.stringify(m);
  } catch (e) { out.textContent = JSON.stringify({ error: String(e) }); }
});
</script></body></html>`, html.EscapeString(u+sep+"fitframe=1"), ww, hh, strings.ReplaceAll(string(opts), "</", "<\\/"))
}

// framedPage serves the page with the framing ban relaxed to this origin.
func framedPage(h http.Handler, w http.ResponseWriter, r *http.Request) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}
	w.Header().Del("X-Frame-Options")
	body := rec.Body.Bytes()
	if d := r.URL.Query().Get("defect"); defectCSS[d] != "" || overlayJS[d] != "" || vaultJS[d] != "" || questJS[d] != "" { // a page that must fail: the same page with one defect put back, from this viewer's own file
		add := ""
		if defectCSS[d] != "" {
			add += `<link rel="stylesheet" href="/__fit/defect-` + d + `.css">`
		}
		if overlayJS[d] != "" || vaultJS[d] != "" || questJS[d] != "" {
			add += `<script src="/__fit/defect-` + d + `.js"></script>`
		}
		body = bytes.Replace(body, []byte("</head>"), []byte(add+"</head>"), 1)
		w.Header().Del("Content-Length")
	}
	if csp := w.Header().Get("Content-Security-Policy"); csp != "" {
		w.Header().Set("Content-Security-Policy", strings.Replace(csp, "frame-ancestors 'none'", "frame-ancestors 'self'", 1))
	}
	w.WriteHeader(rec.Code)
	w.Write(body)
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
	Outside []struct {
		El     string `json:"el"`
		Right  int    `json:"right"`
		Bottom int    `json:"bottom"`
	} `json:"outside"`
	Empty   []string `json:"empty"`
	Unready []string `json:"unready"`
	Pop     *struct {
		Open   bool   `json:"open"`
		Sel    string `json:"sel"`
		Inside bool   `json:"inside"`
		Covers bool   `json:"covers"`
		Sheet  any    `json:"sheet"`
	} `json:"pop"`
	Expect map[string]bool `json:"expect"`
	Flow   bool            `json:"flow"`
	Nav    *fitNav         `json:"nav"`
	Builds *fitBuilds      `json:"builds"`
	Chars  *fitChars       `json:"chars"`
	Vault  *fitVault       `json:"vault"`
	Ov     *struct {
		Open       []string `json:"open"`
		Menu       []string `json:"menu"`
		MenuScroll []string `json:"menuScroll"`
		Cut        []string `json:"cut"`
		Under      []string `json:"under"`
		Charter    []string `json:"charter"`
		Small      []string `json:"small"`
		Words      []string `json:"words"`
	} `json:"ov"`
	CV *struct {
		Zero  []string `json:"zero"`
		Cut   []string `json:"cut"`
		Tiles []string `json:"tiles"`
		Diag  string   `json:"diag"`
		Phone struct {
			ArtLabel []string `json:"artLabel"`
			Bar      []string `json:"bar"`
			Acts     []string `json:"acts"`
			Lock     []string `json:"lock"`
			Hover    []string `json:"hover"`
			Grids    []string `json:"grids"`
		} `json:"phone"`
		Inner struct {
			Size  int    `json:"size"`
			At    [2]int `json:"at"`
			Panel string `json:"panel"`
		} `json:"inner"`
	} `json:"cv"`
	Clicked string          `json:"clicked"`
	Missing string          `json:"missing"`
	Hits    map[string]bool `json:"hits"`
	Slots   []struct {
		Slot string `json:"slot"`
	} `json:"slots"`
	Zones struct {
		N       int    `json:"n"`
		Rows    int    `json:"rows"`
		MinH    int    `json:"minH"`
		Cut     int    `json:"cut"`
		Menu    bool   `json:"menu"`
		Host    string `json:"host"`
		Overlap int    `json:"overlap"`
		// SwitchOverlap is how many zone buttons the Act/Area switch lies on.
		SwitchOverlap int `json:"switchOverlap"`
	} `json:"zones"`
	// Video: the video card's controls and the controls that touch its caption or another text (card videoctl)
	Video struct {
		Controls int      `json:"controls"`
		Overlap  []string `json:"overlap"`
	} `json:"video"`
	OrderCut int    `json:"orderCut"` // order rows whose title is cut while the reason beside it still shows
	Covered  int    `json:"covered"`
	Hole     int    `json:"hole"`
	HoleAt   [2]int `json:"holeAt"`
}

// fitInner is the largest empty square inside a panel of the character screen that still reads as a layout gap (px). The first
// review's dead zones were 330 px (no mercenary) and 145 px (the selected box).
const fitInner = 135

// fitInnerPhone: the same on a phone, where a tall screen leaves the storage tab spare height under the stash (the cells are
// limited by the width, ten of them) that no layout can fill with anything but more boxes.
const fitInnerPhone = 180

// fitGap is the widest empty strip between slots that is not dead space: the cockpit gap (8 px) and two pixels.
const fitGap = 10

// broken says the measurement itself is not usable (a harness failure, not a verdict on the page): this fails a run for
// every page, including the ones that must fail.
func (r fitResult) broken() string {
	switch {
	case r.Error != "":
		return "the measurement failed: " + r.Error
	case r.VW == 0:
		return "no measurement"
	}
	return ""
}

// fitGlyphSlack is the height, in px, by which a one-line box may be shorter than its own text (line-height 1 with a taller glyph box).
const fitGlyphSlack = 2

// phoneScroll reports a cockpit page in phone mode that is taller than the window: it grows with its content and the page scrolls
// by design (web/utilities.css, "phone: the tab grows to its content"), so the vertical scroll, what lies below the window and the
// scrollbar strip are not defects. Sideways scroll, cut boxes and a page that is NOT taller than the window (its frame must be full)
// are still judged.
func (r fitResult) phoneScroll() bool {
	scrolls := r.ScrollH > r.VH || r.VW-r.ScrollW >= 8 // or a vertical scrollbar takes its strip (a page one pixel over the window)
	if r.VW >= 700 || !scrolls || r.ScrollW > r.VW {
		return false
	}
	for _, s := range r.Slots {
		if s.Slot == "phoneTabs" {
			return true
		}
	}
	return false
}

// judge returns what is wrong with a measured page; an empty list is a fit.
func (r fitResult) judge() []string {
	if r.Flow { // the navigation pages (nav-NAME): a document that scrolls on purpose; the page's own fit lane judges its content, this one the navigation
		if r.Nav == nil {
			return []string{"the page has no navigation (nav.js is not loaded)"}
		}
		bad := r.Nav.judge(r.VW)
		bad = append(bad, r.Builds.judge(r.VW, r.ScrollW)...)
		if r.Builds != nil { // the Build tab's detail page presses Details and expects Choose to be reachable
			bad = append(bad, r.clickVerdict()...)
		}
		return bad
	}
	var bad []string
	ps := r.phoneScroll()
	if r.ScrollH > r.VH && r.Vault == nil && !ps { // the Vault is one scrolling document (card vaultdesktop): only sideways counts
		bad = append(bad, fmt.Sprintf("the page scrolls: scrollHeight %d > innerHeight %d", r.ScrollH, r.VH))
	}
	if r.ScrollW > r.VW {
		bad = append(bad, fmt.Sprintf("the page scrolls sideways: scrollWidth %d > innerWidth %d", r.ScrollW, r.VW))
	}
	for _, o := range r.Overflow {
		if o.DW == 0 && o.DH > 0 && o.DH <= fitGlyphSlack {
			continue // a glyph box a pixel or two taller than a one-line box (line-height 1): nothing is hidden
		}
		bad = append(bad, fmt.Sprintf("%s overflows its box by %d px high, %d px wide", o.El, o.DH, o.DW))
	}
	for _, o := range r.Outside {
		if ps && o.Right <= r.VW {
			continue // the phone page grows with its content: below the window is the scroll, not a defect
		}
		bad = append(bad, fmt.Sprintf("%s ends outside the window (right %d, bottom %d)", o.El, o.Right, o.Bottom))
	}
	for _, e := range r.Empty {
		bad = append(bad, "slot "+e+" holds no visible panel")
	}
	for _, u := range r.Unready {
		bad = append(bad, "panel "+u+" is not ready")
	}
	if len(r.Slots) == 0 && r.Chars == nil && r.Vault == nil {
		bad = append(bad, "no [data-cockpit-slot] on the page")
	}
	if c := r.Chars; c != nil {
		bad = append(bad, c.judge(r.VW)...)
	}
	bad = append(bad, r.Nav.judge(r.VW)...)
	if v := r.Vault; v != nil {
		bad = append(bad, v.judge()...)
	}
	if r.Hole > fitGap && r.Chars == nil && r.Vault == nil && !ps { // a scrolling phone page: the hole is the scrollbar's strip
		bad = append(bad, fmt.Sprintf("dead space: a %d px square is covered by no slot at %d,%d", r.Hole, r.HoleAt[0], r.HoleAt[1]))
	}
	if r.CV != nil {
		for _, z := range r.CV.Zero {
			bad = append(bad, "content with no width: "+z)
		}
		for _, c := range r.CV.Cut {
			bad = append(bad, "a values tile is cut: "+c)
		}
		for _, t := range r.CV.Tiles {
			bad = append(bad, "a values tile is stretched: "+t)
		}
		// the phone's storage tab and the art cells (card uiphone): what the 2026-11-15 screenshot showed
		for _, p := range []struct {
			what string
			list []string
		}{{"label over art: ", r.CV.Phone.ArtLabel}, {"toolbar cut: ", r.CV.Phone.Bar}, {"action buttons: ", r.CV.Phone.Acts}, {"lock line: ", r.CV.Phone.Lock},
			{"horizontal overflow: ", r.CV.Phone.Hover}, {"grid: ", r.CV.Phone.Grids}} {
			for _, x := range p.list {
				bad = append(bad, p.what+x)
			}
		}
		lim := fitInner
		if r.VW <= 480 {
			lim = fitInnerPhone
		}
		if r.CV.Inner.Size > lim {
			bad = append(bad, fmt.Sprintf("dead space inside %s: a %d px square is empty at %d,%d", r.CV.Inner.Panel, r.CV.Inner.Size, r.CV.Inner.At[0], r.CV.Inner.At[1])+" ["+r.CV.Diag+"]")
		}
	}
	exp := make([]string, 0, len(r.Expect))
	for sel, ok := range r.Expect {
		if !ok {
			exp = append(exp, sel)
		}
	}
	sort.Strings(exp)
	for _, sel := range exp {
		bad = append(bad, "expected "+sel+" on the page: it is absent or has no size")
	}
	if z := r.Zones; z.N > 0 {
		if z.Rows < 10 {
			bad = append(bad, fmt.Sprintf("the zone map shows %d rows of its 10", z.Rows))
		}
		if z.MinH < 20 {
			bad = append(bad, fmt.Sprintf("a zone row is %d px high (least 20)", z.MinH))
		}
		if z.Cut > 0 {
			bad = append(bad, fmt.Sprintf("%d zone names have no room (under 24 px wide)", z.Cut))
		}
		if z.Overlap > 0 && !(z.Host == "self" && r.VW <= 480) { // the phone's sheet replaces the map (J: a zone tap swaps the sheet)
			bad = append(bad, fmt.Sprintf("the zone menu (in %s) covers %d zone buttons", z.Host, z.Overlap))
		}
	}
	if r.Zones.SwitchOverlap > 0 {
		bad = append(bad, fmt.Sprintf("the map switch covers %d zone buttons", r.Zones.SwitchOverlap))
	}
	if len(r.Video.Overlap) > 0 {
		bad = append(bad, "video control covers text: "+strings.Join(r.Video.Overlap, "; "))
	}
	if r.OrderCut > 0 {
		bad = append(bad, fmt.Sprintf("%d order titles are cut while their reason shows", r.OrderCut))
	}
	if r.Pop != nil {
		switch {
		case !r.Pop.Open:
			bad = append(bad, r.Pop.Sel+" did not open")
		case !r.Pop.Inside:
			bad = append(bad, r.Pop.Sel+" is not inside the window")
		case r.Pop.Covers && r.Pop.Sheet == "" && r.Pop.Sel == "#inv-tip":
			bad = append(bad, r.Pop.Sel+" covers the cell it belongs to")
		}
	}
	if o := r.Ov; o != nil {
		if len(o.Open) > 1 {
			bad = append(bad, "more than one overlay is open at once: "+strings.Join(o.Open, ", "))
		}
		for _, x := range o.Menu {
			bad = append(bad, "menu item "+x+" is covered by another layer")
		}
		for _, x := range o.MenuScroll {
			bad = append(bad, "menu: "+x)
		}
		for _, x := range o.Cut {
			bad = append(bad, "sheet text is cut, not wrapped: "+x)
		}
		for _, x := range o.Under {
			bad = append(bad, "a sheet leaves the free area (under the header or past the window): "+x)
		}
		for _, x := range o.Charter {
			bad = append(bad, "sheet: "+x)
		}
		for _, x := range o.Small {
			bad = append(bad, "target under 44 px: "+x)
		}
		for _, x := range o.Words {
			bad = append(bad, "quest words (no chapter counter, no margin number): "+x)
		}
	}
	return append(bad, r.clickVerdict()...)
}

// clickVerdict judges the press the harness made (click=) and the elements it expects to be hittable afterwards (hit=).
func (r fitResult) clickVerdict() []string {
	var bad []string
	if r.Missing != "" {
		bad = append(bad, "there is nothing to press at "+r.Missing+" (the page lacks it)")
	}
	if r.Clicked != "" {
		sels := make([]string, 0, len(r.Hits))
		for sel := range r.Hits {
			sels = append(sels, sel)
		}
		sort.Strings(sels)
		for _, sel := range sels {
			if sel == "#btn-layout" && r.VW < 700 {
				continue // a phone cannot edit the layout: the button is not offered there (shell.js), so it cannot be covered
			}
			if !r.Hits[sel] {
				bad = append(bad, "after pressing "+r.Clicked+", "+sel+" is not hittable (the sheet is clipped or covered)")
			}
		}
	}
	return bad
}

// fitBuilds is the Build tab as the page shows it (fit.js buildsPage).
type fitBuilds struct {
	Cards  int      `json:"cards"`
	Open   int      `json:"open"`
	Cols   int      `json:"cols"`
	Past   []string `json:"past"`
	Boxes  []string `json:"boxes"`
	Cut    []string `json:"cut"`
	Small  []string `json:"small"`
	Notext []string `json:"notext"`
}

// judge says what is wrong with the cards: one column under 700 px and a grid above (while no detail is open: the open card
// takes the whole row), the page scrolls and no box inside it does, nothing past the window or cut, 44 px buttons, a text
// equivalent on every star row. A nil one (a page without the list) is not judged.
func (b *fitBuilds) judge(vw, scrollW int) []string {
	if b == nil {
		return nil
	}
	var bad []string
	if b.Cards == 0 {
		bad = append(bad, "the build list has no card")
	}
	if b.Open == 0 && b.Cards > 1 {
		if vw < fitNavPhone && b.Cols != 1 {
			bad = append(bad, fmt.Sprintf("build cards in %d columns at %d px wide, want 1 on a phone", b.Cols, vw))
		}
		if vw >= fitNavPhone && b.Cols < 2 {
			bad = append(bad, fmt.Sprintf("build cards in %d column at %d px wide, want a grid", b.Cols, vw))
		}
	}
	if scrollW > vw {
		bad = append(bad, fmt.Sprintf("the page scrolls sideways: scrollWidth %d > innerWidth %d", scrollW, vw))
	}
	for _, s := range b.Past {
		bad = append(bad, "build card ends outside the window: "+s)
	}
	for _, s := range b.Boxes {
		bad = append(bad, "build cards hold a scroll box: "+s)
	}
	for _, s := range b.Cut {
		bad = append(bad, "build card text is cut: "+s)
	}
	for _, s := range b.Small {
		bad = append(bad, "build card button too small: "+s)
	}
	for _, s := range b.Notext {
		bad = append(bad, "build card stars lack their text equivalent: "+s)
	}
	return bad
}

// fitChars is what fit.js returns for the Characters page (charsPage).
type fitChars struct {
	Clipped []string `json:"clipped"`
	Small   []string `json:"small"`
	Covered []string `json:"covered"`
	Outside []string `json:"outside"`
	Under   string   `json:"under"`
	Sheet   string   `json:"sheet"`
	Hole    int      `json:"hole"`
	HoleAt  [2]int   `json:"holeAt"`
	Cards   int      `json:"cards"`
	Loading bool     `json:"loading"`
	Open    bool     `json:"open"`
}

// fitCharsHole is the largest empty square inside the Characters list that is not a gap (px): a missing card in a row of cards is as high as a
// card (about 190 px); the phone's single column has none. fitCharsHolePhone is the same on a phone.
const (
	fitCharsHole      = 300
	fitCharsHolePhone = 240
)

// judge is what is wrong with a measured Characters page (docs/charselect-ui.md); vw tells the phone from the desk.
func (c *fitChars) judge(vw int) []string {
	var bad []string
	if c.Loading {
		bad = append(bad, "the Characters page is still loading")
	}
	for _, x := range c.Clipped {
		bad = append(bad, "clipped text: "+x)
	}
	for _, x := range c.Small {
		bad = append(bad, "target under 44 px: "+x)
	}
	for _, x := range c.Covered {
		bad = append(bad, "covered control: "+x)
	}
	for _, x := range c.Outside {
		bad = append(bad, "ends outside the window: "+x)
	}
	if c.Under != "" {
		bad = append(bad, "under the header: "+c.Under)
	}
	if c.Sheet != "" {
		bad = append(bad, "sheet cut: "+c.Sheet)
	}
	lim := fitCharsHole
	if vw <= 480 {
		lim = fitCharsHolePhone
	}
	if c.Hole > lim {
		bad = append(bad, fmt.Sprintf("dead space in the list: a %d px square is empty at %d,%d", c.Hole, c.HoleAt[0], c.HoleAt[1]))
	}
	return bad
}

// fitVault is what fit.js returns for the Ethereal Vault page (vaultPage).
type fitVault struct {
	Clipped []string `json:"clipped"`
	Small   []string `json:"small"`
	Covered []string `json:"covered"`
	Outside []string `json:"outside"`
	Under   string   `json:"under"`
	Sheet   string   `json:"sheet"`
	Unnamed []string `json:"unnamed"`
	Holes   []string `json:"holes"`
	Cells   int      `json:"cells"`
	State   string   `json:"state"`
	Open    bool     `json:"open"`
	Meter   bool     `json:"meter"`
	Mode    string   `json:"mode"`
}

// maxVaultCells is the most cells the Vault page may have in the DOM at once: a full window plus its margin, on the tallest
// supported window and the narrowest cells (the vault itself holds thousands).
const maxVaultCells = 500

// judge is what is wrong with a measured Vault page (docs/vault.md).
func (v *fitVault) judge() []string {
	var bad []string
	if v.State == "loading" {
		bad = append(bad, "the Vault page is still loading")
	}
	if !v.Meter {
		bad = append(bad, "no capacity meter with a value")
	}
	for _, x := range v.Clipped {
		bad = append(bad, "clipped text: "+x)
	}
	for _, x := range v.Small {
		bad = append(bad, "target under 44 px: "+x)
	}
	for _, x := range v.Covered {
		bad = append(bad, "covered control: "+x)
	}
	for _, x := range v.Outside {
		bad = append(bad, "ends outside the window: "+x)
	}
	if v.Under != "" {
		bad = append(bad, "under the header: "+v.Under)
	}
	if v.Sheet != "" {
		bad = append(bad, "sheet cut: "+v.Sheet)
	}
	for _, x := range v.Unnamed {
		bad = append(bad, "control with no name: "+x)
	}
	for _, x := range v.Holes {
		bad = append(bad, "placeholder: "+x)
	}
	if v.Cells > maxVaultCells {
		bad = append(bad, fmt.Sprintf("cells drawn for the whole list: %d (the grid is windowed)", v.Cells))
	}
	return bad
}

// mustFail names, by file prefix, the pages that must FAIL and the verdict each must fail with. Another verdict, or a
// broken measurement, does not count: only the reason the fixture was built for.
var mustFail = map[string]string{
	"fit-overflow-fixture-":    "overflows its box",
	"fit-overflow-nested-":     "overflows its box",
	"fit-overflow-fakescroll-": "overflows its box",
	"fit-overflow-classic-":    "the page scrolls",
	"fit-blank-rail-":          "holds no visible panel",
	"fit-tall-frame-":          "ends outside the window",
	// the overlay policy (card overlayfix): the real page with one defect put back (fixtures_overlay.go)
	"fit-overlaydefect-under-":  "covered by another layer",
	"fit-overlaydefect-clip-":   "sheet text is cut",
	"fit-overlaydefect-header-": "leaves the free area",
	"fit-overlaydefect-two-":    "more than one overlay is open",
	// the "..." menu (card menufit2): one defect each, put back by defectCSS "mn-*" (fixtures_menu.go)
	"fit-menudefect-noscroll-": "cannot scroll",
	"fit-menudefect-primary-":  "is below the fold",
	"fit-menudefect-hint-":     "shows no hint",
	"fit-moredefect-clip-":     "the More list shows",
	"fit-moredefect-phone-":    "the More sheet is",
	// the quest board (slice S5): one defect each, put back by defectCSS and overlayJS in fixtures_quests.go
	"fit-questdefect-clip-":  "sheet text is cut",
	"fit-questdefect-small-": "target under 44 px",
	"fit-questdefect-words-": "quest words",
	"fit-questdefect-two-":   "more than one overlay is open",
	// the phone's storage tab (card uiphone): the real page with one defect of the 2026-11-15 screenshot put back (defectCSS in fixtures_phone.go)
	"fit-phonedefect-label-":     "label over art",
	"fit-phonedefect-acts-":      "action buttons: ",
	"fit-phonedefect-cutchips-":  "toolbar cut: ",
	"fit-phonedefect-hover-":     "horizontal overflow: ",
	"fit-phonedefect-narrow-":    "grid: ",
	"fit-phonedefect-resetclip-": "#btn-reset is not hittable",
	// the video card (card videoctl): the toolbar put back over the caption (defectCSS vid-ctl in fixtures_phone.go)
	"fit-viddefect-ctl-": "video control covers text",
	// the Characters page (docs/charselect-ui.md): one defect each, put back by defectCSS in fixtures_chars.go
	"fit-charsdefect-clip-":    "clipped text",
	"fit-charsdefect-wide-":    "ends outside the window",
	"fit-charsdefect-under-":   "under the header",
	"fit-charsdefect-small-":   "target under 44 px",
	"fit-charsdefect-hole-":    "dead space in the list",
	"fit-charsdefect-sheet-":   "sheet cut",
	"fit-charsdefect-covered-": "covered control",
	// the Ethereal Vault page (docs/vault.md): one defect each, put back by defectCSS and vaultJS in fixtures_vault.go
	"fit-vaultdefect-clip-":       "clipped text",
	"fit-vaultdefect-wide-":       "ends outside the window",
	"fit-vaultdefect-under-":      "under the header",
	"fit-vaultdefect-small-":      "target under 44 px",
	"fit-vaultdefect-covered-":    "covered control",
	"fit-vaultdefect-sheet-":      "sheet cut",
	"fit-vaultdefect-tip-":        "is not inside the window",
	"fit-vaultdefect-unwindowed-": "cells drawn for the whole list",
	"fit-vaultdefect-noname-":     "control with no name",
	"fit-vaultdefect-enabled-":    "placeholder: ",
	"fit-vaultdefect-ethenabled-": "placeholder: ",
	// the shared navigation (card uinav2): one defect each, put back by defectCSS "nav-*" (fixtures_nav.go)
	"fit-navdefect-covers-": "the bottom bar covers content",
	"fit-navdefect-small-":  "navigation target too small",
	"fit-navdefect-sheet-":  "the More sheet overlaps the bottom bar",
	"fit-navdefect-wide-":   "navigation is a bar=true",
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
		if msg := r.broken(); msg != "" {
			lines = append(lines, "FAIL "+name+": "+msg)
			ok = false
			continue
		}
		bad := r.judge()
		// the browser window must really be the size the name says, or the numbers are for another page
		if wh := reSize.FindStringSubmatch(name); wh != nil && (strconv.Itoa(r.VW) != wh[1] || strconv.Itoa(r.VH) != wh[2]) {
			lines = append(lines, fmt.Sprintf("FAIL %s: the viewport is %dx%d, not %sx%s", name, r.VW, r.VH, wh[1], wh[2]))
			ok = false
			continue
		}
		why := ""
		for prefix, w := range mustFail {
			if strings.HasPrefix(name, prefix) {
				why = w
			}
		}
		hit := ""
		for _, b := range bad {
			if why != "" && strings.Contains(b, why) {
				hit = b
			}
		}
		switch {
		case why != "" && hit == "":
			lines = append(lines, fmt.Sprintf("FAIL %s: must fail with %q; verdict %v", name, why, bad))
			ok = false
		case why != "":
			lines = append(lines, fmt.Sprintf("ok   %s: failed as it must (%s)", name, hit))
		case len(bad) > 0:
			lines = append(lines, "FAIL "+name+": "+strings.Join(bad, "; "))
			ok = false
		default:
			extra := ""
			if r.Zones.N > 0 {
				extra = fmt.Sprintf(", %d zones in %d rows (least %d px)", r.Zones.N, r.Zones.Rows, r.Zones.MinH)
				if r.Zones.Menu {
					extra += ", menu open in " + r.Zones.Host
				}
			}
			if c := r.Chars; c != nil {
				lines = append(lines, fmt.Sprintf("ok   %s: %dx%d, no scroll, %d cards, largest empty square %d px", name, r.VW, r.VH, c.Cards, c.Hole))
				continue
			}
			if v := r.Vault; v != nil {
				lines = append(lines, fmt.Sprintf("ok   %s: %dx%d, no scroll, %d cells, %s view, state %q", name, r.VW, r.VH, v.Cells, v.Mode, v.State))
				continue
			}
			lines = append(lines, fmt.Sprintf("ok   %s: %dx%d, no scroll, %d slots, largest gap %d px%s", name, r.VW, r.VH, len(r.Slots), r.Hole, extra))
		}
	}
	return lines, ok, nil
}

// fitNav is the shared navigation as the page shows it (fit.js navPage).
type fitNav struct {
	Missing bool     `json:"missing"`
	Bar     bool     `json:"bar"`
	Labels  []string `json:"labels"`
	Current int      `json:"current"`
	Tiny    []string `json:"tiny"`
	Covers  []string `json:"covers"`
	Inside  bool     `json:"inside"`
	Host    string   `json:"host"`
	Sheet   *struct {
		Inside     bool `json:"inside"`
		ClearOfBar bool `json:"clearOfBar"`
		First      bool `json:"first"`
	} `json:"sheet"`
}

// fitNavPhone is the width under which the navigation is a bottom bar (web/nav.json "breakpoint").
const fitNavPhone = 700

// judge says what is wrong with the navigation; a nil one (a page without it) is not judged.
func (n *fitNav) judge(vw int) []string {
	if n == nil {
		return nil
	}
	if n.Missing {
		return []string{"the page has no navigation (nav.js did not draw it)"}
	}
	var bad []string
	if got := strings.Join(n.Labels, ","); got != "Live,Gear,Skills,Vault,More" {
		bad = append(bad, "navigation lists "+got+", want Live,Gear,Skills,Vault,More")
	}
	if n.Current != 1 {
		bad = append(bad, fmt.Sprintf("navigation marks %d destinations as current, want exactly 1", n.Current))
	}
	if n.Bar != (vw < fitNavPhone) {
		bad = append(bad, fmt.Sprintf("navigation is a bar=%v at %d px wide (a bar under %d px, the band above)", n.Bar, vw, fitNavPhone))
	}
	if !n.Inside {
		bad = append(bad, "navigation is not inside the window")
	}
	for _, t := range n.Tiny {
		bad = append(bad, "navigation target too small: "+t)
	}
	for _, c := range n.Covers {
		bad = append(bad, "the bottom bar covers content: "+c)
	}
	if s := n.Sheet; s != nil {
		if !s.Inside {
			bad = append(bad, "the More sheet is not inside the window")
		}
		if !s.ClearOfBar {
			bad = append(bad, "the More sheet overlaps the bottom bar")
		}
		if !s.First {
			bad = append(bad, "the More sheet holds no focus after opening")
		}
	}
	return bad
}
