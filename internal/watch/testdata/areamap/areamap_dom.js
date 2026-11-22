// DOM goldens of the area-map slot panel (web/panels/area-map) in a stub DOM, and the geometry rule behind it: the drawing is
// scaled to fit its box whole, letterboxed, never cropped (the user saw a minimap cropped at 1914x930). Every state goes
// through the panel's own mount/update and is compared with areamap_<state>.txt (UPDATE=1 rewrites).
// Usage: node areamap_dom.js path/to/web/panels
const assert = require('assert'), fs = require('fs'), path = require('path'), vm = require('vm');
const { mk } = require('../panel_harness.js');
const dir = process.argv[2];

// A loader that also keeps the top-level consts (the harness one keeps functions only) and takes imported names as extras.
function load(file, extra) {
  const src = fs.readFileSync(file, 'utf8').replace(/^export /gm, '').replace(/^import [\s\S]*?;\n/gm, '');
  const names = [...src.matchAll(/^(?:function|const) (\w+)/gm)].map((m) => m[1]);
  const c = Object.assign({ Math, Object, JSON, Date, Uint8Array, encodeURIComponent, console, Array, String, Number, getComputedStyle: () => ({ getPropertyValue: () => '' }), typeof: undefined }, extra || {});
  c.window = c;
  vm.runInNewContext(src + '\n;this.__x = {' + names.join(',') + '};', c);
  return c.__x;
}
const M = load(path.join(dir, 'map', 'panel.js'));
const P = load(path.join(dir, 'area-map', 'panel.js'), M);

function node(tag, cls, text) {
  const e = mk(tag);
  e.className = cls || '';
  if (text !== undefined) e.textContent = text;
  const app = e.append;
  e.append = (...k) => app(...k.map((x) => (typeof x === 'string' ? { tag: '#text', textContent: x, children: [], dataset: {}, attrs: {} } : x)));
  return e;
}
// the canvas records what is drawn: the cell rectangles (fillRect with an area fill) and every stroke point
function recorder(rec) {
  return new Proxy({}, { get: (t, k) => {
    if (k === 'fillRect') return (x, y, w, h) => rec.rects.push([x, y, w, h]);
    if (k === 'lineTo' || k === 'moveTo') return (x, y) => rec.pts.push([x, y]);
    return k in t ? t[k] : () => ({ width: 1 });
  }, set: (t, k, v) => { t[k] = v; return true; } });
}
function make(store) {
  const rec = { rects: [], pts: [] }, body = node('div', 'd-slotbody'), meta = node('span', 'meta'), actions = node('span', 'd-btngroup'), root = node('div', 'd-panel');
  const c = { body, meta, actions, el: (tag, cls, text) => { const e = node(tag, cls, text); if (tag === 'canvas') { e.clientWidth = 300; e.clientHeight = 200; e.getContext = () => recorder(rec); } return e; },
    store: (k) => ({ get: () => (k in store ? store[k] : null), set: (v) => { store[k] = v; } }), hide() {}, state() {}, toast() {}, later() {}, token: 't' };
  root.append(body);
  return { root, c, rec, body };
}
function dump(e, ind) {
  if (e.tag === '#text') return e.textContent.trim() ? ind + JSON.stringify(e.textContent) + '\n' : '';
  const a = [];
  if (e.className) a.push('class=' + JSON.stringify(e.className));
  for (const k of Object.keys(e.dataset).sort()) a.push('data-' + k + '=' + JSON.stringify(e.dataset[k]));
  for (const k of Object.keys(e.attrs).sort()) a.push(k + '=' + JSON.stringify(e.attrs[k]));
  for (const k of ['title', 'type']) if (e[k]) a.push(k + '=' + JSON.stringify(e[k]));
  if (e.hidden) a.push('hidden');
  let s = ind + '<' + e.tag + (a.length ? ' ' + a.join(' ') : '') + '>' + (e.textContent ? ' ' + JSON.stringify(e.textContent) : '') + '\n';
  for (const k of e.children) s += dump(k, ind + '  ');
  return s;
}

