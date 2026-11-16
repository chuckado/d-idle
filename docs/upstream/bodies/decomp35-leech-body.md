## What changes

A player's life and mana leech works as FUN_0057c420 does. d2d computed `physical (whole points) x steal x Drain / 10000` in whole points, with no cap and no difficulty term; game.exe does three other things:

1. **256ths.** FUN_0057c420 shifts the steal left by 6, takes `MulDiv(physical, steal', 100)` of the record's physical in 256ths, `MulDiv(.., Drain, 100)` unless Drain is 100, and divides by 64 toward zero (`0x57c47d..0x57c556`). FUN_0057a980 and FUN_0057aa00 add that many 256ths of life and mana. The whole-point floor made a 6 percent leech on a 10 hp blow heal 0.
2. **The difficulty's divisor.** For a player attacker (`[attacker] == 0`, `0x57c494`) the shifted steal is divided by DifficultyLevels `LifeStealDivisor` (row `+0x28` of FUN_00611d30) or `ManaStealDivisor` (`+0x2c`): 1 / 2 / 3 in Normal, Nightmare, Hell (`0x57c4b8..0x57c4ce`). d2d had none, so a leech hero healed twice as much in Nightmare and three times in Hell.
3. **The physical is cut to the target's life.** FUN_0057c6c0 loads the record's physical, reads stat 6 of the defender and stores the smaller back before calling FUN_0057c420 (`0x57c7f8..0x57c826`). A killing blow leeches only the life that was left.

`rules::player_blow` now calls a new `leech_fine` (`components/rules/combat.hpp`) with the physical after resistance in 256ths; `Blow::life` and `Blow::mana` hold 256ths and `Fight::land` adds them as they are (`strike.cpp`). `Target` gets `life_fine` (the target's life in 256ths, -1: `hit_points x 256`) and the two divisors, filled from `DifficultyLevels` in `gamedata_load.cpp` and `Monster::target`. Smite does not leech (its record is marked built), as before. `test_combat` pinned the old whole-point value (`50 x 10 x 50 / 10000`); it now pins 640 (256ths).

Two commits: the oracle (`tools/emu/leech.py`, `leech_harness.cpp`, a README row), then the claim.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0057b7d0` `0x57bb00..0x57bb21` | a player attacker's stats 60 and 62 go into the record at `+0x38` and `+0x3c`, unscaled |
| `FUN_0057c6c0` `0x57c7f8..0x57c826` | the record's physical `+8` is cut to the defender's stat 6, then FUN_0057c420 |
| `FUN_0057c420` `0x57c466..0x57c474` | the monster's Drain of the game's difficulty (byte `row + 0xa0 + difficulty`); 0 or less: no leech |
| `FUN_0057c420` `0x57c47d..0x57c4ce` | steal `<< 6`, then `/ LifeStealDivisor` or `/ ManaStealDivisor` for a player attacker |
| `FUN_0057c420` `0x57c4e6..0x57c556` | `MulDiv(physical, steal', 100)`, `MulDiv(.., Drain, 100)` unless 100, `/ 64` toward zero, then FUN_0057aa00 (mana) and FUN_0057a980 (life) |
| `FUN_00483360` | the MulDiv those lines use (already ported as `mul_div`, `rules.hpp`) |

## Counted differentials (V run, tree = upstream `b6e6940d` plus the branch)

Oracle: `leech.py game` runs FUN_0057c6c0 on game.exe natively (a player attacker, a monster defender of class 60, record flags apply + hit, the case's physical, steal, Drain and difficulty, the defender's life as stat 6) and reads the amounts FUN_0057a980 and FUN_0057aa00 receive. `leech.py compare` sets them against `leech_harness`, which runs d2d's real `player_blow` (the weapon's physical is the case's, the seed is advanced to a hit) and prints what `Fight::land` credits. Grid: 3 difficulties x 13 steals x 8 physical sizes (down to a fraction of a point), overkill rows (target life 1 to 10 points), Drain rows {0, 1, 25, 50, 75, 100, 125}, then random.

| cases | differ before (parent) | differ after |
|---|---|---|
| seed 7: 20,351 | **17,097** | **0** |
| seed 1: 60,351 | **50,626** | **0** |

Self-check (`leech.py selfcheck`: the rule read from the disassembly, run against game.exe's column): 0 of 20,351 and 0 of 60,351 differ.

Mutants of the claim, seed 7 (each is caught by the compare): no divisor 11,620 differ, no cut to the target's life 3,133, whole points 17,068, no Drain term 8,376.

Build and tests (V run, GCC 14, Release, on the builder): `presubmit.py --quick --base upstream/main` 0 FAIL, `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, `test_combat` ok, `test_game` rc 0. The full ctest result is in the card result of the branch's task.

## In play

hp healed per landed blow (Drain 100), game.exe against the parent: a 5 percent steal on a 10 hp blow 0.50 against 0.00; an 8 percent steal on 20 hp 1.60 against 1.00; a 20 percent steal on 40 hp 8.00 in Normal, 4.00 in Nightmare, 2.66 in Hell against 8.00 on the parent in all three.

## Limits

- Only the player's own blows leech (`by_player`), as before: a mercenary, a pet or a monster attacker does not leech in d2d (game.exe scales a monster's steal by the SrcDam share at `0x57ba64`; not claimed).
- Where a caller builds a `Target` by hand (no `life_fine`), the life is `hit_points x 256`.
- Not claimed: elemental or crushing damage is never leeched in either build; Open Wounds, crit and deadly strike rolls.

## Answers to the presubmit WARNs

- Unit kinds: the diff touches the player's blow only. In game.exe the divisor is applied when the attacker is a player (`0x57c494`); the merc and pets do not credit a leech in d2d (`land` adds it `by_player` only), so there is no sibling path to change.
- Build: `--quick` skips it; the build and ctest numbers are above.

## How to re-check

```sh
python3 tools/pr_guard.py upstream/main HEAD
cmake --build build --target test_combat test_game
cd tools/emu && uv sync
set -o pipefail
uv run python leech.py gen 20000 7 > /tmp/c.tsv
uv run python leech.py game /tmp/c.tsv > /tmp/g.tsv
g++ -O2 -std=c++23 -I../../components/rules -I../../components -I../../components/d2s -I../../components/txt \
    -o /tmp/leech_harness leech_harness.cpp && /tmp/leech_harness < /tmp/c.tsv > /tmp/d.tsv
uv run python leech.py selfcheck /tmp/c.tsv /tmp/g.tsv   # 0 of 20351 differ
uv run python leech.py compare /tmp/c.tsv /tmp/g.tsv /tmp/d.tsv   # parent: 17097 differ; claim: 0, exit 0
../../build/tests/test_combat
```

---

Agent: Claude Code
