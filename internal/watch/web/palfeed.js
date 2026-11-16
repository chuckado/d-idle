/* Lossless frame feed player (docs/video-feed.md). Prototype, not loaded by any page yet. Wire format: see
   internal/watch/palfeed/codec.go; this file must stay byte-compatible with it (tools/videofeed/palfeed.test.js decodes
   the Go fixtures). Pure functions first (node-testable), the DOM player last. */
(function (root) {
  'use strict';
  var HEADER = 28, KEY = 1;

  function u16(b, o) { return b[o] | (b[o + 1] << 8); }
  function u32(b, o) { return (b[o] | (b[o + 1] << 8) | (b[o + 2] << 16) | (b[o + 3] << 24)) >>> 0; }

  // parseHeader reads a message's fixed header; throws on anything that is not version 1.
  function parseHeader(m) {
    if (m.length < HEADER || m[0] !== 0x50 || m[1] !== 0x46 || m[2] !== 1) throw new Error('palfeed: bad header');
    var h = { key: (m[3] & KEY) !== 0, w: u16(m, 4), h: u16(m, 6), tile: m[8], seq: u32(m, 10), base: u32(m, 14),
              gameMs: u32(m, 18), n: u32(m, 22), dx: (m[26] << 24) >> 24, dy: (m[27] << 24) >> 24 };
    if (!h.w || !h.h || !h.tile || m.length !== HEADER + h.n) throw new Error('palfeed: bad message');
    return h;
  }

  // Decoder keeps the RGBA frame a canvas shows. apply(header, raw) takes the INFLATED payload (the caller inflates:
  // zlib.inflateRawSync in node, DecompressionStream('deflate-raw') in a browser). Two buffers alternate, so a scroll
  // tile can read the old frame while the new one is written; `pix` is always the newest.
  function Decoder() { this.w = 0; this.h = 0; this.seq = 0; this.have = false; this.pix = null; this.spare = null; this.dirty = null; }
  Decoder.prototype.needsKey = function (h) { return !h.key && (!this.have || this.w !== h.w || this.h !== h.h || this.seq !== h.base); };
  Decoder.prototype.apply = function (h, raw) {
    if (this.needsKey(h)) { var e = new Error('palfeed: need key'); e.needKey = true; throw e; }
    var w = h.w, ht = h.h, t = h.tile, tx = Math.ceil(w / t), ty = Math.ceil(ht / t), key = h.key, pos = 0, old = this.pix;
    var out, nm = key ? 0 : (tx * ty + 3) >> 2;
    if (raw.length < nm) throw new Error('palfeed: short payload');
    pos = nm;
    if (key) out = new Uint8ClampedArray(w * ht * 4);
    else { out = this.spare && this.spare.length === w * ht * 4 ? this.spare : new Uint8ClampedArray(w * ht * 4); out.set(old); }
    var x0min = w, y0min = ht, x1max = 0, y1max = 0, dx = h.dx, dy = h.dy;
    for (var j = 0; j < ty; j++) for (var i = 0; i < tx; i++) {
      var mode = 1;
      if (!key) { var k = j * tx + i; mode = (raw[k >> 2] >> (2 * (k & 3))) & 3; if (mode === 0) continue; }
      var x0 = i * t, y0 = j * t, tw = Math.min(t, w - x0), th = Math.min(t, ht - y0);
      if (mode !== 3 && pos + tw * th * 3 > raw.length) throw new Error('palfeed: short payload');
      for (var y = y0; y < y0 + th; y++) {
        var o = (y * w + x0) * 4, pr = 0, pg = 0, pb = 0, sy = y - dy;
        for (var x = 0; x < tw; x++, o += 4) {
          var r = 0, g = 0, b = 0;
          if (mode >= 2) {   // shifted reference, black outside the frame
            var sx = x0 + x - dx;
            if (sy >= 0 && sy < ht && sx >= 0 && sx < w) { var so = (sy * w + sx) * 4; r = old[so]; g = old[so + 1]; b = old[so + 2]; }
            if (mode === 3) { out[o] = r; out[o + 1] = g; out[o + 2] = b; out[o + 3] = 255; continue; }
          } else if (!key) { r = old[o]; g = old[o + 1]; b = old[o + 2]; }
          if (key) { pr = (pr + raw[pos]) & 255; pg = (pg + raw[pos + 1]) & 255; pb = (pb + raw[pos + 2]) & 255; out[o] = pr; out[o + 1] = pg; out[o + 2] = pb; }
          else { out[o] = r ^ raw[pos]; out[o + 1] = g ^ raw[pos + 1]; out[o + 2] = b ^ raw[pos + 2]; }
          out[o + 3] = 255; pos += 3;
        }
      }
      x0min = Math.min(x0min, x0); y0min = Math.min(y0min, y0); x1max = Math.max(x1max, x0 + tw); y1max = Math.max(y1max, y0 + th);
    }
    this.spare = key ? null : old; this.pix = out;
    this.w = w; this.h = ht; this.seq = h.seq; this.have = true;
    this.dirty = key ? { x: 0, y: 0, w: w, h: ht } : (x1max > x0min ? { x: x0min, y: y0min, w: x1max - x0min, h: y1max - y0min } : null);
    return this.dirty;
  };

  // splitter turns arbitrary chunks of the u32-length-prefixed body into whole messages.
  function Splitter() { this.buf = new Uint8Array(0); }
  Splitter.prototype.push = function (chunk) {
    var b = new Uint8Array(this.buf.length + chunk.length), out = [], p = 0;
    b.set(this.buf, 0); b.set(chunk, this.buf.length);
    while (b.length - p >= 4) {
      var n = u32(b, p);
      if (n > (16 << 20)) throw new Error('palfeed: message too large');
      if (b.length - p - 4 < n) break;
      out.push(b.subarray(p + 4, p + 4 + n)); p += 4 + n;
    }
    this.buf = b.subarray(p);
    return out;
  };

  // fit picks the canvas CSS size: an integer multiple of the frame in DEVICE pixels, so pixels stay square and sharp
  // on any devicePixelRatio; below 1x it falls back to the exact fit (a phone narrower than the frame).
  function fit(frameW, frameH, availCssW, availCssH, dpr) {
    dpr = dpr || 1;
    var s = Math.floor(Math.min(availCssW * dpr / frameW, availCssH * dpr / frameH));
    if (s >= 1) return { cssW: frameW * s / dpr, cssH: frameH * s / dpr, scale: s };
    var f = Math.min(availCssW / frameW, availCssH / frameH);
    return { cssW: frameW * f, cssH: frameH * f, scale: f * dpr };
  }

  // detect asks the host whether the feed exists; anything but a 200 with paletteFeed:true keeps the MJPEG img.
  function detect(fetchFn, base, headers) {
    return fetchFn(base + '/api/palfeed', { headers: headers || {}, cache: 'no-store' })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (j) { return j && j.paletteFeed === true && j.version === 1 ? j : null; })
      .catch(function () { return null; });
  }

  function browserInflate(bytes) {
    var ds = new DecompressionStream('deflate-raw');
    return new Response(new Blob([bytes]).stream().pipeThrough(ds)).arrayBuffer().then(function (a) { return new Uint8Array(a); });
  }

  // Player: fetch the stream, decode in order, draw only the newest state once per animation frame (latest wins).
  // o: {canvas, url, headers, fetch, inflate, raf, onFallback(reason), onStats(s)}
  function Player(o) {
    this.o = o; this.dec = new Decoder(); this.ctx = o.canvas.getContext('2d', { alpha: false }); this.img = null;
    this.stopped = false; this.pending = null; this.raf = o.raf || (root.requestAnimationFrame && root.requestAnimationFrame.bind(root));
    this.frames = 0; this.drawn = 0; this.bytes = 0; this.failures = 0; this.lowRate = false; this.ctl = null;
  }
  Player.prototype.setLowRate = function (on) { this.lowRate = !!on; if (this.ctl) this.ctl.abort(); };
  Player.prototype.stop = function () { this.stopped = true; if (this.ctl) this.ctl.abort(); };
  Player.prototype.start = function () { this._run(); };
  Player.prototype._run = function () {
    var self = this, o = this.o, inflate = o.inflate || browserInflate, sp = new Splitter();
    this.ctl = typeof AbortController === 'function' ? new AbortController() : { abort: function () {} };
    var url = o.url + (this.lowRate ? (o.url.indexOf('?') < 0 ? '?' : '&') + 'fps=10' : '');
    var chain = Promise.resolve();   // messages are applied strictly in order
    var f = o.fetch || root.fetch.bind(root);
    f(url, { headers: o.headers || {}, signal: this.ctl.signal, cache: 'no-store' }).then(function (r) {
      if (!r.ok || !r.body) throw new Error('status ' + r.status);
      var rd = r.body.getReader();
      return (function pump() {
        return rd.read().then(function (c) {
          if (c.done) throw new Error('ended');
          sp.push(c.value).forEach(function (m) {
            self.bytes += m.length + 4;
            chain = chain.then(function () {
              var h = parseHeader(m);
              if (self.dec.needsKey(h)) throw Object.assign(new Error('need key'), { needKey: true });
              return Promise.resolve(inflate(m.subarray(HEADER))).then(function (raw) { self._got(self.dec.apply(h, raw)); });
            });
          });
          return chain.then(pump);
        });
      })();
    }).catch(function (e) {
      if (self.stopped) return;
      if (e && e.name === 'AbortError') { if (!self.stopped) self._run(); return; }   // low-rate switch: reconnect
      self.failures++;
      if (self.failures >= 3) { if (o.onFallback) o.onFallback(String(e && e.message || e)); return; }
      setTimeout(function () { if (!self.stopped) self._run(); }, 500 * self.failures);   // reconnect = new key frame
    });
  };
  Player.prototype._got = function (dirty) {
    this.failures = 0; this.frames++;
    if (dirty) this.pending = this.pending ? union(this.pending, dirty) : dirty;
    if (this.raf && !this.rafOn) { var self = this; this.rafOn = true; this.raf(function () { self.rafOn = false; self._draw(); }); }
  };
  Player.prototype._draw = function () {
    var d = this.dec, p = this.pending;
    if (!p || !d.have) return;
    this.pending = null;
    var c = this.o.canvas;
    if (c.width !== d.w || c.height !== d.h) { c.width = d.w; c.height = d.h; this.img = null; p = { x: 0, y: 0, w: d.w, h: d.h }; }
    if (!this.img || this.img.data !== d.pix) this.img = new ImageData(d.pix, d.w, d.h);
    this.ctx.putImageData(this.img, 0, 0, p.x, p.y, p.w, p.h);
    this.drawn++;
  };
  function union(a, b) {
    var x0 = Math.min(a.x, b.x), y0 = Math.min(a.y, b.y), x1 = Math.max(a.x + a.w, b.x + b.w), y1 = Math.max(a.y + a.h, b.y + b.h);
    return { x: x0, y: y0, w: x1 - x0, h: y1 - y0 };
  }

  var api = { parseHeader: parseHeader, Decoder: Decoder, Splitter: Splitter, fit: fit, detect: detect, Player: Player, HEADER: HEADER };
  if (typeof module !== 'undefined' && module.exports) module.exports = api; else root.PalFeed = api;
})(typeof window !== 'undefined' ? window : globalThis);
