// usage: node shots.mjs BASE OUTDIR TAG   pictures at 412x915 of the Quests tab and the Vault page, dark and light; OLD=1 puts the pre-fix CSS back.
// Also lists inner scroll boxes of the other phone pages (Skills, Points, Return, Inventory) so a scroll trap shows up in the log.
import { spawn } from 'node:child_process';
import fs from 'node:fs';
const [base, out, tag] = process.argv.slice(2);
const dir = fs.mkdtempSync((process.env.TMPDIR || '/tmp') + '/shots-');
const port = 9460 + (process.pid % 50);
const ch = spawn(process.env.CHROME || 'google-chrome-stable', ['--headless=new', '--remote-debugging-port=' + port, '--user-data-dir=' + dir, '--no-sandbox', '--disable-gpu', '--hide-scrollbars', 'about:blank'], { stdio: 'ignore' });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
for (let i = 0; i < 60; i++) { try { await fetch(`http://127.0.0.1:${port}/json`); break; } catch { await sleep(200); } }
const t = await (await fetch(`http://127.0.0.1:${port}/json/new?about:blank`, { method: 'PUT' })).json();
const ws = new WebSocket(t.webSocketDebuggerUrl); await new Promise((r) => (ws.onopen = r));
let id = 0; const pend = new Map();
ws.onmessage = (m) => { const d = JSON.parse(m.data); if (d.id && pend.has(d.id)) { pend.get(d.id)(d.result || d.error); pend.delete(d.id); } };
const send = (method, params = {}) => new Promise((r) => { const i = ++id; pend.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
const ev = async (e) => (await send('Runtime.evaluate', { expression: e, returnByValue: true })).result.value;
await send('Page.enable'); await send('Page.setBypassCSP', { enabled: true });
await send('Emulation.setDeviceMetricsOverride', { width: 412, height: 915, deviceScaleFactor: 1, mobile: true });
await send('Emulation.setTouchEmulationEnabled', { enabled: true, maxTouchPoints: 5 });
if (process.env.OLD) { const css = '.d-shell[data-template=cockpit]{block-size:100dvh!important;overflow:hidden!important}.qb-list{max-block-size:266px!important;overflow-y:auto!important}.vt,.vt-page{block-size:100dvh!important}.vt-main{overflow-y:auto!important;min-block-size:0!important}.vt-slots{display:none!important}'; await send('Page.addScriptToEvaluateOnNewDocument', { source: `document.addEventListener('DOMContentLoaded', () => { const s = document.createElement('style'); s.textContent = ${JSON.stringify(css)}; document.head.appendChild(s); });` }); }
fs.mkdirSync(out, { recursive: true });
const pages = { quests: '/watch?first=0&needs=0&quests=running&tab=Quests', vault: '/watch/vault?vault=pop' };
for (const [n, p] of Object.entries(pages)) for (const th of ['dark', 'light']) {
  await send('Page.navigate', { url: base + p + '&theme=' + th }); await sleep(3500);
  const r = await send('Page.captureScreenshot', { format: 'png' });
  fs.writeFileSync(`${out}/mobile-${n}-${th}-${tag}.png`, Buffer.from(r.data, 'base64'));
}
for (const p of ['/watch/skills', '/watch/points', '/watch/return', '/watch/inventory', '/watch/characters']) {
  await send('Page.navigate', { url: base + p + '?theme=dark' }); await sleep(3000);
  console.log(p, await ev(`JSON.stringify({sh: document.scrollingElement.scrollHeight, vh: innerHeight, sw: document.scrollingElement.scrollWidth, inner: [...document.querySelectorAll('body *')].filter((n) => /^(auto|scroll)$/.test(getComputedStyle(n).overflowY) && n.scrollHeight > n.clientHeight + 1 && n.clientHeight > 0).map((n) => (n.className || n.tagName) + ' ' + n.clientHeight + '/' + n.scrollHeight).slice(0, 6), wide: [...document.querySelectorAll('body *')].filter((n) => n.getBoundingClientRect().width > 500).slice(0, 4).map((n) => (n.id || n.className || n.tagName) + ' ' + Math.round(n.getBoundingClientRect().right)), bodyOverflow: getComputedStyle(document.body).overflowY})`));
}
ch.kill(); process.exit(0);
