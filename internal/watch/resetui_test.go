package watch

import (
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
import { phase, refusal, mountReset, COPY } from './resetui.mjs';
const mk = (tag) => ({ tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false,
  append(...c) { this.children.push(...c); }, appendChild(c) { this.children.push(c); return c; }, setAttribute(k, v) { this.attrs[k] = v; },
  addEventListener(n, f) { (this.listeners[n] = this.listeners[n] || []).push(f); }, focus() { doc.activeElement = this; } });
const doc = { activeElement: null, body: mk('body'), listeners: {}, createElement: mk, addEventListener(n, f) { this.listeners[n] = f; } };
const ui = { btn: (label) => { const b = mk('button'); b.textContent = label; return b; } };
const toasts = [], posts = [];
let view = { schema: 'watch.reset/1', state: 'idle' }, reply = null;
const later = (fn) => { later.q.push(fn); };
later.q = [];
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const out = {};
out.phaseIdle = phase({ schema: 'watch.reset/1', state: 'idle' });
out.phaseRes = phase({ schema: 'watch.reset/1', state: 'resetting', reason: 'leaving the game' });
out.phaseHard = phase({ schema: 'watch.reset/1', state: 'rejoining', hard: true });
out.phaseFail = phase({ schema: 'watch.reset/1', state: 'failed', reason: 'x' });
out.refOff = refusal({ code: 'off' }); out.refRun = refusal({ code: 'measured_run' }); out.refNone = refusal(null);
mountReset({ ui, toast: (s, t) => toasts.push(s + ':' + t), doc, later,
  post: (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return Promise.resolve(reply); }, get: () => Promise.resolve(view) });
await new Promise((r) => setTimeout(r, 5));
const nodes = all(doc.body), sheet = nodes.find((n) => n.id === 'reset-confirm'), strip = nodes.find((n) => n.id === 'reset-progress');
const go = nodes.find((n) => n.id === 'reset-go'), hard = nodes.find((n) => n.dataset.key === 'reset-hard');
out.copy = nodes.some((n) => n.textContent === COPY);
out.buttons = nodes.filter((n) => n.tag === 'button' && !n.hidden).map((n) => n.textContent);
out.startHidden = sheet.hidden && strip.hidden && hard.hidden;
doc.listeners['d2:reset-ask']({ detail: {} });
out.opened = !sheet.hidden; out.focusGo = doc.activeElement === go;
reply = { code: 'measured_run', error: 'no' };
go.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.refusedStays = !sheet.hidden && nodes.some((n) => n.className === 'd-reset-err' && !n.hidden && n.textContent === 'A measured run is never reset');
reply = view = { schema: 'watch.reset/1', state: 'resetting', reason: 'leaving the game', since: 'x' };
go.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.posted = posts.slice(); out.closedAfter = sheet.hidden; out.strip = strip.hidden ? '' : strip.textContent;
view = { schema: 'watch.reset/1', state: 'idle', last: { result: 'done' } };
while (later.q.length) later.q.shift()(); await new Promise((r) => setTimeout(r, 5));
out.stripGone = strip.hidden; out.toasts = toasts;
view = { schema: 'watch.reset/1', state: 'failed', reason: 'the game did not come back' };
doc.listeners['d2:reset-ask']({ detail: {} });
reply = { schema: 'watch.reset/1', state: 'resetting' }; go.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
while (later.q.length) later.q.shift()(); await new Promise((r) => setTimeout(r, 5));
out.hardShown = !hard.hidden;
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
		`"copy":true`, `"buttons":["Cancel","Reset"]`, `"startHidden":true`, `"opened":true`, `"focusGo":true`, `"refusedStays":true`,
		`"posted":["/api/reset {}","/api/reset {}"]`, `"closedAfter":true`, `"strip":"Resetting... leaving the game"`, `"stripGone":true`,
		"good:Reset done", `"hardShown":true`, `"text":"Restarting the bot..."`, `"refOff":"Reset is not available on this hero yet"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in\n%s", want, got)
		}
	}
}
