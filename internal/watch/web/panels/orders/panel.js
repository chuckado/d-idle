// The Orders panel (docs/ui/redesign-2026-11-14/C-orders.md, slices S4 and S5). It draws GET /api/commands
// (watch.commands/1): the verb row, one safety line, one 28 px line per player command in the order the bot runs them,
// the blocked lines, and the plan as the last, grey line. S5 makes the verb row live: every button posts one command to
// POST /api/commands (the dispatcher over the old routes), each accepted command shows a five-second Undo toast, the
// More sheet lists the errands, and the keys g t r p o z / and Esc work. The zone map (panels/zone-map) imports sendCommand
// and its Undo toast from here, so there is one client for the one door. Cancel posts to the route the row itself names.
// textContent only.
const CHIP = { queued: 'queued', waiting: 'waiting', active: 'active', failed: 'failed' };

// mark is the narrow first column: ">" the row that runs, the place number of a row that waits, "!" a failure.
export function mark(r, place) {
  if (r.state === 'active') return '>';
  if (r.state === 'failed') return '!';
  return String(place);
}

// reason is the muted text after the title: the row's own reason, else its phase.
export function reason(r) {
  return (r.why && r.why.text) || r.phase || '';
}

export function openCount(rows) {
  return (rows || []).filter((r) => r.state !== 'failed').length;
}

// ---- the one client of POST /api/commands ----
export const COMMANDS = '/api/commands';
const UNDO_MS = 5000;

// Words of a refusal: the dispatcher's own reply, never a status code.
export function refusal(j) {
  if (!j) return 'no answer from the hero';
  if (j.error) return j.code === 'off' ? 'not available on this hero yet' : j.error; // a specific reason is always shown as it is
  if (j.code === 'off') return 'not available on this hero yet';
  if (j.code === 'locked') return j.error || 'not unlocked yet';
  return j.error || 'could not do that';
}

// The toast text of an accepted command: the order's own text when the reply carries one, else a short word per intent.
const ACK = { repair: 'Repair asked', farm_gold: 'Farming gold', pause: 'Paused', resume: 'Resumed', loop: 'Loop started', town: 'Going to town',
  identify: 'Identify queued', sell_now: 'Sell queued', revive_merc: 'Mercenary revive asked', fight_here: 'Fighting here' };
export function ack(intent, j) {
  if (j && j.order && j.order.text) return 'Order given: ' + j.order.text;
  if (j && j.entry && j.entry.text) return 'Queued: ' + j.entry.text;
  return ACK[intent] || 'Done';
}

let lastUndo = null, keysOff = null;
let lastRows = [];

// replacing is how many open orders "Go now" would cancel first (the hub's replace mode closes every open order, queued or running;
// queue entries, the repair, the farm and the Loop are not orders in that sense).
export function replacing() {
  return lastRows.filter((r) => /^order:/.test(r.id || '') && (r.state === 'queued' || r.state === 'active' || r.state === 'waiting')).length;
}
const REPLACES = ['go', 'fight_here', 'town'];
// A toast with Undo that stays five seconds (the shell's toast with an action). Undo posts the request the dispatcher named;
// the key z does the same while the toast is up.
export function toastUndo(ctx, text, undo) {
  if (!undo) { ctx.toast('good', text); return; }
  const mine = { undo };
  lastUndo = mine;
  ctx.toast('good', text, { label: 'Undo', ms: UNDO_MS, run: () => { if (lastUndo === mine) doUndo(ctx); } });
  ctx.later(() => { if (lastUndo === mine) lastUndo = null; }, UNDO_MS);
}

export function doUndo(ctx) {
  const u = lastUndo;
  if (!u) return Promise.resolve(false);
  lastUndo = null;
  return ctx.post(u.undo.path, u.undo.body).then((j) => {
    if (j && j.ok === false) ctx.toast('warn', refusal(j));
    else ctx.toast('info', 'Undone');
    return !(j && j.ok === false);
  });
}

// sendCommand posts {intent, args, mode}. On success it shows the Undo toast and resolves to the reply; a refusal resolves
// to the reply too (ok: false) so the caller can show a charter or its own line; the toast says why unless quiet.
export function sendCommand(ctx, intent, args, mode, quiet) {
  const body = { intent, args: args || {} };
  if (mode) body.mode = mode;
  // Undo takes back only the new order, so the toast says what Go now cancelled before it
  const gone = REPLACES.includes(intent) && mode !== 'queue' ? replacing() : 0;
  return ctx.post(COMMANDS, body).then((j) => {
    if (j && !j.error && j.schema === 'watch.pause/1') j = Object.assign({}, j, { ok: true }); // a pause view without the flag is an accepted press
    if (j && j.ok) { toastUndo(ctx, ack(intent, j) + (gone ? '; ' + gone + ' earlier ' + (gone === 1 ? 'order' : 'orders') + ' cancelled' : ''), j.undo); if (ctx.refresh) ctx.refresh(); return j; }
    if (!quiet) ctx.toast('warn', refusal(j));
    return j || { ok: false };
  });
}

