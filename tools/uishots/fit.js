// The fit check's measurements (tools/uishots fit.go): run from the harness page on a page of the same origin (an iframe of
// the exact viewport size) and returns plain numbers; fit.go judges them. No page code is touched.
// opts.scroll: the ids of the panels whose manifest says scroll: true (the only place a data-scroll list is accepted).
export function measure(win, opts) {
  const d = win.document, se = d.scrollingElement, vw = win.innerWidth, vh = win.innerHeight, decl = new Set((opts && opts.scroll) || []);
  const vis = (n) => { const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  const cs = (n) => win.getComputedStyle(n);
  // a list that scrolls on purpose: marked data-scroll, inside a panel that declares scroll, and really a scroll container
  // data-vscroll: a region of the page that scrolls on purpose (the phone's storage canvas, the item sheet); it only counts while the browser really scrolls it
  const scroller = (n) => (n.hasAttribute('data-scroll') && /^(auto|scroll)$/.test(cs(n).overflowY) && decl.has((n.closest('[data-panel]') || { dataset: {} }).dataset.panel)) || (n.hasAttribute('data-vscroll') && /^(auto|scroll)$/.test(cs(n).overflowY));
  // data-hscroll: a row that scrolls sideways on purpose (the phone's chip row)
  const hscroller = (n) => n.hasAttribute('data-hscroll') && /^(auto|scroll)$/.test(cs(n).overflowX);
  // a few words of the element's own text name it in the verdict (a cell, a slot, a chip), so the fix is found without a debugger
  const name = (n) => { const p = n.closest('[data-panel]'); return (p ? p.dataset.panel : '?') + ' ' + n.tagName.toLowerCase() + (n.className && typeof n.className === 'string' ? '.' + n.className.trim().split(/\s+/)[0] : '') + (n.children.length < 3 && n.textContent ? ' "' + n.textContent.trim().slice(0, 16) + '"' : '') + ' [' + n.scrollHeight + '/' + n.clientHeight + ' ' + n.scrollWidth + '/' + n.clientWidth + ']'; };
  const over = [];
  const test = (n, label) => {
    if (!n.clientHeight && !n.clientWidth) return; // inline boxes and SVG parts have no client box
    const dh = n.scrollHeight - n.clientHeight, dw = n.scrollWidth - n.clientWidth;
    const ell = cs(n).textOverflow === 'ellipsis'; // one line cut with an ellipsis: width only, by design
    const clamp = cs(n).webkitLineClamp && cs(n).webkitLineClamp !== 'none'; // a text cut to N lines with an ellipsis: by design
    const bad = (dh > 1 && !scroller(n) && !clamp) || (dw > 1 && !ell && !scroller(n) && !hscroller(n));
    if (bad) over.push({ el: label, dh: dh > 1 ? dh : 0, dw: dw > 1 && !ell ? dw : 0 });
  };
  for (const n of d.querySelectorAll('[data-cockpit-slot]')) if (vis(n)) test(n, 'slot ' + n.dataset.cockpitSlot);
  for (const root of d.querySelectorAll('[data-cockpit-panel], [data-fit-deep]')) {
    if (!vis(root) || root.closest('.d-drawer')) continue; // a drawer is a side sheet that scrolls inside itself, not part of the frame
    test(root, root.dataset.fitDeep !== undefined ? 'band' : 'panel ' + (root.dataset.panel || '?'));
    // every descendant too: content cut one level below the root is still cut. A scroll container's own children are its business.
    for (const n of root.querySelectorAll('*')) {
      const sc = n.closest('[data-scroll],[data-vscroll],[data-hscroll]');
      if (!vis(n) || sc && sc !== n && root.contains(sc) && (scroller(sc) || hscroller(sc))) continue;
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
    const r = n.closest('.d-drawer') || (n.closest('[data-vscroll]') && scroller(n.closest('[data-vscroll]'))) ? null : vis(n);
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
  for (const n of d.querySelectorAll('[data-cockpit-panel], .d-tabbar, .update-banner')) { const r = vis(n); if (r) mark(r); } // the update banner takes its strip off the frame (--d-banner-h): covered, not dead
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
  // the map switch (Act / Area) sits on the top edge of a map tile and covers no zone button; an order title is never cut while its reason still shows
  const sw = d.querySelector('.d-mapswitch'), swr = sw && vis(sw);
  zones.switchOverlap = swr ? zb.filter((x) => x[1].left < swr.right - 1 && x[1].right > swr.left + 1 && x[1].top < swr.bottom - 1 && x[1].bottom > swr.top + 1).length : 0;
  const orderCut = [...d.querySelectorAll('[data-panel="orders"] .o-row:not(.o-plan)')].filter((r) => {
    const t = r.querySelector('.o-t:not(.o-wide)'), re = r.querySelector('.o-r');
    return t && vis(t) && t.scrollWidth > t.clientWidth + 1 && re && re.clientWidth > 0;
  }).length;
  return {
    orderCut, zones, vw, vh, scrollH: se.scrollHeight, scrollW: se.scrollWidth, overflow: over, outside: out, cv: cvx, empty, unready, slots,
    covered: Math.round((covered * 100) / (cols * rows)), hole: best * cs2, holeAt: [(bx - best + 1) * cs2, (by - best + 1) * cs2],
  };
}

// An opened sheet must be hittable: the element at the centre of each target is the target or inside it.
export function hits(win, selectors) {
  const d = win.document, out = {};
  for (const sel of selectors) {
    const n = d.querySelector(sel);
    if (n && n.scrollIntoView && !n.closest('.cv-sel')) n.scrollIntoView({ block: 'start' }); // a target below the fold of a scrolling page is reached by scrolling, as a thumb does
    const r = n && n.getBoundingClientRect();
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
  // what the page had done when it was measured: a hole on a page that was not finished loading says so
  const g = (id) => d.getElementById(id), diag = [d.readyState, 'keys ' + d.querySelectorAll('[data-key]').length, 'sub "' + ((g('merc-sub') || {}).textContent || '') + '"', 'hs ' + ((g('hsheet') || { children: [] }).children.length),
    'mercbox ' + (g('merc-doll') ? g('merc-doll').parentNode.getBoundingClientRect().width | 0 : -1), win.d2InvActions ? 'act' : 'no-act', win.d2InvTip ? 'tip' : 'no-tip', win.d2InvMerc ? 'merc' : 'no-merc', 'errors [' + (win.d2CvErrors || []).join(' | ') + ']'].join(', ');
  return { zero, cut, tiles, inner, diag, phone: phoneChecks(win, d, vis, cs) };
}

// The phone's storage tab and the art cells (card uiphone, the bug report of 2026-11-15, a 412 px screen). Each list names what the user saw:
//  artLabel: a name painted over the picture of a cell (any width)
//  bar:      a toolbar part cut at the bar's edge or covered (the Sell chip); a chip row that is no scroller is cut
//  acts:     the sheet's buttons clipped, covered, under 44 px, or the sheet running off the window
//  lock:     the "Stash from anywhere" line cut with an ellipsis, or missing on the storage tab
//  hover:    something wider than the scrolling canvas (a sideways scroll)
//  grids:    a cell under 32 px, or a grid that leaves its box's width unused
function phoneChecks(win, d, vis, cs) {
  const o = { artLabel: [], bar: [], acts: [], lock: [], hover: [], grids: [], phone: win.innerWidth <= 700 };
  const nm = (n) => n.tagName.toLowerCase() + (n.id ? '#' + n.id : typeof n.className === 'string' && n.className ? '.' + n.className.trim().split(/\s+/)[0] : '') + ' "' + (n.textContent || '').trim().slice(0, 14) + '"';
  for (const n of d.querySelectorAll('.cv .d-cells .item[data-shown]')) {
    if (!vis(n)) continue;
    const own = [...n.childNodes].some((c) => c.nodeType === 3 && c.nodeValue.trim());
    if (own && parseFloat(cs(n).fontSize) > 0.5) o.artLabel.push(nm(n) + ' font ' + cs(n).fontSize);
  }
  if (!o.phone) return o;
  const bar = d.getElementById('cv-bar'), br = bar && vis(bar);
  if (br) {
    const hs = (n) => n.hasAttribute('data-hscroll') && /^(auto|scroll)$/.test(cs(n).overflowX);
    for (const n of bar.querySelectorAll('*')) {
      const r = vis(n);
      if (!r || n.closest('[hidden]')) continue;
      const sc = n.closest('[data-hscroll]');
      if (sc && sc !== n && hs(sc)) continue; // a chip inside a row that scrolls: cut at the edge by design
      if (r.right > br.right + 0.5 || r.left < br.left - 0.5 || r.bottom > br.bottom + 0.5 || r.top < br.top - 0.5) o.bar.push(nm(n) + ' ends outside the bar [' + Math.round(r.left) + '..' + Math.round(r.right) + ' of ' + Math.round(br.left) + '..' + Math.round(br.right) + ']');
      else if (n.matches('button, input')) {
        const at = d.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
        if (!at || !(at === n || n.contains(at))) o.bar.push(nm(n) + ' is covered');
      }
    }
    const ch = d.getElementById('cv-chips');
    if (ch && vis(ch) && ch.scrollWidth > ch.clientWidth + 1 && !hs(ch)) o.bar.push('the chip row is cut and does not scroll [' + ch.scrollWidth + '/' + ch.clientWidth + ']');
  }
  const sel = d.getElementById('cv-sel'), sr = sel && vis(sel), acts = sel && sel.querySelector('.cv-acts');
  if (sr && acts) {
    const ar = acts.getBoundingClientRect(), vh = win.innerHeight;
    if (sr.bottom > vh + 1 || sr.right > win.innerWidth + 1) o.acts.push('the sheet runs off the window [' + Math.round(sr.bottom) + '/' + vh + ']');
    if (acts.scrollHeight > acts.clientHeight + 1) o.acts.push('the button row is cut [' + acts.scrollHeight + '/' + acts.clientHeight + ']');
    for (const b of acts.querySelectorAll('button')) {
      const r = vis(b);
      if (!r) continue;
      if (r.bottom > ar.bottom + 1 || r.bottom > vh + 1 || r.top < sr.top - 1) o.acts.push(nm(b) + ' is clipped');
      else if (r.height < 43.5) o.acts.push(nm(b) + ' is ' + Math.round(r.height) + ' px high (least 44)');
      else if (b.scrollWidth > b.clientWidth + 1 || b.scrollHeight > b.clientHeight + 1) o.acts.push(nm(b) + ' text is cut');
      else { const at = d.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2); if (!at || !(at === b || b.contains(at))) o.acts.push(nm(b) + ' is covered'); }
    }
  }
  const pack = d.getElementById('cv') && d.getElementById('cv').dataset.tab === 'pack';
  if (pack) {
    const lock = d.getElementById('cv-lock'), lr = lock && vis(lock);
    if (!lr) o.lock.push('the lock line is not on the storage tab');
    else if (lock.scrollWidth > lock.clientWidth + 1 || lock.scrollHeight > lock.clientHeight + 1 || (cs(lock).textOverflow === 'ellipsis' && cs(lock).whiteSpace === 'nowrap')) o.lock.push('the lock line is cut [' + lock.scrollWidth + '/' + lock.clientWidth + ' ' + lock.scrollHeight + '/' + lock.clientHeight + ']');
    const cv = d.getElementById('inv'), cr = cv && vis(cv);
    if (cr) {
      if (cv.scrollWidth > cv.clientWidth + 1) o.hover.push('the canvas scrolls sideways [' + cv.scrollWidth + '/' + cv.clientWidth + ']');
      for (const n of cv.querySelectorAll('.cv-storez *')) {
        const r = vis(n);
        if (r && !n.closest('[hidden]') && (r.right > cr.right + 1 || r.left < cr.left - 1)) { o.hover.push(nm(n) + ' reaches past the canvas'); break; }
      }
      const z = d.querySelector('.cv-storez'), zr = z && vis(z);
      if (zr) {
        const pad = parseFloat(cs(z).paddingLeft) + parseFloat(cs(z).paddingRight);
        for (const id of ['pack', 'belt', 'stash']) {
          const g = d.getElementById(id), r = g && vis(g);
          if (!r) { o.grids.push('#' + id + ' is not on the storage tab'); continue; }
          if (r.width < zr.width - pad - 14) o.grids.push('#' + id + ' is ' + Math.round(r.width) + ' px wide in a box of ' + Math.round(zr.width - pad));
          const c = g.querySelector('.cell'), cw = c && c.getBoundingClientRect().width;
          if (c && id !== 'belt' && cw < 32) o.grids.push('#' + id + ' cells are ' + Math.round(cw) + ' px (least 32)');
        }
      }
    }
  }
  return o;
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
