# warptrigger: the trigger areas of warps, read from the engine (card warptrigger, 2026-11-22)

> status: current. Card `warptrigger` (first slice of `docs/design/fable-review-2026-11-22/map-knowledge.md`, user answer Q3: yes, a read verb under ADR 0022). Tags: V = measured in a named run, S = source reading of the private fork, U = unverified.

## What the engine does (S, fork source `components/game/world.hpp`, `world/commands.cpp`, `world/travel.cpp`; V by the test below)

A warp (cave mouth, stairs) is NOT taken by stepping on its tile. It is taken by a CLICK: `World::warp_clicked(warp, x, y)` is true when the click point, in world pixels relative to the tile unit's subtile (`(x - y) * 80`, `(x + y) * 40`), lies inside LvlWarp's Select box. A click inside sets `take_warp`; the hero then walks to the tile unit and `use_warp` fires once `World::warp_close` holds (unit distance under `kWarpReach`, 5 subtiles). So the real "trigger area" is two cell sets: the cells whose click takes the warp, and the cells from which that click is taken at once. In subtile space the first is a parallelogram (the Select box is a rectangle in screen space), not a disc. The bot's guessed bounds (`mouthTriggerR` 15, `mouthCoreR` 10 subtiles around the exit position) were fitted to logged clicks (DRILLS16 to 18).

## The fork verb (private fork, branch `didle/warp-triggers-20261122` off `didle/devp18-20261121`, code commit c2a9e006, tests up to 98deb79e8; pushed to the private remote the same day)

`debug warptriggers` (read-only, on the allow-list, no `--lab-verbs` needed): one JSON line `{level, unit:"subtile", scan, reach_rule, warps:[{index, to, unit:[x,y], select:[x,y,dx,dy], click:[[y,x0,x1]...], reach:[[y,x0,x1]...], truncated}]}`, `index` and `to` as `debug warps`. Both sets are asked of the engine's own predicates over the unit's subtile +-40, tested at the subtile centre (x + 0.5) / 5 like a client click; nothing is copied from the rule. `truncated` is true when a run touches the scan edge.

V (builder long lane, binary of c2a9e006, `tests/warptriggers_d2d.py`, seed 5, `--lab-verbs` SETUP for `debug goto`, `debug warp`, `debug kill`): on the Blood Moor (2), the Den of Evil (8) and Cave Level 1 (9), 4 warps:
- table: count, destination and tile-unit subtile equal `debug warps`, and its click cell lies in the click set: 0 differences;
- engine: for 6 listed click cells per warp, a real `cmd move` onto the cell from a listed free reach cell takes the warp (the level becomes the destination); for 6 cells just outside the set it does not: 0 + 0 differences (24 + 24 real clicks). Click sets were 32 to 66 cells, reach sets 69 cells, never truncated;
- read-only: `state`, `debug warps`, `debug presets` and the verb itself are equal before and after repeated reads, and the verb answers without `--lab-verbs`;
- `levelmonsters_d2d.py` and `lab_verbs_d2d.py` (the allow-list test) still pass.
Test lessons: a move onto the hero's own position is the controller's stop (the click on the cell the hero stands on never takes the warp), and a monster's blow interrupts the walk, so the test kills the level's monsters (SETUP) and never clicks the stand cell.

## Public side

`d2dctl.Client.WarpTriggers` (+ fake `Fix.WarpTriggers`), `d2dworld` Options.WarpRead (knob `nav.warpRead`, default 0) gives each exit a `d2map.Trigger` (click and reach runs in the map's coordinates), and the mouth gate (`internal/nav/mouthgate.go`) uses it: with the knob on and an exit that has an area, a click is refused only when it lands within `readTriggerMargin` (2 subtiles) of a cell that takes the warp (or is clipped back along its segment to the last cell that does not); exits without an area (stock d2d, verb missing, mismatch of destinations) keep the guessed 15 and 10 bounds. Tests: `internal/nav/warpread_test.go` (Paladin and Sorceress walks of the Tower Cellar 5 fixture: the click 13.5 from the stairs that the guessed bound refused goes through; a click on the area is refused; knob off or no area keeps the old refusal), `internal/d2dworld/triggers_test.go`, `internal/d2dctl/warptriggers_test.go`.

## Not verified

- U: the scenario `act1.forgotten.countess` with `nav.warpRead` 1 (needs the new fork binary in the Proving Ground; the knob is 0 by default, so the default behaviour is unchanged). A bot walking in with the verb is the first live check: `quests.layoutRead`-style, one run with `nav.warpRead` 1 on a Tower Cellar 5 game.
- U: Game.exe has no such verb; nothing is planned there (ADR 0023).
- The unit is a DS1 preset position per game; the area moves with it, so it is read per level on arrival (d2dworld does, with the other warp reads).
