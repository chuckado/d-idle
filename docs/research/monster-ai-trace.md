# Monster AI: what game.exe 1.14d does for the Act 1 and 2 monsters, vs d2d and our bot (agent monsterai, 2026-11-01)

> status: research, current. Question: how does game.exe run a woken monster's AI (target, approach, attack choice, flee, ranged, packs, difficulty), and where do d2d and the bot differ? Tags: V read = read in the game.exe disassembly (`disasm.py`); V run = game.exe's own code run under the emulator against d2d's real rules (counts below); D = 1.14d patched tables read with `mpq-cat` (docs/process/mpq-data-reading.md); S = an existing upstream proof or doc; I = inference; U = unverified. d2d = upstream f8aab0e8 (`components/rules/monsters.hpp`, `components/game/ai.cpp`). Classic and expansion tables as the bot extracts them. Offline only: no game, no slot, nothing pushed or posted. Prior traces this builds on: docs/research/mon-wake-trace.md (who notices whom: 35 subtiles, cadence), monster-attack-trace.md (the hit rolls), monster-mods-trace.md.

## Verdict

1. **Five Act 1 and 2 core thinks are exact in d2d**: Skeleton, Goatman, Quill Rat, Griswold and the Smith ran 3,000 cases each in game.exe's own code (two seeds), replayed through `rules::mon_think`: **0 differ**. The Fallen is exact too in the one regime that matters most (no leader command word: 1,500 cases, 0 differ; with a command word 2, which no Act 1 Fallen leader sends, 510 of 1,500 differ, I).
2. **The Act 1 hand ports that were never oracle-run are not all right.** Proven at the level of the think's output (D1 to D4): the Zombie, Brute, Blood Hawk, Corrupt Rogue and Corrupt Lancer thinks in game.exe set a **pace** (a speed bonus on the next run or walk, `FUN_005de190`) that d2d's port leaves out (649, 1,820, 1,686, 917 and 1,656 of 3,000 cases each), and the **Brute runs** in game.exe where d2d walks (D2). What the missing pace costs in speed depends on game.exe's run base speed, which was not read (U, section 5).
3. **Upstream's own oracle (`tools/emu/mon_ais.py`) already covers 91 think entries with 4,018 recorded game.exe cases** (S, counted from the committed corpus `tests/mon_ais*_cases.inc`); the Act 1 core (Skeleton, Zombie, Fallen, Fallen Shaman, Goatman, Quill Rat, Brute, Blood Hawk, Wraith, Corrupt Rogue, Skeleton Archer, Arach, Fetish, Andariel, Griswold, Smith) was not in it. This card adds a harness for them (`tools/emu/mai_ais.py`, 25 AIs, no upstream file edited) and ran it twice.
4. **No PR candidate**: the differential is new (not the existing harness), the effect of the missing pace (speed) is U, and no fork branch was made. Drill recipe in section 7.
5. **The bot kites and steps back as if monsters walk** (B1): the Zombie runs with a pace of 100 and the Brute runs, in game.exe.

## Q1. Where AI is decided (V read, S)

| Step | Address | What | Tag |
|---|---|---|---|
| Driver | FUN_005b1740 | stun, door, flee (FUN_005b10e0), the search (FUN_005b1650: type 1 FUN_005de890 search then stand/wander, type 2 search only, types 4 / 5 FUN_005de9d0), speech and leash (FUN_005b13e0), then the MonAI think | S + V read (mon-wake-trace section 1) |
| MonAI table | 0x73ca18, 16 bytes a row: +0 search type, +4 an init think (nine rows), +8 the think, +12 the other think | `tools/emu/mai_table.py` dumped all 148 MonAI.txt rows (D for the names, V run for the addresses; the 146 distinct names of monstats are rows of it, `None` and `Idle` aside) | V run |
| Think params block | [0] AI control (scratch words +0x14 / +0x18 / +0x1c), [2] target, [5] distance, [6] in melee, [7] MonStats row; aip1..5 at MonStats +0x56 + 6k + 2 x difficulty | upstream scene, re-read in the thinks below | S + V read |
| Acts | FUN_005ddf90 mode (4 A1, 5 A2, 9 S2 ...), FUN_005dead0 / FUN_005de000 skill (+ sequence), FUN_005dec80 walk, FUN_005ded40 / FUN_005defb0 run, FUN_005def80 approach, FUN_005defe0 back off, FUN_005de6d0 keep distance, FUN_005df7d0 circle, FUN_005de200 wander, FUN_005de080 stand, FUN_005de190 **pace** | hooked by the harness; each think ends in exactly one | V run |
| Aggro radius | FUN_005dd7f0 | 35 subtiles for every row (no `aidist` set in 734 of 735 rows: D), no sight needed outdoors; being hit never sets aware | S + V run (mon-wake-trace: 141,088 / 0 differ) |
| Think period | FUN_005a73e0 | `aidel` frames: 15 Normal, 14 Nightmare, 13 Hell for every Act 1 core type (D, table below); 0 -> 15 | V run (55,268 / 0 differ, mon-wake-trace) |

Group behaviour: the Fallen leader (`dist < 15`, aip1 %) and the Fallen Shaman (aip1 %) send command kind 1 to their group (FUN_005f04b2 / FUN_005f1558 store 1 before FUN_0058f730); a minion acts on it only inside its own think, after its own target search (V read; mon-wake-trace). A Fallen **flees** a dying Fallen: the think's first loop looks at the group's four ids (+0x38), any unit in DT within 15 subtiles (FUN_005dc530 < 0xf) sets state 1, clears the command (FUN_0058ed10), **sets pace 50** (FUN_005de190(50, 0) at `0x5f03a8`..`0x5f03b0`) and backs off 12 (FUN_005defe0, 0xc) (V read, `0x5f02c0`..`0x5f03df`); d2d: `dying` (monsters.hpp:1525-1528) sets the state, clears the command and backs off but sets no pace (D11), not exercised by the harness (limit L1).

