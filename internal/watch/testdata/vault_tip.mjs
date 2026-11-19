// The Vault page's tooltip is the inventory tooltip (inventory-tip.js), fed the vault rows by toTipView. This test draws a weapon
// and a jewel, identified and unidentified, from vault rows and from the same items as inventory rows: the text must be equal
// (one code path), an unidentified row shows the base facts, "Unidentified" and the item level and NO property line, and
// the tap and hover show the node. Usage: node vault_tip.mjs <vault-logic.js copy as .mjs> <inventory-tip.js>
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import { pathToFileURL } from 'node:url';

const L = await import(pathToFileURL(process.argv[2]).href);
let h = null;
const window = { __tipTest: (x) => { h = x; }, addEventListener() {} };
vm.runInNewContext(fs.readFileSync(process.argv[3], 'utf8'), { window, document: {}, Math, Object, Array, String, Number, RegExp, JSON });

const sword = { name: 'Phase Blade', code: 'zvb', quality: 'magic', q: 4, ilvl: 40, level: 40, identified: true, ethereal: false, sockets: 0, state: 'recorded', status: 'revealed', value: 100,
  base: { name: 'Phase Blade', minDam: 5, maxDam: 21, reqStr: 25, reqDex: 136, speed: -30, dur: 24 }, props: ['+20 to Strength', '10% Increased Attack Speed'], sellPrice: 240, art: 'invzvb' };
const jewel = { name: 'Jewel', code: 'jew', quality: 'magic', q: 4, ilvl: 30, level: 30, identified: true, state: 'recorded', status: 'revealed', value: 10,
  base: { name: 'Jewel' }, props: ['+8% Enhanced Damage'], sellPrice: 50, kind: 'jewel' };
const unid = (r) => Object.assign({}, r, { identified: false, props: undefined, status: 'recorded' });
const text = (it) => { const b = h.build(it, ''); return [b.title].concat(b.lines.map((l) => l.k + ':' + l.t)); };
const asInv = (r) => { const c = Object.assign({}, r, { sell: r.sellPrice }); return c; };

let n = 0;
for (const [name, r] of [['weapon', sword], ['jewel', jewel]]) {
  for (const row of [r, unid(r)]) {
    const tv = L.toTipView([Object.assign({ id: 'v-' + name }, row)]).view.backpack[0];
    assert.deepEqual(text(tv), text(asInv(row)), name + ': vault tooltip differs from the inventory one');
    n++;
  }
  const idl = text(L.toTipView([Object.assign({ id: 'a' }, r)]).view.backpack[0]).join('\n');
  assert.ok(/prop:\+\d+/.test(idl), name + ' identified shows its stat lines');
  assert.ok(!/Unidentified/.test(idl));
  const ul = text(L.toTipView([Object.assign({ id: 'b' }, unid(r))]).view.backpack[0]).join('\n');
  assert.ok(/Unidentified/.test(ul), name + ' unidentified says so');
  assert.ok(/Item level: \d+/.test(ul));
  assert.ok(!/prop:/.test(ul), name + ' unidentified shows no property line: ' + ul);
  assert.ok(!/Enhanced|Strength|Attack Speed/.test(ul));
}
const sw = text(L.toTipView([Object.assign({ id: 'c' }, unid(sword))]).view.backpack[0]).join('\n');
assert.ok(/One-Hand damage: 5 to 21/.test(sw) && /Required dexterity: 136/.test(sw), 'base stats of the unidentified weapon: ' + sw);
// the sell price of the tooltip is the vendor price, not the Sell errand's state word
assert.equal(L.toTipView([Object.assign({ id: 'd', sell: 'queued' }, sword)]).view.backpack[0].sell, 240);
assert.equal(L.toTipView([Object.assign({ id: 'd', sell: 'queued' }, { name: 'x' })]).view.backpack[0].sell, 0);
// the art key is the inventory rule
assert.equal(L.artKey({ art: 'invzvb' }), 'item-invzvb'); assert.equal(L.artKey({ art: 'invzvb', tint: 't1c2' }), 'item-invzvb~t1c2');
assert.equal(L.artKey({ art: '../x' }), ''); assert.equal(L.artKey({}), '');
console.log('ok', n);
