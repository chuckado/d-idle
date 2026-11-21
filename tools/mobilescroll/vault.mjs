// Vault page on a phone (card mobilequests). usage: node vault.mjs URL [wxh]
// 1. one finger drag scrolls the document and the LAST grid item can be brought into view, no inner scroll box holds the grid;
// 2. the quick filter chips: counts add up to the page's total, a chip narrows the grid to its count, All brings it back, the pick survives a reload.
import { spawn } from 'node:child_process';
import fs from 'node:fs';
const CH = process.env.CHROME || 'google-chrome-stable';
const [url, size = '412x915'] = process.argv.slice(2);
const [W, H] = size.split('x').map(Number);
const dir = fs.mkdtempSync((process.env.TMPDIR || '/tmp') + '/vscroll-');
const port = 9400 + (process.pid % 50);
const ch = spawn(CH, ['--headless=new', '--remote-debugging-port=' + port, '--user-data-dir=' + dir, '--no-sandbox', '--disable-gpu', '--hide-scrollbars', 'about:blank'], { stdio: 'ignore' });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
for (let i = 0; i < 60; i++) { try { await fetch(`http://127.0.0.1:${port}/json`); break; } catch { await sleep(200); } }
const t = await (await fetch(`http://127.0.0.1:${port}/json/new?about:blank`, { method: 'PUT' })).json();
const ws = new WebSocket(t.webSocketDebuggerUrl); await new Promise((r) => (ws.onopen = r));
let id = 0; const pend = new Map();
ws.onmessage = (m) => { const d = JSON.parse(m.data); if (d.id && pend.has(d.id)) { pend.get(d.id)(d.result || d.error); pend.delete(d.id); } };
const send = (method, params = {}) => new Promise((r) => { const i = ++id; pend.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
const ev = async (expr) => { const r = await send('Runtime.evaluate', { expression: expr, returnByValue: true, awaitPromise: true }); return r.result ? r.result.value : r; };
await send('Page.enable'); await send('Page.setBypassCSP', { enabled: true });
await send('Emulation.setDeviceMetricsOverride', { width: W, height: H, deviceScaleFactor: 1, mobile: true });
await send('Emulation.setTouchEmulationEnabled', { enabled: true, maxTouchPoints: 5 });
if (process.env.OLD) { const css = '.vt,.vt-page{block-size:100dvh!important}.vt-main{overflow-y:auto!important;min-block-size:0!important}.vt-slots{display:none!important}'; await send('Page.addScriptToEvaluateOnNewDocument', { source: `document.addEventListener('DOMContentLoaded', () => { const s = document.createElement('style'); s.textContent = ${JSON.stringify(css)}; document.head.appendChild(s); });` }); }
const fails = [];
const sep = url.includes('?') ? '&' : '?';
async function load(theme) { await send('Page.navigate', { url: `${url}${sep}theme=${theme}` }); await sleep(3500); }
const swipe = async (y0, dist) => { await send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x: Math.round(W / 2), y: y0 }] }); for (let i = 1; i <= 12; i++) { await send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x: Math.round(W / 2), y: y0 - Math.round((dist * i) / 12) }] }); await sleep(16); } await send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] }); await sleep(500); };
await ev('try { localStorage.clear(); } catch (e) {}');
for (const theme of ['dark', 'light']) {
  await load(theme);
  const n = await ev(`document.querySelectorAll('#pack > *').length`);
  if (!n) { fails.push(`${theme}: the grid is empty`); continue; }
  const inner = await ev(`(() => { const o = []; for (let e = document.querySelector('#pack'); e && e !== document.body; e = e.parentElement) { const c = getComputedStyle(e); if (/^(auto|scroll)$/.test(c.overflowY) && e.scrollHeight > e.clientHeight + 1) o.push(e.className + ' ' + e.clientHeight + '/' + e.scrollHeight); } return o; })()`);
  if (inner.length) fails.push(`${theme}: the grid sits in an inner scroll box: ${inner.join('; ')}`);
  const sh = await ev('document.scrollingElement.scrollHeight');
  if (sh <= H + 2) fails.push(`${theme}: the page is not taller than the screen (${sh}); the fixture needs more items`);
  for (let i = 0; i < 40; i++) { const before = await ev('scrollY'); await swipe(Math.round(H * 0.8), 500); if ((await ev('scrollY')) === before) { if (i === 0) fails.push(`${theme}: a finger drag does not scroll the page (scrollY ${before})`); break; } }
  const vis = await ev(`(() => { const g = document.querySelectorAll('#pack > *'); const r = g[g.length - 1].getBoundingClientRect(); return r.bottom <= innerHeight + 1 && r.top >= 0 ? 1 : 0; })()`);
  if (!vis && sh > H + 2) fails.push(`${theme}: the last grid item cannot be brought into view`);
  console.log(`${theme}: ${n} cells, scrollH ${sh}, last visible ${vis}`);
}
// chips (dark): counts, filter, All, persistence
await load('dark');
const chips = await ev(`[...document.querySelectorAll('.vt-slot')].map((b) => [b.dataset.slot, b.textContent, b.getBoundingClientRect().height])`);
if (!Array.isArray(chips) || chips.length < 2) fails.push('no quick filter chips');
else {
  const num = (s) => Number((/(\d+)$/.exec(s) || [0, -1])[1]);
  const total = await ev(`(window.__vt && 0) || Number((/(\\d+)/.exec(document.querySelector('.vt-shown') ? document.querySelector('.vt-shown').textContent : '') || [0, -1])[1])`);
  const slots = chips.filter((c) => c[0]);
  if (chips[0][0] !== '' ) fails.push('the first chip is not All');
  for (const c of chips) if (c[2] < 43.5) fails.push(`chip ${c[1]} is ${c[2]} px high (< 44)`);
  const all = await ev(`document.querySelectorAll('#pack > *').length`);
  const pick = slots.find((c) => num(c[1]) > 0 && num(c[1]) < all) || slots[0];
  await ev(`document.querySelector('.vt-slot[data-slot="${pick[0]}"]').click()`); await sleep(1500);
  const shown = await ev(`document.querySelectorAll('#pack > *').length`);
  if (shown !== num(pick[1])) fails.push(`chip ${pick[1]}: the grid shows ${shown} items, the chip says ${num(pick[1])}`);
  await ev('location.reload()'); await sleep(3500);
  const pressed = await ev(`document.querySelector('.vt-slot[aria-pressed=true]') && document.querySelector('.vt-slot[aria-pressed=true]').dataset.slot`);
  if (pressed !== pick[0]) fails.push(`the chosen chip did not survive a reload (${pressed})`);
  await ev(`document.querySelector('.vt-slot[data-slot=""]').click()`); await sleep(1500);
  const back = await ev(`document.querySelectorAll('#pack > *').length`);
  if (back !== all) fails.push(`All shows ${back} items, the page had ${all}`);
  console.log('chips', JSON.stringify(chips.map((c) => c[1])), 'picked', pick[1], 'shown', shown, 'all', all, 'total', total);
}
ch.kill();
if (fails.length) { console.log('FAIL\n' + fails.join('\n')); process.exit(1); }
console.log('ok'); process.exit(0);
