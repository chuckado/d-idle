// measureMap(win, variant): the Map-tab rules of card mapphone, read from the live page. Returns one line
// "MAPPHONE PASS|FAIL variant: numbers; reasons". Rules: the visible map canvas is more than 200 px tall, its toolbar (-, +,
// Layers) and the Act | Area switch do not intersect the phone tab row, and the canvas holds drawn pixels unless the bot has
// not seen the hero yet (then the panel says why).
export function measureMap(win, variant) {
  const d = win.document, R = (n) => { const r = n && n.getBoundingClientRect(); return r && r.width && r.height ? r : null; };
  const vis = (sel) => [...d.querySelectorAll(sel)].map(R).filter(Boolean)[0] || null;
  const cv = vis('[data-panel="area-map"] canvas'), ctl = vis('[data-panel="area-map"] .ctl'), sw = vis('.d-mapswitch'), tabs = vis('.d-tabbar');
  const hit = (a, b) => !!(a && b && a.left < b.right && a.right > b.left && a.top < b.bottom && a.bottom > b.top);
  const why = [], num = {};
  const rect = (r) => r && [Math.round(r.left), Math.round(r.top), Math.round(r.right), Math.round(r.bottom)];
  num.canvas = rect(cv); num.ctl = rect(ctl); num.switch = rect(sw); num.tabs = rect(tabs);
  const empty = d.querySelector('[data-panel="area-map"] .empty'), msg = empty && !empty.hidden && R(empty) ? empty.textContent.trim() : '';
  if (!cv) why.push('no map canvas is visible');
  else if (cv.height <= 200) why.push('canvas height ' + Math.round(cv.height) + ' px, want > 200');
  if (tabs && hit(ctl, tabs)) why.push('map toolbar intersects the tab row');
  if (tabs && hit(sw, tabs)) why.push('Act | Area switch intersects the tab row');
  if (cv && cv.height > 0) {
    let px = 0;
    try { const c = d.querySelector('[data-panel="area-map"] canvas'), g = c.getContext('2d'), im = g.getImageData(0, 0, c.width, c.height).data; for (let i = 3; i < im.length; i += 4) if (im[i]) px++; } catch (e) { px = -1; }
    num.pixels = px;
    if (variant === 'waiting') { if (!msg) why.push('waiting state shows no message'); } else if (px === 0) why.push('canvas has no drawn pixels' + (msg ? ' (message: ' + msg + ')' : ''));
  }
  if (msg) num.message = msg;
  // the heights from the panel up to the page (a collapsed ancestor shows here): selector-ish name and px
  const ch = [];
  for (let n = d.querySelector('[data-panel="area-map"] canvas'); n && n !== d.body; n = n.parentElement) ch.push((n.className || n.tagName).toString().split(' ')[0] + ':' + Math.round(n.getBoundingClientRect().height) + (n.dataset && n.dataset.grow ? 'g' : ''));
  num.chain = ch.join('<');
  return 'MAPPHONE ' + (why.length ? 'FAIL' : 'PASS') + ' ' + variant + ' ' + win.innerWidth + 'x' + win.innerHeight + ': ' + JSON.stringify(num) + (why.length ? '; ' + why.join('; ') : '');
}