// The errands of the More sheet: label, intent, and what the row list says is on.
export function moreItems(rows) {
  const farming = (rows || []).some((r) => r.intent === 'farm_gold' && r.state === 'active');
  return [
    { intent: 'identify', label: 'Identify all', args: { all: true } },
    { intent: 'sell_now', label: 'Sell now' },
    { intent: 'revive_merc', label: 'Revive mercenary' },
    { intent: 'farm_gold', label: farming ? 'Stop farming gold' : 'Farm gold', args: { on: !farming } },
    { intent: 'reset', label: 'Reset (leave the game and rejoin)', ask: true },
    { intent: 'characters', label: 'Characters (switch, create, delete)', go: '/watch/characters' }, // its own page (docs/charselect-ui.md)
  ];
}

// Keys are off while a text field has focus, in overlay mode, and with a modifier held (the browser's own shortcuts).
export function typing(t) {
  const tag = t && t.tagName ? String(t.tagName).toLowerCase() : '';
  return tag === 'input' || tag === 'textarea' || tag === 'select' || !!(t && t.isContentEditable);
}

export function mount(root, ctx) {
  const head = ctx.el('div', 'o-head');
  const title = ctx.el('span', '', 'Orders'), count = ctx.el('em', '');
  head.append(title, count);
  const verbs = ctx.el('div', 'o-verbs');
  const safety = ctx.el('div', 'o-safety');
  safety.hidden = true;
  const list = ctx.el('div', 'o-list');
  list.dataset.scroll = '1';
  const plan = ctx.el('div', 'o-row o-plan');
  const why = ctx.el('div', 'o-why');
  why.hidden = true;
  ctx.body.append(head, verbs, why, safety, list, plan);
  let view = null, down = false;

  // the More sheet: the errands as a list with a filter ("/" focuses it), inside this tile, over the list
  const ts = ctx.tileSheet('self'), sheet = ts.el, filter = ctx.el('input', 'o-filter'), items = ctx.el('div', 'o-items');
  sheet.hidden = true;
  filter.type = 'search'; filter.placeholder = 'Type to find a command'; filter.setAttribute('aria-label', 'Find a command');
  items.dataset.scroll = '1';
  const closeSheet = () => { sheet.hidden = true; };
  const sheetHead = ctx.el('div', 'o-sheethead', 'More'), sx = ctx.el('button', 'o-x', '×');
  sx.type = 'button'; sx.setAttribute('aria-label', 'Close'); sx.addEventListener('click', closeSheet);
  sheetHead.appendChild(sx);
  sheet.append(sheetHead, filter, items);
  ts.place();
  function drawItems() {
    ctx.ui.clear(items);
    const q = (filter.value || '').toLowerCase();
    moreItems(view && view.rows).filter((m) => m.label.toLowerCase().includes(q)).forEach((m) => {
      const b = ctx.el('button', 'o-item', m.label);
      b.type = 'button'; b.dataset.key = 'more-' + m.intent;
      b.addEventListener('click', () => {
        closeSheet();
        if (m.go) location.assign(m.go + (/^(dark|light|auto)$/.test(new URLSearchParams(location.search).get('theme') || '') ? '?theme=' + new URLSearchParams(location.search).get('theme') : ''));
        else if (m.ask) document.dispatchEvent(new CustomEvent('d2:reset-ask', { detail: { from: b } })); // its own route and confirm sheet (resetui.js)
        else sendCommand(ctx, m.intent, m.args);
      });
      items.appendChild(b);
    });
  }
  function openSheet() { filter.value = ''; drawItems(); sheet.hidden = false; filter.focus(); }
  filter.addEventListener('input', drawItems);
  filter.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') { const f = items.children[0]; if (f && f.click) f.click(); }
  });

  function verb(vb) {
    const paused = view && view.paused;
    switch (vb.intent) {
      case 'go': document.dispatchEvent(new CustomEvent('d2:goto')); return;
      case 'more': if (sheet.hidden) openSheet(); else closeSheet(); return;
      case 'pause': sendCommand(ctx, paused ? 'resume' : 'pause'); return;
      default: sendCommand(ctx, vb.intent);
    }
  }
  const verbByKey = (k) => ((view && view.verbs) || []).find((vb) => vb.intent === k && vb.enabled);

  // keys: one listener per page; a remount replaces it
  if (typeof document !== 'undefined' && document.addEventListener) {
    if (keysOff) keysOff();
    const onKey = (e) => {
      const k = e.key;
      if (k === 'Escape') { if (!sheet.hidden) { closeSheet(); e.preventDefault(); } document.dispatchEvent(new CustomEvent('d2:zone-close')); return; }
      if (e.defaultPrevented || e.ctrlKey || e.metaKey || e.altKey || typing(e.target || {}) || (document.documentElement && document.documentElement.dataset.mode === 'overlay')) return;
      if (down) return;
      const hit = { g: 'go', t: 'town', r: 'repair', p: 'pause' }[k];
      if (hit) { const vb = verbByKey(hit); if (vb) { e.preventDefault(); verb(vb); } return; }
      if (k === 'z') { e.preventDefault(); doUndo(ctx); }
      else if (k === 'o') { e.preventDefault(); list.focus(); }
      else if (k === '/') { e.preventDefault(); openSheet(); }
    };
    document.addEventListener('keydown', onKey);
    keysOff = () => document.removeEventListener('keydown', onKey);
  }
  list.tabIndex = 0;
  list.setAttribute('aria-label', 'Orders');

  function cancel(r, btn) {
    btn.disabled = true;
    ctx.post(r.cancel.path, r.cancel.body).then((j) => {
      if (j && j.ok === false) ctx.toast('warn', j.error || 'could not cancel');
      btn.disabled = false;
    });
  }

  function row(r, place) {
    const e = ctx.el('div', 'o-row');
    e.dataset.s = r.state;
    e.dataset.id = r.id;
    const why = reason(r);
    e.title = r.text + (why ? ' - ' + why : '');
    const chipText = r.next && r.state !== 'active' ? 'next' : CHIP[r.state] || r.state;
    const chip = ctx.el('span', 'o-chip', chipText);
    chip.dataset.s = r.next && r.state !== 'active' ? 'next' : r.state;
    e.append(ctx.el('span', 'o-ix', mark(r, place)), ctx.el('span', 'o-t', r.text), ctx.el('span', 'o-r', why), chip);
    if (r.can && r.can.cancel && r.cancel) {
      const x = ctx.el('button', 'o-x', '×');
      x.type = 'button';
      x.dataset.key = 'x-' + r.id;
      x.setAttribute('aria-label', 'Cancel: ' + r.text);
      x.addEventListener('click', () => cancel(r, x));
      e.appendChild(x);
    } else {
      e.appendChild(ctx.el('span', 'o-x'));
    }
    return e;
  }

  return {
    update(v, err) {
      if (!v) {
        if (err && (err.status === 404 || err.status === 503)) { ctx.hide(true); return; }
        down = true;
        verbs.querySelectorAll && verbs.querySelectorAll('button').forEach((b) => { b.disabled = true; });
        for (const b of verbs.children) b.disabled = true;
        ctx.hide(false);
        safety.hidden = false;
        safety.dataset.tone = 'muted';
        safety.textContent = 'No connection to the hero. Orders are kept and shown when it is back.';
        return;
      }
      ctx.hide(false);
      down = false;
      view = v;
      const rows = v.rows || [];
      lastRows = rows;
      count.textContent = openCount(rows) + ' open';
      ctx.ui.clear(verbs);
      const whys = [];
      (v.verbs || []).forEach((vb) => {
        const b = ctx.ui.btn(vb.label, { size: 'sm', intent: vb.intent });
        b.dataset.key = 'verb-' + vb.intent;
        b.title = vb.reason || vb.label;
        if (!vb.enabled) { b.disabled = true; if (vb.reason) whys.push(vb.label + ': ' + vb.reason); } // a dead button always has its reason next to it
        else b.addEventListener('click', () => verb(vb));
        verbs.appendChild(b);
      });
      why.hidden = !whys.length;
      why.textContent = whys.join('; ');
      if (!sheet.hidden) drawItems();
      safety.hidden = !(v.safety && v.safety.active);
      safety.dataset.tone = 'warn';
      safety.textContent = v.safety && v.safety.active ? v.safety.text : '';
      ctx.ui.clear(list);
      let place = 0;
      rows.forEach((r) => { if (r.state === 'queued' || r.state === 'waiting') place++; list.appendChild(row(r, place)); });
      (v.blocked || []).forEach((b) => {
        const e = ctx.el('div', 'o-row');
        e.dataset.s = 'blocked';
        e.title = b.text;
        e.append(ctx.el('span', 'o-ix', '▲'), ctx.el('span', 'o-t o-wide', b.text));
        list.appendChild(e);
      });
      if (!rows.length && !(v.blocked || []).length) {
        list.appendChild(ctx.el('div', 'o-empty', 'No orders. The hero follows the plan.'));
      }
      ctx.ui.clear(plan);
      plan.hidden = !v.plan;
      if (v.plan) {
        plan.title = v.plan.text;
        plan.append(ctx.el('span', 'o-ix', '~'), ctx.el('span', 'o-t', 'Then the plan'), ctx.el('span', 'o-r', v.plan.text));
      }
      ctx.state('ready');
    },
  };
}
