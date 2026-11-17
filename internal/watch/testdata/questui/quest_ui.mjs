// The quest board UI (slice S5): the state-to-words mapping, the button each state offers, the Hand in read-back and the refusal
// text, with a fake API fed by the goldens of GET /api/quests (testdata/quests) and a stub DOM. Usage: node quest_ui.mjs web/panels
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

const dir = path.resolve(process.argv[2]);
const gold = (n) => JSON.parse(fs.readFileSync(path.join(dir, '..', '..', 'testdata', 'quests', n + '.json'), 'utf8'));
const B = await import(pathToFileURL(path.join(dir, 'quest-board/panel.js')));
const S = await import(pathToFileURL(path.join(dir, 'quest-strip/panel.js')));
const L = await import(pathToFileURL(path.join(dir, 'quest-list/panel.js')));

// ---- stub DOM and ctx ----
function node(tag, cls, text) {
  const listeners = {}, attrs = {};
  const e = { tag, className: cls || '', dataset: {}, children: [], hidden: false, disabled: false, title: '', attrs, listeners,
    setAttribute(k, v) { attrs[k] = String(v); }, getAttribute(k) { return attrs[k]; },
    addEventListener(t, f) { (listeners[t] = listeners[t] || []).push(f); },
    click() { (listeners.click || []).forEach((f) => f({})); },
    append(...k) { for (const x of k) e.children.push(typeof x === 'string' ? { tag: '#text', textContent: x, children: [], dataset: {}, attrs: {} } : x); },
    appendChild(k) { e.children.push(k); return k; } };
  let own = text === undefined ? '' : String(text);
  Object.defineProperty(e, 'textContent', { get: () => own + e.children.map((c) => c.textContent).join(''), set: (v) => { own = String(v); e.children.length = 0; }, enumerable: true });
  return e;
}
const walk = (n, f) => { f(n); n.children.forEach((c) => walk(c, f)); };
const all = (n, pred) => { const o = []; walk(n, (x) => { if (pred(x)) o.push(x); }); return o; };
const byKey = (n, k) => all(n, (x) => x.dataset && x.dataset.key === k)[0];
const texts = (n) => n.textContent;

function ctxFor(posts, reply) {
  const body = node('div', 'd-slotbody');
  const c = { body, hidden: null, st: '', toasts: [], refreshed: 0, drawers: [], timers: [],
    el: node, ui: { clear(n) { n.children.length = 0; } },
    hide(on) { c.hidden = !!on; }, state(s) { c.st = s; }, toast(sev, t) { c.toasts.push([sev, t]); },
    later(fn, ms) { c.timers.push([fn, ms]); }, refresh() { c.refreshed++; }, openDrawer(id) { c.drawers.push(id); return true; },
    post(p, b) { posts.push([p, b]); return Promise.resolve(typeof reply === 'function' ? reply(p, b) : reply); } };
  return c;
}
const tick = () => new Promise((r) => setImmediate(r));
const clone = (v) => JSON.parse(JSON.stringify(v));
const fresh = () => { B.resetShared(); return B.shared; };

// ---- the words for every state ----
for (const cls of ['paladin', 'sorceress']) {
  const L0 = fresh();
  const offers = gold('offers-' + cls), running = gold('running-' + cls), awaiting = gold('awaiting-' + cls);
  let m = B.stripModel(offers, L0);
  assert.equal(m.kind, 'offer', cls);
  assert.equal(m.head, 'Offer: Den of Evil, safe');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.quest, m.primary.chunk], ['Start', 'start', 'den_of_evil', 'clear_den']);
  m = B.stripModel(running, L0);
  assert.equal(m.kind, 'running');
  assert.equal(m.head, 'Running: Den of Evil');
  assert.equal(m.sub, 'Walk blood moor, next: Akara');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.confirm], ['Stop', 'cancel', true], 'a stop asks first');
  m = B.stripModel(awaiting, L0);
  assert.equal(m.kind, 'awaiting');
  assert.equal(m.head, 'Akara is waiting: hand in Den of Evil');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.chunk], ['Hand in', 'handin', 'clear_den']);
  assert.equal(m.secondary, null, 'one primary button at a time');
}

// the board is off: the exact knob, no button
{
  const m = B.stripModel({ schema: 'watch.quests/1', enabled: false, reason: 'the quest board is off (knob quests.board is 0)' }, fresh());
  assert.equal(m.kind, 'off');
  assert.match(m.sub, /Turn on the Quest board in settings: set quests\.board to 1/);
  assert.equal(m.primary, null);
  assert.equal(B.stripModel(null, fresh()).kind, 'gone');
}

