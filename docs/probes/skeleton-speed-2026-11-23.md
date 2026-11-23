# Skeleton minions: speed and pathing on the lab d2d (2026-11-23)

> status: probe record, current. Question (user, 2026-11-22): "skeleton minions seem way too fast and their pathfinding is off". Result: no deviation found; d2d moves a Necromancer's skeletons exactly as the game data and the traced game.exe follow rule say. No PR draft, no fork rebase.

Tags: V = measured on the lab d2d (lockstep, 1 tick = 40 ms, throwaway Necromancer, `--no-save`, setup verbs only, no writes after the run starts that change a minion); S = read in source or docs; U = not verified. Game data: only table and column names and numbers are quoted.

## 1. Which d2d, and what upstream has since

- Lab build: the devp20 build (fork branch `didle/vaultsell-20261122`, tip `63b1cecd`, off upstream `e80f59042` of 2026-11-21). V: `git merge-base --is-ancestor` says upstream #675 (`f8ec8bb9f`, "each mode at game.exe's own rate, a walk at its walk %", FUN_00623f50) IS in it. The "sandbox may lack the #675 fix" worry is wrong: it has it.
- Upstream main now (`ff41f2f4c`, 75 commits after `e80f59042`). Commits that touch pets, minions, walking or rates (S, from `git log` and `git show --stat`):

| Commit | What it changes | Effect on a skeleton's speed or path |
|---|---|---|
| `de438be3f` | a monster's turn goes on a step a frame after its walk ends (FUN_00648640); an NPC's and a pet's still snap | facing only; none on speed |
| `5239da23a` | a monster's strike waits for FUN_005539b0's tick at its mode rate (S1 plays at its walk rate) | attack timing of monsters; none on walking |
| `2d522f16f` | the same for a pet's strike (it waited at the AnimData speed alone: no IAS, chill or walk rate on a moving S1..S4) | attack timing, and only for a slowed or moving-S mode; a skeleton's plain A1 is unchanged (S, U not run) |
| `2cd871e9f` | a pet's unit seed comes off the game seed | RNG of the pet (wander offsets); not speed |
| `8e10dcdf5` | a seal's and the Blade's minions follow their boss (FUN_0058f380) | not a Necromancer pet |
| `1cd0f0730` | a revived monster keeps its corpse's sighted flag | Revive only |
| `c5c3edda8`, `556d49e3e`, `8412086ca`, ... | merc Frozen Armor, bird flight, client drawing | unrelated |

