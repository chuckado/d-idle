// "N need you" in the cockpit band, toned by the worst severity among the needs; "all quiet" and muted at zero. From
// GET /api/glance (needs). The rows themselves are the Needs you panel's; this is only the count.
const RANK = { info: 1, warn: 2, danger: 3 };

// worst is the highest severity of the needs ('' for none).
export function worst(needs) {
  let w = '';
  for (const n of needs || []) if ((RANK[n.sev] || 1) > (RANK[w] || 0)) w = RANK[n.sev] ? n.sev : 'info';
  return w;
}

export function label(n) { return n ? n + (n > 1 ? ' need you' : ' needs you') : 'all quiet'; }

export function mount(root, ctx) {
  const pill = ctx.el('span', 'pill', '...');
  pill.dataset.n = '0';
  pill.dataset.tone = 'none';
  root.append(pill);
  return {
    update(v) {
      if (!v || !Array.isArray(v.needs)) { ctx.hide(true); return; }
      ctx.hide(false);
      const n = v.needs.length;
      pill.dataset.n = String(n);
      pill.dataset.tone = n ? worst(v.needs) : 'none';
      pill.textContent = label(n);
      pill.title = n ? v.needs.map((x) => x.text).filter(Boolean).join('; ') : '';
    },
  };
}
