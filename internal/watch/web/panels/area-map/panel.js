// The area map as a cockpit slot panel (docs/ui/redesign-2026-11-14/briefs/S6-areamap-drawers-phone.md): the same
// /api/senses reply and the same pure helpers as the full `map` panel, drawn to FIT the box the slot gives it, whole,
// letterboxed, never cropped. Two controls only: zoom, and a Layers popover (trail window, coverage, heat, events).
// No timeline, replay, export or legend text. The zone name and the explored share are the tile's label (ctx.meta).
// View only. The drawing functions below are the `map` panel's, with their state passed in (the `map` panel stays as it
// was, so Classic and Debug are untouched); folding the two onto one module is a later cleanup.
import { union, nbParam, TRAIL_WINDOWS, trailWindow, trailPlan, trailAlpha, trailSig, activityShare, ACT_TOKEN, ACT_FALLBACK,
  EVENT_MODES, eventShown, HEAT_MODES, heatMax, heatAlpha, coverSummary } from '../map/panel.js';

const TAU = 6.2832;
const HEAT_COL = { kills: 2, time: 3, damage: 4 };

function decode(m) {
  const n = m.w * m.h, out = new Uint8Array(n), r = m.rle || [];
  let o = 0;
  for (let i = 0; i + 1 < r.length; i += 2) for (let c = 0; c < r[i + 1] && o < n; c++) out[o++] = r[i];
  return out;
}

// The explored share of the area in whole percent, or -1 when the reply carries no coverage of this area.
export function exploredPct(cov) {
  if (!cov || !cov.cells) return -1;
  const total = cov.total || cov.cells.length;
  if (!total) return -1;
  let seen = cov.seen;
  if (seen === undefined) { seen = 0; cov.cells.forEach((c) => { if (c[2] === 1) seen++; }); }
  return Math.max(0, Math.min(100, Math.round(100 * seen / total)));
}

// The tile's label: "Blood Moor 62%".
export function label(area, pct) { return (area || '') + (area && pct >= 0 ? ' ' + pct + '%' : ''); }

// The box of what there is to draw, in the area's own frame: the cells that are not empty (an area's grid is mostly margin, and
// scaling the grid would leave the room small in its tile). Cached per cell array.
const boxes = new WeakMap();
export function bounds(map, cells) {
  let b = boxes.get(cells);
  if (b) return b;
  let x0 = map.w, y0 = map.h, x1 = -1, y1 = -1;
  for (let y = 0; y < map.h; y++) for (let x = 0; x < map.w; x++) {
    if (!cells[y * map.w + x]) continue;
    if (x < x0) x0 = x; if (x > x1) x1 = x; if (y < y0) y0 = y; if (y > y1) y1 = y;
  }
  b = x1 < 0 ? { x0: map.x0, y0: map.y0, x1: map.x0 + map.w * map.step, y1: map.y0 + map.h * map.step }
    : { x0: map.x0 + x0 * map.step, y0: map.y0 + y0 * map.step, x1: map.x0 + (x1 + 1) * map.step, y1: map.y0 + (y1 + 1) * map.step };
  boxes.set(cells, b);
  return b;
}

// The world to fit: the current area's and the neighbours' drawn cells, plus the hero and the trail wherever they are, so
// nothing the picture holds can fall outside it.
export function worldBox(map, cells, nbs, pts) {
  const u = Object.assign({}, bounds(map, cells));
  const grow = (x0, y0, x1, y1) => { u.x0 = Math.min(u.x0, x0); u.y0 = Math.min(u.y0, y0); u.x1 = Math.max(u.x1, x1); u.y1 = Math.max(u.y1, y1); };
  nbs.forEach((n) => { const b = bounds(n.map, n.cells); grow(b.x0 + n.dx, b.y0 + n.dy, b.x1 + n.dx, b.y1 + n.dy); });
  pts.forEach((p) => grow(p.x, p.y, p.x + 1, p.y + 1));
  return u;
}

