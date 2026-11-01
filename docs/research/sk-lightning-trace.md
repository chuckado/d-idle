# Lightning, Chain Lightning, Lightning Mastery: game.exe 1.14d vs d2d vs the bot (decomp row 17)

> status: card sklight, 2026-11-01, offline (no game, no d2d slot, no Go change). Decomp row 17 (SKLIGHT, sorc-skill-matrix rank 4). Row 18 (SKNOVA: Frost Nova ring, Static Field radius) is NOT taken: its ring tables are a separate path. This trace touches Nova only where it shares code with Chain Lightning: the NextHit rule (D1) applies to every NextHit missile, `nova` and `frostnova` included. Follows docs/research/sk-missile-trace.md (SKMISS: spawn, flight, roll) and docs/research/sk-mana-trace.md (SKMANA).

Tags: **V read** disassembled by this agent with upstream's `tools/emu/disasm.py` on the 1.14d Game.exe (addresses given); **V run** run natively under the emulator (none in this trace); **D** patched 1.14d tables read through upstream's `emu.mpq_file` (Skills.txt, Missiles.txt, States.txt); **S** d2d source (upstream main 72132b99, file:line), upstream docs (`docs/research/re/*.md`) or an earlier trace of ours (named); **I** inference, including the output of the model `tools/emu/sk_light_rehit.py`; **U** not verified.

## Summary

- **Hits per cast and hop rule (V read + D).** Chain Lightning (do 26, FUN_005ca1b0) makes one `chainlightning` missile and stores calc1 = (26 + (lvl - 1)) / 5 in it (5 at level 1, 9 at 20). Hit function 12 (FUN_005aa730) makes the next missile from the hit spot to the next target with one fewer, while more than 1 is left: **calc1 hits per cast at most**. The next target: FUN_0056bd10 over the rooms near the caster, within 20 subtiles (aurarange `par1`; sHitPar1 is 0) of the hit spot, hostile, alive, targetable, not in town, the line from the CASTER clear of mask 4, **not in state 0x56 `justhit`**; then the lowest unit id above the struck one's, else the lowest id up to it. d2d copies all of this except the `justhit` test.
- **NextHit is a unit state in game.exe, a per-missile timer in d2d (D1, V read both sides).** Any hit by a missile whose row has NextHit (FUN_005adb00) puts state 0x56 `justhit` on the unit for NextDelay frames (4 for `chainlightning`, `nova`, `frostnova`). For those frames every NextHit missile, from any cast, passes the unit by, and Chain Lightning's hop search skips it. d2d has no such state. Its NextHit rows keep a `hit_at` list on each missile, so a new hop missile (or another nova missile) may strike the same unit at once. Consequences (I, model): in d2d the hop leaving the struck monster re-hits that monster on its first frame in about 73 % of hops; game.exe never does. In d2d two close monsters also ping-pong, where game.exe's chain ends.
- **Piercing Lightning hits one unit a frame and skips only the last one (D2, V read; SKMISS D4 resolved).** FUN_005ae1f0 tests the subtiles a frame entered and hands **one** unit (the first that passes the filter) to FUN_005adf10, then stops for that frame. The filter skips the missile's last unit hit (missile data +0x18 / +0x1c, active when the row has LastCollide, FUN_0064a470). d2d strikes every touching monster in the same frame, each at most once per missile. Model (I): game.exe deals about 10 % more Lightning hits per cast in packs of 2 to 6 on the line. About 9 % of monsters are hit twice, alternating with a neighbour. About 1.4 % of monsters d2d hits are never hit by game.exe.
- **Damage (S + V read).** Each hop is a new missile whose damage is built at creation from the caster's current stats (FUN_0059fa30 -> FUN_0064b860: synergies, Lightning Mastery %; S, PR #154), rolled by FUN_005a8910 (SKMISS D1: top exclusive), resisted less pierce (S, PR #135). No decay per hop. The hit function runs BEFORE the damage and does not depend on it, so a Chain Lightning that strikes a lightning-immune monster (0 damage) still hops (V read, FUN_005adf10 0x5ae0e6..0x5ae18e). d2d: the same (strike then hop, S).
- **To-hit (V read + D + S).** The missile's to-hit roll (FUN_0057d9b0) is gated by the Missiles.txt ToHit column (+0x18b), not by Skills.txt UseAttackRate. `lightningbolt` and `chainlightning` have no ToHit: they never miss, even though Skills.txt gives both UseAttackRate 1. d2d tests `missile.info->to_hit`: equal.
- **Cast length (S, K1).** Lightning and Chain Lightning are SQ with seqtrans SC: 19 steps at 0 FCR (760 ms) in game.exe. d2d on upstream main casts them as plain SC (`cast_missile` sets kModeSC, `fight.hpp:550` "seqtrans isn't read"). The held local candidate `fix/lightning-sequence` fixes this (D3, existing).
- **Bot (described, not changed).** B1: the one `CastGapMs` 650 for every skill is shorter than the 760 ms Lightning / Chain Lightning cast, the same command-drop trap the knob's own comment warns about. B2: immune targets skip Chain Lightning entirely, though in game.exe its hops reach the pack behind. B3: ranges 25 vs reach 35.2 (SKMISS B1). B4: no use of the hop radius or of lines for Lightning.
- **PR candidate: none.** D1 and D2 are read on both sides and modelled, not proven on a d2d build. The existing harness (`sk_missile_hit`) covers the roll only. How to build a differential is under each deviation.

