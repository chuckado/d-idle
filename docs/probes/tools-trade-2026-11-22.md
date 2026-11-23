# Tools of the Trade stuck at a border exit (card toolstrade, 2026-11-22)

> status: current. Tags: V = measured in a named run, S = source reading, U = unverified. Setups are CHEATED (scenario `act1.tools.malus`: level 22, gold, waypoints, potions): mechanics only.

## Report

Sandbox (release a756ec2, nav.warpRead=1, nav.fullGrid=1): `quest.tools_of_the_trade/to-barracks` failed three times with `nav: no path: reached the closest reachable point {141 5} to {141 0}` and the chunk parked `gave_up`. The goal is a border-row cell (y 0) of a cloister level.

## Root cause (V, builder long lane, scenario `act1.tools.malus` Paladin 11, bot logs)

1. Both knob settings (sandbox knobs on, and both off) fail the same way before the fix: at the Monastery Gate the first `ExitTo` 27 picks the gate at {141 0}, the plan looks complete, the walk ends 5 subtiles short ("reached the closest reachable point {142 5} to {141 0}") and `ExitTo` returns that error. `openGateDoor` (the door at (141,2) in front of the border cells, mapverify) only ran when no gate plan existed at the start of an attempt, never after a walk that ended short. The board step then retried; the retry opened the door. With three of those in a row (the doors are met again after every trip to town) the step parks.
2. The same shape applies at Outer Cloister to Barracks (the report names the Barracks; U: the Barracks strip has a door of the same class, not reproduced because the scenario hero is chased at arrival, see below).
3. Not a data hole and not a knob effect: `nav.fullGrid` / `nav.warpRead` change nothing here (V, identical logs).

## Fix

`internal/nav/exit.go`: when the walk to the gate ends in `ErrNoPath`, `doorAfterShortWalk` (`internal/nav/gatedoor.go`) observes again and runs `openGateDoor` at the new position, then `ExitTo` loops. Bounded by the existing door counters (`doorTried`, `doorApproach`); when no door is found the old error is returned. Class-free geometry.
Must-fail test `TestExitToOpensTheDoorAfterAShortWalk` (two sub-tests; the world is class-free, so the class label is only a name): with the call removed it fails with the live error text `no path: reached the closest reachable point`. `go test -race -count=20 -run TestExitTo ./internal/nav` passes.

## Scenario after the fix (V, sandbox knobs, run k3-202347 fixgw)

First attempt now opens the door (141,2) from (142,4) and `Entered Outer Cloister` (before: only on the retry). The run then stops on a different problem: arriving in Outer Cloister at (141,199) six hostiles stand around the hero (life 144 of 253, mana 5) and it chickens; a later re-entry from town ends in Tamoe Highland ("left level 26 for 27 but ended in 7"). Both are separate findings: combat/safety at the cloister arrival with a CHEATED level-22 kit, and the Monastery Gate border push after a town return. The Barracks leg, the Smith, the Malus and the Charsi reward were NOT reached in any run: full-chunk GREEN is NOT shown.

## Not done

