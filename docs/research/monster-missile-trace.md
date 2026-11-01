# Monster missiles for Acts 1 and 2: game.exe 1.14d vs d2d vs our bot (card monstermissile, 2026-11-01)

> status: current. Offline only: no game, no d2d slot, nothing pushed or posted. d2d read at upstream main `23e2baec` (PR checkout dir on the worker host, the SKMISS candidate head: two commits above the base 72132b99, `components/game/ai.cpp` untouched by it). game.exe md5 2b3600c1... under upstream's unicorn harness (`tools/emu`, `drlg.boot()` loads every data table with game.exe's own loaders). Scope: the rule-coverage gap "missiles 658 of 684 untested" (docs/coverage/gaps.tsv rank 1), monster-side missiles that hit the player, Acts 1 and 2 (levels.txt `Act` 0 and 1 spawn lists: 71 launch rows) plus the brief's extra shooters (Fetish blowdart, Corpse Spitter, Maggot Queen) and the Act 1 / 2 bosses with a missile (Andariel, Blood Raven, Radament, the Summoner). The player-skill missile path is docs/research/sk-missile-trace.md; the hit on the player (to-hit, block, dodge, roll) is docs/research/monster-attack-trace.md and is only extended here.

Tags: **V read** read with `tools/emu/disasm.py` (address, instructions); **V run** game.exe's own code run under the emulator by this agent (the d2d side of a V run is a Python port of d2d's expression read at file:line, NOT the real d2d binary: said wherever it matters); **D** a 1.14d table (`mpq-cat` on the patch stack, the lowercase `patch_d2.mpq` trap of docs/process/mpq-data-reading.md avoided); **S** an existing proof or upstream doc; **I** inference; **U** unverified.

## Verdict

1. **Normal Act 1 / early Act 2 arrows, spikes and Fallen Shaman fire: d2d equals game.exe in damage.** At level 1 (Normal: the missile level is 0 + 1) none of the level brackets bite, so spike, cr_arrow, skbowarrow, blowgun, pantherjav and shafire damage agree with d2d's `src x SrcDamage / 128 + row` (V run: 213 launches, owner damage 0 differ from MonLvl x pct / 100, see Q3).
2. **Nightmare and Hell: d2d's MissA path drops the row's physical level brackets (D1).** game.exe adds `(level - 1) x MinLevDam / MaxLevDam` (levels 1 / 4 / 8) to the physical part of the missile; d2d never reads `MinLevDam` / `MaxLevDam` of a missile row. Quill Rat spike: game.exe 19-26 (NM) and 44-61 (Hell) against d2d's 10-17 and 23-40; Corpse Spitter 70-100 / 130-160 against 25-55. 28 of 213 launches differ on it, none in Normal.
3. **Poison and fire carried by the ROW are missing on d2d's MissA path (D3)**, read not run (d2d side): Sand Maggot / Maggot Queen spit (poison 6 / 12 / 19 points over 8 / 12 / 16 s in Normal), Unraveler and Radament mummy spit (7 to 28 points), the Act 2 poison pot, the fire pot. For the Sand Maggot the poison IS the attack (its physical part is under a tenth of a point).
4. **D2: d2d treats a row's MinDamage as whole points whatever HitShift** (goospit shift 1: game 2-6 256ths, d2d 1-3 points: six rows, Normal).
5. **The block gate (a) of monster-attack-trace is wider than its early-game examples**: 25 of the 49 distinct Act 1 / 2 rows (21 of them projectiles) have no physical part (V run: record max 0): Fallen Shaman fire, Skeleton Mage bolts, Bighead lightning, Unraveler / Radament bolts and spit, the pots. game.exe cannot block them; d2d can (a shield build).
6. **Bot:** no per-missile threat model. The dodge reacts to any hostile missile inside 4 subtiles; an arrow covers that in 0.14 to 0.24 s (Q6), the sidestep list names four kinds none of which an Act 1 / 2 shooter fires (B1-B4).
7. **No PR candidate**: the proof of D1 / D2 / D3 is a game.exe run against a port of d2d's expression, not the existing differential against the real d2d build (rules at the end).

## Questions answered

