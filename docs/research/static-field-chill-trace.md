# Static Field and cold length / chill: game.exe 1.14d vs d2d vs our bot (agent DECOMPSTATIC, 2026-10-31)

> status: research, current (card decompstatic, decomp-queue item 6). Offline only: no game, no d2d slot; nothing pushed, posted or filed. Follows docs/research/skill-damage-trace.md ("Not covered"); the chill's speed effect is in docs/research/walk-speed-trace.md (not redone). d2d read at upstream main 72132b99 (worktree `static` under the PR checkout dir on the worker host). Tags: **V** = run in game.exe's own code under unicorn (oracle named) or read with `tools/emu/disasm.py` (said which); **D** = patched 1.14d tables via mpq-cat (docs/process/mpq-data-reading.md); **S** = source read (d2d or our Go code, file:line); **I** = inferred; **U** = unknown.

## Answer in three lines

1. **Static Field** = `MulDiv(life >> 8, 25, 100)` whole points of the target's CURRENT life, at most life - 1 (it never kills on its own), nothing at or under StaticFieldMin % of max life (0 / 33 / 50, expansion games only), a negative lightning resistance cancelled (MulDiv(dmg, 100, 100 - res)), then the normal lightning path in 256ths with pierce; it hits monsters AND hostile players in aurarange `5 + (lvl - 1)` subtiles (squared integer distance), alive, attackable, out of town, in line of sight; no boss / champion / minion rule. d2d deviates (whole points from the ceiling of life, no floor, negative resistance amplifies, no pierce, no line of sight): 309,664 oracle cases, 115,103 differ before, 0 after the local candidate `fix/static-field` (V).
2. **Cold length**: the skill's ELen (FUN_00644f20, already proven) goes into the record (+0x30), is zeroed by stat 0x99 (cannot be frozen), halved by 0x76 (half freeze duration), reduced by cold resistance less cold pierce like damage, and handed to the chill **whatever the damage came to**; the chill divides by MonsterColdDivisor for monsters with a negative coldeffect (1 / 2 / 4), skips coldeffect 0, lasts at least 1 tick, only ever lengthens. d2d chills only when the resisted damage is at least one whole point and ignores 0x99 / 0x76: 28,739 of 101,170 oracle cases differ (12,804 of 81,559 with neither stat) (V). No d2d patch yet.
3. **Bot**: no damage model for either. Static Field is an opener on elites/bosses only, `StaticCasts` 4 per monster, a life floor 25 / 33 / 50 %; game.exe's floor in Normal is 0, so the bot's 25 % is tuning, not a rule (S). Chill and freeze are not modelled anywhere (S). Drills: see DRILL RECIPE.

## Q1 Static Field in game.exe

