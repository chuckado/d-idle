# decomp28: what interrupts a Town Portal cast, the bot's crowd guard, B2 to B4, and the danger units fit

> status: finished 2026-11-15 (card decomp28, Sonnet). Offline: the emulator image on the worker host (game.exe's own functions, docs/emu-image.md); no game, no live d2d, nothing posted, no Go change. Tags: V run (run here against game.exe's real code), V read (disassembly), D (MPQ tables), S (an earlier doc), I (inferred), U (unverified). Consumers: the survival reflex layer (R6, docs/survival/design-2026-11-14.md), `internal/danger` (units per clear), the matrices (section 9). Data: `docs/research/data28/`. Fit report: [docs/survival/units-fit-2026-11-15.md](../survival/units-fit-2026-11-15.md). Data file for the generator: `internal/danger/areaunits.json` (not read by any code yet).

## 1. Answer in ten lines

1. **No hit, block or knock-back can interrupt a Town Portal or Identify cast in game.exe** (V run, 14,400 cases): while the scroll or tome row is the used skill and the cast is running, the only new modes the game accepts are death (0), the unit's own next attack or cast (A1, A2, SC, TH), S1 (13, reached only by an Amazon's Dodge or Avoid) and 18. Get-hit (4), block (9), knock-back (19), walk, run and idle are refused.
2. So the bot's R6 premise ("a melee hit cuts a portal cast", `internal/resource/abort.go:79`, `internal/combat/safety.go:146`) is **false on game.exe**. d2d agrees only with the P04 claim (`fid/hitrec-interrupt-r1`); before it d2d cut the cast (decomp26 D2).
3. The guard's **effect** is still defensible: the cast plus the walk into the portal are exposure time (the death in `docs/fixportalmelee.md` was damage taken during the cast, not a cut). The guard is **over-cautious** (it fires on a count of three adjacent, whatever the life and the damage) and **wrongly reasoned**. A design card (section 5) refines it; nothing was removed or changed.
4. **Buffs and normal casts ARE interrupted** by a hit (335 of 357 rows); the buff recast-after-a-cut logic (B3) and `fightBuff`'s skip with three adjacent are right.
5. Two things the bot's model lacks: the **Concentration aura** (state 0x2a) makes a cast uninterruptable with the chance of its stat 164, and the **Concentrate** skill's state 0xf makes everything uninterruptable except death (V run, branch logic; the real stat values U). A Hammerdin standing in Concentration is partly protected from cuts.
6. **B4 is a real finding** (not "U, no change"): `internal/power/margin.go:181` applies the hero's block to every ranged attacker, but game.exe lets a shield block a missile only when its physical record is over 0 (decomp26 Q10 to Q13): Fallen Shaman fire, Radament's bolt, vampire fire and the mummies' poison are unblockable.
7. **Danger fit**: the shipped units-per-clear estimate has a 94 percent median error over 67 Act 1 and Act 2 Normal areas (Act 1 alone 41 percent, Act 2 alone 147 percent). The missing term is the level's real area: game.exe's generator builds 8x8-tile rooms, and units per clear are about 0.02 to 0.08 per tile. With the tile count as a data column and the maze rate tied to MonDen the median error is 16 percent in sample, 17 percent leaving each area out and 17 percent leaving each family of areas out (mazes 13 to 14, outdoors 19, **presets 54: not fixable from tables**).
8. **Pack size does not help**: the exponent of the mean group size is 0.0 to 0.1; the generator's `MonDen x groupMean x k` multiplies by a term the population does not follow, and `MonDen` itself only matters in mazes (exponent 0.41).
9. **The per-screen density is wrong in mazes**: `Zone.Density.PerScreen` divides by the Levels.txt size (200 x 200 for mazes); the real area is 400 to 4,500 tiles, so the shipped value is 10 to 30 times too low (game.exe: 10 to 22 units per 280-tile screen; shipped 0.4 to 3). The "bad" encounter in a maze is therefore close to the typical one.
10. Nothing deviates in a way that makes a PR candidate here: the interrupt claim is already P04; the Concentration chance is a d2d deviation by V read only (no differential run).