None of them changes how fast a Raise Skeleton minion walks or where it goes. The two that matter for speed (#675 mode rates, #659/#662 pet follow and revive) are already in the lab build. Rebasing onto upstream main would not change what is measured below, so no rebase was done (decision (4a) not taken).

## 2. Expected speed (from the game data and the rules)

- MonStats `necroskeleton`: Velocity 13, Run 15 (S, the repo's own table loader on the lab data; same read for `necromage` 12 / 14). CharStats Necromancer: WalkVelocity 6, RunVelocity 9.
- Unit conversion (S, docs/fixrun.md and FUN_00621360 / FUN_00623f50 as quoted in tools/emu/verify_walk_mercpet.py): one tick moves `velocity x 256 / 4096` subtiles, so 0.375 subtile per tick at velocity 6 (hero walking, 1.875 cells/s) and 0.5625 at 9 (hero running, 2.81 cells/s). A skeleton at the table's Velocity 13 moves 0.8125 subtile per tick (4.06 cells/s); at Run 15 (walk % 115) 0.9375 (4.69 cells/s).
- The follow rule (S, upstream `rules::necropet_think` / `PetFollow::follow`, cited there as FUN_005e4cf0, FUN_005e3ea0, FUN_005e45d0): over 0x32 subtiles (10 cells) from the owner it teleports to it; 0x1d subtiles or more it follows with mode 0 at the pace `Run x 100 / Velocity - 100` (15 percent for the skeleton); otherwise mode 1 retraces the owner's footsteps at a pace of `rand(40) + 40` percent (40 to 79) when it has to catch up. Velocity x (100 + pace) / 100 therefore tops out at 13 x 179 / 100 = 23.3 (7.27 cells/s). Idle pets within the reach (7 or 8 subtiles plus a quarter of the other pets) stand or wander 4 subtiles.
- Independent check of the walk % chain: tools/emu/verify_walk_mercpet.py (card verifyprs2, game.exe's own FUN_00623f50 under the emulator, pet cases: 89,544 differing cases before the fix, 0 after; the fix is what #675 became upstream). Not re-run here.

## 3. Measured speed (V)

Probe: `docs/probes/skeleton_speed_probe.py` (own devctl session, `docs/probes/skeleton_speed_lib.py`), analysis `docs/probes/skeleton_speed_analyze.py`. Setup: `debug goto`, `debug points 70 3` (Raise Skeleton, 3 points, Skeleton Mastery 5), three Fallen spawned and killed, three casts of Raise Skeleton at the corpses, then the hero moves with `cmd move`; every tick: hero position (`state` cam) and `debug minions` (id, x, y, mode). `debug kill` runs every third tick so that no natural monster stays alive (the pets are not in the level's monster list). Traces: `skeleton-speed-trace-cave.csv`, `skeleton-speed-trace-bloodmoor.csv` (columns phase, tick, hero x y, minion id, x y, mode), live-foe counts `skeleton-speed-foes-*.csv`.

Speeds are the median over 5-tick windows with movement, in cells per second (25 ticks/s); 1 cell = 5 subtiles.

| Unit | Mode | Observed cells/s | Observed subtile/tick | Expected subtile/tick | Observed / expected |
|---|---|---|---|---|---|
| Hero | walk (cmd move) | 1.85 to 1.89 | 0.370 to 0.379 | 0.375 (velocity 6) | 0.99 to 1.01 |
| Hero | run (key r) | 2.80 | 0.560 | 0.5625 (velocity 9) | 1.00 |
| Skeleton | following, WL | 4.03 to 4.06 (Blood Moor, area 2, leg to the row) | 0.807 to 0.812 | 0.8125 (velocity 13) | 0.99 to 1.00 |
| Skeleton | walk % 115 (mode 0 follow) | 4.65 | 0.931 | 0.9375 (velocity 15) | 0.99 |
| Skeleton | catch-up, mode 1 pace | 6.4 to 7.3 (max windows) | up to 1.46 | 1.16 to 1.45 (velocity 18.2 to 23.3) | within the band |

Per-tick velocity histogram of all moving skeleton ticks in the cave run (step x 80, rounded to velocity units, 1,270 moving ticks): 13: 552, 12: 183, 14: 137, 10: 117, 15: 70, 3: 67; a tail at 21 to 23: 16 ticks, none above 23. The Blood Moor run (1,800 moving ticks, partly a foe chase) has the same peak (13: 823) and the same ceiling, none above 23, with 91 ticks at 21 to 23. So the speeds sit on the table's numbers and the pace band of the follow rule (V). Catch-up sample (cave, minion 66, hero running at 2.8 cells/s): ticks 1724 to 1729 step 22.1 / 22.9 velocity units (6.9 to 7.2 cells/s) while 2.1 to 1.2 cells behind the hero.

Relative to the hero: a skeleton at the table's walk speed is 2.17 x a walking hero (13 / 6), 1.44 x a running hero (13 / 9), and on a catch-up it is up to 7.27 / 1.875 = 3.9 x a walking hero and 2.6 x a running one. That is what "too fast" looks like: it follows from the table and the rule, not from a defect in the engine step. A bot hero that mostly walks (docs/fixrun.md: the walking share was unexplained) makes the ratio worse. Whether the real game's skeletons feel slower than this is U (no game.exe trace of a Necromancer's pet on screen was taken here).

## 4. Pathing (V, positions in the CSVs)

| Check | Result |
|---|---|
| Stuck in a moving mode (below 0.005 cells per tick for 10 ticks or more) | 0 runs in the cave, 0 in the Blood Moor (one minion: 1 run in an earlier, longer cave route) |
| Walking through walls | not checked against the level grid (U): no position was compared with `debug grid`; nothing odd stood out in the tables |
| Corridor with corners (cave level 1, seed 5, about 21 cells of hero route, a corner or two) | the three skeletons follow; route length 21.4 / 30.8 / 40.8 cells against the hero's 21.0 (walking leg; they wander, see below), lag up to 6.9 cells, mean distance to the hero 1.4 to 3.6 |
| Lag then snap | In the Blood Moor a minion fell 10.0 cells behind (tick 572: hero 35.22, 16.50; minion 44.26, 20.83, mode WL) and at tick 573 it was put at 33.5, 19.5 (3.4 cells from the hero, mode NU): the 0x32-subtile teleport of `necropet_think` (S), exactly at 10 cells. Earlier, longer cave routes (not saved) showed the same snap at 9.8 to 13.9 cells before it, in total 5 snaps in about 3,000 ticks. |
| Idle at a distance | At rest a pet stands 1.5 to 3.6 cells off (inside its reach) and does not follow a hero that has started to walk until the hero is about 5 to 7 cells away (an earlier, foe-contaminated Blood Moor trace that is not in the CSVs: one of three stood still for 2.4 s, the other two for 0.7 s and 1.7 s; in the saved cave trace the delay after the first `cmd move` was not measured separately). The rule's `stand` / `stay` (15 percent, 15 frames) explains a delay, not 2.4 s: U (not traced in game.exe here). |
| Direction reversals | 4 to 9 per minion in the cave corridor leg, 13 to 16 in the Blood Moor legs. They are the idle wander (4 subtiles) and a foe chase (a natural monster spawned and was killed within 3 ticks: `skeleton-speed-foes-bloodmoor.csv` has a live foe in 68 of 99 samples, the cave file in 0 of 75; the Blood Moor legs are therefore partly a foe chase, not a clean follow; A_idle and the leg to the row are clean). |
| Different route from game.exe | not measurable here: no game.exe trace of the same scene. The follow target spots (8 ahead of the owner's path end, then halfway, then the owner's footstep ring) are the rule of `PetFollow::follow`, which cites FUN_005e3ea0; U for the byte-level path (FUN_ path function not compared). |

Dead ends: a first cave route chosen on the 1-cell grid wedged the hero itself at a gap (the same position with and without pets: the control run with no pets stood there too), so it is a probe artifact, not a pet effect; the route was rebuilt on the subtile grid with obstacles grown by one subtile and cut to 125 steps.

## 5. Decision

- Not fixed upstream, and nothing to fix found: the measured speed ratio is 1.00 of the table and of the follow rule's band; the only pathing oddities (standing idle at 2 to 3.5 cells, up to 10 cells of lag before the teleport, a wander of 4 subtiles) are the rule as upstream traced it.
- No PR draft: there is no differential against game.exe showing a deviation, so it does not clear the proof standard.
- If the user still sees something wrong, the cheapest next checks (ready to run, need a game):
  1. On a real game.exe with a Necromancer (level 6 or more, Raise Skeleton 3 points), count frames for a skeleton over a 20-cell straight run of the hero and compare with the table above (13 per tick step 0.8125 subtile).
  2. Same with a walking and a running hero, to check the catch-up band (40 to 79 percent) against the screen.
  3. A standing hero 3.5 cells from a skeleton: does it walk up to 1 cell (`gap <= 1` back-off) or stand?
- The idle delay in section 4 (first trace: 0.7 to 2.4 s) is the one measured behaviour not explained by the traced rule. It is a candidate for a game.exe trace of FUN_005e4cf0 (the think interval of a NecroPet), not a claim.

## 6. Re-run

`python3 docs/probes/skeleton_speed_probe.py <d2d> <outdir> 5 <area>` with `D2_MPQ_DIR` set (area 2 Blood Moor, 9 cave level 1) on a lockstep d2d with `--lab-verbs --read-combat`; then `python3 docs/probes/skeleton_speed_analyze.py <outdir>/trace_area<N>.csv <outdir>/foes_area<N>.csv`. Needs one lab slot for about 10 s of wall time per area.
