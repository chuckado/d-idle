// DOM goldens of web/inventory-canvas.js (card uic2): paints fixture views into a stub page and compares the serialised
// result with testdata/charscreen/canvas_<case>.txt (UPDATE=1 rewrites them). Cases: the hero alone (no mercenary), the four
// views of C1 (testdata/itemusable: Sorceress and Paladin, with and without a Desert Guard) for the requirement marks, one
// cell per reason code, a selected pack item with its action buttons, and the filters. Usage: node canvas_dom.js inventory-canvas.js
const fs = require('fs'), vm = require('vm'), path = require('path');
const { mk, ser } = require('./stub.js');
const dir = __dirname, data = path.join(dir, '..');
const src = fs.readFileSync(process.argv[2], 'utf8');

function page(view) {
  const ids = {};
  const get = (id) => ids[id] || (ids[id] = Object.assign(mk('div'), { id }));
  ['pack', 'stash', 'belt', 'equip', 'cv-gold', 'cv-stashgold', 'cv-sell', 'cv-repair', 'cv-sel', 'cv-sets', 'hsheet', 'equipmeta', 'gearprompt', 'cv', 'queue', 'cv-log'].forEach(get);
  ids.cv.dataset.tab = 'hero';
  (view.backpack || []).forEach((it) => { const n = mk('button'); n.dataset.key = 'item-' + it.id; n.title = it.name; ids.pack.appendChild(n); });
  (view.stash || []).forEach((it) => { const n = mk('button'); n.dataset.key = 'item-' + it.id; n.title = it.name; ids.stash.appendChild(n); });
  Object.keys(view.equipment || {}).forEach((s) => { const n = mk('button'); n.className = 'd-slot2'; n.dataset.slot = s; n.dataset.key = 'slot-' + s; ids.equip.appendChild(n); });
  const document = { getElementById: get, createElement: mk, addEventListener() {}, querySelector: () => null, querySelectorAll: () => [], createTextNode: (t) => ({ textContent: t }) };
  let h = null;
  const window = { __canvasTest: (x) => { h = x; }, d2InvActions: { say() {}, post() {}, equip() {}, unequip() {}, stash() {}, take() {}, identify() {}, trash() {}, open() {} } };
  const ctx = { document, window, JSON, Math, Object, Array, String, Number, RegExp, Date, console, Promise, fetch: () => new Promise(() => {}) };
  vm.runInNewContext(src, ctx);
  return { h, ids };
}
let bad = 0;
function check(name, got) {
  const file = path.join(dir, 'canvas_' + name + '.txt');
  if (process.env.UPDATE) { fs.writeFileSync(file, got); return; }
  if (!fs.existsSync(file)) { console.error('missing golden ' + file); bad++; return; }
  if (fs.readFileSync(file, 'utf8') !== got) { console.error('canvas differs from ' + file + '\n' + got); bad++; }
}
const dump = (p, keys) => keys.map((k) => '## ' + k + '\n' + ser(p.ids[k], 0)).join('');
const heroJSON = { level: 7, life: 86, maxLife: 120, mana: 105, maxMana: 140, xp: 900, xpNext: 1400, xpPct: 62.4, stats: { resists: { fire: 12, cold: 30, light: -8, poison: 5 } } };
const flat = (p) => ['pack', 'stash', 'belt'].map((k) => p.ids[k].children.map((n) => k + ' ' + n.dataset.key + ' usable=' + (n.dataset.usable || '') + ' who=' + (n.dataset.who || '') +
  ' badge=' + n.children.map((b) => b.textContent + '/' + b.dataset.tone + '/' + b.title).join('|') + ' title=' + n.title).join('\n')).join('\n') + '\n';

