# DROPMODEL: what the Act 1 Normal areas, monsters and bosses drop (game.exe treasure rolls as data)

> status: research, 2026-11-15, agent decomp31 (task 505). Offline: no game, no d2d slot, nothing posted, no Go changed. Builds on treasure-trace.md (the roll, classes, gold), areamon-trace.md (the room populations) and xp-gold-trace.md. Tags: V run (game.exe's own code under the emulator, `$D2_EMU_DIR` image of docs/emu-image.md, compared with an exact oracle or with d2d's rule), V read (disassembly, in the traces named), D (patched 1.14d tables via the emulator's loader), S (an earlier trace of ours), I inferred, U unverified. Consumer cards: levelsim `-drops`, the Where-to line, the loot-rules examples (section 9). Data: `internal/plandata/dropmodel-normal.json` (schema `dropmodel/1`).

## Answer in eight lines

1. **A Normal Act 1 area is a potion and a magic-item source mostly because of its few special units, not its trash.** Cold Plains per clear: the one superunique (Bishibosh) is 22 % of the healing potions and 39 % of the magic or better items; the 0.8 uniques add 18 % and 16 %; 191 plain kills add 55 % and 41 % (V run, section 5). Cave 2 and the Catacombs are 63 % unique-driven.
2. **Potion supply per minute (healing / mana, expected, one player, MF 0)**: Blood Moor 0.25 / 0.13; Cold Plains 0.86 / 0.47; Stony Field 0.85 / 0.48; Dark Wood 1.11 / 0.60; Black Marsh 0.84 / 0.53; Cave 2 1.18 / 0.57; Catacombs 2 0.94 / 0.51; Catacombs 3 1.02 / 0.54 (the rate is the time model of areamon, I; the per-kill counts are V run). Per 1,000 kills the overworld is 35 to 76 healing and 18 to 43 mana, the dungeon floors with 2 to 3 uniques 110 to 205 (section 4).
3. **Item quality is poor in Normal Act 1.** 58 to 88 percent of gear drops are normal or low (88 in Blood Moor, 73 to 78 in Cold Plains and Stony Field, 58 to 62 in Cave 2 and the Catacombs), 5 to 8 percent superior; magic 2 to 21 per 1,000 kills on the overworld, 37 to 46 in Cave 2 and the Catacombs; rare 0.9 to 4.3, set 0.6 to 1.9, unique 0.1 to 0.8 per 1,000 kills (table 3). Every base is a clvl 1 to 10 base: all gear dropped in Act 1 Normal has a required level of at most 10 except 3.6 per 1,000 kills in Catacombs 3 (V run): a clvl 15 to 25 hero finds magic and rare gear here, never a base for its level.
4. **Superuniques are the supply points.** Every Act 1 Super A/B/C boss drops exactly 6 items: 4 potions (2.6 healing, 1.2 mana, 0.2 rejuvenation), 1.4 magic, 0.37 rare, 0.05 set and 0.02 unique gear per kill (V run, section 6). Griswold and the Smith drop 3 items, no potions. Blood Raven 3.4 items, almost no potions (0.18). Andariel 4.9 items (0.62 potions), first kill 3.7 items and no potions.
5. **The Countess is the rune source of Act 1: 1.89 runes per kill, El to Ral** (Tal 0.46, Ral 0.31, Eth 0.33, Tir 0.24). The 2.28 of treasure-trace.md section 5 is the exact odds before the 6-item cap; the native roll with the cap is 1.89 (V run, 20,000 rolls, z -75 against the uncapped exact odds; the class has P(7 or more drops) = 0.32). Outside her room no Act 1 area gives more than 1.2 runes per 1,000 kills (Catacombs 3 and 4).
6. **d2d's roll equals game.exe's on this exact key set**: 102 (class, item level) keys of the population, 40,800 rolls (one and three players, MF 0 and 300): 40,800 equal, 0 differ (V run, section 7). The class choice was already equal on 78,309 cases (S, treasure-trace.md). No d2d deviation of the drop roll; the open deviations are about WHICH units d2d spawns (D1, U).
7. **Bot findings**: the tables the survival economy and the simulator use credit no item supply (B1, B2); `data.SupplyOf` is not read by any code (B1). No PR candidate.
8. **Not done**: chests, barrels and urns per area (counts and the area-to-tier map; only the per-round class tables are in the data, section 6), magic find above 0, more than one player, and every U in section 8.

## Q1..: questions and answers

| id | question | answer | tag |
|---|---|---|---|
| Q1 | Which class does a unit roll? | superunique flag (data +0x16 bit 2): the SuperUniques row `TC`; else flag 4 (champion): MonStats TreasureClass2; flag 8 (unique): TreasureClass3; else TreasureClass1 (a unique's minions carry flag 0x10 only, so TC1). Normal: no level move. Same rule as treasure-trace.md Q1; the harness `dm_diff.py` reuses it | V read + V run (S: `tre_select.py`, 78,309 equal) |
| Q2 | Which item level? | the unit's stat 12 (at least 1): Normal = the MonStats Level column, champion +2, unique and superunique +3 (the population's own stat, read from the unit lines) | V run (population), S (rule) |
| Q3 | What does one roll make? | `FUN_0055a6d0` natively, 20,000 rolls per distinct (class, ilvl) key, 117 keys including the quest classes; the made item's code and quality are tallied. Exact odds of the class records (`tre_act.Table.one`) agree: 1,521 cells; 3 cells above |z| 3.5: the Countess rune cell (-75) and the Andariel gold cell (-3.9) are the 6-item cap, which the exact odds do not model; the third (Champ C jewels, 0.0015 against 0.0008 expected) is noise at that probability, the next largest |z| is 3.0 | V run |
| Q4 | Which units does an area hold? | game.exe's own room population, 12 seeds, 36 Act 1 levels (areamon method: `ms_spawn.py`, `FUN_00573cb0` stat writes): class, flags, level of every unit with experience above 0 | V run |
| Q5 | What drops per kill, per clear, per minute? | sections 4 and 5; per minute = per clear / the clear time of `internal/levelplan/areaxp.csv` (a TIME MODEL, I) | V run x I |
| Q6 | Who drops the potions and the good gear? | section 5: the few superuniques, uniques and champions; plain kills are 14 to 55 % of the supply | V run |
| Q7 | Which bases? | Weapons.txt / Armor.txt required level of the base: all Act 1 Normal gear is levelreq at most 10 (3.6 per 1,000 kills in Catacombs 3 are 11 to 25); a base the hero can wear at clvl 25 is every drop, strength and dexterity not modelled | D + V run |
| Q8 | What do bosses drop? | section 6 | V run |
| Q9 | Chests, barrels? | per-round class tables only (section 6); area counts and the area-to-tier map U | V run (rolls), U (counts) |

## 1. Method

1. Population: `bash tools/emu/am_batch.sh POP 1 12 3` (areamon, 36 levels, Normal, expansion set); one line per unit `U lid seed n class flags statlvl f7 stats`.
2. `tools/emu/dm_model.py POP OUT.json 20000` (run as `bash tools/emu/tre_run.sh dm_model.py ...` on the worker host): per unit the class of Q1 and ilvl of Q2, a native roll per distinct key (`drops.Oracle.roll`: `FUN_00654e00`, `FUN_0055a6d0`, one player, MF 0, GF 0; made items hooked at `FUN_0055a550`), tallied by item code and quality, then the area sums over the 12 seeds, the monster rows and the boss rows. The exact odds `tre_act.Table.one` are compared cell by cell (hp, mp, scroll, gear, rune, gem, jewelry, gold, key, rejuvenation).
3. `tools/emu/dm_pack.py` turns the output into `internal/plandata/dropmodel-normal.json`; `tools/emu/dm_diff.py` rolls the same 102 keys through d2d's `drop-dump` (pristine fork build) and game.exe.
4. Classes of interest by item code: potions by tier (hp1..5, mp1..5, rejuvenation, stamina and others), town portal and identify scrolls, tomes, keys, gems by grade, runes by number, rings and amulets, jewels, charms, weapons and armor by quality (plain = low and normal, superior, magic, rare, set, unique) and by base required level (early at most 10, mid 11 to 25, late above 25). Quality is the quality asked when the item is made (`FUN_00558640`).
5. Gold is excluded (xp-gold-trace.md); one player; no magic find, no gold find.

## 2. What a plain kill gives, by monster (per 1,000 kills, V run)

Per monster name, plain kills only, at the item level the area gave it (`monsters` in the data, `name:kind` keys). Healing, mana, town portal scrolls, magic or better items, rare or better items per 1,000 kills:

| monster | class | ilvl | seen | hp | mp | TP | magic+ | rare+ |
|---|---|---|---|---|---|---|---|---|
| fallen1 / zombie1 | Act 1 H2H A | 1 | 3,587 / 458 | 36 | 18 | 7 | 10.2 | 2.05 |
| fallen2 / brute2 / corruptrogue2 | H2H B | 5 | 5,593 | 33 | 20 | 9 | 11.7 | 1.75 |
| fallen3 | H2H B | 7.1 | 2,562 | 37 | 19 | 7 | 12.7 | 2.32 |
| fallen4 | H2H C | 10.1 | 2,417 | 33 | 21 | 7 | 12.4 | 2.78 |
| fallenshaman1 / foulcrow1 | Cast A | 2 / 4 | 660 / 334 | 28 / 30 | 38 | 15 | 11.2 / 10.8 | 1.55 / 2.10 |
| fallenshaman3 / bighead2 | Cast C | 9 / 11 | 218 / 448 | 28 / 31 | 40 | 13 / 15 | 11.6 / 10.5 | 2.2 / 1.5 |
| cr_archer1 / sk_archer1 | Miss A / B | 4 / 5 | 281 / 509 | 30 | 15 / 17 | 7 / 6 | 11.4 / 11.6 | 2.45 / 2.6 |
| skeleton2 / skmage_fire2 | H2H B / Cast C | 6 / 10 | 394 / 349 | 36 / 30 | 19 / 38 | 8 / 14 | 11.2 / 10.8 | 1.9 / 1.75 |
| quillrat1..2 | Quill 1 / 2 | 1 / 5 | 254 / 224 | 34 / 32 | 18 | 7 | **0** | **0** |
| quillrat3 | Quill 3 | 8 | 390 | 31 | 18 | 6 | 3.0 | 0.05 |

A plain kill gives 0.050 to 0.070 potions (finding: the Cast classes carry twice the mana and twice the scrolls); the magic or better share of the gear is the same 10 to 13 per 1,000 kills in every class (I: the quality odds are those of the Act 1 base items at these item levels). Quill rats (own classes `Quill 1..4`) drop potions and scrolls but no good gear.

## 3. Quality of the gear drops (per 1,000 kills of the area, weapons and armor only)

| area | plain | superior | magic | rare | set | unique | early bases | mid | late |
|---|---|---|---|---|---|---|---|---|---|
| Blood Moor | 79.1 | 6.7 | 2.2 | 0.86 | 0.55 | 0.12 | 89.5 | 0.0 | 0.0 |
| Cold Plains | 95.4 | 8.0 | 14.1 | 3.02 | 0.91 | 0.40 | 121.8 | 0.0 | 0.0 |
| Stony Field | 94.0 | 7.8 | 20.9 | 3.67 | 1.16 | 0.66 | 128.1 | 0.0 | 0.0 |
| Dark Wood | 87.7 | 7.0 | 16.1 | 2.55 | 0.98 | 0.68 | 115.0 | 0.0 | 0.0 |
| Black Marsh | 95.9 | 7.6 | 11.7 | 1.21 | 0.85 | 0.48 | 117.7 | 0.0 | 0.0 |
| Tamoe Highland | 82.4 | 6.5 | 11.3 | 1.31 | 0.80 | 0.41 | 102.8 | 0.0 | 0.0 |
| Cave Level 1 | 95.6 | 8.0 | 23.2 | 4.29 | 1.17 | 0.59 | 132.9 | 0.0 | 0.0 |
| Cave Level 2 | 91.7 | 7.7 | 44.0 | 3.04 | 1.83 | 0.79 | 149.0 | 0.0 | 0.0 |
| Catacombs Level 2 | 89.1 | 7.1 | 37.2 | 2.93 | 1.76 | 0.68 | 138.8 | 0.0 | 0.0 |
| Catacombs Level 3 | 84.1 | 7.2 | 45.6 | 3.25 | 1.94 | 0.83 | 139.3 | 3.6 | 0.0 |
| Jail Level 1 | 84.9 | 6.9 | 31.2 | 2.56 | 1.54 | 0.59 | 127.7 | 0.0 | 0.0 |
Rings, amulets, jewels and charms are in the data (`jewelry_*`, `jewel_any`, `charm`); they are 5 to 10 percent of the magic or better count. "Plain" folds low and normal; set 0.6 to 1.9 and unique 0.1 to 0.8 per 1,000 kills come from the quality roll of every class, not from a special class.

## 4. Per area (all 36 Act 1 levels, per 1,000 kills and per minute; V run counts, I time)

hp, mp, rejuv, TP scroll, magic+, rare+, set/unique are per 1,000 kills; the five per-minute columns multiply the same counts by kills per minute (units per clear over the clear time of areaxp.csv, I). "usable25" = magic or better weapons and armor on a base a clvl 25 hero can wear by level.

| lid | area | units | kills/min (I) | hp | mp | rejuv | TP | magic+ | rare+ | set/unique | HP/min | MP/min | magic+/min | rare+/min | usable25/min |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 2 | Blood Moor | 97 | 7.13 | 35 | 18 | 2.3 | 6.8 | 8.0 | 1.6 | 0.7 | 0.25 | 0.13 | 0.06 | 0.011 | 0.03 |
| 3 | Cold Plains | 198 | 14.56 | 59 | 32 | 4.5 | 8.2 | 25 | 4.5 | 1.4 | 0.86 | 0.47 | 0.36 | 0.065 | 0.27 |
| 4 | Stony Field | 152 | 11.21 | 76 | 43 | 6.0 | 9.2 | 33 | 5.6 | 1.9 | 0.85 | 0.48 | 0.37 | 0.063 | 0.30 |
| 5 | Dark Wood | 231 | 17.02 | 65 | 35 | 5.3 | 8.4 | 27 | 4.3 | 1.7 | 1.11 | 0.60 | 0.46 | 0.073 | 0.35 |
| 6 | Black Marsh | 199 | 14.65 | 58 | 36 | 4.6 | 9.6 | 21 | 2.6 | 1.4 | 0.84 | 0.53 | 0.30 | 0.038 | 0.21 |
| 7 | Tamoe Highland | 194 | 14.26 | 59 | 33 | 4.5 | 7.8 | 20 | 2.6 | 1.3 | 0.84 | 0.47 | 0.29 | 0.037 | 0.20 |
| 8 | Den of Evil | 77 | 6.34 | 68 | 36 | 5.3 | 7.8 | 36 | 7.7 | 1.9 | 0.43 | 0.23 | 0.23 | 0.049 | 0.18 |
| 9 | Cave Level 1 | 124 | 7.27 | 82 | 42 | 6.4 | 7.8 | 36 | 6.2 | 1.9 | 0.60 | 0.30 | 0.26 | 0.045 | 0.21 |
| 10 | Underground Passage Level 1 | 195 | 7.96 | 59 | 33 | 4.8 | 8.6 | 21 | 2.7 | 1.5 | 0.47 | 0.27 | 0.17 | 0.021 | 0.11 |
| 11 | Hole Level 1 | 124 | 7.27 | 64 | 37 | 5.3 | 8.8 | 23 | 2.8 | 1.6 | 0.47 | 0.27 | 0.17 | 0.021 | 0.12 |
| 12 | Pit Level 1 | 148 | 7.57 | 63 | 32 | 4.8 | 7.3 | 23 | 3.2 | 1.4 | 0.48 | 0.24 | 0.17 | 0.024 | 0.12 |
| 13 | Cave Level 2 | 46 | 7.25 | 162 | 78 | 14 | 7.4 | 57 | 5.8 | 2.7 | 1.18 | 0.57 | 0.41 | 0.042 | 0.36 |
| 14 | Underground Passage Level 2 | 48 | 7.32 | 153 | 77 | 13 | 7.9 | 53 | 5.6 | 2.8 | 1.12 | 0.56 | 0.39 | 0.041 | 0.33 |
| 15 | Hole Level 2 | 36 | 6.85 | 164 | 83 | 14 | 8.7 | 59 | 6.0 | 3.1 | 1.12 | 0.57 | 0.41 | 0.041 | 0.35 |
| 16 | Pit Level 2 | 50 | 7.4 | 140 | 68 | 12 | 7.2 | 50 | 5.5 | 2.5 | 1.03 | 0.51 | 0.37 | 0.041 | 0.31 |
| 17 | Burial Grounds | 38 | 6.95 | 35 | 21 | 1.9 | 8.7 | 55 | 12 | 2.5 | 0.24 | 0.15 | 0.38 | 0.083 | 0.31 |
| 18 | Crypt | 29 | 4.1 | 192 | 91 | 16 | 7.4 | 103 | 19 | 4.5 | 0.78 | 0.37 | 0.42 | 0.080 | 0.38 |
| 19 | Mausoleum | 26 | 3.84 | 150 | 73 | 12 | 7.7 | 55 | 5.4 | 2.6 | 0.57 | 0.28 | 0.21 | 0.021 | 0.18 |
| 21 | Tower Cellar Level 1 | 20 | 3.34 | 201 | 113 | 18 | 11 | 76 | 7.4 | 3.4 | 0.67 | 0.38 | 0.25 | 0.025 | 0.22 |
| 22 | Tower Cellar Level 2 | 16 | 2.85 | 200 | 110 | 18 | 11 | 74 | 7.3 | 3.4 | 0.57 | 0.31 | 0.21 | 0.021 | 0.18 |
| 23 | Tower Cellar Level 3 | 21 | 3.36 | 168 | 95 | 15 | 11 | 61 | 6.2 | 2.8 | 0.56 | 0.32 | 0.21 | 0.021 | 0.17 |
| 24 | Tower Cellar Level 4 | 16 | 2.79 | 206 | 112 | 18 | 10 | 79 | 7.7 | 3.6 | 0.57 | 0.31 | 0.22 | 0.022 | 0.19 |
| 25 | Tower Cellar Level 5 | 33 | 6.65 | 52 | 41 | 3.6 | 13 | 65 | 8.9 | 4.1 | 0.35 | 0.27 | 0.43 | 0.059 | 0.36 |
| 26 | Monastery Gate | 13 | 4.63 | 96 | 53 | 7.7 | 9.3 | 34 | 4.1 | 2.1 | 0.44 | 0.25 | 0.16 | 0.019 | 0.13 |
| 27 | Outer Cloister | 49 | 7.41 | 80 | 41 | 6.3 | 8.0 | 25 | 2.9 | 1.4 | 0.59 | 0.30 | 0.18 | 0.021 | 0.14 |
| 28 | Barracks | 100 | 6.87 | 157 | 78 | 13 | 7.9 | 64 | 6.9 | 3.3 | 1.08 | 0.53 | 0.44 | 0.048 | 0.39 |
| 29 | Jail Level 1 | 63 | 5.9 | 121 | 70 | 10.0 | 9.9 | 44 | 4.9 | 2.3 | 0.71 | 0.42 | 0.26 | 0.029 | 0.21 |
| 30 | Jail Level 2 | 90 | 6.68 | 136 | 75 | 11 | 9.1 | 59 | 9.6 | 3.0 | 0.91 | 0.50 | 0.39 | 0.064 | 0.34 |
| 31 | Jail Level 3 | 89 | 6.67 | 111 | 65 | 9.1 | 9.7 | 39 | 4.5 | 2.1 | 0.74 | 0.43 | 0.26 | 0.030 | 0.21 |
| 32 | Inner Cloister | 8 | 3.33 | 106 | 59 | 8.8 | 9.4 | 39 | 4.6 | 2.2 | 0.35 | 0.20 | 0.13 | 0.015 | 0.11 |
| 33 | Cathedral | 36 | 6.89 | 151 | 94 | 13 | 13 | 83 | 17 | 3.8 | 1.04 | 0.65 | 0.57 | 0.114 | 0.51 |
| 34 | Catacombs Level 1 | 78 | 6.37 | 141 | 75 | 12 | 8.8 | 49 | 5.5 | 2.6 | 0.90 | 0.48 | 0.31 | 0.035 | 0.27 |
| 35 | Catacombs Level 2 | 96 | 6.83 | 137 | 74 | 12 | 9.2 | 50 | 5.5 | 2.6 | 0.94 | 0.51 | 0.34 | 0.038 | 0.29 |
| 36 | Catacombs Level 3 | 69 | 6.14 | 167 | 88 | 16 | 9.9 | 59 | 6.2 | 2.9 | 1.02 | 0.54 | 0.36 | 0.038 | 0.32 |
| 37 | Catacombs Level 4 | 50 | 7.34 | 57 | 36 | 4.1 | 11 | 60 | 20 | 11 | 0.42 | 0.26 | 0.44 | 0.150 | 0.38 |
| 38 | Tristram | 56 | 7.51 | 182 | 89 | 16 | 7.3 | 81 | 8.5 | 4.1 | 1.37 | 0.67 | 0.61 | 0.064 | 0.55 |
Healing tiers per 1,000 kills (the data carries `potion_hp1..5`, `potion_mp1..5`): Blood Moor 26 minor + 9 light; Cold Plains 44 + 15; Stony Field 56 + 20; Dark Wood 49 + 17; Black Marsh 43 + 15; Cave 2 119 + 43; Catacombs 2 100 + 37; Catacombs 3 89 minor + 58 light + 20 greater (the first greater healing potions; mana: 36 + 37 + 15). Gems are all chipped (6 to 10 per 1,000 kills); flawed gems start in Catacombs 3 (1.4). Runes: none on the floors except the Countess (section 6) and 0.9 and 1.2 per 1,000 kills in Catacombs 3 and 4 (Andariel adds 0.05 per kill). Keys 6 to 8 per 1,000 kills, identify scrolls about 3.

Reading the table for a farming choice: per minute the best supply is Dark Wood (1.11 healing, 0.60 mana, 0.46 magic+) on the overworld and Tristram, Cave 2, Barracks and Cathedral below; per kill Cave 2 and the Tower Cellars 1 to 4 (160 to 205 healing per 1,000 kills) are 2.5 to 3x the overworld because they hold 1 to 3 uniques in 16 to 46 units. Blood Moor is the poorest field in everything (0.25 healing per minute, 0.06 magic+).

## 5. Where the supply comes from (kind split, per clear)

| area | kind | units | healing | share | magic+ | share |
|---|---|---|---|---|---|---|
| Cold Plains | plain | 190.6 | 6.44 | 55 % | 2.01 | 41 % |
| | minion (of a unique) | 5.4 | 0.20 | 2 % | 0.07 | 1 % |
| | champion | 0.25 | 0.32 | 3 % | 0.10 | 2 % |
| | unique | 0.83 | 2.16 | 18 % | 0.81 | 16 % |
| | superunique (Bishibosh) | 1.0 | 2.60 | 22 % | 1.93 | 39 % |
| Cave 2 | plain / minion | 35.8 / 7.8 | 1.21 / 0.27 | 16 / 4 % | 0.38 / 0.10 | 14 / 4 % |
| | champion / unique | 1.0 / 1.8 | 1.30 / 4.75 | 17 / 63 % | 0.39 / 1.78 | 15 / 67 % |
| Catacombs 2 | plain / minion | 76.8 / 14.8 | 2.41 / 0.47 | 18 / 4 % | 0.89 / 0.17 | 19 / 4 % |
| | champion / unique | 1.5 / 3.3 | 1.95 / 8.39 | 15 / 63 % | 0.58 / 3.15 | 12 / 66 % |

Consequence for the bot: a hero that skips the pack with the blue or gold name (a unique, a champion) in these areas gives up most of the supply; a minion drops as a plain kill (I: flag 0x10 only, TC1). A flag-4 unit takes TreasureClass2 before flag 8 (V read); in the population these are the units with flags `0xd`; the dormant meaning of flag bit 0x1 is U.

## 6. Bosses and containers (per kill, V run)

Per kill: total items, potions (healing + mana + rejuvenation), magic, rare, set, unique gear, runes, and magic or better of all item types.

| boss | area | class | ilvl | items | potions | magic | rare | set | unique | runes | magic+ | first kill |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Andariel | Catacombs Level 4 | Andariel | 12 | 4.93 | 0.62 | 1.13 | 0.40 | 0.30 | 0.13 | 0.05 | 2.10 | first kill Andarielq 3.732 drops, magic+ 3.458 |
| Bishibosh | Cold Plains | Act 1 Super A | 5 | 6.00 | 4.00 | 1.42 | 0.37 | 0.05 | 0.02 | 0.00 | 1.93 |  |
| Blood Raven | Burial Grounds | Blood Raven | 10 | 3.42 | 0.18 | 1.19 | 0.31 | 0.04 | 0.02 | 0.00 | 1.70 |  |
| Boneash | Cathedral | Act 1 Super C | 17 | 6.00 | 4.00 | 1.40 | 0.39 | 0.05 | 0.02 | 0.00 | 1.93 |  |
| Bonebreak | Crypt | Act 1 Super A | 5 | 6.00 | 4.00 | 1.42 | 0.37 | 0.05 | 0.02 | 0.00 | 1.93 |  |
| Coldcrow | Cave Level 1 | Act 1 Super A | 7 | 6.00 | 4.00 | 1.44 | 0.36 | 0.05 | 0.02 | 0.00 | 1.94 |  |
| Corpsefire | Den of Evil | Act 1 Super A | 4 | 6.00 | 4.00 | 1.43 | 0.36 | 0.06 | 0.02 | 0.00 | 1.93 |  |
| Griswold | Tristram | Griswold | 8 | 3.00 | 0.00 | 0.92 | 0.06 | 0.04 | 0.02 | 0.00 | 1.07 |  |
| Pitspawn Fouldog | Jail Level 2 | Act 1 Super B | 14 | 6.00 | 4.00 | 1.41 | 0.38 | 0.06 | 0.02 | 0.00 | 1.94 |  |
| Rakanishu | Stony Field | Act 1 Super B | 8 | 6.00 | 4.00 | 1.43 | 0.37 | 0.05 | 0.03 | 0.00 | 1.94 |  |
| The Countess | Tower Cellar Level 5 | Countess | 11 | 5.44 | 0.46 | 1.23 | 0.10 | 0.07 | 0.03 | 1.89 | 1.53 |  |
| The Smith | Barracks | Smith | 13 | 3.00 | 0.00 | 0.92 | 0.06 | 0.04 | 0.02 | 0.00 | 1.08 |  |
| Treehead WoodFist | Dark Wood | Act 1 Super B | 8 | 6.00 | 4.00 | 1.43 | 0.37 | 0.05 | 0.03 | 0.00 | 1.94 |  |
The Super A/B/C classes make the same 6 items at every ilvl 4 to 17 (Corpsefire, Bishibosh, Bonebreaker, Coldcrow, Rakanishu, Treehead, Pitspawn, Boneash): the item level only changes the bases, which stay at levelreq at most 10. The Countess: runes per kill r01 0.11, r02 0.07, r03 0.24, r04 0.16, r05 0.33, r06 0.21, r07 0.46, r08 0.31 (El .. Ral), none higher (V run). Andariel's first kill rolls `Andarielq` (no potions, 3.7 items, 0.07 runes); later kills 4.9 items and 0.05 runes. Her 4.9 items include 0.30 set and 0.13 unique gear per kill: the best set chance of Act 1 per kill (a plain Cold Plains kill gives 0.0009 set and 0.0004 unique gear).

Containers (the per-round class tables; `chests` in the data, `Act 1 Chest A/B/C` at item level 1, 3, 5, 8, 11): per opened round 0.11 healing, 0.06 mana, 0.024 TP scrolls, 0.027 gems, 0.04 to 0.05 magic or better items and 0.28 to 0.39 plain gear (chest C), at every item level tested: the rolled table does not depend on the level inside Act 1. A round is empty 25 % of the time (S treasure-trace.md section 6); how many rounds a chest, barrel or urn gives and how many of them an area holds were not extracted (U). Barrels and urns have no class of their own in Act 1 (V: no `Act 1 Barrel` row in the class list); their table is U.

## 7. d2d against game.exe on this data (V run)

`bash tools/emu/tre_run.sh dm_diff.py POP 300` (`drop-dump` of the pristine fork build, `$TRE_DROPDUMP`, against the native roll): **102 keys, 33 classes, 40,800 rolls (1 player MF 0 and 3 players MF 300), 40,800 equal, 0 differ.** d2d's class selection reads the same MonStats and SuperUniques columns (V read: `components/game/loot.hpp` lines 86 to 101 of the read checkout: superunique row, then `Boss::champion`, `Boss::unique`, else the plain class, quest class, `tc_upgrade` only past Normal).

## Deviation list (D1..) d2d vs game.exe

| id | game.exe | d2d | tag | differential |
|---|---|---|---|---|
| D1 | the population above: unit counts, kinds (flag 4 champion, flag 8 unique, 0x10 minion) and levels per Act 1 level, from `FUN_00573cb0` and the pack code | d2d chooses its own pack kinds through `monster.boss` and its room fill; not compared with these 12 seeds | U (areamon D1 says the same for counts) | can be built: run the fork's `monsters` verb on the 12 seeds of the same areas (docs/forkdbg.md), compare class, kind and level multisets with the `U` lines of `pop.txt`; oracle = this population. Needs the lab binary |
| D2 | the roll `FUN_0055a6d0` with MF and players | equal on 40,800 rolls | V run | exists (`dm_diff.py`) |
| D3 | quest class (TreasureClass4) from the killer's quest flags: Andariel's first kill 3.7 items without potions | the acting player's flags | I (treasure-trace.md D7) | exists in part (`tre_select.py`), pet killer case to build |
| D4 | monster level = row level + 3 unique, +2 champion; berserker champions +3 | d2d's berserker champions +0, boss stats at the boosted level | S (monster-mods-trace.md D8, 1,746 cases) | exists there; moves only the ilvl of three champion kinds, no count |
| D5 | landing floor and the 6-item cap per kill (`FUN_00555da0`) | cap in the roll equal (the 40,800 rolls include it); landing spot differs (treasure-trace.md D5) | V run (cap), S (landing) | cap: exists; landing: exists upstream |
| D6 | containers per area and tier (`FUN_00585f60` family) | `chest_tc(...)` | S (treasure-trace.md section 6: 20,000 container jobs equal) ; the per-area count U | per-area object population: build from `objgroups` like the monsters |

Proposed rows for docs/upstream/pipeline-2026-11-14.md (not edited here): "decomp31 D1 population kinds per Act 1 area, U, no PR; D2 drop roll on the Act 1 key set, equal, none". **No PR candidate**: no d2d deviation was proven (mismatches before 0). Nothing was built in the fork checkout.

## Bot findings (B1..), described, not changed

| id | where | what the code assumes | the rule | tag |
|---|---|---|---|---|
| B1 | `internal/survival/economy.go:30-37` (`Economy`), `internal/data/supply.go:25-33` (`supplyTable`, `SupplyOf`) | the economy is gold income minus potions burned; item supply is not credited; `SupplyOf` is read by no non-test code (V: grep) | a hero in Dark Wood is paid 1.11 healing + 0.60 mana potions per minute (V run), which offsets part of the potion burn the economy charges and is not counted; the table's single plain row (0.057 potions, 0.070 for casters) is the H2H or Cast average, not the area's, and has no unique or boss share | V run |
| B2 | `tools/levelsim/model.go` (XP only), docs/levelsim.md section 2 "survival ... potions, gold, items outside the model" | zones are valued by XP per minute | the same area differs 3x in supply per kill (Cave 2 162 healing per 1,000 kills, Cold Plains 59) and a boss kill is 22 to 39 % of an area's clear; `dropmodel-normal.json` gives the missing columns | V run |
| B3 | `internal/lootrules/` examples and `internal/loot/pickit_starter.json` (rules for magic and better gear, "jewelry to id") | thresholds written without a drop rate | 58 to 88 % of Act 1 gear drops are plain; a rare or better gear piece is one per 50 kills in the Crypt to one per 625 in Blood Moor; every base is clvl 1..10: in Cave 1 a rare gear piece is one per 230 kills (4.3 per 1,000, V run) and a rare weapon about one per 450 (I: half of the gear) | V run + D |

## Drill recipe for the fork verbs (docs/forkdbg.md, docs/forkcheats.md; run later, not here)

The lab binary, one hidden slot, fresh throwaway character with zero setup writes for rates; kills by `debug mon set <id> life 1` then a cast.
- **T1 plain rates**: `debug spawn fallen2 1` in Dark Wood (item level 5), 2,000 kills, `ground` and `dump items`: expect hp 33, mp 20, TP 9, magic or better 11.7 per 1,000 kills (+/- 2 sigma of about 3 for hp). U until run.
- **T2 superunique**: spawn the Super A class monster (`debug spawn` a superunique if the verb supports it, else run the map location) and kill 30: expect exactly 6 items, 4 potions each time.
- **T3 Countess**: Tower Cellar 5 (area id 25) 20 kills in separate games: runes mean 1.89, never above r08 (Ral), items 5.44 mean.
- **T4 population (D1)**: `monsters` in Cold Plains and Cave 2, 12 seeds: units 198 and 46, flags: champion 0.25 and 1.0 per clear, unique 0.83 and 1.8 per clear.
- **T5 Andariel**: first kill 3.7 items and no potions; second 4.9.

## Matrix and queue cell text (proposed)

- docs/research/treasure-trace.md section 5: the Countess row's "2.28 runes" should read "2.28 uncapped exact, 1.89 with the 6-item cap (decomp31, V run)"; add "Chest and barrel counts per area" to its dead ends as still open.
- docs/levelplan3.md / levelsim.md: area value columns per minute from `dropmodel-normal.json` (`per_min`), section 9.
- `docs/coverage/rule-coverage.md`: no new rule row; `dm_diff.py` adds 40,800 equal rolls to the `treasure` row's evidence.

## 8. Unverified and dead ends

- U: the meaning of flag bit 0x1, whether every `0xd` unit is a champion pack member in d2d's pack builder (D1).
- U: container counts and the area-to-tier map; barrel and urn tables.
- U: strength and dexterity of the bases ("usable25" counts level only); the unique and set item picked per quality (the base can be a high-level unique base).
- U: kill rates (the I time model of areamon); nothing here ran live.
- Dead end: the exact-odds tool `tre_act.py` over-counts a class that can make more than 6 items (Countess runes 2.28 against 1.89, Andariel gold and gear about 1.5 percent); the data comes from the native rolls, not from the exact odds.
- Dead end: the NAS builder container for the emulator run; a plain listing there did not answer within 2 minutes and the emulator image is documented on the worker host only (docs/emu-image.md). The run used the worker host, 1 core at nice 10: about 3 minutes for the 12-seed population, 9 for the rolls.
- No magic find: MF 100 and 300 shift the quality only; the table of treasure-trace.md section 3 is the reference (champion class, MF 0 / 100 / 300).

## 9. Follow-up cards (not built here)

1. levelsim `-drops`: add `potion_hp`, `potion_mp` and `magic_plus` per-minute columns of `dropmodel-normal.json` to the zone value, with the potion price from `internal/survival` as a credit against the potion burn (B1, B2).
2. A "what do I find here" line in the Where-to menu: the area's per-1,000-kill and per-minute potions, magic or better items and the Countess or boss line, read from the JSON (keys by area name).
3. Loot-rules examples: write the Starter rule thresholds with the drop rates of section 4 (B3) and a doc line "rare gear on a clvl 15 to 25 hero's base is one per 230 kills in Cave 1".
4. Container counts per area (objgroups population plus the chest tier map) and the barrel table (U), then add `containers` to the JSON.
5. D1 differential (fork `monsters` verb, 12 seeds).
6. Merge `data.SupplyOf` with `dropmodel` (one source) and read the area rows instead of the plain/caster rows.

## Data file schema (`internal/plandata/dropmodel-normal.json`, `dropmodel/1`, additive)

Top keys: `schema`, `version` (1), `difficulty` (normal), `conditions`, `units`, `keys` (the meaning of each value), `check` (the exact-odds comparison), `areas`, `monsters`, `bosses`, `chests`. `areas` is keyed by the Levels.txt name (`Cold Plains`, `Cave Level 2`) with `lid`, `seeds`, `units_per_clear`, `kinds_per_clear` (plain, minion, champion, unique, boss), `clear_min` and `kills_per_min` (I), `per_clear`, `per_1000_kills`, `per_min`, `per_clear_by_kind` (the core keys). `monsters` is keyed `<MonStats name>:<kind>` with `tc`, `observed`, `ilvl_mean`, `per_1000_kills`. `bosses` is keyed by the SuperUniques name (or `Andariel`, `Blood Raven`) with `area`, `class`, `tc`, `ilvl`, `drops`, `per_kill` and an optional `first_kill` (quest class). `chests` is keyed by class, then item level, with per-round counts. All values are expected counts of dropped items; gold is not in the file; a key that is absent is 0. Consumers must read by name and ignore unknown keys; a new key is added without a version bump, a changed meaning bumps `version`.

## Reproduce

```
bash tools/emu/am_batch.sh POP 1 12 3                                   # worker host; POP = the population of areamon-trace.md
bash tools/emu/tre_run.sh dm_model.py POP dm.json 20000                  # about 9 minutes at nice 10
python3 tools/emu/dm_pack.py dm.json internal/levelplan/areaxp.csv internal/plandata/dropmodel-normal.json
bash tools/emu/tre_run.sh dm_diff.py POP 300                             # d2d against game.exe on the key set
```

Limits: one player, MF 0, GF 0, Normal; 12 population seeds (means; the per-area min and max are in areamon-table.csv); 20,000 rolls per key (a per-1,000 value of 1 has a standard error near 0.1); the time model is I; no live check.
