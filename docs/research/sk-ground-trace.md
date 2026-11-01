# Ground and area skills: game.exe 1.14d vs d2d vs the bot (decomp row 22, SKGROUND)

> status: card skground, 2026-11-01, offline (no game, no d2d slot, nothing pushed or posted). Decomp row 22 (sorc-skill-matrix rank 9). Finished: **Blizzard** (cast, centre, shard spawn by game.exe run, shard hit by read) and **Meteor** (cast, impact, blast, fire hand-off by read). Read to the dispatch level: **Frozen Orb**, **Thunder Storm**. Not reached in game.exe: **Hydra**, **Fire Wall** (tables and d2d side only). The board item should stay open for those two and for the U rows below.

Tags: **V read** disassembled by this agent with upstream's `tools/emu/disasm.py` on the patched 1.14d game.exe (addresses given); **V run** game.exe's own code run natively under the emulator by this agent (`tools/emu/sk_ground_blizz.py`); **D** the patched 1.14d tables as game.exe's MPQ stack serves them (`tools/emu/sk_ground_tbl.py tree`); **S** d2d source at upstream main 173b79cc (a scratch `git worktree` on the worker host), an upstream doc, or an earlier trace of ours (named); **I** inference or arithmetic; **U** not verified.

## Answer in eight lines

