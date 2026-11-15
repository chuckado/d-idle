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
    stash: [{ id: 20, code: 'amu', name: 'Jade Amulet', quality: 'magic', identified: true, w: 1, h: 1, stashOp: 'take', props: ['+9 Mana'] }],
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
  const q = JSON.parse(JSON.stringify(v)); q.queue = [{ id: 20, kind: 'take', state: 'queued', wait: 'town', name: 'Jade Amulet' }];
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
  const anchors = {}, docL = {}, winL = {}, boxes = {}, fetched = [], body = mk('body'), prevented = [];
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
  vm.runInNewContext(src, { document, window, JSON, Math, Object, Array, String, Number, RegExp, Promise, console, fetch, location: { protocol: 'http:', hostname: 'h' }, CustomEvent: function (t, o) { this.type = t; this.detail = o && o.detail; } });
  const fire = (t, e) => (docL[t] || []).forEach((f) => f(e));
  return { h, document, body, getA, docL, boxes, fetched, fire, setReply: (r) => { reply = r; }, setView: (x) => { current = x; }, prevented, ev: (target, extra) => Object.assign({ target, preventDefault() { prevented.push(1); }, stopPropagation() {}, clientX: 120, clientY: 130 }, extra) };
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
    const q = JSON.parse(JSON.stringify(v)); q.queue = [{ id: 20, kind: 'take', state: 'queued', wait: 'town', name: 'Jade Amulet' }];
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
  // the cards page keeps its click menu
  {
    const v = view('Sorceress'), p = actPage(actSrc, v); delete p.body.dataset.canvas; p.h.init(); p.h.onView(v);
    p.fire('click', p.ev(p.getA('item-1'))); assert.strictEqual(p.h.state().key, 'item-1');
  }
  if (bad) process.exit(1);
  console.log('ok');
})().catch((e) => { console.error(e); process.exit(1); });
