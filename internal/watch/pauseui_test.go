package watch

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pauseui.js in a stub DOM (node, skipped without it): no button without can_freeze, the press, the banner and the dimmed
// state, the refusal toast, a hero with no pause desk, and the Space rule.
func TestPauseUIJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/pauseui.js")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pauseui.mjs"), js, 0o600)
	harness := `
import { phase, bannerText, clock, spaceToggles, mountPause } from './pauseui.mjs';
const mk = (tag) => ({ tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false,
  classList: { add() {} }, closest() { return null; },
  append(...c) { this.children.push(...c); }, appendChild(c) { this.children.push(c); return c; }, insertBefore(c) { this.children.unshift(c); return c; },
  setAttribute(k, v) { this.attrs[k] = v; }, addEventListener(n, f) { (this.listeners[n] = this.listeners[n] || []).push(f); } });
const html = mk('html');
const pop = mk('div'); pop.id = 'menu-more';
const doc = { documentElement: html, activeElement: null, body: mk('body'), listeners: {}, createElement: mk,
  getElementById: (id) => (id === 'menu-more' ? pop : null), querySelector: () => null, addEventListener(n, f) { this.listeners[n] = f; } };
const ui = { btn: (label) => { const b = mk('button'); b.textContent = label; return b; } };
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const out = {};
const frozen = { schema: 'watch.pause/1', state: 'paused', frozen: true, can_freeze: true, since: '2026-11-14T10:00:00', cap_min: 60, monsters_near: true };
out.none = phase(null); out.running = phase({ schema: 'watch.pause/1', state: 'running' });
out.can = phase({ schema: 'watch.pause/1', state: 'running', can_freeze: true });
out.walk = phase({ schema: 'watch.pause/1', state: 'paused' });
out.frozen = phase(frozen);
out.banner = bannerText(phase(frozen)); out.plain = bannerText(phase({ schema: 'watch.pause/1', state: 'paused', frozen: true }));
out.clock = clock('2026-11-14T10:05:00'); out.clockBad = clock('nope');
const key = (o) => Object.assign({ key: ' ', code: 'Space', repeat: false }, o);
const d2 = (a) => ({ activeElement: a, querySelector: () => null });
const el = (sel) => ({ closest: (s) => (s.split(',').map((x) => x.trim()).includes(sel) ? {} : null) });
out.sp = [spaceToggles(key({}), d2(null)), spaceToggles(key({}), d2(el('input'))), spaceToggles(key({}), d2(el('button'))), spaceToggles(key({}), d2(el('a'))),
  spaceToggles(key({ repeat: true }), d2(null)), spaceToggles(key({ ctrlKey: true }), d2(null)), spaceToggles(key({ key: 'a', code: 'KeyA' }), d2(null)),
  spaceToggles(key({}), { activeElement: null, querySelector: () => ({}) })];
let view = { schema: 'watch.pause/1', state: 'running' }, reply = null; const posts = [], toasts = [];
const later = (fn) => { later.q.push(fn); }; later.q = [];
const m = mountPause({ ui, toast: (s, t) => toasts.push(s + ':' + t), doc, later, post: (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return Promise.resolve(reply); }, get: () => Promise.resolve(view) });
await new Promise((r) => setTimeout(r, 5));
const nodes = all(doc.body).concat(all(pop));
const btn = nodes.find((n) => n.id === 'btn-pause'), banner = nodes.find((n) => n.id === 'pause-banner');
out.inMenu = pop.children[0] === btn;
out.hiddenWithoutCan = btn.hidden && banner.hidden && html.dataset.paused === undefined;
view = { schema: 'watch.pause/1', state: 'running', can_freeze: true }; while (later.q.length) later.q.shift()(); await new Promise((r) => setTimeout(r, 5));
out.shown = !btn.hidden && btn.textContent === 'Pause';
reply = frozen; btn.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.posted = posts.slice(); out.paused = html.dataset.paused; out.bannerOn = !banner.hidden; out.label = btn.textContent;
const go = nodes.find((n) => n.dataset.key === 'pause-resume');
reply = { schema: 'watch.pause/1', state: 'running', can_freeze: true }; go.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.resumed = posts.slice(1); out.dimGone = html.dataset.paused === undefined && banner.hidden;
reply = { schema: 'watch.pause/1', state: 'running', can_freeze: true, error: 'could not freeze the game: x' };
doc.listeners.keydown({ ...key({}), target: null, preventDefault() { out.prevented = true; } }); await new Promise((r) => setTimeout(r, 5));
out.space = posts.slice(2); out.toasts = toasts;
const doc2 = { ...doc, getElementById: () => null, body: mk('body'), documentElement: mk('html') };
const later2 = (fn) => { later2.q.push(fn); }; later2.q = [];
view = null; const m2 = mountPause({ ui, toast() {}, doc: doc2, later: later2, post: () => Promise.resolve(null), get: () => Promise.resolve(null) });
await new Promise((r) => setTimeout(r, 5));
out.noDesk = all(doc2.body).find((n) => n.id === 'pause-banner').hidden && later2.q.length === 0;
console.log(JSON.stringify(out));
`
	f := filepath.Join(dir, "h.mjs")
	os.WriteFile(f, []byte(harness), 0o600)
	b, err := exec.Command(node, f).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, b)
	}
	got := string(b)
	for _, want := range []string{
		`"none":{"shown":false`, `"can":{"shown":true,"paused":false,"label":"Pause"`, `"walk":{"shown":false`, `"frozen":{"shown":true,"paused":true,"label":"Resume"`,
		`"banner":"Paused since 10:00 (resumes by itself after 60 min). Monsters are near."`, `"plain":"Paused"`, `"clock":"10:05","clockBad":""`,
		`"sp":[true,false,false,false,false,false,false,false]`, `"inMenu":true`, `"hiddenWithoutCan":true`, `"shown":true`,
		`"posted":["/api/pause {\"on\":true}"]`, `"paused":"1"`, `"bannerOn":true`, `"label":"Resume"`, `"resumed":["/api/pause {\"on\":false}"]`,
		`"dimGone":true`, `"space":["/api/pause {\"on\":true}"]`, `"prevented":true`, `warn:could not freeze the game: x`, `"noDesk":true`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
}

