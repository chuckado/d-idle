// Stream monitor of the game video card (docs/video-feed.md section 12): a rolling 30 s record of what the page received and
// showed, drawn on one small canvas at a few Hz, and copied as JSON. Off by default (the card's Stats button, or ?pfdebug).
// The recorder is plain data (ring buffers, no DOM, no per-frame allocation) so it is node-testable; the drawing is last.
// Both players feed it: the lossless palette player and the MJPEG reader.
export const WINDOW_S = 30;
export const HITCH_MS = 52;   // "over 50 ms": two refresh intervals of a 60 Hz display are 33.3 ms, three are 50.0 ms plus timer noise
export const MAX_GAPS = 1200;
export const MAX_EVENTS = 64;
export const HEIGHT_FULL = 266;
export const HEIGHT_COMPACT = 208;

// StreamMonitor keeps one-second buckets for the last WINDOW_S + 1 seconds, the individual presentation gaps (ms between two
// shown frames) of the window, and the events. clock() is wall ms (Date.now), injected for tests.
export class StreamMonitor {
  constructor(clock) {
    this.clock = clock || Date.now;
    const n = WINDOW_S + 1;
    this.t0 = 0;                                   // wall second of bucket 0 of the ring: bucket i holds second sec0 + i (mod n)
    this.sec = new Float64Array(n).fill(-1);       // the second a ring slot currently holds (-1 = empty)
    this.deliv = new Int32Array(n); this.shown = new Int32Array(n); this.drop = new Int32Array(n);
    this.bytes = new Float64Array(n); this.keys = new Int32Array(n);
    this.dly = new Float64Array(n); this.dlyN = new Int32Array(n);          // presentation delay: sum and count per second
    this.thr = new Float64Array(n).fill(NaN); this.cpu = new Float64Array(n).fill(NaN); this.ring = new Float64Array(n).fill(NaN);
    this.gapT = new Float64Array(MAX_GAPS); this.gapV = new Float32Array(MAX_GAPS); this.gapN = 0; this.gapHead = 0;
    this.evT = new Float64Array(MAX_EVENTS); this.evK = new Array(MAX_EVENTS); this.evL = new Array(MAX_EVENTS); this.evN = 0; this.evHead = 0;
    this.lastShown = 0; this.hitches = [];         // wall ms of hitch gaps (trimmed to the window)
    this.target = 0;                               // the source's expected fps (the chart's target line)
    this.player = '';                              // palette | jpeg
    this.serverAt = 0;
  }
  _slot(now) {
    const s = Math.floor(now / 1000), i = ((s % this.sec.length) + this.sec.length) % this.sec.length;
    if (this.sec[i] !== s) {                       // the slot held an older second: reuse it
      this.sec[i] = s; this.deliv[i] = 0; this.shown[i] = 0; this.drop[i] = 0; this.bytes[i] = 0; this.keys[i] = 0;
      this.dly[i] = 0; this.dlyN[i] = 0; this.thr[i] = NaN; this.cpu[i] = NaN; this.ring[i] = NaN;
    }
    return i;
  }
  // arrive: one message or part of n bytes reached the page (key: a key frame or key-sized message)
  arrive(bytes, key) { const i = this._slot(this.clock()); this.deliv[i]++; this.bytes[i] += bytes; if (key) { this.keys[i]++; this.event('key', (bytes >> 10) + ' kB'); } }
  // shown: a picture was drawn; the gap since the previous one is recorded and a hitch flagged above HITCH_MS
  shownFrame() {
    const now = this.clock(), i = this._slot(now);
    this.shown[i]++;
    if (this.lastShown) {
      const g = now - this.lastShown, h = this.gapHead;
      this.gapT[h] = now; this.gapV[h] = g; this.gapHead = (h + 1) % MAX_GAPS; if (this.gapN < MAX_GAPS) this.gapN++;
      if (g > HITCH_MS) { this.hitches.push(now); this.event('hitch', Math.round(g) + ' ms'); }
    }
    this.lastShown = now;
  }
  dropped(n) { const i = this._slot(this.clock()); this.drop[i] += n; }
  delay(ms) { const i = this._slot(this.clock()); this.dly[i] += ms; this.dlyN[i]++; }
  // server: one sample of the host's numbers (cgroup throttle share, process cpu, ring rate), stamped with the host's clock
  server(s) { if (!s) return; const i = this._slot(this.clock()); this.thr[i] = s.throttlePct; this.cpu[i] = s.cpuPct; this.ring[i] = s.ringFps; this.serverAt = s.atMs || 0; }
  event(kind, label) {
    const h = this.evHead; this.evT[h] = this.clock(); this.evK[h] = kind; this.evL[h] = label || '';
    this.evHead = (h + 1) % MAX_EVENTS; if (this.evN < MAX_EVENTS) this.evN++;
  }
  // series: the last WINDOW_S whole seconds ending at the current one, oldest first, into the caller's arrays (no allocation)
  series(out) {
    const now = this.clock(), end = Math.floor(now / 1000), n = this.sec.length;
    for (let k = 0; k < WINDOW_S; k++) {
      const s = end - (WINDOW_S - 1 - k), i = ((s % n) + n) % n, ok = this.sec[i] === s;
      out.deliv[k] = ok ? this.deliv[i] : 0; out.shown[k] = ok ? this.shown[i] : 0; out.drop[k] = ok ? this.drop[i] : 0;
      out.kB[k] = ok ? this.bytes[i] / 1024 : 0; out.keys[k] = ok ? this.keys[i] : 0;
      out.delay[k] = ok && this.dlyN[i] ? this.dly[i] / this.dlyN[i] : NaN;
      out.thr[k] = ok ? this.thr[i] : NaN; out.cpu[k] = ok ? this.cpu[i] : NaN; out.ring[k] = ok ? this.ring[i] : NaN;
    }
    out.endSec = end; return out;
  }
  // summary over the window: current fps (last whole second), kB/s, p95 gap, hitches in the window
  summary() {
    const now = this.clock(), from = now - WINDOW_S * 1000;
    let n = 0; const g = [];
    for (let k = 0; k < this.gapN; k++) { const j = (this.gapHead - 1 - k + MAX_GAPS) % MAX_GAPS; if (this.gapT[j] < from) break; g.push(this.gapV[j]); n++; }
    g.sort((a, b) => a - b);
    while (this.hitches.length && this.hitches[0] < from) this.hitches.shift();
    const s = Math.floor(now / 1000) - 1, i = ((s % this.sec.length) + this.sec.length) % this.sec.length, ok = this.sec[i] === s;
    return { fps: ok ? this.shown[i] : 0, delivered: ok ? this.deliv[i] : 0, kBps: ok ? Math.round(this.bytes[i] / 1024) : 0,
             p95: n ? g[Math.min(n - 1, Math.ceil(n * 0.95) - 1)] : 0, p50: n ? g[n >> 1] : 0, max: n ? g[n - 1] : 0, hitches: this.hitches.length, frames: n };
  }
  // fpsAvg: mean shown-frames per second over the last n whole seconds (the current, partial second excluded); null with no data at all
  fpsAvg(n) {
    const end = Math.floor(this.clock() / 1000) - 1, len = this.sec.length; let sum = 0, any = false;
    for (let k = 0; k < n; k++) { const sec = end - k, i = ((sec % len) + len) % len; if (this.sec[i] === sec) { any = true; sum += this.shown[i]; } }
    return any ? sum / n : null;
  }
  // events of the window, newest last
  events(out) {
    const from = this.clock() - WINDOW_S * 1000; out.length = 0;
    for (let k = this.evN - 1; k >= 0; k--) { const j = (this.evHead - 1 - k + MAX_EVENTS) % MAX_EVENTS; if (this.evT[j] >= from) out.push({ t: this.evT[j], kind: this.evK[j], label: this.evL[j] }); }
    return out;
  }
  // json: everything a single paste needs (times in ISO and ms, timezone offset, build, player, viewport, server sample time)
  json(meta) {
    const ser = this.series({ deliv: [], shown: [], drop: [], kB: [], keys: [], delay: [], thr: [], cpu: [], ring: [] }), now = this.clock();
    const d = new Date(now), start = new Date((ser.endSec - WINDOW_S + 1) * 1000);
    const gaps = []; const from = now - WINDOW_S * 1000;
    for (let k = 0; k < this.gapN; k++) { const j = (this.gapHead - 1 - k + MAX_GAPS) % MAX_GAPS; if (this.gapT[j] < from) break; gaps.unshift([Math.round(this.gapT[j]), Math.round(this.gapV[j] * 10) / 10]); }
    const clean = (a) => a.map((v) => (Number.isFinite(v) ? Math.round(v * 100) / 100 : null));
    return JSON.stringify(Object.assign({
      schema: 'watch.streamstats/1', start: start.toISOString(), end: d.toISOString(), startMs: start.getTime(), endMs: now, windowS: WINDOW_S,
      tzOffsetMin: -d.getTimezoneOffset(), player: this.player, targetFps: this.target, serverSampleAtMs: this.serverAt,
      summary: this.summary(), perSecond: { deliveredFps: ser.deliv, shownFps: ser.shown, droppedFps: ser.drop, kB: clean(ser.kB), keys: ser.keys,
        delayMs: clean(ser.delay), serverThrottlePct: clean(ser.thr), serverCpuPct: clean(ser.cpu), serverRingFps: clean(ser.ring) },
      gaps, events: this.events([]) }, meta || {}));
  }
}