## Q2. Per-AI rules of the Act 1 and 2 killers (V read of the think, D for the numbers; aip labels from MonAI.txt)

| Id | AI | Vel / Run | aip1..aip5 Normal | aip Hell | aidel N / NM / H |
|---|---|---|---|---|---|
| skeleton1 | Skeleton | 3 / 3 | approach? 60; stall time 15; attack? 75; att1/att2? 75; aip5 - | 70 10 85 75 - | 15 / 14 / 13 |
| zombie1 | Zombie | 1 / 3 | approach? 30; aware dist 10; att1/att2? -; aip4 20; aip5 - | 50 20 - 30 - | 15 / 14 / 13 |
| fallen1 | Fallen | 5 / 5 | cmd:attack? 30; attack? 10; att1/att2? 50; aip4 20; aip5 - | 50 20 70 30 - | 15 / 14 / 13 |
| fallenshaman1 | FallenShaman | 4 / 4 | rez?/cmd? 45; shoot? 60; melee?/circle? 100; rez dist 24; shoot dist 15 | 55 70 80 26 15 | 15 / 14 / 13 |
| goatman1 | Goatman | 5 / 5 | approach? 75; stall time 10; attack? 80; aip4 -; aip5 - | 80 9 80 - - | 15 / 14 / 13 |
| quillrat1 | QuillRat | 3 / 3 | activate dist 10; shoot? 35; walk dist -; aip4 2; aip5 - | 20 50 - 2 - | 15 / 14 / 13 |
| brute1 | Brute | 3 / 3 | circle? -; attack? 0; att1/att2? 100; aip4 45; aip5 - | - 0 100 35 - | 15 / 14 / 13 |
| foulcrow1 | BloodHawk | 4 / 4 | charge? 30; wander? 90; attack? 5; run vel 50; charge vel 100 | 65 90 10 70 100 | 15 / 14 / 13 |
| wraith1 | Wraith | 5 / 5 | approach? 50; stall time 12; attack? 70; aip4 -; aip5 - | 60 8 80 - - | 15 / 14 / 13 |
| corruptrogue1 | CorruptRogue | 5 / 5 | approach? 60; stall time 15; attack? 75; run vel 100; run? 20 | 80 7 85 100 75 | 15 / 14 / 13 |
| sk_archer1 | SkeletonBow | 4 / 4 | shoot? 75; stall time 15; approach? 50; walk steps 5; tgt dist 6 | 85 13 50 5 6 | 15 / 14 / 13 |
| cr_archer1 | CorruptArcher | 4 / 4 | approach? 60; shoot? 70; stall time 14; run? 20; always run dist 20 | 70 88 9 40 20 | 15 / 14 / 13 |
| andariel | Andariel | 8 / 8 | spray in melee? 30; stall? 10; fire or engage? 30; spray from dist? 50; aip5 - | 35 6 34 60 - | 15 / 11 / 9 |
| duriel | Duriel | 10 / 15 | holy freeze level 5; smite? 33; jab? 50; att2? -; charge? - | 6 33 50 - - | 15 / 15 / 15 |

Reading the table (D): the Zombie's aip2 "aware dist" is 10 (Hell 20); the Fallen's aip2 is 10 (Hell 20): it walks in only from that distance unless hit or rallied (S, mon-wake-trace); the Fallen Shaman raises a corpse at aip1 % within aip4 = 24 subtiles and shoots (Skill2, `shafire`) inside aip5 = 15 at aip2 %; the Quill Rat shoots spikes (A2) at aip2 = 35 % once the player is inside aip1 = 10 or after a hit, else wanders; the Corrupt and Skeleton Archers use the second-target search (FUN_005ddc30) and shoot at aip2 / aip1 %. Hell raises the notice and attack chances and shortens the stall (Skeleton aip1 60 to 70, Fallen aip1 30 to 50, Fallen Shaman aip2 60 to 70).

Difficulty differences (D, V run for the cadence): `aidel` 15 / 14 / 13 (Andariel 15 / 11 / 9); the aips above; the thinks also read the difficulty directly in two places that d2d's code names (S, d2d comments, not isolated in the runs): the Corrupt Rogue's run-in range `20 - 3 x difficulty` (monsters.hpp:1453) and Blood Raven's zombie cap `8 + 2 x difficulty`.

## Q3. What a woken monster does, per AI (this card's differential, section 4)

