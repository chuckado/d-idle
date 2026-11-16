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

test('player: draws once per animation frame with the newest state (latest wins)', async () => {
  const wire = Buffer.concat([lenp(rd('key.pf')), rd('key.pf'), lenp(rd('delta.pf')), rd('delta.pf')]);
  const puts = [], rafs = [];
  global.ImageData = function (d, w, h) { this.data = d; this.width = w; this.height = h; };
  const canvas = { width: 0, height: 0, getContext: () => ({ putImageData: (...a) => puts.push(a.slice(3)) }) };
  let released;
  const body = { getReader: () => { let n = 0; return { read: () => n++ === 0 ? Promise.resolve({ done: false, value: new Uint8Array(wire) }) : new Promise((r) => { released = r; }) }; } };
  const p = new PF.Player({ canvas, url: '/stream.pfeed', fetch: () => Promise.resolve({ ok: true, body }), raf: (cb) => rafs.push(cb),
    inflate: (b) => new Uint8Array(zlib.inflateRawSync(b)) });
  p.start();
  for (let i = 0; i < 20 && p.frames < 2; i++) await new Promise((r) => setTimeout(r, 5));
  assert.strictEqual(p.frames, 2, 'both messages decoded');
  assert.strictEqual(rafs.length, 1, 'one animation frame requested for two messages');
  rafs.shift()();
  assert.strictEqual(puts.length, 1, 'one draw');
  assert.deepStrictEqual(puts[0], [0, 0, 96, 70], 'the union of key and delta is the whole frame');
  assert.ok(rgbEqual(p.dec.pix, rd('b.rgba')));
  p.stop();
});

function fakeBody(chunks) {
  return { getReader: () => { let n = 0; return { read: () => n < chunks.length ? Promise.resolve({ done: false, value: new Uint8Array(chunks[n++]) }) : new Promise(() => {}) }; } };
}
const settle = async (cond) => { for (let i = 0; i < 40 && !cond(); i++) await new Promise((r) => setTimeout(r, 5)); };

test('player: a delta after a draw stays a dirty rect (one ImageData per decoder buffer)', async () => {
  const puts = [], rafs = [];
  global.ImageData = function (d, w, h) { this.data = d; this.width = w; this.height = h; };
  const canvas = { width: 0, height: 0, getContext: () => ({ putImageData: (...a) => puts.push(a.slice(3)) }) };
  const p = new PF.Player({ canvas, url: '/s', fetch: () => Promise.resolve({ ok: true, body: fakeBody([Buffer.concat([lenp(rd('key.pf')), rd('key.pf')])]) }),
    raf: (cb) => rafs.push(cb), inflate: (b) => new Uint8Array(zlib.inflateRawSync(b)) });
  p.start();
  await settle(() => p.frames === 1);
  rafs.shift()();
  p._got(p.dec.apply(PF.parseHeader(rd('delta.pf')), new Uint8Array(zlib.inflateRawSync(rd('delta.pf').subarray(PF.HEADER)))));
  rafs.shift()();
  assert.deepStrictEqual(puts[0], [0, 0, 96, 70]);
  assert.ok(puts[1][2] < 96 || puts[1][3] < 70, 'second draw is a sub-rectangle, got ' + puts[1]);
  p.stop();
});

test('player with onEnd: errors go to the caller, bad streams are fatal, onFrame and onBytes count', async () => {
  global.ImageData = function (d, w, h) { this.data = d; };
  const canvas = { width: 0, height: 0, getContext: () => ({ putImageData() {} }) };
  const run = async (wire) => {
    const ev = { frames: 0, bytes: 0, end: null, logs: [] };
    const p = new PF.Player({ canvas, url: '/s', fetch: () => Promise.resolve({ ok: true, body: fakeBody([wire]) }), raf: () => {},
      inflate: (b) => new Uint8Array(zlib.inflateRawSync(b)), onFrame: () => ev.frames++, onBytes: (n) => (ev.bytes += n),
      onEnd: (e) => { ev.end = e; }, log: (m) => ev.logs.push(m) });
    p.start();
    await settle(() => ev.end);
    p.stop();
    return ev;
  };
  let ev = await run(Buffer.concat([lenp(rd('key.pf')), rd('key.pf')]).subarray(0, 4)); // only a length: stream ends later; no error yet
  assert.strictEqual(ev.end, null);
  ev = await run(Buffer.concat([lenp(rd('delta.pf')), rd('delta.pf')])); // a delta first: needs a key, reconnect helps
  assert.ok(ev.end && ev.end.needKey && !ev.end.fatal);
  const junk = Buffer.alloc(40, 7);
  ev = await run(Buffer.concat([lenp(junk), junk])); // garbage: header check fails, a retry cannot help
  assert.ok(ev.end && ev.end.fatal, 'garbage is fatal');
  assert.ok(ev.logs.length === 1 && /palfeed/.test(ev.logs[0]));
  ev = await run(Buffer.concat([lenp(rd('key.pf')), rd('key.pf'), lenp(junk), junk]));
  assert.ok(ev.frames === 1 && ev.bytes > 0 && ev.end.fatal);
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
