// "Needs you" at the top of the cockpit rail (J-second-attempt.md section 2): at most three one-line rows, each with a
// shape by severity, the text with a muted sub text, and one or two buttons the reply names; a fourth line "+N more" and
// the "..." in the header open the full list in a popover, where the notification switch lives (off by default; a
// desktop notification for a NEW red row only). With no rows the panel is hidden and takes no height. From GET
// /api/glance (needs). A button posts exactly the `post` and `body` of its action. textContent only.
const RANK = { danger: 3, warn: 2, info: 1 };
const SHOWN = 3;
const SHOWN_SHORT = 2; // a short window (the cockpit's COCKPIT_SHORT, not a phone): Orders keeps its room

// ordered is the needs, worst first (stable inside a severity); a severity this page does not know counts as info.
export function ordered(needs) {
  const r = (n) => RANK[n.sev] || 1;
  return (needs || []).map((n, i) => [n, i]).sort((a, b) => r(b[0]) - r(a[0]) || a[1] - b[1]).map((x) => x[0]);
}

export function overflow(needs, shown = SHOWN) { return Math.max(0, (needs || []).length - shown); }

export function sevOf(n) { return RANK[n.sev] ? n.sev : 'info'; }

// fresh are the red rows of `needs` that were not red in `seen` (a Set of ids).
export function fresh(needs, seen) { return (needs || []).filter((n) => n.sev === 'danger' && !seen.has(n.id)); }

export function mount(root, ctx) {
  const head = ctx.el('div', 'n-head');
  const dots = ctx.el('button', 'n-dots', '...');
  dots.type = 'button';
  dots.dataset.key = 'menu';
  dots.setAttribute('aria-label', 'All needs and notification setting');
  dots.setAttribute('aria-haspopup', 'true');
  head.append(ctx.el('span', '', 'Needs you'), dots);
  const list = ctx.el('div', 'n-list');
  list.setAttribute('data-scroll', ''); // a long question or five answers on a short box scroll inside the list
  const pop = ctx.el('div', 'n-pop');
  pop.setAttribute('popover', 'auto');
  pop.setAttribute('role', 'dialog');
  pop.setAttribute('aria-label', 'Everything that needs you');
  ctx.body.append(head, list, pop);

  const notify = ctx.store('notify');
  let on = notify.get() === '1';
  let sig = ''; // the last drawn reply: an identical poll leaves the rows (and a focused button) alone
  let seen = null; // red ids of the last reply; null until the first one (a page that opens on a red row does not ping)
  const toggle = () => { if (pop.togglePopover) pop.togglePopover(); };
  dots.addEventListener('click', toggle);

  function act(a, btn) {
    if (a.post === '/api/reset') { document.dispatchEvent(new CustomEvent('d2:reset-ask', { detail: { from: btn } })); return; } // the confirm sheet decides
    btn.disabled = true;
    ctx.post(a.post, a.body).then((j) => {
      if (j && j.ok === false) ctx.toast('warn', j.error || 'could not do that');
      btn.disabled = false;
    });
  }

  function acts(n) {
    const box = ctx.el('span', 'n-acts');
    // a hero question shows every answer (some points have five); the other needs keep their two buttons
    (n.actions || []).slice(0, n.kind === 'question' ? 8 : 2).forEach((a, i) => {
      const b = ctx.el('button', i === 0 ? 'n-btn n-p' : 'n-btn', a.label);
      b.type = 'button';
      b.dataset.key = 'act-' + n.id + '-' + i;
      b.addEventListener('click', () => act(a, b));
      box.append(b);
    });
    return box;
  }

  function row(n) {
    const e = ctx.el('div', 'n-row');
    e.dataset.sev = sevOf(n);
    e.dataset.id = n.id;
    if (n.kind === 'question') e.dataset.q = ctx.short && ctx.short() ? 's' : '1'; // two lines of text (one on a short window), the answers on a line of their own
    e.title = n.text + (n.sub ? ' - ' + n.sub : '');
    const t = ctx.el('span', 'n-t');
    t.append(ctx.el('b', '', n.text));
    if (n.sub) t.append(ctx.el('span', 'n-s', ' ' + n.sub));
    e.append(ctx.el('span', 'n-shape'), t, acts(n));
    return e;
  }

  function sw() {
    const b = ctx.el('button', 'n-sw', 'Desktop notification for a new red alert');
    b.type = 'button';
    b.dataset.key = 'notify';
    b.setAttribute('role', 'switch');
    b.setAttribute('aria-checked', String(on));
    b.dataset.on = on ? '1' : '0';
    b.append(ctx.el('em', '', on ? 'on' : 'off'));
    b.addEventListener('click', () => {
      on = !on;
      notify.set(on ? '1' : '0');
      b.setAttribute('aria-checked', String(on));
      b.dataset.on = on ? '1' : '0';
      b.lastChild.textContent = on ? 'on' : 'off';
      if (on && typeof Notification !== 'undefined' && Notification.permission === 'default') Notification.requestPermission();
    });
    return b;
  }

  let seenQ = null; // question ids of the last reply: a new one is announced once, like a new red row
  function pingQuestions(needs) {
    const now = new Set(needs.filter((n) => n.kind === 'question').map((n) => n.id));
    if (seenQ) for (const n of needs) if (n.kind === 'question' && !seenQ.has(n.id)) ctx.toast('info', 'The hero asks: ' + n.text);
    seenQ = now;
  }

  function ping(needs) {
    const now = new Set(needs.filter((n) => n.sev === 'danger').map((n) => n.id));
    if (seen && on && typeof Notification !== 'undefined' && Notification.permission === 'granted') {
      for (const n of fresh(needs, seen)) new Notification(n.text, { body: n.sub || '', tag: n.id });
    }
    seen = now;
  }

  return {
    update(v) {
      if (!v || !Array.isArray(v.needs)) { sig = ''; ctx.hide(true); return; }
      const needs = ordered(v.needs);
      ping(needs);
      pingQuestions(needs);
      const shown = ctx.short && ctx.short() ? SHOWN_SHORT : SHOWN; // a short window shows two rows: Orders keeps its room
      const now = JSON.stringify([needs, shown]);
      if (now === sig) return;
      sig = now;
      ctx.ui.clear(list);
      ctx.ui.clear(pop);
      if (!needs.length) { ctx.hide(true); ctx.state('ready'); return; }
      ctx.hide(false);
      needs.slice(0, shown).forEach((n) => list.append(row(n)));
      // the "+N more" link takes the place of the "..." in the header: the same popover, no line of its own
      const extra = overflow(needs, shown);
      dots.textContent = extra ? '+' + extra + ' more' : '...';
      dots.dataset.more = extra ? '1' : '';
      const all = ctx.el('div', 'n-all');
      needs.forEach((n) => all.append(row(n)));
      pop.append(all, sw());
      ctx.state('ready');
    },
  };
}