// SlowGate is the card's "slow" state with hysteresis, one update per second: on after `secs` seconds in a row below `on` fps, off after
// `secs` seconds in a row above `off` fps. update(fps) returns 'on' or 'off' at the second it flips, else ''. A value between the two
// thresholds keeps the state and resets the count, so a rate hovering at 14 to 16 cannot flip it.
export class SlowGate {
  constructor(on, off, secs) { this.on = on; this.off = off; this.secs = secs; this.slow = false; this.n = 0; }
  update(fps) {
    if (!this.slow) { this.n = fps < this.on ? this.n + 1 : 0; if (this.n >= this.secs) { this.slow = true; this.n = 0; return 'on'; } }
    else { this.n = fps > this.off ? this.n + 1 : 0; if (this.n >= this.secs) { this.slow = false; this.n = 0; return 'off'; } }
    return '';
  }
}

// ---- drawing: one canvas, no allocation per draw beyond a few strings; call at 4 to 10 Hz, never per picture ----
// Colours: the Okabe-Ito set (distinguishable with the common colour-vision deficiencies) on the overlay's own dark card, so the
// contrast does not depend on the page theme; hitches are vermillion AND drawn taller, so colour is never the only cue.
export const PALETTE = { bg: 'rgba(12,12,14,0.80)', fg: '#f2f2f2', dim: '#b9b9bd', grid: 'rgba(255,255,255,0.16)', deliv: '#56B4E9', shown: '#E69F00',
  drop: '#CC79A7', bytes: 'rgba(86,180,233,0.55)', hitch: '#D55E00', key: '#F0E442', thr: '#009E73', ok: '#a6d8a6' };

