# Sorceress passives, synergies and +skills: game.exe 1.14d vs d2d vs the bot (decomp row 23, SKPASS)

> status: card skpass, 2026-11-01, offline (no game, no d2d slot, nothing pushed or posted). Decomp row 23 (sorc-skill-matrix rank 10). Answers the user's question of 2026-11-01: how the Sorceress passives work and whether this version has synergies.

Tags: **V read** disassembled by this agent with upstream's `tools/emu/disasm.py` on the patched 1.14d game.exe (addresses given); **V run** game.exe's own code or table load run under the emulator by this agent (`tools/emu/sk_pass_tbl.py`); **D** the patched 1.14d tables as game.exe's MPQ stack serves them (`sk_pass_tbl.py skills|stats|states`); **S** d2d source at upstream main 72132b99 (still upstream's head on 2026-11-01) or an earlier trace (named); **I** inference or arithmetic on V/D (the replay and damage numbers of `tools/emu/sk_pass_plan.py` are all I); **U** not verified.

## Answer in five lines

1. **Yes, 1.14d has synergies** (D): every Sorceress attack skill but Static Field, Nova and Thunder Storm carries an `EDmgSymPerCalc` of other skills' **base points** (`blvl`) x a percent; Frozen / Shiver / Chilling Armor have duration and freeze-length synergies; Ice Blast a freeze-length synergy (table in section 2).
2. **Masteries are passive stats** put on the caster by FUN_00646d60 when the skill is learned and refreshed when its level changes (V read): Lightning Mastery = stat 330 `passive_ltng_mastery` = 50 + 12 (L-1) % **damage**, Fire Mastery = stat 329 = 30 + 7 (L-1) % damage, **Cold Mastery = stat 335 `passive_cold_pierce` = 20 + 5 (L-1) % off the monster's cold resistance, not damage** (D). The damage masteries are added once when the missile is made (S, skill-damage-trace); the pierce is subtracted in the hit's resistance step FUN_0057be00 **only when the monster's resistance is under 100**, floored at -100 (V read): a mastery never breaks an immunity.
3. **Warmth** = stat 27 `manarecoverybonus` = 30 + 12 (L-1) % (D), applied by the regen FUN_005806f0 as `MulDiv(per-frame base, 100 + stat 27, 100)` on a truncated integer base (V read, SKMANA): Warmth 1 adds nothing under 47 max mana and +20-29 % above. d2d implements it the same way (S; V-d2d DRILLS15 M4).
4. **+skills** (FUN_00644180, V read): +all skills (127), skill shrine +2 (state 134 `shrine_skill`), +class (83), +tab (188), oskill (97, at most 3 on an own-class skill), +elemental skills (126, by the row's element byte: Lightning Mastery and Cold Mastery have none) and +single skill (107). They feed the level used for damage, mana, duration and the passives' stats (`lvl`, FUN_006442a0 flag 1); **synergies read base points only** (`blvl`, flag 0). d2d splits them the same way (S). **Our bot reads base + stat 107 only** (B2).
5. **Lightning Mastery's first point is clvl 30** (reqlevel 30, D), and the leveler does NOT save the refused points: it spends them on the next legal row. Because Chain Lightning needs a Lightning point (reqskill, D) and the FRESH table's first Lightning row is clvl 40, the real order is **Lightning 1 at clvl 25 (27 without quest points), Chain Lightning from 26 (28), Lightning Mastery from 30, five by 33 (34)** (I, replay of the leveler's own rule; B1). Every later row is reachable.

## 1. Questions answered (Q1..)

| id | question | answer | where | tag |
|---|---|---|---|---|
| Q1 | How does a passive skill put its effect on the unit? | FUN_00646d60(unit, skill): Skills row; needs `passivestate` (+0x94) > 0; no skill entry (FUN_00643810) or the row's +0x80 state on the unit: the passivestate's stat list is removed (FUN_006256b0 / 006277e0 / 00626cd0). Else level = FUN_006442a0(unit, entry, **1**) (with +skills); list = FUN_00643620(unit, state, lvl); skipped when the list's stat 0x15f already equals the level; else each passivestat k (+0x98, up to the first empty) **set** (FUN_00627150) to passivecalc k (+0xa4, FUN_00646ca0) on the layer `passiveitype` (+0x96, 0 when none); stats 0x15e = skill, 0x15f = level; the state turned on (FUN_00639e30). | 0x646d60..0x646f10 | V read |
| Q2 | When is it applied / refreshed? | The point-spend path calls FUN_00646d60 right after adding the skill (FUN_00647280, call at 0x55bb52). FUN_00646f20(unit) walks the table of skills with a passivestate (tables +0xbb4 / +0xbb0) and refreshes each one the unit has **whose state is already on** (calls at 0x56de44, 0x56de7e, 0x56dee1, 0x56df25, 0x570122, 0x4c7b0f, 0x4d9161: item and stat callbacks, I). So a +skills item changes Lightning Mastery's percent at once. | 0x646f20, callers | V read (callers' roles I) |
| Q3 | Lightning / Fire Mastery: magnitude and use | Lightning Mastery row 63: passivestate 72 `lightningmastery`, passivestat1 330 = `ln12`, Param1 50, Param2 12 -> 50 + 12 (L-1) % (L1 50, L5 98, L10 158, L20 278). Fire Mastery row 61: 329 = 30 + 7 (L-1) %. Read once when the missile is made: FUN_00644d50 / 00644e40 flag 1 add `MulDiv(dmg, stat, 100)` with the element byte table (fire 329, lightning 330, cold 331, poison 332) on the caster; the missile carries no mastery stat. Applies to the min too (the synergy min gate does not cover the mastery): Lightning's min 1 becomes 1.5 at LM 1, 3.8 at LM 20. Static Field does not use it (its damage is a percent of life, FUN_005c96a0, no FUN_00644d50; I from the static-field read). | D (`sk_pass_tbl.py skills`, `record 63 0x94:2 0x98:2`: 72, 330 V run); formula S `skill-damage-trace.md` |
| Q4 | Cold Mastery | Row 65: passivestate `coldmastery`, passivestat1 **335 `passive_cold_pierce`** = 20 + 5 (L-1) (L1 20, L5 40, L20 115). Stat 331 (`passive_cold_mastery`, the cold damage mastery the byte table reads) is granted by no Sorceress skill. The hit's element descriptors at 0x732980 (12 x 0x2c: +4 resistance stat, +8 max-resistance stat, +0xc pierce stat) map fire 39 / 40 / 333, lightning 41 / 42 / 334, cold 43 / 44 / 335 (also the two cold-length rows), poison 45 / 46 / 336, magic 37 / 38 / none, physical 36 / none. | D; table V run (`sk_pass_tbl.py res`) |
| Q5 | Where is a pierce applied, and against immunes? | FUN_0057c1e0 builds the context (+8 attacker, +0xc defender, +0x14 = defender is a monster and FUN_0063ee90(defender)), FUN_0057bf80 -> FUN_0057be00: res = defender's stat; with a pierce stat: **subtracted when res < 100, or when the defender is not such a monster** (players, I for hirelings); defender not a monster: + the difficulty penalty (expansion: DifficultyLevels via FUN_00611d30; classic -20 / -50) and capped at 75 + max-res stat (at most 95); then res <= -100 -> -100. A monster at 100+ keeps its resistance: Cold Mastery never breaks a cold immunity, and the immune hit deals 0 (MulDiv(dmg, 100 - 100, 100)). | 0x57be00..0x57bf05, 0x57bf80, 0x57c1e0 | V read |
| Q6 | Warmth | Row 37: passive, passivestate `warmth`, passivestat1 27 `manarecoverybonus` = `ln12` = 30 + 12 (L-1) % (L20 258 %), EType fire (so +fire skills raise it). Regen FUN_005806f0 a frame: `max(maxmana256 / 3000, 1)`, then `MulDiv(base, 100 + stat 27, 100)`, + stat 26 (SKMANA V read). Per frame in 256ths (I): max 45: 3 -> Warmth 1: 3 (no gain), Warmth 5: 5; max 55: 4 -> 5 (+25 %); max 100: 8 -> 10; max 200: 17 -> 22. Fill time at 55 max: 141 s -> 113 s. | D; V read 0x5806f0 (SKMANA); arithmetic I |
| Q7 | Telekinesis | srvstfunc 12 = FUN_005c9030 (table 0x732140), srvdofunc 21 = **FUN_005c98f0** (table 0x7322b0). Do: players only; target from FUN_00553540; by target type (jump table 0x5c9b34): a **monster** (both units out of town): damage record from the skill (FUN_0056e170 physical, FUN_0056e0c0 element: lightning EMin 1 / EMax 2, +1 a level, HitShift 8), the row's hit flags (+0x12e / +0x130 / +0x134), **flags | 9 (knockback) only when rand(100) (FUN_0045c390 on the caster's seed) < Skills +0x14c = Param2 = 35**, then FUN_0057c6c0 + FUN_0057cee0; an **item** of item type 4, 0x16, 0x26, 0x38, 9 or 0x29 (FUN_00629bb0): picked up (FUN_00563560), another item: FUN_00553380 (I: refused / drop); an **object**: operated (FUN_00584420). Its plan value: the prerequisite of Teleport and the Energy Shield mana ratio `calc2 = 32 - Telekinesis blvl` (D). | 0x5c98f0..0x5c9b2e; +0x14c = 35 V run (`record 43 0x14c`) | V read (item type names U) |
| Q8 | +skills: what adds to a skill's level | FUN_00644180(entry), unit in esi: `entry +0x2c` (source U) + FUN_00644150 (+2 under state 134 `shrine_skill`) + stat 127 `item_allskills`; a player's own-class skill: + stat 83 (param class) + stat 188 (param class x 8 + tab - 1, when the SkillDesc tab byte is set) + **min(stat 97 oskill, 3)**; a player's other-class skill: + stat 97 in full; a non-player: stat 97 capped at 3 when the entry has base points; then + stat 126 `item_elemskill` with the param = the row's element byte +0x1dc (Fire Bolt 1, Charged Bolt / Telekinesis 2, Ice Bolt 4; Lightning Mastery 0, V run) + stat 107 `item_singleskill` (param skill). | 0x644180..0x644295, 0x644150 | V read |
| Q9 | Base level vs level | FUN_006442a0(unit, entry, flag): `entry +0x28` (base points); flag 1 and `entry +0x34 == -1` (the character's own entry, not an item charge): + FUN_00644180; negative -> 0; at most FUN_00611830(0) (table 0x96c8a8[0], value U). Flag 1 feeds the cast level (damage brackets, mana cost SKMANA, `ln12` durations, Charged Bolt's bolt count), and the passives (Q1); the calc operand 41 `blvl` reads flag 0 (skill-damage-trace, FUN_00646460): **all synergies in section 2 use base points**, so +skills never raise a synergy. | 0x6442a0..0x6442fb | V read (+ S skill-damage-trace for operand 41) |
| Q10 | Lightning Mastery reqlevel vs the table's clvl 26 row | Skills reqlevel 30 (D). The leveler's rule refuses it while clvl < 30 (`internal/leveler/leveler.go:272`) and the planner marks it blocked and goes on to later rows (`internal/leveler/plan.go:82-95`): the point is spent elsewhere, not saved. Effective first point clvl 30; all later LM rows reachable (LM 20 at clvl 57 with quest points). See B1 for the knock-on (Lightning, Chain Lightning). | D + S (our code) + I (replay `sk_pass_plan.py`) |

