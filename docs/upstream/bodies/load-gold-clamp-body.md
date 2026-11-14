## What changes

game.exe's loader `FUN_0056b180` puts the saved purse and stash gold back through `FUN_0056af80` (`0x56afb4..0x56b00f`) and `FUN_0053ff00`, which sets the stat to 0 when it would pass its cap: `clvl x 10000` for the purse (`FUN_00622e70`), 2,500,000 for the stash (`FUN_00623460`). `World::take_character` kept both; it now zeroes them under the same tests (equal to the cap is kept).

Two commits: the oracle (`tools/emu/fl_gold.py`, `test_game --load-gold`; prints what the parent does), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0056b180` -> `FUN_0056af80` `0x56afb4..0x56b00f` | gold and stash gold are restored through `FUN_0053ff00` |
| `FUN_0053ff00` | a stat that would pass its cap is set to 0 |
| `FUN_00622e70`, `FUN_00623460` | the caps: `clvl x 10000`, 2,500,000 |

## Counted differentials

`fl_gold.py gen` writes 539 cases (11 levels x 7 purse values x 7 stash values around the caps); `fl_gold.py game` runs the real loader on them; `test_game --load-gold` prints d2d's `World::enter` for the same cases.

| Tree | cases that differ (of 539) |
|---|---|
| parent (oracle commit only) | 363 |
| with the claim | 0 |

`test_game` passes (it fails on the parent once the claim's unit block is added: an over-cap purse stays).

Reachability needs an edited or imported save; the PR does not claim a play path.
