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
  assert.equal(m.sub, 'step 2 of 4, Walk blood moor, next: Akara', 'the step list says where the hero is');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.confirm], ['Stop', 'cancel', true], 'a stop asks first');
  m = B.stripModel(awaiting, L0);
  assert.equal(m.kind, 'awaiting');
  assert.equal(m.head, 'Akara is waiting: hand in Den of Evil');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.chunk], ['Hand in', 'handin', 'clear_den']);
  assert.equal(m.secondary, null, 'one primary button at a time');
}

// a hold (appendix B.4): "On hold" with the hold's sentence and a Stop, never "Running" with a step that does not move
for (const cls of ['paladin', 'sorceress']) {
  const v = clone(gold('running-' + cls));
  v.running.reason = { code: 'needs_potions', text: 'Going to town for potions: 3 of 4 healing potions.', remedy: 'chore' };
  const m = B.stripModel(v, fresh());
  assert.equal(m.kind, 'held', cls);
  assert.equal(m.label, 'On hold');
  assert.equal(m.head, 'Den of Evil: on hold');
  assert.equal(m.sub, 'Going to town for potions: 3 of 4 healing potions.');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.confirm], ['Stop', 'cancel', true]);
  const b = clone(gold('running-' + cls)); // a parked run says the park's sentence
  b.blocked = Object.assign({}, b.running, { state: 'blocked', why: 'old', reason: { code: 'died_too_often', text: 'Den of Evil is stuck: the hero died 2 times on this part.', remedy: 'player' } }); delete b.running;
  assert.equal(B.stripModel(b, fresh()).sub, 'Den of Evil is stuck: the hero died 2 times on this part.');
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
  const d2 = clone(done); d2.handIn = { quest: 'den_of_evil', chunk: 'clear_den', state: 'done', awarded: [{ kind: 'ether', amount: 5, key: 'quest/den_of_evil/-/normal' }] };
  assert.equal(B.settle(d2, L3).text, 'Done: Den of Evil   +5 Ether');
  assert.equal(B.paysText([{ kind: 'ether', amount: 2 }]), 'pays 2 Ether');
  assert.equal(B.paysText(undefined), '', 'no wallet, no chip');
  assert.equal(B.awardText([]), '');
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
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/cancel', { chunk: 'clear_den', confirm: true }], 'the second tap says so');
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
  assert.doesNotMatch(t, /auto: coming/, 'the auto switch is live now');
  assert.ok(byKey(c.body, 'auto-den_of_evil'), 'a live switch while the API offers it');
  assert.match(t, /auto: always you/, 'a quest the data keeps for the player is greyed with the reason');
  assert.ok(!byKey(c.body, 'auto-search_for_cain'), 'no switch on a never quest');
  const rows = all(c.body, (x) => x.className === 'qb-row').map((x) => x.dataset.group);
  assert.ok(rows.indexOf('ready') < rows.indexOf('available') && rows.indexOf('available') < rows.indexOf('locked'), 'ready, then not yet, then locked');
  // every part has its own state; Start only where ready
  const starts = all(c.body, (x) => x.dataset && /^start-/.test(x.dataset.key || ''));
  assert.deepEqual(starts.map((x) => x.dataset.key), ['start-den_of_evil-clear_den', 'start-sisters_burial_grounds-blood_raven']);
  byKey(c.body, 'start-sisters_burial_grounds-blood_raven').click(); await tick();
  assert.deepEqual(posts[0], ['/api/quests/sisters_burial_grounds/start', { chunk: 'blood_raven' }]);
  byKey(c.body, 'mode-idle').click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/mode', { mode: 'idle' }]);
  // the auto switch posts {on} to the real route; the page reads the new flag from the next GET
  let sw = byKey(c.body, 'auto-den_of_evil');
  assert.equal(sw.attrs.role, 'switch');
  assert.equal(sw.attrs['aria-checked'], 'false');
  sw.click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/auto', { on: true }]);
  const v2 = clone(v); v2.quests.find((q) => q.id === 'den_of_evil').auto = true;
  b.update(v2);
  sw = byKey(c.body, 'auto-den_of_evil');
  assert.equal(sw.attrs['aria-checked'], 'true');
  sw.click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/den_of_evil/auto', { on: false }]);
  // the board off: the cell says so, no switch
  const v3 = clone(v); v3.autoStart = { available: false, note: '' };
  b.update(v3);
  assert.match(texts(c.body), /auto: off/);
  // running: the Now row, with the stop confirm
  b.update(gold('running-paladin'));
  assert.match(texts(c.body), /Running: Den of Evil/);
  assert.equal(texts(byKey(c.body, 'now-cancel')), 'Stop');
  // history
  const h = gold('running-paladin'); h.history = [{ quest: 'x', chunk: 'y', title: 'Den of Evil', confirmed: true }];
  b.update(h);
  assert.match(texts(c.body), /Recent/);
  assert.match(texts(c.body), /handed in/);
  assert.doesNotMatch(texts(c.body), /Ether/, 'no wallet, no Ether words');
  // with a wallet: the offer chip, the payoff under the quest row and in Recent
  const w = gold('offers-paladin'); w.offers[0].pays = [{ kind: 'ether', amount: 2, key: 'k' }];
  b.update(w);
  assert.match(texts(c.body), /pays 2 Ether/);
  const pd = gold('running-paladin'); pd.quests[0].state = 'done'; pd.quests[0].chunks[0].state = 'done'; pd.quests[0].paid = [{ kind: 'ether', amount: 2 }];
  pd.history = [{ quest: 'x', chunk: 'y', title: 'Den of Evil', confirmed: true, paid: [{ kind: 'ether', amount: 2 }] }];
  b.update(pd);
  assert.match(texts(c.body), /\+2 Ether/);
  assert.match(texts(c.body), /handed in, \+2 Ether/);
  // off: the empty state names the knob
  b.update({ schema: 'watch.quests/1', enabled: false, reason: 'the quest board is off (knob quests.board is 0)' });
  assert.match(texts(c.body), /Turn on the Quest board in settings/);
  assert.match(texts(c.body), /quests\.board = 1/);
  assert.equal(c.st, 'ready');
  // no answer
  b.update(null);
  assert.match(texts(c.body), /did not answer/);
}

