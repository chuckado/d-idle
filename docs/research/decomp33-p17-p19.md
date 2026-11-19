# P17 sequence rate, P18 Smite range, P19 Holy Bolt heal: re-derived against upstream main (DECOMP33)

> status: research, 2026-11-15, agent decomp33 (Sonnet), offline (emulator and fork builds on the NAS builder container, no game, no live d2d, no Go change, nothing posted; fork branches pushed to the private forge remote only). Builds on [decomp29-skilldamage.md](decomp29-skilldamage.md) (the deviation rows D3 to D5), [skiasgame2-trace.md](skiasgame2-trace.md) D1, [palskill-trace.md](palskill-trace.md) D1 to D8, [pal-skill-trace.md](pal-skill-trace.md) D1, [../upstream/pipeline-2026-11-14.md](../upstream/pipeline-2026-11-14.md) rows P17 to P19.

Tags: **V run** = game.exe 1.14d's own code run in the emulator this card (`tools/emu/dc33_seqrate.py`, `dc33_smite.py`, `dc33_holybolt.py`, copies of the fork's `seq_rate.py`, `smite_damage.py`, `holybolt_heal.py`) against d2d's real code in the fork's test binaries; **V read** = disassembly read; **D** = patched 1.14d table cell; **S** = an earlier trace of ours; **I** inferred; **U** unverified. Upstream main at the time: `d1a10a38`.

## Answer

1. **P17 is a proven deviation and has a fork branch.** A player's sequence skill in mode 0x12 loses 30 points of its attack percent in game.exe; d2d times it at the plain rate. 7,600 cells over the nine attack-branch sequence skills: parent 6,941 differ, claim 0. Branch `fid/seq-ias-r1` (oracle `320e3491`, claim `e177a4fa`), ready for the verifier.
2. **P18 is a proven deviation and has a fork branch.** Of the five Smite rows the palskill card listed, upstream already fixed the block one (D8); two remain after re-derivation: stat 111 missing from Smite's range (D1) and the maximum not forced one point above the minimum (D2). 75,600 cells: parent 63,709 differ, claim 0. Branch `fid/smite-r1` (oracle `486dc3bd`, claim `772458b6`), ready for the verifier. Reach is small (three uniques, one runeword, the Zakarum Shield).
3. **P19 is not a deviation for what d2d models.** Holy Bolt's heal of the caster's pets is in upstream (`skill_missile_hits`, hit function 7, `test_holy_bolt`): 2,240 cells (40 levels x 7 life/max pairs x 8 seeds) equal game.exe's FUN_005a9fb0, 0 differ, the draw equal. What is missing is a target d2d does not model: the mercenary (a monster whose owner is the player) is not struck by player missiles in d2d, and party allies need a party. No PR; an oracle-only branch `fid/holybolt-heal-r1` (head `21e0afb5`) holds the differential.

## P17: the sequence rate

