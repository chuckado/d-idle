// The "older layouts render the same cells" guard of the cockpit slice (S1b). It plans every layout file that has no
// template at five widths, in the page and in the editor, and hashes what the shell would place (cells, spans, order,
// hidden, fallbacks) together with which panels selectPanels mounts. The expected hashes (layout_cells.golden) were taken
// with the layout.js of before the cockpit existed; a change there that moves a cell fails this test.
// Usage: node layout_cells.js path/to/layout.js layout_cells.golden [--write] layout.json...
const crypto = require('crypto'), fs = require('fs'), path = require('path'), assert = require('assert');
const { load } = require('./panel_harness.js');
const L = load(process.argv[2]);
const golden = process.argv[3], write = process.argv[4] === '--write', files = process.argv.slice(write ? 5 : 4);
const conf = { aliases: { 'hero-status': 'hero-vitals', 'hero-detail': 'hero-vitals', whyline: 'why', 'hero-now': 'now', voice: 'diary' }, compat: [], reserved: ['risk', 'statusline', 'needs-you', 'orders'] };
const page = ['video', 'activity', 'hero-vitals', 'map', 'gear', 'why', 'tally', 'currency', 'thoughts', 'now', 'orders-map', 'queue', 'rules', 'spend', 'chronicle',
  'diary', 'quests', 'cards', 'locked', 'bestiary', 'atlas', 'alerts', 'journal'].map((id, i) => ({ id, size: i % 3 ? 's' : 'm', hidden: id === 'locked', rows: id === 'now' ? 5 : 0 }));
const manifests = [{ id: 'video', band: 'stage' }, { id: 'hero-status', band: 'status' }, { id: 'hero-meters', band: 'status' }, { id: 'alerts-count', band: 'status' },
  { id: 'journey-chip', band: 'status' }, { id: 'hero-vitals', band: 'side' }, { id: 'statusline', band: 'status', optIn: true }, { id: 'needs-you', band: 'side', optIn: true }, { id: 'onboarding', band: 'side', optIn: true }];
const lines = [];
for (const f of files) {
  const layout = JSON.parse(fs.readFileSync(f, 'utf8')), name = path.basename(f);
  const h = crypto.createHash('sha256');
  for (const w of [390, 768, 1280, 1440, 1920]) for (const edit of [false, true]) h.update(JSON.stringify(L.plan(layout, w, '', conf, page, edit)));
  h.update(JSON.stringify(L.plan(layout, 1920, 'overlay', conf, page)));
  for (const mode of ['', 'overlay']) h.update(JSON.stringify(L.selectPanels(manifests, layout, mode).map((m) => m.id)));
  lines.push(name + ' ' + h.digest('hex'));
}
if (write) { fs.writeFileSync(golden, lines.join('\n') + '\n'); process.exit(0); }
const want = fs.readFileSync(golden, 'utf8').trim().split('\n');
assert.deepStrictEqual(lines, want, 'a layout without a template no longer renders the same cells');
