// node tools/videofeed/stream3.js DIR PORT: serves the pacing bench (bench3.html, palfeed.js) and DIR/stream.pf as a paced
// feed on /stream.pfeed: one message per 40 ms (25 Hz stamps in the headers) delivered with 0 to 25 ms of network-like jitter
// and now and then an 80 ms stall after which the backlog arrives in a burst (TCP never reorders: an arrival is never earlier
// than the one before). ?jitter=0 turns the jitter off. The page POSTs its numbers to /result: DIR/bench3-result.txt.
const http = require('node:http'), fs = require('node:fs'), path = require('node:path');
const [dir, port] = [process.argv[2], +process.argv[3]];
const wire = fs.readFileSync(path.join(dir, 'stream.pf'));
const msgs = [];
for (let p = 0; p + 4 <= wire.length;) { const n = wire.readUInt32LE(p); msgs.push(wire.subarray(p, p + 4 + n)); p += 4 + n; }
const files = { '/bench3.html': path.join(__dirname, 'bench3.html'), '/palfeed.js': path.join(__dirname, '../../internal/watch/web/palfeed.js') };
http.createServer((q, r) => {
  const u = new URL(q.url, 'http://x');
  if (q.method === 'POST' && u.pathname === '/result') { const c = []; q.on('data', (d) => c.push(d)); q.on('end', () => { fs.writeFileSync(path.join(dir, 'bench3-result.txt'), Buffer.concat(c)); r.end('ok'); }); return; }
  if (u.pathname === '/stream.pfeed') {
    r.writeHead(200, { 'Content-Type': 'application/x-palfeed', 'Cache-Control': 'no-store' });
    const jitter = u.searchParams.get('jitter') !== '0', t0 = process.hrtime.bigint();
    const now = () => Number(process.hrtime.bigint() - t0) / 1e6;
    let k = 0, prev = 0, seed = 7;
    const rnd = () => (seed = (seed * 1103515245 + 12345) & 0x7fffffff) / 0x7fffffff;
    (function next() {
      if (k >= msgs.length || r.destroyed) return r.end();
      const at = Math.max(prev, k * 40 + (jitter ? rnd() * 25 + (rnd() < 0.03 ? 80 : 0) : 0));
      prev = at;
      setTimeout(() => { if (!r.destroyed) r.write(msgs[k]); k++; next(); }, Math.max(0, at - now()));
    })();
    q.on('close', () => r.destroy());
    return;
  }
  const f = files[u.pathname]; if (!f) { r.writeHead(404); return r.end(); }
  r.writeHead(200, { 'Content-Type': f.endsWith('.js') ? 'text/javascript' : 'text/html' }); fs.createReadStream(f).pipe(r);
}).listen(port, '127.0.0.1');