| Address | Rule | Tag |
|---|---|---|
| `FUN_005c9800` (do 20) | skill row 42; DifficultyLevels row by game +0x6d (`FUN_00611d30`, null: no cast); params: +0 = calc(+0x13c calc2 = par3 = 0), +4 = calc(+0x138 calc1 = par4 = 25), +8 = **game +0x70 (expansion) ? row +0x40 StaticFieldMin : 0**, +0xc ELen (`FUN_00644f20` flag 1; 0), +0x10 EType (+0x1dc, ltng); radius = calc(+0x64 aurarangecalc `ln12` = 5 + (lvl-1) x 1), filter = +0x50 aurafilter 34691 (0x8783); `FUN_0056b7e0(game, caster, 0, 0, radius, filter, FUN_005c96a0, params)` | V read; field names V (loader descriptors 0x616fc5 calc1 +0x138, 0x616feb calc2 +0x13c, 0x6155a6 aurarangecalc +0x64, 0x615506 aurafilter +0x50, 0x61805b StaticFieldMin +0x40); values D |
| `FUN_0056b7e0` | centre = the caster (x / y 0); for every unit in the rooms near it except the caster: `dx^2 + dy^2 <= r^2` (`FUN_006492a0`, integer subtiles), then the filter `FUN_0056b3e0`, then the callback | V read |
| `FUN_0056b3e0` filter 0x8783 | bit 1 players (not mode 0 death / 17 dead), bit 2 monsters (not mode 0 / 12); 0x80 unit flag +0xc4 bit 2, 0x400 bit 3 (attackable, valid target: I from names); 0x100 not in a town room (`FUN_0061ab00`); 0x8000 `FUN_00554200` (hostile to the caster); 0x200 line of sight (`FUN_0064e260` mask 4) | V read; flag names I |
| `FUN_005c96a0` (per unit) | `life = stat6 >> 8`, < 1: nothing; floor: `life <= MulDiv(maxlife >> 8, floor, 100)`: nothing (returns 0, no record, no get-hit); `dmg = MulDiv(life, calc1, 100)`, `life - dmg < 1`: `dmg = life - 1`; `dmg <<= 8`; `dmg = max(dmg, calc2)`; record (`FUN_0056c8e0`: lightning +0x1c, resist stat 0x29); stat 0x29 < 0: `dmg = MulDiv(dmg, 100, 100 - res)`; record flags +0x60 or 0xd, +0x64 = 1, +4 = 0x4001; `FUN_0057c6c0(.., 1, rec)`, `FUN_0057cee0` | V (oracle `tools/emu/static_field.py`, 309,664 cases) |
| `FUN_0057c6c0` -> `FUN_0057c1e0` | the record through the resistance less the attacker's pierce (0x14e), >= -100, >= 100 nothing (`FUN_0057be00` / `FUN_0057bf80`), life in 256ths, < 0x100 after = dead (PR #135) | V (same oracle, native) |

Consequences (D / I from the rule): with calc1 25 the `life - 1` cap never binds (it would at calc1 >= 100 or a big calc2: the oracle covers both); life under 1 point or 1..1.99 points: 0 damage; res -100 still 25 %, res -150 gives 20 % (the cancel divides by 250, the resistance is capped at -100); pierce on a negative resistance can push it above 25 %. Bosses, champions, minions: no special case in these functions; the floor and the resistance are the only limits (I: no other caller path read). Players: hit when hostile (PvP), same formula (the PvP damage divisor, if any, is not in these functions: U).

## Q1 d2d (S, upstream 72132b99) and the deviations

`components/game/fight/strike.cpp` `kStaticField` in `Fight::spot`: `resisted(hit_points * pct / 100, res[3])`, `land()`; `within()` = `hypot * 5 <= aurarange` over targetable monsters.

| id | d2d before | game.exe | tag |
|---|---|---|---|
| S1 | life = `hit_points` (whole, the ceiling: life_frac ignored) | `stat6 >> 8` (the floor) | V |
| S2 | whole points through `resisted`, frac dropped | 256ths through FUN_0057bf80 | V |
| S3 | no floor | StaticFieldMin 0 / 33 / 50 % (expansion) | V |
| S4 | negative resistance multiplies (res -50: 37.5 %) | cancelled (25 %) | V |
| S5 | no pierce | attacker's lightning pierce | V |
| S6 | no `life - 1` cap, no calc2 minimum | both (no effect with 1.14d data) | V |
| S7 | monsters only, no line of sight, float distance | players too, line of sight (0x200), integer squared distance | V read / S; not in the candidate (geometry) |

Differential (`static_field.py` vs `test_game --static-field`, both in the candidate): boundary grid floors {0, 33, 50} x 13 resistances x max life 1..6 x every life in 256ths (209,664), plus 100,000 random (max life 1..3000, res -200..150, pierce 0..60, floor, calc1 0..150, calc2 0..0x2000). Before (line-for-line copy of the old code): 115,103 differ; after: 0 (md5 5583e13b of game.txt). Slices of the before: Normal-like (no floor, no pierce, calc1 25) res 0..99: 10,607 of 34,380; res < 0: 9,747 of 29,055; immune: 0 of 17,526; floor skips: 25,490 of 113,553.

