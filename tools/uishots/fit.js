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
  const chx = d.body.hasAttribute('data-chars') ? charsPage(win, d, vis, cs) : null; // the Characters page (docs/charselect-ui.md)
  const vtx = d.body.hasAttribute('data-vault') ? vaultPage(win, d, vis, cs) : null; // the Ethereal Vault page (docs/vault.md)
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
  // the video card (card videoctl): no control box (a button) may touch the caption or any other text of the card; the controls sit in a
  // toolbar under the picture, and the picture's own corner text (caption, badge, rate line) must stay readable
  const vpn = d.querySelector('[data-panel="video"]'), vo = [];
  let vn = 0;
  if (vpn && vis(vpn)) {
    const hasText = (n) => [...n.childNodes].some((c) => c.nodeType === 3 && c.textContent.trim());
    const ctls = [...vpn.querySelectorAll('button, [role="button"]')].filter((n) => vis(n) && cs(n).visibility !== 'hidden');
    const txts = [...vpn.querySelectorAll('*')].filter((n) => hasText(n) && !n.closest('button') && vis(n) && cs(n).visibility !== 'hidden');
    vn = ctls.length;
    for (const c of ctls) {
      const a = c.getBoundingClientRect();
      for (const t of txts) {
        if (t.contains(c) || c.contains(t)) continue;
        const b = t.getBoundingClientRect();
        if (a.left < b.right - 1 && a.right > b.left + 1 && a.top < b.bottom - 1 && a.bottom > b.top + 1) vo.push(name(c) + ' covers ' + name(t));
      }
    }
  }
  const video = { controls: vn, overlap: vo };
  return {
    video, orderCut, zones, vw, vh, scrollH: se.scrollHeight, scrollW: se.scrollWidth, overflow: over, outside: out, cv: cvx, chars: chx, vault: vtx, empty, unready, slots,
    covered: Math.round((covered * 100) / (cols * rows)), hole: best * cs2, holeAt: [(bx - best + 1) * cs2, (by - best + 1) * cs2],
  };
}