// The view that fits the whole stitched world into w x h (the smaller of the two scales: letterboxed), times 1.5^zoom
// around the hero once zoomed. Pure.
export function fitView(u, w, h, zoom, hero, pan) {
  const uw = Math.max(1, u.x1 - u.x0), uh = Math.max(1, u.y1 - u.y0);
  const k0 = Math.min(w / uw, h / uh) * 0.96, k = k0 * Math.pow(1.5, zoom); // 4 % margin so the border is never touched
  let cx = (u.x0 + u.x1) / 2, cy = (u.y0 + u.y1) / 2;
  if (zoom > 0 && hero) { cx = hero.x; cy = hero.y; }
  return { k, ox: w / 2 - (cx + pan.x) * k, oy: h / 2 - (cy + pan.y) * k };
}

function drawArea(g, css, m, cl, dx, dy, v) {
  for (let y = 0; y < m.h; y++) for (let x = 0; x < m.w; x++) {
    const c = cl[y * m.w + x];
    if (!c) continue;
    g.fillStyle = c === 1 ? css('--d-surface-3', '#2d2720') : css('--d-line-strong', '#4a4036');
    g.fillRect(v.ox + (m.x0 + dx + x * m.step) * v.k, v.oy + (m.y0 + dy + y * m.step) * v.k, Math.ceil(m.step * v.k), Math.ceil(m.step * v.k));
  }
  const at = (p) => [v.ox + (p.x + dx + 0.5) * v.k, v.oy + (p.y + dy + 0.5) * v.k];
  g.fillStyle = css('--d-info', '#7ab0e8');
  (m.exits || []).forEach((e) => { const a = at(e); g.fillRect(a[0] - 4, a[1] - 4, 8, 8); });
  g.fillStyle = css('--d-accent', '#c7a44a');
  (m.waypoints || []).forEach((w) => {
    const b = at(w);
    g.beginPath(); g.moveTo(b[0], b[1] - 6); g.lineTo(b[0] + 5, b[1]); g.lineTo(b[0], b[1] + 6); g.lineTo(b[0] - 5, b[1]); g.closePath(); g.fill();
  });
}

function drawHeat(g, css, s, at, v) {
  const heat = s.heat;
  if (s.heatMode === 'off' || !heat || !heat.cells || heat.area !== (s.map && s.map.areaId)) return;
  const i = HEAT_COL[s.heatMode], max = heatMax(heat.cells, s.heatMode), side = (heat.cell || 7) * v.k;
  g.save();
  g.fillStyle = css('--d-bad', '#ec7d6c');
  heat.cells.forEach((c) => {
    const al = heatAlpha(c[i], max);
    if (!al) return;
    const a = at({ x: c[0], y: c[1] });
    g.globalAlpha = al; g.fillRect(a[0] - 0.5 * v.k, a[1] - 0.5 * v.k, side, side);
  });
  g.restore();
}

function drawCoverage(g, css, s, at, v) {
  const cov = s.cov;
  if (!s.covOn || !cov || !cov.cells) return;
  const half = (cov.cell || 8) / 2 * v.k, side = half * 2;
  g.save();
  g.lineWidth = 1;
  cov.cells.forEach((c) => {
    const a = at({ x: c[0], y: c[1] }), x = a[0] - half, y = a[1] - half;
    if (c[2] === 1) { g.globalAlpha = 0.28; g.fillStyle = css('--d-good', '#6fbf73'); g.fillRect(x, y, side, side); }
    else if (c[2] === 2) {
      g.globalAlpha = 0.9; g.strokeStyle = css('--d-bad', '#ec7d6c'); g.strokeRect(x, y, side, side);
      g.beginPath(); g.moveTo(x, y + side); g.lineTo(x + side, y); g.stroke();
    } else { g.globalAlpha = 0.35; g.strokeStyle = css('--d-text-2', '#a89d8a'); g.strokeRect(x, y, side, side); }
  });
  if (cov.target) {
    const a = at({ x: cov.target[0], y: cov.target[1] });
    g.globalAlpha = 1; g.lineWidth = 2; g.strokeStyle = css('--d-accent', '#c7a44a');
    g.beginPath(); g.arc(a[0], a[1], Math.max(6, half * 0.7), 0, TAU); g.stroke();
  }
  g.restore();
}

