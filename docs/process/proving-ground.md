# The Proving Ground: scenarios that play the real bot and read the game

> status: current (card provingground, 2026-11-19). Source: the design review `docs/design/fable-review-2026-11-19/` (05 section 5, 06 R1 and R2) and the owner's decision of 2026-11-19 to build it first. Tags: V verified (command or test named), S secondary, U unverified. Nothing here changes bot behaviour or the merge gate.

## 1. What it is

A scenario is a small, deterministic lab-slot run in which the REAL bot code plays a declared situation and the verdict is read from GAME STATE afterwards (quest records, item codes, object states, area, gold, life), never from log lines. It exists so that a behaviour is "done" when its scenario is green, and so that a stuck hero (the Search for Cain tree, 17 hours in the sandbox) is a red line in minutes.

Parts (all stdlib Go, one data directory):

| part | where | what |
|---|---|---|
| format and logic | `internal/scenario` | scenario, assertion and result types, the SETUP builder, tier selection and matrix, coverage, determinism check (V `go test -race -count=20`) |
| tool | `cmd/d2scenario` | `list validate plan setup eval coverage provenance report determinism gate-check run` |
| data | `tools/provingground/scenarios/*.json` | scenarios (`didle.scenario/1`), `behaviours.json` (the behaviour catalogue) |
| builder wrapper | `tools/provingground/run.sh` | builds d2bot and d2scenario, memory guard, runs in lab slots |

## 2. Writing a scenario

One JSON object in `tools/provingground/scenarios/*.json` (unknown fields are refused; ids are lower case with `.`, `-`, `_`).

- `kind`: `step` (one quest chunk alone; declares `quest` {quest, chunk} of `data/questboard` and `step` {action, target, nodes}), `behaviour` (anything outside quests: chores, belt, recovery, loot, flail) or `e2e`.
- `pre`: the minimal declared precondition: `level`, `gold`, `stash`, `difficulty`, `waypoints`, `quests` (record, word), `board` (the chunk the player accepted), `items`, `belt`, `merc`, `start` (`town` default, or `field` with `area`). The same declaration is what the arbiter's shadow mode and the questspec checker read as the step's pre facts; `assert` is its post facts.
- `trigger`: the bot's `--goals` and tuning knobs; `driver: board-taps` (with `tap_chunks`, `tap_quest`) runs the bot as the sandbox does (`--ladder --watch`, the quest board is built) and a tapper taps Start on the offered chunk and Hand in when the board waits, nothing else (V the harness of the unattended Search for Cain run, `docs/probes/cainl3/poll.py`; the tapper is `cmd/d2scenario/tap.go`, its taps are in the result and `<run>.taps.log`). A driver scenario does not use `pre.board`. `budget.virtual_min`: the timeout, in VIRTUAL minutes (a scenario without one is refused). `matrix`: classes and seeds.
- `assert`: kinds `item_in_pack`, `item_absent`, `quest_bits` (record, set mask, clear mask), `area_is`, `area_in`, `gold_between`, `stash_between`, `level_at_least`, `life_pct_at_least`, `object_state`, `area_changes_le`, `portal_open` (a portal stands in that level, `debug portals`), `no_writes`. Each cites the fact it reads. A fact the observer could not read makes the assertion UNOBSERVED.
- `covers`: behaviour ids of `behaviours.json`; a step scenario also covers its `quest.<quest>.<chunk>`. `touches`: path prefixes; the merge gate runs the fast scenarios whose prefixes match the changed files.
- `needs`, `build` (`lab` or `sandbox`), `provenance` (what is CHEATED), `status: pending` with `pending_reason` for a skeleton.

Dev loop: `d2scenario setup cain.scroll -class Sorceress -seed 3` shows exactly what will be written; `d2scenario eval cain.scroll -state s.json` evaluates a state file offline; `d2scenario plan -id cain.scroll` is the dry run.

## 3. The setup builder (no saves checked in)

