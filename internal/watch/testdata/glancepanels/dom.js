// DOM goldens and behaviour tests of the cockpit's Needs you panel and the strip (progress, journey-strip, health), in a
// stub DOM: each state is rendered through the panel's own mount/update and compared with <panel>_<state>.txt in this
// directory (UPDATE=1 rewrites). Usage: node dom.js path/to/web/panels
const assert = require('assert'), fs = require('fs'), path = require('path'), { load, mk } = require('../panel_harness.js');
const dir = process.argv[2];

function node(tag, cls, text) {
  const e = mk(tag);
  e.className = cls || '';
  let own = '';
  Object.defineProperty(e, 'textContent', { get: () => own, set: (v) => { own = v; e.children.length = 0; }, enumerable: true });
  if (text !== undefined) e.textContent = text;
  const app = e.append;
  e.append = (...k) => app(...k.map((x) => (typeof x === 'string' ? { tag: '#text', textContent: x, children: [], dataset: {}, attrs: {} } : x)));
  Object.defineProperty(e, 'lastChild', { get: () => e.children[e.children.length - 1] });
  e.togglePopover = () => { e.popOpen = !e.popOpen; };
  return e;
}
const ui = { clear(n) { n.children.length = 0; } };
const doc = { createElementNS: (ns, t) => node(t) };
function ctxFor(posts, store) {
  const root = node('div', 'd-panel'), body = node('div', 'd-slotbody');
  root.ownerDocument = doc;
  root.append(body);
  const c = { body, el: node, ui, hidden: false, st: '', toasts: [],
    hide(on) { c.hidden = !!on; root.hidden = !!on; }, state(s) { c.st = s; }, toast(sev, t) { c.toasts.push([sev, t]); },
    store: (k) => ({ get: () => (k in store ? store[k] : null), set: (v) => { store[k] = v; } }),
    post(p, b) { posts.push([p, b]); return Promise.resolve({ ok: true }); } };
  return { root, c };
}
function dump(e, ind) {
  if (e.tag === '#text') return e.textContent.trim() ? ind + JSON.stringify(e.textContent) + '\n' : '';
  const a = [];
  if (e.className) a.push('class=' + JSON.stringify(e.className));
  for (const k of Object.keys(e.dataset).sort()) a.push('data-' + k + '=' + JSON.stringify(e.dataset[k]));
  for (const k of Object.keys(e.attrs).sort()) a.push(k + '=' + JSON.stringify(e.attrs[k]));
  if (e.title) a.push('title=' + JSON.stringify(e.title));
  if (e.type) a.push('type=' + JSON.stringify(e.type));
  if (e.hidden) a.push('hidden');
  let s = ind + '<' + e.tag + (a.length ? ' ' + a.join(' ') : '') + '>' + (e.textContent ? ' ' + JSON.stringify(e.textContent) : '') + '\n';
  for (const k of e.children) s += dump(k, ind + '  ');
  return s;
}
let bad = 0;
function golden(id, st, got, c, hides) {
  const f = path.join(__dirname, id + '_' + st + '.txt');
  if (process.env.UPDATE === '1') { fs.writeFileSync(f, got); return; }
  if (!fs.existsSync(f)) { console.error('missing golden ' + f); bad++; return; }
  if (fs.readFileSync(f, 'utf8') !== got) { console.error('golden differs: ' + f + '\n' + got); bad++; }
  assert.strictEqual(c.hidden, !!hides, id + ' ' + st + ' hidden');
}
const panel = (id) => load(path.join(dir, id, 'panel.js'), { Notification: undefined });
function render(id, data, store) {
  const P = panel(id), posts = [], { root, c } = ctxFor(posts, store || {});
  const api = P.mount(root, c);
  api.update(data);
  return { P, root, c, api, posts };
}

