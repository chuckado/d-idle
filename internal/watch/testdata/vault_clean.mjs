// The Vault page chrome (card vaultclean, TestVaultCleanChromeJS): no empty fold, one action bar, status chips that go, one marked chip,
// the unidentified badge only on unidentified tiles, the docked card only while an item is open. Two fixture states: all identified; mixed.
import { readFileSync } from 'fs';
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
const post = async () => reply;
const gold = (total, order) => ({ enabled: true, charges: 0, free: false, balance: 0, queued: 0, working: 0, done: 0, methods: ['gold'],
  bar: { total, default: total ? 'gold' : '', line: '', options: total ? [{ id: 'gold', label: 'Gold', line: 'g', ok: true, count: total, ids: items.map((i) => i.id), cost: total * 80, skipped: [] }] : [] },
  gold: { on: true, order: order || null, stash: 5000, scrollGold: 80, confirmAt: 0 } });
const sell = { enabled: true, queued: 0, working: 0, sold: 0, gold: 0 };
const page = mountVault({ doc, root, get, post, later, clear, texts: T, now: () => clock });
const bar = () => cls('vt-actionbar');

// ---- state 1: everything identified, nothing to reveal ----
items = ['v1', 'v2', 'v3'].map((id) => item(id, true));
act = { reveal: gold(0), sell };
await page.start(); await tick();
ok(bar().length === 1, 'exactly one action bar element: ' + bar().length);
ok(bar()[0].hidden, 'with nothing to do the action bar is hidden entirely');
ok(cls('vt-fold').length === 0 && cls('vt-fold-caret').length === 0 && !all(root).some((n) => (n.className || '').includes('vt-revealbar')), 'no empty fold or chevron element renders when N = 0');
ok(cls('vt-keeprow').length === 0 && cls('vt-msg').length === 0, 'the old bands are gone');
ok(cls('vt-ab-reveal')[0].hidden, 'no Reveal button when N = 0');
ok(cls('vt-unid').length === 0 && !all(root).some((n) => n.dataset && n.dataset.tag === 'unid'), 'no unidentified badge on identified tiles');
const cells = cls('vt-cell');
ok(cells.length === 3 && !cells.some((c) => all(c).some((n) => n.textContent === '?')), 'no ? anywhere in an identified tile');
const aside = cls('vt-detail')[0];
ok(aside && aside.hidden, 'the item card is hidden while nothing is open');

// ---- state 2: mixed; something to reveal ----
items = [item('v1', true), item('v2', false), item('v3', false, { pin: true })];
act = { reveal: gold(2), sell };
await page.reload(false);
ok(bar().length === 1 && !bar()[0].hidden, 'the bar shows when there is something to reveal');
const go = find('reveal-bar-go');
ok(!go.hidden && go.children.length === 2 && go.children[0].textContent === 'Reveal 2 items with gold' && go.children[1].textContent === 'cost 160, stash 5,000' && /vt-sub/.test(go.children[1].className), 'the cost is a muted subtitle of the Reveal button: ' + shownText(go));
const tiles = cls('vt-cell');
const unid = (c) => all(c).filter((n) => n.dataset && n.dataset.tag === 'unid');
ok(unid(tiles[0]).length === 0, 'the identified tile has no badge');
ok(unid(tiles[1]).length === 1 && unid(tiles[1])[0].textContent === '?' && unid(tiles[2]).length === 1, 'each unidentified tile has exactly one ? badge');
ok(all(tiles[0]).every((n) => n.textContent !== '?'), 'no ? text on the identified tile');

// the docked card only while an item is open
page.state.detailId = 'v2'; page.draw();
ok(!aside.hidden && cls('vt-body')[0].dataset.detail === '1', 'the card shows while an item is open');
page.state.detailId = ''; page.draw();
ok(aside.hidden && !cls('vt-body')[0].dataset.detail, 'the card goes when nothing is open');

// the grid has no frame: no ancestor carries a fixed height or an overflow box
const grid = page.grid;
let anc = [], n = grid;
while (n) { anc.push(n); n = n.parent; }
ok(anc.every((a) => !(a.style && (a.style.height || a.style.blockSize || a.style.overflow || a.style.maxHeight))), 'no ancestor of the grid has an inline height or overflow');
ok(!anc.some((a) => /vt-frame|vt-scroll|vt-box/.test(a.className || '')), 'no frame class around the grid');

