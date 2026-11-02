# Paladin skill rules: Smite, Zeal, Sacrifice, Holy Bolt, Blessed Hammer, the aura set; game.exe 1.14d vs d2d vs the bot (decomp queue: Paladin skills)

> status: card palskill, 2026-11-01, offline (no game, no d2d slot, no Go change). Follows docs/classes/paladin-probe.md, paladin-act1-recon.md (section 6 rows 1 & 5), and the SKMISS / SKMANA traces (docs/research/sk-missile-trace.md, sk-mana-trace.md). Updated for Issue #3 (Smite, Zeal and Sacrifice dispatch in game.exe 1.14d vs d2d's skill_built).
> Authored by an AI coding assistant.

Tags: **V read** disassembled in game.exe 1.14d (addresses and functions given, no disassembly or decompiled code reproduced); **V run** run natively in game.exe under the emulator by this agent (script named) and, where a d2d column exists, compared line for line with the real d2d build (upstream main, patched 1.14d tables, linked against scratch builds); **D** from the patched 1.14d tables (Skills.txt, Missiles.txt, States.txt, MonStats.txt, Items.txt read with mpq reader from the patch stack); **S** d2d source or an upstream doc / earlier trace of ours (named); **I** inference; **U** not verified. A rule known only from reading is **V read** and is not offered as proof of anything (the model split of 2026-11-01: only V run numbers may carry a PR).

## Summary

- **Smite (V read + V run).** Do 150 (FUN_005ce9f0) requires a shield. Base damage is shield base min / max damage (Items.txt +0xfe/+0xff), plus stat 17/18 min / max damage added directly by Holy Shield (state 0x65, FUN_00647bc0 / FUN_00647d00). Damage percent bonus is calc1 (ln34 = 15 + 15*(lvl-1)%), plus stat 25 (off-weapon ED), plus stat 0 (Strength) via ItemGetStrBonus (FUN_00629860 = 100, so +1% damage per point of Strength). Stun length is calc2 (min(250, ln12) = min(250, 15 + 5*(lvl-1)) frames; hard-capped at 250 frames = 10.0 seconds). Knockback is set via ResultFlags = 8. Smite **unconditionally hits** because FUN_005ce9f0 forces bit 0 of the hit flags after FUN_0057ec10 (or ax, 1): shield block, weapon block, defense and attack rating are completely bypassed (a blocking target may trigger a block animation, but still takes full damage, stun, and knockback). Mana cost is 2 mana (manashift 8, mana 2, minmana 1, paid in do). A held button repeats Smite attack animation (mode S1) each cycle, paying 2 mana per swing.
- **Zeal (V read + V run).** Start 37 (FUN_005daf40) and Do 13 (FUN_005dbc60). Hit count is calc1 (min(par5 + lvl - 1, par6) with par5=2, par6=5): 2 hits at level 1, 3 at level 2, 4 at level 3, and 5 hits at level 4+ (cap 5). Damage percent bonus is calc2 (((lvl < 5) ? 0 : ((lvl-4) * 6)) + Sacrifice.blvl * 12%). Attack rating bonus is ToHit 10, LevToHit 10 (10 + 10*(lvl-1)%). Mana cost is 2 mana (manashift 8, mana 2, minmana 1), **paid once at srvstfunc** when the Zeal sequence starts; individual swings 2..5 cost 0 mana. When a target dies or leaves melee reach during the sequence (FUN_00622c40 fails), Do 13 searches within melee reach for another enemy using filter 0x20003 (FUN_0056bd10). In game.exe, Zeal uses frame rollback (Param2 = 100, FUN_004efc20) to execute rapid consecutive strikes. A held button executes a fresh 2..5 hit Zeal sequence as soon as the previous sequence finishes, paying 2 mana per sequence.
- **Sacrifice (V read + V run).** Start 29 (FUN_005ce790) and Do 64 (FUN_005ce8e0). Damage percent bonus is calc1 (180 + 15*(lvl-1)% + Redemption.blvl * 15% + Fanaticism.blvl * 5%). Attack rating bonus is ToHit 20, LevToHit 7 (20 + 7*(lvl-1)%). Mana cost is 0 mana (mana 0, lvlmana 0, minmana 0). Self-damage is calc2 (par3 = 8%): Do 64 takes the physical damage dealt, **caps it at the target's current hitpoints** (min(phys_damage, target_hp)), and inflicts 8% of that capped value onto the caster as direct internal damage (flag 0x1000). If self-damage exceeds current player life, the player dies. A held button repeats Sacrifice attack animation (mode A1, seqnum 2) each cycle, taking 8% life cost on each landed hit.
- **d2d implementation in `skill_built` (S).** Contrary to early concerns that Smite, Zeal, and Sacrifice might fall back to plain attack, d2d's `skill_built` (components/game/fight.hpp:62-75) explicitly implements all three:
  - Smite: `skill.srvdofunc == ServerDoFunction::kSmite` -> true.
  - Zeal: `runs(skill, ServerStartFunction::kRepeatedHit, ServerDoFunction::kRepeatedHit)` -> true.
  - Sacrifice: `runs(skill, ServerStartFunction::kSacrifice, ServerDoFunction::kSacrifice)` -> true.
  None of Smite, Zeal, or Sacrifice falls back to plain attack in d2d.
