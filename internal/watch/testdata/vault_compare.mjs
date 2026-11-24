// The Vault item sheet on a phone (TestVaultCompareJS): the selected cell stays above the bottom sheet, the sheet carries the tooltip
// lines, and one button cycles Compare off -> hero -> merc -> off with the stat changes of the inventory tooltip's own comparison.
// Prints {fails}. Two classes' worth of gear (a Paladin's shield and sword, a Sorceress's staff and rings) and the merc kinds.
import { readFileSync } from 'fs';
import vm from 'node:vm';
import { createRequire } from 'module';
globalThis.dToast = createRequire(import.meta.url)('./toast.cjs');
import { mountVault } from './vault.mjs';
import { clearDelta, comparePlan, nextCompare, validCompare, COMPARE_MODES } from './vault-logic.mjs';
const T = JSON.parse(readFileSync(new URL('./vault.json', import.meta.url), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };

// the real tooltip script gives lines and compare (window.d2InvTip), as on the page
const win = { addEventListener() {}, matchMedia: () => ({ matches: true }), getComputedStyle: () => ({ position: 'fixed', getPropertyValue: () => '', fontSize: '16px' }), innerHeight: 640, scrolled: [],
  scrollBy(x, y) { win.scrolled.push(y); }, localStorage: { m: {}, getItem(k) { return this.m[k] ?? null; }, setItem(k, v) { this.m[k] = v; } } };
const dom0 = { getElementById: () => null, addEventListener() {}, body: {}, documentElement: { clientWidth: 360, clientHeight: 640 } };
vm.runInNewContext(readFileSync(new URL('./inventory-tip.cjs', import.meta.url), 'utf8'), { window: win, document: dom0, Math, Object, Array, String, Number, RegExp, JSON });
const tip = win.d2InvTip;
ok(tip && typeof tip.compare === 'function' && typeof tip.lines === 'function', 'the tooltip script exposes lines and compare');

// ---- 1. where the sheet sits: the cell must clear it ----
// The defect: the floating tip is placed beside the cell (place()) and the bottom sheet (z-index above the tip) covers it for a cell in the last rows.
for (const [vw, vh] of [[360, 640], [390, 844]]) {
  const sheetTop = vh - Math.round(vh * 0.6); // max-block-size 60dvh
  const cell = { top: vh - 90, bottom: vh - 20 }; // a cell in the last row of the screen
  const d = clearDelta(cell, sheetTop, 0, 0);
  ok(d > 0, vw + 'x' + vh + ': a last-row cell is under the sheet and must scroll up (delta ' + d + ')');
  ok(cell.bottom - d <= sheetTop - 8 + 0.5, vw + 'x' + vh + ': after the scroll the cell ends above the sheet');
  ok(clearDelta({ top: 60, bottom: 130 }, sheetTop, 0, 0) === 0, vw + 'x' + vh + ': a cell in the top part does not move');
  ok(clearDelta({ top: -40, bottom: 20 }, sheetTop, 0, 0) < 0, vw + 'x' + vh + ': a cell above the view scrolls down into it');
  const tall = clearDelta({ top: sheetTop - 30, bottom: sheetTop + 200 }, sheetTop, 0, 0);
  ok(sheetTop - 30 - tall >= 8 - 0.5, 'a cell taller than the band keeps its top in view');
}

// ---- 2. the compare cycle ----
ok(JSON.stringify(COMPARE_MODES) === '["off","hero","merc"]', 'modes');
ok(nextCompare('off') === 'hero' && nextCompare('hero') === 'merc' && nextCompare('merc') === 'off', 'off -> hero -> merc -> off');
ok(validCompare('both') === 'off' && validCompare(null) === 'off' && validCompare('merc') === 'merc', 'a stored junk value is off');

// ---- 3. deltas, empty slot, rings, two-hander, merc ----
const base = (o) => Object.assign({ id: 1, name: 'x', quality: 'magic', identified: true, props: [], base: {} }, o);
const cand = (o) => tip.compare && base(o);
const cmp = (n, o) => tip.compare(n, o);
const inv = { equipment: {
  torso: base({ name: 'Quilted Armor', base: { name: 'Quilted Armor', minAC: 30, maxAC: 30 }, props: ['+10 to Strength'] }),
  rhand: base({ name: 'Broad Sword', base: { name: 'Broad Sword', minDam: 4, maxDam: 12 } }),
  lhand: base({ name: 'Buckler', base: { name: 'Buckler', minAC: 4, maxAC: 4 } }),
  rring: base({ name: 'Ring A', props: ['+5 to Dexterity'] }), lring: base({ name: 'Ring B', props: ['+20 to Life'] }) },
  merc: { owned: true, equipment: { mtorso: base({ name: 'Leather Armor', base: { name: 'Leather Armor', minAC: 14, maxAC: 14 } }) } } };
const armor = base({ name: 'Splint Mail', base: { name: 'Splint Mail', minAC: 60, maxAC: 60 }, props: ['+15 to Strength'] });
const heroSh = { wearable: true, slots: [{ slot: 'torso', label: 'armour', can: true }], merc: { owned: true, kind: 'ironwolf', can: true, slot: 'mtorso' } };
let p = comparePlan('off', heroSh, inv, armor, cmp, T);
ok(p.state === 'off' && !p.blocks.length, 'off draws nothing');
p = comparePlan('hero', undefined, undefined, armor, cmp, T);
ok(p.state === 'loading', 'no data yet: loading');
p = comparePlan('hero', heroSh, inv, armor, cmp, T);
const row = (b, label) => b.rows.find((r) => r.label === label);
ok(p.state === 'ok' && p.blocks.length === 1 && p.blocks[0].worn === 'Quilted Armor', 'hero: the worn armour is the one in the same slot');
ok(row(p.blocks[0], 'Defense') && row(p.blocks[0], 'Defense').delta === '+30' && row(p.blocks[0], 'Defense').tone === 'up', 'defense 30 -> 60 is +30, better: ' + JSON.stringify(p.blocks[0].rows));
ok(row(p.blocks[0], 'Strength') && row(p.blocks[0], 'Strength').delta === '+5', 'strength +10 -> +15 is +5');
const worse = comparePlan('hero', heroSh, inv, base({ name: 'Rags', base: { name: 'Rags', minAC: 10, maxAC: 10 } }), cmp, T);
ok(row(worse.blocks[0], 'Defense').tone === 'dn' && row(worse.blocks[0], 'Defense').delta === '-20', 'a worse item is red (dn): ' + JSON.stringify(worse.blocks[0].rows));
// empty slot
const emptyInv = { equipment: {}, merc: { owned: true, equipment: {} } };
p = comparePlan('hero', heroSh, emptyInv, armor, cmp, T);
ok(p.state === 'ok' && p.blocks[0].empty && p.blocks[0].rows.length === 0, 'an empty slot says nothing is equipped');
p = comparePlan('merc', heroSh, emptyInv, armor, cmp, T);
ok(p.state === 'ok' && p.blocks[0].empty && p.blocks[0].slot === 'mtorso', 'merc: empty merc slot');
// both ring slots
const ring = base({ name: 'New Ring', props: ['+30 to Life'] });
const ringSh = { wearable: true, slots: [{ slot: 'lring', label: 'left ring', side: 'left', can: true }, { slot: 'rring', label: 'right ring', side: 'right', can: true }], merc: { owned: true, kind: 'ironwolf', can: false, why: '' } };
p = comparePlan('hero', ringSh, inv, ring, cmp, T);
ok(p.blocks.length === 2 && p.blocks.map((b) => b.worn).join() === 'Ring B,Ring A', 'a ring shows both ring slots: ' + p.blocks.map((b) => b.worn));
ok(row(p.blocks[0], 'Life').delta === '+10' && row(p.blocks[1], 'Life').delta === '+30', 'ring deltas per slot');
// two-hander with a shield worn: the shield comes off, said in words
const twoSh = { wearable: true, slots: [{ slot: 'rhand', label: 'right hand', can: true, note: 'takes the shield off (a two-hander uses both hands)' }], merc: { owned: true, kind: 'desert', can: true, slot: 'mrhand' } };
p = comparePlan('hero', twoSh, inv, base({ name: 'Great Sword', base: { name: 'Great Sword', twoHanded: true, twoMinDam: 10, twoMaxDam: 30 } }), cmp, T);
ok(/Takes off Buckler/.test(p.blocks[0].note), 'two-hander names the shield it pushes off: ' + p.blocks[0].note);
// a shield while a two-hander is worn
const shSh = { wearable: true, slots: [{ slot: 'lhand', label: 'left hand', can: false, why: 'a two-handed weapon is worn: take it off first' }], merc: { owned: true, kind: 'rogue', can: false, why: 'a rogue cannot use that' } };
p = comparePlan('hero', shSh, inv, base({ name: 'Kite Shield', base: { name: 'Kite Shield', minAC: 8, maxAC: 8 } }), cmp, T);
ok(p.state === 'ok' && /two-handed weapon is worn/.test(p.blocks[0].note), 'a shield against a worn two-hander says why');
// merc cannot wear it / not hired / out of view
p = comparePlan('merc', ringSh, inv, ring, cmp, T);
ok(p.state === 'refused' && p.mercNo && /Merc cannot wear this/.test(p.why), 'merc cannot wear a ring: ' + p.why);
p = comparePlan('merc', shSh, inv, armor, cmp, T);
ok(p.state === 'refused' && /Merc cannot wear this/.test(p.why) && /rogue/.test(p.why), 'a rogue cannot use a shield: ' + p.why);
p = comparePlan('merc', Object.assign({}, heroSh, { merc: { owned: false, why: 'no mercenary is hired' } }), inv, armor, cmp, T);
ok(p.state === 'refused' && !p.mercNo && /no mercenary is hired/.test(p.why), 'no merc hired');
p = comparePlan('merc', Object.assign({}, heroSh, { merc: { owned: true, kind: '', can: false, why: 'your mercenary is out of view, so his kind is not known yet' } }), inv, armor, cmp, T);
ok(p.state === 'refused' && !p.mercNo && /out of view/.test(p.why), 'merc out of view is not "cannot wear"');
p = comparePlan('hero', false, inv, armor, cmp, T);
ok(p.state === 'refused' && p.why === T.cmpNoData, 'a bot without the equip route says the data is missing');
p = comparePlan('hero', { wearable: false, why: 'charms are not equipped: belt and backpack items are out of scope for Equip', slots: [] }, inv, armor, cmp, T);
ok(p.state === 'refused' && /charms/.test(p.why), 'a charm is not worn');

// ---- 4. the page: the button, its name, the saved choice, the block ----
const mk = (tag) => { const n = { tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, value: '', id: '', style: { setProperty() {} },
  classList: { add(c) { n.className += ' ' + c; }, remove() {}, contains() { return false; }, toggle() {} },
  append(...c) { n.children.push(...c); c.forEach((x) => { x.parent = n; }); }, appendChild(c) { c.parent = n; n.children.push(c); return c; }, insertBefore(c, ref) { const i = n.children.indexOf(ref); c.parent = n; n.children.splice(i < 0 ? n.children.length : i, 0, c); return c; },
  replaceChildren(...c) { n.children = c; c.forEach((x) => { x.parent = n; }); }, setAttribute(k, v) { n.attrs[k] = v; }, getAttribute(k) { return n.attrs[k] ?? null; }, hasAttribute(k) { return k in n.attrs; }, removeAttribute(k) { delete n.attrs[k]; },
  addEventListener(e, f) { (n.listeners[e] = n.listeners[e] || []).push(f); }, focus() { doc.activeElement = n; }, contains(x) { return all(n).includes(x); }, querySelectorAll() { return []; }, querySelector(sel) { return query(n, sel); },
  closest() { return /vt-cell/.test(n.className) ? n : null; }, getBoundingClientRect() { return n.rect || { top: 0, bottom: 0, left: 0, right: 0, height: 0 }; }, click() { return fire(n, 'click'); } };
  return n; };
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => { c.parent = e; all(c, o); }); return o; };
const query = (n, sel) => { const m = /\[data-(\w+)="([^"]*)"\]/.exec(sel); return all(n).find((x) => m && x.dataset && x.dataset[m[1]] === m[2]) || null; };
const doc = { activeElement: null, body: mk('body'), createElement: mk, addEventListener() {} };
const tick = () => new Promise((r) => setTimeout(r, 5));
const fire = async (n, e, arg) => { for (const f of n.listeners[e] || []) await f(arg || { key: '', preventDefault() {}, target: n }); await tick(); };
const root = mk('div'); doc.body.append(root);
const find = (k) => all(root).find((n) => n.dataset && n.dataset.key === k);
const shown = (n) => (n.hidden ? '' : (n.children && n.children.length ? n.children.map(shown).join(' ') : n.textContent || ''));

