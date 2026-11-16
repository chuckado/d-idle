// node --test tools/videofeed/palfeed.test.js   (node 20+; stdlib only)
const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');
const PF = require('../../internal/watch/web/palfeed.js');
const dir = path.join(__dirname, '../../internal/watch/palfeed/testdata');
const rd = (n) => new Uint8Array(fs.readFileSync(path.join(dir, n)));
const apply = (dec, m) => { const h = PF.parseHeader(m); return dec.apply(h, new Uint8Array(zlib.inflateRawSync(m.subarray(PF.HEADER)))); };
const rgbEqual = (pix, want) => { for (let i = 0; i < want.length; i += 4) for (let c = 0; c < 3; c++) if (pix[i + c] !== want[i + c]) return false; return true; };

test('decodes the Go key frame and delta byte for byte', () => {
  const dec = new PF.Decoder();
  const d1 = apply(dec, rd('key.pf'));
  assert.deepStrictEqual(d1, { x: 0, y: 0, w: 96, h: 70 });
  assert.ok(rgbEqual(dec.pix, rd('a.rgba')));
  const d2 = apply(dec, rd('delta.pf'));
  assert.ok(rgbEqual(dec.pix, rd('b.rgba')));
  assert.ok(d2.w < 96 && d2.h < 70, 'dirty rect is smaller than the frame');
  assert.strictEqual(dec.seq, 2);
});

test('decodes a scrolled delta (modes 2 and 3) to the exact frame', () => {
  const dec = new PF.Decoder();
  apply(dec, rd('wkey.pf'));
  const m = rd('wdelta.pf'), h = PF.parseHeader(m);
  assert.deepStrictEqual([h.dx, h.dy], [-5, 3]);
  apply(dec, m);
  const rgb = Buffer.alloc(dec.pix.length / 4 * 3);
  for (let i = 0, o = 0; i < dec.pix.length; i += 4, o += 3) { rgb[o] = dec.pix[i]; rgb[o + 1] = dec.pix[i + 1]; rgb[o + 2] = dec.pix[i + 2]; }
  const want = fs.readFileSync(path.join(dir, 'wb.sha256'), 'utf8').trim();
  assert.strictEqual(require('node:crypto').createHash('sha256').update(rgb).digest('hex'), want);
});

test('a delta without its base asks for a key frame', () => {
  const dec = new PF.Decoder();
  assert.throws(() => apply(dec, rd('delta.pf')), (e) => e.needKey === true);
  apply(dec, rd('key.pf'));
  apply(dec, rd('delta.pf'));
  assert.throws(() => apply(dec, rd('delta.pf')), (e) => e.needKey === true, 'base seq moved on');
});

test('rejects garbage', () => {
  assert.throws(() => PF.parseHeader(new Uint8Array(10)));
  const m = rd('key.pf'); m[2] = 9;
  assert.throws(() => PF.parseHeader(m));
  assert.throws(() => PF.parseHeader(rd('key.pf').subarray(0, 40)));
});

test('splitter reassembles messages from any chunking', () => {
  const a = rd('key.pf'), b = rd('delta.pf');
  const wire = Buffer.concat([lenp(a), a, lenp(b), b]);
  for (const step of [1, 3, 7, 1000, wire.length]) {
    const sp = new PF.Splitter(), got = [];
    for (let i = 0; i < wire.length; i += step) got.push(...sp.push(new Uint8Array(wire.subarray(i, i + step))));
    assert.strictEqual(got.length, 2, 'step ' + step);
    assert.deepStrictEqual(Buffer.from(got[1]), Buffer.from(b));
  }
});
function lenp(m) { const b = Buffer.alloc(4); b.writeUInt32LE(m.length); return b; }

