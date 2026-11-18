## What changes

A player's block holds game.exe's length. `FUN_00623f50`'s block branch (`0x62403d..0x624084`, a player's mode 9) takes `rate = max(speed x (50 + E(FBR) + 50 with Holy Shield's state 0x65) / 100, 1)`, `E = 120 v / (120 + v)` (stat 102); `FUN_005539b0` ends the mode `max((256 frames - 1) / rate, 1) + 1` ticks later (the get-hit's clock). d2d used `speed_frames(frames, fbr)` x 40 ms: no 50, no Holy Shield. A Paladin's sword block (PABL1HS, 3 frames at 256) is 6 ticks at FBR 0 in game.exe, 3 in d2d. `rules::block_rate` and `block_ticks`, one line in `Fight::set_pmode`.

Two commits: the oracle (`tools/emu/block_len.py`, `tests/game/test_hit_recovery.cpp`; it builds alone and prints what the parent does), then the claim. The finding is in `docs/research/re/combat.md` ("Hit recovery: the length"), the oracle's row in `tools/emu/README.md`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00623f50` `0x62403d..0x624084` | the block branch (a player's mode 9): the rate with stat 102 and state 0x65 |
| `FUN_00621740` (table row 3 at `0x6e8e24`) | `120 v / (120 + v)` for stat 102 |
| `FUN_005539b0` | the mode-end event `max(n, 1) + 1` ticks later |

The game side is game.exe's own code under the emulator (`block_len.py game`: `FUN_00623f50`'s block branch from `0x624020`, then `FUN_005539b0`) on every cell d2d produces: its ticks equal the closed form above in 34,314 of 34,314 cells (7 classes x main-hand weapons x FBR 0..300 x Holy Shield off / on; a few seconds on one core).

## Claims (a batch only)

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `371a4791` | the oracle: game.exe's block length per cell, `test_hit_recovery --block-length` prints d2d's `Fight::set_pmode(BL)` hold for the same cells; `compare` and `closed` exit 1 when no cell was compared | `0x62403d..0x624084` | "How to re-check" |
| `3264d61d` | a player's block holds `block_ticks` (Holy Shield's 50 included) | `0x62403d..0x624084`, `FUN_00621740`, `FUN_005539b0` | "How to re-check" |

On the oracle commit alone 17,195 of the 34,314 cells differ (16,989 of 17,157 without Holy Shield, 206 of 17,157 with it); with the claim 0 differ. The unit block in `test_hit_recovery` pins a Paladin's breakpoints: FBR 0 / 13 / 32 / 86 give 6 / 5 / 4 / 3 ticks, with Holy Shield 3 and 2.

## How to re-check

```sh
set -o pipefail
cd tools/emu
../../build/tests/test_hit_recovery --block-length | grep '^D ' > d2d.tsv
uv run python block_len.py game d2d.tsv > game.tsv
uv run python block_len.py compare game.tsv d2d.tsv     # claim: 34314 cells, 0 differ; oracle commit alone: 17195 differ
cd ../..
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure
```

## Checklist

- [x] A test in `tests/` (`test_hit_recovery`: the rules, and a Paladin's block through `set_pmode`)
- [x] The finding in `docs/research/re/` (`combat.md`; the stale line in `skills.md` that said `speed_frames` shortens GH / BL is corrected); no `todo.md` / `unverified.md` row existed for it
- [x] `python3 tools/pr_guard.py` passes

Limits:
- Which record a class and weapon pick (PABL1HS, AMBLHTH, ...) is d2d's on both sides of the comparison; the oracle takes frames and speed from d2d's AnimData record, as the claim is the length, not the choice.
- The block throttle (`15 + FBR / 8` frames between two animations) and entering BL from any mode are untouched. After this change a d2d Paladin in a dense pack can block back to back at the real length without game.exe's throttle, and still never blocks out of an attack.
- Monster and merc blocks (game.exe mode 6, same branch) are not changed here.
- Negative FBR is outside the run (0..300); `block_rate` floors at 1.

---

Agent: Claude Code
