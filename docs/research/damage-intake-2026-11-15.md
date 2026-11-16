# Damage intake of a surrounded hero, Den of Evil (agent dmgcheck, 2026-11-15)

> status: research, current. Question (user report): a fight where the hero stood against a wall in a giant group of Den of Evil monsters and took zero damage. A fidelity regression, or what game.exe also does? Tags: V read (code or git read), V run (observed running), D (data tables / the sandbox log), S (an existing proof doc), I (inferred), U (unverified). **What was NOT done:** the offline d2d scenario run. The builder was unreachable from this box (ssh stalled), game data and the fork build live only on the worker host, and the brief bans anything on it. So there is no V run in this doc for d2d vs game.exe counts; the scenario is a recipe (section 7).

## Verdict

1. **No evidence of a fidelity regression; the observed intake is what the rules predict (D, I).** In the Den a clvl 5-6 hero (defence 28) is hit by a level 1 Fallen about 6 percent of its swings (hit rule V run, 0 differ in docs/research/monster-attack-trace.md) for 1 to 2 hp. A pack of 5 adjacent Fallen without any flee delivers about 0.24 hp/s; the live Den rate over 1,535 s of fighting was 0.42 hp/s, and a fight of 20 s or longer ended with no damage in 2 of 21 cases (10 percent, the model says 4 percent for 5 attackers, up to 60 percent when pack flees dominate).
2. **The devp8 build does not contain our recent PRs (V read, git ancestry).** The lineage `didle/rebase-20261114b` (tip 75891650) stops at upstream PR #494. #515 to #544 are absent, so #539, #541, #542, #544 are not in it, and #547, #548, #553, #557, #559, P17, P18 (branches `fid/*`, none merged into the lineage) cannot be causes. Without #548 and #557 the old rules apply (get-hit restarts; a hit interrupts any swing), both of which make the hero take MORE, not less, damage.
3. **Plausible extra reason, same in game.exe (V read, I on size): the Fallen flee rule.** Fallen, Carver and Devilkin share one think. Any pack mate in DT within 15 subtiles makes every Fallen of the pack, on its next think, set state 1, drop its command, pace 50 and back off 12 (FUN_005f02c0, `0x5f02c0`..`0x5f03df`; d2d: thinks/act1.hpp `fallen()`, game/ai.cpp:987, game/ai/act1.cpp:26). Den monsters have 5.9 hp on average, a Paladin kills one per swing, so deaths are near continuous and the pack is pushed off more often than it swings. This is the V-read twin on both sides; its off time per flee is U (section 5).
4. **No PR candidate**: no d2d deviation is proven; nothing built. Fork lineage branch names untouched, nothing pushed.

## Q1. Live Den numbers (D, sandbox log of the last 6 h, read only via the container log)

Source: `docker logs --since 6h` of the sandbox container (one read, no writes, no verbs). 358 minute lines. The character changes at 02:03 UTC from a clvl 17 test hero to the fresh Paladin (clvl 1 to 6), so the window below is the new hero only.

| Window (area 8, clvl 1-6) | Value |
|---|---|
| Fights logged (`combat fight area=8`) | 138, 1,535 s, 637 kills, 649 damage = **0.42 hp/s** |
| Same in Blood Moor (area 2) | 45 fights, 265 s, 133 kills, 29 damage = 0.11 hp/s |
| Fights 20 s or longer | 21, damage 0.40 hp/s, 2 with zero damage |
| Fights 10 s or longer | 51, 5 with zero damage |
| Big fights (maxHostiles 8 or more) | 33 fights, 631 s, 300 damage, 4 with zero damage, the longest zero one 28 s with 18 kills and 11 hostiles in range |
| Minutes in the Den (both ends in area 8) | 26 minutes, 321 kills, 376 damage, 4 potions; 3 minutes with 10 or more kills and at most 5 damage |

Reading: damage is not zero on average; single fights and minutes with zero or near zero damage exist at the rate the rules allow. The `maxHostiles` field counts hostiles in range, not adjacent ones (internal/combat/fight.go:368), so a "giant group" in the log is not a count of attackers (B1). The window has 13 `Save and Exit requested (chicken / leave game)` lines in 57 minutes and many minutes with the life minimum at 45 to 55 percent, so this hero does get hurt in the Den; the zero-damage fight is one tail case, not the norm; no potion use sits in the quiet minutes. The intake card's `lifeLost` field is not deployed.

## Q2. What the rules expect (D + V run for the hit rule; I for the cadence)

| Piece | Rule | Tag |
|---|---|---|
| Hit chance | AR x 100 / (AR + DEF) x 2 x mlvl / (mlvl + clvl), clamp 5..95; d2d equal to game.exe `0057d9b0` on 100,000 cases | V run (monster-attack-trace.md) |
| Den units | mean level 1.2, mean life 5.9, 849 units in the room population | V run (monster-stats-trace.md) |
| Fallen mlvl 1, AR 8, swing 1 to 2 | at clvl 6, DEF 28: 6.3 percent; mlvl 2 (AR 12, swing 1 to 3): 15 percent | D + formula |
| Cadence | think every aidel = 15 frames; in melee aip3 = 50 percent swing (Normal), else 30 percent S2 or stand 10; swing 10 frames | D + V run (mon-wake-trace, monster-ai-trace) |
| Flee | a pack mate in DT within 15 subtiles: state 1, command 0, pace 50, back off 12 | V read (FUN_005f02c0), d2d code read |