test('fit gives integer device-pixel scales', () => {
  assert.deepStrictEqual(PF.fit(800, 600, 1700, 1000, 1), { cssW: 800, cssH: 600, scale: 1 });
  assert.deepStrictEqual(PF.fit(800, 600, 900, 700, 2), { cssW: 800, cssH: 600, scale: 2 });  // dpr 2: 2 device px per frame px = 1 css px
  assert.deepStrictEqual(PF.fit(800, 600, 1700, 1300, 1), { cssW: 1600, cssH: 1200, scale: 2 });
  const phone = PF.fit(800, 600, 390, 800, 3);                 // 1170 device px wide: 1x device scale, 390 css wide
  assert.strictEqual(phone.scale, 1); assert.strictEqual(phone.cssW, 800 / 3);
  const tiny = PF.fit(800, 600, 300, 800, 1);                 // narrower than the frame: exact fit
  assert.ok(Math.abs(tiny.cssW - 300) < 1e-9);
});

test('detect keeps the MJPEG img unless the host says paletteFeed', async () => {
  const f = (status, body) => () => Promise.resolve({ ok: status === 200, json: () => Promise.resolve(body) });
  assert.ok(await PF.detect(f(200, { paletteFeed: true, version: 1, width: 800, height: 600 }), ''));
  assert.strictEqual(await PF.detect(f(404, null), ''), null);
  assert.strictEqual(await PF.detect(f(503, null), ''), null);
  assert.strictEqual(await PF.detect(f(200, { paletteFeed: true, version: 2 }), ''), null);
  assert.strictEqual(await PF.detect(() => Promise.reject(new Error('net')), ''), null);
});

// --- the player, driven by a fake clock, a fake animation frame and a body we push chunks into ---
function rig(extra) {
  const rafs = [], puts = [], clock = { t: 0 };
  global.ImageData = function (d, w, h) { this.data = d; this.width = w; this.height = h; };
  const canvas = { width: 0, height: 0, getContext: () => ({ putImageData: (...a) => puts.push(a.slice(3)) }) };
  const waiters = [], chunks = [];
  const body = { getReader: () => ({ read: () => chunks.length ? Promise.resolve({ done: false, value: chunks.shift() }) : new Promise((r) => waiters.push(r)) }) };
  const ev = { frames: 0, bytes: 0, end: null, logs: [] };
  const p = new PF.Player(Object.assign({ canvas, url: '/s', fetch: () => Promise.resolve({ ok: true, body }), raf: (cb) => rafs.push(cb), now: () => clock.t,
    inflate: (b) => new Uint8Array(zlib.inflateRawSync(b)), onFrame: () => ev.frames++, onBytes: (n) => (ev.bytes += n), onEnd: (e) => { ev.end = e; }, log: (m) => ev.logs.push(m) }, extra));
  const push = async (buf) => { const v = new Uint8Array(buf); const w = waiters.shift(); if (w) w({ done: false, value: v }); else chunks.push(v); await settle2(); };
  // run the next animation frame at time ts
  const frame = async (ts) => { clock.t = ts; const cb = rafs.shift(); if (cb) cb(ts); await settle2(); };
  return { p, puts, rafs, clock, push, frame, ev };
}
const settle2 = async () => { for (let i = 0; i < 6; i++) await new Promise((r) => setImmediate(r)); };
// a wire message from a fixture with its seq, base seq and game ms patched (the payload is reused: only the header matters to the clock)
function patch(buf, seq, base, gameMs) {
  const m = Buffer.from(buf); m.writeUInt32LE(seq, 10); m.writeUInt32LE(base, 14); m.writeUInt32LE(gameMs >>> 0, 18);
  return Buffer.concat([lenp(m), m]);
}

test('player: messages due together draw once with the union (latest wins)', async () => {
  const r = rig();
  r.p.start();
  await settle2();
  await r.push(Buffer.concat([patch(rd('key.pf'), 1, 0, 1000), patch(rd('delta.pf'), 2, 1, 1040)]));
  assert.strictEqual(r.p.stats().delivered, 2);
  await r.frame(100000); // far past both due times
  assert.strictEqual(r.puts.length, 1, 'one draw');
  assert.deepStrictEqual(r.puts[0], [0, 0, 96, 70], 'the union of key and delta is the whole frame');
  const st = r.p.stats();
  assert.strictEqual(st.decoded, 2); assert.strictEqual(st.presented, 1); assert.strictEqual(st.dropped, 1);
  assert.ok(rgbEqual(r.p.dec.pix, rd('b.rgba')));
  r.p.stop();
});

