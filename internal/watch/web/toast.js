// The one toast (card uitoast, docs/ui-design-system.md "Toasts and severity"): a short message that goes by itself, with a Dismiss
// button or one action (Undo). Every page uses it: the shell panels, the Vault, the live screen, the alert feed.
// ONE severity scale for toast, note and chip: info, good, warn, bad. Old names map in (alarm and danger to bad, ok to good).
// ONE table of auto-dismiss times. The stack sits in #toasts (role=status, aria-live=polite; a bad toast is role=alert, announced at once).
// Text goes through textContent only. Classic script: window.dToast, or module.exports for the node test.
(function (g) {
  'use strict';
  const SEVERITIES = ['info', 'good', 'warn', 'bad'];
  const ALIAS = { alarm: 'bad', danger: 'bad', error: 'bad', ok: 'good', warning: 'warn' };
  const TIMES = { info: 6000, good: 6000, warn: 10000, bad: 20000, action: 8000 }; // ms; a toast with an action uses 'action' unless it says ms
  const MAX = 4; // the stack keeps this many, oldest first out

  const sev = (s) => { const v = ALIAS[s] || s; return SEVERITIES.includes(v) ? v : 'info'; };

  // create(doc, { host, later, clear, dismissText }): a stack on doc. later(fn, ms) and clear(handle) default to the timers.
  function create(doc, o) {
    o = o || {};
    const later = o.later || ((fn, ms) => setTimeout(fn, ms));
    const clear = o.clear || ((h) => clearTimeout(h));
    function host() {
      if (o.host) return o.host;
      const h = doc.getElementById && doc.getElementById('toasts');
      if (h) return h;
      const b = doc.createElement('div');
      b.id = 'toasts'; b.className = 'd-toasts';
      b.setAttribute('role', 'status'); b.setAttribute('aria-live', 'polite');
      doc.body.append(b);
      return b;
    }
    // show(sev, text, act): act = { label, run, ms } adds one action button that runs once; returns the toast node
    function show(s, text, act) {
      const t = doc.createElement('div');
      t.className = 'd-toast';
      t.dataset.sev = sev(s);
      t.setAttribute('role', sev(s) === 'bad' ? 'alert' : 'status');
      const msg = doc.createElement('span'); msg.textContent = text; t.append(msg);
      const b = doc.createElement('button');
      b.className = 'd-btn'; b.type = 'button'; b.textContent = act ? act.label : (o.dismissText || 'Dismiss');
      b.dataset.size = 'sm'; b.dataset.variant = 'ghost'; b.dataset.key = 'toast-dismiss';
      const box = host();
      let timer = null;
      const detach = (n) => { n.hidden = true; if (n.remove) n.remove(); else if (Array.isArray(box.children) && box.children.includes(n)) box.children.splice(box.children.indexOf(n), 1); }; // the else: a stub DOM
      const remove = () => { if (timer !== null) { clear(timer); timer = null; } detach(t); };
      b.addEventListener('click', () => { remove(); if (act && act.run) act.run(); });
      t.append(b);
      box.append(t);
      const mine = () => [...box.children].filter((n) => n.className === 'd-toast');
      while (mine().length > MAX) { const old = mine()[0]; if (old.dismiss) old.dismiss(); else detach(old); }
      timer = later(remove, (act && act.ms) || (act ? TIMES.action : TIMES[sev(s)]));
      t.dismiss = remove;
      return t;
    }
    return { show, host };
  }

  let main = null;
  const api = { SEVERITIES, TIMES, MAX, sev, create, show(s, text, act) { main = main || create(g.document); return main.show(s, text, act); } };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else g.dToast = api;
})(typeof window !== 'undefined' ? window : globalThis);
