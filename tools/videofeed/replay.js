// node tools/videofeed/replay.js stream.pf   decode timing of a recorded wire stream (written by TestReplay with D2FEED_MSGS)
const fs = require('node:fs'), zlib = require('node:zlib');
const PF = require('../../internal/watch/web/palfeed.js');
const wire = new Uint8Array(fs.readFileSync(process.argv[2]));
const msgs = new PF.Splitter().push(wire);
const dec = new PF.Decoder(), inflateMs = [], applyMs = [], bytes = [];
for (const m of msgs) {
  const h = PF.parseHeader(m);
  const t0 = process.hrtime.bigint();
  const raw = new Uint8Array(zlib.inflateRawSync(m.subarray(PF.HEADER)));
  const t1 = process.hrtime.bigint();
  dec.apply(h, raw);
  const t2 = process.hrtime.bigint();
  if (!h.key) { inflateMs.push(Number(t1 - t0) / 1e6); applyMs.push(Number(t2 - t1) / 1e6); bytes.push(m.length); }
}
const pc = (a, p) => { const s = [...a].sort((x, y) => x - y); return s[Math.min(s.length - 1, Math.max(0, Math.ceil(s.length * p / 100) - 1))]; };
console.log(`node ${process.version}: ${msgs.length} messages, ${bytes.length} deltas`);
console.log(`delta bytes p50 ${pc(bytes, 50)} p95 ${pc(bytes, 95)}`);
console.log(`inflate ms p50 ${pc(inflateMs, 50).toFixed(3)} p95 ${pc(inflateMs, 95).toFixed(3)}`);
console.log(`apply   ms p50 ${pc(applyMs, 50).toFixed(3)} p95 ${pc(applyMs, 95).toFixed(3)}`);
