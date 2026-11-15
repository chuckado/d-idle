// "Health" cell of the cockpit strip (J-second-attempt.md section 2): five lights, each with a label and one word. A light
// the bot does not report is a hollow disc with "not reported"; it is never drawn green. A click opens the detail in a
// popover. From GET /api/glance (health). The readiness word is safe, risky or too hard: words only, never a number.
const LIGHTS = [['life', 'Life'], ['gear', 'Gear'], ['readiness', 'Ready'], ['merc', 'Merc'], ['supplies', 'Supply']];
const KNOWN = ['ok', 'warn', 'bad'];

// lightOf is the state a light is drawn in: ok, warn, bad or unknown (anything the page does not know).
export function lightOf(l) { return l && KNOWN.includes(l.light) ? l.light : 'unknown'; }

export function wordOf(l) { return lightOf(l) === 'unknown' ? 'not reported' : (l.word || ''); }

// worst is the worst state among the lights that report; unknown ones do not count.
export function worst(h) {
  let w = 'ok', seen = false;
  for (const [k] of LIGHTS) {
    const s = lightOf(h && h[k]);
    if (s === 'unknown') continue;
    seen = true;
    if (s === 'bad' || (s === 'warn' && w === 'ok')) w = s;
  }
  return seen ? w : 'unknown';
}

export function mount(root, ctx) {
  const btn = ctx.el('button', 'h-btn');
  btn.type = 'button';
  btn.dataset.key = 'open';
  btn.setAttribute('aria-haspopup', 'true');
  const head = ctx.el('span', 'h-h');
  head.append(ctx.el('b', '', 'Health'));
  const lights = ctx.el('span', 'h-lights');
  const detail = ctx.el('dl', 'h-detail'); // shown only when the box is tall (the phone's Progress tab), where the lights leave room
  btn.append(head, lights, detail);
  const pop = ctx.el('div', 'h-pop');
  pop.setAttribute('popover', 'auto');
  pop.setAttribute('role', 'dialog');
  pop.setAttribute('aria-label', 'Health in detail');
  ctx.body.append(btn, pop);
  btn.addEventListener('click', () => { if (pop.togglePopover) pop.togglePopover(); });

  let sig = ''; // the last drawn health: an identical poll redraws nothing
  return {
    update(g) {
      const h = g && g.health;
      if (!h) { sig = ''; ctx.hide(true); return; }
      ctx.hide(false);
      const now = JSON.stringify(h);
      if (now === sig) return;
      sig = now;
      ctx.ui.clear(lights);
      ctx.ui.clear(pop);
      ctx.ui.clear(detail);
      const dl = ctx.el('dl', 'h-rows');
      for (const [k, name] of LIGHTS) {
        const l = h[k] || {}, s = lightOf(l), w = wordOf(l);
        const e = ctx.el('span', 'h-light');
        e.dataset.l = s;
        e.dataset.k = k;
        e.title = name + ': ' + w + (l.detail ? '. ' + l.detail : '');
        e.append(ctx.el('i'), ctx.el('b', '', name), ctx.el('span', '', w));
        lights.append(e);
        const said = w + (l.detail && s !== 'unknown' ? '. ' + l.detail : '');
        dl.append(ctx.el('dt', '', name), ctx.el('dd', '', said));
        const dd = ctx.el('dd', '', said);
        dd.dataset.l = s;
        detail.append(ctx.el('dt', '', name), dd);
      }
      btn.dataset.worst = worst(h);
      pop.append(dl);
      ctx.state('ready');
    },
  };
}
