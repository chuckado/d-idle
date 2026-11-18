# Sandbox d2d build devp10 (2026-11-17)

> status: current. Task: TASK-623 (sandbox3). Tags: V = verified in this run (how said), S = read from a tool or log, U = unverified. Image built and LOADED only; not deployed. Follows `docs/process/sandbox-devp9.md`.

devp9 lacks two fork verb sets the merged bot code uses: the vault verbs (`ground full`, `ground take`, `item full`, `item create`, `item vault`, `vault stats`, flag `--vault-verbs`, `docs/d2d-vault-verbs.md`) and the labels verb (`labels on|off|toggle|status`, cap `ui.labels`, needs `--item-verbs` or `--lab-verbs`). devp10 adds both on current upstream main.

## Fork branch

- `didle/rebase-20261117` in the private d2d fork, pushed to the private remote the same day (no force-push). Base: upstream main f1533770 (2026-11-17, PR 597). Tip: 9df9b936. V (`git push` output).
- Built by cherry-picking with `-x`: the 39 non-merge commits of `didle/rebase-20261115` (all that were not upstream), then the six vault commits of `didle/vault-verbs` (tip f2ed52c8) that are not in the 20261115 lineage, then `didle/labels-verb` (58f45660), then one build fix. V.
- One conflict: `tests/CMakeLists.txt` (both branches add an `add_test` after `item_verbs_d2d`): kept both `vault_verbs_d2d` and `labels_d2d`.
- Build fix (9df9b936): upstream 9f32547f removed `Monster::last_hp` (Lightning Enchanted fires on landed hits, not life drops); the lab verb `debug mon set life` no longer resets it (the line was the only user).
- Dropped as upstream: nothing further. The 20261115 lineage had already dropped the warp patches (PR 490, 547); a cherry-pick of all 39 applied clean, and `git grep` of upstream/main finds no framefeed, item-verbs, lab-verbs or vault code. V.

## Carried (fork-only) patches in devp10

Lab patch set port and lab/test cheats behind `--lab-verbs` (debug act/hit/merc/give/quest/spawn, missiles, skill drills); `--overlay-dir` table overlay; SPDX and fork wording; `dump merc` (incl. live merc gear list); framefeed verb (`ced1a0f5`, the ring file) and `cursor on|off`; icon art and Steam art clean-up; item verbs (equip, unequip, move, identify, drop, gold stash, `--item-verbs` gate); debug warps click point; merc-equip / merc-unequip; `dump stats` hero block with gear sums and damage; README-DIDLE; vault verbs (spike plus V1); labels verb.

## Upstream PRs of ours included (base f1533770)

Everything merged upstream as of today, including 586, 589 (gamble), 590 (umod) and the earlier ones listed in devp9 (557, 553, 548, 547, 541, 539, 530, 490, 445, 444 and older). S (the brief names 586, 589, 590; `git log upstream/main ^485fe535` has 214 commits since the devp9 base).

## Build and tests

- `tools/ops.sh run build-fork.sh didle/rebase-20261117 --tests --jobs 2 --timeout 3600` through `tools/agentrun.sh`, NAS builder, -j2. First run failed at `debug_skills.cpp` (`last_hp`, above); second run: d2d 622 s, tests 61 s, both rc 0. V. The script's copy to the worker host's `fork/bin` failed for the d2d binary (a branch name with a slash makes a missing directory; the test binaries copied after the directories were made). The binary used is on the builder: `/data/build/didle/rebase-20261117-9df9b936-d2d/apps/d2d/d2d`, sha256 prefix 44123393.
- Headless verb tests on the builder (`D2_MPQ_DIR=/data/gamedata`, nice, scripts in the lab dir): `item_verbs_d2d` ok, `lab_verbs_d2d` ok, `vault_verbs_d2d` ok (0 failures), `labels_d2d` ok, `cursor_verb_d2d` ok (the framefeed has no dedicated test; the cursor test runs the frame path, so framefeed itself is U), `merc_verbs_d2d` ok, `overlay_d2d` ok. V (all rc 0).
- Combined flags: with `--item-verbs --vault-verbs` `caps` lists `ui.labels`, the item and merc verbs and `ground.full,ground.take,item.full,item.create,item.vault,vault.stats`. V (one-off script, output read).
- Test binaries with game data on the builder: `test_rules`, `test_combat`, `test_monsters`, `test_protocol`, `test_game` rc 0 (V); `test_next_hit`, `test_armor_hit`, `test_freeze_resume` need an oracle mode and case files (rc 77, SKIP); the oracle comparisons were not run (U).

## Image

- `tools/dev-update.sh --image d2dbot:devp10 --d2d /data/build/didle/rebase-20261117-9df9b936-d2d/apps/d2d/d2d` with `D2_LABRUN_HOST=<sandbox host>` and `local.env` loaded (restricted key, `tools/imgbuild.sh`), image id 51e4773cf727. `tools/imgbuild.sh save d2dbot:devp10 OUT.tar.gz` (168 MB), `gunzip -c OUT.tar.gz | docker load` on the sandbox host. V.
- Checked with a throwaway `docker run --rm --entrypoint /opt/d2d/d2d d2dbot:devp10 --help`: lists `--vault-verbs`, `--item-verbs`, `--lab-verbs`; the binary hash equals the builder's. V. The running container `d2play-20261108` (still `devp9`, healthy) and the compose files were not touched.

## Switch later (needs the user's go)

Flags are set in the compose environment `D2BOT_ARGS` of `iv/iv.yml` in `~/play-build-20261108` on the sandbox host (S: read there; value now `--d2d-args --item-verbs`). The entrypoint runs `eval exec d2bot $args ${D2BOT_ARGS}` and the bot flag `-d2d-args` takes ONE string of space-separated d2d arguments, so two flags need quotes. Change: `D2BOT_ARGS: '--d2d-args "--item-verbs --vault-verbs"'` (a copy `iv/iv2.yml` is the cleaner way: no edit of the existing file). The labels verb needs no extra flag (`--item-verbs`). Do not add `--lab-verbs` (cheat verbs).

Steps (same as devp9, section Deploy): `docker stop -t 60 d2play-20261108`; back up the volume (`docker run --rm --user 0 -v d2play-data:/data:ro --entrypoint tar d2dbot:devp10 czf - -C /data .` into a `bak-devp9` file); then
`D2_IMAGE=d2dbot:devp10 docker compose -p play-build-20261108 -f compose.yml -f override.yml -f dev.yml -f feed.yml -f iv/iv2.yml up -d --no-deps d2dbot`
with the exports of `up.sh` (`D2_GAME_DIR`, `D2_WATCH_ORIGIN`, `D2_PUBLISH`, `D2_DEV_URL`, `D2_DEV_PUBKEY`); then `tools/dev-update.sh --serve start` and one `--commit origin/main --to d2play-20261108`. What changes: the d2d binary (new upstream fidelity fixes since 485fe535, vault and labels verbs available), the d2d command line (one more flag, `--vault-verbs`); volume, game mount and bot knobs are unchanged. The bot's vault code stays inert until its knob turns it on (`loot.vault`, S: `internal/loot/vault.go`).

## Rollback

`devp9` stays loaded. Same compose command with `D2_IMAGE=d2dbot:devp9` and `-f iv/iv.yml` (the old `--item-verbs` only); data back with the backup tarball if the hero was saved by devp10 and something is wrong (a hero saved by a newer d2d is still a classic d2s, S).