// a wide area (100 x 20 subtiles) in a 300 x 200 box: the old square canvas cropped the width of such a room
const map = { areaId: 8, area: 'Cold Plains', version: 3, x0: 0, y0: 0, step: 1, w: 100, h: 20, rle: [1, 100 * 20], exits: [{ x: 98, y: 10 }], waypoints: [{ x: 2, y: 10 }] };
const cells = (n) => Array.from({ length: n }, (_, i) => [i * 3, 10, i % 5 === 0 ? 2 : 1]);
const base = { tiers_unlocked: [1, 2], version: 3, map, hero: { x: 40, y: 10 }, units: [],
  trail: [{ ago: 50000, x: 5, y: 10, a: 0 }, { ago: 30000, x: 30, y: 10, a: 1 }, { ago: 8000, x: 40, y: 10, a: 3 }, { ago: 500, x: 41, y: 10, a: 3 }], trail_area: 8,
  trail_legend: ['walk', 'combat', 'town', 'farm', 'flee', 'rest', 'loot'], events: [{ ago: 20000, x: 30, y: 10, k: 'kill' }], coverage: { area: 8, total: 20, cells: cells(20), cell: 4 } };
const states = {
  trail: [base],
  empty: [{ tiers_unlocked: [1], version: 0, waiting: true }],
  nodata: [null],
  locked: [{ tiers_unlocked: [] }],
};
let bad = 0;
for (const name of Object.keys(states)) {
  const { root, c, body } = make({});
  const api = P.mount(root, c);
  states[name].forEach((r) => api.update(r));
  const out = dump(root, '');
  const f = path.join(__dirname, 'areamap_' + name + '.txt');
  if (process.env.UPDATE === '1') { fs.writeFileSync(f, out); continue; }
  if (!fs.existsSync(f) || fs.readFileSync(f, 'utf8') !== out) { console.error('areamap_' + name + '.txt differs\n' + out); bad++; }
}
if (bad) process.exit(1);