## 2. Questions and answers (task 1: the interrupt rule)

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | Which code decides whether a unit may enter a new mode while it casts? | Two tests, both needed, called by `FUN_005809d0` (player mode switch) and `FUN_00580a70` (the knock path): `FUN_0057edd0` (current mode against new mode, class gates; `0x580a02`, `0x580adc`) and `FUN_0057eec0` (`0x580a1c`, `0x580af6`; the gate of the used skill; also called at `0x54d7b0`, `0x581147`). Both are skipped only when the last argument of the switch is non-zero. | V read |
| Q2 | Which modes can a hit start, and from where? | The player branch of `FUN_0057cee0` picks by result flags: `0x80` (state 65 Dodge) and `0x100` (state 66 Avoid) go to mode 13 (S1) at `0x57d1f8`; `0x200` (state 68 Evade) calls `FUN_00553380`; `0x8010` (block) goes to mode 9 at `0x57d2c5`; `2` (death) mode 0 at `0x57d324`; `8` (knock) mode 0x13 at `0x57d356`; `4` (get-hit) mode 4 at `0x57d394`. State names from States.txt. | V read + D |
| Q3 | What does `FUN_0057eec0` do with a used skill whose `interrupt` column is blank? | The switch on the new mode minus 7 (jump table `0x57f000`, bytes at `0x57f008`): new modes below 7 (idle, walk, run, get-hit, town modes) are refused; of 7 to 18 only 7, 8, 10, 11, 13, 18 may pass, and only when `FUN_0057ed70` (the unit's queued end event is no more than 5 frames away from now) says so; 9, 12, 14 to 17 are refused. New mode 0 (death) is accepted before the used skill is looked at; state 0x36 refuses even that (V run). | V read (`0x57eec0..0x57f008`) + V run |
| Q4 | With the `interrupt` bit set? | The unit enters any mode unless it has state 0x2a (Concentration aura) and a roll of 0 to 99 (`FUN_0045c390`, the unit's seed) is below its stat 0xa4 (164, `skill_concentration`), or state 0xf (Concentrate): then it refuses every mode but idle (current mode NU, `0x57ef91..0x57ef97`) and death. | V read; V run with a stat of 50 and rolls 10 and 90 (the real stat per aura level is U) |
| Q5 | Result for the Town Portal scroll (219) and tome (220), the Identify scroll (217), Zeal (106) | Used skill with the cast running (current mode SC, end event queued): allowed new modes = {0, 7, 8, 10, 11, 13, 18} (`data28/tpint-summary.txt`). The same set from A1, TH, WL, RN as the current mode; from NU everything. With the end event already more than 5 frames past or none queued: {0}. Concentration or Concentrate states change nothing for these rows. | V run (`tools/emu/dc28_tpint.py game`, 7 classes x 8 skills x 5 state sets x 3 event cases x 6 current modes x 20 new modes = 100,800 evaluations; 14,400 Paladin lines, 0 class-dependent lines) |
| Q6 | Result for an interruptible cast (Fire Bolt 36, Teleport 54, Attack 0) | Every new mode 0 to 19 is allowed while the event is live; with no event queued only {0, 1, 4, 5, 9, 17}, so a hit (4) and a block (9) still interrupt. State 0xf: {0}. State 0x2a with the roll under the stat: {0}; with it over: everything. | V run |
| Q7 | So what ends a Town Portal cast? | Death. An Amazon's Dodge or Avoid (mode 13), reached from flags `0x80` / `0x100`. The unit's own next command that is an attack or cast (7, 8, 10, 11) once its end event is near. Nothing else: not a hit, not a block, not a knock-back, not a walk click. Stun (state 21) is a state that stops the result flag 4, not a mode (`movement-and-recovery-trace.md` row 49): not run (U). | V run (modes), V read (flags) |
| Q8 | Does reading the scroll set the used skill? | The scroll rows exist in Skills.txt (ids 217 to 220, `interrupt` blank, D) and decomp26 found their `+7` bit clear; that the scroll or tome read goes through the skill start (so that `FUN_00620250` is non-null during the cast) is not read here. | I; drill R3 (section 8) |
| Q9 | How long is the exposure? | `docs/fixcrowd.md` (S, DRILLS11 live on d2d): the portal object appears 4 steps (320 ms) after the click, the character stays in cast mode 6 steps; `docs/fixportalmelee.md` quotes about 1.3 s from the soak log. The two do not agree and neither was measured on game.exe. | U |

## 3. Deviations (d2d against game.exe)

| id | Rule in game.exe | Rule in d2d | Tag | Differential |
|---|---|---|---|---|
| D1 | A hit (4), block (9) and knock-back (19) are refused while the TP or Identify row is the used skill | parent: every cast is cut by a hit (decomp26 D2); claim `a78f66a3`: `Fight::hit_interrupts` (`components/game/fight/tick.cpp:576`) reads the row's `interrupt` for a hit | V run (game side) + V read (d2d) | yes for the hit: P04's `interrupt_swing.py`; the block and knock-back cells are I (no d2d test emits them): extend `interrupt_swing.py game` with new modes 9 and 19 (this card's `dc28_tpint.py` is the game side) and `test_hit_recovery` with a block and a knock case |
| D2 | A cast with `interrupt` set is uninterruptable when the unit has Concentration (state 0x2a) and the roll is under stat 164, and always while the Concentrate state (0xf) is on | not implemented: no read of state 0x2a, 0xf or stat 164 near `hit_interrupts` (grep of `components/game`) | V read (d2d absence), V run (game branch) | buildable: the game side is `dc28_tpint.py` with real stat values (Skills.txt aura table for skill 123 at levels 1 to 20); the d2d side needs a player with the aura and a seeded roll; the oracle must compare the roll sequence (unit seed `+0x20`), so it needs the RNG of d2d's hit path |
| D3 | An Amazon's Dodge or Avoid (flags `0x80` / `0x100`) enters S1 and can end a cast | not looked at | U | buildable with the same game-side harness and class 0 |

No PR candidate: D1's hit part is P04 (already proven and verified elsewhere); D2 and D3 have no d2d-side differential run (the rule needs one first; recipe above).

## 4. Bot findings of decomp26 (B1 to B4), checked

| id | Where | Result |
|---|---|---|
| B1 | `internal/resource/abort.go:79` (`castCut`), `internal/resource/types.go:81` (`CastAdjacentMax` 3), `internal/combat/safety.go:146` (`meleeCastGuard`), `internal/combat/crowd.go:18,43`, `internal/combat/mercfield.go:52`, `internal/combat/resource.go:103` | The premise is false on game.exe (Q5, Q7). Verdict: **wrong reasoning, right instinct, over-cautious as written**. Detail and the proposed refinement: section 5. |
| B2 | `internal/town/trip.go:210` (`openPortal`, 3 s wait) | Confirmed: nothing in game.exe makes a hit stop the portal, so "town portal did not appear" after a hit is a d2d (parent) symptom or another cause. The 3 s is generous (portal at 320 ms, S). No change; the warning stays as a signal of a real problem. |
| B3 | `internal/combat/fight.go:310` (armor upkeep), `internal/combat/fighter.go:679` (`buffPass`, `watchCast` cut test, `buffTriesMax`) | Confirmed and right: Frozen Armor and every row with `interrupt` set is cut by a hit (Q6), so the recast-on-cut logic is live; Zeal and Smite chains are never cut. `fightBuff`'s skip with three or more adjacent (`fighter.go:748`) rests on the same true premise for buffs (not for the portal). One addition: with Concentration on (Hammerdin), some cuts do not happen; the logic is fine either way. |
| B4 | the Fallen Shaman threat in the bot's models | **Found**: `internal/power/margin.go:181` (`hitOn`) multiplies the hit chance by `1 - BlockPct/100` for every attacker including `Ranged` ones; `internal/power/zone.go:27` (`meanMon`) builds the ranged mean monster without a block flag. In game.exe a missile is blockable only if its physical record is over 0 (decomp26 Q10 to Q13: `FUN_005ad730` -> `FUN_0057dfb0`); the shaman's, Radament's, vampires' and countess's fire and the mummies' poison are not (54 of 57 skill-launch records have no physical). The combat threat table (`internal/combat/affix.go:66`) does not use block, so the combat side has no error. Effect: a 60 percent block Paladin is credited with 60 percent less ranged damage from casters than he takes; the zone margin for Black Marsh, Dark Wood, the Catacombs and all caster-heavy zones is optimistic. Proposed change section 5.2. |
| B5 (new) | `internal/danger/generate.go:186` (PerScreen) | `Zone.Density.PerScreen = Units / (SizeX x SizeY) x ScreenTiles` uses the Levels.txt size, which is 200 x 200 for mazes and 80 x 80 for outdoors; the code comment of `params.go` already says those sizes are not the real areas. See section 7 and the table of the fit report. |
| B6 (tool note) | `tools/emu/mode_switch.py` | `NAMES` lists mode 18 as KB; the hit code pushes 0x13 for a knock, so 19 is the knock-back mode and 18 is not shown to be one (I: sequence). Labels only; the cells are by number. |

## 5. Design cards (not applied; the user decides)

### 5.1 Card: the portal escape judged by exposure, not by a cut

Evidence: Q5, Q7 (V run); the death that created the rule was damage during the cast (`docs/fixportalmelee.md`: 14 life, 3 to 5 undead adjacent, the scroll was spent and the character died 22 ms of wall time later). The guard's two readings: (a) three adjacent hits cut the cast: false. (b) the cast and the entry take time during which attackers hit: true.
Today: `Adjacent >= 3` and in a fight and the decision was a portal: drink if a potion is ready, else leave the game. At 90 percent life with a Rejuvenation ready this still drinks and the portal is skipped; at 30 percent with no potion it chickens, which is the right answer for the death case.
Proposal (data-driven, keeps the safety net): replace the count by an expected-damage test over the exposure. `exposure = castSeconds + enterSeconds` (knobs with today's guess as defaults, measured later by drill R3), `incoming = sum of Dps of the adjacent monsters` (the danger table's `Monster.Dps`, internal/danger, already per monster; the resource input carries `PackDanger` and `Hostiles`); the guard fires when `incoming * exposure >= life * margin` (margin knob, default 1.0 to start; with the default exposure of about 1.3 s the same three-adjacent soak case still fires). Where the damage is unknown (no table), keep the count rule. Build-specific data: the exposure comes from the class's cast frames (Skills.txt / class data), not a literal; the test is written against two classes (Sorceress and Paladin). Files: `internal/resource/abort.go`, `internal/combat/safety.go`, the knob in `internal/env/tuning_combat.go`, the drill wait in `internal/drill`.
Risk: if `Dps` is under-estimated the guard stops firing where it protected; that is why it is a card, not a patch, and why drill R3 (portal exposure on the lab binary) comes first.
Alternative with no code risk: leave the behaviour and correct the comments and docs to the true reason (exposure, not interruption) and tag the three-adjacent number as a judgement (U). I recommend doing this at once.

### 5.2 Card: block does not apply to unblockable ranged attacks

`internal/power/margin.go:181` and `internal/power/zone.go`: add `Unblockable bool` to `Mon` (true for a ranged monster whose missile has a zero physical record; the monster table needs one column per ranged species: `MonStats` skill and `Missiles.txt` `MinDamage/MaxDamage` and the monster's `A1MinD` for arrows; the decomp26 oracle `tools/emu/missile_block.py` decides each row) and skip the block factor for it. Test against a shield Paladin and a Sorceress. Data: a table `internal/danger` can read, built by the oracle. Until then the cheap bound: apply block to a ranged monster only when its hazards list has no element (arrows) (I).

## 6. What the bot does around a cast (read, not changed)

`internal/town/trip.go:210` (`openPortal`) right-clicks the scroll in the open pack, closes the panels, and waits up to 3 s for the portal object; no step depends on the cast not being cut. `watchCast` (`fighter.go:728`) reports a cut when the player mode is 4 at any poll, which is true for an interruptible buff and never for a TP read (the mode stays 10). Mode 4 during a TP read in game.exe cannot occur (Q5), so a cut flag in a TP drill is a d2d defect, not a fact about the game.

## 7. Task 3: the danger units fit

### 7.1 Method (V run)

1. Game side: `tools/emu/dc28_batch.sh` runs upstream's room population of every Act 2 Normal level that has monsters for seeds 1 to 12 (the Act 1 run of decomp18 was re-run: **identical to `areamon-table.csv` on all 36 rows**, V run, a check that the image is deterministic). 33 Act 2 levels; 31 usable.
2. `tools/emu/dc28_rooms.py` reads the same level from game.exe's generator (`FUN_0061b730` room list): every room of every level is 8 x 8 tiles (words at room +0x3c and +0x40), so the real area is `rooms x 64` tiles. `dc28_perseed.py` joins units and tiles per seed.
3. `tools/emu/dc28_fit.py` (stdlib only, input `data28/areas-input.csv`) fits and reports; `internal/danger/areaunits.json` is the data file.

### 7.2 Findings (all on the 67 usable Normal areas of Acts 1 and 2)

| Finding | Number | Tag |
|---|---|---|
| Shipped estimate (`MonDen x groupMean x k`), median error | 94 percent (Act 1: 41, Act 2: 147); mazes 115, outdoors 147, presets 18 (Act 1 fit) | V run vs I |
| Tiles model: `units = tiles x exp(a [+ b ln MonDen])` | in sample 16, leave-one-area-out 17, leave-one-family-out 17; Act 1 fit predicting Act 2 only: 48 | V run data, I model |
| By type, leave-one-area-out | mazes (DrlgType 1) 13, outdoors (3) 19, presets (2) 54 | V run |
| Maze rate | `exp(-5.8175) x MonDen^0.4111` units per tile (0.041 at MonDen 600, 0.075 at 2,200) | I (fit) |
| Group size term | exponent of the mean group size 0.03 to 0.1 when added: no effect | I (fit) |
| Units follow tiles inside one level | across the 12 seeds: Tal Rasha's tombs correlation 0.78 to 0.96, Cave 1 0.92, Pit 1 0.81, Underground Pass 1 0.72 (levels whose room count varies) | V run (`data28/perseed-*.csv`) |
| Outdoors | tiles 5,200 to 6,300 in every area but Burial Grounds (1,920); MonDen 520 (Act 1) and 880 (Act 2) give the same 0.02 to 0.04 units per tile: MonDen does not matter | V run |
| Presets (Monastery, Cloisters, Tristram, the second caves, Catacombs 4) | units per tile 0.012 to 0.086, no model; Monastery Gate has 13 units on 1,152 tiles and Inner Cloister 8 on 360: their monsters come from the preset maps | V run, U (cause) |
| Levels with no population | `45` (a 16-room stub, 1 or 2 units in 7 of 12 seeds; not investigated, U) and `73` (Duriel's lair, the boss only) are excluded | V run |
| Per screen | game.exe 10 to 22 units per 280-tile screen in mazes, 5 to 11 outdoors; shipped: mazes 0.4 to 3 (Levels.txt 200 x 200), outdoors 7 to 32 | V run vs the generator |

What formula terms are missing: (1) the level's real area (rooms x 64) as a data column per area (not in the Levels.txt sizes; from the generator); (2) a per-type rate, with `MonDen` only for mazes and a power below 1; (3) for presets, the preset map's own monster count, which only the emulator (or the DS1 files) gives; (4) not needed: pack size, uniques and champions (champions are 0 in Normal Act 1 and 2; uniques and minions are in the units).

### 7.3 Data file and wiring

`internal/danger/areaunits.json` (`danger.areaunits/1`): per area `unitsMean/Min/Max`, `tiles`, `perScreen`, the fit parameters, the exclusions. It is **additive and unused**: a wiring card (Go) would read it as `Density.Units` and `Density.PerScreen` for a measured area and use `fit.*` with a tile count for an unmeasured one. The override file (`overrides.json`, cap of 5 per class) is not the place for 67 rows. Not measured: Nightmare and Hell (the Levels.txt `MonDen(N)` / `MonDen(H)` columns and the monster levels differ; Act 1's Normal and Nightmare dens are equal, Acts 3 to 5 were not run). The Normal Act 3 to 5 areas follow by the same recipe: `tools/emu/dc28_batch.sh OUT 1 12 4 <ids> 0`, `dc28_rooms.py`, `dc28_perseed.py`, `dc28_fit.py`.

Held-out reading, stated plainly: a fit made on Act 1 alone predicts Act 2 with a 48 percent median (the mazes of Act 2 have MonDen up to 2,200 against 1,056 in Act 1: the fit extrapolates). The 17 percent holds with both acts in the training set and with whole families held out; it does not prove Acts 3 to 5.

## 8. Drill recipe for the fork verbs (lab binary; NOT run)

Docs: `docs/forkdbg.md`, `docs/d2d-fork.md`, `docs/forkcheats.md`.

| id | Confirms | Steps |
|---|---|---|
| R3 (decomp26) | Q5, Q7, B1, B2, D1 | with the P04 claim on the binary: read a TP scroll with the real inventory click, `debug hit 20` at steps 2 and 3, repeat 20 times, with 1 and with 5 spawned zombies adjacent: expect mode 10 held, the portal at step 4, the scroll spent once. Log life per step: this gives `castSeconds + enterSeconds` and the damage taken, the numbers card 5.1 needs. |
| R8 | Q8 (the used skill during a scroll) | the same, with `debug hit` replaced by a block roll (a Paladin with a shield and `debug stat` block 75, spawn a skeleton archer): expect no mode 9 during the cast. A mode 9 means d2d refuses less than game.exe (D1 block). |
| R9 | D2 | a Paladin with Concentration at a high level and Blessed Hammer: 200 casts under `debug hit` at the cast's second tick; count cuts; game.exe expects `stat 164 percent` uninterrupted (read the stat with `debug stat`); d2d expects 0. |
| R10 | B4 | a shield Paladin (block 60), `debug spawn` a Fallen Shaman: 200 bolts: 0 blocked on game.exe, about 120 on the parent (decomp26 R6). With the zone margin: `survival` input for Black Marsh before and after card 5.2. |

## 9. Matrix-cell updates (proposed; not applied)

- `sorc-skill-matrix.md`, Town Portal row (scroll and tome): "not interruptible by a hit, block or knock-back (V run decomp28 Q5); only death, an Amazon's dodge or avoid, or the unit's own next attack or cast ends it". Every other Sorceress cast: "interrupted by a hit, unless Concentration state 0x2a rolls under stat 164 or the Concentrate state is on (V run, Q4)".
- `pal-skill-matrix.md`: Zeal, Smite, Concentrate rows: "refuse every mode but death while swung; Concentration aura gives a chance for the others (D2)".
- `docs/research/decomp26-trace.md` B1 and B4: "answered in decomp28; B4 is a bot defect in the power margin".
- `docs/survival/danger.md` line 9 (units per clear fit): "the fit of docs/survival/units-fit-2026-11-15.md replaces 'median error outdoor 20, preset 18, maze 55' (the 41 percent over Act 1 is 94 percent over Acts 1 and 2); PerScreen needs the real area".
- `docs/research/areamon-trace.md` D1: population counts per area for Act 2 now exist (`data28/areas-input.csv`).

## 10. Dead ends and limits

- The exposure of a cast (Q9) was not measured on game.exe; the two repo numbers disagree.
- `FUN_00553380` (the Evade branch, flag `0x200`) was not read; the stun state (21) path was not run.
- The real values of stat 164 per Concentration level (Skills.txt aura table) were not read; the oracle used 50.
- Presets, Nightmare, Hell and Acts 3 to 5 are not modelled; two levels (45, 73) give no population.
- Monster counts are units with experience over 0 (as `am_model.py`); champions are 0 in these areas.

## 11. Reproduce (the worker host; one core, nice 10)

```
# interrupt oracle (seconds)
$PY tools/emu/dc28_tpint.py game > game.tsv ; $PY tools/emu/dc28_sum.py        # reads /tmp/d28_game.tsv (edit the path)
# population and rooms
bash tools/emu/dc28_batch.sh OUT 1 12 4 41,42,... 0         # 12 seeds, 4 per call, about 2 min for Act 2
$PY tools/emu/dc28_rooms.py 1-12 2,3,...,74 0 > rooms.txt
$PY tools/emu/dc28_perseed.py OUT rooms.txt > perseed.csv
# fit (no emulator)
python3 tools/emu/dc28_fit.py report ; python3 tools/emu/dc28_fit.py json > internal/danger/areaunits.json
```
Helpers: `dc28_dis.py` (disassembly of a range), `dc28_callers.py` (call sites of a function), `dc28_units.py` (units per area).

## 12. Glossary

No change. Proposed line for the coordinator: "Unit per clear: a monster unit (plain, champion, unique or minion with experience over 0) that game.exe's room population places in one level of one game; 'tiles' is the number of 8 x 8 rooms times 64."
