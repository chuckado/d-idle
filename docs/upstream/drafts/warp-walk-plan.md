# Draft (local, NOT posted): "A clicked warp's walk is planned to its tile unit"

> status: r3, **needs the second verifier pass** (card prfix4, 2026-11-15; the first, verifyp02, asked for the changes applied here). Branch `fid/warp-walk-r3` on the private fork remote, head `d442f595`, two commits on upstream `4e054f43`: the oracle `c6f5b409` (builds alone; its `compare` fails on the parent), then the claim `d442f595`. Backup of r2 before the squash: `fid/warp-walk-r2-pre-squash` (`3b4439b3`). Not pushed to GitHub; posting waits for the verifier and the user's go. Body for the PR: [../bodies/warp-walk-plan-body.md](../bodies/warp-walk-plan-body.md) (same text, below only the verifier map).

PR title: `A clicked warp's walk is planned to its tile unit`

## What the second pass has to look at (the changed parts)

| Verifier finding (verifier-p02-2026-11-14.md) | What r3 does |
|---|---|
| 1 the test asserts nothing | `test_warp_walk` with no argument checks 8 starts against game.exe's path and outcome; reverting `plan_unit` makes 6 of 8 fail, exit 1 (V run) |
| 2 `compare` cannot tell parent from claim | the dump records the path `World::tick` plans (`T`); `compare` judges it against game.exe's first path and the outcome against the model; exit 1 on the parent in both sweeps, 0 on the claim (2.5 cell sweep) |
| 3 scope, Act 1 only | every level 1..136 with warps; 61,206 starts compared |
| 4 "0 after" overclaimed | the claim is the plan; 157 outcome differences outside the plan are listed (level 92 warp 0, level 81) |
| 5 commit order | oracle, then claim; 9 commits squashed to 2 |
| 6 body a sketch | below |
| 7 collision.md numbers | rewritten with the new counts |

---

## What changes

A click on a warp reaches game.exe as a walk to a unit target, type 5, like a click on an item, an NPC or a monster (`FUN_00548b00` case 5 -> `FUN_00548a50` -> `FUN_00580a70` mode 3). `FUN_00649970` takes the path's end from the tile unit (`FUN_006498a0`: its subtile, near 1), walks a blocked end back and plans the type 7 path with the unit-target rules. `World::tick` planned a clicked warp as a walk to a spot (`to_unit` false), so its path differed: the end got the spot rules (near 0, no search guard) and the route to a tile unit in rock often had 3 or 4 points where game.exe has 1.

One line in `World::tick`: `plan_unit = to_unit || actor->take_warp >= 0`, used for the `player_walk` call only. The re-plan rule and the end-of-walk re-path still use `to_unit`, so a warp keeps the single server re-walk (`walk_pending`). `to_unit` is already true for an NPC, an object, a portal and a monster, so for those nothing changes.

This claim is the plan. It is not a claim that every start now ends as game.exe's walk does: some outcome differences remain outside the plan and are listed under the limits.

Two commits: the oracle (`test_warp_walk --dump`, `tools/emu/warp_walk_path.py`, a README row; it builds on its own and its `compare` fails on the parent), then the claim (`tick.cpp`, the test's checks against game.exe's answers, a paragraph in `docs/research/re/collision.md`).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00548b00` `0x548bf9..0x548c1f` | the tile-unit case: `FUN_00641530` the unit distance; above `0x32` nothing; 4 or less takes the warp (`0x548c2f`); else `FUN_00548a50` with `edi = 5`, the target type. The same call with `edi = 4` (item, `0x548cd8`), `1` (`0x548d9c`), `0` (`0x548e1d`), `2` (`0x548bcc`): a warp is walked as any other unit target |
| `FUN_00548a50` | `FUN_00580a70(.., 0, 3, type, id, 0)`, then `FUN_00641f20` |
| table `0x6e1744`, entry 3 | `0x57f190`, mode 3's starter |
| `FUN_0057f190` | `FUN_00648b90` stores the target in the path (`0x57f1d2`), then `FUN_0057f090` (`0x57f1da`) |
| `FUN_0057f090` | `FUN_00648dc0` (a player's path back to type 7), `FUN_00649970(path, unit, 0)` at `0x57f117` |
| `FUN_006498a0` | the end is the target's static path `+0xc / +0x10` for types 2, 4, 5; the near byte by the jump table at `0x649958`: types 0 and 1 -> `0x649943` (1), 2 -> `0x649930`, 3 -> `0x649950` (1), 4 -> `0x64992b` (2), 5 is past the table: 1 |

## Counted differentials (tree = upstream `4e054f43` plus the branch)

`test_warp_walk --dump SEEDS FILE` clicks the centre of each warp's Select box through the real `World::apply` and `World::tick`, from a start a given distance off the warp's tile unit, and writes d2d's blocked bits round the unit, the path the tick planned, and whether the warp was taken (600 ticks, no monsters). `warp_walk_path.py game` runs game.exe's `FUN_00649970` (with `FUN_006498a0`, `FUN_00648120`, `FUN_00648050`, the type 7 pather `FUN_00679ed0`), the walker's frames `FUN_00650840` and the distance `FUN_00641530` under the emulator on those bits: its first path, and the outcome of its walk model (a walk, the distance test, one more walk, idle). `compare` judges every start: the plan the tick made against game.exe's first path, and the outcome against the model; it exits 1 on a plan that differs, on a start missing or listed twice, on a start that leaves the oracle's window, and on more outcome differences than `--outcome-max` (default 0).

| Sweep (levels 1..136 with warps) | starts | plan = game.exe, parent | plan = game.exe, claim | outcome differs, parent | outcome differs, claim |
|---|---|---|---|---|---|
| 2.5 cells, 8 directions, 30 seeds (`--dump 30`) | 21,023 parent, 21,021 claim | 20,507, **439 differ** | 20,974, **0 differ** | **146** (Act 1: 30, Acts 2 to 5: 116) | **0** |
| 1.5 / 2.5 / 3.5 / 5 / 7 cells, 16 directions, 6 seeds (`--dump 6 --dists 1.5,2.5,3.5,5,7 --dirs 16`) | 40,189 parent, 40,185 claim | 39,446, **542 differ** | 39,981, **0 differ** | 211 | 157 (see limits: not the plan) |

The two builds are two runs, so a start's neighbours differ a little (a level's walk flags fill room by room): each dump is compared against its own game.exe column, and the start counts differ by a few. Of the starts not counted as "equal": 219 on the claim (33 and 186) are starts where game.exe's first step is blocked and d2d's walk ends in the same tick, which clears the path, so the tick's plan is not visible; 32 on the claim (60 on the parent) are in a town, where the tick's crowd holds NPCs and the oracle has none. They do not fail the check; they are counted apart. The path the tick leaves after its first tick can lack the first point or two that the walker passed in that tick; `compare` accepts that and prints how many (0 points for 60,747 starts, 1 for 59, 2 for 149).

The test (no argument; it needs the game data and SKIPs without) runs 8 starts of seed 1 (levels 17, 46, 52, 76, 113 and 118, where the spot plan differs; 2 and 6 as controls) and checks the tick's path and whether the warp is taken against game.exe's answers (taken from the oracle). Built with the claim's `tick.cpp` put back to the parent's: **6 of the 8 fail, exit 1**; with the claim: 8 of 8, exit 0.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_warp_walk
set -o pipefail
D2_MPQ_DIR=<game data> build/tests/test_warp_walk       # 8 of 8 on the claim; 6 FAIL lines without plan_unit
cd tools/emu && uv sync
D2_MPQ_DIR=<game data> ../../build/tests/test_warp_walk --dump 30 /tmp/d.txt
uv run python warp_walk_path.py game /tmp/d.txt > /tmp/g.tsv
uv run python warp_walk_path.py compare /tmp/d.txt /tmp/g.tsv     # claim: OK, exit 0; parent: PLAN DIFFERS lines, FAIL, exit 1
D2_MPQ_DIR=<game data> ../../build/tests/test_warp_walk --dump 6 /tmp/w.txt --dists 1.5,2.5,3.5,5,7 --dirs 16
uv run python warp_walk_path.py game /tmp/w.txt > /tmp/wg.tsv
uv run python warp_walk_path.py compare /tmp/w.txt /tmp/wg.tsv --outcome-max 157   # claim: OK (its plan line: 0 differ)
```

