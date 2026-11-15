# Draft (local, NOT posted): "A clicked warp's walk is planned to its tile unit"

> status: r2, **ready for the independent verifier** (card decomp23, 2026-11-15). Branch `fid/warp-walk-r2` on the private fork remote, head `3b4439b3`: oracle commit `18439e09` (builds alone, prints the parent's counts), claim commit `4016ff0d`, later commits style, a `docs/research/re/collision.md` row, the oracle script `tools/emu/warp_walk_path.py` and its README row. Base: upstream `4e054f43`. Supersedes the shim `d7d365f0` (not game.exe's rule, see limits).

PR title: `A clicked warp's walk is planned to its tile unit`

## What changes

A click on a warp reaches game.exe as a walk to a unit target of type 5 (`FUN_00548a50` -> `FUN_00580a70` mode 3 -> `FUN_0057f190`); `FUN_00649970` takes the path's end from that unit (`FUN_006498a0`: a static unit's own subtile, near 1) and plans a type 7 path with the unit-target rules. `World::tick` planned it as a walk to a spot (`to_unit` false), which differs in the pather's near distance and the search guard. One line: `plan_unit = to_unit || actor->take_warp >= 0` for the `player_walk` call only (the client's re-issue, which `to_unit` also controls, does not apply to a warp: the server re-walks it once, `walk_pending`).

## Claims

| Claim | Commit | Evidence | Re-check |
|---|---|---|---|
| d2d's unit-target plan equals game.exe's `FUN_00649970` path in 5,857 of 5,857 starts (every Act 1 to 5 warp, 8 start directions 2.5 cells off its tile unit, 30 seeds; 324 distinct level / warp / direction combinations) | `18439e09` | V run: `warp_walk_path.py compare` | `test_warp_walk --dump 30 d.txt`; `warp_walk_path.py game d.txt > g.tsv`; `warp_walk_path.py compare d.txt g.tsv` -> `first path: 5857 equal, 0 differ of 5857` |
| planned as a spot, d2d differs from game.exe in 31 starts (2 combinations), and its outcome (warp taken after the two walks) in 30 (level 17 warp 0, direction 3) | `18439e09` | V run: the `F` lines and the `W` lines against the game column | the same files |
| after the claim commit the outcome differs in 0 of 5,857 (4,847 taken, as game.exe's walk model) | `4016ff0d` | V run | `test_warp_walk --dump 30 d.txt` on the claim build, then the game column on its dump; `test_warp_walk --check` prints OK |

## Limits (U, say so in the PR)

The grid is d2d's (`Level::unit_blocked`); the oracle has no rooms (the room lookups are hooked to "in room"); the walk speed is the Sorceress's walking speed with no FRW; one combination of 324 changes outcome. The full ctest was not run on the branch. The earlier fork shim (`nearest_free(unit)`) is NOT proposed: against the same oracle it differs from game.exe in 319 starts (222 taken that game.exe stalls, 97 lost).
