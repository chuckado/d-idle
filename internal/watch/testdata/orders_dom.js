// DOM golden of the Orders panel (panels/orders, docs/ui/redesign-2026-11-14/C-orders.md): the five fixtures of
// testdata/commands (both classes) plus a view with eight rows, the lost connection and a route that is off; and the
// POST each cancel cross sends (an order, a queue entry, the loop, a repair, an item action, the gold farm).
// UPDATE=1 rewrites testdata/orders_dom.txt. Usage: node orders_dom.js path/to/panel.js
const assert = require('assert'), fs = require('fs'), path = require('path');
const { load, shellCtx } = require('./panel_harness.js');
const P = load(process.argv[2]);
const dir = path.join(__dirname, 'commands');
const files = fs.readdirSync(dir).filter((f) => /^view-.*\.json$/.test(f)).sort();
assert.ok(files.length >= 12, 'the handler goldens are missing: ' + files.length);
const view = (n) => JSON.parse(fs.readFileSync(path.join(dir, 'view-' + n + '.json'), 'utf8'));

function mount(post) {
  const { root, c } = shellCtx({}, {}, []);

  c.ui = { btn: (t, o) => { const b = c.el('button', 'd-btn', t); b.dataset = Object.assign({}, o); return b; }, chip: (t) => c.el('span', 'd-chip', t), clear(n) { n.children.length = 0; } };
  c.post = (p, b) => { post.push(p + ' ' + JSON.stringify(b)); return Promise.resolve({ ok: true }); };
  c.toasts = [];
  c.toast = (s, t) => c.toasts.push(s + ': ' + t);
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
  const golden = path.join(__dirname, 'orders_dom.txt');
  if (process.env.UPDATE) fs.writeFileSync(golden, out);
  assert.strictEqual(out, fs.readFileSync(golden, 'utf8'));
  console.log('ok');
});
