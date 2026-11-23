// web/settings.js in the stub DOM (node test, card uisettings): the Settings sheet on the shared sheet component: focus in, Escape and focus
// back, the two sections with visible labels, a reason line for every off control (why.js), the Rules calls and the Alerts settings.
'use strict';
const path = require('path');
const { Ev, makeDoc } = require('./nav_dom.cjs');
const dir = process.argv[2];
const Sheet = require(path.join(dir, 'sheet.cjs')), Toast = require(path.join(dir, 'toast.cjs')), Why = require(path.join(dir, 'why.cjs')), Set = require(path.join(dir, 'settings.cjs'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const key = (n, k) => n.dispatchEvent(Object.assign(new Ev('keydown', { cancelable: true }), { key: k }));
const tick = () => new Promise((r) => setTimeout(r, 0));
function newDoc() {
  const doc = makeDoc();
  doc.N.prototype.remove = function () { if (this.parent) { this.parent.children = this.parent.children.filter((x) => x !== this); this.parent = null; } };
  return doc;
}
const rule = (k, text, revoked) => ({ key: k, text, risk: 'low', revoked: !!revoked });
// a fake page: /api/rules answers the current list; the answer route records its posts
function rig(rules, over) {
  const doc = newDoc(), opener = doc.createElement('button'); opener.textContent = 'Settings'; doc.body.appendChild(opener); opener.focus();
  const st = { rules, posts: [], answer: { ok: true }, kv: {}, getFails: false };
  const fetch = (url, o) => {
    if (o && o.method === 'POST') { st.posts.push({ url, body: JSON.parse(o.body), creds: o.credentials }); return Promise.resolve({ status: 200, json: () => Promise.resolve(st.answer) }); }
    if (st.getFails) return Promise.resolve({ ok: false, status: 500, json: () => Promise.resolve({}) });
    return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve({ rules: st.rules, live: st.rules.filter((r) => !r.revoked).length }) });
  };
  const env = Object.assign({ doc, sheet: Sheet, toast: Toast, fetch, base: () => Promise.resolve('http://answer.test'), store: { get: (k) => st.kv[k] || null, set: (k, v) => { st.kv[k] = v; } },
    notif: null, opener, section: 'rules', later: () => 0 }, over);
  return { doc, opener, st, env };
}
const nodes = (d) => d.body.all();
const text = (d) => nodes(d).map((n) => n.textContent).join('|');
const byKey = (d, k) => nodes(d).filter((n) => n.dataset && n.dataset.key === k);

