// uic4 (review c2c3 R6, R7, R11, R12): inventory.js and inventory-actions.js run together in one stub page, so the pending outline is
// checked on the cells inventory.js really builds, the menu's focus on the confirm step is checked, and a touch press or a drag
// never opens the menu. Usage: node c4.js web/inventory.js web/inventory-actions.js
const fs = require('fs'), vm = require('vm'), assert = require('assert');
let focused = null;
function mk(tag) {
  const e = { tag, className: '', dataset: {}, children: [], textContent: '', attrs: {}, listeners: {}, style: {}, hidden: false, title: '', type: '', parentNode: null, offsetWidth: 200, offsetHeight: 100 };
  e.classList = { add(c) { e.className = (e.className + ' ' + c).trim(); }, contains(c) { return e.className.split(/\s+/).includes(c); }, remove(c) { e.className = e.className.split(/\s+/).filter((x) => x && x !== c).join(' '); } };
  Object.defineProperty(e, 'firstChild', { get() { return e.children[0] || null; } });
  e.appendChild = (k) => { k.parentNode = e; e.children.push(k); return k; };
  e.removeChild = (k) => { e.children = e.children.filter((x) => x !== k); return k; };
  e.setAttribute = (k, v) => { e.attrs[k] = String(v); };
  e.getAttribute = (k) => (k in e.attrs ? e.attrs[k] : null);
  e.hasAttribute = (k) => k in e.attrs;
  e.removeAttribute = (k) => { delete e.attrs[k]; };
  e.addEventListener = (t, f) => { (e.listeners[t] = e.listeners[t] || []).push(f); };
  e.fire = (t, ev) => (e.listeners[t] || []).forEach((f) => f(Object.assign({ target: e, preventDefault() {} }, ev)));
  e.focus = () => { focused = e; };
  e.closest = (sel) => { for (let n = e; n; n = n.parentNode) if (sel === '[data-key]' && n.dataset && n.dataset.key) return n; return null; };
  e.contains = (x) => { for (let n = x; n; n = n.parentNode) if (n === e) return true; return false; };
  e.getBoundingClientRect = () => ({ left: 100, top: 100, right: 150, bottom: 140 });
  e.querySelector = (s) => { const c = s.slice(1); const f = (n) => { for (const k of n.children) { if (k.className.split(/\s+/).includes(c)) return k; const r = f(k); if (r) return r; } return null; }; return f(e); };
  e.querySelectorAll = (s) => { const out = []; const f = (n) => n.children.forEach((k) => { if (s === 'button.act' ? (k.tag === 'button' && k.className.split(/\s+/).includes('act')) : k.className.split(/\s+/).includes(s.slice(1))) out.push(k); f(k); }); f(e); return out; };
  return e;
}
function CustomEvent(type, init) { this.type = type; this.detail = init && init.detail; }
function page() {
  const els = {}, dl = {}, fetches = [];
  const get = (i) => els[i] || (els[i] = mk('div'));
  const body = mk('body'); body.dataset.canvas = '';
  const document = { body, documentElement: { clientWidth: 1000, clientHeight: 700 }, getElementById: get, createElement: mk, activeElement: null,
    addEventListener: (t, f) => { (dl[t] = dl[t] || []).push(f); },
    dispatchEvent: (ev) => { (dl[ev.type] || []).forEach((f) => f(ev)); return true; },
    querySelector(sel) {
      const m = /\[data-key="(.+)"\]/.exec(sel); if (!m) return null;
      let hit = null; const walk = (e) => { if (!hit && e.dataset.key === m[1]) hit = e; e.children.forEach(walk); };
      Object.keys(els).forEach((k) => walk(els[k])); return hit;
    },
    querySelectorAll: () => [] };
  let inv = null, act = null;
  const window = { addEventListener() {}, __invTest: (x) => { inv = x; }, __actTest: (x) => { act = x; } };
  const ctx = { document, window, console, JSON, Math, Object, Array, String, Number, Promise, setInterval() {}, location: { search: '' }, Image: function () {}, CustomEvent,
    fetch: (u, o) => { fetches.push({ u, o }); return Promise.reject(new Error('no network in the stub')); } };
  window.window = window;
  vm.runInNewContext(fs.readFileSync(process.argv[2], 'utf8').replace("typeof window !== 'undefined' && window.__invTest", 'window.__invTest'), ctx);
  vm.runInNewContext(fs.readFileSync(process.argv[3], 'utf8'), ctx);
  act.init();
  return { els, document, dl, inv, act, fetches, get body() { return body; }, press(t, ev) { (dl[t] || []).forEach((f) => f(Object.assign({ preventDefault() {} }, ev))); } };
}
const item = (id, name, x, extra) => Object.assign({ id, code: 'rin', name, quality: 'magic', identified: true, x, y: 0, w: 1, h: 1 }, extra || {});
const view = (queue) => ({ schema: 10, class: 'Sorceress', inTown: false, gold: 1, stashGold: 2, equipment: {}, belt: [], stash: [], recent: [], queue: queue || [],
  backpack: [item(11, 'Ring', 0, { stashOp: 'stash', canWear: ['lring'] }), item(12, 'Staff', 1, { stashOp: 'stash' })] });
