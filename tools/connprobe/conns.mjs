// usage: node conns.mjs URL reloads waitMs [tabs]
import { spawn } from 'node:child_process';
import fs from 'node:fs';
const CH = process.env.CHROME || 'google-chrome-stable';
const [url, reloads = '0', waitMs = '30000', tabs = '1'] = process.argv.slice(2);
const dir = fs.mkdtempSync((process.env.TMPDIR || '/tmp') + '/connprobe-');
const ch = spawn(CH, ['--headless=new', '--remote-debugging-port=9333', '--user-data-dir=' + dir, '--no-sandbox', '--disable-gpu', 'about:blank'], { stdio: 'ignore' });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function targets() { for (let i = 0; i < 50; i++) { try { return await (await fetch('http://127.0.0.1:9333/json')).json(); } catch { await sleep(200); } } throw new Error('no chrome'); }
async function newTab() {
  const t = await (await fetch('http://127.0.0.1:9333/json/new?about:blank', { method: 'PUT' })).json();
  const ws = new WebSocket(t.webSocketDebuggerUrl); await new Promise((r) => ws.onopen = r);
  let id = 0; const pend = new Map(); const reqs = new Map();
  ws.onmessage = (m) => { const d = JSON.parse(m.data);
    if (d.id && pend.has(d.id)) { pend.get(d.id)(d.result); pend.delete(d.id); return; }
    const p = d.params;
    if (d.method === 'Network.requestWillBeSent') reqs.set(p.requestId, { url: p.request.url.replace(/^https?:\/\/[^/]+/, '').replace(/[?&]r=\d+/, ''), type: p.type, state: 'pending', t: Date.now() });
    else if (d.method === 'Network.loadingFinished' || d.method === 'Network.loadingFailed') { const r = reqs.get(p.requestId); if (r) r.state = d.method.slice(15); }
    else if (d.method === 'Network.responseReceived') { const r = reqs.get(p.requestId); if (r) r.resp = p.response.status; }
  };
  const send = (method, params = {}) => new Promise((r) => { const i = ++id; pend.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
  await send('Network.enable'); await send('Page.enable');
  return { send, reqs, ws };
}
function open(reqs, label) {
  const o = [...reqs.values()].filter((r) => r.state === 'pending' && Date.now() - r.t > 3000);
  const by = {}; for (const r of o) { const k = r.type + ' ' + r.url.split('?')[0]; by[k] = (by[k] || 0) + 1; }
  console.log(label, 'open:', o.length, JSON.stringify(by));
}
await targets();
const T = [];
for (let i = 0; i < Number(tabs); i++) { const t = await newTab(); T.push(t); await t.send('Page.navigate', { url }); }
await sleep(Number(waitMs));
T.forEach((t, i) => open(t.reqs, 'tab' + i + ' load'));
for (let n = 1; n <= Number(reloads); n++) {
  const t = T[0]; t.reqs.clear(); await t.send('Page.reload'); await sleep(Number(waitMs));
  open(t.reqs, 'after reload ' + n);
}
let tot = 0; for (const t of T) tot += [...t.reqs.values()].filter((r) => r.state === 'pending' && Date.now() - r.t > 3000).length;
console.log('total tabs open', tot);
ch.kill(); process.exit(0);
