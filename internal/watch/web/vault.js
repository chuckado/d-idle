// The Ethereal Vault page (docs/vault.md, UI section; design docs/design/loot-vault-2026-11-16.md section 6, slice V4).
// Read side plus Pin, Unpin and Discard: GET /api/vault (watch.vault/1, paged, filtered on the server), GET /api/vault/item/<id>,
// POST /api/vault/{pin,unpin,discard} on the answer listener. Keep, Reveal and Sell are not here (V5, V6): the Keep button is
// drawn disabled with its reason. The grid and the tooltip are the inventory page's (components.css cells, inventory-tip.js,
// fed the rows as an inventory view); every text comes from /watch/vault.json or the API and goes in through textContent.
// Every press is read back: the list is fetched again and what it says is what the page shows.
import { PAGE, QUALITY, SORTS, activeCount, capacityView, createSelection, defaultFilters, fromQuery, toQuery, toTipView, when } from '/watch/vault-logic.js';

export const ROUTE = '/api/vault';
const POLL_MS = 6000;
const LONG_PRESS_MS = 500;
const TYPED_ABOVE = 10; // a discard of more items than this asks for the number typed
const MAX_LOADED = 600; // the page loads this many rows at most before "Show more" stops (the list is for choosing, search narrows it)

const fmt = (s, o) => String(s || '').replace(/\{(\w+)\}/g, (m, k) => (o && o[k] !== undefined ? o[k] : ''));

