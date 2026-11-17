// The Quest board sheet (docs/design/quest-board-2026-11-16.md section 4.3; slice S5) and the logic the quest strip and the
// phone's Quests tab share: the words for every state of GET /api/quests, the button each state offers, the action calls and
// the Hand in read-back. Words only: no margin number, no chapter counter. The page decides what happened from the next GET,
// never from a POST reply (the Pause bug of 2026-11-16: a reply without `ok` was shown as "could not do that"); a refusal is
// shown with the reason the API gave. This file is the drawer panel `quest-board`; `quest-strip` and `quest-list` import it.
// textContent only, no timers, no requests of their own (ctx.post, ctx.later).

export const KNOB = 'quests.board';
export const WORD = { safe: 'safe', risky: 'risky', too_hard: 'too hard' };
export const MODES = [['train', 'Train'], ['idle', 'Idle'], ['farm', 'Farm']];
const ROMAN = ['', 'I', 'II', 'III', 'IV', 'V'];
const HAND_STALE_MS = 6 * 60 * 1000; // the desk gives a hand-in 5 minutes; a longer wait is a lost reply
const NOW = { accepted: 1, in_progress: 1, awaiting_handin: 1, blocked: 1 };

export function actName(n) { return 'Act ' + (ROMAN[n] || n); }
export function wordOf(w) { return WORD[w] || ''; }

// humanize turns a step or chunk id into words: "walk_blood_moor" -> "Walk blood moor".
export function humanize(id) {
  const s = String(id || '').replace(/[_-]+/g, ' ').trim();
  return s ? s[0].toUpperCase() + s.slice(1) : '';
}

export function minutesText(m) {
  if (!Array.isArray(m) || !m[0]) return '';
  return m[1] && m[1] !== m[0] ? 'about ' + m[0] + ' to ' + m[1] + ' min' : 'about ' + m[0] + ' min';
}

// paysText is the meta payout in words; empty until the vault wallet exists (the API omits `pays`, slice S10).
export function paysText(p) {
  return p && p.meta > 0 ? '+' + p.meta + ' meta' : '';
}

export function modeLine(mode) {
  return { train: 'Training', idle: 'Idle in town', farm: 'Farming for loot' }[mode] || 'Training';
}

// What the page remembers between polls, shared by every quest panel of the page: the hand-in the player pressed (so the strip
// says "Handing in..." at once, before the next GET), the chunk last seen running or waiting (so its completion is announced
// once), and the chunks already announced.
export const shared = { handing: null, last: null, paid: new Set(), error: null };
export function resetShared() { shared.handing = null; shared.last = null; shared.paid = new Set(); shared.error = null; }

// The video card's quest line. quest-strip sets the text from each GET; the video panel shows it (it has no endpoint of its own).
export const questCaption = { text: '', subs: new Set() };
export function setCaption(t) {
  if (questCaption.text === t) return;
  questCaption.text = t;
  questCaption.subs.forEach((f) => f(t));
}

// captionOf is the quest line for a view: the quest, what the hero does in words and who is next. Empty: no caption.
export function captionOf(v) {
  if (!v || !v.enabled) return '';
  if (v.awaiting) return v.awaiting.questTitle + ': waiting for you at ' + (v.awaiting.npc || 'the quest giver');
  if (v.blocked) return v.blocked.questTitle + ': stuck';
  const r = v.running;
  if (!r) return '';
  return r.questTitle + ': ' + (r.step ? humanize(r.step).toLowerCase() : 'starting') + (r.npc ? ', next ' + r.npc : '');
}

const key = (q, c) => q + '/' + c;

function chunkOf(v, quest, chunk) {
  const q = ((v && v.quests) || []).find((x) => x.id === quest);
  const c = q && (q.chunks || []).find((x) => x.id === chunk);
  return q && c ? { q, c } : null;
}

export function chunkDone(v, quest, chunk) {
  const f = chunkOf(v, quest, chunk);
  return !!f && f.c.state === 'done';
}

function handFor(v, r) { // the hand-in record of the view when it is about this run
  const h = v && v.handIn;
  return h && r && h.quest === r.quest && h.chunk === r.chunk ? h : null;
}

