// The Vault page against a fake bot with an ETHER wallet (TestVaultEtherUIJS).
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

const vaultView = (count, capacity) => ({ schema: 'watch.vault/1', mode: 'on', hero: 'Sorc', total: 0, offset: 0, limit: 60, items: [], kinds: [], types: [],
  capacity: { count, capacity, pins: 0, pinLimit: 0, percent: Math.floor(100 * count / capacity), level: 'ok' }, upgrade: { available: false, currency: 'Ether', reason: 'needs the vault switched on', max: 2400 } });
let wallet = null, vcount = 100, vcap = 400, posts = [], reply = null;
const eth = (balance, owned) => ({ schema: 'watch.ether/1', enabled: true, currency: 'Ether', balance, chip: 'Ether ' + balance, owned, capacity: vcap,
  next: owned.length ? { id: 'v2', add: 300, price: 40, canAfford: balance >= 40, short: Math.max(0, 40 - balance), capacity: vcap + 300 } : { id: 'v1', add: 200, price: 20, canAfford: balance >= 20, short: Math.max(0, 20 - balance), capacity: vcap + 200 },
  downgrade: owned.length ? (vcount <= vcap - 200 ? { available: true, step: owned[owned.length - 1], refund: 20 } : { available: false, step: 'v1', refund: 20, reason: 'the vault holds ' + vcount + ' records and the lower tier fits ' + (vcap - 200) + ': discard ' + (vcount - vcap + 200) + ' first' }) : { available: false, reason: 'no capacity step is owned' } });
const get = async (p) => { if (p === '/api/ether') return wallet && typeof wallet === 'function' ? wallet() : wallet; return vaultView(vcount, vcap); };
const post = async (p, b) => { posts.push(p + ' ' + JSON.stringify(b)); return typeof reply === 'function' ? reply(p, b) : reply; };
const later = (f) => { later.q.push(f); }; later.q = [];
const page = mountVault({ doc, root, get, post, later, texts: T, now: () => 0 });
await page.start(); await tick();

// wallet off (loot.vault 0): no chip, the placeholder stands, a press says why and posts nothing
wallet = { schema: 'watch.ether/1', enabled: false };
await page.reload(false);
ok(find('wallet').hidden, 'no chip without a wallet');
ok(find('upgrade').attrs['aria-disabled'] === 'true' && find('downgrade').hidden, 'placeholder buttons');
await find('upgrade').click();
ok(posts.length === 0 && has('needs the vault switched on'), 'a placeholder press posts nothing and says why');

// a wallet with 35 Ether: the chip, a real Upgrade, no Refund yet
wallet = eth(35, []);
await page.reload(false);
ok(!find('wallet').hidden && find('wallet').textContent === 'Ether 35', 'wallet chip');
ok(find('upgrade').attrs['aria-disabled'] === 'false' && find('upgrade').textContent === 'Upgrade +200 for 20 Ether', 'upgrade enabled with the price: ' + find('upgrade').textContent);
ok(find('downgrade').hidden, 'no refund with nothing owned');

// the confirm sheet: words, cancel posts nothing
await find('upgrade').click();
ok(!sheet().hidden && has('capacity 400 becomes 600') && has('you have 35, 15 after'), 'confirm sheet words');
await find('sheet-cancel').click();
ok(sheet().hidden && posts.length === 0, 'cancel posts nothing');

// pay: the post carries the step, the page shows the read-back, not the hope
await find('upgrade').click();
reply = () => { vcap = 600; wallet = eth(15, ['v1']); return { ok: true, changed: true, step: 'v1', price: 20, balance: 15, capacity: 600 }; };
await find('sheet-go').click();
ok(posts[0] === '/api/vault/upgrade {"step":"v1"}', 'post body ' + posts[0]);
ok(find('wallet').textContent === 'Ether 15' && has('Capacity is now 600. 15 Ether left.'), 'read back after the press');
ok(sheet().hidden, 'sheet closed');

// a short balance: aria-disabled with the number, a press says it, no sheet
ok(find('upgrade').attrs['aria-disabled'] === 'true' && has('Need 25 more Ether'), 'short balance reason');
posts = [];
await find('upgrade').click();
ok(sheet().hidden && posts.length === 0, 'no sheet and no post when short');

// refund does not fit: the reason is shown and nothing is posted; then it fits and the sheet shows the refund
vcount = 550; wallet = eth(15, ['v1']); await page.reload(false);
ok(!find('downgrade').hidden && find('downgrade').attrs['aria-disabled'] === 'true' && has('discard 150 first'), 'refund refused with the reason');
await find('downgrade').click();
ok(sheet().hidden && posts.length === 0, 'refused refund opens nothing');
vcount = 100; wallet = eth(15, ['v1']); await page.reload(false);
ok(find('downgrade').attrs['aria-disabled'] === 'false' && find('downgrade').textContent === 'Refund last step for 20 Ether', 'refund label');
await find('downgrade').click();
ok(has('Gives back step v1 for 20 Ether (15 becomes 35)') && has('holds 100 records'), 'refund sheet words');
reply = () => { vcap = 400; wallet = eth(35, []); return { ok: true, changed: true, step: 'v1', price: 20, balance: 35, capacity: 400 }; };
await find('sheet-go').click();
ok(posts[0] === '/api/vault/downgrade {"step":"v1"}' && has('Refunded. Capacity is now 400. 35 Ether in the wallet.'), 'refund pressed and read back');

// refusals from the server are never drawn as success and carry their numbers
wallet = eth(35, []); await page.reload(false);
await find('upgrade').click();
reply = { ok: false, code: 'insufficient', need: 7, error: 'raw' };
await find('sheet-go').click();
ok(has('Not enough Ether: 7 more needed.') && !has('Capacity is now'), '402 text');
wallet = eth(35, ['v1']); vcount = 560; vcap = 600; await page.reload(false);
reply = { ok: false, code: 'in-use', count: 560, fitsTier: 400, error: 'raw' };
await page.pressEther('downgrade', 'v1');
ok(has('The vault holds 560 records and the lower tier fits 400: discard 160 first.'), '409 in-use text');
reply = null; await page.pressEther('upgrade', 'v2');
ok(has(T.noAnswer), 'no answer text');
reply = { ok: true, already: true, step: 'v1', balance: 35, capacity: 600 }; await page.pressEther('upgrade', 'v1');
ok(has(T.etherAlready), 'an already-owned step is said, not charged');
reply = { ok: false, code: 'zzz', error: 'the server says' }; await page.pressEther('upgrade', 'v1');
ok(has('the server says'), 'unknown code falls back to the server sentence');

// the poll refreshes the balance when only it moved
wallet = eth(99, ['v1']); later.q.splice(0).forEach((f) => f()); await tick(); await tick();
ok(find('wallet').textContent === 'Ether 99', 'poll refreshes the chip: ' + find('wallet').textContent);

console.log(JSON.stringify({ fails }));
