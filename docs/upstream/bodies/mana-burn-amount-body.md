## What changes

A Mana Burn monster's landed melee hit takes `lo * 256` up to `hi * 256` mana points off the player in game.exe, where `lo` / `hi` are the whole points d2d keeps in `Monster::mana_lo` / `mana_hi`; d2d took `lo` to `hi` points. `FUN_005a1f90` stores the burn as `x << 8` in stats 62 / 63 (the getter returns it unchanged: ValShift 0), and the damage build (`FUN_0057b7d0` `0x57ba6c..0x57baac`, a plain monster attacker) rolls those stats with `FUN_0057a8e0`, which shifts by 8 again. The record's mana field is `(lo << 16) + draw` up to `hi << 16` in 256ths, which `FUN_0057c6c0` takes off the player's mana (`FUN_0057aa60`).

In play, a Normal level 9 unique with Mana Burn (`lo` 4, `hi` 7; MonLvl L-DM 7 x 66 / 100 and x 100 / 100) took 4 to 7 mana in d2d and takes 1,024 to 1,792 in game.exe, so a hit empties any normal mana pool. MonUMod row 25 can be picked for a unique in every difficulty (upick 6 in Normal, Nightmare and Hell); its minions' constants are 0 in Normal, so in Normal only the unique burns.

