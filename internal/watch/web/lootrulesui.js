// The Loot Rules sheet (docs/lootrules.md, wave 2a): opened by the "..." menu (the d2:lootrules-ask event). It shows the rule
// set in plain language, lets you switch rules off and move them (Custom), or pick the Starter preset and turn four knobs
// (gear to sell, lowest rune, socketed bases, smallest gold pile), and lists the last decisions with the rule behind each one
// ("why was this kept, sold or left"). Saving is one PUT of the whole set; the server checks it and names the line of an error.
// textContent only; every string from the server is data. One overlay at a time (web/overlay.js, docs/ui-overlays.md).
export const ROUTE = '/api/lootrules';
export const ASK = 'd2:lootrules-ask';
export const SCHEMA = 'watch.lootrules/1';
export const STAGES = ['Today\'s code decides every item. The rules below are only compared with it.', 'The rules decide what to pick up. Today\'s code still decides what to stash, sell and drop.', 'The rules decide what to pick up and what to stash or drop. Selling is still today\'s code.'];
const SCOPE = { ground: 'on the ground', tidy: 'in the pack', sell: 'at the vendor' };
const DID = { pickup: 'picked up', ignore: 'left alone', keep: 'kept', stash: 'stashed', sell: 'sold', drop: 'dropped' };
const GEAR = [['magic', 'Magic and better'], ['rare', 'Rare and better'], ['never', 'Never']];

export function stageText(n) { return STAGES[Math.max(0, Math.min(2, n | 0))]; }

// refusal is the words of a refused call (the route's own code, never a status code).
export function refusal(r) {
  if (!r || !r.j) return 'no answer from the hero';
  const j = r.j;
  if (j.code === 'off') return 'Loot rules are not available on this hero yet';
  if (j.code === 'conflict') return 'The rules changed somewhere else. Reload and apply your change again.';
  if (j.code === 'invalid' && j.errors && j.errors.length) return j.errors.slice(0, 5).map((e) => (e.line ? 'line ' + e.line + ': ' : e.path ? e.path + ': ' : '') + e.msg).join('; ');
  return j.error || 'could not save the rules';
}

export function decisionLine(r) {
  const act = DID[r.rules] || r.rules;
  const diff = r.today && r.today !== r.rules ? ' (today\'s code: ' + (DID[r.today] || r.today) + ')' : '';
  return r.item + ' ' + (SCOPE[r.scope] || r.scope) + ': ' + act + diff + '. ' + (r.why || '') + ' [' + r.rule + ']';
}

