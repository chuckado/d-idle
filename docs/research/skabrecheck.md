# SKABRECHECK: two open claims of the SKRESIST A/B, run against game.exe and the lab binary's source (card skabrecheck)

> status: card skabrecheck, 2026-11-01, offline (no game, no d2d slot, no Go; nothing pushed or posted). Settles the two claims left open by docs/research/model-ab-skresist.md (section 5 E1, section 4 Sonnet D3). Worker host only, one run at a time, nice 10, never more than two processes (a capacity check first: idle above 25 percent).

Tags: **V read** read in the disassembly (upstream `tools/emu/disasm.py`) or in the d2d source; **V run** game.exe's own code run natively under the emulator and compared case by case with the real d2d code built from the lab binary's source; **D** patched 1.14d tables as game.exe's MPQ stack serves them (`tools/emu/skab_tbl.py`); **S** an earlier trace of ours; **I** inferred; **U** unverified.

**d2d side = the lab binary's source.** The differential is built from the fork branch `lab2-freezeresume` at `ee5e06d8`, the commit named in the lab binary `d2d.lab2-freezeresume` (its name carries that hash). The test binaries are linked from that tree (`test_skab_resist`, `test_skab_phys`, built in its build directory, `rules::missile_blow` real). I did NOT feed cases to the lab binary itself (it has no oracle mode); that the binary equals its source is the binary's own naming (U, not hash-checked here). The arm's own source pin (`7d5b3a5c`) is older; this card's numbers come from the newer one.

## Verdicts

