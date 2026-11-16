# Act 1 Normal bosses and quest monsters: what the guided journey must survive (BOSS1, decomp19)

> status: current (2026-11-14, card BOSS1, agent decomp19 on Sonnet). Offline: no game, no d2d slot, no Go change, nothing pushed or posted. Question: for the Act 1 Normal quest bosses (Blood Raven, the Countess, the Smith, Andariel) and the Act 1 uniques of the journey (Corpsefire, Bishibosh, Coldcrow, Rakanishu, Treehead WoodFist, Griswold): life, damage, speed, skills, immunities, minions, experience, the minimum safe clvl for the Paladin the bot plays (plain Attack under Might, Zeal from clvl 12), the best tactic with the bot's primitives, where game.exe and d2d differ, and where the bot's boss handling differs.
>
> Tags: **V run** = game.exe's own room population run in the emulator (`tools/emu/bs_spawn.sh` -> `umod_spawn.py`, seeds 1 to 4, Normal, the worker host); **V read** = read in a disassembly or source by an earlier trace; **D** = the patched 1.14d tables (`tools/emu/mpq-table.sh`; derived numbers only); **S** = an earlier trace of ours (named); **I** = inferred (arithmetic in `tools/emu/bs_model.py`; it does NOT run game.exe); **U** = unverified. The dps is `tools/emu/pm_dps.py` (docs/research/pal-skill-matrix.md section 4). Not in the brief's list: Coldworm the Burrower and the Dark Elder are Act 2 superuniques (Superuniques.txt rows 15 and 17, `maggotqueen1` / `zombie5`); Pitspawn Fouldog did not appear in any run area (U where he stands).

## Answer