function drawEvent(g, css, a, e) {
  const r = (e.big || e.k === 'levelup' || e.k === 'death') ? 6 : 4, x = a[0], y = a[1];
  g.lineWidth = 2; g.lineCap = 'round';
  switch (e.k) {
    case 'kill': g.strokeStyle = css('--d-bad', '#ec7d6c'); g.beginPath(); g.moveTo(x - r, y - r); g.lineTo(x + r, y + r); g.moveTo(x + r, y - r); g.lineTo(x - r, y + r); g.stroke(); break;
    case 'death': g.strokeStyle = css('--d-text', '#eee'); g.fillStyle = css('--d-bad', '#ec7d6c'); g.beginPath(); g.arc(x, y, r, 0, TAU); g.fill(); g.stroke(); break;
    case 'damage': g.fillStyle = css('--d-warn', '#e0a030'); g.beginPath(); g.moveTo(x, y - r); g.lineTo(x + r, y + r); g.lineTo(x - r, y + r); g.closePath(); g.fill(); break;
    case 'potion': g.strokeStyle = css('--d-info', '#7ab0e8'); g.beginPath(); g.arc(x, y + 1, r - 1, 0, TAU); g.moveTo(x, y - r + 1); g.lineTo(x, y - r - 2); g.stroke(); break;
    case 'loot': g.fillStyle = css('--d-accent', '#c7a44a'); g.beginPath(); g.arc(x, y, r - 1, 0, TAU); g.fill(); break;
    case 'levelup': g.fillStyle = css('--d-accent', '#c7a44a'); g.beginPath(); for (let i = 0; i < 10; i++) { const rr = i % 2 ? r / 2 : r, an = i * Math.PI / 5 - Math.PI / 2; g.lineTo(x + rr * Math.cos(an), y + rr * Math.sin(an)); } g.closePath(); g.fill(); break;
    case 'chicken': g.strokeStyle = css('--d-warn', '#e0a030'); g.lineWidth = 3; g.beginPath(); g.moveTo(x - r, y - r); g.lineTo(x + r, y + r); g.moveTo(x + r, y - r); g.lineTo(x - r, y + r); g.stroke(); break;
    case 'portal': g.strokeStyle = css('--d-info', '#7ab0e8'); g.beginPath(); g.arc(x, y, r, 0, TAU); g.arc(x, y, r / 2, 0, TAU); g.stroke(); break;
    case 'waypoint': g.fillStyle = css('--d-accent', '#c7a44a'); g.beginPath(); g.moveTo(x, y - r); g.lineTo(x + r, y); g.lineTo(x, y + r); g.lineTo(x - r, y); g.closePath(); g.fill(); break;
    case 'roamfail': g.strokeStyle = css('--d-bad', '#ec7d6c'); g.strokeRect(x - r + 1, y - r + 1, 2 * r - 2, 2 * r - 2); break;
    case 'gatehold': g.fillStyle = css('--d-warn', '#e0a030'); g.fillRect(x - r + 1, y - r + 1, 2 * r - 2, 2 * r - 2); break;
    case 'ignore': {
      const left = e.s ? Math.max(0, 1 - e.ago / 1000 / e.s) : 1;
      g.strokeStyle = css('--d-warn', '#e0a030'); g.beginPath(); g.arc(x, y, r + 2, 0, TAU * left); g.stroke();
      g.lineWidth = 1; g.beginPath(); g.moveTo(x - r, y + r); g.lineTo(x + r, y - r); g.moveTo(x - r, y); g.lineTo(x, y - r); g.moveTo(x, y + r); g.lineTo(x + r, y); g.stroke(); break;
    }
    case 'fallback': g.fillStyle = css('--d-text-2', '#a89d8a'); g.beginPath(); g.moveTo(x - r, y - r); g.lineTo(x + r, y - r); g.lineTo(x, y + r); g.closePath(); g.fill(); break;
    case 'yield': g.strokeStyle = css('--d-text-2', '#a89d8a'); g.beginPath(); g.moveTo(x - r, y); g.lineTo(x + r, y); g.stroke(); break;
    default: break;
  }
}

