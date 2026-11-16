# Sandbox d2d build devp9 (2026-11-15)

> status: current. Task: TASK-535 (sandbox2). Tags: V = verified in this run (how said), S = read from a tool or log, U = unverified.

The play sandbox (container `d2play-20261108`, docs/dev-update.md, ops repo `docs/play-sandbox.md`) ran image `d2dbot:devp8` (fork lineage `didle/rebase-20261114b` plus `sandbox-herostats` and `sandbox-mercverbs`), which stopped at upstream PR 494. `devp9` is the same carried patch set on current upstream main.

## Fork branch

- Branch `didle/rebase-20261115` in the private d2d fork, on the private remote (same day push). Base: upstream main 485fe535 (2026-11-15, after PR 560). Head: fabae935 (README-DIDLE line added). V (`git push` printed the new branch).
- Built by cherry-picking, in order, every non-merge commit of the devp8 lineage (41 commits: the lab patch set port, lab and test cheats, overlay, SPDX, dump merc, framefeed, cursor, icon art, item verbs, warp debug verbs, merc verbs, hero stats) with `-x`. One conflict (apps/d2d/main.cpp, the input-device block: kept upstream's touch code and our `click_mods_event(event)` call). V.
- Dropped as now upstream: `c0555966`, its revert `a1e50a79` and `75891650` (the nearest-free walk to a clicked warp): upstream has the warp arrival rule (PR 490), the Select-box click (66aece1a) and the walk to the tile unit (PR 547). V by reading the commits; the the Black Marsh warp start-direction count that justified `75891650` was not re-run (U).
- Kept (not upstream, fork-only): lockstep and time-scale, dump verbs (stats hero block, gear sums, merc live gear list), lab and test cheats behind `--lab-verbs`, `--overlay-dir`, framefeed and `cursor`, item verbs incl. merc-equip / merc-unequip, debug warps click point (`060c8901`, `c3e24fcb`), icon art, README-DIDLE.

## Merged upstream PRs now in the build (author shatteredeagle, from `gh pr list --state merged`)

Base 485fe535 contains every merged PR of ours as of today, newest first: 557 (big hit interrupts a swing or cast only if the skill allows), 553 (a missile with no physical damage cannot be blocked), 548 (a big hit does not restart a get-hit), 547 (clicked warp walk to its tile unit), 541 (block holds game.exe's length), 539 (saved purse or stash over its cap), 530 (swing starts at the weapon's start frame), 490 (warp lands on the nearest free subtile), 445, 444 and the older trace and fidelity PRs. devp8 had the lineage only up to 494 (S: the branch merge-base was 73bf565b, 412 upstream commits older than today's main).

## Build and image

- d2d: `tools/ops.sh run build-fork.sh didle/rebase-20261115 --tests` (the NAS builder, d2-builder container, `ccache`, jobs 6; about 7 minutes cold). d2d sha256 prefix a74efecd, 17,678,096 bytes. V (script output).
- Builder tests (`ctest -R` of the 8 built binaries inside the builder): test_rules, test_combat, test_monsters, test_protocol pass; test_armor_hit, test_freeze_resume, test_game and test_next_hit are Skipped there (no game data on the builder). The data-backed oracle runs were not done: running the test binaries on the worker host was refused by the permission classifier and the brief says nothing runs there but the deploy. Not run: U. The 8 targets do not include the new fork-only tests (hero stats, merc verbs); they were built in the earlier lineage and not re-run here (U).
- Image: `tools/dev-update.sh --image d2dbot:devp9 --d2d /data/build/didle/rebase-20261115-fabae935-d2d/apps/d2d/d2d` with `D2_LABRUN_HOST=<sandbox host>` (context on the builder, build on the NAS through `tools/imgbuild.sh`, one DEV=1 image, a few seconds because d2d is prebuilt). Moved with `tools/imgbuild.sh save d2dbot:devp9 OUT.tar.gz` then `docker load` on the sandbox host (166 MB). V.

## Deploy

Same compose set as the devp8 swap (project `play-build-20261108`: compose.yml, override.yml, dev.yml, feed.yml, iv/iv.yml for `D2BOT_ARGS=--d2d-args --item-verbs`), `--no-deps d2dbot`, `D2_IMAGE=d2dbot:devp9`:
1. `docker stop -t 60 d2play-20261108` (the bot saves the hero on TERM).
2. Volume backup (read only, after the stop so the save is consistent): `docker run --rm --user 0 -v d2play-data:/data:ro --entrypoint tar d2dbot:devp9 czf - -C /data .` into `~/play-build-20261108/bak-devp8/d2play-data-pre-devp9.tgz` (79.7 MB, 1568 entries) on the sandbox host. Restore of a backup: stop, `tar xzf` back into the volume the same way. The `dev-sandbox.sh restore` tool only restores its own reset archives, so it was not usable for a plain backup (S: read in docs/dev-reset.md).
3. `docker compose -p play-build-20261108 -f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv.yml up -d --no-deps d2dbot`. The volume `d2play-data` and the game mount are untouched.
4. `tools/dev-update.sh --serve start`, then `tools/dev-update.sh --commit origin/main --to d2play-20261108` with `D2_LABRUN_HOST=<sandbox host>` (and the values of `local.env` loaded: without `D2_WORKER_HOST` the update is REJECTED with a placeholder host): `UPDATED ... 0.0.261115193109 -> 0.0.261116033250 in 15s`. `bot.farmReset` was not touched (tuning.json unchanged).

## Rollback

- Bot only: `D2_LABRUN_HOST=<sandbox host> tools/dev-update.sh --rollback --to d2play-20261108`.
- Image: on the sandbox host in `~/play-build-20261108` run the swap command of step 3 with `D2_IMAGE=d2dbot:devp8` (the image stays loaded; the exports are in the step 3 line and the update URL and key are read from the container). Data back to before: restore `bak-devp8/d2play-data-pre-devp9.tgz`.

## Verification

Read only (container logs, state files; no game input, no live-hero API, no token). Window 03:33 to 03:43 UTC after the bot update, V:

| Check | devp8 (10 min before the swap) | devp9 |
|---|---|---|
| container | healthy | healthy, 0 restarts |
| kills in 10 min | 167 | 180 |
| xp from kills | 4010 | 3939 |
| `Save and Exit` lines | 1 (a chicken, no burst) | 1 (a chicken, no burst) |
| ERROR lines | not counted | 0 |
| hero | clvl 7, alive | alive, loads, clvl 8 by the end (level up in the window) |
| watch page | n/a | `/` answers 200; the inventory state file is intact |

Damage taken was not logged as a number (U): life stayed in the normal range in `zone entry` lines (131/135 at the end). `bot.farmReset` left at its default (tuning.json untouched, no farm key present). Not exercised: `dev-update.sh --rollback` itself (S: the previous release is still in `/data/didle`, the image devp8 is still loaded) and the watch UI in a browser (U).
