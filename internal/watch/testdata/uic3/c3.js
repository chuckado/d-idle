// uic3 (the tooltip, the right-click menu, gestures, instant feedback of the character canvas): DOM goldens and behaviour checks of
// web/inventory-tip.js and web/inventory-actions.js with a small DOM stub. UPDATE=1 rewrites the goldens.
// Usage: node c3.js path/to/inventory-tip.js path/to/inventory-actions.js
const fs = require('fs'), vm = require('vm'), assert = require('assert'), path = require('path');
const dir = __dirname;
let bad = 0;
const check = (name, got) => {
  const file = path.join(dir, name + '.txt');
  if (process.env.UPDATE) { fs.writeFileSync(file, got); return; }
  if (!fs.existsSync(file)) { console.error('missing golden ' + file); bad++; return; }
  if (fs.readFileSync(file, 'utf8') !== got) { console.error('differs from ' + file + '\n' + got); bad++; }
};

let focused = null;
function mk(tag) {
  const e = { tag, className: '', dataset: {}, children: [], textContent: '', attrs: {}, listeners: {}, hidden: false, style: {}, id: '' };
  e.classList = { add(c) { e.className = (e.className + ' ' + c).trim(); } };
  Object.defineProperty(e, 'firstChild', { get() { return e.children[0] || null; } });
  e.appendChild = (k) => { k.parent = e; e.children.push(k); return k; };
  e.removeChild = (k) => { e.children = e.children.filter((x) => x !== k); return k; };
  e.setAttribute = (k, v) => { e.attrs[k] = String(v); };
  e.getAttribute = (k) => (k in e.attrs ? e.attrs[k] : null);
  e.hasAttribute = (k) => k in e.attrs;
  e.removeAttribute = (k) => { delete e.attrs[k]; };
  e.addEventListener = (t, f) => { (e.listeners[t] = e.listeners[t] || []).push(f); };
  e.contains = (x) => x === e || e.children.some((c) => c.contains && c.contains(x));
  e.offsetWidth = 270; e.offsetHeight = 200;
  e.focus = (o) => { if (!o || o.preventScroll !== true) throw new Error('focus must pass preventScroll'); focused = e; };
  e.scrollIntoView = () => { throw new Error('nothing may scroll the page'); };
  e.querySelector = (s) => { const c = s.slice(1); const f = (n) => { for (const k of n.children) { if (k.className.split(/\s+/).includes(c)) return k; const r = f(k); if (r) return r; } return null; }; return f(e); };
  e.querySelectorAll = (s) => { const out = []; const f = (n) => n.children.forEach((k) => { if (s === 'button.act' ? (k.tag === 'button' && k.className.split(/\s+/).includes('act')) : k.className.split(/\s+/).includes(s.slice(1))) out.push(k); f(k); }); f(e); return out; };
  return e;
}
function ser(e, d) {
  const attrs = Object.keys(e.attrs).sort().map((k) => k + '=' + e.attrs[k]);
  const data = Object.keys(e.dataset).sort().map((k) => 'data-' + k + '=' + e.dataset[k]);
  const cls = e.className.split(/\s+/).filter(Boolean).sort().join('.');
  let out = '  '.repeat(d) + e.tag + (cls ? '.' + cls : '') + ' [' + attrs.concat(data).join(' ') + ']' + (e.textContent ? ' ' + JSON.stringify(e.textContent) : '') + (e.hidden ? ' hidden' : '') + '\n';
  e.children.forEach((c) => { out += ser(c, d + 1); });
  return out;
}