## Q2 cold length, chill and freeze

| Address | Rule | Tag |
|---|---|---|
| `FUN_00644f20` | length = ELen + brackets + ELenSymPerCalc share (Ice Bolt 150 + 35 / lvl, Frost Nova 200 + 25 / lvl, Ice Blast 75 + 5 / lvl; Glacial Spike no ELen: its freeze is elsewhere, U) | V (skill-damage-trace, 0 of 116,000); D |
| `FUN_0056c8e0` | cold: +0x24 += dmg, +0x30 = length, +0x60 = 0x30, resist stat 0x2b; etype 12 (freeze): +0x24, +0x34 freeze length | V read |
| `FUN_0057c140` | target stat 0x99: cold and freeze lengths 0; else stat 0x76: both halved (`sar 1`) | V (oracle `tools/emu/cold_length.py`) |
| `FUN_0057c1e0` table 0x732980 | entry 5 = field 0x30 with res 0x2b and pierce 0x14f, entry 6 = 0x34: the lengths go through FUN_0057bf80 like damage (res >= 100: 0; -100 floor) | V (oracle) |
| `0057c8a0..0057c8a8` | `FUN_0057af80(attacker, rec +0x30)` after the life subtraction, no test on the damage | V (oracle) |
| `FUN_0057af80` chill | length <= 0: nothing; monster: coldeffect = MonStats byte +0x168 + difficulty (signed), 0: never chilled; negative: length / MonsterColdDivisor (DifficultyLevels +0x18: 1 / 2 / 4, D); a player: -50, no divisor; length >= 1; no chill state 0xb: state on, stat list until frame + length, stats 0x43 / 0x44 / 0x45 += coldeffect, `FUN_00623f50` (speed: walk-speed-trace); running: extended only if the new end is later; then a 20 % roll sets state 0x6b (I: the visual), monsters only | V read |
| `FUN_0057b230` freeze (+0x34) | a player: chill instead (I: player freeze is elsewhere, U); a monster with state 0x36, nothing; MonStats boss (`FUN_0063e9f0`), a special one (`FUN_005a0180` flag 8: champion / unique), a hireling: chill instead; else length / MonsterFreezeDivisor (+0x14), state 1 until frame + length, extended only | V read |

d2d (S): `missile_blow` (components/rules/combat.hpp) sets `chill_ticks = resisted(elen, res)` only `if (resisted_damage > 0)`; `land()` (strike.cpp 'Chilled' block) has the coldeffect / divisor / >= 1 / extend rules (match by reading) and a ponytail for the 20 % state.

| id | d2d | game.exe | tag |
|---|---|---|---|
| C1 | no chill when the resisted cold damage is under one point | chills whenever the resisted length > 0 | V: 12,804 of 81,559 cases (stats 0) |
| C2 | stats 0x99 / 0x76 not read | zero / halve the length | V: the rest of 28,739 of 101,170 |
| C3 | (freeze path not compared) | bosses, champions, uniques, hirelings get chill instead of freeze | V read; d2d not traced (U) |

For the bot's early skills C1 needs a monster with high cold resistance (Ice Bolt 3..5 points at res 75 %); C2 needs the stats (players' items, rare monster mods): low impact (I).

## Q3 the bot's model (S, file:line at this worktree)

- Static Field: opener in `internal/combat/fight.go:679-726` (`staticOpener`), only `tg.Elite || tg.Boss`, up to `StaticCasts` 4 per monster (`internal/env/tuning_combat.go:77`, knob `staticCasts`), Blood Raven 0 (`:44`, U), life floor 25 / 33 / 50 % (`tuning_combat.go:77`, `fighter.go:737`), range `(lvl+4)*2/3` yards (`internal/combat/skills.go:48`, "Kolbot", UNVERIFIED). No damage model; the sim pins only the press count (`sim_test.go:503-518`).
- Frost Nova: crowd cast at 3+ hostiles within 4 subtiles, mana floor 25 (`internal/combat/castfirst.go:20-122`); Ice Bolt / Glacial Spike: plain attack skills (classdata). No chill, freeze or slow modelled; no kiting keyed on chill.

