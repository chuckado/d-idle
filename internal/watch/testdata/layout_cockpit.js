// The cockpit template's pure logic (web/layout.js, S1b): the plan per width, slots left out without holes, which panels
// mount, the editor's swap and divider moves, and the round trip a Save and a reload make. Usage:
// node layout_cockpit.js path/to/layout.js path/to/cockpit.json
const assert = require('assert'), fs = require('fs'), { load } = require('./panel_harness.js');
const L = load(process.argv[2]);
const eq = (a, b, msg) => assert.strictEqual(JSON.stringify(a), JSON.stringify(b), msg);
const file = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'));
const all = (id) => true;

assert.ok(L.isCockpit(file));
assert.ok(!L.isCockpit({ schema: 'watch.layout/1', breakpoints: {} }), 'no template, no cockpit');
assert.ok(!L.isCockpit(Object.assign({}, file, { template: 'cockpit/9' })), 'an unknown template is not read');
assert.ok(!L.isCockpit(Object.assign({}, file, { slots: Object.assign({}, file.slots, { stage: [] }) })), 'a cockpit needs its stage');
eq([390, 699, 700, 1699, 1700, 2560].map((w) => L.cockpitMode(w, '')), ['phone', 'phone', 'page', 'page', 'wide', 'wide']);
assert.strictEqual(L.cockpitMode(1920, 'overlay'), '');
assert.strictEqual(L.cockpitPlan(file, 1920, 'overlay', all), null, 'the overlay never lays out a cockpit');
assert.strictEqual(L.cockpitPlan({ schema: 'watch.layout/1', breakpoints: {} }, 1440, '', all), null);

// every panel mounted: the page, the wide page (two rail columns) and the phone tabs
let p = L.cockpitPlan(file, 1440, '', all);
eq(p.band, ['statusline', 'hero-meters', 'needs-count']);
assert.strictEqual(p.stage, 'video');
eq(p.strip, ['progress', 'journey-strip', 'health']);
eq(p.rail, [['needs-you', 'orders', 'zone-map']]);
assert.strictEqual(p.has, 'str');
eq(p.sizes, { strip: 112, railMin: 400 });
eq(p.tabs, []);
p = L.cockpitPlan(file, 1920, '', all);
eq(p.top, ['needs-you', 'orders']);
eq(p.rail, [['zone-map'], ['area-map']]);
p = L.cockpitPlan(file, 390, '', all);
eq(p.tabs.map((t) => t.name), ['Orders', 'Map', 'Progress']);
eq(p.strip, []); eq(p.rail, []);
assert.strictEqual(p.has, 'sp');

// today's panels: only the picture and the meters exist; the rest of the layout is left out and nothing is in its place
const today = (id) => id === 'video' || id === 'hero-meters';
p = L.cockpitPlan(file, 1440, '', today);
eq(p.band, ['hero-meters']); assert.strictEqual(p.stage, 'video'); eq(p.strip, []); eq(p.rail, []);
assert.strictEqual(p.has, 's', 'the picture takes the whole frame');
assert.ok(p.left.includes('orders') && p.left.includes('progress'));
p = L.cockpitPlan(file, 390, '', today);
assert.strictEqual(p.has, 's'); eq(p.tabs, []);
// a column that is empty on the wide page leaves the other one alone
p = L.cockpitPlan(file, 1920, '', (id) => id === 'video' || id === 'zone-map');
eq(p.top, []); eq(p.rail, [['zone-map']]);
assert.strictEqual(p.has, 'sr');
// only the top panels: the rail is them alone
p = L.cockpitPlan(file, 1920, '', (id) => id === 'video' || id === 'orders');
eq(p.top, ['orders']); eq(p.rail, []); assert.strictEqual(p.has, 'sr');