const need = (id, sev, text, sub, actions) => ({ id, kind: 'k', sev, text, sub, actions: actions || [] });
const verb = (label, post, body) => ({ label, post, body });
const gear = need('gear', 'danger', 'Weapon broken', 'repair is about 180 gold', [verb('Repair now', '/api/repair', { why: 'gear' })]);
const raven = need('raven', 'warn', 'Go for Blood Raven at level 9?', '', [verb('Go now', '/api/orders', { verb: 'go', zone: 'Cold Plains' }), verb('Not yet', '/api/journey/hold')]);
const belt = need('belt', 'info', 'Belt is empty', 'buy potions in town', [verb('Buy', '/api/buy')]);
const merc = need('merc', 'warn', 'Mercenary is dead', '', [verb('Revive', '/api/merc/revive')]);
const stash = need('stash', 'info', 'Stash is full', '', []);
const glance = (o) => Object.assign({ schema: 'watch.glance/1', needs: [], health: {}, progress: {} }, o);

// ---- needs-you
const NY = {
  empty: [], 'one-red': [gear], 'red-amber': [raven, gear], 'five-rows': [gear, raven, belt, merc, stash],
};
for (const [st, needs] of Object.entries(NY)) {
  const r = render('needs-you', glance({ needs }));
  golden('needs-you', st, dump(r.root, ''), r.c, needs.length === 0);
}
{ // severity order, the overflow line and the cap of the visible rows
  const r = render('needs-you', glance({ needs: NY['five-rows'] }));
  const rows = r.root.children[0].children[1].children;
  assert.strictEqual(rows.length, 3, 'three rows');
  assert.deepStrictEqual(rows.map((x) => x.dataset.id), ['gear', 'raven', 'merc'], 'worst first, then the order of the reply');
  const more = r.root.children[0].children[0].children[1]; // the header's button: "+2 more" in place of the "..."
  assert.strictEqual(more.textContent, '+2 more');
  more.fire('click');
  assert.ok(r.root.children[0].children[2].popOpen, 'the header button opens the popover');
  assert.strictEqual(r.root.children[0].children[2].children[0].children.length, 5, 'the popover lists all five');
  assert.strictEqual(render('needs-you', glance({ needs: [gear] })).root.children[0].children[0].children[1].textContent, '...', 'the dots when nothing overflows');
}
{ // a button posts exactly the post and body the reply names, and nothing else
  const r = render('needs-you', glance({ needs: [raven] }));
  const acts = r.root.children[0].children[1].children[0].children[2].children;
  assert.deepStrictEqual(acts.map((b) => b.textContent), ['Go now', 'Not yet']);
  acts[0].fire('click');
  acts[1].fire('click');
  assert.deepStrictEqual(r.posts, [['/api/orders', { verb: 'go', zone: 'Cold Plains' }], ['/api/journey/hold', undefined]]);
  const three = need('x', 'info', 'x', '', [verb('a', '/api/a'), verb('b', '/api/b'), verb('c', '/api/c')]);
  const r2 = render('needs-you', glance({ needs: [three] }));
  assert.strictEqual(r2.root.children[0].children[1].children[0].children[2].children.length, 2, 'at most two buttons per row');
}
{ // the panel hides when the needs go away and when the reply fails; an identical reply is not redrawn
  const r = render('needs-you', glance({ needs: [gear] }));
  assert.strictEqual(r.c.hidden, false);
  const first = r.root.children[0].children[1].children[0];
  r.api.update(glance({ needs: [gear] }));
  assert.strictEqual(r.root.children[0].children[1].children[0], first, 'same reply, same nodes');
  r.api.update(glance({ needs: [] }));
  assert.strictEqual(r.c.hidden, true);
  r.api.update(glance({ needs: [gear] }));
  assert.strictEqual(r.c.hidden, false);
  r.api.update(null);
  assert.strictEqual(r.c.hidden, true);
}
{ // notifications: off by default, a NEW red row only, and the first reply never pings
  const sent = [];
  function N(t, o) { sent.push([t, o.tag]); }
  N.permission = 'granted'; N.requestPermission = () => {};
  const P = load(path.join(dir, 'needs-you', 'panel.js'), { Notification: N });
  const store = {}, posts = [], { root, c } = ctxFor(posts, store);
  const api = P.mount(root, c);
  api.update(glance({ needs: [gear] }));
  api.update(glance({ needs: [gear, need('life', 'danger', 'Life is low')] }));
  assert.deepStrictEqual(sent, [], 'off by default');
  const pop = root.children[0].children[2];
  const sw = pop.children[1];
  assert.strictEqual(sw.attrs.role, 'switch');
  assert.strictEqual(sw.attrs['aria-checked'], 'false');
  sw.fire('click');
  assert.strictEqual(store.notify, '1');
  assert.strictEqual(sw.attrs['aria-checked'], 'true');
  api.update(glance({ needs: [gear, need('life', 'danger', 'Life is low'), raven]}));
  assert.deepStrictEqual(sent, [], 'an amber row or a known red one does not ping');
  api.update(glance({ needs: [gear, need('life', 'danger', 'Life is low'), need('dead', 'danger', 'Mercenary is dead', 'revive')] }));
  assert.deepStrictEqual(sent, [['Mercenary is dead', 'dead']], 'a new red row pings once');
  api.update(glance({ needs: [gear, need('life', 'danger', 'Life is low'), need('dead', 'danger', 'Mercenary is dead', 'revive')] }));
  assert.strictEqual(sent.length, 1);
  const P2 = load(path.join(dir, 'needs-you', 'panel.js'), { Notification: N });
  const { root: r3, c: c3 } = ctxFor([], { notify: '1' });
  const a3 = P2.mount(r3, c3);
  a3.update(glance({ needs: [gear] }));
  assert.strictEqual(sent.length, 1, 'a page that opens on a red row does not ping');
}