| Q | Answer | Tag |
|---|---|---|
| Q1 | **How a monster fires a MissA row.** `005a7670` (attack handler, reached from the MonAI think) -> `005a6d50(unit, which)`: the row id from `0063e6b0` (MissA1 / MissA2 / MissS1 / MissC), the missile LEVEL = `DifficultyLevels row (0x611d30(game difficulty byte +0x6d)) +0x10` + 1 (0 / 3 / 7 + 1 = 1 / 4 / 8; d2d's `monster_skill_level(1, difficulty)`, same numbers), replaced by the unit's stat 12 when the monster-data flag bit 9 (unit +0xc4 >> 9 & 1) is set (which monsters: U); then `0056ecb0(game, unit, row, .., level)` -> missile create `0059fa30`. Shaman fire / Unraveler bolts (`srvdofunc 85`, `005cb0c0`): the skill's `srvmissilea` stem + (its number + MonStats `TransLvl`): fallenshaman1..4 fire shafire1..4, unraveler1..3 unholybolt1..3 (row names from the table, the V run used the same rule) | V read (`005a6d50`: `mov eax,[eax+0x10]; add eax,1`; `test edx,1; push 0xc; call 00625480`); D |
| Q2 | **What damage the missile carries (the launch record).** `0059f900` -> `0064b860` builds a 0x7c-byte record, `0064ac60` copies it into the missile's stats (0x15 / 0x16 physical min / max in 256ths, 0x30..0x3b the elements, 0x19 damage %, flags). Row without a `Skill` (every monster MissA row): physical min = `0064af20` = `(MinDamage + level bracket(MinLevDam1..5, level) + synergy calc) << HitShift`; max `0064aff0` alike; the level bracket is `00644b70`: `(level - 1) x b1` for level <= 8, then 8 x b1 + (level - 8) x b2 ... (V read, 5 steps at 8 / 16 / 22 / 28). For a monster OWNER with `SrcDamage` (+0x12d, a byte, the 128ths) S > 0: the owner's stat 0x15 / 0x16 (its damage in the mode set by `005a4f50`: A1 -> A1, A2 -> A2, SC / S1 -> S1) `<< 8`, `x S / 128` toward zero (`0064bc43..0064bc63`: `imul; cdq; and edx,0x7f; add; sar 7`), ADDED to the row part (`0064bdf7`: `add esi,[ebp-0x2c]`). Elements: `0064b100 / 0064b1d0 / 0064b2a0` = (EMin / Emax + MinELev / MaxELev brackets at the level) `<< HitShift`, ELen + ELevLen brackets, stored by the row's EType (`0064be18` switch); the owner's own El1..3 elements (pct = chance) ride along via the mode stats. A monster with an INVENTORY (`monstats.inventory` = 1: of the Act 1 / 2 shooters only Blood Raven) takes the damage from its equipped weapon instead (`0064bac5` -> `00622830`, `00645830`: random per launch, V run 1-5 / 1-19 / 1-6 points on the same row) | V read + V run (213 launches, `tools/emu/mm_launch.py`: the missile's own stats read back with `00625480` after `0059fa30`); weapon formula U |
| Q3 | **Difficulty scaling.** (a) The owner's damage stat the record starts from is MonLvl x percent / 100 by difficulty (`006538a0` at spawn, `005a4f50` per mode): read back from the spawned game.exe monster and compared with `L-DM / L-DM(N) / L-DM(H)` x the A1 / A2 / S1 percent / 100 of monstats.txt: **0 of 213 differ** (the S rule of monster-attack-trace, now on the launch rows). (b) The missile level 1 / 4 / 8 enters the row's brackets, physical AND elemental: Fallen Shaman fire 1-4 (Normal) -> 31-34 (NM) -> 71-74 (Hell) points for shafire1, shafire2..4 add 1 / 2 / 3 on the top; Skeleton Mage fire 4-7 -> 21-52 -> 53-129 (own El rows x MonLvl, not the row). (c) Rows with no bracket and no owner part (poison per-frame) still grow through MinELev: goospit1 8 -> 26 -> 50 per frame (V run, `tools/emu/mm_elev.py` calling `0064b100` at level 1 / 4 / 8). (d) What does not scale: arrows and javelins in the Normal early game (level 1), the pots' reach, ToHit (a table flag, not a number) | V run (A, B, C); D |
| Q4 | **Per-row data and the V run record** (table below): velocity, range, physical / element ranges per difficulty, poison totals, ToHit, blockable, reach | D + V run |
| Q5 | **The hit on the player.** `005adf10`: to-hit roll `0057d9b0` only when the row's ToHit (+0x18b) is set (17 of the 49 rows: spike, cr_arrow, skbowarrow, pantherjav 1/2/4, blowgun, raven1; every shaman bolt, mage bolt, spit, mummy, pot and Bighead bolt always hits), then `005ad730`: Avoid (a missile) / Evade / Dodge `0057dd60`, then block `0057dfb0(.., missile 1, can-block = record +8 != 0)`, then the roll `005a89a0` and apply `0057c6c0` (all of this: monster-attack-trace Q2, 100,000 cases on block and roll). New here: which rows pass the can-block gate = the missile's stat 0x16 > 0 after launch: 24 rows blockable, 25 not (lists below) | V read (earlier trace); V run (record max per row) |
| Q6 | **Flight and reach.** path velocity per frame = ((Vel + VelLev x level / 8) << 8) x 75 / 100 / 4096 subtiles, Range frames, 25 frames per second (S: sk-missile-trace section 1-2, upstream `missiles.py` proof). Reach at level 1 = Range x Vel x 0.046875: Fallen Shaman bolt 15-30 subtiles at 9-19 subtiles per second, Skeleton Archer arrow 45 at 28, Corrupt Archer arrow 26-38 at 16-23, Quill Rat spike 19-38 at 12-23, Sand Maggot spit 11 at 12, Unraveler spit 5.6 at 6, Blood Raven arrow 37.5, javelin 75. VelLev adds speed with the level: spike1 +1 Vel per level (NM level 4: 13, Hell level 8: 18) | D + S (arithmetic; not re-run) |

### Q4 table: game.exe's record per row (V run, tables D). Points, N / NM / H = Normal / Nightmare / Hell, level 1 / 4 / 8

Physical and element come from the missile's own stats right after `0059fa30` (so before resistance, armor, block, the roll). Poison total = per-frame stat 0x39 x length 0x3b / 256. "El chance" rows add an element on a fraction of the hits (the monster's El Pct, shown as N / NM / H percent); one launch cannot show it, so it is not in the cells. Blood Raven's rows carry her weapon (random).

| row (id) | shooters (source) | physical pts N / NM / H | element pts N / NM / H | poison total pts (per-frame x len / 256) N / NM / H | ToHit | blockable | reach |
|---|---|---|---|---|---|---|---|
| bighead1 (2) | bighead1 (MissA2) | - / - / - | ltng 6-8 / ltng 43-55 / ltng 121-152 | - / - / - | - | no | 15.0 (vel 8, range 40) |
| bighead2 (3) | bighead2 (MissA2) | - / - / - | ltng 6-8 / ltng 43-55 / ltng 121-152 | - / - / - | - | no | 22.5 (vel 12, range 40) |
| bighead3 (4) | bighead3 (MissA2) | - / - / - | ltng 3-4 / ltng 43-55 / ltng 121-152 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| bighead4 (5) | bighead4 (MissA2) | - / - / - | ltng 4-6 / ltng 43-55 / ltng 121-152 | - / - / - | - | no | 18.8 (vel 10, range 40) |
| spike1 (7) | quillrat1 (MissA2) [El chance 0/15/20 %: element is a per-hit chance] | 1-2 / 19-26 / 44-61 | - / - / - | - / - / - | 1 | yes | 18.8 (vel 10, range 40) |
| spike2 (8) | quillrat2 (MissA2) [El chance 0/5/15 %: element is a per-hit chance] | 1-4 / 19-27 / 44-62 | - / - / - | - / - / - | 1 | yes | 24.4 (vel 13, range 40) |
| spike3 (9) | quillrat3 (MissA2) [El chance 0/5/15 %: element is a per-hit chance] | 1-5 / 19-28 / 44-63 | - / - / - | - / - / - | 1 | yes | 30.0 (vel 16, range 40) |
| spike4 (10) | quillrat4 (MissA2) [El chance 0/5/15 %: element is a per-hit chance] | 2-5 / 22-31 / 51-70 | - / - / - | - / - / - | 1 | yes | 37.5 (vel 20, range 40) |
| cr_arrow1 (13) | cr_archer1 (MissA1) | 2-3 / 18-27 / 45-68 | - / - / - | - / - / - | 1 | yes | 26.2 (vel 14, range 40) |
| cr_arrow2 (14) | cr_archer2 (MissA1) [El chance 0/10/20 %: element is a per-hit chance] | 2-3 / 18-27 / 45-68 | - / - / - | - / - / - | 1 | yes | 30.0 (vel 16, range 40) |
| cr_arrow3 (15) | cr_archer3 (MissA1) [El chance 0/10/20 %: element is a per-hit chance] | 3-4 / 18-27 / 45-68 | - / cold 12-18 / cold 30-45 | - / - / - | 1 | yes | 33.8 (vel 18, range 40) |
| cr_arrow4 (16) | cr_archer4 (MissA1) [El chance 0/10/20 %: element is a per-hit chance] | 4-6 / 18-27 / 45-68 | - / - / - | - / - / - | 1 | yes | 37.5 (vel 20, range 40) |
| shafire1 (22) | fallenshaman1 (Skill2:ShamanFire) | - / - / - | fire 1-4 / fire 31-34 / fire 71-74 | - / - / - | - | no | 15.0 (vel 8, range 40) |
| shafire2 (23) | fallenshaman2 (Skill2:ShamanFire) | - / - / - | fire 2-6 / fire 32-36 / fire 72-76 | - / - / - | - | no | 18.8 (vel 10, range 40) |
| shafire3 (24) | fallenshaman3 (Skill2:ShamanFire) | - / - / - | fire 2-7 / fire 32-37 / fire 72-77 | - / - / - | - | no | 22.5 (vel 12, range 40) |
| shafire4 (25) | fallenshaman4 (Skill2:ShamanFire) | - / - / - | fire 2-8 / fire 32-38 / fire 72-78 | - / - / - | - | no | 30.0 (vel 16, range 40) |
| andarielspray (32) | andariel (Skill1:AndrialSpray) | 7-10 / 22-25 / 42-45 | - / - / - | 50.0 / 215.0 / 470.0 | - | yes | 28.1 (vel 15, range 40) |
| mummy1 (63) | unraveler1 (MissA2) | - / - / - | - / - / - | 7.0 / 16.4 / 28.9 | - | no | 5.6 (vel 5, range 24) |
| mummy2 (64) | unraveler2 (MissA2) | - / - / - | - / - / - | 14.1 / 32.8 / 57.8 | - | no | 5.6 (vel 5, range 24) |
| mummy3 (65) | unraveler3 (MissA2) | - / - / - | - / - / - | 21.1 / 56.2 / 103.1 | - | no | 5.6 (vel 5, range 24) |
| mummy4 (66) | radament (MissA2) | - / - / - | - / - / - | 28.1 / 75.0 / 137.5 | - | no | 5.6 (vel 5, range 24) |
| goospit1 (70) | maggotqueen1 (MissA2) | <0.1-<0.1 / <0.1-<0.1 / 0.2-0.2 | - / - / - | 6.2 / 20.3 / 39.1 | - | yes | 11.2 (vel 10, range 24) |
| goospit2 (71) | maggotqueen2 (MissA2) | <0.1-<0.1 / <0.1-<0.1 / 0.2-0.2 | - / - / - | 11.7 / 32.8 / 60.9 | - | yes | 11.2 (vel 10, range 24) |
| goospit3 (72) | maggotqueen3 (MissA2) | <0.1-<0.1 / <0.1-0.1 / 0.2-0.2 | - / - / - | 18.8 / 46.9 / 84.4 | - | yes | 11.2 (vel 10, range 24) |
| unholybolt1 (77) | radament (Skill3:UnHolyBolt) | - / - / - | mag 3-6 / mag 30-33 / mag 66-69 | - / - / - | - | no | 22.5 (vel 12, range 40) |
| unholybolt2 (78) | unraveler2 (Skill3:UnHolyBolt) | - / - / - | mag 3-7 / mag 30-34 / mag 66-70 | - / - / - | - | no | 22.5 (vel 12, range 40) |
| unholybolt3 (79) | unraveler3 (Skill3:UnHolyBolt) | - / - / - | mag 4-8 / mag 31-35 / mag 67-71 | - / - / - | - | no | 22.5 (vel 12, range 40) |
| frostnova (119) | summoner (Skill2:Frost Nova) | - / - / - | cold 2-4 / cold 8-11.5 / cold 16-21.5 | - / - / - | - | no | 15.8 (vel 24, range 14) |
| skmage1 (125) | skmage_pois3 (MissA1) | - / - / - | - / - / - | 23.4 / 78.1 / 195.3 | - | no | 16.9 (vel 9, range 40) |
| skmage2 (126) | skmage_cold4 (MissA1) | - / - / - | cold 6-12 / cold 18-46 / cold 45-114 | - / - / - | - | no | 22.5 (vel 12, range 40) |
| skmage3 (127) | skmage_fire1 (MissA1) | - / - / - | fire 4-7 / fire 21-52 / fire 53-129 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage3 (127) | skmage_fire2 (MissA1) | - / - / - | fire 4-8 / fire 21-52 / fire 53-129 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage3 (127) | skmage_fire3 (MissA1) | - / - / - | fire 7-12 / fire 21-52 / fire 53-129 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage3 (127) | skmage_fire4 (MissA1) | - / - / - | fire 7-13 / fire 21-52 / fire 53-129 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage4 (128) | skmage_ltng1 (MissA1) | - / - / - | ltng 2-9 / ltng 12-62 / ltng 30-152 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage4 (128) | skmage_ltng2 (MissA1) | - / - / - | ltng 2-10 / ltng 12-62 / ltng 30-152 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage4 (128) | skmage_ltng3 (MissA1) | - / - / - | ltng 4-15 / ltng 12-62 / ltng 30-152 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| skmage4 (128) | skmage_ltng4 (MissA1) | - / - / - | ltng 4-16 / ltng 12-62 / ltng 30-152 | - / - / - | - | no | 26.2 (vel 14, range 40) |
| vampirefirewallmaker (130) | summoner (Skill4:VampireFirewall) | - / - / - | - / - / - | - / - / - | - | no | 5.6 (vel 12, range 10) |
| vampiremeteorcenter (133) | vampire1 (Skill3:VampireMeteor) | - / - / - | fire 5-10 / fire 20-25 / fire 40-45 | - / - / - | - | no | 0.0 (vel 0, range 60) |
| raven1 (135) | bloodraven (MissA1) [weapon: random] | 1-5 / 1-19 / 1-6 | fire 1-5, ltng 1-8 / - / - | 12.0 / - / 7.0 | 1 | yes | 37.5 (vel 16, range 50) |
| raven1 (135) | bloodraven (Skill2:Quick Strike) [weapon: random] | 1-5 / 2-7 / 1-12 | - / cold 3-15 / ltng 1-97 | - / - / - | 1 | yes | 37.5 (vel 16, range 50) |
| skbowarrow1 (196) | sk_archer1 (MissA1) | 2-3 / 10-15 / 26-38 | - / - / - | - / - / - | 1 | yes | 45.0 (vel 24, range 40) |
| skbowarrow2 (197) | sk_archer2 (MissA1) | 3-5 / 10-15 / 26-38 | - / - / - | - / - / - | 1 | yes | 45.0 (vel 24, range 40) |
| skbowarrow3 (198) | sk_archer3 (MissA1) | 4-6 / 10-15 / 26-38 | - / - / - | - / - / - | 1 | yes | 45.0 (vel 24, range 40) |
| skbowarrow4 (199) | sk_archer4 (MissA1) | 4-6 / 10-15 / 26-38 | fire 2-4 / fire 7-17 / fire 19-41 | - / - / - | 1 | yes | 45.0 (vel 24, range 40) |
| skbowarrow5 (200) | sk_archer5 (MissA1) | 5-8 / 10-15 / 26-38 | ltng 0-9 / ltng 1-24 / ltng 3-91 | - / - / - | 1 | yes | 45.0 (vel 24, range 40) |
| pantherjav1 (226) | slinger1 (MissA1) [El chance 0/5/15 %: element is a per-hit chance] | 5-12 / 15-26 / 38-64 | - / - / - | - / - / - | 1 | yes | 75.0 (vel 20, range 80) |
| pantherjav2 (227) | slinger2 (MissA1) [El chance 0/5/15 %: element is a per-hit chance] | 5-12 / 15-26 / 38-64 | - / - / - | - / - / - | 1 | yes | 75.0 (vel 20, range 80) |
| pantherjav4 (229) | slinger4 (MissA1) [El chance 0/5/15 %: element is a per-hit chance] | 6-15 / 15-26 / 38-64 | - / fire 9-17 / fire 30-57 | - / - / - | 1 | yes | 75.0 (vel 20, range 80) |
| pantherpotorange (238) | slinger5 (MissA1) | 3-5 / 18-20 / 38-40 | fire 2-5 / fire 23-26 / fire 51-54 | - / - / - | - | yes | 96.0 (vel 16, range 128) |
| pantherpotgreen (239) | slinger6 (MissA1) | - / - / - | - / - / - | 28.1 / 84.4 / 159.4 | - | no | 96.0 (vel 16, range 128) |
| blowgun (266) | fetishblow1 (MissA1) [El chance 0/7/12 %: element is a per-hit chance] | 1-2 / 6-10 / 15-25 | - / - / - | - / - / - | 1 | yes | 84.4 (vel 20, range 90) |
| blowgun (266) | fetishblow2 (MissA1) [El chance 0/7/12 %: element is a per-hit chance] | 2-4 / 6-10 / 15-25 | - / - / - | - / - / - | 1 | yes | 84.4 (vel 20, range 90) |
| regurgitatorcorpse (341) | regurgitator1 (MissA2) | 25-55 / 70-100 / 130-160 | - / - / - | - / - / - | - | yes | 131.2 (vel 14, range 200) |
| leapknockback (346) | sandleaper1 (Skill1:Leap) | - / - / - | - / - / - | - / - / - | - | no | 0.0 (vel 0, range 2) |
| countessfirewallmaker (653) | corruptrogue3 (Skill1:CountessFirewall) | - / - / - | - / - / - | - / - / - | - | no | 3.9 (vel 12, range 7) |
labrun: orion exited 0 after 1s

Not in the table (own rows, one line each, V run): Andariel spray (id 32): phys 7-10 / 22-25 / 42-45, poison 50 / 215 / 470 total; Summoner Frost Nova (119): cold 2-4 / 8-11.5 / 16-21.5; Countess / vampire fire walls and meteors (653, 130, 133): no physical part (the wall's fire is its own hit function, sk-ground2-trace); Sand Leaper (346): no damage record.

The lists the run printed (rows by the can-block gate, `mm_launch.py` last lines):
- physical part, blockable: andarielspray blowgun cr_arrow1-4 goospit1-3 pantherjav1,2,4 pantherpotorange raven1 regurgitatorcorpse skbowarrow1-5 spike1-4 (24 rows)
- no physical part, not blockable: bighead1-4 countessfirewallmaker frostnova leapknockback mummy1-4 pantherpotgreen shafire1-4 skmage1-4 unholybolt1-3 vampirefirewallmaker vampiremeteorcenter (25 rows)
- ToHit set (the AR roll runs): blowgun cr_arrow1-4 pantherjav1,2,4 raven1 skbowarrow1-5 spike1-4 (17 rows)

## The differential (V run on the game side)

`tools/emu/mm_launch.py` (copied next to upstream's `missiles.py`, run by `tools/emu/mm_run.sh`; the tables from `mpq-cat` into one scratch dir): per difficulty a fresh booted Blood Moor game (the game's difficulty byte 0 / 1 / 2, expansion on); per launch row the REAL monster is spawned (`00555230`, class = hcIdx), its mode stats set by game.exe's own `005a4f50`, the missile made by `0059fa30` with the real level (0 / 3 / 7 + 1), and the missile's stats read with `00625480`. 71 rows x 3 difficulties = **213 launches**; all of Normal, NM and Hell for the same rows. Oracles in the script:
- the reading: owner damage x S / 128 + row part with brackets, 6 differ out of 213 -- all Blood Raven (the weapon path, not modelled; with brackets left out of the reading it was 36, so the brackets are the whole rest);
- owner stat vs MonLvl x percent / 100: 0 of 213 differ;
- d2d's expression (`ai.cpp:2014-2017`, `src_min x src_damage / 128 + missile_info.min`, whole points, no bracket, no HitShift) as a PORT: 43 of 213 differ = 28 level bracket (D1), 6 HitShift (D2), 6 Blood Raven weapon, 3 Andariel (her spray goes through `andariel_missile`, not this path: port not applicable, not counted);
- elements on the MissA path (monster El rows with the shot's mode, `ai.cpp:2019-2020`): 125 cases compared, 4 differ (Blood Raven's weapon element, the pot's row fire x 3), 28 skipped (El Pct 1..99, a chance);
- `mm_elev.py`: `0064b100 / 0064b1d0 / 0064b2a0` / `0064af20 / 0064aff0` called at level 1 / 4 / 8: goospit1 element 8 / 26 / 50 per frame, physical 2-6 / 20-24 / 44-48 (256ths); shafire1 element min 256 / 7936 / 18176.

## Deviations d2d vs game.exe

| D | Rule, game.exe | Rule, d2d | Tag | Differential |
|---|---|---|---|---|
| D1 | physical min / max carry `(level - 1) x MinLevDam / MaxLevDam` at the missile level (1 / 4 / 8): spike1 NM +9, Hell +21; goospit, regurgitatorcorpse (+45 / +105), pantherpotorange, spike1-4 | `missile_info.min` is the raw `MinDamage` column; the loader reads only `MinELev` brackets (`gamedata_load.cpp:357`, `gamedata.hpp:712`); `ai.cpp:2016-2017` adds raw min / max | V run (game) + V-d2d read (`ai.cpp`, no `min_lev` for missiles in components) | 28 of 213 launches (none Normal); real-d2d version buildable: an apps `test_game` mode that spawns a monster at difficulty d, fires its MissA through `Fight` and prints `missile.src.a2_min / a2_max`, diffed with this script's `game256` column |
| D2 | `MinDamage << HitShift` in 256ths (`0064af20`, `shl cl`, byte +0x196): goospit shift 1 -> 2-6 256ths | raw column as whole points: 1-3 points | V run + V-d2d read | 6 of 213 (Normal sand maggots and maggot queens, each of three rows x 2) |
| D3 | the row's own element rides on the missile: poison goospit1-3 (8 / 10 / 12 per frame x 200 / 300 / 400 frames), mummy1-4 (18 x 100..400), pantherpotgreen, fire pantherpotorange | the MissA path (`ai.cpp:2011-2022`) carries only the monster's own El1..3; `row_damage` of the row is read only in `andariel_missile` (skill missiles) | V run (game side), V-d2d read only | not run on d2d; 12 launch rows (9 missile rows) x 3 difficulties = 36 launches in the table; same buildable test as D1 (print `missile.src.elements`) |
| D4 | can-block = missile physical max (stat 0x16) != 0: 25 of 49 rows unblockable (list above) | always blockable (`combat.hpp` monster_blow, monster-attack-trace D(a)) | V run (record) + S (earlier 100,000 cases on the gate) | yes: `mon_missile.py block` (existing); the table here gives the rows |
| D5 | Blood Raven (the one Act 1 / 2 shooter with an inventory): damage from her weapon, random per launch, plus the weapon's elements | A1 stat x 128 / 128 | V run (the spread) + V read (`0064bac5`); formula U | none; build: spawn bloodraven, read her weapon with the item dump, compare the record |
| D6 (I) | missile level replaced by the monster's level when monster-data flag bit 9 is set (`005a6d50`) | always bonus + 1 | V read only; which monsters carry the flag U | none yet: a population sweep reading unit +0xc4 bit 9 over the spawn lists (monsters.py population) |

Also from earlier traces, unchanged: d2d rolls whole points inclusive, game.exe 256ths top exclusive (D5 of sk-missile-trace, D(roll) of monster-attack-trace); the MissA ToHit gate (b) and the A2-shooter AR at impact (monster-attack-trace).

## Bot findings (described, not changed)

- **B1 The dodge has no threat model.** `internal/combat/fight.go:492-507` dodges every hostile missile inside `DodgeRadius` (default 4 subtiles, `internal/env/tuning_combat.go:85`) once per 1.2 s, whatever it carries. A Skeleton Archer arrow covers 4 subtiles in 3.6 frames (0.14 s, 1.13 subtile per frame), a Corrupt Archer arrow 0.17-0.24 s, a Quill Rat spike 0.17-0.34 s; only the slow bolts leave time (Fallen Shaman fire 0.43 s, Unraveler spit 0.7 s, Sand Maggot spit 0.34 s at 4 subtiles but 1.0 s from its 11-subtile reach). With a poll plus a click round trip of the order of 100-300 ms (I) the radius cannot dodge arrows and spikes; it can dodge shaman and mage bolts, the cheap ones in Normal.
- **B2 The sidestep list never matches Act 1 / 2 shooters.** `internal/d2/missiles.go:45-57` names four kinds (fire wall, hydra, chaos, anything containing "lightning"); the shooters' rows are spike, cr_arrow, skbowarrow, shafire, goospit, mummy, skmage1-4, bighead, unholybolt, blowgun, pantherjav (all MissileOther, never sidestepped, `internal/combat/dodge.go:77-84`); the knob is off by default anyway (`tuning_combat.go:85`, MissileDodge 0).
- **B3 Cheap versus costly hits are not told apart.** No table of missile id -> damage and element exists (grep: `hostileMissiles`, `internal/combat/position.go:92`, lists positions only). Per this trace the costly ones are NM / Hell: Skeleton Mage fire 53-129 and lightning 30-152 (unblockable, resist-gated), the Quill Rat spike 44-70 (blockable), the Unraveler / Radament poison 29-137 total, Sand Maggot poison 39-84 total, Fallen Shaman fire 31-34 (NM) and 71-74 (Hell). The cheap ones are all of Normal Act 1: shaman fire 1-4, spike 1-5, arrows 2-6. A table keyed by row id (this doc's Q4) is the missing input.
- **B4 Kill-first rules are name-based.** `internal/combat/target.go:76` names shaman, oblivion, unraveler, bloodraven; correct as a heuristic (the shamans' fire scales to 71-74 in Hell, the Unraveler / Radament poison to 29-137), but Skeleton Mages (`skmage_*`: the biggest unblockable single hits, 53-129 fire, 30-152 lightning in Hell) are in no list.
- **B5 The Fetish blowdart family** (`fetishblow1-4`: 1-4 / 6-10 / 15-25 points, ToHit, blockable, reach 84) is not a danger case (`internal/combat/rules.go:69` treats it only as an explosion keep-range rule for the Stygian Doll); it does not appear in Acts 1 / 2 spawn lists (D), so Act 3 only.
- **B6 d2d side of the telemetry.** the d2d dump feeds no `Snapshot.Missiles` (B1 of monster-attack-trace) so all of the above is moot on d2d until it does.

## PR candidate

None. D1 / D2 / D3 are shown as a game.exe record against a port of d2d's expression plus a source read; the existing differential harness (SKMISS style) compares game.exe with the REAL d2d build (`test_missile_hit`, `rules::missile_blow`), and no such real-d2d path exists for the monster launch in `Ai::tick` (the record is built inline at `ai.cpp:2014-2017`, not in `rules`). To make it a candidate: move the record build into `rules::monster_missile_damage(info, owner_stats, mode, level)` (a small upstream refactor), add the three terms (D1 bracket via the existing `level_bonus`, D2 `<< hitshift`, D3 the row element via `row_damage`), a `test_monster_missile` mode printing min / max / elements per (row, difficulty), and diff its output with `mm_launch.py`'s `game256` columns; before > 0 (28 + 6 + 36 expected), after 0 is the gate. An independent verifier is needed before any PR text. Nothing was built here.

## Drill recipe (lab only; docs/forkdbg.md, docs/forkcheats.md; running a game is not part of this card)

Binary `bin/d2d.lab-debug-skills` as `D2D_BIN`, one hidden slot, a throwaway Sorceress `--no-save`, `d2drill -probe` so every verb is ledgered. Blood Moor (`debug goto 2`), `debug missiles` (rows: owner type 1 = a monster, speed in cells per second = subtiles x 5 per frame x 25 / 5, ttl frames), `dump stats` raw[6] life, `monsters`.
- **M1 (D1, the lab already shows the surface):** `debug spawn quillrat1 1` next to her (the record comes from the table, so no `debug mon set`); a Nightmare / Hell lab game is needed (no difficulty verb is documented in forkdbg.md: U whether one exists), `step` until the spike flies, read the player's life before and after each hit (raw[6]) over 30 hits: game.exe NM spike 19-26 points, d2d today 10-17; Hell 44-61 vs 23-40. Same for `cr_archer1` / `sk_archer1` (arrows, no bracket: equal in every difficulty, the control).
- **M2 (D2, D3):** `debug spawn sandmaggot1 1` (Act 2 level `debug goto <act 2 id>`), `step` through one spit: game.exe puts 8 poison per frame for 200 frames (6.25 points total) on the player (`dump stats` poison state, `debug states`); d2d today: the hit does ~2 points of physical and no poison state. Unraveler (`unraveler1`) the same: poison 18 per frame for 100 frames.
- **M3 (D4):** give the player a shield with a high block chance (cheat ledger), `debug spawn fallenshaman1 6` and count blocks over 200 bolts: game.exe 0 (no physical part), d2d > 0.
- **M4 (D5, D6):** `debug spawn bloodraven 1`, 30 arrows, histogram of the player's damage: game.exe a spread by her weapon (1-5 Normal), d2d the A1 range.
- Expected on d2d today: every row of Q4 with a bracket (spike NM / Hell, goospit, regurgitatorcorpse) deals less than the table; the poison rows deal none; shaman fire in NM / Hell is equal (the skill path applies the row brackets, `ai.cpp:384-390`).

## Matrix cell updates (the monster / missile matrix is not on main: proposed text)

| Row | Cell | New value |
|---|---|---|
| rule-coverage `missiles` 2-5, 7-10, 13-16, 22-25, 32, 63-66, 70-72, 77-79, 125-128, 130, 133, 135, 196-200, 226, 227, 229, 238, 239, 266, 341, 346, 653 | status | READ-ONLY (game side V run, d2d side read): ledger.tsv line added; promote to TESTED-DEVIATION when the real-d2d test above exists |
| risk map rank 1 (monster missiles) | evidence | docs/research/monster-missile-trace.md: Normal agrees, NM / Hell physical brackets missing (D1), row elements missing (D3) |
| monster-attack-trace D(a) can-block | rows | 25 of 49 Act 1 / 2 rows unblockable in game.exe (list above) |

## Dead ends and corrections

- First pass read the monstats `A1MinD` column as points: they are percentages of the MonLvl damage (`L-DM`); a launch run with them showed 106 "owner differs" lines, all my table misreading. The V run reads the stat from the spawned unit instead.
- First pass launched every row at level 1 and concluded "no brackets, no scaling": wrong; the row's brackets and the 1 / 4 / 8 levels are the main NM / Hell effect (`005a6d50`). The difficulty bonus is read there; the unit-level flag is open (D6).
- Elements first compared from single launches flagged chance rows (El Pct 1..99) as differences; they are now skipped and counted.
- Brief's names with no missile: Cold Crow, Scarab, Carrion Bird, Gloombat, Vulture have no MissA / MissS / MissC row and no missile skill in monstats.txt (V, `mm_tbl.py` over all enabled rows): melee only. "Desert / Dune spitters": the Corpse Spitter (`regurgitator1-3`, 25-55 physical Normal) and Sand Maggots (poison); the Corpse Spitter is not in the Act 1 / 2 spawn lists (Act 3).
- Not run: the real d2d binary (D1-D3 d2d side), flight of these rows (S only), the hit roll on the player (monster-attack-trace), Blood Raven's weapon formula, flag bit 9.

## Reproduce (the worker host)

```sh
# tables: mpq-cat through a scratch dir with a lowercase patch_d2.mpq (docs/process/mpq-data-reading.md): monstats missiles skills levels monlvl
python3 tools/emu/mm_tbl.py <tables dir>      # all enabled rows with a missile source
python3 tools/emu/mm_acts.py <tables dir>     # Act 1 / 2 rows (stat columns are percents of the MonLvl damage, no points)
tools/emu/mm_run.sh "" <tables dir>           # game.exe launches: prints the summary lines quoted above (about 7 s)
tools/emu/mm_run.sh "" <tables dir> mm_elev.py
python3 tools/emu/mm_summary.py <launch output> <tables dir>   # the Q4 table
```

## Shared-file lines for the merger

In `openspec/changes/monstermissile/merge-lines.txt`.
