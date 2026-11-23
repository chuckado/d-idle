// web/nav.js in a stub DOM: the page-to-destination map, the bar and the band, one aria-current, the More sheet
// (focus in, Escape, focus back), the Skills sub-tabs. Prints {"fails": [...]}; the Go test reads it.
'use strict';
const fs = require('fs'), path = require('path');
const { Ev, makeDoc, page } = require('./nav_dom.cjs');
Object.defineProperty(globalThis, 'CustomEvent', { value: Ev, configurable: true, writable: true });
const dir = process.argv[2];
const cfg = JSON.parse(fs.readFileSync(path.join(dir, 'nav.json'), 'utf8'));
const { pickCurrent, createNav, subtabs } = require(path.join(dir, 'nav.cjs'));
const fails = [];
const ok = (c, m) => { if (!c) fails.push(m); };
const want = ['Live', 'Gear', 'Skills', 'Vault', 'More'];

// 1. the address decides the destination
const cases = [['/watch', '', 'live'], ['/watch/', '', 'live'], ['/watch/inventory', '', 'gear'], ['/watch/skills', '', 'skills'], ['/watch/points', '', 'skills'], ['/watch/vault', '', 'vault'],
  ['/watch/characters', '', 'more'], ['/watch/return', '', 'more'], ['/watch/policy', '', 'more'], ['/watch/hall', '', 'more'], ['/watch/nowhere', '', null]];
for (const [p, q, d] of cases) ok(pickCurrent(cfg, p, q).dest === d, p + ' -> ' + d + ', got ' + pickCurrent(cfg, p, q).dest);
ok(pickCurrent(cfg, '/watch/skills', '').sub === 'tree', 'skills is the Tree sub-tab');
ok(pickCurrent(cfg, '/watch/points', '').sub === 'points', 'points is the Points sub-tab');
ok(pickCurrent(cfg, '/watch/points', '?tab=build').sub === 'build', 'points?tab=build is the Build sub-tab');
ok(pickCurrent(cfg, '/watch/vault', '').sub === null, 'the Vault has no sub-tab');

// 2. every page: five destinations, exactly one aria-current, on the bar (narrow) and in the band (wide)
function mount(pathname, search, narrow, kind) {
  const doc = makeDoc(); page(doc, kind);
  let narrowCb = null;
  const env = { doc, cfg, loc: { pathname, search }, narrow: () => narrow.v, width: () => 1200, store: { get: () => null, set() {} }, later: (f) => f(), onNarrow: (f) => { narrowCb = f; }, onResize() {} };
  const n = createNav(env); n.mount();
  return { doc, n, flip(v) { narrow.v = v; narrowCb(); } };
}
for (const [p, , d] of cases.filter((c) => c[2])) {
  for (const kind of ['band', 'bare']) for (const nar of [true, false]) {
    const m = mount(p, '', { v: nar }, kind);
    const items = m.n.nav.children;
    ok(items.map((i) => i.children[1].textContent).join() === want.join(), p + ': five destinations in order, got ' + items.map((i) => i.children[1].textContent));
    ok(items.filter((i) => i.attrs['aria-current'] === 'page').length === 1, p + ': exactly one aria-current');
    ok(items.find((i) => i.attrs['aria-current']).dataset.nav === d, p + ': the current one is ' + d);
    ok(items.filter((i) => i.href).map((i) => i.href).join() === '/watch?live=1,/watch/inventory,/watch/skills,/watch/vault', p + ': the four links');
    const inBody = m.doc.body.children.includes(m.n.nav);
    ok(inBody === nar, p + ': the nav is in <body> on a phone and in the band above (' + kind + ', narrow=' + nar + ')');
    ok(nar || kind !== 'band' || m.doc.querySelector('[data-nav-host]').children[0] === m.n.nav, p + ': in the page band, first');
  }
}

// 3. the bar follows the window
{
  const m = mount('/watch/vault', '', { v: false }, 'bare');
  ok(m.doc.body.dataset.nav === 'band', 'wide: body says band');
  m.flip(true);
  ok(m.doc.body.children.includes(m.n.nav) && m.doc.body.dataset.nav === 'bar', 'narrow: the nav moved to <body>');
  m.flip(false);
  ok(!m.doc.body.children.includes(m.n.nav), 'wide again: out of <body>');
}