| Claim | Verdict | One line |
|---|---|---|
| 1a. Fire / Lightning Mastery grant the pierce stats 333 / 334 (Sonnet B2) | **REFUTED** (D, V run of the loader path) | Fire Mastery writes stat 329 only, Lightning Mastery 330 only, Cold Mastery 335 only; 333 / 334 are written by no Sorceress mastery. |
| 1b. game.exe applies stats 333 / 334 differently from d2d | **REFUTED on the missile path** (V run: 0 differences, 73,622 pierce cases); **UNCOVERED on the melee path** (V read: d2d's `player_blow` has no pierce input; no run) | Same gate, same floor, same 100 line, for fire, lightning and cold. |
| 2. missile PHYSICAL damage: d2d cuts the roll to whole points before the resist (Sonnet D3) | **CONFIRMED** (V run, 101,626 cases: 84,271 differ on game.exe's own roll; 0 after a 6-line fix for res of -100 or more) | Local PR candidate `local/skabrecheck-phys` in the fork checkout, not pushed. |

## Claim 1: pierce

### Q1 What writes stats 333..336? (D)

`python skab_tbl.py` over Skills.txt (357 rows, all classes), ItemStatCost and Properties.txt, read through game.exe's MPQ stack:

- ItemStatCost 329..336 = `passive_fire_mastery`, `passive_ltng_mastery`, `passive_cold_mastery`, `passive_pois_mastery`, `passive_fire_pierce`, `passive_ltng_pierce`, `passive_cold_pierce`, `passive_pois_pierce` (D). So 333 is fire pierce, 334 lightning pierce, 336 poison pierce by name, not by order (the bot's comment at internal/combat/target.go:260 calls them "ItemStatCost order, S": now D).
- Skills.txt rows that name a pierce stat in any column: **two**: row 65 Cold Mastery (`passivestat1 = passive_cold_pierce`, calc `ln12`) and row 62 **Hydra** (`passivestat1 = passive_fire_mastery`, `passivestat2 = passive_fire_pierce`, both calcs `stat('..'.accr)`: Hydra copies the caster's mastery and fire pierce onto its own unit; reading of `.accr` as "the caster's total" is I).
- Fire Mastery (row 61): `passivestat1 = passive_fire_mastery` only (329). Lightning Mastery (row 63): `passive_ltng_mastery` only (330). Neither has a second stat. This matches `sk-passives-trace.md` Q3 / Q4 (S) and closes E1 of the A/B document: **Sonnet's "Fire / Lightning Mastery carry stats 333 / 334" is wrong** (D).
- Item sources of 333 / 334: Properties.txt `pierce-fire` / `pierce-ltng` / `pierce-cold` / `pierce-pois` (D) write exactly these stats on the wearer. So fire / lightning pierce on a Sorceress comes from gear (and Hydra's copy), never from a skill she levels.

### Q2 The rule at hit time (S from `sk-resist-trace.md` Q4, re-run here)

FUN_0057c1e0 -> FUN_0057bf80 -> FUN_0057be00 on the descriptor rows at 0x732980 (+0xc pierce stat: fire 333 / 0x14d, lightning 334 / 0x14e, cold 335 / 0x14f, poison 336 / 0x150): the pierce is subtracted from the defender's resist only when the resist is under 100 (a plain monster), floored at -100; then `MulDiv(damage, 100 - res, 100)`, 0 at res 100 or more (V read in the earlier trace; this card re-ran it, below). d2d: `rules::missile_blow` (components/rules/combat.hpp, the `res < 100` line, `max(res - pierce[etype], -100)`) and `Fight::pierce()` (fight/missiles.cpp:538, `psum[333..336]` in fire, lightning, cold, poison order; V read).

### Q3 Differential (V run)

`tools/emu/skab_run.sh resist 7`: `sk_resist.py gen 100000 7` (32,604-style boundary grid plus 100,000 random: **147,628 cases**; own resist -150..1000, pierce 0 / 15 / 60 / 115 on the grid and 0..200 random, curse delta 0..-300), game.exe side `sk_resist.py game` (FUN_005c3540, FUN_0057be00, FUN_0057c6c0 natively), d2d side `test_skab_resist` (= `sk_resist_test.cpp`: `adj` and `effres` copy d2d's lines, `dealt` is the REAL `missile_blow`). The totals reproduce the Opus run of the earlier card exactly (seed 7: 88,224 `adj`, 92,149 `effres`, 14,289 `dealt` differ), now on the lab binary's source.

Split by `tools/emu/skab_pierce.py` (cases with a pierce stat above 0, per element; `pure` = no curse):

| element | class | cases | `dealt` differ | `effres` differ |
|---|---|---|---|---|
| fire 333 | pure (no curse) | 2,722 | 0 | 0 |
| lightning 334 | pure | 2,632 | 0 | 0 |
| cold 335 | pure | 2,756 | 0 | 0 |
| all three | pure, on the 100 line (own resist 99 / 100 / 101) | 85 + 87 + 84 | 0 | 0 |
| all three | pure, result near the floor (own - pierce <= -95) | 351 + 342 + 364 | 0 | 0 |
| all three | a curse on an own resist under 100 (the cut does not apply) | 6,260 + 6,252 + 6,313 = 18,825 | 0 | 0 |
| all three | a curse on an own resist of 100 or more (**D1 of the SKRESIST card, the one-fifth cut, not this claim**) | 46,687 | 8,847 | 43,933 |

Reading (V run): for fire and lightning exactly as for cold, with or without a plain curse, d2d's pierce gate, floor and 100 line equal game.exe on **26,935 pierce cases (8,110 pure + 18,825 curse) with 0 differences**; every difference in the whole file sits in the already-known D1 (curse on an immune base) or D2 (magic / physical under -100) classes. The arm's Fire / Lightning Mastery text has no rule difference behind it either.

### Q4 The melee path (V read, not run)

`player_blow` (combat.hpp, the player's melee, strike.cpp:564 and ten other call sites) takes no pierce argument: weapon elements (`fighter.elem`), Vengeance's shares and conversion go through `resisted(.., res)` on the raw resist. game.exe puts the same pierce through FUN_0057be00 for every record element, melee included (the context is built once per hit by FUN_0057c1e0; V read of the earlier trace). So a melee build with an item that has `pierce-fire` / `pierce-ltng` and a weapon element is hit by a rule difference (the Opus D3 of the SKRESIST card). **UNCOVERED by a run**: no melee differential carries a pierce stat. How to build it is in D2 below.

## Claim 2: the physical roll of a missile

### Q5 game.exe (V read, then V run)

FUN_005a89a0 (the record builder for a missile hit) at 0x5a89a0: it reads stat 0x16 (max) then 0x15 (min) of the missile with FUN_00625480; **if either is 0 or less the physical is 0**; if min is over max they are swapped; the physical is `min + rand(max - min)` on the missile's seed (+0x20, FUN_0045c3e0; exclusive of max), stored at record +8 **in 256ths**; stat 0x19 (percent), 0x79 / 0x7a (versus demon / undead) scale it by `MulDiv`-style `x * pct / 100` (0 on a skill missile, I). The fire pair (0x30 / 0x31 with 0x149) is rolled next, the same way, to record +0x10. FUN_0057c6c0 then applies the record: FUN_0057bf80 takes the flat reduction, **MulDiv(field, 100 - res, 100) on the 256ths** (FUN_00483360), and the total comes off life stat 6 in 256ths (V read, `sk-missile-trace.md` / `pr-missile-damage-fine-proof.md`). So game.exe keeps the physical fraction, as it does for an element.

### Q6 d2d (V read, lab binary's source `ee5e06d8`)

`rules::missile_blow`, physical part: `physical = rng.range(phys_lo, phys_hi) >> 8; blow.phys += physical; blow.damage += resisted(physical, res[0])`. Whole points before the resist (up to a point lost per hit, more at a negative resist), **`frac` is never set for the physical** (it is set only in the element branch). This settles the "open tension" of the A/B document (section 4: `Blow::frac` "only from missiles"): `frac` exists for the element only; Sonnet's reading was right, and `sk-melee2-trace.md`'s "frac only from missiles" means the element. `Rng::range(min, max)` is inclusive of max (the known D1 of `sk-missile-trace.md`), and a zero min with a positive max is rolled by d2d and not by game.exe (this run measures it, D3 below).

### Q7 Differential (V run)

`tools/emu/skab_run.sh phys 7`: `skab_phys.py gen 100000 7` = **101,626 cases** (a boundary grid of physical min / max in 256ths from 0 to 25,600, resist -150..150 for physical, plus 30 percent with a fire pair, plus 100,000 random up to 0x180000 (10 percent) / 0x8000), game.exe side `skab_phys.py game` (FUN_005a89a0 on the case's seed, then FUN_0057c6c0; life lost in 256ths; hooked, not the claim: the stat getter, life_apply.py's unit kinds, events and side effects), d2d side `test_skab_phys` (`skab_phys_test.cpp`, the REAL `missile_blow`). Two d2d columns: **A** the same seed (includes the roll range), **B** d2d fed game.exe's own roll (`min = max = rp`, so no draw: isolates the whole-point arithmetic). Before the fix (lab binary's source), class by class, 256ths compared case by case:

| class | n | A differ | B differ | mean game minus d2d on B (points) |
|---|---|---|---|---|
| physical, res -100..-1 | 23,403 | 23,395 | 23,383 | +1.205 |
| physical, res 0 | 266 | 260 | 260 | +0.480 |
| physical, res 1..99 | 23,462 | 23,431 | 23,414 | +0.720 |
| physical, res >= 100 | 11,884 | 0 | 0 | 0 |
| physical, res under -100 | 11,532 | 11,526 | 11,525 | -114.0 (the floor, D4) |
| physical + fire pair, res -100..-1 / 0 / 1..99 | 10,211 / 287 / 10,313 | 10,197 / 275 / 10,286 | 10,180 / 261 / 10,252 | +1.20 / +0.40 / +0.71 |
| physical + fire pair, res >= 100 | 5,136 | 4,222 | 0 | 0 |
| physical + fire pair, res under -100 | 4,996 | 4,995 | 4,996 | -117.6 |
| zero min (lo or hi 0) | 25 + 12 + 62 + 13 + 24 | 10 + 5 + 12 + 6 + 0 | 0 | 0 |
| **total** | **101,626** | **88,620** | **84,271** | |

Every B difference is the truncation (a value of `rp & 255` lost before the resist, then scaled, at most 255/256 of a point times (100 - res)/100) or the floor class (res under -100). Spot values (case 3001: roll 26,237, res 93, game.exe takes 1,836 in 256ths = 26,237 x 7 / 100; d2d takes 7 whole points = 1,792). At res 0 the mean loss is 0.480 points a landed hit (= half a point, the mean fraction 255/512); at res 1..99 about 0.72 points; the whole-point rule never deals a fraction, so a missile that rolls under one point deals 0.

### Q8 The fix and the proof (V run, local branch, not pushed)

Branch `local/skabrecheck-phys` (one commit on `ee5e06d8` in a separate worktree of the fork repo; never pushed, never posted; AGENTS.md's same-day-push rule is overridden by this brief): the physical part of `missile_blow` becomes

```
-        const int physical = rng.range(damage.phys_lo, damage.phys_hi) >> 8;
-        blow.phys += physical;
-        blow.damage += resisted(physical, target.res[0]);
+        const int roll = rng.range(damage.phys_lo, damage.phys_hi);   // 256ths
+        blow.phys += roll >> 8;
+        const int fine = resisted_fine(roll, target.res[0]);
+        blow.damage += fine >> 8;
+        blow.frac += fine & 255;
```

(`resisted_fine` already exists above it for the element; no other file). The same harness (`tools/emu/skab_fix.sh before|after`, the test compiled standalone against each tree's header: the physical test needs only the header-only rules code; the standalone `before` reproduces the built test exactly, 84,271) gives **B differ 84,271 before, 16,526 after, and 0 after for every case with res of -100 or more**; the 16,526 left are all the res under -100 class (D4 floor, latent: MonStats physical min is 0 and Amplify Damage gives exactly -100, `sk-resist-trace.md` D2). A differs 88,398 after (roll range, the SKMISS D1, a separate claim). A PR for this alone would carry the floor as a separate line (`max(res, -100)`) only if the owner wants D4 closed with it.

PR status: **candidate, NOT ready**. Before it can go anywhere: (a) an independent verifier with its own harness (the harness here is this card's own); (b) the caller audit: `strike.cpp:464` hands `frac` to `hurt` only when `taken == blow.damage` (Bone Armor / soak drops it) and `Fight::land` for a missile with `srcdam > 0` first runs `player_blow` and then `missile_blow` on the same `Blow` (Blade Fury, Blade Shield), so the sum must be checked on those two skills; (c) deterministic-hash scenarios that include a physical missile change (any Druid / Assassin / Barbarian scenario with Raven, Twister, Tornado, Shock Wave, Blade Fury, Blade Shield); (d) upstream's PR standard (docs/process/upstream-pr-standard.md). Reach: the skills with a physical MinDam / MaxDam at base (D, `skab_tbl.py phys`): Magic Arrow (1-1), Holy Shield, Raven, the Druid summons' hits, Twister, Tornado, Shock Wave, Blade Sentinel, Blade Fury, Blade Shield, Mind Blast, traps and the monster rows; **no Sorceress skill** (so no effect on the bot's current plan).

## Deviation list (d2d vs game.exe)

| D | game.exe | d2d (`ee5e06d8`) | Tag | Differential |
|---|---|---|---|---|
| D1 | missile physical rolled in 256ths, resisted in 256ths, life takes the fraction (Q5) | rolled, shifted to whole points, then resisted; `frac` never set for the physical (Q6) | **V run, CONFIRMED**: 84,271 of 101,626 differ on game.exe's roll; 0 after the fix at res >= -100 | yes, built and run: `skab_phys.py` + `test_skab_phys` + `skab_fix.sh` |
| D2 | pierce on melee records (the same FUN_0057be00, every record element) | `player_blow` has no pierce input; weapon elements, Vengeance, conversion use the raw resist | V read both sides; **no run** | buildable: extend `sk_melee2.py` (game: FUN_0057b420 / FUN_0057b7d0 then FUN_0057c6c0 with stat 333 / 334 / 335 on the attacker unit, a weapon with an element, resist 0..99) and `sk_melee2_test.cpp` (`player_blow` with `Fighter::elem`); oracle = the record's life lost; expected: every case with pierce > 0 and resist under 100 differs |
| D3 | a physical min or max of 0 means no physical roll (Q5) | `phys_hi > 0` rolls `range(0, hi)` (Q6) | V run (the zero-min class: 99 cases, d2d deals damage game.exe does not; mean up to 12 points) | yes (same harness); **reach U**: no Skills.txt row has base MinDam 0 with MaxDam above 0 (D, `skab_tbl.py phys`); the level columns were not scanned |
| D4 | -100 floor on every resist, the physical included | `resisted(.., res)` / `resisted_fine` without a floor on the physical and magic paths | V run (16,526 left after the fix, all res < -100); latent (`sk-resist-trace.md` D2) | yes (same harness) |
| D5 | physical and fire rolls on one seed: physical first, then the fire pair, `min + rand(max - min)` | the same order, `range` inclusive of max (SKMISS D1, known) | V run: A differs 88,398 after the fix | yes (same harness, column A) |
| equal | pierce on missiles (fire 333, lightning 334, cold 335): gate under 100, floor -100, the 100 line, with or without a plain curse | the same | V run: 73,622 pierce cases, differences only in the curse-on-immune class (D1 of SKRESIST) | done |

## Bot findings (described, not changed; main at the time of the card)

- **B1** the bot already takes the pierce per element from the character's stat list by id, whatever skill or item set it: `PierceFromStats` (internal/combat/target.go:259-270, `pierceStat = {0, 0, 333, 334, 335, 336}`), subtracted below 100 and floored at -100 in `Target.EffRes` (internal/combat/rules.go:159-166), the same rule as game.exe (V run, Q3). The A/B Sonnet B2 ("no pierce") describes a tree before that code and is outdated; its Fire / Lightning Mastery sentence is wrong (Q1). Nothing to change. Tag upgrade only: the comment at target.go:260 may say "D: ItemStatCost names" for 333 / 334 / 336.
- **B2** nothing in the bot reads stats 329 / 330 as pierce (`grep` of internal for the stat ids and the mastery names finds no use): no model error to fix (V read).
- **B3** a Sorceress build must not plan fire / lightning pierce from skills: the only sources are gear properties (`pierce-fire`, `pierce-ltng`) and Hydra's copy of the caster's stat (D). A gear scorer that wants it would read the property, not a skill level.
- No bot finding for claim 2: no Sorceress skill reaches the physical missile roll.

## Dead ends and limits

- The lab binary cannot run an oracle; the differential links the fork's rules and game libraries built from its source instead (U for "binary equals source").
- A first standalone attempt was unnecessary for `before` (the built test ran); the standalone compile is used only to prove the `after` header without rebuilding the game library (its `before` numbers equal the built test's, 84,271).
- Not covered: absorb stats (142..149) and the flat reduction on the physical path (stats 34 / 35 assumed 0 on monsters); Hydra's missile carrying the caster's pierce in d2d (I only); the soak (Bone Armor) interaction with a nonzero physical `frac`; the whole-point physical in `player_blow` (the melee2 D1, 80,250 of 80,632, S).
- The physical test is a two-seed-free single run (seed 7); the arithmetic is deterministic, a second seed adds nothing to the B column but was not run (U).

## Drill recipe (fork verbs, docs/forkdbg.md and docs/forkcheats.md; the lab binary `d2d.lab2-freezeresume`, NOT run here)

One hidden slot, a throwaway offline character, lockstep, every cheat ledgered with the reason "SKABRECHECK drill Rn".

- **R1 fire / lightning pierce is item-only (Q1, claim 1a)**: `debug points 61 5` (Fire Mastery 5) and `debug points 63 5` (Lightning Mastery, needs clvl 30: `debug stat 12 30`), then `dump` and read the stat list: expected stat 329 = 30 + 7 x 4 and 330 = 50 + 12 x 4, stats 333 / 334 absent or 0 (U until read).
- **R2 pierce on the missile path (claim 1b)**: `debug spawn zombie1 1 near`, `debug mon set <id> life 100000 maxlife 100000 res light 50`; cast Lightning 200 times with and without a stat 334 of 30 on the character (`debug stat 334 30` if the verb takes a pierce stat: U) and sum the life lost: expected 0.50 vs 0.80 of the unreduced damage in both game.exe and d2d (res 50 less 30 = 20). Then `res light 100`: 0 in both with any pierce.
- **R3 physical missile (claim 2)**: needs a class with a physical missile, which the Sorceress is not (Druid Raven or Assassin Blade Fury throwaway; creating one is not this card's to do): `debug mon set <id> life 100000 maxlife 100000 res phys 0`, 500 casts of a skill whose missile has a small physical (Magic Arrow 1-1 at weapon-less levels), compare the mean life lost per hit: game.exe keeps fractions (the base after the roll; Magic Arrow is 1.0), d2d truncates; a skill whose roll is under one point shows 0 in d2d (U).

## Matrix-cell updates (proposed text for docs/research/sorc-skill-matrix.md)

- Row 8 (SKRESIST): append "SKABRECHECK (2026-11-01): pierce rule equal on the missile path for fire / lightning / cold (V run, 26,935 pierce cases, 0 differ); Fire / Lightning Mastery are damage % (stats 329 / 330), 333 / 334 come from gear properties and Hydra's copy (D); docs/research/skabrecheck.md".
- Fire Mastery (61) and Lightning Mastery (63) rows: replace any "pierce" text with "[D] stat 329 / 330 damage only, no pierce".
- Cold Mastery (65): unchanged ("pierce stat 335 under 100, V run", now 2,756 pure cases 0 differ on the lab binary's source).
- Hydra (62): add "(D) passivestat1 fire_mastery and passivestat2 fire_pierce copy the caster's stats onto Hydra; whether d2d's Hydra uses `psum` for the same: I".
- Physical missile rows (no Sorceress row): "d2d truncates the physical roll to whole points, game.exe keeps 256ths: V run 84,271 of 101,626 (docs/research/skabrecheck.md D1)".

## Missing glossary terms (for the report)

"pure pierce case", "game's roll vs same seed (the A and B columns)".

## Shared-file lines for the merger

See `openspec/changes/skabrecheck/merge-lines.txt`.
