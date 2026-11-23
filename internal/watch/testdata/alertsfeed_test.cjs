// web/alerts-feed.js (node test, card uialerts): off means no request and no toast; sound or granted notifications turn the feed on;
// only NEW alerts raise one toast; a cue only with the sound on; a desktop notification only with permission and a hidden tab.
'use strict';
const path = require('path');
const F = require(path.join(process.argv[2], 'alertsfeed.cjs'));
const fails = [], ok = (c, m) => { if (!c) fails.push(m); };
const tick = () => new Promise((r) => setTimeout(r, 0));
function rig(o) {
  const st = { reqs: [], toasts: [], notes: [], beeps: 0, kv: Object.assign({}, o.kv), alerts: [], seq: 0, timers: [] };
  class N { constructor(t, opt) { st.notes.push([t, opt.body]); } }
  N.permission = o.perm || 'default';
  class AC { constructor() { st.beeps++; this.currentTime = 0; this.destination = {}; } createOscillator() { return { frequency: {}, connect() {}, start() {}, stop() {} }; } createGain() { return { gain: {}, connect() {} }; } }
  const env = { doc: { hidden: !!o.hidden }, store: { get: (k) => st.kv[k] || null }, notif: o.noNotif ? null : N, audio: AC,
    toast: { show: (s, t) => st.toasts.push([s, t]) }, later: (fn) => { st.timers.push(fn); return st.timers.length; }, clear() {},
    fetch: (u) => { st.reqs.push(u); const since = Number(/since=(\d+)/.exec(u)[1]); return Promise.resolve({ ok: true, json: () => Promise.resolve({ seq: st.seq, alerts: st.alerts.filter((a) => a.seq > since) }) }); } };
  return { st, env, N };
}
const add = (st, text, sev) => { st.seq++; st.alerts.push({ seq: st.seq, text, severity: sev || 'warn' }); };
(async () => {
  { // both off: nothing is fetched, nothing shown
    const r = rig({}); add(r.st, 'old'); F.start(r.env); await tick(); add(r.st, 'new');
    for (let i = 0; i < 3; i++) { const f = r.st.timers.shift(); if (f) f(); await tick(); }
    ok(r.st.reqs.length === 0 && r.st.toasts.length === 0 && r.st.beeps === 0 && r.st.notes.length === 0, 'preference off: no request, toast, cue or notification');
  }
  { // sound on: the first answer is the baseline; a new alert raises one toast and one cue; a repeat raises nothing
    const r = rig({ kv: { 'd2lod.alerts.sound': '1' } }); add(r.st, 'old one'); F.start(r.env); await tick(); await tick();
    ok(r.st.toasts.length === 0, 'the alerts already there are not replayed');
    add(r.st, 'Life under 25%.', 'alarm'); r.st.timers.shift()(); await tick(); await tick();
    ok(r.st.toasts.length === 1 && r.st.toasts[0][0] === 'alarm' && r.st.toasts[0][1] === 'Life under 25%.', 'a new alert raises one toast with its severity');
    ok(r.st.beeps === 1, 'and one cue');
    r.st.timers.shift()(); await tick(); await tick();
    ok(r.st.toasts.length === 1 && /since=2/.test(r.st.reqs[r.st.reqs.length - 1]), 'a poll with nothing new raises nothing; it asks since the last seq');
    ok(r.st.notes.length === 0, 'no desktop notification without permission');
  }
  { // notifications granted, sound off: toast and notification only with the tab hidden, no cue
    const r = rig({ perm: 'granted', hidden: true }); F.start(r.env); await tick(); await tick();
    add(r.st, 'Short Staff is broken.'); r.st.timers.shift()(); await tick(); await tick();
    ok(r.st.toasts.length === 1 && r.st.notes.length === 1 && r.st.notes[0][1] === 'Short Staff is broken.' && r.st.beeps === 0, 'granted: a toast and a notification, no cue');
    const v = rig({ perm: 'granted', hidden: false }); F.start(v.env); await tick(); await tick(); add(v.st, 'x'); v.st.timers.shift()(); await tick(); await tick();
    ok(v.st.notes.length === 0 && v.st.toasts.length === 1, 'tab visible: no desktop notification');
    const d = rig({ perm: 'denied', hidden: true }); F.start(d.env); await tick();
    ok(d.st.reqs.length === 0, 'denied and sound off: the feed stays off');
  }
  { // turning the sound off stops the requests; on again starts from the current alert
    const r = rig({ kv: { 'd2lod.alerts.sound': '1' } }); F.start(r.env); await tick(); await tick();
    r.st.kv['d2lod.alerts.sound'] = '0'; const n = r.st.reqs.length; add(r.st, 'while off'); r.st.timers.shift()(); await tick(); await tick();
    ok(r.st.reqs.length === n && r.st.toasts.length === 0, 'sound switched off: no request');
    r.st.kv['d2lod.alerts.sound'] = '1'; r.st.timers.shift()(); await tick(); await tick(); r.st.timers.shift()();  await tick(); await tick();
    ok(r.st.toasts.length === 0, 'switched on again: what came while off is not replayed');
  }
  console.log(JSON.stringify({ fails })); process.exit(0);
})();
