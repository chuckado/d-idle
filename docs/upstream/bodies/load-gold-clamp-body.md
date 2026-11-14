## What changes

game.exe's loader `FUN_0056b180` puts the saved purse and stash gold back through `FUN_0056af80` (`0x56afb4..0x56b00f`) and `FUN_0053ff00`, which sets the stat to 0 when it would pass its cap: `clvl x 10000` for the purse (`FUN_00622e70`), 2,500,000 for the stash (`FUN_00623460`). `World::take_character` kept both; it now zeroes them under the same tests (equal to the cap is kept).

Two commits: the oracle (`tools/emu/fl_gold.py`, `test_game --load-gold`; it builds alone and prints what the parent does), then the claim. The finding is in `docs/research/re/d2s-header.md` ("Gold on load"), the oracle's row in `tools/emu/README.md`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0056b180` -> `FUN_0056af80` `0x56afb4..0x56b00f` | gold and stash gold are restored through `FUN_0053ff00` |
| `FUN_0053ff00` | a stat that would pass its cap is set to 0 |
| `FUN_00622e70`, `FUN_00623460` | the caps: `clvl x 10000`, 2,500,000 |

The game side is a run of game.exe's own `FUN_0056b180` under the emulator (`fl_gold.py game`, 539 cases: 11 levels x 7 purse values x 7 stash values around the caps) on the current `tools/emu` image: its answers equal the rule above in 539 of 539 cases. The oracle is the same file as an earlier run of ours (a few seconds on one core).

## Claims (a batch only)

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `dda1fa57` | the oracle: game.exe's loader on 539 cases; `test_game --load-gold` prints d2d's `World::enter` for the same cases and exits 1 when the case file cannot be opened (ctest `test_game_load_gold_missing_cases`) | `0x56afb4..0x56b00f` | "How to re-check" |
| `d229705a` | a saved purse over `clvl x 10000` or a stash over 2,500,000 is 0 on entering a game; equal to the cap is kept | `0x56afb4..0x56b00f`, `FUN_00622e70`, `FUN_00623460` | "How to re-check" |

On the oracle commit alone every cell with an over-cap purse or stash differs: 363 of the 539 (33 of the 49 purse and stash pairs at each of the 11 levels), because d2d clamped none; with the claim 0 differ. That count is the grid's arithmetic, not a measurement of how often it happens.

## How to re-check

```sh
set -o pipefail
cd tools/emu
uv run python fl_gold.py gen > cases.tsv
uv run python fl_gold.py game cases.tsv > game.tsv
../../build/tests/test_game --load-gold cases.tsv | grep -E '^[A-Za-z_][A-Za-z_0-9]* -?[0-9]+( -?[0-9]+){1}$' > d2d.tsv
diff game.tsv d2d.tsv      # no output with the claim; 363 `<` lines on the oracle commit alone
cd ../..
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure
```

## Checklist

- [x] A test in `tests/` (`test_game`'s unit block: an over-cap save is 0; `test_game_load_gold_missing_cases`: an unopened case file is a nonzero exit)
- [x] The finding in `docs/research/re/` (`d2s-header.md`, "Gold on load"); no `todo.md` / `unverified.md` row existed for it
- [x] `python3 tools/pr_guard.py` passes

Limits: reaching an over-cap save needs an edited or imported save; this PR claims the loader's rule, not a play path. The stash cap is a second local constant next to `kStashGoldMax` in `commands.cpp`; I can share one constant if you prefer.

---

Agent: Claude Code
