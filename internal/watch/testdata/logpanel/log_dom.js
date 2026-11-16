// DOM goldens of the log drawer's panel (web/panels/log) per tab, and the pure helpers. Each state goes through the panel's own
// mount/update and is compared with logpanel_<state>.txt (UPDATE=1 rewrites). The time of a row is the viewer's own zone, so the
// goldens print it as "<time>" and the helper is checked on its own with a fixed zone.
// Usage: node log_dom.js path/to/web/panels
process.env.TZ = 'UTC';
const assert = require('assert'), fs = require('fs'), path = require('path'), vm = require('vm');
const { mk } = require('../panel_harness.js');
const src = fs.readFileSync(path.join(process.argv[2], 'log', 'panel.js'), 'utf8').replace(/^export /gm, '');
const names = [...src.matchAll(/^(?:function|const) (\w+)/gm)].map((m) => m[1]);
const c0 = { Math, Object, JSON, Date, Array, String };
vm.runInNewContext(src + '\n;this.__x = {' + names.join(',') + '};', c0);
const P = c0.__x;

function node(tag, cls, text) {
  const e = mk(tag);
  e.className = cls || '';
  if (text !== undefined) e.textContent = text;
  e.replaceChildren = (...k) => { e.children = k; };
  e.focus = () => {};
  return e;
}
function make(store) {
  const body = node('div', 'd-slotbody'), root = node('div', 'd-panel'), calls = { refresh: 0 };
  const c = { body, meta: node('span'), actions: node('span'), el: node, store: (k) => ({ get: () => (k in store ? store[k] : null), set: (v) => { store[k] = v; } }),
    hide() {}, state() {}, toast() {}, later() {}, refresh() { calls.refresh++; } };
  root.append(body);
  return { root, c, calls };
}
function dump(e, ind) {
  const a = [];
  if (e.className) a.push('class=' + JSON.stringify(e.className));
  for (const k of Object.keys(e.dataset).sort()) a.push('data-' + k + '=' + JSON.stringify(e.dataset[k]));
  for (const k of Object.keys(e.attrs).sort()) a.push(k + '=' + JSON.stringify(e.attrs[k]));
  if (e.tabIndex !== undefined) a.push('tabindex=' + e.tabIndex);
  if (e.hidden) a.push('hidden');
  const text = /(^| )when( |$)/.test(e.className) ? '<time>' : e.textContent;
  let s = ind + '<' + e.tag + (a.length ? ' ' + a.join(' ') : '') + '>' + (text ? ' ' + JSON.stringify(text) : '') + '\n';
  for (const k of e.children) s += dump(k, ind + '  ');
  return s;
}
const row = (id, ago, kind, text, level, quality) => Object.assign({ id, ago, kind, text }, level ? { level } : {}, quality ? { quality } : {});
const replies = {
  items: { tab: 'items', rows: [row('d2', 8, 'drop', 'picked up Stormshield (Blood Moor)', '', 'unique'), row('d1', 70, 'drop', 'picked up Short Staff (Cold Plains)', '', 'normal')] },
  now: { tab: 'now', rows: [row('f2', 5, 'order', 'Order changed: farm Cold Plains.'), row('f1', 30, 'story', 'A quiet moor.')] },
  journal: { tab: 'journal', rows: [row('j2', 20, 'learned', 'It will remember that.'), row('j1', 90, 'answer', 'You said yes to the Den of Evil.')] },
  story: { tab: 'story', rows: [], note: 'The story opens at level 5.' },
  alerts: { tab: 'alerts', rows: [row('a2', 40, 'life', 'Life under 25% in Blood Moor.', 'alarm'), row('a1', 400, 'gear', 'Short Staff is broken.', 'warn')] },
  empty: { tab: 'now', rows: [] },
};
let bad = 0;
for (const name of Object.keys(replies)) {
  const store = { tab: replies[name].tab }, { root, c } = make(store), api = P.mount(root, c);
  assert.strictEqual(api.query(), '?tab=' + replies[name].tab, 'the tab rides the poll');
  api.update(replies[name]);
  const out = dump(root, '');
  const f = path.join(__dirname, 'logpanel_' + name + '.txt');
  if (process.env.UPDATE === '1') { fs.writeFileSync(f, out); continue; }
  if (!fs.existsSync(f) || fs.readFileSync(f, 'utf8') !== out) { console.error('logpanel_' + name + '.txt differs\n' + out); bad++; }
}
{ // the error state: nothing held yet
  const { root, c } = make({}), api = P.mount(root, c);
  api.update(null);
  const out = dump(root, ''), f = path.join(__dirname, 'logpanel_error.txt');
  if (process.env.UPDATE === '1') fs.writeFileSync(f, out); else if (!fs.existsSync(f) || fs.readFileSync(f, 'utf8') !== out) { console.error('logpanel_error.txt differs\n' + out); bad++; }
}
if (bad) process.exit(1);

// picking a tab: the next poll asks for it at once, a reply for the tab just left is ignored
{
  const store = {}, { root, c, calls } = make(store), api = P.mount(root, c);
  assert.strictEqual(api.query(), '?tab=items', 'first tab is Items');
  const bar = root.children[0].children[0].children[0];
  bar.children[3].fire('click');
  assert.strictEqual(api.query(), '?tab=story');
  assert.strictEqual(calls.refresh, 1, 'asks the shell to poll now');
  assert.strictEqual(store.tab, 'story', 'remembered per viewer');
  api.update(replies.now);
  assert.strictEqual(root.children[0].children[0].children[1].children.length, 0, 'a late reply of the old tab is not drawn');
  api.update(replies.story);
  assert.ok(/opens at level 5/.test(root.children[0].children[0].children[2].textContent));
  bar.children[3].fire('keydown', { key: 'ArrowRight' });
  assert.strictEqual(api.query(), '?tab=alerts', 'arrow keys move between tabs');
}
// times are the viewer's: 20 s ago at 14:00:30 UTC is 14:00:10; an old row carries its date
assert.strictEqual(P.when(20, Date.UTC(2026, 10, 14, 14, 0, 30), 'en-US'), '02:00:10 PM');
assert.strictEqual(P.when(86400, Date.UTC(2026, 10, 14, 14, 0, 30), 'en-US'), 'Nov 13 02:00:30 PM');
console.log('logpanel ok');
