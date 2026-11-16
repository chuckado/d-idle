// Item action menu of the inventory page (docs/inventory-actions.md, docs/ui-design-system.md components 13 and 14):
// click, tap or press Enter on an item (backpack, stash, belt, a worn slot) and a small menu offers Equip, Mark for
// sale, Keep, Trash, and Identify. A choice QUEUES an action the bot performs at its next open opportunity; this file never
// touches the game. A phone gets the same menu as a bottom sheet. The queue card is redrawn here with the states
// queued, doing, done, failed and cancelled, a reason line, the result of a finished errand and a Cancel button for a
// queued request. Above the queue sit the two town errands (docs/inventory-actions.md, schema 5): "Buy identify scrolls"
// with a count box (empty = fill the tome) and "Identify all (Cain)", a locked tile with its teaser until Cain is rescued.
// Input: the view /api/inventory sends, handed over by inventory.js as a 'd2:inventory' event on document (the same
// feed the item tooltip uses). Text goes in through textContent only; nothing scrolls the page (focus is always taken
// with preventScroll) and the menu is position: fixed.
// EXTENSION POINT, used since the drag and drop pass (web/inventory-drag.js, docs/gearsets.md): every doll slot carries
// data-slot (and data-key="slot-<slot>" on the hero's); a drop calls the public API at the end of this file: equip(id,
// slot, target), unequip(slot, target), and trashConfirm(key) for a trash (the same confirm step as the menu).
(function () {
  'use strict';
  var MENU_ID = 'inv-act';
  var TRIP = 'waiting for a trip home'; // the stash is reached only in town: the bot walks home at the next safe moment
  var LABEL = { head: 'helm', neck: 'amulet', torso: 'armour', rhand: 'right hand', lhand: 'left hand', rring: 'right ring', lring: 'left ring', belt: 'belt', feet: 'boots', gloves: 'gloves' };
  var lastPtr = 'mouse', view = null, menu = null, scrim = null, openKey = '', openSig = '', confirming = false, kbd = false, sheet = false, openAt = null;

  function el(t, c, x) { var e = document.createElement(t); if (c) e.className = c; if (x !== undefined) e.textContent = x; return e; }
  function clear(n) { while (n.firstChild) n.removeChild(n.firstChild); }
  function say(t) { var m = document.getElementById('msg'); if (!m) return; m.textContent = t || ''; m.hidden = !t; }
  function slotLabel(s) { return LABEL[s] || s; }
  // the character canvas (body[data-canvas], card uic3): a click only selects, the menu is the right-click menu with a key on every
  // entry, and the gestures (double-click, Enter, Ctrl-click, Alt-click, Delete) run actions. The cards page keeps its click menu.
  function canvasMode() { return !!(document.body && document.body.dataset && document.body.dataset.canvas !== undefined); }
  var KEYHINT = { identify: 'Enter', equip: 'Enter', unequip: 'Enter', take: 'Ctrl+click', stash: 'Ctrl+click', sell: 'Alt+click', trash: 'Del', merc: 'M', repair: 'R' };

  // ---- data ----
  function find(key) {
    if (!view || !key) return null;
    var m = /^(item|slot)-(.+)$/.exec(key);
    if (!m) return null;
    if (m[1] === 'slot') return { it: view.equipment && view.equipment[m[2]], slot: m[2], place: 'worn' };
    var id = Number(m[2]), hit = null;
    [['pack', view.backpack], ['stash', view.stash], ['belt', view.belt], ['cube', view.cube]].forEach(function (l) {
      (l[1] || []).forEach(function (x) { if (x.id === id) hit = { it: x, slot: '', place: l[0] }; });
    });
    return hit;
  }
  function requestsFor(id) { return (view && view.queue || []).filter(function (r) { return r.id === id; }); }
  function kindOf(r) { return r.kind || 'equip'; }
  function verb(r) {
    var pre = r.groupName ? 'Set "' + r.groupName + '": ' : '';
    switch (kindOf(r)) {
      case 'unequip': return pre + 'Take off ' + r.name + ' (' + slotLabel(r.slot) + ')';
      case 'trash': return 'Trash ' + r.name;
      case 'identify': return r.place === 'stash' ? 'Identify ' + r.name + ' in the stash' : 'Identify ' + r.name;
      case 'stash': return 'Move ' + r.name + ' to the stash';
      case 'take': return pre + 'Take ' + r.name + ' out of the stash';
      case 'buyid': case 'cain': case 'gold': return r.name; // the server words the errand
    }
    return pre + 'Equip ' + r.name + ' in ' + slotLabel(r.slot);
  }

  // model builds the menu of one item: {title, cls, rows: [{text, state, cancel}], entries: [{act, label, reason, slot}]}.
  // An entry with a reason is shown but not available. The server checks everything again.
  function model(f) {
    var it = f.it, v = view;
    var reqs = requestsFor(it.id), trashQ = reqs.some(function (r) { return kindOf(r) === 'trash'; });
    var m = { target: 'hero', /* the menu structure is per target: the mercenary stage adds its own entries */ title: it.name || it.code || 'item', cls: 'q-' + (it.quality || 'normal'), id: it.id, key: f.place === 'worn' ? 'slot-' + f.slot : 'item-' + it.id, rows: [], entries: [] };
    reqs.forEach(function (r) {
      var row = { text: verb(r), state: r.state, kind: kindOf(r), cancel: r.state === 'queued' };
      if (r.state === 'queued' && (r.place === 'stash' || kindOf(r) === 'stash' || kindOf(r) === 'take') && r.reason) row.text += ' (' + TRIP + ')';
      m.rows.push(row);
    });
    // identify: offered on an unidentified item only; the server refuses an identified one
    if (it.identified === false) {
      var idQ = reqs.some(function (r) { return kindOf(r) === 'identify'; });
      m.entries.push({ act: 'identify', label: 'Identify', reason: it.noIdentify || (idQ ? 'it is already queued to be identified' : '') || (trashQ ? 'it is queued to be trashed: cancel that first' : '') });
    }
    var cv = canvasMode();
    // equip
    if (f.place === 'pack') {
      var can = it.canWear || [];
      if (!can.length) m.entries.push({ act: 'equip', label: 'Equip', reason: it.noWear || 'it cannot be worn right now' });
      else can.forEach(function (s) {
        m.entries.push({ act: 'equip', slot: s, label: can.length > 1 ? 'Equip (' + slotLabel(s) + ')' : 'Equip', reason: trashQ ? 'it is queued to be trashed: cancel that first' : '' });
      });
    } else {
      if (!cv) m.entries.push({ act: 'equip', label: 'Equip', reason: f.place === 'worn' ? 'it is already worn' : (it.noWear || 'move it to the backpack first') });
      if (f.place === 'worn') { // schema 7: take it off into the backpack (also what dragging it onto the pack does)
        var offQ = reqs.some(function (r) { return kindOf(r) === 'unequip'; });
        m.entries.push({ act: 'unequip', slot: f.slot, label: 'Take off', reason: it.noUnequip || (offQ ? 'it is already queued to come off' : '') });
      }
    }
    if (cv && f.place === 'pack' && it.usable && it.usable.merc) { // the mercenary's side: his verdict first, then whether the bot can dress him at all
      var mu = it.usable.merc, mh = (mu.reasons || []).filter(function (r) { return !r.warn; })[0];
      m.entries.push({ act: 'merc', label: 'Give to mercenary', reason: !mu.ok ? (mh ? mh.text : 'he cannot use it') : (v.merc && v.merc.actWhy) || '' });
    }
    if (cv && f.place === 'worn' && it.maxDur && !(it.dur > 0)) m.entries.push({ act: 'repair', label: 'Repair now', reason: '' });
    // stash moves (schema 6): "Move to stash" for a backpack item, "Take out" for a stash item. A stash job of the item is
    // queued once; an identify of a stash item runs in place, so it needs no take-out.
    if (it.stashOp === 'stash' || it.stashOp === 'take') {
      var busy = reqs.some(function (r) { return kindOf(r) === 'stash' || kindOf(r) === 'take' || (kindOf(r) === 'identify' && r.place === 'stash'); });
      m.entries.push({ act: it.stashOp, label: it.stashOp === 'stash' ? 'Move to stash' : 'Take out', reason: it.noStash || (busy ? 'a stash action for it is already queued' : '') || (trashQ ? 'it is queued to be trashed: cancel that first' : '') });
    }
    // sell and keep marks
    var sold = it.mark === 'sell';
    m.entries.push({ act: 'sell', label: sold ? 'Clear sell mark' : 'Mark for sale', reason: sold ? '' : (it.noSell || (f.place !== 'pack' ? 'only backpack items are sold' : '')) });
    m.entries.push({ act: 'keep', label: it.mark === 'keep' ? 'Clear keep mark' : 'Keep', reason: '' });
    // trash, with its confirm step
    var tr = { act: 'trash', label: confirming ? 'Confirm: trash it' : 'Trash', reason: it.noTrash || (trashQ ? 'it is already queued to be trashed' : '') };
    if (confirming && !tr.reason) tr.confirm = true;
    m.entries.push(tr);
    if (cv) m.entries.forEach(function (e) { if (KEYHINT[e.act]) e.key = KEYHINT[e.act]; });
    return m;
  }

  // ---- network (same gate as the other writes: the answer listener, cookie or token) ----
  function base() {
    return fetch('/api/config', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (c) {
      return c.answerBase ? c.answerBase : location.protocol + '//' + location.hostname + ':' + c.answerPort;
    });
  }
  // ---- what the page shows while and after a request (card uic3) ----
  // fb[id] = {state: 'sending' | 'done' | 'refused', slot, code, text, seen}. 'sending' is the dashed outline of a request
  // the server has not answered (the instant lane waits up to four seconds); 'done' and 'refused' stay for two views. The item
  // is never moved on the page before the view says so: a refusal leaves everything where it was.
  var fb = {};
  // What the cell says while the request runs (within one frame of the click: setFb runs before the POST is sent).
  var DOING = { identify: 'Identifying...', equip: 'Equipping...', unequip: 'Taking off...', stash: 'Stashing...', take: 'Taking...', trash: 'Dropping...', gold: 'Moving gold...', cain: 'Identifying...' };
  var DONE_MS = 2000, FAST_MS = 200, timer = 0;
  function targetOf(path, body) {
    body = body || {};
    return { id: body.id === undefined ? 0 : body.id, slot: body.slot || '', kind: String(path).replace(/^.*\//, '') };
  }
  function setFb(t, state, code, text) {
    if (!t.id && !t.slot) return;
    var k = t.id || t.slot;
    fb[k] = { state: state, id: t.id, slot: t.slot, kind: t.kind, code: code || '', text: text || '', seen: 0 };
    markFeedback();
    // the success highlight lasts two seconds by the clock, not by the number of views (the poll is fast now)
    if (state === 'done' && typeof setTimeout === 'function') {
      var mine = fb[k];
      setTimeout(function () { if (fb[k] === mine) { delete fb[k]; markFeedback(); } }, DONE_MS);
    }
  }
  function refuse(id, text, code) { setFb({ id: id, slot: '', kind: '' }, 'refused', code, text); if (text) say(text); }
  // No scroll or tome: refuse at once and point at the way out, never wait (the button gets a ring for a few seconds).
  function offerBuy() {
    if (!ensureErrands() || !errBuy) return;
    errBuy.dataset.offer = '';
    if (typeof setTimeout === 'function') setTimeout(function () { delete errBuy.dataset.offer; }, 8000);
    errNote('No identify scroll or tome: use "Buy identify scrolls".');
  }
  function noToolText(t) { return /scroll|tome/i.test(t || ''); }
  function markFeedback() {
    Array.prototype.forEach.call(document.querySelectorAll ? document.querySelectorAll('[data-result], [data-sending]') : [], function (n) { delete n.dataset.result; delete n.dataset.sending; delete n.dataset.doing; n.removeAttribute('data-why'); });
    Object.keys(fb).forEach(function (k) {
      var x = fb[k];
      [x.id ? 'item-' + x.id : '', x.slot ? 'slot-' + x.slot : ''].forEach(function (key) {
        var a = key && anchorOf(key); if (!a || !a.dataset) return;
        if (x.state === 'sending') { a.dataset.sending = x.kind || ''; a.dataset.doing = DOING[x.kind] || 'Working...'; }
        else { a.dataset.result = x.state; if (x.text) a.setAttribute('data-why', x.text); }
      });
    });
  }
  function ageFeedback() {
    Object.keys(fb).forEach(function (k) { if (fb[k].state !== 'sending' && ++fb[k].seen > 2) delete fb[k]; });
  }
  function post(path, body) {
    var t = targetOf(path, body);
    setFb(t, 'sending');
    return base().then(function (u) {
      return fetch(u + path, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    }).then(function (r) { return r.json().then(function (j) { if (!r.ok) { var e = new Error(j.error || ('error ' + r.status)); e.code = j.code || ''; throw e; } return j; }); })
      .then(function (j) {
        say(''); errNote('');
        var q = j && j.request, st = q && q.state;
        if (st === 'done') setFb(t, 'done', '', q.result || ''); else if (st === 'failed') setFb(t, 'refused', q.code || '', q.reason || ''); else delete fb[t.id || t.slot];
        return refresh();
      }, function (e) { say(e.message); setNote(e.message); errNote(e.message); setFb(t, 'refused', e.code, e.message); if (t.kind === 'identify' && noToolText(e.message)) offerBuy(); return refresh(); });
  }
  function refresh() {
    return fetch('/api/inventory', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (v) {
      onView(v, true);
      // the poll of inventory.js is every 4 s; our own change shows at once in the modules that listen (drag, sets) and
      // inventory.js repaints the grids on this event, so the cell the answer is about is the cell on the page
      if (typeof CustomEvent === 'function' && document.dispatchEvent) document.dispatchEvent(new CustomEvent('d2:inventory-fresh', { detail: v }));
    }, function () {});
  }
  // The public calls, also the hook of the later drag and drop pass.
  function equip(id, slot, target) { var b = { id: id, slot: slot || '' }; if (target && target !== 'hero') b.target = target; return post('/api/inventory/equip', b); }
  function unequip(slot, target) { var b = { slot: slot }; if (target && target !== 'hero') b.target = target; return post('/api/inventory/unequip', b); }
  function trash(id) { return post('/api/inventory/trash', { id: id, confirm: true }); }
  function cancel(id, kind) { return post('/api/inventory/cancel', { id: id, kind: kind || 'equip' }); }
  function identify(id) { return post('/api/inventory/identify', { id: id }); }
  function buyIdentify(n) { return post('/api/inventory/buyid', { count: n || 0 }); }
  function cainIdentify() { return post('/api/inventory/cain', {}); }
  function stash(id) { return post('/api/inventory/stash', { id: id }); }
  function take(id) { return post('/api/inventory/take', { id: id }); }
  function gold(op, n) { return post('/api/inventory/gold', { op: op, count: n }); }

  function run(e, m) {
    if (e.reason) { setNote(e.reason); refuse(m.id, e.reason); if (e.act === 'identify' && noToolText(e.reason)) offerBuy(); return; }
    if (e.act !== 'trash') confirming = false;
    switch (e.act) {
      case 'equip': closeMenu(true); equip(m.id, e.slot); return;
      case 'unequip': closeMenu(true); unequip(e.slot); return;
      case 'sell': { var it = find(m.key).it; closeMenu(true); post('/api/inventory/mark', { id: m.id, mark: it.mark === 'sell' ? '' : 'sell' }); return; }
      case 'keep': { var k = find(m.key).it; closeMenu(true); post('/api/inventory/mark', { id: m.id, mark: k.mark === 'keep' ? '' : 'keep' }); return; }
      case 'identify': closeMenu(true); identify(m.id); return;
      case 'stash': closeMenu(true); stash(m.id); return;
      case 'take': closeMenu(true); take(m.id); return;
      case 'merc': closeMenu(true); equip(m.id, '', 'merc'); return;
      case 'repair': closeMenu(true); post('/api/repair/request', {}); return;
      case 'trash':
        if (!e.confirm) { confirming = true; draw(true); return; } // the first click only asks
        closeMenu(true); trash(m.id); return;
    }
  }

  // ---- menu DOM ----
  function setNote(t) { if (menu) { var n = menu.querySelector('.act-note'); if (n) n.textContent = t || ''; } }
  function ensure() {
    if (menu) return menu;
    menu = el('div', 'inv-act'); menu.id = MENU_ID; menu.setAttribute('role', 'menu'); menu.hidden = true;
    scrim = el('div', 'inv-act-scrim'); scrim.hidden = true;
    scrim.addEventListener('click', function () { closeMenu(true); });
    document.body.appendChild(scrim); document.body.appendChild(menu);
    return menu;
  }
  function render(m) {
    var n = ensure(); clear(n);
    n.setAttribute('aria-label', 'Actions for ' + m.title);
    n.appendChild(el('div', 'act-title ' + m.cls, m.title));
    m.rows.forEach(function (r) {
      var row = el('div', 'act-state'); row.dataset.state = r.state; row.dataset.kind = r.kind;
      row.appendChild(el('span', 'act-st', r.state)); row.appendChild(el('span', 'act-tx', r.text));
      if (r.cancel) {
        var c = el('button', 'act act-cancel', 'Cancel'); c.type = 'button'; c.dataset.act = 'cancel'; c.setAttribute('role', 'menuitem');
        c.addEventListener('click', function () { closeMenu(true); cancel(m.id, r.kind); });
        row.appendChild(c);
      }
      n.appendChild(row);
    });
    m.entries.forEach(function (e) {
      var b = el('button', 'act', e.label); b.type = 'button'; b.dataset.act = e.act + (e.slot ? '-' + e.slot : '');
      b.setAttribute('role', 'menuitem');
      if (e.key) b.appendChild(el('kbd', '', e.key));
      if (e.act === 'trash') b.dataset.danger = '';
      if (e.confirm) b.dataset.confirm = '';
      if (e.reason) { b.setAttribute('aria-disabled', 'true'); b.dataset.off = ''; }
      b.addEventListener('click', function () { run(e, m); });
      n.appendChild(b);
      if (e.reason) { var w = el('div', 'act-why', e.reason); w.dataset.for = b.dataset.act; n.appendChild(w); }
    });
    var note = el('div', 'act-note'); note.setAttribute('role', 'status'); n.appendChild(note);
    if (confirming) { var no = el('button', 'act', 'Keep it'); no.type = 'button'; no.dataset.act = 'back'; no.setAttribute('role', 'menuitem'); no.addEventListener('click', function () { confirming = false; draw(true); }); n.appendChild(no); }
  }
  function place(a, w, h, vw, vh) { // beside the anchor, else below or above; always inside the viewport
    var mg = 8, x = a.right + mg + w <= vw - mg ? a.right + mg : a.left - mg - w >= mg ? a.left - mg - w : Math.round((a.left + a.right) / 2 - w / 2), y = a.top;
    if (x !== a.right + mg && x !== a.left - mg - w) y = a.bottom + mg + h <= vh - mg ? a.bottom + mg : a.top - mg - h;
    return { x: Math.max(mg, Math.min(Math.round(x), vw - w - mg)), y: Math.max(mg, Math.min(Math.round(y), vh - h - mg)) };
  }
  function anchorOf(key) { return document.querySelector('[data-key="' + key + '"]'); }
  function position() {
    if (sheet) { menu.style.transform = ''; return; }
    var a = openAt ? null : anchorOf(openKey); if (!a && !openAt) { closeMenu(false); return; }
    var p = place(openAt ? { left: openAt.x, right: openAt.x, top: openAt.y, bottom: openAt.y } : a.getBoundingClientRect(), menu.offsetWidth, menu.offsetHeight, document.documentElement.clientWidth, document.documentElement.clientHeight);
    menu.style.transform = 'translate(' + p.x + 'px, ' + p.y + 'px)';
  }
  function focusables() { return menu ? menu.querySelectorAll('button.act') : []; }
  function draw(keepFocus) {
    var f = find(openKey);
    if (!f || !f.it) { closeMenu(false); return; }
    var m = model(f), sig = JSON.stringify(m) + (confirming ? 'c' : '');
    if (sig === openSig && !menu.hidden) return;
    var act = keepFocus && document.activeElement && document.activeElement.dataset ? document.activeElement.dataset.act : '';
    render(m); openSig = sig;
    menu.dataset.mode = sheet ? 'sheet' : 'pop'; menu.hidden = false; scrim.hidden = !sheet;
    var list = focusables(), target = null;
    // on the confirm step the focus is on "Keep it": a second Enter or Space must never be the one that trashes (review c2c3 R12)
    if (confirming) for (var c = 0; c < list.length; c++) if (list[c].dataset.act === 'back') target = list[c];
    for (var i = 0; !target && i < list.length; i++) if (act && list[i].dataset.act === act) target = list[i];
    if (keepFocus && !target) for (var j = 0; j < list.length; j++) if (!list[j].hasAttribute('aria-disabled')) { target = list[j]; break; }
    if (target) target.focus({ preventScroll: true });
    position();
  }
  function openMenu(key, at) {
    var f = find(key);
    if (!f || !f.it) return;
    ensure();
    confirming = false; openKey = key; openSig = ''; openAt = at || null;
    if (window.matchMedia) sheet = window.matchMedia('(pointer: coarse)').matches || window.matchMedia('(max-width: 40rem)').matches;
    menu.hidden = true;
    var a = anchorOf(key); if (a) a.setAttribute('aria-haspopup', 'menu');
    draw(false);
    var list = focusables(); // the first available entry takes focus so the keyboard works at once
    for (var i = 0; i < list.length; i++) if (!list[i].hasAttribute('aria-disabled')) { list[i].focus({ preventScroll: true }); break; }
  }
  function closeMenu(refocus) {
    var key = openKey;
    openKey = ''; openSig = ''; confirming = false; openAt = null;
    if (menu) menu.hidden = true;
    if (scrim) scrim.hidden = true;
    if (refocus && key) { var a = anchorOf(key); if (a && a.focus) a.focus({ preventScroll: true }); }
  }

  // ---- queue card ----
  function renderQueue() {
    var q = document.getElementById('queue'); if (!q || !view) return;
    var t = document.getElementById('t-queue'); if (t) t.textContent = 'Item actions';
    clear(q);
    var rows = (view.queue || []).concat(view.recent || []);
    if (!rows.length) { q.appendChild(el('p', 'd-hint', canvasMode() ? 'Nothing queued. Double-click an item to run its first action; right-click shows all of them.' : 'Nothing queued. Click an item to equip, sell, identify or trash it.')); return; }
    var ul = el('ul', 'd-list');
    rows.forEach(function (r) {
      var li = el('li'); li.dataset.state = r.state;
      li.appendChild(el('span', 't', r.state));
      li.appendChild(el('span', 'txt', verb(r)));
      if (r.reason) li.appendChild(el('span', 'sub', r.reason + (r.state === 'failed' && r.code ? ' [' + r.code + ']' : '')));
      if (r.state === 'queued' && r.reason && isStashJob(r)) { li.dataset.trip = 'home'; li.appendChild(el('span', 'sub', TRIP)); }
      if (r.result) li.appendChild(el('span', 'sub', r.result));
      if (r.state === 'queued') {
        var c = el('button', 'act act-cancel', 'Cancel'); c.type = 'button'; c.dataset.key = 'cancel-' + kindOf(r) + '-' + r.id;
        c.addEventListener('click', function () { cancel(r.id, kindOf(r)); });
        li.appendChild(c);
      }
      ul.appendChild(li);
    });
    q.appendChild(ul);
  }

  function isStashJob(r) { var k = kindOf(r); return k === 'stash' || k === 'take' || k === 'gold' || (k === 'identify' && r.place === 'stash'); }

  // ---- stash gold (schema 6): an amount box with Deposit and Withdraw, updated in place like the errand strip ----
  var gBox = null, gAmt = null, gDep = null, gWd = null, gHint = null, gSig = '', gDirty = false;
  function ensureGold() {
    if (gBox) return gBox;
    var host = document.getElementById('stashgold'); if (!host) return null;
    gBox = host; host.setAttribute('role', 'group'); host.setAttribute('aria-label', 'Stash gold');
    var lab = el('label', 'err-count', 'Gold'); lab.setAttribute('for', 'sg-n');
    gAmt = el('input', 'err-n'); gAmt.id = 'sg-n'; gAmt.setAttribute('type', 'number'); gAmt.setAttribute('min', '1'); gAmt.setAttribute('step', '1'); gAmt.setAttribute('inputmode', 'numeric'); gAmt.setAttribute('aria-label', 'How much gold to move');
    gAmt.addEventListener('input', function () { gDirty = true; });
    gDep = el('button', 'act', 'Deposit gold'); gDep.type = 'button'; gDep.dataset.act = 'deposit';
    gWd = el('button', 'act', 'Withdraw gold'); gWd.type = 'button'; gWd.dataset.act = 'withdraw';
    [[gDep, 'deposit', 'depositMax', 'noDeposit'], [gWd, 'withdraw', 'withdrawMax', 'noWithdraw']].forEach(function (c) {
      c[0].addEventListener('click', function () {
        var o = view && view.stashOps || {}, why = o[c[3]], n = Number(String(gAmt.value).trim());
        if (why) { errNote(why); return; }
        if (!(n >= 1 && Math.floor(n) === n && n <= o[c[2]])) { errNote('Enter a whole amount from 1 to ' + o[c[2]] + '.'); return; }
        gDirty = false; gold(c[1], n);
      });
    });
    gHint = el('p', 'd-hint err-hint');
    host.appendChild(lab); host.appendChild(gAmt); host.appendChild(gDep); host.appendChild(gWd); host.appendChild(gHint);
    return host;
  }
  function renderGold() {
    var o = view && view.stashOps;
    if (!ensureGold()) return;
    gBox.hidden = !o;
    if (!o) return;
    var sig = JSON.stringify(o) + '|' + view.gold + '|' + view.stashGold;
    if (sig === gSig) return;
    gSig = sig;
    if (!gDirty) gAmt.value = '';
    gAmt.setAttribute('max', String(Math.max(o.depositMax, o.withdrawMax, 1)));
    gHint.textContent = 'Backpack ' + view.gold + ' gold, stash ' + view.stashGold + ' gold. The stash opens only in town: the hero walks home at the next safe moment.';
    offBtn(gDep, o.noDeposit); offBtn(gWd, o.noWithdraw);
  }

  // ---- town errands (above the queue) ----
  var errBox = null, errNoteEl = null, errCount = null, errBuy = null, errBuyHint = null, errCainWrap = null, errCain = null, errCainHint = null, errInstantEl = null, errDirty = false, errSig = '', cainStart = -1;
  function errNote(t) { if (errNoteEl) errNoteEl.textContent = t || ''; }
  function offBtn(b, reason) { if (reason) { b.setAttribute('aria-disabled', 'true'); b.dataset.off = ''; } else { b.removeAttribute('aria-disabled'); delete b.dataset.off; } }
  function ensureErrands() {
    if (errBox) return errBox;
    var box = document.getElementById('errands'); if (!box) return null;
    errBox = box; box.setAttribute('role', 'group'); box.setAttribute('aria-label', 'Identify commands');
    var buy = el('div', 'err-buy');
    var lab = el('label', 'err-count', 'Scrolls'); lab.setAttribute('for', 'err-n');
    errCount = el('input', 'err-n'); errCount.id = 'err-n'; errCount.setAttribute('type', 'number'); errCount.setAttribute('min', '0'); errCount.setAttribute('max', '20'); errCount.setAttribute('step', '1'); errCount.setAttribute('inputmode', 'numeric'); errCount.setAttribute('aria-label', 'How many identify scrolls to buy (empty fills the tome)');
    errCount.addEventListener('input', function () { errDirty = true; });
    errBuy = el('button', 'act', 'Buy identify scrolls'); errBuy.type = 'button'; errBuy.dataset.act = 'buyid';
    errBuy.addEventListener('click', function () {
      var why = errBuy.hasAttribute('aria-disabled') ? errBuyHint.textContent : '';
      if (why) { errNote(why); return; }
      var raw = String(errCount.value === undefined ? '' : errCount.value).trim(), n = raw === '' ? 0 : Number(raw);
      var max = (view && view.identify && view.identify.buyMax) || 20;
      if (!(n >= 0 && n <= max && Math.floor(n) === n)) { errNote('Pick a whole number from 1 to ' + max + ', or leave it empty to fill the tome.'); return; }
      buyIdentify(n);
    });
    errBuyHint = el('p', 'd-hint err-hint'); buy.appendChild(lab); buy.appendChild(errCount); buy.appendChild(errBuy); buy.appendChild(errBuyHint);
    errCainWrap = el('div', 'err-cain'); errCainWrap.dataset.state = 'ready';
    errCain = el('button', 'act', 'Identify all (Cain)'); errCain.type = 'button'; errCain.dataset.act = 'cain';
    errCain.addEventListener('click', function () {
      if (errCain.hasAttribute('aria-disabled')) { errNote(errCainHint.textContent); return; }
      cainIdentify();
    });
    errCainHint = el('p', 'd-hint err-hint teaser'); errCainWrap.appendChild(errCain); errCainWrap.appendChild(errCainHint);
    errNoteEl = el('p', 'd-hint err-note'); errNoteEl.setAttribute('role', 'status');
    errInstantEl = el('p', 'd-hint err-instant'); errInstantEl.hidden = true;
    box.appendChild(errInstantEl); box.appendChild(buy); box.appendChild(errCainWrap); box.appendChild(errNoteEl);
    return box;
  }
  // renderErrands updates the strip in place (no rebuild: a number being typed must survive the 4 s poll).
  function renderErrands() {
    var id = view && view.identify;
    if (!ensureErrands()) return;
    errBox.hidden = !id;
    var ins = view.instant;
    errInstantEl.hidden = !(ins && ins.enabled);
    errInstantEl.textContent = ins && ins.enabled ? 'Instant actions: on' + (ins.why ? ' (closed now: ' + ins.why + ')' : '') : '';
    if (!id) return;
    // "Identify all" shows read-back progress: the unidentified count the view reports, never a guess
    var cainReq = (view.queue || []).some(function (r) { return kindOf(r) === 'cain' && (r.state === 'queued' || r.state === 'doing'); });
    if (cainReq && cainStart < 0) cainStart = id.unid;
    if (!cainReq) cainStart = -1;
    var prog = cainReq ? (cainReq && (view.queue || []).some(function (r) { return kindOf(r) === 'cain' && r.state === 'doing'; })
      ? 'Identified ' + Math.max(cainStart - id.unid, 0) + ' of ' + cainStart : 'waiting: not in town') : '';
    var sig = JSON.stringify(id) + (view.stashOps ? 's' : '') + prog;
    if (sig === errSig) return;
    errSig = sig;
    errCount.setAttribute('placeholder', id.tome >= 0 ? 'fill (' + id.buyDefault + ')' : 'fill');
    if (!errDirty) errCount.value = '';
    var have = id.tome >= 0 ? 'Tome of Identify: ' + id.tome + ' of ' + id.tomeCap : 'No Tome of Identify';
    errBuyHint.textContent = id.buyWhy || (have + ', ' + id.scrolls + ' loose scroll' + (id.scrolls === 1 ? '' : 's') + '. Never spends below the repair and potion reserve.');
    offBtn(errBuy, id.buyWhy);
    errCainWrap.hidden = !id.cain;
    if (id.cain === 'locked') {
      errCainWrap.className = 'err-cain d-locked'; errCainWrap.dataset.state = 'locked';
      errCainHint.textContent = id.cainWhy || 'opens when Cain is rescued';
      offBtn(errCain, id.cainWhy || 'locked');
    } else if (id.cain) {
      errCainWrap.className = 'err-cain'; errCainWrap.dataset.state = 'open';
      errCainHint.textContent = id.cainWhy || (id.unid + ' unidentified item' + (id.unid === 1 ? '' : 's') + ' carried. Free: no scroll is used.');
      offBtn(errCain, id.cainWhy);
    }
    if (prog) errCainHint.textContent = prog;
    // with the stash backend: say what Cain's Identify covers (S: the decomp finding, docs/inventory-actions.md)
    if (!prog && view.stashOps && id.cain) errCainHint.textContent += ' Covers the backpack, cube and worn items, not the stash: identify a stash item from its own menu.';
  }

  // markPending flags the grid cells of items that wait in the queue (a stash move, a take-out, an identify): the cell gets
  // data-pending=<kind>, styled as a dashed outline in inventory-actions.css. Cells of settled requests are cleared.
  var pendingKeys = [];
  function markPending() {
    pendingKeys.forEach(function (k) { var a = anchorOf(k); if (a && a.dataset) { delete a.dataset.pending; delete a.dataset.wait; delete a.dataset.waiting; } });
    pendingKeys = [];
    ((view && view.queue) || []).forEach(function (r) {
      if (!r.id || (r.state !== 'queued' && r.state !== 'doing')) return;
      var k = 'item-' + r.id, a = anchorOf(k);
      if (a && a.dataset) {
        a.dataset.pending = kindOf(r); if (r.wait) a.dataset.wait = r.wait; else delete a.dataset.wait;
        var w = waitText(r); if (w) { a.dataset.waiting = w; a.setAttribute('data-why', w); } else delete a.dataset.waiting;
        pendingKeys.push(k);
      }
      if (r.slot && kindOf(r) === 'equip') { var sa = anchorOf('slot-' + r.slot); if (sa && sa.dataset) { sa.dataset.pending = kindOf(r); pendingKeys.push('slot-' + r.slot); } }
    });
    markFeedback();
  }

  // ---- events ----
  // Short polling while a request is waiting or running: the page learns of a finished request within FAST_MS, not at the next
  // 4 s poll of inventory.js. It stops by itself when nothing is pending or sending.
  function busy() {
    if (Object.keys(fb).some(function (k) { return fb[k].state === 'sending'; })) return true;
    return ((view && view.queue) || []).some(function (r) { return r.state === 'queued' || r.state === 'doing'; });
  }
  function schedulePoll() {
    if (timer || typeof setTimeout !== 'function' || !busy()) return;
    timer = setTimeout(function () { timer = 0; if (busy()) refresh(); }, FAST_MS);
  }
  // A request that failed without the player pressing anything (lost item, no safe moment, ...) is said once, at the item and in
  // the message line: it never just disappears from the queue. The first view only learns what is old.
  var announced = null;
  function announceFailures(v) {
    var fresh = (v.recent || []).filter(function (r) { return r.state === 'failed'; });
    if (announced === null) { announced = {}; fresh.forEach(function (r) { announced[r.id + ':' + r.at] = 1; }); return; }
    fresh.forEach(function (r) {
      var key = r.id + ':' + r.at;
      if (announced[key]) return;
      announced[key] = 1;
      if (r.reason) { say(r.reason); if (r.id) setFb({ id: r.id, slot: '', kind: '' }, 'refused', r.code, r.reason); }
    });
  }
  function waitText(r) {
    if (r.state !== 'queued') return '';
    if (r.wait === 'town') return 'waiting: not in town';
    var ins = view && view.instant;
    return 'waiting: ' + ((ins && ins.why) || 'a safe moment (no monster near)');
  }
  function onView(v, fresh) {
    view = v;
    if (!fresh) ageFeedback();
    announceFailures(v);
    markPending();
    schedulePoll();
    // inventory.js hands the view over BEFORE it rebuilds the doll and the cells and redraws the queue card: the marks and our
    // own card are drawn once that call has returned, on the cells that are on the page then (review c2c3 R6)
    Promise.resolve().then(function () { if (view !== v) return; markPending(); renderQueue(); renderErrands(); renderGold(); });
    if (openKey) draw(true);
  }
  function keyOf(target) {
    var a = target && target.closest ? target.closest('[data-key]') : null;
    return a && a.dataset.key && /^(item|slot)-/.test(a.dataset.key) ? a.dataset.key : '';
  }
  // ---- gestures of the canvas ----
  // entryFor: the first entry of the menu (what Enter and a double-click run), or the one for an act; {e, m} or null.
  // Without an act it is the first entry that moves the item (identify, equip, take off, stash, take out): marks and the
  // trash are never what a double-click does, so an item with none of those has no double-click action.
  var PRIMARY = { identify: 1, equip: 1, unequip: 1, stash: 1, take: 1 };
  function entryFor(key, act) {
    var f = find(key); if (!f || !f.it) return null;
    var m = model(f), hit = null;
    m.entries.forEach(function (e) { if (!hit && (act ? e.act === act : PRIMARY[e.act])) hit = e; });
    return hit ? { e: hit, m: m } : null;
  }
  // do: run an entry from a gesture. A refused one says why at the item and in the message line and sends nothing.
  function gesture(key, act) {
    var h = entryFor(key, act), f = find(key);
    if (!h || !f) return false;
    if (h.e.reason) { refuse(f.it.id, h.e.reason); if (h.e.act === 'identify' && noToolText(h.e.reason)) offerBuy(); return true; }
    if (h.e.act === 'trash') { trashConfirm(key); return true; }
    run(h.e, h.m);
    return true;
  }
  // entries(key): the menu's own model of one item (its queued rows with their Cancel, its entries with the reasons they are refused), for
  // the selected box of the canvas: one list, so the box, the menu and the double-click can no longer disagree
  function entries(key) {
    var f = find(key); if (!f || !f.it) return null;
    var m = model(f);
    return { id: m.id, rows: m.rows.map(function (r) { return { text: r.text, state: r.state, kind: r.kind, cancel: r.cancel }; }),
      entries: m.entries.map(function (e) { return { act: e.act, slot: e.slot || '', label: e.label, reason: e.reason || '', key: e.key || '', confirm: !!e.confirm }; }) };
  }
  // runEntry: the box's button press; a refused entry says why (at the item and in the message line) and sends nothing, a trash
  // opens the menu on its confirm step (the same two steps as everywhere)
  function runEntry(key, act, slot) {
    var f = find(key); if (!f || !f.it) return false;
    var m = model(f), hit = null;
    m.entries.forEach(function (e) { if (!hit && e.act === act && (!slot || e.slot === slot)) hit = e; });
    if (!hit) return false;
    if (hit.reason) { refuse(f.it.id, hit.reason); return true; }
    if (hit.act === 'trash') { trashConfirm(key); return true; }
    run(hit, m);
    return true;
  }
  function cancelRow(key, kind) { var f = find(key); if (f && f.it) cancel(f.it.id, kind); }
  function first(key) { var h = entryFor(key); return h ? { act: h.e.act, label: h.e.label, reason: h.e.reason || '' } : null; }
  function gestures() {
    document.addEventListener('contextmenu', function (e) {
      var key = keyOf(e.target); if (!key || !find(key)) return;
      e.preventDefault();
      // a held finger is a drag (inventory-drag.js); Android's long-press contextmenu must not open a sheet under it, and iOS sends none:
      // a touch screen opens the menu from the selected box's "All actions" button instead
      if (document.body.hasAttribute('data-dragging') || lastPtr === 'touch') return;
      if (typeof CustomEvent === 'function') document.dispatchEvent(new CustomEvent('d2:select', { detail: key }));
      openMenu(key, { x: e.clientX, y: e.clientY });
    });
    document.addEventListener('dblclick', function (e) {
      var key = keyOf(e.target); if (key && find(key)) { e.preventDefault(); gesture(key, ''); }
    });
    document.addEventListener('click', function (e) {
      if (menu && menu.contains(e.target)) return;
      var key = keyOf(e.target); if (!key || !find(key)) return;
      if (e.ctrlKey || e.metaKey) { var f = find(key); gesture(key, f.place === 'stash' ? 'take' : 'stash'); }
      else if (e.altKey) gesture(key, 'sell');
    });
    document.addEventListener('keydown', function (e) {
      var t = e.target, typing = t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT');
      var key = keyOf(t); if (!key || typing || openKey || !find(key)) return;
      if (e.key === 'Enter' && !e.ctrlKey && !e.altKey) { e.preventDefault(); gesture(key, ''); }
      else if (e.key === 'Delete') { e.preventDefault(); gesture(key, 'trash'); }
      else if (e.key === 'ContextMenu' || (e.key === 'F10' && e.shiftKey)) { e.preventDefault(); openMenu(key); }
      else if ((e.key === 'm' || e.key === 'M') && !e.ctrlKey && !e.metaKey) { var h = entryFor(key, 'merc'); if (h) { e.preventDefault(); gesture(key, 'merc'); } }
      else if ((e.key === 'r' || e.key === 'R') && !e.ctrlKey && !e.metaKey) { var r = entryFor(key, 'repair'); if (r) { e.preventDefault(); gesture(key, 'repair'); } }
    });
  }

  function init() {
    if (canvasMode()) gestures();
    document.addEventListener('d2:inventory', function (e) { onView(e.detail); });
    document.addEventListener('keydown', function (e) {
      kbd = true;
      if (!openKey) return;
      if (e.key === 'Escape') { closeMenu(true); return; }
      var list = Array.prototype.slice.call(focusables()), i = list.indexOf(document.activeElement);
      if (!list.length) return;
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Home' || e.key === 'End' || e.key === 'Tab') {
        var n = e.key === 'Home' ? 0 : e.key === 'End' ? list.length - 1 : e.key === 'ArrowUp' || (e.key === 'Tab' && e.shiftKey) ? (i <= 0 ? list.length - 1 : i - 1) : (i + 1) % list.length;
        e.preventDefault(); list[n].focus({ preventScroll: true });
      }
    });
    document.addEventListener('pointerdown', function (e) { kbd = false; lastPtr = e.pointerType || 'mouse'; });
    document.addEventListener('pointerdown', function (e) {
      if (openKey && menu && !menu.contains(e.target) && keyOf(e.target) !== openKey) closeMenu(false);
    });
    document.addEventListener('click', function (e) {
      if (menu && menu.contains(e.target)) return;
      var key = keyOf(e.target);
      if (!key || !find(key)) return;
      if (canvasMode()) { if (openKey) closeMenu(false); return; } // the canvas: a click selects (inventory-canvas.js), the menu is the right-click menu
      if (key === openKey) { closeMenu(true); return; }
      openMenu(key);
    });
    window.addEventListener('resize', function () { if (openKey && menu && !menu.hidden) position(); }, { passive: true });
    window.addEventListener('scroll', function () { if (openKey && menu && !menu.hidden) position(); }, { passive: true });
  }

  // trashConfirm opens the item's menu already on the confirm step (the drag of a tile outside every grid): the same two
  // clicks as the menu, so a drop alone never trashes.
  function trashConfirm(key) { openMenu(key); if (!openKey) return; var f = find(key); if (f && f.it && !f.it.noTrash) { confirming = true; draw(true); } }
  var api = { post: post, offerBuy: offerBuy, base: base, refresh: refresh, stash: stash, take: take, gold: gold, equip: equip, unequip: unequip, trashConfirm: trashConfirm, say: say, note: setNote, trash: trash, cancel: cancel, first: first, entries: entries, runEntry: runEntry, cancelRow: cancelRow, gesture: gesture, refuse: refuse, identify: identify, buyIdentify: buyIdentify, cainIdentify: cainIdentify, open: openMenu, close: function () { closeMenu(false); } };
  if (typeof window !== 'undefined') window.d2InvActions = api;
  if (typeof window !== 'undefined' && window.__actTest) {
    window.__actTest({ poll: function () { return timer; }, announce: announceFailures, init: init, onView: onView, model: model, entryFor: entryFor, gesture: gesture, first: first, fb: function () { return fb; }, post: post,  find: find, open: openMenu, close: closeMenu, place: place,
      state: function () { return { key: openKey, node: menu, confirming: confirming, sheet: sheet }; }, renderQueue: renderQueue, renderErrands: renderErrands, renderGold: renderGold, goldBox: function () { return { box: gBox, amt: gAmt, dep: gDep, wd: gWd, hint: gHint }; }, errands: function () { return { instant: errInstantEl, hint: errCainHint, box: errBox, count: errCount, buy: errBuy, cain: errCain, wrap: errCainWrap, note: errNoteEl }; }, setSheet: function (s) { sheet = !!s; } });
    return;
  }
  init();
})();