1. **Minimum safe clvl (I, Paladin plan v2, no block, no resist, armour grade rising with the level, at most half of the life lost standing toe to toe, 4+ healing potions):** Corpsefire 4, Bishibosh 5, Coldcrow 6, Rakanishu 6 (8 minions), Treehead WoodFist 8, Blood Raven 8 (9 comfortable), Griswold 10, the Countess herself 10 (the cellars decide: 11 to 13), the Smith 16 (14 with a shield's block and potions, U), Andariel 18 with a merc or 20 alone, 8 potions and 3 antidotes. The campaign floors: Blood Raven 8, Countess 11, Smith 10, Andariel 16 (14 with a living merc) (D data/campaign/normal.json, internal/plandata/boss_levels.go).
2. **Blood Raven (V run): 113 life, defence 38, level 10, experience 181, all elements and magic 50 percent, physical 0, no minions of her own**. Her arrows: 4 to 6 physical plus a random weapon element (fire 1 to 5 or lightning 1 to 8). Standing melee costs a clvl 8 Paladin about a third of its life; she kites (walk home at 3.3 cells a second, Quick Strike 40 percent of her casts, raises up to 8 zombies).
3. **The Countess (V run): only about 100 life** (`corruptrogue3`, level 8 raised to 11, life x4 of 26, defence 40, fire resist 75 from Fire Enchanted, experience 395); her damage is low (4 to 9 a hit). She is not the reason for the floor: the fire walls (CountessFirewall, skill level 10, about 8 life a second per wall on fire, I) and the five cellar levels are.
4. **The Smith is a SUPERUNIQUE, not an act boss (V run, class 402, flags 0x0b, mods 22.5.22): level 10 -> 13, life x4 = 404, experience 905, Extra Strong (+135 percent damage, +90 percent to-hit), poison resist -50, 21 to 25 a hit.** This corrects docs/research/uniqmods-trace.md Q6 and its act-boss table (x1). He is the most dangerous Normal fight below clvl 16 for a plain melee Paladin: at clvl 12 a toe-to-toe fight costs about twice the life.
5. **Andariel (V run): level 12, life 1,024, defence 60, experience 1,282; fire -50 (takes half again), cold 50, lightning 50, POISON 80 (NOT immune), physical and magic 0.** Melee 6 to 19 (+ poison), spray 7 to 10 + 50 to 100 poison over 16 s, poison bolt 5 to 7 + 100 to 200 over 32 s (S covmissiles2 Q2 / Q4, V run). The Paladin needs 60 s (clvl 12, Zeal 1) to 32 s (clvl 20, game.exe's Zeal) to kill her; d2d's Zeal is slower (Zeal gap, section D7).
6. **Bot findings (section B): B1 the Andariel script walks a melee Paladin away from her every 1.5 s (class-blind kite), B2 Smith script floor 8 against 16, B3 Andariel floor 16 is 2 to 4 levels under the model, B4 the Countess gate 14 is for the cellars, not her.** No PR candidate (section P).

## Q1..Q8

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | What stats does game.exe give each boss at spawn in Normal? | See the table below: level, life, experience, defence, resists and mods from game.exe's own population (FUN_00573cb0 rows, FUN_005a2120 / FUN_005a49b0 mods): act bosses (Blood Raven, Andariel: flags 0x08 only) x1 life, level +0, experience x1; superuniques (Corpsefire, Bishibosh, Rakanishu, Treehead, Griswold, the Countess, the Smith, Coldcrow) level +3, life x4, experience x5, plus their fixed mods and the quest mod 22 | V run (bs_spawn.sh, 4 seeds a level; Blood Raven and Andariel in 4 of 4, each superunique in at least 1 of 4) |
| Q2 | How do life, damage, to-hit and experience scale with difficulty? | Row percent x the MonLvl `L-` column at the level (life, defence, experience at the MonStats level; damage and to-hit at the raised level), Normal / Nightmare / Hell = the Level, Level(N), Level(H) column of the row | V run (S monster-stats-trace Q2, 242,220 rows; S pr-mods-proof: damage and to-hit read at the raised level); NM / H values D + I |
| Q3 | Blood Raven's attack set? | Row: AI `BloodRaven`, A1 = missile `raven1` (ToHit set: the AR roll runs), Skill1 `Nest` (seq_bloodravencast, raises zombies, at most 8 + 2 x difficulty at once, offset 5..19 from the target), Skill2 `Quick Strike` (seq_brquickstrike, rolled when `roll < (difficulty + 4) x 10`: 40 percent in Normal), walk 6 / run 8, speed percent 75 | D, V read (S monsterai2 Q8); the minion class `zombie2` D |
| Q4 | Countess's attack set? | AI `CorruptRogue` (aip 80 / 9 / 85), A1 melee 51..117 percent, Skill1 `CountessFirewall` level 10 in mode A1 (missile 653 maker -> `countessfirewall` 331: Range 1000 + 1 a level frames, fire 3 to 5 + 4 a level, HitShift 1, DamageRate 41); five minions of her own class with the Fire Enchanted rider (V run: pack members 18 to 30 life, level 11, experience 395) | D, V run for the pack; the AI think and the wall's damage per frame are U (not traced; I: a firewall hits every frame, S sk-ground2-trace Q20, so about 8 a second at level 10 if HitShift 1 is x2 / 256 per hit) |
| Q5 | The Smith's attack set? | AI `Smith`, A1 melee 101..129 percent (9 to 11 at level 13, x2.35 from Extra Strong), no skills, no missiles, walk = run 7; El1 fire on A1 has no percent (no fire damage in Normal, D) | D, V run (mods, damage factor) |
| Q6 | Andariel's attack set and AI? | AI `Andariel`: in melee `roll < 30` spray else A1; out of melee 10 percent stand, then spray or bolt (Skill1 / Skill2), else walk (0x5f5830); a hit adds poison El1 (pois, 100 percent, 33 / 33 over 30: unit not resolved, U); minion none; the quest mod 22 only | V run (S monsterai2 Q1: her think equals game.exe's on 3,830 of 3,830 cases); the A1 poison total U |
| Q7 | Immunities and resists a Paladin meets? | table below; only Bishibosh is immune to something (fire 140: Fire Enchanted + Magic Resistant). Holy Fire is wasted on Bishibosh and the Countess (75), halved on Blood Raven (50). The Paladin's physical melee faces 0 physical resist on every boss | V run |
| Q8 | What do they drop? | TreasureClass by row (D): Blood Raven `Blood Raven`, the Countess `Countess` (runes: 5.75 items, 2.28 runes, S treasure-trace), Griswold / the Smith 3 drops, Andariel `Andariel`, Act 1 superuniques `Act 1 Super A/B` (6 drops, 4 potions, 1.87 gear) | D, S uniqmods-trace |

### Boss table (Normal from V run; NM / H life from the tables with the NM / H unique factor x3 / x2 (D + I); experience x5 on superuniques)

| Boss | Area (Levels id) | Level N (NM / H) | Life N (NM / H) | A1 hit N | Key skills | Immunities / resists (fire / cold / light / poison) | Minions | Exp N | Min safe clvl | Tactic |
|---|---|---|---|---|---|---|---|---|---|---|
| Corpsefire | Den of Evil (8) | 4 (39 / 70) | 38 (2,007 / 10,485) | 2-6 | melee; spectral hit mod 27 | 20 / 20 / 20 / 50 | 5 zombies | 165 | 4 | melee; stand in the cave mouth |
| Bishibosh | Cold Plains (3) | 5 (39 / 71) | 28 (1,140 / 6,128) | 1-4 | Resurrect, ShamanFire; Fire Enchanted + Magic Res | 140 (immune) / 40 / 40 / 0 | 2 fallen | 160 | 5 | melee; no Holy Fire; fire on death |
| Blood Raven | Burial Grounds (17) | 10 (43 / 88) | 113 (9,832 / 57,829) | 4-6 + element | raven1 arrows, Nest zombies, Quick Strike | 50 / 50 / 50 / 50 (magic 50, physical 0) | raised zombies (0 own) | 181 | 8 | clear the ring at the mouth, fight at the entry, TP at 40 percent life |
| Coldcrow | Cave 1 Cold Plains (9) | 7 (40 / 71) | 64 (1,257 / 6,339) | 3-4 + cold 3-5 | arrows; Cold Enchanted | 0 / 75 / 0 / 0 | 4 archers | 270 | 6 | close in; chill slows |
| Rakanishu | Stony Field (4) | 8 (41 / 71) | 26 (711 / 3,380) | 3-6 + ltng 3-6 | Lightning Enchanted + Extra Fast (speed 175 percent) | 0 / 0 / 75 / 0 | 8 fallen | 210 | 6 | back to a wall; lightning on death |
| Treehead WoodFist | Dark Wood (5) | 8 (41 / 72) | 122 (3,156 / 15,413) | 9-14 (strong x2.35) | Extra Strong, Extra Fast (speed 175 percent), rage | 0 / 0 / 0 / 0 | 2 brutes | 420 | 8 | fight at a chokepoint; he closes fast |
| Griswold | Tristram (38) | 8 (42 / 87) | 412 (11,436 / 73,774) | 3-9 | Cursed mod 7; walk 3, run 4 (slow) | 0 / 0 / 0 / 50 | none | 955 | 10 | optional; kite is free (he is slow) |
| The Countess | Tower Cellar 5 (25) | 11 (42 / 72) | 96 (1,504 / 6,947) | 4-9 + fire 5-8 | CountessFirewall level 10; Fire Enchanted | 75 / 0 / 0 / 0 | 5 own class | 395 | 10 (cellars 11-13) | step off fire walls, melee her from the side; she dies in 5-10 s |
| The Smith | Barracks (28) | 13 (45 / 82) | 404 (9,294 / 46,682) | 21-25 (x2.35) | Extra Strong; melee only | 0 / 0 / 0 / -50 | none (Barracks crowd) | 905 | 16 (14 with block) | fight at the arrival point, retreat and drink at 50 percent |
| Andariel | Catacombs 4 (37) | 12 (49 / 75) | 1,024 (24,800 / 60,031) | 6-19 + poison | spray 7-10 + 50-100 poison, bolt 5-7 + 100-200 poison (L1 rows) | -50 / 50 / 50 / 80 | none | 1,282 | 18 with a merc, 20 alone | antidote before the first hit; stay in melee reach; potions at 50 percent |

Level, experience, life, resists in the Normal column are V run (the mods the run printed: Bishibosh `8.9.22`, Rakanishu `17.6.22`, Treehead `13.5.6.22`, Corpsefire `27.22`, Coldcrow `18.22`, Countess `9.22`, Smith `22.5.22`, Griswold `7.22`, Blood Raven `12.22`, Andariel `22`). Life shown is the average of the row's range; the V run values are one roll each: Blood Raven 113, Andariel 1,024, Countess 104, Smith 404, Griswold 412, Corpsefire 48, Bishibosh 20, Coldcrow 68, Rakanishu 20, Treehead 152. NM / H levels are D with the +3 added, the NM / H life uses MonStats `MinHP(N)` / `(H)` at the Level(N) / Level(H) column (D + I; uniqmods-trace section Q10 saw Bishibosh 870 to 1,299 at NM, consistent in order of magnitude). The damage and AR in the table use the raised level (S pr-mods-proof).

## Exposure model (I; `python3 tools/emu/bs_model.py <tables dir> fight`)

Hero: Paladin plan v2 stats (clvl 8: Str 38 Dex 21 Vit 46; 12: 46 / 21 / 58; 16: 51 / 21 / 73; 20: 55 / 23 / 87), life 55 + 2 per clvl + 3 per Vit above 25 (S wiki; CharStats LifePerLevel 8 and LifePerVitality 12 in fourths, D), Might 4 / 8 / 8 / 8, Zeal 1 / 4 / 4 from clvl 12, Short Sword, armour defence 25 / 50 / 80 / 110 (I), no block, no resist. Monster hit chance: AR / (AR + defence) x 2 mlvl / (mlvl + clvl), 5 to 95 percent (S sk-melee-trace); period 32 ticks (A1 about 16 frames + aidel 15, I, plus or minus a third).

| clvl (life) | Blood Raven | Countess | Smith | Andariel melee only | Andariel dps of the hero (game.exe / d2d), ttk |
|---|---|---|---|---|---|
| 8 (132) | 33 percent of life lost standing | 36 | 699 | 961 | 7.5 dps, 137 s |
| 12 (176) | 8 | 8 | 208 | 251 | 17.1 / 12.5 dps, 60 / 82 s |
| 16 (229) | 3 | 3 | 78 | 90 | 27.6 / 15.9 dps, 37 / 64 s |
| 20 (279) | 2 | 2 | 45 | 50 | 31.7 / 18.2 dps, 32 / 56 s |

(ttk = life / the dps against her defence 60, Zeal from clvl 12; d2d's Zeal strikes 15 ticks apart against game.exe's 7, S pal-zeal-trace.) Not in these numbers: Andariel's spray and bolt (a poison of 4.7 life a second per active hit, whether hits stack is U: if each hit's remaining total is added to the next, the rate runs to tens of life a second, so poison, not melee, decides the fight); the Countess' walls; Blood Raven's arrows from range and her zombies; block (a Paladin's shield stops a good share of the Smith's blows, U for the rate). Antidote and Cleansing cut the poison length (S covpal-trace Cleansing; the Antidote's size U).

## Deviations d2d vs game.exe (all S from earlier traces unless marked)

| D | Rule in game.exe | Rule in d2d | Tag | Differential |
|---|---|---|---|---|
| D1 | Blood Raven: walks home when the TARGET is 50 or more from home, pace 100 on the walks (2.3 x a plain walk), Quick Strike roll | no pace, only her own home distance | V run (monsterai2 D4, D5: 820 / 162 of 1,500 differ) | yes: `mai2_run.sh BloodRaven` |
| D2 | Andariel's spray / bolt poison: per frame, 50 to 100 points in total at level 1 | stored the whole total as a per-frame value (x31); the subject (`local/missilepoison-skillrows`) is in upstream main per pr-queue.md section 2 | V run (covmissiles2 D2: 1,640 of 1,640) | yes; whether the lab binary (the d2d the bot plays against) has it is U: drill 1 |
| D3 | Andariel missile physical gains level brackets (42 to 45 at level 8) | raw row (7 to 10 always) | V run (covmissiles2 D1: from level 4); Normal level 1 agrees (I) | yes (`cm2_run.sh`); matters from Nightmare |
| D4 | Superunique and unique life, defence and experience read at the MonStats level (Countess level 8 rows), only damage and to-hit at +3 | all four at +3 (life 1.3 x, defence 2.5 x in early Normal; for the bosses here, I: Countess 36 / 27 = 1.33 x, Griswold 1.6 x, the Smith 1.27 x) | V run (pr-mods-proof D1 narrowed) | yes (`boss_level.py`); upstream queue status of `fix/boss-stats` U (not in pr-queue.md) |
| D5 | The Smith is a superunique (SuperUniques row 20): +3, life x4, experience x5, Extra Strong | not read: whether d2d spawns him with the superunique mods or as an act boss is U | V run (game side) | would need a d2d spawn read: drill 4 |
| D6 | `countessfirewall` lives Range + LevRange x level frames | `max(range, 1)`, no LevRange | V run (covmissiles2 D3: 845 of 845), effect one frame a level | yes |
| D7 | Zeal: next strike 7 ticks after the last | 15 ticks | V run (pal-zeal-trace, 24 of 24); fixed upstream by 276f85c0 per pr-queue.md 1b, lab binary U | yes (`zd_zeal_test`) |

## Bot findings (described, not changed)

| B | Where | Finding | Tag |
|---|---|---|---|
| B1 | `internal/boss/act1.go:157` (`Kite: Range 10, Every 1.5 s, Away 12`), `internal/boss/run.go:590-596` | the kite step has no class gate: a melee Paladin within 10 subtiles of Andariel walks 12 away every 1.5 s and never stays in reach; the script was written for the caster dance (Kolbot, S). Expect a fight of repeated approaches while spray and bolt land | V read; not run |
| B2 | `internal/boss/act1.go:137` (Smith `MinLevel[0]` 8), data/campaign/normal.json (`tools_of_the_trade` floor 10, recommended 13) | the model's toe-to-toe loss is 208 percent of life at clvl 12 and 78 percent at 16; floor 10 and the script's 8 send a clvl 10 Paladin into a 404-life Extra Strong fight (17 life a second against a 12 dps hero) | I |
| B3 | `internal/plandata/boss_levels.go:23` (Paladin Andariel 17 / 15, floors 16 / 14) | a clvl 16 Paladin loses 90 percent of its life in melee alone before poison (section exposure model); the floor is a warning, not a wall (`tuning_boss.go` text), so the bot goes in | I |
| B4 | `internal/boss/act1.go:117` (Countess `MinLevel` 14, `MinPotions` 2) | the boss has about 100 life and 4 to 9 damage: the 14 is the cellars (and the older route floor 20 without a merc in docs/research/andariel-chain-needs.md); the knob mixes two risks | I |
| B5 | `internal/env/tuning_boss.go` Blood Raven text "her 0 percent resist" (melee) | correct: physical resist 0, every element and magic 50 (V run); B2 of monsterai2 already asks to tag the 50 percent D | V run |
| B6 | docs/research/uniqmods-trace.md Q6 table and the act-boss row (x1 for the Smith) | wrong for the Smith (V run above): proposed text in the matrix section | V run |
| B7 | `internal/boss/act1.go` Andariel `PreAntidoteBelow: 75`, `MinPotions: 3` against the gate kit (`internal/task/gate.go`) | the script's own kit (3) is below the gate kit; poison stacking and the potion rate decide the fight, no number behind 3 | I |

`bossPhase` / `OnPhase` hooks do not exist in `internal/boss`, `internal/bot`, `internal/combat` (grep; the wiring audit removed them): phases are the script's `Hazards`, `RetreatPct`, `EscapePct`, `DrinkPct`.

## PR candidates

No PR candidate. D2 is already merged upstream per pr-queue.md section 2 (nothing left to prove, the lab binary state is U), D1 / D4 are proven by earlier harnesses but have no built "after" run (no patched build under this card: no fork access time box), D5 / D6 / D7 are covered. Draft text stays in docs/research (`pr-mods-proof.md`, `monsterai2-trace.md`); nothing added to docs/upstream/pr-queue.md by this card. If the user wants D1 (Blood Raven pace) taken up: draft: needs fork patch + verifier.

## Drill recipe (fork verbs; docs/forkdbg.md, docs/d2d-fork.md, docs/forkcheats.md; one hidden slot, throwaway character; nothing run by this card)

1. D2 / poison size: `debug spawn andariel 1`, `debug goto 37`, stand in reach, `step` through one spray, `dump stats` and `debug states`: expect 50 to 100 poison points over 16 s (game.exe) rather than 1,500 to 3,100 (the old d2d). Also read how a second hit changes the rate (stacking, U): two sprays 2 s apart.
2. Superunique stats (D4, D5): `debug spawn smith 1` then `monsters`: game.exe life 404, level 13, damage 21 to 25; and and compare the level, life and damage with the Q1 row.
3. Blood Raven (D1): the monsterai2 drill 2 (walk the player 60 subtiles from her spot, `step 1`, read her position): expect her home walk as soon as the player is 50 away at 3.3 cells a second.
4. Poison and Countess wall damage per frame (U in Q4): `debug spawn corruptrogue3 1 unique`, let her cast, `dump stats` life every 5 ticks while standing on one wall.
5. A1 anim frames and the melee period (I, plus or minus a third): `step 1` and read `monsters` modes for Smith and Andariel over 20 swings.

## Matrix cell updates (proposed text)

- docs/research/uniqmods-trace.md Q6 / act-boss table: replace "Smith: act boss, x1" by "The Smith is a superunique (SuperUniques row 20): level 10 -> 13, life x4, experience 181 -> 905, mods 22.5.22 (Extra Strong), V run 2026-11-14 (boss1-trace.md)"; Blood Raven 113 life, Andariel 1,024 life and her resists (fire -50, poison 80) go into the Blood Raven and Andariel rows (the open "HP conflict" of docs/sme/blood-raven-caster.md section 1 resolves to 113: V run).
- docs/research/pal-skill-matrix.md: add a boss row under 3.1 Sacrifice: "against Andariel (defence 60) Sacrifice 1 does 21.0 dps at clvl 12 against Zeal 1's 17.1 (game.exe) and 12.5 (d2d): the bot never casts it (B1) and loses about a quarter of the fight time (I)".
- docs/research/pal-skill-matrix.md row Holy Fire: "wasted on Bishibosh (fire immune), the Countess (75), halved on Blood Raven; use Might against bosses (V run resists)".

## Dead ends and limits

- The Countess' AI think and the wall's per-frame damage were not traced (U); no differential exists for `CorruptRogue` (not in `mai2_six`).
- Whether poison hits stack in 1.14d is not read here (U); the dominant risk of the Andariel fight rests on it.
- Pitspawn Fouldog, Bonebreak, Boneash: Bonebreak appeared in Crypt 1 (class 0, mods 5.8.22, level 5, life 40) and Boneash in the Cathedral (class 276, level 17, 8.5.18.22); Pitspawn did not appear in the cave areas run (10, 11).
- The exposure model assumes the boss swings every 32 ticks; the anim frames were not read (drill 5).
- Ran: `D2_LABRUN_HOST=<worker> tools/emu/run-on-worker.sh --ttl 900 bs_spawn.sh` (areas 17 25 28 37 38 8, seeds 1 to 4, Normal, 72 s) and `... bs_spawn.sh 0 1-4 3 4 5 9 10 11 18 33` (110 s); tables with `tools/emu/mpq-table.sh monstats|monlvl|skills|missiles|superuniques|monumod|levels` on the worker host; `python3 tools/emu/bs_model.py <dir> table|fight` locally.
