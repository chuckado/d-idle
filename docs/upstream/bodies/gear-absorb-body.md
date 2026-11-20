## What changes

A player's gear absorb (stats 142 to 149) takes part of every elemental or magic field of a monster blow AND gives the absorbed amount back as life, as `FUN_0057bf10` and `FUN_0057a980` do. d2d read none of the eight stats, so an item or charm with an absorb property did nothing against a monster's fire, lightning, cold or magic damage.

**The reduction.** `FUN_0057c1e0` calls `FUN_0057bf80` on each row of the table at `0x732980`; for the fire, lightning, magic and cold rows `FUN_0057bf80` ends with `FUN_0057bf10` (`0x57c02b`, skipped when the record ignores resistances, `0x57c01e`). In order, on the field in 256ths:

1. the flat magic reduction (stat 35) and the resistance, as d2d already did;
2. the absorb percent (fire 142, lightning 144, magic 146, cold 148), at most 40: `MulDiv(field, percent, 100)` comes off (`0x57bf2a..0x57bf48`);
3. the absorb flat (143, 145, 147, 149), whole points `<< 8`: at most what is left comes off (`0x57bf4d..0x57bf6f`).

A field the first step left under 0 is not skipped: the percent moves it toward 0 and the flat clears it to 0 (a magic field of 2 points, magic reduction 3, absorb 36 percent: -256 becomes -164).

**The heal.** `FUN_0057bf10` adds what it takes to the record's `+0x48` (`0x57bf48`, `0x57bf6c`; nothing else in `0x579000..0x57e000` writes it). `FUN_0057c6c0` hands that total to `FUN_0057a980` (`0x57c844..0x57c84a`) BEFORE it takes the record total `+0x4c` off the life (`0x57c852..0x57c874`). `FUN_0057a980` (`0x57a980..0x57a9f0`): a total of 0 or less, a dead unit or state `0x5c`: nothing; else life = `min(life + total, max life)`. So an absorb lowers the hit and refills that much life, up to max, one record at a time. At full life the heal is capped away, so the player sees the same as a lower hit; below full life it is more: 20 percent cold absorb on a 100-point cold hit (after the resistance) takes 100 off the parent, 80 off a reduction-only change, 60 off in game.exe. A hit that would have killed can be survived, because the heal comes first.

**Code.** `rules::absorbed_fine` and `rules::elemental_took` (`combat.hpp`) are the reduction and what it took; `monster_blow` passes the defender's `Fighter::absorb_pct` / `absorb_flat` (new, per element type) and records the taken amount per field; `make_fighter` reads the eight stats; `StruckBlow::Part` carries both so the armors' recompute (Cyclone Armor, Energy Shield: `elemental_fine`) keeps the same order and the absorbed total follows the fields it leaves. `rules::absorb_heal` is the heal, run per blow in `Fight::hurt_side` before the totals come off (the cap is hit per blow, as game.exe applies one record at a time: two blows in a frame are two capped heals, not one summed heal). `d2s_items.hpp` names the stats.

