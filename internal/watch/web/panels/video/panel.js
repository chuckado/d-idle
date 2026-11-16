import { videoController, openStream, streamingSupported, registerVideo } from '../../mjpeg.js';

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
  frame.append(shot, cv, wait, badge, retry);
  ctx.body.append(frame);

  const full = ctx.el('button', 'd-btn', 'Full screen');
  full.type = 'button'; full.dataset.size = 'sm'; full.dataset.key = 'full';
  full.setAttribute('aria-label', 'Full screen');
  const col = ctx.el('button', 'd-btn', 'Collapse');
  col.type = 'button'; col.dataset.size = 'sm'; col.dataset.key = 'collapse';
  ctx.actions.append(full, col);

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
  ctx.actions.append(lowBtn);
  const hint = ctx.el('p', 'd-empty', '');
  hint.hidden = true; hint.setAttribute('role', 'status');
  ctx.body.append(hint);
  let slowSince = 0, bytes = [];
  function setLow(on) {
    low.set(on ? '1' : '0');
    lowBtn.setAttribute('aria-pressed', String(on));
    hint.hidden = true; slowSince = 0;
    ctl.reconnect();
  }
  lowBtn.addEventListener('click', () => setLow(low.get() !== '1'));
  // measured here and shown in the card head (also read by window.d2video.stats(): a friend can screenshot it)
  function stats() {
    const now = Date.now();
    while (bytes.length && now - bytes[0][0] > 2000) bytes.shift();
    const kb = bytes.reduce((a, b) => a + b[1], 0) / 2 / 1024;
    return { fps: Math.round(ctl.fps() * 10) / 10, kBps: Math.round(kb), low: low.get() === '1' };
  }
  function watchRate() {
    const st = stats();
    if (ctl.state === 'live' && st.fps < HINT_FPS && !st.low) {
      if (!slowSince) slowSince = Date.now();
      if (Date.now() - slowSince >= HINT_AFTER && hint.hidden) {
        hint.textContent = 'The picture runs at ' + Math.round(st.fps) + ' fps here. Low-rate mode may look smoother. ';
        const b = ctx.el('button', 'd-btn', 'Use low-rate mode'); b.type = 'button'; b.dataset.size = 'sm';
        b.addEventListener('click', () => setLow(true));
        hint.append(b); hint.hidden = false;
      }
    } else { slowSince = 0; if (!hint.hidden && st.fps >= HINT_FPS) hint.hidden = true; }
    // the rate line shows while slow or with ?debug; a steady line elsewhere would only repaint the head
    const dbg = typeof location !== 'undefined' && /[?&]debug\b/.test(location.search || '');
    const text = (dbg || (ctl.state === 'live' && st.fps < HINT_FPS)) ? 'stream: ' + st.fps + ' fps, ' + st.kBps + ' KB/s' + (st.low ? ', low rate' : '') : '';
    if (ctx.meta.textContent !== text) ctx.meta.textContent = text;
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
  function palFit() {
    const P = PF();
    if (!P || !P.fit || mode !== 'pal' || !frame.clientWidth) return;
    const r = P.fit(cv.width, cv.height, frame.clientWidth, frame.clientHeight, (typeof window !== 'undefined' && window.devicePixelRatio) || 1);
    cv.style.inlineSize = r.cssW + 'px'; cv.style.blockSize = r.cssH + 'px'; // whole device pixels per frame pixel: sharp at any zoom
  }
  function toMjpeg(why) {
    if (mode === 'mjpeg') return;
    log('falling back to MJPEG: ' + why);
    mode = 'mjpeg'; palFails = 0;
    cv.hidden = true;
    ctl.reconnect();
  }
  function openPal(onFrame, onEnd) {
    const P = PF();
    player = new P.Player({
      canvas: cv, url: src(palUrl),
      onFrame: () => { palFails = 0; if (cv.hidden) { cv.hidden = false; shot.hidden = true; wait.hidden = true; palFit(); } onFrame(); },
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
      const tq = ctx.token ? '?t=' + encodeURIComponent(ctx.token) : '';
      const g = await fetch('/api/glance' + tq, { cache: 'no-store', credentials: 'same-origin' });
      const fm = g.ok ? ((await g.json()).stream || {}).formats : null;
      if (!fm || fm.indexOf('palfeed') < 0) return;
      if (!PF()) await import('../../palfeed.js');
      const P = PF();
      if (!P || typeof DecompressionStream !== 'function' || typeof ImageData !== 'function') { log('lossless feed offered, this browser cannot decode it'); return; }
      let d = null; // 503 until the first frame: the probe itself wakes the ring, so ask again a few times
      for (let i = 0; i < 5 && !d; i++) { if (i) await new Promise((r) => setTimeout(r, 1000)); d = await P.detect(fetch, '', {}, tq); }
      if (!d) { log('lossless feed offered but not ready, staying on MJPEG'); return; }
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
    return openStream(src(), (b, type) => { bytes.push([Date.now(), b.length]); show(b, type); onFrame(); }, onEnd);
  }
  // <img> fallback: the browser plays the MJPEG itself; 'error' (refused, ended) is the only end signal
  let imgEnd = null, imgFrame = null;
  function openImg(onFrame, onEnd) {
    imgFrame = onFrame; imgEnd = onEnd;
    shot.src = src();
    return () => { imgEnd = imgFrame = null; };
  }
  shot.addEventListener('load', () => {
    if (streaming) { blobErrs = 0; decoded(); } else if (imgFrame) imgFrame();
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
  ctl.start();
  tryPal();
  return {};
}
