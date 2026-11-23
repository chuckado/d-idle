/* Update banner (TASK-122). Self-contained: polls /api/update/status on this origin, posts to the answer route.
   Never silent: nothing is downloaded or applied without a click. */
(function () {
  'use strict';
  var answerBase = '', bar = null, busy = false, later = '';
  function el(tag, text, cls) { var e = document.createElement(tag); if (text) e.textContent = text; if (cls) e.className = cls; return e; }
  function ensure() {
    if (bar) return bar;
    bar = el('div', '', 'update-banner');
    bar.setAttribute('role', 'status');
    bar.style.cssText = 'position:fixed;left:0;right:0;bottom:0;z-index:50;padding:8px 12px;background:#222;color:#eee;font:14px sans-serif;display:none;gap:8px;align-items:center;flex-wrap:wrap';
    document.body.appendChild(bar);
    return bar;
  }
  function post(path, body) {
    return fetch(answerBase + path, {
      method: 'POST', credentials: 'include',
      headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body || {})
    }).then(function (r) { return r.json(); }).catch(function () { return { error: 'the answer route is not reachable' }; });
  }
  function button(label, fn) {
    var b = el('button', label);
    b.style.cssText = 'padding:4px 10px;background:#444;color:#fff;border:1px solid #888;border-radius:4px;font:inherit;cursor:pointer;min-height:32px';
    b.onclick = function () { if (busy) return; busy = true; fn().then(function () { busy = false; refresh(); }); };
    return b;
  }
  function render(s) {
    var b = ensure(), a = s.available, lines = [];
    b.textContent = '';
    if (s.rollback) {
      b.appendChild(el('span', 'Update ' + s.rollback.from + ' failed and was rolled back: ' + s.rollback.reason));
      b.appendChild(button('Dismiss', function () { return post('/api/update/seen'); }));
    } else if (s.state === 'downloading') {
      b.appendChild(el('span', 'Downloading ' + (s.step || '') + ' ' + Math.round((s.progress || 0) * 100) + '%'));
      b.appendChild(button('Cancel', function () { return post('/api/update/cancel'); }));
    } else if (s.state === 'applying') {
      b.appendChild(el('span', 'Applying the update (' + (s.step || '') + '). The page reconnects when the bot is back.'));
    } else if (s.state === 'ready' && a) {
      b.appendChild(el('span', 'Update ' + a.version + ' is downloaded and verified. ' + (a.notes || '')));
      b.appendChild(button('Install and restart', function () {
        return post('/api/update/apply', { confirm: a.version }).then(function (j) { if (j && j.error && window.dSheet) { window.dSheet.notice(document, { title: 'The update did not start', lines: [j.error] }); } });
      }));
      b.appendChild(button('Cancel', function () { return post('/api/update/cancel'); }));
    } else if (a && a.needs_image) {
      b.appendChild(el('span', 'Update ' + a.version + ' needs a new image: run docker compose pull, then docker compose up -d.'));
    } else if (a && later === a.version) {
      b.style.display = 'none';
      return space();
    } else if (a) {
      b.appendChild(el('span', 'Update available: ' + a.version + '. ' + (a.notes || '')));
      b.appendChild(button('Download', function () { return post('/api/update/download'); }));
      b.appendChild(button('Later', function () { later = a.version; return Promise.resolve(); }));
    } else if (s.state === 'error' && s.last_error) {
      b.appendChild(el('span', 'Update problem: ' + s.last_error));
      b.appendChild(button('Dismiss', function () { return post('/api/update/cancel'); }));
    } else {
      b.style.display = 'none';
      return space();
    }
    b.style.display = 'flex';
    space();
  }
  // the watch cockpit is a fixed frame: it gives the banner's height back to the page (--d-banner-h, shell.css) instead of lying under it
  function space() {
    var h = bar && bar.style.display !== 'none' ? bar.offsetHeight : 0;
    document.documentElement.style.setProperty('--d-banner-h', h + 'px');
  }
  window.addEventListener('resize', function () { space(); });
  function refresh() {
    fetch('/api/update/status', { credentials: 'include' }).then(function (r) { return r.json(); }).then(render).catch(function () {});
  }
  fetch('/api/config', { credentials: 'include' }).then(function (r) { return r.json(); }).then(function (c) {
    answerBase = c.answerBase || (c.answerPort ? location.protocol + '//' + location.hostname + ':' + c.answerPort : '');
  }).catch(function () {}).then(function () {
    refresh();
    setInterval(refresh, 5000);
  });
})();
