// Long-lived requests of the video panel (web/panels/video/panel.js): the page keeps at most ONE video transport open (the palette
// feed OR the MJPEG stream, never both), the capability probe runs before any stream opens, a probe that stalls is aborted and
// falls back to MJPEG, and leaving the page closes the stream. A browser keeps about six connections per host; the event stream
// takes one, so every extra stream starves the panels' own fetches (the viewer showed NO SIGNAL).
// Usage: node video_conns.js path/to/panel.js
const assert = require('assert'), path = require('path');
const { load, shellCtx } = require('./panel_harness.js');
const dir = path.dirname(process.argv[2]);
const rest = Object.assign({}, load(path.join(dir, 'brightness.js')), load(path.join(dir, '../../streamstats.js')));
rest.drawStats = rest.draw;
const tick = (ms) => new Promise((r) => setTimeout(r, ms || 5));

// fetch stub that counts the requests still open: a stream stays open until its signal aborts; /api/glance and the probe answer
// at once unless stall is set, then they wait for their signal too
function env(stall) {
  const open = new Set(), log = [], players = [];
  const pending = (u, o) => new Promise((res, rej) => {
    const h = { u }; open.add(h);
    if (o && o.signal) o.signal.addEventListener('abort', () => { open.delete(h); rej(Object.assign(new Error('aborted'), { name: 'AbortError' })); });
  });
  const fetch = (u, o) => {
    log.push(u);
    if (/stream\.mjpg/.test(u)) return pending(u, o);
    if (stall) return pending(u, o);
    return Promise.resolve({ ok: true, json: () => Promise.resolve({ stream: { formats: ['mjpeg', 'palfeed'] } }) });
  };
  const PalFeed = {
    negotiate: (f, tq) => f('/api/glance' + tq).then(() => ({ paletteFeed: true, version: 1 })),
    fitFill: () => ({ cssW: 800, cssH: 600, scale: 1, integer: true }),
    Player: function (o) { const h = { u: o.url }; this.start = () => { open.add(h); players.push(h); }; this.stop = () => { open.delete(h); }; },
  };
  const wl = {};
  const addEventListener = (t, f) => { (wl[t] = wl[t] || []).push(f); };
  return { open, log, players, wl, extra: { PalFeed, fetch, addEventListener, AbortController, DecompressionStream: function () {}, ImageData: function () {}, console: { info() {} } } };
}
async function mount(stall, cfg) {
  const e = env(stall), timers = [];
  const web = { fetch: e.extra.fetch, ReadableStream: function () {}, AbortController, URL: Object.assign(function () {}, { createObjectURL() { return 'blob:x'; }, revokeObjectURL() {} }), Blob: function () {} };
  const M = Object.assign({}, load(path.join(dir, '../../mjpeg.js'), web), rest);
  const P = load(process.argv[2], Object.assign({}, M, e.extra));
  const s = shellCtx(Object.assign({ frameUrl: '/stream.mjpg', palFeedUrl: '/stream.pfeed' }, cfg || {}), {}, timers);
  let peak = 0; const tr = setInterval(() => { peak = Math.max(peak, e.open.size); }, 1);
  P.mount(s.root, s.c);
  await tick(40);
  clearInterval(tr);
  return Object.assign(e, { s, timers, peak: () => Math.max(peak, e.open.size) });
}

(async () => {
  // the probe answers: only the palette feed is open, the MJPEG stream was never requested
  let m = await mount(false);
  assert.strictEqual(m.players.length, 1, 'palette player');
  assert.ok(!m.log.some((u) => /stream\.mjpg/.test(u)), 'no MJPEG request next to the palette feed: ' + m.log.join(' '));
  assert.ok(m.peak() <= 1, 'peak open requests ' + m.peak());

  // the probe stalls (a busy host): while it waits no stream is open; after the timeout it is aborted and ONE MJPEG stream opens
  m = await mount(true);
  assert.ok(!m.log.some((u) => /stream\.mjpg/.test(u)), 'no MJPEG while the probe is pending');
  const t = m.timers.find((x) => x.ms === 6000);
  assert.ok(t, 'the probe has a timeout');
  t.fn(); await tick(40);
  const mj = [...m.open].filter((h) => /stream\.mjpg/.test(h.u)), probes = [...m.open].filter((h) => !/stream\.mjpg/.test(h.u));
  assert.strictEqual(mj.length, 1, 'one MJPEG stream after the timeout');
  assert.strictEqual(probes.length, 0, 'the stalled probe was aborted: ' + probes.map((h) => h.u));

  // leaving the page closes the stream
  assert.ok(m.wl.pagehide && m.wl.pagehide.length, 'pagehide handler');
  m.wl.pagehide.forEach((f) => f({}));
  await tick(10);
  assert.strictEqual(m.open.size, 0, 'no request open after pagehide');

  // no palette feed offered: one MJPEG stream at once
  m = await mount(false, { palFeedUrl: '' });
  assert.strictEqual(m.open.size, 1);
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
