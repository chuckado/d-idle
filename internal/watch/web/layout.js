// Layout manifest logic (docs/ui-design-system.md section 6.4), pure: no DOM, so the node tests run it as is. The
// shell (shell.js) applies a plan to the page; the editor (layoutedit.js) moves entries with the same packer as the
// mockup editor (docs/mockups/ui/editor.html).

export const WIDTHS = [['phone', 0], ['tablet', 640], ['laptop', 1024], ['wide', 1440]]; // the breakpoints of section 4.2

export function breakpointFor(width, mode) {
  if (mode === 'overlay') return 'overlay';
  let name = 'phone';
  for (const [n, min] of WIDTHS) if (width >= min) name = n;
  return name;
}

// The breakpoint entry for a name; a missing one falls back to the next wider (never to overlay), then to null.
export function pickBreakpoint(layout, name) {
  if (!layout || layout.schema !== 'watch.layout/1' || !layout.breakpoints) return null;
  const order = WIDTHS.map((w) => w[0]);
  const from = order.indexOf(name);
  const names = from < 0 ? [name] : order.slice(from);
  for (const n of names) {
    const bp = layout.breakpoints[n];
    if (bp && Array.isArray(bp.panels) && bp.panels.length) return { name: n, cols: bp.cols || layout.cols || 12, panels: bp.panels };
  }
  return null;
}

// ---- the packer (the editor's): first free cell top-down, and push down on a collision ----
export function occ(panels, skip) {
  const m = {};
  for (const p of panels) {
    if (p === skip || p.hidden) continue; // a hidden panel reserves nothing
    const h = p.collapsed ? 1 : p.rowSpan;
    for (let r = p.row; r < p.row + h; r++) for (let c = p.col; c < p.col + p.colSpan; c++) m[r + ':' + c] = p.id;
  }
  return m;
}
export function free(m, c, r, w, h) {
  for (let y = r; y < r + h; y++) for (let x = c; x < c + w; x++) if (m[y + ':' + x]) return false;
  return true;
}
export function placeNear(panels, p) {
  const m = occ(panels, p), h = p.collapsed ? 1 : p.rowSpan;
  for (let r = p.row; r < p.row + 400; r++) if (free(m, p.col, r, p.colSpan, h)) { p.row = r; return; }
}
function placeFrom(panels, p, cols, fromRow) {
  const m = occ(panels, p);
  for (let r = fromRow; r < fromRow + 400; r++) {
    for (let c = 1; c + p.colSpan - 1 <= cols; c++) if (free(m, c, r, p.colSpan, p.rowSpan)) { p.col = c; p.row = r; return; }
  }
}

const SPAN = { s: [3, 5], m: [6, 6], l: [12, 6] }; // columns and rows of a panel the layout does not mention

