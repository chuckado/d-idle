// The overlay manager (web/overlay.js) in a stub DOM: one overlay open at a time, in both directions, for every pair of kinds;
// Escape closes the top-most and gives the focus back to the opener; a closed overlay does not steal Escape.
// Usage: node overlay_dom.mjs path/to/overlay.mjs
import assert from 'node:assert';
import { pathToFileURL } from 'node:url';
const { createOverlays } = await import(pathToFileURL(process.argv[2]).href);

function stub() {
  const listeners = {};
  const doc = { addEventListener(t, f) { (listeners[t] = listeners[t] || []).push(f); }, fire(t, e) { (listeners[t] || []).forEach((f) => f(e)); } };
  const node = (name) => { const n = { name, hidden: true, closed: 0, children: [], contains(x) { return x === n || n.children.includes(x); }, focus() { n.focused = (n.focused || 0) + 1; doc.fire('focusin', { target: n }); }, isConnected: true }; return n; };
  return { doc, node, listeners };
}
const names = ['menu', 'sheet-zone-map', 'sheet-orders', 'drawer-log', 'drawer-onboarding', 'reset'];

// every ordered pair: open A, then B: only B is open, A got its close called
for (const a of names) for (const b of names) {
  if (a === b) continue;
  const { doc, node } = stub(), o = createOverlays(doc, null), els = {};
  for (const n of names) { els[n] = node(n); o.register(n, els[n], () => { els[n].closed++; els[n].hidden = true; }); }
  els[a].hidden = false; o.sync();
  assert.deepStrictEqual(o.open(), [a], 'a alone: ' + a);
  els[b].hidden = false; o.sync();
  assert.deepStrictEqual(o.open(), [b], a + ' then ' + b + ': only the last is open');
  assert.strictEqual(els[a].closed, 1, a + ' was closed through its own close()');
  assert.strictEqual(els[a].hidden, true);
}

// Escape closes the top-most and returns the focus to the opener (the button that was focused outside every overlay)
{
  const { doc, node } = stub(), o = createOverlays(doc, null), more = node('more'), menu = node('menu'), sheet = node('sheet');
  o.register('menu', menu, () => { menu.hidden = true; });
  o.register('sheet', sheet);
  more.focus(); // the opener
  menu.hidden = false; o.sync();
  const ev = { key: 'Escape', defaultPrevented: false, prevented: false, preventDefault() { this.prevented = true; this.defaultPrevented = true; } };
  const before = more.focused;
  doc.fire('keydown', ev);
  assert.deepStrictEqual(o.open(), [], 'Escape closed the menu');
  assert.ok(ev.prevented);
  assert.strictEqual(more.focused, before + 1, 'focus went back to the opener');
  // nothing open: Escape is not eaten
  const ev2 = { key: 'Escape', defaultPrevented: false, preventDefault() { this.defaultPrevented = true; } };
  doc.fire('keydown', ev2);
  assert.strictEqual(ev2.defaultPrevented, false, 'a closed manager leaves Escape alone');
  // an overlay that already handled Escape (defaultPrevented) is not closed twice
  sheet.hidden = false; o.sync();
  doc.fire('keydown', { key: 'Escape', defaultPrevented: true, preventDefault() {} });
  assert.deepStrictEqual(o.open(), ['sheet']);
  // Escape closes only the top-most: menu opens over a sheet that stayed (impossible by policy, so closed by it)
  menu.hidden = false; o.sync();
  assert.deepStrictEqual(o.open(), ['menu']);
  assert.strictEqual(sheet.hidden, true);
}

// a closed overlay opened again counts as new (its sequence restarts)
{
  const { doc, node } = stub(), o = createOverlays(doc, null), a = node('a'), b = node('b');
  o.register('a', a); o.register('b', b);
  a.hidden = false; o.sync(); a.hidden = true; o.sync(); b.hidden = false; o.sync(); a.hidden = false; o.sync();
  assert.deepStrictEqual(o.open(), ['a']);
  assert.strictEqual(b.hidden, true);
}
console.log('overlay_dom ok');