**Players only.** game.exe's functions test no unit type, so a mercenary's absorb gear would act the same way (I: the hireling's item-stat transfer was not read). d2d keeps the merc's life in whole points with a fractional carry and no heal-back path, so `merc_fighter` zeroes the absorb: the merc takes none, before and after (no change for it). Pets have no gear. A co-op partner's hero is a player and takes the new rule.

Two commits on upstream `f1533770`: the oracle (`tools/emu/gear_absorb.py`, `tests/game/test_gear_absorb.cpp --life`, a README row; it builds alone and prints what the parent does: the player's life, through `monster_blow`, `Foe::take` and the real `Fight::hurt_side`), then the claim (the rules, the heal, `docs/research/re/combat.md`, and the tests below).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| table `0x732980`, stride `0x2c` | fire `0x10`: absorb stats `0x8e` / `0x8f`; lightning `0x1c`: `0x90` / `0x91`; magic `0x20`: `0x92` / `0x93`; cold `0x24`: `0x94` / `0x95`; the physical, poison, chill and freeze rows hold `-1` |
| `FUN_0057bf80` `0x57bfd0`, `0x57bfdc..0x57bff9`, `0x57c02b` | flat reduction, resistance (`FUN_0057be00`), then `FUN_0057bf10` |
| `FUN_0057bf10` `0x57bf2a..0x57bf48` | percent: 40 or more is 40, 0 or less is skipped, `MulDiv(field, p, 100)` added to `+0x48` |
| `FUN_0057bf10` `0x57bf4d..0x57bf6f` | flat: `stat << 8` (skipped when 0 or less), `min(flat, field)` (signed) added to `+0x48` |
| `FUN_0057c6c0` `0x57c844..0x57c84a` | `+0x48` to `FUN_0057a980` before the record total comes off |
| `FUN_0057a980` `0x57a980..0x57a9f0` | `min(life + total, FUN_00625d10)`; nothing for a total of 0 or less, a dead unit (`FUN_005541b0`) or state `0x5c` |
| `FUN_0057c6c0` `0x57c852..0x57c874` | life minus `+0x4c`; under `0x100` left is 0 |

## Counted differentials (V run, tree = upstream `f1533770` plus the branch)

Oracle: `gear_absorb.py game` runs `FUN_0057c6c0` natively on a player defender (a classic Normal game: no difficulty penalty; the cap is 75 plus the max stat) and prints the player's LIFE in 256ths after one to three blows (each blow one to three of fire, lightning, cold, magic; the defender's resistances, flat magic reduction, absorb percent and flat, a Cyclone Armor pool optional; the life from one point to full). `test_gear_absorb --life` prints the same from d2d's real `monster_blow`, `Foe::take` and `Fight::hurt_side` on a player with that life. Grid: type x field 1 to 400 x resistance -100 to 75 x absorb percent {0, 1 to 60, 40, 41} x flat 0 to 80 x magic reduction x pool x life.

| cases | differ before (oracle commit alone) | differ after (claim) |
|---|---|---|
| seed 7: 20,000 (7,923 with a Cyclone pool) | **9,639**, exit 1 | **0**, exit 0 |
| seed 5001, fresh: 20,000 (8,038 with a pool) | **9,519**, exit 1 | **0**, exit 0 |

Mutants of the claim (local, never pushed): each aborts `test_combat` or `test_gear_absorb_rules` or `test_gear_absorb`: no heal; heal after the total comes off (both `tick.cpp`: `test_gear_absorb` aborts on the case with the capped heal); a merc keeping its absorb (`merc.cpp`: the merc check aborts); flat before percent, percent and flat swapped, percent cap 41 (`test_combat` aborts; the first two also abort the rules test); magic and cold stats swapped in `make_fighter` (`test_combat`); the heal not capped at max life, a negative field healing, the total subtracted before the heal, the armors' recompute dropping the absorb, `Foe::take` dropping it (`test_gear_absorb_rules` aborts). 12 of 12 aborted by at least one test (none of them is the Energy Shield recompute: see Limits). The header mutants are compiled into the one translation unit of `test_gear_absorb_rules` (no library, so no unmutated inline copy can win) and into `test_combat`; the `.cpp` ones are real rebuilds of `tick.cpp` / `merc.cpp` and the test.

Build and tests (V run, GCC 14, Release, `-j2`, the builder): full build of the branch (1,183 targets) and `ctest -j2` 358 of 359 pass in one run; the one failure, `smoke_d2d`, is a window test that fails under the shared builder's load and passes rerun alone (78 s, `ctest -R smoke_d2d`): 359 of 359. 13 tests skip in this environment (their data), as on the same tree without the branch. `test_combat`, `test_gear_absorb` and `test_gear_absorb_rules` pass.

## In play

A hero with 20 percent cold absorb below full life: a 100-point cold hit after the resistance costs 60 life in game.exe (80 by the reduction alone, 100 on the parent); at 90 of 100 life it leaves 20 (the heal is capped at max life, then 80 comes off). Flat absorb (1 to 15 points on charms and rarer pieces) refills as much as it takes off. At full life the heal is capped away.

## Limits

- Only a player defender's blow through `monster_blow` (melee and missile hits that build their record there) and the armors' recompute. Not claimed: the burn path (`ai.hpp` burn), the blast of objects (`objects.cpp`), a monster's own element roll of a missile (`FUN_005a89a0`).
- Energy Shield: `energy_shield_blows` passes the absorb through the same `elemental_fine` call as Cyclone Armor's recompute, but no test runs it with an absorb (the oracle covers Cyclone Armor); not claimed beyond the shared function.
- A merc defender: game.exe applies the same functions to any unit (I); d2d's merc takes none here, as before.
- The ignore-resistances record (`0x57c01e`; flags `0x100`, `0x200`, `0x400` by attacker kind, `0x57c309..0x57c371`) skips the absorb in game.exe; d2d models no such record.
- A magic field's magic resistance (stat 37, `0x25`) is applied by game.exe on this path and not by d2d (decomp37 T3, not built here): a magic field with a non-zero magic resistance stays wrong.
- State `0x5c` (death_delay) blocks the heal in game.exe; d2d has no such state on the player (the dead check stands in).
- The oracle runs a classic Normal game (no penalty); the order of penalty and cap is the resistance's, unchanged. Absorb of a poison field does not exist in game.exe (the poison row holds `-1`).

## Claims

| Commit | Claim or change | Evidence | Re-check |
|---|---|---|---|
| `982a65c1` | the oracle and `test_gear_absorb --life CASES` | the parent's life differs from game.exe in 9,639 of 20,000 (seed 7) and 9,519 (seed 5001) | `gear_absorb.py compare` prints `# total: 20000 cases, 9639 differ`, exit 1 |
| `270b9799` | a player's absorb after the resistance (percent then flat), the heal-back per blow before the total, the armors' recompute carrying it, the merc left out | claim 0 of 20,000 twice; ten pinned lives equal game.exe's on the rules alone and through `Fight::hurt_side`; mutants abort | `compare` exit 0; `ctest --test-dir build -R 'test_combat\|test_gear_absorb' --output-on-failure` |

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_combat test_gear_absorb test_gear_absorb_rules
# gear_absorb.py game needs the emulator environment of the neighbouring tools/emu rows (D2_MPQ_DIR or d2d.cfg)
cd tools/emu && uv sync
set -o pipefail
uv run python gear_absorb.py gen 20000 7 > /tmp/c.tsv
uv run python gear_absorb.py game /tmp/c.tsv > /tmp/g.tsv
../../build/tests/test_gear_absorb --life /tmp/c.tsv | grep -E '^[0-9]+ -?[0-9]+$' > /tmp/d.tsv
uv run python gear_absorb.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent (oracle commit alone): 9639 differ, exit 1; claim: 0, exit 0 (also exit 1 on an empty file, unequal lengths, a malformed line or ids out of step)
cd ../..
ctest --test-dir build -R 'test_combat|test_gear_absorb' --output-on-failure    # D2_MPQ_DIR and D2_PATCH_INSTALLER set; test_gear_absorb skips without them, test_gear_absorb_rules needs no data
```

## Answers to the presubmit WARNs

- `siblings`: the diff touches the player's intake and `merc_fighter` (to leave the merc out). Merc: no change (its absorb is zeroed, as it was zero before); pets: no gear; monsters: carry no gear stats; co-op partner: a player, takes the rule. game.exe would apply it to a hireling too (I), which is why the Limits say so.
- `build`, `gcc`: built and ctest run (Release, GCC): full build of the branch (1,183 targets) and `ctest -j2` 358 of 359 pass in one run; the one failure, `smoke_d2d`, is a window test that fails under the shared builder's load and passes rerun alone (78 s, `ctest -R smoke_d2d`): 359 of 359. 13 tests skip in this environment (their data), as on the same tree without the branch. `test_combat`, `test_gear_absorb` and `test_gear_absorb_rules` pass.

---

Agent: Claude Code
