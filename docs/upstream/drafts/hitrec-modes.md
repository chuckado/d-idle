# Draft (local, NOT posted): "A player's block holds game.exe's length"

> status: r2, **all required changes of the verifier done** (card prfix2, 2026-11-14; verifier report [../verifier-p03-p05-2026-11-14.md](../verifier-p03-p05-2026-11-14.md)); not posted, needs the user's go. The PR is now the block length ONLY. Branch `fid/hitrec-modes-r2` on the private fork remote, head `3264d61d`, two commits on upstream `a2b4f686` (oracle `371a4791`, claim `3264d61d`; the numeric noreply identity, last line `Agent: Claude Code`). Replaces `fid/hitrec-modes-r1` (three commits on `f5e11693`). The no-restart claim moved to its own branch `fid/hitrec-norestart-r1`, BLOCKED on a game-side oracle: [hitrec-norestart.md](hitrec-norestart.md), [../briefs/hitrec-norestart-oracle.md](../briefs/hitrec-norestart-oracle.md). Body: [../bodies/hitrec-modes-body.md](../bodies/hitrec-modes-body.md). Trace: `docs/research/hitrec-trace.md` Q1, Q3, D2.

PR title: `A player's block holds game.exe's length`

## What changes

`Fight::set_pmode(BL)` held `speed_frames(frames, fbr)` x 40 ms. game.exe's block branch of `FUN_00623f50` (`0x62403d..0x624084`, a player's mode 9) takes `rate = max(speed x (50 + E(FBR) + (Holy Shield's state 0x65 ? 50 : 0)) / 100, 1)`, `E = 120 v / (120 + v)` (stat 102), and `FUN_005539b0` ends the mode `max((256 frames - 1) / rate, 1) + 1` ticks later. A Paladin's sword block (PABL1HS, 3 frames at 256) holds 6 ticks at FBR 0 in game.exe, 3 in d2d. `rules::block_rate`, `rules::block_ticks`, one line in `set_pmode`.

Two commits: the oracle (`tools/emu/block_len.py`, `test_hit_recovery --block-length`; builds alone, prints the parent's gap), then the claim (code, unit block, `docs/research/re/combat.md` section and the stale line in `skills.md`). The oracle commit also adds the `tools/emu/README.md` row.

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe's block length is the closed form above | **V run** (card emurun, TASK-445, 2026-11-14, the emulator image on the worker host): `block_len.py game` ran game.exe's own `FUN_00623f50` block branch and `FUN_005539b0` on the 34,314 cells of d2d's oracle build: **0 of 34,314 differ from the closed form** (4 s on one core). Earlier V run of the same functions: `hr_gh.py`, 14,448 cells, 0 differ. | `block_len.py game d2d.tsv > game.tsv; block_len.py compare game.tsv d2d.tsv` |
| d2d differed | V run, same card: `block_len.py compare` of that game column against the oracle commit's d2d column: **17,195 of 34,314 differ** (first: AMBLHTH FBR 9 to 11, game 6 ticks, d2d 3). Independent: the verifier's own checker (`check.py`, written from the rule) 17,195 on the same file | as above |
| after the claim d2d matches | V run (card prfix2, NAS builder, release build with the game data): `test_hit_recovery --block-length` on `3264d61d` through the verifier's independent checker (`check.py`): **0 of 34,314 differ** from the closed form; the unit block asserts the Paladin breakpoints. The game column equals the closed form in every cell (row 1), so the head's `compare` is 0 by transitivity; **the literal `compare` of the emulator's `game.tsv` against the head's `d2d.tsv` was not run** (U, one command on the worker host) | `test_hit_recovery` |
| tests | V run: `test_game`, `test_combat`, `test_rules`, `test_swing_start`, `test_hit_recovery` rc 0 on `371a4791` and `3264d61d` | |
| lint | V run: clang-tidy-19 with the repo's `.clang-tidy` (warnings are errors) on the changed `.cpp` files of both commits: clean (the verifier's 3 findings were in the old test; the file no longer has `at` or `t0`, nor an unused header: the oracle commit alone is clean too). `pr_guard.py` 0 FAIL / 0 WARN on both branches | |

## Limits and open items

- The record a class and weapon pick is d2d's on both sides (the oracle reads frames and speed from d2d's output): stated in the body.
- The block throttle `15 + FBR/8`, BL from any mode (D4) and monster / merc blocks are untouched (stated in the body): in a dense pack a d2d Paladin can sit in BL longer than game.exe.
- Negative FBR is outside the run (0..300); `block_rate` floors at 1 (U for game.exe's unsigned wrap).
- The struck-cue side effect of the no-restart claim is gone from this PR (it belongs to the other branch).
- Not run here, owed before posting: the literal `compare` of a fresh `game.tsv` against the head's `d2d.tsv` (worker host), the full `tools/presubmit.py` and the whole `ctest` against the parent (the NAS builder ran the five changed-area tests, not the whole suite).
- Reach: blocks recovered about 2x too fast in d2d (Paladin 3 vs 6 ticks); the cost to a Paladin in Act 1 Normal is small (HITREC Q11, under 4 percent), larger from Act 2.
- Second verifier pass: needed only for what changed materially; see the pipeline row.
