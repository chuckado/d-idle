# Sandbox image devp16 (loaded, not switched)

> status: current (2026-11-21, card devp16). The image is loaded on the worker host; switching the container is the coordinator's step with the user's go. Earlier note: docs/process/sandbox-devp15.md.

## What devp16 is
- One private fork branch, `didle/devp16-20261121`: the level-objects branch (devp15 code, level monsters, and `objpresets` in `debug levelmonsters`) merged with the tests-only fix branch of devp15 (the vault_verbs race and the smoke save-mtime fix). V: `git merge-base --is-ancestor` showed the fix branch was not contained, and the merge touched only two files under tests/.
- Defaults unchanged: `--ladder-runewords` stays OFF.
- Built per docs/process/fork-rebase.md (long lane on the builder, d2d plus the runeword test), then `tools/dev-update.sh --image d2dbot:devp16 --d2d <binary>` (a docker build only, load, never `up`).

## Tests (V: run on the builder against the built binary)
- rc 0: test_runeword, levelmap, levelmonsters, levelobjects, input_items_diff, input_verbs, lab_verbs, vault_verbs, item_verbs, item_colour, item_pic, labels, merc_verbs, cursor_verb, overlay.
- smoke_d2d fails at line 512 (`cmd skill didn't hurt zombie`), the same assertion as on devp13, devp14 and devp15: pre-existing, not a blocker.
- Image smoke: the image's d2d starts and prints its banner (git hash unknown: prebuilt binary).

## Sandbox unchanged (V)
`docker ps` before and after the load: the sandbox container has the same id, image (devp15), healthy state and start time. No container was created, restarted or stopped.

## Recreate and rollback
Private: the recreate command for the sandbox compose set takes the image tag as its only argument; recreate with `d2dbot:devp16`, roll back with `d2dbot:devp15`. Save data lives in the mounted data directory, not in the image. Details are in the private ops notes.
