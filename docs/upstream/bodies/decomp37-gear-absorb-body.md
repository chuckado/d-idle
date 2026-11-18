## What changes

A player's gear absorb (stats 142 to 149) takes part of every elemental or magic field of a monster blow, as `FUN_0057bf10` does. d2d read none of the eight stats, so an item or charm with an absorb property did nothing against a monster's fire, lightning, cold or magic damage.

`FUN_0057c1e0` calls `FUN_0057bf80` on each row of the table at `0x732980`; for the fire, lightning, magic and cold rows `FUN_0057bf80` ends with `FUN_0057bf10` (`0x57c02b`, skipped when the record ignores resistances, `0x57c01e`). In order, on the field in 256ths:

1. the flat magic reduction (stat 35) and the resistance, as d2d already did;
2. the absorb percent (fire 142, lightning 144, magic 146, cold 148), at most 40: `MulDiv(field, percent, 100)` comes off (`0x57bf2a..0x57bf48`);
3. the absorb flat (143, 145, 147, 149), whole points `<< 8`: at most what is left comes off (`0x57bf4d..0x57bf6f`).

A field the first step left under 0 is not skipped: the percent moves it toward 0 and the flat clears it to 0 (a magic field of 2 points, magic reduction 3, absorb 36 percent: -256 becomes -164).

`rules::absorbed_fine` (`combat.hpp`) is that function; `elemental_fine` takes the two numbers and calls it after the resistance; `monster_blow` passes the defender's `Fighter::absorb_pct` / `absorb_flat` (new, per element type); `make_fighter` reads the eight stats; `StruckBlow::Part` carries them so the armors' recompute (Cyclone Armor, Energy Shield: `elemental_fine`) keeps the same order. `d2s_items.hpp` names the stats.

Two commits: the oracle (`tools/emu/gear_absorb.py`, `gear_absorb_harness.cpp`, a README row), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| table `0x732980`, stride `0x2c` | fire `0x10`: absorb stats `0x8e` / `0x8f`; lightning `0x1c`: `0x90` / `0x91`; magic `0x20`: `0x92` / `0x93`; cold `0x24`: `0x94` / `0x95`; the physical, poison, chill and freeze rows hold `-1` |
| `FUN_0057bf80` `0x57bfd0`, `0x57bfdc..0x57bff9`, `0x57c02b` | flat reduction, resistance (`FUN_0057be00`), then `FUN_0057bf10` |
| `FUN_0057bf10` `0x57bf2a..0x57bf48` | percent: 40 or more is 40, 0 or less is skipped, `MulDiv(field, p, 100)` |
| `FUN_0057bf10` `0x57bf4d..0x57bf6f` | flat: `stat << 8` (skipped when 0 or less), `min(flat, field)` taken (signed) |

## Counted differentials (V run, tree = upstream `ed007569` plus the branch)

Oracle: `gear_absorb.py game` runs `FUN_0057c1e0` natively on a player defender (a classic Normal game: no difficulty penalty; the cap is 75 plus the max stat) with one element field in the record, the defender's resistance, flat magic reduction and absorb stats set through the stat getter, and prints the field and the record total (`+0x4c`) in 256ths. `gear_absorb_harness` runs d2d's real `monster_blow` on an `exact` record (one El row of min == max, no physical). Grid: type (fire, lightning, cold, magic) x field 1 to 400 x resistance -100 to 75 x absorb percent {0, 1 to 60, 40, 41} x flat x magic reduction (including past the field).

| cases | differ before (parent) | differ after |
|---|---|---|
| seed 7: 20,000 (16,607 with an absorb) | **13,189** | **0** |
| seed 1: 60,000 (50,062 with an absorb) | **39,678** | **0** |

The cases without an absorb are equal on the parent too (3,393 and 9,938): the existing reduction and resistance order is unchanged.

Mutants of the claim (each is caught by the compare), seed 7: no cap at 40 2,271 differ, flat before the percent 1,556, flat not limited to the field 7,983, negative fields skipped 4,883, absorb before the resistance 1,826. `test_combat` pins game.exe's fields (the resistance alone 19,200, with 20 percent 15,360, the percent cap and the flat 5,120, a flat past the field 0, the flat after the resistance 7,680, a negative field -164 and cleared 0, and the stat to element mapping of `make_fighter`); each of the five mutants aborts on its own assert.

Build and tests (V run, GCC 14, Release, `-j2`, the builder): `test_combat` ok, `test_game` rc 0. Full ctest: see the card result of the task (run on a tree holding this branch and the Open Wounds one).

## In play

A hero with 20 percent cold absorb takes 20 percent less of a cold hit that got through the resistance in game.exe; on the parent the same hit is unchanged. Flat absorb (1 to 15 points on charms and rarer pieces) takes a hit's whole field down to 0 on small hits.

## Limits

- Only a player defender's blow through `monster_blow` (melee and missile hits that build their record there) and the armors' recompute. Not claimed: the burn path (`ai.hpp` burn), the blast of objects (`objects.cpp`), a monster's own element roll of a missile (`FUN_005a89a0`), a merc or pet defender (their fighters carry no gear stats in d2d).
- The absorbed amount game.exe adds to the record (`+0x48`) is not read by anything d2d models.
- The oracle runs a classic Normal game (no penalty); the order of penalty and cap is the resistance's, unchanged.
- Absorb of a poison field does not exist in game.exe (the poison row holds `-1`).

## Answers to the presubmit WARNs

- Unit kinds: the diff touches the player's intake only; the merc and pets are monsters in `Foe` with default (zero) absorb, so nothing changes for them.
- Build: `--quick` skips it; the numbers are above.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_combat test_game
# gear_absorb.py game needs the emulator environment of the neighbouring tools/emu rows (D2_MPQ_DIR or d2d.cfg)
cd tools/emu && uv sync
set -o pipefail
uv run python gear_absorb.py gen 20000 7 > /tmp/c.tsv
uv run python gear_absorb.py game /tmp/c.tsv > /tmp/g.tsv
g++ -O2 -std=c++23 -I../../components/rules -I../../components -I../../components/d2s -I../../components/txt \
    -o /tmp/gear_absorb_harness gear_absorb_harness.cpp && /tmp/gear_absorb_harness < /tmp/c.tsv > /tmp/d.tsv
uv run python gear_absorb.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: 13189 differ, exit 1; claim: 0, exit 0 (also exit 1 on an empty file, unequal lengths, a malformed line or ids out of step)
../../build/tests/test_combat
```

---

Agent: Claude Code
