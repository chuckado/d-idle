// The Pause control (docs/pause.md): one Pause/Resume button in the "..." menu (or the links row of a page without it), a
// "Paused" banner with a Resume button, and <html data-paused> that dims the game view (pause.css). The server decides
// what Pause means (it freezes the game where the backend can, else it walks home first); this file shows the button only
// when GET /api/pause says the game can be frozen, so there is never a dead button. Space toggles it unless a field, a
// button, a link or a dialog has the focus. textContent only; no inline styles.
export const ROUTE = '/api/pause';
const POLL_MS = 3000;

// phase turns a pause view into what the page shows. shown: the button exists; paused: the game is frozen.
export function phase(v) {
  const none = { shown: false, paused: false, label: 'Pause', since: '', near: false, cap: 0, text: '' };
  if (!v || v.schema !== 'watch.pause/1') return none;
  const paused = v.frozen === true;
  const shown = paused || v.can_freeze === true;
  return {
    shown, paused, label: paused ? 'Resume' : 'Pause', since: paused && v.since ? v.since : '', near: paused && v.monsters_near === true,
    cap: paused && v.cap_min ? v.cap_min : 0, text: paused ? 'Paused' : '',
  };
}

// clock is "HH:MM" of an ISO time in the viewer's own zone ('' when it does not parse).
export function clock(iso) {
  const t = new Date(iso);
  if (!iso || isNaN(t.getTime())) return '';
  return String(t.getHours()).padStart(2, '0') + ':' + String(t.getMinutes()).padStart(2, '0');
}

// bannerText is the one line of the banner: "Paused since 14:05", the cap and the monsters hint as plain info.
export function bannerText(p) {
  let s = 'Paused';
  const c = clock(p.since);
  if (c) s += ' since ' + c;
  if (p.cap) s += ' (resumes by itself after ' + p.cap + ' min)';
  if (p.near) s += '. Monsters are near.';
  return s;
}

// spaceToggles says whether a key event is the Space shortcut: a bare Space, no repeat, and the focus is not somewhere Space
// already means something (a field, a button, a link, a menu or a dialog).
export function spaceToggles(e, doc) {
  if (!e || (e.key !== ' ' && e.code !== 'Space') || e.repeat || e.ctrlKey || e.metaKey || e.altKey || e.shiftKey || e.defaultPrevented) return false;
  const a = doc.activeElement, t = e.target || a;
  const hot = (n) => n && n.closest && n.closest('input, textarea, select, button, a, summary, [contenteditable], [role="button"], [role="menuitem"], [role="textbox"], [role="dialog"]');
  if (hot(a) || hot(t)) return false;
  if (doc.querySelector && doc.querySelector('[role="dialog"]:not([hidden]), dialog[open]')) return false;
  return true;
}

export function mountPause(d) {
  const doc = d.doc || document;
  const html = doc.documentElement;
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const btn = d.ui.btn('Pause', { size: 'sm', variant: 'ghost' });
  btn.id = 'btn-pause';
  btn.dataset.key = 'pause-toggle';
  btn.classList.add('d-pause-btn');
  btn.hidden = true;
  btn.setAttribute('aria-label', 'Pause the game');
  const menu = doc.getElementById && doc.getElementById('menu-more');
  const links = doc.querySelector && doc.querySelector('.d-band-status .d-links');
  const pair = menu && menu.querySelector && menu.querySelector('.d-menu-primary'); // the primary row (Pause, Reset) never scrolls out of reach
  if (pair) pair.insertBefore(btn, pair.firstChild);
  else if (menu) menu.insertBefore(btn, menu.firstChild);
  else if (links) links.appendChild(btn);
  const banner = el('div', 'd-pause-banner');
  banner.id = 'pause-banner';
  banner.hidden = true;
  banner.setAttribute('role', 'status');
  const glyph = el('span', 'd-pause-glyph', '❚❚');
  glyph.setAttribute('aria-hidden', 'true');
  const line = el('span', 'd-pause-line', 'Paused');
  const go = d.ui.btn('Resume', { variant: 'primary' });
  go.dataset.key = 'pause-resume';
  go.classList.add('d-pause-resume');
  banner.append(glyph, line, go);
  doc.body.appendChild(banner);

  let cur = phase(null), busy = false, timer = null, off = false, seen = false;
  function draw(v) {
    cur = phase(v);
    btn.hidden = !cur.shown;
    btn.textContent = cur.label;
    btn.setAttribute('aria-label', cur.paused ? 'Resume the game' : 'Pause the game');
    btn.setAttribute('aria-pressed', String(cur.paused));
    banner.hidden = !cur.paused;
    line.textContent = bannerText(cur);
    if (cur.paused) html.dataset.paused = '1'; else delete html.dataset.paused;
  }
  function toggle() {
    if (busy || !cur.shown) return Promise.resolve();
    busy = true; btn.disabled = go.disabled = true;
    const want = !cur.paused;
    const done = () => { busy = false; btn.disabled = go.disabled = false; };
    return d.post(ROUTE, { on: want }).then((j) => {
      if (j && j.schema === 'watch.pause/1') {
        done(); draw(j);
        if (j.error) d.toast('warn', j.error); // the reason the desk gave, as it is
        return;
      }
      // an odd or missing answer is not a verdict: read the state back and judge by that (the game may have frozen anyway)
      return d.get(ROUTE).then((v) => {
        done();
        if (v && v.schema === 'watch.pause/1') {
          draw(v);
          if (phase(v).paused !== want) d.toast('warn', (j && j.error) || 'Pause did not go through');
        } else d.toast('warn', (j && j.error) || 'no answer from the hero');
      });
    });
  }
  function poll() {
    timer = null;
    d.get(ROUTE).then((v) => {
      if (v === null) { // never answered: no pause desk on this hero (measured run), no button; a later miss keeps the last state
        if (!seen) { off = true; draw(null); return; }
        timer = d.later(poll, POLL_MS * 3);
        return;
      }
      seen = true;
      draw(v);
      if (!off) timer = d.later(poll, POLL_MS);
    });
  }
  btn.addEventListener('click', () => toggle());
  go.addEventListener('click', () => toggle());
  doc.addEventListener('keydown', (e) => {
    if (!spaceToggles(e, doc) || !cur.shown) return;
    e.preventDefault();
    toggle();
  });
  poll();
  return { draw, toggle, state: () => cur };
}
