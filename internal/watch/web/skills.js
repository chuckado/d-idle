// Skills page (docs/skilltrees.md): the hero's class skill trees from GET /api/skills. Read only. Text through textContent
// only; the only inline style writes are the tooltip transform (CSSOM, allowed under style-src 'self'); tile placement is
// by class names. Icons are the user's own game art decoded at request time (/watch/art/skill-<class>-<cel>.png, see
// docs/paperdoll-art-build.md); a missing icon leaves a text tile. The tooltip is one element, refilled on hover or focus.
(function () {
  'use strict';
  var $ = function (id) { return document.getElementById(id); };
  var view = null, sig = '', tip = null, openId = null, kindSel = 'all';
  // Skill kinds (n.kind from skills.txt flags, internal/skilltree KindOf): label, glyph, one-line explanation. A node with
  // no known kind gets no badge and is counted only under All.
  var KINDS = [
    { k: 'active', label: 'Active', glyph: '\u25B6', note: 'cast or used by the hero' },
    { k: 'aura', label: 'Aura', glyph: '\u25CE', note: 'switched on once, stays on' },
    { k: 'passive', label: 'Passive', glyph: '\u25C6', note: 'always on, nothing to cast' }
  ];
  function kindInfo(k) { for (var i = 0; i < KINDS.length; i++) if (KINDS[i].k === k) return KINDS[i]; return null; }
  try { var saved = localStorage.getItem('sk.kind'); if (saved === 'all' || kindInfo(saved)) kindSel = saved; } catch (e) { /* storage blocked */ }

  function el(t, c, x) { var e = document.createElement(t); if (c) e.className = c; if (x !== undefined) e.textContent = x; return e; }
  function clear(n) { while (n.firstChild) n.removeChild(n.firstChild); }
  function say(t) { var m = $('msg'); m.textContent = t || ''; m.hidden = !t; }
  function pct(v, of) { return (Math.round(v / of * 10000) / 100) + '%'; }
  function num(x) { return Math.round(x * 100) / 100; }
  function abbr(n) { return n.split(/\s+/).map(function (w) { return w.charAt(0); }).join('').slice(0, 3).toUpperCase(); }

  // lines returns [{k, t}] for one node: the same facts the page tiles show plus what the tables give.
  function lines(n) {
    var L = [], add = function (k, t) { if (t) L.push({ k: k, t: t }); };
    add('desc', n.short);
    add('desc', n.long && n.long !== n.short ? n.long : '');
    var ki = kindInfo(n.kind); add('type', ki ? ki.label + ': ' + ki.note : '');
    add('req', 'Required level: ' + n.reqLevel);
    (n.prereqs || []).forEach(function (p) { add(p.have ? 'req' : 'bad', 'Requires: ' + p.name + (p.have ? '' : ' (not learned)')); });
    var cur = 'Skill level: ' + n.base + (n.bonus ? ' (+' + n.bonus + ' from items, level ' + n.level + ')' : '') + ' of ' + n.maxLevel;
    add(n.base || n.bonus ? 'good' : 'dim', cur);
    if (n.manaNow !== undefined) add('base', 'Mana cost now: ' + num(n.manaNow));
    if (n.manaNext !== undefined) add('base', 'Mana cost at level ' + (n.level + 1) + ': ' + num(n.manaNext));
    if (n.synergy && n.synergy.length) add('dim', 'Synergies: ' + n.synergy.join(', '));
    if (!n.canAdd && n.why && n.base === 0) add('bad', 'Cannot learn yet: ' + n.why);
    return L;
  }

  function tipEl() {
    if (tip) return tip;
    tip = el('div', 'inv-tip sk-tip'); tip.id = 'sk-tip'; tip.setAttribute('role', 'tooltip'); tip.hidden = true;
    document.body.appendChild(tip);
    return tip;
  }
  function showTip(tile, n) {
    var t = tipEl();
    clear(t);
    t.appendChild(el('div', 'tip-name q-normal', n.name));
    lines(n).forEach(function (l) { t.appendChild(el('div', 'tip-line', l.t)).setAttribute('data-k', l.k); });
    t.hidden = false; openId = n.id;
    tile.setAttribute('aria-describedby', 'sk-tip');
    var r = tile.getBoundingClientRect(), w = t.offsetWidth, h = t.offsetHeight, vw = document.documentElement.clientWidth, vh = window.innerHeight, m = 8;
    var x = r.right + m; if (x + w > vw - m) x = r.left - m - w; if (x < m) { x = Math.max(m, Math.min(vw - w - m, r.left)); }
    var y = Math.max(m, Math.min(vh - h - m, r.top));
    if (x < r.right && x + w > r.left) y = r.bottom + m + h < vh ? r.bottom + m : Math.max(m, r.top - m - h); // overlaps the tile: below or above
    t.style.transform = 'translate(' + Math.round(x) + 'px,' + Math.round(y) + 'px)';
  }
  function hideTip() {
    if (tip) tip.hidden = true;
    openId = null;
    var d = document.querySelectorAll('[aria-describedby="sk-tip"]');
    for (var i = 0; i < d.length; i++) d[i].removeAttribute('aria-describedby');
  }

  function tile(n, allocOn, tr) {
    var d = el('div', 'sk-tile');
    d.style.setProperty('--l', pct(n.x, tr.w)); d.style.setProperty('--t', pct(n.y, tr.h)); d.style.setProperty('--w', pct(n.w, tr.w));
    d.tabIndex = 0; d.setAttribute('role', 'group'); d.setAttribute('data-state', n.state); d.setAttribute('data-id', n.id);
    d.setAttribute('aria-label', n.name + ', level ' + n.base + (n.bonus ? ' plus ' + n.bonus : '') + ' of ' + n.maxLevel + (kindInfo(n.kind) ? ', ' + kindInfo(n.kind).label.toLowerCase() : '') + (n.state === 'locked' ? ', locked' : ''));
    if (allocOn && n.canAdd) d.setAttribute('data-act', '1');
    var ab = el('span', 'sk-ab', n.icon ? '' : abbr(n.name));
    if (n.icon) {
      var img = new Image(); img.alt = ''; img.decoding = 'async'; img.loading = 'lazy'; img.src = '/watch/art/' + n.icon + '.png';
      img.addEventListener('error', function () { img.remove(); ab.textContent = abbr(n.name); });
      d.appendChild(img);
    }
    d.appendChild(ab);
    var ki = kindInfo(n.kind);
    if (ki) {
      d.setAttribute('data-kind', ki.k);
      var b = el('span', 'sk-kind'); b.setAttribute('data-kind', ki.k);
      var g = el('span', 'sk-kg', ki.glyph); g.setAttribute('aria-hidden', 'true');
      b.appendChild(g); b.appendChild(el('span', 'sk-kt', ki.label));
      d.appendChild(b);
    }
    if (n.level > 0) d.appendChild(el('span', 'sk-lv', n.bonus ? n.base + '+' + n.bonus : String(n.base)));
    d.addEventListener('mouseenter', function () { showTip(d, n); });
    d.addEventListener('mouseleave', hideTip);
    d.addEventListener('focus', function () { showTip(d, n); });
    d.addEventListener('blur', hideTip);
    d.addEventListener('click', function (e) {
      if (d.getAttribute('data-act') === '1') { allocate(n.id); return; }
      if (e.pointerType === 'touch' || e.pointerType === 'pen') { if (openId === n.id) hideTip(); else showTip(d, n); }
    });
    return d;
  }

  // The allocation control spends one point through POST /api/hero/points/skill (docs/points.md). A skill the hero cannot
  // use yet that is an attack asks once: the reply says needsConfirm and the page asks the player before it sends confirm.
  function allocate(id, confirm) {
    fetch('/api/config', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (c) {
      var u = c.answerBase ? c.answerBase : location.protocol + '//' + location.hostname + ':' + c.answerPort;
      return fetch(u + '/api/hero/points/skill', { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ skill: id, confirm: !!confirm }) });
    }).then(function (r) {
      return r.json().then(function (j) {
        if (j.needsConfirm && !confirm) { if (window.confirm(j.reason)) allocate(id, true); return; }
        if (!r.ok) throw new Error(j.reason || j.error || ('error ' + r.status));
        say(''); load(true);
      });
    }).catch(function (e) { say(e.message); });
  }

  var SVGNS = 'http://www.w3.org/2000/svg';
  // arrows draws the prerequisite lines from the numbers the server sent (edge midpoints of the two tiles, grid units).
  // One SVG with the board's own viewBox, stretched to the board like the tiles are: nothing is measured, so the first
  // paint is final and a resize or a late image changes nothing.
  function arrows(tree) {
    var svg = document.createElementNS(SVGNS, 'svg');
    svg.setAttribute('class', 'sk-arrows'); svg.setAttribute('aria-hidden', 'true');
    svg.setAttribute('viewBox', '0 0 ' + tree.w + ' ' + tree.h); svg.setAttribute('preserveAspectRatio', 'none');
    (tree.links || []).forEach(function (l) {
      var have = l.have ? '1' : '0', ln = document.createElementNS(SVGNS, 'line');
      ['x1', 'y1', 'x2', 'y2'].forEach(function (k) { ln.setAttribute(k, l[k]); });
      ln.setAttribute('data-have', have); ln.setAttribute('data-from', l.from); ln.setAttribute('data-to', l.to);
      svg.appendChild(ln);
      var hd = document.createElementNS(SVGNS, 'polygon'), dx = l.x2 - l.x1, dy = l.y2 - l.y1, len = Math.sqrt(dx * dx + dy * dy) || 1;
      var ux = dx / len, uy = dy / len, bx = l.x2 - ux * 9, by = l.y2 - uy * 9;
      var r2 = function (v) { return Math.round(v * 100) / 100; };
      hd.setAttribute('points', r2(l.x2) + ',' + r2(l.y2) + ' ' + r2(bx - uy * 4.5) + ',' + r2(by + ux * 4.5) + ' ' + r2(bx + uy * 4.5) + ',' + r2(by - ux * 4.5));
      hd.setAttribute('data-have', have);
      svg.appendChild(hd);
    });
    return svg;
  }

  // kindBar draws the filter chips with their counts (All first); the counts cover every tree of the hero.
  function kindBar(v) {
    var bar = $('kindbar'); if (!bar) return;
    clear(bar);
    var counts = { all: 0 };
    KINDS.forEach(function (x) { counts[x.k] = 0; });
    (v.trees || []).forEach(function (tr) { tr.nodes.forEach(function (n) { counts.all++; if (kindInfo(n.kind)) counts[n.kind]++; }); });
    [{ k: 'all', label: 'All' }].concat(KINDS).forEach(function (x) {
      var c = el('button', 'd-chip sk-chip'); c.type = 'button'; c.setAttribute('data-kind', x.k);
      c.setAttribute('aria-pressed', kindSel === x.k ? 'true' : 'false');
      c.appendChild(el('span', '', x.label)); c.appendChild(el('b', '', String(counts[x.k])));
      c.addEventListener('click', function () {
        kindSel = x.k;
        try { localStorage.setItem('sk.kind', kindSel); } catch (e) { /* storage blocked */ }
        render(view);
      });
      bar.appendChild(c);
    });
  }

  function render(v) {
    var host = $('trees'); clear(host); hideTip();
    kindBar(v);
    var meta = $('summeta');
    if (!v.available) {
      meta.textContent = '';
      host.appendChild(el('p', 'd-hint', 'No skill trees yet: the hero or its class is not known.'));
      return;
    }
    meta.textContent = v.class + ', level ' + v.level + ' | unspent points: ' + v.unspent + ' | spent: ' + v.spent;
    v.trees.forEach(function (tr, i) {
      var card = el('section', 'd-card'), head = el('div', 'd-card-head'), h = el('h2', '', 'Tree ' + (i + 1));
      h.id = 'tree-' + tr.tab; card.setAttribute('aria-labelledby', h.id);
      head.appendChild(h); head.appendChild(el('span', 'meta', tr.spent + ' point' + (tr.spent === 1 ? '' : 's') + ' spent'));
      card.appendChild(head);
      var board = el('div', 'sk-board'); if (kindSel !== 'all') board.setAttribute('data-filtered', '1'); board.setAttribute('role', 'group'); board.setAttribute('aria-label', 'Skills of tree ' + (i + 1));
      board.appendChild(arrows(tr));
      tr.nodes.forEach(function (n) {
        var t = tile(n, v.allocate, tr);
        if (kindSel !== 'all' && n.kind !== kindSel) t.hidden = true;
        board.appendChild(t);
      });
      card.appendChild(board); host.appendChild(card);
    });
  }
  function load(force) {
    return fetch('/api/skills', { cache: 'no-store' }).then(function (r) {
      if (!r.ok) throw new Error(r.status === 503 ? 'The skills page is off or the hero is not known yet.' : 'Could not read the skill trees (' + r.status + ').');
      return r.text();
    }).then(function (txt) {
      say('');
      if (txt === sig && !force) return; // unchanged: no rewrite, no reflow, an open tooltip stays
      sig = txt; view = JSON.parse(txt); render(view);
    }).catch(function (e) { say(e.message); });
  }
  document.addEventListener('keydown', function (e) { if (e.key === 'Escape') hideTip(); });
  document.addEventListener('click', function (e) { if (!e.target.closest || !e.target.closest('.sk-tile')) hideTip(); });
  load(); setInterval(load, 5000);
})();
