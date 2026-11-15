// The log drawer's content (docs/ui/redesign-2026-11-14/briefs/S6-areamap-drawers-phone.md): four tabs, Now, Journal, Story,
// Alerts, one fetched at a time through GET /api/log?tab= (the shell's poll carries the tab in query(); the drawer only polls
// while it is open). Rows arrive newest first with "seconds ago"; the time shown is the viewer's own clock and zone.
// An open question is a Needs you row, not a log line, so this panel never shows one.
export const TABS = [['now', 'Now'], ['journal', 'Journal'], ['story', 'Story'], ['alerts', 'Alerts']];

// "14:03:22" in the viewer's zone for a row that is `ago` seconds old; past half a day the date is added.
export function when(ago, nowMs, lang) {
  const d = new Date(nowMs - Math.max(0, ago || 0) * 1000);
  const t = d.toLocaleTimeString(lang, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  return ago > 43200 ? d.toLocaleDateString(lang, { month: 'short', day: 'numeric' }) + ' ' + t : t;
}

export function mount(root, ctx) {
  const el = ctx.el, store = ctx.store('tab');
  let tab = 'now';
  try { const v = store.get(); if (TABS.some((t) => t[0] === v)) tab = v; } catch (e) { /* storage blocked */ }
  const wrap = el('div', 'lg'), bar = el('div', 'tabs'), list = el('div', 'rows'), note = el('p', 'note');
  bar.setAttribute('role', 'tablist');
  list.setAttribute('role', 'tabpanel');
  list.dataset.scroll = '1'; // the one list that may be longer than its box
  const btns = TABS.map(([id, name]) => {
    const b = el('button', 'd-btn', name);
    b.type = 'button'; b.dataset.size = 'sm'; b.dataset.key = 'tab-' + id; b.setAttribute('role', 'tab');
    b.addEventListener('click', () => pick(id));
    b.addEventListener('keydown', (e) => {
      const i = TABS.findIndex((t) => t[0] === tab), step = e.key === 'ArrowRight' ? 1 : e.key === 'ArrowLeft' ? -1 : 0;
      if (!step) return;
      e.preventDefault();
      pick(TABS[(i + step + TABS.length) % TABS.length][0], true);
    });
    return b;
  });
  bar.append(...btns);
  wrap.append(bar, list, note);
  ctx.body.append(wrap);
  function paintTabs() { btns.forEach((b, i) => { const on = TABS[i][0] === tab; b.setAttribute('aria-selected', String(on)); b.tabIndex = on ? 0 : -1; }); }
  function rows(v) {
    list.replaceChildren
      ? list.replaceChildren()
      : (list.children = []);
    const now = Date.now();
    (v.rows || []).forEach((r) => {
      const row = el('div', 'row');
      row.dataset.kind = r.kind || '';
      if (r.level) row.dataset.level = r.level;
      row.append(el('span', 'when', when(r.ago, now)), el('span', 'txt', r.text));
      list.appendChild(row);
    });
    note.textContent = (v.rows || []).length ? '' : (v.note || 'Nothing here yet.');
  }
  function pick(id, focus) {
    tab = id;
    try { store.set(id); } catch (e) { /* storage blocked */ }
    paintTabs();
    note.textContent = 'Loading...';
    if (list.replaceChildren) list.replaceChildren(); else list.children = [];
    if (ctx.refresh) ctx.refresh();
    if (focus) btns[TABS.findIndex((t) => t[0] === id)].focus();
  }
  paintTabs();
  return {
    query() { return '?tab=' + tab; },
    update(r) {
      if (!r) { if (!list.children.length) note.textContent = 'The log is not reachable right now.'; return; }
      if (r.tab !== tab) return; // a reply for the tab the viewer just left
      rows(r);
    },
  };
}