// ---- the views: two classes, nothing in the page knows which
const side = (ok, ...reasons) => ({ ok, reasons });
const req = (stat, need, have, what) => ({ code: 'requirements', stat, need, have, text: 'needs ' + what + ' (you have ' + have + ')' });
function view(cls) {
  const v = {
    schema: 10, class: cls, inTown: true, gold: 100, stashGold: 50, queue: [], recent: [],
    merc: { state: 'ready', owned: true, alive: true, kind: 'desert', level: 14, actWhy: 'the bot cannot change the mercenary\'s gear yet' },
    equipment: {
      rhand: { id: 90, code: 'sst', name: 'Short Staff', quality: 'normal', identified: true, w: 1, h: 3, props: ['Two-Hand Damage: 1 to 5'], maxDur: 18, dur: 18 },
      lring: { id: 91, code: 'rin', name: 'Raven Grip', quality: 'rare', identified: true, w: 1, h: 1, props: ['Fire Resist +8%', '+6 Life'] },
      rring: { id: 92, code: 'rin', name: 'Coral Band', quality: 'magic', identified: true, w: 1, h: 1, props: ['+10 Life'] },
      torso: { id: 93, code: 'qui', name: 'Quilted Armor', quality: 'magic', identified: true, w: 2, h: 3, props: ['Defense: 11'], maxDur: 24, dur: 0 },
    },
    backpack: [
      { id: 1, code: 'rin', name: 'Plain Ring', quality: 'normal', identified: true, w: 1, h: 1, canWear: ['lring', 'rring'], stashOp: 'stash', usable: { hero: side(true), merc: side(false, { code: 'merctype', text: 'a Desert Guard wears helm, armor, spear or polearm' }) } },
      { id: 2, code: 'cap', name: 'War Cap', quality: 'magic', identified: true, w: 2, h: 2, props: ['Defense: 14'], req: { lvl: 9 }, noWear: 'needs level 9 (you have 7)', stashOp: 'stash', usable: { hero: side(false, req('lvl', 9, 7, 'level 9')), merc: side(true) } },
      { id: 3, code: 'plt', name: 'Light Plate', quality: 'normal', identified: true, w: 2, h: 3, req: { str: 41 }, noWear: 'needs 41 strength (you have 38)', usable: { hero: side(false, req('str', 41, 38, '41 strength')), merc: side(true) } },
      { id: 4, code: 'bow', name: 'Hunter Bow', quality: 'normal', identified: true, w: 2, h: 3, req: { dex: 45, str: 20 }, noWear: 'needs 45 dexterity (you have 30)', usable: { hero: side(false, req('dex', 45, 30, '45 dexterity')) } },
      { id: 5, code: 'tar', name: 'Targe', quality: 'magic', identified: true, w: 2, h: 2, req: { str: 16, class: ['Paladin'] }, noWear: 'Paladin only', usable: { hero: side(false, { code: 'class', text: 'only a Paladin can use it' }), merc: side(false, { code: 'merctype', text: 'a Desert Guard wears helm, armor, spear or polearm' }) } },
      { id: 6, code: 'rin', name: 'Ring', quality: 'rare', identified: false, w: 1, h: 1, noIdentify: '', usable: { hero: side(false, { code: 'identified', text: 'it is not identified' }) } },
      { id: 7, code: 'plt', name: 'Chipped Plate', quality: 'normal', identified: true, w: 2, h: 3, maxDur: 20, dur: 0, canWear: ['torso'], props: ['Defense: 22'], usable: { hero: side(true, { code: 'broken', text: 'it is broken: repair it', warn: true }) } },
      { id: 8, code: 'stf', name: 'War Staff', quality: 'magic', identified: true, w: 1, h: 3, canWear: ['rhand'], props: ['Two-Hand Damage: 2 to 8', '+2 to Charged Bolt'], req: { lvl: 6 }, stashOp: 'stash', usable: { hero: side(true), merc: side(true) } },
      { id: 9, code: 'rin', name: 'Gold Ring', quality: 'rare', identified: true, w: 1, h: 1, canWear: ['lring', 'rring'], props: ['Fire Resist +11%', '+12 Life', '+3 Dexterity'], stashOp: 'stash', usable: { hero: side(true), merc: side(true) } },
    ],
    stash: [{ id: 20, code: 'amu', name: 'Topaz Amulet', quality: 'magic', identified: true, w: 1, h: 1, stashOp: 'take', props: ['+9 Mana'] }],
    belt: [{ id: 30, code: 'hp1', name: 'Minor Healing Potion', quality: 'normal', identified: true, w: 1, h: 1 }],
  };
  return v;
}
const MENUS = { pack: 'item-1', packunmet: 'item-3', unid: 'item-6', stash: 'item-20', worn: 'slot-rhand', wornbroken: 'slot-torso', belt: 'item-30', mercok: 'item-2' };

// ---- the tooltip
function tipPage(src, vw, vh) {
  const anchors = {}, docL = {}, winL = {}, boxes = {}, body = mk('body'), gone = {};
  body.dataset.canvas = '';
  const getA = (key, rect) => {
    if (!anchors[key]) {
      const a = mk('button'); a.dataset.key = key; a.attrs.title = 'native title'; a.closest = () => a;
      a.getBoundingClientRect = () => rect || { left: 100, top: 100, right: 150, bottom: 140 };
      anchors[key] = a;
    }
    return anchors[key];
  };
  const document = { body, documentElement: { clientWidth: vw || 1000, clientHeight: vh || 700 }, createElement: mk, getElementById: (i) => (boxes[i] = boxes[i] || mk('div')),
    querySelector: (s) => { const m = /data-key="([^"]+)"/.exec(s); return m ? (gone[m[1]] ? null : getA(m[1])) : null; }, addEventListener: (t, f) => { (docL[t] = docL[t] || []).push(f); } };
  let h = null;
  const window = { addEventListener: (t, f) => { (winL[t] = winL[t] || []).push(f); }, __tipTest: (x) => { h = x; } };
  vm.runInNewContext(src, { document, window, JSON, Math, Object, Array, String, Number, RegExp, console });
  return { h, document, body, getA };
}
const tipSrc = fs.readFileSync(process.argv[2], 'utf8'), actSrc = fs.readFileSync(process.argv[3], 'utf8');
for (const cls of ['Sorceress', 'Paladin']) {
  const v = view(cls);
  const tips = { noreq: 'item-1', level: 'item-2', str: 'item-3', dex: 'item-4', classitem: 'item-5', unid: 'item-6', broken: 'item-7', cmp: 'item-8', rings: 'item-9', worn: 'slot-torso', stash: 'item-20' };
  for (const [name, key] of Object.entries(tips)) {
    const p = tipPage(tipSrc); p.h.init(); p.h.onView(v); p.h.show(key, 'hover');
    check('tip_' + cls.toLowerCase() + '_' + name, ser(p.h.state().node, 0));
  }
  // a queued move for town shows the TOWN badge and its wait
  const q = JSON.parse(JSON.stringify(v)); q.queue = [{ id: 20, kind: 'take', state: 'queued', wait: 'town', name: 'Topaz Amulet' }];
  const p = tipPage(tipSrc); p.h.init(); p.h.onView(q); p.h.show('item-20', 'hover');
  check('tip_' + cls.toLowerCase() + '_town', ser(p.h.state().node, 0));
}

