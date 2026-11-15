// The fit check's measurements (tools/uishots fit.go): run from the harness page on a page of the same origin (an iframe of
// the exact viewport size) and returns plain numbers; fit.go judges them. No page code is touched.
// opts.scroll: the ids of the panels whose manifest says scroll: true (the only place a data-scroll list is accepted).
export function measure(win, opts) {
  const d = win.document, se = d.scrollingElement, vw = win.innerWidth, vh = win.innerHeight, decl = new Set((opts && opts.scroll) || []);
  const vis = (n) => { const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  const cs = (n) => win.getComputedStyle(n);
  // a list that scrolls on purpose: marked data-scroll, inside a panel that declares scroll, and really a scroll container
  const scroller = (n) => n.hasAttribute('data-scroll') && /^(auto|scroll)$/.test(cs(n).overflowY) && decl.has((n.closest('[data-panel]') || { dataset: {} }).dataset.panel);
  // a few words of the element's own text name it in the verdict (a cell, a slot, a chip), so the fix is found without a debugger
  const name = (n) => { const p = n.closest('[data-panel]'); return (p ? p.dataset.panel : '?') + ' ' + n.tagName.toLowerCase() + (n.className && typeof n.className === 'string' ? '.' + n.className.trim().split(/\s+/)[0] : '') + (n.children.length < 3 && n.textContent ? ' "' + n.textContent.trim().slice(0, 16) + '"' : '') + ' [' + n.scrollHeight + '/' + n.clientHeight + ' ' + n.scrollWidth + '/' + n.clientWidth + ']'; };
  const over = [];
  const test = (n, label) => {
    if (!n.clientHeight && !n.clientWidth) return; // inline boxes and SVG parts have no client box
    const dh = n.scrollHeight - n.clientHeight, dw = n.scrollWidth - n.clientWidth;
    const ell = cs(n).textOverflow === 'ellipsis'; // one line cut with an ellipsis: width only, by design
    const bad = (dh > 1 && !scroller(n)) || (dw > 1 && !ell && !scroller(n));
    if (bad) over.push({ el: label, dh: dh > 1 ? dh : 0, dw: dw > 1 && !ell ? dw : 0 });
  };
  for (const n of d.querySelectorAll('[data-cockpit-slot]')) if (vis(n)) test(n, 'slot ' + n.dataset.cockpitSlot);
  for (const root of d.querySelectorAll('[data-cockpit-panel], [data-fit-deep]')) {
    if (!vis(root)) continue;
    test(root, root.dataset.fitDeep !== undefined ? 'band' : 'panel ' + (root.dataset.panel || '?'));
    // every descendant too: content cut one level below the root is still cut. A scroll container's own children are its business.
    for (const n of root.querySelectorAll('*')) {
      if (!vis(n) || n.closest('[data-scroll]') && n.closest('[data-scroll]') !== n && scroller(n.closest('[data-scroll]'))) continue;
      test(n, name(n));
      // the deep band: a part that reaches out of the band's box is named (the band is not clipped itself, so only its parts can say what is cut)
      if (root.dataset.fitDeep !== undefined) {
        const rr = root.getBoundingClientRect(), nr = n.getBoundingClientRect();
        if (nr.bottom > rr.bottom + 1 || nr.top < rr.top - 1) over.push({ el: 'band part ' + name(n), dh: Math.round(Math.max(nr.bottom - rr.bottom, rr.top - nr.top)), dw: 0 });
      }
    }
  }
  // The character screen (card uic4, review c2c3 R14): checks the first fit check let through. Only on a page that has the canvas frame.
  const cvx = d.body.hasAttribute('data-canvas') ? charChecks(win, d, vis, cs) : null;
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
  // the zone map (S5): every zone row keeps 20 px and a readable name, all rows are on show, and the open zone menu (a sheet in
  // another tile) covers none of the zone buttons
  const zb = [...d.querySelectorAll('[data-panel="zone-map"] .zm-n:not(.zm-row)')].map((n) => [n, vis(n)]).filter((x) => x[1]);
  const sheet = d.querySelector('.d-tilesheet[data-panel="zone-map"]:not([hidden])'), sr = sheet && vis(sheet);
  const tops = new Set(zb.map((x) => Math.round(x[1].top)));
  const zones = { n: zb.length, rows: tops.size, minH: zb.length ? Math.round(Math.min(...zb.map((x) => x[1].height))) : 0,
    cut: zb.filter((x) => { const nm = x[0].querySelector('.zm-nm'); return nm && nm.clientWidth < 24; }).length, menu: !!sr, host: sheet ? sheet.dataset.host : '',
    overlap: sr ? zb.filter((x) => x[1].left < sr.right - 1 && x[1].right > sr.left + 1 && x[1].top < sr.bottom - 1 && x[1].bottom > sr.top + 1).length : 0 };
  return {
    zones, vw, vh, scrollH: se.scrollHeight, scrollW: se.scrollWidth, overflow: over, outside: out, cv: cvx, empty, unready, slots,
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

// What the first fit check missed on the character screen (review c2c3, R1 R2 R3 R4 R8 R10 R13):
//  zero:  a box with content and no width (the storage side column on a phone: 0 px wide, nothing says so)
//  cut:   a value or label of a values tile that an ellipsis cut ("L...", "3..."): a number never reads cut
//  tiles: a values tile far taller than what it holds (148 by 157 px for "Level 14")
//  inner: the largest empty square inside a panel, where painted things are text, controls, cells, tiles, lists
function charChecks(win, d, vis, cs) {
  const zero = [], cut = [], tiles = [];
  const label = (n) => { const p = n.closest('[data-panel]'); return (p ? p.dataset.panel : '?') + ' ' + n.tagName.toLowerCase() + (typeof n.className === 'string' && n.className ? '.' + n.className.trim().split(/\s+/)[0] : ''); };
  for (const panel of d.querySelectorAll('[data-cockpit-panel]')) {
    const pr = vis(panel);
    if (!pr) continue;
    for (const n of panel.querySelectorAll('*')) {
      if (n.closest('[hidden]')) continue;
      const st = cs(n);
      if (st.display === 'none' || st.display === 'contents' || st.visibility === 'hidden' || /^(SCRIPT|STYLE|OPTION|TEMPLATE)$/.test(n.tagName)) continue;
      const r = n.getBoundingClientRect();
      if (r.width < 1 && r.height > 4 && (n.children.length || (n.textContent || '').trim())) zero.push(label(n));
    }
  }
  for (const n of d.querySelectorAll('.cv-sheet label, .cv-sheet b')) {
    if (vis(n) && n.scrollWidth > n.clientWidth + 1) cut.push(label(n) + ' "' + (n.textContent || '').trim().slice(0, 14) + '" [' + n.scrollWidth + '/' + n.clientWidth + ']');
  }
  for (const t of d.querySelectorAll('.cv-sheet > div')) {
    const r = vis(t);
    if (!r) continue;
    const need = [...t.children].reduce((h, c) => h + c.getBoundingClientRect().height, 0);
    if (r.height > 2.4 * need + 6 && r.height > 60) tiles.push(label(t) + ' ' + Math.round(r.width) + 'x' + Math.round(r.height) + ' holds ' + Math.round(need) + ' px');
  }
  // dead space inside the panels: the largest empty square among what is painted (4 px cells)
  const cell = 4;
  let inner = { size: 0, at: [0, 0], panel: '' };
  for (const panel of d.querySelectorAll('[data-cockpit-panel]')) {
    const pr = vis(panel);
    if (!pr) continue;
    const cols = Math.ceil(pr.width / cell), rows = Math.ceil(pr.height / cell), cov = new Uint8Array(cols * rows);
    const mark = (r) => {
      for (let y = Math.max(0, Math.floor((r.top - pr.top) / cell)); y < Math.min(rows, Math.ceil((r.bottom - pr.top) / cell)); y++) {
        for (let x = Math.max(0, Math.floor((r.left - pr.left) / cell)); x < Math.min(cols, Math.ceil((r.right - pr.left) / cell)); x++) cov[y * cols + x] = 1;
      }
    };
    // the panel's own padding is not empty space: the border of the zone
    mark({ left: pr.left, right: pr.right, top: pr.top, bottom: pr.top + 12 }); mark({ left: pr.left, right: pr.right, top: pr.bottom - 12, bottom: pr.bottom });
    mark({ left: pr.left, right: pr.left + 12, top: pr.top, bottom: pr.bottom }); mark({ left: pr.right - 12, right: pr.right, top: pr.top, bottom: pr.bottom });
    for (const n of panel.querySelectorAll('.d-cells, .d-doll, .cv-sheet > div, .cv-log, .cv-rules, .cv-pop, .cv-mextra, button, input, img, li, .cv-life, .cv-st, .cv-setbar > *')) {
      if (n.closest('[hidden]')) continue;
      const r = vis(n);
      if (r) mark(r);
    }
    const walker = d.createTreeWalker(panel, 4); // text nodes: where the words are, not the box they sit in
    for (let t = walker.nextNode(); t; t = walker.nextNode()) {
      if (!(t.nodeValue || '').trim() || (t.parentNode.closest && t.parentNode.closest('[hidden]'))) continue;
      const rg = d.createRange(); rg.selectNodeContents(t);
      for (const r of rg.getClientRects()) if (r.width > 0 && r.height > 0) mark(r);
    }
    let best = 0, bx = 0, by = 0, prev = new Int16Array(cols + 1), cur = new Int16Array(cols + 1);
    for (let y = 0; y < rows; y++) {
      cur.fill(0);
      for (let x = 0; x < cols; x++) {
        if (cov[y * cols + x]) continue;
        const sq = Math.min(prev[x], prev[x + 1], cur[x]) + 1;
        cur[x + 1] = sq;
        if (sq > best) { best = sq; bx = x; by = y; }
      }
      [prev, cur] = [cur, prev];
    }
    if (best * cell > inner.size) inner = { size: best * cell, at: [Math.round(pr.left + (bx - best + 1) * cell), Math.round(pr.top + (by - best + 1) * cell)], panel: panel.dataset.panel || '?' };
  }
  return { zero, cut, tiles, inner };
}

// The elements a page says must be on it, with a size (opts.expect): the machine form of "this state is drawn".
export function expects(win, selectors) {
  const d = win.document, out = {};
  for (const sel of selectors) {
    const n = d.querySelector(sel), r = n && n.getBoundingClientRect();
    out[sel] = !!r && r.width > 0 && r.height > 0;
  }
  return out;
}
