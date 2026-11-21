// The Ethereal Vault page (docs/vault.md, UI section; design docs/design/loot-vault-2026-11-16.md section 6, slice V4).
// Read side plus Pin, Unpin, Discard and Keep (slice V5): GET /api/vault (watch.vault/1, paged, filtered on the server), GET
// /api/vault/item/<id>, POST /api/vault/{pin,unpin,discard,keep,reveal,sell,cancel} on the answer listener. Keep only QUEUES: the bot creates the item
// at its next town visit, so the page shows the states queued, creating, kept and failed from the GET, never from the reply.
// Reveal and Sell (slice V6) also only queue: the price is shown before the click (revealPlan) and the states come from the GET. The grid and the tooltip are the inventory page's (components.css cells, inventory-tip.js,
// fed the rows as an inventory view); every text comes from /watch/vault.json or the API and goes in through textContent.
// Every press is read back: the list is fetched again and what it says is what the page shows.
import { PAGE, equipPlan, equipUpgradeWord, artKey, windowRange, QUALITY, SORTS, activeCount, capacityView, createSelection, defaultFilters, fromQuery, revealBarModel, revealPlan, sellPlan, toQuery, toTipView, when } from '/watch/vault-logic.js';

export const ROUTE = '/api/vault';
const POLL_MS = 6000;
const LONG_PRESS_MS = 500;
const TYPED_ABOVE = 10; // a discard of more items than this asks for the number typed
const MAX_LOADED = 6000; // rows kept in memory at most (a vault holds a few thousand); only the rows near the view are drawn
const LOAD = 200; // the server's page size cap: scrolling near the end of what is loaded fetches the next page
const OVER_ROWS = 4; // rows drawn above and below the view

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
  const S = { eth: null, f: d.filters || defaultFilters(), items: [], total: 0, view: null, facets: { kinds: T.kinds || [], types: [] }, err: false, loading: true, sel: createSelection(),
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
  const wallet = el('span', 'vt-wallet'); wallet.dataset.key = 'wallet'; wallet.hidden = true; // the account's Ether balance (GET /api/ether), absent while loot.vault is 0
  const upgrade = btn(T.upgrade, { size: 'md', key: 'upgrade' });
  upgrade.setAttribute('aria-disabled', 'true'); upgrade.setAttribute('aria-describedby', 'vt-upgrade-why');
  const upgradeWhy = el('span', 'vt-upgrade-why', T.upgradeWhy); upgradeWhy.id = 'vt-upgrade-why';
  const downgrade = btn(T.downgrade, { size: 'md', key: 'downgrade', variant: 'ghost' }); downgrade.hidden = true;
  downgrade.setAttribute('aria-describedby', 'vt-down-why');
  const downWhy = el('span', 'vt-down-why'); downWhy.id = 'vt-down-why';
  meterBox.append(meterLabel, meter, meterText, pinsText, wallet, upgrade, downgrade, upgradeWhy, downWhy);
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
  const keptChip = btn(T.chipKept, { key: 'chip-kept', variant: 'ghost' }); keptChip.classList.add('vt-chip'); keptChip.setAttribute('aria-pressed', 'false');
  keptChip.addEventListener('click', () => { S.f.state = S.f.state === 'kept' ? '' : 'kept'; changed(); });
  chips.appendChild(keptChip);
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
  const keepB = btn(T.keep, { key: 'keep' });
  const revealB = btn(T.reveal, { key: 'reveal' }), sellB = btn(T.sell, { key: 'sell' });
  const etherB = btn(T.revealEther, { key: 'reveal-ether' }); etherB.hidden = true;
  const acts = el('div', 'vt-acts'); acts.append(keepB, revealB, etherB, sellB, pinB, unpinB, discB);
  bar.append(selBtn, selText, selAll, acts);
  // ---- the Reveal bar: one bar for the items of the identify filter, a payment picker with one cost line each ----
  const RT = T.revealBar || {};
  const RB = { pay: '', confirming: false };
  const rb = el('section', 'vt-revealbar'); rb.hidden = true; rb.setAttribute('aria-label', RT.label || '');
  const rbTitle = el('p', 'vt-rb-title'); rbTitle.setAttribute('role', 'status');
  const rbOpts = el('div', 'vt-rb-opts'); rbOpts.setAttribute('role', 'radiogroup'); rbOpts.setAttribute('aria-label', RT.payLabel || '');
  const rbGo = btn('', { key: 'reveal-bar-go' }), rbYes = btn(RT.confirmYes, { key: 'reveal-bar-yes' }), rbNo = btn(RT.confirmNo, { key: 'reveal-bar-back', variant: 'ghost' });
  const rbAsk = el('span', 'vt-rb-ask'); rbAsk.setAttribute('role', 'alert');
  const rbWhy = el('span', 'vt-rb-why');
  const rbRow = el('div', 'vt-rb-row'); rbRow.append(rbGo, rbAsk, rbYes, rbNo, rbWhy);
  const rbSkip = el('details', 'vt-rb-skip'), rbSkipSum = el('summary'), rbSkipList = el('ul'); rbSkip.append(rbSkipSum, rbSkipList);
  const rbOrder = el('div', 'vt-rb-order'), rbSteps = el('ol', 'vt-rb-steps'), rbProg = el('span', 'vt-rb-prog'), rbNote = el('span', 'vt-rb-note');
  const rbCancel = btn(RT.cancel, { key: 'reveal-bar-cancel', variant: 'ghost' });
  rbOrder.append(rbSteps, rbProg, rbNote, rbCancel);
  rb.append(rbTitle, rbOpts, rbRow, rbSkip, rbOrder);
  const keepRow = el('div', 'vt-keeprow'); keepRow.hidden = true;
  const keepLine = el('span', 'vt-keepline'); keepLine.setAttribute('role', 'status');
  const keepNowB = btn(T.keepNow, { key: 'keep-now', variant: 'ghost' }); keepNowB.title = T.keepNowWhy;
  keepRow.append(keepLine, keepNowB);
  const actRow = el('div', 'vt-keeprow vt-actrow'); actRow.hidden = true;
  const actLine = el('span', 'vt-keepline'); actLine.setAttribute('role', 'status');
  const buyB = btn(T.buyScrolls, { key: 'buy-scrolls', variant: 'ghost' }); buyB.title = T.buyScrollsWhy; buyB.hidden = true;
  const costLine = el('span', 'vt-costline'); costLine.setAttribute('role', 'status'); costLine.hidden = true;
  actRow.append(costLine, actLine, buyB);
  const promptRow = el('div', 'vt-keeprow vt-promptrow'); promptRow.hidden = true;
  const promptLine = el('span', 'vt-keepline'); promptLine.setAttribute('role', 'status');
  const promptYes = btn(T.promptYes, { key: 'prompt-yes' }), promptNo = btn(T.promptNo, { key: 'prompt-no', variant: 'ghost' });
  promptRow.append(promptLine, promptYes, promptNo);
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

  // quick filter chips: one tap for a slot (card mobilequests). A chip is the Type filter with a name and a count; the slot list is data (vault.json "slots").
  const slotRow = el('div', 'vt-slots'); slotRow.setAttribute('role', 'group'); slotRow.setAttribute('aria-label', T.slotsLabel || ''); slotRow.hidden = !(T.slots && T.slots.length);
  const SLOT_KEY = 'vt-slot';
  const slotStore = (v) => { try { if (v === undefined) return localStorage.getItem(SLOT_KEY) || ''; localStorage.setItem(SLOT_KEY, v); } catch (e) { /* storage blocked */ } return ''; };
  function pickSlot(id) { S.f.type = id; slotStore(id); changed(); }
  function drawSlots() {
    const n = {}; (S.facets.types || []).forEach((t) => { n[t.id] = t.n; });
    slotRow.replaceChildren();
    const mk = (id, label, count) => { const b = el('button', 'd-btn vt-slot', label + (count === undefined ? '' : ' ' + count)); b.type = 'button'; b.dataset.variant = 'ghost'; b.dataset.slot = id; b.setAttribute('aria-pressed', String(S.f.type === id)); b.addEventListener('click', () => pickSlot(id)); return b; };
    slotRow.appendChild(mk('', T.slotAll || 'All'));
    (T.slots || []).forEach((s) => { if (n[s.id] || S.f.type === s.id) slotRow.appendChild(mk(s.id, s.label, n[s.id] || 0)); });
  }
  { const was = slotStore(); if (was && !S.f.type) S.f.type = was; }
  const page = el('div', 'vt-page');
  page.append(head, banner, slotRow, filters, rb, bar, keepRow, actRow, promptRow, msg, body);
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
    if (!keep) main.scrollTop = 0; // a new query starts at the top
    const limit = keep ? Math.min(Math.max(S.items.length, PAGE), MAX_LOADED) : PAGE;
    const [j] = await Promise.all([fetchPage(0, Math.min(limit, 200), !S.view || !S.facets.types.length), refreshEther()]);
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
    drawSlots();
    S.sig = sig(j);
    S.sel.prune(order());
    draw();
  }
  async function loadMore() {
    if (S.items.length >= S.total || S.items.length >= MAX_LOADED) return;
    if (S.busy) return;
    S.busy = true;
    let added = 0;
    try {
      const j = await fetchPage(S.items.length, LOAD, false);
      if (!j) { setMsg('bad', T.loadError); return; }
      S.view = j; S.total = j.total;
      const have = new Set(order());
      const fresh = j.items.filter((i) => !have.has(i.id));
      added = fresh.length;
      absorb(j, fresh, false);
    } finally { S.busy = false; }
    if (added && S.items.length < S.total && rendered.last >= S.items.length - OVER_ROWS * geo.cols) loadMore(); // still near the end (a tall window): the next page
  }
  const sig = (j) => j.total + '|' + (j.capacity && j.capacity.count) + '|' + (j.items[0] && j.items[0].id) + '|' + j.mode + '|' + (j.capacity && j.capacity.pins) + '|' + (j.keep ? j.keep.queued + ',' + j.keep.creating + ',' + j.keep.done + ',' + j.keep.line : '') + '|' + (j.act ? [j.act.reveal.queued, j.act.reveal.working, j.act.reveal.done, j.act.reveal.charges, j.act.reveal.free, j.act.sell.queued, j.act.sell.working, j.act.sell.sold, j.act.reveal.line, j.act.sell.line, j.act.prompt ? j.act.prompt.ids.length : 0].join(',') : '');
  async function poll() {
    if (!S.sheet) {
      const j = await d.get(ROUTE + toQuery(S.f, { limit: 1 }));
      if (j && j.schema === 'watch.vault/1' && sig(j) !== S.sig) await reload(true);
      else if (await refreshEther()) drawHead(); // the balance moves with earning, not with the list
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
    slotRow.querySelectorAll('[data-slot]').forEach((b) => b.setAttribute('aria-pressed', String(b.dataset.slot === S.f.type)));
    sockSel.value = S.f.sockets;
    if (doc.activeElement !== lo) lo.value = S.f.ilvlMin;
    if (doc.activeElement !== hi) hi.value = S.f.ilvlMax;
    ethC.c.setAttribute('aria-pressed', String(S.f.eth)); unidC.c.setAttribute('aria-pressed', String(S.f.unid)); pinC.c.setAttribute('aria-pressed', String(S.f.pinned));
    stateSel.value = S.f.state;
    keptChip.setAttribute('aria-pressed', String(S.f.state === 'kept'));
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
    drawEther(v);
    let b = cv.banner;
    if (v && v.mode === 'off') b = T.offBanner;
    banner.hidden = !b; banner.textContent = b || ''; banner.dataset.level = v && v.mode === 'off' ? 'info' : cv.level;
  }
  // The Upgrade and Refund buttons and the wallet chip. With the wallet (GET /api/ether enabled) the buttons are real: a press opens
  // the confirm sheet. A short balance or a vault that does not fit the lower tier keeps the button aria-disabled and says why;
  // without the wallet (loot.vault 0) the old placeholder stands, with the server's reason.
  function drawEther(v) {
    const e = S.eth && S.eth.enabled ? S.eth : null, u = v && v.upgrade;
    wallet.hidden = !e; wallet.textContent = e ? e.chip : '';
    wallet.setAttribute('aria-label', e ? fmt(T.walletLabel, { chip: e.chip }) : '');
    if (!e) {
      if (u && u.next) upgrade.textContent = fmt(T.upgradeFor, { add: u.next.add, price: u.next.price, cur: u.currency });
      else upgrade.textContent = u && !u.next && v.capacity.capacity ? T.upgradeMax : T.upgrade;
      upgrade.setAttribute('aria-disabled', 'true');
      upgradeWhy.textContent = u && u.reason ? u.reason : T.upgradeWhy;
      downgrade.hidden = true; downWhy.textContent = '';
      return;
    }
    const n = e.next;
    if (n) {
      upgrade.textContent = fmt(T.upgradeFor, { add: n.add, price: n.price, cur: e.currency });
      upgrade.setAttribute('aria-disabled', String(!n.canAfford));
      upgradeWhy.textContent = n.canAfford ? '' : fmt(T.upgradeShort, { n: n.short, cur: e.currency });
    } else {
      upgrade.textContent = T.upgradeMax; upgrade.setAttribute('aria-disabled', 'true'); upgradeWhy.textContent = '';
    }
    const dn = e.downgrade, owned = e.owned && e.owned.length > 0;
    downgrade.hidden = !owned;
    if (owned) {
      downgrade.textContent = fmt(T.downgradeFor, { refund: dn.refund, cur: e.currency });
      downgrade.setAttribute('aria-disabled', String(!dn.available));
      downWhy.textContent = dn.available ? '' : (dn.reason || '');
    } else downWhy.textContent = '';
  }
  async function refreshEther() {
    const j = await d.get('/api/ether');
    const next = j && j.schema === 'watch.ether/1' ? j : null;
    const changed = JSON.stringify(next) !== JSON.stringify(S.eth);
    S.eth = next;
    return changed;
  }
  function drawBar() {
    const n = S.sel.size, ro = !S.view || S.view.mode === 'off';
    selBtn.setAttribute('aria-pressed', String(S.sel.mode));
    selBtn.textContent = S.sel.mode ? T.selectOff : T.select;
    selText.textContent = n ? fmt(T.selected, { n }) : (S.sel.mode ? T.selectNone : '');
    selAll.hidden = !S.sel.mode || !S.items.length;
    acts.hidden = !n;
    for (const b of [pinB, unpinB, discB]) { b.disabled = ro; }
    const k = S.view && S.view.keep;
    keepB.textContent = fmt(T.keepSel, { n });
    keepB.disabled = ro || !k;
    keepB.title = !k ? T.keepNote : (!k.enabled && k.why ? k.why : '');
    keepRow.hidden = !(k && k.line);
    keepLine.textContent = k && k.line ? k.line : '';
    keepNowB.hidden = !(k && (k.queued + k.creating) > 0);
    keepNowB.disabled = ro || !k || !k.enabled;
    drawAct(n, ro);
  }
  function drawRevealBar(ro) {
    const a = S.view && S.view.act, m = revealBarModel(a && a.reveal, RB.pay, RT);
    rb.hidden = !m.show;
    if (!m.show) return;
    rbTitle.textContent = m.total ? m.title : '';
    rbTitle.title = m.line || '';
    rbOpts.replaceChildren(...m.options.map((o) => {
      const b = el('button', 'vt-rb-opt'); b.type = 'button'; b.dataset.key = 'pay-' + o.id;
      b.setAttribute('role', 'radio'); b.setAttribute('aria-checked', o.checked ? 'true' : 'false');
      if (!o.ok) { b.setAttribute('aria-disabled', 'true'); b.title = o.why; }
      b.appendChild(el('strong', '', o.label)); b.appendChild(el('span', 'vt-rb-line', o.line));
      b.addEventListener('click', () => { if (o.ok) { RB.pay = o.id; RB.confirming = false; draw(); } });
      return b;
    }));
    rbOpts.hidden = !m.total;
    const ask = RB.confirming && m.confirm && m.sel === 'gold';
    rbGo.hidden = !m.total || ask; rbGo.textContent = m.go; rbGo.disabled = ro || m.goOff;
    rbGo.title = m.goOff ? m.why : '';
    rbAsk.hidden = rbYes.hidden = rbNo.hidden = !ask; rbAsk.textContent = ask ? m.confirmAsk : '';
    rbWhy.textContent = m.total && m.goOff ? m.why : ''; rbWhy.hidden = !rbWhy.textContent;
    const tail = [m.note].filter(Boolean);
    rbSkip.hidden = !(m.skippedN > 0 || tail.length);
    rbSkipSum.textContent = m.skippedN > 0 ? m.skipTitle : (RT.skipNote || '');
    rbSkipList.replaceChildren(...m.skipped.map((x) => el('li', '', x.n + ' · ' + x.why)), ...tail.map((t) => el('li', '', t)));
    const o = m.order;
    rbOrder.hidden = !o;
    if (o) {
      rbSteps.replaceChildren(...o.steps.map((x) => { const li = el('li', x.current ? 'cur' : x.past ? 'past' : '', x.label); if (x.current) li.setAttribute('aria-current', 'step'); return li; }));
      rbProg.textContent = o.progress + (o.spent ? ' · ' + o.spent + ' gold' : ''); rbNote.textContent = o.note;
      rbCancel.hidden = !o.active; rbCancel.disabled = ro;
    }
  }
  // Reveal and Sell: the buttons, the cost line for the selection, the queue sentences and the opt-in prompt
  function drawAct(n, ro) {
    const a = S.view && S.view.act, ids = S.sel.ids();
    const rp = revealPlan(S.items, ids, a), sp = sellPlan(S.items, ids);
    revealB.textContent = fmt(T.revealSel, { n: rp.n || n });
    revealB.disabled = ro || !a || !a.reveal.enabled;
    revealB.title = !a ? T.revealNote : (!a.reveal.enabled && a.reveal.why ? a.reveal.why : (!rp.n && n ? T.revealNone : ''));
    sellB.textContent = fmt(T.sellSel, { n: sp.n || n });
    sellB.disabled = ro || !a || !a.sell.enabled;
    sellB.title = !a ? T.sellNote : (!a.sell.enabled && a.sell.why ? a.sell.why : (!sp.n && n ? T.sellNone : ''));
    let cost = '';
    if (a && n && rp.n) {
      cost = (rp.n === 1 ? T.revealCostOne : fmt(T.revealCost, { n: rp.n, c: rp.charges })) + (rp.free ? T.revealFree : (rp.have >= 0 ? fmt(T.revealHave, { have: rp.have }) : '')) + (rp.short ? T.revealShort : '');
    }
    // Ether replaces the scroll: no charge is used. Shown only while Ether is a payer and Cain is not here (Cain is free).
    const per = a ? a.reveal.etherPer || 0 : 0, bal = a ? a.reveal.balance : -1, eOn = per > 0 && !(a && a.reveal.free);
    etherB.hidden = !eOn;
    if (eOn) {
      const need = rp.n * per;
      etherB.textContent = fmt(T.revealEtherSel, { n: rp.n || n });
      etherB.disabled = ro || !rp.n || bal < need;
      etherB.title = !rp.n ? T.revealNone : (bal < need ? fmt(T.etherShort, { need, have: bal }) : '');
      if (cost) cost += fmt(T.revealEtherOr, { cost: need, have: bal });
    }
    costLine.textContent = cost; costLine.hidden = !cost;
    buyB.hidden = !(cost && rp.short); buyB.disabled = ro;
    drawRevealBar(ro);
    const line = a ? [a.reveal.line, a.sell.line].filter(Boolean).join(' · ') : '';
    actLine.textContent = line;
    actRow.hidden = !(cost || line);
    const p = a && a.prompt;
    promptRow.hidden = !p || ro;
    promptLine.textContent = p ? fmt(T.promptText, { short: p.short, n: p.ids.length, gold: p.gold }) : '';
  }

  function cellLabel(it) {
    const bits = [it.name, T.qualities[it.q] || it.quality, fmt(T.ilvl, { n: it.ilvl })];
    if (!it.identified) bits.push(T.unidTag);
    if (it.pin) bits.push(T.pinTag);
    if (it.ethereal) bits.push(T.ethTag);
    if (it.keep) bits.push(T.keepStates[it.keep] + (it.keep === 'failed' && it.reason ? ': ' + it.reason : ''));
    else if (it.state !== 'recorded') bits.push(it.state);
    if (it.reveal) bits.push(T.revealStates[it.reveal] + (it.reveal === 'failed' && it.revealReason ? ': ' + it.revealReason : ''));
    if (it.sell) bits.push(T.sellStates[it.sell] + (it.sell === 'failed' && it.sellReason ? ': ' + it.sellReason : ''));
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
    const ak = artKey(it);
    if (ak) { // the game's icon when it loads; a failed load removes it and the name stays
      const im = el('img', 'ico'); im.setAttribute('alt', ''); im.setAttribute('src', '/watch/art/' + ak + '.png');
      im.addEventListener('error', () => { im.remove(); b.classList.remove('has-art'); });
      b.appendChild(im); b.classList.add('has-art');
    }
    b.appendChild(el('span', 'vt-name', it.name));
    const tags = el('span', 'vt-tags');
    if (!it.identified) { const t = el('b', 'vt-tag', '?'); t.dataset.tag = 'unid'; tags.appendChild(t); }
    if (it.pin) { const t = el('b', 'vt-tag', 'P'); t.dataset.tag = 'pin'; tags.appendChild(t); }
    if (it.ethereal) { const t = el('b', 'vt-tag', 'E'); t.dataset.tag = 'eth'; tags.appendChild(t); }
    if (it.keep) { const t = el('b', 'vt-tag', { queued: 'Q', creating: '...', kept: 'K', failed: '!' }[it.keep] || 'K'); t.dataset.tag = 'keep-' + it.keep; tags.appendChild(t); }
    else if (it.state !== 'recorded' && !it.sell) { const t = el('b', 'vt-tag', it.state.slice(0, 1).toUpperCase()); t.dataset.tag = 'state'; tags.appendChild(t); }
    if (it.reveal) { const t = el('b', 'vt-tag', { queued: 'R', working: '...', failed: '!' }[it.reveal] || 'R'); t.dataset.tag = 'reveal-' + it.reveal; tags.appendChild(t); }
    if (it.sell) { const t = el('b', 'vt-tag', { queued: '$', selling: '...', sold: '$', failed: '!' }[it.sell] || '$'); t.dataset.tag = 'sell-' + it.sell; tags.appendChild(t); }
    b.appendChild(tags);
    if (S.sel.has(it.id)) b.appendChild(el('span', 'vt-check-mark', '✓'));
    if (S.mode === 'list') {
      const meta = [T.qualities[it.q] || it.quality, fmt(T.ilvl, { n: it.ilvl })];
      if (it.area) meta.push(it.area);
      if (it.found) meta.push(when(it.found, now()));
      if (it.keep) meta.push(keepHint(it));
      if (it.reveal) meta.push(revealHint(it));
      if (it.sell) meta.push(sellHint(it));
      b.appendChild(el('span', 'vt-meta', meta.join(' · ')));
    }
    return b;
  }
  function keepHint(it) {
    return it.keep === 'queued' ? T.keepQueuedHint : it.keep === 'creating' ? T.keepCreatingHint
      : it.keep === 'kept' ? fmt(T.keepKeptHint, { where: it.reason || (it.where === 'inv' ? 'backpack' : 'stash') })
      : fmt(T.keepFailedHint, { why: it.reason || '' });
  }
  function revealHint(it) {
    return it.reveal === 'queued' ? T.revealQueuedHint : it.reveal === 'working' ? T.revealWorkingHint : fmt(T.revealFailedHint, { why: it.revealReason || '' });
  }
  function sellHint(it) {
    return it.sell === 'queued' ? T.sellQueuedHint : it.sell === 'selling' ? T.sellWorkingHint
      : it.sell === 'sold' ? (it.value ? fmt(T.sellSoldHint, { gold: it.value }) : T.sellSoldUnknown) : fmt(T.sellFailedHint, { why: it.sellReason || '' });
  }
  // ---- the windowed grid: only the rows near the view are cells; padding on the grid stands in for the rest ----
  let rendered = { first: -1, last: -1 }, geo = { cols: 1, rowH: 0 };
  function measure() { // columns and row pitch of the laid-out grid; the columns are worked out from the width the way auto-fill does (the browser reports one track for a grid with no cells yet)
    const cs = win && win.getComputedStyle ? win.getComputedStyle(grid) : null;
    if (!cs) return geo;
    const rem = parseFloat(win.getComputedStyle(doc.documentElement).fontSize) || 16, gap = parseFloat(cs.columnGap) || 0;
    const minW = (parseFloat(cs.getPropertyValue('--vt-cell-min')) || 4.75) * rem;
    const inner = grid.clientWidth - (parseFloat(cs.paddingInlineStart) || 0) - (parseFloat(cs.paddingInlineEnd) || 0);
    const cols = Math.floor((inner + gap) / (minW + gap));
    const row = parseFloat(cs.gridAutoRows) + (parseFloat(cs.rowGap) || 0);
    geo = { cols: S.mode === 'list' ? 1 : Math.max(1, cols), rowH: row > 0 ? row : 0 };
    return geo;
  }
  function rangeFor(mustIdx) {
    const g = measure(), n = S.items.length;
    let top = 0;
    if (main.getBoundingClientRect) top = main.getBoundingClientRect().top - grid.getBoundingClientRect().top;
    let r = windowRange({ n, cols: g.cols, rowH: g.rowH, top, view: main.clientHeight, over: OVER_ROWS });
    if (mustIdx >= 0 && g.rowH > 0 && (mustIdx < r.first || mustIdx >= r.last)) { // keyboard to a cell outside the window
      r = windowRange({ n, cols: g.cols, rowH: g.rowH, top: Math.floor(mustIdx / g.cols) * g.rowH, view: main.clientHeight, over: OVER_ROWS });
    }
    return r;
  }
  function drawList(mustIdx) {
    const tv = toTipView(S.items);
    S.tip = tv;
    grid.dataset.mode = S.mode;
    const had = doc.activeElement && cellOf(doc.activeElement) ? doc.activeElement.dataset.vid : '';
    const r = rangeFor(mustIdx === undefined ? -1 : mustIdx);
    rendered = { first: r.first, last: r.last };
    grid.style.setProperty('--vt-pad-top', r.padTop + 'px'); grid.style.setProperty('--vt-pad-bot', r.padBot + 'px');
    const cells = [];
    for (let i = r.first; i < r.last; i++) cells.push(cell(S.items[i], i + 1));
    grid.replaceChildren(...cells);
    if (had) { const again = grid.querySelector('[data-vid="' + (win && win.CSS && win.CSS.escape ? win.CSS.escape(had) : had) + '"]'); if (again && again.focus) again.focus({ preventScroll: true }); }
    if (typeof doc.dispatchEvent === 'function' && typeof CustomEvent !== 'undefined') doc.dispatchEvent(new CustomEvent('d2:inventory', { detail: tv.view })); // the tooltip reads rows from here
    viewBtn.textContent = S.mode === 'grid' ? T.viewList : T.viewGrid;
    if (r.last >= S.items.length - OVER_ROWS * geo.cols && S.items.length < S.total) loadMore(); // near the end of what is loaded: fetch the next page
  }
  function onScroll() { // redraw only when the view leaves the rows already drawn (cheap check: no redraw while the drawn rows cover the view)
    const g = measure(), n = S.items.length;
    if (!(g.rowH > 0) || !n) return;
    const top = Math.max(0, main.getBoundingClientRect().top - grid.getBoundingClientRect().top);
    const fr = Math.floor(top / g.rowH) * g.cols, lr = Math.ceil((top + main.clientHeight) / g.rowH) * g.cols;
    if (rendered.first < 0 || (fr < rendered.first && rendered.first > 0) || (lr > rendered.last && rendered.last < n)) drawList();
  }
  main.addEventListener('scroll', onScroll, { passive: true });
  const RO = d.ResizeObserver || (win && win.ResizeObserver); // the list's height settles after the first draw (filters fold, fonts load): draw again when it changes
  if (RO) { const ro = new RO(onScroll); ro.observe(main); ro.observe(grid); }
  if (win && win.addEventListener) win.addEventListener('resize', () => { measure(); drawList(); });
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
    if (it.keep) { const kp = el('p', 'vt-dkeep', T.keepStates[it.keep] + ': ' + keepHint(it)); kp.dataset.keep = it.keep; kp.setAttribute('role', 'status'); card.appendChild(kp); }
    if (it.reveal) { const rv = el('p', 'vt-dkeep', T.revealStates[it.reveal] + ': ' + revealHint(it)); rv.dataset.keep = it.reveal === 'failed' ? 'failed' : it.reveal; rv.setAttribute('role', 'status'); card.appendChild(rv); }
    if (it.sell) { const sl = el('p', 'vt-dkeep', T.sellStates[it.sell] + ': ' + sellHint(it)); sl.dataset.keep = it.sell === 'failed' ? 'failed' : it.sell; sl.setAttribute('role', 'status'); card.appendChild(sl); }
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
    row.appendChild(mk(T.keep, 'keep', act.keep, () => keepAct([it.id], false)));
    const eqv = S.view && S.view.equip; // one-tap Equip: only when the bot offers it (loot.vaultEquip)
    if (eqv && eqv.enabled && it.state === 'recorded') { const eb = btn(T.equip, { key: 'd-equip' }); eb.addEventListener('click', () => openEquip(it, eb)); row.appendChild(eb); }
    row.appendChild(mk(T.reveal, 'reveal', act.reveal, () => revealAct([it.id], false)));
    const ra = S.view && S.view.act && S.view.act.reveal;
    if (ra && ra.etherPer > 0 && !ra.free && !it.identified) row.appendChild(mk(fmt(T.revealEtherOne, { cost: ra.etherPer }), 'reveal-ether', { ok: ra.balance >= ra.etherPer, reason: fmt(T.etherShort, { need: ra.etherPer, have: ra.balance }) }, () => etherAct([it.id])));
    row.appendChild(mk(T.sell, 'sell', act.sell, () => openSell([it.id], row.lastChild)));
    if (it.reveal === 'queued' || it.sell === 'queued' || it.keep === 'queued') row.appendChild(mk(T.cancelQueued, 'cancel', { ok: true }, () => act1('/api/vault/cancel', [it.id]), 'ghost'));
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
  grid.addEventListener('keydown', (e) => { // arrows, Home, End and Page keys move between cells (also past the drawn window); Escape ends a selection
    const c = cellOf(e.target);
    if (e.key === 'Escape' && S.sel.size) { S.sel.clear(); drawBar(); drawList(); e.preventDefault(); return; }
    if (!c || !/^(Arrow(Left|Right|Up|Down)|Home|End|Page(Up|Down))$/.test(e.key)) return;
    const i = S.items.findIndex((x) => x.id === c.dataset.vid), cols = Math.max(1, measure().cols), page = Math.max(1, Math.floor(main.clientHeight / (geo.rowH || 1)) - 1) * cols;
    const step = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -cols, ArrowDown: cols, PageUp: -page, PageDown: page }[e.key];
    let to = e.key === 'Home' ? 0 : e.key === 'End' ? S.items.length - 1 : i + step;
    if ((e.key === 'ArrowUp' || e.key === 'ArrowDown') && (to < 0 || to >= S.items.length)) return; // a column does not wrap
    to = Math.min(S.items.length - 1, Math.max(0, to));
    if (i < 0 || to === i) return;
    e.preventDefault();
    if (to < rendered.first || to >= rendered.last) drawList(to);
    const nx = grid.querySelector('[data-vid="' + (win && win.CSS && win.CSS.escape ? win.CSS.escape(S.items[to].id) : S.items[to].id) + '"]');
    if (nx) { nx.focus(); if (nx.scrollIntoView) nx.scrollIntoView({ block: 'nearest' }); }
  });
  selBtn.addEventListener('click', () => {
    if (S.sel.mode) S.sel.clear(); else { S.sel.mode = true; S.fopen = false; syncFilters(); } // the folded filter panel closes when selecting starts: on a phone it left no room for the list and the cost line
    drawBar(); drawList();
  });
  selAll.addEventListener('click', () => { S.sel.all(order()); drawBar(); drawList(); });

  // ---- presses ----
  const REAL = /\/(keep|reveal|sell|cancel)$/;
  const CHUNK = 150; // the answer body is capped at 4 KiB: the server takes this many ids at a time
  async function act(route, ids, confirm, opts) {
    let j = null;
    const failed = [], done = [], already = [];
    for (let i = 0; i < ids.length; i += CHUNK) {
      const body = { ids: ids.slice(i, i + CHUNK) };
      if (confirm) body.confirm = true;
      if (opts && opts.now) body.now = true;
      if (opts && opts.ether) body.ether = true;
      const r = await d.post(route, body);
      if (!r) { j = null; break; }
      j = r;
      (r.done || []).forEach((x) => done.push(x));
      (r.failed || []).forEach((x) => failed.push(x));
      (r.already || []).forEach((x) => already.push(x));
      if (!r.done && !r.failed) { if (r.error && REAL.test(route)) failed.push(r); break; } // a refusal of the whole request (read only, bad body)
    }
    if (!j) { setMsg('bad', T.noAnswer); await reload(true); return false; }
    const isKeep = REAL.test(route); // the queue routes name the real cause (the server's sentence)
    const now = opts && opts.now;
    const word = route.endsWith('/pin') ? T.okPin : route.endsWith('/unpin') ? T.okUnpin : route.endsWith('/keep') ? (now ? T.okKeepNow : T.okKeep)
      : route.endsWith('/reveal') ? (opts && opts.ether ? T.okRevealEther : now ? T.okRevealNow : T.okReveal) : route.endsWith('/sell') ? (now ? T.okSellNow : T.okSell) : route.endsWith('/cancel') ? T.okCancel : T.okDiscard;
    if (!failed.length && j.ok) {
      const al = already.length;
      setMsg('good', al && al === done.length ? fmt(T.okKeepAlready, { n: al }) : fmt(word, { n: done.length - al }) + (al ? ' ' + fmt(T.okKeepAlready, { n: al }) : ''));
    } else {
      // Keep names the real cause (the server's sentence); the other routes use the short texts of vault.json
      const say = (x) => (isKeep && x.error) || T.refusals[x.code] || x.error;
      const why = failed.length ? say(failed[0]) : (say(j) || T.noAnswer);
      setMsg('bad', done.length ? fmt(T.partial, { done: done.length, failed: failed.length, why }) : why);
    }
    S.sel.clear(); if (route.endsWith('/discard') && done.includes(S.detailId)) S.detailId = '';
    await reload(true);
    if (S.detailId) openDetail(S.detailId);
    return done.length > 0;
  }
  const act1 = (route, ids) => act(route, ids, false);
  const keepAct = (ids, now) => act('/api/vault/keep', ids, false, { now });
  const revealAct = (ids, now) => act('/api/vault/reveal', ids, false, { now });
  const etherAct = (ids) => act('/api/vault/reveal', ids, false, { ether: true });
  etherB.addEventListener('click', () => {
    if (!S.sel.size || etherB.disabled) return;
    const p = revealPlan(S.items, S.sel.ids(), S.view && S.view.act);
    if (!p.n) { setMsg('info', T.revealNone); return; }
    etherAct(p.ids);
  });
  const sellAct = (ids, now) => act('/api/vault/sell', ids, true, { now });
  // The bar's press: free and Ether use the ordinary reveal routes; gold is an order the bot fills at the next town visit.
  async function goldOrder(ids, cost, confirm) {
    const r = await d.post('/api/vault/revealgold', { ids, maxGold: cost, confirm: !!confirm });
    RB.confirming = false;
    if (!r) setMsg('bad', T.noAnswer);
    else if (r.ok) setMsg('good', RT.sent);
    else setMsg('bad', ((r.failed && r.failed[0] && r.failed[0].error) || r.error || T.noAnswer));
    await reload(true);
  }
  rbGo.addEventListener('click', () => {
    const rev = S.view && S.view.act && S.view.act.reveal, m = revealBarModel(rev, RB.pay, RT);
    const o = m.sel && rev.bar.options.find((x) => x.id === m.sel);
    if (!o || m.goOff) return;
    if (o.id === 'gold') { if (o.confirm) { RB.confirming = true; draw(); return; } goldOrder(o.ids, o.cost, false); } else if (o.id === 'ether') etherAct(o.ids); else revealAct(o.ids, false);
  });
  rbYes.addEventListener('click', () => {
    const rev = S.view && S.view.act && S.view.act.reveal, m = revealBarModel(rev, RB.pay, RT);
    const o = m.sel === 'gold' && rev.bar.options.find((x) => x.id === 'gold');
    if (o) goldOrder(o.ids, o.cost, true);
  });
  rbNo.addEventListener('click', () => { RB.confirming = false; draw(); });
  rbCancel.addEventListener('click', async () => { const r = await d.post('/api/vault/revealgold/cancel', {}); setMsg(r && r.ok ? 'good' : 'bad', r && r.ok ? RT.cancelled : ((r && r.error) || T.noAnswer)); await reload(true); });
  pinB.addEventListener('click', () => { if (S.sel.size) act1('/api/vault/pin', S.sel.ids()); });
  unpinB.addEventListener('click', () => { if (S.sel.size) act1('/api/vault/unpin', S.sel.ids()); });
  discB.addEventListener('click', () => { if (S.sel.size) openDiscard(S.sel.ids(), discB); });
  keepB.addEventListener('click', () => { if (S.sel.size && !keepB.disabled) keepAct(S.sel.ids(), false); });
  revealB.addEventListener('click', () => {
    if (!S.sel.size || revealB.disabled) return;
    const p = revealPlan(S.items, S.sel.ids(), S.view && S.view.act);
    if (!p.n) { setMsg('info', T.revealNone); return; }
    revealAct(p.ids, false); // only what a reveal can still do something with; the server refuses the rest anyway
  });
  sellB.addEventListener('click', () => {
    if (!S.sel.size || sellB.disabled) return;
    const p = sellPlan(S.items, S.sel.ids());
    if (!p.n) { setMsg('info', T.sellNone); return; }
    openSell(S.sel.ids(), sellB);
  });
  buyB.addEventListener('click', async () => { // the manual Buy action of the inventory page: the bot never buys scrolls on its own
    const r = await d.post('/api/inventory/buyid', {});
    setMsg(r && !r.error ? 'good' : 'bad', r && !r.error ? T.okBuy : (r && r.error) || T.noAnswer);
  });
  promptYes.addEventListener('click', () => { const p = S.view && S.view.act && S.view.act.prompt; if (p) act('/api/vault/sell', p.ids, true, {}); });
  promptNo.addEventListener('click', async () => { await d.post('/api/vault/sellprompt', { dismiss: true }); await reload(true); });
  keepNowB.addEventListener('click', async () => { // any queued id will do: the server only needs one that is queued to ask for the trip
    const q = S.items.find((i) => i.keep === 'queued' || i.keep === 'creating');
    const j = q ? null : await d.get(ROUTE + '?state=keepq&limit=1');
    const id = q ? q.id : j && j.items && j.items[0] && j.items[0].id;
    if (id) keepAct([id], true); else setMsg('info', T.keepAlready);
  });

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
  // ---- the sell sheet: the count, the price and the pinned note; a queued sale can be cancelled before the visit ----
  function openSell(ids, opener) {
    sheetCard.replaceChildren();
    const p = sellPlan(S.items, ids), n = p.n || ids.length;
    if (p.n) ids = p.ids;
    sheetCard.setAttribute('aria-label', n === 1 ? T.sellOne : fmt(T.sellTitle, { n }));
    sheetCard.appendChild(el('h2', '', n === 1 ? T.sellOne : fmt(T.sellTitle, { n })));
    sheetCard.appendChild(el('p', '', fmt(T.sellBody, { gold: p.gold })));
    if (p.pinned) sheetCard.appendChild(el('p', '', fmt(T.sellPinnedSkip, { n: p.pinned })));
    const go = btn(n === 1 ? T.sellGoOne : fmt(T.sellGo, { n }), { key: 'sheet-go' });
    const cancel = btn(T.cancel, { key: 'sheet-cancel', variant: 'ghost' });
    const row = el('div', 'vt-row'); row.append(cancel, go); sheetCard.appendChild(row);
    S.sheet = { opener, first: cancel };
    sheetWrap.hidden = false;
    cancel.addEventListener('click', closeSheet);
    go.addEventListener('click', async () => { go.disabled = true; closeSheet(true); await sellAct(ids, false); });
    cancel.focus();
  }
  // ---- the Equip sheet: what is worn against this item, can-equip with the reason, then one button per way to wear it ----
  let eqTimer = null;
  function openEquip(it, opener) {
    sheetCard.replaceChildren();
    sheetCard.setAttribute('aria-label', fmt(T.equipTitle, { name: it.name }));
    S.sheet = { opener, first: null, equip: it.id };
    sheetWrap.hidden = false;
    equipDraw(it, null);
    equipLoad(it);
  }
  async function equipLoad(it) {
    const sh = await d.get('/api/vault/equip/' + encodeURIComponent(it.id));
    if (!S.sheet || S.sheet.equip !== it.id) return; // closed or another sheet meanwhile
    equipDraw(it, sh || false);
    const p = sh ? equipPlan(sh) : null;
    if (eqTimer && d.clear) d.clear(eqTimer);
    eqTimer = null;
    if (p && p.state === 'active' && d.later) eqTimer = d.later(() => equipLoad(it), 3000);
  }
  async function equipPress(it, slot, now) {
    const r = await d.post('/api/vault/equip', { ids: [it.id], slot, now: !!now });
    if (r && r.ok) setMsg('good', T.equipSent);
    else setMsg('bad', (r && r.failed && r.failed[0] && r.failed[0].error) || (r && r.error) || T.noAnswer);
    await reload(true);
    equipLoad(it);
  }
  function equipDraw(it, sh) {
    sheetCard.replaceChildren();
    sheetCard.appendChild(el('h2', '', fmt(T.equipTitle, { name: it.name })));
    const close = btn(T.equipClose, { key: 'sheet-cancel', variant: 'ghost' });
    close.addEventListener('click', () => closeSheet());
    if (sh === null) { sheetCard.appendChild(el('p', '', T.equipLoading)); sheetCard.appendChild(close); if (S.sheet) S.sheet.first = close; return; }
    if (sh === false) { sheetCard.appendChild(el('p', '', T.equipNone)); sheetCard.appendChild(close); if (S.sheet) S.sheet.first = close; return; }
    const pl = equipPlan(sh);
    const st = el('p', 'vt-eq-state'); st.setAttribute('role', 'status'); st.dataset.state = pl.state;
    if (pl.state === 'off') st.textContent = fmt(T.equipOff, { why: pl.offWhy });
    else if (pl.state === 'blocked') st.textContent = fmt(T.equipBlocked, { why: pl.why });
    else if (pl.state === 'reveal') st.textContent = T.equipReveal;
    else if (pl.order) {
      const o = pl.order;
      if (pl.state === 'done') st.textContent = (o.displaced && o.displaced.length ? fmt(T.equipDoneSwap, { slot: o.worn, old: o.displaced.join(', '), where: o.where || 'vault' }) : fmt(T.equipDone, { slot: o.worn }));
      else if (pl.state === 'failed') st.textContent = fmt(T.equipFailed, { why: o.reason || '' });
      else st.textContent = T.equipStates[o.state] || o.state;
    }
    if (st.textContent) sheetCard.appendChild(st);
    if (sh.req && sh.req.length) sheetCard.appendChild(el('p', 'vt-eq-req', fmt(T.equipReq, { req: sh.req.join(', ') })));
    (sh.slots || []).forEach((x) => {
      const box = el('div', 'vt-eq-slot'); box.dataset.slot = x.slot;
      box.appendChild(el('h3', '', x.label));
      const w = el('p', 'vt-eq-worn', x.worn ? T.equipWorn + ': ' + x.worn.name : T.equipWorn + ': ' + T.equipEmpty); box.appendChild(w);
      if (x.worn && x.worn.lines && x.worn.lines.length) { const ul = el('ul', 'vt-eq-lines'); x.worn.lines.forEach((l) => ul.appendChild(el('li', '', l))); box.appendChild(ul); }
      box.appendChild(el('p', 'vt-eq-can', x.can ? T.equipCan : fmt(T.equipCant, { why: x.why || '' })));
      const up = equipUpgradeWord(x);
      if (x.can && x.worn) box.appendChild(el('p', 'vt-eq-up', up === 'better' ? T.equipBetter : up === 'worse' ? T.equipWorse : T.equipUnknown));
      if (x.note) box.appendChild(el('p', 'vt-eq-note', x.note));
      sheetCard.appendChild(box);
    });
    const recall = el('label', 'vt-label'); const cb = el('input'); cb.type = 'checkbox'; cb.dataset.key = 'equip-recall'; recall.append(cb, doc.createTextNode(' ' + T.equipRecall));
    const row = el('div', 'vt-row'); row.dataset.key = 'equip-row';
    if (pl.state === 'ready') {
      pl.buttons.forEach((b) => {
        const label = b.slot === 'auto' ? (pl.ring ? T.equipGoAuto : T.equipGo) : b.slot === 'left' ? T.equipGoLeft : T.equipGoRight;
        const bb = btn(label, { key: 'equip-' + b.slot, variant: b.slot === 'auto' ? undefined : 'ghost' });
        if (!b.can) { bb.setAttribute('aria-disabled', 'true'); bb.title = b.why; }
        bb.addEventListener('click', () => { if (b.can) equipPress(it, b.slot, cb.checked); });
        row.appendChild(bb);
      });
      sheetCard.appendChild(recall);
    } else if (pl.state === 'reveal') {
      const rb = btn(T.equipRevealGo, { key: 'equip-reveal' });
      rb.addEventListener('click', () => { closeSheet(true); revealAct([it.id], false); });
      row.appendChild(rb);
    } else if (pl.state === 'active') {
      const cx = btn(T.equipCancel, { key: 'equip-cancel', variant: 'ghost' });
      cx.addEventListener('click', async () => { const r = await d.post('/api/vault/equipcancel', { ids: [it.id] }); if (!(r && r.ok)) setMsg('bad', (r && r.failed && r.failed[0] && r.failed[0].error) || T.noAnswer); await reload(true); equipLoad(it); });
      row.appendChild(cx);
    }
    const take = btn(T.equipTakeOut, { key: 'equip-takeout', variant: 'ghost' });
    take.addEventListener('click', () => { closeSheet(true); keepAct([it.id], false); });
    if (pl.state !== 'active' && pl.state !== 'done') row.appendChild(take);
    row.appendChild(close);
    sheetCard.appendChild(row);
    sheetCard.appendChild(el('p', 'vt-eq-help', T.equipNote));
    if (S.sheet) S.sheet.first = close;
  }
  function closeSheet(noFocus) {
    if (eqTimer && d.clear) { d.clear(eqTimer); eqTimer = null; }
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
  upgrade.addEventListener('click', () => {
    const e = S.eth && S.eth.enabled ? S.eth : null;
    if (!e || !e.next || !e.next.canAfford) { setMsg('info', upgradeWhy.textContent || T.upgradeMax); return; }
    openConfirm('upgrade', upgrade);
  });
  downgrade.addEventListener('click', () => {
    const e = S.eth && S.eth.enabled ? S.eth : null;
    if (!e || !e.downgrade.available) { setMsg('info', downWhy.textContent); return; }
    openConfirm('downgrade', downgrade);
  });

  // ---- the Ether sheets: confirm, press, read back ----
  function openConfirm(kind, opener) {
    const e = S.eth;
    if (!e || !e.enabled) return;
    const up = kind === 'upgrade', n = e.next, dn = e.downgrade;
    if (up ? !n : !dn.available) return;
    sheetCard.replaceChildren();
    const title = up ? T.upgradeTitle : T.downgradeTitle;
    sheetCard.setAttribute('aria-label', title);
    sheetCard.appendChild(el('h2', '', title));
    const cap = S.view && S.view.capacity ? S.view.capacity.capacity : e.capacity;
    const body = up
      ? fmt(T.upgradeBody, { add: n.add, from: e.capacity, to: n.capacity, price: n.price, cur: e.currency, bal: e.balance, after: e.balance - n.price })
      : fmt(T.downgradeBody, { step: dn.step, refund: dn.refund, cur: e.currency, bal: e.balance, after: e.balance + dn.refund, from: e.capacity, count: S.view && S.view.capacity ? S.view.capacity.count : 0, cap });
    sheetCard.appendChild(el('p', '', body));
    const go = btn(up ? fmt(T.upgradeGo, { price: n.price, cur: e.currency }) : fmt(T.downgradeGo, { refund: dn.refund, cur: e.currency }), { key: 'sheet-go', variant: up ? undefined : 'danger' });
    const cancel = btn(T.cancel, { key: 'sheet-cancel', variant: 'ghost' });
    const row = el('div', 'vt-row'); row.append(cancel, go); sheetCard.appendChild(row);
    S.sheet = { opener, first: cancel };
    sheetWrap.hidden = false;
    cancel.addEventListener('click', closeSheet);
    go.addEventListener('click', async () => { if (go.disabled) return; go.disabled = true; closeSheet(true); await pressEther(kind, up ? n.id : dn.step); });
    cancel.focus();
  }
  // The text of a refusal, with its numbers: the code decides, the numbers come from the answer.
  function etherWhy(r) {
    if (!r) return T.noAnswer;
    const t = T.refusals[r.code];
    if (!t) return r.error || T.noAnswer;
    return fmt(t, { need: r.need, count: r.count, fits: r.fitsTier, over: (r.count || 0) - (r.fitsTier || 0), cur: S.eth && S.eth.currency || '' });
  }
  async function pressEther(kind, step) {
    const r = await d.post(kind === 'upgrade' ? '/api/vault/upgrade' : '/api/vault/downgrade', { step });
    await reload(true); // the page shows what the wallet and the vault say now, not what the press hoped for
    if (r && r.ok) {
      const cur = S.eth && S.eth.currency || '';
      if (r.already) setMsg('info', T.etherAlready);
      else setMsg('good', fmt(kind === 'upgrade' ? T.okUpgrade : T.okDowngrade, { capacity: r.capacity, balance: r.balance, price: r.price, cur }));
    } else setMsg('bad', etherWhy(r));
    return r;
  }

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
  if (d.overlays) d.overlays.register('vault-sheet', sheetWrap, closeSheet); // the page's one modal: Escape and the one-at-a-time rule are the manager's (docs/ui-overlays.md)
  return { start, reload, draw, state: S, changed, openDiscard, openSell, openConfirm, pressEther, refreshEther, closeSheet, act, main, grid, loadTip };
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
  let overlays = null;
  try { overlays = (await import('/watch/overlay.js')).overlays; } catch (e) { /* the sheet closes itself on Escape */ }
  const vt = mountVault({ doc, root, overlays, win: window, get, post, texts, now: Date.now, filters: f, backHref: back, IntersectionObserver: window.IntersectionObserver, ResizeObserver: window.ResizeObserver,
    later: (fn, ms) => setTimeout(fn, ms), clear: (t) => clearTimeout(t) });
  vt.loadTip();
  vt.start();
}
if (typeof document !== 'undefined' && document.body && document.body.dataset && document.body.dataset.vault !== undefined) boot();
