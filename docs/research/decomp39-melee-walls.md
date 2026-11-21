# decomp39: can a player's melee hit through a wall? game.exe 1.14d vs d2d vs the bot (and the DoT sub-point death rule)

> status: current (2026-11-17, card decomp39, TASK-624, agent d2-decomp on Sonnet). Offline: no game, no d2d slot, no Go changed, nothing posted upstream. The game.exe side ran under the emulator (`docs/emu-image.md`) and the d2d side as fork builds, both on the builder container (nothing on the desktop worker, nothing heavy on the head box). Triggered by the user's observation of 2026-11-17: the sandbox Paladin (Zeal and Attack, d2d lineage) hit monsters on the other side of walls, cliffs and cave walls, more than once. Follows [decomp35-combat.md](decomp35-combat.md), [dc36-lane-b.md](dc36-lane-b.md) and [sk-melee-trace.md](sk-melee-trace.md) (its Q1 names the reach test `FUN_00622c40` as read, not run on a player's swing). Scratch tooling here: `tools/emu/dc39_reach.py` (the fork branch carries it as `player_reach.py`) and `tools/emu/dc39_dot.py`.

Tags: **V run** = game.exe's own code under the emulator, compared case by case with a d2d build of the real `Fight` (the numbers are counts of cases); **V read** = disassembly read (`FUN_` addresses of 1.14d `Game.exe`); **S** = an earlier proof of ours or upstream's, named; **I** inferred; **U** unverified.

## Answer in eight lines

