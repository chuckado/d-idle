# 05. Architecture: what to keep, what to rebuild, and whether to start over

> status: REVIEW for the owner (design review 2026-11-19, card fablereview, docs only). Nothing here is decided. Part of [00-summary.md](00-summary.md).
>
> Tags: **[V]** read or computed by me or a helper of this review from the source named; **[S]** another doc's claim; **[U]** unverified or judgment. "Helper F" is the code map of this tree at `989432fcd` (read-only agent, 2026-11-19); its numbers are `wc -l` and `git log` outputs.

## 0. Verdict in eight lines

1. **No complete overhaul.** The parts that are expensive to get right are right: the backend seam, the data layer that reads the player's own game tables, the d2d adapters (thin), the fidelity lane with an emulator, the loot rules engine, the power-against-danger survival model, the watch UI. A rewrite would re-buy all of that and fix none of the actual problem.
2. **Rebuild the decision core** (about 25,000 of 248,000 lines: `internal/bot` goal and loop arbitration, `internal/task` quest handlers). It is an ordered if-chain with incident patches, and it is where every case study in the brief happened.
3. **The root cause is not in the code. It is that almost nothing is run before it is built on.** The repository is 20 days old, has 8,014 commits and about 105 merges a day, and the proven ground is still Normal Act 1 [V `git log`, docs/capabilities.md, docs/probes/treefix3-2026-11-19.md]. Two earlier reviews said "verify before building the next one" [V docs/research/project-review-2026-11-10.md R6]. As a rule it did not hold. It has to become a machine.
4. That machine is a **Proving Ground**: lockstep d2d scenarios (quest chains, act runs, class smokes) that run in minutes, on every merge that touches behaviour and nightly on seeds. "Done" for a behaviour card means its scenario is green there. The engine for it exists (the lockstep fork runs 46 to 62 times real time with the bot attached [V docs/d2d-loop.md]); it is used for survival A/B runs and almost nothing else.
5. **Finish the survival model and make it the one judge** of "can this hero be here": zone choice, quest readiness, the difficulty transition and the build viability estimate of [04](04-stat-skill-agency.md) all ask the same margin. It replaces 54 mechanisms and about 150 knobs by its own design text [V docs/survival/README.md].
6. **Give the fork a small bot API** (level room rows on entry, allocate verbs, object state) instead of treating our own engine like a remote window ([02](02-fidelity-and-gaps.md) section 5).
7. **Keep game.exe as an oracle, drop it as a second product backend** in practice: emulator differentials yes, Wine soak lanes and statistical A:B verdicts no.
8. Cost of the core rebuild: about 45 to 70 agent-cards (60 to 90 minutes each) over three to four weeks, most of it in shadow mode next to the old chain. Cost of a full rewrite: several hundred cards and the loss of a month of fidelity evidence. Not close.

## 1. The size of things (so the verdict has a denominator)

| fact | value | source |
|---|---|---|
| first commit | 2026-10-30 | [V `git log --reverse`] |
| commits, merges | 8,014 commits; 2,438 merges; 736 merges in the last 7 days | [V `git log --merges --since`] |
| Go, non-test / test | 247,771 / 156,912 lines; 106 packages under `internal/`, 56 under `cmd/` | [V `wc -l`, helper F] |
| largest packages (non-test) | watch 24,207; bot 18,517; cmd/d2bot 13,591; town 12,060; combat 9,424 | [V] |
| docs, task cards, knobs | 1,471 markdown files; 644 task files; 419 knob rows | [V `find`, helper F] |
| proven live ground | Normal Act 1; the friend sandbox hero is a Paladin around clvl 16, stuck on the third Act 1 quest | [V docs/probes/search-for-cain-2026-11-17.md] |
| first hour, fresh Paladin | clvl 5 to 6, four recorded runs | [V docs/probes/firsthour.md] |