test('player: a frame is not shown before its due time and is shown after it', async () => {
  const r = rig();
  r.p.start(); await settle2();
  r.clock.t = 5000;
  await r.push(patch(rd('key.pf'), 1, 0, 1000));
  await r.frame(5001);          // due = game + (arrival - game) + delay: about 33 ms after arrival
  assert.strictEqual(r.puts.length, 0, 'not due yet');
  assert.ok(r.rafs.length >= 1, 'asks for another animation frame');
  await r.frame(5016); assert.strictEqual(r.puts.length, 0);
  await r.frame(5050); assert.strictEqual(r.puts.length, 1, 'due');
  r.p.stop();
});

test('player: one chunk, one message after the other: a late message older than 80 ms is taken at once', async () => {
  const r = rig();
  r.p.start(); await settle2();
  r.clock.t = 1000;
  await r.push(patch(rd('key.pf'), 1, 0, 0));
  r.clock.t = 1500; // the browser was busy for half a second
  await r.push(patch(rd('delta.pf'), 2, 1, 40));
  await r.frame(1501);
  assert.strictEqual(r.puts.length, 1, 'both taken, one draw');
  assert.ok(r.p.stats().latencyMs >= 0);
  r.p.stop();
});

// Presentation jitter: frames made at 30 fps game time arrive with 0 to 40 ms of delay; animation frames run at 60 Hz.
// The clock must make the intervals between draws more even than presenting at the first animation frame after arrival.
test('player: the presentation clock evens out arrival jitter and bounds the delay', async () => {
  const r = rig();
  r.p.start(); await settle2();
  let seed = 12345; const rnd = () => (seed = (seed * 1103515245 + 12345) & 0x7fffffff) / 0x7fffffff;
  const N = 150, step = 1000 / 30, arrivals = [];
  let prevAt = 0; // TCP never reorders: an arrival is never earlier than the one before
  for (let k = 0; k < N; k++) { const at = Math.max(prevAt, 2000 + k * step + rnd() * 40); prevAt = at; arrivals.push({ k, g: Math.round(k * step), at }); }
  const naive = []; let ai = 0, lastNaive = -1;
  for (let ts = 2000; ts < 2000 + N * step + 300; ts += 1000 / 60) {
    let got = false;
    while (ai < arrivals.length && arrivals[ai].at <= ts) {
      const a = arrivals[ai++]; r.clock.t = a.at;
      await r.push(a.k === 0 ? patch(rd('key.pf'), 1, 0, a.g) : patch(rd('delta.pf'), a.k + 1, a.k, a.g)); got = true;
    }
    if (got) { if (lastNaive >= 0) naive.push(ts - lastNaive); lastNaive = ts; }
    await r.frame(ts);
  }
  const st = r.p.stats();
  const sd = (v) => { const m = v.reduce((a, b) => a + b, 0) / v.length; return Math.sqrt(v.reduce((a, b) => a + (b - m) * (b - m), 0) / v.length); };
  assert.ok(st.delivered === N, 'delivered ' + st.delivered);
  assert.ok(st.presented >= N * 0.85 && st.presented <= N, 'presented ' + st.presented + ' of ' + N);
  assert.ok(st.latencyMs <= 80, 'latency ' + st.latencyMs);
  assert.ok(st.jitterMs < sd(naive), 'jitter ' + st.jitterMs.toFixed(1) + ' ms against naive ' + sd(naive).toFixed(1) + ' ms');
  assert.ok(st.delayMs >= 33 && st.delayMs <= 70, 'delay ' + st.delayMs);
  r.p.stop();
});

