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

  var view = { stats: [] };
  function load() {
    loadPlan();
    fetch('/api/hero/points', { cache: 'no-store' }).then(function (r) { return r.json(); }).then(function (j) {
      if (j.code) { say(j.error || 'the points desk is not available'); return; }
      view = j; render(j);
    }).catch(function () { say('could not read the points sheet'); });
  }
  load();
  setInterval(function () { if (!busy && !document.hidden) load(); }, 5000);
})();
