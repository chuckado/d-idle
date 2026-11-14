// The fit check's measurements (tools/uishots fit.go): run inside a page of the same origin (an iframe of the exact
// viewport size) and returns plain numbers; fit.go judges them. No page code is touched.
export function measure(win) {
  const d = win.document, se = d.scrollingElement, vw = win.innerWidth, vh = win.innerHeight;
  const vis = (n) => { const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  // an element or one of its ancestors that scrolls on purpose (a list marked data-scroll) is excepted
  const excepted = (n) => !!n.closest('[data-scroll]');
  const over = [];
  for (const n of d.querySelectorAll('[data-slot], [data-slotpanel]')) {
    if (!vis(n) || excepted(n)) continue;
    const dh = n.scrollHeight - n.clientHeight, dw = n.scrollWidth - n.clientWidth;
    if (dh > 1 || dw > 1) over.push({ el: n.dataset.slot ? 'slot ' + n.dataset.slot : 'panel ' + (n.dataset.panel || '?'), dh, dw });
  }
  // dead space: the frame at 2 px cells, every slot's rectangle marked; the largest empty square must not be wider than a gap
  const cs = 2, cols = Math.ceil(vw / cs), rows = Math.ceil(vh / cs), cov = new Uint8Array(cols * rows);
  const slots = [];
  let covered = 0;
  for (const n of d.querySelectorAll('[data-slot]')) {
    const r = vis(n);
    if (!r) continue;
    slots.push({ slot: n.dataset.slot, x: Math.round(r.left), y: Math.round(r.top), w: Math.round(r.width), h: Math.round(r.height) });
    for (let y = Math.max(0, Math.floor(r.top / cs)); y < Math.min(rows, Math.ceil(r.bottom / cs)); y++) {
      for (let x = Math.max(0, Math.floor(r.left / cs)); x < Math.min(cols, Math.ceil(r.right / cs)); x++) cov[y * cols + x] = 1;
    }
  }
  for (let i = 0; i < cov.length; i++) covered += cov[i];
  let best = 0, bx = 0, by = 0, prev = new Int16Array(cols + 1), cur = new Int16Array(cols + 1);
  for (let y = 0; y < rows; y++) {
    cur.fill(0);
    for (let x = 0; x < cols; x++) {
      if (cov[y * cols + x]) continue;
      const s = Math.min(prev[x], prev[x + 1], cur[x]) + 1; // largest empty square ending here (cells)
      cur[x + 1] = s;
      if (s > best) { best = s; bx = x; by = y; }
    }
    [prev, cur] = [cur, prev];
  }
  return {
    vw, vh, scrollH: se.scrollHeight, scrollW: se.scrollWidth, overflow: over, slots,
    covered: Math.round((covered * 100) / (cols * rows)), hole: best * cs, holeAt: [(bx - best + 1) * cs, (by - best + 1) * cs],
  };
}
