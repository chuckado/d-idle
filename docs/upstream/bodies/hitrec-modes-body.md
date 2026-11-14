## What changes

1. A player's block holds game.exe's length. `FUN_00623f50`'s block branch (`0x62403d..0x624084`) takes `rate = max(speed x (50 + E(FBR) + 50 with Holy Shield's state 0x65) / 100, 1)`, `E = 120 v / (120 + v)`; `FUN_005539b0` ends the mode `max((256 frames - 1) / rate, 1) + 1` ticks later (the get-hit's clock). d2d used `speed_frames(frames, fbr)` x 40 ms. A Paladin's sword block is 6 ticks at FBR 0 in game.exe, 3 in d2d. `rules::block_rate` and `block_ticks`, one line in `Fight::set_pmode`.
2. A hit never restarts a get-hit. `FUN_0057edd0` refuses GH and BL from GH and BL; d2d tested BL only. One more condition in the get-hit test of `Fight::hurt_side`.

Three commits: the oracle (`tools/emu/block_len.py`, `tests/game/test_hit_recovery.cpp`; prints what the parent does), then one commit per claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00623f50` `0x62403d..0x624084` | the block branch (a player's mode 9): the rate with stat 102 and state 0x65 |
| `FUN_00621740` (table row 3 at `0x6e8e24`) | `120 v / (120 + v)` for stat 102 |
| `FUN_005539b0` | the mode-end event `max(n, 1) + 1` ticks later |
| `FUN_0057edd0` | may the unit enter mode N from its mode: GH (4) and BL (9) refused from GH, BL, DT, DD |

## Counted differentials

`block_len.py` and `test_hit_recovery --block-length` (d2d's real `set_pmode(BL)`: 34,314 cells, 7 classes x main-hand weapons x FBR 0..300 x Holy Shield off / on).

| Tree | cells that differ from game.exe's rule (of 34,314) |
|---|---|
| parent (oracle commit only) | 17,195 |
| with the block claim | 0 |

`test_hit_recovery --restart` (two big hits on a Paladin 5 ticks apart through `hurt_side`): parent `R 10 15`, with the claim `R 10 10`.
