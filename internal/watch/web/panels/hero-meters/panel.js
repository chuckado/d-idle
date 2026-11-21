// Life and mana in the top band (the approved mockup's status tile): who and where in one line, two thin meters with
// their numbers, from GET /api/hero. A value the bot has not reported shows as a dash, never as zero.
export function level(cur, max) {
  if (!max) return '';
  const p = (100 * cur) / max;
  return p < 25 ? 'bad' : p < 50 ? 'warn' : '';
}

export function mount(root, ctx) {
  const where = ctx.el('span', 'where');
  const life = ctx.ui.meter('life', 'life'), mana = ctx.ui.meter('mana', 'mana');
  life.root.classList.add('thin');
  mana.root.classList.add('thin');
  life.root.dataset.key = 'life';
  mana.root.dataset.key = 'mana';
  // in the cockpit band (the shell marks the root) the panel is slim: two thin bars and one number each, no line of words
  const slim = root.dataset.cockpitPanel === 'band';
  if (slim) root.dataset.slim = '1'; else root.append(where);
  root.append(life.root, mana.root);
  return {
    update(h) {
      if (!h || !h.known) { ctx.hide(true); return; }
      ctx.hide(false);
      where.textContent = [h.name, h.class ? h.class + (h.level ? ' lvl ' + h.level : '') : '', h.area].filter(Boolean).join(', ');
      const lt = slim ? (h.life) : h.life + '/' + h.maxLife, mt = slim ? h.mana : h.mana + '/' + h.maxMana;
      life.set(h.maxLife ? (100 * h.life) / h.maxLife : 0, h.maxLife ? String(lt) : '-', level(h.life, h.maxLife));
      mana.set(h.maxMana ? (100 * h.mana) / h.maxMana : 0, h.maxMana ? String(mt) : '-');
      if (slim && h.maxLife) life.root.title = 'life ' + h.life + ' of ' + h.maxLife;
      if (slim && h.maxMana) mana.root.title = 'mana ' + h.mana + ' of ' + h.maxMana;
    },
  };
}
