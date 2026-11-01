# Champion, unique and minion modifiers: game.exe 1.14d vs d2d vs our bot (DECOMPMODS, decomp-queue item 13)

> status: partly superseded (2026-11-01, card buildmods): D1 is narrower than written here: game.exe reads a boss's AR and damage per hit at the RAISED level (FUN_005a4f50, stat 12), so only life, defence and experience come from the base level; d2d's AR and damage were right. Built and proven: [pr-mods-proof.md](pr-mods-proof.md).

> status: current (2026-10-31, card decompmods). Offline only: no game instance, no d2d slot; nothing pushed, posted or filed. d2d read at upstream main 72132b99 (worktree `trace/mods` under the PR checkout dir on the worker host). Tags: V = read in game.exe 1.14d (md5 2b3600c1..., `tools/emu/disasm.py`) or run in its own code under unicorn by this agent; V-d2d = d2d source read at file:line; D = 1.14d tables as game.exe's own loader holds them (drlg.boot) or `monstats.table`; S = upstream doc (`docs/research/re/monsters.md`); I = inference; U = unverified. Oracles: `tools/emu/umod_spawn.py` (real room population, every stat write), `tools/emu/umod_stats.py` (the real MonUMod inits on synthetic units) + `tools/emu/umod_stats_harness.cpp` (d2d's real `rules::boss_stats` from `components/rules/uniques.hpp`, plus `make_boss`'s application copied line for line).

## Answer in five lines

1. **The modifier inits themselves match d2d for champions, uniques and superuniques** (157,618 cases, exhaustive singles over every MonStats row x 3 difficulties + 100,000 random combos): life %, damage / to-hit %, Extra Fast, the resistance adds, Fire / Lightning Enchanted damage, Mana Burn, Stone Skin, fanatic, possessed are equal, with five narrow exceptions (table below).
2. **D1 (big): d2d reads a champion's / unique's / superunique's MonLvl stats at the boosted level; game.exe reads them at the base level** and only raises stat 12 (+3, champion net +2) and multiplies the base experience (x5, champion x3). V: in the real population of six Act 1 levels x 12 seeds, all 113 random champions / uniques / superuniques had their life / defence / AR / damage rows read at the MonStats level (`0x573f7e`), the +3 came later (`FUN_005a0e40`). In Normal Act 1 this makes d2d's bosses about 1.9x the life, 2.5x the defence, 2.9x the AR, 1.3x the damage and 1.75x the experience of game.exe's (Bishibosh: level 2 rows vs level 5).
3. **D2 (big): d2d's minions get none of what game.exe gives them**: level +3 and experience x5 (all 401 traced minions), and their unique's mods run on them with the minion flag (Extra Strong +67 % damage / +45 % AR, Extra Fast the full speed bonus, Nightmare / Hell enchanted damage and Mana Burn on the minion rows). d2d tags minions with no mods (`objgroups.cpp:648/668/777`).
4. **Bot**: the early-Normal danger model is right to rate Extra Strong / Fast / Teleport / Mana Burn as 1 and Fire Enchanted as 1, but misses that a Strong or Fast unique's minions carry it (game.exe), that Lightning Enchanted answers every hit of a Charged Bolt volley (Normal: no extra danger), and its champion rule of thumb says +50 % damage where game.exe gives +90 % (Normal). The resistance fallback (`AffixResists`) uses a different immunity rule.
5. **PR candidate**: D1 + D2 + the ghostly speed (G1) are clean, proven rule-level deviations; not built in this time box (no patched d2d build, no det run): row in `docs/research/pr-bundle-queue.md`.

## game.exe's rules (with addresses)

### When the mods are applied (V)