// slice S8: the countdown of an auto-start, the levels to go in words, the away card, the step list
{
  const L0 = fresh();
  const v = gold('offers-paladin');
  v.autoStart = { available: true, note: '', delaySec: 20, countdown: { quest: 'den_of_evil', questTitle: 'Den of Evil', chunk: 'clear_den', endsAt: '2026-11-17T12:00:20Z', secondsLeft: 12 } };
  let m = B.stripModel(v, L0);
  assert.equal(m.kind, 'countdown');
  assert.equal(m.head, 'Den of Evil starts by itself in 12 seconds');
  assert.deepEqual([m.primary.label, m.primary.verb, m.primary.quest, m.primary.chunk], ['Cancel', 'cancel', 'den_of_evil', 'clear_den']);
  v.autoStart.countdown.secondsLeft = 1;
  assert.match(B.stripModel(v, L0).head, /in 1 second$/);
  // the strip tile shows it and Cancel posts to the cancel route
  const posts = [];
  const c = ctxFor(posts, { ok: true });
  const s = S.mount(node('div', 'd-panel'), c);
  s.update(v);
  assert.equal(byKey(c.body, 'primary').textContent, 'Cancel');
  byKey(c.body, 'primary').click(); await tick();
  assert.deepEqual(posts[0], ['/api/quests/den_of_evil/cancel', { chunk: 'clear_den' }]);
  // levels to go, in words only
  assert.equal(B.levelsText(2), 'two levels to go');
  assert.equal(B.levelsText(1), 'one level to go');
  assert.equal(B.levelsText(0), '');
  assert.equal(B.levelsText(40), 'many levels to go');
  const t = gold('offers-paladin');
  t.offers.forEach((o) => { o.startable = false; o.readiness = 'risky'; o.levelsToGo = 3; o.reasons = [{ code: 'level', text: 'not yet' }]; });
  assert.equal(B.stripModel(t, fresh()).sub, 'three levels to go');
  t.meter = { mode: 'train', target: 'den_of_evil', levelsToGo: 2 };
  assert.equal(B.stripModel(t, fresh()).sub, 'two levels to go', 'the meter of the training target wins');
  for (const q of [m.head, B.stripModel(t, fresh()).sub]) assert.doesNotMatch(q, /\d+ level/, 'no level number on the page');
  // the step list
  const r = gold('running-paladin');
  assert.deepEqual(r.running.steps.map((x) => x.state), ['done', 'now', 'todo', 'todo']);
  assert.equal(B.stepText(r.running), 'step 2 of 4');
  assert.equal(B.stepText({ stepCount: 1, stepNo: 1 }), '');
  assert.equal(B.stepText({ stepCount: 2, stepNo: 2, partNo: 1, partCount: 4 }), 'part 1 of 4, step 2 of 2', 'the chain length leads');
  assert.equal(B.stepText({ stepCount: 1, stepNo: 1, partNo: 3, partCount: 4 }), 'part 3 of 4');
  const ra = clone(r); ra.running.auto = true;
  assert.match(B.stripModel(ra, fresh()).sub, /started by itself/);
}
{
  fresh();
  const posts = [];
  const c = ctxFor(posts, { ok: true });
  const b = B.mount(node('div', 'd-panel'), c);
  const v = gold('offers-paladin');
  v.whileAway = { events: [{ at: '2026-11-17T14:02:00Z', quest: 'den_of_evil', questTitle: 'Den of Evil', chunk: 'clear_den', kind: 'started', text: 'Started Den of Evil by itself' }, { at: '2026-11-17T14:20:00Z', quest: 'den_of_evil', questTitle: 'Den of Evil', chunk: 'clear_den', kind: 'handedIn', text: 'Handed in Den of Evil' }] };
  b.update(v);
  assert.match(texts(c.body), /While you were away/);
  assert.match(texts(c.body), /Started Den of Evil by itself/);
  assert.match(texts(c.body), /Handed in Den of Evil/);
  byKey(c.body, 'away-dismiss').click(); await tick();
  assert.deepEqual(posts.at(-1), ['/api/quests/away/dismiss', {}]);
  b.update(gold('offers-paladin'));
  assert.doesNotMatch(texts(c.body), /While you were away/, 'the card goes when the API stops sending it');
  // a stop in the last day is a plain-words note with the local time; one without a usable time says nothing
  const sv = gold('offers-paladin');
  sv.stopped = [{ quest: 'search_for_cain', questTitle: 'Search for Cain', at: '2026-11-18T14:11:00Z' }, { quest: 'x', questTitle: 'Broken', at: 'nope' }];
  b.update(sv);
  const hh = new Date('2026-11-18T14:11:00Z');
  const hhmm = String(hh.getHours()).padStart(2, '0') + ':' + String(hh.getMinutes()).padStart(2, '0');
  assert.match(texts(c.body), new RegExp('Search for Cain'));
  assert.match(texts(c.body), new RegExp('Stopped by you at ' + hhmm));
  assert.doesNotMatch(texts(c.body), /Broken/);
  b.update(gold('offers-paladin'));
  assert.doesNotMatch(texts(c.body), /Stopped by you/, 'gone once the server drops it');
  // a countdown is a Now row with its Cancel
  const cd = gold('offers-paladin'); cd.autoStart = { available: true, countdown: { quest: 'den_of_evil', questTitle: 'Den of Evil', chunk: 'clear_den', secondsLeft: 9 } };
  b.update(cd);
  assert.match(texts(c.body), /Den of Evil starts by itself in 9 seconds/);
  assert.equal(texts(byKey(c.body, 'now-cancel')), 'Cancel');
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

// the video card's caption: the strip publishes it, the video panel shows it
{
  fresh();
  assert.equal(B.captionOf(null), '');
  assert.equal(B.captionOf(gold('offers-paladin')), '', 'an offer is not a caption');
  assert.equal(B.captionOf(gold('running-paladin')), 'Den of Evil: walk blood moor, next Akara');
  assert.equal(B.captionOf(gold('awaiting-paladin')), 'Den of Evil: waiting for you at Akara');
  const bl = gold('running-paladin'); bl.blocked = bl.running; bl.running = null;
  assert.equal(B.captionOf(bl), 'Den of Evil: stuck');
  const seen = []; B.questCaption.subs.add((t) => seen.push(t));
  const c = ctxFor([], { ok: true });
  const s = S.mount(node('div', 'd-panel'), c);
  s.update(gold('running-paladin')); s.update(gold('running-paladin')); s.update(gold('awaiting-paladin')); s.update({ enabled: false });
  assert.deepEqual(seen, ['Den of Evil: walk blood moor, next Akara', 'Den of Evil: waiting for you at Akara', ''], 'published on change only');
  assert.doesNotMatch(fs.readFileSync(path.join(dir, 'video/panel.js'), 'utf8'), /fetch\(/);
}
// ---- every part state offers a button or says why not (user bug 2026-11-21: Blood Raven "waiting for you", no button) ----
{
  const act = (verb, label, x) => Object.assign({ verb, label }, x || {});
  // the table of the board's chunk states: what the server sends, the one button the row must render (null: a stated reason)
  const table = [
    { state: 'ready', action: act('start', 'Start'), btn: 'start', post: 'start' },
    { state: 'ready', action: null, why: 'Another quest is running', btn: null },
    { state: 'available', action: null, readiness: 'too_hard', why: 'too hard for now', btn: null },
    { state: 'locked', action: null, why: 'after Have Akara read the scroll', btn: null },
    { state: 'accepted', action: act('cancel', 'Stop', { confirm: true }), btn: 'cancel' },
    { state: 'in_progress', action: act('cancel', 'Stop', { confirm: true }), btn: 'cancel' },
    { state: 'blocked', action: act('retry', 'Try again'), btn: 'retry', post: 'retry' },
    { state: 'awaiting_handin', action: act('handin', 'Hand in'), btn: 'handin', post: 'handin', withAwaiting: true },
    { state: 'awaiting_handin', action: act('handin', 'Hand in'), btn: 'handin', post: 'handin' }, // no `awaiting` object
    { state: 'awaiting_handin', action: act('start', 'Start'), btn: 'start', post: 'start', reason: { code: 'handin_ready', text: 'The field work is already done: Start goes straight to the hand-in.' } },
    { state: 'awaiting_handin', action: act('handin', 'Handing in...', { disabled: true, why: 'x' }), btn: 'handin', disabled: true },
    { state: 'done', action: null, btn: null },
    { state: 'skipped', action: null, btn: null },
  ];
  for (const tab of [false, true]) for (const r of table) {
    fresh();
    const posts = [], c = ctxFor(posts, { ok: true });
    const b = tab ? L.mount(node('div', 'd-panel'), c) : B.mount(node('div', 'd-panel'), c);
    const v = clone(gold('offers-paladin'));
    const q = v.quests.find((x) => x.id === 'sisters_burial_grounds');
    const ch = q.chunks[0];
    Object.assign(ch, { state: r.state, readiness: r.readiness, why: r.why });
    if (r.action) ch.action = r.action; else delete ch.action;
    if (r.reason) ch.reason = r.reason;
    q.state = r.state;
    v.offers = v.offers.filter((o) => o.chunk !== 'blood_raven');
    v.running = v.awaiting = null;
    b.update(v);
    const what = JSON.stringify(r) + ' tab=' + tab;
    const line = all(c.body, (x) => x.className === 'qb-chunk' && x.dataset.state === r.state && /Blood Raven/.test(texts(x)))[0] || all(c.body, (x) => x.className === 'qb-chunk' && /Blood Raven/.test(texts(x)))[0];
    if (r.state === 'done' || r.state === 'skipped') continue; // finished rows show no part line
    assert.ok(line, 'part line ' + what);
    const key = r.btn && r.btn + '-sisters_burial_grounds-blood_raven';
    const buttons = all(line, (x) => x.tag === 'button');
    if (!r.btn) {
      assert.equal(buttons.length, 0, 'no button ' + what);
      assert.ok(texts(line).length > ch.title.length + 3, 'a stated reason ' + what);
      continue;
    }
    assert.equal(buttons.length, 1, 'exactly one button ' + what);
    assert.equal(buttons[0].dataset.key, key, what);
    assert.equal(!!buttons[0].disabled, !!r.disabled, what);
    if (r.post) {
      buttons[0].click(); await tick();
      assert.deepEqual(posts[0], ['/api/quests/sisters_burial_grounds/' + r.post, { chunk: 'blood_raven' }], what);
    }
  }
}

console.log('quest ui ok');
