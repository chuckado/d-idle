// node --test tools/videofeed/streamstats.test.js: the stream monitor's recorder (ring buffers, window, hitches, events, JSON).
const test = require('node:test'), assert = require('node:assert'), fs = require('node:fs'), vm = require('node:vm'), path = require('node:path');
const src = fs.readFileSync(path.join(__dirname, '../../internal/watch/web/streamstats.js'), 'utf8').replace(/^export /gm, '');
const ctx = { Date, Math, Float64Array, Float32Array, Int32Array, Array, Object, JSON, Number, String };
vm.runInNewContext(src + '\n;this.X = { StreamMonitor, SlowGate, draw, PALETTE, WINDOW_S, HITCH_MS, HEIGHT_FULL, HEIGHT_COMPACT };', ctx);
const { StreamMonitor, SlowGate, draw, WINDOW_S, HITCH_MS, HEIGHT_FULL, HEIGHT_COMPACT } = ctx.X;
const mk = () => ({ deliv: [], shown: [], drop: [], kB: [], keys: [], delay: [], thr: [], cpu: [], ring: [] });
function clock(t0) { const c = { t: t0, now: () => c.t }; return c; }

test('per-second buckets over a 30 s window, old seconds fall out', () => {
  const c = clock(1_000_000_000_000), m = new StreamMonitor(c.now);
  for (let s = 0; s < 40; s++) { for (let k = 0; k < 30; k++) { c.t = 1_000_000_000_000 + s * 1000 + k * 33; m.arrive(1024, false); m.shownFrame(); } }
  c.t = 1_000_000_000_000 + 39 * 1000 + 990;
  const ser = m.series(mk());
  assert.strictEqual(ser.deliv.length, WINDOW_S);
  assert.ok(ser.deliv.every((v) => v >= 29 && v <= 31), 'about 30 a second: ' + ser.deliv.join(','));
  assert.ok(ser.kB.every((v) => v > 28 && v < 32));
  c.t += 5 * 60 * 1000; // five minutes of silence: the window is empty again
  const e = m.series(mk()); assert.ok(e.deliv.every((v) => v === 0));
});

test('a gap over 50 ms is a hitch, with an event and a count; p95 and fps summarise the window', () => {
  const c = clock(2_000_000_000_000), m = new StreamMonitor(c.now);
  for (let k = 0; k < 90; k++) { c.t += 33; m.shownFrame(); }
  c.t += 120; m.shownFrame();                  // one 120 ms stall
  for (let k = 0; k < 30; k++) { c.t += 33; m.shownFrame(); }
  const s = m.summary();
  assert.strictEqual(s.hitches, 1); assert.strictEqual(s.max, 120); assert.ok(s.p95 <= 34, 'p95 ' + s.p95);
  const ev = m.events([]); assert.ok(ev.length === 1 && ev[0].kind === 'hitch' && /120/.test(ev[0].label), JSON.stringify(ev));
  c.t += 31_000; assert.strictEqual(m.summary().hitches, 0, 'a hitch leaves the window after 30 s');
});

test('ring reuse never mixes seconds; delay averages per second; server samples are NaN until set', () => {
  const c = clock(3_000_000_000_000), m = new StreamMonitor(c.now);
  m.delay(40); m.delay(60); c.t += 1000; m.delay(100);
  m.server({ throttlePct: 12.5, cpuPct: 20, ringFps: 29, atMs: c.t });
  const ser = m.series(mk());
  assert.strictEqual(ser.delay[WINDOW_S - 2], 50); assert.strictEqual(ser.delay[WINDOW_S - 1], 100);
  assert.strictEqual(ser.thr[WINDOW_S - 1], 12.5); assert.ok(Number.isNaN(ser.thr[WINDOW_S - 2]));
});

