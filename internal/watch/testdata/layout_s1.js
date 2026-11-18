// The S1 layout rules in node (docs/ui/redesign-2026-11-14/D-layouts.md): old layouts keep their cells, the Standard and
// Simple presets resolve against today's panel list, opt-in panels stay unmounted, a stand-in saves as its tile.
//   node layout_s1.js golden <layout.js> <layout.json>...        prints the cells of every file at every width (JSON)
//   node layout_s1.js check <layout.js> <golden.json> <web dir> <compat dir>
const assert = require('assert'), fs = require('fs'), path = require('path'), { load } = require('./panel_harness.js');
const eq = (a, b, msg) => assert.strictEqual(JSON.stringify(a), JSON.stringify(b), msg);
const WIDTHS = [390, 768, 1280, 1920];
const conf = { aliases: { 'hero-status': 'hero-vitals', 'hero-detail': 'hero-vitals', whyline: 'why', 'hero-now': 'now', voice: 'diary' }, compat: [], reserved: ['risk'] };
// the page of the layout tests since the shell existed: every panel and card the old layouts name
const page = ['video', 'activity', 'hero-vitals', 'map', 'gear', 'why', 'tally', 'currency', 'thoughts', 'now', 'orders-map', 'queue', 'rules', 'spend', 'chronicle',
  'diary', 'quests', 'cards', 'locked', 'bestiary', 'atlas', 'alerts', 'journal'].map((id) => ({ id, size: 's' }));
const cells = (p, noOrder) => p.items.map((i) => [i.id, i.col, i.row, i.colSpan, i.rowSpan, !!i.collapsed, noOrder ? '' : i.order, !!i.placed].join()).sort();

function goldenOf(L, files) {
  const out = {};
  for (const f of files) {
    const layout = JSON.parse(fs.readFileSync(f, 'utf8'));
    out[path.basename(f)] = {};
    for (const w of WIDTHS) { const p = L.plan(layout, w, '', conf, page); out[path.basename(f)][w] = { name: p.name, flow: p.flow, unknown: p.unknown, cells: cells(p) }; }
  }
  return out;
}

function noOverlap(p) {
  const m = {};
  for (const it of p.items) {
    const h = it.collapsed ? 1 : it.rowSpan;
    assert.ok(it.col >= 1 && it.col + it.colSpan - 1 <= p.cols, it.id + ' inside the columns');
    for (let r = it.row; r < it.row + h; r++) for (let c = it.col; c < it.col + it.colSpan; c++) {
      assert.ok(!m[r + ':' + c], it.id + ' overlaps ' + m[r + ':' + c]);
      m[r + ':' + c] = it.id;
    }
  }
}

if (process.argv[2] === 'golden') {
  const L = load(process.argv[3]);
  process.stdout.write(JSON.stringify(goldenOf(L, process.argv.slice(4)), null, 1) + '\n');
  process.exit(0);
}

const [, , , layoutJS, goldenFile, webDir, compatDir] = process.argv;
const L = load(layoutJS);
const readJSON = (f) => JSON.parse(fs.readFileSync(f, 'utf8'));

// 1. Old layouts keep their cells: the three editor exports (compat fixtures) and the pre-redesign default (now classic)
//    plan to exactly the cells the shell before S1 gave them (the golden was produced by that shell's layout.js).
{
  const golden = readJSON(goldenFile);
  const files = ['layout-v1-mockup.json', 'layout-v1-live.json', 'layout-v1-command.json'].map((n) => path.join(compatDir, n)).concat(path.join(webDir, 'layouts/classic.json'));
  const now = goldenOf(L, files);
  eq(Object.keys(now).sort(), Object.keys(golden).sort(), 'the same files');
  for (const f of Object.keys(golden)) for (const w of WIDTHS) eq(now[f][w], golden[f][w], f + ' at ' + w + ' plans to the cells it had before S1');
}

