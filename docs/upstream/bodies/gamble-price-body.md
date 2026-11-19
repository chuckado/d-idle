## What changes

Gheed's gamble price (`FUN_0062efb0` mode 2 -> `FUN_00629370`) prices the item by `FUN_006287d0`, the Items record's `normcode` when it has one, else its own code. `rules::gamble_price` looked the stock item's own code up, so an upgraded base (an exceptional or an elite one, which the stock roll `FUN_005786a0` makes) was priced from its own cost and level. It now prices the normal version, as game.exe does.

In play, at clvl 45 an elite Hand Axe (`7ha`, a Tomahawk) cost 51,186 in d2d and costs 12,933 in game.exe; an exceptional one (`9ha`, a Hatchet) cost less than in game.exe below clvl 45 (clvl 30: 5,493 against 6,896).

Two commits: the oracle (`tools/emu/gamble_price.py`, `tests/game/test_gamble_price.cpp`, a README row; it builds alone, prints what the parent does, and its test has no failing assert), then the claim (`components/rules/rules.hpp`, the finding in `docs/research/re/store.md` (Gambling: the normal code and the version branch), and the asserts in `test_gamble_price` that pin game.exe's 771, 6,896, 12,933 and 11,640 for `hax`, `9ha`, `7ha` at clvl 1, 30, 45 and 45 with 10 percent reduced prices).

FUN_00629370 has one caller, `0x62f07a` in mode 2 (a scan of `.text` for `call` / `jmp` and of every section for pointers), so buy, sell and repair prices are untouched. d2d's two callers of `rules::gamble_price` are the payment (`store_gamble`) and the hover price (`apps/d2d/store.cpp`); both take the fix.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0062efb0` `0x62f040..0x62f0bb` | mode 2: reduced = stat 0x57 capped at 99; the item's version (item data `+0x30`, `FUN_0062a670`) at least 1: price = `FUN_00629370(item, clvl)` less `MulDiv(price, reduced, 100)` |
| `FUN_00629370` `0x629379..0x62938b` | the base is `FUN_006287d0(item)`: `normcode` (Items `+0x84`) when non-zero, else the code (`+0x80`) |
| `FUN_00629370` `0x6293cc..0x6293e5` | `rin ` and `amu ` cost their `+0xd4` column; every other base reads the normal record's uber and ultra codes (`+0x88` / `+0x8c`), stack (`+0xe4` / `+0xe8`) and cost, and mixes the exceptional and elite costs by `pb` / `pu` |
| `FUN_006287d0` callers | `0x62937a` (`FUN_00629370`), `0x62f09e` (mode 2, version 0) and `0x630282` (a bit writer, not a price) |

The game side is game.exe's own `FUN_0062efb0` under the emulator on the price oracle's rig: the 369 codes a gamble tab can hold (gamble.txt's 125 plus the exceptional and elite codes their `+0x88` / `+0x8c` name) x clvl 1..99 x reduced prices 0 and 10 (73,062 cases), and a second grid of the same codes x clvl 1, 7, 30, 45, 63, 99 x reduced 1, 3, 7, 13, 25, 33, 50, 77, 99 (19,926 cases), because game.exe subtracts a `MulDiv` and d2d truncates.

## Claims

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `f9a127e1` | the oracle and `test_gamble_price --gamble-price CASES` | 73,062 cases; 48,310 differ on this commit alone (an upgraded base is priced by its own record); the extra grid: 13,176 of 19,926 differ | `gamble_price.py compare` prints `48310 differ`, exit 1 |
| `48d8f6ee` | an upgraded gamble base is priced as its normal version, with the finding and the pinned asserts | 0 of 73,062 and 0 of 19,926 differ; at clvl 45 `7ha` game 12,933, parent 51,186; `9ha` game 12,933, parent 11,033 | `gamble_price.py compare` prints `73062 cases, 0 differ`, exit 0 |

Mutants (local, each caught): the parent rule under the claim's test aborts (`== 771`); the normcode taken only for an exceptional base (the elites by their own record) leaves 24,156 of 73,062 and 6,588 of 19,926 cases differing and aborts the test (`== 771`, for `7ha`).

## How to re-check

```sh
set -o pipefail
# the d2d side needs D2_MPQ_DIR (as test_game); gamble_price.py game needs the emulator's game image and a built mpq-cat, as the neighbouring tools/emu rows
cd tools/emu
uv run python gamble_price.py gen > cases.tsv
uv run python gamble_price.py game cases.tsv > game.tsv
../../build/tests/test_gamble_price --gamble-price cases.tsv | grep -E '^[a-z0-9]+\.[0-9]+\.[0-9]+ -?[0-9]+$' > d2d.tsv
uv run python gamble_price.py compare game.tsv d2d.tsv   # claim: 73062 cases, 0 differ, exit 0; oracle commit alone: 48310 differ, exit 1
cd ../..
python3 tools/pr_guard.py
```

`compare` also exits 1 on an empty or unequal file, a malformed line or ids out of step.

## Limits

- Mode 2 takes `FUN_00629370` when the item's version (item data `+0x30`, `FUN_0062a670`; 101 in an expansion game, 0 in a classic one) is at least 1; with 0 the price is the normal record's `+0xd4` (`0x62f09d..0x62f0af`). d2d models the expansion branch only; the classic one is not compared. A classic game prices by `+0xd4`, which d2d does not model.
- The stock roll itself (`FUN_00578790`: the 14 codes, levels and qualities) is not part of this change and was not run. The probe shows the price does not depend on quality, identified, ethereal, the NPC (Gheed, Akara, Charsi) or the difficulty. `elx` (not a gamble base) prices negative in game.exe and is outside the grid.

## Answers to the presubmit WARNs

- `data-gating`: the test reads the 1.14d tables through `D2_MPQ_DIR` and skips without them, as `test_load_gold`.
- `build`, `gcc`: built and ctest run (Release, GCC, on the builder): 2 of 349 fail (`test_outdoor`, `smoke_d2d`), the same 2 as on upstream main `def80df7` (2 of 348); `test_gamble_price`, `test_rules` and `test_thief_mod` pass.

## Checklist

- [x] A test in `tests/` (fix or port): `tests/game/test_gamble_price.cpp`
- [x] The finding in `docs/research/re/`, its `todo.md` / `unverified.md` row moved (trace): `docs/research/re/store.md`
- [x] `python3 tools/pr_guard.py` passes (0 FAIL, 0 WARN)

---

Agent: Claude Code
