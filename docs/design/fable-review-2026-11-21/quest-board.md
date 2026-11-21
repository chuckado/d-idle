# Quest board state machine: design review (Fable, 2026-11-21)

> status: proposal, decision-ready. Docs only: no Go and no behaviour changes on main. Card `fableboard` (TASK-714). The questions in section 7 are for the user; nothing here is an ADR.

Code read at `origin/main` `5b12ce5e7` and at `agent/boardprereq` `e77b53911`. That branch was in flight when the review started and merged to main while it was written (merge queue entry #1019, head `3cedddeeb`, CI still pending when I looked; S `mergeq list`). "Today" and "main" below mean `5b12ce5e7`; lines marked `e77b53911` are the merged fix. Tags: **V** = read in the code at that commit (file:line given) or run by me; **S** = read in a probe doc or card, not re-run; **U** = unverified. I ran no game and no test; every "V" below is a code read unless it says otherwise.

Appendices: [A, evidence per bug](quest-board-appendix-a-evidence.md); [B, tables: reason codes and wording, API, invariants and scenarios, panel spec](quest-board-appendix-b-tables.md).

## 1. Verdict in ten lines

1. The six bugs are one bug: **the board stores no lifecycle.** A chunk's state is recomputed every tick as `Derive(evidence now, tap records)` (V `internal/questboard/state.go:137`), on the design premise that evidence only moves forward (V `docs/design/quest-board-2026-11-16.md` section 1.5). Evidence does not only move forward, so every state can run backwards, and the one thing that is stored (the tap record) changes its meaning when it does.
2. Each fix of 2026-11-20 added a rule for one (state, record) combination that the derive function had left undefined. There are more such combinations than were fixed (appendix A lists five that are still open).
3. The fix is small in idea: **store the player's intent as a run with an explicit status, and latch completion.** Evidence stays derived; it becomes an input event to a transition table instead of being the state.
4. Dependencies become an invariant of the stored set (a live run's prerequisites are done or live), checked on every write and at load, not a test made once at the tap.
5. Every state in which nothing runs carries a reason code from a closed list, with a remedy class that says who acts: the bot (restock, wait, train), the player (tap), or nobody (failed by the game).
6. API verbs act on the stored run, so they are idempotent by construction: Start means "a live run exists", Stop means "no live run of this chunk or of anything that waits for it".
7. With the arbiter, the board is one intent source. Its holds are exported as world facts that the upkeep sources read; the board never orders a town trip itself.
8. Migration: 10 small cards, scenarios first, behind one knob, with the plan file dual-written for one release so a rollback loses nothing.
9. `agent/boardprereq` (merged meanwhile) is the right behaviour for the interim: accept implies the chain, Stop implies dependants. Three points in it for the follow-up are in section 6.2.
10. Eleven questions for the user in section 7; the two that shape the work are Q1 (store the lifecycle) and Q4 (what happens when exact evidence goes back).

## 2. Root cause

### 2.1 How state is made today

| layer | what it holds | survives a restart | where (V) |
|---|---|---|---|
| plan file | `Accepted` records (quest, chunk, difficulty, time, mode, auto), skips, opt-ins | yes | `internal/questboard/plan.go:18-25,66-80` |
| derive | the nine states, computed from evidence and the records | recomputed | `state.go:195-261` |
| `Exec.rs` | started, failures, deaths, the park reason | no | `exec.go:92-100` |
| `boardRT` | hand-in ran, chunk seen done once, walked to the giver, last failure time, started key, queued text | no | `internal/bot/questboard.go:54-67` |
| `task.Executor` | blocked groups, fail and strike counters | no | `internal/task/executor.go:476,517-520` |
| auto controller | countdown, muted offers, tried hand-ins | no | `questboard/auto.go:72-82` |

The design says "never handler progress, everything else is derived" (V design 1.5). In practice four unpersisted stores each hold a piece of the lifecycle, because the derive function alone could not express it. They are cleared by different verbs: Stop clears `Exec.rs` and three of the `boardRT` maps but not `reached` and not the task executor's block (V `exec.go:362-363`, `questboard.go:607-614`); Try again clears `Exec.rs` only (V `exec.go:393-397`), so a chunk parked by the task executor is stuck again on the next slice (V `questboard.go:632-634`, `executor.go:291-292`; U: I found no other caller that clears the group on the board path, `Unpark` is called only from `staletarget.go:163,244`).

### 2.2 The three false premises

**P1. Evidence is monotonic.** It is not.
- A chunk that ends with an item or an area is done only while the proxy holds: "the stones chunk reads done only while the hero stands in Tristram (or holds a portal) and the scroll chunk only while the key is in the pack" (V comment and code `internal/bot/questboard.go:534-544`). The patch was an in-process map `reached`; a restart empties it.
- A hand-in is done at T3 because a process-local flag says the dialogue ran (V `questboard.go:528-530,905-907`); a restart before the save confirms it reverts the chunk.
- A fresh memory read may deny a proxy (V `state.go:269-271`), and a quest record bit can go back (bug 5, S from the brief).

**P2. A tap record has one meaning.** It has four, and it is never removed when the chunk completes (the only removal is Cancel, V `plan.go:292-306`):
1. consent to run now: `open()` (V `exec.go:214-222`);
2. proof that the board ran the chunk, which gates awards: `touched` (V `plan.go:135-142`, `state.go:348`);
3. the history row (V `desk.go:394-400`);
4. queue membership and order, since #1012 (V `exec.go:224-231`).

Whether a record is a live run or a finished one is itself derived (`c.State != Done`, V `exec.go:220`, and the `finished` callback at `exec.go:316-319`). So when P1 fails, a finished record becomes a live run again. That is how a plan file comes to hold two open records without anyone tapping twice (mechanism V in code; whether the sandbox report of 2026-11-20 got there this way or by two taps is U, I did not find the report itself).

**P3. A check at the tap is enough.** Start is checked against the evidence of that moment (`CanStart`, V `state.go:382-399`); the run is decided against the evidence of a later tick (`Next`, V `exec.go:467-502`). Nothing defines what happens when the two differ. The catch-all is the `default:` branch of `Next` (V `exec.go:495-499`): any combination nobody thought of becomes `DecBlocked` with a text such as "chunk is locked", no transition out, and (see 2.4) no row on the page.

### 2.3 The six bugs as one table

| # | symptom | undefined combination | why it follows from 2.2 | fixed by |
|---|---|---|---|---|
| 1 | "more than one chunk is accepted: parked, not running any" | two records whose derived state is not done | P2: records are never closed, "open" is derived; the old `Next` refused the case (V diff of `b28a34020`) | #1012: sort by data order, first runs |
| 2 | Stop on a queued chunk answers "not running", withdraws nothing | record exists, derived state `locked` | the verb switched on the derived state, not on the record it mutates (V `desk.go:767-779`) | #1012: `HasRecord` |
| 3 | stones waits for ever for "after Have Akara read the scroll" | live record whose prerequisite has no record | P3: the order was tested at the tap; the stored set has no invariant; `Next` default branch (V `state.go:284-288`, `exec.go:495`) | `agent/boardprereq` |
| 4 | guard holds a step (potions under 4), `needs_attention`, no restock | accepted chunk, step refused for a fixable reason | a refusal has no type: `fieldReady` returns the same danger fault, with no chores step, for "too weak" and for "3 potions" (V `internal/task/gate.go:269-281`); the board counts it as a failure and parks after three (V `questboard.go:747-752`, `exec.go:513-520`). Design 2.3 ("a wall refuses: not blocked, the bot buys them") is implemented only for chunks that are not accepted: `Missing` is read after the `rec` return (V `state.go:240-252`) | open |
| 5 | evidence goes back, a downstream chunk is already accepted | completed chunk becomes open while dependants are live | P1 and P2 together; no rule exists | open |
| 6 | the player cannot tell why nothing runs | any non-running state | reasons are free text from six producers; the page shows some of them (2.4) | open |

### 2.4 Why the page cannot say why (bug 6)

- There are two "blocked" notions that do not agree: the derived state `blocked` needs a record and an executor park (V `state.go:217-221`); the decision `DecBlocked` is "anything that is not run or await" (V `exec.go:495-499`).
- The view picks the card from the derived state: awaiting, blocked, **else running** (V `desk.go:417-424`). In the deadlock of bug 3 the active chunk derives `locked`, so the page shows "Running: Search for Cain" with a Stop button while the hero trains, and no Needs-you row is raised (the row needs `v.Blocked`, V `desk.go:545`). Try again answers "is not stuck" (V `desk.go:798-801`).
- `in_progress` does not mean the hero is on it: it is set when any step reads satisfied or the process once started the chunk (V `exec.go:168-176`), also while a recovery holds it.
- The reason an accepted chunk waits is computed (`queued`, V `questboard.go:944-959,1004,1021`) and put in the view (V `desk.go:414-416,425`), and the further queued chunks are in `extra` (V `desk.go:428-430`), but the panel reads neither: `grep -n "queued\|extra" internal/watch/web/panels/quest-board/panel.js internal/watch/web/panels/quest-strip/panel.js` finds no use of `run.queued`, `strip.queued` or `v.extra` (V, run 2026-11-21).
- The reason texts are sentences built in Go at six places (`ChunkView.Why`, `Decision.Why`, `boardRT.queued`, `AutoView.Why`, `OfferView.Why`, `Reply.Reason`). No code, so the page cannot choose a button from the reason, and `agent/boardprereq` already has to detect a case by the text prefix `"after "` (V `exec.go:367` at `e77b53911`).

## 3. The clean model

One sentence: **facts are derived, intent is stored, completion is latched.**

### 3.1 Three kinds of data

| kind | examples | stored | changes by |
|---|---|---|---|
| **facts** (per tick) | quest word and tier, chunk checks, readiness word, hero in town, potions, gold | no | the game |
| **runs** (the player's intent) | "run search_for_cain/stones", its status, its hold or park, who accepted it, its place in the queue | yes, plan file v2 | transitions of 3.3 only |
| **latches** (completion marks) | "scroll was done at T3 on 2026-11-20, raised to T1 later" | yes, plan file v2 | 3.6 only |

The nine wire states stay as they are for the page; they become a projection: a chunk with a live run shows the run's status, a chunk without one shows its offer status.

### 3.2 States

Offer status (derived, no storage; today's rules, V `state.go:195-261`):

| offer status | rule |
|---|---|
| `skipped` | difficulty file or the player's skip |
| `done` | latch present, or exact quest evidence |
| `failed` | the game failed the quest (shown as `blocked`, code `game_failed`) |
| `locked` | a `requires` quest is not done on T1/T2, or an `after` chunk is neither latched nor live |
| `available` | no handler, or readiness too hard or unknown, or a readiness fact is missing |
| `ready` | readiness safe or risky; also a chunk whose pending bit is exact although nobody started it (Start then goes straight to the hand-in, case C5 of appendix A) |

Run status (stored):

| run status | wire state | meaning | who must act |
|---|---|---|---|
| `queued` | `accepted` | live, waits for an earlier run of the same chain | nobody (bot) |
| `running` | `in_progress` | the head run; the hero works on it or is about to | nobody (bot) |
| `held` | `in_progress` plus `hold` | the head run cannot proceed now for a reason the bot resolves or waits out | nobody (bot) |
| `awaiting` | `awaiting_handin` | field work done, the hero waits for the tap | player |
| `handing_in` | `awaiting_handin` plus `handIn.running` | the dialogue runs | nobody (bot) |
| `parked` | `blocked` | the bot gave up; reason code and counters kept | player |
| `finished` | `done` | closed with a latch; kept as history and award proof | none |
| `stopped` | offer status | closed by the player or by a cascade; kept as history | none |

`held` is the new state. It is the difference between "needs you" and "the bot is on it", which today is one free-text line.

### 3.3 Transitions of a run

Causes: **P** player (API), **B** bot (executor report), **E** evidence (a fact changed), **G** game event (death, new game, difficulty change), **L** load (restart).

| from | event | cause | to | notes |
|---|---|---|---|---|
| none | start(c), c ready | P, or auto | `queued` then head | accepts the unmet `after` chain with it (3.4 D1), one save |
| none | start(c), c not startable | P | none | refused with the offer's reason code |
| `queued` | it becomes the first live run in data order | B | `running` | |
| `running` | step refused, remedy bot | B | `held(code)` | no failure counted |
| `held` | the hold's fact clears | E | `running` | |
| `held` | hold older than its budget | B | `parked(hold_timeout, code)` | every hold code has a budget (appendix B.1) |
| `running` | step failed, under the cap | B | `running` | counter up; spacing by the arbiter back-off |
| `running` | failures or deaths at the cap | B, G | `parked(code)` | counters are stored with the park |
| `running` | objective met, chunk ends handin | E | `awaiting` | T3 is enough to stop and ask |
| `running` | chunk done, ends item or auto | E | `finished` | latch written in the same save |
| `awaiting` | handin tap | P, or auto when away | `handing_in` | waits for town; not a failure in the field |
| `handing_in` | read-back done | E | `finished` | latch T3 or exact; award on exact |
| `handing_in` | attempts used up | B | `parked(handin_failed)` | |
| `parked` | retry | P | `running` | clears the board counters and the task executor group |
| any live | stop | P | `stopped` | cascades to dependants (D2) |
| any live | exact evidence says done | E | `finished` | done by hand meanwhile: quiet, the award rule is unchanged |
| any live | quest failed by the game | G | `parked(game_failed)` | no retry offered |
| any live | difficulty changed | G | unchanged, not shown | runs are per difficulty, as today |
| any live | restart | L | `queued`, `awaiting` or `parked` as stored; the head becomes `running` | holds are recomputed; a park survives |
| `finished` | latch cleared by exact evidence | E | see 3.6 | never reopens by itself |

Every cell that is not in the table is a no-op that answers `already`. The table is total over (status x event): that is invariant I8 and the unit test that replaces the `default:` branch.

A quest's state stays the aggregate of its chunks (V `state.go:318-345`), with one change: the order is `parked` > `awaiting` > `held` > `running` > `queued` > offer states, so a quest that needs the player always says so.

### 3.4 Dependency semantics

- **D1. Accept implies the prerequisite chain.** Starting chunk C creates runs for every `after` ancestor of C in the same quest that is neither latched nor live, in data order, in one save, marked `by: chain`. Each must be startable (ready, or already live); if one is not, nothing is written and the refusal names it with its code. This is what `agent/boardprereq` does (V `exec.go:310-319,369-405` at `e77b53911`).
- **D1 limit.** Quest-level `requires` (another quest, T1/T2) are never accepted by implication: another quest has its own risk and its own consent (Andariel is `hazard`, `autoStart never`, V `data/questboard/act1.json`).
- **D1 does not run the later parts.** "A tap never covers the next part" (V `exec.go:289`) stays true: the chain ends at C.
- **D2. Stop implies dependants.** Stopping C stops every live run of the same quest that waits for C, directly or not. The reply lists them (`changed`), the toast names them. Also in `agent/boardprereq`.
- **D3. One head.** At most one run is `running`, `held`, `awaiting` or `handing_in` per hero and difficulty. The head is the first live run in data order. When the head is parked nothing runs: everything behind it stays `queued` with code `waits_for`, because by D4 and D5 it depends on the head.
- **D4. Closure (the invariant bug 3 broke).** For every live run, each `after` ancestor is latched or has a live run. Checked on every write; at load a violation is repaired by adding the missing ancestors when they are startable, else the run is parked with `prereq_missing`. The repair is a load step with a log line, never a write inside `Next`.
- **D5. One quest at a time** (today's rule "one chunk at a time", V `exec.go:310-312`, generalised): all live runs belong to one quest. A Start for another quest is refused with `busy` and the name of the running quest. Queueing across quests is Q7.

### 3.5 Holds and parks: every stop has a reason code

A reason is `{code, args, remedy, since}`; the sentence is built from the code by one table (appendix B.1 has the exact wording). Remedy classes:

| remedy | meaning | examples (codes) | what the bot does |
|---|---|---|---|
| `chore` | a town visit fixes it | `needs_potions`, `needs_gold_withdraw`, `needs_repair`, `needs_kit` | goes to town now and does it (bug 4) |
| `wait` | time or another run fixes it | `waits_for`, `fight_first`, `recovery_first`, `gold_order_first`, `order_first`, `area_banned`, `handin_in_town`, `confirming` | waits; the mode does not take over for short waits |
| `train` | levels or gold fix it | `too_weak`, `route_closed`, `needs_gold_farm` | the between-chunks mode runs with this chunk as its target |
| `player` | a tap fixes it | `handin_wanted`, `paused`, `gave_up`, `died_too_often`, `handin_failed`, `prereq_regressed`, `prereq_missing`, `quest_locked`, `hold_timeout` | waits, Needs-you row |
| `none` | nothing fixes it here | `game_failed`, `unsupported` | shows it |

To make bug 4 impossible the task layer has to say which kind a refusal is. Today `Fault{Kind, Why, Insert}` (V `gate.go:244-250`) carries a kind but the guards use `KindDanger` for both a level floor and a potion count. The change: the guard reasons are produced as a code plus numbers (`potions 3 < 4`, `clvl 12 < 20`, `no merc`), and the potion and gold codes carry the chores step, as `gatePotions` already does for the boss gates (V `internal/task/act1.go:99-110`). No class or build literal is needed: the floors are already data (V `internal/bot/routeguard.go:17-18`, `data/questboard/readiness.json`).

### 3.6 Completion latch and evidence regression

- **Latch.** When a chunk first reads done, the board stores `{quest, chunk, difficulty, tier, at, source}` in the same save that closes its run. This replaces the in-process `reached` and `handed` maps (V `questboard.go:54-55`), so a restart no longer walks back to the Dark Wood for a second scroll.
- **R1. Silence of a proxy never clears a latch.** The key leaving the pack, the hero leaving Tristram, a restart: the chunk stays done.
- **R2. A latch only rises** (T3 to T1 or T2). An award is paid when it becomes exact, once, by ledger key (unchanged, V `state.go:347-362`).
- **R3. Exact contradiction clears a latch.** Only a T1 or T2 read that says "not done" where the latch says done, on evidence that covers this chunk (the quest word for a one-chunk quest, as `memoryDenies` does today, V `state.go:269-271`; a sub-record bit for a part, when one exists). It is logged as a `regressed` event with both readings.
- **R4. What a regression does.** The chunk returns to its offer status. Nothing is re-run silently. If live runs depend on it, the head is parked with `prereq_regressed` and the Needs-you row offers "Run it again" (accepts the regressed chunk `by: chain`) and "Stop". Reason: a regression of exact evidence is either a reader fault or a real loss, and in both cases redoing field work unasked is the CAIN-L3 fault again. Alternative (re-run by itself when the quest allows auto-start) is Q4.
- **R5. Awards are never clawed back.**
- **R6. Done by hand, never ran by the board:** unchanged, quiet, no award (V `state.go:185-190`).

### 3.7 Persistence across restarts

Plan file version 2 adds `runs` and `latches`; `accepted` stays and is written from the live and finished runs for one release, so an old binary still reads the file (today a future version is a load error, V `plan.go:214-215`).

| stored | not stored |
|---|---|
| runs with status `queued`, `awaiting`, `parked`, `finished`, `stopped`; park code and counters; `by`; accepted time | `running` and `held` (the head is recomputed), hold reasons, countdowns, step position |
| latches | readiness words, facts |

Changes of behaviour that this implies, each on purpose:
- A park survives a restart. Today a restart clears it and the chunk is "simply tried again" (V `exec.go:92-93`), so a chunk parked for two deaths gets two more after every restart at night.
- A chunk that was done at T3 stays done after a restart.
- Migration v1 to v2 at load: a record whose chunk reads done becomes a `finished` run with a latch at the tier read; any other becomes `queued`; then D4 repair runs once.

### 3.8 API semantics (idempotent)

Verbs state a target condition on the stored runs. Repeating a verb changes nothing and answers `ok: true, already: true`. Every reply carries `changed` (the runs it touched, with their new state and reason) and the refusal code; the page still reads the result from the GET (V rule in `desk.go:12-15`). Full table in appendix B.2. The two that were wrong on 2026-11-20:

| verb | target condition | wrong before | now |
|---|---|---|---|
| start(c) | a live run of c exists (with its chain, D1) | refused or deadlocked depending on the derived state | one rule, refusals carry the offer's code |
| cancel(c) | no live run of c or of its dependants | "not running" while a record existed (bug 2) | acts on the record; `already` only when no run existed |

### 3.9 Invariants

Short list (assertion sketches and the scenario that carries each are in appendix B.3):

| id | invariant | checked by |
|---|---|---|
| I1 | at most one head run per hero and difficulty | unit, on every store write; scenario `board.sm.one-head` |
| I2 | closure: prerequisites of a live run are latched or live (D4) | unit on write and load; `board.prereq-not-accepted` |
| I3 | the head is the first live run in data order; nothing runs past a parked head; data order is a topological order of `after` | questspec rule 12; `board.two-accepted-chunks` |
| I4 | no live run for a latched chunk; a run finishes and its latch is written in one save | unit; `board.sm.restart-mid-chain` |
| I5 | a live run that is not `running` has a reason code from the closed list | unit over the table; `board.view.reason-always` |
| I6 | a latch never falls except by a logged `regressed` event | unit; `board.sm.regress-proxy` |
| I7 | start, cancel, retry are idempotent | unit (apply twice, compare files); `board.sm.stop-idempotent` |
| I8 | the transition table is total; `Next` has no default branch | unit over status x event |
| I9 | save, load, decide equals decide (holds aside) | unit; `board.park.persist-restart` |
| I10 | every hold code has a remedy and a budget; no hold outlives its budget | unit; `board.hold.*` |
| I11 | the card the page shows equals the decision kind | unit on `Desk.View`; `board.view.reason-always` |
| I12 | an award key pays once | exists (V `ledger_test.go`) |

## 4. The board and the arbiter

The arbiter is in shadow (V `docs/arbiter.md`, `internal/arbiter`). Its classes, low to high: filler, objective, upkeep, player, reflex (V `internal/arbiter/arbiter.go:16-24`). Its quest source today is a single fact, `objective` (V `internal/arbiter/sources.go:112-122`).

### 4.1 The board as an intent source

The board proposes at most one intent, from the head run, and reads only world facts:

| head run | intent | class | urgency | area | resumable |
|---|---|---|---|---|---|
| `running` | `quest/chunk:<q>/<c>` reason "you started <title>: <step>" | objective | 0.6; 0.9 when the chunk is `noGold` | the next step's area | yes |
| `held`, remedy `chore` | none; the board exports facts instead (4.2) | | | | |
| `held`, remedy `wait` | `quest/hold` only when it must stand still (hand-in in town); else none | objective | 0.6 | town or any | yes |
| `held`, remedy `train` | none; the filler (mode) runs with the chunk as pinned target | | | | |
| `awaiting` | `quest/await` "waiting at <npc> for your tap" | objective | 0.3 | town | yes |
| `handing_in` | `quest/handin` | **player** | 1 | town | no |
| `parked`, none | none | | | | |

Why these numbers: they reproduce today's order without a special case. Recovery is objective 0.8 and the gold order 0.7 (V `sources.go:100-111`), so both stay ahead of a normal chunk (V today: `questboard.go:984,1000-1005`), and a `noGold` chunk at 0.9 passes the recovery (V today: `chunkNeedsNoGold`, `questboard.go:920-926`). Whether a Start tap should be class `player` instead is Q8.

### 4.2 Holds become facts, not orders

A source may not read another source (V `docs/arbiter.md` "Sources read only View"). So the board does not ask for a town trip. It publishes what the head chunk needs as facts, and lifts its hold when the facts are met:

| hold code | fact the board raises | source that acts | class |
|---|---|---|---|
| `needs_potions` | `potions_min` = max(build line, the guard's floor) | resupply | upkeep |
| `needs_gold_withdraw` | `need_gold` (stash counts) | resupply, with its withdraw step (V `sources.go:83-98`) | upkeep |
| `needs_gold_farm` | `recovery` | recovery | objective 0.8 |
| `too_weak` | pinned target of the filler | level | filler |

Upkeep is above objective, so the restock wins as soon as the hold exists, and the batching rule makes it one town visit. Bug 4 then has no code path: a guard that refuses for potions cannot leave the hero idle, because the same number that refuses is the number the resupply source reads.

### 4.3 Rules against the other families

| against | rule | replaces today |
|---|---|---|
| survival, chicken (reflex) | always preempts; the chunk intent is resumable and continues after it. A death is reported to the board (counter), a chicken is not a failure | `StepDied` and `StepWaiting` mapping (V `questboard.go:737-743`) |
| player order, Town, Pause (player) | preempts; the run stays live with hold `paused` or `order_first` | `DecPaused` (V `exec.go:492-494`) |
| town chores (upkeep) | preempt a chunk between steps; never during `handing_in` | implicit in `r.do` |
| recovery, gold order (objective, higher urgency) | ahead of a chunk unless `noGold`; the hold names them | `boardPreemptsWhy` (V `questboard.go:975-1056`) |
| level and farm goals (filler) | a live head run always wins; filler runs while the run is held for `train`, parked, or absent | `boardPreempts`, `boardBetween` |
| commitment | the arbiter lease (120 virtual seconds) keeps a started chunk from being swapped out step by step | `startedKey` and the 20 s queue grace (V `questboard.go:62-67,941-942`) |
| failure spacing | arbiter back-off (30 s doubling) | `boardFailGap` 15 s and the forced sleep in town (V `questboard.go:27,773-782`) |

### 4.4 Before the arbiter is live

One function decides for the board per tick and both call sites of `goal.go` use its result. Today `Next` is called up to three times in a tick from different branches (V `questboard.go:660,937,992`, `goal.go:121,128,131`), each adding its own conditions, and `agent/boardprereq` puts a store write inside `Next` (V `exec.go:624` at `e77b53911`), although `Next` is documented as "calling it twice changes nothing" (V `exec.go:466`) and the view path (`State`) does not repair, so page and loop can disagree for a tick. The interim rule: `Next` is pure; repairs happen at load and in the verbs.

## 5. Migration plan

Ten cards. Scenarios first; one knob `quests.boardSM` (0 = today's code); the plan file is dual-written until card 10. "Owns" are the paths the card would need; the coordinator sets the real `owns:`.

| # | card | what | owns (paths) | scenarios, tests | risk | rollback |
|---|---|---|---|---|---|---|
| 1 | `boardsm-scn` | scenario file `board.json` with the ids of appendix B.3, pending or red; two runner additions: assert kind `board` (GET /api/quests at the end: chunk state and reason code) and recorded API replies | `tools/provingground/scenarios/board.json`, `internal/scenario`, `cmd/d2scenario` | the set is red on main by design | low: runner only | delete the file |
| 2 | `boardsm-model` | pure machine: run, latch, reason codes, the transition table, closure, head selection. No wiring, no I/O | `internal/questboard/sm*.go` | table tests I1 to I10 offline, two builds as fixtures | low | delete |
| 3 | `boardsm-store` | plan v2 (runs, latches), v1 import at load, dual-write of `accepted`, D4 repair at load | `internal/questboard/plan*.go` | round trip, import of the five v1 shapes of appendix A, old binary reads the new file | medium: the player's file | v1 block stays valid; `.bak` exists (V `plan.go:194-195`) |
| 4 | `boardsm-exec` | `Exec` on the machine behind the knob; `Next` pure and total; latch replaces `reached` and `handed`; Retry clears the task executor group | `internal/questboard/exec.go`, `internal/bot/questboard.go` | `board.sm.*`, `board.park.persist-restart`, cain chain | medium | knob 0 |
| 5 | `boardsm-holds` | guard refusals as codes with a remedy; potion and gold refusals insert the chores step; hold budgets | `internal/task/gate.go`, `internal/task/kind.go`, `internal/bot/routeguard.go`, `internal/bot/questboard.go` | `board.hold.restock`, `.restock-stash-gold`, `.broke`, `.too-weak`; Paladin and Sorceress | medium: touches the death-spiral guards | knob 0 keeps the old fault kinds |
| 6 | `boardsm-api` | verbs on runs, `changed` in replies, `hold` and `queue` in the view (additive to `watch.quests/1`) | `internal/questboard/desk.go`, `internal/watch/questsapi.go`, contract file | `board.sm.stop-idempotent`, `.start-idempotent`, `.stop-cascades`, `board.view.reason-always` | low | additive fields |
| 7 | `boardsm-panel` | the panel spec of appendix B.4 | `internal/watch/web/panels/quest-board`, `quest-strip`, `tools/uishots/fixtures_quests.go` | panel model tests, one fixture per row of B.4 | low | page only |
| 8 | `boardsm-checker` | questspec rules 12 to 14 (appendix B.3) in the merge gate | `internal/questspec`, `data/questspec` | `TestShippedChains` | low | allow-list |
| 9 | `boardsm-arbiter` | the board intent source and the exported facts in shadow; read the divergence lines of one sandbox night | `internal/arbiter/sources.go`, `internal/bot/shadow.go` | arbiter replay tests | low: shadow | knob `bot.arbiterShadow` |
| 10 | `boardsm-default` | knob default 1 after the scenario set is green on a determinism pair; one release later remove the old paths and the dual-write | knob table, `internal/questboard`, `docs/questboard.md` | full set | medium | knob |

Order: 1, 2, 3, then 4 and 5 (different files, can run in parallel), 6, 7; 8 and 9 any time after 2. `agent/boardprereq` is on main and is the behaviour the scenarios of card 1 pin.

## 6. Notes on the code as it is

### 6.1 Smaller findings (all V, details in appendix A)

- Try again does not reach the task executor's block (2.1).
- The page shows "Running" for an active chunk that is not running (2.4).
- A park does not survive a restart (3.7).
- `open()` sorts by data order and relies on the data order being a topological order of `after`; `def.go` checks cycles only (V `def.go:466-496`, `exec.go:224-231`).
- The design's `quests.handinIdleMin`, `quests.resumeWindowMin` and `quests.retryMax` (V design 2.1, 2.2) are not knobs in `internal/env/tuning_quests.go` (V, grep): an awaiting hero stands at the giver without limit (V `questboard.go:848-867`). See Q6.

### 6.2 `agent/boardprereq` (V at `e77b53911`)

It implements D1 and D2 and is the right interim. Three points for the follow-up, none a reason to revert:
1. `afterOnly` recognises the case by the reason text (`strings.HasPrefix(c.Why, "after ")`). A wording change breaks the logic; the reason needs a code (card 2).
2. The chain accept returns before `CanStart`, so the chunk the player tapped is accepted without its readiness word being looked at (it derives `locked`, so there is none). After the chain runs, that chunk starts even if it would have been "too hard". With the model, readiness of the tapped chunk is checked when it becomes the head (hold `too_weak`).
3. The repair writes from inside `Next` (4.4).

## 7. Questions for the user

Each has my recommendation first. Nothing is decided here.

| # | question | recommendation | alternative |
|---|---|---|---|
| Q1 | May the board store the lifecycle of a run (status, park, completion latch) in the hero's plan file? This amends design 1.5 "never handler progress". | Yes. It stores the player's intent and "seen done", never the step position; the six bugs come from not having it. | Keep deriving and add a rule per incident; I expect more incidents with every multi-part quest of Acts 2 to 5. |
| Q2 | Start on a later part of a quest accepts the earlier, unfinished parts too (D1)? | Yes, with the toast naming them. | Refuse and say "start part 1 first". |
| Q3 | Stop on a part also stops the parts that wait for it (D2)? | Yes, the toast names them. | Stop only the tapped part; the others would wait with code `prereq_missing`. |
| Q4 | Exact evidence goes back for a finished part while later parts are accepted: | Park and ask ("Run it again" or "Stop"); never redo field work unasked. | Re-run by itself when the quest allows auto-start. |
| Q5 | Does a parked chunk stay parked after a bot restart? | Yes. | Today: the restart clears it and the bot tries again. |
| Q6 | How long does the hero stand at the giver waiting for the hand-in tap? | 30 virtual minutes (the design's `quests.handinIdleMin`), then the mode takes over; the chunk stays awaiting and the tap sends the hero back. | Without limit (today). |
| Q7 | May the player queue chunks of a second quest while one runs? | No for now: one quest at a time, the refusal names the running one. Revisit after Act 2. | A queue across quests in tap order. |
| Q8 | Priority of a started chunk against the hero's own needs: | Keep today's order: survival, your orders and Pause, town chores, gold recovery and the Farm gold order, then the chunk, then level and farm goals. A chunk that buys nothing passes the recovery. | Make a Start tap a player order above recovery; the hero may then start a quest broke. |
| Q9 | A guard refuses a step for something a town visit fixes (potions, repair, gold in the stash): the hero goes to town at once and continues, with no failure counted? | Yes. | Count it as a failure (today: parks after three). |
| Q10 | A hold the bot cannot resolve in time (default proposal: 20 virtual minutes for chores, 60 for training holds): | Park with the reason and a Needs-you row. | Hold without limit and show the reason only. |
| Q11 | May the coordinator turn section 5 into the ten cards? | Yes. | |

Proposed text for the decisions index, if Q1 to Q5 are answered as recommended (for the coordinator or the user's session to place; I did not edit `docs/decisions.md`):

> Quest board lifecycle (user, date): the board stores each accepted chunk as a run with a status and latches completion in the hero's plan file; evidence stays derived and is an input, not the state. Accept implies the unfinished earlier parts of the same quest, Stop implies the parts that wait for it, a park survives a restart, and a regression of exact evidence parks and asks. Amends quest-board design 1.5.
