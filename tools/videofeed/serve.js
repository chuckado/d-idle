// node tools/videofeed/serve.js DIR PORT: serves bench.html (the page POSTs its numbers to /result: DIR/bench-result.txt), palfeed.js and DIR/stream.pf on 127.0.0.1 for the browser bench
const http = require('node:http'), fs = require('node:fs'), path = require('node:path');
const [dir, port] = [process.argv[2], +process.argv[3]];
const files = { '/bench.html': path.join(__dirname, 'bench.html'), '/palfeed.js': path.join(__dirname, '../../internal/watch/web/palfeed.js'), '/stream.pf': path.join(dir, 'stream.pf') };
http.createServer((q, r) => {
  if (q.method === 'POST' && q.url === '/result') { const c = []; q.on('data', (d) => c.push(d)); q.on('end', () => { fs.writeFileSync(path.join(dir, 'bench-result.txt'), Buffer.concat(c)); r.end('ok'); }); return; }
  const f = files[q.url]; if (!f) { r.writeHead(404); return r.end(); } r.writeHead(200, { 'Content-Type': q.url.endsWith('.js') ? 'text/javascript' : q.url.endsWith('.html') ? 'text/html' : 'application/octet-stream' }); fs.createReadStream(f).pipe(r); }).listen(port, '127.0.0.1');