// stripModel is what the strip tile says and offers for one view. local is `shared` (or a test's copy); now is Date.now().
// Returns {kind, label, head, sub, primary, secondary, quest, chunk}: kind is one of gone, off, handingin, confirming, failed,
// awaiting, blocked, running, offer, training, idle. A button is {label, verb, quest, chunk, confirm, disabled, why}.
export function stripModel(v, local = shared, now = Date.now()) {
  if (!v) return { kind: 'gone', label: '', head: 'The quest board did not answer', sub: '', primary: null, secondary: null };
  if (!v.enabled) return { kind: 'off', label: 'Quest board', head: 'The quest board is off', sub: offText(v), primary: null, secondary: null };
  const paused = !!v.paused;
  const btn = (label, verb, r, extra) => Object.assign({ label, verb, quest: r.quest, chunk: r.chunk }, extra || {});
  const aw = v.awaiting, bl = v.blocked, run = v.running;
  const who = (r) => r.npc || 'the quest giver';
  if (aw) {
    const h = handFor(v, aw);
    const handing = local.handing && local.handing.quest === aw.quest && local.handing.chunk === aw.chunk && now - local.handing.at < HAND_STALE_MS;
    const m = { quest: aw.quest, chunk: aw.chunk, secondary: null };
    if (h && h.state === 'confirming') return Object.assign(m, { kind: 'confirming', label: 'Handed in', head: 'Handed in to ' + who(aw) + '. Checking that the game agrees', sub: aw.questTitle, primary: null });
    if (h && h.state === 'running' || handing && !(h && (h.state === 'failed' || h.state === 'blocked'))) {
      return Object.assign(m, { kind: 'handingin', label: 'Handing in', head: 'Handing in ' + aw.questTitle + '...', sub: 'Talking to ' + who(aw) + '. The picture shows it', primary: btn('Handing in...', 'handin', aw, { disabled: true }) });
    }
    if (h && (h.state === 'failed' || h.state === 'blocked') || local.error && local.error.quest === aw.quest) {
      const why = (h && h.reason) || (local.error && local.error.reason) || 'the hand-in did not go through';
      return Object.assign(m, { kind: 'failed', label: 'Hand-in failed', head: 'Could not hand in ' + aw.questTitle, sub: why, primary: btn('Try again', 'handin', aw) });
    }
    return Object.assign(m, { kind: 'awaiting', label: 'Waiting for you', head: who(aw) + ' is waiting: hand in ' + aw.questTitle, sub: 'The hero stands there and waits for your tap', primary: btn('Hand in', 'handin', aw) });
  }
  if (bl) {
    return { kind: 'blocked', label: 'Stuck', quest: bl.quest, chunk: bl.chunk, head: bl.questTitle + ' is stuck', sub: bl.why || 'the step failed',
      primary: btn('Try again', 'retry', bl), secondary: btn('Stop', 'cancel', bl, { confirm: true }) };
  }
  if (run) {
    const step = run.step ? humanize(run.step) : 'Starting';
    const sub = [step, run.npc ? 'next: ' + run.npc : ''].filter(Boolean).join(', ');
    return { kind: 'running', label: 'Running', quest: run.quest, chunk: run.chunk, head: 'Running: ' + run.questTitle, sub: sub || run.title,
      primary: btn('Stop', 'cancel', run, { confirm: true }), secondary: null };
  }
  const offers = v.offers || [];
  const ok = offers.filter((o) => o.startable);
  const best = ok.find((o) => o.readiness === 'safe') || ok[0];
  if (best) {
    const w = wordOf(best.readiness);
    const sub = best.sentence || ((best.reasons || [])[0] || {}).text || best.title;
    return { kind: 'offer', label: 'Offer', quest: best.quest, chunk: best.chunk, head: 'Offer: ' + best.questTitle + (w ? ', ' + w : ''), sub,
      primary: btn('Start', 'start', best, paused ? { disabled: true, why: 'Paused: resume the bot first' } : {}), secondary: null };
  }
  const hard = offers[0];
  if (hard && v.mode !== 'idle') {
    const why = ((hard.reasons || [])[0] || {}).text || hard.why || hard.sentence || '';
    return { kind: 'training', label: modeLine(v.mode), head: (v.mode === 'farm' ? 'Farming before ' : 'Training toward ') + hard.questTitle, sub: why || 'Not ready for it yet', primary: null, secondary: null };
  }
  return { kind: 'idle', label: modeLine(v.mode), head: modeLine(v.mode), sub: offers.length ? 'Quests wait for you in the board' : 'No quest is on offer right now', primary: null, secondary: null };
}

export function offText(v) {
  if (v && v.reason && !/quests\.board|knob/i.test(v.reason)) return v.reason;
  return 'Turn on the Quest board in settings: set ' + KNOB + ' to 1';
}

