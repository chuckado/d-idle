# Quest board model (slice S1) and executor (slice S3)

> status: S1 model built and tested offline (card quests1, TASK-577); S3 executor built and tested offline behind the knob `quests.board` (default 0; card quests3, TASK-585), NOT verified live (the drill below has not run). Design: [docs/design/quest-board-2026-11-16.md](design/quest-board-2026-11-16.md). Tags: V verified by the named test or file, S secondary, U unverified live, I design choice.

## What is here

| piece | where | notes |
|---|---|---|
| definitions | `data/questboard/act1.json`, `act2.json`, `families.json`; embedded by `data/questboard/embed.go` | schema `quests.def/1`; strict decode (an unknown field is an error) |
| loader, validation | `internal/questboard/def.go`, `load.go` | unknown class, act, quest, chunk, difficulty, kind, risk: an error naming it (V `TestParseErrors`, `TestLookupsNameTheUnknown`); no silent fallback |
| states, derive | `internal/questboard/state.go` | nine states, `Derive`, `CanStart`, the `Transitions` table with the evidence tier each needs |
| plan records | `internal/questboard/plan.go` | per hero `<dir>/<hero>.quests.json` through `internal/store.File` (tmp, fsync, `.bak`, quarantine, future version refused) |

Quest ids are the save ids (`internal/saves` questTable; V `TestQuestIDsAreSaveIDs` reads that file). `taskId` carries the task-layer id where it differs (`sisters_burial`, `sisters_to_slaughter`).

## Rules the code enforces

- Derive is a pure function of its input: derive twice, same result (V `view` helper in every test). Handler progress is never stored; the only store is the player's decisions (accepted, skipped, auto-start opt-ins, seen toasts).
- Evidence: a quest prerequisite unlocks only on T1 (save) or T2 (memory); T3 (proxy) never unlocks (V `TestTransitionTableEvidence`, `TestChunkOrderAnyThenAssemble`). T3 may stop the hero at the NPC and show "done, confirming"; it pays nothing until T1/T2 confirms.
- Evidence conflict (I): for a single-chunk quest a fresh T2 "not done" beats a T3 "done"; a stale T1 "not done" does not beat a proxy (shown as confirming). A part of a multi-chunk quest has no bit of its own, so its proxy stands.
- Chunk order (I): a chunk's `after` is satisfied by any evidence of the earlier chunk (an offer skips no work, the handler re-derives); a chunk's or quest's `requires` need T1/T2.
- Start tap: `CanStart` refuses anything that is not `ready` (locked, out of order, available, done); `Store.Accept` refuses a second accepted chunk in the same difficulty.
- Readiness: S2 is not built. Without a readiness word a chunk stays `available` ("readiness not known"); the tests stand in for S2 with the legacy floor per class family (clvl by melee, caster, summoner) on Paladin and Sorceress.
- Meta currency (design section 7): `Award.Key` is `quest/<quest>/<chunk or ->/<difficulty>`, never the hero (once per account per quest per difficulty). A quest the board never started (`byHand`) pays nothing: no backfill (user decision 2026-11-16). T3-only completion pays nothing yet. The amounts are data placeholders (U) in the `meta` columns, about 5 for a first Normal story quest and 2 for an optional chore.
- `autoStart: never` rows (hazard risk) refuse the opt-in (V `TestStoreSkipAndAutoStart`); the loader requires `autoStart never` on every `hazard` quest.
- Build flexibility: no class or skill literal in Go; the class-to-family table is `families.json`, the floors are the `legacyFloor` column per quest.

## The data and where it came from

Act 1 merges `data/quests/act1-hints.json` (giver, clvl floors, EXP/UNSURE tags), `act1-givers.json` (accept and hand-in picks) and `docs/goals/execution.md` section 3 (steps, boss, rewards, Req); the per-field `tags` map names the source of each field. Act 2 has no hint or giver table, so it comes from execution.md section 3 and docs/progression/quests.md alone and every row is U until a live drill. Chunks: Cain's chain is four (scroll, read, free Cain, report), the Horadric Staff is three (cube, shaft in any order, assemble after both and the Tainted Sun amulet). The old hint and giver files stay until S9 retires them.

Known gaps (U): the hand-in picks of Akara, Atma, Cain, Charsi (`imbue`), Warriv (`Go East`) and Meshif (`Sail East`) are not in `internal/talk/tables.go` as read here; `minutes` are my estimates (I); the Andariel `softAfter` the Burial quest is the hint table's UNSURE availability, kept soft on purpose so it cannot lock a player out; whether Radament gates the Staff (execution.md says V, kept as a hard `requires`).

