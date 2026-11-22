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

  // Automatic Allocate Skills (docs/autoalloc.md): GET /api/hero/build, POST the pick (it locks), refusals read back in words.
  var bview = null, pickerOpen = false;
  function renderBuild(b) {
    bview = b;
    var st = b.state || {};
    $('bmeta').textContent = st.locked ? 'locked' : (b.autoAlloc ? 'on' : 'off');
    var line = st.locked
      ? 'Locked to ' + (st.label || st.build) + ' since level ' + st.lockedAtLevel + '. Respec to change.'
      : 'No build chosen: points are banked for you. Choose a build and the hero puts every new point where it says.';
    if (!b.autoAlloc) line += ' The allocator is switched off (knob agency.autoAlloc), so nothing is spent automatically yet.';
    $('bstate').textContent = line;
    var h = $('bheld'); h.hidden = !(b.held && b.held.length);
    h.textContent = h.hidden ? '' : b.held.map(function (x) { return x.skill + ' (level ' + x.clvl + '): ' + x.reason; }).join(' ');
    var o = $('boff'); o.hidden = !(b.offBuild && b.offBuild.length);
    o.textContent = o.hidden ? '' : 'Off-build points you spent yourself (the allocator plans around them): ' + b.offBuild.join(', ') + '.';
    $('brespecline').textContent = b.respec ? 'A respec now would cost ' + b.respec.price + ' gold and refund every skill and stat point. ' + b.respec.note : '';
    var bo = $('bopen');
    bo.disabled = !!st.locked;
    bo.title = st.locked ? 'Locked. Respec to change the build.' : '';
    bo.onclick = function () { pickerOpen = !pickerOpen; renderPicker(); };
    renderPicker();
  }
  function renderPicker() {
    var ul = $('bpicker'), b = bview; clear(ul);
    $('bopen').setAttribute('aria-expanded', pickerOpen ? 'true' : 'false');
    ul.hidden = !pickerOpen || !b;
    if (ul.hidden) return;
    if (!b.builds.length) { ul.appendChild(el('li', 'pt-build', 'No build files for this class.')); return; }
    b.builds.forEach(function (x) {
      var li = el('li', 'pt-build'); li.setAttribute('data-tier', x.tier); li.setAttribute('data-chosen', b.state.build === x.name ? '1' : '0');
      var h = el('h3', '', x.label + ' '); h.appendChild(el('span', 'pt-tier', '(' + x.tier + ')')); li.appendChild(h);
      var bad = (x.needs || []).filter(function (n) { return n.tier === 'unsupported'; });
      var meh = (x.needs || []).length - bad.length;
      if (bad.length) li.appendChild(el('p', 'pt-warn', 'The hero cannot use these yet and will hold the point for you: ' + bad.map(function (n) { return n.skill; }).join(', ') + '.'));
      if (meh) li.appendChild(el('p', '', meh + ' more skill' + (meh === 1 ? ' is' : 's are') + ' not proven in the fighter yet.'));
      var d = el('details'); d.appendChild(el('summary', '', 'First ' + x.preview.length + ' levels'));
      var pl = el('ul', 'pt-prev-list');
      x.preview.forEach(function (r) {
        var parts = (r.skills || []).map(function (s) { return s.skill + (s.points > 1 ? ' x' + s.points : '') + (s.tier && s.tier !== 'supported' ? ' [' + s.tier + ']' : ''); });
        var stats = Object.keys(r.stats || {}).map(function (k) { return '+' + r.stats[k] + ' ' + k; });
        pl.appendChild(el('li', '', 'Level ' + r.clvl + ': ' + parts.concat(stats).join(', ')));
      });
      d.appendChild(pl); li.appendChild(d);
      var btn = el('button', 'd-btn', b.state.build === x.name ? 'Chosen' : 'Choose and lock'); btn.type = 'button'; btn.setAttribute('data-size', 'sm');
      btn.disabled = !!b.state.locked;
      btn.addEventListener('click', function () { pickBuild(x, false); });
      li.appendChild(btn);
      ul.appendChild(li);
    });
  }
  function pickBuild(x, confirmed) {
    var price = bview && bview.respec ? bview.respec.price : '?';
    if (!confirmed && !window.confirm('Lock to ' + x.label + '? Every new point follows it. Changing it later means a respec (about ' + price + ' gold now) that refunds all skill and stat points.')) return;
    post('/api/hero/build', { build: x.name, confirm: confirmed }).then(function (j) {
      if (j.needsConfirm && !confirmed) { if (window.confirm(j.reason)) pickBuild(x, true); return; }
      if (!j.ok) say(j.reason || j.error || ('error ' + j.status)); else { say(''); pickerOpen = false; }
      loadBuild();
    }).catch(function () { say('the answer route is not reachable'); });
  }
  function loadBuild() {
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
