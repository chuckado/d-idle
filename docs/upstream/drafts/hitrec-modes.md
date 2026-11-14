# Draft (local, NOT posted): "A player's block holds game.exe's length, and a hit never restarts a get-hit"

> status: r1, **ready for the independent verifier, with one open item: the game-side emulator run** (card decomp21, TASK-440, 2026-11-14); not posted. Branch `fid/hitrec-modes-r1` on the private fork remote, head `2cad31d5`, three commits on upstream `f5e11693` (oracle `c7dd8ead`, block claim `5a08affb`, restart claim `2cad31d5`; numeric noreply identity, last line `Agent: Claude Code`). Not stacked on swing-start any more: that PR's content is in upstream main (`379a2916`). Trace: `docs/research/hitrec-trace.md` (branch `agent/decomp20` until merged) Q1, Q3, Q4.1, D2, D3. Body: [../bodies/hitrec-modes-body.md](../bodies/hitrec-modes-body.md).

PR title: `A player's block holds game.exe's length, and a hit never restarts a get-hit`

## What changes

Two claims of `Fight::set_pmode` / `Fight::hurt_side`, one commit each after the oracle:

1. **Block length** (`rules::block_rate`, `rules::block_ticks`, the BL branch of `set_pmode`). game.exe's block branch of `FUN_00623f50` (`0x62403d..0x624084`, a player's mode 9) takes `rate = max(speed x (50 + E(FBR) + (Holy Shield's state 0x65 ? 50 : 0)) / 100, 1)`, `E = 120 v / (120 + v)` (stat 102), and `FUN_005539b0` ends the mode `max((256 frames - 1) / rate, 1) + 1` ticks later (the get-hit's clock, `get_hit_ticks`). d2d used `speed_frames(frames, fbr)` x 40 ms: no 50, no Holy Shield. A Paladin's sword block (PABL1HS, 3 frames at 256) holds 6 ticks at FBR 0 in game.exe, 3 in d2d.
2. **No restart** (`tick.cpp`, the get-hit test after a landed blow). `FUN_0057edd0` refuses GH and BL from GH and BL; d2d tested only BL, so every big hit inside a get-hit restarted the hold. The test adds `pmode != GH`.

Not in this PR (see Limits): the interrupt rule for swings (the used-skill lifetime is U), the block throttle `15 + FBR/8`, Sacrifice, the knock-back path's own `set_pmode(GH)`.

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe's block length is the closed form above | game side: S, V run by the HITREC trace's `hr_gh.py` (14,448 cells, 0 differ from the closed form, the game's own `FUN_00623f50` and `FUN_005539b0` under the emulator). **Not re-run on the new oracle file `block_len.py game`: the emulator was off limits in this card.** | `block_len.py game d2d.tsv`, then `compare` |
| d2d's block length differed | V run (the d2d side is the real `Fight::set_pmode(BL)`): `test_hit_recovery --block-length` on the oracle commit, 34,314 cells (7 classes x main-hand weapons x FBR 0..300 x Holy Shield off / on) against the closed form (`block_len.py closed`): **17,195 differ** (16,989 of 17,157 without Holy Shield, 206 of 17,157 with it) | build the oracle commit, run the two lines of `block_len.py`'s docstring |
| after the claim | the same 34,314 cells on `5a08affb`: **0 differ**; the unit block (Paladin breakpoints 0 / 13 / 32 / 86 -> 6 / 5 / 4 / 3 and Holy Shield 0 / 86 -> 3 / 2) | `test_hit_recovery` |
| game.exe does not restart a get-hit | S: HITREC trace Q4.1, `hr_modeswitch.py` (`FUN_0057edd0` over every mode pair: GH and BL refused from GH, BL, DT, DD; V run, 0 differ from the table) | `hr_modeswitch.py` |
| d2d restarted | V run: `test_hit_recovery --restart` (real `hurt_side`, two big hits on a Paladin 5 ticks apart): oracle commit and block commit print `R 10 15` (second hold ends 15 ticks after the first hit; game.exe's at 10) | build, run |
| after the claim | `R 10 10`; `test_hit_recovery` asserts it | `test_hit_recovery` |

## Limits and open items

- **Blocked:** `block_len.py game` (and the `hr_modeswitch.py` re-run) needs the emulator image on the worker host; until then the game column is S (the earlier V run of the same functions), not a fresh V run. The verifier should run it.
- Negative FBR is outside the domain (the run covered 0..300): `block_rate` follows `effective_speed` and floors at 1, U for game.exe's unsigned wrap there (the get-hit branch wraps, the block branch was not run below 0).
- A block animation plays only when no mode is set in d2d (`foe.blocked && pmode < 0`); game.exe plays it from any interruptible mode at most once per `15 + FBR/8` frames (D4, V read only): untouched, so BL after GH and the throttle stay as before.
- The knock-back branch (`tick.cpp`, `foe.knocked_from`) also calls `set_pmode(GH)` and is not changed here.
- Reach: blocks recover about 2x too fast in d2d without the claim (Paladin 3 vs 6 ticks), and a pack extends a flinch (15 vs 10); the cost to a Paladin in Act 1 Normal is small (HITREC Q11, under 4 percent), larger from Act 2.
- Checks still to run on the fork checkout before posting: `python3 tools/presubmit.py --quick`, `python3 tools/pr_guard.py`, ctest equal to the parent (four known data-set failures), `tools/emu/README.md` row for `block_len.py` (not added: the upstream README is a shared file; the verifier decides).
