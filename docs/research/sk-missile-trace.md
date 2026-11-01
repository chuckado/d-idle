# Sorceress missile pipeline: game.exe 1.14d vs d2d vs the bot (decomp row 14)

> status: card skmiss, 2026-11-01, offline (no game, no d2d slot). Decomp row 14: how the plan's missile skills (Fire Bolt, Charged Bolt, Ice Bolt, Frost Nova, Lightning, Chain Lightning; Ice Blast and Nova checked alongside) are spawned, flown and landed. The skill matrix (docs/research/sorc-skill-matrix.md) is not on main yet: the cell updates are listed at the end instead of edited there.

Tags: **V** read with upstream's `tools/emu/disasm.py` or run natively under the emulator by this agent (say which); **D** from the patched 1.14d tables (Skills.txt / Missiles.txt read with `mpq-cat` from the patch stack); **S** d2d source or upstream doc (file, on upstream main 72132b99) or an earlier V trace of ours (named); **I** inference; **U** not verified.

## Summary

- **Spawn.** The skill START (FUN_0056f640) only checks the mana for these skills; the mana is PAID in the skill's DO (FUN_0056f7f0), after the srvdofunc and the missile are made, at the cast's action frame. d2d pays at the cast command (V read; S). Charged Bolt makes calc1 = min(24, 3 + (lvl - 1)) bolts, all from the caster's subtile at the same target point, each with its index for the wiggle (V read + D). Plain bolts (Fire Bolt, Ice Bolt, Lightning) are one missile from the caster's subtile to the target point through FUN_0059fa30 (S, V read).
- **Flight.** Upstream's flight model (velocity ((Vel + VelLev x lvl / 8) << 8) x 75 / 100, Range frames, the 128-row aim table, the subtile walk, Charged Bolt's 77-frame wiggle) is already proven against game.exe by upstream's `tools/emu/missiles.py` (S); not redone here. Reach per missile in the table below (D + S arithmetic).
- **Hit.** The element's min / max (synergies, mastery) are built when the missile is made (FUN_0064b860, S: PR #154 skill-damage proof), the roll is FUN_005a8910: **nothing unless both ends are over 0, then min + rand(max - min) on the missile's own seed: the max is never rolled** (V read + V run), then resistance less pierce, MulDiv, absorb and life in 256ths (S: PR #135 life-apply proof).
- **Deviation D1 (proven on the real d2d build):** d2d's `missile_blow` rolls `Rng::range(min, max)` = min + rand(max - min + 1): the top 256th is rollable and the draw's modulus differs. On the same seed the two draws differ in 103,981 of 104,000 cases (the modulus is max - min + 1 instead of max - min); the monster's life after the hit differs in 31,459 (the rest are kills or immunities either way). Local upstream candidate `fix/missile-roll-top`: 0 after.
- **Deviation D2 (read, not run):** d2d spends a missile skill's mana at the command, game.exe at the action frame: a cast cut off by a get-hit costs mana on d2d, none in game.exe. Structural; drill recipe below.
- **Bot:** casts by Kolbot ranges well inside each missile's reach (UNVERIFIED but conservative); mana gate rounds the cost up (safe); no model of the mana drop timing. Findings B1-B4.

## 1. Spawn: what the cast creates

| Address | Rule | Tag |
|---|---|---|
| `FUN_0056f640` (skill start; ecx game, edx unit, skill, arg, flag) | flag 0: FUN_0056c160 mana check (stat 8 >= `(mana + lvlmana x (lvl - 1)) << manashift`, at the skill level FUN_006442a0 flag 1), refused without mana. Then srvstfunc (table 0x732140 by Skills +0x2c, < 0x5b): **none (all the plan's missile skills, D): ebx = 1 and NO payment**; a start function that succeeds pays (FUN_0056bfe0) here unless the skill's +8 flag (mask 0x6ce27c) is set | V read |
| `FUN_0056f7f0` (skill do) | srvdofunc (table 0x7322b0 by Skills +0x2e, < 0xbf) runs; with a srvmissile (+0x46) one missile goes through FUN_0056ecb0 (flags 0x21) or FUN_0056ee90 (0x420), by Skills +4 & 0x6ce26c; nothing made and the do function returned 0: return, no payment (0x56fa3d). Then (players and monsters) no start function OR the +8 flag clear: **FUN_0056bfe0 pays the mana here** (0x56fa57..0x56fa7a); with the +8 flag set the mana is checked again first (0x56f91d -> FUN_0056c160). Then the delay (FUN_0056ef90) | V read |
| when the do runs | the cast mode's action frame (upstream skills.md "Missile skills (phase 4)"; cast lengths: docs/research/cast-trace-gameexe.md, V) | S |
| `FUN_005c9300` (srvdofunc 17, Charged Bolt) | row FUN_005d3cf0(unit, skill) (the srvmissilea row, D: `chargedbolt`); count = FUN_00646ca0(unit, calc1 at Skills +0x138, skill, lvl); per bolt: params flags 0x21, owner the caster, start the caster's subtile (FUN_0045adf0 / FUN_0045ae20), the target point of FUN_0056d2c0 (the same for every bolt), skill and level, callback FUN_005c9290 with the bolt's index (0..count-1); a bolt is made only when the target point is non-zero; returns 1 whatever the count (so the do pays) | V read |
| Charged Bolt count | calc1 `min(24,ln12)` = min(24, Param1 3 + Param2 1 x (lvl - 1)): 3 bolts at level 1, 22 at 20, 24 from 22 | D |
| `FUN_005c9290` | range capped at 77 frames, seed {target x + index, 666}, path type 10 (range / 2 hops of 2 subtiles, each the 8-way way to the target turned -1 / 0 / +1 by a seed draw) | S (upstream missiles.md, proven by `missiles.py`) |
| `FUN_0059fa30` | missile create: start subtile centre, unit vector to the target (FUN_0064fc60), velocity ((Vel + VelLev x lvl / 8) << 8) x 75 / 100, range Range + LevRange x lvl frames, damage built at once (FUN_0059f900 -> FUN_0064b860) | S (upstream missiles.md; skill-damage proof) |
| Frost Nova / Nova (srvdofunc 22) | a ring of missiles from the caster; the ring's count and directions come from tables at 0x6e1288 | S (upstream "not yet": d2d sends 64 evenly spaced); U here |
| Chain Lightning (srvdofunc 26, FUN_005ca1b0) | one `chainlightning` missile carrying calc1 = ln34 / 5 hops (Param3 26 + Param4 1 x (lvl - 1)) / 5: 5 at level 1, 9 at 20; each hit (SrvHitFunc 12, FUN_005aa730) launches the next from the hit spot to another foe within sHitPar1 / aurarange | S (d2d comments, fight/missiles.cpp:597-604); D for the calc |

> Frost Nova / Nova ring: resolved by docs/research/sk-nova-trace.md (64 missiles, tables 0x6e1288 / 0x6e1388, reach 12-12.7 subtiles; D3 superseded).

The plan's skills (D, Skills.txt patched): no srvstfunc on any; srvdofunc none for Fire Bolt (36), Ice Bolt (39), Lightning (49) (srvmissile `firebolt`, `icebolt`, `lightningbolt`), 17 for Charged Bolt (38, srvmissilea `chargedbolt`), 22 for Frost Nova (44, `frostnova`), 26 for Chain Lightning (53, `chainlightning`). Mana (`mana`, `lvlmana`, `manashift`, `minmana`): Fire Bolt 5/0/7/1 (2.5 points), Charged Bolt 24/4/5/1 (0.75 + 0.125 per level), Ice Bolt 3/0/8/1, Frost Nova 9/1/8/1, Lightning 16/1/7/1 (8 + 0.5 per level), Chain Lightning 9/1/8/1.

### Missile rows (D, Missiles.txt patched) and reach (arithmetic on the S rules)

| Row (id) | Vel | Range | Size | CollideKill | SrvHitFunc | other | path velocity / frame | reach (subtiles = Range x vel / 4096) |
|---|---|---|---|---|---|---|---|---|
| firebolt (58) | 20 | 50 | 1 | 1 | 0 | Explosion fireexplode (visual) | 3840 (0.94 subtile) | 46.9 |
| icebolt (59) | 12 | 50 | 1 | 1 | 0 | Explosion iceexplode; ELen 150 chill (docs/research/freeze-rules-trace.md) | 2304 (0.56) | 28.1 |
| chargedbolt (56) | 12 | 98 -> 77 | 1 | 1 | 0 | wiggle path, HitClass 64 | 2304 (0.56) | 43.3 of path; the net distance is a random walk |
| lightningbolt (98) | 30 | 25 | 2 | **0** (passes through) | 0 | | 5760 (1.41) | 35.2 |
| chainlightning (93) | 30 | 25 | 2 | 1 | 12 (chain) | NextHit 1, NextDelay 4 | 5760 (1.41) | 35.2 per hop |
| frostnova (119) | 24 | 14 | 1 | 0 | 0 | Accel -1000, MaxVel 24, NextHit 1 / 4 | 4608 (1.13), -1000 every fifth frame | about 12 (I: 4 frames at 4608, 5 at 3608, 5 at 2608) |

No LevRange or VelLev on any of them (D): reach does not grow with the level. Ice Blast (`iceblast`, 91): Vel 12, Range 50, pSrvDmgFunc 4.

## 2. Flight and collision

S (upstream docs/research/re/missiles.md, proven frame by frame by upstream `tools/emu/missiles.py` for firebolt, icebolt, chargedbolt and others; `tests/monsters/test_monsters.cpp` replays 18 flights): each frame srvdofunc 1 (FUN_005ae1f0) moves the missile (FUN_00650840: every fifth frame velocity += Accel, clamped to MaxVel; delta = unit vector x (velocity x 0x400 >> 6) >> 12; the subtiles entered are tested against the collision mask, a blocked one ends the flight at the last free subtile), then Range - 1 (< 1: ends **without a unit test that frame**), then the units in each entered subtile by FUN_00641cb0 (missile Size x unit size footprint) through the CollideType 3 filter (targetable, not the last unit hit, hostile unless CollideFriend).

- Who it hits: hostile, targetable units (monsters for the player's missiles); not the caster, the merc or pets (S: FUN_005a8850 -> FUN_005a8730 filter; d2d strikes `monsters` only).
- The "one-tick overshoot" question: on its last frame (Range left reaching 0) the missile still moves but tests no unit (S). d2d's `missile_tick` fires the range-end burst at `now + 40 >= dies` for the hit functions that burst at the end, and its moving-flight hit test lives in `missile_fly` (ai.hpp); whether its last frame tests units was not re-run here (U; upstream's `missiles.py` compares end frames against a copy of the rules header, not the full Fight path).
- Lightning (CollideKill 0) passes through. game.exe's filter skips only the LAST unit hit (S); d2d keeps a `struck` list of every unit hit by that missile (S, fight/missiles.cpp:594): a bolt that alternates between two overlapping monsters could hit one twice in game.exe and once in d2d. Rare, structural (D4, no oracle).
- Chain Lightning: NextHit 1 / NextDelay 4: d2d keeps `hit_at` per monster and re-hits after 4 frames (S, :584-589).

## 3. The hit: damage, order of operations

| Step | game.exe | Tag | d2d (upstream main) | Same? |
|---|---|---|---|---|
| element min / max | at creation: FUN_0064b860 -> FUN_00644d50 / FUN_00644e40 (base + level brackets, << HitShift, synergy share by MulDiv when (EMin + brackets) << HitShift > 0x100 or EMinLev1 != 0, then mastery % by MulDiv), stored as the missile's stats 0x30 / 0x31 (fire), 0x32 / 0x33 (lightning), 0x36 / 0x37 (cold) + length 0x38 | S (skill-damage proof, PR #154 merged) | `missile_damage` -> `elem_damage` at strike time with the caster's current env (`Fight::strike`) | yes (116,000 cases 0 differ on #154; re-checked here: min / max columns of the differential below) |
| roll | FUN_005a89a0 -> FUN_005a8910(missile, min stat, max stat, mastery stat): min stat <= 0 or max stat <= 0: 0; swap if min > max; mastery read **on the missile** (0: never set there); `min + FUN_0045c3e0(missile seed, max - min)`: rand(n) = 0 for n < 1, else the seed step's low % n (& (n-1) for a power of two). Record fields: phys +0x08 (stats 0x15 / 0x16, same rule), fire +0x10, magic +0x20, lightning +0x1c, cold +0x24, cold length +0x30, poison +0x28 | V read (0x5a8910..0x5a8999, 0x5a89a0..0x5a8b00) + V run (oracle below) | `missile_blow`: `rng.range(elo, ehi)` = elo + rand(ehi - elo + 1), rolled whenever ehi > 0 | **no: D1** |
| magic damage reduction, resistance, pierce | FUN_0057c1e0 -> FUN_0057bf80: flat reduction first, then resistance less pierce (FUN_0057be00: a plain monster gets no difficulty penalty, no 75 cap; floor -100), MulDiv(fine, 100 - res, 100), res >= 100: 0 | S (life-apply proof, PR #135 merged) | `resisted_fine` | yes (re-checked here: life column) |
| absorb | FUN_0057bf10 (stats 0x90 / 0x91 heal and reduce) | S | not modelled for monsters (no monster absorbs in 1.14d Act 1, I) | out of domain |
| damage dealt | life stat 6 in 256ths minus the total; under 0x100 left: dead (0) | S | `hurt()` with `life_frac` | yes |
| cold side effect | chill length resisted like the damage, freeze rules | S (docs/research/freeze-rules-trace.md, static-field-chill-trace.md) | fix/static-field-chill and fix/freeze-rules candidates | covered there |

Not in play for these skills (D): no physical part (no SrcDam, no min/max damage), no `CltHitSubMissile` / `HitSubMissile` on any of the rows, no SrvHitFunc on the bolts (the client's fireexplode / iceexplode are visual).

## 4. d2d's version (S, upstream main 72132b99)

- `Fight::cast_missile` (components/game/fight/missiles.cpp:218): checks `cost_of` against mana, **pays at once** (`pay`), sets SC (or A1/TH), stores the target; `Fight::strike` (fight/strike.cpp:505) calls `fire` at the action frame.
- `Fight::fire` (:251): plain rows `send(dx, dy, range)` with range = Range + LevRange x lvl; Charged Bolt (do 17) `max(calc1, 1)` bolts with `bolt = k` (the wiggle in `missile_fly`, ai.hpp:591: 77 frames, seed {to_x + k, 666}); Chain Lightning one missile with `hops = calc1`; Nova / Frost Nova 64 missiles round the caster at range Range + calc1.
- `rules::missile_velocity`, `missile_range`, `missile_aim`, `wiggle_points`, `missile_touches` (components/rules/missiles.hpp): the S flight rules.
- `Fight::skill_missile_hits` (:540) -> `strike` (:635) -> `rules::missile_damage` + `rules::missile_blow` (components/rules/combat.hpp:409) -> `land` -> `hurt`.
- Lockstep: d2d steps the game only when stepped (one 40 ms frame per step), so these per-frame rules run once per step.

## 5. The bot's use (our repo)

- `internal/combat/classdata/skills.json` ranges (subtiles, Kolbot's, UNVERIFIED in the file): Fire Bolt 20, Ice Bolt 20, Ice Blast 20, Charged Bolt 10, Lightning 25, Chain Lightning 25, Nova 7, Frost Nova 5. All inside the missile's reach in section 1 (46.9 / 28.1 / 28.1 / ~43 path / 35.2 / 35.2 / - / ~12): the bot casts from closer than it has to (safe; costs walking, docs/research/walk-speed-trace.md).
- `internal/combat/skills.go`: `Cost = ceil(ManaCost(level))` and `PickAttack` refuses a skill whose `Cost > mana`; `internal/data/skills.go` `ManaCost` = `(mana + lvlmana x (lvl - 1)) << manashift / 256`, at least minmana: the game's formula (S: skills.md, V read of FUN_0056c160's arithmetic), rounded up by the bot.
- `internal/drill/d2.go`: Fire Bolt (36) as the cast-mode probe (V live: mode SC ~10 ticks).
- d2d rig / commands: the bot casts through the d2d control socket (`cmd skill <id> x y unit`, rig in internal/d2dcombat); on d2d the mana drops at the command, in game.exe at the action frame (D2).

Bot findings (for the primitives fix queue; none is lethal):

| id | finding | tag | impact |
|---|---|---|---|
| B1 | skill ranges (skills.json, Kolbot) are well inside the missile reach: Fire Bolt 20 vs 46.9, Lightning / Chain Lightning 25 vs 35.2 subtiles; Charged Bolt 10 is right to stay short (the bolts wander) | D + S (rules) | the bot walks closer than needed before casting; walking is 68 % of kill time (walk-speed-trace). Candidate: Fire Bolt / Ice Bolt / Lightning ranges from Missiles.txt (Range x path velocity / 4096) less a margin, after a live M3 check |
| B2 | the mana gate rounds the cost up (`Cost = ceil`, `PickAttack` refuses `Cost > mana`); game.exe compares 256ths (FUN_0056c160) | V read + S | refuses a 2.5-mana Fire Bolt at 2.5..2.99 mana: harmless |
| B3 | `ManaBurnTracker` books a cast's cost at the command (internal/combat/fight.go:710, castfirst.go:166) while game.exe takes it at the action frame and not at all for a cast cut off by a get-hit | S (our code) + V read | the drop can land one sample later (a one-cast phantom burn spike, then a phantom gain) and an interrupted cast books spending that never happened; matters only near the flee / chicken thresholds (Hell) |
| B4 | no model of the missile's flight time: a Fire Bolt at 20 subtiles lands about 21 frames (0.85 s) after the action frame, an Ice Bolt about 36 frames, Lightning about 14 (I: distance / path velocity) | I | a "hit landed?" check right after the cast sees no damage yet; matters for any casts-per-kill or retarget logic that reads the monster's life at once |

## 6. Deviations

| id | what | d2d | game.exe | tag | magnitude / bot impact | candidate |
|---|---|---|---|---|---|---|
| D1 | element roll range | min + rand(max - min + 1), rolled when max > 0 | min + rand(max - min), 0 unless both ends > 0 | V run, real build (104,000 cases: 31,459 -> 0) | 1/512 point lower mean per hit (Charged Bolt 4-8 at level 1: mean 6.000 vs 5.998); no kill-rate effect; changes every roll drawn on a seed (determinism vs game.exe) | `fix/missile-roll-top` (local) |
| D2 | when a missile skill's mana is paid | at the command (`cast_missile`) | in the do, after the missile, at the action frame (FUN_0056f7f0) | V read (game), S (d2d) | a cast cut off before its action frame (get-hit on a big hit, d2d tick.cpp:600) costs mana on d2d, none in game.exe; the mana bar drops ~7-9 ticks later in game.exe (cast-trace) | not built: needs a stateful test (drill D-M below) |
| D3 | Frost Nova / Nova ring | 64 evenly spaced | tables at 0x6e1288 | S (upstream "not yet") | count and spacing of the ring decide how many monsters a nova hits at range | upstream's open item |
| D4 | piercing Lightning re-hit | every unit once per missile | not the LAST unit hit | S read | rare (two overlapping monsters) | none |
| D5 | physical roll | `range(lo, hi) >> 8` (whole points, inclusive) | min + rand(max - min) in 256ths, both ends > 0 | V read | not the Sorceress (no physical part); arrows / throws | listed only |

## 7. Differential (V)

Oracle: `tools/emu/sk_missile_hit.py` (upstream candidate `tools/emu/missile_hit.py`) runs, natively in game.exe under the emulator, FUN_00644d50 / FUN_00644e40 (min / max with synergies and mastery, skill_damage.py's setup), then FUN_005a89a0 on a missile unit carrying only the element's min / max stats and the case's seed, then FUN_0057c6c0 on that record (life_apply.py's setup). The d2d side is `test_missile_hit --missile-hit` (`tools/emu/sk_missile_hit_test.cpp`, upstream `tests/game/test_missile_hit.cpp`): the REAL `rules::missile_damage`, `rules::missile_blow` (Rng set to the case's seed) and `hurt()` with the loaded patched tables. Output `id min max life_after` (256ths) on both sides, `diff`.

Cases (`gen 60000 1`, 104,000): structured 44,000 = 8 skills (Fire Bolt, Charged Bolt, Ice Bolt, Frost Nova, Ice Blast, Nova, Lightning, Chain Lightning) x levels 1..20 x synergy levels 0 / 1 / 5 / 10 / 20 on the skill's own synergy skills x mastery stat 0 / 50 / 98 / 158 / 278 (Lightning Mastery's levels 1, 5, 10, 20: 50 + 12 (lvl - 1), D) x resistance -100 / -50 / 0 / 20 / 25 / 33 / 50 / 75 / 99 / 100 / 150 (pierce 0 / 15 / 60 against a non-immune), life 1..300; plus 60,000 random (levels to 40, mastery to 400, synergies to 40, resistance -100..150, pierce to 100, life 1..400). Seeds: 32-bit low and high words, masked in the Python generator and in the C++ reader. Difficulty does not enter this path for a plain monster (no resistance penalty, FUN_0057be00, S); FCR changes when the missile is made, not its damage (cast-trace, V).

| Build (the worker host) | lines differ of 104,000 | min / max / life | by skill (life) |
|---|---|---|---|
| upstream main 72132b99 + the test mode (BEFORE) | 31,459 | 0 / 0 / 31,459 | Fire Bolt 4,043, Charged Bolt 6,478, Ice Bolt 3,616, Frost Nova 4,184, Ice Blast 2,161, Nova 3,883, Lightning 3,372, Chain Lightning 3,722 |
| + the claim (AFTER, `fix/missile-roll-top` head) | 0 (md5 0c91da88... both sides) | 0 / 0 / 0 | |
| seed 7 (`gen 20000 7`, 64,000 cases) | before 16,479, after 0 | life only | |

The oracle also counts a line-for-line copy of d2d's old roll against game.exe's record: wrong on 103,981 of 104,000; the reading (min + rand(max - min), 0 unless both ends > 0) wrong on 0. The min / max columns equal everywhere: PR #154's synergy gate and MulDiv hold for the plan's skills with masteries.

What the numbers mean for play (I, arithmetic): the mean roll drops by 1/512 point (Charged Bolt 4-8 at level 1: 6.000 -> 5.998; Lightning 1-40: 20.50 -> 20.498); casts per kill and kills per hour for the early Sorceress do not move measurably. The value is fidelity and determinism: with game.exe's draw the same seed gives the same damage.

## 7b. PR candidate `fix/missile-roll-top` (local, the worker host, PR checkout dir `skmiss`; nothing pushed or posted)

Base upstream main 72132b99. Commits authored and committed by the user's numeric-ID noreply identity (copied from fix/static-field), body `Agent: Claude Code`, no trailer:

- `9ba8d369` tooling: `tools/emu/missile_hit.py`, `tests/game/test_missile_hit.cpp` (own file; test_game.cpp is at its length limit), a row in `tools/emu/README.md`.
- `23e2baec` the claim: `rules::missile_blow` rolls `low + rng(high - low)` with low / high the sorted ends, 0 unless low > 0 (components/rules/combat.hpp), a test_combat check (top 256th never rolled over 1,000 seeds; min 0 deals nothing), a "The roll" section in docs/research/re/missiles.md.

Checks (V, the worker host): real builds before / after as section 7 (104,000 and 64,000 cases, 0 after on the final head); pr_guard 0 FAIL / 0 WARN on the range and per commit; presubmit --quick 0 FAIL, 2 WARN (doc-claims "never": FUN_005a8910 has one path, read; build skipped by --quick); ruff (upstream rule set F,E9,B,BLE001,PLW1510,S602,S604) clean; gcc -Wshadow -Wunused -Wsign-compare full build: 0 warnings on added lines; clang-tidy (repo config) on the changed TUs: 0 on added lines after adding `d2s_items.hpp` (misc-include-cleaner); the remaining findings are the pre-existing short binding names in untouched files (outdoor.hpp, merc.hpp, apps/); test_rules, test_combat, test_protocol, test_game (with the data dir) rc 0.

Open: an independent verifier (second agent), the det state-hash run on a patched lab binary (needs the d2d slot; not run by this card), the user's go before any push. Limits for the PR body: the physical roll (`range(lo, hi) >> 8`, D5) is unchanged and outside the claim; every player-side caller of `rules::missile_blow` takes the change (skill missiles, auras and novas in fight/states.cpp, strike.cpp, pets' and the merc's skills, traps in tick.cpp: 12 call sites, S); monsters' missiles on the player use a different function (`game::missile_blow` in ai.hpp, unchanged; their roll is in docs/research/monster-attack-trace.md); d2d draws on the fight's Rng, game.exe on the missile's own seed: the harness sets the Rng to the case's seed, so the claim is the bound, not the seed source.

## 8. Drill recipe (live, a throwaway Sorceress on one instance, offline)

Fork debug verbs by name: `debug spawn`, `debug points`, `debug skill`, `debug missiles`, `debug stat`. MISSING verbs the matrix proposed and these drills need: `debug states`, `debug mon set` (pin a monster's life / resistances), `debug cast` (cast at a unit or point by id, no input), `debug mana full`.

- M1 (D2, mana timing, the main check): `debug skill` Fire Bolt 1, `debug stat` mana known (e.g. 20). Cast at a far point; read mana (stat 8, 256ths) every tick from the command: game.exe drops 640 (2.5 points) on the action-frame tick (SC event; at 0 FCR about tick 8-9 of 13, cast-trace), d2d on the command tick. Then let a monster hit the caster during a cast (`debug spawn` a Fallen next to her, `debug mon set` its damage high enough for a get-hit: more than max life / 12): game.exe: no missile, no mana spent; d2d before a fix: mana spent.
- M2 (Charged Bolt count): levels 1, 5, 20 via `debug skill`; count new missiles with `debug missiles` on the action frame: 3, 7, 22; every bolt's start = the caster's subtile; all expire by 77 frames.
- M3 (reach): Fire Bolt / Ice Bolt / Lightning at an empty point 60 subtiles away; `debug missiles` each frame: the last position about 47 / 28 / 35 subtiles from the start, after 50 / 50 / 25 frames.
- M4 (roll, D1): `debug mon set` a monster's life to 6 points and lightning resistance 0; Charged Bolt 1 (4-8 = 1024..2048 256ths): many single bolts; the life left after one bolt is never exactly 6 x 256 - 2048 in game.exe (the top is exclusive); before the fix d2d leaves it there about 1 in 1025 hits. Statistical only (needs thousands of hits): a weak drill; the oracle is the real proof.
- M5 (Lightning pass-through): two monsters in a line: one Lightning hits both; Chain Lightning at level 1 jumps at most 5 times, never back to the one it just hit when another is in range.

Expected in game.exe: as above. Expected on d2d today: M1 differs (mana at the command, mana lost on an interrupted cast), M2, M3, M5 equal (S), M4 differs only in the top 256th.

## 9. Matrix cell updates (sorc-skill-matrix.md is not on main: apply when it lands)

| Row (matrix) | Cell | New value |
|---|---|---|
| Fire Bolt | missile flight / reach | S (upstream missiles.py) ; reach 46.9 subtiles, 50 frames (D) |
| Fire Bolt | damage roll and resist | V (this trace: 104,000 cases; D1 deviation, candidate fix/missile-roll-top) |
| Fire Bolt, all missile skills | mana timing | V read: paid in the do at the action frame (FUN_0056f7f0), checked at the start (FUN_0056f640); d2d pays at the command (D2) |
| Charged Bolt | bolt count / spawn | V read + D: min(24, 3 + (lvl - 1)) bolts from the caster's subtile at one target point, wiggle by index (FUN_005c9300) |
| Ice Bolt | reach / damage | D reach 28.1 subtiles; V damage chain (chill: freeze-rules-trace) |
| Lightning | pass-through | D CollideKill 0; S last-unit filter vs d2d's struck list (D4) |
| Chain Lightning | first missile, hops | D calc1 = (26 + (lvl - 1)) / 5 hops; S hop search (d2d comments) |
| Frost Nova | ring | U (nova tables 0x6e1288, D3); reach about 12 subtiles (I) |

## 10. Not traced (U)

- The nova tables (0x6e1288): count and directions of Frost Nova's / Nova's ring.
- d2d's last-frame unit test against game.exe's "no unit test on the frame the range ends".
- The exact tick (from the command) of the do in game.exe's server for SC (cast-trace covers the client tick; the server's event-frame path was not re-read).
- FUN_0056d2c0's target point when the target is a unit that moved during the cast.
- Det state-hash on a patched lab binary (needs the d2d slot): open.