For scale, read from the survey's clones [V `wc -l` on the read-only clones, 2026-11-19]: kolbot is about 96,000 lines of JavaScript and plays all seven classes through the endgame; Kolbot-SoloPlay adds about 64,000 lines and levels 30 builds from 1 to 99; the original Koolo was about 21,600 lines of Go. Those bots stand on an in-process engine API and fifteen years of community play, and they have no tests, no fidelity lane and no player UI. The comparison is not "we are slow". It is: **our lines went into things other than playing the game.** `internal/watch` alone is larger than all of original Koolo.

## 2. What is good and stays

| part | why it stays | evidence |
|---|---|---|
| Backend seam: `env.Env` (state reader 8 methods, input controller 12, clock 2) and consumer-side contracts in `internal/bot/contracts.go` | behaviour code does not import the d2d client; the "duplicate" pairs are thin adapters (d2dcombat 468 lines over combat 9,424; d2dtown 431 over town 12,060) | [V helper F] |
| Data from the player's own install (`internal/data`): skills, monsters, levels, items, runewords, cube, hirelings | no game data in the repo, and the tables are the same for every class | [V internal/data/skills.go] |
| Emulator-derived skill damage for 131 skills of all seven classes (`internal/power/skilldamage.json`), danger profiles per zone for all three difficulties (`internal/danger`) | this is the asset no public bot has; it is what makes free skill allocation thinkable | [V internal/power; helper G] |
| The survival model design (power, danger, margin, dial, ledger, archetypes) | it names the disease correctly: "every incident gets a new rule or a moved number" | [V docs/survival/README.md] |
| Loot rules engine with three-valued logic and no equip action | better than NIP for our product ([01](01-landscape-and-comparison.md)) | [V docs/lootrules.md] |
| Quest board model: definitions as data, evidence tiers, states derived and never stored | the model is sound; its executor is the weak part | [V docs/questboard.md] |
| Fidelity lane and the upstream relationship | about 25 merged PRs; differentials with tens of thousands of cases | [V docs/upstream/pr-queue.md] |
| Lockstep fork, journal and divergence finder | the base of the Proving Ground | [V docs/d2d-determinism.md] |
| Honesty tags and the store layer (`internal/store`) | they are why this review could be written from the docs | |

## 3. What hurts, with the case studies

### 3.1 Arbitration is an if-chain with five patches for the same disease

`runGoal` in `internal/bot/goal.go` (1,886 lines) is a fixed priority list of about twenty `if step() { return true }` hooks: points, player pause, player action, errand trip, vault keep, active trip, repair trip, belt fill, flail hold, chores with a cooldown, heal-only trip, low-life gate, gear, claim, save point, claim trip, difficulty, farm, board preemption, gold farm, recovery, board, task, then leveling [V goal.go:21-145]. Each hook decides alone from its own state. Where two of them disagree, a later patch adds a guard:

| guard | trigger | what it does | where |
|---|---|---|---|
| flail guard | 6 area changes in 10 minutes without progress | holds the hero 5 minutes, doubling | `internal/bot/flail.go` (added after the Dark Wood ping-pong, 2026-11-18) |
| ping-pong | 4 switches between two areas in 5 virtual minutes | marks both areas cleared | goal.go:1472 |
| explore gave up | 2 explore failures without a kill | marks the area empty | goal.go:1439 |
| border loop | 5 unwanted border crossings in a minute | skips the area 3 to 48 minutes | loop.go:1283 |
| arrival bounce | 3 bounces in 2 minutes | skips the "deeper" area | loop.go:1348 |

[V helper F]. All five detect the same thing after the fact: **subsystems that each made a locally sensible decision and undid each other.** The comment in `flail.go` says it in plain words: "the board, the area-done trip and the waypoint fallback undid each other". The three-hour town ping-pong, the hero that went broke with 128,000 gold in the stash, and the Start tap that was ignored for 13 minutes because `boardStep` was only called from one branch [V docs/questboard.md "Which goals run the board"] are the same class of fault: no single place knows what the hero is committed to and what a change of plan costs.

