// The Labels button of web/panels/video/panel.js against a stub DOM: hidden until GET /api/labels says available, shows
// the state the game read back, a click POSTs the opposite and shows the answer, a refusal shows a toast and keeps the state.
// Usage: node video_labels_panel.js path/to/panel.js
const assert = require('assert'), path = require('path');
const { load, shellCtx, controlsOf } = require('./panel_harness.js');
const M0 = load(path.join(path.dirname(process.argv[2]), '../../mjpeg.js'));
const B = load(path.join(path.dirname(process.argv[2]), 'brightness.js'));
const S = load(path.join(path.dirname(process.argv[2]), '../../streamstats.js'));
const M = Object.assign({}, M0, S, B, { drawStats: S.draw });
const tick = (ms) => new Promise((r) => setTimeout(r, ms || 5));

async function mount(get) {
  const timers = [], posts = [], toasts = [];
  const fetch = (u) => /\/api\/labels/.test(u) ? Promise.resolve({ ok: true, json: () => Promise.resolve(get()) }) : Promise.resolve({ ok: true, json: () => Promise.resolve({ stream: { formats: ['mjpeg'] } }) });
  const P = load(process.argv[2], Object.assign({}, M, { fetch, console: { info() {} } }));
  const s = shellCtx({ frameUrl: '/stream.mjpg' }, {}, timers);
  s.c.get = (u) => fetch(u).then((r) => r.json());
  s.c.post = (p, b) => { posts.push([p, b]); return Promise.resolve(s.answer(b)); };
  s.c.toast = (k, t) => toasts.push([k, t]);
  P.mount(s.root, s.c);
  await tick(20);
  const btn = controlsOf(s.c).find((b) => b.dataset && b.dataset.key === 'labels');
  return { s, btn, posts, toasts, timers };
}

(async () => {
  let state = { ok: true, on: false, available: false };
  let m = await mount(() => state);
  assert.ok(m.btn, 'the button exists');
  assert.strictEqual(m.btn.hidden, true, 'hidden while the game cannot show labels');
  assert.strictEqual(m.btn.dataset.size, 'sm');

  state = { ok: true, on: true, available: true };
  m = await mount(() => state);
  assert.strictEqual(m.btn.hidden, false); assert.strictEqual(m.btn.attrs['aria-pressed'], 'true', 'shows the read-back state');

  state = { ok: true, on: false, available: true };
  m = await mount(() => state);
  assert.strictEqual(m.btn.attrs['aria-pressed'], 'false');
  m.s.answer = (b) => ({ ok: true, on: b.on, available: true });
  m.btn.fire('click'); await tick();
  assert.strictEqual(JSON.stringify(m.posts), '[["/api/labels",{"on":true}]]');
  assert.strictEqual(m.btn.attrs['aria-pressed'], 'true', 'the answer sets the state');
  m.btn.fire('click'); await tick();
  assert.strictEqual(JSON.stringify(m.posts[1]), '["/api/labels",{"on":false}]');
  assert.strictEqual(m.btn.attrs['aria-pressed'], 'false');

  // a refusal: the state stays what the game says and the reason is shown
  m.s.answer = () => ({ ok: false, on: false, available: true, code: 'refused', reason: 'the game did not take it' });
  m.btn.fire('click'); await tick();
  assert.strictEqual(m.btn.attrs['aria-pressed'], 'false'); assert.strictEqual(m.toasts.length, 1);
  assert.strictEqual(m.btn.disabled, false, 'enabled again after the answer');

  // the game goes away (unsupported after a swap): the button hides again at the next read
  m.s.answer = () => ({ ok: false, on: false, available: false, code: 'unsupported', reason: 'this game build cannot show item labels' });
  m.btn.fire('click'); await tick();
  assert.strictEqual(m.btn.hidden, true);
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
