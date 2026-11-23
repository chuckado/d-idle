// The quest tile of the cockpit strip (design 4.2; slice S5): what the board is doing in words and the one button that fits,
// from GET /api/quests. Offer: "Offer: Den of Evil, safe" [Start]. Running: the step in words [Stop]. Waiting at the NPC:
// [Hand in]; the tile then says "Handing in..." until the next GET shows the result. A stuck chunk: [Try again] [Stop]. With
// the board off (knob quests.board 0) the tile hides itself; the Journey tile it once replaced was removed in slice S9. The page decides what
// happened from the GET; a refusal shows the reason the API gave. textContent only.
import { stripModel, settle, act, shared, MODES, modeLine, setCaption, captionOf } from '../quest-board/panel.js';

export { stripModel };

export function mount(root, ctx) {
  const label = ctx.el('span', 'qs-label');
  const side = ctx.el('span', 'qs-side');
  const head = ctx.el('div', 'qs-h');
  head.append(label, side);
  const line = ctx.el('div', 'qs-line');
  const sub = ctx.el('div', 'qs-sub');
  const acts = ctx.el('div', 'qs-acts');
  const modes = ctx.el('div', 'qs-mode');
  modes.setAttribute('role', 'radiogroup');
  modes.setAttribute('aria-label', 'What the hero does between quests');
  const box = ctx.el('div', 'qs-box');
  const row = ctx.el('div', 'qs-row');
  row.append(acts, modes);
  box.append(head, line, sub, row);
  ctx.body.append(box);

  let sig = '', stopArm = '', last = null;

  function redraw() { sig = ''; draw(last); }

  function press(b, el) {
    const id = b.verb + b.quest + b.chunk;
    if (b.confirm && stopArm !== id) {
      stopArm = id;
      redraw();
      ctx.later(() => { if (stopArm === id) { stopArm = ''; redraw(); } }, 4000);
      return;
    }
    stopArm = '';
    el.disabled = true;
    act(ctx, b.verb, b.quest, b.chunk, shared, redraw, !!b.confirm).then(() => { el.disabled = false; });
  }

  function button(b, primary, k) {
    const armed = b.confirm && stopArm === b.verb + b.quest + b.chunk;
    const e = ctx.el('button', primary ? 'qs-btn qs-p' : 'qs-btn', armed ? 'Stop it?' : b.label);
    e.type = 'button';
    e.dataset.key = k;
    if (armed) e.dataset.armed = '1';
    if (b.disabled) e.disabled = true;
    e.addEventListener('click', () => press(b, e));
    return e;
  }

  function draw(v) {
    last = v;
    const m = stripModel(v);
    box.dataset.kind = m.kind;
    box.dataset.short = ctx.short && ctx.short() ? '1' : '0'; // a short window (92 px strip): the kind label goes
    label.textContent = m.label;
    line.textContent = m.head;
    line.title = m.head;
    sub.textContent = m.sub;
    sub.title = m.sub;
    ctx.ui.clear(acts);
    ctx.ui.clear(modes);
    ctx.ui.clear(side);
    for (const [b, primary, k] of [[m.primary, true, 'primary'], [m.secondary, false, 'secondary']]) {
      if (!b) continue;
      const e = button(b, primary, k);
      acts.append(e);
      if (b.disabled) { // off: the reason is a visible line beside the button, named by aria-describedby (card uireason)
        const w = ctx.el('p', 'd-why', b.why || 'Not available right now.');
        w.id = 'qs-why-' + k;
        w.title = w.textContent;
        e.setAttribute('aria-describedby', w.id);
        acts.append(w);
      }
    }
    if (v && v.enabled) {
      if (!m.secondary) {
      const all = ctx.el('button', 'qs-btn qs-all', 'All quests');
      all.type = 'button';
      all.dataset.key = 'all';
      all.setAttribute('aria-label', 'Open the quest board');
      all.addEventListener('click', () => ctx.openDrawer('quest-board'));
      acts.append(all);
      }
      for (const [id, text] of MODES) {
        const b = ctx.el('button', 'qs-seg', text);
        b.type = 'button';
        b.dataset.key = 'mode-' + id;
        b.setAttribute('role', 'radio');
        b.setAttribute('aria-checked', String(v.mode === id));
        b.dataset.on = v.mode === id ? '1' : '0';
        b.setAttribute('aria-label', modeLine(id));
        b.addEventListener('click', () => { b.disabled = true; act(ctx, 'mode', '', id, shared, redraw).then(() => { b.disabled = false; }); });
        modes.append(b);
      }
      if (v.running || v.awaiting || v.blocked) side.textContent = 'after this quest: ' + modeLine(v.mode).toLowerCase();
    }
  }

  return {
    update(v) {
      setCaption(captionOf(v));
      const paid = settle(v);
      if (paid) ctx.toast('good', paid.text);
      if (!v) { if (!sig) ctx.hide(true); ctx.state('ready'); return; } // one failed poll keeps the tile as it was
      if (!v.enabled) { sig = ''; ctx.hide(true); ctx.state('ready'); return; }
      ctx.hide(false);
      const s = JSON.stringify([Object.assign({}, v, { generated: 0 }), shared.handing, shared.error, stopArm]);
      if (s !== sig) { sig = s; draw(v); }
      ctx.state('ready');
    },
  };
}
