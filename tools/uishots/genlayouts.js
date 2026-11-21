// Writes the layout files of internal/watch/web (run from the repo root: node tools/uishots/genlayouts.js tools/uishots/layout-rows.json):
//  - layouts/{mockup,live,command}.json: the mockup editor's own preset code (docs/mockups/ui/editor.html), as it exports them;
//  - layouts/{simple,standard,stream,debug}.json and layout.json (the default, Standard): the redesign presets of the rows
//    file (docs/ui/redesign-2026-11-14/D-layouts.md; the same rows and rules as docs/ui/redesign-2026-11-14/layouts/gen.py).
// layouts/classic.json is the pre-redesign default, frozen: it is not generated (it had drifted from the mockup rows).
const fs = require('fs'), vm = require('vm');
const src = fs.readFileSync('docs/mockups/ui/editor.html', 'utf8').split('\n');
const a = src.findIndex((l) => l.startsWith('var LIB = ')), b = src.findIndex((l) => l.startsWith('function load()'));
const code = src.slice(a, b).filter((l) => !/^var \$|^var canvas/.test(l)).join('\n');
const ctx = {}; vm.runInNewContext(code + '\nthis.x={blank,fromPreset,BPS,byId,PRESETS,LIB,place};', ctx);
const { blank, BPS, PRESETS, byId } = ctx.x;
// the editor's presets overlap in places (it resolves an overlap only on a drop); the shipped files carry the same
// push-down the editor and the shell apply, so tools/uilint's no-overlap rule holds
const occ = (ps, skip) => { const m = {}; ps.forEach((p) => { if (p === skip) return; for (let r = p.row; r < p.row + p.rowSpan; r++) for (let c = p.col; c < p.col + p.colSpan; c++) m[r + ':' + c] = 1; }); return m; };
const fromPreset = (name, bp) => {
  const out = [];
  for (const p of ctx.x.fromPreset(name, bp)) {
    for (let m = occ(out), ok = false; !ok; p.row++) {
      ok = true;
      for (let r = p.row; r < p.row + p.rowSpan && ok; r++) for (let c = p.col; c < p.col + p.colSpan; c++) if (m[r + ':' + c]) { ok = false; break; }
      if (ok) break;
    }
    out.push(p);
  }
  return out;
};
const out = 'internal/watch/web/';
fs.mkdirSync(out + 'layouts', { recursive: true });
// argv[3]: also write the raw exports (overlaps and all) there, as the editor's Export JSON gives them (compat fixtures)
if (process.argv[3]) for (const name of Object.keys(PRESETS)) {
  const o = blank(); Object.keys(BPS).forEach((bp) => { o.breakpoints[bp].panels = ctx.x.fromPreset(name, bp); });
  const f = process.argv[3] + '/layout-v1-' + name + '.json';
  if (!fs.existsSync(f)) fs.writeFileSync(f, JSON.stringify(o, null, 2) + '\n'); // fixtures are only added, never rewritten
}
for (const name of Object.keys(PRESETS)) {
  const o = blank(); Object.keys(BPS).forEach((bp) => { o.breakpoints[bp].panels = fromPreset(name, bp); });
  fs.writeFileSync(out + 'layouts/' + name + '.json', JSON.stringify(o, null, 2) + '\n');
}
// ---- the redesign presets ----
const rowsFile = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
const web = out;
const shell = JSON.parse(fs.readFileSync(web + 'shell.json', 'utf8')).layout;
const ids = new Set([...shell.compat, ...shell.reserved, ...Object.keys(shell.aliases)]), pbands = {};
for (const d of fs.readdirSync(web + 'panels')) {
  const f = web + 'panels/' + d + '/panel.json';
  if (!fs.existsSync(f)) continue;
  const m = JSON.parse(fs.readFileSync(f, 'utf8'));
  ids.add(m.id); pbands[m.id] = m.band;
}
const bandOf = (id) => rowsFile.bands[id] || pbands[id] || (shell.compat.includes(id) ? 'grid' : 'side');
const every = [...ids].filter((i) => !rowsFile.skipHidden.includes(i)).sort();

function entries(rows) {
  const shown = rows.map((r) => r[0]);
  const es = rows.map(([id, col, row, colSpan, rowSpan], n) => ({ id, band: bandOf(id), col, row, colSpan, rowSpan, collapsed: false, order: n }));
  for (const e of es) if (rowsFile.fallback[e.id] && !shown.includes(rowsFile.fallback[e.id])) e.fallback = rowsFile.fallback[e.id];
  let row = Math.max(...es.map((e) => e.row + e.rowSpan));
  for (const id of every) { // every other panel is named and switched off, so the shell does not pack it after the rest
    if (shown.includes(id)) continue;
    es.push({ id, band: bandOf(id), col: 1, row, colSpan: 12, rowSpan: 2, collapsed: false, order: es.length, hidden: true });
    row += 2;
  }
  return es;
}

// phone and tablet: the grid's panels in the flow order, stacked; on a tablet two half-width neighbours share a row
function flowed(spec, tablet) {
  const span = {};
  for (const [id, , , w, h] of spec.grid) span[id] = [w, h];
  const flow = spec.flow || spec.grid.map((r) => [r[0], 12]);
  let rows = [], r = 1;
  for (const [id, w] of flow) { rows.push([id, 1, r, tablet ? w : 12, span[id][1]]); r += span[id][1]; }
  if (tablet) {
    const fixed = [];
    for (let k = 0; k < rows.length;) {
      const a = rows[k], b = rows[k + 1];
      if (a[3] === 6 && b && b[3] === 6) { const h = Math.max(a[4], b[4]); fixed.push([a[0], 1, a[2], 6, h], [b[0], 7, a[2], 6, h]); k += 2; } else { fixed.push(a); k++; }
    }
    const packed = [];
    r = 1;
    for (let k = 0; k < fixed.length;) {
      const a = fixed[k], b = fixed[k + 1];
      if (a[3] === 6 && b && b[1] === 7) { packed.push([a[0], 1, r, 6, a[4]], [b[0], 7, r, 6, b[4]]); r += a[4]; k += 2; } else { packed.push([a[0], 1, r, a[3], a[4]]); r += a[4]; k++; }
    }
    rows = packed;
  }
  return entries(rows);
}

const redesign = {};
for (const [name, spec] of Object.entries(rowsFile.presets)) {
  const grid = entries(spec.grid);
  redesign[name] = {
    schema: 'watch.layout/1', cols: 12, rowHeight: 48, density: spec.density, theme: 'dark', status: rowsFile.status,
    breakpoints: {
      phone: { width: 390, cols: 12, panels: flowed(spec, false) },
      tablet: { width: 768, cols: 12, panels: flowed(spec, true) },
      laptop: { width: 1280, cols: 12, panels: grid },
      wide: { width: 1920, cols: 12, panels: grid },
    },
  };
  fs.writeFileSync(out + 'layouts/' + name + '.json', JSON.stringify(redesign[name], null, 2) + '\n');
}
fs.writeFileSync(out + 'layout.json', JSON.stringify(redesign.standard, null, 2) + '\n'); // the default is Standard