// plan(layout, width, mode, conf, targets): where every target goes. targets are the placeable elements of the page
// in reading order ({id, size}); conf is the layout block of /watch/panels.json (aliases, compat, reserved). Returns
// null when no breakpoint applies (the shell then places by manifest), else
// { name, cols, rowHeight, flow, items: [{id, col, row, colSpan, rowSpan, collapsed, order, placed}], unknown: [ids] }.
// flow (phone and tablet): rows are not fixed, the shell uses order and the column span and lets the rows flow.
// A hidden target (a locked feature, a route that is off) takes no cell and the panels under it rise into the room, so the
// page has no dead bands. With edit set (the editor) the cells are the same as on the normal page, and the hidden
// panels the file names come along as { hidden: true } at their file position: they are saved again but reserve no
// cell (they once did, invisibly, which is why a panel could not be dropped into room that looked free).
export function plan(layout, width, mode, conf, targets, edit) {
  const bpName = breakpointFor(width, mode);
  const bp = pickBreakpoint(layout, bpName);
  if (!bp) return null;
  const aliases = (conf && conf.aliases) || {}, reserved = (conf && conf.reserved) || [];
  const known = {};
  targets.forEach((t) => { known[t.id] = t; });
  const items = [], used = {}, unknown = [], kept = [], off = [];
  let maxRow = 0, maxOrder = -1;
  // An entry whose id is not on the page (a new tile that is not built yet) puts its `fallback` in the same cell, and
  // shows it even when the layout also lists that target as hidden. An id that is on the page ignores the fallback.
  const res = (id) => (known[id] ? id : known[aliases[id]] ? aliases[id] : '');
  const forced = new Set(), entries = bp.panels.map((e) => {
    const fb = !res(e.id) && e.fallback ? res(e.fallback) : '';
    if (!fb) return e;
    forced.add(fb);
    return Object.assign({}, e, { id: fb, via: e.id }); // the entry's own hidden stays: a stand-in the user switched off stays off
  });
  for (const e of entries) {
    if (!e.via && forced.has(res(e.id))) continue; // the entry that was hidden stays hidden while it stands in
    const id = res(e.id);
    if (!id) { if (!reserved.includes(e.id) && !e.fallback) unknown.push(e.id); continue; }
    if (used[id]) continue; // the first entry wins when an alias and its target are both named
    used[id] = true;
    const colSpan = Math.max(1, Math.min(bp.cols, e.colSpan | 0)), col = Math.max(1, Math.min(bp.cols - colSpan + 1, e.col | 0));
    const it = { id, col, row: Math.max(1, e.row | 0), colSpan, rowSpan: Math.max(1, e.rowSpan | 0), collapsed: !!e.collapsed, order: e.order | 0, placed: true };
    if (e.via) it.via = e.via; // saved again as the entry it stands in for (toPanels)
    // an entry with hidden: true is a panel the user switched off in the editor: no cell, and the shell hides it
    if (known[id].rows) it.rowSpan = known[id].rows; // fit content: the stored span is not used (below)
    if (e.hidden) { off.push(id); if (edit) kept.push(Object.assign(it, { hidden: true, off: true })); continue; }
    if (known[id].hidden) { if (edit) kept.push(Object.assign(it, { hidden: true })); continue; }
    placeNear(items, it); // a file with overlapping entries (the editor's presets have some) still loads: push down
    items.push(it);
    maxRow = Math.max(maxRow, it.row + (it.collapsed ? 1 : it.rowSpan) - 1);
    maxOrder = Math.max(maxOrder, it.order);
  }
  // panels the layout does not mention are packed after the placed ones, in page order
  const flow = bpName === 'phone' || bpName === 'tablet';
  for (const t of targets) {
    if (used[t.id] || t.hidden) continue;
    const [w, h0] = SPAN[t.size] || SPAN.s, h = t.rows || h0;
    const it = { id: t.id, col: 1, row: maxRow + 1, colSpan: flow && bpName === 'phone' ? bp.cols : Math.min(bp.cols, flow ? Math.max(w, 6) : w), rowSpan: h, collapsed: false, order: ++maxOrder, placed: false };
    placeFrom(items, it, bp.cols, maxRow + 1);
    items.push(it);
  }
  if (!flow) { // each panel rises in its own columns into the room a hidden one left
    const done = [];
    for (const it of items.slice().sort((x, y) => x.row - y.row || x.col - y.col)) {
      const m = occ(done), h = it.collapsed ? 1 : it.rowSpan;
      while (it.row > 1 && free(m, it.col, it.row - 1, it.colSpan, h)) it.row--;
      done.push(it);
    }
  }
  items.push(...kept);
  return { name: bp.name, bp: bpName, cols: bp.cols, rowHeight: bpName === 'overlay' ? 40 : (layout.rowHeight || 48), flow, items, unknown, off };
}

// The entries of a plan as a breakpoint's panels array (what the editor saves), order from row then column.
export function toPanels(items, bands) {
  return items.slice().sort((a, b) => a.row - b.row || a.col - b.col).map((it, i) => {
    const o = { id: it.id, band: (bands && bands[it.id]) || 'grid', col: it.col, row: it.row, colSpan: it.colSpan, rowSpan: it.rowSpan, collapsed: !!it.collapsed, order: i };
    if (it.off) o.hidden = true; // only a switched-off panel carries the key, so other entries are as the mockup editor writes them
    if (it.via) { o.id = it.via; o.fallback = it.id; } // a stand-in is saved as the new tile it stands in for
    return o;
  });
}

// withShadows(saved, original): the layout the editor saved, with back the entries a stand-in displaced. A fallback
// target that the file lists as hidden is not an item of the plan while it stands in (the stand-in has its cell), so
// the editor's own export lacks that entry; without it the target would be packed after the placed panels in the
// stored copy. Only breakpoints the editor wrote and entries that name a fallback are looked at; the input is not changed.
export function withShadows(saved, original) {
  const out = JSON.parse(JSON.stringify(saved));
  for (const [name, bp] of Object.entries(out.breakpoints || {})) {
    const was = ((original.breakpoints || {})[name] || {}).panels || [];
    let order = bp.panels.reduce((m, p) => Math.max(m, p.order | 0), -1);
    for (const e of bp.panels.slice()) {
      if (!e.fallback || bp.panels.some((p) => p.id === e.fallback)) continue;
      const old = was.find((p) => p.id === e.fallback);
      if (old) bp.panels.push(Object.assign({}, old, { order: ++order }));
    }
  }
  return out;
}

