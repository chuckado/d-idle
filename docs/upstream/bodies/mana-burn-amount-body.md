## What changes

A Mana Burn monster's landed melee hit takes `lo * 256` up to `hi * 256` mana points off the player in game.exe, where `lo` / `hi` are the whole points d2d keeps in `Monster::mana_lo` / `mana_hi`; d2d took `lo` to `hi` points. `FUN_005a1f90` stores the burn as `x << 8` in stats 62 / 63 (the getter returns it unchanged: ValShift 0), and the damage build (`FUN_0057b7d0` `0x57ba6c..0x57baac`, a plain monster attacker) rolls those stats with `FUN_0057a8e0`, which shifts by 8 again. The record's mana field is `(lo << 16) + draw` up to `hi << 16` in 256ths, which `FUN_0057c6c0` takes off the player's mana (`FUN_0057aa60`).

In play, a Normal level 9 unique with Mana Burn (`lo` 4, `hi` 7) took 4 to 7 mana in d2d and takes 1,024 to 1,792 in game.exe, so a hit empties any normal mana pool.

Two commits: the oracle (`tools/emu/mana_burn.py`, `tests/game/test_mana_burn.cpp`, a README row; it builds alone, prints what the parent does, and `cmp` exits 1 on a mismatch), then the claim (`rules::mana_burn_fine` in `components/rules/uniques.hpp`, its two uses in `ai.cpp` and `fight/tick.cpp`: `Foe::mana_burn` is now in 256ths).

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
| `a6093f9d` | the oracle and `test_mana_burn --mana-burn CASES` | 2,000 cases (8 ranges x 250 seeds); 1,750 differ on this commit alone (d2d outside game.exe's range); reseed 1,800 of 2,000 | `mana_burn.py cmp game.tsv d2d.tsv` prints `1750 ...`, exit 1 |
| `d7a97516` | a hit takes `lo << 16` up to `hi << 16` in 256ths | 0 of 2,000 and 0 of 2,000 outside the range; the rule alone (`--mana-burn-roll`) equals game.exe's 2,000 of 2,000 byte for byte on each set | `cmp` prints `ok`, exit 0, and `cmp ... exact` the same |

Mutant (local, caught): `<< 15` instead of `<< 16` leaves 1,750 of 2,000 cases differing in `cmp ... exact`, exit 1.

## How to re-check

```sh
set -o pipefail
cd tools/emu
uv run python mana_burn.py gen > cases.tsv
uv run python mana_burn.py game cases.tsv > game.tsv
../../build/tests/test_mana_burn --mana-burn cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
uv run python mana_burn.py cmp game.tsv d2d.tsv          # claim: ok, exit 0; oracle commit alone: exit 1
../../build/tests/test_mana_burn --mana-burn-roll cases.tsv | grep -E '^[0-9]+ ' > roll.tsv
uv run python mana_burn.py cmp game.tsv roll.tsv exact   # the rule alone, byte for byte
cd ../..
python3 tools/pr_guard.py
```

## Limits

- The monster's own melee caller was not run; the chain starts at the gate with the record's first dword 0, as the existing swing oracles do. A skill-driven hit with the record bit `8` set (no mana leech) would take no burn in game.exe either way.
- The monster also heals by the mana it burns in game.exe (`FUN_0057c420`, game side V run); d2d has no heal. Not part of this change (no d2d differential yet).
- Under 1 point of mana left after a burn is set to 0 in game.exe (`FUN_0057aa60`); d2d keeps the fraction. Not changed here.
- The whole test suite was not run on this branch (the new test only).

---
Agent: Claude Code