## For S2 to S4

S3 builds `Facts` from `quest.Reader` (`Info.Status`, `Pending`, `Evidence` map to `Fact` and `Tier` by value: `EvProxy`=1, `EvSave`=2, `EvMemory`=3) and `ChunkFact` from the handler checks; S4 calls `CanStart` then `Store.Accept`. No knobs were added (the design's `quests.*` knobs belong to S3).

## Executor (slice S3)

Knob `quests.board` (default 0, docs/knobs.md). With 0 nothing here runs: `Exec` answers `ErrOff`, `boardStep` is never called, the planner plays quests as before (the only change on that path is the executor-creation block of `taskStep` moved into `ensureTaskExecutor`, same code; V by the unchanged `internal/bot` and `internal/task` suites, `TestQuestBoardKnobOffKeepsThePlanner`).

| piece | where |
|---|---|
| API-free interface for S4 | `questboard.Exec`: `Offers()`, `Accept(q, chunk)`, `HandIn(ctx, q, chunk)`, `State()`, plus `Cancel` (the player's Stop), `Reset`, `Pause`, `Unblock` (Try again), `SetEnabled`, `SetMode` |
| loop interface | `Next()` says idle, run, await, blocked, paused or busy; `Report(decision, StepResult)` feeds a slice back; `Tick(ctx, stepper)` is both for drivers without a guard |
| seams | `Env` (difficulty, facts, per-chunk handler checks, readiness) and `Actions` (hand-in dialogue, stop); `questboard.Fake` implements both over a scripted world for tests and the S4 API |
| bot adapter | `internal/bot/questboard.go`: `Runner.NewQuestBoard(board, store)` builds the Exec, `boardStep` is the loop tick, hooked in `goal.go` before `taskStep` |
| plan store | `Store.AcceptIf` lets a finished chunk's record stay (proof the board ran it, awards) without blocking the next Start |

Flow (tests: `internal/questboard/exec_test.go` on a Paladin and a Sorceress with the Fake, `internal/bot/questboard_test.go` on the real runner with the campaign rig):

1. `Offers` lists the Ready chunks (`Startable` false while a chunk is accepted). Readiness is `Board.LegacyReadiness` (the S1 legacy floor per class family, safe at the floor, risky within 2 levels) until S2.
2. `Accept` records the tap for the CURRENT chunk only. It refuses a chunk that is not Ready and a second accepted chunk, naming the one that runs (V `TestAcceptRunAwaitHandIn`, `TestTwoAcceptedRecordsNeverRun`, `TestStartCoversOnlyTheCurrentChunk`). The same tap twice is a no-op.
3. The loop calls `Next`. Accepted or in progress: one slice of the quest's task-layer handler (`task.Handler`) minus its hand-in step, run by the shared task executor inside the runner's guard (same `r.do` reactions, death policy and watchdog as a campaign step). The chunk is awaiting when the quest stage reads Cleared or better and no needed handler step is left (`ChunkFact.Objective`), or on the pending bit with T1/T2.
4. Awaiting: the hero walks to the giver (one open-and-close of the dialog, once per chunk and process) and rests in town. It never talks the reward by itself (V `TestQuestBoardAcceptRunStopHandIn`).
5. `HandIn` (the player's tap) needs the chunk Awaiting. The dialogue runs on the loop goroutine (a one-slot mailbox, so the guard applies), then the board re-derives: Done with T1/T2 is `done` and returns the meta awards; a proxy-only Done (the dialogue ran, the save has not caught up) is `confirming` and pays nothing until T1/T2 agree; a quest word that still reads not done retries (at most `MaxHandIn` 3 per tap), then parks the chunk with a reason. A repeated tap after success returns `Already` and runs no dialogue. A fresh T2 "not done" beats a proxy "done" (S1 rule, V `TestHandInReadBack`).

Retry, park, resume:

| situation | behaviour | test |
|---|---|---|
| a slice fails | counted at most once per 15 s in the bot (`boardFailGap`, so a gear gate does not park in 3 s); 3 in a row park the chunk with `gave up after 3 tries: <why>` | `TestFailureDeathParkResume` |
| nothing left to run but the quest does not read finished | counted as a failure (no silent loop) | bot adapter, U |
| the hero dies | the bot's own recovery runs; the chunk resumes from the evidence; the second death on one chunk parks it (`died 2 times`) | same |
| parked | state Blocked with the reason (Needs-you row in S5); the hero trains meanwhile; `Unblock` or a restart tries again | same |
| restart | no handler progress is stored; counters and parks are in memory; the chunk is re-derived from the quest word, the handler checks and the plan record (InProgress shows from any satisfied step) | `TestRestartAtEveryStep` |
| Reset, knob off | `Reset` clears counters and parks and calls `Actions.Stop`; the record stays accepted | `TestResetPauseCancel`, `TestSwitchOffKeepsRecords` |
| Pause | a running chunk holds; an awaiting one stays awaiting | `TestResetPauseCancel` |
| Stop (`Cancel`) | the record goes, the chunk is Ready again | same |

Between chunks (nothing accepted, parked or paused-free) the hero trains: one `task.StepLevel(clvl+1)` slice per tick, the same path the campaign's level step uses. The modes (train, idle, farm) are slice S6.

Limits and unknowns:

- Only single-chunk quests with a task-layer handler run (U for the rest): a quest without a handler or with several chunks (Cain, the Staff, Khalim's pieces) has readiness unknown, so it is never offered; per-chunk handlers are slice S7. Act 1 quests have handlers; Act 2 to 5 handlers exist in `internal/task/act2..5.go` and run the same way once their data rows are offered, none drilled.
- The hand-in step of a handler is the Talk step whose id ends in `/reward` (V by grep, ten handlers). Quests whose hand-in is another talk (Alkor, Tyrael, Jerhyn rows) are S7.
- `ChunkFact.Steps` marks the share of satisfied handler steps over the board's named steps (I: the names do not match the handler ids); it drives the strip text only, never a decision.
- A mid-walk cancel is not implemented: Stop and Reset take effect at the next slice boundary (U).
- The reader is the save by default (`QuestsLive` false), which only changes on Save and Exit: a hand-in then reads `confirming` until the next save, and a restart before it shows the chunk awaiting again (the second tap talks the NPC once more; harmless for a done quest, U).
- Wiring: `cmd/d2bot` does not construct the board yet; S4 calls `Runner.NewQuestBoard` next to the API. Until then `boardStep` with the knob on and no board only trains (logged once).
- Not in S3: auto-start (S8), the meta ledger (S10; `HandInResult.Awards` carries the keys), modes (S6), the API (S4).

## Live drill, Act 1 (checklist, NOT run)

Throwaway character (test-characters rule: create it normally, no setup writes if possible; log it in docs/probes/test-characters.md), one hidden instance, `tools/live.sh`, `quests.board` = 1, a board wired through S4 or a drill command that calls `Offers`, `Accept`, `HandIn` on the runner. Writes after start must be 0 (`d2cheat ledger --since-run <run>`).

1. Fresh Normal character, clvl 1 to 3, Rogue Encampment. Knob 1, campaign goal. Expect: the hero trains, no Akara talk, no quest travel, `Offers` lists Den of Evil as Ready (Paladin from clvl 1, Sorceress from clvl 2).
2. Den of Evil: `Accept`. Expect the log `quest accepted`, the walk to the Den, the clear, the return to camp, a single Akara dialog open and close, `Next` = await. Expect NO reward talk. Stand 60 s: nothing happens. Then `HandIn`: expect one Akara reward dialog, the skill point appears (check the unspent skill points), outcome `done` if the memory reader is live, `confirming` with the save reader. Tap again: `already`, no dialog.
3. Restart the bot while the Den clear runs and again while awaiting: it must resume or stand awaiting without redoing the clear.
4. Death: let the hero die once in the Den (or note the first natural death): after recovery the chunk resumes; a second death parks it with a reason.
5. Blood Raven: same from clvl 6 to 8 (Paladin floor 7, Sorceress 8; below it the offer must be absent or risky). `Accept`, Burial Grounds, Blood Raven, Kashya awaiting (Kashya row U: the approach open-and-close may fail; record it), `HandIn`, the Rogue claim is a separate step and must not run by itself.
6. One chunk at a time: with the Den running, `Accept` Blood Raven must be refused naming the Den.
7. Evidence: note which tier (T1 save, T2 memory) answered at each step in the log, and whether the hero stood next to the NPC.
8. Record in docs/probes: tiers, durations per chunk, whether the approach dialog opened, any parked chunk and its reason. Leave the hero in town, no panels open; stop the instance with `tools/game.sh stop`.