// ---- progress
const lvl = (o) => Object.assign({ from: 7, to: 8, pct: 62, etaSeconds: 1500 }, o);
const series = [8e3, 9e3, 9.5e3, 1e4, 11e3, 10e3, 12e3, 12.4e3];
const PG = {
  measuring: { level: lvl({ pct: 4, etaSeconds: null }), xpPerHour: null, xpSeries: [], goldPerHour: null, goldSeries: [] },
  full: { level: lvl(), xpPerHour: 12400, xpSeries: series, goldPerHour: 2100, goldSeries: [1900, 2000, 2100] },
  nohero: { level: { from: 0, to: 0, pct: null, etaSeconds: null }, xpPerHour: null, xpSeries: [], goldPerHour: null, goldSeries: [] },
};
for (const [st, progress] of Object.entries(PG)) {
  const r = render('progress', glance({ progress }));
  golden('progress', st, dump(r.root, ''), r.c, false);
}
{
  const P = panel('progress');
  assert.strictEqual(P.short(12400), '12.4k'); assert.strictEqual(P.short(950), '950'); assert.strictEqual(P.short(null), '');
  assert.strictEqual(P.eta(1500), 'about 25 min'); assert.strictEqual(P.eta(7800), 'about 2 h 10 min'); assert.strictEqual(P.eta(3600), 'about 1 h'); assert.strictEqual(P.eta(null), '');
  assert.deepStrictEqual(JSON.parse(JSON.stringify(P.points([0, 5, 10]))), [[2, 13], [30, 8], [58, 3]]); assert.deepStrictEqual(JSON.parse(JSON.stringify(P.points([3, 3]))), [[2, 8], [58, 8]]); assert.strictEqual(P.points([3]).length, 0);
  assert.strictEqual(render('progress', glance({ progress: PG.full })).root.children[0].children[0].children.length, 3, 'head, meter row, rates');
  assert.ok(render('progress', glance({ progress: PG.measuring })).c.st === 'ready');
  assert.strictEqual(render('progress', glance({})).c.hidden, false, 'an empty progress object still draws');
  assert.strictEqual(render('progress', null).c.hidden, true);
}

