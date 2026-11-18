// The overlay manager (docs/ui-overlays.md): ONE transient overlay is open at a time. The "..." menu, the zone and More
// sheets, the drawers and the Reset confirm register their element here; when one opens (its `hidden` goes false) every other
// one is closed, and Escape closes the one opened last and gives the focus back to what opened it. The overlays keep their own
// open and close code (they set `hidden`); the manager only watches. Layers: the "..." menu is a browser top-layer popover, the
// sheets and drawers are z-index layers of tokens.css (--d-z-*), see docs/ui-overlays.md.

export function createOverlays(doc, Observer) {
  const all = new Map(); // id -> { el, close, opener, seq }
  let seq = 0, outside = null; // outside: the last focused element that is in no overlay (what Escape returns to)
  const isOpen = (o) => !o.el.hidden;
  const inside = (n) => { for (const o of all.values()) if (o.el.contains && o.el.contains(n)) return true; return false; };
  const openList = () => [...all.entries()].filter(([, o]) => isOpen(o)).sort((a, b) => a[1].seq - b[1].seq).map(([id]) => id);

  // sync notices which overlays are open now; a newly opened one closes the rest. Called by the observer and by tests.
  function sync() {
    const open = [...all.values()].filter(isOpen);
    for (const o of all.values()) if (!isOpen(o)) o.seq = 0;
    const fresh = open.filter((o) => !o.seq);
    for (const o of fresh) { o.seq = ++seq; o.opener = outside; }
    if (fresh.length) {
      const keep = fresh[fresh.length - 1]; // several at once: the last registered wins
      for (const o of open) if (o !== keep) { closeOne(o, false); }
    }
  }
  function closeOne(o, back) {
    const opener = o.opener;
    o.seq = 0; o.opener = null;
    if (o.close) o.close(); else o.el.hidden = true;
    o.el.hidden = true;
    if (back && opener && opener.focus && opener.isConnected !== false) opener.focus({ preventScroll: true });
  }
  function register(id, el, close) {
    all.set(id, { el, close, opener: null, seq: 0 });
    if (Observer) new Observer(sync).observe(el, { attributes: true, attributeFilter: ['hidden'] });
    sync();
  }
  function escape() { const top = openList().pop(); if (!top) return false; closeOne(all.get(top), true); return true; }
  function closeAll() { for (const o of all.values()) if (isOpen(o)) closeOne(o, false); }

  if (doc && doc.addEventListener) {
    doc.addEventListener('focusin', (e) => { if (e.target && !inside(e.target)) outside = e.target; });
    // a press that does not focus its button (Safari) still names the opener
    doc.addEventListener('click', (e) => { const b = e.target && e.target.closest && e.target.closest('button'); if (b && !inside(b)) outside = b; }, true);
    doc.addEventListener('keydown', (e) => { if (e.key === 'Escape' && !e.defaultPrevented && escape()) e.preventDefault(); });
  }
  return { register, sync, escape, closeAll, open: openList };
}

export const overlays = createOverlays(typeof document !== 'undefined' ? document : null, typeof MutationObserver !== 'undefined' ? MutationObserver : null);
