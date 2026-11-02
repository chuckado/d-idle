# Champions, uniques, minions and the Act 1 / 2 superuniques: game.exe 1.14d vs d2d vs our bot (UNIQMODS, follow-up of decomp-queue item 13)

> status: current (2026-11-01, card uniqmods, agent d2-decomp on Sonnet). Offline, docs and tools only: no game, no d2d slot, no Go change; nothing pushed or posted. **Upgrade run (second session, 2026-11-01): the worker-host emulator ran game.exe's own room population again (new section Q10: 24 runs, 8 seeds each, three difficulties, Acts 1 and 2, about 3,700 unit lines) and confirms the earlier S-run Normal values byte for byte; tag "V run" below means that run. It is game.exe against itself, not a differential against d2d (the d2d half of the differential is the 157,618-case run in monster-mods-trace.md).** The first session had no emulator, so older "S-run" tags in Q2..Q8 remain as written unless Q10 restates the number. Every number below is either **D** (the 1.14d patch-layer tables, read by header name), **V read** (disassembly of game.exe, listings left by the DECOMPMODS card), **S-run** (the DECOMPMODS run log of game.exe's own code under the emulator, `umod_spawn.py`, 526 lines, re-summarised here by `tools/emu/um_summ.py`, not re-run), **S** (upstream docs `re/monsters.md`, `re/superuniques.md`), **I** or **U**. This doc does not repeat [monster-mods-trace.md](monster-mods-trace.md) (what each init writes, the 157,618-case differential; read it first) or [pr-mods-proof.md](pr-mods-proof.md) (D1 narrowed, D2, D3, B built and merged upstream as #157); it adds the pick rules, the pack and minion counts, the Act 1 / 2 superunique records, the multiplier summary by class and difficulty, and the bot side. d2d read at the upstream checkout the DECOMPMODS follow-up used (components/rules/uniques.hpp, components/game/objgroups.cpp).

## Answer in six lines

1. **Pick rules (V read, D)**: a random unique rolls `rand(1) + 1 + difficulty` mods (Normal 1, Nightmare 2, Hell 3; FUN_005a0760), weighted by MonUMod `upick`: 12 rows of weight 6 in Normal (72 in all; Magic Resistant has upick 0 in Normal), 13 rows in Nightmare and Hell (78). A champion (20 %) takes one of 5 kinds at weight 1 each. Exclusions: row disabled, expansion-only rows (version 100) in a classic game, MonType (`exclude1/2`, through a type-inheritance bit matrix), fPick (Extra Fast needs the walk mode, Teleportation the A1 mode, Multiple Shots a flag test).
2. **Pack sizes (V read + S-run)**: a champion event is 2..4 units (leader plus `rand(3) + 1`, FUN_0054e1e0, V read; six packs seen: 3, 3, 3, 4, 4, 4); a random unique brings 3..6 minions (S-run: 44 uniques, min 3, max 6, mean 4.5). The leader holds the champion kind, the followers all carry mod 16.
3. **Multipliers (D + S-run)**: life x2 / x3 / x4 for minion / champion / unique in Normal (x1.75 / 2.5 / 3 NM, x1.5 / 2 / 2 Hell); level +3 unique and minion, +2 champion; experience x5 unique, superunique and minion, x3 champion, x1 act boss (Blood Raven, the Smith); damage / to-hit bonuses from the MonUMod `constants` times the DifficultyLevels 90 / 75 / 66 percent.
4. **Act 1 Normal superuniques (S-run, 12 seeds each)**: Bishibosh mods 8 9 (+22), 2 minions; Rakanishu 17 6 (+22), 8 minions; Treehead WoodFist 13 5 6 (+22), 2 minions; Corpsefire 27 (+22), 5 minions. The minions carry the mods' minion effects (Rakanishu's and Treehead's minions are Extra Fast, Treehead's Extra Strong), not the resistances. Act 2 superunique records: not extracted (U), recipe in the drill section.
5. **d2d**: the pick rules match for Acts 1 and 2 (one new, low-impact deviation D1: fPick 1 / 3 untested, and D2: exclusion by MonType parent, U). Nothing here is a PR candidate.
6. **Bot**: seven findings B1..B7; the biggest are that champion packs are 2..4, not "3-5" (B1), that the +50 % champion damage rule ignores berserkers (+270 %) (B2), and that the fallback resist model applies unique-only resist bonuses to minions (B4).

## Q1. The MonUMod table (D)