`scenario.BuildSetup(scenario, class, seed)` is a pure function from the declared `pre` to (a) lab verbs, each with a ledger reason naming scenario, class and seed (`debug stat 12/13/14/15`, `statpts`, `skillpts`, `wp`, `quest set`, `give <code> pack [q<n>] [unid]`, `give <code> <n> [unid]` (an item on the ground beside the hero), a life percentage by `debug hit` (the live life is read first), `belt clear`, `give <code> belt`, `merc hire`, `goto` for a field start), (b) the accepted board chunk, written with the real `questboard` plan store into `<instance>/state/quests/<hero>.quests.json` (the same code the bot reads), (c) the knob overrides as `state/tuning.json`, (d) provenance tags: `CHEATED` when the state could not be reached by play (a warped position, a given quest item, a hand-set quest word or waypoint, a free merc, a changed difficulty), `SETUP-only` otherwise. A precondition with no known verb (an item into the stash, a life percentage) is listed as `Unsupported` and the run is BLOCKED, never faked. Verb argument forms that are not verified live are marked U in the code (`debug difficulty`).

A hero is created fresh by the normal path (`--d2d-class`), the verbs are applied on a lab d2d (`--lab-verbs`, lockstep), the hero is saved and that d2d exits.

## 4. How a run works (V design from tools/survival/abrun.sh, U until the first green run)

