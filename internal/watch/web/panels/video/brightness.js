// Viewer-side Brightness for the game picture (per browser, never the game's own gamma). It only sets --vid-bright on the
// frame; panel.css turns it into filter: brightness() on the picture. Default 100% leaves the pixels untouched (the
// stylesheet applies no filter then). Additive: one button in the card's actions, remembered through ctx.store.
export const LEVELS = [
  { label: 'Darkest', pct: 70 },
  { label: 'Dark', pct: 85 },
  { label: 'Default', pct: 100 },
  { label: 'Bright', pct: 115 },
  { label: 'Brightest', pct: 130 },
];
const DEFAULT = 2;

export function levelFor(saved) {
  const i = LEVELS.findIndex((l) => String(l.pct) === String(saved));
  return i < 0 ? DEFAULT : i;
}

export function mountBrightness(ctx, frame) {
  const mem = ctx.store('brightness');
  const btn = ctx.el('button', 'd-btn', '');
  btn.type = 'button'; btn.dataset.size = 'sm'; btn.dataset.key = 'brightness';
  btn.title = 'Brightness of this picture in your browser only; the game itself is not changed';
  let cur = DEFAULT;
  function apply(i, remember) {
    cur = i;
    const l = LEVELS[i];
    if (l.pct === 100) frame.style.setProperty('--vid-bright', '1');
    else frame.style.setProperty('--vid-bright', String(l.pct / 100));
    frame.dataset.bright = String(l.pct);
    btn.textContent = 'Brightness: ' + l.label;
    btn.setAttribute('aria-label', 'Brightness ' + l.label + ' (' + l.pct + '%), press to change');
    if (remember) mem.set(String(l.pct));
  }
  apply(levelFor(mem.get()), false);
  btn.addEventListener('click', () => apply((cur + 1) % LEVELS.length, true));
  ctx.actions.append(btn);
  return btn;
}