Dead ends written down: 0x14d / 0x14e / 0x14f immediates in the full listing are mostly unrelated (UI and string ids); the pierce is found only through the descriptor table at 0x732980. `FUN_00611830(0)` is filled at run time (0x96c8a8 in .data/.bss); not read.

## 2. Synergies in 1.14d (D, patched Skills.txt; `sk_pass_tbl.py skills`)

All synergy operands are `skill('<name>'.blvl)`: base points (Q9).

| Skill (row) | Synergy column | Fed by | Per point |
|---|---|---|---|
| Fire Bolt (36) | EDmgSymPerCalc | Fire Ball, Meteor | +16 % fire damage |
| Charged Bolt (38) | EDmgSymPerCalc | Lightning | +6 % |
| Ice Bolt (39) | EDmgSymPerCalc | Frost Nova, Ice Blast, Glacial Spike, Blizzard, Frozen Orb | +15 % |
| Frozen Armor (40) | auralencalc | Shiver Armor, Chilling Armor | +250 frames = **+10 s** duration (3000 + 300 (L-1) frames base) |
| Frozen Armor (40) | calc1 (freeze length on a striker) | Shiver Armor, Chilling Armor | +5 % of `ln56` (30 + 3 (L-1) frames) |
| Inferno (41) | EDmgSymPerCalc | Warmth | +13 % |
| Static Field (42) | none | | |
| Telekinesis (43) | none (it feeds Energy Shield calc2 and is Teleport's prerequisite) | | |
| Frost Nova (44) | EDmgSymPerCalc | Blizzard, Frozen Orb | +10 % |
| Ice Blast (45) | EDmgSymPerCalc / ELenSymPerCalc | Ice Bolt, Blizzard, Frozen Orb / Glacial Spike | +8 % / +10 % freeze length |
| Blaze (46) | EDmgSymPerCalc | Warmth, Fire Wall | +4 %, +1 % |
| Fire Ball (47) | EDmgSymPerCalc | Fire Bolt, Meteor | +14 % |
| Nova (48) | none | | |
| Lightning (49) | EDmgSymPerCalc | Charged Bolt, Chain Lightning, Nova | +8 % |
| Shiver Armor (50) | auralencalc / EDmgSymPerCalc | Frozen Armor, Chilling Armor | +10 s / +9 % |
| Fire Wall (51) | EDmgSymPerCalc | Warmth, Inferno | +4 %, +1 % |
| Enchant (52) | EDmgSymPerCalc | Warmth | +9 % |
| Chain Lightning (53) | EDmgSymPerCalc | Charged Bolt, Lightning, Nova | +4 % |
| Teleport (54) | none | | |
| Glacial Spike (55) | EDmgSymPerCalc / auralencalc (freeze length) | Ice Bolt, Ice Blast, Frozen Orb / Blizzard | +5 % / +3 % |
| Meteor (56) | EDmgSymPerCalc | Fire Bolt, Fire Ball | +5 % |
| Thunder Storm (57) | none | | |
| Energy Shield (58) | calc2 (mana per absorbed point) | Telekinesis | `32 - TK blvl` |
| Blizzard (59) | EDmgSymPerCalc | Ice Bolt, Ice Blast, Glacial Spike | +5 % |
| Chilling Armor (60) | auralencalc / EDmgSymPerCalc | Frozen Armor, Shiver Armor | +10 s / +7 % |
| Hydra (62) | EDmgSymPerCalc | Fire Bolt, Fire Ball | +3 % |
| Lightning Mastery (63), Cold Mastery (65), Fire Mastery (61), Warmth (37) | none (they are the passive stats of Q3, Q4, Q6) | | |
| Frozen Orb (64) | EDmgSymPerCalc | Ice Bolt | +2 % |

The synergy is applied as `dmg += MulDiv(dmg, sum x par, 100)` to the min (only when the min is over one point or EMinLev1 is set: Lightning and Chain Lightning keep min 1) and the max (S, skill-damage-trace, V run 116,000 cases). No column of the 1.14d table gives a Sorceress skill a ToHit or physical (`DmgSymPerCalc`) synergy.

## 3. The plan against the synergies (I, `tools/emu/sk_pass_plan.py`)

Replay of `Build.Plan` + `Leveler.rule` over `build-fallback-lightning-v2.json` (FRESH) with quest points Den +1 at clvl 7, Radament +1 at 20, Izual +2 at 33 (assumed; without them everything below comes 2 levels later). Damage in whole points per hit (Charged Bolt per bolt), synergy % in brackets, Lightning Mastery included, no items, no resistance.

| clvl | points (CB / L / CL / LM) | Charged Bolt | Lightning | Chain Lightning |
|---|---|---|---|---|
| 10 | 8 / 0 / 0 / 0 | 5.5-7.5 (+0) | - | - |
| 15 | 12 / 0 / 0 / 0 | 7.5-9.5 (+0) | - | - |
| 20 | 17 / 0 / 0 / 0 | 10.5-12.5 (+0) | - | - |
| 24 | 20 / 0 / 0 / 0 | 13.5-15.5 (+0) | - | - |
| 25 | 20 / 1 / 0 / 0 | 14.3-16.4 (+6) | 1-104 (+160) | - |
| 27 | 20 / 1 / 2 / 0 | 14.3-16.4 (+6) | 1-110 (+176) | 1-94 (+84) |
| 29 | 20 / 1 / 4 / 0 | 14.3-16.4 (+6) | 1-117 (+192) | 1-134 (+84) |
| 30 | 20 / 1 / 4 / 1 | 21.5-24.6 (+6) | 1.5-175 (+192) | 1.5-202 (+84) |
| 33 | 20 / 1 / 5 / 5 | 28.3-32.5 (+6) | 2.0-238 (+200) | 2.0-306 (+84) |
| 37 | 20 / 1 / 9 / 5 | 28.3-32.5 (+6) | 2.0-263 (+232) | 2.0-474 (+84) |
| 42 | 20 / 6 / 9 / 5 | 36.4-41.7 (+36) | 2.0-526 (+232) | 2.0-525 (+104) |
| 50 | 20 / 6 / 9 / 13 | 54.0-62.0 (+36) | 2.9-781 (+232) | 2.9-780 (+104) |
| 57 | 20 / 6 / 9 / 20 | 69.4-79.7 (+36) | 3.8-1004 (+232) | 3.8-1002 (+104) |

What the structure says (for the owner of the build; nothing changed):
- Charged Bolt 20 by clvl 24 is the synergy core: it gives Lightning +160 % and Chain Lightning +80 % from their first point, while Charged Bolt's own synergy (Lightning x 6 %) stays at +6 % until clvl 37 and +36 % at most.
- Chain Lightning, the main attack from clvl 26, never gets more than +104 %: its feeders are Charged Bolt (20, done), Lightning (1 until clvl 37, 6 at most) and Nova (never). Lightning carries +232 % from clvl 37 but is second in the attack order and gets 6 points.
- Lightning Mastery's first point is the single biggest step (+50 % on every lightning hit, Charged Bolt included); each later point adds 12 points of percent to 150-278 %, so its relative gain falls from 8 % to 4 % a point.
- Points with no damage synergy: Frozen Armor's second point (row 22: +5 % armor, +12 s; Shiver / Chilling Armor would add +10 s each but are not in the plan), Frost Nova 1 (feeds nothing lightning), Telekinesis 1 (Teleport's prerequisite; Energy Shield is in the lightning profile's buff list but never learned, so its calc2 synergy is unused), Warmth (clvl 91), Static Field (clvl 86+; no synergy, a boss tool). None is a lightning damage point wasted; the only obvious unused synergy is Nova (the same 8 % / 4 % as Charged Bolt, so it would matter only after Charged Bolt is maxed).
- END table (replay with `cold`): Ice Bolt 6, Ice Blast 12, Glacial Spike 6 give Blizzard +120 % from its first point to clvl 60; Ice Bolt itself reaches +585 % but is not cast late; Frost Nova +200 % at Blizzard 20; Warmth 1 at clvl 11 (max mana about 55: +25 % regen, I); Cold Mastery pierce -20 % at clvl 47 to -80 % at 60, never against an immune.

## 4. d2d's version (S, upstream main 72132b99)

- Passives: `rules::passive_stats` (components/rules/skills.hpp:741) evaluates every passivestat of each skill with a level (env.level, with items), summoners' skipped; `Fight::update_fighters` (components/game/fight/strike.cpp:118..) recomputes them each update into `psum`. Equal result to FUN_00646d60's set-on-change for the Sorceress (I); the +0x80 state hold-off is not modelled (Paladin auras only, its own comment).
- Masteries: `elem_damage(.., mastery = true)` reads `env.stat(329 + etype)` once (S, V by the skill-damage differential). Cold Mastery's 335 reaches the hit through `Fight::pierce()` (components/game/fight/missiles.cpp:538) into `rules::missile_blow` (components/rules/combat.hpp:423): `res < 100` -> `max(res - pierce, -100)`: **equal to FUN_0057be00 for monster targets** (V read vs S). Its own comment says the game.exe code was not traced; it now is (Q5).
- Warmth: `rules::mana_per_frame` (components/rules/rules.hpp:880) `base x (100 + stat 27) / 100 + stat 26` with stat 27 from the fighter (`fighter.mana_regen`, combat.hpp:155), passives included: equal (plain multiply = MulDiv below 0x100000, I).
- +skills: `rules::item_skill_bonus` (skills.hpp:882) cites FUN_00644180 and has 127, 83, 188, 97 (cap 3 own class), 126, 107; the shrine comes in as `extra` (fight.hpp:141); `entry +0x2c` is not kept (its own comment). Synergies through operand 41 `env.base_level` (skills.hpp eval_calc kCalcSkill): equal split (S).
- Telekinesis (do 21): `components/game/fight/strike.cpp:757`: the missile blow on the attacked monster with `blow.knockback = true` always; no item or object branch in do 21 (by grep).

## 5. Deviations d2d vs game.exe (D1..)

| id | rule | game.exe | d2d | tag | differential |
|---|---|---|---|---|---|
| D1 | Telekinesis knockback | rand(100) < Param2 (35): knockback flags 9 only then (0x5c99c0..0x5c9a47) | knockback on every hit (strike.cpp:757) | V read + V run (Param2 at +0x14c) / S | none under tools/emu; could be built: run FUN_005c98f0 natively for a player and a monster target, hook FUN_00553540 (target), FUN_00554200 / FUN_0061ab00 (not town), FUN_0057c6c0 / FUN_0057cee0 (record the flags word), seed the caster's RNG per case; oracle = the knockback bit over 100,000 seeds (expect 35.0 %) vs d2d's 100 %. No plan impact (Telekinesis is never cast). |
| D2 | Telekinesis on items / objects | picks up items of types 4, 0x16, 0x26, 0x38, 9, 0x29 (FUN_00563560) and operates objects (FUN_00584420) | do 21 handles a monster only | V read / S (grep) | structural only (no pure function); not worth a harness for the plan |
| D3 | skill level bonus `entry +0x2c` | added in FUN_00644180 | not kept (own comment) | V read / S | none; the source of +0x2c is U (a drill or a read of its writers would come first) |

Clean (no deviation, risk moves from U to V read / S): passive stat values and the with-items level (Q1); mastery once at creation (S, earlier differential); pierce gate and floor for monster targets (Q5); Warmth in regen (Q6); the +skills stat list and the blvl / lvl split (Q8, Q9) by reading. A blvl != lvl differential could be built on the existing `skill_damage.py`: hook FUN_006442a0 to return different values for flag 0 and flag 1 (base vs with items) and give d2d's `CalcEnv` the same split; oracle = FUN_00644d50 / 00644e40 output; it would make Q9 V run.

**PR candidate: none.** D1-D3 have no existing differential (the brief's bar: proven with the existing harness, mismatches before > 0, after 0), and none touches the plan's casts.

## 6. Bot findings (B1..; described, not changed)

| id | finding | file:line | tag | impact |
|---|---|---|---|---|
| B1 | The FRESH table is not spent in table order past clvl 24: the Lightning Mastery rows (26, 28-31) are refused until clvl 30 and the Chain Lightning rows (27, 32-39) until Lightning has a point (Skills reqskill1), so the planner (blocked -> next row) puts the clvl 25 / 27 point into the **clvl 40 Lightning row**; then Chain Lightning from clvl 26 (28), Lightning Mastery 1 a level from 30 to 5 at 33 (34), Chain Lightning to 9 at 37, Lightning 2-6 at 38-42, LM 6-20 at 43-57. The matrix's reading (Chain Lightning 27, Lightning 40) never happens; the attack band switches at clvl 26 (`internal/combat/classdata/profiles.json`, lightning band minLevel 26), one level after Lightning is learned with quest points | `internal/leveler/plan.go:82-95`, `internal/leveler/leveler.go:272-279`, `internal/plandata/build-fallback-lightning-v2.json` rows 26-40 | I (replay of the bot's own rule) + D | the order is legal and arguably better (Lightning's +160 % synergy is available early); but anyone reading the table expects Chain Lightning at 27 without Lightning: worth a table note or reorder by the build owner |
| B2 | `Skill.Effective()` = base + stat 107 only; game.exe adds 127 (+all), 83 (+class), 188 (+tab), 97 (oskill, cap 3), 126 (+elemental, by element byte), the shrine's +2, and caps | `internal/d2/skills.go:16`, `internal/d2/player.go:190`; used by `internal/combat/skills.go:141,148`, `internal/combat/classdata.go:251`, `cmd/d2bot/display.go:180` | V read vs our code | with +class / +tab / +all gear (Sorceress orbs, Leaf's +3 fire skills = 126) or a skill shrine the bot's level and mana cost are low: Charged Bolt and Lightning cost 0.5 more mana per missing level (lvlmana 4 << 5 and 1 << 7, / 256), Chain Lightning 1.0 more (1 << 8); KnownSkills keeps the skill (base > 0). Small early, larger once +skills gear exists |
| B3 | The leveler counts base points (`Skill.Level`, +0x28) for what is owned | `internal/leveler/leveler.go:185-188` | V read | correct: matches game.exe's point count and the synergies' blvl; +skills gear does not distort the plan |
| B4 | No damage, synergy or mastery model: `Skill.Synergies()` is unused outside tests; skills are chosen by profile order, not expected damage | `internal/data/skills.go:144`, `internal/combat/skills.go:125` | S (our code) | nothing wrong today; if a damage-based pick is ever added it must use blvl for synergies, lvl for brackets and the mastery once (section 2) |
| B5 | Immunity threshold `Res >= 100` | `internal/combat/target.go:102` | V read (Q5) | correct: Cold Mastery (END) never breaks a cold immunity, so the cold-immune fallback stays right with any Cold Mastery level |
| B6 | Warmth's value depends on max mana: 0 under 47 max mana, +20-29 % above (I); the bot's regen model is linear in the bonus (SKMANA B2, `internal/resource/types.go`) | `internal/resource/types.go` (ManaPerSec) | V read + I | END's Warmth at clvl 11 helps (+25 % at 55 max); the model over-credits Warmth on tiny pools (already SKMANA B2) |

## 7. Drill recipe (lab binary, fork verbs on main: docs/forkdbg.md, docs/d2d-fork.md, docs/forkcheats.md; not run by this card)

Setup: one hidden slot, lockstep, a fresh throwaway Sorceress `--no-save`, every cheat ledgered with the reason "SKPASS drill Pn". `debug stat 12 30` (clvl 30, so Lightning Mastery can take points). Expected values are game.exe's (from the reads above); d2d's expectation in brackets.

- **P1 Lightning Mastery 0 vs 5 (Q3).** `debug points 49 1` (Lightning L1, 1-40 with no synergy: keep Charged Bolt at 0; whether `debug points` skips the Charged Bolt prerequisite is U, else accept CB n and expect x (1 + 0.08 n) on the max), `debug spawn zombie1 1 near`, `debug mon set <id> life 100000 maxlife 100000 res light 0`. 200 x (`debug mana full`, `debug cast 49 unit <id>`, `step 20`), sum the life lost. Then `debug points 63 5` and repeat. Expect mean per hit 20.5 -> 40.6 points (x 1.98) and the smallest hit 1 -> 1.98 (mastery on the min, the synergy gate not). [d2d: the same; equal by S.] Readback of the passive itself: `dump stats` stat 330 = 98 after `debug points 63 5` (and 50 at 1).
- **P2 Warmth regen (Q6).** `debug stat 9 <100 x 256>`, `debug stat 8 0`, `step 750`, `dump stats` raw[8]: no Warmth 8 x 750 = 6,000; `debug points 37 1`: 10 x 750 = 7,500; `debug points 37 5` (78 %): 14 x 750 = 10,500. Repeat at max 45 (11,520): 2,250 / 2,250 / 3,750. [d2d equal; DRILLS15 M4 already showed the 45 / Warmth 1 case.] Check that `debug stat 9` holds (the fork may recompute max mana on update; then set energy instead and compute max from it).
- **P3 Cold Mastery pierce and immunity (Q4, Q5).** `debug points 39 1` (Ice Bolt 3-5), spawn, `debug mon set <id> life 100000 maxlife 100000 res cold 50`; 200 casts; then `debug points 65 1` (-20): damage x 0.70 / 0.50 = 1.40; `res cold 95` with Cold Mastery 20 (115): resistance -20 -> x 1.20 / 0.05; `res cold 100` with Cold Mastery 20: 0 damage, no chill. [d2d equal by S.]
- **P4 +skills, blvl vs lvl (Q8, Q9).** Needs an item with a chosen +skills property; the fork's `debug give` makes random affixes only (forkcheats). Proposed verb for a later fork card: `debug give <code> pack prop <stat> <param> <value>`. With it: Charged Bolt 5 base, Lightning 1, an item +3 Lightning Skills tab (188, param 1 x 8 + tab - 1): Charged Bolt cast at L8 (bolts 3 + 7 = 10, `debug missiles`; mana `(24 + 4 x 7) << 5` = 1,664 256ths), while Lightning's synergy from Charged Bolt stays 5 x 8 = 40 % (Lightning max 40 x 1.40 = 56). Without the verb: U.
- **P5 Telekinesis knockback rate (D1).** `debug points 43 1`, spawn, big life, 200 x `debug cast 43 unit <id>`; count knockbacks (the monster's mode / position jump in `monsters`). Expect 35 % in game.exe [d2d 100 %].
- **P6 Leveler replay (B1), offline.** `python3 tools/emu/sk_pass_plan.py internal/plandata/build-fallback-lightning-v2.json quests`; live check on a FRESH character crossing clvl 24-30: the leveler log shows `Lightning` at the clvl 25 / 27 spend and `Lightning Mastery` from 30.

## 8. Matrix cell updates (docs/research/sorc-skill-matrix.md; proposed text, not edited by this card)

| Row / place | Cell | New value |
|---|---|---|
| inventory row 37 Warmth | effect | passive stat 27 = 30 + 12 (L-1) %, MulDiv on the truncated per-frame regen base: 0 under 47 max mana, +20-29 % above at L1 [V read FUN_005806f0 + D, sk-passives-trace]; EType fire (+fire skills count) |
| inventory row 43 Telekinesis | rules | do 21 FUN_005c98f0: lightning 1-2 (+1 / level) on a monster, knockback at 35 % (Param2), picks up gold / potion-class items, operates objects; d2d always knocks back (D1) [V read] |
| inventory row 63 Lightning Mastery | plan use | first point clvl 30 (reqlevel); refused points go to later rows, not saved: Lightning 1 at clvl 25 (27 without quest points), Chain Lightning from 26 (28), LM 5 at 33 (34) [I replay, sk-passives-trace B1] |
| note under the inventory (Lightning Mastery) | text | replace "the later spend of the saved points was not read, U" with: "the planner does not save them: blocked rows pass their points to the next legal row; the Lightning row of clvl 40 is spent at clvl 25-27 because Chain Lightning needs Lightning (sk-passives-trace B1)" |
| inventory row 65 Cold Mastery | effect | stat 335 `passive_cold_pierce` = 20 + 5 (L-1), subtracted from the monster's cold resistance only when it is under 100, floored at -100; no damage % [V read FUN_0057be00 + D] |
| band table A3 / A4 / A5 | skills | A3: Lightning (1), Chain Lightning 1-5, Lightning Mastery 1-5 from clvl 30; A4 / A5: Chain Lightning 5-9, Lightning 1-6, LM 5-8 [I replay] |
| 3.6 Telekinesis (a) | rules | as row 43 above [V read] |
| 3.8 Lightning Mastery (a) | rules | stat 330 = 50 + 12 (L-1) % on the caster by FUN_00646d60 at the with-items level, set when learned, refreshed on level change (FUN_00646f20); added once at missile creation to min and max (min 1 -> 1.5 at L1) [V read + S]; reqlvl 30 [D] |
| 3.8 Lightning Mastery (b) | d2d | `passive_stats` recomputed each update, mastery once in `elem_damage`: equal [S] |
| END rows Cold Mastery / Warmth | (a) / (b) | as rows 65 / 37 above; d2d: pierce `res < 100` gate and -100 floor equal (S vs V read); Warmth in `mana_per_frame` equal [S] |
| K3 (damage formula) | add | synergies read base points (`blvl`, FUN_006442a0 flag 0); cast level, mana, durations and passives read the level with +skills (flag 1: FUN_00644180's 127 / 83 / 188 / 97 cap 3 / 126 / 107 / shrine +2) [V read] |
| new K row: +skills in the bot | tag | the bot's `Effective()` adds only stat 107 (B2) [V read vs code] |
| section 3 tally / "Skills with no V on their own behaviour" | text | remove Telekinesis, Cold Mastery and Warmth from the list (now V read) |

## 9. Not traced (U)

- The source of `entry +0x2c` in FUN_00644180 and the cap FUN_00611830(0) (0x96c8a8[0]).
- The roles of FUN_00646f20's callers (item equip, stat callbacks: I from their neighbourhood only).
- FUN_0063ee90 (the "defender is such a monster" test in the hit context): which monsters fail it (I: player-side pets and hirelings get the player rules: pierce on immunes, penalty, cap).
- Telekinesis's item type ids 4 / 0x16 / 0x26 / 0x38 / 9 / 0x29 by name; its start function FUN_005c9030 (pays mana in the start, SKMANA).
- Item-granted skills at 0 base points (whether an entry exists for +all-skills-only skills): S from d2d's FUN_0055b800 comment, not read.

## Shared-file lines for the merger

docs/decomp-queue.md, row 23, status column (replace `proposed (SKILLINV 2026-11-01)`):

```
done 2026-11-01 (SKPASS, Opus): passives FUN_00646d60 (set at learn, refreshed FUN_00646f20, level with +skills); Lightning Mastery stat 330 = 50 + 12 (L-1) % damage, once at creation; Cold Mastery stat 335 = pierce, only under 100 resistance, floor -100 (FUN_0057be00, V read); Warmth stat 27 MulDiv on the truncated regen base (0 under 47 max mana); +skills FUN_00644180 (127 / 83 / 188 / 97 cap 3 / 126 / 107 / shrine +2) feed lvl, synergies read blvl; synergy table from the patched Skills.txt; d2d D1 Telekinesis knockback always (game.exe 35 %), D2 no item / object branch, D3 entry +0x2c; no PR candidate; bot B1 (leveler spends the clvl 40 Lightning row at 25-27, LM from 30), B2 (Effective() = base + 107 only); docs/research/sk-passives-trace.md; tools/emu/sk_pass_tbl.py, sk_pass_plan.py
```

docs/BACKLOG.md (two lines, decomp / bot section):

```
- SKPASS B1: the FRESH lightning table is spent out of order past clvl 24 (Lightning row 40 at clvl 25-27, Chain Lightning 26+, LM 30+); build owner to reorder or annotate rows 26-40 (docs/research/sk-passives-trace.md B1)
- SKPASS B2: Skill.Effective() reads base + stat 107 only; game.exe adds 127 / 83 / 188 / 97 / 126 / shrine (FUN_00644180); affects mana cost and spec level with +skills gear (docs/research/sk-passives-trace.md B2)
```

docs/INDEX.md (research list):

```
- docs/research/sk-passives-trace.md: Sorceress passives (masteries, Warmth, Telekinesis), the 1.14d synergy table, +skills base vs level; decomp row 23 (SKPASS)
```

docs/research/README.md (trace list):

```
- sk-passives-trace.md: SKPASS, masteries / pierce / Warmth / Telekinesis / synergies / +skills (V read, D, I replay)
```

CHANGELOG.md: none (docs only).

docs/OWNERS (while the card is open): `docs/research/sk-passives-trace.md skpass`, `tools/emu/sk_pass_* skpass`, `openspec/changes/skpass/** skpass`.

Glossary terms missing (for the report, not edited here): "blvl / lvl (base level vs level with +skills)", "pierce (resistance reduction stat)", "synergy".