// The overlay policy (docs/ui-overlays.md): what is open now, and what is wrong with it.
//  open:  the names of the visible transient overlays (the "..." menu, tile sheets, drawers, the Reset confirm); at most one
//  menu:  menu items whose centre is not the item or inside it (covered by a sheet or clipped), each after scrolling to it
//  menuScroll: the primary row below the fold, a menu taller than its window that cannot scroll or gives no hint (card menufit2)
//  cut:   text in a tile sheet that is wider than its box (a line that is cut, not wrapped)
//  under: a sheet whose top is above the header's bottom, or whose bottom is past the window
//  charter: the "Pay and go" button of the zone sheet when it is not visible inside the sheet
export function overlayState(win) {
  const d = win.document, vis = (n) => { if (!n) return null; const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  const open = [], menu = [], cut = [], under = [], charter = [];
  const band = d.querySelector('.d-band-status'), bandBottom = band ? band.getBoundingClientRect().bottom : 0;
  const pop = d.getElementById('menu-more'), menuScroll = [];
  if (vis(pop)) {
    open.push('menu');
    // card menufit2: every entry must be reachable. The primary row (Pause, Reset) is on show without scrolling; when the groups
    // are taller than the window the body must scroll and say so; then each entry is scrolled to and must be on top there.
    const body = pop.querySelector('.d-menu-body'), pr = pop.getBoundingClientRect();
    for (const n of pop.querySelectorAll('.d-menu-primary .d-btn')) {
      const r = vis(n);
      if (r && (r.bottom > pr.bottom + 0.5 || r.bottom > win.innerHeight)) menuScroll.push('primary action "' + (n.textContent || '?').trim() + '" is below the fold');
    }
    if (body && body.scrollHeight - body.clientHeight > 1) {
      if (!/auto|scroll/.test(win.getComputedStyle(body).overflowY)) menuScroll.push('the menu is ' + (body.scrollHeight - body.clientHeight) + ' px taller than its window and cannot scroll');
      else if (win.getComputedStyle(body, '::after').content === 'none') menuScroll.push('the menu scrolls but shows no hint that there is more');
    }
    for (const n of pop.querySelectorAll('.d-btn, .d-menu-gh')) {
      if (n.scrollIntoView && body && body.contains(n)) n.scrollIntoView({ block: 'center' });
      const r = vis(n);
      if (!r) continue;
      const at = d.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
      if (!at || !(at === n || n.contains(at))) menu.push((n.textContent || n.id || '?').trim().slice(0, 20));
    }
    if (body) body.scrollTop = 0;
  }
  for (const sh of d.querySelectorAll('.d-tilesheet')) {
    const r = vis(sh);
    if (!r) continue;
    for (const it of sh.querySelectorAll('.o-items')) { // the Orders More sheet (card morefit): the list shows min(3, commands) rows in the sheet and never clips without a scroll
      if (!vis(it)) continue;
      const rows = [...it.querySelectorAll('.o-item')], sr = sh.getBoundingClientRect(), lr = it.getBoundingClientRect(), shown = rows.filter((x) => { const q = x.getBoundingClientRect(); return q.height > 0 && q.top >= Math.max(sr.top, lr.top) - 1 && q.bottom <= Math.min(sr.bottom, lr.bottom) + 1 && q.bottom <= win.innerHeight + 1; }).length;
      if (shown < Math.min(3, rows.length)) menuScroll.push('the More list shows ' + shown + ' of ' + rows.length + ' commands in a sheet ' + Math.round(sr.height) + ' px high');
      if (it.scrollHeight - it.clientHeight > 1 && !/auto|scroll/.test(win.getComputedStyle(it).overflowY)) menuScroll.push('the More list is ' + (it.scrollHeight - it.clientHeight) + ' px taller than its box and cannot scroll');
      if (sh.scrollHeight - sh.clientHeight > 1 && !/auto|scroll/.test(win.getComputedStyle(sh).overflowY)) menuScroll.push('the More sheet is ' + (sh.scrollHeight - sh.clientHeight) + ' px taller than its tile and cannot scroll');
    }
    open.push('sheet:' + (sh.dataset.panel || '?'));
    if (r.top < bandBottom - 1 || r.bottom > win.innerHeight + 1) under.push((sh.dataset.panel || '?') + ' top ' + Math.round(r.top) + ' bottom ' + Math.round(r.bottom) + ' (header ends ' + Math.round(bandBottom) + ', window ' + win.innerHeight + ')');
    for (const n of sh.querySelectorAll('*')) {
      if (/^(SELECT|INPUT|OPTION)$/.test(n.tagName) || !vis(n)) continue;
      if (n.scrollWidth - n.clientWidth > 1 && n.clientWidth > 0) cut.push((n.className || n.tagName) + ' "' + (n.textContent || '').trim().slice(0, 24) + '" ' + n.scrollWidth + '/' + n.clientWidth);
    }
    const sb = sh.querySelector('.zm-charter .d-btn'), sr = sb && vis(sb);
    if (sb && !(sr && sr.top >= r.top - 1 && sr.bottom <= r.bottom + 1 && sr.left >= r.left - 1 && sr.right <= r.right + 1)) charter.push('Pay and go is outside the sheet');
  }
  for (const n of d.querySelectorAll('.d-drawer')) {
    if (!vis(n)) continue;
    open.push('drawer:' + (n.dataset.drawer || '?'));
    if (n.dataset.drawer !== 'quest-board') continue;
    for (const x of n.querySelectorAll('*')) { // the quest sheet wraps its text; nothing in it may be wider than its box
      if (/^(SELECT|INPUT|OPTION)$/.test(x.tagName) || !vis(x)) continue;
      if (x.scrollWidth - x.clientWidth > 1 && x.clientWidth > 0) cut.push((x.className || x.tagName) + ' "' + (x.textContent || '').trim().slice(0, 24) + '" ' + x.scrollWidth + '/' + x.clientWidth);
    }
  }
  // the quest panels (strip tile, sheet, phone tab): touch targets on a phone and the words the design forbids (a chapter counter, a number for the margin)
  const small = [], words = [];
  for (const p of d.querySelectorAll('[data-panel^="quest-"]')) {
    if (!vis(p)) continue;
    if (win.innerWidth <= 480) for (const b of p.querySelectorAll('button')) { const r = vis(b); if (r && r.height < 43.5) small.push((b.dataset.key || b.textContent || '?').trim().slice(0, 20) + ' ' + Math.round(r.height) + ' px'); }
    const t = p.textContent || '';
    const m = t.match(/chapter\s*\d+\s*of\s*\d+|margin\D{0,8}\d|\d+(\.\d+)?\s*%/i);
    if (m) words.push(p.dataset.panel + ': "' + m[0] + '"');
  }
  if (vis(d.getElementById('reset-confirm'))) open.push('reset');
  return { open, menu, menuScroll, cut, under, charter, small, words };
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

// The Characters page (docs/charselect-ui.md, brief J-second-attempt.md): what the cockpit checks cannot see on a page that is one header and one
// list that scrolls inside the window. Lists, each naming what a person would see:
//  clipped:  text cut by its own box (an ellipsis or overflow), or sticking out of the card, row, sheet or bar that holds it
//  small:    a control under 44 px high
//  covered:  a control with something else at its centre (the header over the list, a sheet over a button)
//  outside:  a header, list, card or sheet that ends outside the window
//  under:    the list starting under the header
//  sheet:    an open sheet that scrolls or is cut (its buttons may be out of reach)
//  hole:     the largest empty square inside the list, above the end of its content (what follows the last card is the end of a list, not a gap)
function charsPage(win, d, vis, cs) {
  const vw = win.innerWidth, vh = win.innerHeight;
  const o = { clipped: [], small: [], covered: [], outside: [], under: '', sheet: '', hole: 0, holeAt: [0, 0], cards: 0, loading: false, open: false };
  const nm = (n) => n.tagName.toLowerCase() + (typeof n.className === 'string' && n.className ? '.' + n.className.trim().split(/\s+/)[0] : '') + ' "' + (n.textContent || '').trim().slice(0, 16) + '"';
  const root = d.getElementById('cs-root'), main = d.getElementById('cs-main'), head = root && root.querySelector('.cs-head');
  if (!root || !main || !head) { o.clipped.push('the page has no header or list'); return o; }
  const sheet = d.getElementById('cs-sheet'), open = !!sheet && !sheet.hidden && !!vis(sheet);
  o.open = open;
  o.cards = d.querySelectorAll('.cs-card').length;
  o.loading = [...d.querySelectorAll('.cs-state p')].some((p) => /^Loading/.test(p.textContent));
  const live = (n) => vis(n) && !n.closest('[hidden]');
  const box = 'cs-card,.cs-delrow,.cs-state,.cs-sheet-card,.cs-note,.cs-strip,.cs-head';
  for (const n of root.querySelectorAll('*')) {
    if (!live(n) || !([...n.childNodes].some((c) => c.nodeType === 3 && c.nodeValue.trim()))) continue;
    if (n.clientWidth && (n.scrollWidth > n.clientWidth + 1 || (/^(hidden|clip|auto|scroll)$/.test(cs(n).overflowY) && n.scrollHeight > n.clientHeight + 1))) { o.clipped.push(nm(n) + ' [' + n.scrollWidth + '/' + n.clientWidth + ' ' + n.scrollHeight + '/' + n.clientHeight + ']'); continue; }
    const c = n.closest('.' + box);
    if (!c) continue;
    const rg = d.createRange();
    rg.selectNodeContents(n);
    const r = rg.getBoundingClientRect(), cr = c.getBoundingClientRect();
    if (r.width && (r.right > cr.right + 1 || r.left < cr.left - 1)) o.clipped.push(nm(n) + ' sticks out of ' + (typeof c.className === 'string' ? c.className.split(/\s+/)[0] : 'its box'));
  }
  const ctl = [...root.querySelectorAll('button, a.d-btn, input:not([type=radio]), label.cs-chipbtn')].filter(live);
  for (const n of ctl) { const r = n.getBoundingClientRect(); if (r.height < 43.5) o.small.push(nm(n) + ' is ' + Math.round(r.height) + ' px high'); }
  const mr = main.getBoundingClientRect(), hr = head.getBoundingClientRect();
  if (mr.top < hr.bottom - 1) o.under = 'the list starts at ' + Math.round(mr.top) + ' px, under the header that ends at ' + Math.round(hr.bottom);
  for (const n of ctl) {
    const r = n.getBoundingClientRect(), inSheet = !!n.closest('#cs-sheet'), inMain = main.contains(n);
    if (open !== inSheet) continue; // an open sheet: only its own controls; behind the scrim is by design
    const cx = r.left + r.width / 2, cy = r.top + r.height / 2;
    if (cx < 0 || cy < 0 || cx > vw || cy > vh || (inMain && (cy < mr.top || cy > mr.bottom))) continue; // scrolled out of the list: a thumb scrolls to it
    const at = d.elementFromPoint(cx, cy);
    if (!at || !(at === n || n.contains(at) || at.contains(n))) o.covered.push(nm(n));
  }
  for (const n of root.querySelectorAll('.cs-head, .cs-main, .cs-strip, .cs-card, .cs-delrow, .cs-sheet-card')) {
    if (!live(n)) continue;
    const r = n.getBoundingClientRect(), tall = n.matches('.cs-head, .cs-main, .cs-sheet-card');
    if (r.right > vw + 1 || r.left < -1 || (tall && (r.bottom > vh + 1 || r.top < -1))) o.outside.push(nm(n) + ' right ' + Math.round(r.right) + ' bottom ' + Math.round(r.bottom));
  }
  const card = d.querySelector('.cs-sheet-card');
  if (open && card && (card.scrollHeight > card.clientHeight + 1)) o.sheet = 'the sheet scrolls or is cut [' + card.scrollHeight + '/' + card.clientHeight + ']';
  if (!open) {
    const cell = 4, cols = Math.ceil(mr.width / cell);
    let bottom = mr.top;
    for (const n of main.querySelectorAll('.cs-card, .cs-delrow, .cs-state, .cs-sec > *')) { const r = vis(n); if (r && !n.closest('[hidden]')) bottom = Math.max(bottom, r.bottom); }
    bottom = Math.min(bottom, mr.bottom, vh);
    const rows = Math.max(0, Math.ceil((bottom - mr.top) / cell)), cov = new Uint8Array(Math.max(1, cols * rows));
    const mark = (r) => {
      for (let y = Math.max(0, Math.floor((r.top - mr.top) / cell)); y < Math.min(rows, Math.ceil((r.bottom - mr.top) / cell)); y++) {
        for (let x = Math.max(0, Math.floor((r.left - mr.left) / cell)); x < Math.min(cols, Math.ceil((r.right - mr.left) / cell)); x++) cov[y * cols + x] = 1;
      }
    };
    for (const n of main.querySelectorAll('.cs-delrow, .cs-state, h2, button, a.d-btn, .cs-chip, .cs-mark, .cs-badge')) { const r = vis(n); if (r && !n.closest('[hidden]')) mark(r); }
    for (const n of main.querySelectorAll('.cs-card')) { // a card's own padding is its border, not a gap; what it holds is painted by the parts above and the words below
      const r = vis(n);
      if (!r) continue;
      mark({ left: r.left, right: r.right, top: r.top, bottom: r.top + 12 }); mark({ left: r.left, right: r.right, top: r.bottom - 12, bottom: r.bottom });
      mark({ left: r.left, right: r.left + 12, top: r.top, bottom: r.bottom }); mark({ left: r.right - 12, right: r.right, top: r.top, bottom: r.bottom });
    }
    const walker = d.createTreeWalker(main, 4);
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
    o.hole = best * cell; o.holeAt = [Math.round(mr.left + (bx - best + 1) * cell), Math.round(mr.top + (by - best + 1) * cell)];
  }
  return o;
}

// The Characters page measured again after the harness pressed something (an open sheet is judged as it stands).
export function charsNow(win) {
  const vis = (n) => { const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  return charsPage(win, win.document, vis, (n) => win.getComputedStyle(n));
}

// The Ethereal Vault page (docs/vault.md, UI section): a header, filters, a bar, one list that scrolls inside the window and, wide, a card.
// Lists, each naming what a person would see:
//  clipped:  text cut by its own box (not an ellipsis or a line clamp: those are by design, the tooltip has the full name), or sticking out of its card
//  small:    a control under 44 px high
//  covered:  a control with something else at its centre
//  outside:  a part of the page that ends outside the window
//  under:    the list starting under the header, filters or bar
//  sheet:    an open sheet that scrolls or is cut
//  unnamed:  a control a screen reader would announce with no name
//  holes:    an Upgrade without a wallet that is not disabled, an Upgrade that is enabled while it says the balance is short, a missing Keep button (Keep is real since V5)
function vaultPage(win, d, vis, cs) {
  const vw = win.innerWidth, vh = win.innerHeight;
  const o = { clipped: [], small: [], covered: [], outside: [], under: '', sheet: '', unnamed: [], holes: [], cells: 0, state: '', open: false, meter: false, mode: '' };
  const nm = (n) => n.tagName.toLowerCase() + (typeof n.className === 'string' && n.className ? '.' + n.className.trim().split(/\s+/)[0] : '') + ' "' + (n.textContent || '').trim().slice(0, 16) + '"';
  const root = d.getElementById('vt-root'), main = d.getElementById('vt-main'), head = root && root.querySelector('.vt-head');
  if (!root || !main || !head) { o.clipped.push('the page has no header or list'); return o; }
  const live = (n) => vis(n) && !n.closest('[hidden]');
  const sheet = root.querySelector('.vt-sheet'), open = !!sheet && !sheet.hidden && !!vis(sheet);
  o.open = open;
  o.cells = d.querySelectorAll('.vt-cell').length;
  const sc = d.querySelector('.vt-card[data-kind]');
  o.state = sc ? sc.dataset.kind : '';
  o.mode = (d.querySelector('.vt-grid') || { dataset: {} }).dataset.mode || '';
  o.meter = !!d.querySelector('.vt-meter[role="meter"][aria-valuenow]');
  const box = 'vt-card,.vt-sheet-card,.vt-banner,.vt-head,.vt-dcard,.vt-cell,.vt-bar';
  for (const n of root.querySelectorAll('*')) {
    if (!live(n) || n.closest('.d-sr-only') || !([...n.childNodes].some((c) => c.nodeType === 3 && c.nodeValue.trim()))) continue; // screen-reader-only text is clipped to a pixel on purpose
    const c0 = cs(n);
    const by = c0.textOverflow === 'ellipsis' || (c0.webkitLineClamp && c0.webkitLineClamp !== 'none'); // cut with an ellipsis or a clamp: by design
    if (n.clientWidth && !by && (n.scrollWidth > n.clientWidth + 1 || (/^(hidden|clip|auto|scroll)$/.test(c0.overflowY) && n.scrollHeight > n.clientHeight + 1))) { o.clipped.push(nm(n) + ' [' + n.scrollWidth + '/' + n.clientWidth + ' ' + n.scrollHeight + '/' + n.clientHeight + ']'); continue; }
    const c = n.closest('.' + box);
    if (!c || by) continue;
    const rg = d.createRange();
    rg.selectNodeContents(n);
    const r = rg.getBoundingClientRect(), cr = c.getBoundingClientRect();
    if (r.width && (r.right > cr.right + 1 || r.left < cr.left - 1)) o.clipped.push(nm(n) + ' sticks out of ' + (typeof c.className === 'string' ? c.className.split(/\s+/)[0] : 'its box'));
  }
  const ctl = [...root.querySelectorAll('button, a.d-btn, input, select')].filter(live);
  for (const n of ctl) { const r = n.getBoundingClientRect(); if (r.height < 43.5) o.small.push(nm(n) + ' is ' + Math.round(r.height) + ' px high'); }
  const above = [head, root.querySelector('.vt-banner'), root.querySelector('.vt-filters'), root.querySelector('.vt-bar'), root.querySelector('.vt-keeprow'), root.querySelector('.vt-actrow'), root.querySelector('.vt-promptrow'), root.querySelector('.vt-msg')].filter((n) => n && live(n));
  const mr = main.getBoundingClientRect(), edge = Math.max(...above.map((n) => n.getBoundingClientRect().bottom));
  if (mr.top < edge - 1) o.under = 'the list starts at ' + Math.round(mr.top) + ' px, under the parts above it that end at ' + Math.round(edge);
  for (const n of ctl) {
    const r = n.getBoundingClientRect(), inSheet = !!n.closest('.vt-sheet'), inMain = main.contains(n);
    if (open !== inSheet) continue; // an open sheet: only its own controls; behind the scrim is by design
    const cx = r.left + r.width / 2, cy = r.top + r.height / 2;
    if (cx <= 2 || cy <= 2 || cx >= vw - 2 || cy >= vh - 2 || (inMain && (cy <= mr.top + 2 || cy >= mr.bottom - 2))) continue; // scrolled out of the list (a centre within 2 px of an edge is out, the browser returns no element there): a thumb scrolls to it
    const fp = n.closest('.vt-fpanel');
    if (fp) { const pr = fp.getBoundingClientRect(); if (cy < pr.top || cy > pr.bottom) continue; } // inside the folded panel's own scroll
    const at = d.elementFromPoint(cx, cy);
    if (!at || !(at === n || n.contains(at) || at.contains(n) || (n.labels && [...n.labels].some((l) => l.contains(at))))) o.covered.push(nm(n) + (at ? ' under ' + nm(at) : ' at ' + Math.round(cx) + ',' + Math.round(cy) + ' of ' + vw + 'x' + vh));
  }
  for (const n of root.querySelectorAll('.vt-head, .vt-filters, .vt-bar, .vt-body, .vt-main, .vt-grid, .vt-detail, .vt-sheet-card, .vt-banner')) {
    if (!live(n)) continue;
    const r = n.getBoundingClientRect(), tall = n.matches('.vt-head, .vt-body, .vt-main, .vt-sheet-card');
    if (r.right > vw + 1 || r.left < -1 || (tall && (r.bottom > vh + 1 || r.top < -1))) o.outside.push(nm(n) + ' right ' + Math.round(r.right) + ' bottom ' + Math.round(r.bottom));
  }
  const card = d.querySelector('.vt-sheet-card');
  if (open && card && card.scrollHeight > card.clientHeight + 1) o.sheet = 'the sheet scrolls or is cut [' + card.scrollHeight + '/' + card.clientHeight + ']';
  const text = (n) => (n.textContent || '').trim();
  for (const n of root.querySelectorAll('button, a.d-btn, input, select, [role="meter"]')) {
    if (!live(n) && !n.matches('[role="meter"]')) continue;
    const lb = n.getAttribute('aria-labelledby'), byId = lb ? lb.split(/\s+/).map((i) => d.getElementById(i)).filter(Boolean).map(text).join('') : '';
    const lab = n.labels && n.labels.length ? text(n.labels[0]) : '';
    const name = (n.getAttribute('aria-label') || byId || (/^(button|a)$/i.test(n.tagName) ? text(n) : '') || lab || n.getAttribute('title') || '').trim();
    if (!name) o.unnamed.push(nm(n));
  }
  // Keep is real since V5 and only has to exist; Upgrade is a placeholder until the wallet answers (no visible wallet chip), and with a
  // wallet it may be enabled only when nothing says the balance is short
  const chip = root.querySelector('[data-key="wallet"]'), walletOn = !!chip && live(chip);
  if (!root.querySelector('[data-key="keep"]')) o.holes.push('keep button is missing');
  {
    const b = root.querySelector('[data-key="upgrade"]');
    if (!b) o.holes.push('upgrade button is missing');
    const off = !b || b.getAttribute('aria-disabled') === 'true';
    if (b && walletOn) {
      const why = root.querySelector('.vt-upgrade-why');
      if (!off && why && why.textContent.trim()) o.holes.push('upgrade is enabled while it says "' + why.textContent.trim().slice(0, 30) + '"');
    } else if (b && !off) o.holes.push('upgrade is enabled');
  }
  return o;
}

// The Vault page measured again after the harness pressed something (an open sheet is judged as it stands).
export function vaultNow(win) {
  const vis = (n) => { const r = n.getBoundingClientRect(); return r.width > 0 && r.height > 0 ? r : null; };
  return vaultPage(win, win.document, vis, (n) => win.getComputedStyle(n));
}