const tick = () => new Promise((r) => setImmediate(r));

(async () => {
  // R6: the dashed outline and the TOWN mark are on the cells inventory.js built, on the first paint and after every repaint
  {
    const p = page(), q = [{ id: 11, kind: 'stash', name: 'Ring', state: 'queued', wait: 'town', reason: 'waiting for a trip home' }];
    p.inv.paint(view(q)); await tick();
    let cell = p.els.pack.children.find((c) => c.dataset.key === 'item-11');
    assert.ok(cell, 'the cell exists');
    assert.strictEqual(cell.dataset.pending, 'stash', 'first paint: the queued stash move is outlined');
    assert.strictEqual(cell.dataset.wait, 'town', 'first paint: the TOWN mark');
    const old = cell;
    p.inv.paint(view(q.concat([{ id: 12, kind: 'stash', name: 'Staff', state: 'queued' }]))); await tick();
    cell = p.els.pack.children.find((c) => c.dataset.key === 'item-11');
    assert.notStrictEqual(cell, old, 'the repaint built new cells');
    assert.strictEqual(cell.dataset.pending, 'stash', 'after a repaint the outline is on the new cell');
    assert.strictEqual(cell.dataset.wait, 'town', 'after a repaint the TOWN mark is on the new cell');
    assert.strictEqual(p.els.pack.children.find((c) => c.dataset.key === 'item-12').dataset.pending, 'stash', 'a second request');
    p.inv.paint(view([])); await tick();
    cell = p.els.pack.children.find((c) => c.dataset.key === 'item-11');
    assert.ok(!('pending' in cell.dataset) && !('wait' in cell.dataset), 'a settled request leaves no outline');
  }
  // R7: the answer to the page's own request repaints the grids at once (d2:inventory-fresh), not at the next poll
  {
    const p = page(); p.inv.paint(view());
    const moved = view(); moved.backpack = moved.backpack.slice(1);
    p.document.dispatchEvent(new CustomEvent('d2:inventory-fresh', { detail: moved })); await tick();
    assert.ok(!p.els.pack.children.some((c) => c.dataset.key === 'item-11'), 'the ring is gone from the pack when the fresh view says so');
  }
  // R12: Delete opens the menu on its confirm step with the focus on "Keep it"; Enter there never trashes or equips
  {
    const p = page(); p.inv.paint(view()); await tick();
    const cell = p.els.pack.children.find((c) => c.dataset.key === 'item-11');
    p.press('keydown', { key: 'Delete', target: cell });
    const st = p.act.state();
    assert.strictEqual(st.key, 'item-11', 'Delete opens the menu');
    assert.ok(st.confirming, 'the menu is on the confirm step');
    assert.strictEqual(focused.dataset.act, 'back', 'the focus is on Keep it, not on Equip or the confirm button');
    focused.fire('click', {}); // what Enter on a focused button does in a browser
    assert.ok(!p.act.state().confirming, 'Keep it leaves the confirm step');
    assert.strictEqual(p.fetches.filter((f) => /trash|equip/.test(f.u || '')).length, 0, 'nothing was sent');
    // the first click on Trash puts the focus on Keep it too
    p.act.open('item-11');
    const trash = p.act.state().node.querySelectorAll('button.act').find((b) => b.dataset.act === 'trash');
    trash.fire('click', {});
    assert.strictEqual(focused.dataset.act, 'back', 'after the first click on Trash the focus is on Keep it');
  }
  // R11: a touch press or a drag in progress never opens the menu from contextmenu
  {
    const p = page(); p.inv.paint(view()); await tick();
    const cell = p.els.pack.children.find((c) => c.dataset.key === 'item-11');
    p.press('pointerdown', { pointerType: 'touch' });
    p.press('contextmenu', { target: cell, clientX: 5, clientY: 5 });
    assert.strictEqual(p.act.state().key, '', 'a touch long-press opens no menu');
    p.press('pointerdown', { pointerType: 'mouse' });
    p.body.setAttribute('data-dragging', '');
    p.press('contextmenu', { target: cell, clientX: 5, clientY: 5 });
    assert.strictEqual(p.act.state().key, '', 'no menu under a drag');
    p.body.removeAttribute('data-dragging');
    p.press('contextmenu', { target: cell, clientX: 5, clientY: 5 });
    assert.strictEqual(p.act.state().key, 'item-11', 'a mouse right-click opens it');
  }
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
