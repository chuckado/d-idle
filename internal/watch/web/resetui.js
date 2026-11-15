// The Reset control (docs/reset.md): one confirm sheet and one progress strip, shared by the "..." menu, the Orders More
// sheet and the Needs you chip (they all dispatch the d2:reset-ask event). The sheet says what Reset does and has two
// buttons; a third, quiet button (restart the bot) appears only after a failed reset. The strip shows "Resetting..." with
// the bot's own reason from GET /api/reset until the state is idle again, then a toast says done or failed. textContent only.
export const ROUTE = '/api/reset';
export const ASK = 'd2:reset-ask';
export const COPY = 'Leave the game and rejoin; the bot resumes its plan and orders.';
const POLL_MS = 2000;
const GIVE_UP_MS = 4 * 60 * 1000; // a hard reset restarts the process: polls fail for a while, then stop

// phase turns a reset view into what the strip shows: active (the strip is up), text, and failed.
export function phase(v) {
  if (!v || v.schema !== 'watch.reset/1') return { active: false, text: '', failed: false };
  if (v.state === 'resetting' || v.state === 'rejoining') {
    const w = v.hard ? 'Restarting the bot' : 'Resetting';
    return { active: true, text: w + '...' + (v.reason ? ' ' + v.reason : ''), failed: false };
  }
  if (v.state === 'failed') return { active: false, text: 'Reset failed: ' + (v.reason || 'the game did not come back'), failed: true };
  return { active: false, text: '', failed: false };
}

// refusal is the words of a refused press (the route's own code, never a status code).
export function refusal(j) {
  if (!j) return 'no answer from the hero';
  if (j.code === 'off') return 'Reset is not available on this hero yet';
  if (j.code === 'measured_run') return 'A measured run is never reset';
  return j.error || 'could not reset';
}

export function mountReset(d) {
  const doc = d.doc || document;
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const wrap = el('div', 'd-reset-sheet');
  wrap.id = 'reset-confirm';
  wrap.hidden = true;
  wrap.setAttribute('role', 'dialog');
  wrap.setAttribute('aria-modal', 'true');
  wrap.setAttribute('aria-labelledby', 'reset-title');
  const card = el('div', 'd-reset-card');
  const title = el('h2', '', 'Reset the hero?');
  title.id = 'reset-title';
  const copy = el('p', '', COPY);
  const err = el('p', 'd-reset-err', '');
  err.hidden = true;
  const row = el('div', 'd-reset-row');
  const go = d.ui.btn('Reset', { variant: 'primary' });
  go.id = 'reset-go';
  go.dataset.key = 'reset-go';
  const no = d.ui.btn('Cancel', { variant: 'ghost' });
  no.dataset.key = 'reset-cancel';
  const hard = d.ui.btn('Restart the bot instead', { size: 'sm', variant: 'ghost' });
  hard.dataset.key = 'reset-hard';
  hard.hidden = true;
  row.append(no, go);
  card.append(title, copy, err, row, hard);
  wrap.appendChild(card);
  const strip = el('div', 'd-reset-strip');
  strip.id = 'reset-progress';
  strip.hidden = true;
  strip.setAttribute('role', 'status');
  doc.body.append(wrap, strip);

  let opener = null, timer = null, watching = false, since = 0, lastState = 'idle';
  const close = () => { wrap.hidden = true; if (opener && opener.focus) opener.focus({ preventScroll: true }); opener = null; };
  const open = (from) => {
    opener = from || (doc.activeElement && doc.activeElement !== doc.body ? doc.activeElement : null);
    err.hidden = true;
    wrap.hidden = false;
    go.disabled = false;
    go.focus();
  };

  function draw(v) {
    const p = phase(v);
    strip.hidden = !p.active;
    if (p.active) strip.textContent = p.text;
    hard.hidden = !p.failed;
    if (p.failed) { err.hidden = false; err.textContent = p.text; }
    const st = v && v.state ? v.state : 'idle';
    if (lastState !== st && watching) {
      if (st === 'idle' && v.last) d.toast(v.last.result === 'done' ? 'good' : 'warn', v.last.result === 'done' ? 'Reset done: back in the game, carrying on with the plan' : 'Reset failed: ' + (v.last.reason || ''));
      if (st === 'failed') d.toast('warn', p.text);
    }
    if (st === 'resetting' || st === 'rejoining') watching = true;
    else if (st === 'idle' || st === 'failed') watching = false;
    lastState = st;
    return st === 'resetting' || st === 'rejoining';
  }

  function poll() {
    timer = null;
    d.get(ROUTE).then((v) => {
      const busy = draw(v);
      if (busy || (!v && watching)) {
        if (!since) since = Date.now();
        if (Date.now() - since < GIVE_UP_MS) timer = d.later(poll, POLL_MS);
        else { watching = false; strip.hidden = true; }
      } else since = 0;
    });
  }
  const kick = () => { watching = true; since = 0; if (!timer) poll(); };

  function press(hardOn) {
    go.disabled = hard.disabled = true;
    d.post(ROUTE, hardOn ? { hard: true } : {}).then((j) => {
      go.disabled = hard.disabled = false;
      if (j && j.schema === 'watch.reset/1') { wrap.hidden = true; opener = null; draw(j); kick(); return; }
      err.hidden = false;
      err.textContent = refusal(j);
    });
  }
  go.addEventListener('click', () => press(false));
  hard.addEventListener('click', () => press(true));
  no.addEventListener('click', close);
  wrap.addEventListener('click', (e) => { if (e.target === wrap) close(); });
  wrap.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.preventDefault(); close(); return; }
    if (e.key === 'Tab') { // two or three buttons: keep the focus inside the sheet
      const f = [no, go, hard].filter((b) => !b.hidden && !b.disabled);
      if (!f.length) return;
      const i = f.indexOf(doc.activeElement), n = e.shiftKey ? (i <= 0 ? f.length - 1 : i - 1) : (i === f.length - 1 ? 0 : i + 1);
      e.preventDefault();
      f[n].focus();
    }
  });
  doc.addEventListener(ASK, (e) => open(e && e.detail && e.detail.from));
  poll(); // a reload during a reset shows the strip at once; an idle hero answers once and stops
  return { open, close, draw };
}
