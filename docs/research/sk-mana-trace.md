# Player mana: check, payment, timing, regen and potions; game.exe 1.14d vs d2d vs the bot (decomp row 19)

> status: card skmana, 2026-11-01, offline (no game, no d2d slot). Decomp row 19 (SKMANA, sorc-skill-matrix rank 6): how game.exe checks and pays a player skill's mana, when, how mana comes back (regen, potions, leech), the skill delay, and what d2d and the bot assume. Follows docs/research/sk-missile-trace.md (SKMISS), which found the pay-at-the-action-frame rule for the missile skills.

Tags: **V read** disassembled with upstream's `tools/emu/disasm.py` by this agent (addresses given); **V run** run natively in game.exe under the emulator by this agent (`tools/emu/sk_mana.py`); **D** from the patched 1.14d tables (Skills.txt, misc.txt, states.txt as game.exe loads them or read with `mpq-cat` from the patch stack); **S** d2d source (upstream main 72132b99) or an upstream / earlier trace (named); **I** inference; **U** not verified.

## Summary

- **Cost (V read + V run).** `(mana + lvlmana x max(lvl - 1, 0)) << manashift` in 256ths (Skills +0x18a, +0x18c, +0x188), the level from FUN_006442a0 flag 1 (with +skills). d2d's `rules::mana_cost` and the bot's `ManaCost` use the same formula: **equal on every row and level where the minmana floor does not bite** (130,049 of 130,840 cases, 37,388 payments equal to the 256th).
- **The check has no floor, the payment has one (V read).** The server's check FUN_0056c160 (and the client's FUN_00647540) compares stat 8 with the formula only; the payment FUN_0056bfe0 takes `max(formula, minmana x 256)`, refuses (takes nothing) when mana is short of that, and the skill do ignores its answer: **the cast happens for free**. A row with mana 0 and lvlmana 0 pays nothing at all, minmana or not.
- **When it is paid (V read).** A skill with a srvstfunc (Sorceress: Inferno, Telekinesis, Thunder Storm, Hydra) pays in the START (FUN_0056f640) when the start function succeeds; every other skill (Fire Bolt, Charged Bolt, Ice Bolt, Frozen Armor, Frost Nova, Static Field, Teleport, Lightning, Chain Lightning, Blizzard, ...) pays in the DO (FUN_0056f7f0) at the cast mode's action frame, after the srvdofunc and the missile. Only `usemanaondo` rows (Plague Javelin, Blade Fury: D) re-check the mana in the do. A cast cut off before its action frame (get-hit, death, a new mode) **costs nothing**.
- **d2d** checks and pays `mana_cost` (with the floor) at the command on every cast path (`Fight::cast`, `cast_missile`, `cast_summon`, `start_swing`): D1 timing (SKMISS D2, restated), D2 the floor in the check and the paid amount (proven: 791 of 130,840 cases, all explained), D3 interrupted casts cost mana.
- **Regen (V read)** FUN_005806f0 a frame: `max(maxmana / 3000, 1)` 256ths (ManaRegen 120 x 25 frames), x `(100 + stat 27) / 100` truncated, + stat 26; equal in d2d (S + V-d2d fidelity 1a). **The bot's model is 100 s to fill and linear in the bonus: 1.2x to 1.8x too fast on early pools (B2).**
- **Potions (D + V read):** mp1..mp5 = 20 / 40 / 80 / 150 / 250 mana over 128 frames (5.1 s), x2 for the Sorceress (FUN_0062a620); the bot's sim adds 40 at once (B3).
- **Lab binary and PR #128 (V by git):** the private branches lab6, lab-port-148 and lab-debug-skills contain upstream's #128 merge (af825eda); the old `lab` branch does not. Which binary the soak harness runs today: U.
- **PR candidate:** none built. D2 is proven on the real d2d build but only bites below 1 mana point (Teleport level 25+, high-level arrows / bone skills) or on Might (a mana-free aura d2d charges 1 point); D1 / D3 need a stateful harness (drill below).

## 1. The check

