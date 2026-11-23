# Appendix A: evidence per bug (quest board review 2026-11-21)

> status: appendix of [quest-board.md](quest-board.md). Code at `origin/main` `5b12ce5e7` unless a line says `e77b53911` (`agent/boardprereq`). **V** = I read the cited lines; **S** = from a probe doc or card; **U** = unverified. No test and no game was run for this review.

Paths are under `internal/` unless they start with `docs/`, `data/` or `tools/`. `qb/` is `internal/questboard/`, `bot/qb.go` is `internal/bot/questboard.go`.

## A.1 How one chunk state is computed (the order matters)

`deriveChunk` returns at the first rule that matches (V `qb/state.go:195-261`):

| order | rule | result | lines |
|---|---|---|---|
| 1 | quest word done | `done` | 201-203 |
| 2 | chunk's own check done, not denied by a memory read | `done` | 204-206 |
| 3 | quest failed by the game | `blocked` | 208-212 |
| 4 | skipped by the player | `skipped` | 213-216 |
| 5 | record exists and the executor parked it | `blocked` | 217-221 |
| 6 | a `requires` quest is not done on T1/T2, or an `after` chunk is not done | `locked` | 222-225, 273-290 |
| 7 | no handler | `available` | 227-230 |
| 8 | ends handin and (pending bit exact, or record and objective met) | `awaiting_handin` | 233-239 |
| 9 | record exists | `accepted`, or `in_progress` when "running" | 240-247 |
| 10 | no record: readiness | `available` or `ready` | 248-260 |

Consequences read off the table (all V):
- A record does not protect a chunk from rules 6 and 7. An accepted chunk can derive `locked` or `available`; `Next` then takes the default branch (`qb/exec.go:495-499`).
- Rule 8 needs no record when the pending bit is exact: a chunk nobody started can derive `awaiting_handin`.
- "Running" in rule 9 is `any step satisfied OR the process started it` (`qb/exec.go:168-176`), not "the hero is on it".
- A readiness fact (`Missing`) is read only in rule 10, so only for a chunk without a record (`qb/state.go:251-252`).

## A.2 The six bugs

### Bug 1: "more than one chunk is accepted: parked, not running any"

- Before #1012 `Next` refused any plan with two open records (V removed lines in `git show b28a34020 -- internal/questboard/exec.go`).
- How two open records can exist although `acceptLocked` refuses a second one (V `qb/exec.go:310-312`):
  1. A record is never removed when its chunk completes; `Store.Cancel` is the only removal (V `qb/plan.go:292-306`). "Finished" is derived at the moment of the next accept (V `qb/exec.go:316-319`, `qb/plan.go:270-281`).
  2. Completion of an item or auto chunk is a proxy that lapses; the in-process `reached` map covers it until the process ends (V `bot/qb.go:534-544`). Completion of a hand-in at T3 is the in-process `handed` flag (V `bot/qb.go:528-530,905-907`).
  3. So: part 1 done by proxy, part 2 accepted (allowed, part 1 counts as finished), restart or lapse, part 1 reads open again: two open records.
- U: which path the sandbox hero of 2026-11-20 took. The scenario that pins the state writes the two records directly through the store (V `cmd/d2scenario/run.go:451-458`, comment "a plan file that holds two is the state under test").
- Fix #1012 (V `qb/exec.go:224-231`): sort open records by data order, run the first, list the rest as `extra`.

### Bug 2: Stop on a queued chunk answered "not running" and withdrew nothing

- `Desk.Cancel` chose by derived state; a queued chunk derives `locked` (rule 6) and fell into the "nothing to stop" reply (V old lines in `git show b28a34020 -- internal/questboard/desk.go`).
- Fix #1012: ask the store first (`HasRecord`, V `qb/desk.go:771-778`, `qb/exec.go:336-341`).
- Same shape still present in `Desk.Retry`: it acts only when the derived state is `blocked` (V `qb/desk.go:798-801`).

### Bug 3: stones waits for ever for "after Have Akara read the scroll"