const item = (id, extra) => Object.assign({ id, code: 'ltp', name: 'Splint Mail ' + id, state: 'recorded', quality: 'magic', q: 4, ilvl: 30, level: 30, identified: true, kind: 'armor', value: 30, pin: false,
  base: { name: 'Splint Mail', minAC: 60, maxAC: 60, reqStr: 100 }, props: ['+15 to Strength'] }, extra);
const items = [item('a'), item('b')];
const view = () => ({ schema: 'watch.vault/1', mode: 'on', hero: 'Sorc', total: items.length, offset: 0, limit: 300, items, kinds: [], types: [], act: null, marked: null,
  capacity: { count: items.length, capacity: 400, pins: 0, pinLimit: 0, percent: 1, level: 'ok' }, upgrade: { available: false, currency: 'Ether', reason: 'x', max: 2400 } });
let equipReply = heroSh;
const gets = [];
const get = async (p) => { gets.push(p); if (p === '/api/ether') return { schema: 'watch.ether/1', enabled: false }; if (p === '/api/inventory') return inv; if (p.startsWith('/api/vault/equip/')) return equipReply; if (p.startsWith('/api/vault/item/')) return { item: items[0], lines: [], actions: {} }; return view(); };
const run = async (stored) => {
  win.localStorage.m = stored ? { 'vt-compare': stored } : {};
  root.children = [];
  const page = mountVault({ toastHost: root, doc, root, get, post: async () => ({ ok: true }), later: () => ({}), clear() {}, texts: T, now: () => 0, win });
  await page.start(); await tick();
  return page;
};
let page = await run('');
await page.state.sel; // no-op
const cellsOf = () => all(root).filter((n) => (n.className || '').split(/\s+/).includes('vt-cell'));
const open = async (i) => { const c = cellsOf()[i]; await fire(page.grid, 'click', { target: c, preventDefault() {} }); await tick(); };
await open(0);
let b = find('d-compare');
ok(b && b.textContent === 'Compare: off', 'the button starts off: ' + (b && b.textContent));
ok(b && /Compare with equipped gear: off/.test(b.attrs['aria-label']), 'its accessible name states the mode: ' + (b && b.attrs['aria-label']));
ok(b && b.tag === 'button' && b.attrs.disabled === undefined, 'it is a focusable button');
ok(find('d-tip') && /Base defense: 60/.test(shown(find('d-tip'))) && /Required strength: 100/.test(shown(find('d-tip'))), 'the sheet carries the tooltip base lines: ' + (find('d-tip') && shown(find('d-tip'))));
ok(!/Quilted/.test(shown(find('d-cmp'))), 'off: no block');
await b.click(); await tick(); await tick();
b = find('d-compare');
ok(b.textContent === 'Compare: hero' && /hero/.test(b.attrs['aria-label']), 'press 1: hero: ' + b.textContent);
ok(win.localStorage.m['vt-compare'] === 'hero', 'the choice is saved on this device');
ok(/Hero: armour/.test(shown(find('d-cmp'))) && /Equipped: Quilted Armor/.test(shown(find('d-cmp'))) && all(find('d-cmp')).some((n) => n.dataset && n.dataset.tone === 'up' && /^Defense: 30 > 60/.test(n.textContent) && n.children[0].textContent === '+30' && /better/.test(n.children[1].textContent)), 'hero block: ' + shown(find('d-cmp')));
ok(doc.activeElement === b || doc.activeElement === find('d-compare'), 'focus stays on the button after the press');
await b.click(); await tick(); await tick();
b = find('d-compare');
ok(b.textContent === 'Compare: merc' && /Merc: armour/.test(shown(find('d-cmp'))) && /Leather Armor/.test(shown(find('d-cmp'))), 'press 2: merc: ' + shown(find('d-cmp')));
await b.click(); await tick();
ok(find('d-compare').textContent === 'Compare: off' && win.localStorage.m['vt-compare'] === 'off' && !/Armor/.test(shown(find('d-cmp'))), 'press 3: off');
// a stored choice is used at the next visit; junk and a throwing storage fall back to off
page = await run('merc'); await open(0);
ok(find('d-compare').textContent === 'Compare: merc', 'the saved mode is used on the next visit: ' + find('d-compare').textContent);
page = await run('both'); await open(0);
ok(find('d-compare').textContent === 'Compare: off', 'a junk stored value is off');
const saved = win.localStorage; win.localStorage = { getItem() { throw new Error('blocked'); }, setItem() { throw new Error('blocked'); } };
page = await run(''); win.localStorage = saved; // run() resets the object: use a throwing one directly
win.localStorage = { getItem() { throw new Error('blocked'); }, setItem() { throw new Error('blocked'); } };
root.children = [];
const page2 = mountVault({ toastHost: root, doc, root, get, post: async () => ({ ok: true }), later: () => ({}), clear() {}, texts: T, now: () => 0, win });
await page2.start(); await tick(); page = page2; await open(0);
ok(find('d-compare').textContent === 'Compare: off', 'blocked storage: the page still draws, compare is off');
await find('d-compare').click(); await tick();
ok(find('d-compare').textContent === 'Compare: hero', 'blocked storage: the press still cycles');
win.localStorage = saved;
// the merc cannot wear it
equipReply = Object.assign({}, heroSh, { merc: { owned: true, kind: 'rogue', can: false, why: '' } });
page = await run('merc'); await open(0);
ok(/Merc cannot wear this/.test(shown(find('d-cmp'))), 'the page says Merc cannot wear this: ' + shown(find('d-cmp')));
console.log(JSON.stringify({ fails }));