- **Holy Bolt (V read + V run).** The missile `holybolt` (id 55, Vel 20, Range 50, CollideFriend, LastCollide, Holy) has server hit function 7 (FUN_005a9fb0) with sHitPar1 1 (heals allies) and sHitPar2 1 (undead only). A hit unit that is the caster's own pet / merc (FUN_005542c0) or an ally in the caster's party (FUN_00554d20) is **healed** by a roll of calc1..calc2 (ln12 / ln34, Prayer synergy) and the bolt ends; every other unit is **damaged only if it is undead** (return 3), and **passed through with no damage** (return 4) if it is not. Players are always passed. The undead test (FUN_0063e990: MonStats +0xd & masks 3 or 4) and the demon test (FUN_0063e940: mask 5) agree with d2d's `undead` / `demon` flags on **all 734 monster ids** (V run, 0 differ; 123 undead, 185 demon). d2d implements the filter and not the heal (its source says so).
- **Blessed Hammer (V read + V run).** Do 73 (FUN_005d0040) pays at the action frame (no start function), makes one `blessedhammer` missile (92) from the caster, sets path type 14 (FUN_00648cf0, FUN_00649970) and scales the missile's magic min / max (stats 0x34 / 0x35) by **half of the caster's Concentration damage %** (FUN_006461d0: state 42 concentration, stat 25, x Param1 4 / 8, toward zero: Concentration 1 +30 %, 10 +97 %, 20 +172 %: V run). A hit on an undead or a demon adds **+50 % of the rolled magic damage** (pSrvDmgFunc 5, FUN_005ad2f0, dParam1 / dParam2 = 50). The flight is a **spiral out of the caster's own subtile**, not a bolt at the target: V run radius 1.6 subtiles at frame 10, 4.5 at 25, 6.5 at 50, 10.4 at 75, 10.8 at 100, 13.0 at 120 (Range 120 frames), 2.5 to 2.9 turns (about 0.57 a second on average), the same at every level and every aim (the aim only sets the phase). A stationary Quill Rat standing in its orbit was hit at distances 1, 2, 4, 5, 6, 8, 9, 10 and **not** at 3, 7, 11 to 14 (V run, once each).
- **Damage numbers (V run, 0 differ of 128,640 cases).** Element min / max with synergies for Holy Bolt, Blessed Hammer, Holy Fire, Holy Freeze, Holy Shock, Fist of the Heavens (8,640 cases), plus Smite / Zeal / Sacrifice ED, AR, stun duration, hit counts, and self-damage scaling (120,000 cases): game.exe equals d2d / formula verification with 0 mismatches.
- **Auras (V read + V run, 0 differ of 21,760 cases).** Do 65 (FUN_005cf010, friendly) and 66 / 81 (FUN_005cf3a0 / FUN_005d0920, Holy Fire / Shock / Freeze, Sanctuary, Conviction) read six aurastat ids and calcs, the five passive stats, the range (ln12: Might / Prayer / Defiance / Concentration 16 + 2 per level, Vigor 16, 19, 22 .. 73 at 20, Holy Fire / Freeze / Shock 6 + 1 per level) and the period; game.exe's calc machine equals d2d's `eval_calc` on every aura row and slot. Paladin perdelay is 50 frames (2 s).
- **Deviations d2d vs game.exe:**
  - D1 Holy Bolt's heal (absent in d2d).
  - D2 Concentration's 50 % share of Blessed Hammer's damage (absent in d2d).
  - D3 Blessed Hammer's +50 % against undead and demons (absent in d2d).
  - D4 Blessed Hammer's path (d2d uses mathematical Archimedean spiral model).
  - D5 Aura per-pulse mana and Might command floor (Might charged 1 mana in d2d).
  - D6 Aura reach to merc/party (d2d player only).
  - D7 Repeat hits of one hammer (d2d limits 1 hit per monster per hammer).
  - D8 Smite blockable in d2d: `combat.hpp:499` checks `target.block` without `!swing.smite`, allowing blocking monsters to nullify Smite damage; game.exe forces hit bit (or ax, 1), so Smite is unblockable.
  - D9 Zeal missing frame rollback in d2d: `strike.cpp:901` waits for full `pmode_until` (~14-16 frames) between every Zeal strike; game.exe uses 100% frame rollback (Param2 = 100, FUN_004efc20) to strike rapidly (~4 frames).

## 1. Smite, Zeal and Sacrifice Dispatch (game.exe 1.14d)

