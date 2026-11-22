// The Characters page (docs/charselect.md Phase B, docs/charselect-ui.md): list, switch, create, delete (archive first,
// the typed name) and restore, over GET /api/characters (watch.characters/1) and its POST routes. Every text comes from
// /watch/characters.json or the API; textContent only. Every press is read back: the list is fetched again and what it
// says is what the page shows, so a refusal is never drawn as a success. The class table and the name rule come from
// GET /api/characters/rules, never from this file.
export const ROUTE = '/api/characters';
export const RULES = '/api/characters/rules';
const POLL_MS = 2000;
const GIVE_UP_MS = 4 * 60 * 1000; // a switch restarts the bot: polls fail for a while, then the strip says so

export const fmt = (s, o) => String(s || '').replace(/\{(\w+)\}/g, (m, k) => (o && o[k] !== undefined ? o[k] : ''));

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
// when turns a timestamp into words ("3 h ago"); '' when it is not a date.
export function when(iso, nowMs) {
  const t = Date.parse(iso || '');
  if (!isFinite(t)) return '';
  const s = Math.max(0, Math.round(((nowMs === undefined ? Date.now() : nowMs) - t) / 1000));
  if (s < 60) return 'just now';
  if (s < 3600) return Math.floor(s / 60) + ' min ago';
  if (s < 86400) return Math.floor(s / 3600) + ' h ago';
  if (s < 172800) return 'yesterday';
  if (s < 30 * 86400) return Math.floor(s / 86400) + ' days ago';
  const d = new Date(t);
  return d.getUTCDate() + ' ' + MONTHS[d.getUTCMonth()] + ' ' + d.getUTCFullYear();
}

// checkName is the live hint of the name field: the rule of the API (rules.name), then duplicates. level: hint, bad, good.
export function checkName(name, rules, chars, T) {
  const r = (rules && rules.name) || { min: 2, max: 15, separators: '-_' };
  const o = { min: r.min, max: r.max, name };
  if (!name) return { ok: false, level: 'hint', text: fmt(T.nameHint, o) };
  if (name.length < r.min) return { ok: false, level: 'bad', text: fmt(T.nameTooShort, o) };
  if (name.length > r.max) return { ok: false, level: 'bad', text: fmt(T.nameTooLong, o) };
  let sep = 0;
  for (let i = 0; i < name.length; i++) {
    const c = name[i];
    if (/[A-Za-z]/.test(c)) continue;
    if (r.separators.indexOf(c) < 0) return { ok: false, level: 'bad', text: fmt(T.nameChars, o) };
    if (++sep > 1 || i === 0 || i === name.length - 1) return { ok: false, level: 'bad', text: fmt(T.nameSep, o) };
  }
  const low = name.toLowerCase();
  if ((chars || []).some((c) => String(c.name).toLowerCase() === low)) return { ok: false, level: 'bad', text: fmt(T.nameTaken, o) };
  return { ok: true, level: 'good', text: fmt(T.nameOk, o) };
}

// refusalText is the words of a refused press: the page's text for the route's code, else the route's own error.
export function refusalText(j, T) {
  if (!j || typeof j !== 'object') return T.noAnswer;
  return (T.refusals && T.refusals[j.code]) || j.error || T.noAnswer;
}

// strip turns the list view and our own pending press into what the Switching strip shows.
export function strip(v, pend, nowMs, T) {
  const none = { active: false, failed: false, text: '' };
  if (v && v.switch) return { active: true, failed: false, text: fmt(v.switch.creating ? T.stripCreate : T.stripSwitch, { to: v.switch.to, from: v.switch.from }) };
  if (!pend) return none;
  if (v && v.active === pend.to) return none;
  if (nowMs - pend.at > GIVE_UP_MS) return { active: false, failed: true, text: fmt(T.stripFail, { to: pend.to }) };
  return { active: true, failed: false, text: fmt(T.stripWait, { to: pend.to }) };
}