## Honest limits

- The grid is d2d's (`Level::unit_blocked`), as is the start; the oracle has no rooms (the room lookups answer "in room"), no stamps and no NPC crowd, and the unit sizes are the player's 2 and a tile unit's 0 (`drlg.md`). The walker is the Sorceress's, at her walking speed, with no faster run/walk.
- The oracle builds the path record with the tile unit as the target (type 5) itself; that a click becomes that walk is the read of the addresses above, not something the emulator derives. The outcome model is the read rule of `FUN_00548b00` (distance, one more walk), not a run of that function.
- Outcome differences outside the plan remain. In the 1.5 to 7 cell sweep 157 of 40,185 starts still differ on the claim (211 on the parent): level 92 warp 0 (145 starts: d2d does not take the warp, the model takes it at distance 3) and level 81 warps 0 to 3 (12 starts: d2d takes it, the model stalls at distance 8). Their first paths equal game.exe's, so they are not this rule; candidates are an object or monster stamp near the tile unit (the oracle has none) and the per-tick reach (`warp_close` each tick, against game.exe's test when a walk ends); neither was checked (U). The 2.5 cell sweep has none.
- A start's state depends on the starts before it on the same level (the room walk flags fill as a level is entered), so the test's starts are ones that reproduce in isolation (level 92 warp 3 did not and is not in it); the sweeps record each start's own bits.
- Not covered: an operated trap door's walk (`take_warp` is also set in `objects.cpp`; a walk after it is now planned as a unit walk, U), and a Pickup or Resurrect walk whose point lies in a warp's Select box (the target becomes the warp, as before).
- Not in this PR: an item pick-up is the same unit walk with near 2 (`0x64992b`) while d2d plans it as a spot walk, and objects get near 2 or `FUN_006497d0` where `player_walk` gives near 1 to every unit; both are separate claims.
- d2d now stalls somewhat more often on a warp whose tile unit is in rock, exactly where game.exe's walk stalls. Our fork's earlier workaround (aim the walk at the nearest free subtile) is replaced by this rule; the bot's own click-point workaround stays. Neither is part of this PR.

## Checklist

- [x] A test in `tests/` citing the game.exe addresses, and the oracle in `tools/emu/`, in the same pull request (oracle commit first and building alone, claim second); the test fails without the fix
- [x] The finding in `docs/research/re/collision.md`
- [x] `python3 tools/pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN
- [x] `python3 tools/presubmit.py --quick --base upstream/main`: 0 FAIL, 2 WARN (build and gcc skipped by `--quick`; built, tested and run by hand)
- [x] `ctest -j3` on a full build (game data, no patch installer): 303 tests on the branch, 301 pass or skip; `test_outdoor` (abort) and `smoke_d2d` fail, the same two as on upstream `4e054f43` (302 tests, same two); `test_game`, `test_combat` and the new `test_warp_walk` pass; each commit builds alone
- [x] clang-tidy-19 and clang-format-19 on `world/tick.cpp` and `tests/game/test_warp_walk.cpp`: 0 findings, 0 format warnings

---

Agent: Claude Code
