/* Lossless frame feed player (docs/video-feed.md). Prototype, not loaded by any page yet. Wire format: see
   internal/watch/palfeed/codec.go; this file must stay byte-compatible with it (tools/videofeed/palfeed.test.js decodes
   the Go fixtures). Pure functions first (node-testable), the DOM player last. */
(function (root) {
  'use strict';
  var HEADER = 28, KEY = 1;

  // an error a retry cannot help: the stream itself is broken (the player then falls back to MJPEG)
  function fatal(msg) { var e = new Error(msg); e.fatal = true; return e; }
  function u16(b, o) { return b[o] | (b[o + 1] << 8); }
  function u32(b, o) { return (b[o] | (b[o + 1] << 8) | (b[o + 2] << 16) | (b[o + 3] << 24)) >>> 0; }

  // parseHeader reads a message's fixed header; throws on anything that is not version 1.
  function parseHeader(m) {
    if (m.length < HEADER || m[0] !== 0x50 || m[1] !== 0x46 || m[2] !== 1) throw fatal('palfeed: bad header');
    var h = { key: (m[3] & KEY) !== 0, w: u16(m, 4), h: u16(m, 6), tile: m[8], seq: u32(m, 10), base: u32(m, 14),
              gameMs: u32(m, 18), n: u32(m, 22), dx: (m[26] << 24) >> 24, dy: (m[27] << 24) >> 24 };
    if (!h.w || !h.h || !h.tile || m.length !== HEADER + h.n) throw fatal('palfeed: bad message');
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
    if (raw.length < nm) throw fatal('palfeed: short payload');
    pos = nm;
    if (key) out = new Uint8ClampedArray(w * ht * 4);
    else { out = this.spare && this.spare.length === w * ht * 4 ? this.spare : new Uint8ClampedArray(w * ht * 4); out.set(old); }
    var x0min = w, y0min = ht, x1max = 0, y1max = 0, dx = h.dx, dy = h.dy;
    for (var j = 0; j < ty; j++) for (var i = 0; i < tx; i++) {
      var mode = 1;
      if (!key) { var k = j * tx + i; mode = (raw[k >> 2] >> (2 * (k & 3))) & 3; if (mode === 0) continue; }
      var x0 = i * t, y0 = j * t, tw = Math.min(t, w - x0), th = Math.min(t, ht - y0);
      if (mode !== 3 && pos + tw * th * 3 > raw.length) throw fatal('palfeed: short payload');
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

  // splitter turns arbitrary chunks of the u32-length-prefixed body into whole messages. One growing buffer with a read
  // offset: a chunk is copied once and no array is built per chunk (the old concatenation allocated the whole backlog
  // per 16 to 64 KB chunk, which a busy main thread felt as GC pauses). The returned views are valid until the next push:
  // the caller must consume them (parse, hand to inflate: a Blob copies at once) before pushing again.
  function Splitter() { this.buf = new Uint8Array(1 << 16); this.off = 0; this.end = 0; }
  Splitter.prototype.push = function (chunk) {
    if (this.end + chunk.length > this.buf.length) {
      var live = this.end - this.off, need = live + chunk.length;
      var next = need > this.buf.length ? new Uint8Array(Math.max(need, this.buf.length * 2)) : this.buf;
      next.set(this.buf.subarray(this.off, this.end), 0);
      this.buf = next; this.off = 0; this.end = live;
    }
    this.buf.set(chunk, this.end); this.end += chunk.length;
    var out = [], b = this.buf;
    while (this.end - this.off >= 4) {
      var n = u32(b, this.off);
      if (n > (16 << 20)) throw fatal('palfeed: message too large');
      if (this.end - this.off - 4 < n) break;
      out.push(b.subarray(this.off + 4, this.off + 4 + n)); this.off += 4 + n;
    }
    if (this.off === this.end) this.off = this.end = 0;
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

  // fitFill is the canvas CSS size that fills the box: the largest size that keeps the frame's aspect (contain), at
  // any fractional scale, rounded to whole DEVICE pixels so the edges stay clean. `integer` says the scale is a whole
  // number of device pixels per frame pixel (then nearest-neighbour is as sharp as it can be; at a fractional scale
  // nearest-neighbour shimmers when the picture scrolls, so the page smooths there unless Sharp pixels is on).
  function fitFill(frameW, frameH, availCssW, availCssH, dpr) {
    dpr = dpr || 1;
    var s = Math.min(availCssW * dpr / frameW, availCssH * dpr / frameH);
    if (!(s > 0)) return { cssW: frameW, cssH: frameH, scale: 1, integer: true };
    var devW = Math.max(1, Math.round(frameW * s)), devH = Math.max(1, Math.round(frameH * s));
    return { cssW: devW / dpr, cssH: devH / dpr, scale: s, integer: Math.abs(s - Math.round(s)) < 0.005 };
  }

  // detect asks the host whether the feed exists; anything but a 200 with paletteFeed:true keeps the MJPEG img.
  function detect(fetchFn, base, headers, query) {
    return fetchFn(base + '/api/palfeed' + (query || ''), { headers: headers || {}, cache: 'no-store' })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (j) { return j && j.paletteFeed === true && j.version === 1 ? j : null; })
      .catch(function () { return null; });
  }

  // negotiate: does this host offer the feed to this page? /api/glance must list "palfeed" under stream.formats and the
  // capability probe must answer (it is 503 until the first frame, and the probe wakes the ring, so ask up to `tries`
  // times a second apart). Resolves to the probe reply or null; never rejects. tq is the query string with the token.
  function negotiate(fetchFn, tq, tries) {
    tq = tq || '';
    return fetchFn('/api/glance' + tq, { cache: 'no-store', credentials: 'same-origin' })
      .then(function (g) { return g.ok ? g.json() : null; })
      .then(function (j) {
        var fm = j && j.stream && j.stream.formats;
        if (!fm || fm.indexOf('palfeed') < 0) return null;
        var n = 0;
        return (function ask() {
          return detect(fetchFn, '', {}, tq).then(function (d) {
            if (d || ++n >= (tries || 5)) return d;
            return new Promise(function (r) { setTimeout(r, 1000); }).then(ask);
          });
        })();
      })
      .catch(function () { return null; });
  }

  function browserInflate(bytes) {
    var ds = new DecompressionStream('deflate-raw');
    return new Response(new Blob([bytes]).stream().pipeThrough(ds)).arrayBuffer().then(function (a) { return new Uint8Array(a); });
  }

  // Player: fetch the stream, inflate each message as it arrives, and PRESENT on requestAnimationFrame by a presentation
  // clock (below). o: {canvas, url, headers, fetch, inflate, raf, now, onFrame(), onBytes(n), onEnd(err), onFallback(reason),
  // log(msg)}. With onEnd the caller owns reconnects (the video panel's controller: backoff, stall watchdog, hidden tabs);
  // without it the player retries itself 3 times (a reconnect is a new key frame) and then calls onFallback. An error
  // carries .fatal when a retry cannot help (a broken stream: bad header, short payload, oversized message).
  //
  // Presentation clock (docs/video-feed.md section 11). Messages arrive with network and GC jitter; showing each at its own
  // arrival moment makes steady motion look uneven. Every message carries the d2d game time of its frame. The player keeps
  // the smallest (arrival - gameMs) of the last 2 s, which is the transit of the least delayed message, and presents the
  // frame of game time g at g + that + D, with D one frame interval plus 1.5 times the measured jitter, clamped to 1..2 frames
  // (never above MAX_DELAY ms). A message that is due is decoded in order and ONE draw shows the newest of them (latest wins);
  // a message older than MAX_AGE ms at presentation time is never waited for. When the game clock does not track the wall
  // clock (a fast-forwarded lockstep game) the due time is capped at arrival + D, which is "present on the next animation
  // frame". Counters: delivered messages, decoded, presented draws, dropped (decoded but never drawn on their own: merged
  // into a later draw), and the standard deviation of the intervals between draws.
  var MAX_DELAY = 70, MAX_AGE = 80, WINDOW_MS = 2000;
  function Player(o) {
    this.o = o; this.dec = new Decoder(); this.ctx = o.canvas.getContext('2d', { alpha: false, desynchronized: true });
    this.stopped = false; this.pending = null; this.raf = o.raf || (root.requestAnimationFrame && root.requestAnimationFrame.bind(root));
    this.now = o.now || (root.performance && root.performance.now.bind(root.performance)) || Date.now;
    this.frames = 0; this.drawn = 0; this.bytes = 0; this.failures = 0; this.lowRate = false; this.ctl = null; this.gen = 0; this.imgs = [];
    this.q = []; this.transits = []; this.jit = 0; this.step = 33; this.lastGame = -1; this.lastDraw = -1; this.ivs = [];
    this.work = 0; this.delivered = 0; this.decoded = 0; this.dropped = 0; this.lat = 0; this.rafOn = false;
  }
  Player.prototype.setLowRate = function (on) { this.lowRate = !!on; this._abort(); };
  Player.prototype._abort = function () { if (this.ctl) this.ctl.abort(); };
  Player.prototype.stop = function () { this.stopped = true; this.gen++; this._abort(); };
  Player.prototype.start = function () { this._run(); };
  // stats: counters since the connection started; fps are over the last 2 s of presented draws and arrivals
  Player.prototype.stats = function () {
    var n = this.ivs.length, mean = 0, sd = 0, i;
    for (i = 0; i < n; i++) mean += this.ivs[i];
    if (n) mean /= n;
    for (i = 0; i < n; i++) sd += (this.ivs[i] - mean) * (this.ivs[i] - mean);
    return { delivered: this.delivered, decoded: this.decoded, presented: this.drawn, dropped: this.dropped, queue: this.q.length,
             presentFps: mean ? 1000 / mean : 0, jitterMs: n ? Math.sqrt(sd / n) : 0, latencyMs: this.lat, delayMs: this._delay(), workMs: this.work };
  };
  Player.prototype._delay = function () { return Math.max(this.step, Math.min(MAX_DELAY, this.step + 1.5 * this.jit)); };
  // _arrive stamps a message with its due time. g is the frame's game time in ms.
  Player.prototype._arrive = function (h, at) {
    var g = h.gameMs, T = this.transits;
    if (this.o.noClock) return at;   // measurement only: present at the first animation frame after arrival (the old behaviour)
    if (this.lastGame >= 0 && g === this.lastGame) return at;   // a heartbeat repeats the last frame's time: nothing to schedule
    if (this.lastGame >= 0) {
      if (g < this.lastGame - 1000) { T.length = 0; this.jit = 0; }   // a restarted d2d: a new clock
      else if (g > this.lastGame) { var d = g - this.lastGame; if (d < 200) this.step += (Math.min(d, 70) - this.step) * 0.1; }
    }
    if (g > this.lastGame) this.lastGame = g;
    var tr = at - g;
    T.push([at, tr]);
    while (T.length && at - T[0][0] > WINDOW_MS) T.shift();
    var m = tr;
    for (var i = 0; i < T.length; i++) if (T[i][1] < m) m = T[i][1];
    this.jit += (Math.abs(tr - m) - this.jit) * 0.05;
    var due = g + m + this._delay();
    return Math.min(due, at + MAX_DELAY);   // a game clock that does not follow the wall clock: present on the next frame
  };
  Player.prototype._run = function () {
    var self = this, o = this.o, inflate = o.inflate || browserInflate, sp = new Splitter(), gen = ++this.gen;
    this.dec = new Decoder(); this.pending = null; this.q = []; this.transits = []; this.lastGame = -1; // a new connection starts from a key frame
    this.ctl = typeof AbortController === 'function' ? new AbortController() : { abort: function () {} };
    var url = o.url + (this.lowRate ? (o.url.indexOf('?') < 0 ? '?' : '&') + 'fps=10' : '');
    var f = o.fetch || root.fetch.bind(root);
    f(url, { headers: o.headers || {}, signal: this.ctl.signal, cache: 'no-store', credentials: 'same-origin' }).then(function (r) {
      if (!r.ok || !r.body) throw new Error('status ' + r.status);
      var rd = r.body.getReader();
      return (function pump() {
        return rd.read().then(function (c) {
          if (c.done) throw new Error('ended');
          var msgs = sp.push(c.value), at = self.now();
          for (var i = 0; i < msgs.length; i++) {
            var m = msgs[i], h = parseHeader(m), e = { h: h, at: at, due: self._arrive(h, at), raw: null, err: null };
            self.bytes += m.length + 4; self.delivered++;
            if (o.onBytes) o.onBytes(m.length + 4);
            // inflate at once (a Blob copies the bytes synchronously, so the splitter may reuse its buffer); decoding waits for the due time
            self._inflate(e, inflate(m.subarray(HEADER)));
            self.q.push(e);
          }
          self._kick();
          return pump();
        });
      })();
    }).catch(function (e) {
      if (self.stopped || gen !== self.gen && !(e && e.name === 'AbortError')) return;
      if (e && e.name === 'AbortError') { if (!self.stopped && gen === self.gen) self._run(); return; }   // low-rate switch: reconnect
      if (o.log) o.log('palfeed: ' + (e && e.message || e));
      if (o.onEnd) { o.onEnd(e); return; }
      self.failures++;
      if (e && e.fatal || self.failures >= 3) { if (o.onFallback) o.onFallback(String(e && e.message || e)); return; }
      setTimeout(function () { if (!self.stopped) self._run(); }, 500 * self.failures);   // reconnect = new key frame
    });
  };
  Player.prototype._inflate = function (e, result) {
    var self = this;
    Promise.resolve(result).then(function (raw) { e.raw = raw; self._kick(); }, function (err) { e.err = err; self._kick(); });
  };
  Player.prototype._kick = function () {
    if (this.rafOn || !this.raf || this.stopped) return;
    var self = this; this.rafOn = true;
    this.raf(function (ts) { self.rafOn = false; self._tick(ts); });
  };
  // _tick runs on an animation frame: decode what is due, in order, draw once. A decode error ends the stream like a fetch error.
  Player.prototype._tick = function (ts) {
    if (this.stopped) return;
    var q = this.q, applied = 0, shown = 0, newest = null;
    if (ts === undefined) ts = this.now();
    var w0 = this.now();
    // the head may only be applied when its payload is inflated; later ones wait behind it (strict order)
    while (q.length && (q[0].raw || q[0].err)) {
      var e = q[0];
      var late = ts - e.at > MAX_AGE;   // too old to wait for: take everything that is ready
      if (!late && e.due > ts) break;
      q.shift();
      if (e.err) { this._fail(e.err); return; }
      try {
        if (this.dec.needsKey(e.h)) throw Object.assign(new Error('need key'), { needKey: true });
        var d = this.dec.apply(e.h, e.raw);
        this.decoded++; applied++; newest = e;
        if (d) { this.pending = this.pending ? union(this.pending, d) : d; shown++; }
      } catch (err) { this._fail(err); return; }
    }
    if (applied) {
      this.frames += applied;
      this.failures = 0;
      if (this.o.onFrame) this.o.onFrame();
      if (this.pending) this._draw(ts, shown, newest);   // an empty delta (a heartbeat) changes nothing: it is not a dropped picture
    }
    this.work += (this.now() - w0 - this.work) * 0.05;
    if (q.length) this._kick();   // more to present later; an idle player sleeps until the next message
  };
  Player.prototype._fail = function (err) {
    this.q = [];
    if (this.o.log) this.o.log('palfeed: ' + (err && err.message || err));
    this.gen++; this._abort();   // this connection is over; the caller reconnects (a key frame) or falls back
    if (this.o.onEnd) this.o.onEnd(err);
    else if (this.o.onFallback) this.o.onFallback(String(err && err.message || err));
  };
  Player.prototype._draw = function (ts, applied, newest) {
    var d = this.dec, p = this.pending;
    this.pending = null;
    var c = this.o.canvas;
    if (c.width !== d.w || c.height !== d.h) { c.width = d.w; c.height = d.h; this.imgs = []; p = { x: 0, y: 0, w: d.w, h: d.h }; }
    // the decoder alternates two pixel buffers: one ImageData per buffer, made once (a dirty rect then stays a dirty rect)
    var imgs = this.imgs, img = null;
    for (var i = 0; i < imgs.length; i++) if (imgs[i].data === d.pix) img = imgs[i];
    if (!img) { img = new ImageData(d.pix, d.w, d.h); imgs.push(img); if (imgs.length > 4) imgs.shift(); }
    this.ctx.putImageData(img, 0, 0, p.x, p.y, p.w, p.h);
    this.drawn++;
    this.dropped += Math.max(0, applied - 1);
    this.lat += ((ts - newest.at) - this.lat) * 0.1;
    if (this.lastDraw >= 0) { this.ivs.push(ts - this.lastDraw); if (this.ivs.length > 120) this.ivs.shift(); }
    this.lastDraw = ts;
  };
  function union(a, b) {
    var x0 = Math.min(a.x, b.x), y0 = Math.min(a.y, b.y), x1 = Math.max(a.x + a.w, b.x + b.w), y1 = Math.max(a.y + a.h, b.y + b.h);
    return { x: x0, y: y0, w: x1 - x0, h: y1 - y0 };
  }

  var api = { parseHeader: parseHeader, Decoder: Decoder, Splitter: Splitter, fit: fit, fitFill: fitFill, detect: detect, negotiate: negotiate, Player: Player, HEADER: HEADER };
  if (typeof module !== 'undefined' && module.exports) module.exports = api; else root.PalFeed = api;
})(typeof window !== 'undefined' ? window : globalThis);