function drawTrail(g, css, s, at) {
  const plan = s.plan, win = s.win;
  if (!plan.pts.length) return;
  const P = plan.pts, col = (k) => css(ACT_TOKEN[k] || ACT_TOKEN[0], ACT_FALLBACK[k] || ACT_FALLBACK[0]);
  g.save();
  g.lineCap = 'round';
  for (let i = 1; i < P.length; i++) {
    if (P[i].e) continue;
    const k = P[i].a || 0, a = at(P[i - 1]), b = at(P[i]);
    g.strokeStyle = col(k); g.lineWidth = k === 1 ? 4 : 2; g.setLineDash(k === 6 ? [4, 3] : []);
    g.globalAlpha = k === 1 ? Math.max(0.6, trailAlpha((P[i - 1].ago + P[i].ago) / 2, win)) : trailAlpha((P[i - 1].ago + P[i].ago) / 2, win);
    g.beginPath(); g.moveTo(a[0], a[1]); g.lineTo(b[0], b[1]); g.stroke();
  }
  g.globalAlpha = 1; g.lineWidth = 2; g.setLineDash([]);
  P.forEach((p) => {
    if (!p.e) return;
    const a = at(p);
    g.strokeStyle = col(0); g.beginPath(); g.moveTo(a[0] - 4, a[1] - 4); g.lineTo(a[0] + 4, a[1] + 4); g.stroke();
  });
  g.fillStyle = css('--d-text', '#eee');
  plan.stalls.forEach((st) => { const a = at(st); g.globalAlpha = trailAlpha(st.ago, win); g.beginPath(); g.arc(a[0], a[1], 3, 0, TAU); g.fill(); });
  g.globalAlpha = 1;
  (s.evs || []).forEach((e) => { if (e.ago <= win && eventShown(e, s.evMode)) drawEvent(g, css, at(e), e); });
  g.restore();
}

// drawScene(g, css, s, w, h): the whole picture in a w x h box (CSS px). s holds map, cells, nbs, hero, units, heat, heatMode,
// cov, covOn, plan, win, evs, evMode, zoom, pan. Returns the view (the tests read the fit from it).
export function drawScene(g, css, s, w, h, unitsDraw, cv) {
  g.clearRect(0, 0, w, h);
  if (!s.map || !s.cells) return null;
  const v = fitView(worldBox(s.map, s.cells, s.nbs, (s.hero ? [s.hero] : []).concat(s.plan.pts)), w, h, s.zoom, s.hero, s.pan);
  g.globalAlpha = 0.55;
  s.nbs.forEach((n) => drawArea(g, css, n.map, n.cells, n.dx, n.dy, v));
  g.globalAlpha = 1;
  drawArea(g, css, s.map, s.cells, 0, 0, v);
  const at = (p) => [v.ox + (p.x + 0.5) * v.k, v.oy + (p.y + 0.5) * v.k];
  drawHeat(g, css, s, at, v);
  drawCoverage(g, css, s, at, v);
  drawTrail(g, css, s, at);
  if (s.hero) {
    const hh = at(s.hero);
    g.strokeStyle = css('--d-bad', '#ec7d6c'); g.lineWidth = 3;
    g.beginPath(); g.moveTo(hh[0] - 7, hh[1]); g.lineTo(hh[0] + 7, hh[1]); g.moveTo(hh[0], hh[1] - 7); g.lineTo(hh[0], hh[1] + 7); g.stroke();
  }
  if (unitsDraw && s.units.length) unitsDraw(g, at, s.units, css, cv);
  return v;
}

const WIN_LABEL = (ms) => (ms < 60000 ? '30 s' : ms / 60000 + ' min');