| Step | Rule | Tag |
|---|---|---|
| unit made | `FUN_00573cb0`: level = MonStats Level (Normal; expansion NM / Hell: the area level unless a boss-flagged row), stat 12 set, then life (rolled, `<< 8`), defence, AR, damage from `FUN_006538a0(row, .., level)` at THAT level (`0x573f7e`: stack +0xc) | V (disasm; population trace) |
| random unique / champion | `FUN_005a43e0` -> `FUN_005a09e0` (make) -> `FUN_005a0760` (champion 20 % or unique, mods) -> `FUN_005a2120(game, unit, 1)` | V |
| `FUN_005a2120` | (1) minions first (`FUN_005a0c00`, `rand(4) + 3` of minion1, else its own type, flag 0x10); (2) the fixed mods 1..4 at `0x6e2168` (rndname, hpmultiply, light, leveladd), each `init(unit, mod, 1)` from the table at `0x73c008`, then `FUN_0058f380(unit, mod, 0, FUN_005a2100)`: the same init on every party member (its minions) with flag 0; (3) its own 9 mod bytes (`data +0x1c`) the same way | V |
| superunique | `FUN_005a49b0` -> `FUN_005a2120` (S; d2d's `superunique_mods` for the mod list) | S |
| act boss | `FUN_005b1cf0`: `FUN_005a4850(mod, 1)` per mod, no fixed mods: no level, life or experience change (12 traced units: Blood Raven et al., all unchanged) | V |
| champion pack | `FUN_005a48c0`: flags \|= 5, mod 16, `FUN_005a2120` (fixed mods too) | V |

Init table (`0x73c008`, D from game.exe memory): 1 `5a0ce0`, 2 `5a0dc0`, 4 `5a0e40`, 5 `5a17e0`, 6 `5a1910`, 8 / 27 / 28 `5a1370`, 9 `5a1990`, 16 `5a0e80`, 17 `5a1b00`, 18 `5a1c70`, 23 `5a1e00`, 25 `5a1f90`, 26 `5a1600`, 30 `5a1650`, 36 `5a1080`, 37 `5a11f0`, 38 `5a1230`, 39 `5a1280`, 41 `5a1330`. No init: 3 light, 7 cursed, 29 multishot (fight hooks only).

### What each init writes (V: disasm read and run under unicorn, `umod_stats.py`)

| Mod | Function | Boss flag 1 (champion / unique) | Flag 0 (minion) |
|---|---|---|---|
| 2 hpmultiply | `5a0dc0` -> `5a0d20` | life (256ths) += MulDiv(life, row 4+d (champion flag 4) else 7+d, 100); **stat 0x4a (life regen) set to 0** | life += row 1+d; regen kept |
| 4 leveladd | `5a0e40` | stat 12 += 3, stat 13 (exp) x5 | **same: +3, x5** (no flag test) |
| 16 champion | `5a0e80` | stat 12 -1; exp = exp - 2 (exp / 5); +dmg% (0x19) / +AR% (0x77) = rows 11 / 10 x DifficultyLevels +0x34 / 100, damage part halved for BaseId 0x76; velocity% (0x43): ghostly (36) **none**, fanatic as fast, else +20; none when MonStats Velocity <= 0 | nothing |
| 36 ghostly | `5a1080` | flag 0x40, DR (0x24) = 80, `5a0e80`, cold min / max += MonLvl L-DM at stat 12 x rows 22 / 25 + d, cold length += 150 | nothing |
| 37 fanatic | `5a11f0` | item_armor_percent (0x10) = -70, `5a0e80` | nothing |
| 38 possessed | `5a1230` | flag 0x20, life +100 %, `5a0e80` | nothing |
| 39 berserker | `5a1280` | life -75 %, dmg% += 300 x bonus / 100 (halved for BaseId 0x76), AR% += 300 x bonus / 100; no `5a0e80` | nothing |
| 5 strong | `5a17e0` | rows 15 / 13 x bonus / 100 (dmg halved for BaseId 0x76) | **rows 14 / 12 x bonus** |
| 6 fast | `5a1910` | velocity% += clamp(2048 / Velocity - 128, 10, 100) | **same** (no flag test) |
| 9 / 17 / 18 enchanted | `5a1990` / `5a1b00` / `5a1c70` | element min / max += MonLvl L-DM (stat 12 at that moment) x rows 28 / 31 + d; cold also length += stat12 x 5 + 100; then `5a1370` | **rows 16 / 19 + d** (Normal 0, NM 33 / 50, Hell 33 / 50 %); no resist |
| 25 mana burn | `5a1f90` | manadrain min / max (0x3e / 0x3f) = (L-DM x rows 28 / 31) `<< 8`; `5a1370` | rows 16 / 19 (Normal 0) |
| 23 poison hit | `5a1e00` | S (upstream doc), not differenced | |
| 8 / 27 / 28 resist, spectral, stone skin | `5a1370` | needs flag; stone skin doubles defence (0x1f) FIRST; then counts immunities (fire, light, cold, poison, DR, MR >= 100): >= 2 -> nothing more; MR +40 cold, fire, light (each < 100, recount between); FE / CE / LE / poison +75; mana burn MR +20; spectral +20 cold, fire, light (each < 75); stone skin DR + 50 | nothing |
| 26 teleport | `5a1600` | needs flag: skill 184 level 1, AI flag 0x20 | nothing |
| 30 aura | `5a1650` | aura skill (S: table `0x73bf68`; d2d `boss_aura`) | not run here |

D (game.exe's tables): MonUMod constants 20 / 100 75 50 / 200 150 100 / 300 200 100 / 75 100 / 50 100 / 75 150 / minion elem 0 33 33, 0 50 50 / champion 33 33 33, 50 50 50 / unique 66 66 66, 100 100 100; DifficultyLevels Unique / ChampionDamageBonus 90 / 75 / 66; upick Normal for mod 8 (Magic Resistant) is 0: a random Normal unique never rolls it; Normal uniques get exactly 1 mod (`rand(1) + 1 + 0`).

### In the fight (S + structural V)

Hook table `0x73c0b8` (D, six events a mod): 7 cursed event 3 `5a2530` (3 in 4 on its seed, Amplify Damage at stat12 / 5 + 1); 9 FE events 1 / 2 `5a25f0` / `5a2620` (death blast); 17 LE events 1 / 2 / 4 `5a37d0` / `5a29a0` / `5a2ba0` (takes damage: 8 `lightunique` bolts, at most every 10 frames); 18 CE events 1 / 2 (death nova); 27 spectral events 0 / 5; 29 multishot event 5 `5a3610`; teleport AI `5b11f0` (V: AI flag 0x20, rand(100) < 40, then life < 30 % or (class not flagged by `457490` and AI param +0x14 < 10), rand(100) < 15, a free spot in its room). Handlers need flag 8 (S): minions do not explode, bolt or teleport.

## d2d's implementation, per mod

| Mod | d2d (V-d2d) | Verdict |
|---|---|---|
| 1 rndname, 2 hpmultiply, 4 leveladd | `boss_stats` (uniques.hpp:250-263): hp %, level_add +3, exp_mult 5 for uniques; minion hp row only | stats: equal on the same base; **D1** base level; **D2** minion level / exp; regen zeroing not modelled (d2d has no monster regen) |
| 16 champion | uniques.hpp:264-271 | equal except BaseId 0x76 halving (H1) |
| 36 ghostly | -33 velocity (uniques.hpp:270), DR 80, cold rows 22 / 25, length 150 | **G1**: game.exe gives no speed change (dead branch at `0x5a1062`) |
| 37 fanatic, 38 possessed | equal | equal |
| 39 berserker | no level_add, exp_mult 1 (uniques.hpp:264) | **B-level**: game.exe +3 level, exp x5 (fixed mods run for it); dmg halving ponytail-noted (H1) |
| 5 strong | equal for bosses | **D2** minions; H1 |
| 6 fast | equal for bosses | **D2** minions |
| 8 / 27 / 28 resists | `resist()` lambda | equal, except stone skin's defence doubling is skipped when the monster already has 2 immunities (S1: 723 + 253 cases, NM / Hell only) |
| 9 / 17 / 18 enchanted damage | one `elem` slot (uniques.hpp:320-326) | equal for one enchant; **E1** a unique with two enchants keeps only the last (NM / Hell); **E2** Cold Enchanted's cold length (stat12 x 5 + 100) missing (d2d elem_len 0) |
| 25 mana burn | ai.cpp:2953 | equal for bosses (whole points vs 256ths: same value); D2 minions NM / Hell |
| 23 poison hit | ai.cpp:2960 | not differenced (act boss / family only) |
| 7 cursed | ai.cpp:2689: melee only (ponytail), 3 in 4, level / 5 + 1 | structural match; the missile-hit case is a known ponytail |
| 26 teleport | ai.cpp:2766: 40 %, then low life or a shooter within 10, 15 %; room as an 8x8 block | structural match (I: the `457490` class test read as "a shooter"); U: room-spot choice |
| 29 multishot | ai.cpp:2025: two more missiles, unique / superunique only | structural match |
| 17 LE bolts, 9 FE / 18 CE deaths | fight/deaths.cpp:89-111 | structural match (earlier traces); LE needs a whole-point drop in d2d (any 256th in game.exe), fires at once in GH |
| 27 spectral hit | ai.cpp:3041 `attack_starts` | structural match (rows 28 / 31) |
| 30 aura | `boss_aura`, `Fight::monster_auras` | not run (aura init hooked out of the oracle); U |
| 24 thief | not built; disabled in MonUMod | equal (never rolled) |
| 13 rage, 11 durieldead, 3 light | none | equal (no init, no hooks) |
| 41 always_run_ai (init `5a1330`), 14 / 15 / 19 / 21 / 31 / 32 / 34 / 35 / 40 / 42 hooks | family_mods lists some | not traced (none in Act 1) |

## Deviations (real d2d code: `rules::boss_stats` compiled from upstream 72132b99; game.exe's own inits)

Oracle run: 157,618 cases (all MonStats rows with a Normal level, 3 difficulties, kinds c / u / m with each single mod + 100,000 random u / s / m combos), seed 1; game.tsv md5 9844cc49..., d2d.tsv 39e25809.... 76,827 cases differ in at least one field (c 3,372 of 8,730; u 6,190 of 57,707; s 9,480 of 33,396; m 57,785 of 57,785).

| id | deviation | evidence | cases (same-base differential) | early Act 1 Normal |
|---|---|---|---|---|
| D1 | boss stats from MonLvl at level + 3 / + 2 in d2d (`monster_stats(.., boss.level_add)`, ai.cpp:2927); game.exe at the base level | population: 113 of 113 bosses read rows at the base level; `FUN_005a0e40` only writes stats 12 / 13 | structural (not in the same-base diff by design); every random champion / unique / superunique | yes: life 1.9x, AR 2.9x, defence 2.5x, damage 1.3x, exp 1.75x at L2 -> 5 (MonLvl L-HP 9 / 17, L-TH 12 / 35, L-AC 12 / 30, L-DM 3 / 4, L-XP 40 / 70) |
| D2 | minions: no +3 level, no x5 exp, none of the boss's mods' minion effects | population: 401 of 401 minions +3, x5, life x2; oracle | lvl / exp 57,785 of 57,785; strong dmg / AR 3,303 + 376; fast 2,928 + 376; NM / H enchanted 1,878 / 1,908 / 1,920 + combos; mana 1,867 + 379 | yes: XP from minions x5 in game.exe; Extra Strong / Fast packs |
| B-lvl | berserker champions: d2d level +0 and exp x1; game.exe +3, x5 | oracle | 1,746 of 1,746 berserker cases | yes (LoD rows; a 1-in-5 champion kind) |
| G1 | ghostly champions -33 % velocity in d2d; game.exe none | disasm `0x5a0ff8` returns before `0x5a1062`; oracle | 1,554 (all ghostly with Velocity > 0) | yes: d2d's ghostly packs walk a third slower |
| E2 | Cold Enchanted's cold length missing | oracle `cold_len` | 3,333 + 499 + 2,290 + 4,470 | yes (Normal CE uniques) |
| E1 | two enchants on one unique: d2d keeps the last | oracle | ~2,500 combos | no (Normal uniques have one mod) |
| H1 | BaseId 0x76 family: champion / strong / berserker damage % halved in game.exe | oracle (row 118) | 24 x 5 kinds + 54 + 31 + 10 | no (not an Act 1 family) |
| S1 | stone skin defence x2 skipped by d2d with 2 immunities | oracle | 204 + 49 + 10 + ~25 | no |
| R1 | life regen zeroed for bosses (stat 0x4a) | oracle | n/a: d2d models no monster regen | no impact on d2d |
| A1 | damage % applied to whole-point min / max at spawn in d2d vs to the 256ths roll in game.exe (`0057b420`) | read (both) | not counted (rounding: up to 1 hp less per hit) | small |

## The bot's model (internal/, V-d2d of our own code)

| Where | Assumption | game.exe | Finding |
|---|---|---|---|
| `internal/resource/potion.go:47` championBonus | "champions have about 3x life and +50 % damage" (SME rule of thumb) | Normal: life x3 (+200 %), damage +90 %, AR +67 %, speed +20 %; stats at the base level (D1) | B1: the damage number is low; on d2d champions are also ~1.7-2.9x tougher in AR / life (D1), so drills on d2d overstate champion danger vs game.exe |
| `internal/combat/affix.go:79` TargetDanger | Normal: Strong / Fast / Teleport / Mana Burn / Stone Skin / MR / FE -> 1; aura -> 2; LE nothing | LE bolts on every hit taken (10-frame limit): a Charged Bolt volley into an LE unique costs up to 8 x 5 lightning a trigger | B2: LE should count at least as FE (1) for a low-level Sorceress (Soaky's LE Carver death) |
| `internal/combat/affix.go:113` PackDanger | minions raise danger only in NM / Hell | the boss's Strong / Fast reach its minions in every difficulty | B3: rate the whole pack, not only the unique, when the unique has Strong or Fast |
| `internal/combat/affix.go:21` AffixResists | skips an add that would make a third immunity; counts magic + 4 elements | skips everything once 2 immunities exist (counting DR and MR too), MR adds stop at 2, enchanted +75 always; spectral +20 / stone skin DR not modelled | B4: fallback only (live stats preferred); NM / Hell edge |
| `internal/combat/target.go:14` Enchants ids | ids = MonUMod rows, read from MonsterData.Enchants | the server keeps them at MonsterData +0x1c (9 bytes, V); the client copy is U | B5: UNVERIFIED live (no unique seen); drill T1 |
| `internal/combat/resource.go:119` weakHostiles | champion / unique +2 / +3 levels | level stat yes, but stats from the base level (D1) | note only |
| archived `docs/archive/monster-attack-trace-2026-10-30.md` row "every unique: row 13 +tohit% 100 x 90 / 100" | every unique +90 % AR | only Extra Strong gives AR % (u () case: 0) | B6: the archived Rakanishu "x 1.9" AR row is wrong (his mods 17, 6: no AR bonus) |

## Drill recipe (fork debug commands by name; observables)

T1 enchant ids: `spawn` a unique with a known mod (e.g. Bishibosh in Cold Plains, MR + FE) on a throwaway character; read the unit's enchant list: expect `8 9` (+ 22 questcomplete on the server copy), and minions `8 9` (copied, MonUMod byte +7). T2 D1 on d2d vs game.exe: same unique, read max life / level / defence: game.exe Bishibosh level 5 with life x4 of a level-2 fallen shaman roll (traced: 5 -> 20, 7 -> 28); d2d today shows level-5 rows (life ~1.9x). T3 minion XP: kill one minion of a random unique: game.exe XP = 5x a normal one of its type at the same level; d2d 1x today. T4 ghostly: `spawn` a ghostly champion pack and time its walk over 20 subtiles: game.exe = a normal champion's minus the +20 %; d2d 33 % slower. T5 Extra Strong / Fast minions: a Strong unique's minion hits ~1.67x a plain one's damage; a Fast unique's minions close in as fast as the unique. T6 CE: get hit by a Cold Enchanted unique: chill lasts (mlvl x 5 + 100) frames in game.exe; d2d no chill length today. T7 LE: one Charged Bolt volley into an LE unique: count `lightunique` missiles (8 per trigger, at most one trigger per 10 frames). Leave the character in town, no panels open.

## Reproduce (the worker host; upstream worktree on main 72132b99; game.exe md5 2b3600c1)

```sh
cd <upstream>/tools/emu && cp <this repo>/tools/emu/umod_*.py <this repo>/tools/emu/umod_stats_harness.cpp .
D2_MPQ_DIR=<mpq dir> <venv>/bin/python umod_spawn.py 1-12 3 0        # also levels 2 4 5 17 8: 526 boss / minion lines
D2_MPQ_DIR=<mpq dir> <venv>/bin/python umod_stats.py 100000 1 > game.tsv   # 157,618 cases, about 4 min
g++ -O2 -std=c++23 $(for d in ../../components/*/; do echo -n "-I$d "; done) -o umod_stats_harness umod_stats_harness.cpp
./umod_stats_harness < game.tsv > d2d.tsv && <venv>/bin/python umod_stats.py --compare game.tsv d2d.tsv
```

Open: no patched d2d build (no after side), no det run (needs the d2d slot), NM / Hell level source (area level) not compared, aura init and poison hit not run, the client's enchant copy (B5), the `457490` class test of teleport, LE's sub-point trigger.
