# Sandbox d2d build devp9 (2026-11-15)

> status: current. Task: TASK-535 (sandbox2). Tags: V = verified in this run (how said), S = read from a tool or log, U = unverified.

The play sandbox (container `d2play-20261108`, docs/dev-update.md, ops repo `docs/play-sandbox.md`) ran image `d2dbot:devp8` (fork lineage `didle/rebase-20261114b` plus `sandbox-herostats` and `sandbox-mercverbs`), which stopped at upstream PR 494. `devp9` is the same carried patch set on current upstream main.

## Fork branch

- Branch `didle/rebase-20261115` in the private d2d fork, on the private remote (same day push). Base: upstream main 485fe535 (2026-11-15, after PR 560). Head: fabae935 (README-DIDLE line added). V (`git push` printed the new branch).
- Built by cherry-picking, in order, every non-merge commit of the devp8 lineage (41 commits: the lab patch set port, lab and test cheats, overlay, SPDX, dump merc, framefeed, cursor, icon art, item verbs, warp debug verbs, merc verbs, hero stats) with `-x`. One conflict (apps/d2d/main.cpp, the input-device block: kept upstream's touch code and our `click_mods_event(event)` call). V.
- Dropped as now upstream: `c0555966`, its revert `a1e50a79` and `75891650` (the nearest-free walk to a clicked warp): upstream has the warp arrival rule (PR 490), the Select-box click (66aece1a) and the walk to the tile unit (PR 547). V by reading the commits; the Black Marsh to Forgotten Tower start-direction count that justified `75891650` was not re-run (U).
- Kept (not upstream, fork-only): lockstep and time-scale, dump verbs (stats hero block, gear sums, merc live gear list), lab and test cheats behind `--lab-verbs`, `--overlay-dir`, framefeed and `cursor`, item verbs incl. merc-equip / merc-unequip, debug warps click point (`060c8901`, `c3e24fcb`), icon art, README-DIDLE.

## Merged upstream PRs now in the build (author shatteredeagle, from `gh pr list --state merged`)

Base 485fe535 contains every merged PR of ours as of today, newest first: 557 (big hit interrupts a swing or cast only if the skill allows), 553 (a missile with no physical damage cannot be blocked), 548 (a big hit does not restart a get-hit), 547 (clicked warp walk to its tile unit), 541 (block holds game.exe's length), 539 (saved purse or stash over its cap), 530 (swing starts at the weapon's start frame), 490 (warp lands on the nearest free subtile), 445, 444 and the older trace and fidelity PRs. devp8 had the lineage only up to 494 (S: the branch merge-base was 73bf565b, 412 upstream commits older than today's main).

## Build and image

See the section below (filled in after the run).