// ---- journey-strip
const chapters = (n, st, held) => Array.from({ length: 9 }, (_, i) => ({ id: 'ch' + (i + 1), state: i + 1 < n ? 'done' : i + 1 === n ? (held ? 'held' : st) : 'todo' }));
const jr = (n, o) => ({ n, of: 9, chapters: chapters(n, 'open', o && o.held), next: { id: 'ch' + n, title: 'Blood Raven', band: 'too early (floor 9)', text: 'Level 7 is under the floor of 9.' },
  actions: [verb('Go now', '/api/journey/go', { chapter: 'ch' + n }), verb('Hold', '/api/journey/hold', { chapter: 'ch' + n })] });
const JS = {
  'chapter-1': jr(1), 'chapter-3': jr(3), 'chapter-9': jr(9),
  blocked: Object.assign(jr(3, { held: true }), { next: { id: 'ch3', title: 'Blood Raven', band: 'blocked', text: 'The path is blocked.' } }),
  absent: undefined,
};
for (const [st, journey] of Object.entries(JS)) {
  const r = render('journey-strip', glance({ journey }));
  golden('journey-strip', st, dump(r.root, ''), r.c, false);
}
{
  const r = render('journey-strip', glance({ journey: JS['chapter-3'] }));
  const acts = r.root.children[0].children[0].children[2].children[1].children;
  acts[0].fire('click'); acts[1].fire('click');
  assert.deepStrictEqual(r.posts, [['/api/journey/go', { chapter: 'ch3' }], ['/api/journey/hold', { chapter: 'ch3' }]]);
  const dots = r.root.children[0].children[0].children[1].children.filter((x) => x.tag === 'button');
  assert.strictEqual(dots.length, 9);
  dots[1].fire('click');
  assert.ok(r.root.children[0].children[1].popOpen, 'a dot opens the chapter list');
  const P = panel('journey-strip');
  assert.strictEqual(P.current(JS['chapter-9'].chapters), 8); assert.strictEqual(P.current([{ state: 'done' }, { state: 'done' }]), 1);
  assert.deepStrictEqual(JSON.parse(JSON.stringify(P.line({ chapters: [{ state: 'done' }], next: null }))), ['Every chapter is done', '']);
}

// ---- health
const L = (light, word, detail) => ({ light, word, detail });
const HV = { life: L('ok', '150/150'), gear: L('ok', 'sound'), readiness: L('ok', 'safe'), merc: L('ok', 'Iron Wolf'), supplies: L('ok', 'stocked') };
const HS = {
  'all-ok': HV,
  'gear-broken': Object.assign({}, HV, { gear: L('bad', 'broken', 'The weapon is broken.'), readiness: L('warn', 'risky', 'Lightning resistance is under the line.') }),
  'two-unknown': Object.assign({}, HV, { merc: { light: 'unknown', word: 'unknown' }, supplies: { light: 'unknown', word: 'unknown' } }),
};
for (const [st, health] of Object.entries(HS)) {
  const r = render('health', glance({ health }));
  golden('health', st, dump(r.root, ''), r.c, false);
}
{
  const P = panel('health');
  assert.strictEqual(P.worst(HV), 'ok'); assert.strictEqual(P.worst(HS['gear-broken']), 'bad'); assert.strictEqual(P.worst({}), 'unknown');
  assert.strictEqual(P.wordOf({ light: 'unknown', word: 'unknown' }), 'not reported'); assert.strictEqual(P.lightOf({ light: 'purple' }), 'unknown');
  const r = render('health', glance({ health: HS['two-unknown'] }));
  r.root.children[0].children[0].fire('click');
  assert.ok(r.root.children[0].children[1].popOpen);
  const out = dump(r.root, '');
  assert.ok(out.includes('"not reported"'), 'unknown says not reported');
  assert.ok(out.includes('data-l="unknown"'), 'and is drawn hollow');
}

if (bad) process.exit(1);
console.log('ok');