| Field / Mechanic | Smite (Id 97) | Zeal (Id 106) | Sacrifice (Id 96) | Tag |
|---|---|---|---|---|
| srvstfunc | None | **37** (`FUN_005daf40`) | **29** (`FUN_005ce790`) | D, V read |
| srvdofunc | **150** (`FUN_005ce9f0`) | **13** (`FUN_005dbc60`) | **64** (`FUN_005ce8e0`) | D, V read |
| anim / seqnum | S1 / None | A1 / None | A1 / **2** | D, V read |
| Base weapon / shield | Shield equipped (FUN_0063c8f0); reads shield min/max from Items.txt +0xfe/+0xff | Mainhand weapon physical damage | Mainhand weapon physical damage (SrcDam 128 = 100%) | D, V read |
| Holy Shield synergy | Active state 0x65 adds Holy Shield stat 17/18 min/max to base shield damage | None | None | V read, V run |
| Hit count | 1 hit per action cycle | **calc1**: `min((par5 + lvl - 1), par6)` (par5=2, par6=5): L1=2, L2=3, L3=4, L4+=5 | 1 hit per action cycle | D, V read, V run |
| Damage % bonus | **calc1**: `ln34` = `15 + 15*(lvl-1)%` | **calc2**: `((lvl < 5) ? 0 : ((lvl-4) * 6)) + Sacrifice.blvl * 12%` | **calc1**: `(180 + 15*(lvl-1)) + Redemption.blvl * 15% + Fanaticism.blvl * 5%` | D, V read, V run |
| Stat bonuses to damage | stat 25 (off-weapon ED) + stat 0 (Strength) via ItemGetStrBonus = 100 (+1% per Str) | stat 25 (ED) + stat 0 (Strength) / stat 2 (Dexterity) per weapon type | stat 25 (ED) + stat 0 (Strength) / stat 2 (Dexterity) per weapon type | V read |
| Attack Rating bonus | None (always hits, bypasses AR check) | `ToHit 10, LevToHit 10` = `10 + 10*(lvl-1)%` | `ToHit 20, LevToHit 7` = `20 + 7*(lvl-1)%` | D, V read, V run |
| Hit check & Block bypass | Calls FUN_0057ec10 then executes **`or ax, 1`**: ignores defense, shield block, and weapon block | Standard melee hit check FUN_0057ec10 | Standard melee hit check FUN_0057ec10 | V read |
| Stun duration | **calc2**: `min(250, ln12)` = `min(250, 15 + 5*(lvl-1))` frames; cap 250 frames (10.0 s) | None | None | D, V read, V run |
| Knockback | ResultFlags = 8 (bit 3 set) | None | None | D, V read |
| Mana cost | 2 mana (`manashift 8, mana 2, minmana 1`), paid in srvdofunc 150 | 2 mana (`manashift 8, mana 2, minmana 1`), **paid once at srvstfunc 37** | 0 mana (`mana 0, lvlmana 0, minmana 0`) | D, V read, V run |
| Life cost / Self-damage | None | None | **calc2**: `par3` = 8% of `min(phys_damage_dealt, target_current_hp)`. Direct internal damage (0x1000 flag); lethal | D, V read, V run |
| Target reacquisition | Target locked | If current target dies or leaves reach (FUN_00622c40), searches reach for new target (0x20003 filter, FUN_0056bd10) | Target locked | V read |
| Animation & Rollback | S1 full cycle | A1 with Param2 = 100 (% frame rollback via FUN_004efc20 / FUN_0056e210) | A1 full cycle | D, V read |
| Held button behavior | Repeats Smite attack each animation cycle (S1), paying 2 mana per swing | Repeats full Zeal sequence (2..5 hits) upon completion if target in reach, paying 2 mana per sequence | Repeats Sacrifice attack each animation cycle (A1), paying 8% life of damage dealt per hit | V read |

### Smite Dispatch Details (FUN_005ce9f0, srvdofunc 150)
1. **Shield Requirement (V read):** Calls FUN_0063c8f0 on caster; if no shield is equipped, returns 0 immediately. Reads base shield damage from `Items.txt` row of the shield: min damage at offset `+0xfe`, max damage at `+0xff`.
2. **Holy Shield Bonus (V read + V run):** Checks if state 0x65 (`holy_shield`, D) is active on the caster (FUN_00639df0). If active, calls FUN_00647bc0 and FUN_00647d00 on the Holy Shield statlist to retrieve min and max damage bonus, and adds them directly to the shield's base damage.
3. **Damage Scaling (V read + V run):** Calls FUN_00646ca0 with calc1 (`ln34 = 15 + 15*(lvl-1)%`). Passes shield item, caster, min/max base damage, and calc1 to FUN_0057b420. FUN_0057b420 queries ItemGetStrBonus (FUN_00629860), which returns 100 for shields, adding `Strength * 100 / 100 = +1%` per point of Strength. Off-weapon ED (`stat 25`) is added. Min damage is clamped to -90% floor if negative.
4. **Stun & Knockback (V read + V run):** Evaluates calc2 (`min(250, ln12)` = `min(250, 15 + 5*(lvl-1))` frames), writing the stun length to the blow packet offset `+0x44`. ResultFlags (Skills.txt +0x130 = 8) is merged into packet flags (bit 3 = knockback).
5. **Always-Hit Rule (V read):** FUN_005ce9f0 calls FUN_0057ec10 to perform standard defense and block checks, but immediately follows with `or ax, 1`. Because bit 0 indicates hit landed, Smite cannot miss. Even if a target rolls a block (bit 4), the hit still lands and delivers full damage, stun, and knockback.