// settle looks at a view for the Hand in payoff and returns {title, text} once per finished chunk, else null. The chunk last
// seen running, waiting or stuck is remembered; the view saying it is done is the proof (not a reply).
export function settle(v, local = shared, now = Date.now()) {
  if (!v || !v.enabled) return null;
  const cur = v.awaiting || v.blocked || v.running;
  if (local.handing) {
    const h = handFor(v, local.handing);
    const stale = now - local.handing.at >= HAND_STALE_MS;
    if (stale || chunkDone(v, local.handing.quest, local.handing.chunk) || (h && (h.state === 'failed' || h.state === 'blocked' || h.state === 'done'))) local.handing = null;
  }
  if (local.error && (!v.awaiting || v.awaiting.quest !== local.error.quest)) local.error = null;
  let out = null;
  if (local.last && chunkDone(v, local.last.quest, local.last.chunk)) {
    const k = key(local.last.quest, local.last.chunk);
    if (!local.paid.has(k)) {
      local.paid.add(k);
      const pays = paysText(v.handIn && handFor(v, local.last) ? v.handIn.pays : null);
      out = { title: local.last.title, text: 'Done: ' + local.last.title + (pays ? '   ' + pays : '') };
    }
    local.last = null;
  }
  if (cur) local.last = { quest: cur.quest, chunk: cur.chunk, title: cur.questTitle };
  else if (local.last && !chunkDone(v, local.last.quest, local.last.chunk)) local.last = null; // cancelled or reset: nothing to announce
  return out;
}

// act is the one place a button reaches the API: it posts, shows a refusal with the API's own reason and asks for a fresh GET.
// It returns the reply (or null when the route is unreachable; ctx.post has said so already).
export async function act(ctx, verb, quest, chunk, local = shared, redraw = null) {
  const mode = verb === 'mode';
  const path = mode ? '/api/quests/mode' : '/api/quests/' + encodeURIComponent(quest) + '/' + verb;
  // the auto toggle (slice S8, route not built yet) sends {auto: bool} in the place of a chunk
  const body = mode ? { mode: chunk } : verb === 'auto' ? { auto: !!chunk } : chunk ? { chunk } : {};
  if (verb === 'handin') { local.handing = { quest, chunk, at: Date.now() }; local.error = null; if (redraw) redraw(); }
  const j = await ctx.post(path, body);
  if (j && j.ok === false) {
    const reason = j.reason || j.error || 'refused with code ' + (j.code || 'unknown') + ' and no reason';
    if (verb === 'handin') { local.handing = null; local.error = { quest, chunk, reason }; }
    ctx.toast('warn', reason);
  } else if (!j && verb === 'handin') local.handing = null;
  if (redraw) redraw();
  ctx.refresh();
  return j;
}

// ---- the board ----

const GROUPS = [['ready', 'Ready'], ['available', 'Not yet'], ['locked', 'Locked'], ['done', 'Done'], ['skipped', 'Skipped']];
const GROUP_OF = (s) => (NOW[s] ? 'now' : s === 'ready' || s === 'available' || s === 'locked' || s === 'done' || s === 'skipped' ? s : 'locked');

// boardModel: the quests grouped by act, each with its state group and chunk rows, plus the Now line, the history and the mode.
export function boardModel(v, local = shared, now = Date.now()) {
  if (!v) return { empty: 'The quest board did not answer. Trying again.' };
  if (!v.enabled) return { empty: offText(v), off: true };
  const offerOf = (q, c) => (v.offers || []).find((o) => o.quest === q && o.chunk === c);
  const acts = new Map();
  for (const q of v.quests || []) {
    const chunks = (q.chunks || []).map((c) => {
      const o = offerOf(q.id, c.id);
      const w = wordOf(c.readiness || (o && o.readiness));
      const reason = (o && (o.sentence || ((o.reasons || [])[0] || {}).text)) || c.why || '';
      return { id: c.id, title: c.title, state: c.state, word: w, reason, minutes: o ? minutesText(o.minutes) : '', startable: !!(o && o.startable), why: o && !o.startable ? o.why : '', pays: o ? paysText(o.pays) : '' };
    });
    const g = GROUP_OF(q.state);
    if (!acts.has(q.act)) acts.set(q.act, []);
    acts.get(q.act).push({ id: q.id, title: q.title, act: q.act, state: q.state, group: g, auto: !!q.auto, chunks });
  }
  const order = (g) => (g === 'now' ? -1 : GROUPS.findIndex((x) => x[0] === g));
  const list = [...acts.keys()].sort((a, b) => a - b).map((n) => ({ act: n, name: actName(n), rows: acts.get(n).sort((a, b) => order(a.group) - order(b.group)) }));
  const cur = stripModel(v, local, now);
  const nowRow = ['running', 'awaiting', 'blocked', 'handingin', 'confirming', 'failed'].includes(cur.kind) ? cur : null;
  const hist = (v.history || []).slice(0, 5).map((h) => ({ title: h.title, text: (h.confirmed ? 'handed in' : 'started') + (paysText(h.pays) ? ', ' + paysText(h.pays) : '') }));
  const open = list.find((a) => a.rows.some((r) => r.group !== 'done' && r.group !== 'skipped'));
  return { acts: list, now: nowRow, history: hist, mode: v.mode || 'train', modeActs: !!v.modeActs, actLabel: (open ? open.name : list.length ? list[list.length - 1].name : 'Quests') + ', ' + humanize(v.difficulty || 'normal'),
    auto: v.autoStart || { available: false, note: '' }, paused: !!v.paused, empty: list.length ? '' : 'No quests to show yet' };
}