### Q1 to Q6

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | What does game.exe do to a player's sequence rate? | FUN_00623f50's attack branch (`0x6241d7..0x62430a`): `esi = FUN_00621740(unit, 0) + stat 68`; `0x624290..0x62429b`: `cmp [unit], 0; jne; cmp [unit+0x10], 0x12; jne; sub esi, 0x1e`; clamp 15..175; rate = record `+0xc` x esi / 100, stored at `+0x4c` and `+0x3c` | V read; V run (below) |
| Q2 | Which skills take that branch? | FUN_006216e0 (`0x6216f5..0x62172c`) sends mode 0x12 with record byte `+0x12` (seqtrans) = 0xa (SC) to the cast branch; FUN_00621580 (`0x6215f4..0x621622`) sends a mode whose row (`0x6e8a00 + 20 mode`, `+8` set) has an active skill whose record `+6` & `[0x6ce274]` (= 8, V run) is set: UseAttackRate. V run on the 357 loaded records, mode 0x12, probe inputs stat 68 = 100, IAS 0, stat 69 = 130: 35 records give 179 (= 256 x 70 / 100, attack with -30): Jab 10, Impale 19, Double Swing 133, Double Throw 140, Frenzy 147, the four Assassin finishers 259 to 274 (Fists of Fire, Claws of Thunder, Dragon Claw, Blades of Ice) and monster skills; 2 give 256 (cast: Lightning 49, Chain Lightning 53); 79 give 332 (the other-animation branch with stat 69: Charge 107, Leap Attack 143, Whirlwind 151, Inferno 41, Arctic Blast 230 and the monster rows) | V run (`seq_rate.py table`) |
| Q3 | Is the rate the one the sequence steps by? | mode 0x12 runs FUN_00621260 (sequence init: `+0x3c` = 0x100, countdown `+0x48`) at `0x62442f` and then FUN_00623f50 at `0x62443f`, so the rate it stores replaces the init value; FUN_00623e00 takes `+0x3c` off the countdown each tick (`0x623e34`) | V read |
| Q4 | Length of a sequence at a rate | `ceil(256 steps / rate)` ticks: upstream's own reading (`docs/research/re/skills.md` "Cast rate", the merged Lightning sequence); not re-run here | S, U |
| Q5 | What does d2d do? | `Fight::start_sequence` (`components/game/fight/strike.cpp`): `attack_ticks(seq.size(), A1 AnimData speed, ias, wsm, sias + chill)`, no term for the skill; `swing_mode` plays SQ as A1 | V read; V run (the parent's grid) |
| Q6 | Is the mode really 0x12 for these skills? | the record byte `+0x10` (anim) is 0x12 for each of them (V run, table); that the unit's mode is the skill's anim is d2d's and the maintainers' existing reading of FUN_004cc5b0 (`swing_mode`) | V run (record), S (mode) |

### Differential (V run)

Game side `tools/emu/dc33_seqrate.py` (fork: `tools/emu/seq_rate.py`): the real Skills.txt records of the booted game.exe (`skill_tips.boot`), FUN_00623f50 natively in mode 0x12 for a player unit with the case's record as active skill (the stat getters return stat 93 = IAS and stat 68 = 100 - WSM + skill speed, `FUN_00620f00` is a no-op and the record's `+0xc` is d2d's speed). d2d side `test_seq_rate --seq-rate`: the real `Fight::set_pmode` + `start_sequence` for a hero of each class wearing each weapon; 17,600 cells; 7,600 compared (the other 10,000 are Lightning, Chain Lightning, Charge, Leap Attack, Whirlwind, Inferno and Arctic Blast, listed by name in the `CHECK` lines).

| 7,600 cells | differ | exit |
|---|---|---|
| parent `d1a10a38` + oracle commit | **6,941** (e.g. Jab, bare hands, IAS 0: d2d 13 ticks, game 19; Frenzy 17 vs 25; a chilled Jab at WSM 0: 26 vs 66) | 1 |
| claim | **0** | 0 |

`compare` exits 1 on an empty file and on a short line (both run). Mutants: `kSequenceRatePenalty = 0` and the dropped call-site term each abort `test_seq_rate`. Fork checks: `presubmit --quick` 0 FAIL, `pr_guard` 0/0, clang-tidy-19 0 findings in changed files, ctest 5 failures on both the parent (326) and the claim (327), the same five.

### Deviations (d2d vs game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | The 30 points: game.exe -30 for mode 0x12 + UseAttackRate + not SC; d2d none (fixed by `fid/seq-ias-r1`) | V run both sides | exists (above) |
| D2 | AnimData speed of a sequence: game.exe counts mode 0x12 on the `GH` record (FUN_00620f00 with no weapon: AMGHHTH, SOGHHTH, ..., speed 256 for all seven classes, V run); d2d uses the class's A1 record. The two tables differ for the Assassin's claws (A1 208, GH 256), a Paladin with a staff or a two-hand thrusting weapon (GH 192), a Druid with a one-hand sword (GH 248): record speeds read from the loaded AnimData (V run), but the lookup with a weapon was not run (the held-weapon path of FUN_00620f00 needs an item record the rig does not build) | V run (hth, tables), U (weapon path) | buildable: hook FUN_0063bde0 and FUN_006335f0 as `swing_start.py` does and call FUN_00620f00 for mode 0x12 with each weapon class key; compare to d2d's `player_anim(kModeA1).speed` |
| D3 | Sequences with UseAttackRate unset and not SC (Charge, Leap Attack, Whirlwind, Inferno, Arctic Blast): game.exe takes the rate from stat 69 (`other` branch); d2d times them at the attack rate | V run (the branch), U (stat 69 for a player) | buildable once stat 69's source is read |
| D4 | The step count and the `ceil(256 steps / rate)` model of a sequence | S upstream, U | FUN_005539b0 with `+0x30` set and the sequence object (FUN_006633b0 / FUN_006633d0 / FUN_006633f0) hooked |

### Bot findings (described, not changed)

- B1: no plan uses a sequence skill (Sorceress and Paladin plans); `internal/combat/rules.go:241` lists Double Swing, Frenzy and Whirlwind as physical names for damage typing only. Nothing in the bot depends on the length.
- B2: the resource simulator's melee swing is a constant (`internal/resource/sim.go:133`, `:150`, `:382`, see skiasgame2-trace B1); a sequence skill would be 46 percent longer at 0 IAS than a plain swing of the same frames.

### PR candidate

`fid/seq-ias-r1`: yes (the existing harness shape, parent differs, claim 0), ready for the independent verifier. Body `docs/upstream/bodies/seq-ias-body.md`, verifier brief `docs/upstream/briefs/verify-seq-ias.md`. Not posted.

## P18: Smite

### Re-derivation against upstream main

| Row (palskill-trace) | Upstream `d1a10a38` | Verdict |
|---|---|---|
| D1 stat 111 on both ends (FUN_0057b420 `0x57b4c3..0x57b4d6`) | `smite_lo / hi` = the shield's row (`combat.hpp` make_fighter) | **deviates**, fixed on the branch |
| D2 max at least min + 1 point (`0x57b4ea..0x57b4f4`) | none for Smite (the plain swing has it) | **deviates**, fixed on the branch |
| D3 -90 floor on the sum | gear part first | reach none (I), not claimed |
| D4 rounding with a negative percent | not re-run | not claimed |
| D5 256ths | upstream keeps 256ths (`phys_fine`) | already matches (the grid compares 256ths) |
| D6 to-hit roll | no roll, hit forced | not a deviation (palskill dead end), a different random stream only |
| D7 Smite without a shield | swings with 0 | U, not claimed |
| D8 (smite-zeal-sacrifice) block cancels Smite | `if (!swing.smite) return blow;` after the block roll: fixed upstream | already matches |

### Differential (V run)

`dc33_smite.py` (fork `smite_damage.py`): `emu.Emu()` on game.exe, FUN_0057b420 with the shield path (item non-null, `edx = 0`), the unit's seed words from the case; d2d `test_smite_damage --smite-damage`: `make_fighter` + `player_blow(swing.smite)` on a hand-made shield row (rules library only). 75,600 cells.

| 75,600 cells | differ |
|---|---|
| parent + oracle commit | **63,709** (stat 111 != 0: 42,116 of 43,200; min = max with stat 111 = 0: 4,315 of 4,320) |
| claim | **0** |

Mutant (no maximum rule) aborts the test. Checks as P17: presubmit 0 FAIL, `pr_guard` 0/0, tidy 0 findings in changed files, ctest the same five failures on the parent and the claim, `test_combat` passes after its Smite bounds went from exactly 15 to 15..16 (a 10 to 10 shield now rolls 10 to 11).

### Deviations still open

| D | Rule | Tag | Differential |
|---|---|---|---|
| D5 | -90 floor on the sum (game) vs gear part first (d2d) | V run (palskill, 214 of 1,732) | exists; no gear reaches it |
| D6 | rounding of a negative percent (1/256 differences) | V run, cause I | exists (`pal2` negative-percent rows); no PR |
| D7 | Smite with no shield: game.exe returns 0 and does nothing | V read | buildable: FUN_005ce9f0 with FUN_0063c8f0 hooked to fail |

### Bot findings

- B3: the Paladin plan never casts Smite (`internal/combat/rules.go:237` only classes it as a physical skill); the rule matters only for a shield-bash opener. `internal/gear/gear.go` and `gear_test.go:336` reason about Smite and shields as a skill name, no damage model.

### PR candidate

`fid/smite-r1`: yes, ready for the verifier; the verifier should weigh the small reach (body Limits). Body `docs/upstream/bodies/smite-range-body.md`, brief `docs/upstream/briefs/verify-smite-range.md`.

## P19: Holy Bolt heal

### Re-derivation

FUN_005a9fb0 (hit function 7), `0x5a9fb0..0x5aa111`: with the missile row's sHitPar1 (`+0x4c`) set and the owner found (FUN_00552fd0), FUN_005542c0 (TargetPet: the struck unit is type 1 and its owner is the missile's owner; skill flag `+7 & [0x6ce26c]`) or FUN_00554d20 (TargetAlly: the party test; `[0x6ce270]`) makes the unit a friend; its life becomes `min(life + roll, max)` with `roll = FUN_0045c3e0(seed at +0x20, calc2 << 8 - calc1 << 8) + calc1 << 8` (calcs by FUN_00646ca0), a roll of 0 or less taking FUN_005a8910's missile stat; the missile ends (return 1) and plays its overlay (word `+0x36`). Otherwise the unit goes through the undead and demon filter (`hit_par2`).

