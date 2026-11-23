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
