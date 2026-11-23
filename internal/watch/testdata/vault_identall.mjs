// Identify all (TASK identifyall, TestVaultIdentifyAllJS): the toolbar button counts the WHOLE vault, hides at 0, the confirm sheet words enough and short
// stashes, the order payload carries scope all_unrevealed, and the stash accounting holds on two fixture states.
import { readFileSync } from 'fs';
import { createRequire } from 'module';
globalThis.dToast = createRequire(import.meta.url)('./toast.cjs');
const dWhy = createRequire(import.meta.url)('./why.cjs');
import { mountVault } from './vault.mjs';
const T = JSON.parse(readFileSync(new URL('./vault.json', import.meta.url), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const mk = (tag) => { const n = { tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, value: '', id: '', style: { setProperty() {} },
  classList: { add(c) { n.className += ' ' + c; }, remove() {}, contains() { return false; }, toggle() {} },
  append(...c) { n.children.push(...c); }, appendChild(c) { c.parent = n; n.children.push(c); return c; }, replaceChildren(...c) { n.children = c; c.forEach((x) => { x.parent = n; }); }, setAttribute(k, v) { n.attrs[k] = v; }, getAttribute(k) { return n.attrs[k]; }, removeAttribute(k) { delete n.attrs[k]; },
  addEventListener(e, f) { (n.listeners[e] = n.listeners[e] || []).push(f); }, focus() { doc.activeElement = n; }, contains() { return false; }, querySelectorAll() { return []; }, click() { return fire(n, 'click'); } };
  return n; };
const doc = { activeElement: null, body: mk('body'), createElement: mk, addEventListener() {} };
const tick = () => new Promise((r) => setTimeout(r, 5));
const fire = async (n, e, arg) => { for (const f of n.listeners[e] || []) await f(arg || { key: '', preventDefault() {}, target: n }); await tick(); };
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => { c.parent = e; all(c, o); }); return o; };
const root = mk('div'); doc.body.append(root);
const find = (k) => all(root).find((n) => n.dataset && n.dataset.key === k);
const cls = (c) => all(root).filter((n) => (n.className || '').split(/\s+/).includes(c));
const shownText = (n) => (n.hidden ? '' : (n.children && n.children.length ? n.children.map(shownText).join(' ') : n.textContent || ''));

let clock = 0, timers = [], items = [], act = null, marked = null;
const later = (f, ms) => { const t = { f, at: clock + ms }; timers.push(t); return t; };
const clear = (t) => { timers = timers.filter((x) => x !== t); };
const advance = async (ms) => { clock += ms; const due = timers.filter((t) => t.at <= clock); timers = timers.filter((t) => t.at > clock); for (const t of due) await t.f(); await tick(); };
const item = (id, identified, extra) => Object.assign({ id, code: 'cap', name: 'Cap ' + id, state: 'recorded', quality: 'magic', q: 4, ilvl: 5, identified, kind: 'armor', value: 10, pin: false }, extra);
const view = () => ({ schema: 'watch.vault/1', mode: 'on', hero: 'Sorc', total: items.length, offset: 0, limit: 60, items, kinds: [], types: [], act, marked,
  capacity: { count: 4, capacity: 400, pins: 0, pinLimit: 0, percent: 1, level: 'ok' }, upgrade: { available: false, currency: 'Ether', reason: 'x', max: 2400 } });
const get = async (p) => (p === '/api/ether' ? { schema: 'watch.ether/1', enabled: false } : view());
let reply = { ok: true, done: ['v1'] };
const posts = [];
const post = async (p, b) => { posts.push([p, b]); return reply; };
const gold = (total, order, un, stash) => ({ enabled: true, charges: 0, free: false, balance: 0, queued: 0, working: 0, done: 0, methods: ['gold'],
  bar: { total, default: total ? 'gold' : '', line: '', options: total ? [{ id: 'gold', label: 'Gold', line: 'g', ok: true, count: total, ids: items.map((i) => i.id), cost: total * 80, skipped: [] }] : [] },
  gold: { on: true, order: order || null, stash: stash === undefined ? 5000 : stash, scrollGold: 80, confirmAt: 0, unrevealed: un === undefined ? total : un } });
const sell = { enabled: true, queued: 0, working: 0, sold: 0, gold: 0 };
const page = mountVault({ toastHost: root, doc, root, get, post, later, clear, texts: T, now: () => clock });
const bar = () => cls('vt-actionbar');


const ident = () => find('identify-all');
const sheet = () => cls('vt-sheet')[0];

