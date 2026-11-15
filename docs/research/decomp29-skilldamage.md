# Skill damage, cost and speed as data for the power estimator (DECOMP29)

> status: research, 2026-11-15, agent decomp29 (Sonnet), offline (emulator and a d2d test binary on the worker host; no game, no live d2d, no Go change, nothing pushed or posted). Data file: `internal/power/skilldamage.json` (schema `power.skilldamage/1`). Builds on [pal-skill-matrix.md](pal-skill-matrix.md), [sorc-skill-matrix.md](sorc-skill-matrix.md), [skill-damage-trace.md](skill-damage-trace.md), [skiasgame2-trace.md](skiasgame2-trace.md), [cast-trace-gameexe.md](cast-trace-gameexe.md), [sk-mana-trace.md](sk-mana-trace.md), [xp-gold-trace.md](xp-gold-trace.md) (method), [../survival/power.md](../survival/power.md) (consumer).

Tags: **V run** = game.exe 1.14d's own code run in the emulator this card (`tools/emu/dc29_*.py`); **D** = patched 1.14d table cell (`mpq-table.sh`, AnimData read by `dc29_build.py`); **S** = an earlier trace of ours; **I** = inferred (arithmetic on V / D numbers); **U** = unverified. A number called V run below was printed by game.exe's function, not typed from a table.

## Answer

1. **The damage of every class skill, per level and per synergy set, is now data.** 131 skills (all seven classes plus plain Attack, Throw, Kick), levels 1 to 30, three synergy sets (none, every other skill at 1, every other skill at 20), element min / max / length (FUN_00644d50 / e40 / f20, flag 1), the skill's own physical min / max (FUN_00647bc0 / d00), the four calc slots (FUN_00646ca0: ED percent, hit count, stun, bolt count ... named by `calc_src`) and the mana paid (FUN_0056c160 check and FUN_0056bfe0 payment). All V run.
2. **d2d's element damage equals game.exe's on all class skills.** `skill_damage.py game` against d2d's `elem_damage` (the fork's `fix/skill-damage` build, the merged #146 state): **24,750 cases (skills 6..280 x levels 1..30 x 3 synergy sets), 0 differ** (V run, section 3). The two real deviations found by DECOMPSKILL (the min gate on 1-x lightning skills, MulDiv above 0x100000) are fixed upstream and show no residue on this larger grid.
3. **The speed side is a table, not a constant.** Plain cast takes `ceil(256 N / rate) - 1` ticks, a sequence cast and an attack `ceil(256 N / rate)`, with `rate` from game.exe's own FUN_00623f50 (V run grid, FCR 0..200, IAS 0..200, WSM -20..+30) and `N` the AnimData frame count (D). Sorceress cast 13 ticks at 0 FCR (breakpoints 9, 20, 37, 63, 105, 200), Paladin one-hand swing 15, Barbarian one-hand 16. The estimator's `swingSeconds` 1.28 (`internal/power/profiles.json:270`) is **2.1 to 2.5 times** the real 0.52 to 0.60 s of those cases (B1).
4. **No new d2d deviation, no PR candidate.** The differential for elemental damage is 0; the calc slots and physical min / max of each class are 0-diff in the per-class cards (S, section 5); the open deviations that move dps (Zeal spacing, sequence -30, Smite) are carried by earlier cards and recorded as pipeline rows P17 to P19 (not built).

## 1. Method (oracle first)

1. `tools/emu/mpq-table.sh skills|missiles` (D): the Skills.txt and Missiles.txt of the patched stack; `d9_animdata.sh` + a parser in `dc29_build.py` (D): AnimData frame counts per class, mode and weapon class (the parse consumes the file exactly: 570,304 bytes).
2. `tools/emu/dc29_skill.py all 1 30` (V run): boots `skill_tips.boot()` (game.exe under unicorn, as `skill_damage.py`), hooks only the stat getter FUN_00625480 (stat 0xc = clvl 30, masteries 0, else 0), the unit's skill record FUN_00643810 and its level FUN_006442a0, and calls game.exe's own FUN_00644d50 / e40 / f20 (element, flag 1), FUN_00647bc0 / d00 (physical) and FUN_00646ca0 on `calc1..4` for every skill id 6..280 and level 1..30 (8,250 pairs x 3 synergy sets), plus ids 0..5 at level 1. 8 seconds of game CPU per 1,000 pairs.
3. `dc29_skill.py mana 1 30` (V run): the real check FUN_0056c160 and payment FUN_0056bfe0 with a huge mana pool (sk_mana.py's rig); `paid` / 256 is the cost.
4. `tools/emu/dc29_rate.py` (V run): FUN_00623f50's attack branch (mode 7, AnimData speed 256, IAS 0..200 x WSM 0, -10, -20, +10, +20, +30) and cast branch (mode 10, FCR 0..200) in the skiasgame2 rig (stat getters hooked, the rate arithmetic is game.exe's).
5. `tools/emu/dc29_cmp.sh` (V run, d2d side): `dc29_cases.py` writes 24,750 cases; `skill_damage.py game` and the fork's `test_game --skill-damage` run them; `diff` counts the mismatches (0). Reproduce: `D2_LABRUN_HOST=<worker host> tools/emu/run-on-worker.sh --ttl 900 dc29_cmp.sh`. md5 (first 8): cases 8e185b79, game side 4f0d5661.
6. `tools/emu/dc29_build.py` turns the runs and tables into `internal/power/skilldamage.json`; `dc29_doc.py` prints the tables below from the JSON (no hand-typed numbers).