(async () => {
  // 1. the sheet: a titled modal dialog, focus inside, Escape closes and the focus goes back
  {
    const r = rig([rule('a', 'When asked X, I do Y.'), rule('b', 'When asked P, I do Q.', true)]);
    const h = Set.open(r.env); await h.ready;
    ok(h.sheet.attrs.role === 'dialog' && h.sheet.attrs['aria-modal'] === 'true', 'Settings is a modal dialog');
    ok(r.doc.getElementById(h.sheet.attrs['aria-labelledby']).textContent === 'Settings', 'titled Settings');
    ok(h.sheet.contains(r.doc.activeElement), 'focus is inside the sheet');
    ok(nodes(r.doc).filter((n) => n.tag === 'h3').map((n) => n.textContent).join() === 'Rules,Alerts', 'two sections: Rules, Alerts');
    ok(h.confirm.textContent === 'Close' && !h.cancel.parent, 'one Close button');
    key(h.node, 'Escape');
    ok(!r.doc.body.children.includes(h.node) && r.doc.activeElement === r.opener, 'Escape closes it and the focus goes back to the opener');
  }
  // 2. Rules: the list, one Revoke per live rule, a labelled button, the call and the reload
  {
    const r = rig([rule('a', 'When asked X, I do Y.'), rule('b', 'When asked P, I do Q.', true)]);
    const h = Set.open(r.env); await h.ready;
    ok(text(r.doc).includes('When asked X, I do Y.') && text(r.doc).includes('revoked'), 'the rules and the revoked mark are listed');
    const rev = byKey(r.doc, 'revoke');
    ok(rev.length === 1 && rev[0].textContent === 'Revoke' && /When asked X/.test(rev[0].attrs['aria-label']), 'one visible Revoke for the live rule, named for its rule');
    rev[0].click(); await tick(); await tick();
    ok(r.st.posts.length === 1 && r.st.posts[0].url === 'http://answer.test/api/revoke' && r.st.posts[0].body.key === 'a' && r.st.posts[0].creds === 'include', 'revoke posts the key to the answer route with the cookie');
    ok(/Rule revoked/.test(text(r.doc)), 'a good toast says so');
    r.st.answer = { ok: false, error: 'not allowed' };
    r.st.rules = [rule('c', 'Another rule.')]; await h.reload();
    byKey(r.doc, 'revoke')[0].click(); await tick(); await tick();
    ok(/not allowed/.test(text(r.doc)), 'a refused call shows the error');
    ok(nodes(r.doc).some((n) => n.dataset && n.dataset.sev === 'bad'), 'as a bad toast');
  }
  // 3. Revoke all: disabled with a reason when nothing is active; confirm sheet; the call
  {
    const r = rig([rule('b', 'When asked P, I do Q.', true)]);
    const h = Set.open(r.env); await h.ready;
    const all = byKey(r.doc, 'revoke-all')[0];
    ok(all.disabled && all.textContent === 'Revoke all', 'Revoke all is off with no live rule');
    ok(Why.unexplained(nodes(r.doc)).length === 0, 'and says why: ' + Why.unexplained(nodes(r.doc)).join('; '));
    r.st.rules = [rule('a', 'x'), rule('c', 'y')]; await h.reload();
    ok(!all.disabled && Why.unexplained(nodes(r.doc)).length === 0, 'on with live rules');
    all.click(); await tick();
    const back = r.doc.body.children.filter((n) => n.className === 'd-sheet-back');
    ok(back.length === 2, 'a confirm sheet opens over Settings');
    const ask = back[1].children[0];
    ok(ask.attrs.role === 'alertdialog', 'the confirm is the danger dialog');
    ask.all().find((n) => n.className.includes('d-sheet-ok')).click(); await tick(); await tick(); await tick();
    ok(r.st.posts.length === 1 && r.st.posts[0].url.endsWith('/api/revoke-all') && r.st.posts[0].body.confirm === true, 'confirming posts revoke-all');
  }
  // 4. the rules cannot be loaded
  {
    const r = rig([rule('a', 'x')]); r.st.getFails = true;
    const h = Set.open(r.env); await h.ready;
    ok(text(r.doc).includes('The rules could not be loaded.'), 'a load failure is said');
    ok(byKey(r.doc, 'revoke-all')[0].disabled && Why.unexplained(nodes(r.doc)).length === 0, 'Revoke all is off with a reason');
  }
  // 5. Alerts: the sound checkbox keeps the device setting, the notification button follows the permission
  {
    const states = [[null, /no desktop notifications/, true], [{ permission: 'granted' }, /are on while this tab is open/, true], [{ permission: 'denied' }, /blocked in the browser settings/, true], [{ permission: 'default', requestPermission() { this.permission = 'granted'; this.asked = 1; return Promise.resolve('granted'); } }, /asks only when you press/, false]];
    for (const [N, re, off] of states) {
      const r = rig([]);
      r.env.notif = N;
      const h = Set.open(r.env); await h.ready;
      const b = byKey(r.doc, 'alert-notify')[0];
      ok(b.disabled === off, 'notification button ' + (off ? 'off' : 'on') + ' for ' + JSON.stringify(N && N.permission));
      ok(re.test(text(r.doc)), 'the line under it says: ' + re);
      ok(Why.unexplained(nodes(r.doc)).length === 0, 'every off control has a reason (' + (N && N.permission) + '): ' + Why.unexplained(nodes(r.doc)).join('; '));
      if (N && N.permission === 'default') { b.click(); await tick(); await tick(); ok(N.asked === 1 && b.disabled && /are on/.test(text(r.doc)), 'pressing it asks, then shows it is on'); }
    }
    const r = rig([]); r.env.store.set('d2lod.alerts.sound', '1');
    const h = Set.open(r.env); await h.ready;
    const sound = byKey(r.doc, 'alert-sound')[0];
    ok(sound.checked === true, 'the sound checkbox shows the saved setting');
    const lab = nodes(r.doc).find((n) => n.tag === 'label' && n.attrs.for === sound.id);
    ok(lab && /Play a sound on a new alert/.test(lab.children.map((c) => c.textContent).join('')) && lab.contains(sound), 'the checkbox has a visible label');
    sound.checked = false; sound.dispatchEvent(new Ev('change')); ok(r.st.kv['d2lod.alerts.sound'] === '0', 'unchecking stores 0 under the key the old page reads');
  }
  // 5b. the Loot rules entry stays reachable from Settings (the old Inventory entry): a labelled link to the live screen's sheet
  {
    const r = rig([]); const h = Set.open(r.env); await h.ready;
    const a = byKey(r.doc, 'lootrules-link')[0];
    ok(a && a.tag === 'a' && a.textContent === 'Loot rules...' && a.href === '/watch?live=1&lootrules=1', 'Settings has a Loot rules link to the live screen sheet');
    ok(a && r.doc.getElementById(a.attrs['aria-describedby']) && /picks up, stashes and sells/.test(r.doc.getElementById(a.attrs['aria-describedby']).textContent), 'with a visible line saying what it is');
  }
  // 6. every button and checkbox carries a visible label
  {
    const r = rig([rule('a', 'x')]); const h = Set.open(r.env); await h.ready;
    for (const n of nodes(r.doc).filter((n) => n.tag === 'button')) ok(String(n.textContent).trim() !== '', 'a button without a visible label: ' + (n.dataset.key || n.className));
  }
  console.log(JSON.stringify({ fails }));
  process.exit(0);
})();