// ---- a refused requirement never renders as met (every item, every side)
{
  const v = view('Paladin'), p = tipPage(tipSrc); p.h.init(); p.h.onView(v);
  const all = v.backpack.concat(v.stash, v.belt);
  let seen = 0;
  all.forEach((it) => {
    const f = { it, slot: '', place: 'pack' }, r = p.h.richOf(it, f);
    r.who.forEach((w) => {
      const side = it.usable[w.who];
      w.rows.forEach((row) => { if (row.ok) assert.ok(!/needs|only|not identified|broken/.test(row.text), 'a met row words a refusal: ' + row.text); });
      const hard = (side.reasons || []).filter((x) => !x.warn);
      if (hard.length) { seen++; assert.ok(w.rows.some((row) => !row.ok && !row.warn), it.name + ': a refused item has no unmet row'); assert.strictEqual(w.ok, false); }
      hard.forEach((x) => { assert.ok(!w.rows.some((row) => row.ok && row.text === x.text), 'refusal shown as met: ' + x.text); });
    });
  });
  assert.ok(seen >= 6, 'the fixture should refuse several items');
  // the gap is the server's numbers: level 9 against 7 is 2 short, strength 41 against 38 is 3
  const lv = p.h.reqRows(v.backpack[1], v.backpack[1].usable.hero);
  assert.ok(lv.some((x) => !x.ok && /Level 9 \(7\)\s+2 short/.test(x.text)), JSON.stringify(lv));
  const st = p.h.reqRows(v.backpack[2], v.backpack[2].usable.hero);
  assert.ok(st.some((x) => !x.ok && /Strength 41 \(38\)\s+3 short/.test(x.text)), JSON.stringify(st));
  // no requirement and nothing in the way: one met row
  assert.strictEqual(JSON.stringify(p.h.reqRows(v.backpack[0], v.backpack[0].usable.hero)), JSON.stringify([{ ok: true, text: 'no requirements' }]));
  // an item with no verdict (a potion) has no requirement block
  assert.strictEqual(p.h.richOf(v.belt[0], { it: v.belt[0], slot: '', place: 'belt' }).who.length, 0);
  // the mercenary block is there only while one is employed
  const nm = JSON.parse(JSON.stringify(v)); nm.merc = { state: 'none', owned: false };
  const p2 = tipPage(tipSrc); p2.h.init(); p2.h.onView(nm);
  assert.strictEqual(JSON.stringify(p2.h.richOf(nm.backpack[0], { it: nm.backpack[0], slot: '', place: 'pack' }).who.map((w) => w.who)), JSON.stringify(['hero']));
}

// ---- the comparison: old to new with signed deltas, both rings, both hands
{
  const v = view('Sorceress'), p = tipPage(tipSrc); p.h.init(); p.h.onView(v);
  const c = p.h.compare(v.backpack[7], v.equipment.rhand);
  assert.strictEqual(JSON.stringify(c.map((r) => [r.label, r.old, r.now, r.delta, r.tone])), JSON.stringify([['Damage', '1-5', '2-8', '+2', 'up'], ['Charged Bolt', '-', '+2', '+2', 'up']]));
  const lr = p.h.compare(v.backpack[8], v.equipment.lring).map((r) => r.label + ' ' + r.delta);
  assert.strictEqual(JSON.stringify(lr), JSON.stringify(['Fire Resist +3', 'Life +6', 'Dexterity +3']));
  const rr = p.h.compare(v.backpack[8], v.equipment.rring).map((r) => r.label + ' ' + r.delta);
  assert.ok(rr.includes('Life +2') && rr.includes('Fire Resist +11'), rr.join('|'));
  // a value only the worn item has goes to nothing and counts as a loss
  const loss = p.h.compare(v.backpack[0], v.equipment.lring);
  assert.ok(loss.every((r) => r.tone === 'dn' && r.now === '-'), JSON.stringify(loss));
  // both ring slots show, one block each; an empty hand compares with nothing
  const rich = p.h.richOf(v.backpack[8], { it: v.backpack[8], slot: '', place: 'pack' });
  assert.strictEqual(JSON.stringify(rich.against.map((a) => a.slot)), JSON.stringify(['lring', 'rring']));
  const e = JSON.parse(JSON.stringify(v)); delete e.equipment.rhand;
  const p3 = tipPage(tipSrc); p3.h.init(); p3.h.onView(e);
  assert.strictEqual(JSON.stringify(p3.h.richOf(e.backpack[7], { it: e.backpack[7], slot: '', place: 'pack' }).against), JSON.stringify([]));
  // the tooltip of a worn item compares with nothing; no more than six rows
  assert.strictEqual(p.h.richOf(v.equipment.rhand, { it: v.equipment.rhand, slot: 'rhand', place: 'worn' }).against.length, 0);
  const many = { name: 'X', quality: 'magic', identified: true, props: ['+1 A', '+1 B', '+1 C', '+1 D', '+1 E', '+1 F', '+1 G', '+1 H'] };
  assert.strictEqual(p.h.compare(many, { name: 'Y', props: [] }).length, 6);
}

// ---- placement: inside the window and off the cell at the eight sizes, for every cell of a grid
{
  const p = tipPage(tipSrc), pl = p.h.place;
  const sizes = [[1440, 900], [1920, 1080], [1366, 768], [1914, 930], [1904, 1000], [1536, 730], [1280, 620], [390, 844]];
  let n = 0;
  for (const [vw, vh] of sizes) {
    for (const th of [120, 260, 420, 560]) {
      for (let x = 0; x + 58 <= vw; x += 29) {
        for (let y = 0; y + 58 * 4 <= vh; y += 41) {
          for (const [cw, ch] of [[58, 58], [58, 116], [116, 174], [116, 232]]) {
            if (x + cw > vw || y + ch > vh) continue;
            const a = { left: x, top: y, right: x + cw, bottom: y + ch };
            if (vw <= 640) { // a phone: the sheet takes the edge away from the cell
              const edge = p.h.sheetEdge(a, vh);
              assert.ok(edge === 'top' || edge === 'bottom');
              if (ch < vh * 0.3) assert.ok(edge === 'top' ? a.top >= vh / 2 - ch : a.bottom <= vh / 2 + ch);
              continue;
            }
            const o = pl(a, 270, th, vw, vh), h = Math.min(th, vh - 16);
            assert.ok(o.x >= 8 && o.y >= 8 && o.x + 270 <= vw - 8 && o.y + h <= vh - 8, 'outside ' + JSON.stringify([vw, vh, a, o]));
            const overlap = !(o.x + 270 <= a.left || o.x >= a.right || o.y + h <= a.top || o.y >= a.bottom);
            assert.ok(!overlap, 'covers its cell ' + JSON.stringify([vw, vh, th, a, o]));
            n++;
          }
        }
      }
    }
  }
  assert.ok(n > 5000, 'placements checked: ' + n);
}