// 2. Standard and Simple against today's page: the panels the shell mounts (opt-in ones only when named), every pre-shell card.
const shell = readJSON(path.join(webDir, 'shell.json')).layout;
const manifests = fs.readdirSync(path.join(webDir, 'panels')).map((d) => readJSON(path.join(webDir, 'panels', d, 'panel.json')));
function today(layout) {
  const mounted = L.selectPanels(manifests.filter((m) => !m.mergedInto), layout);
  return mounted.filter((m) => m.band !== 'status').map((m) => ({ id: m.id, size: m.size, hidden: false })).concat(shell.compat.map((id) => ({ id, size: 'm' })));
}
const FALLBACKS = { 'needs-you': 'gear', orders: 'now', 'zone-map': 'orders-map', health: 'why', progress: 'tally', log: 'thoughts' }; // D.4
const built = (id) => manifests.some((m) => m.id === id); // a slice that ships a panel takes it off the fallback list (S4: orders)
for (const name of ['standard', 'simple', 'stream', 'debug']) {
  const layout = readJSON(path.join(webDir, 'layouts', name + '.json'));
  const targets = today(layout), have = new Set(targets.map((t) => t.id));
  for (const t of targets) assert.ok(!['needs-you', 'orders', 'health', 'progress', 'zone-map', 'log', 'statusline', 'needs-count'].filter((id) => !built(id)).includes(t.id), name + ': no new panel exists yet');
  for (const w of WIDTHS) {
    const p = L.plan(layout, w, '', shell, targets);
    eq(p.unknown, [], name + ' at ' + w + ': no unknown id');
    if (!p.flow) noOverlap(p);
    assert.strictEqual(new Set(p.items.map((i) => i.id)).size, p.items.length, name + ': a panel once');
    const bp = layout.breakpoints[p.name].panels;
    for (const e of name === 'debug' ? [] : bp) { // Debug also lists the old tiles, so a new tile's cell stays empty there until it ships
      if (e.hidden) continue;
      // no empty cell: every shown entry became an item, at its cell (a fallback is the target the entry names)
      const want = have.has(e.id) ? e.id : e.fallback;
      assert.ok(want && have.has(want), name + ' ' + p.name + ': ' + e.id + ' has nothing to show');
      const it = p.items.find((i) => i.id === want);
      assert.ok(it && it.placed, name + ' ' + p.name + ': ' + want + ' is placed');
      assert.strictEqual(it.via, have.has(e.id) ? undefined : e.id, name + ': via');
      eq([it.col, it.colSpan, it.rowSpan], [e.col, e.colSpan, e.rowSpan], name + ' ' + p.name + ': ' + e.id + ' keeps its cell');
      if (!p.flow && !e.collapsed) assert.strictEqual(it.row, e.row, name + ' ' + p.name + ': ' + e.id + ' keeps its row');
    }
    if (name === 'standard' || name === 'simple') for (const [id, fb] of Object.entries(FALLBACKS)) {
      if (!bp.some((e) => e.id === id && !e.hidden) || built(id)) continue;
      const it = p.items.find((i) => i.via === id);
      assert.ok(it && it.id === fb, name + ' ' + p.name + ': ' + id + ' shows ' + fb);
      assert.ok(!p.items.some((i) => i.id === fb && !i.via), name + ': ' + fb + ' is not placed twice');
    }
  }
  // the status band: the layout's list against the manifests; a not yet built panel uses its fallback
  if (name === 'standard' || name === 'simple' || name === 'stream' || name === 'debug') {
    eq(L.statusIds(layout, manifests), ['hero-status', 'hero-meters', 'alerts-count'], name + ': status band');
    eq(L.selectPanels(manifests, layout).filter((m) => m.band === 'status').map((m) => m.id).sort(), ['alerts-count', 'hero-meters', 'hero-status'], name + ': mounted status panels');
  }
}
// an old layout has no status key: the band is the page's own list
{
  const classic = readJSON(path.join(webDir, 'layouts/classic.json'));
  assert.strictEqual(L.statusIds(classic, manifests), null);
  eq(L.selectPanels(manifests, classic).length, manifests.filter((m) => !m.cockpitOnly && !m.optIn).length, 'classic mounts every panel but the cockpit-only and opt-in ones');
}