export function mount(root, ctx) {
  const el = ctx.el;
  const box = el('div', 'am'), cv = el('canvas', 'cv'), empty = el('div', 'empty', 'The map opens with the first unlock.');
  cv.setAttribute('role', 'img');
  cv.setAttribute('aria-label', 'Top-down map of the area; the cross is the hero');
  const btn = (text, key, label) => {
    const b = el('button', 'd-btn', text);
    b.type = 'button'; b.dataset.size = 'sm'; b.dataset.key = key; b.setAttribute('aria-label', label);
    return b;
  };
  const zin = btn('+', 'zin', 'Zoom in'), zout = btn('-', 'zout', 'Zoom out'), lay = btn('Layers', 'layers', 'Map layers: trail, coverage, heat, events');
  lay.setAttribute('aria-expanded', 'false');
  const ctl = el('div', 'ctl'), pop = el('div', 'layers');
  pop.hidden = true;
  ctl.append(zout, zin, lay);

  // Layers: the trail window, coverage, heat and events, each remembered per viewer (the keys of the `map` panel are not
  // shared: this panel's ctx.store is its own).
  const tstore = ctx.store('trailwin'), covStore = ctx.store('coverage'), evStore = ctx.store('events'), heatStore = ctx.store('heat');
  let win = 60000, covOn = false, evMode = 'key', heatMode = 'off';
  try {
    win = trailWindow(tstore.get()); covOn = covStore.get() === '1';
    const e = evStore.get(), h = heatStore.get();
    if (EVENT_MODES.indexOf(e) >= 0) evMode = e;
    if (HEAT_MODES.indexOf(h) >= 0) heatMode = h;
  } catch (e) { /* storage blocked */ }
  const winBtns = TRAIL_WINDOWS.map((ms) => btn(WIN_LABEL(ms), 'trail-' + ms, 'Trail window ' + WIN_LABEL(ms)));
  const covBtn = btn('Coverage', 'coverage', 'Show the exploration coverage grid');
  const heatBtn = btn('Heat: ' + heatMode, 'heat', 'Heat layer: kills, time spent or damage taken, per cell');
  const evBtn = btn('Events: ' + evMode, 'events', 'Event markers on the trail: key events, all, or off');
  const row = (name, ...bs) => { const r = el('div', 'row'); r.append(...bs); return [el('div', 'lbl', name), r]; };
  pop.append(...row('Trail', ...winBtns), ...row('Layers', covBtn, heatBtn, evBtn));
  // the picture and its message share a box that has a height of its own; on a phone the toolbar and the Layers list are in the flow above it (utilities.css), never a float over the tab row
  const stage = el('div', 'stage');
  stage.append(cv, empty);
  box.append(stage, ctl, pop);
  ctx.body.append(box);

  let map = null, cells = null, hero = null, units = [], version = 0, nbs = [], cov = null, heat = null, trail = [], trailArea = 0, evs = [];
  let plan = { pts: [], stalls: [] }, sig = '', zoom = 0, pan = { x: 0, y: 0 }, drag = null, wantHeat = 0, polls = 0, legendNames = [];
  const nbCache = {};
  const g = cv.getContext('2d');
  const css = (name, fb) => getComputedStyle(root).getPropertyValue(name).trim() || fb;
  const size = () => ({ w: Math.max(1, cv.clientWidth || 320), h: Math.max(1, cv.clientHeight || 240), dpr: window.devicePixelRatio || 1 });
  function draw() {
    const z = size(), bw = Math.round(z.w * z.dpr), bh = Math.round(z.h * z.dpr);
    if (cv.width !== bw || cv.height !== bh) { cv.width = bw; cv.height = bh; }
    if (g.setTransform) g.setTransform(z.dpr, 0, 0, z.dpr, 0, 0);
    drawScene(g, css, { map, cells, nbs, hero, units, heat, heatMode, cov, covOn, plan, win, evs, evMode, zoom, pan }, z.w, z.h, window.sensesUnitsDraw, cv);
    empty.hidden = !!(map && cells);
    if (map && cells) cv.title = ctx.meta.textContent || '';
  }
  function setLabel() { ctx.meta.textContent = map ? label(map.area || 'area ' + map.areaId, exploredPct(cov)) : ''; }
  function replan(force) {
    plan = trailPlan(trail, win);
    winBtns.forEach((b, i) => b.setAttribute('aria-pressed', String(TRAIL_WINDOWS[i] === win)));
    const s = trailSig(plan, win, trailArea, evs.length) + '|' + zoom + '|' + heatMode + covOn + evMode;
    if (force || s !== sig) { sig = s; draw(); }
  }
  covBtn.setAttribute('aria-pressed', String(covOn));
  winBtns.forEach((b, i) => b.addEventListener('click', () => { win = TRAIL_WINDOWS[i]; try { tstore.set(String(win)); } catch (e) { /* storage blocked */ } replan(true); }));
  covBtn.addEventListener('click', () => {
    covOn = !covOn;
    try { covStore.set(covOn ? '1' : '0'); } catch (e) { /* storage blocked */ }
    covBtn.setAttribute('aria-pressed', String(covOn)); draw();
  });
  heatBtn.addEventListener('click', () => {
    const was = heatMode;
    heatMode = HEAT_MODES[(HEAT_MODES.indexOf(heatMode) + 1) % HEAT_MODES.length];
    try { heatStore.set(heatMode); } catch (e) { /* storage blocked */ }
    heatBtn.textContent = 'Heat: ' + heatMode; heatBtn.setAttribute('aria-pressed', String(heatMode !== 'off'));
    if (was === 'off') wantHeat = 0;
    draw();
  });
  evBtn.addEventListener('click', () => {
    evMode = EVENT_MODES[(EVENT_MODES.indexOf(evMode) + 1) % EVENT_MODES.length];
    try { evStore.set(evMode); } catch (e) { /* storage blocked */ }
    evBtn.textContent = 'Events: ' + evMode; draw();
  });
  lay.addEventListener('click', () => { pop.hidden = !pop.hidden; lay.setAttribute('aria-expanded', String(!pop.hidden)); });
  zin.addEventListener('click', () => { zoom = Math.min(zoom + 1, 6); draw(); });
  zout.addEventListener('click', () => { zoom = Math.max(zoom - 1, 0); if (!zoom) pan = { x: 0, y: 0 }; draw(); });
  cv.addEventListener('pointerdown', (ev) => { drag = { x: ev.clientX, y: ev.clientY }; if (cv.setPointerCapture) cv.setPointerCapture(ev.pointerId); });
  cv.addEventListener('pointerup', () => { drag = null; });
  cv.addEventListener('pointermove', (ev) => {
    if (!drag || !map || !zoom) return;
    const u = worldBox(map, cells, nbs, (hero ? [hero] : []).concat(plan.pts)), k = fitView(u, size().w, size().h, zoom, hero, pan).k;
    pan = { x: pan.x - (ev.clientX - drag.x) / k, y: pan.y - (ev.clientY - drag.y) / k };
    drag = { x: ev.clientX, y: ev.clientY };
    draw();
  });
  if (typeof ResizeObserver === 'function') new ResizeObserver(() => draw()).observe(stage); // the slot's box changes with the window

  function takeNeighbours(list) {
    const keep = {};
    nbs = [];
    (list || []).forEach((n) => {
      let c = nbCache[n.areaId];
      if (n.map) c = nbCache[n.areaId] = { version: n.version, map: n.map, cells: decode(n.map) };
      if (!c || c.version !== n.version) return;
      keep[n.areaId] = true;
      nbs.push({ areaId: n.areaId, dx: n.dx, dy: n.dy, map: c.map, cells: c.cells });
    });
    Object.keys(nbCache).forEach((k) => { if (!keep[k]) delete nbCache[k]; });
  }

  return {
    query() {
      const heatOn = heatMode !== 'off', due = heatOn && (polls++ % 10 === 0 || wantHeat === 0);
      if (due) wantHeat = 1;
      return '?v=' + version + (due ? '&hist=1&heat=1' : '') + (Object.keys(nbCache).length ? '&nb=' + encodeURIComponent(nbParam(nbCache)) : '');
    },
    update(r) {
      if (!r) { if (!map) { empty.textContent = 'No map data right now.'; empty.hidden = false; } return; }
      if (!r.tiers_unlocked || r.tiers_unlocked.indexOf(1) < 0) { empty.textContent = 'The map opens with the first unlock.'; empty.hidden = !!map; return; }
      if (r.map) { map = r.map; cells = decode(map); version = r.version; pan = { x: 0, y: 0 }; }
      if (r.history && r.history.heat) heat = r.history.heat;
      hero = r.hero || hero;
      takeNeighbours(r.neighbours);
      units = r.units || [];
      cov = r.coverage && r.coverage.area === (map && map.areaId) ? r.coverage : null;
      if (r.trail) { trail = r.trail; trailArea = r.trail_area || 0; evs = r.events || []; legendNames = r.trail_legend || []; } else if (!r.waiting) { trail = []; evs = []; }
      empty.textContent = map ? '' : (r.waiting ? 'Waiting for the first steps...' : 'No map yet.');
      setLabel();
      replan(true);
      const share = activityShare(plan, legendNames);
      if (map) cv.title = (ctx.meta.textContent || '') + (share ? ' · ' + share : '');
    },
  };
}