function hms(ms) { const d = new Date(ms), p = (n) => (n < 10 ? '0' : '') + n; return p(d.getHours()) + ':' + p(d.getMinutes()) + ':' + p(d.getSeconds()); }
function tzAbbr(ms) { try { const m = /\(([^)]+)\)/.exec(new Date(ms).toString()); const a = m && m[1].split(/\s+/).map((w) => w[0]).join(''); return (a && a.length <= 5 ? a : '') || 'UTC' + (-new Date(ms).getTimezoneOffset() / 60 >= 0 ? '+' : '') + (-new Date(ms).getTimezoneOffset() / 60); } catch (e) { return ''; } }

// draw paints the monitor on ctx (a 2D context of w x h CSS pixels, already scaled for devicePixelRatio). info: {player, build, view,
// target, ...}. scratch is a reusable series object. compact: the phone variant (smaller, no frame-time strip).
export function draw(ctx, w, h, mon, info, scratch, compact) {
  const P = PALETTE, now = mon.clock(), ser = mon.series(scratch), sum = mon.summary();
  ctx.clearRect(0, 0, w, h);
  ctx.fillStyle = P.bg; ctx.fillRect(0, 0, w, h);
  const px = compact ? 10 : 11, pad = 6, L = 30, R = w - pad, W = R - L;
  ctx.font = px + 'px ui-monospace, Menlo, Consolas, monospace'; ctx.textBaseline = 'top';
  const endMs = ser.endSec * 1000 + 999, startMs = endMs - WINDOW_S * 1000 + 1;
  // header: window, length, build, player, view
  ctx.fillStyle = P.fg;
  ctx.fillText(hms(startMs) + ' - ' + hms(now) + ' ' + tzAbbr(now) + '  ' + WINDOW_S + ' s', pad, pad);
  ctx.fillStyle = P.dim;
  ctx.fillText((info.player || '?') + ' | ' + (info.build || 'build ?') + ' | ' + (info.view || ''), pad, pad + px + 2);
  let y = pad + 2 * (px + 2) + 2;
  const xs = (k) => L + (k + 0.5) * W / WINDOW_S;
  // 1. fps lines: delivered, shown, dropped; the target as a dashed line
  const hF = compact ? 44 : 58;
  let top = Math.max(30, mon.target ? mon.target * 1.2 : 0); for (let k = 0; k < WINDOW_S; k++) top = Math.max(top, ser.deliv[k], ser.shown[k]);
  top = Math.ceil(top / 10) * 10;
  const yf = (v) => y + hF - Math.min(v, top) / top * hF;
  ctx.strokeStyle = P.grid; ctx.lineWidth = 1; ctx.beginPath(); ctx.moveTo(L, y + hF + 0.5); ctx.lineTo(R, y + hF + 0.5); ctx.stroke();
  ctx.fillStyle = P.dim; ctx.fillText(String(top), 2, y - 1); ctx.fillText('fps', 2, y + hF - px);
  if (mon.target) { ctx.setLineDash && ctx.setLineDash([3, 3]); ctx.strokeStyle = P.ok; ctx.beginPath(); ctx.moveTo(L, yf(mon.target)); ctx.lineTo(R, yf(mon.target)); ctx.stroke(); ctx.setLineDash && ctx.setLineDash([]); }
  const line = (a, col, wd) => { ctx.strokeStyle = col; ctx.lineWidth = wd; ctx.beginPath(); for (let k = 0; k < WINDOW_S; k++) { k ? ctx.lineTo(xs(k), yf(a[k])) : ctx.moveTo(xs(k), yf(a[k])); } ctx.stroke(); };
  line(ser.deliv, P.deliv, 1.5); line(ser.shown, P.shown, 2); line(ser.drop, P.drop, 1);
  y += hF + 6;
  // 2. kB/s bars with key-frame ticks and the server throttle share as a green line (0 to 100 percent)
  const hB = compact ? 24 : 32; let topB = 64; for (let k = 0; k < WINDOW_S; k++) topB = Math.max(topB, ser.kB[k]); topB = Math.ceil(topB / 64) * 64;
  ctx.fillStyle = P.bytes; for (let k = 0; k < WINDOW_S; k++) { const bh = ser.kB[k] / topB * hB; ctx.fillRect(xs(k) - W / WINDOW_S * 0.4, y + hB - bh, W / WINDOW_S * 0.8, bh); }
  ctx.fillStyle = P.key; for (let k = 0; k < WINDOW_S; k++) if (ser.keys[k]) ctx.fillRect(xs(k) - 2, y, 4, 3);
  let hasThr = false; for (let k = 0; k < WINDOW_S; k++) if (ser.thr[k] === ser.thr[k]) hasThr = true;
  if (hasThr) { ctx.strokeStyle = P.thr; ctx.lineWidth = 1.5; ctx.beginPath(); let st = false; for (let k = 0; k < WINDOW_S; k++) { const v = ser.thr[k]; if (v !== v) { st = false; continue; } const yy = y + hB - Math.min(v, 100) / 100 * hB; st ? ctx.lineTo(xs(k), yy) : ctx.moveTo(xs(k), yy); st = true; } ctx.stroke(); }
  ctx.fillStyle = P.dim; ctx.fillText('kB/s ' + topB, 2, y - 1);
  y += hB + 6;
  // 3. frame-time strip: one tick per shown frame, height = the gap (ms), vermillion and tall above 50 ms
  if (!compact) {
    const hS = 34, cap = 120; ctx.strokeStyle = P.grid; ctx.beginPath(); const y50 = y + hS - HITCH_MS / cap * hS; ctx.moveTo(L, y50); ctx.lineTo(R, y50); ctx.stroke();
    const from = now - WINDOW_S * 1000;
    for (let k = 0; k < mon.gapN; k++) { const j = (mon.gapHead - 1 - k + MAX_GAPS) % MAX_GAPS, t = mon.gapT[j]; if (t < from) break; const g = mon.gapV[j], x = L + (t - from) / (WINDOW_S * 1000) * W;
      ctx.fillStyle = g > HITCH_MS ? P.hitch : P.shown; const gh = Math.min(g, cap) / cap * hS; ctx.fillRect(x, y + hS - gh, g > HITCH_MS ? 2 : 1, gh); }
    ctx.fillStyle = P.dim; ctx.fillText('gap ms', 2, y - 1); ctx.fillText('50', 6, y50 - px);
    y += hS + 6;
  }
  // 4. presentation delay line (ms)
  const hD = compact ? 18 : 24; let topD = 100; for (let k = 0; k < WINDOW_S; k++) if (ser.delay[k] === ser.delay[k]) topD = Math.max(topD, ser.delay[k]); topD = Math.ceil(topD / 50) * 50;
  ctx.strokeStyle = P.deliv; ctx.lineWidth = 1.2; ctx.beginPath(); let st = false;
  for (let k = 0; k < WINDOW_S; k++) { const v = ser.delay[k]; if (v !== v) { st = false; continue; } const yy = y + hD - v / topD * hD; st ? ctx.lineTo(xs(k), yy) : ctx.moveTo(xs(k), yy); st = true; } ctx.stroke();
  ctx.fillStyle = P.dim; ctx.fillText('delay ms ' + topD, 2, y - 1);
  y += hD + 4;
  // time axis: ticks every 5 s, labels at both ends, and "now"
  ctx.strokeStyle = P.dim; ctx.lineWidth = 1; ctx.beginPath();
  for (let s = 0; s <= WINDOW_S; s += 5) { const x = L + s / WINDOW_S * W; ctx.moveTo(x, y); ctx.lineTo(x, y + (s % 10 === 0 ? 5 : 3)); }
  ctx.moveTo(L, y); ctx.lineTo(R, y); ctx.stroke();
  ctx.fillStyle = P.dim; ctx.textAlign = 'left'; ctx.fillText(hms(startMs), L, y + 6); ctx.textAlign = 'right'; ctx.fillText(hms(now) + ' now', R, y + 6); ctx.textAlign = 'left';
  y += px + 10;
  // events with their times: hitch, resync, key, hold, fallback (the most recent three, newest last)
  const ev = mon.events([]); let k0 = Math.max(0, ev.length - 3), line2 = '';
  for (let k = k0; k < ev.length; k++) line2 += (line2 ? '  ' : '') + hms(ev[k].t) + ' ' + ev[k].kind + (ev[k].label ? ' ' + ev[k].label : '');
  // summary line
  ctx.fillStyle = sum.hitches ? P.hitch : P.fg;
  ctx.fillText(sum.fps + ' fps  ' + sum.kBps + ' kB/s  gap p95 ' + Math.round(sum.p95) + ' ms  hitches ' + sum.hitches, pad, y);
  y += px + 3; ctx.fillStyle = P.dim; ctx.fillText(line2 || 'no events', pad, y);
  return y + px + pad;   // the height used
}
