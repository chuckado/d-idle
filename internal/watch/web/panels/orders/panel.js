// The Orders panel (docs/ui/redesign-2026-11-14/C-orders.md, slice S4: read-only except the cancel cross). It draws
// GET /api/commands (watch.commands/1): the verb row, one safety line, one 28 px line per player command in the order
// the bot runs them, the blocked lines, and the plan as the last, grey line. The verbs are listed but inert here (the
// write side is S5). Cancel posts to the route the row itself names; nothing else is posted. textContent only.
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
  ctx.body.append(head, verbs, safety, list, plan);

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
        ctx.hide(false);
        safety.hidden = false;
        safety.dataset.tone = 'muted';
        safety.textContent = 'No connection to the hero. Orders are kept and shown when it is back.';
        return;
      }
      ctx.hide(false);
      const rows = v.rows || [];
      count.textContent = openCount(rows) + ' open';
      ctx.ui.clear(verbs);
      (v.verbs || []).forEach((vb) => {
        const b = ctx.ui.btn(vb.label, { size: 'sm', intent: vb.intent, inert: '1' });
        b.title = vb.reason || 'Orders are given from the order card for now';
        verbs.appendChild(b);
      });
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
