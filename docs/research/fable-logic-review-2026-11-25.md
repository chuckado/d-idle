# End-to-end logic review of the bot (2026-11-25)

> status: REVIEW for the owner (Fable, card fablereview2, TASK-953). Docs only: no Go changed, no game run, nothing posted. Nothing here is decided. It follows the design review of 2026-11-19 (`docs/design/fable-review-2026-11-19/`) and the three of 2026-11-21 (`docs/design/fable-review-2026-11-21/`); it does not repeat them, it measures what happened since.
>
> The owner's ask: "an end to end review of the bot's logic, look for room for improvements, ways to make it more reliable / controllable. As it stands it seems like we are constantly chasing edge cases." Added the same day: research other Diablo II bots, other ARPG bots and game AI for patterns to take.
>
> Tags: **[V]** read or run by me in this tree at `b8cc86348` (file and line, command, or a log line of the two sandbox heroes read on 2026-11-25); **[S]** another doc's claim or my own prior knowledge of a public project, not re-read; **[U]** unverified or judgment. Log times are the log's own UTC stamps. Sizes: **S** = 1 to 2 agent cards (under one Sonnet agent-day), **M** = 3 to 6 cards (2 to 3 days), **L** = 7 to 15 cards (about a week).

## Executive summary

1. **The diagnosis of 2026-11-19 was right and was not acted on.** Three structural replacements sit switched off: the intent arbiter (`bot.arbiterShadow`, no commit since 2026-11-19), the one supply need (`town.plan` 0) and the stored quest runs (`quests.boardSM` 0). In the same six days the legacy path took about 76 patch branches for one quest (`toolstrade` to `toolstrade76`), 58 new knob rows and 2,000 more lines in `internal/bot` [V section 2].
2. **There is no single arbiter.** A decision passes about 16 checks in `step`, about 30 ordered hooks in `runGoal`, 11 cancel causes in the guard and a failsafe ladder; state is 98 loop fields and 53 "not before T" timestamps [V section 1].
3. **Top cause classes** in 178 fix cards since 2026-11-04 (a card can count in several): supply and economy 49, per-quest special cases 45, navigation geometry 39, two deciders undoing each other 29, class-specific combat 23, thresholds moved by one incident 20 [V section 2].
4. **A refusal is still treated as a failure.** The quest sandbox logged 296 `action failed ... not ready` warnings in nine hours, 226 in two hours; four separate exemptions try to keep refusals off the ladder [V section 3].
5. **Safety is not an independent layer.** A reflex works by cancelling the running action and hoping it returns. The newest `bug:true` death: 23 s under the drink line with 3 healing potions, inside a town chore that was running in a field [V section 3].
6. **Lab green, live red:** lab heroes are constructed, live heroes are grown (clvl 12, median 35 gold, empty belt). The Proving Ground is not in the merge gate [V section 5].
7. **The operator can start and stop quests but cannot see why.** `/api/why` gives one reason code; rejected alternatives and refusals are nowhere [V section 4].

**The one structural change to do first:** make a step's result a typed outcome, where a refusal is data that names what it needs and the fact that would end it, handled in one place (proposal P2). It removes the refusal-as-failure class, replaces timers by wake conditions, and is the record an explain page and the arbiter both read. Before it, one small non-structural PR: an invariant monitor that only logs (P1).

**Not to do:** a rewrite; a sixth loop guard; a behaviour-tree or planner library; another floor moved after one death; another knob per fix.

## 1. Architecture map: how a decision is made today

### 1.1 The path of one tick [V `internal/bot/loop.go:60-141`, `goal.go:23-145`, `loop.go:1045-1118`, `loop.go:403-451`]

```
Run loop (one goroutine, the only one that sends input)
  step():   frozen? -> settings -> poll snapshot -> game lost? -> new goal? -> stuck ladder? -> not in game?
            -> death pending / dead wait -> campaign stall -> drought -> finish recovery -> corpse in town
            -> player reset -> no goal -> error wait -> guarded exit -> paused (hazard or paused fight)
  runGoal(): prepare -> points -> checkpoint -> player pause -> recall | player items | errand | vault keep
            -> idle rest -> running trip -> repair trip -> belt fill -> flail hold -> chores (cooldown, emergency)
            -> low-life gate -> gear | claim -> save point -> claim trip -> difficulty -> farm runs
            -> board preempts -> board holds area -> gold farm -> recovery -> board -> task -> target/travel/sweep
  do(action): runs ONE action under a context; a guard watches it every bot.watchMs and may cancel it with a cause:
            game lost, left game, died, pause, SAFETY(drink/portal/chicken/crowd), drought, re-plan, player command,
            checkpoint, points, trip due, stuck
  after the action: do() maps (error, cause) to a reaction: carry out the safety action, open an emergency trip,
            wait on a gate, count a failure, or climb the ladder (skip target, reset to town, leave and re-enter,
            restart the process, pause)
```

Perception is one snapshot per tick plus an observer with its own mutex (`observe.go`, 1,275 lines). Safety is decided by a pure function (`combat.Decide` through `Fighter.Safety`, `internal/combat/fighter.go:240`), which is good; it is CARRIED OUT only by the guard cancelling the action, or by the fighter's own loop, or by `hazard` while paused.

### 1.2 Who arbitrates what

