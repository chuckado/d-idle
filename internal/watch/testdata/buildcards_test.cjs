// web/buildcards.js in the stub DOM: the small card and the detail view as an indented text golden (testdata/buildcards_dom.txt,
// UPDATE=1 rewrites it), and the confirm sheet's focus, Tab, Escape and press-outside behaviour. The DOM does not depend on the
// window width (only the CSS does), so one golden stands for 390x844 and 1366x768; the layout at both sizes is the fit lane's
// (tools/uishots.sh --fit --only 'builds-*'). Usage: node buildcards_test.cjs DIR (DIR holds buildcards.cjs, nav_dom.cjs, data.json).
'use strict';
const fs = require('fs'), path = require('path');
const { Ev, makeDoc } = require('./nav_dom.cjs');
const dir = process.argv[2];
const BC = require(path.join(dir, 'buildcards.cjs'));
const data = JSON.parse(fs.readFileSync(path.join(dir, 'data.json'), 'utf8'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };

function newDoc() {
  const doc = makeDoc();
  doc.N.prototype.remove = function () { if (this.parent) { this.parent.children = this.parent.children.filter((x) => x !== this); this.parent = null; } };
  return doc;
}
function dump(n, depth, out) {
  const a = Object.keys(n.attrs).sort().map((k) => k + '=' + n.attrs[k]);
  for (const k of Object.keys(n.dataset).sort()) a.push('data-' + k + '=' + n.dataset[k]);
  for (const k of ['id', 'type', 'name', 'value']) if (n[k]) a.push(k + '=' + n[k]);
  if (n.disabled) a.push('disabled');
  if (n.checked) a.push('checked');
  const cls = n.className ? '.' + n.className.split(' ').join('.') : '';
  out.push('  '.repeat(depth) + n.tag + cls + (a.length ? ' [' + a.join(' ') + ']' : '') + (n.textContent ? ' ' + JSON.stringify(n.textContent) : ''));
  n.children.forEach((c) => dump(c, depth + 1, out));
  return out;
}

// 1. the golden
const doc = newDoc();
const card = BC.cardNode(doc, data.card, { open: false });
let out = ['== small card ==', ...dump(card, 0, [])];
const cases = [['detail, auto-allocation on', { autoAlloc: true, locked: false }], ['detail, auto-allocation off', { autoAlloc: false, locked: false }],
  ['detail, locked to another build', { autoAlloc: true, locked: true, chosen: 'x', label: 'Another build' }], ['detail, this build chosen', { autoAlloc: true, locked: true, chosen: data.detail.variants[0].name, label: 'This' }]];
for (const [name, st] of cases) out.push('== ' + name + ' ==', ...dump(BC.detailNode(doc, data.detail, st, {}), 0, []));
out = out.join('\n') + '\n';
const golden = path.join(dir, 'buildcards_dom.txt');
if (process.env.UPDATE) fs.writeFileSync(process.env.UPDATE, out);
else if (fs.readFileSync(golden, 'utf8') !== out) { fs.writeFileSync(path.join(dir, 'buildcards_dom.got'), out); fails.push('the DOM differs from testdata/buildcards_dom.txt (UPDATE=<path> rewrites it)'); }

// 2. accessibility facts, independent of the golden
const rates = card.all().filter((n) => n.className === 'bc-r');
ok(rates.length === 4, 'the small card shows four ratings, got ' + rates.length);
ok(rates[0].children[0].textContent.trim() === 'Defense', 'the first one is Defense');
for (const r of rates) {
  const st = r.children[1];
  ok(st.attrs.role === 'img' && /^[A-Za-z() ]+: [1-5] of 5 stars, (estimate|measured)/.test(st.attrs['aria-label']), 'text equivalent: ' + st.attrs['aria-label']);
  ok(/^[★☆]{5}$/.test(st.textContent), 'visible stars: ' + st.textContent);
}
ok(card.all().some((n) => n.className === 'bc-src' && n.textContent === 'Estimate'), 'the card says Estimate next to the stars');
ok(card.all().filter((n) => n.tag === 'button').length === 1 && card.all().find((n) => n.tag === 'button').attrs['aria-expanded'] === 'false', 'one Details button, collapsed');
let detailsCalled = 0; BC.cardNode(doc, data.card, { onDetails: () => { detailsCalled++; } }).all().find((n) => n.tag === 'button').click();
ok(detailsCalled === 1, 'Details calls back');
const offDet = BC.detailNode(doc, data.detail, { autoAlloc: false, locked: false }, {});
const offBtn = offDet.all().find((n) => n.className.includes('bc-choose'));
ok(offBtn.disabled && offDet.all().some((n) => n.className === 'bc-why' && n.textContent === 'Auto-allocation is off'), 'off: Choose is disabled and the reason line is visible');
ok(offDet.all().filter((n) => n.className === 'bc-r').length === 7, 'the detail shows all seven ratings');

// 3. the confirm sheet
{
  const d = newDoc(), opener = d.createElement('button'); d.body.appendChild(opener); opener.focus();
  let got = null;
  const s = BC.openSheet(d, { title: 'Choose X?', lines: ['You can change later, but it costs gold to respec.'], variants: [{ name: 'a', label: 'A', tier: 'supported' }, { name: 'b', label: 'B', tier: 'experimental' }], opener, onConfirm: (n) => { got = n; } });
  ok(s.sheet.attrs.role === 'dialog' && s.sheet.attrs['aria-modal'] === 'true' && s.sheet.attrs['aria-labelledby'], 'a modal dialog with a label');
  ok(d.activeElement && s.sheet.contains(d.activeElement), 'focus is inside the sheet after opening');
  ok(d.body.children.filter((n) => n.className === 'd-sheet-back').length === 1, 'exactly one sheet');
  const radios = s.sheet.all().filter((n) => n.tag === 'input');
  ok(radios.length === 2 && radios[0].checked, 'two variants: two radios, the first picked');
  // Tab from the last control wraps to the first
  s.confirm.focus(); s.node.dispatchEvent(Object.assign(new Ev('keydown', { cancelable: true }), { key: 'Tab' }));
  ok(d.activeElement === radios[0], 'Tab from the last control wraps to the first');
  s.node.dispatchEvent(Object.assign(new Ev('keydown', { cancelable: true }), { key: 'Escape' }));
  ok(!d.body.children.includes(s.node), 'Escape removes the sheet');
  ok(d.activeElement === opener, 'and the focus goes back to the opener');
  // press on the backdrop closes, a press inside does not
  const s2 = BC.openSheet(d, { title: 'T', lines: [], variants: [{ name: 'a', label: 'A', tier: 'supported' }], opener, onConfirm: (n) => { got = n; } });
  ok(s2.sheet.all().filter((n) => n.tag === 'input').length === 0, 'one variant: no radios');
  s2.sheet.click(); ok(d.body.children.includes(s2.node), 'a press inside the sheet keeps it');
  s2.node.click(); ok(!d.body.children.includes(s2.node) && d.activeElement === opener, 'a press on the backdrop closes it and returns the focus');
  // choosing the second variant then confirming
  const s3 = BC.openSheet(d, { title: 'T', lines: [], variants: [{ name: 'a', label: 'A', tier: 'supported' }, { name: 'b', label: 'B', tier: 'x' }], opener, onConfirm: (n) => { got = n; } });
  const r3 = s3.sheet.all().filter((n) => n.tag === 'input'); r3[1].dispatchEvent(new Ev('change')); s3.confirm.click();
  ok(got === 'b' && !d.body.children.includes(s3.node), 'Confirm hands over the picked variant and closes');
  // Cancel
  const s4 = BC.openSheet(d, { title: 'T', lines: [], variants: [], opener, onConfirm: () => { got = 'wrong'; } });
  s4.cancel.click(); ok(got === 'b' && d.activeElement === opener, 'Cancel closes without confirming');
}
console.log(JSON.stringify({ fails }));
