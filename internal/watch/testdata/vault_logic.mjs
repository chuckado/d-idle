// Node test of web/vault-logic.js (run by TestVaultLogicJS): query round trips, the client mirror of the server filter and sort,
// the multi-select with shift ranges, the capacity words. Usage: node vault_logic.mjs <path to vault-logic.js>
import assert from 'node:assert/strict';
import { pathToFileURL } from 'node:url';

const L = await import(pathToFileURL(process.argv[2]).href);
const T = { overBanner: 'Over by {over} ({n} held)', fullBanner: 'Full' };
const kinds = [{ id: 'unique', quality: [7] }, { id: 'rune', codePrefix: ['r0', 'r1', 'r2', 'r3'] }];
let n = 0;
const ok = (name, fn) => { fn(); n++; };

const row = (o) => Object.assign({ id: 'v-' + Math.random(), name: 'Cap', code: 'cap', q: 4, quality: 'magic', ilvl: 10, sockets: 0, ethereal: false, identified: true, pin: false, state: 'recorded', value: 5, found: '2026-11-16T10:00:00Z' }, o);

ok('default filters make an empty query', () => assert.equal(L.toQuery(L.defaultFilters()), ''));
ok('query round trip', () => {
  const f = Object.assign(L.defaultFilters(), { q: ' war ', quality: [7, 4], kind: 'unique', type: 'swor', ilvlMin: '20', ilvlMax: '30', sockets: '3', eth: true, unid: true, pinned: true, state: 'kept', sort: 'ilvl' });
  const q = L.toQuery(f, { offset: 60, limit: 60, facets: true });
  const p = new URLSearchParams(q);
  assert.equal(p.get('q'), 'war'); assert.equal(p.get('quality'), '4,7'); assert.equal(p.get('identified'), '0'); assert.equal(p.get('eth'), '1');
  assert.equal(p.get('offset'), '60'); assert.equal(p.get('facets'), '1'); assert.equal(p.get('sort'), 'ilvl'); assert.equal(p.get('sockets'), '3');
  const g = L.fromQuery(q);
  assert.deepEqual(g.quality.sort(), [4, 7]); assert.equal(g.kind, 'unique'); assert.equal(g.ilvlMin, '20'); assert.equal(g.unid, true); assert.equal(g.q, 'war'); assert.equal(g.sort, 'ilvl');
});
ok('sockets 0 is a filter, empty is not', () => {
  assert.equal(L.toQuery(Object.assign(L.defaultFilters(), { sockets: '0' })), '?sockets=0');
  assert.equal(L.activeCount(Object.assign(L.defaultFilters(), { sockets: '0' })), 1);
  assert.equal(L.activeCount(L.defaultFilters()), 0);
});
ok('a bad address is ignored, not trusted', () => {
  const f = L.fromQuery('?quality=99,rare&sort=drop&ilvlMin=-3&kind=x');
  assert.deepEqual(f.quality, []); assert.equal(f.sort, 'new'); assert.equal(f.ilvlMin, '');
});
ok('active count: sort does not count, an ilvl range is one', () => {
  const f = Object.assign(L.defaultFilters(), { sort: 'name', ilvlMin: '5', ilvlMax: '9', quality: [4, 6], eth: true });
  assert.equal(L.activeCount(f), 3);
});

ok('matches mirrors the server', () => {
  const f = L.defaultFilters();
  assert.equal(L.matches(row({}), f, kinds), true);
  assert.equal(L.matches(row({ state: 'discarded' }), f, kinds), false, 'live states only by default');
  assert.equal(L.matches(row({ state: 'discarded' }), Object.assign({}, f, { state: 'discarded,trimmed' }), kinds), true);
  assert.equal(L.matches(row({ q: 7 }), Object.assign({}, f, { kind: 'unique' }), kinds), true);
  assert.equal(L.matches(row({ q: 4 }), Object.assign({}, f, { kind: 'unique' }), kinds), false);
  assert.equal(L.matches(row({ code: 'r05' }), Object.assign({}, f, { kind: 'rune' }), kinds), true);
  assert.equal(L.matches(row({ code: 'cap' }), Object.assign({}, f, { kind: 'rune' }), kinds), false);
  assert.equal(L.matches(row({ ilvl: 29 }), Object.assign({}, f, { ilvlMin: '30' }), kinds), false);
  assert.equal(L.matches(row({ ilvl: 31 }), Object.assign({}, f, { ilvlMax: '30' }), kinds), false);
  assert.equal(L.matches(row({ sockets: 2 }), Object.assign({}, f, { sockets: '2' }), kinds), true);
  assert.equal(L.matches(row({ sockets: 0 }), Object.assign({}, f, { sockets: '0' }), kinds), true);
  assert.equal(L.matches(row({ ethereal: false }), Object.assign({}, f, { eth: true }), kinds), false);
  assert.equal(L.matches(row({ identified: true }), Object.assign({}, f, { unid: true }), kinds), false);
  assert.equal(L.matches(row({ pin: true }), Object.assign({}, f, { pinned: true }), kinds), true);
  assert.equal(L.matches(row({ name: 'Rare War Sword' }), Object.assign({}, f, { q: 'war sw' }), kinds), true);
  assert.equal(L.matches(row({ q: 6 }), Object.assign({}, f, { quality: [4, 7] }), kinds), false);
  assert.equal(L.matches(row({ q: 4 }), Object.assign({}, f, { kind: 'missing' }), kinds), false, 'an unknown kind matches nothing');
});
ok('sortItems', () => {
  const a = row({ id: 'a', name: 'b item', q: 4, ilvl: 10, value: 3, found: '2026-11-16T10:00:00Z' });
  const b = row({ id: 'b', name: 'A item', q: 7, ilvl: 5, value: 9, found: '2026-11-16T09:00:00Z' });
  const c = row({ id: 'c', name: 'c item', q: 7, ilvl: 50, value: 1, found: '2026-11-16T11:00:00Z' });
  const ids = (s) => L.sortItems([a, b, c], s).map((x) => x.id).join('');
  assert.equal(ids('new'), 'cab'); assert.equal(ids('old'), 'bac'); assert.equal(ids('quality'), 'cba'); assert.equal(ids('ilvl'), 'cab');
  assert.equal(ids('value'), 'bac'); assert.equal(ids('name'), 'bac'); assert.equal(ids('nope'), 'cab');
});

