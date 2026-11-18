## What changes

A player's melee swing starts, and its blow lands, only where game.exe's reach test allows. d2d's `Fight::engage` started the swing within 1.1 cells (5.5 subtiles, centre to centre) and `Fight::hit` landed within 1.6 cells (8 subtiles), with nothing between the units tested, so a Paladin swung at and hit monsters on the other side of a wall.

1. **The test.** The client ends its walk to a monster (`FUN_00480c80` at `0x480d0f`) and the server starts the swing (`FUN_00549ba0` at `0x549cc8`) when `FUN_00622c40(player, monster, FUN_00622dc0(monster))` holds, and the blow's to-hit (`FUN_0057ec10` at `0x57ec7e..0x57eca6`) asks `FUN_00622c40(player, monster, 1)` first (a false answer is no hit and no roll). `FUN_00622c40`: a monster of class 0x102 / 0x105 is in reach under range + 8; the unit distance `FUN_00641530` at 0 or less is in reach; else it needs `range + extra + 1 >= distance` and no collision word with a bit of mask 0x804 on the line (`FUN_00622aa0`). `rules::melee_reach` and `rules::unit_distance` (both already in the tree for NPCs) are that function.
2. **What the fix does.** `Fight::melee_reach_to(monster, blow)` asks it with the unit distance of the player (size 2) and the monster's SizeX, the equipped weapon's range (`weapon_range()`), `extra` 1 for a moving monster (a start) or always (a blow), and the line over `Level::blocked(.., 0x804)`. `engage` and the default `hit` use it. Whirlwind, Charge and Leap Attack keep the reach they had (their own landing is not traced).
3. **What stays.** Two cases game.exe allows stay allowed: units whose sizes overlap (`across + down < size1 + size2`, up to 3 subtiles for two size 2 units) have no line to cut, and a gap in a diagonal wall lets the line through.

Two commits on upstream `895f8a67`: the oracle (`tools/emu/player_reach.py`, `tests/game/test_player_reach.cpp`, a README row), then the claim (code, 16 pinned game.exe cases in the test, a note in `docs/research/re/combat.md`).

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_00480c80` `0x480d06..0x480d1f` | the client keeps walking until `FUN_00622c40(player, target, FUN_00622dc0(target))` holds (any target but an object) |
| `FUN_00549ba0` `0x549cbf..0x549cc8` | the server's start asks the same for a melee skill on a unit; false sends the player walking (`FUN_00548a50`) |
| `FUN_00622c40` | tentacle class under range + 8; apart at 0 or less in reach; else `range + extra + 1 >= apart` and `FUN_00622aa0(.., 0x804)` clear |
| `FUN_0057ec10` `0x57ec7e..0x57eca6` | `extra = 1 + 2 * (attacker is a monster)`; `FUN_00622c40` false returns no hit |
| `FUN_0056f070` `0x56f115..0x56f1e9` | the do of Attack hits the target `FUN_00553540` returns; no other distance test |

## Counted differentials (V run, tree = upstream `895f8a67` plus the branch)

Oracle: `player_reach.py game` runs `FUN_00622c40` and `FUN_00622dc0` for real over a random wall map (the room's collision words, `FUN_00641530` and the line `FUN_00622aa0` -> `FUN_0064e260` real), `player_reach.py hit` runs `FUN_0057ec10` with its rolls fixed; `test_player_reach --player-reach` / `--player-hit` print what the real `Fight::engage` / `Fight::hit` do on the same cases.

| cases | start: parent | start: claim | blow: parent | blow: claim |
|---|---|---|---|---|
| seed 1 (`gen 20000 1`): 30,584 | 2,799 | **0** | 9,869 | **0** |
| seed 7 (`gen 60000 7`): 70,584 | 7,210 | **0** | 24,094 | **0** |

The cases: every offset within 10 subtiles x monster size 1, 2, 3 x standing or walking x weapon range 0..3 with no wall, then random pairs within 9 subtiles with a wall (a line of 1 to 5 subtiles, horizontal, vertical or diagonal, or a 2 x 2 pillar) and scattered cells, words 4, 0x800, 0x804 (cut) and 1, 0x400 (not in the mask).

On the parent, d2d starts through a wall in 1,546 of the 1,661 cases where game.exe's distance would allow it and only a wall stops it, and its blow lands in all 2,047 such cases; most of the rest is the distance (game.exe's reach for two size 2 units is about 3 subtiles, not 5.5).

Mutants of the claim (local, never pushed), each aborted by `test_player_reach`:

| Mutant | Pin that fails |
|---|---|
| the start back on `hypot <= 1.1` | 1 |
| the line test off | 2 |
| mask 0x4 instead of 0x804 | 5 |
| the moving extra left off the start | 3 |
| the blow's extra left off | 1 |
| the monster's size read as 2 | 15 |
| the blow on the old 1.6 cells | 2 |
| the weapon range read as 0 | 11 |

`test_player_reach` runs 16 pinned cases from game.exe's output without the emulator; the grid itself needs it.

Build and tests (V run, GCC, Release, the builder): the 19 test binaries that touch `Fight` build and pass or skip (the oracle-mode ones) on the claim, `test_zeal_chain` and `test_game` among them. The whole `ctest` was not run by the author (a full build did not finish on the shared builder); the verifier's job.

## In play

The swing starts at about 3 subtiles from a standing size 2 monster (4 while it walks) instead of 5.5; a wall of mask 0x804 on the line between two units 4 or more subtiles apart now stops both the start and the blow. A player clicking a monster behind a wall walks round (or fails to, as in game.exe) instead of swinging at it.

## Limits

- Monsters' melee on the player and the pets (`FUN_0057d4f0` asks the same test with `extra` 3) is not changed: d2d's `kMeleeReach + 0.3` hypot stays (not run).
- Whirlwind, Charge and Leap Attack keep their reach (the same blow test applies to them in game.exe; not run).
- Which packets set the server's flag in `FUN_00549ba0` was not decoded; the client's walk end alone gives the same decision in a single-player game.
- A vulture's moving rule (BaseId 0x6e rows) is passed only for the row `mon_bin[0x6e]`.
