// uic3: the hero sheet tiles of web/inventory-canvas.js read stats.attrs, stats.defense and stats.damage of /api/hero; a value
// the backend did not send has no tile (never a zero). Usage: node sheet.js path/to/inventory-canvas.js
const fs = require('fs'), vm = require('vm'), assert = require('assert');
const { mk } = require('../charscreen/stub.js');
const src = fs.readFileSync(process.argv[2], 'utf8');
let h = null;
const document = { getElementById: (id) => Object.assign(mk('div'), { id }), createElement: mk, addEventListener() {}, querySelector: () => null, querySelectorAll: () => [] };
vm.runInNewContext(src, { document, window: { __canvasTest: (x) => { h = x; }, d2InvActions: {} }, JSON, Math, Object, Array, String, Number, RegExp, Date, console, Promise, fetch: () => new Promise(() => {}) });
const base = { level: 7, life: 86, maxLife: 120, mana: 105, maxMana: 140, xp: 900, xpNext: 1400, xpPct: 62.4 };
const labels = (hero) => h.heroTiles({ equipment: {} }, hero).map((t) => t.l + '=' + t.v);
const full = Object.assign({}, base, { stats: { resists: { fire: 12, cold: 30, light: -8, poison: 5 }, attrs: { str: 38, dex: 25, vit: 30, energy: 22 }, defense: 214, damage: { min: 4, max: 9 } } });
assert.strictEqual(labels(full).join('|'), ['Level=7', 'Strength=38', 'Dexterity=25', 'Vitality=30', 'Energy=22', 'Life=86/120', 'Mana=105/140', 'Defense=214', 'Bonus damage=4-9', 'Fire res=12%', 'Cold res=30%', 'Light res=-8%', 'Poison res=5%', 'xp this level=62%', 'xp left=500', 'Gear=ok'].join('|'));
// nothing read beyond the resistances: no attribute, defence or damage tile
const part = Object.assign({}, base, { stats: { resists: { fire: 1, cold: 2, light: 3, poison: 4 } } });
assert.ok(!labels(part).some((l) => /^(Strength|Dexterity|Vitality|Energy|Defense|Damage)=/.test(l)), labels(part).join('|'));
// a defence of 0 that WAS read is a value; an absent one is not
assert.ok(labels(Object.assign({}, base, { stats: { resists: { fire: 0, cold: 0, light: 0, poison: 0 }, defense: 0 } })).includes('Defense=0'));
// no /api/hero yet: the sheet draws what the view knows (gear) and nothing else
assert.strictEqual(labels(null).join('|'), 'Gear=ok');
console.log('ok');