**The fix is an arbiter, not a sixth guard.** Every subsystem proposes an intent (priority class, urgency, where it must happen, estimated cost, reason). One arbiter picks, with four rules the chain cannot express:

1. **Commitment.** The running intent holds a lease in virtual time; only a higher class (a reflex, a player order) or its own source ends it early.
2. **Travel is priced.** An area change that would undo the previous one inside a window is refused unless the world changed (a kill, experience, a quest fact, gold, the belt), and the refusal is logged with both reasons.
3. **Batching.** When any intent takes the hero to town, every source is asked what it wants while there. Resupply, hand-in, stash, withdraw gold, repair and identify become one visit. The 128,000-gold case disappears because "withdraw" is part of the visit, not a separate chore rule.
4. **A trace.** Every tick yields "chose X; rejected Y because Z". That line is the "why" the watch UI keeps asking for.

Prototype: branch `agent/fable-proto-arbiter`, package `internal/proto/arbiter` (scenario tests for the ping-pong against a first-match baseline, the stash-gold case, preemption, starvation, a golden decision trace). Results are in [00-summary.md](00-summary.md) section "Prototypes". Migration is by shadow, the way the loot rules went in: the arbiter runs beside the chain and logs what it WOULD do; the differences are read on real sandbox logs; then one family of hooks at a time moves over behind a knob.

### 3.2 Quest chains: sound model, unverifiable execution (Search for Cain)

What happened [V docs/probes/search-for-cain-2026-11-17.md, -l0, -l1, -l2, -45, -5b, treefix3]:

