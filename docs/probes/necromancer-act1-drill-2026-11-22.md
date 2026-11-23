# Necromancer earned Act 1 drill (card necrodrill, 2026-11-22)

> status: current. Verdict: BLOCKED, NOT ready to play. Tags: V verified in a named run, S read in a log, U unverified. Builder LONG lane, lab d2d devp20 (`--read-combat`), `combat.minionRead` 1, build `necromancer_summoner_sme`.

## 1. Runs

| run | what | verdict |
|---|---|---|
| 20261122-172117-b149 (V) | `necro.act1-earned`: a fresh clvl 1 Necromancer, town start, goal `campaign:normal_act1`, NO setup write, `writes after start: 0` | BLOCKED: quest record 6 (Andariel) not observed in 400 virtual minutes. The job was cut at 3604 s wall by the 60-minute hard limit of a detached job (the scenario's own 140-minute wall budget never applied); the result file was written at 3586 s |
| 20261122-182247-6965 (V) | stage 1 as the existing act1 scenarios with `-class Necromancer -seed 11` (CHEATED level fixtures, default table `necromancer_summoner_normal30`): `act1.den.clear` and `act1.burial.blood_raven` | `act1.den.clear` RED (record 1 read 0x0014, wanted bit 0 set; Start was tapped only 4 wall minutes before the end); `act1.burial.blood_raven` BLOCKED (bot run cut before its virtual budget; the hero ended at clvl 16 in the Rogue Encampment) |

The ledger refusal ("missing bot_commit") is a launch detail: `run.sh` reads `git rev-parse` in the synced mirror, which has no `.git`; passing `env D2LOD_BOT_COMMIT=<hash>` on the labrun command line fixes it (V: no refusal in run 6965).

## 2. How far the fresh hero got (V from the bot log and result file of run b149)

- clvl 14, 5381 kills, 77 boss kills, 0 deaths, 168 potions, 13 level-ups in 24008 active seconds (11448 xp/h). Level by wall-time hour: 1, 6, 9, 12, 13, 13, 14.
- Quests: Den of Evil done (`Quest done: den_of_evil` after about five virtual minutes; record 1 ended 0x2001). Nothing else: no Blood Raven, no Cain, no Countess, no Andariel.
- Where the time went: 88 "Quest: fighting (Level up to N)" states (N = 12 to 15), 45 "Recovery: farming gold", 102 town trips (heal below the resume line 39, healing potions 25), 327 area changes (Rogue Encampment to Cold Plains and back about 60 times), stash 29341 gold with 423 in the purse, skill points all spent (15 hard points, `skill_levels` Raise Skeleton 5, Skeleton Mastery 5, Clay Golem 1).
- What stopped it: the board said "Training toward Sisters to the Slaughter: N levels to go" (10 lines) and the bot kept levelling; 38 "quest board step" actions failed, 37 with `exit to Cold Plains: nav: click {149 330}: no path: the click would cross the trigger area of an unwanted exit`, three failsafe-ladder escalations, and one `back to field ... portal led to another level`. The campaign goal without the board-taps driver starts no chunk (autoStart is never), so only the Den (a goal of its own) completed.

## 3. Blockers, ranked

1. Nav: the Rogue Encampment exit to Cold Plains click is refused as crossing an unwanted exit trigger (37 failures in one run). It sits on every Act 1 quest. Not Necromancer-specific (U: the Paladin runs used the same exit; the dev logs of act1more show the same family for the Outer Cloister).
2. Quest start: the drill cannot start the chunks after the Den. The `board-taps` driver taps one quest per scenario; a whole-act run needs a driver that taps every offered chunk in order (the Cain scenario taps only its own). Until then an earned multi-quest drill cannot be built from the scenario format.
3. Pace: clvl 14 after 400 virtual minutes (11.4k xp/h) with 102 town trips; the "Level up to N" training loop is the time sink, and the hero at that level is below the board's readiness line for Blood Raven (U: the board's reason is not logged, same finding as act1more finding 3). The army kills (5381 kills, 0 deaths), so this is throughput, not survival.
4. Tooling: a detached labrun job has a 60-minute hard limit regardless of `budget.wall_min`; a 400-virtual-minute drill needs stages, and `act1.burial.blood_raven` for a Necromancer hit a bot-run cut before its budget (cause U, the log tail shows a normal stop).
5. Den for the Necromancer under the act1 scenario (CHEATED level 10, default table) did not complete in 45 virtual minutes although the fresh hero finished it in five (U why: different build table and fixture level; read the Den log of run 6965 before drawing a conclusion).

