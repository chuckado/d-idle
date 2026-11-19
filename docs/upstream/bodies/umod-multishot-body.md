## What changes

The check that decides whether a MonUMod row may be rolled on a monster (`FUN_005a03e0`, called by the champion pick `FUN_005a0500` and the unique pick `FUN_005a0600` on each weighted row) refuses the Multiple Shots mod (fPick 2) on a monster whose MonStats flag `+0xc & 2` (isMelee) or `+0xe & 2` (nomultishot) is set. `unique_detail::can_take` tested only the first (`type.melee`). 11 rows (vampire1..9, slinger5, slinger6) refuse the mod in game.exe and accepted it in d2d. `MonType::no_multishot` reads the `nomultishot` column and `can_take` tests it.

Two commits: the oracle (`tools/emu/umod_can_take.py`, `tests/game/test_umod_can_take.cpp`, a README row), then the claim (`gamedata_load.cpp`, `montypes.hpp`, `uniques.hpp`, the finding in `docs/research/re/monsters.md`, and the asserts in `test_umod_can_take` that pin game.exe's counts: vampire1 and slinger5 take 40 of 43 rows, sk_archer1 41; the Multiple Shots row is refused for the first two and taken by the third).

In play: row 29 (Multiple Shots) has `champion` empty and upick 6 in every difficulty, so only random uniques are affected, and no superunique is of the 11 classes. These monsters can be uniques in Act 1 (Catacombs level 3: vampire5), Act 2 (Dry Hills and Lost City: slinger5, slinger6; the seven Tal Rasha's Tombs and the Arcane Sanctuary: vampire1), Act 3 (Travincal, the Temples, the Durance: vampire2..4) and Act 5 in Nightmare and Hell only (vampire6..8; `nmon` of Worldstone Keep level 3, Throne of Destruction and Baal's Temple 1). Such a unique also draws its other mods from a smaller total weight, so the draw itself changes.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005a03e0` `0x5a046e..0x5a04c9` | fPick 2: refused when `MonStats +0xc & mask` or `+0xe & mask` (mask `[0x6ce26c]` = 2, read from the image) |
| `FUN_005a0500` `0x5a056f..0x5a057f` | the champion pick calls it for each row with weight and `enabled` |
| `FUN_005a0600` `0x5a067c` | the unique pick calls it; `FUN_005a0600` is called from `FUN_005a0760` (`0x5a08f3`, a random unique) and `FUN_005a49b0` (`0x5a4b3f`, a superunique's extra mods) |

The column semantic is read, not inferred: for all 734 MonStats rows, bit 1 of the dword at `+0xc` equals MonStats.txt's `isMelee` and bit 17 (`+0xe & 2`) equals `nomultishot`, 0 mismatches. The 11 rows with bit 17 are vampire1..5 (131..135), slinger5, slinger6 (394, 395), vampire6..8 (695..697) and vampire9 (731).

The game side is game.exe's own `FUN_005a03e0` under the emulator for every one of the 734 MonStats rows x 43 MonUMod rows (31,562 pairs, an expansion game).

## Claims

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `7bd3ddec` | the oracle and `test_umod_can_take --umod-can-take` / `--monster-ids` | 31,562 pairs; 11 differ on this commit alone: mod row 29 on rows 131..135 (vampire1..5), 394, 395 (slinger5, 6), 695..697 (vampire6..8), 731 (vampire9) | `umod_can_take.py compare` prints `11 differ`, exit 1 |
| `fdea73d1` | the Multiple Shots mod is refused on a `nomultishot` monster, with the finding and the pinned counts | 0 of 31,562 differ | `umod_can_take.py compare` prints `31562 pairs, 0 differ`, exit 0 |

Mutants (local): the parent rule under the claim's test aborts (vampire1 and slinger5 refuse the fPick 2 row); `nomultishot` refusing every row leaves 440 pairs differing and aborts the test (`taken(vampire) == 40`).

## How to re-check

```sh
set -o pipefail
# the d2d side needs D2_MPQ_DIR (as test_game); umod_can_take.py game needs the emulator's game image and a built mpq-cat, as the neighbouring tools/emu rows
cd tools/emu
uv run python umod_can_take.py game > game.tsv
../../build/tests/test_umod_can_take --umod-can-take | grep -E '^[0-9]+ [0-9]+ [01]$' > d2d.tsv
uv run python umod_can_take.py compare game.tsv d2d.tsv   # claim: 31562 pairs, 0 differ, exit 0; oracle commit alone: 11 differ, exit 1
cd ../..
python3 tools/pr_guard.py
```

`compare` also exits 1 on an empty or unequal file, a malformed line or ids out of step.

## Limits

- Only the pair rule. How often a vampire or slinger unique rolled Multiple Shots in d2d before the fix was not counted.
- The classic-game refusal (a row id of 100 or more when the second argument is 0) is not run; the oracle passes an expansion game.
- Mods that game.exe gives without `FUN_005a03e0` (an act boss's, a family's, Skills sumumod, a minion's copy) do not use `can_take` in d2d either, and are not touched.

## Answers to the presubmit WARNs

- `siblings`: no skill or hit path is touched; the change is the mod pick's eligibility, used by the champion, unique and superunique picks (`FUN_005a0500`; `FUN_005a0600` from `FUN_005a0760` and `FUN_005a49b0`; `pick` in d2d).
- `data-gating`: the test reads the 1.14d tables through `D2_MPQ_DIR` and skips without them, as `test_load_gold`.
- `doc-claims`: every MonStats row x every MonUMod row; the grid is exhaustive (734 x 43).
- `build`, `gcc`: built and ctest run (Release, GCC, on the builder): 2 of 349 fail (`test_outdoor`, `smoke_d2d`), the same 2 as on upstream main `def80df7`; `test_umod_can_take`, `test_rules` and `test_thief_mod` pass.

## Checklist

- [x] A test in `tests/` (fix or port): `tests/game/test_umod_can_take.cpp`
- [x] The finding in `docs/research/re/`, its `todo.md` / `unverified.md` row moved (trace): `docs/research/re/monsters.md`
- [x] `python3 tools/pr_guard.py` passes (0 FAIL, 0 WARN)

---

Agent: Claude Code
