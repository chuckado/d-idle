// The Vault header's links (card uinav): Back is the live cockpit, plus Gear and Skills. vault.js in a stub DOM.
import { readFileSync } from 'fs';
import { mountVault } from './vault.mjs';
const T = JSON.parse(readFileSync(new URL('./vault.json', import.meta.url), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const mk = (tag) => { const n = { tag, children: [], hidden: false, className: '', textContent: '', dataset: {}, attrs: {}, listeners: {}, disabled: false, value: '', id: '', style: { setProperty() {} },
  classList: { add(c) { n.className += ' ' + c; }, remove() {}, contains() { return false; }, toggle() {} },
  append(...c) { n.children.push(...c); }, appendChild(c) { c.parent = n; n.children.push(c); return c; }, replaceChildren(...c) { n.children = c; c.forEach((x) => { x.parent = n; }); }, setAttribute(k, v) { n.attrs[k] = v; }, getAttribute(k) { return n.attrs[k]; }, removeAttribute(k) { delete n.attrs[k]; },
  addEventListener(e, f) { (n.listeners[e] = n.listeners[e] || []).push(f); }, focus() {}, contains() { return false; }, querySelectorAll() { return []; } };
  return n; };
const doc = { activeElement: null, body: mk('body'), createElement: mk, addEventListener() {} };
const all = (e, o = []) => { o.push(e); (e.children || []).forEach((c) => all(c, o)); return o; };
const root = mk('div'); doc.body.append(root);
const get = async (p) => (p === '/api/ether' ? { schema: 'watch.ether/1', enabled: false } : { schema: 'watch.vault/1', mode: 'on', hero: 'Sorc', total: 0, offset: 0, limit: 60, items: [], kinds: [], types: [], act: null, marked: null,
  capacity: { count: 0, capacity: 400, pins: 0, pinLimit: 0, percent: 0, level: 'ok' }, upgrade: { available: false, currency: 'Ether', reason: 'x', max: 2400 } });
const later = () => ({}); const clear = () => {};
const page = mountVault({ doc, root, get, post: async () => ({ ok: true }), later, clear, texts: T, now: () => 0, backHref: '/watch?live=1' });
await page.start(); await new Promise((r) => setTimeout(r, 5));
const links = all(root).filter((n) => n.tag === 'a' && n.href);
const by = (h) => links.find((n) => n.href === h);
ok(by('/watch?live=1') && by('/watch?live=1').dataset.key === 'back', 'a Back link to the live cockpit');
ok(by('/watch/inventory') && by('/watch/inventory').textContent === 'Inventory', 'an Inventory link: ' + links.map((n) => n.href));
ok(by('/watch/skills') && by('/watch/skills').textContent === 'Skills', 'a Skills link');
ok(!links.some((n) => n.href === '/watch'), 'no link to the bare /watch');
console.log(JSON.stringify({ fails }));