| concern | where it is decided | its own state |
|---|---|---|
| town trip (potions, pack, repair, heal) | `needChores` hook plus the guard's `tripDue`, plus `beltFillStep`, `repairTripStep`, `lowLifeGate`, emergency trips opened inside `do` and `handleSafety` | `trip`, `lastTrip`, `forceTrip`, `beltFillNext`, `repairRetryAt`, `healFails`, `noPortalSince`, `lowLifeUntil` |
| quest step | `boardPreempts`, `boardHoldsArea`, `boardStep`, `taskStep` (four hooks at three places in the chain) | `task`, `questPending`, `questTries`, the board's runtime, the plan file |
| may the hero enter this area | at least eleven judges: gear gate, kit gate, route guard, early road, Barracks and Forgotten Tower floors, Den floor, Andariel floor, readiness, quest floors (`quests.legacyWall`), area bans, the zone model in shadow | each its own notes, bans and data source |
| danger | `combat.Decide`, crowd rules, cast guard, potion burst, stale pack, checkpoint | fighter drink times, `lastSafety`, `tpSeen` |
| death and corpse | `openRecovery`, `finishRecovery`, `corpseTownStep`, `corpseSweep`, the death rule | `rec`, `corpseTry`, the corpse-in-town record |
| poverty | recovery, gold farm, poverty mode, the kit gate, the shop withdrawal | `broke`, `starved`, `rebuyBar`, `shopGold` |
| leveling area | plan table, `hold` hysteresis, the zone model (shadow), area skips | `stepArea`, `stepLevel`, `cleared`, bans |
| loops between all of them | flail guard, ping-pong, explore gave up, border loop, arrival bounce | five separate memories |

Answer to "one arbiter or many triggers": **many.** The order of the `if` chain is the only arbitration, each hook reads its own private state, and five after-the-fact guards catch hooks that undo each other (named in the 2026-11-19 review; all five are still there, and the flail guard fired twice in the quest sandbox on 2026-11-26, `flailing: holding`, cycle "Dark Wood <-> Rogue Encampment: 6 area changes in 2 min") [V log].

### 1.3 Where state is duplicated or implicit [V counted with grep on non-test files]

| measure | value |
|---|---|
| fields in `loopState` (`internal/bot/runner.go:238`) | 98 (96 on 2026-11-19) |
| timestamp fields of the form "not before T" in `internal/bot` | 53 |
| duration constants written in Go in `internal/bot` | 72 |
| retry or strike counters incremented in bot, task, town, nav, questboard, boss, cmd/d2bot | 27 sites |
| area-id literals compared in Go (bot, task, town, cmd/d2bot) | 19 lines, for example `kitSafeTier = {2, 3}` and `dungeonMouths` in `kitgate.go:18-22` |
| functions named after one place in `internal/bot` | `barracksGuardWhy`, `towerGuardWhy`, `denFloor`, `fillAndariel`, `andarielLow`, `andarielKitOn`, `bossPurse` |
| control flow by error wording | `planningFault` matches "not ready" (`task.go:1298`); the watch hub matches the wording of the gate error (`internal/watch/hub.go:652`) |
| knob rows in `docs/knobs.md` | 560 (502 at the 2026-11-19 review commit) |

The implicit state machine is "which timestamp has not expired yet". No table says which modes exist, what ends each one and how long each may last.

## 2. The edge-case treadmill

### 2.1 Method

178 items: every task card created since 2026-11-04 labelled `bug` or `bot-fix` outside the watch UI and the fidelity lane (162), plus the 16 `toolstrade52` to `toolstrade76` fix commits that have no card on main yet. Classified by keywords on the title, so a card can fall in several classes and 20 fall in none; I read every title and hand-checked the class lists. It is a sizing, not an audit [V `backlog/tasks/`, `git log --since=2026-11-24`; U the exact numbers].

### 2.2 Counts

| # | cause class | items | example titles |
|---|---|---|---|
| C | supply and economy: several definitions of "needs" and "can afford" | 49 | "every town visit buys up to the route guard kit, stash gold pays"; "kit mana minimum vs belt plan"; "the poverty advice asks the town for what its sell chore sells (advice counted 4, the chore sold 0)" |
| H | per-quest or per-boss special case | 45 | "the Malus comes from the Malus stand"; "Andariel's run clears the lair door crowd first"; "the Smith run sweeps the arrival crowd first" |
| D | navigation geometry and walk stalls | 39 | "ExitTo door openings are not attempts"; "unproven border strip is a frontier"; "a door that already reads open is not used again" |
| A | two deciders undo each other | 29 | "no home loop for a thinned kit"; "a stale-pack hold does not send a board chunk's trip home to the first field"; "Den of Evil and Blood Moor shuttle: potion-count flip evicts the hero" |
| I | class-specific combat | 23 | "Paladin lands no hit in lockstep"; "Amazon with no javelins and no gold stands idle swinging" |
| G | threshold moved by one incident | 20 | "the Dark Wood wants 180 life (melee) / 100 (caster)"; "Andariel floors recalibrated"; "the Barracks life floor is waived four levels above the level floor" |
| F | a fact the backend does not give or gives wrong | 17 | "d2d source never sets Snapshot.Hovered"; "tree class from the preset" |
| B | a refusal counted as a failure, or a retry counter | 16 | "chickens count against their own limit, not the three of a death"; "hold_timeout keeps the readiness reason"; "retry only a parked chunk" |
| K | the failure was not visible | 16 | "Cain scroll chunk deferred silently by the recovery"; "why the Tower Cellars are not farmed" |
| E | stale state after portal, death, new game | 12 | "a worn cap left the body on a pickup: unwear matched the ground item id against a worn item with the same number" |
| J | tests, races, flakes | 11 | |