// ---- the menu and the gestures
function actPage(src, v) {
  const anchors = {}, docL = {}, winL = {}, boxes = {}, fetched = [], body = mk('body'), prevented = [], timers = [];
  body.dataset.canvas = '';
  const getA = (key) => {
    if (!anchors[key]) { const a = mk('button'); a.dataset.key = key; a.closest = () => a; a.getBoundingClientRect = () => ({ left: 100, top: 100, right: 150, bottom: 140 }); anchors[key] = a; }
    return anchors[key];
  };
  const document = { body, documentElement: { clientWidth: 1000, clientHeight: 700 }, createElement: mk, getElementById: (i) => (boxes[i] = boxes[i] || mk('div')),
    querySelector: (s) => { const m = /data-key="([^"]+)"/.exec(s); return m ? getA(m[1]) : null; }, querySelectorAll: () => Object.values(anchors), dispatchEvent: () => true,
    addEventListener: (t, f) => { (docL[t] = docL[t] || []).push(f); } };
  let h = null, reply = { status: 200, body: { ok: true, request: { state: 'queued', how: 'queue' } } }, current = v;
  const window = { addEventListener: (t, f) => { (winL[t] = winL[t] || []).push(f); }, __actTest: (x) => { h = x; } };
  const fetch = (u, o) => {
    fetched.push({ u, body: o && o.body });
    if (/config$/.test(u)) return Promise.resolve({ ok: true, json: () => Promise.resolve({ answerBase: 'http://x' }) });
    if (/inventory$/.test(u)) return Promise.resolve({ ok: true, json: () => Promise.resolve(current) });
    return Promise.resolve({ ok: reply.status < 300, status: reply.status, json: () => Promise.resolve(reply.body) });
  };
  vm.runInNewContext(src, { setTimeout: (f, ms) => { timers.push({ f, ms }); return timers.length; }, document, window, JSON, Math, Object, Array, String, Number, RegExp, Promise, console, fetch, location: { protocol: 'http:', hostname: 'h' }, CustomEvent: function (t, o) { this.type = t; this.detail = o && o.detail; } });
  const fire = (t, e) => (docL[t] || []).forEach((f) => f(e));
  return { h, window, document, body, getA, docL, boxes, fetched, timers, fire, setReply: (r) => { reply = r; }, setView: (x) => { current = x; }, prevented, ev: (target, extra) => Object.assign({ target, preventDefault() { prevented.push(1); }, stopPropagation() {}, clientX: 120, clientY: 130 }, extra) };
}
const flush = () => new Promise((r) => setTimeout(r, 0));
const post = (p) => p.fetched.filter((f) => !/config$|inventory$/.test(f.u));

