package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// lootrulesui.js in a stub DOM (node, skipped without it): nothing is sent before a press, the Starter mode applies a preset
// with the four knobs, the Custom mode moves and switches rules off and saves the whole set as custom, a refusal is shown
// with its line or field, the recent decisions say what differs from today's code, the sheet registers with the overlay manager.
func TestLootRulesUIJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	js, err := webFS.ReadFile("web/lootrulesui.js")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "lootrulesui.mjs"), js, 0o600)
	harness := `
import { mountLootRules, stageText, refusal, decisionLine, previewLine, previewSummary } from './lootrulesui.mjs';
const mk = (tag) => ({ tag, children: [], hidden: false, className: '', set textContent(v) { this._t = v; if (v === '') this.children = []; }, get textContent() { return this._t; }, _t: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, checked: false, value: '', classList: { toggle() {} },
  append(...c) { this.children.push(...c); }, appendChild(c) { this.children.push(c); return c; }, setAttribute(k, v) { this.attrs[k] = v; },
  addEventListener(n, f) { (this.listeners[n] = this.listeners[n] || []).push(f); }, focus() { doc.activeElement = this; }, click() { this.clicked = true; } });
const doc = { activeElement: null, body: mk('body'), listeners: {}, createElement: mk, addEventListener(n, f) { this.listeners[n] = f; } };
const ui = { btn: (label) => { const b = mk('button'); b.textContent = label; return b; } };
const toasts = [], puts = [], posts = [], registered = [];
let postReply = { ok: true, seq: 3 };
let reply = { ok: true, seq: 2 };
const PV = { schema: 'watch.lootpreview/1', name: 'x', changed: 2, asOfItems: 9, totals: { sold: 3, stashed: 2, dropped: 0, kept: 4 }, note: 'The dry run does not know your marks.',
  items: [{ id: 1, name: 'Ring', place: 'pack', changed: true, wasTidy: 'stash', tidy: { action: 'stash', rule: 't-rare-jewelry', why: 'Will stash: rare jewelry' }, wasSell: 'sell', sell: { action: 'keep', rule: 's-jewelry-stash', why: 'Kept: you own a Tome of Identify' } },
    { id: 2, name: 'Cap', place: 'pack', changed: false, wasTidy: 'keep', tidy: { action: 'keep', rule: 'default', why: 'Left alone' }, wasSell: 'keep', sell: { action: 'keep', rule: 'default', why: 'x' } }] };
const rules = [
  { id: 'g-runes', scope: ['ground'], action: 'stash', reason: 'rune' },
  { id: 'g-gold', scope: ['ground'], action: 'keep', reason: 'gold' },
  { id: 't-sell', scope: ['pack'], action: 'sell', reason: 'vendor junk', off: true }];
const view = (preset, extra) => Object.assign({ schema: 'watch.lootrules/1', seq: 1, stage: 1, presets: [{ id: 'starter', name: 'Starter' }, { id: 'class:Amazon', name: 'Starter for Amazon' }],
  active: { schema: 'lootrules/1', seq: 1, name: 'Starter', limits: { maxActions: 12 }, preset, rules }, plain: { 'g-runes': 'On the ground: stash when it is a rune', 'g-gold': 'On the ground: keep gold', 't-sell': 'In the pack: sell junk' },
  options: { sellGear: 'rare', minRune: 3, minGold: 5, keepSockets: false }, canUndo: false,
  sellKinds: [{ id: 'good-gear', title: 'Good gear', text: 'Kept.' }, { id: 'jewelry', title: 'Jewelry', text: 'Sold unidentified until you own a Tome of Identify.' }], protected: [{ ID: 'invariant:mark', Text: 'Items you marked stay.' }],
  recent: [{ scope: 'ground', item: 'El Rune', today: 'ignore', rules: 'pickup', rule: 'g-runes', why: 'Will pick up: rune' }, { scope: 'sell', item: 'Cap', today: 'sell', rules: 'sell', rule: 't-sell', why: 'Will sell' }] }, extra || {});
let cur = view('starter');
let gets = 0;
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const api = mountLootRules({ ui, toast: (s, t) => toasts.push(s + ':' + t), doc, overlays: { register: (id) => registered.push(id) },
  get: () => { gets++; return Promise.resolve(cur); }, put: (p, b) => { puts.push(JSON.stringify(b)); return Promise.resolve({ status: 200, j: reply }); },
  post: (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return Promise.resolve({ status: 200, j: p.endsWith('/preview') ? PV : postReply }); } });
const nodes = all(doc.body), sheet = nodes.find((n) => n.id === 'lootrules-sheet');
const byKey = (k) => all(sheet).filter((n) => n.dataset && n.dataset.key === k);
const wait = () => new Promise((r) => setTimeout(r, 5));
const out = { registered, startHidden: sheet.hidden, noGetYet: gets === 0, stages: [stageText(0), stageText(1), stageText(2), stageText(9)] };
doc.listeners['d2:lootrules-ask']({ detail: {} }); await wait();
out.opened = !sheet.hidden; out.gets = gets; out.noPutOnOpen = puts.length === 0; out.mode = api.mode;
out.stage = nodes.find((n) => n.id === 'lootrules-stage').textContent;
const get = (id) => all(sheet).find((n) => n.id === id);
out.knobs = { gear: get('lr-gear').value, rune: get('lr-rune').value, gold: get('lr-gold').value, sock: get('lr-sock').checked, preset: get('lr-preset').value };
out.recent = get('lr-recent').children.map((n) => n.textContent);
out.differs = get('lr-recent').children.map((n) => n.dataset.differs || '');
out.sellKinds = get('lr-sellkinds').children.map((li) => li.children.map((c) => c.textContent).join(''));
out.undoOff = byKey('lr-undo')[0].disabled; out.exportHref = byKey('lr-export')[0].href;
// preview posts the pending set, shows the headline and the changed rows, saves nothing
get('lr-rune').value = '6'; byKey('lr-preview')[0].listeners.click[0](); await wait();
out.previewPost = posts[0]; out.previewHead = get('lr-preview-head').textContent; out.previewRows = get('lr-preview-list').children.map((n) => n.textContent); out.putsAfterPreview = puts.length;
out.previewLine = previewLine(PV.items[0]); out.previewNone = previewSummary({ changed: 0, asOfItems: 0, totals: {} });
// Starter apply: the four knobs and the preset, no rules
get('lr-rune').value = '8'; get('lr-gold').value = '40'; get('lr-sock').checked = true; get('lr-gear').value = 'never'; get('lr-preset').value = 'class:Amazon';
byKey('lr-apply')[0].listeners.click[0](); await wait();
out.applyPut = puts[0]; out.toasts = toasts.slice(); out.getsAfterSave = gets;
// a bad number never posts
get('lr-rune').value = '99'; byKey('lr-apply')[0].listeners.click[0](); await wait();
out.badNumber = { puts: puts.length, err: all(sheet).find((n) => n.className === 'd-lr-err').textContent };
// Custom: switch, move the first rule down, switch the second off, save
cur = view('custom'); doc.listeners['d2:lootrules-ask']({ detail: {} }); await wait();
out.customMode = api.mode;
const items = () => get('lr-rules').children;
out.rows = items().map((li) => li.children[1].textContent);
byKey('lr-down')[0].listeners.click[0]();
out.moved = items().map((li) => li.children[1].textContent);
const firstOn = items()[0].children[0]; firstOn.checked = false; firstOn.listeners.change[0]();
byKey('lr-save')[0].listeners.click[0](); await wait();
const saved = JSON.parse(puts[puts.length - 1]);
out.savedIds = saved.rules.map((r) => r.id + (r.off ? ':off' : '')); out.savedPreset = saved.preset; out.savedSeq = saved.seq;
// a refusal says where
reply = { ok: false, code: 'invalid', errors: [{ path: 'rules[3].when', msg: 'bad value' }, { msg: 'unexpected comma', line: 4, col: 2 }] };
byKey('lr-save')[0].listeners.click[0](); await wait();
out.refused = all(sheet).find((n) => n.className === 'd-lr-err').textContent;
reply = { ok: false, code: 'conflict' };
byKey('lr-save')[0].listeners.click[0](); await wait();
out.conflict = all(sheet).find((n) => n.className === 'd-lr-err').textContent;
out.refusals = [refusal(null), refusal({ j: { code: 'off' } }), refusal({ j: { error: 'boom' } })];
out.line = decisionLine({ scope: 'tidy', item: 'Ring', today: 'keep', rules: 'stash', rule: 't-rare-jewelry', why: 'Will stash: rare jewelry' });
// undo: enabled once the view says so, posts the version it looked at, reloads
cur = view('custom', { canUndo: true, seq: 5 }); doc.listeners['d2:lootrules-ask']({ detail: {} }); await wait();
out.undoOn = !byKey('lr-undo')[0].disabled;
byKey('lr-undo')[0].listeners.click[0](); await wait();
out.undoPost = posts[posts.length - 1]; out.undoToast = toasts[toasts.length - 1];
postReply = { ok: false, code: 'no-undo' }; byKey('lr-undo')[0].listeners.click[0](); await wait();
out.undoRefused = all(sheet).find((n) => n.className === 'd-lr-err').textContent;
// import: nothing is sent until a file is chosen; a big or bad file never posts; a good one posts with the version
postReply = { ok: true, seq: 6 }; const fi = byKey('lr-file')[0]; const before = posts.length;
byKey('lr-import')[0].listeners.click[0](); out.importIdle = posts.length === before;
fi.files = [{ size: 300000, text: () => Promise.resolve('{}') }]; fi.listeners.change[0](); await wait();
out.importBig = all(sheet).find((n) => n.className === 'd-lr-err').textContent; out.bigPosted = posts.length - before;
fi.files = [{ size: 5, text: () => Promise.resolve('{nope') }]; fi.listeners.change[0](); await wait();
out.importBad = all(sheet).find((n) => n.className === 'd-lr-err').textContent; out.badPosted = posts.length - before;
fi.files = [{ size: 30, text: () => Promise.resolve('{"schema":"lootrules/1","rules":[]}') }]; fi.listeners.change[0](); await wait();
out.importPost = posts[posts.length - 1]; out.importToast = toasts[toasts.length - 1];
// Escape closes
sheet.listeners.keydown[0]({ key: 'Escape', preventDefault() {} });
out.closed = sheet.hidden;
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
	t.Log(j)
	for _, want := range []string{
		`"registered":["lootrules"]`, `"startHidden":true`, `"noGetYet":true`, `"opened":true`, `"gets":1`, `"noPutOnOpen":true`, `"mode":"starter"`,
		`"knobs":{"gear":"rare","rune":"3","gold":"5","sock":false,"preset":"starter"}`,
		`"recent":["El Rune on the ground: picked up (today's code: left alone). Will pick up: rune [g-runes]","Cap at the vendor: sold. Will sell [t-sell]"]`,
		`"differs":["1",""]`,
		`"applyPut":"{\"schema\":\"lootrules/1\",\"seq\":1,\"name\":\"Starter\",\"limits\":{\"maxActions\":12},\"preset\":\"class:Amazon\",\"options\":{\"sellGear\":\"never\",\"minRune\":8,\"minGold\":40,\"keepSockets\":true},\"rules\":[]}"`,
		`"toasts":["good:Loot rules saved"]`, `"getsAfterSave":2`, `"badNumber":{"puts":1,"err":"Lowest rune must be from 1 to 33"}`,
		`"customMode":"custom"`, `"rows":["On the ground: stash when it is a rune","On the ground: keep gold","In the pack: sell junk"]`,
		`"moved":["On the ground: keep gold","On the ground: stash when it is a rune","In the pack: sell junk"]`,
		`"savedIds":["g-gold:off","g-runes","t-sell:off"]`, `"savedPreset":"custom"`,
		`"refused":"rules[3].when: bad value; line 4: unexpected comma"`, `"conflict":"The rules changed somewhere else. Reload and apply your change again."`,
		`"sellKinds":["Good gear: Kept.","Jewelry: Sold unidentified until you own a Tome of Identify."]`, `"undoOff":true`, `"exportHref":"/api/lootrules/export"`,
		`"previewPost":"/api/lootrules/preview {\"schema\":\"lootrules/1\",\"seq\":1,\"name\":\"Starter\",\"limits\":{\"maxActions\":12},\"preset\":\"starter\",\"options\":{\"sellGear\":\"rare\",\"minRune\":6,\"minGold\":5,\"keepSockets\":false},\"rules\":[]}"`,
		`"previewHead":"2 items would be handled differently. With this set the bot would sell 3, stash 2, drop 0 and keep 4 of the 9 items in the pack and stash now. Nothing was changed."`,
		`"previewRows":["Ring (pack): at the vendor: sold now, kept with this set. Kept: you own a Tome of Identify","The dry run does not know your marks."]`, `"putsAfterPreview":0`,
		`"previewNone":"The dry run changes nothing: every item gets the answer it gets now. With this set the bot would sell 0, stash 0, drop 0 and keep 0 of the 0 items in the pack and stash now. Nothing was changed."`,
		`"undoOn":true`, `"undoPost":"/api/lootrules/undo {\"seq\":5}"`, `"undoToast":"good:Went back to the earlier rules"`, `"undoRefused":"There is no earlier version to go back to."`,
		`"importIdle":true`, `"importBig":"That file is too large to be a rules file (the limit is 256 KB)."`, `"bigPosted":0`, `"importBad":"That file is not valid JSON."`, `"badPosted":0`,
		`"importPost":"/api/lootrules/import?seq=5 {\"schema\":\"lootrules/1\",\"rules\":[]}"`, `"importToast":"good:Rules imported. Undo goes back."`,
		`"line":"Ring in the pack: stashed (today's code: kept). Will stash: rare jewelry [t-rare-jewelry]"`, `"closed":true`,
	} {
		if !strings.Contains(j, want) {
			t.Errorf("missing %s", want)
		}
	}
}
