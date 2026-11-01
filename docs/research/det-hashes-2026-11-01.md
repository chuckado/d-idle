# Det state hashes for three upstream candidates (agent DETHASH, card dethash)

> status: current research record. Tags: V = run by us (how), S = read in a doc or source, U = unverified.

## What the harness is (V)

Upstream's own tree has NO det harness (no `det`, `state_hash` or `determinism` in its tests, tools, CI workflow or CONTRIBUTING; V by grep). The "det state hashes" of the PR standard are OUR scripted scenarios on a headless d2d with the lab patch set (`--lockstep`, `step n`, `monsters`, `debug ...` verbs; docs/d2d-fork.md section 5, docs/d2d-lab6.md): a fresh d2d process per run (seed 3, `--start-screen ingame`), scripted `step` blocks, an md5 over the `monsters` dump after each block. It needs a d2d slot on the worker host (`d2dslot.sh run`, hidden headless d2d, game data dir for the MPQs) but NOT game.exe, ptrace or a live character. It is the only deterministic-hash tool we have; it covers monsters (positions, hp, state), not mercs, pets, items or the player's own position.

The lab patch set is not on upstream main, so each build is upstream main (or the candidate head) plus the lab patch commit `3a629371` of the fork (lab set ported onto upstream 03f42254, branch `lab-port-148`), cherry-picked onto the tree. Two conflicts (the shift-click belt code, `town/update.cpp` and `rules.hpp`) were resolved to upstream's version (upstream has since merged the same feature); no test or devctl code uses what was dropped. All four builds carry the identical lab commit (`094ec6cc`).

## Builds (V, worker host, worktrees under the PR checkout dir `dethash/`)

| build | tree | d2d sha256 (first 12) |
|---|---|---|
| base | upstream main 72132b99 (still the tip at the run) + lab commit | c899f1690... |
| 155 | PR #155 head 6270d162 (claim d9d64d53 + tooling) + lab commit | e3d1f5126... |
| chill | `fix/static-field-chill` head 79c9a496 (155 + claim 46ef6c75 + tooling) + lab commit | c4b9428d8... |
| walk3 | `fix/walk-velocity-v3` head 8db11206 (five claim commits + tooling), base 72132b99, + lab commit | 00c202f32... |

Candidate commits were replayed with `git cherry-pick 72132b99..<head>` on top of the lab commit (no conflicts); the authors' worktrees and branches were not touched. `cmake -DCMAKE_BUILD_TYPE=Release -DD2D_WERROR=OFF`, target `d2d` only.

## Scenarios and commands (V)

`det7.py` (scratch on the worker host; det6.py of the lab plus three claim-driven modes), one run per scenario and build: `D2D_BIN=bin/d2d.NAME ~/d2d-lab/d2dslot.sh run -w 600 -- python3 det7.py SCENARIO NAME`. Each run: `debug stat 12 L`, `debug goto 2` (Blood Moor), warp next to the zombies and quill rats, `click 400 500`, 8 blocks of 500 (passive) or 250 steps, `monsters` after each. Scenarios:

- passive: clvl 8, zombies standing (det6 passive, class 1).
- fight: clvl 30, warped next to the pack, zombies walk (det6 FIGHT=1).
- cold: class 4 sorceress, clvl 30, Ice Bolt (39) cast at the first monster twice per block (`cmd skill`).
- cold2: the same with Frozen Orb (64).
- static: the same with Static Field (42).

Baseline self-check: base run twice per scenario gives identical hashes in all five (V). The harness reproduces the recorded lab hashes on this base: passive `d1ca6c46...` and fight `5f906303...` match docs/d2d-fork.md section 5 (a cross-check that the ported lab set does not shift the old results).

## Hashes (md5 of the `monsters` lines, first 12 hex; V)

| scenario | base | 155 | chill | walk3 |
|---|---|---|---|---|
| passive | d1ca6c463f45 | same | same | same |
| fight | 5f90630302ee | same | same | **1124443e0e4e** |
| cold (Ice Bolt) | bba8362ccde9 | same | same | **5ef0cf795284** |
| cold2 (Frozen Orb) | bba8362ccde9 | same | same | **5ef0cf795284** |
| static | 34a1b9ef9f44 | same | same | **ef7c7c56bf1b** |

The full-log hash (every reply except frame, music and sfx counters) agrees with the table: identical for 155 and chill in every scenario, different for walk3 except passive.

## Per candidate

### PR #155 (Static Field only, head 6270d162): DET-CLEAN

All five scenarios identical to base. Static Field IS cast and hits in the static scenario (zombie hp 6/8, 8/10 in every build), but the claimed changes (verifier: synthetic-only calc2 and life - 1 branches, MDR and absorb not modelled) do not show in these monsters, so "unchanged" means no regression, not that the claimed field was exercised. U: whether any det scenario reaches the claimed branch.

### chill (fix/static-field-chill, head 79c9a496): DET-CLEAN with a coverage gap

All five scenarios identical to base, including the two cold scenarios. Gap (V): in the cold scenarios the missiles fire (Frost Nova poke: 64 missiles) but the monster rows never changed (hp 8/8 throughout, and cold and cold2 give the SAME monster hash, so the monsters do not react to which cold skill is cast); the zombies' resists or immunities mean no damage and no visible chill in `monsters`. So the det run shows the chill commit moved nothing it should not have, but it does NOT show a cold missile reaching the changed `missile_blow` path. A scenario with a cold-vulnerable monster and a visible chill state is needed before the det step is considered to cover the claim (U).

### walk velocity v3 (head 8db11206): DET-CHANGES-EXPLAINED

passive unchanged (zombies stand). fight, cold, cold2: 5 of 56 monster rows differ, first at block 4 (first differing row: zombie #1, fight base 36.50 38.54 vs 36.50 38.41), all rows of walking zombies (state WL on both sides); static: 16 of 56 rows, first at block 1 (zombie #1, 35.30 38.64 vs 35.51 38.78, same hp 6/8); one differing row in static is a zombie that stopped (NU) at the last sample but walked in earlier blocks, and its earlier positions differ too. In every scenario: no row of a monster that never walked differs, hp and state letters equal, the player's life and camera lines are identical. Explanation (S, claim commits d8e63d4a, a5cf6415): a walking monster now moves at MonStats Velocity with the walk percent as a sum, and a chilled one adds coldeffect; that is the claimed change in monster walk speed, so walking zombies shift by fractions of a subtile per tick. Not covered by det (V, no merc or pet in any scenario): the merc and pet commits (455e366f, 8e5ff4cd) and the player velocitypercent commit (1c3fa4e5, the player's position is not hashed). Per-commit attribution was not bisected (U); the explanation is at the level of the claim.

## Limits

- Baseline is upstream main plus the ported lab set, not pure upstream (upstream has no scripted-scenario harness).
- 25 runs, then 7 reruns of runs that died at start-up (the previous d2d unlinking the shared devctl socket; the harness now waits for the old process; V: the rerun hashes equal the other runs of the same scenario).
- Monsters only: zombies and quill rats in Blood Moor, one seed (3).
- Nothing was posted or pushed; the worker host scratch is `~/d2d-lab/pr/dethash/` (builds, out/, results.txt, det7.py).
