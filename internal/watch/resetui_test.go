package watch

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// resetui.js in a stub DOM (node, skipped without it): the sheet copy, the press, the strip, the refusal and the done toast.
func TestResetUIJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/resetui.js")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "resetui.mjs"), js, 0o600)
	harness := `
import { phase, refusal, mountReset, COPY, COOLDOWN_MS, LABEL } from './resetui.mjs';
const mk = (tag) => ({ tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false,
  append(...c) { this.children.push(...c); }, appendChild(c) { this.children.push(c); return c; }, setAttribute(k, v) { this.attrs[k] = v; },
  addEventListener(n, f) { (this.listeners[n] = this.listeners[n] || []).push(f); }, focus() { doc.activeElement = this; } });
const doc = { activeElement: null, body: mk('body'), listeners: {}, createElement: mk, addEventListener(n, f) { this.listeners[n] = f; } };
const ui = { btn: (label) => { const b = mk('button'); b.textContent = label; return b; } };
const toasts = [], posts = [];
let view = { schema: 'watch.reset/1', state: 'idle' }, reply = null, clock = 100000;
const later = (fn) => { later.q.push(fn); };
later.q = [];
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const tick = () => new Promise((r) => setTimeout(r, 5));
const out = {};
out.phaseIdle = phase({ schema: 'watch.reset/1', state: 'idle' });
out.phaseRes = phase({ schema: 'watch.reset/1', state: 'resetting', reason: 'leaving the game' });
out.phaseHard = phase({ schema: 'watch.reset/1', state: 'rejoining', hard: true });
out.phaseStuck = phase({ schema: 'watch.reset/1', state: 'resetting', hard: true, reason: 'the bot was stuck, restarting it' });
out.phaseFail = phase({ schema: 'watch.reset/1', state: 'failed', reason: 'x' });
out.refOff = refusal({ code: 'off' }); out.refRun = refusal({ code: 'measured_run' }); out.refNone = refusal(null);
const btn = mk('button'); btn.textContent = 'Reset';
mountReset({ ui, button: btn, now: () => clock, toast: (s, t) => toasts.push(s + ':' + t), doc, later,
  post: (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return Promise.resolve(reply); }, get: () => Promise.resolve(view) });
await tick();
const nodes = all(doc.body), sheet = nodes.find((n) => n.id === 'reset-confirm'), strip = nodes.find((n) => n.id === 'reset-progress');
const go = nodes.find((n) => n.id === 'reset-go'), hard = nodes.find((n) => n.dataset.key === 'reset-hard');
const text = nodes.find((n) => n.className === 'd-reset-text');
out.copy = nodes.some((n) => n.textContent === COPY);
out.label = btn.attrs['aria-label'] === LABEL && LABEL === 'Reset bot (leaves the game and rejoins)';
out.startHidden = sheet.hidden && strip.hidden && hard.hidden && btn.textContent === 'Reset';
// one tap: the request goes out and the feedback is there before the reply
reply = view = { schema: 'watch.reset/1', state: 'resetting', reason: 'leaving the game', since: 'x' };
btn.listeners.click[0]();
out.instant = { sheetHidden: sheet.hidden, label: btn.textContent, strip: strip.hidden ? '' : text.textContent, toast: toasts.slice(), posted: posts.slice() };
await tick();
// a second press inside the cooldown (and the menu/orders event) is a no-op
btn.listeners.click[0](); doc.listeners['d2:reset-ask']({ detail: {} }); await tick();
out.cooldownPosts = posts.length;
out.busyLabel = btn.textContent;
view = { schema: 'watch.reset/1', state: 'idle', last: { result: 'done' } };
while (later.q.length) later.q.shift()(); await tick();
out.stripGone = strip.hidden; out.labelBack = btn.textContent; out.toasts = toasts.slice();
// after the cooldown a press goes out again; the menu event is the same one tap
clock += COOLDOWN_MS + 1;
doc.listeners['d2:reset-ask']({ detail: {} }); await tick();
out.afterCooldown = posts.length;
// a refusal gives the button back and says why
clock += COOLDOWN_MS + 1; reply = { code: 'measured_run', error: 'no' };
btn.listeners.click[0](); await tick();
out.refused = { label: btn.textContent, toast: toasts[toasts.length - 1], strip: strip.hidden };
clock += 1; reply = { schema: 'watch.reset/1', state: 'resetting' };
view = { schema: 'watch.reset/1', state: 'failed', reason: 'the game did not come back' };
btn.listeners.click[0](); await tick();
out.retryAfterRefusal = posts.length;
// a failed reset offers the restart; only that one has a confirm
while (later.q.length) later.q.shift()(); await tick();
out.failedStrip = !strip.hidden && !hard.hidden && text.textContent;
hard.listeners.click[0]();
out.hardSheet = !sheet.hidden && doc.activeElement === go;
reply = { schema: 'watch.reset/1', state: 'resetting', hard: true, reason: 'restarting the bot' };
go.listeners.click[0](); await tick();
out.hardPosted = posts[posts.length - 1]; out.hardClosed = sheet.hidden;
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
		`"copy":true`, `"label":true`, `"startHidden":true`,
		`"instant":{"sheetHidden":true,"label":"Resetting...","strip":"Resetting... leaving the game","toast":["info:Reset sent"],"posted":["/api/reset {}"]}`,
		`"cooldownPosts":1`, `"busyLabel":"Resetting..."`, `"stripGone":true`, `"labelBack":"Reset"`, "good:Reset done",
		`"afterCooldown":2`, `"refused":{"label":"Reset","toast":"warn:A measured run is never reset","strip":true}`, `"retryAfterRefusal":4`,
		`"failedStrip":"Reset failed: the game did not come back"`, `"hardSheet":true`, `"hardPosted":"/api/reset {\"hard\":true}"`, `"hardClosed":true`,
		`"phaseStuck":{"active":true,"text":"Resetting... the bot was stuck, restarting it"`, `"text":"Restarting the bot..."`, `"refOff":"Reset is not available on this hero yet"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
}

// Every module the shell imports from /watch/ is served: a missing route blanks the whole page (the fit check caught it once).
func TestShellImportsAreServed(t *testing.T) {
	js, _ := webFS.ReadFile("web/shell.js")
	sim, err := NewSim("sorceress", 1)
	if err != nil {
		t.Skip(err)
	}
	h := sim.Hub.ViewHandler()
	n := 0
	for _, line := range strings.Split(string(js), "\n") {
		i := strings.Index(line, "from '/watch/")
		if !strings.HasPrefix(line, "import ") || i < 0 {
			continue
		}
		p := line[i+len("from '"):]
		p = p[:strings.Index(p, "'")]
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		if w.Code != 200 {
			t.Errorf("%s: %d", p, w.Code)
		}
		n++
	}
	if n < 2 {
		t.Fatalf("found %d imports", n)
	}
}