Source: `monumod.txt` of the 1.14d patch layer (read from the worker host's extracted copy; 43 rows, ids 0..42; version 0 rows are in Classic, 100 are the expansion rows 32..42). `tools/emu/um_tab.py monumod|weights` prints these columns.

| id | mod | unique weight (N / NM / H) | champion weight | fPick | exclude | constants (see Q4) |
|---|---|---|---|---|---|---|
| 1, 2, 3, 4 | rndname, hpmultiply, light, leveladd | fixed mods of every unique (FUN_005a2120, 0x6e2168) | | | | |
| 5 | strong | 6 / 6 / 6 | | | | 150 |
| 6 | fast | 6 / 6 / 6 | | 3 | | 100 |
| 7 | curse | 6 / 6 / 6 | | | | 300 |
| 8 | resist (Magic Resistant) | **0** / 6 / 6 | | | | 200 |
| 9 | fire (Fire Enchanted) | 6 / 6 / 6 | | | | 100 |
| 17 | lightning (Lightning Enchanted) | 6 / 6 / 6 | | | sandleaper | 33 |
| 18 | cold (Cold Enchanted) | 6 / 6 / 6 | | | | 33 |
| 25 | manahit (Mana Burn) | 6 / 6 / 6 | | | | 50 |
| 26 | teleport | 6 / 6 / 6 | | 1 | | 50 |
| 27 | spectralhit | 6 / 6 / 6 | | | | 50 |
| 28 | stoneskin | 6 / 6 / 6 | | | | 66 |
| 29 | multishot | 6 / 6 / 6 | | 2 | | 66 |
| 30 | aura (Aura Enchanted) | 6 / 6 / 6 | | | | 66 |
| 16, 36, 37, 38, 39 | champion, ghostly, fanatic, possessed, berserk | | 1 each | | | 16: 0 |
| 10..15, 19..24, 31..35, 40..42 | poisondead, durieldead, bloodraven, rage, spcdamage, partydead, hireable, scarab, killself, questcomplete, poisonhit, thief (disabled), goboom, firespike_explode, suicideminion_explode, ai_after_death, shatter_on_death, worms_on_death, always_run_ai, lightningdeath | no pick weight: only fixed mods, superunique records or act-boss code give them | | | | |

Sums (D, `um_tab.py weights`): unique weights 72 / 78 / 78 over 12 / 13 / 13 rows. So the chance that a first pick is a given mod in Normal is 6/72 = 8.3 % before the fPick and exclude filters, Magic Resistant 0 %; in Nightmare and Hell 6/78 = 7.7 % each. Later picks of the same monster are drawn without the mods it already holds. A champion's kind is one of five at 1/5 (rows 36..39 are expansion rows, so in an expansion game all five). Tag D.

## Q2. When the mods are rolled (V read, S)

| Step | Address | Rule | Tag |
|---|---|---|---|
| quota | FUN_005be020 | the level's MonUMin / MonUMax (Levels.txt) and the room counter decide whether a room's spawn becomes a unique / champion event; a superunique counts as one of the quota; Normal Act 1: Blood Moor and Den none, Cold Plains 1 / 1, Stony Field 1 / 2, Dark Wood 1 / 2; Nightmare 4 / 5, 4 / 5, 4 / 6, 4 / 6 | V read, D (spawn-density-trace section 3) |
| the roll | FUN_005a0760 (`005a0820`..`005a0830`) | `rand(100) <` MonUMod row 0 constant (20): the monster becomes a champion (flag byte |= 4) and FUN_005a0500 picks one kind from the champion weights; else `rand(1) + 1 + difficulty` unique mods (`005a0887`: the difficulty is `[game + 0x6d]`), FUN_005a0600 picks from the unique weights with a 9-byte "taken" array, so none twice; the list lives at monster data +0x1c (9 bytes, a monster already holding 8 or more is skipped, the count is capped so the total stays at most 9) | V read |
| allowed | FUN_005a03e0 | see Q3 | V read |
| the company | FUN_005a0c00 (unique), FUN_0054e1e0 (champion) | see Q5 | V read, S-run |
| the mods' effects | FUN_005a2120 | fixed mods 1..4 then the unit's own, each through the init table at 0x73c008 with flag 1 on the boss and flag 0 on every party member | S (monster-mods-trace.md), V read |
| superunique | FUN_005a49b0 | SuperUniques Mod1..3, then the quest mod 22 for every superunique (an empty hook for Act 1, upstream `re/superuniques.md`), then `difficulty` more unique mods by `upick`; minions `MinGrp..MaxGrp` each plus `difficulty` when both are set | S, V (S-run lines `8.9.22`, `17.6.22`) |
| act boss | FUN_005b1cf0 | mods 12 / 22 (Blood Raven), 22 (Andariel, the Smith) and flag 8 only: no level, life or experience change | S, S-run (Blood Raven: level 10 -> 10, exp x1, life 113 -> 113) |

Normal uniques therefore carry exactly one rolled mod (S-run: every random `f9` line of Cold Plains, Stony Field and Dark Wood holds one mod, one exception below), Nightmare two, Hell three. The superuniques' own mods are fixed in the record and keep that number whatever the difficulty; they gain `difficulty` extra rolled mods on top.

## Q3. Which monsters can take which mod (V read of FUN_005a03e0; the DECOMPMODS listing)

Read at `005a03e0`..`005a04f3` (arg1 = the MonUMod row, ebx = the monster unit, arg2 = the expansion flag):

1. `cmp byte [row + 6], 0` (enabled): a disabled row (24 thief) is refused. V read.
2. If arg2 is 0 (classic) and `word [row + 4] >= 0x64` (version): refused. In an expansion game every row passes. V read.
3. `exclude1` / `exclude2` (words at `row + 0xa`, `+ 0xc`, MonType ids; `<= 0` means none): each calls FUN_005a0070(monster, type), which reads the monster's MonStats MonType (`word [MonStats + 0x1c]`) and tests one bit of a MonType x MonType matrix at `tables + 0xc44` (row length `+ 0xc48`, bit = type `& 0x1f`): a bit matrix, i.e. "this monster's type is, or derives from, the excluded type". A set bit refuses the row. Only row 17 (Lightning Enchanted) has an exclude (`sandleaper`). V read; which MonTypes derive from `sandleaper`: U (montype.txt not extracted).
4. fPick (byte `row + 9`): **1** (Teleportation): FUN_0046c140(class, 4) must be true (mode bit 4 = A1, a melee attack animation; FUN_0046c140's bit table is MonStats2 `mDT mNU mWL mGH mA1 ...`, upstream `re/monster-ai.md`: bit 2 = mWL, bit 4 = A1) else refused; **3** (Extra Fast): FUN_0046c140(class, 2) (mode WL, walk) must be true else refused; **2** (Multiple Shots): refused when `MonStats byte + 0xc & [0x6ce26c]` is set, else allowed only when `byte + 0xe & [0x6ce26c]` is clear (the flag meaning is not read here; upstream and d2d call it "melee": S, U); `fPick 0`, anything else: allowed. V read.

Table result (D, `tools/emu/um_fpick.py`, MonStats x MonStats2): the enabled, non-NPC rows without the walk mode or without A1 are crow nests, maggot eggs and queens (Coldworm's row 284..288), bone walls and prisons, towers, traps, tentacles, catapults, the Summoner (no walk: Extra Fast refused), doom knight 3, Act 5 barbarians and suicide minions (no A1: Teleportation refused). None of these is a regular Act 1 / 2 spawn that can roll mods (crow nests swap their unique for a foul crow, `re/monsters.md` section "The unique's type", S; Coldworm and the Summoner are act bosses that never go through the roll). Tag D for the table, I for the "no Act 1 / 2 impact" reading (the area spawn lists were not intersected).

## Q4. What each class gets (summary of monster-mods-trace.md plus D and S-run)

`constants` rows (D, by id; triples are Normal / NM / Hell): 0 champion chance 20; 1..3 minion life +100 / 75 / 50 %; 4..6 champion +200 / 150 / 100 %; 7..9 unique +300 / 200 / 100 %; 10 champion to-hit 75, 11 champion damage 100; 12 minion to-hit 50, 13 unique to-hit 100; 14 minion strong damage 75, 15 unique strong damage 150; 16..21 minion elemental min / max (0 33 33, 0 50 50); 22..27 champion's (33 33 33, 50 50 50); 28..33 unique's (66 66 66, 100 100 100). Damage and to-hit bonuses are scaled by DifficultyLevels Unique / ChampionDamageBonus 90 / 75 / 66 % (S-run confirms Normal: Strong unique `dmg% 135, th% 90`, Strong minion `67 / 45`, champion `90 / 67`).

| | life | level | experience | damage | to-hit | speed | other (Normal, S-run) |
|---|---|---|---|---|---|---|---|
| minion | x2 / 1.75 / 1.5 | +3 | x5 | strong +67 / 56 / 49 % | strong +45 / 37 / 33 % | fast: yes | enchanted elements 0 in Normal (33 / 50 NM and Hell); no resist bonus; the hooks (explosion, bolts, teleport, curse) need flag 8 and do not run |
| champion (any kind, followers are plain 16) | x3 / 2.5 / 2 | +2 | x3 | +90 / 75 / 66 % | +67 / 56 / 49 % | +20 % | regeneration 0 |
| berserker (39) | x0.75 / 0.625 / 0.5 | +3 (exp x5: the fixed mods run for it) | x5 | +270 / 225 / 198 % | same +270 | none | the champion +90 does not apply (S-run line: `dmg% 270 th% 270`, life 7 -> 5.25) |
| ghostly (36) | x3 | +2 | x3 | +90 and a cold hit 1-2 at level 5 (S-run) | +67 | **none** (game.exe) | physical damage resist 80 |
| fanatic (37) | x3 | +2 | x3 | +90 | +67 | fast formula (S-run: +100) | item armour % -70 (S-run `armor% -70`) |
| possessed (38) | x6 (I: +100 % on the x3 life, no S-run line) | +2 | x3 | +90 | +67 | +20 % | cannot be cursed (flag 0x20, S) |
| unique | x4 / 3 / 2 | +3 | x5 | strong +135 / 112 / 99 % | strong +90 / 75 / 66 % | fast +10..+100 (formula 2048 / Velocity - 128) | regeneration 0; Magic Resistant +40 cold / fire / lightning, the three Enchanted +75 own element, Mana Burn +20 magic, Stone Skin defence x2 and +50 physical, Spectral +20 on the three elements (each only below 100, none past two immunities) |
| superunique | same as unique | +3 | x5 | same | same | same | mods fixed by the record; minions as above |
| act boss (Blood Raven, Smith, Andariel) | x1 | x1 | x1 | x1 | x1 | x1 | flag 8 only: gold name, TreasureClass3 |

Tags: D for the constants, V read for the inits (monster-mods-trace.md), S-run for the Normal values in the table (Cold Plains to Dark Wood, area ids 3, 4, 5, 8, 17). Nightmare and Hell numbers are D arithmetic on the same tables (the two-row run did not cover them: U for the Hell level source, which is the area level except for boss-flagged MonStats rows).

## Q5. Pack sizes and minion counts

| Event | Count | Source |
|---|---|---|
| champion pack | leader plus `seed % 3 + 1` more = 2..4 units, each follower made at radius 4 by FUN_005b2f70 and given mod 16 by FUN_005a48c0; the leader holds the kind | V read (`0054e1e0`..`0054e25b`: LCG step, `x - 3 * (x / 3) + 1`, `push 4` radius, `push 0x10` mod); S-run: six packs 3, 3, 3, 4, 4, 4, one non-16 kind per three of them |
| random unique | `rand(4) + 3` minions = 3..6 of its MonStats `minion1` (else its own type) at radius 3 | S (`re/monsters.md`), S-run: 44 uniques, minions 3 to 6 (mean 4.5) |
| superunique | `rand(MaxGrp - MinGrp + 1) + MinGrp` plus `difficulty` when both are set, of `minion1` (else its own type) | S, S-run (Normal: below) |
| act boss | none of its own (Blood Raven: 0 in 12 seeds; her archers are her room's normal groups) | S-run |

## Q6. Act 1 and 2 superuniques

Act 1, Normal (S-run: `umod_spawn.py` on game.exe's own room population, 12 seeds each, minions counted by `um_summ.py`; monster rows by hcIdx from the same patch tables, D):

| superunique | area | base (hcIdx) | mods (+22) | level (base -> +3) | experience | life x | minions (type: count) | notes |
|---|---|---|---|---|---|---|---|---|
| Bishibosh | Cold Plains | fallenshaman1 (58) | 8, 9 | 2 -> 5 | 32 -> 160 | x4 | fallen1: 2 | S-run: fire 140 / light 40 / cold 40 (Fire Enchanted +75 and Magic Resistant +40 over the row's base; the 140 is above 100, so immune), fire damage 2-4; minions 0 / 0 / 0 resists (the minion init needs flag 1) |
| Rakanishu | Stony Field | fallen2 (20) | 17, 6 | 5 -> 8 | 42 -> 210 | x4 | fallen2: 8 | S-run: speed +100, lightning damage 3-6, lightning resist 75; minions Extra Fast, no bolts |
| Treehead WoodFist | Dark Wood | brute2 (24) | 13, 5, 6 | 5 -> 8 | 84 -> 420 | x4 | brute2: 2 | S-run: damage +135, to-hit +90, speed +100; minions damage +67 / to-hit +45 / speed +100; **mod 13 (rage) is in the record**, has no init and no hook (monster-mods-trace.md: equal) |
| Corpsefire | Den of Evil | zombie1 (5) | 27 | 1 -> 4 | 33 -> 165 | x4 | zombie1: 5 | S-run: spectral hit resists 20 / 20 / 20 over the base (poison 50 base); minions 0 / 0 / 0 |
| Blood Raven | Burial Grounds | bloodraven (267), an act boss | 12, 22 | 10 -> 10 | 181 | x1 | none | S-run: all stats unchanged by mods |
| Griswold, the Countess, the Smith | Tristram, Tower Cellar, Barracks | | | | | | | **not in the run log** (U): Griswold superunique with his own AI; the Countess hcIdx 6 with stat 0x76, quest 5, AI 0xd (S, `re/superuniques.md`); the Smith is an act boss (mod 22, `re/superuniques.md`, S); levels: Smith row Level 10 (D, MonStats), the others U |

Act 2 (U, not extracted this session): Radament (superunique 10, class 229) and Fangskin (superunique 12) are named in the bot data (internal/boss/act2.go, docs/act2.md; V from the bot's own table read of 2026-10-28). The rest of the Act 2 superuniques (Beetleburst, Creeping Feature, Coldworm the Burrower, Fire Eye, Dark Elder and the others) are not in any document of this repo; their mods, minions and levels need `um_extract.sh superuniques` (reads the table through mpq-cat) on the worker host and then the run in the drill section.

Observation, unresolved (U, dead end): one S-run line outside the superunique list, seed 7 Cold Plains, a random unique (flags `9`, not `b`) of class 28 (brute1) holds mods `13.5` and five minions with the same mods. A random unique in Normal should hold one mod (Q2), and 13 has no `upick`, so the pick cannot produce it; the same pair `13.5` is the front of Treehead's record (`13.5.6`). Candidate readings: a Treehead variant made with a wrong flag byte by the harness, or a base the row's SuperUniques entry replaces. Not read in the disassembly; a drill (G6) or one more spawn log resolves it.

## Q7. Experience

Kill experience is the monster's stat 13 (mods applied at spawn: x5 for unique, superunique and minion by leveladd FUN_005a0e40, `lea eax, [eax + eax*4]`; champions net x3 by FUN_005a0e80; berserker x5 because its fixed mods run but FUN_005a0e80 does not) fed to the award path of [xp-award-trace.md](xp-award-trace.md) (level-difference table, ExpRatio, no party bonus alone). S-run values, Normal: Bishibosh 32 -> 160, Rakanishu 42 -> 210, Treehead 84 -> 420, Corpsefire 33 -> 165, a champion at level 7 54 -> 162, a minion the same x5 as its unique, Blood Raven 181 -> 181. d2d at the merged state gives minions the same x5 (pr-mods-proof.md D2). Tag S-run + V read.

## Q8. Drops (links)

[treasure-trace.md](treasure-trace.md) section 5 (D, V run in that card): champions Act 1 Champ A (100 %, 3.0 items), uniques Unique A (5.0 items, the minions use their own plain class), superuniques of Acts 1 and 2 `Act N Super A / B / C` (6 drops, 4 potions, 1.87 gear), Griswold and the Smith 3 drops, the Countess her own class (5.75 items, 2.28 runes), Radament `pk5`. The class a boss gets follows the monster flag: FUN_005a6600 tests flag 4 (champion) for TC2, flag 8 for TC3 (`re/superuniques.md`, S); a minion carries only 0x10 and so keeps the plain class. The treasure-trace numbers hold for all of this; nothing new is claimed here.

## Q9. What the mods change in the fight (S, structural V)

Unchanged from monster-mods-trace.md ("In the fight"): Cursed (7, FUN_005a2530, 3 in 4 on its seed, Amplify Damage level `stat12 / 5 + 1`, melee only in d2d), Fire Enchanted death blast (9), Lightning Enchanted bolts on being hit (17, eight `lightunique`, at most every 10 frames), Cold Enchanted death nova (18), Spectral Hit (27), Multiple Shots (29, ranged only), Teleportation AI (FUN_005b11f0: AI flag 0x20, 40 %, then low life or a shooter nearby, 15 %). The handlers need flag 8, so a minion never explodes, bolts, curses or teleports (S + structural V: the same flag test as the stat inits). Extra Fast / Strong reach the minions (S-run).

## Q10. Upgrade run: Normal, Nightmare, Hell and Act 2 (V run, game.exe's own code)

Command: `tools/emu/run-on-worker.sh --ttl 3000 um_all.sh` (24 runs of `umod_spawn.py` seeds 1..8 over Levels ids 3, 4, 5, 8, 17 (Act 1: Cold Plains, Stony Field, Dark Wood, Den of Evil, Burial Grounds) and 41, 43, 44 (Act 2: Rocky Waste, Far Oasis, Lost City) at difficulty 0, 1, 2; 245 s), then `um_agg.py` over the logs. Method: game.exe populates the rooms; every stat write by FUN_00627260 / FUN_006272b0 is hooked; units with flag 4, 8 or 0x10 are printed. About 3,700 unit lines; no d2d side (so no mismatch count: this is a V run of the game, not a differential). The Normal Act 1 lines equal the first session's S-run lines (Bishibosh `8.9.22` fire / light / cold 140 / 40 / 40, Rakanishu `17.6.22` speed +100, Blood Raven `12.22` unchanged).

| fact | Normal | Nightmare | Hell | tag |
|---|---|---|---|---|
| random unique: mods in its list (counting the rolled ones only) | 1 (54 of 65 lines; 3 more carry a fixed mod, see below; 8 are Blood Raven) | 2 (150) or 3 (20) | 3 (223) or 4 (26), 2 (8 = Blood Raven) | V run for the counts; I that the 3rd (NM) and 4th (Hell) entries are mostly the fixed mod 20 of scarab classes (only the Normal 20.x lines were read one by one) |
| unique level add / experience ratio | +3 / x5 (57 of 57) | +3 / x5 (162) | +3 / x5 (249) | V run |
| champion level add / experience ratio | +2 / x3 (31); +3 / x5 for berserker (4) | +2 / x3 (95); berserker +3 / x5 (11) | +2 / x3 (152); berserker +3 / x5 (16) | V run |
| minion level add / experience ratio | +3 / x5 (425) | +3 / x5 (968) | +3 / x5 (1353) | V run |
| Act 1 superuniques: add / ratio | +3 / x5 (48 lines over three difficulties, every one) | same | same | V run |
| Bishibosh (class 58) level, experience, max life | level 2 -> 5, exp 32 -> 160, hp 20..32 | level 36 -> 39, exp 3852 -> 19260, hp 870..1299 | level 68 -> 71, exp 42138 -> 210690, hp 4096..5534 | V run (the Level(N) / Level(H) column of the row drives the growth, mods only add +3) |
| champion kind mixes (leader's `kind` mod) | 16 plain 25, berserker 39: 4, fanatic 37: 4, possessed 38: 2 | 16: 76, 39: 10, 37: 9, 38: 5, 36: 4 | 16: 115, 39: 14, 37: 10, 36: 9, 38: 5 (+ `20.16`, `13.16` rare) | V run |
| Blood Raven (act boss) | +0 / x1 | +0 / x1 | +0 / x1 | V run (8 of 8 per difficulty in Burial Grounds) |

Act 2 superuniques seen (Normal, Far Oasis, area 43, 8 of 8 seeds): **Beetleburst** = class 92 (scarab2), mods `20.8.22`, level 16 -> 19, exp 256 -> 1280, life 188..240 at 4x the row (hpcreate 47..60), resists fire / light / cold 40 / 40 / 40 (Magic Resistant 8 applies +40 to each; no champion bonuses): V run. Mod 20 (MonUMod "scarab", id 20, xfer 0, no pick weight, D) is a leader-marker with no stat init; it is also the front of every random unique of MonStats classes scarab1 / scarab2 (SetBoss 1, BossXfer 1, PartyMin 2..3, PartyMax 4; D, the MonStats table via `mpq-table.sh monstats`). In Normal the class 91 (scarab1) random uniques show `20.17` (Lightning Enchanted) and `20.9` (Fire Enchanted): mod 20 is the extra entry in the list, so "one rolled mod" still holds. The other Act 2 superuniques (Fangskin, Creeping Feature, Bloodwitch, Radament, Coldworm, the Summoner) live in areas outside the eight runs: not run (U).

Dead ends resolved or narrowed by this run: (a) the earlier `13.5` unique of class 28 (brute1) in Cold Plains seed 7 appears once in the 8 Normal seeds and 0 times in the 8 NM / Hell seeds of any area; the pick still cannot produce 13 (no `upick`), the reading "a Treehead-style fixed list on a flag-9 unit" stays U (G6). (b) The "unique mod count" of 11 two-mod lines in Normal is 8 Blood Raven (`12.22`, flag 9, +0) plus the 3 above: the one-rolled-mod rule is confirmed for 54 of 54 random uniques with no fixed mod.

What this run does NOT show: minion counts per boss in NM / Hell (the log prints every unit but `um_agg.py` does not group them by boss; the unit totals of 425 / 968 / 1353 `0x10` lines include the champion followers, so no per-boss count is claimed here), Act 2 areas outside 41 / 43 / 44, and any monster behaviour (AI, damage in a hit): those need the drill recipe. No d2d value was compared in this session.

## Deviations d2d vs game.exe

The proven ones are in monster-mods-trace.md and pr-mods-proof.md (D1 life / defence / experience level, D2 minion level and experience, D3 ghostly speed, B berserker, merged upstream as #157; E1, E2, H1, S1 classified, not claimed). New here, all read-only:

| id | game.exe | d2d | tag | differential |
|---|---|---|---|---|
| D1 | Extra Fast only on a monster with the walk mode, Teleportation only with A1 (FUN_005a03e0 fPick 3 / 1) | `can_take` (rules/uniques.hpp) tests neither; its comment says so (ponytail) | V read (game), V-d2d read (d2d), D (the rows in Q3) | Buildable: run FUN_005a03e0 under unicorn for every (MonStats row, mod 1..42, expansion flag) with the tables game.exe loaded, compare with `unique_detail::can_take` on the same rows (the umod_stats harness style, `tools/emu/umod_stats.py`). Impact in Acts 1 / 2: none seen (Q3). Not built |
| D2 | MonType exclusion by type inheritance (bit matrix, FUN_005a0070) | the exact type string only (comment in `can_take`) | V read, I for the d2d effect | same differential as D1; only row 17 / `sandleaper` has an exclude, so the effect is whether any MonType derives from sandleaper (U) |
| D3 | multishot refused by a MonStats flag pair (`+0xc`, `+0xe` bytes against a mask) | `type.melee` | V read (game), S (what the flags mean), U | same differential; the d2d column choice is the open item |
| D4 | a random unique may hold up to 9 mod bytes; the count cap is "total at most 9", start refused at 8 | `min(count, 8)` | V read | none needed (a random unique rolls at most 3) |

**PR candidate: none.** The existing differential (`umod_stats.py` + harness, `boss_level.py`) covers stats, not the pick rules, and its D / B / G deviations are already merged upstream. The new D1..D4 have no run (mismatches before = 0 cases counted), so none may be offered; D1 / D2 would need the new differential above and show impact in Acts 1 / 2 first.

## Bot findings (described, not changed)

| id | file:line | what the bot does | the rule | tag |
|---|---|---|---|---|
| B1 | `internal/combat/target.go:134-138` (`championGroupMin = 3`, comment "groups of 3-5") | assumes champion packs of 3 to 5 | 2..4 (leader plus `rand(3) + 1`); six packs in the run: 3, 3, 3, 4, 4, 4; a pack of 2 never reaches the min, a "5" does not exist | V read, S-run |
| B2 | `internal/resource/potion.go:47-54` (`championBonus`), same file's comment "3x life and +50 % damage" | one number for every champion | life x3 holds; damage is +90 % in Normal, +270 % for a berserker (life x0.75), +67 % AR, and the kinds are not told apart; kinds are visible in the enchant list (36..39) which `affixByMod` drops | D, S-run |
| B3 | `internal/combat/target.go:36-39` (`affixByMod`), `affixesOf` | knows 13 of the 30 pick mods; no 13 rage, 22, 23 poisonhit, 12, the champion kinds 36 ghostly (physical resist 80), 37 fanatic (armour -70, fast), 38 possessed, 39 berserker | a Normal Act 1 / 2 pack can be a berserker (Normal +270 % damage on a 5 hp monster) or fanatic (fast); the bot sees an unmodified champion | V read, S-run |
| B4 | `internal/combat/affix.go:21-52` (`AffixResists`) used by `resistsOf` (target.go) | adds Magic Resistant +40, Enchanted +75, Mana Burn +20 to any monster carrying the id, minions included | the resist inits need the boss flag; a minion gets no resist (S-run: Bishibosh's two minions read 0 / 0 / 0 fire / light / cold, Corpsefire's likewise); fallback only (the live stat list wins) and only if minions carry the ids on the client (T1) | S-run, U for the client copy |
| B5 | `internal/combat/target.go:66-67, 219-221` (`killFirstAffixes = Cursed \| ManaBurn \| MultiShot \| Aura`) | a minion carrying one of these ids is killed first | the hooks (curse, aura bonus beyond the party, multishot) belong to flag 8 units; Mana Burn damage on a minion is 0 in Normal; so a minion with the id is not the threat | S (flag 8 handlers), I |
| B6 | `internal/combat/affix.go:79-127` (`TargetDanger`, `PackDanger`, minion test at 123) | pack danger from the boss's mods; minions count only outside Normal (`diff != Normal && minions >= 3`) | in Normal a Strong or Fast unique brings 3..6 minions that carry Strong / Fast too (Treehead's two, Rakanishu's eight are Extra Fast, S-run): the rating is by the boss only | S-run |
| B7 | `internal/combat/target.go:194-196` and `internal/d2/monflags.go:28-42` (`MonKindMinion` after the boss test) | `Boss` is "unique with a name": every random unique is a Boss for `TargetDanger` / `Avoid` | consistent with the flag byte (random uniques have names); but the superunique bit (0x02, Bishibosh reads 0x0a) is not used, so Act 1 / 2 superuniques (fixed mods, `difficulty` extra) are not told from random uniques; no effect on danger today | I |

B1, B3..B6 need no live check beyond the drill T1 (client enchant bytes).

## Drill recipe (lab binary; docs/forkdbg.md, docs/forkcheats.md; nothing run here)

Use one hidden slot, a throwaway character on a field level, the lab binary of docs/d2d-fork.md, `debug spawn <monster> [n] [champion|unique [mods]]` (forkcheats) and `debug mon set` / `monsters` to read life, level and mods. Each drill confirms one U.

| id | confirms | steps and expected |
|---|---|---|
| G1 | the Act 2 superunique records | on the worker host: `tools/emu/run-on-worker.sh mpq-table.sh superuniques` (writes the table to stdout), then read the `Mod1..3`, `MinGrp`, `MaxGrp`, `Class` of the Act 2 rows; extend Q6 |
| G2 | D1 (fPick 1 / 3) | `spawn <a row from um_fpick.py> 1 unique 6` then `unique 26`: d2d accepts it (game.exe would not roll it); the row is a nest, egg, turret or trap, so the visible effect is nil in Acts 1 / 2: the drill only proves d2d's side |
| G3 | pick weights | spawn 10,000 random uniques in a field level by `debug goto` loops and tally mods: Normal expect no 8, the other twelve at about 8.3 % each, one mod each; Nightmare 2 mods, 13 rows |
| G4 | champion pack size | count packs in the same sweep: expect 2..4, followers mod 16 only, the leader one of 16 / 36..39 at 1 / 5 |
| G5 | B4 (minion enchant copy) | spawn Bishibosh's pack on the original (not available) or a d2d unique with mods 8 9: read a minion's enchant list; expect the same bytes as the boss, resist stats 0 |
| G6 | the `13.5` unique | spawn `brute1` unique with mods 13 5 and read the flag byte: `0x9` vs `0xb`; compare with the S-run seed 7 line (Cold Plains) |

## Matrix and ledger updates (proposed)

- `docs/research/monster-mods-trace.md` row "Bot B3 (PackDanger)": add B6 above (Normal minions carry Strong / Fast).
- `docs/difficulty.md` section 4 (not edited here): the pack rule "champions in groups of 2-4 (all but one are Champion ...)" is half wrong: the leader holds the kind, the followers are plain champions, and the size is 2..4 (V read FUN_0054e1e0); "Unique gets 1 extra mod in NM, 2 in Hell" reads as 2 / 3 mods total (Normal 1): V read FUN_005a0760.
- `docs/coverage/ledger.tsv`: **no new row**. The Q10 run (about 3,700 unit lines) is game.exe against itself, with no d2d oracle compared, so it cannot be a TESTED-* row; the existing line `formulas F09 TESTED-DEVIATION docs/research/monster-mods-trace.md 157618` stays the evidence. `docs/coverage/rule-coverage.md` was not regenerated.

## Dead ends

- First session: the runner refused remote scripts, so nothing was run; the upgrade session (Q10) ran the Normal / Nightmare / Hell populations and the Far Oasis superunique, but still not Q3's D1..D3 differential against d2d nor the Act 2 superuniques outside areas 41 / 43 / 44; the tools added here (`um_extract.sh`, `um_tab.py`, `um_fpick.py`, `um_summ.py`) are the offline half and were run only locally on copies of tables the DECOMPMODS and CHILL cards had already extracted.
- The `13.5` random unique (Q6) is unexplained.
- FUN_0046c140's mask table and the `0x6ce26c` flag in the Multiple Shots branch were taken from upstream's doc and a read of the branch, not decoded further.

## Reproduce

```sh
# local, on copies of the tables (monumod.txt, monstats.txt, monstats2.txt from the patch layer)
python3 -I tools/emu/um_tab.py weights monumod.txt
python3 -I tools/emu/um_tab.py rows monstats.txt 58 20 24 5 267 402
python3 -I tools/emu/um_fpick.py monstats.txt monstats2.txt
python3 -I tools/emu/um_summ.py spawn-normal.txt     # the DECOMPMODS umod_spawn.py log (526 lines)
# worker host, the Q10 run (writes $D2_SCRATCH/um_d<diff>_a<area>.txt, then aggregates them)
tools/emu/run-on-worker.sh --ttl 3000 um_all.sh
tools/emu/run-on-worker.sh um_agg.py
```

## Shared-file lines for the merger

See `openspec/changes/uniqmods/merge-lines.txt`.