### Zeal Dispatch Details (FUN_005daf40 / FUN_005dbc60, srvstfunc 37 / srvdofunc 13)
1. **Sequence Initialization (V read + V run):** In `srvstfunc 37` (FUN_005daf40), evaluates calc1 (`min((par5 + lvl - 1), par6)` with par5=2, par6=5). Sets remaining hit count on the skill node via FUN_0044560: level 1 has 2 hits, level 2 has 3, level 3 has 4, level 4+ has 5 hits. Pays 2 mana (FUN_00644b10).
2. **Strike Execution (V read + V run):** In `srvdofunc 13` (FUN_005dbc60), verifies the target is alive and in melee reach (FUN_00622c40). If the target died or left reach, invokes target finder FUN_0056bd10 with enemy filter 0x20003 within melee reach (FUN_0056e510) to switch to another target.
3. **Damage & AR Bonus (V read + V run):** Attack rating bonus is computed by FUN_006449f0: `ToHit + (lvl - 1) * LevToHit` = `10 + 10*(lvl-1)%`. Damage bonus is calc2: `((lvl < 5) ? 0 : ((lvl-4) * 6)) + Sacrifice.blvl * 12%`.
4. **Rapid Chaining via Rollback (V read):** After each strike, decrements remaining hit count. If hits remain (> 0), calls FUN_004efc20 (reading Param2 = 100, frame rollback) and FUN_0056e210 to chain the next hit rapidly without replaying the full animation recovery.

### Sacrifice Dispatch Details (FUN_005ce790 / FUN_005ce8e0, srvstfunc 29 / srvdofunc 64)
1. **Damage & AR Bonus (V read + V run):** Attack rating bonus is `ToHit + (lvl - 1) * LevToHit` = `20 + 7*(lvl-1)%`. Damage bonus is calc1: `180 + 15*(lvl-1)% + Redemption.blvl * 15% + Fanaticism.blvl * 5%`. Mana cost is 0.
2. **Self-Damage Cap at Target HP (V read + V run):** In `srvdofunc 64` (FUN_005ce8e0), physical damage dealt is compared to target current hitpoints (`stat 6`, FUN_00625480). If damage dealt exceeds target HP, damage is clamped to target HP: `capped_damage = min(phys_damage, target_hp)`.
3. **Self-Damage Infliction (V read + V run):** Evaluates calc2 (Param3 = 8%). Calculates `8% * capped_damage` via MulDiv (FUN_00483360). Inflicts this damage on the caster with internal direct damage flag `0x1000` (FUN_0057c6c0 / FUN_0057cee0). This damage can reduce player HP to 0 and kill the player.

---

## 2. Holy Bolt

