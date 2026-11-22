// The Vault Sell button and the instant sale against a fake bot (TestVaultSellJS): the button never silently differs from the selection,
// the reasons are shown, the confirm sheet states the gold and the stash after, the press sells at once. Two fixture states (the screenshot of 2026-11-22: 260 selected).
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


let clock = 0, timers = [], items = [], act = null;
const later = (f, ms) => { const t = { f, at: clock + ms }; timers.push(t); return t; };
const clear = (t) => { timers = timers.filter((x) => x !== t); };
const item = (id, extra) => Object.assign({ id, code: 'cap', name: 'Cap ' + id, state: 'recorded', quality: 'magic', q: 4, ilvl: 5, identified: true, kind: 'armor', value: 30, pin: false }, extra);
const view = () => ({ schema: 'watch.vault/1', mode: 'on', hero: 'Sorc', total: items.length, offset: 0, limit: 300, items, kinds: [], types: [], act, marked: null,
  capacity: { count: items.length, capacity: 400, pins: 0, pinLimit: 0, percent: 1, level: 'ok' }, upgrade: { available: false, currency: 'Ether', reason: 'x', max: 2400 } });
const get = async (p) => (p === '/api/ether' ? { schema: 'watch.ether/1', enabled: false } : view());
const posts = [];
let sellReply = null, quoteReply = null;
const post = async (route, body) => {
  posts.push({ route, body });
  if (route === '/api/vault/sellquote') return quoteReply(body);
  if (route === '/api/vault/sell') return sellReply(body);
  return { ok: true, done: [] };
};
const reveal0 = { enabled: false, charges: 0, free: false, balance: 0, queued: 0, working: 0, done: 0, methods: ['gold'], bar: { total: 0, default: '', line: '', options: [] }, gold: { on: false } };
const page = mountVault({ doc, root, get, post, later, clear, texts: T, now: () => clock });
const allIds = () => items.map((i) => i.id);

// the 260 of the screenshot: 11 can be sold, 169 are already waiting for the next town visit, 20 are pinned, 60 are gone from the vault
const build = () => {
  const out = [];
  for (let i = 0; i < 11; i++) out.push(item('r' + i));
  for (let i = 0; i < 169; i++) out.push(item('q' + i, { state: 'sellq', sell: 'queued' }));
  for (let i = 0; i < 20; i++) out.push(item('p' + i, { pin: true }));
  for (let i = 0; i < 60; i++) out.push(item('s' + i, { state: 'sold', sell: 'sold' }));
  return out;
};

// ---- fixture 1: the old queue (deploy 31): "Sell (11)" with 260 selected must say what it leaves out ----
items = build();
act = { reveal: reveal0, sell: { enabled: true, instant: false, queued: 169, working: 0, sold: 0, gold: 0 } };
await page.start(); await tick();
await find('select-toggle').click(); await find('select-all').click();
page.state.sel.all(allIds()); page.draw();
ok(/260 selected/.test(shownText(cls('vt-ab-sel')[0])), '260 selected: ' + shownText(cls('vt-ab-sel')[0]));
const sb = find('sell');
ok(sb.textContent === 'Sell 11 of 260', 'the button states the count it will act on: ' + sb.textContent);
const why = find('sell-why');
ok(why && !why.hidden, 'the reasons are on screen, not only in a tooltip');
ok(/169 already waiting to be sold/.test(why.textContent) && /20 pinned/.test(why.textContent) && /60 already sold, kept or discarded/.test(why.textContent) && /Not selling 249 of 260/.test(why.textContent), 'the breakdown names every group: ' + why.textContent);
ok(sb.title === why.textContent, 'the tooltip carries the same sentence');
// a selection that is all sellable says nothing extra
items = build().slice(0, 11); act.sell.queued = 0; await page.reload(false);
page.state.sel.clear(); page.state.sel.all(allIds()); page.draw();
ok(find('sell').textContent === 'Sell (11)' && find('sell-why').hidden, 'all sellable: "Sell (11)" and no reason line: ' + find('sell').textContent);

// ---- fixture 2: instant sale; queued ones are sellable now; quote, confirm sheet with the stash after, ack, chunks, the message ----
items = build();
act = { reveal: reveal0, sell: { enabled: true, instant: true, confirmAt: 2000, queued: 169, working: 0, sold: 0, gold: 0 } };
await page.reload(false);
page.state.sel.clear(); page.state.sel.all(allIds()); page.draw();
ok(find('sell').textContent === 'Sell 180 of 260', 'instant: queued items count, 80 stay out: ' + find('sell').textContent);
ok(/20 pinned/.test(find('sell-why').textContent) && !/waiting to be sold/.test(find('sell-why').textContent), 'queued is no reason any more: ' + find('sell-why').textContent);
quoteReply = (b) => { const ids = b.ids; const lines = ids.filter((id) => !/^[ps]/.test(id)).map((id) => ({ id, gold: 30 }));
  return { asked: ids.length, count: lines.length, lines, total: lines.length * 30, stash: 2000, stashAfter: 2000 + lines.length * 30, room: 2498000, confirmAt: 2000, needConfirm: false,
    reasons: [{ code: 'pinned', n: ids.filter((x) => x[0] === 'p').length }, { code: 'state', n: ids.filter((x) => x[0] === 's').length }].filter((r) => r.n) }; };
