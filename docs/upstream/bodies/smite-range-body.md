## What changes

Smite hits with the shield's damage. FUN_005ce9f0 calls FUN_0057b420 with the shield as the item (`edx = 0`), the shield's `mindam` and `maxdam` shifted left by 8 plus Holy Shield's own range, the skill's ED and SrcDam 0x80 (`0x5ceb31`). On that path FUN_0057b420 does two things before the percents that d2d did not:

1. It adds stat 111 (`item_normaldamage`, "dmg" in the item properties) shifted left by 8 to both ends (`0x57b4c3..0x57b4d6`), and makes a minimum below 1 (in 256ths) one point (`0x57b4d8..0x57b4e0`). `Fighter::smite_lo` and `smite_hi` took the shield's row alone, while the plain weapon swing already adds the unit's stat 111.
2. When the maximum is not above the minimum it makes the maximum the minimum plus one point (`0x57b4ea..0x57b4f4`: `cmp [max], min; jg; add min, 0x100`). A shield whose `maxdam` equals its `mindam` (the Zakarum Shield, 46 and 46) therefore rolls over a range, not a fixed value; the plain swing has the same rule in `make_fighter` (`high = max(high << 8, low + 256)`).

Both are in `make_fighter` (the shield block) and in `player_blow`'s Smite branch (`components/rules/combat.hpp`). `tests/test_combat.cpp` pinned a shield 10 to 10 at 15 points; it now allows 15 or 16 (the range is 15 to 16.5).

Two commits: the oracle (`tools/emu/smite_damage.py`, `tests/test_smite_damage.cpp --smite-damage`, README row; it prints what `player_blow` rolls on the parent), then the claim (the two rules, the changed `test_combat` bounds, 28 frozen cells and a range check in the test).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_005ce9f0` `0x5ceb31` | the call: FUN_0057b420 with the shield, min and max (<< 8) plus Holy Shield's, ED = calc1, SrcDam 0x80 |
| `FUN_0057b420` `0x57b4c3..0x57b4f4` | stat 111 << 8 added to min and max; min < 1 becomes 0x100; max <= min becomes min + 0x100 |
| `FUN_0057b420` `0x57b4f7..0x57b5c5` | the percent: stat 25 + the item's Str and Dex bonus (FUN_00629860 / FUN_006298a0) + the ED, floored at -90, as MulDiv on each end; the roll is FUN_0045c3e0 on the unit's seed |

## Counted differentials (V run, tree = upstream `d1a10a38` plus the branch)

Oracle: `smite_damage.py compare` loads the seed's two words into a unit and runs FUN_0057b420 natively with the case's inputs for each cell `test_smite_damage --smite-damage` prints, and compares game.exe's roll with d2d's `player_blow(swing.smite)` physical in 256ths, whole. The grid: 7 shield rows (1..2, 5..5, 46..46, 9..30, 12..12, 1..1, 20..60) x Holy Shield none or 3..9 points x ED {0, 15, 45, 150, 300} x stat 111 {-30, 0, 1, 2, 5} x stat 25 {0, 20, 100} x Str / Dex {30 / 30, 100 / 80} x the shield's bonuses {110 / 0, 0 / 100, 75 / 75, 0 / 0} x gear stats 17 and 18 {0, 15, 40} x 3 seeds = 75,600 cells. No percent is negative (a negative percent's rounding and the -90 floor are other rules, not claimed).

| 75,600 cells | differ before (parent) | differ after |
|---|---|---|
| stat 111 = 0, min < max | 0 | 0 |
| stat 111 = 0, min = max | **4,315** of 4,320 | **0** |
| stat 111 != 0 | **42,116** of 43,200 | **0** |
| stat 111 != 0, min = max | **17,278** of 17,280 | **0** |
| all | **63,709** | **0** |

Mutant (V run): dropping the maximum rule (`base_high = base_low + 256`) aborts `test_smite_damage` at a frozen cell. The stat 111 term alone was not mutated separately (the 42,116 differing cells above are the check).

Build and tests (V run, GCC 14, Release, on the builder): `presubmit.py --quick --base upstream/main` 0 FAIL (4 WARN, answered below), `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, clang-tidy-19 on the 265 translation units that include a changed file: 0 findings in the changed files (5 `misc-include-cleaner` errors for `std::erase_if` in five other files this branch does not touch). Full ctest: parent 5 of 326 fail (`test_outdoor`, `smoke_d2d`, `touch_controls_d2d`, `touch_corpse_d2d`, `pad_use_d2d`), the claim the same 5 of 327. `test_combat` passes. Linux GCC only (U for other platforms).

## Limits

- Reach: stat 111 sits on three uniques and one runeword in the patched tables (`item_normaldamage`, none a magic affix), and the only vanilla shield with `maxdam == mindam` is the Zakarum Shield. The change is small in play; the claim is the rule, not a dps effect.
- Whether a stat 111 of the other weapon set counts while Smite is used (the hand toggling) is not claimed: the claim reads the player's summed stats as the plain swing does.
- Not claimed: the -90 floor (it applies to the sum in game.exe, to the gear part first in d2d; no gear gets there), the rounding of a negative percent, the to-hit roll (game.exe draws one and forces the hit; d2d draws none: the same outcome, a different random stream), Smite without a shield (game.exe does nothing; d2d swings with 0).

## Answers to the presubmit WARNs

- Hit paths (melee, skill, missile): the kick branch of `player_blow` has its own damage rule (FUN_00646280) and is not touched; Smite is the only caller of `swing.smite`; a missile has no shield damage.
- Data gating: `test_smite_damage` reads no game data (it builds the shield's row by hand); the word is in a comment about armor.txt.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_smite_damage test_combat
cd tools/emu && uv sync
set -o pipefail
../../build/tests/test_smite_damage --smite-damage > /tmp/smite.txt
uv run python smite_damage.py compare /tmp/smite.txt    # last line: 75600 cells  same (parent: 63709 differ)
../../build/tests/test_smite_damage
../../build/tests/test_combat
```

---

Agent: Claude Code