// no startable offer: training words from the first reason, never a number
{
  const v = gold('offers-paladin');
  v.offers.forEach((o) => { o.startable = false; o.readiness = 'too_hard'; o.reasons = [{ code: 'level', text: 'She would likely kill you. Train first' }]; });
  let m = B.stripModel(v, fresh());
  assert.equal(m.kind, 'training');
  assert.equal(m.head, 'Training toward Den of Evil');
  assert.equal(m.sub, 'She would likely kill you. Train first');
  assert.equal(m.primary, null);
  v.mode = 'idle';
  m = B.stripModel(v, fresh());
  assert.equal(m.kind, 'idle');
  assert.equal(m.head, 'Idle in town');
  v.mode = 'farm'; m = B.stripModel(v, fresh());
  assert.equal(m.head, 'Farming before Den of Evil');
}

// a paused hero: Start is shown but cannot be pressed, with the reason
{
  const v = gold('offers-paladin'); v.paused = true;
  const m = B.stripModel(v, fresh());
  assert.equal(m.primary.disabled, true);
  assert.match(m.primary.why, /resume/i);
}

// a stuck chunk: Try again first, Stop second (with a confirm), the board's reason in the tile
{
  const v = gold('running-paladin');
  v.blocked = Object.assign({}, v.running, { state: 'blocked', why: 'could not find the Cairn Stone' }); v.running = null;
  const m = B.stripModel(v, fresh());
  assert.equal(m.kind, 'blocked');
  assert.equal(m.sub, 'could not find the Cairn Stone');
  assert.deepEqual([m.primary.label, m.primary.verb, m.secondary.label, m.secondary.confirm], ['Try again', 'retry', 'Stop', true]);
}

// ---- the Hand in moment, from the GET only ----
{
  const L0 = fresh();
  const aw = gold('awaiting-paladin');
  L0.handing = { quest: 'den_of_evil', chunk: 'clear_den', at: Date.now() };
  let m = B.stripModel(aw, L0);
  assert.equal(m.kind, 'handingin', 'pressed: Handing in at once, before the next GET');
  assert.equal(m.head, 'Handing in Den of Evil...');
  assert.equal(m.primary.disabled, true);
  // the GET says the dialogue runs / is read back
  const run = clone(aw); run.handIn = { quest: 'den_of_evil', chunk: 'clear_den', state: 'running' };
  assert.equal(B.stripModel(run, fresh()).kind, 'handingin', 'a hand-in started elsewhere (Needs you row) shows too');
  const conf = clone(aw); conf.handIn = { quest: 'den_of_evil', chunk: 'clear_den', state: 'confirming' };
  m = B.stripModel(conf, fresh());
  assert.equal(m.kind, 'confirming');
  assert.match(m.head, /Checking that the game agrees/);
  // the GET says it failed: the API's own reason and a Try again
  const bad = clone(aw); bad.handIn = { quest: 'den_of_evil', chunk: 'clear_den', state: 'failed', reason: 'Akara did not answer: the dialogue never opened' };
  L0.handing = { quest: 'den_of_evil', chunk: 'clear_den', at: Date.now() };
  m = B.stripModel(bad, L0);
  assert.equal(m.kind, 'failed');
  assert.equal(m.sub, 'Akara did not answer: the dialogue never opened');
  assert.equal(m.primary.label, 'Try again');
  // a reply that refused (nothing in the GET yet) keeps the reason on the tile until the next press
  const L1 = fresh();
  L1.error = { quest: 'den_of_evil', reason: 'the hero is not at Akara yet' };
  m = B.stripModel(aw, L1);
  assert.equal(m.kind, 'failed');
  assert.equal(m.sub, 'the hero is not at Akara yet');
  // a hand-in lost for six minutes is not shown for ever
  const L2 = fresh();
  L2.handing = { quest: 'den_of_evil', chunk: 'clear_den', at: Date.now() - 7 * 60 * 1000 };
  assert.equal(B.stripModel(aw, L2).kind, 'awaiting');
}

// the payoff: announced once, only when the GET says the chunk is done
{
  const L0 = fresh();
  assert.equal(B.settle(gold('running-paladin'), L0), null);
  assert.equal(B.settle(gold('awaiting-paladin'), L0), null);
  const done = gold('offers-paladin');
  done.quests[0].state = 'done'; done.quests[0].chunks[0].state = 'done'; done.offers.shift();
  const p = B.settle(done, L0);
  assert.deepEqual(p, { title: 'Den of Evil', text: 'Done: Den of Evil' });
  assert.equal(B.settle(done, L0), null, 'once');
  assert.equal(B.stripModel(done, L0).kind, 'offer', 'the next offer shows at once');
  assert.match(B.stripModel(done, L0).head, /Blood Raven|Sisters/);
  // a cancelled chunk is not a payoff
  const L1 = fresh();
  B.settle(gold('running-paladin'), L1);
  assert.equal(B.settle(gold('offers-paladin'), L1), null);
  // a payout shows only when the API says it paid
  const L3 = fresh();
  const aw = gold('awaiting-paladin'); B.settle(aw, L3);
  const d2 = clone(done); d2.handIn = { quest: 'den_of_evil', chunk: 'clear_den', state: 'done', pays: { meta: 5 } };
  assert.equal(B.settle(d2, L3).text, 'Done: Den of Evil   +5 meta');
  // the page never says a number for a margin or a chapter
  for (const v of [gold('offers-paladin'), gold('running-paladin'), gold('awaiting-paladin')]) {
    const m = B.stripModel(v, fresh());
    assert.doesNotMatch(m.head + ' ' + m.sub, /chapter\s*\d|margin\D{0,8}\d|\d+\s*%/i);
  }
}