1. **How a ground cast reaches the ground (Blizzard, Meteor):** srvdofunc 28 (FUN_005ca3e0) -> FUN_0056ede0 makes ONE srvmissilea row AT the target point (params flags 1, owner the caster, skill and level), refused (no missile, do returns 0, so no mana is paid: SKMISS) when the point is more than 0x64 by FUN_006417f0 from the caster. d2d launches the same row at the point with no distance cap (V read; S).
2. **Blizzard** = `blizzardcenter` standing 100 frames (Range, D). Its missile do 10 (FUN_005aea60) calls the shard spawner FUN_005a9820 each frame: **one shard when age % calc2 == 0** (calc2 = Param3 = 4 at every level, D), so **25 shards a cast** at most; the shard's spot is **cx + rand(2r) - r, cy + rand(2r) - r with r = calc1 - 1 = 6 (radius Param1 7): offsets -6..5, a 12 x 12 square, not centred**; the draw is on the centre's seed reset to **{cx + age, 666}** each time; a spot failing collision mask 5 gets no shard (V read + **V run, 64,200 cases, 0 differ from the reading**).
3. **A shard** (`blizzard1`, Size 2, Range 9, no hit function, CollideKill 0, LastCollide 1) stands 9 frames; missile do 3 (FUN_005ae480) marks 0x40 in its room's collision at its subtile (FUN_0064cb90) and runs the generic step FUN_005ae1f0: per frame at most ONE unit in the Size-2 footprint (FUN_00641cb0), not the last unit it hit, takes a hit with a fresh roll of Blizzard's cold damage (V read; footprint S upstream missiles.md; that the standing shard's tested subtile is its own: I).
4. **Meteor** = `meteorcenter` standing 60 frames (Range, D), then its hit function 14 (FUN_005aabb0) with no unit: **one damage roll (FUN_005a89a0) applied to every unit in the blast** by FUN_0056bad0, radius sHitPar1 or else aurarangecalc = Param1 6 subtiles at every level (D: Param2 empty), then the fire (`meteorfire`, HitSubMissile1) for FUN_004cc7c0's frames through FUN_005aaa90 (V read; the fire's duration S from d2d's trace).
5. **Frozen Orb** (missile do 15 FUN_005af030, hit 29 FUN_005abb00): a `frozenorbbolt` every Param1 = 1 frame toward direction d of 64 (tables 0x6e2b78 / 0x6e2a78), d += Param2 19; at the range end 16 `frozenorbnova` (every sHitPar1 = 4th of the 64 ways, tables 0x6e2738 / 0x6e2638). d2d's do 15 / hit 29 match this reading (V read; S).
6. **Thunder Storm** (srvdofunc 29 FUN_005ca4d0): the first run starts state `thunderstorm` for auralencalc = 800 + 200 (lvl - 1) frames with stats 0x15e / 0x15f (skill, level); while it lasts the periodic run picks a unit within FUN_005c8bc0's radius after the last one struck (FUN_0056bd10) and strikes it at once with a `thunderstorm1` (FUN_005adf10 flag 1). d2d's states.cpp follows the same reading (V read; S).
7. **Skill delay** (Blizzard 45, Meteor 30, Frozen Orb 25, Hydra 40, Fire Wall 35 frames, D): FUN_0056ef90, called from the skill do (FUN_0056f7f0, at the action frame, SKMISS), puts state 121 `skilldelay` (D) on a PLAYER for `delay` frames (timer event 0xc, end callback 0x56e900). d2d does not model it for the player (K7 confirmed: `delay` is read only for pets). Where the server refuses a cast during the state was not found (U).
8. **d2d deviations:** D1 Blizzard shard offsets -(R-1)..(R-1) on d2d (13 values for R 7) vs -(R-1)..(R-2) in game.exe (12): 14.8 % of d2d shards fall on a row or column game.exe never uses (V run game side, port d2d side); D4 shard footprint (d2d: every monster within 0.5 cell, about 4x the game's area for a size-1 monster; I/U, the largest effect); D5 Meteor rolls per monster on d2d; D7 skill delay. **No PR candidate** (section 7). **Bot:** the delay timer starts at the input, not at the action frame (B1).

## 1. Questions answered (Q1..)

| id | question | answer | where | tag |
|---|---|---|---|---|
| Q1 | Which functions run each skill? | Skills.txt (D): Blizzard (59) and Meteor (56) srvdofunc 28 (0x7322b0[28] = FUN_005ca3e0), srvmissilea `blizzardcenter` / `meteorcenter`; Fire Wall (51) srvdofunc 24 (FUN_005c9ea0), `firewallmaker` + `firewall`; Thunder Storm (57) srvstfunc 13 (0x732140[13] = FUN_005c91c0), srvdofunc 29 (FUN_005ca4d0), periodic, `thunderstorm1`; Hydra (62) srvstfunc 14 (FUN_005c9220), srvdofunc 144 (FUN_005ca910), summon `hydra1`; Frozen Orb (64) no functions, plain srvmissile `frozenorb`. Missile do table 0x73c768 (do 3 FUN_005ae480, 5 FUN_005ae520, 6 FUN_005ae680, 10 FUN_005aea60, 15 FUN_005af030, 16 FUN_005af170; do 1 FUN_005b0bc0 = jmp FUN_005ae1f0); hit table 0x73c840 (14 FUN_005aabb0, 29 FUN_005abb00) | tables dumped with `sk_ground_tbl.py dw` | D + V read |
| Q2 | How does a ground cast reach the ground? | FUN_005ca3e0(game ecx, unit edx, skill, level): Skills record +0x48 (srvmissilea) valid; unit +0xc4 |= 0x40; FUN_0056e450(game, unit, row) must pass; FUN_0056ede0(game, unit, skill, level, row, 0, 0): with x = y = 0 the point is FUN_0056d2c0's (the target unit's spot or the clicked point); FUN_006417f0(unit, x, y) > 0x64: return 0 (nothing made; the do then pays no mana, SKMISS Q-table); else params {flags 1, owner, skill, level, start = the POINT, row} -> FUN_0059fa30. The do returns 1 only when the missile was made | 0x5ca3e0..0x5ca46d, 0x56ede0..0x56ee84 | V read |
| Q3 | Blizzard's centre | `blizzardcenter` (158): Vel 0, Range 100, pSrvDoFunc 10, SubMissile1 `blizzard1`, NoMultiShot (D). Do 10 (FUN_005aea60) each frame: calc1 (Skills +0x138 = `par1` = 7) and calc2 (+0x13c = `par3` = 4) at the missile's skill / level (FUN_0064a280 / FUN_0064a210, FUN_00646ca0), then FUN_005a9820(eax calc2, esi missile; game, calc1, SubMissile1 row (+0x18), 5), then FUN_005ae1f0 (the step: Range - 1, end at 0). Duration 100 frames = 4.0 s at every level (Param4's note "Change duration in Missiles.xls", D) | 0x5aea60..0x5aeb52 | V read + D |
| Q4 | Where Blizzard's shards fall | FUN_005a9820: freq 0 -> none; age = FUN_0064a3b0 (Range total - Range left, missile data +0x14 words +0xe / +0x10); age % freq != 0 -> none; owner (FUN_00552fd0) and room (FUN_00620bb0) needed; seed (missile +0x20) := {subtile x + age, 0x29a} (FUN_00650e40); r = calc1 - 1; dx = rand(2r) - r, dy = rand(2r) - r (FUN_0045c3e0, x first); spot = centre subtile + (dx, dy); FUN_0064cb30(room, x, y, 5) nonzero -> none; else params {flags 3, owner, row, start spot, skill, level} -> FUN_0059fa30 | 0x5a9820..0x5a9921 | V read + **V run** (section 5) |
| Q5 | How many shards, how often | one each 4 frames at ages 0, 4, .., 96: 25 a cast, fewer where a spot is blocked; radius 7 and frequency 4 at every level (calc1 `par1`, calc2 `par3`, no level term) | | D + V run (`made` column) |
| Q6 | What a shard does | `blizzard1` (159): Vel 0, Range 9, Size 2, CollideType 3, LastCollide 1, AlwaysExplode, NoMultiShot, ReturnFire, GetHit, no pSrvHitFunc, no CollideKill, Skill Blizzard (D). Do 3 (FUN_005ae480): a moving path -> the plain step; standing: FUN_0064cb90 ORs 0x40 into its room's collision word at its subtile, then FUN_005ae1f0: Range - 1 (end at 0, no unit test that frame); CollideType 3's filter (0x73c720 / 0x73c724); the path's collision bits (FUN_00648eb0) must be non-zero; for each subtile in the path's list (FUN_00648f40) a collision test (FUN_0064d9b0) then FUN_00641cb0 finds ONE unit (footprint by Size 2 and the unit's size) -> FUN_005adf10(unit, 0) and the frame ends. With no hit function the hit is the default damage: the skill's cold min / max rolled min + rand(max - min) on the missile's seed (SKMISS Q-table, FUN_005a89a0), resistances as SKRESIST. CollideKill 0: the shard stays; the filter skips only the LAST unit hit (S, sk-missile-trace D4) | 0x5ae480..0x5ae510, 0x5ae1f0..0x5ae3fa | V read (chain); S (footprint, filter); **I** (the standing shard's list holds its own subtile: no damage path is left otherwise) |
| Q7 | Blizzard damage | EMin 45 EMax 75, brackets 15/30/45/55/65 and 16/31/46/56/66, HitShift 8, synergy (Ice Bolt + Ice Blast + Glacial Spike) x par8 5 %; ELen 100 (chill) (D); the min / max build is SKMISS's (S, PR #154 proof) | | D + S |
| Q8 | Meteor's centre and impact | `meteorcenter` (101): Vel 0, Range 60, do 1, pSrvHitFunc 14, AlwaysExplode, HitSubMissile1 `meteorfire`, sHitPar2 1 (D). It stands 60 frames (2.4 s) at the point and goes off at the range end with no unit (FUN_005ae1f0 -> FUN_005adf10(0, 1) -> hit 14) | | D + V read |
| Q9 | Meteor's blast | FUN_005aabb0: radius = missile +0x4c (sHitPar1, blank) else calc of Skills +0x64 (aurarangecalc `ln12` = Param1 6 + Param2 (blank) x (lvl - 1) = 6 at every level), at least 1; ONE record by FUN_005a89a0 (the roll), OR'd with the row's hit flags (+0xa8 / +0xac); FUN_0056bad0(game, owner, x, y, radius, record, 0) -> FUN_0056b7e0 (unit flags 0x8583, callback 0x56ba90) gives that SAME record to every unit found; returns 1 when anything was hit, else 3 | 0x5aabb0..0x5aad3b, 0x56bad0..0x56bb12 | V read |
| Q10 | Meteor's fire | after the blast, HitSubMissile1 (+0x24) >= 0: FUN_004cc7c0(skill, level) (frames; d2d's trace: Param3 30 + (lvl - 1) x Param4 15) and missile +0x50 (sHitPar2 1, "flames increment", at least 1) into FUN_005aaa90 (placement). `meteorfire` (240): do 5, pSrvDmgFunc 3, Size 2, DamageRate 41, EMin 15 / Emax 25 brackets 4-6, HitShift 3, ApplyMastery, Inferno synergy x 3 (D) | 0x5aad00..0x5aad2d | V read (call); S (frames); FUN_005aaa90 and do 5 not read (U) |
| Q11 | Meteor damage | EMin 80 EMax 100, brackets 23/39/79/81/83 and 25/41/81/83/85, HitShift 8, synergy (Fire Bolt + Fire Ball) x par8 5 % (D) | | D |
| Q12 | Frozen Orb | `frozenorb` (260): Vel 10, Range 30, do 15, hit 29, Param1 1, Param2 19, sHitPar1 4, SubMissile1 `frozenorbbolt` (Vel 18, Range 25, CollideKill), HitSubMissile1 `frozenorbnova` (do 16, Vel 24, Range 25) (D). FUN_005af030: freq = Param1 (+0x38) at least 1; age % freq > 0: plain step; else a bolt {flags 2, owner, parent, skill, level, row +0x18} aimed along entry d of 0x6e2b78 (x) / 0x6e2a78 (y), d = FUN_0064a730(missile) & 63, then d := (d + Param2 (+0x3c)) & 63 (FUN_0064a710), FUN_0059fa30, step. FUN_005abb00 (hit 29): nothing while Range left != 0 (returns 2: flies over units); at the end the nova row (+0x24) {flags 2, ...} along entries 0, 4, 8, .., 60 of 0x6e2738 / 0x6e2638 (step 4 x sHitPar1 bytes): 16 missiles | 0x5af030..0x5af16a, 0x5abb00..0x5abc30 | V read + D |
| Q13 | Thunder Storm | FUN_005ca4d0(game, unit, skill, level): Skills +0x48 row and +0x80 aurastate valid; skill entry FUN_006439f0; state on (FUN_00639df0) and the entry not busy (FUN_006444d0): target = FUN_0056bd10(game, unit, 0, 0, FUN_005c8bc0(skill) radius, 3, last target FUN_006444a0, 0); found: FUN_005ca470 makes the bolt, FUN_005adf10(bolt, unit, 1) strikes at once, FUN_00555600, FUN_00571aa0, last target := unit id (FUN_00644560); none: last target -1. State off: FUN_0056e970 starts it, length = calc of +0x60 (auralencalc `ln12` = 800 + 200 (lvl - 1) frames, D), stats 0x15e skill, 0x15f level. Period: perdelay (100 - dm56) x par4 / 100 + par3 (D; where it is evaluated not read, S d2d) | 0x5ca4d0..0x5ca6b8 | V read + D |
| Q14 | Skill delay | FUN_0056ef90 (eax delay, edi unit): end = game frame (+0xa8) + delay; only unit type 0 (player) and delay != 0; state 121 (`skilldelay`, States.txt, D) found by FUN_006256b0 or made (FUN_006251f0 flags 2, FUN_006252d0, end callback 0x56e900 via FUN_00625ce0, FUN_00626e10, FUN_00639db0 flag on); its end set (FUN_006260b0) and a timer event 0xc at that frame (FUN_005417d0). Called from the skill do after the payment (SKMISS Q-table: FUN_0056f7f0) | 0x56ef90..0x56f018 | V read + D |
| Q15 | Mana (D, formula S skills.md) | `(mana + lvlmana (L - 1)) << manashift / 256`: Blizzard 23/1/8 = 23 + (L - 1); Meteor 34/1/7 = 17 + (L - 1) / 2; Frozen Orb 50/1/7 = 25 + (L - 1) / 2; Hydra 40/1/7 = 20 + (L - 1) / 2; Fire Wall 22/1/8 = 22 + (L - 1); Thunder Storm 19/0/8 = 19; minmana 1 everywhere | | D + S |
| Q16 | Fire Wall, Hydra (tables only) | Fire Wall: do 24 (FUN_005c9ea0, not read), `firewallmaker` (Vel 12, Range 7 + LevRange 2 x lvl, do 6, SubMissile1 `firewall`), `firewall` (do 5, pSrvDmgFunc 3, Range 90, DamageRate 41, Skill Fire Wall), HitShift 4, EMin 15 EMax 20 brackets 9/14/21 (D). Hydra: srvstfunc 14, do 144, summon `hydra1` (petmax 99), sumskill HydraMissile at lvl, duration Param1 250 frames, HitShift 7, EMin 28 EMax 39 (D) | | D only (game code not reached: U) |

Dead ends: the `push 0x79` hits in the full listing other than FUN_0056ef90 are not the delay state: 0x57b642 / 0x57b863 / 0x5a8b22 read STAT 0x79 (FUN_00625500), 0x58b004 / 0x58bd1e pass 0x79 as a plain argument; the check that refuses a cast during `skilldelay` is probably a state-bit test (no immediate), not found. FUN_005b0bc0 (missile do 1) is a one-instruction jump to FUN_005ae1f0, which is why a dword search for 0x5ae1f0 finds no table entry. The three 64-entry x tables 0x6e1288 (nova), 0x6e2b78 (orb bolts), 0x6e2738 (orb nova) hold the same values (dumped); the y tables were not compared (I: same).

## 2. Blizzard and Meteor as formulas

```
cast (do 28):   centre = srvmissilea at the point; none if dist > 0x64 (no mana then)
Blizzard:       for age in 0..99: if age % 4 == 0:
                   seed = {cx + age, 666}; dx = rand(12) - 6; dy = rand(12) - 6
                   shard at (cx+dx, cy+dy) unless collision & 5 there
shard:          9 frames; each frame <= 1 unit in its Size-2 footprint, not the last one hit:
                   cold = min + rand(max - min) (skill level, synergies), resist, chill ELen
Meteor:         60 frames after the do: one roll R; every unit within 6 takes R; then fire
                   for 30 + 15 (L - 1) frames (S)
delay:          state skilldelay on the player for `delay` frames from the do (action frame)
```

Worked numbers (I, arithmetic on the rules): a Blizzard covers 144 subtiles (12 x 12) with 25 shards over 4 s; a size-1 monster standing still is in a shard's footprint (5 subtiles, Q6) for about 25 x 5 / 144 = 0.87 shards a cast on average; bigger monsters have bigger footprints (upstream missiles.md, S: missile 2 / unit 2 |dx| + |dy| <= 2, 13 subtiles; missile 2 / unit 3 the 5 x 3 cross, 13 subtiles): about 2.3 shards a cast. The pattern's centre of mass is at (-0.5, -0.5) subtile from the point. Meteor at level 1: 80-100 points (HitShift 8) to everything within 6 subtiles, the same number for all.

## 3. d2d (S, upstream main 173b79cc)

- Cast: `Fight::fire` kMissileAtTarget (fight/missiles.cpp:345): `launch` at (to_x, to_y), range = Range + LevRange x lvl; no distance cap.
- Blizzard centre: `missile_tick` (:460): do 10 / 25 with SubMissile1: every = calc2, radius = calc1; `age % every == 0` -> offset `rng(2 radius - 1) - (radius - 1)` per axis (fight Rng), shard launched with its row's Range; no collision check at the spot.
- Shard hit: standing friendly missiles go through the `missiles_update` callback in fight/tick.cpp:390-411: every targetable monster within `hypot <= 0.5f` cells (2.5 subtiles), once per missile (`struck`), `skill_missile_hits` -> `strike` (one roll per monster).
- Meteor: hit 14 is in `kEnd` (:480): `burst` when `now + 40 >= dies` -> `area(missile, x, y, radius)` (:695: every monster with `hypot x 5 <= radius`, each `strike` rolling its own) then 18 spread points (`kSpreadX / kSpreadY`, every sHitPar2-th) of the fire lasting `par[2] + (lvl - 1) par[3]` frames.
- Frozen Orb: do 15 (:485) every Param1 frames along `kNovaRing[turn & 63]`, `turn += Param2`; hit 29 burst (:765) every sHitPar1-th of 64.
- Thunder Storm: fight/states.cpp:236-258, as Q13 (next id after the last, round to the lowest, within Param7 x 1 subtile), `missile_blow` directly.
- Skill delay: `Skill::delay` (components/rules/skills.hpp:413) is read for pets only (fight/pets.cpp:586); the player can recast a Blizzard at once (K7).

## 4. Deviations (d2d vs game.exe)

| id | rule | game.exe | d2d | tag | reach | differential |
|---|---|---|---|---|---|---|
| D1 | Blizzard shard offset range | rand(2r) - r, r = radius - 1: -6..5 at radius 7 (FUN_005a9820) | rng(2 radius - 1) - (radius - 1): -6..6 | **V run** game side (64,200 cases, the reading 0 differ; the offset +r never made at any radius); d2d by port of the line (S) | every Blizzard (and every do 10 / 25 row): 1 - (12/13)^2 = 14.8 % of d2d's shards land on the +6 row or column, which game.exe never uses; d2d's pattern is centred, game.exe's sits half a subtile toward -x / -y | game side yes (`sk_ground_blizz.py`); real d2d build not built (the draw is inside `Fight::missile_tick`; a test would need a Fight with a level and a Rng hook: how in section 7) |
| D2 | shard seed | the centre's seed set to {cx + age, 666} per shard: the pattern is a function of the centre's x subtile and age only | the fight's shared Rng | V run (game) + S | determinism only (same cast spot, same shards in game.exe) | as D1 |
| D3 | shard on a blocked spot | none (FUN_0064cb30 mask 5) | launched anyway | V read (game) + S | Blizzard next to walls: d2d shards land inside walls; the hit test is distance-only, so they can still hit a monster next to the wall | port only |
| D4 | shard hit area and repeats | one unit a frame in the Size-2 footprint (size-1 unit: the subtile and its 4 neighbours), not the last unit hit, 9 frames, a fresh roll per hit | every monster within 0.5 cell (about 19.6 subtiles of area), each once a shard | V read (chain) + S (footprint) + **I** (own subtile tested); d2d S | the largest effect: for size-1 monsters d2d's shard area is about 4x game.exe's, so d2d over-rates Blizzard's damage per cast; two monsters in one footprint can each be hit more than once in game.exe (alternation past LastCollide), once on d2d | not built: needs missiles.py's Game (a real FUN_0059fa30 shard stepped by the REAL FUN_005ae480 with one monster placed at each offset; oracle = which offsets get FUN_005adf10 called) |
| D5 | Meteor blast roll | one roll for all units (FUN_005aabb0 -> FUN_005a89a0 once -> FUN_0056bad0) | `area` -> `strike` per monster, a roll each | V read + S | same mean, other spread; changes every draw after the first (determinism vs game.exe) | port only; a real-build check would compare the per-monster damage set of one burst (all equal in game.exe) |
| D6 | Meteor blast shape | FUN_0056b7e0 with flags 0x8583 and callback 0x56ba90: shape not read | `hypot x 5 <= radius` circle | U (game) + S | unknown | not built |
| D7 | skill delay | state `skilldelay` on the player for `delay` frames from the do (FUN_0056ef90) | none for the player | V read + S (K7 confirmed) | d2d lets a Blizzard / Meteor / Frozen Orb be recast every cast time: a d2d soak over-rates ground builds' damage per second | not built: stateful (drill G3) |
| D8 | ground cast distance | no missile, no mana, when FUN_006417f0 > 0x64 | no cap | V read + S | only casts at more than about 100 subtiles (beyond the screen: I) | none |

Not deviations at the read depth: Frozen Orb's bolt cadence, turn and end nova (Q12 = d2d), Thunder Storm's target rule (Q13 = d2d), Blizzard's 25 shards (age % 4 with d2d's age from `born`; whether both run the do on the creation frame: U), Meteor's 60-frame fall (d2d bursts at `now + 40 >= dies`: frame 59 or 60 not checked, U).