// ---- which panels the shell mounts ----
export const PRESETS = ['classic', 'simple', 'standard', 'stream', 'debug'];

// The ids a layout names on any breakpoint (placed, not hidden) and in its status list. A cockpit's slot ids are not here: the
// cockpit mounts its own (selectPanels), and the OBS strip under a cockpit mounts what it mounts under any layout.
function named(layout) {
  const ids = new Set();
  for (const bp of Object.values((layout && layout.breakpoints) || {})) {
    for (const e of (bp && bp.panels) || []) {
      if (!e.hidden) { ids.add(e.id); if (e.fallback) ids.add(e.fallback); }
    }
  }
  for (const e of (layout && layout.status) || []) { ids.add(typeof e === 'string' ? e : e.id); if (e.fallback) ids.add(e.fallback); }
  return ids;
}

// statusIds(layout, list): the status band's panel ids the layout chooses, or null when it has no `status` key (the page's
// own list then applies). list is the manifests ({id, ...}). An entry that is not a panel uses its fallback, else is skipped.
export function statusIds(layout, list) {
  if (!layout || !Array.isArray(layout.status)) return null;
  const have = new Set(list.map((m) => m.id)), out = [];
  for (const e of layout.status) {
    const id = typeof e === 'string' ? e : e && e.id, fb = e && e.fallback;
    const pick = have.has(id) ? id : fb && have.has(fb) ? fb : '';
    if (pick && !out.includes(pick)) out.push(pick);
  }
  return out;
}

// selectPanels(list, layout): the manifests the shell mounts. An optIn panel only when the layout in force names it;
// with a `status` key (not in overlay mode), only the status-band panels it chooses. Without a layout nothing optIn is mounted.
export function selectPanels(list, layout, mode) {
  if (isCockpit(layout) && mode !== 'overlay') { const c = cockpitIds(layout); return list.filter((m) => m.slot && c.has(m.id)); } // no old tile in a cockpit
  const names = named(layout), status = mode === 'overlay' ? null : statusIds(layout, list); // the OBS strip keeps the page's own band
  return list.filter((m) => (!m.optIn || names.has(m.id)) && (m.band !== 'status' || status === null || status.includes(m.id)));
}

// ---- the editor's move: push down, cascade, close up (a dashboard grid) ----
const rows = (p) => (p.collapsed ? 1 : p.rowSpan);
const hits = (a, b) => a !== b && a.col < b.col + b.colSpan && b.col < a.col + a.colSpan && a.row < b.row + rows(b) && b.row < a.row + rows(a);

// moveItem(items, id, rect, cols): the layout after panel id takes rect ({col, row, colSpan, rowSpan, collapsed}; any
// part may be left out). It never changes items: the result is a copy, so a preview that is dropped costs nothing and
// the layout before the drag is still there to go back to. The moved panel lands exactly on its cells; every panel it
// covers is pushed below it, the ones those cover are pushed in turn; then each panel rises in its columns as far as
// there is room (the moved one first where two start level), which is what the normal page does too. Hidden panels
// neither block nor move. A move onto free cells changes nothing else.
export function moveItem(items, id, rect, cols) {
  const out = items.map((i) => Object.assign({}, i));
  const it = out.find((i) => i.id === id && !i.hidden);
  if (!it) return out;
  Object.assign(it, rect);
  it.colSpan = Math.max(1, Math.min(cols, it.colSpan | 0));
  it.col = Math.max(1, Math.min(cols - it.colSpan + 1, it.col | 0));
  it.row = Math.max(1, it.row | 0);
  it.rowSpan = Math.max(1, it.rowSpan | 0);
  const vis = out.filter((i) => !i.hidden);
  const queue = [it];
  for (let guard = 0; queue.length && guard < 20000; guard++) {
    const a = queue.shift();
    for (const b of vis.slice().sort((x, y) => x.row - y.row || x.col - y.col)) {
      if (b !== it && hits(a, b)) { b.row = a.row + rows(a); queue.push(b); }
    }
  }
  rise(vis, it);
  return out;
}