// ---- the call: a refusal shows the API's reason; success asks for a fresh GET ----
{
  const L0 = fresh();
  const posts = [];
  let c = ctxFor(posts, { ok: false, code: 'busy', reason: 'Den of Evil is running; stop it first' });
  await B.act(c, 'start', 'sisters_burial_grounds', 'blood_raven', L0);
  assert.deepEqual(posts[0], ['/api/quests/sisters_burial_grounds/start', { chunk: 'blood_raven' }]);
  assert.deepEqual(c.toasts, [['warn', 'Den of Evil is running; stop it first']]);
  c = ctxFor([], { ok: false, code: 'weird' });
  await B.act(c, 'start', 'a', 'b', L0);
  assert.match(c.toasts[0][1], /code weird/, 'a refusal with no reason still names its code, never "could not do that"');
  assert.doesNotMatch(c.toasts[0][1], /could not do that/);
  c = ctxFor([], { ok: false, code: 'not_awaiting', reason: 'the hero is not at Akara yet' });
  await B.act(c, 'handin', 'den_of_evil', 'clear_den', L0);
  assert.equal(L0.handing, null);
  assert.equal(L0.error.reason, 'the hero is not at Akara yet');
  c = ctxFor(posts, { ok: true, pending: true });
  L0.error = null;
  await B.act(c, 'handin', 'den_of_evil', 'clear_den', L0);
  assert.ok(L0.handing, 'accepted: Handing in until the GET says so');
  assert.equal(c.refreshed, 1);
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/handin', { chunk: 'clear_den' }]);
  await B.act(c, 'mode', '', 'farm', L0);
  assert.deepEqual(posts.at(-1), ['/api/quests/mode', { mode: 'farm' }]);
  await B.act(c, 'cancel', 'den_of_evil', '', L0);
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/cancel', {}]);
}

// ---- the strip tile: buttons through a stub DOM ----
{
  const L0 = fresh(), posts = [];
  const c = ctxFor(posts, { ok: true });
  const s = S.mount(node('div', 'd-panel'), c);
  s.update(gold('offers-paladin'));
  assert.equal(c.hidden, false);
  const btn = byKey(c.body, 'primary');
  assert.equal(texts(btn), 'Start');
  assert.equal(btn.disabled, false);
  assert.match(texts(c.body), /Offer: Den of Evil, safe/);
  assert.equal(all(c.body, (x) => x.dataset.key && x.dataset.key.startsWith('mode-')).length, 3, 'train, idle, farm');
  assert.equal(all(c.body, (x) => x.dataset.key === 'mode-train')[0].attrs['aria-checked'], 'true', 'Train is the default');
  btn.click(); await tick();
  assert.deepEqual(posts[0], ['/api/quests/den_of_evil/start', { chunk: 'clear_den' }]);
  byKey(c.body, 'all').click();
  assert.deepEqual(c.drawers, ['quest-board']);
  byKey(c.body, 'mode-farm').click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/mode', { mode: 'farm' }]);
  // running: Stop needs a second press
  s.update(gold('running-paladin'));
  let stop = byKey(c.body, 'primary');
  assert.equal(texts(stop), 'Stop');
  const n = posts.length;
  stop.click(); await tick();
  assert.equal(posts.length, n, 'the first press only asks');
  stop = byKey(c.body, 'primary');
  assert.equal(texts(stop), 'Stop it?');
  stop.click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/cancel', { chunk: 'clear_den' }]);
  // waiting at the NPC: Hand in, then Handing in... at once, and the reply never decides
  s.update(gold('awaiting-paladin'));
  const hand = byKey(c.body, 'primary');
  assert.equal(texts(hand), 'Hand in');
  hand.click();
  assert.match(texts(c.body), /Handing in Den of Evil\.\.\./, 'the tile changes before any reply');
  await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/handin', { chunk: 'clear_den' }]);
  assert.equal(texts(byKey(c.body, 'primary')), 'Handing in...');
  assert.equal(byKey(c.body, 'primary').disabled, true);
  // the GET says done: the toast fires once, the tile shows the next offer
  const done = gold('offers-paladin');
  done.quests[0].state = 'done'; done.quests[0].chunks[0].state = 'done'; done.offers.shift();
  s.update(done);
  assert.deepEqual(c.toasts.filter((t) => t[0] === 'good'), [['good', 'Done: Den of Evil']]);
  s.update(done);
  assert.equal(c.toasts.filter((t) => t[0] === 'good').length, 1);
  assert.match(texts(c.body), /Offer: /);
  // off: the tile hides itself and leaves the Journey tile alone
  s.update({ enabled: false, reason: 'off' });
  assert.equal(c.hidden, true);
}