// the hero alone: no mercenary block, no sets, nothing selected by the player (the first pack item is)
{
  const v = JSON.parse(fs.readFileSync(path.join(data, 'itemusable', 'view_Sorceress_nomerc.json'), 'utf8'));
  v.gold = 1240; v.stashGold = 5300; v.sellPending = 2; v.class = 'Sorceress';
  const p = page(v); p.h.setView(v); p.h.setHero(heroJSON); p.h.setRepair({ estimate: 180, request: null, trip: null, worst: 'Short Staff is broken.' }); p.h.paint(v);
  check('hero_only', JSON.stringify(p.h.heroTiles(v, heroJSON)) + '\n' + dump(p, ['cv-gold', 'cv-stashgold', 'cv-sell', 'cv-repair', 'cv-sets', 'cv-sel']) + flat(p));
}
// the four views of C1: the marks only read `usable`
for (const n of ['Sorceress_nomerc', 'Sorceress_desert', 'Paladin_nomerc', 'Paladin_desert']) {
  const v = JSON.parse(fs.readFileSync(path.join(data, 'itemusable', 'view_' + n + '.json'), 'utf8'));
  const p = page(v); p.h.setView(v); p.h.paint(v);
  check('usable_' + n, flat(p));
}
// one cell per reason code, hero and mercenary verdicts apart
{
  const R = (code, stat, text, warn) => ({ code, stat, text, warn: !!warn });
  const no = (r) => ({ ok: false, reasons: [r] }), ok = { ok: true };
  const cases = [
    ['usable by both', ok, ok], ['hero only', ok, no(R('merctype', '', 'a Desert Guard wears helm, armor, spear or polearm'))], ['merc only (level)', no(R('requirements', 'lvl', 'needs level 9 (you are 7)')), ok],
    ['nobody (str)', no(R('requirements', 'str', 'needs 41 strength (you have 38)')), no(R('requirements', 'str', 'needs 41 strength (he has 40)'))],
    ['nobody (dex)', no(R('requirements', 'dex', 'needs 45 dexterity (you have 25)')), no(R('requirements', 'dex', 'needs 45 dexterity (he has 40)'))],
    ['class', no(R('class', '', 'Paladin only')), no(R('class', '', 'Paladin only'))], ['unidentified', no(R('identified', '', 'not identified')), no(R('identified', '', 'not identified'))],
    ['two-hander', no(R('twohand', '', 'the shield is worn')), ok], ['slot', no(R('slot', '', 'not equipment for this class')), ok], ['merc dead', ok, no(R('dead', '', 'he is dead'))],
    ['broken (warn only)', { ok: true, reasons: [R('broken', '', 'broken', true)] }, ok], ['hero only, no merc', ok, null],
  ];
  const v = { schema: 8, class: 'Sorceress', gold: 0, equipment: {}, belt: [], stash: [], backpack: cases.map((c, i) => ({ id: i + 1, name: c[0], quality: 'normal', identified: true, usable: { hero: c[1], merc: c[2] || undefined } })) };
  const p = page(v); p.h.setView(v); p.h.paint(v);
  check('codes', flat(p));
}
// the selected item and its buttons: a wearable pack item (two ring slots), a refused one, an unidentified one, a stash item, a worn item
{
  const v = JSON.parse(fs.readFileSync(path.join(data, 'itemusable', 'view_Sorceress_desert.json'), 'utf8'));
  v.backpack[1].canWear = ['lring', 'rring'];
  v.backpack[2].identified = false; v.backpack[2].noIdentify = '';
  v.stash = [{ id: 90, name: 'Ring', quality: 'rare', identified: true, stashOp: 'take', noStash: 'the stash opens only in town', x: 0, y: 0, w: 1, h: 1 }];
  const p = page(v); p.h.setView(v);
  const out = [];
  for (const k of ['item-' + v.backpack[0].id, 'item-' + v.backpack[1].id, 'item-' + v.backpack[2].id, 'item-90', 'slot-' + Object.keys(v.equipment)[0]]) {
    p.h.select(k); p.ids['cv-sel'].dataset.sig = ''; while (p.ids['cv-sel'].children.length) p.ids['cv-sel'].removeChild(p.ids['cv-sel'].lastChild);
    p.h.drawSel(); out.push('## ' + k + '\n' + ser(p.ids['cv-sel'], 0));
  }
  check('selected', out.join(''));
}
// filters dim and never hide
{
  const v = JSON.parse(fs.readFileSync(path.join(data, 'itemusable', 'view_Sorceress_desert.json'), 'utf8'));
  const lines = [];
  for (const [k, x] of [['usable', true], ['merc', true], ['unid', true], ['sell', true], ['q', 'staff']]) {
    const p = page(v); p.h.setFilter(k, x);
    lines.push(k + ': ' + v.backpack.map((it) => it.id + '=' + (p.h.matches(it) ? 'show' : 'dim')).join(' '));
  }
  check('filters', lines.join('\n') + '\n');
}
if (bad) process.exit(1);
console.log('ok');