## 1. game.exe: what a cast makes and how a hop is chosen

| Q | Address | Rule | Tag |
|---|---|---|---|
| Q1 Chain Lightning do | `FUN_005ca1b0` (do 26) | srvmissilea row (Skills +0x48), sets unit +0xc4 bit 6 on the caster, calc1 (Skills +0x138) by FUN_00646ca0, one missile by FUN_0056ecb0, then **FUN_0064a710(missile, calc1)** = missile data +0x28 (the hop count; read back by FUN_0064a730); returns 1 (so the do pays the mana, SKMANA) | V read (0x5ca1b0..0x5ca253) |
| Q1 hits per cast | Skills.txt `ln34 / 5`, Param3 26, Param4 1 | 5 at level 1, 6 at 5, 7 at 10, 8 at 15, 9 at 20 (integer division) | D |
| Q2 hit function 12 | `FUN_005aa730` (ecx game, edx missile, arg struck unit) | owner FUN_00552fd0; hops = +0x28; **hops <= 1: no hop**. Radius = Missiles +0x4c (sHitPar1) when > 0, else the skill's calc at Skills record +0x64 (aurarange) at the missile's skill / level, at least 1. Next = FUN_0056bd10(owner, missile subtile x / y, radius, flags 0x88583, struck unit's id). None or the struck unit itself: no hop. Else a missile is made (FUN_0059fa30) with params flags 0x21, owner, target the next unit, row = the same, start = the missile's subtile, aim = the next unit's subtile, skill and level of the missile; hops - 1 stored. **Returns 3** (bit 1: the missile ends, bit 2: deal the damage) | V read (0x5aa730..0x5aa8a0) |
| Q2 radius | `chainlightning` sHitPar1 0 ("range (0 = skill)"), Chain Lightning aurarangecalc `par1`, Param1 20 | 20 subtiles at every level, measured from the hit spot | D |
| Q2 search | `FUN_0056bd10` -> `FUN_0056b7e0` | flags OR 0xa783 = 0x8a783. Rooms: the OWNER's room (FUN_00620bb0) and its near list (FUN_00619790), town rooms skipped (0x2000, FUN_0061ab00), a room-circle prefilter FUN_0056b770. Units in each room except the owner; distance FUN_006492a0 = dx^2 + dy^2 in subtiles <= radius^2 | V read |
| Q2 filter | `FUN_0056b3e0` with 0x8a783 | 0x1 players, 0x2 monsters (a monster in mode 0 or 12, dying or dead, refused); 0x80 unit +0xc4 bit 2 and 0x400 bit 3 (targetable / hittable flags); 0x100 not standing in town; 0x8000 hostile to the owner (FUN_00554200); **0x80000 not in state 0x56** (FUN_00639df0); 0x200 the line from the OWNER's subtile to the unit's subtile clear of mask 4 (FUN_0064e260) | V read (0x56b3e0..0x56b6a6) |
| Q2 pick | callback `0x56bcd0` | the lowest unit id (+0xc) above the struck id; else the lowest id at or below it (which can be the struck unit, the caller then gives up) | V read |
| Q2 justhit | state 0x56 | States.txt row 86 = `justhit` | D |

Who can be hopped to: the struck monster is always in `justhit` at that moment, because FUN_005adb00 sets it before the hit function runs (section 2). So in game.exe the "itself" fallback only fires when no other monster qualifies. A monster hit by the previous hop less than 4 frames ago is skipped too. That is about 5.6 subtiles of flight at 1.41 subtiles a frame (I).

## 2. game.exe: a missile meeting a unit

| Q | Address | Rule | Tag |
|---|---|---|---|
| Q3 per frame | `FUN_005ae1f0` (srvdofunc 1) | after the move: range - 1 (< 1 ends), units tested only when range left <= the activation word (FUN_0064a640, set at creation to Range - Activate: FUN_0059fa30 0x59fd01..0x59fd0c). Then for each subtile the frame entered (FUN_00648f40) with a collision hit (FUN_0064d9b0 on the CollideType mask): FUN_00641cb0 looks for ONE unit there through the CollideType filter; found: **FUN_005adf10(missile, unit) and return, the frame is over**; none and a wall bit: end | V read (0x5ae1f0..0x5ae3fa) |
| Q3 filter | `FUN_005a8850` -> `FUN_005a8730` (CollideType 3) | unit +0xc4 bits 3 and 2; NextHit row (+0x188): not in state 0x56; **FUN_0064a470: not the missile's last unit hit**; hostile, or the row's CollideFriend (+0x187) | V read |
| Q3 last unit | `FUN_0064a470` / `FUN_0064a400` | one slot (missile data +0x18 type, +0x1c id), compared only when the row's LastCollide flag (Missiles +4 bit 0) is set. Set on every hit (FUN_005adf10 0x5adfae) and at creation to the params' +4 (the owner: FUN_0059fa30 0x59fdbb) | V read |
| Q3 the hit | `FUN_005adf10` | flags: deal 2, end 1 if CollideKill (+0x186). NextHit filter, last-unit filter (return 1, nothing), last unit := this one. ToHit (+0x18b): FUN_0057d9b0(owner, unit, missile stat 0x13); a miss still sets `justhit` and the event, no damage. Then FUN_005adb00, FUN_005c0c30 (the hit event), the SrvHitFunc (table 0x73c840 by +0xe), **whose return replaces the flags** (bit 4: stop), then with bit 2 the roll FUN_005a89a0, the SrvDmgFunc (+0x10), FUN_005adcd0 (apply) | V read (0x5adf10..0x5ae1e9) |
| Q4 NextHit state | `FUN_005adb00` | row NextHit (+0x188) and a unit: a timed stat list ending at the game frame (+0xa8) + NextDelay (+0x189), state 0x56 set on the unit, cleared by callback 0x5adaf0 when it ends | V read; the exact last frame (inclusive / exclusive) I |
| Q4 field names | Missiles record | CollideKill +0x186, CollideFriend +0x187, NextHit +0x188, NextDelay +0x189, ToHit +0x18b, Activate +0x136 | S (upstream re/missiles.md, from FUN_00661b20's column list) |

Rows (D): `lightningbolt` (98) Vel 30, Range 25, Size 2, CollideKill 0, LastCollide 1, no NextHit, no SrvHitFunc. `chainlightning` (93) Vel 30, Range 25, Size 2, CollideKill 1, LastCollide 1, **NextHit 1, NextDelay 4**, SrvHitFunc 12. `nova` (90) Vel 24, Range 13, Accel -1000, Size 1, CollideKill 0, LastCollide 1, NextHit 1, NextDelay 4. `frostnova` (119): the same NextHit 1 / 4. No ToHit and no Activate on any of them.

What this means in play (I):
- Lightning: at most one hit a frame and 25 frames, so 25 hits at most per bolt. A lone monster is hit once: it stays the last unit while the bolt crosses it. Two monsters under the bolt at once take turns, and each can be hit 2-3 times while they stay under its footprint (about 2-3 frames for a size-2 monster at 1.41 subtiles a frame). In a tight clump, a monster passed in a frame that went to a neighbour can be missed.
- Chain Lightning: the hop missile starts on the struck monster but cannot touch it for 4 frames. By then it has flown about 5.6 subtiles. After a hop to a close neighbour (arriving within 4 frames), the previous monster is still `justhit`. Two monsters within a few subtiles of each other are each hit once and the chain ends. It does not bounce A-B-A-B.
- One Chain Lightning protects nothing for long: `justhit` lasts 4 frames, and a second cast 760 ms later is far outside it.

## 3. Damage: synergies, Lightning Mastery, resistance, immunity

| Q | Rule | Tag |
|---|---|---|
| Q5 base | Lightning EMin 1 EMax 40, EMaxLev1..5 8 / 12 / 20 / 28 / 36, HitShift 8; Chain Lightning EMin 1 EMax 40, EMaxLev 11 / 13 / 15 / 15 / 15; both min stay 1 (the min gate) | D; formula S (skill-damage trace K3, PR #154) |
| Q5 synergies | Lightning: (Charged Bolt + Chain Lightning + Nova blvl) x 8 %; Chain Lightning: (Charged Bolt + Lightning + Nova blvl) x 4 %; Charged Bolt: Lightning blvl x 6 % | D (EDmgSymPerCalc) |
| Q5 mastery | Lightning Mastery `passive_ltng_mastery` = ln12 = 50 + 12 (lvl - 1) % (reqlevel 30), on the caster, applied once by MulDiv when the missile is made | D; S (K3) |
| Q5 when built | each hop is a new missile: min / max built from the caster's stats at THAT hop's creation (FUN_0059fa30 -> FUN_0059f900 -> FUN_0064b860); d2d builds at strike time from the current env | S; equal unless a stat changes during the flight (I) |
| Q5 roll | FUN_005a8910: min + rand(max - min), 0 unless both > 0 | V (SKMISS D1, real builds 104,000 cases; candidate fix/missile-roll-top) |
| Q6 resistance | flat magic reduction, resistance less pierce (no 75 cap on a plain monster, floor -100), MulDiv; resistance >= 100: 0 | S (PR #135); immunity not broken by pierce (S) |
| Q6 immune and hops | hit function 12 runs before the damage and whatever it is (0x5ae0e6..0x5ae18e): a Chain Lightning that lands on an immune hops on with its full count | V read |
| Q6 to-hit | Missiles ToHit (+0x18b) decides; `lightningbolt` / `chainlightning` have none: no to-hit roll; Skills UseAttackRate 1 on both rows is not read on this path | V read (0x5ae04c..0x5ae076) + D |

## 4. d2d (upstream main 72132b99)

- `Fight::fire` (components/game/fight/missiles.cpp:322): Chain Lightning `send(dx, dy, range).hops = calc1`. Lightning a plain missile (`skill_missile`: plain = no start / do function).
- `Fight::skill_missile_hits` (missiles.cpp:540): NextHit rows check `missile.hit_at` (per missile, :587-591), strike, push to `struck` only for non-NextHit rows (:596). Hit function 12 (:605-624): hops > 1, radius sHitPar1 else aurarange, rooms near the PLAYER, not town, `targetable()`, `align != 2`, squared subtile distance, `sight_blocked` from the player with mask 4, `rules::ring_pick(found, struck index)` (rules.hpp:2021), none or itself: no hop; launch from the missile's spot with hops - 1. Commit 527f6f30 ("Chain Lightning's hop by FUN_0056b7e0 / FUN_0056bd10") is where this came from.
- The per-frame unit loop (components/game/fight/tick.cpp:394-411): for every monster in index order that `missile_struck` (any subtile entered this frame, `rules::missile_touches`), skip those in `struck`, call `skill_missile_hits`; false (not spent) goes on to the next monster in the SAME frame.
- No `justhit` state anywhere in components/ (grep: only `next_hit` / `next_delay` in gamedata). Upstream re/skills.md:928 describes NextHit as "strike a unit again NextDelay frames on", the per-missile reading. Upstream re/missiles.md:114 notes the client's state 0x56 skip, but the server side was not carried over.
- Ids: d2d's ring pick uses the monster's index in `monsters`. game.exe uses the unit id (+0xc). The two orders agree when the vector is in creation order and ids are given out in creation order (I; not checked).
- Cast: `cast_missile` (missiles.cpp:218) sets kModeSC for every non-A1 / TH missile skill; `start_sequence` reads seqnum but "seqinput / seqtrans aren't read" (fight.hpp:546-551). Lightning / Chain Lightning take the plain SC length.

## 5. Deviations (d2d vs game.exe)

| id | what | game.exe | d2d | tag | magnitude | differential |
|---|---|---|---|---|---|---|
| D1 | NextHit | a hit sets unit state 0x56 `justhit` for NextDelay frames; every NextHit missile skips such a unit, and Chain Lightning's hop search skips it (flag 0x80000) | per-missile `hit_at`; no unit state; the hop search may find the monster just struck or the one before | V read both sides | Model (I, `sk_light_rehit.py` scenario B, 20,000 hops, seeds 1 and 7): the hop's first hit is the monster just struck in 73.2 % of hops in d2d, 0 % in game.exe (window 3 or 4 alike). By angle from the struck monster to the next target: 0-45 degrees 100 %, 135-180 degrees 28 %. So d2d's Chain Lightning spends most of its hops on the first monster, with full damage each time. Also, by reading (not modelled): two close monsters ping-pong A-B-A in d2d where game.exe stops. The same rule decides how often a nova's 64 d2d missiles may each hit one monster (`nova`, `frostnova` are NextHit rows: hand to SKNOVA) | **none exists.** Buildable: (a) game.exe side natively under the emulator: FUN_005adf10 with a fake missile row (NextHit 1, NextDelay 4) and two units, calling it twice at frames f and f + k, plus FUN_0056bd10 with stubbed rooms (sk_tele_spot.py's room stubs) and a unit in state 0x56. Oracle: hit / no hit, chosen id. (b) d2d side: a test mode in tests/game driving `Fight` with 2-3 monsters at given subtiles, one Chain Lightning cast, printing every strike (monster, frame). Inputs: monster count 1-4, sizes 1-3, positions within 25 subtiles, hops 2-9. This is the stateful harness SKMANA also lacks |
| D2 | piercing missile, units per frame | one unit a frame (the first that passes the filter in the first subtile with a unit), the last unit skipped | every touching unit in the same frame, each once per missile (`struck`) | V read both sides | Model (I, scenario A, 20,000 casts): game.exe 3.15 hits per cast vs d2d 2.86 (+10 %), packs of 6: 4.86 vs 4.28. 8.8 % of monsters hit twice by game.exe, 1.4 % hit by d2d and missed by game.exe; 26 % of casts differ somewhere. The game.exe order of units within one subtile is U (index order in the model) | **none exists.** Buildable like D1(a): FUN_005ae1f0 on a missile with a stubbed room collision grid and N units (upstream missiles.py already runs the flight). Oracle: the unit handed to FUN_005adf10 each frame. d2d side: the same test mode as D1(b) with Lightning |
| D3 | Lightning / Chain Lightning cast length | SQ, seqtrans SC: 19 steps at 0 FCR, 11 at 194 | plain SC (13 ticks after PR #128) | V (K1, cast-trace-gameexe.md) + S | -32 % cast time on d2d for the lightning plan's main attack: d2d over-rates kills per hour from clvl 26 | **exists**: `cast_rates.py` sequence lines; local candidate `fix/lightning-sequence` 2036fd3f (held for the maintainer, not mine; empty diff after) |
| D4 | hop pick order | unit id (+0xc) | index in `monsters` | S + V read | equal while both are creation order (I) | small: one check that d2d's monster vector order is the id order after deaths / spawns |
| D5 | hop missile damage build time | at the hop's creation | at strike time | S | stat changes in flight only | not worth one |

Equal (no deviation found): hop count calc1 and the "hops > 1" test; radius 20 from the hit spot; rooms near the caster; the line from the caster with mask 4; town excluded; the hop starts at the missile's subtile and aims at the next unit's subtile with the row's Range; no to-hit roll; the hop does not depend on the damage (immune still hops); damage formula and mastery (S, #154).

## 6. The bot (internal/combat; described, not changed)

| id | finding | where | tag | impact |
|---|---|---|---|---|
| B1 | one cast gap for every skill, `CastGapMs` 650, while Lightning / Chain Lightning take 19 frames = 760 ms at 0 FCR (and 18 from FCR 7). The knob's own comment says it MUST stay above the cast animation, or commands are dropped mid-cast (1.6-1.75 commands per real cast at the old 450) | internal/env/tuning_combat.go:38-42, :79; internal/combat/fight.go:717-719 | S (our code) + V (K1) | from clvl 26 (Chain Lightning first in the order) about every second command may land mid-cast on game.exe at low FCR; on d2d without fix/lightning-sequence the cast is 13 ticks and the bug is hidden (D3). Candidate: a per-skill gap from the cast length (K1 formula) |
| B2 | an immune target drops Chain Lightning entirely (`PickAttack` skips any skill whose element the target is immune to) | internal/combat/skills.go:179; target.go:102 | S + V read (Q6) | in game.exe a Chain Lightning cast at a lightning immune still hops at full count to the monsters around it (0 damage on the immune). The bot instead falls to Fire Bolt / ignores the pack member. Matters from NM (lightning immunes); opportunity, not a bug |
| B3 | Lightning / Chain Lightning range 25 subtiles vs 35.2 reach | internal/combat/classdata/skills.json:1 | D + S (SKMISS B1) | restated; Chain Lightning's real reach is 35.2 to the first hit plus 20 per hop |
| B4 | no model of the hop (calc1 hits, 20-subtile radius) or of Lightning's line: the attack order (Chain Lightning, Lightning, Charged Bolt, Fire Bolt) is fixed and every cast aims at one target | internal/combat/classdata/profiles.json:47-56 | S | Chain Lightning is the right first pick for packs; for a lone target Lightning (same 1-40 base, +8 % synergy vs +4 %, mana 8 + 0.5 / lvl vs 9 + 1 / lvl) is cheaper and stronger per cast. Tuning, not a fidelity bug |
| B5 | mana gate and cost: `Cost = ceil(ManaCost)` (Lightning manashift 7: 8 + 0.5 (lvl - 1), Chain Lightning 9 + (lvl - 1)) | internal/combat/skills.go, internal/data/skills.go | S (SKMISS B2, SKMANA) | equal up to the round-up; nothing new |

## 7. PR candidates

No PR candidate. D1 and D2 are read on both sides and quantified only by a model (`sk_light_rehit.py`: line copies of d2d's flight and footprint rules plus a model of the game.exe rule, tag I). The existing differential (`tools/emu/sk_missile_hit.py` + `sk_missile_hit_test.cpp`) covers the damage roll, not flights or hits per frame, so it cannot show mismatches before > 0 and after = 0 for these. D3 already has its own held candidate (`fix/lightning-sequence`). No fork branch was made.

## 8. The model run (I)

`python3 tools/emu/sk_light_rehit.py 20000 1` (and seed 7): pure Python, no game code run.

Scenario A (Lightning, 25 frames, 2-6 monsters of size 1-3 within 3 subtiles of the line):

| monsters on the line | casts | hits / cast game.exe rule | d2d rule | casts that differ | monsters hit 2+ (game.exe) | hit by d2d, missed by game.exe |
|---|---|---|---|---|---|---|
| 2 | 3,919 | 1.451 | 1.394 | 4.4 % | 3.1 % | 0.2 % |
| 4 | 4,035 | 3.110 | 2.840 | 24.5 % | 7.9 % | 1.2 % |
| 6 | 4,029 | 4.861 | 4.276 | 49.4 % | 11.8 % | 2.1 % |
| all | 20,000 | 3.154 | 2.856 | 26.1 % | 8.8 % | 1.4 % |

Seed 7: 3.142 vs 2.857, 25.5 % differ. A lone monster is hit exactly once by both rules (checked, sizes 1-3).

Scenario B (Chain Lightning, the first hop after a hit; B within 20 subtiles of the hit spot): first hit is the struck monster in 73.2 % (seed 1) / 73.3 % (seed 7) of hops with the d2d rule, always on the hop's first frame; 0.00 % with the game.exe rule. The game.exe result is geometric: after 4 frames the hop is about 5.6 subtiles away and out of the struck monster's footprint.

Limits: open room (no walls), units in index order within one subtile, no monster movement, ping-pong over several hops not simulated.

## 9. Drill recipe (live, later; writing it is allowed, running it is not)

Fork verbs (docs/forkdbg.md, lab binary per docs/d2d-fork.md and docs/forkcheats.md): `debug stat 12 <clvl>`, `debug points <id> <n>` (Lightning 49, Chain Lightning 53, Lightning Mastery 63), `debug mana full`, `debug spawn <monster> <n> near`, `debug mon set <id> life <n> maxlife <n> res light <r>`, `debug cast <skill> unit <id>` / `x y`, `step N`, `monsters`, `debug missiles`. On d2d these show d2d's behaviour; the game.exe side needs a hidden game.exe instance and memory reads of monster life only (no debug verbs there).

- L1 (D1, the hop re-hit): two Fallen, A and B, 6 subtiles apart, both `debug mon set life 1000 maxlife 1000 res light 0`. Chain Lightning level 1 (5 hits) cast at A. Read both monsters' life every `step 1`, and use `debug missiles` for the hop missiles. game.exe expects A hit once, then B, then (A no longer `justhit` if the hop took more than 4 frames) A again, alternating, at most 5 hits. With A and B 3 subtiles apart, expect A, B and the chain ending (A still `justhit`). d2d today expects A hit 2+ times in a row in most casts.
- L2 (D1, nova share): one Fallen at 3 subtiles, life 1000, Nova level 1. game.exe expects one hit per cast. d2d expects as many hits as ring missiles touch it (count the life drops). Hand to SKNOVA.
- L3 (D2, Lightning per frame): three Fallen in a line 5, 6 and 7 subtiles from the caster, life 1000, res light 0, Lightning level 1, 20 casts. Count life drops per monster per cast. game.exe expects alternating double hits on neighbours. d2d expects exactly one hit each per cast.
- L4 (immune hop, Q6): A with `res light 100`, B and C normal within 10 subtiles. Chain Lightning at A: A takes 0, and B and C take hits in both game.exe and d2d.
- L5 (B1, cast gap): a game.exe instance, Sorceress with 0 FCR, Chain Lightning on a pack. Count cast commands vs SC/SQ mode starts in the bot log. Expect more than 1.0 commands per real cast at `castGapMs` 650, and 1.0 at 800.
- L6 (D3): cast length in ticks on the lab binary with and without fix/lightning-sequence: 19 / 13 at 0 FCR.

## 10. Matrix-cell updates (docs/research/sorc-skill-matrix.md section 3; proposed text)

| Row | Cell | New value |
|---|---|---|
| 3.9 Chain Lightning | (a) hops | calc1 = (26 + (lvl - 1)) / 5 hits per cast (5..9), next target within 20 subtiles of the hit spot, rooms near the caster, line from the caster clear of mask 4, not in `justhit`, next unit id above the struck one else the lowest [V read FUN_005ca1b0, FUN_005aa730, FUN_0056bd10, FUN_0056b3e0; sk-lightning-trace.md Q1-Q2] |
| 3.9 Chain Lightning | (b) d2d status | hop rule copied (527f6f30) except state `justhit` (D1: hop re-hits the struck monster, ping-pong; model 73 % of hops) [V read both sides, I magnitude]; cast plain SC (D3, held fix/lightning-sequence) |
| 3.9 Chain Lightning | (c) bot use | add: cast gap 650 < 760 ms cast (B1); skipped on immunes though it hops (B2) [S] |
| 3.10 Lightning | (a) behaviour | one unit a frame, the last unit hit skipped (LastCollide), CollideKill 0, no NextHit, no to-hit, 25 frames, reach 35.2 subtiles [V read FUN_005ae1f0, FUN_005adf10, FUN_0064a470; D] |
| 3.10 Lightning | (b) d2d status | every touching monster each frame, once per missile (D2: model -10 % hits per cast in packs) [V read both sides, I magnitude] |
| 3.8 Lightning Mastery | (a) rules | unchanged; add "applied at each hop's creation (each hop a new missile)" [S] |
| 3.4 Frost Nova, Nova | (a) behaviour | add: NextHit 1 / 4 = unit state `justhit` in game.exe, per-missile in d2d (D1 of sk-lightning-trace.md; ring count still SKNOVA) [V read] |
| section 5 rank 4 | status | done 2026-11-01 (SKLIGHT): sk-lightning-trace.md |

## 11. Not traced (U) and dead ends

- The order of units FUN_00641cb0 returns within one subtile (D2's tie rule).
- The exact frame `justhit` ends (FUN_006251f0's expiry: frame + NextDelay inclusive or exclusive). The model shows no sensitivity between 3 and 4.
- Whether the missile leaves at the sequence's event-1 step for SQ casts (cast-trace U, unchanged).
- Dead end: Skills.txt UseAttackRate on Lightning / Chain Lightning looked like a to-hit roll; it is not read on the missile hit path (ToHit is the Missiles column, Q6).
- Dead end: first read of the hop flags as 0x88583 (the push). FUN_0056bd10 ORs 0xa783, so the line test (0x200) is in. d2d's comment (0x8a783) is right.

## Shared-file lines for the merger

docs/decomp-queue.md, row 17, replace the status cell with:

`done 2026-11-01 (SKLIGHT, Opus): do 26 FUN_005ca1b0 / hit 12 FUN_005aa730 / hop search FUN_0056bd10 + FUN_0056b3e0 / collision FUN_005ae1f0 + FUN_005adf10 read: calc1 = (26 + lvl - 1) / 5 hits, radius 20 from the hit spot, line from the caster, next id above the struck; D1 NextHit is unit state 0x56 justhit in game.exe, per missile in d2d (model: d2d hop re-hits the struck monster in 73 % of hops, ping-pong); D2 Lightning one unit a frame, last unit skipped vs d2d all once (model +10 % hits in game.exe); D3 cast SC vs SQ 19 (held fix/lightning-sequence); immune still hops (V read); no PR candidate (no flight harness); bot B1 cast gap 650 < 760 ms, B2 immune skips Chain Lightning; docs/research/sk-lightning-trace.md; tools/emu/sk_light_rehit.py`

docs/decomp-queue.md, results list, add:

`- 17 Lightning / Chain Lightning (SKLIGHT, 2026-11-01): game.exe hits a unit with at most one missile hit a frame (FUN_005ae1f0), skips the missile's last unit (LastCollide, FUN_0064a470), and NextHit sets unit state 0x56 justhit for NextDelay frames (FUN_005adb00) that every NextHit missile and the hop search (flag 0x80000) skip; d2d keeps per-missile hit_at and hits every touching monster once: D1 (hop re-hits / ping-pong, also the novas), D2 (Lightning per frame), model only (sk_light_rehit.py), differentials buildable (stateful Fight test mode + FUN_005adf10 / FUN_0056bd10 under the emulator); bot B1-B5`

docs/BACKLOG.md, add:

`- SKLIGHT follow-ups: build the stateful d2d Fight test mode + game.exe FUN_005adf10 / FUN_0056bd10 oracle for D1 (justhit) and D2 (one unit a frame); SKNOVA to check D1 on the 64-missile nova; bot B1 per-skill cast gap (Lightning / Chain Lightning 760 ms at 0 FCR) and B2 Chain Lightning on immunes (docs/research/sk-lightning-trace.md)`

docs/INDEX.md, research section, add:

`- docs/research/sk-lightning-trace.md: Lightning / Chain Lightning hops, NextHit justhit state, one hit a frame; d2d D1-D5, bot B1-B5, drills L1-L6 (SKLIGHT)`

docs/research/README.md, add:

`- sk-lightning-trace.md (SKLIGHT, 2026-11-01): Chain Lightning hop rule and NextHit as unit state justhit; Lightning one unit a frame; no PR candidate`

CHANGELOG.md: none (docs and a model script only).

Glossary: `justhit` (state 0x56, the NextHit unit state) and "hop" (one Chain Lightning missile after the first) are not in GLOSSARY.md; proposed for the coordinator, not added.
