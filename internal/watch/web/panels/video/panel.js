import { videoController, openStream, streamingSupported, registerVideo } from '../../mjpeg.js';
import { questCaption } from '../quest-board/panel.js';
import { StreamMonitor, SlowGate, draw as drawStats, HEIGHT_FULL, HEIGHT_COMPACT } from '../../streamstats.js';

// Game video (stage band), shown only when the server offers a stream (/api/config frameUrl). The stream is read by
// web/mjpeg.js (fetch, split into JPEG parts), so the page knows when each frame arrives: no frame for 3 s while the tab
// is visible tears the stream down and reconnects (1 s, 2 s, 4 s ... 10 s), showing "reconnecting" instead of a frozen
// picture and "stream down" with a Retry button after five failures in a row. A tab hidden for 10 s closes the stream
// (the browser throttles it anyway); visible, focus, pageshow and online reopen it at once. window.d2video.reconnect()
// lets the desktop app reload it after a sleep. Browsers without streaming fetch keep the plain MJPEG <img>.
// The token, if any, rides along from the page address. Collapse is remembered, full screen asks the panel.
// The picture always fits its box: it keeps 4:3, scales to the width and the height the cell gives and sits centred.
export function mount(root, ctx) {
  const url = ctx.config && ctx.config.frameUrl;
  if (!url) { ctx.hide(true); return {}; }
  root.dataset.aspect = '4/3'; // the layout editor keeps the cell in the picture's shape (panel version 2)
  const frame = ctx.el('div', 'frame');
  const shot = ctx.el('img', 'shot');
  shot.alt = 'live game window';
  shot.width = 800; shot.height = 600;
  shot.hidden = true;
  const cv = ctx.el('canvas', 'shot pal'); // the lossless feed draws here (docs/video-feed.md); hidden until it delivers
  cv.hidden = true; cv.width = 800; cv.height = 600;
  cv.setAttribute('aria-label', 'live game window');
  const wait = ctx.el('p', 'd-empty', 'Waiting for the game picture...');
  const badge = ctx.el('span', 'badge', '');
  badge.hidden = true;
  badge.setAttribute('role', 'status');
  const retry = ctx.el('button', 'd-btn', 'Retry');
  retry.type = 'button'; retry.dataset.size = 'sm'; retry.hidden = true;
  const cap = ctx.el('div', 'qcap'); // the quest board's current step as a small caption on the picture's top-left edge (never over the HUD at the bottom)
  cap.hidden = true;
  cap.setAttribute('role', 'status');
  cap.tabIndex = 0; cap.setAttribute('aria-expanded', 'false');
  const capToggle = () => { const on = cap.dataset.open !== '1'; cap.dataset.open = on ? '1' : '0'; cap.setAttribute('aria-expanded', String(on)); };
  cap.addEventListener('click', capToggle);
  cap.addEventListener('keydown', (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); capToggle(); } });
  frame.append(shot, cv, wait, badge, retry, cap);
  ctx.body.append(frame);
  // The controls are a toolbar BELOW the picture, never over it (a phone's five buttons wrapped across the top of the picture and
  // covered the quest caption and the game's own text; the fit check now fails any control that touches a caption or other text of
  // the card). It wraps, so a button added later never overflows: add one with window.d2video.controls.add({key, label, title,
  // onClick, pressed}) (the Labels toggle uses it). Only Collapse stays in the card's title bar actions.
  const bar = ctx.el('div', 'vtools');
  bar.setAttribute('role', 'toolbar'); bar.setAttribute('aria-label', 'Video controls');
  bar.setAttribute('data-hscroll', ''); // one row that scrolls sideways when the card is narrow: the picture keeps its height
  ctx.body.append(bar);
  const statLine = ctx.el('span', 'vstat', ''); // the rate line (was the card head's meta text, which sat over the picture)
  statLine.setAttribute('role', 'status'); statLine.hidden = true;
  function addControl(def) {
    const b = ctx.el('button', 'd-btn', def.label);
    b.type = 'button'; b.dataset.size = 'sm'; b.dataset.key = def.key;
    if (def.title) b.title = def.title;
    if (def.pressed !== undefined) b.setAttribute('aria-pressed', String(!!def.pressed));
    if (def.onClick) b.addEventListener('click', def.onClick);
    bar.insertBefore ? bar.insertBefore(b, statLine) : bar.append(b);
    return b;
  }
  bar.append(statLine);


  const full = ctx.el('button', 'd-btn', 'Full screen');
  full.type = 'button'; full.dataset.size = 'sm'; full.dataset.key = 'full';
  full.setAttribute('aria-label', 'Full screen');
  const col = ctx.el('button', 'd-btn', 'Collapse');
  col.type = 'button'; col.dataset.size = 'sm'; col.dataset.key = 'collapse';
  bar.insertBefore ? bar.insertBefore(full, statLine) : bar.append(full);
  ctx.actions.append(col);

  const saved = ctx.store('collapsed');
  function collapse(on, remember) {
    ctx.body.hidden = on;
    col.setAttribute('aria-pressed', String(on));
    col.textContent = on ? 'Expand' : 'Collapse';
    if (remember) saved.set(on ? '1' : '0');
  }
  collapse(saved.get() === '1', false);
  col.addEventListener('click', () => collapse(!ctx.body.hidden, true));
  full.addEventListener('click', () => {
    if (document.fullscreenElement) document.exitFullscreen();
    else if (root.requestFullscreen) root.requestFullscreen();
  });

  // Low-rate mode (per browser): the server sends this page at most LOW_FPS parts a second, so a slow main thread or
  // graphics path (software compositing, a busy tab) spends less on decoding and drawing. The hint offers it when the page
  // measures under HINT_FPS for HINT_AFTER ms.
  const LOW_FPS = 8, HINT_FPS = 15, HINT_AFTER = 6000;
  const low = ctx.store('lowrate');
  const lowBtn = ctx.el('button', 'd-btn', 'Low rate');
  lowBtn.type = 'button'; lowBtn.dataset.size = 'sm'; lowBtn.dataset.key = 'lowrate';
  lowBtn.setAttribute('aria-pressed', String(low.get() === '1'));
  lowBtn.title = 'Smoother on a slow screen: about ' + LOW_FPS + ' pictures a second';
  bar.insertBefore ? bar.insertBefore(lowBtn, statLine) : bar.append(lowBtn);
  // the low-fps hint: ONE line under the toolbar (text with an ellipsis, the button never cut), so it cannot squeeze the picture
  const hint = ctx.el('div', 'vhint', '');
  const hintT = ctx.el('span', 'vhint-t', '');
  const hintBtn = ctx.el('button', 'd-btn', 'Use low-rate mode'); hintBtn.type = 'button'; hintBtn.dataset.size = 'sm';
  hint.hidden = true; hint.setAttribute('role', 'status');
  hint.append(hintT, hintBtn);
  ctx.body.append(hint);
  let bytes = [];
  const mon = new StreamMonitor(); // the 30 s record the Stats overlay draws; both players feed it
  function setLow(on) {
    low.set(on ? '1' : '0');
    lowBtn.setAttribute('aria-pressed', String(on));
    hint.hidden = true; slow = false; gate.slow = false; gate.n = 0;
    ctl.reconnect();
  }
  lowBtn.addEventListener('click', () => setLow(low.get() !== '1'));
  hintBtn.addEventListener('click', () => setLow(true));
  // measured here and shown in the card head (also read by window.d2video.stats(): a friend can screenshot it)
  function stats() {
    const now = Date.now();
    while (bytes.length && now - bytes[0][0] > 2000) bytes.shift();
    const kb = bytes.reduce((a, b) => a + b[1], 0) / 2 / 1024;
    const out = { fps: Math.round(ctl.fps() * 10) / 10, kBps: Math.round(kb), low: low.get() === '1' };
    if (mode === 'pal' && player && player.stats) {
      const ps = player.stats();   // shown = draws, dropped = decoded and merged into a later draw, jitter = sd of the draw intervals
      Object.assign(out, { shown: Math.round(ps.presentFps * 10) / 10, dropped: ps.dropped, jitterMs: Math.round(ps.jitterMs * 10) / 10, latencyMs: Math.round(ps.latencyMs), lossless: true });
    }
    return out;
  }
  // The slow state has hysteresis: on after 3 s below SLOW_ON fps, off after 3 s above SLOW_OFF, so one stall (or a 1 s
  // bucket that lands on 14 and then 16) cannot flip the hint and the rate line on and off. The fps is the mean of the
  // monitor's shown frames over the last 3 whole seconds (not the controller's 2 s window of arrivals).
  const SLOW_ON = 12, SLOW_OFF = 18, SLOW_SECS = 3;
  const gate = new SlowGate(SLOW_ON, SLOW_OFF, SLOW_SECS);
  let slow = false, slowAt = 0;
  const stamp = () => new Date().toISOString();
  function watchRate() {
    const st = stats();
    const avg = mon.fpsAvg(3), f = avg === null ? st.fps : avg;
    if (ctl.state === 'live') {
      const flip = gate.update(f);
      if (flip) { slow = flip === 'on'; if (slow) slowAt = Date.now(); mon.event('slow-' + flip, Math.round(f) + ' fps'); log(stamp() + ' slow ' + flip + ': ' + f.toFixed(1) + ' fps'); }
    }
    if (slow && !st.low) {
      if (Date.now() - slowAt >= 3000 && hint.hidden) {
        hintT.textContent = 'Runs at ' + Math.round(f) + ' fps here. Low rate may look smoother.';
        hint.hidden = false;
      }
    } else if (!hint.hidden && !slow) hint.hidden = true;
    // the rate line shows while slow or with ?debug; a steady line elsewhere would only repaint the head
    const dbg = typeof location !== 'undefined' && /[?&]debug\b/.test(location.search || '');
    const text = !ovOn && (dbg || slow) ? 'stream: ' + st.fps + ' fps, ' + st.kBps + ' KB/s' + (st.lossless ? ', shown ' + st.shown + ', dropped ' + st.dropped + ', jitter ' + st.jitterMs + ' ms, delay ' + st.latencyMs + ' ms' : '') + (st.low ? ', low rate' : '') : '';
    if (statLine.textContent !== text) { statLine.textContent = text; statLine.hidden = !text; }
    ctx.later(watchRate, 1000);
  }

  function src(base) {
    const q = [];
    if (ctx.token) q.push('t=' + encodeURIComponent(ctx.token));
    if (low.get() === '1') q.push('fps=' + LOW_FPS); // the server skips parts for this viewer only
    q.push('r=' + Date.now());
    return (base || url) + '?' + q.join('&');
  }

  // Lossless feed (docs/video-feed.md): used only when /api/config offers palFeedUrl, /api/glance lists "palfeed" under
  // stream.formats and the capability probe answers; every failure keeps (or returns to) the MJPEG path below, logged
  // to the console only. mode is read by the controller's open() on each (re)connect.
  let mode = 'mjpeg', palFails = 0, player = null, ro = null;
  const palUrl = ctx.config && ctx.config.palFeedUrl;
  const PF = () => (typeof window !== 'undefined' && window.PalFeed) || null;
  const log = (m) => { if (typeof console !== 'undefined' && console.info) console.info('video: ' + m); };
  // Stream monitor overlay (docs/video-feed.md section 12): off by default, the Stats button (remembered per browser) or ?pfdebug
  // turns it on. It draws a 30 s record of both players on one canvas at 5 Hz, in a corner that leaves the game HUD alone.
  const statsMem = ctx.store('stats');
  const statsBtn = ctx.el('button', 'd-btn', 'Stats');
  statsBtn.type = 'button'; statsBtn.dataset.size = 'sm'; statsBtn.dataset.key = 'stats';
  statsBtn.title = 'Show a 30 second chart of the picture: frames per second, kB/s, gaps and hitches';
  const copyBtn = ctx.el('button', 'd-btn', 'Copy 30 s');
  copyBtn.type = 'button'; copyBtn.dataset.size = 'sm'; copyBtn.dataset.key = 'statscopy'; copyBtn.hidden = true;
  copyBtn.title = 'Copy the last 30 seconds as JSON (times, fps, kB/s, gaps, events, server numbers)';
  bar.insertBefore ? bar.insertBefore(statsBtn, statLine) : bar.append(statsBtn);
  bar.insertBefore ? bar.insertBefore(copyBtn, statLine) : bar.append(copyBtn);
  const ov = ctx.el('canvas', 'statsov'); ov.hidden = true; ov.setAttribute('role', 'img');
  ov.setAttribute('aria-label', 'stream statistics of the last 30 seconds');
  frame.append(ov, badge, retry); // appended last: the badge and Retry stay above the overlay
  const scratch = { deliv: new Array(30).fill(0), shown: new Array(30).fill(0), drop: new Array(30).fill(0), kB: new Array(30).fill(0), keys: new Array(30).fill(0),
    delay: new Array(30).fill(NaN), thr: new Array(30).fill(NaN), cpu: new Array(30).fill(NaN), ring: new Array(30).fill(NaN), endSec: 0 };
  const pfdebug = typeof location !== 'undefined' && /[?&]pfdebug\b/.test(location.search || '');
  let ovOn = false, stopPoll = null, polledMode = '', lastHolds = -1, ovTick = 0;
  const viewInfo = () => (mode === 'pal' ? fitInfo : frame.clientWidth + 'x' + frame.clientHeight + ' @dpr' + (((typeof window !== 'undefined' && window.devicePixelRatio) || 1)));
  function ovPoll() {   // the host's numbers once a second, while the overlay is on and the lossless feed is the player
    const want = ovOn && mode === 'pal' ? 'pal' : '';
    if (want === polledMode) return;
    if (stopPoll) { stopPoll(); stopPoll = null; }
    polledMode = want;
    const P = PF();
    if (!want || !P || !P.poll) return;
    const tq = ctx.token ? '?t=' + encodeURIComponent(ctx.token) : '';
    stopPoll = P.poll(fetch, '/api/palfeed' + tq, (j) => {
      const st = j && j.stats; if (!st) return;
      mon.server(st.server);
      if (lastHolds >= 0 && st.holds > lastHolds) mon.event('hold', '+' + (st.holds - lastHolds));
      lastHolds = st.holds;
    }, 1000);
  }
  function ovDraw() {
    if (!ovOn) return;
    const w = Math.max(160, Math.min(380, frame.clientWidth - 12)), compact = frame.clientWidth < 520, h = compact ? HEIGHT_COMPACT : HEIGHT_FULL;
    const dpr = (typeof window !== 'undefined' && window.devicePixelRatio) || 1;
    if (ov.width !== Math.round(w * dpr) || ov.height !== Math.round(h * dpr)) {
      ov.width = Math.round(w * dpr); ov.height = Math.round(h * dpr);
      ov.style.setProperty('--ov-w', w + 'px'); ov.style.setProperty('--ov-h', h + 'px');
    }
    const g = ov.getContext('2d'); g.setTransform(dpr, 0, 0, dpr, 0, 0);
    drawStats(g, w, h, mon, { player: mon.player, build: ctx.config && ctx.config.build, view: viewInfo(), delay: player && player.stats ? player.stats().delayMs : 0 }, scratch, compact);
    ovPoll();
    ctx.later(ovDraw, 200);
  }
  function setStats(on, remember) {
    ovOn = on; ov.hidden = !on; copyBtn.hidden = !on;
    statsBtn.setAttribute('aria-pressed', String(on));
    if (remember) statsMem.set(on ? '1' : '0');
    if (on) { ovDraw(); } else { ovPoll(); }
  }
  statsBtn.addEventListener('click', () => setStats(!ovOn, true));
  copyBtn.addEventListener('click', () => {
    const txt = mon.json({ build: ctx.config && ctx.config.build, view: viewInfo(), page: typeof location !== 'undefined' ? location.pathname : '', userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : '' });
    const done = () => { copyBtn.textContent = 'Copied'; ctx.later(() => { copyBtn.textContent = 'Copy 30 s'; }, 1500); };
    if (typeof navigator !== 'undefined' && navigator.clipboard && navigator.clipboard.writeText) navigator.clipboard.writeText(txt).then(done, () => ctx.toast && ctx.toast('Could not copy'));
    else ctx.toast && ctx.toast('Clipboard not available here');
  });

  // The canvas fills the frame box (contain, 4:3 kept) at ANY scale, rounded to whole device pixels. Sharp pixels off (the
  // default): nearest-neighbour only at a whole-number scale of 1x or more, smooth otherwise (a fractional scale with
  // nearest-neighbour shimmers while the picture scrolls); on: always nearest-neighbour.
  let fitInfo = '';
  const sharp = ctx.store('sharp');
  const sharpBtn = ctx.el('button', 'd-btn', 'Sharp pixels');
  sharpBtn.type = 'button'; sharpBtn.dataset.size = 'sm'; sharpBtn.dataset.key = 'sharp';
  sharpBtn.setAttribute('aria-pressed', String(sharp.get() === '1'));
  sharpBtn.title = 'Hard pixel edges instead of smooth scaling (lossless picture only)';
  sharpBtn.hidden = true; // shown while the lossless picture is on
  bar.insertBefore ? bar.insertBefore(sharpBtn, statLine) : bar.append(sharpBtn);
  sharpBtn.addEventListener('click', () => {
    const on = sharp.get() !== '1';
    sharp.set(on ? '1' : '0'); sharpBtn.setAttribute('aria-pressed', String(on)); palFit();
  });
  function palFit() {
    const P = PF();
    if (!P || !P.fitFill || mode !== 'pal' || !frame.clientWidth) return;
    const r = P.fitFill(cv.width, cv.height, frame.clientWidth, frame.clientHeight, (typeof window !== 'undefined' && window.devicePixelRatio) || 1);
    cv.style.setProperty('--pal-w', r.cssW + 'px'); cv.style.setProperty('--pal-h', r.cssH + 'px'); fitInfo = cv.width + 'x' + cv.height + ' @' + (Math.round(r.scale * 100) / 100) + 'x dpr' + (((typeof window !== 'undefined' && window.devicePixelRatio) || 1));
    cv.dataset.smooth = (sharp.get() !== '1' && !(r.integer && r.scale >= 1)) ? '1' : '0';
  }
  function toMjpeg(why) {
    if (mode === 'mjpeg') return;
    log('falling back to MJPEG: ' + why); mon.event('fallback', why);
    mode = 'mjpeg'; palFails = 0;
    cv.hidden = true; sharpBtn.hidden = true;
    ctl.reconnect();
  }
  function openPal(onFrame, onEnd) {
    const P = PF();
    mon.player = 'palette'; mon.target = 30;
    player = new P.Player({
      canvas: cv, url: src(palUrl), mon,
      onFrame: () => { palFails = 0; if (cv.hidden) { cv.hidden = false; shot.hidden = true; wait.hidden = true; sharpBtn.hidden = false; palFit(); } onFrame(); },
      onBytes: (n) => bytes.push([Date.now(), n]),
      log,
      onEnd: (e) => {
        if (e && e.fatal) { toMjpeg(e.message); return; }
        if (++palFails >= 3) { toMjpeg('feed failed ' + palFails + ' times'); return; }
        onEnd(e);
      },
    });
    player.start();
    const p = player;
    return () => { p.stop(); if (player === p) player = null; };
  }
  async function tryPal() {
    if (!palUrl || typeof fetch !== 'function') return;
    try {
      if (!PF()) await import('../../palfeed.js');
      const P = PF();
      if (!P || typeof DecompressionStream !== 'function' || typeof ImageData !== 'function') { log('lossless feed: this browser cannot decode it'); return; }
      const d = await P.negotiate(fetch, ctx.token ? '?t=' + encodeURIComponent(ctx.token) : '');
      if (!d) return; // not offered (glance without palfeed) or not ready: MJPEG stays
      mode = 'pal';
      if (typeof ResizeObserver === 'function') { ro = new ResizeObserver(palFit); ro.observe(frame); }
      ctl.reconnect();
    } catch (e) { log('lossless feed not used: ' + (e && e.message || e)); }
  }

  // fetch path: each part becomes an object URL on the <img>; while one decodes only the newest waits (latest wins)
  let streaming = streamingSupported(), blobErrs = 0;
  let shown = '', busy = false, pending = null;
  function show(bytes, type) {
    if (busy) { pending = [bytes, type]; return; }
    busy = true;
    const u = URL.createObjectURL(new Blob([bytes], { type }));
    if (shown) URL.revokeObjectURL(shown);
    shown = u;
    shot.src = u;
  }
  function decoded() {
    busy = false;
    if (pending) { const p = pending; pending = null; show(p[0], p[1]); }
  }
  function openFetch(onFrame, onEnd) {
    mon.player = 'jpeg'; mon.target = 0;
    return openStream(src(), (b, type) => { bytes.push([Date.now(), b.length]); mon.arrive(b.length, false); show(b, type); onFrame(); }, onEnd);
  }
  // <img> fallback: the browser plays the MJPEG itself; 'error' (refused, ended) is the only end signal
  let imgEnd = null, imgFrame = null;
  function openImg(onFrame, onEnd) {
    imgFrame = onFrame; imgEnd = onEnd;
    shot.src = src();
    return () => { imgEnd = imgFrame = null; };
  }
  shot.addEventListener('load', () => {
    if (streaming) { blobErrs = 0; decoded(); mon.shownFrame(); } else if (imgFrame) imgFrame();
    shot.hidden = false; wait.hidden = true;
  });
  shot.addEventListener('error', () => {
    if (streaming) {
      decoded();
      // frames that never decode (a page policy without blob: images): fall back to the browser's own MJPEG <img>
      if (++blobErrs >= 3) { streaming = false; ctl.reconnect(); }
      return;
    }
    shot.hidden = true; wait.hidden = false;
    if (imgEnd) imgEnd();
  });

  const timers = new Map();
  let tick = 0;
  const ctl = videoController({
    stall: () => (streaming || mode === 'pal' ? 3000 : 0),
    open: (onFrame, onEnd) => (mode === 'pal' ? openPal(onFrame, onEnd) : streaming ? openFetch(onFrame, onEnd) : openImg(onFrame, onEnd)),
    later(fn, ms) { const id = ++tick; timers.set(id, true); ctx.later(() => { if (timers.delete(id)) fn(); }, ms); return id; },
    cancel(id) { timers.delete(id); },
    now: () => Date.now(),
    state(st) {
      ctx.state(st === 'live' ? 'ready' : st);
      if (st === 'reconnecting' || st === 'down' || st === 'paused') mon.event(st);
      badge.hidden = st === 'live' || st === 'connecting';
      badge.textContent = st === 'down' ? 'stream down' : st === 'paused' ? 'paused (tab hidden)' : st === 'reconnecting' ? 'reconnecting...' : '';
      retry.hidden = st !== 'down';
    },
  });
  retry.addEventListener('click', () => ctl.reconnect());
  ctl.stats = stats;
  watchRate();
  if (typeof document.addEventListener === 'function') {
    document.addEventListener('visibilitychange', () => ctl.visibility(document.visibilityState === 'hidden'));
  }
  if (typeof window.addEventListener === 'function') {
    window.addEventListener('focus', () => ctl.wake());
    window.addEventListener('online', () => ctl.wake());
    window.addEventListener('pageshow', (e) => { if (e && e.persisted) ctl.reconnect(); else ctl.wake(); });
  }
  registerVideo(ctl);
  if (typeof window !== 'undefined' && window.d2video) window.d2video.controls = { add: addControl };
  ctl.start();
  tryPal();
  if (pfdebug || statsMem.get() === '1') setStats(true, false);
  // the quest caption: quest-strip (the quest board's tile) publishes the text, this panel only shows it; without the tile (or in a
  // page where the import is not there) the caption stays empty
  if (typeof questCaption !== 'undefined') {
    const show = (t) => { if (cap.textContent !== t) { cap.textContent = t; cap.title = t; cap.dataset.open = '0'; } cap.hidden = !t; };
    questCaption.subs.add(show);
    show(questCaption.text);
  }
  return {};

}
