// The lossless-feed path of web/panels/video/panel.js against a stub DOM and a fake PalFeed: used only when /api/config
// has palFeedUrl, glance lists "palfeed" and the probe answers; a fatal stream error or three failures return to MJPEG.
// Usage: node video_pal_panel.js path/to/panel.js
const assert = require('assert'), { load, shellCtx } = require('./panel_harness.js');
const path = require('path');
const M = load(path.join(path.dirname(process.argv[2]), '../../mjpeg.js'));
const tick = (ms) => new Promise((r) => setTimeout(r, ms || 5));

function env(formats, detectOK) {
  const players = [], urls = [];
  const PalFeed = {
    detect: () => Promise.resolve(detectOK ? { paletteFeed: true, version: 1 } : null),
    fit: () => ({ cssW: 800, cssH: 600, scale: 1 }),
    Player: function (o) { this.o = o; this.stopped = false; players.push(this); urls.push(o.url); this.start = () => {}; this.stop = () => { this.stopped = true; }; },
  };
  const fetch = (u) => Promise.resolve({ ok: true, json: () => Promise.resolve({ stream: { formats } }) });
  return { players, urls, extra: { PalFeed, fetch, DecompressionStream: function () {}, ImageData: function () {}, console: { info() {} } } };
}
async function mount(formats, detectOK, cfg, store) {
  const e = env(formats, detectOK), timers = [];
  const P = load(process.argv[2], Object.assign({}, M, e.extra));
  const s = shellCtx(Object.assign({ frameUrl: '/stream.mjpg', palFeedUrl: '/stream.pfeed' }, cfg || {}), store || {}, timers);
  P.mount(s.root, s.c);
  await tick(20);
  const frame = s.c.body.children[0];
  return Object.assign(e, { s, frame, shot: frame.children[0], cv: frame.children[1], timers });
}

(async () => {
  // glance lists palfeed + probe ok: the canvas player opens with the token, the canvas shows on the first frame
  let m = await mount(['mjpeg', 'palfeed'], true);
  assert.strictEqual(m.players.length, 1, 'pal player opened');
  assert.ok(/^\/stream\.pfeed\?t=tok&r=\d+$/.test(m.urls[0]), m.urls[0]);
  assert.strictEqual(m.cv.hidden, true);
  m.players[0].o.onFrame();
  assert.strictEqual(m.cv.hidden, false); assert.strictEqual(m.shot.hidden, true); assert.strictEqual(m.s.c.st, 'ready');
  // a fatal stream error returns to MJPEG: the canvas hides and the next open is the fetch stream
  m.players[0].o.onEnd(Object.assign(new Error('palfeed: bad header'), { fatal: true }));
  assert.strictEqual(m.cv.hidden, true);
  await tick();
  assert.ok(m.players[0].stopped, 'player stopped');
  assert.strictEqual(m.players.length, 1, 'no second pal player after the fallback');
  assert.ok(/^\/stream\.mjpg\?t=tok/.test(m.shot.src) || m.s.c.st !== 'ready', 'MJPEG again');

  // three plain failures in a row do the same; a frame in between resets the count
  m = await mount(['mjpeg', 'palfeed'], true);
  m.players[0].o.onEnd(new Error('ended')); m.players[0].o.onEnd(new Error('ended'));
  m.players[0].o.onFrame(); // resets
  m.players[0].o.onEnd(new Error('ended')); m.players[0].o.onEnd(new Error('ended'));
  assert.strictEqual(m.cv.hidden, false, 'two failures after a frame keep the feed');
  m.players[0].o.onEnd(new Error('ended'));
  assert.strictEqual(m.cv.hidden, true, 'the third failure in a row falls back');

  // low rate rides on the feed's URL too
  m = await mount(['mjpeg', 'palfeed'], true, {}, { lowrate: '1' });
  assert.ok(/fps=8/.test(m.urls[0]), m.urls[0]);

  // not offered: formats without palfeed, no palFeedUrl, probe refused: MJPEG only, no player
  for (const [f, ok, cfg] of [[['mjpeg'], true, {}], [['mjpeg', 'palfeed'], true, { palFeedUrl: '' }]]) {
    m = await mount(f, ok, cfg);
    assert.strictEqual(m.players.length, 0);
    assert.ok(/^\/stream\.mjpg|^\/stream/.test(m.shot.src));
  }
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
