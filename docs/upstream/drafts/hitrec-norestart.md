# Draft (local, NOT posted): "A hit never restarts a get-hit" (r3)

> status: r3 2026-11-15 (card prfix3). Superseded by the PR body [bodies/hitrec-norestart-body.md](../bodies/hitrec-norestart-body.md); this page is the r1 history. The r2 verifier (`docs/upstream/verifier-p03b-2026-11-14.md` on branch `agent/verifyp03c`, agent/verifyp03c) found the claim also refused a big hit during a knockback slide, which game.exe allows (FUN_0057edd0 allows mode 4 from 0x13); r3 fixes that, extends the oracle to modes 0..19 with a knocked-back player, makes `compare` fail on missing or malformed lines, and fixes two stale comments. Fork branch `fid/hitrec-norestart-r3`: probe `530d8df9`, oracle `c3cc0993`, claim `b18acf7e`, stacked on `fid/hitrec-modes-r2` (#541, open: post only after it merges, then a second verifier pass).

PR title (when unblocked): `A hit never restarts a get-hit`

## What changes

`FUN_0057edd0` (may the unit enter mode N from its current mode) refuses GH (4) and BL (9) from GH, BL, DT and DD; d2d tested BL only, so every big hit inside a get-hit called `set_pmode(GH)` again and restarted the hold: two big hits 5 ticks apart held 15 ticks in d2d, 10 in game.exe. One more condition in the get-hit test of `Fight::hurt_side` (`tick.cpp`) and a comment fix. The server's get-hit call goes through `FUN_005809d0` -> `FUN_0057edd0` (call at `0x57d383`, V read).

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe refuses GH and BL from GH and BL | S: the HITREC trace ran `FUN_0057edd0` over every mode pair (`hr_modeswitch.py`, game.exe's own code); **the script is not committed anywhere upstream-bound and has no diff mode**: this is the missing oracle | the brief |
| d2d restarted | V run (card prfix2, NAS builder): `test_hit_recovery --restart` on the probe commit `d8d35ca9`: `R 10 15`; `--hit-modes` (one big hit in each mode 0..18 with a hold running): the GH cell restarted the hold, BL is already refused | `test_hit_recovery --restart`, `--hit-modes` |
| after the claim | V run: `R 10 10`; `--hit-modes` GH cell not restarted (the only difference between the two builds); `test_game`, `test_combat`, `test_rules`, `test_swing_start`, `test_hit_recovery` rc 0; clang-tidy-19 clean on the changed files | |

## Open items (why it is blocked)

- The game-side oracle (`tools/emu/mode_switch.py`, `game` and `compare`, README row), written and run on the worker host: the exact steps are in the brief.
- A refused big hit no longer replays `play_player_hit` (the struck cry) through `set_pmode(GH)`; whether game.exe still sends the struck cue (flag `0x8000`) when `FUN_005809d0` refuses is not traced (U): say so in the body or trace it.
- The boundary tick (a second hit exactly at the hold's end) is V read only (events before units); add a test once decided.
- Swing modes (A1, KK, S1, S3, S4) stay uninterrupted: that is P04, not this PR.
