// The support bundle sheet (docs/support-bundle.md): opened by the "..." menu (the d2:support-ask event). Three steps in
// one dialog: the choices (screenshot and character save, both off), the preview of exactly what would be sent (files,
// sizes, the log excerpt, anything the scan blocked) with Send or Save to file instead, and the support code. Nothing is
// sent before the press of Send. textContent only; every string from the server is data.
export const ROUTE = '/api/support';
export const BUNDLE = '/api/support/bundle';
export const SEND = '/api/support/send';
export const DOWNLOAD = '/api/support/download';
export const ASK = 'd2:support-ask';
export const SCHEMA = 'watch.support/1';
export const INTRO = 'This collects the bot log of the last hours, settings, health and version numbers, with home paths, addresses, names and keys removed. You see everything before anything is sent.';
export const SAVE_WORDS = 'This is your character save file: your character, its items and its name. Leave it off unless the owner asks for it.';

export function size(n) {
  if (n < 1024) return n + ' B';
  if (n < 1048576) return (n / 1024).toFixed(1) + ' KB';
  return (n / 1048576).toFixed(1) + ' MB';
}

// refusal is the words of a refused call (the route's own code, never a status code).
export function refusal(j) {
  if (!j) return 'no answer from the hero';
  if (j.code === 'off') return 'Support bundles are not available on this hero yet';
  if (j.code === 'gone') return 'That bundle expired: build it again';
  if (j.code === 'blocked') return 'This bundle still holds private-looking text, so it is not sent';
  return j.error || 'could not do that';
}

