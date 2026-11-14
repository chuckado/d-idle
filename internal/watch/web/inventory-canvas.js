// The character canvas of the inventory page (card uic2, docs/ui/redesign-2026-11-14/briefs/C2-character-canvas.md). It
// does what the cards page had no place for: the bar (search, four filters that dim and never hide, gold, repair, sell),
// the stat sheets, the set chips, the selected item with its actions as buttons, the requirement marks on the cells
// (from `req` and `usable` of the view only: no rule is computed here), the two tabs of a phone, and the trimming of the
// two lists that must never scroll. The doll, the cells, the merc doll, the errands and the queue are still drawn by
// inventory.js, inventory-merc.js and inventory-actions.js into the same element ids; this file reads the view they get
// (the d2:inventory event), draws after them (a microtask) and never writes to the bot except through window.d2InvActions.
// Text goes in through textContent only; the only inline style writes are CSSOM custom properties.
(function () {
  'use strict';
  var $ = function (id) { return document.getElementById(id); };
  var view = null, hero = null, repair = null, selKey = '', askTrash = '', heroAt = 0, drawnSheet = '', popOpen = false;
  var F = { q: '', usable: false, merc: false, unid: false, sell: false };
  var FILTERS = [['usable', 'Usable by me', 'Mine'], ['merc', 'By mercenary', 'Merc'], ['unid', 'Unidentified', 'Unid'], ['sell', 'Marked sell', 'Sell']];
  var SLOTWORD = { head: 'helm', neck: 'amulet', torso: 'armour', rhand: 'right hand', lhand: 'left hand', rring: 'right ring', lring: 'left ring', belt: 'belt', feet: 'boots', gloves: 'gloves' };
  var REASON = { class: 'CLS', identified: 'UNID', twohand: '2H', slot: 'SLOT', dead: 'DEAD', merctype: 'TYPE', broken: 'BRK' }; // the badge word of a reason code
  var STAT = { lvl: 'LVL', str: 'STR', dex: 'DEX' };
  var SHORT = { 'Deposit gold': 'Deposit', 'Withdraw gold': 'Withdraw', 'Buy identify scrolls': 'Buy scrolls', 'Identify all (Cain)': 'Identify all' };

  function el(t, c, x) { var e = document.createElement(t); if (c) e.className = c; if (x !== undefined) e.textContent = x; return e; }
  function clear(n) { while (n.firstChild) n.removeChild(n.firstChild); }
  function api() { return (typeof window !== 'undefined' && window.d2InvActions) || null; }
  function num(n) { return String(Math.round(Number(n) || 0)).replace(/\B(?=(\d{3})+(?!\d))/g, ','); }
  function setVar(n, k, v) { if (n.style && n.style.setProperty) n.style.setProperty(k, String(v)); else if (n.style) n.style[k] = String(v); }
  // a redraw is skipped when the signature of what it shows is unchanged; the hash keeps the attribute short
  function hash(t) { var h = 5381, i; t = String(t); for (i = 0; i < t.length; i++) h = ((h << 5) + h + t.charCodeAt(i)) | 0; return (h >>> 0).toString(36); }
  function say(t) { var a = api(); if (a && a.say) a.say(t); }

  // ---- the view: where an item is ----
  function locate(v, key) {
    var m = /^(item|slot)-(.+)$/.exec(key || '');
    if (!v || !m) return null;
    if (m[1] === 'slot') { var w = v.equipment && v.equipment[m[2]]; return w ? { it: w, slot: m[2], place: 'worn' } : null; }
    var id = Number(m[2]), hit = null;
    [['pack', v.backpack], ['stash', v.stash], ['belt', v.belt], ['cube', v.cube]].forEach(function (l) {
      (l[1] || []).forEach(function (x) { if (x.id === id) hit = { it: x, slot: '', place: l[0] }; });
    });
    return hit;
  }
  function allItems(v) {
    var o = [];
    [v.backpack, v.stash, v.belt].forEach(function (l) { (l || []).forEach(function (x) { o.push(x); }); });
    Object.keys(v.equipment || {}).forEach(function (k) { if (v.equipment[k]) o.push(v.equipment[k]); });
    return o;
  }
  function brokenCount(v) {
    var n = 0;
    Object.keys(v.equipment || {}).forEach(function (k) { var it = v.equipment[k]; if (it && it.maxDur && !(it.dur > 0)) n++; });
    return n;
  }
  function side(it, who) { return it && it.usable && it.usable[who] ? it.usable[who] : null; }
  function hardReasons(s) { return s && s.reasons ? s.reasons.filter(function (r) { return !r.warn; }) : []; }

  // ---- bar ----
  function drawBar(v) {
    $('cv-gold').textContent = num(v.gold);
    $('cv-stashgold').textContent = v.stashGold || v.stashGold === 0 ? num(v.stashGold) : '-';
    var n = v.sellPending || 0, b = $('cv-sell');
    b.textContent = 'Sell marked (' + n + ')';
    if (n > 0) b.removeAttribute('aria-disabled'); else b.setAttribute('aria-disabled', 'true');
    b.title = n > 0 ? 'Walk home and sell the items marked for sale' : 'Nothing is marked for sale';
  }
  function drawRepair(v, r) {
    var b = $('cv-repair');
    if (!r) { b.hidden = true; return; }
    b.hidden = false;
    var q = r.request, pend = q && q.state === 'requested', t = r.trip;
    var bt = t && t.button ? t.button : { label: pend ? 'Cancel repair' : 'Repair now', enabled: true, action: pend ? 'cancel' : 'request' };
    var est = t ? t.estimate : r.estimate;
    b.textContent = bt.action === 'cancel' ? 'Cancel repair' : 'Repair' + (est ? ' ' + num(est) + ' g' : '');
    var broken = brokenCount(v) > 0;
    if (broken) b.dataset.danger = ''; else delete b.dataset.danger;
    b.dataset.action = bt.action;
    if (!bt.enabled || bt.action === 'none') b.setAttribute('aria-disabled', 'true'); else b.removeAttribute('aria-disabled');
    var tip = [];
    if (r.worst) tip.push(r.worst);
    if (t && t.line) tip.push(t.line); else if (q && q.state) tip.push('Last request: ' + q.state);
    if (bt.reason) tip.push(bt.reason);
    b.title = tip.join('. ') || bt.label;
  }
  function onRepairClick() {
    var b = $('cv-repair'), a = api();
    if (!a || b.hasAttribute('aria-disabled')) { if (b.title) say(b.title); return; }
    a.post(b.dataset.action === 'cancel' ? '/api/repair/cancel' : '/api/repair/request', {});
  }
  function chips() {
    var host = $('cv-chips'); if (host.firstChild) return;
    FILTERS.forEach(function (f) {
      var b = el('button', 'cv-chip'); b.type = 'button'; b.dataset.filter = f[0]; b.setAttribute('aria-pressed', 'false');
      b.appendChild(el('span', 'f', f[1])); b.appendChild(el('span', 's', f[2])); b.title = f[1];
      b.addEventListener('click', function () { F[f[0]] = !F[f[0]]; b.setAttribute('aria-pressed', F[f[0]] ? 'true' : 'false'); dim(); });
      host.appendChild(b);
    });
  }
  // the four filters and the search dim what does not match, they never hide a cell (and a dimmed item still answers a click)
  function matches(it) {
    var u = side(it, 'hero'), m = side(it, 'merc');
    if (F.usable && !(u && u.ok)) return false;
    if (F.merc && !(m && m.ok)) return false;
    if (F.unid && it.identified !== false) return false;
    if (F.sell && it.mark !== 'sell') return false;
    if (F.q) {
      var hay = ((it.name || '') + ' ' + (it.code || '') + ' ' + (it.props || []).join(' ') + ' ' + (it.quality || '')).toLowerCase();
      if (hay.indexOf(F.q) < 0) return false;
    }
    return true;
  }
  function dim() {
    if (!view) return;
    var on = F.usable || F.merc || F.unid || F.sell || F.q;
    [['pack', view.backpack], ['stash', view.stash]].forEach(function (g) {
      var host = $(g[0]); if (!host) return;
      (g[1] || []).forEach(function (it) {
        var n = host.querySelector('[data-key="item-' + it.id + '"]'); if (!n) return;
        if (on && !matches(it)) n.dataset.dim = ''; else delete n.dataset.dim;
      });
    });
  }

  // ---- the requirement marks on the cells ----
  function badge(n, text, tone, title) {
    var old = n.querySelector && n.querySelector('.bd'); if (old) n.removeChild(old);
    var b = el('span', 'bd', text); b.dataset.tone = tone; if (title) b.title = title; n.appendChild(b);
  }
  function reasonWord(r) { return r.code === 'requirements' ? (STAT[r.stat] || 'REQ') : (REASON[r.code] || String(r.code || '?').toUpperCase().slice(0, 4)); }
  function marks(v) {
    [['pack', v.backpack], ['stash', v.stash], ['belt', v.belt]].forEach(function (g) {
      var host = $(g[0]); if (!host) return;
      (g[1] || []).forEach(function (it) {
        var n = host.querySelector('[data-key="item-' + it.id + '"]'); if (!n) return;
        var u = side(it, 'hero'), m = side(it, 'merc');
        delete n.dataset.usable; delete n.dataset.who;
        if (!u) return;
        if (u.ok) { if ((u.reasons || []).some(function (r) { return r.warn; })) badge(n, 'BRK', 'warn', 'Broken: repair it before it is worn'); return; }
        var hard = hardReasons(u), why = hard.length ? hard[0].text : 'you cannot use it';
        if (m && m.ok) { n.dataset.who = 'merc'; badge(n, 'MERC', 'info', 'Only your mercenary can use it. ' + why); n.title = (n.title || '') + ' - ' + why; return; }
        n.dataset.usable = 'no';
        badge(n, hard.length ? reasonWord(hard[0]) : 'NO', 'bad', why);
        n.title = (n.title || '') + ' - ' + why;
      });
    });
  }

  // ---- stat sheets ----
  function tile(label, value, tone, title) { return { l: label, v: value, t: tone || '', k: title || '' }; }
  function heroTiles(v, h) {
    var t = [];
    if (h && h.level) t.push(tile('Level', h.level));
    if (h && h.maxLife) t.push(tile('Life', num(h.life) + ' / ' + num(h.maxLife), h.life <= 0 ? 'bad' : ''));
    if (h && h.maxMana) t.push(tile('Mana', num(h.mana) + ' / ' + num(h.maxMana)));
    var r = h && h.stats && h.stats.resists;
    if (r) [['Fire res', 'fire'], ['Cold res', 'cold'], ['Light res', 'light'], ['Poison res', 'poison']].forEach(function (k) { t.push(tile(k[0], r[k[1]] + '%', r[k[1]] < 0 ? 'bad' : '')); });
    if (h && h.xpNext > 0) {
      var left = h.xpNext - h.xp;
      t.push(tile('xp to next', h.xpPct === null || h.xpPct === undefined ? '-' : Math.floor(h.xpPct) + '%'));
      t.push(tile('xp left', left > 0 ? num(left) : '0'));
    }
    var b = brokenCount(v);
    t.push(tile('Gear', b ? b + ' broken' : 'ok', b ? 'bad' : 'good'));
    return t;
  }
  function drawSheet(host, tiles, cols) {
    var sig = hash(JSON.stringify(tiles) + cols);
    if (host.dataset.sig === sig) return;
    host.dataset.sig = sig;
    clear(host);
    tiles.forEach(function (x) {
      var d = el('div'); if (x.k) d.title = x.k;
      d.appendChild(el('label', '', x.l));
      var b = el('b', '', String(x.v)); if (x.t) b.dataset.tone = x.t; d.appendChild(b);
      host.appendChild(d);
    });
    setVar(host, '--n', cols);
    host.hidden = !tiles.length;
  }
  function heroLabel(v, h) {
    var bits = [];
    if (v.class) bits.push(v.class + (h && h.level ? ' ' + h.level : ''));
    var m = $('equipmeta'); if (!m) return;
    var broken = brokenCount(v);
    if (broken) bits.push(broken + ' broken');
    m.textContent = bits.join(' · ');
  }

  // ---- gear set chips: the preview and the full panel (rename, delete, save) open in one popover ----
  function setWarn(v, s) {
    var pool = allItems(v);
    return (s.slots || []).some(function (x) {
      return pool.some(function (it) { var u = side(it, s.target === 'merc' ? 'merc' : 'hero'); return it.name === x.name && it.code === x.code && u && !u.ok; });
    });
  }
  function drawSets(v) {
    var bar = $('cv-sets'), s = v.sets;
    var sig = hash(JSON.stringify([s, v.merc && v.merc.owned, v.backpack, v.stash, v.equipment, popOpen]));
    if (bar.dataset.sig === sig) return;
    bar.dataset.sig = sig;
    clear(bar);
    bar.hidden = !s;
    if (!s) return;
    bar.appendChild(el('span', 'cv-lab', 'Sets'));
    if (!s.open) {
      var lk = el('button', 'cv-chip', 'Locked'); lk.type = 'button'; lk.title = s.why || 'Gear sets are locked'; lk.addEventListener('click', function () { pop(true); });
      bar.appendChild(lk); return;
    }
    (s.sets || []).forEach(function (x) {
      var c = el('button', 'cv-chip', x.name); c.type = 'button'; c.dataset.set = x.id;
      c.title = (x.target === 'merc' ? 'Mercenary: ' : '') + (x.slots || []).map(function (q) { return (SLOTWORD[q.slot] || q.slot) + ': ' + q.name; }).join(', ');
      if (x.applying) c.dataset.applying = '';
      if (setWarn(v, x)) { c.dataset.warn = ''; c.title += ' (an item here is under its requirement)'; }
      c.addEventListener('click', function () { pop(true); });
      bar.appendChild(c);
    });
    var add = el('button', 'cv-chip', '+'); add.type = 'button'; add.title = 'Save what you wear now as a set'; add.setAttribute('aria-label', 'Save a gear set');
    add.addEventListener('click', function () { pop(true); });
    bar.appendChild(add);
    fitSets();
  }
  // chips that do not fit the row give way to one "+n" chip that opens the panel (the bar never overflows, never scrolls)
  function fitSets() {
    var bar = $('cv-sets'); if (!bar || !bar.clientWidth || bar.scrollWidth <= bar.clientWidth + 1) return;
    var chips = Array.prototype.filter.call(bar.children, function (c) { return c.dataset && c.dataset.set; }), cut = 0, more = el('button', 'cv-chip', '');
    more.type = 'button'; more.addEventListener('click', function () { pop(true); });
    bar.insertBefore(more, bar.lastChild);
    while (chips.length && bar.scrollWidth > bar.clientWidth + 1) { bar.removeChild(chips.pop()); cut++; more.textContent = '+' + cut; more.title = cut + ' more gear sets'; }
  }
  function pop(open) {
    popOpen = open; var p = $('cv-pop'); p.hidden = !open;
    if (open) { var f = p.querySelector('button, input'); if (f && f.focus) f.focus({ preventScroll: true }); }
  }

  // ---- the selected item: its facts and its actions as buttons (the list of the action menu) ----
  function actions(f) {
    var it = f.it, a = api(), out = [], key = (f.place === 'worn' ? 'slot-' + f.slot : 'item-' + it.id);
    function add(label, fn, reason, o) { out.push({ label: label, fn: fn, reason: reason || '', primary: o && o.primary, confirm: o && o.confirm, act: o && o.act }); }
    if (it.identified === false) add('Identify', function () { a.identify(it.id); }, it.noIdentify, { primary: true, act: 'identify' });
    if (f.place === 'pack') {
      var can = it.canWear || [];
      if (!can.length) add('Equip', null, it.noWear || 'it cannot be worn right now', { act: 'equip' });
      else can.forEach(function (s) { add(can.length > 1 ? 'Equip (' + (SLOTWORD[s] || s) + ')' : 'Equip', function () { a.equip(it.id, s); }, '', { primary: it.identified !== false, act: 'equip-' + s }); });
    } else if (f.place === 'worn') add('Take off', function () { a.unequip(f.slot); }, it.noUnequip, { act: 'unequip' });
    if (it.stashOp === 'stash' || it.stashOp === 'take') add(it.stashOp === 'stash' ? 'To stash' : 'Take out', function () { (it.stashOp === 'stash' ? a.stash : a.take)(it.id); }, it.noStash, { act: it.stashOp });
    var sold = it.mark === 'sell';
    add(sold ? 'Clear sell mark' : 'Mark for sale', function () { a.post('/api/inventory/mark', { id: it.id, mark: sold ? '' : 'sell' }); }, sold ? '' : (it.noSell || (f.place !== 'pack' ? 'only backpack items are sold' : '')), { act: 'sell' });
    add(it.mark === 'keep' ? 'Clear keep mark' : 'Keep', function () { a.post('/api/inventory/mark', { id: it.id, mark: it.mark === 'keep' ? '' : 'keep' }); }, '', { act: 'keep' });
    var ask = askTrash === key;
    add(ask ? 'Confirm: trash' : 'Trash', function () { if (!ask) { askTrash = key; drawSel(); } else { askTrash = ''; a.trash(it.id); } }, it.noTrash, { confirm: ask, act: 'trash' });
    return out;
  }
  function whereWord(f) {
    return f.place === 'worn' ? 'Worn: ' + (SLOTWORD[f.slot] || f.slot) : f.place === 'pack' ? 'In the pack' : f.place === 'stash' ? 'In the stash' : f.place === 'belt' ? 'On the belt' : 'In the cube';
  }
  function drawSel() {
    var box = $('cv-sel'); if (!box) return;
    var f = view && selKey ? locate(view, selKey) : null;
    var sig = hash(JSON.stringify([f, askTrash === selKey, view && view.merc && view.merc.state]));
    if (box.dataset.sig === sig) { fitActs(); return; }
    box.dataset.sig = sig;
    clear(box);
    if (!f || !f.it) {
      box.appendChild(el('p', 'cv-hint', 'Pick an item or a slot to see it here.'));
      box.dataset.empty = ''; return;
    }
    delete box.dataset.empty;
    var it = f.it, h = el('h3', 'q-' + (it.quality || 'normal'), (it.name || it.code || 'item') + (it.identified === false ? ' (unidentified)' : '')); h.title = h.textContent; box.appendChild(h);
    var bits = [whereWord(f)]; if (it.level) bits.push('item level ' + it.level); if (it.maxDur) bits.push('durability ' + (it.dur || 0) + '/' + it.maxDur);
    box.appendChild(el('p', '', bits.join(' · ')));
    var why = el('div', 'cv-why'), u = side(it, 'hero'), m = side(it, 'merc');
    if (u) {
      var hr = hardReasons(u);
      if (hr.length) hr.forEach(function (r) { var p = el('p', '', 'You: ' + r.text); p.dataset.tone = 'bad'; p.title = r.text; why.appendChild(p); });
      else { var ok = el('p', '', 'You can use it'); ok.dataset.tone = 'good'; why.appendChild(ok); }
    }
    if (m) {
      var mr = hardReasons(m), p2 = el('p', '', m.ok ? 'He can use it' : 'He: ' + (mr.length ? mr[0].text : 'no')); p2.dataset.tone = m.ok ? 'good' : 'bad'; why.appendChild(p2);
    }
    if (it.why) why.appendChild(el('p', '', 'Why: ' + it.why));
    if ((it.props || []).length) why.appendChild(el('p', '', it.props.slice(0, 2).join(', ')));
    box.appendChild(why);
    var acts = el('div', 'cv-acts'), a = api();
    if (a) actions(f).forEach(function (e) {
      var b = el('button', '', e.label); b.type = 'button'; b.dataset.act = e.act || ''; if (e.primary) b.dataset.primary = ''; if (e.confirm) b.dataset.confirm = '';
      if (e.reason) { b.setAttribute('aria-disabled', 'true'); b.title = e.reason; }
      b.addEventListener('click', function () { if (e.reason) { say(e.reason); return; } if (e.fn) e.fn(); });
      acts.appendChild(b);
    });
    box.appendChild(acts);
    fitActs();
  }
  // a list that does not fit its box loses its tail, never scrolls; the action buttons end with More, which opens the full menu
  function fitActs() {
    var acts = document.querySelector && document.querySelector('#cv-sel .cv-acts');
    if (!acts || !acts.clientHeight || acts.dataset.fit) return;
    var cut = 0;
    while (acts.children.length > 1 && acts.scrollHeight > acts.clientHeight + 1) { acts.removeChild(acts.lastChild); cut++; }
    if (cut) {
      acts.removeChild(acts.lastChild);
      var more = el('button', '', 'More...'); more.type = 'button'; more.dataset.act = 'more'; more.title = 'Every action for this item';
      more.addEventListener('click', function () { var a = api(); if (a && a.open && selKey) a.open(selKey); });
      acts.appendChild(more);
    }
    acts.dataset.fit = '';
  }
  var trimming = false;
  function trimQueue() {
    var q = $('queue'); if (!q || trimming || !q.clientHeight) return;
    var ul = q.querySelector('ul'), log = $('cv-log');
    if (log) { if (ul) delete log.dataset.empty; else log.dataset.empty = ''; }
    if (!ul) return;
    trimming = true;
    while (ul.children.length > 1 && q.scrollHeight > q.clientHeight + 1) ul.removeChild(ul.lastChild);
    trimming = false;
  }
  // slots that would take the selected pack or stash item are outlined green, the others red with the server's reason
  function fitSlots() {
    var f = view && selKey ? locate(view, selKey) : null, it = f && f.it, host = $('equip');
    if (!host) return;
    // the reason most slots share ("does not fit that slot") is no news: a slot is marked red, with its reason, only when its
    // reason differs from the common one (the slot the item belongs to but may not go to now)
    var common = '', n = {}, why = (it && it.noWearSlot) || {};
    Object.keys(why).forEach(function (k) { n[why[k]] = (n[why[k]] || 0) + 1; if (n[why[k]] > (n[common] || 0)) common = why[k]; });
    Array.prototype.forEach.call(host.querySelectorAll('.d-slot2'), function (s) {
      delete s.dataset.fit; delete s.dataset.why;
      if (!it || f.place === 'worn' || !(it.canWear || it.noWearSlot)) return;
      var k = s.dataset.slot;
      if ((it.canWear || []).indexOf(k) >= 0) s.dataset.fit = 'ok';
      else if (why[k] && why[k] !== common) { s.dataset.fit = 'no'; s.dataset.why = why[k]; }
    });
  }
  // the errand buttons come from inventory-actions.js with long words; a four-cell column needs short ones (the long word stays as the title)
  function shorten() {
    Array.prototype.forEach.call(document.querySelectorAll('.cv-side .act'), function (b) {
      var t = b.textContent;
      if (SHORT[t]) { b.title = b.title || t; b.textContent = SHORT[t]; }
    });
  }

  // ---- hero values: /api/hero, read at most every eight seconds with the page's own poll ----
  function loadHero() {
    var now = Date.now(); if (now - heroAt < 8000) return; heroAt = now;
    fetch('/api/hero', { cache: 'no-store' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (h) { if (h) { hero = h; paintSheets(); } }, function () {});
  }
  function paintSheets() {
    if (!view) return;
    var t = heroTiles(view, hero);
    drawSheet($('hsheet'), t, Math.max(1, Math.ceil(t.length / 2)));
    heroLabel(view, hero);
  }

  // ---- phone tabs ----
  function tabs() {
    var host = $('cv-tabbar'); if (!host || host.firstChild) return;
    [['hero', 'Hero and mercenary'], ['pack', 'Pack and stash']].forEach(function (x) {
      var b = el('button', '', x[1]); b.type = 'button'; b.dataset.tab = x[0]; b.setAttribute('aria-pressed', x[0] === $('cv').dataset.tab ? 'true' : 'false');
      b.addEventListener('click', function () {
        $('cv').dataset.tab = x[0];
        Array.prototype.forEach.call(host.children, function (c) { c.setAttribute('aria-pressed', c === b ? 'true' : 'false'); });
        relayout();
      });
      host.appendChild(b);
    });
  }

  function relayout() {
    var go = function () { delete_fit(); var b = $('cv-sets'); if (b) { b.dataset.sig = ''; if (view) drawSets(view); } trimQueue(); fitActs(); };
    if (typeof requestAnimationFrame === 'function') requestAnimationFrame(go); else go();
  }
  function delete_fit() { var a = document.querySelector('#cv-sel .cv-acts'); if (a) delete a.dataset.fit; }

  // ---- events ----
  // the page never opens on an empty box: the first item of the pack (else the stash) is selected until the player picks one
  function ensureSel(v) {
    if (selKey && locate(v, selKey)) return;
    var first = (v.backpack || [])[0] || (v.stash || [])[0];
    selKey = first ? 'item-' + first.id : ''; askTrash = '';
  }
  function markSel() {
    Array.prototype.forEach.call(document.querySelectorAll('.d-slot2.sel, .item.sel'), function (n) { n.classList.remove('sel'); });
    var n = selKey && document.querySelector('[data-key="' + selKey + '"]'); if (n) n.classList.add('sel');
  }
  function paint(v) {
    ensureSel(v); markSel();
    drawBar(v); drawRepair(v, repair); marks(v); dim(); drawSets(v); shorten(); fitSlots(); paintSheets();
    var g = $('gearprompt'); if (g) g.title = g.textContent;
    drawSel(); trimQueue(); loadHero();
  }
  function onView(e) {
    view = e.detail;
    var p = $('gearprompt'); if (p) p.title = p.textContent;
    Promise.resolve().then(function () { if (view === e.detail) paint(view); });
  }
  function init() {
    chips(); tabs();
    $('cv-q').addEventListener('input', function () { F.q = $('cv-q').value.trim().toLowerCase(); dim(); });
    $('cv-sell').addEventListener('click', function () {
      var b = $('cv-sell'), a = api();
      if (b.hasAttribute('aria-disabled')) { say('Nothing is marked for sale.'); return; }
      if (a) a.post('/api/inventory/sellnow', {});
    });
    $('cv-repair').addEventListener('click', onRepairClick);
    $('cv-popx').addEventListener('click', function () { pop(false); });
    document.addEventListener('d2:inventory', onView);
    document.addEventListener('d2:inventory-fresh', function (e) { view = e.detail; drawBar(view); drawSel(); });
    document.addEventListener('d2:repair', function (e) { repair = e.detail; if (view) drawRepair(view, repair); });
    document.addEventListener('click', function (e) {
      var k = e.target && e.target.closest ? e.target.closest('[data-key]') : null;
      if (!k || !/^(item|slot)-/.test(k.dataset.key || '')) return;
      selKey = k.dataset.key; askTrash = ''; if (view) { markSel(); fitSlots(); drawSel(); }
    });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && popOpen) { pop(false); return; }
      var t = e.target, typing = t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA');
      if (e.key === '/' && !typing && !e.ctrlKey && !e.metaKey && !e.altKey) { e.preventDefault(); $('cv-q').focus({ preventScroll: true }); }
      if (e.key === 'Escape' && t === $('cv-q')) { $('cv-q').value = ''; F.q = ''; dim(); $('cv-q').blur(); }
    });
    window.addEventListener('resize', relayout, { passive: true });
    if (typeof ResizeObserver === 'function') {
      var ro = new ResizeObserver(relayout);
      ['cv-log', 'cv-sel'].forEach(function (id) { var n = $(id); if (n) ro.observe(n); });
    }
    if (typeof MutationObserver === 'function' && $('queue')) new MutationObserver(function () { if (!trimming) { trimQueue(); } }).observe($('queue'), { childList: true });
    if (typeof MutationObserver === 'function' && $('errands')) new MutationObserver(function () { shorten(); }).observe($('errands').parentNode, { childList: true, subtree: true });
  }

  if (typeof window !== 'undefined' && window.__canvasTest) {
    window.__canvasTest({ heroTiles: heroTiles, locate: locate, marks: marks, matches: matches, actions: actions, setWarn: setWarn, reasonWord: reasonWord,
      setView: function (v) { view = v; }, setHero: function (h) { hero = h; }, setRepair: function (r) { repair = r; }, paint: paint, setFilter: function (k, x) { F[k] = x; }, drawBar: drawBar, drawSel: drawSel, drawSets: drawSets, select: function (k) { selKey = k; askTrash = ''; } });
    return;
  }
  init();
})();
