# Sandbox image devp30 (loaded, not switched)

> status: current (2026-11-25, round didle/rebase-20261125 of docs/process/fork-rebase.md). The image is loaded on the sandbox host next to devp29; the running container is untouched and stays on `d2dbot:devp27`. Switching needs the user's go. Tags: V = verified (how said), S = read from a tool or log, U = unverified. Previous notes: docs/process/sandbox-devp16.md (older), the private ops notes for devp27 and devp29.

## What devp30 is
- Fork branch `didle/rebase-20261125` on the private remote, tip `418eb931`: a MERGE (no rewrite, no force-push) of upstream main `d96a39adf` into `didle/directequip-20261125` (devp29: `item equipdirect`, `unequipdirect`, `wearset`) plus `didle/stepverbs3-20261124` (devp28: `debug rooms up`). V: `git merge-base --is-ancestor` shows the devp29 branch holds labverbs, stepverbs, stepverbs2 and rebase-20261124, but NOT stepverbs3, so stepverbs3 was merged in first (clean, 4 files).
- Behind upstream: 294 commits before (merge base `ff6d47a66`, 300 files), 0 after (V: `git rev-list --count HEAD..upstream/main`).
- Patches already upstream: none to drop. `git cherry upstream/main HEAD` finds no identical patch-id (upstream squashes), but the content of our merged PRs (#727 pet search order, #728 merc regenerates life, #708 merc RogueMissile, earlier #699 #693 #680 #678 #675) is identical on both sides: the merge made no conflict there and `git diff upstream/main HEAD` holds no pet, merc or monster-AI hunk (V, 116 files differ, all lab verbs, tests and docs of the fork).
- Upstream's big change in the range is a file move: `apps/d2d` is split into `audio/ frontend/ input/ items/ panels/ render/ ...`, `world.hpp` and `rules.hpp` lose functions into other headers. All conflicts come from that.

## Conflicts and how they were resolved (S: by reading both sides; V: the build and the tests below)
| File | Cause | Resolution |
|---|---|---|
| apps/d2d/CMakeLists.txt | source list reorganised upstream | upstream's list; our `devctl/debug_skills.cpp` and `lockstep.cpp` added in sort order |
| apps/d2d/main.cpp | includes moved to subdirectories | upstream's block; `devctl_dump.hpp`, `lockstep.hpp` kept |
| components/game/gamedata/gamedata.hpp | contents split into other headers | upstream's file; `wait_for_level_builds` and `<atomic>` re-added; `teleport_pad_target` declared in `level.hpp` next to `near_rooms` (compile error found it) |
| components/game/world.hpp (+ world/travel.cpp) | `warp_clicked` / `warp_close` moved to travel.cpp | declarations upstream's; our whole-pixel compare in `warp_clicked` and `warp_click_point` (lab read) moved into travel.cpp |
| components/rules/rules.hpp | functions split to `items/item_gen.hpp` | upstream's file; `ladder_runewords()` (flag `--ladder-runewords`, default off) put beside `runeword_row` in item_gen.hpp; the rules.cpp use merged by itself |
After the merge two compile fixes in our own files: `debug_lab.cpp` includes `panels/panels.hpp`, and uses `game_data->layouts.layout_of`.

## Binary and image
- Binary built from `418eb931` on the builder long lane (cmake flags of build-fork.sh, d2d only, -j3; tests off): `/data/d2d-lab/qdbin/d2d-devp30-rebase`, 18172888 bytes, sha256 prefix `fb37a805851746cd` (V). No existing binary replaced.
- Image `d2dbot:devp30`, id `c073131dfc45` (V: `docker images` on the sandbox host), built on the NAS with `D2_LABRUN_HOST=<the builder> tools/dev-update.sh --image d2dbot:devp30 --d2d <builder path>`, moved with `tools/imgbuild.sh save` and `docker load`. V: the binary inside the image has the same sha256 prefix as the builder's and `--help` lists `--lab-verbs`, `--item-verbs`, `--vault-verbs`. Load only: the sandbox container kept its id, image (devp27) and healthy state (V: `docker ps` and `docker inspect` before and after).

## Tests (V, builder, `D2_MPQ_DIR=/data/gamedata`)
- rc 0: devp29, devp28, devp26, stepverbs, vault_verbs, item_verbs, lab_verbs, merc_verbs, levelmap, levelmonsters, levelobjects, combat_reads, input_items_diff, input_verbs, input_verbs2_diff, creditgold, itemprice, itemdecode, warptriggers, pad_use, item_colour, item_pic, labels, cursor_verb, overlay, smoke_d2d (smoke passes here; it failed on upstream main earlier).
- Flaky, not a regression (V by reruns): `labverbs_d2d` and `paygold_d2d` each failed once and passed on rerun; the devp29 binary shows the same flake (labverbs 1 of 3, paygold 0 of 3). Cause is the known session-copy read race (the next frame refreshes the session copy; the check reads it at once).
- Proving Ground (V, one job on the long lane, `PG_DIR=/data/d2d-lab/pg-rebase25`): `act1.burial.blood_raven` Paladin seed 11 GREEN (60 vmin), `vault.equip-v2-swap` and `vault.equip-v2-set10` GREEN (3 vmin each; devp29 semantics). Each carries the scenario's own `[CHEATED setup]` tag.
- U: upstream's own ctest suite was not built (d2d only); the 294 upstream commits were not played in a live game; no Proving Ground run of other scenarios.

## What changes for the bot (vs devp29)
Nothing in the verbs. The behaviour of d2d follows upstream's 294 commits (animated tile tick 0x28 ms, item unit ids per game, merc and pet fixes, others; list: `git log ff6d47a66..d96a39adf --first-parent` in the fork). Default knobs are unchanged: `--ladder-runewords` off.

## Switch, backup and rollback (user go needed; the coordinator runs it)
- Backup first: the sandbox container's compose set is copied to a `bak-devp30` folder on the sandbox host before the recreate (same as devp27).
- Switch: the recreate script of the sandbox compose set takes the image tag only: `recreate-current.sh d2dbot:devp30` (its tuning, supervisor and operator overlays stay). Save data lives in the mounted data directory, not in the image.
- Rollback: `recreate-current.sh d2dbot:devp27` (the running image today; `d2dbot:devp29` is loaded too and holds the direct equip verbs on the older base, `d2dbot:devp22` is the older fallback). Images are kept.
- Watch after a switch: healthy, restarts 0, no ERROR line, hero progressing, 60 minutes (as for devp27).
