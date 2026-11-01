# Monster resistances and immunities by difficulty: game.exe 1.14d vs d2d vs the bot (decomp row 21, SKRESIST)

> status: card skresist, 2026-11-01, offline (no game, no d2d slot, nothing pushed or posted). Decomp row 21 (sorc-skill-matrix rank 8). Question: how does game.exe 1.14d build a monster's elemental resistances and immunities by difficulty (MonStats columns, difficulty penalties, immunity rows, the 100 line and the -100 floor, boss modifiers), what do Lower Resist / Conviction / Cold Mastery-style reductions do to the final damage, and does d2d give the same numbers?

Tags: **V read** disassembled by this agent with upstream's `tools/emu/disasm.py` on the patched 1.14d game.exe (addresses given); **V run** game.exe's own code run natively under the emulator by this agent (`tools/emu/sk_resist.py`), or the real d2d build run by this agent; **D** the patched 1.14d tables as game.exe's MPQ stack serves them (`sk_resist.py tbl`, `sk_pass_tbl.py`); **S** d2d source at upstream main 72132b99 or an earlier trace of ours (named); **I** inference or arithmetic; **U** not verified.

## Answer in six lines

1. **Base resistances = the MonStats column of the game's difficulty, nothing else.** At spawn FUN_00573cb0 sets the six base stats 0x24 / 0x25 / 0x27 / 0x29 / 0x2b / 0x2d from MonStats ResDm / ResMa / ResFi / ResLi / ResCo / ResPo of difficulty 0..2 (a hireling always reads the Normal column) (V read). **A monster gets no difficulty penalty, no 75 cap, no maximum-resist stat**: those are for players and hirelings only (FUN_0057be00, V read). Boss / champion / minion adds (Magic Resistant, the enchanteds, Stone Skin, the two-immunity limit) come on top at spawn (FUN_005a1370; S, monster-mods-trace.md, matched by d2d).
2. **Immunity is base + curses >= 100**, decided at the hit: FUN_0057bf80 turns res >= 100 into 0 damage (V read; S PR #135 proof). There is no other immunity flag.
3. **Pierce** (Cold Mastery's stat 335, items' 333 / 334 / 336) is subtracted only while the monster's total resistance is under 100; at 100+ it is ignored for a monster (players and hirelings take it at any value). **The floor is -100 for every damage type**, physical and magic too (V read + V run).
4. **Curses and auras on an immune (the 1/5 rule)**: FUN_005c3540, called by the curse callback (srvdofunc 30) and the aura callback (srvdofunc 65 / 66), divides a NEGATIVE change to stat 36 / 37 / 39 / 41 / 43 / 45 by 5 (truncated toward 0) when the monster's BASE value of that stat is >= 100; players and hirelings take the full value (V read + V run, 147,628 cases). Lower Resist level 1 (-31) takes a lightning 110 Hell monster to 104 (still immune); Conviction level 20 (-125) to 85 (I, arithmetic on the rule).
5. **d2d deviation D1 (V run on the real build):** d2d adds the full curse / aura value on an immune (`Fight::apply_target_stats`): 88,101 of 147,628 cases differ in the reduction, 14,355 in the damage dealt (seed 7: 88,224 / 14,289). Everything else on the missile path matches: base columns, pierce under 100 only, -100 floor for the four elements, MulDiv, the 100 line.
6. **Bot:** reads the merged stat list (the game's own total, so curses come in already divided): its `>= 100` immunity test is right; it ignores pierce (Cold Mastery) in its damage factor and has no -100 floor (B1, B2). No PR candidate (section 7).

## 1. Questions answered (Q1..)

| id | question | answer | where | tag |
|---|---|---|---|---|
| Q1 | Where do a monster's resistances come from at spawn? | FUN_00573cb0(game, ..., unit in esi): MonStats record (stride 0x1a8 index -> MonStats record stride 0x134); difficulty index = game byte +0x6d clamped to 0..2, **0 when FUN_0063ee90(unit) (a hireling class)**; base stats set with FUN_00627260: 0x24 = record +0x144 + 2 diff (ResDm), 0x25 = +0x14a (ResMa), 0x27 = +0x150 (ResFi), 0x29 = +0x156 (ResLi), 0x2b = +0x15c (ResCo), 0x2d = +0x162 (ResPo) (int16 each); 0x14 toblock from +0xa3 + diff; level 0xc from +0xaa (expansion and difficulty > 0 and not a boss row: the area's level, FUN_0061a1b0 / FUN_0061dca0). No difficulty adjustment of the resistances anywhere in this function. | 0x573e83..0x573ef0, 0x573dc1..0x573e03 | V read |
| Q2 | Which rows are immune per difficulty? | MonStats rows with a value / immune (>= 100) / negative: Normal fire 122 / 1 / 2, lightning 104 / 0 / 0, cold 131 / 0 / 0, poison 191 / 22 / 2, physical 60 / 0, magic 50 / 0; Nightmare fire 149 / 11, lightning 142 / 15, cold 166 / 16, poison 218 / 48, physical 74 / 3, magic 68 / 0; Hell fire 372 / 113, lightning 364 / 105, cold 397 / 137, poison 392 / 131, physical 472 / 20, magic 140 / 11 (max 145, cold max 1000). Counts are rows, not spawn weights; unused rows are included. | `sk_resist.py tbl` | D |
| Q3 | Boss / champion / unique / minion modifiers | Applied once at spawn by FUN_005a2120 -> FUN_005a1370 (Magic Resistant +40 fire / cold / light, Fire / Cold / Lightning Enchanted +75 by difficulty rows, Stone Skin DR, Spectral; nothing more once two immunities exist). Read, differentialled and matched against d2d's `resist()` lambda in docs/research/monster-mods-trace.md (157,618 cases; residual S1 Stone Skin defence). That they are BASE stats (so they count for the 1/5 rule) is from the setter used there. | monster-mods-trace.md rows 8 / 9 / 17 / 18 / 27 / 28 | S (adds); I (base list) |
| Q4 | The resistance a hit uses | FUN_0057c1e0 builds the context (+0 game, +4 DifficultyLevels row FUN_00611d30, +8 attacker, +0xc defender, +0x14 = defender is type 1 and NOT FUN_0063ee90 = **a plain monster**); per element FUN_0057bf80 -> FUN_0057be00 on the descriptor at 0x732980 (+4 resist stat, +8 max stat, +0xc pierce stat): res = total stat (FUN_00625480: the merged list); pierce subtracted when res < 100, or when the defender is NOT a plain monster; not a plain monster and not DR / MR: + DifficultyLevels ResistPenalty (expansion; classic -20 / -50 by difficulty) and capped at 75 + max stat (DR without a max stat: 50), at most 95; res <= -100 becomes -100 (every stat, monsters included); physical with the attacker under state 0x2f (`sanctuary`, D) against FUN_0063e990 (a MonStats flag-byte test, undead by the masks, I): DR = 0. | 0x57be00..0x57bf05, 0x57c1e0..0x57c24b | V read + V run (oracle `effres` column) |
| Q5 | Damage from the resistance | FUN_0057bf80: flat reduction first (stat 0x22 physical, 0x23 magic, << 8), then dmg > 0 and res != 0: res >= 100 -> 0, else MulDiv(dmg, 100 - res, 100) (FUN_00483360's two branches); negative res multiplies, at most x2 by the floor. | 0x57bf80..0x57bff4 | S (PR #135 proof) + V run (oracle `dealt`) |
| Q6 | Curses and auras on an immune | FUN_005c3540(unit ecx, stat edx, value): stat 36 / 37 / 39 / 41 / 43 / 45 (byte table 0x5c35b0: the max-resist stats 38 / 40 / 42 / 44 are passed through), value <= 0, unit not a player (type 0) and not a hireling (FUN_0063ee90), BASE stat (FUN_006253b0 -> FUN_00624ed0: the vector at statlist +0x24) >= 100: return value / 5 (cdq-style truncation toward 0); else the value. Callers: FUN_005c35c0 and FUN_005c3de0 (curse target callbacks of srvdofunc 30 = FUN_005c37c0, table 0x7322b0) for the first aurastat and the five others; FUN_005cedc0 (aura callback of srvdofunc 65 = FUN_005cf010 and 66 = FUN_005cf3a0). The adjusted value is what goes into the state's stat list (FUN_00627150), so the merged total the hit reads is base + value / 5. In FUN_005c35c0 a FIRST aurastat that comes out 0 applies no curse at all (0x5c3639) | 0x5c3540..0x5c35a3, 0x5c35c0..0x5c371b, 0x5ceeae, table dwords 0x7322b0 + 4 x 30 / 65 / 66 | V read + V run (`adj` column) |
| Q7 | The curse values (D) | Lower Resist (91) `-dm56` on fire / light / cold / poison, Param5 25, Param6 70; Conviction (123) `-min(ln34,150)` on fire / cold / light (30 + 5 (L-1)) plus `skill_armor_percent -dm56`; Amplify Damage (66) `-par5` = -100 damageresist; Decrepify (87) `par5` = -50 damageresist (+ speed, damage, attack rate). Lower Resist L1 = 25 + 45 x (110 x 1 / 7) / 100 = 31 (I: the dm formula from upstream skills.md). No 1.14d curse lowers magic resistance. | Skills.txt patched | D (values), I (dm arithmetic) |
| Q8 | Masteries and pierce | Cold Mastery = stat 335 `passive_cold_pierce` 20 + 5 (L-1); Fire / Lightning Mastery are damage %, not pierce (S, sk-passives-trace.md Q3 / Q4). Pierce works only under 100 (Q4): it never breaks an immunity, but after a curse brings an immune under 100 it applies in full on top. | sk-passives-trace.md | S + V read |
| Q9 | What an immune hit still does | Total 0 after resistances: no get-hit (FUN_0057cb00: none under one point; S pr-get-hit-rule-proof.md); no chill / freeze (the cold length is cut by the same resistance and no chill under one point of cold; S static-field-chill-trace.md C1, freeze-rules-trace.md C5); Static Field: lightning-immune takes 0 (its record goes through the same path; S). Knock-back flags, the hit events (wake-up, Lightning Enchanted's on-hit charged bolts) on a zero-damage hit: not traced (U). | | S / U |
| Q10 | The resistance penalty | DifficultyLevels ResistPenalty 0 / -40 / -100 (D) is added in FUN_0057be00 only for defenders that are NOT plain monsters (players, hirelings). The monster's resistances are the same table numbers in every difficulty's hit path; Nightmare / Hell monsters are tougher only because their own columns are higher. | 0x57be3d..0x57be72 | V read + D |

Dead ends: the 0x66666667 (divide by 5) hits in the full listing are mostly UI and drawing code; the one next to a `cmp eax, 0x64` in the server range is FUN_005c3540 (the rule). FUN_00573cb0's difficulty clamp has a dead branch for value 1 (it falls to `edi = al`).

## 2. The rules as a formula (monster defender)

```
base[s]  = MonStats Res*(difficulty)[s] + boss-mod adds          (spawn, base stats)
adj      = v / 5 (toward 0)  if v < 0 and base[s] >= 100  else v  (each curse / aura stat, at apply time)
total    = base[s] + sum(adj)                                      (merged stat list)
eff      = total - pierce  if pierce stat and total < 100          (monsters: no penalty, no cap)
eff      = max(eff, -100)                                          (all six types)
dealt    = 0 if eff >= 100 else MulDiv(dmg - flat, 100 - eff, 100) (256ths)
```

Worked numbers (I, arithmetic on the V rules): Hell lightning 110 monster: Lower Resist 1 (-31 -> -6): 104, immune; Conviction 1 (-30 -> -6): 104; Conviction 20 (-125 -> -25): 85, takes 15 %, then Cold Mastery-style pierce applies (for cold); Conviction 25+ (-150 -> -30): 80. A cold 1000 row (D) can never be broken. A non-immune cold 75 boss with Cold Mastery 20 (115 pierce): 75 - 115 = -40 -> 140 % damage.

## 3. d2d (S, upstream main 72132b99; same in the skmiss checkout)

- Base: `gamedata_load.cpp:224` reads ResDm..ResPo with the difficulty suffix "" / "(N)" / "(H)"; `Monster::target()` (ai.hpp:366) = `type.diff[difficulty].res + boss_res`. Same as Q1 (no penalty) [S + V read].
- Boss adds: `rules::boss_stats` / `resist()` (monster-mods-trace.md: equal apart from S1).
- Curses and auras: `Fight::target_of` (fight/missiles.cpp:800) applies the monster's curse / cry and the player's enemy aura through `apply_target_stats` (:813): **`target.res[i] += value` with no immunity check** (D1).
- Missile and skill hits: `rules::missile_blow` (combat.hpp:411): elements fire..poison take the pierce when res < 100 and are floored at -100 there; magic (etype 4) and physical are used as is: **no -100 floor** (D2, unreachable with 1.14d data on this path, below).
- Melee / weapon hits: `rules::player_blow` (combat.hpp:470-545): physical `resisted(phys, res[0])`, the weapon's and gear's fire / lightning / cold (and conversions) by `resisted(..., res[i])`: **no pierce and no -100 floor** for the elements (D3).
- Sanctuary's undead DR bypass: not modelled (no `skill_bypass_undead` in components/) (D4).

## 4. Deviations (d2d vs game.exe)

| id | rule | game.exe | d2d | tag | reach in 1.14d | differential |
|---|---|---|---|---|---|---|
| D1 | curse / aura reduction on an immune | negative change / 5 when base >= 100 (FUN_005c3540) | full value (`apply_target_stats`) | **V run, real build**: 147,628 cases, `adj` 88,101 differ, `dealt` 14,355 (seed 7: 88,224 / 14,289); 0 cases differ outside "base >= 100 and v < 0" | Hell (100+ immune rows per element, Q2) and NM: Lower Resist from Necromancers / wands / charges, Conviction (Paladin, a hireling's Infinity), Amplify Damage / Decrepify on physical immunes (Hell DR 100: 20 rows): d2d breaks immunities game.exe keeps (Lower Resist 1 on lightning 110: d2d 79, takes 21 %; game.exe 104, immune) | yes: `sk_resist.py game` vs `test_resist --resist` (this card); exists |
| D1b | a curse whose first aurastat rounds to 0 on an immune | curse not applied at all (0x5c3639) | applied | V read | unreachable with 1.14d skills (smallest first-stat value on an immune is Lower Resist's -25 -> -5) | not built (I: unreachable) |
| D2 | -100 floor on magic and physical (missile path) | floored | not floored | **V run, real build** (magic: 1,958 / 1,977 cases with total < -100 differ in `dealt`) | unreachable with 1.14d data: no curse lowers MR, min ResMa 10, Amplify -100 on min ResDm 10 gives -90 (D) | yes (same harness) |
| D3 | melee / weapon elements: pierce and floor | pierce under 100, floor -100 (same FUN_0057be00 for every record) | `resisted()` without pierce or floor | V read (game) + S (d2d) | reachable: Conviction 25+ (-150) on any fire / cold / light res under 50 gives -100..-150 (d2d up to x2.5 instead of x2), Lower Resist 20+ (-70) on fire -50 rows; any -enemy-res item (Griffon's, Facets) on a melee build | port only: needs a melee record harness (FUN_0057b7d0 + FUN_0057c6c0 natively vs `player_blow`); could be built on hit_absorb.py's setup |
| D4 | Sanctuary undead DR | DR = 0 when the attacker has state `sanctuary` and the defender is undead (I) | none | V read (game), S (d2d) | Paladin only | not built (needs FUN_0063e990's mask meaning) |

Limits not covered: hireling defenders (penalty, cap, Normal column) and players as defenders; absorb; the poison per-tick path; the curse VALUE calcs (dm56 / ln34) were not differentialled (d2d's `eval_calc` is shared with PR #154's proven calcs, I); cold length cut by the immune 1/5 total (same total, I).

## 5. The differential (V run)

Oracle `tools/emu/sk_resist.py` (runs next to upstream's tools/emu: `emu.py`, `life_apply.py`): per case it runs natively FUN_005c3540 (curse adjustment; base from the case), puts base + adj as the monster's total, then FUN_0057be00 on the element's descriptor row (effective resistance; context +0x14 = plain monster) and FUN_0057c6c0 with only that element's record field set (the damage dealt, life in 256ths, life large enough that no case kills). d2d side: `tests/game/test_resist.cpp` (our copy `tools/emu/sk_resist_test.cpp`) built on upstream main 72132b99 on the worker host: `adj` / `effres` follow d2d's code line for line, `dealt` is the REAL `rules::missile_blow` for magic / fire / lightning / cold (elo = ehi = the roll, so no draw); physical and poison print "-" (whole-point / per-tick paths, not this claim).

Cases (`gen 100000 SEED`): exhaustive grid 47,628 = 6 types x 21 base values (-50, -1, 0, 1, 20, 50, 75, 95, 99, 100, 101, 104, 105, 110, 115, 120, 125, 145, 150, 200, 1000) x 14 curse values (0, -1, -4, -5, -9, -10, -25, -31, -50, -70, -100, -150, +20, +75) x 4 pierce (0, 15, 60, 115; 1 for physical / magic) x 9 rolls (1, 255, 256, 257, 1000, 4096, 25600, 0x100000, 0x100001); plus 100,000 random (base -150..1000, curse -200..+40 or 0, pierce 0..200, roll 1..0x180000). xorshift32 on a 32-bit masked seed; the C++ side only reads.

| run (the worker host) | cases | adj differ | effres differ | dealt differ | categories |
|---|---|---|---|---|---|
| game.exe vs REAL d2d build, seed 1 | 147,628 | 88,101 | 92,035 | 14,355 | all "base >= 100 and curse < 0" (D1) or "total < -100" on magic / physical (D2); 0 other |
| game.exe vs REAL d2d build, seed 7 | 147,628 | 88,224 | 92,149 | 14,289 | same |
| d2d port (`sk_resist.py port`) vs REAL d2d build, seeds 1 and 7 | 147,628 each | 0 | 0 | 0 | the port is d2d |

md5: c1.tsv a1ea534a..., c7.tsv 1eab443c..., g1.tsv fd1425e8..., g7.tsv cef33a48... Pierce against an immune: game.exe and d2d agree (d2d's `res < 100` gate): this closes the contradiction in pr-missile-damage-fine-proof.md's limits ("game.exe lets it through for a plain monster"): it does NOT (V read 0x57be20..0x57be29, V run); sk-passives-trace.md Q5 was right.

Commands (the worker host; paths as on the worker host's checkouts):

```sh
cd <upstream>/tools/emu
python sk_resist.py gen 100000 1 > c1.tsv
python sk_resist.py game c1.tsv > g1.tsv          # about 80 s
<build>/tests/test_resist --resist c1.tsv > d1.tsv
python sk_resist.py cmp c1.tsv g1.tsv d1.tsv
```

## 6. The bot (our repo; described, not changed)

- Immunity test: `Target.Immune` = `HasRes && Res[d] >= 100` (internal/combat/target.go:102), `Monster.Immune` (internal/data/monsters.go:88). Right line (Q2 / Q5).
- Source of `Res`: `refineResists` (internal/combat/fighter.go:736) reads the unit's stat list through `ReadStatList` (internal/d2/stats.go:61: the merged vector at +0x48 when flag 0x80000000, the one FUN_00625480 reads, V read 0x624f60): **the game's own total, curses already divided by 5**. Fallback `resistsOf` (target.go:233): MonStats row of the difficulty + `AffixResists`.

| id | finding | file:line | tag | impact |
|---|---|---|---|---|
| B1 | `expFactor` multiplies by (100 - res) / 100 with no -100 floor | internal/combat/rules.go:144-149 | V read (game) + S (our code) | over-rates a target under -100 (only with Lower Resist / Conviction on negative rows): ranking only |
| B2 | pierce ignored: `DamageFactor` / `PickByResist` / `ImmuneWorkaround` use `Res[d]`, not `Res[d] - pierce` | internal/combat/rules.go:155-165, 175, 244 | V read + S | a Cold Mastery Sorceress under-rates cold against resistant, non-immune targets (cold 75 boss with Cold Mastery 20: real 140 %, bot 25 %): `resistTolerance` 0.6 may switch her off cold when cold is the best element (END build) |
| B3 | the fallback `AffixResists` differs from FUN_005a1370's limits (skips adds per third immunity; counts magic) | internal/combat/affix.go:21 | S (monster-mods-trace.md B-list) | fallback only (no stat list) |
| B4 | comments say the live list "includes curses and difficulty penalties"; monsters have no difficulty penalty (Q10) | internal/combat/target.go:232, fighter.go:735 | V read | wording; no behaviour |
| B5 | on the d2d backend the bot reads d2d's resistances, which carry D1: a cursed Hell immune shows under 100 on d2d and the bot attacks with the element; on game.exe it would skip | internal/d2denv/reader.go:152 (`ReadStatList` = d2d's `UnitStats`) | I (from D1; d2d's exported stats not read) | fidelity of d2d soaks with curses / Conviction only |

## 7. PR candidate

**No PR candidate built.** D1 is proven on the real build with an existing-style differential (88,101 -> 0 would need the fix: `apply_target_stats` taking the monster's base `res + boss_res` and dividing a negative value by 5 when it is >= 100, plus D1b's "first stat 0: no curse"). Not built in this card's time box and the card has no `fork-branch:` / `upstream-pr:` resource; it is a strong candidate for a follow-up card (one claim, harness ready: `sk_resist.py` + `test_resist.cpp`, before 88,101 / 14,355, after expected 0 in `adj` and `dealt` with D2 excluded or fixed alongside). D2 is unreachable in 1.14d data; D3 needs a melee harness first.

## 8. Drill recipe (lab binary, fork verbs; for a later live card, not run here)

Verbs (docs/forkdbg.md, docs/forkcheats.md): `debug difficulty`, `debug goto`, `debug spawn`, `debug points`, `debug skill`, `step N`, `cmd skill`, `monsters` (resists per monster). Missing verbs the drills want: `debug mon set <id> res <type> <v>` (pin a base), `debug curse <id> <skill> <lvl>` (or a Necromancer helper).

- R1 (base by difficulty, Q1): `debug difficulty` 0 / 1 / 2, `debug spawn` zombie1 near; `monsters`: fire / light / cold / poison equal MonStats ResFi.. of that difficulty (Normal zombie1 cold 50 / poison 75, D); a Hell lightning-immune row shows its 100+ value, no penalty.
- R2 (D1, the main check): Hell, spawn a lightning-immune row; cast Lower Resist level 1 on it (a Necromancer character or a wand with charges); `monsters`: game.exe 104 (110 - 6), d2d before a fix 79. Then Lightning at it: game.exe 0 damage, d2d damage.
- R3 (pierce never breaks immunity, Q4 / Q8): Sorceress with Cold Mastery 20 vs a cold-immune: Ice Bolt deals 0; vs a cold 75 monster: hp loss matches 140 % of the roll (oracle interval).
- R4 (floor, D3): Paladin Conviction 25 + a fire-damage weapon vs a fire 0 monster: damage x2.0 of the fire part (game.exe), x2.5 on d2d before a fix.
- R5 (Q9): a cold-immune hit by Ice Bolt: no chill state, no get-hit mode.

## 9. Matrix cell updates (docs/research/sorc-skill-matrix.md; proposed text for the merger)

| Row (matrix) | Cell | New value |
|---|---|---|
| Fire Bolt / Fire Ball etc. (a) resist, immunity (line 99) | resist, immunity | base = MonStats ResFi of the difficulty, no penalty / cap for monsters (FUN_00573cb0, FUN_0057be00 V read); immune at total >= 100 (0 damage, no get-hit, S); pierce only under 100, floor -100 (V run, sk-resist-trace.md); curses on an immune / 5 (FUN_005c3540, V run; d2d D1) |
| Lightning / Chain Lightning (a) resist, immunity (line 113) | resist, immunity | as fire with ResLi; lightning-immune: Lower Resist 1 leaves 104 (immune) in game.exe, 79 on d2d (D1, V run 147,628 cases) |
| Cold Mastery (line 228) | the pierce use | V read + V run: stat 335 subtracted from the monster's cold total only while it is under 100 (never breaks an immunity), floor -100 (sk-resist-trace.md Q4 / Q8); bot ignores pierce in its damage factor (B2) |
| Static Field (line 32 / 146) | resist | lightning-immune takes 0 (same path); bot check `tg.Immune` right (S) |
| rank 8 SKRESIST (line 252) | status | done (sk-resist-trace.md): D1 proven on the real build, no PR built; D3 melee needs a harness |
| NM / Hell gate (line 305) | SKRESIST | traced: monster resist build, immunity line, curse 1/5, pierce; open: D1 fix, melee D3 |

## 10. Not traced (U)

- What a zero-damage hit still triggers (wake-up / aggro, Lightning Enchanted's on-hit bolts, knock-back).
- FUN_0063e990's flag masks (undead or not) and Sanctuary's other effects.
- Whether d2d exports curse-adjusted resistances to the bot's reader (B5).
- The curse value calcs (dm56, ln34) on game.exe vs d2d (not run).

## Shared-file lines for the merger

docs/decomp-queue.md, row 21 (replace the status cell): `done (SKRESIST 2026-11-01, docs/research/sk-resist-trace.md)`; and append to the finished list:

```
- 21 monster resists (SKRESIST, 2026-11-01): base = MonStats Res* of the difficulty at spawn (FUN_00573cb0; hirelings the Normal column), no difficulty penalty / 75 cap / max stat for monsters (FUN_0057be00), immune at total >= 100, pierce only under 100, floor -100 for all six types (V read + V run); curses / auras on an immune: a negative change / 5 when the BASE stat >= 100 (FUN_005c3540 from srvdofunc 30 / 65 / 66, V read + V run); D1 d2d applies the full value (`apply_target_stats`): real build 147,628 cases, 88,101 adj / 14,355 dealt differ (seed 7 88,224 / 14,289), 0 outside the rule; D2 magic / physical floor (unreachable in 1.14d data), D3 melee elements without pierce / floor (read), D4 Sanctuary; no PR built (follow-up candidate); bot B1 no floor, B2 pierce ignored, B5 d2d soaks carry D1; docs/research/sk-resist-trace.md
```

docs/BACKLOG.md (fix queue, one line each):

```
- SKRESIST D1: d2d curse / aura reduction on immunes not divided by 5 (FUN_005c3540); harness ready (tools/emu/sk_resist.py + sk_resist_test.cpp), before 88,101 / 14,355; follow-up card for a local fix branch (docs/research/sk-resist-trace.md section 7)
- SKRESIST D3: d2d melee elements without pierce and -100 floor (combat.hpp player_blow); needs a melee record harness
- SKRESIST B2: bot damage factor ignores pierce (Cold Mastery) (internal/combat/rules.go:155-175)
- SKRESIST B1: bot expFactor has no -100 floor (internal/combat/rules.go:144)
```

docs/INDEX.md (research list):

```
- docs/research/sk-resist-trace.md: monster resistances / immunities by difficulty, curse 1/5 rule, pierce, floor; game.exe vs d2d (D1 real build) vs bot
```

docs/research/README.md (if it lists traces): `sk-resist-trace.md - SKRESIST, decomp row 21 (monster resists, immunities, curses, pierce)`.

CHANGELOG.md: none (docs and tooling only).

Glossary (proposed, not written): "immunity breaking (1/5 rule)": a curse or aura lowering a resistance whose base is 100 or more applies a fifth of its value (FUN_005c3540).