Two commits on upstream `ed007569`: the oracle `dc6d881a` (`tools/emu/mana_burn.py`, `tests/game/test_mana_burn.cpp`, a README row; it builds alone and prints what the parent does: its d2d side runs `monster_update` and then the real `Fight::hurt_side` on a player with 60,000 mana and prints the mana lost, not the field; `compare` exits 1 on a mismatch), then the claim `177189d6` (`rules::mana_burn_fine` in `components/rules/uniques.hpp`, its two uses in `ai.cpp` and `fight/tick.cpp`: `Foe::mana_burn` is now in 256ths; the `docs/research/re/monsters.md` finding; and the no-argument asserts in `test_mana_burn`, which ctest runs: game.exe's amounts for eight cases from the rule, and with the game data the player's mana after a burn through `Fight::hurt_side` and a Normal level 9 unique's hit end to end).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005a1f90` `0x5a2088..0x5a20c8` | `x = MonLvl DM x MonUMod constant / 100`, stored by the stat setter as `x << 8` in stats `0x3e` / `0x3f` |
| `FUN_00625480` | stats 62 / 63 have ValShift 0: setter 1,024 then getter 1,024 on a fake stat list |
| `FUN_0057b7d0` `0x57ba64..0x57baac` | a monster attacker (type 1, not a hireling), record flag bit 8 clear: `rec +0x3c += FUN_0057a8e0(0x3e, 0x3f)` scaled by the SrcDam byte / 128 |
| `FUN_0057a8e0` / `FUN_0057a880` | `(min << 8) + rand((max << 8) - (min << 8))` on the unit's seed |
| `FUN_0057c6c0` `0x57c879..0x57c87e` -> `FUN_0057aa60` | the player's mana (256ths) minus the record's `+0x3c`; under `0x100` left is 0 |

The game side is game.exe's own gate `FUN_0057dbf0` and resolver `FUN_0057d4f0` (record `+0 = 0`) under the emulator on a monster attacker with stats 62 / 63 = `lo << 8` / `hi << 8` and a player with 60,000 mana; the stat code is a dict that the checked identity of the real getter and setter justifies; events and the reach are hooked as in `vse_run.py`.

## Claims

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `dc6d881a` | the oracle and `test_mana_burn --mana-burn CASES` | 2,000 cases (8 ranges x 250 seeds); the parent d2d is outside game.exe's range in 1,750 of 2,000 through `Fight::hurt_side` | `mana_burn.py compare game.tsv d2d.tsv` prints `1750 ...`, exit 1 |
| `177189d6` | a hit takes `lo << 16` up to `hi << 16` in 256ths | claim 0 of 2,000 outside the range; game.exe equals its own rule on 2,000 of 2,000; the rule alone (`--mana-burn-roll`) equals game.exe's 2,000 of 2,000 byte for byte (`compare ... exact`) | `compare` prints `ok`, exit 0; `ctest -R test_mana_burn` passes |

Reseed: seeds 5001..5200 x 10 ranges, including the power-of-two bounds `(1, 2)` and `(0, 5)`, a fixed `(111, 111)` and Hell's `(60, 111)`: 2,000 of 2,000 in range and exact on the claim, 1,600 outside on the parent.

Mutants (local, never pushed; all abort the claim's no-argument `test_mana_burn`): one shift (`<< 8`), `<< 15`, no shift, the max stat for both, an inclusive bound (`rng(high - low + 1)`), `tick.cpp` keeping the old `<< 8`, `tick.cpp` subtracting points (`>> 8`), no clamp at 0.

Build and tests (V run, builder, Release, GCC): upstream main `ed007569` ctest 355 of 355; the claim 356 of 356 (the one added test is `test_mana_burn`).

## How to re-check

```sh
set -o pipefail
cd tools/emu
python3 mana_burn.py gen > cases.tsv
python3 mana_burn.py game cases.tsv > game.tsv
../../build/tests/test_mana_burn --mana-burn cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
python3 mana_burn.py compare game.tsv d2d.tsv          # claim: ok, exit 0; oracle commit alone: exit 1
../../build/tests/test_mana_burn --mana-burn-roll cases.tsv | grep -E '^[0-9]+ ' > roll.tsv
python3 mana_burn.py compare game.tsv roll.tsv exact   # the rule alone, byte for byte
cd ../..
ctest --test-dir build -R test_mana_burn
python3 tools/pr_guard.py
```

The d2d side needs `D2_MPQ_DIR` and `D2_PATCH_INSTALLER` (the test skips without 1.14d's MonStats). `compare` exits 1 on an empty or unequal file, a malformed line or cases out of step.

## Answers to the presubmit WARNs

- `siblings`: game.exe's `FUN_0057aa60` takes any defender's stat 8; d2d applies a burn to the player only (`Fight::hurt_side` reads `side[0]`), so a merc's mana is not burned, before or after; pets have no mana; a co-op partner is a player and takes the new amount.
- `build`, `gcc`: built and ctest run (Release, GCC): the same failures as upstream main (none: 355 of 355 there, 356 of 356 here).

## Limits

- The monster's own melee caller was not run; the chain starts at the gate with the record's first dword 0, as the existing swing oracles do. A skill-driven hit with the record bit `8` set (no mana leech) would take no burn in game.exe either way.
- Energy Shield: game.exe also takes the shield's calc1 % of the burn and pays `MulDiv(part, calc2, 16)` mana for it (`FUN_005ca6c0`: the table `0x6e30e0` lists the record's `+0x3c` with flag 1, taken only for a plain monster attacker, `0x5ca78d..0x5ca7ca`); d2d's shield works on the five damage fields only. Not changed here.
- `FUN_0057b7d0` has seven callers (the melee gate, `FUN_0056f070` and five skill functions); d2d burns only on a monster's melee swing. Whether a monster's missile or skill hit burns mana in game.exe is not run; nothing there is changed.
- The monster also heals by the mana it burns in game.exe (`FUN_0057c420`, game side V run); d2d has no heal. Out of scope: there is no d2d differential for it yet.
- Under 1 point of mana left after a burn is set to 0 in game.exe (`FUN_0057aa60`); d2d keeps the fraction. Not changed here.
- Not run: Windows, macOS, Clang; SrcDam other than 0x80.

## Checklist

- [x] A test in `tests/`: `tests/game/test_mana_burn.cpp`
- [x] The finding in `docs/research/re/`: `monsters.md` (MonUMod row 25)
- [x] `python3 tools/pr_guard.py` passes (0 FAIL, 0 WARN)

---
Agent: Claude Code