export function mountSupport(d) {
  const doc = d.doc || document;
  const el = (tag, cls, text) => { const e = doc.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const wrap = el('div', 'd-sup-sheet');
  wrap.id = 'support-sheet';
  wrap.hidden = true;
  wrap.setAttribute('role', 'dialog');
  wrap.setAttribute('aria-modal', 'true');
  wrap.setAttribute('aria-labelledby', 'support-title');
  const card = el('div', 'd-sup-card');
  const title = el('h2', '', 'Send a support bundle');
  title.id = 'support-title';
  const intro = el('p', '', INTRO);
  const err = el('p', 'd-sup-err', '');
  err.hidden = true;

  const form = el('div', 'd-sup-form');
  const shot = el('input'); shot.type = 'checkbox'; shot.id = 'support-shot'; shot.dataset.key = 'support-shot';
  const shotL = el('label', 'd-sup-check'); shotL.append(shot, el('span', '', 'Include a screenshot of the game view'));
  const save = el('input'); save.type = 'checkbox'; save.id = 'support-save'; save.dataset.key = 'support-save';
  const saveL = el('label', 'd-sup-check'); saveL.append(save, el('span', '', 'Include my character save file'));
  const saveNote = el('p', 'd-sup-note', SAVE_WORDS);
  form.append(shotL, saveL, saveNote);

  const prev = el('div', 'd-sup-prev');
  prev.hidden = true;
  const files = el('ul', 'd-sup-files');
  const logH = el('p', 'd-sup-note', 'The end of the log, as it will be sent:');
  const log = el('pre', 'd-sup-log');
  const blocked = el('div', 'd-sup-blocked');
  prev.append(files, logH, log, blocked);

  const done = el('div', 'd-sup-done');
  done.hidden = true;
  const doneMsg = el('p', '', 'Sent. Your support code is:');
  const codeEl = el('p', 'd-sup-code', '');
  codeEl.id = 'support-code';
  const doneNote = el('p', 'd-sup-note', 'Send this code to the owner. Nothing else was uploaded.');
  done.append(doneMsg, codeEl, doneNote);

  const row = el('div', 'd-sup-row');
  const no = d.ui.btn('Cancel', { variant: 'ghost' });
  no.dataset.key = 'support-cancel';
  const build = d.ui.btn('Preview', { variant: 'primary' });
  build.id = 'support-build';
  build.dataset.key = 'support-build';
  const send = d.ui.btn('Send', { variant: 'primary' });
  send.id = 'support-send';
  send.dataset.key = 'support-send';
  const saveBtn = el('a', 'd-sup-link', 'Save to file instead');
  saveBtn.id = 'support-save-file';
  saveBtn.dataset.key = 'support-save-file';
  const back = d.ui.btn('Back', { variant: 'ghost', size: 'sm' });
  back.dataset.key = 'support-back';
  row.append(no, build, send);
  const alt = el('div', 'd-sup-alt');
  alt.append(saveBtn, back);
  card.append(title, intro, err, form, prev, done, row, alt);
  wrap.appendChild(card);
  doc.body.appendChild(wrap);

  let opener = null, id = '', step = 'form';
  const show = (s) => {
    step = s;
    form.hidden = s !== 'form';
    prev.hidden = s !== 'prev';
    done.hidden = s !== 'done';
    build.hidden = s !== 'form';
    send.hidden = s !== 'prev' || !state.canSend;
    saveBtn.hidden = s !== 'prev' || !state.canSave;
    back.hidden = s !== 'prev';
    alt.hidden = s !== 'prev';
    intro.hidden = s === 'done';
    no.textContent = s === 'done' ? 'Close' : 'Cancel';
  };
  const state = { canSend: false, canSave: false };
  const close = () => { wrap.hidden = true; if (opener && opener.focus) opener.focus({ preventScroll: true }); opener = null; };
  const open = (from) => {
    opener = from || (doc.activeElement && doc.activeElement !== doc.body ? doc.activeElement : null);
    err.hidden = true;
    shot.checked = save.checked = false; // both start off every time
    id = '';
    state.canSend = state.canSave = false;
    show('form');
    wrap.hidden = false;
    build.disabled = false;
    build.focus();
  };
  const fail = (m) => { err.hidden = false; err.textContent = m; };

  function draw(p) {
    id = p.id;
    state.canSend = !!p.canSend;
    state.canSave = !!p.canSave;
    files.textContent = '';
    for (const f of p.files || []) files.appendChild(el('li', '', f.name + '  ' + size(f.size)));
    log.textContent = p.logPreview || '(no log lines)';
    blocked.textContent = '';
    const fs = p.findings || [];
    blocked.hidden = !fs.length;
    if (fs.length) {
      blocked.appendChild(el('p', 'd-sup-err', 'Not sent: this bundle still holds private-looking text (' + fs.length + '). Nothing leaves this computer.'));
      for (const f of fs.slice(0, 20)) blocked.appendChild(el('p', 'd-sup-note', f.file + ' line ' + f.line + ' (' + f.class + '): ' + f.context));
    } else if (!p.configured) {
      blocked.hidden = false;
      blocked.appendChild(el('p', 'd-sup-note', 'Uploading is not set up on this install. Save the file and send it to the owner yourself.'));
    }
    if (p.dropped && p.dropped.length) files.appendChild(el('li', 'd-sup-note', 'Left out: ' + p.dropped.join(', ')));
    saveBtn.setAttribute('href', DOWNLOAD + '?id=' + encodeURIComponent(p.id));
    saveBtn.setAttribute('download', '');
    show('prev');
    (state.canSend ? send : saveBtn.hidden ? back : saveBtn).focus();
  }

  build.addEventListener('click', () => {
    build.disabled = true;
    err.hidden = true;
    d.post(BUNDLE, { includeScreenshot: shot.checked, includeSave: save.checked }).then((j) => {
      build.disabled = false;
      if (j && j.schema === SCHEMA) draw(j); else fail(refusal(j));
    });
  });
  send.addEventListener('click', () => {
    send.disabled = true;
    err.hidden = true;
    d.post(SEND, { id }).then((j) => {
      send.disabled = false;
      if (j && j.schema === SCHEMA && j.code) {
        codeEl.textContent = j.code;
        show('done');
        no.focus();
        d.toast('good', 'Support bundle sent: code ' + j.code);
      } else fail(refusal(j) + (j && j.error && j.code === 'send' ? '' : ''));
    });
  });
  back.addEventListener('click', () => { err.hidden = true; show('form'); build.focus(); });
  no.addEventListener('click', close);
  wrap.addEventListener('click', (e) => { if (e.target === wrap) close(); });
  wrap.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.preventDefault(); close(); return; }
    if (e.key === 'Tab') {
      const f = [shot, save, no, build, send, saveBtn, back].filter((b) => !b.hidden && !b.disabled && !(b.parentNode && b.parentNode.hidden));
      if (!f.length) return;
      const i = f.indexOf(doc.activeElement), n = e.shiftKey ? (i <= 0 ? f.length - 1 : i - 1) : (i === f.length - 1 ? 0 : i + 1);
      e.preventDefault();
      f[n].focus();
    }
  });
  doc.addEventListener(ASK, (e) => open(e && e.detail && e.detail.from));
  return { open, close, get step() { return step; } };
}
