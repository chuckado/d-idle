// A small DOM stub for the character canvas tests (card uic2): enough of document and element for inventory-canvas.js and
// inventory-merc.js to paint into and for the result to be serialised. Not a browser: layout (clientHeight and the like) is 0,
// so the fit-to-box trimming is not exercised here; the fit check of tools/uishots.sh --fit does that in a real browser.
function mk(tag) {
  const e = { tag, className: '', dataset: {}, children: [], textContent: '', attrs: {}, listeners: {}, style: {}, hidden: false, title: '', parentNode: null, id: '',
    clientHeight: 0, scrollHeight: 0 };
  const cl = () => e.className.split(/\s+/).filter(Boolean);
  e.classList = { add(c) { if (!cl().includes(c)) e.className = (e.className + ' ' + c).trim(); }, remove(c) { e.className = cl().filter((x) => x !== c).join(' '); } };
  Object.defineProperty(e, 'firstChild', { get() { return e.children[0] || null; } });
  Object.defineProperty(e, 'lastChild', { get() { return e.children[e.children.length - 1] || null; } });
  e.appendChild = (k) => { k.parentNode = e; e.children.push(k); return k; };
  e.removeChild = (k) => { e.children = e.children.filter((x) => x !== k); return k; };
  e.setAttribute = (k, v) => { e.attrs[k] = String(v); };
  e.getAttribute = (k) => (k in e.attrs ? e.attrs[k] : null);
  e.removeAttribute = (k) => { delete e.attrs[k]; };
  e.hasAttribute = (k) => k in e.attrs;
  e.addEventListener = (t, f) => { (e.listeners[t] = e.listeners[t] || []).push(f); };
  e.focus = () => {};
  const walk = (n, f) => { for (const c of n.children) { f(c); walk(c, f); } };
  const test = (n, sel) => {
    let m;
    if ((m = /^\[data-key="(.+)"\]$/.exec(sel))) return n.dataset.key === m[1];
    if ((m = /^\.([a-z0-9-]+)$/.exec(sel))) return n.className.split(/\s+/).includes(m[1]);
    if ((m = /^([a-z0-9]+)$/.exec(sel))) return n.tag === m[1];
    return false;
  };
  e.querySelector = (sel) => { let hit = null; for (const part of sel.split(',')) { walk(e, (c) => { if (!hit && test(c, part.trim())) hit = c; }); } return hit; };
  e.querySelectorAll = (sel) => { const out = []; for (const part of sel.split(',')) walk(e, (c) => { if (test(c, part.trim())) out.push(c); }); return out; };
  e.closest = () => null;
  return e;
}
function ser(e, d) {
  const attrs = Object.keys(e.attrs).sort().map((k) => k + '=' + e.attrs[k]);
  const dat = Object.keys(e.dataset).sort().map((k) => 'data-' + k + '=' + e.dataset[k]);
  const cls = e.className.split(/\s+/).filter(Boolean).sort().join('.');
  let out = '  '.repeat(d) + e.tag + (cls ? '.' + cls : '') + ' [' + attrs.concat(dat).join(' ') + ']' + (e.textContent ? ' ' + JSON.stringify(e.textContent) : '') +
    (e.title ? ' title=' + JSON.stringify(e.title) : '') + (e.hidden ? ' hidden' : '') + '\n';
  e.children.forEach((c) => { out += ser(c, d + 1); });
  return out;
}
module.exports = { mk, ser };
