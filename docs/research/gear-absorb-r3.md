# Gear absorb r3: the heal-back, a life oracle, players only (game.exe 1.14d vs d2d vs the bot)

> status: current (2026-11-17, card gearabsorb3, TASK-621, agent d2-general on Sonnet). Offline: no game, no d2d slot, no Go changed, nothing posted upstream. The game.exe side ran under the emulator and the d2d side as fork builds, both on the builder container. Follows [decomp37-combat.md](decomp37-combat.md) (T1) and the verifier's finding that r2 FAILS ([../upstream/verifier-decomp37-2026-11-17.md](../upstream/verifier-decomp37-2026-11-17.md), A1 to A6). Consumer: the independent re-verification, then the posting agent ([../upstream/pr-queue.md](../upstream/pr-queue.md)). Draft body: [../upstream/bodies/gear-absorb-body.md](../upstream/bodies/gear-absorb-body.md).

Tags: **V run** = game.exe's own code under the emulator (`docs/emu-image.md`), compared with a d2d build of the real code; **V read** = disassembly read (addresses given); **S** = an earlier doc; **I** inferred; **U** unverified.

## Answer

r2 modelled half of game.exe's absorb: it lowered the hit. game.exe also gives the absorbed amount back as life. r3 adds that, measures the player's life (not an internal field) and keeps mercenaries out. Branch `fid/gear-absorb-r3` on the private fork remote (never posted): oracle commit `982a65c1`, claim `270b9799` on upstream `f1533770` (the heads are final at the card result).

## Q1..Q6 (game.exe)

| Q | Answer | Tag |
|---|---|---|
| Q1 Where is the absorbed amount kept? | `FUN_0057bf10` adds what it takes (percent, then flat) to the record's `+0x48` (`0x57bf48`, `0x57bf6c`); no other code in `0x579000..0x57e000` writes `+0x48` (the verifier's scan). It is signed: a field the flat reduction left under 0 gives a negative share, and the percent of a negative field is a negative `MulDiv`. | V read |
| Q2 Who reads it, and in what order against the damage? | `FUN_0057c6c0` calls `FUN_0057c1e0` (fills `+0x48`, `+0x4c`), then at `0x57c844` loads `[edi+0x48]` and calls `FUN_0057a980` (heal), and only then (`0x57c852..0x57c874`) takes the record total `+0x4c` off stat 6. So the heal comes FIRST: a blow that would kill is survived when the heal plus the life covers it. | V read, V run (case 3 of the pins: a heal that does not save; id 9: the cap) |
| Q3 What does `FUN_0057a980` do? | total at or under 0: nothing; the unit dead (`FUN_005541b0`) or under state `0x5c` (death_delay): nothing; else stat 6 = `min(stat6 + total, FUN_00625d10)` (max life; a life already above max is brought down to it). Returns the amount applied. | V read (`0x57a980..0x57a9f0`), V run |
| Q4 Min / clamp rules on the subtraction | `+0x4c` above 0: life = life - total; a result under `0x100` (under one point) is set to 0 (dead). A total at or under 0 subtracts nothing and does not kill. | V read (`0x57c852..0x57c874`), V run |
| Q5 Same for all four elements and both kinds? | Yes: one record total `+0x48`, the sum over fire, lightning, cold and magic of (field before the absorb minus after), percent and flat alike. A negative field in the same record lowers the sum. | V read, V run (id 8: three fields; id 7: negative) |
| Q6 Unit types | `FUN_0057c6c0` tests the unit type only to skip a monster flagged by `FUN_00457490` (`0x57c76a..0x57c775`) and a dead unit; `FUN_0057c1e0`, `FUN_0057bf80`, `FUN_0057bf10` and `FUN_0057a980` read the stat list of whatever unit they get. So a hireling (a monster unit with the merc's item stats on its list) would absorb and heal too (I: not read for the hireling's item-stat transfer). Monsters carry no absorb stats from MonStats (U: a MonProp that sets one was not looked for). d2d's merc keeps its life in whole points with a fractional carry, so r3 does not model it: the merc takes none, as on the parent. | V read for the functions, I for the merc, U for a monster |

Mana Burn and leech do not touch `+0x48`: Mana Burn is its own record field (FUN_0057aa60 path), leech heals the attacker after the hit (decomp35).

## Deviations (d2d against game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | The absorb of gear stats 142..149 after the resistance (percent capped at 40, then flat). d2d parent: none. r2 and r3: as game.exe. | V run | yes: `gear_absorb.py`, counts in the body |
| D2 | The heal-back of the absorbed amount, per blow, before the total. d2d parent and r2: none. r3: `absorb_heal` in `Fight::hurt_side`. | V run | yes, the same oracle (life) |
| D3 | A merc's gear absorb and heal. game.exe: acts (I). d2d: none, before and after. | I | buildable: the merc's life is whole points with `life_frac` (tick.cpp `hurt_side`); the oracle is the same record on a type-1 unit whose stat list carries the merc's gear |
| D4 | State `0x5c` (death_delay) on the heal. game.exe: no heal. d2d: no such state on the player is modelled; the dead check stands in. | U | would need a unit under that state |
| D5 | Absorb on the ignore-resistances record (`0x57c01e`; flags `0x100`/`0x200`/`0x400` by attacker kind), the burn and blast paths, a missile's own element roll, a magic field's magic resistance (decomp37 T3). | U | not built |

## Bot findings (described, not changed)