// 4. the More sheet: opens with focus inside, Escape closes it and the focus goes back to More
{
  const m = mount('/watch/inventory', '', { v: true }, 'band');
  const more = m.n.moreButton(), sheet = m.n.sheet;
  ok(sheet.hidden && more.attrs['aria-expanded'] === 'false', 'the sheet starts closed');
  more.focus(); more.click();
  ok(!sheet.hidden && more.attrs['aria-expanded'] === 'true', 'More opens the sheet');
  ok(m.doc.activeElement && sheet.contains(m.doc.activeElement), 'focus is inside the sheet after opening');
  const first = m.doc.activeElement;
  first.dispatchEvent(Object.assign(new Ev('keydown'), { key: 'Escape' }));
  ok(sheet.hidden && more.attrs['aria-expanded'] === 'false', 'Escape closes the sheet');
  ok(m.doc.activeElement === more, 'focus goes back to More');
  more.click(); ok(!sheet.hidden, 'opens again');
  m.doc.body.dispatchEvent(new Ev('click')); // a press outside
  ok(sheet.hidden, 'a press outside closes it');
  // the entries: characters, return, hall, policy, help, old layout; theme in the settings group
  const links = sheet.all().filter((n) => n.tag === 'a').map((n) => n.href);
  for (const h of ['/watch/characters', '/watch/return', '/watch/hall', '/watch/policy', '/watch?live=1&drawer=onboarding', '/watch?layout=classic']) ok(links.includes(h), 'the sheet links ' + h);
  ok(sheet.all().some((n) => n.id === 'btn-theme'), 'the sheet holds the theme button (adopted)');
  ok(sheet.id === 'menu-more' && more.id === 'btn-more', 'ids the screenshot suite and fit check use');
  // the live screen can answer Help in place
  let seen = null; m.doc.addEventListener('d2:nav', (e) => { seen = e.detail.id; e.preventDefault(); });
  const help = sheet.all().find((n) => n.dataset.nav === 'help'); const ce = new Ev('click', { cancelable: true }); help.dispatchEvent(ce);
  ok(seen === 'help' && ce.defaultPrevented, 'Help raises d2:nav and the page may keep the click');
  // drop / hide
  m.n.drop('help'); ok(help.hidden, 'a page can drop an entry');
  // a group the page adds (the live screen's Hero / Tools)
  const b = m.doc.createElement('button'); m.n.group('tools', 'Tools', [b], true);
  ok(sheet.all().includes(b), 'dNav.group adds nodes');
}
// More is current on its own pages
{
  const m = mount('/watch/policy', '', { v: false }, 'bare');
  ok(m.n.moreButton().attrs['aria-current'] === 'page', 'More is current on /watch/policy');
  const cur = m.n.sheet.all().filter((n) => n.attrs && n.attrs['aria-current']);
  ok(cur.length === 1 && cur[0].dataset.nav === 'policy', 'and the sheet marks Policy');
}
// 4b. Settings: an entry on every page's sheet opens the Settings sheet in place (the page gets no navigation)
for (const [p, kind] of [['/watch/inventory', 'band'], ['/watch/vault', 'bare'], ['/watch/hall', 'bare'], ['/watch', 'band']]) {
  const doc = makeDoc(); page(doc, kind);
  let asked = null;
  const env = { doc, cfg, loc: { pathname: p, search: '' }, narrow: () => true, width: () => 390, store: { get: () => null, set() {} }, later: (f) => f(), onNarrow() {}, onResize() {}, settings: (o, sec) => { asked = [o, sec]; } };
  const n = createNav(env); n.mount();
  const a = n.sheet.all().find((x) => x.dataset.nav === 'settings');
  ok(a && a.tag === 'a' && a.textContent === 'Settings', p + ': the More sheet has a Settings entry');
  const ce = new Ev('click', { cancelable: true }); a.dispatchEvent(ce);
  ok(asked && asked[0] === a && ce.defaultPrevented, p + ': pressing Settings opens the sheet in place');
}

// 5. the theme button of a page that has none
{
  const m = mount('/watch/vault', '', { v: false }, 'bare');
  const t = m.n.sheet.all().find((n) => n.id === 'btn-theme');
  ok(t && /Theme: dark/.test(t.textContent), 'a theme button is made when the page has none');
  t.click(); ok(m.doc.documentElement.dataset.theme === 'light', 'it switches the theme');
}
// 6. sub-tabs of the Skills area
for (const [p, q, sub] of [['/watch/skills', '', 'tree'], ['/watch/points', '', 'points'], ['/watch/points', '?tab=build', 'build']]) {
  const doc = makeDoc(); const cur = pickCurrent(cfg, p, q);
  const bar = subtabs({ doc, cfg }, cur);
  ok(bar && bar.children.map((a) => a.textContent).join() === 'Tree,Points,Build', p + q + ': three sub-tabs');
  ok(bar.children.filter((a) => a.attrs['aria-current']).length === 1 && bar.children.find((a) => a.attrs['aria-current']).dataset.sub === sub, p + q + ': ' + sub + ' is current');
  ok(doc.body.dataset.subtab === sub, 'body[data-subtab] is ' + sub);
}
ok(subtabs({ doc: makeDoc(), cfg }, pickCurrent(cfg, '/watch/vault', '')) === null, 'no sub-tabs on the Vault');
console.log(JSON.stringify({ fails }));
