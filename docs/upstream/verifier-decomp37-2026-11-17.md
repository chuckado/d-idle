# Verifier: decomp37, gear absorb and Open Wounds (2026-11-17)

> status: current, 2026-11-17 (card verifydecomp37, TASK-613, Opus, singly; independent verifier before posting). Nothing posted. Subjects: private fork branches `fid/gear-absorb-r2` (oracle `00e7b018`, claim `82eced5f`) and `fid/open-wounds-r2` (oracle `857ece45`, claim `6d4b506d`), both on upstream `ed007569`; bodies `docs/upstream/bodies/decomp37-gear-absorb-body.md` and `decomp37-open-wounds-body.md`; trace `docs/research/decomp37-combat.md` (TASK-607). Checklist from `verifier-decomp35-2026-11-16.md`, `verifier-decomp36-2026-11-16.md` and `verifier-decomp38-2026-11-17.md` (branch `agent/verifydecomp38`). The fixed `fid/open-wounds-r3` is on the private fork remote only.

Tags: **V run** (run by this verifier today, NAS builder), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdicts

| PR | Verdict | Post |
|---|---|---|
| Gear absorb | **r2 FAIL (incomplete mechanic: the heal)** | nothing yet. The reduction rule is game.exe's and r2 implements it exactly, but game.exe also gives the absorbed amount back as life (`FUN_0057a980`); r2 does not model it and the body denies it. Needs an r3 by a worker (A1 to A6 below) |
| Open Wounds | **r2 FAIL (tests), r3 PASS WITH CHANGES** | `fid/open-wounds-r3`: oracle `@@OWO@@`, claim `@@OWC@@`, on upstream `ed007569` (rebases clean on upstream main `84c86b40`); body after W1 to W7 |

### Gear absorb: why r2 cannot be posted

1. **The absorbed amount heals the defender (blocker).** `FUN_0057bf10` adds what it takes to the record's `+0x48` (`0x57bf48`, `0x57bf6c`; nothing else in `0x579000..0x57e000` writes `+0x48`, V read, a scan). `FUN_0057c6c0` hands `+0x48` to **`FUN_0057a980`** (`0x57c844..0x57c84a`) BEFORE it takes the record total `+0x4c` off stat 6 (`0x57c852..0x57c874`). `FUN_0057a980` (V read, `0x57a980..0x57a9f0`): absorbed at or under 0, the unit dead (`FUN_005541b0`) or under state `0x5c` (`death_delay`, States.txt row 92): nothing; else stat 6 = `min(life + absorbed, FUN_00625d10(unit))` (max life). So in game.exe an absorb both lowers the hit and refills that much life, up to max. r2 only lowers the hit. A player at full life sees the same in both (the heal is capped); below full life, with 20 % cold absorb on a 100-point cold hit (after the resistance), the player loses 100 on the parent, 80 on r2 and 60 in game.exe. The body's Limits line "The absorbed amount game.exe adds to the record (`+0x48`) is not read by anything d2d models" hides this, and "In play" ("takes 20 percent less") is wrong for game.exe.
2. **The oracle measures the field, not the life (high).** `gear_absorb.py` prints the element field and `+0x4c`; `+0x48` and the life are never compared, which is how the heal went unseen (the decomp38 lesson).
3. **The merc is changed too, and the body says it is not (high, claims).** `Fight::merc_foe` builds the merc's `Foe` from `merc_fighter()` (`tick.cpp:286..290`), which is `make_fighter` over the merc's gear (`merc.cpp:57..105`), so from r2 on a merc's gear absorb lowers a monster's hit on it. That is probably right (`FUN_0057bf80` / `FUN_0057bf10` test no unit type; I), but the body's WARN answer ("the merc and pets ... default (zero) absorb, so nothing changes for them") and `combat.md` ("Not modelled: absorb on a ... mercenary defender") are false. The merc would also heal in game.exe (`FUN_0057a980` takes any unit).
4. **No test reaches the armors' recompute or `Foe::take` (medium).** Mutants that drop the absorb from `absorb_blows` / `energy_shield_blows` (a6) or from `Foe::take`'s parts (a7) pass `test_combat` and the oracle (section Mutants). Energy Shield and Cyclone Armor take from the raw field before `FUN_0057c1e0` (S, the `combat.hpp` comment), so the recompute is where a sorceress with Energy Shield gets her absorb.

