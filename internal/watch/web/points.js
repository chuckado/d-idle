// Points page (docs/points.md): GET /api/hero/points, one + button per attribute, the autopilot switch and the trail.
// Writes go to the answer route (another port) and always read the sheet back afterwards; text through textContent only.
(function () {
  'use strict';
  var $ = function (id) { return document.getElementById(id); };
  var base = null, busy = false;

  function el(t, c, x) { var e = document.createElement(t); if (c) e.className = c; if (x !== undefined) e.textContent = x; return e; }
  function clear(n) { while (n.firstChild) n.removeChild(n.firstChild); }
  function say(t) { var m = $('msg'); m.textContent = t || ''; m.hidden = !t; }

  function answerBase() {
    if (base) return Promise.resolve(base);
    return fetch('/api/config', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (c) {
      base = c.answerBase ? c.answerBase : location.protocol + '//' + location.hostname + ':' + c.answerPort;
      return base;
    });
  }

  function post(path, body) {
    return answerBase().then(function (u) {
      return fetch(u + path, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    }).then(function (r) { return r.json().then(function (j) { j.status = r.status; return j; }); });
  }

  function act(path, body, done) {
    if (busy) return;
    busy = true; say('');
    post(path, body).then(function (j) {
      if (!j.ok) say(j.reason || j.error || ('error ' + j.status));
      else if (done) done(j);
    }).catch(function () { say('the answer route is not reachable'); }).then(function () { busy = false; load(); });
  }

  function render(s) {
    $('sum').textContent = [s.hero, s.class, s.level ? 'level ' + s.level : ''].filter(Boolean).join(' - ');
    $('waiting').textContent = s.waiting || (s.note && !s.stats.length ? s.note : 'Nothing is waiting: every point is spent.');
    var sk = $('skillline'); clear(sk);
    if (s.skillPoints > 0) {
      sk.appendChild(document.createTextNode(s.skillPoints + (s.skillPoints === 1 ? ' skill point' : ' skill points') + ' to spend: '));
      var a = el('a', 'd-link', 'open the skill trees'); a.href = s.skillsPage || '/watch/skills'; sk.appendChild(a);
    }
    $('statmeta').textContent = s.statPoints + ' unspent';
    var ul = $('stats'); clear(ul);
    s.stats.forEach(function (r) {
      var li = el('li', 'pt-stat');
      var nm = el('span', 'pt-name', r.name); li.appendChild(nm);
      var plus = el('button', 'd-btn pt-plus', '+'); plus.type = 'button';
      plus.setAttribute('aria-label', 'Add one point to ' + r.name);
      plus.disabled = !s.enabled || !r.canAdd;
      plus.addEventListener('click', function () { act('/api/hero/points/stat', { stat: r.key }); });
      li.appendChild(plus);
      li.appendChild(el('p', 'pt-prev', r.now + (r.spent ? ' (' + r.spent + ' spent by you)' : '') + ' - ' + r.preview));
      var need = r.needsGear > 0 ? 'Your gear needs ' + r.needsGear + ' ' + r.name + '.' : '';
      if (need) { var p = el('p', 'pt-need', need); p.setAttribute('data-short', r.needsGear > r.now ? '1' : '0'); li.appendChild(p); }
      if (!r.canAdd && r.why && s.statPoints > 0) li.appendChild(el('p', 'pt-why', r.why));
      ul.appendChild(li);
    });
    $('note').textContent = (s.note || '') + (s.verbs && s.verbs.via ? (s.note ? ' ' : '') + 'Spent through: ' + s.verbs.via + '.' : '');
    var g = $('gear'); clear(g);
    (s.gear || []).forEach(function (x) {
      var t = x.item + ' (' + x.slot + '): ' + [x.str ? x.str + ' Strength' : '', x.dex ? x.dex + ' Dexterity' : ''].filter(Boolean).join(', ');
      if (x.shortStr || x.shortDex) t += ' - you lack ' + [x.shortStr ? x.shortStr + ' Strength' : '', x.shortDex ? x.shortDex + ' Dexterity' : ''].filter(Boolean).join(' and ');
      g.appendChild(el('li', '', t));
    });
    if (!(s.gear || []).length) g.appendChild(el('li', '', 'Nothing you wear has a requirement.'));
    $('autodesc').textContent = s.autopilot
      ? 'On: the hero spends new points by itself from its plan table.'
      : 'Off: new points are banked and you are asked. Nothing is spent behind your back.';
    var ab = $('auto'); ab.textContent = s.autopilot ? 'Turn autopilot off' : 'Turn autopilot on';
    ab.onclick = function () { act('/api/hero/points/autopilot', { on: !s.autopilot }); };
    var tr = $('trail'); clear(tr);
    (view.recent || []).forEach(function (e) {
      var li = el('li', '', new Date(e.at).toLocaleTimeString() + ' - ' + e.action + (e.target ? ' ' + e.target : '') + ' - ' + (e.ok ? 'ok' : (e.reason || e.code || 'refused')) + ' (' + e.remote + (e.agent ? ', ' + e.agent : '') + ')');
      li.setAttribute('data-ok', e.ok ? '1' : '0'); tr.appendChild(li);
    });
    if (!(view.recent || []).length) tr.appendChild(el('li', '', 'Nothing yet.'));
  }

  // The shadow planner's line (GET /api/hero/plan, docs/agency.md): off or unreachable hides it.
  function loadPlan() {
    fetch('/api/hero/plan', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (j) {
      var b = $('buildline');
      b.textContent = j && j.enabled && j.line ? j.line : '';
      b.hidden = !b.textContent;
      if (j && j.warnings && j.warnings.length) b.title = j.warnings.map(function (w) { return w.text; }).join('\n');
    }).catch(function () { $('buildline').hidden = true; });
  }

  // The Build tab (docs/autoalloc.md, card buildcards): GET /api/hero/build gives cards from the catalogue; a card's detail comes
  // on demand (?detail=ID); Choose opens ONE confirm sheet and POSTs the table (variant) the player picked, which locks it.
  var bview = null, bkey = '', open = {}, details = {}, sheet = null, BC = window.dBuildCards;
  function bstate() { var st = bview.state || {}; return { autoAlloc: !!bview.autoAlloc, locked: !!st.locked, chosen: st.build, label: st.label }; }
  function renderBuild(b) {
    var key = JSON.stringify(b);
    if (key === bkey) return; // nothing changed: keep the DOM, the focus and the open details as they are
    bkey = key; bview = b;
    var st = b.state || {};
    $('bmeta').textContent = st.locked ? 'locked' : (b.autoAlloc ? 'on' : 'off');
    var line = st.locked
      ? 'Locked to ' + (st.label || st.build) + ' since level ' + st.lockedAtLevel + '. Respec to change.'
      : 'No build chosen: points are banked for you. Choose a build and the hero puts every new point where it says.';
    $('bstate').textContent = line;
    var off = $('boffnote'); off.hidden = !!b.autoAlloc;
    off.textContent = b.autoAlloc ? '' : 'Auto-allocation is off (knob agency.autoAlloc): you can read about the builds here, but choosing one stays closed until it is on.';
    var u = b.unspent || {};
    var act = $('bact');
    if (act) {
      act.hidden = !st.locked;
      act.textContent = !st.locked ? '' : 'Spent by the build so far this run: ' + (b.spent || 0) + '. Unspent now: ' + (u.stat || 0) + ' stat, ' + (u.skill || 0) + ' skill.' + (b.waiting && ((u.stat || 0) + (u.skill || 0)) > 0 ? ' Waiting: ' + b.waiting : '');
    }
    var pl = $('bplan');
    if (pl) {
      pl.hidden = !(b.plan || b.noPlan);
      pl.textContent = pl.hidden ? '' : [b.plan, b.noPlan].filter(Boolean).join(' ');
    }
    var h = $('bheld'); h.hidden = !(b.held && b.held.length);
    h.textContent = h.hidden ? '' : b.held.map(function (x) { return x.skill + ' (level ' + x.clvl + '): ' + x.reason; }).join(' ');
    var o = $('boff'); o.hidden = !(b.offBuild && b.offBuild.length);
    o.textContent = o.hidden ? '' : 'Off-build points you spent yourself (the allocator plans around them): ' + b.offBuild.join(', ') + '.';
    $('brespecline').textContent = b.respec ? 'A respec now would cost ' + b.respec.price + ' gold and refund every skill and stat point. ' + b.respec.note : '';
    renderCards();
  }
  function renderCards() {
    var ul = $('bcards'), b = bview; clear(ul);
    var err = $('bcerr'); err.hidden = !b.cardsError; err.textContent = b.cardsError || '';
    if (!(b.cards || []).length) { if (!b.cardsError) ul.appendChild(el('li', 'bc-empty', b.class ? 'No builds are listed for this class yet.' : 'The hero is not in a game yet, so no builds are listed.')); return; }
    b.cards.forEach(function (c) {
      var li = BC.cardNode(document, c, { open: !!open[c.id], onDetails: function (card, btn) { toggle(card, btn); } });
      if (open[c.id] && details[c.id]) li.appendChild(BC.detailNode(document, details[c.id], bstate(), { onChoose: choose }));
      ul.appendChild(li);
    });
  }
  function toggle(card, btn) {
    if (open[card.id]) { open[card.id] = false; renderCards(); return; }
    open[card.id] = true;
    if (details[card.id]) { renderCards(); return; }
    fetch('/api/hero/build?detail=' + encodeURIComponent(card.id), { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (j) {
      if (j.code || j.error) { open[card.id] = false; say(j.error || 'the build details are not available'); return; }
      details[card.id] = j; renderCards();
    }).catch(function () { open[card.id] = false; say('could not read the build details'); });
  }
  function choose(det, btn) {
    var vs = det.variants || [];
    var exp = function (n) { return vs.filter(function (v) { return v.name === n && v.tier === 'experimental'; }).length > 0; };
    sheet = BC.chooseSheet(document, det, bview && bview.respec ? bview.respec.price : undefined, { opener: btn, onCancel: function () { sheet = null; },
      onConfirm: function (name) { sheet = null; pickBuild(name, exp(name)); } });
  }
  function pickBuild(name, confirmed) {
    post('/api/hero/build', { build: name, confirm: confirmed }).then(function (j) {
      if (j.needsConfirm && !confirmed) { pickBuild(name, true); return; }
      say(j.ok ? '' : (j.reason || j.error || ('error ' + j.status)));
      loadBuild(); $('t-build').focus();
    }).catch(function () { say('the answer route is not reachable'); });
  }
  function loadBuild() {
    if (sheet) return; // a sheet is open: do not redraw under it
    fetch('/api/hero/build', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (j) {
      if (j.code) { $('bstate').textContent = j.error || 'the build picker is not available'; return; }
      renderBuild(j);
    }).catch(function () { $('bstate').textContent = 'could not read the build state'; });
  }

  var view = { stats: [] };
  function load() {
    loadPlan(); loadBuild();
    fetch('/api/hero/points', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (j) {
      if (j.code) { say(j.error || 'the points desk is not available'); return; }
      view = j; render(j);
    }).catch(function () { say('could not read the points sheet'); });
  }
  load();
  setInterval(function () { if (!busy && !document.hidden) load(); }, 5000);
})();
