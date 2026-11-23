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

  // The build confirm: the shared sheet (web/sheet.js) with the variant picker. Kept as openSheet for points.js and the test.
  function openSheet(doc, o) {
    const S = typeof module !== 'undefined' && module.exports ? require('./sheet.cjs') : g.dSheet;
    return S.open(doc, Object.assign({ confirmText: 'Choose and lock' }, o));
  }

  const api = { stars, ratingRow, sourceWord, cardNode, detailNode, variantNode, openSheet };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else g.dBuildCards = api;
})(typeof window !== 'undefined' ? window : globalThis);