| # | failure | class of fault |
|---|---|---|
| 1 | the tree preset is only listed when the hero is within about 13 to 18 cells; `find` explored 100 s per try and three tries parked the chunk | the hero's knowledge model was never designed ([02](02-fidelity-and-gaps.md) section 5) |
| 2 | progress (42 to 79 percent explored) was counted as failure | no notion of progress in the step contract |
| 3 | the stones' lit state read as `Mode: -1` on d2d, so the solver could never see a stone light | a fact the step needs had no reader on this backend |
| 4 | quest items matched by display name; the name in code was wrong for item code `bkd` | identity by label instead of by code (the same fault as loot quality "guessed from labels") |
| 5 | pickup called immediately after the tree click, before the scroll exists | no postcondition wait |
| 6 | everything after the tree had never run anywhere | no end-to-end run before shipping |
| 7 | after three fix merges (#934, #948, #960) the hero stood at the tree for 17 hours; the remaining hypothesis is a `hidden=1` preset row | each fix was tested against a fake built from the last hypothesis, not against the engine |

The data model had the chain right. The executor is hand-written Go per quest (`internal/task/act1.go` `cainNodes`, `cmd/d2bot/cain.go`, `cmd/d2bot/objects.go`, `internal/bot/task.go`) [V helper F], and nothing could check it without a game.

How three public bots do the same step [V helper A: kolbot `core/Common/Cain.js:46-142` and `core/Misc.js:368-402`, jaenster-kolbot `bots/AutoPlay/Questing/TheSearchForCain.js`, Koolo `internal/run/rescue_cain.go`]:

| step | what they do | what we did |
|---|---|---|
| find the tree | ask the engine for the PRESET of object class 30 in the level and walk to within 5 tiles of its coordinates; fetch the object unit again after arriving | explore until an object list near the hero contains class 30 |
| click it | up to 7 attempts; kolbot uses a real mouse click (down, 20 ms, up) for its list of quest chests, which includes the tree, and the interact packet only for ordinary chests; it first walks to an offset of (+1, +2) from the object | one `cmd interact <index>` |
| know it worked | the OBJECT'S STATE changed: kolbot polls `unit.mode` non-zero for up to 1 s per attempt, Koolo loops until the object is no longer selectable | none (the next call is the pickup) |
| take the scroll | poll for the item by CLASS ID (524) on the ground or in the pack for 1 to 3 s, then pick it up; several retries | pick up by display name at once |
| stones | loop over the five stone objects until each has `mode` set; no order knowledge needed | a solver that needed a lit state the backend could not read |

Every row on the left is "a postcondition read from the engine". That is the pattern to copy (an idea, not code): **an interaction is finished when a named fact changes, and the fact is read by id, not by label.** It also gives one more cheap hypothesis for the still-open tree fault: the interact verb may not be the path the engine takes for this object while a pointer click is (U; one lab drill decides it).

Three changes, in this order:

1. **Make chains checkable without a game.** A step declares its action, target (object class by name, item by CODE), precondition and postcondition facts, how the target becomes known (always, when near, room then explore, spawned by a step), and a budget. A static checker then asks, per backend: does every fact have a reader; is every precondition produced by an earlier postcondition; is every step's completion observable; is the chain resumable from facts alone. Faults 2, 3, 4 and 5 above are checker findings. Prototype: `internal/proto/questspec` on branch `agent/fable-proto-arbiter`, with the Cain chain as data and two capability sets (d2d as the probes describe it, and fixed).
2. **Run every chain in lockstep before it reaches a hero** (section 5). Fault 6 and 7 are only found by the engine.
3. **Level room knowledge on entry** (option K1) removes fault 1 for every future quest object.

The quest board's player-facing model (accept, run, hand in, evidence tiers) does not change.

### 3.3 "The bot never estimates": the survival model is half switched on

The survival review of 2026-11-14 found 54 mechanisms and about 150 knobs that compare proxies (clvl, life percent, gold lines) with hand-set numbers, and nine disagreeing notions of "may the hero be here" [V docs/survival/README.md]. The replacement (danger profile, power estimate, margin, dial, ledger) is built through slice S-D and sits behind `bot.zoneModel` = 0. Its A/B record: Paladin passes on the 8 tuning seeds (experience ratio 1.005, chickens 15 against 2) and on holdout 2; one holdout failed because the chickens came from the recovery state, which bypasses the model; the Sorceress has never reached the three-pair minimum with a pass [V docs/probes/survival-ab-reask-2026-11-17.md; helper G table].

Reading: the model is not the problem. The holdout failure is the if-chain again (recovery picks its own area and never asks). Recommendation: do not tune per class by seed sets forever. (a) Route EVERY area choice through the model (recovery, gold farm, quest travel) as part of the arbiter work, so there is one judge; (b) make its inputs real (weapon, defence and block read live are still "derived" or "default" [V docs/survival/power.md]); (c) then run the class matrix in the Proving Ground. The per-class tuning burden the brief worries about is mostly archetype data plus the capability table of [04](04-stat-skill-agency.md), not seven hand-tuned controllers.

### 3.4 Knobs and docs as load

419 knobs is a symptom of 3.1 and 3.3: each incident got a number. The arbiter and the margin delete families of them (cooldowns, level floors, bans). 1,471 docs and 644 cards for 20 days of work mean an agent, or the owner, cannot hold the system; the diet decided on 2026-11-07 (D5) has not happened [V the counts]. I do not propose another docs rule. The Proving Ground report becomes the status page: what is green there is what the bot can do.

## 4. The engine boundary: we own a fork and use it like a remote desktop

| today | better | why it is allowed |
|---|---|---|
| stat and skill points spent by clicking UI buttons and icons at table coordinates, on d2d too | the fork's `stat` and `skillpt` verbs (they exist; no non-test caller) [V helper G] | same game rule path; a measured run counts an input verb as input, not a write |
| object lists only near the hero; room rows ignored by the parser until 2026-11-18 | a read verb for the level's room rows on entry (K1) | read only |
| object state (stone lit) from `debug objects`, "may be lab-only" [S search-for-cain-45] | object state in the normal dump | read only |
| ground item quality guessed from the label text [V docs/lootrules.md section 1] | quality, item level and flags in the ground dump (the vault's capture verbs already read a full item [S docs/design/loot-vault-2026-11-16.md]) | read only; a player sees the label colour, the engine knows the quality |
| difficulty start through a lab cheat | a start option or the upstream picker | normal game flow |

Policy line to record (proposed, question Q5): "The fork may expose anything the engine already knows to the bot as a READ. Reads never make a run CHEATED. What the hero is allowed to ACT on is a product decision (K0, K1, K2)." The Game.exe memory backend can already read all of it; d2d was held to a stricter standard by accident.

### 4a. The input layer (added on the owner's note of 2026-11-19; full audit in [07-input-layer.md](07-input-layer.md))

The owner observed that the bot clicks where it could use hotkeys. The audit [V helper I, helper H] says:

- On d2d, combat already uses direct verbs (cast, belt, pickup, move, interact). The clicks are in: aura and left-skill selection through the skill popup (because `cmd select` does not set the right skill, deviation R08), stat and skill points (the verbs exist, unused), town portal through the inventory (one incident: a dropped cast that spent the charge), vendor tabs and cells, selling, repair, gold, identify, cube, waypoint rows, and a priming click after every level change (two bounce incidents). On Game.exe skills are F-keys the bot binds itself with read-back and potions are keys 1 to 4; the rest is pixel clicks.
- kolbot and d2bsng press no skill hotkey at all. They send the game's own client messages (select skill, cast, belt, pick up, interact, buy, sell, add point, take waypoint) and fall back to clicks. Hotkeys are what bots without engine access use.
- A hover-first click on d2d costs about 600 ms; an aura switch through the popup is two of them; on the watch video the popup and inventory flash open.

Architecture decision proposed (question Q13): **the fork's action verbs are the game's client-to-server messages, one verb per message, routed to the same handler; the bot prefers verb, then key, then verified click; behaviour packages talk only to the intent-level `SemanticInput` interface (it exists and has one production caller); every intent ends with a postcondition read from state.** A verb equal to a message a real client can send is input, not a cheat, which also gives the crutch rule a precise definition of "the real input channel" on d2d. Each verb is promoted from "lab" to "input" by a differential (UI path against verb, identical state for N ticks). No binding bootstrap is needed on d2d; a hotkey map GENERATED from the planner's slots ([04](04-stat-skill-agency.md)) replaces the hand-written `layout` in the profiles, because with player-chosen skills a hand-written layout cannot exist. Cost: about 10 to 14 bot cards and 6 to 8 small fork verbs. Success measure in the Proving Ground: pixel clicks per virtual hour on d2d, target zero.

## 5. The Proving Ground (the one thing I would build first)

What exists [V]: `--lockstep` fork, 46 to 62 times real time with the bot attached (8 virtual hours in about 8 wall minutes) (docs/d2d-loop.md); `tools/smoke.sh` and `cmd/d2smoke` with a rules file, "never run on the game yet" (docs/smoke.md); `internal/d2dreplay`; the survival A/B harness and `abreport`; the journal and `d2ddet compare` (docs/d2d-determinism.md); the builder with a long lane (docs/process/build-hosts-2026-11-14.md).

What is missing is the product that ties them together:

| layer | content | wall time | when it runs |
|---|---|---|---|
| L0 static | questspec checker over every chain and backend capability set; build-file validation; capability table against the skill tables | seconds | every merge (CI) |
| L1 scenario | one lockstep scenario per quest chunk and per primitive, from a prepared save (a hero parked one step before, created by the normal path and recorded under the provenance rule): "Cain: scroll" passes when the quest fact changes within a virtual budget | 1 to 3 minutes each | every merge that touches `internal/bot`, `task`, `questboard`, `nav`, `town`, `combat`; the set is chosen by the files touched |
| L2 act run | fresh hero, class x build x seed, plays until the act boss dies or a virtual cap; records level at milestones, deaths, chickens, stalls, the arbiter trace of every stall | 10 to 30 minutes per run at 50x | nightly on the long lane: 2 classes x 3 seeds today, growing with the class roadmap |
| L3 matrix | per class gate: every skill in the capability table at "supported" is cast in a scripted fight and its effect read back | minutes | when a class or skill tier changes |

Rules that make it bite:

- A behaviour card names its L1 scenario in the task; the merge queue runs it; no green, no merge. A feature that cannot have a scenario yet (a new act) merges behind a knob that the release builds do not turn on.
- The board, the changelog and `docs/capabilities.md` take "verified" from the Proving Ground report, never from a sentence in a result line.
- The friend sandbox gets only builds whose L2 run passed. The sandbox is where the 17-hour stall was found; a 3-minute scenario would have found it first.

Comparison with d2bsng (correcting the stage-1 survey) [V helper A read the clone: `CLAUDE.md` "Simulator", `docs/host_hooks.md`]: the simulator is NOT in the public repository. It is built from a separate private repository that links the public runtime and runs the real 1.14d Game.exe in a console process on a virtual clock; the public repo holds only ten link-time host hooks (script thread start and join, a call quantum on native calls, an observer that can hash or print each native result, a frame runner). The doc states "runs stay deterministic" and 157 to 175 times real time on a fresh Sorceress (3 x 600 virtual seconds); HOW the clock, the random seed and the frame stepping are made deterministic is not documented, the hashing-for-replay use is an inference, and the leaderboard scoring is not described. Its public tests are 188 doctest cases (pathfinding against about 20 captured collision grids with a reference A*, the frame loop against fakes), none of which plays the game. So d2bsng is a precedent for the IDEA (a bot regression run on a virtual clock, scored) and not a design we can read. We are in fact further along in the open: our lockstep fork, the journal with state-hash checkpoints and the divergence finder are documented and ours. Two lessons still transfer: (1) **determinism of the whole loop is a feature worth finishing**. Ours diverges at tick 2 for bot-side reasons that are mostly identified (the guard goroutine fixed; the goal manager clock and farm timers still wall-clocked) [V docs/d2d-determinism.md section 2]. With a deterministic loop a failing seed is a reproducible bug report and a fix can be proven on the same seed. Budget 3 to 5 cards. (2) **A scored run as the public face of quality**: one number per build and class per night.

Capacity: an L2 run needs one core for d2d and one for the bot for 10 to 30 minutes. Six runs a night are well inside the long lane. This is not a throughput problem.

## 6. Two backends

Decisions already say d2d is the product and game.exe the oracle (ADR 0009, 0013, current truth). In the code the Game.exe path is still a full second backend (`internal/mem`, `memfd`, `input`, `viewport`, `launch`, `charsel`, `cmd/d2game`, pixel and panel logic in `invdesk`) and its interaction model (hover, click, panel rows) leaks into shared code [V helper F]. The acceptance text still requires a statistical A:B between 6 game.exe and 6 d2d lanes.

Recommendation: keep reading game.exe (the emulator differential, the LAN `.d2s` oracle), stop planning to PLAY on it. Concretely: no new feature needs to work on the Game.exe backend; the shared interfaces may grow d2d-shaped methods (room rows, allocate verbs) with "unsupported" on the old path; the A:B acceptance is replaced by differential coverage on the hero's path ([02](02-fidelity-and-gaps.md) section 3). This needs the owner's decision (Q7) because ADR 0015's acceptance names the A:B.

## 7. Overhaul options, honestly compared

| option | what it is | cost (agent-cards) | risk | verdict |
|---|---|---|---|---|
| O1 keep patching | another guard per incident | low per week, unbounded in total | every new class and act multiplies interactions; seven classes x five acts x three difficulties will not survive this | no |
| O2 targeted core rebuild (recommended) | Proving Ground; arbiter in shadow then live; questspec and checker; survival model as the single judge; fork bot API; capability-driven combat selection ([04](04-stat-skill-agency.md)) | 45 to 70 over 3 to 4 weeks | medium: shadow mode and the Proving Ground bound it; the old chain stays behind a knob until the arbiter wins on logs | yes |
| O3 new bot core, keep adapters and data | rewrite `internal/bot`, `task`, `combat` selection from scratch on the existing seams | 150 to 250 | high: the incident knowledge in today's hooks (each comment cites a real failure) is lost unless ported as scenarios first; but if O2's arbiter shows the hooks cannot be wrapped as sources, this is the fallback, and the Proving Ground scenarios are then its test suite | fallback only |
| O4 complete overhaul | new repo or new language, in-process bot inside the fork | several hundred | highest; discards the seam that lets the same bot read game.exe; GPL boundary (ADR 0003) says no d2d code in this repo | no |

An in-process bot (the D2BS shape) deserves one honest paragraph because it is why kolbot's scripts are short. Inside the engine every question is a function call. We do not need to move in to get that: the control channel can answer the same questions if the fork exposes them (section 4). Speed is the remaining difference (157x against about 50x), and that is not our bottleneck.

## 8. Merge and builder throughput

| measure | value | source |
|---|---|---|
| merges to main | 736 in 7 days, about 105 a day | [V `git log`] |
| agent merges that changed no Go (week to 2026-11-10) | 252 of 440 | [S project-review-2026-11-10] |
| shared generated files rewritten per merge | five (STATUS block, BACKLOG, CHANGELOG, OWNERS, INDEX), now generated at merge | [V AGENTS.md, user decision 2026-11-11] |

The pipeline is not slow; it is fast at merging work that has not been run. Three changes:

1. **Change the unit of work from a slice of a layer to a slice of the game.** "Quest board S1 to S10" and "Vault V0 to V7" were built layer by layer, each "offline, NOT verified live" [V docs/questboard.md status lines]. A vertical card is "the hero completes Den of Evil through the board on d2d, scenario green", and it touches model, executor, API and UI at once. Fewer, larger cards (the 400-line rule in `docs/agents/rules-short.md` pushes the other way and should bend for these).
2. **One gate that matters more than the rest**: the L1 scenario set. The current gate set proves the code compiles, is formatted, leaks nothing and passes fakes. Fakes encode the author's belief about the engine; the Cain fixes passed all of them.
3. **Builder etiquette stays** (one machine, merge gates first, long lane for long jobs). The Proving Ground adds minutes per behaviour merge and one nightly long-lane job; it removes the multi-day loops of "fix, deploy to the sandbox, wait, read 20 hours of logs".

## 9. Test strategy to stop whole-class regressions

| risk | guard |
|---|---|
| a change for class A breaks class B | L2 nightly matrix over every class at tier "playable"; L3 skill matrix per class |
| a build or capability table edit changes behaviour silently | golden "derived plan" per shipped build at clvl 6, 12, 18, 24, 30 (the freebuild prototype's agreement table is the first version of this) |
| a rule only passes for the default Sorceress | already an AGENTS.md rule (two builds per test); extend to "one caster, one melee, one summoner" once the Necromancer is playable |
| a quest chain edit breaks a later chunk | L0 checker plus the chain's L1 scenarios |
| arbitration regressions | golden decision traces for the incident scenarios (ping-pong, broke with stash gold, ignored Start tap, recovery bypass) |

## 10. What could go wrong with this plan

- The lockstep fork diverges from what friends run (lab verbs on or off, upstream rebases). Mitigation: the Proving Ground runs the SAME binary and flags as the release (the smoke recipe already runs with lab verbs off).
- Scenario saves rot when the engine or the save format changes. Mitigation: saves are produced by a script from a fresh hero (normal path), not checked in as blobs.
- The arbiter becomes a second if-chain with nicer names. Mitigation: sources may not read each other's state; the only shared inputs are world facts and the arbiter's own trace. The prototype enforces this by construction.
- Shadow mode never ends. Mitigation: a dated exit rule per hook family (the loot rules precedent has a per-stage switch).
- K1 knowledge makes the hero feel omniscient. Mitigation: it is rooms, not monsters; the watch UI's senses tiers still gate what the PLAYER sees.