1. `tools/d2dslot.sh run` takes a lab slot (private socket, ports, home).
2. SETUP on the lab d2d, ledgered. Town start: save and quit. Field start: the same d2d (verbs on) is handed to the bot with `--attach --d2d-attach`; the position is a declared CHEATED SETUP and `no_writes` is UNOBSERVED because that d2d keeps verbs on.
3. The bot (`d2bot --backend d2d --managed`, lockstep, `D2DSOAK_HOURS` = the virtual budget) plays on a d2d WITHOUT lab verbs: gated debug verbs are refused, so no gameplay write can happen after the start; the refusals are counted into `writes after start: N`. `d2dbackend` stops the run on the virtual clock.
4. OBSERVE: a read-only d2d (no lab verbs) loads the saved hero and the observer reads state: `state` (level, gold, life), `debug level` (area), `dump player` (stash gold), `items` (code, place), `debug quest list` (records), `debug objects` (class, state). No log line is read. If the bot run left no newer save the run is BLOCKED.
5. EVALUATE: assertions on that state give GREEN, RED, BLOCKED or PENDING. A result JSON per run, `results.json`, a one-screen report, the bot log as `Trace` (the human's view of what it chose and why, never asserted).

Limits to know (U): a reloaded hero stands in town, so `area_is` and `object_state` need a field start (live observation of the same d2d); the bot may not save on exit (then BLOCKED, the first run shows); the observer's parse of `debug objects` columns, `items` `panel=` for the stash and `debug quest list` follow the L0/L2 probe docs but were not run through this tool yet.

## 5. Statuses, determinism, green

- GREEN: every assertion passed on observed state, `writes after start: 0` where asserted. RED: an assertion failed on observed state. BLOCKED: the run could not construct or observe what it needs (a missing build such as the sandbox d2d, a setup with no verb, an unobserved fact, an infrastructure failure): never a pass and it fails the gate. PENDING: a skeleton, listed and counted, never run, does not fail the gate.
- A CHEATED setup is mechanics evidence only (provenance rule, `docs/probes/test-characters.md`): a green scenario says the bot does X from that declared state, not that a played hero reaches it. `d2scenario provenance` prints the rows for that file.
- Determinism: the same scenario, class and seed twice must give the same verdict (hard) and the same state digest (a differing digest with the same verdict is a warning while the bot loop still has wall-clock decisions, `docs/d2d-determinism.md`). `d2scenario determinism a.json b.json` exits 3 on a verdict difference. A flaky scenario is a failed scenario: fix it or demote it; do not retry until green.
- Isolation: one hero, one instance directory, one slot per run; no shared state; at most 4 lab slots in parallel (the builder memory guard, `run.sh` refuses below 2500 MB available).

## 6. Running (builder only; BUILDER is the builder host name of your local.env)

The builder is one machine; scenario runs go on the LONG lane, one job at a time, never the worker host, nothing heavy on the head box. Merge gates have priority on the main lane.

```
# dry run, no game (the head box is fine):
go run ./cmd/d2scenario plan -tier fast
go run ./cmd/d2scenario run -dry-run -tier full
# the one-scenario dev loop:
labrun BUILDER --lane long --detach --pull provingground-out -- tools/provingground/run.sh -id cain.scroll -class Paladin -seed 11 -parallel 1
labrun watch <job>        # then: go run ./cmd/d2scenario report provingground-out/<stamp>/results.json
# fast tier and a determinism pair:
labrun BUILDER --lane long --detach --pull provingground-out -- tools/provingground/run.sh -tier fast -parallel 2
```

Output: `provingground-out/<stamp>/results.json`, one `<id>_<class>_<seed>.json` per run (assertions, observed state, digest, tags) and `.bot.log` (the trace).

## 7. Tiers and the gate proposal (nothing applied)

| tier | content | when |
|---|---|---|
| fast | scenarios with `tier: fast`, one class and one seed each, 30 to 60 of them at the target size, minutes each | proposal: a merge step for changes under `internal/bot`, `task`, `questboard`, `nav`, `town`, `combat`, `cmd/d2bot`, selected by `-changed` against each scenario's `touches` |
| full | every scenario times its class and seed matrix | nightly on the long lane |
| nightly | act runs (fresh hero to a boss), later | nightly on the long lane |

Proposal for the merge gate (to be decided by the owner and the merge owner; this card applies none of it):
1. After the whole-repo gate passes on the main lane, the merge queue computes the changed files and calls `labrun BUILDER --lane long -- tools/provingground/run.sh -tier fast -changed <files>`; `d2scenario gate-check results.json` decides (exit 3 on RED or BLOCKED).
2. Start in report-only mode (the verdict is posted, the merge is not blocked) until the fast set has been green on `main` for a week and the determinism pair is clean; then make it blocking. A behaviour card names its scenario in the task; no green, no merge; a feature that cannot have one yet merges behind a knob the release does not turn on.
3. Dry-run mode already exists (`-dry-run`): the queue can print what it would run on every merge before the step is wired.
4. Cost control: the long lane has one job at a time and the main lane has priority, so a fast gate is queued behind at most one other long job; the set is chosen by touched files and runs at most 4 slots.

## 8. The behaviour catalogue and coverage

`behaviours.json` lists what the bot does (the runGoal hooks, chores, resource policy, recovery, loot, travel, merc, difficulty) with the code entry of each; every quest-board chunk is derived from `data/questboard` (`quest.<quest>.<chunk>`). `d2scenario coverage` maps each behaviour to the scenarios that name it and lists COVERED, PENDING-ONLY and UNCOVERED. `d2scenario validate` fails on a scenario naming an unknown behaviour or chunk. Add a behaviour when a hook is added; a hook with no scenario shows up as UNCOVERED, which is the point.

## 9. What exists today (2026-11-19)

Catalogue: see `d2scenario coverage`. Scenarios: Search for Cain steps `cain.scroll` (fast), `cain.scroll.sandbox` (full; BLOCKED without `D2D_BIN_SANDBOX`), `cain.scroll.field` (full), `cain.read` (fast), `cain.report` (full), `cain.stones` and `cain.free_cain` (PENDING, reasons in the file); behaviours `town.stash-gold-withdraw` (fast) and `flail.area-cycle` (PENDING: the area-change count is not an observable fact yet) plus six PENDING skeletons (belt refill, chicken and heal, rare ring, repair, merc hire, Nightmare start). 

First live results (V, builder long lane, lab d2d build of 2026-11-17, bot of this branch, Paladin clvl 17 seed 11, `tools/provingground/run.sh`; output `provingground-out/`, not committed):

| scenario | verdict | what the state and the trace show |
|---|---|---|
| `cain.scroll.field` (hero warped into the Dark Wood, board chunk accepted) | RED, 842 s wall | the bot reaches the tree step (`quest board: step Take the Inifuss bark`), then `still exploring for object class 30 (0 s of 720 s)` and cycles Dark Wood, Cold Plains and the camp (zone entries 5,1,5,6,5,1,5,6 ...); no `bks` in the pack at the end. This is the sandbox stall reproduced on the lab slot. |
| `cain.scroll` (from the camp, travel included) | RED, 349 s wall | the chunk starts (`step Activate waypoint`), the step is cancelled by a fight (`context canceled`), the chunk stays `queued: waiting for the fight to end` and the hero trains; no `bks`. Not the tree stall: the chunk never gets back to the tree. U whether that is a scenario artifact of the `reach:30` goal. |
| first attempt | BLOCKED by design after the fix | the instance id was too long (bot refused it) and the run was wrongly judged RED; now a bot that exits with an error is BLOCKED, never RED. |

Not run live yet: `cain.read`, `cain.report`, `town.stash-gold-withdraw`, `cain.scroll.sandbox` (needs the sandbox d2d build), and the same-seed determinism pair.

## 10. Dead ends and open points

- FINDING (V, first live run 2026-11-19): with `--goals campaign` a headless bot never builds the quest board runtime (`boardStep` reads `qbRT()`, which only the watch API or `boardPreempts` create), so an accepted chunk is ignored and the hero trains (the log shows `level plan` and no board line). A scenario that needs the board therefore uses a level or farm goal (`reach:30`), where `boardPreempts` builds the board and runs the accepted chunk, as in the sandbox where the player taps Start on a hero with a farm goal.

- A save cannot hold a field position (the game re-enters in town), so isolating a step that starts in a field needs the live hand-over (field start) or the bot walking there from town (the budget then includes travel).
- The board needs the player's Start tap; the scenario writes the accepted record in the plan store, as the tap does. A chunk that ends on a hand-in (`read`, `report`) stops at the NPC and waits for the player's Hand in; those scenarios may need a hand-in tap in the trigger (U).
- Anti-flail needs a level-change counter the observer does not have (fork read verb or sampling); until then it is PENDING.
- `quest_bits` after the tree click is not asserted: which record bit the tree sets is not known (U). The scroll in the pack is the fact.

## 11. Observables added by the expansion card (pgexpand, 2026-11-19)

| fact | how it is read | tag |
|---|---|---|
| area changes (`area_changes_le`) | the number of `zone entry` lines of the bot's log after the first (`state.area_changes`, `area_trail`); one line per area change (`internal/bot/lifelog.go`). The one exception to "never a log line": the observer cannot hold the control socket during the run (the d2d takes one client), and the d2d has no level-change counter. | S |
| portal open (`portal_open`) | `debug portals` rows (which, level), live observation only; the way to Tristram is a row for level 4 (V L1), kept from Tristram too | V lab, U shipped |
| item on the ground | `ground`, only on a live (field) observation; a reloaded hero's floor is the town and is not read | V verb |
| stone lit state | `debug objects` NU to ON (V L1), readable only while the hero stands in Stony Field: the bot's own check, not asserted; the scenario asserts the portal instead | V lab |
| shipped verbs | a field start needs the lab d2d (verbs on); the shipped build refuses the warp, so `cain.stones.shipped` is BLOCKED by `blockedBefore` with that reason; whether the shipped build lists portals is U | decision |

## 12. State after the expansion card (2026-11-19)

19 scenarios: 14 runnable, `cain.stones.shipped` BLOCKED by design, 3 PENDING (repair, merc, nightmare). Fast tier first run: GREEN 4, RED 3 (details and findings: `docs/probes/provingground-run-2026-11-19.md`). Coverage: 25 of 94 behaviours covered after merging main (22 of 91 on this branch alone; was 10), 6 pending-only, 63 uncovered; all five Search for Cain chunks have a scenario. Determinism pair for `town.resupply` Paladin seed 11 submitted, result in the run record.
