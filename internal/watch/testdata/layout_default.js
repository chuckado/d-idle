// The default layout of S7 (web/layout.js): which stored browser choice counts, and the slot self-check.
// Usage: node layout_default.js path/to/web/layout.js path/to/web/layout.json
const assert = require('assert'), fs = require('fs'), { load } = require('./panel_harness.js');
const L = load(process.argv[2]);
const def = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'));
// the shipped default is the cockpit
assert.ok(L.isCockpit(def), 'web/layout.json is a cockpit');
// a Simple that no menu click wrote (an older shell pinned it for a new hero) is no choice; every real choice is kept
assert.strictEqual(L.effectiveChoice('simple', ''), '', 'auto-pinned simple follows the default');
assert.strictEqual(L.effectiveChoice('simple', '1'), 'simple', 'a menu choice of simple is kept');
for (const c of ['classic', 'standard', 'stream', 'debug', 'saved', 'cockpit']) {
  assert.strictEqual(L.effectiveChoice(c, ''), c, c + ' is kept without the marker');
  assert.strictEqual(L.effectiveChoice(c, '1'), c, c + ' is kept with the marker');
}
assert.strictEqual(L.effectiveChoice('', ''), '');
assert.strictEqual(L.effectiveChoice(undefined, ''), '');
// the self-check: every slot panel must be offered
const ids = [...L.cockpitIds(def)];
assert.ok(ids.length >= 9, 'the cockpit names its panels: ' + ids.join());
assert.strictEqual(JSON.stringify(L.missingSlots(def, ids)), JSON.stringify([]), 'all present');
assert.strictEqual(JSON.stringify(L.missingSlots(def, ids.filter((i) => i !== 'orders'))), JSON.stringify(['orders']), 'one missing is named');
assert.strictEqual(JSON.stringify(L.missingSlots({ schema: 'x' }, [])), JSON.stringify([]), 'not a cockpit: nothing to check');
console.log('ok');