test('player: a heartbeat (same game time) is applied at once and does not disturb the clock', async () => {
  const r = rig();
  r.p.start(); await settle2();
  r.clock.t = 100;
  await r.push(patch(rd('key.pf'), 1, 0, 500));
  await r.frame(300);
  const before = r.p.stats().delayMs;
  r.clock.t = 1300;
  await r.push(patch(rd('delta.pf'), 1, 1, 500)); // seq stays 1 = base 1: an empty delta in real life; the payload here is a real delta
  await r.frame(1301);
  assert.ok(Math.abs(r.p.stats().delayMs - before) < 1, 'delay unchanged');
  r.p.stop();
});

test('player with onEnd: errors go to the caller, bad streams are fatal, onFrame and onBytes count', async () => {
  const run = async (wire) => {
    const r = rig();
    r.p.start(); await settle2();
    await r.push(wire);
    await r.frame(100000);
    await settle2();
    r.p.stop();
    return r.ev;
  };
  let ev = await run(Buffer.concat([lenp(rd('delta.pf')), rd('delta.pf')])); // a delta first: needs a key, reconnect helps
  assert.ok(ev.end && ev.end.needKey && !ev.end.fatal);
  const junk = Buffer.alloc(40, 7);
  ev = await run(Buffer.concat([lenp(junk), junk])); // garbage: the header check fails, a retry cannot help
  assert.ok(ev.end && ev.end.fatal, 'garbage is fatal');
  assert.ok(ev.logs.length >= 1 && /palfeed/.test(ev.logs[0]));
  ev = await run(Buffer.concat([lenp(rd('key.pf')), rd('key.pf'), lenp(junk), junk]));
  assert.ok(ev.bytes > 0 && ev.end.fatal);
});

test('negotiate needs palfeed in glance stream.formats and a ready probe', async () => {
  const mk = (formats, probe) => (u) => Promise.resolve(u.indexOf('/api/glance') === 0
    ? { ok: true, json: () => Promise.resolve({ stream: { formats } }) }
    : { ok: probe === 200, json: () => Promise.resolve({ paletteFeed: true, version: 1 }) });
  assert.ok(await PF.negotiate(mk(['mjpeg', 'palfeed'], 200), '?t=x', 1));
  assert.strictEqual(await PF.negotiate(mk(['mjpeg'], 200), '', 1), null);
  assert.strictEqual(await PF.negotiate(mk(undefined, 200), '', 1), null);
  assert.strictEqual(await PF.negotiate(mk(['mjpeg', 'palfeed'], 503), '', 1), null, 'probe not ready after the tries');
  assert.strictEqual(await PF.negotiate(() => Promise.reject(new Error('net')), '', 1), null);
});

test('fitFill fills the box at any scale, keeps 4:3, rounds to whole device pixels', () => {
  let r = PF.fitFill(800, 600, 1000, 700, 1);       // 1075 px window: width 1000, height limits at 700
  assert.ok(Math.abs(r.cssW / r.cssH - 4 / 3) < 0.01, 'aspect ' + r.cssW + 'x' + r.cssH);
  assert.ok(r.cssH === 700 && r.cssW === 933 && !r.integer && r.scale > 1.16 && r.scale < 1.17, JSON.stringify(r));
  r = PF.fitFill(800, 600, 1600, 1200, 1); assert.ok(r.integer && r.scale === 2 && r.cssW === 1600);
  r = PF.fitFill(800, 600, 400, 300, 1);   assert.ok(!r.integer && r.scale === 0.5, 'a phone card shrinks the picture');
  r = PF.fitFill(800, 600, 801, 700, 2);   // dpr 2: whole device pixels, so css sizes are multiples of 0.5
  assert.ok(Number.isInteger(r.cssW * 2) && Number.isInteger(r.cssH * 2), JSON.stringify(r));
  assert.ok(r.cssW <= 801 + 1e-9 && r.cssH <= 700 + 1e-9);
  r = PF.fitFill(800, 600, 0, 0, 1);       assert.ok(r.cssW > 0, 'a hidden box gives a harmless size');
});