d2d (`Fight::skill_missile_hits`, `missiles.cpp`, `test_holy_bolt`): the same for a pet of the caster (`actor->pets`), with the fight rng; the party case needs a party (solo: none, as FUN_00540710 returns -1), the overlay is not played.

### Differential (V run)

`dc33_holybolt.py` (fork `holybolt_heal.py`): `skill_damage.Oracle` boot, FUN_005a9fb0 natively with the real calc machine and RNG, hooks for the owner, the missile's skill and level, the pet test (yes), the life getter and setter. d2d side `test_holy_bolt_heal --heal` (the real `skill_missile_hits`, `world.rng` loaded with the seed words): Holy Bolt levels 1 to 40, seven life/max pairs (1/100 ... 100/100, 5/6), 8 seeds: 2,240 cells, **0 differ** (1,603 cells end at the cap, 320 are a full-life pet). The comparison detects a change (one edited line: 1 differ). The build was of the oracle commit before the identifier rename (`212232f3`; the rename in `21e0afb5` is lint only).

### Deviations

| D | Rule | Tag | Differential |
|---|---|---|---|
| D1 | The mercenary (a monster owned by the player) is a Holy Bolt target in game.exe; d2d's player missiles never strike it, so it is not healed | V read both | buildable: a d2d probe with the merc in the fight (`actor->merc`) and a missile over it; game side as above with FUN_005542c0 run for a merc-like unit. A feature (a missile collision loop for the merc), not one rule: no PR now |
| D2 | Party allies (FUN_00554d20) | V read | needs a party in d2d: none solo |
| D3 | The bolt's overlay (`+0x36`, FUN_00621e40) | V read | cosmetic |
| D4 | The roll comes from the fight rng, not the missile seed (`ponytail` in the source) | S | the same distribution; the 2,240 cells load the seed into both |

