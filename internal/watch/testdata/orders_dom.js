// DOM golden of the Orders panel (panels/orders, docs/ui/redesign-2026-11-14/C-orders.md): the five fixtures of
// testdata/commands (both classes) plus a view with eight rows, the lost connection and a route that is off; and the
// POST each cancel cross sends (an order, a queue entry, the loop, a repair, an item action, the gold farm).
// UPDATE=1 rewrites testdata/orders_dom.txt. Usage: node orders_dom.js path/to/panel.js
const assert = require('assert'), fs = require('fs'), path = require('path');
const { load, shellCtx } = require('./panel_harness.js');
const listeners = {};
const doc = { addEventListener(t, f) { (listeners[t] = listeners[t] || []).push(f); }, removeEventListener(t, f) { listeners[t] = (listeners[t] || []).filter((x) => x !== f); },
  dispatchEvent(e) { (listeners[e.type] || []).forEach((f) => f(e)); }, documentElement: { dataset: {} } };
const P = load(process.argv[2], { document: doc, CustomEvent: class { constructor(type, o) { this.type = type; this.detail = o && o.detail; } } });
const dir = path.join(__dirname, 'commands');
const files = fs.readdirSync(dir).filter((f) => /^view-.*\.json$/.test(f)).sort();
assert.ok(files.length >= 12, 'the handler goldens are missing: ' + files.length);
const view = (n) => JSON.parse(fs.readFileSync(path.join(dir, 'view-' + n + '.json'), 'utf8'));

function mount(post) {
  const { root, c } = shellCtx({}, {}, []);

  c.ui = { btn: (t, o) => { const b = c.el('button', 'd-btn', t); b.dataset = Object.assign({}, o); return b; }, chip: (t) => c.el('span', 'd-chip', t), clear(n) { n.children.length = 0; } };
  c.post = (p, b) => { post.push(p + ' ' + JSON.stringify(b)); return Promise.resolve(c.reply || { ok: true }); };
  const sheet = c.el('div', 'd-tilesheet');
  sheet.hidden = true;
  c.tileSheet = () => ({ el: sheet, place() {} });
  c.later = () => 0;
  c.toasts = [];
  c.toast = (s, t, a) => c.toasts.push(s + ': ' + t + (a ? ' [' + a.label + ' ' + a.ms + ']' : ''));
  return { c, inst: P.mount(root, c) };
}
// every element as tag.class, its data and text, indented; hidden ones in brackets
function dump(e, d) {
  const data = Object.keys(e.dataset || {}).sort().map((k) => k + '=' + e.dataset[k]).join(',');
  let s = '  '.repeat(d) + (e.hidden ? '[' : '') + e.tag + (e.className ? '.' + e.className.split(' ').join('.') : '') + (data ? ' {' + data + '}' : '') +
    (e.textContent ? ' ' + JSON.stringify(e.textContent) : '') + (e.title ? ' title=' + JSON.stringify(e.title) : '') + (e.disabled ? ' disabled' : '') + (e.hidden ? ']' : '') + '\n';
  for (const k of e.children) s += dump(k, d + 1);
  return s;
}

const cases = [];
for (const f of files) cases.push([f.replace(/^view-|\.json$/g, ''), view(f.replace(/^view-|\.json$/g, ''))]);
const eight = view('three-one-failed-sorceress');
for (let i = 0; i < 5; i++) eight.rows.splice(1, 0, { id: 'order:' + (20 + i), intent: 'go', text: 'Farm Blood Moor for ' + (10 + i) + ' min', source: 'map', state: 'queued',
  why: { code: 'order_queued', text: 'after the running order' }, can: { cancel: true, retry: false }, cancel: { path: '/api/orders/cancel', body: { id: 20 + i } } });
cases.push(['eight-rows', eight]);

let out = '';
for (const [name, v] of cases) {
  const { c, inst } = mount([]);
  inst.update(v);
  out += '== ' + name + (c.hidden ? ' (panel hidden)' : '') + '\n' + dump(c.body, 0);
}
for (const [name, status] of [['no connection', 0], ['route off', 404]]) {
  const { c, inst } = mount([]);
  inst.update(null, { status, fails: 3 });
  out += '== ' + name + (c.hidden ? ' (panel hidden)' : '') + '\n' + dump(c.body, 0);
}

// the helpers
assert.strictEqual(P.mark({ state: 'active' }, 1), '>');
assert.strictEqual(P.mark({ state: 'failed' }, 1), '!');
assert.strictEqual(P.mark({ state: 'waiting' }, 2), '2');
assert.strictEqual(P.reason({ phase: 'going' }), 'going');
assert.strictEqual(P.openCount([{ state: 'active' }, { state: 'failed' }, { state: 'queued' }]), 2);

