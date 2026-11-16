## What changes

The check that decides whether a MonUMod row may be rolled on a monster (`FUN_005a03e0`, called by the champion pick `FUN_005a0500` and the unique pick `FUN_005a0600` on each weighted row) refuses the Multiple Shots mod (fPick 2) on a monster whose MonStats flag `+0xc & 2` (isMelee) or `+0xe & 2` is set. `unique_detail::can_take` tested only the first (`type.melee`). The second is the `nomultishot` column: 11 rows (vampire1..9, slinger5, slinger6) refuse the mod in game.exe and accepted it in d2d. `MonType::no_multishot` reads the column and `can_take` tests it.

Two commits: the oracle (`tools/emu/umod_can_take.py`, `tests/game/test_umod_can_take.cpp`), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005a03e0` `0x5a046e..0x5a04c9` | fPick 2: refused when `MonStats +0xc & mask` or `+0xe & mask` (mask `[0x6ce26c]` = 2) |
| `FUN_005a0500` `0x5a056f..0x5a057f` | the pick calls it for each row with weight and `enabled` |

The game side is game.exe's own `FUN_005a03e0` under the emulator for every one of the 734 MonStats rows x 43 MonUMod rows (31,562 calls, an expansion game).

## Claims

| Commit | Claim or change | Evidence |
|---|---|---|
| `21d235d0` | the oracle and `test_umod_can_take --umod-can-take` / `--monster-ids` | 31,562 pairs; 11 differ on this commit alone: mod row 29 on rows 131..135 (vampire1..5), 394, 395 (slinger5, 6), 695..697 (vampire6..8), 731 (vampire9) |
| `6d1fc304` | the Multiple Shots mod is refused on a `nomultishot` monster | 0 of 31,562 differ; the test asserts it for vampire1 |

## How to re-check

```sh
set -o pipefail
cd tools/emu
uv run python umod_can_take.py game > game.tsv
../../build/tests/test_umod_can_take --umod-can-take | grep -E '^[0-9]+ [0-9]+ [01]$' > d2d.tsv
diff game.tsv d2d.tsv      # no output with the claim; 11 `<` lines on the oracle commit alone
cd ../..
python3 tools/pr_guard.py
```

## Limits

Only the pair rule. How often a unique of these rows spawns, and in which acts, was not counted; the effect is a unique vampire or slinger rolling Multiple Shots (and the pick's total weight differing for the other mods of those rows). That `+0xe & 2` is the `nomultishot` column is inferred from the equal count (11 rows, the same 11 names), not read from the loader.

## Checklist

- [ ] `python3 tools/presubmit.py --quick` 0 FAIL, `python3 tools/pr_guard.py` 0 / 0, ctest as the parent (not run yet; verifier)
- [ ] each commit builds alone (both built on the builder: yes)