### 2.3 The clearest exhibit: one threshold, one day [V `git log --since=2026-11-25`]

| commit | rule for the road to the Dark Wood |
|---|---|
| toolstrade67 | a floor-only chunk reads too hard more than two levels under its floor |
| toolstrade69 | early road guard as data: class level and kit |
| toolstrade70 | Cain scroll floor 11 melee, 13 caster, 13 summoner ("a clvl 8 Paladin died twice") |
| toolstrade72 | also max life 110 melee, 70 caster ("a clvl 8 Paladin with 69 life died twice") |
| toolstrade74 | the early road is a Ready hold, not a strike |
| toolstrade75 | the Dark Wood wants 180 life (melee), 100 (caster) |

Five of the seven deaths of the quest sandbox hero that day are in the Dark Wood on the same quest step [V log]. Each commit is locally sensible. Together they are a hand-tuned number chasing a sample of one, while the survival model that computes this margin for every area logs `zone shadow ... word=safe` beside it and is not asked on quest travel [V log; `bot.zoneModel` 0].

### 2.4 Top five classes and the property that removes each

| class | why it keeps producing cards | structural property that makes it impossible |
|---|---|---|
| **C supply/economy (49)** | the chores hook, the kit gate, the route guard, the early road, the shop, the recovery and poverty mode each count potions, gold and scrolls their own way [V the titles are mostly "X now counts like Y"] | ONE count and ONE need per excursion, read by every guard and the shop. It exists (`internal/supply`, ADR 0028) at `town.plan` 0. The property is reached by finishing and switching it on, then deleting the other counters |
| **H per-quest special case (45)** | a quest step is hand-written Go; success is inferred, not read | a step declares its target by id and its postcondition as a fact; one generic executor does approach, interact, wait for the fact, retry. `internal/questspec` holds the declaration side; the executor side is still per quest |
| **D nav geometry (39)** | every new map shape adds a rule to the walker (doors, strips, mouths) | not removable by structure; reducible by (a) asking the engine for more (ADR 0022 read verbs) and (b) one walk contract with progress as the only success signal. I rank this class below the others on purpose: it is real game complexity, the others are self-made |
| **A deciders undo each other (29)** | no place knows what the hero is committed to | one arbiter with leases and priced travel (ADR 0020 item 3). Exists in shadow; see 6.3 for why its migration rule cannot be met as written |
| **G thresholds by incident (20)** plus **B refusal as failure (16)** | a gate is a number in a row or a knob, its refusal is an error, and the answer to a death is a new number | one admission function whose verdict is data (ok, or refused with a NEED), numbers derived from one margin model with per-row overrides that carry their evidence count, and a refusal that never touches a failure counter |

### 2.5 Why the treadmill runs (the process side) [V]

- ADR 0020 says bot changes merge only with a green scenario; the pipeline audit of 2026-11-24 found the merge worker runs no scenario tier and that fixes merged while their scenario was red (`docs/process/pipeline-audit-2026-11-24.md` row 6, repeated in the "Current truth" block).
- Every structural fix is behind a staged knob with default 0, and every live incident is fixed on the path the default runs. The shadow lines show the cost: 93 rate-limited divergence lines in nine hours say "chain chose quest (or recovery), arbiter would choose resupply/refill (belt N under N)"; the hero then died in the Dark Wood with an empty belt [V log].
- The loop is: live death, a card within the hour, a number or a guard, merge, next death. It is fast and it converges on one hero in one place; it does not converge on a class of heroes.

## 3. Reliability

### 3.1 Invariants and where they are (not) enforced

| invariant | status | evidence |
|---|---|---|
| never leave town without the kit | partly: kit gate for a recovering hero, route guard per named road, early road per two areas; no single check at the town exit | deaths at 19:04 (Underground Passage, clvl 5, belt 0 healing, 0 scrolls) and 22:10 and 00:56 (Dark Wood, belt 0) [V log] |
| never leave town hurt | not enforced at a chokepoint | 21 of 277 field entries in nine hours were under 50 percent life (`zone entry` lines); two at `life=0/176`, the second 21 s before the death below [V log] |
| a town chore runs in town | not enforced per step | at 02:51:15 the hero arrives in Cold Plains while the trip is still in its chores stage; the shop chore walks 21 s toward a town NPC "not in the unit table", state `resupplying`, 12 hostiles; death at 02:51:36 [V log] |
| a safety decision is carried out within a bounded time | no bound, no measurement | same death: `lastSafety=chicken`, `bug=true`, "3 healing potions in the belt and none drunk in the last minute (23 s in the field, 23 s under the drink line)". Root cause U; the structural reason is that a reflex needs the running action to return first |
| a refusal is not a failure | four partial exemptions (`errAreaGated` in `do`, `planningFault`, `parkStep`, `noteStepRefused`), none covers the board step | 296 `action failed what="quest board step" err="not ready: early road ..."` warnings, `consecutive` up to 4 of the 5 that start the ladder (`bot.maxStepFailures` 5) [V log, `loop.go:472`, `questboard.go:911`] |
| never walk a route below its floor | enforced per named route, in Go functions per place | section 1.3 |
| a corpse is retrieved or consciously abandoned | mostly: retries, fresh game, town corpse step, a sweep every minute | the sweep was added on 2026-11-25 after "a death record read Equipped 0 with five pieces on the corpse" (toolstrade64): the invariant held only after an incident showed it did not |
| every mode has an exit and a timeout | per mode by hand; no table | about fourteen ways the hero stands still: player pause, frozen pause, self pause, leveler pause, error wait, gated wait, planning block, ban wait, low-life wait, fresh hold, flail hold, board holds area, parked chunk, await hand-in. Each has its own exit rule |
| the purse recovers | not an invariant at all | nine hours, clvl 1 to 12, median gold 35, maximum 482, 16 `poverty mode: on`, 136 "no withdrawal for the shop, stash gold reads 0", 150 town trips (one per 3.6 minutes), 94 of 527 minutes in town chores [V log] |