let fakeStash = 2000;
sellReply = (b) => { fakeStash += b.ids.length * 30; return { ok: true, done: b.ids, sale: { id: 's-1', state: 'done', items: b.ids.length, done: b.ids.length, gold: b.ids.length * 30, stashAfter: fakeStash, line: 'x' } }; };
posts.length = 0;
await find('sell').click();
const quotes = posts.filter((p) => p.route === '/api/vault/sellquote');
ok(quotes.length === 2 && quotes[0].body.ids.length === 150 && quotes[1].body.ids.length === 110, 'the quote goes in chunks of 150: ' + quotes.map((q) => q.body.ids.length));
const sheet = () => all(root).filter((n) => !n.hidden && n.dataset && n.dataset.key === 'sheet-go');
ok(sheet().length === 1, 'the total (5,400) is above loot.vaultSellConfirm: the sheet asks first');
const sheetText = shownText(root);
ok(/Sell 180 items for 5,400 gold\?/.test(sheetText) && /Stash after: 7,400 \(2,000 \+ 5,400\)/.test(sheetText), 'the sheet states the gold and the stash after: ' + (sheetText.match(/Sell 180[^]{0,160}/) || [''])[0]);
ok(/Not selling 80 of 260/.test(sheetText) && /20 pinned/.test(sheetText) && /60 already sold/.test(sheetText), 'and who is left out');
ok(!posts.some((p) => p.route === '/api/vault/sell'), 'nothing is sold before the confirm');
await sheet()[0].click();
const sells = posts.filter((p) => p.route === '/api/vault/sell');
ok(sells.length === 2 && sells.every((s) => s.body.confirm === true && s.body.ack === true) && sells[0].body.ids.length === 150 && sells[1].body.ids.length === 30, 'one sale per chunk, with confirm and ack: ' + JSON.stringify(sells.map((s) => [s.body.ids.length, s.body.ack])));
ok(!sells.some((s) => s.body.ids.some((id) => /^[ps]/.test(id))), 'only the quoted records are sent');
ok(/Sold 180 items for 5,400 gold\. Stash gold: 7,400\./.test(shownText(root)), 'the message says what was paid: ' + (shownText(root).match(/Sold[^]{0,80}/) || [''])[0]);

// below the threshold: no sheet, sold at once, no ack
items = build().slice(0, 11); await page.reload(false);
page.state.sel.clear(); page.state.sel.all(allIds()); page.draw();
quoteReply = (b) => ({ asked: b.ids.length, count: b.ids.length, lines: b.ids.map((id) => ({ id, gold: 30 })), total: b.ids.length * 30, stash: 100, stashAfter: 100 + b.ids.length * 30, room: 2499900, confirmAt: 2000, needConfirm: false, reasons: [] });
posts.length = 0;
await find('sell').click();
const s2 = posts.filter((p) => p.route === '/api/vault/sell');
ok(s2.length === 1 && s2[0].body.ack !== true && s2[0].body.ids.length === 11 && sheet().length === 0, 'a small sale sells at once without a sheet: ' + JSON.stringify(s2.map((s) => s.body.ack)));

// the item card shows the engine's price for the open item
items = build().slice(0, 3); await page.reload(false);
quoteReply = (b) => ({ asked: b.ids.length, count: b.ids.length, lines: b.ids.map((id) => ({ id, gold: 43 })), total: b.ids.length * 43, stash: 100, stashAfter: 143, room: 1, confirmAt: 2000, needConfirm: false, reasons: [] });
page.state.detailId = 'r0'; page.draw(); await tick(); await tick();
const sv = find('sell-value');
ok(sv && !sv.hidden && /Sell value 43 gold/.test(sv.textContent), 'the card shows the engine price: ' + (sv && sv.textContent));

// nothing sellable: the page says why instead of doing nothing
items = build().slice(11 + 169, 11 + 169 + 20); await page.reload(false);
page.state.sel.clear(); page.state.sel.all(allIds()); page.draw();
quoteReply = (b) => ({ asked: b.ids.length, count: 0, lines: [], total: 0, stash: 100, stashAfter: 100, room: 2499900, confirmAt: 2000, needConfirm: false, reasons: [{ code: 'pinned', n: b.ids.length }] });
posts.length = 0;
ok(find('sell').textContent === 'Sell 0 of 20', 'nothing sellable: ' + find('sell').textContent);
await find('sell').click();
ok(/Nothing selected can be sold: 20 pinned/.test(shownText(root)) && !posts.some((p) => p.route === '/api/vault/sell'), 'the reason is shown and nothing is posted: ' + (shownText(root).match(/Nothing[^]{0,60}/) || [''])[0]);

console.log(JSON.stringify({ fails }));