// pause.css uses tokens only, the banner and the button keep a 44 px target, and the page serves both new files.
func TestPauseAssets(t *testing.T) {
	css, err := webFS.ReadFile("web/pause.css")
	if err != nil {
		t.Fatal(err)
	}
	s := string(css)
	for _, bad := range []string{"#", "rgb(", "hsl(", "px"} {
		for _, line := range strings.Split(s, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "/*") {
				continue
			}
			if strings.Contains(line, bad) && !(bad == "px" && strings.Contains(line, "var(--d-hit)")) {
				t.Errorf("pause.css line has %q outside a token: %s", bad, strings.TrimSpace(line))
			}
		}
	}
	if strings.Count(s, "var(--d-hit)") < 3 || !strings.Contains(s, `:root[data-paused] [data-panel="video"] .frame::after`) {
		t.Error("the 44 px targets or the dimmed game view are missing")
	}
	sim, err := NewSim("sorceress", 1)
	if err != nil {
		t.Skip(err)
	}
	h := sim.Hub.ViewHandler()
	for _, p := range []string{"/watch/pauseui.js", "/watch/pause.css"} {
		w := httptest_get(h, p)
		if w != 200 {
			t.Errorf("%s: %d", p, w)
		}
	}
	html, _ := webFS.ReadFile("web/watch.html")
	if !strings.Contains(string(html), "/watch/pause.css") {
		t.Error("watch.html does not link pause.css")
	}
}

func httptest_get(h http.Handler, p string) int {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
	return w.Code
}
