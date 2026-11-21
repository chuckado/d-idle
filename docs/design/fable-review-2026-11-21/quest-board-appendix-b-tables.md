# Appendix B: tables (quest board review 2026-11-21)

> status: appendix of [quest-board.md](quest-board.md); proposal. Everything here is design (I) unless tagged V. Wording follows the existing page style (V `internal/questboard/desk.go:534-592`, `internal/watch/web/panels/quest-board/panel.js:124-177`): sentences, no numbers for margins, the quest title as the player knows it.

Placeholders: `{quest}` quest title, `{part}` chunk title, `{npc}` giver, `{n}` a count, `{have}`/`{need}` counts, `{area}` area name, `{why}` a sentence from the producer.

## B.1 Reason codes and the exact wording

One table builds the sentence from the code; Go passes code and arguments, never a sentence. The same code is in the view (`hold.code`), in the log line and in the scenario assertions.

### Holds (the bot acts or waits; the run stays live; no failure is counted)

| code | remedy | the player reads | budget, then | today |
|---|---|---|---|---|
| `waits_for` | wait | Starts after {part}. | none (bounded by the head run) | "waits for q/c" in `extra`, not shown |
| `fight_first` | wait | Starts when the fight is over. | 20 s, then it starts anyway | V `bot/questboard.go:1021,941-942` |
| `recovery_first` | wait | Starts after the hero has refilled the purse and restocked. | 60 min, then `hold_timeout` | V `bot/questboard.go:1004` |
| `gold_order_first` | wait | Your Farm gold order comes first. | none (the player's order) | log only, V `bot/questboard.go:985` |
| `order_first` | wait | Your order comes first; {quest} continues after it. | none | not shown |
| `needs_potions` | chore | Going to town for potions: {have} of {need} {kind} potions. | 20 min, then `hold_timeout` | failure, parks after three (bug 4) |
| `needs_gold_withdraw` | chore | Taking gold from the stash to restock. | 20 min | chores step cannot pay (S act1more probe) |
| `needs_repair` | chore | Going to town to repair. | 20 min | |
| `needs_kit` | chore | Going to town for {item}. | 20 min | design 2.3 "needs 3 antidotes: buying them" |
| `needs_gold_farm` | train | Not enough gold to restock: farming gold first. | 60 min | recovery, reason not shown |
| `too_weak` | train | Too hard for now: {why}. Training toward it. | 60 min | "not ready: ..." as a failure |
| `route_closed` | train | The road to {area} is closed for now: {why}. | 60 min | "not ready: route guard (...)" as a failure |
| `area_banned` | wait | {area} is closed after {n} deaths; trying again later. | until the ban ends | "area banned after ..." as a failure |
| `handin_in_town` | wait | Hand-in is on its way: the hero is going to town. | 5 min (V `handInTimeout`, `desk.go:33`) | stays queued, not shown |
| `confirming` | wait | Handed in. Checking that the game agrees. | until the next save | V `panel.js:135` |
| `loading` | wait | Waiting for the game. | 2 min | silent |

### Needs you (the run is `awaiting` or `parked`; a Needs-you row is raised)

| code | the player reads | buttons | today |
|---|---|---|---|
| `handin_wanted` | {npc} is waiting: hand in {quest} | Hand in | V `desk.go:542` |
| `paused` | Paused: the hero continues when you resume. | (Resume is the page's own button) | V `bot/questboard.go:1045` |
| `gave_up` | {quest} is stuck: gave up after {n} tries: {why} | Try again, Stop | V `exec.go:518` |
| `died_too_often` | {quest} is stuck: the hero died {n} times on this part. | Try again, Stop | V `exec.go:526` |
| `handin_failed` | The hand-in did not go through: {why} | Hand in, Stop | V `desk.go:589` |
| `prereq_regressed` | {part} no longer reads as done, and {quest} needs it. | Run it again, Stop | none (bug 5) |
| `prereq_missing` | {part} has to be done first and cannot start now: {why} | Stop | deadlock (bug 3) |
| `quest_locked` | {quest} is locked again: {other quest} no longer reads as done. | Stop | shown as running (C1) |
| `hold_timeout` | Still waiting after {n} minutes: {the hold's sentence} | Keep waiting, Stop | none |

### Nobody can fix it here

| code | the player reads | buttons |
|---|---|---|
| `game_failed` | Failed in this game. It can be done again in the next difficulty. | none |
| `unsupported` | Not supported yet: {why} | Stop, when a run exists |

### Why Start is not offered (a chunk without a run)

| code | the player reads on the row | Start button |
|---|---|---|
| `locked_quest` | Locked until {other quest} is done. | absent |
| `starts_with` | Starts with: {parts}. | enabled (D1: accepts the chain) |
| `too_hard` | Too hard for now: {why} | absent |
| `readiness_unknown` | Not judged yet. | absent |
| `missing` | {why} (a readiness fact, such as a level) | absent |
| `busy` | {running quest} is running; stop it first. | disabled, this text as the hint |
| `paused` | Paused: resume the bot first. | disabled, this text as the hint |
| `handin_ready` | The field work is already done: Start goes straight to the hand-in. | enabled (case C5 of appendix A) |
| `board_off`, `no_hero` | as today (V `desk.go:333,343`) | absent |

## B.2 API semantics

All verbs: `POST /api/quests/<quest>/<verb>` with `{"chunk": "<id>"}` (V routes `internal/watch/questsapi.go:11-20`). A missing chunk means: for start, the first startable chunk of the quest; for the others, the quest's head run. Reply fields today: `ok`, `already`, `pending`, `code`, `reason`, `quest`, `chunk`, `state` (V `desk.go:86-98`); added: `changed: [{quest, chunk, state, reason: {code, text}}]`.

| verb | target condition | when it changes something | `already` (200, nothing changed) | refusals (code, status) |
|---|---|---|---|---|
| `start` | a live run of the chunk exists | creates the run and its unmet `after` chain in one save; for `handin_ready` creates it as `awaiting` | a live run exists (any status) | `done` 409; `locked_quest`, `too_hard`, `readiness_unknown`, `missing`, `unsupported` 409 with the row's sentence; `busy` 409 naming the running quest; `paused` 409; `prereq_not_startable` 409 naming the part and its code |
| `cancel` | no live run of the chunk or of its dependants | closes them as `stopped`; stops the running slice; clears the counters, the task executor group and a pending hand-in request; a countdown of the auto-start is cancelled as today | no live run and no countdown: "Nothing of {quest} is started." | `done` 409 only when the chunk is finished and nothing depends on it |
| `retry` | the run is not parked | `parked` to `queued` or head; clears the board counters and the task executor group; for `prereq_regressed` with `{"again": true}` accepts the regressed part | the run is live and not parked: "{quest} is not stuck." | `not_started` 409 when no run exists; `game_failed` 409 |
| `handin` | a hand-in was attempted for the awaiting run | `awaiting` to `handing_in`; in the field it holds `handin_in_town` and runs at the next town visit (V today `bot/questboard.go:873-879`) | one is in flight (202); the chunk is finished (200, settles an unpaid award as today, V `desk.go:828-833`) | `not_awaiting` 409 with the run's reason; `paused` 409 |
| `auto` | the opt-in equals `on` | writes the opt-in | equal already | `auto_never` 409 |

Rules that hold for every verb:
1. The verb reads and writes the stored runs. The derived state is used only to refuse a Start.
2. Applying a verb twice gives the same file as applying it once (I7).
3. A refusal never changes the file. A reply with `ok: false` always has `code` and `reason`.
4. An auto-start goes through the same `start` transition with `by: auto`; it never accepts a chain whose members are not all `safe` (the auto rule "only safe", V `auto.go:215-241`, applied to each member).

## B.3 Invariants, scenarios, checker rules

### Runner additions the scenarios need (migration card 1)

- Assert kind `board`: `{kind: "board", quest, chunk, state, reason}` evaluated on a `GET /api/quests` taken at the end of the run (and, with `from_min`/`to_min`, on samples taken every 30 virtual seconds). Today's kinds are in V `internal/scenario/state.go:82-100`; none reads the board.
- Assert kind `api_reply`: `{kind: "api_reply", index, status, ok, already, code}` on the recorded reply of `trigger.api[index]` (the calls exist, V `internal/scenario/scenario.go:149-158`; U whether replies are recorded today).
- Trigger `restart_at_s`: stop and start the bot process once at a virtual time, same game and hero (U whether the runner can do this; if not, the restart scenarios are split in two runs sharing a state directory).
- No scenario writes game state after the start (CRUTCH RULE): exact-evidence regression (R3, R4) is therefore tested offline only, with `questboard.Fake`.

### Scenarios (ids for `tools/provingground/scenarios/board.json`)

| id | setup | trigger | assertion sketch | invariant |
|---|---|---|---|---|
| `board.two-accepted-chunks` (exists) | records scroll and stones | none | `item_in_pack bks or bkd`; `log_line "parked" max 0` | I3 |
| `board.stop-parked-chunk` (exists) | as above | cancel stones at 90 s | `api_reply ok, already false`; `board stones state ready or locked` | I7 |
| `board.prereq-not-accepted` (boardprereq) | record stones only | none | `log_line "stones waits for scroll"`; `board scroll state in_progress or done` | I2 |
| `board.prereq-start-tap` (boardprereq) | none | start stones | `api_reply ok`; `board scroll state accepted or in_progress`; `board stones reason waits_for` | D1 |
| `board.sm.stop-cascades` | records scroll, read, stones | cancel scroll | `api_reply changed has read and stones`; `board stones state ready or locked`; `log_line "search_for_cain" run lines max 0 after the stop` | D2 |
| `board.sm.stop-idempotent` | record scroll | cancel scroll twice | second `api_reply ok, already true`; plan file equal after both | I7 |
| `board.sm.start-idempotent` | none | start scroll twice | second `already true`; one run in the view | I7 |
| `board.sm.one-head` | records of the Cain chain | none, 30 virtual min | sampled `board`: never two chunks `in_progress` or `awaiting_handin` | I1 |
| `board.sm.busy-second-quest` | record scroll | start den_of_evil | `api_reply ok false, code busy`; reason names Search for Cain | D5 |
| `board.sm.restart-mid-chain` | none | start read (chain), restart after the scroll is taken and read | `area_changes_le` bound; `log_line "step Take the Scroll" max 1`; `board scroll state done` after the restart | I4, I6 |
| `board.sm.regress-proxy` | hero in Tristram with stones done | town trip (the portal and key are gone) | `board stones state done`; `log_line "step Touch the cairn stones" max 0` | I6 |
| `board.park.persist-restart` | chunk that parks (too-weak hero, death cap) | restart after the park | `board state blocked reason died_too_often` after the restart; `log_line "starting chunk" max 0` after it | I9 |
| `board.retry.runs-a-step` | parked chunk, cause removed in setup | retry | `log_line "quest board: step"` min 1 after the retry | A.4 row 1 |
| `board.hold.restock` | Countess chunk, 3 healing potions, gold 5000 | start | `log_line "hold needs_potions"`; `log_line "parked" max 0`; hero reaches Black Marsh (`area_in`); `no_writes` | I10, bug 4 |
| `board.hold.restock-stash-gold` | as above, purse 0, stash 5000 | start | `log_line "needs_gold_withdraw"`; `stash_between` lower than at the start; `log_line "parked" max 0` | I10 |
| `board.hold.broke` | purse 0, stash 0, 3 potions | start | sampled `board reason needs_gold_farm`; never a live run without a reason | I5, I10 |
| `board.hold.too-weak` | hero under the guard floor | start (risky offer) | `board reason too_weak or route_closed`; the level rises (`level_at_least`); `log_line "parked" max 0` inside the budget | I10 |
| `board.hold.timeout` | as `broke`, budget 70 virtual min | start | `board state blocked reason hold_timeout` | I10 |
| `board.view.reason-always` | the Cain chain run | taps | sampled every 30 s: a live chunk that is not `in_progress` with a moving step has a `reason.code` of B.1; the running card only while the bot log shows a quest step | I5, I11 |
| `board.handin.field-tap` | awaiting chunk, hero sent to the field by an order | handin | `api_reply 202`; `board reason handin_in_town`; later `quest_bits` set | B.2 |
| `board.handin.without-run` | pending bit exact, no record (C5) | start | run is `awaiting`; Hand in card present | C5 |
| `board.auto.chain-safe-only` | auto on, part 1 safe, part 2 risky | none | part 1 starts by itself, part 2 does not | B.2 rule 4 |

Each scenario runs on two builds (Paladin and Sorceress, the matrix the existing board scenarios use, V `cain.json`), per the build-flexibility rule.

### Offline unit tests (card 2), one per invariant

- I8 totality: for every run status and every event of 3.3, `Step(status, event)` returns a status and a reason code from the list; no panic, no empty code.
- I7: for every reachable plan in a generated set (all subsets of the Cain chain's runs, each status), `apply(v, apply(v, p)) == apply(v, p)` for start, cancel, retry.
- I2 and I1: after every verb and every event on that set, closure and one-head hold.
- I9: `decide(load(save(p))) == decide(p)` with the head recomputed.
- R3, R4: with `questboard.Fake`, clear an exact quest word under a live dependant: the head parks with `prereq_regressed`, nothing runs, `retry {again}` accepts the part.

### questspec checker rules (card 8)

| rule | check | why |
|---|---|---|
| 12 | in every quest the chunk order of `data/questboard` is a topological order of `after` | the head is chosen by data order (V `exec.go:224-231`); `def.go` checks cycles only |
| 13 | a chunk whose last chain step has only volatile readers for its posts (caps `volatile`) is marked `latched` in the chain data, and the list of such chunks equals the list the board latches | the `reached` patch was found by a lab drill; the checker can find the next one from data (V `docs/questspec.md`: readers carry `volatile`) |
| 14 | every guard reason a step's `Ready` can return maps to a code of B.1 with a remedy (a table test over the fault producers in `internal/task/gate.go` and the per-act gates) | bug 4: a refusal without a type |

## B.4 What the player sees (the existing board panel)

The panel keeps its layout: the Now row (shared with the strip), the mode bar, the acts with their quests and parts, history, stopped (V `panel.js:233-264`). Changes are in what each state says and offers.

### The Now row and the strip

| run status | label | head | sub line | primary | secondary |
|---|---|---|---|---|---|
| `running` | Running | Running: {quest} | part k of n, step k of n, {step}, next: {npc} | Stop (two taps) | none |
| `held`, remedy chore, wait or train | On hold | {quest}: on hold | the hold's sentence (B.1) | Stop (two taps) | none |
| `queued` head before its first step | Starting | {quest} is starting | the hold's sentence, or "Starts at the hero's next free moment." | Stop (two taps) | none |
| `awaiting` | Waiting for you | {npc} is waiting: hand in {quest} | The hero stands there and waits for your tap | Hand in | Stop (two taps) |
| `handing_in` | Handing in | Handing in {quest}... | Talking to {npc}. The picture shows it | Handing in... (disabled) | none |
| `handing_in`, `confirming` | Handed in | Handed in to {npc}. Checking that the game agrees | {quest} | none | none |
| `parked`, retry possible | Stuck | {quest} is stuck | the park's sentence | Try again | Stop (two taps) |
| `parked` `prereq_regressed` | Needs you | {quest} cannot continue | {part} no longer reads as done | Run it again | Stop (two taps) |
| `parked` `prereq_missing`, `quest_locked` | Needs you | {quest} cannot continue | the sentence | Stop (two taps) | none |
| `parked` `hold_timeout` | Needs you | {quest} is still on hold | the hold's sentence | Keep waiting | Stop (two taps) |
| `parked` `game_failed` | Failed | {quest} failed in this game | It can be done again in the next difficulty | none | none |
| no run, countdown | Starting | as today (V `panel.js:155-158`) | | Cancel | none |
| no run | Offer, Training, Idle | as today (V `panel.js:160-176`) | | Start | none |

New against today: the "On hold" row (today a held chunk shows "Running" with a step that does not move), Stop on the awaiting row, and the three Needs-you rows that have no Try again.

Under the Now row, when further runs are queued: one line "Then: {part}, {part}" (the `queue` of the view; today `extra`, not shown).

### Part rows in the quest list

| state of the part | word on the row | reason line | button | disabled when, with the hint |
|---|---|---|---|---|
| `ready` | safe or risky | the readiness sentence; minutes; pays | Start | paused: "Paused: resume the bot first"; another quest runs: "{quest} is running; stop it first" |
| `locked`, only by earlier parts (`starts_with`) | after {part} | Starts with: {parts} | Start | as `ready` |
| `locked` by another quest | locked | Locked until {other quest} is done | none | |
| `available` | too hard, or not yet | the row sentence of B.1; levels to go in words | none | |
| `accepted` (queued) | queued | Starts after {part} | Stop (two taps) | |
| `in_progress` | running, or on hold | the step, or the hold's sentence | Stop (two taps) | |
| `awaiting_handin` | waiting for you | {npc} is waiting | Hand in | paused: "Paused: resume the bot to hand in" |
| `blocked` | stuck, or needs you | the park's sentence | the Now row's buttons | |
| `done`, not confirmed | done, confirming | Checking that the game agrees | none | |
| `done` | done | the Ether paid, if any | none | |
| `skipped` | skipped | Skipped in this difficulty, or by you | none | |

Rules for the page:
1. A button that is shown but cannot act is disabled and its `title` and the line under it carry the reason; a button is never shown for a verb the server would answer with `already`.
2. The page chooses buttons from `reason.code` and `state`, never from the sentence.
3. Every live part shows either a moving step or a reason sentence. A live part with neither is a bug (I5); the page then shows "No reason was given (report this)" so it is visible in a screenshot.
4. After a Stop that cascaded, the toast reads: "{quest} stopped. Also stopped: {parts}." (V `agent/boardprereq` already words it, `desk.go:786-792` at `e77b53911`).
5. After a Start that accepted a chain: "{part} started. First: {parts}."
6. Needs-you rows carry the same sentence and the same buttons as the Now row (V today `desk.go:536-555`); a hold never raises a Needs-you row, a park always does.
