# Sorceress armors and Energy Shield: game.exe 1.14d vs d2d vs the bot (decomp row 15)

> status: card skarmor, 2026-11-01, offline (no game, no d2d slot, no Go change). Decomp row 15 (SKARMOR, sorc-skill-matrix rank 2): Frozen Armor, Shiver Armor, Chilling Armor and Energy Shield (Cold Mastery where it touches them): the state each sets, defence / absorb and their scaling, durations, the striker effects and their exceptions, recast and replacement, the state ids the bot can read; d2d's version; the bot's use. Follows docs/research/freeze-rules-trace.md (DECOMPFREEZE: the freeze rule C3 / C5 and the oracle reused here), docs/research/sk-mana-trace.md (when the mana is paid) and docs/fixrecast.md (FIXRECAST: the durations and the recast timer).

Tags: **V read** disassembled with upstream's `tools/emu/disasm.py` by this agent (addresses given); **V run** run natively in game.exe under the emulator by this agent (`tools/emu/sk_armor.py`); **D** from the patched 1.14d tables as game.exe loads them (Skills.txt, States.txt, Events.txt read from the MPQ stack by the emulator's own loader); **S** d2d source (upstream main 72132b99) or an earlier trace (named); **I** inference; **U** not verified.

## Summary

- **Frozen Armor's striker effect is the freeze rule, nothing else (V run).** Event 2 (`damagedinmelee`, FUN_005cac50) builds an empty hit record with only the freeze length (+0x34 = calc1) and the "no on-hit events" flag (+4 |= 0x20) and applies it with FUN_0057c6c0 (flag 1): the length is cut by the striker's cold resistance less the wearer's cold pierce (Cold Mastery), and FUN_0057b230 decides: a plain monster is frozen only at coldeffect < 0, for length / MonsterFreezeDivisor, and **not chilled**; a champion, unique, superunique, act boss (MonStats boss) or hireling is **chilled, never frozen**; a cold-immune striker gets nothing. Run natively on 184,771 cases: equal to the freeze oracle `freeze_rules.py` on every case (0 mismatches). Only a blow whose physical damage (+8) is above 0 triggers it.
- **d2d (S)** `Fight::buff_events` stuns AND chills every damaging striker for the raw calc1 frames: any kind, any coldeffect, any resistance (Andariel, Bishibosh, cold immunes included). Against the native rule every damaging case differs (D1). The freeze PRs (#155 / #156 line, freeze-rules-trace C3 / C5) fix the missile path only; this path is not covered. No d2d-side harness runs `buff_events`, so **no PR candidate** (section 8).
- **One armor at a time (V read + D).** Do 18 (FUN_005c9480) first calls FUN_0056c740, which removes every state of the new state's States.txt `group` (frozenarmor 10, chillingarmor 20, shiverarmor 88, bonearmor 14 are all group 1), the old copy of itself included, then creates the state with the full auralen. So **a recast always restarts the full duration** (and any lower level replaces a higher one), and Shiver replaces Frozen. d2d removes only the same skill's state: Frozen + Shiver + Chilling stack, their armor % add up (D2).
- **Energy Shield (V read)** FUN_005ca6c0: per damage field (physical, fire, lightning, cold, magic, plus three fields only from a non-hireling monster), absorbed = `min(field x calc1 / 100, mana x 16 / calc2)`, mana -= `absorbed x calc2 / 16`, all in 256ths; **at 0 mana the state is removed**. d2d works on the total in whole points and never removes the state (D3, D4).
- **Durations (D, matches FIXRECAST):** Frozen / Shiver 3000 + 300 (L-1) frames + 250 per point in the two other armors (120 s / 132 s at L1 / L2, +10 s a point); Chilling 3600 + 150 (L-1) + 250 a point (144 s / 150 s); Energy Shield 3600 + 1500 (L-1) (144 s / 204 s), no synergy.
- **Bot (V read of our code):** the recast timer is right, but `Fighter.Buff` runs only when `needBuff` is set (new game, back from town, goal reset, death): **inside one long clear the armor lapses at 120 s and is not recast until the next town trip** (B1). Shiver / Chilling / Energy Shield states are not in `buffStates` (B2); the d2d backend never fills `Player.States` (B3).

## Q1. Do 18: the state, its length, the armor percent (V read + D)

| Address | Rule | Tag |
|---|---|---|
| `FUN_005c9480` (ECX game, EDX unit, skill, level) | Skills row; aurastate (+0x80) must be a valid state. Unit +0xc4 \|= 0x40 (the "missile made" flag of SKMANA: the do counts as done, mana paid). `FUN_0056c740(unit, state, 1)`: group removal (Q4). A 0x24 block {unit, unit, skill, level, length = calc(Skills +0x60 auralencalc, unit, skill, level), -1, 0, state, callback 0x5c9420} to `FUN_0056e970` (create the timed state); `FUN_005c6cc0` and `FUN_005c6dc0` (the aurastats onto the state's stat list), stats 0x15e = skill, 0x15f = level on that list; the aura events: `FUN_005c0b50(unit, 1, state)`, then for each of up to three (Skills +0x84 event, +0x8a function) `FUN_0056e740`; `FUN_00639e30(unit, state, 1)` | V read (0x5c9480..0x5c9636) |
| `FUN_0056e970` | for a curse (FUN_0063a420) length cut by stat 0x6d; existing list of the same state: same skill and level: end := now + length (refresh, event 0xc rescheduled); same skill, new level LOWER: refused (returns 0); otherwise the old list is removed and a new one made | V read (0x56e970..0x56eae6) |
| auralencalc | Frozen Armor `ln34 + (Shiver blvl + Chilling blvl) x par7`: 3000 + 300 (L-1) + 250 per point; Shiver the same with (Frozen + Chilling); Chilling `ln34` = 3600 + 150 (L-1) + 250 x (Frozen + Shiver); Energy Shield `ln12` = 3600 + 1500 (L-1) | D |
| armor percent | aurastat1 `skill_armor_percent` (stat 171) = ln12: Frozen 30 + 5 (L-1), Shiver 45 + 6 (L-1), Chilling 45 + 5 (L-1) | D |
| how stat 171 enters the defence | not read here | U |
| level used | FUN_005c9480's level argument (the cast level, +skills included: SKMANA FUN_006442a0 flag 1); synergies use blvl (hard points) | I (from the calc text and SKMANA) |
| mana | Frozen 7, Shiver 11, Chilling 17, Energy Shield 5 points flat (lvlmana 0), paid in the do at the action frame | S (sk-mana-trace) + D |
| in town | InTown 1 for all four | D |

States the bot can read (D, States.txt row = id): **frozenarmor 10, chillingarmor 20, energyshield 30, shiverarmor 88** (bonearmor 14). Group 1: 10, 14, 20, 88; energyshield has no group. The freeze is state 1 (`freeze`), the chill state 11 (`cold`).

## Q2. Frozen Armor's striker effect (event 2)

| Address | Rule | Tag |
|---|---|---|
| `FUN_005cac50` (ECX game; wearer, striker, the striker's hit record, skill, level) | wearer and striker present, striker a player (0) or monster (1); the hit record's +8 (physical damage, 256ths) > 0, else nothing. A zeroed 0x70 record, +4 \|= 0x20; +0x34 = calc(Skills +0x138 calc1, wearer, skill, level); `FUN_0057c6c0(game, wearer -> striker, 1, record)`; the overlay Skills +0x110 on the striker | V read (0x5cac50..0x5cad33); V run (the oracle passes through it) |
| `FUN_0057c6c0` | hostile check; flag 1 runs `FUN_0057c1e0` (the lengths resisted, row 6 of 0x732980: +0x34 by cold res 0x2b less cold pierce 0x14f); record +4 & 0x20 skips the on-hit events (5 / 1 melee, 6 / 2 missile, 10 / 9 at the end): the reflected hit fires nothing; then FUN_0057af80 (+0x30 = 0: no chill from the length) and **FUN_0057b230(+0x34)** | V read (0x57c6c0..0x57c97b) |
| `FUN_0057b230` | as freeze-rules-trace.md: player striker: a chill; monster with state 0x36: nothing; MonStats boss / monster-data flag 8 (champions, uniques, superuniques, act bosses) / hireling: chill for the length (FUN_0057af80: / MonsterColdDivisor at coldeffect < 0, nothing at coldeffect 0), no freeze; else coldeffect >= 0: nothing; else freeze (state 1) for length / MonsterFreezeDivisor, no chill | S (freeze-rules-trace, V run there) + V run here |
| calc1 | `ln56 x (100 + (Shiver blvl + Chilling blvl) x par8) / 100`: (30 + 3 (L-1)) x (100 + 5 per point) / 100 frames: L1 30 (1.2 s), L2 33, L20 87 | D (arithmetic I) |
| Cold Mastery | passive_cold_pierce (stat 335 = 0x14f) on the wearer lowers the striker's cold resistance for the length (no effect at res >= 100: immune) | D + V read (row 6 pierce 0x14f) |

Oracle (V run): `tools/emu/sk_armor.py game` runs FUN_005cac50 natively on top of `freeze_rules.Game` (the same hooks: stats, DifficultyLevels row, MonStats row, the state lists), with the calc hooked to the case's frames (every call checked to be calc1 of skill 40: 184,765 calls, 0 others) and the overlay call stubbed. Cases (`gen`): 3 difficulties x 5 monster kinds x boss column x 6 coldeffects x 13 resistances x every calc1 value of levels 1..20 with 0 / 1 / 5 / 20 / 40 synergy points, a pierce block, damage-0 cases, 4,000 random: **184,771**. On the worker host (upstream main 72132b99 read worktree, the emulator's MPQ stack):

| comparison | mismatches |
|---|---|
| `game` (FUN_005cac50) vs `direct` (freeze_rules.py: FUN_0057c6c0 on a record with +0x34 = frames, no cold damage) | **0 of 184,771** |
| `game` vs `model` (d2d's buff_events as read: stun and chill = max(frames, 1) on any damaging striker; NOT a d2d run) | 184,765 of 184,771 (every damaging case): specials 146,061 (chill only, never frozen), plain monsters 15,741 (freeze only, resisted and divided, never chilled), coldeffect >= 0 14,300 (nothing), cold immune 8,663 (nothing) |

(The bucket sizes follow the case mix, not the game; the point is that game.exe never sets both states and d2d always does.) Case file sha256 prefix f626469a14b8e962, game output e3db7ed22b8d1a4d (worker host, `~/d2d-lab/pr/skarmor-out/`).

What the early Sorceress gets (I, from the rule and D): an Act 1 Fallen (cold res 0, coldeffect < 0) hitting her in Normal is frozen 30 frames (1.2 s) at level 1, 15 in Nightmare, 7 in Hell; Bishibosh, Corpsefire, Blood Raven, a champion pack and Andariel are at most chilled (FUN_0057af80: not at coldeffect 0), never stopped; a cold-immune striker gets nothing; a striker whose hit was only elemental (no physical) triggers nothing.

## Q3. Shiver Armor (event 3) and Chilling Armor (event 1)

| Address | Rule | Tag |
|---|---|---|
| `FUN_005cad40` (Shiver, `attackedinmelee`) | striker a player or monster; a zeroed record filled by `FUN_0056e0c0` (the skill's element: cold EMin 12..EMax 16 by level, HitShift 7, ELen 100 + 25 / 50 a level, synergy 9 % a point of Frozen + Chilling); +0x60 \|= 0xd, +0x64 = 1, +4 = 0x4021 (hit, no on-hit events); `FUN_0057c6c0(game, wearer -> striker, 1, record)` (resisted cold damage and chill length, FUN_0057af80's chill rules), `FUN_0057cee0`, the overlay | V read (0x5cad40..0x5cae2e) + D |
| when event 3 fires (on a miss too?) | not in FUN_0057c6c0 (which fires 1 / 2 / 5 / 6 / 9 / 10 only); the caller not found here | U |
| `FUN_005cab40` (Chilling, `hitbymissile`) | the missile's owner (FUN_00552fd0), hostile (FUN_00554200), the missile's row ReturnFire (Missiles +5 & [0x6ce26c] = 2); then the skill's srvmissilea (`chillingarmorbolt`, +0x48) from the wearer at the owner's position (FUN_0045adf0 / FUN_0045ae20), level as cast (FUN_0059fa30, flags 0x20) | V read (0x5cab40..0x5cac43) + D |
| events table | Events.txt rows: 0 hitbymissile, 1 damagedinmelee, 2 damagedbymissile, 3 attackedinmelee, 5 domeleedamage, 6 domissiledamage, 9 kill, 10 killed, 11 absorbdamage | D |

d2d (S): Shiver `land(missile_blow(missile_damage(...)))` on every `melee_by` striker (ai.cpp:2697; pushed on every resolved swing, hit or not, as far as read: I); Chilling `return_fire` launches `srvmissilea` at the shooter's spot for each ReturnFire owner. Both match the reads in structure (I); the damage and chill go through the missile path SKMISS / chill PR #155 cover. Not run.

## Q4. Recast, replacement and stacking (V read + D)

- `FUN_0056c740(unit, state, include_self)`: States.txt row (stride 0x3c) +0x1e = group; group 0: nothing. Else every state with that group the unit has (the new one too, include_self = 1 from do 18) is switched off (FUN_00639db0 off) and its list freed (FUN_006277e0, FUN_00626cd0). V read (0x56c740..0x56c83a).
- So for the three armors: a recast of the same armor at any time restarts the full duration (the old list is gone before FUN_0056e970 runs, so even its "lower level refused" branch never triggers); casting another armor removes the first. Mana is the only cost of an early recast.
- Energy Shield (do 23, `FUN_005c9c10`): no group call (and no group); FUN_0056e970 refreshes at the same level, replaces at a higher level, **refuses at a lower level** (e.g. after a +skills item came off). V read.
- d2d `Fight::start_state` (fight/states.cpp:99): `erase_if(state.skill == skill.id)` then push: same-skill refresh (equal), but no group: Frozen + Shiver + Chilling coexist and `update_fighters` (strike.cpp:133-139) sums every self state's aurastats, so their armor % add; a lower-level Energy Shield replaces. S.
- d2d applies Frozen Armor's percent to the defence it reports: fork vetting `debug states set 40` def 6 -> 7 (6 x 1.30 truncated). S (docs/forkdbg.md vetting, live on d2d).

## Q5. Energy Shield's absorb (event 24)

| Address | Rule | Tag |
|---|---|---|
| `FUN_005ca6c0` (ECX game; wearer, attacker, record, skill, level) | wearer has the state; pct = calc1 (+0x138, `min(edmn, 95)`; EMin 20, EMinLev1 5 for levels 2-8, ...: D, the edmn scaling I); pct <= 0: nothing. mana = stat 8 (256ths); ratio = calc2 (+0x13c, `par5 - Telekinesis blvl` = 32 - TK), at least 1. "from a monster" = no attacker, or a monster that is not a hireling (FUN_0063ee90) | V read (0x5ca6c0..0x5ca790) |
| the loop | table 0x6e30e0, 8 x {record offset, monster-only}: {+8, +0x10, +0x1c, +0x24, +0x20} always, {+0x38, +0x3c, +0x40} only from a monster. Per field > 0: a = MulDiv(field, pct, 100); a = min(a, MulDiv(mana, 16, ratio)); field -= a; mana -= MulDiv(a, ratio, 16), floored at 0 | V read (0x5ca7b1..0x5ca837) + V run (table bytes read in the image) |
| after | stat 8 := mana; the prgoverlay (+0x10c) by direction when anything was absorbed; **mana <= 0: the state's list is freed (the shield ends)** | V read (0x5ca839..0x5ca8f0) |
| field names | +8 physical, +0x10 fire, +0x1c lightning, +0x24 cold (S, upstream skills.md: Bone Armor / Cyclone Armor); +0x20 magic, +0x38 / +0x3c / +0x40 (I: the leech / drain fields); poison is not in the table (V read: no poison field) | S / I |
| order against resistances and other absorbs | event 11 is fired outside FUN_0057c6c0; not traced | U |

d2d (S, fight/states.cpp:131-145): `part = damage x clamp(calc1, 0, 95) / 100` on the total `foe.damage` (points), `can = min(part, (mana >> 8) x 16 / per16)`, mana -= `(can x per16 / 16) << 8`; the state is never removed at 0 mana.

## 6. Deviations (d2d vs game.exe)

| id | what | d2d (S, 72132b99) | game.exe | tag | differential under tools/emu | bot impact |
|---|---|---|---|---|---|---|
| D1 | Frozen Armor striker effect | `buff_events` (fight/states.cpp:150-161; merc copy `merc_buff_events` :186-197): stun_until and chill_until = now + max(calc1, 1) frames on every damaging striker, no resistance, no MonsterFreezeDivisor, any coldeffect, bosses / champions / uniques / immunes included | FUN_005cac50 -> FUN_0057c6c0 -> FUN_0057b230: resisted by cold res less pierce; plain monster frozen only at coldeffect < 0 for length / MonsterFreezeDivisor and not chilled; specials chilled only; immune nothing | game V run (184,771 cases, = freeze oracle on all); d2d S | game side yes (`sk_armor.py game`); d2d side **none**: to build, a test mode beside `test_freeze_rules` that makes a Fight with a frozenarmor self state, one monster from the case (kind flags, MonStats boss byte, coldeffect, cold res, difficulty), a Foe with the monster in `melee_by` and `damaged_by`, calls `buff_events`, prints the chill and stun ends in ticks; oracle `sk_armor.py game`, cases `sk_armor.py gen` | d2d is much kinder: Andariel and every champion stand for 1.2 s per damaging melee hit; NM / Hell freezes 2x / 4x too long; cold immunes stopped. Inflates d2d's melee-survival and Andariel numbers for any Sorceress with the armor up |
| D2 | armor group | same skill only; Frozen + Shiver + Chilling stack, armor % summed | one group-1 state at a time (FUN_0056c740 + States group 1) | V read + D / S | none (stateful); drill A2 | none today (the bot casts one armor per group, B4) |
| D3 | Energy Shield at 0 mana | state stays; absorbs again as mana regenerates | state removed (FUN_005ca6c0 0x5ca8b0) | V read / S | none; drill E1 | with ES learned, d2d keeps a shield the game drops; the bot's timer (B2) would not see the drop in game.exe either |
| D4 | Energy Shield arithmetic | total damage, whole points, mana floored to points, cost truncated to points | per field, 256ths, MulDiv per field; the monster-only fields; poison never | V read / S | could be built as a pure-function differential: game `FUN_005ca6c0` natively (hooks: FUN_006256b0 state present, FUN_00646ca0 calc1 / calc2, stat get / set) on a record with the case's fields and mana, vs a d2d test of `Fight::absorb`'s Energy Shield block (needs factoring it out: it is a member) | small per hit; d2d charges up to 1 point less a hit, absorbs nothing under 1 mana point |
| D5 | Frozen Armor trigger | `taken.damage > 0` (points, any element: ai.cpp:2698) | record +8 (physical, 256ths) > 0 | V read / S | none | an elemental-only melee striker is frozen on d2d only; a sub-point physical hit only in game.exe |
| D6 | Energy Shield lower-level recast | replaces | refused (FUN_0056e970) | V read / S | none | none in the plan |
| D7 | calc1 <= 0 | 1 frame (`max(.., 1)`) | no freeze | V read / S | none | unreachable (calc1 >= 30) |

Not deviations (equal by read, I): the durations (d2d `start_state` uses auralen x 40 ms), same-skill refresh, Chilling's return bolt, Shiver's element path (whose damage / chill rules SKMISS and #155 cover).

## 7. The bot (our repo, main at 2f34d65)

| id | finding | where | tag | impact |
|---|---|---|---|---|
| B1 | `Fighter.Buff` is called only from `sweepPass` when `needBuff` is set, and `needBuff` is set only at a new game (`forgetGame`), after a town trip, at a goal reset and after death. The FIXRECAST limit (114 s at L1) is only looked at then: inside one clear longer than 120 s the armor lapses and stays off until the next edge | internal/bot/goal.go:1050-1054 (and :139, :465, :477; loop.go:174; death.go:404); internal/combat/fighter.go:585 | V read (our code); the lapse I | Frozen Armor (and its freeze, Q2) missing in long Act 1 clears; F2 was fixed in `buffDue` but `buffDue` is not polled. Candidate: call `Buff` between fights (e.g. each `sweepPass`, or from the fight loop when no hostile is near), letting `buffDue` decide |
| B2 | `buffStates` knows Frozen Armor 10 and Bone Armor 14 only; Shiver 88, Chilling 20, Energy Shield 30 run on the clock. Energy Shield ends at 0 mana in game.exe (Q5), which the clock cannot see | internal/combat/necro.go:40 | V read + D | ES (when learned) can be gone for most of its timer; add 20 / 30 / 88 to the map |
| B3 | the d2d backend never fills `Player.States` (`internal/d2dsrc/player.go` builds the player with no states), so on d2d every armor runs on the clock and the "learned" path is game.exe-only | internal/d2dsrc/player.go:38, :424 | V read (grep) | the FIXRECAST state proof can only be done on game.exe or with the fork's `debug states` as the reader |
| B4 | each `buffs` group casts its first KNOWN skill (`Frozen Armor, Shiver Armor, Chilling Armor`): a character with points in Shiver or Chilling but also in Frozen always casts Frozen. Correct for the one-armor rule (Q4, only one can be up), a choice rather than a bug | internal/combat/fighter.go:597-619; internal/combat/classdata/profiles.json:27-33, :66-74 | V read | none on the default tables (only Frozen is bought) |
| B5 | `AuraFrames` leaves out the +250 frames a synergy point; `buffLimit` uses `sp.Level` (points + items): both err short, so recast early, never late; a recast while up costs only mana (Q4: full duration restarts) | internal/data/skills.go:291; internal/combat/necro.go:489-501 | V read + D | none (confirms FIXRECAST's assumption "recast refreshes": V read now) |
| B6 | stale comment: `Buff` says the state is not readable; `buffDue` reads it | internal/combat/fighter.go:583-584 | V read | doc only |

## 8. PR candidate

**No PR candidate.** D1 is proven on the game side (V run, 184,771 cases) and d2d's rule is plain from the source, but the existing differential (`freeze_rules.py` + `freeze_rules_test.cpp`) drives d2d's `Fight::strike` (hit function 13), not `buff_events`; there is no d2d-side run of the Frozen Armor path, so "mismatches before > 0 on the real build" is not shown. The fix shape, for whoever builds the harness in section 6: make `buff_events` call the same freeze function the freeze candidate (fix/freeze-rules) gives `strike`, with the length resisted by the striker's cold res less the wearer's pierce and the trigger on physical damage; then run `sk_armor.py gen / game` against the new test mode. D2-D4 need stateful drills first.

## 9. Drill recipe (live, a throwaway Sorceress on one hidden instance, offline, every cheat ledgered; lab binary with the FORKDBG verbs; nothing run by this card)

Verbs: `debug points`, `debug stat`, `debug spawn`, `step N`, `dump`, `monsters`, `cmd skill`; FORKDBG (docs/forkdbg.md): `debug states` / `debug states set <skill> [lvl]` / `debug states clear <skill>|all`, `debug mon set`, `debug cast`, `debug mana full`.

- **A1 Frozen Armor live recast proof (FIXRECAST recipe, merged).** `debug points 40 1`, bind it, `debug mana full`, cast with `cmd skill 40 <own x> <own y>` (or `debug cast 40 <own x> <own y>`); `debug states` must show `buffs=40:1` and `def` x 1.30 truncated; note the tick. Pause lockstep, `step 25` units to 100 virtual s, then to 114 s: the bot's `buff skill=Frozen Armor` line must appear by about 114 s, before the state ends at 120 s (3,000 frames; `step` to 121 s with the bot idle: `buffs` empty). Level 2 (`debug points 40 2`): recast at about 126 s, state ends at 132 s (3,300 frames). Because of B1 the bot only looks at the timer when `Buff` runs: trigger a `needBuff` edge after 114 s (a town trip or a goal reset) or the line will not appear, which itself confirms B1. game.exe side (state 10 in `Player.States`): the same steps on a game.exe instance with the bot's state reader (docs/fixrecast.md); expected equal durations.
- **A2 group (D2).** `debug points 40 1`, `debug points 50 1` (Shiver), `debug cast 40 ...` then `debug cast 50 ...`: game.exe: only shiverarmor (88) is on, defence +45 %; d2d today: `buffs=40:1,50:1,` and defence +75 %. Then cast 40 again: game.exe only 10.
- **A3 refresh.** Cast 40, `step 1500` (60 s), cast 40 again, `step 2999`: still up in both; `step 2`: gone at 180 s (game.exe and d2d both, equal expected).
- **A4 striker effect (D1, D5).** `debug spawn fallen1 1` next to her, `debug mon set <id> dmg 1 1 res cold 0 life 1000 maxlife 1000` (a damaging blow, no death), Frozen Armor up, `debug stat 6` high, let it swing, `monsters` each tick after the hit: game.exe state 1 (freeze) for 30 ticks, no state 11; d2d today stun + chill for 30. Variants: `res cold 50`: game.exe 15 ticks; `res cold 100`: game.exe nothing, d2d stun + chill; Cold Mastery (`debug points 65 n`) with `res cold 50`: game.exe longer by the pierce; a champion or unique (Corpsefire in the Den of Evil, or `debug spawn` of a champion if the verb supports it: U): game.exe state 11 only, never 1. Nightmare / Hell: freeze length / 2, / 4 (needs a character there: U).
- **A5 Shiver / Chilling (equal expected).** Shiver up, a Fallen swinging and missing (`debug mon set` low to-hit is not a field: use a high defence via `debug stat` if available, U): does it take cold damage on a miss (Q3 U)? Chilling up, a Fallen shaman's fire bolt: a `chillingarmorbolt` in `debug missiles` from her to the shaman.
- **E1 Energy Shield at 0 mana (D3, D4).** `debug points 58 1`, `debug cast 58 <own x> <own y>`, `debug stat 8 768` (3 points), a 50-point physical hit (`debug hit 50` or a Fallen with `dmg 50 50`): game.exe absorbs 1.5 points (20 % of 50 is 10, capped at 3 x 16 / 32), mana 0, state 30 removed; d2d absorbs 1 point, mana 1 point, `buffs=58:1` stays. `debug mana full` afterwards: game.exe no shield until recast; d2d still shielded.

## 10. Matrix cell updates (docs/research/sorc-skill-matrix.md; proposed text, not edited by this card)

| Row | Cell | New value |
|---|---|---|
| 3.3 Frozen Armor | (a) rules | do 18 FUN_005c9480: state frozenarmor (10, group 1: replaces Shiver / Chilling / Bone Armor), armor % ln12 = 30 + 5 (L-1), length auralencalc 3000 + 300 (L-1) + 250 a synergy point frames, a recast restarts the full length [V read + D, sk-armor-trace Q1 / Q4] |
| 3.3 Frozen Armor | (a) hit effect | event 2 FUN_005cac50: a damaging (physical > 0) melee striker gets a hit with only the freeze length calc1 = (30 + 3 (L-1)) x (1 + 5 % a synergy point) frames through FUN_0057c6c0 / FUN_0057b230: resisted by cold res less pierce, plain monsters frozen at coldeffect < 0 for length / MonsterFreezeDivisor and not chilled, specials (champion, unique, superunique, boss, hireling) chilled only, immunes nothing [V run: sk_armor.py, 184,771 cases = freeze oracle] |
| 3.3 Frozen Armor | (b) d2d status | `buff_events` stuns and chills every damaging striker for raw calc1 frames, no resistance, divisor or boss rule: deviation D1 (sk-armor-trace; no d2d harness yet, no PR candidate); armor % stacks with Shiver / Chilling (D2) [S] |
| 3.3 Frozen Armor | (c) bot use | recast by `buffDue` / `buffLimit` (FIXRECAST: state 10 or duration - 6 s) but only when `Buff` runs on a `needBuff` edge: lapses inside long clears (B1) [V read] |
| 3.11 Energy Shield | (a) rules | do 23, state energyshield 30 (no group), length 3600 + 1500 (L-1); event 24 FUN_005ca6c0: per field (phys, fire, light, cold, magic; three more from monsters only; never poison) min(field x calc1 %, mana x 16 / calc2), mana -= absorbed x calc2 / 16 in 256ths; the state ends at 0 mana; a lower-level recast is refused [V read] |
| 3.11 Energy Shield | (b) d2d status | total damage in whole points, state kept at 0 mana: D3 / D4 [S] |
| Shiver Armor / Chilling Armor (inventory rows) | rules | Shiver event 3 FUN_005cad40 (its cold through FUN_0057c6c0), Chilling event 1 FUN_005cab40 (chillingarmorbolt to a ReturnFire missile's owner); states 88 / 20, group 1; lengths 3000 + 300 (L-1) / 3600 + 150 (L-1) + 250 a synergy point [V read + D] |
| Cold Mastery | touches | its cold pierce lengthens Frozen Armor's freeze on resistant strikers (not immunes) [V read] |
| 4 coverage summary | Frozen Armor, Energy Shield | move from "no V on their own behaviour" to V (read / run) |
| 7 open questions | "Frozen Armor's freeze versus chill" | answered: sk-armor-trace Q2 |

## 11. Not traced (U)

- How stat 171 (skill_armor_percent) enters the defence in game.exe (d2d: of the total, truncated: S).
- Where event 3 (attackedinmelee) and event 11 (absorbdamage) are fired, so whether Shiver hits on a miss and where Energy Shield sits against resistances and Bone / Cyclone absorb.
- The names of record fields +0x20, +0x38..+0x40 (I).
- Live: everything (section 9).

## Dead ends

- The skmana read worktree has no `freeze_rules.py` (it lives in this repo and on the freeze candidate branch); the oracle here imports this repo's copy, run from a scratch directory with the upstream `tools/emu` on the path.
- sk_armor.py's `classify` buckets by the first rule that applies (special before immune), so its counts describe the case mix, not a game property.

## Shared-file lines for the merger

docs/decomp-queue.md, row 15, status cell:

```
done 2026-11-01 (SKARMOR): Frozen Armor event 2 = the freeze rule (FUN_005cac50 -> FUN_0057c6c0 -> FUN_0057b230; V run, 184,771 cases equal to freeze_rules.py): resisted, plain monsters frozen at coldeffect < 0 / MonsterFreezeDivisor and not chilled, specials chilled only, immunes nothing; group-1 armors replace each other, recast restarts the full length; Energy Shield per field in 256ths, ends at 0 mana; d2d D1 (buff_events stun + chill raw, all monsters) D2 (armors stack) D3 / D4 (ES) D5-D7; bot B1 (Buff only on needBuff edges: armor lapses in long clears) B2-B6; no PR candidate (no d2d harness for buff_events); docs/research/sk-armor-trace.md
```

docs/BACKLOG.md (fix queue, bot):

```
- SKARMOR B1: Fighter.Buff runs only on needBuff edges (internal/bot/goal.go:1050); Frozen Armor lapses after 120 s inside one clear. Poll Buff between fights and let buffDue decide. B2: add Shiver 88, Chilling 20, Energy Shield 30 to buffStates (internal/combat/necro.go:40). docs/research/sk-armor-trace.md section 7.
- SKARMOR D1 harness: d2d test mode for Fight::buff_events vs tools/emu/sk_armor.py (then a freeze-rule PR candidate for Frozen Armor). docs/research/sk-armor-trace.md sections 6 and 8.
```

docs/INDEX.md (research list):

```
- [research/sk-armor-trace.md](research/sk-armor-trace.md): Frozen / Shiver / Chilling Armor and Energy Shield, game.exe vs d2d vs the bot (decomp row 15, SKARMOR)
```

docs/research/README.md:

```
- sk-armor-trace.md: armors and Energy Shield (SKARMOR, decomp row 15); oracle tools/emu/sk_armor.py
```

CHANGELOG.md (unreleased, docs):

```
- docs: SKARMOR trace of the Sorceress armors and Energy Shield (docs/research/sk-armor-trace.md, tools/emu/sk_armor.py): Frozen Armor's striker effect follows the freeze rule in game.exe (V run), d2d stuns and chills every striker (D1); bot finding B1 (buffs only recast on needBuff edges).
```

docs/OWNERS (changecheck warns until these are in; the merger drops them at archive time as for skmana):

```
docs/research/sk-armor-trace.md  skarmor  skarmor
tools/emu/sk_armor.py  skarmor  skarmor
openspec/changes/skarmor/**  skarmor  skarmor
```

Glossary: no new term needed ("aura event", "state group" are used in their Skills.txt / States.txt column sense).
