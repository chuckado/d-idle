package watch

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCharactersRulesGolden(t *testing.T) {
	api := NewCharactersAPI() // no desk bound: the rules are static
	w := httptest.NewRecorder()
	api.RulesHandler().ServeHTTP(w, httptest.NewRequest("GET", CharactersRulesRoute, nil))
	if w.Code != 200 {
		t.Fatalf("GET %d %s", w.Code, w.Body)
	}
	charGolden(t, "rules", w.Body.String())
}

// The page, its script, styles and texts are served by the view handler, and the texts name every refusal code the
// backend can send (a code without a text would show the server's raw sentence).
func TestCharactersPageServedAndTexts(t *testing.T) {
	h := (&Hub{}).ViewHandler()
	for path, ctype := range map[string]string{
		"/watch/characters": "text/html", "/watch/characters.js": "text/javascript", "/watch/characters.css": "text/css", "/watch/characters.json": "application/json",
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), ctype) || w.Body.Len() == 0 {
			t.Errorf("%s: %d %q", path, w.Code, w.Header().Get("Content-Type"))
		}
	}
	var tx struct {
		Refusals map[string]string `json:"refusals"`
	}
	b, err := webFS.ReadFile("web/characters.json")
	if err != nil || json.Unmarshal(b, &tx) != nil {
		t.Fatal("characters.json unreadable")
	}
	re := regexp.MustCompile(`no\((?:http\.Status\w+|\d+), "([a-z_]+)"`)
	files, _ := filepath.Glob("../charsel/*.go")
	n := 0
	for _, f := range files {
		src, _ := os.ReadFile(f)
		for _, m := range re.FindAllStringSubmatch(string(src), -1) {
			n++
			if tx.Refusals[m[1]] == "" {
				t.Errorf("no text for the refusal code %q", m[1])
			}
		}
	}
	if n < 15 {
		t.Errorf("found only %d codes: the scan is wrong", n)
	}
	for file, want := range map[string]string{"web/nav.json": "characters-open", "web/panels/orders/panel.js": "/watch/characters"} { // the two entry points
		src, err := webFS.ReadFile(file)
		if err != nil || !strings.Contains(string(src), want) {
			t.Errorf("%s lacks the Characters entry %q", file, want)
		}
	}
	for _, c := range []string{"off", "origin", "token", "body"} { // the route gates
		if tx.Refusals[c] == "" {
			t.Errorf("no text for %q", c)
		}
	}
}

// characters.js in a stub DOM (node, skipped without it): the list, the confirm sheets, the typed name, the create rules,
// the read-back after a press, refusals that are never drawn as success, and the error states.
func TestCharactersUIJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	dir := t.TempDir()
	for _, n := range []string{"characters.js", "characters.json"} {
		b, err := webFS.ReadFile("web/" + n)
		if err != nil {
			t.Fatal(err)
		}
		name := n
		if n == "characters.js" {
			name = "characters.mjs"
		}
		os.WriteFile(filepath.Join(dir, name), b, 0o600)
	}
	rules, _ := os.ReadFile("testdata/characters/rules.json")
	os.WriteFile(filepath.Join(dir, "rules.json"), rules, 0o600)
	list, _ := os.ReadFile("testdata/characters/list-two-classes.json")
	os.WriteFile(filepath.Join(dir, "list.json"), list, 0o600)
	h := filepath.Join(dir, "h.mjs")
	os.WriteFile(h, []byte(charactersHarness), 0o600)
	b, err := exec.Command(node, h).CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, b)
	}
	var out map[string]any
	if err := json.Unmarshal(b[strings.Index(string(b), "{"):], &out); err != nil {
		t.Fatalf("%v\n%s", err, b)
	}
	if fails, _ := out["fails"].([]any); len(fails) > 0 {
		t.Fatalf("checks failed: %v", fails)
	}
}