function rise(vis, first) {
  const done = [];
  for (const x of vis.slice().sort((p, q) => p.row - q.row || (p === first ? -1 : q === first ? 1 : 0) || p.col - q.col)) {
    const m = occ(done);
    while (x.row > 1 && free(m, x.col, x.row - 1, x.colSpan, rows(x))) x.row--;
    done.push(x);
  }
}

// setOff(items, id, on, cols): the layout after the user hides (on) or shows a panel. A hidden panel keeps its cells in
// the file and reserves nothing; the others rise into its room. Shown again it takes its cells back and pushes down.
export function setOff(items, id, on, cols) {
  const out = items.map((i) => Object.assign({}, i));
  const it = out.find((i) => i.id === id);
  if (!it || (it.hidden && !it.off)) return out; // a panel the page itself hides is not the user's to switch
  if (on) { it.off = true; it.hidden = true; rise(out.filter((i) => !i.hidden)); return out; }
  delete it.off; delete it.hidden;
  return moveItem(out, id, {}, cols);
}

// Whether two layouts put every panel on the same cells (the editor: is there anything to commit).
export function sameCells(a, b) {
  const key = (i) => [i.id, i.col, i.row, i.colSpan, i.rowSpan, !!i.collapsed, !!i.off].join();
  return a.length === b.length && a.map(key).sort().join('|') === b.map(key).sort().join('|');
}

// ---- fit content: a panel's rows come from what it shows, not from the file ----
// A panel that declares fit (panel.json "fit": true; the pre-shell cards through shell.json) takes the rows its
// rendered content needs at the width the layout gives it. The stored rowSpan stays in the file (old files and editor
// exports load unchanged, and a loader without this reads it as before) but is not used for such a panel: a span
// taller than the content left a dead area in the card (the game picture is 4:3 of its width, whatever was saved).
export function fitRows(heightPx, rowh, gap) {
  return Math.max(1, Math.ceil((heightPx - 0.5 + gap) / (rowh + gap))); // half a pixel of rounding never costs a row
}

// The height a card needs: from its top edge to the bottom of its lowest visible child, plus its own bottom padding
// and border. cs is the card's computed style. 0 when nothing is visible in it.
export function contentHeight(n, cs) {
  const top = n.getBoundingClientRect().top;
  let bottom = top, any = false;
  for (const c of n.children) {
    if (c.hidden || !c.getClientRects || !c.getClientRects().length) continue;
    bottom = Math.max(bottom, c.getBoundingClientRect().bottom);
    any = true;
  }
  if (!any) return 0;
  return bottom - top + (parseFloat(cs.paddingBottom) || 0) + (parseFloat(cs.borderBottomWidth) || 0);
}

// ---- the cockpit template (cockpit/1; docs/ui/redesign-2026-11-14/J-second-attempt.md section 3) ----
// A cockpit layout names panels per slot instead of cells of a grid. Everything here is pure; the shell (shell.js)
// puts the panels into the slot elements and shell.css does the geometry.
export const COCKPIT = 'cockpit/1';
export const COCKPIT_WIDE = 1700;  // from this width the rail is two columns (railWide)
export const COCKPIT_PHONE = 700;  // under this width: band, picture, tabs
export const COCKPIT_SHORT = 760;  // under this height (not on a phone) the strip is at most 92 px (the renders)
export const STRIP_SHORT = 92;
export const COCKPIT_LIMITS = { strip: [48, 240], railMin: [240, 800] };
const COCKPIT_DEFAULTS = { strip: 112, railMin: 400 };
const idList = (v) => (Array.isArray(v) ? v : typeof v === 'string' ? [v] : []).filter((s) => typeof s === 'string' && s);

export function isCockpit(layout) {
  return !!layout && layout.schema === 'watch.layout/1' && layout.template === COCKPIT && !!layout.slots && idList(layout.slots.stage).length === 1;
}

// The pools of a cockpit's slots: [name, array] pairs (the arrays are the layout's own, so a copy can edit them).
function pools(slots) {
  const out = [['band', slots.band], ['stage', slots.stage], ['strip', slots.strip], ['rail', slots.rail]];
  if (slots.railWide) out.push(['railWide.top', slots.railWide.top], ['railWide.left', slots.railWide.left], ['railWide.right', slots.railWide.right]);
  for (const n of Object.keys(slots.phoneTabs || {})) out.push(['phoneTabs.' + n, slots.phoneTabs[n]]);
  return out;
}

// Every panel id the cockpit names, in any arrangement.
export function cockpitIds(layout) {
  const out = new Set();
  if (!isCockpit(layout)) return out;
  for (const [, v] of pools(layout.slots)) idList(v).forEach((id) => out.add(id));
  return out;
}