Expected, `tools/emu/dmg_model.py` (Monte Carlo of the rows above, I, not an emulator run): 5 adjacent level 1 Fallen, no flee, 20 s: 4.8 hp (0.24 hp/s), P(no damage) 4 percent; 8 attackers 0.38 hp/s, 0.6 percent; 3 (a corner) 0.14 hp/s, 15 percent. With 40 percent of thinks pre-empted by a flee (3 s off): 5 attackers 0.07 hp/s, P(zero) 38 percent. Level 2 Fallen (15 percent hit): 0.75 hp/s without flee, 0.23 with 40 percent flee. Live 0.42 hp/s sits between: consistent with partial flee suppression and a mix of 1-2 level units (I).

Geometry (D, I): against a wall a unit has 5 neighbour cells, in a corner 3, in open ground 8; reach is `melee_rng + 1` by unit distance (ai.cpp:977), so a second ring of larger monsters may also reach; the model's N is the attackers in reach.

## Q3. Build ancestry (V read)

`git merge-base --is-ancestor` against `origin/didle/rebase-20261114b`: the tips of the PR worktrees (hit-interrupt a78f66a3, missile block cd314f2c, melee block oracle 3d3d6b25, no restart b18acf7e, knock-in-gethit 4e911f21) are all not in it; its newest merged upstream PR is #494. `sandbox-mercverbs` and the zeal and swing branches sit on the same base. Upstream PRs absent: #515 to #527, #533 to #542, #544.

## Deviations D1..

- **D1 (build, not a rule):** the sandbox build lacks #541 to #557. Effect on intake: the hero's get-hit and swing interrupts follow the old d2d rules (I: more hero downtime, so more damage, not less); monster get-hit restarts on each hit (more monster stun, so less damage from one that survives a hit). Net size U; differential exists (the `fid/*` oracles `tools/emu/*hitrec*` per PR).
- **D2 (U):** when the Fallen's back off of 12 cannot be taken (wall, crowd), d2d's `fallen()` falls through to the attack branch (thinks/act1.hpp:171-173). game.exe's FUN_005defe0 failure path was not read for the same case. Differential could be built: run FUN_005f02c0 in the `tools/emu/mai_ais.py` harness with a scene where the back-off act reports failure, compare the returned act with `rules::mon_think`; oracle = game.exe's returned act and the `0x5f03xx` store sequence.
- **D3 (U):** time spent fleeing: d2d runs a monster's walk from the step budget and path (FUN_005defe0 walk, pace 50 steps); game.exe's pace (FUN_005de190 stores velocity percent consumed by the next mode) was named in monster-ai-trace.md (D1 there; for the Fallen now ported in this lineage). The actual seconds off per flee are not measured on either side. Differential: the emulator think harness covers the think output only; the seconds need a stepped scene (recipe 7).

## Bot findings (B, not changed)

- **B1:** internal/combat/fight.go:368 `maxHostiles` = candidates in range, not adjacent; internal/combat/castfirst.go:259 logs it, so reports of "surrounded" cannot be checked from the log. A count of hostiles within melee reach and the life lost per fight (the intake card) would settle this class of report.
- **B2:** internal/combat/crowd.go:113 steps away from a crowd of 7 or more at a life percentage (`combat crowd: stepping away`, 144 lines in the window); this moves the hero off walls, so it also changes how often a hero is "against a wall" in the log (I).

## Matrix-cell proposals

- docs/research/fidelity-matrix.md, Fallen flee row: "V read both sides (FUN_005f02c0 = thinks/act1.hpp `fallen()`); off time U; blocked back-off D2 U".
- A sandbox-build row: "devp8 lineage = upstream #494 plus lab patches; #515-#557 absent (V read)".

## 7. Drill recipe (not run)

On the builder (never the game host), from the fork checkout of this lineage and, as the control, upstream main:
1. Build the tests (`tools/ops.sh run build-fork.sh <branch> --tests` needs the worker checkout; if forbidden, stage a `git archive` of the branch into the builder by hand).
2. Write one test `test_den_crowd` on the pattern of `tests/game/test_den_party.cpp` (needs `D2_MPQ_DIR` and `D2_PATCH_INSTALLER`): `World` in the Den (`DenQuest::kDen`), a Paladin with the live stats (clvl 6, life 113, defence 28, weapon 2-7), six to ten Den units from the room population (Fallen mlvl 1-2, a Shaman, a Zombie), the hero on a cell with one wall side (then a corner), no hero actions, `world.tick()` for 500 ticks (20 s): count `Monster` swings (mode A1 starts), hits (life drops), damage; 200 seeds. Run on the lineage and on upstream main.
3. Oracle: expectation from section 2 per seed; for game.exe the think outputs via `tools/emu/mai_ais.py` (Fallen, 25 AIs) with the `dying` input set and unset, plus the blocked back-off scene of D2.
4. Add the kill process: kill one monster every 1.2 s (hero swing stand-in) and record the share of thinks that returned the flee act.
5. Compare: swings, hits, damage per 20 s must be within the Monte Carlo band of `tools/emu/dmg_model.py`; if the lineage differs from upstream main by more than the band, bisect the `fid/*` branches into it one by one.
Fork verbs for a live confirm later (docs/forkdbg.md): `dump monsters` and the lifeloss counter from the intake card on one throwaway character, hero against a wall; read only.

## Shared-file lines for the merger

BACKLOG: - DMGCHECK: surrounded-hero zero damage in the Den: no regression found, sandbox build lacks PRs #515-#557, Fallen flee rule explains quiet fights (docs/research/damage-intake-2026-11-15.md)
INDEX: - docs/research/damage-intake-2026-11-15.md: Den of Evil damage intake, build ancestry, flee rule, drill recipe
CHANGELOG: - docs: damage intake of a surrounded hero in the Den (docs/research/damage-intake-2026-11-15.md).
