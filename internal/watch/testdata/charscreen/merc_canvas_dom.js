// DOM goldens of the character canvas mode of web/inventory-merc.js (card uic2): the host carries data-canvas, there is no fold
// and no n/a box. States: none, dead, desert, ironwolf (the shield slot, a six-column doll), unknown, absent (the box hides), and
// a dead mercenary whose desk offers a revive verb. Usage: node merc_canvas_dom.js web/inventory-merc.js
const fs = require('fs'), vm = require('vm'), path = require('path');
const { mk, ser } = require('./stub.js');
const dir = __dirname, data = path.join(dir, '..');
const src = fs.readFileSync(process.argv[2], 'utf8');

function page() {
  const box = mk('div'), host = mk('div'), sub = mk('span');
  host.dataset.canvas = ''; host.hidden = true; box.hidden = true; box.appendChild(host);
  const document = { getElementById: (i) => (i === 'merc-doll' ? host : i === 'merc-sub' ? sub : null), createElement: mk, createTextNode: (t) => ({ textContent: t, tag: '#text', children: [], dataset: {}, className: '', attrs: {} }), addEventListener() {} };
  let h = null;
  const window = { __mercTest: (x) => { h = x; } };
  vm.runInNewContext(src, { document, window, location: { search: '' }, JSON, Math, Object, Array, String, Number, RegExp, console, Image: undefined });
  return { h, host, box, sub };
}
let bad = 0;
const check = (name, got) => {
  const file = path.join(dir, 'merc_canvas_' + name + '.txt');
  if (process.env.UPDATE) { fs.writeFileSync(file, got); return; }
  if (!fs.existsSync(file)) { console.error('missing golden ' + file); bad++; return; }
  if (fs.readFileSync(file, 'utf8') !== got) { console.error('merc canvas differs from ' + file + '\n' + got); bad++; }
};
const view = (n) => JSON.parse(fs.readFileSync(path.join(data, 'inventory_merc_' + n + '.json'), 'utf8'));
const dump = (p) => 'label: ' + p.sub.textContent + '\nbox hidden: ' + p.box.hidden + '\n' + ser(p.host, 0);

for (const n of ['none', 'desert', 'ironwolf', 'dead', 'unknown', 'absent']) {
  const p = page(); p.h.render(view(n)); check(n, dump(p));
}
{
  const v = view('dead'); v.merc.revive = { path: '/api/inventory/revive', cost: 600 };
  const p = page(); p.h.render(v); check('dead_revive', dump(p));
}
if (bad) process.exit(1);
console.log('ok');