- Full chunk GREEN, Sorceress 3, and the both-knobs-off run after the fix.
- A check of every Act 1 border exit for a reachable approach cell (needs the lab map reads per level; the existing `exitsread` probe reads the link table only). Proposed: extend `tools/probes/exitsread.py` with a `debug blocked` strip scan per border link and assert a cell connected to the interior.
- parkrelease (#1090): the park stores release `a756ec2-dirty`; `smReleaseRetry` (internal/questboard/exec_sm.go) clears a `gave_up` park whose stored release differs from the running one, once (S, by reading; the next bot release differs, so the step retries once).

## Round 2 (card toolstrade2)

- V (run k3-202347 fixgw, bot log): after a town trip the hero arrives at the Monastery Gate at (162,86), 4 subtiles from the south border; the walk to the door at (140,78) was planned via (162,89) and (147,89), onto the strip toward level 7, and the level changed ("left level 26 for 27 but ended in 7"). Level ids are right (27 is the Outer Cloister); not an off-by-one: the planner only penalised tile exits, never border strips.
- Fix: `penalizeBorders` (internal/nav/nav.go) puts a soft cost on the strip cells toward every neighbour other than `wantExit`, except near the goal. Must-fail test `TestWalkKeepsOffTheBorderStripOfAnotherLevel`.
- Scenario hero: `act1.tools.malus` now clvl 19 with a Rogue merc and cap, quilted armor, short sword and buckler (CHEATED, mechanics only).

## Round 3 (card toolstrade3)

- V (run k3-212315 off, bot log): the Monastery Gate door budget (`maxDoorTries` 2 per door per game) was used up by the first two visits; from the third visit on every ExitTo failed with "gate to level 27 is on the map but not reachable (cut off, no way out found)" and the board parked. The level is generated again after each town trip with its doors shut. Fix: `doorTried` and `doorApproach` are reset at the start of each ExitTo call. Must-fail test `TestExitToOpensTheSameDoorOnEveryVisit`.
- Scenario `act1.tools.barracks`: the chunk from a field start in the Outer Cloister (CHEATED, level 26) to reach the Barracks leg without the arrival crowd deciding.
- Observed, not fixed: with the clvl 19 hero and a merc the run keeps going home (merc dies, portal casts cancelled by the crowd rule, `guard/route/chores/home ran but is still needed`) and parks after three tries: a combat and safety finding at the Outer Cloister arrival, not nav.

## Round 4 (card toolstrade4)

- V (run k3-213843 sbx, log): the re-entry failure had a second cause. The arrival pocket of the Monastery Gate has no door in view ("cut-off gate, no door beside it", nearest object 42 subtiles), so `pocketTarget` picked the nearest connected cell with wall clearance: (97,89), on the south strip toward the Tamoe Highland, and the walk there changed the level. Fix: `pocketTarget` and `nearestReachableTo` skip cells on or within 2 subtiles of the strips toward other levels (`otherStrips`). Must-fail test `TestApproachCellKeepsOffOtherLevelsStrip`.
- V (same run, scenario `act1.tools.barracks`, field start in the Outer Cloister, level 26): Outer Cloister to Barracks works: door (141,2) used from (141,4), `Entered Barracks`, then Jail 1 and 2 (the `reach:30` goal walked on). So the Barracks leg is not broken by a missing door.
- V (run k3-215844 sbx act1.tools.malus Paladin 11 after the fix): the re-entry now reaches `Entered Outer Cloister`. The run then parks on `guard/route/chores/home` (the hero wants town, the portal cast is cancelled by the crowd rule, three tries): the arrival crowd of the Outer Cloister against a CHEATED clvl 19 kit. Smith, Malus and the Charsi reward are NOT reached: GREEN is not shown. Sorceress 3 and the knobs-off after-fix runs were not made.

## Round 5 (card toolstrade5)

- V (runs k3-222048, log): at the Outer Cloister the hero fights the arrival crowd, drinks one potion (4 to 3) and `guardedExit` fires: `guard: leaving guarded area 27 first (3 healing potions < 4)`, `crowd:true`. The route guard is the gate verdict (a thin kit closes the road to a hero setting out); `guardedExit` reused it for a hero already on the road. The walk home failed (no waypoint on level 27, the portal cast refused by the crowd rule, the border walk back read as "not readable"), `guard/route/chores/home ran but is still needed`, three tries, parked.
- Fix: `task.RouteLeaveWhy` (same level grounds, healing and mana floors halved) used by `guardedExit` through `areaLeaveWhy`; entering the road is unchanged. Test `TestRouteLeaveWhyToleratesAThinnedKit` (non-caster and caster). U: whether half is the right slack; the knob lives in the code (`kitWhyAt`), not in `docs/knobs.md`.
- Probe `tools/probes/strips.py` (item 6): per Act 1 level the border links (nearby minus warps), the side from the offset (U heuristic), and a verdict open / pocket / closed from the grid band next to the edge and a flood fill; unit test `tools/probes/strips_test.py` (python3 -m unittest strips_test in tools/probes). Lab run not made yet: `python3 tools/probes/strips.py D2D DATA 1,2,3,4,5,6,7,26,27,28,29,30,31,32,33 5`.
- Scenario `act1.tools.malus` hero level 28 (CHEATED: the clvl 19 kit with no stat points kills too slowly for the cloister crowd; the guard loop is then a scenario artefact, the real sandbox hero got to y=5 of the Barracks border).

## Round 6 (card toolstrade6), live evidence of the sandbox 2026-11-23

- V (coordinator, sandbox log 04:00 to 04:10Z, deploy 39): the chunk walks the long way to the Outer Cloister (04:07:05), 33 s later `Entered Rogue Encampment` with `guard/route/chores/home` and `quest board: hold needs_potions: not ready: route guard (3 healing potions < 4)`; the hero had 5 to 6 potions at the cloister and lost 2 to 4 in 30 s; restock, the chunk restarts at Talk to Charsi, the walk repeats: a loop.
- Cause (S, reading): besides `guardedExit` (toolstrade5) the NEXT STEP's `fieldReady` (`internal/task/gate.go`) applies the gate verdict to a hero that is already on the road. Fix: `fieldReady` skips the route block when the hero's area is guarded and `RouteLeaveWhy` is empty. Test `TestFieldReadyOnTheRoadToleratesAThinnedKit` (non-caster and caster; fails without the exception).
- Potion tier and count (S): `internal/town/belt/plan.json` allows tier 4 (greater) from clvl 13; `pickPotion` buys the best affordable tier in stock, so Minor potions in the belt come from the vendor stock or the budget (U: not read from the sandbox shop), not from the cap. Not changed.

## Round 7 (card toolstrade7): the Smith is not found in the lab Barracks

- V (lab runs k3-230123 sbx and k3-233156 off, Paladin 11, level 28, merc, worn kit, CHEATED): with the fixes of rounds 1 to 6 the hero reaches the Barracks in both knob settings (`Entered Barracks`, step `to-barracks` ran). `Kill The Smith` then fails: `boss presets level 28 presets 5 monsterPresets null`, `preset 1/754 not found`, `boss: boss not found in the arena` after 4 to 46 s, three times, then `needs_attention`. Setting the quest word of record 3 to started (0x000C) changes nothing (k3-000016).
- V (probe `tools/probes/smith.py`, seed 5, `debug goto 28`): the lab verbs have no monster list; `debug presets` lists objects only (doors, barrels, chests, a shrine), no monster preset. The Barracks is 50 by 56 tiles.
- Not known (U): whether the Smith exists in the lab d2d Barracks at all (rooms up only near the hero; a quest condition of the real game, S, "Tools of the Trade active") or the match (class 402, name "The Smith", unique number -1) fails to read him. This card adds a note listing class/unique number/name of the monsters in view when the boss is not found, so the next run decides. Smith kill, Malus pick and the Charsi hand-in are NOT reached: the chunk is not green.

## Round 8 (card toolstrade7), the right verb

- S (docs/process/layout-read.md): `debug presets` lists objects; the monster presets of all rooms, superuniques included, come from `debug levelmonsters` and reach the bot only with `quests.layoutRead` 1 (d2dworld adds each superunique preset as an NPC preset with its name). The Smith script asked for table id 754 (U) without a name, so even with the knob on it could not match. Fix: the anchor carries `Name: "The Smith"` (the Countess works this way). Test `TestSmithAnchorResolvesByName` (fails without the name).
- Consequence for the sandbox: the Smith step needs `quests.layoutRead` 1 there too (without it the hero explores the Barracks for him, U).

## Round 9 (card toolstrade8): the Smith's room is behind doors

- V (run k3-005744 lr, lab d2d devp20 binary `PG_D2D_BIN=.../d2d.devp20`, knobs `nav.fullGrid 1, nav.warpRead 1, quests.layoutRead 1`, Paladin 11 level 28): `boss presets level 28 presets 66 monsterPresets ["754@23,57"]`, so preset 754 IS The Smith, at subtile (23,57). The earlier runs used the default lab binary (`survhold/bin/d2d.lab`), which has no `debug levelmonsters`: that was the whole "not found" (and the knob was off in some). With the preset the boss script's `goto anchor` fails: `nav: no path: reached the closest reachable point {114 257} to {23 57}` then `{23 57} is unreachable`: the Barracks has closed doors (`debug presets` lists classes 15 and 16, "Closed Door"), the Smith's room is behind them. `monsters in view` at the arena: classes 60, 21, 384, 172, 46, no Smith (his room is not up).
- Fix: `Navigator.GoTo` opens the door between the character and an unreachable spot (`openDoorToward`, up to four doors per call, no approach walk), then walks again. Test `TestGoToOpensTheDoorToASpotBehindIt` (fails without it).
- Needs for the sandbox: `quests.layoutRead` 1 and the d2d with `debug levelmonsters` (devp16 and later; the sandbox image is devp22).

## Round 10 (card toolstrade9)

- S (reading `internal/bot/loop.go`, `task.go`): the guard cancels the step on `SafetyDrinkLife` (`safetyInterrupt`), `do` drinks and the loop plans again; the Smith step re-enters `KillBoss`, which waited `bossRestartGap` (30 s) since the last start: 30 s standing in the Barracks crowd after every drink. Fix: a drink-type interrupt clears the start mark so the run restarts at once (`TestBossRunRestartsAtOnceAfterADrink`; the gap stays for every other cancel, `TestBossRunStartsAtMostOncePer30s`). The boss run itself restarts from `locate`, which walks to the anchor from where the hero stands, so no position is lost.
- Merc (U): the scenario pre sets `merc: 0`; the `minute` lines of the runs do not print the merc state. The route guard lines show `(merc=alive)` at the start of the chunk and `(merc=dead)` later in the same runs: the merc dies in the Outer Cloister crowd and is revived at the next town visit.
- `quests.layoutRead` prerequisite: see `docs/process/layout-read.md` (new section). Test `TestBossNotFoundNamesTheLayoutKnob`.
