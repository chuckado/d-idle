// web/sheet.js in the stub DOM (node test, card uisheet): roles, focus in, Tab wrap, Escape, backdrop, return of focus, danger, notice, ask().
'use strict';
const path = require('path');
const { Ev, makeDoc } = require('./nav_dom.cjs');
const S = require(path.join(process.argv[2], 'sheet.cjs'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const key = (n, k, extra) => n.dispatchEvent(Object.assign(new Ev('keydown', { cancelable: true }), { key: k }, extra));
const sheets = (d) => d.body.children.filter((n) => n.className === 'd-sheet-back');
function newDoc() {
  const doc = makeDoc();
  doc.N.prototype.remove = function () { if (this.parent) { this.parent.children = this.parent.children.filter((x) => x !== this); this.parent = null; } };
  return doc;
}
(async () => {
  const d = newDoc(), opener = d.createElement('button'); d.body.appendChild(opener); opener.focus();
  let got = '';
  const s = S.open(d, { title: 'Revoke?', lines: ['one', 'two'], danger: true, confirmText: 'Revoke', opener, onConfirm: () => { got = 'ok'; }, onCancel: () => { got = 'cancel'; } });
  ok(s.sheet.attrs.role === 'alertdialog' && s.sheet.attrs['aria-modal'] === 'true', 'a danger sheet is a modal alertdialog');
  ok(s.sheet.attrs['aria-labelledby'] && d.getElementById(s.sheet.attrs['aria-labelledby']).textContent === 'Revoke?', 'labelled by its title');
  ok(s.sheet.attrs['aria-describedby'] && d.getElementById(s.sheet.attrs['aria-describedby']).textContent === 'one', 'described by its first line');
  ok(s.confirm.attrs['data-variant'] === 'danger' && s.confirm.textContent === 'Revoke', 'the danger confirm button is the danger variant');
  ok(s.sheet.contains(d.activeElement), 'focus moves into the sheet');
  s.confirm.focus(); key(s.node, 'Tab');
  ok(d.activeElement === s.cancel, 'Tab from the last control wraps to the first');
  key(s.node, 'Tab', { shiftKey: true });
  ok(d.activeElement === s.confirm, 'Shift+Tab from the first wraps to the last');
  key(s.node, 'Escape');
  ok(sheets(d).length === 0 && d.activeElement === opener && got === 'cancel', 'Escape closes, calls onCancel and returns the focus');
  const s2 = S.open(d, { title: 'T', opener, onCancel: () => { got = 'backdrop'; } });
  s2.sheet.click(); ok(sheets(d).length === 1, 'a press inside keeps the sheet');
  s2.node.click(); ok(sheets(d).length === 0 && got === 'backdrop' && d.activeElement === opener, 'a press on the backdrop closes it');
  const s3 = S.open(d, { title: 'T', opener, onConfirm: () => { got = 'confirmed'; } });
  ok(s3.sheet.attrs.role === 'dialog' && s3.confirm.attrs['data-variant'] === 'primary', 'a plain sheet is a dialog with a primary confirm');
  s3.confirm.click(); ok(got === 'confirmed' && sheets(d).length === 0, 'Confirm calls onConfirm and closes');
  const n = S.notice(d, { title: 'Oops', lines: ['it failed'], opener });
  const sn = sheets(d)[0];
  ok(sn && sn.all().filter((x) => x.tag === 'button').length === 1, 'a notice has one button');
  sn.all().find((x) => x.tag === 'button').click(); await n;
  ok(sheets(d).length === 0, 'the notice closes and resolves');
  const a = S.ask(d, { title: 'Sure?', opener }); sheets(d)[0].all().find((x) => x.className.includes('d-sheet-ok')).click();
  ok((await a) === true, 'ask resolves true on confirm');
  const a2 = S.ask(d, { title: 'Sure?', opener }); key(sheets(d)[0], 'Escape');
  ok((await a2) === false, 'ask resolves false on cancel');
  S.open(d, { title: 'one' }); S.open(d, { title: 'two' });
  ok(sheets(d).length === 2 && new Set(sheets(d).map((x) => x.all().find((y) => y.tag === 'h2').id)).size === 2, 'two sheets get distinct ids');
  console.log(JSON.stringify({ fails }));
})();