| id | finding (bot side, for the fix queue) | where | tag |
|---|---|---|---|
| B1 | Static Field range `(lvl+4)*2/3`: game.exe's radius is `5 + (lvl-1)` = `lvl + 4` subtiles, which is `(lvl+4) x 2/3` yards (1 yard = 1.5 subtiles, I); right only if the bot compares it in yards, 1.5x too small if in subtiles: check the unit at the call site | skills.go:48, fight.go:722 | S + V rule; unit I |
| B2 | Normal floor 25 % is tuning; game.exe has none in Normal: after 4 casts a 100 % monster is at 31.6 % (0.75^4), so the floor does not bind for 4 casts; a 5th+ cast would still work in game.exe | tuning_combat.go:77 | S + V rule |
| B3 | the life read the floor uses must be the CURRENT life %: game.exe compares whole points with MulDiv; within 1 point of the floor the bot may cast once for nothing (a cast that makes no record) | fight.go:722 | I |
| B4 | no chill model: kite/back-off timings assume unchilled monsters; chilled monsters move at coldeffect % (walk-speed-trace) for ELen / divisor ticks (Ice Bolt lvl 1: 6 s in Normal, monsters with coldeffect < 0 3 s in Nightmare) | combat (none) | S |
| B5 | Static Field in Nightmare / Hell on d2d hits below 33 / 50 % (S3): bot is unaffected (its floor stops first) | - | I |

## DRILL RECIPE (what a live check should see)

Observables are memory reads (life stat 6 in 256ths, state 0xb, the monster's mode), one Static Field cast at a time, on a throwaway character, Normal unless said:
1. **25 % of current life**: a full-life monster with lightning res 0 and life L points (L >= 4): after one cast life = (L - floor(L/4)) x 256 exactly; four casts: each takes 25 % of what is left (whole points, floor). On d2d after the candidate the same; before, a monster whose life has a fraction loses one point more.
2. **Never kills**: cast on a monster at 1..3 points: no change (damage 0); at 4..7: it loses 1. Never 0 life.
3. **Negative resistance**: a monster cursed below 0 lightning res (or an Act 1 monster with negative res, D: monstats ResLi) loses the same 25 %; d2d before: more.
4. **Floor (Nightmare / Hell)**: at or under 33 % / 50 % of max life no damage and no get-hit (the cast makes no record).
5. **Radius / walls**: a monster at 5 + (lvl - 1) subtiles is hit, one subtile more is not; behind a wall it is not hit (d2d hits it: S7).
6. **Chill**: Ice Bolt on a coldeffect -50 monster in Normal: state 0xb for ELen ticks (150 + 35 (lvl - 1)); in Nightmare half; Frost Nova 200 + 25 (lvl - 1); a monster with coldeffect 0 never gets 0xb; a cold-resistant monster that takes < 1 point still gets 0xb in game.exe (C1).
Private fork debug commands by name only: the control verbs `stat` and `skill` (present in the fork's control code, V grep; their exact arguments not checked here), test cheats only through `internal/cheat` / `cmd/d2cheat` (life / level setup, logged); a verb that dumps a monster's life in 256ths and its states was not found (U): read them from memory instead.

## PR candidate (local only)

Branch `fix/static-field` on upstream 72132b99 in the worker host's `static` worktree: d9d64d53 claim (rules `static_field_fine`, DifficultyLevels `static_field_min`, `kStaticField` in strike.cpp, ponytail for S7), 6270d162 tooling (`tools/emu/static_field.py`, `test_game --static-field`). Upstream test_game / test_rules / test_combat rc 0, pr_guard 0 FAIL 0 WARN, shadowing sweep clean. Proof / draft / pack: docs/research/pr-static-{proof,draft,evidence-pack}.md. Not pushed, no verifier yet. C1 / C2 (cold) proven by the oracle only (Python before side); no patch: next candidate.