Hooked, not the claim (as skill_damage.py): the stat getter, the skill-record lookup, the level lookup, and for the rate the stat getters and a few side-effect stubs (skiasgame2-trace.md lists them). **What the emulator did not evaluate:** a skill's whole `srvdofunc` (target search, hit roll, block, resist, leech, party) and the missile flight; the unit's `ulvl` terms use clvl 30 (Raven, Fenris, Spirit Wolf, Grizzly show clvl-30 values only); masteries are 0 (a mastery is a plain percent of the element damage, V read FUN_00644c90); stat-driven bonuses (Strength / Dexterity percent, gear ED, Sacrifice's price) are the weapon side, not the skill row.

## 2. Questions answered

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | What are the min / max element damage and length of each skill at level L? | `skilldamage.json` `skills.<name>.elem.<set>.{min,max,len}` (256ths; length in ticks); T1 shows the focus skills. Fire Bolt L1 3 to 6, L20 45.5 to 60.5; Fire Ball L20 199.5 to 226.5; Blizzard L20 570 to 619 | V run |
| Q2 | What do synergies add? | `elem.all20` against `elem.none`: every named synergy at 20 points. Linear per point: `all1` minus `none` is the share of the 1-point synergy sum. Fire Bolt L20: 45.5 to 60.5 alone, 336.7 to 447.7 with Fire Ball and Meteor at 20 | V run |
| Q3 | Lightning skills' min? | The min stays 1 whatever the synergy (the gate: synergy added only when the base min is over one point, FUN_00644d50 0x644db3); Lightning L20 1 to 272, with synergies 1 to 1577 | V run (SKDMG) |
| Q4 | Physical skills (Zeal, Bash, Fend ...)? | The skill's own `phys` min / max (Skills.txt MinDam: War Cry, Shock Wave, Fury ...), plus the weapon share `srcdam` / 128 and the ED percent in a calc slot (T2). Sacrifice ED 180 at L1, 465 at L20, 865 with the synergies (Redemption, Fanaticism); Zeal calc1 hits 2, 3, 4, 5; calc2 ED 0 at L1, 96 at L20 | V run |
| Q5 | Mana per cast? | `mana[level]` (points, V run through the real check and payment); formula columns kept in `mana_cols` (cost = (mana + lvlmana (L - 1)) << manashift, in 256ths, S sk-mana-trace). Fire Bolt 2.5 flat, Charged Bolt 3 + 0.5 per level, Frost Nova 9 + 1 per level | V run |
| Q6 | How long is one cast or swing? | T3 and T4 below; the arrays `rate.attack_wsm0[ias]` and `rate.cast[fcr]` in the JSON; `frames.<class>.<mode>.<weapon class>`. Sorceress SC 14 frames, 13 ticks at 0 FCR; Paladin A1 15 / 14 / 13 ticks at IAS 0 / 9 / 19 with a one-hand sword | V run (rate), D (frames) |
| Q7 | Which skills run at the cast rate? | `anim` SC and the sequence skills whose `seqtrans` is SC (Lightning, Chain Lightning, Inferno? see U1); the rest at the attack rate. JSON `rate_kind` | V read (cast-trace-gameexe.md) |
| Q8 | Per-skill delay (cooldown)? | `delay_frames` in ticks: Blizzard 45, Meteor 30, Hydra 40, Fire Wall 35, Frozen Orb 25 (D); the period of a delayed skill is the larger of delay and cast ticks (I) | D, I |
| Q9 | Area and pierce notes | `missile.pierce`, `vel`, `range`, `explosion` from Missiles.txt (D, raw strings); radius per skill sits in calc / params (Meteor `radius` calc1 par1, Holy Fire 6 + level - 1 subtiles, S pal-skill-matrix) | D |

## 3. d2d differential (V run)

| Piece | Cases | Mismatches | Command |
|---|---|---|---|
| element min / max / length, all class skills | 24,750 (skills 6..280 x levels 1..30 x none / all1 / all20, clvl 30, masteries 0) | **0** | `run-on-worker.sh dc29_cmp.sh` |
| FCR 0..200 and IAS 0..200 rate | game side printed (1,446 lines) | not re-run against d2d here: S skiasgame2 (198,988 cases, 0) and cast-trace-gameexe (2,807 plain + 1,203 sequence lines, 0 after the cast PR) | `dc29_rate.py` |
| mana | game side printed (8,250 pairs) | not re-run here: S sk-mana-trace (130,840 cases, 0 differ with the real d2d `mana_cost`) | `dc29_skill.py mana` |
| calc slots, physical min / max | game side printed | not re-run here: S per-class cards, 0 differ (section 5) | `dc29_skill.py all` |

## 4. Tables (generated by `dc29_doc.py` from the JSON)

Damage in points (256ths / 256), no synergy unless the column says `+syn`. Mana in points.

### T1 Element damage (points) per skill level, no synergy; `+syn` = every named synergy at 20 (level 20 row)

| Skill (id, class) | anim | mana L1 / L10 / L20 / L30 | L1 | L5 | L10 | L15 | L20 | L30 | L20 +syn | length L1 / L20 (ticks) |
|---|---|---|---|---|---|---|---|---|---|---|
| Fire Bolt (36, sor, fire) | SC | 2.5 / 2.5 / 2.5 / 2.5 | 3-6 | 9-12 | 17.5-22.5 | 27.5-37.5 | 45.5-60.5 | 161.5-186.5 | 336.7-447.7 | 0 |
| Charged Bolt (38, sor, ltng) | SC | 3 / 7.5 / 12.5 / 17.5 | 2-4 | 4-6 | 6.5-8.5 | 9-11 | 13.5-15.5 | 28.5-30.5 | 29.7-34.1 | 0 |
| Lightning (49, sor, ltng) | SQ | 8 / 12.5 / 17.5 / 22.5 | 1-40 | 1-72 | 1-120 | 1-180 | 1-272 | 1-552 | 1-1577.6 | 0 |
| Chain Lightning (53, sor, ltng) | SQ | 9 / 18 / 28 / 38 | 1-40 | 1-84 | 1-143 | 1-208 | 1-281 | 1-431 | 1-955.4 | 0 |
| Frost Nova (44, sor, cold) | SC | 9 / 18 / 28 / 38 | 2-4 | 10-14 | 22-28.5 | 37-46 | 56-67.5 | 106-122.5 | 280-337.5 | 200 / 675 |
| Blizzard (59, sor, cold) | SC | 23 / 32 / 42 / 52 | 45-75 | 105-139 | 210-249 | 360-404 | 570-619 | 1120-1179 | 2280-2476 | 100 / 100 |
| Meteor (56, sor, fire) | SC | 17 / 21.5 / 26.5 / 31.5 | 80-100 | 172-200 | 319-357 | 514-562 | 869-927 | 1679-1757 | 2607-2781 | 0 |
| Fire Ball (47, sor, fire) | SC | 5 / 9.5 / 14.5 / 19.5 | 6-14 | 32-44 | 74.5-91.5 | 132-154 | 199.5-226.5 | 364.5-401.5 | 1316.7-1494.9 | 0 |
| Ice Bolt (39, sor, cold) | SC | 3 / 3 / 3 / 3 | 3-5 | 7-11 | 14-20.5 | 24-33 | 38-49.5 | 78-94.5 | 608-792 | 150 / 815 |
| Glacial Spike (55, sor, cold) | SC | 10 / 14.5 / 19.5 / 24.5 | 16-24 | 44-54 | 91-103.5 | 156-171 | 225-242.5 | 375-397.5 | 900-970 | 0 |
| Frozen Orb (64, sor, cold) | SC | 25 / 29.5 / 34.5 / 39.5 | 40-45 | 80-87 | 134-143.5 | 194-206 | 262-276.5 | 407-426.5 | 366.8-387.1 | 200 / 675 |
| Nova (48, sor, ltng) | SC | 15 / 24 / 34 / 44 | 1-20 | 25-52 | 57-94 | 92-139 | 131-188 | 221-298 | 131-188 | 0 |
| Holy Bolt (101, pal, mag) | SC | 2 / 2.5625 / 3.1875 / 3.8125 | 8-16 | 40-48 | 84-94 | 134-149 | 196-220 | 358-404 | 4116-4620 | 0 |
| Blessed Hammer (112, pal, mag) | SC | 5 / 7.25 / 9.75 / 12.25 | 12-16 | 44-48 | 88-92 | 138-142 | 196-200 | 326-330 | 1293.6-1320 | 0 |
| Fist of the Heavens (121, pal, ltng) | SC | 25 / 25 / 25 / 25 | 150-200 | 210-260 | 315-365 | 465-515 | 675-725 | 1225-1275 | 1620-1740 | 0 |
| Holy Fire (102, pal, fire) |  | 0 / 0 / 0 / 0 | 1-3 | 3-5 | 6.5-8.5 | 11.5-13.5 | 18.5-20.5 | 43.5-45.5 | 107.3-118.9 | 0 |
| Holy Shock (118, pal, ltng) |  | 0 / 0 / 0 / 0 | 1-10 | 1-34 | 1-68 | 1-108 | 1-156 | 1-278 | 1-655.2 | 0 |
| Fire Arrow (7, ama, fire) | A1 | 3 / 4.125 / 5.375 / 6.625 | 1-4 | 9-12 | 21-24 | 36-39 | 63-70 | 195-222 | 214.2-238 | 0 |
| Cold Arrow (11, ama, cold) | A1 | 3.5 / 4.625 / 5.875 / 7.125 | 3-4 | 11-12 | 22-23 | 34.5-35.5 | 53-56 | 151-160 | 180.2-190.4 | 100 / 670 |
| Lightning Bolt (20, ama, ltng) | TH | 6 / 8.25 / 10.75 / 13.25 | 1-40 | 1-88 | 1-160 | 1-250 | 1-380 | 1-900 | 1-1292 | 0 |
| Lightning Fury (35, ama, ltng) | TH | 10 / 14.5 / 19.5 / 24.5 | 1-40 | 1-120 | 1-240 | 1-390 | 1-580 | 1-1060 | 1-1044 | 0 |
| Poison Javelin (15, ama, pois) | TH | 4 / 6.25 / 8.75 / 11.25 | 0.12-0.19 | 0.38-0.44 | 0.81-0.91 | 1.44-1.61 | 2.31-2.56 | 4.94-5.22 | 7.86-8.71 | 200 / 1150 |
| Plague Javelin (25, ama, pois) | TH | 7 / 11.5 / 16.5 / 21.5 | 0.31-0.5 | 1.06-1.25 | 2.38-2.56 | 4.25-4.44 | 7.12-7.31 | 19.62-19.81 | 21.38-21.94 | 75 / 265 |
| Charged Strike (24, ama, ltng) | A1 | 4 / 6.25 / 8.75 / 11.25 | 1-30 | 1-78 | 1-146 | 1-226 | 1-322 | 1-562 | 1-2898 | 0 |
| Power Strike (14, ama, ltng) | A1 | 2 / 4.25 / 6.75 / 9.25 | 1-16 | 1-88 | 1-214 | 1-394 | 1-646 | 1-1366 | 1-5814 | 0 |
| Immolation Arrow (27, ama, fire) | A1 | 6 / 10.5 / 15.5 / 20.5 | 12-23 | 60-71 | 142-153 | 257-268 | 416-427 | 776-787 | 1248-1281 | 0 |
| Teeth (67, nec, mag) | SC | 3 / 7.5 / 12.5 / 17.5 | 2-4 | 6-8 | 11-14 | 16-21.5 | 23-31 | 43-56 | 299-403 | 0 |
| Bone Spear (84, nec, mag) | SC | 7 / 9.25 / 11.75 / 14.25 | 16-24 | 48-56 | 90-98 | 135-143 | 192-204 | 372-394 | 1267.2-1346.4 | 0 |
| Bone Spirit (93, nec, mag) | SC | 12 / 16.5 / 21.5 / 26.5 | 20-30 | 84-98 | 166-185 | 251-275 | 340-369 | 530-569 | 1972-2140.2 | 0 |
| Poison Nova (92, nec, pois) | SC | 20 / 20 / 20 / 20 | 1-1.81 | 2-2.81 | 3.5-4.31 | 5.38-6.19 | 8-8.81 | 16.38-17.19 | 40-44.06 | 50 / 50 |
| Poison Dagger (73, nec, pois) | A1 | 3 / 5.25 / 7.75 / 10.25 | 0.14-0.31 | 0.45-0.62 | 0.92-1.09 | 1.51-1.68 | 2.25-2.42 | 4.05-4.22 | 20.25-21.8 | 50 / 240 |
| Fire Trauma (251, ass, fire) | S2 | 3 / 4.125 / 5.375 / 6.625 | 3-4 | 9-14 | 21.5-32.5 | 41.5-60 | 85.5-113.5 | 277.5-335.5 | 1008.9-1339.3 | 0 |
| Fists of Fire (259, ass, fire) | SQ | 2 / 2 / 2 / 2 | 6-10 | 26-30 | 61-67 | 111-122 | 201-221 | 501-551 | 683.4-751.4 | 0 |
| Blades of Ice (274, ass, cold) | SQ | 3 / 3 / 3 / 3 | 15-35 | 47-67 | 91-111 | 141-161 | 231-259 | 531-579 | 600.6-673.4 | 100 / 290 |
| Claws of Thunder (269, ass, ltng) | SQ | 4 / 4 / 4 / 4 | 1-80 | 1-160 | 1-300 | 1-500 | 1-780 | 1-1580 | 1-2028 | 0 |
| Shock Field (256, ass, ltng) | S2 | 6 / 6 / 6 / 6 | 1-10 | 1-22 | 1-43 | 1-73 | 1-119 | 1-271 | 1-904.4 | 0 |
| Lightning Sentry (271, ass, ltng) | S2 | 20 / 20 / 20 / 20 | 1-20 | 1-60 | 1-122 | 1-202 | 1-314 | 1-654 | 1-2574.8 | 0 |
| Wake of Fire Sentry (262, ass, fire) | S2 | 13 / 13 / 13 / 13 | 5-10 | 13-18 | 25-30 | 40-45 | 63-72 | 133-152 | 264.6-302.4 | 0 |
| Firestorm (225, dru, fire) | SC | 4 / 4 / 4 / 4 | 0.05-0.09 | 0.23-0.28 | 0.53-0.61 | 0.92-1.08 | 1.44-1.67 | 3.62-4.05 | 14.66-17.05 | 0 |
| Eruption (234, dru, fire) | SC | 15 / 15 / 15 / 15 | 15-25 | 39-49 | 81-91 | 141-151 | 217-227 | 401-419 | 1258.6-1316.6 | 0 |
| Arctic Blast (230, dru, cold) | SQ | 0.375 / 0.5156 / 0.6719 / 0.8281 | 0.33-0.62 | 1.33-1.62 | 2.64-2.97 | 4.05-4.45 | 5.58-6.06 | 9.36-10.03 | 39.05-42.44 | 100 / 385 |
| Fire Claws (239, dru, fire) | A1 | 4 / 4 / 4 / 4 | 15-20 | 47-52 | 95-100 | 155-160 | 247-260 | 491-528 | 4594.2-4836 | 0 |
| Rabies (238, dru, pois) | S3 | 10 / 10 / 10 / 10 | 0.19-0.44 | 0.69-0.94 | 1.38-1.62 | 2.16-2.41 | 3.19-3.44 | 6.69-6.94 | 14.66-15.81 | 100 / 290 |
| Hurricane (250, dru, cold) | SC | 30 / 30 / 30 / 30 | 25-50 | 53-78 | 94-119 | 144-169 | 202-227 | 342-367 | 929.2-1044.2 | 50 / 50 |
| Volcano (244, dru, fire) | SC | 25 / 25 / 25 / 25 | 8-10 | 16-18 | 30-32 | 50-52 | 78-80 | 160-166 | 452.4-464 | 0 |
| Molten Boulder (229, dru, fire) | SC | 10 / 14.5 / 19.5 / 24.5 | 6-12 | 22-32 | 48-63 | 83-103 | 130-155 | 260-295 | 338-403 | 0 |
| Armageddon (249, dru, fire) | SC | 35 / 35 / 35 / 35 | 25-75 | 85-139 | 170-231 | 270-341 | 390-471 | 702-809 | 3666-4427.4 | 0 |

### T2 Weapon skills: srcdam, to-hit bonus and the calc slots (levels 1 / 10 / 20; none | +syn at 20)

| Skill (id, class) | anim | srcdam | ToHit + per level | mana L1 / L20 | calc1 | calc2 | calc3 | calc4 | calc source (calc1 ; calc2) |
|---|---|---|---|---|---|---|---|---|---|
| Attack (0, None) | A1 | 128 | None + None | None / None | 0 | 0 | 0 | 0 | `-; -` |
| Static Field (42, sor) | SC | None | None + None | 9.0 / 9.0 | 25/25/25 \| 25 | 0 | 0 | 0 | `par4; par3` |
| Corpse Explosion (74, nec) | SC | None | None + None | 15.0 / 34.0 | 70/70/70 \| 70 | 120/120/120 \| 120 | 50/50/50 \| 50 | 0 | `par1; par2` |
| Sacrifice (96, pal) | A1 | 128 | 20 + 7 | 0.0 / 0.0 | 180/315/465 \| 865 | 8/8/8 \| 8 | 0 | 0 | `ln12+skill('Redemption'.blvl)*par8+skill('Fanaticism'.blvl)*par7; par3` |
| Zeal (106, pal) | A1 | 128 | 10 + 10 | 2.0 / 2.0 | 2/5/5 \| 5 | 0/36/96 \| 336 | 0 | 0 | `min((par5 + lvl -1), par6); ((lvl < 5) ? 0 : ((lvl-4) * par4) )+skill('Sacrifice'.blvl)*pa` |
| Smite (97, pal) | S1 | None | None + None | 2.0 / 2.0 | 15/150/300 \| 300 | 15/60/110 \| 110 | 0 | 0 | `ln34; min(250,ln12)` |
| Vengeance (111, pal) | A1 | 128 | 20 + 10 | 4.0 / 8.75 | 70/124/184 \| 424 | 70/124/184 \| 424 | 70/124/184 \| 424 | 0 | `ln12+skill('Resist Fire'.blvl)*par8+skill('Salvation'.blvl)*par7; ln12+skill('Resist Cold'` |
| Charge (107, pal) | SQ | 128 | 50 + 15 | 9.0 / 9.0 | 100/325/575 \| 1375 | 0 | 0 | 0 | `ln34+(skill('Vigor'.blvl)+skill('Might'.blvl))*par8; -` |
| Bash (126, bar) | A1 | 128 | None + None | 2.0 / 2.0 | 50/95/145 \| 245 | 1/10/20 \| 20 | 0 | 0 | `ln12+skill('Stun'.blvl)*par8; ln34` |
| Double Swing (133, bar) | SQ | 128 | 15 + 5 | 1.0 / 0.0 | 0/0/0 \| 200 | 0 | 50/50/50 \| 50 | 0 | `skill('Bash'.blvl)*par8; -` |
| Frenzy (147, bar) | SQ | 128 | 100 + 7 | 1.5 / 1.5 | 90/135/185 \| 505 | 0 | 0 | 0/0/0 \| 20 | `ln12+(skill('Double Swing'.blvl) + skill('Taunt'.blvl))*par8; -` |
| Concentrate (144, bar) | A1 | 128 | 60 + 10 | 2.0 / 2.0 | 70/115/165 \| 465 | 0 | 0 | 0/0/0 \| 20 | `ln12+skill('Bash'.blvl)*par8+skill('Battle Orders'.blvl)*par7; -` |
| Whirlwind (151, bar) | SQ | 128 | None + 5 | 12.5 / 22.0 | -50/22/102 \| 102 | 0 | 0 | 0 | `ln12; -` |
| Berserk (152, bar) | A1 | 128 | 100 + 15 | 4.0 / 4.0 | 150/285/435 \| 835 | 68/41/33 \| 33 | 0 | 100/100/100 \| 100 | `ln12+(skill('Howl'.blvl) + skill('Shout'.blvl))*par8; par4-min(((110*lvl)/(lvl+6)*(par4-pa` |
| Double Throw (140, bar) | SQ | 128 | 20 + 10 | 1.0 / 1.0 | 0/0/0 \| 160 | 0 | 0 | 0 | `skill('Double Swing'.blvl)*par8; -` |
| Leap Attack (143, bar) | SQ | 128 | 50 + 15 | 9.0 / 9.0 | 100/370/670 \| 870 | 0 | 0 | 0 | `ln34+skill('Leap'.blvl)*par8; -` |
| War Cry (154, bar) | SC | None | None + None | 10.0 / 29.0 | 0 | 0 | 0 | 0 | `-; -` |
| Jab (10, ama) | SQ | 128 | 10 + 9 | 2.0 / 6.75 | -15/12/42 \| 42 | 0 | 0 | 0 | `ln34; -` |
| Impale (19, ama) | SQ | 128 | 100 + 25 | 3.0 / 3.0 | 300/525/775 \| 775 | 46/30/25 \| 25 | 1/1/1 \| 1 | 0 | `ln12; par6-dm34` |
| Fend (30, ama) | A1 | 128 | 40 + 10 | 5.0 / 5.0 | 12/12/12 \| 12 | 70/160/260 \| 260 | 0 | 0 | `12; ln34` |
| Strafe (26, ama) | A1 | 96 | None + None | 11.0 / 11.0 | 4/10/10 \| 10 | 5/50/100 \| 100 | 2/4/7 \| 7 | 0 | `min(par3 + lvl - 1, par4); ln12` |
| Multiple Shot (12, ama) | A1 | 96 | None + None | 4.0 / 23.0 | 2/11/21 \| 21 | 1/1/1 \| 1 | 2/2/2 \| 2 | 0 | `min(24,ln12); par3` |
| Guided Arrow (22, ama) | A1 | 128 | None + None | 8.0 / 3.25 | 0/45/95 \| 95 | 0 | 0 | 0 | `ln34; -` |
| Tiger Strike (254, ass) | A1 | 128 | 15 + 7 | 1.0 / 1.0 | 100/280/480 \| 480 | 0 | 0 | 0 | `ln12; -` |
| Dragon Claw (260, ass) | SQ | 128 | 40 + 25 | 2.0 / 2.0 | 50/95/145 \| 225 | 0 | 0 | 0 | `ln12 + skill('Claw Mastery'.blvl)*par8; -` |
| Dragon Talon (255, ass) | KK | None | 20 + 35 | 6.0 / 6.0 | 1/2/4 \| 4 | 57/84/92 \| 92 | 36/75/87 \| 87 | 36/75/87 \| 87 | `lvl/6+1; dm34` |
| Cobra Strike (265, ass) | A1 | 128 | 15 + 7 | 2.0 / 2.0 | 0 | 0 | 0 | 0 | `-; -` |
| Royal Strike (280, ass) | A1 | 128 | 15 + 7 | 4.0 / 4.0 | 6/6/6 \| 6 | 0 | 0 | 0 | `par2; -` |
| Blade Fury (266, ass) | SQ | 96 | None + None | 1.0 / 3.375 | 0 | 0 | 0 | 0 | `-; -` |
| Fury (248, dru) | A1 | 128 | 50 + 7 | 4.0 / 4.0 | 2/5/5 \| 5 | 100/253/423 \| 423 | 0 | 0 | `min((par5 + lvl -1), par6); ln34` |
| Maul (233, dru) | A1 | 128 | 20 + 10 | 3.0 / 3.0 | 0 | 3/8/13 \| 13 | 0 | 0 | `0; lvl/par7 + par8` |
| Feral Rage (232, dru) | A1 | 128 | 20 + 10 | 3.0 / 3.0 | 50/95/145 \| 145 | 3/8/13 \| 13 | 0 | 0 | `ln56; lvl/par7 + par8` |
| Hunger (242, dru) | S3 | 128 | 50 + 10 | 3.0 / 3.0 | -75/-75/-75 \| -75 | 72/152/176 \| 176 | 72/152/176 \| 176 | 0 | `par5; dm12` |
| Shock Wave (243, dru) | SC | None | None + None | 7.0 / 7.0 | 5/5/5 \| 5 | 0 | 0 | 0 | `5; -` |

### T3 Cast frames and FCR breakpoints (plain cast = `ceil(256 N / rate) - 1`, V run rate, see doc)

| Class | SC frames (HTH) | breakpoints (FCR>= : ticks) |
|---|---|---|
| Amazon | 20 | 0: 19, 7: 18, 14: 17, 22: 16, 32: 15, 48: 14, 68: 13, 99: 12, 152: 11 |
| Sorceress | 14 | 0: 13, 9: 12, 20: 11, 37: 10, 63: 9, 105: 8, 200: 7 |
| Necromancer | 16 | 0: 15, 9: 14, 18: 13, 30: 12, 48: 11, 75: 10, 125: 9 |
| Paladin | 16 | 0: 15, 9: 14, 18: 13, 30: 12, 48: 11, 75: 10, 125: 9 |
| Barbarian | 14 | 0: 13, 9: 12, 20: 11, 37: 10, 63: 9, 105: 8, 200: 7 |
| Druid | 15 (anim speed 208) | 0: 18, 4: 17, 10: 16, 19: 15, 30: 14, 46: 13, 68: 12, 99: 11, 163: 10 |
| Assassin | 17 | 0: 16, 8: 15, 16: 14, 27: 13, 42: 12, 65: 11, 102: 10, 174: 9 |

### T4 Attack frames (A1) per weapon class and IAS breakpoints (WSM 0; `ceil(256 N / rate)`)

| Class | wclass | A1 frames | breakpoints (IAS>= : ticks) |
|---|---|---|---|
| Amazon | 1HS | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Amazon | 2HS | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Amazon | BOW | 14 | 0: 14, 9: 13, 20: 12, 37: 11, 63: 10, 105: 9, 200: 8 |
| Amazon | HTH | 13 | 0: 13, 10: 12, 23: 11, 42: 10, 72: 9, 133: 8 |
| Amazon | STF | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Amazon | XBW | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Sorceress | 1HS | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Sorceress | 2HS | 24 | 0: 24, 6: 23, 11: 22, 18: 21, 26: 20, 35: 19, 48: 18, 65: 17, 86: 16, 125: 15, 180: 14 |
| Sorceress | BOW | 17 | 0: 17, 8: 16, 16: 15, 27: 14, 42: 13, 65: 12, 102: 11, 174: 10 |
| Sorceress | HTH | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Sorceress | STF | 18 | 0: 18, 8: 17, 15: 16, 26: 15, 39: 14, 58: 13, 86: 12, 138: 11 |
| Sorceress | XBW | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Necromancer | 1HS | 19 | 0: 19, 7: 18, 15: 17, 23: 16, 35: 15, 52: 14, 78: 13, 117: 12, 194: 11 |
| Necromancer | 2HS | 23 | 0: 23, 6: 22, 11: 21, 19: 20, 27: 19, 39: 18, 52: 17, 70: 16, 99: 15, 142: 14 |
| Necromancer | BOW | 18 | 0: 18, 8: 17, 15: 16, 26: 15, 39: 14, 58: 13, 86: 12, 138: 11 |
| Necromancer | HTH | 15 | 0: 15, 9: 14, 19: 13, 32: 12, 54: 11, 86: 10, 152: 9 |
| Necromancer | STF | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Necromancer | XBW | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Paladin | 1HS | 15 | 0: 15, 9: 14, 19: 13, 32: 12, 54: 11, 86: 10, 152: 9 |
| Paladin | 2HS | 18 | 0: 18, 8: 17, 15: 16, 26: 15, 39: 14, 58: 13, 86: 12, 138: 11 |
| Paladin | BOW | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Paladin | HTH | 14 | 0: 14, 9: 13, 20: 12, 37: 11, 63: 10, 105: 9, 200: 8 |
| Paladin | STF | 18 | 0: 18, 8: 17, 15: 16, 26: 15, 39: 14, 58: 13, 86: 12, 138: 11 |
| Paladin | XBW | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Barbarian | 1HS | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Barbarian | 2HS | 18 | 0: 18, 8: 17, 15: 16, 26: 15, 39: 14, 58: 13, 86: 12, 138: 11 |
| Barbarian | BOW | 15 | 0: 15, 9: 14, 19: 13, 32: 12, 54: 11, 86: 10, 152: 9 |
| Barbarian | HTH | 12 | 0: 12, 11: 11, 26: 10, 48: 9, 86: 8, 180: 7 |
| Barbarian | STF | 19 | 0: 19, 7: 18, 15: 17, 23: 16, 35: 15, 52: 14, 78: 13, 117: 12, 194: 11 |
| Barbarian | XBW | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Druid | 1HS | 19 | 0: 19, 7: 18, 15: 17, 23: 16, 35: 15, 52: 14, 78: 13, 117: 12, 194: 11 |
| Druid | 2HS | 21 | 0: 21, 7: 20, 13: 19, 20: 18, 30: 17, 44: 16, 63: 15, 86: 14, 129: 13, 200: 12 |
| Druid | BOW | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Druid | HTH | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Druid | STF | 17 | 0: 17, 8: 16, 16: 15, 27: 14, 42: 13, 65: 12, 102: 11, 174: 10 |
| Druid | XBW | 20 | 0: 20, 7: 19, 14: 18, 22: 17, 32: 16, 48: 15, 68: 14, 99: 13, 152: 12 |
| Assassin | 1HS | 15 | 0: 15, 9: 14, 19: 13, 32: 12, 54: 11, 86: 10, 152: 9 |
| Assassin | 2HS | 23 | 0: 23, 6: 22, 11: 21, 19: 20, 27: 19, 39: 18, 52: 17, 70: 16, 99: 15, 142: 14 |
| Assassin | BOW | 16 | 0: 16, 9: 15, 18: 14, 30: 13, 48: 12, 75: 11, 125: 10 |
| Assassin | HTH | 11 | 0: 11, 13: 10, 29: 9, 56: 8, 113: 7 |
| Assassin | STF | 19 | 0: 19, 7: 18, 15: 17, 23: 16, 35: 15, 52: 14, 78: 13, 117: 12, 194: 11 |
| Assassin | XBW | 21 | 0: 21, 7: 20, 13: 19, 20: 18, 30: 17, 44: 16, 63: 15, 86: 14, 129: 13, 200: 12 |

## 5. Deviations d2d vs game.exe that move dps (carried from earlier cards; none new here)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | Zeal: game.exe starts the next strike `action` ticks later (7 at 0 IAS); d2d the full swing (15). The fork lineage carries the upstream fix (Dravis `276f85c0`, residual P01) | V run both sides (pal-zeal-trace.md); status of upstream S (pipeline row M4) | exists (pal-zeal-trace.md, `zd_*`) |
| D2 | Idle tick between held swings in d2d (pal-zeal-trace D2) | S | exists |
| D3 | Sequence skills (SQ) through the attack branch: game.exe takes 30 points off the rate; d2d none (Jab, Impale, Double Swing, Double Throw, Frenzy) | V read + V run game side (skiasgame2 D1); d2d V read | buildable: `Fight::start_sequence` probe from `ms_build.sh`, inputs frames 8..24 x rate, oracle `skiasgame2_rate.py seq` (pipeline P17) |
| D4 | Smite: stat 111 omitted, min = max, rounding, block cancels Smite in d2d | V run 20,500 cases per seed (palskill-trace.md D1 to D5), D8 V read | exists (pal2_test); pipeline P18 |
| D5 | Holy Bolt: allies are healed in game.exe, d2d has no heal | V read / V run game side (pal-skill-trace.md D1) | buildable (friendly target probe); pipeline P19 |
| D6 | Static Field: no difficulty floor, no "life - 1", a negative lightning resistance raises d2d's damage | V read (skill-damage-trace.md) + V run chill (static-field-chill-trace.md) | exists for the chill; the damage rules need `static_field.py` |
| D7 | Plain cast takes one tick less than a sequence (cast PR, merged); no deviation left (cast-trace-gameexe.md) | S | n/a |

Per-class calc / physical sources (S): Paladin palskill-trace.md (40,720 cases), Sorceress sor-rest-trace.md and sorc-skill-matrix.md, Necromancer nec-skill-trace.md (27,800), Barbarian skbarb-trace.md (27,800), Amazon skamazon-trace.md (27,800), Druid skdruid-trace.md, Assassin skassassin-trace.md. **Not re-run here** (only the element grid and the rates were compared with d2d this card).

## 6. Bot findings (described, not changed)

| B | Finding |
|---|---|
| B1 | `internal/power/profiles.json:270` `swingSeconds` 1.28 (tag U) is used for every hero with a readable weapon (`internal/power/margin.go:133`, `hit * mean / pr.SwingSeconds`). Real: Paladin 1HS 15 ticks = 0.60 s, Sorceress staff 18 ticks = 0.72 s, Barbarian 1HS 16 ticks = 0.64 s at 0 IAS (T4). Offence is understated by about 2x at 0 IAS and the gap grows with IAS. |
| B2 | `profiles.json:271` `defaultDpsPerClvl` 2.5 (`margin.go:141`, `* h.Level`): a clvl-24 Sorceress gets 60 dps; Fire Bolt 20 with no synergy deals 102 dps (T1 + T3, I) and Fire Ball 20 about 410. The spell side is not modelled at all, so any caster margin is dominated by the default flag. |
| B3 | `profiles.json` `manaPerSecond` 4.0 (U): Fire Bolt costs 2.5 mana per 13 ticks = 4.8 mana per second (0 FCR); Frost Nova 17 at L1; Blizzard 44 (delay-bound, so 23 to 81 / 1.8 s). One number for all spells hides a factor of 10. |
| B4 | The power model has no skill-delay term (Blizzard 45, Meteor 30 ticks); a delay-bound skill is `max(delay, cast ticks)` per cast (I). |

## 7. How the power estimator should use `skilldamage.json` (follow-up card; not wired here)

Schema `power.skilldamage/1` (all numbers are game.exe values; Go reads it with `encoding/json`, no table is mirrored into code):

```
version, schema, units, assumptions, formulas
rate.attack_wsm0[ias 0..200]        FUN_00623f50 attack rate at WSM 0, AnimData speed 256
rate.attack_by_wsm["-20"]..         the same for other WSM (stat 68 = 100 - WSM)
rate.cast[fcr 0..200]               FUN_00623f50 cast rate (stat 105), speed 256
frames.<class>.<mode>.<wclass> = {frames, speed}   AnimData; class = Skills.txt charclass (ama sor nec pal bar dru ass)
skills.<Skill name> = {id, class, anim, seqtrans, rate_kind (cast | attack | attack_sq), kind (power.profiles skillKinds or null),
   srcdam, uses_attack_rate, tohit_pct, tohit_per_level, etype, hitshift, req_level, delay_frames,
   mana_cols, mana[level], calc_src[4], calc_desc[4], params, levels[1..30], missile{name,pierce,vel,range,explosion},
   elem.{none,all1,all20}.{min,max,len}[level] (256ths), phys.{none,all20}.{min,max}[level] (256ths),
   calc.{none,all1,all20}[level][4]}
```

Key by skill NAME from the character (the skill the combat profile picks), never by id; the class code comes from the entry. To compute damage per second by archetype:

- **Caster (element spell, `rate_kind` cast):** `avg = (elem.S.min + elem.S.max) / 512` points at the hero's level (+ item skills) and synergy set S chosen from the hero's actual synergy points (linear: `none + p/20 x (all20 - none)` is an upper bound since `all20` assumes 20 points of each named synergy; exact when the hero's points are known, evaluate the table row for the real set); `ticks = ceil(256 N / rate.cast[fcr]) - 1` (sequence skills, `anim` SQ with `seqtrans` SC: no - 1; Lightning and Chain Lightning 19 steps for N, S cast-trace-gameexe); `period = max(ticks, delay_frames) / 25` s; `dps_one = avg x (1 - res) / period` (masteries: x (1 + mastery / 100)); bolts per cast: `calc[0]` for Charged Bolt (each bolt a roll; hit share U), `calc[0]` count for Zeal. Mana: `mana[level] / period` per second; a pool P lasts `P / (mana per second)`.
- **Bruiser / ranged physical (`attack`):** `hit = weapon_avg x (srcdam / 128) x (100 + ED + stat% + skill ED) / 100` where ED is the calc slot named by `calc_src` (Sacrifice calc1, Zeal calc2, Vengeance calc1 to 3 per element, Bash calc1 ...), x `hit_chance` (to-hit with the skill's `tohit_pct + tohit_per_level (L - 1)`, S sk-melee-trace); `ticks = ceil(256 N / rate.attack_wsm0[ias])`; multiple strikes per cast (Zeal 2 to 5, Fury) at the per-strike spacing of the side being modelled (game.exe 7 ticks, d2d 15: D1); `dps = hits x hit / (ticks / 25)`.
- **Summoner and aura builds:** the table gives the master's own attack; summons need the monster tables (not in this file, U).
- **Quality flags:** the per-skill damage and the rates are `measured`-grade (game.exe values); the hero's own stats (ED gear, IAS, FCR, weapon) keep their existing term quality; a missing FCR / IAS read keeps `derived` at 0 (the table's first column).

Examples (I; one target, no resist, 0 FCR, no synergy, `dc29_doc.py` arithmetic): Fire Bolt L1 8.7, L10 38.5, L20 101.9 dps at 4.81 mana per second; Ice Bolt L20 84 dps; Fire Ball L20 410 dps at 27.9 mana per second; Frost Nova L20 119; Glacial Spike L20 450; Lightning L20 180 (19 ticks, min 1); Blizzard L20 about 1143 before the 45-tick delay (about 330 with it: 594 per 1.8 s, I).

## 8. Unverified (U) and dead ends

- U1: `seqtrans` of Inferno / Arctic Blast (SQ, not SC): what rate they get is U (cast-trace-gameexe.md line 38); `rate_kind` `attack_sq` marks them.
- U2: the number of sequence steps for sequence skills other than Lightning / Chain Lightning (`seqnum`, sequences.d2, not read).
- U3: hit share of multi-bolt skills (Charged Bolt, Lightning Fury, Strafe) against one target: needs the missile flight (SKMISS, S) and the monster size; the table gives per-bolt damage and the count.
- U4: whether the druid's cast speed 208 enters Skills `anim` for all druid casts (formula linear in speed, V read skiasgame2 Q2; the rig ran speed 256 only): T3's Druid row is I.
- U5: the physical min / max and calc slots were run through game.exe this card but only compared with d2d in earlier cards (0-diff per class cards, S); no new differential for them.
- Dead end: the NAS builder container has no emulator image (`labrun tower`: no `~/d2d-lab/emu`), so the run used the worker host under `tools/emu/run-on-worker.sh` (a niced single process, well under 4 cores, `agentreap` list empty at the end).

## 9. PR candidates

No PR candidate: the element differential is 0 mismatches before and after (nothing to fix); the dps-moving deviations (D1 to D5) were not proven by the existing differential in this card (rules of the briefs: before > 0, after 0). Pipeline rows P17 to P19 added to [../upstream/pipeline-2026-11-14.md](../upstream/pipeline-2026-11-14.md); no branch, nothing pushed.

## 10. Drill recipe (lab, one hidden slot, `docs/forkdbg.md`, `docs/forkcheats.md`; do not run while the user plays)

1. Fresh Sorceress with Fire Bolt at level N: `debug cast 36 near` (forkdbg) against `debug mon set <id> life 5000 maxlife 5000 res fire 0`; read the target's life before and after one cast through `dump monsters`: the loss in 256ths must lie in `elem.none` min..max of the JSON at level N (U confirm; the d2d side is already 0-diff, this checks the fork build the bot runs).
2. Cast period: 20 casts of `debug cast 36` with the d2d tick counter (`dump stats` frame field); `ticks per cast` must be 13 at FCR 0 (T3), 12 with FCR 9.
3. Paladin swing: `debug cast 0` with a one-hand sword; ticks per swing 15 at IAS 0 (T4).
4. Mana: `debug mana full` then one cast; `mana before - after` = `mana[level]` of the JSON.
5. Synergy: `debug points <skill id> 20` on the named synergy skills; the next cast's damage must lie in `elem.all20` range (the points command is lab-only, forkcheats).

## 11. Matrix-cell updates (proposed text)

- `sorc-skill-matrix.md` damage cells of Fire Bolt, Ice Bolt, Charged Bolt, Fire Ball, Glacial Spike, Blizzard, Meteor, Frost Nova, Lightning, Chain Lightning, Static Field: replace the `[I]` per-level ranges by `[V run, decomp29-skilldamage.md T1; internal/power/skilldamage.json]`; cast and mana cells: `[V run, T3 and mana[]]` (Sorceress SC 14 frames, 13 ticks plain, FCR breakpoints 9 / 20 / 37 / 63 / 105 / 200). Row K7 (delay not modelled): add the `delay_frames` values as D.
- `pal-skill-matrix.md` section 2 (rates): `15 ticks per swing` stays; add `Paladin 1HS breakpoints IAS 9, 19, 32, 54, 86, 152 (V run rate, T4)`; section 3.6 Holy Bolt damage ranges 8-16, 40-48, 84-94 at L1, 5, 10 and 196-220 at L20 (V run, T1, replaces the earlier L1 to 4 only).
- `docs/survival/power.md` "What is readable today": spell and skill damage row: `readable as data from internal/power/skilldamage.json (V run); wiring pending (card to follow)`.

## 12. Glossary

Missing term for GLOSSARY.md (proposal only): "synergy set" (a named assignment of skill levels to the other skills of a class when evaluating a skill's damage).