const charactersHarness = `
import { readFileSync } from 'fs';
const here = (f) => new URL(f, import.meta.url);
import { mountCharacters, checkName, when, strip, fmt, refusalText } from './characters.mjs';
const T = JSON.parse(readFileSync(here('./characters.json'), 'utf8')), RULES = JSON.parse(readFileSync(here('./rules.json'), 'utf8')), LIST = JSON.parse(readFileSync(here('./list.json'), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const mk = (tag) => { const n = { tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, value: '', checked: false, id: '',
  classList: { add(c) { n.className += ' ' + c; } },
  append(...c) { n.children.push(...c); }, appendChild(c) { n.children.push(c); return c; }, replaceChildren(...c) { n.children = c; }, setAttribute(k, v) { n.attrs[k] = v; },
  addEventListener(e, f) { (n.listeners[e] = n.listeners[e] || []).push(f); }, focus() { doc.activeElement = n; }, contains() { return false; }, click() { return fire(n, 'click'); } };
  return n; };
const doc = { activeElement: null, body: mk('body'), createElement: mk, addEventListener() {}, listeners: {} };
const fire = async (n, e, arg) => { for (const f of n.listeners[e] || []) await f(arg || { key: '', preventDefault() {}, target: n }); await tick(); };
const tick = () => new Promise((r) => setTimeout(r, 5));
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const find = (k) => all(doc.body).find((n) => n.dataset && n.dataset.key === k);
const texts = () => all(doc.body).filter((n) => !n.hidden).map((n) => n.textContent).filter(Boolean);
const has = (s) => texts().some((t) => t.includes(s));

// pure parts
ok(checkName('', RULES, [], T).level === 'hint', 'empty name is a hint');
ok(checkName('A', RULES, [], T).text.startsWith('Too short'), 'short');
ok(checkName('Abcdefghijklmnop', RULES, [], T).text.startsWith('Too long'), 'long');
ok(checkName('Ab1', RULES, [], T).text === T.nameChars, 'chars');
ok(checkName('-Ab', RULES, [], T).text === T.nameSep && checkName('A-b-c', RULES, [], T).text === T.nameSep, 'separators');
ok(checkName('A-b', RULES, [], T).ok, 'one separator is fine');
ok(checkName('zapp', RULES, LIST.characters, T).text.includes('already exists'), 'duplicate ignores case');
ok(when('2026-11-15T09:00:00Z', Date.parse('2026-11-15T12:00:00Z')) === '3 h ago' && when('', 0) === '', 'when');
ok(strip(null, { to: 'Zapp', at: 0 }, 1000, T).text.startsWith('Starting Zapp'), 'strip waits for the bot');
ok(strip(null, { to: 'Zapp', at: 0 }, 5 * 60 * 1000, T).failed, 'strip gives up');
ok(refusalText({ code: 'exists', error: 'raw' }, T) === T.refusals.exists && refusalText({ code: 'zzz', error: 'raw' }, T) === 'raw' && refusalText(null, T) === T.noAnswer, 'refusal text');

// the page against a fake bot
let view = JSON.parse(JSON.stringify(LIST)), posts = [], reply = null, getFail = null, rules = RULES, clock = Date.parse('2026-11-15T12:00:00Z');
const get = async (p) => { if (p === '/api/characters/rules') return rules; return getFail || view; };
const post = async (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return typeof reply === 'function' ? reply(p, b) : reply; };
const later = (f) => { later.q.push(f); }; later.q = [];
const page = mountCharacters({ doc, get, post, later, texts: T, now: () => clock });
await page.start(); await tick();

ok(all(doc.body).filter((n) => n.className === 'cs-card' || n.className.startsWith('cs-card')).length === 3, 'two heroes and the new tile');
ok(!!find('badge-active') && !find('switch-Holy') && !find('switch-Zapp').disabled, 'active badge, no switch on the active hero');
ok(has('Level 24 Paladin') && has('Playing now: Holy') && has('Nothing deleted'), 'card text');

// switch: the sheet words, then the press, then the strip from the read-back
await find('switch-Zapp').click();
const sheet = all(doc.body).find((n) => n.id === 'cs-sheet');
ok(!sheet.hidden && has('Leave the game now and start Zapp? The current hero is saved.') && doc.activeElement === find('sheet-go'), 'switch sheet words and focus');
reply = { ok: false, code: 'no_supervisor', error: 'raw' };
await find('sheet-go').click();
ok(!sheet.hidden && has(T.refusals.no_supervisor) && !has('Switching to Zapp'), 'a refusal stays in the sheet and is not a success');
reply = { schema: 'watch.characters.result/1', action: 'switch', name: 'Zapp' };
view = Object.assign(JSON.parse(JSON.stringify(LIST)), { switch: { from: 'Holy', to: 'Zapp', since: 'x' } });
await find('sheet-go').click();
ok(sheet.hidden && !find('strip').hidden && find('strip').textContent.startsWith('Switching to Zapp'), 'strip while switching');
ok(find('switch-Zapp').disabled && find('delete-Zapp').disabled, 'no second press while switching');
ok(later.q.length > 0, 'the strip polls');
view = Object.assign(JSON.parse(JSON.stringify(LIST)), { active: 'Zapp' }); view.characters[0].active = false; view.characters[1].active = true;
while (later.q.length) later.q.shift()(); await tick(); await tick();
ok(find('strip').hidden && has('Now playing Zapp.'), 'strip gone, done note from the read-back');
getFail = null; view = JSON.parse(JSON.stringify(LIST)); await page.refresh();

// delete: the active hero is refused with the reason, no API call
posts = [];
await find('delete-Holy').click();
ok(has('The hero in the game is never deleted') && !find('sheet-go') && posts.length === 0, 'active hero: reason, no go');
await find('sheet-cancel').click();
ok(sheet.hidden && doc.activeElement === find('delete-Holy'), 'focus returns to the opener');
// delete: the typed name
await find('delete-Zapp').click();
const typed = all(doc.body).find((n) => n.id === 'cs-typed');
ok(has('characters-archive/Zapp/') && find('sheet-go').disabled && doc.activeElement === typed, 'type-the-name sheet starts off');
typed.value = 'zapp'; await fire(typed, 'input'); ok(find('sheet-go').disabled, 'wrong case keeps it off');
typed.value = 'Zapp'; await fire(typed, 'input'); ok(!find('sheet-go').disabled, 'exact name turns it on');
reply = { ok: false, code: 'confirm_mismatch', error: 'raw' };
await find('sheet-go').click();
ok(!sheet.hidden && has(T.refusals.confirm_mismatch) && !has('is deleted'), 'refused delete is no success');
reply = { schema: 'watch.characters.result/1', action: 'delete', name: 'Zapp', archive: 'Zapp@20261115T120000Z' };
view = JSON.parse(JSON.stringify(LIST)); view.characters.pop(); view.deleted = [{ id: 'Zapp@20261115T120000Z', name: 'Zapp', class: 'sorceress', level: 9, archivedAt: '2026-11-15T12:00:00Z' }];
await find('sheet-go').click();
ok(sheet.hidden && has('Zapp is deleted. Archived as Zapp@20261115T120000Z') && !!find('deleted-Zapp@20261115T120000Z') && has('Level 9 Sorceress, deleted just now'), 'deleted and listed under Deleted');
ok(posts[posts.length - 1] === '/api/characters/delete {"name":"Zapp","confirm":"Zapp"}', 'delete body');
// the bot says done but the list disagrees
view.characters.push(LIST.characters[1]); await page.refresh();
reply = { schema: 'watch.characters.result/1', action: 'delete', name: 'Zapp' };
await find('delete-Zapp').click(); const t2 = all(doc.body).find((n) => n.id === 'cs-typed'); t2.value = 'Zapp'; await fire(t2, 'input'); await find('sheet-go').click();
ok(has('is still listed'), 'read-back catches a delete that did not happen');
// restore
posts = [];
view = JSON.parse(JSON.stringify(LIST)); view.characters.pop(); view.deleted = [{ id: 'Zapp@20261115T120000Z', name: 'Zapp', class: 'sorceress', level: 9, archivedAt: '2026-11-15T12:00:00Z' }];
await page.refresh();
await find('restore-Zapp@20261115T120000Z').click();
ok(has('Restore Zapp?') && has('It is not started'), 'restore sheet');
reply = { schema: 'watch.characters.result/1', action: 'restore', name: 'Zapp' }; view = JSON.parse(JSON.stringify(LIST));
await find('sheet-go').click();
ok(posts[0] === '/api/characters/restore {"archive":"Zapp@20261115T120000Z"}' && has('Zapp is back in the list'), 'restore posted and read back');

// create: the rules come from the API, the class chips from its table, hardcore is a note
posts = [];
await find('new-top').click();
const nm = all(doc.body).find((n) => n.id === 'cs-name');
const radios = all(doc.body).filter((n) => n.type === 'radio'), go = () => find('sheet-go');
ok(radios.length === RULES.classes.length && radios.length === 7 && has('Hardcore is not available yet.') && doc.activeElement === nm, 'class chips from the table, hardcore note, focus');
ok(go().disabled, 'create starts off');
nm.value = 'Zapp'; await fire(nm, 'input'); ok(go().disabled && has('already exists'), 'duplicate named in words');
nm.value = 'Newb'; await fire(nm, 'input'); ok(go().disabled && has('Pick a class.'), 'class needed');
radios[3].checked = true; await fire(radios[3], 'change');
ok(!go().disabled && go().textContent === 'Create Newb', 'valid name and class');
reply = { schema: 'watch.characters.result/1', action: 'create', name: 'Newb' };
view = Object.assign(JSON.parse(JSON.stringify(LIST)), { switch: { from: 'Holy', to: 'Newb', creating: true, since: 'x' } });
await go().click();
ok(posts[0] === '/api/characters {"name":"Newb","class":"paladin"}' && find('strip').textContent.startsWith('Creating Newb'), 'create body from the table id and the creating strip');
// Escape closes the top sheet
view = JSON.parse(JSON.stringify(LIST)); page.state.pend = null; await page.refresh();
await find('new-top').click(); const sh2 = all(doc.body).find((n) => n.id === 'cs-sheet');
await fire(sh2, 'keydown', { key: 'Escape', preventDefault() {} });
ok(sh2.hidden && doc.activeElement === find('new-top'), 'Escape closes and returns focus');
// no create when the backend cannot
view = Object.assign(JSON.parse(JSON.stringify(LIST)), { caps: { create: false, switch: false, delete: true, restore: true, why: 'a measured run is in progress (D2BOT_RUN)' } });
await page.refresh();
ok(find('new-top').disabled && find('switch-Zapp').disabled && has('Switching is off: a measured run is in progress (D2BOT_RUN).'), 'caps.why in words, buttons off');
// error states
getFail = { code: 'off', error: 'characters are not available' }; view = null; page.state.view = null; await page.refresh();
ok(has('Character management needs the bot under its own supervisor') && has('docs/charselect.md') && !find('retry'), 'off state names the doc, no retry');
getFail = { code: 'no_save_dir', error: 'x' }; await page.refresh(); ok(has('No saves yet') && !!find('retry'), 'no save folder');
getFail = null; view = { schema: 'watch.characters/1', active: '', caps: { create: true, switch: true, delete: true, restore: true }, characters: [], deleted: [] };
await page.refresh(); ok(has('No saved heroes') && !!find('new-tile-btn'), 'empty list');
getFail = { }; page.state.view = null; view = null; getFail = null; const g2 = get; 
console.log(JSON.stringify({ fails }));
`