// 3. optIn: mounted only when the layout names it (placed, not hidden); a fixture manifest list
{
  const list = [{ id: 'a', band: 'side', optIn: true }, { id: 'b', band: 'side' }, { id: 'c', band: 'grid', optIn: true }, { id: 's', band: 'status', optIn: true }, { id: 'o', band: 'status' }];
  const lay = (panels, status) => Object.assign({ schema: 'watch.layout/1', breakpoints: { wide: { cols: 12, panels } } }, status ? { status } : {});
  const e = (id, extra) => Object.assign({ id, col: 1, row: 1, colSpan: 4, rowSpan: 2, order: 0 }, extra);
  const ids = (l) => L.selectPanels(list, l).map((m) => m.id);
  eq(ids(lay([e('a')])), ['a', 'b', 'o'], 'named: mounted');
  eq(ids(lay([e('a', { hidden: true })])), ['b', 'o'], 'named but hidden: not mounted');
  eq(ids(lay([e('b')])), ['b', 'o'], 'not named: not mounted');
  eq(ids(null), ['b', 'o'], 'no layout: not mounted');
  eq(ids(lay([e('c', { hidden: false })])), ['b', 'c', 'o'], 'a grid panel');
  eq(ids(lay([e('x', { fallback: 'a' })])), ['a', 'b', 'o'], 'a layout that names an opt-in panel as a fallback names it');
  eq(ids(lay([e('b')], ['s', 'o'])), ['b', 's', 'o'], 'a status key names the opt-in status panel');
  eq(ids(lay([e('b')], ['o'])), ['b', 'o'], 'status key without s');
  eq(ids(lay([e('b')], [{ id: 's2', fallback: 'o' }])), ['b', 'o'], 'status entry that is not a panel uses its fallback');
  eq(ids(lay([e('b')], [{ id: 's2' }])), ['b'], 'status entry that is not a panel and has no fallback is skipped');
}

// 4. a stand-in saves as the tile it stands in for, and the entry it displaced comes back (Edit layout still saves)
{
  const layout = readJSON(path.join(webDir, 'layouts/standard.json'));
  const targets = today(layout), p = L.plan(layout, 1920, '', shell, targets), pe = L.plan(layout, 1920, '', shell, targets, true); // pe: what the editor holds
  const bands = {};
  layout.breakpoints.wide.panels.forEach((e) => { bands[e.id] = e.band; });
  const saved = JSON.parse(JSON.stringify(layout));
  saved.breakpoints.wide.panels = L.toPanels(pe.items, bands);
  const out = L.withShadows(saved, layout);
  assert.strictEqual(JSON.stringify(out.status), JSON.stringify(layout.status), 'the status key survives');
  const w = out.breakpoints.wide.panels;
  for (const [id, fb] of Object.entries(FALLBACKS)) {
    if (built(id)) continue; // a built panel is no stand-in
    const e = w.find((x) => x.id === id), g = w.find((x) => x.id === fb);
    assert.ok(e && e.fallback === fb && !e.hidden, id + ' is saved as itself with its fallback');
    assert.ok(g && g.hidden, fb + ' keeps its hidden entry');
  }
  assert.strictEqual(new Set(w.map((x) => x.order)).size, w.length, 'orders stay unique');
  const again = L.plan(out, 1920, '', shell, targets);
  eq(cells(again, true), cells(p, true), 'the saved file plans to the same cells');
}
// 5. R1: a stand-in the user switched off stays off after saving and loading again
{
  const layout = readJSON(path.join(webDir, 'layouts/standard.json'));
  const targets = today(layout), saved = JSON.parse(JSON.stringify(layout));
  for (const bp of Object.values(saved.breakpoints)) { const e = bp.panels.find((x) => x.id === 'needs-you'); e.hidden = true; }
  const p = L.plan(saved, 1920, '', shell, targets, true);
  assert.ok(!p.items.some((i) => i.id === 'gear' && !i.hidden), 'a switched-off stand-in is not shown');
  assert.ok(p.off.includes('gear'), 'it is switched off');
  const out = L.withShadows(Object.assign({}, saved, { breakpoints: { wide: { cols: 12, panels: L.toPanels(p.items, {}) } } }), saved);
  const e = out.breakpoints.wide.panels.find((x) => x.id === 'needs-you');
  assert.ok(e && e.hidden && e.fallback === 'gear', 'saved as the hidden tile with its fallback');
  assert.ok(!L.plan(out, 1920, '', shell, targets).items.some((i) => i.id === 'gear'), 'still off after a reload');
}
// 6. R2: the OBS strip (overlay mode) keeps the page's own status band, whatever the layout lists
{
  const layout = readJSON(path.join(webDir, 'layouts/standard.json'));
  eq(L.selectPanels(manifests, layout, 'overlay').filter((m) => m.band === 'status').map((m) => m.id).sort(), manifests.filter((m) => m.band === 'status' && !m.cockpitOnly).map((m) => m.id).sort(), 'overlay band');
}
console.log('ok');