What r2 gets right (V run; keep it in r3): the row table, the order (flat magic reduction, resistance, percent capped at 40 with `MulDiv`, then flat `<< 8` capped at what is left, signed), a negative field moved toward 0 and then cleared, magic with no resistance on this path, the stat-to-element mapping (magic 146 / 147, cold 148 / 149). The parent differs in 13,189 of 20,000 (decomp37's seed 7, game side byte-identical) and 13,242 of 20,000 (fresh seed 5001); r2 in 0. d2d reads none of the eight stats on the parent (V read: `git grep` of `components` and `apps` for the names and the ids).

### Open Wounds: why r2 cannot be posted, and r3

1. **The d2d side of the oracle is a copy of the rule, not d2d's path (high).** `open_wounds_harness.cpp` decides `special` itself from the case's data word and multiplies `open_wounds_fine` by 200; it never runs `Monster::target` (where `Target::special` is set from the `Boss` kind), `player_blow`, `Fight::land` (where `bleed_rate` is set) or `Fight::monster_dots` (where the life goes). The claim commit also rewrites the harness, so the oracle's measurement changes between the two commits (the decomp38 r1 finding). r3 keeps the emulator compare as the rule check and adds `tests/game/test_open_wounds.cpp`: a real `player_blow` on `Fight::target_of`, `Fight::land`, then 250 frames of `Fight::monster_dots`, on a fallen as a plain monster, a champion, a unique, a superunique and a minion at levels 2, 30, 61; the life lost must be `200 v / 256` with game.exe's `v` (49 / 24, 436 / 218, 1,426 / 713; V run, `open_wounds.py game`).
2. **No test pins the path (blocker).** r2's `test_combat` pins only `open_wounds_fine`. Mutants that set `special` for a minion, ignore it in `player_blow`, take the rate per second instead of per frame, shorten the wound or bring back the whole-point truncation pass `test_combat` and the oracle; the halving rounded up passes `test_combat` (all four halved pins are even). r3's tests abort every one (section Mutants), and the new test fails on the parent (section Build).
3. **The compare passes with nothing compared (medium).** A case file of player rows only gives `# total: 1 cases, 0 differ`, exit 0 (V run); a missing file is a traceback. r3: `# total: N monster cases compared, M differ`, exit 1 when none is compared, `cannot read` on a missing file.
4. **The body's seed-7 counts do not reproduce (medium).** `open_wounds.py gen 200 7` (r2's script) gives 1,982 lines (244 player rows), 1,738 compared: the parent differs in **1,729** (plain 849 of 858, special 880 of 880), not "1,748 of 2,000 (plain 859 of 867, special 889 of 889)" (those came from the scratch `dc37_ow.py`, whose case file differs). Seed 1 reproduces 16,463, but of **16,526 compared** (7,095 plain, 9,431 special), not "21,782 compared" (that includes the 5,256 player rows).
5. **The WARN answer is false (medium, claims).** "the merc's and the pets' blows carry no Open Wounds in d2d (`player_blow` is the only reader)": `merc_strike` calls `player_blow` with `merc_fighter()` (the merc's gear) and the merc's level (`deaths.cpp:409..411`), and a Shadow swings `actor->player_combat` with its own monster level (`pets.cpp:300..303`); both reach `Fight::land`. So a merc's or a Shadow's Open Wounds takes the new rule with the attacker's own level, as game.exe's handler reads the attacker's stat 12 (`0x5bfed8`).

## Game side, re-derived (V read by this verifier, then V run)

### Gear absorb

- **Row table** `0x732980` (V run, dumped from the image, 12 rows of `0x2c`: record offset, resist, max resist, pierce, absorb percent, absorb flat, reduction index, ...). Fire `0x10`: `0x8e / 0x8f`; lightning `0x1c`: `0x90 / 0x91`; cold `0x24`: `0x94 / 0x95`; magic `0x20` (resist `0x25`, max `0x26`, no pierce): `0x92 / 0x93`; physical, the two poison rows, poison length, chill, freeze and the rows `+0x38 / +0x3c / +0x40` hold -1 for both. So 142 / 143 fire, 144 / 145 lightning, 146 / 147 magic, 148 / 149 cold, as r2 maps them.
- **`FUN_0057bf80`** (`0x57bf80..0x57c057`): a field at or under 0 is stored 0 and the function stops. With the third argument set (the record ignores resistances), a positive resistance becomes 0 and the flat reduction is skipped (`0x57bfb4..0x57bfcb`); else the flat reduction `[ctx + idx x 4 + 0x1c]` comes off (`0x57bfcd..0x57bfd0`). If the field is still above 0 and the resistance not 0: `MulDiv(field, 100 - min(res, 100), 100)` (`0x57bfd4..0x57bff4`). Then, **only when the argument is 0**, `FUN_0057bf10` (`0x57c01e..0x57c02b`).
- **`FUN_0057bf10`** (`0x57bf10..0x57bf75`): `p = stat(row +0x10)`; `p >= 40` is 40, `p <= 0` skipped; `a = MulDiv(field, p, 100)` (`FUN_00483360`), `+0x48 += a`, `field -= a`. `f = stat(row +0x14) << 8`; `f <= 0` skipped; `a = f < field ? f : field` (signed), `+0x48 += a`, `field -= a`. r2's `absorbed_fine` is this with d2d's `mul_div` (the branches of `FUN_00483360`, V read `rules.hpp:1675`).
- **The ignore argument**: `FUN_0057c1e0` sets it for a monster attacker (type 1) when the record has flag `0x100`, `0x200` or `0x400` by the attacker's kind (`0x57c309..0x57c371`); no d2d monster blow models such a record (U which records carry them).
- **The heal**: item 1 above (V read; not run under the emulator).

### Open Wounds

- **Handler table** `0x7325b0` (V run, dumped): entry 15 `0x5bfe60`, entry 16 `0x5bffc0` (Crushing Blow).
- **`FUN_005bfe60`** (`0x5bfe60..0x5bffbb`): the chance is the attacker's stat through `FUN_00625500` (the event's packed stat id `>> 16`), `rand % 100 < chance` on the attacker's seed (`0x5bfe9e..0x5bfecb`). `L = max(stat 12 of the attacker, 1)` (`0x5bfed8..0x5bfee7`). **`FUN_005bfda0`** (`0x5bfda0..0x5bfe59`): `L <= 1` gives 0; to 15 `(L - 1) x 9`; to 30 `126 + (L - 15) x 18`; to 45 `396 + (L - 30) x 27`; to 60 `801 + (L - 45) x 36`; above `1,341 + (L - 60) x 45`, as r2's `open_wounds_fine`. Plus 40 (`0x5bff20`). Struck type 0 (a player): `/ 4` toward 0 and `/ 2` more for event 6; type 1 (a monster): halved toward 0 when `FUN_005a0180(unit, 12)` (`0x5bff3c..0x5bff56`), event 5 and 6 alike; any other type unchanged. To `FUN_0056e970`: 200 frames, stat `0x4a`, the negated value, state `0x3e` (`0x5bff58..0x5bffa2`).
- **V run** (`open_wounds.py game`): game.exe's values at levels 1, 2, 30, 61, 99 for words 0 / 0x10 (not halved) and 4 / 8 / 0xa (halved): 40 / 20, 49 / 24, 436 / 218, 1,426 / 713, 3,136 / 1,568.
- **I, as decomp37**: which spawner sets bit 4 or 8 (d2d's champion, unique and superunique as "special", minion not; its stun code uses the same mapping). **U, new**: an act boss (MonStats `boss`, not a placed unique) is `Boss::none` in d2d and takes the whole wound; whether game.exe's act-boss spawn sets bit 4 or 8 was not read.
- **U, as decomp37 D8**: d2d replaces the rate and restarts the 8 s on a second wound; `FUN_0056e970`'s rule was not read.

## Differentials (V run)

d2d side: each tree's harness compiled against that tree's rules headers, compared with the branch's `compare`.

| Tree | gear seed 7 (20,000) | gear seed 5001 (20,000, fresh) | Open Wounds seed 7 (1,738 compared) | seed 5001 (3,069 compared, fresh) | seed 1 (16,526 compared) |
|---|---|---|---|---|---|
| parent `ed007569` / oracle commit | **13,189**, exit 1 | **13,242**, exit 1 | **1,729**, exit 1 | **3,054**, exit 1 | **16,463**, exit 1 |
| claim (r2; r3 has the same code) | 0, exit 0 | 0, exit 0 | 0, exit 0 | 0, exit 0 | 0, exit 0 |

The gear cases without an absorb are equal on the parent (3,393 and 3,268): the reduction and resistance order is unchanged.

## Mutants (V run)

Local mutants only (never pushed). `test_combat` compiled alone against each tree (its `test-links` are the header-only rules) and run; the oracle compare as in Differentials (seed 7 / seed 5001, cases that differ). `test_open_wounds` built from each tree under the builder's lane lock (strike.cpp and ai.hpp mutants).

| Mutant | r2 `test_combat` | oracle compare | r3 tests |
|---|---|---|---|
| a1 percent cap 41 | aborts (`5120`) | 2,271 / 2,279 | n/a (gear r3 not built) |
| a2 absorb before the resistance | aborts (`7680`) | 1,826 / 1,744 | |
| a3 flat in 256ths, not `<< 8` | aborts (`7680`) | 5,007 / 4,996 | |
| a4 negative field skipped | aborts (`-164`) | 4,883 / 4,882 | |
| a5 magic and cold stats swapped in `make_fighter` | aborts (the mapping) | not reached (the harness sets the fields) | |
| a6 recompute (`absorb_blows`, `energy_shield_blows`) without absorb | **passes** | **0** (not reached) | |
| a7 `Foe::take` drops the absorb | **passes** | **0** (not reached) | |
| a8 percent rounded (`+ 50`) | aborts (`-164`) | 3,310 / 3,195 | |
| a9 flat before the percent | aborts (`5120`) | 1,556 / 1,487 | |
| a10 `MulDiv(field, 100 - p, 100)` | aborts (`-164`) | 5,482 / 5,475 | |
| o1 / q1 no halving | aborts | 880 / 1,668 | `test_combat` aborts |
| o5 / q5 halving rounded up | **passes** | 432 / 836 | `test_combat` aborts (`player_blow` pin, 24) |
| o6 breakpoint 16 | aborts | 17 / 35 | |
| o8 / q8 `player_blow` ignores `special` | **passes** | **0** (harness computes `special` itself) | `test_combat` aborts |
| q2 a minion is special | **passes** (V read: no test builds a `Monster`) | **0** | @@Q2@@ |
| q3 rate per 25 ms (per second confusion) | **passes** | **0** | @@Q3@@ |
| q4 wound 200 x 25 ms | **passes** | **0** | @@Q4@@ |
| q9 whole points a second (the parent's truncation in `Fight::land`) | **passes** | **0** | @@Q9@@ |

So r2's tests and oracles miss six of the gear and Open Wounds path mutants; r3's two tests abort every Open Wounds mutant tried.


@@BUILD@@

## RNG luck

Neither rule draws: the gear grid is deterministic (`exact` records, no roll), and the Open Wounds value has none (the chance roll is fixed at 100 on both sides and is not part of the claim). The decomp35 200-seed reseed does not apply; both oracles ran on a fresh case set instead (seed 5001), with 0 on the claims. The new end-to-end test draws only the hit roll (AR 1,000,000, up to 20 tries, `blow.hit` asserted).

## Required work for gear absorb (r3, a worker)

- **A1 (blocker).** Model the heal: the absorbed total of a blow (what `absorbed_fine` takes on the four fields, the recompute's included) goes to the defender's life before the hit's damage, capped at max life, not when dead or under `death_delay` (`FUN_0057a980`, `0x57c844..0x57c84a`). In d2d that is `Fight::hurt_side` for the player and the merc, per blow as game.exe applies one record at a time (an aggregate heal capped once differs when two blows land in one frame: say which is done). Pin it end to end on a player below max life.
- **A2 (high).** The oracle measures the life: run `FUN_0057c6c0` (or `FUN_0057c1e0`, `FUN_0057a980` and the stat-6 subtract) on a player below max life and print the life after; the d2d side through `Fight::hurt_side` (as `test_mana_burn` r2).
- **A3 (high, claims).** Body and `combat.md`: the merc takes its gear absorb (and the heal) through `merc_fighter`; drop "default (zero) absorb" and "Not modelled: ... mercenary defender". Pets have no gear (true).
- **A4 (medium).** Pin the recompute: an Energy Shield or Cyclone Armor blow with gear absorb in `test_combat` (kills a6), and `Foe::take` carrying it (a7).
- **A5 (medium, claims).** Limits: the ignore-resistances record (`0x57c01e`; flags `0x100 / 0x200 / 0x400` by attacker kind) skips the absorb, d2d models no such record; the player's magic resistance on this path is decomp37 T3 (not built), so a magic field with a non-zero magic resistance stays wrong.
- **A6 (low, template).** As W6 below; fix the decomp37 trace's line that the merc carries no gear stats (D6 there).

## Required changes to the Open Wounds body (`decomp37-open-wounds-body.md`, for r3)

- **W1 (high).** Commits and tree: "upstream `ed007569` plus `@@OWO@@` (oracle) and `@@OWC@@` (claim)". The claim also adds `tests/game/test_open_wounds.cpp` (a blow through `Fight::land` and `monster_dots` on a plain monster, a champion, a unique, a superunique and a minion; game.exe's life lost at levels 2, 30, 61) and the `player_blow` pins in `test_combat` (49, halved 24); ctest runs both. The oracle's compare exits 1 when no monster case is compared.
- **W2 (high).** Counts: seed 7 "1,738 compared (244 player rows not compared): the parent differs in 1,729 (plain 849 of 858, special 880 of 880)"; seed 1 "16,526 compared (5,256 player rows not compared), 16,463 differ"; add the fresh seed 5001 row (3,069 compared, 3,054 differ, 0 after). Say the emulator compare's d2d side is the rule (`open_wounds_fine`), and `test_open_wounds` checks the path.
- **W3 (high).** Mutants: r3's list (section Mutants) instead of the four, with the test that aborts each.
- **W4 (medium, claims).** WARN answers: replace the unit-kinds line by "A merc (`merc_strike`: `player_blow` over its gear, its level) and a Shadow (the owner's fighter, the Shadow's level) wound through the same `player_blow` and `Fight::land`, so they take the new rule with their own level, as game.exe's handler reads the attacker's stat 12; a player struck is not modelled (no monster carries the property in d2d); a co-op partner's hero is a player (the same path)." `build` and `gcc`: "built and ctest run (Release, GCC): @@CTESTLINE@@".
- **W5 (medium, claims).** Limits: add "An act boss (MonStats boss, not a placed unique) is `Boss::none` in d2d and takes the whole wound; whether game.exe's act-boss spawn sets bit 4 or 8 of the data word was not read."
- **W6 (low, template).** The template's Claims table (`Commit | Claim or change | Evidence | Re-check`, a row per commit) and Checklist, ticked: tests (`tests/test_combat.cpp`, `tests/game/test_open_wounds.cpp`), finding (`docs/research/re/combat.md`), `pr_guard` 0 FAIL 0 WARN. Re-check block: `ctest --test-dir build -R 'test_combat|test_open_wounds' --output-on-failure` (`D2_MPQ_DIR` set; `test_open_wounds` skips without the data).
- **W7 (low).** "In play": add the halved line, e.g. level 30 over 8 s: 340 points on a plain monster, 170 on a unique (game.exe; d2d after r3 the same, V run by `test_open_wounds`).

## For the coordinator (not body changes)

- `decomp37-combat.md` section 5 and the TASK-607 result cite Open Wounds counts that r2's scripts do not reproduce (W2), and D6 there says the merc carries no gear stats: it does (`merc_fighter`).
- Gear absorb r3 is a worker card (A1 to A6, Sonnet); this verifier's checklist then re-runs.

@@HYGIENE@@

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill (decomp37's recipes R1..R4 not run); nothing ran on the worker host.
- game.exe's consumer of the wound (`FUN_005a6920`) and the heal (`FUN_0057a980`) are read, not run under the emulator.
- The act-boss data word, `FUN_0056e970`'s stacking rule, the order of crush and Open Wounds on one blow (decomp37 D7).

@@RECHECK@@