### 3.2 Watchdogs that exist [V]

Process: step watchdog with stack dump (`cmd/d2bot/watchdog.go`), heap guard, supervisor restart. Loop: `stuckCheck` and the four-rung ladder with burst caps, campaign stall (warn, fresh game, stop), drought, the five loop guards, the death rule (streak, deaths per hour), repeat-site ban, trip budget, panic-safe step. Detection after the fact is well covered.

### 3.3 Missing

1. **Safety latency**: time from a non-none `Safety` verdict to the potion leaving the belt or the portal cast. Not measured, not bounded, not tested.
2. **Cancellation latency**: no test that every action returns within a bound after its context is cancelled. The whole safety design rests on this property.
3. **Refusal storm**: the same step refused N times with no fact changed (226 in two hours went unnoticed by every watchdog because each refusal "ran").
4. **Mode invariants**: "resupplying outside town for more than a few seconds", "life unreadable or zero while alive".
5. **Economy stall**: gold below the kit cost for H hours while kills continue.
6. **Town thrash**: trips per hour (the flail guard counts area changes without kills; a hero that kills a little between 17 trips an hour passes it).

### 3.4 The two sandbox logs

| | farm sandbox (hero Doofer) | quest sandbox (hero Questy) |
|---|---|---|
| window readable | 58 minutes (2026-11-25 13:17 to 14:15), nothing after | 8.9 hours (2026-11-25 18:18 to 2026-11-26 03:13) |
| deaths, `bug:true` | 0, 0 | 7, 1 |
| chicken exits (Save and Exit, re-enter) | not counted | 77 |
| town trips | 10 | 150 (41 "heal below the resume line", 20 "life low, cannot regenerate") |
| nav stuck / nav stall lines | 41 / 391 | 313 / 2,219 |
| action failed | 0 | 321 (296 refusals) |
| longest stretch without a kill | not computed | under 10 minutes: no long stall, the hero is busy losing |

Reading: the farm hero on a path that has been patched for three weeks is quiet. The fresh quest hero reproduces, in nine hours, most of the cause classes of section 2. That is the gap between "the code handles the cases we have seen" and "the code has the property".

## 4. Controllability

### 4.1 What the owner can steer [V routes in `internal/watch`, `internal/ops/ops.go`, `docs/operator.md`]

| channel | verbs | consistent? |
|---|---|---|
| quest board (viewer) | Start, Stop, Hand in, Retry, auto, Try anyway | yes, one desk (`questboard.Desk`) |
| operator socket (`D2BOT_OPERATOR=1`) | start, retry, cancel, handin, auto, status, start --try-anyway | same desk, audited, rate limited; quest verbs ONLY |
| orders | go to area, fight here, come home (minutes, forever, until level) | separate queue; overlaps the board's "Level up to N" chunks |
| player pause | walk-home pause, frozen pause | two meanings behind one button, documented (`docs/pause.md`) |
| Farm gold, repair, recall, reset | one-purpose orders, each its own hook in the chain | each has its own "waits for a safe moment" rule |
| policy sheet, loot rules, pacts, points, vault | data the hero follows | fine |
| knobs | 560 rows | the real steering wheel in practice, and not discoverable: a behaviour is the product of several (`town.plan`, `quests.boardSM`, `route.earlyRoute`, `quests.legacyWall`, `resource.*`) |

Overlaps: three ways to say "go level" (a leveling goal, an order, a board chunk); two ways to say "stand still" plus about twelve the bot chooses itself (3.1). Missing on the operator socket: hold, come home, go to area, park a quest for N hours, and any read of WHY.

### 4.2 "What is the bot doing and why"

- `/api/why` and the whyline: one of 26 reason codes for the activity that started last (`internal/bot/whyline.go`). It says what, not what else was considered or what is blocking.
- `/api/thoughts`: ladder questions, not decisions.
- The arbiter records "chose X, rejected Y because Z" every tick; `Runner.Arbiter()` (`internal/bot/shadow.go:33`) has no caller outside the package. The only place its reasoning is visible is a rate-limited log line.
- Refusals appear as WARN `action failed`, which reads as a fault.

There is no single place. The cheapest honest answer today is `docker logs | grep`, which is what this review did.

### 4.3 Cost of a decision log and an explain endpoint

| piece | size | note |
|---|---|---|
| a ring of the last 200 decisions: tick time, chosen hook family, reason code, the refusals standing with their needs | S | falls out of P2; without P2 it can only repeat the whyline |
| `GET /api/explain` and operator `status --why` | S | read-only; the socket allowlist grows by a read, which needs a one-line amendment of ADR 0030 |
| arbiter decision in the same record (chosen, leased until, top two rejected) | S | the data exists per tick; 2.6 microseconds per tick measured in `docs/arbiter.md` |
| a "Why not?" row per board chunk on the watch page | M | watch work, after the API |