// ---- the order chip is transient ----
act = { reveal: gold(0, { id: 'g-9', state: 'revealed', total: 9, want: 9, done: 9, paid: 720, stashAfter: 564833, line: 'x' }), sell };
await page.reload(false);
const chip = cls('vt-ab-order')[0];
ok(!chip.hidden && shownText(chip).includes('9 of 9 revealed') && shownText(chip).includes('720 gold paid'), 'the finished order shows as a chip first: ' + shownText(chip));
await advance(9000);
ok(!chip.hidden, 'still there before 10 s');
await advance(1500);
ok(chip.hidden && bar()[0].hidden, 'the chip and the bar are gone after about 10 s');
await page.reload(false);
ok(chip.hidden, 'a reload does not bring a seen order back');
act = { reveal: gold(0, { id: 'g-10', state: 'queued', total: 2, want: 2, done: 0, paid: 0, line: 'x' }), sell };
await page.reload(false);
await advance(30000);
ok(!chip.hidden, 'a running order stays until it ends');
act = { reveal: gold(0, { id: 'g-11', state: 'revealed', total: 2, want: 2, done: 2, paid: 160, stashAfter: 1, line: 'x' }), sell };
await page.reload(false);
ok(!chip.hidden, 'a new ended order shows');
await find('status-dismiss').click();
ok(chip.hidden, 'dismiss hides the chip at once');

// ---- the toast ----
reply = { ok: true, done: ['v1'] };
await page.act('/api/vault/pin', ['v1']);
const toast = cls('vt-toast')[0];
ok(!toast.hidden && shownText(toast).includes('Pinned'), 'a press says what it did in a toast: ' + shownText(toast));
await advance(10500);
ok(toast.hidden, 'the toast goes after 10 s');
await page.act('/api/vault/pin', ['v1']);
await find('toast-dismiss').click();
ok(toast.hidden, 'the toast dismisses');

// ---- Marked for sale is one toolbar chip that opens a sheet ----
act = { reveal: gold(0), sell };
marked = { n: 2, ids: ['v1', 'v2'], names: ['Cap v1', 'Cap v2'] };
await page.reload(false);
const mc = find('marked-chip');
ok(mc && !mc.hidden && mc.textContent === 'Marked for sale 2', 'one Marked for sale chip: ' + (mc && mc.textContent));
ok(mc.parent === cls('vt-tools')[0], 'the chip sits in the toolbar');
ok(!find('marked-sell'), 'no Sell button in the page until the sheet opens');
await mc.click();
const sh = cls('vt-sheet')[0];
ok(!sh.hidden && find('marked-sell') && /Sell these 2 now/.test(find('marked-sell').textContent), 'the sheet has the list and the Sell button');
ok(shownText(sh).includes('Cap v1'), 'the sheet lists the items');
page.closeSheet();
marked = null;

// ---- the toolbar holds chips, search, Filters, sort, view ----
const tb = cls('vt-tools')[0];
for (const k of ['search', 'filters-toggle', 'sort', 'view-toggle', 'select-toggle']) ok(find(k) && find(k).parent === tb, k + ' is in the toolbar');
ok(cls('vt-slots')[0].parent === tb, 'the category chips are in the toolbar');

// ---- selecting: N selected + the actions + Done, in the same one bar ----
act = { reveal: gold(2), sell };
await page.reload(false);
await find('select-toggle').click();
await find('select-all').click();
ok(bar().length === 1 && find('select-done').parent.parent === bar()[0] || find('select-done').parent === cls('vt-ab-sel')[0], 'the selection actions live in the one bar');
ok(/3 selected/.test(shownText(cls('vt-ab-sel')[0])) && !find('keep').hidden && !find('sell').hidden && !find('pin').hidden && !find('discard').hidden, 'N selected and the actions show');
ok(cls('vt-ab-reveal')[0].hidden, 'the idle Reveal button gives way to the selection actions');
await find('select-done').click();
ok(cls('vt-ab-sel')[0].hidden && !cls('vt-ab-reveal')[0].hidden, 'Done returns to the idle bar');

console.log(JSON.stringify({ fails }));
