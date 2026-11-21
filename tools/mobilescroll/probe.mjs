// Phone scroll probe (card mobilequests). usage: node probe.mjs BASEURL [wxh]
// For each phone tab a one-finger drag that starts over the tab's content must scroll the PAGE (window.scrollY changes), no
// content box may be a nested scroller the finger can get stuck in, and the page must not scroll sideways. Exit 1 on a failure.
import { spawn } from 'node:child_process';
import fs from 'node:fs';
const CH = process.env.CHROME || 'google-chrome-stable';
const base = process.argv[2];
const [W, H] = (process.argv[3] || '412x915').split('x').map(Number);
const dir = fs.mkdtempSync((process.env.TMPDIR || '/tmp') + '/mscroll-');
const port = 9340 + (process.pid % 50);
const ch = spawn(CH, ['--headless=new', '--remote-debugging-port=' + port, '--user-data-dir=' + dir, '--no-sandbox', '--disable-gpu', '--hide-scrollbars', 'about:blank'], { stdio: 'ignore' });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function up() { for (let i = 0; i < 60; i++) { try { return await (await fetch(`http://127.0.0.1:${port}/json`)).json(); } catch { await sleep(200); } } throw new Error('no chrome'); }
await up();
const t = await (await fetch(`http://127.0.0.1:${port}/json/new?about:blank`, { method: 'PUT' })).json();
const ws = new WebSocket(t.webSocketDebuggerUrl); await new Promise((r) => (ws.onopen = r));
let id = 0; const pend = new Map();
ws.onmessage = (m) => { const d = JSON.parse(m.data); if (d.id && pend.has(d.id)) { pend.get(d.id)(d.result || d.error); pend.delete(d.id); } };
const send = (method, params = {}) => new Promise((r) => { const i = ++id; pend.set(i, r); ws.send(JSON.stringify({ id: i, method, params })); });
const ev = async (expr) => { const r = await send('Runtime.evaluate', { expression: expr, returnByValue: true, awaitPromise: true }); return r.result ? r.result.value : r; };
await send('Page.enable');
await send('Emulation.setDeviceMetricsOverride', { width: W, height: H, deviceScaleFactor: 1, mobile: true });
await send('Emulation.setTouchEmulationEnabled', { enabled: true, maxTouchPoints: 5 });
const fails = [];
if (process.env.OLD) { // the CSS before the fix, put back: a fixed frame and a small scroll box around the quest list
  await send('Page.setBypassCSP', { enabled: true });
  const css = '.d-shell[data-template=cockpit]{block-size:100dvh!important;overflow:hidden!important}.qb-list{max-block-size:266px!important;overflow-y:auto!important}';
  await send('Page.addScriptToEvaluateOnNewDocument', { source: `document.addEventListener('DOMContentLoaded', () => { const s = document.createElement('style'); s.textContent = ${JSON.stringify(css)}; document.head.appendChild(s); });` });
}
const tabs = (process.env.TABS || 'Orders,Map,Quests,Progress').split(',');
const sep = base.includes('?') ? '&' : '?';
for (const tab of tabs) {
  for (const theme of ['dark', 'light']) {
    await send('Page.navigate', { url: `${base}${sep}theme=${theme}&tab=${tab}` });
    await sleep(3500);
    const name = `${tab}/${theme}`;
    // nested scrollers inside the tab body: a box whose content is taller than it is and that scrolls by itself
    const info = await ev(`(() => { const b = document.querySelector('.d-slottabs'); if (!b) return { err: 'no tab body' };
      const bad = []; for (const n of b.querySelectorAll('*')) { const cs = getComputedStyle(n); if (/^(auto|scroll)$/.test(cs.overflowY) && n.scrollHeight > n.clientHeight + 1 && n.clientHeight > 0) bad.push((n.className||n.tagName)+' '+n.clientHeight+'/'+n.scrollHeight); }
      const r = b.getBoundingClientRect(); const se = document.scrollingElement;
      return { bad, x: Math.round(innerWidth/2), y: Math.round(Math.min(innerHeight-60, Math.max(r.top+40, innerHeight*0.7))), sh: se.scrollHeight, vh: innerHeight, sw: se.scrollWidth, cw: document.documentElement.clientWidth, scale: visualViewport.scale, vw: innerWidth, bodyBottom: Math.round(r.bottom), tab: document.querySelector('.d-tabbar [aria-selected=true],.d-tabbar [data-on]') ? 1 : 0 }; })()`);
    if (info.err) { fails.push(`${name}: ${info.err}`); continue; }
    if (info.sw > info.cw) { const w = await ev(`[...document.querySelectorAll('body *')].filter((n) => n.getBoundingClientRect().right > document.documentElement.clientWidth + 1 && n.getBoundingClientRect().width > 0).slice(0, 6).map((n) => (n.dataset.panel || n.className || n.tagName) + ' ' + Math.round(n.getBoundingClientRect().right)).join('; ')`); fails.push(`${name}: wider than the screen: ${w}`); if (process.env.DEBUG) console.log(await ev(`['.d-shell','.d-main','[data-cockpit-slot=stage]','[data-cockpit-slot=phoneTabs]','.d-slottabs','.d-slottabs > *','.d-band-status'].map((q) => { const n = document.querySelector(q); return n ? q + ' ' + Math.round(n.getBoundingClientRect().width) + ' ' + getComputedStyle(n).gridTemplateColumns : q + ' none'; }).join('\\n')`)); }
    if (info.sw > info.vw) fails.push(`${name}: page scrolls sideways (${info.sw} > ${info.vw})`);
    if (info.bad.length) fails.push(`${name}: nested scroller(s) with content cut: ${info.bad.join('; ')}`);
    const before = await ev('scrollY');
    if (info.sh > info.vh + 2) {
      // swipe up with one finger starting over the tab content
      const dist = Math.min(400, info.sh - info.vh), y0 = info.y;
      await send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x: info.x, y: y0 }] });
      for (let i = 1; i <= 12; i++) { await send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x: info.x, y: y0 - Math.round((dist * i) / 12) }] }); await sleep(16); }
      await send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
      await sleep(600);
      const after = await ev('scrollY');
      if (!(after > before) && process.env.DEBUG) console.log(await ev(`(() => { const o = []; for (let n = document.elementFromPoint(${info.x}, ${info.y}); n; n = n.parentElement) o.push((n.dataset && n.dataset.panel || n.className || n.tagName) + ':' + getComputedStyle(n).touchAction + ':' + getComputedStyle(n).overflowY); return o.join(' < '); })()`));
      if (!(after > before)) fails.push(`${name}: a one-finger drag over the content did not scroll the page (scrollY ${before} -> ${after}, scrollHeight ${info.sh}, viewport ${info.vh})`);
    } else if (tab === 'Quests') fails.push(`${name}: the Quests tab fits the screen (scrollHeight ${info.sh}); the fixture must be taller than the screen`);
    console.log(`${name}: scrollH ${info.sh} vh ${info.vh} vw ${info.vw} sw ${info.sw} cw ${info.cw} scale ${info.scale} y ${info.y} nested ${info.bad.length}`);
  }
}
ch.kill();
if (fails.length) { console.log('FAIL\n' + fails.join('\n')); process.exit(1); }
console.log('ok'); process.exit(0);