export function mountLootRules(d) {
  const doc = d.doc || document;
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const wrap = el('div', 'd-lr-sheet');
  wrap.id = 'lootrules-sheet';
  wrap.hidden = true;
  wrap.setAttribute('role', 'dialog');
  wrap.setAttribute('aria-modal', 'true');
  wrap.setAttribute('aria-labelledby', 'lootrules-title');
  const card = el('div', 'd-lr-card');
  const title = el('h2', '', 'Loot rules');
  title.id = 'lootrules-title';
  const stage = el('p', 'd-lr-note', '');
  stage.id = 'lootrules-stage';
  const err = el('p', 'd-lr-err', '');
  err.hidden = true;
  err.setAttribute('role', 'alert');

  const modes = el('div', 'd-lr-modes');
  const bStarter = d.ui.btn('Starter', { size: 'sm' });
  bStarter.dataset.key = 'lr-mode-starter';
  const bCustom = d.ui.btn('Custom', { size: 'sm' });
  bCustom.dataset.key = 'lr-mode-custom';
  modes.append(bStarter, bCustom);

  // Starter: a preset and four knobs
  const simple = el('div', 'd-lr-simple');
  const field = (label, input) => { const l = el('label', 'd-lr-field'); l.append(el('span', '', label), input); return l; };
  const preset = el('select'); preset.id = 'lr-preset'; preset.dataset.key = 'lr-preset';
  const gear = el('select'); gear.id = 'lr-gear'; gear.dataset.key = 'lr-gear';
  for (const [v, t] of GEAR) { const o = el('option', '', t); o.value = v; gear.appendChild(o); }
  const rune = el('input'); rune.type = 'number'; rune.min = '1'; rune.max = '33'; rune.id = 'lr-rune'; rune.dataset.key = 'lr-rune';
  const gold = el('input'); gold.type = 'number'; gold.min = '0'; gold.max = '100000'; gold.id = 'lr-gold'; gold.dataset.key = 'lr-gold';
  const sock = el('input'); sock.type = 'checkbox'; sock.id = 'lr-sock'; sock.dataset.key = 'lr-sock';
  const sockL = el('label', 'd-lr-check'); sockL.append(sock, el('span', '', 'Keep socketed bases (runeword bases)'));
  const apply = d.ui.btn('Apply', { variant: 'primary' });
  apply.id = 'lr-apply';
  apply.dataset.key = 'lr-apply';
  simple.append(field('Preset', preset), field('Pick up gear only to sell it', gear), field('Keep runes from number (1 = every rune)', rune),
    field('Smallest gold pile to pick up (0 = every pile)', gold), sockL, apply);

  // Custom: the whole list
  const custom = el('div', 'd-lr-custom');
  const customNote = el('p', 'd-lr-note', 'First match wins, top to bottom. Switch a rule off or move it, then save. Saving here makes the set Custom.');
  const list = el('ol', 'd-lr-rules');
  list.id = 'lr-rules';
  const save = d.ui.btn('Save order and switches', { variant: 'primary' });
  save.id = 'lr-save';
  save.dataset.key = 'lr-save';
  custom.append(customNote, list, save);

  const prot = el('div', 'd-lr-prot');
  const protB = d.ui.btn('Always protected', { size: 'sm', variant: 'ghost' });
  protB.dataset.key = 'lr-protected';
  const protL = el('ul', 'd-lr-plist');
  protL.hidden = true;
  protB.addEventListener('click', () => { protL.hidden = !protL.hidden; });
  prot.append(protB, protL);

  const why = el('div', 'd-lr-why');
  why.append(el('h3', '', 'Why was this kept, sold or left?'));
  const recent = el('ul', 'd-lr-recent');
  recent.id = 'lr-recent';
  const more = d.ui.btn('Show more', { size: 'sm', variant: 'ghost' });
  more.dataset.key = 'lr-more';
  why.append(recent, more);

  const row = el('div', 'd-lr-row');
  const closeB = d.ui.btn('Close', { variant: 'ghost' });
  closeB.dataset.key = 'lr-close';
  row.append(closeB);
  card.append(title, stage, err, modes, simple, custom, prot, why, row);
  wrap.appendChild(card);
  doc.body.appendChild(wrap);

  let opener = null, v = null, mode = 'starter', order = [], off = {}, shown = 8;
  const fail = (m) => { err.hidden = !m; err.textContent = m || ''; };
  const close = () => { wrap.hidden = true; if (opener && opener.focus) opener.focus({ preventScroll: true }); opener = null; };

  function setMode(m) {
    mode = m;
    simple.hidden = m !== 'starter';
    custom.hidden = m !== 'custom';
    bStarter.setAttribute('aria-pressed', String(m === 'starter'));
    bCustom.setAttribute('aria-pressed', String(m === 'custom'));
  }

  function drawRules() {
    list.textContent = '';
    order.forEach((r, i) => {
      const li = el('li', 'd-lr-rule');
      const on = el('input'); on.type = 'checkbox'; on.checked = !off[r.id]; on.dataset.key = 'lr-on'; on.setAttribute('aria-label', 'Rule on: ' + r.id);
      on.addEventListener('change', () => { off[r.id] = !on.checked; li.classList.toggle('is-off', !on.checked); });
      li.classList.toggle('is-off', !!off[r.id]);
      const text = el('span', 'd-lr-text', (v.plain && v.plain[r.id]) || r.reason);
      const up = d.ui.btn('Up', { size: 'sm', variant: 'ghost' }); up.dataset.key = 'lr-up'; up.setAttribute('aria-label', 'Move up: ' + r.id); up.disabled = i === 0;
      const dn = d.ui.btn('Down', { size: 'sm', variant: 'ghost' }); dn.dataset.key = 'lr-down'; dn.setAttribute('aria-label', 'Move down: ' + r.id); dn.disabled = i === order.length - 1;
      up.addEventListener('click', () => { if (i > 0) { [order[i - 1], order[i]] = [order[i], order[i - 1]]; drawRules(); } });
      dn.addEventListener('click', () => { if (i < order.length - 1) { [order[i + 1], order[i]] = [order[i], order[i + 1]]; drawRules(); } });
      const act = el('span', 'd-lr-act'); act.append(up, dn);
      li.append(on, text, act);
      list.appendChild(li);
    });
  }

  function drawRecent() {
    recent.textContent = '';
    const rs = (v && v.recent) || [];
    if (!rs.length) recent.appendChild(el('li', 'd-lr-note', 'No decisions yet. They appear here as the bot looks at items.'));
    for (const r of rs.slice(0, shown)) {
      const li = el('li', 'd-lr-dec', decisionLine(r));
      if (r.today && r.today !== r.rules) li.dataset.differs = '1';
      recent.appendChild(li);
    }
    more.hidden = rs.length <= shown;
  }

  function draw(j) {
    v = j;
    stage.textContent = stageText(j.stage);
    const a = j.active || {};
    order = (a.rules || []).slice();
    off = {};
    for (const r of order) if (r.off) off[r.id] = true;
    preset.textContent = '';
    for (const p of j.presets || []) { const o = el('option', '', p.name); o.value = p.id; o.title = p.description || ''; preset.appendChild(o); }
    preset.value = a.preset && a.preset !== 'custom' ? a.preset : 'starter';
    const o = j.options || {};
    gear.value = o.sellGear || 'magic';
    rune.value = String(o.minRune || 1);
    gold.value = String(o.minGold || 0);
    sock.checked = o.keepSockets !== false;
    protL.textContent = '';
    for (const p of j.protected || []) protL.appendChild(el('li', '', p.Text || p.text || ''));
    setMode(a.preset === 'custom' ? 'custom' : 'starter');
    drawRules();
    drawRecent();
  }

  const body = (extra) => {
    const a = v.active;
    return Object.assign({ schema: a.schema, seq: v.seq, name: a.name, limits: a.limits }, extra);
  };
  function send(b, btn) {
    btn.disabled = true; fail('');
    return d.put(ROUTE, b).then((r) => {
      btn.disabled = false;
      if (r && r.j && r.j.ok) { d.toast('good', 'Loot rules saved'); return reload(); }
      fail(refusal(r));
    });
  }
  function reload() {
    return d.get(ROUTE).then((j) => { if (j && j.schema === SCHEMA) draw(j); else fail(refusal({ j: j || { code: 'off' } })); });
  }
  apply.addEventListener('click', () => {
    const n = Math.round(Number(rune.value)), g = Math.round(Number(gold.value));
    if (!(n >= 1 && n <= 33)) { fail('Lowest rune must be from 1 to 33'); return; }
    if (!(g >= 0 && g <= 100000)) { fail('Smallest gold pile must be from 0 to 100000'); return; }
    send(body({ preset: preset.value, options: { sellGear: gear.value, minRune: n, minGold: g, keepSockets: sock.checked }, rules: [] }), apply);
  });
  save.addEventListener('click', () => {
    send(body({ preset: 'custom', rules: order.map((r) => Object.assign({}, r, { off: !!off[r.id] || undefined })), defaults: v.active.defaults, groups: v.active.groups, ack: v.active.ack, comment: v.active.comment }), save);
  });
  bStarter.addEventListener('click', () => setMode('starter'));
  bCustom.addEventListener('click', () => setMode('custom'));
  more.addEventListener('click', () => { shown = 40; drawRecent(); });
  closeB.addEventListener('click', close);
  wrap.addEventListener('click', (e) => { if (e.target === wrap) close(); });
  wrap.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.preventDefault(); close(); return; }
    if (e.key === 'Tab') {
      const f = [bStarter, bCustom, preset, gear, rune, gold, sock, apply, protB, more, closeB].filter((b) => !b.hidden && !b.disabled && !(b.parentNode && b.parentNode.hidden));
      if (!f.length) return;
      const i = f.indexOf(doc.activeElement), n = e.shiftKey ? (i <= 0 ? f.length - 1 : i - 1) : (i === f.length - 1 ? 0 : i + 1);
      e.preventDefault();
      f[n].focus();
    }
  });
  const open = (from) => {
    opener = from || (doc.activeElement && doc.activeElement !== doc.body ? doc.activeElement : null);
    fail(''); shown = 8; v = null;
    wrap.hidden = false;
    card.scrollTop = 0;
    bStarter.focus({ preventScroll: true });
    return reload();
  };
  doc.addEventListener(ASK, (e) => open(e && e.detail && e.detail.from));
  if (d.overlays) d.overlays.register('lootrules', wrap, () => { wrap.hidden = true; });
  return { open, close, reload, get mode() { return mode; } };
}
