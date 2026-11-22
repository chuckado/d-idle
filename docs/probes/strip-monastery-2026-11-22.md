# Monastery Gate to Outer Cloister: why nav finds "no usable gate" (card mapverify, 2026-11-22)

> status: current. Probe `tools/probes/strip.py` (lab d2d of the fork branch `didle/rebase-20261122-fix1`, 611fc53b, `--headless --lab-verbs`, seed 5, one process per level pair; setup `debug goto` is CHEATED, mechanics only). Tags: V = measured in a named run, S = source or doc reading, U = unverified.

## Verdict

**V: the border strip is not "unreadable", it was never built.** `debug layout` (the act placements the bot's `d2dworld` abuts) does not list Outer Cloister (27) at all: 40 s after arrival at the Monastery Gate (26) the layout has levels 1 to 7, 17, 26 and 39 only. `debug nearby` does list it, `27 0 -40` (origin 40 tiles above ours), from the first read on. `d2dworld.adjacent` takes neighbours from the layout only, so level 27 never became an `Adjacent` entry: the failure line `nav no usable gate` shows `borderCells 0 unproven 0 cands 0` while `pending` counts only the levels the layout does list. The exits probe (`exitsread-2026-11-22.md`) read the link table, which is why it saw the edge and the bot did not.

The same holds for the next two links of the chain, V (same probe): Outer Cloister to Barracks (`28 -18 -56` in `nearby`, 28 missing from the layout) and Inner Cloister to Cathedral (`33 -4 -34`). Own level 27 and 32 are not in the layout either; their size and origin come from `debug level`.

## The strip (V, seed 5)

| link | edge | open cells in the level's own last row (subtile) | open past the edge (`debug blocked`) | note |
|---|---|---|---|---|
| 26 -> 27 | north, 64 tiles wide, 18 tall | 2 cells, x 140..141, row 0 only; rows 1 and 2 are wall across the whole width, the interior opens at row 3 | tiles 17..39 | closed by an object, see below |
| 27 -> 28 | north | 2 cells (rows 0 and 1), 4 cells (row 2), row 3 open: a connected 2-wide gap | open | reachable on the grid |
| 32 -> 33 | north | none in row 0 or 1; 6 cells in row 2 (x about 45..50), row 3 open | open | the gap ends two rows short of the strip row |

Monastery Gate: `debug objects` lists object 5 (class 24, mode `op8`, state `NU`) at (28.3, 0.5) tiles, in the gap, and the engine's collision (`debug blocked` at 0.1 tile steps) is blocked at y 0.3 and 0.5 for every x while open at y -0.1, 0.1 (a 0.3 tile pocket) and from 0.7 on. V (`tools/probes/objclass`, objects.txt of the builder's game data): class 24 is `Door`, IsDoor true, selectable. The pocket at row 0 is cut off from the interior by that closed door; S that using it opens the cells (the scenario run below decides; and `debug warps` and `debug warptriggers` both answer empty for level 26, so no tile or trigger carries this exit). The old reading "the neighbour's collision is built late" is not what happens here; the `pending 1` of the old log was another abutting level.

## Fix (card mapverify, `internal/d2dworld/border.go` `placeLinked`)

A level that `debug nearby` links but the layout lacks is placed from its offset: the extent across the shared edge is the offset, the extent along it is our own (the strip scan keeps only cells whose far side `across` finds open). A level whose offset leaves a gap, a corner or is far away (town from Cold Plains, `-2232 32`) is not placed. Must-fail tests: `TestPlaceLinkedGivesAGateToALevelTheLayoutLacks` (four sides, gap, corner, far, own size unknown, listed level untouched) and `TestLinkedLevelMissingFromLayoutGetsBorderCells` (fails with the call removed: no border cells to level 5). Geometry only: nothing in it is class or build specific.

## Second and third part of the fix: the doors (card mapverify, `internal/nav/gatedoor.go`)

The placement alone moved the failure one step on (V, scenario `act1.andariel.slaughter`, Paladin 11, CHEATED setup): `nav: no path: gate to level 27 is on the map but not reachable (cut off, no way out found)`. The Monastery Gate has TWO doors in the way, both objects.txt "Door" (`tools/probes/objclass`: classes 24 and 27, IsDoor true): class 27 at (28.1, 15.7) tiles across the arrival area, class 24 at (28.3, 0.5) on the border cells. d2d's collision (and so the grid) keeps a closed door as wall.

- `ExitTo` now calls `openGateDoor` when no gate is reachable: it floods the grid from the character and from the gate, takes the door object (map presets and snapshot objects, objects.txt `IsDoor`) whose nearest cell on each side is closest (both within 14 subtiles), walks beside it, uses it once (at most twice per door and game), forces the world to re-read the grid (`d2dworld.World.Refetch`) and plans again. If no door is listed yet (the presets list the rooms that are up; from the arrival point 5 of the 11 objects are) it first walks to the reachable cell nearest the gate and looks again.
- Found on the way, S by reading and V by the test: `Navigator.invalidate` ("a door opened") did nothing, because `fieldFor` builds its key from the stored key's nonce, so the two always matched; the planner kept the closed door. It now drops the cached field.
- Tests (must-fail, each fails with its step removed): `TestExitToOpensTheDoorBesideACutOffGate` (the door appears in the map and the snapshot only when the hero is within 60 subtiles, so the approach step is needed too; removing `approachGate` fails it), `TestExitToLeavesAnObjectThatIsNoDoorAlone`, `TestRefetchReadsTheGridAgain`. The geometry is class-free: nothing here knows a build or a class (the test world is the town wall of `gateWorld`).

## Result (V, builder lab slot, lab d2d 611fc53b)

`act1.andariel.slaughter`, Paladin 11: GREEN, 90 virtual minutes, `Entered Outer Cloister` asserted and seen, no death, writes after start 0 (before: RED on that assertion in three runs, `provingground-act1more-2026-11-20.md`). The log shows door (140,78) used, the exit walk, door (141,2) used, `Entered Outer Cloister`, then the Barracks, 10 `gate is cut off by a door` lines in all (the doors are met again after every return from town). Sorceress 3 (CHEATED setup, run 20261122-125549-f30d): RED on the same assertion, but never reached the Monastery Gate: the lab Sorceress is out of mana all run (2006 `combat out of mana` lines, `nav no usable gate` only for Rogue Encampment to Blood Moor in town), so the doors were not exercised for a caster. Two classes are therefore not shown live; the nav tests are class-free geometry.

Not verified: a real game.exe or the sandbox (d2play, devp18): the geometry is d2d's. The door interaction is the bot's click on the object; whether the real game opens the same door the same way is U. The tag CHEATED applies (setup puts the hero at level 26 with a clvl 26 kit): mechanics evidence about nav, not about play.
