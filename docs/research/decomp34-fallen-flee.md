# decomp34: the Fallen's flee scan (damage intake D2/D3) and the cost of a refused Teleport (P13)

> status: research, 2026-11-16, agent decomp34 (Sonnet), offline (emulator and fork builds on the NAS builder; no game, no live d2d, no Go change, nothing posted). Fork branch `fid/tele-refund-r1` (head `0a439860`) pushed to the private remote only. Follows [damage-intake-2026-11-15.md](damage-intake-2026-11-15.md) D2/D3, [monster-ai-trace.md](monster-ai-trace.md) D11, [../probes/fid/fid-teleport.md](../probes/fid/fid-teleport.md) D4, [../upstream/pipeline-2026-11-14.md](../upstream/pipeline-2026-11-14.md) row P13.

Tags: **V run** = game.exe 1.14d's own code run in the emulator (`tools/emu/dc34_*.py`, copies of the fork's `tele_mana.py`) against d2d's real build where stated; **V read** = disassembly read; **I** inferred; **U** unverified. Candidates were picked by (value to bot fidelity x tractability) from the damage-intake D2/D3, P07, P12, P13 and the un-traced rows: the Fallen flee (Den of Evil is the first-hour fight) and P13 (a Sorceress's hop is the bot's main move, and a differential shape exists).

## Answers

1. **D2 (a Fallen whose back-off of 12 is blocked): no deviation.** game.exe's FUN_005f02c0 sets state 1, drops the command, sets pace 50 and then goes on with the rest of the think when FUN_005defe0 fails (`0x5f03c3..0x5f03d6`); d2d's `fallen()` does the same. 1,122 of 1,122 cases with a failing back-off (real game.exe think, V run) went on to the later calls. Upstream's own Act 1 corpus already carries the `dying` word and the failed-move input for the Fallen (`tests/monsters/test_monster_ai.cpp:450..463`, `mon_ais_act1_cases.inc`), so the think itself is covered (S).
2. **The flee SCAN is not what d2d models, and the difference is real but tiny (D-ring below).** game.exe does not look at the Fallen's group; it reads a ring of the last four deaths of each near room. d2d looks at every monster of the level. They agree in 3,976 of 4,000 cases of a scene built from the real code; the 24 that differ are all ring overflow. No PR: the fix needs a per-room death ring in d2d and the effect is rare (section 3).
3. **P13 (a refused Teleport costs mana in d2d, not in game.exe) is a proven deviation with a fork branch.** 927 and 970 of 2,000 cases cost 24 points that game.exe does not take; after the claim 0. Branch `fid/tele-refund-r1`, body [bodies/tele-refund-body.md](../upstream/bodies/tele-refund-body.md). Ready for the independent verifier (not run).

## 1. The Fallen's flee scan (Q1..Q5)

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | What does FUN_005f02c0 scan at `0x5f032e..0x5f03df`? | `FUN_00620bb0(monster)` is the unit's room (types 2/4/5: `[unit+0x2c][0]`; others: the path's room via FUN_00648a80); `FUN_00619790(room, &list, &count)` gives `room[0]` (the near rooms) and `room[+0x24]` (their count); for each near room, the four words at `room + 0x38` are unit ids; each is looked up with FUN_00552f60(game, 1, id); a hit in mode 0 (DT) with `FUN_005dc530` (the AI distance) under 15 is the trigger | V read |
| Q2 | Who writes the ring? | `FUN_0061afa0(room, id)` (`0x61afa0..0x61afc2`): `room[+0x38 + 4 * head] = id; head = (head + 1) & 3`, head at byte `+0x14`. Its only caller is the death handler `FUN_005a6ff0` at `0x5a704d`, with the dying unit's own room (`FUN_00620bb0`) and id, for EVERY monster class | V read; V run (below) |
| Q3 | Does the real death handler fill it as read? | Spawned fallen1 and zombie1 units through FUN_00555230 in a Blood Moor game; FUN_005a6ff0 on five of them in one room: the unit's mode becomes 0 and the ring reads `[1,0,0,0] [1,2,0,0] [1,2,3,0] [1,2,3,4] [5,2,3,4]` (head 1 after the fifth): the fifth death overwrites the first | V run (`dc34_ring.py probe2`) |
| Q4 | Does the near list include the unit's own room? | yes, always (6 of 6 sampled rooms), and the touching rooms; rooms are 40 x 40 subtiles, so a unit within 15 of the Fallen is always in a near room | V run (`dc34_ring.py probe`) |
| Q5 | The distance? | `FUN_005dc530` = d2d's `ai_distance` (`(min + 2 max) / 2`): 4,000 cases agree on every one under or over 15 | V run (same run as below) |