| Address | Rule | Tag |
|---|---|---|
| `FUN_0056c160` (ecx unit) | not a player (unit type != 0): 1. Current skill FUN_00620250: none: 0. An item-charge entry (FUN_00643ad0 != -1): `charges left (entry +0x38) > 0`. Skills row FUN_00644140, level FUN_006442a0(unit, entry, 1) - 1 floored at 0; cost = `(lvlmana (+0x18c, signed) x that + mana (+0x18a, signed)) << manashift (+0x188, byte)`. srvdofunc 0x74 (Werewolf / Werebear / Delirium change) and the unit already shifted (FUN_0063a400: unit +0xc8 bit 3): 1 (unshifting is free). Else `stat 8 >= cost`. **No minmana, no startmana** | V read (0x56c160..0x56c206) |
| `FUN_0056f640` (skill start; flag arg 0 for a player's cast) | FUN_0056c160 first: 0 refuses the skill (returns 0; FUN_0056faf0 then puts a player back to neutral via FUN_005809d0). Then InTown (Skills +5 bit 0) vs a town room, the target restriction (+0x18f), srvstfunc (table 0x732140, +0x2c < 0x5b) | V read |
| client `FUN_00647960` -> `FUN_00647540` | status 1 (need mana, the `<class>_needmana_1` voice) when `stat 8 < formula` (the same arithmetic, no floor); under state 0x72 (`blood_mana`, D) the formula is compared with **life** (stat 6) | V read (0x647540..0x647633) |
| client `FUN_006440f0` | status 1 also when `startmana` (Skills +0x184) > 0 and stat 8 < startmana x 256, unless state 12 (`inferno`, D): Inferno 6, Arctic Blast 6, Blade Fury 6 (D); no other Sorceress row has one | V read + D |

Skills record offsets (V read, matching upstream skills.md): +0x184 startmana, +0x186 minmana, +0x188 manashift, +0x18a mana, +0x18c lvlmana, +0x190 delay calc, +0x2c srvstfunc, +0x2e srvdofunc, +0x46 srvmissile. The flag bits use the mask table at 0x6ce268 (1 << n): +4 bit 0 decquant, bit 1 lob, bit 4 passive, bit 5 aura; +5 bit 0 InTown, bit 2 InGame; +8 bit 5 usemanaondo (V read of the uses; the bit names I from Skills.txt column order, matched by the client's status 5 / 6 / 3 and upstream's AttackNoMana at +7 & 0x10).

## 2. The payment and its timing

| Address | Rule | Tag |
|---|---|---|
| `FUN_0056bfe0` (ecx game, edx unit, skill, level) | not a player: 1. The current entry is the skill and an item charge: FUN_0056bec0 (one charge, no mana). Row: **mana 0 and lvlmana 0: returns 0, pays nothing**. cost = `(lvlmana x max(level - 1, 0) + mana) << manashift`, then `max(cost, minmana (+0x186) << 8)`. State 0x72 (blood_mana): FUN_005d2b60 pays it from life. Else **stat 8 < cost: returns 0, nothing taken**; else FUN_006272b0(stat 8, -cost), returns 1 | V read (0x56bfe0..0x56c109) |
| `FUN_0056f640` | after a srvstfunc that succeeded: pays (FUN_0056bfe0) unless usemanaondo or the flag arg | V read (0x56f78c..0x56f7aa) |
| `FUN_0056f7f0` (skill do; args skill, level, pay flag, ...) | usemanaondo: FUN_0056c160 again first (0: no do). srvdofunc (table 0x7322b0), srvmissile (FUN_0056ecb0 / FUN_0056ee90); nothing made and the do returned 0: return without paying (0x56fa3d). With the pay flag: **no srvstfunc, or usemanaondo: FUN_0056bfe0, its answer ignored** (0x56fa57..0x56fa7a); decquant: FUN_0056c3f0 (ammo); then the delay (below) | V read |
| `FUN_0056fc50` | the do at the action frame: FUN_0056f7f0(skill, level, pay 1, 0, 0), then FUN_0056cda0 | V read |
| `FUN_00580460` (player mode event) | event 1 or 2 of the mode's frame table runs FUN_0056fc50 unless the unit's "missile made" flag (+0xc4 bit 6, set by the do, cleared at the mode start 0x57fed8) is on | V read |
| `FUN_0057fe90` (player mode start) | clears that flag and runs FUN_0056faf0 -> FUN_0056f640(skill, level, 0) | V read |
| when the action frame comes | SC at 0 FCR: 13 ticks, the event around tick 8-9 (docs/research/cast-trace-gameexe.md, SKMISS) | S |

What follows (I, from the reads): mana leaves at the command (start) for the four start-function Sorceress skills, at the action frame for all the others; a cast whose mode is replaced before the event (get-hit, knockback, death) runs no do and pays nothing; a cast that passed the check but finds less mana at the action frame (e.g. Energy Shield drained it in between) is cast and pays nothing. Under blood_mana (a monster curse, skill 310, D) the cast is paid from life.

## 3. Differential (V run, game.exe natively vs the real d2d build)

Oracle `tools/emu/sk_mana.py`: FUN_0056c160 then FUN_0056bfe0 natively, for a player whose current skill is the case's skill at the case's level with the case's mana (256ths); output `id check paid` (paid 0 when the check fails: the start refuses). d2d side `tools/emu/sk_mana_test.cpp` (upstream `tests/game/test_sk_mana.cpp`, local only): d2d's rule on every cast path, `lvl > 0 && mana >= cost_of` then `pay` = the REAL `rules::mana_cost` on the loaded patched Skills.txt.

Cases (`gen 40000 1`, 130,840): every one of the 357 Skills rows x levels 1..60 x mana at 0, formula - 1 / = / + 1, minmana x 256 - 1 / = / + 1 (90,840), plus 40,000 random (levels 1..99, mana 0..0x20000). Build: upstream main 72132b99 on the worker host.

| result | cases |
|---|---|
| equal (check and paid) | 130,049 (37,388 of them a payment, equal to the 256th) |
| game.exe casts and pays nothing, d2d refuses (mana >= formula, < minmana x 256) | 549 |
| game.exe casts and pays nothing, d2d pays 256 (Might: mana 0, lvlmana 0, minmana 1) | 242 |
| by skill (all 791) | Might 98: 422; Summon Resist 89 (44 - 3 / lvl): 91; Teleport 54 (24 - 1 / lvl, from level 25): 74; Dopplezon 28: 71; Bone Prison 88: 67; Guided Arrow 22: 66 |

Every mismatch is the floor (D2); no row differs in the formula. For the plan's skills (Fire Bolt, Charged Bolt, Ice Bolt, Frozen Armor, Static Field, Telekinesis, Frost Nova, Lightning, Chain Lightning, Blizzard, Energy Shield) d2d's check and cost equal game.exe at every level 1..60; Teleport from level 25 only, with less than 1 mana point.

Not covered by this harness: the timing (D1, D3) and the start-paying rows' interaction with a failed start function; a stateful drill below covers them.

## 4. Regen, potions, leech, delay

| Address | Rule | Tag |
|---|---|---|
| `FUN_005806f0` (each frame, from the player's event 3 callback FUN_00580810) | unless state 0x55 (`nomanaregen`, D): base = `max(maxmana (FUN_00625d60) / (CharStats ManaRegen +0x3a x 25, 7500 when 0), 1)` (ManaRegen 120 for every class, D: 3000 frames), then `MulDiv(base, 100 + stat 27 manarecoverybonus, 100)` (FUN_00483360); then + stat 26 manarecovery (also under nomanaregen); a full globe ends state 0x6a (`manapot`); capped at max - mana, floored at -mana | V read (0x5806f0..0x5807f0), matching upstream combat.md (S) |
| d2d | `rules::mana_per_frame` + `Fight::apply_regen` | S; d2d-fidelity 1a measured `floor(256 max / 3000) x 25 / 256` per s on d2d (V-d2d) |
| what it means | 256ths a frame are integers: max 35 -> 2 a frame (0.195 / s, 179 s to fill), 45 -> 3 (0.293 / s, 153 s), 100 -> 8 (0.78 / s, 128 s), 250 -> 21 (2.05 / s, 122 s). The bonus multiplies the truncated base: Warmth level 1 (+30 %, matrix) adds nothing while the base is 2 or 3 (max mana under 47) | I (arithmetic on V) |
| potions `FUN_005be3f0` | misc.txt stat `manarecovery`: mp1 20, mp2 40, mp3 80, mp4 150, mp5 250 points over `len` 128 frames (5.12 s), joined with what is left of a running one | D (misc.txt via mpq-cat) + S (upstream combat.md) |
| `FUN_0062a620` | the class factor on a mana potion: Amazon, Paladin, Assassin x1.5 (`v + (v >> 1)`); **Sorceress**, Necromancer, Druid **x2**; Barbarian x1 (byte table 0x62a668, jump table 0x62a65c) | V read |
| potion doubling | energy > 0: `rand(energy) >> 1` percent chance to double | S (upstream combat.md) |
| leech | mana stat 62 (manadrainmindam) on a physical hit, x the monster's Drain; d2d `blow.mana = phys x mana_steal x drain / 10000` (combat.hpp:517) | S; not read here (U) |
| skill delay `FUN_0056ef90` | players only: state 0x79 (`skilldelay`, D) until frame + the row's delay calc (+0x190, FUN_00646ca0); Sorceress: Fire Wall 35, Meteor 30, Blizzard 45, Hydra 40, Frozen Orb 25 (D) | V read; d2d loads `delay` but uses it only for pets (`fight/pets.cpp:580`), not the player (S, matrix K7) |
| cast length / FCR | docs/research/cast-trace-gameexe.md (V; PR #128) | S |

## 5. d2d's version (S, upstream main 72132b99)

- `rules::mana_cost` (components/rules/skills.hpp:518): `lvl < 1: 0`, else `max(formula, minmana x 256)` (comment cites FUN_0056c160, which has no floor).
- `Fight::cast` (fight/tick.cpp:81), `cast_missile` (fight/missiles.cpp:218), `cast_summon` (fight/summons.cpp:52), `start_swing` (fight/strike.cpp:223): `cost_of` = 0 from a charge, else `mana_cost`; `mana < cost`: `need_mana` voice and refuse (or a plain swing with AttackNoMana); else **`pay` at once**, then the mode.
- Regen `rules::mana_per_frame` (rules.hpp:880), potions `rules::potion_amount / potion_rate`: equal (S, fidelity 1a).
- No blood_mana life payment, no player skill delay, no startmana check (S, by grep).

## 6. The bot's model (our repo)

- `internal/data/skills.go` `ManaCost`: the same formula with the minmana floor, as a float; `internal/combat/skills.go` `Cost = ceil(ManaCost)`; `PickAttack` / `resource.AttackChoice` refuse a skill whose Cost > mana.
- `internal/resource/types.go`: `manaFullSecs = 100` (U), `ManaPerSec = max / 100 x (1 + bonus / 100)`; `RestSecs` and the sim (`sim.go`: mana potion +40 at once, rejuv +60 %) use it.
- `internal/resource/hell.go` `ManaBurnTracker.Observe(now, mana, max, spent)`: the caller passes the cast's cost as spent when it sends the command (`internal/combat/fight.go`, `castfirst.go`).
- `internal/d2dcombat/meter.go`: ManaSpent = the sum of mana drops between samples (timing-neutral).

| id | finding | tag | impact |
|---|---|---|---|
| B1 | ManaBurnTracker books a cast at the command; game.exe takes it at the action frame for all but Inferno / Telekinesis / Thunder Storm / Hydra, and not at all for a cast cut off before it (SKMISS B3, now with the start-paying list) | V read + S | a phantom one-cast spike then a phantom gain; an interrupted cast books spending that never happened; only near the Hell flee / chicken lines |
| B2 | regen model 100 s to fill, linear: game.exe 120 s at large pools and slower on small ones (integer 256ths a frame): max 35 = 179 s, 45 = 153 s, 100 = 128 s; the bonus multiplies the truncated base (Warmth 1 adds 0 under 47 max mana) | V read + I | `RestSecs` 1.2x-1.8x too short early: rests end before the mana is there, the sim (Blood Moor clvl 8, max 45) overstates mana by ~1.5x. Candidate: `ManaPerSec = floor(256 max / 3000) x (100 + bonus) / 100 x 25 / 256` + flat |
| B3 | sim drinks a mana potion as +40 at once; game.exe: 20 / 40 / 80 / 150 / 250 x2 for the Sorceress over 5.12 s, joining a running one | D + V read | sim only; a re-drink inside 5 s wastes less than the bot assumes is fine, an instant check right after a drink sees part of it |
| B4 | the gate uses `ceil` and the minmana floor: stricter than game.exe (which casts with mana >= formula and then may pay nothing) | V run | harmless (Teleport 25+ under 1 point) |
| B5 | no model of blood_mana (casts paid from life; a monster curse, skill 310) | V read + D | Act 5 / Hell only (I); a caster could kill itself casting; out of the early plan |
| B6 | the bot's skill cooldown (`SkillSpec.Delay`, internal/combat/skills.go) matches game.exe's per-player skilldelay state; d2d does not enforce it for players (matrix K7) | V read + S | on d2d the bot could cast Blizzard faster than game.exe allows: a d2d-only XP/h boost for the END table |

## 7. Deviations (d2d vs game.exe)

| id | what | d2d | game.exe | tag | differential under tools/emu | bot impact |
|---|---|---|---|---|---|---|
| D1 | when mana is paid (non-start skills) | at the command | at the action frame (FUN_0056f7f0) | V read / S | none (stateful); drill M1 | the mana bar drops ~8 ticks earlier on d2d |
| D2 | floor in the check and the paid amount | `mana >= max(formula, minmana x 256)` and pays that; Might costs 1 point | check `mana >= formula`; pays the floor only if mana covers it, else nothing (the cast still happens); mana 0 / lvlmana 0 rows pay nothing | V run on the real build | `sk_mana.py` + `sk_mana_test.cpp`: 791 of 130,840, all explained | none for the plan (Teleport 25+ under 1 point) |
| D3 | interrupted cast | paid | free | V read / S | none; drill M2 | d2d costs mana game.exe would not under get-hits |
| D4 | mana drops between start and action frame | already paid | cast, nothing paid | V read / S | none | only with Energy Shield or mana drain |
| D5 | blood_mana | not modelled | from life (FUN_005d2b60) | V read / S | none | late game |
| D6 | player skill delay (state 0x79) | not enforced | enforced (FUN_0056ef90) | V read / S (matrix K7) | none | END table on d2d only |
| D7 | startmana (Inferno, Arctic Blast, Blade Fury) | no check | client status 1 under startmana x 256 (FUN_006440f0) | V read / S | none | not in the plan |

PR candidate: not built. D2 is the only one with a differential and its magnitude is below one mana point outside Might; D1 / D3 / D4 need d2d to split the check (command) from the payment (the strike at the action frame: `Fight::strike` already runs `fire` there), which is the change worth proposing once the drill confirms it live.

## 8. Drill recipe (live, a throwaway Sorceress on one instance, offline)

Verbs FORKDBG is adding: `debug states` (list a unit's states), `debug mon set` (pin a monster's life / damage / resistances), `debug cast` (cast by id at a unit or point, no input), `debug mana full`. Existing: `debug stat`, `debug points`, `debug spawn`, `step N`, `dump`.

- M1 (D1, pay timing): `debug points 36 1`, `debug stat 8` = 20 x 256. `debug cast 36` at an empty point, then `step 1` x 14 with `dump` each: game.exe's mana drops 640 on the action-frame tick (8-9 of 13); d2d before a fix on tick 0. Repeat with Telekinesis 43 (start-paying): both drop at tick 0.
- M2 (D3, interrupted cast): `debug spawn` a Fallen next to the caster, `debug mon set` its damage above max life / 12 (a get-hit), cast Fire Bolt; when the get-hit lands before the action frame: game.exe no missile and mana unchanged; d2d mana spent.
- M3 (D2, the floor): Teleport level 25 (`debug points 54 25`), `debug stat 8` = 128 (half a point): game.exe teleports and mana stays 128; d2d refuses (need-mana voice). Also mana 0: game.exe teleports (formula 0 <= 0).
- M4 (regen, B2): `debug stat 9` = 45 x 256, `debug stat 8` 0, `step 750` (30 s): game.exe +2,250 256ths (3 a frame = 8.8 points); the bot's model says 13.5 points. With Warmth 1 (`debug points 37 1`): unchanged (3 x 130 / 100 = 3).
- M5 (potions, B3): drink mp1 on a Sorceress: +40 points spread over 128 frames (`dump` every 16 steps: +5 each), state `manapot` in `debug states` until the globe is full or the frames end.
- M6 (D6, delay): Blizzard (`debug points 59 1`), `debug cast 59` twice in a row: game.exe refuses the second until 45 frames after the first's action frame (`debug states` shows `skilldelay`); d2d casts both.

Expected in game.exe: as above. Expected on d2d today: M1, M2, M3, M6 differ; M4, M5 equal (S).

## 9. Matrix cell updates (docs/research/sorc-skill-matrix.md; not edited by this card)

| Row | Cell | New value |
|---|---|---|
| K5 (mana cost formula) | tag | V run: FUN_0056c160 / FUN_0056bfe0 natively vs d2d's `mana_cost`, 130,840 cases, equal except the minmana floor (D2: the check has none, the payment refuses silently and the cast is free) |
| every skill row | (a) mana | V (formula equal at levels 1..60); paid at the action frame except Telekinesis / Hydra / Thunder Storm / Inferno (start) |
| Teleport | (a) mana | from level 25 the formula is <= 0: castable with any mana, pays 1 point when it has it, else nothing (V run) |
| Warmth | regen | V read: MulDiv on the truncated per-frame base; level 1 adds nothing under 47 max mana |
| K2 (PR #128 in the lab binary) | tag | V by git for the branches (lab6, lab-port-148, lab-debug-skills have it; `lab` not); the running binary U |
| K7 (skill delay) | tag | V read: FUN_0056ef90, state 0x79, players only; d2d does not enforce it (D6) |

## 10. Not traced (U)

- The four start functions' own refusals (Telekinesis, Thunder Storm, Hydra, Inferno): what a failed start does with the mana (none paid: the payment sits after a non-zero start result, V read; what the client shows: U).
- Leech's server path (address not read here).
- Monster mana (non-players skip both check and payment: V read; monster casters have no mana in 1.14d: I).
- The exact action-frame tick for each cast mode on the server (SKMISS section 10).
