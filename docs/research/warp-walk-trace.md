# Warp walk-stall: game.exe vs d2d (P02), first pass

> status: partial 2026-11-14 (decomp22, Sonnet). S-1 done (V run); the game.exe walk-to-a-blocked-goal side is NOT traced (U); no patch, no PR candidate. Tags: V run (run here), V read (disassembly read, capstone over the PE image), S (an earlier doc), I (inferred), U (unverified). Brief: `docs/upstream/briefs/warp-walk.md`; symptom: `docs/probes/warpclick-2026-11-14.md`.

## Questions

**Q1. Is the stall fixed on current upstream?** NO (V run). Harness: new test `tests/game/test_warp_walk.cpp --sweep N` (fork branch `fid/warp-walk-r1` `d7151557`, on upstream `4e054f43`; the private fork remote): for every level 1..39 with warps and every warp, 8 starts 2.5 cells off the warp's tile unit, a click at the centre of the warp's Select box (`World::warp_clicked` accepts it), through the real `World::apply` and `World::tick`, monsters cleared every tick, 600 ticks, taken = the level changed. 30 seeds, 12,480 starts; 6,623 start cells are on a blocked subtile and are not run (listed `-1`), 5,857 are run.

| Build | warp taken | of | share |
|---|---|---|---|
| upstream `4e054f43` (`d7151557` adds only the test) | 4,877 | 5,857 | 83.3 percent |
| plus the fork shim, target = `nearest_free(unit)` (`d7d365f0`, same change as `75891650`) | 4,972 | 5,857 | 84.9 percent |

Of the 5,857 starts, the tile unit stands on a blocked subtile (plus mask, `Level::unit_blocked`) in 3,510: parent 2,746 taken (78.2 percent), shim 2,841. The remaining 2,347 (unit free) are 2,131 taken on both builds (90.8 percent): 216 of those fail on both with the unit free, so a blocked unit is not the only cause of a stall (U for the others: start pockets that no walk leaves, or the Select-box point; I). Per warp table: `docs/research/data22/warp_sweep_by_warp.tsv`.

**The shim is not a clean fix.** Against the parent, the shim takes 192 starts the parent missed and LOSES 97 the parent took (10 warps, e.g. level 3 to 9: 13, level 7 to 12: 16, level 9 to 13: 12, level 11 to 15: 10, level 12 to 16: 9). So `nearest_free(unit)` is not game.exe's rule (I): either the free subtile game.exe walks to differs, or game.exe's planner handles the blocked goal itself. It cannot go up as P02 until the game side is known.

**Does this reproduce the brief's "0 of 8" (Black Marsh to the Forgotten Tower, level 6 to 20, seed 7)?** Not in this harness: level 6 to 20 takes 85 of 122 starts (69.7 percent) on both builds over 30 seeds and the tower's unit is free in these seeds (`unit_blocked` 0). The probe's `dirs.sh` and `tileclick.sh` harness is not in any tree (the scripts are gone), so the two numbers are different measurements; the earlier "0 of 8" is therefore U here, not contradicted.

**Q2. What does game.exe do on a click on a warp? (V read, `FUN_00548b00`)** Case 5 (use a warp, called with the unit id and a flag): look up the unit (`FUN_00552f60`, type 5), `FUN_00641530` (the unit distance between the player and the tile unit, size-based) must be at most 0x32 = 50 (`0x548bfe`); if it is 4 or less (`cmp eax, 4; jle`, `0x548c07`) the warp is taken (`FUN_005550b0`, `0x548c2f`); else (5 to 50) it queues the walk: `FUN_00548a50` (`0x548c1f`, mode 5) which calls `FUN_00580a70` (arg 3) and `FUN_00641f20`, which flags the unit (`+0x150 = 1`) and queues the action through `FUN_00460780`. This is the rule d2d implements in `World::warp_close` (under 5) and `warp_clicked`: S (`0ddb6fe7`) and V read here. NOT traced (U): `FUN_00460780`'s queued action, the server's run-to-unit loop that follows it and the path planner's handling of a goal subtile that is blocked (the reason this item is open).

**Q3. Does the game take the warp from the same starts?** U: no game.exe column exists for the approach walk.

## Deviations

| Id | d2d rule | game.exe rule | Tag | Differential |
|---|---|---|---|---|
| D1 | A clicked warp's walk targets the tile unit's own position (`commands.cpp` Move, `actor->target`); where it is blocked 21.8 percent of the blocked-unit starts (764 of 3,510) stall short of distance 5 (parent) | unknown: game.exe "enters there" (S, warpclick record) | the d2d side V run, game side U | could be built: run the walk-to-unit server loop (`FUN_00580c20` / `FUN_00641f20` -> `FUN_00460780` and the path planner `FUN_006503f0` family named in the d2d comments) under the emulator on the Level's walk grid (the `sk_tele_spot.py` style grid input already feeds `FUN_0064e7b0`), same (grid, start, unit) cases as `test_warp_walk`, oracle = player unit distance after N frames |
| D2 | the shim's `nearest_free` ring order is not `FUN_0064dea0`'s (`Level::warp_spot` already implements that order for arrivals: rings 1 to 49, least |dx|+|dy|) | for the target, unknown | I | the same differential; also try `warp_spot` as the shim target in `test_warp_walk` (one more build, a cheap first check; not run) |
| D3 | 216 starts with a free unit also fail on both builds | unknown | U | the same differential |

## Bot findings

None new. The bot's click point (`d2map.Exit` click columns, TASK-412) is unchanged; B-none.

## PR candidate

None. The deviation is not proven: the game side is untraced and the shim regresses 97 starts, so the fix may not be the target. The branch `fid/warp-walk-r1` holds only the sweep test (commit 1) and the experiment commit (`d7d365f0`, labelled as such, not a candidate).

## Drill recipe (nothing run)

In the lab (docs/forkdbg.md, docs/d2d-fork.md): `debug warps` for the unit cell and click point of a warp, `debug goto 6` (Black Marsh), place the hero with the lab verbs 2.5 cells off the tower's unit in one direction, click the point `debug warps` prints, `step` per tick and read `state` for the level; repeat for the 8 directions. The question for game.exe cannot be drilled (d2d is the only live engine); it needs the emulator differential of D1.

## Matrix and queue lines

- `docs/upstream/pipeline-2026-11-14.md` rows P02 and P03b carry a one-line result (done in this branch).
- Proposed for `docs/upstream/pr-queue.md` row 9, "Evidence" cell: "S-1 re-test 2026-11-14 on `4e054f43`: 4,877 of 5,857 starts take the warp, shim 4,972 but loses 97 (decomp22, docs/research/warp-walk-trace.md)".
- docs/decomp-queue.md: `decomp22: P03b no-restart claim confirmed by game.exe's FUN_0057edd0 (V run, 0 of 12 differ after, 1 before); P02 S-1: stall persists on upstream 4e054f43, shim loses 97 starts, game side untraced.`
