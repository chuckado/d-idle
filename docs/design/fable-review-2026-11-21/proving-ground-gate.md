# Proving Ground, merge gate and builder scheduling: design review (2026-11-21)

> status: proposal for the owner (card fablepg, TASK-712). Docs only: nothing here changes Go, the merge scripts, the builder or a scenario. Companion: [proving-ground-gate-catalog.md](proving-ground-gate-catalog.md) (scenario catalogue and coverage map). Builds on `docs/design/fable-review-2026-11-19/05-architecture-and-overhaul-options.md` sections 5, 8 and 9 and on ADR 0020; it does not amend either.

Tags: V = verified in this review (the command or file:line is named), S = secondary (a doc or a brief says so), U = unverified or an estimate. No game and no builder job was started for this review; every number comes from files already on disk.

## 0. Verdict in ten lines

1. The digest differences are almost certainly ours, not the engine's. The harness drives the board with a WALL-CLOCK tapper against a game that runs on a virtual clock, the bot still has known wall-clock inputs, and the digest hashes every observed number, asserted or not. (V code read; U until the journal pair in 2.5 is run.)
2. The engine was shown deterministic under identical commands for 12000 ticks on 2026-10-30 (V `docs/d2d-determinism.md` section 6). Nobody has repeated that pair since the board, the watch API, points and the vault were added. The tool for it exists (`cmd/d2ddet`) and the Proving Ground does not use it.
3. Split "same verdict" from "same world": gate on an ASSERTION DIGEST (only the facts a scenario asserts), keep the full state digest as a drift report, and use the journal to locate a divergence.
4. Five tiers, T0 to T4. Only T0 and T1 (no game, seconds) block every merge now. T2 (a smoke set of town-start scenarios, about 10 wall minutes) runs after the push, report-only, and becomes blocking when it has earned it.
5. A scenario has a stability state (candidate, stable, quarantined). Only stable scenarios gate. A verdict flip on unchanged binaries quarantines the scenario automatically; there is no retry-until-green.
6. The merge gate costs a median of 435 s and a p90 of 1990 s per code merge (V, 78 merges since 2026-11-17). The median is compute, the tail is queueing behind agent jobs. Cut the median with the Go test cache and prebuilt check tools; cut the tail with one locked job, merge priority and agent quotas.
7. Whole-repo and race safety stay: whole-repo vet and test still run for every code merge (cached), CI still runs it after the push, and a nightly uncached whole-repo `-race` run is added.
8. The merge worker needs a supervisor, a heartbeat and a stuck-item alarm: it timed out at 3600 s five times and sat idle for 471 minutes with items queued (V worker log and queue log).
9. One results ledger (JSONL, one record per run, written by the runner) and one board page replace verdict tables in probe docs. A doc cites a run id.
10. Scenario setup moves into a hero library with a self-check, so a hero with banked points, no weapon or three potions is BLOCKED(setup) in seconds instead of RED after 30 minutes.

## 1. Evidence base