## 5. Testability

### 5.1 What exists [V]

| layer | what it proves | limit |
|---|---|---|
| unit tests on fakes (`mem.Fake`, rigs, sims; 137 test files in `internal/bot` alone) | the code against the author's model of the game | the model is the last hypothesis (the 2026-11-19 Cain case) |
| offline scenario corpus (`docs/scenarios.md`, six recorded failures as `TestScenario_*`) | a recorded failure stays fixed; has a knob-off proof | six scenarios; hand-built |
| arbiter replay (`cmd/d2arbiter`, one fixture) | the arbiter on a recorded area sequence | the facts in the fixture are assumed, not logged (`docs/arbiter.md`) |
| Proving Ground (`internal/scenario`, 25 files in `tools/provingground/scenarios`) | the real bot on the lockstep engine, verdict from game state | not in the merge gate; setup is declared and CHEATED by design |
| drills, quest drill matrix | one chunk at a time on a stand-in hero | stand-ins: note 69 of the matrix ran Andariel with a clvl 21 stand-in; the table boss has 1,024 life and the model "reads the boss alone as one melee monster" (`docs/probes/andariel-gate-recalibration-2026-11-24.md`) |
| lockstep survival A/B | zone choice on seed sets | only the leveling path |

### 5.2 The gap between lab green and live red

1. **State distribution.** A scenario declares `pre` (level, gold, belt, start). The live hero that fails is the one nobody would declare: clvl 8, 69 life, 4 gold, empty belt, no scroll, a corpse in another area. The failures are in the product of economy state and task state, and the lab samples the diagonal.
2. **Duration.** Class A and C faults need tens of minutes of coupled play (trip, death, poverty, trip). A step scenario ends before the coupling starts.
3. **Oracles.** A scenario asserts its own end state. No scenario fails because the hero made 17 town trips in an hour on the way to green.
4. **Gate.** Nothing runs at merge (2.5).

### 5.3 Regression strategy

| # | what | catches | size |
|---|---|---|---|
| T1 | **Invariant monitors as the universal oracle** (P1): the same checks log live and turn any lockstep run red | every class that shows as a broken invariant, without a scenario author predicting it | S |
| T2 | **Incident to scenario**: at each death, park, flail hold or monitor hit the bot writes the scenario `pre` of that moment (level, life, gold, stash, belt and pack counts, kit, quest bits, area, accepted chunk). A nightly job replays the newest N on the lockstep engine | the state-distribution gap; the scenario corpus grows from live play, not from memory | M |
| T3 | **Property tests on the pure deciders**: `combat.Decide`, `supply.Solve`, `arbiter.Decide`, the admission function of P4. Random inputs; properties such as "a reflex always wins", "more potions never makes the verdict worse", "no intent starves", "a refusal's need, once met, admits" | whole input classes of B, C, G at unit-test cost | S each |
| T4 | **Recorded sessions**: one `tick` line per second with the facts the shadow adapter already collects (`internal/bot/shadow.go`); `d2arbiter -follow` replays them. Makes the arbiter fixtures real | class A before live; gives the migration an outcome measure (6.3) | S |
| T5 | **Soak with fault injection** on lockstep: empty the belt, zero the purse, drop life, fail the portal, time out the control socket (14 such timeouts in the quest log), remove an NPC from the unit list. Monitors are the oracle | E, F and the recovery paths nobody walks on purpose | M; the injections are lab verbs of the fork (ADR 0014 DIAGNOSIS; never in a measured run) |
| T6 | **Cancellation-latency test**: every action registered with `do` is cancelled at random points against a fake; it must return within a bound | the safety design's hidden assumption | S |
| T7 | report-only scenario tier in the merge queue (already proposed by the pipeline audit) | 2.5 | S (ops repo) |

## 6. Prior art and what to steal

Licence rule: ideas only, in our own words and code; nothing was downloaded or run. Where a project is copyleft the note says so; it matters only if someone copies code, which this review does not propose.

### 6.1 One paragraph each

