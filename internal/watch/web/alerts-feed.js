// The alert feed of the new pages (card uialerts; review E10 gap): the pages after the pre-shell one had no feed that plays the stored
// Alerts settings (web/settings.js). Same source as the old page (web/alerts.js): GET /api/alerts?since=N every few seconds. While the
// sound setting is on, or the browser has granted desktop notifications, a NEW alert raises one toast (toast.js), a short cue and, with
// the tab hidden, a desktop notification. With both off nothing is fetched and nothing changes. The old page keeps its own
// alerts.js: this feed does not start where #alertspane exists. Text goes through textContent only.
// Classic script: starts itself in a browser; module.exports for the node test (internal/watch/testdata/alertsfeed_test.cjs).
(function (g) {
  'use strict';
  const SOUND_KEY = 'd2lod.alerts.sound'; // the key settings.js and alerts.js share
  const CUES = { info: [[520, 90]], good: [[520, 90], [780, 120]], warn: [[400, 140], [400, 140]], alarm: [[300, 200], [220, 320]] };
  const EVERY = 4000, GIVE_UP = 5;

  // env: { doc, fetch, store: {get}, notif (Notification or null), toast (dToast), audio (AudioContext class or null), later(fn, ms) }
  function start(env) {
    let seq = 0, first = true, fails = 0, stopped = false, timer = null;
    const soundOn = () => env.store.get(SOUND_KEY) === '1';
    const notifOn = () => !!env.notif && env.notif.permission === 'granted';
    const active = () => soundOn() || notifOn();

    function beep(sev) {
      const cue = CUES[sev];
      if (!env.audio || !cue || !soundOn()) return;
      try {
        const ac = new env.audio(); let t = ac.currentTime;
        for (const [hz, ms] of cue) {
          const o = ac.createOscillator(), gn = ac.createGain();
          o.frequency.value = hz; gn.gain.value = 0.05; o.connect(gn); gn.connect(ac.destination);
          o.start(t); o.stop(t + ms / 1000); t += ms / 1000;
        }
      } catch (e) { /* no audio here */ }
    }
    function note(a) {
      if (!notifOn() || !env.doc.hidden) return;
      try { new env.notif('d.idle', { body: a.text, tag: 'd2lod-' + a.seq }); } catch (e) { /* not allowed here */ }
    }
    function raise(a) { env.toast.show(a.severity, a.text); beep(a.severity); note(a); }

    function tick() {
      if (stopped) return Promise.resolve();
      if (!active()) { first = true; return Promise.resolve(); } // off: no request; coming back on starts from the then-current alert
      return env.fetch('/api/alerts?since=' + seq, { cache: 'no-store', credentials: 'same-origin' }).then((r) => { if (!r.ok) throw new Error(String(r.status)); return r.json(); })
        .then((v) => {
          fails = 0;
          for (const a of v.alerts || []) if (!first) raise(a);
          first = false; seq = v.seq || seq;
        }, () => { fails++; });
    }
    function loop() {
      tick().then(() => { if (!stopped && fails < GIVE_UP) timer = env.later(loop, EVERY); });
    }
    loop();
    return { stop() { stopped = true; if (env.clear && timer !== null) env.clear(timer); }, tick };
  }

  const api = { start, SOUND_KEY };
  if (typeof module !== 'undefined' && module.exports) { module.exports = api; return; }
  const d = g.document;
  if (!d || d.getElementById('alertspane') || !g.dToast) return; // the old page has its own feed
  start({ doc: d, fetch: g.fetch.bind(g), store: { get(k) { try { return g.localStorage.getItem(k); } catch (e) { return null; } } },
    notif: g.Notification || null, toast: g.dToast, audio: g.AudioContext || g.webkitAudioContext || null, later: (fn, ms) => setTimeout(fn, ms), clear: (h) => clearTimeout(h) });
})(typeof window !== 'undefined' ? window : globalThis);