// cancel: every cross posts the route its row names, and the routes are the existing ones
const posts = [];
const { c, inst } = mount(posts);
const mixed = view('cancel-all-sorceress');
inst.update(mixed);
const crosses = [];
(function walk(e) { if (e.className === 'o-x' && e.tag === 'button') crosses.push(e); e.children.forEach(walk); })(c.body);
const wantRows = mixed.rows.filter((r) => r.can.cancel);
assert.strictEqual(crosses.length, wantRows.length, 'one cross per cancellable row');
assert.ok(wantRows.length >= 6, 'the mixed fixture has a row of every source');
crosses.forEach((x) => x.fire('click'));
setImmediate(() => {
  assert.deepStrictEqual(posts.slice().sort(), [
    '/api/inventory/cancel {"id":7,"kind":"equip"}',
    '/api/inventory/sets/cancel {"group":"g1"}',
    '/api/orderloop {"action":"stop"}',
    '/api/orderqueue {"action":"cancel","id":3}',
    '/api/orders/cancel {"id":12}',
    '/api/repair/cancel {}',
    '/api/farmgold {"on":false}',
  ].sort(), 'each cancel reaches the route of its own source');
  // a row that cannot be cancelled has no cross: the item being done, a failed row
  const none = view('three-one-failed-sorceress');
  const { c: c2, inst: i2 } = mount([]);
  i2.update(none);
  let xs = 0;
  (function walk(e) { if (e.className === 'o-x' && e.tag === 'button') xs++; e.children.forEach(walk); })(c2.body);
  assert.strictEqual(xs, none.rows.filter((r) => r.can.cancel).length);
  // the verb row is live: every button posts one command to the one door; the toast carries Undo for five seconds
  const vp = [], vm = mount(vp), vv = view('three-one-failed-sorceress');
  vm.c.reply = { ok: true, order: { text: 'Fight here for 20 min' }, undo: { path: '/api/commands/cancel', body: { id: 'order:9' } } };
  vm.inst.update(vv);
  const vb = {};
  (function walk(e) { if (e.tag === 'button' && e.dataset && e.dataset.key && e.dataset.key.startsWith('verb-')) vb[e.dataset.intent] = e; e.children.forEach(walk); })(vm.c.body);
  assert.deepStrictEqual(Object.keys(vb).sort(), ['go', 'more', 'pause', 'repair', 'town'].sort(), 'the five verbs');
  vb.town.fire('click'); vb.repair.fire('click'); vb.pause.fire('click');
  let gotGo = 0; doc.addEventListener('d2:goto', () => gotGo++);
  vb.go.fire('click');
  setImmediate(() => {
    assert.deepStrictEqual(vp, ['/api/commands {"intent":"town","args":{}}', '/api/commands {"intent":"repair","args":{}}', '/api/commands {"intent":"pause","args":{}}']);
    assert.strictEqual(gotGo, 1, 'Go to asks the zone map');
    assert.ok(vm.c.toasts.every((t) => /\[Undo 5000\]$/.test(t)), 'each accepted command shows Undo for five seconds: ' + vm.c.toasts);
    // z undoes the last one; keys g t r p work, and are off while typing
    vp.length = 0;
    doc.dispatchEvent({ type: 'keydown', key: 'z', target: {}, preventDefault() {} });
    doc.dispatchEvent({ type: 'keydown', key: 't', target: { tagName: 'INPUT' }, preventDefault() {} });
    doc.dispatchEvent({ type: 'keydown', key: 't', target: {}, preventDefault() {} });
    setImmediate(() => {
      assert.deepStrictEqual(vp, ['/api/commands/cancel {"id":"order:9"}', '/api/commands {"intent":"town","args":{}}']);
      // a refused command says why in words, and the paused view reads Resume
      const rp = [], rm = mount(rp);
      rm.c.reply = { ok: false, code: 'locked', error: 'not unlocked yet' };
      rm.inst.update(vv);
      let tb; (function walk(e) { if (e.dataset && e.dataset.key === 'verb-town') tb = e; e.children.forEach(walk); })(rm.c.body);
      tb.fire('click');
      setImmediate(() => {
        assert.ok(rm.c.toasts.some((t) => /warn: not unlocked yet/.test(t)), 'refusal in words: ' + rm.c.toasts);
        finish();
      });
    });
  });
});
function finish() {
  const golden = path.join(__dirname, 'orders_dom.txt');
  if (process.env.UPDATE) fs.writeFileSync(golden, out);
  assert.strictEqual(out, fs.readFileSync(golden, 'utf8'));
  console.log('ok');
}