// a short window (under 760 px high, not a phone) has a strip of at most 92 px; the layout's own size is kept otherwise
eq(L.cockpitPlan(file, 1536, '', all, 730).sizes.strip, 92);
eq(L.cockpitPlan(file, 1280, '', all, 620).sizes.strip, 92);
eq(L.cockpitPlan(file, 1366, '', all, 768).sizes.strip, 112);
eq(L.cockpitPlan(file, 1440, '', all, 900).sizes.strip, 112);
eq(L.cockpitPlan(file, 390, '', all, 600).sizes.strip, 112);
const small = JSON.parse(JSON.stringify(file)); small.sizes.strip = 80;
eq(L.cockpitPlan(small, 1280, '', all, 620).sizes.strip, 80, 'a strip smaller than 92 stays');

// sizes: defaults, clamped to the limits, set by the editor inside them
eq(L.cockpitSizes({}), { strip: 112, railMin: 400 });
eq(L.cockpitSizes({ sizes: { strip: 5, railMin: 9999 } }), { strip: 48, railMin: 800 });
let s = L.cockpitSize(file, { strip: 150.4, railMin: 100 });
eq(s.sizes, { strip: 150, railMin: 240 });
eq(file.sizes, { strip: 112, railMin: 400 }, 'the layout it started from is not changed');

// swap: two panels trade places in every arrangement; one that is not in the layout replaces
let sw = L.cockpitSwap(file, 'progress', 'orders');
eq(sw.slots.strip, ['orders', 'journey-strip', 'health']);
eq(sw.slots.rail, ['needs-you', 'progress', 'zone-map']);
eq(sw.slots.railWide.top, ['needs-you', 'progress']);
eq(sw.slots.phoneTabs.Orders, ['needs-you', 'progress']);
eq(sw.slots.phoneTabs.Progress, ['orders', 'journey-strip', 'health']);
eq(L.cockpitSwap(sw, 'progress', 'orders'), file, 'swapping back restores the file');
sw = L.cockpitSwap(file, 'video', 'fx-new');
assert.strictEqual(sw.slots.stage, 'video' === file.slots.stage ? 'fx-new' : '');
eq(L.cockpitSwap(file, 'health', 'health'), file);
// a stage given as a list swaps too
const listy = JSON.parse(JSON.stringify(file)); listy.slots.stage = ['video'];
eq(L.cockpitSwap(listy, 'video', 'orders').slots.stage, ['orders']);

// the editor: swap two tiles, move both dividers, Save (a JSON round trip, as POST and GET make it), reload: same arrangement
const edited = L.cockpitSize(L.cockpitSwap(file, 'health', 'zone-map'), { strip: 140, railMin: 480 });
const reloaded = JSON.parse(JSON.stringify(edited));
for (const w of [390, 1440, 1920]) eq(L.cockpitPlan(reloaded, w, '', all), L.cockpitPlan(edited, w, '', all), 'same arrangement at ' + w);
p = L.cockpitPlan(reloaded, 1440, '', all);
eq(p.strip, ['progress', 'journey-strip', 'zone-map']); eq(p.rail, [['needs-you', 'orders', 'health']]);
eq(p.sizes, { strip: 140, railMin: 480 });

// which panels mount: a cockpit mounts only slot panels it names, whatever else the page has; other layouts as before
const list = [{ id: 'video', slot: true, band: 'stage' }, { id: 'hero-meters', slot: true, band: 'status' }, { id: 'hero-status', band: 'status' }, { id: 'map', band: 'side' },
  { id: 'orders', slot: true, optIn: true, band: 'side' }, { id: 'unused', slot: true, band: 'side' }];
eq(L.selectPanels(list, file, '').map((m) => m.id), ['video', 'hero-meters', 'orders']);
// the OBS strip is not a cockpit: under a cockpit it mounts exactly what it mounts under Classic (no opt-in panel of a slot)
const classic = JSON.parse(fs.readFileSync(process.argv[3].replace('cockpit.json', 'classic.json'), 'utf8'));
eq(L.selectPanels(list, file, 'overlay').map((m) => m.id), L.selectPanels(list, classic, 'overlay').map((m) => m.id));
assert.ok(!L.selectPanels(list, file, 'overlay').some((m) => m.id === 'orders'));
eq(L.cockpitIds(file).size > 8, true);
console.log('layout_cockpit ok');