## 5. The differential (V run, game side)

Oracle `tools/emu/sk_ground_blizz.py` (runs next to upstream's tools/emu, `emu.py`): per case `id cx cy age freq radius` it calls game.exe's FUN_005a9820 natively (EAX freq, ESI a missile unit, stack game / radius / row 159 / 5) with the age, owner, room, centre subtile, skill / level, the spot's collision (free) and the missile maker hooked; the seed set FUN_00650e40, the draws FUN_0045c3e0 and the memset run as game code. `cmp` checks every line against the reading (Q4) and counts game.exe offsets against d2d's range.

```sh
# the worker host, from the card's tools/emu helper (copies the script into the scratch upstream checkout)
tools/emu/sk_ground_run.sh <upstream> sk_ground_blizz.py gen 20000 1 > b1.cases
tools/emu/sk_ground_run.sh <upstream> sk_ground_blizz.py game b1.cases > b1.game     # about 1 s
tools/emu/sk_ground_run.sh <upstream> sk_ground_blizz.py cmp b1.cases b1.game
```

Cases: grid 2,100 = radius 1 / 2 / 3 / 7 / 8 / 12 / 20 x age 0..99 x 3 centres, frequency 4; plus N random (centre 1,000..61,000 subtiles, age 0..99, frequency 1..6, radius 7 three times in four else 1..25). xorshift32 on a masked seed.

| run (the worker host) | cases | shards made | reading differs | game offset outside d2d's range | offsets at radius 7 |
|---|---|---|---|---|---|
| seed 1 (`gen 20000 1`) | 22,100 | 8,744 | 0 | 0 | game -6..5 (12 values), d2d -6..6 (13); d2d-only {+6} |
| seed 7 (`gen 40000 7`) | 42,100 | 16,868 | 0 | 0 | same |

md5 (first 8): b1.cases 1f104d8d, b1.game 52d7a91a, b7.cases cdac5c6a, b7.game adfe14a3. Every radius 2..21 shows the same "d2d-only = {+(R-1)}"; radii 22..25 also list a few interior values game.exe did not draw in that sample (few random cases at those radii: sampling, not a rule). Radius 1: 0 both.

What is NOT covered by this run: the collision check (hooked free), the shard's hit (D4), d2d's real build (port of one line).

## 6. The bot (our repo; described, not changed)

- Ground skills are `kind: ground` in `internal/combat/classdata/skills.json` (Blizzard 30, Meteor 30 subtiles; Thunder Storm is `buff`): aimed at the monster's spot (`KindGround`, internal/combat/skills.go:29).
- Delay: `SkillSpec.Delay` = Skills.txt `delay` x 40 ms (internal/combat/skills.go:118-119); `skillReady` (internal/combat/fighter.go:773-779) compares with `lastCastSkill`, booked when the input is sent (fighter.go:724-725), whatever `err` says.

| id | finding | file:line | tag | impact |
|---|---|---|---|---|
| B1 | the delay clock starts at the input; game.exe starts `skilldelay` in the do at the action frame (about 7-9 frames later at 0 FCR, cast-trace) | internal/combat/fighter.go:724-725, 773-779 | V read (game) + S (our code) | the bot tries the next Blizzard about 0.3 s before the game allows it; what the server does with that cast is U (G3) |
| B2 | the delay is booked even when the input failed or the cast was cut off (no missile, no state in game.exe) | internal/combat/fighter.go:724-725 | S (our code) + V read | the bot waits a full delay for a Blizzard that never happened |
| B3 | the bot does not read state 121 `skilldelay` from memory, which would give the game's own timer | internal/combat/fighter.go:773 | I | candidate: read the player's state flags for 121 instead of a local clock (needs a state-flag offset check, VERIFIED LIVE first) |
| B4 | aiming at the monster's spot: shards fall in -6..5 subtiles around it, centred at (-0.5, -0.5); a size-1 monster is in about 0.87 shard footprints a cast (I) | internal/combat/skills.go:29 | I | none for a crowd; for a single small target Blizzard is weaker than the tooltip suggests: Glacial Spike / Ice Blast during the delay (the END profile already does) is right |
| B5 | on the d2d backend Blizzard is stronger (D4 area, D7 no delay) and Meteor rolls per target (D5): d2d kill times for the END build are optimistic | internal/d2denv (backend) | I (from D4 / D7) | read d2d END-build soak numbers as an upper bound |

## 7. PR candidate

**No PR candidate.** D1 is proven on the game side by a run and on the d2d side only by a port of one line: no real-build before/after exists, so the standard (mismatches before > 0 on the real build, 0 after) is not met. Building it: a `tests/game/test_ground.cpp` mode that constructs a Fight with a flat level, places one `blizzardcenter` with a fixed level, replaces the fight Rng's draws by game.exe's seed rule (or only collects the offset SET over many seeds), steps `missile_tick` 100 frames and prints each shard's offset from the centre; before: the +6 offsets appear; after the one-line fix (`rng(2 (radius - 1)) - (radius - 1)`), none. One claim commit would suffice. D4 / D7 need more than one line (footprint by sizes; a delay state). Not built in this card's time box; the card has no `fork-branch:` / `upstream-pr:` resource.

## 8. Drill recipe (lab binary, fork verbs; for a later live card, not run here)

Verbs (docs/forkdbg.md, docs/forkcheats.md): `debug stat 12`, `debug points 59 <n>` / `56` / `64`, `debug goto`, `debug spawn zombie1 <n> near`, `step N`, `cmd skill <id> X Y`, `debug missiles` (unit, missile id, position, ttl, owner), `monsters` (hp). Missing verbs: `debug states` (needed for G3), `debug mon set` (pin hp).

- G1 (D1, Q4): Blizzard 1 at an empty point (X, Y) in a flat area; `step 1` and `debug missiles` each frame for 100 frames: record every `blizzard1` position minus (X, Y). game.exe: 25 shards (fewer only next to walls), offsets -6..5, the same 25 offsets for every cast at the same X subtile; d2d before a fix: -6..6, different each cast.
- G2 (D4): one zombie (size 1) parked at (X + a, Y + b) for a grid of (a, b); cast Blizzard many times; count hits per shard from hp drops: game.exe hits only when a shard lands on the zombie's subtile or a 4-neighbour; d2d within 2.5 subtiles.
- G3 (D7, B1): with `debug states player` (missing verb) watch state 121 after a Blizzard: it should run 45 frames from the cast's action frame; send a second `cmd skill 59` 10, 30, 50 frames after the first: game.exe refuses (or delays) the ones inside the state, d2d casts all.
- G4 (D5): Meteor 1 on 5 zombies packed within 3 subtiles with known hp: every zombie loses the SAME hp in game.exe, different amounts on d2d.
- G5 (Q12): Frozen Orb at a far point: `debug missiles` shows a `frozenorbbolt` every frame turning 19 / 64 each, then 16 `frozenorbnova` at the end (frame 30).

## 9. Matrix cell updates (docs/research/sorc-skill-matrix.md; proposed text for the merger)

| Row (matrix) | Cell | New value |
|---|---|---|
| 59 Blizzard (line 49) and section 3.13 (line 227) | meteor count, radius, tick timing | V run + V read (sk-ground-trace.md): centre 100 frames, a shard every 4 frames (25 a cast), offsets -6..5 (radius 7, not centred) on the seed {cx + age, 666}, none on a blocked spot; a shard stands 9 frames, one unit a frame in its Size-2 footprint; d2d D1 (offsets -6..6), D4 (0.5-cell hit area), D7 (no delay) |
| 56 Meteor (line 46) | ground behaviour | V read: centre at the point 60 frames, one roll for every unit within 6 subtiles (FUN_005aabb0), fire 30 + 15 (L - 1) frames (S); d2d D5 (a roll per monster), D6 shape U |
| 64 Frozen Orb (line 54) | missile behaviour | V read: a bolt every frame turning 19 / 64, 16-missile nova at frame 30; d2d matches |
| 57 Thunder Storm (line 47) | behaviour | V read: state 800 + 200 (L - 1) frames, strikes the next unit id within Param7 each period; d2d matches |
| 51 Fire Wall, 62 Hydra | behaviour | not reached (SKGROUND left them open) |
| K7 (line 84) | delay | V read: FUN_0056ef90 puts state 121 `skilldelay` on the player for `delay` frames from the do; d2d reads `delay` for pets only (confirmed); the refusing check not located (U) |
| rank 9 SKGROUND (line 253) | status | partly done (sk-ground-trace.md): Blizzard, Meteor finished; Frozen Orb, Thunder Storm read at dispatch level; Hydra, Fire Wall open |

## 10. Not traced (U)

- Whether a standing shard's entered-subtile list (FUN_00648f40) holds its own subtile (Q6, the I under D4), and the footprint for size-2 / size-3 monsters at missile Size 2 as game.exe runs it.
- FUN_0056b7e0's unit search shape and flags 0x8583 (Meteor's blast, D6); FUN_005aaa90 (fire placement); do 5 / pSrvDmgFunc 3 (the fire's per-frame damage, DamageRate 41).
- The server check that refuses a cast while `skilldelay` is on (G3).
- The do on the creation frame (shard at age 0) and the exact impact frame of Meteor on both sides.
- Fire Wall (FUN_005c9ea0, do 6 FUN_005ae680, do 5 FUN_005ae520), Hydra (FUN_005c9220, FUN_005ca910, the hydra's AI and missile) in game.exe.
- Frozen Orb's initial direction (FUN_0064a730 on a new missile) and the y tables 0x6e2a78 / 0x6e2638 vs d2d's `kNovaRing`.

## Shared-file lines for the merger

docs/decomp-queue.md, row 22 (replace the status cell): `partly done (SKGROUND 2026-11-01, docs/research/sk-ground-trace.md: Blizzard, Meteor done; Frozen Orb, Thunder Storm read; Hydra, Fire Wall open)`; and append to the finished list:

```
- 22 ground skills (SKGROUND, 2026-11-01, partial): ground casts make the srvmissilea row AT the point (FUN_005ca3e0 -> FUN_0056ede0, none past 0x64, then no mana); Blizzard: centre 100 frames, FUN_005a9820 a shard every calc2 4 frames (25), offsets rand(2r) - r with r = radius - 1 (-6..5 at 7) on the seed {cx + age, 666}, none on a blocked spot (V run 64,200 cases, 0 differ from the reading); shard 9 frames, one unit a frame in its Size-2 footprint (V read, I); Meteor: 60 frames, one roll for all within 6 (FUN_005aabb0); Frozen Orb, Thunder Storm match d2d at read depth; skill delay = state 121 skilldelay from the do (FUN_0056ef90); d2d D1 offsets -6..6 (14.8 % of shards off game.exe's grid), D4 0.5-cell hit area (about 4x), D5 Meteor roll per monster, D7 no player delay; no PR (D1 not on a real build); bot B1 delay clock at the input, B2 booked on failed casts; Hydra, Fire Wall open; docs/research/sk-ground-trace.md
```

docs/BACKLOG.md (fix queue, one line each):

```
- SKGROUND D1: d2d Blizzard shard offsets -(R-1)..(R-1) vs game.exe -(R-1)..(R-2) (FUN_005a9820); game-side oracle tools/emu/sk_ground_blizz.py; needs a real-build test mode before a fix branch (docs/research/sk-ground-trace.md section 7)
- SKGROUND D4: d2d standing-shard hit area 0.5 cell vs game.exe Size-2 footprint, one unit a frame (needs a missiles.py Game oracle)
- SKGROUND D7: d2d has no player skill delay (state 121 skilldelay, FUN_0056ef90); drill G3 first
- SKGROUND B1/B2: bot delay clock starts at the input and is booked on failed casts (internal/combat/fighter.go:724-725)
- SKGROUND open: Hydra and Fire Wall in game.exe; Meteor blast shape FUN_0056b7e0
```

docs/INDEX.md (research list):

```
- docs/research/sk-ground-trace.md: Blizzard, Meteor (done), Frozen Orb, Thunder Storm (read), skill delay; game.exe vs d2d (D1 game-side run) vs bot
```

docs/research/README.md (if it lists traces): `sk-ground-trace.md - SKGROUND, decomp row 22 (ground / area skills, skill delay)`.

Community board item SKGROUND: `partly done 2026-11-01 (Blizzard, Meteor; Frozen Orb, Thunder Storm read): keep OPEN for Hydra, Fire Wall, D4 oracle and the D1 real-build test`.

CHANGELOG.md: none (docs and tooling only).

Glossary (proposed, not written): "skill delay (skilldelay)": the per-skill cooldown of Skills.txt `delay`, state 121 on the player for that many frames from the skill's action frame (FUN_0056ef90).