- **Skeleton, Goatman, Quill Rat, Griswold, Smith: same as d2d** (3,000 cases each, 0 differ, V run; Quill Rat and Fetish scenes carry no leader command because d2d's ports never receive one, a documented `ponytail` in monsters.hpp). Skeleton: walk in at aip1 %, in melee aip3 % (aip4 % A1, else A2), else stand aip2. Goatman and Wraith: aip1 % close in, in melee aip3 % A1, else stand aip2.
- **Zombie** (`FUN_005efe20`): in melee A1 / A2 (aip4 %); a target it was hit by, or one inside aip2 at aip1 %, or Burial Grounds (level 0x11, FUN_005eff0e..0x5eff17): **FUN_005de190(pace 100, 0) then run**; else wander 3. d2d runs without the pace (D1).
- **Brute** (`FUN_005efb80`): not in melee: pace = `100 - clamp(life %, 40, 100)` (0 to 60) at `0x5efcbf`..`0x5efccd`, then **run** (`FUN_005ded40`, 7); in melee A1 / A2, circle 4, stand 15. d2d walks without a pace (D2).
- **Blood Hawk**: charge (state 1, then a swing); the game's think sets paces of 0, 1 or 100 (V run; MonAI.txt names aip4 "run vel" and aip5 "charge vel", 50 / 100 in Normal, I that they are the source); d2d returns the same acts with pace 0 and different step counts (D3).
- **Fallen** (`FUN_005f02c0`): returns at once in DT / DD; flees a dying kin (above); a leader command of kind 1 charges (walk in; in melee aip3 % then A1 / A2 by aip4); otherwise rally (aip1 %, leader within 15), walk in from aip2, else 30 % wander 3 / stand 10. d2d equal for no command and kind 1.
- **Fallen Shaman** (`FUN_005f1440`): in melee aip3 % A1; aip1 % rally; if aware and not state 4, the corpse search `FUN_005f1380` (Fallen class 0x13 or 0x3a, DD, within aip4 squared) and aip1 % raise (Skill1, `FUN_005de000`); inside aip5 aip2 % shoot (Skill2); else aip3 % circle 3, else stand 10 (V read). d2d's `ai.cpp` feeds the corpse as "within 80 subtiles" for a unique shaman instead of aip4 squared (ai.cpp:880-886, I; U which rows).
- **Corrupt Rogue / Lancer, Skeleton Archer, Wraith**: run in past `20 - 3 x difficulty` (Rogue), aip5 (Lancer) with a pace; archers shoot or keep range with FUN_005de6d0.
- **Andariel** (`FUN_005f5830`, boss), **Duriel** (`FUN_005f67b0`, covered by upstream's corpus: 48 cases, S): not decomposed here beyond the table (Andariel has its own d2d function `andariel_think`, monsters.hpp:1056, not reachable through `mon_think`; L2).

## 4. The differential (V run)

Harness (all in this repo, scratch d2d tree on the worker host): `tools/emu/mai_ais.py` imports upstream `mon_ais.py` (scene, hooks, case format) and adds the think addresses from the MonAI table, a scene hook for the Fallen Shaman's corpse search and AI flags (`FUN_005a0180`, callback `FUN_005f1380`), a zero block for the Lancer's `params[8]`, the Fallen's mode (NU) and an empty room scan; `mai_patch.py` turns the replay block of upstream's `tests/test_monsters.cpp` into a counter over the new cases (rally as an effect, a back-off the game did not make flagged instead of asserting). Each case: a random scene (difficulty, frame, distance, in melee, got hit, life %, aips, skills, spots, seeds) run once in game.exe's own think (hooks outside the claim: the acts, the searches, FUN_005dd2b0 got-hit, FUN_00621f20 life, FUN_005ddc30 the second target; the second target is set to the primary at the same distance, as `ai.cpp` feeds it), the act, pace, effects, AI control and seed after written as a case line; d2d's `rules::mon_think` replays the line. Compared: act, n, x, y, unit, wait, event, pace, steps, the three control words, the seed pair, the effects and the back-offs asked.

Run A: seed 20261101, 1,500 cases per AI, commands as the scene draws them (37,500 cases). Run B: seed 7, no leader commands (37,500 cases). Per AI, A + B = 3,000 cases:

| AI | cases | differ | family (V run unless marked) |
|---|---|---|---|
| Skeleton | 3,000 | 0 | |
| Goatman | 3,000 | 0 | |
| Quill Rat | 3,000 | 0 | no commands in the scene (ponytail) |
| Griswold | 3,000 | 0 | |
| Smith | 3,000 | 0 | |
| Fallen | 3,000 | 510 | all with command word 2 (A only); run B (no command) 1,500 / 0 |
| Zombie | 3,000 | 649 | every run act: pace 100 vs 0 (D1) |
| Brute | 3,000 | 1,820 | every run act: game runs, d2d walks; pace 0..60 vs 0 (D2) |
| Blood Hawk | 3,000 | 1,686 | pace and steps (D3) |
| Corrupt Rogue | 3,000 | 917 | run n 3 vs 0, pace 1 / 100 vs 0 (D4) |
| Corrupt Lancer | 3,000 | 1,656 | run pace 100 vs 0 (D4) |
| Arach | 3,000 | 888 | run vs walk, none vs idle (D5) |
| Fetish | 3,000 | 182 | pace 50 vs 0; a few acts (D6) |
| Wraith | 3,000 | 494 | `keep` y field (D7) |
| Skeleton Archer | 3,000 | 1,482 | unit label (second vs primary target, a scene label) 1,336, `keep` y 146 (D7) |
| Fallen Shaman | 3,000 | 176 | skill slot n, unit label, `seq` vs `skill` for the raise (U, section 6) |
| Gargoyle Trap | 3,000 | 124 | nest-like init state; U |
| Foul Crow Nest | 3,000 | 1,034 | `gone` vs `die` (init state of the scene, U) |
| Blood Raven | 3,000 | 2,162 | second search, home / around acts; scene states (U) |
| Andariel, Dark Wanderer, Elemental Beast, Flying Scimitar, Quill Mother, Town Rogue | 3,000 each | 3,000 each | d2d `mon_think` returns `untraced` (no port in that function; Andariel has `andariel_think`): not a finding about the rule, L2 |

Totals: run A 37,500 cases, 16,120 differ; run B 37,500, 15,660 (9,000 cases in each run are the six untraced AIs). Baseline: upstream's unmodified `test_monsters` on f8aab0e8 passes (rc 0, V run), replaying its 4,018 committed cases. The test bites: the same replay flags every Zombie run in 649 cases, so a pace-aware port would show 0.

## 5. Deviations, d2d vs game.exe

| # | Rule in game.exe | d2d | Tag | Differential |
|---|---|---|---|---|
| D1 | Zombie: `FUN_005de190(100, 0)` before the run (`0x5eff19`..`0x5eff21`) | `return { MonAct::run }`, no pace (monsters.hpp:1401) | think output V run + V read (649 / 3,000, every run act); speed effect U (what the pace does to the run speed: ai.cpp:2692-2694 adds it to the Run column; game.exe's run base is unread) | yes: `mai_ais.py Zombie`, the run act's pace |
| D2 | Brute: not in melee, pace `100 - clamp(life %, 40, 100)` then **run** (`0x5efca3`..`0x5efcdc`) | `walk` (monsters.hpp:1435-1442) | V run + V read (1,820 / 3,000); effect U | yes (same) |
| D3 | Blood Hawk: pace from aip4 / aip5 on its charge and run, steps | pace 0 | V run (1,686 / 3,000); root of the steps field unread (I) | yes |
| D4 | Corrupt Rogue: run with n 3 and pace 1 / 100; Corrupt Lancer: run with pace 100 | `run` n 0, pace 0 | V run (917, 1,656 / 3,000) | yes |
| D5 | Arach (Act 1 Catacombs 2): runs where d2d walks | walk | V run (888 / 3,000); root unread | yes |
| D6 | Fetish: pace 50 on its walk; a few acts differ with the leader state | pace 0 | V run (182 / 3,000) | yes |
| D7 | Wraith / Skeleton Archer `keep` (FUN_005de6d0): y field 2 in every case | y 0 | V run (494, 146); the argument behind y unread: U whether it is a flag that changes the walk | yes |
| D8 | Fallen with a leader command word that is not 1: another branch (`0x5f068b`) | ignores any word but 1 | V run (510 / 1,500 in run A); no Act 1 Fallen sender sets another word (the two stores at `0x5f04b2` and `0x5f1558` are 1): I, unreachable | yes (command word in the scene) |
| D9 | Fallen Shaman corpse: callback FUN_005f1380 within aip4 squared (and only when aware and not state 4) | `ai.cpp` 880-886: a unique shaman takes any Fallen / Shaman corpse within 80 subtiles | V read (callback), V-d2d read; reach U (which unique shamans exist) | no: the d2d side is in `ai.cpp`, not in `mon_think`; build from the Fallen Shaman scene with the corpse distance as an input and `ai.cpp`'s feed (needs a header-only copy of the lines 860-900 loop) |
| D10 | RN base speed and what stat 0x43 does to it | `ai.cpp:2693-2694` | U | no existing one; build: run `FUN_005a7c20` for a monster in RN with stat 0x43 = 100 and read the path speed (needs the unit, stats, path objects), oracle = `cells_per_sec(Run) x (100 + pace) / 100` |
| D11 | Fallen flees a dying kin with pace 50 (`0x5f03a8`) | `dying` branch: state 1, command 0, away 12, no pace (monsters.hpp:1525-1528) | V read only; the harness's room scan answers "none" so it was not run | yes, to build: hook FUN_00619790 to return one DT unit within 15 and FUN_00552f60 by id, with `dying = true` in the replay |

Not listed as findings (they stay U, section 6): FallenShaman residual, Gargoyle Trap, Foul Crow Nest, Blood Raven.

## 6. Limits and dead ends

- L1: `dying` (a Fallen flees a dying kin), the leader's rally to the group, the Fallen Shaman's corpse find (a boolean here), and the second-target search (set equal to the primary) are inputs the case line carries as scene values; their d2d feeds (`ai.cpp:860-910`) are not in the differential. The room scan `FUN_00619790` is hooked to "no unit".
- L2: six AIs have no `mon_think` port (3,000 `untraced` each); Andariel's `andariel_think` and `countess_think`-style functions need their own case format.
- L3: 25 of the 148 MonAI rows were run (this card) and 91 by upstream's corpus; the rest (Act 3 to 5, town NPCs, traps, pets) are not Act 1 and 2 killers or are not thinks (`None`, `Idle`, `Npc`).
- L4: damage, to-hit and speed are not in the think: only the choice of act and its pace. Boss packs (Andariel's, Duriel's minions) not decomposed.
- L5: `CorruptLancer` reads a byte at `params[8] + 0xe` (a pointer); unmodelled, a zero block; d2d's port has no input for it (U what it is).
- Dead end: the first harness run had every Fallen think returning `none` (the unit's mode word 0 returns at once at `0x5f02e5`); fixed by giving it NU. A second dead end: `FUN_005a0180(unit, mask)` is a flag test on monster data +0x16, not a mode test; it decides the shaman's search branch.

## 7. Bot findings (described, not changed)

- B1: the stepback and kite rules assume a monster that can be walked away from. `internal/combat/fight.go:628-632` steps back from one melee monster within `stepBackDist` 3 below 60 % life, and `KiteDistance` 10 / `KiteRetreat` 12 (`internal/env/tuning_combat.go:85`) retreat on foot; the Zombie runs with a pace of 100 and the Brute runs with up to 60 (V run for the act and pace), so a retreat does not break contact with either. The Skeleton, Goatman and Fallen only walk (Fallen Vel = Run = 5, D), so the retreat works there; a fleeing Fallen runs at pace 50 (D11). Speed effect U (D10).
- B2: the Fallen packs flee a dying kin: with the pack in a screen the bot sees monsters walking away and back; `internal/combat` has no notion of it (comment at fight.go:621-624 documents the pack case only as a kiting trap). Whether the bot should wait is U.
- B3: `ReachSec` and the approach budget (`fight.go:487`, `MaxApproaches` 12) treat a target that does not come as unreachable; a Fallen stands still until aip2 = 10 subtiles unless hit or rallied (S, mon-wake-trace), so the bot's "target does not approach" is the normal Fallen behaviour, not a failure.

## 8. PR candidates

None. The deviations D1 to D8 are proven on a new harness, not the existing differential, their effect (speed) is U (D10), and no fork branch was made. If D10 is closed and the Zombie / Brute speed matters, the fix shape is: `*input.pace = 100` before the Zombie run (monsters.hpp:1401), `*input.pace = 100 - clamp(life_pct, 40, 100)` and `MonAct::run` for the Brute (monsters.hpp:1436), with the replay as the proof (mismatches before: 649 and 1,820 of 3,000; after: 0 expected, not built).

## 9. Drill recipe for the fork verbs (docs/forkdbg.md, docs/forkcheats.md; not run)

1. Zombie speed (D1, D10): `debug spawn zombie1 1 near`, `debug goto` a field level, read `monsters` positions every tick with `step 1` while the zombie is in RN (walks at Vel 1, runs at Run 3). Cells per second from consecutive positions; compare with `cells_per_sec(Run) x (100 + pace) / 100`: d2d's value is what the bot sees; game.exe's is the V run pace 100 on a base of Run (U which).
2. Brute: same with `brute1`; `debug mon set <id> life <n> maxlife <n>` to 50 % and 100 %: d2d walks both times; game.exe runs, faster when hurt.
3. Blood Hawk (`foulcrow1`), Corrupt Rogue (`corruptrogue1`): charge speed (aip4 / aip5 "run vel" / "charge vel" 50 / 100).
4. Fallen Shaman corpse reach (D9): `debug spawn fallenshaman1 1 unique`, `debug spawn fallen1 4`, `debug kill` one at 30, 60, 78 subtiles and read whether the shaman raises it (`monsters` mode); game.exe's rule: only within aip4 = 24.

## 10. Matrix and ledger proposals

- docs/coverage/ledger.tsv (proposed text, applied in this card): Skeleton, Goatman, QuillRat, Griswold, Smith TESTED-0-DIFF (3,000 each, aspect: the think's act, pace, effects and seed); Fallen TESTED-0-DIFF for no-command (1,500, aspect noted); Zombie, Brute, BloodHawk, CorruptRogue, CorruptLancer TESTED-DEVIATION (D1 to D4).
- docs/research/d2d-fidelity-riskmap.md rank 5 `mon_wake`: wake and cadence stay clean; add "think proper: D1 to D4 open".
- docs/coverage/gaps.tsv row 2 (monster.ai): stays (135 of 146 rows still untested after this card; the Act 3 to 5 AIs are upstream-corpus S only).

## 11. Reproduce (the worker host)

```sh
# d2d scratch tree: git worktree of upstream f8aab0e8 at ~/d2d-lab/pr/monsterai-src; game.exe read-only copy under ~/d2d-lab/pr/monsterai-out/data/bin
tools/emu/mai_build.sh                 # test_monsters, 2 jobs, about 1.5 min
tools/emu/mai_all.sh                   # runs A and B, about 6 min; prints `mai AI cases differ` lines (the table above)
python3 tools/emu/mai_table.py <monai.txt>   # with the upstream tools/emu on PYTHONPATH: the MonAI table
python3 tools/emu/mai_aitable.py <tables> <table dump> <traced names> <upstream inc dir> <counts>   # section 12
```

## 12. Per-AI table: all 146 distinct monstats AIs (row counts match docs/coverage/denominators.tsv: 146 AIs, 734 monstats rows)

"d2d mon_think port" is membership in `traced_ai` (monsters.hpp:1296); "upstream corpus cases" are recorded game.exe cases in `tests/mon_ais*_cases.inc` (other-thinks excluded); "Act 1-2 placed" are monstats Ids named in a levels.txt `mon` / `nmon` / `umon` column of an Act 1 or 2 level (superuniques and bosses are placed elsewhere and show `-`).

| AI | monstats rows | Act 1-2 placed (Id) | search type | think (+8) | aidel N/NM/H | d2d mon_think port | upstream corpus cases | this run cases / differ |
|---|---|---|---|---|---|---|---|---|
| 7TIllusion | 1 | - | 1 | 5ea080 | 15/15/15 | no | - | - |
| AbyssKnight | 1 | - | 1 | 5fab80 | 15/14/13 | yes | 30 | - |
| Ancient | 3 | - | 2 | 5ef1a0 | 15/13/11 | yes | 75 | - |
| AncientStatue | 3 | - | 0 | 5eeaa0 | 15/15/15 | no | - | - |
| Andariel | 2 | - | 1 | 5f5830 | 15/11/9 | no | - | 3000 / 3000 |
| Arach | 7 | arach1 | 1 | 5f4510 | 15/14/13 | yes | - | 3000 / 888 |
| ArcaneTower | 1 | - | 1 | 5e0f60 | 15/15/15 | yes | 18 | - |
| AssassinSentry | 4 | - | 0 | 5ea3d0 | 15/15/15 | no | - | - |
| BaalCrab | 1 | - | 0 | 5fcfe0 | 15/13/12 | yes | 180 | - |
| BaalCrabClone | 1 | - | 0 | 5fd210 | 15/14/12 | yes | 162 | - |
| BaalMinion | 3 | - | 1 | 5ef910 | 15/14/13 | yes | 30 | - |
| BaalTaunt | 1 | - | 1 | 5ef710 | 15/15/15 | yes | - | - |
| BaalTentacle | 5 | - | 1 | 5ef820 | 15/15/15 | yes | 18 | - |
| BaalThrone | 1 | - | 2 | 5ef320 | 15/15/15 | yes | - | - |
| BaalToStairs | 1 | - | 1 | 5ef620 | 15/15/15 | yes | - | - |
| Baboon | 8 | baboon1 | 1 | 5f0cd0 | 15/14/13 | yes | 114 | - |
| BatDemon | 7 | batdemon1 | 1 | 5f5040 | 15/14/13 | yes | 89 | - |
| Bighead | 10 | bighead1, bighead2, bighead3, bighead4 | 1 | 5eff50 | 15/14/13 | yes | 48 | - |
| BladeCreeper | 1 | - | 0 | 5ea540 | 15/15/15 | no | - | - |
| BloodHawk | 8 | foulcrow1, foulcrow2, foulcrow3 | 1 | 5f00e0 | 15/14/13 | yes | - | 3000 / 1686 |
| BloodLord | 7 | - | 1 | 5e36f0 | 15/14/13 | yes | 24 | - |
| BloodRaven | 1 | - | 1 | 5e6320 | 15/13/10 | yes | - | 3000 / 2162 |
| BoneWall | 1 | - | 0 | 5e0400 | 15/15/15 | no | - | - |
| Brute | 9 | brute1, brute2, brute3, brute4 | 1 | 5efb80 | 15/14/13 | yes | - | 3000 / 1820 |
| Buffy | 1 | - | 0 | 5e7f50 | 15/15/15 | no | - | - |
| Catapult | 4 | - | 1 | 5e34c0 | 15/15/15 | yes | 12 | - |
| CatapultSpotter | 4 | - | 1 | 5ee040 | 15/15/15 | yes | 18 | - |
| ClawViper | 5 | clawviper2, clawviper3 | 1 | 5f1b60 | 15/14/13 | yes | 66 | - |
| ClawViperEx | 5 | - | 1 | 5f1de0 | 15/14/13 | yes | 66 | - |
| CorruptArcher | 8 | cr_archer1, cr_archer2, cr_archer3, cr_archer4 | 1 | 5f5a20 | 15/14/13 | yes | 88 | - |
| CorruptLancer | 9 | cr_lancer1, cr_lancer2, cr_lancer3 | 1 | 5f5d50 | 15/14/13 | yes | - | 3000 / 1656 |
| CorruptRogue | 6 | corruptrogue1, corruptrogue2, corruptrogue3, corruptrogue4 | 1 | 5f0b00 | 15/14/13 | yes | - | 3000 / 917 |
| CycleOfLife | 2 | - | 2 | 5ec8c0 | 15/15/15 | no | - | - |
| DarkWanderer | 1 | - | 1 | 5ea130 | 15/15/15 | yes | - | 3000 / 3000 |
| DeathMauler | 6 | - | 1 | 5ee260 | 15/15/15 | yes | 24 | - |
| DeathSentry | 1 | - | 0 | 5ea980 | 15/15/15 | no | - | - |
| DesertTurret | 4 | - | 1 | 5e0980 | 15/15/15 | yes | 12 | - |
| Diablo | 2 | - | 1 | 5e9170 | 15/13/12 | yes | 108 | - |
| DoomKnight | 3 | - | 1 | 5faa90 | 15/14/13 | yes | 18 | - |
| DruidBear | 1 | - | 0 | 5ed730 | 15/15/15 | no | - | - |
| DruidWolf | 2 | - | 0 | 5ed710 | 15/15/15 | no | - | - |
| Duriel | 2 | - | 1 | 5f67b0 | 15/15/15 | yes | 48 | - |
| ElementalBeast | 4 | - | 1 | 5f6b70 | 15/15/15 | no | - | 3000 / 3000 |
| EvilHole | 6 | - | 1 | 5fb410 | 15/15/15 | no | - | - |
| Fallen | 8 | fallen1, fallen2, fallen3, fallen4 | 1 | 5f02c0 | 15/14/13 | yes | - | 3000 / 510 |
| FallenShaman | 8 | fallenshaman1, fallenshaman2, fallenshaman3, fallenshaman4 | 1 | 5f1440 | 15/14/13 | yes | - | 3000 / 176 |
| Fetish | 16 | fetish1 | 1 | 5f53e0 | 15/14/13 | yes | - | 3000 / 182 |
| FetishBlowgun | 8 | - | 1 | 5e1250 | 15/14/13 | yes | 120 | - |
| FetishShaman | 8 | - | 1 | 5f9a80 | 15/14/13 | yes | 54 | - |
| FingerMage | 5 | - | 1 | 5fa380 | 15/14/13 | yes | 42 | - |
| FlyingScimitar | 1 | - | 1 | 5f6ca0 | 15/13/10 | no | - | 3000 / 3000 |
| FoulCrowNest | 4 | crownest1, crownest2, crownest3 | 1 | 5f6650 | 15/15/15 | yes | - | 3000 / 1034 |
| FrogDemon | 3 | - | 5 | 5f8260 | 15/14/13 | yes | 54 | - |
| FrozenHorror | 5 | - | 1 | 5e3530 | 15/14/13 | yes | 42 | - |
| GargoyleTrap | 1 | - | 1 | 5f9490 | 15/15/15 | yes | - | 3000 / 124 |
| GenericSpawner | 1 | - | 1 | 5e61b0 | 15/15/15 | yes | 18 | - |
| Goatman | 10 | goatman1, goatman2, goatman3, goatman4, goatman5 | 1 | 5f12a0 | 15/14/13 | yes | - | 3000 / 0 |
| GoodNpcRanged | 1 | - | 0 | 5e7ac0 | 15/15/15 | no | - | - |
| GreaterMummy | 10 | unraveler1, unraveler2, unraveler3 | 1 | 5f2b10 | 15/14/13 | yes | 42 | - |
| Griswold | 1 | - | 1 | 5e5ac0 | 15/13/10 | yes | - | 3000 / 0 |
| HellMeteor | 1 | - | 1 | 5f56d0 | 15/15/15 | yes | 12 | - |
| HighPriest | 4 | - | 1 | 5e0490 | 15/13/12 | yes | 54 | - |
| Hireable | 5 | - | 0 | 5e52d0 | 1/1/1 | no | - | - |
| Hydra | 3 | - | 2 | 5e9e60 | 15/15/15 | yes | 18 | - |
| Idle | 61 | - | 0 | 5b0cd0 | 15/14/13 | no | - | - |
| Imp | 8 | - | 1 | 5e2ff0 | 15/14/13 | yes | 63 | - |
| InvisoPet | 1 | - | 0 | 5ea7a0 | 15/15/15 | no | - | - |
| InvisoSpawner | 1 | - | 1 | 5e0160 | 15/15/15 | no | - | - |
| Izual | 1 | - | 1 | 5f89b0 | 15/12/8 | yes | 59 | - |
| JarJar | 1 | - | 0 | 5e7590 | 15/15/15 | no | - | - |
| MaggotEgg | 6 | - | 1 | 5f6530 | 15/15/15 | yes | 18 | - |
| MaggotLarva | 6 | - | 1 | 5f6220 | 15/14/13 | yes | 18 | - |
| Megademon | 7 | - | 1 | 5e0c80 | 15/13/11 | yes | 60 | - |
| Mephisto | 1 | - | 1 | 5f78b0 | 15/10/6 | yes | 90 | - |
| Minion | 11 | - | 1 | 5e1b60 | 15/14/13 | yes | 48 | - |
| MinionSpawner | 9 | - | 1 | 5e2bd0 | 15/15/15 | yes | 18 | - |
| Mosquito | 4 | - | 1 | 5f3730 | 15/14/13 | yes | 36 | - |
| MosquitoNest | 4 | - | 0 | 5e0260 | 15/15/15 | no | - | - |
| Mummy | 6 | mummy1, mummy2, mummy3, mummy4 | 1 | 5f2850 | 15/14/13 | yes | 36 | - |
| Navi | 1 | - | 0 | 5e7e20 | 15/15/15 | no | - | - |
| NecroPet | 9 | - | 0 | 5e4cf0 | 15/15/15 | no | - | - |
| Nihlathak | 1 | - | 2 | 5ee5d0 | 15/15/15 | yes | 99 | - |
| Npc | 35 | - | 0 | 5e7130 | 15/15/15 | no | - | - |
| NpcBarb | 2 | - | 2 | 5edc50 | 15/15/15 | yes | 36 | - |
| NpcOutOfTown | 2 | - | 0 | 5e7880 | 15/15/15 | no | - | - |
| NpcStationary | 4 | - | 0 | 5e73a0 | 15/15/15 | no | - | - |
| OblivionKnight | 4 | - | 1 | 5faf00 | 15/14/13 | yes | 60 | - |
| Overseer | 6 | - | 1 | 5e27a0 | 15/14/13 | yes | 42 | - |
| PantherJavelin | 9 | slinger1, slinger2, slinger4, slinger5, slinger6 | 1 | 5e1080 | 15/14/13 | yes | 30 | - |
| PantherWoman | 6 | pantherwoman1, pantherwoman2, pantherwoman3, pantherwoman4 | 1 | 5f22b0 | 15/14/13 | yes | 30 | - |
| PinHead | 6 | blunderbore1, blunderbore2 | 1 | 5f6340 | 15/14/13 | yes | 24 | - |
| PutridDefiler | 5 | - | 1 | 5efa90 | 15/15/15 | yes | 30 | - |
| QuillMother | 5 | - | 1 | 5fb2a0 | 15/15/15 | no | - | 3000 / 3000 |
| QuillRat | 8 | quillrat1, quillrat2, quillrat3, quillrat4 | 1 | 5f1140 | 15/14/13 | yes | - | 3000 / 0 |
| Raven | 1 | - | 2 | 5ecc10 | 15/15/15 | no | - | - |
| ReanimatedHorde | 6 | - | 1 | 5e1540 | 15/14/13 | yes | 30 | - |
| Regurgitator | 4 | - | 1 | 5fa710 | 15/14/13 | yes | 60 | - |
| SandLeaper | 7 | sandleaper1, sandleaper2, sandleaper3 | 1 | 5f20a0 | 15/14/13 | yes | 36 | - |
| SandMaggot | 7 | sandmaggot1, sandmaggot2, sandmaggot3 | 4 | 5f1800 | 15/14/13 | yes | 30 | - |
| SandMaggotQueen | 5 | - | 0 | 5f9cf0 | 15/15/15 | yes | 24 | - |
| SandRaider | 10 | sandraider1, sandraider2, sandraider3 | 1 | 5f0700 | 15/14/13 | yes | 133 | - |
| Sarcophagus | 1 | - | 1 | 5f6a10 | 15/15/15 | yes | 18 | - |
| Scarab | 8 | scarab1, scarab2, scarab3, scarab4 | 1 | 5f2540 | 15/14/13 | yes | 54 | - |
| ShadowMaster | 1 | - | 2 | 5eb970 | 15/15/15 | no | - | - |
| ShadowWarrior | 1 | - | 2 | 5eafa0 | 15/15/15 | no | - | - |
| SiegeBeast | 5 | - | 1 | 5e1900 | 15/14/13 | yes | 66 | - |
| SiegeTower | 1 | - | 1 | 5e1860 | 15/15/15 | yes | 12 | - |
| Skeleton | 9 | hellbovine, skeleton1, skeleton2, skeleton3, skeleton4, skeleton5 | 1 | 5efcf0 | 15/14/13 | yes | - | 3000 / 0 |
| SkeletonBow | 11 | sk_archer1, sk_archer2, sk_archer3, sk_archer4, sk_archer5 | 1 | 5f6070 | 15/14/13 | yes | - | 3000 / 1482 |
| SkeletonMage | 27 | skmage_cold4, skmage_fire1, skmage_fire2, skmage_fire3, skmage_fire4, skmage_ltng1... | 1 | 5f96c0 | 15/15/15 | yes | 36 | - |
| Smith | 2 | - | 1 | 5e3890 | 15/13/12 | yes | - | 3000 / 0 |
| Spirit | 1 | - | 1 | 5e3840 | 15/15/15 | yes | 12 | - |
| Succubus | 8 | - | 1 | 5e1e00 | 15/14/13 | yes | 24 | - |
| SuccubusWitch | 9 | - | 1 | 5e2120 | 15/14/13 | yes | 81 | - |
| SuicideMinion | 8 | - | 1 | 5e1d30 | 15/14/13 | yes | 18 | - |
| Summoner | 1 | - | 1 | 5f85c0 | 15/13/10 | yes | 48 | - |
| Swarm | 5 | swarm1, swarm2 | 1 | 5f2460 | 15/14/13 | yes | 18 | - |
| Tentacle | 3 | - | 2 | 5f8f80 | 15/15/15 | yes | 36 | - |
| TentacleHead | 3 | - | 2 | 5f9270 | 15/15/15 | yes | 30 | - |
| ThornHulk | 5 | - | 1 | 5f4850 | 15/14/13 | yes | 42 | - |
| Totem | 3 | - | 0 | 5ed9e0 | 15/15/15 | no | - | - |
| TownRogue | 1 | - | 1 | 5e7d60 | 15/15/15 | no | - | 3000 / 3000 |
| Towner | 6 | - | 0 | 5e7540 | 15/15/15 | no | - | - |
| Trap-LeftArrow | 1 | - | 1 | 5fb7e0 | 15/15/15 | yes | 18 | - |
| Trap-Melee | 1 | - | 1 | 5fba60 | 15/15/15 | no | - | - |
| Trap-Missile | 2 | - | 2 | 5fb5b0 | 15/15/15 | no | - | - |
| Trap-Nova | 1 | - | 1 | 5fb9b0 | 15/15/15 | no | - | - |
| Trap-Poison | 1 | - | 2 | 5fb900 | 15/15/15 | no | - | - |
| Trap-RightArrow | 1 | - | 1 | 5fb6c0 | 15/15/15 | yes | 18 | - |
| TrappedSoul | 2 | - | 1 | 5e9f10 | 15/15/15 | yes | 30 | - |
| UberBaal | 1 | - | 1 | 5fd200 | 15/13/12 | yes | 6 | - |
| UberDiablo | 1 | - | 1 | 5e9df0 | 15/13/12 | yes | 6 | - |
| UberIzual | 1 | - | 1 | 5f8c80 | 15/12/8 | yes | 42 | - |
| UberMephisto | 1 | - | 1 | 5f81c0 | 15/10/6 | yes | 6 | - |
| Vampire | 9 | vampire1, vampire5 | 1 | 5f4a70 | 15/14/13 | yes | 126 | - |
| Vendor | 2 | - | 0 | 5e9e00 | 15/15/15 | no | - | - |
| VileDog | 10 | - | 1 | 5fa280 | 15/14/13 | yes | 18 | - |
| VileMother | 5 | - | 1 | 5fa010 | 15/14/13 | yes | 54 | - |
| Vines | 1 | - | 2 | 5ec6c0 | 15/15/15 | no | - | - |
| Vulture | 5 | vulture1, vulture2 | 1 | 5f3170 | 15/14/13 | yes | 72 | - |
| WillOWisp | 8 | - | 1 | 5f39b0 | 15/14/13 | yes | 66 | - |
| Wraith | 9 | wraith1, wraith2, wraith3, wraith4 | 1 | 5f0a20 | 15/14/13 | yes | - | 3000 / 494 |
| Wussie | 1 | - | 2 | 5ee3c0 | 15/15/15 | yes | - | - |
| ZakarumPriest | 6 | - | 1 | 5f72d0 | 15/14/13 | yes | 24 | - |
| ZakarumZealot | 5 | - | 1 | 5f6e60 | 15/14/13 | yes | 61 | - |
| Zombie | 5 | zombie1, zombie2, zombie3, zombie5 | 1 | 5efe20 | 15/14/13 | yes | - | 3000 / 649 |

## Shared-file lines for the merger

See `openspec/changes/monsterai/merge-lines.txt`.