// mount builds the drawer panel; renderInto is shared with quest-list (the phone's Quests tab).
export function mount(root, ctx) { return mountBoard(root, ctx, false); }

export function mountBoard(root, ctx, tab) {
  const box = ctx.el('div', 'qb-box');
  ctx.body.append(box);
  let sig = '', stopArm = '', last = null;
  const local = shared;
  box.dataset.tab = tab ? '1' : '0';

  function btn(text, cls, k, fn, label) {
    const b = ctx.el('button', cls, text);
    b.type = 'button';
    b.dataset.key = k;
    if (label) b.setAttribute('aria-label', label);
    if (fn) b.addEventListener('click', () => fn(b));
    return b;
  }

  function press(verb, quest, chunk, confirm, b) {
    const id = verb + quest + chunk;
    if (confirm && stopArm !== id) { // a stop asks once: the second press within four seconds does it
      stopArm = id;
      redraw();
      ctx.later(() => { if (stopArm === id) { stopArm = ''; redraw(); } }, 4000);
      return;
    }
    stopArm = '';
    b.disabled = true;
    act(ctx, verb, quest, chunk, local, redraw).then(() => { b.disabled = false; });
  }

  function modeBar(m) {
    const bar = ctx.el('div', 'qb-mode');
    bar.setAttribute('role', 'radiogroup');
    bar.setAttribute('aria-label', 'What the hero does between quests');
    for (const [id, label] of MODES) {
      const b = btn(label, 'qb-seg', 'mode-' + id, () => { b.disabled = true; act(ctx, 'mode', '', id, local, redraw).then(() => { b.disabled = false; }); });
      b.setAttribute('role', 'radio');
      b.setAttribute('aria-checked', String(m.mode === id));
      b.dataset.on = m.mode === id ? '1' : '0';
      bar.append(b);
    }
    if (!m.modeActs) bar.title = 'Saved for this hero. The modes start to act in a later update';
    return bar;
  }

  function autoCell(m, row) {
    const c = ctx.el('span', 'qb-auto');
    if (!m.auto.available) {
      c.dataset.off = '1';
      c.append(ctx.el('span', '', 'auto: coming'));
      c.title = m.auto.note || 'Starting by itself when you are away arrives in a later update';
      return c;
    }
    const on = !!row.auto;
    const b = btn('auto: ' + (on ? 'on' : 'off'), 'qb-sw', 'auto-' + row.id, () => { b.disabled = true; act(ctx, 'auto', row.id, !on, local, redraw).then(() => { b.disabled = false; }); }, 'Start ' + row.title + ' by itself when you are away and it is safe');
    b.setAttribute('role', 'switch');
    b.setAttribute('aria-checked', String(on));
    b.dataset.on = on ? '1' : '0';
    b.title = 'When you are away, start this by itself when it is safe';
    c.append(b);
    return c;
  }

  function chunkLine(m, row, c, multi, idx) {
    const li = ctx.el('li', 'qb-chunk');
    li.dataset.state = c.state;
    const label = multi ? 'Part ' + (idx + 1) + ': ' + c.title : c.title;
    const stateWord = c.state === 'ready' || c.state === 'available' ? c.word : { locked: 'locked', done: 'done', skipped: 'skipped', accepted: 'started', in_progress: 'running', awaiting_handin: 'waiting for you', blocked: 'stuck' }[c.state] || c.state;
    const t = ctx.el('span', 'qb-t');
    t.append(ctx.el('b', '', label));
    li.title = label + (c.reason ? ' - ' + c.reason : '');
    li.append(t);
    if (stateWord) { const w = ctx.el('span', 'qb-w', stateWord); w.dataset.w = c.word || c.state; li.append(w); }
    if (c.reason && c.state !== 'done') li.append(ctx.el('span', 'qb-r', c.reason));
    const meta = [c.minutes, c.pays].filter(Boolean).join(', ');
    if (meta && c.state !== 'done') li.append(ctx.el('span', 'qb-m', meta));
    if (c.state === 'ready' && c.startable) {
      const dis = m.paused;
      const b = btn('Start', 'qb-btn qb-p', 'start-' + row.id + '-' + c.id, (x) => press('start', row.id, c.id, false, x), 'Start ' + row.title + (multi ? ', part ' + (idx + 1) : ''));
      if (dis) { b.disabled = true; b.title = 'Paused: resume the bot first'; }
      li.append(b);
    } else if (c.state === 'ready' && c.why) li.append(ctx.el('span', 'qb-r', c.why));
    return li;
  }

  function redraw() { sig = ''; render(last); }

  function render(v) {
    last = v;
    const m = boardModel(v, local);
    ctx.ui.clear(box);
    if (m.empty && !m.acts) { // off, or no answer: say how to fix it
      const e = ctx.el('div', 'qb-empty');
      e.dataset.kind = m.off ? 'off' : 'gone';
      e.append(ctx.el('b', '', m.off ? 'The quest board is off' : 'The quest board did not answer'), ctx.el('p', '', m.empty));
      if (m.off) e.append(ctx.el('code', '', KNOB + ' = 1'));
      box.append(e);
      return;
    }
    const head = ctx.el('div', 'qb-head');
    head.append(ctx.el('span', 'qb-act', m.actLabel));
    if (!tab) head.append(modeBar(m)); // the phone's tile above the list has the switch
    box.append(head);
    if (m.now) {
      const n = ctx.el('div', 'qb-now');
      n.dataset.kind = m.now.kind;
      const t = ctx.el('span', 'qb-t');
      t.append(ctx.el('b', '', m.now.head), ctx.el('span', 'qb-r', m.now.sub));
      n.append(t);
      for (const b of [m.now.primary, m.now.secondary].filter(Boolean)) {
        const armed = b.confirm && stopArm === b.verb + b.quest + b.chunk;
        const x = btn(armed ? 'Stop it?' : b.label, b === m.now.primary ? 'qb-btn qb-p' : 'qb-btn', 'now-' + b.verb, (e) => press(b.verb, b.quest, b.chunk, b.confirm, e));
        if (armed) x.dataset.armed = '1';
        if (b.disabled) x.disabled = true;
        n.append(x);
      }
      box.append(n);
    }
    const list = ctx.el('div', 'qb-list');
    list.setAttribute('data-scroll', '');
    for (const a of m.acts) {
      const sec = ctx.el('section', 'qb-act-sec');
      sec.append(ctx.el('h3', '', a.name));
      const ul = ctx.el('ul', 'qb-rows');
      for (const row of a.rows) {
        const li = ctx.el('li', 'qb-row');
        li.dataset.group = row.group;
        const multi = row.chunks.length > 1;
        const gl = GROUPS.find((g) => g[0] === row.group);
        const done = row.group === 'done' || row.group === 'skipped';
        const t = ctx.el('span', 'qb-t');
        t.append(ctx.el('b', '', row.title));
        if (multi) t.append(ctx.el('span', 'qb-m', ' ' + row.chunks.length + ' parts'));
        li.append(t, ctx.el('span', 'qb-g', row.group === 'now' ? 'now' : gl ? gl[1] : ''));
        if (!done && row.group !== 'locked') li.append(autoCell(m, row));
        const inner = ctx.el('ul', 'qb-chunks');
        // a quest of one part shows its line once; a quest of several lists every part with its own word and Start
        const rows = done ? [] : multi ? row.chunks : row.chunks.slice(0, 1);
        rows.forEach((c, i) => inner.append(chunkLine(m, row, c, multi, i)));
        if (inner.children.length) li.append(inner);
        ul.append(li);
      }
      sec.append(ul);
      list.append(sec);
    }
    if (m.history.length) {
      const h = ctx.el('section', 'qb-act-sec qb-hist');
      h.append(ctx.el('h3', '', 'Recent'));
      const ul = ctx.el('ul', 'qb-rows');
      for (const x of m.history) { const li = ctx.el('li', 'qb-row qb-h'); li.append(ctx.el('b', '', x.title), ctx.el('span', 'qb-r', x.text)); ul.append(li); }
      h.append(ul);
      list.append(h);
    }
    box.append(list);
  }

  return {
    update(v) {
      const paid = settle(v, local);
      if (paid) ctx.toast('good', paid.text);
      const s = JSON.stringify([v && Object.assign({}, v, { generated: 0 }), local.handing, local.error, stopArm]);
      if (s === sig) { ctx.state('ready'); return; }
      sig = s;
      render(v);
      ctx.state('ready');
    },
  };
}