- **kolbot, D2BS scripts** [S, read by the 2026-11-19 survey: `docs/research/public-d2-bots-2026-11-19.md` 2.3, 7.4]. Pattern: a flat library (`Town`, `Pather`, `Pickit`, `Attack`, `Config`) and run scripts; ONE `Town.doChores` in a fixed order; the belt need is two arrays (`BeltColumn`, `MinColumn`); survival is two numbers. Fixes for us: class C (one town pass, one definition of the belt need). Cost: none beyond P5. Risk: its simplicity rests on an in-process engine API. No licence file on most of it: ideas only.
- **Kolbot-SoloPlay** [S, same survey 2.5]. Pattern: a run index where each script carries `preReq`, `skipIf`, `shouldRun` predicates over quest state, level and class; when broke it reruns cheap farms. Fixes: class G and H: readiness as predicates in a table beside the run, not in code per place. Cost: P4 is our form of it.
- **Koolo** (Go, D2R, MIT) [S, survey 5.2 and 2026-11-19 review 3.2]. Pattern: a health manager apart from the run logic, chicken lists as config, object interaction that loops until the object's state changes. Fixes: H (postcondition by state), reflex independence. Cost: P9 and P3.
- **Botty** (Python, vision, MIT) [S, survey 5.1]. Pattern: three threads from `main`: the bot, a health monitor, a death monitor; a screenshot on every anomaly. Fixes: 3.3 item 1: health does not wait for the run logic. Cost: P3. Risk: our single-input rule and lockstep determinism (see P3).
- **Map overlays (MapAssist style)** [S, survey 5.4]. Data layer only; nothing for decisions. The transferable idea is a debug view of what the bot believes, which the senses page already is.
- **Baritone** (Minecraft pathing bot, LGPL-3.0) [V fetched `PathingControlManager.java` 2026-11-25]. Pattern: every behaviour is a process with a priority; each tick the highest active one is asked for a command; it may answer DEFER; a process marked TEMPORARY (a combat pause) takes the tick without resetting the ones below, a non-temporary one calls `onLostControl` on every lower process, which resets their state. Fixes: class A, and class E through the explicit lost-control hook: stale trip or step state after a preemption is cleared by contract instead of by a timestamp. Cost: two fields on the arbiter's `Intent` and one method on `Source`; S inside P6.
- **mod-playerbots** (bots for an open MMO server core, GPL-2.0-or-later) [V fetched `Engine.cpp` 2026-11-25]. Pattern: strategies are named sets of triggers; a trigger pushes actions with a relevance number into a queue; multipliers scale relevance; an action has prerequisites (queued first), alternatives (queued when it fails) and continuers; strategies are switched at run time by a short command (plus name, minus name, question mark to list); a debug strategy reports every executed action with its relevance, and a rolling history of recent actions is kept. Fixes: B (a prerequisite is a need, not an error), controllability (strategies as named, switchable bundles instead of 560 knobs), the decision trace. Cost: the prerequisite idea is P2; named strategy bundles are a later M. Risk: relevance numbers become the next knob jungle if not derived.
- **OpenKore** (Ragnarok Online client bot, GPL-2.0) [V fetched `control/config.txt` 2026-11-25]. Pattern: behaviour is almost entirely declarative: repeatable blocks (`useSelf_item`, `buyAuto`, `teleportAuto_*`) that share ONE condition vocabulary (`hp`, `sp`, `monsters`, `aggressives`, `inTown`, `onlyWhenSafe`, `timeout`). Fixes: G: a threshold is a row with conditions in a shared grammar; the operator can read it. We have this for loot (`docs/lootrules.md`) and not for safety or town. Cost: M for a safety and supply rule table. Risk: users write contradictory rows; it needs the three-valued evaluation the loot rules have.
- **Path of Exile bot frameworks of the Exilebuddy line** [U, from memory, closed source, not fetched]. Pattern: a task manager that walks an ordered list of tasks, each returning "I acted" or "next". It is the same shape as `runGoal`, and that community's plugins grew the same symptom (tasks that fight over town runs). A cautionary match, not something to take.
- **Behaviour trees** (Halo 2, later the TreeSharp library used by several MMO and ARPG bot hosts) [S general knowledge; the bot hosts U]. Pattern: a priority selector re-evaluated from the root every tick, nodes return Success, Failure or RUNNING, and reusable decorators wrap nodes (cooldown, timeout, repeat until). Our chain already is a priority selector. What it lacks is the two other parts: an explicit status vocabulary (P2) and decorators as shared code in place of 53 hand-written timestamps. Take those two ideas; do not take a tree library.
- **Utility AI** [S general knowledge]. Pattern: every option gets a score from a few normalised considerations; inertia favours the running one. The arbiter's urgency inside a class plus its lease is this. Take: derive urgency from measured quantities (margin, belt fraction, gold gap), never hand-set per incident.
- **GOAP, planners** [S general knowledge]. Pattern: actions with preconditions and effects, a planner searches. Planning is not our problem (the route to Andariel is known); checking is. Take only the declaration: a step's preconditions and effects as data, which `internal/questspec` already checks. Do not add a planner.
- **Hierarchical state machines, statecharts** [S general knowledge]. Pattern: named states with entry and exit actions, a deadline per state, transitions in a table. Fixes: "every mode has an exit and a timeout" becomes a table a test can read. Fits the trip (portal home, chores, settle, portal back) and the death recovery, which are true sequences. Cost: M for the trip machine.
- **Runtime assurance (Simplex architecture), subsumption** [S general knowledge, robotics]. Pattern: a small, independently checked safety controller sits beside the complex controller and can always take the actuator; the complex one is never trusted to yield. This is the exact property 3.1 lacks. Cost: P3.
- **Script frameworks for RuneScape-like clients** [U, the documentation page did not load]. Pattern as I know it: task or tree nodes with `accept` and `execute`, and a house idiom of "sleep until condition or timeout" instead of fixed sleeps. The idiom is the postcondition wait of P9.
- **Diablo III, Diablo IV, Grim Dawn, Last Epoch tooling** [U]. I found nothing public that I could verify beyond what the survey lists; no claim made.

### 6.2 Ranked shortlist