export function mountCharacters(d) {
  const doc = d.doc || document, T = d.texts, now = d.now || Date.now;
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const btn = (label, o) => {
    const b = el('button', 'd-btn', label);
    b.type = 'button';
    b.dataset.size = (o && o.size) || 'md';
    if (o && o.variant) b.dataset.variant = o.variant;
    if (o && o.key) b.dataset.key = o.key;
    return b;
  };
  const S = { view: null, rules: null, err: null, pend: null, waiting: false, last: '', keys: {}, sheet: null };

  // ---- the page frame ----
  const root = d.root || doc.body;
  const head = el('header', 'cs-head');
  const title = el('h1', '', T.title);
  const nowEl = el('span', 'cs-now', '');
  const newTop = btn(T.newHero, { variant: 'primary', key: 'new-top' });
  head.append(title, nowEl, newTop);
  const stripEl = el('div', 'cs-strip', '');
  stripEl.setAttribute('role', 'status'); stripEl.hidden = true; stripEl.dataset.key = 'strip';
  const noteEl = el('div', 'cs-note', '');
  noteEl.hidden = true;
  const noteText = el('span', 'cs-note-text', '');
  const noteX = btn(T.dismiss, { variant: 'ghost', key: 'note-x' });
  noteEl.append(noteText, noteX);
  const main = el('main', 'cs-main');
  main.id = 'cs-main'; main.tabIndex = -1; main.dataset.vscroll = '1';
  const sheetWrap = el('div', 'cs-sheet');
  sheetWrap.id = 'cs-sheet'; sheetWrap.hidden = true;
  sheetWrap.setAttribute('role', 'dialog'); sheetWrap.setAttribute('aria-modal', 'true'); sheetWrap.setAttribute('aria-labelledby', 'cs-sheet-title');
  const sheetCard = el('div', 'cs-sheet-card');
  sheetCard.dataset.vscroll = '1';
  sheetWrap.appendChild(sheetCard);
  root.append(head, stripEl, noteEl, main, sheetWrap);

  const setNote = (sev, text) => {
    noteEl.hidden = !text;
    noteEl.dataset.sev = sev || 'info';
    noteEl.setAttribute('role', sev === 'warn' ? 'alert' : 'status');
    noteText.textContent = text || '';
  };
  noteX.addEventListener('click', () => { setNote('', ''); main.focus(); });

  const caps = () => (S.view && S.view.caps) || {};
  const busy = () => !!(S.view && S.view.switch) || !!S.pend;
  const className = (id, fallback) => {
    const c = S.rules && (S.rules.classes || []).find((x) => x.id === id);
    return c ? c.name : fallback || id || '';
  };

  // ---- the list ----
  function card(c) {
    const li = el('li', 'cs-card');
    li.dataset.class = c.class || ''; li.dataset.key = 'char-' + c.name;
    if (c.active) li.dataset.active = '1';
    const top = el('div', 'cs-top');
    const mark = el('span', 'cs-mark', (c.className || c.class || '?').slice(0, 1));
    mark.setAttribute('aria-hidden', 'true');
    const id = el('div', 'cs-id');
    const name = el('h3', 'cs-name', c.name);
    id.append(name, el('p', 'cs-sub', fmt(T.level, { level: c.level, class: c.className || c.class })));
    top.append(mark, id);
    if (c.active) { const b = el('span', 'cs-badge', T.playing); b.dataset.key = 'badge-active'; top.appendChild(b); }
    const chips = el('div', 'cs-chips');
    const chip = (text, lvl) => { const s = el('span', 'cs-chip', text); if (lvl) s.dataset.level = lvl; chips.appendChild(s); };
    chip(c.expansion ? T.expansion : T.classic);
    if (c.hardcore) chip(T.hardcore, 'warn');
    if (c.died) chip(T.died, 'bad');
    if (c.damaged) chip(T.damaged, 'bad');
    if (c.archived && c.archived.length) chip(fmt(T.archivedCopies, { n: c.archived.length }));
    const w = when(c.lastPlayed, now());
    const last = el('p', 'cs-last', w ? fmt(T.lastPlayed, { when: w }) : T.lastUnknown);
    const acts = el('div', 'cs-acts');
    const del = btn(T.delete, { variant: 'danger', key: 'delete-' + c.name });
    del.setAttribute('aria-label', T.delete + ' ' + c.name);
    del.disabled = busy() && !c.active;
    del.addEventListener('click', () => openDelete(c, del));
    S.keys[del.dataset.key] = del;
    if (!c.active) { // the hero in the game has the badge instead of a Switch
      const sw = btn(T.switch, { variant: 'primary', key: 'switch-' + c.name });
      sw.setAttribute('aria-label', fmt(T.switchTo, { name: c.name }));
      sw.disabled = !caps().switch || !!c.damaged || busy();
      sw.addEventListener('click', () => openSwitch(c, sw));
      S.keys[sw.dataset.key] = sw;
      acts.appendChild(sw);
    }
    acts.appendChild(del);
    li.append(top, chips, last);
    if (c.damaged) li.appendChild(el('p', 'cs-warn', c.damaged));
    li.appendChild(acts);
    return li;
  }

  function newTile() {
    const li = el('li', 'cs-card cs-newcard');
    li.dataset.key = 'new-tile';
    const b = btn(T.newTile, { key: 'new-tile-btn' });
    b.classList.add('cs-newbtn');
    b.disabled = !S.rules || !caps().create || busy();
    b.addEventListener('click', () => openCreate(b));
    S.keys['new-tile-btn'] = b;
    li.append(b, el('p', 'cs-sub', T.newTileSub));
    return li;
  }

  function deletedRow(a) {
    const li = el('li', 'cs-delrow');
    li.dataset.key = 'deleted-' + a.id;
    const text = el('div', 'cs-deltext');
    text.append(el('b', '', a.name), el('span', 'cs-sub', delMeta(a)));
    const b = btn(T.restore, { key: 'restore-' + a.id });
    b.setAttribute('aria-label', T.restore + ' ' + a.name);
    b.disabled = !caps().restore || busy();
    b.addEventListener('click', () => openRestore(a, b));
    S.keys[b.dataset.key] = b;
    li.append(text, b);
    return li;
  }
  const delMeta = (a) => {
    const w = when(a.archivedAt, now());
    return a.level ? fmt(T.deletedMeta, { level: a.level, class: className(a.class, a.class), when: w }) : fmt(T.deletedMetaNoLevel, { when: w });
  };

  function panel(titleText, body, retry) {
    const p = el('section', 'cs-state');
    p.setAttribute('role', 'status');
    p.append(el('h2', '', titleText), el('p', '', body));
    if (retry) { const b = btn(T.retry, { variant: 'primary', key: 'retry' }); b.addEventListener('click', () => refresh()); p.appendChild(b); }
    return p;
  }

  function draw() {
    S.keys = {};
    const kids = [];
    const v = S.view;
    if (!v && !S.err) kids.push(panel(T.title, T.loading));
    else if (!v) {
      const e = S.err;
      const m = { off: [T.errOffTitle, T.errOffBody], no_save_dir: [T.errNoSaveTitle, T.errNoSaveBody], noanswer: [T.errNoAnswerTitle, T.errNoAnswerBody] }[e.code];
      kids.push(m ? panel(m[0], m[1], e.code !== 'off') : panel(T.errOtherTitle, e.text || T.noAnswer, true));
    } else {
      nowEl.textContent = v.active ? fmt(T.playingNow, { name: v.active }) : '';
      const sec = el('section', 'cs-sec');
      const h = el('h2', '', T.heroes);
      sec.appendChild(h);
      const c = caps();
      if (!c.switch && c.why) { const w = el('p', 'cs-why', fmt(T.switchOff, { why: c.why })); w.id = 'cs-why'; sec.appendChild(w); }
      if (!(v.characters || []).length) sec.appendChild(panel(T.emptyTitle, T.emptyBody));
      const ul = el('ul', 'cs-grid');
      ul.setAttribute('aria-label', T.heroes);
      for (const ch of v.characters || []) ul.appendChild(card(ch));
      ul.appendChild(newTile());
      sec.appendChild(ul);
      kids.push(sec);
      const del = el('section', 'cs-sec cs-deleted');
      del.appendChild(el('h2', '', T.deleted + ((v.deleted || []).length ? ' (' + v.deleted.length + ')' : '')));
      if (!(v.deleted || []).length) del.appendChild(el('p', 'cs-sub', T.deletedEmpty));
      else { const dl = el('ul', 'cs-dellist'); for (const a of v.deleted) dl.appendChild(deletedRow(a)); del.appendChild(dl); }
      kids.push(del);
    }
    main.replaceChildren(...kids);
    newTop.hidden = !v;
    newTop.disabled = !v || !S.rules || !caps().create || busy();
    S.keys['new-top'] = newTop;
    const st = strip(v, S.pend, now(), T);
    stripEl.hidden = !st.active && !st.failed;
    stripEl.textContent = st.text;
    stripEl.dataset.state = st.failed ? 'failed' : 'busy';
    return st;
  }

  // ---- reading ----
  let timer = null, since = 0;
  async function refresh() {
    const j = await d.get(ROUTE);
    if (j && j.schema === 'watch.characters/1') {
      S.view = j; S.err = null; S.waiting = false;
      if (S.pend && !j.switch && j.active === S.pend.to) { setNote('good', fmt(T.doneSwitch, { name: j.active })); S.pend = null; }
    } else if (S.pend) {
      S.waiting = true; // the bot is restarting: keep the last view, the strip says so
    } else if (j && j.code) S.err = { code: j.code, text: refusalText(j, T) }, S.view = null;
    else S.err = { code: 'noanswer' }, S.view = null;
    const st = draw();
    schedule(st);
    return j;
  }
  function schedule(st) {
    if (timer) return;
    if (st.active && (!since || now() - since < GIVE_UP_MS)) {
      if (!since) since = now();
      timer = d.later(() => { timer = null; refresh(); }, POLL_MS);
    } else if (!st.active) since = 0;
  }
  async function loadRules() {
    const j = await d.get(RULES);
    S.rules = j && j.schema === 'watch.characters.rules/1' ? j : null;
  }

  // ---- the sheet: one modal at a time, Escape and the scrim close it, Tab stays inside ----
  function closeSheet() {
    if (!S.sheet) return;
    const key = S.sheet.opener;
    S.sheet = null;
    sheetWrap.hidden = true;
    sheetCard.replaceChildren();
    const t = (key && S.keys[key]) || (key === 'new-top' ? newTop : null);
    (t && !t.disabled ? t : main).focus();
  }
  function openSheet(o) {
    S.sheet = { opener: o.opener, focus: [] };
    const h = el('h2', '', o.title);
    h.id = 'cs-sheet-title';
    const err = el('p', 'cs-err', '');
    err.hidden = true; err.setAttribute('role', 'alert');
    const row = el('div', 'cs-row');
    const no = btn(o.noLabel || T.cancel, { variant: 'ghost', key: 'sheet-cancel' });
    no.addEventListener('click', closeSheet);
    row.appendChild(no);
    let go = null;
    if (o.goLabel) {
      go = btn(o.goLabel, { variant: o.danger ? 'danger' : 'primary', key: 'sheet-go' });
      go.id = 'cs-go';
      go.disabled = !!o.goOff;
      row.appendChild(go);
    }
    sheetCard.replaceChildren(h, ...(o.body || []), err, row);
    sheetWrap.dataset.kind = o.kind;
    sheetWrap.hidden = false;
    S.sheet.focus = [...(o.fields || []), no, go].filter(Boolean);
    S.sheet.err = err; S.sheet.go = go; S.sheet.no = no;
    (o.first || go || no).focus();
    if (go) go.addEventListener('click', async () => {
      go.disabled = true; no.disabled = true;
      err.hidden = true;
      const out = await o.run();
      no.disabled = false;
      if (out && out.ok) { closeSheet(); return; }
      go.disabled = o.goOff ? o.goOff() : false;
      err.hidden = false; err.textContent = out && out.text ? out.text : T.noAnswer;
      refresh(); // read back: the list shows what really is
    });
    return S.sheet;
  }
  sheetWrap.addEventListener('click', (e) => { if (e.target === sheetWrap) closeSheet(); });
  sheetWrap.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.preventDefault(); closeSheet(); return; }
    if (e.key === 'Tab' && S.sheet) {
      const f = S.sheet.focus.filter((n) => !n.hidden && !n.disabled);
      if (!f.length) return;
      const i = f.indexOf(doc.activeElement), n = e.shiftKey ? (i <= 0 ? f.length - 1 : i - 1) : (i === f.length - 1 ? 0 : i + 1);
      e.preventDefault();
      f[n].focus();
    }
  });
  doc.addEventListener && doc.addEventListener('keydown', (e) => { if (e.key === 'Escape' && S.sheet && !sheetWrap.contains(e.target)) closeSheet(); });

  // a press: post, then read the list back; the result is judged by what the list says
  async function act(route, body, verify, ok) {
    const j = await d.post(route, body);
    if (!j || j.ok === false || j.schema !== 'watch.characters.result/1') {
      refresh();
      return { ok: false, text: refusalText(j, T) };
    }
    await ok(j);
    const v = await refresh();
    const bad = v && v.schema === 'watch.characters/1' ? verify(v, j) : '';
    if (bad) setNote('warn', bad);
    return { ok: true };
  }

  function openSwitch(c, opener) {
    const run = () => act(ROUTE + '/switch', { name: c.name }, () => '', (j) => {
      S.pend = { to: c.name, from: S.view && S.view.active, at: now() };
      setNote('good', fmt(T.startedSwitch, { name: c.name }));
    });
    openSheet({ kind: 'switch', opener: opener.dataset.key, title: fmt(T.switchTitle, { name: c.name }), body: [el('p', '', fmt(T.switchBody, { name: c.name }))], goLabel: fmt(T.switchTo, { name: c.name }), run });
  }

  function openDelete(c, opener) {
    if (c.active) {
      openSheet({ kind: 'delete-active', opener: opener.dataset.key, title: fmt(T.deleteActiveTitle, { name: c.name }), body: [el('p', '', fmt(T.deleteActive, { name: c.name }))], noLabel: T.close });
      return;
    }
    const lab = el('label', 'cs-label', fmt(T.deleteTypeLabel, { name: c.name }));
    const input = doc.createElement('input');
    input.className = 'cs-input'; input.id = 'cs-typed'; input.type = 'text';
    input.setAttribute('autocomplete', 'off'); input.setAttribute('autocapitalize', 'off'); input.setAttribute('spellcheck', 'false');
    lab.htmlFor = 'cs-typed';
    const body = [el('p', '', fmt(T.deleteBody1, { name: c.name })), el('p', '', fmt(T.deleteBody2, { name: c.name })), lab, input];
    const sh = openSheet({
      kind: 'delete', opener: opener.dataset.key, title: fmt(T.deleteTitle, { name: c.name }), body, fields: [input], first: input, danger: true,
      goLabel: fmt(T.deleteGo, { name: c.name }), goOff: () => input.value !== c.name,
      run: () => act(ROUTE + '/delete', { name: c.name, confirm: input.value }, (v) => ((v.characters || []).some((x) => x.name === c.name) ? fmt(T.gone, { name: c.name }) : ''),
        (j) => setNote('good', fmt(T.doneDelete, { name: c.name, archive: j.archive || '' }))),
    });
    const sync = () => { sh.go.disabled = input.value !== c.name; };
    input.addEventListener('input', sync);
    input.addEventListener('keydown', (e) => { if (e.key === 'Enter' && !sh.go.disabled) { e.preventDefault(); sh.go.click(); } });
  }

  function openRestore(a, opener) {
    openSheet({
      kind: 'restore', opener: opener.dataset.key, title: fmt(T.restoreTitle, { name: a.name }), body: [el('p', '', fmt(T.restoreBody, { name: a.name, meta: delMeta(a) }))], goLabel: fmt(T.restoreGo, { name: a.name }),
      run: () => act(ROUTE + '/restore', { archive: a.id }, (v) => ((v.characters || []).some((x) => x.name === a.name) ? '' : fmt(T.missing, { name: a.name })),
        () => setNote('good', fmt(T.doneRestore, { name: a.name }))),
    });
  }

  function openCreate(opener) {
    const c = caps(), r = S.rules;
    if (!c.create || !r) {
      const why = !r ? T.rulesMissing : fmt(T.createOff, { why: c.why || '' });
      openSheet({ kind: 'create-off', opener: opener.dataset.key, title: T.createTitle, body: [el('p', '', why)], noLabel: T.close });
      return;
    }
    const lab = el('label', 'cs-label', T.nameLabel);
    const input = doc.createElement('input');
    input.className = 'cs-input'; input.id = 'cs-name'; input.type = 'text'; lab.htmlFor = 'cs-name';
    input.setAttribute('autocomplete', 'off'); input.setAttribute('autocapitalize', 'off'); input.setAttribute('spellcheck', 'false');
    input.setAttribute('aria-describedby', 'cs-hint');
    const hint = el('p', 'cs-hint', '');
    hint.id = 'cs-hint'; hint.setAttribute('role', 'status');
    const fs = el('fieldset', 'cs-classes');
    fs.appendChild(el('legend', '', T.classLabel));
    const radios = [];
    let picked = '';
    for (const k of r.classes || []) {
      const lb = el('label', 'cs-chipbtn');
      lb.dataset.class = k.id;
      const rd = doc.createElement('input');
      rd.type = 'radio'; rd.name = 'cs-class'; rd.value = k.id;
      rd.addEventListener('change', () => { picked = k.id; sync(); });
      radios.push(rd);
      lb.append(rd, el('span', '', k.name));
      fs.appendChild(lb);
    }
    const hc = el('p', 'cs-sub', r.hardcore ? '' : T.hardcoreNote);
    hc.hidden = !!r.hardcore;
    const body = [lab, input, hint, fs, hc];
    const ok = () => checkName(input.value, r, S.view && S.view.characters, T).ok && !!picked;
    const sh = openSheet({
      kind: 'create', opener: opener.dataset.key, title: T.createTitle, body, fields: [input, ...radios], first: input, goLabel: T.createGoBlank, goOff: () => !ok(),
      run: () => act(ROUTE, { name: input.value, class: picked }, () => '', () => {
        S.pend = { to: input.value, from: S.view && S.view.active, at: now(), creating: true };
        setNote('good', fmt(T.startedCreate, { name: input.value }));
      }),
    });
    function sync() {
      const n = checkName(input.value, r, S.view && S.view.characters, T);
      hint.textContent = n.ok && !picked ? T.classPick : n.text;
      hint.dataset.level = n.ok && !picked ? 'hint' : n.level;
      input.setAttribute('aria-invalid', String(n.level === 'bad'));
      sh.go.disabled = !ok();
      sh.go.textContent = n.ok ? fmt(T.createGo, { name: input.value }) : T.createGoBlank;
    }
    input.addEventListener('input', sync);
    input.addEventListener('keydown', (e) => { if (e.key === 'Enter' && !sh.go.disabled) { e.preventDefault(); sh.go.click(); } });
    sync();
  }
  newTop.addEventListener('click', () => openCreate(newTop));

  if (d.overlays) d.overlays.register('characters-sheet', sheetWrap, closeSheet); // the page's one modal: Escape and the one-at-a-time rule are the manager's (docs/ui-overlays.md)

  async function start() {
    draw();
    await Promise.all([loadRules(), refresh()]);
    draw();
    return S;
  }
  return { start, refresh, draw, state: S, openCreate, closeSheet, main };
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
  const texts = await get('/watch/characters.json');
  const root = doc.getElementById('cs-root');
  if (!texts || !root) { if (root) root.textContent = 'The Characters page could not load its texts.'; return; }
  let overlays = null;
  try { overlays = (await import('/watch/overlay.js')).overlays; } catch (e) { /* the sheet closes itself on Escape */ }
  mountCharacters({ doc, root, overlays, get, post, later: (f, ms) => setTimeout(f, ms), texts, now: Date.now }).start();
}
if (typeof document !== 'undefined' && document.body && document.body.dataset && document.body.dataset.chars !== undefined) boot();