test('json carries the whole timing context: ISO window, tz offset, player, build, server sample time, events', () => {
  const c = clock(Date.UTC(2026, 10, 16, 12, 0, 30)), m = new StreamMonitor(c.now);
  m.player = 'palette'; m.target = 30;
  for (let k = 0; k < 60; k++) { c.t += 33; m.arrive(5000, k === 10); m.shownFrame(); }
  m.event('resync', 'need key'); m.server({ throttlePct: 3, cpuPct: 9, ringFps: 29, atMs: c.t - 40 });
  const j = JSON.parse(m.json({ build: 'v1', view: '1075x700 @1.17x' }));
  assert.strictEqual(j.schema, 'watch.streamstats/1'); assert.strictEqual(j.windowS, 30);
  assert.ok(/^2026-11-16T12:0\d:\d\dZ$/.test(j.start.replace(/\.\d+Z/, 'Z')) && j.end > j.start, j.start + ' ' + j.end);
  assert.strictEqual(typeof j.tzOffsetMin, 'number'); assert.strictEqual(j.player, 'palette'); assert.strictEqual(j.build, 'v1');
  assert.strictEqual(j.serverSampleAtMs, c.t - 40);
  assert.ok(j.events.some((e) => e.kind === 'resync') && j.events.some((e) => e.kind === 'key'));
  assert.strictEqual(j.perSecond.shownFps.length, 30); assert.ok(j.gaps.length > 40 && Array.isArray(j.gaps[0]));
});

test('draw paints every chart, the time axis and the summary without throwing, in full and compact form', () => {
  const calls = [], texts = [];
  const g = new Proxy({}, { get: (t, k) => (k === 'fillText' ? (s) => texts.push(String(s)) : k in t ? t[k] : () => { calls.push(k); }), set: (t, k, v) => { t[k] = v; return true; } });
  const c = clock(Date.UTC(2026, 10, 16, 15, 4, 5)), m = new StreamMonitor(c.now);
  m.player = 'palette'; m.target = 30;
  for (let k = 0; k < 400; k++) { c.t += k === 200 ? 130 : 33; m.arrive(4000, k % 150 === 0); m.shownFrame(); m.delay(60); }
  m.server({ throttlePct: 40, cpuPct: 15, ringFps: 29, atMs: c.t });
  for (const compact of [false, true]) {
    texts.length = 0;
    const used = draw(g, compact ? 300 : 360, 220, m, { player: 'palette', build: 'v1', view: '1075x700' }, mk(), compact);
    assert.ok(used > 80 && used <= (compact ? HEIGHT_COMPACT : HEIGHT_FULL), 'height ' + used + ' against the canvas ' + (compact ? HEIGHT_COMPACT : HEIGHT_FULL));
    const all = texts.join('|');
    assert.ok(/\d\d:\d\d:\d\d/.test(all), 'timestamps'); assert.ok(/now/.test(all), 'now marker'); assert.ok(/hitches 1/.test(all), 'hitch count: ' + all);
    assert.ok(/palette/.test(all) && /v1/.test(all) && /30 s/.test(all), 'header');
  }
});

test('SlowGate: a rate hovering around 15 never flips; 3 s below 12 turns on, 3 s above 18 turns off', () => {
  const g = new SlowGate(12, 18, 3), flips = [];
  for (let k = 0; k < 60; k++) { const f = g.update(k % 2 ? 14 : 16); if (f) flips.push(f); }
  assert.deepStrictEqual(flips, [], 'hovering at 14 to 16 flips nothing');
  const seq = [13, 11, 11, 11, 17, 19, 19, 19];
  const out = seq.map((f) => g.update(f));
  assert.deepStrictEqual(out, ['', '', '', 'on', '', '', '', 'off']);
  // a single stall second inside good fps does not turn it on, and one good second inside a slow stretch does not turn it off
  const h = new SlowGate(12, 18, 3); const r = [30, 30, 5, 30, 30, 5, 30].map((f) => h.update(f)); assert.ok(r.every((x) => x === ''));
});
