package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// supportui.js in a stub DOM (node, skipped without it): both boxes start off, nothing is posted before a press, the
// preview shows the files and the log, a blocked bundle has no Send, no endpoint means save only, Send shows the code.
func TestSupportUIJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/supportui.js")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "supportui.mjs"), js, 0o600)
	harness := `
import { mountSupport, size, refusal, INTRO, SAVE_WORDS } from './supportui.mjs';
const mk = (tag) => ({ tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, checked: false,
  append(...c) { this.children.push(...c); }, appendChild(c) { this.children.push(c); return c; }, setAttribute(k, v) { this.attrs[k] = v; },
  addEventListener(n, f) { (this.listeners[n] = this.listeners[n] || []).push(f); }, focus() { doc.activeElement = this; } });
const doc = { activeElement: null, body: mk('body'), listeners: {}, createElement: mk, addEventListener(n, f) { this.listeners[n] = f; } };
const ui = { btn: (label) => { const b = mk('button'); b.textContent = label; return b; } };
const toasts = [], posts = [];
let reply = null;
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const out = { size: [size(12), size(2048), size(3 * 1048576)], refs: [refusal(null), refusal({ code: 'off' }), refusal({ code: 'blocked' }), refusal({ error: 'boom' })] };
mountSupport({ ui, toast: (s, t) => toasts.push(s + ':' + t), doc, post: (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return Promise.resolve(reply); } });
const nodes = all(doc.body), sheet = nodes.find((n) => n.id === 'support-sheet');
const byId = (id) => nodes.find((n) => n.id === id);
const shot = byId('support-shot'), save = byId('support-save'), build = byId('support-build'), send = byId('support-send'), link = byId('support-save-file'), code = byId('support-code');
out.copy = nodes.some((n) => n.textContent === INTRO) && nodes.some((n) => n.textContent === SAVE_WORDS);
out.startHidden = sheet.hidden;
doc.listeners['d2:support-ask']({ detail: {} });
out.opened = !sheet.hidden; out.defaultsOff = !shot.checked && !save.checked; out.noPostOnOpen = posts.length === 0;
out.visibleOnForm = { build: !build.hidden, send: !send.hidden, link: !link.hidden };
const file = (n, s) => ({ name: n, size: s });
reply = { schema: 'watch.support/1', id: 'abc 1', files: [file('MANIFEST.txt', 300), file('log.txt', 4096)], logPreview: 'line one\nline two', findings: [], canSend: true, canSave: true, configured: true };
shot.checked = true;
build.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.buildPost = posts[0];
out.fileLines = all(sheet).filter((n) => n.tag === 'li').map((n) => n.textContent);
out.log = all(sheet).find((n) => n.tag === 'pre').textContent;
out.preview = { send: !send.hidden, link: !link.hidden, buildHidden: build.hidden, href: link.attrs.href };
reply = { schema: 'watch.support/1', code: 'ABCD2345' };
send.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.sendPost = posts[1]; out.code = code.textContent; out.toasts = toasts;
// reopening resets the boxes and the step
doc.listeners['d2:support-ask']({ detail: {} });
out.reopenOff = !shot.checked && !save.checked && !build.hidden && send.hidden;
// a blocked bundle: no Send, no Save, the findings are listed, a refusal on send is words
reply = { schema: 'watch.support/1', id: 'x', files: [file('log.txt', 10)], logPreview: '', findings: [{ file: 'log.txt', line: 3, class: 'ip', context: 'peer [***]' }], canSend: false, canSave: false, configured: true };
build.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.blocked = { send: !send.hidden, link: !link.hidden, text: all(sheet).map((n) => n.textContent).filter((x) => x.includes('log.txt line 3')) };
// no endpoint: save only
reply = { schema: 'watch.support/1', id: 'y', files: [file('log.txt', 10)], logPreview: 'a', findings: [], canSend: false, canSave: true, configured: false };
doc.listeners['d2:support-ask']({ detail: {} });
build.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
out.unconfigured = { send: !send.hidden, link: !link.hidden, said: all(sheet).some((n) => n.textContent.startsWith('Uploading is not set up')) };
// a failed send keeps the sheet and says why
reply = { schema: 'watch.support/1', id: 'z', files: [file('log.txt', 10)], logPreview: 'a', findings: [], canSend: true, canSave: true, configured: true };
doc.listeners['d2:support-ask']({ detail: {} });
build.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
reply = { code: 'send', error: 'too many uploads just now' };
send.listeners.click[0](); await new Promise((r) => setTimeout(r, 5));
const err = all(sheet).find((n) => n.className === 'd-sup-err');
out.failed = { open: !sheet.hidden, err: !err.hidden && err.textContent };
console.log(JSON.stringify(out));
`
	os.WriteFile(filepath.Join(dir, "h.mjs"), []byte(harness), 0o600)
	b, err := exec.Command(node, filepath.Join(dir, "h.mjs")).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	var o map[string]any
	if err := json.Unmarshal(b, &o); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	j := string(b)
	for _, want := range []string{
		`"size":["12 B","2.0 KB","3.0 MB"]`, `"copy":true`, `"startHidden":true`, `"opened":true`, `"defaultsOff":true`, `"noPostOnOpen":true`,
		`"visibleOnForm":{"build":true,"send":false,"link":false}`,
		`"buildPost":"/api/support/bundle {\"includeScreenshot\":true,\"includeSave\":false}"`,
		`"fileLines":["MANIFEST.txt  300 B","log.txt  4.0 KB"]`, `"log":"line one\nline two"`,
		`"preview":{"send":true,"link":true,"buildHidden":true,"href":"/api/support/download?id=abc%201"}`,
		`"sendPost":"/api/support/send {\"id\":\"abc 1\"}"`, `"code":"ABCD2345"`, `"reopenOff":true`,
		`"blocked":{"send":false,"link":false`, `"unconfigured":{"send":false,"link":true,"said":true}`,
		`"failed":{"open":true,"err":"too many uploads just now"}`,
	} {
		if !strings.Contains(j, want) {
			t.Errorf("missing %s in\n%s", want, j)
		}
	}
	if !strings.Contains(j, "Not sent") && !strings.Contains(j, "log.txt line 3 (ip): peer [***]") {
		t.Errorf("findings not listed:\n%s", j)
	}
}