## 4. Next cards

- `nav: Rogue Encampment to Cold Plains exit click refused as an unwanted exit trigger` (owner nav; blocker 1).
- `scenario driver: tap every offered chunk of a campaign in order (board-taps all)` (owner scenario/provingground; blocker 2), then stage the drill as three jobs each under 60 minutes: Rogue Encampment to Cain (record 4), to Countess, to Andariel, restoring the saved hero between stages.
- `quest board: log why a chunk is not startable at the hero's level` (blocker 3, also act1more finding 3).
- `scenario: necro stage runs with the SME build and the Necromancer matrix on act1.*` (blocker 5).
- Test character: none created by this drill (the builder runner creates and deletes its own heroes; the provenance is the scenario's `provenance` text). A hidden-instance Necromancer in docs/probes/test-characters.md waits for the 23:00-07:00 EST window.

## 5. Follow-up runs: the exit refusal, the knobs, and why the Necromancer runs are cut (2026-11-22, evening)

### 5.1 Is the Cold Plains exit refusal fixed by `nav.warpRead` and `nav.fullGrid`? (V for the counts, U for the conclusion)
| run | scenario, seed | knobs | verdict | `unwanted exit` lines, `mouth gate` lines |
|---|---|---|---|---|
| 20261122-190413-c261 | `act1.burial.blood_raven`, Paladin 11 and 3 | warpRead 1, fullGrid 1 | GREEN | 0, 0 |
| 20261122-190413-c261 | same, Necromancer 11 and 3 | warpRead 1, fullGrid 1 | BLOCKED (cut, section 5.2) | 0, 0 |
| 20261122-195405-6208 | same, Necromancer 11 | OFF (defaults) | BLOCKED (cut, section 5.2) | 0, 0 |
| 20261122-195405-6208 | `act1.den.clear`, Necromancer 11 | warpRead 1, fullGrid 1 | RED, record 1 0x0014 as before | 0, 0 |

The refusal does not appear in these scenarios with the knobs OFF either, so they cannot show that the knobs remove it. The 37 refusals of run b149 came from a fresh hero whose board step walked from the Blood Moor to Cold Plains after hours of play, knobs OFF; no staged scenario reproduces that walk. Evidence for a default flip of `nav.warpRead` and `nav.fullGrid`: none from this card (mapverify #1108, 55 to 0 mouth-gate refusals on the Countess, remains the only measurement). A direct test needs a stage that starts the hero in the Blood Moor and sends it to Cold Plains with the knobs OFF first (to reproduce) and ON second.

### 5.2 Why the Necromancer `act1.burial.blood_raven` runs are cut (V, bot log of 20261122-190434 seed 11)
The hero is not slow, it is stuck from the first minute. clvl 16 fixture, 8000 gold, 6 mana potions in the pack and none in the belt:
1. the town bank chore deposits 6000 of 8000 gold (`byChore stash -6000`, `town.bankFirst` default; the act1more scenarios set it to 0, the Necromancer run used the same file and was cut anyway);
2. the kit gate asks for mana potions 0/3 and the shop plan says `wantMP 0` because the pack holds 6 (`packMP 6, beltMP 0`);
3. the quest step then refuses with `not ready: 0 mana potions < 2 (caster)` (internal/task/gate.go:59, World.ManaPotions = `in.Belt.MP`, internal/bot/task.go:357, belt only), forever: `needs_attention` from 19:04:47 to the stop at 19:33:40, 16 kills in 29 wall minutes.
The guard counts belt plus pack for mana only at `town.plan` stage 2 or `quests.boardSM` 1 (internal/bot/routeguard.go:368, 395), both default 0 and absent from the scenario knobs, while healing potions already count the pack (task.go:356). A Paladin never meets it (no mana need). So the cut is a bot defect, not the 60-minute job limit and not a Necromancer pace problem: the same family as act1more finding 2 (belt against pack).

### 5.3 Cards
- `quest gate: count pack mana potions like healing ones` (owner task/bot): fix `World.ManaPotions` at stage 0 and 1, or make `quests.boardSM` the default; must-fail test on Necromancer and Sorceress with mana in the pack and none in the belt. Then rerun `act1.burial.blood_raven` for the Necromancer.
- `scenario stage: Blood Moor to Cold Plains walk, knobs OFF then ON` to settle the default flip.
- `act1.den.clear` for the Necromancer stays RED with and without the knobs (record 1 0x0014 against 0x0001); the Den log of run 6208 (V, read): 409 kills, 4 boss kills, 0 deaths, 33 potions in 45 virtual minutes, but the hero went back to town again and again for `low on mana potions` and `mp belt columns below 50%` (Cave, area 9, then area 13 and Cold Plains), so mana potion supply, not survival, set the pace and the Den bit was not set in time.

## 6. Stage results after the fixes (2026-11-22, late evening) and the revised blocker list

| stage run | verdict | note |
|---|---|---|
| 20261122-205642-49a0 `act1.burial.blood_raven`, Necromancer 11 | GREEN (60 virtual minutes) | after the pack-mana fix (manapack, #1135); before it: stuck 29 minutes on "0 mana potions < 2" |
| 20261122-205642-49a0 `act1.den.clear`, Necromancer 11 | RED, then GREEN | RED cause: the bank chore moved 6000 of 8000 gold into the stash, the hero left the Den after 36 s to farm gold (denfix, #1140); with `town.bankFirst 0` GREEN for the Necromancer and the Paladin (job 20261122-212018-1980). Record 1 0x0014 = 0x4 Akara talked, 0x10 U, not the completed bit 0 |
| 20261122-213319-0b57 `nav.bloodmoor-coldplains`, Paladin and Necromancer, seeds 11 and 3, knobs OFF and ON | GREEN 8 of 8 | 0 `unwanted exit`, 0 `mouth gate` lines in all eight logs, Cold Plains entered in each: the Blood Moor to Cold Plains walk does not reproduce the 37 refusals of the fresh drill with or without `nav.warpRead` and `nav.fullGrid`. So there is still no default-flip evidence for the two knobs from this card |

Revised blockers for the earned Act 1 drill (ranked):
1. Whole-act run: the tap driver now exists (`trigger.tap_quest` list or `*`, tapall #1130) but no multi-quest stage scenario uses it yet; stages must restart from declared level fixtures (the runner cannot restore a save) and each job must stay under the 60-minute limit of a detached job.
2. Bank chore against the quest kit: with the default `town.bankFirst` the purse is emptied before the kit is bought (card in progress; a Necromancer needs many mana potions, so it shows first).
3. Fresh-hero pace: clvl 14 in 400 virtual minutes (section 2); not re-measured after the fixes.
4. The fresh-drill exit refusal (37 lines, knobs OFF) is not reproduced by any staged scenario: U which walk and which position produced it.
Is the drill feasible in stages now? Yes for single quests from fixtures: Den and Blood Raven are GREEN for the Necromancer on the lab d2d. Not yet for the chain Rogue Encampment to Andariel; the next step is stage scenarios (`tap_quest: "*"`, a level fixture and a 55-minute budget each) for Cain, the Countess and Andariel. Still NOT ready to play.

## 7. Bank chore against the kit: does `town.plan` stage 3 cover it? (card bankdoc)

Question: with `town.bankFirst` 1 (the default) the bank chore moved 6000 of 8000 gold to the stash, the Necromancer was short in purse and kit and left the Den to farm gold. Is that fixed by `town.plan` 3 (purse: one stash stop settles the purse to the carry line plus the planned spend; wealth for the recovery and poverty rules is purse plus stash, `wealthFor` in internal/bot/broke.go)?

| run (builder long lane, lab d2d, `act1.den.clear` seed 11, CHEATED level 10 with 8000 gold) | knobs | verdict |
|---|---|---|
| 20261122-205642-49a0 and 195405-6208, Necromancer | `town.bankFirst` 1, `town.plan` 0 | RED (record 1 0x0014; Den left after 36 s, 409 kills farming gold) |
| 20261122-213936-7ae3, Sorceress | `bankFirst` 1, plan 0 | GREEN |
| 20261122-212018-1980, Necromancer and Paladin | `bankFirst` 0 | GREEN, GREEN |
| 20261122-213936-7ae3, Necromancer | `bankFirst` 1, plan 3 | GREEN |
| 20261122-213936-7ae3, Sorceress | `bankFirst` 1, plan 3 | GREEN |

Reading (V for the verdicts; U for why the Sorceress is spared): at stage 0 the defect shows for the class that needs many mana potions (the Necromancer's kit gate asks 3 mana potions and its fights drain them), not for the Sorceress or the Paladin in the same scenario; stage 3 removes it for the Necromancer without touching the Sorceress. Stage 3 therefore already covers the defect when it is on, and no withdraw-when-short rule is written here.

Proposal (NOT applied, defaults stay): set `town.plan` 3 as the default stage for the friends sandbox (it runs stage 1 today) and for the scenarios of Necromancer builds. Before flipping, a second seed and a Paladin run at stage 3 (`bankFirst` 1) should be green; one seed per class was run. The scenario fix of denfix (`town.bankFirst` 0 in `act1.den.clear`) can then be dropped.

## 8. Staged drill, night run (2026-11-23, 00:20 to 01:30 EST; builder long lane, lab d2d devp20, CHEATED fixtures, mechanics only)

| stage | scenario | verdict | evidence |
|---|---|---|---|
| 1 | `necro.act1-stage1` (new): level 14 fixture, 8000 gold, tap driver `den_of_evil,sisters_burial_grounds`, build `necromancer_summoner_sme`, Necromancer 11 (job 20261123-002147-61e0) | BLOCKED | the driver started the Den at 00:22:13, the Den was done and handed in at 00:22:27 (14 s, V), Blood Raven started at 00:23:11; then `quest.sisters_burial/to-burial` failed: the hero had walked Den, Stony Field, Underground Passage, Dark Wood, Black Marsh, Tamoe Highland, Monastery Gate, Outer Cloister and ended in the Barracks (areas 26 to 28), `exit to Jail Level 1: nav: no path`, `exit to Outer Cloister: nav: no exit to that level: level 27 from 28`, step blocked at 00:23:14. After that the log holds only `d2dtrace` lines: the d2d tick counter froze at 21189 (`virtual` 848 s) for 49 minutes until the 50-minute wall limit cut the run |
| 2 | `act1.forgotten.countess` with `-class Necromancer -seed 11` (existing scenario, level 24 fixture, job 20261123-011245-cb7e) | GREEN (60 virtual minutes, 494 s wall) | record 5 bit 0 set |
| (Tools of the Trade) | not run | the live board still reads `quest.tools_of_the_trade/kill: boss not found in the arena` at 23:49 EST (read-only docker log): the Tools fix was not deployed when this was written |

Findings (blocker list, ranked):
1. Blood Raven from the Den: the route step `to-burial` sends the Necromancer through the Monastery road into the Barracks, where it finds no exit and stops. The earlier GREEN Blood Raven run (level 16 fixture, no Den first) did not do this, so the difference is the start (level 14, just after the Den) or the level plan; U which. Not Necromancer-specific in code (U: a Paladin stage 1 was not run).
2. After the step fails, the bot sits idle and the game clock stops (49 minutes with no tick): a failed chunk should park and return the hero to its level plan, not freeze. The `needs_attention` state does not step the game.
3. The Tools of the Trade fix is not live (`boss not found in the arena`).
4. The chain Rogue Encampment to Andariel still has no whole run: stages restart from fixtures and each job stays under 60 minutes; stage 1 and the Tools chunk are blocked, the Countess stage is green for the Necromancer.
Not ready to play.

## 9. Thresholds of the necro scenarios from lab traces (card drill2)
From run 20261122-170602 (one run per scenario, CHEATED, so a floor rather than a tuned number): `necro.upkeep-pack` 71 kills, 5 Raise Skeleton raises that grew the army, 2 Raise Skeletal Mage, 1 Clay Golem, 15 Amplify Damage curse lines; `necro.curse-unique` 75 kills, 20 curse lines, 5 skeleton raises; `necro.revive-after-death` 87 kills, 1 Revive that grew the army, 6 skeleton raises; `necro.mana-budget-10min` 94 kills, 5 skeleton raises, 18 curse lines. The scenario floors are now: upkeep Raise Skeleton raises >= 3 (was 2), curse lines >= 8 (was 1), mana budget kills >= 50 (was 40); Revive stays at 1 raise. All four GREEN again with the new floors (job 20261123-012125-173a). One run per scenario: the variance is U; the floors sit at about half of the observed counts.

## 10. Reruns after failspend (#1170) and the two blocker cards (night of 2026-11-23, builder long lane, lab d2d devp20, sandbox nav knobs `nav.warpRead` 1 and `nav.fullGrid` 1, CHEATED fixtures)

### The 49-minute clock freeze (V, bot log of run 20261123-002147-61e0 and code reading)
Not a lab artifact of the runner and not a hang: a spin. The board step `quest.sisters_burial/to-burial` was refused with a `ui` error (`exit to Outer Cloister: nav: no exit to that level`), which ran nothing. `boardRun` returned true without any clock sleep, so under the lockstep clock no game time passed, the 15-second failure gap (`boardFailGap`) never elapsed, the three counted failures that park a chunk never came, and the loop polled snapshots about 130 times a second while the d2d tick counter stood still. In real play the wall clock moves, so the chunk would park after three tries; in the lab it burned the wall budget. Fix merged as failspend (#1170): a failed non-idle board step spends a third of the gap (`failSpend`, unit test). V: the rerun (job 20261123-023115-028d) parks the chunk within 4 virtual seconds (`parked ... gave up after 3 tries`) and then runs its full 90 virtual minutes in 1272 s.

### Stage 1 reruns
| run | fixture | verdict | what it shows |
|---|---|---|---|
| 20261123-023115-028d | level 14 | RED (a death in the Outer Cloister, Rust Shard) | after the Den hand-in the level plan (window 14-18, Catacombs 3) walks the hero Dark Wood, Black Marsh, Tamoe Highland, Monastery Gate, Outer Cloister, Barracks while the Blood Raven chunk waits in its start gate (`start gate: waiting for the fight to end`); the chunk then starts in the Barracks and `to-burial` cannot route back (`nav no usable gate from 28 to 27`, `border not readable`). With the fix it parks and the hero goes on into the Monastery and dies at clvl 14 |
| 20261123-025322-bafd | level 16 (the fixture of the green Blood Raven runs) | RED: record 1 read 0x0004 | the Den chunk reaches `awaiting_handin` 11 s after Start (00:00:02 to 00:00:13) and the hand-in fails 3 times: `no skill point after talking to Akara (quest not complete, or the dialog...)`; Blood Raven stays `ready` and is never started in 90 virtual minutes |

So stage 1 is still NOT green. Two concrete defects:
1. The Den chunk of a level 16 Necromancer is judged done after 11 s (no kill reads as complete) and the hand-in is refused by the game: the board's completion test for `clear_den` reads a state the fixture sets (U: quest word 0x0004 from the SETUP, the save, not the kill). The denfix runs (level 10 fixture) clear the Den for real and are GREEN; the difference is the fixture level (U).
2. The level plan and a queued quest start run at the same time (a chunk waiting 30 s for the fight to end while the plan travels toward the Monastery), and the nav cannot route back from the Barracks to the Outer Cloister.

### Stage 2 and 3
Stage 2 (`act1.forgotten.countess`, Necromancer 11): GREEN (job 20261123-011245-cb7e, 494 s). Stage 3 (Andariel) and the Tools of the Trade stage were not run: the Tools fix was not live at 23:49 EST, and Andariel needs a geared hero (act1more finding 3).

### Verdict
The earned Act 1 drill is NOT green for the Necromancer: the Den to Blood Raven stage fails on a Den completion test and a plan/quest overlap, the Countess stage passes, Andariel is unrun. Not ready to play. Cards: `quest board: clear_den completion test at a fixture level` (owner questboard), `level plan: do not travel while a chunk waits in its start gate` (owner bot), `nav: route back from the Barracks to the Outer Cloister` (owner nav, toolstrade family).

## 11. The three blocker cards (night of 2026-11-23, card stage1fix; builder long lane, lab d2d devp20, sandbox nav knobs, CHEATED fixtures)

| blocker | what was found | fix | evidence |
|---|---|---|---|
| nav from the Barracks back to the Outer Cloister | V (run 20261123-023115-028d log): arriving from the Outer Cloister the hero is 60 subtiles from the south strip; it walks to the strip cell twice, and there `nav no usable gate from 28 to 27 (exits 1, adjacent 0, borderCells 0, unproven 4, pending 1, fullyKnown)`: d2d never reads across the edge from the strip, so no border cell shows and ExitTo ends in ErrNoExit | `exit.go`: after the strip was approached and still no gate shows, the unproven strip cells are the gate (`gatesFromCells`, log line `nav: crossing an unproven border strip`): walking over the strip changes the level | unit test `TestExitToCrossesAnUnprovenStripThatNeverProves` (Paladin, Necromancer; fails without the fix with the live error `nav: no exit to that level`). NOT shown in a scenario: `nav.barracks-return` (field start by `debug goto 28`, hero at (65,212)) is RED for both classes with and without the fix because the planner finds no plan to the strip from there; it is marked pending with that reason |
| level plan against a queued quest start | V (same run): the chunk was accepted in the Stony Field and started 30 s later in the Barracks: a goal travel is one blocking action over many areas, so the start gate (`boardQueueGrace` 20 s) was only evaluated after the plan had walked the hero Dark Wood, Black Marsh, Tamoe Highland, Monastery Gate, Outer Cloister, Barracks | `goal.go`/`questboard.go`: the goal travel yields once per chunk at the next hop when a started chunk is runnable (`boardYield`; quest steps do not yield; a chunk that cannot start never starves the travel) | unit test `TestBoardYieldOncePerChunk` (Necromancer, Paladin). Not shown in the scenario (below) |
| Den hand-in at a fixture level | V: the Den was cleared in game (live memory word `den_of_evil=0x2016` at 03:58:01, 62 kills, Corpsefire dead) in 5 s wall; the tap driver then tapped Hand in and, in the first reruns, Start on Blood Raven one second later, which cancelled the hand-in dialogue (`quest board: waiting: context canceled`). Record 1 bit 0 is the reward taken (0x0001), 0x0004 the dialogue with Akara, 0x2000 the cleared state, so the scenario assertion (bit 0) is right | `tap.go`: the driver does not tap Start while the chunk of its last Hand in still shows the hand-in button or awaiting_handin | unit test `TestTapperWaitsForTheHandInBeforeTheNextStart`. The scenario is NOT green yet (below) |

### Stage 1 after all three (job 20261123-035544-7536): still RED
`necro.act1-stage1` (level 16 fixture, tap list `den_of_evil,sisters_burial_grounds`): the Start tap on Blood Raven no longer interrupts the hand-in, yet record 1 still reads 0x0004 at the end. The log shows the Den cleared (0x2016) and then a mana-potion town trip (`kit: mana potions 0/3`) while the Hand in tap lands; the Akara talk completes (`talk done`, 10 s virtual) with `skillPointsBefore 0, after 0`, so no reward is paid, the board parks the chunk after three tries, and the hero's `back to field` fails (`portal click did not change the level`, then `d2d.sock` gone). The same Den chunk alone is GREEN at clvl 10, 14 and 16 (`act1.den.clear`, `necro.den-l14`, `necro.den-l16`, jobs 20261122-212018-1980 and 20261123-033710-afbd), so the difference is the combination with the stage-1 setup (items, waypoint 1, `combat.minionRead` 1, the second quest in the tap list). U which of these; next card: `stage 1: find what the Den hand-in needs that den.clear provides` (read the Akara dialog and the skill-point read-back under the stage-1 fixture).

### Verdict
The earned Act 1 drill is still NOT green for the Necromancer: Den to Blood Raven fails at the Den hand-in in the stage-1 setup (cause open), the Countess stage is GREEN, Andariel is unrun. Not ready to play. What is fixed (offline tests, not yet shown in a scenario): the return leg from the Barracks, the plan overlap with a queued chunk, the driver's Start-during-hand-in.