| B | Finding | file:line | Tag |
|---|---|---|---|
| B1 | The survival model folds the percent absorb into the resist as a lower intake (`rt.Intake(p.Resist.Fire, p.Absorb.FirePct)`) and has no flat-absorb or heal term. In game.exe an absorb both lowers the hit AND refills that much life up to max: below full life a blow of F points with absorbed amount A costs F - 2A (not F - A), and a hit that would kill at life L survives when L + A covers F - A. At full life the heal is capped away (same as lowering the hit). The power estimate is conservative below full life and exact at full life for the percent part | `internal/power/hero.go:118..123` | I (arithmetic from V run rule) |
| B2 | `internal/d2/player.go:47..50` documents absorb as "applied to an elemental hit" only | `internal/d2/player.go:47` | I |

## Evidence

| What | Result | Tag |
|---|---|---|
| Oracle, seed 7: `gear_absorb.py gen 20000 7`, `game` (FUN_0057c6c0 on a player, life after one to three blows, 7,923 with a Cyclone pool) against `test_gear_absorb --life` | parent (oracle commit alone) **9,639 differ**, exit 1; claim **0**, exit 0 | V run |
| Oracle, seed 5001 (fresh) | parent **9,519 differ**, exit 1; claim **0**, exit 0 | V run |
| Ten pinned lives (heal capped, Cyclone pool, below max, a heal that does not save, two blows per frame, flat, order with a resistance, a negative field, three fields, percent cap with a capped heal) | equal game.exe's, through the rules alone (`test_gear_absorb_rules`, no game data) and through `Fight::hurt_side` (`test_gear_absorb`) | V run (game side), the same lines in both tests |
| The merc | `merc_fighter()` carries no absorb even when its stat sum does (a potion state's stats join the sum like gear: the resistance given that way reaches its `Fighter`, the absorb does not) | V run (d2d) |
| `compare` exit codes | 1 on a difference, an empty file, unequal lengths, ids out of step, non-numeric values, an unreadable file | V run |
| Mutants | 12 of 12 abort (no heal, heal after the total, merc keeping absorb, flat before percent, percent and flat swapped, cap 41, magic and cold swapped, heal uncapped, negative heals, total before heal, recompute without absorb, `Foe::take` without absorb); detail in the body | V run |
| Build and ctest | 1,183 targets built; `ctest -j2` 358 of 359, `smoke_d2d` failed once under load and passes alone (359 of 359); 13 skips for data | V run |
| `pr_guard.py upstream/main HEAD`, `presubmit.py --quick --base upstream/main` | 0 FAIL 0 WARN; 0 FAIL, WARN siblings / build / gcc (answered in the body) | V run |


## What changed against r2 (the verifier's A1 to A6)

| A | r3 |
|---|---|
| A1 heal | `rules::absorb_heal`, per blow, before the totals come off, capped at max life (the per-blow choice: game.exe applies one record at a time, so two blows in a frame are two capped heals; pinned by case 4, 30 not 20) |
| A2 oracle on the life | `gear_absorb.py game` runs FUN_0057c6c0 for real and prints the life; the d2d side is `test_gear_absorb --life` through `Fight::hurt_side` |
| A3 merc | decided from the decompilation: game.exe's functions test no unit type (I for the hireling's item stats), d2d's merc has whole-point life and no heal path, so r3 keeps the merc OUT on purpose (`merc_fighter` zeroes the absorb, pinned by the potion-state check) and the body says so; r2's "default (zero) absorb" wording is gone |
| A4 recompute | case 1 (Cyclone pool) pins `absorb_blows`, and `Foe::take` carrying the absorb: mutants drop either and the rules test aborts. Energy Shield's recompute (`energy_shield_blows`) carries the same call but NO test runs it with an absorb (the oracle has Cyclone Armor only): a mutant dropping the absorb there passes (U, open: an Energy Shield oracle case) |
| A5 limits | the ignore-resistances record, the magic resistance (decomp37 T3), death_delay are in the body's Limits |
| A6 template | Claims table, re-check block, siblings answer in the body |

Reseed: neither test draws (every blow is `exact` and `sure`, the pool and life are inputs), so the 200-seed luck check has nothing to reseed; the oracle ran on a second fresh seed (5001) instead.

## Drill recipe (lab only, nothing run)

Fork verbs of [../forkdbg.md](../forkdbg.md) and [../forkcheats.md](../forkcheats.md), lab binary of [../d2d-fork.md](../d2d-fork.md), a hidden slot, a throwaway character (setup writes declared and logged; a character with written stats shows the mechanism only).
- **R1 (D2):** `debug stat 148 20` (20 percent cold absorb), set life to 90 of a 100-life character with the life-setting verb, `debug spawn` a monster with a cold El row at a damage of 100 after resistance, let one hit land, read the life: game.exe 20, the parent dead, r2 10.
- **R2 (cap):** the same at full life: the life after is max minus the post-absorb hit on both game.exe and r2 (the heal is capped).
- **R3 (D3):** a merc with a gear piece carrying an absorb property: read the merc's life after a hit; game.exe (I) heals it.

## Matrix and queue updates (proposed text)

- `docs/upstream/pr-queue.md`: gear absorb row: r2 FAIL (verifier 2026-11-17) replaced by `fid/gear-absorb-r3`, oracle and claim commits as in the body; verifier-ready.
- `docs/research/decomp37-combat.md` D6 line: the merc is not "no gear stats": `merc_fighter` is `make_fighter` over its gear; r3 zeroes its absorb on purpose.

## Shared-file lines for the merger

See `backlog/merge-lines/agent-gearabsorb3.txt`.
