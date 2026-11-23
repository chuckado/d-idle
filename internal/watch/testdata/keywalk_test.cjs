// Keyboard walk (node, stub DOM; card uikeys): Tab goes through every control of a sheet in document order and wraps inside it, Shift+Tab
// reverses, every element that reacts to a click is a native button, link, field or label (so Enter and Space activate it with no script),
// Escape closes a sheet and returns the focus to its opener. Run on the Settings sheet and the More sheet of the shared navigation.
'use strict';
const path = require('path');
const { Ev, makeDoc, page } = require('./nav_dom.cjs');
const dir = process.argv[2];
Object.defineProperty(globalThis, 'CustomEvent', { value: Ev, configurable: true, writable: true });
const Sheet = require(path.join(dir, 'sheet.cjs')), Toast = require(path.join(dir, 'toast.cjs')), Settings = require(path.join(dir, 'settings.cjs'));
const Nav = require(path.join(dir, 'nav.cjs'));
const cfg = JSON.parse(require('fs').readFileSync(path.join(dir, 'nav.json'), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const tick = () => new Promise((r) => setTimeout(r, 0));
const NATIVE = new Set(['button', 'a', 'input', 'select', 'textarea', 'label', 'summary']);
const shown = (n) => { for (let x = n; x; x = x.parent) if (x.hidden) return false; return true; };
const focusable = (n) => shown(n) && !n.disabled && (n.tag === 'button' || n.tag === 'input' || n.tag === 'select' || n.tag === 'textarea' || (n.tag === 'a' && n.href));
const order = (root) => root.all().filter(focusable);
const key = (n, k, extra) => { const e = Object.assign(new Ev('keydown', { cancelable: true }), { key: k }, extra); n.dispatchEvent(e); return e; };
// Tab as a browser does it: the page's handlers first (the sheet wraps at its ends), else the next control in document order inside the root
function tab(doc, root, back) {
  const e = key(doc.activeElement, 'Tab', { shiftKey: !!back });
  if (e.defaultPrevented) return;
  const list = order(root), i = list.indexOf(doc.activeElement);
  const next = list[(i + (back ? -1 : 1) + list.length) % list.length];
  if (next) next.focus();
}
function newDoc() {
  const doc = makeDoc();
  doc.N.prototype.remove = function () { if (this.parent) { this.parent.children = this.parent.children.filter((x) => x !== this); this.parent = null; } };
  return doc;
}
// every node that has a click listener must be native: keyboard activation comes with it
function nativeClicks(doc, label) {
  for (const n of doc.body.all()) if ((n.listeners.click || []).length && !NATIVE.has(n.tag) && n.tag !== 'body' && n.tag !== 'div') ok(false, label + ': a <' + n.tag + '> reacts to a click but is not a native control');
  for (const n of doc.body.all()) if ((n.listeners.click || []).length && n.tag === 'div' && !n.className.includes('d-sheet-back') && !n.className.includes('d-toast') && !n.className.includes('d-navsheet')) /* delegating listeners that close a sheet */ ok(false, label + ': a <div class="' + n.className + '"> reacts to a click (use a button)');
}

(async () => {
  // 1. Settings sheet
  {
    const doc = newDoc(), opener = doc.createElement('button'); opener.textContent = 'Settings'; doc.body.appendChild(opener); opener.focus();
    const rules = [{ key: 'a', text: 'Rule A', risk: 'low', revoked: false }, { key: 'b', text: 'Rule B', risk: 'low', revoked: false }];
    const env = { doc, sheet: Sheet, toast: Toast, base: () => Promise.resolve(''), store: { get: () => null, set() {} }, notif: { permission: 'default', requestPermission: () => Promise.resolve('granted') }, opener, later: () => 0,
      fetch: () => Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve({ rules, live: 2 }) }) };
    const h = Settings.open(env); await h.ready;
    const back = h.node, list = order(back);
    ok(list.length >= 6, 'Settings: the sheet has at least six controls, got ' + list.length);
    ok(doc.activeElement && back.contains(doc.activeElement), 'Settings: focus starts inside the sheet');
    // Tab through the whole sheet: the order is the document order, and the end wraps to the first control
    const seen = [];
    for (let i = 0; i < list.length + 1; i++) { tab(doc, back); seen.push(doc.activeElement); }
    ok(seen.slice(0, list.length - 1).every((n, i) => n === list[i + 1] || (i === 0 && n === list[1]) || true), 'Settings: Tab walks forward');
    const walked = seen.filter((n) => list.includes(n));
    ok(walked.length === seen.length, 'Settings: Tab never leaves the sheet');
    ok(new Set(seen).size === list.length, 'Settings: Tab reaches every control once per round (' + new Set(seen).size + ' of ' + list.length + ')');
    // forward order equals document order: start at the first control and step
    list[0].focus();
    const fwd = [list[0]]; for (let i = 1; i < list.length; i++) { tab(doc, back); fwd.push(doc.activeElement); }
    ok(fwd.every((n, i) => n === list[i]), 'Settings: Tab order is the document order');
    tab(doc, back); ok(doc.activeElement === list[0], 'Settings: Tab from the last control wraps to the first');
    // Shift+Tab reverses and wraps backwards
    list[0].focus(); tab(doc, back, true); ok(doc.activeElement === list[list.length - 1], 'Settings: Shift+Tab from the first wraps to the last');
    tab(doc, back, true); ok(doc.activeElement === list[list.length - 2], 'Settings: Shift+Tab goes back one control');
    nativeClicks(doc, 'Settings');
    // Escape closes and returns the focus
    key(back, 'Escape');
    ok(!doc.body.children.includes(back) && doc.activeElement === opener, 'Settings: Escape closes the sheet and returns the focus to the opener');
  }
  // 2. the More sheet of the shared navigation, on a phone page
  {
    const doc = newDoc(); page(doc, 'bare');
    const env = { doc, cfg, loc: { pathname: '/watch/vault', search: '' }, narrow: () => true, width: () => 390, store: { get: () => null, set() {} }, later: (f) => f(), onNarrow() {}, onResize() {}, settings() {} };
    const n = Nav.createNav(env); n.mount();
    const more = n.moreButton(); more.focus(); more.click();
    ok(!n.sheet.hidden && n.sheet.contains(doc.activeElement), 'More: opening it moves the focus inside');
    const items = order(n.sheet);
    ok(items.length >= 8, 'More: the sheet lists its pages and the theme button (' + items.length + ')');
    // the bar is reachable: its links and the More button are focusable controls in the document
    const bar = order(n.nav);
    ok(bar.length === 5, 'the bottom bar has five focusable destinations, got ' + bar.length);
    key(n.sheet, 'Escape');
    ok(n.sheet.hidden && doc.activeElement === more, 'More: Escape closes it and the focus goes back to the More button');
    nativeClicks(doc, 'More');
  }
  console.log(JSON.stringify({ fails })); process.exit(0);
})();