| # | pattern | from | kills | lands in |
|---|---|---|---|---|
| 1 | a refusal is a prerequisite that gets queued, not an error | mod-playerbots; behaviour-tree statuses | B, K | P2 |
| 2 | independent safety controller that never waits for the main logic | Botty, Koolo, Simplex | the `bug:true` deaths | P3 |
| 3 | priority processes with TEMPORARY takeover and a lost-control reset | Baritone | A, E | P6 |
| 4 | one town pass, one belt need | kolbot | C | P5 |
| 5 | readiness as predicates in the run table | SoloPlay | G, H | P4 |
| 6 | interact until the object's state changes; identity by id | kolbot, Koolo | H, F | P9 |
| 7 | a shared condition grammar for thresholds, readable by the operator | OpenKore; our own loot rules | G, controllability | P10 |
| 8 | debug strategy: every executed action with its score, and a rolling history | mod-playerbots | K, controllability | P7 |

## 7. Proposals

### 7.1 Ranked list

| # | what | kills | size | risk | deletes | touches |
|---|---|---|---|---|---|---|
| **P1** | **Invariant monitor.** A table of about twelve checks run once per tick on the snapshot and loop state; a hit logs one line `invariant` with a name and facts and counts in the minute line. Log only. First rows: resupplying outside town over 10 s; field entry under the resume line; safety verdict not carried out within 2 s; same step refused 10 times with no fact changed; trips per hour over N; gold under kit cost for 2 h; life zero while alive | K; finds A, B, C, E live and in lab | S | none (no behaviour) | nothing now; later the five loop guards become rows that detect and no longer act | `docs/contracts/soak-gates.md` gains invariant counts as a later step, with the owner's go |
| **P2** | **Typed step outcome.** A step ends as Done, Progress, Refused{need, wake}, Failed{fault} or Interrupted{cause}. `Refused` names what is missing (kit lines, life, level, gold, an item) and the fact that ends it. One handler: a refusal never increments a failure counter, is logged once per change, and becomes the hero's next concern (a trip, a farm, a level target) or a park with a wake condition | B whole; much of A and K; the ground for P4, P6, P7 | M | low to medium: touches `do`, `boardSlice`, `taskStep`; behaviour stays the same where the refusal already led to a trip | `planningFault`'s word match, the `errAreaGated` special cases, the wording match in `internal/watch/hub.go:652`, and over time most `*Until` backoff fields | `docs/contracts/bot-loop.md`; ADR 0025 ("guard refusals fixed by town trigger a restock") is the same idea for one case |
| **P3** | **Reflex lane.** (a) Measure safety latency. (b) T6 cancellation test. (c) A drink is sent by the safety check itself through the input controller under its lock; it does not cancel the action. Portal and chicken keep cancelling, with a hard deadline after which the reflex acts anyway. (d) Town chores verify "in town" before each NPC step and abort the trip otherwise | the `bug:true` class; 3.1 rows 3 and 4 | M | medium: it changes the rule "the guard only watches" (`docs/bot.md`) for one action. On lockstep the guard already runs on the acting goroutine (`loop.go:1008`), so determinism holds [U until run] | the drink branch of the guard's cancel list; `safetyDue` storm limiter | `docs/bot.md` invariant, `docs/contracts/combat-loop.md`; fits the checkpoint design `docs/design/danger-protocol-2026-11-25.md` |
| **P4** | **One admission function** `MayEnter(area, purpose)` and `MayLeaveTown(purpose)` with a `Refused{need}` verdict; the eleven judges become providers behind it; called at the travel chokepoint and at the town exit. Floors are rows keyed by area and archetype (as `plandata/earlyroutes.json` already is), no function per place | G, C at the edges, the "left town hurt or without kit" rows | M to L | medium; do it provider by provider with the shadow-then-flip habit | `barracksGuardWhy`, `towerGuardWhy`, `denFloor`, `andarielLow`, `kitSafeTier` and `dungeonMouths` literals, `routeAllowed` duplicates | ADR 0011 (readiness is power against danger: this is its one entry point), build-flexibility rule (rows by archetype, no class literal) |
| **P5** | **Finish and flip `town.plan`.** Stage 2 on both sandboxes now, default after a week of monitor counts; then stages 3 to 5; then delete the old counters | C (the largest class) | M | medium; mitigated by P1 counts before and after | the old belt, kit and shop counters; several `resource.*` knobs | ADR 0028 as decided; no change |
| **P6** | **Arbiter: change the migration rule, then migrate.** Exit rule becomes outcome, not agreement (6.3). Add TEMPORARY intents and a lost-control hook. Sources read P2 refusals instead of re-deriving facts. Order unchanged: survive and player, chores and claim, gold farm and recovery, quest, level | A, E | L | medium to high; it is the core. The shadow has run without a panic (`panics=0` in every minute line read) | the five loop guards, the chores cooldown family, `forceTrip`, `claimSkipUntil` and kin | ADR 0020 item 3 as decided; `docs/arbiter.md` migration table (a doc, not an ADR) |
| **P7** | **Explain.** Decision ring, `/api/explain`, operator `status --why`, later a "Why not?" row per chunk | K; controllability | S | none | the need to grep logs | ADR 0030: the allowlist gains a read verb. This is a proposed change to a decision: say yes or no |
| **P8** | **Incident to scenario** (T2) plus the report-only tier (T7) | the lab and live gap | M | low | hand-built stand-in heroes over time | ADR 0020 item 1, ADR 0027; test-character provenance: an incident `pre` is a recorded state, tagged as SETUP |
| **P9** | **Generic step executor with postconditions** for interact, pick up, talk, use: approach, act, wait for the declared fact, bounded retries, progress is not failure. New chunks may only use it | H, F | L | medium; per chunk behind the scenario of that chunk | `cainNodes`-style handlers one by one | `internal/questspec`, ADR 0021, ADR 0022 |
| **P10** | **Threshold governance.** A floor or line changes only with (a) the margin model's number or (b) a seed set, never one death; rows carry `evidence: n`; a fix card may not add a knob without removing one or the owner's word. Later: safety and supply thresholds as rows in the loot-rule grammar | G; the knob count | S process, M for the table | low | knobs | ADR 0011; `docs/knobs.md` |
| **P11** | **Operator verbs** hold, home, go, park-quest-for-N-hours on the socket, mapped onto the existing orders (`go`, `fight`, `home`) | controllability | S | low | nothing | ADR 0030 allowlist: a proposed change to a decision |
| **P12** | **Fault-injection soak** (T5) | E, F, recovery paths | M | low (lab only) | nothing | ADR 0014: DIAGNOSIS writes in the lab, never measured |