1. **game.exe does not let a player's melee go through a wall, with two authentic exceptions.** The swing starts, and the blow lands, only when `FUN_00622c40(player, monster, extra)` holds: the unit distance (`FUN_00641530`) within the weapon's range + extra + 1 (always at 0 or less), **and no collision word with a bit of mask 0x804 on the line between the two units** (`FUN_00622aa0`, the sight function upstream already proved). `extra` is 1 while the monster moves for the start (`FUN_00622dc0`), always 1 for the player's blow (`FUN_0057ec10` at `0x57eca6`). V read; `FUN_00622c40` and the blow's chain V run.
2. **The exceptions (authentic, V run):** units whose sizes overlap (a size 2 player and a size 2 monster at a distance of 3 subtiles or less, `across + down < size1 + size2`) have no line to cut, so a thin wall (one or two subtiles) between two units that close does not stop a hit; and the line is one straight run of subtiles, so a gap in a diagonal wall can let it through. Words 0x1 (walk-only) and 0x400 (an object's) are not in the mask and never cut it.
3. **d2d deviates, in both places.** `Fight::engage` starts the swing within 1.1 cells (5.5 subtiles, centre to centre) with no test of anything between; `Fight::hit` lands within 1.6 cells (8 subtiles). Against game.exe on 101,168 cases (two seeds, wall maps and open ground): the swing start differs in 10,009 (parent) and the blow in 33,963; d2d starts or lands through a wall in 1,546 of 1,661 start cases and 2,047 of 2,047 blow cases where game.exe's distance would allow it and only the wall stops it. Most of the difference is the distance itself: game.exe's reach for two size 2 units is about 3 subtiles, not 5.5.
4. **Fix proven:** `fid/melee-reach-r1` (two commits on upstream `895f8a67`): after the claim 0 of 101,168 differ on the start and 0 of 101,168 on the blow (seeds 1 and 7), 13 pinned game.exe values in `test_player_reach`, 8 of 8 mutants abort it. Verifier-ready after the full `ctest` (section 7).
5. **The bot does not choose a wall-crossing target on purpose.** It clicks a monster within the skill's range (7 subtiles by default) and lets the game walk it into reach (`internal/combat/fight.go:721`); the line test it has (`GridLOS`, a missile bit) is skipped for a melee click. It was d2d's engage that swung through the wall once the click brought the player within 5.5 subtiles. B1 to B3.
6. **DoT (CodeRabbit on #600):** game.exe does NOT apply the hit's rule (life under 256, one whole point, is dead: `FUN_0057c6c0` `0x57c865`) to poison and bleeding. `FUN_005a6920` kills when life is under 1 in 256ths (`0x5a69ec`); 140 of 140 emulator cases equal `life + net`, and 42 of them keep a monster alive between 1 and 255. d2d's `monster_dots` already has that rule (no sub-point rule); what differs is only the grain (section 8). No change proposed.
7. **Not done:** monsters' melee against the player and the pets (`FUN_0057d4f0` asks the same test with `extra` 3), Whirlwind, Charge and Leap Attack (they keep d2d's own reach, D4), the walk's end path in a real map with a live player (needs the lab).
8. **How often:** section 6 (static count on real d2d maps).

## 1. Method

- game.exe side: `tools/emu/dc39_reach.py` (the emulator of `docs/emu-image.md` on the builder; `game` and `hit` modes). Real: `FUN_00622c40`, `FUN_00622dc0` with `FUN_00463860` / `FUN_00463900` over a small MonStats table, `FUN_00641530` (the unit distance), `FUN_00622aa0` -> `FUN_00622920` -> `FUN_0064e260` (the line over the room's collision words), and for the blow `FUN_0057ec10` itself. Hooked: the unit's room (`FUN_00620bb0`: one fake 64 x 64 subtile room), the unit's size (`FUN_00620510`), the weapon's range (`FUN_00622870`) and, for the blow, the rolls inside `FUN_0057ec10` (`FUN_00554200` a foe, `FUN_005a4ee0` strikeable, `FUN_0057d9b0` the to-hit passes, `FUN_0057dfb0` no block). The units are fake (type, path words `+2` / `+6` the subtile position, mode `+0x10`).
- d2d side: `tests/game/test_player_reach.cpp` on the fork, the REAL `World` / `Fight` in the Blood Moor with its collision words (`Level::walk`) overwritten by each case's walls: `--player-reach` prints what `Fight::engage` does (swing started or walk on), `--player-hit` what `Fight::hit` does (the blow landed within 100 tries; armor 0, so the 95 percent cap cannot hide a case). Cases are at subtile centres.
- Cases (`gen N SEED`): every offset within 10 subtiles of a fixed point x monster size 1, 2, 3 x walking or standing x weapon range 0..3 with no wall (10,584), then N random pairs within 9 subtiles with one structured wall (horizontal, vertical, diagonal, anti-diagonal of 1 to 5 subtiles, or a 2 x 2 pillar) plus 0 to 3 scattered cells, words from {4, 4, 4, 0x800, 0x804, 1, 0x400}. Seed 1: 30,584 cases; seed 7: 70,584 (the 10,584 open cases are in both); 101,168 lines in all, each run twice (start and blow).
- Compare with `player_reach.py compare GAME D2D` (exit 1 on any difference, an empty or unequal file, a malformed line or ids out of step) or `dc36_cmp.py`.

## 2. Q1: what decides that a player's melee swing may start on a monster (game.exe)

| Q | Answer | Tag |
|---|---|---|
| Q1.1 Where does the click on a monster go? | the client walks the player to the unit and sends the attack when `FUN_00480c80` stops asking to walk: for every target type except an object (type 2) the answer is `FUN_00622c40(player, target, FUN_00622dc0(target)) == 0` (`0x480d06..0x480d1f`: `neg / sbb / add 1` turns the result into "walk on"). | V read; the function itself V run by upstream (`melee_reach.py`, S) and here |
| Q1.2 What does the server do? | `FUN_00549ba0` (the cast on a unit; callers `FUN_00549d80`, `FUN_00549e00`, `FUN_0054a050`...): the skill's range class `FUN_00645460` 1 (melee) with a target and the flag set asks `FUN_00622c40(player, target, FUN_00622dc0(target))` (`0x549cbf..0x549cc8`); false: the player is sent walking (`FUN_00548a50`), true: `FUN_00580a70` starts the skill. Which packets set the flag (`[ebp+0x14]`) was not decoded. | V read; flag I |
| Q1.3 What is the test? | `FUN_00622c40(unit, target, extra)`: a monster target of class 0x102 / 0x105 (tentacle, head) is in reach under range + 8; `FUN_00641530` at 0 or less is in reach; else `range + extra + 1 >= apart` **and** `FUN_00622aa0(unit, target, 0x804)` finds no cut line. The range is the player's weapon `rangeadder` (`FUN_00622870` through `FUN_0063c9b0`), a monster's MeleeRng. | V read; V run (here and upstream `melee_reach.py`, 0 differ) |
| Q1.4 What is the unit distance? | `FUN_00641530` = `rules::unit_distance`: an 8 x 8 table (`0x6eb180`) under 8 subtiles apart with both sizes under 4, else the deltas less the half sizes, the smaller plus twice the larger. For two size 2 units: apart 0 at a delta of up to (3, 0) and (2, 2), 2 at (4, 0), 4 at (5, 0). So a standing size 2 monster is in reach (range 0, extra 0) out to about 3 subtiles, a walking one out to 4. | V read; V run (the printed reach map, section 4) |
| Q1.5 What is the line? | `FUN_00622aa0` -> `FUN_00622920` -> `FUN_0064e260`: the ends step in by min(size, 2) along the longer axis (both on a diagonal); units whose `across + down < size1 + size2` always see each other; every subtile of the straight line, ends included, is tested for `word & 0x804`. | V read; V run (upstream `sight.py`, S; here through `FUN_00622c40`) |
| Q1.6 What else tests distance before the blow? | Nothing. The do of Attack (`FUN_0056f070` `0x56f115..0x56f1e9`) hits the target `FUN_00553540` returns (the unit's skill target) with `FUN_0057ec10`, `FUN_0057b7d0`, `FUN_0057dbf0`. The strike-event path `FUN_0056ecb0` -> `FUN_0059fa30` is the ranged weapon's (bow, thrown), not a melee swing. | V read |
| Q1.7 What does the blow test? | `FUN_0057ec10` (`0x57ec7e..0x57eca6`): `edi = 1 + 2 * (the attacker is a monster)`, then `FUN_00622c40(attacker, target, edi)` (the third stack argument, 0 in do 1, only adds to the range); false returns no hit and no roll. So a player's blow needs range + 2 and no wall; `FUN_0057d4f0`'s follow-up reads the same call (`paldur-trace.md`, S). | V read; V run (`FUN_0057ec10` itself, rolls fixed) |

## 3. Q2: what d2d does, and the deviations (D1..)

| id | game.exe | d2d (parent `895f8a67`) | tag | differential |
|---|---|---|---|---|
| D1 | the swing starts at `FUN_00622c40(.., FUN_00622dc0(target))`: distance by the unit table, no cut line | `Fight::engage` (`components/game/fight/tick.cpp:248` before the fix): `hypot(monster - player) <= kMeleeReach` (1.1 cells = 5.5 subtiles), sizes, weapon range, a moving monster and walls ignored | V run, both sides | yes, built: `tools/emu/player_reach.py` (`fid/melee-reach-r1`); parent: 2,799 (seed 1) and 7,210 (seed 7) differ |
| D2 | the blow needs `FUN_00622c40(.., 1)`: range + 2, no cut line | `Fight::hit` (`strike.cpp:675`): `hypot <= kMeleeReach + 0.5` (1.6 cells = 8 subtiles), no wall | V run, both sides | yes, built: same oracle, `hit` mode; parent: 9,869 and 24,094 differ |
| D3 | a thin wall between units whose sizes overlap does not stop a hit (authentic) | not modelled either way (d2d hits there too) | V run (pin 12 `30 30 33 30`, wall at 31) | covered by the same oracle; no deviation after the claim |
| D4 | Whirlwind, Charge, Leap Attack: the same blow test applies to each hit | `hit(now, 5.f)`, `hit(now, kMeleeReach + 0.5f)`: their own reach | V read (game); not run | not built: Whirlwind needs the move events of a real sequence (skbarb D1); Charge and Leap place the player by d2d's own model |
| D5 | a monster's melee on the player and the pets asks `FUN_00622c40(.., 3)` (range + 4, the same line) | `kMeleeReach + 0.3` hypot (`ai.cpp:1943`, `2491`, `2627`, `pets.cpp`) | V read only | not built: needs the monster think and attack frame with a collision map; buildable on `Fight` with a monster at (dx, dy) and a wall, oracle `FUN_0057d4f0` / `FUN_0057ec10` with a monster attacker |

### Numbers (V run; `dc39_reach.py` against the real `Fight`, parent = the oracle commit `5ff84fb6`, claim = `38426e2d`)

| run | cases | start: differ on the parent | start: differ after the claim | blow: parent | blow: claim |
|---|---|---|---|---|---|
| seed 1 (`gen 20000 1`) | 30,584 | 2,799 | 0 | 9,869 | 0 |
| seed 7 (`gen 60000 7`) | 70,584 | 7,210 | 0 | 24,094 | 0 |

What the parent gets wrong (seed 1 / seed 7, same cases):

| | start | blow |
|---|---|---|
| d2d acts, game.exe does not, no wall in the way (the distance alone) | 2,270 / 5,710 | 9,362 / 22,554 |
| game.exe acts, d2d does not (big or long-range cases: size 3, polearm) | 150 / 333 | 0 / 0 |
| pairs in game.exe's distance reach that a wall cuts | 411 / 1,250 | 507 / 1,540 |
| of those, d2d acts anyway (through the wall) | 379 / 1,167 | 507 / 1,540 |

The wall count is the lower bound: it counts only pairs close enough for game.exe's distance; d2d's wider disc adds through-the-wall cases that game.exe refuses for both reasons.

## 4. The reach map (V run, no wall, standing monster, weapon range 0; `#` = game.exe starts the swing; dx across, dy down, 1 subtile per column)

```
game.exe, standing  d2d parent, same    game.exe, walking   game.exe, polearm (range 2)
size 2 monster      (any size / range)  size 2 monster      size 2, standing
...............     ...............     ...............     ...............
...............     .....#####.....     ...............     ...............
...............     ....#######....     ......###......     .....#####.....
......###......     ...#########...     ....#######....     ....#######....
.....#####.....     ..###########..     ....#######....     ...#########...
....#######....     ..###########..     ...#########...     ...#########...
....#######....     ..###########..     ...#########...     ...#########...
....#######....     ..###########..     ...#########...     ...#########...
.....#####.....     ..###########..     ....#######....     ...#########...
......###......     ...#########...     ....#######....     ....#######....
...............     ....#######....     ......###......     .....#####.....
...............     .....#####.....     ...............     ...............
...............     ...............     ...............     ...............
```

(From the 10,584 open cases of seed 1; the parent's disc is radius 5.5 for every size, range and mode. The fix reads each case from `FUN_00641530`, so size, range and a moving monster all move the edge.)

## 5. Bot findings (B1..), described, NOT changed

| id | file:line | finding |
|---|---|---|
| B1 | `internal/combat/fight.go:721` | `meleeGo = pr.MeleeClass() && spec.Kind == KindMelee && tg.Dist <= spec.Range && onScreen`: a melee click goes out at any monster up to the skill's range (`defaultMeleeRange` 7 subtiles, U, `classdata.go:178`; `resource.go:22` `meleeRange` 10) and skips the line test at `:726`, relying on the game to walk into reach (comment at `:718`). That is right for game.exe: the client keeps walking until `FUN_00622c40` holds. On a d2d without the claim the same click makes the player swing from up to 5.5 subtiles, through a wall. The bot did not pick a wall-crossing target; d2d let it. |
| B2 | `internal/combat/reach.go:141` | `GridLOS` rounds a line over the cells and tests the `BlockLOS` bit (a missile barrier), with no unit sizes and no 0x800 door word. It is the right idea for a shot (mask 4) but is not the melee test (mask 0x804, sizes stepped in): it should not be used to predict whether a melee swing can start. It is skipped for melee anyway (B1). |
| B3 | `internal/combat/classdata/profiles.json:241`, `:372` (`melee.reach` 4) and `fight.go:711-722` | `reach := pr.Melee.Reach` (4 subtiles) is the walk target for an approach and `inReach := tg.Dist <= reach` is euclid. With game.exe's rule a standing size 2 monster is hit from about 3 subtiles (4 walking), so 4 is close to right; a polearm (range 2) reaches 5 to 6. No change needed; a soak that counts "time in reach" on d2d will shorten once the claim lands. |
| B4 | soak records on the current d2d | every melee kill record from a d2d without `fid/melee-reach-r1` includes swings and blows from up to 5.5 / 8 subtiles and through walls: kills per minute and time to reach a pack are optimistic by an amount this doc does not measure (U). They are not a fidelity gate input. |

## 6. How often does the geometry occur in real maps (static count)

One Act 1 map each (the generator's map seed 4660, `Level::walk` as game.exe's room grids hold it), every free subtile (a player-sized unit can stand: `Level::unit_blocked`) paired with every free subtile around it (`tools`: scratch `test_wall_pairs` on the fork, branch `scratch/dc39-pairs`, not for upstream). "Cut" = a collision word of mask 0x804 on `FUN_00622aa0`'s line (what stops a hit in game.exe); "walk-blocked only" = a mask 9 word on the line and no mask 0x804 one (a cliff or cave edge that stops walking but not a missile: game.exe's melee crosses it as d2d's does). V run for the counts; U for how often a fight stands on such a pair.

| level | free cells with a cut pair within 5.5 | pairs within 5.5 that a wall cuts | pairs in game.exe's reach that a wall cuts | pairs within 8 (the parent's blow) cut | walk-blocked only: in game.exe's reach / within 5.5 |
|---|---|---|---|---|---|
| 2 Blood Moor | 2,010 of 82,465 (2.4 %) | 8,838 of 7,196,774 (0.12 %) | 634 of 2,788,898 (0.02 %) | 42,071 (0.29 %) | 392 (0.01 %) / 24,437 (0.34 %) |
| 3 Cold Plains | 2,127 of 97,228 (2.2 %) | 9,933 of 8,812,314 (0.11 %) | 442 of 3,371,088 (0.01 %) | 55,853 (0.32 %) | 418 (0.01 %) / 14,429 (0.16 %) |
| 5 Dark Wood | 10,448 of 99,156 (10.5 %) | 54,964 of 8,809,628 (0.62 %) | 5,188 of 3,380,768 (0.15 %) | 206,149 (1.17 %) | 576 (0.02 %) / 20,443 (0.23 %) |
| 8 Den of Evil | 42 of 28,842 (0.1 %) | 90 of 2,434,544 (0.00 %) | 0 of 957,620 | 817 (0.02 %) | 26 (0.00 %) / 5,484 (0.23 %) |
| 9 Cave Level 1 | 224 of 99,382 (0.2 %) | 493 of 8,806,700 (0.01 %) | 4 of 3,401,568 | 4,352 (0.02 %) | 94 (0.00 %) / 10,593 (0.12 %) |
| 10 Underground Passage 1 | 244 of 99,838 (0.2 %) | 626 of 8,770,282 (0.01 %) | 20 of 3,398,536 | 9,731 (0.06 %) | 88 (0.00 %) / 11,798 (0.13 %) |

Reading it: in these maps a wall (mask 0x804) lies between two stand-able cells that are close enough for d2d's old swing in 0.1 to 0.6 percent of the pairs outdoors (Blood Moor, Cold Plains, Dark Wood: trees and cliffs) and almost never in the Den and the caves, whose walls are thick. Inside game.exe's own reach the cut is 0.01 to 0.15 percent. So the sighting needs a player at a wall's edge with a monster just behind it; it is rare per pair, common enough over hours of play next to cliffs and trees (U: the share of swings). The deviation is mostly that d2d swings from further (the disc of 5.5 against about 3); the through-wall part is the tail. Cave and cliff edges that block walking but not missiles are crossed in game.exe as well (the walk-blocked column; authentic).

## 7. PR candidate

| branch (private fork remote) | head | commits | proof | verifier |
|---|---|---|---|---|
| `fid/melee-reach-r1` | `38426e2d` on upstream `895f8a67` | oracle `5ff84fb6` (`tools/emu/player_reach.py`, `tests/game/test_player_reach.cpp`, README row), claim `38426e2d` (`Fight::melee_reach_to` for `engage` and the default `hit`, 13 pins, `docs/research/re/combat.md` finding) | V run: 30,584 + 70,584 cases x 2 modes, parent 2,799 + 7,210 start and 9,869 + 24,094 blow differ, claim 0; 8 of 8 mutants abort `test_player_reach` (list below); the pins run without the emulator | not yet; body [decomp39-melee-reach-body.md](../upstream/bodies/decomp39-melee-reach-body.md) |

Mutants (local, never pushed, `mut-reach-1..8` and `6b`): engage back to the hypot; the line test off; mask 0x4 instead of 0x804; the moving extra ignored on the start; the blow's extra ignored; the monster's size ignored (survived the first pin set, killed by pins 14 to 16); a blow on the old 1.6 cells; the weapon range ignored.

Tests run on the claim (V run, the builder, Release, GCC): the 19 test binaries that touch `Fight` (`test_player_reach`, `test_zeal_chain`, `test_game`, `test_game_missiles`, `test_sk_ground`, `test_bone_wall_lay`, `test_cast_mana`, `test_cast_overlay`, `test_pet_town`, `test_quests`, `test_skill_ready`, `test_sk_mana`, `test_tele_mana`, `test_tele_spot`, `test_touch_input`, plus four oracle-mode tests that skip without arguments) pass or skip as before. The whole `ctest` was not run: a full build was started and stopped at 441 of 1,190 targets on the shared builder (about three hours at its share). The verifier runs it on the worker host.

Not run for the candidate: `tools/presubmit.py --quick`, `pr_guard.py` (the verifier does them), the reseed (there is no draw: both sides are deterministic per case; the blow loop retries the to-hit roll, armor 0).

## 8. The DoT sub-point death rule (CodeRabbit on #600, second topic)

| Q | Answer | Tag |
|---|---|---|
| Q8.1 What kills a monster in a hit? | `FUN_0057c6c0` `0x57c85e..0x57c86e`: `life = stat6 - damage` (256ths); `cmp eax, 0x100; jge keep; xor eax, eax`: **life under 256 is 0 and the monster is dead** (one whole point). d2d's `hurt()` (`ai.cpp:2904`: `hit_points < 1 \|\| (hit_points == 1 && life_frac > 0)`) is that rule (`life_frac` is the carried damage fraction: life = `hit_points x 256 - life_frac`), and so is the merc's (`tick.cpp:666`). | V read both |
| Q8.2 What kills it in poison or bleeding? | `FUN_005a6920` (every frame): `net = stat 0x4a`; a negative net under state 0x34's regen block is applied by `life += net`, clamped to the max, and `cmp edi, 1; jge keep; xor edi, edi` (`0x5a69ec..0x5a69f3`): **life under 1 (256ths) is 0 and the monster dies; a life between 1 and 255 survives the tick**. So the DoT does NOT use the hit's one-point rule. | V read; V run (`dc39_dot.py`: `FUN_005a6920` for real, 140 of 140 cases equal `life + net` with the death at under 1; 42 cases keep a life of 1..255 alive) |
| Q8.3 A guard | a damaging tick (net below 0) on a life under 256 does nothing when `FUN_0061ab00(room)` is true: that is `FUN_006426a0` on the level id, true for levels 1, 40, 75, 103, 109 (the five towns) only (V run over levels 1..109). A monster at under one point in a town is not finished by poison. | V run |
| Q8.4 What does d2d do? | `Fight::monster_dots` (`strike.cpp:610`): the net rate accumulates in a double (`dot_acc`), whole points come off `hit_points` (`hit_points -= whole`) and `alive()` is `hit_points > 0`. With life = `hit_points x 256 - life_frac` that dies when the whole points reach `hit_points`, i.e. at life under 0, where game.exe dies under 1: **the same threshold**, and no extra one-point rule for DoT, as in game.exe. | V read |
| Q8.5 What differs? | the grain: game.exe takes the rate in 256ths every frame; d2d takes whole points when the accumulator crosses 1. A monster at `hit_points == 1` with `life_frac` of `f` (life `256 - f`) dies in game.exe after `(256 - f) / net` frames and in d2d when 1 whole point has accumulated (`256 / net` frames): up to under one point of life, about `255 / |net|` frames (poison rates are tens of 256ths a frame: under 0.4 s), later in d2d. The town guard (Q8.3) is not in d2d. | I (arithmetic from V read both); not run on d2d |

**Verdict for the review comment:** do not add the hit's rule to `monster_dots`; game.exe does not have it there (Q8.2). A differential for the grain (D6) would run `FUN_005a6920` for N frames against `Fight::monster_dots` on a `Monster` with a poison rate at a chosen life; not built (the effect is under one point).

| id | game.exe | d2d | tag | differential |
|---|---|---|---|---|
| D6 | DoT grain: 256ths a frame, dead under 1; skipped under 256 in a town level | whole points, dead at `hit_points <= 0`; no town guard | V read / V run (game), V read (d2d) | buildable: `Fight` with a Monster, poison rate set, N frames, life read against `dc39_dot.py` run for N frames |

## 9. Drill recipe (fork verbs, run later in the lab; nothing here ran a game)

Verbs of [../forkdbg.md](../forkdbg.md) and [../forkcheats.md](../forkcheats.md), lab binary [../d2d-fork.md](../d2d-fork.md). Confirms the U items and the user's sighting.

1. **A wall between a Paladin and a Fallen (D1, D2):** a throwaway Paladin (no setup writes needed): `debug spawn fallen1 1 near` against a cave or Den wall (the Den of Evil has 1 to 2 tile walls; Cave level 1 as well), stand on the far side of a wall at about 4 subtiles, left-click the Fallen. Lab binary without the claim: the swing starts and the monster loses life through the wall (read `monsters` before and after). With the claim (branch binary): the player walks round, or, when the units are within 3 subtiles, hits across a thin wall (the authentic D3 case).
2. **The reach edge (D1):** `debug spawn fallen1` at known offsets east of the player (1..6 subtiles; `debug pos` / `dump stats` for the player's spot) on open ground; click it and read the player's mode and spot when the first swing starts: game.exe's edge is about 3 subtiles for a standing size 2 monster, 5.5 on the parent.
3. **Zeal's chain (D2):** `debug spawn fallen1 3` in a row behind a wall and in front of it; Zeal's retargeting uses `in_reach()` (already game.exe's rule, S), so the chain skips the ones behind the wall; the first blow is the one under test.
4. **DoT (D6):** `debug spawn zombie1`, `debug stat` its life to 100 in 256ths where the verb allows (U), poison it (a Poison Javelin or Nec skill) in the Blood Moor and in town (none spawn there), and count the frames to death.

## 10. Matrix cell updates (proposed text; the owners apply them)

- **fidelity-matrix, new row (skills / monsters boundary), "Player melee reach and line"**: `FUN_00622c40` (start, blow) | DEV | `Fight::engage` and `Fight::hit` used hypot 5.5 / 8 subtiles with no wall test: 10,009 of 101,168 starts and 33,963 of 101,168 blows differ (V run, decomp39-melee-walls.md); fix `fid/melee-reach-r1`, 0 after; monsters' own melee (D5), Whirlwind, Charge and Leap (D4) not run.
- **pal-skill-matrix** (Zeal and Attack rows): add "starts within 3 subtiles (standing size 2) in game.exe, 5.5 in d2d; a wall on the line stops the start and the blow in game.exe, not in d2d (V run); a thin wall inside 3 subtiles does not stop either (authentic)".
- **MA09 monster melee**: add "monsters' melee reach on the player: game.exe `FUN_00622c40(.., 3)` with the line, d2d hypot 1.3 (D5, V read only)".
- **MA / death rows (poison, bleeding)**: "DoT kills under 1 in 256ths, not under 256 (Q8.2, V run); the town guard (Q8.3); d2d's grain differs by under one point (D6)".

## 11. Dead ends and what stayed unverified

- **Strike events / `FUN_0056ecb0`:** I first read the swing as a melee missile created at the target (`FUN_0059fa30` with its 100-subtile test). It is the ranged weapons' path (`FUN_0064f460` true in do 1); the melee do hits the target directly (Q1.6). Recorded so nobody follows it again.
- **Which packet sets `FUN_00549ba0`'s flag** and which skill range class values (`FUN_00645460`: 1, 3, 4) map to which skills: not decoded (I). The client's walk end (`FUN_00480c80`) is enough for d2d's single-player case.
- **Vulture rows and 0x102 / 0x105 in `melee_reach_to`:** the vulture's BaseId rule is passed only for row `mon_bin[0x6e]` (the other vulture rows have it as a BaseId: U); the tentacle test is d2d's own existing one.
- **A real map with a live player:** only the static count (section 6) and the pinned cases; a lab drill (section 9) shows the user's sighting itself.
- **Glossary:** "melee reach" and "line cut" are not in `GLOSSARY.md`; proposed: "Melee reach: game.exe's `FUN_00622c40` test (unit distance within weapon range + extra + 1 and no wall of mask 0x804 on the line) that gates a melee swing's start and a blow." (to the coordinator).

## Shared-file lines for the merger

BACKLOG / CHANGELOG / INDEX lines are generated from the task card (TASK-624). The matrix cells above are for their owners; the pipeline rows (`docs/upstream/pipeline-2026-11-14.md`) would get "P23 melee reach and line (`fid/melee-reach-r1`)".