ok('selection: toggle, mode, anchor', () => {
  const s = L.createSelection();
  assert.equal(s.size, 0); assert.equal(s.mode, false);
  assert.equal(s.toggle('a'), true); assert.equal(s.mode, true); assert.equal(s.anchor, 'a');
  assert.equal(s.toggle('a'), false); assert.equal(s.size, 0);
  s.toggle('b'); assert.deepEqual(s.ids(), ['b']);
});
ok('selection: shift range from the anchor, either direction', () => {
  const order = ['a', 'b', 'c', 'd', 'e', 'f'], s = L.createSelection();
  s.toggle('b'); s.range(order, 'd');
  assert.deepEqual(s.ids().sort(), ['b', 'c', 'd']); assert.equal(s.anchor, 'd');
  s.range(order, 'a');
  assert.deepEqual(s.ids().sort(), ['a', 'b', 'c', 'd']);
  const t = L.createSelection(); t.range(order, 'c'); // no anchor: a plain toggle
  assert.deepEqual(t.ids(), ['c']);
});
ok('selection: select all, prune, clear', () => {
  const s = L.createSelection();
  s.all(['a', 'b', 'c']); assert.equal(s.size, 3);
  s.prune(['b', 'c', 'x']); assert.deepEqual(s.ids().sort(), ['b', 'c']);
  s.prune([]); assert.equal(s.size, 0); assert.equal(s.mode, false, 'the mode ends with the last item');
  s.toggle('a'); s.clear(); assert.equal(s.size, 0); assert.equal(s.anchor, null); assert.equal(s.mode, false);
  s.mode = true; assert.equal(s.mode, true);
});

ok('capacity words', () => {
  assert.deepEqual(L.capacityView({ count: 412, capacity: 600, level: 'ok' }, T), { text: '412 / 600', pct: 69, level: 'ok', now: 412, max: 600, banner: '' });
  assert.equal(L.capacityView({ count: 600, capacity: 600, level: 'full' }, T).banner, 'Full');
  const o = L.capacityView({ count: 450, capacity: 400, level: 'over', over: 50 }, T);
  assert.equal(o.banner, 'Over by 50 (450 held)'); assert.equal(o.pct, 100);
  assert.equal(L.capacityView(null, T).text, '0 / 0');
});
ok('when', () => {
  const now = Date.parse('2026-11-16T12:00:00Z');
  assert.equal(L.when('2026-11-16T11:59:30Z', now), 'just now'); assert.equal(L.when('2026-11-16T11:00:00Z', now), '1 h ago');
  assert.equal(L.when('2026-11-15T10:00:00Z', now), 'yesterday'); assert.equal(L.when('2026-09-01T10:00:00Z', now), '1 Sep 2026'); assert.equal(L.when('x', now), '');
});
ok('tip view numbers the rows and keeps the vault id', () => {
  const tv = L.toTipView([row({ id: 'v-a' }), row({ id: 'v-b' })]);
  assert.equal(tv.view.backpack[1].id, 2); assert.equal(tv.view.backpack[1].vid, 'v-b'); assert.equal(tv.byN.get(1), 'v-a');
});
ok('discard ask', () => { assert.equal(L.discardAsk(10, 10).typed, false); assert.equal(L.discardAsk(11, 10).typed, true); });
console.log('vault logic: ' + n + ' groups ok');