// ---- the board sheet ----
{
  fresh();
  const posts = [];
  const c = ctxFor(posts, { ok: true });
  const b = B.mount(node('div', 'd-panel'), c);
  const v = gold('offers-paladin');
  b.update(v);
  const t = texts(c.body);
  assert.match(t, /Act I, Normal/);
  assert.match(t, /Den of Evil/);
  assert.match(t, /Sisters' Burial Grounds/);
  assert.match(t, /The Search for Cain/);
  assert.match(t, /3 parts|4 parts/, 'a quest of parts says so');
  assert.match(t, /Radament's Lair/);
  assert.match(t, /Act II/, 'grouped by act');
  assert.match(t, /safe/);
  assert.match(t, /risky/);
  assert.match(t, /too hard/);
  assert.match(t, /auto: coming/, 'the auto toggle is a stub until S8');
  assert.ok(!byKey(c.body, 'auto-den_of_evil'), 'no live toggle while the API says unavailable');
  const rows = all(c.body, (x) => x.className === 'qb-row').map((x) => x.dataset.group);
  assert.ok(rows.indexOf('ready') < rows.indexOf('available') && rows.indexOf('available') < rows.indexOf('locked'), 'ready, then not yet, then locked');
  // every part has its own state; Start only where ready
  const starts = all(c.body, (x) => x.dataset && /^start-/.test(x.dataset.key || ''));
  assert.deepEqual(starts.map((x) => x.dataset.key), ['start-den_of_evil-clear_den', 'start-sisters_burial_grounds-blood_raven']);
  byKey(c.body, 'start-sisters_burial_grounds-blood_raven').click(); await tick();
  assert.deepEqual(posts[0], ['/api/quests/sisters_burial_grounds/start', { chunk: 'blood_raven' }]);
  byKey(c.body, 'mode-idle').click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/mode', { mode: 'idle' }]);
  // the auto toggle comes alive when the API offers it
  const v2 = clone(v); v2.autoStart = { available: true, note: '' };
  b.update(v2);
  const sw = byKey(c.body, 'auto-den_of_evil');
  assert.equal(sw.attrs.role, 'switch');
  assert.equal(sw.attrs['aria-checked'], 'false');
  // running: the Now row, with the stop confirm
  b.update(gold('running-paladin'));
  assert.match(texts(c.body), /Running: Den of Evil/);
  assert.equal(texts(byKey(c.body, 'now-cancel')), 'Stop');
  // history
  const h = gold('running-paladin'); h.history = [{ quest: 'x', chunk: 'y', title: 'Den of Evil', confirmed: true }];
  b.update(h);
  assert.match(texts(c.body), /Recent/);
  assert.match(texts(c.body), /handed in/);
  // off: the empty state names the knob
  b.update({ schema: 'watch.quests/1', enabled: false, reason: 'the quest board is off (knob quests.board is 0)' });
  assert.match(texts(c.body), /Turn on the Quest board in settings/);
  assert.match(texts(c.body), /quests\.board = 1/);
  assert.equal(c.st, 'ready');
  // no answer
  b.update(null);
  assert.match(texts(c.body), /did not answer/);
}

// the phone's Quests tab is the same list
{
  fresh();
  const c = ctxFor([], { ok: true });
  const l = L.mount(node('div', 'd-panel'), c);
  l.update(gold('offers-sorceress'));
  assert.match(texts(c.body), /Den of Evil/);
  assert.equal(all(c.body, (x) => x.className === 'qb-box')[0].dataset.tab, '1');
}

// the video card's caption
{
  const V = fs.readFileSync(path.join(dir, 'video/panel.js'), 'utf8');
  const fn = new Function(V.match(/export function captionText[\s\S]*?\n}\n/)[0].replace('export ', '') + '; return captionText;')();
  assert.equal(fn(null), '');
  assert.equal(fn({ title: 'Den of Evil', state: 'in_progress', step: 'walk_blood_moor', npc: 'Akara' }), 'Den of Evil: walk blood moor, next Akara');
  assert.equal(fn({ title: 'Den of Evil', state: 'awaiting_handin', npc: 'Akara' }), 'Den of Evil: waiting for you at Akara');
  assert.equal(fn({ title: 'Den of Evil', state: 'blocked' }), 'Den of Evil: stuck');
}
console.log('quest ui ok');