(async () => {
  for (const cls of ['Sorceress', 'Paladin']) {
    for (const [name, key] of Object.entries(MENUS)) {
      const p = actPage(actSrc, view(cls)); p.h.init(); p.h.onView(view(cls)); p.h.open(key);
      check('menu_' + cls.toLowerCase() + '_' + name, ser(p.h.state().node, 0));
    }
  }
  // the mercenary entry is listed greyed with a reason: the verdict first, then the bot's own
  {
    const v = view('Paladin'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
    const e = (k, a) => p.h.entryFor(k, a).e;
    assert.match(e('item-1', 'merc').reason, /Desert Guard wears/);
    assert.match(e('item-2', 'merc').reason, /cannot change the mercenary/);
    assert.strictEqual(e('item-2', 'merc').key, 'M');
    // the first entry is what Enter and a double-click run; for a worn item it is Take off, never Equip
    assert.strictEqual(p.h.first('slot-rhand').act, 'unequip');
    assert.strictEqual(p.h.first('item-1').act, 'equip');
    assert.strictEqual(p.h.first('item-3').act, 'equip'); assert.match(p.h.first('item-3').reason, /strength/);
    assert.strictEqual(p.h.first('item-6').act, 'identify');
    assert.strictEqual(p.h.first('item-20').act, 'take');
    assert.strictEqual(p.h.first('item-30'), null, 'a potion has no double-click action');
    assert.ok(p.h.entryFor('slot-torso', 'repair'), 'a broken worn item offers repair');
    assert.ok(!p.h.entryFor('slot-rhand', 'repair'));
  }
  // a click only selects: it never opens the menu on the canvas; a right-click opens it at the pointer
  {
    const v = view('Sorceress'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
    const a = p.getA('item-1');
    p.fire('click', p.ev(a)); assert.strictEqual(p.h.state().key, '');
    p.fire('contextmenu', p.ev(a)); assert.strictEqual(p.h.state().key, 'item-1'); assert.strictEqual(p.prevented.length, 1);
    const eq = p.h.state().node.querySelectorAll('button.act').find((b) => b.dataset.act.startsWith('equip'));
    assert.ok(eq.children.some((c) => c.tag === 'kbd' && c.textContent === 'Enter'), 'every entry shows its key');
    assert.ok(/translate\(/.test(p.h.state().node.style.transform), 'placed by a transform');
    p.fire('keydown', { key: 'Escape', target: a }); assert.strictEqual(p.h.state().key, '');
    p.fire('keydown', { key: 'ContextMenu', target: a, preventDefault() {} }); assert.strictEqual(p.h.state().key, 'item-1', 'the menu key');
  }
  // double-click and Enter run the first entry; a refused first entry sends nothing and says why
  {
    const v = view('Sorceress'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v); await flush();
    p.setReply({ status: 200, body: { ok: true, request: { id: 1, state: 'done', how: 'fast' } } });
    p.fire('dblclick', p.ev(p.getA('item-1')));
    assert.ok('sending' in p.getA('item-1').dataset, 'dashed while the answer is awaited');
    assert.ok(!('result' in p.getA('item-1').dataset), 'the pending look is never the done look');
    await flush(); await flush(); await flush();
    const sent = post(p); assert.strictEqual(sent.length, 1); assert.match(sent[0].u, /\/equip$/);
    assert.deepStrictEqual(JSON.parse(sent[0].body), { id: 1, slot: 'lring' });
    assert.strictEqual(p.getA('item-1').dataset.result, 'done'); assert.ok(!('sending' in p.getA('item-1').dataset));
    assert.strictEqual(p.getA('slot-lring').dataset.result, 'done');
    // two more views and the mark is gone
    p.h.onView(v); p.h.onView(v); p.h.onView(v); assert.ok(!('result' in p.getA('item-1').dataset));
    // a refused first entry (strength): nothing sent, red at the item, the reason in the message line, never done
    const n = post(p).length;
    p.fire('dblclick', p.ev(p.getA('item-3')));
    assert.strictEqual(post(p).length, n, 'a refused action sends nothing');
    assert.strictEqual(p.getA('item-3').dataset.result, 'refused'); assert.match(p.getA('item-3').attrs['data-why'], /strength/);
    assert.match(p.boxes.msg.textContent, /strength/);
    // Enter does the same as a double-click
    p.fire('keydown', { key: 'Enter', target: p.getA('item-6'), preventDefault() {} });
    await flush(); await flush(); await flush();
    assert.match(post(p).pop().u, /\/identify$/);
  }
  // the server refuses (409 with the verb's code): the item stays, the cell is red, the reason and the code are shown
  {
    const v = view('Paladin'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
    p.setReply({ status: 409, body: { code: 'cursor', error: 'something is held on the cursor' } });
    p.fire('dblclick', p.ev(p.getA('item-1')));
    await flush(); await flush(); await flush(); await flush();
    assert.strictEqual(p.getA('item-1').dataset.result, 'refused'); assert.ok(!('sending' in p.getA('item-1').dataset));
    assert.match(p.getA('item-1').attrs['data-why'], /cursor/);
    assert.match(p.boxes.msg.textContent, /cursor/);
    assert.ok(!Object.values(p.h.fb()).some((x) => x.state === 'done'));
    // the same request answered failed inside a 200 (the lane's own record) is refused too
    p.setReply({ status: 200, body: { ok: true, request: { id: 1, state: 'failed', code: 'verify', reason: 'the game did not show the change' } } });
    p.fire('dblclick', p.ev(p.getA('item-9')));
    await flush(); await flush(); await flush(); await flush();
    assert.strictEqual(p.getA('item-9').dataset.result, 'refused');
  }
  // queued (the lane is off, or the hero is in the field): pending from the view, with a TOWN badge when it waits for town
  {
    const v = view('Sorceress'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
    p.setReply({ status: 200, body: { ok: true, request: { id: 20, kind: 'take', state: 'queued', how: 'queue', wait: 'town' } } });
    const q = JSON.parse(JSON.stringify(v)); q.queue = [{ id: 20, kind: 'take', state: 'queued', wait: 'town', name: 'Topaz Amulet' }];
    p.setView(q);
    p.fire('click', p.ev(p.getA('item-20'), { ctrlKey: true }));
    await flush(); await flush(); await flush(); await flush();
    assert.match(post(p)[0].u, /\/take$/);
    assert.ok(!('result' in p.getA('item-20').dataset), 'queued is not done');
    assert.strictEqual(p.getA('item-20').dataset.pending, 'take'); assert.strictEqual(p.getA('item-20').dataset.wait, 'town');
    // the menu of that item lists the queued move with Cancel
    p.h.open('item-20'); const names = p.h.state().node.children.map((c) => c.className); assert.ok(names.includes('act-state'));
  }
  // Ctrl-click moves between pack and stash, Alt-click marks for sale, Delete asks first
  {
    const v = view('Sorceress'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
    p.fire('click', p.ev(p.getA('item-1'), { ctrlKey: true })); await flush(); await flush();
    assert.match(post(p)[0].u, /\/stash$/);
    p.fire('click', p.ev(p.getA('item-20'), { ctrlKey: true })); await flush(); await flush();
    assert.match(post(p)[1].u, /\/take$/);
    p.fire('click', p.ev(p.getA('item-2'), { altKey: true })); await flush(); await flush();
    const mark = post(p)[2]; assert.match(mark.u, /\/mark$/); assert.deepStrictEqual(JSON.parse(mark.body), { id: 2, mark: 'sell' });
    const n = post(p).length;
    p.fire('keydown', { key: 'Delete', target: p.getA('item-3'), preventDefault() {} });
    assert.strictEqual(post(p).length, n, 'Delete only asks');
    assert.strictEqual(p.h.state().key, 'item-3'); assert.strictEqual(p.h.state().confirming, true, 'one inline confirmation');
    const yes = p.h.state().node.querySelectorAll('button.act').find((b) => 'confirm' in b.dataset);
    yes.listeners.click[0](); await flush(); await flush();
    assert.match(post(p).pop().u, /\/trash$/);
  }

  // ---- the 21 validity cases of I.10, as the server states them (usable.hero / usable.merc, req, noWear): the tooltip rows and the
  // menu entries must follow the verdict, and a refusal is never drawn as a met requirement or an offered action
  {
    const R = (code, text, extra) => Object.assign({ code, text }, extra || {});
    const rq = (stat, need, have, what) => R('requirements', 'needs ' + what + ' (you have ' + have + ')', { stat, need, have });
    const merc = { state: 'ready', owned: true, alive: true, kind: 'desert', level: 14, actWhy: 'the bot cannot change the mercenary\'s gear yet' };
    // [no, wearer, item name, req, hero reasons, merc reasons (null: no mercenary), item extras, expect]
    const C = [
      [1, 'Sorceress 7', 'Sabre', { str: 30 }, [], [], { canWear: ['rhand'] }, { equip: 'ok', hero: [/^✓ Strength 30$/] }],
      [2, 'Sorceress 7', 'War Staff', { str: 47 }, [rq('str', 47, 38, '47 strength')], [], { noWear: 'needs 47 strength (you have 38)' }, { equip: 'refused', hero: [/^✗ Strength 47 \(38\)\s+9 short$/] }],
      [3, 'Sorceress 7', 'Helm', { lvl: 12 }, [rq('lvl', 12, 7, 'level 12')], [], { noWear: 'needs level 12 (you have 7)' }, { equip: 'refused', hero: [/^✗ Level 12 \(7\)\s+5 short$/] }],
      [4, 'Sorceress', 'Paladin shield', { class: ['Paladin'] }, [R('class', 'only a Paladin can use it')], [], { noWear: 'Paladin only' }, { equip: 'refused', hero: [/^✗ only a Paladin can use it$/] }],
      [5, 'Paladin', 'Paladin shield', { class: ['Paladin'] }, [], [], { canWear: ['lhand'] }, { equip: 'ok', hero: [/^✓ Paladin only$/] }],
      [6, 'Paladin', 'Sorceress orb', { class: ['Sorceress'] }, [R('class', 'only a Sorceress can use it')], [], { noWear: 'Sorceress only' }, { equip: 'refused', hero: [/^✗ only a Sorceress can use it$/] }],
      [7, 'Amazon', 'Long Bow', { dex: 60 }, [rq('dex', 60, 40, '60 dexterity')], [], { noWear: 'needs 60 dexterity (you have 40)' }, { equip: 'refused', hero: [/^✗ Dexterity 60 \(40\)\s+20 short$/] }],
      [8, 'Barbarian', 'Primal Helm', { class: ['Barbarian'] }, [], [], { canWear: ['head'] }, { equip: 'ok', hero: [/^✓ Barbarian only$/] }],
      [8.5, 'Druid', 'Primal Helm', { class: ['Barbarian'] }, [R('class', 'only a Barbarian can use it')], [], { noWear: 'Barbarian only' }, { equip: 'refused', hero: [/^✗ only a Barbarian can use it$/] }],
      [9, 'Necromancer', 'Wand', { class: ['Necromancer'] }, [], [], { canWear: ['rhand'] }, { equip: 'ok', hero: [/^✓ Necromancer only$/] }],
      [9.5, 'Assassin', 'Claw', { class: ['Assassin'] }, [R('class', 'only an Assassin can use it')], [], { noWear: 'Assassin only' }, { equip: 'refused', hero: [/^✗ only an Assassin can use it$/] }],
      [10, 'any', 'Ring', {}, [R('identified', 'it is not identified')], [], { identified: false, noWear: 'it is not identified' }, { equip: 'refused', identify: 'ok', hero: [/^✗ it is not identified$/] }],
      [11, 'any', 'Broken Sword', {}, [R('broken', 'it is broken: repair it', { warn: true })], [], { canWear: ['rhand'], maxDur: 24, dur: 0 }, { equip: 'ok', hero: [/^! it is broken: repair it$/] }],
      [12, 'any', 'Great Axe', {}, [R('twohand', 'a two-handed weapon and a shield cannot be worn together; take the other one off first')], [], { noWear: 'a two-handed weapon and a shield cannot be worn together; take the other one off first' }, { equip: 'refused', hero: [/^✗ a two-handed weapon and a shield/] }],
      [12.5, 'any', 'Great Axe, pack full', {}, [R('nospace', 'there is no room in the pack for the shield')], [], { noWear: 'there is no room in the pack for the shield' }, { equip: 'refused', hero: [/^✗ there is no room/] }],
      [13, 'any', 'Ring, both fingers full', {}, [], [], { canWear: ['lring', 'rring'] }, { equip: 'ok', cmp: ['lring', 'rring'], hero: [/^✓ no requirements$/] }],
      [14, 'hero after a stat change', 'Worn Plate', { str: 60 }, [rq('str', 60, 55, '60 strength')], [], { noWear: 'needs 60 strength (you have 55)' }, { equip: 'refused', hero: [/^✗ Strength 60 \(55\)\s+5 short$/] }],
      [15, 'Desert Guard 14, Str 60', 'Spear', { str: 40 }, [], [], { canWear: ['rhand'] }, { equip: 'ok', merc: [/^✓ Strength 40$/], mercEntry: 'bot' }],
      [15.5, 'Desert Guard 14, Str 60', 'Polearm', { str: 80 }, [], [rq('str', 80, 60, '80 strength')], { canWear: ['rhand'] }, { equip: 'ok', merc: [/^✗ Strength 80 \(60\)\s+20 short$/], mercEntry: 'verdict' }],
      [15.6, 'Desert Guard 14', 'Bow', {}, [], [R('merctype', 'a Desert Guard wears helm, armor, spear or polearm')], { canWear: ['rhand'] }, { equip: 'ok', merc: [/^✗ a Desert Guard wears/], mercEntry: 'verdict' }],
      [19, 'any mercenary, dead', 'Helm', {}, [], [R('dead', 'he is dead')], { canWear: ['head'] }, { equip: 'ok', merc: [/^✗ he is dead$/], mercEntry: 'verdict' }],
      [20, 'no mercenary', 'Helm', {}, [], null, { canWear: ['head'] }, { equip: 'ok', noMerc: true }],
      [21, 'any', 'Ethereal Sword (U: which number the server enforces)', { str: 25, approx: true }, [rq('str', 25, 20, '25 strength')], [], { noWear: 'needs 25 strength (you have 20)' }, { equip: 'refused', hero: [/^✗ Strength 25 \(20\)\s+5 short$/] }],
    ];
    let n = 0;
    for (const cls of ['Sorceress', 'Paladin']) {
      C.forEach(([no, who, name, req, hr, mr, extra, exp]) => {
        const v = view(cls); v.merc = mr === null ? { state: 'none', owned: false } : merc;
        const it = Object.assign({ id: 700, code: 'x', name, quality: 'magic', identified: true, w: 1, h: 2, stashOp: 'stash', req, usable: { hero: { ok: !hr.some((r) => !r.warn), reasons: hr } } }, extra);
        if (mr !== null) it.usable.merc = { ok: !mr.some((r) => !r.warn), reasons: mr };
        v.backpack = [it];
        const tp = tipPage(tipSrc); tp.h.init(); tp.h.onView(v);
        const rich = tp.h.richOf(it, { it, slot: '', place: 'pack' }), where = no + ' ' + who + ' ' + name;
        const heroRows = (rich.who.find((w) => w.who === 'hero') || { rows: [] }).rows.map((r) => (r.ok ? '✓ ' : r.warn ? '! ' : '✗ ') + r.text);
        const mercRows = (rich.who.find((w) => w.who === 'merc') || { rows: [] }).rows.map((r) => (r.ok ? '✓ ' : r.warn ? '! ' : '✗ ') + r.text);
        (exp.hero || []).forEach((re) => assert.ok(heroRows.some((r) => re.test(r)), where + ': hero rows ' + JSON.stringify(heroRows)));
        (exp.merc || []).forEach((re) => assert.ok(mercRows.some((r) => re.test(r)), where + ': merc rows ' + JSON.stringify(mercRows)));
        // a refused hard reason is never a met row
        hr.filter((r) => !r.warn).forEach((r) => assert.ok(!heroRows.some((x) => x.startsWith('✓') && x.includes(r.text)), where + ': refusal shown as met'));
        if (exp.noMerc) assert.ok(!rich.who.some((w) => w.who === 'merc'), where + ': merc block without a mercenary');
        if (exp.cmp) assert.strictEqual(JSON.stringify(rich.against.map((a) => a.slot)), JSON.stringify(exp.cmp), where);
        const ap = actPage(actSrc, v); ap.h.init(); ap.h.onView(v);
        const eq = ap.h.entryFor('item-700', 'equip');
        if (exp.equip === 'ok') assert.ok(!eq.e.reason, where + ': equip must be offered'); else assert.ok(eq.e.reason, where + ': equip must be refused with a reason');
        if (exp.identify === 'ok') assert.ok(!ap.h.entryFor('item-700', 'identify').e.reason, where + ': identify offered');
        else assert.ok(!ap.h.entryFor('item-700', 'identify'), where + ': identify offered on an identified item');
        const me = ap.h.entryFor('item-700', 'merc');
        if (exp.noMerc) assert.ok(!me, where + ': merc entry without a mercenary');
        else if (exp.mercEntry === 'verdict') assert.ok(me.e.reason && me.e.reason === mr.filter((r) => !r.warn)[0].text, where + ': merc entry carries his verdict ' + (me && me.e.reason));
        else if (exp.mercEntry === 'bot') assert.match(me.e.reason, /cannot change the mercenary/, where);
        // a refused equip sent as a gesture never reaches the network
        if (exp.equip !== 'ok' && !exp.identify) { const before = post(ap).length; ap.fire('dblclick', ap.ev(ap.getA('item-700'))); assert.strictEqual(post(ap).length, before, where + ': a refused equip was sent'); assert.strictEqual(ap.getA('item-700').dataset.result, 'refused', where); }
        n++;
      });
    }
    assert.strictEqual(n, 2 * C.length);
  }

  // ---- every kind of action shows its pending, done, refused and queued state the same way (api calls, as the buttons and the drag make them)
  {
    const kinds = [['equip', (a) => a.equip(1, 'lring'), ['item-1', 'slot-lring']], ['unequip', (a) => a.unequip('rhand'), ['slot-rhand']], ['trash', (a) => a.trash(2), ['item-2']],
      ['identify', (a) => a.identify(6), ['item-6']], ['stash', (a) => a.stash(1), ['item-1']], ['take', (a) => a.take(20), ['item-20']], ['mark', (a) => a.post('/api/inventory/mark', { id: 2, mark: 'sell' }), ['item-2']]];
    for (const [kind, call, keys] of kinds) {
      const v = view('Paladin'), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
      const at = (st) => keys.forEach((k) => st(p.getA(k).dataset));
      // pending while the answer is awaited, never the done look
      p.setReply({ status: 200, body: { ok: true, request: { id: 1, state: 'done', how: 'fast' } } });
      let pr = call(p.window.d2InvActions); at((d) => { assert.ok('sending' in d, kind + ': dashed while sending'); assert.ok(!('result' in d), kind + ': pending looks done'); });
      await pr; at((d) => { assert.strictEqual(d.result, 'done', kind); assert.ok(!('sending' in d)); });
      // refused: the server's reason and code, red, nothing marked done
      p.setReply({ status: 409, body: { code: 'cursor', error: 'something is held on the cursor' } });
      pr = call(p.window.d2InvActions); await pr; at((d) => assert.strictEqual(d.result, 'refused', kind));
      assert.match(p.boxes.msg.textContent, /cursor/, kind);
      // queued: neither done nor refused; the view's queue entry supplies the dashed outline
      p.setReply({ status: 200, body: { ok: true, request: { id: 1, state: 'queued', how: 'queue', wait: 'safe' } } });
      p.h.onView(v); p.h.onView(v); p.h.onView(v);
      pr = call(p.window.d2InvActions); await pr; at((d) => { assert.ok(!('result' in d), kind + ': queued shown as a result'); });
    }
  }
  // ---- idfast: Identify feels instant. Optimistic state, the fast poll, the 2 s highlight, the revert with the reason, the wait reason
  for (const cls of ['Sorceress', 'Paladin']) {
    const v = view(cls), p = actPage(actSrc, v); p.h.init(); p.h.onView(v);
    const cell = p.getA('item-6');
    // the click shows "Identifying..." in the same tick, before any answer; the button of the menu is not offered twice
    p.fire('keydown', { key: 'Enter', target: cell, preventDefault() {} });
    assert.ok('sending' in cell.dataset, 'optimistic state at once');
    assert.strictEqual(cell.dataset.doing, 'Identifying...');
    // success: the read-back view is what ends it; the highlight is done for 2 s by the clock
    p.setReply({ status: 200, body: { ok: true, request: { id: 6, kind: 'identify', state: 'done', how: 'fast' } } });
    await flush(); await flush(); await flush(); await flush();
    assert.strictEqual(cell.dataset.result, 'done'); assert.ok(!('sending' in cell.dataset) && !('doing' in cell.dataset));
    const exp = p.timers.filter((t) => t.ms === 2000); assert.strictEqual(exp.length, 1, 'one 2 s expiry for the highlight');
    exp[0].f(); assert.ok(!('result' in cell.dataset), 'highlight gone after 2 s');
    // refused by the server: reverts with the reason, never shown as done
    p.setReply({ status: 409, body: { code: 'requirements', error: 'nothing to read it with' } });
    p.fire('keydown', { key: 'Enter', target: cell, preventDefault() {} });
    await flush(); await flush(); await flush(); await flush();
    assert.strictEqual(cell.dataset.result, 'refused'); assert.match(cell.attrs['data-why'], /nothing/);
    assert.match(p.boxes.msg.textContent, /nothing/);
    // queued: the item says what it waits for; the poll runs at 200 ms while a request is pending and stops when none is
    const q = JSON.parse(JSON.stringify(v)); q.queue = [{ id: 6, kind: 'identify', state: 'queued', wait: 'safe', name: 'Ring' }];
    p.timers.length = 0; p.h.onView(q);
    assert.strictEqual(cell.dataset.waiting, 'waiting: a safe moment (no monster near)');
    assert.strictEqual(p.timers.filter((t) => t.ms === 200).length, 1, 'fast poll scheduled');
    q.instant = { enabled: true, why: 'a panel is open' }; p.h.onView(q);
    assert.strictEqual(cell.dataset.waiting, 'waiting: a panel is open');
    const q2 = JSON.parse(JSON.stringify(v)); q2.queue = []; p.timers.length = 0; p.h.onView(q2);
    assert.strictEqual(p.timers.filter((t) => t.ms === 200).length, 0, 'no poll with nothing pending');
    assert.ok(!('waiting' in cell.dataset));
    // a request that failed on its own (the item was sold) is said once, not dropped
    const lost = JSON.parse(JSON.stringify(v)); lost.recent = [{ id: 6, kind: 'identify', state: 'failed', code: 'lost', at: 5, reason: 'Identify request for Hand Axe lost: the item was sold, dropped or moved' }];
    p.boxes.msg.textContent = ''; p.h.onView(lost);
    assert.match(p.boxes.msg.textContent, /Hand Axe lost/);
    // the indicator of the lane
    const on = JSON.parse(JSON.stringify(v)); on.identify = { tome: 3, tomeCap: 20, scrolls: 0, unid: 2, buyDefault: 5, buyMax: 20, cain: 'open' }; on.instant = { enabled: true };
    p.h.onView(on); await flush(); await flush();
    const er = p.h.errands(); assert.strictEqual(er.instant.hidden, false); assert.strictEqual(er.instant.textContent, 'Instant actions: on');
    on.instant = { enabled: false }; p.h.onView(on); await flush(); await flush(); assert.strictEqual(p.h.errands().instant.hidden, true);
    // Identify all: progress from the view's own count, waiting text before it starts
    const c1 = JSON.parse(JSON.stringify(on)); c1.instant = { enabled: true }; c1.queue = [{ id: 0, kind: 'cain', state: 'doing' }]; p.h.onView(on); p.h.onView(c1); await flush(); await flush();
    const c2 = JSON.parse(JSON.stringify(c1)); c2.identify.unid = 1; p.h.onView(c2); await flush(); await flush();
    assert.strictEqual(p.h.errands().hint.textContent, 'Identified 1 of 2');
    // no tome: the click is refused at once (nothing sent) and the Buy button is offered
    const nt = JSON.parse(JSON.stringify(v)); nt.backpack.forEach((it) => { if (it.id === 6) it.noIdentify = 'no scroll or tome: buy identify scrolls'; }); nt.identify = on.identify; p.h.onView(nt); await flush();
    const sent = post(p).length; p.fire('keydown', { key: 'Enter', target: cell, preventDefault() {} });
    assert.strictEqual(post(p).length, sent); assert.ok('offer' in p.h.errands().buy.dataset, 'Buy identify scrolls offered');
  }
  // the cards page keeps its click menu
  {
    const v = view('Sorceress'), p = actPage(actSrc, v); delete p.body.dataset.canvas; p.h.init(); p.h.onView(v);
    p.fire('click', p.ev(p.getA('item-1'))); assert.strictEqual(p.h.state().key, 'item-1');
  }
  if (bad) process.exit(1);
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