export function mountVault(d) {
  const doc = d.doc || document, T = d.texts, now = d.now || Date.now, win = d.win || (typeof window !== 'undefined' ? window : null);
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const btn = (label, o) => {
    const b = el('button', 'd-btn', label);
    b.type = 'button';
    b.dataset.size = (o && o.size) || 'md';
    if (o && o.variant) b.dataset.variant = o.variant;
    if (o && o.key) b.dataset.key = o.key;
    return b;
  };
  const S = { f: d.filters || defaultFilters(), items: [], total: 0, view: null, facets: { kinds: T.kinds || [], types: [] }, err: false, loading: true, sel: createSelection(),
    detail: null, detailId: '', mode: 'grid', fopen: false, sheet: null, msg: '', msgLevel: '', sig: '', lastLoaded: 0, tip: null };
  const root = d.root || doc.body;

  // ---- header: title, capacity meter, the placeholder Upgrade ----
  const head = el('header', 'vt-head');
  const back = el('a', 'd-btn', T.back);
  back.dataset.variant = 'ghost'; back.dataset.size = 'md'; back.dataset.key = 'back'; back.href = d.backHref || '/watch'; back.setAttribute('aria-label', T.backLabel);
  const h1 = el('h1', '', T.title);
  const heroName = el('span', 'vt-hero');
  const meterBox = el('div', 'vt-capacity');
  const meterLabel = el('span', 'd-sr-only', T.capacity); meterLabel.id = 'vt-meter-label';
  const meter = el('div', 'vt-meter'); meter.setAttribute('role', 'meter'); meter.setAttribute('aria-labelledby', 'vt-meter-label'); meter.setAttribute('aria-valuemin', '0');
  const fill = el('span', 'vt-fill'); meter.appendChild(fill);
  const meterText = el('span', 'vt-count');
  const pinsText = el('span', 'vt-pins');
  const upgrade = btn(T.upgrade, { size: 'md', key: 'upgrade' });
  upgrade.setAttribute('aria-disabled', 'true'); upgrade.setAttribute('aria-describedby', 'vt-upgrade-why');
  const upgradeWhy = el('span', 'vt-upgrade-why', T.upgradeWhy); upgradeWhy.id = 'vt-upgrade-why';
  meterBox.append(meterLabel, meter, meterText, pinsText, upgrade, upgradeWhy);
  head.append(back, h1, heroName, meterBox);

  const banner = el('p', 'vt-banner'); banner.hidden = true; banner.setAttribute('role', 'status');

  // ---- filters ----
  const filters = el('section', 'vt-filters'); filters.setAttribute('aria-label', T.filters);
  const frow = el('div', 'vt-frow');
  const search = el('input', 'vt-input vt-search'); search.type = 'search'; search.placeholder = T.searchHolder; search.setAttribute('aria-label', T.searchLabel); search.maxLength = 64; search.dataset.key = 'search';
  const ftoggle = btn(T.filters, { key: 'filters-toggle', variant: 'ghost' }); ftoggle.setAttribute('aria-expanded', 'false'); ftoggle.setAttribute('aria-controls', 'vt-fpanel'); ftoggle.classList.add('vt-ftoggle');
  const sortSel = el('select', 'vt-input'); sortSel.setAttribute('aria-label', T.sortLabel); sortSel.dataset.key = 'sort';
  SORTS.forEach((s) => { const o = el('option', '', T.sorts[s]); o.value = s; sortSel.appendChild(o); });
  const viewBtn = btn(T.viewList, { key: 'view-toggle', variant: 'ghost' }); viewBtn.setAttribute('aria-label', T.viewLabel); viewBtn.classList.add('vt-viewtoggle');
  frow.append(search, ftoggle, sortSel, viewBtn);
  const fpanel = el('div', 'vt-fpanel'); fpanel.id = 'vt-fpanel';
  const chips = el('div', 'vt-chips'); chips.setAttribute('role', 'group'); chips.setAttribute('aria-label', T.qualityLabel);
  const qbtn = {};
  Object.keys(QUALITY).forEach((id) => {
    const b = btn(T.qualities[id], { key: 'q-' + id, variant: 'ghost' }); b.classList.add('vt-chip'); b.dataset.q = QUALITY[id]; b.setAttribute('aria-pressed', 'false');
    b.addEventListener('click', () => {
      const n = Number(id), i = S.f.quality.indexOf(n);
      if (i < 0) S.f.quality.push(n); else S.f.quality.splice(i, 1);
      changed();
    });
    qbtn[id] = b; chips.appendChild(b);
  });
  const sel = (key, label) => { const s = el('select', 'vt-input'); s.setAttribute('aria-label', label); s.dataset.key = key; return s; };
  const kindSel = sel('kind', T.kindLabel), typeSel = sel('type', T.typeLabel), sockSel = sel('sockets', T.socketsLabel);
  const num = (key, label, hold) => { const i = el('input', 'vt-input vt-num'); i.type = 'number'; i.min = '0'; i.max = '200'; i.inputMode = 'numeric'; i.placeholder = hold || label; i.setAttribute('aria-label', label); i.dataset.key = key; return i; };
  const lo = num('ilvl-min', T.ilvlMin, T.ilvlMinHold), hi = num('ilvl-max', T.ilvlMax, T.ilvlMaxHold);
  const check = (key, label) => { const b = btn(label, { key, variant: 'ghost' }); b.classList.add('vt-check'); b.setAttribute('aria-pressed', 'false'); return { l: b, c: b }; };
  const ethC = check('eth', T.ethLabel), unidC = check('unid', T.unidLabel), pinC = check('pinned', T.pinnedLabel);
  const stateSel = sel('state', T.stateLabel);
  [['', T.stateLive], ['kept', T.stateKept], ['discarded,trimmed', T.stateGone], ['recorded,keepq,sellq,kept,sold,discarded,trimmed', T.stateAll]].forEach(([v, l]) => { const o = el('option', '', l); o.value = v; stateSel.appendChild(o); });
  const clear = btn(T.clearFilters, { key: 'clear-filters', variant: 'ghost' });
  const fgrid = el('div', 'vt-fgrid');
  fgrid.append(kindSel, typeSel, sockSel, lo, hi, stateSel, ethC.l, unidC.l, pinC.l, clear);
  fpanel.append(chips, fgrid);
  filters.append(frow, fpanel);

  // ---- selection bar ----
  const bar = el('section', 'vt-bar'); bar.setAttribute('aria-label', T.select);
  const selBtn = btn(T.select, { key: 'select-toggle', variant: 'ghost' }); selBtn.setAttribute('aria-pressed', 'false');
  const selText = el('span', 'vt-seltext'); selText.setAttribute('aria-live', 'polite');
  const selAll = btn(T.selectAll, { key: 'select-all', variant: 'ghost' });
  const pinB = btn(T.pin, { key: 'pin' }), unpinB = btn(T.unpin, { key: 'unpin' }), discB = btn(T.discard, { key: 'discard', variant: 'danger' });
  const keepB = btn(T.keep, { key: 'keep' }); keepB.setAttribute('aria-disabled', 'true'); keepB.setAttribute('aria-describedby', 'vt-keep-why');
  const keepWhy = el('span', 'vt-keep-why', T.keepNote); keepWhy.id = 'vt-keep-why';
  const acts = el('div', 'vt-acts'); acts.append(pinB, unpinB, discB);
  bar.append(selBtn, selText, selAll, acts, keepB, keepWhy);
  const msg = el('p', 'vt-msg'); msg.hidden = true; msg.setAttribute('role', 'status');

  // ---- the list and the card ----
  const body = el('div', 'vt-body');
  const main = el('main', 'vt-main'); main.id = 'vt-main'; main.tabIndex = 0; main.setAttribute('aria-label', T.title); main.dataset.vscroll = '1';
  const state = el('div', 'vt-state');
  const grid = el('div', 'vt-grid'); grid.id = 'pack'; grid.setAttribute('role', 'group'); grid.setAttribute('aria-label', T.title); // the id is the tooltip script's hook (inventory-tip.js)
  const more = btn('', { key: 'more', variant: 'ghost' });
  const shown = el('p', 'vt-shown');
  main.append(state, grid, shown, more);
  const aside = el('aside', 'vt-detail'); aside.setAttribute('aria-label', T.detailEmpty);
  body.append(main, aside);

  // ---- the confirm sheet ----
  const sheetWrap = el('div', 'vt-sheet'); sheetWrap.hidden = true;
  const sheetCard = el('div', 'vt-sheet-card'); sheetCard.setAttribute('role', 'dialog'); sheetCard.setAttribute('aria-modal', 'true');
  sheetWrap.appendChild(sheetCard);

  const page = el('div', 'vt-page');
  page.append(head, banner, filters, bar, msg, body);
  root.replaceChildren ? root.replaceChildren(page, sheetWrap) : (root.appendChild(page), root.appendChild(sheetWrap));

  // ---- loading ----
  const url = (extra) => ROUTE + toQuery(S.f, extra);
  const order = () => S.items.map((i) => i.id);
  async function fetchPage(offset, limit, facets) {
    const j = await d.get(url({ offset, limit, facets }));
    if (!j || j.schema !== 'watch.vault/1') return null;
    return j;
  }
  async function reload(keep) {
    S.loading = !S.view; S.err = false;
    const limit = keep ? Math.min(Math.max(S.items.length, PAGE), MAX_LOADED) : PAGE;
    const j = await fetchPage(0, Math.min(limit, 200), !S.view || !S.facets.types.length);
    if (!j) { S.err = true; S.loading = false; draw(); return; }
    let items = j.items;
    if (limit > 200) { // a reload that keeps a long list fetches it in pages of 200
      for (let off = 200; off < limit && items.length < j.total; off += 200) {
        const nx = await fetchPage(off, Math.min(200, limit - off), false);
        if (!nx) break;
        items = items.concat(nx.items);
      }
    }
    absorb(j, items, true);
  }
  function absorb(j, items, replace) {
    S.view = j; S.total = j.total; S.err = false; S.loading = false;
    S.items = replace ? items : S.items.concat(items);
    if (j.kinds && j.kinds.length) S.facets.kinds = j.kinds;
    if (j.types) { S.facets.types = j.types; fillTypes(); }
    S.sig = sig(j);
    S.sel.prune(order());
    draw();
  }
  async function loadMore() {
    if (S.items.length >= S.total || S.items.length >= MAX_LOADED) return;
    const j = await fetchPage(S.items.length, PAGE, false);
    if (!j) { setMsg('bad', T.loadError); return; }
    S.view = j; S.total = j.total;
    const have = new Set(order());
    absorb(j, j.items.filter((i) => !have.has(i.id)), false);
  }
  const sig = (j) => j.total + '|' + (j.capacity && j.capacity.count) + '|' + (j.items[0] && j.items[0].id) + '|' + j.mode + '|' + (j.capacity && j.capacity.pins);
  async function poll() {
    if (!S.sheet) {
      const j = await d.get(ROUTE + toQuery(S.f, { limit: 1 }));
      if (j && j.schema === 'watch.vault/1' && sig(j) !== S.sig) await reload(true);
    }
    if (d.later) d.later(poll, POLL_MS);
  }

  // ---- drawing ----
  function setMsg(level, text) { S.msg = text; S.msgLevel = level; msg.hidden = !text; msg.textContent = text; msg.dataset.level = level || ''; }
  function fillTypes() {
    const cur = S.f.type;
    kindSel.replaceChildren(); typeSel.replaceChildren();
    const k0 = el('option', '', T.kindAll); k0.value = ''; kindSel.appendChild(k0);
    S.facets.kinds.forEach((k) => { const o = el('option', '', k.label); o.value = k.id; kindSel.appendChild(o); });
    const t0 = el('option', '', T.typeAll); t0.value = ''; typeSel.appendChild(t0);
    S.facets.types.forEach((t) => { const o = el('option', '', t.label + ' (' + t.n + ')'); o.value = t.id; typeSel.appendChild(o); });
    if (cur && !S.facets.types.some((t) => t.id === cur)) { const o = el('option', '', cur); o.value = cur; typeSel.appendChild(o); }
    kindSel.value = S.f.kind; typeSel.value = cur;
  }
  function syncFilters() {
    if (doc.activeElement !== search) search.value = S.f.q;
    sortSel.value = S.f.sort;
    Object.keys(qbtn).forEach((id) => qbtn[id].setAttribute('aria-pressed', String(S.f.quality.includes(Number(id)))));
    kindSel.value = S.f.kind; typeSel.value = S.f.type;
    sockSel.value = S.f.sockets;
    if (doc.activeElement !== lo) lo.value = S.f.ilvlMin;
    if (doc.activeElement !== hi) hi.value = S.f.ilvlMax;
    ethC.c.setAttribute('aria-pressed', String(S.f.eth)); unidC.c.setAttribute('aria-pressed', String(S.f.unid)); pinC.c.setAttribute('aria-pressed', String(S.f.pinned));
    stateSel.value = S.f.state;
    const n = activeCount(S.f);
    ftoggle.textContent = n ? fmt(T.filtersN, { n }) : T.filters;
    ftoggle.setAttribute('aria-expanded', String(S.fopen));
    filters.dataset.open = S.fopen ? '1' : '';
    clear.hidden = n === 0;
  }
  const sockOpts = () => {
    sockSel.replaceChildren();
    [['', T.socketsAny], ['0', T.socketsNone], ['1', T.socketsOne]].concat([2, 3, 4, 5, 6].map((n) => [String(n), fmt(T.socketsN, { n })])).forEach(([v, l]) => { const o = el('option', '', l); o.value = v; sockSel.appendChild(o); });
  };
  sockOpts();

  function drawHead() {
    const v = S.view, c = v && v.capacity;
    heroName.textContent = v && v.hero ? v.hero : '';
    const cv = capacityView(c, T);
    meter.dataset.level = cv.level;
    meter.setAttribute('aria-valuemax', String(cv.max)); meter.setAttribute('aria-valuenow', String(cv.now)); meter.setAttribute('aria-valuetext', cv.text);
    fill.style.setProperty('--vt-pct', cv.pct + '%');
    meterText.textContent = cv.text;
    pinsText.textContent = c && c.pinLimit ? fmt(T.pinnedOf, { n: c.pins, max: c.pinLimit }) : '';
    const u = v && v.upgrade;
    if (u && u.next) upgrade.textContent = fmt(T.upgradeFor, { add: u.next.add, price: u.next.price, cur: u.currency });
    else upgrade.textContent = u && !u.next && v.capacity.capacity ? T.upgradeMax : T.upgrade;
    upgradeWhy.textContent = u && u.reason ? u.reason : T.upgradeWhy;
    let b = cv.banner;
    if (v && v.mode === 'off') b = T.offBanner;
    banner.hidden = !b; banner.textContent = b || ''; banner.dataset.level = v && v.mode === 'off' ? 'info' : cv.level;
  }
  function drawBar() {
    const n = S.sel.size, ro = !S.view || S.view.mode === 'off';
    selBtn.setAttribute('aria-pressed', String(S.sel.mode));
    selBtn.textContent = S.sel.mode ? T.selectOff : T.select;
    selText.textContent = n ? fmt(T.selected, { n }) : (S.sel.mode ? T.selectNone : '');
    selAll.hidden = !S.sel.mode || !S.items.length;
    acts.hidden = !n;
    for (const b of [pinB, unpinB, discB]) { b.disabled = ro; }
  }

  function cellLabel(it) {
    const bits = [it.name, T.qualities[it.q] || it.quality, fmt(T.ilvl, { n: it.ilvl })];
    if (!it.identified) bits.push(T.unidTag);
    if (it.pin) bits.push(T.pinTag);
    if (it.ethereal) bits.push(T.ethTag);
    if (it.state !== 'recorded') bits.push(it.state);
    return bits.join(', ');
  }
  function cell(it, n) {
    const b = el('button', 'item vt-cell'); b.type = 'button';
    b.dataset.key = 'item-' + n; b.dataset.vid = it.id; b.dataset.q = it.quality; b.classList.add('q-' + (it.quality === 'normal' || it.quality === 'superior' ? 'normal' : it.quality));
    if (it.ethereal && (it.quality === 'normal' || it.quality === 'superior')) b.className = b.className.replace('q-normal', 'q-low');
    b.setAttribute('aria-label', cellLabel(it));
    b.setAttribute('aria-pressed', S.sel.has(it.id) ? 'true' : 'false');
    if (S.sel.has(it.id)) b.classList.add('on');
    if (S.detailId === it.id) b.classList.add('sel');
    b.appendChild(el('span', 'vt-name', it.name));
    const tags = el('span', 'vt-tags');
    if (!it.identified) { const t = el('b', 'vt-tag', '?'); t.dataset.tag = 'unid'; tags.appendChild(t); }
    if (it.pin) { const t = el('b', 'vt-tag', 'P'); t.dataset.tag = 'pin'; tags.appendChild(t); }
    if (it.ethereal) { const t = el('b', 'vt-tag', 'E'); t.dataset.tag = 'eth'; tags.appendChild(t); }
    if (it.state !== 'recorded') { const t = el('b', 'vt-tag', it.state.slice(0, 1).toUpperCase()); t.dataset.tag = 'state'; tags.appendChild(t); }
    b.appendChild(tags);
    if (S.sel.has(it.id)) b.appendChild(el('span', 'vt-check-mark', '✓'));
    if (S.mode === 'list') {
      const meta = [T.qualities[it.q] || it.quality, fmt(T.ilvl, { n: it.ilvl })];
      if (it.area) meta.push(it.area);
      if (it.found) meta.push(when(it.found, now()));
      b.appendChild(el('span', 'vt-meta', meta.join(' · ')));
    }
    return b;
  }
  function drawList() {
    const tv = toTipView(S.items);
    S.tip = tv;
    grid.dataset.mode = S.mode;
    grid.replaceChildren();
    S.items.forEach((it, i) => grid.appendChild(cell(it, i + 1)));
    if (typeof doc.dispatchEvent === 'function' && typeof CustomEvent !== 'undefined') doc.dispatchEvent(new CustomEvent('d2:inventory', { detail: tv.view })); // the tooltip reads rows from here
    viewBtn.textContent = S.mode === 'grid' ? T.viewList : T.viewGrid;
  }
  function drawState() {
    state.replaceChildren();
    const v = S.view, n = activeCount(S.f);
    let card = null;
    if (S.err && !S.items.length) {
      card = [T.loadError, '', [[T.retry, () => reload(false), 'retry']]];
    } else if (S.loading) {
      card = [T.loading, '', []];
    } else if (v && !S.items.length) {
      if (n || S.f.state) card = [T.noMatchTitle, '', [[T.clearFilters, resetFilters, 'clear-empty']]];
      else if (v.mode === 'off') card = [T.offTitle, T.offBody, []];
      else card = [T.emptyTitle, T.emptyBody, []];
    }
    state.hidden = !card;
    if (!card) return;
    const box = el('div', 'vt-card');
    box.dataset.kind = S.err ? 'error' : S.loading ? 'loading' : n ? 'nomatch' : (v && v.mode === 'off' ? 'off' : 'empty');
    box.appendChild(el('h2', '', card[0]));
    if (card[1]) box.appendChild(el('p', '', card[1]));
    card[2].forEach(([label, fn, key]) => { const b = btn(label, { key }); b.addEventListener('click', fn); box.appendChild(b); });
    state.appendChild(box);
  }
  function drawMore() {
    const left = S.total - S.items.length;
    shown.textContent = S.items.length ? fmt(T.showing, { n: S.items.length, total: S.total }) : '';
    more.hidden = left <= 0 || S.items.length >= MAX_LOADED;
    more.textContent = fmt(T.more, { n: Math.min(PAGE, left) });
  }

  function lines(it, det) {
    const out = [];
    if (!it.identified) out.push(T.unidTag);
    out.push(fmt(T.ilvl, { n: it.ilvl }));
    if (it.sockets) out.push(fmt(T.socketsN, { n: it.sockets }));
    if (it.ethereal) out.push(T.ethTag);
    if (det && det.lines) det.lines.forEach((l) => { if (!out.includes(l)) out.push(l); });
    return out;
  }
  function drawDetail() {
    aside.replaceChildren();
    const it = S.items.find((x) => x.id === S.detailId);
    if (!it) { aside.appendChild(el('p', 'vt-hint', T.detailEmpty)); return; }
    const det = S.detail && S.detail.item && S.detail.item.id === it.id ? S.detail : null;
    const card = el('div', 'vt-dcard');
    const nm = el('h2', 'vt-dname', it.name); nm.dataset.q = it.quality; nm.classList.add('q-' + (it.quality === 'normal' || it.quality === 'superior' ? 'normal' : it.quality));
    card.appendChild(nm);
    const st = el('p', 'vt-dstate', (T.qualities[it.q] || it.quality) + ' · ' + it.status);
    card.appendChild(st);
    const ul = el('ul', 'vt-dlines');
    lines(it, det).forEach((l) => ul.appendChild(el('li', '', l)));
    card.appendChild(ul);
    const w = det && det.where;
    if (w && w.time) card.appendChild(el('p', 'vt-dfound', fmt(w.area ? T.foundIn : T.found, { when: when(w.time, now()), area: w.area || '' })));
    else if (it.found) card.appendChild(el('p', 'vt-dfound', fmt(it.area ? T.foundIn : T.found, { when: when(it.found, now()), area: it.area || '' })));
    const row = el('div', 'vt-drow');
    const act = det && det.actions || {};
    const mk = (label, key, a, fn, variant) => { const b = btn(label, { key: 'd-' + key, variant }); if (!a || !a.ok) { b.setAttribute('aria-disabled', 'true'); if (a && a.reason) b.title = a.reason; } b.addEventListener('click', () => { if (a && a.ok) fn(); }); return b; };
    row.append(it.pin ? mk(T.unpin, 'unpin', act.unpin, () => act1('/api/vault/unpin', [it.id])) : mk(T.pin, 'pin', act.pin, () => act1('/api/vault/pin', [it.id])),
      mk(T.discard, 'discard', act.discard, () => openDiscard([it.id], row.lastChild), 'danger'));
    const keep = btn(T.keep, { key: 'd-keep' }); keep.setAttribute('aria-disabled', 'true'); keep.title = T.keepNote;
    row.appendChild(keep);
    card.appendChild(row);
    aside.appendChild(card);
  }
  function draw() {
    syncFilters(); drawHead(); drawBar(); drawState(); drawList(); drawMore(); drawDetail();
    const empty = !S.items.length;
    grid.hidden = empty;
  }

  // ---- filters changed ----
  function changed() {
    const q = toQuery(S.f);
    if (win && win.history && win.history.replaceState) { try { win.history.replaceState(null, '', (win.location ? win.location.pathname : '/watch/vault') + q + (S.detailId ? '#' + S.detailId : '')); } catch (e) { /* a sandboxed frame */ } }
    S.sel.clear(); S.items = []; S.total = 0;
    reload(false);
    syncFilters();
  }
  function resetFilters() { const keep = S.f.sort; S.f = defaultFilters(); S.f.sort = keep; changed(); }
  let searchTimer = null;
  search.addEventListener('input', () => {
    S.f.q = search.value;
    if (d.later) { if (searchTimer) d.clear && d.clear(searchTimer); searchTimer = d.later(() => { searchTimer = null; changed(); }, 250); } else changed();
  });
  search.addEventListener('keydown', (e) => { if (e.key === 'Escape' && search.value) { search.value = ''; S.f.q = ''; changed(); } });
  sortSel.addEventListener('change', () => { S.f.sort = sortSel.value; changed(); });
  kindSel.addEventListener('change', () => { S.f.kind = kindSel.value; changed(); });
  typeSel.addEventListener('change', () => { S.f.type = typeSel.value; changed(); });
  sockSel.addEventListener('change', () => { S.f.sockets = sockSel.value; changed(); });
  stateSel.addEventListener('change', () => { S.f.state = stateSel.value; changed(); });
  lo.addEventListener('change', () => { S.f.ilvlMin = lo.value; changed(); });
  hi.addEventListener('change', () => { S.f.ilvlMax = hi.value; changed(); });
  ethC.c.addEventListener('click', () => { S.f.eth = !S.f.eth; changed(); });
  unidC.c.addEventListener('click', () => { S.f.unid = !S.f.unid; changed(); });
  pinC.c.addEventListener('click', () => { S.f.pinned = !S.f.pinned; changed(); });
  clear.addEventListener('click', resetFilters);
  ftoggle.addEventListener('click', () => { S.fopen = !S.fopen; syncFilters(); });
  viewBtn.addEventListener('click', () => { S.mode = S.mode === 'grid' ? 'list' : 'grid'; drawList(); });
  more.addEventListener('click', loadMore);
  if (typeof d.IntersectionObserver === 'function') { // load the next page when the end of the list scrolls into view
    new d.IntersectionObserver((es) => { if (es.some((e) => e.isIntersecting)) loadMore(); }, { root: main }).observe(shown);
  }

  // ---- selection and the card ----
  function openDetail(id) {
    S.detailId = id; S.detail = null;
    drawList(); drawDetail();
    d.get('/api/vault/item/' + encodeURIComponent(id)).then((j) => { if (j && j.item && S.detailId === id) { S.detail = j; drawDetail(); } });
  }
  function cellOf(target) { return target && target.closest ? target.closest('.vt-cell') : null; }
  let lpTimer = null, lpFired = false;
  grid.addEventListener('pointerdown', (e) => {
    const c = cellOf(e.target); lpFired = false;
    if (!c || (e.pointerType === 'mouse' && e.button !== 0) || !d.later) return;
    lpTimer = d.later(() => { lpTimer = null; lpFired = true; S.sel.toggle(c.dataset.vid); drawBar(); drawList(); }, LONG_PRESS_MS);
  });
  const endPress = () => { if (lpTimer && d.clear) d.clear(lpTimer); lpTimer = null; };
  grid.addEventListener('pointerup', endPress); grid.addEventListener('pointercancel', endPress); grid.addEventListener('pointermove', (e) => { if (lpTimer && (e.movementX || e.movementY) && Math.abs(e.movementX) + Math.abs(e.movementY) > 6) endPress(); });
  grid.addEventListener('contextmenu', (e) => { if (e.pointerType === 'touch' || lpFired) e.preventDefault(); });
  grid.addEventListener('click', (e) => {
    const c = cellOf(e.target);
    if (!c) return;
    if (lpFired) { lpFired = false; e.preventDefault(); return; }
    const id = c.dataset.vid;
    if (e.shiftKey) S.sel.range(order(), id);
    else if (e.ctrlKey || e.metaKey || S.sel.mode) S.sel.toggle(id);
    else { openDetail(id); return; }
    if (!S.sel.size) S.sel.mode = false;
    drawBar(); drawList();
  });
  grid.addEventListener('keydown', (e) => { // arrows move between cells; Escape ends a selection
    const c = cellOf(e.target);
    if (e.key === 'Escape' && S.sel.size) { S.sel.clear(); drawBar(); drawList(); e.preventDefault(); return; }
    if (!c || !/^Arrow(Left|Right|Up|Down)$/.test(e.key)) return;
    const cells = [...grid.querySelectorAll('.vt-cell')], i = cells.indexOf(c), top = cells[0].offsetTop;
    const cols = Math.max(1, cells.findIndex((x) => x.offsetTop !== top) < 0 ? cells.length : cells.findIndex((x) => x.offsetTop !== top));
    const step = e.key === 'ArrowLeft' ? -1 : e.key === 'ArrowRight' ? 1 : e.key === 'ArrowUp' ? -cols : cols;
    const nx = cells[i + step];
    if (nx) { nx.focus(); e.preventDefault(); }
  });
  selBtn.addEventListener('click', () => { if (S.sel.mode) S.sel.clear(); else S.sel.mode = true; drawBar(); drawList(); });
  selAll.addEventListener('click', () => { S.sel.all(order()); drawBar(); drawList(); });

  // ---- presses ----
  const CHUNK = 150; // the answer body is capped at 4 KiB: the server takes this many ids at a time
  async function act(route, ids, confirm) {
    let j = null;
    const failed = [], done = [];
    for (let i = 0; i < ids.length; i += CHUNK) {
      const r = await d.post(route, confirm ? { ids: ids.slice(i, i + CHUNK), confirm: true } : { ids: ids.slice(i, i + CHUNK) });
      if (!r) { j = null; break; }
      j = r;
      (r.done || []).forEach((x) => done.push(x));
      (r.failed || []).forEach((x) => failed.push(x));
      if (!r.done && !r.failed) break; // a refusal of the whole request (read only, bad body)
    }
    if (!j) { setMsg('bad', T.noAnswer); await reload(true); return false; }
    const word = route.endsWith('/pin') ? T.okPin : route.endsWith('/unpin') ? T.okUnpin : T.okDiscard;
    if (!failed.length && j.ok) setMsg('good', fmt(word, { n: done.length }));
    else {
      const why = failed.length ? (T.refusals[failed[0].code] || failed[0].error) : (T.refusals[j.code] || j.error || T.noAnswer);
      setMsg('bad', done.length ? fmt(T.partial, { done: done.length, failed: failed.length, why }) : why);
    }
    S.sel.clear(); if (route.endsWith('/discard') && done.includes(S.detailId)) S.detailId = '';
    await reload(true);
    if (S.detailId) openDetail(S.detailId);
    return done.length > 0;
  }
  const act1 = (route, ids) => act(route, ids, false);
  pinB.addEventListener('click', () => { if (S.sel.size) act1('/api/vault/pin', S.sel.ids()); });
  unpinB.addEventListener('click', () => { if (S.sel.size) act1('/api/vault/unpin', S.sel.ids()); });
  discB.addEventListener('click', () => { if (S.sel.size) openDiscard(S.sel.ids(), discB); });

  // ---- the discard sheet: a modal with the count, a typed number above ten items, Escape and a focus loop ----
  function openDiscard(ids, opener) {
    sheetCard.replaceChildren();
    const n = ids.length, typed = n > TYPED_ABOVE;
    sheetCard.setAttribute('aria-label', n === 1 ? T.discardOne : fmt(T.discardTitle, { n }));
    sheetCard.appendChild(el('h2', '', n === 1 ? T.discardOne : fmt(T.discardTitle, { n })));
    sheetCard.appendChild(el('p', '', T.discardBody));
    let input = null;
    const go = btn(n === 1 ? T.discardGoOne : fmt(T.discardGo, { n }), { key: 'sheet-go', variant: 'danger' });
    if (typed) {
      const lab = el('label', 'vt-label', fmt(T.discardTyped, { n })); input = el('input', 'vt-input'); input.inputMode = 'numeric'; input.autocomplete = 'off'; input.dataset.key = 'sheet-typed';
      lab.appendChild(input); sheetCard.appendChild(lab);
      go.disabled = true;
      input.addEventListener('input', () => { go.disabled = input.value.trim() !== String(n); });
    }
    const cancel = btn(T.cancel, { key: 'sheet-cancel', variant: 'ghost' });
    const row = el('div', 'vt-row'); row.append(cancel, go); sheetCard.appendChild(row);
    S.sheet = { opener, first: cancel };
    sheetWrap.hidden = false;
    cancel.addEventListener('click', closeSheet);
    go.addEventListener('click', async () => { if (go.disabled) return; go.disabled = true; closeSheet(true); await act('/api/vault/discard', ids, true); });
    (input || cancel).focus();
  }
  function closeSheet(noFocus) {
    if (!S.sheet) return;
    const o = S.sheet.opener; S.sheet = null; sheetWrap.hidden = true; sheetCard.replaceChildren();
    if (noFocus !== true && o && o.focus) o.focus();
  }
  sheetWrap.addEventListener('click', (e) => { if (e.target === sheetWrap) closeSheet(); });
  doc.addEventListener('keydown', (e) => {
    if (!S.sheet) return;
    if (e.key === 'Escape') { e.preventDefault(); closeSheet(); return; }
    if (e.key === 'Tab') { // keep the focus inside the sheet
      const f = [...sheetCard.querySelectorAll('button:not([disabled]), input')];
      if (!f.length) return;
      const i = f.indexOf(doc.activeElement);
      if (e.shiftKey && i <= 0) { f[f.length - 1].focus(); e.preventDefault(); } else if (!e.shiftKey && (i < 0 || i === f.length - 1)) { f[0].focus(); e.preventDefault(); }
    }
  });
  // keep the disabled look from the aria-disabled placeholders: a press says why instead of doing nothing
  for (const [b, why] of [[upgrade, () => upgradeWhy.textContent], [keepB, () => T.keepNote]]) b.addEventListener('click', () => setMsg('info', why()));

  async function start() {
    syncFilters(); fillTypes(); draw();
    await reload(false);
    const h = (win && win.location && win.location.hash || '').slice(1);
    if (h && S.items.some((i) => i.id === h)) openDetail(h);
    else if (h) { // a deep link to an item that is not on the first page: ask for it, and show it when the vault has it
      const j = await d.get('/api/vault/item/' + encodeURIComponent(h));
      if (j && j.item) { S.items = [j.item].concat(S.items); S.detailId = h; S.detail = j; draw(); }
    }
    if (d.later) d.later(poll, POLL_MS);
    return S;
  }
  // the tooltip script wires the grid by its id when it loads, so it loads once the grid exists, then gets the rows again
  function loadTip() {
    if (!doc.createElement || !doc.body) return;
    const sc = doc.createElement('script');
    sc.src = '/watch/inventory-tip.js';
    sc.onload = () => { if (S.tip && typeof CustomEvent !== 'undefined') doc.dispatchEvent(new CustomEvent('d2:inventory', { detail: S.tip.view })); };
    doc.body.appendChild(sc);
  }
  return { start, reload, draw, state: S, changed, openDiscard, closeSheet, act, main, grid, loadTip };
}