Differential (V run on the game side, d2d side a port): `dc34_flee.py flee 4000 1` spawns a real Fallen and 0 to 7 other units (Fallen or Zombie, 70 percent at up to 22 subtiles, 80 percent killed through FUN_005a6ff0, in up to 83 rooms), then runs the REAL FUN_005f02c0 (acts and command calls hooked, scan functions real) and records whether the dying branch ran (state 1, pace 0x32, back-off). `dc34_cmp.py` applies d2d's rule (`ai.cpp:1008..1011`: any other monster in DT within `ai_distance` < 15) as a Python port.

| | cases | game flees, d2d predicts the same | differ |
|---|---|---|---|
| seed 1 | 4,000 | 3,976 | **24**, every one a room with 5 or more deaths (ring overflow) |

Fall-through (D2): of the cases with a failing back-off, 1,122 of 1,122 continued with the walk call after it (V run).

### Deviations (d2d vs game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D-ring | game.exe: only the last 4 deaths of each near room are known to the scan; a fifth death in a room evicts the oldest, even while it is still in DT. d2d: every monster in DT counts | V read; V run game side (ring contents, 24 of 4,000 cases); d2d side V read plus a port | game side exists (`dc34_flee.py`); the d2d side needs a real `think()` call with a pack: build it from `tests/game/test_target_cursed.cpp`'s World plus a `Level` and `think(game_data, level, monster, foes, ...)` with the pack's modes set, compare `monster.fleeing` |
| D3 | Seconds spent off per flee (pace 50, a walk of 12) | not measured on either side | I: the walk and pace code is covered by the 122,646-case move differential; a stepped scene would give the seconds |
| D2 | Blocked back-off | none (above) | covered by upstream's corpus |

Size (I): the ring only loses a flee when 5 monsters die in one room within one death animation (an area kill: Blizzard, Frost Nova, Fire Wall) and the one near the Fallen is the oldest. Single-target killing never overflows. For the bot's Paladin in the Den: none. A d2d fix is a per-room ring of 4 ids plus the near-room list: invasive for a rare case, so no PR candidate this round.

### Bot findings

- B1 (unchanged from the damage intake): `internal/combat/fight.go:368` `maxHostiles` counts hostiles in range, not adjacent.
- B2: the bot cannot tell a flee from a retreat from the monster list alone; a Fallen packs' "everyone walks away after a kill" is game.exe behaviour for any monster dying within 15, not only a Fallen kin (Q2: the ring takes every class). The earlier trace's "a dying Fallen" and "the group" wording was too narrow; the scan is by room and covers any monster.

## 2. P13: a refused Teleport costs no mana (Q6..Q9)

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q6 | Where does game.exe pay a skill with no start function? | in FUN_0056f7f0 after the do: `0x56fa51..0x56fa7a`, through FUN_0056bfe0; for a skill with no missile a zero do return leaves at `0x56fa3d..0x56fa41` before it | V read; V run |
| Q7 | What does do 27 return on a refusal? | 0; a landing 1 (FUN_00554ea0 put the unit somewhere) | V run (the oracle: 927 and 970 refused cases give 0 payments) |
| Q8 | What does d2d do? | `Fight::owe` leaves the cost owed (Teleport has no srvstfunc), `settle_cast()` pays it on the action frame BEFORE `spot()` (`strike.cpp` SC branch), and the Teleport refusals `break` without giving it back | V read; V run (parent) |
| Q9 | Counts | below | V run |