### 7.2 The next five PRs

1. **invmon** (P1): the monitor table, log lines, minute-line counters. Two builds in its test (default Sorceress and Paladin facts). No behaviour change.
2. **outcome1** (P2, first slice): the outcome type; `boardSlice` and `taskStep` return it; refusals log as INFO `step refused` with `need=` and `wake=`, once per change, and never reach `fail`. Scenario: the early-road refusal of 2026-11-26 produces one line and one kit trip, not 226 warnings.
3. **reflex1** (P3 a, b, d): safety latency in the autopsy and the minute line; the cancellation-latency test; the in-town check before each chore step. Scenario from the 02:51 death (trip in chores stage, hero placed in a field with hostiles).
4. **incident1** (P8): the bot writes a scenario `pre` at each death and monitor hit; the first three scenarios are the Questy deaths; the report-only tier is the ops-repo half.
5. **explain1** (P7): the ring, `/api/explain`, `ops quest status --why` (after the owner's yes on the ADR 0030 line).

Beside them, not a PR: set `town.plan` 2 on both sandboxes and read the monitor counts for three days (P5).

### 7.3 The arbiter's migration rule cannot be met as written [V `docs/arbiter.md`, log]

The table moves a family when "divergence is under 2 (or 5) percent of ticks". The live divergences are mostly cases where the arbiter is right and the chain is wrong: "chain chose quest, arbiter would choose resupply/refill (belt N under N)" before deaths with an empty belt; "chain chose chores, arbiter would choose survive/drink (life N% under N% with hostiles near)". Agreement with a faulty chain is the wrong bar: the better the arbiter, the further it is from passing. Replace it by an outcome rule: on recorded sessions (T4) replayed with `-follow`, and on the incident scenarios (P8), the arbiter-led run has fewer monitor hits (P1) than the chain-led run and no new ones; then the family moves behind its knob on one sandbox. Second point: the sources re-derive each hook's predicate from "pure facts" because the hooks' own predicates have side effects (`docs/arbiter.md`, "The adapter"). That is a second implementation of every rule. With P2 a source reads the standing refusals and needs, and the duplication goes.

### 7.4 What NOT to do

- **No rewrite of `internal/bot`.** The parts that hurt are the outcome contract, the admission judges and the arbitration; each has a seam.
- **No sixth guard**, and no new "X stops counting after N" rule. Each one is a class B or A card in waiting.
- **No behaviour-tree, GOAP or utility library.** Two ideas from them (statuses, decorators) are worth a few hundred lines of our own.
- **No floor moved after a single death.** 2.3 is what that costs.
- **No new knob as the default answer to a fix.** 58 rows in six days.
- **No further structural work parked at default 0 without a flip date and a measure.** Three parked replacements are the main finding of this review.
- **Do not wait for the arbiter to fix supply or safety.** P3 and P5 are independent of it.
- **Do not tune the nav class (D) structurally.** It is the one large class that is mostly the game's own complexity; keep fixing it case by case, with a fixture each time, as the toolstrade cards did.

### 7.5 Decisions this review proposes to change (explicit)

| decision | proposed change |
|---|---|
| ADR 0030 allowlist | add read verb `why` (P7) and the order verbs hold, home, go, park (P11); same guard rails |
| `docs/bot.md` invariant "the guard only watches" | a drink may be sent by the safety check itself (P3 c) |
| `docs/arbiter.md` migration exit rule | outcome on replay and incident scenarios instead of agreement (7.3) |
| ADR 0020 item 1 as practised | make the scenario tier real, report-only first (P8; the pipeline audit already asks for it) |

Not proposed for change: ADR 0011 (P4 is its implementation), ADR 0014 (the injections of P12 are lab DIAGNOSIS writes, never in a measured run), ADR 0028, ADR 0029, the build-flexibility rule (every new row is keyed by archetype or build data).

## 8. Limits of this review

- The card classification is by title keywords, hand-checked, not a root-cause read of 178 diffs [U].
- The farm sandbox log held only 58 minutes; nothing is concluded from it beyond the table in 3.4.
- The root cause of the 02:51 death (why the hero stood in Cold Plains during the chores stage, and why `life` read 0 from the town on) is not established [U]. The lines are quoted as observed.
- I did not run the Proving Ground, any test or any game. No whole-repo command was run.
- Prior art marked [U] is from memory; the D2 bot facts marked [S] rest on the 2026-11-19 survey's reads, not on a re-read today.
- Sizes are estimates for Sonnet agents with a scenario each; P4, P6 and P9 will split into several cards.
