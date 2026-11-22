// Vault page on a desktop window (card vaultdesktop). usage: node vault.mjs BASEURL "1366x768 1920x1080 ..."
// For every size and fixture state: the page ITSELF scrolls (mouse wheel and the End key change scrollY), no inner scroll box holds the
// grid, the LAST grid item can be brought into view, a wide window shows 6+ columns, and the Reveal block is a one-liner that unfolds and
// remembers its state. OLD=1 puts the 100dvh frame back (the check must then fail).
import { spawn } from 'node:child_process';
import fs from 'node:fs';
const CH = process.env.CHROME || 'google-chrome-stable';
const [base, sizes = '1366x768 1920x1080 1280x720 900x700'] = process.argv.slice(2);
const dir = fs.mkdtempSync((process.env.TMPDIR || '/tmp') + '/dscroll-');
const port = 9500 + (process.pid % 50);
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
if (process.env.OLD) { const css = '.vt,.vt-page{block-size:100dvh!important;min-block-size:0!important}.vt-body{overflow:hidden!important;min-block-size:0!important}.vt-main{overflow-y:auto!important;min-block-size:0!important}'; await send('Page.addScriptToEvaluateOnNewDocument', { source: `document.addEventListener('DOMContentLoaded', () => { const s = document.createElement('style'); s.textContent = ${JSON.stringify(css)}; document.head.appendChild(s); });` }); }
const fails = [];
const wheel = async (W, H, dy) => { for (let i = 0; i < 6; i++) { await send('Input.dispatchMouseEvent', { type: 'mouseWheel', x: Math.round(W / 2), y: Math.round(H * 0.6), deltaX: 0, deltaY: dy / 6 }); await sleep(30); } await sleep(300); };
const key = async (k, code, vk) => { for (const type of ['keyDown', 'keyUp']) await send('Input.dispatchKeyEvent', { type, key: k, code, windowsVirtualKeyCode: vk, ...(type === 'keyDown' && k === ' ' ? { text: ' ' } : {}) }); await sleep(500); };
async function load(state, extra = '') { await send('Page.navigate', { url: `${base}/watch/vault?vault=${state}&theme=dark${extra}` }); await sleep(3500); }
for (const size of sizes.split(/\s+/)) {
  const [W, H] = size.split('x').map(Number);
  await send('Emulation.setDeviceMetricsOverride', { width: W, height: H, deviceScaleFactor: 1, mobile: false });
  for (const state of ['pop', 'revealing', 'big']) {
    const tag = `${size} ${state}`;
    await ev('try { localStorage.clear(); } catch (e) {}');
    await load(state);
    const n = await ev(`document.querySelectorAll('#pack > *').length`);
    if (!n) { fails.push(`${tag}: the grid is empty`); continue; }
    const inner = await ev(`(() => { const o = []; for (let e = document.querySelector('#pack'); e && e !== document.body; e = e.parentElement) { const c = getComputedStyle(e); if (/^(auto|scroll)$/.test(c.overflowY) && e.scrollHeight > e.clientHeight + 1) o.push(e.className + ' ' + e.clientHeight + '/' + e.scrollHeight); } return o; })()`);
    if (inner.length) fails.push(`${tag}: the grid sits in an inner scroll box: ${inner.join('; ')}`);
    const sh = await ev('document.scrollingElement.scrollHeight');
    if (sh <= H + 2) { fails.push(`${tag}: the page is not taller than the window (${sh}); the fixture needs more items`); continue; }
    await wheel(W, H, 900);
    const y1 = await ev('scrollY');
    if (!(y1 > 0)) fails.push(`${tag}: the mouse wheel does not scroll the page (scrollY ${y1})`);
    await ev('scrollTo(0, 0); document.activeElement && document.activeElement.blur()'); await sleep(200);
    await key('End', 'End', 35);
    const y2 = await ev('scrollY');
    if (!(y2 > 0)) fails.push(`${tag}: the End key does not scroll the page (scrollY ${y2})`);
    await ev('scrollTo(0, 0)'); await sleep(200);
    await key(' ', 'Space', 32);
    if (!((await ev('scrollY')) > 0)) fails.push(`${tag}: the Space key does not scroll the page`);
    await ev('scrollTo(0, document.scrollingElement.scrollHeight)'); await sleep(400);
    const vis = await ev(`(() => { const g = document.querySelectorAll('#pack > *'); const r = g[g.length - 1].getBoundingClientRect(); return r.bottom <= innerHeight + 1 && r.top >= 0 ? 1 : 0; })()`);
    const bottom = await ev('Math.round(scrollY + innerHeight) >= document.scrollingElement.scrollHeight - 2 ? 1 : 0');
    if (!vis && state !== 'big') fails.push(`${tag}: the last grid item cannot be brought into view`); // 'big' is windowed and loads more at the end: there only the bottom of the page counts
    if (!bottom && state === 'big' && !((await ev('document.scrollingElement.scrollHeight')) > sh)) fails.push(`${tag}: the end of the page cannot be reached`);
    if (state === 'pop' && W >= 1600) {
      await ev('scrollTo(0, 0)');
      const cols = await ev(`new Set([...document.querySelectorAll('#pack > *')].map((c) => Math.round(c.getBoundingClientRect().left))).size`);
      if (cols < 6) fails.push(`${tag}: only ${cols} grid columns on a ${W} px window`);
      console.log(`${tag}: ${cols} columns`);
    }
    console.log(`${tag}: ${n} cells, scrollH ${sh}, wheel ${y1}, End ${y2}, last visible ${vis}`);
  }
}
// the folds (default 1366x768): the Reveal block is a one-liner until opened; the choice survives a reload
await send('Emulation.setDeviceMetricsOverride', { width: 1366, height: 768, deviceScaleFactor: 1, mobile: false });
await ev('try { localStorage.clear(); } catch (e) {}');
await load('revealing', '');
const fold = `document.querySelector('.vt-revealbar .vt-fold')`;
if (!(await ev(`!!${fold}`))) fails.push('the Reveal bar has no fold button (.vt-fold)');
else {
  const optsVis = () => ev(`(() => { const o = document.querySelector('.vt-rb-opts'); return !!o && o.getBoundingClientRect().height > 0; })()`);
  const barH = await ev(`document.querySelector('.vt-revealbar').getBoundingClientRect().height`);
  console.log('reveal bar folded:', !(await optsVis()), 'height', Math.round(barH));
  if (barH > 60) fails.push(`the folded Reveal bar is ${Math.round(barH)} px high (> 60)`);
  await ev(`${fold}.click()`); await sleep(400);
  if (!(await optsVis())) fails.push('the fold button does not open the payment options');
  await load('revealing', '');
  if (!(await optsVis())) fails.push('the open state of the Reveal bar did not survive a reload');
  await ev(`${fold}.click()`); await sleep(300); await load('revealing', '');
  if (await optsVis()) fails.push('the closed state of the Reveal bar did not survive a reload');
}
ch.kill();
if (fails.length) { console.log('FAIL\n' + fails.join('\n')); process.exit(1); }
console.log('ok'); process.exit(0);
