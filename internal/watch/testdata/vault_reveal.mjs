// The Vault page Reveal bar against a fake bot (TestVaultRevealBarJS): the picker, the confirm step, the paid order chain and the stash gold line.
import { readFileSync } from 'fs';
import { mountVault } from './vault.mjs';
const T = JSON.parse(readFileSync(new URL('./vault.json', import.meta.url), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const mk = (tag) => { const n = { tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, value: '', id: '', style: { setProperty() {} },
  classList: { add(c) { n.className += ' ' + c; }, remove() {}, contains() { return false; }, toggle() {} },
  append(...c) { n.children.push(...c); }, appendChild(c) { n.children.push(c); return c; }, replaceChildren(...c) { n.children = c; }, setAttribute(k, v) { n.attrs[k] = v; }, getAttribute(k) { return n.attrs[k]; }, removeAttribute(k) { delete n.attrs[k]; },
  addEventListener(e, f) { (n.listeners[e] = n.listeners[e] || []).push(f); }, focus() { doc.activeElement = n; }, contains() { return false; }, querySelectorAll() { return []; }, click() { return fire(n, 'click'); } };
  return n; };
const doc = { activeElement: null, body: mk('body'), createElement: mk, addEventListener() {} };
const fire = async (n, e, arg) => { for (const f of n.listeners[e] || []) await f(arg || { key: '', preventDefault() {}, target: n }); await tick(); };
const tick = () => new Promise((r) => setTimeout(r, 5));
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const root = mk('div'); doc.body.append(root);
const find = (k) => all(root).find((n) => n.dataset && n.dataset.key === k);
const texts = () => all(root).filter((n) => !n.hidden).map((n) => n.textContent).filter(Boolean);
const has = (s) => texts().some((t) => t.includes(s));
const sheet = () => all(root).find((n) => n.className === 'vt-sheet');

const bar = (opts, def, total) => ({ total, default: def, line: 'line', options: opts });
const o = (id, x) => Object.assign({ id, label: id, line: id + ' line', ok: true, count: 4, ids: ['v1', 'v2', 'v3', 'v4'], cost: 0, skipped: [] }, x);
let act = null;
let items = [];
const view = () => ({ schema: 'watch.vault/1', mode: 'on', hero: 'Sorc', total: items.length, offset: 0, limit: 60, items, kinds: [], types: [], act,
  capacity: { count: 4, capacity: 400, pins: 0, pinLimit: 0, percent: 1, level: 'ok' }, upgrade: { available: false, currency: 'Ether', reason: 'x', max: 2400 } });
const rev = (b, gold) => ({ enabled: true, charges: 0, free: false, perItem: 1, balance: 665, queued: 0, working: 0, done: 0, bar: b, gold });
const sell = { enabled: true, queued: 0, working: 0, sold: 0, gold: 0 };
let posts = [], reply = null;
const get = async (p) => (p === '/api/ether' ? { schema: 'watch.ether/1', enabled: false } : view());
const post = async (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return typeof reply === 'function' ? reply(p, b) : reply; };
const later = (f) => { later.q.push(f); }; later.q = [];
const page = mountVault({ doc, root, get, post, later, texts: T, now: () => 0 });
const bartext = () => all(find('reveal-bar-go') ? root : root).filter((n) => !n.hidden).map((n) => n.textContent).join(' | ');

// gold is the cheapest full payer: it is picked, the title and the one press say what it costs; no notice paragraph
act = { reveal: rev(bar([o('free', { ok: false, count: 0, ids: [], why: 'no scroll' }), o('ether', { count: 2, ids: ['v1', 'v2'], cost: 2, skipped: [{ n: 2, why: 'the vault holds no affixes for them' }] }), o('gold', { cost: 320, line: 'Gold: 320 (stash 700)' })], 'gold', 4), { on: true, order: null, stash: 700 }), sell };
await page.start(); await tick();
const rbx = all(root).find((n) => n.className === 'vt-revealbar');
ok(rbx && !rbx.hidden, 'the bar shows');
ok(has('Reveal 4 items in this filter'), 'title');
ok(find('pay-gold').attrs['aria-checked'] === 'true' && find('pay-ether').attrs['aria-checked'] === 'false', 'default is gold');
ok(find('reveal-bar-go').textContent === 'Reveal 4 items for 320 gold (stash 700)', 'press words: ' + find('reveal-bar-go').textContent);
ok(!has('Cain is not rescued'), 'no notice paragraph');

// the hidden payments (user decision 2026-11-21): with methods ['gold'] the page draws only the gold option; the Free and Ether
// options, their buttons, the cost line and the scroll notices are absent or hidden, and the gold press words stay
const act0 = act;
posts = [];
act = { reveal: Object.assign(rev(bar([o('gold', { cost: 320, line: 'Gold: 320 (stash 700)' })], 'gold', 4), { on: true, order: null, stash: 700 }), { methods: ['gold'], etherPer: 1, charges: 3, note: 'Cain is not rescued: scrolls', line: 'scrolls: 3' }), sell };
await page.reload(false);
ok(find('pay-gold') && !find('pay-free') && !find('pay-ether'), 'only the gold option is in the DOM');
ok(find('reveal-bar-go').textContent === 'Reveal 4 items for 320 gold (stash 700)', 'gold press words: ' + find('reveal-bar-go').textContent);
ok(find('reveal').hidden && find('reveal-ether').hidden && find('buy-scrolls').hidden, 'the scroll, Ether and Buy buttons are hidden');
ok(!has('Cain is not rescued') && !has('scrolls: 3') && !has('no scroll or tome charge'), 'no scroll or Cain text: ' + texts().join(' | '));
ok(has('Gold: 320 (stash 700)'), 'the gold line reads Gold: <total> (stash <S>)');
// the selection toolbar: ONE reveal button, labelled with gold; no charges, Ether or scroll word anywhere in the Reveal UI
items = ['v1', 'v2', 'v3'].map((id) => ({ id, code: 'cap', name: 'Cap', state: 'recorded', quality: 'magic', identified: false, kind: 'armor', value: 10, pin: false }));
act = { reveal: Object.assign(rev(bar([o('gold', { cost: 160, count: 2, line: 'Gold: 160 (stash 200)' })], 'gold', 3), { on: true, order: null, stash: 200, scrollGold: 80, confirmAt: 0 }), { methods: ['gold'], etherPer: 1, charges: 7, balance: 937, note: 'Cain is not rescued', line: 'Revealing 16 items: at your next town visit' }), sell };
await page.reload(false);
await find('select-toggle').click(); await find('select-all').click();
ok(find('reveal-gold') && !find('reveal-gold').hidden && find('reveal-gold').textContent === 'Reveal with gold (3)', 'the one gold button: ' + (find('reveal-gold') || {}).textContent);
ok(find('reveal').hidden && find('reveal-ether').hidden && find('buy-scrolls').hidden, 'Reveal (charges), Reveal with Ether and Buy are hidden');
ok(has('Reveal 3 items for 160 gold (stash 200); the stash pays for 2'), 'one gold cost line: ' + texts().join(' | '));
ok(!texts().some((t) => /\bether\b|charge|scroll|cain/i.test(t)), 'no Ether, charges, scroll or Cain text while gold-only: ' + texts().filter((t) => /\bether\b|charge|scroll|cain/i.test(t)).join(' | '));
posts = []; reply = { ok: true, done: ['v1', 'v2'] };
await find('reveal-gold').click();
ok(posts[0] && posts[0].startsWith('/api/vault/revealgold {"ids":["v1","v2"],"maxGold":160'), 'the press orders what the stash covers: ' + posts[0]);
items = []; posts = [];
// all three named: the old picker is back
act = { reveal: Object.assign(rev(bar([o('free', { ok: false, count: 0, ids: [], why: 'no scroll' }), o('gold', { cost: 320, line: 'Gold: 320 (stash 700)' })], 'gold', 4), { on: true, order: null, stash: 700 }), { methods: ['free', 'ether', 'gold'] }), sell };
await page.reload(false);
ok(find('pay-free') && find('pay-gold'), 'with methods free, ether, gold the free option shows again');
act = act0;
await page.reload(false);

// the Ether pick and its press go to the ordinary route
await find('pay-ether').click();
ok(find('reveal-bar-go').textContent === 'Reveal 2 with 2 Ether' && has('2 skipped') && has('the vault holds no affixes'), 'ether words and the skipped line: ' + find('reveal-bar-go').textContent);
reply = { ok: true, done: ['v1', 'v2'] };
await find('reveal-bar-go').click();
ok(posts[0] === '/api/vault/reveal {"ids":["v1","v2"],"ether":true}', 'ether post ' + posts[0]);

// gold under the threshold: one press posts the order with the quoted cap
posts = []; reply = { ok: true, done: ['v1', 'v2', 'v3', 'v4'], order: { id: 'g-1', state: 'revealed', line: 'Gold reveal done: 4 items revealed for 320 gold. Stash gold: 380' } };
await find('pay-gold').click();
await find('reveal-bar-go').click();
ok(posts[0] === '/api/vault/revealgold {"ids":["v1","v2","v3","v4"],"maxGold":320,"confirm":false}', 'gold post ' + posts[0]);
ok(has('Gold reveal done: 4 items revealed for 320 gold. Stash gold: 380'), 'the order line says what was paid and what the stash holds');

// over the confirm threshold: the first press only asks, Back posts nothing, Yes posts with confirm
posts = [];
act = { reveal: rev(bar([o('gold', { cost: 5800, confirm: true })], 'gold', 4), { on: true }), sell };
await page.reload(false);
await find('reveal-bar-go').click();
ok(posts.length === 0 && has('Pay 5,800 gold from your stash'), 'the confirm step asks first');
ok(find('reveal-bar-go').hidden, 'the press is replaced by the question');
await find('reveal-bar-back').click();
ok(posts.length === 0 && !find('reveal-bar-go').hidden, 'Back posts nothing');
await find('reveal-bar-go').click(); await find('reveal-bar-yes').click();
ok(posts[0] === '/api/vault/revealgold {"ids":["v1","v2","v3","v4"],"maxGold":5800,"confirm":true}', 'confirmed post ' + posts[0]);

// a payment not yet confirmed: the chain, the progress, Cancel; no second order is offered
posts = []; reply = { ok: true };
act = { reveal: rev(bar([o('gold', { ok: false, why: 'a gold reveal is already running', count: 0, ids: [] })], '', 4), { on: true, order: { id: 'g-1', state: 'queued', total: 4, want: 4, done: 0, paid: 0, line: 'x' } }), sell };
await page.reload(false);
ok(has('Queued') && has('Paid') && has('Revealed') && has('0 of 4 revealed'), 'chain and progress');
ok(find('reveal-bar-go').disabled, 'no second order');
await find('reveal-bar-cancel').click();
ok(posts[0] === '/api/vault/revealgold/cancel {}' && has('Gold reveal cancelled.'), 'cancel ' + posts[0]);

// a paid and revealed order: the stash gold after the payment, no cancel
act = { reveal: rev(bar([], '', 0), { on: true, order: { id: 'g-1', state: 'revealed', total: 4, want: 4, done: 4, paid: 5800, stashAfter: 1400, line: 'x' } }), sell };
await page.reload(false);
ok(find('reveal-bar-cancel').hidden && has('4 of 4 revealed') && has('5,800 gold paid') && has('Stash gold: 1,400'), 'revealed order shows the payment and the stash gold: ' + texts().join(' | '));

// a failed order shows why
act = { reveal: rev(bar([], '', 0), { on: true, order: { id: 'g-2', state: 'failed', total: 2, want: 2, done: 0, paid: 0, note: 'the gold was refused: short of gold', line: 'x' } }), sell };
await page.reload(false);
ok(has('Failed') && has('short of gold') && find('reveal-bar-cancel').hidden, 'failed order');

// a finished order has no cancel
act = { reveal: rev(bar([], '', 0), { on: true, order: { id: 'g-1', state: 'partial', total: 2, want: 4, done: 2, paid: 160, stashAfter: 3, note: 'short of gold', line: 'x' } }), sell };
await page.reload(false);
ok(find('reveal-bar-cancel').hidden && has('short of gold') && has('Done in part'), 'finished order');

console.log(JSON.stringify({ fails }));
