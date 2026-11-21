# WARPWALK: a clicked warp's walk, game.exe vs d2d vs the bot (P02), plus P04 drill R1 and P06 S-1 (decomp23, TASK-473)

> status: finished 2026-11-15 by agent decomp23 (Sonnet). Offline: the emulator image (`docs/emu-image.md`) on the worker host and the NAS builder for d2d builds; no game, no live d2d, no Go change, nothing posted. Follows [warp-walk-trace.md](warp-walk-trace.md) (decomp22: the S-1 re-test and the sweep, game side untraced), [../probes/warpclick-2026-11-14.md](../probes/warpclick-2026-11-14.md) and [../upstream/briefs/warp-walk.md](../upstream/briefs/warp-walk.md). Tags: **V run** (game.exe's own code run natively under the emulator against an oracle written here, files named), **V read** (disassembly, addresses and the rule in words), **S** (an earlier doc or upstream source), **I** (inferred), **U** (unverified). Names from `d2-name-map-2026-11-12.md` are references only; no game code is pasted.

## Summary

1. **game.exe walks a clicked warp as a UNIT target, and d2d did not (V run, V read).** The server plans it as path type 7 with a target unit (`FUN_00649970`; the end comes from the tile unit by `FUN_006498a0`, near 1; a blocked end is walked back by `FUN_00648120`), d2d planned it as a walk to a spot (`tick.cpp`: `to_unit` is false for `take_warp`). Run natively on d2d's own collision bits at every start of the decomp22 sweep (5,857 starts, 324 distinct level / warp / direction combinations because most levels are not random): the unit-target plan equals game.exe's path in **5,857 of 5,857 starts**; the spot plan differs in 31 (2 distinct combinations: level 17 warp 0 direction 3, level 19 warp 0 direction 7).
2. **The walk's outcome differs in exactly one combination (30 starts, V run).** Level 17 warp 0 from direction 3: game.exe's two walks end 9 units short and the warp is not taken; d2d took it. After the fix (`fid/warp-walk-r2`) d2d agrees with the game.exe model in **5,857 of 5,857** (4,847 taken of 5,857; before 4,877 taken, 30 differ).
3. **So the "stall" is almost entirely game.exe's own rule on d2d's grid.** Unit free: 2,347 starts, 2,131 taken, 216 not, identical on both sides. Unit in rock: 3,510 starts, 2,716 taken, 794 not (game.exe model). The probe's "game.exe enters there" (docs/probes/warpclick-2026-11-14.md) was a belief (U), not a trace; this trace shows game.exe's own path code stalls the same starts (V run), with the grid caveat in Q5.
4. **The fork shim (`nearest_free(unit)`, `d7d365f0`) is NOT game.exe's rule (V run).** Against the same game.exe model it differs in 319 starts (44 distinct combinations): 222 starts it takes that game.exe stalls, 97 it loses that game.exe takes. It must not go up as P02; the parent's two rules (end walked back, then the pather) are game.exe's.
5. **A small PR candidate exists** (local branch pushed to the private fork only): `fid/warp-walk-r2` head `3b4439b3`, oracle commit `18439e09`, claim commit `4016ff0d`. Mismatches before 30 (outcome) / 31 (path), after 0 of 5,857. Thin (one combination of 324); see section PR candidate for why it is still worth the verifier's time and what is U.
6. **P04 drill R1, the emulator part: the used-skill pointer does NOT outlive a swing (V run).** Every mode start that is not a skill (idle, walk, run) clears it; the swing's end handler goes through the idle start. A Paladin who last used Zeal flinches again as soon as the swing ends. 2,856 swing-end runs and 714 walk-start runs, 0 failures.
7. **P06 S-1: the missile block gate is still open on the current upstream tree (V read)**; no commit fixes it. Section P06.

## Part A. The warp walk

### Q1. What does game.exe do from a click on a warp to the first step? (V read)

Addresses are game.exe 1.14d.

1. C->S 0x13 on a tile unit -> `FUN_00548b00` case 5 (S, `collision.md` "pending interact"): unit distance `FUN_00641530` at most 4 takes the warp (`FUN_005550b0`); 5 to 50 calls `FUN_00548a50`.
2. `FUN_00548a50` -> `FUN_00580a70(game, unit, 0, 3, type 5, id, 0)`: target lookup `FUN_00552f60`, the gates `FUN_0057edd0` / `FUN_0057eec0`, a player's pending fields +0x150 / +0x154 cleared, then the starter `0x6e1744 + 3 * 8` = `0x6e175c` -> **`FUN_0057f190`** (table value read; modes 2 and 3 share it).
3. `FUN_0057f190(mode, target)`: `FUN_00648b90` stores the target in the unit's path (+0x58 the unit, +0x5c its type, +0x60 its id), then `FUN_0057f090(mode)`: `FUN_00648dc0` puts a player's path back to type 7, **`FUN_00649970(path, unit, 0)`** plans, `FUN_00620210(unit, 0)` clears the used skill, `FUN_00648780` asks for points and the mode events are queued (`FUN_00553570`, `FUN_00553990`, `FUN_00553f00`); no points -> idle (`FUN_0057f020`).
4. `FUN_00649970`: **`FUN_006498a0` sets the path's end (+0x10 / +0x12) from the target**: for the types read as static (2, 4, 5: a tile unit is type 5) the unit's static path +0xc / +0x10, for 0, 1, 3 the dynamic position; and it returns the "near" byte, 1 for type 5 (stored at ctx +0x14). If the end is blocked (`FUN_0064d910`, pattern +0x48, mask +0x50 0x1c09) and the path has flag 0x1000 (types 0, 1, 7: `FUN_00648cf0` ors 0x21900 for type 7), `FUN_00648120` walks it back toward the start (3 tries a round; start reached -> **no path**) and, for a player, `FUN_00648050` pushes it on. Then the pather `0x6eb6d8[7]` = `FUN_00679ed0` (toward `FUN_00679c80`; accepted when its end is within near of the target and not the start; else the search `FUN_0067b850` within 18 subtiles; else the toward path), then `FUN_0064fe40` drops leading points on the start subtile.
5. The walk runs frames (`FUN_00650840`, S: 16.16 subtiles, `rules::walk_frame`). It ends (path done, blocked, no path), `FUN_00580c20` goes idle and calls `FUN_00548b00` again (`LAB_00580e73`): in reach (4 or less) it takes the warp; out of reach it makes ONE more walk (pending +0x150 cleared by `LAB_00580e88`) and the second end is plain idle (S, collision.md; d2d: `walk_pending`, `test_walk_item_twice`). A tile unit's +0x150 pending is set by `FUN_00641f20` (`LAB_00641f37`).

### Q2. Does d2d's plan equal game.exe's? (V run for the path, one deviation D1)

Oracle (new): `tools/emu/ww_path.py` (copied to the fork as `tools/emu/warp_walk_path.py`, commit `ffa7f842` and later). For each start of d2d's sweep it runs, under unicorn, game.exe's own `FUN_00648cf0`, `FUN_00649970` (with `FUN_006498a0`, `FUN_00648120`, `FUN_00648050`, `FUN_00679ed0` / `FUN_00679c80` / `FUN_0067b850`, `FUN_0064fe40`), `FUN_00650840` frames and `FUN_00641530`, on a fake path / unit / tile unit record. Hooked (the oracle's inputs, not claims): `FUN_0064d910` and `FUN_0064ff90` answer from d2d's blocked bits (the player's shape, 129 x 129 subtiles round the tile unit, taken at the start's own world state; 0 window misses), the room lookup `FUN_00463740` (always a room), the stamp lift `FUN_00649560` / `FUN_00649400` (none), the unit size `FUN_00620510` (player 2, tile 0), the chase check `FUN_006503f0` (go on: the unit never moves). d2d side: `test_warp_walk --dump N FILE` (fork `fid/warp-walk-r2`): the same bits, d2d's `player_walk` path planned as a unit target (`D` lines) and as a spot (`F` lines), the three `rules::` pathers raw (`E` lines, window re-check `M 0` in all 5,857) and the real `World::apply` / `World::tick` sweep result (`W` lines).

| Plan | starts | equal to game.exe's path | differ |
|---|---|---|---|
| d2d `player_walk` as a unit target (`D`) | 5,857 | 5,857 | 0 |
| d2d `player_walk` as a spot (`F`, what `World::tick` planned for a warp) | 5,857 | 5,826 | 31 (2 of 324 combinations: level 17 warp 0 direction 3, level 19 warp 0 direction 7) |

The wrapper (walk-back + pather) and the three `rules::` pathers therefore match game.exe on 5,857 REAL starts, on top of the random-grid oracles (`walk_back.py`, `moves.py`). One artefact of the first runs is written down so nobody repeats it: the dump's start was first printed rounded from a float product and landed one subtile off for 41 starts; the hero is now read the way the tick reads it (`Fine` to float), and the first-path differences vanished (the 41 were the harness, not d2d).

### Q3. Do both sides take the warp from the same starts? (V run)

Outcome model (V read for the rule, V run for every part that has an address): path request, frames until `FUN_00650840` returns 0, `FUN_00641530` at the end (taken at 4 or less), else ONE more walk from where the hero stands, then idle. d2d side: the real tick (600 ticks, level change = taken).

| tile unit | d2d (parent) taken / of | game.exe model taken / of | outcome differs |
|---|---|---|---|
| free | 2,131 / 2,347 | 2,131 / 2,347 | 0 |
| in rock | 2,746 / 3,510 | 2,716 / 3,510 | 30 (all level 17 warp 0 direction 3, 30 seeds) |
| all | 4,877 / 5,857 | 4,847 / 5,857 | 30 |

After `fid/warp-walk-r2` (`plan_unit = to_unit || take_warp >= 0` in `tick.cpp`): d2d 4,847 / 5,857, outcome differs **0**; the spot plan's single bad combination is gone. The lone start type: tile unit in rock at distance 9 behind a wall; game.exe's second walk finds no free spot (`FUN_00648120` reaches the start: no path) and idles; d2d's spot plan found a way round and took the warp.

The sweep's counts are 30 seeds of 324 distinct combinations (the random levels differ per seed; the others repeat 30 times): the counts above weigh each combination by how many seeds ran it, they are not 5,857 independent trials.

### Q4. Is the fork shim game.exe's rule? NO (V run)

`d7d365f0` (target `nearest_free(unit)`), built on the oracle commit and run through the same dump: takes 4,972 of 5,857; against the game.exe model it differs in 319 starts (44 of 324 combinations), 222 more taken than game.exe and 97 fewer. So the shim helps a bot that stalls, but it is not a fix of a d2d deviation: game.exe's walk to the tile unit stalls the same starts (V run). A bot that wants the warp from a stalled start must do what a player does (click again from where it stands, or step closer); see B1.

### Q5. What is not shown here (U, written down)

- **The grid is d2d's.** Both sides use `Level::unit_blocked` bits; that the bits at warp tile units and around them equal game.exe's own collision map (the DRLG-built walk flags) is covered by the earlier DRLG / walk-flag oracles (S: `drlg_walk_oracle.py`, `drlg.md`), not re-run here. In particular 3,510 of 5,857 starts have the tile unit in rock (28 of 52 warps in seed 1); if d2d's marking of those cells differs from game.exe's the stall counts move, the planning rule does not.
- **Velocity.** The oracle walks at `path_velocity(walk velocity, 100)` of the Sorceress (the sweep's character); the live tick uses `walk_percent`, which is 100 for a hero without FRW (I, not read for this sweep).
- **Per-tick reach.** d2d tests `warp_close` every tick, game.exe only when the packet arrives and when a walk ends (S, collision.md); the sweep shows no start where that alone changes the outcome after the fix (V run, 0 of 5,857), so it is not a measured deviation; a hero with FRW or a different start distance may differ (U).
- **Rooms.** The oracle has no rooms: `FUN_00463740` / `FUN_00647fb0` answer "in room" for every point (hooked). A tile whose unit is in another room than the hero is not modelled (U); the real stall of Black Marsh to the Forgotten Tower (level 6 to 20) from the probe's 0 of 8 was not reproduced (decomp22: 69.7 percent here), so that exact figure is still U.
- **Inside level 20** (20 to 6, 20 to 21: U in the probe): the oracle takes the map as given; those warps are in the sweep like the others and take like the others on both sides; the probe's 0 of 8 there is not explained (U).

## Deviations d2d vs game.exe (warp walk)

| Id | d2d rule | game.exe rule | Tag | Differential |
|---|---|---|---|---|
| D1 | A clicked warp's walk is planned as a walk to a spot (`tick.cpp` `to_unit` false: near 0, no search guard) | planned as a unit target: end from the tile unit, near 1, `FUN_0067b740` guard (V read `FUN_006498a0`, `FUN_00649970`) | V run (31 path starts, 30 outcome starts, one combination) | yes, built: `test_warp_walk --dump` + `warp_walk_path.py`; fixed on `fid/warp-walk-r2` |
| D2 | the fork shim aims at `nearest_free(unit)` | the end is the tile unit's own subtile; blocked: `FUN_00648120` walk-back | V run (319 starts differ) | yes, same harness (build `d7d365f0` on the oracle commit) |
| D3 | d2d tests the reach each tick | game.exe tests it at the packet and at each walk end, one more walk | V read; no start differs after D1 (V run) | covered by the outcome model (a start where it matters would show as a `W` mismatch) |
| D4 | none found in the planner: `rules::toward_path`, `search_path`, `player_path`, `end_walk_back` equal game.exe on 5,857 real starts | | V run | the same harness |

## Bot findings (described, not changed)

- **B1** `internal/nav/exit.go:112` walks to the tile and stops within 2, then `pushThrough` (`:140`, defined `:349`) clicks the warp's Select box. By Q3 a click from 2.5 cells away is stalled in 17 percent of this sweep's starts on game.exe's own rule (216 of 2,347 with the unit free, 9 percent; 794 of 3,510 with the unit in rock, 23 percent): the click is not guaranteed to take the warp. The bot already retries (step off and back `:148-153`, `gateClicks` `:160`); that is the right shape and needs no rule change. An unmet case would be a stalled start where every retry stalls too; not measured.
- **B2** The click point comes from `d2dctl.Warp` `HasClick` (`gateCands`, `internal/nav/exit.go:227-229`); the oracle clicks the same Select-box centre (`test_warp_walk`), so the sweep's numbers apply to the bot's click.
- **B3** An older d2d without the click columns keeps the tile click (`exit.go:225-226`); it works only inside the Select box (S, the probe).

## PR candidate

**Yes, local and thin.** Fork branch `fid/warp-walk-r2` (pushed to the private fork remote; head `3b4439b3`; oracle `18439e09` builds alone and prints the parent's numbers, claim `4016ff0d` = `plan_unit` plus `--check`, later commits style, `docs/research/re/collision.md` row, the oracle script and its README row). Mismatches (outcome): before 30 of 5,857, after 0; path (spot plan): before 31, after 0 (the `D` column equals game.exe's in 5,857 of 5,857). Why still worth it: it is the only measured planning difference left between d2d and `FUN_00649970` for warps, it replaces the shim with a one-line rule that game.exe's code proves, and the oracle (a game.exe walk to a tile unit on 5,857 real starts) is the first for the approach walk. Honest limits for the verifier: one combination of 324 changes outcome; U items in Q5; the full ctest was NOT run (only `test_warp_walk` builds: `--check` prints OK, `presubmit --quick` 0 FAIL, `pr_guard upstream/main` 0 FAIL 0 WARN). Verifier brief: [../upstream/briefs/verify-warp-walk-r2.md](../upstream/briefs/verify-warp-walk-r2.md). Draft PR text: [../upstream/drafts/warp-walk-plan.md](../upstream/drafts/warp-walk-plan.md). Nothing is posted.

## Drill recipe for the fork verbs (nothing run; docs/forkdbg.md, docs/forkcheats.md)

- **W1 (D1, level 17 warp 0, direction 3):** lab binary, `debug goto 17`, `debug warps` (the tile unit cell and click point), place the hero with the lab verbs 2.5 cells off the unit on the (-1, +1) diagonal, click the printed point, `step` and `state`: expect the level NOT to change after two walks (game.exe model); the parent changes it.
- **W2 (Black Marsh to the tower, level 6 to 20, the probe's 0 of 8):** the same from each of the 8 directions; expect the stalled set to equal this oracle's (it takes 85 of 122 starts here, decomp22), not the probe's 0 of 8; any 0 of 8 left is a grid or room difference (Q5), not the walk.
- **W3 (shim):** the same starts on a build with `d7d365f0`; expect the 319 differences by direction.

## Matrix-cell updates (proposed text)

- `docs/upstream/pr-queue.md` row 9, Evidence: "decomp23 (V run): d2d's plan as a unit target equals game.exe's path in 5,857 of 5,857 starts; the stall is game.exe's own rule on the grid; the shim differs from game.exe in 319 starts; fix `fid/warp-walk-r2` (plan as a unit): outcome mismatches 30 before, 0 after".
- `docs/upstream/pipeline-2026-11-14.md` rows P02 and P04 (done in this branch) carry one line each.
- No skill-matrix row: warp walking is not a skill.

## Part B. P04 drill R1: does the used-skill pointer outlive a swing? (emulator steps)

Brief: [../upstream/briefs/hitrec-interrupt.md](../upstream/briefs/hitrec-interrupt.md), the blocker "the used-skill pointer's lifetime"; the live drill R1 of `hitrec-trace.md` (branch `agent/decomp20`).

| Step | What it asks | Answer | Tag |
|---|---|---|---|
| R1.1 | where is the used skill kept | the unit's skill list +0x10, set by `FUN_00643990` through `FUN_00620210(unit, skill)`, read by `FUN_00620250` | V read |
| R1.2 | who sets it | `FUN_00580a70` when a mode starts with a non-null skill (`0x580a47`); the other non-null callers are a packet handler (`0x56aeaa`, a skill id 0x5a path) and monster AI (`0x5de03a`) | V read (the two others U for what they serve) |
| R1.3 | who clears it | `FUN_0057f020` (the idle start: `0x57f061`), the walk and run starters `FUN_0057f090` (`0x57f11f`, reached from the table entry for modes 2 and 3), and mode starters `0x57fdff` / `0x57fe6f`; the callers of `FUN_0057f020` include the swing-end handler `FUN_00581020` (`0x5810b1`) and the walk-end code in `FUN_00580c20` | V read |
| R1.4 | after the swing ends, does a hit flinch | `FUN_00581020` for every attack-like mode (A1, A2, SC, TH, S1..S4) and every one of 357 Skills.txt rows as the used skill: used skill 0 and mode NU afterwards, 2,856 of 2,856; the gate (`FUN_0057eec0`, as the get-hit call at `0x57d383` passes) allows GH afterwards for all rows; a null used skill returns "allowed" at `0x57ef10` | V run (`tools/emu/ww_used.py` A to C) |
| R1.5 | after a walk or run start | the table starters for modes 2 and 3 clear it for all 357 rows, 714 of 714; so a Paladin that walks off after Zeal flinches | V run (`ww_used.py` E); NU start from any mode 1,785 of 1,785 (D) |
| R1.6 | the gate while the swing runs | refused exactly for the 22 rows whose record byte +7 bit 0x80 is clear, 0 of 357 differ (decomp20's `hr_interrupt.py` result reproduced in this rig) | V run (A) |
| R1.7 | with Attack as the right-hand skill | the client's selection never reaches the server's used skill; only a started mode sets it (R1.2) | V read; the live part is not needed for the lifetime |

Verdict: **the lifetime is the swing**: a Zeal or Smite swing is not interrupted (the gate refuses while it runs), and the first tick after it ends the unit flinches like any other. That removes the blocker of P04: the claim "interrupt rule = the used skill's `interrupt` column while a swing or cast runs" is safe, and d2d's current "never interrupt a melee swing" (D1 of hitrec-trace) is wrong for Attack and Sacrifice and right for Zeal and Smite. Still U for P04 (not part of R1): the d2d side differential (a `Fight` run at tick 3 of an Attack swing, hitrec-trace R2), which needs a d2d build; that the TP scroll path goes through SC in d2d (D5).

Files: `tools/emu/ww_used.py` (run with `bash tools/emu/ww_run.sh ww_used.py` on the worker host). Remaining live-only: none of R1 (R2 to R6 stay as hitrec-trace lists them).

## Part C. P06 S-1: the missile block gate on current upstream (V read)

Fork checkout, `git fetch upstream`, `upstream/main` = `4e054f43` (2026-11-15). `components/rules/combat.hpp` `monster_blow(..., bool missile, ...)` at line 945: lines 952 and 953 roll the player's shield block (`defender.block`) and weapon block (`defender.weapon_block`) for every blow with `!sure`, and the missile call (`components/game/ai.hpp:1090`, `monster_blow(..., missile.src, true, rng, true)`) goes through them: there is **no test for a non-zero physical damage** before the block. The gate is NOT fixed upstream (V read; `git log -S"defender.block" 7557a54d..upstream/main -- components/rules/combat.hpp` lists no commit). P06 stays open as the pipeline row says; its game-side oracle exists (decomp DECOMPMON), the d2d side and the patch are the work of the P06 brief. Not started here.

## Dead ends and limits

- First dump: grids were taken once per warp (a level's walk flags fill room by room, so later starts of the same warp saw other bits): now one grid per start, and the E / M lines prove the window equal when the paths are asked (0 of 5,857 differ).
- First dump: the data loader logs on stdout in the middle of a line: the dump now goes to its own file (`--dump N FILE`).
- First dump: the start was printed rounded, one subtile off for 41 starts (an artefact of the harness, written in Q2).
- A python port (`moves.py`'s `player_path`) and the C++ `rules::player_path` on the same bits agreed with game.exe's `FUN_00679ed0` for the one case that looked like a pather deviation (`E` line check); nothing was left of it.
- The full ctest of the fork (the 4 known data-set failures) was not run; the build holds only `test_warp_walk` (and `test_game`, built, not run).

## Files and reproduce

In this repo: `tools/emu/ww_path.py` (`game`, `compare`), `ww_outcome.py` (the outcome table), `ww_search.py` (pather-level check), `ww_used.py` (P04 R1), `ww_dis.py` (disassembly and caller scan helper), `ww_run.sh` (runs a script in the emulator image). Fork (private remote): `fid/warp-walk-r2`.

```sh
# NAS builder: build test_warp_walk of fid/warp-walk-r2, then
D2_MPQ_DIR=<game data> test_warp_walk --dump 30 d.txt           # 5,857 starts: W, G, D, F, E lines, S 4877 5857 6623 on the parent
# worker host (emulator image):
source $D2_EMU_DIR/env.sh; export PYTHONPATH=$EMU
$PY tools/emu/ww_path.py game d.txt > g.tsv                       # game.exe's column
python3 tools/emu/ww_path.py compare d.txt g.tsv                  # first path: 5857 equal, 0 differ
python3 tools/emu/ww_outcome.py d.txt g.tsv                       # parent: differs in 30; fix build: 0
bash tools/emu/ww_run.sh ww_used.py                               # P04 R1: A..E all 0 failed
```
