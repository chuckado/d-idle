// Item tooltip of the inventory page (docs/inventory-tooltip.md, docs/ui-design-system.md components 13 and 14): hover or
// keyboard focus on desktop, tap to open and tap away to close on a phone, for every item cell and filled doll slot.
// One element, created once and refilled; the page's 4 s poll only re-renders it when the open item's text changed, and
// nothing here scrolls the page: the tip is position: fixed, placed by a transform, and takes no focus.
// Input: the view /api/inventory sends, handed over by inventory.js as a 'd2:inventory' event on document. Text goes in
// through textContent only. The lines are the game's own tooltip text when the backend has it (item.tip, d2d "items"
// verb, S), else built from the game tables' base facts (item.base) and the props the snapshot carries.
(function () {
  'use strict';
  var TIP_ID = 'inv-tip';
  var ARMOR_LINE = /^(defen[cs]e|one-hand damage|two-hand damage|throw damage|damage|durability|required|quantity|item level|sockets?|socketed|ethereal|smite damage|chance to block|quality)/i;
  var lastType = 'mouse', view = null, openKey = '', openSig = '', mode = '', kbd = false, node = null, host = null;

  function el(t, c, x) { var e = document.createElement(t); if (c) e.className = c; if (x !== undefined) e.textContent = x; return e; }
  function clear(n) { while (n.firstChild) n.removeChild(n.firstChild); }

  // The colour class of the name: the quality, except that a plain item that is ethereal or socketed is grey like in
  // the game (low quality is grey too).
  function nameClass(it) {
    var q = it.quality || 'normal';
    if ((q === 'normal' || q === 'superior') && (it.ethereal || it.sockets)) return 'q-low';
    return 'q-' + q;
  }
  function range(a, b) { return a === b ? String(a) : a + ' to ' + b; }

  // build returns {cls, lines: [{k, t}], badges: [text]} for one item. slot is the doll slot key or ''.
  function build(it, slot) {
    var L = [], has = function (re) { return L.some(function (l) { return re.test(l.t); }); };
    var add = function (k, t) { if (t) L.push({ k: k, t: t }); };
    var b = it.base || null, tip = it.tip || [];
    var title = it.name || it.code || '';
    if (tip.length) {
      title = tip[0];
      tip.slice(1).forEach(function (t) { add(ARMOR_LINE.test(t) ? 'base' : 'prop', t); });
    } else {
      if (b && b.name && b.name !== title) add('type', b.name);
      if (b) {
        if (b.minAC || b.maxAC) add('base', 'Base defense: ' + range(b.minAC, b.maxAC));
        // weapons.txt: a true two-hander keeps its damage in the two-hand columns; everything else uses min and max
        if (b.twoHanded && b.twoMaxDam) add('base', 'Two-Hand damage: ' + range(b.twoMinDam, b.twoMaxDam));
        else if (b.minDam || b.maxDam) add('base', (b.twoHanded ? 'Two-Hand' : 'One-Hand') + ' damage: ' + range(b.minDam, b.maxDam));
        if (b.block) add('base', 'Base chance to block: ' + b.block + '%');
        if (b.speed) add('base', 'Base weapon speed: ' + b.speed);
      }
      if (it.maxDur) add('base', 'Durability: ' + (it.dur || 0) + ' of ' + it.maxDur + (!it.dur ? ' (broken)' : ''));
      else if (b && b.dur) add('base', 'Durability: ' + b.dur);
      if (b && b.reqStr) add('req', 'Required strength: ' + b.reqStr);
      if (b && b.reqDex) add('req', 'Required dexterity: ' + b.reqDex);
      if (b && b.reqLvl) add('req', 'Required level: ' + b.reqLvl);
      if (it.ethereal) add('eth', 'Ethereal (cannot be repaired)');
      if (it.sockets) add('sock', 'Socketed (' + (it.filled || 0) + '/' + it.sockets + ')');
      (it.props || []).forEach(function (p) { add('prop', p); });
    }
    if (!it.identified) add('bad', 'Unidentified');
    else if (!tip.length && !(it.props || []).length && (it.quality === 'magic' || it.quality === 'rare' || it.quality === 'set' || it.quality === 'unique' || it.quality === 'crafted')) {
      add('note', 'Magic properties are not read on this backend.');
    }
    if (it.sockets && tip.length && !has(/socket/i)) add('sock', 'Socketed (' + (it.filled || 0) + '/' + it.sockets + ')');
    if (it.ethereal && tip.length && !has(/ethereal/i)) add('eth', 'Ethereal (cannot be repaired)');
    if (it.level) add('meta', 'Item level: ' + it.level);
    if (it.sell) {
      var exact = !it.identified || it.quality === 'normal' || it.quality === 'low' || it.quality === 'superior';
      add('meta', (exact ? 'Sells for ' : 'Sells for at least ') + it.sell + ' gold');
    }
    var badges = [];
    if (it.mark === 'sell') badges.push('SELL');
    if (it.mark === 'keep') badges.push('KEEP');
    return { cls: nameClass(it), title: title, lines: L, badges: badges };
  }

  // place returns the top-left of the tip so that it sits beside the anchor (right, else left; on a window too narrow for
  // both, on the side above or below with more room) and stays inside the viewport with an 8 px margin. a: {left, top,
  // right, bottom}. Beside the anchor it never overlaps it; above or below it does not either unless the tip is taller than
  // either side (the CSS caps its height, so that is only a window smaller than the cell).
  function place(a, w, h, vw, vh) {
    var m = 8, x, y;
    h = Math.min(h, vh - 2 * m);
    if (a.right + m + w <= vw - m) { x = a.right + m; y = a.top; }
    else if (a.left - m - w >= m) { x = a.left - m - w; y = a.top; }
    else {
      x = Math.round((a.left + a.right) / 2 - w / 2);
      var below = vh - m - (a.bottom + m), above = a.top - m - m;
      y = h <= below || (h > above && below >= above) ? a.bottom + m : a.top - m - h;
    }
    x = Math.max(m, Math.min(x, vw - w - m));
    y = Math.max(m, Math.min(y, vh - h - m));
    return { x: Math.round(x), y: Math.round(y) };
  }
  // the phone and touch form: a bottom sheet, moved to the top when the anchor sits in the lower half so it never covers it
  function sheetEdge(a, vh) { return (a.top + a.bottom) / 2 > vh / 2 ? 'top' : 'bottom'; }
  function narrow() { return !!(typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(max-width: 40rem)').matches); }

  // ---- the character canvas: where the item is, who can use it, what it changes (card uic3) ----
  // Only on the canvas page (body[data-canvas]); the cards page keeps the plain tooltip. Every verdict comes from the view
  // (item.req, item.usable, item.canWear, view.queue, view.merc): the page words them and, for the comparison, subtracts
  // the numbers of two tooltip lines. It never decides whether an item is usable.
  var SLOTWORD = { head: 'helm', neck: 'amulet', torso: 'armour', rhand: 'right hand', lhand: 'left hand', rring: 'right ring', lring: 'left ring', belt: 'belt', feet: 'boots', gloves: 'gloves' };
  var MERCKIND = { rogue: 'Rogue', desert: 'Desert Guard', ironwolf: 'Iron Wolf', barbarian: 'Barbarian' };
  var STATWORD = { lvl: 'Level', str: 'Strength', dex: 'Dexterity' };
  var QWORD = { equip: 'Equip', unequip: 'Take off', trash: 'Drop', identify: 'Identify', stash: 'Move to the stash', take: 'Take out of the stash', sell: 'Sell' };
  function isRich() { return !!(document.body && document.body.dataset && document.body.dataset.canvas !== undefined); }
  function words(slot) { return SLOTWORD[slot] || slot; }

  function whereLine(f, id) {
    var t = f.place === 'worn' ? 'Worn by you, ' + words(f.slot) : f.place === 'stash' ? 'In the stash' : f.place === 'belt' ? 'On the belt' : f.place === 'cube' ? 'In the cube' : 'In the pack';
    var q = (view && view.queue || []).filter(function (r) { return r.id === id && (r.state === 'queued' || r.state === 'doing'); })[0], badge = '';
    if (q) {
      var k = q.kind || 'equip';
      t += ' · ' + (q.state === 'doing' ? 'doing: ' : 'queued: ') + (QWORD[k] || k);
      if (q.wait === 'town') { t += ' (waits for town)'; badge = 'TOWN'; } else if (q.wait === 'safe') t += ' (next safe moment)';
    }
    return { text: t, badge: badge };
  }

  // One row of a requirement block: ok is true only for a requirement the view says is met. A refusal (hard reason) is
  // never ok; a warning (broken) is neither: tone 'warn'.
  function reqRows(it, side) {
    var rows = [], reasons = side && side.reasons || [], req = it.req || {}, used = {};
    function gap(r) { var n = Number(r.need), h = Number(r.have); return n > 0 && h >= 0 && n > h ? n - h : 0; }
    reasons.forEach(function (r, i) {
      if (r.code === 'requirements' && r.stat) used[r.stat] = i;
      if (r.code === 'class') used.class = i;
    });
    ['lvl', 'str', 'dex'].forEach(function (k) {
      if (!req[k]) return;
      var i = used[k];
      if (i === undefined) rows.push({ ok: true, text: STATWORD[k] + ' ' + req[k] });
      else { var g = gap(reasons[i]); rows.push({ ok: false, text: STATWORD[k] + ' ' + req[k] + ' (' + (reasons[i].have === undefined ? '?' : reasons[i].have) + ')' + (g ? '  ' + g + ' short' : '') + (reasons[i].approx ? ' (about)' : '') }); }
    });
    if (req.class && req.class.length) {
      if (used.class === undefined) rows.push({ ok: true, text: req.class.join(', ') + ' only' });
      else rows.push({ ok: false, text: reasons[used.class].text || (req.class.join(', ') + ' only') });
    }
    reasons.forEach(function (r) {
      if (r.code === 'class' || (r.code === 'requirements' && r.stat)) return;
      rows.push(r.warn ? { ok: false, warn: true, text: r.text } : { ok: false, text: r.text });
    });
    if (side && !rows.length) rows.push({ ok: true, text: 'no requirements' });
    return rows;
  }
  function who(it, merc) {
    var u = it.usable || {}, out = [];
    if (u.hero) out.push({ who: 'hero', head: 'You', ok: u.hero.ok, rows: reqRows(it, u.hero) });
    if (u.merc && merc) out.push({ who: 'merc', head: (MERCKIND[merc.kind] || 'Mercenary') + (merc.level ? ' ' + merc.level : ''), ok: u.merc.ok, rows: reqRows(it, u.merc) });
    return out;
  }

  // The numbers of the display lines: defence, damage range, and "+N something" properties.
  function parseLine(t) {
    var m;
    if ((m = /^(?:base )?defen[cs]e:?\s*(\d+)(?:\s*(?:to|-)\s*(\d+))?/i.exec(t))) return { key: 'defense', label: 'Defense', v: m[2] ? (Number(m[1]) + Number(m[2])) / 2 : Number(m[1]), txt: m[2] ? m[1] + '-' + m[2] : m[1] };
    if ((m = /damage:\s*(\d+)\s*(?:to|-)\s*(\d+)/i.exec(t)) && !/^smite/i.test(t)) return { key: 'damage', label: 'Damage', v: (Number(m[1]) + Number(m[2])) / 2, txt: m[1] + '-' + m[2] };
    if ((m = /^([+-]\d+)%?\s+(?:to\s+)?(.+)$/.exec(t))) return { key: m[2].toLowerCase(), label: m[2], v: Number(m[1]), txt: m[1] };
    if ((m = /^(.+?):?\s+([+-]\d+)%?$/.exec(t))) return { key: m[1].toLowerCase(), label: m[1], v: Number(m[2]), txt: m[2] };
    return null;
  }
  function statsOf(it) {
    var out = {}, order = [];
    build(it, '').lines.forEach(function (l) {
      if (l.k === 'req' || l.k === 'meta' || l.k === 'sock' || l.k === 'eth' || l.k === 'note') return;
      var p = parseLine(l.t);
      if (p && !out[p.key]) { out[p.key] = p; order.push(p.key); }
    });
    return { m: out, order: order };
  }
  function signed(n) { n = Math.round(n * 10) / 10; return (n > 0 ? '+' : '') + n; }
  // compare lists, for one pair of items, each value as old to new with its signed change; values that only the worn item has
  // are shown going to nothing. Rows are cut to six: the tip never grows past the screen.
  function compare(nu, old) {
    var a = statsOf(nu), b = statsOf(old), keys = a.order.slice(), rows = [];
    b.order.forEach(function (k) { if (!a.m[k]) keys.push(k); });
    keys.forEach(function (k) {
      var n = a.m[k], o = b.m[k], d = (n ? n.v : 0) - (o ? o.v : 0);
      rows.push({ label: (n || o).label, old: o ? o.txt : '-', now: n ? n.txt : '-', delta: signed(d), tone: d > 0 ? 'up' : d < 0 ? 'dn' : 'eq' });
    });
    return rows.slice(0, 6);
  }
  // the slots an item would take that are worth comparing: those the view says it can go in now, else those whose refusal
  // is not the one most slots share ("does not fit that slot"); none when every slot says the same thing
  function targets(it, f) {
    if (f.place === 'worn' || !view || !view.equipment) return [];
    var can = it.canWear || [], why = it.noWearSlot || {}, n = {}, common = '', out = can.slice();
    Object.keys(why).forEach(function (k) { n[why[k]] = (n[why[k]] || 0) + 1; if (n[why[k]] > (n[common] || 0)) common = why[k]; });
    Object.keys(why).forEach(function (k) { if (why[k] !== common && out.indexOf(k) < 0) out.push(k); });
    return out.filter(function (k) { return view.equipment[k]; });
  }
  function against(it, f) {
    return targets(it, f).map(function (k) { var w = view.equipment[k]; return { slot: k, name: w.name || w.code || 'item', rows: compare(it, w) }; }).filter(function (c) { return c.rows.length; });
  }

  function richOf(it, f) {
    var merc = view && view.merc && view.merc.owned ? view.merc : null;
    return { where: whereLine(f, it.id), who: who(it, merc), against: f.place === 'worn' ? [] : against(it, f), hint: hintFor(f) };
  }
  function hintFor(f) {
    var a = typeof window !== 'undefined' && window.d2InvActions && window.d2InvActions.first;
    var k = f.place === 'worn' ? 'slot-' + f.slot : 'item-' + (f.it && f.it.id), e = a ? a(k) : null;
    return (e ? 'Double-click: ' + e.label + (e.reason ? ' (' + e.reason + ')' : '') + '. ' : '') + 'Right-click: all actions.';
  }

  function ensure() {
    if (node) return node;
    node = el('div', 'inv-tip'); node.id = TIP_ID; node.setAttribute('role', 'tooltip'); node.hidden = true;
    document.body.appendChild(node);
    return node;
  }
  function renderRich(n, r) {
    var w = el('div', 'tip-where', r.where.text); w.dataset.k = 'where'; n.appendChild(w);
    if (r.where.badge) { var bd = el('span', 'tip-badge', r.where.badge); bd.dataset.badge = 'town'; w.appendChild(bd); }
    r.who.forEach(function (x) {
      var sec = el('div', 'tip-sec'); sec.dataset.who = x.who; sec.dataset.ok = x.ok ? 'y' : 'n';
      sec.appendChild(el('div', 'tip-hd', x.head));
      x.rows.forEach(function (row) {
        var li = el('div', 'tip-req', (row.ok ? '\u2713 ' : row.warn ? '! ' : '\u2717 ') + row.text);
        li.dataset.ok = row.ok ? 'y' : row.warn ? 'w' : 'n'; sec.appendChild(li);
      });
      n.appendChild(sec);
    });
    r.against.forEach(function (c) {
      var sec = el('div', 'tip-sec'); sec.dataset.cmp = c.slot;
      sec.appendChild(el('div', 'tip-hd', 'Against ' + c.name + ' (' + words(c.slot) + ')'));
      c.rows.forEach(function (row) {
        var li = el('div', 'tip-cmp'); li.dataset.tone = row.tone;
        li.appendChild(el('span', 'k', row.label)); li.appendChild(el('span', 'v', row.old + ' \u2192 ' + row.now)); li.appendChild(el('b', 'd', row.delta));
        sec.appendChild(li);
      });
      n.appendChild(sec);
    });
    n.appendChild(el('div', 'tip-hint', r.hint));
  }
  function render(t) {
    var n = ensure(); clear(n);
    if (t.rich) n.dataset.rich = ''; else delete n.dataset.rich;
    var h = el('div', 'tip-name ' + t.cls, t.title); n.appendChild(h);
    t.lines.forEach(function (l) { var d = el('div', 'tip-line', l.t); d.dataset.k = l.k; n.appendChild(d); });
    if (t.badges.length) {
      var row = el('div', 'tip-badges');
      t.badges.forEach(function (x) { var s = el('span', 'tip-badge', x); s.dataset.badge = x.toLowerCase(); row.appendChild(s); });
      n.appendChild(row);
    }
    if (t.rich) renderRich(n, t.rich);
  }

  function find(key) {
    if (!view || !key) return null;
    var m = /^(item|slot)-(.+)$/.exec(key);
    if (!m) return null;
    if (m[1] === 'slot') return { it: view.equipment && view.equipment[m[2]], slot: m[2], place: 'worn' };
    var id = Number(m[2]), hit = null, pl = '';
    [['pack', view.backpack], ['stash', view.stash], ['belt', view.belt], ['cube', view.cube]].forEach(function (l) { (l[1] || []).forEach(function (x) { if (x.id === id) { hit = x; pl = l[0]; } }); });
    return { it: hit, slot: '', place: pl };
  }
  function anchorOf(key) { return document.querySelector('[data-key="' + key + '"]'); }
  function keyOf(target) {
    var a = target && target.closest ? target.closest('[data-key]') : null;
    if (!a || !a.dataset.key || !/^(item|slot)-/.test(a.dataset.key)) return '';
    return a.dataset.key;
  }

  function position(key) {
    var a = anchorOf(key), n = ensure();
    if (!a) { hide(); return; }
    var r = a.getBoundingClientRect(), vh = document.documentElement.clientHeight;
    if (isRich() && (mode === 'tap' || narrow())) { n.dataset.sheet = sheetEdge(r, vh); n.style.transform = ''; return; }
    delete n.dataset.sheet;
    var p = place(r, n.offsetWidth, n.offsetHeight, document.documentElement.clientWidth, vh);
    n.style.transform = 'translate(' + p.x + 'px, ' + p.y + 'px)';
  }
  function show(key, how) {
    var f = find(key);
    if (!f || !f.it) { hide(); return; }
    var t = build(f.it, f.slot);
    if (isRich()) t.rich = richOf(f.it, f);
    var sig = JSON.stringify(t);
    var n = ensure(), a = anchorOf(key);
    if (a) { if (a.removeAttribute) a.removeAttribute('title'); a.setAttribute('aria-describedby', TIP_ID); }
    if (openKey && openKey !== key) { var o = anchorOf(openKey); if (o) o.removeAttribute('aria-describedby'); }
    if (sig !== openSig || n.hidden) { render(t); openSig = sig; }
    openKey = key; mode = how;
    n.hidden = false;
    position(key);
  }
  function hide() {
    if (openKey) { var o = anchorOf(openKey); if (o && o.removeAttribute) o.removeAttribute('aria-describedby'); }
    openKey = ''; openSig = ''; mode = '';
    if (node) node.hidden = true;
  }

  // the page's poll: refill the open tip only when its text changed, or close it when the item left
  function onView(v) {
    view = v;
    if (!openKey) return;
    var f = find(openKey);
    if (!f || !f.it) { hide(); return; }
    show(openKey, mode);
  }

  function wire(root) {
    root.addEventListener('pointerover', function (e) {
      if (e.pointerType && e.pointerType !== 'mouse') return;
      var k = keyOf(e.target); if (k && mode !== 'tap') show(k, 'hover');
    });
    root.addEventListener('pointerout', function (e) {
      if (mode === 'hover' && keyOf(e.target) === openKey && !(e.relatedTarget && keyOf(e.relatedTarget) === openKey)) hide();
    });
    root.addEventListener('focusin', function (e) { var k = keyOf(e.target); if (k && kbd) show(k, 'focus'); });
    root.addEventListener('focusout', function (e) { if (mode === 'focus' && keyOf(e.target) === openKey) hide(); });
    root.addEventListener('click', function (e) {
      var k = keyOf(e.target);
      if (!k || (lastType !== 'touch' && lastType !== 'pen')) return;
      if (openKey === k && mode === 'tap') hide(); else show(k, 'tap');
    });
  }
  function init() {
    ['equip', 'pack', 'belt', 'stash'].forEach(function (id) { var r = document.getElementById(id); if (r) wire(r); });
    document.addEventListener('d2:inventory', function (e) { onView(e.detail); });
    document.addEventListener('pointerdown', function (e) {
      lastType = e.pointerType || 'mouse'; kbd = false;
      if (mode === 'tap' && !(node && node.contains && node.contains(e.target)) && keyOf(e.target) !== openKey) hide();
    }, true);
    document.addEventListener('keydown', function (e) {
      kbd = true;
      if (e.key === 'Escape' && openKey) hide();
    });
    window.addEventListener('scroll', function () { if (openKey) position(openKey); }, { passive: true });
    window.addEventListener('resize', function () { if (openKey) position(openKey); });
  }

  if (typeof window !== 'undefined' && window.__tipTest) {
    window.__tipTest({ build: build, richOf: richOf, reqRows: reqRows, compare: compare, parseLine: parseLine, sheetEdge: sheetEdge, place: place, nameClass: nameClass, onView: onView, show: show, hide: hide, init: init, state: function () { return { key: openKey, mode: mode, node: node }; } });
    return;
  }
  init();
})();
