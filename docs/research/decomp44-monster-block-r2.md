# decomp44: `fid/monster-block-r2`, the rebase of the monster block gate after the verifier's GO WITH FIXES

> status: current, 2026-11-20 (agent decomp44, Sonnet; nothing posted upstream). Applies the fixes of [decomp43-monster-block-verify.md](decomp43-monster-block-verify.md) section 6 (F1 to F5, F7) to the claim of [decomp43-monster-block-trace.md](decomp43-monster-block-trace.md). The PR body is [../upstream/bodies/monster-block-body.md](../upstream/bodies/monster-block-body.md). The pets question (the verifier's D6) is [decomp44-pet-block-trace.md](decomp44-pet-block-trace.md).

Tags: **V run** (run here on the builder, long lane), **V read**.

## 1. The branch

| | |
|---|---|
| Fork branch | `fid/monster-block-r2` (private fork remote, pushed; `r1` untouched) |
| Base | upstream `psi29a/d2decomped` main `c1b38311` (fetched 2026-11-20; r1 was on `8ae5cfcb`) |
| Commits (three, author and committer the public noreply identity, `Agent: Claude Code` the only trailer) | `8b1c4cd2` the oracle (`monster_block.py`, `test_monster_block --monster-block`); `fddcba54` the claim; `32e23986` the finish (README row, `MB_SPOT`, skip code, `compare` integers) |
| `pr_guard.py upstream-gh/main fid/monster-block-r2` | 0 FAIL, 0 WARN (V run) |

F1: the one conflict was `components/game/ai.hpp` (upstream `13c0b939`, `gear_res` in the resistance loop): resolved as `res[i] += boss_res[i] + gear_res[i];` with `block_chance(game_data)` as the block argument. F2: `test_monster_block` returns `D2D_TEST_SKIPPED` in every mode when the MPQs are missing (one helper, `exit_without_data`; the no-argument mode self-checks the helper on a missing directory before it touches data, so a green run proves the skip code: the helper is silent there). The 1.14d tables guard returns the skip code in both modes too. F7: `compare` accepts plain integers (`-?[0-9]+`) only. F3: the body says three commits.

## 2. Results (V run, builder long lane, fresh build directories of the rebased tree and its parent)

| Check | Result |
|---|---|
| Claim pins (`test_monster_block`, no argument, with data) | `ok`, rc 0, no SKIP line |
| Skip, MPQ directory missing, no argument | rc 77 (it was 0 on r1) |
| Skip, MPQ directory missing, `--monster-block` | rc 77 |
| `gen` | byte-identical to the stored `cases.tsv` |
| d2d claim output (2,469 cases) | byte-identical to r1's (`d2d-claim4.tsv`) |
| game.exe rerun | byte-identical to the stored `game.tsv` (27 s) |
| **claim against game.exe** | **0 of 2,469 differ**, rc 0 (93 skipped as before, 300 blocking) |
| **parent (the oracle commit alone) against game.exe** | **228 of 2,469 differ**, rc 1 (195 + 33 as before) |
| Room control `MB_SPOT=5` | 0 of 2,469, rc 0 |
| `compare` input checks (`1_0`, `+5`, empty, short, malformed d2d file) | all rc 1; a good triple rc 0 |

The mutants of the verifier (m1 to m19) were not rerun on r2: the rule lines are unchanged by the rebase (only the `gear_res` term joined the resistance loop next to the call), the body says so.

## 3. Full suite

The full default build and ctest of the rebased tree is queued on the builder long lane (build dir `/data/build/d44`; section 4 says where to read it).

## 4. State of the full ctest at the last edit of this doc

Not finished when this doc was committed (2026-11-20 14:30 EST): the build of the default target set (889 steps left at 14:27) runs on the long lane after the targeted checks of section 2, then `ctest -j2` with the data of the builder. Where to read it: the builder container files `/data/build/d44.log` (build) and `/data/build/d44-ctest.out` (ctest; its `tests passed` line); on the worker side `labrun builder-status` shows the job `d44-build` then `d44-ctest`. Every check of section 2 is independent of it and passed. The coordinator records the ctest line in the merge note when it lands; a failure of an unrelated test is to be compared against the parent tree (`/data/build/d44p`, only `test_monster_block` built there).