### Bot findings

- B4: no party play and no merc-heal logic in the bot (`internal/combat` has no Holy Bolt cast); nothing depends on D1.

### PR candidate

None: d2d matches game.exe for what it models; D1 is a feature gap with no differential on a fix. The oracle-only branch `fid/holybolt-heal-r1` is kept (precedent: the melee-block oracle branch).

## Drill recipes (lab, one hidden slot; verbs of docs/forkdbg.md; not run)

- P17 (U for the AnimData record and the length): with a Barbarian wielding a sword and a Frenzy skill, `debug spawn` a dummy, `step 1` per tick, read the player mode and `state` from the first swing tick to the return to NU at IAS 0 and at +60: prediction after the PR `ceil(256 x steps / (speed x (70 + eff(ias)) / 100))` with steps from `sequences.hpp` (17 for the sword at the dump's weapon class), speed 256; Assassin with claws at the same: 208 (d2d) against 256 (D2 above).
- P18: a Paladin with a shield carrying stat 111 (`debug` item grant of a unique with `item_normaldamage`), Smite against a dummy at a fixed seed, read the damage from `dump`; the Zakarum Shield without a percent: damage 46 to 47 points.
- P19: a Paladin and a hired merc, a Holy Bolt cast through the merc: game.exe heals it (life up by calc1 to calc2); d2d's bolt passes.

## Matrix-cell updates (proposed text for `pal-skill-matrix.md` and `sorc-skill-matrix.md`)

- pal-skill-matrix, Smite row: "Damage range: stat 111 added and max >= min + 1 point (FUN_0057b420 0x57b4c3..0x57b4f4): d2d lacked both, 63,709 of 75,600 cells differ, fork branch `fid/smite-r1` [V run, decomp33]; the block no longer cancels it (upstream)".
- pal-skill-matrix, Holy Bolt row: "Heal of the caster's pets: d2d equals game.exe, 2,240 of 2,240 cells [V run, decomp33]; the merc and party allies are not targets in d2d [V read]".
- skiasgame2-trace D1 / sorc-skill-matrix plain Attack row: "sequence skills (UseAttackRate, seqtrans not SC) run at the attack rate minus 30 points [V run, 7,600 cells]; fork branch `fid/seq-ias-r1`".

## Unverified and dead ends

- U: the AnimData record game.exe counts a sequence at with a weapon; the step count; stat 69 for a player's non-attack sequences; Windows and macOS builds.
- Dead end: calling FUN_00620f00 with a fake inventory pointer faults inside FUN_0063bde0 (needs the 0x1020304 inventory magic and a full item record); a plain hook of FUN_0063bde0 returns no record because the later weapon-class steps need more of the item.
- Dead end: the first game-side rig read stat ids from the wrong hook argument and classified every skill as the other branch; fixed (`arg(1)` is the stat id).
- Process note: the builder container has the emulator tree and the game image (`/data/prfix6-emu`), so every run here used it, none the worker host.

## Shared-file lines for the merger

See `backlog/merge-lines/decomp33.txt`. Proposed glossary terms: none missing (sequence, UseAttackRate, seqtrans are in the skills docs).
