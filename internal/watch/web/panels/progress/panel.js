// "Progress" cell of the cockpit strip (J-second-attempt.md section 2): the level meter with "N%, about M min", xp/h with a
// 30-point sparkline and gold/h. A rate still being measured shows a dash with a tooltip. A click opens a popover with the
// numbers behind it. From GET /api/glance (progress). Every number carries its unit or label; textContent only.
const NS = 'http://www.w3.org/2000/svg';
const W = 60, H = 16;

// short writes a rate: 12400 -> 12.4k, 950 -> 950.
export function short(n) {
  if (typeof n !== 'number' || !(n >= 0)) return '';
  if (n >= 1000) return (n / 1000).toFixed(n >= 100000 ? 0 : 1) + 'k';
  return String(Math.round(n));
}

// eta is "about 25 min" or "about 2 h 10 min"; '' while the rate builds up.
export function eta(sec) {
  if (sec == null || !(sec > 0)) return '';
  const m = Math.max(1, Math.round(sec / 60));
  if (m < 60) return 'about ' + m + ' min';
  const h = Math.floor(m / 60), r = m % 60;
  return 'about ' + h + ' h' + (r ? ' ' + r + ' min' : '');
}

// points is the sparkline as "x,y x,y ..." in a W by H box (a flat series runs through the middle); [] for fewer than two samples.
export function points(series) {
  const s = (series || []).filter((x) => typeof x === 'number');
  if (s.length < 2) return [];
  const lo = Math.min(...s), hi = Math.max(...s);
  return s.map((x, i) => [Math.round((i * (W - 4) / (s.length - 1) + 2) * 10) / 10, hi === lo ? H / 2 : Math.round((H - 3 - ((x - lo) / (hi - lo)) * (H - 6)) * 10) / 10]);
}

export function mean(series) {
  const s = (series || []).filter((x) => typeof x === 'number');
  return s.length ? s.reduce((a, b) => a + b, 0) / s.length : null;
}

const WAIT = 'Still measuring: a rate needs a few minutes of play.';

export function mount(root, ctx) {
  const btn = ctx.el('button', 'pg-btn');
  btn.type = 'button';
  btn.dataset.key = 'open';
  btn.setAttribute('aria-haspopup', 'true');
  const head = ctx.el('span', 'pg-h'), lvl = ctx.el('em', '');
  head.append(ctx.el('b', '', 'Progress'), lvl);
  const to = ctx.el('span', 'pg-to');
  const meter = ctx.el('progress', 'pg-meter');
  meter.setAttribute('max', '100');
  const said = ctx.el('span', 'pg-said');
  const m = ctx.el('span', 'pg-m');
  m.append(to, meter, said);
  const xp = ctx.el('span', 'pg-xp'), xpv = ctx.el('strong', '');
  const svg = root.ownerDocument.createElementNS(NS, 'svg'), line = root.ownerDocument.createElementNS(NS, 'polyline'), dot = root.ownerDocument.createElementNS(NS, 'circle');
  svg.setAttribute('class', 'pg-spark');
  svg.setAttribute('viewBox', '0 0 ' + W + ' ' + H);
  svg.setAttribute('aria-hidden', 'true');
  dot.setAttribute('r', '2');
  svg.append(line, dot);
  const spark = svg;
  xp.append(xpv, ctx.el('label', '', 'xp/h'), spark);
  const gd = ctx.el('span', 'pg-gold'), gdv = ctx.el('strong', '');
  gd.append(gdv, ctx.el('label', '', 'gold/h'));
  const v = ctx.el('span', 'pg-v');
  v.append(xp, gd);
  btn.append(head, m, v);
  const pop = ctx.el('div', 'pg-pop');
  pop.setAttribute('popover', 'auto');
  pop.setAttribute('role', 'dialog');
  pop.setAttribute('aria-label', 'Progress in detail');
  ctx.body.append(btn, pop);
  btn.addEventListener('click', () => { if (pop.togglePopover) pop.togglePopover(); });

  function rate(el, n, unit) {
    if (n == null) { el.textContent = '\u2013'; el.title = WAIT; el.dataset.wait = '1'; return; }
    el.textContent = short(n);
    el.title = Math.round(n) + ' ' + unit;
    delete el.dataset.wait;
  }

  function rows(p, lv) {
    const out = [];
    if (lv.pct != null && lv.from) out.push(['Level', 'level ' + lv.from + ' to ' + lv.to + ': ' + Math.round(lv.pct) + '% done']);
    const e = eta(lv.etaSeconds);
    out.push(['Time to next level', e || 'still measuring']);
    out.push(['XP now', p.xpPerHour == null ? 'still measuring' : short(p.xpPerHour) + ' xp/h']);
    const ax = mean(p.xpSeries);
    if (ax != null) out.push(['XP average', short(ax) + ' xp/h over the last ' + p.xpSeries.length + ' min']);
    out.push(['Gold now', p.goldPerHour == null ? 'still measuring' : short(p.goldPerHour) + ' gold/h']);
    const ag = mean(p.goldSeries);
    if (ag != null) out.push(['Gold average', short(ag) + ' gold/h over the last ' + p.goldSeries.length + ' min']);
    return out;
  }

  return {
    update(g) {
      const p = g && g.progress;
      if (!p) { ctx.hide(true); return; }
      ctx.hide(false);
      const lv = p.level || {};
      lvl.textContent = lv.from ? 'level ' + lv.from : 'no hero yet';
      to.textContent = lv.to ? 'to ' + lv.to : '';
      const pct = lv.pct == null ? null : Math.max(0, Math.min(100, lv.pct));
      if (pct == null) { meter.removeAttribute('value'); } else meter.setAttribute('value', String(Math.round(pct)));
      const e = eta(lv.etaSeconds);
      said.textContent = pct == null ? 'measuring' : Math.round(pct) + '%' + (e ? ', ' + e : ', measuring pace');
      said.title = pct == null ? WAIT : said.textContent;
      meter.setAttribute('aria-label', 'Progress to the next level');
      rate(xpv, p.xpPerHour, 'xp per hour');
      rate(gdv, p.goldPerHour, 'gold per hour');
      const pts = points(p.xpSeries);
      spark.setAttribute('class', pts.length ? 'pg-spark' : 'pg-spark pg-none');
      line.setAttribute('points', pts.map((q) => q.join(',')).join(' '));
      if (pts.length) { dot.setAttribute('cx', String(pts[pts.length - 1][0])); dot.setAttribute('cy', String(pts[pts.length - 1][1])); }
      ctx.ui.clear(pop);
      const dl = ctx.el('dl', 'pg-rows');
      for (const [k, t] of rows(p, lv)) dl.append(ctx.el('dt', '', k), ctx.el('dd', '', t));
      pop.append(dl);
      ctx.state('ready');
    },
  };
}