// ---- the fit rule ----
{
  const u = P.worldBox(map, new Uint8Array(100 * 20).fill(1), [], []);
  for (const [w, h] of [[300, 200], [200, 300], [97, 400], [1200, 180], [40, 40]]) {
    const v = P.fitView(u, w, h, 0, null, { x: 0, y: 0 });
    const x0 = v.ox + u.x0 * v.k, x1 = v.ox + u.x1 * v.k, y0 = v.oy + u.y0 * v.k, y1 = v.oy + u.y1 * v.k;
    assert.ok(x0 >= 0 && x1 <= w + 1e-6 && y0 >= 0 && y1 <= h + 1e-6, 'the whole world is inside ' + w + 'x' + h + ': ' + [x0, x1, y0, y1]);
    assert.ok(Math.abs((x1 - x0) - w) < 0.05 * w + 1e-6 || Math.abs((y1 - y0) - h) < 0.05 * h + 1e-6, 'it fills one axis (letterboxed on the other)');
  }
  const { root, c, rec } = make({}), api = P.mount(root, c);
  api.update(base);
  assert.ok(rec.rects.length > 0, 'cells were drawn');
  for (const [x, y, w, h] of rec.rects) assert.ok(x >= 0 && y >= 0 && x + w <= 300 + 1 && y + h <= 200 + 1, 'a cell outside the 300x200 box: ' + [x, y, w, h]);
  // zoomed in around the hero: the world may leave the box (that is what zoom is), the hero stays at the centre
  const v = P.fitView(u, 300, 200, 2, { x: 40, y: 10 }, { x: 0, y: 0 });
  assert.ok(Math.abs(v.ox + 40.5 * v.k - 150) < v.k && Math.abs(v.oy + 10.5 * v.k - 100) < v.k, 'zoom centres on the hero');
}
// the margin of an area's grid is not part of the picture: a 100 x 30 grid with a 40 x 10 room is fitted as the room
{
  const m = { areaId: 9, area: 'Den', version: 1, x0: 5, y0: 7, step: 1, w: 100, h: 30 }, cl = new Uint8Array(100 * 30);
  for (let y = 10; y < 20; y++) for (let x = 30; x < 70; x++) cl[y * 100 + x] = 1;
  assert.strictEqual(JSON.stringify(P.bounds(m, cl)), JSON.stringify({ x0: 35, y0: 17, x1: 75, y1: 27 }), 'the box of the drawn cells, in the frame of the area');
  assert.strictEqual(JSON.stringify(P.worldBox(m, cl, [], [{ x: 4, y: 40 }])), JSON.stringify({ x0: 4, y0: 17, x1: 75, y1: 41 }), 'a hero or a trail point outside the room still fits');
  assert.strictEqual(JSON.stringify(P.bounds({ x0: 0, y0: 0, step: 1, w: 4, h: 4 }, new Uint8Array(16))), JSON.stringify({ x0: 0, y0: 0, x1: 4, y1: 4 }), 'an empty grid falls back to the whole grid');
}
// ---- label, share, query ----
assert.strictEqual(P.exploredPct({ total: 20, cells: cells(20) }), Math.round(100 * 16 / 20), 'seen cells over the total');
assert.strictEqual(P.exploredPct({ total: 200, seen: 50, cells: [] }), 25, 'the server count wins');
assert.strictEqual(P.exploredPct(null), -1);
assert.strictEqual(P.label('Cold Plains', 62), 'Cold Plains 62%');
assert.strictEqual(P.label('Cold Plains', -1), 'Cold Plains');
{
  const { root, c } = make({}), api = P.mount(root, c);
  assert.strictEqual(api.query(), '?v=0');
  api.update(base);
  assert.strictEqual(c.meta.textContent, 'Cold Plains 80%', 'the label is the zone and the explored share');
  assert.strictEqual(api.query(), '?v=3');
  api.update(null);
  assert.strictEqual(api.query(), '?v=3', 'a failed poll keeps what it holds');
}
{
  const st = { heat: 'kills' }, { root, c } = make(st), api = P.mount(root, c);
  assert.ok(/hist=1&heat=1/.test(api.query()), 'a stored heat layer asks for the cells at once: ' + api.query());
}
// ---- phone layout (card mapphone, bug 2026-11-22): the toolbar comes before the picture in the flow, the picture sits in its own
// box, and the CSS gives that box a height and puts nothing over the tab row (the pixel check is tools/uishots.sh --map) ----
{
  const { root, c, body } = make({}), api = P.mount(root, c), am = body.children[0];
  assert.deepStrictEqual(am.children.map((k) => k.className), ['ctl', 'layers', 'stage'], 'toolbar and Layers list first, then the picture box');
  const stage = am.children[2];
  assert.deepStrictEqual(stage.children.map((k) => k.className), ['cv', 'empty'], 'the canvas and its message share the box');
  api.update({ tiers_unlocked: [1, 2], waiting: true });
  assert.strictEqual(stage.children[1].textContent, 'Waiting for the first steps...', 'a hero not seen yet says why');
  assert.strictEqual(stage.children[1].hidden, false);
  const css = fs.readFileSync(path.join(dir, 'area-map', 'panel.css'), 'utf8'), ph = css.split('\n').filter((l) => l.includes('data-cmode="phone"'));
  const rule = (sel) => { const l = ph.find((x) => x.includes(sel + ' {')); assert.ok(l, 'a phone rule for ' + sel); return l; };
  assert.ok(/position: static/.test(rule('.ctl')), 'the toolbar is in the flow on a phone');
  assert.ok(/position: static/.test(rule('.layers')), 'the Layers list is in the flow on a phone');
  assert.ok(/aspect-ratio: 1/.test(rule('.stage')) && /min-block-size: 2[6-9]\dpx|min-block-size: [3-9]\d\dpx/.test(rule('.stage')), 'the picture box has a height of its own (over 200 px)');
}
console.log('areamap ok');