| Address | Rule | Tag |
|---|---|---|
| Skills 101 (D) | `srvmissile holybolt`, no srvstfunc / srvdofunc, anim SC, `range none`, TargetPet 1, TargetAlly 1, reqlevel 6, mana 32 / lvlmana 1 / manashift 4 / minmana 1, EType mag EMin 8 (levels +8 +10 +13 +16 +20 brackets) EMax 16 (8, 11, 15, 18, 23), HitShift 8, EDmgSymPerCalc `(Blessed Hammer + Fist of the Heavens) x par8 (50)`, calc1 / calc2 `ln12 / ln34 x (100 + Prayer x par7 (15)) / 100` with Param1..4 = 1, 2, 6, 4 (min heal 1 + 2 per level, max 6 + 4 per level) | D |
| Missiles 55 (D) | pCltDoFunc 1, pSrvDoFunc 1, **pSrvHitFunc 7**, sHitPar1 1 ("heals allies?"), sHitPar2 1 ("0 all, 1 undead, 2 demons"), Vel 20 = MaxVel, Range 50, Size 1, CollideType 3, CollideKill 1, **CollideFriend 1**, LastCollide 1, Holy 1, ProgOverlay healing, no pSrvDmgFunc | D |
| mana | no srvstfunc, so the mana is **paid in the do** at the action frame (FUN_0056f7f0) after the missile; cost `(32 + lvl - 1) << 4` in 256ths = 2.00 mana at level 1, 2.56 at 10, 3.19 at 20 (paid 512 / 656 / 816 256ths) | V run (sk_mana.py, 12,000 cases below) + V read (SKMANA) |
| `FUN_005a9fb0` (ecx game, edx missile, [esp+4] unit hit; returns the hit flags: bit 1 ends the missile, bit 2 damages the unit, 4 passes) | no unit: 1. `owner = FUN_00552fd0(missile)`. sHitPar1 != 0 and an owner: **friendly** when `FUN_005542c0(owner, unit, skill)` (Skills flag +7 bit 1 = TargetPet, and the unit's root owner id (FUN_0058f090) is the owner's id: the caster's pets / merc) **or** `FUN_00554d20(owner, unit, skill)` (+7 bit 2 = TargetAlly, caster a player, the unit's root unit has the caster's party id, `FUN_00554630`, not 0xffff): then heal = `FUN_00646ca0(owner, calc1) << 8` .. `calc2 << 8`, rolled `min + rand(max - min)` on the missile's own seed (missile +0x20), <= 0 falls back to the missile's physical roll (FUN_005a8910, stats 0x15 / 0x16); stat 6 (life) += heal, capped at max life (FUN_00625d10); overlay Skills +0x36 when > 0; return 1 | V read |
| same, not friendly (hostile, or no owner) | a **player** (type 0): return 3 if sHitPar2 == 0, else **4** (pass). A **monster** (type 1): sHitPar2 0: 3; 1: **3 if FUN_0063e990 (undead) else 4**; 2: 3 if FUN_0063e940 (demon) else 4. Any other unit type: 4 | V read; the monster branch V run below |
| `FUN_0063e990` / `FUN_0063e940` | MonStats row (record 0x1a8, table +0xa78) byte +0xd: undead = `& (mask[3] | mask[4])` (masks at 0x6ce268 + 4 n = 1 << n: 8 | 0x10, the lUndead and hUndead columns), demon = `& mask[5]` (0x20) | V read + V run |

**Q1. Does Holy Bolt hit non-undead monsters?** No: it passes through them with no damage and keeps flying (return 4). **V read** (FUN_005a9fb0 0x5aa112..0x5aa16c); the filter's input, the undead / demon flag, **V run** over every monster id: 123 undead and 185 demon ids match d2d byte-for-byte.

**Q2. Undead conversion versus heal?** There is no conversion. A friendly unit is healed, anything hostile is damaged (undead) or passed (the rest). **V read**.

**Q3. Mana?** Paid in the do, 2.0 to 3.2 mana per cast over levels 1..20.

**Q4. Damage?** Magic damage 8..16 at level 1, 84..94 at level 10, 196..220 at level 20 with no synergy; with synergies up to 4,116..4,620 at 20 (V run).

---

## 3. Blessed Hammer

| Address | Rule | Tag |
|---|---|---|
| Skills 112 (D) | `srvdofunc 73`, `srvmissilea blessedhammer`, cltdofunc 35, anim SC, reqlevel 18, reqskill Holy Bolt, mana 20 / lvlmana 1 / manashift 6 / minmana 1 (paid 1,280 256ths = 5.00 mana at level 1, 7.25 at 10, 9.75 at 20: V run), EType mag EMin 12 .. EMax 16, HitShift 8, **Param1 4 "Concentration influence in 8ths"**, Param8 14, EDmgSymPerCalc `(Vigor + Blessed Aim) x par8 (14)` | D |
| Missiles 92 (D) | pSrvDoFunc 1, **pSrvDmgFunc 5**, dParam1 50 ("bonus dmg % undead"), dParam2 50 ("demon"), Vel 18, MaxVel 30, Accel 250, Range 120, Size 1, CollideType 3, **LastCollide 1, Pierce 1**, no CollideKill, no NextHit, EType mag, 16 directions | D |
| `FUN_005d0040` (do 73) | row `FUN_005d3cf0` (`blessedhammer`); missile through FUN_0059fa30 (flags 0x20). Path `FUN_00648cf0(path, 14)` and `FUN_00649970(path, missile, 0)`. Then `infl = FUN_006461d0(caster, skill)`; when non-zero missile stats 0x34 / 0x35 scaled by `MulDiv(stat, 100 + infl, 100)`. Return 1 | V read |
| `FUN_006461d0` (unit, skill) | 0 unless unit has state 0x2a (`concentration`); else state's stat 0x19 (25, damagepercent) x Param1 (4) / 8, toward zero (+30% at L1, +172% at L20) | V read + **V run** |
| `FUN_005ad2f0` (pSrvDmgFunc 5) | if dParam1 > 0 and unit is undead, +50% of rolled magic damage added. If dParam2 > 0 and demon, +50% added | V read |
| flight | path type 14, spiral out of caster subtile: radius 1.6 at frame 10 to 13.0 at 120, 2.5 to 2.9 turns | V run |

**Q5. What does the missile do?** The hammer spirals out of the caster's own subtile, never homing toward the target; it hits what its orbit crosses (V run).

**Q6. Damage chain.** Min / max base times Concentration influence -> roll at hit -> undead/demon +50% -> magic resistance.

**Q7. Hammer synergy.** `(Vigor + Blessed Aim) x 14%`, Concentration enters through the influence rule.

---

## 4. The auras (V read + V run)

| Aura (id) | do | range (subtiles, ln12) | stat 1 (calc) | other stats | mana per pulse | Tag |
|---|---|---|---|---|---|---|
| Might (98) | 65 | 16 + 2 (L-1): 16, 18, 34, 54 at 1, 2, 10, 20 | damagepercent = 40 + 10 (L-1) | state `might` also on allies (auratargetstate) | none | V run + D |
| Prayer (99) | 65 | same | hitpoints = edns (512, 768, 2816, 6400 at 1, 2, 10, 20) | | (16 + 3 (L-1)) << 4 / 256: 1.0, 1.2, 2.7, 4.6 | V run + V read |
| Defiance (104) | 65 | same | skill_armor_percent = 70 + 10 (L-1) | | none | V run + D |
| Concentration (113) | 65 | same | damagepercent = 60 + 15 (L-1) | skill_concentration = 20 (Param5) | none | V run + D |
| Vigor (115) | 65 | 16, 19, 43, 73 | staminarecoverybonus = skill_staminapercent = 50 + 25 (L-1) | velocitypercent (dm56) = 13, 18, 36, 43 at 1, 2, 10, 20 | none | V run + D |
| Holy Fire (102) | 66 | 6 + (L-1): 6, 7, 15, 25 | none (no aurastat) | passivestat firemindam / firemaxdam = 6 / 18 .. 111 / 123; enemy burn per pulse | none | V run + D |
| Holy Freeze (114) | 81 | 6 + (L-1) | resistances -30..-54 on enemies | passive cold min/max 10/15 .. 185/190; slow | none | V run + D |
| Holy Shock (118) | 66 | 6 + (L-1) | | passive lightning min/max 1/60 .. 1/936 | none | V run + D |

- **Pulse (V read).** Scheduled every 50 frames (2 s) aligned to frame 1 mod 50 (FUN_0056cda0).
- **Mana (V read).** `(lvlmana x (lvl - 1) + mana) << manashift`. If > 0, sets state 85 (`nomanaregen`) and pays mana (Prayer only).

---

## 5. d2d's version (S, upstream main)

- **Smite, Zeal, Sacrifice in `skill_built` (S):** `components/game/fight.hpp:62-75` checks all three skills explicitly:
  - Smite: `skill.srvdofunc == ServerDoFunction::kSmite` -> true.
  - Zeal: `runs(skill, ServerStartFunction::kRepeatedHit, ServerDoFunction::kRepeatedHit)` -> true.
  - Sacrifice: `runs(skill, ServerStartFunction::kSacrifice, ServerDoFunction::kSacrifice)` -> true.
  None of them falls back to plain attack.
- **Smite in d2d (`strike.cpp:357-369`, `combat.hpp:498-505`):** Sets `swing.smite = true`, evaluates Holy Shield stat bonus from active self-states (`skill_phys`), calc0 ED %, calc1 stun ticks, and knockback (`result_flags & 8`). In `combat.hpp:498`, `!swing.smite` correctly skips the attack rating check. However, **`combat.hpp:499` checks `target.block > 0` without checking `!swing.smite`**: monsters with block can block Smite and take 0 damage, deviating from game.exe where Smite cannot be blocked (D8).
- **Zeal in d2d (`strike.cpp:244-248, 370-373, 901-906`):** Sets `actor->kicks_left = calc1 - 1`. Sets calc1 ED % on each hit. Next target selected via `next_target()`. However, `strike.cpp:901` waits for `now_ms >= actor->pmode_until` (the full ~14-16 frame A1 animation duration) between every strike of Zeal, missing the 100% frame rollback (Param2 = 100) that makes Zeal rapid in game.exe (D9).
- **Sacrifice in d2d (`strike.cpp:374-377, 603-610`):** Sets calc0 ED % and `swing.srcdam = skill->srcdam`. In line 605-610, calculates `dealt = min(resisted(blow.phys, res[0]), max(hp_before, 0))` and adds `(dealt << 8) * calc1 / 100` to `actor->self_hurt`. Matches game.exe's HP-capped self-damage rule.
- **Holy Bolt:** `Fight::skill_missile_hits` filters by `type_info.undead` / `demon`; does not heal friendly side.
- **Blessed Hammer:** Spirals out using `atan2` and radius formula; struck monsters hit once per missile; no Concentration influence; no +50% vs undead/demons.
- **Auras:** `Fight::aura_pulse` pulses every 50 frames; sums aurastats permanently on player; no mana charged, player only.

---

## 6. The bot's use (our repo; described, not changed)

- `internal/combat/classdata/profiles.json` paladin: band minLevel 18 `requires Blessed Hammer`: attack ladder Blessed Hammer, Zeal, Smite, Attack; aura ladder Prayer (restLifePct 70), Concentration, Holy Fire, Might; Zeal band 12; layout keys Might f1, Holy Fire f2, Concentration f3, Vigor f4, Holy Shield f5, Thorns f6, Resist Fire f7, Prayer f8.
- `internal/combat/classdata/skills.json`: Smite, Zeal (`hold`), Sacrifice are `melee` left-hand skills; Holy Bolt `bolt` left hand; Blessed Hammer `ground` left hand `hold`.
- `internal/plandata/build-paladin-normal30-v2.json`: Might x8, Sacrifice 1, Prayer 1, Smite 1, Zeal 4, then Holy Fire to 19, Holy Bolt, Blessed Hammer, Blessed Aim, Concentration to clvl 30.

| id | finding | tag | impact |
|---|---|---|---|
| B1 | Blessed Hammer spirals out of the caster, never homes; hit only at orbit crossings | V run | bot casts when monster is within ~10 subtiles and repeats |
| B2 | Concentration raises Blessed Hammer (+0.5 x its damage %); Might does not affect it | V read + V run | keep Concentration active while casting hammers |
| B3 | Resting Prayer costs mana per pulse and sets state 85 (no mana regen); d2d charges nothing | V read + D | d2d soak overstates rest efficiency |
| B4 | Holy Bolt damages only undead; passes through demons and living | V read + V run | guard any Holy Bolt use with undead check |
| B5 | Might costs 0 mana on game.exe; d2d charges minmana 1 point on command | V run | 1 mana spent on reselecting Might in d2d |
| B6 | Smite always hits and stuns up to 10 s; shield base damage + Holy Shield bonus + Strength % ED; costs 2 mana | V read + V run | excellent single-target lock against tough foes or bosses; independent of weapon damage/durability |
| B7 | Zeal hit count scales 2 -> 3 -> 4 -> 5 at skill levels 1, 2, 3, 4+; costs 2 mana once per sequence | V read + V run | putting 4 points into Zeal reaches the max 5 hits; holding the key repeats 5-hit sequences |
| B8 | Sacrifice costs 0 mana but inflicts 8% of damage dealt as self-damage (capped at foe HP); lethal | V read + V run | high risk at low character life; safe only when life leech or potion buffer exceeds self-hurt |

---

## 7. Deviations d2d vs game.exe (Paladin)

| id | rule | game.exe | d2d | tag | differential today / buildable |
|---|---|---|---|---|---|
| D1 | Holy Bolt heals friendly units | Caster's pets/merc and party allies get calc1..calc2 life (FUN_005a9fb0) | Not built | V read / S | None |
| D2 | Concentration raises Blessed Hammer | Missile magic min/max scaled by half of Concentration % (+30..+172%) | Absent | V read + V run | None |
| D3 | Blessed Hammer +50% vs undead/demons | FUN_005ad2f0 adds 50% of rolled magic damage | Absent | V read, S | None |
| D4 | Blessed Hammer path | Orbit from caster subtile, radius 13 at 120 frames | Mathematical model | V run, S | None |
| D5 | Aura per-pulse mana | Prayer costs mana per pulse, state 85; Might 0 | None per pulse; Might charged 1 on select | V read, V run | Might: SKMANA harness |
| D6 | Aura reach | Every ally within range (merc, pets, party) | Player only | V read, S | None |
| D7 | Repeat hits of one hammer | Orbit crossings can hit again (LastCollide) | Struck list: 1 hit per monster | V read | None |
| D8 | Smite unblockable | FUN_005ce9f0 forces hit bit (`or ax, 1`), ignoring monster block | `combat.hpp:499` checks `target.block > 0` without `!swing.smite`, allowing blocks | V read, S | Buildable: test `player_blow` with `swing.smite = true` against target with block > 0 |
| D9 | Zeal rapid frame rollback | Param2 = 100 (% frame rollback via FUN_004efc20); strikes chain rapidly (~4 frames) | Waits for full `pmode_until` (~14-16 frames) between every strike | V read, S | Buildable: measure Zeal sequence duration in `Fight::strike` |

---

## 8. Differentials (V run, native verification)

| Differential | Cases | game.exe vs d2d / model | Notes |
|---|---|---|---|
| undead / demon flags | 734 monster ids | **0 differ** (md5 0a050ae1...) | 123 undead, 185 demon |
| element min / max / length | 8,640 cases | **0 differ** | 6 Paladin elemental skills x levels 1..40 x synergies |
| aura calcs | 21,760 cases | **0 differ** | 22 aura skills x levels 1..60, range, perdelay, aurastats |
| mana costs | 12,000 cases | 400 differ (all Might, SKMANA D2) | Paladin skills 96..125 x levels 1..40; Smite=2, Zeal=2, Sacrifice=0 |
| Smite, Zeal, Sacrifice formulas (`scratch/verify_palskills.py`) | 120,000 cases | **0 differ** | Zeal hits (2..5), Zeal ED/AR, Sacrifice ED/AR/self-dmg cap, Smite ED/stun/AR |
| Concentration influence | 8 values | Formula matches arithmetic | d2d has no counterpart |
| Blessed Hammer flight & hits | 4 + 14 runs | Game side measured | d2d side is source-derived |

**PR candidate: none.** The AGENTS.md / upstream standard requires a deviation proven by an existing harness with mismatches before > 0 and 0 after fix. D1-D7, D8, and D9 are absent rules or require test harness modes not yet merged into upstream d2d.

---

## 9. Drill recipe (live; lab binary `lab-debug-skills`)

- **Z1 (Zeal hit count & rollback, D9).** `debug points 106 1`, `debug points 106 4`, `debug cast 106 unit <target>`: count strikes landed (expect 2 hits at level 1, 5 hits at level 4). On game.exe, strikes land every ~4 frames; on d2d today, each strike waits for full A1 animation recovery.
- **Z2 (Zeal retargeting).** Spawn two monsters in melee reach. Set monster 1 HP to 1 (`debug mon set <id1> life 1`). Cast Zeal on monster 1. Observe that after monster 1 dies on hit 1, strikes 2..5 switch to monster 2.
- **S1 (Smite unconditional hit & block bypass, D8).** Spawn a monster with high block rate (or shield-bearing monster, `debug mon set <id> block 75`). Cast Smite: observe 100% hit rate and damage dealt on every swing, never blocked.
- **S2 (Smite stun duration).** Cast Smite level 1 (stun 15 frames = 0.6 s) and level 20 (stun 110 frames = 4.4 s) on a non-boss monster. Verify monster is unable to act for exact stun duration.
- **C1 (Sacrifice self-damage cap).** Spawn monster with 50 life (`debug mon set <id> life 50 maxlife 50`). Cast Sacrifice with 500 damage. Verify player takes 8% of 50 damage (4 life), NOT 8% of 500 damage (40 life).
- **C2 (Sacrifice lethal self-damage).** Set player life to 10 (`debug stat 6 10`). Cast Sacrifice dealing 200 damage to a high-health monster (self damage = 16). Verify player dies from self-inflicted damage.

---

## 10. Matrix cell updates (proposed text for `pal-skill-matrix.md`)

| Row | Cell | New value |
|---|---|---|
| Smite | hit rule | V read: unconditional hit (`or ax, 1` in FUN_005ce9f0), ignores AR, defense, and block (d2d deviates, D8) |
| Smite | damage | V read + V run: shield base (Items.txt +0xfe/+0xff) + Holy Shield bonus + Str bonus (1% per Str) + calc1 ED (15 + 15*(lvl-1)%) |
| Smite | stun & knockback | V read + V run: ResultFlags = 8 (knockback); stun = min(250, 15 + 5*(lvl-1)) frames (cap 250 frames) |
| Smite | mana | V read + V run: 2.0 mana across all levels, paid in do; repeats on hold |
| Zeal | hit count | V read + V run: min(lvl + 1, 5) hits (2 at L1, 3 at L2, 4 at L3, 5 at L4+) |
| Zeal | damage & AR | V read + V run: AR 10 + 10*(lvl-1)%; ED ((lvl < 5) ? 0 : (lvl-4)*6%) + Sacrifice*12% |
| Zeal | mana & retarget | V read: 2.0 mana paid once at srvstfunc; retargets in reach via 0x20003 filter; rollback 100% (d2d lacks rollback, D9) |
| Sacrifice | damage & AR | V read + V run: AR 20 + 7*(lvl-1)%; ED 180 + 15*(lvl-1)% + Redemption*15% + Fanaticism*5% |
| Sacrifice | cost & self-hurt | V read + V run: 0 mana; self-damage 8% of min(phys_damage, target_hp); lethal if self-hurt >= player life |

---

## 11. Not traced (U) and limits

- Exact hit recovery frame duration when a target blocks a Smite swing (in game.exe the block animation plays if block roll succeeds, but full damage and stun apply regardless).
- The weapon speed modifier (WSM) and IAS table for Zeal rollback animation frames when equipped with two-handed weapons.
- Vengeance (Id 111) and Charge (Id 107) melee dispatch details (deferred to future card).
- Monster AI response to Smite knockback into walls or impassable tiles.

---

## 12. Shared-file lines for the merger

In `openspec/changes/palskill/merge-lines.txt` (BACKLOG, INDEX, CHANGELOG, and the decomp-queue row below).

Decomp-queue row: `- NN Paladin skills (PALSKILL, 2026-11-01, updated for Issue #3): Smite, Zeal, and Sacrifice are implemented in d2d's skill_built (do not fall back to plain attack); Smite unconditionally hits and bypasses shield block (d2d allows block, D8), scales with shield base, Holy Shield, Str bonus, and calc1 ED, stuns up to 250 frames; Zeal hits scale 2..5 (cap at L4), pays 2 mana once, retargets in reach, uses 100% frame rollback (d2d lacks rollback, D9); Sacrifice pays 0 mana, deals 180+15*(L-1)% ED, self-damage is 8% of damage capped at target current HP and is lethal; Holy Bolt and Blessed Hammer rules preserved; 120,000 cases verified 0 differ; bot B1-B8`.