- Data: `stones` is `after: [read]`, `read` is `after: [scroll]` (V `data/questboard/act1.json`, quest `search_for_cain`).
- Record for `stones` only (the player stopped `scroll`). `stones` derives `locked` with the reason "after Have Akara read the scroll" (rule 6, V `qb/state.go:284-288`). `open()` returns it (V `qb/exec.go:216-222`), `Next` answers `DecBlocked` (V `qb/exec.go:495-499`), `boardStep` logs "parked" once and runs the mode (V `bot/qb.go:680-683`). Nothing can change the evidence, because nothing runs.
- On the page (V): `Status.Active` is the locked chunk; the view shows it as the running card (`qb/desk.go:417-424`); no Needs-you row (`qb/desk.go:545`); Try again says "is not stuck" (`qb/desk.go:798-801`).
- Fix (V `e77b53911`, merged as queue entry #1019 while this review was written): accept of a chunk locked only by `after` accepts the chain (`qb/exec.go:310-319`); `Next` repairs a lone later record (`qb/exec.go:624`); Stop withdraws dependants (`qb/exec.go:492-501`, `qb/desk.go:779-793`).

### Bug 4: the route guard blocks a step (potions under 4), `needs_attention`, no restock

- The guards: `fieldReady` returns `Fault{Kind: KindDanger}` with no `Insert` for a banned area, the route guard, the tower guard and the gear gate alike (V `task/gate.go:263-284`). The guard text carries the potion count (S "tower guard (3 healing potions < 4)", `docs/probes/provingground-countess-2026-11-20.md` finding 4 and "Earlier live result").
- The executor: a danger fault without an insert is retried, then escapes to a block of the group with `needs_attention` (V `task/executor.go:517-520,542-560,476-486`).
- The board: `OutRetry` or `OutBlocked` is `StepFailed`; a danger refusal makes the tick idle and rests (V `bot/qb.go:747-752,773-783`); three counted failures park the chunk (V `qb/exec.go:513-520`).
- A potion shortage is handled correctly elsewhere: `gatePotions` and `gateReady` return `KindResource` with a chores step (V `task/act1.go:99-110`, `task/gate.go:249`). The field guards do not.
- Design intent not met: "a wall refuses (kit missing, no merc, route guard): not blocked ... buying them" (V `docs/design/quest-board-2026-11-16.md` section 2.3).
- S: 25 virtual minutes idle in town with 3 healing potions and gold 700 after banking 4160 (same probe doc); the banking side is the related finding of `docs/probes/provingground-act1more-2026-11-20.md` (bank chore deposits the purse, the chores step cannot pay).

### Bug 5: quest evidence regresses with a downstream chunk accepted

- S (brief): a record bit went back while a later chunk was accepted. I did not see the log.
- V in code: there is no rule. A regressed prerequisite turns the later accepted chunk `locked` (rule 6) and the page shows it as running (bug 3 shape); a regressed finished chunk with its record still in the file becomes an open record again and, being earlier in data order, runs again by itself (`qb/exec.go:214-231`), which is the second scroll walk of lab drill CAIN-L3 (V comment `bot/qb.go:536-539`).

### Bug 6: the player cannot tell why nothing runs

| what the bot knows | where it is produced (V) | reaches the page? |
|---|---|---|
| why an accepted chunk has not started (fight, recovery, paused, parked) | `bot/qb.go:1004,1021,1042,1045` to `View.Running.Queued` and `Strip.Queued` (`qb/desk.go:414-416,425`) | no: neither panel reads `queued` (V grep, 2026-11-21) |
| further accepted chunks and what they wait for | `View.Extra` (`qb/desk.go:428-430`) | no: not read by the panels (V grep) |
| the decision is `blocked` for a non-park reason | `Decision.Why` (`qb/exec.go:495-499`) | no: only the bot log (`bot/qb.go:681`) |
| the step was refused, counted 1 or 2 of 3 | `runState.lastWhy` (`qb/exec.go:513-516`), `View.Running.Failures` | the count is in the view; the panel does not show it or the reason (V `panel.js:149-153`) |
| auto-start held back | `AutoView.Why` (`qb/desk.go:353`) | U: not checked in the panel |

## A.3 Five combinations that are still undefined after #1012 and boardprereq

| id | combination | what happens (V unless marked) | lines |
|---|---|---|---|
| C1 | record, and a `requires` quest no longer reads done on T1/T2 | derives `locked`; `Next` default branch; page shows the running card; `boardprereq` repairs only `after` locks | `qb/state.go:274-283`, `qb/exec.go:495`, `qb/desk.go:417-424` |
| C2 | record, and the chunk has no handler (data changed, `unsupported` set) | derives `available`; as C1 | `qb/state.go:227-230` |
| C3 | finished record of an earlier part reads open again after a restart | it is the first open record in data order and runs again | `bot/qb.go:534-544`, `qb/exec.go:224-231` |
| C4 | finished record of ANOTHER quest reads open again while a second quest runs | both are open; the earlier quest in data order becomes active, the running one is listed as "waits for" it | `qb/exec.go:214-231`, `qb/desk.go:428-430` |
| C5 | no record, pending bit exact | chunk row says `awaiting_handin`, but there is no active chunk, so no Hand in card and no Needs-you row; the API accepts a hand-in for it; no award (never `touched`) | `qb/state.go:233-239`, `qb/exec.go:433-443,610-618`, `qb/state.go:348` |

## A.4 Smaller findings

| finding | evidence (V) |
|---|---|
| Try again clears the board's counters only; a group blocked in the task executor makes the plan `Stuck` again | `qb/exec.go:393-397`, `bot/qb.go:632-634`, `task/executor.go:291-292`; `Unpark` callers: `bot/staletarget.go:163,244` only (U whether another path clears it) |
| Stop does not clear `reached` nor the task executor group | `bot/qb.go:607-614` |
| A park is lost at a restart and the chunk is tried again | `qb/exec.go:92-93` |
| `Next` runs up to three times per tick from different branches | `bot/qb.go:660,937,992`; `bot/goal.go:121,128,131` |
| Data order is assumed to be a topological order of `after`; only cycles are checked | `qb/exec.go:224-231`, `qb/def.go:466-496`; the chain order check of questspec rule 11 compares chain steps with chunk order, not `after` (`questspec/board.go:39-41`) |
| The step marks on the strip are a share of satisfied handler steps spread over the named steps, and "running" is derived from them | `bot/qb.go:510-516`, `qb/exec.go:168-171` |
| An awaiting hero stands at the giver without limit; `quests.handinIdleMin`, `quests.resumeWindowMin`, `quests.retryMax` of the design are not knobs | `bot/qb.go:848-867`; grep of `env/tuning_quests.go` |
| `boardprereq`: chain accept returns before `CanStart`; the tapped chunk's readiness is never looked at | `qb/exec.go:310-319` at `e77b53911` |
| `boardprereq`: logic keyed on the display text `"after "` | `qb/exec.go:367` at `e77b53911` |

## A.5 Plan file shapes the v1 import must handle (migration card 3)

| shape | v1 content | v2 result |
|---|---|---|
| S1 | one record, chunk not done | one run `queued` (becomes the head) |
| S2 | records whose chunks read done | runs `finished` with a latch at the tier read now |
| S3 | finished earlier part and a live later part | `finished` plus latch, and `queued` |
| S4 | a lone later part (its prerequisite has no record) | D4 repair: add the chain when startable, else `parked(prereq_missing)` |
| S5 | open records of two quests (C4) | the quest of the newest record stays live; the other quest's open records become `stopped` with a log line and a "stopped by the update" history row (Q7 decides if they queue instead) |

Records that name an id the data no longer has are a load error today (V `qb/plan.go:155-190`); the import keeps that.

## A.6 What I did not check

- No run of `go test`, no Proving Ground job, no live game. The scenarios `board.two-accepted-chunks`, `board.stop-parked-chunk`, `board.prereq-not-accepted` and `board.prereq-start-tap` are reported as not run on the builder by their cards (S TASK-703 and TASK-709 `result:` lines).
- The sandbox report of 2026-11-20 and the log of the evidence regression (bug 5).
- The readiness model (`qb/readiness.go`), the Ether ledger and the audit ring beyond their interfaces.
- Acts 2 to 5 data: the review reads Act 1 chains only; the model is meant for all acts but multi-part quests there (staff, Khalim, seals, Ancients) were not walked through it.
