# Monster base stats: game.exe 1.14d vs d2d vs our bot (MONSTATS, coverage gap monster.stats)

> status: current (2026-11-01, card monsterstats). Offline only: no game, no d2d slot, nothing pushed, posted or filed. Question (docs/coverage/rule-coverage.md: monster.stats 0 of 734): how does game.exe build the stats of a plain monster (life, defence, attack rating, damage, experience, resistances, the difficulty and player-count paths, the 32-bit rounding), does d2d give the same numbers, and what does the bot assume? Includes the per-kill experience and its level-gap step (docs/research/baseline-audit.md flag: d2d pays +20 percent in the Blood Moor, +60 percent in the Den).
>
> Tags: **V read** disassembled by this agent (upstream's `tools/emu/disasm.py` on the patched 1.14d game.exe; addresses given); **V run** game.exe's own code run natively under the emulator and compared field by field with d2d's real compiled code (`rules::monster_stats` / `monster_ranges` / `player_*_pct` from upstream main 7d5b3a5c, built on the worker host); **D** the patched 1.14d tables as game.exe's loader holds them; **S** an earlier trace of ours (named); **I** inference; **U** unverified. Oracles (this card): `tools/emu/ms_stats.py` (game side, `FUN_006538a0` and the player-count helpers), `tools/emu/ms_d2d_test.cpp` + `ms_build.sh` (d2d side), `tools/emu/ms_spawn.py` + `ms_spawn_cmp.py` (the real unit stats of a populated level), `ms_area_xp.py` / `ms_table.py` (tables), `ms_diff.sh` / `ms_pop.sh` / `ms_run.sh` (drivers, worker host).

## Answer in seven lines

1. **For a plain Normal monster d2d's stats equal game.exe's.** The stat rows (life min / max, defence, to-hit and damage of A1 / A2 / S1, experience) are equal in **242,220 of 242,220 cases** (every MonStats row x 3 difficulties x monster level 1..110; V run), and the real unit stats of **85,213 plain Normal monsters** of the Act 1 and Act 2 levels (12 seeds, all levels 2..74 that populate) show 0 differences in level, life range, experience, defence and the six base resistances (V run, 123 rows seen). The player-count tables agree for 0..16 players (17 of 17, V run).
2. **XP per kill is not a rule gap.** A kill's experience is the unit's stat 13, set once at spawn from MonLvl x MonStats Exp % (equal, above); the award step (level gap, Experience.txt ratio) was traced in docs/research/xp-award-trace.md and d2d has had it since upstream 01337ba3. The real populations (Normal, plain monsters) average **22.0 XP per kill in the Blood Moor and 26.1 in the Den** at a clvl at most 5 above the monster, within 3 percent of d2d's sweep (21.5 / 25.4, xp-award-trace.md) and above the live original's 18.3 / 18.1 (bot kill mix): the +20 / +60 percent audit flags are composition, not a rule deviation (section 5, I for the cause).
3. **D1 (NM / Hell level, V run):** in Nightmare and Hell game.exe gives a ratio-row monster the AREA's level (Levels.txt MonLvl2Ex / 3Ex), d2d the monster's own Level(N) / Level(H) column: 33,693 of 82,171 Nightmare units (41.0 percent) sit at another level; d2d's mean experience is 6.4 percent and life 4 percent below game.exe's there. Known to d2d's author (comment on `MonStats`); not Normal, so it does not touch the soak baseline.
4. **D2 (monster life regeneration, V read + I):** game.exe gives every monster a life regeneration, stat 0x4a = (max life in 256ths x MonStats DamageRegen) >> 12 (DamageRegen 2 for 600 of 734 rows), added to life by FUN_00580610 in the event FUN_00580810; d2d's monsters have none (ai.cpp:964 ponytail). Size: max life / 2048 per frame at DamageRegen 2 (1.2 percent of max life per second if it runs every frame, U), nothing for a monster of under 8 life (Fallen, Quill Rat: the stat is life / 8 in whole 256ths), 1/256 life per frame for a 12-life Zombie.
5. **Bot (not changed):** B1 `areaXPPerKill` (10 x the area's monster level) is 1.3 to 2.2 times too low in the lowest areas (Blood Moor 10 vs 22.0, Cave 1 20 vs 35.2) and within 22 percent from Stony Field up; it ties Cold Plains, Cave 1 and Cave 2 where the real order is Cave 1 > Cave 2 > Cold Plains. B2 the Normal fallback levels (Den 2, Cave 2 3) differ from Levels.txt (1, 2). B3 `ScaledHP` is the right formula for ratio rows but is dead code and wrong for the 31 noRatio pet rows. B4 the resource simulation's monster life (12 / 24 / 24 / 45) is 2.5 to 10 times the real Blood Moor life (1 to 12).
6. **PR candidate: none** (D1 is NM / Hell and needs area plumbing the harness lacks; D2 has no d2d regeneration to patch against a harness).
7. **V run vs not:** everything in lines 1 to 3 is V run (game.exe code in the emulator against d2d's compiled code; counts in section 3). D2's consumer and formula are V read; "every frame, after damage" is U. The bot findings are source reads of our code against V run numbers.

## 1. Questions answered (Q1..)

| id | question | answer | where | tag |
|---|---|---|---|---|
| Q1 | Which difficulty and level does a monster unit use? | `FUN_00573cb0` (the monster unit maker, unit in esi, game in `[ebp+8]`, MonStats record in ebx): difficulty = game byte +0x6d, 0 when 0, 2 when >= 2, 0 for a hireling class (`FUN_0063ee90`); level = MonStats int16 at +0xaa + 2 x difficulty (Level, Level(N), Level(H)); if game dword +0x70 is set, difficulty > 0 and the row's flag byte +0xc misses the two boss masks (`0x6ce270`, `0x6ce280`), the level is the area's (`FUN_0061a1b0`, S skill-resist Q1); (the area-level branch is from the skill-resist trace; the call target `FUN_0061a1b0` is read at 0x573e48). | 0x573dc1..0x573e4d | V read |
| Q2 | How are life, defence, to-hit, damage and experience derived? | `FUN_00573cb0` zeroes a 0x38-byte buffer and calls `FUN_006538a0(row, expansion, difficulty, level, 7, out)`; the expansion argument is 1 when game byte +0x6a or dword +0x74 is set. Flag bits of that call (V run, `ms_stats.py flags`): 1 life min / max, 2 defence, 4 experience, 8 A1 (to-hit + damage), 0x10 A2, 0x20 S1. Per MonStats row and difficulty: MonLvl `L-<col>(N/H)` for the level x the MonStats percentage / 100 (truncated), or the MonStats number itself for a noRatio row (flag +0xc & 4). | 0x573f5b..0x573f7e; table equal to d2d in 242,220 cases | V read + V run |
| Q3 | How is life rolled? | min + rand(max - min + 1) on the unit seed (`FUN_0045c3e0`, ecx = unit +0x20), plus MulDiv(life, player-count %, 100) (`FUN_005738f0`), capped below 0x800000 (0x7fffff), stored x 256 in stats 6 and 7. Observed endpoints on the common rows equal the row range (Fallen 1..4, Zombie 7..12, Quill Rat 1..5). | 0x573f83..0x573fc8 | V read; V run (life inside d2d's range in all 85,213 units) |
| Q4 | Defence? | stat 0x1f = the row's defence value as returned, no scaling (no difficulty or armor-percent step in the maker). | 0x573fd3 | V read; V run (0 of 85,213 differ) |
| Q5 | Experience and the player-count bonus? | stat 13 = exp + MulDiv(exp, pct, 100), pct = `FUN_00573910(players)`: table `0x6e15b4` for 0..8, else (5 n + 130) x 2; life uses `FUN_005738f0`: table `0x6e1590` else (n - 2) x 50. d2d's `player_life_pct` / `player_exp_pct` equal for 0..16 (17 of 17). | 0x573fdb..0x573ff1, 0x5738f0, 0x573910 | V read; V run |
| Q6 | Attack rating and damage? | The unit carries no AR or damage stat of its own for a plain monster: the three to-hit values and three damage ranges of Q2 are read again at every attack (`FUN_005a4f50`, stat 12 = the unit's level; docs/research/pr-mods-proof.md). Rows equal d2d's `to_hit`, `to_hit2`, `to_hit_s1`, `a1..s1_min / max` in all 242,220 cases (d2d additionally clamps max >= min: no row is affected). The melee roll and block were differenced in docs/research/monster-attack-trace.md (0 of 300,672). | | V run (rows); S (attack path) |
| Q7 | Resistances by difficulty? | base stats 0x24 / 0x25 / 0x27 / 0x29 / 0x2b / 0x2d = MonStats ResDm / ResMa / ResFi / ResLi / ResCo / ResPo of the difficulty; no penalty or cap for monsters. Real units: 0 of 85,213 differ from d2d's `type.diff[d].res` (Normal); NM / Hell units 0 differ at the game's level. | docs/research/sk-resist-trace.md Q1 | S + V run |
| Q8 | The 32-bit rounding? | all truncating integer: `base x percent / 100` per row (equal to d2d at every level, so equal overflow behaviour up to level 110), `MulDiv` (`FUN_00483360`, 64-bit intermediate) for the player bonus and the experience bonus, life x 256 stored in an int32 (cap 0x7fffff00). | | V run (rows), V read (maker) |
| Q9 | Does a monster regenerate? | yes: stat 0x4a = (max life x 256 x DamageRegen) >> 12 set at spawn (V run: 1 for a 11-life Zombie, 26 for an NM Fallen of 211 life, 80 for an NM Zombie of 642: all equal to `(life256 x 2) >> 12`); `FUN_00580610` adds it to stat 6 (capped at max life, floor 0x100) and `FUN_00580810` reschedules event 3 for the next frame; the code at 0x57c8e2..0x57c93b (the damage path, taken while life stays above 0, with stat 0x4a != 0) removes and re-adds event 3 for the next frame. | 0x574030..0x574058, 0x580610..0x58066d, 0x580810, 0x57c8fd | V read; trigger and frequency I |

## 2. game.exe's path (read)

`FUN_00573cb0` (monster unit maker), the part that matters here, in order:

1. 0x573dc1..0x573ded: difficulty (byte +0x6d clamped 0..2; 0 for hirelings). 0x573e15: level = MonStats int16 `[row + difficulty*2 + 0xaa]`; the area-level branch at 0x573e23..0x573e4d (needs game +0x70, difficulty > 0, not a boss row).
2. 0x573f3e..0x573f54: expansion argument = (byte game+0x6a != 0) or (dword game+0x74 != 0). **A harness trap (V run, section 3): the stock emulator game has +0x6a = 0 and +0x70 != 0, which gives the CLASSIC Nightmare / Hell columns (defence and Hell life differ) with the expansion area level.** `ms_spawn.py` sets +0x6a = 1 (a LoD game); Normal is the same either way.
3. 0x573f5b..0x573f7e: `memset(buf, 0, 0x38)`, then `FUN_006538a0(row, expansion, difficulty, level, 7, buf)`: buf = life min, life max, defence, (to-hit), experience.
4. 0x573f83..0x573fb4: `life = rand(max - min + 1) + min` on the unit seed; `life += MulDiv(life, [ebp-0x24], 100)` (player-count %); `if life >= 0x800000: life = 0x7fffff`.
5. 0x573fb6..0x573fc8: stat 7 and stat 6 = life << 8. 0x573fd3: stat 0x1f = defence. 0x573fdb..0x573ff1: stat 13 = exp + MulDiv(exp, [ebp-0x20], 100).
6. 0x573ff6..0x574058: `[row + 0x1a0]` (MonStats DamageRegen, range-checked: negative or >= 0x1000 aborts with a data error 0xde / 0xdf) x life(256ths) >> 12 (with an overflow guard: when life > 0x7fffffff / rate the shift comes first) -> stat 0x4a.
7. Then the resistance stats, the MonProp rows, the level stat (stat 12) etc. (S: sk-resist-trace.md Q1, monster-mods-trace.md); the boss mods run later (`FUN_005a2120`) and raise stat 12 after the rows were read (S, pr-mods-proof.md).

d2d: `rules::monster_stats` (components/rules/monsters.hpp) does the same per field (`pct()`: base x percent / 100 or the noRatio value; life `rng.range(min, max)`, x (100 + player %) / 100 capped at 0x7fffff; experience the same); `monster_ranges` gives min / max at the type's level. The type's level comes from `MonType::level[difficulty]` (Level, Level(N), Level(H) of MonStats) plus `level_add` for bosses. d2d has no regeneration (`ai.cpp:964`, V-d2d).

## 3. Oracle runs (counts)

All on the worker host, 2026-11-01, `tools/emu/ms_diff.sh` and `tools/emu/ms_pop.sh`; d2d side = the upstream scratch build of the SKGROUND2 / SORCREST line (main 7d5b3a5c, `monster_stats` unchanged there), built by `ms_build.sh`; data = the patched 1.14d stack.

| run | game side | d2d side | cases | differ |
|---|---|---|---|---|
| R1 stat rows | `FUN_006538a0` (flags 0x0f, 0x10, 0x20 per case), all 734 MonStats rows x difficulty 0..2 x level 1..110 | `monster_ranges` + `monster_stats` with the type level set to the case level | **242,220** | **0** in minhp, maxhp, ac, th (A1), A1 / A2 / S1 min / max, exp, A2 and S1 to-hit; life roll of d2d (32 rolls) inside the game's range in every case except 2 rows (below) |
| R2 players | `FUN_005738f0`, `FUN_00573910`, ecx = 0..16 | `player_life_pct`, `player_exp_pct` | 17 | 0 |
| R3 Normal units | the real room population (`objgroups.quest_stubbed`, 12 seeds x up to 72 levels, Act 1: Levels ids 2..39, Act 2: 41..74; a level that is not populated gives no units) with the stat setter hooked: 91,960 units, 85,569 plain (flag byte & 0x1c = 0); 335 further units with flag byte 0 but level +3 / exp x5 (boss-pack members, excluded); 21 without a row | d2d rows at the unit's level | **85,213** compared | **0** in level (stat 12 = the row level = d2d's level), life (inside d2d's range), experience, defence and the six resistances |
| R4 Nightmare units | same, difficulty 1, 12 seeds, expansion byte set | same | 82,171 compared (82,766 plain, 595 pack members excluded) | 33,693 at another level than d2d's Level(N) (**D1**); 0 in life range, exp, defence, resistances at the game's level |
| R5 Hell units | same, difficulty 2 | same | see section 3a | D1 |

Not matched in R1: two rows, `mephistospirit` and `minionspawner1`, have MaxHP 0 below MinHP (88 / ...): d2d rolls `range(min, 0)` = min, game.exe computes `rand(max - min + 1)` with a negative span (not run; rows of summons, U). Both are in no Act 1 or 2 population. 46,420 R1 cases are rows with no life (objects, traps) where d2d's `max(.., 1)` gives 1 and the game 0; irrelevant.

Mean values of the real populations (Normal, plain monsters, all seeds): experience 196.5, life 61.3, equal to d2d's at its level (R3).

### 3a. Hell

The Hell and Nightmare units use the area's level; the run is slow (about 7 minutes a difficulty). The Hell numbers are in the closing section when finished; if absent, Hell is not run (U) and D1 is shown for Nightmare only.

## 4. Real values: Act 1 and 2 (Normal, V run, game.exe rows at the monster's own level)

`ms_table.py` on the R1 output; life is min..max rolled, damage per hit min..max (A1), experience before the level-gap step.

| monster | level | life | defence | to-hit A1 | A1 dmg | A2 dmg | exp |
|---|---|---|---|---|---|---|---|
| Fallen | 1 | 1..4 | 5 | 8 | 1..2 | 1..2 | 18 |
| Zombie | 1 | 7..12 | 5 | 8 | 1..3 | 2..3 | 33 |
| Quill Rat | 1 | 1..5 | 5 | 8 | 1..2 | | 21 |
| Fallen Shaman | 2 | 5..9 | 10 | 12 | 1..3 | | 32 |
| Brute | 2 | 11..19 | 10 | 12 | 2..3 | 3..4 | 48 |
| Skeleton | 2 | 7..11 | 10 | 12 | 1..3 | 1..3 | 34 |
| Corrupt Rogue | 2 | 5..9 | 10 | 12 | 1..3 | | 31 |
| Goatman | 4 | 16..20 | 12 | 23 | 3..5 | | 54 |
| Foul Crow | 4 | 2..6 | 3 | 23 | 1..2 | | 22 |
| Corrupt Archer | 4 | 13..19 | 20 | 66 | 2..3 | | 54 |
| Skeleton Archer | 5 | 11..24 | 25 | 89 | 2..3 | | 62 |
| Wraith | 7 | 19..37 | 35 | 52 | 3..6 | | 83 |
| Fetish | 10 | 22..38 | 50 | 80 | 5..8 | | 96 |
| Big Head | 12 | 26..50 | 45 | 85 | 6..9 | | 113 |
| Vulture | 14 | 32..51 | 70 | 125 | 9..12 | | 140 |
| Mummy | 13 | 29..69 | 65 | 113 | 5..10 | 5..10 | 132 |
| Sand Raider | 13 | 59..92 | 65 | 213 | 6..9 | 9..12 | 180 |
| Scarab | 14 | 32..53 | 70 | 125 | 5..15 | 5..12 | 211 |
| Andariel | 12 | 1024 | 60 | 169 | 6..19 | | 1282 |
| Blood Raven | 10 | 113 | 38 | 225 | 4..6 | | 181 |
| Duriel | 22 | 3995 | 112 | 444 | 19..25 | 19..22 | 6007 |

Mean experience per kill of the real plain populations, Normal, and what it pays a character of level c (game: the level-gap step of `FUN_0057e2f0`, tables 0x6e1668 / 0x6e1694, applied to every unit; bot: the formula of `areaXPPerKill` (internal/bot/recoveryxp.go) applied to each unit's own level; the bot itself uses the area's Levels.txt level, see B1). `ms_area_xp.py`, 12 seeds; cells are `game / formula on unit levels`.

| area (Levels id) | units | mean exp | mean level | mean life | c = 1 | c = 5 | c = 10 | c = 15 | c = 20 |
|---|---|---|---|---|---|---|---|---|---|
| Blood Moor (2) | 1163 | 22.0 | 1.0 | 4.2 | 22.0 / 10.0 | 22.0 / 10.0 | 4.9 / 2.4 | 0.4 / 0.5 | 0.4 / 0.5 |
| Cold Plains (3) | 2287 | 25.6 | 1.5 | 5.2 | 25.6 / 14.7 | 25.6 / 14.7 | 8.8 / 5.3 | 0.5 / 0.7 | 0.5 / 0.7 |
| Stony Field (4) | 1630 | 37.3 | 3.1 | 9.7 | 37.3 / 31.1 | 37.3 / 31.1 | 24.1 / 23.3 | 1.5 / 1.6 | 1.5 / 1.6 |
| Dark Wood (5) | 2625 | 47.0 | 5.0 | 9.1 | 47.0 / 50.3 | 47.0 / 50.3 | 47.0 / 50.3 | 2.5 / 2.9 | 2.2 / 2.6 |
| Black Marsh (6) | 2248 | 52.0 | 5.5 | 14.0 | 51.2 / 54.2 | 52.0 / 54.6 | 52.0 / 54.6 | 8.2 / 8.2 | 2.4 / 2.8 |
| Den of Evil (8) | 849 | 26.1 | 1.2 | 5.9 | 26.1 / 12.2 | 26.1 / 12.2 | 7.4 / 3.8 | 0.6 / 0.6 | 0.6 / 0.6 |
| Cave 1 (9) | 1352 | 35.2 | 2.0 | 9.1 | 35.2 / 19.7 | 35.2 / 19.7 | 15.9 / 9.6 | 1.2 / 1.0 | 1.2 / 1.0 |
| Cave 2 (13) | 430 | 27.2 | 1.5 | 6.0 | 27.2 / 15.3 | 27.2 / 15.3 | 10.1 / 6.1 | 0.6 / 0.8 | 0.6 / 0.8 |
| Burial Grounds (17) | 447 | 40.2 | 2.0 | 11.3 | 40.2 / 20.0 | 40.2 / 20.0 | 16.5 / 8.6 | 1.6 / 1.0 | 1.6 / 1.0 |
| Outer Cloister (27) | 552 | 76.4 | 8.4 | 19.9 | 35.2 / 41.2 | 76.4 / 83.9 | 76.4 / 83.9 | 55.2 / 59.9 | 3.2 / 4.3 |
| Catacombs 2 (35) | 914 | 97.2 | 10.6 | 29.1 | 6.0 / 8.4 | 85.5 / 96.1 | 97.2 / 106.0 | 97.2 / 106.0 | 19.7 / 18.7 |
| Rocky Waste (41) | 1523 | 159.3 | 14.0 | 42.2 | 2.8 / 2.7 | 23.1 / 20.7 | 159.3 / 140.0 | 159.3 / 140.0 | 128.4 / 113.3 |
| Lost City (44) | 1671 | 218.6 | 16.8 | 78.0 | 6.6 / 3.4 | 6.6 / 3.4 | 146.8 / 114.4 | 215.8 / 168.3 | 215.8 / 168.3 |
| Sewers 1 (47) | 1001 | 140.0 | 13.0 | 53.3 | 2.1 / 2.5 | 49.8 / 46.7 | 140.0 / 130.0 | 140.0 / 130.0 | 86.4 / 80.7 |

(Units include every plain monster the room population placed, bosses and their packs excluded; the mean is per unit, not per kill the bot makes.) At c = 1 and c = 5 the real Blood Moor / Cold Plains / Den pay 22 to 26; the live original paid 18.3 (Blood Moor) and 18.1 (Den) because the bot kills Fallen far more often than Zombies (xp-award-trace.md: 357 Fallen 123 Quill Rats 117 Zombies in d2d's sweep vs 39 / 17 / 6 live). V run for the means; the cause is I.

## 5. Deviations (d2d vs game.exe)

| id | rule in game.exe | rule in d2d | tag | differential |
|---|---|---|---|---|
| D1 | NM / Hell: a ratio-row monster not on a boss row takes the area's level (`FUN_00573cb0` 0x573e23..0x573e4d -> `FUN_0061a1b0`, Levels.txt MonLvl2Ex / 3Ex); life, exp, defence, to-hit, damage all read at that level (equal function, R1) | `MonType::level[difficulty]` = MonStats Level(N) / Level(H); the source comment on `MonStats` says "Nightmare / Hell take the area's level (FUN_0061dca0) for a ratio row, not Level(N) / Level(H), which waits for those difficulties" | V run: R4, 33,693 of 82,171 plain NM units (41.0 percent) at another level (Skeleton 4: 49 vs Level(N) 43, 4,216 units; Hell Bovine 64 vs 61; Fallen 37 vs 36); d2d's mean exp 4,815 vs 5,144 (-6.4 percent), life 903 vs 940 (-4 percent), and for the affected units exp 5,124 vs 5,926 (-14 percent). Hell: section 3a | yes: `ms_pop.sh` (game: real population; d2d: rows at `type.level`); a fix needs the area level passed to `monster_stats` (not in the harness) |
| D2 | every monster has life regeneration stat 0x4a = (life256 x DamageRegen) >> 12, added to life each frame by the regen event (`FUN_00580610` / `FUN_00580810`) once hurt (I) | none (`ai.cpp:964`: "regen (stat 0x4a) is 0, d2d's monsters having none") | V read (formula, adder), V run (the stat values on 4 shapes of unit); trigger and per-frame rate I; effect on kill time U | could be built: run `FUN_0057c8d0` (damage) then `FUN_00580810` for N frames on a populated unit and read stat 6; oracle = d2d's life (constant). No d2d harness exists for monster life over time |
| D3 | `rand(max - min + 1)` for MaxHP below MinHP (`mephistospirit`, `minionspawner1`: 2 rows, summons) | `max(range(min, max), 1)` = min | U (not run); outside Act 1 / 2 | yes, trivially: `FUN_0045c3e0` with edx = max - min + 1 negative; not built, irrelevant for the soak |

Equal (no deviation, V run): all stat rows at all levels (R1), player-count tables (R2), Normal unit level, life range, experience, defence, resistances (R3), to-hit and damage rows including the A2 / S1 ones. The d2d clamp `a1_max >= a1_min` never fires. Boss / minion deviations are in monster-mods-trace.md and pr-mods-proof.md (not repeated).

## 6. Bot findings (described, NOT changed)

| id | where | finding | tag |
|---|---|---|---|
| B1 | internal/bot/recoveryxp.go:54-56 `areaXPPerKill` = 10 x the area's monster level (Levels.txt column, `areaMLvl`) x gap share | At clvl 5 the bot's numbers are Blood Moor 10, Cold Plains 20, Stony Field 40, Dark Wood 50, Black Marsh 60, Cave 1 20, Cave 2 20 (Burial 30); the real means of a plain kill (V run, table in section 4) are 22.0, 25.6, 37.3, 47.0, 52.0, 35.2, 27.2 (40.2): the bot is 0.45x (Blood Moor), 0.57x (Cave 1), 0.74x (Cave 2), 0.78x (Cold Plains) low and 1.07x to 1.15x high from Stony Field up (Act 2 at clvl 10: Rocky Waste 140 vs 159.3, Lost City 114 vs 146.8, i.e. up to 22 percent low). The picker's order at clvl 5 and 10 is Black Marsh > Dark Wood > Stony Field > (Cold Plains = Cave 1 = Cave 2, a tie broken by the first found) > Blood Moor; the real order inside the tie is Cave 1 > Cave 2 > Cold Plains (35.2 / 27.2 / 25.6 at clvl 5). Small, and only for a recovering Normal Act 1 character; any xp/kill threshold built on the number would be off by the ratios above. The "base of 10 per monster level" in the comment is a fit, not data. | V run (game side), S (bot code read) |
| B2 | internal/bot/recoveryxp.go:16-17 `normalMLvl` | the fallback table says Den 2, Cave 2 3 (wiki recollection); Levels.txt MonLvl1Ex (D, `ms_stats.py levels`) is Den 1, Cave 2 2, Burial Grounds 3; the mean level of the real units is Den 1.2, Cave 2 1.5, Burial 2.0 (a Normal unit uses MonStats Level, not the area column: 0x573e15). Used only when the install data is missing. `areaMLvl` (internal/bot/geargate.go:155) reads the Levels.txt column, which is an upper bound of the units' levels in Normal (within about 1). | D + V run |
| B3 | internal/data/db.go:413-421 `ScaledHP` | formula `HPMax x monlvl HP % / 100` equals game.exe's (`L-HP x maxHP % / 100`, R1) for ratio rows; it ignores the noRatio flag (31 rows, all pets and summons: golems, valkyrie, wolves, sentries), does not clamp the level to the last MonLvl row (game: yes), and returns the maximum, not the roll. Only tests call it, so no effect. The doc comment says "UNVERIFIED": it is V run now for hostile rows. | V run + S |
| B4 | internal/resource/sim.go:122,139 `BloodMoorSorc` `MonsterHP: {12, 24, 24, 45}` | real Blood Moor Normal monsters have 1..4 (Fallen), 1..5 (Quill Rat), 7..12 (Zombie) life, mean 4.2; a clvl 8 Sorceress there kills a 12-life monster with one Fire Bolt of the scenario's damage 8 only by luck: the simulated monsters are 2.5 to 10 times too sturdy. The scenario's calibration is not documented (U); only the resource sim uses it. | V run + S |
| B5 | internal/bot/recoveryxp.go:33-48 `xpLevelPct` | equals game.exe's level-gap step (tables 0x6e1668 / 0x6e1694: 100 percent to a gap of 5, then 80.9 / 62.1 / 43.0 / 23.8 / 5.1 below and 87.9 / 68.0 / 35.9 / 14.8 / 2.0 above; clvl >= 25 above the monster: clvl / mlvl) up to float rounding (game: truncating MulDiv). No finding. | S (xp-award-trace.md) + I |

## 7. PR candidates

None. D1 is Nightmare / Hell only and a fix needs the area level handed to `rules::monster_stats` (a design change in `ai.cpp` / `fight.cpp`, three call sites) that the existing harness (which sets the type's level) cannot prove "after = 0" for; D2 is a missing feature (monster regeneration), not a wrong number, and its rate and trigger are U; D3 is two summon rows. The existing differential shows 0 mismatches for Normal: nothing to patch.

## 8. Drill recipe for the fork verbs (docs/forkdbg.md, docs/d2d-fork.md, docs/forkcheats.md); NOT run

The lab binary is d2d, so a drill shows d2d's side only; the game.exe side stays the emulator numbers above. Each item says what to read and what d2d should show (to catch a d2d regression, or a change in the lab binary).

- **S1 stat rows (Normal)**: `debug goto 2` (Blood Moor), `monsters` (life per unit), `dump stats`: expect each Fallen life in 1..4, Zombie 7..12, Quill Rat 1..5 (row table in section 4). With `debug mon set <id> maxlife <n>` the range is bypassed; do not use it for this read.
- **S2 experience per kill (U for the real game)**: kill known units at clvl 1 and 5 with `debug cast` + `debug mon set life 1`: `dump stats` experience delta must be 18 / 33 / 21 (Fallen / Zombie / Quill Rat; section 4). The same on game.exe is the live check already listed in docs/d2d-fidelity.md ("log the monster name with every XP gain").
- **D2 regeneration (U)**: `debug spawn` a Zombie, `debug mon set <id> life 3 maxlife 11`, `step 250`, `monsters`: d2d shows life 3 (no regeneration). On game.exe the same read should show the life rising by about 1/256 per frame (V read formula, I rate); a larger monster (any 100-life row, `debug mon set maxlife 100`) rises 12/256 per frame. Needs the original (live lab, not done).
- **D1 (U)**: a Nightmare lab game (if the lab can start one, check docs/d2d-fork.md): `monsters` levels of a Skeleton 4 (d2d shows Level(N) = 43; the emulator population gave game.exe 49 where it stood) and a Fallen (36 vs 37). d2d must show the Level(N) column; the area levels are in Levels.txt MonLvl2Ex (`ms_stats.py levels`).

## 9. Matrix and table updates (proposed text)

- docs/coverage/ledger.tsv (applied by this card): `monster.stats all TESTED-DEVIATION docs/research/monster-stats-trace.md 242220 2026-11-01` with note `stat rows R1 (life, defence, to-hit, damage, exp: 0 differ), units R3 85,213 0 differ; NM / Hell D1 level (MonStats Level(N) vs area), D2 regen; resists checked on the 123 rows the populations hold`; and `docs/coverage/gaps.tsv` row 3 removed (the generator refuses a gap with a status), `rule-coverage.md` regenerated.
- docs/research/baseline-audit.md row "Monster base stats: 734 rows, none compared" -> "monster base stats: stat rows equal d2d's in 242,220 cases and 85,213 real Normal units (docs/research/monster-stats-trace.md); not a source of the XP/h gap; NM / Hell level and regeneration open".
- docs/research/baseline-audit.md row "Blood Moor XP per kill +20 percent, Den +60 percent": add "per-kill XP is equal on the same monsters (R1, R3); real plain Blood Moor mean 22.0, Den 26.1 (V run); the live original's 18.3 / 18.1 is the bot's kill mix (I)".
- docs/research/d2d-fidelity-riskmap.md row 3 (XP per kill by monster): done for the rows: "closed for Normal, stat and award rules equal (monster-stats-trace.md, xp-award-trace.md)".
- docs/research/sorc-skill-matrix.md: no monster stat rows; cells that read "monster life untraced" (Fire Bolt / Ice Bolt kills-per-cast notes) can cite the table in section 4 for Act 1 life values.

## 10. Dead ends and limits

- The first Nightmare / Hell population run gave 40,923 "defence" and 26,070 Hell "life" mismatches. Cause: the harness game has byte +0x6a = 0, so `FUN_006538a0` read the classic NM / Hell columns (probe `ms_stats.py probe`: row 19 level 36 NM defence 314 classic vs 369 expansion). Fixed by setting +0x6a in `ms_spawn.py` (hook at 0x573f41); only the Normal numbers do not depend on it. U: whether other game fields the emulator leaves at 0 differ from a LoD game (nothing else in the maker reads them).
- Plain units with level +3, life x2, exp x5 and a zero flag byte (335 in Normal, 595 in NM): champion-pack members read after population; excluded as bosses (their rows are read at the base level: pr-mods-proof.md). Their flag byte is not the cause here; U which path leaves it 0.
- The R3 comparison uses the unit's rolled life against d2d's range, not a seed-for-seed roll; d2d's `Rng` and the game's seed stream were not aligned (the roll order depends on the whole population). The range endpoints are reached on the common rows.
- Resistances: R3 / R4 compare the six base stats of the 123 rows the Act 1 / 2 populations hold against d2d's `type.diff[d].res`; the rest of the 734 rows rest on S (sk-resist-trace.md reads the same columns).
- Boss, champion, unique and minion stats, missiles and the melee roll are in earlier traces (monster-mods-trace.md, pr-mods-proof.md, monster-attack-trace.md, monster-missile-trace.md); not repeated.
- Not traced: monster-vs-monster, pets' stats (hireling rows), `MonProp` effects on the stats (to-hit and damage mods of NM / Hell rows are applied at the end of the maker: `monster_props` in d2d covers speed, rate, crush, thorns; U for the others), the quest and act bosses' spawn paths.
- `ms_scan.py` found seven readers of stat 0x4a (0x57c8ff, 0x580613, 0x5a692d, 0x5a6948, 0x5f0d67, 0x5f0f78, 0x5f50c0); only the first two were read (the rest are AI, UI and merc code, U).

## Shared-file lines for the merger

See `openspec/changes/monsterstats/merge-lines.txt`.
