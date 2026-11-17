# Quest board model (slice S1), executor (S3) and API (S4)

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

S3 builds `Facts` from `quest.Reader` (`Info.Status`, `Pending`, `Evidence` map to `Fact` and `Tier` by value: `EvProxy`=1, `EvSave`=2, `EvMemory`=3) and `ChunkFact` from the handler checks; S4 (done, section "API and stores") calls `CanStart` then `Store.Accept` through `Desk`. No knobs were added (the design's `quests.*` knobs belong to S3).

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

1. `Offers` lists the Ready chunks (`Startable` false while a chunk is accepted). Readiness is `Board.LegacyReadiness` (the S1 legacy floor per class family, safe at the floor, risky within 2 levels): the S2 model (`readiness.go`, a `Word` per chunk) is merged, but the bot adapter does not build its survival inputs yet, so wiring it into `Env.Ready` is a follow-up (U).
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

## Readiness (slice S2)

> Built and tested offline (card quests2, TASK-586). Pure query: no bot code imports it, `bot.zoneModel` stays 0 and nothing here changes zone choice. Tags as above.

| piece | where |
|---|---|
| thresholds (data) | `data/questboard/readiness.json` (`quests.readiness/1`): dial lines `enter`/`stay` per cautious/balanced/bold, kit minimums (`minPotions`, `minResist` per difficulty, U), default dial |
| core | `internal/questboard/readiness.go`: `Thresholds.Assess(AssessInput) Readiness`, reason codes, the sentence |
| power adapter | `internal/questboard/readiness_power.go`: `HeroFrom(power.Hero)`, `EncounterFrom(name, kind, power.Result, k)` (k = survival ledger `KZoneFor * KHeroFor`, the same division the controller does) |

- Word rule (V, same as `survival.word` and `power.ZoneMargin.Word`; a test compares them): `safe` when the Low margin is at or over the dial's `enter`, `risky` when Best is at or over `stay`, else `too hard`. The word of a chunk is the worst of its encounters (the boss, and the zone on the way when the caller passes it). The dial lines mirror `internal/power/profiles.json` (V `TestDialsMirrorThePowerProfiles`).
- `Readiness{Word, Reasons, LegacyFloor, Unknown, FromLegacy, Sentence, Debug}`; the margin is only in `Debug` (tools), never in the sentence. The S1 string type `Readiness` is now `Word` (`ReadySafe`, `ReadyRisky`, `ReadyTooHard`, `ReadyUnknown`); `Input.Ready` maps chunk keys to `Word`.
- Reason codes (closed): `level_under`, `no_merc`, `kit_potions`, `kit_resist`, `margin_low`, `margin_wall`, `offence_unread`, `legacy_only`, `no_data`. Order: fixable first (merc, kit), then level, then the margin, then offence. The level number appears only in `level_under` and is the legacy floor (user asked for it beside the margin).
- Legacy floor fallback (I): where no encounter has a margin (boss missing from the tables, hero unread), the old clvl floor of the quest's family decides: at or over the floor `safe`, under it `risky`, never `too hard` (warn, never wall). `FromLegacy` is true so a later auto-start can refuse to trust that `safe`.
- Unknown: no margin and no floor gives `Unknown`, an empty word, reason `no_data`; S1 keeps the chunk `available`, never offered (V `TestUnknownReadinessKeepsQuestAvailable`).
- Tests (V, `internal/questboard/readiness_test.go`): Paladin and Sorceress on made-up zones sized to the logged outcomes (Cold Plains clvl 4 and 5 safe, Black Marsh clvl 4 too hard, clvl 12 with a merc safe, without a merc risky; docs/survival/README.md). The Sorceress rows are regression rows, not logged outcomes. A synthetic `survival.Ledger` with two merc deaths and a death in Black Marsh turns a safe word into a harder one. U: real zone and boss tables are not used (the install is on the worker host); no live run.
- Not done (for S3): choosing the encounters per chunk. Quest rows carry only a `boss` id, not an area; S3 resolves `boss` through `danger.Table.FindBoss`/`power.BossPack` and passes the zone of the route, plus the dial from the hero's setting.

## API and stores (slice S4)

> Built and tested offline (card quests4). Behind the knob `quests.board` (default 0): with 0 every action answers `off` and the GET shows `enabled:false` with the reason; nothing else changes. NOT verified live; the curl list below has not run. No new knob (the design's `quests.*` knobs arrive with S6 to S8). Tags: tests are V (they ran), the hero's real margin words and the live behaviour are U.

| piece | where |
|---|---|
| desk (the API face of `Exec`) | `internal/questboard/desk.go`: `Desk.View()`, `Start`, `HandIn`, `Cancel`, `Retry`, `SetMode`; every action returns a `Reply` |
| routes | `internal/watch/questsapi.go`: GET on the viewer mux, POSTs on the answer mux (same `policyGate` as repair, pause and farm gold: origin list, token, JSON only, small body, unknown field refused) |
| wiring | `cmd/d2bot/questwire.go` (`wireQuestBoard`, `bindQuests`, glance rows), called next to `wireJourney` in both backends; `bot.Runner.SetQuestBoardMaker`, `QuestBoard`, `QuestReadiness`, `SetQuestMargin` |
| readiness | `bot.Deps.QuestMargin` (survival model, read only, `survivalShadow.questMargin`) feeds `questboard.Thresholds.Assess`; `bot.zoneModel` is untouched |
| stores | `<state>/quests/<hero>.quests.json` + `.bak` (`internal/store.File`); mode stored in the same file (`Plan.Mode`, default `train`) |

Routes (V `internal/watch/questsapi_test.go`, goldens in `internal/watch/testdata/quests/`):

| method, path | body | reply |
|---|---|---|
| GET `/api/quests` | none | the view, `Cache-Control: no-store`, derived at most once per second (cached) |
| POST `/api/quests/<id>/start` | `{}` or `{"chunk":"<id>"}` | accepts the CURRENT chunk (the offered one, never the next part). 200 |
| POST `/api/quests/<id>/handin` | `{}` or `{"chunk":".."}` | 202 `pending:true`: the dialogue runs on the bot loop (can wait for a town trip or a recovery, bounded to 5 min); the outcome is `handIn` in the next GET |
| POST `/api/quests/<id>/cancel` | same | stops an accepted chunk (record removed, chunk ready again) |
| POST `/api/quests/<id>/retry` | same | clears the park of a blocked chunk ("Try again") |
| POST `/api/quests/mode` | `{"mode":"train|idle|farm"}` | stored per hero; slice S6 gives it behaviour (`modeActs:false` until then) |

Reply (every action): `{ok, already?, pending?, code?, reason?, quest, chunk, state, mode?}`. `ok:true` with `already:true` is a repeated press (idempotent). A refusal is `ok:false` with a `code` and a `reason` in words, never "could not do that" (the Pause bug of 2026-11-16: the page now reads state back from the GET and uses `ok`). Codes and HTTP status: `off` 409 (quests.board is 0), `no_hero` 503 (no snapshot yet), `paused` 409 (the player's pause is on), `busy` 409 (another chunk runs, named in the reason), `not_ready` 409 (not offered; the reason carries the evidence or readiness text), `not_awaiting` 409 (hand-in before the hero stands at the NPC), `done` 409, `unknown` 404, `bad_mode` 400; gate failures are 401 (token), 403 (origin), 405, 415, 400 (body), 503 (desk not bound).

GET fields: `schema watch.quests/1`, `enabled`, `reason` (when off), `difficulty`, `paused`, `mode`, `modeActs`, `autoStart{available:false}` (stub until S8; the per-quest `auto` flag is read from the plan store), `offers[]` (quest, chunk, title, act, risk, `readiness` word, `reasons[]`, `legacyFloor`, `sentence`, `startable`, `why`, `auto`; `pays` is omitted until the wallet exists, S10), `quests[]` (every quest of the difficulty with its chunks and states), `running`, `awaiting`, `blocked`, `handIn` (running, done, confirming, blocked or failed), `history[]` (newest 20 chunks the board ran), `needs[]`, `strip`, `extra`. Words only: no numeric margin and no chapter counter anywhere (V `TestDeskViewJSONShape`).

Glance (`/api/glance`): additive `quest{quest,title,chunk,part,state,step,npc}` for the video-card strip and Needs-you rows `quest_handin` ("Akara is waiting: hand in Den of Evil", button Hand in -> the handin route) and `quest_blocked` (Try again, Stop). The glance reads the same cached view.

Decisions and gaps (I unless tagged):
- The board is built per hero by the runner (the name is only known from the first snapshot): `SetQuestBoardMaker` opens `<hero>.quests.json`; a hero switch rebuilds it. The plan files carry the hero name, so `internal/charsel/perchar.json` does NOT list them (parking would hide the other hero's plan); V `TestQuestPlanFilesAreNotParkedBySwitch`.
- Readiness: `Assess` replaces the legacy word where the margin query has data (boss from the danger table, hero readable). Where `Assess` answers Unknown the legacy word stands (a quest without a floor stays safe) so no quest vanishes from the board because of missing data. The boss margin is divided by the ledger's hero correction only (a boss has no zone key; U). Queries are throttled to once per 30 s per level and difficulty. The potion count is `unread` -> no "short of potions" reason.
- "bot paused" is the player's pause (`PauseRunning` is the only running state), including the walk-home phase.
- A POST needs a body (`{}`); an empty body is a 400 like the other control routes.
- Auto-start flags, modes and meta awards are stubs: `autoStart.available:false`, `modeActs:false`, no `pays` (slices S6, S8, S10; the award keys are already idempotent per account in `Award.Key`, the ledger itself waits for the vault wallet).
- Not done: the strip's step names are the board's descriptive names, not the handler's (S3 note); restart persistence of `accepted` is V offline (`TestDeskModeAndRestartPersistence`), not live.

### Live check list for the API (NOT run; needs `quests.board` = 1 and a hidden instance, `tools/live.sh`)

Replace `$H` with the watch address, `$T` with the answer token (never paste it into a log); POST bodies are JSON.

1. `curl -s $H/api/quests | jq '{enabled,reason,mode,offers:[.offers[]|{quest,readiness,reasons,legacyFloor,startable}]}'` with the knob 0: `enabled:false`, reason names the knob. Set the knob to 1: offers appear; Den of Evil is `safe`; no number for a margin anywhere.
2. `curl -s -X POST -H "Authorization: Bearer $T" -H 'Content-Type: application/json' -d '{}' $H/api/quests/den_of_evil/start`: `ok:true`, `state:accepted`. Repeat: `ok:true, already:true`. Start Blood Raven meanwhile: 409 `busy`, the reason names the Den.
3. Poll `GET /api/quests` every 5 s: `running.step` moves, `strip.npc` is Akara, glance `quest` matches. Press Pause in the UI: `paused:true`; a Start then answers `paused`.
4. When `awaiting` appears (and the Needs row `quest_handin`): POST `handin`: 202 `pending:true`; poll until `handIn.state` is `done` or `confirming`; a second POST answers `ok:true, already:true`; the dialogue ran once (log `talk Akara/reward` once).
5. `POST /api/quests/mode {"mode":"farm"}` then restart the bot: GET shows `mode:farm`. Wrong token: 401; foreign `Origin`: 403.
6. Restart the bot while a chunk runs and while awaiting: `running`/`awaiting` come back from the plan file and the evidence, the chunk is not redone. Writes after start must be 0.

## Watch UI (slice S5)

> Built offline (card quests5), NOT verified live. Tags: node tests and the fit run are V (they ran); the look on a real hero and a real phone is U. Everything below is dormant while `quests.board` = 0: the quest tile hides itself, the Journey tile stays and the sheet says how to turn the board on.

| piece | where |
|---|---|
| `quest-strip` tile (strip slot, after `journey-strip`) | `internal/watch/web/panels/quest-strip`: one line of state in words, one primary button, "All quests", the mode switch (Train, Idle, Farm) |
| `quest-board` sheet | the same-named drawer panel (menu entry "Quests", or "All quests" on the tile); registered with `overlay.js` like the other drawers: one overlay at a time, Escape closes it |
| `quest-list` | the phone's Quests tab (Orders, Map, Quests, Progress), the same list without a drawer; its CSS is the sheet's with the id swapped (`TestQuestListCSSIsTheBoardCSS`) |
| shared logic | `quest-board/panel.js`: `stripModel` (state to words and button), `boardModel` (acts, parts, groups), `settle` (the payoff), `act` (the one place a button reaches the API), `captionOf` |
| video card line | `quest-strip` publishes `questCaption`, `video` shows it top-left of the picture (never over the HUD); it has no endpoint of its own (the v1 manifest stays) |
| Needs you | quest rows ride the existing glance; `sev: "event"` has an accent (not an alarm) and a refusal shows `reason`, not "could not do that" |
| Journey | glance gets `questBoard: true` while the board is on (`GlanceSources.QuestOn`, wired in `cmd/d2bot/questwire.go`); `journey-strip` then hides. The old tile, chip and panel stay until S9 |

States of the tile (words, from GET /api/quests only): offer ("Offer: Den of Evil, safe" [Start]), training (no button), running ("Running: Den of Evil", the step in words, [Stop] with a second press to confirm), waiting at the NPC ("Akara is waiting: hand in Den of Evil" [Hand in]), handing in ("Handing in Den of Evil..." as soon as the button is pressed, no button), confirming ("Checking that the game agrees"), failed (the API's reason, [Try again]), stuck ([Try again] [Stop]), idle (the mode line). A Start tap covers the current chunk only. Hand in: the page decides success from the GET (`handIn.state` done, or the chunk `done`), then shows the payoff toast once ("Done: Den of Evil", plus "+N meta" only when the API sends `pays`) and the tile shows the next offer. A refused POST shows the reply's `reason` (or its code when the reason is empty), never a generic line. No margin number and no chapter counter anywhere: the fit run fails a page that has one (`quest words`).

Stubs: the per-quest auto-start cell reads "auto: coming" while `autoStart.available` is false (S8; with `available` true it is a switch posting `{auto}` to `/api/quests/<id>/auto`, route U); the mode switch is stored but acts in S6 (tooltip says so); `pays` is absent until S10.

Gaps (I): the API has no step list, so the tile names the current step only (no "step 2 of 4" dots); no "levels to go" count (the first readiness reason is shown); the offer Needs-you row ("Blood Raven is ready") is not raised (the tile carries the offer); the quest-board sheet is the 440 px drawer, not a rail-width sheet.

Checks: `go test ./internal/watch/ -run Quest` (node: `testdata/questui/quest_ui.mjs`, fake API fed by the S4 goldens, both builds); `go run ./tools/uilint`; fit pages `quest-*` and `quest-sheet-*` in `tools/uishots.sh --fit` (all eight sizes plus 412x780, 360x640 for sheets), must-fail `questdefect-clip|small|words|two` (`tools/uishots/fixtures_quests.go`: the real Desk over the Fake world, variants by the `quests=` query), pictures with `tools/uishots.sh --pics --only 'quest-*'`.