// ---- the page in a browser ----
async function boot() {
  const doc = document, qs = new URLSearchParams(location.search);
  let theme = 'dark';
  try { theme = localStorage.getItem('d2theme') || 'dark'; } catch (e) { /* storage blocked */ }
  if (/^(dark|light|auto)$/.test(qs.get('theme') || '')) theme = qs.get('theme');
  doc.documentElement.dataset.theme = /^(light|auto)$/.test(theme) ? theme : 'dark';
  const json = (r) => r.json().catch(() => null);
  const get = (p) => fetch(p, { cache: 'no-store' }).then(json).catch(() => null);
  let cfg = {};
  try { cfg = (await get('/api/config')) || {}; } catch (e) { /* the post says so */ }
  const base = cfg.answerBase || (cfg.answerPort ? location.protocol + '//' + location.hostname + ':' + cfg.answerPort : '');
  const post = (p, b) => fetch(base + p, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(b || {}) }).then(json).catch(() => null);
  const texts = await get('/watch/vault.json');
  const root = doc.getElementById('vt-root');
  if (!texts || !root) { if (root) root.textContent = 'The Ethereal Vault page could not load its texts.'; return; }
  const f = fromQuery(location.search);
  const back = qs.get('theme') ? '/watch?theme=' + qs.get('theme') : '/watch';
  const vt = mountVault({ doc, root, win: window, get, post, texts, now: Date.now, filters: f, backHref: back, IntersectionObserver: window.IntersectionObserver,
    later: (fn, ms) => setTimeout(fn, ms), clear: (t) => clearTimeout(t) });
  vt.loadTip();
  vt.start();
}
if (typeof document !== 'undefined' && document.body && document.body.dataset && document.body.dataset.vault !== undefined) boot();
