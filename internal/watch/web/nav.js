// The one navigation (docs/ui-design-system.md "Navigation", review docs/design/fable-review-2026-11-22 section B3).
// web/nav.json lists the five destinations; this script draws them on EVERY page: a bottom bar under 700 px, links in the
// top band above (a band the page already has, else one this script adds), one aria-current, and the More sheet
// (the pages that do not fit the bar, plus whatever the page registers with dNav.group). Text goes through textContent only.
// Classic script: window.dNav on a page, module.exports for the node test (internal/watch/testdata/nav_test.cjs).
(function (g) {
  'use strict';

  const trim = (p) => (p.length > 1 && p.endsWith('/') ? p.slice(0, -1) : p);

  // pickCurrent maps an address to the destination and, inside an area, the sub-tab. Unknown addresses mark nothing.
  function pickCurrent(cfg, pathname, search) {
    const p = trim(pathname || '/');
    let dest = null;
    for (const d of cfg.destinations) if ((d.match || []).some((m) => trim(m) === p)) { dest = d.id; break; }
    let sub = null;
    const area = dest && cfg.areas && cfg.areas[dest];
    if (area) {
      const tab = new URLSearchParams(search || '').get('tab');
      const hit = area.find((s) => { const u = new URL(s.href, 'http://x'); return trim(u.pathname) === p && (u.searchParams.get('tab') || '') === (tab || ''); })
        || area.find((s) => trim(new URL(s.href, 'http://x').pathname) === p && !new URL(s.href, 'http://x').searchParams.get('tab'));
      sub = hit ? hit.id : null;
    }
    return { dest, sub };
  }

  // createNav draws into env.doc. env: { doc, cfg, loc: {pathname, search}, narrow(): bool, store: {get, set}, onNarrow(fn) }
  function createNav(env) {
    const doc = env.doc, cfg = env.cfg, T = cfg.texts || {};
    const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
    const cur = pickCurrent(cfg, env.loc.pathname, env.loc.search);
    const nav = el('nav', 'd-nav');
    nav.setAttribute('aria-label', T.main || 'Main');
    nav.dataset.nav = '';
    const sheet = el('div', 'd-menu-pop d-navsheet');
    sheet.id = 'menu-more';
    sheet.setAttribute('role', 'dialog');
    sheet.setAttribute('aria-label', T.moreTitle || 'More');
    sheet.setAttribute('popover', 'manual');
    sheet.hidden = true;
    const primary = el('div', 'd-menu-primary'); primary.dataset.slot = 'primary'; // Reset and Pause go here (the live screen); empty it takes no room
    const body = el('div', 'd-menu-body'); body.dataset.scroll = '1';
    sheet.append(primary, body);
    let moreBtn = null, host = null, band = null;
    const groups = new Map();

    const item = (d) => {
      const a = el(d.sheet ? 'button' : 'a', 'd-nav-i');
      a.dataset.nav = d.id;
      a.append(el('span', 'd-nav-g'), el('span', 'd-nav-t', d.label));
      if (d.sheet) { a.type = 'button'; a.id = 'btn-more'; a.setAttribute('aria-haspopup', 'dialog'); a.setAttribute('aria-expanded', 'false'); a.setAttribute('aria-controls', sheet.id); }
      else a.href = d.href;
      if (cur.dest === d.id) a.setAttribute('aria-current', 'page');
      return a;
    };
    for (const d of cfg.destinations) { const a = item(d); if (d.sheet) moreBtn = a; nav.appendChild(a); }

    // ---- the sheet ----
    const place = () => {
      if (!moreBtn || !moreBtn.getBoundingClientRect) return;
      const r = moreBtn.getBoundingClientRect(), w = env.width ? env.width() : 1024;
      sheet.style.setProperty('--menu-top', Math.round(r.bottom + 4) + 'px');
      sheet.style.setProperty('--menu-end', Math.max(8, Math.round(w - r.right)) + 'px');
    };
    const shown = (on) => { try { if (on && !sheet.matches(':popover-open')) sheet.showPopover(); if (!on && sheet.matches(':popover-open')) sheet.hidePopover(); } catch (e) { /* no popover support: the z-index layer stands in */ } };
    // the "more" hint under a scrolling body: set when the body has content below the fold
    const syncMore = () => { body.dataset.more = body.scrollHeight - body.clientHeight > 1 && body.scrollTop + body.clientHeight < body.scrollHeight - 1 ? '1' : '0'; };
    body.addEventListener('scroll', syncMore);
    const isOpen = () => !sheet.hidden;
    function close(back) {
      sheet.hidden = true; shown(false);
      if (moreBtn) moreBtn.setAttribute('aria-expanded', 'false');
      if (back && moreBtn) moreBtn.focus({ preventScroll: true });
    }
    function open() {
      sheet.hidden = false; place(); shown(true); syncMore();
      if (env.later) env.later(syncMore);
      if (moreBtn) moreBtn.setAttribute('aria-expanded', 'true');
      doc.dispatchEvent && doc.dispatchEvent(new g.CustomEvent('d2:nav-sheet', { detail: { open: true } }));
      const first = firstFocus();
      if (first) first.focus({ preventScroll: true });
    }
    const firstFocus = () => {
      for (const grp of groups.values()) { if (grp.body.hidden) continue; for (const n of grp.body.children) if (!n.hidden && n.focus) return n; }
      return primary.children[0] || null;
    };
    if (moreBtn) moreBtn.addEventListener('click', () => (isOpen() ? close(false) : open()));
    sheet.addEventListener('keydown', (e) => { if (e.key === 'Escape' && isOpen()) { e.preventDefault(); e.stopPropagation(); close(true); } });
    doc.addEventListener('click', (e) => { if (isOpen() && !sheet.contains(e.target) && !(moreBtn && moreBtn.contains(e.target))) close(false); });
    sheet.addEventListener('click', (e) => { const t = e.target && e.target.closest ? e.target.closest('a, button:not(.d-menu-gh)') : null; if (t) close(false); });
    if (env.onResize) env.onResize(() => { if (isOpen()) { place(); syncMore(); } });

    // a group is a titled, foldable list of nodes (the old "..." menu's Hero / Tools / Layout)
    function group(id, title, nodes, openNow) {
      let grp = groups.get(id);
      if (!grp) {
        const box = el('div', 'd-menu-group'), gb = el('div', 'd-menu-gb');
        const h = el('button', 'd-menu-gh', title);
        h.type = 'button'; h.dataset.group = id;
        h.addEventListener('click', () => { gb.hidden = !gb.hidden; h.setAttribute('aria-expanded', String(!gb.hidden)); syncMore(); });
        box.append(h, gb); body.appendChild(box);
        grp = { box, body: gb, head: h }; groups.set(id, grp);
        gb.hidden = openNow === false;
        h.setAttribute('aria-expanded', String(openNow !== false));
      }
      grp.body.replaceChildren(...nodes);
      return grp;
    }
    // the shared list: links from nav.json; an item with an id answers 'd2:nav' first (the live screen opens Help in place)
    const links = (cfg.more || []).map((m) => {
      const a = el('a', 'd-btn', m.label);
      a.dataset.size = 'sm'; a.dataset.variant = 'ghost'; a.dataset.nav = m.id; a.dataset.key = m.key || 'menu-' + m.id; a.href = m.href;
      if (cur.dest === 'more' && trim(m.href.split('?')[0]) === trim(env.loc.pathname)) a.setAttribute('aria-current', 'page');
      a.addEventListener('click', (e) => {
        const ev = new g.CustomEvent('d2:nav', { cancelable: true, detail: { id: m.id } });
        doc.dispatchEvent(ev);
        if (ev.defaultPrevented) e.preventDefault();
        else if (m.id === 'settings' && env.settings) { e.preventDefault(); env.settings(a, 'rules'); } // the Settings sheet opens in place on every page
      });
      return a;
    });
    group('pages', T.pages || 'Pages', links, true);
    if (cur.dest === 'more' && moreBtn) moreBtn.setAttribute('aria-current', 'page');

    // the theme button: a page that already has #btn-theme (the shell pages: shell.js owns its click) lends it, else this makes one
    const theme = doc.getElementById && doc.getElementById('btn-theme') || (() => {
      const b = el('button', 'd-btn', (T.theme || 'Theme: ') + 'dark'); b.id = 'btn-theme'; b.type = 'button';
      b.dataset.size = 'sm'; b.dataset.variant = 'ghost';
      b.setAttribute('aria-label', 'Switch colour theme');
      let t = (env.store && env.store.get('d2theme')) || 'dark';
      const q = new URLSearchParams(env.loc.search || '').get('theme');
      if (/^(dark|light|auto)$/.test(q || '')) t = q;
      const apply = (v) => { if (v !== 'light' && v !== 'auto') v = 'dark'; t = v; doc.documentElement.dataset.theme = v; b.textContent = (T.theme || 'Theme: ') + v; };
      b.addEventListener('click', () => { apply({ dark: 'light', light: 'auto', auto: 'dark' }[t] || 'dark'); if (env.store) env.store.set('d2theme', t); });
      apply(t);
      return b;
    })();
    group('settings', 'Settings', [theme], true);

    // ---- placement: the bar sits in <body> on a phone (nothing between it and the viewport can capture it), in the band above ----
    function mount() {
      host = doc.querySelector('[data-nav-host]') || doc.querySelector('.d-band-status .d-links');
      if (!host) { // a page without a band gets one
        band = el('header', 'd-navband');
        const brand = el('div', 'd-navbrand'); brand.append(el('span', 'd-navmark'), el('b', '', 'd.idle'));
        host = el('div', 'd-navhost');
        band.append(brand, host);
        doc.body.insertBefore(band, doc.body.firstChild);
      }
      doc.body.appendChild(sheet);
      settle();
      env.onNarrow(settle);
    }
    function settle() {
      const narrow = env.narrow();
      if (narrow) doc.body.appendChild(nav); else host.insertBefore(nav, host.firstChild);
      if (band) band.hidden = narrow;
      doc.body.dataset.nav = narrow ? 'bar' : 'band';
      if (isOpen()) place();
    }
    return { nav, sheet, primary, body, group, open, close, isOpen, mount, current: cur, moreButton: () => moreBtn,
      hide(id, on) { for (const a of links) if (a.dataset.nav === id) a.hidden = !!on; },
      drop(id) { this.hide(id, true); } };
  }

  // sub-tabs of an area (Skills: Tree, Points, Build): links, the current one marked; inserted after the page's band
  function subtabs(env, cur) {
    const area = cur.dest && env.cfg.areas && env.cfg.areas[cur.dest];
    if (!area) return null;
    const doc = env.doc, T = env.cfg.texts || {};
    const bar = doc.createElement('nav');
    bar.className = 'd-subtabs'; bar.setAttribute('aria-label', T.area || 'Section'); bar.dataset.subtabs = cur.dest;
    for (const s of area) {
      const a = doc.createElement('a');
      a.className = 'd-subtab'; a.href = s.href; a.textContent = s.label; a.dataset.sub = s.id;
      if (cur.sub === s.id) a.setAttribute('aria-current', 'page');
      bar.appendChild(a);
    }
    doc.body.dataset.subtab = cur.sub || '';
    return bar;
  }

  // openSettings loads the sheet, toast and settings scripts once (a page that has them already skips them) and opens the Settings sheet.
  function openSettings(doc, store, opener, section) {
    const load = (name, have) => have ? Promise.resolve() : new Promise((res, rej) => {
      const s = doc.createElement('script'); s.src = '/watch/' + name; s.onload = res; s.onerror = rej; doc.head.appendChild(s);
    });
    return Promise.all([load('sheet.js', g.dSheet), load('toast.js', g.dToast)]).then(() => load('settings.js', g.dSettings)).then(() => {
      const base = () => g.fetch('/api/config', { credentials: 'same-origin' }).then((r) => r.json())
        .then((c) => c.answerBase || (c.answerPort ? g.location.protocol + '//' + g.location.hostname + ':' + c.answerPort : '')).catch(() => '');
      return g.dSettings.open({ doc, sheet: g.dSheet, toast: g.dToast, fetch: g.fetch.bind(g), base, store, notif: g.Notification || null, opener, section });
    });
  }

  // ---- boot (a browser only) ----
  function boot() {
    const doc = g.document;
    let readyRes;
    const ready = new Promise((r) => { readyRes = r; });
    const api = g.dNav = { ready, nav: null };
    const store = { get(k) { try { return g.localStorage.getItem(k); } catch (e) { return null; } }, set(k, v) { try { g.localStorage.setItem(k, v); } catch (e) { /* storage blocked */ } } };
    const start = async () => {
      let cfg;
      try { cfg = await (await g.fetch('/watch/nav.json', { cache: 'no-store', credentials: 'same-origin' })).json(); } catch (e) { return; }
      const mq = g.matchMedia('(max-width: ' + (cfg.breakpoint - 0.02) + 'px)');
      const env = { doc, cfg, loc: g.location, narrow: () => mq.matches, width: () => g.innerWidth, store, later: (fn) => g.requestAnimationFrame(fn),
        onNarrow(fn) { mq.addEventListener('change', fn); }, onResize(fn) { g.addEventListener('resize', fn); } };
      env.settings = (opener, section) => openSettings(doc, store, opener, section);
      const nav = createNav(env);
      nav.mount();
      const want = new URLSearchParams(g.location.search).get('settings');
      if (want) openSettings(doc, store, null, want);
      const strip = subtabs(env, nav.current);
      if (strip) { const h = doc.querySelector('.d-band-status, .d-navband'); const at = h || null; if (at && at.after) at.after(strip); else doc.body.insertBefore(strip, doc.body.firstChild); }
      Object.assign(api, nav, { nav });
      readyRes(api);
    };
    if (doc.readyState === 'loading') doc.addEventListener('DOMContentLoaded', start); else start();
  }

  if (typeof module !== 'undefined' && module.exports) module.exports = { pickCurrent, createNav, subtabs };
  else if (g.document) boot();
})(typeof window !== 'undefined' ? window : globalThis);
