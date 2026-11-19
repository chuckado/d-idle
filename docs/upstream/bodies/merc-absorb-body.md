## What changes

A mercenary's gear absorb (stats 142 to 149) takes part of every elemental or magic field of a monster blow and heals the merc by what it took, as it does for a player (#606). `FUN_0057c6c0` tests no player type: `FUN_0057c1e0` -> `FUN_0057bf80` -> `FUN_0057bf10` read the struck unit's stats and `FUN_0057a980` heals it, whatever the unit is. d2d zeroed the merc's absorb stats (`merc.cpp`: "the absorb without the heal on the merc's life would be half the rule"), so absorb gear on a hireling did nothing.

- `Fight::merc_fighter` keeps the absorb stats `make_fighter` reads from the merc's worn gear (its resistances already came the same way).
- `Fight::hurt_side` gives the merc the heal-back before the blows' totals come off its life: `rules::absorb_heal` over the merc's blows, in 256ths of `life_fine()`, capped at `merc_max_life() << 8`, a blow at a time. Nothing else in the merc's life handling changes (its fraction carry, the dead-under-one-point rule, the crush after the hits, the get-hit).
- `test_gear_absorb` no longer asserts that the merc takes none; `test_merc_absorb` pins game.exe's lives for twelve cases.

Two commits: the oracle (`tools/emu/merc_absorb.py`, `tests/game/test_merc_absorb.cpp --life`, a README row), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0057c6c0` `0x57c6c0..0x57c844` | the strike's gate tests the unit (`FUN_00554200`, `FUN_00620bb0`) and, for a type-1 unit, one MonStats flag of its class row (`FUN_00457490`, bit 0xf); no test names a player |
| `FUN_0057c1e0` -> `FUN_0057bf80` -> `FUN_0057bf10` | the same absorb as the player's (#606) |
| `FUN_0057a980` `0x57a980..` | `life = min(life + amount, max life)`; nothing when the amount is 0 or the unit is dead (`FUN_005541b0`) or in state `0x5c` |

## Counted differentials (V run)

`merc_absorb.py game` boots game.exe with every table loaded (`drlg.boot`: the hireling classes 271, 338, 359, 560 and 561 have their real MonStats rows) and runs `FUN_0057c6c0` on a unit of type 1 of one of those classes, once per blow, printing the unit's life; the MonStats readers (`FUN_00457490`, `FUN_0063e9f0`, `FUN_0063ee90`, `FUN_0063e990`, `FUN_0063e940`) run for real. The same cases on a type-0 unit through `gear_absorb.py game`: the lives are equal in 300 of 300. `test_merc_absorb --life` prints the merc's life from d2d's real `merc_foe`, `monster_blow`, `Foe::take` and `Fight::hurt_side`, the absorb on a worn torso piece.

| | seed 7 | seed 5001 | seed 9091 |
|---|---|---|---|
| parent, 6,000 cases | 3,075 differ | 3,056 | 3,057 |
| claim | 0 | 0 | 0 |

The pins (12 cases) and the compare catch 11 of 11 mutants (no heal, uncapped heal, cap in points, whole-point life, first blow only, percent or flat dropped, fire and cold swapped, negative heal dropped, two-blow heal, the parent). Build and the full ctest: not run yet (waiting for the builder); `presubmit --quick` 0 FAIL, `pr_guard` 0 / 0.

## In play

A Desert Guard at 90 of 100 life with 20 percent cold absorb on its armor, hit by a 100-point cold blow: game.exe heals it to 100, then takes 80 (20 left); the parent took the whole blow and the merc died. The merc's life regeneration, its potions and its aura are untouched.

## Limits

- That a merc's worn item puts its stats on the merc's stat list as it does for resistances is I: d2d's `merc_gear` already sums every property of the worn items into the stats `merc_fighter` reads (resistances, defense, life, to-block), and game.exe has one stat-list path for any unit with an inventory; the equip path itself was not run. The differential starts from the stat list.
- The ignore-resistances record (`0x57c01e`), state `0x5c` on the heal, a magic field's magic resistance (stat 37: decomp37 T3) and the burn and blast paths are as in #606.
- Pets and summons carry no worn gear in d2d.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cd build && ninja test_merc_absorb test_gear_absorb && cd ../tools/emu && uv sync
set -o pipefail
uv run python merc_absorb.py gen 20000 7 > /tmp/c.tsv
uv run python merc_absorb.py game /tmp/c.tsv > /tmp/g.tsv
../../build/tests/test_merc_absorb --life /tmp/c.tsv | grep -E '^[0-9]+ -?[0-9]+$' > /tmp/d.tsv
uv run python merc_absorb.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: differs, exit 1; claim: 0, exit 0
../../build/tests/test_merc_absorb && ../../build/tests/test_gear_absorb
```

---

Agent: Claude Code
