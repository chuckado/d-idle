// The one confirm sheet (card uisheet, docs/ui-design-system.md "Sheets and confirms"). Every "are you sure" on the web pages goes
// through here; window.confirm, alert and prompt are not used anywhere (sheet_noprompt_test.go greps for them).
// o = { title, lines: [text], variants: [{name,label,tier}], advanced: {summary, variants: [{name,label,why,recommended,warn}]}, pick (name), confirmText, cancelText, danger, notice, onConfirm(pick), onCancel, opener, parent }.
// panel: a titled dialog that carries nodes of its own (o.nodes, appended after the lines) and one Close button (confirmText, default "Close"): the Settings sheet.
// notice: a one-button message (the replacement for alert). danger: the confirm button is the red one and the dialog is an alertdialog.
// Focus goes into the sheet, Tab stays inside, Escape, a press on the backdrop or Cancel closes it and returns the focus to the opener.
// advanced: a collapsed button that, when pressed, shows the variants as radios (none are rendered while collapsed); the radios
// carry the plain label, the one-line why, a "recommended" marker and, when warn is set, the warning text. Variant names (plan ids) are values and
// never visible text. Text goes through textContent only. Classic script: window.dSheet, or module.exports for the node test.
(function (g) {
  'use strict';
  let seq = 0;
  function mk(doc, tag, cls, text) {
    const e = doc.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }
  // the controls Tab may land on inside the sheet (a real DOM or the node test's stub)
  const focusables = (sh) => (sh.querySelectorAll ? [...sh.querySelectorAll('button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),a[href]')]
    : sh.all().filter((n) => !n.disabled && (n.tag === 'button' || n.tag === 'input' || n.tag === 'select' || n.tag === 'textarea' || (n.tag === 'a' && n.href))));

  function open(doc, o) {
    const id = 'd-sheet-' + (++seq);
    const back = mk(doc, 'div', 'd-sheet-back');
    const sh = mk(doc, 'div', 'd-sheet');
    sh.setAttribute('role', (o.danger || o.notice) && !o.panel ? 'alertdialog' : 'dialog');
    sh.setAttribute('aria-modal', 'true');
    const title = mk(doc, 'h2', 'd-sheet-title', o.title);
    title.id = id + '-t';
    title.setAttribute('tabindex', '-1');
    sh.setAttribute('aria-labelledby', title.id);
    sh.append(title);
    const lines = o.lines || [];
    lines.forEach((t, i) => { const p = mk(doc, 'p', 'd-sheet-line', t); if (!i) { p.id = id + '-d'; sh.setAttribute('aria-describedby', p.id); } sh.append(p); });
    if (o.nodes) sh.append(...o.nodes);
    let pick = o.pick || (o.variants && o.variants[0] && o.variants[0].name) || '';
    if (o.variants && o.variants.length > 1) {
      const fs = mk(doc, 'div', 'd-sheet-pick');
      fs.setAttribute('role', 'radiogroup');
      fs.setAttribute('aria-label', 'Variant');
      for (const v of o.variants) {
        const lab = mk(doc, 'label', 'd-sheet-opt');
        const r = mk(doc, 'input', '');
        r.type = 'radio';
        r.name = id + '-variant';
        r.value = v.name;
        r.checked = v.name === pick;
        r.addEventListener('change', () => { pick = v.name; });
        lab.append(r, mk(doc, 'span', '', ' ' + v.label + (v.tier ? ' (' + v.tier + ')' : '')));
        fs.append(lab);
      }
      sh.append(fs);
    }
    if (o.advanced && o.advanced.variants && o.advanced.variants.length) {
      const adv = o.advanced, wrap = mk(doc, 'div', 'd-sheet-adv');
      const tog = mk(doc, 'button', 'd-sheet-adv-btn', adv.summary || 'Advanced');
      tog.type = 'button';
      tog.setAttribute('aria-expanded', 'false');
      tog.setAttribute('aria-controls', id + '-adv');
      wrap.append(tog);
      let grp = null;
      tog.addEventListener('click', () => {
        if (grp) { if (grp.remove) grp.remove(); grp = null; tog.setAttribute('aria-expanded', 'false'); return; }
        grp = mk(doc, 'div', 'd-sheet-pick');
        grp.id = id + '-adv';
        grp.setAttribute('role', 'radiogroup');
        grp.setAttribute('aria-label', 'Plan');
        for (const v of adv.variants) {
          const lab = mk(doc, 'label', 'd-sheet-opt');
          lab.dataset.plan = v.name;
          const r = mk(doc, 'input', '');
          r.type = 'radio';
          r.name = id + '-plan';
          r.value = v.name;
          r.checked = v.name === pick;
          r.addEventListener('change', () => { pick = v.name; });
          const txt = mk(doc, 'span', 'd-sheet-opt-text');
          txt.append(mk(doc, 'span', 'd-sheet-opt-name', v.label + (v.recommended ? ' (recommended)' : '')));
          if (v.why) txt.append(mk(doc, 'span', 'd-sheet-opt-why', v.why));
          if (v.warn) txt.append(mk(doc, 'span', 'd-sheet-opt-warn', v.warn));
          lab.append(r, txt);
          grp.append(lab);
        }
        wrap.append(grp);
        tog.setAttribute('aria-expanded', 'true');
      });
      sh.append(wrap);
    }
    const row = mk(doc, 'div', 'd-sheet-row');
    const cancel = mk(doc, 'button', 'd-btn d-sheet-cancel', o.cancelText || 'Cancel');
    cancel.type = 'button';
    const ok = mk(doc, 'button', 'd-btn d-sheet-ok', o.confirmText || (o.panel ? 'Close' : o.notice ? 'OK' : 'Confirm'));
    ok.type = 'button';
    ok.setAttribute('data-variant', o.danger ? 'danger' : 'primary');
    if (!o.notice && !o.panel) row.append(cancel);
    row.append(ok);
    sh.append(row);
    back.append(sh);
    let closed = false;
    function close(restore) {
      if (closed) return;
      closed = true;
      if (back.remove) back.remove();
      if (restore !== false && o.opener && o.opener.focus) o.opener.focus();
    }
    function dismiss() { if (closed) return; close(); if (o.onCancel) o.onCancel(); }
    cancel.addEventListener('click', dismiss);
    ok.addEventListener('click', () => { const p = pick; close(o.notice || o.panel ? true : false); if (o.onConfirm) o.onConfirm(p); });
    back.addEventListener('click', (e) => { if (e.target === back) dismiss(); });
    back.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') { e.preventDefault(); dismiss(); return; }
      if (e.key !== 'Tab') return;
      const fs = focusables(sh);
      if (!fs.length) return;
      const first = fs[0], last = fs[fs.length - 1], at = doc.activeElement;
      if (e.shiftKey && (at === first || at === title)) { e.preventDefault(); last.focus(); }
      else if (!e.shiftKey && at === last) { e.preventDefault(); first.focus(); }
    });
    (o.parent || doc.body).append(back);
    title.focus();
    return { close, node: back, sheet: sh, confirm: ok, cancel };
  }

  // a Promise form for call sites that used window.confirm: resolves true on confirm, false on cancel
  function ask(doc, o) {
    return new Promise((resolve) => { open(doc, Object.assign({}, o, { onConfirm: () => resolve(true), onCancel: () => resolve(false) })); });
  }
  function notice(doc, o) {
    return new Promise((resolve) => { open(doc, Object.assign({}, o, { notice: true, onConfirm: () => resolve(), onCancel: () => resolve() })); });
  }

  const api = { open, ask, notice };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else g.dSheet = api;
})(typeof window !== 'undefined' ? window : globalThis);
