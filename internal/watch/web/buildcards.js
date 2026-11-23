// Build cards (card buildcards, review E6): the pieces of the Build tab as pure DOM builders, so a node test can look at them.
// A card = name, identity line, four ratings (each with a text equivalent), a Details button. The detail view adds the
// summary, pros and cons, all seven ratings, the variants (the plan tables behind the build) and Choose. One confirm
// sheet, never window.confirm. Text goes through textContent only. Classic script: window.dBuildCards, or module.exports for node.
(function (g) {
  'use strict';

  const stars = (n) => '★'.repeat(n) + '☆'.repeat(5 - n);

  function mk(doc, tag, cls, text) {
    const e = doc.createElement(tag);
    if (cls) e.className = cls;
    if (text !== undefined) e.textContent = text;
    return e;
  }

  // One rating: visible "Defense ★★★★☆", the stars carry the text equivalent as aria-label ("Defense: 4 of 5 stars, estimate").
  function ratingRow(doc, r, withNote) {
    const li = mk(doc, 'li', 'bc-r');
    li.dataset.rating = r.key;
    li.append(mk(doc, 'span', 'bc-rl', r.label + ' '));
    const st = mk(doc, 'span', 'bc-stars', stars(r.stars));
    st.setAttribute('role', 'img');
    st.setAttribute('aria-label', r.text);
    li.append(st);
    if (withNote) {
      li.append(mk(doc, 'span', 'bc-rs', ' ' + r.source));
      if (r.note) li.append(mk(doc, 'span', 'bc-rn', ' - ' + r.note));
    }
    return li;
  }

  // "Estimate" when every shown rating is an opinion, "Measured" when all are measured, else both words: never an unlabelled star.
  function sourceWord(ratings) {
    const set = new Set(ratings.map((r) => r.source));
    if (set.size === 1) return set.has('measured') ? 'Measured' : 'Estimate';
    return 'Estimate and measured';
  }

  function list(doc, items, cls) {
    const ul = mk(doc, 'ul', cls);
    for (const t of items) ul.append(mk(doc, 'li', '', t));
    return ul;
  }

  // The small card. o.onDetails(card, button) is called by the Details button.
  function cardNode(doc, card, o) {
    o = o || {};
    const li = mk(doc, 'li', 'bc-card');
    li.dataset.build = card.id;
    li.dataset.open = o.open ? '1' : '0';
    const head = mk(doc, 'h3', 'bc-name', card.name);
    head.id = 'bc-h-' + card.id;
    li.setAttribute('aria-labelledby', head.id);
    li.append(head, mk(doc, 'p', 'bc-id', card.identity));
    const rl = mk(doc, 'ul', 'bc-rates');
    rl.setAttribute('aria-label', 'Ratings');
    for (const r of card.ratings) rl.append(ratingRow(doc, r, false));
    li.append(rl, mk(doc, 'p', 'bc-src', sourceWord(card.ratings)));
    const btn = mk(doc, 'button', 'd-btn bc-details', o.open ? 'Hide details' : 'Details');
    btn.type = 'button';
    btn.setAttribute('aria-expanded', o.open ? 'true' : 'false');
    btn.setAttribute('aria-label', (o.open ? 'Hide details of ' : 'Details of ') + card.name);
    btn.addEventListener('click', () => { if (o.onDetails) o.onDetails(card, btn); });
    li.append(btn);
    return li;
  }

  function variantNode(doc, v) {
    const li = mk(doc, 'li', 'bc-var');
    li.dataset.table = v.name;
    li.dataset.tier = v.tier;
    const h = mk(doc, 'p', 'bc-vn', v.label + ' ');
    h.append(mk(doc, 'span', 'bc-tier', '(' + v.tier + ')'));
    li.append(h);
    const bad = (v.needs || []).filter((n) => n.tier === 'unsupported');
    const rest = (v.needs || []).length - bad.length;
    if (bad.length) li.append(mk(doc, 'p', 'bc-warn', 'The hero cannot use these yet and will hold the point for you: ' + bad.map((n) => n.skill).join(', ') + '.'));
    if (rest) li.append(mk(doc, 'p', 'bc-note', rest + ' more skill' + (rest === 1 ? ' is' : 's are') + ' not proven in the fighter yet.'));
    if ((v.preview || []).length) {
      const d = mk(doc, 'details', 'bc-prev');
      d.append(mk(doc, 'summary', '', 'First ' + v.preview.length + ' levels'));
      const ul = mk(doc, 'ul', 'bc-prev-list');
      for (const r of v.preview) {
        const parts = (r.skills || []).map((s) => s.skill + (s.points > 1 ? ' x' + s.points : '') + (s.tier && s.tier !== 'supported' ? ' [' + s.tier + ']' : ''));
        const st = Object.keys(r.stats || {}).map((k) => '+' + r.stats[k] + ' ' + k);
        ul.append(mk(doc, 'li', '', 'Level ' + r.clvl + ': ' + parts.concat(st).join(', ')));
      }
      d.append(ul);
      li.append(d);
    }
    return li;
  }

  // The detail view. st = { autoAlloc, locked, chosen (the table the hero is locked to), label }, o.onChoose(detail, button).
  function detailNode(doc, det, st, o) {
    o = o || {};
    const box = mk(doc, 'div', 'bc-detail');
    box.dataset.build = det.id;
    box.append(mk(doc, 'p', 'bc-sum', det.summary));
    const pc = mk(doc, 'div', 'bc-pc');
    const pros = mk(doc, 'div', 'bc-pros'), cons = mk(doc, 'div', 'bc-cons');
    pros.append(mk(doc, 'h4', '', 'Pros'), list(doc, det.pros, 'bc-list-pros'));
    cons.append(mk(doc, 'h4', '', 'Cons'), list(doc, det.cons, 'bc-list-cons'));
    pc.append(pros, cons);
    box.append(pc);
    const rh = mk(doc, 'h4', '', 'Ratings (' + sourceWord(det.ratings).toLowerCase() + ')');
    box.append(rh);
    const rl = mk(doc, 'ul', 'bc-rates bc-rates-all');
    rl.setAttribute('aria-label', 'All ratings');
    for (const r of det.ratings) rl.append(ratingRow(doc, r, true));
    box.append(rl);
    if (det.defenseByDifficulty) {
      const d = det.defenseByDifficulty;
      box.append(mk(doc, 'p', 'bc-diff', 'Defense by difficulty: Normal ' + stars(d.normal) + ' (' + d.normal + ' of 5), Hell ' + stars(d.hell) + ' (' + d.hell + ' of 5).'));
    }
    box.append(mk(doc, 'h4', '', 'Variants'));
    const vl = mk(doc, 'ul', 'bc-vars');
    for (const v of det.variants) vl.append(variantNode(doc, v));
    if (!det.variants.length) vl.append(mk(doc, 'li', 'bc-var', 'No plan table of this build is available to the hero yet.'));
    box.append(vl);
    const btn = mk(doc, 'button', 'd-btn bc-choose', 'Choose this build');
    btn.type = 'button';
    btn.setAttribute('data-variant', 'primary');
    const mine = det.variants.some((v) => v.name === st.chosen);
    let why = '';
    if (mine) { btn.textContent = 'Chosen'; why = 'This is your build.'; }
    else if (st.locked) why = 'Locked to ' + (st.label || st.chosen) + '. Respec to change.';
    else if (!st.autoAlloc) why = 'Auto-allocation is off';
    else if (!det.variants.length) why = 'No plan table of this build is available to the hero yet.';
    btn.disabled = !!why;
    btn.addEventListener('click', () => { if (!btn.disabled && o.onChoose) o.onChoose(det, btn); });
    box.append(btn);
    if (why) { const p = mk(doc, 'p', 'bc-why', why); p.id = 'bc-why-' + det.id; btn.setAttribute('aria-describedby', p.id); box.append(p); }
    return box;
  }

  // the controls Tab may land on inside the sheet (a real DOM or the node test's stub)
  const focusables = (sh) => (sh.querySelectorAll ? [...sh.querySelectorAll('button:not([disabled]),input:not([disabled])')]
    : sh.all().filter((n) => !n.disabled && (n.tag === 'button' || n.tag === 'input')));

  // The one confirm sheet. o = { title, lines: [text], variants: [{name,label,tier}], pick (name), confirmText, onConfirm(name), opener, parent }.
  // Focus goes into it, Tab stays inside, Escape or a press on the backdrop or Cancel closes it and puts the focus back on the opener.
  function openSheet(doc, o) {
    const back = mk(doc, 'div', 'bc-back');
    const sh = mk(doc, 'div', 'bc-sheet');
    sh.setAttribute('role', 'dialog');
    sh.setAttribute('aria-modal', 'true');
    const title = mk(doc, 'h2', 'bc-st', o.title);
    title.id = 'bc-sheet-title';
    title.setAttribute('tabindex', '-1');
    sh.setAttribute('aria-labelledby', title.id);
    sh.append(title);
    for (const t of o.lines || []) sh.append(mk(doc, 'p', 'bc-sl', t));
    let pick = o.pick || (o.variants && o.variants[0] && o.variants[0].name) || '';
    if (o.variants && o.variants.length > 1) {
      const fs = mk(doc, 'div', 'bc-pick');
      fs.setAttribute('role', 'radiogroup');
      fs.setAttribute('aria-label', 'Variant');
      for (const v of o.variants) {
        const lab = mk(doc, 'label', 'bc-opt');
        const r = mk(doc, 'input', '');
        r.type = 'radio';
        r.name = 'bc-variant';
        r.value = v.name;
        r.checked = v.name === pick;
        r.addEventListener('change', () => { pick = v.name; });
        lab.append(r, mk(doc, 'span', '', ' ' + v.label + ' (' + v.tier + ')'));
        fs.append(lab);
      }
      sh.append(fs);
    }
    const row = mk(doc, 'div', 'bc-row');
    const cancel = mk(doc, 'button', 'd-btn bc-cancel', 'Cancel');
    cancel.type = 'button';
    const ok = mk(doc, 'button', 'd-btn bc-confirm', o.confirmText || 'Choose and lock');
    ok.type = 'button';
    ok.setAttribute('data-variant', 'primary');
    row.append(cancel, ok);
    sh.append(row);
    back.append(sh);
    let closed = false;
    function close(restore) {
      if (closed) return;
      closed = true;
      if (back.remove) back.remove();
      if (restore !== false && o.opener && o.opener.focus) o.opener.focus();
    }
    cancel.addEventListener('click', () => close());
    ok.addEventListener('click', () => { const p = pick; close(false); if (o.onConfirm) o.onConfirm(p); });
    back.addEventListener('click', (e) => { if (e.target === back) close(); });
    back.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') { e.preventDefault(); close(); return; }
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

  const api = { stars, ratingRow, sourceWord, cardNode, detailNode, variantNode, openSheet };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else g.dBuildCards = api;
})(typeof window !== 'undefined' ? window : globalThis);
