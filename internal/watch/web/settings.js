// The Settings sheet (card uisettings, review docs/design/fable-review-2026-11-22/ui-unification.md C2/E10): the two controls that used
// to live only on the pre-shell page, in the shared sheet component (sheet.js): 'Rules' (revoke one learned rule or all of them:
// GET /api/rules, POST /api/revoke and /api/revoke-all on the answer route) and 'Alerts' (sound on a new alert, desktop notifications:
// the same device settings alerts.js keeps). More (nav.js) opens it on every page. Every control has a visible label; a control that is
// off names why in a visible line beside it (why.js). Text goes through textContent only.
// Classic script: window.dSettings, or module.exports for the node test (internal/watch/testdata/settings_test.cjs).
(function (g) {
  'use strict';
  const SOUND_KEY = 'd2lod.alerts.sound'; // the key alerts.js of the old page reads: one setting for both pages

  function mk(doc, tag, cls, text) {
    const e = doc.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }

  // reasonRow: a control and the line that says why it is off (always in the page, empty text hides it), so the pair stays together
  function row(doc, ctl, id) {
    const line = mk(doc, 'p', 'd-set-why');
    line.id = id;
    ctl.setAttribute('aria-describedby', id);
    const r = mk(doc, 'div', 'd-set-row');
    r.append(ctl, line);
    return { row: r, line };
  }
  const btn = (doc, text, variant, key) => {
    const b = mk(doc, 'button', 'd-btn', text);
    b.type = 'button'; b.dataset.size = 'md'; b.dataset.variant = variant; b.dataset.key = key;
    return b;
  };

  // notifyState: the desktop-notification line and whether its button is usable, from the browser's permission
  function notifyState(N) {
    if (!N) return { on: false, text: 'This browser has no desktop notifications.' };
    if (N.permission === 'granted') return { on: false, text: 'Desktop notifications are on while this tab is open. Closing the tab stops them (no background push).' };
    if (N.permission === 'denied') return { on: false, text: 'Desktop notifications are blocked in the browser settings for this site.' };
    return { on: true, text: 'Works while this tab is open. The browser asks only when you press the button. Background push is not built.' };
  }

  // open(env): env = { doc, sheet (dSheet), toast (dToast), fetch, base(): Promise<answer origin>, store: {get,set}, notif (Notification or null), opener, section }
  function open(env) {
    const doc = env.doc;
    // the sheet's scrim sits above the page's toast corner, so the answers show in a stack inside the sheet (same toasts, same severities)
    const stack = mk(doc, 'div', 'd-set-toasts'); stack.setAttribute('role', 'status'); stack.setAttribute('aria-live', 'polite');
    const toasts = env.toast.create ? env.toast.create(doc, { host: stack }) : env.toast;
    const toast = (s, t) => toasts.show(s, t);
    const id = 'd-set';
    const rules = mk(doc, 'section', 'd-set-sec'); rules.dataset.section = 'rules'; rules.setAttribute('aria-labelledby', id + '-rt');
    const alerts = mk(doc, 'section', 'd-set-sec'); alerts.dataset.section = 'alerts'; alerts.setAttribute('aria-labelledby', id + '-at');
    const rt = mk(doc, 'h3', 'd-set-h', 'Rules'); rt.id = id + '-rt';
    const at = mk(doc, 'h3', 'd-set-h', 'Alerts'); at.id = id + '-at';
    const hint = mk(doc, 'p', 'd-set-hint', 'Every rule the hero learned from you. A rule never revokes itself.');
    const list = mk(doc, 'ul', 'd-set-list');
    const all = btn(doc, 'Revoke all', 'danger', 'revoke-all');
    const allRow = row(doc, all, id + '-allwhy');
    // the Loot rules sheet lives on the live screen (the pre-Settings Inventory page had its entry): one link reaches it from every page
    const loot = mk(doc, 'a', 'd-btn', 'Loot rules...');
    loot.href = '/watch?live=1&lootrules=1'; loot.dataset.size = 'md'; loot.dataset.variant = 'ghost'; loot.dataset.key = 'lootrules-link';
    loot.setAttribute('aria-describedby', id + '-lootwhy');
    const lootWhy = mk(doc, 'p', 'd-set-hint', 'What the bot picks up, stashes and sells. Opens on the live screen.'); lootWhy.id = id + '-lootwhy';
    const lootRow = mk(doc, 'div', 'd-set-row'); lootRow.append(loot, lootWhy);
    rules.append(rt, lootRow, hint, list, allRow.row);

    // ---- Rules ----
    let busy = false, live = 0;
    const origin = () => Promise.resolve(env.base()).then((b) => b || '');
    const post = (path, body) => origin().then((b) => env.fetch(b + path, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }))
      .then((r) => r.json().then((j) => ({ ok: !!j.ok, error: j.error || ('error ' + r.status) })))
      .catch(() => ({ ok: false, error: 'The answer route is not reachable.' }));
    function paint(rs, failed) {
      list.replaceChildren();
      live = failed ? 0 : rs.live || 0;
      const items = failed ? [] : rs.rules || [];
      if (failed) list.append(mk(doc, 'li', 'd-set-empty', 'The rules could not be loaded.'));
      else if (!items.length) list.append(mk(doc, 'li', 'd-set-empty', 'No rule learned yet.'));
      for (const x of items) {
        const li = mk(doc, 'li', 'd-set-rule');
        li.dataset.key = x.key;
        const t = mk(doc, 'span', 'd-set-rt', x.text);
        const meta = mk(doc, 'small', 'd-set-meta', '[' + x.risk + ']' + (x.revoked ? ' revoked' : ''));
        li.append(t, meta);
        if (!x.revoked) {
          const b = btn(doc, 'Revoke', 'danger', 'revoke');
          b.dataset.size = 'sm';
          b.setAttribute('aria-label', 'Revoke this rule: ' + x.text);
          b.addEventListener('click', () => revoke(x.key));
          li.append(b);
        }
        list.append(li);
      }
      all.disabled = live === 0;
      allRow.line.textContent = failed ? 'The rules could not be loaded, so there is nothing to revoke.' : live === 0 ? 'No rule is active, so there is nothing to revoke.' : '';
      allRow.line.hidden = !allRow.line.textContent;
    }
    function load() {
      return env.fetch('/api/rules', { cache: 'no-store', credentials: 'same-origin' }).then((r) => { if (!r.ok) throw new Error(String(r.status)); return r.json(); })
        .then((v) => paint(v, false), () => { paint(null, true); toast('bad', 'The rules could not be loaded.'); });
    }
    function done(res, okText) {
      busy = false;
      if (res.ok) toast('good', okText); else toast('bad', res.error);
      return load();
    }
    function revoke(key) {
      if (busy) return Promise.resolve();
      busy = true;
      return post('/api/revoke', { key }).then((r) => done(r, 'Rule revoked.'));
    }
    all.addEventListener('click', () => {
      if (busy) return;
      env.sheet.ask(doc, { title: 'Revoke every rule?', lines: ['The hero will ask again.'], confirmText: 'Revoke all', danger: true, opener: all }).then((yes) => {
        if (!yes) return;
        busy = true;
        post('/api/revoke-all', { confirm: true }).then((r) => done(r, 'Every rule revoked.'));
      });
    });

    // ---- Alerts ----
    const sound = mk(doc, 'input'); sound.type = 'checkbox'; sound.id = id + '-sound'; sound.dataset.key = 'alert-sound';
    sound.checked = env.store.get(SOUND_KEY) === '1';
    sound.addEventListener('change', () => env.store.set(SOUND_KEY, sound.checked ? '1' : '0'));
    const soundLab = mk(doc, 'label', 'd-set-check'); soundLab.setAttribute('for', sound.id);
    soundLab.append(sound, mk(doc, 'span', '', 'Play a sound on a new alert'));
    const note = btn(doc, 'Turn on desktop notifications', 'primary', 'alert-notify');
    const noteRow = row(doc, note, id + '-notewhy');
    function paintNote() {
      const s = notifyState(env.notif);
      note.disabled = !s.on;
      noteRow.line.textContent = s.text;
    }
    note.addEventListener('click', () => {
      const p = env.notif && env.notif.requestPermission();
      Promise.resolve(p).then(paintNote, paintNote);
    });
    paintNote();
    alerts.append(at, mk(doc, 'p', 'd-set-hint', 'Kept on this device. They apply to the alert feed while this tab is open.'), soundLab, noteRow.row);

    const h = env.sheet.open(doc, { title: 'Settings', panel: true, nodes: [stack, rules, alerts], confirmText: 'Close', opener: env.opener, parent: env.parent });
    h.sheet.className += ' d-set';
    const ready = load().then(() => {
      const want = env.section === 'alerts' ? alerts : env.section === 'rules' ? rules : null;
      if (want && want.scrollIntoView) want.scrollIntoView({ block: 'start' });
    });
    return Object.assign(h, { ready, reload: load });
  }

  const api = { open, notifyState, SOUND_KEY };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else g.dSettings = api;
})(typeof window !== 'undefined' ? window : globalThis);