// ---- state 1: 2 of 3 unidentified in the list, 30 in the whole vault; a filter that shows 1 item ----
items = [item('v1', true), item('v2', false), item('v3', false)];
act = { reveal: gold(2, null, 30), sell };
await page.start(); await tick();
ok(ident() && !ident().hidden && ident().textContent === 'Identify all (30)', 'the button counts the whole vault, not the list or the bar: ' + (ident() && ident().textContent));
ok(ident().parent === cls('vt-tools')[0], 'the button is in the toolbar');
page.state.f.q = 'v2'; page.state.items = [items[1]]; page.draw();
ok(!ident().hidden && ident().textContent === 'Identify all (30)', 'a filter and a search do not change the count');
await find('select-toggle').click();
ok(!ident().hidden && ident().textContent === 'Identify all (30)', 'still shown while selecting (select mode)');
await find('select-all').click();
ok(!ident().hidden && ident().textContent === 'Identify all (30)', 'still shown with a selection');
await find('select-done').click();

// enough gold: 30 x 80 = 2,400 of 5,000
await ident().click();
ok(!sheet().hidden && shownText(sheet()).includes('Identify 30 items for 2,400 gold? Stash 5,000, left after: 2,600'), 'confirm text, stash enough: ' + shownText(sheet()));
ok(find('sheet-go').textContent === 'Identify 30', 'the confirm button names the count: ' + find('sheet-go').textContent);
posts.length = 0;
await find('sheet-go').click();
ok(posts.length === 1 && posts[0][0] === '/api/vault/revealgold', 'one POST to the gold route: ' + JSON.stringify(posts));
const pl = posts[0] && posts[0][1];
ok(pl && pl.scope === 'all_unrevealed' && pl.maxGold === 2400 && pl.confirm === true && !('ids' in pl), 'the payload carries scope all_unrevealed and no ids: ' + JSON.stringify(pl));
ok(sheet().hidden, 'the sheet closes');

// ---- state 2: stash short: 1,000 buys 12 of 30 ----
act = { reveal: gold(2, null, 30, 1000), sell };
await page.reload(false);
await ident().click();
ok(shownText(sheet()).includes('Identify 12 of 30 for 960 gold (stash 1,000): the rest later'), 'confirm text, stash short: ' + shownText(sheet()));
ok(find('sheet-go').textContent === 'Identify 12', 'the button says Identify K: ' + find('sheet-go').textContent);
posts.length = 0;
await find('sheet-go').click();
ok(posts[0] && posts[0][1].maxGold === 960 && posts[0][1].scope === 'all_unrevealed', 'short stash: cap is the K price');

// stash below one reveal: disabled, no sheet
act = { reveal: gold(2, null, 30, 50), sell };
await page.reload(false);
const identWhy = find('why-identify-all') || all(root).find((n) => n.id === ident().attrs['aria-describedby']);
ok(ident().disabled && identWhy && !identWhy.hidden && /Not enough gold/.test(identWhy.textContent) && !ident().attrs.title, 'a stash below one reveal disables the button with a visible reason line: ' + (identWhy && identWhy.textContent));

// a running order disables it
act = { reveal: gold(2, { id: 'g-1', state: 'paid', total: 12, want: 30, done: 0, paid: 960, line: 'x' }, 30, 40), sell };
await page.reload(false);
ok(ident().disabled, 'disabled while an order runs');

// ---- hidden at 0 ----
act = { reveal: gold(0, null, 0), sell };
await page.reload(false);
ok(ident().hidden, 'hidden when nothing is unrevealed');

// ---- the progress chip follows the order as today ----
act = { reveal: gold(0, { id: 'g-2', state: 'revealed', total: 12, want: 30, done: 12, paid: 960, stashAfter: 40, line: 'x' }, 18), sell };
await page.reload(false);
ok(!ident().hidden && ident().textContent === 'Identify all (18)', 'the rest remain countable after a partial order');
ok(shownText(cls('vt-ab-order')[0]).includes('12 of 30'), 'the chip shows the order: ' + shownText(cls('vt-ab-order')[0]));

// a partial answer (some records could not be decoded) says so in the toast
act = { reveal: gold(2, null, 30), sell };
await page.reload(false);
reply = { ok: false, done: ['a', 'b'], failed: [{ id: 'c', code: 'nosealed', error: 'x' }], order: { id: 'g-3', state: 'revealed', line: 'Gold reveal done: 29 items revealed' } };
await ident().click(); await find('sheet-go').click();
ok(/Gold reveal done: 29 items revealed 1 could not be read/.test(shownText(cls('d-toast')[0])), 'the toast names the records that stay unrevealed: ' + shownText(cls('d-toast')[0]));

for (const w of dWhy.unexplained(all(root))) fails.push('why: ' + w);
console.log(JSON.stringify({ fails }));