| what | where | tag |
|---|---|---|
| runner, digest, tapper | `cmd/d2scenario/run.go`, `tap.go`, `internal/scenario/determinism.go`, `state.go` | V read at 5b12ce5e7 |
| 38 scenarios, 97 behaviours (32 covered, 7 pending-only, 58 uncovered) | `go run ./cmd/d2scenario list`, `coverage` | V run on the head box |
| run records | `docs/probes/provingground-*.md` (8 files) | S (the authors' V) |
| loop determinism study | `docs/d2d-determinism.md` | S (V there, 2026-10-30) |
| merge queue history | the merge worker's `queue.jsonl` (1020 items) and `worker.log`, ops side | V parsed with a script, items 900 to 1020 |
| merge gate code | ops `tools/merge.sh` lines 475 to 510, `tools/mergeq.py` | V read |
| builder state at 22:14 EST on 2026-11-20 | `labrun builder-status` (light lane, read-only) | V |
| waits of 30 to 60 min, load 40 to 180, silent worker stop | the brief | S |

## 2. Why two runs of one seed differ

### 2.1 The three places a difference can come from

| layer | what it would look like in a journal pair (`d2ddet compare`) |
|---|---|
| engine (d2d) | identical commands at identical ticks, then a different reply or checkpoint hash |
| bot | identical state hashes so far, then a different COMMAND |
| harness | an input that is not in the bot at all arrives at a different tick (a tap, a cut, a different binary) |

### 2.2 Harness sources (all V by code read; none measured as THE cause yet)

| # | source | evidence | scenarios hit |
|---|---|---|---|
| H1 | The board tapper runs on the wall clock: it waits 20 wall seconds, then polls every 5 wall seconds. The game runs 5 to 60 times faster than real time (S `internal/scenario/scenario.go:268` comment), so one poll interval is 25 virtual seconds to 5 virtual minutes, and the ratio moves with builder load. The Start tap therefore lands at a different game tick in every run. | `cmd/d2scenario/run.go:517`, `:519`, `tap.go:128` | all 13 `board-taps` scenarios: the Blood Raven pair, `cain.chain`, Den, Countess, Andariel |
| H2 | `trigger.api` calls fire at wall seconds after the start. | `tap.go:157` | `board.stop-parked-chunk` and the pending `points.api-*` |
| H3 | The digest hashes the WHOLE observed state: purse, stash, life percent, every carried item, the area trail from the bot log, stats. A scenario asserting one quest bit gets a new digest when the hero ends with 712 gold instead of 700. | `internal/scenario/determinism.go:13-33`; `town.resupply`: gold 700 against 712, same verdict (S run record 2026-11-19) | all |
| H4 | The area trail and area-change count come from the bot's log and are inside the digest. | `run.go` `addTrail`, `state.go` `AreaTrail` | all |
| H5 | The run is cut by `--run-for` wall minutes (default virtual/5 x 1.6 + 10). The result records the full virtual budget without checking that the bot logged `virtual run length reached`. On a loaded builder a run can be cut early and still be judged. | `run.go:474`, `run.go:307`, `scenario.go:267-268` | long scenarios under load |
| H6 | The result does not record the bot commit, the d2d binary hash, the host, the load or the pace ratio. One "pair" in the records ran two different bot builds (S pgfix4 section 4: run A before a wiring fix). | `internal/scenario/result.go` | every pair so far |
| H7 | Results are lost on the way back: two of four `results.json` of the `cain.chain` widening did not survive the pull (S pgfix4). The digest pair could not be compared at all. | probe doc | evidence, not behaviour |
| H8 | The observer d2d is started without a seed and vendor shelves were once counted as floor (S pgfix4 2.2). | `run.go:312`, `internal/d2dinput/lifecycle.go:89` | `item_absent ... ground` |
| H9 | `log_line` windows map wall time linearly onto the virtual budget. | `docs/process/proving-ground.md` section 2 (marked U there) | 9 scenarios with `log_line` |
| H10 | The second worker's lab d2d is a different build file from the builder's; same commit or not is U. | `docs/process/second-worker.md:359` | every cross-host comparison |

### 2.3 Bot sources (wall clock inside the decision loop)

| # | source | evidence | status |
|---|---|---|---|
| B1 | `goals.Manager` ticks on `time.Now` (no `Clock` is passed on the d2d path). | `cmd/d2bot/d2dbackend.go:357`; open since 2026-10-30 (`docs/d2d-determinism.md` section 2 row 3) | open |
| B2 | Farm time budget uses `time.Now` and `time.Since`. | `internal/bot/farm.go:143`, `:149` | open; hits the `farm:` scenarios (`chicken.recovery`, `loot.rare-ring-classified`) |
| B3 | The soak stop is not tick-based: one run sends one more read where the other already steps. | `docs/d2d-determinism.md` section 6 | open |
| B4 | Watch and ladder code take a `decide.Clock`; which clock the d2d path wires is U. With `--ladder --watch` (every board scenario) HTTP handlers also inject state from another goroutine. | `cmd/d2bot/ladder.go:253-285` | U, not audited |
| B5 | Runner heartbeat and stall watchdog read the wall clock. On a builder at load 40+ a slow wall minute can look like a stall. | `internal/bot/runner.go:756`, `loop.go:28`, `:45` | U whether a decision follows |
| B6 | Map iteration order in decision code was never audited. | `docs/d2d-determinism.md` section 2, last paragraph | U |
| B7 | A pickup is checked 8 ms later and not seen (a read race). | S `provingground-cain-2026-11-19.md` finding 2 | open, harmless so far |

Not sources (S, same doc): nav and combat dice are seeded with a constant; nav, combat, town and the runner timers use the env clock; the guard runs on the stepping goroutine since the `Pacer.Watch` fix.

### 2.4 Engine sources

| # | source | status |
|---|---|---|
| E1 | Neighbour levels were built on a wall-time thread pool and installed when ready. | fixed under lockstep by the fork's link-wait change (S determinism doc section 6). U: that the current lab binary (built after the fork rebase of 2026-11-15) still carries it. One `strings` or a fork `git log` check settles it. |
| E2 | `frame` and `cam=` in replies are wall-driven. | not bot inputs; masked in the journal (S) |
| E3 | One more start-up field of the `state` reply differs between runs and is unidentified. | open (S section 6); `D2DCTL_JOURNAL_RAW` exists to find it |
| E4 | Seed plumbing. The seed reaches the prep d2d (`run.go:258`) and the bot's d2d (`run.go:471`, `lifecycle.go:89-90`). In a field start the bot attaches to the prep d2d, so one seed applies. | V read. U: whether a LOADED save re-rolls drops and layouts from `--seed` alone or also from the save's own state. |

Reading the evidence: the one pair with identical full digests is `cain.stones` Paladin 11 (S), a FIELD start where the chunk was accepted at setup and the tapper did nothing useful. The pairs that differ are board-taps or town-start runs with shopping. That pattern fits H1 and H3 and does not need an engine cause. It is still a pattern, not a proof.

### 2.5 The cheap proof (one card, about 40 builder minutes)

1. Take two scenarios: `town.resupply` (no taps) and `act1.burial.blood_raven` (taps), Paladin 11.
2. Run each three times on one host with `D2DCTL_JOURNAL=<file>` set for the bot (the journal is already in `internal/d2dctl`; the runner only has to pass the variable).
3. `d2ddet compare -ignore state A.jnl B.jnl` names the first difference and its class: a COMMAND on equal hashes is the bot or the harness, a REPLY or checkpoint under equal commands is the engine.
4. Expected (U): `town.resupply` diverges late or not at all; Blood Raven diverges at the Start tap. If instead a checkpoint differs under equal commands, section 2.4 becomes the first priority and the fork owner gets the journal.

### 2.6 Three digests, three jobs

| digest | content | rule |
|---|---|---|
| D0 verdict | GREEN, RED, BLOCKED | hard, as today |
| D1 assertion digest (new) | scenario id, class, seed and, for each assertion in order, the fact it read: discrete facts by value (quest word, item code present, area, skill levels, hand skill, portal level), range facts by their pass or fail bucket only (gold, stash, wealth, life percent, area changes, log line counts) | must be equal for a scenario to be `stable` |
| D2 state digest (today's) | the whole observed state | reported as drift, never gates; the goal is to shrink drift as B1 to B3 and H1 are fixed |
| D3 journal | command stream and checkpoint hashes | diagnosis only |

D1 is deliberately close to the verdict. Its extra value is the discrete facts: two GREEN runs where one ended with quest word 0x2011 and the other with 0x2019 are not the same run.

### 2.7 N-run policy and quarantine

- A new scenario is `candidate`. It becomes `stable` after 5 consecutive runs of one class and seed with equal D0 and D1, on unchanged binaries, at least two of them while another slot is busy (parallel 2). Until then it never gates anything.
- Nightly: every stable scenario runs once; three rotate through a same-night pair. All results go to the ledger (section 5).
- A flip of D0 or D1 on unchanged bot commit and d2d hash is a FLAKE event. The scenario goes to `quarantined` at once, by the tool, not by a person. It keeps running nightly, is shown amber, gates nothing, and a card for its owner is proposed. Out of quarantine: a root cause in the ledger note and 5 clean runs again.
- A blocking gate run that fails may be re-run once. If the second run passes, the merge goes on AND the scenario is quarantined: the first result was a flake by definition.
- Honest limit: 5 clean runs still pass a scenario with a 10 percent flake rate 59 percent of the time. The nightly history is what catches it; the page shows the last 30 runs per scenario.

### 2.8 Fix order

1. H6 and H5 first (record provenance, BLOCK a cut run): half a day, and every later pair becomes trustworthy.
2. D1 (2.6) and the journal pair (2.5).
3. H1 and H2: taps on the virtual clock. Recommended form: a scripted answer source inside the bot (`--scripted-taps <file>`, rules such as "tap Start on chunk X at the first tick the offer is startable", "tap Hand in 2 virtual seconds after the board waits"), evaluated on the stepping goroutine like the guard. The HTTP tapper stays for sandbox-style e2e runs, which are then never `stable`.
4. B1 and B2: two small changes, known since 2026-10-30. B3 with them.
5. B4 to B6 only if the journal pair still shows a bot-side divergence.

## 3. Tiers

### 3.1 The model

| tier | what runs | game | wall budget | where |
|---|---|---|---|---|
| T0 static | `d2scenario validate`, `coverage`, `plan -dry-run`; questspec checker; build-file and plan-table validation; the static half of the setup self-check (6.2) | no | under 30 s in total | head box, CI |
| T1 offline | `d2scenario eval` over recorded state fixtures (the state file of the last green run of each stable scenario: assertion and observer-parser regressions); decision-trace goldens for the incident cases (ping-pong, broke with stash gold, ignored Start tap); `mem.Fake` package tests | no | seconds each, inside `go test` | anywhere |
| T2 smoke | stable, town-start, no taps, at most 3 wall minutes each, one class and seed: today `heal.town`, `town.resupply`, `belt.refill-after-fight`, `town.stash-gold-withdraw`, `points.bank-by-default`, `points.autopilot-spends-by-verb`, `cain.report`, `town.tp-scroll-hoard` | lab slot | 10 minutes at 2 slots (measured singles: 8 s, 94 s, 149 s, 140 s, 31 s; S run records) | builder long lane or the second worker |
| T3 full | every runnable step and behaviour scenario times its class matrix; one seed per class at night, the second seed weekly | lab slot | 5 to 35 minutes each; the full set about 5 to 6 hours of slot time (U, from the per-scenario wall times in the records) | long lane, at night |
| T4 e2e and soak | `cain.chain`, act runs from a fresh hero (EARNED), survival A/B waves | lab slot | 15 to 150 minutes each | long lane at night, second worker when allowed |

T1 is honest about what it is: it guards the harness and the arbitration, not the engine. Its fixtures come from real runs and carry the run id they were cut from.

### 3.2 Which tier gates which event

| event | blocks | reports only |
|---|---|---|
| agent branch, before "done" | T0; T1; the card's OWN named scenario, run once by the agent (a run id in the card result) | |
| merge | T0 and T1 (they are part of the Go gate) | T2 after the push (stage 1); T2 blocking for merges that touch the `touches` prefixes of a smoke scenario (stage 2, criteria in 4.6) |
| nightly | nothing (it reports) | T2, T3, the determinism pairs, one T4 item; a RED stable scenario opens a hold-or-fix line on the board for the coordinator |
| release cut | T0 to T3 green on the release commit; the smoke set clean five times; at least one EARNED act run green; no quarantined scenario among those the release notes cite | T4 soaks |

The current `tier` values map as fast to T2 or T3 (see the catalogue: five of today's twelve `fast` scenarios are board-taps or RED and do not belong in a gate), full to T3, nightly to T4.

### 3.3 Setup honesty tags

Four tags, one per run, stored in the ledger:

| tag | meaning | may be cited for |
|---|---|---|
| EARNED | the hero was created by the normal path and every later state came from the bot playing (including a save exported by an earlier EARNED run) | behaviour, balance, release evidence |
| SETUP | set by verbs to a state a played hero reaches (today's `SETUP-only`) | behaviour of the step; not balance |
| CHEATED | a state play does not reach this way (warp, given quest item, quest word, waypoints, level jump) | mechanics only |
| IMPOSSIBLE | stats or items no played hero can have | mechanics only; never on the release page |

Today 0 scenarios are EARNED, 12 are SETUP and 26 are CHEATED or IMPOSSIBLE (V `d2scenario provenance`; the catalogue lists each). The original design asked for prepared saves made by the normal path (`05-architecture` section 5, row L1); the build went to verbs because a save cannot hold a field position. Section 6.3 proposes the way back.

## 4. Merge gate and builder scheduling

### 4.1 What the queue log says (V, items 900 to 1020, 2026-11-17 to 2026-11-20)

| measure | value |
|---|---|
| code merges with timings | 78 |
| gates phase per code merge | min 75 s, p25 280 s, median 435 s, p75 593 s, p90 1990 s; 12 of 78 over 20 minutes; 13.8 hours in total |
| preflight phase | median 59 s, p90 68 s (three `go run` invocations) |
| docs-only merge | median 145 s, p90 319 s |
| wait in the queue before the merge starts | median 525 s, p90 6461 s, max 48993 s |
| `gate go-whole failed` | 8 since item 700 on 7 branches; 6 of those branches merged on a later queue entry, two of them on the very next entry (988 to 989, 1007 to 1008) |
| `merge.sh timed out after 3600s` | 5 (items 895, 906, 948, 953, 962) |
| worker restarts in the log | 4 `worker started` lines; idle stretches over 60 minutes with items queued: 9, the longest 471 minutes (2026-11-20, 01:33 EST) |

Two facts explain most of it. First, `merge.sh` calls the builder twice per code merge (whole-repo at line 480, race at line 499) and each call queues again on the main lane, so an agent job can slip between them. `labrun BUILDER --gate` already does both in one locked job (V `labrun --help`) and the merge does not use it. Second, at the time of reading the main lane held the merge gate and one agent's own `go vet && go test` behind it, and the long lane had one job running and three waiting (V builder-status). Agents use the same lane as the gate for work the gate will repeat.

The failing output of a go-whole gate is not kept: `result_text` is empty and `log_tail` holds only the JSON line (V items 927, 988, 1007, 1010). "Flaky or contention" cannot be answered from the records. That is the first thing to fix.

### 4.2 Per merge, nightly, release

| check | today | proposed |
|---|---|---|
| leakcheck, hygiene, contracts, backlog | every merge | same, from prebuilt binaries (preflight 59 s to an estimated 15 s, U) |
| `go vet ./...` and `go test -short ./...` (208 packages) | every code merge, `-count=1` | every code merge, WITHOUT `-count=1`, with a persistent build and test cache on the builder: packages whose inputs did not change answer from the cache |
| `-race` on touched packages | every code merge, second builder job | same check, in the same locked job |
| whole-repo uncached, whole-repo `-race` | CI after the push (uncached); no whole-repo race | CI unchanged; nightly uncached whole-repo `-race` on the second build host; red result uses the existing red hold |
| T0 and T1 scenarios | not separate | inside the Go gate |
| T2 smoke | not run | after the push, report-only, then blocking (4.6) |
| T3, T4, determinism pairs | by hand, by agents | nightly, by one scheduler job |

Why the cache keeps the safety: Go's test cache keys a result on the package's compiled inputs, the files and environment variables the test read, and the flags; a change anywhere in a package's dependency closure re-runs it. So "whole repo" is still evaluated for every merge. What is lost is the re-roll of unrelated flaky tests on every merge, which is exactly the cost we want to drop; the nightly uncached run takes that job. U: whether the builder mirror keeps file times stable enough for a high hit rate (measure before relying on it), and the hit rate itself.

One side effect: `merge.sh` treats a whole-repo gate faster than 20 s as "did not run" (`MERGE_GATE_MIN_SECS`, lines 485 to 487). A cached gate can be that fast. Replace the timing guess by a positive marker: the count of `ok` and `no test files` lines must equal the package count.

### 4.3 Cost estimate

| | today (V) | proposed (U, estimate) |
|---|---|---|
| median code merge, gates phase | 435 s | 120 to 200 s (cache; internal/bot and its dependents still run when touched) |
| p90 code merge, gates phase | 1990 s | under 600 s (one job, merge priority; no queueing behind agent jobs) |
| preflight | 59 s | about 15 s |
| builder time on gates | 13.8 h in 3.7 days | about 5 h for the same merges |
| added: T2 smoke after the push | none | 10 wall minutes of two slots on the long lane per behaviour merge, batched: one smoke run covers every merge pushed while it ran |
| added: nightly | none | about 6 hours of slot time (T3) plus 1 to 2 hours (pairs, T4), between 23:00 and 07:00 EST |

The estimate that matters most is the first row and it is the least certain. The card measures the hit rate over 20 real merges before the switch is made default.

### 4.4 Sharding across hosts

| work | host | why |
|---|---|---|
| merge gate | builder, main lane, reserved | one writer, one cache, low priority container (6 CPU, 8 GB; S ops builder doc) |
| agents' own package tests | head box with `-p 1` for one package; the Mac build host for more (10 cores, load 1.2 at the time of reading; V `labrun hosts`; pure Go tests need no game) | takes agents off the main lane |
| T2 smoke, the agent's own scenario | builder long lane, raised from one job to two slots when memory allows (a lab d2d holds 0.7 to 0.8 GB, a bot 80 MB; S ops builder doc) | the long lane is the scarce resource today: three jobs waiting |
| T3, T4, pairs | long lane at night; the second worker when the owner allows it (Q4) | 16 cores idle against a 6 CPU container |
| nightly whole-repo race | the Mac build host or the second worker | keeps the builder free for scenarios |

A result from another host counts for determinism only after the d2d binaries are the same file by hash (H10) and one cross-host pair has been compared.

### 4.5 Contention: reservations, priorities, backpressure

1. One locked job per merge (the existing `--gate` form or an equivalent wrapper), so nothing interleaves.
2. Merge class on the main lane: a job submitted with the merge class goes to the head of the main queue. It never stops a running job.
3. Agent quota, enforced in the launcher: per agent at most one running and one waiting job on the main lane, and two on the long lane. Over quota the launcher exits 75 with the queue shown. A queue of 3 to 7 jobs from one agent becomes impossible, and an agent sees the wait before it plans around it.
4. Whole-repo `go vet ./...` or `go test ./...` from a branch is refused on the main lane with a pointer ("the merge runs this; test your packages"). An override flag exists for the coordinator.
5. A wait budget on every agent job: over 20 minutes in the queue, the job is cancelled and the agent told to go on with offline work (AGENTS.md already says never block on a long run).
6. Load guard: the gate does not START while the 1-minute load of the builder host is above a limit (start with 3 times the container's CPU count), up to 10 minutes; after that it runs and marks the result `contended`. A failure in a `contended` run with only timeout-shaped failures is re-run once and logged as a flake candidate; any other failure fails at once.
7. Scenario jobs at night belong to ONE scheduler job, not to agents.

### 4.6 Scenario gate at merge: stages

| stage | rule | exit criterion |
|---|---|---|
| 0 (now) | T0 and T1 in the Go gate | |
| 1 | T2 after the push, report-only, into the ledger and onto the page | 7 consecutive days with the smoke set green on `main` and no flake event |
| 2 | T2 blocking, before the push, only for merges whose changed files match a smoke scenario's `touches`; one re-run allowed, which quarantines (2.7) | owner decision (Q2) |
| 3 | a behaviour card must name a stable scenario; the agent's branch run id is checked by the merge | owner decision; ADR 0020 item 1 already states the intent |

ADR 0020 says bot changes merge only with a green scenario. Today that is enforced by convention on scenarios that are not yet deterministic. The stages make the rule true in the order the evidence allows, instead of blocking merges on a flaky gate.

### 4.7 Watchdog and failure surfacing

- Run the merge worker under a service manager with restart on exit (the ops repo already carries a unit and timer for another job). A stopped worker then restarts in seconds and the restart is a log line.
- Heartbeat: the worker touches a file every tick. A separate timer checks three things every 5 minutes: heartbeat older than 10 minutes; an item `running` for more than 45 minutes; items queued and nothing started for 30 minutes while CI of main is green. Any of them sends the existing phone push once and sets a red tile on the board.
- Replace the single 3600 s timeout by phase budgets (preflight 5 minutes, gate 30, push and CI poll 12) so a hang names its phase.
- Store with every failed gate: the failing test names, the last 60 lines, the builder load at start and end, the job id. Append one line per failure to a flake ledger; two non-reproducing failures of one test put it on the quarantine list (`tools/ci/flaky.txt` exists and is empty).
- `mergeq status` exits non-zero when a watchdog condition holds, so the coordinator's own tick notices without reading text.

## 5. One results store

### 5.1 The record

One JSON line per run, appended by the runner itself at the end of `exec-one` (not by the agent, not from a pulled file):

```
{"v":1,"run":"20261121-021455-9c1e/cain.report/Paladin/11","scenario":"cain.report","class":"Paladin","seed":11,
 "tier":"T2","stability":"stable","verdict":"GREEN","reason":"","d1":"9f2c...","d2":"40ec13d9","journal":"",
 "setup":"CHEATED","setup_tags":["quest-word"],"selfcheck":"ok",
 "bot_commit":"5b12ce5e7","bot_dirty":false,"d2d_sha256":"...","host":"builder","lane":"long","parallel":2,
 "load1":7.5,"pace":41.2,"virtual_min":10,"wall_s":31,"cut":false,"writes_after_start":0,
 "event":"nightly","branch":"main","card":"","job":"20261121-021455-9c1e","started":"2026-11-21T07:14:55Z"}
```

Rules: the ledger is append-only; `host` is a role name, never a machine name; a run with `bot_dirty` true or `event` `manual` never changes a scenario's stability; artifacts (state file, bot log, taps, journal) stay on the host that ran them under the run id for 14 days, the ledger line stays for good.

### 5.2 Where it lives and how it is shown

- Source of truth: `pg-results.jsonl` on the board host, written through one small ingest command that the runner calls (so the builder, the second worker and a dev run all land in one file). Plain JSONL because the project is stdlib-only; the board server can index it in memory.
- Board page `/pg`: a grid of scenario by class with the latest verdict, stability, setup tag, D1, host, commit and age; a click opens the last 30 runs with flips marked; filters by quest, class, system and tier; a release view that shows only EARNED and SETUP runs on the chosen commit.
- Repo: one generated file, `docs/process/proving-ground-status.md`, rewritten by the nightly job from the ledger (verdict per scenario and class on `main`, with run ids). It is generated like the STATUS block and never edited by hand.
- Probe docs keep findings and root causes. Their verdict tables stop: a probe doc cites run ids, and the existing eight get a `> status:` pointer to the page once it exists.
- `docs/probes/test-characters.md` rows for scenario heroes come from the ledger (`d2scenario provenance` already prints them with a `DATE` placeholder; the ledger has the dates).

## 6. Harness hygiene

### 6.1 What went wrong (S, from the run records)

| incident | cost |
|---|---|
| heroes raised by stat verbs kept every point banked after the agency P0 change: 55 life at clvl 32, no attack skill | the whole first pass of the second run was RED for setup reasons; the survival harness heroes were crippled the same way |
| Sorceress setup had no weapon; the readiness gate refused the chunk | one RED |
| purse smaller than the kit plus the first shopping; the bank chore moved the purse to the stash | several REDs over three cards (Den, Blood Raven, chicken) |
| Minor potions below the potion floor were sold at the first town visit; a full pack sold down to the spare limit | Countess held for 60 virtual minutes |
| level equal to the `reach:` goal: the hero idled | one RED |
| ground item given before the warp: a GREEN that proved nothing | one false GREEN |

The same repair was then applied scenario by scenario: `resource.pointsDecisions` appears 28 times in the scenario files (V grep), the survival harness sets it in a shell string and has its own weapon rule (`tools/survival/abrun.sh`), and `internal/scenario/setup.go` has no weapon concept at all (V grep).

### 6.2 The hero library and the self-check

A package `internal/scenario/hero` (shared with `tools/ab/prep`) that turns a short declaration into setup verbs:

```
"hero": {"profile": "played", "level": 16, "build": "paladin_default", "purse": "kit+shopping", "potions": "guard+2", "waypoints": "by-level"}
```

- `profile: played` means: stat and skill points spent as the named build's table spends them at that level (data from the plan tables, never a skill id in Go; an unknown class or build is an error, never a Sorceress default); the class's starter weapon equipped; belt and pack potions of the tier the potion floor keeps at that level; purse derived from the kit price and the first-visit shopping list, not a literal; waypoints a hero of that level has.
- `profile: banked` is the explicit opposite, for the points scenarios.
- Anything the scenario still sets by hand overrides the profile and is listed in the provenance text.

Self-check, run on the prep d2d after the verbs and before the save, from game state:

| check | fails when |
|---|---|
| weapon | no weapon in hand for a class or build that needs one |
| points | unspent stat or skill points differ from what the profile declares |
| attack | no attack skill bound on either hand |
| life | max life below a floor for the level (catches a hero with no vitality) |
| potions | belt plus pack healing or mana potions below the guard the bot enforces for the first step |
| purse | purse below the recovery kit price plus the first shopping list |
| goal | the trigger goal is already met at the start |
| order | a ground item was given before the position verb |

A failed check ends the run as BLOCKED with `selfcheck: <name>` in about the 20 seconds a setup takes. The static half (goal, order, unknown build, purse literal below the kit) runs in `d2scenario validate`, so it is T0.

Guard against the next silent default change: one T1 test builds the `played` profile for two classes and asserts the self-check passes against a recorded state; a product default that strands scenario heroes then fails the Go gate of the merge that introduces it, which is where the agency P0 change should have been caught.

### 6.3 Earned saves

A nightly T4 act run from a fresh hero exports its save at milestones (Den done, clvl 12, Cain freed, before Andariel) with its run id, bot commit and d2d hash. A step scenario may then declare `pre.from_save: <milestone>` instead of verbs and is EARNED. Field starts still need the warp and stay CHEATED. U: whether a save from one d2d build loads unchanged on the next; the card starts with that check.

## 7. Rollout

Effort in agent-days of the usual one-PR size; "ops" paths are in the private ops repo.

| # | card | what | owns | days | risk |
|---|---|---|---|---|---|
| 1 | pgprov | result records bot commit, d2d hash, host role, load, pace; a run cut by the wall limit is BLOCKED | `internal/scenario/result.go`, `cmd/d2scenario/run.go` | 0.5 | low |
| 2 | pgdigest | assertion digest D1, `determinism` compares D1 hard and D2 as drift | `internal/scenario/determinism.go`, `cmd/d2scenario/cmds.go` | 1 | low |
| 3 | pgjournal | runner passes the journal variable; `determinism-run -n N`; the pair of 2.5 run and written up | `cmd/d2scenario`, `docs/probes` | 1.5 plus 40 builder minutes | low |
| 4 | pgtaps | scripted taps and API calls on the virtual clock inside the bot | `cmd/d2bot/ladder.go`, `cmd/d2scenario/tap.go`, `internal/scenario` | 2 | medium (touches the bot's answer path) |
| 5 | pgclock | goals clock, farm timers and the soak stop on the env clock; tests for two builds | `cmd/d2bot/d2dbackend.go`, `internal/bot/farm.go`, `internal/goals` | 1.5 | medium |
| 6 | pgstability | stability state, quarantine rule, tier retag per the catalogue, T1 state fixtures | `internal/scenario`, `tools/provingground/scenarios` | 1.5 | low |
| 7 | pghero | hero library with the `played` and `banked` profiles; survival prep uses it | `internal/scenario/hero`, `tools/ab/prep`, `tools/survival` | 2.5 | medium |
| 8 | pgselfcheck | self-check on the prep d2d and its static half in `validate` | `internal/scenario/hero`, `cmd/d2scenario/run.go` | 1 | low |
| 9 | pgledger | the JSONL ledger, the ingest command, results written on the host (no pull loss) | `cmd/d2scenario`, ops board tools | 1.5 | low |
| 10 | pgpage | board page `/pg` and the generated status doc; pointers in the eight probe docs | ops board, `docs/process/proving-ground-status.md` | 2 | low |
| 11 | mgonejob | the merge gate as one locked builder job; failing tests and tail stored; positive ran-marker; flake ledger | ops `tools/merge.sh`, `tools/mergeq.py` | 1 | low |
| 12 | mgcache | cached whole-repo gate measured over 20 merges, then default; nightly uncached whole-repo race with red hold | ops `tools/merge.sh`, builder config | 1.5 | medium |
| 13 | mglanes | merge class, agent quotas, whole-repo refusal, wait budget, load guard | ops launcher and builder lock tools | 2 | medium |
| 14 | mgwatchdog | service unit, heartbeat, three alarms, phase budgets, status exit code | ops `tools/mergeq.py`, `tools/systemd` | 1 | low |
| 15 | pgnightly | one nightly scheduler job: T2, T3, pairs, one T4; stage 1 smoke after each push | ops, `tools/provingground` | 1.5 | medium |
| 16 | pgearned | earned save chain and `pre.from_save` | `internal/scenario`, `tools/provingground` | 3 | high (save portability U) |

Order: 11 and 14 first (they stop today's losses and need nothing else); 1, 2, 3 next (they answer the determinism question); then 7 and 8, 9 and 10, 4 and 5, 6, 12 and 13, 15; 16 last. About 25 agent-days in total. Cards 4, 5 and 7 change bot or harness behaviour and each needs its own scenario run id in the result.

## 8. Open questions for the owner

1. **Is the verdict plus the assertion digest (D0 and D1) the determinism bar for a gating scenario, with the full state digest as a drift report only?** Recommendation: yes. Requiring bit-identical worlds first would hold every gate behind fork and loop work of unknown size.
2. **When does the smoke set block a merge?** Recommendation: stage 1 (report-only after the push) now; stage 2 after 7 green days with no flake event, by your decision then.
3. **Drop `-count=1` from the merge gate and rely on the Go test cache, with an uncached whole-repo race run every night?** Recommendation: yes, after the 20-merge measurement of card 12. The uncached run still happens in CI after every push.
4. **A standing allowance for the second worker at night** (for example 23:00 to 07:00 EST, at most 3 runs, niced, the sandbox container untouched)? The last allowance ended on 2026-11-20 at 9 am EST. Recommendation: yes for T3, T4 and the nightly race run; without it the nightly needs about 8 hours of the single long lane and will collide with agents' own runs.
5. **Agent quotas on the builder** (one running and one waiting on the main lane, two on the long lane, whole-repo runs refused from branches)? Recommendation: yes. It is the cheapest fix for the 30 to 60 minute waits.
6. **Where does the ledger live?** Recommendation: on the board host next to the status files, with one generated status doc in the repo. The alternative (commit every result) adds a commit per run to a repo that already takes about 100 merges a day.
7. **May the tool quarantine a scenario on its own?** Recommendation: yes; a person takes it out of quarantine, never puts it in.
8. **Release evidence: is at least one EARNED act run required for a release cut, and are CHEATED scenarios excluded from release notes?** Recommendation: yes to both, starting with the release after card 16.
9. **Raise the long lane from one job to two slots?** Recommendation: yes when the memory guard allows it (two lab d2d plus bots are under 2 GB), because the long lane, not the main lane, is where scenario work queues.
10. **Should probe docs stop carrying verdict tables once the page exists?** Recommendation: yes; findings stay in probe docs, verdicts are run ids.

### Proposed text (not written anywhere; the owner or the coordinator decides)

For `docs/decisions.md`, if Q1, Q2, Q5 and Q7 are answered yes:

> Proving Ground gate (2026-11-21): a scenario gates only when it is `stable` (5 consecutive runs with equal verdict and assertion digest on unchanged binaries). A flip quarantines it automatically. The smoke tier runs report-only after each push until it has been green for 7 days. Builder: merges have priority on the main lane; an agent holds at most one running and one waiting main-lane job and does not run whole-repo tests from a branch.

For AGENTS.md, section "Your report", one line:

> A Proving Ground verdict is cited as a run id from the results ledger, with its setup tag (EARNED, SETUP, CHEATED, IMPOSSIBLE).

## 9. Not done, and limits of this review

- No run was made. Every determinism statement in section 2 is a code reading plus other cards' run records; the proof is card 3.
- The cost table in 4.3 is an estimate except for the "today" column.
- The builder lock tools and the launcher were read only through their help text and the ops docs; card 13 may find that priorities already exist in some form (U).
- The survival A/B harness was read at the shell level only (`tools/survival/*.sh`); `abreport` and the prep tool were not read. Its non-identical repeats (S `docs/process/second-worker.md` section 5) share B1 to B3 and the campaign goal, not the tapper.
- `decide.Clock` wiring on the d2d path (B4) and map-order use in decision code (B6) were not audited.
- The local `main` of the shared clone was 2517 commits behind the remote; this branch was started from `origin/main` (5b12ce5e7) instead.
