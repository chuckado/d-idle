## What changes

Gheed's gamble price (`FUN_0062efb0` mode 2 -> `FUN_00629370`) prices the item by `FUN_006287d0`, the Items record's `normcode` when it has one, else its own code. `rules::gamble_price` looked the stock item's own code up, so an upgraded base (an exceptional or an elite one, which the stock roll `FUN_005786a0` makes) was priced from its own cost and level. It now prices the normal version, as game.exe does.

Two commits: the oracle (`tools/emu/gamble_price.py`, `tests/game/test_gamble_price.cpp`; it builds alone and prints what the parent does), then the claim. The oracle's row is in `tools/emu/README.md`.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0062efb0` `0x62f062..0x62f09a` | mode 2: price = `FUN_00629370(item, clvl)` less the reduced-prices percentage (stat 0x57) |
| `FUN_00629370` `0x629379..0x62938b` | the base is `FUN_006287d0(item)`: `normcode` (Items `+0x84`) when non-zero, else the code (`+0x80`) |
| `FUN_00629370` `0x6293cc..0x6293e5` | `rin ` and `amu ` cost their `+0xd4` column; every other base mixes its own cost with the exceptional and elite costs by `pb` / `pu` |

The game side is game.exe's own `FUN_0062efb0` under the emulator on the price oracle's rig, 73,062 cases: the 369 codes a gamble tab can hold (gamble.txt's 125 plus the exceptional and elite codes their `+0x88` / `+0x8c` name) x clvl 1..99 x reduced prices 0 and 10.

## Claims

| Commit | Claim or change | Evidence |
|---|---|---|
| `373a2cf4` | the oracle and `test_gamble_price --gamble-price CASES` | 73,062 cases; 48,310 differ on this commit alone (an upgraded base is priced by its own record) |
| `eba73ede` | an upgraded gamble base is priced as its normal version | 0 of 73,062 differ; examples at clvl 45: `7ha` game 12,933, parent 51,186; `9ha` game 12,933, parent 11,033 |

## How to re-check

```sh
set -o pipefail
cd tools/emu
uv run python gamble_price.py gen > cases.tsv
uv run python gamble_price.py game cases.tsv > game.tsv
../../build/tests/test_gamble_price --gamble-price cases.tsv | grep -E '^[a-z0-9]+\.[0-9]+\.[0-9]+ -?[0-9]+$' > d2d.tsv
diff game.tsv d2d.tsv      # no output with the claim; 48,310 `<` lines on the oracle commit alone
cd ../..
python3 tools/pr_guard.py
```

## Limits

The stock roll itself (`FUN_00578790`: the 14 codes, levels and qualities) is not part of this change and was not run. The data word `+0x30` that mode 2 reads (`FUN_0062a670`) is not understood; the probe shows the price does not depend on quality, identified, ethereal, the NPC or the difficulty. `elx` (not a gamble base) prices negative in game.exe and is outside the grid.

## Checklist

- [ ] `python3 tools/presubmit.py --quick` 0 FAIL, `python3 tools/pr_guard.py` 0 / 0, ctest as the parent (not run yet; verifier)
- [ ] each commit builds alone (both built on the builder: yes)
