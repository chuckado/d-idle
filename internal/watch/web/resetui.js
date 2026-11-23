// The Reset control (docs/reset.md): a break-glass, so ONE TAP. The always-visible header button, the "..." menu entry, the
// Orders More entry and the Needs you chip (the last three dispatch d2:reset-ask) all send the soft reset at once: no
// confirm, feedback within 100 ms (the button says Resetting..., the strip says Reset sent and then shows the bot's reason), a 3 s cooldown against a double press. A confirm sheet exists only for the hard variant (restart the process),
// reached from the strip after a failed reset. textContent only.
export const ROUTE = '/api/reset';
export const ASK = 'd2:reset-ask';
export const COPY = 'Restart the bot process; it picks up its plan and orders again (about half a minute).';
export const COOLDOWN_MS = 3000;
export const LABEL = 'Reset bot (leaves the game and rejoins)';
const STUCK = 'the bot was stuck, restarting it'; // the server's words for an escalated reset (internal/bot/playerreset.go)
const POLL_MS = 2000;
const GIVE_UP_MS = 4 * 60 * 1000; // a hard reset restarts the process: polls fail for a while, then stop

// phase turns a reset view into what the strip shows: active (the strip is up), text, and failed.
export function phase(v) {
  if (!v || v.schema !== 'watch.reset/1') return { active: false, text: '', failed: false };
  if (v.state === 'resetting' || v.state === 'rejoining') {
    if (v.hard && v.reason === STUCK) return { active: true, text: 'Resetting... the bot was stuck, restarting it', failed: false };
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
  const now = d.now || (() => Date.now());
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const wrap = el('div', 'd-reset-sheet');
  wrap.id = 'reset-confirm';
  wrap.hidden = true;
  wrap.setAttribute('role', 'dialog');
  wrap.setAttribute('aria-modal', 'true');
  wrap.setAttribute('aria-labelledby', 'reset-title');
  const card = el('div', 'd-reset-card');
  const title = el('h2', '', 'Restart the bot?');
  title.id = 'reset-title';
  const copy = el('p', '', COPY);
  const err = el('p', 'd-reset-err', '');
  err.hidden = true;
  const row = el('div', 'd-reset-row');
  const go = d.ui.btn('Restart', { variant: 'primary' });
  go.id = 'reset-go';
  go.dataset.key = 'reset-go';
  const no = d.ui.btn('Cancel', { variant: 'ghost' });
  no.dataset.key = 'reset-cancel';
  row.append(no, go);
  card.append(title, copy, err, row);
  wrap.appendChild(card);
  const strip = el('div', 'd-reset-strip');
  strip.id = 'reset-progress';
  strip.hidden = true;
  strip.setAttribute('role', 'status');
  const stripText = el('span', 'd-reset-text', '');
  const hard = d.ui.btn('Restart the bot', { size: 'sm', variant: 'ghost' });
  hard.dataset.key = 'reset-hard';
  hard.hidden = true;
  strip.append(stripText, hard);
  doc.body.append(wrap, strip);
  const btn = d.button || null; // the header button (shell.js): always visible, one tap
  if (btn) btn.setAttribute('aria-label', LABEL);

  let opener = null, timer = null, watching = false, since = 0, lastState = 'idle', lastPress = -Infinity, sending = false;
  const close = () => { wrap.hidden = true; if (opener && opener.focus) opener.focus({ preventScroll: true }); opener = null; };
  const openHard = (from) => {
    opener = from || (doc.activeElement && doc.activeElement !== doc.body ? doc.activeElement : null);
    err.hidden = true;
    wrap.hidden = false;
    go.disabled = false;
    go.focus();
  };
  const label = (busy) => { if (btn) { btn.textContent = busy ? 'Resetting...' : 'Reset'; btn.dataset.busy = busy ? '1' : ''; btn.setAttribute('aria-busy', busy ? 'true' : 'false'); } };

  function draw(v) {
    const p = phase(v);
    const st = v && v.state ? v.state : 'idle';
    strip.hidden = !(p.active || p.failed);
    if (!strip.hidden) stripText.textContent = p.text;
    hard.hidden = !p.failed;
    label(p.active);
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
        else { watching = false; strip.hidden = true; label(false); }
      } else since = 0;
    });
  }
  const kick = () => { watching = true; since = 0; if (!timer) poll(); };

  // press sends the soft reset at once. Everything the player sees happens before the request returns.
  function press() {
    const t = now();
    if (sending || t - lastPress < COOLDOWN_MS) return; // idempotent: a second press while resetting is a no-op
    lastPress = t;
    sending = true;
    label(true);
    stripText.textContent = 'Reset sent. Resetting... leaving the game'; // no toast: toasts sit top right, over this very button (fit caught it)
    strip.hidden = false;
    hard.hidden = true;
    d.post(ROUTE, {}).then((j) => {
      sending = false;
      if (j && j.schema === 'watch.reset/1') { draw(j); kick(); return; }
      lastPress = -Infinity;
      label(false);
      strip.hidden = true;
      d.toast('warn', refusal(j));
    });
  }
  function pressHard() {
    go.disabled = true;
    d.post(ROUTE, { hard: true }).then((j) => {
      go.disabled = false;
      if (j && j.schema === 'watch.reset/1') { wrap.hidden = true; opener = null; lastPress = now(); draw(j); kick(); return; }
      err.hidden = false;
      err.textContent = refusal(j);
    });
  }
  go.addEventListener('click', pressHard);
  hard.addEventListener('click', () => openHard(hard));
  no.addEventListener('click', close);
  if (btn) btn.addEventListener('click', press);
  wrap.addEventListener('click', (e) => { if (e.target === wrap) close(); });
  wrap.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.preventDefault(); close(); return; }
    if (e.key === 'Tab') { // two buttons: keep the focus inside the sheet
      const f = [no, go].filter((b) => !b.disabled);
      if (!f.length) return;
      const i = f.indexOf(doc.activeElement), n = e.shiftKey ? (i <= 0 ? f.length - 1 : i - 1) : (i === f.length - 1 ? 0 : i + 1);
      e.preventDefault();
      f[n].focus();
    }
  });
  doc.addEventListener(ASK, () => press());
  poll(); // a reload during a reset shows the strip at once; an idle hero answers once and stops
  return { press, openHard, close, draw };
}
