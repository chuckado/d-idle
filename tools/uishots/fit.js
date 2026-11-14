// The fit check's measurements (tools/uishots fit.go): run from the harness page on a page of the same origin (an iframe of
// the exact viewport size) and returns plain numbers; fit.go judges them. No page code is touched.
// opts.scroll: the ids of the panels whose manifest says scroll: true (the only place a data-scroll list is accepted).
export function measure(win, opts) {
  const d = win.document, se = d.scrollingElement, vw = win.innerWidth, vh = win.innerHeight, decl = new Set((opts && opts.scroll) || []);
  const vis = (n) => { const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  const cs = (n) => win.getComputedStyle(n);
  // a list that scrolls on purpose: marked data-scroll, inside a panel that declares scroll, and really a scroll container
  const scroller = (n) => n.hasAttribute('data-scroll') && /^(auto|scroll)$/.test(cs(n).overflowY) && decl.has((n.closest('[data-panel]') || { dataset: {} }).dataset.panel);
  const name = (n) => { const p = n.closest('[data-panel]'); return (p ? p.dataset.panel : '?') + ' ' + n.tagName.toLowerCase() + (n.className && typeof n.className === 'string' ? '.' + n.className.trim().split(/\s+/)[0] : ''); };
  const over = [];
  const test = (n, label) => {
    if (!n.clientHeight && !n.clientWidth) return; // inline boxes and SVG parts have no client box
    const dh = n.scrollHeight - n.clientHeight, dw = n.scrollWidth - n.clientWidth;
    const ell = cs(n).textOverflow === 'ellipsis'; // one line cut with an ellipsis: width only, by design
    const bad = (dh > 1 && !scroller(n)) || (dw > 1 && !ell && !scroller(n));
    if (bad) over.push({ el: label, dh: dh > 1 ? dh : 0, dw: dw > 1 && !ell ? dw : 0 });
  };
  for (const n of d.querySelectorAll('[data-cockpit-slot]')) if (vis(n)) test(n, 'slot ' + n.dataset.cockpitSlot);
  for (const root of d.querySelectorAll('[data-cockpit-panel]')) {
    if (!vis(root)) continue;
    test(root, 'panel ' + (root.dataset.panel || '?'));
    // every descendant too: content cut one level below the root is still cut. A scroll container's own children are its business.
    for (const n of root.querySelectorAll('*')) {
      if (!vis(n) || n.closest('[data-scroll]') && n.closest('[data-scroll]') !== n && scroller(n.closest('[data-scroll]'))) continue;
      test(n, name(n));
    }
  }
  // boxes that end outside the window: a frame taller or wider than the viewport that something clips
  const out = [];
  for (const n of d.querySelectorAll('[data-cockpit-slot], [data-cockpit-panel]')) {
    const r = vis(n);
    if (r && (r.right > vw + 1 || r.bottom > vh + 1 || r.left < -1 || r.top < -1)) out.push({ el: n.dataset.cockpitSlot ? 'slot ' + n.dataset.cockpitSlot : 'panel ' + (n.dataset.panel || '?'), right: Math.round(r.right), bottom: Math.round(r.bottom) });
  }
  // a slot (the band and the stage apart) that shows no panel is a blank hole, and a panel still loading or failed is not a fit
  const empty = [], unready = [];
  for (const n of d.querySelectorAll('[data-cockpit-slot]')) {
    if (!vis(n) || n.dataset.cockpitSlot === 'band' || n.dataset.cockpitSlot === 'stage') continue;
    if (![...n.querySelectorAll('[data-cockpit-panel]')].some((p) => vis(p))) empty.push(n.dataset.cockpitSlot);
  }
  for (const p of d.querySelectorAll('[data-cockpit-panel]')) if (vis(p) && /^(loading|error)$/.test(p.dataset.state || '')) unready.push(p.dataset.panel + ':' + p.dataset.state);
  // dead space: the frame at 2 px cells; covered by the band, the stage (its letterbox is the one named exception), the visible
  // panel boxes and the phone tab bar. The largest empty square must not be wider than a gap.
  const cs2 = 2, cols = Math.ceil(vw / cs2), rows = Math.ceil(vh / cs2), cov = new Uint8Array(cols * rows);
  const slots = [];
  const mark = (r) => {
    for (let y = Math.max(0, Math.floor(r.top / cs2)); y < Math.min(rows, Math.ceil(r.bottom / cs2)); y++) {
      for (let x = Math.max(0, Math.floor(r.left / cs2)); x < Math.min(cols, Math.ceil(r.right / cs2)); x++) cov[y * cols + x] = 1;
    }
  };
  for (const n of d.querySelectorAll('[data-cockpit-slot]')) {
    const r = vis(n);
    if (!r) continue;
    slots.push({ slot: n.dataset.cockpitSlot, x: Math.round(r.left), y: Math.round(r.top), w: Math.round(r.width), h: Math.round(r.height) });
    if (n.dataset.cockpitSlot === 'band' || n.dataset.cockpitSlot === 'stage') mark(r);
  }
  for (const n of d.querySelectorAll('[data-cockpit-panel], .d-tabbar')) { const r = vis(n); if (r) mark(r); }
  let covered = 0;
  for (let i = 0; i < cov.length; i++) covered += cov[i];
  let best = 0, bx = 0, by = 0, prev = new Int16Array(cols + 1), cur = new Int16Array(cols + 1);
  for (let y = 0; y < rows; y++) {
    cur.fill(0);
    for (let x = 0; x < cols; x++) {
      if (cov[y * cols + x]) continue;
      const s = Math.min(prev[x], prev[x + 1], cur[x]) + 1;
      cur[x + 1] = s;
      if (s > best) { best = s; bx = x; by = y; }
    }
    [prev, cur] = [cur, prev];
  }
  return {
    vw, vh, scrollH: se.scrollHeight, scrollW: se.scrollWidth, overflow: over, outside: out, empty, unready, slots,
    covered: Math.round((covered * 100) / (cols * rows)), hole: best * cs2, holeAt: [(bx - best + 1) * cs2, (by - best + 1) * cs2],
  };
}

// An opened sheet must be hittable: the element at the centre of each target is the target or inside it.
export function hits(win, selectors) {
  const d = win.document, out = {};
  for (const sel of selectors) {
    const n = d.querySelector(sel), r = n && n.getBoundingClientRect();
    if (!r || !r.width) { out[sel] = false; continue; }
    const at = d.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
    out[sel] = !!at && (at === n || n.contains(at));
  }
  return out;
}