Game side: `tools/emu/dc34_telemana.py game N SEED` (fork copy `tele_mana.py`) runs `FUN_0056f7f0(skill 54, level, pay 1, 1, 0)` natively with the real do 27 on `sk_tele_spot.py`'s rooms; FUN_0056c160 (the mana check, `sk_mana.py`'s claim) is hooked to pass, FUN_0056bfe0 is recorded. The cases are `sk_tele_spot.py gen N SEED --no-units`. d2d side: `test_tele_mana --tele-mana` (real `Fight::cast_missile` and World ticks, mana against a control run, small pool so regeneration is under 400 of 6,144, chunks of 100 cases per process because the World's timers drift after about 190 casts in one process).

| 2,000 cases | refused (game 0, d2d paid before) | landed (both pay 6,144) | differ before | differ after |
|---|---|---|---|---|
| seed 1 | 927 | 1,073 | **927** | 0 |
| seed 2 | 970 | 1,030 | **970** (+2 test artifacts) | 0 (+2) |

The 2 artifacts of seed 2 are hops onto the caster's own subtile (the test reads a landing from the position); both sides pay equally. `dc34_tmcmp.py` exits 1 on a difference and on an empty or unequal file (both run). The self-check of `test_tele_mana` fails on the parent (`refused.paid == 0`), passes on the claim; `test_cast_mana` and `test_tele_spot` pass on the claim (V run, `tm2` binaries).

### Deviations

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D4 (P13) | refused hop: game.exe no payment, d2d pays 24 at level 1 | V run both sides | yes: fixed by `fid/tele-refund-r1` |
| D4b | other skills with no missile whose do can return 0 (Leap, Telekinesis without a target, curses): same rule in game.exe | V read only | build by the same oracle with the other do hooked to refuse |

### PR candidate

`fid/tele-refund-r1` (oracle commits `c97378b5`..`0c753690`, claim `4386559e`, tools `0a439860`) on upstream `b6e6940d`: yes, ready for the independent verifier. The oracle commits are five small steps because the fork remote denies a rewrite; the verifier's rebase should squash them. Body [bodies/tele-refund-body.md](../upstream/bodies/tele-refund-body.md). Not posted. Checks not run: `presubmit.py`, `pr_guard.py`, clang-tidy, full ctest (the verifier's pass).

### Bot findings

- B3: a refused hop costs 24 mana on d2d, so the `waitMana` the bot sits through after a failed hop is d2d-only (sk-teleport-trace B2); after the fix it vanishes. `hopCell` aims only at known walkable cells, so the bot rarely sends a refused hop.

## 3. Candidates not taken (reason)

| Row | Why not |
|---|---|
| P07 (monster block gate) | needs a per-spawn component pick (the shield slot) that d2d does not model at all; a fix is a model, not a rule |
| P12 (hire list once) | U whether the lists differ; low value |
| P09 / P11 | known small or no oracle |

## 4. Drill recipe for the fork verbs (lab, not run)

- R1 (Fallen flee scan): the Den or a throwaway Blood Moor on the lab binary (docs/forkdbg.md): `debug spawn fallen1 6`, stand near, `debug kill` one monster at a time and `monsters` mode each 15 frames: the others should show the walk-away within the same think after any kill within 15 subtiles, a Zombie kill included; then kill five in one room (Frost Nova) and see the oldest's neighbours stop fleeing. Expected game.exe: ring semantics above; d2d: every DT monster.
- R2 (P13): lab binary, a Sorceress with Teleport (docs/forkcheats.md setup on a throwaway), `debug stat 8` to read mana, cast at a point in a room two away; game.exe expects mana unchanged, the parent d2d minus 24 at level 1, the claim unchanged.

## 5. Matrix-cell updates (proposed)

- fidelity-matrix Fallen flee row: "scan = the last 4 deaths of each near room, any class (V read, ring contents V run); d2d scans every DT monster; differ only on a fifth death in one room (24 of 4,000, V run game side); blocked back-off equal (V run, 1,122 of 1,122)".
- SK-TELE row: "refused-hop mana: V run both sides, 927 and 970 of 2,000 differ on the parent, 0 on `fid/tele-refund-r1`".

## 6. Dead ends

- First harness run of the Fallen scan reported every case as fleeing: the rooms' rings and the earlier cases' units stay (the dead stay in DT forever with no ticks); fixed by zeroing every ring per case.
- `FUN_00620250` returning the fake record for the whole payment run crashed the mana path; the check is hooked to pass and the record is left null (`dc34_telemana.py`).
- The d2d side with a 100,000-point mana pool gave noise from regeneration (a rate of the maximum); a 100-point pool and per-100-case processes removed it.
- `World::monsters` does not exist (the field is `Fight::monsters`); the level's spawns in a test world block a d2d landing (units live in the Crowd) while the synthetic game rooms have none.
- Glossary: no missing term. Proposed wording for monster-ai-trace.md D11: "scan by the room's death ring, not by group".
