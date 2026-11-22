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

Monastery Gate: `debug objects` lists object 5 (class 24, mode `op8`, state `NU`) at (28.3, 0.5) tiles, in the gap, and the engine's collision (`debug blocked` at 0.1 tile steps) is blocked at y 0.3 and 0.5 for every x while open at y -0.1, 0.1 (a 0.3 tile pocket) and from 0.7 on. The pocket at row 0 is cut off from the interior by the object: V that a closed gate-like object sits there; S that it is the Monastery gate door and that clicking it opens the cell (not tried: no live game, and `debug warps` and `debug warptriggers` both answer empty for level 26, so no tile or trigger carries this exit). The old reading "the neighbour's collision is built late" is not what happens here; the `pending 1` of the old log was another abutting level.

## Fix (card mapverify, `internal/d2dworld/border.go` `placeLinked`)

A level that `debug nearby` links but the layout lacks is placed from its offset: the extent across the shared edge is the offset, the extent along it is our own (the strip scan keeps only cells whose far side `across` finds open). A level whose offset leaves a gap, a corner or is far away (town from Cold Plains, `-2232 32`) is not placed. Must-fail tests: `TestPlaceLinkedGivesAGateToALevelTheLayoutLacks` (four sides, gap, corner, far, own size unknown, listed level untouched) and `TestLinkedLevelMissingFromLayoutGetsBorderCells` (fails with the call removed: no border cells to level 5). Geometry only: nothing in it is class or build specific.

What the fix does NOT do: open the Monastery gate. With it, 26 -> 27 gets a border contact (2 cells) that the planner cannot reach, 27 -> 28 and the Barracks chain get reachable contacts. See the result section for the scenario run.
