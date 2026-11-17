// "Journey" cell of the cockpit strip (J-second-attempt.md section 2): the chapters as a dotted path (done, current, held,
// to come), the next chapter with how ready the hero is in one line, and the buttons the reply names (Go now, Hold). A
// click on a dot opens the chapter list in a popover. From GET /api/glance (journey). A button posts exactly the `post`
// and `body` of its action; the old Journey panel is not touched. textContent only.
const WORD = { done: 'done', open: 'current', held: 'held', skipped: 'skipped', todo: 'to come' };

export function stateWord(s) { return WORD[s] || 'to come'; }

// current is the index of the chapter the hero is on: the first open or held one, else the first still to come, else the last.
export function current(chs) {
  let i = chs.findIndex((c) => c.state === 'open' || c.state === 'held');
  if (i < 0) i = chs.findIndex((c) => c.state === 'todo');
  return i < 0 ? chs.length - 1 : i;
}

// line is the one line under the path as [title, band]: "Blood Raven", "too early (floor 9)".
export function line(j) {
  const n = j.next;
  if (!n) return [j.chapters.length && j.chapters.every((c) => c.state === 'done' || c.state === 'skipped') ? 'Every chapter is done' : 'No next chapter yet', ''];
  return [n.title || '', n.band || ''];
}

export function mount(root, ctx) {
  const head = ctx.el('span', 'jr-h'), prog = ctx.el('em', '');
  head.append(ctx.el('b', '', 'Journey'), prog);
  const path = ctx.el('div', 'jr-path');
  path.setAttribute('role', 'group');
  path.setAttribute('aria-label', 'Chapters');
  const sub = ctx.el('span', 'jr-sub');
  const acts = ctx.el('span', 'jr-acts');
  const row = ctx.el('div', 'jr-row');
  row.append(sub, acts);
  const box = ctx.el('div', 'jr-box');
  box.append(head, path, row);
  const pop = ctx.el('div', 'jr-pop');
  pop.setAttribute('popover', 'auto');
  pop.setAttribute('role', 'dialog');
  pop.setAttribute('aria-label', 'Journey chapters');
  ctx.body.append(box, pop);
  const toggle = () => { if (pop.togglePopover) pop.togglePopover(); };

  function act(a, b) {
    b.disabled = true;
    ctx.post(a.post, a.body).then((j) => {
      if (j && j.ok === false) ctx.toast('warn', j.error || 'could not do that');
      b.disabled = false;
    });
  }

  function absent(text) {
    prog.textContent = '';
    ctx.ui.clear(path);
    ctx.ui.clear(acts);
    ctx.ui.clear(pop);
    ctx.ui.clear(sub);
    sub.append(ctx.el('span', '', text));
    sub.title = text;
    box.dataset.empty = '1';
  }

  let sig = ''; // the last drawn journey: an identical poll leaves the dots (and a focused one) alone
  return {
    update(g) {
      if (!g || g.questBoard) { sig = ''; ctx.hide(true); return; } // with the quest board on, the quest tile stands in this cell
      ctx.hide(false);
      const j = g.journey;
      const now = JSON.stringify(j || null);
      if (now === sig) return;
      sig = now;
      if (!j || !Array.isArray(j.chapters) || !j.chapters.length) { absent('not reported'); ctx.state('ready'); return; }
      delete box.dataset.empty;
      const chs = j.chapters, cur = current(chs);
      prog.textContent = 'chapter ' + (j.n || cur + 1) + ' of ' + (j.of || chs.length);
      ctx.ui.clear(path);
      chs.forEach((c, i) => {
        if (i) { const l = ctx.el('i', ''); if (chs[i - 1].state === 'done') l.dataset.s = 'done'; path.append(l); }
        const d = ctx.el('button', 'jr-dot', String(i + 1));
        d.type = 'button';
        d.dataset.s = c.state;
        if (i === cur) d.dataset.cur = '1';
        d.setAttribute('aria-label', 'Chapter ' + (i + 1) + ', ' + stateWord(c.state));
        d.setAttribute('aria-haspopup', 'true');
        d.addEventListener('click', toggle);
        path.append(d);
      });
      const [ttl, band] = line(j), text = [ttl, band].filter(Boolean).join(': ');
      ctx.ui.clear(sub);
      sub.append(ctx.el('b', '', ttl));
      if (band) sub.append(ctx.el('span', '', ': ' + band));
      sub.title = j.next && j.next.text ? text + ' - ' + j.next.text : text;
      sub.dataset.band = j.next && j.next.band ? j.next.band : '';
      ctx.ui.clear(acts);
      (j.actions || []).slice(0, 2).forEach((a, i) => {
        const b = ctx.el('button', i === 0 ? 'jr-btn jr-p' : 'jr-btn', a.label);
        b.type = 'button';
        b.dataset.key = 'act-' + i;
        b.addEventListener('click', () => act(a, b));
        acts.append(b);
      });
      ctx.ui.clear(pop);
      const dl = ctx.el('dl', 'jr-rows');
      chs.forEach((c, i) => {
        dl.append(ctx.el('dt', '', String(i + 1)), ctx.el('dd', '', c.id + ' (' + stateWord(c.state) + ')'));
      });
      pop.append(dl);
      if (j.next && j.next.text) pop.append(ctx.el('p', 'jr-next', j.next.text));
      // the same verbs in the popover: on a narrow cell the strip keeps only the first one (the cell is under 270 px)
      const pa = ctx.el('span', 'jr-acts jr-pacts');
      (j.actions || []).slice(0, 2).forEach((a, i) => {
        const b = ctx.el('button', i === 0 ? 'jr-btn jr-p' : 'jr-btn', a.label);
        b.type = 'button';
        b.dataset.key = 'pact-' + i;
        b.addEventListener('click', () => { act(a, b); if (pop.hidePopover) pop.hidePopover(); });
        pa.append(b);
      });
      if (pa.children.length) pop.append(pa);
      ctx.state('ready');
    },
  };
}