export function cockpitMode(width, mode) {
  if (mode === 'overlay') return '';
  return width < COCKPIT_PHONE ? 'phone' : width >= COCKPIT_WIDE ? 'wide' : 'page';
}

export function cockpitSizes(layout) {
  const z = (layout && layout.sizes) || {}, out = {};
  for (const k of Object.keys(COCKPIT_DEFAULTS)) {
    const [lo, hi] = COCKPIT_LIMITS[k], v = z[k] | 0;
    out[k] = v ? Math.max(lo, Math.min(hi, v)) : COCKPIT_DEFAULTS[k];
  }
  return out;
}

// cockpitPlan(layout, width, mode, mounted, height): where each mounted panel goes. mounted(id) says a slot panel of that id
// is on the page; a name that is not mounted (not built yet, not a slot panel) is left out and its neighbours take the
// room (no stand-in tile). -> null when the layout is no cockpit or the mode is overlay, else
// { mode, band, stage, strip, top, rail: [[ids]], tabs: [{name, ids}], sizes, has, left }
// rail is one column, or on a wide page the two columns of railWide (left 1.7 parts, right 1) under top, which spans both
// (the accepted renders: Needs you and Orders across, the two maps side by side under them). has is the letters of the
// slots that hold something ('s' stage, 't' strip, 'r' rail, 'p' phone tabs) for the CSS; left are the ids the layout
// names that are not mounted. height (the window's, optional) shortens the strip on a short window.
export function cockpitPlan(layout, width, mode, mounted, height) {
  const m = cockpitMode(width, mode);
  if (!m || !isCockpit(layout)) return null;
  const sl = layout.slots, left = [], keep = (v) => idList(v).filter((id) => (mounted(id) ? true : (left.push(id), false)));
  const band = keep(sl.band), stage = keep(sl.stage)[0] || '';
  let strip = [], rail = [], top = [], tabs = [];
  if (m === 'phone') {
    tabs = Object.keys(sl.phoneTabs || {}).map((name) => ({ name, ids: keep(sl.phoneTabs[name]) })).filter((t) => t.ids.length);
  } else {
    strip = keep(sl.strip);
    if (m === 'wide' && sl.railWide) {
      top = keep(sl.railWide.top);
      rail = [keep(sl.railWide.left), keep(sl.railWide.right)].filter((c) => c.length);
    } else rail = [keep(sl.rail)].filter((c) => c.length);
  }
  const sizes = cockpitSizes(layout);
  if (m !== 'phone' && height > 0 && height < COCKPIT_SHORT) sizes.strip = Math.min(sizes.strip, STRIP_SHORT);
  const has = (stage ? 's' : '') + (strip.length ? 't' : '') + (top.length || rail.length ? 'r' : '') + (tabs.length ? 'p' : '');
  return { mode: m, band, stage, strip, top, rail, tabs, sizes, has, left };
}

// cockpitSwap(layout, a, b): a copy of the layout with the panels a and b exchanged wherever they stand (every
// arrangement: page, wide, phone tabs). When b is not in the layout, a is replaced by b. Same id twice: unchanged.
export function cockpitSwap(layout, a, b) {
  const out = JSON.parse(JSON.stringify(layout));
  if (!isCockpit(out) || a === b) return out;
  const sw = (id) => (id === a ? b : id === b ? a : id);
  const s = out.slots;
  for (const k of ['band', 'stage', 'strip', 'rail']) if (s[k] !== undefined) s[k] = Array.isArray(s[k]) ? s[k].map(sw) : sw(s[k]);
  if (s.railWide) for (const k of ['top', 'left', 'right']) if (s.railWide[k] !== undefined) s.railWide[k] = idList(s.railWide[k]).map(sw);
  for (const n of Object.keys(s.phoneTabs || {})) s.phoneTabs[n] = idList(s.phoneTabs[n]).map(sw);
  return out;
}

// cockpitSize(layout, {strip, railMin}): a copy with the sizes the editor's dividers set, inside the limits.
export function cockpitSize(layout, want) {
  const out = JSON.parse(JSON.stringify(layout));
  out.sizes = Object.assign({}, cockpitSizes(layout), out.sizes || {});
  for (const k of Object.keys(COCKPIT_LIMITS)) {
    if (want[k] === undefined) continue;
    const [lo, hi] = COCKPIT_LIMITS[k];
    out.sizes[k] = Math.max(lo, Math.min(hi, Math.round(want[k])));
  }
  return out;
}
